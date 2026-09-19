// Package luld provides a production-grade LULD circuit breaker engine
// with LULD band calculation, market-wide circuit breakers (10%/15%/20%),
// session state management, and straddle-state trading halt logic.
package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Enums & Data Structures
// ---------------------------------------------------------------------------

// SessionState represents the current market session.
type SessionState string

const (
	SessionPreMarket        SessionState = "PRE_MARKET"
	SessionContinuous       SessionState = "CONTINUOUS"
	SessionClosingAuction   SessionState = "CLOSING_AUCTION"
	SessionHalted           SessionState = "HALTED"
	SessionClosed           SessionState = "CLOSED"
)

// SymbolTier classifies instruments for band width selection.
type SymbolTier int

const (
	Tier1NiftyFnO SymbolTier = iota // ±5% bands (S&P/Nifty equivalents)
	Tier2General                     // ±10% bands
)

// CircuitBreakerLevel represents market-wide circuit breaker thresholds.
type CircuitBreakerLevel int

const (
	CBLevel1 CircuitBreakerLevel = iota // 10% decline
	CBLevel2                             // 15% decline
	CBLevel3                             // 20% decline
)

// LULDBands represents the current price bands for a symbol.
type LULDBands struct {
	Symbol          string
	Tier            SymbolTier
	ReferencePrice  float64
	UpperBandPrice  float64
	LowerBandPrice  float64
	BandWidthPct    float64
	InStraddleState bool
	StraddleCount   int       // consecutive straddle events
	StateExpiresAt  time.Time
	LastUpdated     time.Time
}

// MarketWideCircuitBreaker tracks index-level circuit breaker state.
type MarketWideCircuitBreaker struct {
	Level           CircuitBreakerLevel
	ReferenceIndex  float64
	ThresholdPct    float64
	TriggerPrice    float64
	Triggered       bool
	TriggeredAt     time.Time
	HaltDuration    time.Duration
	ResumesAt       time.Time
}

// OrderValidationResult encapsulates the result of price validation.
type OrderValidationResult struct {
	Valid          bool
	RejectReason   string
	Symbol         string
	Price          float64
	UpperBand      float64
	LowerBand      float64
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	tier1BandPct         = 5.0
	tier2BandPct         = 10.0
	straddleTimeoutSecs  = 15 // seconds in straddle before halt
	maxStraddleCount     = 3  // max straddles before extended halt
	cbLevel1Pct          = 10.0
	cbLevel2Pct          = 15.0
	cbLevel3Pct          = 20.0
)

// ---------------------------------------------------------------------------
// Engine
// ---------------------------------------------------------------------------

// LULDCircuitBreakerEngine manages per-symbol LULD bands and market-wide
// circuit breakers with session state awareness.
type LULDCircuitBreakerEngine struct {
	mu              sync.RWMutex
	bands           map[string]*LULDBands
	session         SessionState
	marketBreakers  [3]*MarketWideCircuitBreaker
	sessionStarted  time.Time
}

// NewLULDCircuitBreakerEngine creates and initialises the engine.
func NewLULDCircuitBreakerEngine(indexReferencePrice float64) *LULDCircuitBreakerEngine {
	eng := &LULDCircuitBreakerEngine{
		bands:          make(map[string]*LULDBands),
		session:        SessionPreMarket,
		sessionStarted: time.Now().UTC(),
	}
	eng.initMarketBreakers(indexReferencePrice)
	return eng
}

// initMarketBreakers sets up the 10/15/20 % circuit breakers.
func (e *LULDCircuitBreakerEngine) initMarketBreakers(refIndex float64) {
	levels := [3]struct {
		pct      float64
		duration time.Duration
	}{
		{cbLevel1Pct, 45 * time.Minute},
		{cbLevel2Pct, 105 * time.Minute}, // 1h45m
		{cbLevel3Pct, 0},                 // remainder of day
	}

	for i, l := range levels {
		e.marketBreakers[i] = &MarketWideCircuitBreaker{
			Level:          CircuitBreakerLevel(i),
			ReferenceIndex: refIndex,
			ThresholdPct:   l.pct,
			TriggerPrice:   refIndex * (1.0 - l.pct/100.0),
			HaltDuration:   l.duration,
		}
	}
}

// ---------------------------------------------------------------------------
// Session Management
// ---------------------------------------------------------------------------

// TransitionSession changes the market session state.
func (e *LULDCircuitBreakerEngine) TransitionSession(newState SessionState) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	validTransitions := map[SessionState][]SessionState{
		SessionPreMarket:      {SessionContinuous, SessionClosed},
		SessionContinuous:     {SessionClosingAuction, SessionHalted, SessionClosed},
		SessionClosingAuction: {SessionClosed},
		SessionHalted:         {SessionContinuous, SessionClosed},
		SessionClosed:         {SessionPreMarket},
	}

	allowed, ok := validTransitions[e.session]
	if !ok {
		return fmt.Errorf("unknown current session state %s", e.session)
	}
	for _, s := range allowed {
		if s == newState {
			e.session = newState
			return nil
		}
	}
	return fmt.Errorf("invalid transition from %s to %s", e.session, newState)
}

// GetSession returns the current session state.
func (e *LULDCircuitBreakerEngine) GetSession() SessionState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.session
}

// ---------------------------------------------------------------------------
// LULD Band Management
// ---------------------------------------------------------------------------

