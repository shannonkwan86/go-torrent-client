package torrentfile

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/shannonkwan86/go-torrent-client/peers"

	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	FailureReason string `bencode:"failure reason"`
	Interval      int    `bencode:"interval"`
	Peers         string `bencode:"peers"`
}

const trackerDiscoveryRounds = 3

func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
		"event":      []string{"started"},
		"numwant":    []string{"500"},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

// discoverPeers performs several announces because public trackers commonly
// return a random subset containing many stale peers.
func (t *TorrentFile) discoverPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	type announceResult struct {
		peers []peers.Peer
		err   error
	}
	results := make(chan announceResult, trackerDiscoveryRounds)
	for i := 0; i < trackerDiscoveryRounds; i++ {
		go func() {
			batch, err := t.requestPeers(peerID, port)
			results <- announceResult{peers: batch, err: err}
		}()
	}

	seen := make(map[string]peers.Peer)
	var lastErr error
	for i := 0; i < trackerDiscoveryRounds; i++ {
		result := <-results
		if result.err != nil {
			lastErr = result.err
			continue
		}
		for _, peer := range result.peers {
			seen[peer.String()] = peer
		}
	}

	if len(seen) == 0 {
		if lastErr != nil {
			return nil, fmt.Errorf("peer discovery failed after %d announces: %w", trackerDiscoveryRounds, lastErr)
		}
		return nil, fmt.Errorf("tracker returned no peers after %d announces", trackerDiscoveryRounds)
	}

	result := make([]peers.Peer, 0, len(seen))
	for _, peer := range seen {
		result = append(result, peer)
	}
	return result, nil
}

func (t *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "go-torrent-client/1.0")
	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracker returned HTTP status %s", resp.Status)
	}

	trackerResp := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}
	if trackerResp.FailureReason != "" {
		return nil, fmt.Errorf("tracker failure: %s", trackerResp.FailureReason)
	}
	if trackerResp.Peers == "" {
		return nil, fmt.Errorf("tracker returned no peers")
	}

	return peers.Unmarshal([]byte(trackerResp.Peers))
}
