package bloom

import (
	"errors"
	"sync"
)

var (
	ErrElementNotPresent = errors.New("cannot remove element: not found in counting bloom filter")
)

// CountingBloomFilter uses 8-bit counters to support element deletion without rebuilding the filter.
type CountingBloomFilter struct {
	mu           sync.RWMutex
	m            uint64  // number of counters
	k            uint32  // number of hash functions
	counters     []uint8 // 8-bit counters (saturates at 255)
	elementCount uint64
}

// NewCountingBloomFilter creates a Counting Bloom filter.
func NewCountingBloomFilter(expectedElements uint64, falsePositiveRate float64) *CountingBloomFilter {
	m, k := CalculateOptimalParams(expectedElements, falsePositiveRate)
	return &CountingBloomFilter{
		m:            m,
		k:            k,
		counters:     make([]uint8, m),
		elementCount: 0,
	}
}

// Add increments the counters corresponding to the hash values.
func (c *CountingBloomFilter) Add(data []byte) {
	indices := HashValues(data, c.m, c.k)
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, idx := range indices {
		if c.counters[idx] < 255 {
			c.counters[idx]++
		}
	}
	c.elementCount++
}

// Contains checks if all corresponding counters are > 0.
func (c *CountingBloomFilter) Contains(data []byte) bool {
	indices := HashValues(data, c.m, c.k)
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, idx := range indices {
		if c.counters[idx] == 0 {
			return false
		}
	}
	return true
}

// Remove decrements the counters corresponding to the hash values.
func (c *CountingBloomFilter) Remove(data []byte) error {
	indices := HashValues(data, c.m, c.k)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify presence first
	for _, idx := range indices {
		if c.counters[idx] == 0 {
			return ErrElementNotPresent
		}
	}

	for _, idx := range indices {
		if c.counters[idx] > 0 && c.counters[idx] < 255 {
			c.counters[idx]--
		}
	}

	if c.elementCount > 0 {
		c.elementCount--
	}
	return nil
}

// ElementCount returns count of active elements.
func (c *CountingBloomFilter) ElementCount() uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.elementCount
}
