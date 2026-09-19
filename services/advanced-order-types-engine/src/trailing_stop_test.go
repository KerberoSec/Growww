package main

import (
	"testing"
)

func TestTrailingStop_SellRatchetsUpwardsAndNeverDownwards(t *testing.T) {
	engine := NewTrailingStopEngine()

	// BTC-USDT SELL order: Initial price 60,000 USDT (60000 * 1e8)
	// Trailing Delta: 200 bps (2.00% = 1,200 USDT)
	initialPrice := uint64(6000000000000)
	order := &TrailingStopOrder{
		OrderID:          "ts-sell-1",
		UserID:           "u1",
		Symbol:           "BTC-USDT",
		Side:             "SELL",
		QuantityE8:       100000000,
		TrailingDeltaBps: 200, // 2%
	}

	if err := engine.RegisterOrder(order, initialPrice); err != nil {
		t.Fatalf("RegisterOrder failed: %v", err)
	}

	// Initial Stop should be: 60,000 - 2% (1,200) = 58,800 USDT
	expectedInitialStop := uint64(5880000000000)
	snap, _ := engine.GetOrder("ts-sell-1")
	if snap.CurrentStopE8 != expectedInitialStop {
		t.Fatalf("Expected initial stop %d, got %d", expectedInitialStop, snap.CurrentStopE8)
	}

	// Price moves UP to 65,000 USDT (+5,000)
	priceUp1 := uint64(6500000000000)
	triggered := engine.OnPriceTick("BTC-USDT", priceUp1)
	if len(triggered) != 0 {
		t.Fatalf("Unexpected trigger on upward price movement")
	}

	// Stop should ratchet UP to: 65,000 - 2% (1,300) = 63,700 USDT
	expectedStopUp1 := uint64(6370000000000)
	snap, _ = engine.GetOrder("ts-sell-1")
	if snap.HighWaterMarkE8 != priceUp1 {
		t.Errorf("Expected high water mark %d, got %d", priceUp1, snap.HighWaterMarkE8)
	}
	if snap.CurrentStopE8 != expectedStopUp1 {
		t.Fatalf("Expected ratcheted stop %d, got %d", expectedStopUp1, snap.CurrentStopE8)
	}

	// Price dips slightly to 64,500 USDT (above stop level of 63,700)
	priceDip := uint64(6450000000000)
	triggered = engine.OnPriceTick("BTC-USDT", priceDip)
	if len(triggered) != 0 {
		t.Fatalf("Unexpected trigger on minor dip above stop level")
	}

	// Invariant Check: Stop MUST NOT move down! It must stay at 63,700 USDT
	snap, _ = engine.GetOrder("ts-sell-1")
	if snap.CurrentStopE8 != expectedStopUp1 {
		t.Fatalf("INVARIANT VIOLATED: Sell trailing stop dropped from %d to %d!", expectedStopUp1, snap.CurrentStopE8)
	}

	// Price drops to or below stop price: 63,700 USDT -> MUST TRIGGER!
	priceHit := uint64(6370000000000)
	triggered = engine.OnPriceTick("BTC-USDT", priceHit)
	if len(triggered) != 1 {
		t.Fatalf("Expected order to trigger at stop level, got %d triggered", len(triggered))
	}
	if triggered[0].OrderID != "ts-sell-1" {
		t.Errorf("Expected order ts-sell-1 to trigger, got %s", triggered[0].OrderID)
	}
	if !triggered[0].Triggered {
		t.Errorf("Order flag Triggered should be true")
	}
	if triggered[0].TriggerPriceE8 != priceHit {
		t.Errorf("Expected trigger price %d, got %d", priceHit, triggered[0].TriggerPriceE8)
	}
}

