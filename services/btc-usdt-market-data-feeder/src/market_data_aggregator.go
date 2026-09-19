package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// PriceSource identifies a market data source
type PriceSource string

const (
	SourceBinance  PriceSource = "BINANCE"
	SourceCoinbase PriceSource = "COINBASE"
	SourceKraken   PriceSource = "KRAKEN"
	SourceOKX      PriceSource = "OKX"
	SourceOracle   PriceSource = "ORACLE"
)

// SourceTick represents a price tick from a specific data source
type SourceTick struct {
	Source      PriceSource `json:"source"`
	PriceUSD   float64     `json:"price_usd"`
	VolumeUSD  float64     `json:"volume_usd"`
	TimestampMs int64      `json:"timestamp_ms"`
}

// AggregatedPrice is the final computed multi-source price
type AggregatedPrice struct {
	Symbol     string      `json:"symbol"`
	VWAP       float64     `json:"vwap"`
	TWAP       float64     `json:"twap"`
	OraclePrice float64    `json:"oracle_price"`
	MedianPrice float64    `json:"median_price"`
	SourceCount int        `json:"source_count"`
	Timestamp  time.Time   `json:"timestamp"`
}

// VWAPWindow holds time-windowed volume-weighted data
type VWAPWindow struct {
	SumPriceVolume float64
	SumVolume      float64
}

// TWAPWindow holds time-weighted price accumulation
type TWAPWindow struct {
	Prices []float64
}

// MarketDataAggregator aggregates BTC/USDT data from multiple sources
type MarketDataAggregator struct {
	mu           sync.RWMutex
	sources      map[PriceSource]SourceTick
	vwap         VWAPWindow
	twap         TWAPWindow
	oraclePrice  float64
	maxTWAPSamples int
}

// NewMarketDataAggregator creates a new multi-source BTC/USDT aggregator
func NewMarketDataAggregator(maxTWAPSamples int) *MarketDataAggregator {
	if maxTWAPSamples <= 0 {
		maxTWAPSamples = 100
	}
	return &MarketDataAggregator{
		sources:        make(map[PriceSource]SourceTick),
		maxTWAPSamples: maxTWAPSamples,
	}
}

// IngestSourceTick processes an incoming price tick from a data source
func (a *MarketDataAggregator) IngestSourceTick(tick SourceTick) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.sources[tick.Source] = tick

	// Update VWAP accumulator
	a.vwap.SumPriceVolume += tick.PriceUSD * tick.VolumeUSD
	a.vwap.SumVolume += tick.VolumeUSD

	// Update TWAP sliding window
	a.twap.Prices = append(a.twap.Prices, tick.PriceUSD)
	if len(a.twap.Prices) > a.maxTWAPSamples {
		a.twap.Prices = a.twap.Prices[1:]
	}

	if tick.Source == SourceOracle {
		a.oraclePrice = tick.PriceUSD
	}
}

// GetVWAP returns the volume-weighted average price
func (a *MarketDataAggregator) GetVWAP() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.vwap.SumVolume == 0 {
		return 0
	}
	return a.vwap.SumPriceVolume / a.vwap.SumVolume
}

// GetTWAP returns the time-weighted average price
func (a *MarketDataAggregator) GetTWAP() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if len(a.twap.Prices) == 0 {
		return 0
	}
	var sum float64
	for _, p := range a.twap.Prices {
		sum += p
	}
	return sum / float64(len(a.twap.Prices))
}

// GetMedianPrice returns the median price across all active sources
func (a *MarketDataAggregator) GetMedianPrice() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var prices []float64
	for _, tick := range a.sources {
		prices = append(prices, tick.PriceUSD)
	}
	if len(prices) == 0 {
		return 0
	}

	// Simple sort for median
	for i := 0; i < len(prices); i++ {
		for j := i + 1; j < len(prices); j++ {
			if prices[j] < prices[i] {
				prices[i], prices[j] = prices[j], prices[i]
			}
		}
	}

	mid := len(prices) / 2
	if len(prices)%2 == 0 {
		return (prices[mid-1] + prices[mid]) / 2.0
	}
	return prices[mid]
}

// GetAggregatedPrice returns the full multi-source aggregated price
func (a *MarketDataAggregator) GetAggregatedPrice() AggregatedPrice {
	return AggregatedPrice{
		Symbol:      "BTC/USDT",
		VWAP:        a.GetVWAP(),
		TWAP:        a.GetTWAP(),
		OraclePrice: a.getOraclePrice(),
		MedianPrice: a.GetMedianPrice(),
		SourceCount: a.getSourceCount(),
		Timestamp:   time.Now().UTC(),
	}
}

func (a *MarketDataAggregator) getOraclePrice() float64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.oraclePrice
}

func (a *MarketDataAggregator) getSourceCount() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.sources)
}

// ResetVWAP resets the VWAP accumulator for a new calculation window
func (a *MarketDataAggregator) ResetVWAP() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.vwap = VWAPWindow{}
}

// DetectPriceDeviation checks if any source deviates more than maxPct from the median
func (a *MarketDataAggregator) DetectPriceDeviation(maxPct float64) []string {
	median := a.GetMedianPrice()
	if median == 0 {
		return nil
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	var deviants []string
	for src, tick := range a.sources {
		deviation := math.Abs(tick.PriceUSD-median) / median * 100.0
		if deviation > maxPct {
			deviants = append(deviants, fmt.Sprintf("%s: %.2f%% deviation", src, deviation))
		}
	}
	return deviants
}
