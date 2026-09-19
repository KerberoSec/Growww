package main

import (
	"time"
)

type MarketSessionState string

const (
	SessionOpen        MarketSessionState = "OPEN"
	SessionPreOpen     MarketSessionState = "PRE_OPEN"
	SessionPostClose   MarketSessionState = "POST_CLOSE"
	SessionClosed      MarketSessionState = "CLOSED"
	SessionContinuous  MarketSessionState = "CONTINUOUS_24_7"
)

type SessionStatus struct {
	AssetClass  string             `json:"asset_class"`
	Symbol      string             `json:"symbol"`
	State       MarketSessionState `json:"state"`
	OrderEntry  bool               `json:"order_entry_allowed"`
	NextStateAt string             `json:"next_state_at"`
}

// EvaluateMarketSession evaluates whether an order is allowed based on market hours
func EvaluateMarketSession(assetClass string, nowUTC time.Time) SessionStatus {
	// Crypto trades 24/7/365 continuously
	if assetClass == "CRYPTO" || assetClass == "VDA" {
		return SessionStatus{
			AssetClass:  assetClass,
			State:       SessionContinuous,
			OrderEntry:  true,
			NextStateAt: "NEVER",
		}
	}

	// Indian Equity Market: 09:15 - 15:30 IST (Monday - Friday)
	// Convert UTC to IST (+5:30)
	loc, _ := time.LoadLocation("Asia/Kolkata")
	ist := nowUTC.In(loc)

	// Weekend check
	if ist.Weekday() == time.Saturday || ist.Weekday() == time.Sunday {
		return SessionStatus{
			AssetClass:  assetClass,
			State:       SessionClosed,
			OrderEntry:  false, // Off-market orders (AMO) handled separately
			NextStateAt: "MONDAY_0900_IST",
		}
	}

	hour := ist.Hour()
	min := ist.Minute()
	totalMin := hour*60 + min

	// 09:00 - 09:15 IST Pre-Open Call Auction
	if totalMin >= 540 && totalMin < 555 {
		return SessionStatus{
			AssetClass:  assetClass,
			State:       SessionPreOpen,
			OrderEntry:  true,
			NextStateAt: "09:15_IST",
		}
	}

	// 09:15 - 15:30 IST Continuous Trading
	if totalMin >= 555 && totalMin <= 930 {
		return SessionStatus{
			AssetClass:  assetClass,
			State:       SessionOpen,
			OrderEntry:  true,
			NextStateAt: "15:30_IST",
		}
	}

	return SessionStatus{
		AssetClass:  assetClass,
		State:       SessionClosed,
		OrderEntry:  false,
		NextStateAt: "NEXT_TRADING_DAY_0900_IST",
	}
}
