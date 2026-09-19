package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Domain types
// ──────────────────────────────────────────────────────────────────────────────

// ChainID represents a supported EVM chain identifier.
type ChainID uint64

const (
	ChainEthereum  ChainID = 1
	ChainPolygon   ChainID = 137
	ChainArbitrum  ChainID = 42161
	ChainOptimism  ChainID = 10
	ChainAvalanche ChainID = 43114
	ChainBSC       ChainID = 56
	ChainBase      ChainID = 8453
)

// BridgeProvider enumerates supported cross-chain bridge protocols.
type BridgeProvider string

const (
	BridgeCCIP      BridgeProvider = "CHAINLINK_CCIP"
	BridgeWormhole  BridgeProvider = "WORMHOLE"
	BridgeLayerZero BridgeProvider = "LAYERZERO"
	BridgeAxelar    BridgeProvider = "AXELAR"
	BridgeStargate  BridgeProvider = "STARGATE"
)

// RouteStatus tracks the lifecycle of a collateral routing request.
type RouteStatus string

const (
	RouteStatusPending    RouteStatus = "PENDING"
	RouteStatusInFlight   RouteStatus = "IN_FLIGHT"
	RouteStatusConfirmed  RouteStatus = "CONFIRMED"
	RouteStatusFailed     RouteStatus = "FAILED"
	RouteStatusRolledBack RouteStatus = "ROLLED_BACK"
)

// CollateralAsset defines a collateral token on a specific chain.
type CollateralAsset struct {
	Symbol          string  `json:"symbol"`
	ContractAddress string  `json:"contract_address"`
	ChainID         ChainID `json:"chain_id"`
	DecimalsE8      uint8   `json:"decimals"`
}

// BridgeRoute represents a candidate path for transferring collateral cross-chain.
type BridgeRoute struct {
	Provider         BridgeProvider `json:"provider"`
	SourceChain      ChainID        `json:"source_chain"`
	DestinationChain ChainID        `json:"destination_chain"`
	EstimatedFeeUSD  float64        `json:"estimated_fee_usd"`
	EstimatedTimeMs  int64          `json:"estimated_time_ms"`
	MaxAmountE8      uint64         `json:"max_amount_e8"`
	Available        bool           `json:"available"`
}

// CollateralTransferRequest represents an inbound request to move collateral.
type CollateralTransferRequest struct {
	RequestID   string          `json:"request_id"`
	UserID      string          `json:"user_id"`
	Asset       CollateralAsset `json:"asset"`
	AmountE8    uint64          `json:"amount_e8"`
	SourceChain ChainID         `json:"source_chain"`
	DestChain   ChainID         `json:"dest_chain"`
	MaxFeeUSD   float64         `json:"max_fee_usd"`
	Urgency     string          `json:"urgency"` // "LOW", "MEDIUM", "HIGH"
}

