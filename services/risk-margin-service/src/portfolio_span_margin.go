package main

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

// InstrumentType specifies the asset class of a derivative or spot position.
type InstrumentType string

const (
	InstrumentSpot    InstrumentType = "SPOT"
	InstrumentFutures InstrumentType = "FUTURES"
	InstrumentCall    InstrumentType = "CALL"
	InstrumentPut     InstrumentType = "PUT"
)

// SPANPosition defines a single position for SPAN and portfolio margin calculations.
type SPANPosition struct {
	PositionID     string         `json:"position_id"`
	UserID         string         `json:"user_id"`
	Symbol         string         `json:"symbol"`
	CommodityClass string         `json:"commodity_class"` // e.g. "BTC", "ETH", "NIFTY"
	Instrument     InstrumentType `json:"instrument_type"`
	Quantity       float64        `json:"quantity"` // Positive for Long, Negative for Short
	EntryPrice     float64        `json:"entry_price"`
	CurrentPrice   float64        `json:"current_price"`
	Strike         float64        `json:"strike,omitempty"`
	ExpiryDays     float64        `json:"expiry_days,omitempty"`
	Delta          float64        `json:"delta"`
	Gamma          float64        `json:"gamma"`
	Vega           float64        `json:"vega"`
	ImpliedVol     float64        `json:"implied_vol"`
}

// SPANRiskParameters holds risk parameters for a commodity class.
type SPANRiskParameters struct {
	PriceScanRangePct      float64 `json:"price_scan_range_pct"`      // e.g. 0.10 (10% move)
	VolScanRangePct        float64 `json:"vol_scan_range_pct"`        // e.g. 0.15 (15% vol shift)
	ExtremeMoveMultiplier  float64 `json:"extreme_move_multiplier"`   // e.g. 2.0 (2x scan range)
	ExtremeLossFraction    float64 `json:"extreme_loss_fraction"`     // e.g. 0.35 (35% of extreme loss)
	IntraSpreadRateUSD     float64 `json:"intra_spread_rate_usd"`     // Charge per calendar spread contract
	ShortOptionMinRateUSD  float64 `json:"short_option_min_rate_usd"` // Minimum charge per short option
	LiquidityAddonPct      float64 `json:"liquidity_addon_pct"`       // Add-on for concentrated positions
}

// InterCommoditySpreadRule defines correlation offset between two commodity classes.
type InterCommoditySpreadRule struct {
	ClassA           string  `json:"class_a"`
	ClassB           string  `json:"class_b"`
	DeltaRatio       float64 `json:"delta_ratio"`        // Ratio of Class A to Class B delta
	CreditPercentage float64 `json:"credit_percentage"`  // Margin credit % e.g. 0.70 (70% offset)
}

// SPANScenario defines price and volatility shifts for each of the 16 standard SPAN scenarios.
type SPANScenario struct {
	ScenarioNumber int     `json:"scenario_number"`
	PriceShiftMult float64 `json:"price_shift_mult"` // Multiplier on PriceScanRange (-1.0 to 1.0, or extreme)
	VolShiftMult   float64 `json:"vol_shift_mult"`   // Multiplier on VolScanRange (-1.0 to 1.0)
	LossWeight     float64 `json:"loss_weight"`      // Fraction of loss accounted (1.0 or ExtremeLossFraction)
	Description    string  `json:"description"`
}

// Default16SPANScenarios returns standard CME / SEBI 16 SPAN scenario specifications.
func Default16SPANScenarios() []SPANScenario {
	return []SPANScenario{
		{1, 0.0, 1.0, 1.0, "Price Unchanged, Vol Up"},
		{2, 0.0, -1.0, 1.0, "Price Unchanged, Vol Down"},
		{3, 1.0 / 3.0, 1.0, 1.0, "Price Up 1/3, Vol Up"},
		{4, 1.0 / 3.0, -1.0, 1.0, "Price Up 1/3, Vol Down"},
		{5, -1.0 / 3.0, 1.0, 1.0, "Price Down 1/3, Vol Up"},
		{6, -1.0 / 3.0, -1.0, 1.0, "Price Down 1/3, Vol Down"},
		{7, 2.0 / 3.0, 1.0, 1.0, "Price Up 2/3, Vol Up"},
		{8, 2.0 / 3.0, -1.0, 1.0, "Price Up 2/3, Vol Down"},
		{9, -2.0 / 3.0, 1.0, 1.0, "Price Down 2/3, Vol Up"},
		{10, -2.0 / 3.0, -1.0, 1.0, "Price Down 2/3, Vol Down"},
		{11, 1.0, 1.0, 1.0, "Price Up 3/3, Vol Up"},
		{12, 1.0, -1.0, 1.0, "Price Up 3/3, Vol Down"},
		{13, -1.0, 1.0, 1.0, "Price Down 3/3, Vol Up"},
		{14, -1.0, -1.0, 1.0, "Price Down 3/3, Vol Down"},
		{15, 2.0, 0.0, 0.35, "Extreme Price Up (2x range, 35% loss covered)"},
		{16, -2.0, 0.0, 0.35, "Extreme Price Down (2x range, 35% loss covered)"},
	}
}

