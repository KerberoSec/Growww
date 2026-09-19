package main
import "testing"
func TestRFQFlow(t *testing.T) {
	eng := NewRFQEngine(0.5)
	q, err := eng.RequestQuote("BTC/USDT", "BUY", 1.0, 65000); if err != nil { t.Fatal(err) }
	if q.Price <= 65000 { t.Error("buy quote should include spread markup") }
	if err := eng.ExecuteQuote(q.QuoteID); err != nil { t.Fatal(err) }
	if err := eng.ExecuteQuote(q.QuoteID); err == nil { t.Error("double execute should fail") }
}
func TestZeroQty(t *testing.T) {
	eng := NewRFQEngine(0.5); _, err := eng.RequestQuote("BTC/USDT", "BUY", 0, 65000)
	if err == nil { t.Error("expected error for zero qty") }
}
