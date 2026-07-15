package p2p

import (
	"crypto/sha1"
	"encoding/binary"
	"net"
	"testing"
	"time"

	"github.com/shannonkwan86/go-torrent-client/client"
	"github.com/shannonkwan86/go-torrent-client/message"
	"github.com/shannonkwan86/go-torrent-client/peers"
)

func TestPieceProgressIgnoresDuplicateAndRejectsUnrequestedBlock(t *testing.T) {
	clientConn, peerConn := net.Pipe()
	defer clientConn.Close()
	defer peerConn.Close()
	state := pieceProgress{
		index:    0,
		client:   &client.Client{Conn: clientConn},
		buf:      make([]byte, 6),
		backlog:  1,
		pending:  map[int]int{0: 3},
		received: make(map[int]bool),
	}

	writePieceMessage(t, peerConn, 0, []byte("abc"))
	if err := state.readMessage(); err != nil {
		t.Fatalf("readMessage() error = %v", err)
	}
	if state.downloaded != 3 || state.backlog != 0 {
		t.Fatalf("state after first block = downloaded %d, backlog %d", state.downloaded, state.backlog)
	}

	writePieceMessage(t, peerConn, 0, []byte("abc"))
	if err := state.readMessage(); err != nil {
		t.Fatalf("duplicate readMessage() error = %v", err)
	}
	if state.downloaded != 3 || state.backlog != 0 {
		t.Fatalf("duplicate changed state to downloaded %d, backlog %d", state.downloaded, state.backlog)
	}

	writePieceMessage(t, peerConn, 3, []byte("def"))
	if err := state.readMessage(); err == nil {
		t.Fatal("unrequested block error = nil")
	}
}

func writePieceMessage(t *testing.T, conn net.Conn, begin int, data []byte) {
	t.Helper()
	payload := make([]byte, 8+len(data))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	copy(payload[8:], data)
	go func() {
		_, _ = conn.Write((&message.Message{ID: message.MsgPiece, Payload: payload}).Serialize())
	}()
}

func TestDownloadReturnsWhenAllPeersFail(t *testing.T) {
	var peerID, infoHash [20]byte
	data := []byte("one piece")
	torrent := Torrent{
		Peers:       []peers.Peer{{IP: []byte{127, 0, 0, 1}, Port: 1}},
		PeerID:      peerID,
		InfoHash:    infoHash,
		PieceHashes: [][20]byte{sha1.Sum(data)},
		PieceLength: len(data),
		Length:      len(data),
	}

	done := make(chan error, 1)
	go func() {
		_, err := torrent.Download()
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("Download() error = nil, want peer failure")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Download() blocked after all peers failed")
	}
}

func TestSendPieceResultStopsWithoutReceiver(t *testing.T) {
	results := make(chan *pieceResult)
	stop := make(chan struct{})
	close(stop)

	done := make(chan bool, 1)
	go func() {
		done <- sendPieceResult(results, &pieceResult{index: 1}, stop)
	}()

	select {
	case sent := <-done:
		if sent {
			t.Fatal("sendPieceResult() = true after stop, want false")
		}
	case <-time.After(time.Second):
		t.Fatal("sendPieceResult() blocked after stop")
	}
}
