package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCacheAsideAndInvalidationBus(t *testing.T) {
	bus := NewInvalidationBus("growww:cache:invalidate")
	sharedL2 := NewInMemoryL2Store()

	node1 := NewCacheAsideEngine("node-1", sharedL2, bus, 10*time.Minute)
	node2 := NewCacheAsideEngine("node-2", sharedL2, bus, 10*time.Minute)

	callCount := 0
	loader := func(ctx context.Context) (interface{}, error) {
		callCount++
		return "data-v1", nil
	}

	key := FormatKey("prod", "identity", "user", "usr-101")

	// 1. Initial get on Node 1 triggers loader
	val, err := node1.GetOrCompute(context.Background(), key, loader)
	if err != nil || val != "data-v1" {
		t.Fatalf("unexpected load: %v, val: %v", err, val)
	}
	if callCount != 1 {
		t.Errorf("expected callCount 1, got %d", callCount)
	}

	// 2. Second get on Node 1 should hit L1 (no loader call)
	val, _ = node1.GetOrCompute(context.Background(), key, loader)
	if val != "data-v1" || callCount != 1 {
		t.Errorf("expected L1 hit on Node 1")
	}

	// 3. Get on Node 2 should hit shared L2 (no loader call)
	val, _ = node2.GetOrCompute(context.Background(), key, loader)
	if val != "data-v1" || callCount != 1 {
		t.Errorf("expected L2 hit on Node 2 without calling loader")
	}

	// 4. Node 1 invalidates key -> L1 on Node 1 and Node 2 should be invalidated via Pub/Sub
	if err := node1.Invalidate(context.Background(), key); err != nil {
		t.Fatalf("failed to invalidate: %v", err)
	}

	// Give pub/sub synchronous bus a microsecond
	if _, _, found := node2.l1.Get(key); found {
		t.Errorf("expected Node 2 L1 to be invalidated via pub/sub")
	}

	// Now Node 2 fetching should cause miss in L1 & L2 and invoke loader again
	val, _ = node2.GetOrCompute(context.Background(), key, loader)
	if callCount != 2 {
		t.Errorf("expected loader called again after invalidation, got %d", callCount)
	}
}

func TestSlidingWindowRateLimiter(t *testing.T) {
	limiter := NewSlidingWindowRateLimiter()
	key := FormatKey("prod", "ratelimit", "user", "usr-limit-test")

	// Limit: 5 requests per 100ms
	limit := int64(5)
	window := 100 * time.Millisecond

	for i := 0; i < 5; i++ {
		res, err := limiter.Allow(context.Background(), key, limit, window)
		if err != nil || !res.Allowed {
			t.Fatalf("request %d should be allowed, got err: %v", i, err)
		}
		if res.Remaining != limit-int64(i)-1 {
			t.Errorf("expected remaining %d, got %d", limit-int64(i)-1, res.Remaining)
		}
	}

	// 6th request must be rejected
	res, err := limiter.Allow(context.Background(), key, limit, window)
	if err != ErrRateLimitExceeded || res.Allowed {
		t.Fatalf("expected rate limit exceeded, allowed: %v, err: %v", res.Allowed, err)
	}

	// Wait for window to slide
	time.Sleep(110 * time.Millisecond)

	// Now request should be allowed again
	res, err = limiter.Allow(context.Background(), key, limit, window)
	if err != nil || !res.Allowed {
		t.Fatalf("request after window slide should be allowed, got err: %v", err)
	}
}

func TestDistributedLockWithFencing(t *testing.T) {
	lockMgr := NewDistributedLockManager(2 * time.Second)
	resKey := FormatKey("prod", "lock", "order", "ord-999")

	// 1. Acquire lock
	handle1, err := lockMgr.Acquire(context.Background(), resKey, 1*time.Second)
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}
	if handle1.FencingSeq <= 0 {
		t.Errorf("expected positive fencing sequence, got %d", handle1.FencingSeq)
	}

	// 2. Second acquire on same resource while held must fail
	_, err = lockMgr.Acquire(context.Background(), resKey, 1*time.Second)
	if err != ErrLockAcquisitionFailed {
		t.Fatalf("expected ErrLockAcquisitionFailed, got %v", err)
	}

	// 3. Lease extension
	if err := lockMgr.ExtendLease(context.Background(), handle1, 500*time.Millisecond); err != nil {
		t.Fatalf("failed to extend lease: %v", err)
	}

	// 4. Safe release
	if err := lockMgr.Release(context.Background(), handle1); err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	// 5. Subsequent acquire should succeed with higher fencing token sequence
	handle2, err := lockMgr.Acquire(context.Background(), resKey, 1*time.Second)
	if err != nil {
		t.Fatalf("failed to re-acquire lock: %v", err)
	}
	if handle2.FencingSeq <= handle1.FencingSeq {
		t.Errorf("fencing token must be monotonically increasing (%d <= %d)", handle2.FencingSeq, handle1.FencingSeq)
	}
	_ = lockMgr.Release(context.Background(), handle2)
}

