// Package main enhances the NBSE risk-margin service with real-time margin calculation,
// position risk aggregation, and liquidation threshold monitoring.
package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// PositionSide represents the direction of a position.
type PositionSide string

const (
	PositionLong  PositionSide = "LONG"
	PositionShort PositionSide = "SHORT"
)

// LiquidationLevel defines the margin thresholds.
type LiquidationLevel string

const (
	LevelSafe       LiquidationLevel = "SAFE"          // > 150%
	LevelWarning    LiquidationLevel = "WARNING"       // 120% - 150%
	LevelDanger     LiquidationLevel = "DANGER"        // 110% - 120%
	LevelLiquidate  LiquidationLevel = "LIQUIDATE"     // < 110%
)

// Position represents a trading position.
type Position struct {
	PositionID   string       `json:"position_id"`
	UserID       string       `json:"user_id"`
	Symbol       string       `json:"symbol"`
	Side         PositionSide `json:"side"`
	EntryPriceE8 uint64       `json:"entry_price_e8"`
	QuantityE8   uint64       `json:"quantity_e8"`
	Leverage     uint8        `json:"leverage"`
	MarginUsedE8 uint64       `json:"margin_used_e8"`
	OpenedAt     time.Time    `json:"opened_at"`
}

// PositionRiskSummary holds aggregated risk data for a user's positions.
type PositionRiskSummary struct {
	UserID             string           `json:"user_id"`
	TotalPositionValue float64          `json:"total_position_value_usd"`
	TotalMarginUsed    float64          `json:"total_margin_used_usd"`
	TotalUnrealizedPnL float64          `json:"total_unrealized_pnl_usd"`
	MarginLevel        float64          `json:"margin_level_pct"`
	LiquidationLevel   LiquidationLevel `json:"liquidation_level"`
	PositionCount      int              `json:"position_count"`
}

// MarginRequirement defines the margin needed for an order.
type MarginRequirement struct {
	Symbol           string  `json:"symbol"`
	Side             string  `json:"side"`
	NotionalUSD      float64 `json:"notional_usd"`
	InitialMarginPct float64 `json:"initial_margin_pct"`
	MaintenanceMarginPct float64 `json:"maintenance_margin_pct"`
	InitialMarginUSD float64 `json:"initial_margin_usd"`
	MaintenanceMarginUSD float64 `json:"maintenance_margin_usd"`
}

// RealTimeMarginEngine manages positions, margin calculations, and liquidation monitoring.
type RealTimeMarginEngine struct {
	mu                sync.RWMutex
	positions         map[string]*Position           // positionID -> position
	userPositions     map[string][]string             // userID -> positionIDs
	markPrices        map[string]uint64               // symbol -> current mark price E8
	initialMarginPcts map[string]float64              // symbol -> initial margin %
	maintenanceMarginPcts map[string]float64          // symbol -> maintenance margin %
}

// NewRealTimeMarginEngine creates a new RealTimeMarginEngine.
func NewRealTimeMarginEngine() *RealTimeMarginEngine {
	return &RealTimeMarginEngine{
		positions:             make(map[string]*Position),
		userPositions:         make(map[string][]string),
		markPrices:            make(map[string]uint64),
		initialMarginPcts:     make(map[string]float64),
		maintenanceMarginPcts: make(map[string]float64),
	}
}

// SetMarginParams configures initial and maintenance margin percentages for a symbol.
func (e *RealTimeMarginEngine) SetMarginParams(symbol string, initialPct, maintenancePct float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.initialMarginPcts[symbol] = initialPct
	e.maintenanceMarginPcts[symbol] = maintenancePct
}

// UpdateMarkPrice updates the current mark price for a symbol.
func (e *RealTimeMarginEngine) UpdateMarkPrice(symbol string, priceE8 uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.markPrices[symbol] = priceE8
}

