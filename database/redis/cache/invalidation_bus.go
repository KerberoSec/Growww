package cache

import (
	"sync"
	"time"
)

// InvalidationSubscriber receives invalidation events.
type InvalidationSubscriber interface {
	OnInvalidation(msg InvalidationMessage)
}

// InvalidationBus coordinates pub/sub invalidations between L1 local caches.
type InvalidationBus struct {
	mu          sync.RWMutex
	subscribers []InvalidationSubscriber
	channelName string
}

func NewInvalidationBus(channelName string) *InvalidationBus {
	if channelName == "" {
		channelName = "growww:cache:invalidate"
	}
	return &InvalidationBus{
		channelName: channelName,
		subscribers: make([]InvalidationSubscriber, 0),
	}
}

// Subscribe registers a subscriber (typically an L1/L2 engine instance).
func (b *InvalidationBus) Subscribe(sub InvalidationSubscriber) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers = append(b.subscribers, sub)
}

// PublishInvalidation broadcasts a key invalidation across all registered subscribers.
func (b *InvalidationBus) PublishInvalidation(key string, nodeID string) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msg := InvalidationMessage{
		Key:       key,
		NodeID:    nodeID,
		Timestamp: time.Now().UTC(),
	}

	for _, s := range b.subscribers {
		s.OnInvalidation(msg)
	}
}
