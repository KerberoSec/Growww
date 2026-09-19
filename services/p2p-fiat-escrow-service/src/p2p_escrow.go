package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type EscrowState string

const (
	EscrowFunded    EscrowState = "FUNDED"
	EscrowFiatPaid  EscrowState = "FIAT_PAID"
	EscrowReleased  EscrowState = "RELEASED"
	EscrowDisputed  EscrowState = "DISPUTED"
	EscrowCancelled EscrowState = "CANCELLED"
)

type P2PEscrowRecord struct {
	TradeID         string      `json:"trade_id"`
	SellerID        string      `json:"seller_id"`
	BuyerID         string      `json:"buyer_id"`
	CryptoAmountE8  uint64      `json:"crypto_amount_e8"`
	FiatAmountINR   float64     `json:"fiat_amount_inr"`
	State           EscrowState `json:"state"`
	FiatRefNumber   string      `json:"fiat_ref_number"`
	ExpiryTimestamp time.Time   `json:"expiry_timestamp"`
	CreatedAt       time.Time   `json:"created_at"`
}

type P2PEscrowDesk struct {
	mu     sync.RWMutex
	trades map[string]*P2PEscrowRecord
}

func NewP2PEscrowDesk() *P2PEscrowDesk {
	return &P2PEscrowDesk{
		trades: make(map[string]*P2PEscrowRecord),
	}
}

func (d *P2PEscrowDesk) CreateEscrow(id, seller, buyer string, cryptoE8 uint64, fiatINR float64, windowMinutes int) *P2PEscrowRecord {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	rec := &P2PEscrowRecord{
		TradeID:         id,
		SellerID:        seller,
		BuyerID:         buyer,
		CryptoAmountE8:  cryptoE8,
		FiatAmountINR:   fiatINR,
		State:           EscrowFunded,
		ExpiryTimestamp: now.Add(time.Duration(windowMinutes) * time.Minute),
		CreatedAt:       now,
	}
	d.trades[id] = rec
	return rec
}

func (d *P2PEscrowDesk) MarkFiatPaid(id, buyerID, refNumber string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	rec, exists := d.trades[id]
	if !exists {
		return errors.New("escrow trade not found")
	}
	if rec.BuyerID != buyerID {
		return errors.New("only designated buyer can mark paid")
	}
	if rec.State != EscrowFunded {
		return fmt.Errorf("cannot mark paid in state %s", rec.State)
	}

	rec.State = EscrowFiatPaid
	rec.FiatRefNumber = refNumber
	return nil
}

func (d *P2PEscrowDesk) ReleaseEscrow(id, sellerID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	rec, exists := d.trades[id]
	if !exists {
		return errors.New("escrow trade not found")
	}
	if rec.SellerID != sellerID {
		return errors.New("only designated seller can release")
	}

	rec.State = EscrowReleased
	fmt.Printf("[P2P Escrow] Trade %s released successfully: credited %d to buyer %s\n",
		id, rec.CryptoAmountE8, rec.BuyerID)
	return nil
}
