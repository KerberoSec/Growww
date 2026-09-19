package main

import (
	"sync"
	"time"
)

type TriggerType string

const (
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
	CreatedAt       time.Time   `json:"created_at"`
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

func (e *TriggerEngine) RegisterConditionalOrder(order *ConditionalOrder) {
	e.mu.Lock()
	defer e.mu.Unlock()
	order.CreatedAt = time.Now().UTC()
	e.orders[order.OrderID] = order
}

// EvaluateMarketPrice checks resting triggers against the latest mark/last price
func (e *TriggerEngine) EvaluateMarketPrice(symbol string, currentPriceE8 uint64) []*ConditionalOrder {
	e.mu.Lock()
	defer e.mu.Unlock()

	var triggeredOrders []*ConditionalOrder

	for _, order := range e.orders {
		if order.Triggered || order.Symbol != symbol {
			continue
		}

		shouldTrigger := false
		switch order.TriggerType {
		case TriggerStopLoss:
			// Sell Stop Loss triggers when price falls <= stop price
			if order.Side == "SELL" && currentPriceE8 <= order.StopPriceE8 {
				shouldTrigger = true
			} else if order.Side == "BUY" && currentPriceE8 >= order.StopPriceE8 {
				shouldTrigger = true
			}
		case TriggerTakeProfit:
			// Sell Take Profit triggers when price rises >= stop price
			if order.Side == "SELL" && currentPriceE8 >= order.StopPriceE8 {
				shouldTrigger = true
			} else if order.Side == "BUY" && currentPriceE8 <= order.StopPriceE8 {
				shouldTrigger = true
			}
		}

		if shouldTrigger {
			order.Triggered = true
			triggeredOrders = append(triggeredOrders, order)
		}
	}

	return triggeredOrders
}
