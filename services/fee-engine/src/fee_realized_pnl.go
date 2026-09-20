package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// AssetClass defines instrument categories for fee and tax computation.
type AssetClass string

const (
	AssetEquityDelivery AssetClass = "EQUITY_DELIVERY"
	AssetEquityIntraday AssetClass = "EQUITY_INTRADAY"
	AssetFutures        AssetClass = "FUTURES"
	AssetOptions        AssetClass = "OPTIONS"
	AssetCryptoSpot     AssetClass = "CRYPTO_SPOT"
	AssetCryptoPerp     AssetClass = "CRYPTO_PERP"
)

// OrderRole indicates maker or taker execution role.
type OrderRole string

const (
	RoleMaker OrderRole = "MAKER"
	RoleTaker OrderRole = "TAKER"
)

// AccountingMethod defines the cost basis method for realized PnL calculation.
type AccountingMethod string

const (
	MethodFIFO AccountingMethod = "FIFO"
	MethodLIFO AccountingMethod = "LIFO"
	MethodWAC  AccountingMethod = "WAC" // Weighted Average Cost
)

// TradeSide represents buy or sell side.
type TradeSide string

const (
	TradeBuy  TradeSide = "BUY"
	TradeSell TradeSide = "SELL"
)

// ComprehensiveFeeBreakdown details all components of trade fees and statutory levies.
type ComprehensiveFeeBreakdown struct {
	ExecutionID       string     `json:"execution_id"`
	UserID            string     `json:"user_id"`
	Symbol            string     `json:"symbol"`
	AssetClass        AssetClass `json:"asset_class"`
	Side              TradeSide  `json:"side"`
	Quantity          float64    `json:"quantity"`
	Price             float64    `json:"price"`
	NotionalTurnover  float64    `json:"notional_turnover"`
	OrderRole         OrderRole  `json:"order_role"`
	BaseBrokerageRate float64    `json:"base_brokerage_rate_bps"`
	GrossBrokerage    float64    `json:"gross_brokerage"`
	TokenDiscount     float64    `json:"token_discount"`      // Discount applied if paid with utility token
	NetBrokerage      float64    `json:"net_brokerage"`
	STT_CTT           float64    `json:"stt_ctt"`             // Securities / Commodities Transaction Tax
	ExchangeCharges   float64    `json:"exchange_charges"`    // Exchange turnover fees
	SEBICharges       float64    `json:"sebi_charges"`        // Regulatory turnover charges
	StampDuty         float64    `json:"stamp_duty"`          // State / national stamp duty
	GST               float64    `json:"gst"`                 // 18% GST on brokerage + exchange + SEBI charges
	TotalStatutory    float64    `json:"total_statutory"`     // STT + Exchange + SEBI + Stamp + GST
	TotalFeesPayable  float64    `json:"total_fees_payable"`  // NetBrokerage + TotalStatutory
	SGFContribution   float64    `json:"sgf_contribution"`    // 25% of net brokerage allocated to Settlement Guarantee Fund
	TreasuryVault     float64    `json:"treasury_vault"`      // Remainder to treasury
	CalculatedAt      time.Time  `json:"calculated_at"`
}

// VIPTier defines maker/taker fee rates based on 30-day trailing volume.
type VIPTier struct {
	TierLevel     int     `json:"tier_level"`
	MinVolumeUSD  float64 `json:"min_volume_usd"`
	MakerFeeBps   float64 `json:"maker_fee_bps"`
	TakerFeeBps   float64 `json:"taker_fee_bps"`
}

// FeeScheduleConfig holds exchange fee and tax settings.
type FeeScheduleConfig struct {
	VIPTiers             []VIPTier
	UtilityTokenDiscount float64 // e.g. 0.25 (25% discount)
	SGFReservePct        float64 // e.g. 0.25 (25% of exchange net fee)
	GSTRate              float64 // 0.18 (18%)
	SEBIChargeRate       float64 // 0.000001 (₹10 per crore)
}

