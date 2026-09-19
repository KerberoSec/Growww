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
// Domain types for Chainlink CCIP ingestion
// ──────────────────────────────────────────────────────────────────────────────

// CCIPMessageType distinguishes transfer types.
type CCIPMessageType string

const (
	CCIPTokenTransfer    CCIPMessageType = "TOKEN_TRANSFER"
	CCIPArbitraryMessage CCIPMessageType = "ARBITRARY_MESSAGE"
	CCIPProgrammable     CCIPMessageType = "PROGRAMMABLE_TOKEN_TRANSFER"
)

// CCIPLane represents a CCIP lane between source and destination chains.
type CCIPLane struct {
	SourceChainSelector uint64 `json:"source_chain_selector"`
	DestChainSelector   uint64 `json:"dest_chain_selector"`
	OnRampAddress       string `json:"on_ramp_address"`
	OffRampAddress      string `json:"off_ramp_address"`
}

// CCIPMessage represents an inbound CCIP message from a source chain.
type CCIPMessage struct {
	MessageID       string          `json:"message_id"`
	SourceChain     uint64          `json:"source_chain"`
	DestChain       uint64          `json:"dest_chain"`
	SequenceNumber  uint64          `json:"sequence_number"`
	Sender          string          `json:"sender"`
	Receiver        string          `json:"receiver"`
	TokenAmounts    []TokenAmount   `json:"token_amounts"`
	Data            []byte          `json:"data"`
	MessageType     CCIPMessageType `json:"message_type"`
	FeeToken        string          `json:"fee_token"`
	FeeAmountE18    string          `json:"fee_amount_e18"`
	Nonce           uint64          `json:"nonce"`
	BlockNumber     uint64          `json:"block_number"`
	TxHash          string          `json:"tx_hash"`
	Timestamp       int64           `json:"timestamp"`
}

// TokenAmount represents a token transfer amount within a CCIP message.
type TokenAmount struct {
	Token    string `json:"token"`
	AmountE8 uint64 `json:"amount_e8"`
}

// VerificationStatus indicates the trust level of a verified message.
type VerificationStatus string

const (
	VerificationPending   VerificationStatus = "PENDING"
	VerificationConfirmed VerificationStatus = "CONFIRMED"
	VerificationRejected  VerificationStatus = "REJECTED"
	VerificationFinalized VerificationStatus = "FINALIZED"
)

