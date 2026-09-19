package main

import (
	"fmt"
	"sync"
	"time"
)

// MarketTick represents a live price update from real exchange feeds
type MarketTick struct {
	Symbol      string  `json:"symbol"`
	PriceUSD    float64 `json:"price_usd"`
	BidUSD      float64 `json:"bid_usd"`
	AskUSD      float64 `json:"ask_usd"`
	Volume24h   float64 `json:"volume_24h"`
	TimestampMs int64   `json:"timestamp_ms"`
}

// PriceMirrorPipeline mirrors production market ticks into the paper trading sandbox
type PriceMirrorPipeline struct {
	mu           sync.RWMutex
	latestPrices map[string]MarketTick
	subscribers  map[string]chan MarketTick
}

func NewPriceMirrorPipeline() *PriceMirrorPipeline {
	return &PriceMirrorPipeline{
		latestPrices: make(map[string]MarketTick),
		subscribers:  make(map[string]chan MarketTick),
	}
}

// IngestProductionTick mirrors a live market tick and fans out to virtual orderbooks
func (p *PriceMirrorPipeline) IngestProductionTick(tick MarketTick) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.latestPrices[tick.Symbol] = tick

	// Fan out to all connected demo matching engines
	for subID, ch := range p.subscribers {
		select {
		case ch <- tick:
		default:
			// Non-blocking drop if consumer buffer is full
			_ = subID
		}
	}
}

// Subscribe returns a feed channel for a demo paper matching instance
func (p *PriceMirrorPipeline) Subscribe(subID string) <-chan MarketTick {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan MarketTick, 1024)
	p.subscribers[subID] = ch
	return ch
}

func main() {
	pipe := NewPriceMirrorPipeline()
	pipe.IngestProductionTick(MarketTick{
		Symbol:      "BTC/USDT",
		PriceUSD:    68500.00,
		BidUSD:      68498.50,
		AskUSD:      68501.50,
		Volume24h:   1250000000.0,
		TimestampMs: time.Now().UnixMilli(),
	})
	fmt.Println("Live Market Price Mirroring Pipeline active.")
}
