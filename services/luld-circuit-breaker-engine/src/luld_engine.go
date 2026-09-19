package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type LULDBands struct {
	Symbol          string
	ReferencePrice  float64
	LowerBandPrice  float64 // Reference - 5% (Tier 1) / 10% (Tier 2)
	UpperBandPrice  float64 // Reference + 5%
	InStraddleState bool
	StateExpiresAt  time.Time
}

type LULDCircuitBreakerEngine struct {
	mu    sync.RWMutex
	bands map[string]*LULDBands
}

func NewLULDCircuitBreakerEngine() *LULDCircuitBreakerEngine {
	return &LULDCircuitBreakerEngine{
		bands: make(map[string]*LULDBands),
	}
}

// UpdateReferencePrice recalculates dynamic 5-minute rolling VWAP price bands
func (e *LULDCircuitBreakerEngine) UpdateReferencePrice(symbol string, vwap float64, bandPct float64) *LULDBands {
	e.mu.Lock()
	defer e.mu.Unlock()

	b := &LULDBands{
		Symbol:         symbol,
		ReferencePrice: vwap,
		LowerBandPrice: vwap * (1.0 - bandPct/100.0),
		UpperBandPrice: vwap * (1.0 + bandPct/100.0),
	}
	e.bands[symbol] = b
	return b
}

// ValidateOrderPrice verifies order does not violate Limit-Up / Limit-Down bands
func (e *LULDCircuitBreakerEngine) ValidateOrderPrice(symbol string, price float64) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	b, exists := e.bands[symbol]
	if !exists {
		return nil
	}

	if price > b.UpperBandPrice {
		return fmt.Errorf("order price %.2f breached Limit-Up threshold %.2f", price, b.UpperBandPrice)
	}
	if price < b.LowerBandPrice {
		return fmt.Errorf("order price %.2f breached Limit-Down threshold %.2f", price, b.LowerBandPrice)
	}

	return nil
}
