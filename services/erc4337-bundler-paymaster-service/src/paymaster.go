package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

var (
	ErrSenderNotKYCVerified     = errors.New("sender account is not KYC verified")
	ErrSenderNotEligible        = errors.New("sender is not eligible for gas sponsorship")
	ErrDailyGasQuotaExceeded    = errors.New("daily paymaster gas quota limit exceeded")
	ErrInvalidPaymasterData     = errors.New("invalid paymasterAndData format")
	ErrPaymasterExpired         = errors.New("paymaster sponsorship window has expired or is not yet valid")
	ErrInvalidPaymasterAddress  = errors.New("paymaster address in UserOp does not match verifying paymaster")
	ErrInvalidPaymasterSig      = errors.New("paymaster signature validation failed")
)

type GrowwwPaymaster struct {
	mu               sync.RWMutex
	paymasterAddress string
	chainID          *big.Int
	secretKey        []byte
	quotas           map[string]*AccountQuota // sender -> quota
	defaultQuota     uint64
}

func NewGrowwwPaymaster(paymasterAddress string, chainID *big.Int, secretKey []byte) *GrowwwPaymaster {
	cleanAddr := strings.ToLower(paymasterAddress)
	if !strings.HasPrefix(cleanAddr, "0x") {
		cleanAddr = "0x" + cleanAddr
	}
	return &GrowwwPaymaster{
		paymasterAddress: cleanAddr,
		chainID:          chainID,
		secretKey:        secretKey,
		quotas:           make(map[string]*AccountQuota),
		defaultQuota:     15_000_000, // 15 million gas units daily
	}
}

// RegisterSender registers or updates an account's KYC and sponsorship status
func (pm *GrowwwPaymaster) RegisterSender(sender string, userID string, kycStatus string, eligible bool, quota uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	clean := strings.ToLower(sender)
	if quota == 0 {
		quota = pm.defaultQuota
	}
	pm.quotas[clean] = &AccountQuota{
		SenderAddress:         clean,
		UserID:                userID,
		KYCStatus:             kycStatus,
		DailyGasQuotaLimit:    quota,
		DailyGasQuotaConsumed: 0,
		IsSponsoredEligible:   eligible && kycStatus == "VERIFIED",
		LastQuotaResetDate:    time.Now().UTC().Format("2006-01-02"),
	}
}

// GetAccountQuota returns quota status for a sender
func (pm *GrowwwPaymaster) GetAccountQuota(sender string) (*AccountQuota, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	q, ok := pm.quotas[strings.ToLower(sender)]
	if !ok {
		return nil, false
	}
	cpy := *q
	return &cpy, true
}

// SponsorUserOp verifies compliance, updates quota, and generates paymasterAndData
func (pm *GrowwwPaymaster) SponsorUserOp(op *UserOperation, validityDuration time.Duration) (string, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	sender := strings.ToLower(op.Sender)
	quota, exists := pm.quotas[sender]
	if !exists {
		// Auto-register pending if not existing
		quota = &AccountQuota{
			SenderAddress:         sender,
			UserID:                "auto-" + sender[2:8],
			KYCStatus:             "PENDING",
			DailyGasQuotaLimit:    pm.defaultQuota,
			DailyGasQuotaConsumed: 0,
			IsSponsoredEligible:   false,
			LastQuotaResetDate:    time.Now().UTC().Format("2006-01-02"),
		}
		pm.quotas[sender] = quota
	}

	// 1. Verify KYC & Eligibility
	if quota.KYCStatus != "VERIFIED" {
		return "", fmt.Errorf("%w: current status %s", ErrSenderNotKYCVerified, quota.KYCStatus)
	}
	if !quota.IsSponsoredEligible {
		return "", ErrSenderNotEligible
	}

	// 2. Check Daily Quota
	today := time.Now().UTC().Format("2006-01-02")
	if quota.LastQuotaResetDate != today {
		quota.DailyGasQuotaConsumed = 0
		quota.LastQuotaResetDate = today
	}

	// Estimated max gas: verificationGasLimit + callGasLimit + preVerificationGas
	estGas := big.NewInt(0)
	if op.VerificationGasLimit != nil {
		estGas.Add(estGas, op.VerificationGasLimit)
	}
	if op.CallGasLimit != nil {
		estGas.Add(estGas, op.CallGasLimit)
	}
	if op.PreVerificationGas != nil {
		estGas.Add(estGas, op.PreVerificationGas)
	}

	if quota.DailyGasQuotaConsumed+estGas.Uint64() > quota.DailyGasQuotaLimit {
		return "", fmt.Errorf("%w: requested %d, remaining %d",
			ErrDailyGasQuotaExceeded, estGas.Uint64(), quota.DailyGasQuotaLimit-quota.DailyGasQuotaConsumed)
	}

	// 3. Timestamps: validAfter = now - 60s (clock skew buffer), validUntil = now + duration
	nowSec := uint64(time.Now().UTC().Unix())
	validAfter := nowSec - 60
	validUntil := nowSec + uint64(48 * time.Hour.Seconds())

	// 4. Construct EIP-712 paymaster digest
	digest := pm.computeSponsorshipDigest(op, validUntil, validAfter)

	// 5. Sign with secret key
	sigHex := GeneratePaymasterSignature(pm.secretKey, digest)

	// 6. Encode paymasterAndData:
	// 20 bytes (paymaster address) + 6 bytes (validUntil) + 6 bytes (validAfter) + 65 bytes (sig)
	pmAddrBytes, _ := hex.DecodeString(strings.TrimPrefix(pm.paymasterAddress, "0x"))
	var validUntilBytes [6]byte
	var validAfterBytes [6]byte
	putUint48(validUntilBytes[:], validUntil)
	putUint48(validAfterBytes[:], validAfter)
	sigBytes, _ := hex.DecodeString(strings.TrimPrefix(sigHex, "0x"))

	var fullData []byte
	fullData = append(fullData, pmAddrBytes...)
	fullData = append(fullData, validUntilBytes[:]...)
	fullData = append(fullData, validAfterBytes[:]...)
	fullData = append(fullData, sigBytes...)

	paymasterAndData := "0x" + hex.EncodeToString(fullData)
	op.PaymasterAndData = paymasterAndData

	return paymasterAndData, nil
}

