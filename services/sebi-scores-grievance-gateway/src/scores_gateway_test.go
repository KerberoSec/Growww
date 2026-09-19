package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testWebhookSecret = "TEST_SCORES_SECRET_KEY_2026"

func computeTestSignature(payload ScoresWebhookPayload, secret string) string {
	rawBytes, _ := json.Marshal(payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBytes)
	return hex.EncodeToString(mac.Sum(nil))
}

func TestSCORES_Ingestion_Validation(t *testing.T) {
	cfg := ScoresGatewayConfig{
		WebhookSecret: testWebhookSecret,
	}
	userStore := NewMemoryUserMappingStore()
	userStore.RegisterUserPAN("ABCPE1234F", "usr-uuid-9812-alpha")

	gw := NewSCORESGatewayConnector(cfg, userStore)

	// 1. Valid Complaint Ingestion
	intakeDate := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	payload := ScoresWebhookPayload{
		ScoresRegistrationNumber: "SEBIP/2026/0014529",
		ComplainantName:          "Vikram Malhotra",
		ComplainantPAN:           "ABCPE1234F",
		ComplainantEmail:         "vikram.m@example.com",
		ComplainantMobile:        "+919876543210",
		ComplaintCategory:        "NON_CREDIT_OF_FUNDS",
		SubCategory:              "UPI_DEPOSIT_DELAY",
		ComplaintDescription:     "Amount debited from bank via UPI but trading wallet not credited.",
		ReceiptDate:              intakeDate,
		DisputedAmount:           25000.0,
		DematBOID:                "1208160012345678",
	}

	sig := computeTestSignature(payload, testWebhookSecret)
	docket, err := gw.IngestSCORESComplaint(payload, sig)
	if err != nil {
		t.Fatalf("failed to ingest valid SCORES complaint: %v", err)
	}

	if docket.ExternalRegistrationNumber != "SEBIP/2026/0014529" {
		t.Errorf("expected external reg number SEBIP/2026/0014529, got %s", docket.ExternalRegistrationNumber)
	}
	if docket.UserID != "usr-uuid-9812-alpha" {
		t.Errorf("expected user mapping usr-uuid-9812-alpha, got %s", docket.UserID)
	}
	if docket.StatutorySLADays != 21 {
		t.Errorf("expected 21 statutory days, got %d", docket.StatutorySLADays)
	}
	expectedDeadline := intakeDate.AddDate(0, 0, 21)
	if !docket.StatutoryDeadline.Equal(expectedDeadline) {
		t.Errorf("expected deadline %v, got %v", expectedDeadline, docket.StatutoryDeadline)
	}
	if docket.ComplainantNameMasked == "Vikram Malhotra" {
		t.Errorf("expected masked complainant name, got plain text: %s", docket.ComplainantNameMasked)
	}
	if docket.ComplaintHash == "" {
		t.Errorf("expected non-empty deterministic complaint hash")
	}

	// 2. Duplicate Docket Rejection
	_, err = gw.IngestSCORESComplaint(payload, sig)
	if err != ErrDuplicateDocket {
		t.Errorf("expected ErrDuplicateDocket, got: %v", err)
	}

	// 3. Invalid HMAC Signature Rejection
	payload2 := ScoresWebhookPayload{
		ScoresRegistrationNumber: "SEBIP/2026/0099999",
		ComplainantName:          "Rohan Das",
		ComplainantPAN:           "ABCPM5555K",
		ComplaintCategory:        "TRADE_EXECUTION",
		ComplaintDescription:     "Order slipped past limit price",
		ReceiptDate:              intakeDate,
	}
	_, err = gw.IngestSCORESComplaint(payload2, "bad_signature_hex_deadbeef")
	if err != ErrInvalidHMACSignature {
		t.Errorf("expected ErrInvalidHMACSignature, got: %v", err)
	}

	// 4. Invalid PAN Format Rejection
	badPANPayload := ScoresWebhookPayload{
		ScoresRegistrationNumber: "SEBIP/2026/0088888",
		ComplainantName:          "Bad Pan User",
		ComplainantPAN:           "INVALID123", // invalid length and format
		ComplaintCategory:        "OTHER",
		ComplaintDescription:     "Invalid PAN test",
	}
	_, err = gw.IngestSCORESComplaint(badPANPayload, "")
	if err != ErrInvalidPANFormat {
		t.Errorf("expected ErrInvalidPANFormat, got: %v", err)
	}
}

