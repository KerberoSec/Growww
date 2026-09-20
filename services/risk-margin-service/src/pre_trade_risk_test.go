package main

import (
	"testing"
)

func TestPreTradeRisk_ValidateOrder(t *testing.T) {
	cfg := PreTradeRiskConfig{
		MaxOrderNotionalUSD:  100000_00000000, // $100,000 max single order
		MaxPriceDeviationBps: 500,              // 5% fat-finger deviation limit
		MaxDailyNotionalUSD:  500000_00000000, // $500,000 daily
	}
	engine := NewPreTradeRiskEngine(cfg)

	user := &UserAccountRiskProfile{
		UserID:             "USER_TEST_001",
		AvailableBalanceE8: 50000_00000000, // $50,000 balance
		DailyTurnoverE8:    0,
		IsFrozen:           false,
		IsCompliant:        true,
	}
	engine.RegisterUserProfile(user)

	markPrice := uint64(64000_00000000) // $64,000

	// 1. Happy path: Buy 0.5 BTC at $64,100 ($32,050 notional < $50,000 balance)
	err := engine.ValidateOrder("USER_TEST_001", "BUY", 64100_00000000, 50000000, markPrice)
	if err != nil {
		t.Fatalf("expected valid order to pass, got: %v", err)
	}

	// 2. Fat-finger rejection: Price $70,000 vs $64,000 (> 5% deviation)
	err = engine.ValidateOrder("USER_TEST_001", "BUY", 70000_00000000, 10000000, markPrice)
	if err == nil {
		t.Fatalf("expected fat-finger deviation error, got nil")
	}

	// 3. Insufficient balance rejection: Buy 1.0 BTC at $64,000 ($64,000 > $50,000 balance)
	err = engine.ValidateOrder("USER_TEST_001", "BUY", 64000_00000000, 100000000, markPrice)
	if err == nil {
		t.Fatalf("expected insufficient balance error, got nil")
	}

	// 4. Max order notional rejection: Buy 2.0 BTC at $64,000 ($128,000 > $100,000 max limit)
	user.AvailableBalanceE8 = 200000_00000000
	err = engine.ValidateOrder("USER_TEST_001", "BUY", 64000_00000000, 200000000, markPrice)
	if err == nil {
		t.Fatalf("expected order notional exceeded error, got nil")
	}

	// 5. Frozen account rejection
	user.IsFrozen = true
	err = engine.ValidateOrder("USER_TEST_001", "BUY", 64000_00000000, 10000000, markPrice)
	if err == nil {
		t.Fatalf("expected error for frozen account, got nil")
	}
}
