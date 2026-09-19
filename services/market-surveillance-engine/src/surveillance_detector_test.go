package main

import (
	"testing"
	"time"
)

// ──────────────────────────────────────────────────
// Wash Trading Detection Tests
// ──────────────────────────────────────────────────

func TestDetectWashTrading_SelfCross(t *testing.T) {
	engine := NewSurveillanceEngine()

	alert := engine.DetectWashTrading("USER-001", "USER-001", "BTCUSDT", 6500000000000, 100000000)
	if alert == nil {
		t.Fatal("expected wash trading alert for self-cross")
	}
	if alert.Pattern != "WASH_TRADING_SELF_CROSS" {
		t.Errorf("expected pattern WASH_TRADING_SELF_CROSS, got %s", alert.Pattern)
	}
	if alert.ConfidenceRate != 1.00 {
		t.Errorf("expected confidence 1.00, got %.2f", alert.ConfidenceRate)
	}
	if alert.UserID != "USER-001" {
		t.Errorf("expected UserID USER-001, got %s", alert.UserID)
	}
	if alert.Symbol != "BTCUSDT" {
		t.Errorf("expected symbol BTCUSDT, got %s", alert.Symbol)
	}
}

func TestDetectWashTrading_DifferentUsers_NoAlert(t *testing.T) {
	engine := NewSurveillanceEngine()

	alert := engine.DetectWashTrading("USER-001", "USER-002", "ETHUSDT", 3000000000000, 500000000)
	if alert != nil {
		t.Error("expected no alert for different buyer/seller")
	}
}

func TestDetectWashTrading_MultipleAlerts(t *testing.T) {
	engine := NewSurveillanceEngine()

	a1 := engine.DetectWashTrading("USER-X", "USER-X", "BTCUSDT", 100, 100)
	a2 := engine.DetectWashTrading("USER-Y", "USER-Y", "ETHUSDT", 200, 200)

	if a1 == nil || a2 == nil {
		t.Fatal("expected alerts for both self-crosses")
	}
	if a1.AlertID == a2.AlertID {
		t.Error("alert IDs should be unique")
	}
}

// ──────────────────────────────────────────────────
// Spoofing Detection Tests
// ──────────────────────────────────────────────────

func TestDetectSpoofing_FastCancelLargeOrder(t *testing.T) {
	engine := NewSurveillanceEngine()

	// Cancelled in 50ms, notional $500,000
	alert := engine.DetectSpoofing("USER-SPOOF-1", "BTCUSDT", 50, 500000.0)
	if alert == nil {
		t.Fatal("expected spoofing alert for fast cancel of large order")
	}
	if alert.Pattern != "HIGH_FREQUENCY_SPOOFING" {
		t.Errorf("expected pattern HIGH_FREQUENCY_SPOOFING, got %s", alert.Pattern)
	}
	if alert.ConfidenceRate != 0.92 {
		t.Errorf("expected confidence 0.92, got %.2f", alert.ConfidenceRate)
	}
}

func TestDetectSpoofing_SlowCancel_NoAlert(t *testing.T) {
	engine := NewSurveillanceEngine()

	// Cancelled after 5000ms - not spoofing
	alert := engine.DetectSpoofing("USER-LEGIT", "ETHUSDT", 5000, 200000.0)
	if alert != nil {
		t.Error("expected no alert for slow cancel")
	}
}

func TestDetectSpoofing_SmallOrder_NoAlert(t *testing.T) {
	engine := NewSurveillanceEngine()

	// Fast cancel but small notional ($5,000)
	alert := engine.DetectSpoofing("USER-SMALL", "BTCUSDT", 10, 5000.0)
	if alert != nil {
		t.Error("expected no alert for small order even with fast cancel")
	}
}

func TestDetectSpoofing_Boundary_200ms(t *testing.T) {
	engine := NewSurveillanceEngine()

	// Exactly 200ms - should NOT trigger (condition is < 200)
	alert := engine.DetectSpoofing("USER-EDGE", "BTCUSDT", 200, 150000.0)
	if alert != nil {
		t.Error("expected no alert at exactly 200ms boundary")
	}
}

func TestDetectSpoofing_Boundary_100kNotional(t *testing.T) {
	engine := NewSurveillanceEngine()

	// Exactly $100,000 - should NOT trigger (condition is > 100000)
	alert := engine.DetectSpoofing("USER-EDGE2", "BTCUSDT", 50, 100000.0)
	if alert != nil {
		t.Error("expected no alert at exactly $100k notional boundary")
	}
}

// ──────────────────────────────────────────────────
// Circular Trading Detection Tests
// ──────────────────────────────────────────────────

func TestDetectCircularTrading_CyclicRing(t *testing.T) {
	engine := NewSurveillanceEngine()

	now := time.Now().UTC()
	hops := []TradeHop{
		{BuyerID: "A", SellerID: "B", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now},
		{BuyerID: "B", SellerID: "C", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now.Add(1 * time.Second)},
		{BuyerID: "C", SellerID: "A", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now.Add(2 * time.Second)},
	}

	alert := engine.DetectCircularTrading(hops, 10.0) // 10-second window
	if alert == nil {
		t.Fatal("expected circular trading alert for A -> B -> C -> A ring")
	}
	if alert.Pattern != "CIRCULAR_TRADING_RING" {
		t.Errorf("expected pattern CIRCULAR_TRADING_RING, got %s", alert.Pattern)
	}
	if alert.ConfidenceRate != 0.98 {
		t.Errorf("expected confidence 0.98, got %.2f", alert.ConfidenceRate)
	}
}

