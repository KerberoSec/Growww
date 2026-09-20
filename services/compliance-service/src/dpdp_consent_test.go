package main

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestDPDPConsent_GrantAndRevoke(t *testing.T) {
	mgr := NewDPDPConsentManager()
	userID := "USR-8821"

	// Grant essential service and analytics
	entry1, err := mgr.GrantConsent(userID, ConsentEssentialService, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("unexpected error granting essential: %v", err)
	}
	if entry1.Action != ActionGrant || entry1.RecordHash == "" {
		t.Fatalf("invalid entry: %+v", entry1)
	}

	entry2, err := mgr.GrantConsent(userID, ConsentMarketingAnalytics, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("unexpected error granting marketing: %v", err)
	}
	// Cryptographic chain check: entry2.PreviousHash must equal entry1.RecordHash
	if entry2.PreviousHash != entry1.RecordHash {
		t.Fatalf("cryptographic link broken: expected %s, got %s", entry1.RecordHash, entry2.PreviousHash)
	}

	// Revoke marketing analytics
	entry3, err := mgr.RevokeConsent(userID, ConsentMarketingAnalytics, "192.168.1.1", "Mozilla/5.0", "Do not want tracking")
	if err != nil {
		t.Fatalf("unexpected error revoking marketing: %v", err)
	}
	if entry3.Action != ActionRevoke || entry3.PreviousHash != entry2.RecordHash {
		t.Fatalf("cryptographic chain mismatch on revocation: expected %s, got %s", entry2.RecordHash, entry3.PreviousHash)
	}

	// Essential service consent cannot be revoked while account is active
	_, err = mgr.RevokeConsent(userID, ConsentEssentialService, "192.168.1.1", "Mozilla/5.0", "Cancel essential")
	if !errors.Is(err, ErrEssentialConsentRevocation) {
		t.Fatalf("expected ErrEssentialConsentRevocation, got %v", err)
	}

	// Verify state
	state, err := mgr.GetUserConsentState(userID)
	if err != nil {
		t.Fatalf("unexpected error getting state: %v", err)
	}
	if !state.EssentialTradingService {
		t.Fatal("essential service must remain true")
	}
	if state.MarketingAnalytics {
		t.Fatal("marketing analytics must be false after revocation")
	}

	// Verify audit trail length
	trail := mgr.GetConsentAuditTrail(userID)
	if len(trail) != 3 {
		t.Fatalf("expected 3 audit trail entries, got %d", len(trail))
	}
}

func TestDPDPConsent_DataPortabilityExport(t *testing.T) {
	mgr := NewDPDPConsentManager()
	userID := "USR-PORTABLE"

	_, _ = mgr.GrantConsent(userID, ConsentEssentialService, "10.0.0.1", "App/v1.0")
	_, _ = mgr.GrantConsent(userID, ConsentPromotionalSMS, "10.0.0.1", "App/v1.0")

	archiveBytes, err := mgr.ExportDataPortabilityArchive(userID, "user@growww.in", "VERIFIED", "VERIFIED")
	if err != nil {
		t.Fatalf("unexpected error exporting archive: %v", err)
	}

	var profile PortabilityProfile
	if err := json.Unmarshal(archiveBytes, &profile); err != nil {
		t.Fatalf("failed to unmarshal portability archive: %v", err)
	}

	if profile.UserID != userID || profile.Email != "user@growww.in" {
		t.Fatalf("unexpected profile data: %+v", profile)
	}
	if profile.ChecksumSHA256 == "" {
		t.Fatal("expected non-empty SHA256 checksum in export archive")
	}
	if !profile.Consents.PromotionalSMS {
		t.Fatal("expected promotional SMS consent to be reflected in export")
	}
}