// SPANMarginResult contains the detailed breakdown of the SPAN calculation.
type SPANMarginResult struct {
	UserID                  string             `json:"user_id"`
	ScanningRisk            float64            `json:"scanning_risk_usd"`
	IntraCommodityCharge    float64            `json:"intra_commodity_charge_usd"`
	InterCommodityCredit    float64            `json:"inter_commodity_credit_usd"`
	ShortOptionMinimum      float64            `json:"short_option_minimum_usd"`
	LiquidityAddon          float64            `json:"liquidity_addon_usd"`
	GrossMarginRequired     float64            `json:"gross_margin_required_usd"`
	NetSPANMarginRequired   float64            `json:"net_span_margin_required_usd"`
	MarginSavingsUSD        float64            `json:"margin_savings_usd"`
	MarginSavingsPct        float64            `json:"margin_savings_pct"`
	WorstScenarioIndex      int                `json:"worst_scenario_index"`
	WorstScenarioDesc       string             `json:"worst_scenario_desc"`
	ScenarioLosses          [16]float64        `json:"scenario_losses"`
	NetDeltaByCommodity     map[string]float64 `json:"net_delta_by_commodity"`
	CalculatedAt            time.Time          `json:"calculated_at"`
}

// PortfolioSPANEngine computes SPAN margins and inter-commodity offsets.
type PortfolioSPANEngine struct {
	mu           sync.RWMutex
	riskParams   map[string]SPANRiskParameters       // commodityClass -> params
	spreadRules  []InterCommoditySpreadRule
	scenarios    []SPANScenario
}

// NewPortfolioSPANEngine creates a new SPAN engine.
func NewPortfolioSPANEngine() *PortfolioSPANEngine {
	return &PortfolioSPANEngine{
		riskParams:  make(map[string]SPANRiskParameters),
		spreadRules: make([]InterCommoditySpreadRule, 0),
		scenarios:   Default16SPANScenarios(),
	}
}

// SetRiskParameters configures risk parameters for a commodity class.
func (e *PortfolioSPANEngine) SetRiskParameters(commodityClass string, params SPANRiskParameters) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.riskParams[commodityClass] = params
}

// AddInterCommoditySpreadRule adds an offset rule between two correlated commodities.
func (e *PortfolioSPANEngine) AddInterCommoditySpreadRule(rule InterCommoditySpreadRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.spreadRules = append(e.spreadRules, rule)
}

