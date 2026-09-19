package main

import (
	"errors"
	"fmt"
	"time"
)

// MarketOrderSide represents BUY or SELL
type MarketOrderSide string

const (
	SideBuy  MarketOrderSide = "BUY"
	SideSell MarketOrderSide = "SELL"
)

// MarketOrderRequest defines an incoming spot market order
type MarketOrderRequest struct {
	OrderID         string          `json:"order_id"`
	UserID          string          `json:"user_id"`
	Symbol          string          `json:"symbol"`
	Side            MarketOrderSide `json:"side"`
	QuantityE8      uint64          `json:"quantity_e8"`
	MaxSlippageBps  uint32          `json:"max_slippage_bps"` // e.g., 100 = 1.00%
	EstimatedPriceE8 uint64         `json:"estimated_price_e8"`
}

// BookLevel represents a price level in the orderbook
type BookLevel struct {
	PriceE8    uint64
	QuantityE8 uint64
}

// MarketOrderExecutionResult holds fill details and slippage metrics
type MarketOrderExecutionResult struct {
	OrderID       string `json:"order_id"`
	FilledQtyE8   uint64 `json:"filled_qty_e8"`
	AveragePriceE8 uint64 `json:"average_price_e8"`
	TotalNotional uint64 `json:"total_notional"`
	SlippageBps   uint32 `json:"slippage_bps"`
	PlatformFeeE8 uint64 `json:"platform_fee_e8"` // Strictly 0.00%
	Status        string `json:"status"`
	ExecutedAt    time.Time `json:"executed_at"`
}

// ExecuteMarketOrder evaluates slippage bounds against current book levels and computes execution
func ExecuteMarketOrder(req MarketOrderRequest, availableDepth []BookLevel) (*MarketOrderExecutionResult, error) {
	if req.QuantityE8 == 0 {
		return nil, errors.New("quantity must be greater than zero")
	}
	if len(availableDepth) == 0 {
		return nil, errors.New("insufficient liquidity to satisfy market order")
	}

	remaining := req.QuantityE8
	var totalCost uint64
	var filledQty uint64

	for _, level := range availableDepth {
		if remaining == 0 {
			break
		}

		qtyToTake := level.QuantityE8
		if qtyToTake > remaining {
			qtyToTake = remaining
		}

		totalCost += (qtyToTake * level.PriceE8) / 1e8
		filledQty += qtyToTake
		remaining -= qtyToTake
	}

	if filledQty == 0 {
		return nil, errors.New("zero volume matched")
	}

	avgPrice := (totalCost * 1e8) / filledQty

	// Slippage calculation against pre-trade estimated price
	var slippageBps uint32
	if req.EstimatedPriceE8 > 0 {
		var diff uint64
		if req.Side == SideBuy && avgPrice > req.EstimatedPriceE8 {
			diff = avgPrice - req.EstimatedPriceE8
			slippageBps = uint32((diff * 10000) / req.EstimatedPriceE8)
		} else if req.Side == SideSell && avgPrice < req.EstimatedPriceE8 {
			diff = req.EstimatedPriceE8 - avgPrice
			slippageBps = uint32((diff * 10000) / req.EstimatedPriceE8)
		}

		if req.MaxSlippageBps > 0 && slippageBps > req.MaxSlippageBps {
			return nil, fmt.Errorf("slippage limit breached: %d bps > max %d bps", slippageBps, req.MaxSlippageBps)
		}
	}

	return &MarketOrderExecutionResult{
		OrderID:        req.OrderID,
		FilledQtyE8:    filledQty,
		AveragePriceE8: avgPrice,
		TotalNotional:  totalCost,
		SlippageBps:    slippageBps,
		PlatformFeeE8:  0, // Strictly 0.00% universal fee
		Status:         "FILLED",
		ExecutedAt:     time.Now().UTC(),
	}, nil
}