func TestDetectCircularTrading_TooFewHops(t *testing.T) {
	engine := NewSurveillanceEngine()

	now := time.Now().UTC()
	hops := []TradeHop{
		{BuyerID: "A", SellerID: "B", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now},
		{BuyerID: "B", SellerID: "A", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now.Add(1 * time.Second)},
	}

	alert := engine.DetectCircularTrading(hops, 10.0)
	if alert != nil {
		t.Error("expected no alert for fewer than 3 hops")
	}
}

func TestDetectCircularTrading_WindowExceeded(t *testing.T) {
	engine := NewSurveillanceEngine()

	now := time.Now().UTC()
	hops := []TradeHop{
		{BuyerID: "A", SellerID: "B", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now},
		{BuyerID: "B", SellerID: "C", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now.Add(30 * time.Second)},
		{BuyerID: "C", SellerID: "A", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now.Add(60 * time.Second)},
	}

	alert := engine.DetectCircularTrading(hops, 5.0) // 5-second window
	if alert != nil {
		t.Error("expected no alert when window is exceeded")
	}
}

func TestDetectCircularTrading_NoCycle(t *testing.T) {
	engine := NewSurveillanceEngine()

	now := time.Now().UTC()
	hops := []TradeHop{
		{BuyerID: "A", SellerID: "B", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now},
		{BuyerID: "B", SellerID: "C", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now.Add(1 * time.Second)},
		{BuyerID: "C", SellerID: "D", Symbol: "BTCUSDT", QtyE8: 1e8, PriceE8: 6500e8, Timestamp: now.Add(2 * time.Second)},
	}

	alert := engine.DetectCircularTrading(hops, 10.0)
	if alert != nil {
		t.Error("expected no alert when there is no cycle back to start")
	}
}

func TestDetectCircularTrading_FourNodeCycle(t *testing.T) {
	engine := NewSurveillanceEngine()

	now := time.Now().UTC()
	hops := []TradeHop{
		{BuyerID: "A", SellerID: "B", Symbol: "ETHUSDT", QtyE8: 5e8, PriceE8: 3000e8, Timestamp: now},
		{BuyerID: "B", SellerID: "C", Symbol: "ETHUSDT", QtyE8: 5e8, PriceE8: 3000e8, Timestamp: now.Add(1 * time.Second)},
		{BuyerID: "C", SellerID: "D", Symbol: "ETHUSDT", QtyE8: 5e8, PriceE8: 3000e8, Timestamp: now.Add(2 * time.Second)},
		{BuyerID: "D", SellerID: "A", Symbol: "ETHUSDT", QtyE8: 5e8, PriceE8: 3000e8, Timestamp: now.Add(3 * time.Second)},
	}

	alert := engine.DetectCircularTrading(hops, 10.0)
	if alert == nil {
		t.Fatal("expected circular trading alert for 4-node cycle")
	}
}

// ──────────────────────────────────────────────────
// Layering Detection Tests
// ──────────────────────────────────────────────────

func TestDetectLayering_HighCancelRatio(t *testing.T) {
	engine := NewSurveillanceEngine()

	// 10 orders placed, 9 cancelled (90%) within 500ms
	alert := engine.DetectLayering("USER-LAYER-1", "BTCUSDT", 10, 9, 500)
	if alert == nil {
		t.Fatal("expected layering alert for high cancel ratio within window")
	}
	if alert.Pattern != "ORDERBOOK_LAYERING_MANIPULATION" {
		t.Errorf("expected pattern ORDERBOOK_LAYERING_MANIPULATION, got %s", alert.Pattern)
	}
	if alert.ConfidenceRate != 0.95 {
		t.Errorf("expected confidence 0.95, got %.2f", alert.ConfidenceRate)
	}
}

func TestDetectLayering_LowCancelRatio_NoAlert(t *testing.T) {
	engine := NewSurveillanceEngine()

	// Only 3 of 10 cancelled (30%) - below threshold
	alert := engine.DetectLayering("USER-NORMAL", "ETHUSDT", 10, 3, 500)
	if alert != nil {
		t.Error("expected no alert for low cancel ratio")
	}
}

func TestDetectLayering_TooFewOrders_NoAlert(t *testing.T) {
	engine := NewSurveillanceEngine()

	// Only 3 orders placed (need >= 5)
	alert := engine.DetectLayering("USER-FEW", "BTCUSDT", 3, 3, 100)
	if alert != nil {
		t.Error("expected no alert for fewer than 5 orders")
	}
}

func TestDetectLayering_SlowCancelWindow_NoAlert(t *testing.T) {
	engine := NewSurveillanceEngine()

	// High cancel ratio but over 2000ms (need < 1000ms)
	alert := engine.DetectLayering("USER-SLOW", "BTCUSDT", 10, 9, 2000)
	if alert != nil {
		t.Error("expected no alert for slow cancel window")
	}
}

func TestDetectLayering_ExactBoundary(t *testing.T) {
	engine := NewSurveillanceEngine()

	// Exactly 5 orders, 4 cancelled (80%), 999ms
	alert := engine.DetectLayering("USER-BOUNDARY", "BTCUSDT", 5, 4, 999)
	if alert == nil {
		t.Fatal("expected layering alert at exact boundary conditions")
	}
}
