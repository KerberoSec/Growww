package main

import (
	"fmt"
	"testing"
	"time"
)

// ──────────────────────────────────────────────────
// 24h Rolling Ticker Tests
// ──────────────────────────────────────────────────

func TestTickerService_FirstTrade(t *testing.T) {
	ts := NewTickerService()

	ticker := ts.UpdateTicker("BTCUSDT", 6500000000000, 100000000) // 65000 USDT, 1 BTC
	if ticker.OpenE8 != 6500000000000 {
		t.Errorf("expected open 6500000000000, got %d", ticker.OpenE8)
	}
	if ticker.TradeCount != 1 {
		t.Errorf("expected trade count 1, got %d", ticker.TradeCount)
	}
	if ticker.ChangePct != 0.0 {
		t.Errorf("expected 0%% change on first trade, got %.4f", ticker.ChangePct)
	}
}

func TestTickerService_MultipleTradesHighLow(t *testing.T) {
	ts := NewTickerService()

	ts.UpdateTicker("ETHUSDT", 3000_00000000, 10_00000000)
	ts.UpdateTicker("ETHUSDT", 3100_00000000, 5_00000000) // New high
	ts.UpdateTicker("ETHUSDT", 2900_00000000, 8_00000000) // New low
	ticker := ts.UpdateTicker("ETHUSDT", 3050_00000000, 3_00000000)

	if ticker.HighE8 != 3100_00000000 {
		t.Errorf("expected high 310000000000, got %d", ticker.HighE8)
	}
	if ticker.LowE8 != 2900_00000000 {
		t.Errorf("expected low 290000000000, got %d", ticker.LowE8)
	}
	if ticker.TradeCount != 4 {
		t.Errorf("expected 4 trades, got %d", ticker.TradeCount)
	}
	if ticker.LastPriceE8 != 3050_00000000 {
		t.Errorf("expected last price 305000000000, got %d", ticker.LastPriceE8)
	}
}

func TestTickerService_GetTicker_NotFound(t *testing.T) {
	ts := NewTickerService()

	_, ok := ts.GetTicker("NONEXISTENT")
	if ok {
		t.Error("expected not found for nonexistent symbol")
	}
}

func TestTickerService_ChangePct(t *testing.T) {
	ts := NewTickerService()

	ts.UpdateTicker("BTCUSDT", 1000_00000000, 1_00000000) // Open at 1000
	ticker := ts.UpdateTicker("BTCUSDT", 1100_00000000, 1_00000000) // Now 1100 = +10%

	expectedChange := 10.0
	if ticker.ChangePct != expectedChange {
		t.Errorf("expected change %.2f%%, got %.2f%%", expectedChange, ticker.ChangePct)
	}
}

// ──────────────────────────────────────────────────
// Trade History Tests
// ──────────────────────────────────────────────────

func TestTradeHistory_RecordAndRetrieve(t *testing.T) {
	store := NewTradeHistoryStore(100)

	store.RecordTrade(TradeRecord{TradeID: "T-1", Symbol: "BTCUSDT", PriceE8: 6500e8, QuantityE8: 1e8, ExecutedAt: time.Now()})
	store.RecordTrade(TradeRecord{TradeID: "T-2", Symbol: "BTCUSDT", PriceE8: 6510e8, QuantityE8: 2e8, ExecutedAt: time.Now()})

	trades := store.GetRecentTrades("BTCUSDT", 5)
	if len(trades) != 2 {
		t.Errorf("expected 2 trades, got %d", len(trades))
	}
}

func TestTradeHistory_Eviction(t *testing.T) {
	store := NewTradeHistoryStore(3) // Max 3 trades

	for i := 0; i < 5; i++ {
		store.RecordTrade(TradeRecord{
			TradeID:  fmt.Sprintf("T-%d", i),
			Symbol:   "ETHUSDT",
			PriceE8:  uint64(3000+i) * 1e8,
			ExecutedAt: time.Now(),
		})
	}

	trades := store.GetRecentTrades("ETHUSDT", 10)
	if len(trades) != 3 {
		t.Errorf("expected 3 trades after eviction, got %d", len(trades))
	}
	// Most recent should be T-4
	if trades[2].TradeID != "T-4" {
		t.Errorf("expected most recent trade T-4, got %s", trades[2].TradeID)
	}
}

func TestTradeHistory_EmptySymbol(t *testing.T) {
	store := NewTradeHistoryStore(100)

	trades := store.GetRecentTrades("EMPTY", 10)
	if len(trades) != 0 {
		t.Errorf("expected 0 trades for unknown symbol, got %d", len(trades))
	}
}

