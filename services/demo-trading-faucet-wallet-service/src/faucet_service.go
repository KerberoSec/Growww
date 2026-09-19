package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type DemoBalance struct {
	UserID          string    `json:"user_id"`
	VirtualUsdtE6   uint64    `json:"virtual_usdt_e6"` // 10,000 USDT = 10,000 * 1e6
	VirtualBtcE8    uint64    `json:"virtual_btc_e8"`  // 1 BTC = 1 * 1e8
	LastClaimedAt   time.Time `json:"last_claimed_at"`
	ClaimsCount     uint32    `json:"claims_count"`
}

type DemoFaucetService struct {
	mu       sync.RWMutex
	accounts map[string]*DemoBalance
	cooldown time.Duration
}

func NewDemoFaucetService() *DemoFaucetService {
	return &DemoFaucetService{
		accounts: make(map[string]*DemoBalance),
		cooldown: 24 * time.Hour,
	}
}

// AutoCreditNewUser automatically credits 10,000 virtual USDT and 1 virtual BTC
func (s *DemoFaucetService) AutoCreditNewUser(userID string) (*DemoBalance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.accounts[userID]; exists {
		return nil, errors.New("user already registered with demo wallet")
	}

	bal := &DemoBalance{
		UserID:        userID,
		VirtualUsdtE6: 10_000 * 1_000_000,
		VirtualBtcE8:  1 * 100_000_000,
		LastClaimedAt: time.Now().UTC(),
		ClaimsCount:   1,
	}
	s.accounts[userID] = bal
	fmt.Printf("[Demo Faucet] Credited 10,000 USDT and 1 BTC to demo user %s\n", userID)
	return bal, nil
}

// ClaimPeriodicFaucet allows topping up demo balance after cooldown
func (s *DemoFaucetService) ClaimPeriodicFaucet(userID string) (*DemoBalance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bal, exists := s.accounts[userID]
	if !exists {
		return nil, errors.New("demo account not found; run auto-credit first")
	}

	if time.Since(bal.LastClaimedAt) < s.cooldown {
		return nil, fmt.Errorf("faucet cooldown active; retry after %v", s.cooldown-time.Since(bal.LastClaimedAt))
	}

	bal.VirtualUsdtE6 += 10_000 * 1_000_000
	bal.VirtualBtcE8 += 1 * 100_000_000
	bal.LastClaimedAt = time.Now().UTC()
	bal.ClaimsCount++

	return bal, nil
}
