package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type SubAccountRole string

const (
	RoleReadOnlyAnalyst SubAccountRole = "READ_ONLY_ANALYST"
	RoleTradingOnlyBot  SubAccountRole = "TRADING_ONLY_BOT"
	RoleFullManager     SubAccountRole = "FULL_MANAGER"
)

var (
	ErrMaxSubAccountsExceeded = errors.New("maximum 100 sub-accounts per institutional master account exceeded")
	ErrSubAccountNotFound     = errors.New("sub-account not found")
	ErrMasterAccountNotFound  = errors.New("master account entity not found")
	ErrPermissionDenied       = errors.New("sub-account role does not permit this operation")
	ErrCrossMasterTransfer    = errors.New("cross-master transfer blocked: sub-accounts belong to different corporate entities")
	ErrInsufficientBalance    = errors.New("insufficient balance for internal transfer")
	ErrInvalidTransferAmount  = errors.New("transfer amount must be > 0")
)

type SubAccount struct {
	SubAccountID string            `json:"sub_account_id"`
	MasterID     string            `json:"master_id"`
	Label        string            `json:"label"`
	Role         SubAccountRole    `json:"role"`
	APIKeys      map[string]bool   `json:"api_keys"`
	BalancesE8   map[string]uint64 `json:"balances_e8"` // asset -> balance
	CreatedAt    time.Time         `json:"created_at"`
	IsFrozen     bool              `json:"is_frozen"`
}

type MasterAccount struct {
	MasterID     string                 `json:"master_id"`
	CompanyName  string                 `json:"company_name"`
	CIN          string                 `json:"cin"` // Corporate Identification Number
	SubAccounts  map[string]*SubAccount `json:"sub_accounts"`
	MasterWallet map[string]uint64      `json:"master_wallet_e8"`
	CreatedAt    time.Time              `json:"created_at"`
}

type InternalTransferReceipt struct {
	TransferID       string    `json:"transfer_id"`
	MasterID         string    `json:"master_id"`
	FromAccountID    string    `json:"from_account_id"`
	ToAccountID      string    `json:"to_account_id"`
	Asset            string    `json:"asset"`
	AmountE8         uint64    `json:"amount_e8"`
	FeeE8            uint64    `json:"fee_e8"` // Strictly 0 for internal transfers
	ExecutedAt       time.Time `json:"executed_at"`
	SettlementStatus string    `json:"settlement_status"`
}

type SubAccountHierarchyManager struct {
	mu      sync.RWMutex
	masters map[string]*MasterAccount
	subs    map[string]*SubAccount // subID -> SubAccount
	maxSubs int
}

func NewSubAccountHierarchyManager() *SubAccountHierarchyManager {
	return &SubAccountHierarchyManager{
		masters: make(map[string]*MasterAccount),
		subs:    make(map[string]*SubAccount),
		maxSubs: 100,
	}
}

// RegisterMasterAccount initializes a corporate institutional entity
func (m *SubAccountHierarchyManager) RegisterMasterAccount(masterID, companyName, cin string) (*MasterAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.masters[masterID]; exists {
		return nil, errors.New("master account already exists")
	}

	master := &MasterAccount{
		MasterID:     masterID,
		CompanyName:  companyName,
		CIN:          cin,
		SubAccounts:  make(map[string]*SubAccount),
		MasterWallet: make(map[string]uint64),
		CreatedAt:    time.Now().UTC(),
	}
	m.masters[masterID] = master
	return master, nil
}

// CreateSubAccount creates an isolated child account under a master corporate account
func (m *SubAccountHierarchyManager) CreateSubAccount(masterID, subID, label string, role SubAccountRole) (*SubAccount, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	master, exists := m.masters[masterID]
	if !exists {
		return nil, ErrMasterAccountNotFound
	}

	if len(master.SubAccounts) >= m.maxSubs {
		return nil, ErrMaxSubAccountsExceeded
	}

	if _, exists := m.subs[subID]; exists {
		return nil, errors.New("sub-account ID already in use")
	}

	sub := &SubAccount{
		SubAccountID: subID,
		MasterID:     masterID,
		Label:        label,
		Role:         role,
		APIKeys:      make(map[string]bool),
		BalancesE8:   make(map[string]uint64),
		CreatedAt:    time.Now().UTC(),
		IsFrozen:     false,
	}

	master.SubAccounts[subID] = sub
	m.subs[subID] = sub

	return sub, nil
}

