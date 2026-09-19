package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Domain types for GIFT City IFSC funding
// ──────────────────────────────────────────────────────────────────────────────

// FundingStatus tracks the lifecycle of a GIFT City funding request.
type FundingStatus string

const (
	FundingPending           FundingStatus = "PENDING"
	FundingComplianceCheck   FundingStatus = "COMPLIANCE_CHECK"
	FundingComplianceCleared FundingStatus = "COMPLIANCE_CLEARED"
	FundingCredited          FundingStatus = "CREDITED"
	FundingRejected          FundingStatus = "REJECTED"
	FundingReturned          FundingStatus = "RETURNED"
)

// IFSCAComplianceResult captures the outcome of IFSCA regulatory checks.
type IFSCAComplianceResult struct {
	KYCVerified       bool   `json:"kyc_verified"`
	LRSLimitOK        bool   `json:"lrs_limit_ok"`        // Liberalised Remittance Scheme limit check
	PurposeCodeValid  bool   `json:"purpose_code_valid"`
	SourceOfFundsOK   bool   `json:"source_of_funds_ok"`
	SanctionsScreened bool   `json:"sanctions_screened"`
	OverallPass       bool   `json:"overall_pass"`
	RejectReason      string `json:"reject_reason,omitempty"`
}

// WireIngress represents an incoming USD wire transfer to GIFT City.
type WireIngress struct {
	WireID           string  `json:"wire_id"`
	SenderName       string  `json:"sender_name"`
	SenderBankSWIFT  string  `json:"sender_bank_swift"`
	ReceiverIFSCCode string  `json:"receiver_ifsc_code"` // GIFT City IFSC branch
	AmountUSD        float64 `json:"amount_usd"`
	PurposeCode      string  `json:"purpose_code"` // RBI purpose code (e.g., "S0001")
	InvestorPAN      string  `json:"investor_pan"`
	SenderCountry    string  `json:"sender_country"`
	ReceivedAt       int64   `json:"received_at"`
}

