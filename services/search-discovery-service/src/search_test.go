package main
import "testing"
func TestSearch(t *testing.T) {
	svc := NewSearchService()
	svc.IndexAsset(SearchableAsset{ID: "1", Symbol: "BTC", Name: "Bitcoin", Tags: []string{"crypto", "layer1"}})
	svc.IndexAsset(SearchableAsset{ID: "2", Symbol: "ETH", Name: "Ethereum", Tags: []string{"crypto", "smart-contracts"}})
	results := svc.Search("bitcoin", 10)
	if len(results) != 1 || results[0].Symbol != "BTC" { t.Error("expected BTC") }
}
func TestAutocomplete(t *testing.T) {
	svc := NewSearchService()
	svc.IndexAsset(SearchableAsset{Symbol: "BTC"}); svc.IndexAsset(SearchableAsset{Symbol: "BCH"}); svc.IndexAsset(SearchableAsset{Symbol: "ETH"})
	matches := svc.Autocomplete("B", 10)
	if len(matches) != 2 { t.Errorf("expected 2 matches, got %d", len(matches)) }
}