// ComputePortfolioMargin calculates SPAN margin with scanning risk, calendar spreads, SOM, and portfolio offsets.
func (e *PortfolioSPANEngine) ComputePortfolioMargin(userID string, positions []SPANPosition) (*SPANMarginResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(positions) == 0 {
		return &SPANMarginResult{
			UserID:       userID,
			CalculatedAt: time.Now().UTC(),
		}, nil
	}

	// 1. Group positions by commodity class and aggregate delta/vega
	byCommodity := make(map[string][]SPANPosition)
	netDeltaByCommodity := make(map[string]float64)
	var shortOptionCount float64

	for _, pos := range positions {
		byCommodity[pos.CommodityClass] = append(byCommodity[pos.CommodityClass], pos)

		// Effective delta in units of underlying
		posDelta := pos.Quantity
		if pos.Instrument == InstrumentCall || pos.Instrument == InstrumentPut {
			posDelta = pos.Quantity * pos.Delta
			if pos.Quantity < 0 {
				shortOptionCount += math.Abs(pos.Quantity)
			}
		}
		netDeltaByCommodity[pos.CommodityClass] += posDelta
	}

	// 2. Evaluate 16 SPAN scenarios across all positions
	var scenarioLosses [16]float64
	worstLoss := -math.MaxFloat64
	worstIdx := 0

	for sIdx, sc := range e.scenarios {
		totalPortfolioPnL := 0.0

		for _, pos := range positions {
			params, ok := e.riskParams[pos.CommodityClass]
			if !ok {
				// Default fallback risk parameters
				params = SPANRiskParameters{
					PriceScanRangePct:     0.10,
					VolScanRangePct:       0.20,
					ExtremeMoveMultiplier: 2.0,
					ExtremeLossFraction:   0.35,
					ShortOptionMinRateUSD: 50.0,
				}
			}

			priceShiftPct := sc.PriceShiftMult * params.PriceScanRangePct
			volShiftPct := sc.VolShiftMult * params.VolScanRangePct

			// Calculate instrument PnL under scenario
			var posPnL float64
			switch pos.Instrument {
			case InstrumentSpot, InstrumentFutures:
				// Linear delta payoff
				priceChange := pos.CurrentPrice * priceShiftPct
				posPnL = pos.Quantity * priceChange

			case InstrumentCall, InstrumentPut:
				// First-order (Delta, Vega) + second-order (Gamma) Taylor expansion
				priceChange := pos.CurrentPrice * priceShiftPct
				volChange := pos.ImpliedVol * volShiftPct

				deltaPnL := pos.Delta * priceChange
				gammaPnL := 0.5 * pos.Gamma * priceChange * priceChange
				vegaPnL := pos.Vega * (volChange * 100.0) // Vega is per 1% vol

				posPnL = pos.Quantity * (deltaPnL + gammaPnL + vegaPnL)
			}

			totalPortfolioPnL += posPnL
		}

		// Scenario loss = -PnL, scaled by scenario loss weight
		scenarioLoss := -totalPortfolioPnL * sc.LossWeight
		scenarioLosses[sIdx] = scenarioLoss

		if scenarioLoss > worstLoss {
			worstLoss = scenarioLoss
			worstIdx = sIdx
		}
	}

	scanningRisk := math.Max(0.0, worstLoss)

	// 3. Intra-Commodity (Calendar Spread) Charge
	intraCommodityCharge := 0.0
	for commClass, commPositions := range byCommodity {
		params := e.riskParams[commClass]
		// Find spreads across different expiry dates
		longQtyByExpiry := make(map[float64]float64)
		shortQtyByExpiry := make(map[float64]float64)

		for _, p := range commPositions {
			if p.Quantity > 0 {
				longQtyByExpiry[p.ExpiryDays] += p.Quantity
			} else {
				shortQtyByExpiry[p.ExpiryDays] += math.Abs(p.Quantity)
			}
		}

		if len(longQtyByExpiry) > 1 || len(shortQtyByExpiry) > 1 {
			// Compute calendar spread pairs
			spreadPairs := 0.0
			var totalLong, totalShort float64
			for _, q := range longQtyByExpiry {
				totalLong += q
			}
			for _, q := range shortQtyByExpiry {
				totalShort += q
			}
			spreadPairs = math.Min(totalLong, totalShort)
			intraCommodityCharge += spreadPairs * params.IntraSpreadRateUSD
		}
	}

	// 4. Inter-Commodity Spread Credit (Portfolio Margin Offset)
	interCommodityCredit := 0.0
	for _, rule := range e.spreadRules {
		deltaA, okA := netDeltaByCommodity[rule.ClassA]
		deltaB, okB := netDeltaByCommodity[rule.ClassB]

		if !okA || !okB {
			continue
		}

		// Offset occurs when one commodity is net long and the other is net short
		if deltaA*deltaB < 0 {
			absA := math.Abs(deltaA)
			absB := math.Abs(deltaB)

			ratio := rule.DeltaRatio
			if ratio <= 0 {
				ratio = 1.0
			}

			// Paired delta contracts
			matchedDeltaA := math.Min(absA, absB*ratio)
			matchedDeltaB := matchedDeltaA / ratio

			// Calculate margin relief based on individual scanning risks
			paramsA := e.riskParams[rule.ClassA]
			paramsB := e.riskParams[rule.ClassB]

			marginA := matchedDeltaA * paramsA.PriceScanRangePct * 100.0 // notional proxy
			marginB := matchedDeltaB * paramsB.PriceScanRangePct * 100.0
			baseMargin := marginA + marginB

			credit := baseMargin * rule.CreditPercentage
			interCommodityCredit += credit
		}
	}

	// 5. Short Option Minimum (SOM)
	somCharge := 0.0
	for commClass, commPositions := range byCommodity {
		params := e.riskParams[commClass]
		rate := params.ShortOptionMinRateUSD
		if rate <= 0 {
			rate = 25.0
		}
		for _, p := range commPositions {
			if (p.Instrument == InstrumentCall || p.Instrument == InstrumentPut) && p.Quantity < 0 {
				somCharge += math.Abs(p.Quantity) * rate
			}
		}
	}

	// 6. Liquidity Add-on
	liquidityAddon := 0.0
	for commClass, commPositions := range byCommodity {
		params := e.riskParams[commClass]
		if params.LiquidityAddonPct > 0 {
			var grossNotional float64
			for _, p := range commPositions {
				grossNotional += math.Abs(p.Quantity) * p.CurrentPrice
			}
			if grossNotional > 1_000_000 { // Large concentration threshold
				liquidityAddon += (grossNotional - 1_000_000) * params.LiquidityAddonPct
			}
		}
	}

	// Gross margin requirement (without portfolio offset credit)
	grossMargin := scanningRisk + intraCommodityCharge + liquidityAddon
	if grossMargin < somCharge {
		grossMargin = somCharge
	}

	// Net SPAN Margin after applying offsets
	netMargin := math.Max(0.0, scanningRisk+intraCommodityCharge-interCommodityCredit) + liquidityAddon
	if netMargin < somCharge {
		netMargin = somCharge
	}

	savingsUSD := math.Max(0.0, grossMargin-netMargin)
	savingsPct := 0.0
	if grossMargin > 0 {
		savingsPct = (savingsUSD / grossMargin) * 100.0
	}

	return &SPANMarginResult{
		UserID:                userID,
		ScanningRisk:          scanningRisk,
		IntraCommodityCharge:  intraCommodityCharge,
		InterCommodityCredit:  interCommodityCredit,
		ShortOptionMinimum:    somCharge,
		LiquidityAddon:        liquidityAddon,
		GrossMarginRequired:   grossMargin,
		NetSPANMarginRequired: netMargin,
		MarginSavingsUSD:      savingsUSD,
		MarginSavingsPct:      savingsPct,
		WorstScenarioIndex:    worstIdx + 1,
		WorstScenarioDesc:     e.scenarios[worstIdx].Description,
		ScenarioLosses:        scenarioLosses,
		NetDeltaByCommodity:   netDeltaByCommodity,
		CalculatedAt:          time.Now().UTC(),
	}, nil
}

