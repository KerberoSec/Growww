package main

import (
	"errors"
	"testing"
)

func TestCrossCollateral_SEBIHaircutBrackets(t *testing.T) {
	// Group 1: 12% VaR + 3.5% ELM = 15.5%, floored at 20.0%
	h1 := ComputeHaircut(Group1Equity, 12.0)
	if h1 != 20.0 {
		t.Fatalf("expected Group 1 haircut to be 20.0%%, got %.2f%%", h1)
	}

	// Group 1: 22% VaR + 3.5% ELM = 25.5%
	h1High := ComputeHaircut(Group1Equity, 22.0)
	if h1High != 25.5 {
		t.Fatalf("expected Group 1 haircut to be 25.5%%, got %.2f%%", h1High)
	}

	// Group 2: Midcap floored at 30.0%
	h2 := ComputeHaircut(Group2Equity, 15.0)
	if h2 != 30.0 {
		t.Fatalf("expected Group 2 haircut to be 30.0%%, got %.2f%%", h2)
	}

	// Group 3: Smallcap floored at 50.0%
	h3 := ComputeHaircut(Group3Equity, 20.0)
	if h3 != 50.0 {
		t.Fatalf("expected Group 3 haircut to be 50.0%%, got %.2f%%", h3)
	}
}

func TestCrossCollateral_PledgeLifecycleAndMarginUtilization(t *testing.T) {
	engine := NewCrossCollateralEngine(84.0) // 84.0 INR per USDT

	// Pledge 100 shares of Reliance (Group 1) at ₹2,500
	// Gross value = 250,000 INR.
	// Group 1 Haircut = 20% -> Collateral Val = 200,000 INR
	// Margin in USDT = 200,000 / 84 = 2380.95 USDT
	rec, err := engine.PledgeDematStock("PLG-001", "USER-42", "INE002A01018", Group1Equity, 100, 2500.0, 12.0)
	if err != nil {
		t.Fatalf("unexpected error pledging stock: %v", err)
	}
	if rec.CollateralValINR != 200000.0 {
		t.Fatalf("expected collateral value 200000.0, got %.2f", rec.CollateralValINR)
	}

	// Health factor initially infinite (used margin is 0)
	health, call, liq, err := engine.GetCollateralHealth("PLG-001")
	if err != nil || health != 999.0 || call || liq {
		t.Fatalf("unexpected initial health: health=%.2f, call=%v, liq=%v, err=%v", health, call, liq, err)
	}

	// Utilize 1500 USDT margin for spot BTC purchase
	err = engine.UtilizeMargin("PLG-001", 1500.0)
	if err != nil {
		t.Fatalf("unexpected error utilizing margin: %v", err)
	}

	health, call, liq, err = engine.GetCollateralHealth("PLG-001")
	if err != nil || health < 1.5 || call || liq {
		t.Fatalf("unexpected health after utilization: health=%.2f, call=%v, liq=%v", health, call, liq)
	}

	// Attempting to utilize more than available margin should fail
	err = engine.UtilizeMargin("PLG-001", 1000.0) // 1500 + 1000 = 2500 > 2380.95
	if err == nil {
		t.Fatal("expected error exceeding margin capacity")
	}

	// Price of Reliance drops to ₹1,800 (-28%)
	// Gross value = 180,000 INR. Haircut 20% -> Collateral = 144,000 INR = 1714.28 USDT
	// Used margin is 1500 USDT -> Health = 1714.28 / 1500 = 1.1428 -> Margin Call triggered!
	_, err = engine.UpdateMarketPrice("PLG-001", 1800.0)
	if err != nil {
		t.Fatalf("unexpected error updating market price: %v", err)
	}

	health, call, liq, err = engine.GetCollateralHealth("PLG-001")
	if err != nil || !call || liq {
		t.Fatalf("expected margin call: health=%.2f, call=%v, liq=%v", health, call, liq)
	}

	// Price drops further to ₹1,400 -> Collateral = 112,000 INR = 1333.33 USDT
	// Health = 1333.33 / 1500 = 0.888 -> Liquidation triggered!
	_, err = engine.UpdateMarketPrice("PLG-001", 1400.0)
	if err != nil {
		t.Fatalf("unexpected error updating market price: %v", err)
	}

	health, call, liq, err = engine.GetCollateralHealth("PLG-001")
	if err != nil || !call || !liq {
		t.Fatalf("expected liquidation trigger: health=%.2f, call=%v, liq=%v", health, call, liq)
	}

	// Attempt to unpledge while margin is in use must be rejected
	_, err = engine.UnpledgeDematStock("PLG-001")
	if !errors.Is(err, ErrInsufficientMargin) {
		t.Fatalf("expected ErrInsufficientMargin, got: %v", err)
	}

	// Close spot positions and release margin
	err = engine.ReleaseMargin("PLG-001", 1500.0)
	if err != nil {
		t.Fatalf("unexpected error releasing margin: %v", err)
	}

	// Unpledge now succeeds
	unpledged, err := engine.UnpledgeDematStock("PLG-001")
	if err != nil || unpledged.PledgeID != "PLG-001" {
		t.Fatalf("unexpected error unpledging stock: %v", err)
	}

	// Pledge record removed
	_, err = engine.UnpledgeDematStock("PLG-001")
	if !errors.Is(err, ErrPledgeNotFound) {
		t.Fatalf("expected ErrPledgeNotFound, got: %v", err)
	}
}

func TestCrossCollateral_ValidationErrors(t *testing.T) {
	engine := NewCrossCollateralEngine(84.0)

	// Zero shares
	_, err := engine.PledgeDematStock("P1", "U1", "ISIN1", Group1Equity, 0, 100.0, 10.0)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity for 0 shares, got %v", err)
	}

	// Zero price
	_, err = engine.PledgeDematStock("P1", "U1", "ISIN1", Group1Equity, 10, 0.0, 10.0)
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("expected ErrInvalidQuantity for 0 price, got %v", err)
	}

	// Duplicate pledge ID
	_, err = engine.PledgeDematStock("P1", "U1", "ISIN1", Group1Equity, 10, 100.0, 10.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = engine.PledgeDematStock("P1", "U1", "ISIN1", Group1Equity, 10, 100.0, 10.0)
	if !errors.Is(err, ErrDuplicatePledge) {
		t.Fatalf("expected ErrDuplicatePledge, got %v", err)
	}
}
