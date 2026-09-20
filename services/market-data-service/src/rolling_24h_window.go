package main

import (
	"errors"
	"math/big"
	"sync"
	"time"
)

var (
	ErrFutureTimestamp  = errors.New("trade timestamp is in the future (> 5s skew)")
	ErrStaleTimestamp   = errors.New("trade timestamp is older than rolling window")
	ErrInvalidTradeData = errors.New("invalid trade price or quantity")
)

type TradeTick struct {
	TradeID    string    `json:"trade_id"`
	PriceE8    uint64    `json:"price_e8"`
	QuantityE8 uint64    `json:"quantity_e8"`
	Timestamp  time.Time `json:"timestamp"`
}

type TickerStats24h struct {
	Symbol          string    `json:"symbol"`
	LastPriceE8     uint64    `json:"last_price_e8"`
	OpenPrice24hE8  uint64    `json:"open_price_24h_e8"`
	High24hE8       uint64    `json:"high_24h_e8"`
	Low24hE8        uint64    `json:"low_24h_e8"`
	Volume24hE8     uint64    `json:"volume_24h_e8"`
	QuoteVolume24h  float64   `json:"quote_volume_24h"`
	Vwap24hE8       uint64    `json:"vwap_24h_e8"`
	PriceChangePct  float64   `json:"price_change_pct_24h"`
	TradeCount24h   uint64    `json:"trade_count_24h"`
	WindowStart     time.Time `json:"window_start"`
	WindowEnd       time.Time `json:"window_end"`
}

type SymbolWindow struct {
	symbol   string
	ticks    []TradeTick
	window   time.Duration
}

type Rolling24hWindowCalculator struct {
	mu             sync.RWMutex
	windows        map[string]*SymbolWindow
	windowDuration time.Duration
}

func NewRolling24hWindowCalculator(windowDuration time.Duration) *Rolling24hWindowCalculator {
	if windowDuration <= 0 {
		windowDuration = 24 * time.Hour
	}
	return &Rolling24hWindowCalculator{
		windows:        make(map[string]*SymbolWindow),
		windowDuration: windowDuration,
	}
}

// IngestTrade processes an incoming trade tick and updates rolling window stats
func (c *Rolling24hWindowCalculator) IngestTrade(symbol string, tick TradeTick, now time.Time) (*TickerStats24h, error) {
	if tick.PriceE8 == 0 || tick.QuantityE8 == 0 {
		return nil, ErrInvalidTradeData
	}

	// Reject anomalous future timestamps (> 5s drift)
	if tick.Timestamp.After(now.Add(5 * time.Second)) {
		return nil, ErrFutureTimestamp
	}

	cutoff := now.Add(-c.windowDuration)
	if tick.Timestamp.Before(cutoff) {
		return nil, ErrStaleTimestamp
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	w, exists := c.windows[symbol]
	if !exists {
		w = &SymbolWindow{
			symbol: symbol,
			ticks:  make([]TradeTick, 0, 1000),
			window: c.windowDuration,
		}
		c.windows[symbol] = w
	}

	// Append new tick
	w.ticks = append(w.ticks, tick)

	// Evict stale ticks outside the window
	evictIdx := 0
	for evictIdx < len(w.ticks) && w.ticks[evictIdx].Timestamp.Before(cutoff) {
		evictIdx++
	}
	if evictIdx > 0 {
		w.ticks = w.ticks[evictIdx:]
	}

	return c.computeStatsLocked(w, now)
}

// GetStats returns current rolling window stats for a symbol
func (c *Rolling24hWindowCalculator) GetStats(symbol string, now time.Time) (*TickerStats24h, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	w, exists := c.windows[symbol]
	if !exists || len(w.ticks) == 0 {
		return nil, false
	}

	// Purge stale ticks on read
	cutoff := now.Add(-c.windowDuration)
	evictIdx := 0
	for evictIdx < len(w.ticks) && w.ticks[evictIdx].Timestamp.Before(cutoff) {
		evictIdx++
	}
	if evictIdx > 0 {
		w.ticks = w.ticks[evictIdx:]
	}

	if len(w.ticks) == 0 {
		return nil, false
	}

	stats, _ := c.computeStatsLocked(w, now)
	return stats, true
}

func (c *Rolling24hWindowCalculator) computeStatsLocked(w *SymbolWindow, now time.Time) (*TickerStats24h, error) {
	if len(w.ticks) == 0 {
		return nil, errors.New("no ticks in window")
	}

	firstTick := w.ticks[0]
	lastTick := w.ticks[len(w.ticks)-1]

	high := firstTick.PriceE8
	low := firstTick.PriceE8
	totalVolE8 := uint64(0)

	// BigInt accumulators for safe financial multiplication without overflow
	sumNotionalE16 := big.NewInt(0)
	sumQuantityE8 := big.NewInt(0)

	for _, t := range w.ticks {
		if t.PriceE8 > high {
			high = t.PriceE8
		}
		if t.PriceE8 < low {
			low = t.PriceE8
		}
		totalVolE8 += t.QuantityE8

		pBig := new(big.Int).SetUint64(t.PriceE8)
		qBig := new(big.Int).SetUint64(t.QuantityE8)
		notional := new(big.Int).Mul(pBig, qBig)

		sumNotionalE16.Add(sumNotionalE16, notional)
		sumQuantityE8.Add(sumQuantityE8, qBig)
	}

	// VWAP = sum(Price * Qty) / sum(Qty)
	vwapE8 := uint64(0)
	if sumQuantityE8.Sign() > 0 {
		vwapBig := new(big.Int).Div(sumNotionalE16, sumQuantityE8)
		vwapE8 = vwapBig.Uint64()
	}

	// Quote volume = sum(PriceE8 * QtyE8) / 1e16
	e16Divisor := new(big.Float).SetUint64(10000000000000000)
	notionalFloat := new(big.Float).SetInt(sumNotionalE16)
	quoteVolFloat := new(big.Float).Quo(notionalFloat, e16Divisor)
	quoteVol, _ := quoteVolFloat.Float64()

	// Percentage change against open price
	pctChange := 0.0
	if firstTick.PriceE8 > 0 {
		pctChange = (float64(lastTick.PriceE8) - float64(firstTick.PriceE8)) / float64(firstTick.PriceE8) * 100.0
	}

	return &TickerStats24h{
		Symbol:         w.symbol,
		LastPriceE8:    lastTick.PriceE8,
		OpenPrice24hE8: firstTick.PriceE8,
		High24hE8:      high,
		Low24hE8:       low,
		Volume24hE8:    totalVolE8,
		QuoteVolume24h: quoteVol,
		Vwap24hE8:      vwapE8,
		PriceChangePct: pctChange,
		TradeCount24h:  uint64(len(w.ticks)),
		WindowStart:    now.Add(-c.windowDuration),
		WindowEnd:      now,
	}, nil
}
