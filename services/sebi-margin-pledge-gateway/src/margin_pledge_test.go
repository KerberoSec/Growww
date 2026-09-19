package main

import (
	"testing"
)

func TestInitiatePledge_Success(t *testing.T) {
	gw := NewPledgeGateway()

	req, err := gw.InitiatePledge("PLG-001", "CLIENT-A", "INE002A01018", "RELIANCE", 100, 2450.00, 25.0, DepositoryNSDL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Status != PledgeStatusPending {
		t.Errorf("expected PENDING_DEPOSITORY, got %s", req.Status)
	}
	// Collateral = 100 * 2450 * (1 - 0.25) = 183750
	expectedCollateral := 100.0 * 2450.0 * 0.75
	if req.CollateralINR != expectedCollateral {
		t.Errorf("expected collateral %.2f, got %.2f", expectedCollateral, req.CollateralINR)
	}
	if req.Depository != DepositoryNSDL {
		t.Errorf("expected NSDL depository, got %s", req.Depository)
	}
}

func TestInitiatePledge_InvalidISIN(t *testing.T) {
	gw := NewPledgeGateway()

	_, err := gw.InitiatePledge("PLG-002", "CLIENT-A", "INVALID", "TEST", 10, 100.0, 10.0, DepositoryNSDL)
	if err == nil {
		t.Fatal("expected error for invalid ISIN")
	}
}

func TestInitiatePledge_ZeroQuantity(t *testing.T) {
	gw := NewPledgeGateway()

	_, err := gw.InitiatePledge("PLG-003", "CLIENT-A", "INE002A01018", "RELIANCE", 0, 100.0, 10.0, DepositoryNSDL)
	if err == nil {
		t.Fatal("expected error for zero quantity")
	}
}

func TestInitiatePledge_DuplicateID(t *testing.T) {
	gw := NewPledgeGateway()

	_, err := gw.InitiatePledge("PLG-004", "CLIENT-A", "INE002A01018", "RELIANCE", 50, 1000.0, 20.0, DepositoryCDSL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = gw.InitiatePledge("PLG-004", "CLIENT-B", "INE002A01018", "RELIANCE", 50, 1000.0, 20.0, DepositoryCDSL)
	if err == nil {
		t.Fatal("expected error for duplicate pledge ID")
	}
}

func TestConfirmPledge_Success(t *testing.T) {
	gw := NewPledgeGateway()

	_, _ = gw.InitiatePledge("PLG-005", "CLIENT-C", "INE009A01021", "INFY", 200, 1500.0, 20.0, DepositoryNSDL)

	confirmed, err := gw.ConfirmPledge("PLG-005", "NSDL-REF-12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if confirmed.Status != PledgeStatusConfirmed {
		t.Errorf("expected CONFIRMED, got %s", confirmed.Status)
	}
	if confirmed.DepositoryRef != "NSDL-REF-12345" {
		t.Errorf("expected depository ref NSDL-REF-12345, got %s", confirmed.DepositoryRef)
	}

	// Check collateral was credited
	collateral := gw.GetClientCollateral("CLIENT-C")
	expected := 200.0 * 1500.0 * 0.80
	if collateral != expected {
		t.Errorf("expected collateral %.2f, got %.2f", expected, collateral)
	}
}

func TestRejectPledge(t *testing.T) {
	gw := NewPledgeGateway()

	_, _ = gw.InitiatePledge("PLG-006", "CLIENT-D", "INE062A01020", "TCS", 50, 3500.0, 15.0, DepositoryCDSL)

	rejected, err := gw.RejectPledge("PLG-006", "Insufficient DP balance")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rejected.Status != PledgeStatusRejected {
		t.Errorf("expected REJECTED, got %s", rejected.Status)
	}
}

func TestUnpledge_FullLifecycle(t *testing.T) {
	gw := NewPledgeGateway()

	// Initiate -> Confirm -> Unpledge
	_, _ = gw.InitiatePledge("PLG-007", "CLIENT-E", "INE040A01034", "HDFCBANK", 100, 1600.0, 25.0, DepositoryNSDL)
	_, _ = gw.ConfirmPledge("PLG-007", "NSDL-REF-67890")

	collateralBefore := gw.GetClientCollateral("CLIENT-E")
	if collateralBefore <= 0 {
		t.Fatal("expected positive collateral after confirmation")
	}

	unpledged, err := gw.InitiateUnpledge("PLG-007")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if unpledged.Status != PledgeStatusUnpledged {
		t.Errorf("expected UNPLEDGED, got %s", unpledged.Status)
	}

	collateralAfter := gw.GetClientCollateral("CLIENT-E")
	if collateralAfter != 0 {
		t.Errorf("expected zero collateral after unpledge, got %.2f", collateralAfter)
	}
}

func TestUnpledge_PendingFails(t *testing.T) {
	gw := NewPledgeGateway()

	_, _ = gw.InitiatePledge("PLG-008", "CLIENT-F", "INE002A01018", "RELIANCE", 50, 2400.0, 20.0, DepositoryNSDL)

	// Trying to unpledge a PENDING pledge should fail
	_, err := gw.InitiateUnpledge("PLG-008")
	if err == nil {
		t.Fatal("expected error when unpledging a non-confirmed pledge")
	}
}

func TestGetClientPledges(t *testing.T) {
	gw := NewPledgeGateway()

	_, _ = gw.InitiatePledge("PLG-A1", "CLIENT-G", "INE002A01018", "RELIANCE", 10, 100.0, 10.0, DepositoryNSDL)
	_, _ = gw.InitiatePledge("PLG-A2", "CLIENT-G", "INE009A01021", "INFY", 20, 200.0, 15.0, DepositoryCDSL)

	pledges := gw.GetClientPledges("CLIENT-G")
	if len(pledges) != 2 {
		t.Errorf("expected 2 pledges for CLIENT-G, got %d", len(pledges))
	}
}

func TestConfirmPledge_NotFound(t *testing.T) {
	gw := NewPledgeGateway()

	_, err := gw.ConfirmPledge("NONEXISTENT", "REF-000")
	if err == nil {
		t.Fatal("expected error for nonexistent pledge")
	}
}

func TestInitiatePledge_InvalidDepository(t *testing.T) {
	gw := NewPledgeGateway()

	_, err := gw.InitiatePledge("PLG-009", "CLIENT-H", "INE002A01018", "RELIANCE", 10, 100.0, 10.0, "BSE")
	if err == nil {
		t.Fatal("expected error for invalid depository")
	}
}
