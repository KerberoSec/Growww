package main

import "testing"

func TestAnnounceAndAdvance(t *testing.T) {
	svc := NewCorpActionsService()
	ca := &CorpAction{ID: "ca1", Symbol: "INFY", Type: "DIVIDEND", DividendPerShare: 18.0}
	if err := svc.Announce(ca); err != nil { t.Fatal(err) }
	got, _ := svc.GetAction("ca1")
	if got.Phase != Announced { t.Errorf("expected ANNOUNCED, got %s", got.Phase) }
	svc.AdvancePhase("ca1", Paid)
	got2, _ := svc.GetAction("ca1")
	if got2.Phase != Paid { t.Errorf("expected PAID, got %s", got2.Phase) }
}

func TestStockSplit(t *testing.T) {
	svc := NewCorpActionsService()
	ca := &CorpAction{ID: "sp1", Symbol: "ITC", Type: "SPLIT", Ratio: 5.0}
	svc.Announce(ca)
	newQty, newPrice, err := svc.ComputeSplitAdjustment("sp1", 100, 500)
	if err != nil { t.Fatal(err) }
	if newQty != 500 { t.Errorf("5:1 split on 100 should give 500, got %f", newQty) }
	if newPrice != 100 { t.Errorf("price should be 100 after split, got %f", newPrice) }
}

func TestDuplicateAnnounce(t *testing.T) {
	svc := NewCorpActionsService()
	ca := &CorpAction{ID: "d1", Symbol: "HDFC"}
	svc.Announce(ca)
	if err := svc.Announce(&CorpAction{ID: "d1", Symbol: "HDFC"}); err == nil { t.Error("expected error") }
}
