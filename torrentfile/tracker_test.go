package torrentfile

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackpal/bencode-go"
)

func TestBuildTrackerURL(t *testing.T) {
	var infoHash, peerID [20]byte
	copy(infoHash[:], "info-hash-1234567890")
	copy(peerID[:], "peer-id--1234567890")
	tf := TorrentFile{Announce: "https://tracker.example/announce?existing=value", InfoHash: infoHash, Length: 123}

	rawURL, err := tf.buildTrackerURL(peerID, 6881)
	if err != nil {
		t.Fatalf("buildTrackerURL() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, rawURL, nil)
	query := req.URL.Query()
	if query.Get("info_hash") != string(infoHash[:]) || query.Get("peer_id") != string(peerID[:]) {
		t.Error("tracker URL does not preserve binary hashes")
	}
	if query.Get("port") != "6881" || query.Get("left") != "123" || query.Get("compact") != "1" {
		t.Errorf("tracker query = %v, want required announce parameters", query)
	}
	if query.Get("existing") != "value" {
		t.Error("tracker URL dropped existing query parameter")
	}
}

func TestRequestPeers(t *testing.T) {
	var infoHash, peerID [20]byte
	copy(infoHash[:], "info-hash-1234567890")
	copy(peerID[:], "peer-id--1234567890")
	compactPeers := string([]byte{127, 0, 0, 1, 0x1A, 0xE1})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("compact") != "1" {
			t.Errorf("compact query = %q, want 1", r.URL.Query().Get("compact"))
		}
		var body bytes.Buffer
		if err := bencode.Marshal(&body, bencodeTrackerResp{Interval: 60, Peers: compactPeers}); err != nil {
			t.Fatalf("bencode.Marshal() error = %v", err)
		}
		_, _ = w.Write(body.Bytes())
	}))
	defer server.Close()

	tf := TorrentFile{Announce: server.URL, InfoHash: infoHash, Length: 10}
	got, err := tf.requestPeers(peerID, 6881)
	if err != nil {
		t.Fatalf("requestPeers() error = %v", err)
	}
	if got, want := len(got), 1; got != want {
		t.Fatalf("peer count = %d, want %d", got, want)
	}
	if got[0].String() != "127.0.0.1:6881" {
		t.Errorf("peer = %q, want %q", got[0].String(), "127.0.0.1:6881")
	}
}
