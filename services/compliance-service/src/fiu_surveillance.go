package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

const (
	FIUReportingThresholdINR = 1000000.0 // ₹10 Lakh statutory threshold
	StructuringMinThreshold  = 850000.0  // ₹8.5 Lakh lower bound for structuring check
	TravelRuleThresholdINR   = 50000.0   // ₹50,000 Travel Rule threshold for VDAs

	// Prompt 704 Smurfing Constants (< ₹50,000 evasion)
	SmurfingLowerBoundINR = 45000.0 // ₹45,000 lower bound for micro-smurfing
	SmurfingUpperBoundINR = 49999.0 // ₹49,999 upper bound
	SmurfingAggregateINR  = 200000.0 // ₹2 Lakh threshold within 72h window
)

// AMLTypology defines money laundering pattern classifications under PMLA
type AMLTypology string

const (
	TypologyStructuringSmurfing      AMLTypology = "STRUCTURING_SMURFING"
	TypologyRapidMovementOfFunds     AMLTypology = "RAPID_MOVEMENT_OF_FUNDS"
	TypologyDormantAccountSpike      AMLTypology = "DORMANT_ACCOUNT_SUDDEN_SPIKE"
	TypologyHighVelocitySpike        AMLTypology = "HIGH_VELOCITY_FLOW_SPIKE"
	TypologySanctionedJurisdiction   AMLTypology = "SANCTIONED_JURISDICTION_FLOW"
	TypologyMandatoryCashThreshold   AMLTypology = "MANDATORY_CTR_THRESHOLD"
)

// AlertSeverity defines urgency and risk tiers
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "LOW"
	SeverityMedium   AlertSeverity = "MEDIUM"
	SeverityHigh     AlertSeverity = "HIGH"
	SeverityCritical AlertSeverity = "CRITICAL"
)

// TransactionType defines flow categorization
type TransactionType string

const (
	TxTypeDeposit       TransactionType = "DEPOSIT"
	TxTypeWithdrawal    TransactionType = "WITHDRAWAL"
	TxTypeTradeBuy      TransactionType = "TRADE_BUY"
	TxTypeTradeSell     TransactionType = "TRADE_SELL"
	TxTypeTokenTransfer TransactionType = "TOKEN_TRANSFER"
)

// TransactionRecord captures individual transactional events for stream monitoring
type TransactionRecord struct {
	TxID                string          `json:"tx_id"`
	UserID              string          `json:"user_id"`
	AmountINR           float64         `json:"amount_inr"`
	Type                TransactionType `json:"type"`
	Timestamp           time.Time       `json:"timestamp"`
	CounterpartyCountry string          `json:"counterparty_country"`
	PaymentMode         string          `json:"payment_mode"` // UPI, NEFT, RTGS, IMPS, VDA
	DestinationAddress  string          `json:"destination_address"`
}

// UserActivityProfile tracks behavioural baselines and historical activity
type UserActivityProfile struct {
	UserID                   string    `json:"user_id"`
	LastActiveAt             time.Time `json:"last_active_at"`
	BaselineAvgDailyINR      float64   `json:"baseline_avg_daily_inr"`
	AccountCreatedAt         time.Time `json:"account_created_at"`
	RecentTransactions       []TransactionRecord `json:"recent_transactions"`
	CurrentFiatBalanceINR    float64   `json:"current_fiat_balance_inr"`
	LastDepositAmountINR     float64   `json:"last_deposit_amount_inr"`
	LastDepositTimestamp     time.Time `json:"last_deposit_timestamp"`
	RecentTradingVolumeINR   float64   `json:"recent_trading_volume_inr"`
}

