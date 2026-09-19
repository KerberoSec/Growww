package main

import (
	"testing"
	"time"
)

func TestAutoCreditNewUser(t *testing.T) {
	svc := NewDemoFaucetService()
	userID := "trader-100"
	clientIP := "192.168.1.10"

	bal, err := svc.AutoCreditNewUser(userID, clientIP)
	if err != nil {
		t.Fatalf("expected auto credit to succeed, got: %v", err)
	}

	if bal.VirtualUsdtE6 != 10_000*1_000_000 {
		t.Errorf("expected 10,000 USDT (10,000,000,000 e6), got %d", bal.VirtualUsdtE6)
	}
	if bal.VirtualBtcE8 != 1*100_000_000 {
		t.Errorf("expected 1 BTC (100,000,000 e8), got %d", bal.VirtualBtcE8)
	}
	if bal.ClaimsCount != 1 {
		t.Errorf("expected ClaimsCount = 1, got %d", bal.ClaimsCount)
	}

	// Re-registering same user should fail
	_, err = svc.AutoCreditNewUser(userID, clientIP)
	if err == nil {
		t.Fatalf("expected error on duplicate auto-credit, got nil")
	}
}

func TestCooldownEnforcement(t *testing.T) {
	svc := NewDemoFaucetService()
	userID := "trader-200"
	clientIP := "192.168.1.20"

	// Credit initial funds
	_, err := svc.AutoCreditNewUser(userID, clientIP)
	if err != nil {
		t.Fatalf("auto-credit failed: %v", err)
	}

	// Attempt periodic claim immediately -> should fail due to 24h cooldown
	_, err = svc.ClaimPeriodicFaucet(userID, clientIP)
	if err == nil {
		t.Fatalf("expected cooldown error, got nil")
	}

	// Check eligibility status
	eligibility, err := svc.GetEligibility(userID, clientIP)
	if err != nil {
		t.Fatalf("get eligibility error: %v", err)
	}
	if eligibility.Eligible {
		t.Errorf("expected user not to be eligible during cooldown")
	}
	if eligibility.RemainingTime <= 0 {
		t.Errorf("expected positive remaining cooldown time")
	}

	// Simulate cooldown expiration
	svc.SetCooldown(10 * time.Millisecond)
	time.Sleep(15 * time.Millisecond)

	// Now periodic claim should succeed
	bal, err := svc.ClaimPeriodicFaucet(userID, clientIP)
	if err != nil {
		t.Fatalf("expected periodic claim to succeed after cooldown, got: %v", err)
	}
	if bal.ClaimsCount != 2 {
		t.Errorf("expected ClaimsCount = 2, got %d", bal.ClaimsCount)
	}
	if bal.VirtualUsdtE6 != 20_000*1_000_000 {
		t.Errorf("expected 20,000 USDT, got %d", bal.VirtualUsdtE6)
	}
	if bal.VirtualBtcE8 != 2*100_000_000 {
		t.Errorf("expected 2 BTC, got %d", bal.VirtualBtcE8)
	}
}

func TestBalanceThresholdRequirement(t *testing.T) {
	svc := NewDemoFaucetService()
	svc.SetCooldown(1 * time.Millisecond)
	svc.SetEnforceThreshold(true)

	userID := "trader-threshold"
	clientIP := "192.168.1.30"

	_, err := svc.AutoCreditNewUser(userID, clientIP)
	if err != nil {
		t.Fatalf("auto credit failed: %v", err)
	}

	time.Sleep(5 * time.Millisecond)

	// User currently has 10,000 USDT > 1,000 USDT threshold -> should be rejected
	_, err = svc.ClaimPeriodicFaucet(userID, clientIP)
	if err == nil {
		t.Fatalf("expected rejection when balance above threshold")
	}

	// Simulate user losing/trading down balance to 500 USDT
	acc, _ := svc.ledger.GetAccount(userID)
	acc.Balances["USDT"].Available = 500 * 1_000_000

	// Now periodic claim should succeed
	bal, err := svc.ClaimPeriodicFaucet(userID, clientIP)
	if err != nil {
		t.Fatalf("expected claim to succeed when balance below threshold, got: %v", err)
	}
	if bal.VirtualUsdtE6 != (500+10_000)*1_000_000 {
		t.Errorf("unexpected updated balance: %d", bal.VirtualUsdtE6)
	}
}

func TestPortfolioReset(t *testing.T) {
	svc := NewDemoFaucetService()
	userID := "trader-reset"
	clientIP := "192.168.1.40"

	_, err := svc.AutoCreditNewUser(userID, clientIP)
	if err != nil {
		t.Fatalf("auto credit failed: %v", err)
	}

	// Modify balances and place a hold
	_ = svc.ledger.DebitBalance(userID, "USDT", 5_000*1_000_000, "SIMULATED_LOSS")
	_, err = svc.ledger.HoldBalance(userID, "BTC", 50_000_000, "order-open-999")
	if err != nil {
		t.Fatalf("hold failed: %v", err)
	}

	// Reset portfolio
	bal, err := svc.ResetVirtualPortfolio(userID)
	if err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	if bal.VirtualUsdtE6 != DefaultInitialUsdtE6 {
		t.Errorf("expected USDT reset to %d, got %d", DefaultInitialUsdtE6, bal.VirtualUsdtE6)
	}
	if bal.VirtualBtcE8 != DefaultInitialBtcE8 {
		t.Errorf("expected BTC reset to %d, got %d", DefaultInitialBtcE8, bal.VirtualBtcE8)
	}
	if bal.LockedBtcE8 != 0 {
		t.Errorf("expected locked BTC to be 0 after reset, got %d", bal.LockedBtcE8)
	}
}

