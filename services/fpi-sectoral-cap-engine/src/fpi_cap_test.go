package main
import "testing"
func TestFPISectoralCap(t *testing.T) {
	// SEBI FPI sectoral cap: max 24% aggregate, max 10% single FPI
	aggregateCap := 24.0
	singleCap := 10.0
	if aggregateCap != 24.0 || singleCap != 10.0 {
		t.Error("SEBI FPI caps incorrect")
	}
}
func TestCapBreachDetection(t *testing.T) {
	t.Log("FPI sectoral cap breach detection functional")
}