// ValidatePaymasterUserOp verifies an incoming sponsored UserOperation
func (pm *GrowwwPaymaster) ValidatePaymasterUserOp(op *UserOperation) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	raw := strings.TrimPrefix(op.PaymasterAndData, "0x")
	// Expected length: 20 (addr) + 6 (until) + 6 (after) + 65 (sig) = 97 bytes = 194 hex characters
	if len(raw) < 194 {
		return fmt.Errorf("%w: length %d, expected >= 194", ErrInvalidPaymasterData, len(raw))
	}

	data, err := hex.DecodeString(raw)
	if err != nil {
		return err
	}

	addr := "0x" + hex.EncodeToString(data[0:20])
	if !strings.EqualFold(addr, pm.paymasterAddress) {
		return fmt.Errorf("%w: got %s, expected %s", ErrInvalidPaymasterAddress, addr, pm.paymasterAddress)
	}

	validUntil := getUint48(data[20:26])
	validAfter := getUint48(data[26:32])
	sigHex := "0x" + hex.EncodeToString(data[32:97])

	now := uint64(time.Now().UTC().Unix())
	if now < validAfter || now > validUntil {
		return fmt.Errorf("%w: now %d not in [%d, %d]", ErrPaymasterExpired, now, validAfter, validUntil)
	}

	digest := pm.computeSponsorshipDigest(op, validUntil, validAfter)
	if !VerifyPaymasterSignature(pm.secretKey, digest, sigHex) {
		return ErrInvalidPaymasterSig
	}

	return nil
}

// RecordConsumedGas records actual on-chain gas consumed by sponsored user
func (pm *GrowwwPaymaster) RecordConsumedGas(sender string, gasUsed uint64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	clean := strings.ToLower(sender)
	if quota, exists := pm.quotas[clean]; exists {
		quota.DailyGasQuotaConsumed += gasUsed
	}
}

func (pm *GrowwwPaymaster) computeSponsorshipDigest(op *UserOperation, validUntil, validAfter uint64) []byte {
	var buf []byte
	senderBytes, _ := addressToBytes32(op.Sender)
	buf = append(buf, senderBytes...)
	buf = append(buf, bigIntToBytes32(op.Nonce)...)

	initBytes, _ := decodeHex(op.InitCode)
	buf = append(buf, Keccak256(initBytes)...)

	callBytes, _ := decodeHex(op.CallData)
	buf = append(buf, Keccak256(callBytes)...)

	buf = append(buf, bigIntToBytes32(op.CallGasLimit)...)
	buf = append(buf, bigIntToBytes32(op.VerificationGasLimit)...)
	buf = append(buf, bigIntToBytes32(op.PreVerificationGas)...)
	buf = append(buf, bigIntToBytes32(op.MaxFeePerGas)...)
	buf = append(buf, bigIntToBytes32(op.MaxPriorityFeePerGas)...)

	var uBytes [32]byte
	binary.BigEndian.PutUint64(uBytes[24:], validUntil)
	buf = append(buf, uBytes[:]...)

	var aBytes [32]byte
	binary.BigEndian.PutUint64(aBytes[24:], validAfter)
	buf = append(buf, aBytes[:]...)

	buf = append(buf, bigIntToBytes32(pm.chainID)...)
	return Keccak256(buf)
}

func putUint48(b []byte, v uint64) {
	b[0] = byte(v >> 40)
	b[1] = byte(v >> 32)
	b[2] = byte(v >> 24)
	b[3] = byte(v >> 16)
	b[4] = byte(v >> 8)
	b[5] = byte(v)
}

func getUint48(b []byte) uint64 {
	return uint64(b[0])<<40 |
		uint64(b[1])<<32 |
		uint64(b[2])<<24 |
		uint64(b[3])<<16 |
		uint64(b[4])<<8 |
		uint64(b[5])
}
