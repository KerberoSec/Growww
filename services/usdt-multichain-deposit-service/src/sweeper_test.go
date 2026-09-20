package main

import (
	"context"
	"math/big"
	"testing"
)

func TestUSDTSweeper_IngestAndSweepPolicy(t *testing.T) {
	sweeper := NewUSDTSweeper()

	// 1. Ingest small deposit (500 USDT = 500,000,000 micro-units, below 1000 threshold)
	smallAmount := new(big.Int).Mul(big.NewInt(500), big.NewInt(1_000_000))
	dep1 := sweeper.IngestDeposit(
		ChainEthereum,
		"0x1111222233334444555566667777888899990000aaaaabbbbbcccccdddddeeeee",
		"0xUserEthAddress1",
		smallAmount,
	)

	if dep1.Swept {
		t.Fatalf("expected small deposit not to be swept immediately")
	}

	// 2. Ingest large deposit (2500 USDT, exceeds 1000 threshold)
	largeAmount := new(big.Int).Mul(big.NewInt(2500), big.NewInt(1_000_000))
	dep2 := sweeper.IngestDeposit(
		ChainTron,
		"trx_tx_hash_9999888877776666555544443333222211110000",
		"TUserTrxAddress1",
		largeAmount,
	)

	// Execute sweeps
	sweeps, err := sweeper.ExecutePendingSweeps(context.Background())
	if err != nil {
		t.Fatalf("expected successful sweep execution: %v", err)
	}

	// Exactly 1 deposit should have been swept (dep2)
	if len(sweeps) != 1 {
		t.Fatalf("expected 1 swept transaction, got %d", len(sweeps))
	}

	if dep1.Swept {
		t.Fatalf("dep1 should remain unswept")
	}
	if !dep2.Swept {
		t.Fatalf("dep2 should be marked swept")
	}
	if dep2.SweepTxHash == "" {
		t.Fatalf("expected non-empty sweep tx hash for dep2")
	}
}

func TestUSDTSweeper_MultichainVaultTargets(t *testing.T) {
	sweeper := NewUSDTSweeper()

	chains := []SupportedChain{ChainEthereum, ChainTron, ChainPolygon, ChainSolana}
	for _, chain := range chains {
		policy, exists := sweeper.policies[chain]
		if !exists {
			t.Fatalf("expected policy for chain %s", chain)
		}
		if policy.TargetVaultAddress == "" {
			t.Fatalf("expected non-empty cold vault address for %s", chain)
		}
		if policy.MinSweepAmountUSDT.Cmp(big.NewInt(0)) <= 0 {
			t.Fatalf("expected positive min sweep amount for %s", chain)
		}
	}
}
