package src

import (
	"testing"
	"time"
)

func TestTokenBucketRateLimiting(t *testing.T) {
	limiter := NewDistributedTokenBucketLimiter(nil)

	// Public tier: 10 ref/s, burst 20
	key := "client-ip-192.168.1.100"

	// 1. Consume 20 burst tokens
	for i := 0; i < 20; i++ {
		res := limiter.Allow(key, TierPublic, 1.0)
		if !res.Allowed {
			t.Fatalf("request %d should be allowed within burst capacity", i)
		}
	}

	// 2. 21st request should be rejected
	res := limiter.Allow(key, TierPublic, 1.0)
	if res.Allowed {
		t.Fatalf("21st request should be rate limited")
	}
	if res.RetryAfter <= 0 {
		t.Fatalf("expected positive RetryAfter, got %v", res.RetryAfter)
	}

	// 3. Institutional tier has 1000 burst capacity
	instKey := "inst-fund-alpha"
	resInst := limiter.Allow(instKey, TierInstitutional, 100.0)
	if !resInst.Allowed || resInst.RemainingTokens != 900.0 {
		t.Fatalf("expected institutional request to allow 100 cost with 900 remaining, got allowed=%v rem=%.2f", resInst.Allowed, resInst.RemainingTokens)
	}

	// Wait a moment for public key refill
	time.Sleep(120 * time.Millisecond) // ~1 token refilled
	resRefill := limiter.Allow(key, TierPublic, 1.0)
	if !resRefill.Allowed {
		t.Fatalf("expected request after refill to be allowed")
	}
}
