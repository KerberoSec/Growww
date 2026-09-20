package main

import (
	"errors"
	"testing"
	"time"
)

func TestRollingWindow_VWAPAndHighLow(t *testing.T) {
	calc := NewRolling24hWindowCalculator(24 * time.Hour)
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)

	// Trade 1: 1 BTC at $60,000 (60,000e8)
	tick1 := TradeTick{
		TradeID:    "T1",
		PriceE8:    60000_00000000,
		QuantityE8: 1_00000000,
		Timestamp:  now.Add(-2 * time.Hour),
	}
	stats, err := calc.IngestTrade("BTCUSDT", tick1, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.High24hE8 != 60000_00000000 || stats.Low24hE8 != 60000_00000000 {
		t.Fatalf("unexpected high/low: %d / %d", stats.High24hE8, stats.Low24hE8)
	}
	if stats.Vwap24hE8 != 60000_00000000 {
		t.Fatalf("unexpected VWAP: %d", stats.Vwap24hE8)
	}
	if stats.PriceChangePct != 0.0 {
		t.Fatalf("unexpected change pct: %.2f", stats.PriceChangePct)
	}

	// Trade 2: 3 BTC at $70,000
	// Total volume: 4 BTC.
	// Total Notional: (1 * 60,000) + (3 * 70,000) = 270,000 USD
	// Expected VWAP: 270,000 / 4 = $67,500
	tick2 := TradeTick{
		TradeID:    "T2",
		PriceE8:    70000_00000000,
		QuantityE8: 3_00000000,
		Timestamp:  now.Add(-1 * time.Hour),
	}
	stats, err = calc.IngestTrade("BTCUSDT", tick2, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.High24hE8 != 70000_00000000 || stats.Low24hE8 != 60000_00000000 {
		t.Fatalf("unexpected high/low: %d / %d", stats.High24hE8, stats.Low24hE8)
	}
	if stats.Vwap24hE8 != 67500_00000000 {
		t.Fatalf("expected VWAP 67500_00000000, got %d", stats.Vwap24hE8)
	}
	// Change pct: (70000 - 60000)/60000 * 100 = 16.6667%
	if stats.PriceChangePct < 16.66 || stats.PriceChangePct > 16.67 {
		t.Fatalf("expected ~16.67%% change, got %.4f%%", stats.PriceChangePct)
	}
}

func TestRollingWindow_EvictionAfter24Hours(t *testing.T) {
	calc := NewRolling24hWindowCalculator(24 * time.Hour)
	baseTime := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)

	// Ingest tick at 23 hours ago
	tickOld := TradeTick{
		TradeID:    "T-OLD",
		PriceE8:    50000_00000000,
		QuantityE8: 2_00000000,
		Timestamp:  baseTime.Add(-23 * time.Hour),
	}
	_, err := calc.IngestTrade("BTCUSDT", tickOld, baseTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ingest tick at 1 hour ago
	tickNew := TradeTick{
		TradeID:    "T-NEW",
		PriceE8:    65000_00000000,
		QuantityE8: 1_00000000,
		Timestamp:  baseTime.Add(-1 * time.Hour),
	}
	_, err = calc.IngestTrade("BTCUSDT", tickNew, baseTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Advance time by 2 hours -> tickOld is now 25 hours old, should be evicted
	nowAdvanced := baseTime.Add(2 * time.Hour)
	stats, found := calc.GetStats("BTCUSDT", nowAdvanced)
	if !found {
		t.Fatal("expected stats to be found")
	}

	// Only tickNew remains in window!
	if stats.TradeCount24h != 1 {
		t.Fatalf("expected 1 tick after eviction, got %d", stats.TradeCount24h)
	}
	if stats.OpenPrice24hE8 != 65000_00000000 || stats.LastPriceE8 != 65000_00000000 {
		t.Fatalf("expected open and last to be 65000_00000000, got %d", stats.OpenPrice24hE8)
	}
	if stats.Vwap24hE8 != 65000_00000000 {
		t.Fatalf("expected VWAP to be 65000_00000000, got %d", stats.Vwap24hE8)
	}
}

func TestRollingWindow_AnomalousTicksRejected(t *testing.T) {
	calc := NewRolling24hWindowCalculator(24 * time.Hour)
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)

	// Future tick (> 5 seconds ahead)
	futureTick := TradeTick{
		TradeID:    "T-FUT",
		PriceE8:    60000_00000000,
		QuantityE8: 1_00000000,
		Timestamp:  now.Add(10 * time.Second),
	}
	_, err := calc.IngestTrade("BTCUSDT", futureTick, now)
	if !errors.Is(err, ErrFutureTimestamp) {
		t.Fatalf("expected ErrFutureTimestamp, got %v", err)
	}

	// Stale tick (> 24 hours old)
	staleTick := TradeTick{
		TradeID:    "T-STALE",
		PriceE8:    60000_00000000,
		QuantityE8: 1_00000000,
		Timestamp:  now.Add(-25 * time.Hour),
	}
	_, err = calc.IngestTrade("BTCUSDT", staleTick, now)
	if !errors.Is(err, ErrStaleTimestamp) {
		t.Fatalf("expected ErrStaleTimestamp, got %v", err)
	}

	// Zero price tick
	zeroPriceTick := TradeTick{
		TradeID:    "T-ZERO",
		PriceE8:    0,
		QuantityE8: 1_00000000,
		Timestamp:  now,
	}
	_, err = calc.IngestTrade("BTCUSDT", zeroPriceTick, now)
	if !errors.Is(err, ErrInvalidTradeData) {
		t.Fatalf("expected ErrInvalidTradeData, got %v", err)
	}
}
