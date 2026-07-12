package main

import (
	"log"
	"os"

	"github.com/veggiedefender/torrent-client/torrentfile"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: torrent-client <torrent-file> <output-file>")
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
