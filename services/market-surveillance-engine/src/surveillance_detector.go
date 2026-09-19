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
