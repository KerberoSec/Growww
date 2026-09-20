package worm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// WORMStorageEngine manages immutable S3 Object Lock compliance storage.
type WORMStorageEngine struct {
	mu           sync.RWMutex
	objects      map[string][]*ObjectRecord // key: bucket/key -> slice of versions
	anchorer     *BesuAnchorer
	defaultRetention time.Duration
}

// NewWORMStorageEngine creates a new WORM compliance storage engine.
func NewWORMStorageEngine() *WORMStorageEngine {
	return &WORMStorageEngine{
		objects:          make(map[string][]*ObjectRecord),
		anchorer:         NewBesuAnchorer(),
		defaultRetention: 7 * 365 * 24 * time.Hour, // 7-year regulatory retention
	}
}

// PutObject stores a new object version with immutable Object Lock compliance.
func (e *WORMStorageEngine) PutObject(ctx context.Context, req PutObjectRequest) (*PutObjectResponse, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if req.Bucket == "" || req.Key == "" {
		return nil, fmt.Errorf("bucket and key cannot be empty")
	}

	// Calculate SHA-256
	h := sha256.Sum256(req.Payload)
	actualSHA256 := hex.EncodeToString(h[:])

	if req.ExpectedSHA256 != "" && req.ExpectedSHA256 != actualSHA256 {
		return nil, ErrIntegrityMismatch
	}

	// Set default retention if unspecified
	now := time.Now()
	retainUntil := req.RetainUntilDate
	if retainUntil.IsZero() {
		retainUntil = now.Add(e.defaultRetention)
	}

	if retainUntil.Before(now) {
		return nil, ErrInvalidRetentionDate
	}

	mode := req.RetentionMode
	if mode == "" {
		mode = ModeCompliance
	}

	versionID := fmt.Sprintf("v_%d_%s", now.UnixNano(), actualSHA256[:8])
	record := &ObjectRecord{
		Bucket:          req.Bucket,
		Key:             req.Key,
		VersionID:       versionID,
		Payload:         req.Payload,
		SHA256Checksum:  actualSHA256,
		ContentLength:   int64(len(req.Payload)),
		RetentionMode:   mode,
		RetainUntilDate: retainUntil,
		LegalHold:       req.LegalHold,
		CreatedAt:       now,
		IsLatest:        true,
	}

	mapURI := fmt.Sprintf("%s/%s", req.Bucket, req.Key)
	versions := e.objects[mapURI]
	for _, v := range versions {
		v.IsLatest = false
	}
	e.objects[mapURI] = append(versions, record)

	return &PutObjectResponse{
		Bucket:         req.Bucket,
		Key:            req.Key,
		VersionID:      versionID,
		SHA256Checksum: actualSHA256,
		CreatedAt:      now,
	}, nil
}

// GetObject retrieves an object and verifies its cryptographic integrity.
func (e *WORMStorageEngine) GetObject(ctx context.Context, bucket, key, versionID string) (*ObjectRecord, []byte, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	mapURI := fmt.Sprintf("%s/%s", bucket, key)
	versions, exists := e.objects[mapURI]
	if !exists || len(versions) == 0 {
		return nil, nil, ErrObjectNotFound
	}

	var target *ObjectRecord
	if versionID == "" {
		// Get latest
		for i := len(versions) - 1; i >= 0; i-- {
			if versions[i].IsLatest {
				target = versions[i]
				break
			}
		}
		if target == nil {
			target = versions[len(versions)-1]
		}
	} else {
		for _, v := range versions {
			if v.VersionID == versionID {
				target = v
				break
			}
		}
	}

	if target == nil {
		return nil, nil, ErrObjectNotFound
	}

	// Verify cryptographic integrity
	h := sha256.Sum256(target.Payload)
	calculatedSHA256 := hex.EncodeToString(h[:])
	if calculatedSHA256 != target.SHA256Checksum {
		return nil, nil, ErrIntegrityMismatch
	}

	return target, target.Payload, nil
}

