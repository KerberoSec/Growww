package main

import (
	"strings"
	"testing"
	"time"
)

const testSalt = "GROWWW_SALT_COMPLIANCE_2026_TEST"

func TestPANValidator(t *testing.T) {
	v := NewPANValidator(testSalt)

	// Valid PAN tests
	validCases := []struct {
		pan        string
		entityType string
	}{
		{"ABCPE1234F", "INDIVIDUAL"},
		{"ABCCA9876K", "COMPANY"},
		{"AAAHM5555L", "HINDU_UNDIVIDED_FAMILY"},
		{"BBEFP1111Q", "PARTNERSHIP_FIRM"},
	}

	for _, tc := range validCases {
		eType, err := v.ValidateFormat(tc.pan)
		if err != nil {
			t.Fatalf("expected valid PAN for %s, got error: %v", tc.pan, err)
		}
		if eType != tc.entityType {
			t.Errorf("expected entity %s for %s, got %s", tc.entityType, tc.pan, eType)
		}
	}

	// Invalid PAN tests
	invalidPANs := []string{
		"ABC1234F",       // Too short
		"ABCDEF1234",     // Incorrect format
		"ABCDZ1234F",     // Invalid 4th character 'Z'
		"12345ABCDE",     // Digits first
		"ABCDE12345",     // Ending in digit
	}

	for _, bad := range invalidPANs {
		_, err := v.ValidateFormat(bad)
		if err == nil {
			t.Errorf("expected error for invalid PAN '%s', but got nil", bad)
		}
	}

	// Operative and Aadhaar linkage check
	err := v.VerifyPANStatus("ABCPE1234F", PANStatusOperative, true, true)
	if err != nil {
		t.Fatalf("expected operative linked PAN to pass, got: %v", err)
	}

	// Inoperative PAN should fail
	err = v.VerifyPANStatus("ABCPE1234F", PANStatusInoperative, true, true)
	if err == nil {
		t.Errorf("expected inoperative PAN to fail")
	}

	// Unlinked Aadhaar for individual should fail
	err = v.VerifyPANStatus("ABCPE1234F", PANStatusOperative, false, true)
	if err == nil {
		t.Errorf("expected unlinked Aadhaar to fail for individual")
	}
}

func TestAadhaarDataVaultValidator(t *testing.T) {
	v := NewAadhaarDataVaultValidator(testSalt)

	// Valid masked Aadhaar formats
	validMasked := []string{
		"XXXX-XXXX-1234",
		"********5678",
		"XXXXXXXX9999",
	}

	for _, m := range validMasked {
		last4, err := v.ValidateMaskedAadhaar(m)
		if err != nil {
			t.Fatalf("expected masked Aadhaar %s to be valid, got: %v", m, err)
		}
		if len(last4) != 4 {
			t.Errorf("expected 4 digits, got %s", last4)
		}
	}

	// Raw 12-digit Aadhaar violation detection
	rawAadhaars := []string{
		"123456789012",
		"1234-5678-9012",
		"1234 5678 9012",
	}

	for _, raw := range rawAadhaars {
		_, err := v.ValidateMaskedAadhaar(raw)
		if err == nil || !strings.Contains(err.Error(), "CRITICAL COMPLIANCE VIOLATION") {
			t.Errorf("expected critical violation error for raw Aadhaar %s, got: %v", raw, err)
		}
	}

	// Vault token generation
	tok, err := v.GenerateVaultToken()
	if err != nil || !strings.HasPrefix(tok, "ADV-TOK-") {
		t.Fatalf("invalid vault token generation: %s, %v", tok, err)
	}

	// CKYC 14-digit validation
	err = v.ValidateCKYCNumber("20012345678901")
	if err != nil {
		t.Fatalf("expected valid 14-char CKYC, got: %v", err)
	}

	err = v.ValidateCKYCNumber("12345")
	if err == nil {
		t.Errorf("expected short CKYC to fail")
	}
}

