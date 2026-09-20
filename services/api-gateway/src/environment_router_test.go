package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnvironmentRouter_DetermineEnvironment(t *testing.T) {
	router := NewEnvironmentRouter("http://demo-matching:8080", "http://real-matching:8080")

	tests := []struct {
		name        string
		tradingMode string
		environment string
		expectedEnv EnvironmentType
	}{
		{"Default without headers", "", "", EnvDemoTestnet},
		{"Explicit REAL trading mode", "REAL", "", EnvRealMainnet},
		{"Explicit MAINNET trading mode", "MAINNET", "", EnvRealMainnet},
		{"Explicit DEMO trading mode", "DEMO", "", EnvDemoTestnet},
		{"X-Environment REAL header", "", "REAL", EnvRealMainnet},
		{"X-Environment TESTNET header", "", "TESTNET", EnvDemoTestnet},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/orders", nil)
			if tc.tradingMode != "" {
				req.Header.Set("X-Trading-Mode", tc.tradingMode)
			}
			if tc.environment != "" {
				req.Header.Set("X-Environment", tc.environment)
			}

			env := router.DetermineEnvironment(req)
			if env != tc.expectedEnv {
				t.Fatalf("expected env %s, got %s", tc.expectedEnv, env)
			}
		})
	}
}

func TestEnvironmentRouter_RouteTargetAndConfig(t *testing.T) {
	router := NewEnvironmentRouter("http://demo-matching:8080", "http://real-matching:8080")

	// Demo Config
	demoTarget, err := router.RouteTarget(EnvDemoTestnet)
	if err != nil || demoTarget != "http://demo-matching:8080" {
		t.Fatalf("unexpected demo target: %v, err: %v", demoTarget, err)
	}

	demoCfg, err := router.GetConfig(EnvDemoTestnet)
	if err != nil || demoCfg.ChainID != 1337 || demoCfg.KafkaTopicPrefix != "demo." {
		t.Fatalf("unexpected demo config: %+v, err: %v", demoCfg, err)
	}

	// Real Config
	realTarget, err := router.RouteTarget(EnvRealMainnet)
	if err != nil || realTarget != "http://real-matching:8080" {
		t.Fatalf("unexpected real target: %v, err: %v", realTarget, err)
	}

	realCfg, err := router.GetConfig(EnvRealMainnet)
	if err != nil || realCfg.ChainID != 2026 || realCfg.KafkaTopicPrefix != "mainnet." {
		t.Fatalf("unexpected real config: %+v, err: %v", realCfg, err)
	}
}

func TestEnvironmentRouter_ValidateChainID(t *testing.T) {
	router := NewEnvironmentRouter("http://demo-matching:8080", "http://real-matching:8080")

	// Correct chain IDs
	if err := router.ValidateChainID(EnvDemoTestnet, 1337); err != nil {
		t.Fatalf("expected valid testnet chain ID, got %v", err)
	}
	if err := router.ValidateChainID(EnvRealMainnet, 2026); err != nil {
		t.Fatalf("expected valid mainnet chain ID, got %v", err)
	}

	// Cross-contamination chain IDs
	if err := router.ValidateChainID(EnvDemoTestnet, 2026); err == nil {
		t.Fatal("expected error passing mainnet chain ID to demo environment")
	}
	if err := router.ValidateChainID(EnvRealMainnet, 1337); err == nil {
		t.Fatal("expected error passing testnet chain ID to real mainnet")
	}
}

func TestEnvironmentRouter_ValidateRequestBoundary(t *testing.T) {
	router := NewEnvironmentRouter("http://demo-matching:8080", "http://real-matching:8080")

	// Demo order placement with demo:trade scope: Allowed
	err := router.ValidateRequestBoundary(EnvDemoTestnet, "/v1/orders", "demo:trade")
	if err != nil {
		t.Fatalf("expected valid boundary check, got: %v", err)
	}

	// Demo trying to execute withdrawal: Blocked by cross-contamination guard
	err = router.ValidateRequestBoundary(EnvDemoTestnet, "/v1/crypto/withdraw", "demo:trade")
	if err != ErrDemoCrossContamination {
		t.Fatalf("expected ErrDemoCrossContamination, got: %v", err)
	}

	// Demo trying with only live scope: Scope mismatch
	err = router.ValidateRequestBoundary(EnvDemoTestnet, "/v1/orders", "live:trade")
	if err != ErrScopeMismatch {
		t.Fatalf("expected ErrScopeMismatch, got: %v", err)
	}

	// Live environment with live:trade scope: Allowed
	err = router.ValidateRequestBoundary(EnvRealMainnet, "/v1/orders", "live:trade")
	if err != nil {
		t.Fatalf("expected valid boundary check on live, got: %v", err)
	}

	// Live environment with only demo scope: Scope mismatch
	err = router.ValidateRequestBoundary(EnvRealMainnet, "/v1/orders", "demo:trade")
	if err != ErrScopeMismatch {
		t.Fatalf("expected ErrScopeMismatch, got: %v", err)
	}
}

func TestEnvironmentRouter_Middleware(t *testing.T) {
	router := NewEnvironmentRouter("http://demo-matching:8080", "http://real-matching:8080")

	var capturedEnv EnvironmentType
	handler := router.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedEnv = r.Context().Value(EnvironmentKey).(EnvironmentType)
		w.WriteHeader(http.StatusOK)
	}))

	// Case 1: Valid demo order request
	req1 := httptest.NewRequest(http.MethodPost, "/v1/orders", nil)
	req1.Header.Set("X-Environment", "DEMO")
	req1.Header.Set("X-Token-Scope", "demo:trade")
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec1.Code)
	}
	if capturedEnv != EnvDemoTestnet {
		t.Fatalf("expected context env DEMO_TESTNET, got %s", capturedEnv)
	}

	// Case 2: Attempted withdrawal under demo mode (should be 403 Forbidden)
	req2 := httptest.NewRequest(http.MethodPost, "/v1/fiat/payout", nil)
	req2.Header.Set("X-Environment", "DEMO")
	req2.Header.Set("X-Token-Scope", "demo:trade")
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden, got %d", rec2.Code)
	}

	// Case 3: Scope mismatch under real mode (should be 401 Unauthorized)
	req3 := httptest.NewRequest(http.MethodPost, "/v1/orders", nil)
	req3.Header.Set("X-Environment", "REAL")
	req3.Header.Set("X-Token-Scope", "demo:trade")
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req3)

	if rec3.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d", rec3.Code)
	}
}
