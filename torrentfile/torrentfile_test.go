package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackpal/bencode-go"
)

func TestSplitPieceHashes(t *testing.T) {
	pieces := string(bytes.Repeat([]byte{1}, 20)) + string(bytes.Repeat([]byte{2}, 20))

	got, err := splitPieceHashes(pieces)
	if err != nil {
		t.Fatalf("splitPieceHashes() error = %v", err)
	}
	if got, want := len(got), 2; got != want {
		t.Fatalf("hash count = %d, want %d", got, want)
	}
	if got[0][0] != 1 || got[1][0] != 2 {
		t.Errorf("hashes = %v, want distinct 20-byte hashes", got)
	}
}

func TestSplitPieceHashesRejectsInvalidLength(t *testing.T) {
	_, err := splitPieceHashes("too short")
	if err == nil {
		t.Fatal("splitPieceHashes() error = nil, want error")
	}
}

func TestToTorrentFile(t *testing.T) {
	pieces := string(bytes.Repeat([]byte{3}, 20))
	bto := bencodeTorrent{
		Announce: "https://tracker.example/announce",
		Info: bencodeInfo{
			Pieces:      pieces,
			PieceLength: 16,
			Length:      42,
			Name:        "example.txt",
		},
	}

	got, err := bto.toTorrentFile()
	if err != nil {
		t.Fatalf("toTorrentFile() error = %v", err)
	}

	var encoded bytes.Buffer
	if err := bencode.Marshal(&encoded, bto.Info); err != nil {
		t.Fatalf("bencode.Marshal() error = %v", err)
	}
	wantHash := sha1.Sum(encoded.Bytes())
	if got.Announce != bto.Announce || got.Name != bto.Info.Name || got.Length != bto.Info.Length || got.PieceLength != bto.Info.PieceLength {
		t.Errorf("TorrentFile metadata = %#v, want fields from bencode input", got)
	}
	if got.InfoHash != wantHash {
		t.Errorf("InfoHash = %x, want %x", got.InfoHash, wantHash)
	}
	if got, want := len(got.PieceHashes), 1; got != want {
		t.Errorf("piece hash count = %d, want %d", got, want)
	}
}

func TestOpen(t *testing.T) {
	bto := bencodeTorrent{
		Announce: "https://tracker.example/announce",
		Info: bencodeInfo{
			Pieces:      string(bytes.Repeat([]byte{4}, 20)),
			PieceLength: 32,
			Length:      32,
			Name:        "sample.bin",
		},
	}

	var encoded bytes.Buffer
	if err := bencode.Marshal(&encoded, bto); err != nil {
		t.Fatalf("bencode.Marshal() error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "sample.torrent")
	if err := os.WriteFile(path, encoded.Bytes(), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	got, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if got.Name != "sample.bin" || got.Length != 32 || got.PieceLength != 32 {
		t.Errorf("Open() = %#v, want parsed torrent metadata", got)
	}
}
