// Package message 负责 BitTorrent peer wire 消息的编码与解码。
package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	MsgChoke         uint8 = 0
	MsgUnchoke       uint8 = 1
	MsgInterested    uint8 = 2
	MsgNotInterested uint8 = 3
	MsgHave          uint8 = 4
	MsgBitfield      uint8 = 5
	MsgRequest       uint8 = 6
	MsgPiece         uint8 = 7
	MsgCancel        uint8 = 8

	// MaxMessageLength 限制从 peer 接收一条消息时允许分配的最大内存。
	MaxMessageLength uint32 = 1 << 20
)

// Message 表示一条非空的 BitTorrent peer wire 消息。
// keep-alive 没有 ID 和 payload，因此用 nil 表示。
type Message struct {
	ID      uint8
	Payload []byte
}

// Write 将 m 编码为带长度前缀的 peer wire 消息并写入 w。
func (m *Message) Write(w io.Writer) error {
	length := uint32(1 + len(m.Payload))
	if length > MaxMessageLength {
		return fmt.Errorf("message length %d exceeds maximum %d", length, MaxMessageLength)
	}

	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[:4], length)
	buf[4] = m.ID
	copy(buf[5:], m.Payload)

	n, err := w.Write(buf)
	if err != nil {
		return err
	}
	if n != len(buf) {
		return io.ErrShortWrite
	}
	return nil
}

// Read 从 r 解码一条 peer wire 消息。
// 读到 keep-alive 时返回 (nil, nil)。
func Read(r io.Reader) (*Message, error) {
	var lengthBuf [4]byte
	if _, err := io.ReadFull(r, lengthBuf[:]); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf[:])
	if length == 0 {
		return nil, nil
	}
	if length > MaxMessageLength {
		return nil, fmt.Errorf("message length %d exceeds maximum %d", length, MaxMessageLength)
	}

	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}

	return &Message{ID: body[0], Payload: body[1:]}, nil
}
