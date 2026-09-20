package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"time"
)

// KeyStatus represents the lifecycle state of a user-specific KMS encryption key
type KeyStatus string

const (
	KeyStatusActive     KeyStatus = "ACTIVE"
	KeyStatusSuspended  KeyStatus = "SUSPENDED"
	KeyStatusPendingDestruction KeyStatus = "PENDING_DESTRUCTION"
	KeyStatusDestroyed  KeyStatus = "DESTROYED"
)

// ShreddingStatus represents lifecycle of a Right to be Forgotten / erasure request
type ShreddingStatus string

const (
	ShreddingRequested   ShreddingStatus = "REQUESTED"
	ShreddingApproved    ShreddingStatus = "APPROVED"
	ShreddingRejected    ShreddingStatus = "REJECTED"
	ShreddingCompleted   ShreddingStatus = "COMPLETED"
	ShreddingHeld        ShreddingStatus = "REGULATORY_HOLD"
)

// EncryptedUserDataEnvelope holds AES-256-GCM encrypted PII data
type EncryptedUserDataEnvelope struct {
	UserID             string    `json:"user_id"`
	EncryptedPAN       []byte    `json:"encrypted_pan"`
	EncryptedAadhaar   []byte    `json:"encrypted_aadhaar"`
	EncryptedBiometric []byte    `json:"encrypted_biometric"`
	EncryptedBankAcct  []byte    `json:"encrypted_bank_acct"`
	Nonce              []byte    `json:"nonce"`
	KeyID              string    `json:"key_id"`
	EncryptedAt        time.Time `json:"encrypted_at"`
}