// CollateralTransferReceipt is the output of a successful route execution.
type CollateralTransferReceipt struct {
	RequestID     string         `json:"request_id"`
	RouteHash     string         `json:"route_hash"`
	Provider      BridgeProvider `json:"provider"`
	Status        RouteStatus    `json:"status"`
	FeeChargedUSD float64        `json:"fee_charged_usd"`
	EstimatedETA  time.Duration  `json:"estimated_eta"`
	Timestamp     time.Time      `json:"timestamp"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Bridge registry & route scoring
// ──────────────────────────────────────────────────────────────────────────────

// BridgeRegistry maintains available bridges and their route capabilities.
type BridgeRegistry struct {
	mu     sync.RWMutex
	routes []BridgeRoute
}

// NewBridgeRegistry creates a new registry pre-loaded with default routes.
func NewBridgeRegistry() *BridgeRegistry {
	return &BridgeRegistry{
		routes: defaultRoutes(),
	}
}

func defaultRoutes() []BridgeRoute {
	return []BridgeRoute{
		{BridgeCCIP, ChainEthereum, ChainPolygon, 2.50, 180_000, 100_000_00000000, true},
		{BridgeCCIP, ChainEthereum, ChainArbitrum, 1.80, 120_000, 200_000_00000000, true},
		{BridgeCCIP, ChainPolygon, ChainEthereum, 3.20, 300_000, 100_000_00000000, true},
		{BridgeWormhole, ChainEthereum, ChainAvalanche, 1.50, 240_000, 500_000_00000000, true},
		{BridgeWormhole, ChainEthereum, ChainBSC, 1.20, 200_000, 500_000_00000000, true},
		{BridgeLayerZero, ChainArbitrum, ChainOptimism, 0.80, 90_000, 300_000_00000000, true},
		{BridgeLayerZero, ChainOptimism, ChainBase, 0.60, 60_000, 300_000_00000000, true},
		{BridgeAxelar, ChainEthereum, ChainPolygon, 3.00, 150_000, 250_000_00000000, true},
		{BridgeStargate, ChainArbitrum, ChainPolygon, 0.90, 80_000, 400_000_00000000, true},
		{BridgeStargate, ChainBSC, ChainPolygon, 0.70, 100_000, 400_000_00000000, true},
	}
}

// RegisterRoute adds or updates a bridge route.
func (br *BridgeRegistry) RegisterRoute(route BridgeRoute) {
	br.mu.Lock()
	defer br.mu.Unlock()
	br.routes = append(br.routes, route)
}

// FindRoutes returns all available routes between source and destination chains.
func (br *BridgeRegistry) FindRoutes(src, dst ChainID) []BridgeRoute {
	br.mu.RLock()
	defer br.mu.RUnlock()

	var matches []BridgeRoute
	for _, r := range br.routes {
		if r.SourceChain == src && r.DestinationChain == dst && r.Available {
			matches = append(matches, r)
		}
	}
	return matches
}

// ──────────────────────────────────────────────────────────────────────────────
// Route optimizer
// ──────────────────────────────────────────────────────────────────────────────

// OptimizationStrategy configures how the router selects the best route.
type OptimizationStrategy string

const (
	OptimizeCost  OptimizationStrategy = "COST"
	OptimizeSpeed OptimizationStrategy = "SPEED"
	OptimizeBlend OptimizationStrategy = "BLEND" // 60% cost + 40% speed
)

// routeScore computes a normalised score for a route (lower is better).
func routeScore(r BridgeRoute, strategy OptimizationStrategy, maxFee float64, maxTime int64) float64 {
	feePct := r.EstimatedFeeUSD / math.Max(maxFee, 0.01)
	timePct := float64(r.EstimatedTimeMs) / math.Max(float64(maxTime), 1.0)

	switch strategy {
	case OptimizeCost:
		return feePct
	case OptimizeSpeed:
		return timePct
	default: // BLEND
		return 0.6*feePct + 0.4*timePct
	}
}

// SelectOptimalRoute picks the best route from candidates based on strategy.
func SelectOptimalRoute(candidates []BridgeRoute, strategy OptimizationStrategy, maxAmountE8 uint64) (*BridgeRoute, error) {
	if len(candidates) == 0 {
		return nil, errors.New("no candidate routes available")
	}

	// Filter by capacity
	var eligible []BridgeRoute
	for _, c := range candidates {
		if c.MaxAmountE8 >= maxAmountE8 {
			eligible = append(eligible, c)
		}
	}
	if len(eligible) == 0 {
		return nil, fmt.Errorf("no route supports amount %d (e8)", maxAmountE8)
	}

	// Find max fee and max time for normalisation
	var maxFee float64
	var maxTime int64
	for _, c := range eligible {
		if c.EstimatedFeeUSD > maxFee {
			maxFee = c.EstimatedFeeUSD
		}
		if c.EstimatedTimeMs > maxTime {
			maxTime = c.EstimatedTimeMs
		}
	}

	sort.Slice(eligible, func(i, j int) bool {
		return routeScore(eligible[i], strategy, maxFee, maxTime) <
			routeScore(eligible[j], strategy, maxFee, maxTime)
	})

	best := eligible[0]
	return &best, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Collateral Router Service
// ──────────────────────────────────────────────────────────────────────────────

// CollateralRouter orchestrates cross-chain collateral transfers.
type CollateralRouter struct {
	mu       sync.Mutex
	registry *BridgeRegistry
	strategy OptimizationStrategy
	receipts map[string]*CollateralTransferReceipt
}

// NewCollateralRouter creates a router with the given bridge registry.
func NewCollateralRouter(registry *BridgeRegistry, strategy OptimizationStrategy) *CollateralRouter {
	return &CollateralRouter{
		registry: registry,
		strategy: strategy,
		receipts: make(map[string]*CollateralTransferReceipt),
	}
}

// Route processes a cross-chain collateral transfer request.
func (cr *CollateralRouter) Route(req CollateralTransferRequest) (*CollateralTransferReceipt, error) {
	if req.AmountE8 == 0 {
		return nil, errors.New("collateral amount must be > 0")
	}
	if req.SourceChain == req.DestChain {
		return nil, errors.New("source and destination chains must differ")
	}
	if req.RequestID == "" {
		return nil, errors.New("request_id is required")
	}

	candidates := cr.registry.FindRoutes(req.SourceChain, req.DestChain)

	// Apply fee ceiling filter
	if req.MaxFeeUSD > 0 {
		var filtered []BridgeRoute
		for _, c := range candidates {
			if c.EstimatedFeeUSD <= req.MaxFeeUSD {
				filtered = append(filtered, c)
			}
		}
		candidates = filtered
	}

	// Select urgency-driven strategy override
	strategy := cr.strategy
	switch req.Urgency {
	case "HIGH":
		strategy = OptimizeSpeed
	case "LOW":
		strategy = OptimizeCost
	}

	best, err := SelectOptimalRoute(candidates, strategy, req.AmountE8)
	if err != nil {
		return nil, fmt.Errorf("route selection failed: %w", err)
	}

	// Compute deterministic route hash
	payload := fmt.Sprintf("%s:%d:%d:%s:%d", req.RequestID, req.SourceChain, req.DestChain, best.Provider, req.AmountE8)
	h := sha256.Sum256([]byte(payload))
	routeHash := "RH-" + hex.EncodeToString(h[:16])

	receipt := &CollateralTransferReceipt{
		RequestID:     req.RequestID,
		RouteHash:     routeHash,
		Provider:      best.Provider,
		Status:        RouteStatusInFlight,
		FeeChargedUSD: best.EstimatedFeeUSD,
		EstimatedETA:  time.Duration(best.EstimatedTimeMs) * time.Millisecond,
		Timestamp:     time.Now().UTC(),
	}

	cr.mu.Lock()
	cr.receipts[req.RequestID] = receipt
	cr.mu.Unlock()

	fmt.Printf("[CollateralRouter] Routed %s via %s | Fee: $%.2f | ETA: %s | Hash: %s\n",
		req.RequestID, best.Provider, best.EstimatedFeeUSD, receipt.EstimatedETA, routeHash)

	return receipt, nil
}

// GetReceipt retrieves a previously generated receipt by request ID.
func (cr *CollateralRouter) GetReceipt(requestID string) (*CollateralTransferReceipt, error) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	r, ok := cr.receipts[requestID]
	if !ok {
		return nil, fmt.Errorf("receipt not found for request %s", requestID)
	}
	return r, nil
}

// ConfirmTransfer marks a routed transfer as confirmed on-chain.
func (cr *CollateralRouter) ConfirmTransfer(requestID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	r, ok := cr.receipts[requestID]
	if !ok {
		return fmt.Errorf("receipt not found: %s", requestID)
	}
	if r.Status != RouteStatusInFlight {
		return fmt.Errorf("cannot confirm transfer in status %s", r.Status)
	}
	r.Status = RouteStatusConfirmed
	return nil
}

// FailTransfer marks a routed transfer as failed.
func (cr *CollateralRouter) FailTransfer(requestID string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	r, ok := cr.receipts[requestID]
	if !ok {
		return fmt.Errorf("receipt not found: %s", requestID)
	}
	r.Status = RouteStatusFailed
	return nil
}
