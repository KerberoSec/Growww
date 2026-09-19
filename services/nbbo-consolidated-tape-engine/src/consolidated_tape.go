package main

import (
	"sync"
	"time"
)

type VenueQuote struct {
	Venue    string  // NSE, BSE, NBSE
	BidPrice float64
	BidSize  uint64
	AskPrice float64
	AskSize  uint64
}

type NBBORecord struct {
	Symbol      string
	BestBid     float64
	BestBidSize uint64
	BestBidVen  string
	BestAsk     float64
	BestAskSize uint64
	BestAskVen  string
	Spread      float64
	Timestamp   time.Time
}

type ConsolidatedTapeEngine struct {
	mu     sync.RWMutex
	venues map[string]map[string]VenueQuote // symbol -> (venue -> quote)
}

func NewConsolidatedTapeEngine() *ConsolidatedTapeEngine {
	return &ConsolidatedTapeEngine{
		venues: make(map[string]map[string]VenueQuote),
	}
}

// IngestVenueQuote aggregates cross-exchange quotes and calculates National Best Bid and Offer (NBBO)
func (e *ConsolidatedTapeEngine) IngestVenueQuote(symbol string, quote VenueQuote) NBBORecord {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.venues[symbol]; !exists {
		e.venues[symbol] = make(map[string]VenueQuote)
	}
	e.venues[symbol][quote.Venue] = quote

	var bestBid float64
	var bestBidSize uint64
	var bestBidVen string

	var bestAsk float64 = 1e12
	var bestAskSize uint64
	var bestAskVen string

	for v, q := range e.venues[symbol] {
		if q.BidPrice > bestBid {
			bestBid = q.BidPrice
			bestBidSize = q.BidSize
			bestBidVen = v
		}
		if q.AskPrice > 0 && q.AskPrice < bestAsk {
			bestAsk = q.AskPrice
			bestAskSize = q.AskSize
			bestAskVen = v
		}
	}

	spread := 0.0
	if bestAsk < 1e12 && bestBid > 0 {
		spread = bestAsk - bestBid
	}

	return NBBORecord{
		Symbol:      symbol,
		BestBid:     bestBid,
		BestBidSize: bestBidSize,
		BestBidVen:  bestBidVen,
		BestAsk:     bestAsk,
		BestAskSize: bestAskSize,
		BestAskVen:  bestAskVen,
		Spread:      spread,
		Timestamp:   time.Now().UTC(),
	}
}
