package src

import (
	"errors"
	"testing"
	"time"
)

func TestIdempotencyLockerWorkflow(t *testing.T) {
	locker := NewMemoryIdempotencyLocker()
	key := "idemp-trade-tx-99991"
	leaseToken := "node-worker-a-tok-01"

	// 1. First acquisition succeeds
	rec, err := locker.AcquireLease(key, leaseToken, 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error acquiring lease: %v", err)
	}
	if rec.Status != IdempotencyInFlight {
		t.Fatalf("expected status IN_FLIGHT, got %s", rec.Status)
	}

	// 2. Concurrent duplicate acquisition while in flight must return ErrLockConflict
	_, err = locker.AcquireLease(key, "node-worker-b-tok-02", 1*time.Minute)
	if err == nil || !errors.Is(err, ErrLockConflict) {
		t.Fatalf("expected ErrLockConflict for concurrent lease, got %v", err)
	}

	// 3. Commit execution with response payload
	respPayload := []byte(`{"order_id":"ORD-99991","status":"MATCHED"}`)
	if err := locker.CommitExecution(key, leaseToken, 200, respPayload, 24*time.Hour); err != nil {
		t.Fatalf("failed to commit execution: %v", err)
	}

	// 4. Replay of same idempotency key returns committed cached payload without error
	cachedRec, err := locker.AcquireLease(key, "node-worker-c-tok-03", 1*time.Minute)
	if err != nil {
		t.Fatalf("expected cached replay to succeed, got %v", err)
	}
	if cachedRec.Status != IdempotencyCommitted {
		t.Fatalf("expected COMMITTED status, got %s", cachedRec.Status)
	}
	if string(cachedRec.ResponseBody) != string(respPayload) {
		t.Fatalf("cached response mismatch: %s vs %s", string(cachedRec.ResponseBody), string(respPayload))
	}

	// 5. Test rejection flow
	rejKey := "idemp-trade-tx-rejected"
	rejToken := "tok-rej-01"
	_, err = locker.AcquireLease(rejKey, rejToken, 1*time.Minute)
	if err != nil {
		t.Fatalf("failed acquiring lease: %v", err)
	}
	if err := locker.RejectExecution(rejKey, rejToken, "Validation failed"); err != nil {
		t.Fatalf("failed to reject execution: %v", err)
	}

	// After rejection, can acquire lease anew
	_, err = locker.AcquireLease(rejKey, "tok-new-02", 1*time.Minute)
	if err != nil {
		t.Fatalf("expected re-acquisition after rejection to succeed, got %v", err)
	}
}
