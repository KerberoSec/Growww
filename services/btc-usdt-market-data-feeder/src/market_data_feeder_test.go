package main

import (
	"math"
	"testing"
	"time"
)

func TestMarketDataAggregator_VWAP(t *testing.T) {
	agg := NewMarketDataAggregator(100)

	// Source 1: price=65000, volume=1000
	agg.IngestSourceTick(SourceTick{Source: SourceBinance, PriceUSD: 65000, VolumeUSD: 1000, TimestampMs: time.Now().UnixMilli()})
	// Source 2: price=65100, volume=2000
	agg.IngestSourceTick(SourceTick{Source: SourceCoinbase, PriceUSD: 65100, VolumeUSD: 2000, TimestampMs: time.Now().UnixMilli()})

	// VWAP = (65000*1000 + 65100*2000) / (1000+2000) = (65M + 130.2M) / 3000 = 65066.667
	vwap := agg.GetVWAP()
	expected := (65000.0*1000.0 + 65100.0*2000.0) / 3000.0
	if math.Abs(vwap-expected) > 0.01 {
		t.Fatalf("Expected VWAP %.4f, got %.4f", expected, vwap)
	}
}

func TestMarketDataAggregator_TWAP(t *testing.T) {
	agg := NewMarketDataAggregator(5) // Small window

	prices := []float64{64000, 64500, 65000, 65500, 66000}
	for _, p := range prices {
		agg.IngestSourceTick(SourceTick{Source: SourceBinance, PriceUSD: p, VolumeUSD: 100, TimestampMs: time.Now().UnixMilli()})
	}

	twap := agg.GetTWAP()
	expected := (64000.0 + 64500.0 + 65000.0 + 65500.0 + 66000.0) / 5.0
	if math.Abs(twap-expected) > 0.01 {
		t.Fatalf("Expected TWAP %.4f, got %.4f", expected, twap)
	}
}

func TestMarketDataAggregator_TWAPSlidingWindow(t *testing.T) {
	agg := NewMarketDataAggregator(3) // Window of 3

	for i := 0; i < 5; i++ {
		agg.IngestSourceTick(SourceTick{Source: SourceBinance, PriceUSD: float64(60000 + i*1000), VolumeUSD: 100, TimestampMs: time.Now().UnixMilli()})
	}

	// Window should contain last 3: 62000, 63000, 64000
	twap := agg.GetTWAP()
	expected := (62000.0 + 63000.0 + 64000.0) / 3.0
	if math.Abs(twap-expected) > 0.01 {
		t.Fatalf("Expected TWAP %.4f, got %.4f", expected, twap)
	}
}

func TestMarketDataAggregator_MedianPrice(t *testing.T) {
	agg := NewMarketDataAggregator(100)

	agg.IngestSourceTick(SourceTick{Source: SourceBinance, PriceUSD: 65000, VolumeUSD: 100, TimestampMs: time.Now().UnixMilli()})
	agg.IngestSourceTick(SourceTick{Source: SourceCoinbase, PriceUSD: 65200, VolumeUSD: 100, TimestampMs: time.Now().UnixMilli()})
	agg.IngestSourceTick(SourceTick{Source: SourceKraken, PriceUSD: 65100, VolumeUSD: 100, TimestampMs: time.Now().UnixMilli()})

	median := agg.GetMedianPrice()
	// With 3 sources sorted: 65000, 65100, 65200 -> median = 65100
	if median != 65100.0 {
		t.Fatalf("Expected median 65100, got %.4f", median)
	}
}

func TestMarketDataAggregator_OraclePrice(t *testing.T) {
	agg := NewMarketDataAggregator(100)
	agg.IngestSourceTick(SourceTick{Source: SourceOracle, PriceUSD: 65050, VolumeUSD: 0, TimestampMs: time.Now().UnixMilli()})

	ap := agg.GetAggregatedPrice()
	if ap.OraclePrice != 65050.0 {
		t.Fatalf("Expected oracle price 65050, got %.4f", ap.OraclePrice)
	}
}

func TestMarketDataAggregator_PriceDeviation(t *testing.T) {
	agg := NewMarketDataAggregator(100)

	agg.IngestSourceTick(SourceTick{Source: SourceBinance, PriceUSD: 65000, VolumeUSD: 100, TimestampMs: time.Now().UnixMilli()})
	agg.IngestSourceTick(SourceTick{Source: SourceCoinbase, PriceUSD: 65050, VolumeUSD: 100, TimestampMs: time.Now().UnixMilli()})
	agg.IngestSourceTick(SourceTick{Source: SourceKraken, PriceUSD: 70000, VolumeUSD: 100, TimestampMs: time.Now().UnixMilli()}) // Deviant

	deviants := agg.DetectPriceDeviation(2.0) // 2% threshold
	if len(deviants) == 0 {
		t.Fatal("Expected at least one deviant source")
	}
}

func TestPriceMirrorPipeline_SubscribeAndReceive(t *testing.T) {
	pipe := NewPriceMirrorPipeline()
	ch := pipe.Subscribe("demo-engine-1")

	tick := MarketTick{
		Symbol:      "BTC/USDT",
		PriceUSD:    68500.0,
		BidUSD:      68498.5,
		AskUSD:      68501.5,
		Volume24h:   1e9,
		TimestampMs: time.Now().UnixMilli(),
	}

	pipe.IngestProductionTick(tick)

	select {
	case received := <-ch:
		if received.Symbol != "BTC/USDT" {
			t.Fatalf("Expected BTC/USDT, got %s", received.Symbol)
		}
		if received.PriceUSD != 68500.0 {
			t.Fatalf("Expected price 68500, got %.2f", received.PriceUSD)
		}
	default:
		t.Fatal("Expected to receive tick on subscriber channel")
	}
}
