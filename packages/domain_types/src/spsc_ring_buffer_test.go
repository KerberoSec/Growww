package src

import (
	"errors"
	"sync"
	"testing"
)

type TradeTick struct {
	SeqID   uint64
	PriceE8 uint64
	QtyE8   uint64
}

func TestSPSCRingBufferBasicOperations(t *testing.T) {
	q := NewSPSCRingBuffer[TradeTick](4)
	if q.Capacity() != 4 {
		t.Fatalf("expected capacity 4, got %d", q.Capacity())
	}

	// Dequeue from empty
	_, err := q.Dequeue()
	if err == nil || !errors.Is(err, ErrQueueEmpty) {
		t.Fatalf("expected ErrQueueEmpty, got %v", err)
	}

	// Fill to capacity
	for i := uint64(1); i <= 4; i++ {
		err := q.Enqueue(TradeTick{SeqID: i, PriceE8: 100 * i, QtyE8: 10 * i})
		if err != nil {
			t.Fatalf("enqueue failed at item %d: %v", i, err)
		}
	}

	if q.Len() != 4 {
		t.Fatalf("expected len 4, got %d", q.Len())
	}

	// Enqueue when full must error
	err = q.Enqueue(TradeTick{SeqID: 5})
	if err == nil || !errors.Is(err, ErrQueueFull) {
		t.Fatalf("expected ErrQueueFull, got %v", err)
	}

	// Dequeue items in FIFO order
	for i := uint64(1); i <= 4; i++ {
		item, err := q.Dequeue()
		if err != nil {
			t.Fatalf("dequeue failed at item %d: %v", i, err)
		}
		if item.SeqID != i || item.PriceE8 != 100*i {
			t.Fatalf("item mismatch: expected seq %d, got %d", i, item.SeqID)
		}
	}

	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got len %d", q.Len())
	}
}

func TestSPSCRingBufferConcurrentProducerConsumer(t *testing.T) {
	q := NewSPSCRingBuffer[uint64](1024)
	totalItems := uint64(100000)

	var wg sync.WaitGroup
	wg.Add(2)

	// Producer
	go func() {
		defer wg.Done()
		for i := uint64(0); i < totalItems; i++ {
			for {
				err := q.Enqueue(i)
				if err == nil {
					break
				}
				// yield/spin if full
			}
		}
	}()

	// Consumer
	receivedCount := uint64(0)
	lastVal := uint64(0)
	go func() {
		defer wg.Done()
		for receivedCount < totalItems {
			val, err := q.Dequeue()
			if err == nil {
				if receivedCount > 0 && val != lastVal+1 {
					t.Errorf("sequence gap: expected %d, got %d", lastVal+1, val)
				}
				lastVal = val
				receivedCount++
			}
		}
	}()

	wg.Wait()

	if receivedCount != totalItems {
		t.Fatalf("expected %d items processed, got %d", totalItems, receivedCount)
	}
}
