package lossy_test

import (
	"bytes"
	"fmt"
	"github.com/cevatbarisyilmaz/lossy"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"
)

func getConnections() (net.PacketConn, net.Conn, error) {
	packetConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 0,
	})
	if err != nil {
		return nil, nil, err
	}
	conn, err := net.DialUDP("udp", nil, packetConn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		return nil, nil, err
	}
	return packetConn, conn, nil
}

func closeConnections(packetConn net.PacketConn, conn net.Conn) error {
	err := conn.Close()
	cerr := packetConn.Close()
	if cerr != nil && err == nil {
		err = cerr
	}
	return err
}

func TestBandwidth(t *testing.T) {
	t.Parallel()
	const bandwidth = 8 * 1024 // 64 Kbit/s
	const headerOverhead = lossy.UDPv4MinHeaderOverhead
	const messageCount = 32
	const messageSize = 1024
	const idealDuration = time.Second * messageCount * (messageSize + headerOverhead) / bandwidth
	const testSensitivity = time.Millisecond * 100
	pc, c, err := getConnections()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = closeConnections(pc, c)
		if err != nil {
			t.Error(err)
		}
	}()
	c = lossy.Conn(c, bandwidth, 0, 0, 0, headerOverhead)
	pc = lossy.PacketConn(pc, bandwidth, 0, 0, 0, headerOverhead)
	var messages [][]byte
	for i := 0; i < messageCount; i++ {
		message := make([]byte, messageSize)
		rand.Read(message)
		messages = append(messages, message)
	}
	startTime := time.Now()
	go func() {
		for _, message := range messages {
			_, err := c.Write(message)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Millisecond) // Mke sure messages will arrive in order
		}
	}()
	go func() {
		for _, message := range messages {
			_, err := pc.WriteTo(message, c.LocalAddr())
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(time.Millisecond) // Make sure messages will arrive in order
		}
	}()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < messageCount; i++ {
			buffer := make([]byte, messageSize)
			n, err := c.Read(buffer)
			if err != nil {
				t.Fatal(err)
			}
			if n != messageSize {
				t.Fatal("message is smaller than expected")
			}
			if !bytes.Equal(buffer, messages[i]) {
				t.Fatal("wrong message is received")
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < messageCount; i++ {
			buffer := make([]byte, messageSize)
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				t.Fatal(err)
			}
			if addr.String() != c.LocalAddr().String() {
				t.Fatal("hijacked")
			}
			if n != messageSize {
				t.Fatal("message is smaller than expected")
			}
			if !bytes.Equal(buffer, messages[i]) {
				t.Fatal("wrong message is received")
			}
		}
	}()
	wg.Wait()
	dur := time.Now().Sub(startTime)
	if dur < idealDuration-testSensitivity {
		t.Error("transmission took shorter than expected by", idealDuration-dur)
	}
	if dur > idealDuration+testSensitivity {
		t.Error("transmission took longer than expected by", dur-idealDuration)
	}
}

func TestLatency(t *testing.T) {
	t.Parallel()
	const minLatency = time.Millisecond * 10
	const maxLatency = time.Millisecond * 100
	const testSensitivityForMin = time.Millisecond
	const testSensitivityForMax = time.Millisecond * 5
	const messageCount = 32
	const messageSize = 1024
	pc, c, err := getConnections()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = closeConnections(pc, c)
		if err != nil {
			t.Error(err)
		}
	}()
	c = lossy.Conn(c, 0, minLatency, maxLatency, 0, 0)
	pc = lossy.PacketConn(pc, 0, minLatency, maxLatency, 0, 0)
	var messages [][]byte
	for i := 0; i < messageCount; i++ {
		message := make([]byte, messageSize)
		rand.Read(message)
		messages = append(messages, message)
	}
	c1 := make(chan time.Time, messageCount)
	c2 := make(chan time.Time, messageCount)
	go func() {
		for _, message := range messages {
			_, err := c.Write(message)
			if err != nil {
				t.Fatal(err)
			}
			c1 <- time.Now()
			time.Sleep(maxLatency + time.Millisecond) // Make sure messages will arrive in order
		}
	}()
	go func() {
		for _, message := range messages {
			_, err := pc.WriteTo(message, c.LocalAddr())
			if err != nil {
				t.Fatal(err)
			}
			c2 <- time.Now()
			time.Sleep(maxLatency + time.Millisecond) // Make sure messages will arrive in order
		}
	}()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < messageCount; i++ {
			buffer := make([]byte, messageSize)
			n, err := c.Read(buffer)
			if err != nil {
				t.Fatal(err)
			}
			if n != messageSize {
				t.Fatal("message is smaller than expected")
			}
			if !bytes.Equal(buffer, messages[i]) {
				fmt.Println(i)
				t.Fatal("wrong message is received")
			}
			latency := time.Now().Sub(<-c2)
			if latency < minLatency-testSensitivityForMin {
				t.Error("latency is smaller than expected by", minLatency-latency)
			} else if latency > maxLatency+testSensitivityForMax {
				t.Error("latency is greater than expected", latency-maxLatency)
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < messageCount; i++ {
			buffer := make([]byte, messageSize)
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				t.Fatal(err)
			}
			if addr.String() != c.LocalAddr().String() {
				t.Fatal("hijacked")
			}
			if n != messageSize {
				t.Fatal("message is smaller than expected")
			}
			if !bytes.Equal(buffer, messages[i]) {
				fmt.Println(i)
				t.Fatal("wrong message is received")
			}
			latency := time.Now().Sub(<-c1)
			if latency < minLatency-testSensitivityForMin {
				t.Error("latency is smaller than expected by", minLatency-latency)
			} else if latency > maxLatency+testSensitivityForMax {
				t.Error("latency is greater than expected", latency-maxLatency)
			}
		}
	}()
	wg.Wait()
}

