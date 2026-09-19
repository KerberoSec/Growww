package main

import (
	"testing"
	"time"
)

func TestSubmitApplication_Success(t *testing.T) {
	pipeline := NewKYCPipeline(365)

	app, err := pipeline.SubmitApplication("KYC-001", "USER-1", "Rahul Sharma", "ABCDE1234F", "hash123", "1990-01-01", TierBasic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Status != KYCStatusPending {
		t.Errorf("expected PENDING, got %s", app.Status)
	}
	if app.Tier != TierBasic {
		t.Errorf("expected BASIC tier, got %s", app.Tier)
	}
}

func TestSubmitApplication_InvalidPAN(t *testing.T) {
	pipeline := NewKYCPipeline(365)

	_, err := pipeline.SubmitApplication("KYC-002", "USER-2", "Test User", "SHORT", "hash", "2000-01-01", TierBasic)
	if err == nil {
		t.Fatal("expected error for invalid PAN")
	}
}

func TestSubmitApplication_Duplicate(t *testing.T) {
	pipeline := NewKYCPipeline(365)

	_, _ = pipeline.SubmitApplication("KYC-003", "USER-3", "Test User", "ABCDE1234F", "hash", "2000-01-01", TierBasic)
	_, err := pipeline.SubmitApplication("KYC-003", "USER-3", "Test User", "ABCDE1234F", "hash", "2000-01-01", TierBasic)
	if err == nil {
		t.Fatal("expected error for duplicate application")
	}
}

func TestScreenApplicant_SanctionedMatch(t *testing.T) {
	pipeline := NewKYCPipeline(365)
	pipeline.LoadSanctionsList([]string{"Dawood Ibrahim", "Hafiz Saeed"})

	_, _ = pipeline.SubmitApplication("KYC-004", "USER-4", "Dawood Ibrahim", "XYZAB9876P", "hash", "1955-12-26", TierBasic)

	result, err := pipeline.ScreenApplicant("KYC-004")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsSanctioned {
		t.Error("expected sanctioned flag to be true")
	}
	if result.MatchedList != "OFAC_SDN" {
		t.Errorf("expected OFAC_SDN match, got %s", result.MatchedList)
	}
}

func TestScreenApplicant_PEPMatch(t *testing.T) {
	pipeline := NewKYCPipeline(365)
	pipeline.LoadPEPList([]string{"Minister Kumar"})

	_, _ = pipeline.SubmitApplication("KYC-005", "USER-5", "Minister Kumar", "PQRST1234Z", "hash", "1970-05-15", TierStandard)

	result, err := pipeline.ScreenApplicant("KYC-005")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsPEP {
		t.Error("expected PEP flag to be true")
	}
}

func TestScreenApplicant_CleanUser(t *testing.T) {
	pipeline := NewKYCPipeline(365)
	pipeline.LoadSanctionsList([]string{"Bad Actor"})
	pipeline.LoadPEPList([]string{"Politician X"})

	_, _ = pipeline.SubmitApplication("KYC-006", "USER-6", "Clean User", "ABCDE5678K", "hash", "1995-06-01", TierBasic)

	result, err := pipeline.ScreenApplicant("KYC-006")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsSanctioned || result.IsPEP {
		t.Error("expected clean screening result")
	}
}

func TestCalculateRiskScore_LowRisk(t *testing.T) {
	pipeline := NewKYCPipeline(365)

	_, _ = pipeline.SubmitApplication("KYC-007", "USER-7", "Low Risk User", "ABCDE1234F", "hash", "1990-01-01", TierBasic)

	score, level, err := pipeline.CalculateRiskScore("KYC-007", 100000, 5, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score >= 25 {
		t.Errorf("expected low risk score < 25, got %d", score)
	}
	if level != RiskLow {
		t.Errorf("expected LOW risk level, got %s", level)
	}
}

func TestCalculateRiskScore_HighVolumePEP(t *testing.T) {
	pipeline := NewKYCPipeline(365)
	pipeline.LoadPEPList([]string{"High Risk Person"})

	_, _ = pipeline.SubmitApplication("KYC-008", "USER-8", "High Risk Person", "XYZAB1234P", "hash", "1975-03-10", TierAdvanced)
	_, _ = pipeline.ScreenApplicant("KYC-008")

	score, level, err := pipeline.CalculateRiskScore("KYC-008", 10000000, 30, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// PEP(30) + highVol(20) + country(30) + NRI(15) = 95
	if score != 95 {
		t.Errorf("expected risk score 95, got %d", score)
	}
	if level != RiskCritical {
		t.Errorf("expected CRITICAL risk level, got %s", level)
	}
}

func TestCalculateRiskScore_SanctionedAutoMax(t *testing.T) {
	pipeline := NewKYCPipeline(365)
	pipeline.LoadSanctionsList([]string{"Sanctioned Entity"})

	_, _ = pipeline.SubmitApplication("KYC-009", "USER-9", "Sanctioned Entity", "SANCT1234X", "hash", "1960-01-01", TierBasic)
	_, _ = pipeline.ScreenApplicant("KYC-009")

	score, level, err := pipeline.CalculateRiskScore("KYC-009", 0, 0, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score != 100 {
		t.Errorf("expected max risk score 100, got %d", score)
	}
	if level != RiskCritical {
		t.Errorf("expected CRITICAL risk level, got %s", level)
	}
}

func TestApproveApplication_Success(t *testing.T) {
	pipeline := NewKYCPipeline(365)

	_, _ = pipeline.SubmitApplication("KYC-010", "USER-10", "Normal User", "ABCDE1234F", "hash", "1992-07-20", TierStandard)

	app, err := pipeline.ApproveApplication("KYC-010")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Status != KYCStatusApproved {
		t.Errorf("expected APPROVED, got %s", app.Status)
	}
	if app.ReKYCDueDate.IsZero() {
		t.Error("expected re-KYC due date to be set")
	}
}

func TestApproveApplication_SanctionedBlocked(t *testing.T) {
	pipeline := NewKYCPipeline(365)
	pipeline.LoadSanctionsList([]string{"Blocked Person"})

	_, _ = pipeline.SubmitApplication("KYC-011", "USER-11", "Blocked Person", "BLOCK1234X", "hash", "1965-04-10", TierBasic)
	_, _ = pipeline.ScreenApplicant("KYC-011")

	_, err := pipeline.ApproveApplication("KYC-011")
	if err == nil {
		t.Fatal("expected error when approving sanctioned applicant")
	}
}

func TestRejectApplication(t *testing.T) {
	pipeline := NewKYCPipeline(365)

	_, _ = pipeline.SubmitApplication("KYC-012", "USER-12", "Reject User", "REJCT1234X", "hash", "2001-01-01", TierBasic)

	app, err := pipeline.RejectApplication("KYC-012", "Document mismatch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Status != KYCStatusRejected {
		t.Errorf("expected REJECTED, got %s", app.Status)
	}
}

func TestCheckReKYCDue(t *testing.T) {
	pipeline := NewKYCPipeline(365)

	_, _ = pipeline.SubmitApplication("KYC-013", "USER-13", "Old User", "OLDUS1234X", "hash", "1985-01-01", TierBasic)
	_, _ = pipeline.ApproveApplication("KYC-013")

	// Manually set re-KYC date to past
	pipeline.mu.Lock()
	pipeline.applications["KYC-013"].ReKYCDueDate = time.Now().UTC().AddDate(0, 0, -1)
	pipeline.mu.Unlock()

	due := pipeline.CheckReKYCDue()
	if len(due) != 1 {
		t.Errorf("expected 1 re-KYC due, got %d", len(due))
	}
}
