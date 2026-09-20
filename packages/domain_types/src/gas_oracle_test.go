package src

import (
	"testing"
)

func TestEIP1559BaseFeeAdjustment(t *testing.T) {
	currentBaseFee := uint64(1000000000) // 1 Gwei
	gasTarget := uint64(15000000)        // 15M gas target (30M limit)

	// 1. Exact target usage -> base fee unchanged
	nextFee, err := ComputeNextBlockBaseFee(currentBaseFee, gasTarget, gasTarget)
	if err != nil || nextFee != currentBaseFee {
		t.Fatalf("expected unchanged base fee for exact target usage, got %d", nextFee)
	}

	// 2. 100% full block (30M gas used = 2x target) -> max 12.5% increase
	// feeDelta = 1000000000 * 15000000 / (15000000 * 8) = 125,000,000
	fullBlockGas := uint64(30000000)
	nextFeeFull, err := ComputeNextBlockBaseFee(currentBaseFee, fullBlockGas, gasTarget)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedFullFee := currentBaseFee + (currentBaseFee / 8)
	if nextFeeFull != expectedFullFee {
		t.Fatalf("expected 12.5%% increase to %d, got %d", expectedFullFee, nextFeeFull)
	}

	// 3. 0% empty block -> max 12.5% decrease
	nextFeeEmpty, err := ComputeNextBlockBaseFee(currentBaseFee, 0, gasTarget)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedEmptyFee := currentBaseFee - (currentBaseFee / 8)
	if nextFeeEmpty != expectedEmptyFee {
		t.Fatalf("expected 12.5%% decrease to %d, got %d", expectedEmptyFee, nextFeeEmpty)
	}
}

func TestGasOracleRecommendations(t *testing.T) {
	nextBaseFee := uint64(20000000000) // 20 Gwei
	prioSamples := []uint64{
		1000000000, 1500000000, 2000000000, 2500000000, 3000000000,
		3500000000, 4000000000, 5000000000, 6000000000, 10000000000,
	}

	recs := ComputeGasOracle(nextBaseFee, prioSamples)

	if recs.Low.MaxPriorityFeePerGas >= recs.Standard.MaxPriorityFeePerGas {
		t.Fatalf("low priority fee must be <= standard")
	}
	if recs.Standard.MaxPriorityFeePerGas >= recs.Fast.MaxPriorityFeePerGas {
		t.Fatalf("standard priority fee must be <= fast")
	}

	// Fast MaxFee = 2 * 20 Gwei + fast priority
	expectedFastMaxFee := (2 * nextBaseFee) + recs.Fast.MaxPriorityFeePerGas
	if recs.Fast.MaxFeePerGas != expectedFastMaxFee {
		t.Fatalf("expected fast max fee %d, got %d", expectedFastMaxFee, recs.Fast.MaxFeePerGas)
	}
}
