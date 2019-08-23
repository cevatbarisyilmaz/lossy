package lossy_test

import (
	"fmt"
	"github.com/cevatbarisyilmaz/lossy"
	"log"
	"math/rand"
	"net"
	"time"
)

func Example() {
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
	lossyConn := lossy.Conn(conn, 1024, time.Second, time.Second*2, 0.25, lossy.UDPv4MinHeaderOverhead)
	var dataWritten int
	const messageCount = 32
	go func() {
		for i := 0; i < messageCount; i++ {
			message := make([]byte, i*64)
			dataWritten += len(message)
			rand.Read(message)
			_, err := lossyConn.Write(message)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println("Sent", messageCount, "messages with total size of", dataWritten, "bytes")
	}()
	const timeOut = time.Second * 5
	var dataRead int
	var messagesReceived int
	startTime := time.Now()
	for {
		err = packetConn.SetReadDeadline(time.Now().Add(timeOut))
		if err != nil {
			log.Fatal(err)
		}
		buffer := make([]byte, 32*64)
		n, addr, err := packetConn.ReadFrom(buffer)
		if err != nil {
			if nerr, ok := err.(net.Error); ok {
				if nerr.Timeout() {
					break
				}
			}
			log.Fatal(err)
		}
		if addr.String() != conn.LocalAddr().String() {
			log.Fatal("hijacked by", addr.String())
		}
		messagesReceived++
		dataRead += n
	}
	dur := time.Now().Sub(startTime) - timeOut
	fmt.Println("Received", messagesReceived, "messages with total size of", dataRead, "bytes in", dur.Seconds(), "seconds")
}
