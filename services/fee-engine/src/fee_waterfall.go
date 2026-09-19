package main

import (
	"fmt"
	"time"
)

type FeeAllocationBreakdown struct {
	ExecutionID       string    `json:"execution_id"`
	NotionalTurnover  float64   `json:"notional_turnover"`
	PlatformFeeBps    float64   `json:"platform_fee_bps"` // 0.00% universal policy
	PlatformFeeTotal  float64   `json:"platform_fee_total"` // ₹0.00
	SGFReservePaise   uint64    `json:"sgf_reserve_paise"`  // 25% of net revenues to Settlement Guarantee Fund
	TreasuryVault     uint64    `json:"treasury_vault_paise"`
	ComputedAt        time.Time `json:"computed_at"`
}

type FeeEngine struct {
	PlatformFeeRate float64 // 0.00% Launch Invariant
}

func NewFeeEngine() *FeeEngine {
	return &FeeEngine{
		PlatformFeeRate: 0.0000, // 0.00% universal fee
	}
}

// CalculateTradeFees computes fee breakdown enforcing zero-fee launch invariant
func (e *FeeEngine) CalculateTradeFees(execID string, turnoverINR float64) FeeAllocationBreakdown {
	platformFee := turnoverINR * e.PlatformFeeRate // Exactly 0.00

	return FeeAllocationBreakdown{
		ExecutionID:      execID,
		NotionalTurnover: turnoverINR,
		PlatformFeeBps:   0.00,
		PlatformFeeTotal: platformFee,
		SGFReservePaise:  0,
		TreasuryVault:    0,
		ComputedAt:       time.Now().UTC(),
	}
}

func main() {
	engine := NewFeeEngine()
	res := engine.CalculateTradeFees("exec-101", 50000.0)
	fmt.Printf("[Fee Engine] Turnover: ₹%.2f, Platform Fee: ₹%.2f (0.00%% policy)\n",
		res.NotionalTurnover, res.PlatformFeeTotal)
}
