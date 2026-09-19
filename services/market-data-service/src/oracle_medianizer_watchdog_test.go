package main

import "testing"

func TestOracleMedianizer_OutlierClamping(t *testing.T) {
	w := NewMedianizerWatchdog(0.03) // 3% clamp

	ticks := []OraclePriceTick{
		{Source: "Binance", Price: 65000},
		{Source: "Coinbase", Price: 65100},
		{Source: "Kraken", Price: 64950},
		{Source: "ManipulatedOracle", Price: 95000}, // Rogue spike
	}

	median, valid, err := w.ComputeMedianAndFilterClamped(ticks)
	if err != nil {
		t.Fatal(err)
	}

	if median < 64950 || median > 65100 {
		t.Errorf("median out of expected range: %f", median)
	}
	if len(valid) != 3 {
		t.Errorf("expected 3 valid ticks (1 rogue clamped), got %d", len(valid))
	}
}
