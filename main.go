package main

import (
	"log"
	"os"

	"github.com/shannonkwan86/go-torrent-client/torrentfile"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: go-torrent-client <torrent-file>")
	}

	inPath := os.Args[1]

	tf, err := torrentfile.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tf)
}
