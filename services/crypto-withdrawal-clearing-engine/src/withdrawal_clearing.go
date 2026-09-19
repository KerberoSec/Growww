package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type WithdrawalStatus string

const (
	StatusPendingReview WithdrawalStatus = "PENDING_REVIEW"
	StatusTimelocked    WithdrawalStatus = "TIMELOCKED"
	StatusApproved      WithdrawalStatus = "APPROVED"
	StatusBroadcasted   WithdrawalStatus = "BROADCASTED"
	StatusRejected      WithdrawalStatus = "REJECTED"
)

type WhitelistEntry struct {
	Address        string
	Chain          string
	Label          string
	AddedAt        time.Time
	TimelockExpiry time.Time
	Active         bool
}

type WithdrawalRequest struct {
	RequestID       string
	UserID          string
	Asset           string
	DestinationAddr string
	AmountE8        uint64
	Status          WithdrawalStatus
	CreatedAt       time.Time
	ReleaseAt       time.Time
}

type WithdrawalClearingEngine struct {
	mu         sync.RWMutex
	whitelist  map[string]WhitelistEntry // "user:addr" -> entry
	requests   map[string]*WithdrawalRequest
	timeLock24 time.Duration
}

func NewWithdrawalClearingEngine() *WithdrawalClearingEngine {
	return &WithdrawalClearingEngine{
		whitelist:  make(map[string]WhitelistEntry),
		requests:   make(map[string]*WithdrawalRequest),
		timeLock24: 24 * time.Hour,
	}
}

// AddWhitelistAddress adds an external address with mandatory 24-hour cooling time-lock
func (e *WithdrawalClearingEngine) AddWhitelistAddress(userID, address, chain, label string) WhitelistEntry {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	entry := WhitelistEntry{
		Address:        address,
		Chain:          chain,
		Label:          label,
		AddedAt:        now,
		TimelockExpiry: now.Add(e.timeLock24),
		Active:         false, // Becomes active only after 24h
	}
	key := fmt.Sprintf("%s:%s", userID, address)
	e.whitelist[key] = entry
	fmt.Printf("[Withdrawal Security] Address %s enrolled for user %s with 24h timelock\n", address, userID)
	return entry
}

// RequestWithdrawal verifies whitelist status and enforces security cooling periods
func (e *WithdrawalClearingEngine) RequestWithdrawal(reqID, userID, asset, destAddr string, amountE8 uint64) (*WithdrawalRequest, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, destAddr)
	entry, exists := e.whitelist[key]
	if !exists {
		return nil, errors.New("destination address is not on user whitelist")
	}

	if time.Now().UTC().Before(entry.TimelockExpiry) {
		return nil, fmt.Errorf("address timelock active: cannot withdraw until %v", entry.TimelockExpiry)
	}

	req := &WithdrawalRequest{
		RequestID:       reqID,
		UserID:          userID,
		Asset:           asset,
		DestinationAddr: destAddr,
		AmountE8:        amountE8,
		Status:          StatusPendingReview,
		CreatedAt:       time.Now().UTC(),
		ReleaseAt:       time.Now().UTC().Add(15 * time.Minute), // Standard 15m review buffer
	}
	e.requests[reqID] = req
	return req, nil
}
