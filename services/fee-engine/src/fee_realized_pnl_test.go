package main

import (
	"math"
	"testing"
	"time"
)

func TestComprehensiveFeeEngine_VIPAndDiscounts(t *testing.T) {
	cfg := DefaultFeeScheduleConfig()
	engine := NewComprehensiveFeeEngine(cfg)

	// VIP 0: volume = $50,000 => Maker 10 bps, Taker 20 bps
	resVIP0 := engine.CalculateFees("EXEC-1", "USER-1", "BTCUSDT", AssetCryptoSpot, TradeBuy, 1.0, 50000.0, RoleTaker, 50000.0, false)
	expectedBrokerage := 50000.0 * (20.0 / 10000.0) // $100
	if math.Abs(resVIP0.NetBrokerage-expectedBrokerage) > 0.01 {
		t.Errorf("Expected VIP0 brokerage %.2f, got %.2f", expectedBrokerage, resVIP0.NetBrokerage)
	}

	// VIP 0 with utility token discount (25% off)
	resVIP0Discount := engine.CalculateFees("EXEC-2", "USER-1", "BTCUSDT", AssetCryptoSpot, TradeBuy, 1.0, 50000.0, RoleTaker, 50000.0, true)
	expectedDiscount := expectedBrokerage * 0.25
	expectedNet := expectedBrokerage - expectedDiscount
	if math.Abs(resVIP0Discount.NetBrokerage-expectedNet) > 0.01 {
		t.Errorf("Expected discounted net brokerage %.2f, got %.2f", expectedNet, resVIP0Discount.NetBrokerage)
	}

	// SGF Allocation: 25% of net brokerage
	expectedSGF := expectedNet * 0.25
	if math.Abs(resVIP0Discount.SGFContribution-expectedSGF) > 0.01 {
		t.Errorf("Expected SGF contribution %.2f, got %.2f", expectedSGF, resVIP0Discount.SGFContribution)
	}

	// VIP 4: volume = $25,000,000 => Maker 2 bps
	resVIP4 := engine.CalculateFees("EXEC-3", "USER-VIP4", "BTCUSDT", AssetCryptoSpot, TradeBuy, 1.0, 50000.0, RoleMaker, 25_000_000.0, false)
	expectedVIP4Brokerage := 50000.0 * (2.0 / 10000.0) // $10
	if math.Abs(resVIP4.NetBrokerage-expectedVIP4Brokerage) > 0.01 {
		t.Errorf("Expected VIP4 brokerage %.2f, got %.2f", expectedVIP4Brokerage, resVIP4.NetBrokerage)
	}
}

func TestStatutoryLeviesCalculations(t *testing.T) {
	cfg := DefaultFeeScheduleConfig()
	engine := NewComprehensiveFeeEngine(cfg)

	// Test Equity Delivery Buy: 100 shares @ ₹500 = ₹50,000 notional
	// STT: 0.1% = ₹50
	// Stamp duty: 0.015% = ₹7.5
	resDeliveryBuy := engine.CalculateFees("EXEC-EQ-1", "USER-EQ", "RELIANCE", AssetEquityDelivery, TradeBuy, 100, 500.0, RoleTaker, 0, false)
	if math.Abs(resDeliveryBuy.STT_CTT-50.0) > 0.01 {
		t.Errorf("Expected STT ₹50.0, got %.2f", resDeliveryBuy.STT_CTT)
	}
	if math.Abs(resDeliveryBuy.StampDuty-7.5) > 0.01 {
		t.Errorf("Expected stamp duty ₹7.5, got %.2f", resDeliveryBuy.StampDuty)
	}

	// Test Futures Sell: STT is 0.0125% on sell side only
	resFuturesSell := engine.CalculateFees("EXEC-FUT-1", "USER-EQ", "NIFTY-FUT", AssetFutures, TradeSell, 50, 25000.0, RoleTaker, 0, false)
	notionalFut := 50.0 * 25000.0 // 1,250,000
	expectedFutSTT := notionalFut * 0.000125
	if math.Abs(resFuturesSell.STT_CTT-expectedFutSTT) > 0.01 {
		t.Errorf("Expected Futures STT %.2f, got %.2f", expectedFutSTT, resFuturesSell.STT_CTT)
	}
}

