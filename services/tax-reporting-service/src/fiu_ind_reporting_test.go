package main

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
)

// TestFIU_CTR_Success tests valid generation of a Cash Transaction Report (>= ₹10 Lakhs)
func TestFIU_CTR_Success(t *testing.T) {
	engine := NewFIUReportingEngine("INRE00099881", "Growww Securities Private Limited")

	suspect := FIUSuspectProfile{
		CustomerInternalID: "USR-998",
		FullName:          "Aarav Sharma",
		PAN:               "ABCDE1234F",
		AadhaarHash:       "0xabcdef1234567890abcdef1234567890abcdef12",
		CKYCNumber:        "12345678901234",
		DateOfBirth:       "1988-05-14",
		Nationality:       "IND",
		RiskCategory:      "HIGH",
		AddressLine:       "100 MG Road",
		City:              "Bengaluru",
		StateCode:         "KA",
		PinCode:           "560001",
		MobileNumber:      "9876543210",
		Email:             "aarav@example.com",
	}

	txs := []FIUTransactionRecord{
		{
			TransactionID:   "TX-CTR-001",
			TransactionDate: "2026-06-15",
			TransactionTime: "14:30:00",
			TransactionMode: "CBDC",
			TransactionType: "DEPOSIT",
			AmountINR:       1500000.0, // ₹15 Lakhs (>= ₹10 Lakhs)
			AssetCode:       "INR",
			AssetQuantity:   1500000.0,
			IPAddress:       "103.21.244.2",
		},
	}

	report, err := engine.GenerateCTR("BATCH-CTR-202606-001", suspect, txs)
	if err != nil {
		t.Fatalf("expected successful CTR generation, got error: %v", err)
	}

	if report.Header.ReportType != ReportTypeCTR {
		t.Fatalf("expected report type CTR, got %s", report.Header.ReportType)
	}
	if report.Header.ReportingEntityID != "INRE00099881" {
		t.Fatalf("unexpected RE ID: %s", report.Header.ReportingEntityID)
	}
}

// TestFIU_CTR_Rejection_BelowThreshold verifies rejection if aggregate amount is below ₹10 Lakhs
func TestFIU_CTR_Rejection_BelowThreshold(t *testing.T) {
	engine := NewFIUReportingEngine("INRE00099881", "Growww Securities Private Limited")

	suspect := FIUSuspectProfile{
		CustomerInternalID: "USR-777",
		FullName:          "Vikram Verma",
		PAN:               "BNZPK9876Q",
		Nationality:       "IND",
		RiskCategory:      "MEDIUM",
	}

	// ₹8.5 Lakhs is below ₹10 Lakh threshold
	txs := []FIUTransactionRecord{
		{
			TransactionID:   "TX-BELOW-1",
			TransactionDate: "2026-06-15",
			TransactionTime: "11:00:00",
			TransactionMode: "NEFT",
			TransactionType: "DEPOSIT",
			AmountINR:       850000.0,
			AssetCode:       "INR",
			AssetQuantity:   850000.0,
			IPAddress:       "103.22.200.1",
		},
	}

	_, err := engine.GenerateCTR("BATCH-CTR-FAIL", suspect, txs)
	if err == nil {
		t.Fatalf("expected error generating CTR below ₹10 Lakhs threshold, but succeeded")
	}

	if !strings.Contains(err.Error(), "CTR_THRESHOLD_NOT_MET") {
		t.Fatalf("expected CTR_THRESHOLD_NOT_MET error, got: %v", err)
	}
}

// TestFIU_STR_Success tests creation and validation of an STR for money laundering suspicion
func TestFIU_STR_Success(t *testing.T) {
	engine := NewFIUReportingEngine("INRE00099881", "Growww Securities Private Limited")

	suspect := FIUSuspectProfile{
		CustomerInternalID: "USR-SUSPECT-01",
		FullName:          "Ramesh Patel",
		PAN:               "AAACG7777K",
		Nationality:       "IND",
		RiskCategory:      "HIGH",
		AddressLine:       "Satellite Road",
		City:              "Ahmedabad",
		StateCode:         "GJ",
		PinCode:           "380015",
	}

	txs := []FIUTransactionRecord{
		{
			TransactionID:     "TX-STR-01",
			TransactionDate:   "2026-06-16",
			TransactionTime:   "02:15:22",
			TransactionMode:   "CRYPTO_ONCHAIN",
			TransactionType:   "WITHDRAWAL",
			AmountINR:         4500000.0,
			AssetCode:         "BTC",
			AssetQuantity:     0.75,
			TxHash:            "0x9e12345678abcdef012345678abcdef012345678abcdef012345678abcdef01",
			DestinationWallet: "0xTornadoCashProxyAddressExample",
			IPAddress:         "185.220.101.5",
		},
	}

	grounds := FIUSuspicionGrounds{
		SuspicionCategory:     CategoryDarknetMixer,
		DetailedNarrative:     "Customer attempted immediate withdrawal of high-value BTC to known Tornado Cash privacy mixer contract within 10 minutes of deposit",
		ActionTaken:           "FROZEN",
		InvestigatorReference: "AML-INV-2026-889",
	}

	report, err := engine.GenerateSTR("BATCH-STR-202606-001", suspect, txs, grounds)
	if err != nil {
		t.Fatalf("expected successful STR filing, got: %v", err)
	}

	if report.SuspicionGrounds.ActionTaken != "FROZEN" {
		t.Fatalf("expected action taken FROZEN, got %s", report.SuspicionGrounds.ActionTaken)
	}
}

