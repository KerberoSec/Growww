// Package ratelimit provides a production-grade rate limiting service for the
// NBSE sovereign exchange. It implements Token Bucket and Sliding Window
// algorithms with a Redis-compatible in-memory store. Supports per-user,
// per-IP, and per-API-key tiers with configurable limits.
package ratelimit

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Algorithm selects the rate limiting strategy.
type Algorithm string

const (
	AlgoTokenBucket   Algorithm = "token_bucket"
	AlgoSlidingWindow Algorithm = "sliding_window"
)

// TierType categorizes rate limit subjects.
type TierType string

const (
	TierUser   TierType = "user"
	TierIP     TierType = "ip"
	TierAPIKey TierType = "api_key"
)

// TierConfig defines rate limits for a specific tier.
type TierConfig struct {
	Name           string
	TierType       TierType
	Algorithm      Algorithm
	MaxRequests    int64         // max requests in window (sliding window) or bucket capacity (token bucket)
	WindowDuration time.Duration // sliding window size
	RefillRate     float64       // tokens per second (token bucket)
	BurstSize      int64         // max burst capacity (token bucket)
}

// Decision is the result of a rate limit check.
type Decision struct {
	Allowed     bool
	Remaining   int64
	RetryAfter  time.Duration
	Limit       int64
	ResetAt     time.Time
}

