package src

import (
	"errors"
	"math"
)

// =====================================================
// Prompt 180: Insurance Fund Waterfall Formula
// =====================================================

var (
	ErrInsuranceFundDepleted     = errors.New("insurance_fund: fund depleted, triggering ADL")
	ErrWaterfallLayerExceeded    = errors.New("insurance_fund: all waterfall layers exhausted")
)

// WaterfallLayer represents one tier in the default waterfall.
type WaterfallLayer struct {
	Name            string  `json:"name"`
	CapacityPaise   uint64  `json:"capacity_paise"`
	AllocatedPaise  uint64  `json:"allocated_paise"`
	Absorbed        uint64  `json:"absorbed"`
}

// WaterfallResult captures the result of loss absorption through the waterfall.
type WaterfallResult struct {
	TotalLossPaise       uint64           `json:"total_loss_paise"`
	AbsorbedByLayers     []WaterfallLayer `json:"absorbed_by_layers"`
	ResidualLossPaise    uint64           `json:"residual_loss_paise"`
	ADLTriggered         bool             `json:"adl_triggered"`
}

// InsuranceFundWaterfall applies Growww SGF default loss waterfall:
// Layer 1: Defaulting member's posted margin
// Layer 2: Insurance Fund (5% of remaining)
// Layer 3: Mutualised SGF contributions
// Layer 4: Clearing Corporation's skin-in-the-game
// Layer 5: Assessment on surviving members
func InsuranceFundWaterfall(totalLossPaise uint64, layers []WaterfallLayer) WaterfallResult {
	result := WaterfallResult{
		TotalLossPaise:   totalLossPaise,
		AbsorbedByLayers: make([]WaterfallLayer, len(layers)),
	}
	copy(result.AbsorbedByLayers, layers)

	remaining := totalLossPaise
	for i := range result.AbsorbedByLayers {
		if remaining == 0 {
			break
		}
		layer := &result.AbsorbedByLayers[i]
		available := layer.CapacityPaise - layer.AllocatedPaise
		if available == 0 {
			continue
		}
		absorb := remaining
		if absorb > available {
			absorb = available
		}
		layer.Absorbed = absorb
		layer.AllocatedPaise += absorb
		remaining -= absorb
	}

	result.ResidualLossPaise = remaining
	result.ADLTriggered = remaining > 0
	return result
}

// =====================================================
// Prompt 179: Market Making Rebate Accounting Model
// =====================================================

// MakerRebateTier defines volume-tiered rebate schedule.
type MakerRebateTier struct {
	MinMonthlyVolumeE8 uint64  `json:"min_monthly_volume_e8"`
	MakerRebateBps     int32   `json:"maker_rebate_bps"`  // Negative = rebate
	TakerFeeBps        int32   `json:"taker_fee_bps"`
}

// DefaultMakerRebateSchedule returns Growww NBSE fee schedule.
func DefaultMakerRebateSchedule() []MakerRebateTier {
	return []MakerRebateTier{
		{MinMonthlyVolumeE8: 0,                  MakerRebateBps: 1,   TakerFeeBps: 10},   // Tier 0: 0.01% maker, 0.10% taker
		{MinMonthlyVolumeE8: 100_000_00000000,   MakerRebateBps: 0,   TakerFeeBps: 8},    // Tier 1: 0% maker, 0.08% taker
		{MinMonthlyVolumeE8: 500_000_00000000,   MakerRebateBps: -1,  TakerFeeBps: 6},    // Tier 2: -0.01% rebate, 0.06% taker
		{MinMonthlyVolumeE8: 5_000_000_00000000, MakerRebateBps: -2,  TakerFeeBps: 4},    // Tier 3: -0.02% rebate, 0.04% taker
	}
}

// GetTierForVolume returns the applicable rebate tier.
func GetTierForVolume(schedule []MakerRebateTier, monthlyVolumeE8 uint64) MakerRebateTier {
	applicable := schedule[0]
	for _, tier := range schedule {
		if monthlyVolumeE8 >= tier.MinMonthlyVolumeE8 {
			applicable = tier
		}
	}
	return applicable
}

// ComputeMakerFee computes the maker fee or rebate in e8 for a given trade.
// Negative value = rebate credited to market maker.
func ComputeMakerFee(tradeSizeE8 uint64, markPriceE8 uint64, rebateBps int32) int64 {
	notionalE8 := float64(tradeSizeE8) * float64(markPriceE8) / 1e8
	feeE8 := notionalE8 * float64(rebateBps) / 10000.0
	return int64(math.Round(feeE8))
}

// =====================================================
// Prompt 191: Cross-Currency Forex Buffer Math
// =====================================================

var ErrInvalidFXRate = errors.New("forex_buffer: FX rate must be positive")

// ForexBufferResult contains computed INR/USD buffer and converted amount.
type ForexBufferResult struct {
	OriginalAmountE8  uint64  `json:"original_amount_e8"`
	FxRate            float64 `json:"fx_rate"`
	ConvertedE8       uint64  `json:"converted_e8"`
	BufferAmountE8    uint64  `json:"buffer_amount_e8"`
	NetAfterBufferE8  uint64  `json:"net_after_buffer_e8"`
	BufferBps         uint32  `json:"buffer_bps"`
}

// ComputeForexBuffer converts USD to INR with safety buffer applied.
// BufferBps represents percentage buffer held as volatility hedge (e.g. 100 bps = 1%).
func ComputeForexBuffer(usdAmountE8 uint64, usdInrRate float64, bufferBps uint32) (ForexBufferResult, error) {
	if usdInrRate <= 0 {
		return ForexBufferResult{}, ErrInvalidFXRate
	}

	converted := float64(usdAmountE8) * usdInrRate
	convertedE8 := uint64(converted)

	bufferE8 := (convertedE8 * uint64(bufferBps)) / 10000
	netE8 := convertedE8 - bufferE8

	return ForexBufferResult{
		OriginalAmountE8: usdAmountE8,
		FxRate:           usdInrRate,
		ConvertedE8:      convertedE8,
		BufferAmountE8:   bufferE8,
		NetAfterBufferE8: netE8,
		BufferBps:        bufferBps,
	}, nil
}
