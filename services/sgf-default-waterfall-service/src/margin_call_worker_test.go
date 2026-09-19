package main

import "testing"

func TestSGFMarginCallWorker(t *testing.T) {
	worker := NewSGFMarginCallWorker()
	m1 := &ClearingMember{MemberID: "CM_01", CollateralINR: 8000000, RequiredMarginINR: 10000000}
	m2 := &ClearingMember{MemberID: "CM_02", CollateralINR: 15000000, RequiredMarginINR: 10000000}

	worker.RegisterMember(m1)
	worker.RegisterMember(m2)

	calls := worker.EvaluateMarginCalls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 margin call, got %d", len(calls))
	}
	if calls[0].MemberID != "CM_01" {
		t.Errorf("expected CM_01, got %s", calls[0].MemberID)
	}
	if calls[0].MarginDeficitINR != 2000000 {
		t.Errorf("expected 20L deficit, got %f", calls[0].MarginDeficitINR)
	}
}
