package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// InvestorHolding represents an individual user's fractional holding
type InvestorHolding struct {
	InvestorCommitment [32]byte
	Salt               [32]byte
	Balance            uint64
}

// DepositoryReport contains custody holdings from NSDL/CDSL
type DepositoryReport struct {
	ISIN                 string
	DepositoryShareCount uint64
	ReportTimestamp      time.Time
	CustodianID          string
}

// OnChainSupplyReport contains circulating token supply from Hyperledger Besu
type OnChainSupplyReport struct {
	ISIN               string
	OnChainTokenSupply uint64
	ContractAddress    string
	BlockNumber        uint64
	Timestamp          time.Time
}

// SolvencyStatus indicates whether reserves fully cover liabilities
type SolvencyStatus string

const (
	Solvent    SolvencyStatus = "SOLVENT"
	Deficit    SolvencyStatus = "DEFICIT_BREACH"
	Unverified SolvencyStatus = "UNVERIFIED"
)

// AttestationRecord represents a finalized, cryptographically proven PoR epoch
type AttestationRecord struct {
	Epoch                uint64
	ISIN                 string
	DepositoryShareCount uint64
	OnChainTokenSupply   uint64
	MerkleRoot           [32]byte
	Status               SolvencyStatus
	DeficitAmount        uint64
	PublishedAt          time.Time
	LeavesCount          int
}

// PoREngine coordinates daily Proof-of-Reserve generation and investor verification
type PoREngine struct {
	mu           sync.RWMutex
	attestations map[string][]AttestationRecord // isin -> list of epochs
	leavesStore  map[string]map[uint64][]InvestorHolding
}

// NewPoREngine creates a new Proof-of-Reserve engine
func NewPoREngine() *PoREngine {
	return &PoREngine{
		attestations: make(map[string][]AttestationRecord),
		leavesStore:  make(map[string]map[uint64][]InvestorHolding),
	}
}

// ComputeLeafHash generates a zero-PII salted leaf hash: SHA256(commitment + salt + balanceBytes)
func ComputeLeafHash(holding InvestorHolding) [32]byte {
	h := sha256.New()
	h.Write(holding.InvestorCommitment[:])
	h.Write(holding.Salt[:])
	balanceBytes := []byte(fmt.Sprintf("%d", holding.Balance))
	h.Write(balanceBytes)
	var leaf [32]byte
	copy(leaf[:], h.Sum(nil))
	return leaf
}

// BuildMerkleTree builds a sorted-pair Merkle tree from investor holdings
func BuildMerkleTree(holdings []InvestorHolding) ([32]byte, [][32]byte) {
	if len(holdings) == 0 {
		return [32]byte{}, nil
	}

	leaves := make([][32]byte, len(holdings))
	for i, h := range holdings {
		leaves[i] = ComputeLeafHash(h)
	}

	// Sort leaves to guarantee canonical determinism
	sort.Slice(leaves, func(i, j int) bool {
		return hex.EncodeToString(leaves[i][:]) < hex.EncodeToString(leaves[j][:])
	})

	currentLevel := leaves
	for len(currentLevel) > 1 {
		var nextLevel [][32]byte
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				nextLevel = append(nextLevel, hashPair(currentLevel[i], currentLevel[i+1]))
			} else {
				// Odd leaf promotes upward
				nextLevel = append(nextLevel, currentLevel[i])
			}
		}
		currentLevel = nextLevel
	}

	return currentLevel[0], leaves
}

func hashPair(a, b [32]byte) [32]byte {
	// Canonical sorted pair hashing
	h := sha256.New()
	if hex.EncodeToString(a[:]) <= hex.EncodeToString(b[:]) {
		h.Write(a[:])
		h.Write(b[:])
	} else {
		h.Write(b[:])
		h.Write(a[:])
	}
	var res [32]byte
	copy(res[:], h.Sum(nil))
	return res
}

// GenerateProof generates an inclusion proof path for an investor holding
func GenerateProof(target InvestorHolding, sortedLeaves [][32]byte) ([][32]byte, error) {
	targetLeaf := ComputeLeafHash(target)
	targetIdx := -1
	for i, l := range sortedLeaves {
		if l == targetLeaf {
			targetIdx = i
			break
		}
	}
	if targetIdx == -1 {
		return nil, errors.New("target holding not found in leaf set")
	}

	var proof [][32]byte
	currentLevel := sortedLeaves
	idx := targetIdx

	for len(currentLevel) > 1 {
		var nextLevel [][32]byte
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				if i == idx {
					proof = append(proof, currentLevel[i+1])
				} else if i+1 == idx {
					proof = append(proof, currentLevel[i])
				}
				nextLevel = append(nextLevel, hashPair(currentLevel[i], currentLevel[i+1]))
			} else {
				nextLevel = append(nextLevel, currentLevel[i])
			}
		}
		idx = idx / 2
		currentLevel = nextLevel
	}

	return proof, nil
}

// VerifyInclusion verifies client-side that an investor's balance is part of the Merkle root
func VerifyInclusion(target InvestorHolding, proof [][32]byte, root [32]byte) bool {
	computed := ComputeLeafHash(target)
	for _, p := range proof {
		computed = hashPair(computed, p)
	}
	return computed == root
}

// ReconcileAndPublish processes daily reconciliation and anchors an epoch
func (e *PoREngine) ReconcileAndPublish(
	custody DepositoryReport,
	onChain OnChainSupplyReport,
	holdings []InvestorHolding,
) (AttestationRecord, error) {
	if custody.ISIN != onChain.ISIN {
		return AttestationRecord{}, errors.New("ISIN mismatch between depository and on-chain report")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	merkleRoot, _ := BuildMerkleTree(holdings)

	status := Solvent
	var deficit uint64 = 0
	if onChain.OnChainTokenSupply > custody.DepositoryShareCount {
		status = Deficit
		deficit = onChain.OnChainTokenSupply - custody.DepositoryShareCount
	}

	epoch := uint64(len(e.attestations[custody.ISIN]) + 1)
	record := AttestationRecord{
		Epoch:                epoch,
		ISIN:                 custody.ISIN,
		DepositoryShareCount: custody.DepositoryShareCount,
		OnChainTokenSupply:   onChain.OnChainTokenSupply,
		MerkleRoot:           merkleRoot,
		Status:               status,
		DeficitAmount:        deficit,
		PublishedAt:          time.Now().UTC(),
		LeavesCount:          len(holdings),
	}

	e.attestations[custody.ISIN] = append(e.attestations[custody.ISIN], record)

	if e.leavesStore[custody.ISIN] == nil {
		e.leavesStore[custody.ISIN] = make(map[uint64][]InvestorHolding)
	}
	e.leavesStore[custody.ISIN][epoch] = holdings

	return record, nil
}

func (e *PoREngine) GetLatestAttestation(isin string) (AttestationRecord, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	epochs := e.attestations[isin]
	if len(epochs) == 0 {
		return AttestationRecord{}, errors.New("no attestations published for ISIN")
	}
	return epochs[len(epochs)-1], nil
}

func main() {
	fmt.Println("Growww Proof-of-Reserve Verification Service initialized.")
}
