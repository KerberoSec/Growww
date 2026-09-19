package main

import (
	"testing"
)

func TestGenerateDepositAddress_Success(t *testing.T) {
	ws := NewWalletService()

	addr, err := ws.GenerateDepositAddress("USER-1", AssetBTC, "BITCOIN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if addr.Address == "" {
		t.Error("expected non-empty address")
	}
	if addr.UserID != "USER-1" {
		t.Errorf("expected UserID USER-1, got %s", addr.UserID)
	}
	if addr.Network != "BITCOIN" {
		t.Errorf("expected network BITCOIN, got %s", addr.Network)
	}
}

func TestGenerateDepositAddress_Idempotent(t *testing.T) {
	ws := NewWalletService()

	addr1, _ := ws.GenerateDepositAddress("USER-1", AssetETH, "ETHEREUM")
	addr2, _ := ws.GenerateDepositAddress("USER-1", AssetETH, "ETHEREUM")

	if addr1.Address != addr2.Address {
		t.Error("expected idempotent address generation to return same address")
	}
}

func TestGenerateDepositAddress_EmptyUserID(t *testing.T) {
	ws := NewWalletService()

	_, err := ws.GenerateDepositAddress("", AssetBTC, "BITCOIN")
	if err == nil {
		t.Fatal("expected error for empty user ID")
	}
}

func TestCreditDeposit_AndGetBalance(t *testing.T) {
	ws := NewWalletService()

	err := ws.CreditDeposit("USER-2", AssetUSDT, 100_00000000) // 100 USDT
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bal := ws.GetBalance("USER-2", AssetUSDT)
	if bal.Available != 100_00000000 {
		t.Errorf("expected available 10000000000, got %d", bal.Available)
	}
	if bal.Total != 100_00000000 {
		t.Errorf("expected total 10000000000, got %d", bal.Total)
	}
}

func TestCreditDeposit_ZeroAmount(t *testing.T) {
	ws := NewWalletService()

	err := ws.CreditDeposit("USER-2", AssetBTC, 0)
	if err == nil {
		t.Fatal("expected error for zero deposit amount")
	}
}

func TestLockFunds_Success(t *testing.T) {
	ws := NewWalletService()

	_ = ws.CreditDeposit("USER-3", AssetBTC, 5_00000000)
	err := ws.LockFunds("USER-3", AssetBTC, 2_00000000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bal := ws.GetBalance("USER-3", AssetBTC)
	if bal.Available != 3_00000000 {
		t.Errorf("expected available 300000000, got %d", bal.Available)
	}
	if bal.Locked != 2_00000000 {
		t.Errorf("expected locked 200000000, got %d", bal.Locked)
	}
}

func TestLockFunds_InsufficientBalance(t *testing.T) {
	ws := NewWalletService()

	_ = ws.CreditDeposit("USER-4", AssetETH, 1_00000000)
	err := ws.LockFunds("USER-4", AssetETH, 5_00000000)
	if err == nil {
		t.Fatal("expected error for insufficient balance")
	}
}

func TestUnlockFunds_Success(t *testing.T) {
	ws := NewWalletService()

	_ = ws.CreditDeposit("USER-5", AssetUSDT, 10_00000000)
	_ = ws.LockFunds("USER-5", AssetUSDT, 4_00000000)
	err := ws.UnlockFunds("USER-5", AssetUSDT, 4_00000000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bal := ws.GetBalance("USER-5", AssetUSDT)
	if bal.Available != 10_00000000 {
		t.Errorf("expected available 1000000000, got %d", bal.Available)
	}
	if bal.Locked != 0 {
		t.Errorf("expected locked 0, got %d", bal.Locked)
	}
}

func TestWithdrawal_FullLifecycle(t *testing.T) {
	ws := NewWalletService()

	_ = ws.CreditDeposit("USER-6", AssetBTC, 10_00000000)

	wr, err := ws.InitiateWithdrawal("WD-001", "USER-6", AssetBTC, 1_00000000, 10000, "bc1qxyz...", "BITCOIN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wr.Status != WithdrawalPending {
		t.Errorf("expected PENDING status, got %s", wr.Status)
	}

	bal := ws.GetBalance("USER-6", AssetBTC)
	if bal.Available != 8_99990000 {
		t.Errorf("expected available 899990000, got %d", bal.Available)
	}

	completed, err := ws.CompleteWithdrawal("WD-001", "txhash_abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if completed.Status != WithdrawalCompleted {
		t.Errorf("expected COMPLETED status, got %s", completed.Status)
	}

	bal = ws.GetBalance("USER-6", AssetBTC)
	if bal.Total != 8_99990000 {
		t.Errorf("expected total 899990000, got %d", bal.Total)
	}
}

func TestWithdrawal_InsufficientBalance(t *testing.T) {
	ws := NewWalletService()

	_ = ws.CreditDeposit("USER-7", AssetETH, 1_00000000)

	_, err := ws.InitiateWithdrawal("WD-002", "USER-7", AssetETH, 5_00000000, 10000, "0xabc...", "ETHEREUM")
	if err == nil {
		t.Fatal("expected error for insufficient balance")
	}
}

func TestWithdrawal_DuplicateID(t *testing.T) {
	ws := NewWalletService()

	_ = ws.CreditDeposit("USER-8", AssetUSDT, 100_00000000)
	_, _ = ws.InitiateWithdrawal("WD-003", "USER-8", AssetUSDT, 10_00000000, 0, "dest", "TRC20")

	_, err := ws.InitiateWithdrawal("WD-003", "USER-8", AssetUSDT, 10_00000000, 0, "dest", "TRC20")
	if err == nil {
		t.Fatal("expected error for duplicate withdrawal ID")
	}
}

func TestGetAllBalances_MultiAsset(t *testing.T) {
	ws := NewWalletService()

	_ = ws.CreditDeposit("USER-9", AssetBTC, 1_00000000)
	_ = ws.CreditDeposit("USER-9", AssetETH, 10_00000000)
	_ = ws.CreditDeposit("USER-9", AssetUSDT, 1000_00000000)

	balances := ws.GetAllBalances("USER-9")
	if len(balances) != 3 {
		t.Errorf("expected 3 asset balances, got %d", len(balances))
	}
}
