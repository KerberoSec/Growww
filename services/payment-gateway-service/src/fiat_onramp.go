package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type PaymentRail string

const (
	RailUPI  PaymentRail = "UPI"
	RailIMPS PaymentRail = "IMPS"
	RailNEFT PaymentRail = "NEFT"
	RailRTGS PaymentRail = "RTGS"
)

type FiatDepositOrder struct {
	DepositID       string      `json:"deposit_id"`
	UserID          string      `json:"user_id"`
	AmountINR       float64     `json:"amount_inr"`
	EquivalentUSDT  float64     `json:"equivalent_usdt"`
	FXRateINRPerUSD float64     `json:"fx_rate_inr_per_usd"`
	Rail            PaymentRail `json:"rail"`
	BankRefNumber   string      `json:"bank_ref_number"`
	Status          string      `json:"status"`
	CreatedAt       time.Time   `json:"created_at"`
	CompletedAt     time.Time   `json:"completed_at"`
}

type FiatOnRampGateway struct {
	mu           sync.RWMutex
	deposits     map[string]*FiatDepositOrder
	activeFXRate float64 // e.g. 84.50 INR per USDT
}

func NewFiatOnRampGateway(fxRate float64) *FiatOnRampGateway {
	if fxRate <= 0 {
		fxRate = 84.50
	}
	return &FiatOnRampGateway{
		deposits:     make(map[string]*FiatDepositOrder),
		activeFXRate: fxRate,
	}
}

// InitiateDeposit creates a pending INR on-ramp order
func (g *FiatOnRampGateway) InitiateDeposit(depositID, userID string, amountINR float64, rail PaymentRail) (*FiatDepositOrder, error) {
	if amountINR < 100.0 {
		return nil, errors.New("minimum deposit amount is ₹100.00")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	eqUSDT := amountINR / g.activeFXRate
	order := &FiatDepositOrder{
		DepositID:       depositID,
		UserID:          userID,
		AmountINR:       amountINR,
		EquivalentUSDT:  eqUSDT,
		FXRateINRPerUSD: g.activeFXRate,
		Rail:            rail,
		Status:          "PENDING_BANK_CONFIRMATION",
		CreatedAt:       time.Now().UTC(),
	}

	g.deposits[depositID] = order
	fmt.Printf("[Fiat On-Ramp] Initiated deposit %s: ₹%.2f (%.4f USDT via %s)\n", depositID, amountINR, eqUSDT, rail)
	return order, nil
}

// ConfirmBankReceipt processes NPCI / banking webhook and credits USDT balance
func (g *FiatOnRampGateway) ConfirmBankReceipt(depositID, bankRef string) (*FiatDepositOrder, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	order, exists := g.deposits[depositID]
	if !exists {
		return nil, errors.New("deposit order not found")
	}

	if order.Status == "COMPLETED" {
		return order, nil
	}

	order.BankRefNumber = bankRef
	order.Status = "COMPLETED"
	order.CompletedAt = time.Now().UTC()

	fmt.Printf("[Fiat On-Ramp] Confirmed deposit %s with bank ref %s; ready for credit\n", depositID, bankRef)
	return order, nil
}
