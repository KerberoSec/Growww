// Package p2pescrow provides a production-grade P2P fiat-to-crypto escrow
// service with full trade lifecycle, dispute resolution, payment confirmation,
// auto-cancel on expiry, and seller/buyer reputation tracking.
package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Enums & Data Structures
// ---------------------------------------------------------------------------

// EscrowState represents the lifecycle state of a P2P trade.
type EscrowState string

const (
	EscrowCreated   EscrowState = "CREATED"
	EscrowFunded    EscrowState = "FUNDED"
	EscrowFiatPaid  EscrowState = "FIAT_PAID"
	EscrowReleased  EscrowState = "RELEASED"
	EscrowDisputed  EscrowState = "DISPUTED"
	EscrowResolved  EscrowState = "RESOLVED"
	EscrowCancelled EscrowState = "CANCELLED"
	EscrowExpired   EscrowState = "EXPIRED"
)

// DisputeResolution captures the outcome of a dispute.
type DisputeResolution string

const (
	ResolutionNone          DisputeResolution = "NONE"
	ResolutionReleaseBuyer  DisputeResolution = "RELEASE_TO_BUYER"
	ResolutionRefundSeller  DisputeResolution = "REFUND_TO_SELLER"
	ResolutionSplitPartial  DisputeResolution = "SPLIT_PARTIAL"
)

// P2PEscrowRecord tracks the full lifecycle of a single P2P trade.
type P2PEscrowRecord struct {
	TradeID           string            `json:"trade_id"`
	SellerID          string            `json:"seller_id"`
	BuyerID           string            `json:"buyer_id"`
	CryptoAmountE8    uint64            `json:"crypto_amount_e8"`
	FiatAmountINR     float64           `json:"fiat_amount_inr"`
	State             EscrowState       `json:"state"`
	FiatRefNumber     string            `json:"fiat_ref_number"`
	PaymentMethod     string            `json:"payment_method"`
	ExpiryTimestamp   time.Time         `json:"expiry_timestamp"`
	CreatedAt         time.Time         `json:"created_at"`
	FiatPaidAt        time.Time         `json:"fiat_paid_at,omitempty"`
	ReleasedAt        time.Time         `json:"released_at,omitempty"`
	DisputeReason     string            `json:"dispute_reason,omitempty"`
	DisputeResult     DisputeResolution `json:"dispute_result,omitempty"`
	ArbitratorID      string            `json:"arbitrator_id,omitempty"`
}

// UserReputation tracks P2P trading reputation.
type UserReputation struct {
	UserID          string  `json:"user_id"`
	TotalTrades     int     `json:"total_trades"`
	CompletedTrades int     `json:"completed_trades"`
	DisputesRaised  int     `json:"disputes_raised"`
	DisputesLost    int     `json:"disputes_lost"`
	AvgReleaseTime  float64 `json:"avg_release_time_secs"` // avg seconds from FiatPaid to Released
	Score           float64 `json:"score"`                  // 0-100
}

// ---------------------------------------------------------------------------
// Engine
// ---------------------------------------------------------------------------

// P2PEscrowDesk manages the full P2P escrow lifecycle.
type P2PEscrowDesk struct {
	mu          sync.RWMutex
	trades      map[string]*P2PEscrowRecord
	reputations map[string]*UserReputation
}

// NewP2PEscrowDesk creates a new escrow desk.
func NewP2PEscrowDesk() *P2PEscrowDesk {
	return &P2PEscrowDesk{
		trades:      make(map[string]*P2PEscrowRecord),
		reputations: make(map[string]*UserReputation),
	}
}

