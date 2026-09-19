package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrAccountNotFound      = errors.New("virtual account not found")
	ErrAccountAlreadyExists = errors.New("virtual account already exists")
	ErrInsufficientBalance  = errors.New("insufficient available balance")
	ErrInsufficientLocked   = errors.New("insufficient locked balance")
	ErrHoldNotFound         = errors.New("order hold not found")
	ErrHoldAlreadyProcessed = errors.New("order hold already released or settled")
)

type AssetBalance struct {
	Asset     string `json:"asset"`
	Available uint64 `json:"available"`
	Locked    uint64 `json:"locked"`
}

func (ab *AssetBalance) Total() uint64 {
	return ab.Available + ab.Locked
}

type HoldStatus string

const (
	HoldStatusActive   HoldStatus = "ACTIVE"
	HoldStatusReleased HoldStatus = "RELEASED"
	HoldStatusSettled  HoldStatus = "SETTLED"
)

type HoldRecord struct {
	HoldID    string     `json:"hold_id"`
	OrderID   string     `json:"order_id"`
	UserID    string     `json:"user_id"`
	Asset     string     `json:"asset"`
	Amount    uint64     `json:"amount"`
	Status    HoldStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

type VirtualAccount struct {
	UserID    string                   `json:"user_id"`
	Balances  map[string]*AssetBalance `json:"balances"`
	Holds     map[string]*HoldRecord   `json:"holds"` // orderID -> HoldRecord
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

type EntryType string

const (
	EntryTypeDebit  EntryType = "DEBIT"
	EntryTypeCredit EntryType = "CREDIT"
)

type LedgerEntry struct {
	EntryID       string    `json:"entry_id"`
	TransactionID string    `json:"transaction_id"`
	UserID        string    `json:"user_id"`
	Asset         string    `json:"asset"`
	Type          EntryType `json:"type"`
	Amount        uint64    `json:"amount"`
	Reason        string    `json:"reason"`
	Timestamp     time.Time `json:"timestamp"`
}

type VirtualLedger struct {
	mu            sync.RWMutex
	accounts      map[string]*VirtualAccount
	auditLog      []LedgerEntry
	entrySequence uint64
}

func NewVirtualLedger() *VirtualLedger {
	return &VirtualLedger{
		accounts: make(map[string]*VirtualAccount),
		auditLog: make([]LedgerEntry, 0),
	}
}

// CreateAccount registers a new demo account with zero balances.
func (vl *VirtualLedger) CreateAccount(userID string) (*VirtualAccount, error) {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	if _, exists := vl.accounts[userID]; exists {
		return nil, ErrAccountAlreadyExists
	}

	acc := &VirtualAccount{
		UserID: userID,
		Balances: map[string]*AssetBalance{
			"USDT": {Asset: "USDT", Available: 0, Locked: 0},
			"BTC":  {Asset: "BTC", Available: 0, Locked: 0},
		},
		Holds:     make(map[string]*HoldRecord),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	vl.accounts[userID] = acc
	return acc, nil
}

// GetAccount returns the account copy or error if not found.
func (vl *VirtualLedger) GetAccount(userID string) (*VirtualAccount, error) {
	vl.mu.RLock()
	defer vl.mu.RUnlock()

	acc, exists := vl.accounts[userID]
	if !exists {
		return nil, ErrAccountNotFound
	}
	return acc, nil
}

// CreditBalance adds available balance and logs an audit credit entry.
func (vl *VirtualLedger) CreditBalance(userID, asset string, amount uint64, reason string) error {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	acc, exists := vl.accounts[userID]
	if !exists {
		return ErrAccountNotFound
	}

	bal, ok := acc.Balances[asset]
	if !ok {
		bal = &AssetBalance{Asset: asset, Available: 0, Locked: 0}
		acc.Balances[asset] = bal
	}

	bal.Available += amount
	acc.UpdatedAt = time.Now().UTC()

	vl.recordEntry(userID, asset, EntryTypeCredit, amount, reason)
	return nil
}

// DebitBalance deducts available balance and logs an audit debit entry.
func (vl *VirtualLedger) DebitBalance(userID, asset string, amount uint64, reason string) error {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	acc, exists := vl.accounts[userID]
	if !exists {
		return ErrAccountNotFound
	}

	bal, ok := acc.Balances[asset]
	if !ok || bal.Available < amount {
		return ErrInsufficientBalance
	}

	bal.Available -= amount
	acc.UpdatedAt = time.Now().UTC()

	vl.recordEntry(userID, asset, EntryTypeDebit, amount, reason)
	return nil
}

// HoldBalance reserves funds from available to locked balance for an open limit order.
func (vl *VirtualLedger) HoldBalance(userID, asset string, amount uint64, orderID string) (*HoldRecord, error) {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	acc, exists := vl.accounts[userID]
	if !exists {
		return nil, ErrAccountNotFound
	}

	bal, ok := acc.Balances[asset]
	if !ok || bal.Available < amount {
		return nil, fmt.Errorf("%w: requested %d %s, available %d", ErrInsufficientBalance, amount, asset, bal.Available)
	}

	bal.Available -= amount
	bal.Locked += amount
	acc.UpdatedAt = time.Now().UTC()

	hold := &HoldRecord{
		HoldID:    fmt.Sprintf("hold-%s-%d", orderID, time.Now().UnixNano()),
		OrderID:   orderID,
		UserID:    userID,
		Asset:     asset,
		Amount:    amount,
		Status:    HoldStatusActive,
		CreatedAt: time.Now().UTC(),
	}
	acc.Holds[orderID] = hold

	vl.recordEntry(userID, asset, EntryTypeDebit, amount, fmt.Sprintf("ORDER_HOLD:%s", orderID))
	return hold, nil
}

// ReleaseBalance unlocks held balance back to available when an order is cancelled or expires.
func (vl *VirtualLedger) ReleaseBalance(userID, orderID string) error {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	acc, exists := vl.accounts[userID]
	if !exists {
		return ErrAccountNotFound
	}

	hold, exists := acc.Holds[orderID]
	if !exists {
		return ErrHoldNotFound
	}

	if hold.Status != HoldStatusActive {
		return ErrHoldAlreadyProcessed
	}

	bal, ok := acc.Balances[hold.Asset]
	if !ok || bal.Locked < hold.Amount {
		return ErrInsufficientLocked
	}

	bal.Locked -= hold.Amount
	bal.Available += hold.Amount
	hold.Status = HoldStatusReleased
	acc.UpdatedAt = time.Now().UTC()

	vl.recordEntry(userID, hold.Asset, EntryTypeCredit, hold.Amount, fmt.Sprintf("ORDER_RELEASE:%s", orderID))
	return nil
}

// SettleTrade atomically transfers assets between buyer and seller upon trade execution.
func (vl *VirtualLedger) SettleTrade(
	buyerID, sellerID string,
	baseAsset string, baseQty uint64,
	quoteAsset string, quoteAmount uint64,
	tradeID string,
) error {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	buyer, exists := vl.accounts[buyerID]
	if !exists {
		return fmt.Errorf("buyer %s: %w", buyerID, ErrAccountNotFound)
	}

	seller, exists := vl.accounts[sellerID]
	if !exists {
		return fmt.Errorf("seller %s: %w", sellerID, ErrAccountNotFound)
	}

	// Buyer: pays quoteAmount (USDT), receives baseQty (BTC)
	buyerQuoteBal, ok := buyer.Balances[quoteAsset]
	if !ok || buyerQuoteBal.Total() < quoteAmount {
		return fmt.Errorf("buyer %s has insufficient %s", buyerID, quoteAsset)
	}
	if buyerQuoteBal.Locked >= quoteAmount {
		buyerQuoteBal.Locked -= quoteAmount
	} else {
		// If market order executed without prior hold, deduct available
		rem := quoteAmount - buyerQuoteBal.Locked
		buyerQuoteBal.Locked = 0
		if buyerQuoteBal.Available < rem {
			return fmt.Errorf("buyer %s has insufficient available %s", buyerID, quoteAsset)
		}
		buyerQuoteBal.Available -= rem
	}

	buyerBaseBal, ok := buyer.Balances[baseAsset]
	if !ok {
		buyerBaseBal = &AssetBalance{Asset: baseAsset, Available: 0, Locked: 0}
		buyer.Balances[baseAsset] = buyerBaseBal
	}
	buyerBaseBal.Available += baseQty

	// Seller: pays baseQty (BTC), receives quoteAmount (USDT)
	sellerBaseBal, ok := seller.Balances[baseAsset]
	if !ok || sellerBaseBal.Total() < baseQty {
		return fmt.Errorf("seller %s has insufficient %s", sellerID, baseAsset)
	}
	if sellerBaseBal.Locked >= baseQty {
		sellerBaseBal.Locked -= baseQty
	} else {
		rem := baseQty - sellerBaseBal.Locked
		sellerBaseBal.Locked = 0
		if sellerBaseBal.Available < rem {
			return fmt.Errorf("seller %s has insufficient available %s", sellerID, baseAsset)
		}
		sellerBaseBal.Available -= rem
	}

	sellerQuoteBal, ok := seller.Balances[quoteAsset]
	if !ok {
		sellerQuoteBal = &AssetBalance{Asset: quoteAsset, Available: 0, Locked: 0}
		seller.Balances[quoteAsset] = sellerQuoteBal
	}
	sellerQuoteBal.Available += quoteAmount

	buyer.UpdatedAt = time.Now().UTC()
	seller.UpdatedAt = time.Now().UTC()

	vl.recordEntry(buyerID, quoteAsset, EntryTypeDebit, quoteAmount, fmt.Sprintf("TRADE_SETTLE_PAY:%s", tradeID))
	vl.recordEntry(buyerID, baseAsset, EntryTypeCredit, baseQty, fmt.Sprintf("TRADE_SETTLE_RCV:%s", tradeID))
	vl.recordEntry(sellerID, baseAsset, EntryTypeDebit, baseQty, fmt.Sprintf("TRADE_SETTLE_PAY:%s", tradeID))
	vl.recordEntry(sellerID, quoteAsset, EntryTypeCredit, quoteAmount, fmt.Sprintf("TRADE_SETTLE_RCV:%s", tradeID))

	return nil
}

// ResetAccount clears open holds and resets balances to initial demo baseline.
func (vl *VirtualLedger) ResetAccount(userID string, initialUsdtE6, initialBtcE8 uint64) error {
	vl.mu.Lock()
	defer vl.mu.Unlock()

	acc, exists := vl.accounts[userID]
	if !exists {
		return ErrAccountNotFound
	}

	// Release or cancel all active holds
	for _, hold := range acc.Holds {
		if hold.Status == HoldStatusActive {
			hold.Status = HoldStatusReleased
		}
	}

	acc.Balances["USDT"] = &AssetBalance{Asset: "USDT", Available: initialUsdtE6, Locked: 0}
	acc.Balances["BTC"] = &AssetBalance{Asset: "BTC", Available: initialBtcE8, Locked: 0}
	acc.UpdatedAt = time.Now().UTC()

	vl.recordEntry(userID, "USDT", EntryTypeCredit, initialUsdtE6, "PORTFOLIO_RESET")
	vl.recordEntry(userID, "BTC", EntryTypeCredit, initialBtcE8, "PORTFOLIO_RESET")
	return nil
}

func (vl *VirtualLedger) recordEntry(userID, asset string, entryType EntryType, amount uint64, reason string) {
	vl.entrySequence++
	vl.auditLog = append(vl.auditLog, LedgerEntry{
		EntryID:       fmt.Sprintf("LEDGER-%012d", vl.entrySequence),
		TransactionID: fmt.Sprintf("TXN-%d", time.Now().UnixNano()),
		UserID:        userID,
		Asset:         asset,
		Type:          entryType,
		Amount:        amount,
		Reason:        reason,
		Timestamp:     time.Now().UTC(),
	})
}
