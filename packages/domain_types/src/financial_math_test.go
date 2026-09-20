package src

import (
	"testing"
)

func TestInsuranceFundWaterfall(t *testing.T) {
	loss := uint64(80_000_00000000) // 80 Lakhs INR loss

	layers := []WaterfallLayer{
		{Name: "DefaulterMargin",  CapacityPaise: 50_000_00000000}, // 50L
		{Name: "InsuranceFund",    CapacityPaise: 20_000_00000000}, // 20L
		{Name: "MutualisedSGF",    CapacityPaise: 30_000_00000000}, // 30L
		{Name: "CCContribution",   CapacityPaise: 10_000_00000000}, // 10L
		{Name: "MemberAssessment", CapacityPaise: 50_000_00000000}, // 50L
	}

	res := InsuranceFundWaterfall(loss, layers)

	if res.TotalLossPaise != loss {
		t.Fatalf("total loss mismatch: %d", res.TotalLossPaise)
	}
	if res.ResidualLossPaise != 0 {
		t.Fatalf("all layers should cover 80L loss, residual=%d", res.ResidualLossPaise)
	}
	if res.ADLTriggered {
		t.Fatalf("ADL should NOT be triggered when waterfall is sufficient")
	}

	// Layer 1 (DefaulterMargin) should absorb exactly 50L
	if res.AbsorbedByLayers[0].Absorbed != 50_000_00000000 {
		t.Fatalf("DefaulterMargin should absorb 50L, got %d", res.AbsorbedByLayers[0].Absorbed)
	}
	// Layer 2 (InsuranceFund) absorbs remaining 30L capped at 20L capacity
	if res.AbsorbedByLayers[1].Absorbed != 20_000_00000000 {
		t.Fatalf("InsuranceFund should absorb 20L, got %d", res.AbsorbedByLayers[1].Absorbed)
	}
	// Layer 3 (MutualisedSGF) absorbs remaining 10L
	if res.AbsorbedByLayers[2].Absorbed != 10_000_00000000 {
		t.Fatalf("MutualisedSGF should absorb 10L, got %d", res.AbsorbedByLayers[2].Absorbed)
	}
}

func TestADLTriggeredWhenWaterfallExhausted(t *testing.T) {
	loss := uint64(200_000_00000000) // 200L loss
	layers := []WaterfallLayer{
		{Name: "InsuranceFund", CapacityPaise: 50_000_00000000},
	}

	res := InsuranceFundWaterfall(loss, layers)
	if !res.ADLTriggered {
		t.Fatalf("ADL should be triggered when waterfall is exhausted")
	}
	if res.ResidualLossPaise != 150_000_00000000 {
		t.Fatalf("expected residual 150L, got %d", res.ResidualLossPaise)
	}
}

func TestMakerRebateSchedule(t *testing.T) {
	schedule := DefaultMakerRebateSchedule()

	// Tier 0: low volume
	tier0 := GetTierForVolume(schedule, 0)
	if tier0.MakerRebateBps != 1 || tier0.TakerFeeBps != 10 {
		t.Fatalf("tier 0 mismatch: %+v", tier0)
	}

	// Tier 2: 500K volume → -1 bps rebate
	tier2 := GetTierForVolume(schedule, 500_000_00000000)
	if tier2.MakerRebateBps != -1 {
		t.Fatalf("expected tier2 maker rebate -1 bps, got %d", tier2.MakerRebateBps)
	}

	// Compute rebate fee for 1 BTC at 25000 USD with -1 bps rebate
	tradeSizeE8 := uint64(100000000)      // 1 BTC
	markPriceE8 := uint64(2500000000000) // 25000 USD in e8
	fee := ComputeMakerFee(tradeSizeE8, markPriceE8, -1)
	if fee >= 0 {
		t.Fatalf("expected negative rebate fee for market maker, got %d", fee)
	}
}

func TestForexBuffer(t *testing.T) {
	// 1000 USD at 84.50 INR/USD, 100 bps (1%) buffer
	usdAmountE8 := uint64(100000000000) // 1000 USD in e8
	fxRate := 84.50
	bufferBps := uint32(100) // 1%

	res, err := ComputeForexBuffer(usdAmountE8, fxRate, bufferBps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Converted = 1000 * 84.50 = 84500 INR → 8450000000000 in e8
	expectedConverted := uint64(8450000000000)
	if res.ConvertedE8 != expectedConverted {
		t.Fatalf("expected converted %d, got %d", expectedConverted, res.ConvertedE8)
	}

	// Buffer = 1% of converted
	expectedBuffer := expectedConverted / 100
	if res.BufferAmountE8 != expectedBuffer {
		t.Fatalf("expected buffer %d, got %d", expectedBuffer, res.BufferAmountE8)
	}

	// Net = Converted - Buffer
	if res.NetAfterBufferE8 != expectedConverted-expectedBuffer {
		t.Fatalf("net after buffer mismatch")
	}

	// Invalid FX rate
	_, err = ComputeForexBuffer(usdAmountE8, 0, bufferBps)
	if err == nil {
		t.Fatalf("expected error for zero FX rate")
	}
}
