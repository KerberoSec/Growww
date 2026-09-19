package main

import (
	"testing"
	"time"
)

func TestDepthBroadcaster_PushAndRetrieve(t *testing.T) {
	b := NewDepthBroadcaster(16)

	bids := [][2]uint64{{6500000000000, 100000000}, {6499000000000, 200000000}}
	asks := [][2]uint64{{6501000000000, 150000000}, {6502000000000, 250000000}}

	seq := b.PushSnapshot(bids, asks)
	if seq != 1 {
		t.Fatalf("Expected sequence 1, got %d", seq)
	}

	snap, ok := b.GetLatest()
	if !ok {
		t.Fatal("Expected snapshot to exist")
	}
	if snap.Sequence != 1 {
		t.Fatalf("Expected sequence 1, got %d", snap.Sequence)
	}
	if len(snap.Bids) != 2 || len(snap.Asks) != 2 {
		t.Fatalf("Expected 2 bids and 2 asks, got %d/%d", len(snap.Bids), len(snap.Asks))
	}
}

func TestDepthBroadcaster_RingBufferOverwrite(t *testing.T) {
	b := NewDepthBroadcaster(4) // Small ring

	for i := 0; i < 10; i++ {
		b.PushSnapshot([][2]uint64{{uint64(i), 100}}, [][2]uint64{})
	}

	snap, ok := b.GetLatest()
	if !ok {
		t.Fatal("Expected snapshot to exist")
	}
	if snap.Sequence != 10 {
		t.Fatalf("Expected sequence 10, got %d", snap.Sequence)
	}
	// The ring should have overwritten earlier entries
	if snap.Bids[0][0] != 9 {
		t.Fatalf("Expected latest bid price=9, got %d", snap.Bids[0][0])
	}
}

func TestDepthBroadcaster_EmptyBuffer(t *testing.T) {
	b := NewDepthBroadcaster(8)

	_, ok := b.GetLatest()
	if ok {
		t.Fatal("Expected no snapshot from empty buffer")
	}
}

func TestConflationEngine_SubscribeUnsubscribe(t *testing.T) {
	b := NewDepthBroadcaster(16)
	engine := NewDepthConflationEngine(b, ConflationConfig{IntervalMs: 50, MaxDepthLevel: 5})

	ch := engine.Subscribe("ws-client-1")
	if engine.SubscriberCount() != 1 {
		t.Fatalf("Expected 1 subscriber, got %d", engine.SubscriberCount())
	}

	engine.Unsubscribe("ws-client-1")
	if engine.SubscriberCount() != 0 {
		t.Fatalf("Expected 0 subscribers after unsubscribe, got %d", engine.SubscriberCount())
	}
	// Channel should be closed
	_, open := <-ch
	if open {
		t.Fatal("Expected channel to be closed after unsubscribe")
	}
}

func TestConflationEngine_TruncatesDepthLevels(t *testing.T) {
	b := NewDepthBroadcaster(16)
	engine := NewDepthConflationEngine(b, ConflationConfig{IntervalMs: 0, MaxDepthLevel: 3})

	// Push 10 bid levels
	bids := make([][2]uint64, 10)
	for i := range bids {
		bids[i] = [2]uint64{uint64(6500 - i), 100}
	}
	b.PushSnapshot(bids, [][2]uint64{{6501, 50}})

	// Wait to clear any conflation window
	time.Sleep(5 * time.Millisecond)

	msg := engine.TryBroadcast()
	if msg == nil {
		t.Fatal("Expected broadcast message")
	}
	if len(msg.Bids) != 3 {
		t.Fatalf("Expected 3 bid levels (MaxDepthLevel=3), got %d", len(msg.Bids))
	}
}
