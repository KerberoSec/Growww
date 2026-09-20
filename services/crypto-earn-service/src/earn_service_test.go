package src

import (
	"testing"
	"time"
)

func TestEarnService_FlexibleLifecycle(t *testing.T) {
	svc := NewEarnService()

	// 1. Create Flexible USDT product: 500 bps = 5.00% APR
	prod, err := svc.CreateProduct("prod_flex_usdt", "USDT", ProductFlexible, 500, 10_00000000, 1_000_000_00000000)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}
	if prod.DurationDays != 0 {
		t.Errorf("expected 0 duration for flexible, got %d", prod.DurationDays)
	}

	// 2. User deposits 10,000 USDT (10,000 * 1e8)
	depositAmt := int64(10_000_00000000)
	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	sub, err := svc.Subscribe("user_alice", "prod_flex_usdt", depositAmt, startTime)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}
	if sub.PrincipalE8 != depositAmt {
		t.Errorf("expected principal %d, got %d", depositAmt, sub.PrincipalE8)
	}

	// 3. Check yield accrual after 365 days
	oneYearLater := startTime.Add(365 * 24 * time.Hour)
	gross, net, reserve, err := svc.CalculateAccruedYield(sub.SubscriptionID, oneYearLater)
	if err != nil {
		t.Fatalf("yield calc error: %v", err)
	}

	// 10,000 * 5% = 500 USDT (500_00000000)
	expectedGross := int64(500_00000000)
	expectedReserve := int64(50_00000000)
	expectedNet := int64(450_00000000)

	if gross != expectedGross {
		t.Errorf("expected gross %d, got %d", expectedGross, gross)
	}
	if reserve != expectedReserve {
		t.Errorf("expected reserve %d, got %d", expectedReserve, reserve)
	}
	if net != expectedNet {
		t.Errorf("expected net %d, got %d", expectedNet, net)
	}

	// 4. Redeem
	principal, yieldPaid, err := svc.Redeem(sub.SubscriptionID, "user_alice", oneYearLater)
	if err != nil {
		t.Fatalf("failed to redeem: %v", err)
	}
	if principal != depositAmt {
		t.Errorf("expected principal %d, got %d", depositAmt, principal)
	}
	if yieldPaid != expectedNet {
		t.Errorf("expected yield %d, got %d", expectedNet, yieldPaid)
	}

	// Check reserve accumulated
	resBal := svc.GetReserveBalance("USDT")
	if resBal != expectedReserve {
		t.Errorf("expected reserve balance %d, got %d", expectedReserve, resBal)
	}
}

func TestEarnService_FixedTermMaturityAndPenalty(t *testing.T) {
	svc := NewEarnService()

	// Fixed 30-day product: 1200 bps = 12.00% APR
	_, err := svc.CreateProduct("prod_fixed_30", "BTC", ProductFixed30, 1200, 1_00000000, 100_00000000)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	startTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	depositAmt := int64(5_00000000) // 5 BTC

	// Early redemption penalty test
	subEarly, err := svc.Subscribe("user_bob", "prod_fixed_30", depositAmt, startTime)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Bob attempts early withdrawal after 15 days (before 30 days)
	day15 := startTime.Add(15 * 24 * time.Hour)
	princEarly, yieldEarly, err := svc.Redeem(subEarly.SubscriptionID, "user_bob", day15)
	if err != nil {
		t.Fatalf("failed early redeem: %v", err)
	}
	if princEarly != depositAmt {
		t.Errorf("expected principal %d, got %d", depositAmt, princEarly)
	}
	if yieldEarly != 0 {
		t.Errorf("expected 0 yield due to early penalty, got %d", yieldEarly)
	}

	// Full maturity test
	subMaturity, err := svc.Subscribe("user_charlie", "prod_fixed_30", depositAmt, startTime)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	day30 := startTime.Add(30 * 24 * time.Hour)
	princMat, yieldMat, err := svc.Redeem(subMaturity.SubscriptionID, "user_charlie", day30)
	if err != nil {
		t.Fatalf("failed maturity redeem: %v", err)
	}
	if princMat != depositAmt {
		t.Errorf("expected principal %d, got %d", depositAmt, princMat)
	}
	if yieldMat <= 0 {
		t.Errorf("expected positive yield on maturity, got %d", yieldMat)
	}
}

func TestEarnService_CapacityAndValidation(t *testing.T) {
	svc := NewEarnService()

	_, err := svc.CreateProduct("prod_cap", "USDT", ProductFlexible, 600, 100_00000000, 500_00000000)
	if err != nil {
		t.Fatalf("failed create: %v", err)
	}

	now := time.Now()
	// Deposit below min
	_, err = svc.Subscribe("user1", "prod_cap", 50_00000000, now)
	if err == nil {
		t.Errorf("expected error for deposit below min")
	}

	// Valid deposit
	_, err = svc.Subscribe("user1", "prod_cap", 400_00000000, now)
	if err != nil {
		t.Fatalf("valid deposit failed: %v", err)
	}

	// Deposit exceeding capacity (400 + 200 > 500)
	_, err = svc.Subscribe("user2", "prod_cap", 200_00000000, now)
	if err == nil {
		t.Errorf("expected error for capacity exceeded")
	}
}
