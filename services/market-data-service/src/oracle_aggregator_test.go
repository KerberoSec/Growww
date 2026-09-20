package main

import (
	"testing"
	"time"
)

func TestOracleAggregator_ArbitratedMedian(t *testing.T) {
	agg := NewOracleAggregator(10, 200) // 10s max stale, 200 bps (2.0%) max deviation

	now := uint64(time.Now().UnixMilli())

	// Record 3 valid feeds for BTC/USDT
	agg.RecordObservation(PriceFeedObservation{
		Source:      SourcePyth,
		Symbol:      "BTC/USDT",
		PriceE8:     64000_00000000,
		Confidence:  100,
		TimestampMs: now - 500,
	})
	agg.RecordObservation(PriceFeedObservation{
		Source:      SourceChainlink,
		Symbol:      "BTC/USDT",
		PriceE8:     64100_00000000,
		Confidence:  100,
		TimestampMs: now - 300,
	})
	agg.RecordObservation(PriceFeedObservation{
		Source:      SourceBinance,
		Symbol:      "BTC/USDT",
		PriceE8:     64050_00000000,
		Confidence:  100,
		TimestampMs: now - 100,
	})

	price, err := agg.GetArbitratedPrice("BTC/USDT")
	if err != nil {
		t.Fatalf("expected valid arbitrated price, got: %v", err)
	}

	if price.MedianPriceE8 != 64050_00000000 {
		t.Fatalf("expected median 64050.00, got %d", price.MedianPriceE8)
	}
	if price.ContributingFeeds != 3 {
		t.Fatalf("expected 3 contributing feeds, got %d", price.ContributingFeeds)
	}
}

func TestOracleAggregator_StaleFeedRejection(t *testing.T) {
	agg := NewOracleAggregator(2, 200) // 2s max stale

	now := uint64(time.Now().UnixMilli())

	// Stale Pyth feed (5 seconds old)
	agg.RecordObservation(PriceFeedObservation{
		Source:      SourcePyth,
		Symbol:      "ETH/USDT",
		PriceE8:     3500_00000000,
		TimestampMs: now - 5000,
	})
	// Fresh Binance feed
	agg.RecordObservation(PriceFeedObservation{
		Source:      SourceBinance,
		Symbol:      "ETH/USDT",
		PriceE8:     3520_00000000,
		TimestampMs: now - 200,
	})

	price, err := agg.GetArbitratedPrice("ETH/USDT")
	if err != nil {
		t.Fatalf("expected valid price from remaining fresh feed, got: %v", err)
	}

	if price.ContributingFeeds != 1 {
		t.Fatalf("expected exactly 1 fresh contributing feed, got %d", price.ContributingFeeds)
	}
	if price.MedianPriceE8 != 3520_00000000 {
		t.Fatalf("expected 3520.00, got %d", price.MedianPriceE8)
	}
}

func TestOracleAggregator_DivergentFeedError(t *testing.T) {
	agg := NewOracleAggregator(10, 100) // 100 bps (1.0%) max deviation

	now := uint64(time.Now().UnixMilli())

	// Feed 1: Normal price
	agg.RecordObservation(PriceFeedObservation{
		Source:      SourcePyth,
		Symbol:      "SOL/USDT",
		PriceE8:     150_00000000,
		TimestampMs: now - 100,
	})
	// Feed 2: 10% spiked manipulated price (1000 bps deviation)
	agg.RecordObservation(PriceFeedObservation{
		Source:      SourceBinance,
		Symbol:      "SOL/USDT",
		PriceE8:     165_00000000,
		TimestampMs: now - 100,
	})

	_, err := agg.GetArbitratedPrice("SOL/USDT")
	if err == nil {
		t.Fatalf("expected error due to excessive feed dispersion (> 100 bps)")
	}
}
