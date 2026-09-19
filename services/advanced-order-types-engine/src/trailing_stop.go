package main

import (
	"errors"
	"sync"
	"time"
)

// TrailingOrderType defines whether triggered order executes as Market or Limit
type TrailingOrderType string

const (
	TrailingOrderTypeStopMarket TrailingOrderType = "STOP_MARKET"
	TrailingOrderTypeStopLimit  TrailingOrderType = "STOP_LIMIT"
)

// TrailingStopOrder tracks price peaks/troughs dynamically
type TrailingStopOrder struct {
	OrderID               string            `json:"order_id"`
	UserID                string            `json:"user_id"`
	Symbol                string            `json:"symbol"`
	Side                  string            `json:"side"` // SELL (tracks peak) or BUY (tracks trough)
	OrderType             TrailingOrderType `json:"order_type"`
	QuantityE8            uint64            `json:"quantity_e8"`
	TrailingDeltaBps      uint32            `json:"trailing_delta_bps"`  // e.g. 200 = 2.00%
	TrailingDeltaE8       uint64            `json:"trailing_delta_e8"`   // Fixed price delta in E8 (optional)
	ActivationPriceE8     uint64            `json:"activation_price_e8"` // Optional activation threshold
	IsActivated           bool              `json:"is_activated"`
	HighWaterMarkE8       uint64            `json:"high_water_mark_e8"`
	LowWaterMarkE8        uint64            `json:"low_water_mark_e8"`
	CurrentStopE8         uint64            `json:"current_stop_e8"`
	VolatilityBufferBps   uint32            `json:"volatility_buffer_bps"` // Dynamic ATR buffer (e.g. 50 = 0.50%)
	LimitOffsetBps        uint32            `json:"limit_offset_bps"`      // Limit offset for STOP_LIMIT orders
	Triggered             bool              `json:"triggered"`
	TriggeredAt           time.Time         `json:"triggered_at,omitempty"`
	TriggerPriceE8        uint64            `json:"trigger_price_e8"`
	GeneratedLimitPriceE8 uint64            `json:"generated_limit_price_e8"`
	Canceled              bool              `json:"canceled"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

// TrailingStopEngine manages active trailing stop orders and tick adjustments
type TrailingStopEngine struct {
	mu     sync.RWMutex
	orders map[string]*TrailingStopOrder
}

// NewTrailingStopEngine creates a new TrailingStopEngine
func NewTrailingStopEngine() *TrailingStopEngine {
	return &TrailingStopEngine{
		orders: make(map[string]*TrailingStopOrder),
	}
}

// calculateDeltaE8 computes the total trailing delta in E8 given a baseline price
func (o *TrailingStopOrder) calculateDeltaE8(baselinePriceE8 uint64) uint64 {
	var delta uint64
	if o.TrailingDeltaE8 > 0 {
		delta = o.TrailingDeltaE8
	} else if o.TrailingDeltaBps > 0 {
		delta = (baselinePriceE8 * uint64(o.TrailingDeltaBps)) / 10000
	}

	// Add dynamic volatility buffer if set
	if o.VolatilityBufferBps > 0 {
		bufferDelta := (baselinePriceE8 * uint64(o.VolatilityBufferBps)) / 10000
		delta += bufferDelta
	}

	return delta
}

// RegisterOrder registers a trailing stop order and initializes watermarks
func (e *TrailingStopEngine) RegisterOrder(order *TrailingStopOrder, initialPriceE8 uint64) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}
	if order.OrderID == "" {
		return errors.New("order_id cannot be empty")
	}
	if order.Side != "SELL" && order.Side != "BUY" {
		return errors.New("side must be SELL or BUY")
	}
	if initialPriceE8 == 0 {
		return errors.New("initial price must be greater than zero")
	}
	if order.TrailingDeltaBps == 0 && order.TrailingDeltaE8 == 0 {
		return errors.New("either trailing_delta_bps or trailing_delta_e8 must be specified")
	}
	if order.OrderType == "" {
		order.OrderType = TrailingOrderTypeStopMarket
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	order.HighWaterMarkE8 = initialPriceE8
	order.LowWaterMarkE8 = initialPriceE8
	order.CreatedAt = now
	order.UpdatedAt = now
	order.Triggered = false
	order.Canceled = false

	// Activation threshold logic:
	// If activation price is zero, activate immediately.
	// For SELL: activate if initial price >= activation price.
	// For BUY: activate if initial price <= activation price.
	if order.ActivationPriceE8 == 0 {
		order.IsActivated = true
	} else {
		if order.Side == "SELL" && initialPriceE8 >= order.ActivationPriceE8 {
			order.IsActivated = true
		} else if order.Side == "BUY" && initialPriceE8 <= order.ActivationPriceE8 {
			order.IsActivated = true
		} else {
			order.IsActivated = false
		}
	}

	delta := order.calculateDeltaE8(initialPriceE8)
	if order.Side == "SELL" {
		if initialPriceE8 > delta {
			order.CurrentStopE8 = initialPriceE8 - delta
		} else {
			order.CurrentStopE8 = 1 // Floor at 1
		}
	} else { // BUY
		order.CurrentStopE8 = initialPriceE8 + delta
	}

	e.orders[order.OrderID] = order
	return nil
}

// UpdateVolatilityBuffer dynamically adjusts the volatility buffer across all active trailing orders for a symbol
func (e *TrailingStopEngine) UpdateVolatilityBuffer(symbol string, bufferBps uint32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, order := range e.orders {
		if order.Symbol == symbol && !order.Triggered && !order.Canceled {
			order.VolatilityBufferBps = bufferBps
			// Recompute stop level based on current watermarks
			if order.Side == "SELL" {
				delta := order.calculateDeltaE8(order.HighWaterMarkE8)
				if order.HighWaterMarkE8 > delta {
					newStop := order.HighWaterMarkE8 - delta
					// Trailing stop invariant: stop only moves up!
					if newStop > order.CurrentStopE8 {
						order.CurrentStopE8 = newStop
					}
				}
			} else {
				delta := order.calculateDeltaE8(order.LowWaterMarkE8)
				newStop := order.LowWaterMarkE8 + delta
				// Trailing stop invariant: stop only moves down!
				if newStop < order.CurrentStopE8 {
					order.CurrentStopE8 = newStop
				}
			}
			order.UpdatedAt = time.Now().UTC()
		}
	}
}

// OnPriceTick evaluates incoming market tick, ratchets watermarks, and fires triggers
func (e *TrailingStopEngine) OnPriceTick(symbol string, priceE8 uint64) []*TrailingStopOrder {
	e.mu.Lock()
	defer e.mu.Unlock()

	var triggered []*TrailingStopOrder

	for _, order := range e.orders {
		if order.Triggered || order.Canceled || order.Symbol != symbol {
			continue
		}

		// Check activation threshold if order is not yet activated
		if !order.IsActivated && order.ActivationPriceE8 > 0 {
			if order.Side == "SELL" && priceE8 >= order.ActivationPriceE8 {
				order.IsActivated = true
				order.HighWaterMarkE8 = priceE8
				delta := order.calculateDeltaE8(priceE8)
				if priceE8 > delta {
					order.CurrentStopE8 = priceE8 - delta
				}
			} else if order.Side == "BUY" && priceE8 <= order.ActivationPriceE8 {
				order.IsActivated = true
				order.LowWaterMarkE8 = priceE8
				delta := order.calculateDeltaE8(priceE8)
				order.CurrentStopE8 = priceE8 + delta
			}
			order.UpdatedAt = time.Now().UTC()
		}

		if order.Side == "SELL" {
			// Ratchet stop upwards if a new high-water mark is set
			if priceE8 > order.HighWaterMarkE8 {
				order.HighWaterMarkE8 = priceE8
				delta := order.calculateDeltaE8(priceE8)
				var newStop uint64
				if priceE8 > delta {
					newStop = priceE8 - delta
				} else {
					newStop = 1
				}
				// Trailing stop invariant: current stop must strictly ratchet UPWARDS, NEVER downwards!
				if newStop > order.CurrentStopE8 {
					order.CurrentStopE8 = newStop
				}
				order.UpdatedAt = time.Now().UTC()
			} else if order.IsActivated && priceE8 <= order.CurrentStopE8 {
				// Breached trailing stop trigger condition
				order.Triggered = true
				order.TriggerPriceE8 = priceE8
				order.TriggeredAt = time.Now().UTC()
				order.UpdatedAt = time.Now().UTC()

				// If Stop-Limit, generate limit order price
				if order.OrderType == TrailingOrderTypeStopLimit {
					var offset uint64
					if order.LimitOffsetBps > 0 {
						offset = (priceE8 * uint64(order.LimitOffsetBps)) / 10000
					}
					if priceE8 > offset {
						order.GeneratedLimitPriceE8 = priceE8 - offset
					} else {
						order.GeneratedLimitPriceE8 = 1
					}
				}
				triggered = append(triggered, order)
			}
		} else if order.Side == "BUY" {
			// Ratchet stop downwards if a new low-water mark is set
			if priceE8 < order.LowWaterMarkE8 {
				order.LowWaterMarkE8 = priceE8
				delta := order.calculateDeltaE8(priceE8)
				newStop := priceE8 + delta
				// Trailing stop invariant: current stop must strictly ratchet DOWNWARDS, NEVER upwards!
				if newStop < order.CurrentStopE8 {
					order.CurrentStopE8 = newStop
				}
				order.UpdatedAt = time.Now().UTC()
			} else if order.IsActivated && priceE8 >= order.CurrentStopE8 {
				// Breached trailing stop trigger condition
				order.Triggered = true
				order.TriggerPriceE8 = priceE8
				order.TriggeredAt = time.Now().UTC()
				order.UpdatedAt = time.Now().UTC()

				// If Stop-Limit, generate limit order price
				if order.OrderType == TrailingOrderTypeStopLimit {
					var offset uint64
					if order.LimitOffsetBps > 0 {
						offset = (priceE8 * uint64(order.LimitOffsetBps)) / 10000
					}
					order.GeneratedLimitPriceE8 = priceE8 + offset
				}
				triggered = append(triggered, order)
			}
		}
	}

	return triggered
}

// CancelOrder marks an active trailing stop order as canceled
func (e *TrailingStopEngine) CancelOrder(orderID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	order, exists := e.orders[orderID]
	if !exists {
		return errors.New("trailing stop order not found")
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

// GetOrder returns a copy of the trailing stop order state
func (e *TrailingStopEngine) GetOrder(orderID string) (*TrailingStopOrder, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	order, exists := e.orders[orderID]
	if !exists {
		return nil, false
	}
	copyOrder := *order
	return &copyOrder, true
}

// ListActiveOrders returns all active trailing stop orders
func (e *TrailingStopEngine) ListActiveOrders() []*TrailingStopOrder {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var active []*TrailingStopOrder
	for _, o := range e.orders {
		if !o.Triggered && !o.Canceled {
			copyOrder := *o
			active = append(active, &copyOrder)
		}
	}
	return active
}
