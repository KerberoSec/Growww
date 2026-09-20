package src

import (
	"testing"
	"time"
)

func TestPITRPipeline_IngestAndReplay(t *testing.T) {
	p := NewPITRPipeline()

	t0 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	// 1. Insert order 101 at T+0s
	_, err := p.IngestMutation(
		"orders",
		OpInsert,
		"ord_101",
		nil,
		map[string]interface{}{"symbol": "BTC/USDT", "side": "BUY", "qty": 1.5, "status": "PENDING"},
		t0.UnixNano(),
	)
	if err != nil {
		t.Fatalf("ingest error: %v", err)
	}

	// 2. Update order 101 to FILLED at T+5s
	t1 := t0.Add(5 * time.Second)
	_, err = p.IngestMutation(
		"orders",
		OpUpdate,
		"ord_101",
		map[string]interface{}{"status": "PENDING"},
		map[string]interface{}{"symbol": "BTC/USDT", "side": "BUY", "qty": 1.5, "status": "FILLED"},
		t1.UnixNano(),
	)
	if err != nil {
		t.Fatalf("ingest error: %v", err)
	}

	// 3. Insert order 102 at T+10s
	t2 := t0.Add(10 * time.Second)
	_, err = p.IngestMutation(
		"orders",
		OpInsert,
		"ord_102",
		nil,
		map[string]interface{}{"symbol": "ETH/USDT", "side": "SELL", "qty": 10.0, "status": "PENDING"},
		t2.UnixNano(),
	)
	if err != nil {
		t.Fatalf("ingest error: %v", err)
	}

	// Verify current state has both orders, ord_101 is FILLED
	currOrd101, exists := p.GetCurrentRecord("orders", "ord_101")
	if !exists || currOrd101["status"] != "FILLED" {
		t.Errorf("expected ord_101 FILLED in current state, got %+v", currOrd101)
	}

	// 4. Point-In-Time Recovery to T+2s (ord_101 should be PENDING, ord_102 should NOT exist)
	targetPIT := t0.Add(2 * time.Second)
	restored, count, err := p.RestoreToPointInTime(targetPIT)
	if err != nil {
		t.Fatalf("PITR restore failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 mutation replayed up to T+2s, got %d", count)
	}

	ordersAtPIT := restored["orders"]
	rec101, ok101 := ordersAtPIT["ord_101"]
	if !ok101 || rec101["status"] != "PENDING" {
		t.Errorf("expected ord_101 PENDING at T+2s, got %+v", rec101)
	}
	if _, ok102 := ordersAtPIT["ord_102"]; ok102 {
		t.Errorf("ord_102 should not exist at T+2s")
	}
}

func TestPITRPipeline_CryptographicTamperDetection(t *testing.T) {
	p := NewPITRPipeline()

	now := time.Now()
	mut, _ := p.IngestMutation("wallets", OpInsert, "w_1", nil, map[string]interface{}{"balance": 1000}, now.UnixNano())

	// Tamper with mutation data
	mut.AfterData["balance"] = 9999999

	// Attempt PITR restore -> should fail tamper detection
	_, _, err := p.RestoreToPointInTime(now.Add(1 * time.Hour))
	if err == nil {
		t.Fatalf("expected cryptographic tamper detection error, but passed")
	}
}