// -----------------------------------------------------------------------------
// SPAN Risk Calculator Daemon
// -----------------------------------------------------------------------------

// MarginRiskAlert is triggered when a user's margin utilization breaches warning levels.
type MarginRiskAlert struct {
	UserID            string    `json:"user_id"`
	EquityUSD         float64   `json:"equity_usd"`
	RequiredMarginUSD float64   `json:"required_margin_usd"`
	MarginUtilization float64   `json:"margin_utilization_pct"`
	AlertLevel        string    `json:"alert_level"` // "NORMAL", "WARNING", "MARGIN_CALL", "LIQUIDATION"
	Timestamp         time.Time `json:"timestamp"`
}

// SPANRiskDaemonConfig holds daemon configuration settings.
type SPANRiskDaemonConfig struct {
	EvaluationInterval    time.Duration
	WarningThresholdPct   float64 // e.g. 80.0%
	MarginCallThresholdPct float64 // e.g. 90.0%
	LiquidationThresholdPct float64 // e.g. 100.0%
	WorkerPoolSize        int
}

// SPANRiskDaemon runs continuous risk assessment and margin offset evaluation.
type SPANRiskDaemon struct {
	mu            sync.RWMutex
	spanEngine    *PortfolioSPANEngine
	config        SPANRiskDaemonConfig
	userPositions map[string][]SPANPosition // userID -> positions
	userEquities  map[string]float64        // userID -> equityUSD
	alertsChan    chan MarginRiskAlert
	evalQueue     chan string               // userID queue for immediate evaluation
	ctx           context.Context
	cancel        context.CancelFunc
	isRunning     bool
	wg            sync.WaitGroup
}

// NewSPANRiskDaemon creates an instance of the daemon.
func NewSPANRiskDaemon(engine *PortfolioSPANEngine, cfg SPANRiskDaemonConfig) *SPANRiskDaemon {
	if cfg.EvaluationInterval <= 0 {
		cfg.EvaluationInterval = 500 * time.Millisecond
	}
	if cfg.WarningThresholdPct <= 0 {
		cfg.WarningThresholdPct = 80.0
	}
	if cfg.MarginCallThresholdPct <= 0 {
		cfg.MarginCallThresholdPct = 90.0
	}
	if cfg.LiquidationThresholdPct <= 0 {
		cfg.LiquidationThresholdPct = 100.0
	}
	if cfg.WorkerPoolSize <= 0 {
		cfg.WorkerPoolSize = 4
	}

	return &SPANRiskDaemon{
		spanEngine:    engine,
		config:        cfg,
		userPositions: make(map[string][]SPANPosition),
		userEquities:  make(map[string]float64),
		alertsChan:    make(chan MarginRiskAlert, 1000),
		evalQueue:     make(chan string, 1000),
	}
}

