package merkle

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

// AccountLiability represents an individual user's attested balance in satoshis (e8)
type AccountLiability struct {
	UserID    string   `json:"user_id"`
	BalanceE8 uint64   `json:"balance_e8"`
	LeafHash  [32]byte `json:"leaf_hash"`
}

// SolvencyTree manages the Merkle tree construction and proof generation
type SolvencyTree struct {
	Accounts []AccountLiability
	Leaves   [][32]byte
	Levels   [][][32]byte
	Root     [32]byte
}

// BytesCompare compares two 32-byte hashes lexically, matching Solidity's `a <= b`
func BytesCompare(a, b [32]byte) int {
	return bytes.Compare(a[:], b[:])
}

// ComputeLeaf hashes user ID and balance into a 32-byte leaf
// Matches MerkleRegistry.sol: keccak256(abi.encodePacked(keccak256(userID), uint256(balanceE8)))
func ComputeLeaf(userID string, balanceE8 uint64) [32]byte {
	userHash := Keccak256([]byte(userID))
	var balanceBytes [32]byte
	binary.BigEndian.PutUint64(balanceBytes[24:], balanceE8)

	var payload [64]byte
	copy(payload[:32], userHash[:])
	copy(payload[32:], balanceBytes[:])

	return Keccak256(payload[:])
}

// HashPair hashes two 32-byte sibling nodes canonically in sorted order
// Matches MerkleRegistry.sol: if (a <= b) keccak256(a, b) else keccak256(b, a)
func HashPair(a, b [32]byte) [32]byte {
	var combined [64]byte
	if BytesCompare(a, b) <= 0 {
		copy(combined[:32], a[:])
		copy(combined[32:], b[:])
	} else {
		copy(combined[:32], b[:])
		copy(combined[32:], a[:])
	}
	return Keccak256(combined[:])
}

// VerifyProof verifies that a given leaf belongs to the Merkle tree with the given root
func VerifyProof(leaf [32]byte, proof [][32]byte, root [32]byte) bool {
	computed := leaf
	for _, sibling := range proof {
		computed = HashPair(computed, sibling)
	}
	return computed == root
}

// BuildSolvencyTree constructs a canonical Merkle tree from accounts
func BuildSolvencyTree(accounts []AccountLiability) (*SolvencyTree, error) {
	if len(accounts) == 0 {
		return nil, errors.New("cannot build solvency tree with zero accounts")
	}

	leaves := make([][32]byte, len(accounts))
	for i, acc := range accounts {
		leaf := ComputeLeaf(acc.UserID, acc.BalanceE8)
		accounts[i].LeafHash = leaf
		leaves[i] = leaf
	}

	tree := &SolvencyTree{
		Accounts: accounts,
		Leaves:   leaves,
	}

	currentLevel := make([][32]byte, len(leaves))
	copy(currentLevel, leaves)
	tree.Levels = append(tree.Levels, currentLevel)

	for len(currentLevel) > 1 {
		var nextLevel [][32]byte
		for i := 0; i < len(currentLevel); i += 2 {
			if i+1 < len(currentLevel) {
				nextLevel = append(nextLevel, HashPair(currentLevel[i], currentLevel[i+1]))
			} else {
				// Odd number of elements: duplicate last element
				nextLevel = append(nextLevel, HashPair(currentLevel[i], currentLevel[i]))
			}
		}
		tree.Levels = append(tree.Levels, nextLevel)
		currentLevel = nextLevel
	}

	tree.Root = currentLevel[0]
	return tree, nil
}

// GetProof generates the inclusion proof path for account at accountIndex
func (t *SolvencyTree) GetProof(accountIndex int) ([][32]byte, error) {
	if accountIndex < 0 || accountIndex >= len(t.Accounts) {
		return nil, fmt.Errorf("invalid account index: %d", accountIndex)
	}

	var proof [][32]byte
	idx := accountIndex

	for _, level := range t.Levels[:len(t.Levels)-1] {
		isOdd := len(level)%2 != 0
		if idx%2 == 0 {
			if idx+1 < len(level) {
				proof = append(proof, level[idx+1])
			} else if isOdd && idx == len(level)-1 {
				proof = append(proof, level[idx])
			}
		} else {
			proof = append(proof, level[idx-1])
		}
		idx /= 2
	}

	return proof, nil
}

// ValidateSolvency checks that total reserves >= total liabilities
func ValidateSolvency(totalLiabilities, totalReserves uint64) (float64, error) {
	if totalReserves < totalLiabilities {
		return float64(totalReserves) / float64(totalLiabilities), fmt.Errorf(
			"insolvency detected: reserves (%d) < liabilities (%d)",
			totalReserves, totalLiabilities,
		)
	}
	if totalLiabilities == 0 {
		return 1.0, nil
	}
	return float64(totalReserves) / float64(totalLiabilities), nil
}

// Hex converts a 32-byte hash to a 0x prefixed hex string
func Hex(b [32]byte) string {
	return "0x" + hex.EncodeToString(b[:])
}
