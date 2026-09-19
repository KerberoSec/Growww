package main

import "testing"

func TestHotColdRebalance_Sweep(t *testing.T) {
	reb := NewHotColdRebalancer()
	reb.SetBalances("BTC", 15.0, 100.0) // Hot has 15 BTC, capacity is 10 BTC

	prop := reb.EvaluateSweepToCold("BTC")
	if prop == nil {
		t.Fatal("expected sweep proposal when hot balance > capacity")
	}
	if prop.Amount != 7.0 { // 15 - (10 * 0.8) = 7.0
		t.Errorf("expected 7.0 BTC sweep, got %f", prop.Amount)
	}
	if prop.DestinationType != "COLD_VAULT" {
		t.Errorf("expected destination COLD_VAULT, got %s", prop.DestinationType)
	}
}