// DepositToMaster credits master account funds
func (m *SubAccountHierarchyManager) DepositToMaster(masterID, asset string, amountE8 uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	master, exists := m.masters[masterID]
	if !exists {
		return ErrMasterAccountNotFound
	}
	master.MasterWallet[asset] += amountE8
	return nil
}

// TransferInternal routes funds between master and sub-accounts with zero fees
func (m *SubAccountHierarchyManager) TransferInternal(
	transferID, callerSubID string,
	fromAccID, toAccID, asset string,
	amountE8 uint64,
) (*InternalTransferReceipt, error) {
	if amountE8 == 0 {
		return nil, ErrInvalidTransferAmount
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check caller role if initiated from a sub-account
	if callerSubID != "" {
		callerSub, exists := m.subs[callerSubID]
		if !exists {
			return nil, ErrSubAccountNotFound
		}
		if callerSub.Role != RoleFullManager {
			return nil, fmt.Errorf("%w: role %s cannot initiate internal fund transfers", ErrPermissionDenied, callerSub.Role)
		}
	}

	// Resolve source and destination
	var fromMasterID, toMasterID string
	var fromBal *uint64

	// Resolve 'from'
	if master, ok := m.masters[fromAccID]; ok {
		fromMasterID = master.MasterID
		val := master.MasterWallet[asset]
		fromBal = &val
	} else if sub, ok := m.subs[fromAccID]; ok {
		fromMasterID = sub.MasterID
		val := sub.BalancesE8[asset]
		fromBal = &val
	} else {
		return nil, errors.New("source account not found")
	}

	// Resolve 'to'
	if master, ok := m.masters[toAccID]; ok {
		toMasterID = master.MasterID
	} else if sub, ok := m.subs[toAccID]; ok {
		toMasterID = sub.MasterID
	} else {
		return nil, errors.New("destination account not found")
	}

	// Enforce corporate isolation: must belong to the exact same master account
	if fromMasterID != toMasterID {
		return nil, ErrCrossMasterTransfer
	}

	if *fromBal < amountE8 {
		return nil, ErrInsufficientBalance
	}

	// Execute atomic debit & credit
	if master, ok := m.masters[fromAccID]; ok {
		master.MasterWallet[asset] -= amountE8
	} else if sub, ok := m.subs[fromAccID]; ok {
		sub.BalancesE8[asset] -= amountE8
	}

	if master, ok := m.masters[toAccID]; ok {
		master.MasterWallet[asset] += amountE8
	} else if sub, ok := m.subs[toAccID]; ok {
		sub.BalancesE8[asset] += amountE8
	}

	receipt := &InternalTransferReceipt{
		TransferID:       transferID,
		MasterID:         fromMasterID,
		FromAccountID:    fromAccID,
		ToAccountID:      toAccID,
		Asset:            asset,
		AmountE8:         amountE8,
		FeeE8:            0, // Zero fee internal transfer
		ExecutedAt:       time.Now().UTC(),
		SettlementStatus: "SETTLED_INTERNAL_DOUBLE_ENTRY",
	}

	return receipt, nil
}

// GetAggregateBalance calculates total asset balance across master wallet and all child sub-accounts
func (m *SubAccountHierarchyManager) GetAggregateBalance(masterID, asset string) (uint64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	master, exists := m.masters[masterID]
	if !exists {
		return 0, ErrMasterAccountNotFound
	}

	total := master.MasterWallet[asset]
	for _, sub := range master.SubAccounts {
		total += sub.BalancesE8[asset]
	}

	return total, nil
}

// ValidateOrderPlacement validates if a sub-account has permission to place orders
func (m *SubAccountHierarchyManager) ValidateOrderPlacement(subID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sub, exists := m.subs[subID]
	if !exists {
		return ErrSubAccountNotFound
	}

	if sub.IsFrozen {
		return errors.New("sub-account is frozen")
	}

	if sub.Role == RoleReadOnlyAnalyst {
		return fmt.Errorf("%w: read-only analyst cannot execute orders", ErrPermissionDenied)
	}

	return nil
}
