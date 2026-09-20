package cache

import (
	"context"
	"sort"
	"sync"
	"time"
)

// SlidingWindowRateLimiterLuaScript is the production Redis Lua script for atomic sliding window evaluation.
const SlidingWindowRateLimiterLuaScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local clearBefore = now - window

-- Remove timestamps outside the sliding window
redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)

-- Count current requests in window
local currentRequests = redis.call('ZCARD', key)

if currentRequests < limit then
    -- Add current request timestamp
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window)
    return {1, limit - currentRequests - 1} -- Allowed (1), remaining capacity
else
    return {0, 0} -- Blocked (0), 0 capacity
end
`

// RateLimitResult contains rate limiter evaluation outcome.
type RateLimitResult struct {
	Allowed   bool
	Remaining int64
	ResetIn   time.Duration
}

// SlidingWindowRateLimiter evaluates limits using sorted sets.
type SlidingWindowRateLimiter struct {
	mu      sync.Mutex
	windows map[string][]int64 // key -> list of request timestamps (in milliseconds)
}

func NewSlidingWindowRateLimiter() *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		windows: make(map[string][]int64),
	}
}

// Allow evaluates an incoming request against the sliding window limit.
func (r *SlidingWindowRateLimiter) Allow(
	ctx context.Context,
	key string,
	limit int64,
	window time.Duration,
) (RateLimitResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	nowMs := time.Now().UnixMilli()
	windowMs := window.Milliseconds()
	clearBefore := nowMs - windowMs

	timestamps := r.windows[key]

	// 1. ZREMRANGEBYSCORE: Filter out old timestamps
	valid := make([]int64, 0, len(timestamps))
	for _, ts := range timestamps {
		if ts > clearBefore {
			valid = append(valid, ts)
		}
	}

	currentRequests := int64(len(valid))

	if currentRequests < limit {
		// Allowed: add current timestamp
		valid = append(valid, nowMs)
		sort.Slice(valid, func(i, j int) bool { return valid[i] < valid[j] })
		r.windows[key] = valid

		remaining := limit - currentRequests - 1
		return RateLimitResult{
			Allowed:   true,
			Remaining: remaining,
			ResetIn:   window,
		}, nil
	}

	// Blocked
	r.windows[key] = valid
	var resetIn time.Duration = window
	if len(valid) > 0 {
		oldest := valid[0]
		remainingMs := (oldest + windowMs) - nowMs
		if remainingMs > 0 {
			resetIn = time.Duration(remainingMs) * time.Millisecond
		}
	}

	return RateLimitResult{
		Allowed:   false,
		Remaining: 0,
		ResetIn:   resetIn,
	}, ErrRateLimitExceeded
}

// Reset clears the sliding window for a given key.
func (r *SlidingWindowRateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.windows, key)
}
