package main

import (
	"strings"
	"testing"
	"time"
)

func TestAML_SmurfingDetection(t *testing.T) {
	system := NewFIUSurveillanceSystem()
	userID := "USR-SMURF-01"
	now := time.Now().UTC()

	// 5 deposits of ₹48,000 within 24 hours (Total ₹240,000, each < ₹50,000)
	txs := []TransactionRecord{
		{TxID: "TX-1", UserID: userID, AmountINR: 48000.0, Type: TxTypeDeposit, Timestamp: now.Add(-5 * time.Hour), PaymentMode: "UPI"},
		{TxID: "TX-2", UserID: userID, AmountINR: 48000.0, Type: TxTypeDeposit, Timestamp: now.Add(-4 * time.Hour), PaymentMode: "UPI"},
		{TxID: "TX-3", UserID: userID, AmountINR: 48000.0, Type: TxTypeDeposit, Timestamp: now.Add(-3 * time.Hour), PaymentMode: "UPI"},
		{TxID: "TX-4", UserID: userID, AmountINR: 48000.0, Type: TxTypeDeposit, Timestamp: now.Add(-2 * time.Hour), PaymentMode: "UPI"},
	}

	for _, tx := range txs {
		alert := system.EvaluateRealtimeTransaction(tx)
		// Not enough aggregate yet (< ₹200k)
		if alert != nil {
			t.Fatalf("did not expect alert before reaching 4 txs and ₹200k, got: %v", alert)
		}
	}

	// 5th deposit brings total to ₹240,000 (>= ₹200k threshold)
	tx5 := TransactionRecord{
		TxID:        "TX-5",
		UserID:      userID,
		AmountINR:   48000.0,
		Type:        TxTypeDeposit,
		Timestamp:   now,
		PaymentMode: "UPI",
	}

	alert := system.EvaluateRealtimeTransaction(tx5)
	if alert == nil {
		t.Fatal("expected AML smurfing alert for structured deposits below ₹50k")
	}
	if alert.Typology != TypologyStructuringSmurfing {
		t.Errorf("expected typology %s, got %s", TypologyStructuringSmurfing, alert.Typology)
	}
	if alert.RiskScore < 90.0 {
		t.Errorf("expected risk score >= 90.0, got %.2f", alert.RiskScore)
	}
	if !alert.AutomatedFreezeExecuted {
		t.Errorf("expected automated freeze to be executed for risk score >= 90")
	}
	if len(alert.TriggeringTransactionIDs) < 4 {
		t.Errorf("expected at least 4 triggering tx IDs, got %d", len(alert.TriggeringTransactionIDs))
	}
	if alert.EvidenceMerkleRoot == "" {
		t.Error("expected non-empty evidence merkle root")
	}
}

func TestAML_RapidMovementOfFunds(t *testing.T) {
	system := NewFIUSurveillanceSystem()
	userID := "USR-PASS-THROUGH"
	now := time.Now().UTC()

	// 1. Large deposit of ₹1,000,000
	depTx := TransactionRecord{
		TxID:        "TX-DEP-1",
		UserID:      userID,
		AmountINR:   1000000.0,
		Type:        TxTypeDeposit,
		Timestamp:   now,
		PaymentMode: "RTGS",
	}
	alert := system.EvaluateRealtimeTransaction(depTx)
	// CTR alert on ₹10L is normal, but let's test pass-through withdrawal
	_ = alert

	// 2. Rapid withdrawal of ₹950,000 after 10 minutes without trading
	withTx := TransactionRecord{
		TxID:        "TX-WITH-1",
		UserID:      userID,
		AmountINR:   950000.0,
		Type:        TxTypeWithdrawal,
		Timestamp:   now.Add(10 * time.Minute),
		PaymentMode: "IMPS",
	}

	alert2 := system.EvaluateRealtimeTransaction(withTx)
	if alert2 == nil {
		t.Fatal("expected rapid movement of funds alert")
	}
	if alert2.Typology != TypologyRapidMovementOfFunds {
		t.Errorf("expected typology %s, got %s", TypologyRapidMovementOfFunds, alert2.Typology)
	}
	if alert2.Severity != SeverityCritical {
		t.Errorf("expected CRITICAL severity, got %s", alert2.Severity)
	}
	if !alert2.AutomatedFreezeExecuted {
		t.Error("expected automated account freeze for rapid pass-through laundering")
	}
}

