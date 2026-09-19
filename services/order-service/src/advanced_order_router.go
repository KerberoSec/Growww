package main

import (
	"errors"
	"fmt"
	"time"
)

// AlgoOrderType identifies the algorithmic source of a child order
type AlgoOrderType string

const (
	AlgoTypeTWAP         AlgoOrderType = "TWAP"
	AlgoTypeVWAP         AlgoOrderType = "VWAP"
	AlgoTypeIceberg      AlgoOrderType = "ICEBERG"
	AlgoTypeTrailingStop AlgoOrderType = "TRAILING_STOP"
	AlgoTypeOCO          AlgoOrderType = "OCO"
)

// ChildOrderRequest represents an algorithmic child slice routed to the core exchange
type ChildOrderRequest struct {
	ChildOrderID   string          `json:"child_order_id"`
	ParentOrderID  string          `json:"parent_order_id"`
	AlgoType       AlgoOrderType   `json:"algo_type"`
	UserID         string          `json:"user_id"`
	Symbol         string          `json:"symbol"`
	Side           MarketOrderSide `json:"side"`
	OrderType      string          `json:"order_type"` // "LIMIT" or "MARKET"
	PriceE8        uint64          `json:"price_e8"`
	QuantityE8     uint64          `json:"quantity_e8"`
	PostOnly       bool            `json:"post_only"`
	TIF            TimeInForce     `json:"tif"`
	MaxSlippageBps uint32          `json:"max_slippage_bps"`
	LimitPriceCapE8 uint64         `json:"limit_price_cap_e8"`
}

// ChildOrderRoutingResult contains the outcome of an ingested algorithmic child order
type ChildOrderRoutingResult struct {
	ChildOrderID   string    `json:"child_order_id"`
	ParentOrderID  string    `json:"parent_order_id"`
	AlgoType       AlgoOrderType `json:"algo_type"`
	Status         string    `json:"status"` // "RESTING", "FILLED", "REJECTED"
	FilledQtyE8    uint64    `json:"filled_qty_e8"`
	AveragePriceE8 uint64    `json:"average_price_e8"`
	TotalNotionalE8 uint64   `json:"total_notional_e8"`
	PlatformFeeE8  uint64    `json:"platform_fee_e8"` // Strictly 0.00%
	ExecutedAt     time.Time `json:"executed_at"`
	RejectReason   string    `json:"reject_reason,omitempty"`
}

