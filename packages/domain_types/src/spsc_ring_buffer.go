package src

import (
	"errors"
	"sync/atomic"
)

var (
	ErrQueueFull  = errors.New("spsc: ring buffer is full")
	ErrQueueEmpty = errors.New("spsc: ring buffer is empty")
)

// SPSCRingBuffer implements an ultra-low-latency lock-free Single-Producer Single-Consumer circular queue
// optimized with power-of-two masking and cache line isolation.
type SPSCRingBuffer[T any] struct {
	// Head pointer (updated only by consumer)
	head uint64
	_pad0 [7]uint64 // 56 bytes padding to isolate 64-byte cache line

	// Tail pointer (updated only by producer)
	tail uint64
	_pad1 [7]uint64 // 56 bytes padding to isolate 64-byte cache line

	capacity uint64
	mask     uint64
	buffer   []T
}

// NewSPSCRingBuffer creates a new SPSC ring buffer with power-of-2 capacity.
func NewSPSCRingBuffer[T any](minCapacity uint64) *SPSCRingBuffer[T] {
	// Round up to nearest power of 2
	cap := uint64(1)
	for cap < minCapacity {
		cap <<= 1
	}
	if cap < 2 {
		cap = 2
	}

	return &SPSCRingBuffer[T]{
		capacity: cap,
		mask:     cap - 1,
		buffer:   make([]T, cap),
	}
}

// Enqueue inserts an item at the tail. Lock-free, called ONLY by the single producer thread.
func (q *SPSCRingBuffer[T]) Enqueue(item T) error {
	tail := atomic.LoadUint64(&q.tail)
	head := atomic.LoadUint64(&q.head)

	// If buffer is full
	if tail-head >= q.capacity {
		return ErrQueueFull
	}

	slot := tail & q.mask
	q.buffer[slot] = item
	atomic.StoreUint64(&q.tail, tail+1)
	return nil
}

// Dequeue extracts an item from head. Lock-free, called ONLY by the single consumer thread.
func (q *SPSCRingBuffer[T]) Dequeue() (T, error) {
	head := atomic.LoadUint64(&q.head)
	tail := atomic.LoadUint64(&q.tail)

	var zero T
	if head >= tail {
		return zero, ErrQueueEmpty
	}

	slot := head & q.mask
	item := q.buffer[slot]
	// Clear slot for GC
	q.buffer[slot] = zero
	atomic.StoreUint64(&q.head, head+1)
	return item, nil
}

// Len returns current number of unconsumed elements.
func (q *SPSCRingBuffer[T]) Len() uint64 {
	tail := atomic.LoadUint64(&q.tail)
	head := atomic.LoadUint64(&q.head)
	if tail < head {
		return 0
	}
	return tail - head
}

// Capacity returns the total capacity of the buffer.
func (q *SPSCRingBuffer[T]) Capacity() uint64 {
	return q.capacity
}
