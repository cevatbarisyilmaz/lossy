package lossy

import (
	"math/rand"
	"net"
	"sync"
	"time"
)

type packetConn struct {
	net.PacketConn
	minLatency     time.Duration
	maxLatency     time.Duration
	packetLossRate float64
	writeDeadline  time.Time
	closed         bool
	mu             *sync.Mutex
	rand           *rand.Rand
}

// PacketConn wraps the given net.PacketConn and applies latency and packet losses to the written packets
func PacketConn(c net.PacketConn, minLatency, maxLatency time.Duration, packetLossRate float64) net.PacketConn {
	return &packetConn{
		PacketConn:     c,
		minLatency:     minLatency,
		maxLatency:     maxLatency,
		packetLossRate: packetLossRate,
		writeDeadline:  time.Time{},
		closed:         false,
		mu:             &sync.Mutex{},
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *packetConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.writeDeadline.Equal(time.Time{}) && c.writeDeadline.After(time.Now()) {
		return c.PacketConn.WriteTo(p, addr)
	}
	go func() {
		if c.rand.Float64() > c.packetLossRate {
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
