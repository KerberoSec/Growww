package main

import (
	"sync"
	"time"
)

type WatchdogEscrowTrade struct {
	TradeID     string
	BuyerID     string
	SellerID    string
	Status      string // CREATED, FIAT_MARKED_PAID, DISPUTED, CANCELLED, RELEASED
	CreatedAt   time.Time
	PaymentTTL  time.Duration
}

type P2PTimeoutWatchdog struct {
	mu     sync.Mutex
	trades map[string]*WatchdogEscrowTrade
}

func NewP2PTimeoutWatchdog() *P2PTimeoutWatchdog {
	return &P2PTimeoutWatchdog{trades: make(map[string]*WatchdogEscrowTrade)}
}

func (w *P2PTimeoutWatchdog) RegisterTrade(t *WatchdogEscrowTrade) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.trades[t.TradeID] = t
}

func (w *P2PTimeoutWatchdog) SweepExpiredTrades(now time.Time) []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	var cancelled []string

	for id, trade := range w.trades {
		if trade.Status == "CREATED" {
			expiry := trade.CreatedAt.Add(trade.PaymentTTL)
			if now.After(expiry) {
				trade.Status = "CANCELLED"
				cancelled = append(cancelled, id)
			}
		}
	}
	return cancelled
}
