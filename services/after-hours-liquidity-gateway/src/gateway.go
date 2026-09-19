// Package afterhours implements an extended trading hours liquidity gateway for NBSE.
// It manages pre-market and post-market trading sessions with wider spreads
// and specialized order routing logic.
package afterhours

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// SessionType defines the trading session type.
type SessionType string

const (
	SessionPreMarket  SessionType = "PRE_MARKET"
	SessionRegular    SessionType = "REGULAR"
	SessionPostMarket SessionType = "POST_MARKET"
	SessionClosed     SessionType = "CLOSED"
)

// OrderSide represents buy or sell.
type OrderSide string

const (
	SideBuy  OrderSide = "BUY"
	SideSell OrderSide = "SELL"
)

// OrderType defines the order type.
type OrderType string

const (
	OrderLimit  OrderType = "LIMIT"
	OrderMarket OrderType = "MARKET"
)

// SessionConfig defines parameters for an extended hours session.
type SessionConfig struct {
	Type           SessionType `json:"type"`
	StartTime      string      `json:"start_time"`      // HH:MM in IST
	EndTime        string      `json:"end_time"`         // HH:MM in IST
	SpreadMultiplier float64   `json:"spread_multiplier"` // e.g. 2.0x for wider spreads
	MaxOrderSizeINR  float64   `json:"max_order_size_inr"`
	AllowMarketOrders bool     `json:"allow_market_orders"`
}

// Order represents an after-hours order.
type Order struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Symbol      string    `json:"symbol"`
	Side        OrderSide `json:"side"`
	Type        OrderType `json:"type"`
	Price       float64   `json:"price"`
	Quantity    float64   `json:"quantity"`
	FilledQty   float64   `json:"filled_qty"`
	Session     SessionType `json:"session"`
	Status      string    `json:"status"` // PENDING, ROUTED, FILLED, REJECTED, CANCELLED
	SubmittedAt time.Time `json:"submitted_at"`
	RoutedAt    time.Time `json:"routed_at,omitempty"`
	Reason      string    `json:"reason,omitempty"` // rejection reason
}

// SpreadQuote represents the wider spread quote for after-hours.
type SpreadQuote struct {
	Symbol      string  `json:"symbol"`
	BidPrice    float64 `json:"bid_price"`
	AskPrice    float64 `json:"ask_price"`
	Spread      float64 `json:"spread"`
	Multiplier  float64 `json:"multiplier"`
	Session     SessionType `json:"session"`
}

// LiquidityGateway manages extended hours trading sessions.
type LiquidityGateway struct {
	mu           sync.RWMutex
	sessions     map[SessionType]*SessionConfig
	activeSession SessionType
	orders       []*Order
	orderIndex   map[string]*Order // orderID -> order

	// Reference prices from regular session close
	refPrices map[string]float64 // symbol -> last regular close price
}

// NewLiquidityGateway creates a new gateway with default session configs.
func NewLiquidityGateway() *LiquidityGateway {
	gw := &LiquidityGateway{
		sessions: map[SessionType]*SessionConfig{
			SessionPreMarket: {
				Type:             SessionPreMarket,
				StartTime:        "07:00",
				EndTime:          "09:00",
				SpreadMultiplier: 2.0,
				MaxOrderSizeINR:  5_000_000,
				AllowMarketOrders: false,
			},
			SessionPostMarket: {
				Type:             SessionPostMarket,
				StartTime:        "15:30",
				EndTime:          "17:00",
				SpreadMultiplier: 1.5,
				MaxOrderSizeINR:  10_000_000,
				AllowMarketOrders: false,
			},
		},
		activeSession: SessionClosed,
		orders:        make([]*Order, 0),
		orderIndex:    make(map[string]*Order),
		refPrices:     make(map[string]float64),
	}
	return gw
}

// SetReferencePrice sets the closing price from the regular session for a symbol.
func (gw *LiquidityGateway) SetReferencePrice(symbol string, price float64) error {
	if price <= 0 {
		return errors.New("reference price must be positive")
	}
	gw.mu.Lock()
	defer gw.mu.Unlock()
	gw.refPrices[symbol] = price
	return nil
}

// OpenSession activates an extended trading session.
func (gw *LiquidityGateway) OpenSession(sessionType SessionType) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	if sessionType == SessionRegular || sessionType == SessionClosed {
		return fmt.Errorf("cannot open session type: %s", sessionType)
	}
	if _, ok := gw.sessions[sessionType]; !ok {
		return fmt.Errorf("unknown session type: %s", sessionType)
	}
	gw.activeSession = sessionType
	return nil
}

// CloseSession deactivates the current session.
func (gw *LiquidityGateway) CloseSession() {
	gw.mu.Lock()
	defer gw.mu.Unlock()
	gw.activeSession = SessionClosed
}

