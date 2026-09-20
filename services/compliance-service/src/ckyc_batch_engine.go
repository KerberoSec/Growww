package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// CKYCBatchStatus represents batch processing lifecycle state
type CKYCBatchStatus string

const (
	CKYCStatusDraft          CKYCBatchStatus = "DRAFT"
	CKYCStatusValidated      CKYCBatchStatus = "VALIDATED"
	CKYCStatusQueued         CKYCBatchStatus = "QUEUED"
	CKYCStatusUploaded       CKYCBatchStatus = "UPLOADED"
	CKYCStatusAckReceived    CKYCBatchStatus = "ACK_RECEIVED"
	CKYCStatusSuccess        CKYCBatchStatus = "PROCESSED_SUCCESS"
	CKYCStatusPartialSuccess CKYCBatchStatus = "PARTIAL_SUCCESS"
	CKYCStatusRejected       CKYCBatchStatus = "REJECTED"
	CKYCStatusRetryScheduled CKYCBatchStatus = "RETRY_SCHEDULED"
)

// DocumentType represents PoI / PoA classification under CERSAI guidelines
type DocumentType string

const (
	DocTypePAN            DocumentType = "PAN"
	DocTypePassport       DocumentType = "PASSPORT"
	DocTypeVoterID        DocumentType = "VOTER_ID"
	DocTypeDrivingLicense DocumentType = "DRIVING_LICENSE"
	DocTypeAadhaarOffline DocumentType = "AADHAAR_OFFLINE"
)

// CKYCInvestorRecord represents a single applicant's demographic and KYC payload
type CKYCInvestorRecord struct {
	InvestorUUID     string       `xml:"InvestorUUID"`
	FullName         string       `xml:"FullName"`
	Gender           string       `xml:"Gender"` // M, F, T
	DOB              string       `xml:"DateOfBirth"` // DD-MM-YYYY
	PAN              string       `xml:"PAN"`
	MaskedAadhaar    string       `xml:"MaskedAadhaar"` // XXXX-XXXX-1234
	FatherSpouseName string       `xml:"FatherOrSpouseName"`
	MotherName       string       `xml:"MotherName"`
	Occupation       string       `xml:"Occupation"`
	Email            string       `xml:"Email"`
	Mobile           string       `xml:"Mobile"`
	AddressLine1     string       `xml:"AddressLine1"`
	AddressLine2     string       `xml:"AddressLine2,omitempty"`
	City             string       `xml:"City"`
	StateCode        string       `xml:"StateCode"` // e.g. MH, DL, KA
	PINCode          string       `xml:"PINCode"` // 6 digits
	CountryCode      string       `xml:"CountryCode"` // IND
	PoIType          DocumentType `xml:"PoIType"`
	PoINumber        string       `xml:"PoINumber"`
	PoAType          DocumentType `xml:"PoAType"`
	PoANumber        string       `xml:"PoANumber"`
	PhotoFileName    string       `xml:"PhotoFileName"`
	SignatureFileName string      `xml:"SignatureFileName"`
	LivenessScore    float64      `xml:"LivenessScore"`
}

// CERSAI Batch Upload XML Schema
type CKYCBatchUploadXML struct {
	XMLName    xml.Name               `xml:"CKYCBatch"`
	Namespace  string                 `xml:"xmlns,attr"`
	Header     CKYCBatchHeaderXML     `xml:"Header"`
	Records    []CKYCInvestorRecord   `xml:"Records>Record"`
	BatchDigest string                `xml:"BatchDigest"`
}

type CKYCBatchHeaderXML struct {
	BatchID       string `xml:"BatchID"`
	FIEntityCode  string `xml:"FIEntityCode"` // Growww FI Registration Code
	BatchDate     string `xml:"BatchDate"`     // YYYY-MM-DD
	RecordCount   int    `xml:"RecordCount"`
	SchemaVersion string `xml:"SchemaVersion"`
}

