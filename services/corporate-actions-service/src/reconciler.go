package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// CorporateActionType defines the classification of a corporate event.
type CorporateActionType string

const (
	ActionForwardSplit CorporateActionType = "FORWARD_SPLIT"
	ActionReverseSplit CorporateActionType = "REVERSE_SPLIT"
	ActionBonusIssue   CorporateActionType = "BONUS_ISSUE"
	ActionCashDividend CorporateActionType = "CASH_DIVIDEND"
)

// ReconciliationStatus indicates the state of corporate action reconciliation.
type ReconciliationStatus string

const (
	StatusPendingSnapshot ReconciliationStatus = "PENDING_SNAPSHOT"
	StatusAdjusted        ReconciliationStatus = "ADJUSTED"
	StatusReconciled      ReconciliationStatus = "RECONCILED"
	StatusDiscrepancy     ReconciliationStatus = "DISCREPANCY_DETECTED"
	StatusSettled         ReconciliationStatus = "SETTLED"
)

// UserPositionRecord represents a user's share balance pre- and post-adjustment.
type UserPositionRecord struct {
	UserID              string  `json:"user_id"`
	PreActionQuantity   float64 `json:"pre_action_quantity"`
	PostActionQuantity  float64 `json:"post_action_quantity"`
	FractionalRemainder float64 `json:"fractional_remainder"`
	CashInLieuUSD       float64 `json:"cash_in_lieu_usd"`
}

// OpenOrderRecord represents an active order book limit order to be adjusted.
type OpenOrderRecord struct {
	OrderID         string  `json:"order_id"`
	UserID          string  `json:"user_id"`
	Symbol          string  `json:"symbol"`
	OriginalQty     float64 `json:"original_quantity"`
	OriginalPrice   float64 `json:"original_price"`
	AdjustedQty     float64 `json:"adjusted_quantity"`
	AdjustedPrice   float64 `json:"adjusted_price"`
	OriginalNotional float64 `json:"original_notional"`
	AdjustedNotional float64 `json:"adjusted_notional"`
}

// DerivativeContractRecord represents an options or futures contract to be adjusted.
type DerivativeContractRecord struct {
	ContractSymbol string  `json:"contract_symbol"`
	OriginalStrike float64 `json:"original_strike"`
	AdjustedStrike float64 `json:"adjusted_strike"`
	OriginalLotSize float64 `json:"original_lot_size"`
	AdjustedLotSize float64 `json:"adjusted_lot_size"`
}

// DiscrepancyReport records discrepancies found during depository reconciliation.
type DiscrepancyReport struct {
	Field            string  `json:"field"`
	ExpectedValue    float64 `json:"expected_value"`
	ActualValue      float64 `json:"actual_value"`
	DiscrepancyDelta float64 `json:"discrepancy_delta"`
	Severity         string  `json:"severity"` // "WARNING", "CRITICAL"
	Details          string  `json:"details"`
}

// CorporateActionDefinition defines parameters of a corporate action.
type CorporateActionDefinition struct {
	ActionID             string              `json:"action_id"`
	Symbol               string              `json:"symbol"`
	Type                 CorporateActionType `json:"type"`
	SplitNumerator       float64             `json:"split_numerator"`   // e.g. 5 in 5:1 split, 3 in 3:1 bonus
	SplitDenominator     float64             `json:"split_denominator"` // e.g. 1 in 5:1 split, 1 in 3:1 bonus
	PreActionRefPrice    float64             `json:"pre_action_ref_price"`
	ExDateRefPrice       float64             `json:"ex_date_ref_price"`
	RecordDate           time.Time           `json:"record_date"`
	ExDate               time.Time           `json:"ex_date"`
	AllowFractionalShare bool                `json:"allow_fractional_share"` // If false, payout Cash-in-Lieu
}

