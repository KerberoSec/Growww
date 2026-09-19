package main

import (
	"testing"
)

func TestUTXOListener_ProcessTransaction(t *testing.T) {
	listener := NewUTXOListener(3)
	listener.RegisterDepositAddress("bc1qtest123", "user-001")

	dep := listener.ProcessTransaction(
		"txid-abc123", 0, "bc1qtest123", 50000000, 800000, "blockhash-001",
	)
	if dep == nil {
		t.Fatal("Expected deposit to be detected")
	}
	if dep.AmountSats != 50000000 {
		t.Fatalf("Expected 50000000 sats, got %d", dep.AmountSats)
	}
	if dep.Status != StatusConfirming {
		t.Fatalf("Expected CONFIRMING status, got %s", dep.Status)
	}
}

func TestUTXOListener_ConfirmationTransitions(t *testing.T) {
	listener := NewUTXOListener(3)
	listener.RegisterDepositAddress("bc1pTaproot", "user-002")

	dep := listener.ProcessTransaction(
		"txid-xyz789", 1, "bc1pTaproot", 100000000, 800000, "blockhash-002",
	)
	if dep == nil {
		t.Fatal("Expected deposit")
	}

	// Simulate block progression
	listener.UpdateConfirmations("txid-xyz789:1", 800002) // 3 confs -> CREDITED
	dep2, _ := listener.deposits["txid-xyz789:1"]
	if dep2.Status != StatusCredited {
		t.Fatalf("Expected CREDITED at 3 confs, got %s", dep2.Status)
	}

	listener.UpdateConfirmations("txid-xyz789:1", 800005) // 6 confs -> FINALIZED
	dep3, _ := listener.deposits["txid-xyz789:1"]
	if dep3.Status != StatusFinalized {
		t.Fatalf("Expected FINALIZED at 6 confs, got %s", dep3.Status)
	}
}

func TestUTXOListener_UnmonitoredAddress(t *testing.T) {
	listener := NewUTXOListener(3)

	dep := listener.ProcessTransaction(
		"txid-nomatch", 0, "bc1qunknown", 10000, 800000, "blockhash-003",
	)
	if dep != nil {
		t.Fatal("Expected nil for unmonitored address")
	}
}

func TestVerifySPV_ValidProof(t *testing.T) {
	// Test with a trivial single-leaf proof (txid == merkle root when no intermediates)
	proof := SPVProof{
		TxID:               "0000000000000000000000000000000000000000000000000000000000000000",
		MerkleRoot:         "0000000000000000000000000000000000000000000000000000000000000000",
		BlockHeight:        800000,
		IntermediateHashes: []string{},
		TxIndex:            0,
	}

	// With no intermediate hashes, the double-reversed txid should equal merkle root
	// This tests the reversal logic; for a zero hash, reversals are identity
	result := VerifySPV(proof)
	if !result {
		t.Fatal("Expected valid SPV proof for zero-hash identity case")
	}
}
