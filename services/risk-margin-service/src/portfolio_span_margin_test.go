package main

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestSPAN16ScenariosEvaluation(t *testing.T) {
	engine := NewPortfolioSPANEngine()
	engine.SetRiskParameters("BTC", SPANRiskParameters{
		PriceScanRangePct:     0.10, // 10%
		VolScanRangePct:       0.20, // 20%
		ExtremeMoveMultiplier: 2.0,
		ExtremeLossFraction:   0.35,
		IntraSpreadRateUSD:    100.0,
		ShortOptionMinRateUSD: 50.0,
	})

	// Pure Long 1 BTC Futures @ $60,000
	positions := []SPANPosition{
		{
			PositionID:     "P-1",
			UserID:         "TRADER-1",
			Symbol:         "BTC-PERP",
			CommodityClass: "BTC",
			Instrument:     InstrumentFutures,
			Quantity:       1.0,
			CurrentPrice:   60000.0,
			Delta:          1.0,
		},
	}

	result, err := engine.ComputePortfolioMargin("TRADER-1", positions)
	if err != nil {
		t.Fatalf("ComputePortfolioMargin failed: %v", err)
	}

	// For 1 BTC long, 10% price down is $6,000 loss.
	// In Scenario 13 or 14 (Price Down 3/3 = -10%), loss is 1.0 * 60000 * 0.10 = $6,000.
	// In Scenario 16 (Extreme Down -20%, 35% covered), loss is 1.0 * 60000 * 0.20 * 0.35 = $4,200.
	// Worst loss is $6,000.
	if math.Abs(result.ScanningRisk-6000.0) > 1.0 {
		t.Errorf("Expected scanning risk ~6000.0, got %.2f", result.ScanningRisk)
	}
	if result.NetSPANMarginRequired < 6000.0 {
		t.Errorf("Expected Net SPAN margin >= 6000, got %.2f", result.NetSPANMarginRequired)
	}
}

func TestPortfolioMarginOffsetCredit(t *testing.T) {
	engine := NewPortfolioSPANEngine()
	engine.SetRiskParameters("BTC", SPANRiskParameters{
		PriceScanRangePct:     0.10,
		VolScanRangePct:       0.15,
		ExtremeMoveMultiplier: 2.0,
		ExtremeLossFraction:   0.35,
	})
	engine.SetRiskParameters("ETH", SPANRiskParameters{
		PriceScanRangePct:     0.12,
		VolScanRangePct:       0.20,
		ExtremeMoveMultiplier: 2.0,
		ExtremeLossFraction:   0.35,
	})

	// 70% offset between BTC and ETH delta hedged positions
	engine.AddInterCommoditySpreadRule(InterCommoditySpreadRule{
		ClassA:           "BTC",
		ClassB:           "ETH",
		DeltaRatio:       15.0, // 1 BTC ~ 15 ETH
		CreditPercentage: 0.70, // 70% margin relief
	})

	// Hedged portfolio: Long 1 BTC and Short 15 ETH
	positions := []SPANPosition{
		{
			PositionID:     "BTC-POS",
			UserID:         "HEDGE-FUND-1",
			Symbol:         "BTC-PERP",
			CommodityClass: "BTC",
			Instrument:     InstrumentFutures,
			Quantity:       1.0,
			CurrentPrice:   60000.0,
			Delta:          1.0,
		},
		{
			PositionID:     "ETH-POS",
			UserID:         "HEDGE-FUND-1",
			Symbol:         "ETH-PERP",
			CommodityClass: "ETH",
			Instrument:     InstrumentFutures,
			Quantity:       -15.0,
			CurrentPrice:   4000.0,
			Delta:          -1.0,
		},
	}

	result, err := engine.ComputePortfolioMargin("HEDGE-FUND-1", positions)
	if err != nil {
		t.Fatalf("ComputePortfolioMargin failed: %v", err)
	}

	if result.InterCommodityCredit <= 0 {
		t.Errorf("Expected inter-commodity offset credit > 0, got %.2f", result.InterCommodityCredit)
	}
	if result.MarginSavingsUSD <= 0 {
		t.Errorf("Expected margin savings USD > 0, got %.2f", result.MarginSavingsUSD)
	}
	if result.NetSPANMarginRequired >= result.GrossMarginRequired {
		t.Errorf("Net SPAN margin (%.2f) must be less than gross margin (%.2f)",
			result.NetSPANMarginRequired, result.GrossMarginRequired)
	}
}

