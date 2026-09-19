package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Standard statutory thresholds under PMLA 2002 for FIU-IND
const (
	FIU_CTR_THRESHOLD_INR      = 1000000.0 // ₹10 Lakhs statutory threshold for CTR
	FIU_STRUCTURING_LOWER_INR  = 850000.0  // Lower boundary for smurfing/structuring surveillance
	FIU_STRUCTURING_UPPER_INR  = 999999.0  // Upper boundary just below ₹10 Lakhs
)

var ckycRegex = regexp.MustCompile(`^[0-9]{14}$`)

// FIUReportType defines statutory reporting filing types
type FIUReportType string

const (
	ReportTypeCTR FIUReportType = "CTR" // Cash Transaction Report
	ReportTypeSTR FIUReportType = "STR" // Suspicious Transaction Report
)

// SuspicionCategory defines statutory grounds for STR under PMLA
type SuspicionCategory string

const (
	CategoryStructuring        SuspicionCategory = "STRUCTURING"
	CategoryRapidLayering      SuspicionCategory = "RAPID_LAYERING"
	CategoryDarknetMixer       SuspicionCategory = "DARKNET_OR_MIXER"
	CategorySanctionHit        SuspicionCategory = "SANCTIONS_OR_OFAC_HIT"
	CategoryAbnormalKYCDeviation SuspicionCategory = "ABNORMAL_KYC_PROFILE_DEVIATION"
)

// FIUBatchHeader represents the FINnet 2.0 batch transmission header
type FIUBatchHeader struct {
	XMLName              xml.Name      `xml:"BatchHeader" json:"-"`
	BatchID              string        `xml:"BatchID" json:"batch_id"`
	ReportType           FIUReportType `xml:"ReportType" json:"report_type"`
	ReportingEntityID    string        `xml:"ReportingEntityID" json:"reporting_entity_id"` // E.g. "INRE00099881"
	ReportingEntityName  string        `xml:"ReportingEntityName" json:"reporting_entity_name"`
	ReportingCategory    string        `xml:"ReportingCategory" json:"reporting_category"` // "VDASP"
	BatchDate            string        `xml:"BatchDate" json:"batch_date"`
	MonthYear            string        `xml:"MonthYear" json:"month_year"` // MMYYYY
	RecordCount          int           `xml:"RecordCount" json:"record_count"`
}

// FIUSuspectProfile captures customer identity details under PMLA Rule 3
type FIUSuspectProfile struct {
	XMLName           xml.Name `xml:"SuspectProfile" json:"-"`
	CustomerInternalID string   `xml:"CustomerInternalID" json:"customer_internal_id"`
	FullName          string   `xml:"FullName" json:"full_name"`
	PAN               string   `xml:"PAN" json:"pan"`
	AadhaarHash       string   `xml:"AadhaarHash" json:"aadhaar_hash"`
	CKYCNumber        string   `xml:"CKYCNumber,omitempty" json:"ckyc_number,omitempty"`
	DateOfBirth       string   `xml:"DateOfBirth" json:"date_of_birth"`
	Nationality       string   `xml:"Nationality" json:"nationality"` // ISO-3166 "IND"
	RiskCategory      string   `xml:"RiskCategory" json:"risk_category"` // "LOW", "MEDIUM", "HIGH", "PEP"
	AddressLine       string   `xml:"AddressLine" json:"address_line"`
	City              string   `xml:"City" json:"city"`
	StateCode         string   `xml:"StateCode" json:"state_code"` // E.g. "KA", "MH"
	PinCode           string   `xml:"PinCode" json:"pin_code"`
	MobileNumber      string   `xml:"MobileNumber" json:"mobile_number"`
	Email             string   `xml:"Email" json:"email"`
}

