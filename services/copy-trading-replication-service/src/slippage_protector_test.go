package main

import "testing"

func TestValidateFollowerExecution(t *testing.T) {
	// Buy order: Leader filled at 100, follower filled at 100.4 (40 bps -> ok)
	e1 := CopyTradeExecution{LeaderPrice: 100, FollowerPrice: 100.4, MaxSlippageBps: 50}
	if err := ValidateFollowerExecution(e1, true); err != nil {
		t.Errorf("expected pass, got %v", err)
	}

	// Buy order: Leader filled at 100, follower filled at 101 (100 bps -> fail)
	e2 := CopyTradeExecution{LeaderPrice: 100, FollowerPrice: 101.0, MaxSlippageBps: 50}
	if err := ValidateFollowerExecution(e2, true); err == nil {
		t.Error("expected slippage rejection")
	}
}

func TestScaleQuantity(t *testing.T) {
	qty := ScaleQuantity(1.0, 1000000, 100000) // Leader has 10L, follower has 1L -> 0.1x
	if qty != 0.1 {
		t.Errorf("expected 0.1, got %f", qty)
	}
}
