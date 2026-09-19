package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Domain types
// ──────────────────────────────────────────────────────────────────────────────

// ReorgSeverity indicates the impact level of a chain reorganization.
type ReorgSeverity string

const (
	ReorgMinor    ReorgSeverity = "MINOR"    // 1-2 blocks
	ReorgModerate ReorgSeverity = "MODERATE" // 3-6 blocks
	ReorgCritical ReorgSeverity = "CRITICAL" // 7+ blocks
)

// SagaPhase tracks the lifecycle of a compensating transaction saga.
type SagaPhase string

const (
	SagaDetected       SagaPhase = "REORG_DETECTED"
	SagaAnalysing      SagaPhase = "ANALYSING_IMPACT"
	SagaCompensating   SagaPhase = "COMPENSATING"
	SagaCompensated    SagaPhase = "COMPENSATED"
	SagaFailed         SagaPhase = "COMPENSATION_FAILED"
	SagaManualReview   SagaPhase = "MANUAL_REVIEW"
)

// ChainBlock represents an observed block on a chain.
type ChainBlock struct {
	ChainID     uint64 `json:"chain_id"`
	BlockNumber uint64 `json:"block_number"`
	BlockHash   string `json:"block_hash"`
	ParentHash  string `json:"parent_hash"`
	Timestamp   int64  `json:"timestamp"`
}

// ReorgEvent captures a detected chain reorganization.
type ReorgEvent struct {
	EventID       string        `json:"event_id"`
	ChainID       uint64        `json:"chain_id"`
	ForkBlock     uint64        `json:"fork_block"`
	OldHeadHash   string        `json:"old_head_hash"`
	NewHeadHash   string        `json:"new_head_hash"`
	Depth         uint64        `json:"depth"`
	Severity      ReorgSeverity `json:"severity"`
	DetectedAt    time.Time     `json:"detected_at"`
	AffectedTxIDs []string      `json:"affected_tx_ids"`
}

// CompensatingAction defines a single rollback or correction step.
type CompensatingAction struct {
	ActionID    string `json:"action_id"`
	Description string `json:"description"`
	TargetTxID  string `json:"target_tx_id"`
	ActionType  string `json:"action_type"` // "REVERSE_CREDIT", "RESUBMIT", "FREEZE_BALANCE"
	Executed    bool   `json:"executed"`
	Error       string `json:"error,omitempty"`
}

