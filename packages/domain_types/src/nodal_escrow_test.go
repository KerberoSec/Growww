package src

import (
	"errors"
	"testing"
)

func TestNodalEscrowMakerCheckerAndReconciliation(t *testing.T) {
	// 1,00,00,000 INR (1 Cr) = 1,00,00,00,000 Paise. Statutory haircut = 200 bps (2%).
	initialBalancePaise := uint64(1000000000)
	escrow := NewNodalEscrowManager(initialBalancePaise, 200)

	transfers := []PayoutTransfer{
		{
			TransferID:      "TX-001",
			BeneficiaryPAN:  "ABCDE1234F",
			BankAccountNum:  "919876543210",
			IFSC:            "HDFC0001234",
			AmountINR_Paise: 50000000, // 5 Lakhs INR
		},
		{
			TransferID:      "TX-002",
			BeneficiaryPAN:  "WXYZP9876Q",
			BankAccountNum:  "919876543211",
			IFSC:            "ICIC0005678",
			AmountINR_Paise: 30000000, // 3 Lakhs INR
		},
	}

	// 1. Maker creates batch
	batchID := "BATCH-2026-09-20-001"
	makerID := "ops-maker-arun"
	batch, err := escrow.CreatePayoutBatch(batchID, makerID, transfers)
	if err != nil {
		t.Fatalf("failed creating payout batch: %v", err)
	}

	if batch.TotalAmountPaise != 80000000 {
		t.Fatalf("total amount mismatch: expected 80000000, got %d", batch.TotalAmountPaise)
	}

	// 2. Maker cannot approve their own batch (Dual-control enforcement)
	err = escrow.ApprovePayoutBatch(batchID, makerID)
	if err == nil || !errors.Is(err, ErrMakerCheckerConflict) {
		t.Fatalf("expected ErrMakerCheckerConflict when maker attempts self-approval, got %v", err)
	}

	// 3. Independent checker approves batch
	checkerID := "ops-checker-ritik"
	if err := escrow.ApprovePayoutBatch(batchID, checkerID); err != nil {
		t.Fatalf("checker approval failed: %v", err)
	}

	// 4. Execute batch
	if err := escrow.ExecuteBatch(batchID); err != nil {
		t.Fatalf("batch execution failed: %v", err)
	}

	expectedRemainingBalance := initialBalancePaise - 80000000
	// 5. Reconcile with bank report
	matched, _, err := escrow.ReconcileBankReport(expectedRemainingBalance)
	if err != nil || !matched {
		t.Fatalf("reconciliation failed against expected remaining balance: %v", err)
	}

	// 6. Reconciliation mismatch detection
	matchedMismatch, diff, err := escrow.ReconcileBankReport(expectedRemainingBalance - 50000)
	if matchedMismatch || err == nil || diff != 50000 {
		t.Fatalf("expected mismatch of 50000 paise, got matched=%v diff=%d err=%v", matchedMismatch, diff, err)
	}
}