// CERSAI Batch Response XML Schema
type CKYCBatchResponseXML struct {
	XMLName      xml.Name               `xml:"CKYCBatchResponse"`
	BatchID      string                 `xml:"BatchID"`
	Status       string                 `xml:"Status"` // SUCCESS, PARTIAL_SUCCESS, REJECTED
	TotalRecords int                    `xml:"TotalRecords"`
	SuccessCount int                    `xml:"SuccessCount"`
	FailureCount int                    `xml:"FailureCount"`
	Records      []CKYCRecordResponseXML `xml:"RecordResults>RecordResult"`
}

type CKYCRecordResponseXML struct {
	InvestorUUID string `xml:"InvestorUUID"`
	Status       string `xml:"Status"` // SUCCESS, FAILED
	CKYCNumber   string `xml:"CKYCNumber,omitempty"` // 14-digit KIN assigned by CERSAI
	ErrorCode    string `xml:"ErrorCode,omitempty"`
	ErrorMessage string `xml:"ErrorMessage,omitempty"`
}

// CERSAI Search / Download Request & Response Schemas
type CKYCSearchRequestXML struct {
	XMLName      xml.Name `xml:"CKYCSearchRequest"`
	FIEntityCode string   `xml:"FIEntityCode"`
	RequestID    string   `xml:"RequestID"`
	CKYCNumber   string   `xml:"CKYCNumber,omitempty"` // 14-digit KIN if known
	PAN          string   `xml:"PAN,omitempty"`
	DOB          string   `xml:"DateOfBirth,omitempty"` // DD-MM-YYYY
}

type CKYCSearchResponseXML struct {
	XMLName          xml.Name `xml:"CKYCSearchResponse"`
	RequestID        string   `xml:"RequestID"`
	Status           string   `xml:"Status"` // FOUND, NOT_FOUND, ERROR
	CKYCNumber       string   `xml:"CKYCNumber"`
	FullName         string   `xml:"FullName"`
	PAN              string   `xml:"PAN"`
	MaskedAadhaar    string   `xml:"MaskedAadhaar"`
	Gender           string   `xml:"Gender"`
	DOB              string   `xml:"DateOfBirth"`
	KYCDate          string   `xml:"KYCDate"`
	VerifyingFI      string   `xml:"VerifyingFIEntityCode"`
	KYCStatusDesc    string   `xml:"KYCStatusDesc"`
	RiskCategory     string   `xml:"RiskCategory"`
	AddressLine1     string   `xml:"AddressLine1"`
	City             string   `xml:"City"`
	StateCode        string   `xml:"StateCode"`
	PINCode          string   `xml:"PINCode"`
}

// CKYCBatchEngine orchestrates CKYC batch uploads, validations, responses, and registry fetches
type CKYCBatchEngine struct {
	mu           sync.RWMutex
	fiEntityCode string
	batches      map[string]*CKYCBatchState
	panRegex     *regexp.Regexp
	pinRegex     *regexp.Regexp
	kinRegex     *regexp.Regexp
}

type CKYCBatchState struct {
	BatchID        string
	Status         CKYCBatchStatus
	Records        []CKYCInvestorRecord
	XMLPayload     string
	DigestHex      string
	CreatedAt      time.Time
	UploadedAt     time.Time
	ProcessedAt    time.Time
	SuccessCount   int
	FailureCount   int
	AssignedKINs   map[string]string // InvestorUUID -> 14-digit KIN
	Errors         map[string]string // InvestorUUID -> ErrorMessage
}

