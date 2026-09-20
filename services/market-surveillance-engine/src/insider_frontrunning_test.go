package main

import (
	"strings"
	"testing"
	"time"
)

func TestSEBI_InsiderTradingDetection(t *testing.T) {
	engine := NewSEBIMarketSurveillanceEngine()

	symbol := "TCS"
	now := time.Now().UTC()
	announcementTime := now.Add(24 * time.Hour)

	// 1. Register Designated Person (Chief Financial Officer) in SDD
	cfoUserID := "USER-CFO-001"
	engine.RegisterDesignatedPerson(DesignatedPerson{
		PersonID:            cfoUserID,
		Name:                "Rajiv Malhotra",
		PAN:                 "AAAPM1234K",
		CompanySymbol:       symbol,
		Role:                "KMP_CFO",
		TradingWindowClosed: true,
		ConnectedAccounts:   []string{"USER-SPOUSE-002"},
	})

	// 2. Register UPSI Event: Massive Q3 earnings beat and share buyback
	upsi := UPSIEvent{
		EventID:                 "UPSI-TCS-Q3-2026",
		Symbol:                  symbol,
		EventType:               "EARNINGS_AND_BUYBACK",
		AnnouncementTimestamp:   announcementTime,
		ExpectedImpact:          "BULLISH",
		PostPriceChangePercent:  8.5,
		PreAnnouncementPriceE8:  3800e8,
		PostAnnouncementPriceE8: 4120e8, // +₹320 per share
	}
	engine.RegisterUPSIEvent(upsi)

	// 3. CFO executes massive aggressive BUY trades 6 hours before disclosure
	cfoTradeTime := announcementTime.Add(-6 * time.Hour)
	trades := []TradeExecution{
		{
			TradeID:   "TR-CFO-1",
			OrderID:   "ORD-CFO-1",
			UserID:    cfoUserID,
			Symbol:    symbol,
			Side:      "BUY",
			PriceE8:   3800e8,
			QtyE8:     10000e8, // 10,000 shares
			Timestamp: cfoTradeTime,
		},
		{
			TradeID:   "TR-CFO-2",
			OrderID:   "ORD-CFO-2",
			UserID:    cfoUserID,
			Symbol:    symbol,
			Side:      "BUY",
			PriceE8:   3805e8,
			QtyE8:     5000e8, // 5,000 shares
			Timestamp: cfoTradeTime.Add(5 * time.Minute),
		},
		// Legitimate unrelated retail trader
		{
			TradeID:   "TR-RETAIL-1",
			OrderID:   "ORD-RETAIL-1",
			UserID:    "USER-RETAIL-99",
			Symbol:    symbol,
			Side:      "BUY",
			PriceE8:   3800e8,
			QtyE8:     10e8, // 10 shares
			Timestamp: cfoTradeTime,
		},
	}

	// Baseline statistics: CFO normal volume is 500 shares, StdDev 200 shares
	baselines := map[string]UserBaselineStats{
		cfoUserID: {
			MeanDailyVolumeE8: 500e8,
			StdDevVolumeE8:    200e8,
		},
		"USER-RETAIL-99": {
			MeanDailyVolumeE8: 10e8,
			StdDevVolumeE8:    5e8,
		},
	}

	// 4. Run Insider Trading Detection across 48h lookback window
	alerts := engine.DetectInsiderTrading(upsi, trades, baselines, 48)
	if len(alerts) != 1 {
		t.Fatalf("expected exactly 1 insider trading alert, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.AccusedUserID != cfoUserID {
		t.Errorf("expected accused user to be CFO %s, got %s", cfoUserID, alert.AccusedUserID)
	}
	if alert.Typology != TypologyInsiderTradingUPSI {
		t.Errorf("expected typology %s, got %s", TypologyInsiderTradingUPSI, alert.Typology)
	}
	if !alert.DirectionalAlignment {
		t.Error("expected directional alignment to be true (BUY on BULLISH)")
	}
	if alert.VolumeZScore < 10.0 {
		t.Errorf("expected extremely high Volume Z-score, got %.2f", alert.VolumeZScore)
	}
	if alert.EstimatedIllicitGainINR <= 0 {
		t.Errorf("expected positive illicit gain calculation, got ₹%.2f", alert.EstimatedIllicitGainINR)
	}
	if !strings.Contains(alert.LegalCitation, "SEBI (Prohibition of Insider Trading) Regulations") {
		t.Errorf("expected SEBI PIT legal citation, got %s", alert.LegalCitation)
	}

	// 5. Generate SEBI Dossier
	dossier, err := engine.GenerateSEBIDossier(alert.AlertID)
	if err != nil {
		t.Fatalf("failed to generate SEBI dossier: %v", err)
	}
	if len(dossier.EvidenceTimeline) < 3 {
		t.Errorf("expected comprehensive evidence timeline, got %d items", len(dossier.EvidenceTimeline))
	}
	if len(dossier.TamperProofHash) != 64 {
		t.Errorf("expected 64-char SHA-256 evidence hash, got %s", dossier.TamperProofHash)
	}

	// 6. Test Protective Freeze
	engine.TriggerProtectiveFreeze(cfoUserID, "SEBI PIT Section 3 violation: pre-announcement trading")
	frozen, reason := engine.IsAccountFrozen(cfoUserID)
	if !frozen || !strings.Contains(reason, "SEBI PIT") {
		t.Errorf("expected account freeze for CFO, got frozen=%v, reason=%s", frozen, reason)
	}
}

func TestSEBI_FrontrunningDetection(t *testing.T) {
	engine := NewSEBIMarketSurveillanceEngine()

	symbol := "INFY"
	now := time.Now().UTC()

	// 1. Institutional Client submits large block BUY order of 100,000 shares
	parentTime := now.Add(10 * time.Second)
	parentOrder := ParentOrder{
		ParentOrderID:     "PARENT-INST-001",
		ClientUserID:      "INST-MUTUAL-FUND-A",
		Symbol:            symbol,
		Side:              "BUY",
		PriceE8:           1520e8,
		QtyE8:             100000e8,
		NotionalINR:       152000000.0, // ₹15.2 Crore block order
		ArrivalTimestamp:  parentTime,
		ExecutedTimestamp: parentTime.Add(50 * time.Millisecond),
	}

	// 2. Suspect Broker / Prop Desk places aggressive BUY order 120ms before parent arrival
	suspectUserID := "PROP-DESK-VIPER"
	suspectOrderTime := parentTime.Add(-120 * time.Millisecond)

	marketOrders := []OrderEvent{
		{
			OrderID:   "ORD-SUSPECT-BUY",
			UserID:    suspectUserID,
			Symbol:    symbol,
			Side:      "BUY",
			PriceE8:   1510e8,
			QtyE8:     5000e8, // 5,000 shares
			Timestamp: suspectOrderTime,
		},
		// Legitimate order after parent execution (not frontrunning)
		{
			OrderID:   "ORD-LATE-BUY",
			UserID:    "USER-LATE",
			Symbol:    symbol,
			Side:      "BUY",
			PriceE8:   1522e8,
			QtyE8:     100e8,
			Timestamp: parentTime.Add(2 * time.Second),
		},
	}

	// 3. Trade Executions: Suspect gets filled at 1510, parent fills pushing price to 1520, suspect unwinds at 1525
	trades := []TradeExecution{
		// Suspect entry fill (before parent execution)
		{
			TradeID:   "TR-SUSPECT-ENTRY",
			OrderID:   "ORD-SUSPECT-BUY",
			UserID:    suspectUserID,
			Symbol:    symbol,
			Side:      "BUY",
			PriceE8:   1510e8,
			QtyE8:     5000e8,
			Timestamp: parentTime.Add(-50 * time.Millisecond),
		},
		// Suspect exit unwind fill (15 seconds after parent execution at higher price)
		{
			TradeID:   "TR-SUSPECT-EXIT",
			OrderID:   "ORD-SUSPECT-SELL",
			UserID:    suspectUserID,
			Symbol:    symbol,
			Side:      "SELL",
			PriceE8:   1525e8,
			QtyE8:     5000e8,
			Timestamp: parentTime.Add(15 * time.Second),
		},
	}

	// 4. Run Frontrunning Detection with 5000ms max lead window
	alerts := engine.DetectFrontrunning([]ParentOrder{parentOrder}, marketOrders, trades, 5000)
	if len(alerts) != 1 {
		t.Fatalf("expected exactly 1 frontrunning alert, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.SuspectUserID != suspectUserID {
		t.Errorf("expected suspect %s, got %s", suspectUserID, alert.SuspectUserID)
	}
	if alert.ParentOrderID != parentOrder.ParentOrderID {
		t.Errorf("expected parent order ID %s, got %s", parentOrder.ParentOrderID, alert.ParentOrderID)
	}
	if alert.LeadTimeMs != 120 {
		t.Errorf("expected 120ms lead time, got %d ms", alert.LeadTimeMs)
	}
	if alert.ConfidenceRate < 0.95 {
		t.Errorf("expected confidence rate >= 0.95, got %.2f", alert.ConfidenceRate)
	}
	// Profit: (1525 - 1510) * 5000 = 15 * 5000 = ₹75,000
	if alert.EstimatedIllicitGainINR != 75000.0 {
		t.Errorf("expected ₹75,000 illicit profit, got ₹%.2f", alert.EstimatedIllicitGainINR)
	}
	if !strings.Contains(alert.LegalCitation, "SEBI (PFUTP) Regulations") {
		t.Errorf("expected SEBI PFUTP citation, got %s", alert.LegalCitation)
	}

	// 5. Generate SEBI Dossier for Frontrunning
	dossier, err := engine.GenerateSEBIDossier(alert.AlertID)
	if err != nil {
		t.Fatalf("failed to generate SEBI frontrunning dossier: %v", err)
	}
	if dossier.Typology != TypologyFrontrunningPFUTP {
		t.Errorf("expected typology %s, got %s", TypologyFrontrunningPFUTP, dossier.Typology)
	}
	if dossier.EstimatedIllicitGainINR != 75000.0 {
		t.Errorf("expected ₹75,000 illicit gain in dossier, got ₹%.2f", dossier.EstimatedIllicitGainINR)
	}
}