// CreateEscrow initialises a new P2P escrow trade with crypto locked.
func (d *P2PEscrowDesk) CreateEscrow(
	id, seller, buyer string,
	cryptoE8 uint64, fiatINR float64,
	paymentMethod string,
	windowMinutes int,
) (*P2PEscrowRecord, error) {
	if id == "" || seller == "" || buyer == "" {
		return nil, errors.New("trade_id, seller, and buyer are required")
	}
	if seller == buyer {
		return nil, errors.New("seller and buyer cannot be the same user")
	}
	if cryptoE8 == 0 || fiatINR <= 0 {
		return nil, errors.New("amounts must be positive")
	}
	if windowMinutes <= 0 {
		return nil, errors.New("payment window must be positive")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.trades[id]; exists {
		return nil, fmt.Errorf("trade %s already exists", id)
	}

	now := time.Now().UTC()
	rec := &P2PEscrowRecord{
		TradeID:         id,
		SellerID:        seller,
		BuyerID:         buyer,
		CryptoAmountE8:  cryptoE8,
		FiatAmountINR:   fiatINR,
		State:           EscrowFunded,
		PaymentMethod:   paymentMethod,
		ExpiryTimestamp: now.Add(time.Duration(windowMinutes) * time.Minute),
		CreatedAt:       now,
		DisputeResult:   ResolutionNone,
	}
	d.trades[id] = rec
	d.ensureReputation(seller)
	d.ensureReputation(buyer)
	return rec, nil
}

// ConfirmPayment marks fiat as paid by the buyer with a reference number.
func (d *P2PEscrowDesk) ConfirmPayment(tradeID, buyerID, refNumber string) error {
	if refNumber == "" {
		return errors.New("payment reference number is required")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	rec, err := d.getTradeUnsafe(tradeID)
	if err != nil {
		return err
	}
	if rec.BuyerID != buyerID {
		return errors.New("only the designated buyer can confirm payment")
	}
	if rec.State != EscrowFunded {
		return fmt.Errorf("cannot confirm payment in state %s", rec.State)
	}
	now := time.Now().UTC()
	if now.After(rec.ExpiryTimestamp) {
		rec.State = EscrowExpired
		return errors.New("payment window has expired")
	}

	rec.State = EscrowFiatPaid
	rec.FiatRefNumber = refNumber
	rec.FiatPaidAt = now
	return nil
}

// ReleaseEscrow releases crypto to buyer after seller confirms fiat receipt.
func (d *P2PEscrowDesk) ReleaseEscrow(tradeID, sellerID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	rec, err := d.getTradeUnsafe(tradeID)
	if err != nil {
		return err
	}
	if rec.SellerID != sellerID {
		return errors.New("only the designated seller can release escrow")
	}
	if rec.State != EscrowFiatPaid {
		return fmt.Errorf("cannot release in state %s", rec.State)
	}

	rec.State = EscrowReleased
	rec.ReleasedAt = time.Now().UTC()

	// Update reputations
	d.recordCompletion(rec)
	return nil
}

// RaiseDispute opens a dispute on the trade.
func (d *P2PEscrowDesk) RaiseDispute(tradeID, raiserID, reason string) error {
	if reason == "" {
		return errors.New("dispute reason is required")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	rec, err := d.getTradeUnsafe(tradeID)
	if err != nil {
		return err
	}
	if raiserID != rec.BuyerID && raiserID != rec.SellerID {
		return errors.New("only trade participants can raise disputes")
	}
	if rec.State != EscrowFiatPaid && rec.State != EscrowFunded {
		return fmt.Errorf("cannot dispute in state %s", rec.State)
	}

	rec.State = EscrowDisputed
	rec.DisputeReason = reason

	if rep, ok := d.reputations[raiserID]; ok {
		rep.DisputesRaised++
	}
	return nil
}

// ResolveDispute resolves an open dispute via arbitration.
func (d *P2PEscrowDesk) ResolveDispute(
	tradeID, arbitratorID string,
	resolution DisputeResolution,
) error {
	if arbitratorID == "" {
		return errors.New("arbitrator ID is required")
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	rec, err := d.getTradeUnsafe(tradeID)
	if err != nil {
		return err
	}
	if rec.State != EscrowDisputed {
		return fmt.Errorf("trade %s is not in disputed state", tradeID)
	}

	rec.State = EscrowResolved
	rec.DisputeResult = resolution
	rec.ArbitratorID = arbitratorID

	// Track disputes lost
	switch resolution {
	case ResolutionReleaseBuyer:
		if rep, ok := d.reputations[rec.SellerID]; ok {
			rep.DisputesLost++
			rep.CompletedTrades++
		}
		if rep, ok := d.reputations[rec.BuyerID]; ok {
			rep.CompletedTrades++
		}
	case ResolutionRefundSeller:
		if rep, ok := d.reputations[rec.BuyerID]; ok {
			rep.DisputesLost++
		}
	}

	d.recalculateScore(rec.SellerID)
	d.recalculateScore(rec.BuyerID)
	return nil
}

// CancelTrade cancels a trade (only if still FUNDED, i.e. no payment yet).
func (d *P2PEscrowDesk) CancelTrade(tradeID, userID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	rec, err := d.getTradeUnsafe(tradeID)
	if err != nil {
		return err
	}
	if rec.State != EscrowFunded {
		return fmt.Errorf("cannot cancel trade in state %s", rec.State)
	}
	if userID != rec.SellerID && userID != rec.BuyerID {
		return errors.New("only trade participants can cancel")
	}

	rec.State = EscrowCancelled
	return nil
}

// ExpireStale marks all funded trades past their window as expired.
func (d *P2PEscrowDesk) ExpireStale(now time.Time) []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	var expired []string
	for id, rec := range d.trades {
		if rec.State == EscrowFunded && now.After(rec.ExpiryTimestamp) {
			rec.State = EscrowExpired
			expired = append(expired, id)
		}
	}
	return expired
}

// GetReputation returns the reputation for a user.
func (d *P2PEscrowDesk) GetReputation(userID string) (*UserReputation, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rep, ok := d.reputations[userID]
	if !ok {
		return nil, fmt.Errorf("no reputation found for user %s", userID)
	}
	return rep, nil
}

// GetTrade returns a trade record.
func (d *P2PEscrowDesk) GetTrade(tradeID string) (*P2PEscrowRecord, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	rec, ok := d.trades[tradeID]
	if !ok {
		return nil, fmt.Errorf("trade %s not found", tradeID)
	}
	return rec, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func (d *P2PEscrowDesk) getTradeUnsafe(tradeID string) (*P2PEscrowRecord, error) {
	rec, ok := d.trades[tradeID]
	if !ok {
		return nil, fmt.Errorf("trade %s not found", tradeID)
	}
	return rec, nil
}

func (d *P2PEscrowDesk) ensureReputation(userID string) {
	if _, ok := d.reputations[userID]; !ok {
		d.reputations[userID] = &UserReputation{
			UserID: userID,
			Score:  50.0, // neutral starting score
		}
	}
}

func (d *P2PEscrowDesk) recordCompletion(rec *P2PEscrowRecord) {
	releaseSecs := rec.ReleasedAt.Sub(rec.FiatPaidAt).Seconds()

	for _, uid := range []string{rec.SellerID, rec.BuyerID} {
		if rep, ok := d.reputations[uid]; ok {
			rep.TotalTrades++
			rep.CompletedTrades++
			// Running average of release time (seller side metric but tracked for both)
			n := float64(rep.CompletedTrades)
			rep.AvgReleaseTime = rep.AvgReleaseTime*(n-1)/n + releaseSecs/n
		}
	}

	d.recalculateScore(rec.SellerID)
	d.recalculateScore(rec.BuyerID)
}

func (d *P2PEscrowDesk) recalculateScore(userID string) {
	rep, ok := d.reputations[userID]
	if !ok {
		return
	}

	if rep.TotalTrades == 0 {
		rep.Score = 50.0
		return
	}

	completionRate := float64(rep.CompletedTrades) / float64(rep.TotalTrades)
	disputePenalty := float64(rep.DisputesLost) * 5.0
	timePenalty := 0.0
	if rep.AvgReleaseTime > 600 { // > 10 min average release time
		timePenalty = math.Min((rep.AvgReleaseTime-600)/60.0, 15.0) // max 15 pt penalty
	}

	score := completionRate*100.0 - disputePenalty - timePenalty
	rep.Score = math.Max(0.0, math.Min(100.0, score))
}
