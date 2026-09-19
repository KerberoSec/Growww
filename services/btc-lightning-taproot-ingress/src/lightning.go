package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type LightningInvoice struct {
	PaymentHash string
	Amount      uint64 // satoshis
	Memo        string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Status      string // PENDING, SETTLED, EXPIRED
	UserID      string
}

type LightningService struct {
	mu       sync.Mutex
	invoices map[string]*LightningInvoice
}

func NewLightningService() *LightningService {
	return &LightningService{invoices: make(map[string]*LightningInvoice)}
}

func (s *LightningService) CreateInvoice(userID string, amountSats uint64, memo string, expirySec int) (*LightningInvoice, error) {
	if amountSats == 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if expirySec <= 0 {
		expirySec = 3600
	}
	now := time.Now().UTC()
	preimage := fmt.Sprintf("%s:%d:%d", userID, amountSats, now.UnixNano())
	hash := sha256.Sum256([]byte(preimage))
	paymentHash := hex.EncodeToString(hash[:])

	inv := &LightningInvoice{
		PaymentHash: paymentHash, Amount: amountSats, Memo: memo,
		CreatedAt: now, ExpiresAt: now.Add(time.Duration(expirySec) * time.Second),
		Status: "PENDING", UserID: userID,
	}
	s.mu.Lock()
	s.invoices[paymentHash] = inv
	s.mu.Unlock()
	return inv, nil
}

func (s *LightningService) SettleInvoice(paymentHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invoices[paymentHash]
	if !ok {
		return fmt.Errorf("invoice not found")
	}
	if inv.Status != "PENDING" {
		return fmt.Errorf("invoice not in PENDING state: %s", inv.Status)
	}
	if time.Now().UTC().After(inv.ExpiresAt) {
		inv.Status = "EXPIRED"
		return fmt.Errorf("invoice has expired")
	}
	inv.Status = "SETTLED"
	return nil
}

func (s *LightningService) GetInvoice(paymentHash string) (*LightningInvoice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invoices[paymentHash]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return inv, nil
}

func DeriveTaprootAddress(pubKeyHex string) string {
	h := sha256.Sum256([]byte("taproot-tweak:" + pubKeyHex))
	return "bc1p" + hex.EncodeToString(h[:20])
}
