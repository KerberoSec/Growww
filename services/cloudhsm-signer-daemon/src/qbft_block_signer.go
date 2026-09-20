package src

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"sync"
)

var (
	ErrEquivocationDetected  = errors.New("anti-equivocation violation: attempting to sign conflicting block hash for same height/round")
	ErrBlockNumberRegressed  = errors.New("block number regressed below last committed block")
)

// BlockSigningRequest represents an inbound proposal or commit from Hyperledger Besu QBFT
type BlockSigningRequest struct {
	ChainID          uint64
	ValidatorAddress string
	BlockNumber      uint64
	Round            uint32
	BlockHash        [32]byte
	PayloadDigest    [32]byte
}

// SignedBlockResponse contains the hardware signature and verification receipt
type SignedBlockResponse struct {
	BlockNumber uint64
	Round       uint32
	R           *big.Int
	S           *big.Int
	V           uint8
	SignatureHex string
}

// AntiEquivocationRecord tracks previous signatures to prevent double signing
type AntiEquivocationRecord struct {
	BlockNumber uint64
	Round       uint32
	BlockHash   [32]byte
}

// QBFTBlockSigner coordinates consensus signing with hardware anti-equivocation protection
type QBFTBlockSigner struct {
	mu             sync.RWMutex
	pool           *PKCS11SessionPool
	validatorAlias string
	lastSigned     map[string]AntiEquivocationRecord // key: chainId:validatorAddress
	highestBlock   uint64
}

// NewQBFTBlockSigner creates an instance of the QBFT consensus signer
func NewQBFTBlockSigner(pool *PKCS11SessionPool, validatorAlias string) *QBFTBlockSigner {
	return &QBFTBlockSigner{
		pool:           pool,
		validatorAlias: validatorAlias,
		lastSigned:     make(map[string]AntiEquivocationRecord),
	}
}

// SignQBFTBlock verifies anti-equivocation invariants before dispatching hardware signing
func (s *QBFTBlockSigner) SignQBFTBlock(req BlockSigningRequest) (*SignedBlockResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lookupKey := fmt.Sprintf("%d:%s", req.ChainID, req.ValidatorAddress)
	record, exists := s.lastSigned[lookupKey]

	if exists {
		// Equivocation Check: same height and round, but different block hash
		if record.BlockNumber == req.BlockNumber && record.Round == req.Round {
			if record.BlockHash != req.BlockHash {
				return nil, fmt.Errorf("%w (block %d, round %d)", ErrEquivocationDetected, req.BlockNumber, req.Round)
			}
		}

		// Height regression check
		if req.BlockNumber < record.BlockNumber {
			return nil, fmt.Errorf("%w (requested: %d, current: %d)", ErrBlockNumberRegressed, req.BlockNumber, record.BlockNumber)
		}
	}

	// Execute hardware signature
	r, sBig, err := s.pool.SignDigest(s.validatorAlias, req.PayloadDigest[:])
	if err != nil {
		return nil, fmt.Errorf("hsm block signing failed: %w", err)
	}

	// Update anti-equivocation persistent tracking state
	s.lastSigned[lookupKey] = AntiEquivocationRecord{
		BlockNumber: req.BlockNumber,
		Round:       req.Round,
		BlockHash:   req.BlockHash,
	}

	if req.BlockNumber > s.highestBlock {
		s.highestBlock = req.BlockNumber
	}

	sigHex := fmt.Sprintf("%064x%064x", r, sBig)

	return &SignedBlockResponse{
		BlockNumber:  req.BlockNumber,
		Round:        req.Round,
		R:            r,
		S:            sBig,
		V:            27,
		SignatureHex: sigHex,
	}, nil
}

// ComputeCanonicalBlockDigest computes sha256 of block header fields
func ComputeCanonicalBlockDigest(chainID, blockNumber uint64, round uint32, blockHash [32]byte) [32]byte {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d:%d:%d:", chainID, blockNumber, round)))
	h.Write(blockHash[:])
	var digest [32]byte
	copy(digest[:], h.Sum(nil))
	return digest
}
