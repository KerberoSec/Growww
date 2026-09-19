// Package affiliaterebate implements a tiered referral rebate engine for NBSE.
// It supports Bronze/Silver/Gold/Platinum tiers with monthly payout cycles
// and anti-gaming fraud detection heuristics.
package affiliaterebate

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// Tier represents the affiliate program tier level.
type Tier string

const (
	TierBronze   Tier = "BRONZE"
	TierSilver   Tier = "SILVER"
	TierGold     Tier = "GOLD"
	TierPlatinum Tier = "PLATINUM"
)

// TierConfig defines rebate percentages and thresholds for each tier.
type TierConfig struct {
	Tier               Tier    `json:"tier"`
	RebatePercent      float64 `json:"rebate_percent"`       // % of referred user's trading fees
	MinMonthlyVolumeINR float64 `json:"min_monthly_volume_inr"` // Min volume to maintain tier
	MinReferrals       int     `json:"min_referrals"`         // Min active referrals needed
}

// DefaultTierConfigs returns the standard NBSE affiliate tier configuration.
func DefaultTierConfigs() []TierConfig {
	return []TierConfig{
		{Tier: TierBronze, RebatePercent: 10.0, MinMonthlyVolumeINR: 0, MinReferrals: 0},
		{Tier: TierSilver, RebatePercent: 15.0, MinMonthlyVolumeINR: 500_000, MinReferrals: 5},
		{Tier: TierGold, RebatePercent: 20.0, MinMonthlyVolumeINR: 5_000_000, MinReferrals: 20},
		{Tier: TierPlatinum, RebatePercent: 30.0, MinMonthlyVolumeINR: 50_000_000, MinReferrals: 100},
	}
}

// Affiliate represents a referrer in the program.
type Affiliate struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ReferralCode     string    `json:"referral_code"`
	CurrentTier      Tier      `json:"current_tier"`
	TotalReferrals   int       `json:"total_referrals"`
	ActiveReferrals  int       `json:"active_referrals"`
	LifetimeEarnings float64   `json:"lifetime_earnings_inr"`
	JoinedAt         time.Time `json:"joined_at"`
	Flagged          bool      `json:"flagged"` // anti-gaming flag
	FlagReason       string    `json:"flag_reason,omitempty"`
}

// Referral represents a referred user relationship.
type Referral struct {
	ReferrerID      string    `json:"referrer_id"`
	ReferredUserID  string    `json:"referred_user_id"`
	MonthlyVolume   float64   `json:"monthly_volume_inr"`
	MonthlyFees     float64   `json:"monthly_fees_inr"`
	RegisteredAt    time.Time `json:"registered_at"`
	IsActive        bool      `json:"is_active"`
}

// PayoutRecord tracks monthly rebate payouts.
type PayoutRecord struct {
	AffiliateID string    `json:"affiliate_id"`
	Month       string    `json:"month"` // YYYY-MM format
	GrossRebate float64   `json:"gross_rebate_inr"`
	Deductions  float64   `json:"deductions_inr"` // clawbacks from fraud
	NetPayout   float64   `json:"net_payout_inr"`
	PaidAt      time.Time `json:"paid_at"`
	Status      string    `json:"status"` // PENDING, PAID, HELD
}

// FraudSignal indicates potential gaming behaviour.
type FraudSignal struct {
	Type        string  `json:"type"`        // SELF_REFERRAL, WASH_VOLUME, RAPID_CHURN, IP_CLUSTER
	Score       float64 `json:"score"`       // 0.0 - 1.0
	Description string  `json:"description"`
}

// RebateEngine is the core affiliate referral rebate calculator.
type RebateEngine struct {
	mu          sync.RWMutex
	tiers       []TierConfig
	affiliates  map[string]*Affiliate
	referrals   map[string][]*Referral // affiliateID -> referrals
	payouts     []*PayoutRecord

	// Anti-gaming thresholds
	maxSelfReferralIPOverlap float64 // fraction of referrals from same IP cluster
	maxChurnRate             float64 // max fraction of referrals that churn within 30 days
	minDaysBetweenReferrals  int     // minimum days between referral signups from same source
}

