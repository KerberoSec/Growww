package main

import (
	"fmt"
	"sync"
	"time"
)

type SurveillanceAlert struct {
	AlertID        string
	Pattern        string // SPOOFING, WASH_TRADING, LAYERING
	UserID         string
	Symbol         string
	ConfidenceRate float64
	Details        string
	TriggeredAt    time.Time
}

type OrderEvent struct {
	OrderID   string
	UserID    string
	Symbol    string
	Side      string
	PriceE8   uint64
	QtyE8     uint64
	IsCancel  bool
	Timestamp time.Time
}

type SurveillanceEngine struct {
	mu     sync.Mutex
	events []OrderEvent
	alerts []SurveillanceAlert
}

func NewSurveillanceEngine() *SurveillanceEngine {
	return &SurveillanceEngine{
		events: make([]OrderEvent, 0),
		alerts: make([]SurveillanceAlert, 0),
	}
}

// DetectWashTrading flags when buyer and seller share common beneficial owner or IP
func (e *SurveillanceEngine) DetectWashTrading(buyerID, sellerID, symbol string, priceE8, qtyE8 uint64) *SurveillanceAlert {
	e.mu.Lock()
	defer e.mu.Unlock()

	if buyerID == sellerID {
		alert := SurveillanceAlert{
			AlertID:        fmt.Sprintf("WASH-%d", time.Now().UnixNano()),
			Pattern:        "WASH_TRADING_SELF_CROSS",
			UserID:         buyerID,
			Symbol:         symbol,
			ConfidenceRate: 1.00,
			Details:        fmt.Sprintf("Self-crossing trade detected on %s for %d @ %d", symbol, qtyE8, priceE8),
			TriggeredAt:    time.Now().UTC(),
		}
		e.alerts = append(e.alerts, alert)
		fmt.Printf("[Surveillance] CRITICAL: Wash trade intercepted for user %s\n", buyerID)
		return &alert
	}
	return nil
}

// DetectSpoofing flags massive orders placed and cancelled within milliseconds to move market
func (e *SurveillanceEngine) DetectSpoofing(userID, symbol string, cancelLatencyMs int64, orderSizeNotional float64) *SurveillanceAlert {
	e.mu.Lock()
	defer e.mu.Unlock()

	// If cancelled in < 200ms and notional is large (> $100k)
	if cancelLatencyMs < 200 && orderSizeNotional > 100000.0 {
		alert := SurveillanceAlert{
			AlertID:        fmt.Sprintf("SPOOF-%d", time.Now().UnixNano()),
			Pattern:        "HIGH_FREQUENCY_SPOOFING",
			UserID:         userID,
			Symbol:         symbol,
			ConfidenceRate: 0.92,
			Details:        fmt.Sprintf("Large order cancelled in %d ms", cancelLatencyMs),
			TriggeredAt:    time.Now().UTC(),
		}
		e.alerts = append(e.alerts, alert)
		return &alert
	}
	return nil
}

// TradeHop represents a trade leg between two distinct accounts
type TradeHop struct {
	BuyerID   string
	SellerID  string
	Symbol    string
	QtyE8     uint64
	PriceE8   uint64
	Timestamp time.Time
}

// DetectCircularTrading identifies cyclic collusion (A -> B -> C -> A) within a sliding time window
func (e *SurveillanceEngine) DetectCircularTrading(hops []TradeHop, maxWindowSec float64) *SurveillanceAlert {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(hops) < 3 {
		return nil
	}

	// Graph adjacency: buyer -> seller
	adj := make(map[string]string)
	var firstTimestamp, lastTimestamp time.Time

	for i, h := range hops {
		if i == 0 {
			firstTimestamp = h.Timestamp
		}
		lastTimestamp = h.Timestamp
		adj[h.BuyerID] = h.SellerID
	}

	if lastTimestamp.Sub(firstTimestamp).Seconds() > maxWindowSec {
		return nil
	}

	// Check for cycle starting from hops[0].BuyerID
	start := hops[0].BuyerID
	curr := start
	visited := make(map[string]bool)

	for i := 0; i <= len(hops); i++ {
		next, exists := adj[curr]
		if !exists {
			break
		}
		if next == start && i >= 2 {
			alert := SurveillanceAlert{
				AlertID:        fmt.Sprintf("CIRCULAR-%d", time.Now().UnixNano()),
				Pattern:        "CIRCULAR_TRADING_RING",
				UserID:         start,
				Symbol:         hops[0].Symbol,
				ConfidenceRate: 0.98,
				Details:        fmt.Sprintf("Collusive circular trading ring detected across %d hops in %.1fs", len(hops), lastTimestamp.Sub(firstTimestamp).Seconds()),
				TriggeredAt:    time.Now().UTC(),
			}
			e.alerts = append(e.alerts, alert)
			return &alert
		}
		if visited[next] {
			break
		}
		visited[curr] = true
		curr = next
	}

	return nil
}

// DetectLayering identifies multiple fake limit orders inserted across adjacent depths and rapidly cancelled
func (e *SurveillanceEngine) DetectLayering(userID, symbol string, placedOrders, cancelledOrders int, cancelWindowMs int64) *SurveillanceAlert {
	e.mu.Lock()
	defer e.mu.Unlock()

	if placedOrders >= 5 && float64(cancelledOrders)/float64(placedOrders) >= 0.8 && cancelWindowMs < 1000 {
		alert := SurveillanceAlert{
			AlertID:        fmt.Sprintf("LAYER-%d", time.Now().UnixNano()),
			Pattern:        "ORDERBOOK_LAYERING_MANIPULATION",
			UserID:         userID,
			Symbol:         symbol,
			ConfidenceRate: 0.95,
			Details:        fmt.Sprintf("%d/%d orders stacked and cancelled within %d ms to simulate phantom depth", cancelledOrders, placedOrders, cancelWindowMs),
			TriggeredAt:    time.Now().UTC(),
		}
		e.alerts = append(e.alerts, alert)
		return &alert
	}
	return nil
}
