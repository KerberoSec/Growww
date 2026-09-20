package main

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type JWTClaims struct {
	Issuer      string   `json:"iss"`
	Subject     string   `json:"sub"` // User UUID
	Audience    string   `json:"aud"`
	ExpiresAt   int64    `json:"exp"`
	IssuedAt    int64    `json:"iat"`
	JTI         string   `json:"jti"`
	Entity      string   `json:"entity"`   // DOMESTIC, GIFT_CITY
	KYCTier     string   `json:"kyc_tier"` // TIER_1, TIER_2
	Roles       []string `json:"roles"`
	DeviceID    string   `json:"device_id"`
	MFAVerified bool     `json:"mfa_verified"`
}

type AuthorizationRequest struct {
	Path            string
	Method          string
	RequestedEntity string
	Claims          *JWTClaims
}

type IAMAuthorizer struct {
	mu              sync.RWMutex
	blacklistedJTIs map[string]int64 // jti -> expiresAt
}

func NewIAMAuthorizer() *IAMAuthorizer {
	return &IAMAuthorizer{
		blacklistedJTIs: make(map[string]int64),
	}
}

// BlacklistToken revokes a token session immediately (e.g. on logout/suspension)
func (a *IAMAuthorizer) BlacklistToken(jti string, expiresAt int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.blacklistedJTIs[jti] = expiresAt
}

// IsTokenRevoked checks whether JTI has been blacklisted
func (a *IAMAuthorizer) IsTokenRevoked(jti string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	exp, exists := a.blacklistedJTIs[jti]
	if !exists {
		return false
	}
	return time.Now().Unix() <= exp
}

// Authorize evaluates RBAC and ABAC policy rules against request
func (a *IAMAuthorizer) Authorize(req AuthorizationRequest, now time.Time) (bool, error) {
	// 1. Public route bypass
	if req.Path == "/healthz" || req.Path == "/api/v1/public/market-data/ticker" {
		return true, nil
	}

	// 2. Token presence
	if req.Claims == nil {
		return false, errors.New("unauthenticated: missing claims")
	}

	// 3. Token expiration check
	if now.Unix() > req.Claims.ExpiresAt {
		return false, errors.New("unauthenticated: token expired")
	}

	// 4. Token blacklist / revocation check
	if a.IsTokenRevoked(req.Claims.JTI) {
		return false, errors.New("unauthenticated: token has been revoked")
	}

	// 5. Policy evaluation for Trading Orders
	if strings.HasPrefix(req.Path, "/api/v1/orders") {
		if req.Claims.KYCTier != "TIER_2" {
			return false, errors.New("permission denied: TIER_2 KYC required for trading")
		}
		if !hasRole(req.Claims.Roles, "TRADER") {
			return false, errors.New("permission denied: TRADER role required")
		}
		if !req.Claims.MFAVerified {
			return false, errors.New("permission denied: MFA verification required")
		}
		if req.RequestedEntity != "" && req.Claims.Entity != req.RequestedEntity {
			return false, errors.New("permission denied: cross-entity jurisdiction mismatch")
		}
		return true, nil
	}

	// 6. Policy evaluation for Compliance Backoffice
	if strings.HasPrefix(req.Path, "/api/v1/compliance") {
		if !hasRole(req.Claims.Roles, "COMPLIANCE_OFFICER") {
			return false, errors.New("permission denied: COMPLIANCE_OFFICER role required")
		}
		if !req.Claims.MFAVerified {
			return false, errors.New("permission denied: MFA verification required for compliance portal")
		}
		return true, nil
	}

	// 7. Policy evaluation for Super Admin
	if strings.HasPrefix(req.Path, "/api/v1/admin") {
		if !hasRole(req.Claims.Roles, "SUPER_ADMIN") {
			return false, errors.New("permission denied: SUPER_ADMIN role required")
		}
		if !req.Claims.MFAVerified {
			return false, errors.New("permission denied: MFA verification required for admin operations")
		}
		return true, nil
	}

	// Default fallback: general authenticated investor route
	if hasRole(req.Claims.Roles, "INVESTOR") || hasRole(req.Claims.Roles, "TRADER") {
		return true, nil
	}

	return false, errors.New("permission denied: unauthorized route access")
}

func hasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}