// NewRebateEngine creates a new engine with default tier configuration.
func NewRebateEngine() *RebateEngine {
	return &RebateEngine{
		tiers:                    DefaultTierConfigs(),
		affiliates:               make(map[string]*Affiliate),
		referrals:                make(map[string][]*Referral),
		payouts:                  make([]*PayoutRecord, 0),
		maxSelfReferralIPOverlap: 0.5,
		maxChurnRate:             0.7,
		minDaysBetweenReferrals:  1,
	}
}

// RegisterAffiliate registers a new affiliate in the program.
func (e *RebateEngine) RegisterAffiliate(id, userID, referralCode string) (*Affiliate, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if id == "" || userID == "" || referralCode == "" {
		return nil, errors.New("id, userID, and referralCode are required")
	}
	if _, exists := e.affiliates[id]; exists {
		return nil, fmt.Errorf("affiliate %s already registered", id)
	}

	aff := &Affiliate{
		ID:           id,
		UserID:       userID,
		ReferralCode: referralCode,
		CurrentTier:  TierBronze,
		JoinedAt:     time.Now().UTC(),
	}
	e.affiliates[id] = aff
	e.referrals[id] = make([]*Referral, 0)
	return aff, nil
}

// AddReferral registers a new referred user under an affiliate.
func (e *RebateEngine) AddReferral(affiliateID, referredUserID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	aff, ok := e.affiliates[affiliateID]
	if !ok {
		return fmt.Errorf("affiliate %s not found", affiliateID)
	}

	ref := &Referral{
		ReferrerID:     affiliateID,
		ReferredUserID: referredUserID,
		RegisteredAt:   time.Now().UTC(),
		IsActive:       true,
	}
	e.referrals[affiliateID] = append(e.referrals[affiliateID], ref)
	aff.TotalReferrals++
	aff.ActiveReferrals++
	return nil
}

// RecordReferralVolume records monthly trading volume/fees for a referral.
func (e *RebateEngine) RecordReferralVolume(affiliateID, referredUserID string, volume, fees float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	refs, ok := e.referrals[affiliateID]
	if !ok {
		return fmt.Errorf("affiliate %s not found", affiliateID)
	}

	for _, r := range refs {
		if r.ReferredUserID == referredUserID {
			r.MonthlyVolume = volume
			r.MonthlyFees = fees
			return nil
		}
	}
	return fmt.Errorf("referral %s not found under affiliate %s", referredUserID, affiliateID)
}

// DetermineTier computes the appropriate tier for an affiliate based on current metrics.
func (e *RebateEngine) DetermineTier(affiliateID string) (Tier, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	aff, ok := e.affiliates[affiliateID]
	if !ok {
		return "", fmt.Errorf("affiliate %s not found", affiliateID)
	}

	// Calculate total monthly volume from referrals
	totalVolume := 0.0
	refs := e.referrals[affiliateID]
	for _, r := range refs {
		if r.IsActive {
			totalVolume += r.MonthlyVolume
		}
	}

	// Determine highest qualifying tier (iterate in reverse order)
	bestTier := TierBronze
	for i := len(e.tiers) - 1; i >= 0; i-- {
		tc := e.tiers[i]
		if totalVolume >= tc.MinMonthlyVolumeINR && aff.ActiveReferrals >= tc.MinReferrals {
			bestTier = tc.Tier
			break
		}
	}
	return bestTier, nil
}

