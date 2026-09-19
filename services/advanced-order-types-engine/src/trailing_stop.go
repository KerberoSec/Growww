package main

import (
	"sync"
	"time"
)

// TrailingStopOrder tracks price peaks/troughs dynamically
type TrailingStopOrder struct {
	OrderID         string    `json:"order_id"`
	UserID          string    `json:"user_id"`
	Symbol          string    `json:"symbol"`
	Side            string    `json:"side"` // SELL (tracks peak) or BUY (tracks trough)
	TrailingDeltaBps uint32   `json:"trailing_delta_bps"` // e.g. 200 = 2%
	HighWaterMarkE8 uint64    `json:"high_water_mark_e8"`
	LowWaterMarkE8  uint64    `json:"low_water_mark_e8"`
	CurrentStopE8   uint64    `json:"current_stop_e8"`
	Triggered       bool      `json:"triggered"`
	CreatedAt       time.Time `json:"created_at"`
}

type TrailingStopEngine struct {
	mu     sync.RWMutex
	orders map[string]*TrailingStopOrder
}

func NewTrailingStopEngine() *TrailingStopEngine {
	return &TrailingStopEngine{
		orders: make(map[string]*TrailingStopOrder),
	}
}

func (e *TrailingStopEngine) RegisterOrder(order *TrailingStopOrder, initialPriceE8 uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	order.HighWaterMarkE8 = initialPriceE8
	order.LowWaterMarkE8 = initialPriceE8
	order.CreatedAt = time.Now().UTC()

	delta := (initialPriceE8 * uint64(order.TrailingDeltaBps)) / 10000
	if order.Side == "SELL" {
		order.CurrentStopE8 = initialPriceE8 - delta
	} else {
		order.CurrentStopE8 = initialPriceE8 + delta
	}

	e.orders[order.OrderID] = order
}

// OnPriceTick updates high/low water marks and recalculates pegged stop level
func (e *TrailingStopEngine) OnPriceTick(symbol string, priceE8 uint64) []*TrailingStopOrder {
	e.mu.Lock()
	defer e.mu.Unlock()

	var triggered []*TrailingStopOrder

	for _, order := range e.orders {
		if order.Triggered || order.Symbol != symbol {
			continue
		}

		delta := (priceE8 * uint64(order.TrailingDeltaBps)) / 10000

		if order.Side == "SELL" {
			// Higher price lifts the stop level
			if priceE8 > order.HighWaterMarkE8 {
				order.HighWaterMarkE8 = priceE8
				newStop := priceE8 - delta
				if newStop > order.CurrentStopE8 {
					order.CurrentStopE8 = newStop
				}
			} else if priceE8 <= order.CurrentStopE8 {
				order.Triggered = true
				triggered = append(triggered, order)
			}
		} else if order.Side == "BUY" {
			// Lower price drops the stop level
			if priceE8 < order.LowWaterMarkE8 {
				order.LowWaterMarkE8 = priceE8
				newStop := priceE8 + delta
				if newStop < order.CurrentStopE8 {
					order.CurrentStopE8 = newStop
				}
			} else if priceE8 >= order.CurrentStopE8 {
				order.Triggered = true
				triggered = append(triggered, order)
			}
		}
	}

	return triggered
}
