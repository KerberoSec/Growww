package main

import (
	"fmt"
	"sync"
	"time"
)

type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "LOW"
	SeverityMedium   AlertSeverity = "MEDIUM"
	SeverityHigh     AlertSeverity = "HIGH"
	SeverityCritical AlertSeverity = "CRITICAL"
)

type AMLSuspiciousAlert struct {
	AlertID        string        `json:"alert_id"`
	UserID         string        `json:"user_id"`
	RuleTriggered  string        `json:"rule_triggered"`
	Severity       AlertSeverity `json:"severity"`
	TransactionAmt float64       `json:"transaction_amt"`
	Description    string        `json:"description"`
	FiledWithFIU   bool          `json:"filed_with_fiu"`
	CreatedAt      time.Time     `json:"created_at"`
}

type TravelRulePayload struct {
	TransferID             string `json:"transfer_id"`
	OriginatorName         string `json:"originator_name"`
	OriginatorAccount      string `json:"originator_account"`
	OriginatorPhysicalAddr string `json:"originator_physical_addr"`
	BeneficiaryName        string `json:"beneficiary_name"`
	BeneficiaryAccount     string `json:"beneficiary_account"`
	VASPCodeOriginator     string `json:"vasp_code_originator"`
	VASPCodeBeneficiary    string `json:"vasp_code_beneficiary"`
}

type FIUSurveillanceEngine struct {
	mu     sync.Mutex
	alerts []*AMLSuspiciousAlert
}

func NewFIUSurveillanceEngine() *FIUSurveillanceEngine {
	return &FIUSurveillanceEngine{
		alerts: make([]*AMLSuspiciousAlert, 0),
	}
}

// EvaluateStructuring checks if multiple transactions are structured just below statutory reporting threshold (₹10 Lakh)
func (e *FIUSurveillanceEngine) EvaluateStructuring(userID string, txAmounts []float64) *AMLSuspiciousAlert {
	e.mu.Lock()
	defer e.mu.Unlock()

	var countNearThreshold int
	var total float64
	for _, amt := range txAmounts {
		total += amt
		if amt >= 850000.0 && amt < 1000000.0 {
			countNearThreshold++
		}
	}

	if countNearThreshold >= 2 {
		alert := &AMLSuspiciousAlert{
			AlertID:        fmt.Sprintf("STR-%d", time.Now().UnixNano()),
			UserID:         userID,
			RuleTriggered:  "RULE_STRUCTURING_PMLA_SEC12",
			Severity:       SeverityCritical,
			TransactionAmt: total,
			Description:    fmt.Sprintf("Detected %d transactions just below ₹10L reporting threshold", countNearThreshold),
			FiledWithFIU:   false,
			CreatedAt:      time.Now().UTC(),
		}
		e.alerts = append(e.alerts, alert)
		fmt.Printf("[FIU Surveillance] ALERT: Potential structuring flagged for user %s\n", userID)
		return alert
	}

	return nil
}
