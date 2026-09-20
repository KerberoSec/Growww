package src

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

type MutationOp string

const (
	OpInsert MutationOp = "INSERT"
	OpUpdate MutationOp = "UPDATE"
	OpDelete MutationOp = "DELETE"
)

// CDCMutation captures row-level change events from PostgreSQL WAL
type CDCMutation struct {
	MutationID   string                 `json:"mutation_id"`
	SequenceNo   uint64                 `json:"sequence_no"`
	TableName    string                 `json:"table_name"`
	Operation    MutationOp             `json:"operation"`
	RowKey       string                 `json:"row_key"`
	BeforeData   map[string]interface{} `json:"before_data,omitempty"`
	AfterData    map[string]interface{} `json:"after_data,omitempty"`
	TimestampNanos int64                `json:"timestamp_nanos"`
	PreviousHash string                 `json:"previous_hash"`
	CurrentHash  string                 `json:"current_hash"`
}

// PITRPipeline manages CDC ingestion, tamper-evident hash chaining, and replay
type PITRPipeline struct {
	mu           sync.RWMutex
	mutations    []*CDCMutation
	tableSnapshots map[string]map[string]map[string]interface{} // table -> rowKey -> data
	lastSeqNo    uint64
	lastHash     string
}

func NewPITRPipeline() *PITRPipeline {
	return &PITRPipeline{
		mutations:      make([]*CDCMutation, 0),
		tableSnapshots: make(map[string]map[string]map[string]interface{}),
		lastSeqNo:      0,
		lastHash:       "GENESIS_CDC_ROOT_00000000000000000000000000000000",
	}
}

// IngestMutation appends and cryptographically chains a new database mutation
func (p *PITRPipeline) IngestMutation(table string, op MutationOp, key string, before, after map[string]interface{}, tsNanos int64) (*CDCMutation, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if table == "" || key == "" {
		return nil, errors.New("table and row key are required")
	}

	p.lastSeqNo++
	mutID := fmt.Sprintf("mut_%s_%d", table, p.lastSeqNo)

	// Compute tamper-evident SHA256 chain hash including data payload
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%d:%s:%s:%s:%d:%v:%v:%s", mutID, p.lastSeqNo, table, op, key, tsNanos, before, after, p.lastHash)))
	currHash := hex.EncodeToString(h.Sum(nil))

	mut := &CDCMutation{
		MutationID:     mutID,
		SequenceNo:     p.lastSeqNo,
		TableName:      table,
		Operation:      op,
		RowKey:         key,
		BeforeData:     before,
		AfterData:      after,
		TimestampNanos: tsNanos,
		PreviousHash:   p.lastHash,
		CurrentHash:    currHash,
	}

	p.mutations = append(p.mutations, mut)
	p.lastHash = currHash

	// Apply mutation directly to in-memory state
	p.applyMutationToState(mut)

	return mut, nil
}

func (p *PITRPipeline) applyMutationToState(mut *CDCMutation) {
	if _, exists := p.tableSnapshots[mut.TableName]; !exists {
		p.tableSnapshots[mut.TableName] = make(map[string]map[string]interface{})
	}

	table := p.tableSnapshots[mut.TableName]
	switch mut.Operation {
	case OpInsert, OpUpdate:
		table[mut.RowKey] = mut.AfterData
	case OpDelete:
		delete(table, mut.RowKey)
	}
}

// RestoreToPointInTime rolls back or reconstructs state up to exact target timestamp
func (p *PITRPipeline) RestoreToPointInTime(targetTime time.Time) (map[string]map[string]map[string]interface{}, int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	targetNanos := targetTime.UnixNano()
	reconstructed := make(map[string]map[string]map[string]interface{})
	replayedCount := 0

	// Verify cryptographic hash chain integrity before recovery
	prevHash := "GENESIS_CDC_ROOT_00000000000000000000000000000000"
	for _, mut := range p.mutations {
		if mut.PreviousHash != prevHash {
			return nil, 0, fmt.Errorf("cryptographic chain break detected at mutation %s", mut.MutationID)
		}

		h := sha256.New()
		h.Write([]byte(fmt.Sprintf("%s:%d:%s:%s:%s:%d:%v:%v:%s", mut.MutationID, mut.SequenceNo, mut.TableName, mut.Operation, mut.RowKey, mut.TimestampNanos, mut.BeforeData, mut.AfterData, mut.PreviousHash)))
		expectedHash := hex.EncodeToString(h.Sum(nil))
		if mut.CurrentHash != expectedHash {
			return nil, 0, fmt.Errorf("tamper detected at mutation %s", mut.MutationID)
		}
		prevHash = mut.CurrentHash

		// Stop applying mutations after target point in time
		if mut.TimestampNanos > targetNanos {
			continue
		}

		if _, exists := reconstructed[mut.TableName]; !exists {
			reconstructed[mut.TableName] = make(map[string]map[string]interface{})
		}
		table := reconstructed[mut.TableName]

		switch mut.Operation {
		case OpInsert, OpUpdate:
			table[mut.RowKey] = mut.AfterData
		case OpDelete:
			delete(table, mut.RowKey)
		}
		replayedCount++
	}

	return reconstructed, replayedCount, nil
}

// GetCurrentRecord fetches current active state for a table and row
func (p *PITRPipeline) GetCurrentRecord(table, key string) (map[string]interface{}, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	t, ok := p.tableSnapshots[table]
	if !ok {
		return nil, false
	}
	row, exists := t[key]
	return row, exists
}
