package timescaledb_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"growww/database/timescaledb"
)

func TestCompressionManager_LifecycleAndBackfill(t *testing.T) {
	engine := timescaledb.NewTimescaleCandleEngine()
	ctx := context.Background()

	bucketTime := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	tick := timescaledb.TradeTick{
		TradeID:   "t-comp-1",
		Symbol:    "BTC-USDT",
		PriceE8:   55000_00000000,
		QuantityE8: 10_00000000,
		Timestamp: bucketTime,
	}

	if err := engine.IngestTradeTick(ctx, tick); err != nil {
		t.Fatalf("IngestTradeTick failed: %v", err)
	}

	_, err := engine.FinalizeCandle("BTC-USDT", 60, bucketTime)
	if err != nil {
		t.Fatalf("FinalizeCandle failed: %v", err)
	}

	chunkStart := bucketTime.Truncate(24 * time.Hour)
	chunkID := fmt.Sprintf("chunk_BTC-USDT_60_%d", chunkStart.Unix())

	compMgr := timescaledb.NewCompressionManager(engine)

	// Compress chunk
	chunk, err := compMgr.CompressChunk(chunkID)
	if err != nil {
		t.Fatalf("CompressChunk failed: %v", err)
	}

	if chunk.Status != timescaledb.ChunkStatusCompressed {
		t.Fatalf("expected chunk status COMPRESSED, got %s", chunk.Status)
	}
	if chunk.MerkleRoot == "" {
		t.Errorf("expected Merkle root to be calculated")
	}
	if chunk.CompressionRatio <= 0 {
		t.Errorf("expected positive compression ratio, got %f", chunk.CompressionRatio)
	}

	// Double compress should fail
	_, err = compMgr.CompressChunk(chunkID)
	if err == nil {
		t.Errorf("expected error compressing already compressed chunk")
	}

	// Insert historical backfill
	backfillCandle := timescaledb.Candle{
		Bucket:      bucketTime.Add(5 * time.Minute),
		Symbol:      "BTC-USDT",
		IntervalSec: 60,
		OpenE8:      55100_00000000,
		HighE8:      55200_00000000,
		LowE8:       55050_00000000,
		CloseE8:     55150_00000000,
		VolumeE8:    5_00000000,
		Finalized:   true,
	}
	backfillCandle.StateHash = backfillCandle.ComputeStateHash()

	err = compMgr.InsertHistoricalBackfill(chunkID, backfillCandle)
	if err != nil {
		t.Fatalf("InsertHistoricalBackfill failed: %v", err)
	}

	// Verify decompressed content
	decompressed, err := compMgr.DecompressChunkForBackfill(chunkID)
	if err != nil {
		t.Fatalf("DecompressChunkForBackfill failed: %v", err)
	}

	if decompressed.RowCount != 2 {
		t.Errorf("expected row count 2 after backfill, got %d", decompressed.RowCount)
	}
}
