package main

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ADLPriority determines which positions are auto-deleveraged first.
// Ranked by profitability * leverage score, highest first.
type ADLPriority struct {
	UserID         string  `json:"user_id"`
	PositionID     string  `json:"position_id"`
	Side           string  `json:"side"` // LONG or SHORT
	UnrealizedPnL  float64 `json:"unrealized_pnl"`
	LeverageRatio  float64 `json:"leverage_ratio"`
	PriorityScore  float64 `json:"priority_score"`
}

// ADLEvent records an auto-deleveraging execution
type ADLEvent struct {
	EventID          string    `json:"event_id"`
	DefaulterUserID  string    `json:"defaulter_user_id"`
	CounterpartyID   string    `json:"counterparty_id"`
	PositionID       string    `json:"position_id"`
	QuantityE8       uint64    `json:"quantity_e8"`
	SettlementPrice  float64   `json:"settlement_price"`
	SocializedLoss   float64   `json:"socialized_loss"`
	Timestamp        time.Time `json:"timestamp"`
}

// InsuranceFund holds the exchange insurance reserve
type InsuranceFund struct {
	BalanceUSDT float64 `json:"balance_usdt"`
}

// ADLEngine manages Auto-Deleveraging queue, socialized loss distribution,
// and insurance fund drawdown for margin defaults.
type ADLEngine struct {
	mu             sync.Mutex
	queue          []ADLPriority
	events         []ADLEvent
	insuranceFund  InsuranceFund
	eventCounter   uint64
}

// NewADLEngine creates an ADL engine with initial insurance fund balance
func NewADLEngine(insuranceBalance float64) *ADLEngine {
	return &ADLEngine{
		queue:         make([]ADLPriority, 0),
		events:        make([]ADLEvent, 0),
		insuranceFund: InsuranceFund{BalanceUSDT: insuranceBalance},
	}
}

// AddToQueue registers a profitable position for potential ADL deleveraging.
// PriorityScore = abs(UnrealizedPnL) * LeverageRatio
func (e *ADLEngine) AddToQueue(userID, positionID, side string, pnl, leverage float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	score := pnl * leverage
	if score < 0 {
		score = -score
	}

	e.queue = append(e.queue, ADLPriority{
		UserID:        userID,
		PositionID:    positionID,
		Side:          side,
		UnrealizedPnL: pnl,
		LeverageRatio: leverage,
		PriorityScore: score,
	})

	// Maintain sorted order: highest priority score first
	sort.Slice(e.queue, func(i, j int) bool {
		return e.queue[i].PriorityScore > e.queue[j].PriorityScore
	})
}

// GetQueue returns the current ADL priority queue (snapshot)
func (e *ADLEngine) GetQueue() []ADLPriority {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make([]ADLPriority, len(e.queue))
	copy(result, e.queue)
	return result
}

// ResolveDefault processes a margin default by:
//  1. Drawing from the insurance fund first
//  2. If insurance fund is insufficient, socializing losses via ADL
//
// Returns the list of ADL events triggered.
func (e *ADLEngine) ResolveDefault(defaulterID string, lossUSDT float64, settlementPrice float64) ([]ADLEvent, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if lossUSDT <= 0 {
		return nil, errors.New("loss amount must be positive")
	}

	remainingLoss := lossUSDT
	var triggered []ADLEvent

	// Step 1: Insurance fund drawdown
	if e.insuranceFund.BalanceUSDT > 0 {
		drawn := e.insuranceFund.BalanceUSDT
		if drawn > remainingLoss {
			drawn = remainingLoss
		}
		e.insuranceFund.BalanceUSDT -= drawn
		remainingLoss -= drawn
		fmt.Printf("[ADL] Insurance fund absorbed %.4f USDT of default loss\n", drawn)
	}

	if remainingLoss <= 0 {
		return triggered, nil
	}

	// Step 2: Socialize remaining loss via ADL queue
	if len(e.queue) == 0 {
		return triggered, fmt.Errorf("ADL queue empty: %.4f USDT loss unresolved", remainingLoss)
	}

	for remainingLoss > 0 && len(e.queue) > 0 {
		top := e.queue[0]
		e.queue = e.queue[1:]

		socializedAmount := remainingLoss
		// Cap at counterparty's unrealized PnL (absolute value)
		maxAbsorb := top.UnrealizedPnL
		if maxAbsorb < 0 {
			maxAbsorb = -maxAbsorb
		}
		if socializedAmount > maxAbsorb {
			socializedAmount = maxAbsorb
		}

		e.eventCounter++
		evt := ADLEvent{
			EventID:         fmt.Sprintf("ADL-%06d", e.eventCounter),
			DefaulterUserID: defaulterID,
			CounterpartyID:  top.UserID,
			PositionID:      top.PositionID,
			QuantityE8:      uint64(socializedAmount * 1e8),
			SettlementPrice: settlementPrice,
			SocializedLoss:  socializedAmount,
			Timestamp:       time.Now().UTC(),
		}
		triggered = append(triggered, evt)
		e.events = append(e.events, evt)

		remainingLoss -= socializedAmount
		fmt.Printf("[ADL] Socialized %.4f USDT from counterparty %s (position %s)\n",
			socializedAmount, top.UserID, top.PositionID)
	}

	return triggered, nil
}

// GetInsuranceFundBalance returns the current insurance fund balance
func (e *ADLEngine) GetInsuranceFundBalance() float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.insuranceFund.BalanceUSDT
}

// GetADLHistory returns all ADL events
func (e *ADLEngine) GetADLHistory() []ADLEvent {
	e.mu.Lock()
	defer e.mu.Unlock()

	result := make([]ADLEvent, len(e.events))
	copy(result, e.events)
	return result
}
