package main
import ("fmt";"sync";"time")
type CommodityPrice struct { Symbol string; PriceINR float64; Unit string; Timestamp time.Time }
type MCXAdapter struct { mu sync.RWMutex; prices map[string]*CommodityPrice }
func NewMCXAdapter() *MCXAdapter { return &MCXAdapter{prices: make(map[string]*CommodityPrice)} }
func (a *MCXAdapter) UpdatePrice(symbol string, price float64, unit string) {
	a.mu.Lock(); defer a.mu.Unlock()
	a.prices[symbol] = &CommodityPrice{Symbol: symbol, PriceINR: price, Unit: unit, Timestamp: time.Now().UTC()}
}
func (a *MCXAdapter) GetPrice(symbol string) (*CommodityPrice, error) {
	a.mu.RLock(); defer a.mu.RUnlock()
	p, ok := a.prices[symbol]; if !ok { return nil, fmt.Errorf("symbol %s not found", symbol) }; return p, nil
}
func (a *MCXAdapter) GetAllPrices() []*CommodityPrice {
	a.mu.RLock(); defer a.mu.RUnlock()
	var out []*CommodityPrice; for _, p := range a.prices { out = append(out, p) }; return out
}
