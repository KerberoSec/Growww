package main
import ("fmt";"sync";"time")
type EquityQuote struct { Symbol string; Exchange string; LastPrice float64; Open float64; High float64; Low float64; Volume uint64; Timestamp time.Time }
type NSEBSEAdapter struct { mu sync.RWMutex; quotes map[string]*EquityQuote }
func NewNSEBSEAdapter() *NSEBSEAdapter { return &NSEBSEAdapter{quotes: make(map[string]*EquityQuote)} }
func (a *NSEBSEAdapter) IngestQuote(q *EquityQuote) { a.mu.Lock(); defer a.mu.Unlock(); q.Timestamp = time.Now().UTC(); a.quotes[q.Symbol+":"+q.Exchange] = q }
func (a *NSEBSEAdapter) GetBestQuote(symbol string) (*EquityQuote, error) {
	a.mu.RLock(); defer a.mu.RUnlock()
	nse := a.quotes[symbol+":NSE"]; bse := a.quotes[symbol+":BSE"]
	if nse == nil && bse == nil { return nil, fmt.Errorf("no quote for %s", symbol) }
	if nse == nil { return bse, nil }; if bse == nil { return nse, nil }
	if nse.LastPrice <= bse.LastPrice { return nse, nil }; return bse, nil
}