func TestAML_DormantAccountSuddenSpike(t *testing.T) {
	system := NewFIUSurveillanceSystem()
	userID := "USR-DORMANT-01"
	now := time.Now().UTC()

	// Profile inactive for 200 days
	system.RegisterUserProfile(UserActivityProfile{
		UserID:              userID,
		LastActiveAt:        now.Add(-200 * 24 * time.Hour),
		BaselineAvgDailyINR: 20000.0,
		AccountCreatedAt:    now.Add(-300 * 24 * time.Hour),
	})

	tx := TransactionRecord{
		TxID:        "TX-SPIKE",
		UserID:      userID,
		AmountINR:   800000.0, // ₹8 Lakh
		Type:        TxTypeDeposit,
		Timestamp:   now,
		PaymentMode: "NEFT",
	}

	alert := system.EvaluateRealtimeTransaction(tx)
	if alert == nil {
		t.Fatal("expected dormant account activity alert")
	}
	if alert.Typology != TypologyDormantAccountSpike {
		t.Errorf("expected %s, got %s", TypologyDormantAccountSpike, alert.Typology)
	}
}

func TestAML_SanctionedJurisdictionFlow(t *testing.T) {
	system := NewFIUSurveillanceSystem()
	userID := "USR-SANCTION-TX"
	now := time.Now().UTC()

	tx := TransactionRecord{
		TxID:                "TX-IRN",
		UserID:              userID,
		AmountINR:           250000.0,
		Type:                TxTypeTokenTransfer,
		Timestamp:           now,
		CounterpartyCountry: "IRN", // Islamic Republic of Iran (FATF Blacklist)
		PaymentMode:         "VDA",
	}

	alert := system.EvaluateRealtimeTransaction(tx)
	if alert == nil {
		t.Fatal("expected alert for sanctioned jurisdiction transaction")
	}
	if alert.Typology != TypologySanctionedJurisdiction {
		t.Errorf("expected %s, got %s", TypologySanctionedJurisdiction, alert.Typology)
	}
	if alert.RiskScore != 98.0 {
		t.Errorf("expected risk score 98.0, got %.1f", alert.RiskScore)
	}
	if !alert.AutomatedFreezeExecuted {
		t.Error("expected immediate auto freeze for transfer to FATF blacklisted jurisdiction")
	}
}

func TestAML_FIUIND_ReportXMLGeneration(t *testing.T) {
	system := NewFIUSurveillanceSystem()

	alert := &AMLAlertEvent{
		AlertID:                 "AML-ALERT-TEST-99",
		TimestampUTC:            time.Now().UTC(),
		InvestorID:              "INV-RAM-01",
		Severity:                SeverityCritical,
		Typology:                TypologyStructuringSmurfing,
		RiskScore:               94.0,
		Description:             "Smurfing pattern detected under PMLA Sec 12",
		TriggeringTransactionIDs: []string{"TX-01", "TX-02"},
		TotalINRAmount:          240000.0,
		AutomatedFreezeExecuted: true,
		EvidenceMerkleRoot:      "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
	}

	investor := DomesticInvestorProfile{
		InvestorUUID:  "INV-RAM-01",
		DeclaredName:  "Ram Prasad Sharma",
		PAN:           "ABCPS1234E",
		AadhaarMasked: "XXXX-XXXX-8821",
		RiskTier:      RiskTierHigh,
		AccountFrozen: true,
	}

	txs := []TransactionRecord{
		{TxID: "TX-01", Timestamp: time.Now().UTC(), AmountINR: 48000.0, PaymentMode: "UPI", Type: TxTypeDeposit},
		{TxID: "TX-02", Timestamp: time.Now().UTC(), AmountINR: 48000.0, PaymentMode: "UPI", Type: TxTypeDeposit},
	}

	xmlDoc, err := system.GenerateFIUINDReportXML(alert, investor, txs)
	if err != nil {
		t.Fatalf("failed to generate FIU-IND XML: %v", err)
	}

	// Verify XML structure and content
	if !strings.Contains(xmlDoc, `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("missing standard XML declaration")
	}
	if !strings.Contains(xmlDoc, `<FIUReport version="2.0">`) {
		t.Error("missing FIUReport root tag")
	}
	if !strings.Contains(xmlDoc, "<ReportType>STR</ReportType>") {
		t.Error("expected ReportType STR in XML")
	}
	if !strings.Contains(xmlDoc, "<PAN>ABCPS1234E</PAN>") {
		t.Error("expected investor PAN in XML")
	}
	if !strings.Contains(xmlDoc, "<DigestHex>abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789</DigestHex>") {
		t.Error("expected evidence digest in XML")
	}
}
