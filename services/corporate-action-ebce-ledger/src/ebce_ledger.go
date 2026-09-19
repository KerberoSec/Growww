package main

import (
	"fmt"
	"sync"
	"time"
)

type CorporateActionType string
const (
	Earnings   CorporateActionType = "EARNINGS"
	Bonus      CorporateActionType = "BONUS"
	Conversion CorporateActionType = "CONVERSION"
	ExDividend CorporateActionType = "EX_DIVIDEND"
)

type CorporateAction struct {
	ActionID   string
	Symbol     string
	ActionType CorporateActionType
	RecordDate time.Time
	ExDate     time.Time
	Ratio      float64 // e.g. 1:2 bonus = 0.5, dividend per share
	Status     string  // ANNOUNCED, APPLIED, COMPLETED
}

type PortfolioAdjustment struct {
	UserID       string
	Symbol       string
	ActionID     string
	OldQuantity  float64
	NewQuantity  float64
	CashCredit   float64
}

type EBCELedger struct {
	mu          sync.Mutex
	actions     []*CorporateAction
	adjustments []*PortfolioAdjustment
}

func NewEBCELedger() *EBCELedger { return &EBCELedger{} }

func (l *EBCELedger) RegisterAction(action *CorporateAction) {
	l.mu.Lock()
	defer l.mu.Unlock()
	action.Status = "ANNOUNCED"
	l.actions = append(l.actions, action)
}

func (l *EBCELedger) ApplyBonus(action *CorporateAction, userID string, currentQty float64) (*PortfolioAdjustment, error) {
	if action.ActionType != Bonus { return nil, fmt.Errorf("not a bonus action") }
	if action.Ratio <= 0 { return nil, fmt.Errorf("invalid ratio") }
	l.mu.Lock()
	defer l.mu.Unlock()
	bonusShares := currentQty * action.Ratio
	adj := &PortfolioAdjustment{
		UserID: userID, Symbol: action.Symbol, ActionID: action.ActionID,
		OldQuantity: currentQty, NewQuantity: currentQty + bonusShares,
	}
	l.adjustments = append(l.adjustments, adj)
	action.Status = "APPLIED"
	return adj, nil
}

func (l *EBCELedger) ApplyDividend(action *CorporateAction, userID string, qty float64) (*PortfolioAdjustment, error) {
	if action.ActionType != ExDividend { return nil, fmt.Errorf("not a dividend action") }
	l.mu.Lock()
	defer l.mu.Unlock()
	cash := qty * action.Ratio
	adj := &PortfolioAdjustment{
		UserID: userID, Symbol: action.Symbol, ActionID: action.ActionID,
		OldQuantity: qty, NewQuantity: qty, CashCredit: cash,
	}
	l.adjustments = append(l.adjustments, adj)
	action.Status = "APPLIED"
	return adj, nil
}