// SetReferencePriceForSymbol calculates and stores LULD bands.
func (e *LULDCircuitBreakerEngine) SetReferencePriceForSymbol(
	symbol string,
	referencePrice float64,
	tier SymbolTier,
) (*LULDBands, error) {
	if referencePrice <= 0 {
		return nil, errors.New("reference price must be positive")
	}

	bandPct := tier1BandPct
	if tier == Tier2General {
		bandPct = tier2BandPct
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	b := &LULDBands{
		Symbol:         symbol,
		Tier:           tier,
		ReferencePrice: referencePrice,
		UpperBandPrice: referencePrice * (1.0 + bandPct/100.0),
		LowerBandPrice: referencePrice * (1.0 - bandPct/100.0),
		BandWidthPct:   bandPct,
		LastUpdated:    now,
	}
	e.bands[symbol] = b
	return b, nil
}

// GetBands returns the current bands for a symbol.
func (e *LULDCircuitBreakerEngine) GetBands(symbol string) (*LULDBands, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	b, ok := e.bands[symbol]
	if !ok {
		return nil, fmt.Errorf("no bands set for symbol %s", symbol)
	}
	return b, nil
}

// ---------------------------------------------------------------------------
// Order Price Validation
// ---------------------------------------------------------------------------

// ValidateOrderPrice checks if a price is within LULD bands.
func (e *LULDCircuitBreakerEngine) ValidateOrderPrice(
	symbol string,
	price float64,
) OrderValidationResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// If market is halted, reject all orders
	if e.session == SessionHalted || e.session == SessionClosed {
		return OrderValidationResult{
			Valid:        false,
			RejectReason: fmt.Sprintf("market session is %s", e.session),
			Symbol:       symbol,
			Price:        price,
		}
	}

	b, ok := e.bands[symbol]
	if !ok {
		// No bands → allow (new listing, pre-market, etc.)
		return OrderValidationResult{Valid: true, Symbol: symbol, Price: price}
	}

	if price > b.UpperBandPrice {
		return OrderValidationResult{
			Valid:        false,
			RejectReason: fmt.Sprintf("price %.4f exceeds Limit-Up %.4f", price, b.UpperBandPrice),
			Symbol:       symbol,
			Price:        price,
			UpperBand:    b.UpperBandPrice,
			LowerBand:    b.LowerBandPrice,
		}
	}
	if price < b.LowerBandPrice {
		return OrderValidationResult{
			Valid:        false,
			RejectReason: fmt.Sprintf("price %.4f breaches Limit-Down %.4f", price, b.LowerBandPrice),
			Symbol:       symbol,
			Price:        price,
			UpperBand:    b.UpperBandPrice,
			LowerBand:    b.LowerBandPrice,
		}
	}

	return OrderValidationResult{
		Valid:     true,
		Symbol:    symbol,
		Price:     price,
		UpperBand: b.UpperBandPrice,
		LowerBand: b.LowerBandPrice,
	}
}

// ---------------------------------------------------------------------------
// Straddle State Management
// ---------------------------------------------------------------------------

// EnterStraddleState marks a symbol as in straddle (price touching a band).
func (e *LULDCircuitBreakerEngine) EnterStraddleState(symbol string, now time.Time) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	b, ok := e.bands[symbol]
	if !ok {
		return fmt.Errorf("no bands for symbol %s", symbol)
	}

	b.InStraddleState = true
	b.StraddleCount++
	b.StateExpiresAt = now.Add(time.Duration(straddleTimeoutSecs) * time.Second)

	if b.StraddleCount >= maxStraddleCount {
		// Trigger trading halt for the symbol
		e.session = SessionHalted
	}
	return nil
}

// ExitStraddleState clears straddle for a symbol.
func (e *LULDCircuitBreakerEngine) ExitStraddleState(symbol string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if b, ok := e.bands[symbol]; ok {
		b.InStraddleState = false
	}
}

// ---------------------------------------------------------------------------
// Market-Wide Circuit Breakers
// ---------------------------------------------------------------------------

// CheckMarketCircuitBreaker evaluates if index level triggers a circuit breaker.
func (e *LULDCircuitBreakerEngine) CheckMarketCircuitBreaker(
	currentIndex float64,
	now time.Time,
) (*MarketWideCircuitBreaker, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	for i := 2; i >= 0; i-- { // check highest level first
		cb := e.marketBreakers[i]
		if cb.Triggered {
			continue
		}
		if currentIndex <= cb.TriggerPrice {
			cb.Triggered = true
			cb.TriggeredAt = now
			if cb.HaltDuration > 0 {
				cb.ResumesAt = now.Add(cb.HaltDuration)
			}
			e.session = SessionHalted
			return cb, true
		}
	}
	return nil, false
}

// GetMarketBreakers returns all circuit breaker states.
func (e *LULDCircuitBreakerEngine) GetMarketBreakers() [3]*MarketWideCircuitBreaker {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.marketBreakers
}

// IsMarketHalted returns whether the market is currently halted.
func (e *LULDCircuitBreakerEngine) IsMarketHalted() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.session == SessionHalted
}

// DeclinePercent computes the percentage decline from the reference.
func DeclinePercent(reference, current float64) float64 {
	if reference <= 0 {
		return 0
	}
	return math.Abs((reference - current) / reference * 100.0)
}
