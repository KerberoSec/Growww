package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// CryptoAMLDecision represents compliance pre-execution outcome
type CryptoAMLDecision string

const (
	DecisionApprove         CryptoAMLDecision = "APPROVE"
	DecisionReviewEDD       CryptoAMLDecision = "REVIEW_EDD"
	DecisionRejectAndFreeze CryptoAMLDecision = "REJECT_AND_FREEZE"
)

// ExposureCategory represents blockchain transaction entity classifications
type ExposureCategory string

const (
	ExposureSanctionsOFAC     ExposureCategory = "OFAC_SANCTIONS"
	ExposureDarknetMarkets    ExposureCategory = "DARKNET_MARKETS"
	ExposureRansomware        ExposureCategory = "RANSOMWARE"
	ExposureMixersTumblers    ExposureCategory = "MIXERS_TUMBLERS"
	ExposureStolenFundsHack   ExposureCategory = "STOLEN_FUNDS_HACK"
	ExposureTerroristFinance  ExposureCategory = "TERRORIST_FINANCING"
	ExposureScamFraud         ExposureCategory = "SCAM_FRAUD"
	ExposureHighRiskDEX       ExposureCategory = "HIGH_RISK_DEX_BRIDGE"
	ExposureLicensedVASP      ExposureCategory = "LICENSED_EXCHANGE_VASP"
	ExposureCleanP2P          ExposureCategory = "CLEAN_P2P"
)

// Base category risk weights (0 to 100)
var categoryWeights = map[ExposureCategory]float64{
	ExposureSanctionsOFAC:    100.0,
	ExposureTerroristFinance: 100.0,
	ExposureRansomware:       95.0,
	ExposureStolenFundsHack:  90.0,
	ExposureDarknetMarkets:   85.0,
	ExposureMixersTumblers:   75.0,
	ExposureScamFraud:        70.0,
	ExposureHighRiskDEX:      40.0,
	ExposureLicensedVASP:     5.0,
	ExposureCleanP2P:         10.0,
}

// ExposureDetail represents single category exposure percentage and hop distance
type ExposureDetail struct {
	Category    ExposureCategory `json:"category"`
	Percentage  float64          `json:"percentage"`  // 0.0 to 1.0 (e.g. 0.25 = 25%)
	HopDistance int              `json:"hop_distance"`// 1 = direct, 2 = 1 intermediary, 3 = 2 intermediaries
	EntityName  string           `json:"entity_name"`
}

// CryptoAMLScreeningRequest represents an incoming address or transaction screening
type CryptoAMLScreeningRequest struct {
	RequestID       string           `json:"request_id"`
	UserID          string           `json:"user_id"`
	WalletAddress   string           `json:"wallet_address"`
	Blockchain      string           `json:"blockchain"` // "ETHEREUM", "BESU", "BITCOIN", "POLYGON"
	TransactionHash string           `json:"transaction_hash,omitempty"`
	AmountUSD       float64          `json:"amount_usd"`
	Exposures       []ExposureDetail `json:"exposures"`
	EllipticScore   float64          `json:"elliptic_score"`   // 0.0 to 10.0 (provider score)
	ChainalysisRisk string           `json:"chainalysis_risk"` // "Severe", "High", "Medium", "Low"
}

// CryptoAMLScreeningResult represents the authoritative AML decision
type CryptoAMLScreeningResult struct {
	ResultID             string            `json:"result_id"`
	RequestID            string            `json:"request_id"`
	UserID               string            `json:"user_id"`
	WalletAddress        string            `json:"wallet_address"`
	CompositeRiskScore   float64           `json:"composite_risk_score"` // 0 to 100
	Decision             CryptoAMLDecision `json:"decision"`
	HasDirectSanctions   bool              `json:"has_direct_sanctions"`
	HasTerroristFinance  bool              `json:"has_terrorist_finance"`
	STRTriggered         bool              `json:"str_triggered"`
	STRReportID          string            `json:"str_report_id,omitempty"`
	DetailedBreakdown    map[string]float64`json:"detailed_breakdown"`
	BesuComplianceTxHash string            `json:"besu_compliance_tx_hash"`
	ScreenedAt           time.Time         `json:"screened_at"`
}

