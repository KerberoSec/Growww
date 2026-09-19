package main
import "testing"
func TestReferralFlow(t *testing.T) {
	svc := NewReferralService(); code, _ := svc.GenerateCode("usr_A")
	if len(code) < 5 { t.Error("code too short") }
	svc.GenerateCode("usr_B")
	if err := svc.ApplyReferral("usr_B", code); err != nil { t.Fatal(err) }
	svc.CreditCommission("usr_A", 500.0)
	if svc.users["usr_A"].TotalEarnings != 500 { t.Error("expected 500 earnings") }
}
func TestSelfReferral(t *testing.T) {
	svc := NewReferralService(); code, _ := svc.GenerateCode("usr_X")
	if err := svc.ApplyReferral("usr_X", code); err == nil { t.Error("expected self-referral error") }
}
func TestInvalidCode(t *testing.T) {
	svc := NewReferralService(); svc.GenerateCode("usr_Z")
	if err := svc.ApplyReferral("usr_Y", "INVALID"); err == nil { t.Error("expected invalid code error") }
}
