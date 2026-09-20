package src

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrContractNotWhitelisted = errors.New("verifying contract address is not whitelisted for settlement signing")
	ErrGrossValueLimitExceeded = errors.New("settlement batch gross value exceeds single-batch safety collar")
	ErrBatchExpired           = errors.New("settlement batch timestamp has expired")
)

// EIP712Domain represents EIP-712 typed domain separator fields
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           uint64
	VerifyingContract string
}

// SettlementBatchData represents trade clearing batch fields
type SettlementBatchData struct {
	BatchID         string
	MerkleRoot      [32]byte
	TradeCount      uint32
	GrossValueINR   float64
	ExpiryTimestamp int64
}

// EIP712BatchSigner executes domain-separated ECDSA signatures for trade clearing
type EIP712BatchSigner struct {
	mu                   sync.RWMutex
	pool                 *PKCS11SessionPool
	relayerAlias         string
	whitelistedContracts map[string]bool
	maxBatchValueINR     float64
}

// NewEIP712BatchSigner initializes EIP-712 signer
func NewEIP712BatchSigner(
	pool *PKCS11SessionPool,
	relayerAlias string,
	approvedContracts []string,
	maxBatchValueINR float64,
) *EIP712BatchSigner {
	wl := make(map[string]bool)
	for _, c := range approvedContracts {
		wl[strings.ToLower(c)] = true
	}

	if maxBatchValueINR <= 0 {
		maxBatchValueINR = 500000000.0 // Default 50 Crore INR max batch value
	}

	return &EIP712BatchSigner{
		pool:                 pool,
		relayerAlias:         relayerAlias,
		whitelistedContracts: wl,
		maxBatchValueINR:     maxBatchValueINR,
	}
}

// ComputeDomainSeparator calculates EIP-712 domain separator hash
func ComputeDomainSeparator(domain EIP712Domain) [32]byte {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract):%s:%s:%d:%s",
		domain.Name, domain.Version, domain.ChainID, strings.ToLower(domain.VerifyingContract))))
	var res [32]byte
	copy(res[:], h.Sum(nil))
	return res
}

// ComputeBatchStructHash calculates EIP-712 struct hash for settlement batch
func ComputeBatchStructHash(batch SettlementBatchData) [32]byte {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("BatchSettlement(string batchId,bytes32 merkleRoot,uint32 tradeCount,uint256 grossValueINR,uint64 expiry):%s:%x:%d:%.2f:%d",
		batch.BatchID, batch.MerkleRoot, batch.TradeCount, batch.GrossValueINR, batch.ExpiryTimestamp)))
	var res [32]byte
	copy(res[:], h.Sum(nil))
	return res
}

// SignSettlementBatch validates safety collars and computes EIP-712 signature
func (s *EIP712BatchSigner) SignSettlementBatch(
	domain EIP712Domain,
	batch SettlementBatchData,
	currentTimeEpoch int64,
) (sigHex string, digest [32]byte, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 1. Validate contract whitelist
	if !s.whitelistedContracts[strings.ToLower(domain.VerifyingContract)] {
		return "", digest, fmt.Errorf("%w: %s", ErrContractNotWhitelisted, domain.VerifyingContract)
	}

	// 2. Validate safety notional collar
	if batch.GrossValueINR > s.maxBatchValueINR {
		return "", digest, fmt.Errorf("%w: %.2f > max %.2f", ErrGrossValueLimitExceeded, batch.GrossValueINR, s.maxBatchValueINR)
	}

	// 3. Validate expiration
	if batch.ExpiryTimestamp < currentTimeEpoch {
		return "", digest, ErrBatchExpired
	}

	// 4. Compute EIP-712 typed digest: "\x19\x01" || DomainSeparator || StructHash
	domSep := ComputeDomainSeparator(domain)
	structHash := ComputeBatchStructHash(batch)

	hasher := sha256.New()
	hasher.Write([]byte("\x19\x01"))
	hasher.Write(domSep[:])
	hasher.Write(structHash[:])
	copy(digest[:], hasher.Sum(nil))

	// 5. Hardware signing
	r, sBig, err := s.pool.SignDigest(s.relayerAlias, digest[:])
	if err != nil {
		return "", digest, fmt.Errorf("eip712 hsm signing failed: %w", err)
	}

	rBytes := r.Bytes()
	sBytes := sBig.Bytes()
	sigHex = fmt.Sprintf("%064s%064s1b", hex.EncodeToString(rBytes), hex.EncodeToString(sBytes))
	return sigHex, digest, nil
}
