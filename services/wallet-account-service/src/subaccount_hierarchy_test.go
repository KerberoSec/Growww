package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestSubAccountHierarchy_CeilingAndIsolation(t *testing.T) {
	mgr := NewSubAccountHierarchyManager()

	master, err := mgr.RegisterMasterAccount("MASTER-ACME", "Acme Capital LP", "U67120MH2024PTC123456")
	if err != nil {
		t.Fatalf("unexpected error registering master: %v", err)
	}
	if master.MasterID != "MASTER-ACME" {
		t.Fatalf("unexpected master ID: %s", master.MasterID)
	}

	// Create 100 sub-accounts
	for i := 1; i <= 100; i++ {
		subID := fmt.Sprintf("SUB-%03d", i)
		_, err := mgr.CreateSubAccount("MASTER-ACME", subID, fmt.Sprintf("Algo Strategy %d", i), RoleTradingOnlyBot)
		if err != nil {
			t.Fatalf("unexpected error creating sub-account %s: %v", subID, err)
		}
	}

	// 101st sub-account should be rejected
	_, err = mgr.CreateSubAccount("MASTER-ACME", "SUB-101", "Overflow Strategy", RoleTradingOnlyBot)
	if !errors.Is(err, ErrMaxSubAccountsExceeded) {
		t.Fatalf("expected ErrMaxSubAccountsExceeded for 101st account, got %v", err)
	}
}

func TestSubAccountHierarchy_RolePermissions(t *testing.T) {
	mgr := NewSubAccountHierarchyManager()
	_, _ = mgr.RegisterMasterAccount("MASTER-1", "Alpha Hedge Fund", "CIN123")

	_, _ = mgr.CreateSubAccount("MASTER-1", "SUB-READONLY", "Audit Desk", RoleReadOnlyAnalyst)
	_, _ = mgr.CreateSubAccount("MASTER-1", "SUB-TRADER", "HFT Execution Bot", RoleTradingOnlyBot)
	_, _ = mgr.CreateSubAccount("MASTER-1", "SUB-MANAGER", "Risk Manager Desk", RoleFullManager)

	// ReadOnly cannot trade
	err := mgr.ValidateOrderPlacement("SUB-READONLY")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied for ReadOnly trading, got %v", err)
	}

	// Trader can trade
	err = mgr.ValidateOrderPlacement("SUB-TRADER")
	if err != nil {
		t.Fatalf("expected Trader allowed to trade, got %v", err)
	}

	// Manager can trade
	err = mgr.ValidateOrderPlacement("SUB-MANAGER")
	if err != nil {
		t.Fatalf("expected Manager allowed to trade, got %v", err)
	}
}

func TestSubAccountHierarchy_ZeroFeeInternalTransfers(t *testing.T) {
	mgr := NewSubAccountHierarchyManager()
	_, _ = mgr.RegisterMasterAccount("CORP-1", "Beta Trading LLC", "CIN-BETA")

	_, _ = mgr.CreateSubAccount("CORP-1", "SUB-A", "Strategy A", RoleTradingOnlyBot)
	_, _ = mgr.CreateSubAccount("CORP-1", "SUB-B", "Strategy B", RoleFullManager)

	// Fund Master with 10 BTC
	_ = mgr.DepositToMaster("CORP-1", "BTC", 10_00000000)

	// Transfer 3 BTC from Master to SUB-A
	receipt, err := mgr.TransferInternal("TX-001", "", "CORP-1", "SUB-A", "BTC", 3_00000000)
	if err != nil {
		t.Fatalf("unexpected error during transfer: %v", err)
	}
	if receipt.FeeE8 != 0 {
		t.Fatalf("internal transfer fee must be 0, got %d", receipt.FeeE8)
	}

	// Transfer 1 BTC from SUB-A to SUB-B initiated by Manager
	receipt2, err := mgr.TransferInternal("TX-002", "SUB-B", "SUB-A", "SUB-B", "BTC", 1_00000000)
	if err != nil {
		t.Fatalf("unexpected error during sibling transfer: %v", err)
	}
	if receipt2.FeeE8 != 0 {
		t.Fatalf("sibling transfer fee must be 0, got %d", receipt2.FeeE8)
	}

	// Sibling transfer initiated by non-manager (SUB-A has RoleTradingOnlyBot) should be rejected
	_, err = mgr.TransferInternal("TX-003", "SUB-A", "SUB-B", "SUB-A", "BTC", 50000000)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied when non-manager initiates transfer, got %v", err)
	}

	// Check aggregate balance across CORP-1
	// Master (7 BTC) + SUB-A (2 BTC) + SUB-B (1 BTC) = 10 BTC total
	agg, err := mgr.GetAggregateBalance("CORP-1", "BTC")
	if err != nil || agg != 10_00000000 {
		t.Fatalf("expected aggregate balance 10 BTC, got %d (err: %v)", agg, err)
	}
}

func TestSubAccountHierarchy_CrossMasterRejection(t *testing.T) {
	mgr := NewSubAccountHierarchyManager()
	_, _ = mgr.RegisterMasterAccount("FIRM-1", "Firm 1", "CIN1")
	_, _ = mgr.RegisterMasterAccount("FIRM-2", "Firm 2", "CIN2")

	_, _ = mgr.CreateSubAccount("FIRM-1", "SUB-1A", "Strategy 1A", RoleFullManager)
	_, _ = mgr.CreateSubAccount("FIRM-2", "SUB-2A", "Strategy 2A", RoleFullManager)

	_ = mgr.DepositToMaster("FIRM-1", "USDT", 1000_00000000)

	// Attempt transfer across entities
	_, err := mgr.TransferInternal("TX-LEAK", "", "FIRM-1", "SUB-2A", "USDT", 500_00000000)
	if !errors.Is(err, ErrCrossMasterTransfer) {
		t.Fatalf("expected ErrCrossMasterTransfer for cross-firm transfer, got %v", err)
	}
}
