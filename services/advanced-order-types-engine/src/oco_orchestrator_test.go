package main

import (
	"testing"
)

func TestOCOOrchestrator_LimitLegFillCancelsStopLeg(t *testing.T) {
	orchestrator := NewOCOOrchestrator()

	pair := &OCOPair{
		PairID:       "oco-1",
		UserID:       "u1",
		Symbol:       "BTC-USDT",
		Side:         "SELL",
		LimitOrderID: "limit-leg-1",
		LimitPriceE8: 6500000000000, // Profit target limit: 65,000 USDT
		StopOrderID:  "stop-leg-1",
		StopPriceE8:  5800000000000, // Stop loss trigger: 58,000 USDT
		QuantityE8:   100000000,
	}

	if err := orchestrator.RegisterOCO(pair); err != nil {
		t.Fatalf("RegisterOCO failed: %v", err)
	}

	// Limit order fills completely
	canceledStopID, err := orchestrator.OnLimitOrderFill("limit-leg-1")
	if err != nil {
		t.Fatalf("OnLimitOrderFill failed: %v", err)
	}
	if canceledStopID != "stop-leg-1" {
		t.Fatalf("Expected canceled stop ID stop-leg-1, got %s", canceledStopID)
	}

	snap, _ := orchestrator.GetOCO("oco-1")
	if snap.Status != OCOLimitWon {
		t.Errorf("Expected status LIMIT_FILLED, got %s", snap.Status)
	}
	if snap.RemainingQtyE8 != 0 {
		t.Errorf("Expected remaining qty 0, got %d", snap.RemainingQtyE8)
	}
}

func TestOCOOrchestrator_StopLegTriggerCancelsLimitLeg(t *testing.T) {
	orchestrator := NewOCOOrchestrator()

	pair := &OCOPair{
		PairID:       "oco-2",
		UserID:       "u2",
		Symbol:       "ETH-USDT",
		Side:         "SELL",
		LimitOrderID: "limit-leg-2",
		LimitPriceE8: 350000000000, // 3,500 USDT
		StopOrderID:  "stop-leg-2",
		StopPriceE8:  290000000000, // 2,900 USDT
		QuantityE8:   200000000,
	}

	_ = orchestrator.RegisterOCO(pair)

	// Stop price triggers
	canceledLimitID, err := orchestrator.OnStopTrigger("stop-leg-2")
	if err != nil {
		t.Fatalf("OnStopTrigger failed: %v", err)
	}
	if canceledLimitID != "limit-leg-2" {
		t.Fatalf("Expected canceled limit ID limit-leg-2, got %s", canceledLimitID)
	}

	snap, _ := orchestrator.GetOCO("oco-2")
	if snap.Status != OCOStopWon {
		t.Errorf("Expected status STOP_TRIGGERED, got %s", snap.Status)
	}
}

func TestOCOOrchestrator_PartialLimitFillAdjustsStopQty(t *testing.T) {
	orchestrator := NewOCOOrchestrator()

	pair := &OCOPair{
		PairID:       "oco-3",
		UserID:       "u3",
		Symbol:       "BTC-USDT",
		Side:         "SELL",
		LimitOrderID: "limit-leg-3",
		LimitPriceE8: 7000000000000,
		StopOrderID:  "stop-leg-3",
		StopPriceE8:  5500000000000,
		QuantityE8:   100000000, // 1 BTC
	}
	_ = orchestrator.RegisterOCO(pair)

	// Partial fill of 40,000,000 (0.4 BTC) on the limit leg
	stopID, remainingStopQty, err := orchestrator.OnLimitOrderPartialFill("limit-leg-3", 40000000)
	if err != nil {
		t.Fatalf("Partial fill failed: %v", err)
	}
	if stopID != "stop-leg-3" {
		t.Errorf("Expected stop ID stop-leg-3, got %s", stopID)
	}
	if remainingStopQty != 60000000 {
		t.Errorf("Expected remaining stop qty 60000000, got %d", remainingStopQty)
	}

	snap, _ := orchestrator.GetOCO("oco-3")
	if snap.Status != OCOActive {
		t.Errorf("Expected status to remain ACTIVE on partial fill, got %s", snap.Status)
	}
	if snap.RemainingQtyE8 != 60000000 {
		t.Errorf("Expected remaining qty 60000000, got %d", snap.RemainingQtyE8)
	}
}

func TestOCOOrchestrator_PriceValidationRules(t *testing.T) {
	orchestrator := NewOCOOrchestrator()

	// SELL OCO: limit price must be > stop price. Here limit <= stop -> Must reject!
	invalidSell := &OCOPair{
		PairID:       "oco-inv-sell",
		UserID:       "u4",
		Symbol:       "BTC-USDT",
		Side:         "SELL",
		LimitOrderID: "l1",
		LimitPriceE8: 5000000000000,
		StopOrderID:  "s1",
		StopPriceE8:  5500000000000, // Stop is higher than limit for sell -> Invalid!
		QuantityE8:   1000,
	}
	if err := orchestrator.RegisterOCO(invalidSell); err == nil {
		t.Fatal("Expected error for invalid sell OCO price relationship")
	}

	// BUY OCO: limit price must be < stop price. Here limit >= stop -> Must reject!
	invalidBuy := &OCOPair{
		PairID:       "oco-inv-buy",
		UserID:       "u5",
		Symbol:       "BTC-USDT",
		Side:         "BUY",
		LimitOrderID: "l2",
		LimitPriceE8: 6000000000000,
		StopOrderID:  "s2",
		StopPriceE8:  5500000000000, // Limit is higher than stop for buy -> Invalid!
		QuantityE8:   1000,
	}
	if err := orchestrator.RegisterOCO(invalidBuy); err == nil {
		t.Fatal("Expected error for invalid buy OCO price relationship")
	}
}

func TestOCOOrchestrator_CancelOCO(t *testing.T) {
	orchestrator := NewOCOOrchestrator()

	pair := &OCOPair{
		PairID:       "oco-cancel",
		UserID:       "u6",
		Symbol:       "BTC-USDT",
		Side:         "SELL",
		LimitOrderID: "l-can",
		LimitPriceE8: 7000000000000,
		StopOrderID:  "s-can",
		StopPriceE8:  5000000000000,
		QuantityE8:   100,
	}
	_ = orchestrator.RegisterOCO(pair)

	lID, sID, err := orchestrator.CancelOCO("oco-cancel")
	if err != nil {
		t.Fatalf("CancelOCO failed: %v", err)
	}
	if lID != "l-can" || sID != "s-can" {
		t.Fatalf("Expected legs l-can and s-can, got %s and %s", lID, sID)
	}

	snap, _ := orchestrator.GetOCO("oco-cancel")
	if snap.Status != OCOCanceled {
		t.Errorf("Expected CANCELED, got %s", snap.Status)
	}
}
