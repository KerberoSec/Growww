package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type AuditLogRecord struct {
	Index          uint64    `json:"index"`
	PreviousHash   string    `json:"previous_hash"`
	RecordHash     string    `json:"record_hash"`
	EventCategory  string    `json:"event_category"` // AUTH, ORDER, TRADE, SETTLEMENT, REGULATORY
	ActorID        string    `json:"actor_id"`       // Pseudonymous actor ID
	Action         string    `json:"action"`
	PayloadDigest  string    `json:"payload_digest"`
	Timestamp      time.Time `json:"timestamp"`
}

type ImmutableWORMStorage struct {
	mu       sync.Mutex
	records  []AuditLogRecord
	lastHash string
}

func NewImmutableWORMStorage() *ImmutableWORMStorage {
	return &ImmutableWORMStorage{
		records:  make([]AuditLogRecord, 0),
		lastHash: "0000000000000000000000000000000000000000000000000000000000000000",
	}
}

// AppendAuditLog appends an event to the tamper-evident hash-chained WORM ledger
func (s *ImmutableWORMStorage) AppendAuditLog(category, actorID, action, rawPayload string) AuditLogRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := uint64(len(s.records)) + 1
	now := time.Now().UTC()

	payloadHash := sha256.Sum256([]byte(rawPayload))
	payloadDigest := hex.EncodeToString(payloadHash[:])

	recordData := fmt.Sprintf("%d:%s:%s:%s:%s:%s:%d",
		idx, s.lastHash, category, actorID, action, payloadDigest, now.UnixNano())

	recHash := sha256.Sum256([]byte(recordData))
	recHashStr := hex.EncodeToString(recHash[:])

	entry := AuditLogRecord{
		Index:         idx,
		PreviousHash:  s.lastHash,
		RecordHash:    recHashStr,
		EventCategory: category,
		ActorID:       actorID,
		Action:        action,
		PayloadDigest: payloadDigest,
		Timestamp:     now,
	}

	s.records = append(s.records, entry)
	s.lastHash = recHashStr

	return entry
}
