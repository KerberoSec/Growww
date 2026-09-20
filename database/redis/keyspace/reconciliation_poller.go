package keyspace

import (
	"context"
	"sync"
	"time"
)

// ActiveTimer represents an indexed expiration timer in Redis ZSET / SQL.
type ActiveTimer struct {
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ActiveReconciliationPoller compensates for Redis Pub/Sub's at-most-once delivery.
// Periodically polls sorted active timers to dispatch any expired keys that missed pub/sub.
type ActiveReconciliationPoller struct {
	mu         sync.Mutex
	dispatcher *ExpiryDispatcher
	timers     map[string]time.Time // key -> expires_at
	pollTicker *time.Ticker
	stopChan   chan struct{}
	running    bool
}

// NewActiveReconciliationPoller creates a fallback reconciliation poller.
func NewActiveReconciliationPoller(dispatcher *ExpiryDispatcher) *ActiveReconciliationPoller {
	return &ActiveReconciliationPoller{
		dispatcher: dispatcher,
		timers:     make(map[string]time.Time),
		stopChan:   make(chan struct{}),
	}
}

// AddTimer indexes a timer for active tracking.
func (p *ActiveReconciliationPoller) AddTimer(key string, expiresAt time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.timers[key] = expiresAt
}

// RemoveTimer removes an indexed timer once acknowledged or cancelled.
func (p *ActiveReconciliationPoller) RemoveTimer(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.timers, key)
}

// ScanAndDispatchExpired scans for overdue timers and dispatches them.
func (p *ActiveReconciliationPoller) ScanAndDispatchExpired(ctx context.Context, now time.Time) int {
	p.mu.Lock()
	var expiredKeys []string
	for key, exp := range p.timers {
		if !exp.After(now) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	p.mu.Unlock()

	dispatchedCount := 0
	for _, key := range expiredKeys {
		event := ParseKey(key)
		_ = p.dispatcher.Dispatch(ctx, event)
		p.RemoveTimer(key)
		dispatchedCount++
	}

	return dispatchedCount
}