// FIUSTRFiling represents a Suspicious Transaction Report filed to FIU-IND
type FIUSTRFiling struct {
	STRID          string    `json:"str_id"`
	UserID         string    `json:"user_id"`
	WalletAddress  string    `json:"wallet_address"`
	ReasonCode     string    `json:"reason_code"`
	RiskScore      float64   `json:"risk_score"`
	FilingSummary  string    `json:"filing_summary"`
	FiledAt        time.Time `json:"filed_at"`
}

// CryptoAMLScoringEngine orchestrates Elliptic & Chainalysis risk aggregation and transfer compliance
type CryptoAMLScoringEngine struct {
	mu           sync.RWMutex
	sanctioned   map[string]bool // address -> sanctioned
	results      map[string]*CryptoAMLScreeningResult
	strFilings   map[string]*FIUSTRFiling
	decayFactor  float64 // gamma = 0.5 per hop
}

// NewCryptoAMLScoringEngine creates a new CryptoAMLScoringEngine
func NewCryptoAMLScoringEngine() *CryptoAMLScoringEngine {
	return &CryptoAMLScoringEngine{
		sanctioned:  make(map[string]bool),
		results:     make(map[string]*CryptoAMLScreeningResult),
		strFilings:  make(map[string]*FIUSTRFiling),
		decayFactor: 0.5,
	}
}

// AddSanctionedAddress adds an OFAC / UN / MHA sanctioned crypto address
func (e *CryptoAMLScoringEngine) AddSanctionedAddress(address string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sanctioned[address] = true
}

