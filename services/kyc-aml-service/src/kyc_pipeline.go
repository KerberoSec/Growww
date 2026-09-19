// Package main enhances the NBSE KYC/AML service with full KYC pipeline orchestration,
// risk scoring engine, PEP/sanctions screening, and re-KYC scheduling.
package main

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// KYCStatus represents the lifecycle state of a KYC application.
type KYCStatus string

const (
	KYCStatusPending    KYCStatus = "PENDING"
	KYCStatusInProgress KYCStatus = "IN_PROGRESS"
	KYCStatusApproved   KYCStatus = "APPROVED"
	KYCStatusRejected   KYCStatus = "REJECTED"
	KYCStatusExpired    KYCStatus = "EXPIRED"
	KYCStatusReKYC      KYCStatus = "RE_KYC_REQUIRED"
)

// KYCTier defines the KYC verification level.
type KYCTier string

const (
	TierBasic    KYCTier = "BASIC"    // Aadhaar + PAN
	TierStandard KYCTier = "STANDARD" // + Bank verification + Selfie
	TierAdvanced KYCTier = "ADVANCED" // + Income proof + CKYC
)

// RiskLevel represents the AML risk level for a customer.
type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

// KYCApplication represents a customer's KYC submission.
type KYCApplication struct {
	ApplicationID string    `json:"application_id"`
	UserID        string    `json:"user_id"`
	FullName      string    `json:"full_name"`
	PAN           string    `json:"pan"`
	AadhaarHash   string    `json:"aadhaar_hash"` // Hashed Aadhaar for privacy
	DOB           string    `json:"dob"`
	Tier          KYCTier   `json:"tier"`
	Status        KYCStatus `json:"status"`
	RiskScore     int       `json:"risk_score"` // 0-100
	RiskLevel     RiskLevel `json:"risk_level"`
	IsPEP         bool      `json:"is_pep"`
	IsSanctioned  bool      `json:"is_sanctioned"`
	ReKYCDueDate  time.Time `json:"re_kyc_due_date"`
	CreatedAt     time.Time `json:"created_at"`
	ApprovedAt    time.Time `json:"approved_at"`
}

// ScreeningResult holds PEP/sanctions screening outcome.
type ScreeningResult struct {
	UserID       string `json:"user_id"`
	IsPEP        bool   `json:"is_pep"`
	IsSanctioned bool   `json:"is_sanctioned"`
	MatchedList  string `json:"matched_list"`
	Confidence   float64 `json:"confidence"`
}

// KYCPipeline manages the full KYC lifecycle with screening and risk scoring.
type KYCPipeline struct {
	mu              sync.RWMutex
	applications    map[string]*KYCApplication // appID -> application
	userApps        map[string]string          // userID -> latest appID
	sanctionsList   map[string]bool            // normalized names on sanctions list
	pepList         map[string]bool            // normalized names of PEPs
	reKYCIntervalDays int
}

// NewKYCPipeline creates a new KYCPipeline.
func NewKYCPipeline(reKYCIntervalDays int) *KYCPipeline {
	if reKYCIntervalDays <= 0 {
		reKYCIntervalDays = 365 // Default: annual re-KYC
	}
	return &KYCPipeline{
		applications:    make(map[string]*KYCApplication),
		userApps:        make(map[string]string),
		sanctionsList:   make(map[string]bool),
		pepList:         make(map[string]bool),
		reKYCIntervalDays: reKYCIntervalDays,
	}
}

// LoadSanctionsList loads sanctioned entity names for screening.
func (p *KYCPipeline) LoadSanctionsList(names []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, name := range names {
		p.sanctionsList[strings.ToUpper(strings.TrimSpace(name))] = true
	}
}

// LoadPEPList loads politically exposed persons for screening.
func (p *KYCPipeline) LoadPEPList(names []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, name := range names {
		p.pepList[strings.ToUpper(strings.TrimSpace(name))] = true
	}
}