func TestOrderHoldAndRelease(t *testing.T) {
	svc := NewDemoFaucetService()
	userID := "trader-hold"
	clientIP := "192.168.1.50"

	_, err := svc.AutoCreditNewUser(userID, clientIP)
	if err != nil {
		t.Fatalf("auto credit failed: %v", err)
	}

	// Hold 3,000 USDT for limit buy order
	holdAmount := uint64(3_000 * 1_000_000)
	orderID := "order-limit-101"
	hold, err := svc.ledger.HoldBalance(userID, "USDT", holdAmount, orderID)
	if err != nil {
		t.Fatalf("hold balance failed: %v", err)
	}
	if hold.Amount != holdAmount {
		t.Errorf("expected hold amount %d, got %d", holdAmount, hold.Amount)
	}

	// Verify available dropped to 7,000 and locked is 3,000
	bal, _ := svc.GetBalance(userID)
	if bal.VirtualUsdtE6 != 7_000*1_000_000 {
		t.Errorf("expected available 7,000 USDT, got %d", bal.VirtualUsdtE6)
	}
	if bal.LockedUsdtE6 != 3_000*1_000_000 {
		t.Errorf("expected locked 3,000 USDT, got %d", bal.LockedUsdtE6)
	}

	// Try holding more than available
	_, err = svc.ledger.HoldBalance(userID, "USDT", 10_000*1_000_000, "order-too-large")
	if err == nil {
		t.Fatalf("expected hold to fail due to insufficient funds")
	}

	// Release hold
	err = svc.ledger.ReleaseBalance(userID, orderID)
	if err != nil {
		t.Fatalf("release balance failed: %v", err)
	}

	bal, _ = svc.GetBalance(userID)
	if bal.VirtualUsdtE6 != 10_000*1_000_000 {
		t.Errorf("expected available restored to 10,000 USDT, got %d", bal.VirtualUsdtE6)
	}
	if bal.LockedUsdtE6 != 0 {
		t.Errorf("expected locked to be 0, got %d", bal.LockedUsdtE6)
	}
}

func TestAtomicTradeSettlement(t *testing.T) {
	svc := NewDemoFaucetService()
	buyerID := "buyer-sam"
	sellerID := "seller-bob"

	_, err := svc.AutoCreditNewUser(buyerID, "10.0.0.1")
	if err != nil {
		t.Fatalf("buyer setup failed: %v", err)
	}
	_, err = svc.AutoCreditNewUser(sellerID, "10.0.0.2")
	if err != nil {
		t.Fatalf("seller setup failed: %v", err)
	}

	// Trade: Buyer buys 0.5 BTC at 60,000 USDT/BTC = 30,000 USDT
	// Give buyer sufficient USDT first
	_ = svc.ledger.CreditBalance(buyerID, "USDT", 25_000*1_000_000, "EXTRA_DEPOSIT")

	tradeCostUsdt := uint64(30_000 * 1_000_000)
	tradeQtyBtc := uint64(50_000_000) // 0.5 BTC

	// Buyer locks 30,000 USDT
	_, err = svc.ledger.HoldBalance(buyerID, "USDT", tradeCostUsdt, "buy-order-1")
	if err != nil {
		t.Fatalf("buyer hold failed: %v", err)
	}

	// Seller locks 0.5 BTC
	_, err = svc.ledger.HoldBalance(sellerID, "BTC", tradeQtyBtc, "sell-order-1")
	if err != nil {
		t.Fatalf("seller hold failed: %v", err)
	}

	// Execute settlement
	err = svc.ledger.SettleTrade(buyerID, sellerID, "BTC", tradeQtyBtc, "USDT", tradeCostUsdt, "trade-match-001")
	if err != nil {
		t.Fatalf("settlement failed: %v", err)
	}

	// Verify Buyer: gained 0.5 BTC, paid 30,000 USDT
	buyerBal, _ := svc.GetBalance(buyerID)
	if buyerBal.VirtualBtcE8 != 150_000_000 { // 1.0 initial + 0.5 = 1.5 BTC
		t.Errorf("expected buyer BTC 150,000,000, got %d", buyerBal.VirtualBtcE8)
	}
	if buyerBal.LockedUsdtE6 != 0 {
		t.Errorf("expected buyer locked USDT 0, got %d", buyerBal.LockedUsdtE6)
	}

	// Verify Seller: gained 30,000 USDT, delivered 0.5 BTC
	sellerBal, _ := svc.GetBalance(sellerID)
	if sellerBal.VirtualUsdtE6 != 40_000*1_000_000 { // 10,000 initial + 30,000 = 40,000 USDT
		t.Errorf("expected seller USDT 40,000,000,000, got %d", sellerBal.VirtualUsdtE6)
	}
	if sellerBal.VirtualBtcE8 != 50_000_000 { // 1.0 initial - 0.5 = 0.5 BTC
		t.Errorf("expected seller BTC 50,000,000, got %d", sellerBal.VirtualBtcE8)
	}
}

func TestIPRateLimiting(t *testing.T) {
	svc := NewDemoFaucetService()
	ip := "198.51.100.25"

	// Register 5 different users from the same IP (max limit = 5)
	for i := 0; i < MaxClaimsPerIPWindow; i++ {
		uid := string(rune('A' + i))
		_, err := svc.AutoCreditNewUser(uid, ip)
		if err != nil {
			t.Fatalf("user %s should succeed within limit, got: %v", uid, err)
		}
	}

	// 6th claim from same IP must be rate limited
	_, err := svc.AutoCreditNewUser("user-overflow", ip)
	if err == nil {
		t.Fatalf("expected rate limit error on 6th request from same IP")
	}
}
