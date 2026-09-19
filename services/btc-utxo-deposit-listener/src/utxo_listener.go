package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// DepositStatus represents the lifecycle of an on-chain BTC deposit
type DepositStatus string

const (
	StatusDetected   DepositStatus = "DETECTED"
	StatusConfirming DepositStatus = "CONFIRMING"
	StatusCredited   DepositStatus = "CREDITED"
	StatusFinalized  DepositStatus = "FINALIZED"
)

// UTXODeposit represents a tracked Bitcoin deposit UTXO
type UTXODeposit struct {
	TxID          string        `json:"tx_id"`
	Vout          uint32        `json:"vout"`
	Address       string        `json:"address"` // P2TR (bc1p) or P2WPKH (bc1q)
	ScriptPubKey  string        `json:"script_pubkey"`
	AmountSats    uint64        `json:"amount_sats"`
	Confirmations uint32        `json:"confirmations"`
	BlockHeight   uint64        `json:"block_height"`
	BlockHash     string        `json:"block_hash"`
	Status        DepositStatus `json:"status"`
	DetectedAt    time.Time     `json:"detected_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// SPVProof holds Merkle inclusion proof for a Bitcoin transaction
type SPVProof struct {
	TxID        string   `json:"tx_id"`
	MerkleRoot  string   `json:"merkle_root"`
	BlockHeight uint64   `json:"block_height"`
	IntermediateHashes []string `json:"intermediate_hashes"`
	TxIndex     uint32   `json:"tx_index"`
}

// UTXOListener coordinates block parsing, address matching, and SPV verification
type UTXOListener struct {
	mu             sync.RWMutex
	monitoredAddrs map[string]string // Address -> UserID
	deposits       map[string]*UTXODeposit
	minConfirmations uint32
}

// NewUTXOListener creates a new Bitcoin UTXO listener instance
func NewUTXOListener(minConfs uint32) *UTXOListener {
	if minConfs == 0 {
		minConfs = 3 // Default 3 confirmations for trading credit
	}
	return &UTXOListener{
		monitoredAddrs:   make(map[string]string),
		deposits:         make(map[string]*UTXODeposit),
		minConfirmations: minConfs,
	}
}

// RegisterDepositAddress registers a Taproot or SegWit address belonging to a user
func (l *UTXOListener) RegisterDepositAddress(address string, userID string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.monitoredAddrs[address] = userID
}

// ProcessTransaction parses a raw block transaction and checks for monitored deposit outputs
func (l *UTXOListener) ProcessTransaction(txID string, vout uint32, destAddress string, amountSats uint64, blockHeight uint64, blockHash string) *UTXODeposit {
	l.mu.Lock()
	defer l.mu.Unlock()

	userID, exists := l.monitoredAddrs[destAddress]
	if !exists {
		return nil
	}

	key := fmt.Sprintf("%s:%d", txID, vout)
	dep, found := l.deposits[key]
	if !found {
		dep = &UTXODeposit{
			TxID:          txID,
			Vout:          vout,
			Address:       destAddress,
			AmountSats:    amountSats,
			Confirmations: 1,
			BlockHeight:   blockHeight,
			BlockHash:     blockHash,
			Status:        StatusConfirming,
			DetectedAt:    time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		if dep.Confirmations >= l.minConfirmations {
			dep.Status = StatusCredited
		}
		l.deposits[key] = dep
		fmt.Printf("[UTXO Ingress] Detected new BTC deposit for user %s: %d Sats at %s\n", userID, amountSats, key)
	}
	return dep
}

// UpdateConfirmations increments block confirmations and transitions deposit states
func (l *UTXOListener) UpdateConfirmations(txKey string, currentTipHeight uint64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	dep, exists := l.deposits[txKey]
	if !exists || dep.BlockHeight == 0 {
		return
	}

	if currentTipHeight >= dep.BlockHeight {
		dep.Confirmations = uint32(currentTipHeight - dep.BlockHeight + 1)
	}

	if dep.Confirmations >= 6 {
		dep.Status = StatusFinalized
	} else if dep.Confirmations >= l.minConfirmations {
		dep.Status = StatusCredited
	}
	dep.UpdatedAt = time.Now().UTC()
}

// VerifySPV verifies a Merkle inclusion proof against a verified Bitcoin block header Merkle root
func VerifySPV(proof SPVProof) bool {
	currentHash, err := hex.DecodeString(proof.TxID)
	if err != nil {
		return false
	}

	// Reverse endianness to match internal Bitcoin little-endian hashes
	reverseBytes(currentHash)

	for _, hStr := range proof.IntermediateHashes {
		sibling, err := hex.DecodeString(hStr)
		if err != nil {
			return false
		}
		reverseBytes(sibling)

		// Double SHA-256 concatenation
		combined := append(currentHash, sibling...)
		firstPass := sha256.Sum256(combined)
		secondPass := sha256.Sum256(firstPass[:])
		currentHash = secondPass[:]
	}

	reverseBytes(currentHash)
	return hex.EncodeToString(currentHash) == proof.MerkleRoot
}

func reverseBytes(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}

func main() {
	listener := NewUTXOListener(3)
	listener.RegisterDepositAddress("bc1p5d7rjq7g6rd2ee076mqeycesdafapx85e0wxyus3ulmwvd8m5w6qncx9dn", "user-uuid-101")
	fmt.Println("BTC UTXO Taproot SPV Ingress Service active.")
}