func TestSLA_Escalation_Milestones(t *testing.T) {
	cfg := ScoresGatewayConfig{}
	gw := NewSCORESGatewayConnector(cfg, nil)
	timer := NewEscalationTimer(gw)

	intakeDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	payload := ScoresWebhookPayload{
		ScoresRegistrationNumber: "SEBIP/2026/0077777",
		ComplainantName:          "Pooja Mehta",
		ComplainantPAN:           "AAAPM1111L",
		ComplaintCategory:        "DIVIDEND_MISMATCH",
		ComplaintDescription:     "Corporate action entitlement missing",
		ReceiptDate:              intakeDate,
	}

	docket, err := gw.IngestSCORESComplaint(payload, "")
	if err != nil {
		t.Fatalf("failed to ingest: %v", err)
	}

	// Milestone 0: Day 2 (Normal)
	evalDay2 := intakeDate.AddDate(0, 0, 2)
	stage, elapsed, remaining := timer.CalculateEscalationStage(docket, evalDay2)
	if stage != Stage0Normal || elapsed != 2 || remaining != 19 {
		t.Errorf("Day 2: expected Stage 0, elapsed 2, remaining 19; got stage=%d, elapsed=%d, remaining=%d", stage, elapsed, remaining)
	}

	// Milestone 1: Day 7 (Warning)
	evalDay7 := intakeDate.AddDate(0, 0, 7)
	stage, elapsed, remaining = timer.CalculateEscalationStage(docket, evalDay7)
	if stage != Stage1Day7Warning || elapsed != 7 || remaining != 14 {
		t.Errorf("Day 7: expected Stage 1, elapsed 7, remaining 14; got stage=%d, elapsed=%d, remaining=%d", stage, elapsed, remaining)
	}

	// Milestone 2: Day 14 (Principal Compliance Officer Escalation)
	evalDay14 := intakeDate.AddDate(0, 0, 14)
	stage, elapsed, remaining = timer.CalculateEscalationStage(docket, evalDay14)
	if stage != Stage2Day14Escalation || elapsed != 14 || remaining != 7 {
		t.Errorf("Day 14: expected Stage 2, elapsed 14, remaining 7; got stage=%d, elapsed=%d, remaining=%d", stage, elapsed, remaining)
	}

	// Milestone 3: Day 18 (Critical P1 RUNBOOK-34)
	evalDay18 := intakeDate.AddDate(0, 0, 18)
	stage, elapsed, remaining = timer.CalculateEscalationStage(docket, evalDay18)
	if stage != Stage3Day18CriticalP1 || elapsed != 18 || remaining != 3 {
		t.Errorf("Day 18: expected Stage 3, elapsed 18, remaining 3; got stage=%d, elapsed=%d, remaining=%d", stage, elapsed, remaining)
	}

	// Milestone 4: Day 20 (Mandatory Submission Freeze)
	evalDay20 := intakeDate.AddDate(0, 0, 20)
	stage, elapsed, remaining = timer.CalculateEscalationStage(docket, evalDay20)
	if stage != Stage4Day20Freeze || elapsed != 20 || remaining != 1 {
		t.Errorf("Day 20: expected Stage 4, elapsed 20, remaining 1; got stage=%d, elapsed=%d, remaining=%d", stage, elapsed, remaining)
	}

	// Milestone 5: Day 21 (Statutory SLA Breach Boundary)
	evalDay21 := intakeDate.AddDate(0, 0, 21)
	stage, elapsed, remaining = timer.CalculateEscalationStage(docket, evalDay21)
	if stage != Stage5Day21Breached || elapsed != 21 || remaining != 0 {
		t.Errorf("Day 21: expected Stage 5, elapsed 21, remaining 0; got stage=%d, elapsed=%d, remaining=%d", stage, elapsed, remaining)
	}

	// Test Alert Emission upon escalation
	alert := timer.EvaluateDocketSLA(docket, evalDay18)
	if alert == nil {
		t.Fatalf("expected alert for Day 18 escalation, got nil")
	}
	if alert.AlertSeverity != "CRITICAL_P1" {
		t.Errorf("expected CRITICAL_P1 alert severity, got %s", alert.AlertSeverity)
	}

	// Test SLA Dashboard Aggregation
	dashboard := timer.GetSLADashboard(evalDay18)
	if dashboard.TotalActiveDockets != 1 {
		t.Errorf("expected 1 active docket, got %d", dashboard.TotalActiveDockets)
	}
	if dashboard.CriticalBreachRiskCount != 1 {
		t.Errorf("expected 1 critical breach risk count, got %d", dashboard.CriticalBreachRiskCount)
	}
}

