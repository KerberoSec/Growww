package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type SpotOrderStatus string

const (
	OrderNew             SpotOrderStatus = "NEW"
	OrderPartiallyFilled SpotOrderStatus = "PARTIALLY_FILLED"
	OrderFilled          SpotOrderStatus = "FILLED"
	OrderCanceled        SpotOrderStatus = "CANCELED"
	OrderRejected        SpotOrderStatus = "REJECTED"
)

type RealSpotOrder struct {
	OrderID         string          `json:"order_id"`
	UserID          string          `json:"user_id"`
	Symbol          string          `json:"symbol"` // BTC/USDT
	Side            string          `json:"side"`   // BUY or SELL
	PriceE8         uint64          `json:"price_e8"`
	QuantityE8      uint64          `json:"quantity_e8"`
	ExecutedQtyE8   uint64          `json:"executed_qty_e8"`
	CumulativeFeeE8 uint64          `json:"cumulative_fee_e8"` // 0.00% universal policy
	Status          SpotOrderStatus `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type SpotOrderLifecycleService struct {
	mu     sync.RWMutex
	orders map[string]*RealSpotOrder
}

func NewSpotOrderLifecycleService() *SpotOrderLifecycleService {
	return &SpotOrderLifecycleService{
		orders: make(map[string]*RealSpotOrder),
	}
}

// SubmitOrder registers a new spot order into the active execution pipeline
func (s *SpotOrderLifecycleService) SubmitOrder(orderID, userID, symbol, side string, priceE8, qtyE8 uint64) (*RealSpotOrder, error) {
	if symbol != "BTC/USDT" {
		return nil, errors.New("unsupported spot symbol: only BTC/USDT active on this rail")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := &RealSpotOrder{
		OrderID:         orderID,
		UserID:          userID,
		Symbol:          symbol,
		Side:            side,
		PriceE8:         priceE8,
		QuantityE8:      qtyE8,
		ExecutedQtyE8:   0,
		CumulativeFeeE8: 0, // 0.00% launch policy
		Status:          OrderNew,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}

	s.orders[orderID] = order
	fmt.Printf("[Spot Order Service] Order %s accepted: %s %d Sats @ %d (0.00%% fee)\n",
		orderID, side, qtyE8, priceE8)
	return order, nil
}

// RecordFill updates fill progress and completes order when fully matched
func (s *SpotOrderLifecycleService) RecordFill(orderID string, fillQtyE8 uint64) (*RealSpotOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		return nil, errors.New("order not found")
	}

	order.ExecutedQtyE8 += fillQtyE8
	order.UpdatedAt = time.Now().UTC()

	if order.ExecutedQtyE8 >= order.QuantityE8 {
		order.Status = OrderFilled
	} else {
		order.Status = OrderPartiallyFilled
	}

	return order, nil
}
