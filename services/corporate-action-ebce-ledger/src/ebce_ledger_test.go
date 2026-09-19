package main

import (
	"testing"
	"time"
)

func TestBonusApplication(t *testing.T) {
	ledger := NewEBCELedger()
	action := &CorporateAction{ActionID: "ca1", Symbol: "RELIANCE", ActionType: Bonus, Ratio: 1.0, RecordDate: time.Now(), ExDate: time.Now()}
	ledger.RegisterAction(action)
	adj, err := ledger.ApplyBonus(action, "usr_1", 100)
	if err != nil { t.Fatal(err) }
	if adj.NewQuantity != 200 { t.Errorf("1:1 bonus on 100 shares should give 200, got %f", adj.NewQuantity) }
}

func TestDividendApplication(t *testing.T) {
	ledger := NewEBCELedger()
	action := &CorporateAction{ActionID: "ca2", Symbol: "TCS", ActionType: ExDividend, Ratio: 15.0}
	ledger.RegisterAction(action)
	adj, err := ledger.ApplyDividend(action, "usr_2", 50)
	if err != nil { t.Fatal(err) }
	if adj.CashCredit != 750.0 { t.Errorf("expected 750 cash, got %f", adj.CashCredit) }
	if adj.NewQuantity != 50 { t.Error("dividend should not change quantity") }
}
