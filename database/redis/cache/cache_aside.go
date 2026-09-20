package cache

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

// CacheMetrics tracks cache hits, misses, and preemptions.
type CacheMetrics struct {
	L1Hits             uint64
	L2Hits             uint64
	Misses             uint64
	PreemptiveRefreshes uint64
	Invalidations      uint64
}

// L2DistributedStore abstracts the shared distributed Redis cache tier.
type L2DistributedStore interface {
	Get(ctx context.Context, key string) (interface{}, time.Duration, error)
	Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// InMemoryL2Store provides an institutional in-memory simulation of Redis L2.
type InMemoryL2Store struct {
	mu    sync.RWMutex
	items map[string]l1Item
}

func NewInMemoryL2Store() *InMemoryL2Store {
	return &InMemoryL2Store{
		items: make(map[string]l1Item),
	}
}

func (s *InMemoryL2Store) Get(ctx context.Context, key string) (interface{}, time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.items[key]
	if !exists {
		return nil, 0, ErrCacheMiss
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return nil, 0, ErrCacheMiss
	}

	remaining := time.Until(item.expiresAt)
	return item.value, remaining, nil
}

func (s *InMemoryL2Store) Set(ctx context.Context, key string, val interface{}, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	s.items[key] = l1Item{
		value:     val,
		expiresAt: exp,
	}
	return nil
}

func (s *InMemoryL2Store) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key)
	return nil
}

// CacheAsideEngine manages two-tier (L1 Memory + L2 Redis) cache-aside with XFetch jitter.
type flightCall struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type CacheAsideEngine struct {
	nodeID          string
	l1              *L1LocalCache
	l2              L2DistributedStore
	bus             *InvalidationBus
	beta            float64 // XFetch beta parameter (typically 1.0)
	defaultTTL      time.Duration
	metrics         CacheMetrics
	refreshInFlight sync.Map
	flightMu        sync.Mutex
	flights         map[string]*flightCall
}

func NewCacheAsideEngine(nodeID string, l2 L2DistributedStore, bus *InvalidationBus, defaultTTL time.Duration) *CacheAsideEngine {
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Minute
	}
	engine := &CacheAsideEngine{
		nodeID:     nodeID,
		l1:         NewL1LocalCache(),
		l2:         l2,
		bus:        bus,
		beta:       1.0,
		defaultTTL: defaultTTL,
		flights:    make(map[string]*flightCall),
	}

	if bus != nil {
		bus.Subscribe(engine)
	}

	return engine
}

// OnInvalidation handles incoming invalidation pub/sub messages from peer nodes.
func (e *CacheAsideEngine) OnInvalidation(msg InvalidationMessage) {
	// If message came from another node, evict from local L1
	e.l1.Invalidate(msg.Key)
	atomic.AddUint64(&e.metrics.Invalidations, 1)
}

// GetOrCompute loads a value via Cache-Aside: L1 -> L2 -> loader().
// Employs singleflight request deduplication and the XFetch optimal probabilistic algorithm to prevent cache stampedes.
func (e *CacheAsideEngine) GetOrCompute(
	ctx context.Context,
	key string,
	loader func(ctx context.Context) (interface{}, error),
) (interface{}, error) {
	// 1. Check L1 local cache
	val, ttlRemaining, found := e.l1.Get(key)
	if found {
		atomic.AddUint64(&e.metrics.L1Hits, 1)
		// Check XFetch probabilistic early refresh
		e.maybePreemptiveRefresh(ctx, key, ttlRemaining, loader)
		return val, nil
	}

	// 2. Check L2 distributed cache
	valL2, ttlL2, err := e.l2.Get(ctx, key)
	if err == nil {
		atomic.AddUint64(&e.metrics.L2Hits, 1)
		// Populate L1
		e.l1.Set(key, valL2, ttlL2, 10*time.Millisecond)
		e.maybePreemptiveRefresh(ctx, key, ttlL2, loader)
		return valL2, nil
	}

	// 3. Cache miss: use singleflight deduplication so concurrent requests do not stampede the DB
	e.flightMu.Lock()
	if c, inFlight := e.flights[key]; inFlight {
		e.flightMu.Unlock()
		c.wg.Wait()
		if c.err != nil {
			return nil, c.err
		}
		return c.val, nil
	}

	call := &flightCall{}
	call.wg.Add(1)
	e.flights[key] = call
	e.flightMu.Unlock()

	defer func() {
		e.flightMu.Lock()
		delete(e.flights, key)
		e.flightMu.Unlock()
		call.wg.Done()
	}()

	atomic.AddUint64(&e.metrics.Misses, 1)
	start := time.Now()
	freshVal, err := loader(ctx)
	if err != nil {
		call.err = err
		return nil, err
	}
	delta := time.Since(start)

	// Write to L2 and L1
	_ = e.l2.Set(ctx, key, freshVal, e.defaultTTL)
	e.l1.Set(key, freshVal, e.defaultTTL, delta)

	call.val = freshVal
	return freshVal, nil
}

// maybePreemptiveRefresh uses XFetch algorithm:
// Refresh if: -beta * delta * ln(rand()) > remaining_ttl
func (e *CacheAsideEngine) maybePreemptiveRefresh(
	ctx context.Context,
	key string,
	remainingTTL time.Duration,
	loader func(ctx context.Context) (interface{}, error),
) {
	if remainingTTL <= 0 {
		return
	}

	delta := 10 * time.Millisecond // computation duration estimate
	r := rand.Float64()
	if r <= 0 {
		r = 0.0001
	}

	xfetchThreshold := -e.beta * float64(delta.Milliseconds()) * math.Log(r)
	if xfetchThreshold > float64(remainingTTL.Milliseconds()) {
		// Probabilistic condition met: trigger background refresh if not already in flight
		if _, loaded := e.refreshInFlight.LoadOrStore(key, true); !loaded {
			atomic.AddUint64(&e.metrics.PreemptiveRefreshes, 1)
			go func() {
				defer e.refreshInFlight.Delete(key)
				start := time.Now()
				freshVal, err := loader(context.Background())
				if err == nil {
					d := time.Since(start)
					_ = e.l2.Set(context.Background(), key, freshVal, e.defaultTTL)
					e.l1.Set(key, freshVal, e.defaultTTL, d)
				}
			}()
		}
	}
}

// Invalidate updates L2, removes from L1, and broadcasts an invalidation pub/sub event.
func (e *CacheAsideEngine) Invalidate(ctx context.Context, key string) error {
	e.l1.Invalidate(key)
	if err := e.l2.Delete(ctx, key); err != nil {
		return err
	}

	if e.bus != nil {
		e.bus.PublishInvalidation(key, e.nodeID)
	}
	return nil
}

// Metrics returns the operational statistics.
func (e *CacheAsideEngine) Metrics() CacheMetrics {
	return CacheMetrics{
		L1Hits:             atomic.LoadUint64(&e.metrics.L1Hits),
		L2Hits:             atomic.LoadUint64(&e.metrics.L2Hits),
		Misses:             atomic.LoadUint64(&e.metrics.Misses),
		PreemptiveRefreshes: atomic.LoadUint64(&e.metrics.PreemptiveRefreshes),
		Invalidations:      atomic.LoadUint64(&e.metrics.Invalidations),
	}
}
