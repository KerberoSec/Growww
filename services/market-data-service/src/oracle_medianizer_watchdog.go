package main

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

type OraclePriceTick struct {
	Source    string
	Price     float64
	Timestamp time.Time
}

type MedianizerWatchdog struct {
	mu           sync.Mutex
	maxDeviation float64 // e.g. 0.02 for 2% max deviation from median
}

func NewMedianizerWatchdog(maxDeviation float64) *MedianizerWatchdog {
	return &MedianizerWatchdog{maxDeviation: maxDeviation}
}

// ComputeMedianAndFilterClamped filters out rogue oracle feeds and returns median price.
func (w *MedianizerWatchdog) ComputeMedianAndFilterClamped(ticks []OraclePriceTick) (float64, []OraclePriceTick, error) {
	if len(ticks) < 3 {
		return 0, nil, fmt.Errorf("at least 3 oracle sources required for median calculation")
	}

	prices := make([]float64, len(ticks))
	for i, t := range ticks {
		prices[i] = t.Price
	}
	sort.Float64s(prices)

	var median float64
	n := len(prices)
	if n%2 == 0 {
		median = (prices[n/2-1] + prices[n/2]) / 2.0
	} else {
		median = prices[n/2]
	}

	var validTicks []OraclePriceTick
	for _, t := range ticks {
		deviation := math.Abs(t.Price-median) / median
		if deviation <= w.maxDeviation {
			validTicks = append(validTicks, t)
		}
	}

	return median, validTicks, nil
}
