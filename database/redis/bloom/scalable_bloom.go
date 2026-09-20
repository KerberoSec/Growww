package bloom

import (
	"sync"
)

// ScalableBloomFilter dynamically allocates new filter layers as elements are inserted,
// keeping false positive probability bounded by p using tightening factor r (default 0.85).
type ScalableBloomFilter struct {
	mu                 sync.RWMutex
	initialCapacity    uint64
	targetFP           float64
	tighteningRatio    float64
	scaleMultiplier    uint64
	filters            []*StandardBloomFilter
	capacities         []uint64
	totalElements      uint64
}

// NewScalableBloomFilter creates a dynamically scalable Bloom filter.
func NewScalableBloomFilter(initialCapacity uint64, targetFP float64) *ScalableBloomFilter {
	if initialCapacity == 0 {
		initialCapacity = 10000
	}
	if targetFP <= 0 || targetFP >= 1.0 {
		targetFP = 0.001
	}

	tightening := 0.85
	firstFP := targetFP * (1.0 - tightening)

	firstFilter := NewStandardBloomFilter(initialCapacity, firstFP)

	return &ScalableBloomFilter{
		initialCapacity: initialCapacity,
		targetFP:        targetFP,
		tighteningRatio: tightening,
		scaleMultiplier: 2,
		filters:         []*StandardBloomFilter{firstFilter},
		capacities:      []uint64{initialCapacity},
	}
}

// Add inserts an element into the active layer, creating a new layer if capacity is reached.
func (s *ScalableBloomFilter) Add(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	activeIdx := len(s.filters) - 1
	activeFilter := s.filters[activeIdx]

	if activeFilter.ElementCount() >= s.capacities[activeIdx] {
		// Layer saturated: allocate new layer with scaled capacity and tightened error rate
		newCap := s.capacities[activeIdx] * s.scaleMultiplier
		// Tighten FP rate: p_i = p * (1-r) * r^i
		newFP := s.targetFP * (1.0 - s.tighteningRatio)
		for i := 0; i < len(s.filters); i++ {
			newFP *= s.tighteningRatio
		}

		newFilter := NewStandardBloomFilter(newCap, newFP)
		s.filters = append(s.filters, newFilter)
		s.capacities = append(s.capacities, newCap)
		activeFilter = newFilter
	}

	activeFilter.Add(data)
	s.totalElements++
}

// Contains checks all filter layers for element presence.
func (s *ScalableBloomFilter) Contains(data []byte) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, filter := range s.filters {
		if filter.Contains(data) {
			return true
		}
	}
	return false
}

// LayerCount returns the number of stacked layers.
func (s *ScalableBloomFilter) LayerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.filters)
}

// TotalElements returns total count across all layers.
func (s *ScalableBloomFilter) TotalElements() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalElements
}
