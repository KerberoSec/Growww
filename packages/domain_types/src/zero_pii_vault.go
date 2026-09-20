package src

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

var (
	ErrInvalidToken = errors.New("piivault: invalid token or decryption failure")
	ErrExpiredSalt  = errors.New("piivault: cryptographic salt expired")
)

// PIITokenType defines the type of PII identifier being tokenized.
type PIITokenType string

const (
	TokenTypePAN     PIITokenType = "PAN"      // Indian Permanent Account Number
	TokenTypeAadhaar PIITokenType = "AADHAAR"  // 12-digit UIDAI ID
	TokenTypePhone   PIITokenType = "PHONE"    // Mobile phone number
	TokenTypeEmail   PIITokenType = "EMAIL"    // Email address
	TokenTypeBankAcc PIITokenType = "BANK_ACC" // Bank Account Number
)

// SaltEntropyManager manages cryptographic salts with rotation cycles and CSPRNG entropy.
type SaltEntropyManager struct {
	mu           sync.RWMutex
	currentSalt  []byte
	saltVersion  uint32
	createdAt    time.Time
	rotationTTL  time.Duration
	historicSalt map[uint32][]byte
}

// NewSaltEntropyManager creates a new salt manager.
func NewSaltEntropyManager(ttl time.Duration) (*SaltEntropyManager, error) {
	if ttl <= 0 {
		ttl = 30 * 24 * time.Hour // 30-day rotation default
	}

	initialSalt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, initialSalt); err != nil {
		return nil, fmt.Errorf("failed generating cryptographically secure salt: %w", err)
	}

	return &SaltEntropyManager{
		currentSalt:  initialSalt,
		saltVersion:  1,
		createdAt:    time.Now().UTC(),
		rotationTTL:  ttl,
		historicSalt: make(map[uint32][]byte),
	}, nil
}

// GetActiveSalt returns current active salt and version.
func (s *SaltEntropyManager) GetActiveSalt() ([]byte, uint32) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	saltCopy := make([]byte, len(s.currentSalt))
	copy(saltCopy, s.currentSalt)
	return saltCopy, s.saltVersion
}

// RotateSalt rotates the salt, archiving the previous version.
func (s *SaltEntropyManager) RotateSalt() (uint32, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newSalt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, newSalt); err != nil {
		return 0, err
	}

	s.historicSalt[s.saltVersion] = s.currentSalt
	s.currentSalt = newSalt
	s.saltVersion++
	s.createdAt = time.Now().UTC()
	return s.saltVersion, nil
}

// ZeroPIIVault implements DPDP Act 2023 zero-PII storage via AES-256-GCM and HMAC blind indexing.
type ZeroPIIVault struct {
	masterKey   [32]byte
	saltManager *SaltEntropyManager
	gcm         cipher.AEAD
}

// NewZeroPIIVault creates a new zero-PII tokenization vault.
func NewZeroPIIVault(masterKey [32]byte, saltManager *SaltEntropyManager) (*ZeroPIIVault, error) {
	block, err := aes.NewCipher(masterKey[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return &ZeroPIIVault{
		masterKey:   masterKey,
		saltManager: saltManager,
		gcm:         gcm,
	}, nil
}

// Tokenize encrypts plaintext PII with AES-GCM and generates a deterministic blind index for searches.
func (v *ZeroPIIVault) Tokenize(tokenType PIITokenType, plaintext string) (token string, blindIndex string, err error) {
	if plaintext == "" {
		return "", "", errors.New("cannot tokenize empty plaintext")
	}

	// 1. Generate random nonce for AES-GCM
	nonce := make([]byte, v.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}

	// 2. Encrypt with AES-256-GCM
	ciphertext := v.gcm.Seal(nonce, nonce, []byte(plaintext), []byte(tokenType))
	token = hex.EncodeToString(ciphertext)

	// 3. Generate deterministic blind index using HMAC-SHA256 for searching without decrypting
	salt, version := v.saltManager.GetActiveSalt()
	mac := hmac.New(sha256.New, salt)
	mac.Write([]byte(string(tokenType) + ":" + plaintext))
	blindIndex = fmt.Sprintf("v%d:%s", version, hex.EncodeToString(mac.Sum(nil)))

	return token, blindIndex, nil
}

// Detokenize decrypts token back to original plaintext PII.
func (v *ZeroPIIVault) Detokenize(tokenType PIITokenType, token string) (string, error) {
	raw, err := hex.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("%w: invalid hex", ErrInvalidToken)
	}

	nonceSize := v.gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("%w: ciphertext too short", ErrInvalidToken)
	}

	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := v.gcm.Open(nil, nonce, ciphertext, []byte(tokenType))
	if err != nil {
		return "", fmt.Errorf("%w: AEAD tag verification failed", ErrInvalidToken)
	}

	return string(plaintext), nil
}
