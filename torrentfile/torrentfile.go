package torrentfile

import (
	"os"

	"github.com/jackpal/bencode-go"
)

type TorrentFile struct {
	Announce string
	Name     string
	Length   int
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

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

	return &TorrentFile{
		Announce: bto.Announce,
		Name:     bto.Info.Name,
		Length:   bto.Info.Length,
	}, nil
}
