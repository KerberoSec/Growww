package main

import (
	"strings"
	"testing"
)

func TestCKYC_RecordValidation(t *testing.T) {
	engine := NewCKYCBatchEngine("IN_GROWWW_001")

	validRec := CKYCInvestorRecord{
		InvestorUUID:     "INV-CKYC-01",
		FullName:         "Aaditya Verma",
		Gender:           "M",
		DOB:              "15-08-1992",
		PAN:              "ABCDE1234F",
		MaskedAadhaar:    "XXXX-XXXX-9912",
		FatherSpouseName: "Mahesh Verma",
		MotherName:       "Sunita Verma",
		Occupation:       "SALARIED_SOFTWARE",
		Email:            "aaditya.verma@example.com",
		Mobile:           "9876543210",
		AddressLine1:     "Flat 402, High Tech Residency",
		City:             "Bengaluru",
		StateCode:        "KA",
		PINCode:          "560100",
		CountryCode:      "IND",
		PoIType:          DocTypePAN,
		PoINumber:        "ABCDE1234F",
		PoAType:          DocTypeAadhaarOffline,
		PoANumber:        "XXXX-XXXX-9912",
		PhotoFileName:    "photo_inv_01.jpg",
		SignatureFileName: "sign_inv_01.jpg",
		LivenessScore:    0.95,
	}

	err := engine.ValidateRecord(validRec)
	if err != nil {
		t.Fatalf("expected valid record to pass, got: %v", err)
	}

	// Invalid PAN
	badPAN := validRec
	badPAN.PAN = "INVALID_PAN"
	if err := engine.ValidateRecord(badPAN); err == nil {
		t.Error("expected error for invalid PAN format")
	}

	// Invalid PIN
	badPIN := validRec
	badPIN.PINCode = "56010" // 5 digits
	if err := engine.ValidateRecord(badPIN); err == nil {
		t.Error("expected error for invalid PIN code")
	}

	// Incomplete address
	badAddr := validRec
	badAddr.City = ""
	if err := engine.ValidateRecord(badAddr); err == nil {
		t.Error("expected error for missing City")
	}

	// Low liveness
	badLiveness := validRec
	badLiveness.LivenessScore = 0.70
	if err := engine.ValidateRecord(badLiveness); err == nil {
		t.Error("expected error for low biometric liveness score")
	}
}

func TestCKYC_BatchLifecycleAndDigest(t *testing.T) {
	engine := NewCKYCBatchEngine("IN_GROWWW_001")

	recs := []CKYCInvestorRecord{
		{
			InvestorUUID:  "INV-001",
			FullName:      "Kavita Reddy",
			Gender:        "F",
			DOB:           "22-01-1995",
			PAN:           "BCCPR5544K",
			MaskedAadhaar: "XXXX-XXXX-1122",
			AddressLine1:  "Plot 12, Jubilee Hills",
			City:          "Hyderabad",
			StateCode:     "TS",
			PINCode:       "500033",
			CountryCode:   "IND",
			PoIType:       DocTypePAN,
			PoINumber:     "BCCPR5544K",
			PoAType:       DocTypePassport,
			PoANumber:     "Z1234567",
			LivenessScore: 0.94,
		},
		{
			InvestorUUID:  "INV-002",
			FullName:      "Deepak Joshi",
			Gender:        "M",
			DOB:           "10-11-1988",
			PAN:           "AAAPJ8899M",
			MaskedAadhaar: "XXXX-XXXX-3344",
			AddressLine1:  "Sector 18",
			City:          "Gurugram",
			StateCode:     "HR",
			PINCode:       "122001",
			CountryCode:   "IND",
			PoIType:       DocTypePAN,
			PoINumber:     "AAAPJ8899M",
			PoAType:       DocTypeVoterID,
			PoANumber:     "VTR998811",
			LivenessScore: 0.92,
		},
	}

	batchState, err := engine.CreateBatch(recs)
	if err != nil {
		t.Fatalf("failed to create batch: %v", err)
	}

	if batchState.Status != CKYCStatusValidated {
		t.Errorf("expected status %s, got %s", CKYCStatusValidated, batchState.Status)
	}
	if len(batchState.DigestHex) != 64 {
		t.Errorf("expected 64-char SHA-256 batch digest, got %s", batchState.DigestHex)
	}
	if !strings.Contains(batchState.XMLPayload, "<CKYCBatch xmlns=\"http://cersai.org.in/ckyc/v1.2\">") {
		t.Error("missing CERSAI namespace in XML")
	}
	if !strings.Contains(batchState.XMLPayload, "<BatchDigest>") {
		t.Error("missing BatchDigest in XML payload")
	}

	// Mark uploaded
	err = engine.MarkBatchUploaded(batchState.BatchID)
	if err != nil {
		t.Fatalf("failed to mark batch uploaded: %v", err)
	}

	st, _ := engine.GetBatchState(batchState.BatchID)
	if st.Status != CKYCStatusUploaded {
		t.Errorf("expected status UPLOADED, got %s", st.Status)
	}

	// Simulate CERSAI response XML with 14-digit KIN assignment
	mockResponseXML := `<?xml version="1.0" encoding="UTF-8"?>
<CKYCBatchResponse>
  <BatchID>` + batchState.BatchID + `</BatchID>
  <Status>SUCCESS</Status>
  <TotalRecords>2</TotalRecords>
  <SuccessCount>2</SuccessCount>
  <FailureCount>0</FailureCount>
  <RecordResults>
    <RecordResult>
      <InvestorUUID>INV-001</InvestorUUID>
      <Status>SUCCESS</Status>
      <CKYCNumber>50012345678901</CKYCNumber>
    </RecordResult>
    <RecordResult>
      <InvestorUUID>INV-002</InvestorUUID>
      <Status>SUCCESS</Status>
      <CKYCNumber>50012345678902</CKYCNumber>
    </RecordResult>
  </RecordResults>
</CKYCBatchResponse>`

	updatedState, err := engine.ProcessBatchResponseXML(mockResponseXML)
	if err != nil {
		t.Fatalf("failed to process response XML: %v", err)
	}

	if updatedState.Status != CKYCStatusSuccess {
		t.Errorf("expected status %s, got %s", CKYCStatusSuccess, updatedState.Status)
	}
	if updatedState.AssignedKINs["INV-001"] != "50012345678901" {
		t.Errorf("expected 14-digit KIN 50012345678901, got %s", updatedState.AssignedKINs["INV-001"])
	}
	if updatedState.AssignedKINs["INV-002"] != "50012345678902" {
		t.Errorf("expected 14-digit KIN 50012345678902, got %s", updatedState.AssignedKINs["INV-002"])
	}
}

