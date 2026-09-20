package cache

import (
	"sort"
	"sync"
	"time"
)

// OrderBookSide represents bids or asks.
type OrderBookSide string

const (
	SideBid OrderBookSide = "BIDS"
	SideAsk OrderBookSide = "ASKS"
)

// OrderBookCache maintains real-time L2 orderbook depth using sorted sets.
type OrderBookCache struct {
	mu     sync.RWMutex
	symbol string
	bids   map[int64]*OrderBookLevel // price -> level
	asks   map[int64]*OrderBookLevel // price -> level
}

func NewOrderBookCache(symbol string) *OrderBookCache {
	return &OrderBookCache{
		symbol: symbol,
		bids:   make(map[int64]*OrderBookLevel),
		asks:   make(map[int64]*OrderBookLevel),
	}
}

// SetPriceLevel updates or inserts a price level. If quantity <= 0, the level is removed.
func (c *OrderBookCache) SetPriceLevel(side OrderBookSide, price int64, quantity int64, orders int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	targetMap := c.bids
	if side == SideAsk {
		targetMap = c.asks
	}

	if quantity <= 0 {
		delete(targetMap, price)
		return
	}

	targetMap[price] = &OrderBookLevel{
		Price:    price,
		Quantity: quantity,
		Orders:   orders,
	}
}

// GetDepth returns the top N price levels for bids and asks.
// Bids: descending by price (highest bid first)
// Asks: ascending by price (lowest ask first)
func (c *OrderBookCache) GetDepth(depth int) OrderBookDepth {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if depth <= 0 {
		depth = 10
	}

	bidsList := make([]OrderBookLevel, 0, len(c.bids))
	for _, l := range c.bids {
		bidsList = append(bidsList, *l)
	}
	sort.Slice(bidsList, func(i, j int) bool {
		return bidsList[i].Price > bidsList[j].Price // Highest first
	})

	asksList := make([]OrderBookLevel, 0, len(c.asks))
	for _, l := range c.asks {
		asksList = append(asksList, *l)
	}
	sort.Slice(asksList, func(i, j int) bool {
		return asksList[i].Price < asksList[j].Price // Lowest first
	})

	if len(bidsList) > depth {
		bidsList = bidsList[:depth]
	}
	if len(asksList) > depth {
		asksList = asksList[:depth]
	}

	return OrderBookDepth{
		Symbol:    c.symbol,
		Bids:      bidsList,
		Asks:      asksList,
		UpdatedAt: time.Now().UTC(),
	}
}

// GetBestBidAsk returns top of book.
func (c *OrderBookCache) GetBestBidAsk() (bestBid *OrderBookLevel, bestAsk *OrderBookLevel) {
	depth := c.GetDepth(1)
	if len(depth.Bids) > 0 {
		bestBid = &depth.Bids[0]
	}
	if len(depth.Asks) > 0 {
		bestAsk = &depth.Asks[0]
	}
	return bestBid, bestAsk
}