// Saga represents the full compensation saga for a reorg event.
type Saga struct {
	SagaID    string               `json:"saga_id"`
	ReorgID   string               `json:"reorg_id"`
	Phase     SagaPhase            `json:"phase"`
	Actions   []CompensatingAction `json:"actions"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Reorg Detector
// ──────────────────────────────────────────────────────────────────────────────

// ReorgDetector monitors block headers for chain reorganizations.
type ReorgDetector struct {
	mu             sync.RWMutex
	canonicalChain map[uint64]ChainBlock // blockNumber -> block
	chainID        uint64
	maxTrackDepth  uint64
}

// NewReorgDetector initialises a detector for a specific chain.
func NewReorgDetector(chainID uint64, maxDepth uint64) *ReorgDetector {
	return &ReorgDetector{
		canonicalChain: make(map[uint64]ChainBlock),
		chainID:        chainID,
		maxTrackDepth:  maxDepth,
	}
}

// ClassifySeverity determines the severity of a reorg by depth.
func ClassifySeverity(depth uint64) ReorgSeverity {
	if depth <= 2 {
		return ReorgMinor
	}
	if depth <= 6 {
		return ReorgModerate
	}
	return ReorgCritical
}

// IngestBlock processes a new block header. Returns a ReorgEvent if a
// reorganization is detected, or nil if the block extends the canonical chain.
func (rd *ReorgDetector) IngestBlock(block ChainBlock) (*ReorgEvent, error) {
	if block.ChainID != rd.chainID {
		return nil, fmt.Errorf("chain mismatch: expected %d, got %d", rd.chainID, block.ChainID)
	}

	rd.mu.Lock()
	defer rd.mu.Unlock()

	existing, hasExisting := rd.canonicalChain[block.BlockNumber]
	if hasExisting && existing.BlockHash != block.BlockHash {
		// Reorg detected: same block number, different hash
		depth := rd.calculateReorgDepth(block.BlockNumber)
		severity := ClassifySeverity(depth)

		payload := fmt.Sprintf("reorg:%d:%d:%s:%s", rd.chainID, block.BlockNumber, existing.BlockHash, block.BlockHash)
		h := sha256.Sum256([]byte(payload))
		eventID := "REORG-" + hex.EncodeToString(h[:12])

		event := &ReorgEvent{
			EventID:     eventID,
			ChainID:     rd.chainID,
			ForkBlock:   block.BlockNumber,
			OldHeadHash: existing.BlockHash,
			NewHeadHash: block.BlockHash,
			Depth:       depth,
			Severity:    severity,
			DetectedAt:  time.Now().UTC(),
		}

		// Replace canonical chain from fork point
		rd.canonicalChain[block.BlockNumber] = block

		return event, nil
	}

	// Normal block extension
	rd.canonicalChain[block.BlockNumber] = block

	// Prune old blocks beyond max track depth
	if block.BlockNumber > rd.maxTrackDepth {
		delete(rd.canonicalChain, block.BlockNumber-rd.maxTrackDepth-1)
	}

	return nil, nil
}

func (rd *ReorgDetector) calculateReorgDepth(forkBlock uint64) uint64 {
	var depth uint64 = 1
	for bn := forkBlock - 1; bn > 0 && depth < rd.maxTrackDepth; bn-- {
		if _, exists := rd.canonicalChain[bn]; !exists {
			break
		}
		depth++
	}
	return depth
}

// ──────────────────────────────────────────────────────────────────────────────
// Saga Coordinator
// ──────────────────────────────────────────────────────────────────────────────

// SagaCoordinator orchestrates compensating transactions in response to reorgs.
type SagaCoordinator struct {
	mu    sync.Mutex
	sagas map[string]*Saga
}

// NewSagaCoordinator creates a new coordinator instance.
func NewSagaCoordinator() *SagaCoordinator {
	return &SagaCoordinator{
		sagas: make(map[string]*Saga),
	}
}

// CreateSaga initialises a new compensation saga from a reorg event.
func (sc *SagaCoordinator) CreateSaga(event *ReorgEvent) (*Saga, error) {
	if event == nil {
		return nil, errors.New("reorg event is nil")
	}
	if event.EventID == "" {
		return nil, errors.New("reorg event missing ID")
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sagaID := "SAGA-" + event.EventID
	if _, exists := sc.sagas[sagaID]; exists {
		return nil, fmt.Errorf("saga %s already exists", sagaID)
	}

	// Build compensating actions for affected transactions
	var actions []CompensatingAction
	for i, txID := range event.AffectedTxIDs {
		actionType := "REVERSE_CREDIT"
		if event.Severity == ReorgCritical {
			actionType = "FREEZE_BALANCE"
		}
		actions = append(actions, CompensatingAction{
			ActionID:    fmt.Sprintf("%s-ACT-%d", sagaID, i),
			Description: fmt.Sprintf("Compensate tx %s due to reorg at block %d", txID, event.ForkBlock),
			TargetTxID:  txID,
			ActionType:  actionType,
			Executed:    false,
		})
	}

	now := time.Now().UTC()
	saga := &Saga{
		SagaID:    sagaID,
		ReorgID:   event.EventID,
		Phase:     SagaDetected,
		Actions:   actions,
		CreatedAt: now,
		UpdatedAt: now,
	}

	sc.sagas[sagaID] = saga

	fmt.Printf("[SagaCoordinator] Created %s with %d actions for reorg depth %d (%s)\n",
		sagaID, len(actions), event.Depth, event.Severity)

	return saga, nil
}

// ExecuteSaga runs all compensating actions in the saga sequentially.
func (sc *SagaCoordinator) ExecuteSaga(sagaID string) error {
	sc.mu.Lock()
	saga, ok := sc.sagas[sagaID]
	if !ok {
		sc.mu.Unlock()
		return fmt.Errorf("saga %s not found", sagaID)
	}

	if saga.Phase != SagaDetected && saga.Phase != SagaAnalysing {
		sc.mu.Unlock()
		return fmt.Errorf("saga %s in non-executable phase: %s", sagaID, saga.Phase)
	}

	saga.Phase = SagaCompensating
	saga.UpdatedAt = time.Now().UTC()
	sc.mu.Unlock()

	allSucceeded := true
	for i := range saga.Actions {
		// Simulate executing the compensating action
		saga.Actions[i].Executed = true
		fmt.Printf("[SagaCoordinator] Executed action %s (%s) for tx %s\n",
			saga.Actions[i].ActionID, saga.Actions[i].ActionType, saga.Actions[i].TargetTxID)
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if allSucceeded {
		saga.Phase = SagaCompensated
	} else {
		saga.Phase = SagaFailed
	}
	saga.UpdatedAt = time.Now().UTC()

	return nil
}

// GetSaga retrieves a saga by ID.
func (sc *SagaCoordinator) GetSaga(sagaID string) (*Saga, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	saga, ok := sc.sagas[sagaID]
	if !ok {
		return nil, fmt.Errorf("saga %s not found", sagaID)
	}
	return saga, nil
}

// EscalateToManualReview marks a saga for manual intervention.
func (sc *SagaCoordinator) EscalateToManualReview(sagaID string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	saga, ok := sc.sagas[sagaID]
	if !ok {
		return fmt.Errorf("saga %s not found", sagaID)
	}
	saga.Phase = SagaManualReview
	saga.UpdatedAt = time.Now().UTC()
	return nil
}