func TestCKYC_SearchAndProfileMapping(t *testing.T) {
	engine := NewCKYCBatchEngine("IN_GROWWW_001")

	// 1. Generate Search Request XML
	searchXML, err := engine.GenerateSearchRequestXML("REQ-SEARCH-01", "50012345678901", "ABCDE1234F", "15-08-1992")
	if err != nil {
		t.Fatalf("failed to generate search request: %v", err)
	}
	if !strings.Contains(searchXML, "<CKYCSearchRequest>") || !strings.Contains(searchXML, "<CKYCNumber>50012345678901</CKYCNumber>") {
		t.Error("search request XML missing required fields")
	}

	// 2. Parse Search Response XML
	mockSearchResponseXML := `<?xml version="1.0" encoding="UTF-8"?>
<CKYCSearchResponse>
  <RequestID>REQ-SEARCH-01</RequestID>
  <Status>FOUND</Status>
  <CKYCNumber>50012345678901</CKYCNumber>
  <FullName>Kavita Reddy</FullName>
  <PAN>BCCPR5544K</PAN>
  <MaskedAadhaar>XXXX-XXXX-1122</MaskedAadhaar>
  <Gender>F</Gender>
  <DateOfBirth>22-01-1995</DateOfBirth>
  <KYCDate>2024-03-12</KYCDate>
  <VerifyingFIEntityCode>IN000999</VerifyingFIEntityCode>
  <KYCStatusDesc>VERIFIED_NORMAL_KYC</KYCStatusDesc>
  <RiskCategory>LOW</RiskCategory>
  <AddressLine1>Plot 12, Jubilee Hills</AddressLine1>
  <City>Hyderabad</City>
  <StateCode>TS</StateCode>
  <PINCode>500033</PINCode>
</CKYCSearchResponse>`

	searchResult, err := engine.ParseSearchResponseXML(mockSearchResponseXML)
	if err != nil {
		t.Fatalf("failed to parse search response: %v", err)
	}
	if searchResult.CKYCNumber != "50012345678901" {
		t.Errorf("expected KIN 50012345678901, got %s", searchResult.CKYCNumber)
	}

	// 3. Map to DomesticInvestorProfile
	profile := &DomesticInvestorProfile{
		InvestorUUID: "INV-PROFILE-01",
		AssignedTier: KYCTier1BasicOTP, // Prior to CKYC
	}

	err = engine.MapFetchedCKYCToProfile(searchResult, profile)
	if err != nil {
		t.Fatalf("failed to map CKYC to profile: %v", err)
	}

	if profile.CKYCNumber != "50012345678901" {
		t.Errorf("expected CKYC number in profile, got %s", profile.CKYCNumber)
	}
	if profile.AssignedTier != KYCTier2FullCKYC {
		t.Errorf("expected tier upgrade to KYCTier2FullCKYC, got %v", profile.AssignedTier)
	}
	if profile.RiskTier != RiskTierLow {
		t.Errorf("expected RiskTierLow, got %s", profile.RiskTier)
	}
}