// Multiplier calculates the effective factor by which shares are increased.
func (d *CorporateActionDefinition) Multiplier() float64 {
	if d.SplitDenominator <= 0 {
		return 1.0
	}
	switch d.Type {
	case ActionForwardSplit, ActionReverseSplit:
		return d.SplitNumerator / d.SplitDenominator
	case ActionBonusIssue:
		// e.g. 3:1 bonus issue gives 3 new shares for 1 held => total multiplier = 1 + 3/1 = 4.0
		return 1.0 + (d.SplitNumerator / d.SplitDenominator)
	default:
		return 1.0
	}
}

// ReconciliationReport captures the complete audit and verification findings.
type ReconciliationReport struct {
	ActionID                 string                 `json:"action_id"`
	Symbol                   string                 `json:"symbol"`
	Multiplier               float64                `json:"multiplier"`
	TotalPreActionShares     float64                `json:"total_pre_action_shares"`
	TotalPostActionShares    float64                `json:"total_post_action_shares"`
	TotalFractionalCashUSD   float64                `json:"total_fractional_cash_usd"`
	PreActionMarketCap       float64                `json:"pre_action_market_cap"`
	PostActionMarketCap      float64                `json:"post_action_market_cap"`
	DepositoryReportedShares float64                `json:"depository_reported_shares"`
	AdjustedPositionsCount   int                    `json:"adjusted_positions_count"`
	AdjustedOrdersCount      int                    `json:"adjusted_orders_count"`
	AdjustedContractsCount   int                    `json:"adjusted_contracts_count"`
	Status                   ReconciliationStatus   `json:"status"`
	Discrepancies            []DiscrepancyReport    `json:"discrepancies"`
	ReconciledAt             time.Time              `json:"reconciled_at"`
}

// CorporateActionReconciler performs adjustment and multi-way reconciliation.
type CorporateActionReconciler struct {
	mu           sync.RWMutex
	definitions  map[string]*CorporateActionDefinition
	userHoldings map[string]map[string]*UserPositionRecord      // actionID -> userID -> record
	openOrders   map[string]map[string]*OpenOrderRecord         // actionID -> orderID -> record
	derivatives  map[string]map[string]*DerivativeContractRecord // actionID -> contractSymbol -> record
	reports      map[string]*ReconciliationReport
}

// NewCorporateActionReconciler initializes a new corporate action reconciler.
func NewCorporateActionReconciler() *CorporateActionReconciler {
	return &CorporateActionReconciler{
		definitions:  make(map[string]*CorporateActionDefinition),
		userHoldings: make(map[string]map[string]*UserPositionRecord),
		openOrders:   make(map[string]map[string]*OpenOrderRecord),
		derivatives:  make(map[string]map[string]*DerivativeContractRecord),
		reports:      make(map[string]*ReconciliationReport),
	}
}

// RegisterAction adds a new corporate action definition.
func (r *CorporateActionReconciler) RegisterAction(action CorporateActionDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if action.ActionID == "" || action.Symbol == "" {
		return errors.New("actionID and symbol are required")
	}
	if action.SplitNumerator <= 0 || action.SplitDenominator <= 0 {
		return errors.New("split numerator and denominator must be positive")
	}

	r.definitions[action.ActionID] = &action
	r.userHoldings[action.ActionID] = make(map[string]*UserPositionRecord)
	r.openOrders[action.ActionID] = make(map[string]*OpenOrderRecord)
	r.derivatives[action.ActionID] = make(map[string]*DerivativeContractRecord)
	return nil
}

// RecordSnapshot records pre-action user holdings on the record date.
func (r *CorporateActionReconciler) RecordSnapshot(actionID, userID string, preShares float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	holdingsMap, exists := r.userHoldings[actionID]
	if !exists {
		return fmt.Errorf("action %s not registered", actionID)
	}

	holdingsMap[userID] = &UserPositionRecord{
		UserID:            userID,
		PreActionQuantity: preShares,
	}
	return nil
}

