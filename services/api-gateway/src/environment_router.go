package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type EnvironmentType string

const (
	EnvDemoTestnet EnvironmentType = "DEMO_TESTNET"
	EnvRealMainnet EnvironmentType = "REAL_MAINNET"
)

const (
	TestnetChainID uint64 = 1337
	MainnetChainID uint64 = 2026
)

// EnvironmentContextKey is the context key for trading mode
type contextKey string

const (
	EnvironmentKey contextKey = "trading_environment"
	ClaimsScopeKey contextKey = "claims_scope"
)

var (
	ErrDemoCrossContamination = errors.New("cross-contamination blocked: real financial action not allowed in demo mode")
	ErrScopeMismatch          = errors.New("JWT scope insufficient for target trading environment")
	ErrInvalidChainID         = errors.New("chain ID does not match target environment isolation boundary")
	ErrUnknownEnvironment     = errors.New("unknown environment type")
)

// EnvironmentConfig encapsulates segregated parameters for each environment
type EnvironmentConfig struct {
	EnvType          EnvironmentType
	ChainID          uint64
	KafkaTopicPrefix string
	DBSchema         string
	RedisPrefix      string
	MatchingHost     string
}

// EnvironmentRouter inspects request headers, scopes, and routes to isolated backends
type EnvironmentRouter struct {
	DemoConfig EnvironmentConfig
	RealConfig EnvironmentConfig
}

func NewEnvironmentRouter(demoHost, realHost string) *EnvironmentRouter {
	return &EnvironmentRouter{
		DemoConfig: EnvironmentConfig{
			EnvType:          EnvDemoTestnet,
			ChainID:          TestnetChainID,
			KafkaTopicPrefix: "demo.",
			DBSchema:         "demo",
			RedisPrefix:      "demo:",
			MatchingHost:     demoHost,
		},
		RealConfig: EnvironmentConfig{
			EnvType:          EnvRealMainnet,
			ChainID:          MainnetChainID,
			KafkaTopicPrefix: "mainnet.",
			DBSchema:         "mainnet",
			RedisPrefix:      "mainnet:",
			MatchingHost:     realHost,
		},
	}
}

// DetermineEnvironment extracts mode from X-Trading-Mode or X-Environment headers
func (r *EnvironmentRouter) DetermineEnvironment(req *http.Request) EnvironmentType {
	mode := strings.ToUpper(req.Header.Get("X-Trading-Mode"))
	envHdr := strings.ToUpper(req.Header.Get("X-Environment"))

	if mode == "REAL" || mode == "MAINNET" || envHdr == "REAL" || envHdr == "MAINNET" {
		return EnvRealMainnet
	}
	// Default to isolated demo sandbox for user safety
	return EnvDemoTestnet
}

// ValidateRequestBoundary ensures no real-money banking / KYC cross-contamination occurs in demo mode
func (r *EnvironmentRouter) ValidateRequestBoundary(env EnvironmentType, path string, tokenScope string) error {
	cleanPath := strings.ToLower(path)

	// Restricted paths in demo mode to prevent accidental real fund movement
	if env == EnvDemoTestnet {
		if strings.Contains(cleanPath, "/withdraw") ||
			strings.Contains(cleanPath, "/fiat/payout") ||
			strings.Contains(cleanPath, "/bank-transfer") ||
			strings.Contains(cleanPath, "/kyc/verify") {
			return ErrDemoCrossContamination
		}
	}

	// Validate scopes
	scopes := strings.Split(tokenScope, " ")
	hasScope := func(target string) bool {
		for _, s := range scopes {
			if s == target || s == "admin" || s == "trading:all" {
				return true
			}
		}
		return false
	}

	if env == EnvDemoTestnet {
		if !hasScope("demo:trade") && !hasScope("demo") {
			return ErrScopeMismatch
		}
	} else {
		if !hasScope("live:trade") && !hasScope("live") {
			return ErrScopeMismatch
		}
	}

	return nil
}

// ValidateChainID verifies that blockchain transactions strictly target the isolated environment's chain
func (r *EnvironmentRouter) ValidateChainID(env EnvironmentType, chainID uint64) error {
	switch env {
	case EnvDemoTestnet:
		if chainID != TestnetChainID {
			return fmt.Errorf("%w: expected testnet %d, got %d", ErrInvalidChainID, TestnetChainID, chainID)
		}
	case EnvRealMainnet:
		if chainID != MainnetChainID {
			return fmt.Errorf("%w: expected mainnet %d, got %d", ErrInvalidChainID, MainnetChainID, chainID)
		}
	default:
		return ErrUnknownEnvironment
	}
	return nil
}

// RouteTarget returns the appropriate cluster endpoint based on isolation mode
func (r *EnvironmentRouter) RouteTarget(env EnvironmentType) (string, error) {
	switch env {
	case EnvDemoTestnet:
		return r.DemoConfig.MatchingHost, nil
	case EnvRealMainnet:
		return r.RealConfig.MatchingHost, nil
	default:
		return "", ErrUnknownEnvironment
	}
}

// GetConfig returns the full environment configuration bundle
func (r *EnvironmentRouter) GetConfig(env EnvironmentType) (*EnvironmentConfig, error) {
	switch env {
	case EnvDemoTestnet:
		return &r.DemoConfig, nil
	case EnvRealMainnet:
		return &r.RealConfig, nil
	default:
		return nil, ErrUnknownEnvironment
	}
}

// Middleware injects environment type into request context and enforces boundary isolation
func (r *EnvironmentRouter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		env := r.DetermineEnvironment(req)
		tokenScope := req.Header.Get("X-Token-Scope")

		if tokenScope != "" {
			if err := r.ValidateRequestBoundary(env, req.URL.Path, tokenScope); err != nil {
				if errors.Is(err, ErrDemoCrossContamination) {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				if errors.Is(err, ErrScopeMismatch) {
					http.Error(w, err.Error(), http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		ctx := context.WithValue(req.Context(), EnvironmentKey, env)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