func TestPennyDropMatching(t *testing.T) {
	p := NewPennyDropValidator(0.85)

	// Valid match
	score, err := p.ValidatePennyDrop("HDFC0001234", "Rahul Sharma", "Rahul Sharma")
	if err != nil || score < 0.85 {
		t.Fatalf("expected exact match to pass, score=%.3f, err=%v", score, err)
	}

	// Name with honorific
	score, err = p.ValidatePennyDrop("SBIN0004321", "Mr. Rajesh Kumar", "Rajesh Kumar")
	if err != nil || score < 0.85 {
		t.Fatalf("expected honorific stripped match to pass, score=%.3f, err=%v", score, err)
	}

	// Token reordering
	score, err = p.ValidatePennyDrop("ICIC0000001", "Verma Amit", "Amit Verma")
	if err != nil || score < 0.85 {
		t.Fatalf("expected token reordered match to pass, score=%.3f, err=%v", score, err)
	}

	// Completely mismatched name (third party deposit attempt)
	score, err = p.ValidatePennyDrop("HDFC0001234", "Rahul Sharma", "Sunil Gupta")
	if err == nil {
		t.Errorf("expected mismatched name to fail penny drop, but got score: %.3f", score)
	}

	// Invalid IFSC
	_, err = p.ValidatePennyDrop("INVALIDIFSC", "Rahul Sharma", "Rahul Sharma")
	if err == nil {
		t.Errorf("expected invalid IFSC to be rejected")
	}
}

func TestSanctionsAndPEPScreening(t *testing.T) {
	s := NewSanctionsPEPEngine()

	// Screen clear user
	sanctioned, _, isPEP, _, _, frozen := s.ComprehensiveScreening("Aarav Patel", "IND")
	if sanctioned || isPEP || frozen {
		t.Fatalf("expected clear individual to have zero hits")
	}

	// Screen UN sanctioned entity
	sanctioned, src, _, _, _, frozen := s.ComprehensiveScreening("Dawood Ibrahim Kaskar", "IND")
	if !sanctioned || !frozen || src != SanctionsUN {
		t.Fatalf("expected UN sanctions hit and account freeze, got sanctioned=%v, src=%s, frozen=%v", sanctioned, src, frozen)
	}

	// Screen MHA UAPA banned organization
	sanctioned, src, _, _, _, frozen = s.ComprehensiveScreening("Lashkar-e-Taiba", "IND")
	if !sanctioned || !frozen || src != SanctionsMHAUAPA {
		t.Fatalf("expected MHA UAPA hit, got src=%s", src)
	}

	// Screen Domestic PEP
	sanctioned, _, isPEP, pepCat, eddReq, frozen := s.ComprehensiveScreening("Arun Kumar Ministerial", "IND")
	if sanctioned || frozen {
		t.Errorf("PEP should not be frozen automatically")
	}
	if !isPEP || pepCat != PEPDomestic || !eddReq {
		t.Fatalf("expected Domestic PEP with EDD required, got isPEP=%v, cat=%s, edd=%v", isPEP, pepCat, eddReq)
	}

	// Screen FATF Blacklist country (DPRK)
	if !s.IsFATFBlacklisted("PRK") {
		t.Errorf("expected PRK to be on FATF blacklist")
	}
	if s.IsFATFBlacklisted("USA") {
		t.Errorf("USA should not be on FATF blacklist")
	}
}

func TestTierLimitsAndEnforcement(t *testing.T) {
	m := NewTierLimitsManager()

	// Tier 1 limits: Daily ₹25,000, Annual ₹1,00,000
	err := m.CheckLimitEnforcement(KYCTier1BasicOTP, 20000.0, 0.0, 0.0)
	if err != nil {
		t.Fatalf("expected ₹20k withdrawal to pass on Tier 1, got: %v", err)
	}

	// Exceed daily quota
	err = m.CheckLimitEnforcement(KYCTier1BasicOTP, 10000.0, 20000.0, 20000.0)
	if err == nil || !strings.Contains(err.Error(), "daily withdrawal limit exceeded") {
		t.Errorf("expected daily limit violation on Tier 1, got: %v", err)
	}

	// Exceed annual quota
	err = m.CheckLimitEnforcement(KYCTier1BasicOTP, 5000.0, 0.0, 98000.0)
	if err == nil || !strings.Contains(err.Error(), "annual withdrawal limit exceeded") {
		t.Errorf("expected annual limit violation on Tier 1, got: %v", err)
	}

	// Tier 0 (unverified) should reject all transactions
	err = m.CheckLimitEnforcement(KYCTier0Unverified, 100.0, 0.0, 0.0)
	if err == nil {
		t.Errorf("expected Tier 0 unverified user to be blocked from withdrawing")
	}

	// Re-KYC duration checks
	now := time.Now()
	reKYCLow := m.CalculateReKYCDue(RiskTierLow, now)
	if reKYCLow.Sub(now) < 9*365*24*time.Hour {
		t.Errorf("expected ~10 years for low risk re-KYC")
	}

	reKYCHigh := m.CalculateReKYCDue(RiskTierHigh, now)
	if reKYCHigh.Sub(now) > 3*365*24*time.Hour {
		t.Errorf("expected ~2 years for high risk re-KYC")
	}
}