// RegisterOpenOrder records an active order book order to be adjusted on the ex-date.
func (r *CorporateActionReconciler) RegisterOpenOrder(actionID, orderID, userID, symbol string, qty, price float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ordersMap, exists := r.openOrders[actionID]
	if !exists {
		return fmt.Errorf("action %s not registered", actionID)
	}

	ordersMap[orderID] = &OpenOrderRecord{
		OrderID:          orderID,
		UserID:           userID,
		Symbol:           symbol,
		OriginalQty:      qty,
		OriginalPrice:    price,
		OriginalNotional: qty * price,
	}
	return nil
}

// RegisterDerivativeContract registers an options or futures contract to be adjusted.
func (r *CorporateActionReconciler) RegisterDerivativeContract(actionID, contractSymbol string, strike, lotSize float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	derivsMap, exists := r.derivatives[actionID]
	if !exists {
		return fmt.Errorf("action %s not registered", actionID)
	}

	derivsMap[contractSymbol] = &DerivativeContractRecord{
		ContractSymbol:  contractSymbol,
		OriginalStrike:  strike,
		OriginalLotSize: lotSize,
	}
	return nil
}

// ExecuteAdjustment applies stock split or bonus adjustments to all user positions, open orders, and derivatives.
func (r *CorporateActionReconciler) ExecuteAdjustment(actionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	action, ok := r.definitions[actionID]
	if !ok {
		return fmt.Errorf("action %s not found", actionID)
	}

	multiplier := action.Multiplier()
	exPrice := action.ExDateRefPrice
	if exPrice <= 0 && action.PreActionRefPrice > 0 {
		exPrice = action.PreActionRefPrice / multiplier
		action.ExDateRefPrice = exPrice
	}

	// 1. Adjust User Holdings
	holdingsMap := r.userHoldings[actionID]
	for _, rec := range holdingsMap {
		grossQty := rec.PreActionQuantity * multiplier

		if action.AllowFractionalShare {
			rec.PostActionQuantity = grossQty
			rec.FractionalRemainder = 0
			rec.CashInLieuUSD = 0
		} else {
			wholeShares := math.Floor(grossQty)
			fraction := grossQty - wholeShares
			rec.PostActionQuantity = wholeShares
			rec.FractionalRemainder = fraction
			rec.CashInLieuUSD = fraction * exPrice
		}
	}

	// 2. Adjust Open Orders (preserving notional value)
	ordersMap := r.openOrders[actionID]
	for _, ord := range ordersMap {
		adjQty := math.Round(ord.OriginalQty * multiplier)
		adjPrice := ord.OriginalPrice / multiplier

		ord.AdjustedQty = adjQty
		ord.AdjustedPrice = adjPrice
		ord.AdjustedNotional = adjQty * adjPrice
	}

	// 3. Adjust Derivative Contracts
	derivsMap := r.derivatives[actionID]
	for _, d := range derivsMap {
		d.AdjustedStrike = d.OriginalStrike / multiplier
		d.AdjustedLotSize = math.Round(d.OriginalLotSize * multiplier)
	}

	return nil
}

