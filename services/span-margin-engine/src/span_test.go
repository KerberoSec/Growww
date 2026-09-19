package main
import "testing"
func TestLongPositionMargin(t *testing.T) {
	eng := NewSPANEngine()
	positions := []PositionRisk{{Symbol: "BTC", Quantity: 1, CurrentPrice: 65000, IsLong: true}}
	margin := eng.ComputeMargin(positions)
	if margin <= 0 { t.Error("margin should be positive for long position") }
	if margin > 65000*0.20 { t.Error("margin seems too high") }
}
func TestEmptyPositions(t *testing.T) {
	eng := NewSPANEngine()
	margin := eng.ComputeMargin(nil)
	if margin != 0 { t.Errorf("expected 0, got %f", margin) }
}
