package main

import (
	"errors"
	"sync"
	"time"
)

type OCOStatus string

const (
	OCOActive    OCOStatus = "ACTIVE"
	OCOLimitWon  OCOStatus = "LIMIT_FILLED"
	OCOStopWon   OCOStatus = "STOP_TRIGGERED"
	OCOCanceled  OCOStatus = "CANCELED"
)

// OCOPair couples a resting limit maker order and a conditional stop-loss order
type OCOPair struct {
	PairID         string    `json:"pair_id"`
	UserID         string    `json:"user_id"`
	Symbol         string    `json:"symbol"`
	LimitOrderID   string    `json:"limit_order_id"`
	LimitPriceE8   uint64    `json:"limit_price_e8"`
	StopOrderID    string    `json:"stop_order_id"`
	StopPriceE8    uint64    `json:"stop_price_e8"`
	StopLimitE8    uint64    `json:"stop_limit_e8"`
	QuantityE8     uint64    `json:"quantity_e8"`
	Status         OCOStatus `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type OCOOrchestrator struct {
	mu    sync.Mutex
	pairs map[string]*OCOPair
}

func NewOCOOrchestrator() *OCOOrchestrator {
	return &OCOOrchestrator{
		pairs: make(map[string]*OCOPair),
	}
}

func (o *OCOOrchestrator) RegisterOCO(pair *OCOPair) {
	o.mu.Lock()
	defer o.mu.Unlock()
	pair.Status = OCOActive
	pair.CreatedAt = time.Now().UTC()
	o.pairs[pair.PairID] = pair
}

// OnLimitOrderFill cancels the associated stop-loss trigger when limit leg fills
func (o *OCOOrchestrator) OnLimitOrderFill(limitOrderID string) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, pair := range o.pairs {
		if pair.LimitOrderID == limitOrderID && pair.Status == OCOActive {
			pair.Status = OCOLimitWon
			return pair.StopOrderID, nil // Return StopOrderID to be canceled
		}
	}
	return "", errors.New("active OCO pair not found for limit order")
}

// OnStopTrigger cancels the resting limit order when the stop price is hit
func (o *OCOOrchestrator) OnStopTrigger(stopOrderID string) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, pair := range o.pairs {
		if pair.StopOrderID == stopOrderID && pair.Status == OCOActive {
			pair.Status = OCOStopWon
			return pair.LimitOrderID, nil // Return LimitOrderID to be canceled from book
		}
	}
	return "", errors.New("active OCO pair not found for stop order")
}
