// Package main provides the enhanced wallet account service for NBSE.
// Implements deposit address generation, withdrawal processing, balance queries,
// and multi-asset ledger management with double-entry bookkeeping.
package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"time"
)

// AssetType represents the type of asset managed in the wallet.
type AssetType string

const (
	AssetINR  AssetType = "INR"
	AssetUSDT AssetType = "USDT"
	AssetBTC  AssetType = "BTC"
	AssetETH  AssetType = "ETH"
)

// WithdrawalStatus represents the lifecycle state of a withdrawal.
type WithdrawalStatus string

const (
	WithdrawalPending    WithdrawalStatus = "PENDING"
	WithdrawalProcessing WithdrawalStatus = "PROCESSING"
	WithdrawalCompleted  WithdrawalStatus = "COMPLETED"
	WithdrawalFailed     WithdrawalStatus = "FAILED"
	WithdrawalRejected   WithdrawalStatus = "REJECTED"
)

// DepositAddress holds a generated deposit address for a user+asset+network.
type DepositAddress struct {
	UserID    string    `json:"user_id"`
	Asset     AssetType `json:"asset"`
	Network   string    `json:"network"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
}

// WithdrawalRequest represents a withdrawal from user wallet.
type WithdrawalRequest struct {
	WithdrawalID string           `json:"withdrawal_id"`
	UserID       string           `json:"user_id"`
	Asset        AssetType        `json:"asset"`
	AmountE8     uint64           `json:"amount_e8"`
	Destination  string           `json:"destination"`
	Network      string           `json:"network"`
	FeeE8        uint64           `json:"fee_e8"`
	Status       WithdrawalStatus `json:"status"`
	TxHash       string           `json:"tx_hash"`
	CreatedAt    time.Time        `json:"created_at"`
	CompletedAt  time.Time        `json:"completed_at"`
}

// AssetBalance represents the balance of a single asset in the wallet.
type AssetBalance struct {
	Asset     AssetType `json:"asset"`
	Available uint64    `json:"available_e8"`
	Locked    uint64    `json:"locked_e8"`
	Total     uint64    `json:"total_e8"`
}

// WalletService manages multi-asset wallets, deposit addresses, and withdrawals.
type WalletService struct {
	mu          sync.RWMutex
	balances    map[string]map[AssetType]*AssetBalance // userID -> asset -> balance
	addresses   map[string][]*DepositAddress           // userID -> deposit addresses
	withdrawals map[string]*WithdrawalRequest          // withdrawalID -> request
}

// NewWalletService creates a new WalletService instance.
func NewWalletService() *WalletService {
	return &WalletService{
		balances:    make(map[string]map[AssetType]*AssetBalance),
		addresses:   make(map[string][]*DepositAddress),
		withdrawals: make(map[string]*WithdrawalRequest),
	}
}

// GenerateDepositAddress creates a deterministic deposit address for a user/asset/network tuple.
func (ws *WalletService) GenerateDepositAddress(userID string, asset AssetType, network string) (*DepositAddress, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}
	if network == "" {
		return nil, errors.New("network is required")
	}

	// Deterministic address generation using SHA-256 of user+asset+network
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", userID, asset, network)))
	address := fmt.Sprintf("0x%x", hash[:20])

	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Check for existing address
	for _, addr := range ws.addresses[userID] {
		if addr.Asset == asset && addr.Network == network {
			return addr, nil // Return existing address (idempotent)
		}
	}

	da := &DepositAddress{
		UserID:    userID,
		Asset:     asset,
		Network:   network,
		Address:   address,
		CreatedAt: time.Now().UTC(),
	}

	ws.addresses[userID] = append(ws.addresses[userID], da)
	return da, nil
}

// CreditDeposit credits a deposit to the user's available balance.
func (ws *WalletService) CreditDeposit(userID string, asset AssetType, amountE8 uint64) error {
	if amountE8 == 0 {
		return errors.New("deposit amount must be positive")
	}

	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.ensureBalance(userID, asset)
	ws.balances[userID][asset].Available += amountE8
	ws.balances[userID][asset].Total += amountE8

	return nil
}

// GetBalance returns the balance for a specific asset.
func (ws *WalletService) GetBalance(userID string, asset AssetType) AssetBalance {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	if userAssets, ok := ws.balances[userID]; ok {
		if bal, ok := userAssets[asset]; ok {
			return *bal
		}
	}
	return AssetBalance{Asset: asset}
}

// GetAllBalances returns all asset balances for a user.
func (ws *WalletService) GetAllBalances(userID string) []AssetBalance {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	var result []AssetBalance
	if userAssets, ok := ws.balances[userID]; ok {
		for _, bal := range userAssets {
			result = append(result, *bal)
		}
	}
	return result
}

// LockFunds moves funds from available to locked (for orders, withdrawals).
func (ws *WalletService) LockFunds(userID string, asset AssetType, amountE8 uint64) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.ensureBalance(userID, asset)
	bal := ws.balances[userID][asset]

	if bal.Available < amountE8 {
		return fmt.Errorf("insufficient available balance: have %d, need %d", bal.Available, amountE8)
	}

	bal.Available -= amountE8
	bal.Locked += amountE8
	return nil
}

// UnlockFunds moves funds from locked back to available (order cancellation).
func (ws *WalletService) UnlockFunds(userID string, asset AssetType, amountE8 uint64) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.ensureBalance(userID, asset)
	bal := ws.balances[userID][asset]

	if bal.Locked < amountE8 {
		return fmt.Errorf("insufficient locked balance: have %d, need %d", bal.Locked, amountE8)
	}

	bal.Locked -= amountE8
	bal.Available += amountE8
	return nil
}

// InitiateWithdrawal creates a withdrawal request and locks the necessary funds.
func (ws *WalletService) InitiateWithdrawal(
	withdrawalID, userID string,
	asset AssetType,
	amountE8, feeE8 uint64,
	destination, network string,
) (*WithdrawalRequest, error) {
	if amountE8 == 0 {
		return nil, errors.New("withdrawal amount must be positive")
	}
	if destination == "" {
		return nil, errors.New("destination address is required")
	}

	totalDeduct := amountE8 + feeE8

	ws.mu.Lock()
	defer ws.mu.Unlock()

	if _, exists := ws.withdrawals[withdrawalID]; exists {
		return nil, fmt.Errorf("withdrawal ID %s already exists", withdrawalID)
	}

	ws.ensureBalance(userID, asset)
	bal := ws.balances[userID][asset]

	if bal.Available < totalDeduct {
		return nil, fmt.Errorf("insufficient balance: have %d, need %d (amount + fee)", bal.Available, totalDeduct)
	}

	// Lock funds atomically
	bal.Available -= totalDeduct
	bal.Locked += totalDeduct

	req := &WithdrawalRequest{
		WithdrawalID: withdrawalID,
		UserID:       userID,
		Asset:        asset,
		AmountE8:     amountE8,
		Destination:  destination,
		Network:      network,
		FeeE8:        feeE8,
		Status:       WithdrawalPending,
		CreatedAt:    time.Now().UTC(),
	}

	ws.withdrawals[withdrawalID] = req
	return req, nil
}

// CompleteWithdrawal marks withdrawal as completed and deducts locked funds.
func (ws *WalletService) CompleteWithdrawal(withdrawalID, txHash string) (*WithdrawalRequest, error) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	req, exists := ws.withdrawals[withdrawalID]
	if !exists {
		return nil, fmt.Errorf("withdrawal %s not found", withdrawalID)
	}
	if req.Status != WithdrawalPending && req.Status != WithdrawalProcessing {
		return nil, fmt.Errorf("withdrawal %s in invalid state: %s", withdrawalID, req.Status)
	}

	totalDeduct := req.AmountE8 + req.FeeE8
	bal := ws.balances[req.UserID][req.Asset]
	bal.Locked -= totalDeduct
	bal.Total -= totalDeduct

	req.Status = WithdrawalCompleted
	req.TxHash = txHash
	req.CompletedAt = time.Now().UTC()

	return req, nil
}

// ensureBalance initializes balance maps if they don't exist.
func (ws *WalletService) ensureBalance(userID string, asset AssetType) {
	if ws.balances[userID] == nil {
		ws.balances[userID] = make(map[AssetType]*AssetBalance)
	}
	if ws.balances[userID][asset] == nil {
		ws.balances[userID][asset] = &AssetBalance{Asset: asset}
	}
}
