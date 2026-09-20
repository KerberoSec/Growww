package keyspace_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"growww/database/redis/keyspace"
)

func TestKeyspace_OrderAndMarginExpiryDispatch(t *testing.T) {
	dispatcher := keyspace.NewExpiryDispatcher(3, 2*time.Millisecond)
	poller := keyspace.NewActiveReconciliationPoller(dispatcher)
	listener := keyspace.NewRedisKeyspaceListener(dispatcher, poller)
	ctx := context.Background()

	var mu sync.Mutex
	var cancelledOrders []string
	var liquidatedMargins []string

	dispatcher.RegisterHandler(keyspace.EventTypeOrderExpiry, keyspace.ExpiryHandlerFunc(func(ctx context.Context, event keyspace.ExpiryEvent) error {
		mu.Lock()
		defer mu.Unlock()
		cancelledOrders = append(cancelledOrders, event.EntityID)
		return nil
	}))

	dispatcher.RegisterHandler(keyspace.EventTypeMarginCall, keyspace.ExpiryHandlerFunc(func(ctx context.Context, event keyspace.ExpiryEvent) error {
		mu.Lock()
		defer mu.Unlock()
		liquidatedMargins = append(liquidatedMargins, event.EntityID)
		return nil
	}))

	// Simulate Redis __keyevent@0__:expired messages
	err := listener.HandleRawPubSubMessage(ctx, "__keyevent@0__:expired", "order:exp:ORD-9981:BTC-USDT")
	if err != nil {
		t.Fatalf("HandleRawPubSubMessage failed: %v", err)
	}

	err = listener.HandleRawPubSubMessage(ctx, "__keyevent@0__:expired", "margin:call:USR-4421:MARG-001")
	if err != nil {
		t.Fatalf("HandleRawPubSubMessage failed: %v", err)
	}

	mu.Lock()
	if len(cancelledOrders) != 1 || cancelledOrders[0] != "ORD-9981" {
		t.Errorf("expected ORD-9981 cancelled, got %v", cancelledOrders)
	}
	if len(liquidatedMargins) != 1 || liquidatedMargins[0] != "USR-4421" {
		t.Errorf("expected USR-4421 margin call liquidated, got %v", liquidatedMargins)
	}
	mu.Unlock()

	metrics := dispatcher.GetMetrics()
	if metrics.TotalEventsReceived != 2 || metrics.TotalEventsDispatched != 2 {
		t.Errorf("metrics mismatch: %+v", metrics)
	}
}

func TestKeyspace_ActiveReconciliationFallback(t *testing.T) {
	dispatcher := keyspace.NewExpiryDispatcher(3, 2*time.Millisecond)
	poller := keyspace.NewActiveReconciliationPoller(dispatcher)
	ctx := context.Background()

	var sessionCleanups []string
	dispatcher.RegisterHandler(keyspace.EventTypeSessionExpiry, keyspace.ExpiryHandlerFunc(func(ctx context.Context, event keyspace.ExpiryEvent) error {
		sessionCleanups = append(sessionCleanups, event.EntityID)
		return nil
	}))

	pastTime := time.Now().Add(-10 * time.Second)
	poller.AddTimer("session:auth:USR-9000:SESS-abc", pastTime)
	poller.AddTimer("session:auth:USR-9001:SESS-def", time.Now().Add(10*time.Minute)) // future timer

	// Scan overdue timers
	dispatched := poller.ScanAndDispatchExpired(ctx, time.Now())
	if dispatched != 1 {
		t.Fatalf("expected 1 overdue timer dispatched, got %d", dispatched)
	}

	if len(sessionCleanups) != 1 || sessionCleanups[0] != "USR-9000" {
		t.Errorf("expected session USR-9000 cleaned up, got %v", sessionCleanups)
	}
}

func TestKeyspace_DLQAndRetryHandling(t *testing.T) {
	dispatcher := keyspace.NewExpiryDispatcher(2, 2*time.Millisecond)
	ctx := context.Background()

	dispatcher.RegisterHandler(keyspace.EventTypeSLATimer, keyspace.ExpiryHandlerFunc(func(ctx context.Context, event keyspace.ExpiryEvent) error {
		return errors.New("upstream settlement service unavailable")
	}))

	event := keyspace.ParseKey("sla:timer:TASK-881:SETTLEMENT")
	err := dispatcher.Dispatch(ctx, event)
	if err == nil {
		t.Fatalf("expected dispatch error, got nil")
	}

	metrics := dispatcher.GetMetrics()
	if metrics.TotalEventsFailed != 1 {
		t.Errorf("expected 1 failed event, got %d", metrics.TotalEventsFailed)
	}
	if metrics.DLQCount != 1 {
		t.Errorf("expected 1 DLQ entry, got %d", metrics.DLQCount)
	}

	dlqRecords := dispatcher.GetDLQ().GetRecords()
	if len(dlqRecords) != 1 {
		t.Fatalf("expected 1 DLQ record, got %d", len(dlqRecords))
	}
	if dlqRecords[0].Event.EntityID != "TASK-881" {
		t.Errorf("expected DLQ entity TASK-881, got %s", dlqRecords[0].Event.EntityID)
	}
}