// FIUTransactionRecord represents an individual transacted line item
type FIUTransactionRecord struct {
	XMLName             xml.Name `xml:"Transaction" json:"-"`
	TransactionID       string   `xml:"TransactionID" json:"transaction_id"`
	TransactionDate     string   `xml:"TransactionDate" json:"transaction_date"` // YYYY-MM-DD
	TransactionTime     string   `xml:"TransactionTime" json:"transaction_time"` // HH:MM:SS
	TransactionMode     string   `xml:"TransactionMode" json:"transaction_mode"` // "CRYPTO_ONCHAIN", "IMPS", "NEFT", "CBDC"
	TransactionType     string   `xml:"TransactionType" json:"transaction_type"` // "DEPOSIT", "WITHDRAWAL", "BUY", "SELL", "P2P_ESCROW"
	AmountINR           float64  `xml:"AmountINR" json:"amount_inr"`
	AssetCode           string   `xml:"AssetCode" json:"asset_code"`             // "BTC", "ETH", "USDT", "INR"
	AssetQuantity       float64  `xml:"AssetQuantity" json:"asset_quantity"`
	TxHash              string   `xml:"TxHash,omitempty" json:"tx_hash,omitempty"`
	SourceWallet        string   `xml:"SourceWallet,omitempty" json:"source_wallet,omitempty"`
	DestinationWallet   string   `xml:"DestinationWallet,omitempty" json:"destination_wallet,omitempty"`
	IPAddress           string   `xml:"IPAddress" json:"ip_address"`
}

// FIUSuspicionGrounds details the statutory rationale for filing an STR under Section 12 PMLA
type FIUSuspicionGrounds struct {
	XMLName               xml.Name          `xml:"SuspicionGrounds" json:"-"`
	SuspicionCategory     SuspicionCategory `xml:"SuspicionCategory" json:"suspicion_category"`
	DetailedNarrative     string            `xml:"DetailedNarrative" json:"detailed_narrative"`
	ActionTaken           string            `xml:"ActionTaken" json:"action_taken"` // "FROZEN", "ACCOUNT_LOCKED", "MONITORED"
	InvestigatorReference string            `xml:"InvestigatorRef" json:"investigator_ref"`
}

// FIUReportFiling aggregates a full regulatory report (CTR or STR)
type FIUReportFiling struct {
	XMLName          xml.Name               `xml:"FIUReportFiling" json:"-"`
	Header           FIUBatchHeader         `xml:"BatchHeader" json:"header"`
	Suspect          FIUSuspectProfile      `xml:"SuspectProfile" json:"suspect"`
	Transactions     []FIUTransactionRecord `xml:"Transactions>Transaction" json:"transactions"`
	SuspicionGrounds *FIUSuspicionGrounds   `xml:"SuspicionGrounds,omitempty" json:"suspicion_grounds,omitempty"`
}

// FIUSchemaValidationError captures detailed schema validation failure
type FIUSchemaValidationError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// FIUReportingEngine manages CTR/STR generation and statutory schema validation
type FIUReportingEngine struct {
	mu                  sync.Mutex
	reportingEntityID   string
	reportingEntityName string
	filings             map[string]*FIUReportFiling
}

func NewFIUReportingEngine(entityID, entityName string) *FIUReportingEngine {
	return &FIUReportingEngine{
		reportingEntityID:   entityID,
		reportingEntityName: entityName,
		filings:             make(map[string]*FIUReportFiling),
	}
}

