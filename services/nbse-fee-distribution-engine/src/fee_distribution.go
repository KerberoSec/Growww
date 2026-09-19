package main
import ("fmt";"sync")
type FeeDistribution struct { ExchangeShare float64; RegulatoryLevy float64; InvestorProtection float64; SettlementFund float64 }
type FeeDistributor struct { mu sync.Mutex; collections float64; distributions []FeeDistribution }
func NewFeeDistributor() *FeeDistributor { return &FeeDistributor{} }
func (d *FeeDistributor) CollectFee(amount float64) error {
	if amount < 0 { return fmt.Errorf("negative fee") }
	d.mu.Lock(); defer d.mu.Unlock(); d.collections += amount; return nil
}
func (d *FeeDistributor) Distribute() FeeDistribution {
	d.mu.Lock(); defer d.mu.Unlock()
	dist := FeeDistribution{ExchangeShare: d.collections * 0.60, RegulatoryLevy: d.collections * 0.10, InvestorProtection: d.collections * 0.15, SettlementFund: d.collections * 0.15}
	d.distributions = append(d.distributions, dist); d.collections = 0; return dist
}
func (d *FeeDistributor) TotalCollected() float64 { d.mu.Lock(); defer d.mu.Unlock(); return d.collections }