// CalculateMarginRequired computes the margin needed for a new order.
func (e *RealTimeMarginEngine) CalculateMarginRequired(symbol, side string, priceE8, quantityE8 uint64) MarginRequirement {
	e.mu.RLock()
	defer e.mu.RUnlock()

	notionalUSD := (float64(priceE8) / 1e8) * (float64(quantityE8) / 1e8)

	initialPct := e.initialMarginPcts[symbol]
	if initialPct == 0 {
		initialPct = 10.0 // Default 10%
	}
	maintenancePct := e.maintenanceMarginPcts[symbol]
	if maintenancePct == 0 {
		maintenancePct = 5.0 // Default 5%
	}

	return MarginRequirement{
		Symbol:               symbol,
		Side:                 side,
		NotionalUSD:          notionalUSD,
		InitialMarginPct:     initialPct,
		MaintenanceMarginPct: maintenancePct,
		InitialMarginUSD:     notionalUSD * (initialPct / 100.0),
		MaintenanceMarginUSD: notionalUSD * (maintenancePct / 100.0),
	}
}

// OpenPosition records a new position.
func (e *RealTimeMarginEngine) OpenPosition(pos Position) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.positions[pos.PositionID]; exists {
		return fmt.Errorf("position %s already exists", pos.PositionID)
	}

	e.positions[pos.PositionID] = &pos
	e.userPositions[pos.UserID] = append(e.userPositions[pos.UserID], pos.PositionID)
	return nil
}

// CalculateUnrealizedPnL computes unrealized PnL for a position against mark price.
func (e *RealTimeMarginEngine) CalculateUnrealizedPnL(positionID string) (float64, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pos, exists := e.positions[positionID]
	if !exists {
		return 0, fmt.Errorf("position %s not found", positionID)
	}

	markPrice, exists := e.markPrices[pos.Symbol]
	if !exists {
		return 0, fmt.Errorf("no mark price for %s", pos.Symbol)
	}

	entry := float64(pos.EntryPriceE8) / 1e8
	mark := float64(markPrice) / 1e8
	qty := float64(pos.QuantityE8) / 1e8

	var pnl float64
	if pos.Side == PositionLong {
		pnl = (mark - entry) * qty
	} else {
		pnl = (entry - mark) * qty
	}

	return pnl, nil
}

// AggregatePositionRisk computes total risk summary for a user.
func (e *RealTimeMarginEngine) AggregatePositionRisk(userID string, equityUSD float64) (*PositionRiskSummary, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	posIDs, exists := e.userPositions[userID]
	if !exists || len(posIDs) == 0 {
		return &PositionRiskSummary{
			UserID:           userID,
			MarginLevel:      math.MaxFloat64,
			LiquidationLevel: LevelSafe,
		}, nil
	}

	var totalValue, totalMargin, totalPnL float64
	for _, pid := range posIDs {
		pos := e.positions[pid]
		markPrice := e.markPrices[pos.Symbol]

		qty := float64(pos.QuantityE8) / 1e8
		mark := float64(markPrice) / 1e8
		entry := float64(pos.EntryPriceE8) / 1e8

		posValue := mark * qty
		totalValue += posValue
		totalMargin += float64(pos.MarginUsedE8) / 1e8

		if pos.Side == PositionLong {
			totalPnL += (mark - entry) * qty
		} else {
			totalPnL += (entry - mark) * qty
		}
	}

	var marginLevel float64
	if totalMargin > 0 {
		marginLevel = ((equityUSD + totalPnL) / totalMargin) * 100.0
	} else {
		marginLevel = math.MaxFloat64
	}

	level := classifyLiquidationLevel(marginLevel)

	return &PositionRiskSummary{
		UserID:             userID,
		TotalPositionValue: totalValue,
		TotalMarginUsed:    totalMargin,
		TotalUnrealizedPnL: totalPnL,
		MarginLevel:        marginLevel,
		LiquidationLevel:   level,
		PositionCount:      len(posIDs),
	}, nil
}

// classifyLiquidationLevel maps margin level percentage to a liquidation level.
func classifyLiquidationLevel(marginLevelPct float64) LiquidationLevel {
	switch {
	case marginLevelPct > 150:
		return LevelSafe
	case marginLevelPct > 120:
		return LevelWarning
	case marginLevelPct > 110:
		return LevelDanger
	default:
		return LevelLiquidate
	}
}