func TestIntraCommodityCalendarSpreadAndSOM(t *testing.T) {
	engine := NewPortfolioSPANEngine()
	engine.SetRiskParameters("NIFTY", SPANRiskParameters{
		PriceScanRangePct:     0.08,
		VolScanRangePct:       0.15,
		IntraSpreadRateUSD:    120.0,
		ShortOptionMinRateUSD: 75.0,
	})

	// Calendar spread: Long Near Month Future, Short Far Month Future
	positions := []SPANPosition{
		{
			PositionID:     "FUT-NEAR",
			UserID:         "SPREAD-1",
			Symbol:         "NIFTY-OCT",
			CommodityClass: "NIFTY",
			Instrument:     InstrumentFutures,
			Quantity:       2.0,
			CurrentPrice:   25000.0,
			ExpiryDays:     15,
			Delta:          1.0,
		},
		{
			PositionID:     "FUT-FAR",
			UserID:         "SPREAD-1",
			Symbol:         "NIFTY-NOV",
			CommodityClass: "NIFTY",
			Instrument:     InstrumentFutures,
			Quantity:       -2.0,
			CurrentPrice:   25100.0,
			ExpiryDays:     45,
			Delta:          -1.0,
		},
		// Deep OTM short call option for SOM test
		{
			PositionID:     "OPT-SHORT",
			UserID:         "SPREAD-1",
			Symbol:         "NIFTY-28000-CE",
			CommodityClass: "NIFTY",
			Instrument:     InstrumentCall,
			Quantity:       -3.0,
			CurrentPrice:   25000.0,
			Strike:         28000.0,
			ExpiryDays:     15,
			Delta:          0.02,
			Vega:           0.5,
			ImpliedVol:     0.18,
		},
	}

	result, err := engine.ComputePortfolioMargin("SPREAD-1", positions)
	if err != nil {
		t.Fatalf("ComputePortfolioMargin failed: %v", err)
	}

	// 2 pairs of calendar spread * 120 = $240
	if math.Abs(result.IntraCommodityCharge-240.0) > 1.0 {
		t.Errorf("Expected intra-commodity charge 240.0, got %.2f", result.IntraCommodityCharge)
	}

	// SOM: 3 short options * $75 = $225
	if math.Abs(result.ShortOptionMinimum-225.0) > 1.0 {
		t.Errorf("Expected SOM charge 225.0, got %.2f", result.ShortOptionMinimum)
	}
}

func TestSPANRiskDaemonLifecycleAndAlerts(t *testing.T) {
	engine := NewPortfolioSPANEngine()
	engine.SetRiskParameters("SOL", SPANRiskParameters{
		PriceScanRangePct:     0.15,
		VolScanRangePct:       0.25,
		ExtremeMoveMultiplier: 2.0,
		ExtremeLossFraction:   0.35,
	})

	cfg := SPANRiskDaemonConfig{
		EvaluationInterval:      100 * time.Millisecond,
		WarningThresholdPct:     75.0,
		MarginCallThresholdPct:  85.0,
		LiquidationThresholdPct: 95.0,
		WorkerPoolSize:          2,
	}

	daemon := NewSPANRiskDaemon(engine, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := daemon.Start(ctx); err != nil {
		t.Fatalf("Failed to start SPAN risk daemon: %v", err)
	}

	// Register user portfolio with high leverage (Margin ~ $3,000, Equity = $3,200 => Utilization ~ 93% => MARGIN_CALL)
	positions := []SPANPosition{
		{
			PositionID:     "POS-SOL-1",
			UserID:         "WHALE-101",
			Symbol:         "SOL-PERP",
			CommodityClass: "SOL",
			Instrument:     InstrumentFutures,
			Quantity:       100.0, // 100 SOL @ $200 = $20,000 notional, 15% scan = $3,000
			CurrentPrice:   200.0,
			Delta:          1.0,
		},
	}
	daemon.SetUserPortfolio("WHALE-101", positions, 3200.0)

	// Wait for alert on channel
	select {
	case alert := <-daemon.Alerts():
		if alert.UserID != "WHALE-101" {
			t.Errorf("Expected alert for WHALE-101, got %s", alert.UserID)
		}
		if alert.AlertLevel != "MARGIN_CALL" && alert.AlertLevel != "LIQUIDATION" {
			t.Errorf("Expected MARGIN_CALL or LIQUIDATION, got %s", alert.AlertLevel)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timed out waiting for risk daemon alert")
	}

	// Update price to trigger further breach
	daemon.UpdateMarketPrices("SOL-PERP", 220.0, 0.30)

	daemon.Stop()
}
