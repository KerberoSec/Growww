package src

import (
	"math"
	"sync"
	"time"
)

// RateLimitTier represents client tier for rate limiting.
type RateLimitTier string

const (
	TierPublic        RateLimitTier = "PUBLIC"
	TierRetailAuth    RateLimitTier = "RETAIL_AUTH"
	TierInstitutional RateLimitTier = "INSTITUTIONAL"
)

// TierPolicy defines rate limits and burst capacities per tier.
type TierPolicy struct {
	RefillRatePerSec float64 `json:"refill_rate_per_sec"`
	BurstCapacity    float64 `json:"burst_capacity"`
}

// DefaultTierPolicies returns production policy profiles.
func DefaultTierPolicies() map[RateLimitTier]TierPolicy {
	return map[RateLimitTier]TierPolicy{
		TierPublic: {
			RefillRatePerSec: 10.0,
			BurstCapacity:    20.0,
		},
		TierRetailAuth: {
			RefillRatePerSec: 50.0,
			BurstCapacity:    100.0,
		},
		TierInstitutional: {
			RefillRatePerSec: 500.0,
			BurstCapacity:    1000.0,
		},
	}
}

// TokenBucket represents an individual client's token bucket state.
type TokenBucket struct {
	mu            sync.Mutex
	tokens        float64
	burstCapacity float64
	refillRate    float64
	lastRefill    time.Time
}

// RateLimitResult conveys the rate limiting decision and metadata.
type RateLimitResult struct {
	Allowed         bool          `json:"allowed"`
	RemainingTokens float64       `json:"remaining_tokens"`
	RetryAfter      time.Duration `json:"retry_after"`
}

// DistributedTokenBucketLimiter manages token buckets across multiple keys.
type DistributedTokenBucketLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*TokenBucket
	policies map[RateLimitTier]TierPolicy
}

// NewDistributedTokenBucketLimiter initializes a new rate limiter.
func NewDistributedTokenBucketLimiter(policies map[RateLimitTier]TierPolicy) *DistributedTokenBucketLimiter {
	if policies == nil {
		policies = DefaultTierPolicies()
	}
	return &DistributedTokenBucketLimiter{
		buckets:  make(map[string]*TokenBucket),
		policies: policies,
	}
}

// Allow evaluates if a request with given cost can proceed for a key and tier.
func (l *DistributedTokenBucketLimiter) Allow(key string, tier RateLimitTier, cost float64) RateLimitResult {
	if cost <= 0 {
		cost = 1.0
	}

	policy, exists := l.policies[tier]
	if !exists {
		policy = l.policies[TierPublic]
	}

	l.mu.Lock()
	bucket, ok := l.buckets[key]
	if !ok {
		bucket = &TokenBucket{
			tokens:        policy.BurstCapacity,
			burstCapacity: policy.BurstCapacity,
			refillRate:    policy.RefillRatePerSec,
			lastRefill:    time.Now().UTC(),
		}
		l.buckets[key] = bucket
	}
	l.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	now := time.Now().UTC()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens = math.Min(bucket.burstCapacity, bucket.tokens+(elapsed*bucket.refillRate))
	bucket.lastRefill = now

	if bucket.tokens >= cost {
		bucket.tokens -= cost
		return RateLimitResult{
			Allowed:         true,
			RemainingTokens: bucket.tokens,
			RetryAfter:      0,
		}
	}

	deficit := cost - bucket.tokens
	retrySeconds := deficit / bucket.refillRate
	return RateLimitResult{
		Allowed:         false,
		RemainingTokens: bucket.tokens,
		RetryAfter:      time.Duration(retrySeconds * float64(time.Second)),
	}
}
