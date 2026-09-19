package main

import (
	"context"
	"errors"
	"net/http"
)

type EnvironmentType string

const (
	EnvDemoTestnet EnvironmentType = "DEMO_TESTNET"
	EnvRealMainnet EnvironmentType = "REAL_MAINNET"
)

// EnvironmentContextKey is the context key for trading mode
type contextKey string

const EnvironmentKey contextKey = "trading_environment"

// EnvironmentRouter inspects request headers and routes to isolated backends
type EnvironmentRouter struct {
	DemoMatchingHost string
	RealMatchingHost string
}

func NewEnvironmentRouter(demoHost, realHost string) *EnvironmentRouter {
	return &EnvironmentRouter{
		DemoMatchingHost: demoHost,
		RealMatchingHost: realHost,
	}
}

// DetermineEnvironment extracts mode from X-Trading-Mode header
func (r *EnvironmentRouter) DetermineEnvironment(req *http.Request) EnvironmentType {
	mode := req.Header.Get("X-Trading-Mode")
	if mode == "REAL" || mode == "MAINNET" {
		return EnvRealMainnet
	}
	// Default to isolated demo sandbox for user safety
	return EnvDemoTestnet
}

// RouteTarget returns the appropriate cluster endpoint based on isolation mode
func (r *EnvironmentRouter) RouteTarget(env EnvironmentType) (string, error) {
	switch env {
	case EnvDemoTestnet:
		return r.DemoMatchingHost, nil
	case EnvRealMainnet:
		return r.RealMatchingHost, nil
	default:
		return "", errors.New("unknown environment type")
	}
}

// Middleware injects environment type into request context
func (r *EnvironmentRouter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		env := r.DetermineEnvironment(req)
		ctx := context.WithValue(req.Context(), EnvironmentKey, env)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
