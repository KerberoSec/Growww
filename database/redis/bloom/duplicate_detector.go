package bloom

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrDuplicateTransaction = errors.New("REPLAY_ATTACK_DETECTED: duplicate transaction / order ID detected by Bloom filter")
	ErrNonMonotonicNonce    = errors.New("SECURITY_VIOLATION: nonce is not monotonically increasing")
)

// DeduplicationMetrics tracks deduplication throughput.
type DeduplicationMetrics struct {
	TotalChecked        uint64
	UniqueAccepted      uint64
	DuplicatesBlocked   uint64
	ReplaysBlocked      uint64
	EstimatedMemoryKB   uint64
}

// SecurityAuditRecord logs detected replay / duplicate attacks.
type SecurityAuditRecord struct {
	EntityID    string    `json:"entity_id"`
	PayloadHash string    `json:"payload_hash"`
	DetectedAt  time.Time `json:"detected_at"`
	Reason      string    `json:"reason"`
}

// DuplicateDetectorEngine provides sub-millisecond duplicate detection across order ingestion,
// transaction hash validation, and nonce monotonicity checks.
type DuplicateDetectorEngine struct {
	mu           sync.RWMutex
	filter       *RotationalSlidingWindowBloomFilter
	nonces       map[string]uint64 // entityID -> last monotonic nonce
	auditLog     []SecurityAuditRecord
	metrics      DeduplicationMetrics
}

// NewDuplicateDetectorEngine initializes the engine with a 15-minute sliding window across 5 buckets.
func NewDuplicateDetectorEngine() *DuplicateDetectorEngine {
	return &DuplicateDetectorEngine{
		filter:   NewRotationalSlidingWindowBloomFilter(15*time.Minute, 5, 500000, 0.0001), // 0.01% false positive
		nonces:   make(map[string]uint64),
		auditLog: make([]SecurityAuditRecord, 0),
	}
}

// CheckAndRecordTransaction verifies if a transaction hash or order ID has already been seen.
// Rejects duplicates with ErrDuplicateTransaction in <100 microseconds.
func (d *DuplicateDetectorEngine) CheckAndRecordTransaction(txHash string, entityID string, nonce uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.metrics.TotalChecked++
	now := time.Now()

	// 1. Nonce Monotonicity Check
	if nonce > 0 {
		lastNonce, hasNonce := d.nonces[entityID]
		if hasNonce && nonce <= lastNonce {
			d.metrics.ReplaysBlocked++
			d.auditLog = append(d.auditLog, SecurityAuditRecord{
				EntityID:    entityID,
				PayloadHash: txHash,
				DetectedAt:  now,
				Reason:      fmt.Sprintf("Non-monotonic nonce %d <= last %d", nonce, lastNonce),
			})
			return fmt.Errorf("%w: entity=%s current=%d last=%d", ErrNonMonotonicNonce, entityID, nonce, lastNonce)
		}
	}

	// 2. Bloom Filter Replay / Duplicate Check
	h := sha256.Sum256([]byte(txHash))
	hashBytes := h[:]

	isDuplicate := d.filter.TestAndSet(hashBytes, now)
	if isDuplicate {
		d.metrics.DuplicatesBlocked++
		d.auditLog = append(d.auditLog, SecurityAuditRecord{
			EntityID:    entityID,
			PayloadHash: hex.EncodeToString(hashBytes),
			DetectedAt:  now,
			Reason:      "Duplicate transaction hash found in Bloom filter",
		})
		return fmt.Errorf("%w: txHash=%s", ErrDuplicateTransaction, txHash)
	}

	// Record accepted transaction & monotonic nonce
	if nonce > 0 {
		d.nonces[entityID] = nonce
	}
	d.metrics.UniqueAccepted++

	return nil
}

// GetMetrics returns engine metrics.
func (d *DuplicateDetectorEngine) GetMetrics() DeduplicationMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.metrics
}

// GetAuditLog returns security audit log records.
func (d *DuplicateDetectorEngine) GetAuditLog() []SecurityAuditRecord {
	d.mu.RLock()
	defer d.mu.RUnlock()
	res := make([]SecurityAuditRecord, len(d.auditLog))
	copy(res, d.auditLog)
	return res
}
