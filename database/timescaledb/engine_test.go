package timescaledb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"growww/database/timescaledb"
)

func TestTimescaleCandleEngine_IngestAndAggregate(t *testing.T) {
	engine := timescaledb.NewTimescaleCandleEngine()
	ctx := context.Background()

	baseTime := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)

	// Ingest 3 ticks within 1 minute
	ticks := []timescaledb.TradeTick{
		{TradeID: "t1", Symbol: "BTC-USDT", PriceE8: 60000_00000000, QuantityE8: 1_00000000, Timestamp: baseTime.Add(5 * time.Second), IsBuyerMaker: false},
		{TradeID: "t2", Symbol: "BTC-USDT", PriceE8: 60500_00000000, QuantityE8: 2_00000000, Timestamp: baseTime.Add(20 * time.Second), IsBuyerMaker: true},
		{TradeID: "t3", Symbol: "BTC-USDT", PriceE8: 59800_00000000, QuantityE8: 1_00000000, Timestamp: baseTime.Add(45 * time.Second), IsBuyerMaker: false},
	}

	for _, tick := range ticks {
		if err := engine.IngestTradeTick(ctx, tick); err != nil {
			t.Fatalf("unexpected ingest error: %v", err)
		}
	}

	// Query 1-minute candles
	candles, err := engine.QueryCandles("BTC-USDT", 60, baseTime.Add(-time.Minute), baseTime.Add(time.Hour))
	if err != nil {
		t.Fatalf("QueryCandles failed: %v", err)
	}

	if len(candles) != 1 {
		t.Fatalf("expected 1 candle, got %d", len(candles))
	}

	c := candles[0]
	if c.OpenE8 != 60000_00000000 {
		t.Errorf("expected Open %d, got %d", 60000_00000000, c.OpenE8)
	}
	if c.HighE8 != 60500_00000000 {
		t.Errorf("expected High %d, got %d", 60500_00000000, c.HighE8)
	}
	if c.LowE8 != 59800_00000000 {
		t.Errorf("expected Low %d, got %d", 59800_00000000, c.LowE8)
	}
	if c.CloseE8 != 59800_00000000 {
		t.Errorf("expected Close %d, got %d", 59800_00000000, c.CloseE8)
	}
	if c.VolumeE8 != 4_00000000 {
		t.Errorf("expected Volume %d, got %d", 4_00000000, c.VolumeE8)
	}
	if c.TradeCount != 3 {
		t.Errorf("expected TradeCount 3, got %d", c.TradeCount)
	}

	// VWAP calculation verification
	expectedVWAP := uint64(60200_00000000)
	if c.VWAPE8 != expectedVWAP {
		t.Errorf("expected VWAP %d, got %d", expectedVWAP, c.VWAPE8)
	}
}

func TestTimescaleCandleEngine_FinalizeAndCompress(t *testing.T) {
	engine := timescaledb.NewTimescaleCandleEngine()
	ctx := context.Background()

	pastTime := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	tick := timescaledb.TradeTick{
		TradeID:    "t100",
		Symbol:     "ETH-USDT",
		PriceE8:    3500_00000000,
		QuantityE8: 10_00000000,
		Timestamp:  pastTime,
	}

	if err := engine.IngestTradeTick(ctx, tick); err != nil {
		t.Fatalf("IngestTradeTick failed: %v", err)
	}

	finalized, err := engine.FinalizeCandle("ETH-USDT", 60, pastTime)
	if err != nil {
		t.Fatalf("FinalizeCandle failed: %v", err)
	}
	if !finalized.Finalized {
		t.Errorf("candle expected to be finalized")
	}

	chunkStart := pastTime.Truncate(24 * time.Hour)
	chunkID := fmt.Sprintf("chunk_ETH-USDT_60_%d", chunkStart.Unix())

	// Apply compression policy
	compMgr := timescaledb.NewCompressionManager(engine)
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	compressedCount, err := compMgr.ApplyCompressionPolicy(now)
	if err != nil {
		t.Fatalf("ApplyCompressionPolicy failed: %v", err)
	}
	if compressedCount != 1 {
		t.Fatalf("expected 1 chunk compressed, got %d", compressedCount)
	}

	metrics := engine.GetMetrics()
	if metrics.CompressedChunksCount != 1 {
		t.Errorf("expected 1 compressed chunk in metrics, got %d", metrics.CompressedChunksCount)
	}

	// Historical Backfill: decompress, insert, recompress
	histCandle := timescaledb.Candle{
		Bucket:      pastTime.Add(30 * time.Minute),
		Symbol:      "ETH-USDT",
		IntervalSec: 60,
		OpenE8:      3510_00000000,
		HighE8:      3520_00000000,
		LowE8:       3500_00000000,
		CloseE8:     3515_00000000,
		VolumeE8:    5_00000000,
		Finalized:   true,
	}

	if err := compMgr.InsertHistoricalBackfill(chunkID, histCandle); err != nil {
		t.Fatalf("InsertHistoricalBackfill failed: %v", err)
	}

	candles, err := engine.QueryCandles("ETH-USDT", 60, pastTime.Add(-time.Hour), pastTime.Add(time.Hour))
	if err != nil {
		t.Fatalf("QueryCandles failed: %v", err)
	}
	if len(candles) == 0 {
		t.Fatalf("expected to find candles in query")
	}
}

func TestTimescaleCandleEngine_DLQAndCircuitBreaker(t *testing.T) {
	engine := timescaledb.NewTimescaleCandleEngine()
	ctx := context.Background()

	// Ingest invalid tick (zero price)
	badTick := timescaledb.TradeTick{
		TradeID:   "bad-1",
		Symbol:    "BTC-USDT",
		PriceE8:   0,
		QuantityE8: 100,
		Timestamp: time.Now(),
	}

	err := engine.IngestTradeTick(ctx, badTick)
	if err == nil {
		t.Errorf("expected error for bad tick, got nil")
	}

	dlq := engine.GetDLQ()
	if len(dlq) != 1 {
		t.Fatalf("expected 1 DLQ entry, got %d", len(dlq))
	}
	if dlq[0].Tick.TradeID != "bad-1" {
		t.Errorf("expected DLQ trade ID bad-1, got %s", dlq[0].Tick.TradeID)
	}

	// Test circuit breaker
	engine.SetCircuitBreaker(true)
	validTick := timescaledb.TradeTick{
		TradeID:   "good-1",
		Symbol:    "BTC-USDT",
		PriceE8:   60000_00000000,
		QuantityE8: 100,
		Timestamp: time.Now(),
	}
	err = engine.IngestTradeTick(ctx, validTick)
	if err == nil {
		t.Errorf("expected circuit breaker error, got nil")
	}
}

func TestTimescaleCandleEngine_ReconciliationLoop(t *testing.T) {
	engine := timescaledb.NewTimescaleCandleEngine()
	ctx := context.Background()

	tick := timescaledb.TradeTick{
		TradeID:   "t-rec",
		Symbol:    "SOL-USDT",
		PriceE8:   150_00000000,
		QuantityE8: 50_00000000,
		Timestamp: time.Now(),
	}

	_ = engine.IngestTradeTick(ctx, tick)

	reconMgr := timescaledb.NewReconciliationManager(engine)
	report := reconMgr.RunReconciliationLoop()

	if !report.Consistent {
		t.Errorf("expected state to be consistent, errors: %v", report.Errors)
	}
	if report.MerkleRoot == "" {
		t.Errorf("expected non-empty MerkleRoot")
	}
	if report.BesuAnchorTxHash == "" || report.BesuBlockNumber == 0 {
		t.Errorf("expected valid Besu anchoring info")
	}
}