// SubmitApplication creates a new KYC application.
func (p *KYCPipeline) SubmitApplication(appID, userID, fullName, pan, aadhaarHash, dob string, tier KYCTier) (*KYCApplication, error) {
	if appID == "" || userID == "" {
		return nil, errors.New("application ID and user ID are required")
	}
	if fullName == "" {
		return nil, errors.New("full name is required")
	}
	if len(pan) != 10 {
		return nil, errors.New("valid 10-character PAN is required")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.applications[appID]; exists {
		return nil, fmt.Errorf("application %s already exists", appID)
	}

	app := &KYCApplication{
		ApplicationID: appID,
		UserID:        userID,
		FullName:      fullName,
		PAN:           pan,
		AadhaarHash:   aadhaarHash,
		DOB:           dob,
		Tier:          tier,
		Status:        KYCStatusPending,
		RiskScore:     0,
		RiskLevel:     RiskLow,
		CreatedAt:     time.Now().UTC(),
	}

	p.applications[appID] = app
	p.userApps[userID] = appID
	return app, nil
}

// ScreenApplicant runs PEP and sanctions screening against loaded lists.
func (p *KYCPipeline) ScreenApplicant(appID string) (*ScreeningResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	app, exists := p.applications[appID]
	if !exists {
		return nil, fmt.Errorf("application %s not found", appID)
	}

	normalizedName := strings.ToUpper(strings.TrimSpace(app.FullName))

	result := &ScreeningResult{
		UserID:     app.UserID,
		Confidence: 1.0,
	}

	if p.sanctionsList[normalizedName] {
		result.IsSanctioned = true
		result.MatchedList = "OFAC_SDN"
		app.IsSanctioned = true
	}

	if p.pepList[normalizedName] {
		result.IsPEP = true
		result.MatchedList = "INDIA_PEP_REGISTRY"
		app.IsPEP = true
	}

	return result, nil
}

// CalculateRiskScore computes a risk score (0-100) based on multiple factors.
func (p *KYCPipeline) CalculateRiskScore(appID string, monthlyTxVolume float64, countryRisk int, isNRI bool) (int, RiskLevel, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	app, exists := p.applications[appID]
	if !exists {
		return 0, "", fmt.Errorf("application %s not found", appID)
	}

	score := 0

	// PEP adds 30 points
	if app.IsPEP {
		score += 30
	}

	// Sanctioned = automatic max
	if app.IsSanctioned {
		score = 100
		app.RiskScore = score
		app.RiskLevel = RiskCritical
		return score, RiskCritical, nil
	}

	// High monthly volume (> ₹50L) adds 20
	if monthlyTxVolume > 5000000 {
		score += 20
	} else if monthlyTxVolume > 1000000 {
		score += 10
	}

	// Country risk factor (0-40)
	if countryRisk > 40 {
		countryRisk = 40
	}
	score += countryRisk

	// NRI adds 15 (higher EDD requirements)
	if isNRI {
		score += 15
	}

	if score > 100 {
		score = 100
	}

	var level RiskLevel
	switch {
	case score >= 75:
		level = RiskCritical
	case score >= 50:
		level = RiskHigh
	case score >= 25:
		level = RiskMedium
	default:
		level = RiskLow
	}

	app.RiskScore = score
	app.RiskLevel = level
	return score, level, nil
}

// ApproveApplication marks the KYC as approved and schedules re-KYC.
func (p *KYCPipeline) ApproveApplication(appID string) (*KYCApplication, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	app, exists := p.applications[appID]
	if !exists {
		return nil, fmt.Errorf("application %s not found", appID)
	}

	if app.IsSanctioned {
		return nil, errors.New("cannot approve sanctioned applicant")
	}

	app.Status = KYCStatusApproved
	app.ApprovedAt = time.Now().UTC()
	app.ReKYCDueDate = time.Now().UTC().AddDate(0, 0, p.reKYCIntervalDays)

	return app, nil
}

// RejectApplication marks the KYC as rejected.
func (p *KYCPipeline) RejectApplication(appID, reason string) (*KYCApplication, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	app, exists := p.applications[appID]
	if !exists {
		return nil, fmt.Errorf("application %s not found", appID)
	}

	app.Status = KYCStatusRejected
	return app, nil
}

// CheckReKYCDue returns applications that need re-KYC.
func (p *KYCPipeline) CheckReKYCDue() []*KYCApplication {
	p.mu.RLock()
	defer p.mu.RUnlock()

	now := time.Now().UTC()
	var due []*KYCApplication
	for _, app := range p.applications {
		if app.Status == KYCStatusApproved && !app.ReKYCDueDate.IsZero() && now.After(app.ReKYCDueDate) {
			due = append(due, app)
		}
	}
	return due
}

// GetApplication retrieves a KYC application by ID.
func (p *KYCPipeline) GetApplication(appID string) (*KYCApplication, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	app, exists := p.applications[appID]
	if !exists {
		return nil, fmt.Errorf("application %s not found", appID)
	}
	return app, nil
}
