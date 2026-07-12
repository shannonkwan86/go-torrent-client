package main

import (
	"log"
	"os"

	"github.com/shannonkwan86/go-torrent-client/torrentfile"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go-torrent-client <torrent-file> <output-file>")
	}
	inPath := os.Args[1]
	outPath := os.Args[2]

	tf, err := torrentfile.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.DownloadToFile(outPath)
	if err != nil {
		log.Fatal(err)
	}
}