// AMLAlertEvent represents an institutional real-time alert with risk scoring and freeze status
type AMLAlertEvent struct {
	AlertID                 string              `json:"alert_id"`
	TimestampUTC            time.Time           `json:"timestamp_utc"`
	InvestorID              string              `json:"investor_id"`
	Severity                AlertSeverity       `json:"severity"`
	Typology                AMLTypology         `json:"typology"`
	RiskScore               float64             `json:"risk_score"` // 0.0 to 100.0
	Description             string              `json:"description"`
	TriggeringTransactionIDs []string           `json:"triggering_transaction_ids"`
	TotalINRAmount          float64             `json:"total_inr_amount"`
	AutomatedFreezeExecuted bool                `json:"automated_freeze_executed"`
	EvidenceMerkleRoot      string              `json:"evidence_merkle_root"`
	FiledWithFIU            bool                `json:"filed_with_fiu"`
}

// FIU-IND XML Schema Structs for Statutory Reporting
type FIUBatchReport struct {
	XMLName        xml.Name          `xml:"FIUReport"`
	Version        string            `xml:"version,attr"`
	BatchHeader    FIUBatchHeader    `xml:"BatchHeader"`
	ReportDetails  FIUReportDetails  `xml:"ReportDetails"`
	AccountSubject FIUAccountSubject `xml:"AccountSubject"`
	Transactions   FIUTransactionSeq `xml:"Transactions"`
	EvidenceDigest FIUEvidenceDigest `xml:"EvidenceDigest"`
}

type FIUBatchHeader struct {
	BatchID          string `xml:"BatchID"`
	ReportingEntity  string `xml:"ReportingEntityID"`
	ReportType       string `xml:"ReportType"` // STR, CTR
	GeneratedDate    string `xml:"GeneratedDate"`
	StatutoryAct     string `xml:"StatutoryAct"`
}

type FIUReportDetails struct {
	AlertID      string  `xml:"AlertID"`
	TypologyCode string  `xml:"TypologyCode"`
	Severity     string  `xml:"Severity"`
	RiskScore    float64 `xml:"RiskScore"`
	Narrative    string  `xml:"NarrativeSummary"`
}

type FIUAccountSubject struct {
	InvestorUUID  string  `xml:"InvestorUUID"`
	DeclaredName  string  `xml:"DeclaredName"`
	PAN           string  `xml:"PAN"`
	AadhaarMasked string  `xml:"AadhaarMasked"`
	RiskTier      string  `xml:"RiskTier"`
	AccountFrozen bool    `xml:"AccountFrozen"`
}

type FIUTransactionSeq struct {
	TransactionItems []FIUTxItem `xml:"Transaction"`
}

type FIUTxItem struct {
	TxID        string  `xml:"TransactionID"`
	Timestamp   string  `xml:"Timestamp"`
	AmountINR   float64 `xml:"AmountINR"`
	Mode        string  `xml:"Mode"`
	Type        string  `xml:"Type"`
	CountryCode string  `xml:"CounterpartyCountry,omitempty"`
}

type FIUEvidenceDigest struct {
	Algorithm string `xml:"Algorithm"`
	DigestHex string `xml:"DigestHex"`
}

// FIUSurveillanceSystem manages stateful AML anomaly detection and reporting
type FIUSurveillanceSystem struct {
	mu            sync.RWMutex
	alerts        []*AMLAlert
	alertEvents   []*AMLAlertEvent
	userProfiles  map[string]*UserActivityProfile
	fatfBlacklist map[string]bool
}

func NewFIUSurveillanceSystem() *FIUSurveillanceSystem {
	return &FIUSurveillanceSystem{
		alerts:       make([]*AMLAlert, 0),
		alertEvents:  make([]*AMLAlertEvent, 0),
		userProfiles: make(map[string]*UserActivityProfile),
		fatfBlacklist: map[string]bool{
			"PRK": true,
			"IRN": true,
			"MMR": true,
		},
	}
}

// RegisterUserProfile initializes or updates baseline metrics for an investor
func (s *FIUSurveillanceSystem) RegisterUserProfile(profile UserActivityProfile) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userProfiles[profile.UserID] = &profile
}