// FundingReceipt confirms wallet crediting after compliance.
type FundingReceipt struct {
	ReceiptID     string        `json:"receipt_id"`
	WireID        string        `json:"wire_id"`
	InvestorPAN   string        `json:"investor_pan"`
	AmountUSD     float64       `json:"amount_usd"`
	WalletAddress string        `json:"wallet_address"`
	Status        FundingStatus `json:"status"`
	Compliance    IFSCAComplianceResult `json:"compliance"`
	CreditedAt    time.Time     `json:"credited_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Compliance Engine
// ──────────────────────────────────────────────────────────────────────────────

// IFSCAComplianceEngine performs regulatory checks per IFSCA guidelines.
type IFSCAComplianceEngine struct {
	lrsAnnualLimitUSD   float64
	sanctionedCountries map[string]bool
	validPurposeCodes   map[string]bool
}

// NewIFSCAComplianceEngine creates a compliance engine with regulatory defaults.
func NewIFSCAComplianceEngine() *IFSCAComplianceEngine {
	return &IFSCAComplianceEngine{
		lrsAnnualLimitUSD: 250_000.00, // RBI LRS limit
		sanctionedCountries: map[string]bool{
			"KP": true, "IR": true, "SY": true, "CU": true,
		},
		validPurposeCodes: map[string]bool{
			"S0001": true, // Trade-related
			"S0002": true, // Investment
			"S0005": true, // Financial services
			"S0017": true, // Capital account
		},
	}
}

// RunChecks executes all IFSCA compliance checks on a wire transfer.
func (e *IFSCAComplianceEngine) RunChecks(wire WireIngress, cumulativeUSD float64) IFSCAComplianceResult {
	result := IFSCAComplianceResult{
		KYCVerified:       len(wire.InvestorPAN) == 10,
		PurposeCodeValid:  e.validPurposeCodes[wire.PurposeCode],
		SanctionsScreened: !e.sanctionedCountries[wire.SenderCountry],
		SourceOfFundsOK:   wire.AmountUSD > 0 && wire.SenderBankSWIFT != "",
	}

	// LRS limit check: cumulative + new amount must not exceed annual limit
	result.LRSLimitOK = (cumulativeUSD + wire.AmountUSD) <= e.lrsAnnualLimitUSD

	// Overall pass requires all checks to pass
	result.OverallPass = result.KYCVerified &&
		result.LRSLimitOK &&
		result.PurposeCodeValid &&
		result.SanctionsScreened &&
		result.SourceOfFundsOK

	if !result.OverallPass {
		reasons := []string{}
		if !result.KYCVerified {
			reasons = append(reasons, "KYC_FAILED")
		}
		if !result.LRSLimitOK {
			reasons = append(reasons, "LRS_LIMIT_EXCEEDED")
		}
		if !result.PurposeCodeValid {
			reasons = append(reasons, "INVALID_PURPOSE_CODE")
		}
		if !result.SanctionsScreened {
			reasons = append(reasons, "SANCTIONED_COUNTRY")
		}
		if !result.SourceOfFundsOK {
			reasons = append(reasons, "SOURCE_OF_FUNDS_FAILED")
		}
		result.RejectReason = fmt.Sprintf("FAILED: %v", reasons)
	}

	return result
}

// ──────────────────────────────────────────────────────────────────────────────
// Funding Service
// ──────────────────────────────────────────────────────────────────────────────

// GIFTCityFundingService processes USD wire ingress for GIFT City IFSC.
type GIFTCityFundingService struct {
	mu               sync.Mutex
	compliance       *IFSCAComplianceEngine
	receipts         map[string]*FundingReceipt
	cumulativeByPAN  map[string]float64    // PAN -> total USD funded in period
	walletsByPAN     map[string]string     // PAN -> wallet address
}

// NewGIFTCityFundingService creates a new funding service.
func NewGIFTCityFundingService() *GIFTCityFundingService {
	return &GIFTCityFundingService{
		compliance:      NewIFSCAComplianceEngine(),
		receipts:        make(map[string]*FundingReceipt),
		cumulativeByPAN: make(map[string]float64),
		walletsByPAN:    make(map[string]string),
	}
}

// RegisterInvestorWallet maps a PAN to a wallet address.
func (s *GIFTCityFundingService) RegisterInvestorWallet(pan, walletAddress string) error {
	if len(pan) != 10 {
		return errors.New("invalid PAN format")
	}
	if walletAddress == "" {
		return errors.New("wallet address required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.walletsByPAN[pan] = walletAddress
	return nil
}

// ProcessWire handles an incoming USD wire transfer through the GIFT City pipeline.
func (s *GIFTCityFundingService) ProcessWire(wire WireIngress) (*FundingReceipt, error) {
	if wire.WireID == "" {
		return nil, errors.New("wire_id is required")
	}
	if wire.AmountUSD <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Idempotency check
	if _, exists := s.receipts[wire.WireID]; exists {
		return nil, fmt.Errorf("duplicate wire: %s", wire.WireID)
	}

	// Lookup wallet
	wallet, hasWallet := s.walletsByPAN[wire.InvestorPAN]
	if !hasWallet {
		return nil, fmt.Errorf("no wallet registered for PAN %s", wire.InvestorPAN)
	}

	// Run compliance
	cumulative := s.cumulativeByPAN[wire.InvestorPAN]
	compliance := s.compliance.RunChecks(wire, cumulative)

	// Generate receipt ID
	payload := fmt.Sprintf("%s:%s:%.2f", wire.WireID, wire.InvestorPAN, wire.AmountUSD)
	h := sha256.Sum256([]byte(payload))
	receiptID := "GIFT-" + hex.EncodeToString(h[:12])

	status := FundingCredited
	if !compliance.OverallPass {
		status = FundingRejected
	}

	receipt := &FundingReceipt{
		ReceiptID:     receiptID,
		WireID:        wire.WireID,
		InvestorPAN:   wire.InvestorPAN,
		AmountUSD:     wire.AmountUSD,
		WalletAddress: wallet,
		Status:        status,
		Compliance:    compliance,
		CreditedAt:    time.Now().UTC(),
	}

	s.receipts[wire.WireID] = receipt

	if compliance.OverallPass {
		s.cumulativeByPAN[wire.InvestorPAN] += wire.AmountUSD
		fmt.Printf("[GIFT Funding] Credited $%.2f to wallet %s for PAN %s (Receipt: %s)\n",
			wire.AmountUSD, wallet, wire.InvestorPAN, receiptID)
	} else {
		fmt.Printf("[GIFT Funding] REJECTED wire %s: %s\n", wire.WireID, compliance.RejectReason)
	}

	return receipt, nil
}

// GetReceipt retrieves a funding receipt by wire ID.
func (s *GIFTCityFundingService) GetReceipt(wireID string) (*FundingReceipt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.receipts[wireID]
	if !ok {
		return nil, fmt.Errorf("receipt not found: %s", wireID)
	}
	return r, nil
}

// GetCumulativeFunding returns the total USD funded for a given PAN.
func (s *GIFTCityFundingService) GetCumulativeFunding(pan string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cumulativeByPAN[pan]
}
