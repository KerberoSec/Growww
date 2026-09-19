package main
import "testing"
func TestBestQuote(t *testing.T) {
	a := NewNSEBSEAdapter()
	a.IngestQuote(&EquityQuote{Symbol: "RELIANCE", Exchange: "NSE", LastPrice: 2500})
	a.IngestQuote(&EquityQuote{Symbol: "RELIANCE", Exchange: "BSE", LastPrice: 2498})
	best, err := a.GetBestQuote("RELIANCE"); if err != nil { t.Fatal(err) }
	if best.Exchange != "BSE" { t.Errorf("expected BSE (cheaper), got %s", best.Exchange) }
}
func TestNoQuote(t *testing.T) {
	a := NewNSEBSEAdapter(); _, err := a.GetBestQuote("MISSING")
	if err == nil { t.Error("expected error") }
}
