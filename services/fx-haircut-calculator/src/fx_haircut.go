package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// ──────────────────────────────────────────────────────────────────────────────
// Domain types
// ──────────────────────────────────────────────────────────────────────────────

// Currency represents a supported currency code (ISO 4217).
type Currency string

const (
	CurrencyUSD Currency = "USD"
	CurrencyINR Currency = "INR"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyJPY Currency = "JPY"
	CurrencySGD Currency = "SGD"
	CurrencyAED Currency = "AED"
	CurrencyCHF Currency = "CHF"
)

// FXRate represents a real-time foreign exchange rate.
type FXRate struct {
	BaseCurrency  Currency  `json:"base_currency"`
	QuoteCurrency Currency  `json:"quote_currency"`
	Rate          float64   `json:"rate"`
	Timestamp     time.Time `json:"timestamp"`
	Source        string    `json:"source"` // "REUTERS", "BLOOMBERG", "RBI_FBIL"
}

// CollateralType classifies the collateral asset.
type CollateralType string

const (
	CollateralCash     CollateralType = "CASH"
	CollateralBond     CollateralType = "SOVEREIGN_BOND"
	CollateralEquity   CollateralType = "EQUITY"
	CollateralCrypto   CollateralType = "CRYPTO"
	CollateralGold     CollateralType = "GOLD"
)

// HaircutProfile defines the haircut configuration for a collateral type.
type HaircutProfile struct {
	CollateralType  CollateralType `json:"collateral_type"`
	BaseCurrency    Currency       `json:"base_currency"`
	BaseHaircutPct  float64        `json:"base_haircut_pct"`  // Base haircut percentage (0-100)
	VolatilityAddon float64        `json:"volatility_addon"`  // Additional haircut for FX volatility
	LiquidityAddon  float64        `json:"liquidity_addon"`   // Illiquidity premium
	ConcentrationCap float64       `json:"concentration_cap"` // Max % of total collateral pool
}

// HaircutResult contains the computed haircut for a collateral position.
type HaircutResult struct {
	RequestID         string         `json:"request_id"`
	CollateralType    CollateralType `json:"collateral_type"`
	OriginalCurrency  Currency       `json:"original_currency"`
	TargetCurrency    Currency       `json:"target_currency"`
	OriginalAmountE8  uint64         `json:"original_amount_e8"`
	FXRate            float64        `json:"fx_rate"`
	ConvertedAmountE8 uint64         `json:"converted_amount_e8"`
	TotalHaircutPct   float64        `json:"total_haircut_pct"`
	HaircutAmountE8   uint64         `json:"haircut_amount_e8"`
	NetCollateralE8   uint64         `json:"net_collateral_e8"`
	ComputedAt        time.Time      `json:"computed_at"`
}

// ──────────────────────────────────────────────────────────────────────────────
// FX Rate Provider
// ──────────────────────────────────────────────────────────────────────────────

// FXRateProvider serves FX rates with staleness checks.
type FXRateProvider struct {
	mu           sync.RWMutex
	rates        map[string]FXRate // key: "BASE-QUOTE"
	maxStaleness time.Duration
}

// NewFXRateProvider creates a rate provider with the given staleness threshold.
func NewFXRateProvider(maxStaleness time.Duration) *FXRateProvider {
	return &FXRateProvider{
		rates:        make(map[string]FXRate),
		maxStaleness: maxStaleness,
	}
}

func rateKey(base, quote Currency) string {
	return string(base) + "-" + string(quote)
}

