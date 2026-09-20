package cache

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// SafeReleaseLockLuaScript ensures only the lock owner holding the fencing token can release the lock.
const SafeReleaseLockLuaScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`

// LeaseExtensionLuaScript extends the lock lease only if caller still holds it.
const LeaseExtensionLuaScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("pexpire", KEYS[1], ARGV[2])
else
    return 0
end
`

// LockHandle represents an acquired lock with a monotonic fencing token.
type LockHandle struct {
	Resource     string
	FencingToken string
	FencingSeq   int64
	TTL          time.Duration
	AcquiredAt   time.Time
	cancelRenew  context.CancelFunc
}

// DistributedLockManager coordinates distributed mutexes with fencing tokens and lease renewal.
type DistributedLockManager struct {
	mu           sync.Mutex
	locks        map[string]*LockHandle
	globalSeq    int64
	defaultTTL   time.Duration
}

func NewDistributedLockManager(defaultTTL time.Duration) *DistributedLockManager {
	if defaultTTL <= 0 {
		defaultTTL = 5 * time.Second
	}
	return &DistributedLockManager{
		locks:      make(map[string]*LockHandle),
		defaultTTL: defaultTTL,
	}
}

func generateToken() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Acquire acquires a lock on a resource with a timeout.
func (m *DistributedLockManager) Acquire(ctx context.Context, resource string, ttl time.Duration) (*LockHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ttl <= 0 {
		ttl = m.defaultTTL
	}

	now := time.Now()
	existing, held := m.locks[resource]
	if held {
		// Check if expired
		if now.Sub(existing.AcquiredAt) < existing.TTL {
			return nil, ErrLockAcquisitionFailed
		}
		// Expired: allow overwrite
	}

	seq := atomic.AddInt64(&m.globalSeq, 1)
	token := fmt.Sprintf("%s-%d", generateToken(), seq)

	handle := &LockHandle{
		Resource:     resource,
		FencingToken: token,
		FencingSeq:   seq,
		TTL:          ttl,
		AcquiredAt:   now,
	}

	m.locks[resource] = handle
	return handle, nil
}

// Release safely releases the lock, validating fencing token.
func (m *DistributedLockManager) Release(ctx context.Context, handle *LockHandle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if handle.cancelRenew != nil {
		handle.cancelRenew()
	}

	existing, held := m.locks[handle.Resource]
	if !held {
		return ErrLockNotHeld
	}

	if existing.FencingToken != handle.FencingToken {
		return ErrLockNotHeld
	}

	delete(m.locks, handle.Resource)
	return nil
}

// ExtendLease extends the TTL of an actively held lock.
func (m *DistributedLockManager) ExtendLease(ctx context.Context, handle *LockHandle, extraTTL time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, held := m.locks[handle.Resource]
	if !held || existing.FencingToken != handle.FencingToken {
		return ErrLockNotHeld
	}

	existing.TTL += extraTTL
	handle.TTL = existing.TTL
	return nil
}

// StartAutoRenewal starts a background goroutine to renew the lock until context cancellation.
func (m *DistributedLockManager) StartAutoRenewal(ctx context.Context, handle *LockHandle, interval time.Duration) {
	renewCtx, cancel := context.WithCancel(ctx)
	handle.cancelRenew = cancel

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-renewCtx.Done():
				return
			case <-ticker.C:
				if err := m.ExtendLease(renewCtx, handle, handle.TTL); err != nil {
					return
				}
			}
		}
	}()
}
