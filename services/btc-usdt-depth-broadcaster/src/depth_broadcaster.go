package main

import (
	"sync"
	"time"
)

type RingBufferDepthSnapshot struct {
	Sequence  uint64
	Timestamp time.Time
	Bids      [][2]uint64 // [priceE8, qtyE8]
	Asks      [][2]uint64
}

// DepthBroadcaster uses a pre-allocated fixed-size ring buffer for zero-allocation broadcasts
type DepthBroadcaster struct {
	mu        sync.RWMutex
	ring      []RingBufferDepthSnapshot
	capacity  int
	writeHead int
	seq       uint64
}

func NewDepthBroadcaster(capacity int) *DepthBroadcaster {
	if capacity <= 0 {
		capacity = 1024
	}
	return &DepthBroadcaster{
		ring:     make([]RingBufferDepthSnapshot, capacity),
		capacity: capacity,
	}
}

// PushSnapshot places latest L2 depth snapshot into circular ring buffer
func (b *DepthBroadcaster) PushSnapshot(bids, asks [][2]uint64) uint64 {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.seq++
	idx := b.writeHead % b.capacity
	b.ring[idx] = RingBufferDepthSnapshot{
		Sequence:  b.seq,
		Timestamp: time.Now().UTC(),
		Bids:      bids,
		Asks:      asks,
	}
	b.writeHead++
	return b.seq
}

// GetLatest reads the most recent depth snapshot
func (b *DepthBroadcaster) GetLatest() (RingBufferDepthSnapshot, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.writeHead == 0 {
		return RingBufferDepthSnapshot{}, false
	}
	idx := (b.writeHead - 1) % b.capacity
	return b.ring[idx], true
}
