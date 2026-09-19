package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type TierLimitsManager struct {
	mu           sync.RWMutex
	tierLimits   map[KYCTier]TierLimits
	reKYCPeriod  map[RiskTier]time.Duration
}

func NewTierLimitsManager() *TierLimitsManager {
	return &TierLimitsManager{
		tierLimits: map[KYCTier]TierLimits{
			KYCTier0Unverified: {
				Tier:       KYCTier0Unverified,
				DailyINR:   0.0,
				AnnualINR:  0.0,
				StatusDesc: "UNVERIFIED: Zero trading and withdrawal allowed",
			},
			KYCTier1BasicOTP: {
				Tier:       KYCTier1BasicOTP,
				DailyINR:   25000.0,   // ₹25,000 per day
				AnnualINR:  100000.0,  // ₹1,00,000 per year
				StatusDesc: "BASIC_OTP: Instant Aadhaar OTP & PAN verified",
			},
			KYCTier2FullCKYC: {
				Tier:       KYCTier2FullCKYC,
				DailyINR:   500000.0,  // ₹5,00,000 per day
				AnnualINR:  2500000.0, // ₹25,00,000 per year
				StatusDesc: "FULL_CKYC: 14-digit CKYC & Bank Penny Drop verified",
			},
			KYCTier3VCIP: {
				Tier:       KYCTier3VCIP,
				DailyINR:   50000000.0,  // ₹5,00,00,000 (₹5 Cr) per day
				AnnualINR:  500000000.0, // ₹50,00,00,000 (₹50 Cr) per year
				StatusDesc: "INSTITUTIONAL_VCIP: Video CIP & CA Net Worth verified",
			},
		},
		reKYCPeriod: map[RiskTier]time.Duration{
			RiskTierLow:    10 * 365 * 24 * time.Hour, // 10 years
			RiskTierMedium: 8 * 365 * 24 * time.Hour,  // 8 years
			RiskTierHigh:   2 * 365 * 24 * time.Hour,  // 2 years
		},
	}
}

// GetTierLimits returns the statutory limits for a tier
func (m *TierLimitsManager) GetTierLimits(tier KYCTier) TierLimits {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tierLimits[tier]
}

// CalculateReKYCDue computes statutory re-KYC expiration date based on AML risk tier
func (m *TierLimitsManager) CalculateReKYCDue(riskTier RiskTier, verifiedAt time.Time) time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	duration, exists := m.reKYCPeriod[riskTier]
	if !exists {
		duration = 2 * 365 * 24 * time.Hour // default to 2 years for prudence
	}
	return verifiedAt.Add(duration)
}

// CheckLimitEnforcement validates if a requested withdrawal violates tier daily or annual quotas
func (m *TierLimitsManager) CheckLimitEnforcement(tier KYCTier, requestedINR, dailySpentINR, annualSpentINR float64) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limits, exists := m.tierLimits[tier]
	if !exists || tier == KYCTier0Unverified {
		return errors.New("tier 0 (unverified) users cannot withdraw or trade; complete KYC to unlock limits")
	}

	if requestedINR <= 0 {
		return errors.New("requested transaction amount must be strictly positive")
	}

	projectedDaily := dailySpentINR + requestedINR
	if projectedDaily > limits.DailyINR {
		return fmt.Errorf("daily withdrawal limit exceeded: ₹%.2f + ₹%.2f > permitted daily quota ₹%.2f for %s",
			dailySpentINR, requestedINR, limits.DailyINR, limits.StatusDesc)
	}

	projectedAnnual := annualSpentINR + requestedINR
	if projectedAnnual > limits.AnnualINR {
		return fmt.Errorf("annual withdrawal limit exceeded: ₹%.2f + ₹%.2f > permitted annual quota ₹%.2f for %s",
			annualSpentINR, requestedINR, limits.AnnualINR, limits.StatusDesc)
	}

	return nil
}
