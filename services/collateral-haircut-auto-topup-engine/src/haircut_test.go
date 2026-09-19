package main

import "testing"

func TestEffectiveCollateral(t *testing.T) {
	eng := NewHaircutEngine()
	positions := []CollateralPosition{
		{UserID: "u1", Symbol: "INR", Quantity: 100000, PriceINR: 1.0},
		{UserID: "u1", Symbol: "BTC", Quantity: 0.5, PriceINR: 5000000},
	}
	eff, _ := eng.EffectiveCollateral(positions)
	inrPart := 100000.0
	btcPart := 0.5 * 5000000 * 0.75
	expected := inrPart + btcPart
	if eff != expected {
		t.Errorf("got %f, want %f", eff, expected)
	}
}

func TestMarginDeficitWarning(t *testing.T) {
	eng := NewHaircutEngine()
	positions := []CollateralPosition{
		{UserID: "u1", Symbol: "INR", Quantity: 80000, PriceINR: 1.0},
	}
	alert := eng.CheckMarginDeficit(positions, 100000)
	if alert == nil {
		t.Fatal("expected deficit alert")
	}
	if alert.Severity != "WARNING" {
		t.Errorf("expected WARNING, got %s", alert.Severity)
	}
}

func TestNoDeficit(t *testing.T) {
	eng := NewHaircutEngine()
	positions := []CollateralPosition{
		{UserID: "u1", Symbol: "INR", Quantity: 200000, PriceINR: 1.0},
	}
	alert := eng.CheckMarginDeficit(positions, 100000)
	if alert != nil {
		t.Error("expected no alert when collateral exceeds margin")
	}
}