// DefaultFeeScheduleConfig returns Growww's zero-fee schedule (0.00% maker, 0.00% taker across all tiers).
func DefaultFeeScheduleConfig() FeeScheduleConfig {
	return FeeScheduleConfig{
		VIPTiers: []VIPTier{
			{TierLevel: 0, MinVolumeUSD: 0, MakerFeeBps: 0.0, TakerFeeBps: 0.0},
			{TierLevel: 1, MinVolumeUSD: 100_000, MakerFeeBps: 0.0, TakerFeeBps: 0.0},
			{TierLevel: 2, MinVolumeUSD: 1_000_000, MakerFeeBps: 0.0, TakerFeeBps: 0.0},
			{TierLevel: 3, MinVolumeUSD: 5_000_000, MakerFeeBps: 0.0, TakerFeeBps: 0.0},
			{TierLevel: 4, MinVolumeUSD: 20_000_000, MakerFeeBps: 0.0, TakerFeeBps: 0.0},
			{TierLevel: 5, MinVolumeUSD: 50_000_000, MakerFeeBps: 0.0, TakerFeeBps: 0.0},
			{TierLevel: 6, MinVolumeUSD: 100_000_000, MakerFeeBps: 0.0, TakerFeeBps: 0.0},
		},
		UtilityTokenDiscount: 0.0,  // No discount needed — already 0%
		SGFReservePct:        0.0,  // 0.00% — zero fee, zero reserve
		GSTRate:              0.0,  // 0.00% — no GST on platform fee
		SEBIChargeRate:       0.0,  // 0.00% — pass-through only; platform does not charge
	}
}

// ComprehensiveFeeEngine computes tiered fees, statutory levies, and SGF waterfalls.
type ComprehensiveFeeEngine struct {
	mu     sync.RWMutex
	config FeeScheduleConfig
}

// NewComprehensiveFeeEngine creates a new fee calculation engine.
func NewComprehensiveFeeEngine(cfg FeeScheduleConfig) *ComprehensiveFeeEngine {
	return &ComprehensiveFeeEngine{config: cfg}
}

// GetVIPTier returns the applicable fee tier for a 30-day turnover.
func (e *ComprehensiveFeeEngine) GetVIPTier(turnoverUSD float64) VIPTier {
	e.mu.RLock()
	defer e.mu.RUnlock()

	matchedTier := e.config.VIPTiers[0]
	for _, tier := range e.config.VIPTiers {
		if turnoverUSD >= tier.MinVolumeUSD {
			matchedTier = tier
		}
	}
	return matchedTier
}

// CalculateFees computes full fee breakdown enforcing Growww's 0.00% zero-fee policy.
// All brokerage, statutory levies (STT, GST, SEBI, stamp duty), and exchange charges are 0.
func (e *ComprehensiveFeeEngine) CalculateFees(
	execID, userID, symbol string,
	assetClass AssetClass,
	side TradeSide,
	quantity, price float64,
	role OrderRole,
	user30DayVolumeUSD float64,
	payWithUtilityToken bool,
) ComprehensiveFeeBreakdown {
	e.mu.RLock()
	defer e.mu.RUnlock()

	notional := quantity * price

	// Growww zero-fee policy: all fees are 0.00%
	return ComprehensiveFeeBreakdown{
		ExecutionID:       execID,
		UserID:            userID,
		Symbol:            symbol,
		AssetClass:        assetClass,
		Side:              side,
		Quantity:          quantity,
		Price:             price,
		NotionalTurnover:  notional,
		OrderRole:         role,
		BaseBrokerageRate: 0.0,
		GrossBrokerage:    0.0,
		TokenDiscount:     0.0,
		NetBrokerage:      0.0,
		STT_CTT:           0.0, // 0.00% — zero statutory levy
		ExchangeCharges:   0.0, // 0.00% — zero exchange charges
		SEBICharges:       0.0, // 0.00% — zero regulatory fee
		StampDuty:         0.0, // 0.00% — zero stamp duty
		GST:               0.0, // 0.00% — no GST on zero fee
		TotalStatutory:    0.0,
		TotalFeesPayable:  0.0,
		SGFContribution:   0.0,
		TreasuryVault:     0.0,
		CalculatedAt:      time.Now().UTC(),
	}
}

// -----------------------------------------------------------------------------
// Realized PnL Engine (FIFO, LIFO, Weighted Average Cost)
// -----------------------------------------------------------------------------

