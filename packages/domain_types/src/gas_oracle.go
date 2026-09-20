package src

import (
	"errors"
	"math/big"
	"sort"
)

const (
	ElasticityMultiplier uint64 = 2
	BaseFeeMaxChangeDenom uint64 = 8 // 12.5% max change per block
)

var (
	ErrZeroGasTarget = errors.New("gas_oracle: target gas cannot be zero")
)

// GasEstimate holds recommended fee parameters per tier (in Wei / e8).
type GasEstimate struct {
	BaseFee               uint64 `json:"base_fee"`
	MaxPriorityFeePerGas  uint64 `json:"max_priority_fee_per_gas"`
	MaxFeePerGas          uint64 `json:"max_fee_per_gas"`
}

// GasFeeRecommendations conveys Low, Standard, and Fast gas tiers.
type GasFeeRecommendations struct {
	Low      GasEstimate `json:"low"`
	Standard GasEstimate `json:"standard"`
	Fast     GasEstimate `json:"fast"`
}

// ComputeNextBlockBaseFee calculates next block base fee per EIP-1559 consensus rule.
func ComputeNextBlockBaseFee(currentBaseFee, gasUsed, gasTarget uint64) (uint64, error) {
	if gasTarget == 0 {
		return 0, ErrZeroGasTarget
	}

	if gasUsed == gasTarget {
		return currentBaseFee, nil
	}

	if gasUsed > gasTarget {
		gasDelta := gasUsed - gasTarget
		deltaBig := new(big.Int).Mul(big.NewInt(int64(currentBaseFee)), big.NewInt(int64(gasDelta)))
		denom := new(big.Int).Mul(big.NewInt(int64(gasTarget)), big.NewInt(int64(BaseFeeMaxChangeDenom)))
		feeDelta := new(big.Int).Div(deltaBig, denom).Uint64()
		if feeDelta == 0 {
			feeDelta = 1 // Min 1 Wei increment
		}
		return currentBaseFee + feeDelta, nil
	}

	// gasUsed < gasTarget
	gasDelta := gasTarget - gasUsed
	deltaBig := new(big.Int).Mul(big.NewInt(int64(currentBaseFee)), big.NewInt(int64(gasDelta)))
	denom := new(big.Int).Mul(big.NewInt(int64(gasTarget)), big.NewInt(int64(BaseFeeMaxChangeDenom)))
	feeDelta := new(big.Int).Div(deltaBig, denom).Uint64()

	if currentBaseFee <= feeDelta {
		return 1, nil // Never drop to zero
	}
	return currentBaseFee - feeDelta, nil
}

// ComputeGasOracle recommendations based on historical priority fees and next base fee.
func ComputeGasOracle(nextBaseFee uint64, historicalPriorityFees []uint64) GasFeeRecommendations {
	if len(historicalPriorityFees) == 0 {
		historicalPriorityFees = []uint64{1000000000} // 1 Gwei fallback
	}

	sorted := make([]uint64, len(historicalPriorityFees))
	copy(sorted, historicalPriorityFees)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	p10 := sorted[int(float64(len(sorted)-1)*0.10)]
	p50 := sorted[int(float64(len(sorted)-1)*0.50)]
	p90 := sorted[int(float64(len(sorted)-1)*0.90)]

	buildTier := func(priorityFee uint64) GasEstimate {
		// MaxFee = 2 * BaseFee + PriorityFee
		maxFee := (2 * nextBaseFee) + priorityFee
		return GasEstimate{
			BaseFee:              nextBaseFee,
			MaxPriorityFeePerGas: priorityFee,
			MaxFeePerGas:         maxFee,
		}
	}

	return GasFeeRecommendations{
		Low:      buildTier(p10),
		Standard: buildTier(p50),
		Fast:     buildTier(p90),
	}
}