// RouteChildOrder processes an incoming algorithmic child order through pre-trade risk and book execution
func RouteChildOrder(req ChildOrderRequest, bestBidPriceE8, bestAskPriceE8 uint64, availableDepth []BookLevel) (*ChildOrderRoutingResult, error) {
	if req.ChildOrderID == "" || req.ParentOrderID == "" {
		return nil, errors.New("child_order_id and parent_order_id are required")
	}
	if req.QuantityE8 == 0 {
		return nil, errors.New("order quantity must be greater than zero")
	}

	// 1. Validate SEBI operating band collar (+/- 5% of NBBO mid)
	if bestBidPriceE8 > 0 && bestAskPriceE8 > 0 {
		nbboMid := (bestBidPriceE8 + bestAskPriceE8) / 2
		band := (nbboMid * 500) / 10000 // 5.00% band
		minCollar := nbboMid - band
		maxCollar := nbboMid + band

		if req.OrderType == "LIMIT" {
			if req.Side == SideBuy && req.PriceE8 > maxCollar {
				return &ChildOrderRoutingResult{
					ChildOrderID:  req.ChildOrderID,
					ParentOrderID: req.ParentOrderID,
					AlgoType:      req.AlgoType,
					Status:        "REJECTED",
					RejectReason:  fmt.Sprintf("buy limit %d breaches upper NBBO collar %d", req.PriceE8, maxCollar),
					ExecutedAt:    time.Now().UTC(),
				}, nil
			}
			if req.Side == SideSell && req.PriceE8 < minCollar {
				return &ChildOrderRoutingResult{
					ChildOrderID:  req.ChildOrderID,
					ParentOrderID: req.ParentOrderID,
					AlgoType:      req.AlgoType,
					Status:        "REJECTED",
					RejectReason:  fmt.Sprintf("sell limit %d breaches lower NBBO collar %d", req.PriceE8, minCollar),
					ExecutedAt:    time.Now().UTC(),
				}, nil
			}
		}
	}

	// 2. Validate user limit price cap collar
	if req.LimitPriceCapE8 > 0 {
		if req.Side == SideBuy && req.PriceE8 > req.LimitPriceCapE8 {
			return &ChildOrderRoutingResult{
				ChildOrderID:  req.ChildOrderID,
				ParentOrderID: req.ParentOrderID,
				AlgoType:      req.AlgoType,
				Status:        "REJECTED",
				RejectReason:  fmt.Sprintf("price %d breaches user limit buy cap %d", req.PriceE8, req.LimitPriceCapE8),
				ExecutedAt:    time.Now().UTC(),
			}, nil
		}
		if req.Side == SideSell && req.PriceE8 < req.LimitPriceCapE8 {
			return &ChildOrderRoutingResult{
				ChildOrderID:  req.ChildOrderID,
				ParentOrderID: req.ParentOrderID,
				AlgoType:      req.AlgoType,
				Status:        "REJECTED",
				RejectReason:  fmt.Sprintf("price %d breaches user limit sell cap %d", req.PriceE8, req.LimitPriceCapE8),
				ExecutedAt:    time.Now().UTC(),
			}, nil
		}
	}

	// 3. Execution path based on order type
	if req.OrderType == "MARKET" {
		marketReq := MarketOrderRequest{
			OrderID:          req.ChildOrderID,
			UserID:           req.UserID,
			Symbol:           req.Symbol,
			Side:             req.Side,
			QuantityE8:       req.QuantityE8,
			MaxSlippageBps:   req.MaxSlippageBps,
			EstimatedPriceE8: req.PriceE8,
		}
		execResult, err := ExecuteMarketOrder(marketReq, availableDepth)
		if err != nil {
			return &ChildOrderRoutingResult{
				ChildOrderID:  req.ChildOrderID,
				ParentOrderID: req.ParentOrderID,
				AlgoType:      req.AlgoType,
				Status:        "REJECTED",
				RejectReason:  err.Error(),
				ExecutedAt:    time.Now().UTC(),
			}, nil
		}
		return &ChildOrderRoutingResult{
			ChildOrderID:    req.ChildOrderID,
			ParentOrderID:   req.ParentOrderID,
			AlgoType:        req.AlgoType,
			Status:          execResult.Status,
			FilledQtyE8:     execResult.FilledQtyE8,
			AveragePriceE8:  execResult.AveragePriceE8,
			TotalNotionalE8: execResult.TotalNotional,
			PlatformFeeE8:   0, // Strictly 0.00% universal fee
			ExecutedAt:      execResult.ExecutedAt,
		}, nil
	}

	// 4. LIMIT order execution path
	limitOrder := LimitOrder{
		OrderID:    req.ChildOrderID,
		UserID:     req.UserID,
		Symbol:     req.Symbol,
		Side:       req.Side,
		PriceE8:    req.PriceE8,
		QuantityE8: req.QuantityE8,
		PostOnly:   req.PostOnly,
		TIF:        req.TIF,
		CreatedAt:  time.Now().UTC(),
	}

	if err := ValidatePostOnly(limitOrder, bestBidPriceE8, bestAskPriceE8); err != nil {
		return &ChildOrderRoutingResult{
			ChildOrderID:  req.ChildOrderID,
			ParentOrderID: req.ParentOrderID,
			AlgoType:      req.AlgoType,
			Status:        "REJECTED",
			RejectReason:  err.Error(),
			ExecutedAt:    time.Now().UTC(),
		}, nil
	}

	// If no crossing condition, order rests on the order book
	return &ChildOrderRoutingResult{
		ChildOrderID:    req.ChildOrderID,
		ParentOrderID:   req.ParentOrderID,
		AlgoType:        req.AlgoType,
		Status:          "RESTING",
		FilledQtyE8:     0,
		AveragePriceE8:  req.PriceE8,
		TotalNotionalE8: 0,
		PlatformFeeE8:   0, // Strictly 0.00% universal fee
		ExecutedAt:      limitOrder.CreatedAt,
	}, nil
}