func TestPacketLoss(t *testing.T) {
	t.Parallel()
	const packetLossRate = 0.33
	const messageCount = 16 * 1024
	const messageSize = 16
	const testSensitivity = 0.03
	pc, c, err := getConnections()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = closeConnections(pc, c)
		if err != nil {
			t.Error(err)
		}
	}()
	c = lossy.Conn(c, 0, 0, 0, packetLossRate, 0)
	pc = lossy.PacketConn(pc, 0, 0, 0, packetLossRate, 0)
	var messages [][]byte
	for i := 0; i < messageCount; i++ {
		message := make([]byte, messageSize)
		rand.Read(message)
		messages = append(messages, message)
	}
	go func() {
		for _, message := range messages {
			_, err := c.Write(message)
			if err != nil {
				t.Fatal(err)
			}
		}
	}()
	go func() {
		for _, message := range messages {
			_, err := pc.WriteTo(message, c.LocalAddr())
			if err != nil {
				t.Fatal(err)
			}
		}
	}()
	readDeadline := time.Now().Add(messageCount * time.Millisecond)
	err = pc.SetReadDeadline(readDeadline)
	if err != nil {
		t.Fatal(err)
	}
	err = c.SetReadDeadline(readDeadline)
	if err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	var a, b int
	go func() {
		defer wg.Done()
		for {
			buffer := make([]byte, messageSize)
			n, err := c.Read(buffer)
			if err != nil {
				if nerr, ok := err.(net.Error); ok {
					if nerr.Timeout() {
						return
					}
				}
				t.Fatal(err)
			}
			if n != messageSize {
				t.Fatal("message is smaller than expected")
			}
			a++
		}
	}()
	go func() {
		defer wg.Done()
		for {
			buffer := make([]byte, messageSize)
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				if nerr, ok := err.(net.Error); ok {
					if nerr.Timeout() {
						return
					}
				}
				t.Fatal(err)
			}
			if addr.String() != c.LocalAddr().String() {
				t.Fatal("hijacked")
			}
			if n != messageSize {
				t.Fatal("message is smaller than expected")
			}
			b++
		}
	}()
	wg.Wait()
	rate := 1 - (float64(a+b) / (2 * messageCount))
	if rate > packetLossRate+testSensitivity {
		t.Error("packet loss rate is greater than expected by", rate-packetLossRate)
	} else if rate < packetLossRate-testSensitivity {
		t.Error("packet loss rate is smaller than expected by", packetLossRate-rate)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	pc, c, err := getConnections()
	if err != nil {
		t.Fatal(err)
	}
	pc = lossy.PacketConn(pc, 0, 0, 0, 0, 0)
	c = lossy.Conn(c, 0, 0, 0, 0, 0)

	err = c.SetWriteDeadline(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	_, err = c.Write([]byte{1})
	if err == nil {
		t.Error("expected error did not occur")
	}
	err = c.SetDeadline(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	_, err = c.Write([]byte{1})
	if err == nil {
		t.Error("expected error did not occur")
	}
	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.Write([]byte{1})
	if err == nil {
		t.Error("expected error did not occur")
	}

	err = pc.SetWriteDeadline(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	_, err = pc.WriteTo([]byte{1}, c.LocalAddr())
	if err == nil {
		t.Error("expected error did not occur")
	}
	err = pc.SetDeadline(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	_, err = pc.WriteTo([]byte{1}, c.LocalAddr())
	if err == nil {
		t.Error("expected error did not occur")
	}
	err = pc.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = pc.WriteTo([]byte{1}, c.LocalAddr())
	if err == nil {
		t.Error("expected error did not occur")
	}
}