// TestFIU_STR_Rejection_MissingGrounds verifies rejection when suspicion grounds are missing or invalid
func TestFIU_STR_Rejection_MissingGrounds(t *testing.T) {
	engine := NewFIUReportingEngine("INRE00099881", "Growww Securities")

	suspect := FIUSuspectProfile{
		CustomerInternalID: "USR-SUSPECT-02",
		FullName:          "Sunil Mehta",
		PAN:               "BLRGP1234F",
	}

	txs := []FIUTransactionRecord{
		{
			TransactionID:   "TX-002",
			TransactionDate: "2026-06-16",
			AmountINR:       500000.0,
			AssetCode:       "USDT",
			IPAddress:       "127.0.0.1",
		},
	}

	// Case 1: Short narrative (< 20 chars)
	invalidGrounds := FIUSuspicionGrounds{
		SuspicionCategory: CategoryStructuring,
		DetailedNarrative: "Too short",
		ActionTaken:       "MONITORED",
	}

	_, err := engine.GenerateSTR("BATCH-STR-FAIL", suspect, txs, invalidGrounds)
	if err == nil {
		t.Fatalf("expected error on short narrative, but succeeded")
	}
	if !strings.Contains(err.Error(), "MIN_LENGTH") {
		t.Fatalf("expected MIN_LENGTH error, got %v", err)
	}
}

// TestFIU_StructuringDetection tests algorithmic structuring detection algorithm
func TestFIU_StructuringDetection(t *testing.T) {
	// 3 transactions: two near ₹10L limit (₹9,50,000 and ₹9,20,000)
	amounts := []float64{950000.0, 920000.0, 50000.0}
	isStructuring, count, total := DetectStructuring(amounts)

	if !isStructuring {
		t.Fatalf("expected structuring to be flagged for multiple amounts near ₹10L")
	}
	if count != 2 {
		t.Fatalf("expected count near limit to be 2, got %d", count)
	}
	if total != 1920000.0 {
		t.Fatalf("expected total 19,20,000, got %.2f", total)
	}

	// Clean scenario: transactions well below or single transaction
	cleanAmounts := []float64{100000.0, 200000.0, 300000.0}
	isStructuringClean, _, _ := DetectStructuring(cleanAmounts)
	if isStructuringClean {
		t.Fatalf("structuring should not be flagged for standard transaction sizes")
	}
}

// TestFIU_ExportToXMLAndJSON verifies schema-compliant serialization
func TestFIU_ExportToXMLAndJSON(t *testing.T) {
	engine := NewFIUReportingEngine("INRE00099881", "Growww Securities")

	suspect := FIUSuspectProfile{
		CustomerInternalID: "USR-EXPORT-01",
		FullName:          "Priya Nair",
		PAN:               "ABCDE1234F",
		Nationality:       "IND",
		RiskCategory:      "LOW",
	}

	txs := []FIUTransactionRecord{
		{
			TransactionID:   "TX-EXP-1",
			TransactionDate: "2026-06-17",
			TransactionTime: "12:00:00",
			TransactionMode: "IMPS",
			TransactionType: "DEPOSIT",
			AmountINR:       1200000.0,
			AssetCode:       "INR",
			AssetQuantity:   1200000.0,
			IPAddress:       "49.207.210.1",
		},
	}

	report, err := engine.GenerateCTR("BATCH-EXP-001", suspect, txs)
	if err != nil {
		t.Fatalf("failed to generate CTR: %v", err)
	}

	// 1. Test XML Export
	xmlBytes, err := report.ExportToXML()
	if err != nil {
		t.Fatalf("failed to export XML: %v", err)
	}
	if !strings.Contains(string(xmlBytes), "<FIUReportFiling>") || !strings.Contains(string(xmlBytes), "<ReportingEntityID>INRE00099881</ReportingEntityID>") {
		t.Fatalf("XML missing key tags: %s", string(xmlBytes))
	}

	// Verify unmarshaling XML
	var roundTripReport FIUReportFiling
	if err := xml.Unmarshal(xmlBytes, &roundTripReport); err != nil {
		t.Fatalf("failed to unmarshal XML: %v", err)
	}

	// 2. Test JSON Export
	jsonBytes, err := report.ExportToJSON()
	if err != nil {
		t.Fatalf("failed to export JSON: %v", err)
	}
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
	if jsonMap["header"] == nil {
		t.Fatalf("JSON export missing header")
	}
}

