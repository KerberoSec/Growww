package src

import (
	"testing"
)

func TestMoney_ExactArithmetic(t *testing.T) {
	m1, err := NewMoney("INR", 100, 750000000) // 100.75 INR
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m2, err := NewMoney("INR", 50, 500000000) // 50.50 INR
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 100.75 + 50.50 = 151.25 INR
	sum, err := m1.Add(m2)
	if err != nil {
		t.Fatalf("add error: %v", err)
	}

	if sum.Units != 151 {
		t.Errorf("expected 151 units, got %d", sum.Units)
	}
	if sum.Nanos != 250000000 {
		t.Errorf("expected 250,000,000 nanos, got %d", sum.Nanos)
	}

	// Cross-currency addition should fail
	usd, _ := NewMoney("USD", 10, 0)
	_, err = m1.Add(usd)
	if err == nil {
		t.Errorf("expected currency mismatch error")
	}
}

func TestFractionalShare_PrecisionAndFormatting(t *testing.T) {
	share := FromFloatShares(1.543210)
	if share.MicroShares != 1543210 {
		t.Errorf("expected 1543210 micro-shares, got %d", share.MicroShares)
	}
	if share.Format() != "1.543210" {
		t.Errorf("expected '1.543210', got '%s'", share.Format())
	}
	if share.ToFloat() != 1.54321 {
		t.Errorf("expected float 1.54321, got %f", share.ToFloat())
	}
}

func TestISIN_Validation(t *testing.T) {
	validISIN, err := NewISIN("INE002A01018") // Reliance Industries
	if err != nil {
		t.Fatalf("expected valid ISIN: %v", err)
	}
	if validISIN.CountryCode() != "IN" {
		t.Errorf("expected country code IN, got %s", validISIN.CountryCode())
	}

	// Invalid ISINs
	invalidISINs := []string{
		"INE002A0101",   // too short
		"INE002A010189", // too long
		"12E002A01018",  // country code numbers instead of letters
		"ine002a0101A",  // last char letter instead of digit
	}

	for _, bad := range invalidISINs {
		_, err := NewISIN(bad)
		if err == nil {
			t.Errorf("expected '%s' to be invalid ISIN", bad)
		}
	}
}

func TestOrderStateMachine_Transitions(t *testing.T) {
	// Valid lifecycle
	if err := ValidateOrderTransition(OrderStatusPending, OrderStatusRouted); err != nil {
		t.Errorf("pending -> routed failed: %v", err)
	}
	if err := ValidateOrderTransition(OrderStatusRouted, OrderStatusPartiallyFilled); err != nil {
		t.Errorf("routed -> partially_filled failed: %v", err)
	}
	if err := ValidateOrderTransition(OrderStatusPartiallyFilled, OrderStatusFilled); err != nil {
		t.Errorf("partially_filled -> filled failed: %v", err)
	}

	// Illegal transitions
	if err := ValidateOrderTransition(OrderStatusFilled, OrderStatusPending); err == nil {
		t.Errorf("expected error transitioning from terminal FILLED state")
	}
	if err := ValidateOrderTransition(OrderStatusCancelled, OrderStatusRouted); err == nil {
		t.Errorf("expected error transitioning from terminal CANCELLED state")
	}
}

func TestDvPStateMachine_Transitions(t *testing.T) {
	// Valid happy path
	if err := ValidateDvPTransition(DvPStatusCreated, DvPStatusEscrowLocked); err != nil {
		t.Errorf("created -> escrow_locked failed: %v", err)
	}
	if err := ValidateDvPTransition(DvPStatusEscrowLocked, DvPStatusCommitted); err != nil {
		t.Errorf("escrow_locked -> committed failed: %v", err)
	}
	if err := ValidateDvPTransition(DvPStatusCommitted, DvPStatusSettled); err != nil {
		t.Errorf("committed -> settled failed: %v", err)
	}

	// Terminal state
	if err := ValidateDvPTransition(DvPStatusSettled, DvPStatusFailed); err == nil {
		t.Errorf("expected error transitioning from terminal SETTLED state")
	}
}
