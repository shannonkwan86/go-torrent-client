package torrentfile

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jackpal/bencode-go"
	"github.com/shannonkwan86/go-torrent-client/peers"
)

// bencodeTrackerResp 对应 tracker 返回的 bencode 响应。
// Peers 字段仍然是 compact peers 原始数据，后续会交给 peers.Unmarshal 解析。
type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

// buildTrackerURL 根据 torrent 元信息和当前客户端身份构造 tracker announce 请求地址。
// 这里的 info_hash 和 peer_id 必须按原始 20 字节传入，再交给 URL 编码处理。
func (tf *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(tf.Announce)
	if err != nil {
		return "", err
	}

	query := base.Query()
	query.Set("info_hash", string(tf.InfoHash[:]))
	query.Set("peer_id", string(peerID[:]))
	query.Set("port", strconv.Itoa(int(port)))
	query.Set("uploaded", "0")
	query.Set("downloaded", "0")
	query.Set("compact", "1")
	query.Set("left", strconv.Itoa(tf.Length))
	base.RawQuery = query.Encode()

	return base.String(), nil
}

// requestPeers 向 tracker 请求正在参与这个 torrent 的 peer 列表。
func (tf *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peers.Peer, error) {
	trackerURL, err := tf.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(trackerURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var trackerResp bencodeTrackerResp
	if err := bencode.Unmarshal(resp.Body, &trackerResp); err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(trackerResp.Peers))
}