// ActiveSession returns the currently active session type.
func (gw *LiquidityGateway) ActiveSession() SessionType {
	gw.mu.RLock()
	defer gw.mu.RUnlock()
	return gw.activeSession
}

// GetSpreadQuote returns the widened bid/ask spread for a symbol in the current session.
func (gw *LiquidityGateway) GetSpreadQuote(symbol string) (*SpreadQuote, error) {
	gw.mu.RLock()
	defer gw.mu.RUnlock()

	if gw.activeSession == SessionClosed {
		return nil, errors.New("no active session")
	}

	refPrice, ok := gw.refPrices[symbol]
	if !ok {
		return nil, fmt.Errorf("no reference price for %s", symbol)
	}

	cfg := gw.sessions[gw.activeSession]
	// Base spread is 0.1% of reference price; multiplied by session config
	baseSpread := refPrice * 0.001
	widened := baseSpread * cfg.SpreadMultiplier
	halfSpread := widened / 2.0

	return &SpreadQuote{
		Symbol:     symbol,
		BidPrice:   refPrice - halfSpread,
		AskPrice:   refPrice + halfSpread,
		Spread:     widened,
		Multiplier: cfg.SpreadMultiplier,
		Session:    gw.activeSession,
	}, nil
}

// SubmitOrder validates and routes an order for the current extended session.
func (gw *LiquidityGateway) SubmitOrder(order *Order) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	if gw.activeSession == SessionClosed {
		return errors.New("no active session; orders not accepted")
	}
	if order.ID == "" || order.UserID == "" || order.Symbol == "" {
		return errors.New("order ID, user ID, and symbol are required")
	}
	if _, exists := gw.orderIndex[order.ID]; exists {
		return fmt.Errorf("duplicate order ID: %s", order.ID)
	}

	cfg := gw.sessions[gw.activeSession]

	// Validate market orders
	if order.Type == OrderMarket && !cfg.AllowMarketOrders {
		order.Status = "REJECTED"
		order.Reason = "market orders not allowed in extended session"
		order.Session = gw.activeSession
		order.SubmittedAt = time.Now().UTC()
		gw.orders = append(gw.orders, order)
		gw.orderIndex[order.ID] = order
		return fmt.Errorf("market orders not allowed in %s session", gw.activeSession)
	}

	// Validate order size
	orderValue := order.Price * order.Quantity
	if orderValue > cfg.MaxOrderSizeINR {
		order.Status = "REJECTED"
		order.Reason = fmt.Sprintf("order value %.2f exceeds max %.2f", orderValue, cfg.MaxOrderSizeINR)
		order.Session = gw.activeSession
		order.SubmittedAt = time.Now().UTC()
		gw.orders = append(gw.orders, order)
		gw.orderIndex[order.ID] = order
		return fmt.Errorf("order value exceeds session max of %.0f INR", cfg.MaxOrderSizeINR)
	}

	// Price validation: must be within spread bounds
	if refPrice, ok := gw.refPrices[order.Symbol]; ok {
		maxDeviation := refPrice * 0.05 // 5% max deviation from reference
		if order.Price > refPrice+maxDeviation || order.Price < refPrice-maxDeviation {
			order.Status = "REJECTED"
			order.Reason = "price outside acceptable range"
			order.Session = gw.activeSession
			order.SubmittedAt = time.Now().UTC()
			gw.orders = append(gw.orders, order)
			gw.orderIndex[order.ID] = order
			return errors.New("order price outside acceptable deviation from reference price")
		}
	}

	order.Session = gw.activeSession
	order.Status = "ROUTED"
	order.SubmittedAt = time.Now().UTC()
	order.RoutedAt = time.Now().UTC()
	gw.orders = append(gw.orders, order)
	gw.orderIndex[order.ID] = order
	return nil
}

// GetOrder retrieves an order by ID.
func (gw *LiquidityGateway) GetOrder(orderID string) (*Order, error) {
	gw.mu.RLock()
	defer gw.mu.RUnlock()
	o, ok := gw.orderIndex[orderID]
	if !ok {
		return nil, fmt.Errorf("order %s not found", orderID)
	}
	return o, nil
}

// CancelOrder cancels a pending/routed order.
func (gw *LiquidityGateway) CancelOrder(orderID string) error {
	gw.mu.Lock()
	defer gw.mu.Unlock()

	o, ok := gw.orderIndex[orderID]
	if !ok {
		return fmt.Errorf("order %s not found", orderID)
	}
	if o.Status != "PENDING" && o.Status != "ROUTED" {
		return fmt.Errorf("cannot cancel order in %s status", o.Status)
	}
	o.Status = "CANCELLED"
	return nil
}

// SessionOrderCount returns the number of orders in the current or specified session.
func (gw *LiquidityGateway) SessionOrderCount(session SessionType) int {
	gw.mu.RLock()
	defer gw.mu.RUnlock()
	count := 0
	for _, o := range gw.orders {
		if o.Session == session {
			count++
		}
	}
	return count
}