// TradeLot represents an unclosed or partially closed position lot.
type TradeLot struct {
	LotID        string    `json:"lot_id"`
	TradeID      string    `json:"trade_id"`
	Side         TradeSide `json:"side"` // TradeBuy (Long) or TradeSell (Short)
	Quantity     float64   `json:"original_quantity"`
	RemainingQty float64   `json:"remaining_quantity"`
	Price        float64   `json:"price"`
	FeePerUnit   float64   `json:"fee_per_unit"`
	Timestamp    time.Time `json:"timestamp"`
}

// ClosedLotMatch records a matched lot segment against a closing execution.
type ClosedLotMatch struct {
	LotID            string    `json:"lot_id"`
	OpenedAt         time.Time `json:"opened_at"`
	MatchedQuantity  float64   `json:"matched_quantity"`
	EntryPrice       float64   `json:"entry_price"`
	AllocatedEntryFee float64  `json:"allocated_entry_fee"`
	HoldingPeriod    time.Duration `json:"holding_period"`
	IsLongTerm       bool      `json:"is_long_term"` // > 365 days
}

// RealizedPnLEvent records the outcome of closing a position.
type RealizedPnLEvent struct {
	EventID          string           `json:"event_id"`
	UserID           string           `json:"user_id"`
	Symbol           string           `json:"symbol"`
	ClosingSide      TradeSide        `json:"closing_side"` // SELL to close Long, BUY to close Short
	ClosedQuantity   float64          `json:"closed_quantity"`
	ExitPrice        float64          `json:"exit_price"`
	ExitFeeAllocated float64          `json:"exit_fee_allocated"`
	GrossRealizedPnL float64          `json:"gross_realized_pnl"`
	NetRealizedPnL   float64          `json:"net_realized_pnl"`
	TotalEntryFees   float64          `json:"total_entry_fees"`
	TotalFees        float64          `json:"total_fees"`
	ClosedLots       []ClosedLotMatch `json:"closed_lots"`
	Timestamp        time.Time        `json:"timestamp"`
}

// UserSymbolPosition tracks active open lots and cumulative PnL for a user symbol.
type UserSymbolPosition struct {
	UserID              string     `json:"user_id"`
	Symbol              string     `json:"symbol"`
	ActiveSide          TradeSide  `json:"active_side"` // BUY (Net Long) or SELL (Net Short)
	NetOpenQuantity     float64    `json:"net_open_quantity"`
	WeightedAverageCost float64    `json:"weighted_average_cost"`
	OpenLots            []TradeLot `json:"open_lots"`
	CumulativeGrossPnL  float64    `json:"cumulative_gross_pnl"`
	CumulativeNetPnL    float64    `json:"cumulative_net_pnl"`
	CumulativeFeesPaid  float64    `json:"cumulative_fees_paid"`
}

// RealizedPnLEngine manages multi-asset tax-lot matching and realized PnL calculations.
type RealizedPnLEngine struct {
	mu               sync.RWMutex
	accountingMethod AccountingMethod
	positions        map[string]*UserSymbolPosition // "userID:symbol" -> Position
	pnlHistory       []RealizedPnLEvent
}

// NewRealizedPnLEngine creates an engine with the chosen accounting method.
func NewRealizedPnLEngine(method AccountingMethod) *RealizedPnLEngine {
	return &RealizedPnLEngine{
		accountingMethod: method,
		positions:        make(map[string]*UserSymbolPosition),
		pnlHistory:       make([]RealizedPnLEvent, 0),
	}
}

// SetAccountingMethod switches between FIFO, LIFO, and WAC.
func (e *RealizedPnLEngine) SetAccountingMethod(method AccountingMethod) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.accountingMethod = method
}

func posKey(userID, symbol string) string {
	return fmt.Sprintf("%s:%s", userID, symbol)
}

