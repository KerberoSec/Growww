package main

import (
	"testing"
	"time"
)

func TestP2PTimeoutWatchdog_AutoCancel(t *testing.T) {
	wd := NewP2PTimeoutWatchdog()
	baseTime := time.Now()

	t1 := &WatchdogEscrowTrade{TradeID: "T1", Status: "CREATED", CreatedAt: baseTime.Add(-20 * time.Minute), PaymentTTL: 15 * time.Minute}
	t2 := &WatchdogEscrowTrade{TradeID: "T2", Status: "CREATED", CreatedAt: baseTime.Add(-5 * time.Minute), PaymentTTL: 15 * time.Minute}
	t3 := &WatchdogEscrowTrade{TradeID: "T3", Status: "FIAT_MARKED_PAID", CreatedAt: baseTime.Add(-30 * time.Minute), PaymentTTL: 15 * time.Minute}

	wd.RegisterTrade(t1)
	wd.RegisterTrade(t2)
	wd.RegisterTrade(t3)

	cancelled := wd.SweepExpiredTrades(baseTime)
	if len(cancelled) != 1 || cancelled[0] != "T1" {
		t.Fatalf("expected only T1 cancelled, got %v", cancelled)
	}
	if t3.Status != "FIAT_MARKED_PAID" {
		t.Error("trade marked paid must not be auto-cancelled")
	}
}
