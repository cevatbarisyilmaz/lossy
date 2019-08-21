package lossy

import (
	"math/rand"
	"net"
	"sync"
	"time"
)

type conn struct {
	net.Conn
	minLatency     time.Duration
	maxLatency     time.Duration
	packetLossRate float64
	writeDeadline  time.Time
	closed         bool
	mu             *sync.Mutex
}

// Conn wraps the given net.Conn and applies latency and packet losses to the written packets
func Conn(c net.Conn, minLatency, maxLatency time.Duration, packetLossRate float64) net.Conn {
	return &conn{
		Conn:           c,
		minLatency:     minLatency,
		maxLatency:     maxLatency,
		packetLossRate: packetLossRate,
		writeDeadline:  time.Time{},
		closed:         false,
		mu:             &sync.Mutex{},
	}
}

func (c *conn) Write(b []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.writeDeadline.Equal(time.Time{}) && c.writeDeadline.After(time.Now()) {
		return c.Conn.Write(b)
	}
	go func() {
		if rand.Float64() > c.packetLossRate {
			time.Sleep(c.minLatency + time.Duration(float64(c.maxLatency-c.minLatency)*rand.Float64()))
			c.mu.Lock()
			_, _ = c.Conn.Write(b)
			c.mu.Unlock()
		}
	}()
	return len(b), nil
}

func (c *conn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return c.Conn.Close()
}

func (c *conn) SetDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = t
	return c.Conn.SetDeadline(t)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.writeDeadline = t
	return c.Conn.SetWriteDeadline(t)
}
