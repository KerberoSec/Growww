package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type OCOStatus string

const (
	OCOActive   OCOStatus = "ACTIVE"
	OCOLimitWon OCOStatus = "LIMIT_FILLED"
	OCOStopWon  OCOStatus = "STOP_TRIGGERED"
	OCOCanceled OCOStatus = "CANCELED"
)

// OCOPair couples a resting limit maker order and a conditional stop-loss order
type OCOPair struct {
	PairID         string    `json:"pair_id"`
	UserID         string    `json:"user_id"`
	Symbol         string    `json:"symbol"`
	Side           string    `json:"side"` // BUY or SELL
	LimitOrderID   string    `json:"limit_order_id"`
	LimitPriceE8   uint64    `json:"limit_price_e8"`
	StopOrderID    string    `json:"stop_order_id"`
	StopPriceE8    uint64    `json:"stop_price_e8"`
	StopLimitE8    uint64    `json:"stop_limit_e8"`
	QuantityE8     uint64    `json:"quantity_e8"`
	RemainingQtyE8 uint64    `json:"remaining_qty_e8"`
	Status         OCOStatus `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// OCOOrchestrator manages coupled OCO / bracket order lifecycle
type OCOOrchestrator struct {
	mu    sync.RWMutex
	pairs map[string]*OCOPair
}

// NewOCOOrchestrator constructs an OCO orchestrator
func NewOCOOrchestrator() *OCOOrchestrator {
	return &OCOOrchestrator{
		pairs: make(map[string]*OCOPair),
	}
}

// RegisterOCO validates price relationships and registers an active OCO pair
func (o *OCOOrchestrator) RegisterOCO(pair *OCOPair) error {
	if pair == nil {
		return errors.New("pair cannot be nil")
	}
	if pair.PairID == "" || pair.LimitOrderID == "" || pair.StopOrderID == "" {
		return errors.New("pair_id, limit_order_id, and stop_order_id required")
	}
	if pair.QuantityE8 == 0 {
		return errors.New("quantity must be greater than zero")
	}

	// Validate price bounds:
	// For SELL OCO: take profit limit > stop loss trigger
	// For BUY OCO: take profit limit < stop loss trigger
	if pair.Side == "SELL" && pair.LimitPriceE8 <= pair.StopPriceE8 {
		return fmt.Errorf("for sell OCO, limit price (%d) must be strictly greater than stop price (%d)",
			pair.LimitPriceE8, pair.StopPriceE8)
	}
	if pair.Side == "BUY" && pair.LimitPriceE8 >= pair.StopPriceE8 {
		return fmt.Errorf("for buy OCO, limit price (%d) must be strictly less than stop price (%d)",
			pair.LimitPriceE8, pair.StopPriceE8)
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	now := time.Now().UTC()
	pair.Status = OCOActive
	pair.RemainingQtyE8 = pair.QuantityE8
	pair.CreatedAt = now
	pair.UpdatedAt = now
	o.pairs[pair.PairID] = pair
	return nil
}

// OnLimitOrderFill cancels the associated stop-loss trigger when the limit leg fills completely
func (o *OCOOrchestrator) OnLimitOrderFill(limitOrderID string) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, pair := range o.pairs {
		if pair.LimitOrderID == limitOrderID && pair.Status == OCOActive {
			pair.Status = OCOLimitWon
			pair.RemainingQtyE8 = 0
			pair.UpdatedAt = time.Now().UTC()
			return pair.StopOrderID, nil // Return StopOrderID to be canceled immediately
		}
	}
	return "", errors.New("active OCO pair not found for limit order")
}

// OnLimitOrderPartialFill handles partial fills on the limit leg by reducing the contingent stop order quantity
func (o *OCOOrchestrator) OnLimitOrderPartialFill(limitOrderID string, filledQtyE8 uint64) (string, uint64, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, pair := range o.pairs {
		if pair.LimitOrderID == limitOrderID && pair.Status == OCOActive {
			if filledQtyE8 >= pair.RemainingQtyE8 {
				pair.Status = OCOLimitWon
				pair.RemainingQtyE8 = 0
				pair.UpdatedAt = time.Now().UTC()
				return pair.StopOrderID, 0, nil // Fully filled
			}
			pair.RemainingQtyE8 -= filledQtyE8
			pair.UpdatedAt = time.Now().UTC()
			return pair.StopOrderID, pair.RemainingQtyE8, nil
		}
	}
	return "", 0, errors.New("active OCO pair not found for limit order")
}

// OnStopTrigger cancels the resting limit order when the stop price is hit
func (o *OCOOrchestrator) OnStopTrigger(stopOrderID string) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, pair := range o.pairs {
		if pair.StopOrderID == stopOrderID && pair.Status == OCOActive {
			pair.Status = OCOStopWon
			pair.UpdatedAt = time.Now().UTC()
			return pair.LimitOrderID, nil // Return LimitOrderID to be canceled from book
		}
	}
	return "", errors.New("active OCO pair not found for stop order")
}

// CancelOCO cancels both legs of an active OCO order
func (o *OCOOrchestrator) CancelOCO(pairID string) (string, string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	pair, exists := o.pairs[pairID]
	if !exists {
		return "", "", errors.New("oco pair not found")
	}
	if pair.Status != OCOActive {
		return "", "", fmt.Errorf("cannot cancel OCO pair in status %s", pair.Status)
	}

	pair.Status = OCOCanceled
	pair.UpdatedAt = time.Now().UTC()
	return pair.LimitOrderID, pair.StopOrderID, nil
}

// GetOCO returns an OCO pair snapshot
func (o *OCOOrchestrator) GetOCO(pairID string) (*OCOPair, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pair, exists := o.pairs[pairID]
	if !exists {
		return nil, false
	}
	copied := *pair
	return &copied, true
}
