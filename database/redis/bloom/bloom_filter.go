package bloom

import (
	"sync"
)

// StandardBloomFilter implements a high-performance, thread-safe bitset Bloom filter.
type StandardBloomFilter struct {
	mu           sync.RWMutex
	m            uint64   // bit array size
	k            uint32   // number of hash functions
	bits         []uint64 // 64-bit word storage
	elementCount uint64   // count of added elements
}

// NewStandardBloomFilter initializes a filter for expectedElements with falsePositiveRate.
func NewStandardBloomFilter(expectedElements uint64, falsePositiveRate float64) *StandardBloomFilter {
	m, k := CalculateOptimalParams(expectedElements, falsePositiveRate)
	words := (m + 63) / 64
	return &StandardBloomFilter{
		m:            m,
		k:            k,
		bits:         make([]uint64, words),
		elementCount: 0,
	}
}

// Add inserts an element into the Bloom filter.
func (b *StandardBloomFilter) Add(data []byte) {
	indices := HashValues(data, b.m, b.k)
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, idx := range indices {
		wordIdx := idx / 64
		bitIdx := idx % 64
		b.bits[wordIdx] |= (1 << bitIdx)
	}
	b.elementCount++
}

// Contains tests whether an element is probably in the filter (guarantees zero false negatives).
func (b *StandardBloomFilter) Contains(data []byte) bool {
	indices := HashValues(data, b.m, b.k)
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, idx := range indices {
		wordIdx := idx / 64
		bitIdx := idx % 64
		if (b.bits[wordIdx] & (1 << bitIdx)) == 0 {
			return false // Definitive negative
		}
	}
	return true // Probable positive
}

// TestAndSet atomically tests for membership and adds if absent.
// Returns true if the element was ALREADY present (probable duplicate), false if newly added.
func (b *StandardBloomFilter) TestAndSet(data []byte) bool {
	indices := HashValues(data, b.m, b.k)
	b.mu.Lock()
	defer b.mu.Unlock()

	alreadyPresent := true
	for _, idx := range indices {
		wordIdx := idx / 64
		bitIdx := idx % 64
		if (b.bits[wordIdx] & (1 << bitIdx)) == 0 {
			alreadyPresent = false
		}
		b.bits[wordIdx] |= (1 << bitIdx)
	}

	if !alreadyPresent {
		b.elementCount++
	}

	return alreadyPresent
}

// ElementCount returns count of added items.
func (b *StandardBloomFilter) ElementCount() uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.elementCount
}

// BitSize returns m.
func (b *StandardBloomFilter) BitSize() uint64 {
	return b.m
}

// HashCount returns k.
func (b *StandardBloomFilter) HashCount() uint32 {
	return b.k
}