// Alerts returns the channel for streaming margin risk alerts.
func (d *SPANRiskDaemon) Alerts() <-chan MarginRiskAlert {
	return d.alertsChan
}

// SetUserPortfolio registers or updates user positions and collateral equity.
func (d *SPANRiskDaemon) SetUserPortfolio(userID string, positions []SPANPosition, equityUSD float64) {
	d.mu.Lock()
	d.userPositions[userID] = positions
	d.userEquities[userID] = equityUSD
	d.mu.Unlock()

	// Trigger high-priority immediate evaluation
	select {
	case d.evalQueue <- userID:
	default:
	}
}

// UpdateMarketPrices updates mark price and implied volatility across all open positions.
func (d *SPANRiskDaemon) UpdateMarketPrices(symbol string, markPrice, iv float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for userID, positions := range d.userPositions {
		changed := false
		for i := range positions {
			if positions[i].Symbol == symbol {
				positions[i].CurrentPrice = markPrice
				if iv > 0 {
					positions[i].ImpliedVol = iv
				}
				changed = true
			}
		}
		if changed {
			select {
			case d.evalQueue <- userID:
			default:
			}
		}
	}
}

// EvaluateUser calculates margin and checks threshold breach for a single user.
func (d *SPANRiskDaemon) EvaluateUser(userID string) (*SPANMarginResult, *MarginRiskAlert, error) {
	d.mu.RLock()
	positions, exists := d.userPositions[userID]
	equity := d.userEquities[userID]
	d.mu.RUnlock()

	if !exists || len(positions) == 0 {
		return nil, nil, nil
	}

	res, err := d.spanEngine.ComputePortfolioMargin(userID, positions)
	if err != nil {
		return nil, nil, err
	}

	var utilization float64
	if equity > 0 {
		utilization = (res.NetSPANMarginRequired / equity) * 100.0
	} else if res.NetSPANMarginRequired > 0 {
		utilization = 999.99
	}

	var alert *MarginRiskAlert
	var level string
	switch {
	case utilization >= d.config.LiquidationThresholdPct:
		level = "LIQUIDATION"
	case utilization >= d.config.MarginCallThresholdPct:
		level = "MARGIN_CALL"
	case utilization >= d.config.WarningThresholdPct:
		level = "WARNING"
	default:
		level = "NORMAL"
	}

	if level != "NORMAL" {
		alert = &MarginRiskAlert{
			UserID:            userID,
			EquityUSD:         equity,
			RequiredMarginUSD: res.NetSPANMarginRequired,
			MarginUtilization: utilization,
			AlertLevel:        level,
			Timestamp:         time.Now().UTC(),
		}
	}

	return res, alert, nil
}

// Start launches background evaluation workers and periodic sweep loop.
func (d *SPANRiskDaemon) Start(parentCtx context.Context) error {
	d.mu.Lock()
	if d.isRunning {
		d.mu.Unlock()
		return errors.New("daemon is already running")
	}
	d.ctx, d.cancel = context.WithCancel(parentCtx)
	d.isRunning = true
	d.mu.Unlock()

	// 1. Start worker pool for processing immediate evaluation queue
	for i := 0; i < d.config.WorkerPoolSize; i++ {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			for {
				select {
				case <-d.ctx.Done():
					return
				case userID := <-d.evalQueue:
					_, alert, err := d.EvaluateUser(userID)
					if err == nil && alert != nil {
						select {
						case d.alertsChan <- *alert:
						default:
						}
					}
				}
			}
		}()
	}

	// 2. Periodic sweep across all registered users
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		ticker := time.NewTicker(d.config.EvaluationInterval)
		defer ticker.Stop()

		for {
			select {
			case <-d.ctx.Done():
				return
			case <-ticker.C:
				d.mu.RLock()
				userIDs := make([]string, 0, len(d.userPositions))
				for uid := range d.userPositions {
					userIDs = append(userIDs, uid)
				}
				d.mu.RUnlock()

				for _, uid := range userIDs {
					select {
					case d.evalQueue <- uid:
					default:
					}
				}
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down daemon workers.
func (d *SPANRiskDaemon) Stop() {
	d.mu.Lock()
	if !d.isRunning {
		d.mu.Unlock()
		return
	}
	d.cancel()
	d.isRunning = false
	d.mu.Unlock()

	d.wg.Wait()
}