// ValidateReport performs rigorous schema validation against FIU-IND FINnet 2.0 specifications
func (e *FIUReportingEngine) ValidateReport(report *FIUReportFiling) []FIUSchemaValidationError {
	var errs []FIUSchemaValidationError

	// 1. Header validations
	if strings.TrimSpace(report.Header.BatchID) == "" {
		errs = append(errs, FIUSchemaValidationError{
			Field:   "Header.BatchID",
			Rule:    "REQUIRED",
			Message: "BatchID must not be empty",
		})
	}
	if report.Header.ReportType != ReportTypeCTR && report.Header.ReportType != ReportTypeSTR {
		errs = append(errs, FIUSchemaValidationError{
			Field:   "Header.ReportType",
			Rule:    "INVALID_ENUM",
			Message: fmt.Sprintf("ReportType must be CTR or STR, got '%s'", report.Header.ReportType),
		})
	}
	if strings.TrimSpace(report.Header.ReportingEntityID) == "" {
		errs = append(errs, FIUSchemaValidationError{
			Field:   "Header.ReportingEntityID",
			Rule:    "REQUIRED",
			Message: "ReportingEntityID must not be empty",
		})
	}

	// 2. Suspect profile validations
	if !ValidatePAN(report.Suspect.PAN) {
		errs = append(errs, FIUSchemaValidationError{
			Field:   "Suspect.PAN",
			Rule:    "INVALID_FORMAT",
			Message: fmt.Sprintf("Invalid Indian Income Tax PAN format: '%s'", report.Suspect.PAN),
		})
	}
	if strings.TrimSpace(report.Suspect.FullName) == "" {
		errs = append(errs, FIUSchemaValidationError{
			Field:   "Suspect.FullName",
			Rule:    "REQUIRED",
			Message: "Suspect FullName must not be empty",
		})
	}
	if report.Suspect.CKYCNumber != "" && !ckycRegex.MatchString(report.Suspect.CKYCNumber) {
		errs = append(errs, FIUSchemaValidationError{
			Field:   "Suspect.CKYCNumber",
			Rule:    "INVALID_FORMAT",
			Message: fmt.Sprintf("CKYC must be 14 digits, got '%s'", report.Suspect.CKYCNumber),
		})
	}

	// 3. Transactions validations
	if len(report.Transactions) == 0 {
		errs = append(errs, FIUSchemaValidationError{
			Field:   "Transactions",
			Rule:    "MIN_ITEMS",
			Message: "Report must contain at least one transaction record",
		})
	}

	var totalAmountINR float64
	for i, tx := range report.Transactions {
		if tx.AmountINR <= 0 {
			errs = append(errs, FIUSchemaValidationError{
				Field:   fmt.Sprintf("Transactions[%d].AmountINR", i),
				Rule:    "GREATER_THAN_ZERO",
				Message: fmt.Sprintf("Transaction amount must be strictly positive, got %.2f", tx.AmountINR),
			})
		}
		if strings.TrimSpace(tx.TransactionID) == "" {
			errs = append(errs, FIUSchemaValidationError{
				Field:   fmt.Sprintf("Transactions[%d].TransactionID", i),
				Rule:    "REQUIRED",
				Message: "TransactionID must not be empty",
			})
		}
		totalAmountINR += tx.AmountINR
	}

	// 4. CTR specific statutory threshold rule (>= ₹10 Lakhs)
	if report.Header.ReportType == ReportTypeCTR {
		if totalAmountINR < FIU_CTR_THRESHOLD_INR {
			errs = append(errs, FIUSchemaValidationError{
				Field:   "Transactions.TotalAmountINR",
				Rule:    "CTR_THRESHOLD_NOT_MET",
				Message: fmt.Sprintf("CTR filing requires aggregate amount >= ₹10,00,000 (PMLA Rule 3), got ₹%.2f", totalAmountINR),
			})
		}
	}

	// 5. STR specific statutory requirements
	if report.Header.ReportType == ReportTypeSTR {
		if report.SuspicionGrounds == nil {
			errs = append(errs, FIUSchemaValidationError{
				Field:   "SuspicionGrounds",
				Rule:    "REQUIRED_FOR_STR",
				Message: "SuspicionGrounds is mandatory for all STR filings under Section 12 PMLA",
			})
		} else {
			if len(strings.TrimSpace(report.SuspicionGrounds.DetailedNarrative)) < 20 {
				errs = append(errs, FIUSchemaValidationError{
					Field:   "SuspicionGrounds.DetailedNarrative",
					Rule:    "MIN_LENGTH",
					Message: "Detailed narrative must be at least 20 characters explaining the suspicion basis",
				})
			}
			if report.SuspicionGrounds.SuspicionCategory == "" {
				errs = append(errs, FIUSchemaValidationError{
					Field:   "SuspicionGrounds.SuspicionCategory",
					Rule:    "REQUIRED",
					Message: "Suspicion category must be specified",
				})
			}
		}
	}

	return errs
}

