package main

import (
	"math"
	"testing"
	"time"
)

func TestForwardStockSplitReconciliation(t *testing.T) {
	rec := NewCorporateActionReconciler()

	splitDef := CorporateActionDefinition{
		ActionID:             "SPLIT-INFY-5TO1",
		Symbol:               "INFY",
		Type:                 ActionForwardSplit,
		SplitNumerator:       5.0,
		SplitDenominator:     1.0,
		PreActionRefPrice:    1500.0,
		ExDateRefPrice:       300.0,
		RecordDate:           time.Now().Add(-24 * time.Hour),
		ExDate:               time.Now(),
		AllowFractionalShare: false,
	}

	if err := rec.RegisterAction(splitDef); err != nil {
		t.Fatalf("RegisterAction failed: %v", err)
	}

	// Setup Pre-Action Holdings: 3 users
	_ = rec.RecordSnapshot(splitDef.ActionID, "USER-A", 100.0) // 100 * 5 = 500
	_ = rec.RecordSnapshot(splitDef.ActionID, "USER-B", 200.0) // 200 * 5 = 1000
	_ = rec.RecordSnapshot(splitDef.ActionID, "USER-C", 50.0)  // 50 * 5 = 250
	// Total pre-action shares = 350. Expected post-action = 1750.

	// Setup Open Order: Buy 10 shares @ ₹1500 (Notional = ₹15,000)
	_ = rec.RegisterOpenOrder(splitDef.ActionID, "ORD-1", "USER-A", "INFY", 10.0, 1500.0)

	// Setup Derivative Contract: Call 1600 Strike, Lot size 100
	_ = rec.RegisterDerivativeContract(splitDef.ActionID, "INFY-24OCT-1600-CE", 1600.0, 100.0)

	// Execute Adjustment
	if err := rec.ExecuteAdjustment(splitDef.ActionID); err != nil {
		t.Fatalf("ExecuteAdjustment failed: %v", err)
	}

	// Verify Open Order adjustment
	ord := rec.openOrders[splitDef.ActionID]["ORD-1"]
	if ord.AdjustedQty != 50.0 || ord.AdjustedPrice != 300.0 {
		t.Errorf("Order not adjusted correctly: qty=%.2f, price=%.2f", ord.AdjustedQty, ord.AdjustedPrice)
	}
	if math.Abs(ord.AdjustedNotional-ord.OriginalNotional) > 0.01 {
		t.Errorf("Order notional invariant breached: original=%.2f, adjusted=%.2f", ord.OriginalNotional, ord.AdjustedNotional)
	}

	// Verify Derivative Contract adjustment
	deriv := rec.derivatives[splitDef.ActionID]["INFY-24OCT-1600-CE"]
	if deriv.AdjustedStrike != 320.0 || deriv.AdjustedLotSize != 500.0 {
		t.Errorf("Derivative not adjusted correctly: strike=%.2f, lotSize=%.2f", deriv.AdjustedStrike, deriv.AdjustedLotSize)
	}

	// Reconcile with Depository reported shares = 1750
	report, err := rec.ReconcileAndAudit(splitDef.ActionID, 1750.0)
	if err != nil {
		t.Fatalf("ReconcileAndAudit failed: %v", err)
	}

	if report.Status != StatusReconciled {
		t.Errorf("Expected status RECONCILED, got %s", report.Status)
	}
	if report.TotalPostActionShares != 1750.0 {
		t.Errorf("Expected 1750 total post shares, got %.2f", report.TotalPostActionShares)
	}
	if len(report.Discrepancies) != 0 {
		t.Errorf("Expected 0 discrepancies, got %d", len(report.Discrepancies))
	}
	// Market cap invariant
	if math.Abs(report.PreActionMarketCap-report.PostActionMarketCap) > 1.0 {
		t.Errorf("Market cap invariant breached: pre=%.2f, post=%.2f", report.PreActionMarketCap, report.PostActionMarketCap)
	}
}

