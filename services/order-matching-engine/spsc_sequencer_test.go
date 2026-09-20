package ordermatching

import (
	"sync"
	"testing"
)

func TestSPSCQueue_BasicFIFO(t *testing.T) {
	q := NewSPSCQueue[int](16)
	if q.Capacity() != 16 {
		t.Fatalf("expected capacity 16, got %d", q.Capacity())
	}

	for i := 1; i <= 10; i++ {
		if err := q.Push(i); err != nil {
			t.Fatalf("failed to push %d: %v", i, err)
		}
	}

	if q.Length() != 10 {
		t.Fatalf("expected length 10, got %d", q.Length())
	}

	for i := 1; i <= 10; i++ {
		val, err := q.Pop()
		if err != nil {
			t.Fatalf("failed to pop item %d: %v", i, err)
		}
		if val != i {
			t.Fatalf("expected %d, got %d", i, val)
		}
	}

	// Queue should now be empty
	if _, err := q.Pop(); err != ErrQueueEmpty {
		t.Fatalf("expected ErrQueueEmpty, got %v", err)
	}
}

func TestSPSCQueue_CapacityBoundaries(t *testing.T) {
	q := NewSPSCQueue[string](4)
	if q.Capacity() != 4 {
		t.Fatalf("expected capacity 4, got %d", q.Capacity())
	}

	for i := 0; i < 4; i++ {
		if !q.TryPush("item") {
			t.Fatalf("failed to push item %d", i)
		}
	}

	// Next push should fail
	if q.TryPush("overflow") {
		t.Fatalf("expected TryPush to return false when queue is full")
	}

	// Pop 2 items
	q.TryPop()
	q.TryPop()

	// Should allow 2 more items
	if !q.TryPush("item_new_1") || !q.TryPush("item_new_2") {
		t.Fatalf("expected TryPush to succeed after pop")
	}
	if q.TryPush("overflow_again") {
		t.Fatalf("expected TryPush to fail when queue is full again")
	}
}

func TestSPSCQueue_BatchPop(t *testing.T) {
	q := NewSPSCQueue[int](32)
	for i := 0; i < 20; i++ {
		q.Push(i * 10)
	}

	batch := make([]int, 8)
	n1 := q.BatchPop(batch)
	if n1 != 8 {
		t.Fatalf("expected 8 items, got %d", n1)
	}
	for i := 0; i < 8; i++ {
		if batch[i] != i*10 {
			t.Fatalf("expected %d, got %d", i*10, batch[i])
		}
	}

	// Drain remaining
	remaining := make([]int, 32)
	n2 := q.BatchPop(remaining)
	if n2 != 12 {
		t.Fatalf("expected 12 remaining items, got %d", n2)
	}

	// Queue should be empty now
	if q.Length() != 0 {
		t.Fatalf("expected queue length 0, got %d", q.Length())
	}
}

func TestSPSCQueue_ConcurrentProducerConsumer(t *testing.T) {
	const totalItems = 200_000
	q := NewSPSCQueue[uint64](4096)

	var wg sync.WaitGroup
	wg.Add(2)

	// Producer Goroutine
	go func() {
		defer wg.Done()
		for i := uint64(1); i <= totalItems; i++ {
			for !q.TryPush(i) {
				// Spin wait
			}
		}
	}()

	receivedItems := make([]uint64, 0, totalItems)

	// Consumer Goroutine
	go func() {
		defer wg.Done()
		batch := make([]uint64, 64)
		for uint64(len(receivedItems)) < totalItems {
			n := q.BatchPop(batch)
			if n > 0 {
				receivedItems = append(receivedItems, batch[:n]...)
			}
		}
	}()

	wg.Wait()

	if len(receivedItems) != totalItems {
		t.Fatalf("expected %d received items, got %d", totalItems, len(receivedItems))
	}

	// Verify strict monotonicity
	for i := 0; i < totalItems; i++ {
		expected := uint64(i + 1)
		if receivedItems[i] != expected {
			t.Fatalf("ordering mismatch at index %d: expected %d, got %d", i, expected, receivedItems[i])
		}
	}
}

func TestSequencer_MonotonicOrdering(t *testing.T) {
	seq := NewSequencer[string](64)

	id1, err := seq.SequenceAndEnqueue("ORDER_BUY_1", 1000)
	if err != nil || id1 != 1 {
		t.Fatalf("expected seq 1, got %d, err: %v", id1, err)
	}

	id2, err := seq.SequenceAndEnqueue("ORDER_SELL_2", 1005)
	if err != nil || id2 != 2 {
		t.Fatalf("expected seq 2, got %d, err: %v", id2, err)
	}

	item1, ok := seq.PollNext()
	if !ok || item1.SequenceID != 1 || item1.Payload != "ORDER_BUY_1" {
		t.Fatalf("unexpected item 1: %+v", item1)
	}

	item2, ok := seq.PollNext()
	if !ok || item2.SequenceID != 2 || item2.Payload != "ORDER_SELL_2" {
		t.Fatalf("unexpected item 2: %+v", item2)
	}

	if _, ok := seq.PollNext(); ok {
		t.Fatalf("expected empty sequencer")
	}
}
