package main

import (
	"sync"
	"time"
)

type OHLCVCandle struct {
	Symbol     string    `json:"symbol"`
	Interval   string    `json:"interval"` // 1m, 5m, 1h, 1d
	OpenE8     uint64    `json:"open_e8"`
	HighE8     uint64    `json:"high_e8"`
	LowE8      uint64    `json:"low_e8"`
	CloseE8    uint64    `json:"close_e8"`
	VolumeE8   uint64    `json:"volume_e8"`
	Trades     uint32    `json:"trades"`
	BucketTime time.Time `json:"bucket_time"`
}

type OHLCVAggregator struct {
	mu            sync.Mutex
	activeCandles map[string]*OHLCVCandle // "symbol:interval" -> candle
}

func NewOHLCVAggregator() *OHLCVAggregator {
	return &OHLCVAggregator{
		activeCandles: make(map[string]*OHLCVCandle),
	}
}

// IngestTrade aggregates an execution tick into the rolling 1m / 5m OHLCV bucket
func (a *OHLCVAggregator) IngestTrade(symbol string, priceE8, quantityE8 uint64, tradeTime time.Time) *OHLCVCandle {
	a.mu.Lock()
	defer a.mu.Unlock()

	bucket := tradeTime.Truncate(time.Minute)
	key := symbol + ":1m"

	c, exists := a.activeCandles[key]
	if !exists || !c.BucketTime.Equal(bucket) {
		// Flush old bucket to TimescaleDB / ClickHouse and start new
		c = &OHLCVCandle{
			Symbol:     symbol,
			Interval:   "1m",
			OpenE8:     priceE8,
			HighE8:     priceE8,
			LowE8:      priceE8,
			CloseE8:    priceE8,
			VolumeE8:   quantityE8,
			Trades:     1,
			BucketTime: bucket,
		}
		a.activeCandles[key] = c
	} else {
		if priceE8 > c.HighE8 {
			c.HighE8 = priceE8
		}
		if priceE8 < c.LowE8 {
			c.LowE8 = priceE8
		}
		c.CloseE8 = priceE8
		c.VolumeE8 += quantityE8
		c.Trades++
	}

	return c
}
