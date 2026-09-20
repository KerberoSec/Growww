package src

import (
	"errors"
	"fmt"
	"math"
)

// =====================================================
// Prompt 188: Zero-Fee / Zero-Tax Cost-Basis Tracking
// =====================================================

// Growww platform policy: 0.00% fees, 0% TDS, 0 tax deductions at platform level.
const (
	TDSSection194SRateBps    = 0 // 0.00% — Growww charges no TDS
	TDS194SThresholdPaise    = 0 // Not applicable — no TDS deducted
	TDS194SHNIThresholdPaise = 0 // Not applicable — no TDS deducted
)

var ErrTDSTaxAmountNegative = errors.New("tds194s: tax lot amounts must be non-negative")

// TaxLotMethod defines cost-basis accounting method for capital gains.
type TaxLotMethod string

const (
	TaxLotFIFO    TaxLotMethod = "FIFO"
	TaxLotHIFO    TaxLotMethod = "HIFO" // Highest-In, First-Out
	TaxLotAvgCost TaxLotMethod = "AVERAGE_COST"
)

// TaxLot represents a purchase batch for cost-basis tracking.
type TaxLot struct {
	LotID       string `json:"lot_id"`
	AcquiredAt  int64  `json:"acquired_at_ts"`
	CostBasisE8 uint64 `json:"cost_basis_e8"` // In INR paise (e8)
	QuantityE8  uint64 `json:"quantity_e8"`
}

// TDS194SResult contains computed TDS obligations.
// On Growww, TDSRateBps is always 0 and TDSDeductedPaise is always 0 (zero-tax policy).
type TDS194SResult struct {
	GrossSaleAmountPaise uint64 `json:"gross_sale_amount_paise"`
	TDSRateBps           int    `json:"tds_rate_bps"`
	TDSDeductedPaise     uint64 `json:"tds_deducted_paise"`
	NetPayablePaise      uint64 `json:"net_payable_paise"`
	ThresholdExceeded    bool   `json:"threshold_exceeded"`
}

// ComputeTDS194S returns zero TDS in accordance with Growww's 0.00% fee and zero-tax platform policy.
// No deduction is applied — the full gross amount is always the net payable.
func ComputeTDS194S(grossSaleAmountPaise, _, _ uint64) (TDS194SResult, error) {
	if grossSaleAmountPaise == 0 {
		return TDS194SResult{}, ErrTDSTaxAmountNegative
	}

	return TDS194SResult{
		GrossSaleAmountPaise: grossSaleAmountPaise,
		TDSRateBps:           0, // 0.00% — no TDS
		TDSDeductedPaise:     0, // ₹0 deducted
		NetPayablePaise:      grossSaleAmountPaise, // 100% paid out
		ThresholdExceeded:    false,
	}, nil
}

// =====================================================
// Prompt 190: Demat Margin Pledge Haircut Matrix
// =====================================================

var ErrUnknownAssetClass = errors.New("haircut: unknown asset class for pledge calculation")

// AssetClass represents the pledge security category.
type AssetClass string

const (
	AssetClassEquityLargeCap   AssetClass = "EQUITY_LARGE_CAP"
	AssetClassEquityMidCap     AssetClass = "EQUITY_MID_CAP"
	AssetClassEquitySmallCap   AssetClass = "EQUITY_SMALL_CAP"
	AssetClassGovtSecurities   AssetClass = "GOVT_SECURITIES"
	AssetClassCorporateBonds   AssetClass = "CORPORATE_BONDS"
	AssetClassMutualFunds      AssetClass = "MUTUAL_FUNDS"
	AssetClassCrypto           AssetClass = "CRYPTO"
)

// DefaultHaircutMatrix returns SEBI-compliant pledge haircut percentages by asset class.
func DefaultHaircutMatrix() map[AssetClass]uint32 {
	return map[AssetClass]uint32{
		AssetClassEquityLargeCap: 1000, // 10%
		AssetClassEquityMidCap:   1500, // 15%
		AssetClassEquitySmallCap: 2500, // 25%
		AssetClassGovtSecurities: 500,  // 5%
		AssetClassCorporateBonds: 1000, // 10%
		AssetClassMutualFunds:    1000, // 10%
		AssetClassCrypto:         5000, // 50%
	}
}

