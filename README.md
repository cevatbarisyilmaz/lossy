# lossy
[![GoDoc](https://godoc.org/github.com/cevatbarisyilmaz/lossy?status.svg)](https://godoc.org/github.com/cevatbarisyilmaz/lossy)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/cevatbarisyilmaz/lossy?sort=semver)
![GitHub](https://img.shields.io/github/license/cevatbarisyilmaz/lossy)
[![Go Report Card](https://goreportcard.com/badge/github.com/cevatbarisyilmaz/lossy)](https://goreportcard.com/report/github.com/cevatbarisyilmaz/lossy)

Go package to simulate bandwidth, latency and packet loss for net.PacketConn and net.Conn interfaces.

```go
bandwidth := 1024 * 1024 // 8 Mbit/s
headerOverhead := lossy.UDPv4MinHeaderOverhead
minLatency := 10 * time.Millisecond
maxLatency := 100 * time.Millisecond
packetLossRate := 0.1
lossyConn := lossy.Conn(conn, bandwidth, minLatency, maxLatency, packetLossRate, headerOverhead)
```