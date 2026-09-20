package main

import (
	"testing"
	"time"
)

func TestSanctionsPEP_AdvancedScreening_ConfirmedSanctions(t *testing.T) {
	engine := NewSanctionsPEPEngine()

	// Screen exact match for Dawood Ibrahim Kaskar
	req := ScreeningRequest{
		EntityID:            "INV-UN-HIT",
		FullName:            "Dawood Ibrahim Kaskar",
		DateOfBirth:         "1955-12-26",
		NationalityISO3:     "IND",
		ResidentCountryISO3: "PAK",
	}

	result := engine.AdvancedScreen(req)
	if result.Status != ScreeningStatusConfirmedSanctions {
		t.Fatalf("expected status %s, got %s", ScreeningStatusConfirmedSanctions, result.Status)
	}
	if !result.AutoFreezeRequired {
		t.Errorf("expected auto freeze to be required for confirmed sanctions hit")
	}
	if !result.EDDRequired {
		t.Errorf("expected EDD to be required")
	}
	if result.HighestMatchScore < 0.95 {
		t.Errorf("expected score >= 0.95, got %.3f", result.HighestMatchScore)
	}
	if len(result.Matches) == 0 || result.Matches[0].WatchlistSource != SanctionsUN {
		t.Errorf("expected top match to be UN Consolidated list")
	}
}

func TestSanctionsPEP_FuzzyAndPhoneticAliases(t *testing.T) {
	engine := NewSanctionsPEPEngine()

	// Test name permutation: "Kaskar Dawood" matching "Dawood Ibrahim Kaskar"
	req := ScreeningRequest{
		EntityID: "INV-FUZZY-1",
		FullName: "Kaskar Dawood",
	}
	res := engine.AdvancedScreen(req)
	if len(res.Matches) == 0 {
		t.Fatal("expected match for token permuted name")
	}
	if res.Matches[0].MatchID != "UN-002" {
		t.Errorf("expected match ID UN-002, got %s", res.Matches[0].MatchID)
	}

	// Test phonetic variation: "Mohammad Sayeed" vs "Hafiz Muhammad Saeed"
	req2 := ScreeningRequest{
		EntityID: "INV-PHONETIC-2",
		FullName: "Mohammad Sayeed",
	}
	res2 := engine.AdvancedScreen(req2)
	if len(res2.Matches) == 0 {
		t.Fatal("expected match for phonetic variation of Hafiz Muhammad Saeed")
	}
	if res2.Matches[0].MatchID != "UN-003" {
		t.Errorf("expected match ID UN-003, got %s", res2.Matches[0].MatchID)
	}
}

func TestSanctionsPEP_FATFBlacklist(t *testing.T) {
	engine := NewSanctionsPEPEngine()

	// Investor from North Korea (PRK)
	req := ScreeningRequest{
		EntityID:        "INV-FATF-PRK",
		FullName:        "Clean Name Citizen",
		NationalityISO3: "PRK",
	}
	res := engine.AdvancedScreen(req)
	if res.Status != ScreeningStatusConfirmedSanctions {
		t.Fatalf("expected confirmed sanctions for FATF blacklisted country, got %s", res.Status)
	}
	if !res.AutoFreezeRequired {
		t.Error("expected auto-freeze for FATF blacklisted jurisdiction")
	}
}

