package opensearch

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// MarketSurveillanceEngine detects fraudulent and manipulative trading patterns.
type MarketSurveillanceEngine struct {
	mu     sync.RWMutex
	alerts []*SurveillanceAlert
}

func NewMarketSurveillanceEngine() *MarketSurveillanceEngine {
	return &MarketSurveillanceEngine{
		alerts: make([]*SurveillanceAlert, 0),
	}
}

// AnalyzeTradeForWashTrading inspects a trade for wash trading.
// Wash trading pattern: buyer and seller are identical addresses, or close counterparties
// trading the same security at identical prices within short intervals with no genuine change in ownership.
func (s *MarketSurveillanceEngine) AnalyzeTradeForWashTrading(ctx context.Context, trade *TradeDocument) *SurveillanceAlert {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Direct self-trade (identical buyer and seller address)
	if trade.BuyerAddress != "" && trade.BuyerAddress == trade.SellerAddress {
		alert := &SurveillanceAlert{
			AlertID:         fmt.Sprintf("alert-wash-%s", trade.TradeID),
			AlertType:       AlertWashTrading,
			Severity:        SeverityCritical,
			ISIN:            trade.ISIN,
			Symbol:          trade.Symbol,
			Participants:    []string{trade.BuyerAddress},
			ConfidenceScore: 1.0,
			Description:     "Direct self-trade detected: buyer and seller blockchain addresses are identical",
			Evidence: map[string]interface{}{
				"trade_id":         trade.TradeID,
				"price":            trade.PricePerUnitINR,
				"units":            trade.FractionalUnits,
				"buyer_order_id":   trade.BuyerOrderID,
				"seller_order_id":  trade.SellerOrderID,
				"matched_address":  trade.BuyerAddress,
			},
			Status:     "OPEN",
			DetectedAt: time.Now().UTC(),
		}
		s.alerts = append(s.alerts, alert)
		return alert
	}

	return nil
}

// AnalyzeOrdersForSpoofing detects rapid order placements and cancellations designed to spoof depth.
type OrderEvent struct {
	OrderID    string
	UserID     string
	Price      float64
	Quantity   float64
	Side       string
	PlacedAt   time.Time
	CanceledAt time.Time
	IsCanceled bool
}

func (s *MarketSurveillanceEngine) AnalyzeOrdersForSpoofing(
	ctx context.Context,
	isin string,
	events []OrderEvent,
	window time.Duration,
) *SurveillanceAlert {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(events) < 3 {
		return nil
	}

	// Detect if a single user placed large orders and canceled >80% within the window
	userCancels := make(map[string]int)
	userTotal := make(map[string]int)

	for _, ev := range events {
		userTotal[ev.UserID]++
		if ev.IsCanceled && ev.CanceledAt.Sub(ev.PlacedAt) <= window {
			userCancels[ev.UserID]++
		}
	}

	for userID, cancelCount := range userCancels {
		total := userTotal[userID]
		cancelRatio := float64(cancelCount) / float64(total)

		if total >= 3 && cancelRatio >= 0.8 {
			alert := &SurveillanceAlert{
				AlertID:         fmt.Sprintf("alert-spoof-%s-%d", userID, time.Now().UnixNano()),
				AlertType:       AlertSpoofing,
				Severity:        SeverityHigh,
				ISIN:            isin,
				Participants:    []string{userID},
				ConfidenceScore: math.Min(1.0, cancelRatio),
				Description:     fmt.Sprintf("Spoofing pattern: User %s canceled %d of %d orders within %v window", userID, cancelCount, total, window),
				Evidence: map[string]interface{}{
					"user_id":      userID,
					"total_orders": total,
					"cancels":      cancelCount,
					"cancel_ratio": cancelRatio,
				},
				Status:     "OPEN",
				DetectedAt: time.Now().UTC(),
			}
			s.alerts = append(s.alerts, alert)
			return alert
		}
	}

	return nil
}

// GetAlerts returns all surveillance alerts.
func (s *MarketSurveillanceEngine) GetAlerts() []*SurveillanceAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]*SurveillanceAlert, len(s.alerts))
	copy(res, s.alerts)
	return res
}
