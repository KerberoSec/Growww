package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type DematPledgeRecord struct {
	PledgeID       string    `json:"pledge_id"`
	UserID         string    `json:"user_id"`
	DematISIN      string    `json:"demat_isin"`
	PledgedShares  uint64    `json:"pledged_shares"`
	MarketPriceINR float64   `json:"market_price_inr"`
	HaircutPct     float64   `json:"haircut_pct"` // e.g. 20% haircut on Group 1 equities
	CollateralValINR float64 `json:"collateral_val_inr"`
	AllowedMarginUSDT float64 `json:"allowed_margin_usdt"`
	PledgedAt      time.Time `json:"pledged_at"`
}

type CrossCollateralEngine struct {
	mu          sync.RWMutex
	pledges     map[string]*DematPledgeRecord
	usdtInrRate float64
}

func NewCrossCollateralEngine(fxRate float64) *CrossCollateralEngine {
	if fxRate <= 0 {
		fxRate = 84.50
	}
	return &CrossCollateralEngine{
		pledges:     make(map[string]*DematPledgeRecord),
		usdtInrRate: fxRate,
	}
}

// PledgeDematStock pledges CDSL/NSDL demat shares to unlock crypto spot margin
func (e *CrossCollateralEngine) PledgeDematStock(id, userID, isin string, shares uint64, priceINR, haircutPct float64) (*DematPledgeRecord, error) {
	if shares == 0 || priceINR <= 0 {
		return nil, errors.New("invalid share quantity or price")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	grossValueINR := float64(shares) * priceINR
	collateralINR := grossValueINR * (1.0 - (haircutPct / 100.0))
	marginUSDT := collateralINR / e.usdtInrRate

	rec := &DematPledgeRecord{
		PledgeID:          id,
		UserID:            userID,
		DematISIN:         isin,
		PledgedShares:     shares,
		MarketPriceINR:    priceINR,
		HaircutPct:        haircutPct,
		CollateralValINR:  collateralINR,
		AllowedMarginUSDT: marginUSDT,
		PledgedAt:         time.Now().UTC(),
	}

	e.pledges[id] = rec
	fmt.Printf("[Cross-Collateral] Pledged %d shares of %s (Val: ₹%.2f, Margin: %.2f USDT)\n",
		shares, isin, collateralINR, marginUSDT)
	return rec, nil
}
