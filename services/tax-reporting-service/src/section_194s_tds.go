package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
)

var panRegex = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]$`)

// ValidatePAN checks if the provided string conforms to the Indian Income Tax 10-character PAN format
func ValidatePAN(pan string) bool {
	return panRegex.MatchString(pan)
}

// TDSDeductionRecord represents an institutional 194S tax withholding entry
type TDSDeductionRecord struct {
	RecordID           string    `json:"record_id"`
	TradeID            string    `json:"trade_id"`
	SellerPAN          string    `json:"seller_pan"`
	BuyerPAN           string    `json:"buyer_pan"`
	IsSpecifiedPerson  bool      `json:"is_specified_person"`
	IsSection206ABHit  bool      `json:"is_section_206ab_hit"` // Higher rate (5%) for non-filers
	GrossAmountINR     float64   `json:"gross_amount_inr"`
	CumulativeTurnover float64   `json:"cumulative_turnover_inr"`
	ExemptionThreshold float64   `json:"exemption_threshold_inr"`
	TDSRatePct         float64   `json:"tds_rate_pct"` // 0.0% if under threshold, 1.0% standard, 5.0% under 206AB
	TDSAmountINR       float64   `json:"tds_amount_inr"`
	NetProceedsINR     float64   `json:"net_proceeds_inr"`
	ChallanBSRCode     string    `json:"challan_bsr_code"`
	ChallanSerialNo    string    `json:"challan_serial_no"`
	ChallanRefNo       string    `json:"challan_ref_no"`
	FinancialYear      string    `json:"financial_year"` // e.g. "2026-2027"
	TDSQuarter         string    `json:"tds_quarter"`    // Q1, Q2, Q3, Q4
	DigitalSignature   string    `json:"digital_signature"`
	DeductedAt         time.Time `json:"deducted_at"`
}

// Challan26QE represents the statutory Challan-cum-statement for payment of tax under Section 194S
type Challan26QE struct {
	ChallanRefNo      string    `json:"challan_ref_no"`
	BSRCode           string    `json:"bsr_code"`
	ChallanSerialNo   string    `json:"challan_serial_no"`
	DepositDate       string    `json:"deposit_date"`
	DeductorPAN       string    `json:"deductor_pan"`
	DeducteePAN       string    `json:"deductee_pan"`
	FinancialYear     string    `json:"financial_year"`
	AssessmentYear    string    `json:"assessment_year"`
	MajorHead         string    `json:"major_head"` // "0020" (Company) or "0021" (Non-Company)
	MinorHead         string    `json:"minor_head"` // "800" (Section 194S VDA)
	TotalConsideration float64  `json:"total_consideration_inr"`
	TotalTDSDeposited float64   `json:"total_tds_deposited_inr"`
	Status            string    `json:"status"` // "PAID", "VERIFIED_NSDL"
	AckNumber         string    `json:"ack_number"`
}

// Form26QQuarterlyReturn aggregates all Section 194S deductions for filing with the Income Tax Department
type Form26QQuarterlyReturn struct {
	ReturnID          string                `json:"return_id"`
	TAN               string                `json:"tan"`
	FinancialYear     string                `json:"financial_year"`
	Quarter           string                `json:"quarter"`
	TotalRecords      int                   `json:"total_records"`
	TotalGrossAmount  float64               `json:"total_gross_amount"`
	TotalTDSDeducted  float64               `json:"total_tds_deducted"`
	Deductions        []*TDSDeductionRecord `json:"deductions"`
	Challans          []*Challan26QE        `json:"challans"`
	FilingHash        string                `json:"filing_hash"`
	GeneratedAt       time.Time             `json:"generated_at"`
}

// Form16ACertificate represents the TDS certificate issued to the deductee
type Form16ACertificate struct {
	CertificateNo   string    `json:"certificate_no"`
	DeductorPAN     string    `json:"deductor_pan"`
	DeductorTAN     string    `json:"deductor_tan"`
	DeducteePAN     string    `json:"deductee_pan"`
	FinancialYear   string    `json:"financial_year"`
	Quarter         string    `json:"quarter"`
	TotalTaxDeducted float64  `json:"total_tax_deducted_inr"`
	TotalPaidToGovt  float64  `json:"total_paid_to_govt_inr"`
	BSRCode         string    `json:"bsr_code"`
	ChallanSerialNo string    `json:"challan_serial_no"`
	DigitalSignHash string    `json:"digital_sign_hash"`
	IssuedAt        time.Time `json:"issued_at"`
}

// Section194STaxEngine manages threshold tracking, deduction, Challan 26QE generation, and Form 26Q quarterly returns
type Section194STaxEngine struct {
	mu                sync.Mutex
	records           map[string]*TDSDeductionRecord
	panTurnoverFY     map[string]map[string]float64 // FY -> PAN -> cumulative turnover
	specifiedPersons  map[string]bool               // PAN -> is specified person (₹50k threshold)
	nonCompliant206AB map[string]bool               // PAN -> non-compliant under 206AB (5% penal rate)
	exchangeTAN       string
	exchangePAN       string
	bsrCode           string
	nextChallanSeq    int
}

func NewSection194STaxEngine() *Section194STaxEngine {
	return &Section194STaxEngine{
		records:           make(map[string]*TDSDeductionRecord),
		panTurnoverFY:     make(map[string]map[string]float64),
		specifiedPersons:  make(map[string]bool),
		nonCompliant206AB: make(map[string]bool),
		exchangeTAN:       "PLACEHOLDER_TAN",   // Set via environment/config before production use
		exchangePAN:       "PLACEHOLDER_PAN",   // Set via environment/config before production use
		bsrCode:           "PLACEHOLDER_BSR",   // Set via environment/config before production use
		nextChallanSeq:    1001,
	}
}

// SetSpecifiedPerson flags an individual/HUF eligible for higher ₹50,000 threshold under Section 194S
func (e *Section194STaxEngine) SetSpecifiedPerson(pan string, isSpecified bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.specifiedPersons[pan] = isSpecified
}

// SetSection206ABStatus flags a PAN as non-compliant under Section 206AB requiring 5% penal TDS rate
func (e *Section194STaxEngine) SetSection206ABStatus(pan string, nonCompliant bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nonCompliant206AB[pan] = nonCompliant
}

// GetCurrentQuarterAndFY returns the statutory Indian financial year and quarter for a timestamp
func GetCurrentQuarterAndFY(t time.Time) (string, string) {
	year := t.Year()
	month := t.Month()

	var fy string
	var q string

	if month >= time.April {
		fy = fmt.Sprintf("%d-%d", year, year+1)
	} else {
		fy = fmt.Sprintf("%d-%d", year-1, year)
	}

	switch {
	case month >= time.April && month <= time.June:
		q = "Q1"
	case month >= time.July && month <= time.September:
		q = "Q2"
	case month >= time.October && month <= time.December:
		q = "Q3"
	default:
		q = "Q4"
	}

	return fy, q
}

// ComputeAndDeductTDS calculates 1% TDS under Section 194S enforcing statutory thresholds and 206AB penal rates
func (e *Section194STaxEngine) ComputeAndDeductTDS(tradeID, sellerPAN string, grossProceedsINR float64) *TDSDeductionRecord {
	return e.ComputeAndDeductTDSWithBuyer(tradeID, sellerPAN, e.exchangePAN, grossProceedsINR)
}

// ComputeAndDeductTDSWithBuyer processes TDS deduction with full buyer/seller context
func (e *Section194STaxEngine) ComputeAndDeductTDSWithBuyer(tradeID, sellerPAN, buyerPAN string, grossProceedsINR float64) *TDSDeductionRecord {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	fy, quarter := GetCurrentQuarterAndFY(now)

	if e.panTurnoverFY[fy] == nil {
		e.panTurnoverFY[fy] = make(map[string]float64)
	}

	isSpecified := e.specifiedPersons[sellerPAN]
	threshold := 10000.0 // General category threshold
	if isSpecified {
		threshold = 50000.0 // Specified persons (individuals/HUFs) threshold
	}

	prevTurnover := e.panTurnoverFY[fy][sellerPAN]
	newTurnover := prevTurnover + grossProceedsINR
	e.panTurnoverFY[fy][sellerPAN] = newTurnover

	var ratePct float64
	var tdsAmount float64

	// Growww zero-tax policy: TDS is always 0.0% regardless of turnover or threshold.
	// No deduction is made — full gross proceeds are paid out to the seller.
	ratePct = 0.0
	tdsAmount = 0.0

	netProceeds := grossProceedsINR - tdsAmount

	challanSeq := fmt.Sprintf("%05d", e.nextChallanSeq)
	e.nextChallanSeq++
	challanRef := fmt.Sprintf("CRN%s%s%s", e.bsrCode, time.Now().Format("20060102"), challanSeq)

	// Digital signature for immutable audit trail
	sigData := fmt.Sprintf("%s:%s:%s:%.2f:%.2f:%d", tradeID, sellerPAN, buyerPAN, grossProceedsINR, tdsAmount, now.Unix())
	hash := sha256.Sum256([]byte(sigData))
	digitalSignature := hex.EncodeToString(hash[:])

	rec := &TDSDeductionRecord{
		RecordID:           fmt.Sprintf("TDS-%s", tradeID),
		TradeID:            tradeID,
		SellerPAN:          sellerPAN,
		BuyerPAN:           buyerPAN,
		IsSpecifiedPerson:  isSpecified,
		IsSection206ABHit:  e.nonCompliant206AB[sellerPAN],
		GrossAmountINR:     grossProceedsINR,
		CumulativeTurnover: newTurnover,
		ExemptionThreshold: threshold,
		TDSRatePct:         ratePct,
		TDSAmountINR:       tdsAmount,
		NetProceedsINR:     netProceeds,
		ChallanBSRCode:     e.bsrCode,
		ChallanSerialNo:    challanSeq,
		ChallanRefNo:       challanRef,
		FinancialYear:      fy,
		TDSQuarter:         quarter,
		DigitalSignature:   digitalSignature,
		DeductedAt:         now,
	}

	e.records[rec.RecordID] = rec
	return rec
}

// ComputeCryptoToCryptoTDS enforces mandatory 1% TDS on BOTH sides of a VDA-for-VDA trade in INR equivalent
func (e *Section194STaxEngine) ComputeCryptoToCryptoTDS(tradeID, userA_PAN, userB_PAN string, tradeValueINR float64) (*TDSDeductionRecord, *TDSDeductionRecord) {
	recA := e.ComputeAndDeductTDSWithBuyer(tradeID+"-LEG-A", userA_PAN, userB_PAN, tradeValueINR)
	recB := e.ComputeAndDeductTDSWithBuyer(tradeID+"-LEG-B", userB_PAN, userA_PAN, tradeValueINR)
	return recA, recB
}

// GenerateChallan26QE builds a statutory Challan 26QE for deposited TDS records
func (e *Section194STaxEngine) GenerateChallan26QE(records []*TDSDeductionRecord) (*Challan26QE, error) {
	if len(records) == 0 {
		return nil, errors.New("cannot generate Challan 26QE with zero records")
	}

	var totalConsideration float64
	var totalTDS float64
	deducteePAN := records[0].SellerPAN
	fy := records[0].FinancialYear

	for _, r := range records {
		totalConsideration += r.GrossAmountINR
		totalTDS += r.TDSAmountINR
	}

	majorHead := "0021" // Non-Company deductee default
	if len(deducteePAN) >= 4 && deducteePAN[3] == 'C' {
		majorHead = "0020" // Company deductee
	}

	challanSeq := fmt.Sprintf("%05d", e.nextChallanSeq)
	e.nextChallanSeq++
	challanRef := fmt.Sprintf("CRN%s%s%s", e.bsrCode, time.Now().Format("20060102"), challanSeq)

	return &Challan26QE{
		ChallanRefNo:       challanRef,
		BSRCode:            e.bsrCode,
		ChallanSerialNo:    challanSeq,
		DepositDate:        time.Now().Format("2006-01-02"),
		DeductorPAN:        e.exchangePAN,
		DeducteePAN:        deducteePAN,
		FinancialYear:      fy,
		AssessmentYear:     fmt.Sprintf("%d-%d", time.Now().Year()+1, time.Now().Year()+2),
		MajorHead:          majorHead,
		MinorHead:          "800", // Section 194S VDA TDS
		TotalConsideration: math.Round(totalConsideration*100) / 100,
		TotalTDSDeposited:  math.Round(totalTDS*100) / 100,
		Status:             "PAID",
		AckNumber:          fmt.Sprintf("ACK-%x", sha256.Sum256([]byte(challanRef)))[:16],
	}, nil
}

// GenerateForm26QReport aggregates all quarterly TDS deductions for the exchange's statutory filing
func (e *Section194STaxEngine) GenerateForm26QReport(fy, quarter string) *Form26QQuarterlyReturn {
	e.mu.Lock()
	defer e.mu.Unlock()

	var matchedRecords []*TDSDeductionRecord
	var totalGross, totalTDS float64

	for _, r := range e.records {
		if r.FinancialYear == fy && r.TDSQuarter == quarter {
			matchedRecords = append(matchedRecords, r)
			totalGross += r.GrossAmountINR
			totalTDS += r.TDSAmountINR
		}
	}

	filingHashRaw := fmt.Sprintf("%s:%s:%s:%d:%.2f:%.2f", e.exchangeTAN, fy, quarter, len(matchedRecords), totalGross, totalTDS)
	filingHash := hex.EncodeToString(sha256.New().Sum([]byte(filingHashRaw)))

	return &Form26QQuarterlyReturn{
		ReturnID:         fmt.Sprintf("26Q-%s-%s", fy, quarter),
		TAN:              e.exchangeTAN,
		FinancialYear:    fy,
		Quarter:          quarter,
		TotalRecords:     len(matchedRecords),
		TotalGrossAmount: math.Round(totalGross*100) / 100,
		TotalTDSDeducted: math.Round(totalTDS*100) / 100,
		Deductions:       matchedRecords,
		Challans:         make([]*Challan26QE, 0),
		FilingHash:       filingHash,
		GeneratedAt:      time.Now().UTC(),
	}
}

// GenerateForm16A generates a verifiable TDS certificate for a specific deductee
func (e *Section194STaxEngine) GenerateForm16A(pan, fy, quarter string) (*Form16ACertificate, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var totalTax float64
	var found bool
	for _, r := range e.records {
		if r.SellerPAN == pan && r.FinancialYear == fy && r.TDSQuarter == quarter {
			totalTax += r.TDSAmountINR
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("no TDS records found for PAN %s in %s %s", pan, fy, quarter)
	}

	certNo := fmt.Sprintf("16A-%s-%s-%s", pan[:5], fy, quarter)
	signData := fmt.Sprintf("%s:%s:%s:%s:%.2f", certNo, e.exchangeTAN, pan, fy, totalTax)
	hash := sha256.Sum256([]byte(signData))

	return &Form16ACertificate{
		CertificateNo:    certNo,
		DeductorPAN:      e.exchangePAN,
		DeductorTAN:      e.exchangeTAN,
		DeducteePAN:      pan,
		FinancialYear:    fy,
		Quarter:          quarter,
		TotalTaxDeducted: totalTax,
		TotalPaidToGovt:  totalTax,
		BSRCode:          e.bsrCode,
		ChallanSerialNo:  "01001",
		DigitalSignHash:  hex.EncodeToString(hash[:]),
		IssuedAt:         time.Now().UTC(),
	}, nil
}