// EvaluateStructuring checks if multiple transactions are structured to evade the ₹10L reporting threshold (backwards compatible)
func (s *FIUSurveillanceSystem) EvaluateStructuring(userID string, recentAmounts []float64) *AMLAlert {
	s.mu.Lock()
	defer s.mu.Unlock()

	var nearThresholdCount int
	var total float64
	for _, amt := range recentAmounts {
		total += amt
		if amt >= StructuringMinThreshold && amt < FIUReportingThresholdINR {
			nearThresholdCount++
		}
	}

	if nearThresholdCount >= 2 {
		alert := &AMLAlert{
			AlertID:        fmt.Sprintf("STR-%d", time.Now().UnixNano()),
			UserID:         userID,
			RuleTriggered:  "PMLA_SEC12_STRUCTURING_SUSPICION",
			Severity:       "CRITICAL",
			TransactionAmt: total,
			ReportType:     "STR",
			Description:    fmt.Sprintf("Detected %d transactions immediately below statutory ₹10L threshold (Total: ₹%.2f)", nearThresholdCount, total),
			FiledWithFIU:   false,
			CreatedAt:      time.Now().UTC(),
		}
		s.alerts = append(s.alerts, alert)
		return alert
	}

	return nil
}

// EvaluateCashThreshold checks if a single or aggregate transaction meets or exceeds the ₹10 Lakh CTR limit (backwards compatible)
func (s *FIUSurveillanceSystem) EvaluateCashThreshold(userID string, amountINR float64) *AMLAlert {
	s.mu.Lock()
	defer s.mu.Unlock()

	if amountINR >= FIUReportingThresholdINR {
		alert := &AMLAlert{
			AlertID:        fmt.Sprintf("CTR-%d", time.Now().UnixNano()),
			UserID:         userID,
			RuleTriggered:  "PMLA_SEC12_MANDATORY_CTR_THRESHOLD",
			Severity:       "HIGH",
			TransactionAmt: amountINR,
			ReportType:     "CTR",
			Description:    fmt.Sprintf("Transaction amount ₹%.2f meets or exceeds statutory ₹10 Lakh mandatory reporting threshold", amountINR),
			FiledWithFIU:   false,
			CreatedAt:      time.Now().UTC(),
		}
		s.alerts = append(s.alerts, alert)
		return alert
	}

	return nil
}

// ValidateTravelRule verifies IVMS-101 Travel Rule data for VDA transfers exceeding ₹50,000 (backwards compatible)
func (s *FIUSurveillanceSystem) ValidateTravelRule(payload TravelRulePayload) error {
	if payload.AmountFiatINR < TravelRuleThresholdINR {
		return nil
	}

	if strings.TrimSpace(payload.OriginatorName) == "" {
		return errors.New("Travel Rule violation: Originator full name is mandatory for VDA transfers exceeding ₹50,000")
	}

	if strings.TrimSpace(payload.OriginatorAccountID) == "" {
		return errors.New("Travel Rule violation: Originator account identifier is mandatory")
	}

	if strings.TrimSpace(payload.OriginatorNationalID) == "" {
		return errors.New("Travel Rule violation: Originator national ID / physical address is mandatory")
	}

	if strings.TrimSpace(payload.BeneficiaryName) == "" {
		return errors.New("Travel Rule violation: Beneficiary full name is mandatory")
	}

	if strings.TrimSpace(payload.BeneficiaryWalletAddr) == "" {
		return errors.New("Travel Rule violation: Beneficiary wallet destination address is mandatory")
	}

	if strings.TrimSpace(payload.OriginatorVASP) == "" || strings.TrimSpace(payload.BeneficiaryVASP) == "" {
		return errors.New("Travel Rule violation: Originator and Beneficiary VASP entity codes must be present")
	}

	return nil
}

// GetPendingAlerts returns unfiled STR and CTR alerts (backwards compatible)
func (s *FIUSurveillanceSystem) GetPendingAlerts() []*AMLAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	pending := make([]*AMLAlert, 0, len(s.alerts))
	for _, a := range s.alerts {
		if !a.FiledWithFIU {
			pending = append(pending, a)
		}
	}
	return pending
}

