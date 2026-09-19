package main

import (
	"testing"
)

func TestMarketOrder_ExecutionAndNoOverflow(t *testing.T) {
	// Depth:
	// Level 1: 2 BTC at 65,000 USDT (65000 * 1e8)
	// Level 2: 3 BTC at 65,100 USDT (65100 * 1e8)
	depth := []BookLevel{
		{PriceE8: 6500000000000, QuantityE8: 200000000},
		{PriceE8: 6510000000000, QuantityE8: 300000000},
	}

	req := MarketOrderRequest{
		OrderID:          "mo-big-1",
		UserID:           "inst-1",
		Symbol:           "BTC-USDT",
		Side:             SideBuy,
		QuantityE8:       500000000, // 5 BTC
		MaxSlippageBps:   50,        // 0.50%
		EstimatedPriceE8: 6500000000000,
	}

	res, err := ExecuteMarketOrder(req, depth)
	if err != nil {
		t.Fatalf("ExecuteMarketOrder failed: %v", err)
	}

	if res.FilledQtyE8 != 500000000 {
		t.Errorf("Expected filled 500000000, got %d", res.FilledQtyE8)
	}

	// Expected average price: (2*65000 + 3*65100) / 5 = 65,060 USDT
	expectedAvgPrice := uint64(6506000000000)
	if res.AveragePriceE8 != expectedAvgPrice {
		t.Errorf("Expected average price %d, got %d", expectedAvgPrice, res.AveragePriceE8)
	}

	// Fee invariant: strictly 0.00% universal fee
	if res.PlatformFeeE8 != 0 {
		t.Errorf("Platform fee must be strictly 0, got %d", res.PlatformFeeE8)
	}
}

func TestMarketOrder_SlippageLimitBreached(t *testing.T) {
	depth := []BookLevel{
		{PriceE8: 6600000000000, QuantityE8: 100000000}, // 66,000 USDT
	}

	req := MarketOrderRequest{
		OrderID:          "mo-slip-1",
		UserID:           "inst-2",
		Symbol:           "BTC-USDT",
		Side:             SideBuy,
		QuantityE8:       100000000,
		MaxSlippageBps:   50,            // 0.50% max allowed slippage
		EstimatedPriceE8: 6500000000000, // 65,000 USDT estimated price (+1.53% slippage)
	}

	_, err := ExecuteMarketOrder(req, depth)
	if err == nil {
		t.Fatal("Expected error due to slippage breach, got nil")
	}
}

func TestPostOnly_Validation(t *testing.T) {
	bestBid := uint64(6490000000000) // 64,900 USDT
	bestAsk := uint64(6500000000000) // 65,000 USDT

	// Buy order crossing ask -> should fail post-only
	buyCrossing := LimitOrder{
		PriceE8:  6505000000000,
		Side:     SideBuy,
		PostOnly: true,
	}
	if err := ValidatePostOnly(buyCrossing, bestBid, bestAsk); err == nil {
		t.Error("Expected error for crossing buy post-only order")
	}

	// Sell order crossing bid -> should fail post-only
	sellCrossing := LimitOrder{
		PriceE8:  6485000000000,
		Side:     SideSell,
		PostOnly: true,
	}
	if err := ValidatePostOnly(sellCrossing, bestBid, bestAsk); err == nil {
		t.Error("Expected error for crossing sell post-only order")
	}

	// Non-crossing passive buy order -> should pass
	passiveBuy := LimitOrder{
		PriceE8:  6495000000000, // 64,950 USDT (between bid and ask)
		Side:     SideBuy,
		PostOnly: true,
	}
	if err := ValidatePostOnly(passiveBuy, bestBid, bestAsk); err != nil {
		t.Errorf("Unexpected error for passive buy: %v", err)
	}
}

func TestEvaluateTIF(t *testing.T) {
	// FOK incomplete fill -> rejected
	ok, status := EvaluateTIF(TIF_FOK, 50, 100)
	if ok || status != "FOK_REJECTED_INCOMPLETE_FILL" {
		t.Errorf("FOK incomplete fill test failed: ok=%v, status=%s", ok, status)
	}

	// FOK complete fill -> filled
	ok, status = EvaluateTIF(TIF_FOK, 100, 100)
	if !ok || status != "FILLED" {
		t.Errorf("FOK complete fill test failed: ok=%v, status=%s", ok, status)
	}

	// IOC zero fill -> cancelled
	ok, status = EvaluateTIF(TIF_IOC, 0, 100)
	if ok || status != "IOC_CANCELLED_ZERO_FILL" {
		t.Errorf("IOC zero fill test failed: ok=%v, status=%s", ok, status)
	}
}

func TestAdvancedOrderRouter_NBBOCollarAndExecution(t *testing.T) {
	bestBid := uint64(6000000000000) // 60,000 USDT
	bestAsk := uint64(6000000000000) // 60,000 USDT (mid = 60,000, +/-5% = 57,000 to 63,000)

	// 1. Child order breaching upper NBBO collar (+5% is 63,000 USDT)
	breachReq := ChildOrderRequest{
		ChildOrderID:  "twap-slice-1",
		ParentOrderID: "twap-parent-1",
		AlgoType:      AlgoTypeTWAP,
		UserID:        "user-1",
		Symbol:        "BTC-USDT",
		Side:          SideBuy,
		OrderType:     "LIMIT",
		PriceE8:       6400000000000, // 64,000 USDT > 63,000 upper collar
		QuantityE8:    10000000,
	}

	res, err := RouteChildOrder(breachReq, bestBid, bestAsk, nil)
	if err != nil {
		t.Fatalf("RouteChildOrder failed: %v", err)
	}
	if res.Status != "REJECTED" {
		t.Errorf("Expected REJECTED status, got %s", res.Status)
	}

	// 2. Child order within collar -> Successfully rests on book
	validReq := ChildOrderRequest{
		ChildOrderID:  "iceberg-clip-1",
		ParentOrderID: "ice-parent-1",
		AlgoType:      AlgoTypeIceberg,
		UserID:        "user-2",
		Symbol:        "BTC-USDT",
		Side:          SideBuy,
		OrderType:     "LIMIT",
		PriceE8:       5990000000000, // 59,900 USDT (passive bid)
		QuantityE8:    10000000,
		PostOnly:      true,
	}

	res2, err := RouteChildOrder(validReq, bestBid, bestAsk, nil)
	if err != nil {
		t.Fatalf("RouteChildOrder valid failed: %v", err)
	}
	if res2.Status != "RESTING" {
		t.Errorf("Expected RESTING status, got %s", res2.Status)
	}
	if res2.PlatformFeeE8 != 0 {
		t.Errorf("Platform fee must be 0, got %d", res2.PlatformFeeE8)
	}
}
