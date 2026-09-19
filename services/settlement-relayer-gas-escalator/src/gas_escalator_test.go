package main
import ("testing";"time")
func TestSubmitAndConfirm(t *testing.T) {
	esc := NewGasEscalator(20, 500, 25)
	esc.Submit(1, "0xabc"); esc.Confirm(1)
	if esc.pending[1].Status != "CONFIRMED" { t.Error("expected confirmed") }
}
func TestEscalation(t *testing.T) {
	esc := NewGasEscalator(100, 500, 50)
	esc.Submit(2, "0xdef"); esc.pending[2].SubmittedAt = time.Now().Add(-2 * time.Minute)
	escalated := esc.EscalateStuck(1 * time.Minute)
	if len(escalated) != 1 { t.Fatal("expected 1 escalated") }
	if escalated[0].GasPrice != 150 { t.Errorf("expected 150, got %d", escalated[0].GasPrice) }
}