// ─────────────────────────────────────────────────────────────
// Prompt 704: Comprehensive Real-Time AML Stream Monitoring Engine
// ─────────────────────────────────────────────────────────────

// EvaluateRealtimeTransaction processes incoming transactions across all typologies
func (s *FIUSurveillanceSystem) EvaluateRealtimeTransaction(tx TransactionRecord) *AMLAlertEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	userProf, exists := s.userProfiles[tx.UserID]
	if !exists {
		userProf = &UserActivityProfile{
			UserID:               tx.UserID,
			LastActiveAt:         tx.Timestamp,
			AccountCreatedAt:     tx.Timestamp.Add(-30 * 24 * time.Hour),
			BaselineAvgDailyINR:  50000.0,
			RecentTransactions:   make([]TransactionRecord, 0),
		}
		s.userProfiles[tx.UserID] = userProf
	}

	// 1. Check Typology: Sanctioned / High-Risk FATF Jurisdiction Flow
	if tx.CounterpartyCountry != "" && s.fatfBlacklist[strings.ToUpper(tx.CounterpartyCountry)] {
		alert := s.createAlertEvent(
			tx.UserID,
			SeverityCritical,
			TypologySanctionedJurisdiction,
			98.0,
			fmt.Sprintf("Transaction flagged with FATF Call for Action blacklisted jurisdiction %s", tx.CounterpartyCountry),
			[]string{tx.TxID},
			tx.AmountINR,
			true, // Auto freeze
		)
		s.appendTransaction(userProf, tx)
		return alert
	}

	// 2. Check Typology: Dormant Account Sudden Large Spike
	// Inactive for > 180 days suddenly transacting large amount (> ₹5 Lakh)
	isDormant := !userProf.LastActiveAt.IsZero() && tx.Timestamp.Sub(userProf.LastActiveAt) > 180*24*time.Hour
	if isDormant && tx.AmountINR >= 500000.0 {
		alert := s.createAlertEvent(
			tx.UserID,
			SeverityHigh,
			TypologyDormantAccountSpike,
			88.0,
			fmt.Sprintf("Dormant account inactive for >180 days suddenly transacted ₹%.2f", tx.AmountINR),
			[]string{tx.TxID},
			tx.AmountINR,
			false,
		)
		s.appendTransaction(userProf, tx)
		return alert
	}

	// 3. Check Typology: Rapid Movement of Funds / Pass-Through Flow
	// Deposit followed by rapid withdrawal / transfer within 30 mins with minimal trading
	if (tx.Type == TxTypeWithdrawal || tx.Type == TxTypeTokenTransfer) && !userProf.LastDepositTimestamp.IsZero() {
		latency := tx.Timestamp.Sub(userProf.LastDepositTimestamp)
		if latency >= 0 && latency <= 30*time.Minute && userProf.LastDepositAmountINR > 100000.0 {
			// Check if withdrawal covers >= 80% of deposit and trading volume is < 10%
			if tx.AmountINR >= 0.80*userProf.LastDepositAmountINR && userProf.RecentTradingVolumeINR < 0.10*userProf.LastDepositAmountINR {
				alert := s.createAlertEvent(
					tx.UserID,
					SeverityCritical,
					TypologyRapidMovementOfFunds,
					92.0,
					fmt.Sprintf("Rapid movement of funds: ₹%.2f withdrawn/transferred within %.1f mins of ₹%.2f deposit without bona-fide trading", tx.AmountINR, latency.Minutes(), userProf.LastDepositAmountINR),
					[]string{tx.TxID},
					tx.AmountINR,
					true, // Risk score >= 90 triggers automated freeze
				)
				s.appendTransaction(userProf, tx)
				return alert
			}
		}
	}

	// 4. Check Typology: Smurfing / Structuring (< ₹50,000 evasion aggregating > ₹200k in 72h)
	if tx.Type == TxTypeDeposit && tx.AmountINR >= SmurfingLowerBoundINR && tx.AmountINR <= SmurfingUpperBoundINR {
		smurfingTxs := []string{tx.TxID}
		smurfingTotal := tx.AmountINR

		cutoff := tx.Timestamp.Add(-72 * time.Hour)
		for _, pastTx := range userProf.RecentTransactions {
			if pastTx.Type == TxTypeDeposit && pastTx.Timestamp.After(cutoff) &&
				pastTx.AmountINR >= SmurfingLowerBoundINR && pastTx.AmountINR <= SmurfingUpperBoundINR {
				smurfingTxs = append(smurfingTxs, pastTx.TxID)
				smurfingTotal += pastTx.AmountINR
			}
		}

		if len(smurfingTxs) >= 4 && smurfingTotal >= SmurfingAggregateINR {
			alert := s.createAlertEvent(
				tx.UserID,
				SeverityCritical,
				TypologyStructuringSmurfing,
				94.0,
				fmt.Sprintf("Smurfing pattern detected: %d deposits between ₹45k-₹49.9k aggregating to ₹%.2f within 72 hours to evade statutory ₹50k reporting", len(smurfingTxs), smurfingTotal),
				smurfingTxs,
				smurfingTotal,
				true, // >= 90 triggers freeze
			)
			s.appendTransaction(userProf, tx)
			return alert
		}
	}

	// 5. Check Typology: High-Velocity Flow Anomaly (> 300% of baseline)
	if userProf.BaselineAvgDailyINR > 0 && tx.AmountINR > 3.0*userProf.BaselineAvgDailyINR && tx.AmountINR > 500000.0 {
		ratio := (tx.AmountINR / userProf.BaselineAvgDailyINR) * 100
		alert := s.createAlertEvent(
			tx.UserID,
			SeverityMedium,
			TypologyHighVelocitySpike,
			75.0,
			fmt.Sprintf("Transaction velocity anomaly: ₹%.2f is %.1f%% of established daily baseline (₹%.2f)", tx.AmountINR, ratio, userProf.BaselineAvgDailyINR),
			[]string{tx.TxID},
			tx.AmountINR,
			false,
		)
		s.appendTransaction(userProf, tx)
		return alert
	}

	// 6. Check Typology: Mandatory CTR Threshold (₹10 Lakh)
	if tx.AmountINR >= FIUReportingThresholdINR {
		alert := s.createAlertEvent(
			tx.UserID,
			SeverityHigh,
			TypologyMandatoryCashThreshold,
			80.0,
			fmt.Sprintf("Transaction amount ₹%.2f meets statutory mandatory CTR ₹10 Lakh reporting threshold", tx.AmountINR),
			[]string{tx.TxID},
			tx.AmountINR,
			false,
		)
		s.appendTransaction(userProf, tx)
		return alert
	}

	// Normal transaction; record state
	s.appendTransaction(userProf, tx)
	return nil
}

