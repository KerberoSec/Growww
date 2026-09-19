// Package main enhances the NBSE market data service with 24h rolling ticker,
// trade history storage, and order book depth snapshots.
package main

import (
	"fmt"
	"sync"
	"time"
)

// RollingTicker24h represents a 24-hour rolling ticker for a symbol.
type RollingTicker24h struct {
	Symbol        string  `json:"symbol"`
	LastPriceE8   uint64  `json:"last_price_e8"`
	HighE8        uint64  `json:"high_24h_e8"`
	LowE8         uint64  `json:"low_24h_e8"`
	OpenE8        uint64  `json:"open_24h_e8"`
	VolumeE8      uint64  `json:"volume_24h_e8"`
	QuoteVolE8    uint64  `json:"quote_volume_24h_e8"`
	TradeCount    uint64  `json:"trade_count_24h"`
	ChangePct     float64 `json:"change_pct_24h"`
	UpdatedAtMs   int64   `json:"updated_at_ms"`
}

// TradeRecord represents a single executed trade.
type TradeRecord struct {
	TradeID     string    `json:"trade_id"`
	Symbol      string    `json:"symbol"`
	PriceE8     uint64    `json:"price_e8"`
	QuantityE8  uint64    `json:"quantity_e8"`
	BuyerMaker  bool      `json:"buyer_is_maker"`
	ExecutedAt  time.Time `json:"executed_at"`
}

// DepthLevel represents a single price level in the order book.
type DepthLevel struct {
	PriceE8    uint64 `json:"price_e8"`
	QuantityE8 uint64 `json:"quantity_e8"`
	OrderCount int    `json:"order_count"`
}

// DepthSnapshot represents the full order book snapshot.
type DepthSnapshot struct {
	Symbol     string       `json:"symbol"`
	Bids       []DepthLevel `json:"bids"`
	Asks       []DepthLevel `json:"asks"`
	SequenceID uint64       `json:"sequence_id"`
	Timestamp  time.Time    `json:"timestamp"`
}

// TickerService manages 24h rolling tickers.
type TickerService struct {
	mu      sync.RWMutex
	tickers map[string]*RollingTicker24h // symbol -> ticker
}

// NewTickerService creates a new TickerService.
func NewTickerService() *TickerService {
	return &TickerService{
		tickers: make(map[string]*RollingTicker24h),
	}
}

// UpdateTicker updates the 24h rolling ticker with a new trade.
func (ts *TickerService) UpdateTicker(symbol string, priceE8, quantityE8 uint64) *RollingTicker24h {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ticker, exists := ts.tickers[symbol]
	if !exists {
		ticker = &RollingTicker24h{
			Symbol:      symbol,
			OpenE8:      priceE8,
			HighE8:      priceE8,
			LowE8:       priceE8,
			LastPriceE8: priceE8,
			VolumeE8:    quantityE8,
			QuoteVolE8:  (priceE8 * quantityE8) / 1e8,
			TradeCount:  1,
			ChangePct:   0.0,
			UpdatedAtMs: time.Now().UnixMilli(),
		}
		ts.tickers[symbol] = ticker
		return ticker
	}

	ticker.LastPriceE8 = priceE8
	if priceE8 > ticker.HighE8 {
		ticker.HighE8 = priceE8
	}
	if priceE8 < ticker.LowE8 {
		ticker.LowE8 = priceE8
	}
	ticker.VolumeE8 += quantityE8
	ticker.QuoteVolE8 += (priceE8 * quantityE8) / 1e8
	ticker.TradeCount++
	if ticker.OpenE8 > 0 {
		ticker.ChangePct = (float64(priceE8) - float64(ticker.OpenE8)) / float64(ticker.OpenE8) * 100.0
	}
	ticker.UpdatedAtMs = time.Now().UnixMilli()

	return ticker
}

// GetTicker returns the current 24h ticker for a symbol.
func (ts *TickerService) GetTicker(symbol string) (*RollingTicker24h, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	t, ok := ts.tickers[symbol]
	return t, ok
}

// TradeHistoryStore stores recent trade history for each symbol.
type TradeHistoryStore struct {
	mu      sync.RWMutex
	trades  map[string][]TradeRecord // symbol -> trades (ring buffer)
	maxSize int
}

// NewTradeHistoryStore creates a new TradeHistoryStore with max capacity per symbol.
func NewTradeHistoryStore(maxSize int) *TradeHistoryStore {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &TradeHistoryStore{
		trades:  make(map[string][]TradeRecord),
		maxSize: maxSize,
	}
}

// RecordTrade appends a trade to the history, evicting oldest if at capacity.
func (s *TradeHistoryStore) RecordTrade(trade TradeRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	trades := s.trades[trade.Symbol]
	if len(trades) >= s.maxSize {
		trades = trades[1:] // Evict oldest
	}
	s.trades[trade.Symbol] = append(trades, trade)
}

// GetRecentTrades returns the N most recent trades for a symbol.
func (s *TradeHistoryStore) GetRecentTrades(symbol string, limit int) []TradeRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	trades := s.trades[symbol]
	if limit <= 0 || limit > len(trades) {
		limit = len(trades)
	}
	start := len(trades) - limit
	result := make([]TradeRecord, limit)
	copy(result, trades[start:])
	return result
}

// DepthSnapshotManager manages order book depth snapshots.
type DepthSnapshotManager struct {
	mu        sync.RWMutex
	snapshots map[string]*DepthSnapshot // symbol -> latest snapshot
	seqID     uint64
}

// NewDepthSnapshotManager creates a new DepthSnapshotManager.
func NewDepthSnapshotManager() *DepthSnapshotManager {
	return &DepthSnapshotManager{
		snapshots: make(map[string]*DepthSnapshot),
	}
}

// UpdateSnapshot replaces the depth snapshot for a symbol.
func (m *DepthSnapshotManager) UpdateSnapshot(symbol string, bids, asks []DepthLevel) *DepthSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.seqID++
	snap := &DepthSnapshot{
		Symbol:     symbol,
		Bids:       bids,
		Asks:       asks,
		SequenceID: m.seqID,
		Timestamp:  time.Now().UTC(),
	}
	m.snapshots[symbol] = snap
	return snap
}

// GetSnapshot returns the latest depth snapshot for a symbol.
func (m *DepthSnapshotManager) GetSnapshot(symbol string) (*DepthSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snap, exists := m.snapshots[symbol]
	if !exists {
		return nil, fmt.Errorf("no depth snapshot for symbol %s", symbol)
	}
	return snap, nil
}