func TestFIUSurveillanceAndTravelRule(t *testing.T) {
	fiu := NewFIUSurveillanceSystem()

	// Test structuring evaluation: 2 transactions in [8.5L, 10L)
	recent := []float64{900000.0, 950000.0}
	alert := fiu.EvaluateStructuring("USER-101", recent)
	if alert == nil || alert.ReportType != "STR" || alert.Severity != "CRITICAL" {
		t.Fatalf("expected STR alert for cash structuring evasion, got: %+v", alert)
	}

	// Test normal transactions should not trigger structuring
	normalTxs := []float64{50000.0, 100000.0, 20000.0}
	alertNormal := fiu.EvaluateStructuring("USER-102", normalTxs)
	if alertNormal != nil {
		t.Errorf("unexpected STR alert for normal transactions: %+v", alertNormal)
	}

	// Test CTR statutory reporting: transaction >= ₹10 Lakh
	ctrAlert := fiu.EvaluateCashThreshold("USER-103", 1200000.0)
	if ctrAlert == nil || ctrAlert.ReportType != "CTR" {
		t.Fatalf("expected CTR alert for ₹12L transaction, got: %+v", ctrAlert)
	}

	// Test Travel Rule for VDA transfer > ₹50,000
	validPayload := TravelRulePayload{
		TransferID:            "TX-999",
		CryptoCurrency:        "USDT",
		AmountFiatINR:         75000.0, // > 50,000 INR
		OriginatorName:        "Vikram Malhotra",
		OriginatorAccountID:   "ACC-DOM-1234",
		OriginatorNationalID:  "ABCDE1234F",
		OriginatorVASP:        "VASP-GROWWW-IN",
		BeneficiaryName:       "John Doe",
		BeneficiaryWalletAddr: "0x71C...b5",
		BeneficiaryVASP:       "VASP-BINANCE-GLOBAL",
		Compliant:             true,
	}
	err := fiu.ValidateTravelRule(validPayload)
	if err != nil {
		t.Fatalf("expected valid Travel Rule payload to pass, got: %v", err)
	}

	// Missing originator info should fail
	invalidPayload := validPayload
	invalidPayload.OriginatorName = ""
	err = fiu.ValidateTravelRule(invalidPayload)
	if err == nil {
		t.Errorf("expected missing originator name to trigger Travel Rule rejection")
	}
}

func TestSandboxGuardrails(t *testing.T) {
	g := NewSandboxGuardrails()

	// Domestic enrollment within SEBI sandbox limits
	err := g.ValidateDomesticEnrollment(35000.0) // <= ₹50,000
	if err != nil {
		t.Fatalf("expected domestic enrollment to pass, got: %v", err)
	}

	// Exceed domestic portfolio limit
	err = g.ValidateDomesticEnrollment(75000.0) // > ₹50,000 cap
	if err == nil || !strings.Contains(err.Error(), "exceeds SEBI sandbox limit") {
		t.Errorf("expected portfolio limit rejection, got: %v", err)
	}

	// Foreign enrollment within IFSCA sandbox limits
	err = g.ValidateForeignEnrollment(5000.0) // <= $10,000
	if err != nil {
		t.Fatalf("expected foreign enrollment to pass, got: %v", err)
	}

	// Exceed foreign portfolio limit
	err = g.ValidateForeignEnrollment(15000.0) // > $10,000 cap
	if err == nil || !strings.Contains(err.Error(), "exceeds IFSCA sandbox limit") {
		t.Errorf("expected foreign portfolio limit rejection, got: %v", err)
	}

	// Banking rail checks
	err = g.ValidateFiatPaymentRail("UPI")
	if err != nil {
		t.Errorf("UPI must be permitted")
	}

	err = g.ValidateFiatPaymentRail("TETHER_USDT_UNREGULATED")
	if err == nil {
		t.Errorf("unregulated stablecoin payment rails must be rejected")
	}

	// Zero fee model check (0.00% fee at launch)
	err = g.ValidateZeroFeeModel(0, 0, 0.0, 0.0)
	if err != nil {
		t.Fatalf("expected 0.00%% fee to pass: %v", err)
	}

	err = g.ValidateZeroFeeModel(10, 0, 0.0, 0.0)
	if err == nil {
		t.Errorf("non-zero fee at launch must violate invariant")
	}
}

