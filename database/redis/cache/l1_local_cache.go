package cache

import (
	"sync"
	"time"
)

type l1Item struct {
	value      interface{}
	expiresAt  time.Time
	deltaDelta time.Duration // computation time delta for XFetch
}

// L1LocalCache is an in-memory thread-safe local cache.
type L1LocalCache struct {
	mu    sync.RWMutex
	items map[string]l1Item
}

func NewL1LocalCache() *L1LocalCache {
	return &L1LocalCache{
		items: make(map[string]l1Item),
	}
}

// Get retrieves an item from L1. Returns nil if missing or expired.
func (c *L1LocalCache) Get(key string) (interface{}, time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, 0, false
	}

	now := time.Now()
	if !item.expiresAt.IsZero() && now.After(item.expiresAt) {
		return nil, 0, false
	}

	remainingTTL := item.expiresAt.Sub(now)
	return item.value, remainingTTL, true
}

// Set stores a key-value pair in L1 with a TTL.
func (c *L1LocalCache) Set(key string, val interface{}, ttl time.Duration, delta time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	c.items[key] = l1Item{
		value:      val,
		expiresAt:  exp,
		deltaDelta: delta,
	}
}

// Invalidate removes a single key from L1 cache.
func (c *L1LocalCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// InvalidateAll flushes all keys from L1.
func (c *L1LocalCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]l1Item)
}

// Len returns current item count.
func (c *L1LocalCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