func TestMakerChecker_DualControl_And_Resolution(t *testing.T) {
	cfg := ScoresGatewayConfig{}
	gw := NewSCORESGatewayConnector(cfg, nil)
	atrGen := NewATRGenerator(gw)

	payload := ScoresWebhookPayload{
		ScoresRegistrationNumber: "SEBIP/2026/0055555",
		ComplainantName:          "Ananya Roy",
		ComplainantPAN:           "BCCPR8888K",
		ComplaintCategory:        "NON_CREDIT_OF_FUNDS",
		ComplaintDescription:     "Deposit failed but funds debited",
		ReceiptDate:              time.Now().UTC(),
		DisputedAmount:           15000.0,
	}
	docket, err := gw.IngestSCORESComplaint(payload, "")
	if err != nil {
		t.Fatalf("failed to ingest: %v", err)
	}

	// 1. Maker drafts the ATR
	makerID := "officer-maker-001"
	draftReq := DraftATRRequest{
		ComplaintID:           docket.ID,
		MakerUserID:           makerID,
		Disposition:           DispositionResolvedWithRefund,
		ActionTakenSummary:    "Reconciled bank statement with payment aggregator. Refund initiated.",
		DetailedRedressalText: "The UPI transaction was delayed at the intermediary switch. INR 15,000 has been credited back to investor bank account.",
		RefundAmount:          15000.0,
		SettlementBankUTR:     "UTR20260920AXIS99901",
		SupportingDocumentRefs: []string{"DOC-UTR-PROOF-01", "DOC-BANK-ACK-02"},
	}

	atr, err := atrGen.DraftATR(draftReq)
	if err != nil {
		t.Fatalf("failed to draft ATR: %v", err)
	}
	if atr.MakerUserID != makerID {
		t.Errorf("expected maker ID %s, got %s", makerID, atr.MakerUserID)
	}
	if docket.CurrentState != StateATRDraftedMaker {
		t.Errorf("expected state StateATRDraftedMaker, got %s", docket.CurrentState)
	}

	// 2. Dual-Control Violation: Maker attempts to approve own draft
	_, err = atrGen.ApproveATR(atr.ID, makerID, "Self approving")
	if err != ErrDualControlViolation {
		t.Fatalf("expected ErrDualControlViolation for self approval, got: %v", err)
	}

	// 3. Valid Checker Approves ATR
	checkerID := "officer-checker-002"
	atr, err = atrGen.ApproveATR(atr.ID, checkerID, "Verified bank statement and UTR credit proof.")
	if err != nil {
		t.Fatalf("checker approval failed: %v", err)
	}
	if atr.CheckerUserID != checkerID {
		t.Errorf("expected checker ID %s, got %s", checkerID, atr.CheckerUserID)
	}
	if docket.CurrentState != StateATRApprovedChecker {
		t.Errorf("expected state StateATRApprovedChecker, got %s", docket.CurrentState)
	}

	// 4. Attempt submission without Class-3 DSC signing -> should fail
	_, err = gw.SubmitATRToRegulator(atr)
	if err != ErrDSCNotSigned {
		t.Errorf("expected ErrDSCNotSigned, got: %v", err)
	}

	// 5. Apply Class-3 Digital Signature Certificate
	signerDN := "CN=Chief Compliance Officer, O=NBSE / Growww, C=IN"
	atr, err = atrGen.SignATRWithDSC(atr.ID, signerDN)
	if err != nil {
		t.Fatalf("failed to apply DSC: %v", err)
	}
	if !atr.IsDSCSigned || atr.ATRPackageHash == "" {
		t.Errorf("expected DSC signed with package hash, got: %v", atr)
	}

	// Verify report formatting
	formatted, err := atrGen.FormatATRReportText(atr)
	if err != nil || len(formatted) == 0 {
		t.Fatalf("failed to format ATR report text: %v", err)
	}

	// 6. Formally submit ATR to Regulator and anchor receipt on Hyperledger Besu
	subResult, err := gw.SubmitATRToRegulator(atr)
	if err != nil {
		t.Fatalf("failed to submit ATR to regulator: %v", err)
	}

	if subResult.RegulatoryAckNumber == "" {
		t.Errorf("expected non-empty regulatory ack number")
	}
	if subResult.BesuTransactionHash == "" {
		t.Errorf("expected on-chain Besu transaction hash")
	}
	if !subResult.ZeroPIIConfirmed {
		t.Errorf("expected zero PII confirmation")
	}

	// Docket should now be formally closed
	if docket.CurrentState != StateResolvedClosed {
		t.Errorf("expected state StateResolvedClosed, got %s", docket.CurrentState)
	}

	// 7. Modifying or re-submitting an already resolved docket must be rejected
	_, err = gw.SubmitATRToRegulator(atr)
	if err != ErrDocketAlreadyResolved {
		t.Errorf("expected ErrDocketAlreadyResolved, got: %v", err)
	}
}

