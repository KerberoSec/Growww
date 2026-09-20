package ordermatching

import (
	"errors"
	"sync/atomic"
)

var (
	ErrQueueFull  = errors.New("spsc queue is full")
	ErrQueueEmpty = errors.New("spsc queue is empty")
)

// CacheLinePadding prevents false sharing across CPU L1/L2/L3 cache lines (64 bytes x86/ARM)
type CacheLinePadding [64]byte

// SequencedItem represents a tagged payload with a continuous monotonic sequence number
type SequencedItem[T any] struct {
	SequenceID uint64
	TimestampNs int64
	Payload     T
}

// SPSCQueue is a high-throughput, lock-free Single-Producer Single-Consumer circular ring buffer.
// Employs cache-line padding between producer (tail) and consumer (head) indices to eliminate false sharing.
type SPSCQueue[T any] struct {
	_pad0 CacheLinePadding

	// tail is updated only by the Producer
	tail atomic.Uint64

	_pad1 CacheLinePadding

	// head is updated only by the Consumer
	head atomic.Uint64

	_pad2 CacheLinePadding

	buffer   []T
	capacity uint64
	mask     uint64
}

// NewSPSCQueue creates an SPSC queue with a power-of-two buffer capacity
func NewSPSCQueue[T any](minCapacity uint64) *SPSCQueue[T] {
	cap := uint64(1)
	for cap < minCapacity {
		cap <<= 1
	}

	return &SPSCQueue[T]{
		buffer:   make([]T, cap),
		capacity: cap,
		mask:     cap - 1,
	}
}

// Capacity returns the total allocated capacity
func (q *SPSCQueue[T]) Capacity() uint64 {
	return q.capacity
}

// Length returns the current number of available items in the queue
func (q *SPSCQueue[T]) Length() uint64 {
	tail := q.tail.Load()
	head := q.head.Load()
	if tail >= head {
		return tail - head
	}
	return 0
}

// TryPush attempts to enqueue an item without blocking.
// MUST only be invoked by the single designated Producer goroutine.
func (q *SPSCQueue[T]) TryPush(item T) bool {
	tail := q.tail.Load()
	head := q.head.Load()

	// Full condition: tail - head >= capacity
	if tail-head >= q.capacity {
		return false
	}

	q.buffer[tail&q.mask] = item
	// Store with release semantics to ensure item payload write is visible before tail increments
	q.tail.Store(tail + 1)
	return true
}

// Push enqueues an item or returns ErrQueueFull
func (q *SPSCQueue[T]) Push(item T) error {
	if !q.TryPush(item) {
		return ErrQueueFull
	}
	return nil
}

// TryPop attempts to dequeue an item without blocking.
// MUST only be invoked by the single designated Consumer goroutine.
func (q *SPSCQueue[T]) TryPop() (T, bool) {
	var zero T
	head := q.head.Load()
	tail := q.tail.Load()

	// Empty condition: head == tail
	if head >= tail {
		return zero, false
	}

	index := head & q.mask
	item := q.buffer[index]
	q.buffer[index] = zero // clear reference for garbage collection

	// Store with release semantics
	q.head.Store(head + 1)
	return item, true
}

// Pop dequeues an item or returns ErrQueueEmpty
func (q *SPSCQueue[T]) Pop() (T, error) {
	item, ok := q.TryPop()
	if !ok {
		var zero T
		return zero, ErrQueueEmpty
	}
	return item, nil
}

// BatchPop drains up to maxItems from the queue into the provided destination slice.
// Returns the number of items successfully drained.
func (q *SPSCQueue[T]) BatchPop(dest []T) int {
	if len(dest) == 0 {
		return 0
	}

	head := q.head.Load()
	tail := q.tail.Load()

	if head >= tail {
		return 0
	}

	available := tail - head
	count := uint64(len(dest))
	if count > available {
		count = available
	}

	var zero T
	for i := uint64(0); i < count; i++ {
		idx := (head + i) & q.mask
		dest[i] = q.buffer[idx]
		q.buffer[idx] = zero
	}

	q.head.Store(head + count)
	return int(count)
}

// Sequencer provides strict monotonic ordering for matching engine event streams
type Sequencer[T any] struct {
	queue        *SPSCQueue[SequencedItem[T]]
	sequenceCounter atomic.Uint64
}

// NewSequencer constructs a new monotonic SPSC sequencer
func NewSequencer[T any](capacity uint64) *Sequencer[T] {
	return &Sequencer[T]{
		queue: NewSPSCQueue[SequencedItem[T]](capacity),
	}
}

// SequenceAndEnqueue stamps the payload with next contiguous monotonic ID and pushes into the SPSC queue
func (s *Sequencer[T]) SequenceAndEnqueue(payload T, timestampNs int64) (uint64, error) {
	seq := s.sequenceCounter.Add(1)
	item := SequencedItem[T]{
		SequenceID:  seq,
		TimestampNs: timestampNs,
		Payload:     payload,
	}

	if err := s.queue.Push(item); err != nil {
		return 0, err
	}

	return seq, nil
}

// PollNext retrieves the next sequentially ordered item from the queue
func (s *Sequencer[T]) PollNext() (SequencedItem[T], bool) {
	return s.queue.TryPop()
}

// DrainBatch extracts a batch of ordered items for high-throughput batch execution
func (s *Sequencer[T]) DrainBatch(dest []SequencedItem[T]) int {
	return s.queue.BatchPop(dest)
}

// CurrentSequence returns the latest issued sequence number
func (s *Sequencer[T]) CurrentSequence() uint64 {
	return s.sequenceCounter.Load()
}
