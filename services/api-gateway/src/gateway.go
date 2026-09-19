package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Claims represents standard decoded JWT claims
type Claims struct {
	Subject   string   `json:"sub"`
	Roles     []string `json:"roles"`
	KYCTier   int      `json:"kyc_tier"`
	ExpiresAt int64    `json:"exp"`
}

// TokenBucket implements an in-memory token bucket rate limiter
type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	tb.tokens += elapsed * tb.refillRate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

// RateLimiter tracks per-client token buckets
type RateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*TokenBucket
	cap     float64
	rate    float64
}

func NewRateLimiter(capacity, refillRate float64) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
		cap:     capacity,
		rate:    refillRate,
	}
}

func (rl *RateLimiter) Allow(clientKey string) bool {
	rl.mu.Lock()
	tb, exists := rl.buckets[clientKey]
	if !exists {
		tb = NewTokenBucket(rl.cap, rl.rate)
		rl.buckets[clientKey] = tb
	}
	rl.mu.Unlock()

	return tb.Allow()
}

// APIGateway orchestrates routing, rate limiting, authentication, and security headers
type APIGateway struct {
	JWTSecret    []byte
	RateLimiter  *RateLimiter
	Authorizer   *APIKeyAuthorizer
	EnvRouter    *EnvironmentRouter
	MeshConfig   *ZeroTrustMeshConfig
	Routes       map[string]string // prefix -> target upstream
}

func NewAPIGateway(jwtSecret []byte, demoHost, realHost string) *APIGateway {
	return &APIGateway{
		JWTSecret:   jwtSecret,
		RateLimiter: NewRateLimiter(100.0, 50.0), // 100 capacity, 50 req/s refill
		Authorizer:  NewAPIKeyAuthorizer(),
		EnvRouter:   NewEnvironmentRouter(demoHost, realHost),
		MeshConfig:  NewZeroTrustMeshConfig(),
		Routes: map[string]string{
			"/api/v1/orders":     "order-service",
			"/api/v1/matching":   "order-matching-engine",
			"/api/v1/compliance": "compliance-service",
			"/api/v1/tax":        "tax-reporting-service",
			"/api/v1/portfolio":  "portfolio-holdings-service",
			"/api/v1/faucet":     "demo-trading-faucet-wallet-service",
		},
	}
}

// ValidateJWT verifies HMAC-SHA256 signature and unmarshals claims
func (gw *APIGateway) ValidateJWT(tokenStr string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed JWT token")
	}

	headerPayload := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, gw.JWTSecret)
	mac.Write([]byte(headerPayload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if expectedSig != parts[2] {
		return nil, errors.New("invalid signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("payload decode error: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("claims parse error: %w", err)
	}

	if claims.ExpiresAt > 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}

// GenerateJWT creates a test or internal HMAC-SHA256 JWT
func (gw *APIGateway) GenerateJWT(claims Claims) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	hp := header + "." + payload
	mac := hmac.New(sha256.New, gw.JWTSecret)
	mac.Write([]byte(hp))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return hp + "." + sig, nil
}

// ServeHTTP handles the primary HTTP dispatch
func (gw *APIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 1. CORS Headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-Trading-Mode")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. Health check
	if r.URL.Path == "/healthz" || r.URL.Path == "/health" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"HEALTHY","service":"api-gateway","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
		return
	}

	// 3. Rate Limiting by IP or API Key
	clientKey := r.RemoteAddr
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		clientKey = "key:" + apiKey
	}
	if !gw.RateLimiter.Allow(clientKey) {
		http.Error(w, `{"error":"Too Many Requests: Rate limit exceeded"}`, http.StatusTooManyRequests)
		return
	}

	// 4. API Key Authorization if provided
	if rawKey := r.Header.Get("X-API-Key"); rawKey != "" {
		clientIP := strings.Split(r.RemoteAddr, ":")[0]
		if err := gw.Authorizer.AuthorizeRequest(rawKey, clientIP, PermRead); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusForbidden)
			return
		}
	}

	// 5. JWT Auth for protected routes
	if strings.HasPrefix(r.URL.Path, "/api/v1/orders") || strings.HasPrefix(r.URL.Path, "/api/v1/portfolio") {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Unauthorized: Missing or invalid Bearer token"}`, http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := gw.ValidateJWT(tokenStr)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Unauthorized: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), "claims", claims))
	}

	// 6. Upstream routing resolution
	var targetService string
	for prefix, service := range gw.Routes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			targetService = service
			break
		}
	}

	if targetService == "" {
		http.Error(w, `{"error":"Not Found: No route matched"}`, http.StatusNotFound)
		return
	}

	// 7. Trading Mode Isolation Check
	env := gw.EnvRouter.DetermineEnvironment(r)

	// Inject Zero-Trust Mesh security headers
	r.Header.Set("X-Growww-Caller", "api-gateway")
	r.Header.Set("X-Trading-Environment", string(env))

	// Respond with route dispatch receipt
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := map[string]interface{}{
		"status":      "DISPATCHED",
		"target":      targetService,
		"environment": env,
		"path":        r.URL.Path,
		"method":      r.Method,
	}
	json.NewEncoder(w).Encode(resp)
}
