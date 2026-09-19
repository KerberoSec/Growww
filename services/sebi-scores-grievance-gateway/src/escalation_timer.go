package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// EscalationTimer tracks 21-day statutory SLAs and generates alerts across milestones
type EscalationTimer struct {
	mu            sync.RWMutex
	gateway       *SCORESGatewayConnector
	alertsHistory []EscalationAlert
}

// NewEscalationTimer creates a new escalation timer attached to the SCORES gateway
func NewEscalationTimer(gateway *SCORESGatewayConnector) *EscalationTimer {
	return &EscalationTimer{
		gateway:       gateway,
		alertsHistory: make([]EscalationAlert, 0),
	}
}

// CalculateEscalationStage evaluates the current milestone for a docket given an evaluation timestamp
func (t *EscalationTimer) CalculateEscalationStage(docket *ComplaintDocket, evalTime time.Time) (EscalationStage, int, int) {
	if evalTime.IsZero() {
		evalTime = time.Now().UTC()
	}

	durationElapsed := evalTime.Sub(docket.IntakeTimestamp)
	daysElapsed := int(math.Floor(durationElapsed.Hours() / 24.0))
	if daysElapsed < 0 {
		daysElapsed = 0
	}

	totalDays := docket.StatutorySLADays
	if totalDays <= 0 {
		totalDays = 21 // Default SEBI SCORES 2.0 statutory SLA
	}

	daysRemaining := totalDays - daysElapsed
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	var stage EscalationStage
	if daysElapsed >= totalDays {
		stage = Stage5Day21Breached
	} else if daysElapsed >= 20 {
		stage = Stage4Day20Freeze
	} else if daysElapsed >= 18 {
		stage = Stage3Day18CriticalP1
	} else if daysElapsed >= 14 {
		stage = Stage2Day14Escalation
	} else if daysElapsed >= 7 {
		stage = Stage1Day7Warning
	} else {
		stage = Stage0Normal
	}

	return stage, daysElapsed, daysRemaining
}

// EvaluateDocketSLA evaluates and updates a single docket's stage and breach status
func (t *EscalationTimer) EvaluateDocketSLA(docket *ComplaintDocket, evalTime time.Time) *EscalationAlert {
	// If already closed, no escalation needed
	if docket.CurrentState == StateResolvedClosed || docket.CurrentState == StateRejectedInvalid {
		return nil
	}

	newStage, daysElapsed, daysRemaining := t.CalculateEscalationStage(docket, evalTime)
	prevStage := docket.CurrentEscalationStage

	docket.CurrentEscalationStage = newStage
	if newStage == Stage5Day21Breached {
		docket.IsSLABreached = true
	}

	// If stage has escalated, emit an alert
	if newStage > prevStage {
		severity := "INFO"
		msg := fmt.Sprintf("Docket %s has reached escalation stage %d", docket.ExternalRegistrationNumber, newStage)

		switch newStage {
		case Stage1Day7Warning:
			severity = "WARNING"
			msg = fmt.Sprintf("Day 7 Warning: Docket %s has 14 days remaining. Assignee must finalize evidence dossier.", docket.ExternalRegistrationNumber)
		case Stage2Day14Escalation:
			severity = "HIGH"
			msg = fmt.Sprintf("Day 14 Escalation: Docket %s escalated to Principal Compliance Officer (PCO). 7 days remaining.", docket.ExternalRegistrationNumber)
		case Stage3Day18CriticalP1:
			severity = "CRITICAL_P1"
			msg = fmt.Sprintf("RUNBOOK-34 CRITICAL P1 ALERT: Docket %s has only 3 days left. Senior Legal Lead war room convened.", docket.ExternalRegistrationNumber)
		case Stage4Day20Freeze:
			severity = "CRITICAL_P1"
			msg = fmt.Sprintf("Day 20 Mandatory Sign-Off Freeze: Docket %s has less than 24 hours. Immediate Checker sign-off required.", docket.ExternalRegistrationNumber)
		case Stage5Day21Breached:
			severity = "SLA_BREACH"
			msg = fmt.Sprintf("STATUTORY SLA BREACH: Docket %s has exceeded 21 days! Automatic escalation to SEBI Enforcement Panel.", docket.ExternalRegistrationNumber)
		}

		alert := EscalationAlert{
			ComplaintID:                docket.ID,
			ExternalRegistrationNumber: docket.ExternalRegistrationNumber,
			Stage:                      newStage,
			DaysElapsed:                daysElapsed,
			DaysRemaining:              daysRemaining,
			AlertSeverity:              severity,
			Message:                    msg,
			TriggeredAt:                evalTime,
		}

		t.mu.Lock()
		t.alertsHistory = append(t.alertsHistory, alert)
		t.mu.Unlock()

		return &alert
	}

	return nil
}

// EvaluateAll evaluates all active dockets in the gateway and returns generated alerts
func (t *EscalationTimer) EvaluateAll(evalTime time.Time) []EscalationAlert {
	dockets := t.gateway.ListDockets()
	alerts := make([]EscalationAlert, 0)

	for _, d := range dockets {
		if alert := t.EvaluateDocketSLA(d, evalTime); alert != nil {
			alerts = append(alerts, *alert)
		}
	}

	return alerts
}

// GetAlertsHistory returns the complete audit trail of escalation alerts
func (t *EscalationTimer) GetAlertsHistory() []EscalationAlert {
	t.mu.RLock()
	defer t.mu.RUnlock()
	copied := make([]EscalationAlert, len(t.alertsHistory))
	copy(copied, t.alertsHistory)
	return copied
}

// GetSLADashboard computes overall health metrics for all dockets
func (t *EscalationTimer) GetSLADashboard(evalTime time.Time) *SLADashboard {
	if evalTime.IsZero() {
		evalTime = time.Now().UTC()
	}

	dockets := t.gateway.ListDockets()
	dashboard := &SLADashboard{
		DocketsByStage: make(map[int]int),
		CalculatedAt:   evalTime,
	}

	for _, d := range dockets {
		if d.CurrentState == StateResolvedClosed || d.CurrentState == StateRejectedInvalid {
			dashboard.ResolvedCount++
			continue
		}

		dashboard.TotalActiveDockets++
		stage, _, _ := t.CalculateEscalationStage(d, evalTime)
		dashboard.DocketsByStage[int(stage)]++

		if stage >= Stage3Day18CriticalP1 {
			dashboard.CriticalBreachRiskCount++
		}
		if stage == Stage5Day21Breached || d.IsSLABreached {
			dashboard.BreachedCount++
		}
	}

	return dashboard
}
