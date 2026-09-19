package main

import (
	"strings"
	"testing"
)

func TestReorgDetectorNormalBlocks(t *testing.T) {
	det := NewReorgDetector(1, 100)

	for i := uint64(1); i <= 10; i++ {
		event, err := det.IngestBlock(ChainBlock{
			ChainID:     1,
			BlockNumber: i,
			BlockHash:   "0xhash" + string(rune('a'+i)),
			ParentHash:  "0xhash" + string(rune('a'+i-1)),
			Timestamp:   1000 + int64(i),
		})
		if err != nil {
			t.Fatalf("block %d: unexpected error: %v", i, err)
		}
		if event != nil {
			t.Fatalf("block %d: unexpected reorg event", i)
		}
	}
}

func TestReorgDetectorDetectsReorg(t *testing.T) {
	det := NewReorgDetector(1, 100)

	// Ingest block 5
	_, _ = det.IngestBlock(ChainBlock{
		ChainID: 1, BlockNumber: 5, BlockHash: "0xOLD5", ParentHash: "0x4",
	})

	// Ingest a different block 5 -> reorg
	event, err := det.IngestBlock(ChainBlock{
		ChainID: 1, BlockNumber: 5, BlockHash: "0xNEW5", ParentHash: "0x4alt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event == nil {
		t.Fatal("expected reorg event, got nil")
	}
	if event.OldHeadHash != "0xOLD5" || event.NewHeadHash != "0xNEW5" {
		t.Errorf("wrong block hashes: old=%s new=%s", event.OldHeadHash, event.NewHeadHash)
	}
	if !strings.HasPrefix(event.EventID, "REORG-") {
		t.Errorf("event ID missing prefix: %s", event.EventID)
	}
}

func TestClassifySeverity(t *testing.T) {
	cases := []struct {
		depth    uint64
		expected ReorgSeverity
	}{
		{1, ReorgMinor},
		{2, ReorgMinor},
		{3, ReorgModerate},
		{6, ReorgModerate},
		{7, ReorgCritical},
		{50, ReorgCritical},
	}

	for _, tc := range cases {
		got := ClassifySeverity(tc.depth)
		if got != tc.expected {
			t.Errorf("depth %d: expected %s, got %s", tc.depth, tc.expected, got)
		}
	}
}

func TestSagaCoordinatorCreateAndExecute(t *testing.T) {
	coord := NewSagaCoordinator()

	event := &ReorgEvent{
		EventID:       "REORG-abc123",
		ChainID:       1,
		ForkBlock:     100,
		OldHeadHash:   "0xOLD",
		NewHeadHash:   "0xNEW",
		Depth:         3,
		Severity:      ReorgModerate,
		AffectedTxIDs: []string{"TX-001", "TX-002", "TX-003"},
	}

	saga, err := coord.CreateSaga(event)
	if err != nil {
		t.Fatalf("create saga failed: %v", err)
	}
	if saga.Phase != SagaDetected {
		t.Errorf("expected REORG_DETECTED phase, got %s", saga.Phase)
	}
	if len(saga.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(saga.Actions))
	}

	// Execute saga
	err = coord.ExecuteSaga(saga.SagaID)
	if err != nil {
		t.Fatalf("execute saga failed: %v", err)
	}

	updated, _ := coord.GetSaga(saga.SagaID)
	if updated.Phase != SagaCompensated {
		t.Errorf("expected COMPENSATED, got %s", updated.Phase)
	}
	for _, a := range updated.Actions {
		if !a.Executed {
			t.Errorf("action %s not executed", a.ActionID)
		}
	}
}

func TestSagaDuplicateRejected(t *testing.T) {
	coord := NewSagaCoordinator()
	event := &ReorgEvent{EventID: "REORG-dup", ForkBlock: 50, AffectedTxIDs: []string{"TX-1"}}

	_, _ = coord.CreateSaga(event)
	_, err := coord.CreateSaga(event)
	if err == nil {
		t.Fatal("expected duplicate saga error")
	}
}

func TestSagaEscalation(t *testing.T) {
	coord := NewSagaCoordinator()
	event := &ReorgEvent{
		EventID:       "REORG-esc",
		Severity:      ReorgCritical,
		ForkBlock:     200,
		AffectedTxIDs: []string{"TX-X"},
	}

	saga, _ := coord.CreateSaga(event)
	err := coord.EscalateToManualReview(saga.SagaID)
	if err != nil {
		t.Fatalf("escalation failed: %v", err)
	}

	s, _ := coord.GetSaga(saga.SagaID)
	if s.Phase != SagaManualReview {
		t.Errorf("expected MANUAL_REVIEW, got %s", s.Phase)
	}
}

func TestReorgDetectorChainMismatch(t *testing.T) {
	det := NewReorgDetector(1, 100)
	_, err := det.IngestBlock(ChainBlock{ChainID: 137, BlockNumber: 1, BlockHash: "0x1"})
	if err == nil {
		t.Fatal("expected chain mismatch error")
	}
}
