package main

import (
	"testing"
	"time"
)

func TestIAMAuthorizer_PublicRoutes(t *testing.T) {
	auth := NewIAMAuthorizer()

	allowed, err := auth.Authorize(AuthorizationRequest{Path: "/healthz"}, time.Now())
	if !allowed || err != nil {
		t.Errorf("expected public healthz allowed, got %v, err: %v", allowed, err)
	}

	allowed, err = auth.Authorize(AuthorizationRequest{Path: "/api/v1/public/market-data/ticker"}, time.Now())
	if !allowed || err != nil {
		t.Errorf("expected public ticker allowed, got %v, err: %v", allowed, err)
	}
}

func TestIAMAuthorizer_TradingOrderPermissions(t *testing.T) {
	auth := NewIAMAuthorizer()
	now := time.Now()

	validClaims := &JWTClaims{
		Subject:     "usr_trader_01",
		ExpiresAt:   now.Add(15 * time.Minute).Unix(),
		JTI:         "jti_test_1",
		Entity:      "DOMESTIC",
		KYCTier:     "TIER_2",
		Roles:       []string{"TRADER"},
		MFAVerified: true,
	}

	// 1. Valid trader
	allowed, err := auth.Authorize(AuthorizationRequest{
		Path:            "/api/v1/orders",
		Method:          "POST",
		RequestedEntity: "DOMESTIC",
		Claims:          validClaims,
	}, now)
	if !allowed || err != nil {
		t.Fatalf("expected valid trader order allowed: %v", err)
	}

	// 2. KYC Tier 1 trying to trade -> Rejection
	tier1Claims := *validClaims
	tier1Claims.KYCTier = "TIER_1"
	allowed, err = auth.Authorize(AuthorizationRequest{
		Path:   "/api/v1/orders",
		Claims: &tier1Claims,
	}, now)
	if allowed || err == nil {
		t.Errorf("expected rejection for TIER_1 KYC")
	}

	// 3. Entity mismatch -> Rejection
	entityMismatch := *validClaims
	entityMismatch.Entity = "GIFT_CITY"
	allowed, err = auth.Authorize(AuthorizationRequest{
		Path:            "/api/v1/orders",
		RequestedEntity: "DOMESTIC",
		Claims:          &entityMismatch,
	}, now)
	if allowed || err == nil {
		t.Errorf("expected rejection for jurisdiction mismatch")
	}

	// 4. Token expired -> Rejection
	expiredClaims := *validClaims
	expiredClaims.ExpiresAt = now.Add(-1 * time.Minute).Unix()
	allowed, err = auth.Authorize(AuthorizationRequest{
		Path:   "/api/v1/orders",
		Claims: &expiredClaims,
	}, now)
	if allowed || err == nil {
		t.Errorf("expected rejection for expired token")
	}

	// 5. Token blacklisted -> Rejection
	auth.BlacklistToken("jti_test_1", now.Add(10*time.Minute).Unix())
	allowed, err = auth.Authorize(AuthorizationRequest{
		Path:            "/api/v1/orders",
		RequestedEntity: "DOMESTIC",
		Claims:          validClaims,
	}, now)
	if allowed || err == nil {
		t.Errorf("expected rejection for blacklisted token")
	}
}

func TestIAMAuthorizer_ComplianceAndAdminRoles(t *testing.T) {
	auth := NewIAMAuthorizer()
	now := time.Now()

	compClaims := &JWTClaims{
		Subject:     "usr_compliance_01",
		ExpiresAt:   now.Add(15 * time.Minute).Unix(),
		JTI:         "jti_comp_1",
		Roles:       []string{"COMPLIANCE_OFFICER"},
		MFAVerified: true,
	}

	allowed, err := auth.Authorize(AuthorizationRequest{
		Path:   "/api/v1/compliance/audit-logs",
		Claims: compClaims,
	}, now)
	if !allowed || err != nil {
		t.Fatalf("expected compliance officer allowed: %v", err)
	}

	// Compliance officer attempting superadmin route -> Rejection
	allowed, err = auth.Authorize(AuthorizationRequest{
		Path:   "/api/v1/admin/cluster-shutdown",
		Claims: compClaims,
	}, now)
	if allowed || err == nil {
		t.Errorf("expected compliance officer rejected from admin route")
	}
}
