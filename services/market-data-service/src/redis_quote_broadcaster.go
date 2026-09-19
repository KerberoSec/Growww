package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type QuoteData struct {
	Symbol     string  `json:"symbol"`
	LastPrice  float64 `json:"last_price"`
	Bid        float64 `json:"bid"`
	Ask        float64 `json:"ask"`
	Change24h  float64 `json:"change_24h_pct"`
	Volume24h  float64 `json:"volume_24h"`
	UpdatedAt  int64   `json:"updated_at_ms"`
}

type RedisQuoteBroadcaster struct {
	mu          sync.RWMutex
	cache       map[string]QuoteData // Local L1 in-memory cache
	redisPrefix string
}

func NewRedisQuoteBroadcaster(prefix string) *RedisQuoteBroadcaster {
	return &RedisQuoteBroadcaster{
		cache:       make(map[string]QuoteData),
		redisPrefix: prefix,
	}
}

// PublishQuote caches quote in local memory and publishes to Redis Pub/Sub cluster channel
func (b *RedisQuoteBroadcaster) PublishQuote(ctx context.Context, quote QuoteData) ([]byte, error) {
	quote.UpdatedAt = time.Now().UnixMilli()

	b.mu.Lock()
	b.cache[quote.Symbol] = quote
	b.mu.Unlock()

	payload, err := json.Marshal(quote)
	if err != nil {
		return nil, err
	}

	channel := fmt.Sprintf("%s:quotes:%s", b.redisPrefix, quote.Symbol)
	// In production, execute: rdb.Publish(ctx, channel, payload)
	_ = channel
	return payload, nil
}

// GetCachedQuote reads from low-latency local L1 cache
func (b *RedisQuoteBroadcaster) GetCachedQuote(symbol string) (QuoteData, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	q, exists := b.cache[symbol]
	return q, exists
}
