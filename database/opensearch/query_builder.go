package opensearch

import (
	"encoding/json"
	"math"
	"sort"
	"time"
)

// QueryBuilder constructs OpenSearch query DSL JSON payloads.
type QueryBuilder struct {
	mustClauses   []map[string]interface{}
	filterClauses []map[string]interface{}
	shouldClauses []map[string]interface{}
	size          int
	from          int
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		mustClauses:   make([]map[string]interface{}, 0),
		filterClauses: make([]map[string]interface{}, 0),
		shouldClauses: make([]map[string]interface{}, 0),
		size:          50,
		from:          0,
	}
}

func (b *QueryBuilder) Term(field string, value interface{}) *QueryBuilder {
	b.filterClauses = append(b.filterClauses, map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	})
	return b
}

func (b *QueryBuilder) Match(field string, query string) *QueryBuilder {
	b.mustClauses = append(b.mustClauses, map[string]interface{}{
		"match": map[string]interface{}{
			field: query,
		},
	})
	return b
}

func (b *QueryBuilder) RangeTime(field string, gte, lte time.Time) *QueryBuilder {
	b.filterClauses = append(b.filterClauses, map[string]interface{}{
		"range": map[string]interface{}{
			field: map[string]interface{}{
				"gte": gte.Format(time.RFC3339),
				"lte": lte.Format(time.RFC3339),
			},
		},
	})
	return b
}

func (b *QueryBuilder) SetPagination(from, size int) *QueryBuilder {
	b.from = from
	b.size = size
	return b
}

func (b *QueryBuilder) Build() (string, error) {
	boolMap := make(map[string]interface{})
	if len(b.mustClauses) > 0 {
		boolMap["must"] = b.mustClauses
	}
	if len(b.filterClauses) > 0 {
		boolMap["filter"] = b.filterClauses
	}
	if len(b.shouldClauses) > 0 {
		boolMap["should"] = b.shouldClauses
	}

	payload := map[string]interface{}{
		"from": b.from,
		"size": b.size,
		"query": map[string]interface{}{
			"bool": boolMap,
		},
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ComputeOHLCV aggregates a slice of TradeDocuments into discrete time windows.
func ComputeOHLCV(trades []*TradeDocument, window time.Duration) []*Candlestick {
	if len(trades) == 0 {
		return nil
	}

	// Sort trades chronologically
	sortedTrades := make([]*TradeDocument, len(trades))
	copy(sortedTrades, trades)
	sort.Slice(sortedTrades, func(i, j int) bool {
		return sortedTrades[i].ExecutedAt.Before(sortedTrades[j].ExecutedAt)
	})

	buckets := make(map[int64][]*TradeDocument)
	for _, t := range sortedTrades {
		bucketKey := t.ExecutedAt.Truncate(window).Unix()
		buckets[bucketKey] = append(buckets[bucketKey], t)
	}

	var bucketKeys []int64
	for k := range buckets {
		bucketKeys = append(bucketKeys, k)
	}
	sort.Slice(bucketKeys, func(i, j int) bool {
		return bucketKeys[i] < bucketKeys[j]
	})

	results := make([]*Candlestick, 0, len(bucketKeys))
	for _, k := range bucketKeys {
		group := buckets[k]
		windowStart := time.Unix(k, 0).UTC()
		openPrice := group[0].PricePerUnitINR
		closePrice := group[len(group)-1].PricePerUnitINR
		highPrice := -1.0
		lowPrice := math.MaxFloat64
		var totalVolume float64

		for _, tr := range group {
			if tr.PricePerUnitINR > highPrice {
				highPrice = tr.PricePerUnitINR
			}
			if tr.PricePerUnitINR < lowPrice {
				lowPrice = tr.PricePerUnitINR
			}
			totalVolume += tr.FractionalUnits
		}

		results = append(results, &Candlestick{
			ISIN:        group[0].ISIN,
			WindowStart: windowStart,
			Open:        openPrice,
			High:        highPrice,
			Low:         lowPrice,
			Close:       closePrice,
			Volume:      totalVolume,
			TradeCount:  len(group),
		})
	}

	return results
}
