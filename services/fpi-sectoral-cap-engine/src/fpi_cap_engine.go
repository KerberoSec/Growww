package main

import (
	"errors"
	"fmt"
	"sync"
)

type SectoralCapRule struct {
	ISIN              string
	CompanySymbol     string
	Sector            string
	MaxFPICapPct      float64 // e.g. 74% for banking, 100% for telecom
	CurrentFPIHoldPct float64
	RedFlagThreshold  float64 // 3% below sectoral cap
}

type FPICapGuardEngine struct {
	mu    sync.RWMutex
	rules map[string]*SectoralCapRule // ISIN -> Rule
}

func NewFPICapGuardEngine() *FPICapGuardEngine {
	return &FPICapGuardEngine{
		rules: make(map[string]*SectoralCapRule),
	}
}

// ValidateFPIBuyOrder validates order against real-time FEMA / RBI aggregate foreign investment limit
func (e *FPICapGuardEngine) ValidateFPIBuyOrder(isin string, requestedShares, totalPaidUpCapital uint64) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rule, exists := e.rules[isin]
	if !exists {
		return nil // No specific sectoral cap restriction
	}

	incrementalPct := (float64(requestedShares) / float64(totalPaidUpCapital)) * 100.0
	projectedPct := rule.CurrentFPIHoldPct + incrementalPct

	if projectedPct >= rule.MaxFPICapPct {
		return fmt.Errorf("FPI sectoral cap breached for %s: projected %.2f%% >= statutory cap %.2f%%",
			rule.CompanySymbol, projectedPct, rule.MaxFPICapPct)
	}

	if projectedPct >= (rule.MaxFPICapPct - rule.RedFlagThreshold) {
		fmt.Printf("[FPI Surveillance] Warning: %s entered RBI Red Flag threshold (%.2f%%)\n",
			rule.CompanySymbol, projectedPct)
	}

	return nil
}
