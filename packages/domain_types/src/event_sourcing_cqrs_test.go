package src

import (
	"errors"
	"testing"
	"time"
)

func TestEventSourcingOCCAndSnapshots(t *testing.T) {
	store := NewInMemEventStore()
	aggID := "order-agg-8888"

	evt1 := DomainEvent{
		EventID:   "evt-1",
		EventType: "ORDER_CREATED",
		Payload:   []byte(`{"price":25000}`),
	}
	evt2 := DomainEvent{
		EventID:   "evt-2",
		EventType: "MARGIN_LOCKED",
		Payload:   []byte(`{"amount":25000}`),
	}

	// 1. Initial append with expectedVersion = 0
	if err := store.Append(aggID, 0, []DomainEvent{evt1, evt2}); err != nil {
		t.Fatalf("failed initial append: %v", err)
	}

	// Verify events
	evts, err := store.GetEvents(aggID, 0)
	if err != nil || len(evts) != 2 {
		t.Fatalf("expected 2 events, got %d, err=%v", len(evts), err)
	}
	if evts[0].SequenceNumber != 1 || evts[1].SequenceNumber != 2 {
		t.Fatalf("unexpected sequence numbers: %d, %d", evts[0].SequenceNumber, evts[1].SequenceNumber)
	}

	// 2. Concurrency conflict test: attempt append with stale expectedVersion = 0
	evtConflict := DomainEvent{
		EventID:   "evt-conflict",
		EventType: "ORDER_CANCELLED",
	}
	err = store.Append(aggID, 0, []DomainEvent{evtConflict})
	if err == nil || !errors.Is(err, ErrConcurrencyConflict) {
		t.Fatalf("expected ErrConcurrencyConflict, got %v", err)
	}

	// 3. Save snapshot at version 2
	snapshot := AggregateSnapshot{
		AggregateID: aggID,
		Version:     2,
		State:       []byte(`{"status":"PENDING_MATCH"}`),
		Timestamp:   time.Now().UTC(),
	}
	if err := store.SaveSnapshot(snapshot); err != nil {
		t.Fatalf("failed saving snapshot: %v", err)
	}

	retrievedSnapshot, err := store.GetSnapshot(aggID)
	if err != nil || retrievedSnapshot == nil || retrievedSnapshot.Version != 2 {
		t.Fatalf("failed retrieving snapshot")
	}

	// 4. Append subsequent event at version 2
	evt3 := DomainEvent{
		EventID:   "evt-3",
		EventType: "ORDER_MATCHED",
		Payload:   []byte(`{"trade_id":"TRD-100"}`),
	}
	if err := store.Append(aggID, 2, []DomainEvent{evt3}); err != nil {
		t.Fatalf("failed appending event 3: %v", err)
	}

	// Get events since snapshot (fromVersion = 2)
	deltaEvts, err := store.GetEvents(aggID, 2)
	if err != nil || len(deltaEvts) != 1 || deltaEvts[0].SequenceNumber != 3 {
		t.Fatalf("expected 1 delta event with seq 3, got %+v", deltaEvts)
	}
}
