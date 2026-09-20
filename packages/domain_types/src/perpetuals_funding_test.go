package src

import (
	"testing"
)

func TestPerpetualsFundingRateNormal(t *testing.T) {
	// Mark slightly above index (longs pay shorts)
	markE8 := uint64(2500500000000) // 25005 USD in e8
	indexE8 := uint64(2500000000000) // 25000 USD in e8

	res, err := Compute8HFundingRate(markE8, indexE8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// premium = (25005 - 25000) / 25000 = 0.0002 = 0.02%
	expectedPremium := 0.0002
	if abs64(res.PremiumIndex-expectedPremium) > 1e-9 {
		t.Fatalf("expected premium %.6f, got %.6f", expectedPremium, res.PremiumIndex)
	}

	if res.Clipped {
		t.Fatalf("normal funding rate should not be clipped")
	}

	// Funding rate should be in [-0.75%, +0.75%]
	if res.FundingRate > 0.0075 || res.FundingRate < -0.0075 {
		t.Fatalf("funding rate %.6f out of clamp bounds", res.FundingRate)
	}
}

func TestPerpetualsFundingRateClampedHigh(t *testing.T) {
	// Extreme markup — mark 5% above index triggers cap clamp
	markE8 := uint64(2625000000000) // 26250 in e8 (5% above 25000)
	indexE8 := uint64(2500000000000)

	res, err := Compute8HFundingRate(markE8, indexE8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Clipped {
		t.Fatalf("expected funding rate to be clamped at +0.75%%")
	}

	capExpected := float64(FundingCapClampBps) / 10000.0
	if abs64(res.FundingRate-capExpected) > 1e-9 {
		t.Fatalf("expected clamped funding rate %.4f, got %.4f", capExpected, res.FundingRate)
	}
}

func TestFundingFeeCalculation(t *testing.T) {
	// 1 BTC long position at 25000 USD, funding rate 0.01%
	positionSizeE8 := int64(100000000) // 1 BTC in e8
	markPriceE8 := uint64(2500000000000) // 25000 USD in e8
	fundingRate := 0.0001 // 0.01%

	fee := ComputeFundingFee(positionSizeE8, markPriceE8, fundingRate)
	// Notional = 1 BTC * 25000 = 25000 USD
	// Fee = 25000 * 0.0001 = 2.5 USD
	// In e8: 250000000 (approx)
	if fee <= 0 {
		t.Fatalf("expected positive funding fee for long position, got %d", fee)
	}
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
