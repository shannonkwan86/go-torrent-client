package peers

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

// Peer 表示一个可以连接的 BitTorrent 节点地址。
type Peer struct {
	IP   net.IP
	Port uint16
}

// Unmarshal 将 tracker 返回的 compact peers 二进制数据解析成 Peer 列表。
func Unmarshal(peersBin []byte) ([]Peer, error) {
	const peerSize = 6
	if len(peersBin)%peerSize != 0 {
		return nil, fmt.Errorf("malformed peers length: %d", len(peersBin))
	}

	numPeers := len(peersBin) / peerSize
	peers := make([]Peer, numPeers)
	for i := 0; i < numPeers; i++ {
		offset := i * peerSize
		peers[i].IP = net.IP(peersBin[offset : offset+4])
		peers[i].Port = binary.BigEndian.Uint16(peersBin[offset+4 : offset+6])
	}
	return peers, nil

}

// String 将 Peer 格式化成网络连接常用的 "host:port" 字符串。
func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}