// ScreenCryptoTransaction evaluates multi-hop exposures and produces composite AML score & decision
func (e *CryptoAMLScoringEngine) ScreenCryptoTransaction(req *CryptoAMLScreeningRequest) (*CryptoAMLScreeningResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if req == nil {
		return nil, errors.New("request cannot be nil")
	}
	if req.WalletAddress == "" {
		return nil, errors.New("wallet_address is required")
	}

	// 1. Check known blacklisted/sanctioned address database
	if e.sanctioned[req.WalletAddress] {
		strID := fmt.Sprintf("str-ofac-%s", req.WalletAddress[:8])
		str := &FIUSTRFiling{
			STRID:         strID,
			UserID:        req.UserID,
			WalletAddress: req.WalletAddress,
			ReasonCode:    "OFAC_SANCTION_LIST_DIRECT_HIT",
			RiskScore:     100.0,
			FilingSummary: fmt.Sprintf("Address %s is an exact match on statutory OFAC / FIU-IND sanctions list", req.WalletAddress),
			FiledAt:       time.Now().UTC(),
		}
		e.strFilings[strID] = str

		resultID := fmt.Sprintf("aml-res-%s", req.RequestID)
		res := &CryptoAMLScreeningResult{
			ResultID:             resultID,
			RequestID:            req.RequestID,
			UserID:               req.UserID,
			WalletAddress:        req.WalletAddress,
			CompositeRiskScore:   100.0,
			Decision:             DecisionRejectAndFreeze,
			HasDirectSanctions:   true,
			STRTriggered:         true,
			STRReportID:          strID,
			DetailedBreakdown:    map[string]float64{"OFAC_SANCTIONS": 100.0},
			BesuComplianceTxHash: "0x" + hex.EncodeToString(sha256.New().Sum([]byte(strID))),
			ScreenedAt:           time.Now().UTC(),
		}
		e.results[resultID] = res
		return res, nil
	}

	// 2. Mathematical Multi-Hop Exposure Risk Calculation:
	// TotalRisk = min(100, sum( w_i * DirectPct_i ) + sum( gamma^(hop-1) * w_j * IndirectPct_j ))
	totalRisk := 0.0
	breakdown := make(map[string]float64)
	hasDirectSanctions := false
	hasTerroristFinance := false

	for _, exp := range req.Exposures {
		weight, ok := categoryWeights[exp.Category]
		if !ok {
			weight = 20.0
		}

		hop := exp.HopDistance
		if hop <= 1 {
			hop = 1
			if exp.Category == ExposureSanctionsOFAC {
				hasDirectSanctions = true
			}
			if exp.Category == ExposureTerroristFinance {
				hasTerroristFinance = true
			}
		}

		// Geometric decay: gamma^(hop-1)
		decay := math.Pow(e.decayFactor, float64(hop-1))
		contribution := weight * exp.Percentage * decay
		totalRisk += contribution
		breakdown[string(exp.Category)] += contribution
	}

	// Incorporate vendor scores (Elliptic 0-10 normalized to 0-100; Chainalysis)
	if req.EllipticScore > 0 {
		normalizedElliptic := req.EllipticScore * 10.0
		totalRisk = 0.7*totalRisk + 0.3*normalizedElliptic
	}
	if req.ChainalysisRisk == "Severe" && totalRisk < 80.0 {
		totalRisk = 85.0
	} else if req.ChainalysisRisk == "High" && totalRisk < 60.0 {
		totalRisk = 65.0
	}

	if totalRisk > 100.0 {
		totalRisk = 100.0
	}

	// 3. Decision Logic & FIU STR Trigger
	var decision CryptoAMLDecision
	strTriggered := false
	var strID string

	if hasDirectSanctions || hasTerroristFinance || totalRisk >= 70.0 {
		decision = DecisionRejectAndFreeze
		strTriggered = true
		strID = fmt.Sprintf("str-aml-%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%f", req.WalletAddress, totalRisk))))[:16]
		e.strFilings[strID] = &FIUSTRFiling{
			STRID:         strID,
			UserID:        req.UserID,
			WalletAddress: req.WalletAddress,
			ReasonCode:    "HIGH_AML_RISK_SUSPICIOUS_TRANSFER",
			RiskScore:     totalRisk,
			FilingSummary: fmt.Sprintf("Transfer rejected: Composite crypto AML risk score %.2f exceeds critical threshold", totalRisk),
			FiledAt:       time.Now().UTC(),
		}
	} else if totalRisk >= 30.0 {
		decision = DecisionReviewEDD
	} else {
		decision = DecisionApprove
	}

	// 4. Besu On-Chain Transfer Hook Digest
	hookDigest := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%.2f:%d",
		req.WalletAddress, decision, req.Blockchain, totalRisk, time.Now().Unix())))
	besuTx := "0x" + hex.EncodeToString(hookDigest[:])

	resultID := fmt.Sprintf("aml-res-%s", req.RequestID)
	result := &CryptoAMLScreeningResult{
		ResultID:             resultID,
		RequestID:            req.RequestID,
		UserID:               req.UserID,
		WalletAddress:        req.WalletAddress,
		CompositeRiskScore:   totalRisk,
		Decision:             decision,
		HasDirectSanctions:   hasDirectSanctions,
		HasTerroristFinance:  hasTerroristFinance,
		STRTriggered:         strTriggered,
		STRReportID:          strID,
		DetailedBreakdown:    breakdown,
		BesuComplianceTxHash: besuTx,
		ScreenedAt:           time.Now().UTC(),
	}

	e.results[resultID] = result
	return result, nil
}

// VerifyTransferCompliance is the smart contract hook validator for Besu token transfers
func (e *CryptoAMLScoringEngine) VerifyTransferCompliance(fromAddr, toAddr string, amountUSD float64) (bool, string) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.sanctioned[fromAddr] {
		return false, "Sender address is subject to statutory OFAC / AML asset freeze"
	}
	if e.sanctioned[toAddr] {
		return false, "Recipient address is subject to statutory OFAC / AML asset freeze"
	}
	return true, "Transfer compliant with on-chain AML policy"
}