func (s *FIUSurveillanceSystem) appendTransaction(prof *UserActivityProfile, tx TransactionRecord) {
	prof.LastActiveAt = tx.Timestamp
	if tx.Type == TxTypeDeposit {
		prof.LastDepositAmountINR = tx.AmountINR
		prof.LastDepositTimestamp = tx.Timestamp
		prof.CurrentFiatBalanceINR += tx.AmountINR
	} else if tx.Type == TxTypeWithdrawal {
		prof.CurrentFiatBalanceINR = math.Max(0, prof.CurrentFiatBalanceINR-tx.AmountINR)
	} else if tx.Type == TxTypeTradeBuy || tx.Type == TxTypeTradeSell {
		prof.RecentTradingVolumeINR += tx.AmountINR
	}
	prof.RecentTransactions = append(prof.RecentTransactions, tx)
}

func (s *FIUSurveillanceSystem) createAlertEvent(userID string, sev AlertSeverity, typ AMLTypology, score float64, desc string, txIDs []string, totalAmt float64, autoFreeze bool) *AMLAlertEvent {
	alertID := fmt.Sprintf("AML-EVT-%d", time.Now().UnixNano())

	// Compute Merkle evidence root
	evidencePayload := fmt.Sprintf("%s:%s:%s:%.2f:%s", alertID, userID, typ, totalAmt, strings.Join(txIDs, ","))
	hash := sha256.Sum256([]byte(evidencePayload))
	merkleRoot := hex.EncodeToString(hash[:])

	// Auto freeze invariant: If score >= 90.0, freeze is mandatory
	if score >= 90.0 {
		autoFreeze = true
	}

	event := &AMLAlertEvent{
		AlertID:                 alertID,
		TimestampUTC:            time.Now().UTC(),
		InvestorID:              userID,
		Severity:                sev,
		Typology:                typ,
		RiskScore:               score,
		Description:             desc,
		TriggeringTransactionIDs: txIDs,
		TotalINRAmount:          totalAmt,
		AutomatedFreezeExecuted: autoFreeze,
		EvidenceMerkleRoot:      merkleRoot,
		FiledWithFIU:            false,
	}

	s.alertEvents = append(s.alertEvents, event)
	return event
}

