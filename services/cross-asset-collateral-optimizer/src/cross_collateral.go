package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// SEBISecurityGroup defines haircut brackets under SEBI risk framework
type SEBISecurityGroup string

const (
	Group1Equity SEBISecurityGroup = "GROUP_1" // Bluechips (min 20% haircut)
	Group2Equity SEBISecurityGroup = "GROUP_2" // Midcaps (min 30% haircut)
	Group3Equity SEBISecurityGroup = "GROUP_3" // Smallcaps / High-volatility (min 50% haircut)
)

var (
	ErrInvalidQuantity    = errors.New("invalid share quantity or market price")
	ErrPledgeNotFound     = errors.New("pledge record not found")
	ErrInsufficientMargin = errors.New("insufficient free collateral margin to unpledge")
	ErrDuplicatePledge    = errors.New("pledge ID already exists")
)

type DematPledgeRecord struct {
	PledgeID          string            `json:"pledge_id"`
	UserID            string            `json:"user_id"`
	DematISIN         string            `json:"demat_isin"`
	SecurityGroup     SEBISecurityGroup `json:"security_group"`
	PledgedShares     uint64            `json:"pledged_shares"`
	MarketPriceINR    float64           `json:"market_price_inr"`
	HaircutPct        float64           `json:"haircut_pct"`
	CollateralValINR  float64           `json:"collateral_val_inr"`
	AllowedMarginUSDT float64           `json:"allowed_margin_usdt"`
	UsedMarginUSDT    float64           `json:"used_margin_usdt"`
	PledgedAt         time.Time         `json:"pledged_at"`
	LastValuedAt      time.Time         `json:"last_valued_at"`
}

type CrossCollateralEngine struct {
	mu          sync.RWMutex
	pledges     map[string]*DematPledgeRecord
	usdtInrRate float64
}

func NewCrossCollateralEngine(fxRate float64) *CrossCollateralEngine {
	if fxRate <= 0 {
		fxRate = 84.50
	}
	return &CrossCollateralEngine{
		pledges:     make(map[string]*DematPledgeRecord),
		usdtInrRate: fxRate,
	}
}

// ComputeHaircut calculates SEBI mandated VAR + ELM haircut percentage
func ComputeHaircut(group SEBISecurityGroup, baseVarPct float64) float64 {
	elm := 3.5 // Standard Extreme Loss Margin
	total := baseVarPct + elm
	switch group {
	case Group1Equity:
		return math.Max(20.0, total)
	case Group2Equity:
		return math.Max(30.0, total)
	case Group3Equity:
		return math.Max(50.0, total)
	default:
		return 50.0
	}
}

// PledgeDematStock pledges CDSL/NSDL demat shares to unlock crypto spot margin
func (e *CrossCollateralEngine) PledgeDematStock(
	id, userID, isin string,
	group SEBISecurityGroup,
	shares uint64,
	priceINR, baseVarPct float64,
) (*DematPledgeRecord, error) {
	if shares == 0 || priceINR <= 0 {
		return nil, ErrInvalidQuantity
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.pledges[id]; exists {
		return nil, ErrDuplicatePledge
	}

	haircutPct := ComputeHaircut(group, baseVarPct)
	grossValueINR := float64(shares) * priceINR
	collateralINR := grossValueINR * (1.0 - (haircutPct / 100.0))
	marginUSDT := collateralINR / e.usdtInrRate

	now := time.Now().UTC()
	rec := &DematPledgeRecord{
		PledgeID:          id,
		UserID:            userID,
		DematISIN:         isin,
		SecurityGroup:     group,
		PledgedShares:     shares,
		MarketPriceINR:    priceINR,
		HaircutPct:        haircutPct,
		CollateralValINR:  collateralINR,
		AllowedMarginUSDT: marginUSDT,
		UsedMarginUSDT:    0.0,
		PledgedAt:         now,
		LastValuedAt:      now,
	}

	e.pledges[id] = rec
	return rec, nil
}

// UpdateMarketPrice updates real-time valuation of pledged collateral
func (e *CrossCollateralEngine) UpdateMarketPrice(id string, newPriceINR float64) (*DematPledgeRecord, error) {
	if newPriceINR <= 0 {
		return nil, ErrInvalidQuantity
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	rec, exists := e.pledges[id]
	if !exists {
		return nil, ErrPledgeNotFound
	}

	rec.MarketPriceINR = newPriceINR
	grossValueINR := float64(rec.PledgedShares) * newPriceINR
	rec.CollateralValINR = grossValueINR * (1.0 - (rec.HaircutPct / 100.0))
	rec.AllowedMarginUSDT = rec.CollateralValINR / e.usdtInrRate
	rec.LastValuedAt = time.Now().UTC()

	return rec, nil
}

// UtilizeMargin reserves trading margin for crypto spot purchases
func (e *CrossCollateralEngine) UtilizeMargin(id string, amountUSDT float64) error {
	if amountUSDT <= 0 {
		return errors.New("utilization amount must be > 0")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	rec, exists := e.pledges[id]
	if !exists {
		return ErrPledgeNotFound
	}

	if rec.UsedMarginUSDT+amountUSDT > rec.AllowedMarginUSDT {
		return errors.New("exceeds available collateral margin capacity")
	}

	rec.UsedMarginUSDT += amountUSDT
	return nil
}

// ReleaseMargin frees up used margin upon closing crypto spot positions
func (e *CrossCollateralEngine) ReleaseMargin(id string, amountUSDT float64) error {
	if amountUSDT <= 0 {
		return errors.New("release amount must be > 0")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	rec, exists := e.pledges[id]
	if !exists {
		return ErrPledgeNotFound
	}

	if amountUSDT > rec.UsedMarginUSDT {
		rec.UsedMarginUSDT = 0
	} else {
		rec.UsedMarginUSDT -= amountUSDT
	}
	return nil
}

// CollateralHealth checks collateral coverage ratio
// HealthFactor = AllowedMarginUSDT / UsedMarginUSDT
func (e *CrossCollateralEngine) GetCollateralHealth(id string) (healthFactor float64, isMarginCall bool, isLiquidation bool, err error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rec, exists := e.pledges[id]
	if !exists {
		return 0, false, false, ErrPledgeNotFound
	}

	if rec.UsedMarginUSDT == 0 {
		return 999.0, false, false, nil // Infinite health
	}

	healthFactor = rec.AllowedMarginUSDT / rec.UsedMarginUSDT

	// Maintenance margin call threshold: health factor < 1.15
	if healthFactor < 1.15 {
		isMarginCall = true
	}
	// Immediate liquidation trigger: health factor < 1.00
	if healthFactor < 1.00 {
		isLiquidation = true
	}

	return healthFactor, isMarginCall, isLiquidation, nil
}

// UnpledgeDematStock verifies zero active margin utilization before releasing depository shares
func (e *CrossCollateralEngine) UnpledgeDematStock(id string) (*DematPledgeRecord, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	rec, exists := e.pledges[id]
	if !exists {
		return nil, ErrPledgeNotFound
	}

	if rec.UsedMarginUSDT > 0.001 {
		return nil, fmt.Errorf("%w: %.2f USDT currently actively locked in crypto positions",
			ErrInsufficientMargin, rec.UsedMarginUSDT)
	}

	delete(e.pledges, id)
	return rec, nil
}
