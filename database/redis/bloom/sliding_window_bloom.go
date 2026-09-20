package bloom

import (
	"sync"
	"time"
)

// RotationalSlidingWindowBloomFilter maintains W time-sliced Bloom filters.
// Oldest slice is evicted automatically as time progresses, keeping memory bounded.
type RotationalSlidingWindowBloomFilter struct {
	mu           sync.RWMutex
	buckets      []*StandardBloomFilter
	bucketCount  int
	windowSize   time.Duration
	bucketWidth  time.Duration
	lastRotate   time.Time
	capacityPerB uint64
	targetFP     float64
}

// NewRotationalSlidingWindowBloomFilter creates a sliding window Bloom filter.
// e.g. windowSize = 5 minutes, bucketCount = 5 -> each bucket is 1 minute.
func NewRotationalSlidingWindowBloomFilter(windowSize time.Duration, bucketCount int, capacityPerBucket uint64, targetFP float64) *RotationalSlidingWindowBloomFilter {
	if bucketCount < 2 {
		bucketCount = 2
	}
	bucketWidth := windowSize / time.Duration(bucketCount)

	buckets := make([]*StandardBloomFilter, bucketCount)
	for i := 0; i < bucketCount; i++ {
		buckets[i] = NewStandardBloomFilter(capacityPerBucket, targetFP)
	}

	return &RotationalSlidingWindowBloomFilter{
		buckets:      buckets,
		bucketCount:  bucketCount,
		windowSize:   windowSize,
		bucketWidth:  bucketWidth,
		lastRotate:   time.Now(),
		capacityPerB: capacityPerBucket,
		targetFP:     targetFP,
	}
}

// RotateIfNeeded slides the window if time elapsed > bucketWidth.
func (r *RotationalSlidingWindowBloomFilter) RotateIfNeeded(now time.Time) {
	if now.Sub(r.lastRotate) >= r.bucketWidth {
		r.mu.Lock()
		defer r.mu.Unlock()

		steps := int(now.Sub(r.lastRotate) / r.bucketWidth)
		if steps > r.bucketCount {
			steps = r.bucketCount
		}

		for s := 0; s < steps; s++ {
			// Drop oldest (index 0) and append fresh bucket
			r.buckets = r.buckets[1:]
			r.buckets = append(r.buckets, NewStandardBloomFilter(r.capacityPerB, r.targetFP))
		}
		r.lastRotate = now
	}
}

// Add inserts into the newest active bucket (latest index).
func (r *RotationalSlidingWindowBloomFilter) Add(data []byte, now time.Time) {
	r.RotateIfNeeded(now)

	r.mu.RLock()
	defer r.mu.RUnlock()

	activeBucket := r.buckets[len(r.buckets)-1]
	activeBucket.Add(data)
}

// Contains checks all buckets within the active sliding window.
func (r *RotationalSlidingWindowBloomFilter) Contains(data []byte, now time.Time) bool {
	r.RotateIfNeeded(now)

	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, b := range r.buckets {
		if b.Contains(data) {
			return true
		}
	}
	return false
}

// TestAndSet checks if present in ANY bucket; if not, inserts into newest bucket.
func (r *RotationalSlidingWindowBloomFilter) TestAndSet(data []byte, now time.Time) bool {
	r.RotateIfNeeded(now)

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, b := range r.buckets {
		if b.Contains(data) {
			return true // duplicate found in window
		}
	}

	// Add to newest bucket
	r.buckets[len(r.buckets)-1].Add(data)
	return false
}