// GenerateFIUINDReportXML serializes an alert and evidence dossier into official FIU-IND XML schema v2.0
func (s *FIUSurveillanceSystem) GenerateFIUINDReportXML(alert *AMLAlertEvent, profile DomesticInvestorProfile, txs []TransactionRecord) (string, error) {
	reportType := "STR"
	if alert.Typology == TypologyMandatoryCashThreshold {
		reportType = "CTR"
	}

	txItems := make([]FIUTxItem, len(txs))
	for i, t := range txs {
		txItems[i] = FIUTxItem{
			TxID:        t.TxID,
			Timestamp:   t.Timestamp.UTC().Format(time.RFC3339),
			AmountINR:   t.AmountINR,
			Mode:        t.PaymentMode,
			Type:        string(t.Type),
			CountryCode: t.CounterpartyCountry,
		}
	}

	report := FIUBatchReport{
		Version: "2.0",
		BatchHeader: FIUBatchHeader{
			BatchID:         fmt.Sprintf("FIU-BATCH-%d", time.Now().UnixNano()),
			ReportingEntity: "GROWWW_NBSE_FIU_001",
			ReportType:      reportType,
			GeneratedDate:   time.Now().UTC().Format("2006-01-02T15:04:05Z"),
			StatutoryAct:    "Prevention of Money Laundering Act (PMLA) 2002 - Section 12",
		},
		ReportDetails: FIUReportDetails{
			AlertID:      alert.AlertID,
			TypologyCode: string(alert.Typology),
			Severity:     string(alert.Severity),
			RiskScore:    alert.RiskScore,
			Narrative:    alert.Description,
		},
		AccountSubject: FIUAccountSubject{
			InvestorUUID:  profile.InvestorUUID,
			DeclaredName:  profile.DeclaredName,
			PAN:           profile.PAN,
			AadhaarMasked: profile.AadhaarMasked,
			RiskTier:      string(profile.RiskTier),
			AccountFrozen: alert.AutomatedFreezeExecuted || profile.AccountFrozen,
		},
		Transactions: FIUTransactionSeq{
			TransactionItems: txItems,
		},
		EvidenceDigest: FIUEvidenceDigest{
			Algorithm: "SHA-256",
			DigestHex: alert.EvidenceMerkleRoot,
		},
	}

	out, err := xml.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal FIU report XML: %w", err)
	}

	xmlHeader := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	return xmlHeader + string(out), nil
}

// GetAlertEvents returns all recorded real-time AML alert events
func (s *FIUSurveillanceSystem) GetAlertEvents() []*AMLAlertEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copied := make([]*AMLAlertEvent, len(s.alertEvents))
	copy(copied, s.alertEvents)
	return copied
}
