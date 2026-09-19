package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrUserAlreadyRegistered = errors.New("user already registered with demo wallet")
	ErrAccountNotRegistered  = errors.New("demo account not found; run auto-credit first")
	ErrFaucetCooldownActive  = errors.New("faucet cooldown active; please wait before claiming again")
	ErrRateLimitedIP         = errors.New("IP claim limit exceeded; please try again later")
	ErrBalanceNotEligible    = errors.New("balance is above threshold (1,000 USDT); faucet refill not required yet")
)

const (
	DefaultInitialUsdtE6 uint64 = 10_000 * 1_000_000 // 10,000 USDT
	DefaultInitialBtcE8  uint64 = 1 * 100_000_000   // 1 BTC
	RefillThresholdUsdtE6 uint64 = 1_000 * 1_000_000 // 1,000 USDT
	DefaultCooldown              = 24 * time.Hour
	MaxClaimsPerIPWindow         = 5
	IPRateWindow                 = 1 * time.Hour
)

type DemoBalance struct {
	UserID        string    `json:"user_id"`
	VirtualUsdtE6 uint64    `json:"virtual_usdt_e6"`
	VirtualBtcE8  uint64    `json:"virtual_btc_e8"`
	LockedUsdtE6  uint64    `json:"locked_usdt_e6"`
	LockedBtcE8   uint64    `json:"locked_btc_e8"`
	LastClaimedAt time.Time `json:"last_claimed_at"`
	ClaimsCount   uint32    `json:"claims_count"`
}

type EligibilityStatus struct {
	Eligible        bool          `json:"eligible"`
	Reason          string        `json:"reason"`
	RemainingTime   time.Duration `json:"remaining_time"`
	CurrentUsdtE6   uint64        `json:"current_usdt_e6"`
	CurrentBtcE8    uint64        `json:"current_btc_e8"`
}

type DemoFaucetService struct {
	mu            sync.RWMutex
	ledger        *VirtualLedger
	userMeta      map[string]*DemoBalance
	ipClaims      map[string][]time.Time
	cooldown      time.Duration
	refillThresh  uint64
	enforceThresh bool
}

func NewDemoFaucetService() *DemoFaucetService {
	return &DemoFaucetService{
		ledger:        NewVirtualLedger(),
		userMeta:      make(map[string]*DemoBalance),
		ipClaims:      make(map[string][]time.Time),
		cooldown:      DefaultCooldown,
		refillThresh:  RefillThresholdUsdtE6,
		enforceThresh: false, // by default allow claims whenever 24h cooldown expires
	}
}

// SetCooldown allows adjusting cooldown duration (e.g. for testing).
func (s *DemoFaucetService) SetCooldown(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cooldown = d
}

// SetEnforceThreshold toggles the <1,000 USDT threshold requirement.
func (s *DemoFaucetService) SetEnforceThreshold(enforce bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enforceThresh = enforce
}

// AutoCreditNewUser automatically registers and credits 10,000 vUSDT and 1 vBTC upon onboarding.
func (s *DemoFaucetService) AutoCreditNewUser(userID string, clientIP string) (*DemoBalance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.userMeta[userID]; exists {
		return nil, ErrUserAlreadyRegistered
	}

	if err := s.checkIPRateLimitLocked(clientIP); err != nil {
		return nil, err
	}

	// Create account in virtual ledger
	if _, err := s.ledger.CreateAccount(userID); err != nil && !errors.Is(err, ErrAccountAlreadyExists) {
		return nil, fmt.Errorf("failed to create ledger account: %w", err)
	}

	// Credit initial balances
	if err := s.ledger.CreditBalance(userID, "USDT", DefaultInitialUsdtE6, "FAUCET_INITIAL_CREDIT"); err != nil {
		return nil, err
	}
	if err := s.ledger.CreditBalance(userID, "BTC", DefaultInitialBtcE8, "FAUCET_INITIAL_CREDIT"); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	bal := &DemoBalance{
		UserID:        userID,
		VirtualUsdtE6: DefaultInitialUsdtE6,
		VirtualBtcE8:  DefaultInitialBtcE8,
		LockedUsdtE6:  0,
		LockedBtcE8:   0,
		LastClaimedAt: now,
		ClaimsCount:   1,
	}
	s.userMeta[userID] = bal
	s.recordIPClaimLocked(clientIP, now)

	return bal, nil
}

// ClaimPeriodicFaucet allows topping up demo balance after the 24-hour cooldown period.
func (s *DemoFaucetService) ClaimPeriodicFaucet(userID string, clientIP string) (*DemoBalance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bal, exists := s.userMeta[userID]
	if !exists {
		return nil, ErrAccountNotRegistered
	}

	// 1. Check IP rate limits
	if err := s.checkIPRateLimitLocked(clientIP); err != nil {
		return nil, err
	}

	// 2. Check 24-hour cooldown
	elapsed := time.Since(bal.LastClaimedAt)
	if elapsed < s.cooldown {
		rem := s.cooldown - elapsed
		return nil, fmt.Errorf("%w: %v remaining", ErrFaucetCooldownActive, rem.Round(time.Second))
	}

	// 3. Optional balance threshold check
	acc, err := s.ledger.GetAccount(userID)
	if err != nil {
		return nil, err
	}
	currentUsdt := acc.Balances["USDT"].Available
	if s.enforceThresh && currentUsdt >= s.refillThresh {
		return nil, fmt.Errorf("%w: current available %d e6 >= threshold %d e6", ErrBalanceNotEligible, currentUsdt, s.refillThresh)
	}

	// 4. Credit periodic refill
	if err := s.ledger.CreditBalance(userID, "USDT", DefaultInitialUsdtE6, "FAUCET_PERIODIC_REFILL"); err != nil {
		return nil, err
	}
	if err := s.ledger.CreditBalance(userID, "BTC", DefaultInitialBtcE8, "FAUCET_PERIODIC_REFILL"); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	bal.LastClaimedAt = now
	bal.ClaimsCount++
	s.syncBalancesFromLedgerLocked(bal, acc)
	s.recordIPClaimLocked(clientIP, now)

	return bal, nil
}