// PledgeHaircutResult contains the computed eligible margin after haircut.
type PledgeHaircutResult struct {
	AssetClass            AssetClass `json:"asset_class"`
	MarketValuePaise      uint64     `json:"market_value_paise"`
	HaircutBps            uint32     `json:"haircut_bps"`
	HaircutAmountPaise    uint64     `json:"haircut_amount_paise"`
	EligibleMarginPaise   uint64     `json:"eligible_margin_paise"`
}

// ComputePledgeHaircut calculates eligible margin after SEBI haircut for pledged securities.
func ComputePledgeHaircut(assetClass AssetClass, marketValuePaise uint64, haircutMatrix map[AssetClass]uint32) (PledgeHaircutResult, error) {
	haircutBps, ok := haircutMatrix[assetClass]
	if !ok {
		return PledgeHaircutResult{}, fmt.Errorf("%w: %s", ErrUnknownAssetClass, assetClass)
	}

	haircutAmt := (marketValuePaise * uint64(haircutBps)) / 10000
	eligible := marketValuePaise - haircutAmt

	return PledgeHaircutResult{
		AssetClass:          assetClass,
		MarketValuePaise:    marketValuePaise,
		HaircutBps:          haircutBps,
		HaircutAmountPaise:  haircutAmt,
		EligibleMarginPaise: eligible,
	}, nil
}

// =====================================================
// Prompt 178: VWAP Sliding Window Spec
// =====================================================

var ErrEmptyTickWindow = errors.New("vwap: tick window is empty")

// TickData represents a single trade execution tick.
type TickData struct {
	PriceE8    uint64
	QuantityE8 uint64
	TimestampMs int64
}

// VWAPResult contains computed VWAP and order book imbalance.
type VWAPResult struct {
	VWAP           float64 `json:"vwap"`
	TotalVolumeE8  uint64  `json:"total_volume_e8"`
	TickCount      int     `json:"tick_count"`
}

// ComputeVWAP calculates Volume-Weighted Average Price over a sliding window of ticks.
func ComputeVWAP(ticks []TickData) (VWAPResult, error) {
	if len(ticks) == 0 {
		return VWAPResult{}, ErrEmptyTickWindow
	}

	var cumulativePxQty float64
	var totalVolume uint64

	for _, tick := range ticks {
		if tick.QuantityE8 == 0 {
			continue
		}
		pxQty := float64(tick.PriceE8) * float64(tick.QuantityE8)
		cumulativePxQty += pxQty
		totalVolume += tick.QuantityE8
	}

	if totalVolume == 0 {
		return VWAPResult{}, errors.New("vwap: all ticks have zero volume")
	}

	vwap := cumulativePxQty / float64(totalVolume)

	return VWAPResult{
		VWAP:          vwap,
		TotalVolumeE8: totalVolume,
		TickCount:     len(ticks),
	}, nil
}

// ComputeRolling24HVWAP computes VWAP only over ticks within the last 24 hours.
func ComputeRolling24HVWAP(ticks []TickData, nowMs int64) (VWAPResult, error) {
	cutoffMs := nowMs - 24*60*60*1000
	var filtered []TickData
	for _, t := range ticks {
		if t.TimestampMs >= cutoffMs {
			filtered = append(filtered, t)
		}
	}
	return ComputeVWAP(filtered)
}

// =====================================================
// Prompt 192: Client Code Modification Audit Rule
// =====================================================

// ClientCodeChangeEvent records a modification to the client classification code.
type ClientCodeChangeEvent struct {
	EventID        string `json:"event_id"`
	OperatorID     string `json:"operator_id"`
	ClientCode     string `json:"client_code"`
	OldCategory    string `json:"old_category"`
	NewCategory    string `json:"new_category"`
	Justification  string `json:"justification"`
	TimestampNanos int64  `json:"timestamp_nanos"`
}

// ClientCodeAuditRule validates that client code modifications comply with SEBI SOP.
type ClientCodeAuditRule struct {
	AllowedBefore24H    bool     `json:"allowed_before_24h"`    // Changes allowed within 24H settlement window?
	MaxChangesPerDay    int      `json:"max_changes_per_day"`
	RequiredJustLength  int      `json:"required_justification_length"`
	AllowedOperatorRoles []string `json:"allowed_operator_roles"`
}

