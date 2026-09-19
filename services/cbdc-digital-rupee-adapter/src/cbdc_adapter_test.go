package main

import (
	"testing"
)

func TestCBDCBridgeAdapter_SettleTransfer(t *testing.T) {
	adapter := NewCBDCBridgeAdapter("HDFC0001234")

	receipt, err := adapter.SettleCBDCTransfer(CBDCTransferRequest{
		TransferID:  "TXF-001",
		PayerVPA:    "alice@rbi.edr",
		PayeeVPA:    "bob@rbi.edr",
		AmountPaise: 100000, // ₹1000
		Type:        CBDCRetail,
	})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if receipt.SettlementStatus != "SETTLED_INSTANT_RBI_CORE" {
		t.Fatalf("Expected instant settlement, got %s", receipt.SettlementStatus)
	}
	if receipt.RBICBDCTxID == "" {
		t.Fatal("Expected non-empty RBI CBDC TxID")
	}
}

func TestCBDCBridgeAdapter_RejectZeroAmount(t *testing.T) {
	adapter := NewCBDCBridgeAdapter("SBI0001234")

	_, err := adapter.SettleCBDCTransfer(CBDCTransferRequest{
		TransferID:  "TXF-002",
		PayerVPA:    "alice@rbi.edr",
		PayeeVPA:    "bob@rbi.edr",
		AmountPaise: 0,
		Type:        CBDCRetail,
	})
	if err == nil {
		t.Fatal("Expected error for zero amount")
	}
}

func TestCBDCTokenManager_IssueAndTransfer(t *testing.T) {
	adapter := NewCBDCBridgeAdapter("HDFC0001234")
	mgr := NewCBDCTokenManager(adapter)

	token, err := mgr.IssueToken(CBDCRetail, 500000, "alice@rbi.edr") // ₹5000
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if token.Status != TokenIssued {
		t.Fatalf("Expected ISSUED status, got %s", token.Status)
	}

	// Transfer to bob
	receipt, err := mgr.TransferToken(token.TokenID, "alice@rbi.edr", "bob@rbi.edr")
	if err != nil {
		t.Fatalf("Transfer failed: %v", err)
	}
	if receipt == nil {
		t.Fatal("Expected transfer receipt")
	}

	// Verify ownership changed
	updated, ok := mgr.GetToken(token.TokenID)
	if !ok {
		t.Fatal("Token not found after transfer")
	}
	if updated.OwnerVPA != "bob@rbi.edr" {
		t.Fatalf("Expected owner bob@rbi.edr, got %s", updated.OwnerVPA)
	}
}

func TestCBDCTokenManager_RedeemToken(t *testing.T) {
	adapter := NewCBDCBridgeAdapter("AXIS0001234")
	mgr := NewCBDCTokenManager(adapter)

	token, _ := mgr.IssueToken(CBDCRetail, 100000, "charlie@rbi.edr")

	err := mgr.RedeemToken(token.TokenID, "charlie@rbi.edr")
	if err != nil {
		t.Fatalf("Redeem failed: %v", err)
	}

	redeemed, _ := mgr.GetToken(token.TokenID)
	if redeemed.Status != TokenRedeemed {
		t.Fatalf("Expected REDEEMED status, got %s", redeemed.Status)
	}

	// Cannot redeem again
	err = mgr.RedeemToken(token.TokenID, "charlie@rbi.edr")
	if err == nil {
		t.Fatal("Expected error for double redemption")
	}
}

func TestCBDCTokenManager_TransferNonOwner(t *testing.T) {
	adapter := NewCBDCBridgeAdapter("ICICI001")
	mgr := NewCBDCTokenManager(adapter)

	token, _ := mgr.IssueToken(CBDCRetail, 200000, "alice@rbi.edr")

	// Try transfer from non-owner
	_, err := mgr.TransferToken(token.TokenID, "eve@rbi.edr", "bob@rbi.edr")
	if err == nil {
		t.Fatal("Expected error: sender does not own token")
	}
}