func NewCKYCBatchEngine(fiEntityCode string) *CKYCBatchEngine {
	return &CKYCBatchEngine{
		fiEntityCode: fiEntityCode,
		batches:      make(map[string]*CKYCBatchState),
		panRegex:     regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`),
		pinRegex:     regexp.MustCompile(`^[1-9][0-9]{5}$`),
		kinRegex:     regexp.MustCompile(`^[0-9]{14}$`),
	}
}

// ValidateRecord checks strict CERSAI statutory KYC field constraints
func (e *CKYCBatchEngine) ValidateRecord(rec CKYCInvestorRecord) error {
	if strings.TrimSpace(rec.InvestorUUID) == "" {
		return errors.New("InvestorUUID is mandatory")
	}
	if strings.TrimSpace(rec.FullName) == "" {
		return errors.New("FullName is mandatory")
	}
	if !e.panRegex.MatchString(rec.PAN) {
		return fmt.Errorf("invalid PAN format: %s", rec.PAN)
	}
	if !e.pinRegex.MatchString(rec.PINCode) {
		return fmt.Errorf("invalid PIN code: %s (must be 6 numeric digits)", rec.PINCode)
	}
	if rec.Gender != "M" && rec.Gender != "F" && rec.Gender != "T" {
		return fmt.Errorf("invalid Gender: %s (must be M, F, or T)", rec.Gender)
	}
	if strings.TrimSpace(rec.AddressLine1) == "" || strings.TrimSpace(rec.City) == "" || strings.TrimSpace(rec.StateCode) == "" {
		return errors.New("incomplete address: Line1, City, and StateCode are mandatory")
	}
	if rec.CountryCode != "IND" && rec.CountryCode != "IN" {
		return fmt.Errorf("invalid country code for domestic CKYC: %s", rec.CountryCode)
	}
	if rec.LivenessScore < 0.85 {
		return fmt.Errorf("insufficient biometric liveness score: %.2f (minimum 0.85 required)", rec.LivenessScore)
	}
	return nil
}

// CreateBatch bundles a group of validated investor records into an uploadable batch
func (e *CKYCBatchEngine) CreateBatch(records []CKYCInvestorRecord) (*CKYCBatchState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(records) == 0 {
		return nil, errors.New("cannot create empty CKYC batch")
	}

	// Validate all records before batch creation
	for i, rec := range records {
		if err := e.ValidateRecord(rec); err != nil {
			return nil, fmt.Errorf("validation error in record %d (%s): %w", i+1, rec.InvestorUUID, err)
		}
	}

	now := time.Now().UTC()
	batchID := fmt.Sprintf("CKYC-B-%s-%d", e.fiEntityCode, now.UnixNano())

	// Build CERSAI XML payload
	batchXML := CKYCBatchUploadXML{
		Namespace: "http://cersai.org.in/ckyc/v1.2",
		Header: CKYCBatchHeaderXML{
			BatchID:       batchID,
			FIEntityCode:  e.fiEntityCode,
			BatchDate:     now.Format("2006-01-02"),
			RecordCount:   len(records),
			SchemaVersion: "1.2",
		},
		Records: records,
	}

	rawXML, err := xml.MarshalIndent(batchXML, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CKYC batch XML: %w", err)
	}

	// Compute tamper-evident batch digest
	hash := sha256.Sum256(rawXML)
	digestHex := hex.EncodeToString(hash[:])
	batchXML.BatchDigest = digestHex

	finalXMLBytes, err := xml.MarshalIndent(batchXML, "", "  ")
	if err != nil {
		return nil, err
	}
	finalXML := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(finalXMLBytes)

	state := &CKYCBatchState{
		BatchID:      batchID,
		Status:       CKYCStatusValidated,
		Records:      records,
		XMLPayload:   finalXML,
		DigestHex:    digestHex,
		CreatedAt:    now,
		AssignedKINs: make(map[string]string),
		Errors:       make(map[string]string),
	}

	e.batches[batchID] = state
	return state, nil
}

// MarkBatchUploaded simulates/records transmission to the CERSAI gateway
func (e *CKYCBatchEngine) MarkBatchUploaded(batchID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	state, exists := e.batches[batchID]
	if !exists {
		return errors.New("batch not found")
	}
	state.Status = CKYCStatusUploaded
	state.UploadedAt = time.Now().UTC()
	return nil
}

// ProcessBatchResponseXML parses and reconciles the statutory response XML returned by CERSAI
func (e *CKYCBatchEngine) ProcessBatchResponseXML(responseXML string) (*CKYCBatchState, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var resp CKYCBatchResponseXML
	if err := xml.Unmarshal([]byte(responseXML), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal CKYC response XML: %w", err)
	}

	state, exists := e.batches[resp.BatchID]
	if !exists {
		return nil, fmt.Errorf("batch %s not found in local registry", resp.BatchID)
	}

	state.ProcessedAt = time.Now().UTC()
	state.SuccessCount = resp.SuccessCount
	state.FailureCount = resp.FailureCount

	for _, r := range resp.Records {
		if r.Status == "SUCCESS" && e.kinRegex.MatchString(r.CKYCNumber) {
			state.AssignedKINs[r.InvestorUUID] = r.CKYCNumber
		} else {
			state.Errors[r.InvestorUUID] = fmt.Sprintf("[%s] %s", r.ErrorCode, r.ErrorMessage)
		}
	}

	if resp.FailureCount == 0 {
		state.Status = CKYCStatusSuccess
	} else if resp.SuccessCount > 0 {
		state.Status = CKYCStatusPartialSuccess
	} else {
		state.Status = CKYCStatusRejected
	}

	return state, nil
}

// GenerateSearchRequestXML builds a CKYC inquiry query payload
func (e *CKYCBatchEngine) GenerateSearchRequestXML(requestID, ckycNumber, pan, dob string) (string, error) {
	if ckycNumber == "" && (pan == "" || dob == "") {
		return "", errors.New("search criteria must provide either CKYCNumber (14-digits) or PAN + DateOfBirth")
	}

	req := CKYCSearchRequestXML{
		FIEntityCode: e.fiEntityCode,
		RequestID:    requestID,
		CKYCNumber:   ckycNumber,
		PAN:          pan,
		DOB:          dob,
	}

	out, err := xml.MarshalIndent(req, "", "  ")
	if err != nil {
		return "", err
	}
	return `<?xml version="1.0" encoding="UTF-8"?>` + "\n" + string(out), nil
}

// ParseSearchResponseXML extracts investor demographic data and verifies CKYC record
func (e *CKYCBatchEngine) ParseSearchResponseXML(responseXML string) (*CKYCSearchResponseXML, error) {
	var resp CKYCSearchResponseXML
	if err := xml.Unmarshal([]byte(responseXML), &resp); err != nil {
		return nil, fmt.Errorf("invalid CKYC search response XML: %w", err)
	}

	if resp.Status != "FOUND" {
		return &resp, fmt.Errorf("CKYC record not found or error status: %s", resp.Status)
	}

	if !e.kinRegex.MatchString(resp.CKYCNumber) {
		return &resp, fmt.Errorf("invalid 14-digit CKYC number received from registry: %s", resp.CKYCNumber)
	}

	return &resp, nil
}

// MapFetchedCKYCToProfile populates a domestic investor profile with verified CKYC registry data
func (e *CKYCBatchEngine) MapFetchedCKYCToProfile(searchResp *CKYCSearchResponseXML, profile *DomesticInvestorProfile) error {
	if searchResp == nil || searchResp.Status != "FOUND" {
		return errors.New("cannot map unverified or empty CKYC search response")
	}

	profile.CKYCNumber = searchResp.CKYCNumber
	profile.DeclaredName = searchResp.FullName
	profile.PAN = searchResp.PAN
	profile.AssignedTier = KYCTier2FullCKYC // CKYC registry verification confers Tier 2

	if searchResp.RiskCategory == "HIGH" {
		profile.RiskTier = RiskTierHigh
	} else if searchResp.RiskCategory == "MEDIUM" {
		profile.RiskTier = RiskTierMedium
	} else {
		profile.RiskTier = RiskTierLow
	}

	return nil
}

// GetBatchState retrieves the state of a batch
func (e *CKYCBatchEngine) GetBatchState(batchID string) (*CKYCBatchState, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	state, exists := e.batches[batchID]
	if !exists {
		return nil, errors.New("batch not found")
	}
	return state, nil
}