// ResetVirtualPortfolio resets the user's account to the initial 10,000 USDT and 1 BTC.
func (s *DemoFaucetService) ResetVirtualPortfolio(userID string) (*DemoBalance, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bal, exists := s.userMeta[userID]
	if !exists {
		return nil, ErrAccountNotRegistered
	}

	if err := s.ledger.ResetAccount(userID, DefaultInitialUsdtE6, DefaultInitialBtcE8); err != nil {
		return nil, err
	}

	acc, err := s.ledger.GetAccount(userID)
	if err != nil {
		return nil, err
	}

	s.syncBalancesFromLedgerLocked(bal, acc)
	return bal, nil
}

// GetBalance queries current available and locked balances for the user.
func (s *DemoFaucetService) GetBalance(userID string) (*DemoBalance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bal, exists := s.userMeta[userID]
	if !exists {
		return nil, ErrAccountNotRegistered
	}

	acc, err := s.ledger.GetAccount(userID)
	if err != nil {
		return nil, err
	}

	res := *bal
	res.VirtualUsdtE6 = acc.Balances["USDT"].Available
	res.LockedUsdtE6 = acc.Balances["USDT"].Locked
	res.VirtualBtcE8 = acc.Balances["BTC"].Available
	res.LockedBtcE8 = acc.Balances["BTC"].Locked
	return &res, nil
}

// GetEligibility checks whether a user can claim from the faucet.
func (s *DemoFaucetService) GetEligibility(userID string, clientIP string) (*EligibilityStatus, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bal, exists := s.userMeta[userID]
	if !exists {
		return &EligibilityStatus{
			Eligible:      true,
			Reason:        "New user eligible for auto-credit onboarding",
			RemainingTime: 0,
		}, nil
	}

	acc, err := s.ledger.GetAccount(userID)
	if err != nil {
		return nil, err
	}

	elapsed := time.Since(bal.LastClaimedAt)
	if elapsed < s.cooldown {
		return &EligibilityStatus{
			Eligible:      false,
			Reason:        fmt.Sprintf("24h cooldown active; %v remaining", (s.cooldown - elapsed).Round(time.Second)),
			RemainingTime: s.cooldown - elapsed,
			CurrentUsdtE6: acc.Balances["USDT"].Available,
			CurrentBtcE8:  acc.Balances["BTC"].Available,
		}, nil
	}

	if s.enforceThresh && acc.Balances["USDT"].Available >= s.refillThresh {
		return &EligibilityStatus{
			Eligible:      false,
			Reason:        "Balance is above low-funds threshold (1,000 USDT)",
			RemainingTime: 0,
			CurrentUsdtE6: acc.Balances["USDT"].Available,
			CurrentBtcE8:  acc.Balances["BTC"].Available,
		}, nil
	}

	return &EligibilityStatus{
		Eligible:      true,
		Reason:        "Eligible for faucet refill",
		RemainingTime: 0,
		CurrentUsdtE6: acc.Balances["USDT"].Available,
		CurrentBtcE8:  acc.Balances["BTC"].Available,
	}, nil
}

// Ledger returns the underlying virtual ledger for trade matching engine integration.
func (s *DemoFaucetService) Ledger() *VirtualLedger {
	return s.ledger
}

func (s *DemoFaucetService) syncBalancesFromLedgerLocked(bal *DemoBalance, acc *VirtualAccount) {
	bal.VirtualUsdtE6 = acc.Balances["USDT"].Available
	bal.LockedUsdtE6 = acc.Balances["USDT"].Locked
	bal.VirtualBtcE8 = acc.Balances["BTC"].Available
	bal.LockedBtcE8 = acc.Balances["BTC"].Locked
}

func (s *DemoFaucetService) checkIPRateLimitLocked(ip string) error {
	if ip == "" {
		return nil
	}
	now := time.Now().UTC()
	claims, exists := s.ipClaims[ip]
	if !exists {
		return nil
	}

	cutoff := now.Add(-IPRateWindow)
	valid := 0
	for _, t := range claims {
		if t.After(cutoff) {
			valid++
		}
	}
	if valid >= MaxClaimsPerIPWindow {
		return ErrRateLimitedIP
	}
	return nil
}

func (s *DemoFaucetService) recordIPClaimLocked(ip string, t time.Time) {
	if ip == "" {
		return
	}
	cutoff := t.Add(-IPRateWindow)
	recent := make([]time.Time, 0, len(s.ipClaims[ip])+1)
	for _, prev := range s.ipClaims[ip] {
		if prev.After(cutoff) {
			recent = append(recent, prev)
		}
	}
	recent = append(recent, t)
	s.ipClaims[ip] = recent
}

func main() {
	svc := NewDemoFaucetService()
	fmt.Println("Growww / NBSE Demo Trading Faucet & Virtual Wallet Ledger Service initialized.")
	bal, err := svc.AutoCreditNewUser("demo-trader-01", "127.0.0.1")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Credited: %+v\n", bal)
}
