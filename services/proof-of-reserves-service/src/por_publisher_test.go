package main

import (
	"crypto/rand"
	"testing"
	"time"
)

func generateRandomBytes() [32]byte {
	var b [32]byte
	rand.Read(b[:])
	return b
}

func TestPoR_MerkleTreeAndInclusionProof(t *testing.T) {
	holdings := []InvestorHolding{
		{InvestorCommitment: generateRandomBytes(), Salt: generateRandomBytes(), Balance: 100},
		{InvestorCommitment: generateRandomBytes(), Salt: generateRandomBytes(), Balance: 250},
		{InvestorCommitment: generateRandomBytes(), Salt: generateRandomBytes(), Balance: 500},
		{InvestorCommitment: generateRandomBytes(), Salt: generateRandomBytes(), Balance: 1000},
	}

	root, sortedLeaves := BuildMerkleTree(holdings)
	if root == [32]byte{} {
		t.Fatal("Expected non-empty Merkle root")
	}

	// Verify proof generation for holding[2]
	target := holdings[2]
	proof, err := GenerateProof(target, sortedLeaves)
	if err != nil {
		t.Fatalf("Failed to generate inclusion proof: %v", err)
	}

	isValid := VerifyInclusion(target, proof, root)
	if !isValid {
		t.Errorf("Expected valid inclusion proof for target holding")
	}

	// Corrupted balance should fail
	corrupted := target
	corrupted.Balance = 9999
	if VerifyInclusion(corrupted, proof, root) {
		t.Errorf("Expected inclusion proof to fail for corrupted balance")
	}
}

func TestPoR_SolvencyReconciliation(t *testing.T) {
	engine := NewPoREngine()

	custody := DepositoryReport{
		ISIN:                 "INE002A01018",
		DepositoryShareCount: 50000,
		ReportTimestamp:      time.Now(),
		CustodianID:          "NSDL-CUST-001",
	}

	onChain := OnChainSupplyReport{
		ISIN:               "INE002A01018",
		OnChainTokenSupply: 50000,
		ContractAddress:    "0x1234567890123456789012345678901234567890",
		BlockNumber:        123456,
		Timestamp:          time.Now(),
	}

	holdings := []InvestorHolding{
		{InvestorCommitment: generateRandomBytes(), Salt: generateRandomBytes(), Balance: 20000},
		{InvestorCommitment: generateRandomBytes(), Salt: generateRandomBytes(), Balance: 30000},
	}

	record, err := engine.ReconcileAndPublish(custody, onChain, holdings)
	if err != nil {
		t.Fatalf("Reconciliation failed: %v", err)
	}

	if record.Status != Solvent {
		t.Errorf("Expected SOLVENT, got %v", record.Status)
	}
	if record.DeficitAmount != 0 {
		t.Errorf("Expected zero deficit, got %d", record.DeficitAmount)
	}

	// Now simulate fractional reserve deficit breach
	onChainDeficit := onChain
	onChainDeficit.OnChainTokenSupply = 55000 // 5k deficit

	record2, err := engine.ReconcileAndPublish(custody, onChainDeficit, holdings)
	if err != nil {
		t.Fatalf("Reconciliation failed: %v", err)
	}

	if record2.Status != Deficit {
		t.Errorf("Expected DEFICIT_BREACH, got %v", record2.Status)
	}
	if record2.DeficitAmount != 5000 {
		t.Errorf("Expected deficit amount 5000, got %d", record2.DeficitAmount)
	}
}