func TestGrievanceService_HTTPEndpoints(t *testing.T) {
	cfg := ScoresGatewayConfig{
		WebhookSecret: testWebhookSecret,
	}
	svc := NewGrievanceService(cfg)
	mux := http.NewServeMux()
	svc.RegisterHTTPHandlers(mux)

	// 1. Test Webhook Ingestion Endpoint
	payload := ScoresWebhookPayload{
		ScoresRegistrationNumber: "SEBIP/2026/0033333",
		ComplainantName:          "Suresh Nair",
		ComplainantPAN:           "ZZZPN9999Q",
		ComplaintCategory:        "UNAUTHORIZED_TRADE",
		ComplaintDescription:     "Alleged trade executed without confirmation",
		ReceiptDate:              time.Now().UTC(),
		DisputedAmount:           45000.0,
	}
	sig := computeTestSignature(payload, testWebhookSecret)
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/grievance/scores/webhook", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Scores-Signature-256", sig)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}

	var createdDocket ComplaintDocket
	if err := json.NewDecoder(rec.Body).Decode(&createdDocket); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if createdDocket.ExternalRegistrationNumber != "SEBIP/2026/0033333" {
		t.Errorf("unexpected docket number: %s", createdDocket.ExternalRegistrationNumber)
	}

	// 2. Test List Dockets Endpoint
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/grievance/dockets", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", listRec.Code)
	}
	var dockets []*ComplaintDocket
	_ = json.NewDecoder(listRec.Body).Decode(&dockets)
	if len(dockets) != 1 {
		t.Errorf("expected 1 docket, got %d", len(dockets))
	}

	// 3. Test SLA Dashboard Endpoint
	dashReq := httptest.NewRequest(http.MethodGet, "/api/v1/grievance/sla/dashboard", nil)
	dashRec := httptest.NewRecorder()
	mux.ServeHTTP(dashRec, dashReq)
	if dashRec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for dashboard, got %d", dashRec.Code)
	}
	var dash SLADashboard
	_ = json.NewDecoder(dashRec.Body).Decode(&dash)
	if dash.TotalActiveDockets != 1 {
		t.Errorf("expected 1 active docket on dashboard, got %d", dash.TotalActiveDockets)
	}
}