// TokenBucket implements the token bucket algorithm.
type TokenBucket struct {
	capacity   int64
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket.
func NewTokenBucket(capacity int64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     float64(capacity),
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token.
func (tb *TokenBucket) Allow(now time.Time) Decision {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on elapsed time
	if now.After(tb.lastRefill) {
		elapsed := now.Sub(tb.lastRefill).Seconds()
		tb.tokens += elapsed * tb.refillRate
		if tb.tokens > float64(tb.capacity) {
			tb.tokens = float64(tb.capacity)
		}
		tb.lastRefill = now
	}

	if tb.tokens >= 1.0 {
		tb.tokens--
		return Decision{
			Allowed:   true,
			Remaining: int64(tb.tokens),
			Limit:     tb.capacity,
			ResetAt:   now.Add(time.Duration(float64(tb.capacity-int64(tb.tokens)) / tb.refillRate * float64(time.Second))),
		}
	}

	// Calculate retry-after
	retryAfter := time.Duration((1.0 - tb.tokens) / tb.refillRate * float64(time.Second))
	return Decision{
		Allowed:    false,
		Remaining:  0,
		RetryAfter: retryAfter,
		Limit:      tb.capacity,
		ResetAt:    now.Add(retryAfter),
	}
}

// SlidingWindowEntry records a request timestamp.
type SlidingWindowEntry struct {
	Timestamp time.Time
}

// SlidingWindow implements the sliding window log algorithm.
type SlidingWindow struct {
	maxRequests int64
	windowSize  time.Duration
	entries     []SlidingWindowEntry
	mu          sync.Mutex
}

// NewSlidingWindow creates a new sliding window limiter.
func NewSlidingWindow(maxRequests int64, windowSize time.Duration) *SlidingWindow {
	return &SlidingWindow{
		maxRequests: maxRequests,
		windowSize:  windowSize,
		entries:     make([]SlidingWindowEntry, 0),
	}
}

// Allow checks if a request is allowed within the sliding window.
func (sw *SlidingWindow) Allow(now time.Time) Decision {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	// Evict expired entries
	windowStart := now.Add(-sw.windowSize)
	newEntries := make([]SlidingWindowEntry, 0, len(sw.entries))
	for _, e := range sw.entries {
		if e.Timestamp.After(windowStart) {
			newEntries = append(newEntries, e)
		}
	}
	sw.entries = newEntries

	currentCount := int64(len(sw.entries))
	if currentCount < sw.maxRequests {
		sw.entries = append(sw.entries, SlidingWindowEntry{Timestamp: now})
		return Decision{
			Allowed:   true,
			Remaining: sw.maxRequests - currentCount - 1,
			Limit:     sw.maxRequests,
			ResetAt:   now.Add(sw.windowSize),
		}
	}

	// Calculate retry-after from oldest entry
	var retryAfter time.Duration
	if len(sw.entries) > 0 {
		oldestExpiry := sw.entries[0].Timestamp.Add(sw.windowSize)
		retryAfter = oldestExpiry.Sub(now)
		if retryAfter < 0 {
			retryAfter = 0
		}
	}

	return Decision{
		Allowed:    false,
		Remaining:  0,
		RetryAfter: retryAfter,
		Limit:      sw.maxRequests,
		ResetAt:    now.Add(retryAfter),
	}
}

// RateLimiter is the main rate limiting engine that manages per-key limiters.
type RateLimiter struct {
	mu            sync.RWMutex
	tiers         map[string]*TierConfig
	tokenBuckets  map[string]*TokenBucket
	slidingWindows map[string]*SlidingWindow
	defaultTier   *TierConfig
}

// NewRateLimiter creates a new rate limiter with a default tier.
func NewRateLimiter(defaultTier *TierConfig) *RateLimiter {
	return &RateLimiter{
		tiers:          make(map[string]*TierConfig),
		tokenBuckets:   make(map[string]*TokenBucket),
		slidingWindows: make(map[string]*SlidingWindow),
		defaultTier:    defaultTier,
	}
}

// RegisterTier adds a named tier configuration.
func (rl *RateLimiter) RegisterTier(name string, tier *TierConfig) error {
	if name == "" {
		return errors.New("tier name is required")
	}
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.tiers[name] = tier
	return nil
}

// buildKey creates a composite key for limiter lookup.
func buildKey(tierType TierType, identifier string) string {
	return fmt.Sprintf("%s:%s", tierType, identifier)
}

// Check evaluates rate limit for a given identifier under a specific tier.
func (rl *RateLimiter) Check(tierName string, tierType TierType, identifier string) Decision {
	return rl.CheckAt(tierName, tierType, identifier, time.Now())
}

// CheckAt evaluates rate limit at a specific time (useful for testing).
func (rl *RateLimiter) CheckAt(tierName string, tierType TierType, identifier string, now time.Time) Decision {
	rl.mu.RLock()
	tier, ok := rl.tiers[tierName]
	rl.mu.RUnlock()
	if !ok {
		tier = rl.defaultTier
	}

	key := buildKey(tierType, identifier)

	switch tier.Algorithm {
	case AlgoTokenBucket:
		return rl.checkTokenBucket(key, tier, now)
	case AlgoSlidingWindow:
		return rl.checkSlidingWindow(key, tier, now)
	default:
		return Decision{Allowed: false, RetryAfter: time.Second}
	}
}

func (rl *RateLimiter) checkTokenBucket(key string, tier *TierConfig, now time.Time) Decision {
	rl.mu.Lock()
	bucket, ok := rl.tokenBuckets[key]
	if !ok {
		bucket = NewTokenBucket(tier.MaxRequests, tier.RefillRate)
		rl.tokenBuckets[key] = bucket
	}
	rl.mu.Unlock()
	return bucket.Allow(now)
}

func (rl *RateLimiter) checkSlidingWindow(key string, tier *TierConfig, now time.Time) Decision {
	rl.mu.Lock()
	window, ok := rl.slidingWindows[key]
	if !ok {
		window = NewSlidingWindow(tier.MaxRequests, tier.WindowDuration)
		rl.slidingWindows[key] = window
	}
	rl.mu.Unlock()
	return window.Allow(now)
}

// Reset clears the rate limit state for a specific key.
func (rl *RateLimiter) Reset(tierType TierType, identifier string) {
	key := buildKey(tierType, identifier)
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.tokenBuckets, key)
	delete(rl.slidingWindows, key)
}

// Stats returns current statistics.
type Stats struct {
	ActiveBuckets int
	ActiveWindows int
	TierCount     int
}

// GetStats returns current rate limiter statistics.
func (rl *RateLimiter) GetStats() Stats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()
	return Stats{
		ActiveBuckets: len(rl.tokenBuckets),
		ActiveWindows: len(rl.slidingWindows),
		TierCount:     len(rl.tiers),
	}
}
