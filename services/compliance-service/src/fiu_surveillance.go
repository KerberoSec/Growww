package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	FIUReportingThresholdINR = 1000000.0 // ₹10 Lakh statutory threshold
	StructuringMinThreshold  = 850000.0  // ₹8.5 Lakh lower bound for structuring check
	TravelRuleThresholdINR   = 50000.0   // ₹50,000 Travel Rule threshold for VDAs
)

type FIUSurveillanceSystem struct {
	mu     sync.RWMutex
	alerts []*AMLAlert
}

func NewFIUSurveillanceSystem() *FIUSurveillanceSystem {
	return &FIUSurveillanceSystem{
		alerts: make([]*AMLAlert, 0),
	}
}

// EvaluateStructuring checks if multiple transactions are structured to evade the ₹10L reporting threshold
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

// EvaluateCashThreshold checks if a single or aggregate transaction meets or exceeds the ₹10 Lakh CTR limit
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

// ValidateTravelRule verifies IVMS-101 Travel Rule data for VDA transfers exceeding ₹50,000
func (s *FIUSurveillanceSystem) ValidateTravelRule(payload TravelRulePayload) error {
	if payload.AmountFiatINR < TravelRuleThresholdINR {
		// Below statutory Travel Rule threshold; basic checks only
		return nil
	}

	// For transfers >= ₹50,000, enforce strict IVMS-101 schema
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

// GetPendingAlerts returns unfiled STR and CTR alerts
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