// ReconcileAndAudit reconciles adjusted internal positions against depository records and verifies invariants.
func (r *CorporateActionReconciler) ReconcileAndAudit(actionID string, depositoryTotalShares float64) (*ReconciliationReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	action, ok := r.definitions[actionID]
	if !ok {
		return nil, fmt.Errorf("action %s not found", actionID)
	}

	multiplier := action.Multiplier()
	holdingsMap := r.userHoldings[actionID]
	ordersMap := r.openOrders[actionID]
	derivsMap := r.derivatives[actionID]

	var totalPreShares, totalPostShares, totalFractionalShares, totalCashInLieu float64

	for _, rec := range holdingsMap {
		totalPreShares += rec.PreActionQuantity
		totalPostShares += rec.PostActionQuantity
		totalFractionalShares += rec.FractionalRemainder
		totalCashInLieu += rec.CashInLieuUSD
	}

	preMarketCap := totalPreShares * action.PreActionRefPrice
	postMarketCap := (totalPostShares + totalFractionalShares) * action.ExDateRefPrice

	var discrepancies []DiscrepancyReport

	// Check 1: Share Invariant Check (Total Post + Fractional == Total Pre * Multiplier)
	expectedTotalShares := totalPreShares * multiplier
	actualTotalShares := totalPostShares + totalFractionalShares
	deltaShares := math.Abs(actualTotalShares - expectedTotalShares)

	if deltaShares > 1e-4 {
		discrepancies = append(discrepancies, DiscrepancyReport{
			Field:            "TOTAL_SHARES_CONSERVATION",
			ExpectedValue:    expectedTotalShares,
			ActualValue:      actualTotalShares,
			DiscrepancyDelta: deltaShares,
			Severity:         "CRITICAL",
			Details:          "Mismatch between pre-split multiplier calculation and post-adjustment share sum",
		})
	}

	// Check 2: Depository Settlement Match
	if depositoryTotalShares > 0 {
		depDelta := math.Abs(totalPostShares - depositoryTotalShares)
		// Small tolerance for rounding / fractional cash-in-lieu
		if depDelta > 1.0 {
			discrepancies = append(discrepancies, DiscrepancyReport{
				Field:            "DEPOSITORY_LEDGER_MATCH",
				ExpectedValue:    depositoryTotalShares,
				ActualValue:      totalPostShares,
				DiscrepancyDelta: depDelta,
				Severity:         "CRITICAL",
				Details:          fmt.Sprintf("Internal ledger shares (%.2f) differ from Depository clearing feed (%.2f)", totalPostShares, depositoryTotalShares),
			})
		}
	}

	// Check 3: Order Notional Conservation
	for _, ord := range ordersMap {
		notionalDiff := math.Abs(ord.AdjustedNotional - ord.OriginalNotional)
		if ord.OriginalNotional > 0 && (notionalDiff/ord.OriginalNotional) > 0.01 { // > 1% drift due to rounding
			discrepancies = append(discrepancies, DiscrepancyReport{
				Field:            fmt.Sprintf("ORDER_%s_NOTIONAL_DRIFT", ord.OrderID),
				ExpectedValue:    ord.OriginalNotional,
				ActualValue:      ord.AdjustedNotional,
				DiscrepancyDelta: notionalDiff,
				Severity:         "WARNING",
				Details:          "Open order notional changed by > 1% after rounding",
			})
		}
	}

	status := StatusReconciled
	if len(discrepancies) > 0 {
		for _, d := range discrepancies {
			if d.Severity == "CRITICAL" {
				status = StatusDiscrepancy
				break
			}
		}
	}

	report := &ReconciliationReport{
		ActionID:                 actionID,
		Symbol:                   action.Symbol,
		Multiplier:               multiplier,
		TotalPreActionShares:     totalPreShares,
		TotalPostActionShares:    totalPostShares,
		TotalFractionalCashUSD:   totalCashInLieu,
		PreActionMarketCap:       preMarketCap,
		PostActionMarketCap:      postMarketCap,
		DepositoryReportedShares: depositoryTotalShares,
		AdjustedPositionsCount:   len(holdingsMap),
		AdjustedOrdersCount:      len(ordersMap),
		AdjustedContractsCount:   len(derivsMap),
		Status:                   status,
		Discrepancies:            discrepancies,
		ReconciledAt:             time.Now().UTC(),
	}

	r.reports[actionID] = report
	return report, nil
}

// GetReconciliationReport returns the generated audit report.
func (r *CorporateActionReconciler) GetReconciliationReport(actionID string) (*ReconciliationReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rep, ok := r.reports[actionID]
	if !ok {
		return nil, errors.New("report not found")
	}
	return rep, nil
}
