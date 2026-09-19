package affiliaterebate

import (
	"fmt"
	"testing"
)



func TestRegisterAffiliate(t *testing.T) {
	engine := NewRebateEngine()

	aff, err := engine.RegisterAffiliate("aff-1", "user-1", "REF123")
	if err != nil {
		t.Fatalf("RegisterAffiliate failed: %v", err)
	}
	if aff.CurrentTier != TierBronze {
		t.Errorf("expected Bronze tier, got %s", aff.CurrentTier)
	}
	if aff.ReferralCode != "REF123" {
		t.Errorf("unexpected referral code: %s", aff.ReferralCode)
	}

	// Duplicate
	_, err = engine.RegisterAffiliate("aff-1", "user-2", "REF456")
	if err == nil {
		t.Error("expected error on duplicate registration")
	}

	// Missing fields
	_, err = engine.RegisterAffiliate("", "user-3", "REF789")
	if err == nil {
		t.Error("expected error with empty ID")
	}
}

func TestTierDetermination(t *testing.T) {
	engine := NewRebateEngine()
	engine.RegisterAffiliate("aff-1", "user-1", "REF123")

	// Add 25 referrals with volume to qualify for Gold
	for i := 0; i < 25; i++ {
		uid := fmt.Sprintf("ref-user-%d", i)
		engine.AddReferral("aff-1", uid)
		engine.RecordReferralVolume("aff-1", uid, 250_000, 500) // 250K each = 6.25M total
	}

	tier, err := engine.DetermineTier("aff-1")
	if err != nil {
		t.Fatalf("DetermineTier failed: %v", err)
	}
	if tier != TierGold {
		t.Errorf("expected Gold tier (25 refs, 6.25M volume), got %s", tier)
	}
}

func TestTierSilver(t *testing.T) {
	engine := NewRebateEngine()
	engine.RegisterAffiliate("aff-2", "user-2", "REF456")

	// 5 referrals with 100K each = 500K (Silver threshold)
	for i := 0; i < 5; i++ {
		uid := fmt.Sprintf("ref-%d", i)
		engine.AddReferral("aff-2", uid)
		engine.RecordReferralVolume("aff-2", uid, 100_000, 200)
	}

	tier, _ := engine.DetermineTier("aff-2")
	if tier != TierSilver {
		t.Errorf("expected Silver tier, got %s", tier)
	}
}

func TestCalculateMonthlyRebate(t *testing.T) {
	engine := NewRebateEngine()
	engine.RegisterAffiliate("aff-1", "user-1", "REF123")
	engine.AddReferral("aff-1", "ref-1")
	engine.AddReferral("aff-1", "ref-2")

	// Bronze tier = 10% rebate
	engine.RecordReferralVolume("aff-1", "ref-1", 100_000, 1000)
	engine.RecordReferralVolume("aff-1", "ref-2", 200_000, 2000)

	rebate, err := engine.CalculateMonthlyRebate("aff-1")
	if err != nil {
		t.Fatalf("CalculateMonthlyRebate failed: %v", err)
	}
	// 10% of (1000 + 2000) = 300
	if rebate != 300.0 {
		t.Errorf("expected 300.0 rebate, got %.2f", rebate)
	}
}

func TestFlaggedAffiliateBlocked(t *testing.T) {
	engine := NewRebateEngine()
	engine.RegisterAffiliate("aff-1", "user-1", "REF123")
	engine.AddReferral("aff-1", "ref-1")
	engine.RecordReferralVolume("aff-1", "ref-1", 100_000, 1000)

	// Flag for fraud
	engine.FlagAffiliate("aff-1", "suspected wash trading")

	_, err := engine.CalculateMonthlyRebate("aff-1")
	if err == nil {
		t.Error("expected error when calculating rebate for flagged affiliate")
	}
}

func TestFraudDetectionRapidChurn(t *testing.T) {
	engine := NewRebateEngine()
	engine.RegisterAffiliate("aff-1", "user-1", "REF123")

	// Add 10 referrals, make 8 inactive (80% churn > 70% threshold)
	for i := 0; i < 10; i++ {
		uid := fmt.Sprintf("ref-%d", i)
		engine.AddReferral("aff-1", uid)
	}

	// Deactivate 8 referrals manually
	engine.mu.Lock()
	refs := engine.referrals["aff-1"]
	for i := 0; i < 8; i++ {
		refs[i].IsActive = false
	}
	engine.mu.Unlock()

	signals, err := engine.DetectFraud("aff-1")
	if err != nil {
		t.Fatalf("DetectFraud failed: %v", err)
	}

	foundChurn := false
	for _, sig := range signals {
		if sig.Type == "RAPID_CHURN" {
			foundChurn = true
			if sig.Score < 0.7 {
				t.Errorf("expected churn score >= 0.7, got %.2f", sig.Score)
			}
		}
	}
	if !foundChurn {
		t.Error("expected RAPID_CHURN signal not found")
	}
}

func TestFraudDetectionWashVolume(t *testing.T) {
	engine := NewRebateEngine()
	engine.RegisterAffiliate("aff-2", "user-2", "REF456")
	engine.AddReferral("aff-2", "ref-1")

	// Volume but zero fees → wash
	engine.RecordReferralVolume("aff-2", "ref-1", 1_000_000, 0)

	signals, err := engine.DetectFraud("aff-2")
	if err != nil {
		t.Fatalf("DetectFraud failed: %v", err)
	}

	foundWash := false
	for _, sig := range signals {
		if sig.Type == "WASH_VOLUME" {
			foundWash = true
		}
	}
	if !foundWash {
		t.Error("expected WASH_VOLUME signal not found")
	}
}

// Need fmt for test referral IDs