// VerifiedCCIPEvent is a confirmed and indexed CCIP event.
type VerifiedCCIPEvent struct {
	EventID            string             `json:"event_id"`
	Message            CCIPMessage        `json:"message"`
	VerificationStatus VerificationStatus `json:"verification_status"`
	CommitRoot         string             `json:"commit_root"`
	VerifiedAt         time.Time          `json:"verified_at"`
	IndexedAt          time.Time          `json:"indexed_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// CCIP Ingress Service
// ──────────────────────────────────────────────────────────────────────────────

// CCIPIngressService ingests, verifies, and indexes Chainlink CCIP messages.
type CCIPIngressService struct {
	mu              sync.RWMutex
	supportedLanes  map[string]CCIPLane // key: "src-dst"
	events          map[string]*VerifiedCCIPEvent
	seqTracker      map[uint64]uint64 // sourceChain -> last sequence number
	minConfirmations uint64
}

// NewCCIPIngressService creates a new ingress service.
func NewCCIPIngressService(minConfirmations uint64) *CCIPIngressService {
	return &CCIPIngressService{
		supportedLanes:   make(map[string]CCIPLane),
		events:           make(map[string]*VerifiedCCIPEvent),
		seqTracker:       make(map[uint64]uint64),
		minConfirmations: minConfirmations,
	}
}

// RegisterLane adds a supported CCIP lane.
func (s *CCIPIngressService) RegisterLane(lane CCIPLane) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%d-%d", lane.SourceChainSelector, lane.DestChainSelector)
	s.supportedLanes[key] = lane
}

// laneKey returns a lookup key for a CCIP lane.
func laneKey(src, dst uint64) string {
	return fmt.Sprintf("%d-%d", src, dst)
}

// IngestMessage validates, verifies, and indexes an inbound CCIP message.
func (s *CCIPIngressService) IngestMessage(msg CCIPMessage) (*VerifiedCCIPEvent, error) {
	if msg.MessageID == "" {
		return nil, errors.New("message_id is required")
	}
	if msg.Sender == "" || msg.Receiver == "" {
		return nil, errors.New("sender and receiver are required")
	}
	if msg.SourceChain == msg.DestChain {
		return nil, errors.New("source and dest chain must differ")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check supported lane
	key := laneKey(msg.SourceChain, msg.DestChain)
	_, laneExists := s.supportedLanes[key]
	if !laneExists {
		return nil, fmt.Errorf("unsupported CCIP lane: %s", key)
	}

	// Check idempotency
	if _, exists := s.events[msg.MessageID]; exists {
		return nil, fmt.Errorf("duplicate message: %s", msg.MessageID)
	}

	// Verify sequence ordering
	lastSeq, hasSeq := s.seqTracker[msg.SourceChain]
	if hasSeq && msg.SequenceNumber <= lastSeq {
		return nil, fmt.Errorf("out-of-order sequence: got %d, expected > %d", msg.SequenceNumber, lastSeq)
	}
	s.seqTracker[msg.SourceChain] = msg.SequenceNumber

	// Compute commit root (simulated Merkle root)
	payload := fmt.Sprintf("%s:%d:%d:%s:%s:%d", msg.MessageID, msg.SourceChain, msg.DestChain, msg.Sender, msg.Receiver, msg.SequenceNumber)
	h := sha256.Sum256([]byte(payload))
	commitRoot := "0x" + hex.EncodeToString(h[:32])

	// Determine verification status based on confirmations
	status := VerificationPending
	if s.minConfirmations == 0 {
		status = VerificationConfirmed
	}

	now := time.Now().UTC()
	event := &VerifiedCCIPEvent{
		EventID:            "CCIP-EVT-" + msg.MessageID,
		Message:            msg,
		VerificationStatus: status,
		CommitRoot:         commitRoot,
		VerifiedAt:         now,
		IndexedAt:          now,
	}

	s.events[msg.MessageID] = event

	fmt.Printf("[CCIP Ingress] Indexed %s | Lane: %s | Type: %s | Status: %s\n",
		event.EventID, key, msg.MessageType, status)

	return event, nil
}

// ConfirmEvent finalizes a pending event after sufficient on-chain confirmations.
func (s *CCIPIngressService) ConfirmEvent(messageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[messageID]
	if !ok {
		return fmt.Errorf("event not found: %s", messageID)
	}
	if event.VerificationStatus == VerificationRejected {
		return errors.New("cannot confirm a rejected event")
	}
	event.VerificationStatus = VerificationFinalized
	return nil
}

// RejectEvent marks a CCIP event as invalid (e.g., fraudulent proof).
func (s *CCIPIngressService) RejectEvent(messageID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event, ok := s.events[messageID]
	if !ok {
		return fmt.Errorf("event not found: %s", messageID)
	}
	event.VerificationStatus = VerificationRejected
	return nil
}

// GetEvent retrieves an indexed CCIP event.
func (s *CCIPIngressService) GetEvent(messageID string) (*VerifiedCCIPEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.events[messageID]
	if !ok {
		return nil, fmt.Errorf("event not found: %s", messageID)
	}
	return event, nil
}

// GetEventsByChain returns all events originating from a given source chain.
func (s *CCIPIngressService) GetEventsByChain(sourceChain uint64) []*VerifiedCCIPEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*VerifiedCCIPEvent
	for _, evt := range s.events {
		if evt.Message.SourceChain == sourceChain {
			results = append(results, evt)
		}
	}
	return results
}