func TestTrailingStop_BuyRatchetsDownwardsAndNeverUpwards(t *testing.T) {
	engine := NewTrailingStopEngine()

	// ETH-USDT BUY order (dip buying): Initial price 3,000 USDT (3000 * 1e8)
	// Trailing Delta: 300 bps (3.00% = 90 USDT)
	initialPrice := uint64(300000000000)
	order := &TrailingStopOrder{
		OrderID:          "ts-buy-1",
		UserID:           "u2",
		Symbol:           "ETH-USDT",
		Side:             "BUY",
		QuantityE8:       500000000,
		TrailingDeltaBps: 300, // 3%
	}

	_ = engine.RegisterOrder(order, initialPrice)

	// Initial Stop should be: 3,000 + 3% (90) = 3,090 USDT
	expectedInitialStop := uint64(309000000000)
	snap, _ := engine.GetOrder("ts-buy-1")
	if snap.CurrentStopE8 != expectedInitialStop {
		t.Fatalf("Expected initial buy stop %d, got %d", expectedInitialStop, snap.CurrentStopE8)
	}

	// Price falls to 2,800 USDT (new low water mark)
	priceDown1 := uint64(280000000000)
	engine.OnPriceTick("ETH-USDT", priceDown1)

	// Stop should ratchet DOWN to: 2,800 + 3% (84) = 2,884 USDT
	expectedStopDown1 := uint64(288400000000)
	snap, _ = engine.GetOrder("ts-buy-1")
	if snap.LowWaterMarkE8 != priceDown1 {
		t.Errorf("Expected low water mark %d, got %d", priceDown1, snap.LowWaterMarkE8)
	}
	if snap.CurrentStopE8 != expectedStopDown1 {
		t.Fatalf("Expected ratcheted buy stop %d, got %d", expectedStopDown1, snap.CurrentStopE8)
	}

	// Price bounces slightly to 2,850 USDT (below stop level of 2,884)
	priceBounce := uint64(285000000000)
	triggered := engine.OnPriceTick("ETH-USDT", priceBounce)
	if len(triggered) != 0 {
		t.Fatalf("Unexpected trigger on minor bounce below stop level")
	}

	// Invariant Check: Buy stop MUST NOT move up! It must stay at 2,884 USDT
	snap, _ = engine.GetOrder("ts-buy-1")
	if snap.CurrentStopE8 != expectedStopDown1 {
		t.Fatalf("INVARIANT VIOLATED: Buy trailing stop increased from %d to %d!", expectedStopDown1, snap.CurrentStopE8)
	}

	// Price bounces up to or above stop level: 2,884 USDT -> MUST TRIGGER!
	triggered = engine.OnPriceTick("ETH-USDT", expectedStopDown1)
	if len(triggered) != 1 {
		t.Fatalf("Expected order to trigger, got %d", len(triggered))
	}
}

func TestTrailingStop_ActivationPriceThreshold(t *testing.T) {
	engine := NewTrailingStopEngine()

	// SELL order at initial price 50,000 USDT.
	// Activation threshold: 55,000 USDT (only activate after reaching 55,000 USDT).
	order := &TrailingStopOrder{
		OrderID:           "ts-act-1",
		UserID:            "u3",
		Symbol:            "BTC-USDT",
		Side:              "SELL",
		QuantityE8:        100000000,
		TrailingDeltaBps:  200, // 2%
		ActivationPriceE8: 5500000000000,
	}

	_ = engine.RegisterOrder(order, 5000000000000)

	snap, _ := engine.GetOrder("ts-act-1")
	if snap.IsActivated {
		t.Fatal("Order should NOT be activated initially at 50,000 USDT")
	}

	// Price rises to 52,000 USDT -> Still not activated
	engine.OnPriceTick("BTC-USDT", 5200000000000)
	snap, _ = engine.GetOrder("ts-act-1")
	if snap.IsActivated {
		t.Fatal("Order should NOT be activated at 52,000 USDT")
	}

	// Price drops to 48,000 USDT -> Must not trigger because it's not active!
	triggered := engine.OnPriceTick("BTC-USDT", 4800000000000)
	if len(triggered) != 0 {
		t.Fatal("Order must not trigger before activation!")
	}

	// Price rallies to 55,000 USDT -> Activates!
	engine.OnPriceTick("BTC-USDT", 5500000000000)
	snap, _ = engine.GetOrder("ts-act-1")
	if !snap.IsActivated {
		t.Fatal("Order should be activated once price reaches 55,000 USDT")
	}
	// Stop should now be set at 55,000 - 2% (1,100) = 53,900 USDT
	expectedStop := uint64(5390000000000)
	if snap.CurrentStopE8 != expectedStop {
		t.Fatalf("Expected stop %d after activation, got %d", expectedStop, snap.CurrentStopE8)
	}

	// Now a retrace to 53,800 USDT triggers the order!
	triggered = engine.OnPriceTick("BTC-USDT", 5380000000000)
	if len(triggered) != 1 {
		t.Fatalf("Expected activated order to trigger on retrace, got %d", len(triggered))
	}
}

