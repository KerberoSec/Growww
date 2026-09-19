package main

import (
	"errors"
	"sync"
	"time"
)

type TriggerType string

const (
	TriggerStopLossMarket   TriggerType = "STOP_LOSS_MARKET"
	TriggerStopLossLimit    TriggerType = "STOP_LOSS_LIMIT"
	TriggerTakeProfitMarket TriggerType = "TAKE_PROFIT_MARKET"
	TriggerTakeProfitLimit  TriggerType = "TAKE_PROFIT_LIMIT"

	// Legacy aliases
	TriggerStopLoss   TriggerType = "STOP_LOSS"
	TriggerTakeProfit TriggerType = "TAKE_PROFIT"
)

type ConditionalOrder struct {
	OrderID         string      `json:"order_id"`
	UserID          string      `json:"user_id"`
	Symbol          string      `json:"symbol"`
	Side            string      `json:"side"` // BUY or SELL
	TriggerType     TriggerType `json:"trigger_type"`
	StopPriceE8     uint64      `json:"stop_price_e8"`
	LimitPriceE8    uint64      `json:"limit_price_e8"`
	QuantityE8      uint64      `json:"quantity_e8"`
	Triggered       bool        `json:"triggered"`
	TriggeredPriceE8 uint64     `json:"triggered_price_e8"`
	TriggeredAt     time.Time   `json:"triggered_at,omitempty"`
	Canceled        bool        `json:"canceled"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type TriggerEngine struct {
	mu     sync.RWMutex
	orders map[string]*ConditionalOrder
}

func NewTriggerEngine() *TriggerEngine {
	return &TriggerEngine{
		orders: make(map[string]*ConditionalOrder),
	}
}

func (e *TriggerEngine) RegisterConditionalOrder(order *ConditionalOrder) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}
	if order.OrderID == "" {
		return errors.New("order_id cannot be empty")
	}
	if order.StopPriceE8 == 0 {
		return errors.New("stop_price_e8 must be greater than zero")
	}
	if order.QuantityE8 == 0 {
		return errors.New("quantity_e8 must be greater than zero")
	}
	if order.Side != "BUY" && order.Side != "SELL" {
		return errors.New("side must be BUY or SELL")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	order.CreatedAt = now
	order.UpdatedAt = now
	order.Triggered = false
	order.Canceled = false
	e.orders[order.OrderID] = order
	return nil
}

// EvaluateMarketPrice checks resting triggers against the latest mark/last price
func (e *TriggerEngine) EvaluateMarketPrice(symbol string, currentPriceE8 uint64) []*ConditionalOrder {
	e.mu.Lock()
	defer e.mu.Unlock()

	var triggeredOrders []*ConditionalOrder

	for _, order := range e.orders {
		if order.Triggered || order.Canceled || order.Symbol != symbol {
			continue
		}

		shouldTrigger := false
		switch order.TriggerType {
		case TriggerStopLoss, TriggerStopLossMarket, TriggerStopLossLimit:
			// Sell Stop Loss triggers when price falls <= stop price
			if order.Side == "SELL" && currentPriceE8 <= order.StopPriceE8 {
				shouldTrigger = true
			} else if order.Side == "BUY" && currentPriceE8 >= order.StopPriceE8 {
				shouldTrigger = true
			}
		case TriggerTakeProfit, TriggerTakeProfitMarket, TriggerTakeProfitLimit:
			// Sell Take Profit triggers when price rises >= stop price
			if order.Side == "SELL" && currentPriceE8 >= order.StopPriceE8 {
				shouldTrigger = true
			} else if order.Side == "BUY" && currentPriceE8 <= order.StopPriceE8 {
				shouldTrigger = true
			}
		}

		if shouldTrigger {
			now := time.Now().UTC()
			order.Triggered = true
			order.TriggeredPriceE8 = currentPriceE8
			order.TriggeredAt = now
			order.UpdatedAt = now
			triggeredOrders = append(triggeredOrders, order)
		}
	}

	return triggeredOrders
}

// CancelOrder marks an active conditional order as canceled
func (e *TriggerEngine) CancelOrder(orderID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	order, exists := e.orders[orderID]
	if !exists {
		return errors.New("conditional order not found")
	}
	if order.Triggered {
		return errors.New("cannot cancel already triggered order")
	}
	if order.Canceled {
		return errors.New("order is already canceled")
	}

	order.Canceled = true
	order.UpdatedAt = time.Now().UTC()
	return nil
}

// GetOrder returns a conditional order copy
func (e *TriggerEngine) GetOrder(orderID string) (*ConditionalOrder, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	order, exists := e.orders[orderID]
	if !exists {
		return nil, false
	}
	copyOrder := *order
	return &copyOrder, true
}

// ListActiveOrders returns all un-triggered, non-canceled orders
func (e *TriggerEngine) ListActiveOrders() []*ConditionalOrder {
	e.mu.RLock()
	defer e.mu.RUnlock()

	active := make([]*ConditionalOrder, 0)
	for _, o := range e.orders {
		if !o.Triggered && !o.Canceled {
			copyOrder := *o
			active = append(active, &copyOrder)
		}
	}
	return active
}