func TestRealizedPnL_FIFO_MultiLot(t *testing.T) {
	pnlEngine := NewRealizedPnLEngine(MethodFIFO)
	userID := "TRADER-FIFO"
	symbol := "INFY"
	now := time.Now()

	// Lot 1: Buy 100 shares @ ₹1000, fee = ₹10 (opened 400 days ago -> long term)
	time1 := now.Add(-400 * 24 * time.Hour)
	_, err := pnlEngine.ProcessTradeExecution("T-1", userID, symbol, TradeBuy, 100, 1000.0, 10.0, time1)
	if err != nil {
		t.Fatalf("Lot 1 error: %v", err)
	}

	// Lot 2: Buy 100 shares @ ₹1200, fee = ₹12 (opened 30 days ago -> short term)
	time2 := now.Add(-30 * 24 * time.Hour)
	_, err = pnlEngine.ProcessTradeExecution("T-2", userID, symbol, TradeBuy, 100, 1200.0, 12.0, time2)
	if err != nil {
		t.Fatalf("Lot 2 error: %v", err)
	}

	// Sell 150 shares @ ₹1500, fee = ₹25
	// Under FIFO:
	// - 100 shares from Lot 1 @ ₹1000: Gross gain = (1500 - 1000) * 100 = ₹50,000, entry fee = ₹10, IsLongTerm = true
	// - 50 shares from Lot 2 @ ₹1200: Gross gain = (1500 - 1200) * 50 = ₹15,000, entry fee = (50/100)*12 = ₹6, IsLongTerm = false
	// Total Gross PnL = 50,000 + 15,000 = ₹65,000
	// Total Entry Fees = 10 + 6 = ₹16
	// Exit Fee = ₹25
	// Total Fees = ₹41
	// Net Realized PnL = 65,000 - 41 = ₹64,959
	event, err := pnlEngine.ProcessTradeExecution("T-3", userID, symbol, TradeSell, 150, 1500.0, 25.0, now)
	if err != nil {
		t.Fatalf("Sell trade error: %v", err)
	}

	if event == nil {
		t.Fatal("Expected realized PnL event, got nil")
	}

	if math.Abs(event.GrossRealizedPnL-65000.0) > 0.01 {
		t.Errorf("Expected gross PnL 65000, got %.2f", event.GrossRealizedPnL)
	}
	if math.Abs(event.TotalFees-41.0) > 0.01 {
		t.Errorf("Expected total fees 41, got %.2f", event.TotalFees)
	}
	if math.Abs(event.NetRealizedPnL-64959.0) > 0.01 {
		t.Errorf("Expected net PnL 64959, got %.2f", event.NetRealizedPnL)
	}

	if len(event.ClosedLots) != 2 {
		t.Fatalf("Expected 2 closed lots, got %d", len(event.ClosedLots))
	}
	if !event.ClosedLots[0].IsLongTerm {
		t.Error("Expected Lot 1 to be long-term (>365 days)")
	}
	if event.ClosedLots[1].IsLongTerm {
		t.Error("Expected Lot 2 to be short-term (<365 days)")
	}

	// Check remaining position: 50 shares remaining of Lot 2
	pos, err := pnlEngine.GetUserPosition(userID, symbol)
	if err != nil {
		t.Fatalf("GetUserPosition error: %v", err)
	}
	if math.Abs(pos.NetOpenQuantity-50.0) > 0.01 {
		t.Errorf("Expected 50 remaining open shares, got %.2f", pos.NetOpenQuantity)
	}
}