func TestReverseSplitWithCashInLieu(t *testing.T) {
	rec := NewCorporateActionReconciler()

	// 1-for-10 reverse split
	reverseDef := CorporateActionDefinition{
		ActionID:             "REV-SPLIT-1TO10",
		Symbol:               "PENNY",
		Type:                 ActionReverseSplit,
		SplitNumerator:       1.0,
		SplitDenominator:     10.0,
		PreActionRefPrice:    2.0,
		ExDateRefPrice:       20.0,
		AllowFractionalShare: false, // Fraction paid in cash
	}

	_ = rec.RegisterAction(reverseDef)

	// User holds 25 shares: 25 * 0.1 = 2.5 => 2 whole shares + 0.5 fractional * $20 = $10 Cash-in-lieu
	_ = rec.RecordSnapshot(reverseDef.ActionID, "USER-ODD", 25.0)

	_ = rec.ExecuteAdjustment(reverseDef.ActionID)

	holding := rec.userHoldings[reverseDef.ActionID]["USER-ODD"]
	if holding.PostActionQuantity != 2.0 {
		t.Errorf("Expected 2 whole shares, got %.2f", holding.PostActionQuantity)
	}
	if holding.FractionalRemainder != 0.5 {
		t.Errorf("Expected 0.5 fractional remainder, got %.2f", holding.FractionalRemainder)
	}
	if holding.CashInLieuUSD != 10.0 {
		t.Errorf("Expected $10 cash in lieu, got %.2f", holding.CashInLieuUSD)
	}

	report, err := rec.ReconcileAndAudit(reverseDef.ActionID, 2.0)
	if err != nil {
		t.Fatalf("Audit failed: %v", err)
	}
	if report.Status != StatusReconciled {
		t.Errorf("Expected status RECONCILED, got %s", report.Status)
	}
	if report.TotalFractionalCashUSD != 10.0 {
		t.Errorf("Expected total cash in lieu $10.0, got %.2f", report.TotalFractionalCashUSD)
	}
}

func TestBonusIssueReconciliation(t *testing.T) {
	rec := NewCorporateActionReconciler()

	// 3:1 bonus issue: Multiplier is 1 + 3/1 = 4.0
	bonusDef := CorporateActionDefinition{
		ActionID:             "BONUS-RELIANCE-3TO1",
		Symbol:               "RELIANCE",
		Type:                 ActionBonusIssue,
		SplitNumerator:       3.0,
		SplitDenominator:     1.0,
		PreActionRefPrice:    3000.0,
		ExDateRefPrice:       750.0,
		AllowFractionalShare: false,
	}

	_ = rec.RegisterAction(bonusDef)
	_ = rec.RecordSnapshot(bonusDef.ActionID, "INVESTOR-1", 50.0) // 50 * 4 = 200 shares

	_ = rec.ExecuteAdjustment(bonusDef.ActionID)

	holding := rec.userHoldings[bonusDef.ActionID]["INVESTOR-1"]
	if holding.PostActionQuantity != 200.0 {
		t.Errorf("Expected 200 shares after 3:1 bonus issue, got %.2f", holding.PostActionQuantity)
	}

	report, err := rec.ReconcileAndAudit(bonusDef.ActionID, 200.0)
	if err != nil {
		t.Fatalf("Reconcile error: %v", err)
	}
	if report.Status != StatusReconciled {
		t.Errorf("Expected RECONCILED, got %s", report.Status)
	}
}

func TestDepositoryDiscrepancyDetection(t *testing.T) {
	rec := NewCorporateActionReconciler()

	splitDef := CorporateActionDefinition{
		ActionID:         "SPLIT-TCS-2TO1",
		Symbol:           "TCS",
		Type:             ActionForwardSplit,
		SplitNumerator:   2.0,
		SplitDenominator: 1.0,
	}

	_ = rec.RegisterAction(splitDef)
	_ = rec.RecordSnapshot(splitDef.ActionID, "USER-1", 100.0) // 200 expected
	_ = rec.ExecuteAdjustment(splitDef.ActionID)

	// Depository clearing report states only 150 shares (mismatch of 50 shares!)
	report, err := rec.ReconcileAndAudit(splitDef.ActionID, 150.0)
	if err != nil {
		t.Fatalf("Audit error: %v", err)
	}

	if report.Status != StatusDiscrepancy {
		t.Errorf("Expected DISCREPANCY_DETECTED, got %s", report.Status)
	}
	if len(report.Discrepancies) == 0 {
		t.Fatal("Expected discrepancies logged in report, found none")
	}
	if report.Discrepancies[0].Severity != "CRITICAL" {
		t.Errorf("Expected CRITICAL severity, got %s", report.Discrepancies[0].Severity)
	}
}
