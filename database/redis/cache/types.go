package cache

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrLockAcquisitionFailed = errors.New("failed to acquire distributed lock")
	ErrLockNotHeld           = errors.New("lock is not held or fencing token mismatch")
	ErrRateLimitExceeded     = errors.New("rate limit exceeded for sliding window")
	ErrCacheMiss             = errors.New("key not found in cache")
)

// Standard Redis Key Namespace Prefix: growww:<env>:<domain>:<entity>:<id>
func FormatKey(env, domain, entity, id string) string {
	return fmt.Sprintf("growww:%s:%s:%s:%s", env, domain, entity, id)
}

// InvalidationMessage is broadcast via Redis Pub/Sub to invalidate L1 caches.
type InvalidationMessage struct {
	Key       string    `json:"key"`
	NodeID    string    `json:"node_id"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderBookLevel represents a single price level in the L2 depth cache.
type OrderBookLevel struct {
	Price    int64 `json:"price"`    // Scaled by 10,000 (e.g. 25000000 = 2500.0000 INR)
	Quantity int64 `json:"quantity"` // Scaled by 1,000,000 (6 decimals)
	Orders   int   `json:"orders"`   // Count of resting orders at this price
}

// OrderBookDepth represents top bids and asks snapshot.
type OrderBookDepth struct {
	Symbol    string           `json:"symbol"`
	Bids      []OrderBookLevel `json:"bids"` // Sorted descending by price
	Asks      []OrderBookLevel `json:"asks"` // Sorted ascending by price
	UpdatedAt time.Time        `json:"updated_at"`
}

// SessionRecord holds user session state.
type SessionRecord struct {
	UserID       string    `json:"user_id"`
	DeviceID     string    `json:"device_id"`
	JWTID        string    `json:"jwt_jti"`
	Role         string    `json:"role"`
	EntityType   string    `json:"entity_type"`
	LastActiveAt time.Time `json:"last_active_at"`
	IsRevoked    bool      `json:"is_revoked"`
}

// ComplianceStatus represents whitelist state.
type ComplianceStatus string

const (
	ComplianceStatusActive    ComplianceStatus = "ACTIVE"
	ComplianceStatusFrozen    ComplianceStatus = "FROZEN"
	ComplianceStatusSuspended ComplianceStatus = "SUSPENDED"
)
