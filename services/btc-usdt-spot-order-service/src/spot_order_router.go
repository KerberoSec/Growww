package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// OrderType represents the type of spot order
type OrderType string

const (
	OrderTypeLimit      OrderType = "LIMIT"
	OrderTypeMarket     OrderType = "MARKET"
	OrderTypeStopLimit  OrderType = "STOP_LIMIT"
)

// OrderValidationResult contains pre-trade validation details
type OrderValidationResult struct {
	Valid        bool     `json:"valid"`
	Errors       []string `json:"errors,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// CancelResult represents the outcome of a cancel request
type CancelResult struct {
	OrderID     string          `json:"order_id"`
	PrevStatus  SpotOrderStatus `json:"previous_status"`
	NewStatus   SpotOrderStatus `json:"new_status"`
	CancelledAt time.Time       `json:"cancelled_at"`
}

// SpotOrderRouter routes orders and provides validation, cancel and query operations
type SpotOrderRouter struct {
	mu     sync.RWMutex
	svc    *SpotOrderLifecycleService
	minOrderE8 uint64 // Minimum order size in E8
	maxOrderE8 uint64 // Maximum order size in E8
}

// NewSpotOrderRouter creates a new router wrapping the lifecycle service
func NewSpotOrderRouter(svc *SpotOrderLifecycleService, minE8, maxE8 uint64) *SpotOrderRouter {
	if minE8 == 0 {
		minE8 = 1000 // 0.00001 BTC minimum
	}
	if maxE8 == 0 {
		maxE8 = 100000000000 // 1000 BTC maximum
	}
	return &SpotOrderRouter{
		svc:        svc,
		minOrderE8: minE8,
		maxOrderE8: maxE8,
	}
}

// ValidateOrder performs pre-trade validation checks
func (r *SpotOrderRouter) ValidateOrder(symbol, side string, priceE8, qtyE8 uint64) OrderValidationResult {
	result := OrderValidationResult{Valid: true}

	if symbol != "BTC/USDT" {
		result.Valid = false
		result.Errors = append(result.Errors, "only BTC/USDT spot pair is supported")
	}

	if side != "BUY" && side != "SELL" {
		result.Valid = false
		result.Errors = append(result.Errors, "side must be BUY or SELL")
	}

	if qtyE8 < r.minOrderE8 {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("quantity %d below minimum %d", qtyE8, r.minOrderE8))
	}

	if qtyE8 > r.maxOrderE8 {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("quantity %d exceeds maximum %d", qtyE8, r.maxOrderE8))
	}

	if priceE8 == 0 {
		result.Warnings = append(result.Warnings, "price is zero — will be treated as market order")
	}

	return result
}

// RouteOrder validates and routes a spot order to the execution pipeline
func (r *SpotOrderRouter) RouteOrder(orderID, userID, symbol, side string, priceE8, qtyE8 uint64) (*RealSpotOrder, error) {
	validation := r.ValidateOrder(symbol, side, priceE8, qtyE8)
	if !validation.Valid {
		return nil, fmt.Errorf("order validation failed: %v", validation.Errors)
	}
	return r.svc.SubmitOrder(orderID, userID, symbol, side, priceE8, qtyE8)
}

// CancelOrder cancels an open or partially filled order
func (r *SpotOrderRouter) CancelOrder(orderID string) (*CancelResult, error) {
	r.svc.mu.Lock()
	defer r.svc.mu.Unlock()

	order, exists := r.svc.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}

	if order.Status == OrderFilled || order.Status == OrderCanceled {
		return nil, fmt.Errorf("cannot cancel order in %s status", order.Status)
	}

	prevStatus := order.Status
	order.Status = OrderCanceled
	order.UpdatedAt = time.Now().UTC()

	return &CancelResult{
		OrderID:     orderID,
		PrevStatus:  prevStatus,
		NewStatus:   OrderCanceled,
		CancelledAt: order.UpdatedAt,
	}, nil
}

// GetOrder retrieves an order by ID
func (r *SpotOrderRouter) GetOrder(orderID string) (*RealSpotOrder, bool) {
	r.svc.mu.RLock()
	defer r.svc.mu.RUnlock()
	o, exists := r.svc.orders[orderID]
	return o, exists
}

// GetOrdersByUser returns all orders for a given user
func (r *SpotOrderRouter) GetOrdersByUser(userID string) []*RealSpotOrder {
	r.svc.mu.RLock()
	defer r.svc.mu.RUnlock()

	var result []*RealSpotOrder
	for _, o := range r.svc.orders {
		if o.UserID == userID {
			result = append(result, o)
		}
	}
	return result
}
