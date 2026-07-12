package handshake

import (
	"bytes"
	"testing"
)

func TestNewSerializeAndReadRoundTrip(t *testing.T) {
	var infoHash, peerID [20]byte
	copy(infoHash[:], "info-hash-1234567890")
	copy(peerID[:], "peer-id--1234567890")

	original := New(infoHash, peerID)
	wire := original.Serialize()

	if got, want := len(wire), 68; got != want {
		t.Fatalf("serialized length = %d, want %d", got, want)
	}
	if got := wire[0]; got != byte(len("BitTorrent protocol")) {
		t.Errorf("protocol length = %d, want %d", got, len("BitTorrent protocol"))
	}

	got, err := Read(bytes.NewReader(wire))
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if *got != *original {
		t.Errorf("Read() = %#v, want %#v", got, original)
	}
}

func TestReadRejectsZeroProtocolLength(t *testing.T) {
	_, err := Read(bytes.NewReader([]byte{0}))
	if err == nil {
		t.Fatal("Read() error = nil, want error")
	}
}

func TestReadRejectsTruncatedHandshake(t *testing.T) {
	_, err := Read(bytes.NewReader([]byte{19, 'B'}))
	if err == nil {
		t.Fatal("Read() error = nil, want error")
	}
}