// ProcessTradeExecution processes a trade fill, matching against open lots and computing realized PnL.
func (e *RealizedPnLEngine) ProcessTradeExecution(
	tradeID, userID, symbol string,
	side TradeSide,
	quantity, price, totalFees float64,
	timestamp time.Time,
) (*RealizedPnLEvent, error) {
	if quantity <= 0 || price <= 0 {
		return nil, errors.New("quantity and price must be strictly positive")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	key := posKey(userID, symbol)
	pos, exists := e.positions[key]
	if !exists {
		pos = &UserSymbolPosition{
			UserID:   userID,
			Symbol:   symbol,
			OpenLots: make([]TradeLot, 0),
		}
		e.positions[key] = pos
	}

	pos.CumulativeFeesPaid += totalFees
	feePerUnit := totalFees / quantity

	// Case 1: No open positions, or adding to existing position on the same side
	if pos.NetOpenQuantity == 0 || pos.ActiveSide == side {
		pos.ActiveSide = side
		lot := TradeLot{
			LotID:        fmt.Sprintf("LOT-%s-%d", tradeID, len(pos.OpenLots)+1),
			TradeID:      tradeID,
			Side:         side,
			Quantity:     quantity,
			RemainingQty: quantity,
			Price:        price,
			FeePerUnit:   feePerUnit,
			Timestamp:    timestamp,
		}
		pos.OpenLots = append(pos.OpenLots, lot)

		// Update Weighted Average Cost
		totalCost := pos.WeightedAverageCost * pos.NetOpenQuantity + price * quantity
		pos.NetOpenQuantity += quantity
		pos.WeightedAverageCost = totalCost / pos.NetOpenQuantity
		return nil, nil // No position closed, no realized PnL event
	}

	// Case 2: Opposite side trade -> CLOSING position (generates Realized PnL)
	qtyToClose := math.Min(pos.NetOpenQuantity, quantity)
	remainingToClose := qtyToClose

	var closedMatches []ClosedLotMatch
	var totalEntryFeesAllocated float64
	var grossPnL float64

	switch e.accountingMethod {
	case MethodWAC:
		// Weighted Average Cost matching
		wacEntryPrice := pos.WeightedAverageCost
		// Average entry fee per unit
		var avgEntryFeePerUnit float64
		if pos.NetOpenQuantity > 0 {
			var totalFeePool float64
			for _, l := range pos.OpenLots {
				totalFeePool += l.RemainingQty * l.FeePerUnit
			}
			avgEntryFeePerUnit = totalFeePool / pos.NetOpenQuantity
		}

		if pos.ActiveSide == TradeBuy { // Closing Long
			grossPnL = (price - wacEntryPrice) * qtyToClose
		} else { // Closing Short
			grossPnL = (wacEntryPrice - price) * qtyToClose
		}

		totalEntryFeesAllocated = avgEntryFeePerUnit * qtyToClose
		closedMatches = append(closedMatches, ClosedLotMatch{
			LotID:             "WAC-AGGREGATE-LOT",
			OpenedAt:          pos.OpenLots[0].Timestamp,
			MatchedQuantity:   qtyToClose,
			EntryPrice:        wacEntryPrice,
			AllocatedEntryFee: totalEntryFeesAllocated,
			HoldingPeriod:     timestamp.Sub(pos.OpenLots[0].Timestamp),
			IsLongTerm:        timestamp.Sub(pos.OpenLots[0].Timestamp) > 365*24*time.Hour,
		})

		// Reduce open lots proportionally
		ratio := (pos.NetOpenQuantity - qtyToClose) / pos.NetOpenQuantity
		for i := range pos.OpenLots {
			pos.OpenLots[i].RemainingQty *= ratio
		}

	case MethodFIFO:
		// First-In, First-Out: Consume lots from beginning of slice
		for i := 0; i < len(pos.OpenLots) && remainingToClose > 0; i++ {
			lot := &pos.OpenLots[i]
			if lot.RemainingQty <= 0 {
				continue
			}

			matched := math.Min(lot.RemainingQty, remainingToClose)
			lot.RemainingQty -= matched
			remainingToClose -= matched

			entryFee := matched * lot.FeePerUnit
			totalEntryFeesAllocated += entryFee

			var lotPnL float64
			if pos.ActiveSide == TradeBuy {
				lotPnL = (price - lot.Price) * matched
			} else {
				lotPnL = (lot.Price - price) * matched
			}
			grossPnL += lotPnL

			holding := timestamp.Sub(lot.Timestamp)
			closedMatches = append(closedMatches, ClosedLotMatch{
				LotID:             lot.LotID,
				OpenedAt:          lot.Timestamp,
				MatchedQuantity:   matched,
				EntryPrice:        lot.Price,
				AllocatedEntryFee: entryFee,
				HoldingPeriod:     holding,
				IsLongTerm:        holding > 365*24*time.Hour,
			})
		}

	case MethodLIFO:
		// Last-In, First-Out: Consume lots from end of slice
		for i := len(pos.OpenLots) - 1; i >= 0 && remainingToClose > 0; i-- {
			lot := &pos.OpenLots[i]
			if lot.RemainingQty <= 0 {
				continue
			}

			matched := math.Min(lot.RemainingQty, remainingToClose)
			lot.RemainingQty -= matched
			remainingToClose -= matched

			entryFee := matched * lot.FeePerUnit
			totalEntryFeesAllocated += entryFee

			var lotPnL float64
			if pos.ActiveSide == TradeBuy {
				lotPnL = (price - lot.Price) * matched
			} else {
				lotPnL = (lot.Price - price) * matched
			}
			grossPnL += lotPnL

			holding := timestamp.Sub(lot.Timestamp)
			closedMatches = append(closedMatches, ClosedLotMatch{
				LotID:             lot.LotID,
				OpenedAt:          lot.Timestamp,
				MatchedQuantity:   matched,
				EntryPrice:        lot.Price,
				AllocatedEntryFee: entryFee,
				HoldingPeriod:     holding,
				IsLongTerm:        holding > 365*24*time.Hour,
			})
		}
	}

	// Filter out completely consumed lots
	activeLots := make([]TradeLot, 0, len(pos.OpenLots))
	for _, l := range pos.OpenLots {
		if l.RemainingQty > 1e-9 {
			activeLots = append(activeLots, l)
		}
	}
	pos.OpenLots = activeLots

	// Allocate exit fee to the closed portion
	exitFeeAllocated := (qtyToClose / quantity) * totalFees
	totalFeesOnTrade := totalEntryFeesAllocated + exitFeeAllocated
	netPnL := grossPnL - totalFeesOnTrade

	pos.CumulativeGrossPnL += grossPnL
	pos.CumulativeNetPnL += netPnL
	pos.NetOpenQuantity -= qtyToClose

	// If position reversed (trade quantity was larger than open position)
	excessQty := quantity - qtyToClose
	if excessQty > 0 {
		pos.ActiveSide = side
		pos.NetOpenQuantity = excessQty
		pos.WeightedAverageCost = price
		pos.OpenLots = []TradeLot{
			{
				LotID:        fmt.Sprintf("LOT-%s-REVERSED", tradeID),
				TradeID:      tradeID,
				Side:         side,
				Quantity:     excessQty,
				RemainingQty: excessQty,
				Price:        price,
				FeePerUnit:   feePerUnit,
				Timestamp:    timestamp,
			},
		}
	} else if pos.NetOpenQuantity <= 1e-9 {
		pos.NetOpenQuantity = 0
		pos.WeightedAverageCost = 0
		pos.OpenLots = make([]TradeLot, 0)
	}

	event := RealizedPnLEvent{
		EventID:          fmt.Sprintf("PNL-%s-%d", tradeID, time.Now().UnixNano()),
		UserID:           userID,
		Symbol:           symbol,
		ClosingSide:      side,
		ClosedQuantity:   qtyToClose,
		ExitPrice:        price,
		ExitFeeAllocated: exitFeeAllocated,
		GrossRealizedPnL: grossPnL,
		NetRealizedPnL:   netPnL,
		TotalEntryFees:   totalEntryFeesAllocated,
		TotalFees:        totalFeesOnTrade,
		ClosedLots:       closedMatches,
		Timestamp:        timestamp,
	}

	e.pnlHistory = append(e.pnlHistory, event)
	return &event, nil
}

// GetUserPosition retrieves current active position and PnL status.
func (e *RealizedPnLEngine) GetUserPosition(userID, symbol string) (*UserSymbolPosition, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	pos, exists := e.positions[posKey(userID, symbol)]
	if !exists {
		return nil, errors.New("no position found for user and symbol")
	}
	return pos, nil
}
