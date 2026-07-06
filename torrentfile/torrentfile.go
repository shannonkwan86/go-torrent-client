package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

// TorrentFile 是程序内部使用的 torrent 元信息结构。
// 它不是 .torrent 文件的原始形状，而是把原始 bencode 数据整理成后续下载更方便使用的形式。
type TorrentFile struct {
	Announce    string
	Name        string
	Length      int
	PieceLength int
	PieceHashes [][20]byte
	InfoHash    [20]byte
}

// bencodeInfo 对应 .torrent 文件里的 "info" 字典。
// 这个结构用于接收 bencode 解码后的原始字段。
type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

// bencodeTorrent 对应 .torrent 文件的顶层结构。
type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

// splitPieceHashes 将 .torrent 中连续存放的 pieces 字符串拆成一组 20 字节的 SHA-1 hash。
func splitPieceHashes(pieces string) ([][20]byte, error) {

	// 每个 piece hash 都是一个 SHA-1 结果，固定为 20 字节。
	if len(pieces)%20 != 0 {
		return nil, fmt.Errorf("pieces length must be multiple of 20, got %d", len(pieces))
	}
	hashes := make([][20]byte, len(pieces)/20)

	for i := 0; i*20 < len(pieces); i += 1 {
		var start, end int = i * 20, i*20 + 20
		copy(hashes[i][:], []byte(pieces[start:end]))
	}

	return hashes, nil
}

// hash 计算 info 字典的 SHA-1 hash，也就是 BitTorrent 协议里的 info hash。
// 注意：info hash 是对重新 bencode 后的 info 字典计算的，不是对整个 .torrent 文件计算的。
func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	var hash [20]byte
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return hash, err
	}
	hash = sha1.Sum(buf.Bytes())
	return hash, nil
}

// Open 读取 .torrent 文件，并把原始 bencode 数据转换成程序内部使用的 TorrentFile。
func Open(path string) (*TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var bto bencodeTorrent
	if err := bencode.Unmarshal(file, &bto); err != nil {
		return nil, err
	}

	return bto.toTorrentFile()
}

// toTorrentFile 将 bencode 解码后的原始结构转换成程序内部使用的 TorrentFile。
func (bto *bencodeTorrent) toTorrentFile() (*TorrentFile, error) {
	pieceHashes, err := splitPieceHashes(bto.Info.Pieces)
	if err != nil {
		return nil, err
	}

	infoHash, err := bto.Info.hash()
	if err != nil {
		return nil, err
	}
	return &TorrentFile{
		Announce:    bto.Announce,
		Name:        bto.Info.Name,
		Length:      bto.Info.Length,
		PieceLength: bto.Info.PieceLength,
		PieceHashes: pieceHashes,
		InfoHash:    infoHash,
	}, nil
}