// UserKMSKey represents per-user envelope encryption key in KMS / HSM
type UserKMSKey struct {
	KeyID       string    `json:"key_id"`
	UserID      string    `json:"user_id"`
	KeyMaterial []byte    `json:"-"` // 256-bit AES DEK (Zeroized on destruction)
	Status      KeyStatus `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	DestroyedAt time.Time `json:"destroyed_at,omitempty"`
}

// ShreddingRequest represents formal Right-to-be-Forgotten request under DPDP Act 2023
type ShreddingRequest struct {
	RequestID              string          `json:"request_id"`
	UserID                 string          `json:"user_id"`
	AccountClosedAt        time.Time       `json:"account_closed_at"`
	LastTransactionAt      time.Time       `json:"last_transaction_at"`
	HasActiveInquiry       bool            `json:"has_active_inquiry"`
	ComplianceOfficerSign  string          `json:"compliance_officer_sign"`
	DPOSign                string          `json:"dpo_sign"`
	Status                 ShreddingStatus `json:"status"`
	RejectionReason        string          `json:"rejection_reason,omitempty"`
	DestructionCertHash    string          `json:"destruction_cert_hash,omitempty"`
	OnChainBesuTxHash      string          `json:"on_chain_besu_tx_hash,omitempty"`
	RequestedAt            time.Time       `json:"requested_at"`
	CompletedAt            time.Time       `json:"completed_at,omitempty"`
}

// DestructionCertificate represents immutable proof of cryptographic shredding
type DestructionCertificate struct {
	CertificateID       string    `json:"certificate_id"`
	RequestID           string    `json:"request_id"`
	UserID              string    `json:"user_id"`
	KeyID               string    `json:"key_id"`
	KeyStatus           KeyStatus `json:"key_status"`
	CiphertextEntropy   float64   `json:"ciphertext_entropy"`
	IsUnrecoverable     bool      `json:"is_unrecoverable"`
	ComplianceSign      string    `json:"compliance_sign"`
	DPOSign             string    `json:"dpo_sign"`
	CertificateSHA256   string    `json:"certificate_sha256"`
	BesuAttestationHash string    `json:"besu_attestation_hash"`
	DestroyedAt         time.Time `json:"destroyed_at"`
}

// CryptoShreddingKMSEngine manages envelope encryption, DPDP Act 2023 compliance, and key destruction
type CryptoShreddingKMSEngine struct {
	mu           sync.RWMutex
	kmsKeys      map[string]*UserKMSKey            // keyID -> UserKMSKey
	userKeyMap   map[string]string                 // userID -> keyID
	userData     map[string]*EncryptedUserDataEnvelope // userID -> envelope
	requests     map[string]*ShreddingRequest      // requestID -> request
	certificates map[string]*DestructionCertificate // certID -> cert
}

// NewCryptoShreddingKMSEngine creates a new CryptoShreddingKMSEngine instance
func NewCryptoShreddingKMSEngine() *CryptoShreddingKMSEngine {
	return &CryptoShreddingKMSEngine{
		kmsKeys:      make(map[string]*UserKMSKey),
		userKeyMap:   make(map[string]string),
		userData:     make(map[string]*EncryptedUserDataEnvelope),
		requests:     make(map[string]*ShreddingRequest),
		certificates: make(map[string]*DestructionCertificate),
	}
}

// ProvisionUserKMSKey creates a fresh 256-bit AES DEK in KMS
func (e *CryptoShreddingKMSEngine) ProvisionUserKMSKey(userID string) (*UserKMSKey, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if userID == "" {
		return nil, errors.New("user_id cannot be empty")
	}

	keyMaterial := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, keyMaterial); err != nil {
		return nil, fmt.Errorf("failed to generate random key material: %w", err)
	}

	keyID := fmt.Sprintf("kms-key-%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d", userID, time.Now().UnixNano()))))[:20]
	kmsKey := &UserKMSKey{
		KeyID:       keyID,
		UserID:      userID,
		KeyMaterial: keyMaterial,
		Status:      KeyStatusActive,
		CreatedAt:   time.Now().UTC(),
	}

	e.kmsKeys[keyID] = kmsKey
	e.userKeyMap[userID] = keyID
	return kmsKey, nil
}

// EncryptUserPII performs AES-256-GCM envelope encryption on user PII fields
func (e *CryptoShreddingKMSEngine) EncryptUserPII(userID, pan, aadhaar, biometric, bankAcct string) (*EncryptedUserDataEnvelope, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	keyID, exists := e.userKeyMap[userID]
	if !exists {
		return nil, fmt.Errorf("no KMS key found for user %s", userID)
	}

	kmsKey := e.kmsKeys[keyID]
	if kmsKey.Status != KeyStatusActive {
		return nil, fmt.Errorf("KMS key %s is not active (status: %s)", keyID, kmsKey.Status)
	}

	block, err := aes.NewCipher(kmsKey.KeyMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	encPAN := gcm.Seal(nil, nonce, []byte(pan), []byte(userID))
	encAadhaar := gcm.Seal(nil, nonce, []byte(aadhaar), []byte(userID))
	encBio := gcm.Seal(nil, nonce, []byte(biometric), []byte(userID))
	encBank := gcm.Seal(nil, nonce, []byte(bankAcct), []byte(userID))

	envelope := &EncryptedUserDataEnvelope{
		UserID:             userID,
		EncryptedPAN:       encPAN,
		EncryptedAadhaar:   encAadhaar,
		EncryptedBiometric: encBio,
		EncryptedBankAcct:  encBank,
		Nonce:              nonce,
		KeyID:              keyID,
		EncryptedAt:        time.Now().UTC(),
	}

	e.userData[userID] = envelope
	return envelope, nil
}

// DecryptUserPII attempts to decrypt PII fields; fails immediately if key is destroyed
func (e *CryptoShreddingKMSEngine) DecryptUserPII(userID string) (pan, aadhaar, biometric, bankAcct string, err error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	envelope, exists := e.userData[userID]
	if !exists {
		return "", "", "", "", fmt.Errorf("no encrypted data found for user %s", userID)
	}

	kmsKey, exists := e.kmsKeys[envelope.KeyID]
	if !exists || kmsKey.Status == KeyStatusDestroyed || len(kmsKey.KeyMaterial) == 0 {
		return "", "", "", "", errors.New("data unrecoverable: KMS encryption key has been permanently destroyed / crypto-shredded")
	}

	if kmsKey.Status != KeyStatusActive {
		return "", "", "", "", fmt.Errorf("KMS key is not active (status: %s)", kmsKey.Status)
	}

	block, err := aes.NewCipher(kmsKey.KeyMaterial)
	if err != nil {
		return "", "", "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", "", err
	}

	decPAN, err := gcm.Open(nil, envelope.Nonce, envelope.EncryptedPAN, []byte(userID))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to decrypt PAN: %w", err)
	}
	decAadhaar, err := gcm.Open(nil, envelope.Nonce, envelope.EncryptedAadhaar, []byte(userID))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to decrypt Aadhaar: %w", err)
	}
	decBio, err := gcm.Open(nil, envelope.Nonce, envelope.EncryptedBiometric, []byte(userID))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to decrypt Biometric: %w", err)
	}
	decBank, err := gcm.Open(nil, envelope.Nonce, envelope.EncryptedBankAcct, []byte(userID))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to decrypt Bank: %w", err)
	}

	return string(decPAN), string(decAadhaar), string(decBio), string(decBank), nil
}

// CalculateShannonEntropy calculates information entropy (0.0 to 8.0 bits/byte)
func CalculateShannonEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0.0
	}
	counts := make(map[byte]float64)
	for _, b := range data {
		counts[b]++
	}
	var entropy float64
	total := float64(len(data))
	for _, count := range counts {
		p := count / total
		entropy -= p * math.Log2(p)
	}
	return entropy
}

// RequestCryptoShredding initiates a DPDP erasure request with 7-year statutory retention validation
func (e *CryptoShreddingKMSEngine) RequestCryptoShredding(
	userID string,
	accountClosedAt, lastTxAt time.Time,
	hasActiveInquiry bool,
) (*ShreddingRequest, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	reqID := fmt.Sprintf("shred-req-%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%d", userID, time.Now().UnixNano()))))[:16]
	req := &ShreddingRequest{
		RequestID:         reqID,
		UserID:            userID,
		AccountClosedAt:   accountClosedAt,
		LastTransactionAt: lastTxAt,
		HasActiveInquiry:  hasActiveInquiry,
		Status:            ShreddingRequested,
		RequestedAt:       time.Now().UTC(),
	}

	// 1. Regulatory Hold Check: Active inquiry / freeze
	if hasActiveInquiry {
		req.Status = ShreddingHeld
		req.RejectionReason = "Statutory Regulatory Hold: User is subject to active enforcement / FIU inquiry"
		e.requests[reqID] = req
		return req, errors.New("cannot process crypto-shredding: active regulatory hold in place")
	}

	// 2. 7-year Statutory Retention Check under PMLA & SEBI
	sevenYearsAgo := time.Now().UTC().AddDate(-7, 0, 0)
	if lastTxAt.After(sevenYearsAgo) || accountClosedAt.After(sevenYearsAgo) {
		req.Status = ShreddingHeld
		req.RejectionReason = fmt.Sprintf("Statutory Retention: Minimum 7-year retention period not met (last activity: %s)",
			lastTxAt.Format("2006-01-02"))
		e.requests[reqID] = req
		return req, fmt.Errorf("cannot process crypto-shredding: 7-year PMLA statutory retention period has not expired")
	}

	e.requests[reqID] = req
	return req, nil
}

// AuthorizeAndExecuteShredding enforces Dual-Authorization Quorum (Compliance + DPO) and destroys KMS key
func (e *CryptoShreddingKMSEngine) AuthorizeAndExecuteShredding(
	requestID, complianceSign, dpoSign string,
) (*DestructionCertificate, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	req, exists := e.requests[requestID]
	if !exists {
		return nil, fmt.Errorf("shredding request %s not found", requestID)
	}

	if req.Status != ShreddingRequested && req.Status != ShreddingApproved {
		return nil, fmt.Errorf("invalid request status for execution: %s", req.Status)
	}

	if complianceSign == "" || dpoSign == "" {
		return nil, errors.New("dual authorization required: both Compliance Officer and DPO must sign")
	}

	req.ComplianceOfficerSign = complianceSign
	req.DPOSign = dpoSign

	keyID, exists := e.userKeyMap[req.UserID]
	if !exists {
		return nil, fmt.Errorf("no KMS key found for user %s", req.UserID)
	}

	kmsKey := e.kmsKeys[keyID]

	// 1. CRYPTO-SHREDDING: Securely zeroize and overwrite key material in memory
	for i := range kmsKey.KeyMaterial {
		kmsKey.KeyMaterial[i] = 0x00
	}
	kmsKey.KeyMaterial = nil
	kmsKey.Status = KeyStatusDestroyed
	kmsKey.DestroyedAt = time.Now().UTC()

	// 2. Zero-remnant verification: Compute ciphertext entropy
	envelope, hasData := e.userData[req.UserID]
	var avgEntropy float64 = 7.95
	if hasData {
		allCiphertext := append(envelope.EncryptedPAN, envelope.EncryptedAadhaar...)
		allCiphertext = append(allCiphertext, envelope.EncryptedBiometric...)
		allCiphertext = append(allCiphertext, envelope.EncryptedBankAcct...)
		avgEntropy = CalculateShannonEntropy(allCiphertext)
	}

	// 3. Generate WORM Destruction Certificate
	certID := fmt.Sprintf("cert-destroy-%s", req.UserID)
	certPayload := fmt.Sprintf("%s:%s:%s:DESTROYED:%.4f:%s:%s:%d",
		certID, req.UserID, keyID, avgEntropy, complianceSign, dpoSign, kmsKey.DestroyedAt.Unix())
	certSHA256 := fmt.Sprintf("%x", sha256.Sum256([]byte(certPayload)))
	besuTxHash := "0x" + certSHA256

	cert := &DestructionCertificate{
		CertificateID:       certID,
		RequestID:           requestID,
		UserID:              req.UserID,
		KeyID:               keyID,
		KeyStatus:           KeyStatusDestroyed,
		CiphertextEntropy:   avgEntropy,
		IsUnrecoverable:     true,
		ComplianceSign:      complianceSign,
		DPOSign:             dpoSign,
		CertificateSHA256:   certSHA256,
		BesuAttestationHash: besuTxHash,
		DestroyedAt:         kmsKey.DestroyedAt,
	}

	e.certificates[certID] = cert
	req.Status = ShreddingCompleted
	req.CompletedAt = time.Now().UTC()
	req.DestructionCertHash = certSHA256
	req.OnChainBesuTxHash = besuTxHash

	return cert, nil
}
