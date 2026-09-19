package main
import "testing"
func TestGenerateAndSubmit(t *testing.T) {
	svc := NewReportingService()
	r, err := svc.GenerateReport("QUARTERLY_TRADE", "Q3-2026", "SEBI"); if err != nil { t.Fatal(err) }
	if r.Status != "GENERATED" { t.Error("expected GENERATED") }
	svc.SubmitReport(r.ID)
	reports := svc.GetReports()
	if reports[0].Status != "SUBMITTED" { t.Error("expected SUBMITTED") }
}
func TestEmptyType(t *testing.T) {
	svc := NewReportingService(); _, err := svc.GenerateReport("", "Q1", "RBI")
	if err == nil { t.Error("expected error for empty type") }
}
