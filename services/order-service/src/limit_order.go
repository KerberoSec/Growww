package main

import (
	"errors"
	"time"
)

// TimeInForce defines order execution lifespan
type TimeInForce string

const (
	TIF_GTC TimeInForce = "GTC" // Good 'Til Cancelled
	TIF_IOC TimeInForce = "IOC" // Immediate or Cancel
	TIF_FOK TimeInForce = "FOK" // Fill or Kill
)

// LimitOrder represents a standard or post-only spot limit order
type LimitOrder struct {
	OrderID     string          `json:"order_id"`
	UserID      string          `json:"user_id"`
	Symbol      string          `json:"symbol"`
	Side        MarketOrderSide `json:"side"`
	PriceE8     uint64          `json:"price_e8"`
	QuantityE8  uint64          `json:"quantity_e8"`
	PostOnly    bool            `json:"post_only"`
	TIF         TimeInForce     `json:"tif"`
	CreatedAt   time.Time       `json:"created_at"`
}

// ValidatePostOnly verifies that a post-only order does not match immediately
func ValidatePostOnly(order LimitOrder, bestBidPriceE8, bestAskPriceE8 uint64) error {
	if !order.PostOnly {
		return nil
	}

	if order.Side == SideBuy && bestAskPriceE8 > 0 && order.PriceE8 >= bestAskPriceE8 {
		return errors.New("post-only buy limit order crosses spread and would take liquidity")
	}

	if order.Side == SideSell && bestBidPriceE8 > 0 && order.PriceE8 <= bestBidPriceE8 {
		return errors.New("post-only sell limit order crosses spread and would take liquidity")
	}

	return nil
}

// EvaluateTIF verifies whether an order satisfies its Time-In-Force constraints
func EvaluateTIF(tif TimeInForce, matchedQtyE8, targetQtyE8 uint64) (bool, string) {
	switch tif {
	case TIF_FOK:
		if matchedQtyE8 < targetQtyE8 {
			return false, "FOK_REJECTED_INCOMPLETE_FILL"
		}
		return true, "FILLED"
	case TIF_IOC:
		if matchedQtyE8 == 0 {
			return false, "IOC_CANCELLED_ZERO_FILL"
		}
		return true, "PARTIALLY_FILLED_OR_FILLED"
	default:
		return true, "RESTING_OR_FILLED"
	}
}
