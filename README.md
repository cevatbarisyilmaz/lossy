# lossy
[![GoDoc](https://godoc.org/github.com/cevatbarisyilmaz/lossy?status.svg)](https://godoc.org/github.com/cevatbarisyilmaz/lossy)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/cevatbarisyilmaz/lossy?sort=semver)](https://github.com/cevatbarisyilmaz/lossy/releases)
[![GitHub](https://img.shields.io/github/license/cevatbarisyilmaz/lossy)](https://github.com/cevatbarisyilmaz/lossy/blob/master/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/cevatbarisyilmaz/lossy)](https://goreportcard.com/report/github.com/cevatbarisyilmaz/lossy)

Go package to simulate bandwidth, latency and packet loss for net.PacketConn and net.Conn interfaces.

Its main usage is to test robustness of applications and network protocols run over unreliable transport protocols such as UDP or IP.
As a side benefit, it can also be used as outbound bandwidth limiter.

lossy only alters the writing side of the connection, reading side is kept as it is.

## Example

```go
package main

import (
	"fmt"
	"github.com/cevatbarisyilmaz/lossy"
	"log"
	"math/rand"
	"net"
	"time"
)

const packetSize = 64

func main() {
	// Create two connection endpoints over UDP
	packetConn, conn := createConnections()

	// Create a lossy packet connection
	bandwidth := 1048 // 8 Kbit/s
	minLatency := 100 * time.Millisecond
	maxLatency := time.Second
	packetLossRate := 0.33
	headerOverhead := lossy.UDPv4MinHeaderOverhead
	lossyPacketConn := lossy.NewPacketConn(packetConn, bandwidth, minLatency, maxLatency, packetLossRate, headerOverhead)

	// Write some packets via lossy
	var bytesWritten int
	const packetCount = 32
	go func() {
		for i := 0; i < packetCount; i++ {
			packet := createPacket()
			_, err := lossyPacketConn.WriteTo(packet, conn.LocalAddr())
			if err != nil {
				log.Fatal(err)
			}
			bytesWritten += len(packet) + headerOverhead
		}
		fmt.Println("Sent", packetCount, "packets with total size of", bytesWritten, "bytes")
		baseTransmissionDuration := time.Duration(float64(bytesWritten * int(time.Second)) / float64(bandwidth))
		earliestCompletion := baseTransmissionDuration + minLatency
		latestCompletion := baseTransmissionDuration + maxLatency
		fmt.Println("Expected transmission duration is between", earliestCompletion, "and", latestCompletion)
	}()

	// Read packets at the other side
	const timeout = 3 * time.Second
	var packets, bytesRead int
	startTime := time.Now()
	for {
		_ = conn.SetReadDeadline(time.Now().Add(timeout))
		buffer := make([]byte, packetSize)
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		bytesRead += n + headerOverhead
		packets++
	}
	dur := time.Now().Sub(startTime) - timeout
	fmt.Println("Received", packets, "packets with total size of", bytesRead, "bytes in", dur)

	// Close the connections
	_ = lossyPacketConn.Close()
	_ = conn.Close()
}

func createConnections() (net.PacketConn, net.Conn) {
	packetConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0,
	})
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp", nil, packetConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		log.Fatal(err)
	}
	return packetConn, conn
}

func createPacket() []byte {
	packet := make([]byte, packetSize)
	rand.Read(packet)
	return packet
}
```

Output
```
Sent 32 packets with total size of 2944 bytes
Expected transmission duration is between 2.909160305s and 3.809160305s
Received 23 packets with total size of 2116 bytes in 3.2507523s
```
