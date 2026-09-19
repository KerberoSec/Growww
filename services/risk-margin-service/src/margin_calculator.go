package main

import (
	"errors"
	"math"
)

type MarginMode string

const (
	MarginIsolated MarginMode = "ISOLATED"
	MarginCross    MarginMode = "CROSS"
)

type CollateralAsset struct {
	Symbol     string
	AmountE8   uint64
	PriceUSD   float64
	HaircutPct float64 // e.g. 0% for USDT, 15% for BTC, 30% for tokenized equities
}

type CrossMarginAccount struct {
	UserID            string
	Collaterals       []CollateralAsset
	TotalBorrowedUSD  float64
	MaintenanceMargin float64
}

// CalculateEffectiveEquity computes total discounted equity available across all collaterals
func (a *CrossMarginAccount) CalculateEffectiveEquity() float64 {
	var totalEquityUSD float64
	for _, c := range a.Collaterals {
		rawUSD := (float64(c.AmountE8) / 1e8) * c.PriceUSD
		discountedUSD := rawUSD * (1.0 - (c.HaircutPct / 100.0))
		totalEquityUSD += discountedUSD
	}
	return totalEquityUSD
}

// CalculateMarginLevel evaluates Margin Level = Total Asset Value / (Total Borrowed + Maintenance)
func (a *CrossMarginAccount) CalculateMarginLevel() (float64, error) {
	if a.TotalBorrowedUSD <= 0 {
		return math.MaxFloat64, nil
	}
	equity := a.CalculateEffectiveEquity()
	if equity <= 0 {
		return 0.0, errors.New("zero or negative collateral equity")
	}
	marginLevel := (equity / a.TotalBorrowedUSD) * 100.0
	return marginLevel, nil
}
