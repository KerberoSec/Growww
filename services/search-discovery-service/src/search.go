package main
import ("sort";"strings";"sync")
type SearchableAsset struct { ID string; Symbol string; Name string; AssetType string; Tags []string }
type SearchService struct { mu sync.RWMutex; assets []SearchableAsset }
func NewSearchService() *SearchService { return &SearchService{} }
func (s *SearchService) IndexAsset(asset SearchableAsset) { s.mu.Lock(); defer s.mu.Unlock(); s.assets = append(s.assets, asset) }
func (s *SearchService) Search(query string, limit int) []SearchableAsset {
	s.mu.RLock(); defer s.mu.RUnlock(); q := strings.ToLower(query)
	type scored struct { asset SearchableAsset; score int }
	var results []scored
	for _, a := range s.assets {
		score := 0
		if strings.Contains(strings.ToLower(a.Symbol), q) { score += 10 }
		if strings.Contains(strings.ToLower(a.Name), q) { score += 5 }
		for _, tag := range a.Tags { if strings.Contains(strings.ToLower(tag), q) { score += 2 } }
		if score > 0 { results = append(results, scored{a, score}) }
	}
	sort.Slice(results, func(i, j int) bool { return results[i].score > results[j].score })
	var out []SearchableAsset; for i, r := range results { if i >= limit { break }; out = append(out, r.asset) }
	return out
}
func (s *SearchService) Autocomplete(prefix string, limit int) []string {
	s.mu.RLock(); defer s.mu.RUnlock(); p := strings.ToUpper(prefix)
	var matches []string
	for _, a := range s.assets { if strings.HasPrefix(a.Symbol, p) { matches = append(matches, a.Symbol) } }
	if len(matches) > limit { matches = matches[:limit] }; return matches
}
