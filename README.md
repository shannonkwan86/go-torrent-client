# go-torrent-client

[![CircleCI](https://circleci.com/gh/shannonkwan86/go-torrent-client.svg?style=shield)](https://circleci.com/gh/shannonkwan86/go-torrent-client)

Tiny BitTorrent client written in Go. Read the blog post: https://blog.jse.li/posts/torrent/

## Install

```sh
go install github.com/shannonkwan86/go-torrent-client@latest
```

## Usage
Try downloading the current [Debian netinst torrent](https://cdimage.debian.org/debian-cd/current/amd64/bt-cd/)!

```sh
go-torrent-client debian-13.6.0-amd64-netinst.iso.torrent debian.iso
```

## Limitations
* Only supports `.torrent` files (no magnet links)
* Only supports HTTP trackers
* Does not support multi-file torrents
* Strictly leeches (does not support uploading pieces)
