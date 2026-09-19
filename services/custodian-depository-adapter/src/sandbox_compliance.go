package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type DeliveryInstructionSlip struct {
	DISNumber         string    `json:"dis_number"`
	Depository        string    `json:"depository"` // NSDL or CDSL
	SourceBOID        string    `json:"source_boid"`
	TargetCMBPID      string    `json:"target_cmbpid"` // Clearing Member Pool ID
	ISIN              string    `json:"isin"`
	Quantity          uint64    `json:"quantity"`
	SettlementNumber  string    `json:"settlement_number"`
	ExecutionDate     time.Time `json:"execution_date"`
	SandboxStatus     string    `json:"sandbox_status"`
}

type SandboxComplianceManager struct {
	mu                   sync.RWMutex
	maxSandboxUsers      uint32
	maxDomesticPortfolio float64
	currentUsers         uint32
	disRecords           map[string]*DeliveryInstructionSlip
}

func NewSandboxComplianceManager() *SandboxComplianceManager {
	return &SandboxComplianceManager{
		maxSandboxUsers:      10000,
		maxDomesticPortfolio: 50000.0, // ₹50,000 per user cap in sandbox phase
		currentUsers:         0,
		disRecords:           make(map[string]*DeliveryInstructionSlip),
	}
}

// CheckSandboxEligibility verifies user cap and portfolio limits under SEBI Innovation Sandbox
func (m *SandboxComplianceManager) CheckSandboxEligibility(requestedPortfolioINR float64) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.currentUsers >= m.maxSandboxUsers {
		return errors.New("SEBI sandbox participant cohort limit (10,000) reached")
	}

	if requestedPortfolioINR > m.maxDomesticPortfolio {
		return fmt.Errorf("portfolio limit exceeded: ₹%.2f > permitted sandbox cap ₹%.2f",
			requestedPortfolioINR, m.maxDomesticPortfolio)
	}

	return nil
}

// GenerateDIS creates a digital delivery instruction slip linked to physical depository
func (m *SandboxComplianceManager) GenerateDIS(disNo, dep, boid, cmbpid, isin string, qty uint64, settNo string) *DeliveryInstructionSlip {
	m.mu.Lock()
	defer m.mu.Unlock()

	dis := &DeliveryInstructionSlip{
		DISNumber:        disNo,
		Depository:       dep,
		SourceBOID:       boid,
		TargetCMBPID:     cmbpid,
		ISIN:             isin,
		Quantity:         qty,
		SettlementNumber: settNo,
		ExecutionDate:    time.Now().UTC(),
		SandboxStatus:    "SEBI_SANDBOX_VERIFIED",
	}

	m.disRecords[disNo] = dis
	return dis
}