func TestSanctionsPEP_PEPTiersAndEDD(t *testing.T) {
	engine := NewSanctionsPEPEngine()

	// Tier 1 PEP: Arun Kumar Ministerial (Union Minister)
	req := ScreeningRequest{
		EntityID:        "INV-PEP-T1",
		FullName:        "Arun Kumar Ministerial",
		DateOfBirth:     "1962-04-14",
		NationalityISO3: "IND",
	}
	res := engine.AdvancedScreen(req)
	if res.Status != ScreeningStatusPotentialMatch {
		t.Fatalf("expected POTENTIAL_MATCH for PEP, got %s", res.Status)
	}
	if res.AutoFreezeRequired {
		t.Error("PEP match must NOT automatically freeze the account")
	}
	if !res.EDDRequired {
		t.Error("PEP match must mandate Enhanced Due Diligence (EDD)")
	}
	if len(res.Matches) == 0 || !res.Matches[0].IsPEP {
		t.Fatal("expected PEP match record")
	}
	if res.Matches[0].PEPTier != PEPTier1 {
		t.Errorf("expected PEPTier1, got %v", res.Matches[0].PEPTier)
	}

	// Tier 3 PEP: Close Associate / Spouse
	reqClose := ScreeningRequest{
		EntityID: "INV-PEP-T3",
		FullName: "Political Spouse Associate",
	}
	resClose := engine.AdvancedScreen(reqClose)
	if len(resClose.Matches) == 0 || resClose.Matches[0].PEPTier != PEPTier3 {
		t.Errorf("expected PEPTier3 for close associate, got: %v", resClose.Matches)
	}
}

func TestSanctionsPEP_WhitelistFalsePositiveOverride(t *testing.T) {
	engine := NewSanctionsPEPEngine()

	investorID := "INV-FALSE-POS-01"
	req := ScreeningRequest{
		EntityID: investorID,
		FullName: "Arun Kumar Ministerial",
	}

	// Before whitelist: triggers potential PEP match
	resBefore := engine.AdvancedScreen(req)
	if resBefore.Status == ScreeningStatusCleared {
		t.Fatal("expected match prior to whitelist")
	}

	// Add whitelist exception
	engine.AddWhitelistException(WhitelistException{
		InvestorUUID:    investorID,
		MatchedEntityID: "PEP-001",
		Reason:          "Different individual with same name verified via Aadhaar & passport biometric",
		ApprovedBy:      "COMPLIANCE-LEAD-01",
		ApprovedAt:      time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().Add(365 * 24 * time.Hour),
	})

	// After whitelist: should be CLEARED_WHITELISTED
	resAfter := engine.AdvancedScreen(req)
	if resAfter.Status != ScreeningStatusClearedWhitelisted {
		t.Fatalf("expected CLEARED_WHITELISTED, got %s", resAfter.Status)
	}

	// Remove whitelist exception: reverts to match
	engine.RemoveWhitelistException(investorID)
	resReverted := engine.AdvancedScreen(req)
	if resReverted.Status == ScreeningStatusClearedWhitelisted {
		t.Fatal("expected status to revert after whitelist removal")
	}
}

func TestSanctionsPEP_BatchScreening(t *testing.T) {
	engine := NewSanctionsPEPEngine()

	batch := []ScreeningRequest{
		{EntityID: "USR-1", FullName: "Rajesh Sharma", NationalityISO3: "IND"},
		{EntityID: "USR-2", FullName: "Dawood Ibrahim Kaskar", NationalityISO3: "IND"},
		{EntityID: "USR-3", FullName: "Vikramaditya Singh Politician", NationalityISO3: "IND"},
		{EntityID: "USR-4", FullName: "Foreign Clean Investor", NationalityISO3: "PRK"}, // FATF Hit
		{EntityID: "USR-5", FullName: "John Smith", NationalityISO3: "USA"},
	}

	summary, results := engine.BatchScreenInvestors(batch)
	if summary.TotalScreened != 5 {
		t.Fatalf("expected 5 screened, got %d", summary.TotalScreened)
	}
	if summary.TotalCleared != 2 {
		t.Errorf("expected 2 cleared, got %d", summary.TotalCleared)
	}
	if summary.TotalConfirmed != 2 { // Dawood + PRK FATF
		t.Errorf("expected 2 confirmed sanctions, got %d", summary.TotalConfirmed)
	}
	if summary.TotalPotential != 1 { // Vikramaditya PEP
		t.Errorf("expected 1 potential PEP match, got %d", summary.TotalPotential)
	}
	if summary.TotalAutoFrozen != 2 {
		t.Errorf("expected 2 auto-frozen, got %d", summary.TotalAutoFrozen)
	}
	if len(results) != 5 {
		t.Errorf("expected 5 individual results, got %d", len(results))
	}
}