func TestOrderBookL2DepthCache(t *testing.T) {
	ob := NewOrderBookCache("INFY-INR")

	// Insert bids (buyers want highest price)
	ob.SetPriceLevel(SideBid, 15000000, 100000000, 5) // 1500.00, qty 100
	ob.SetPriceLevel(SideBid, 15050000, 50000000, 2)  // 1505.00, qty 50
	ob.SetPriceLevel(SideBid, 14950000, 200000000, 8) // 1495.00, qty 200

	// Insert asks (sellers want lowest price)
	ob.SetPriceLevel(SideAsk, 15100000, 75000000, 3)  // 1510.00, qty 75
	ob.SetPriceLevel(SideAsk, 15080000, 25000000, 1)  // 1508.00, qty 25
	ob.SetPriceLevel(SideAsk, 15150000, 150000000, 4) // 1515.00, qty 150

	depth := ob.GetDepth(2)

	// Bids should be sorted descending: 1505.00, then 1500.00
	if len(depth.Bids) != 2 {
		t.Fatalf("expected 2 bids, got %d", len(depth.Bids))
	}
	if depth.Bids[0].Price != 15050000 {
		t.Errorf("expected best bid 15050000, got %d", depth.Bids[0].Price)
	}
	if depth.Bids[1].Price != 15000000 {
		t.Errorf("expected second bid 15000000, got %d", depth.Bids[1].Price)
	}

	// Asks should be sorted ascending: 1508.00, then 1510.00
	if len(depth.Asks) != 2 {
		t.Fatalf("expected 2 asks, got %d", len(depth.Asks))
	}
	if depth.Asks[0].Price != 15080000 {
		t.Errorf("expected best ask 15080000, got %d", depth.Asks[0].Price)
	}
	if depth.Asks[1].Price != 15100000 {
		t.Errorf("expected second ask 15100000, got %d", depth.Asks[1].Price)
	}

	bestBid, bestAsk := ob.GetBestBidAsk()
	if bestBid.Price != 15050000 || bestAsk.Price != 15080000 {
		t.Errorf("best bid/ask mismatch")
	}

	// Deletion via quantity = 0
	ob.SetPriceLevel(SideBid, 15050000, 0, 0)
	depthAfter := ob.GetDepth(1)
	if depthAfter.Bids[0].Price != 15000000 {
		t.Errorf("expected new best bid 15000000 after deletion, got %d", depthAfter.Bids[0].Price)
	}
}

func TestComplianceCache(t *testing.T) {
	comp := NewComplianceCache()
	addr := "0x71C841832046a64ce5560A8146F255aB87C29bA2"

	// Whitelist address
	comp.SetWhitelist(addr, ComplianceStatusActive)

	status, exists := comp.CheckWhitelist(context.Background(), addr)
	if !exists || status != ComplianceStatusActive {
		t.Errorf("expected active whitelist status")
	}

	// JWT Revocation
	jti := "jwt-uuid-12345"
	if comp.IsJWTRevoked(jti) {
		t.Errorf("new JWT should not be revoked")
	}

	comp.RevokeJWT(jti)
	if !comp.IsJWTRevoked(jti) {
		t.Errorf("revoked JWT must report true")
	}
}

func TestConcurrentCacheStampedeJitter(t *testing.T) {
	bus := NewInvalidationBus("growww:cache:stampede")
	l2 := NewInMemoryL2Store()
	engine := NewCacheAsideEngine("node-test", l2, bus, 1*time.Second)

	key := FormatKey("prod", "market", "quote", "TCS-INR")
	var computeCount int64
	var mu sync.Mutex

	loader := func(ctx context.Context) (interface{}, error) {
		mu.Lock()
		computeCount++
		mu.Unlock()
		time.Sleep(10 * time.Millisecond) // simulate DB delay
		return "quote-data", nil
	}

	var wg sync.WaitGroup
	// 50 concurrent requests
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			val, err := engine.GetOrCompute(context.Background(), key, loader)
			if err != nil || val != "quote-data" {
				t.Errorf("concurrent get failed: %v", err)
			}
		}()
	}
	wg.Wait()

	// Initial load should only happen once or twice under concurrency (not 50 times)
	mu.Lock()
	count := computeCount
	mu.Unlock()

	if count > 5 {
		t.Errorf("cache stampede occurred: %d DB computations executed", count)
	}
}
