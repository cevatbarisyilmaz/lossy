package lossy_test

import (
	"github.com/cevatbarisyilmaz/lossy"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func Test(t *testing.T) {
	const packetLossRate = 0.25
	const minLatency = time.Millisecond * 10
	const maxLatency = time.Millisecond * 100
	udpPacketConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP: net.IPv4(127, 0, 0, 1),
		Port: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	udpConn, err := net.DialUDP("udp", nil, udpPacketConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatal(err)
	}
	packetConn := lossy.PacketConn(udpPacketConn, minLatency, maxLatency, packetLossRate)
	conn := lossy.Conn(udpConn, minLatency, maxLatency, packetLossRate)
	const maxByte = byte(255)
	send := map[byte]time.Time{}
	received := map[byte]time.Time{}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for index := byte(0); true; index++ {
			_, err := conn.Write([]byte{index})
			if err != nil {
				log.Fatal(err)
			}
			send[index] = time.Now()
			time.Sleep(minLatency)
			if index == maxByte {
				break
			}
		}
		wg.Done()
	}()
	go func() {
		wg.Wait()
		err := packetConn.SetDeadline(time.Now().Add(maxLatency * 2))
		if err != nil {
			t.Fatal(err)
		}
	}()
	for {
		buffer := make([]byte, 1)
		_, addr, err := packetConn.ReadFrom(buffer)
		if err != nil {
			if nerr, ok := err.(net.Error); ok {
				if nerr.Timeout() {
					break
				}
			}
			t.Fatal(err)
		}
		if addr.String() != conn.LocalAddr().String() {
			t.Fatal("address mismatch")
		}
		received[buffer[0]] = time.Now()
	}
	const latencySensitivity = time.Millisecond
	var dropped byte
	for i := byte(0); true; i++ {
		if received[i].Equal(time.Time{}) {
			dropped++
		} else {
			latency := received[i].Sub(send[i])
			if latency < minLatency - latencySensitivity {
				t.Error("latency is less than expected")
			} else if latency > maxLatency + latencySensitivity {
				t.Error("latency is more than expected")
			}
		}
		if i == maxByte {
			break
		}
	}
	const droppedSensitivity = packetLossRate * 0.25
	dropRate := float64(dropped) / float64(maxByte)
	if dropRate > packetLossRate + droppedSensitivity {
		t.Error("might be false negative but dropped packet rate is greater than expected by", dropRate - packetLossRate)
	} else if dropRate < packetLossRate - droppedSensitivity {
		t.Error("might be false negative but dropped packet rate is less than expected by", packetLossRate - dropRate)
	}

}