// DeleteObject handles object deletion requests with strict WORM compliance.
func (e *WORMStorageEngine) DeleteObject(ctx context.Context, bucket, key, versionID string, bypassGovernance bool, governanceToken string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	mapURI := fmt.Sprintf("%s/%s", bucket, key)
	versions, exists := e.objects[mapURI]
	if !exists || len(versions) == 0 {
		return ErrObjectNotFound
	}

	var targetIndex = -1
	for i, v := range versions {
		if v.VersionID == versionID {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return ErrObjectNotFound
	}

	target := versions[targetIndex]
	now := time.Now()

	// 1. Legal Hold check
	if target.LegalHold {
		return ErrLegalHoldActive
	}

	// 2. Retention Mode check
	if target.RetentionMode == ModeCompliance {
		if now.Before(target.RetainUntilDate) {
			return ErrObjectLockedInCompliance
		}
	} else if target.RetentionMode == ModeGovernance {
		if now.Before(target.RetainUntilDate) {
			if !bypassGovernance || governanceToken != "GOV_BYPASS_AUTHORIZED_SEC_RULE_17A4" {
				return ErrGovernanceBypassDenied
			}
		}
	}

	// Deletion allowed after retention expiry or with valid governance bypass
	e.objects[mapURI] = append(versions[:targetIndex], versions[targetIndex+1:]...)
	return nil
}

// ExtendRetention extends the retention period. Shortening is strictly forbidden.
func (e *WORMStorageEngine) ExtendRetention(ctx context.Context, bucket, key, versionID string, newRetainUntil time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	mapURI := fmt.Sprintf("%s/%s", bucket, key)
	versions, exists := e.objects[mapURI]
	if !exists {
		return ErrObjectNotFound
	}

	for _, v := range versions {
		if v.VersionID == versionID {
			if newRetainUntil.Before(v.RetainUntilDate) {
				return ErrRetentionShortening
			}
			v.RetainUntilDate = newRetainUntil
			return nil
		}
	}

	return ErrObjectNotFound
}

// SetLegalHold sets or removes a legal hold.
func (e *WORMStorageEngine) SetLegalHold(ctx context.Context, bucket, key, versionID string, hold bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	mapURI := fmt.Sprintf("%s/%s", bucket, key)
	versions, exists := e.objects[mapURI]
	if !exists {
		return ErrObjectNotFound
	}

	for _, v := range versions {
		if v.VersionID == versionID {
			v.LegalHold = hold
			return nil
		}
	}

	return ErrObjectNotFound
}

// SealBatchAndAnchor constructs a Merkle tree of all unanchored objects and anchors it to Hyperledger Besu.
func (e *WORMStorageEngine) SealBatchAndAnchor(ctx context.Context, bucket string) (*BesuAnchorReceipt, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var hashes [][]byte
	var pendingRecords []*ObjectRecord

	for mapURI, versions := range e.objects {
		_ = mapURI
		for _, v := range versions {
			if v.Bucket == bucket && v.BesuTxHash == "" {
				checksumBytes, err := hex.DecodeString(v.SHA256Checksum)
				if err == nil {
					hashes = append(hashes, checksumBytes)
					pendingRecords = append(pendingRecords, v)
				}
			}
		}
	}

	if len(hashes) == 0 {
		return nil, fmt.Errorf("no unanchored objects found in bucket %s", bucket)
	}

	tree, err := NewMerkleTree(hashes)
	if err != nil {
		return nil, err
	}

	rootHex := tree.RootHex()
	receipt, err := e.anchorer.AnchorMerkleRoot(rootHex)
	if err != nil {
		return nil, err
	}

	for _, rec := range pendingRecords {
		rec.MerkleRoot = rootHex
		rec.BesuTxHash = receipt.TransactionHash
		rec.BesuBlockNumber = receipt.BlockNumber
	}

	return receipt, nil
}
