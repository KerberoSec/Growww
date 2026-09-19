package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucket_AllowWithinCapacity(t *testing.T) {
	tb := NewTokenBucket(5, 1.0)
	now := time.Now()

	for i := 0; i < 5; i++ {
		d := tb.Allow(now)
		if !d.Allowed {
			t.Errorf("request %d should be allowed", i)
		}
	}
	// 6th request should be denied
	d := tb.Allow(now)
	if d.Allowed {
		t.Error("6th request should be denied (bucket exhausted)")
	}
	if d.RetryAfter <= 0 {
		t.Error("RetryAfter should be positive")
	}
}

func TestTokenBucket_RefillOverTime(t *testing.T) {
	tb := NewTokenBucket(2, 1.0)
	now := time.Now()

	// Exhaust tokens
	tb.Allow(now)
	tb.Allow(now)

	d := tb.Allow(now)
	if d.Allowed {
		t.Error("should be denied after exhaustion")
	}

	// Wait 2 seconds for refill
	future := now.Add(2 * time.Second)
	d = tb.Allow(future)
	if !d.Allowed {
		t.Error("should be allowed after refill period")
	}
}

func TestSlidingWindow_AllowWithinLimit(t *testing.T) {
	sw := NewSlidingWindow(3, time.Minute)
	now := time.Now()

	for i := 0; i < 3; i++ {
		d := sw.Allow(now)
		if !d.Allowed {
			t.Errorf("request %d should be allowed", i)
		}
		if d.Remaining != int64(2-i) {
			t.Errorf("expected remaining %d, got %d", 2-i, d.Remaining)
		}
	}

	// 4th request should be denied
	d := sw.Allow(now)
	if d.Allowed {
		t.Error("4th request should be denied")
	}
}

func TestSlidingWindow_WindowExpiry(t *testing.T) {
	sw := NewSlidingWindow(2, time.Second)
	now := time.Now()

	sw.Allow(now)
	sw.Allow(now)

	d := sw.Allow(now)
	if d.Allowed {
		t.Error("should be denied")
	}

	// After window expires, requests should be allowed again
	future := now.Add(2 * time.Second)
	d = sw.Allow(future)
	if !d.Allowed {
		t.Error("should be allowed after window expiry")
	}
}

func TestRateLimiter_PerUserTier(t *testing.T) {
	rl := NewRateLimiter(&TierConfig{
		Algorithm:      AlgoTokenBucket,
		MaxRequests:    10,
		RefillRate:     1.0,
	})

	_ = rl.RegisterTier("basic", &TierConfig{
		Algorithm:   AlgoTokenBucket,
		MaxRequests: 3,
		RefillRate:  0.5,
	})

	// User on basic tier
	for i := 0; i < 3; i++ {
		d := rl.Check("basic", TierUser, "user-123")
		if !d.Allowed {
			t.Errorf("request %d should be allowed for basic user", i)
		}
	}
	d := rl.Check("basic", TierUser, "user-123")
	if d.Allowed {
		t.Error("basic tier should be exhausted")
	}
}

func TestRateLimiter_PerIPSlidingWindow(t *testing.T) {
	rl := NewRateLimiter(&TierConfig{
		Algorithm:      AlgoSlidingWindow,
		MaxRequests:    100,
		WindowDuration: time.Minute,
	})

	_ = rl.RegisterTier("ip_limit", &TierConfig{
		Algorithm:      AlgoSlidingWindow,
		MaxRequests:    5,
		WindowDuration: time.Second,
	})

	for i := 0; i < 5; i++ {
		d := rl.Check("ip_limit", TierIP, "192.168.1.1")
		if !d.Allowed {
			t.Errorf("request %d should be allowed", i)
		}
	}
	d := rl.Check("ip_limit", TierIP, "192.168.1.1")
	if d.Allowed {
		t.Error("should be rate limited")
	}

	// Different IP should have its own window
	d = rl.Check("ip_limit", TierIP, "10.0.0.1")
	if !d.Allowed {
		t.Error("different IP should be allowed")
	}
}

func TestRateLimiter_DefaultTierFallback(t *testing.T) {
	rl := NewRateLimiter(&TierConfig{
		Algorithm:   AlgoTokenBucket,
		MaxRequests: 2,
		RefillRate:  1.0,
	})

	// Use unknown tier name - should fall back to default
	d := rl.Check("nonexistent", TierAPIKey, "key-abc")
	if !d.Allowed {
		t.Error("should be allowed with default tier")
	}
}

func TestRateLimiter_Reset(t *testing.T) {
	rl := NewRateLimiter(&TierConfig{
		Algorithm:   AlgoTokenBucket,
		MaxRequests: 1,
		RefillRate:  0.01,
	})

	d := rl.Check("", TierUser, "user-1")
	if !d.Allowed {
		t.Error("first request should be allowed")
	}
	d = rl.Check("", TierUser, "user-1")
	if d.Allowed {
		t.Error("should be rate limited")
	}

	// Reset and try again
	rl.Reset(TierUser, "user-1")
	d = rl.Check("", TierUser, "user-1")
	if !d.Allowed {
		t.Error("should be allowed after reset")
	}
}

func TestRateLimiter_Stats(t *testing.T) {
	rl := NewRateLimiter(&TierConfig{
		Algorithm:   AlgoTokenBucket,
		MaxRequests: 10,
		RefillRate:  1.0,
	})
	_ = rl.RegisterTier("premium", &TierConfig{
		Algorithm:   AlgoTokenBucket,
		MaxRequests: 100,
		RefillRate:  10.0,
	})

	rl.Check("", TierUser, "u1")
	rl.Check("", TierUser, "u2")

	stats := rl.GetStats()
	if stats.ActiveBuckets != 2 {
		t.Errorf("expected 2 active buckets, got %d", stats.ActiveBuckets)
	}
	if stats.TierCount != 1 {
		t.Errorf("expected 1 tier, got %d", stats.TierCount)
	}
}
