package parquet_test

import (
	"fmt"
	"testing"
	"time"

	"growww/database/parquet"
)

func TestParquet_ColumnarSerializationRoundtrip(t *testing.T) {
	batch := parquet.NewColumnarTradeBatch(100)
	baseTime := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)

	symbols := []string{"BTC-USDT", "ETH-USDT", "SOL-USDT"}
	sides := []string{"BUY", "SELL"}

	for i := 0; i < 50; i++ {
		batch.Append(parquet.TradeRecord{
			TradeID:         fmt.Sprintf("trd_%d", i),
			TimestampNs:     baseTime.Add(time.Duration(i) * time.Second).UnixNano(),
			Symbol:          symbols[i%len(symbols)],
			PriceE8:         uint64(50000+i*10) * 100000000,
			QuantityE8:      uint64(1+i%5) * 100000000,
			Side:            sides[i%len(sides)],
			MakerOrderID:    fmt.Sprintf("maker_%d", i),
			TakerOrderID:    fmt.Sprintf("taker_%d", i),
			FeeE8:           500000,
			SettlementBlock: uint64(1000 + i),
		})
	}

	// Serialize
	block, err := parquet.SerializeColumnarTrades(batch)
	if err != nil {
		t.Fatalf("SerializeColumnarTrades failed: %v", err)
	}

	if block.RowCount != 50 {
		t.Fatalf("expected 50 rows, got %d", block.RowCount)
	}
	if len(block.CompressedBytes) == 0 {
		t.Fatalf("expected non-empty compressed bytes")
	}

	// Deserialize
	deserialized, err := parquet.DeserializeColumnarTrades(block.CompressedBytes)
	if err != nil {
		t.Fatalf("DeserializeColumnarTrades failed: %v", err)
	}

	if deserialized.Len() != 50 {
		t.Fatalf("expected 50 deserialized rows, got %d", deserialized.Len())
	}

	for i := 0; i < 50; i++ {
		if deserialized.Symbols[i] != symbols[i%len(symbols)] {
			t.Errorf("row %d: expected symbol %s, got %s", i, symbols[i%len(symbols)], deserialized.Symbols[i])
		}
		if deserialized.Sides[i] != sides[i%len(sides)] {
			t.Errorf("row %d: expected side %s, got %s", i, sides[i%len(sides)], deserialized.Sides[i])
		}
		expectedPrice := uint64(50000+i*10) * 100000000
		if deserialized.PricesE8[i] != expectedPrice {
			t.Errorf("row %d: expected price %d, got %d", i, expectedPrice, deserialized.PricesE8[i])
		}
	}
}

func TestParquet_StorageEngineAndPredicatePushdown(t *testing.T) {
	engine := parquet.NewColumnarParquetStorageEngine()
	exportTime := time.Date(2026, 9, 20, 11, 0, 0, 0, time.UTC)

	batch := parquet.NewColumnarTradeBatch(20)
	// Add trades: 10 BUYs below 60000, 10 BUYs above 60000
	for i := 0; i < 10; i++ {
		batch.Append(parquet.TradeRecord{
			TradeID:     fmt.Sprintf("t_low_%d", i),
			TimestampNs: exportTime.UnixNano(),
			Symbol:      "BTC-USDT",
			PriceE8:     55000_00000000,
			QuantityE8:  1_00000000,
			Side:        "BUY",
		})
	}
	for i := 0; i < 10; i++ {
		batch.Append(parquet.TradeRecord{
			TradeID:     fmt.Sprintf("t_high_%d", i),
			TimestampNs: exportTime.UnixNano(),
			Symbol:      "BTC-USDT",
			PriceE8:     65000_00000000,
			QuantityE8:  2_00000000,
			Side:        "SELL",
		})
	}

	manifest, err := engine.ExportTradeBatch("trades", "BTC-USDT", exportTime, batch)
	if err != nil {
		t.Fatalf("ExportTradeBatch failed: %v", err)
	}

	if manifest.SHA256Checksum == "" {
		t.Errorf("expected SHA-256 checksum in manifest")
	}

	// Query with pushdown filter: price > 60000 and side == SELL
	filtered, err := engine.QueryByPredicate(manifest.PartitionPath, 60000_00000000, 0, "SELL")
	if err != nil {
		t.Fatalf("QueryByPredicate failed: %v", err)
	}

	if filtered.Len() != 10 {
		t.Fatalf("expected 10 filtered rows, got %d", filtered.Len())
	}

	for i := 0; i < filtered.Len(); i++ {
		if filtered.PricesE8[i] != 65000_00000000 || filtered.Sides[i] != "SELL" {
			t.Errorf("unexpected record in filtered output: price=%d, side=%s", filtered.PricesE8[i], filtered.Sides[i])
		}
	}
}

func TestParquet_AnalyticsEngineComputations(t *testing.T) {
	analytics := parquet.NewParquetAnalyticsEngine()
	batch := parquet.NewColumnarTradeBatch(3)

	batch.Append(parquet.TradeRecord{PriceE8: 100_00000000, QuantityE8: 1_00000000, Side: "BUY"})
	batch.Append(parquet.TradeRecord{PriceE8: 200_00000000, QuantityE8: 1_00000000, Side: "BUY"})
	batch.Append(parquet.TradeRecord{PriceE8: 300_00000000, QuantityE8: 2_00000000, Side: "SELL"})

	res, err := analytics.ComputeTradeAnalytics(batch)
	if err != nil {
		t.Fatalf("ComputeTradeAnalytics failed: %v", err)
	}

	if res.TradeCount != 3 {
		t.Errorf("expected 3 trades, got %d", res.TradeCount)
	}
	if res.HighPriceE8 != 300_00000000 {
		t.Errorf("expected High 300, got %d", res.HighPriceE8)
	}
	if res.LowPriceE8 != 100_00000000 {
		t.Errorf("expected Low 100, got %d", res.LowPriceE8)
	}
	if res.TotalVolumeE8 != 4_00000000 {
		t.Errorf("expected TotalVolume 4, got %d", res.TotalVolumeE8)
	}

	// quote = 100*1 + 200*1 + 300*2 = 900
	// vwap = 900 / 4 = 225
	expectedVWAP := uint64(225_00000000)
	if res.VWAPE8 != expectedVWAP {
		t.Errorf("expected VWAP %d, got %d", expectedVWAP, res.VWAPE8)
	}

	if res.BuyVolumeE8 != 2_00000000 || res.SellVolumeE8 != 2_00000000 {
		t.Errorf("expected 2 BUY vol and 2 SELL vol, got %d / %d", res.BuyVolumeE8, res.SellVolumeE8)
	}
	if res.BuyTakerRatio != 0.5 {
		t.Errorf("expected 0.5 buy ratio, got %f", res.BuyTakerRatio)
	}
}
