package src

import (
	"errors"
	"math"
)

// =====================================================
// Prompt 183: Perpetuals 8H Funding Rate & Clamping
// =====================================================

const (
	FundingInterestRate   = 0.0003     // 0.03% per funding interval (3 intervals/day = 0.09%/day)
	FundingCapClampBps    = 75         // Maximum 0.75% per 8H interval
	FundingFloorClampBps  = -75        // Minimum -0.75% per 8H interval
)

var ErrInvalidMarkPrice = errors.New("funding_rate: mark price must be positive")

// FundingRateResult stores the computed 8H funding and premium components.
type FundingRateResult struct {
	PremiumIndex   float64 // (Mark - Index) / Index
	FundingRate    float64 // Clamped composite funding rate
	FundingFeeE8   int64   // Funding fee in e8 for a given position size
	Clipped        bool    // Whether funding was clamped by cap/floor
}

// Compute8HFundingRate computes perpetual futures 8-hour funding rate per Binance/Growww model.
// FundingRate = PremiumIndex + clamp(InterestRate - PremiumIndex, -0.05%, 0.05%)
func Compute8HFundingRate(markPriceE8, indexPriceE8 uint64) (FundingRateResult, error) {
	if markPriceE8 == 0 || indexPriceE8 == 0 {
		return FundingRateResult{}, ErrInvalidMarkPrice
	}

	mark := float64(markPriceE8)
	index := float64(indexPriceE8)

	premium := (mark - index) / index

	// clamp(FundingInterestRate - premium, -0.05%, +0.05%)
	adjustment := FundingInterestRate - premium
	clampMin := -0.0005
	clampMax := 0.0005
	clamped := math.Max(clampMin, math.Min(clampMax, adjustment))

	fundingRate := premium + clamped

	// Apply global cap/floor clamp: max ±0.75%
	capClamp := float64(FundingCapClampBps) / 10000.0
	floorClamp := float64(FundingFloorClampBps) / 10000.0
	clipped := false
	if fundingRate > capClamp {
		fundingRate = capClamp
		clipped = true
	} else if fundingRate < floorClamp {
		fundingRate = floorClamp
		clipped = true
	}

	return FundingRateResult{
		PremiumIndex: premium,
		FundingRate:  fundingRate,
		Clipped:      clipped,
	}, nil
}

// ComputeFundingFee computes the funding fee payment for a position.
// Positive = long pays, Negative = short pays
func ComputeFundingFee(positionSizeE8 int64, markPriceE8 uint64, fundingRate float64) int64 {
	notional := float64(positionSizeE8) * float64(markPriceE8) / 1e8
	feeRaw := notional * fundingRate
	return int64(feeRaw)
}
