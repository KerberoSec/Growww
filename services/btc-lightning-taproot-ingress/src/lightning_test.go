package main

import "testing"

func TestCreateAndSettleInvoice(t *testing.T) {
	svc := NewLightningService()
	inv, err := svc.CreateInvoice("usr_1", 100000, "deposit", 3600)
	if err != nil {
		t.Fatal(err)
	}
	if inv.Status != "PENDING" {
		t.Errorf("expected PENDING, got %s", inv.Status)
	}
	if err := svc.SettleInvoice(inv.PaymentHash); err != nil {
		t.Fatal(err)
	}
	settled, _ := svc.GetInvoice(inv.PaymentHash)
	if settled.Status != "SETTLED" {
		t.Errorf("expected SETTLED, got %s", settled.Status)
	}
}

func TestDoubleSettle(t *testing.T) {
	svc := NewLightningService()
	inv, _ := svc.CreateInvoice("usr_2", 50000, "test", 3600)
	svc.SettleInvoice(inv.PaymentHash)
	if err := svc.SettleInvoice(inv.PaymentHash); err == nil {
		t.Error("expected error on double settle")
	}
}

func TestZeroAmount(t *testing.T) {
	svc := NewLightningService()
	_, err := svc.CreateInvoice("usr_3", 0, "zero", 3600)
	if err == nil {
		t.Error("expected error for zero amount")
	}
}

func TestTaprootAddress(t *testing.T) {
	addr := DeriveTaprootAddress("02abcdef1234567890")
	if len(addr) < 10 || addr[:4] != "bc1p" {
		t.Errorf("invalid taproot address: %s", addr)
	}
}
