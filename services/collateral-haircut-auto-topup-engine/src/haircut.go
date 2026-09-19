package main

import (
	"fmt"
	"sync"
)

type AssetHaircut struct {
	Symbol     string
	HaircutPct float64 // 0-100
	Eligible   bool
}

type CollateralPosition struct {
	UserID    string
	Symbol    string
	Quantity  float64
	PriceINR  float64
}

type HaircutEngine struct {
	mu       sync.RWMutex
	haircuts map[string]AssetHaircut
}

func NewHaircutEngine() *HaircutEngine {
	return &HaircutEngine{
		haircuts: map[string]AssetHaircut{
			"BTC":  {Symbol: "BTC", HaircutPct: 25.0, Eligible: true},
			"ETH":  {Symbol: "ETH", HaircutPct: 30.0, Eligible: true},
			"USDT": {Symbol: "USDT", HaircutPct: 5.0, Eligible: true},
			"INR":  {Symbol: "INR", HaircutPct: 0.0, Eligible: true},
		},
	}
}

func (e *HaircutEngine) EffectiveCollateral(positions []CollateralPosition) (float64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	total := 0.0
	for _, p := range positions {
		hc, ok := e.haircuts[p.Symbol]
		if !ok || !hc.Eligible {
			continue
		}
		gross := p.Quantity * p.PriceINR
		total += gross * (1 - hc.HaircutPct/100.0)
	}
	return total, nil
}

type TopupAlert struct {
	UserID   string
	Deficit  float64
	Severity string // WARNING, CRITICAL, LIQUIDATION
}

func (e *HaircutEngine) CheckMarginDeficit(positions []CollateralPosition, requiredMarginINR float64) *TopupAlert {
	effective, _ := e.EffectiveCollateral(positions)
	if effective >= requiredMarginINR {
		return nil
	}
	deficit := requiredMarginINR - effective
	severity := "WARNING"
	ratio := effective / requiredMarginINR
	if ratio < 0.5 {
		severity = "LIQUIDATION"
	} else if ratio < 0.75 {
		severity = "CRITICAL"
	}
	userID := ""
	if len(positions) > 0 {
		userID = positions[0].UserID
	}
	return &TopupAlert{UserID: userID, Deficit: deficit, Severity: severity}
}

var _ = fmt.Sprintf
