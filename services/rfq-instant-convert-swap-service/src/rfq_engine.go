package main
import ("fmt";"sync";"time")
type Quote struct { QuoteID string; Symbol string; Side string; Quantity float64; Price float64; ExpiresAt time.Time; Status string }
type RFQEngine struct { mu sync.Mutex; quotes map[string]*Quote; spread float64 }
func NewRFQEngine(spreadPct float64) *RFQEngine { return &RFQEngine{quotes: make(map[string]*Quote), spread: spreadPct} }
func (e *RFQEngine) RequestQuote(symbol, side string, qty float64, marketPrice float64) (*Quote, error) {
	if qty <= 0 { return nil, fmt.Errorf("quantity must be positive") }
	e.mu.Lock(); defer e.mu.Unlock()
	price := marketPrice; if side == "BUY" { price *= (1 + e.spread/100) } else { price *= (1 - e.spread/100) }
	q := &Quote{QuoteID: fmt.Sprintf("Q-%d", time.Now().UnixNano()), Symbol: symbol, Side: side, Quantity: qty, Price: price, ExpiresAt: time.Now().Add(10*time.Second), Status: "LIVE"}
	e.quotes[q.QuoteID] = q; return q, nil
}
func (e *RFQEngine) ExecuteQuote(quoteID string) error {
	e.mu.Lock(); defer e.mu.Unlock()
	q, ok := e.quotes[quoteID]; if !ok { return fmt.Errorf("quote not found") }
	if q.Status != "LIVE" { return fmt.Errorf("quote not live: %s", q.Status) }
	if time.Now().After(q.ExpiresAt) { q.Status = "EXPIRED"; return fmt.Errorf("quote expired") }
	q.Status = "EXECUTED"; return nil
}
