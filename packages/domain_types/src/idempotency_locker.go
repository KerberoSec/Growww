package src

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// IdempotencyStatus tracks the execution phase of an idempotent request.
type IdempotencyStatus string

const (
	IdempotencyInFlight  IdempotencyStatus = "IN_FLIGHT"
	IdempotencyCommitted IdempotencyStatus = "COMMITTED"
	IdempotencyRejected  IdempotencyStatus = "REJECTED"
)

var (
	ErrLockConflict  = errors.New("idempotency: concurrent request in flight with same idempotency key")
	ErrLockExpired   = errors.New("idempotency: lease expired")
	ErrAlreadyLocked = errors.New("idempotency: key already committed")
)

// IdempotencyRecord stores cached execution results and lease status.
type IdempotencyRecord struct {
	Key           string            `json:"key"`
	LeaseToken    string            `json:"lease_token"`
	Status        IdempotencyStatus `json:"status"`
	ResponseBody  []byte            `json:"response_body"`
	ResponseCode  int               `json:"response_code"`
	CreatedAt     time.Time         `json:"created_at"`
	ExpiresAt     time.Time         `json:"expires_at"`
	ExecutionMs   int64             `json:"execution_ms"`
}

// IdempotencyLocker defines distributed idempotency lock management contract.
type IdempotencyLocker interface {
	AcquireLease(key string, leaseToken string, ttl time.Duration) (*IdempotencyRecord, error)
	CommitExecution(key string, leaseToken string, responseCode int, responseBody []byte, retention time.Duration) error
	RejectExecution(key string, leaseToken string, reason string) error
	GetRecord(key string) (*IdempotencyRecord, error)
}

// MemoryIdempotencyLocker provides an atomic in-memory implementation matching Redis SETNX and Redlock behavior.
type MemoryIdempotencyLocker struct {
	mu      sync.Mutex
	records map[string]*IdempotencyRecord
}

// NewMemoryIdempotencyLocker creates a new thread-safe idempotency manager.
func NewMemoryIdempotencyLocker() *MemoryIdempotencyLocker {
	return &MemoryIdempotencyLocker{
		records: make(map[string]*IdempotencyRecord),
	}
}

// AcquireLease attempts to acquire an in-flight lease for a key.
// Returns existing committed record if already successfully processed, or ErrLockConflict if in flight.
func (m *MemoryIdempotencyLocker) AcquireLease(key string, leaseToken string, ttl time.Duration) (*IdempotencyRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()

	// Check if key exists
	if record, exists := m.records[key]; exists {
		// If expired, permit takeover
		if now.After(record.ExpiresAt) && record.Status == IdempotencyInFlight {
			delete(m.records, key)
		} else {
			switch record.Status {
			case IdempotencyCommitted:
				// Return existing committed record for zero-duplicate replay
				return record, nil
			case IdempotencyInFlight:
				// Active concurrent execution
				return nil, fmt.Errorf("%w: key=%s lease=%s", ErrLockConflict, key, record.LeaseToken)
			case IdempotencyRejected:
				// Can retry if previous was rejected
				delete(m.records, key)
			}
		}
	}

	// Create new in-flight lease
	record := &IdempotencyRecord{
		Key:        key,
		LeaseToken: leaseToken,
		Status:     IdempotencyInFlight,
		CreatedAt:  now,
		ExpiresAt:  now.Add(ttl),
	}
	m.records[key] = record
	return record, nil
}

// CommitExecution stores the final response payload and updates state to COMMITTED.
func (m *MemoryIdempotencyLocker) CommitExecution(key string, leaseToken string, responseCode int, responseBody []byte, retention time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.records[key]
	if !exists {
		return errors.New("idempotency record not found")
	}

	if record.LeaseToken != leaseToken {
		return errors.New("lease token mismatch during commit")
	}

	record.Status = IdempotencyCommitted
	record.ResponseCode = responseCode
	record.ResponseBody = responseBody
	record.ExpiresAt = time.Now().UTC().Add(retention)
	return nil
}

// RejectExecution marks the execution as REJECTED or deletes the lease to permit immediate retry.
func (m *MemoryIdempotencyLocker) RejectExecution(key string, leaseToken string, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.records[key]
	if !exists {
		return errors.New("idempotency record not found")
	}

	if record.LeaseToken != leaseToken {
		return errors.New("lease token mismatch during reject")
	}

	record.Status = IdempotencyRejected
	return nil
}

// GetRecord returns the stored record if available.
func (m *MemoryIdempotencyLocker) GetRecord(key string) (*IdempotencyRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, exists := m.records[key]
	if !exists {
		return nil, errors.New("record not found")
	}
	return record, nil
}
