package main
import ("math";"testing")
func TestFeeDistribution(t *testing.T) {
	d := NewFeeDistributor(); d.CollectFee(100000)
	dist := d.Distribute()
	if math.Abs(dist.ExchangeShare - 60000) > 0.01 { t.Errorf("exchange share: %f", dist.ExchangeShare) }
	if math.Abs(dist.RegulatoryLevy - 10000) > 0.01 { t.Errorf("regulatory: %f", dist.RegulatoryLevy) }
	if d.TotalCollected() != 0 { t.Error("should be zero after distribute") }
}
func TestNegativeFee(t *testing.T) {
	d := NewFeeDistributor(); if err := d.CollectFee(-100); err == nil { t.Error("expected error") }
}
