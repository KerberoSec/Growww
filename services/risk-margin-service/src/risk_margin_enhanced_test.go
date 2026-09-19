package main

import (
	"math"
	"testing"
	"time"
)

func TestCalculateMarginRequired_Default(t *testing.T) {
	engine := NewRealTimeMarginEngine()

	// No custom params - should use defaults (10% initial, 5% maintenance)
	req := engine.CalculateMarginRequired("BTCUSDT", "BUY", 65000_00000000, 1_00000000)

	expectedNotional := 65000.0
	if math.Abs(req.NotionalUSD-expectedNotional) > 0.01 {
		t.Errorf("expected notional %.2f, got %.2f", expectedNotional, req.NotionalUSD)
	}
	if math.Abs(req.InitialMarginUSD-6500.0) > 0.01 {
		t.Errorf("expected initial margin 6500, got %.2f", req.InitialMarginUSD)
	}
	if math.Abs(req.MaintenanceMarginUSD-3250.0) > 0.01 {
		t.Errorf("expected maintenance margin 3250, got %.2f", req.MaintenanceMarginUSD)
	}
}

func TestCalculateMarginRequired_CustomParams(t *testing.T) {
	engine := NewRealTimeMarginEngine()
	engine.SetMarginParams("ETHUSDT", 20.0, 10.0)

	req := engine.CalculateMarginRequired("ETHUSDT", "SELL", 3000_00000000, 10_00000000)

	expectedNotional := 30000.0
	if math.Abs(req.NotionalUSD-expectedNotional) > 0.01 {
		t.Errorf("expected notional %.2f, got %.2f", expectedNotional, req.NotionalUSD)
	}
	if math.Abs(req.InitialMarginUSD-6000.0) > 0.01 {
		t.Errorf("expected initial margin 6000, got %.2f", req.InitialMarginUSD)
	}
}

func TestUnrealizedPnL_LongProfit(t *testing.T) {
	engine := NewRealTimeMarginEngine()

	pos := Position{
		PositionID:   "POS-001",
		UserID:       "USER-1",
		Symbol:       "BTCUSDT",
		Side:         PositionLong,
		EntryPriceE8: 60000_00000000,
		QuantityE8:   1_00000000,
		MarginUsedE8: 6000_00000000,
		OpenedAt:     time.Now(),
	}
	_ = engine.OpenPosition(pos)
	engine.UpdateMarkPrice("BTCUSDT", 65000_00000000)

	pnl, err := engine.CalculateUnrealizedPnL("POS-001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Long: (65000 - 60000) * 1 = 5000
	if math.Abs(pnl-5000.0) > 0.01 {
		t.Errorf("expected PnL 5000, got %.2f", pnl)
	}
}

func TestUnrealizedPnL_ShortProfit(t *testing.T) {
	engine := NewRealTimeMarginEngine()

	pos := Position{
		PositionID:   "POS-002",
		UserID:       "USER-2",
		Symbol:       "ETHUSDT",
		Side:         PositionShort,
		EntryPriceE8: 3200_00000000,
		QuantityE8:   10_00000000,
		MarginUsedE8: 3200_00000000,
		OpenedAt:     time.Now(),
	}
	_ = engine.OpenPosition(pos)
	engine.UpdateMarkPrice("ETHUSDT", 3000_00000000)

	pnl, err := engine.CalculateUnrealizedPnL("POS-002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Short: (3200 - 3000) * 10 = 2000
	if math.Abs(pnl-2000.0) > 0.01 {
		t.Errorf("expected PnL 2000, got %.2f", pnl)
	}
}

func TestUnrealizedPnL_PositionNotFound(t *testing.T) {
	engine := NewRealTimeMarginEngine()

	_, err := engine.CalculateUnrealizedPnL("NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for nonexistent position")
	}
}

func TestAggregatePositionRisk_SafeLevel(t *testing.T) {
	engine := NewRealTimeMarginEngine()

	pos := Position{
		PositionID:   "POS-003",
		UserID:       "USER-3",
		Symbol:       "BTCUSDT",
		Side:         PositionLong,
		EntryPriceE8: 60000_00000000,
		QuantityE8:   1_00000000,
		MarginUsedE8: 6000_00000000,
		OpenedAt:     time.Now(),
	}
	_ = engine.OpenPosition(pos)
	engine.UpdateMarkPrice("BTCUSDT", 65000_00000000)

	summary, err := engine.AggregatePositionRisk("USER-3", 15000.0) // $15k equity
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Equity(15000) + PnL(5000) / Margin(6000) * 100 = 333%
	if summary.LiquidationLevel != LevelSafe {
		t.Errorf("expected SAFE level, got %s", summary.LiquidationLevel)
	}
	if summary.PositionCount != 1 {
		t.Errorf("expected 1 position, got %d", summary.PositionCount)
	}
}

func TestAggregatePositionRisk_LiquidateLevel(t *testing.T) {
	engine := NewRealTimeMarginEngine()

	pos := Position{
		PositionID:   "POS-004",
		UserID:       "USER-4",
		Symbol:       "BTCUSDT",
		Side:         PositionLong,
		EntryPriceE8: 60000_00000000,
		QuantityE8:   1_00000000,
		MarginUsedE8: 6000_00000000,
		OpenedAt:     time.Now(),
	}
	_ = engine.OpenPosition(pos)
	engine.UpdateMarkPrice("BTCUSDT", 55000_00000000) // Loss

	// Equity(1000) + PnL(-5000) / Margin(6000) * 100 = -66% => LIQUIDATE
	summary, err := engine.AggregatePositionRisk("USER-4", 1000.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.LiquidationLevel != LevelLiquidate {
		t.Errorf("expected LIQUIDATE level, got %s (margin level: %.2f%%)", summary.LiquidationLevel, summary.MarginLevel)
	}
}

func TestAggregatePositionRisk_NoPositions(t *testing.T) {
	engine := NewRealTimeMarginEngine()

	summary, err := engine.AggregatePositionRisk("USER-EMPTY", 10000.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.LiquidationLevel != LevelSafe {
		t.Errorf("expected SAFE for no positions, got %s", summary.LiquidationLevel)
	}
}

func TestClassifyLiquidationLevel(t *testing.T) {
	tests := []struct {
		marginPct float64
		expected  LiquidationLevel
	}{
		{200, LevelSafe},
		{150.01, LevelSafe},
		{140, LevelWarning},
		{120.01, LevelWarning},
		{115, LevelDanger},
		{110.01, LevelDanger},
		{109, LevelLiquidate},
		{50, LevelLiquidate},
		{0, LevelLiquidate},
	}

	for _, tc := range tests {
		got := classifyLiquidationLevel(tc.marginPct)
		if got != tc.expected {
			t.Errorf("margin %.2f%%: expected %s, got %s", tc.marginPct, tc.expected, got)
		}
	}
}

func TestOpenPosition_Duplicate(t *testing.T) {
	engine := NewRealTimeMarginEngine()

	pos := Position{PositionID: "POS-DUP", UserID: "USER-DUP", Symbol: "BTCUSDT", Side: PositionLong}
	_ = engine.OpenPosition(pos)
	err := engine.OpenPosition(pos)
	if err == nil {
		t.Fatal("expected error for duplicate position")
	}
}
