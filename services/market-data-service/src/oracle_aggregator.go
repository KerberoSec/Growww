package main

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// OracleSource represents the market data origin
type OracleSource string

const (
	SourceChainlink OracleSource = "CHAINLINK"
	SourcePyth      OracleSource = "PYTH"
	SourceBinance   OracleSource = "BINANCE"
)

// PriceFeedObservation represents an individual price report
type PriceFeedObservation struct {
	Source      OracleSource
	Symbol      string
	PriceE8     uint64 // Scaled 1e8
	Confidence  uint64
	TimestampMs uint64
}

// AggregatedPrice represents the consolidated arbitrated price
type AggregatedPrice struct {
	Symbol           string
	MedianPriceE8    uint64
	ContributingFeeds int
	DispersionBps    uint32
	TimestampMs      uint64
	Stale            bool
}

// OracleAggregator manages feed collection and robust median arbitration
type OracleAggregator struct {
	mu           sync.RWMutex
	maxStaleMs   uint64
	maxDevBps    uint32
	observations map[string]map[OracleSource]PriceFeedObservation
}

func NewOracleAggregator(maxStaleSeconds uint64, maxDeviationBps uint32) *OracleAggregator {
	if maxStaleSeconds == 0 {
		maxStaleSeconds = 15 // Default 15s staleness
	}
	if maxDeviationBps == 0 {
		maxDeviationBps = 150 // Default 1.5% max deviation
	}
	return &OracleAggregator{
		maxStaleMs:   maxStaleSeconds * 1000,
		maxDevBps:    maxDeviationBps,
		observations: make(map[string]map[OracleSource]PriceFeedObservation),
	}
}

// RecordObservation records an incoming tick from an oracle
func (a *OracleAggregator) RecordObservation(obs PriceFeedObservation) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.observations[obs.Symbol]; !exists {
		a.observations[obs.Symbol] = make(map[OracleSource]PriceFeedObservation)
	}
	a.observations[obs.Symbol][obs.Source] = obs
}

// GetArbitratedPrice computes the resilient median price across healthy non-stale feeds
func (a *OracleAggregator) GetArbitratedPrice(symbol string) (*AggregatedPrice, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	feeds, exists := a.observations[symbol]
	if !exists || len(feeds) == 0 {
		return nil, errors.New("no oracle feeds registered for symbol")
	}

	nowMs := uint64(time.Now().UnixMilli())
	var validPrices []uint64

	for _, obs := range feeds {
		if nowMs > obs.TimestampMs && (nowMs-obs.TimestampMs) > a.maxStaleMs {
			continue // Discard stale feed
		}
		validPrices = append(validPrices, obs.PriceE8)
	}

	if len(validPrices) == 0 {
		return nil, errors.New("all oracle feeds are stale")
	}

	sort.Slice(validPrices, func(i, j int) bool {
		return validPrices[i] < validPrices[j]
	})

	var median uint64
	n := len(validPrices)
	if n%2 == 1 {
		median = validPrices[n/2]
	} else {
		median = (validPrices[n/2-1] + validPrices[n/2]) / 2
	}

	// Calculate maximum dispersion (deviation from median)
	var maxDispersionBps uint32
	for _, p := range validPrices {
		diff := math.Abs(float64(p) - float64(median))
		dispBps := uint32((diff / float64(median)) * 10000)
		if dispBps > maxDispersionBps {
			maxDispersionBps = dispBps
		}
	}

	if maxDispersionBps > a.maxDevBps {
		return nil, fmt.Errorf("oracle feed deviation exceeded limit: %d bps > %d bps", maxDispersionBps, a.maxDevBps)
	}

	return &AggregatedPrice{
		Symbol:            symbol,
		MedianPriceE8:     median,
		ContributingFeeds: n,
		DispersionBps:     maxDispersionBps,
		TimestampMs:       nowMs,
		Stale:             false,
	}, nil
}

func main() {
	_ = NewOracleAggregator(15, 150)
	fmt.Println("Oracle Aggregator & Feed Arbiter Service running.")
}
