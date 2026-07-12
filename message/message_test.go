package message

import (
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"
)

func TestWriteAndReadRoundTrip(t *testing.T) {
	original := &Message{ID: MsgInterested, Payload: []byte{1, 2, 3}}
	var wire bytes.Buffer

	if err := original.Write(&wire); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got, err := Read(&wire)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got == nil {
		t.Fatal("Read() = nil, want message")
	}
	if got.ID != original.ID {
		t.Errorf("ID = %d, want %d", got.ID, original.ID)
	}
	if !bytes.Equal(got.Payload, original.Payload) {
		t.Errorf("Payload = %v, want %v", got.Payload, original.Payload)
	}
}

func TestReadKeepAliveReturnsNil(t *testing.T) {
	got, err := Read(bytes.NewReader([]byte{0, 0, 0, 0}))
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got != nil {
		t.Errorf("Read() = %#v, want nil", got)
	}
}

func TestReadRejectsOversizedMessage(t *testing.T) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], MaxMessageLength+1)

	_, err := Read(bytes.NewReader(length[:]))
	if err == nil {
		t.Fatal("Read() error = nil, want oversized-message error")
	}
}

func TestReadHandlesFragmentedInput(t *testing.T) {
	data := []byte{0, 0, 0, 4, MsgUnchoke, 7, 8, 9}

	got, err := Read(&oneByteReader{reader: bytes.NewReader(data)})
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got == nil {
		t.Fatal("Read() = nil, want message")
	}
	if got.ID != MsgUnchoke || !bytes.Equal(got.Payload, []byte{7, 8, 9}) {
		t.Errorf("Read() = %#v, want ID %d with payload [7 8 9]", got, MsgUnchoke)
	}
}

type oneByteReader struct {
	reader io.Reader
}

func (r *oneByteReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		p = p[:1]
	}
	return r.reader.Read(p)
}

func TestReadReturnsErrorForTruncatedMessage(t *testing.T) {
	_, err := Read(strings.NewReader("\x00\x00\x00\x02\x01"))
	if err == nil {
		t.Fatal("Read() error = nil, want truncated-message error")
	}
}

func TestWriteReturnsErrorForShortWrite(t *testing.T) {
	err := (&Message{ID: MsgInterested}).Write(shortWriter{})
	if err == nil {
		t.Fatal("Write() error = nil, want short-write error")
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	return len(p) - 1, nil
}