// GenerateCTR creates a statutory Cash Transaction Report for transactions exceeding ₹10 Lakhs
func (e *FIUReportingEngine) GenerateCTR(batchID string, suspect FIUSuspectProfile, txs []FIUTransactionRecord) (*FIUReportFiling, error) {
	now := time.Now().UTC()
	header := FIUBatchHeader{
		BatchID:             batchID,
		ReportType:          ReportTypeCTR,
		ReportingEntityID:   e.reportingEntityID,
		ReportingEntityName: e.reportingEntityName,
		ReportingCategory:   "VDASP",
		BatchDate:           now.Format("2006-01-02"),
		MonthYear:           now.Format("012006"),
		RecordCount:         len(txs),
	}

	report := &FIUReportFiling{
		Header:       header,
		Suspect:      suspect,
		Transactions: txs,
	}

	valErrs := e.ValidateReport(report)
	if len(valErrs) > 0 {
		return nil, fmt.Errorf("CTR schema validation failed [%s]: %s", valErrs[0].Rule, valErrs[0].Message)
	}

	e.mu.Lock()
	e.filings[batchID] = report
	e.mu.Unlock()

	return report, nil
}

// GenerateSTR creates a statutory Suspicious Transaction Report
func (e *FIUReportingEngine) GenerateSTR(batchID string, suspect FIUSuspectProfile, txs []FIUTransactionRecord, grounds FIUSuspicionGrounds) (*FIUReportFiling, error) {
	now := time.Now().UTC()
	header := FIUBatchHeader{
		BatchID:             batchID,
		ReportType:          ReportTypeSTR,
		ReportingEntityID:   e.reportingEntityID,
		ReportingEntityName: e.reportingEntityName,
		ReportingCategory:   "VDASP",
		BatchDate:           now.Format("2006-01-02"),
		MonthYear:           now.Format("012006"),
		RecordCount:         len(txs),
	}

	report := &FIUReportFiling{
		Header:           header,
		Suspect:          suspect,
		Transactions:     txs,
		SuspicionGrounds: &grounds,
	}

	valErrs := e.ValidateReport(report)
	if len(valErrs) > 0 {
		return nil, fmt.Errorf("STR schema validation failed [%s]: %s", valErrs[0].Rule, valErrs[0].Message)
	}

	e.mu.Lock()
	e.filings[batchID] = report
	e.mu.Unlock()

	return report, nil
}

// DetectStructuring analyzes transaction amounts to identify structuring (smurfing) patterns just below ₹10 Lakhs
func DetectStructuring(amountsINR []float64) (bool, int, float64) {
	var countNearLimit int
	var total float64

	for _, amt := range amountsINR {
		total += amt
		if amt >= FIU_STRUCTURING_LOWER_INR && amt <= FIU_STRUCTURING_UPPER_INR {
			countNearLimit++
		}
	}

	// If 2 or more transactions fall between ₹8.5L and ₹9.99L, flag structuring
	isStructuring := countNearLimit >= 2
	return isStructuring, countNearLimit, total
}

// ExportToXML serializes the filing into standard FIU-IND XML schema format
func (r *FIUReportFiling) ExportToXML() ([]byte, error) {
	valErrs := (&FIUReportingEngine{}).ValidateReport(r)
	if len(valErrs) > 0 {
		return nil, errors.New("cannot export invalid report to XML: " + valErrs[0].Message)
	}
	return xml.MarshalIndent(r, "", "  ")
}

// ExportToJSON serializes the filing into FINnet 2.0 JSON format
func (r *FIUReportFiling) ExportToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
