package main

import (
	"testing"
)

func TestTriggerEngine_StopLossAndTakeProfit(t *testing.T) {
	engine := NewTriggerEngine()

	// 1. Sell Stop Loss: triggers when price <= 58,000 USDT
	_ = engine.RegisterConditionalOrder(&ConditionalOrder{
		OrderID:     "sl-sell-1",
		UserID:      "u1",
		Symbol:      "BTC-USDT",
		Side:        "SELL",
		TriggerType: TriggerStopLossMarket,
		StopPriceE8: 5800000000000,
		QuantityE8:  100000000,
	})

	// 2. Sell Take Profit: triggers when price >= 65,000 USDT
	_ = engine.RegisterConditionalOrder(&ConditionalOrder{
		OrderID:     "tp-sell-1",
		UserID:      "u1",
		Symbol:      "BTC-USDT",
		Side:        "SELL",
		TriggerType: TriggerTakeProfitMarket,
		StopPriceE8: 6500000000000,
		QuantityE8:  100000000,
	})

	// 3. Buy Stop Loss (short stop): triggers when price >= 64,000 USDT
	_ = engine.RegisterConditionalOrder(&ConditionalOrder{
		OrderID:     "sl-buy-1",
		UserID:      "u2",
		Symbol:      "BTC-USDT",
		Side:        "BUY",
		TriggerType: TriggerStopLossLimit,
		StopPriceE8: 6400000000000,
		LimitPriceE8: 6420000000000,
		QuantityE8:  100000000,
	})

	// Price tick: 60,000 USDT (no triggers)
	triggered := engine.EvaluateMarketPrice("BTC-USDT", 6000000000000)
	if len(triggered) != 0 {
		t.Fatalf("Expected 0 triggers at 60k, got %d", len(triggered))
	}

	// Price tick: 65,500 USDT -> tp-sell-1 and sl-buy-1 should trigger!
	triggered = engine.EvaluateMarketPrice("BTC-USDT", 6550000000000)
	if len(triggered) != 2 {
		t.Fatalf("Expected 2 triggers at 65.5k, got %d", len(triggered))
	}

	// Price tick: 57,500 USDT -> sl-sell-1 should trigger!
	triggered = engine.EvaluateMarketPrice("BTC-USDT", 5750000000000)
	if len(triggered) != 1 {
		t.Fatalf("Expected 1 trigger at 57.5k, got %d", len(triggered))
	}
	if triggered[0].OrderID != "sl-sell-1" {
		t.Errorf("Expected sl-sell-1, got %s", triggered[0].OrderID)
	}

	// All triggered, active list should be empty
	active := engine.ListActiveOrders()
	if len(active) != 0 {
		t.Errorf("Expected 0 active orders left, got %d", len(active))
	}
}

func TestTriggerEngine_CancelConditionalOrder(t *testing.T) {
	engine := NewTriggerEngine()

	_ = engine.RegisterConditionalOrder(&ConditionalOrder{
		OrderID:     "cond-can-1",
		UserID:      "u3",
		Symbol:      "BTC-USDT",
		Side:        "SELL",
		TriggerType: TriggerStopLossMarket,
		StopPriceE8: 5000000000000,
		QuantityE8:  100,
	})

	if err := engine.CancelOrder("cond-can-1"); err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}

	// Price drops below stop, but order is canceled so must not trigger
	triggered := engine.EvaluateMarketPrice("BTC-USDT", 4900000000000)
	if len(triggered) != 0 {
		t.Fatalf("Canceled order triggered: %v", triggered)
	}
}