func TestIdentityCommitmentGenerator(t *testing.T) {
	g := NewIdentityCommitmentGenerator(testSalt)

	panHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	userUUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	comm := g.GenerateDomesticCommitment(panHash, userUUID)
	if !g.VerifyCommitmentIntegrity(comm) {
		t.Fatalf("expected valid 32-byte hex commitment, got: %s", comm)
	}

	// Ensure zero PII is in the commitment
	if strings.Contains(comm, panHash) || strings.Contains(comm, userUUID) {
		t.Errorf("commitment must not leak raw preimage plaintext")
	}
}

func TestEndToEndDomesticOnboardingPipeline(t *testing.T) {
	svc := NewComplianceEngineService(testSalt)

	// Successful onboarding profile
	profile := &DomesticInvestorProfile{
		InvestorUUID:      "usr-domestic-001",
		PAN:               "ABCPS1234P",
		PANStatus:         PANStatusOperative,
		AadhaarPANLinked:  true,
		AadhaarMasked:     "XXXX-XXXX-9876",
		CKYCNumber:        "20098765432100",
		IFSC:              "HDFC0001234",
		BeneficiaryName:   "Priya Sharma",
		DeclaredName:      "Priya Sharma",
		FaceLivenessScore: 0.96,
		AntiSpoofPassed:   true,
		AnnualIncomeINR:   1200000.0,
		Occupation:        "SOFTWARE_ENGINEER",
	}

	approved, err := svc.ProcessDomesticKYCOnboarding(profile)
	if err != nil {
		t.Fatalf("expected successful domestic onboarding, got error: %v", err)
	}

	if approved.AssignedTier != KYCTier2FullCKYC {
		t.Errorf("expected Tier 2 Full CKYC assignment, got: %s", approved.AssignedTier)
	}
	if approved.RiskTier != RiskTierLow {
		t.Errorf("expected Low risk tier, got: %s", approved.RiskTier)
	}
	if approved.CommitmentHash == "" || !strings.HasPrefix(approved.CommitmentHash, "0x") {
		t.Errorf("expected valid on-chain commitment hash, got: %s", approved.CommitmentHash)
	}

	// Test withdrawal limit check
	err = svc.CheckWithdrawalLimit("usr-domestic-001", 100000.0, 0.0, 0.0)
	if err != nil {
		t.Fatalf("expected ₹1L withdrawal to pass for Tier 2, got: %v", err)
	}

	// Exceed Tier 2 daily quota (₹5,00,000)
	err = svc.CheckWithdrawalLimit("usr-domestic-001", 600000.0, 0.0, 0.0)
	if err == nil {
		t.Errorf("expected withdrawal > ₹5L to be blocked for Tier 2")
	}

	// Upgrade to Tier 3 VCIP
	err = svc.UpgradeToTier3VCIP("usr-domestic-001", true, 20000000.0) // ₹2 Cr CA net worth
	if err != nil {
		t.Fatalf("expected Tier 3 upgrade to succeed, got: %v", err)
	}

	// Now ₹10,00,000 withdrawal passes under Tier 3 daily quota (₹5 Cr)
	err = svc.CheckWithdrawalLimit("usr-domestic-001", 1000000.0, 0.0, 0.0)
	if err != nil {
		t.Fatalf("expected ₹10L withdrawal to pass under Tier 3, got: %v", err)
	}
}
