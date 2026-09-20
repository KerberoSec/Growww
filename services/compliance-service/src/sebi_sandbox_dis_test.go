package main

import (
	"errors"
	"testing"
	"time"
)

func TestSEBISandbox_ParticipantCeiling(t *testing.T) {
	mgr := NewSEBISandboxManager()
	mgr.maxRetailUsers = 3 // Set low ceiling for unit testing

	if err := mgr.RegisterParticipant("U1"); err != nil {
		t.Fatalf("unexpected error registering U1: %v", err)
	}
	if err := mgr.RegisterParticipant("U2"); err != nil {
		t.Fatalf("unexpected error registering U2: %v", err)
	}
	if err := mgr.RegisterParticipant("U3"); err != nil {
		t.Fatalf("unexpected error registering U3: %v", err)
	}

	// Idempotent re-registration of existing user succeeds
	if err := mgr.RegisterParticipant("U1"); err != nil {
		t.Fatalf("re-registering U1 should be idempotent, got %v", err)
	}

	// 4th user exceeds ceiling
	err := mgr.RegisterParticipant("U4")
	if !errors.Is(err, ErrParticipantLimitExceeded) {
		t.Fatalf("expected ErrParticipantLimitExceeded, got %v", err)
	}
}

func TestSEBISandbox_ElectronicDISLifecycle(t *testing.T) {
	mgr := NewSEBISandboxManager()

	boid := "1208160012345678" // 16-character BOID
	tpin := "842109"

	slip, err := mgr.CreateElectronicDIS(
		"DIS-101", "USER-ALPHA",
		DepositoryCDSL, boid, "INE002A01018", 100, 250000.0, tpin,
	)
	if err != nil {
		t.Fatalf("unexpected error creating e-DIS: %v", err)
	}
	if slip.Status != DISStatusInitiated {
		t.Fatalf("expected status INITIATED, got %s", slip.Status)
	}

	// Invalid TPIN should reject
	err = mgr.VerifyTPINAndSettle("DIS-101", "000000")
	if !errors.Is(err, ErrInvalidTPIN) {
		t.Fatalf("expected ErrInvalidTPIN, got %v", err)
	}

	// Re-attempt after rejection or re-create
	slip2, err := mgr.CreateElectronicDIS(
		"DIS-102", "USER-ALPHA",
		DepositoryNSDL, boid, "INE002A01018", 50, 125000.0, tpin,
	)
	if err != nil {
		t.Fatalf("unexpected error creating DIS-102: %v", err)
	}

	// Correct TPIN settles
	err = mgr.VerifyTPINAndSettle("DIS-102", tpin)
	if err != nil {
		t.Fatalf("unexpected error settling with correct TPIN: %v", err)
	}
	if slip2.Status != DISStatusSettled || slip2.SettledAt == nil {
		t.Fatalf("expected status SETTLED, got %s", slip2.Status)
	}
}

func TestSEBISandbox_ExposureLimit(t *testing.T) {
	mgr := NewSEBISandboxManager()
	mgr.maxAggregateExpINR = 100000.0 // ₹1 Lakh ceiling for test

	boid := "1208160012345678"
	tpin := "123456"

	// Request exceeding ceiling
	_, err := mgr.CreateElectronicDIS("D1", "U1", DepositoryCDSL, boid, "INE001A01011", 10, 150000.0, tpin)
	if !errors.Is(err, ErrExposureLimitExceeded) {
		t.Fatalf("expected ErrExposureLimitExceeded, got %v", err)
	}
}

func TestSEBISandbox_ValidationErrors(t *testing.T) {
	mgr := NewSEBISandboxManager()

	// Invalid BOID length
	_, err := mgr.CreateElectronicDIS("D1", "U1", DepositoryCDSL, "SHORT", "INE001", 10, 1000.0, "123456")
	if !errors.Is(err, ErrInvalidBOID) {
		t.Fatalf("expected ErrInvalidBOID, got %v", err)
	}

	// Invalid TPIN length
	_, err = mgr.CreateElectronicDIS("D1", "U1", DepositoryCDSL, "1208160012345678", "INE001", 10, 1000.0, "12")
	if !errors.Is(err, ErrInvalidTPIN) {
		t.Fatalf("expected ErrInvalidTPIN, got %v", err)
	}
}

func TestSEBISandbox_MonthlyMilestoneReport(t *testing.T) {
	mgr := NewSEBISandboxManager()
	boid := "1208160012345678"
	tpin := "998877"

	_, _ = mgr.CreateElectronicDIS("DIS-01", "USER-1", DepositoryCDSL, boid, "INE002A01018", 10, 25000.0, tpin)
	_ = mgr.VerifyTPINAndSettle("DIS-01", tpin)

	mgr.LogGrievance(true)  // 1 resolved grievance
	mgr.LogGrievance(false) // 1 pending grievance

	report := mgr.GenerateMonthlyMilestoneReport("2026-09")
	if report.PeriodMonth != "2026-09" {
		t.Fatalf("expected period 2026-09, got %s", report.PeriodMonth)
	}
	if report.TotalTradeCount != 1 {
		t.Fatalf("expected 1 settled trade, got %d", report.TotalTradeCount)
	}
	if report.InvestorGrievances != 2 || report.ResolvedGrievances != 1 {
		t.Fatalf("expected 2 grievances (1 resolved), got %d and %d",
			report.InvestorGrievances, report.ResolvedGrievances)
	}
	if report.SettlementMerkleRoot == "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Fatal("expected non-empty settlement Merkle root")
	}
}

func TestSEBISandbox_DISExpiry(t *testing.T) {
	mgr := NewSEBISandboxManager()
	boid := "1208160012345678"
	tpin := "112233"

	slip, _ := mgr.CreateElectronicDIS("DIS-EXP", "USER-1", DepositoryCDSL, boid, "INE002A01018", 10, 25000.0, tpin)

	// Backdate expiry
	slip.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute)

	err := mgr.VerifyTPINAndSettle("DIS-EXP", tpin)
	if !errors.Is(err, ErrDISAlreadyProcessed) {
		t.Fatalf("expected ErrDISAlreadyProcessed on expired DIS, got %v", err)
	}
	if slip.Status != DISStatusExpired {
		t.Fatalf("expected status EXPIRED, got %s", slip.Status)
	}
}
