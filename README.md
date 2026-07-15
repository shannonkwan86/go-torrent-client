# go-torrent-client

一个用 Go 编写的精简 BitTorrent 下载客户端。项目以学习 BitTorrent 协议和 Go 网络并发为目标，完整走通了 `.torrent` 解析、Tracker Peer 发现、Peer Wire Protocol、流水线分块下载、Piece 完整性校验和文件组装。

> 本项目最初参考了 [Building a BitTorrent client from the ground up in Go](https://blog.jse.li/posts/torrent/) 的实现思路，并在此基础上补充了协议正确性、异常处理、测试和 goroutine 生命周期管理。它适合用于学习和展示，不定位为生产级下载器。

## 功能

- 解析单文件 `.torrent` 元数据
- 对原始 bencode `info` 字典字节计算 InfoHash
- 通过 HTTP Tracker 获取并解析 compact Peer 列表
- 完成 BitTorrent Handshake、Bitfield、Interested、Choke/Unchoke、Have、Request 和 Piece 消息处理
- 将 Piece 拆为最大 16 KiB 的 Block，并维持最多 5 个在途请求
- 支持乱序 Block 拼装、重复响应过滤和未请求 Block 检查
- 使用 SHA-1 校验完整 Piece，失败后重新调度
- 由多个 Peer worker 并发领取 Piece，并按 Piece index 组装文件
- 支持无进展超时、Peer 全部失败检测和可取消的 worker 清理

## 下载流程

```text
.torrent
   │
   ├─ 解析 announce、文件信息和 PieceHash
   └─ 对原始 info 字典计算 InfoHash
              │
              ▼
        HTTP Tracker announce
              │
              ▼
          compact Peer 列表
              │
              ▼
     每个 Peer 启动一个 worker
              │
       Handshake → Bitfield
              │
       Interested → Unchoke
              │
       Piece → 多个 Block 请求
              │
       按偏移拼装并校验 SHA-1
              │
              ▼
      按 Piece index 组装完整文件
```

## 项目结构

| 包 | 职责 |
| --- | --- |
| `torrentfile` | 解析 `.torrent`、计算 InfoHash、请求 Tracker、写出文件 |
| `peers` | 解析 compact IPv4 Peer 地址 |
| `handshake` | 编解码 BitTorrent Handshake |
| `message` | 编解码 Peer Wire Protocol 消息 |
| `bitfield` | 查询和更新 Peer 拥有的 Piece |
| `client` | 管理单个 Peer 的 TCP 连接和协议消息 |
| `p2p` | Piece 调度、Block 流水线、校验、并发与退出协调 |

## 构建与运行

要求 Go 1.13 或更高版本。

```bash
git clone https://github.com/shannonkwan86/go-torrent-client.git
cd go-torrent-client
go build -o go-torrent-client .
```

下载单文件 torrent：

```bash
./go-torrent-client input.torrent output.file
```

例如：

```bash
./go-torrent-client debian-13.6.0-amd64-netinst.iso.torrent debian.iso
```

## 网络与代理

Tracker 请求使用 Go 的标准 HTTP Client，因此可以通过环境变量使用 HTTP 代理：

```bash
HTTP_PROXY=http://127.0.0.1:7897 \
  ./go-torrent-client input.torrent output.file
```

Peer 数据连接使用原始 TCP，不经过 `HTTP_PROXY`。如果日志显示 Tracker 成功但所有 Peer 握手都失败，请检查 VPN/TUN/虚拟网卡是否改变了直连 TCP 的路由。

真实 Debian torrent 已验证能够完成 Tracker 请求、Peer 握手并持续下载 Piece。公共 Tracker 返回的列表通常包含离线或不可直连 Peer，因此少量 `timeout` 和 `EOF` 属于正常现象。

## 测试

```bash
go test ./...
go test -race ./...
```

测试覆盖 bencode 原始 `info` 提取、Tracker 响应、Peer 地址、Handshake、Bitfield、消息编解码、重复/未请求 Block 和 worker 取消等核心行为。

## 当前限制

- 仅支持 `.torrent` 文件，不支持 Magnet 和 DHT
- 仅支持 HTTP Tracker，不支持 UDP Tracker
- 仅支持单文件 torrent
- 只下载，不上传 Piece
- Piece 调度采用共享队列，不支持 rarest-first 和 endgame
- 当前将完整文件保存在内存中，全部 Piece 完成后才写入磁盘
- 不支持断点续传

对于大型文件，内存占用会接近文件大小。后续优先方向是 Piece 校验后使用 `WriteAt` 写入临时文件，完成后再原子重命名。

## License

[GNU General Public License v3.0](LICENSE)