func TestTrailingStop_StopLimitOrderGeneration(t *testing.T) {
	engine := NewTrailingStopEngine()

	// SELL STOP_LIMIT order with LimitOffsetBps = 50 (0.50% limit below stop)
	order := &TrailingStopOrder{
		OrderID:          "ts-limit-1",
		UserID:           "u4",
		Symbol:           "BTC-USDT",
		Side:             "SELL",
		OrderType:        TrailingOrderTypeStopLimit,
		QuantityE8:       100000000,
		TrailingDeltaBps: 200, // 2% stop
		LimitOffsetBps:   50,  // 0.5% limit offset below trigger price
	}

	_ = engine.RegisterOrder(order, 6000000000000)

	// Trigger stop price at 58,800 USDT (58800 * 1e8)
	triggerPrice := uint64(5880000000000)
	triggered := engine.OnPriceTick("BTC-USDT", triggerPrice)
	if len(triggered) != 1 {
		t.Fatalf("Expected trigger, got %d", len(triggered))
	}

	trigOrder := triggered[0]
	// Expected generated limit: 58,800 - 0.5% (294) = 58,506 USDT
	expectedLimit := uint64(5850600000000)
	if trigOrder.GeneratedLimitPriceE8 != expectedLimit {
		t.Fatalf("Expected generated limit price %d, got %d", expectedLimit, trigOrder.GeneratedLimitPriceE8)
	}
}

func TestTrailingStop_DynamicVolatilityBufferAdjustment(t *testing.T) {
	engine := NewTrailingStopEngine()

	order := &TrailingStopOrder{
		OrderID:          "ts-vol-1",
		UserID:           "u5",
		Symbol:           "BTC-USDT",
		Side:             "SELL",
		QuantityE8:       100000000,
		TrailingDeltaBps: 200, // 2%
	}
	_ = engine.RegisterOrder(order, 6000000000000)

	snap, _ := engine.GetOrder("ts-vol-1")
	// Initial stop: 58,800 USDT
	if snap.CurrentStopE8 != 5880000000000 {
		t.Fatalf("Expected stop 5880000000000, got %d", snap.CurrentStopE8)
	}

	// Market spikes in volatility, update volatility buffer to 100 bps (+1.00%)
	engine.UpdateVolatilityBuffer("BTC-USDT", 100)

	snap, _ = engine.GetOrder("ts-vol-1")
	if snap.VolatilityBufferBps != 100 {
		t.Errorf("Expected volatility buffer 100 bps, got %d", snap.VolatilityBufferBps)
	}
}

func TestTrailingStop_Cancellation(t *testing.T) {
	engine := NewTrailingStopEngine()

	order := &TrailingStopOrder{
		OrderID:          "ts-cancel-1",
		UserID:           "u6",
		Symbol:           "BTC-USDT",
		Side:             "SELL",
		QuantityE8:       100000000,
		TrailingDeltaBps: 200,
	}
	_ = engine.RegisterOrder(order, 6000000000000)

	if err := engine.CancelOrder("ts-cancel-1"); err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	// Trigger price reached, but order is canceled so it should NOT fire
	triggered := engine.OnPriceTick("BTC-USDT", 5800000000000)
	if len(triggered) != 0 {
		t.Fatalf("Canceled order should not trigger")
	}

	active := engine.ListActiveOrders()
	if len(active) != 0 {
		t.Errorf("Expected 0 active orders, got %d", len(active))
	}
}
