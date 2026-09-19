package main

import (
	"errors"
	"fmt"
	"sync"
)

// PreTradeRiskConfig holds platform-wide and per-tier risk boundaries
type PreTradeRiskConfig struct {
	MaxOrderNotionalUSD uint64 // e.g. $1,000,000 in e8
	MaxPriceDeviationBps uint32 // e.g. 500 = 5% fat finger threshold
	MaxDailyNotionalUSD uint64 // e.g. $5,000,000 in e8
}

// UserAccountRiskProfile tracks balance, limits, and status
type UserAccountRiskProfile struct {
	UserID             string
	AvailableBalanceE8 uint64
	DailyTurnoverE8    uint64
	IsFrozen           bool
	IsCompliant        bool
}

type PreTradeRiskEngine struct {
	mu       sync.RWMutex
	config   PreTradeRiskConfig
	profiles map[string]*UserAccountRiskProfile
}

func NewPreTradeRiskEngine(cfg PreTradeRiskConfig) *PreTradeRiskEngine {
	return &PreTradeRiskEngine{
		config:   cfg,
		profiles: make(map[string]*UserAccountRiskProfile),
	}
}

func (e *PreTradeRiskEngine) RegisterUserProfile(profile *UserAccountRiskProfile) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.profiles[profile.UserID] = profile
}

// ValidateOrder performs comprehensive pre-trade risk and fat-finger checks
func (e *PreTradeRiskEngine) ValidateOrder(
	userID string,
	side string,
	priceE8 uint64,
	quantityE8 uint64,
	referenceMarkPriceE8 uint64,
) error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	profile, exists := e.profiles[userID]
	if !exists {
		return errors.New("user risk profile not found")
	}

	if profile.IsFrozen {
		return errors.New("account is frozen by surveillance / compliance")
	}

	if !profile.IsCompliant {
		return errors.New("user has not completed mandatory KYC / suitability onboarding")
	}

	// 1. Fat-Finger Price Deviation Check
	if referenceMarkPriceE8 > 0 {
		var diff uint64
		if priceE8 > referenceMarkPriceE8 {
			diff = priceE8 - referenceMarkPriceE8
		} else {
			diff = referenceMarkPriceE8 - priceE8
		}

		deviationBps := uint32((diff * 10000) / referenceMarkPriceE8)
		if deviationBps > e.config.MaxPriceDeviationBps {
			return fmt.Errorf("fat-finger limit breached: price deviation %d bps exceeds max %d bps",
				deviationBps, e.config.MaxPriceDeviationBps)
		}
	}

	// 2. Order Notional Threshold
	orderNotionalE8 := (priceE8 * quantityE8) / 1e8
	if e.config.MaxOrderNotionalUSD > 0 && orderNotionalE8 > e.config.MaxOrderNotionalUSD {
		return fmt.Errorf("order notional %d exceeds maximum single order limit %d",
			orderNotionalE8, e.config.MaxOrderNotionalUSD)
	}

	// 3. Balance Sufficiency (for Buy orders)
	if side == "BUY" {
		if orderNotionalE8 > profile.AvailableBalanceE8 {
			return fmt.Errorf("insufficient available collateral: required %d > available %d",
				orderNotionalE8, profile.AvailableBalanceE8)
		}
	}

	return nil
}

func main() {
	cfg := PreTradeRiskConfig{
		MaxOrderNotionalUSD:  100_000 * 1e8,
		MaxPriceDeviationBps: 500, // 5%
		MaxDailyNotionalUSD:  1_000_000 * 1e8,
	}
	engine := NewPreTradeRiskEngine(cfg)
	fmt.Println("Pre-Trade Risk Engine ready.")
	_ = engine
}