// TestFIU_ValidationFailures tests multiple specific validation error rules
func TestFIU_ValidationFailures(t *testing.T) {
	engine := NewFIUReportingEngine("INRE00099881", "Growww Securities")

	baseSuspect := FIUSuspectProfile{
		CustomerInternalID: "U1",
		FullName:          "Valid Name",
		PAN:               "ABCDE1234F",
	}

	validTx := FIUTransactionRecord{
		TransactionID:   "T1",
		TransactionDate: "2026-06-15",
		AmountINR:       1500000.0,
	}

	// 1. Empty BatchID
	r1 := &FIUReportFiling{
		Header:       FIUBatchHeader{BatchID: "", ReportType: ReportTypeCTR, ReportingEntityID: "INRE1"},
		Suspect:      baseSuspect,
		Transactions: []FIUTransactionRecord{validTx},
	}
	errs := engine.ValidateReport(r1)
	if len(errs) == 0 || errs[0].Field != "Header.BatchID" {
		t.Fatalf("expected BatchID error, got: %v", errs)
	}

	// 2. Invalid ReportType
	r2 := &FIUReportFiling{
		Header:       FIUBatchHeader{BatchID: "B1", ReportType: "INVALID", ReportingEntityID: "INRE1"},
		Suspect:      baseSuspect,
		Transactions: []FIUTransactionRecord{validTx},
	}
	errs = engine.ValidateReport(r2)
	if len(errs) == 0 || errs[0].Field != "Header.ReportType" {
		t.Fatalf("expected ReportType error, got: %v", errs)
	}

	// 3. Empty ReportingEntityID
	r3 := &FIUReportFiling{
		Header:       FIUBatchHeader{BatchID: "B1", ReportType: ReportTypeCTR, ReportingEntityID: ""},
		Suspect:      baseSuspect,
		Transactions: []FIUTransactionRecord{validTx},
	}
	errs = engine.ValidateReport(r3)
	if len(errs) == 0 || errs[0].Field != "Header.ReportingEntityID" {
		t.Fatalf("expected ReportingEntityID error, got: %v", errs)
	}

	// 4. Empty FullName
	r4 := &FIUReportFiling{
		Header:       FIUBatchHeader{BatchID: "B1", ReportType: ReportTypeCTR, ReportingEntityID: "INRE1"},
		Suspect:      FIUSuspectProfile{CustomerInternalID: "U1", FullName: "", PAN: "ABCDE1234F"},
		Transactions: []FIUTransactionRecord{validTx},
	}
	errs = engine.ValidateReport(r4)
	if len(errs) == 0 || errs[0].Field != "Suspect.FullName" {
		t.Fatalf("expected FullName error, got: %v", errs)
	}

	// 5. Invalid CKYC (not 14 digits)
	r5 := &FIUReportFiling{
		Header:       FIUBatchHeader{BatchID: "B1", ReportType: ReportTypeCTR, ReportingEntityID: "INRE1"},
		Suspect:      FIUSuspectProfile{CustomerInternalID: "U1", FullName: "A", PAN: "ABCDE1234F", CKYCNumber: "12345"},
		Transactions: []FIUTransactionRecord{validTx},
	}
	errs = engine.ValidateReport(r5)
	if len(errs) == 0 || errs[0].Field != "Suspect.CKYCNumber" {
		t.Fatalf("expected CKYC error, got: %v", errs)
	}

	// 6. Zero transactions
	r6 := &FIUReportFiling{
		Header:       FIUBatchHeader{BatchID: "B1", ReportType: ReportTypeCTR, ReportingEntityID: "INRE1"},
		Suspect:      baseSuspect,
		Transactions: []FIUTransactionRecord{},
	}
	errs = engine.ValidateReport(r6)
	if len(errs) == 0 || errs[0].Field != "Transactions" {
		t.Fatalf("expected Transactions empty error, got: %v", errs)
	}

	// 7. Non-positive amount or empty TxID
	r7 := &FIUReportFiling{
		Header:       FIUBatchHeader{BatchID: "B1", ReportType: ReportTypeCTR, ReportingEntityID: "INRE1"},
		Suspect:      baseSuspect,
		Transactions: []FIUTransactionRecord{{TransactionID: "", AmountINR: -100.0}},
	}
	errs = engine.ValidateReport(r7)
	if len(errs) < 2 {
		t.Fatalf("expected amount and TxID errors, got: %v", errs)
	}

	// 8. Export to XML fails on invalid report
	_, err := r7.ExportToXML()
	if err == nil {
		t.Fatalf("expected XML export to fail on invalid report")
	}
}
