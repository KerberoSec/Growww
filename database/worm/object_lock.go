package worm

import (
	"errors"
	"time"
)

var (
	ErrObjectLockedInCompliance = errors.New("WORM_VIOLATION: object is locked in COMPLIANCE mode and cannot be deleted or modified")
	ErrLegalHoldActive          = errors.New("WORM_VIOLATION: object is under active Legal Hold and cannot be deleted")
	ErrRetentionShortening      = errors.New("WORM_VIOLATION: retention period cannot be shortened")
	ErrGovernanceBypassDenied   = errors.New("WORM_VIOLATION: privileged governance bypass token required to delete object in GOVERNANCE mode")
	ErrObjectNotFound           = errors.New("object not found in WORM storage")
	ErrIntegrityMismatch        = errors.New("data integrity error: SHA-256 checksum does not match payload")
	ErrInvalidRetentionDate     = errors.New("retain-until-date must be in the future")
)

// ObjectLockMode defines the immutability enforcement level.
type ObjectLockMode string

const (
	ModeCompliance ObjectLockMode = "COMPLIANCE" // No user, not even root/admin, can delete before expiration
	ModeGovernance ObjectLockMode = "GOVERNANCE" // Privileged users with bypass token can modify
)

// ObjectRecord represents a versioned, immutable object in the WORM storage engine.
type ObjectRecord struct {
	Bucket           string         `json:"bucket"`
	Key              string         `json:"key"`
	VersionID        string         `json:"version_id"`
	Payload          []byte         `json:"-"`
	SHA256Checksum   string         `json:"sha256_checksum"`
	ContentLength    int64          `json:"content_length"`
	RetentionMode    ObjectLockMode `json:"retention_mode"`
	RetainUntilDate  time.Time      `json:"retain_until_date"`
	LegalHold        bool           `json:"legal_hold"`
	MerkleRoot       string         `json:"merkle_root,omitempty"`
	BesuTxHash       string         `json:"besu_tx_hash,omitempty"`
	BesuBlockNumber  uint64         `json:"besu_block_number,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	IsLatest         bool           `json:"is_latest"`
}

// PutObjectRequest defines parameters for writing to WORM storage.
type PutObjectRequest struct {
	Bucket          string         `json:"bucket"`
	Key             string         `json:"key"`
	Payload         []byte         `json:"payload"`
	ExpectedSHA256  string         `json:"expected_sha256,omitempty"`
	RetentionMode   ObjectLockMode `json:"retention_mode"`
	RetainUntilDate time.Time      `json:"retain_until_date"`
	LegalHold       bool           `json:"legal_hold"`
}

// PutObjectResponse returns write confirmation.
type PutObjectResponse struct {
	Bucket         string    `json:"bucket"`
	Key            string    `json:"key"`
	VersionID      string    `json:"version_id"`
	SHA256Checksum string    `json:"sha256_checksum"`
	CreatedAt      time.Time `json:"created_at"`
}
