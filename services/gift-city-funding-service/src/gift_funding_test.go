package main

import (
	"strings"
	"testing"
)

func newTestFundingService() *GIFTCityFundingService {
	svc := NewGIFTCityFundingService()
	_ = svc.RegisterInvestorWallet("ABCDE1234F", "0xWallet123")
	_ = svc.RegisterInvestorWallet("XYZPQ9876K", "0xWallet456")
	return svc
}

func validWire(wireID string) WireIngress {
	return WireIngress{
		WireID:           wireID,
		SenderName:       "John Doe",
		SenderBankSWIFT:  "CITIUS33",
		ReceiverIFSCCode: "GIFT0001234",
		AmountUSD:        10_000.00,
		PurposeCode:      "S0002",
		InvestorPAN:      "ABCDE1234F",
		SenderCountry:    "US",
	}
}

func TestProcessWireHappyPath(t *testing.T) {
	svc := newTestFundingService()
	receipt, err := svc.ProcessWire(validWire("WIRE-001"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receipt.Status != FundingCredited {
		t.Errorf("expected CREDITED, got %s", receipt.Status)
	}
	if !receipt.Compliance.OverallPass {
		t.Error("compliance should pass")
	}
	if receipt.WalletAddress != "0xWallet123" {
		t.Errorf("wrong wallet: %s", receipt.WalletAddress)
	}
	if !strings.HasPrefix(receipt.ReceiptID, "GIFT-") {
		t.Errorf("receipt ID format wrong: %s", receipt.ReceiptID)
	}
}

func TestProcessWireSanctionedCountryRejected(t *testing.T) {
	svc := newTestFundingService()
	wire := validWire("WIRE-002")
	wire.SenderCountry = "KP" // North Korea - sanctioned

	receipt, err := svc.ProcessWire(wire)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receipt.Status != FundingRejected {
		t.Errorf("expected REJECTED, got %s", receipt.Status)
	}
	if !strings.Contains(receipt.Compliance.RejectReason, "SANCTIONED_COUNTRY") {
		t.Errorf("expected sanctions reason: %s", receipt.Compliance.RejectReason)
	}
}

func TestProcessWireLRSLimitExceeded(t *testing.T) {
	svc := newTestFundingService()

	// First wire: $200,000
	wire1 := validWire("WIRE-003")
	wire1.AmountUSD = 200_000.00
	_, _ = svc.ProcessWire(wire1)

	// Second wire: $60,000 -> exceeds LRS $250,000 limit
	wire2 := validWire("WIRE-004")
	wire2.AmountUSD = 60_000.00
	receipt, err := svc.ProcessWire(wire2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receipt.Status != FundingRejected {
		t.Errorf("expected REJECTED for LRS breach, got %s", receipt.Status)
	}
	if !strings.Contains(receipt.Compliance.RejectReason, "LRS_LIMIT_EXCEEDED") {
		t.Errorf("expected LRS reason: %s", receipt.Compliance.RejectReason)
	}
}

func TestProcessWireDuplicateRejected(t *testing.T) {
	svc := newTestFundingService()
	_, _ = svc.ProcessWire(validWire("WIRE-DUP"))
	_, err := svc.ProcessWire(validWire("WIRE-DUP"))
	if err == nil {
		t.Fatal("expected duplicate wire error")
	}
}

func TestProcessWireInvalidPurposeCode(t *testing.T) {
	svc := newTestFundingService()
	wire := validWire("WIRE-005")
	wire.PurposeCode = "INVALID"

	receipt, _ := svc.ProcessWire(wire)
	if receipt.Status != FundingRejected {
		t.Errorf("expected REJECTED for invalid purpose, got %s", receipt.Status)
	}
}

func TestProcessWireUnregisteredWallet(t *testing.T) {
	svc := newTestFundingService()
	wire := validWire("WIRE-006")
	wire.InvestorPAN = "NOTWALLET1"

	_, err := svc.ProcessWire(wire)
	if err == nil {
		t.Fatal("expected unregistered wallet error")
	}
}

func TestCumulativeFundingTracking(t *testing.T) {
	svc := newTestFundingService()

	wire1 := validWire("WIRE-007")
	wire1.AmountUSD = 50_000.00
	svc.ProcessWire(wire1)

	wire2 := validWire("WIRE-008")
	wire2.AmountUSD = 30_000.00
	svc.ProcessWire(wire2)

	cumulative := svc.GetCumulativeFunding("ABCDE1234F")
	if cumulative != 80_000.00 {
		t.Errorf("expected cumulative $80,000, got $%.2f", cumulative)
	}
}