// ──────────────────────────────────────────────────
// Depth Snapshot Tests
// ──────────────────────────────────────────────────

func TestDepthSnapshot_UpdateAndRetrieve(t *testing.T) {
	dm := NewDepthSnapshotManager()

	bids := []DepthLevel{
		{PriceE8: 6500e8, QuantityE8: 10e8, OrderCount: 3},
		{PriceE8: 6499e8, QuantityE8: 5e8, OrderCount: 2},
	}
	asks := []DepthLevel{
		{PriceE8: 6501e8, QuantityE8: 8e8, OrderCount: 2},
		{PriceE8: 6502e8, QuantityE8: 12e8, OrderCount: 5},
	}

	snap := dm.UpdateSnapshot("BTCUSDT", bids, asks)
	if snap.SequenceID != 1 {
		t.Errorf("expected sequence 1, got %d", snap.SequenceID)
	}
	if len(snap.Bids) != 2 {
		t.Errorf("expected 2 bid levels, got %d", len(snap.Bids))
	}
	if len(snap.Asks) != 2 {
		t.Errorf("expected 2 ask levels, got %d", len(snap.Asks))
	}

	retrieved, err := dm.GetSnapshot("BTCUSDT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.SequenceID != 1 {
		t.Errorf("expected sequence 1, got %d", retrieved.SequenceID)
	}
}

func TestDepthSnapshot_NotFound(t *testing.T) {
	dm := NewDepthSnapshotManager()

	_, err := dm.GetSnapshot("NONEXISTENT")
	if err == nil {
		t.Error("expected error for nonexistent snapshot")
	}
}

func TestDepthSnapshot_SequenceIncrement(t *testing.T) {
	dm := NewDepthSnapshotManager()

	dm.UpdateSnapshot("BTCUSDT", nil, nil)
	snap := dm.UpdateSnapshot("BTCUSDT", nil, nil)

	if snap.SequenceID != 2 {
		t.Errorf("expected sequence 2, got %d", snap.SequenceID)
	}
}

// ──────────────────────────────────────────────────
// OHLCV Aggregator Tests (existing code)
// ──────────────────────────────────────────────────

func TestOHLCVAggregator_SingleTrade(t *testing.T) {
	agg := NewOHLCVAggregator()
	now := time.Now()

	candle := agg.IngestTrade("BTCUSDT", 6500_00000000, 1_00000000, now)
	if candle.OpenE8 != 6500_00000000 {
		t.Errorf("expected open 650000000000, got %d", candle.OpenE8)
	}
	if candle.Trades != 1 {
		t.Errorf("expected 1 trade, got %d", candle.Trades)
	}
}

func TestOHLCVAggregator_MultipleTradesInBucket(t *testing.T) {
	agg := NewOHLCVAggregator()
	now := time.Now().Truncate(time.Minute)

	agg.IngestTrade("BTCUSDT", 6500_00000000, 1_00000000, now)
	agg.IngestTrade("BTCUSDT", 6600_00000000, 2_00000000, now.Add(10*time.Second))
	candle := agg.IngestTrade("BTCUSDT", 6400_00000000, 1_50000000, now.Add(20*time.Second))

	if candle.HighE8 != 6600_00000000 {
		t.Errorf("expected high 660000000000, got %d", candle.HighE8)
	}
	if candle.LowE8 != 6400_00000000 {
		t.Errorf("expected low 640000000000, got %d", candle.LowE8)
	}
	if candle.CloseE8 != 6400_00000000 {
		t.Errorf("expected close 640000000000, got %d", candle.CloseE8)
	}
	if candle.Trades != 3 {
		t.Errorf("expected 3 trades, got %d", candle.Trades)
	}
}

func TestOHLCVAggregator_NewBucketReset(t *testing.T) {
	agg := NewOHLCVAggregator()
	t1 := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 9, 20, 10, 1, 0, 0, time.UTC) // Next minute

	agg.IngestTrade("BTCUSDT", 6500_00000000, 1_00000000, t1)
	candle := agg.IngestTrade("BTCUSDT", 6700_00000000, 2_00000000, t2)

	// New candle should reset
	if candle.OpenE8 != 6700_00000000 {
		t.Errorf("expected new candle open 670000000000, got %d", candle.OpenE8)
	}
	if candle.Trades != 1 {
		t.Errorf("expected 1 trade in new candle, got %d", candle.Trades)
	}
}
