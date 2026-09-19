package main
import "testing"
func TestMCXPriceUpdate(t *testing.T) {
	a := NewMCXAdapter(); a.UpdatePrice("GOLD", 58500, "10g")
	p, err := a.GetPrice("GOLD"); if err != nil { t.Fatal(err) }
	if p.PriceINR != 58500 { t.Errorf("got %f", p.PriceINR) }
}
func TestNotFound(t *testing.T) {
	a := NewMCXAdapter(); _, err := a.GetPrice("PLATINUM")
	if err == nil { t.Error("expected not found") }
}
func TestGetAll(t *testing.T) {
	a := NewMCXAdapter(); a.UpdatePrice("GOLD", 58500, "10g"); a.UpdatePrice("SILVER", 72000, "kg")
	all := a.GetAllPrices(); if len(all) != 2 { t.Errorf("expected 2, got %d", len(all)) }
}