// UpdateRate publishes a new FX rate.
func (p *FXRateProvider) UpdateRate(rate FXRate) error {
	if rate.Rate <= 0 {
		return errors.New("FX rate must be positive")
	}
	if rate.BaseCurrency == rate.QuoteCurrency {
		return errors.New("base and quote currency must differ")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	key := rateKey(rate.BaseCurrency, rate.QuoteCurrency)
	p.rates[key] = rate

	// Also store inverse
	inverseKey := rateKey(rate.QuoteCurrency, rate.BaseCurrency)
	p.rates[inverseKey] = FXRate{
		BaseCurrency:  rate.QuoteCurrency,
		QuoteCurrency: rate.BaseCurrency,
		Rate:          1.0 / rate.Rate,
		Timestamp:     rate.Timestamp,
		Source:        rate.Source,
	}

	return nil
}

// GetRate retrieves the current FX rate, checking for staleness.
func (p *FXRateProvider) GetRate(base, quote Currency) (*FXRate, error) {
	if base == quote {
		return &FXRate{
			BaseCurrency:  base,
			QuoteCurrency: quote,
			Rate:          1.0,
			Timestamp:     time.Now().UTC(),
			Source:        "IDENTITY",
		}, nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	key := rateKey(base, quote)
	rate, exists := p.rates[key]
	if !exists {
		return nil, fmt.Errorf("no FX rate for %s/%s", base, quote)
	}

	if time.Since(rate.Timestamp) > p.maxStaleness {
		return nil, fmt.Errorf("FX rate %s/%s is stale (age: %s, max: %s)",
			base, quote, time.Since(rate.Timestamp), p.maxStaleness)
	}

	return &rate, nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Haircut Calculator
// ──────────────────────────────────────────────────────────────────────────────

// HaircutCalculator computes FX-adjusted haircuts on multi-currency collateral.
type HaircutCalculator struct {
	mu       sync.RWMutex
	fxProvider *FXRateProvider
	profiles   map[CollateralType]HaircutProfile
}

// NewHaircutCalculator creates a calculator with default profiles.
func NewHaircutCalculator(fxProvider *FXRateProvider) *HaircutCalculator {
	calc := &HaircutCalculator{
		fxProvider: fxProvider,
		profiles:   make(map[CollateralType]HaircutProfile),
	}
	calc.loadDefaultProfiles()
	return calc
}

func (c *HaircutCalculator) loadDefaultProfiles() {
	defaults := []HaircutProfile{
		{CollateralCash, CurrencyUSD, 0.0, 0.5, 0.0, 100.0},
		{CollateralBond, CurrencyUSD, 2.0, 1.0, 0.5, 80.0},
		{CollateralEquity, CurrencyINR, 15.0, 3.0, 2.0, 50.0},
		{CollateralCrypto, CurrencyUSD, 25.0, 10.0, 5.0, 30.0},
		{CollateralGold, CurrencyUSD, 5.0, 2.0, 1.0, 40.0},
	}
	for _, p := range defaults {
		c.profiles[p.CollateralType] = p
	}
}

// SetProfile registers or updates a haircut profile.
func (c *HaircutCalculator) SetProfile(profile HaircutProfile) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.profiles[profile.CollateralType] = profile
}

// GetProfile retrieves a haircut profile by collateral type.
func (c *HaircutCalculator) GetProfile(ct CollateralType) (*HaircutProfile, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	p, ok := c.profiles[ct]
	if !ok {
		return nil, fmt.Errorf("no haircut profile for %s", ct)
	}
	return &p, nil
}

// ComputeHaircut calculates the FX-adjusted haircut for a collateral position.
func (c *HaircutCalculator) ComputeHaircut(
	requestID string,
	collateralType CollateralType,
	originalCurrency Currency,
	targetCurrency Currency,
	amountE8 uint64,
) (*HaircutResult, error) {
	if amountE8 == 0 {
		return nil, errors.New("amount must be > 0")
	}

	c.mu.RLock()
	profile, exists := c.profiles[collateralType]
	c.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("unsupported collateral type: %s", collateralType)
	}

	// Get FX rate
	fxRate, err := c.fxProvider.GetRate(originalCurrency, targetCurrency)
	if err != nil {
		return nil, fmt.Errorf("FX rate lookup failed: %w", err)
	}

	// Convert amount to target currency
	originalFloat := float64(amountE8)
	convertedFloat := originalFloat * fxRate.Rate
	convertedE8 := uint64(math.Round(convertedFloat))

	// Compute total haircut percentage
	totalHaircutPct := profile.BaseHaircutPct + profile.VolatilityAddon + profile.LiquidityAddon

	// Apply FX volatility surcharge for cross-currency conversions
	if originalCurrency != targetCurrency {
		totalHaircutPct += profile.VolatilityAddon // Double FX vol addon for cross-currency
	}

	// Cap at 100%
	if totalHaircutPct > 100.0 {
		totalHaircutPct = 100.0
	}

	haircutFloat := convertedFloat * (totalHaircutPct / 100.0)
	haircutE8 := uint64(math.Round(haircutFloat))
	netE8 := convertedE8 - haircutE8

	result := &HaircutResult{
		RequestID:         requestID,
		CollateralType:    collateralType,
		OriginalCurrency:  originalCurrency,
		TargetCurrency:    targetCurrency,
		OriginalAmountE8:  amountE8,
		FXRate:            fxRate.Rate,
		ConvertedAmountE8: convertedE8,
		TotalHaircutPct:   totalHaircutPct,
		HaircutAmountE8:   haircutE8,
		NetCollateralE8:   netE8,
		ComputedAt:        time.Now().UTC(),
	}

	fmt.Printf("[FXHaircut] %s: %s %d (e8) @ FX %.6f -> %s %d (e8) | Haircut: %.2f%% = %d (e8) | Net: %d (e8)\n",
		requestID, originalCurrency, amountE8, fxRate.Rate, targetCurrency, convertedE8, totalHaircutPct, haircutE8, netE8)

	return result, nil
}