// CalculateMonthlyRebate computes the gross rebate for an affiliate for the current month.
func (e *RebateEngine) CalculateMonthlyRebate(affiliateID string) (float64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	aff, ok := e.affiliates[affiliateID]
	if !ok {
		return 0, fmt.Errorf("affiliate %s not found", affiliateID)
	}
	if aff.Flagged {
		return 0, fmt.Errorf("affiliate %s is flagged for fraud: %s", affiliateID, aff.FlagReason)
	}

	// Find rebate percentage for current tier
	rebatePct := 0.0
	for _, tc := range e.tiers {
		if tc.Tier == aff.CurrentTier {
			rebatePct = tc.RebatePercent
			break
		}
	}

	// Sum fees from all active referrals
	totalFees := 0.0
	for _, r := range e.referrals[affiliateID] {
		if r.IsActive {
			totalFees += r.MonthlyFees
		}
	}

	rebate := totalFees * (rebatePct / 100.0)
	return math.Round(rebate*100) / 100, nil // Round to 2 decimals
}

// DetectFraud runs anti-gaming heuristics on an affiliate's referral network.
func (e *RebateEngine) DetectFraud(affiliateID string) ([]FraudSignal, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	_, ok := e.affiliates[affiliateID]
	if !ok {
		return nil, fmt.Errorf("affiliate %s not found", affiliateID)
	}

	refs := e.referrals[affiliateID]
	signals := make([]FraudSignal, 0)

	if len(refs) == 0 {
		return signals, nil
	}

	// Check 1: Rapid churn detection — high fraction of inactive referrals
	inactiveCount := 0
	for _, r := range refs {
		if !r.IsActive {
			inactiveCount++
		}
	}
	churnRate := float64(inactiveCount) / float64(len(refs))
	if churnRate > e.maxChurnRate {
		signals = append(signals, FraudSignal{
			Type:        "RAPID_CHURN",
			Score:       math.Min(churnRate, 1.0),
			Description: fmt.Sprintf("%.0f%% of referrals churned (threshold: %.0f%%)", churnRate*100, e.maxChurnRate*100),
		})
	}

	// Check 2: Wash volume — referrals with zero fees but non-zero volume
	washCount := 0
	for _, r := range refs {
		if r.IsActive && r.MonthlyVolume > 0 && r.MonthlyFees == 0 {
			washCount++
		}
	}
	if washCount > 0 {
		washScore := float64(washCount) / float64(len(refs))
		signals = append(signals, FraudSignal{
			Type:        "WASH_VOLUME",
			Score:       washScore,
			Description: fmt.Sprintf("%d referrals have volume but zero fees", washCount),
		})
	}

	// Check 3: Rapid signup burst — many referrals on the same day
	dayBuckets := make(map[string]int)
	for _, r := range refs {
		day := r.RegisteredAt.Format("2006-01-02")
		dayBuckets[day]++
	}
	for day, count := range dayBuckets {
		if count > 10 {
			signals = append(signals, FraudSignal{
				Type:        "SIGNUP_BURST",
				Score:       math.Min(float64(count)/50.0, 1.0),
				Description: fmt.Sprintf("%d referrals signed up on %s", count, day),
			})
		}
	}

	return signals, nil
}

// FlagAffiliate marks an affiliate as flagged for fraud, suspending payouts.
func (e *RebateEngine) FlagAffiliate(affiliateID, reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	aff, ok := e.affiliates[affiliateID]
	if !ok {
		return fmt.Errorf("affiliate %s not found", affiliateID)
	}
	aff.Flagged = true
	aff.FlagReason = reason
	return nil
}

// UpdateTier recalculates and persists the affiliate's tier.
func (e *RebateEngine) UpdateTier(affiliateID string) error {
	tier, err := e.DetermineTier(affiliateID)
	if err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.affiliates[affiliateID].CurrentTier = tier
	return nil
}

// GetAffiliate retrieves affiliate details by ID.
func (e *RebateEngine) GetAffiliate(affiliateID string) (*Affiliate, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	aff, ok := e.affiliates[affiliateID]
	if !ok {
		return nil, fmt.Errorf("affiliate %s not found", affiliateID)
	}
	return aff, nil
}