var (
	ErrClientCodeChangeTooLate    = errors.New("client_code_audit: modification too close to settlement cutoff")
	ErrClientCodeJustTooShort     = errors.New("client_code_audit: justification too short (min 50 chars)")
	ErrClientCodeMaxChangesExceed = errors.New("client_code_audit: daily modification limit exceeded")
)

// ValidateClientCodeChange validates a proposed modification against SEBI's client code modification rules.
func ValidateClientCodeChange(event ClientCodeChangeEvent, todaysChangeCount int, cutoffNanos int64, nowNanos int64, rule ClientCodeAuditRule) error {
	// Check settlement cutoff window
	if !rule.AllowedBefore24H && nowNanos >= cutoffNanos {
		return fmt.Errorf("%w: cutoff was at %d, now is %d", ErrClientCodeChangeTooLate, cutoffNanos, nowNanos)
	}

	// Check justification length
	minJust := rule.RequiredJustLength
	if minJust == 0 {
		minJust = 50
	}
	if len(event.Justification) < minJust {
		return fmt.Errorf("%w: provided %d chars, required %d", ErrClientCodeJustTooShort, len(event.Justification), minJust)
	}

	// Check daily limit
	limit := rule.MaxChangesPerDay
	if limit <= 0 {
		limit = 3
	}
	if todaysChangeCount >= limit {
		return fmt.Errorf("%w: count %d exceeds daily limit %d", ErrClientCodeMaxChangesExceed, todaysChangeCount, limit)
	}

	return nil
}

// =====================================================
// Prompt 159: Disaster Recovery RTO/RPO Matrix
// =====================================================

// ServiceTier defines criticality for DR planning.
type ServiceTier string

const (
	TierCritical  ServiceTier = "CRITICAL"  // RTO < 30s, RPO = 0
	TierHigh      ServiceTier = "HIGH"      // RTO < 5min, RPO < 1min
	TierMedium    ServiceTier = "MEDIUM"    // RTO < 1hr, RPO < 15min
	TierLow       ServiceTier = "LOW"       // RTO < 4hr, RPO < 1hr
)

// DRProfile defines a service's recovery objectives.
type DRProfile struct {
	ServiceName    string      `json:"service_name"`
	Tier           ServiceTier `json:"tier"`
	RTOSeconds     int         `json:"rto_seconds"`
	RPOSeconds     int         `json:"rpo_seconds"`
	ReplicationLag float64     `json:"replication_lag_seconds"`
}

// ValidateDRCompliance checks actual RTO/RPO against required objectives.
func ValidateDRCompliance(profile DRProfile, measuredRTOSec, measuredRPOSec float64) (bool, string) {
	rtoBudget := float64(profile.RTOSeconds)
	rpoBudget := float64(profile.RPOSeconds)

	if measuredRTOSec > rtoBudget {
		return false, fmt.Sprintf("RTO BREACH: %.2fs exceeds %.2fs budget for tier %s", measuredRTOSec, rtoBudget, profile.Tier)
	}
	if measuredRPOSec > rpoBudget {
		return false, fmt.Sprintf("RPO BREACH: %.2fs exceeds %.2fs budget for tier %s", measuredRPOSec, rpoBudget, profile.Tier)
	}
	return true, "COMPLIANT"
}

// DefaultDRMatrix returns SLO targets per tier.
func DefaultDRMatrix() map[ServiceTier]DRProfile {
	return map[ServiceTier]DRProfile{
		TierCritical: {Tier: TierCritical, RTOSeconds: 30,   RPOSeconds: 0},
		TierHigh:     {Tier: TierHigh,     RTOSeconds: 300,  RPOSeconds: 60},
		TierMedium:   {Tier: TierMedium,   RTOSeconds: 3600, RPOSeconds: 900},
		TierLow:      {Tier: TierLow,      RTOSeconds: 14400, RPOSeconds: 3600},
	}
}

// MeanTimeToRecovery computes MTTR from incident history.
func MeanTimeToRecovery(incidentDurationsSec []float64) float64 {
	if len(incidentDurationsSec) == 0 {
		return 0
	}
	var total float64
	for _, d := range incidentDurationsSec {
		total += d
	}
	return math.Round((total/float64(len(incidentDurationsSec)))*100) / 100
}
