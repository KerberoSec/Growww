package main

import (
	"strings"
	"testing"
)

func newTestRouter() *CollateralRouter {
	reg := NewBridgeRegistry()
	return NewCollateralRouter(reg, OptimizeBlend)
}

func TestRouteHappyPath(t *testing.T) {
	router := newTestRouter()
	req := CollateralTransferRequest{
		RequestID:   "REQ-001",
		UserID:      "USR-42",
		Asset:       CollateralAsset{Symbol: "USDC", ContractAddress: "0xA0b8...", ChainID: ChainEthereum},
		AmountE8:    10_000_00000000,
		SourceChain: ChainEthereum,
		DestChain:   ChainPolygon,
		MaxFeeUSD:   5.00,
		Urgency:     "MEDIUM",
	}

	receipt, err := router.Route(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receipt.RequestID != "REQ-001" {
		t.Errorf("expected RequestID REQ-001, got %s", receipt.RequestID)
	}
	if receipt.Status != RouteStatusInFlight {
		t.Errorf("expected status IN_FLIGHT, got %s", receipt.Status)
	}
	if receipt.FeeChargedUSD > 5.00 {
		t.Errorf("fee %.2f exceeded max 5.00", receipt.FeeChargedUSD)
	}
	if !strings.HasPrefix(receipt.RouteHash, "RH-") {
		t.Errorf("route hash missing prefix: %s", receipt.RouteHash)
	}
}

func TestRouteRejectsZeroAmount(t *testing.T) {
	router := newTestRouter()
	req := CollateralTransferRequest{
		RequestID:   "REQ-002",
		UserID:      "USR-42",
		AmountE8:    0,
		SourceChain: ChainEthereum,
		DestChain:   ChainPolygon,
	}

	_, err := router.Route(req)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
	if !strings.Contains(err.Error(), "amount must be > 0") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRouteSameChainRejected(t *testing.T) {
	router := newTestRouter()
	req := CollateralTransferRequest{
		RequestID:   "REQ-003",
		UserID:      "USR-42",
		AmountE8:    1_00000000,
		SourceChain: ChainEthereum,
		DestChain:   ChainEthereum,
	}

	_, err := router.Route(req)
	if err == nil {
		t.Fatal("expected error for same source/dest chain")
	}
}

func TestRouteMaxFeeFilter(t *testing.T) {
	router := newTestRouter()
	req := CollateralTransferRequest{
		RequestID:   "REQ-004",
		UserID:      "USR-42",
		AmountE8:    1_00000000,
		SourceChain: ChainEthereum,
		DestChain:   ChainPolygon,
		MaxFeeUSD:   0.01, // impossibly low
		Urgency:     "LOW",
	}

	_, err := router.Route(req)
	if err == nil {
		t.Fatal("expected route failure for impossibly low fee ceiling")
	}
}

func TestSelectOptimalRouteCost(t *testing.T) {
	candidates := []BridgeRoute{
		{BridgeCCIP, ChainEthereum, ChainPolygon, 3.00, 180_000, 100_000_00000000, true},
		{BridgeAxelar, ChainEthereum, ChainPolygon, 1.50, 300_000, 100_000_00000000, true},
	}

	best, err := SelectOptimalRoute(candidates, OptimizeCost, 1_00000000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if best.Provider != BridgeAxelar {
		t.Errorf("expected AXELAR (cheapest), got %s", best.Provider)
	}
}

func TestSelectOptimalRouteSpeed(t *testing.T) {
	candidates := []BridgeRoute{
		{BridgeCCIP, ChainEthereum, ChainPolygon, 3.00, 60_000, 100_000_00000000, true},
		{BridgeAxelar, ChainEthereum, ChainPolygon, 1.50, 300_000, 100_000_00000000, true},
	}

	best, err := SelectOptimalRoute(candidates, OptimizeSpeed, 1_00000000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if best.Provider != BridgeCCIP {
		t.Errorf("expected CCIP (fastest), got %s", best.Provider)
	}
}

func TestConfirmAndFailTransfer(t *testing.T) {
	router := newTestRouter()
	req := CollateralTransferRequest{
		RequestID:   "REQ-005",
		UserID:      "USR-42",
		AmountE8:    5_00000000,
		SourceChain: ChainEthereum,
		DestChain:   ChainArbitrum,
		Urgency:     "HIGH",
	}

	_, err := router.Route(req)
	if err != nil {
		t.Fatalf("route failed: %v", err)
	}

	err = router.ConfirmTransfer("REQ-005")
	if err != nil {
		t.Fatalf("confirm failed: %v", err)
	}

	receipt, _ := router.GetReceipt("REQ-005")
	if receipt.Status != RouteStatusConfirmed {
		t.Errorf("expected CONFIRMED, got %s", receipt.Status)
	}

	// Cannot confirm again
	err = router.ConfirmTransfer("REQ-005")
	if err == nil {
		t.Fatal("expected error confirming already-confirmed transfer")
	}
}
