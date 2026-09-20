package src

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type KeyLifecycleState string

const (
	KeyStateGenerated KeyLifecycleState = "GENERATED"
	KeyStateActive    KeyLifecycleState = "ACTIVE"
	KeyStateRotating  KeyLifecycleState = "ROTATING"
	KeyStateArchived  KeyLifecycleState = "ARCHIVED"
	KeyStateRevoked   KeyLifecycleState = "REVOKED_SHREDDED"
)

var (
	ErrQuorumNotMet           = errors.New("m-of-n ceremony quorum threshold not met")
	ErrInvalidStateTransition = errors.New("illegal key lifecycle state transition")
)

// ManagedKeyRecord tracks operational key status and rotation metadata
type ManagedKeyRecord struct {
	KeyAlias      string
	State         KeyLifecycleState
	CreatedAt     time.Time
	ActivatedAt   time.Time
	RotatedAt     time.Time
	RevokedAt     time.Time
	QuorumSigners []string
}

// KeyCeremonyCoordinator manages M-of-N key rotation and shredding
type KeyCeremonyCoordinator struct {
	mu            sync.RWMutex
	pool          *PKCS11SessionPool
	thresholdM    int
	totalN        int
	keys          map[string]*ManagedKeyRecord
	officerRoster map[string]bool
}

func NewKeyCeremonyCoordinator(pool *PKCS11SessionPool, thresholdM, totalN int, officers []string) *KeyCeremonyCoordinator {
	roster := make(map[string]bool)
	for _, off := range officers {
		roster[off] = true
	}

	return &KeyCeremonyCoordinator{
		pool:          pool,
		thresholdM:    thresholdM,
		totalN:        totalN,
		keys:          make(map[string]*ManagedKeyRecord),
		officerRoster: roster,
	}
}

// ExecuteKeyGenerationCeremony generates and activates a new hardware key under M-of-N authorization
func (c *KeyCeremonyCoordinator) ExecuteKeyGenerationCeremony(
	keyAlias string,
	algo KeyAlgorithm,
	approvingOfficers []string,
) (*ManagedKeyRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Verify M-of-N quorum
	validSigners := 0
	signerSet := make(map[string]bool)
	for _, off := range approvingOfficers {
		if c.officerRoster[off] && !signerSet[off] {
			signerSet[off] = true
			validSigners++
		}
	}

	if validSigners < c.thresholdM {
		return nil, fmt.Errorf("%w: required %d, valid %d", ErrQuorumNotMet, c.thresholdM, validSigners)
	}

	// Generate key inside HSM
	_, err := c.pool.GenerateHardwareKey(keyAlias, algo)
	if err != nil {
		return nil, err
	}

	record := &ManagedKeyRecord{
		KeyAlias:      keyAlias,
		State:         KeyStateActive,
		CreatedAt:     time.Now().UTC(),
		ActivatedAt:   time.Now().UTC(),
		QuorumSigners: approvingOfficers,
	}

	c.keys[keyAlias] = record
	return record, nil
}

// RotateKey transitions current key to archived and promotes replacement key
func (c *KeyCeremonyCoordinator) RotateKey(currentAlias, newAlias string, approvingOfficers []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(approvingOfficers) < c.thresholdM {
		return ErrQuorumNotMet
	}

	curr, existsCurr := c.keys[currentAlias]
	next, existsNext := c.keys[newAlias]
	if !existsCurr || !existsNext {
		return ErrKeyNotFound
	}

	if curr.State != KeyStateActive {
		return fmt.Errorf("%w: current key is not in ACTIVE state", ErrInvalidStateTransition)
	}

	curr.State = KeyStateArchived
	curr.RotatedAt = time.Now().UTC()

	next.State = KeyStateActive
	next.ActivatedAt = time.Now().UTC()

	return nil
}

// ShredKey permanently invalidates and shreds a revoked key
func (c *KeyCeremonyCoordinator) ShredKey(keyAlias string, approvingOfficers []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(approvingOfficers) < c.thresholdM {
		return ErrQuorumNotMet
	}

	keyRecord, exists := c.keys[keyAlias]
	if !exists {
		return ErrKeyNotFound
	}

	keyRecord.State = KeyStateRevoked
	keyRecord.RevokedAt = time.Now().UTC()

	// Invalidate key descriptor in session pool
	c.pool.mu.Lock()
	delete(c.pool.keys, keyAlias)
	c.pool.mu.Unlock()

	return nil
}

// GetKeyRecord returns key state record
func (c *KeyCeremonyCoordinator) GetKeyRecord(keyAlias string) (*ManagedKeyRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rec, exists := c.keys[keyAlias]
	if !exists {
		return nil, ErrKeyNotFound
	}
	return rec, nil
}
