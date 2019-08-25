package lossy

import (
	"math/rand"
	"net"
	"sync"
	"time"
)

type packetConn struct {
	net.PacketConn
	minLatency        time.Duration
	maxLatency        time.Duration
	packetLossRate    float64
	writeDeadline     time.Time
	closed            bool
	mu                *sync.Mutex
	rand              *rand.Rand
	throttleMu        *sync.Mutex
	timeToWaitPerByte float64
	headerOverhead    int
}

// NewPacketConn wraps the given net.PacketConn with a lossy packet connection.
//
// bandwidth is in bytes/second.
// i.e. enter 1024 * 1024 for a 8 Mbit/s connection.
// Enter 0 or a negative value for an unlimited bandwidth.
//
// minLatency and maxLatency is used to create a random latency for each packet.
// maxLatency should be equal or greater than minLatency.
// If bandwidth is not unlimited and there's no other packets waiting to be delivered,
// time to deliver a packet is (len(packet) + headerOverhead) / bandwidth + randomDuration(minLatency, maxLatency)
//
// packetLossRate is chance of a packet to be dropped.
// It should be less than 1 and equal or greater than 0.
//
// headerOverhead is the header size of the underlying protocol of the connection.
// It is used to simulate bandwidth more realistically.
// If bandwidth is unlimited, headerOverhead is ignored.
func NewPacketConn(c net.PacketConn, bandwidth int, minLatency, maxLatency time.Duration, packetLossRate float64, headerOverhead int) net.PacketConn {
	var timeToWaitPerByte float64
	if bandwidth <= 0 {
		timeToWaitPerByte = 0
	} else {
		timeToWaitPerByte = float64(time.Second) / float64(bandwidth)
	}
	return &packetConn{
		PacketConn:        c,
		minLatency:        minLatency,
		maxLatency:        maxLatency,
		packetLossRate:    packetLossRate,
		writeDeadline:     time.Time{},
		closed:            false,
		mu:                &sync.Mutex{},
		rand:              rand.New(rand.NewSource(time.Now().UnixNano())),
		throttleMu:        &sync.Mutex{},
		timeToWaitPerByte: timeToWaitPerByte,
		headerOverhead:    headerOverhead,
	}
}

func (c *packetConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || !c.writeDeadline.Equal(time.Time{}) && c.writeDeadline.Before(time.Now()) {
		return c.PacketConn.WriteTo(p, addr)
	}
	go func() {
		c.throttleMu.Lock()
		time.Sleep(time.Duration(c.timeToWaitPerByte * (float64(len(p) + c.headerOverhead))))
		c.throttleMu.Unlock()
		if c.rand.Float64() >= c.packetLossRate {
			time.Sleep(c.minLatency + time.Duration(float64(c.maxLatency-c.minLatency)*c.rand.Float64()))
			c.mu.Lock()
			_, _ = c.PacketConn.WriteTo(p, addr)
			c.mu.Unlock()
		}
	}()
	return len(p), nil
}

func (c *packetConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return c.PacketConn.Close()
}

func (c *packetConn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = t
	return c.PacketConn.SetDeadline(t)
}

func (c *packetConn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = t
	return c.PacketConn.SetWriteDeadline(t)
}
