package timescaledb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ReconciliationReport contains verification results.
type ReconciliationReport struct {
	Timestamp           time.Time `json:"timestamp"`
	TotalCandlesChecked int       `json:"total_candles_checked"`
	TotalChunksChecked  int       `json:"total_chunks_checked"`
	MerkleRoot          string    `json:"merkle_root"`
	BesuAnchorTxHash    string    `json:"besu_anchor_tx_hash"`
	BesuBlockNumber     uint64    `json:"besu_block_number"`
	Consistent          bool      `json:"consistent"`
	MismatchCount       int       `json:"mismatch_count"`
	Errors              []string  `json:"errors"`
}

// ReconciliationManager runs periodic reconciliation against immutable ledger states.
type ReconciliationManager struct {
	mu            sync.Mutex
	engine        *TimescaleCandleEngine
	lastReport    *ReconciliationReport
	besuBlockNum  uint64
}

// NewReconciliationManager creates a new reconciliation manager.
func NewReconciliationManager(engine *TimescaleCandleEngine) *ReconciliationManager {
	return &ReconciliationManager{
		engine:       engine,
		besuBlockNum: 1000000,
	}
}

// RunReconciliationLoop executes verification of in-memory candles against state hashes and chunk Merkle roots.
func (rm *ReconciliationManager) RunReconciliationLoop() *ReconciliationReport {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	report := &ReconciliationReport{
		Timestamp:  time.Now(),
		Consistent: true,
		Errors:     make([]string, 0),
	}

	rm.engine.mu.RLock()
	var hashes []string

	// Check active candles
	for key, c := range rm.engine.activeCandles {
		expectedHash := c.ComputeStateHash()
		if c.StateHash != expectedHash {
			report.Consistent = false
			report.MismatchCount++
			report.Errors = append(report.Errors, fmt.Sprintf("Hash mismatch for active candle %s: current=%s expected=%s", key, c.StateHash, expectedHash))
		}
		hashes = append(hashes, c.StateHash)
		report.TotalCandlesChecked++
	}

	// Check chunks
	for chunkID, chunk := range rm.engine.chunks {
		if chunk.Status == ChunkStatusCompressed && chunk.MerkleRoot == "" {
			report.Consistent = false
			report.MismatchCount++
			report.Errors = append(report.Errors, fmt.Sprintf("Compressed chunk %s missing Merkle root", chunkID))
		}
		hashes = append(hashes, chunk.MerkleRoot)
		report.TotalChunksChecked++
	}
	rm.engine.mu.RUnlock()

	// Compute aggregate Merkle root across all state hashes
	sort.Strings(hashes)
	h := sha256.New()
	for _, hash := range hashes {
		h.Write([]byte(hash))
	}
	report.MerkleRoot = hex.EncodeToString(h.Sum(nil))

	// Simulate Hyperledger Besu QBFT Block anchoring
	rm.besuBlockNum++
	report.BesuBlockNumber = rm.besuBlockNum
	anchorHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", report.MerkleRoot, report.BesuBlockNumber)))
	report.BesuAnchorTxHash = "0x" + hex.EncodeToString(anchorHash[:])

	rm.engine.mu.Lock()
	rm.engine.metrics.LastReconciliationMs = time.Now().UnixMilli()
	rm.engine.mu.Unlock()

	rm.lastReport = report
	return report
}

// GetLastReport returns the latest reconciliation report.
func (rm *ReconciliationManager) GetLastReport() *ReconciliationReport {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	return rm.lastReport
}
