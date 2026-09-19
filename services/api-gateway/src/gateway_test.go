package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIGateway_Healthz(t *testing.T) {
	gw := NewAPIGateway([]byte("super-secret-key-32bytes-growww!"), "http://demo-matching:8080", "http://real-matching:8080")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	gw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["status"] != "HEALTHY" {
		t.Errorf("expected HEALTHY status, got %v", body["status"])
	}
}

func TestAPIGateway_JWTAuthentication(t *testing.T) {
	secret := []byte("super-secret-key-32bytes-growww!")
	gw := NewAPIGateway(secret, "http://demo-matching:8080", "http://real-matching:8080")

	claims := Claims{
		Subject:   "usr_12345",
		Roles:     []string{"RETAIL_TRADER"},
		KYCTier:   2,
		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
	}

	token, err := gw.GenerateJWT(claims)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	// 1. Request with valid token to protected route
	req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/live", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK with valid JWT, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Request without token
	reqNoAuth := httptest.NewRequest(http.MethodGet, "/api/v1/orders/live", nil)
	wNoAuth := httptest.NewRecorder()
	gw.ServeHTTP(wNoAuth, reqNoAuth)

	if wNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized, got %d", wNoAuth.Code)
	}

	// 3. Request with expired token
	expiredClaims := Claims{
		Subject:   "usr_expired",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}
	expiredToken, _ := gw.GenerateJWT(expiredClaims)
	reqExpired := httptest.NewRequest(http.MethodGet, "/api/v1/orders/live", nil)
	reqExpired.Header.Set("Authorization", "Bearer "+expiredToken)
	wExpired := httptest.NewRecorder()
	gw.ServeHTTP(wExpired, reqExpired)

	if wExpired.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", wExpired.Code)
	}
}

func TestAPIGateway_EnvironmentRouting(t *testing.T) {
	secret := []byte("secret")
	gw := NewAPIGateway(secret, "http://demo-matching:8080", "http://real-matching:8080")

	// Demo environment default
	reqDemo := httptest.NewRequest(http.MethodGet, "/api/v1/faucet/claim", nil)
	wDemo := httptest.NewRecorder()
	gw.ServeHTTP(wDemo, reqDemo)

	var respDemo map[string]interface{}
	json.Unmarshal(wDemo.Body.Bytes(), &respDemo)
	if respDemo["environment"] != string(EnvDemoTestnet) {
		t.Errorf("expected DEMO_TESTNET, got %v", respDemo["environment"])
	}

	// Real mainnet environment with header
	reqReal := httptest.NewRequest(http.MethodGet, "/api/v1/compliance/verify", nil)
	reqReal.Header.Set("X-Trading-Mode", "REAL")
	wReal := httptest.NewRecorder()
	gw.ServeHTTP(wReal, reqReal)

	var respReal map[string]interface{}
	json.Unmarshal(wReal.Body.Bytes(), &respReal)
	if respReal["environment"] != string(EnvRealMainnet) {
		t.Errorf("expected REAL_MAINNET, got %v", respReal["environment"])
	}
}

func TestAPIGateway_RateLimiting(t *testing.T) {
	gw := NewAPIGateway([]byte("secret"), "http://demo:8080", "http://real:8080")
	gw.RateLimiter = NewRateLimiter(2.0, 0.0) // Capacity 2, refill 0 (allows exactly 2 requests)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/compliance/check", nil)
	req.RemoteAddr = "192.168.1.1:12345"

	// Req 1: OK
	w1 := httptest.NewRecorder()
	gw.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Errorf("req 1 failed: %d", w1.Code)
	}

	// Req 2: OK
	w2 := httptest.NewRecorder()
	gw.ServeHTTP(w2, req)
	if w2.Code != http.StatusOK {
		t.Errorf("req 2 failed: %d", w2.Code)
	}

	// Req 3: Rate Limited
	w3 := httptest.NewRecorder()
	gw.ServeHTTP(w3, req)
	if w3.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", w3.Code)
	}
}

func TestAPIGateway_APIKeyAuthorization(t *testing.T) {
	gw := NewAPIGateway([]byte("secret"), "http://demo:8080", "http://real:8080")
	rawKey := "growww_live_ak_987654321"

	gw.Authorizer.RegisterKey(rawKey, APIKeyMetadata{
		UserID:      "usr_inst_1",
		Permissions: PermRead | PermTrade,
		IPWhitelist: []string{"10.0.0.0/24"},
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Active:      true,
	})

	// Valid IP from CIDR
	req := httptest.NewRequest(http.MethodGet, "/api/v1/matching/depth", nil)
	req.Header.Set("X-API-Key", rawKey)
	req.RemoteAddr = "10.0.0.42:54321"
	w := httptest.NewRecorder()
	gw.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with whitelisted IP, got %d: %s", w.Code, w.Body.String())
	}

	// Invalid IP
	reqBlocked := httptest.NewRequest(http.MethodGet, "/api/v1/matching/depth", nil)
	reqBlocked.Header.Set("X-API-Key", rawKey)
	reqBlocked.RemoteAddr = "192.168.1.50:54321"
	wBlocked := httptest.NewRecorder()
	gw.ServeHTTP(wBlocked, reqBlocked)
	if wBlocked.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-whitelisted IP, got %d", wBlocked.Code)
	}
}