func TestRealizedPnL_LIFO(t *testing.T) {
	pnlEngine := NewRealizedPnLEngine(MethodLIFO)
	userID := "TRADER-LIFO"
	symbol := "TCS"
	now := time.Now()

	// Lot 1: Buy 100 shares @ ₹3000
	_, _ = pnlEngine.ProcessTradeExecution("T-1", userID, symbol, TradeBuy, 100, 3000.0, 10.0, now.Add(-5*time.Hour))
	// Lot 2: Buy 100 shares @ ₹3200
	_, _ = pnlEngine.ProcessTradeExecution("T-2", userID, symbol, TradeBuy, 100, 3200.0, 10.0, now.Add(-2*time.Hour))

	// Under LIFO: Sell 100 shares @ ₹3500 matches Lot 2 first!
	// Gross gain = (3500 - 3200) * 100 = ₹30,000
	event, err := pnlEngine.ProcessTradeExecution("T-3", userID, symbol, TradeSell, 100, 3500.0, 15.0, now)
	if err != nil {
		t.Fatalf("Sell error: %v", err)
	}

	if math.Abs(event.GrossRealizedPnL-30000.0) > 0.01 {
		t.Errorf("Expected LIFO gross PnL 30000, got %.2f", event.GrossRealizedPnL)
	}
	if event.ClosedLots[0].EntryPrice != 3200.0 {
		t.Errorf("Expected LIFO to close Lot 2 at 3200, got %.2f", event.ClosedLots[0].EntryPrice)
	}
}

func TestRealizedPnL_WAC(t *testing.T) {
	pnlEngine := NewRealizedPnLEngine(MethodWAC)
	userID := "TRADER-WAC"
	symbol := "HDFC"
	now := time.Now()

	// Lot 1: Buy 100 @ 1000
	_, _ = pnlEngine.ProcessTradeExecution("T-1", userID, symbol, TradeBuy, 100, 1000.0, 10.0, now)
	// Lot 2: Buy 100 @ 1200 => WAC is 1100
	_, _ = pnlEngine.ProcessTradeExecution("T-2", userID, symbol, TradeBuy, 100, 1200.0, 10.0, now)

	// Sell 100 @ 1300 => (1300 - 1100) * 100 = 20,000 gross gain
	event, err := pnlEngine.ProcessTradeExecution("T-3", userID, symbol, TradeSell, 100, 1300.0, 15.0, now)
	if err != nil {
		t.Fatalf("Sell error: %v", err)
	}

	if math.Abs(event.GrossRealizedPnL-20000.0) > 0.01 {
		t.Errorf("Expected WAC gross PnL 20000, got %.2f", event.GrossRealizedPnL)
	}
}

func TestRealizedPnL_ShortPositionAndReversal(t *testing.T) {
	pnlEngine := NewRealizedPnLEngine(MethodFIFO)
	userID := "TRADER-SHORT"
	symbol := "BTCUSDT"
	now := time.Now()

	// Short sell 2 BTC @ $65,000
	_, _ = pnlEngine.ProcessTradeExecution("T-SHORT", userID, symbol, TradeSell, 2.0, 65000.0, 20.0, now)

	// Buy 3 BTC @ $60,000 (Cover 2 BTC short at $5,000 profit each, and flip to 1 BTC Long)
	event, err := pnlEngine.ProcessTradeExecution("T-COVER-FLIP", userID, symbol, TradeBuy, 3.0, 60000.0, 30.0, now)
	if err != nil {
		t.Fatalf("Cover trade error: %v", err)
	}

	// Short profit on 2 BTC: (65000 - 60000) * 2 = $10,000
	if math.Abs(event.GrossRealizedPnL-10000.0) > 0.01 {
		t.Errorf("Expected short gross PnL 10000, got %.2f", event.GrossRealizedPnL)
	}

	// Verify position flipped to Long 1 BTC @ $60,000
	pos, err := pnlEngine.GetUserPosition(userID, symbol)
	if err != nil {
		t.Fatalf("GetUserPosition error: %v", err)
	}
	if pos.ActiveSide != TradeBuy {
		t.Errorf("Expected flipped side BUY, got %s", pos.ActiveSide)
	}
	if math.Abs(pos.NetOpenQuantity-1.0) > 0.01 {
		t.Errorf("Expected 1.0 open BTC long, got %.2f", pos.NetOpenQuantity)
	}
}
