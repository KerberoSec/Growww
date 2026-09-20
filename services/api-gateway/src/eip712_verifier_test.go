package main

import (
	"testing"
	"time"
)

func TestEIP712_OrderVerificationAndReplayDefense(t *testing.T) {
	domain := EIP712Domain{
		Name:              "Growww Exchange",
		Version:           "1.0",
		ChainID:           1337,
		VerifyingContract: "0x1111222233334444555566667777888899990000",
	}

	mgr := NewEIP712Manager(domain)
	trader := "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"

	order0 := SpotOrderData{
		Trader:   trader,
		ISIN:     "INE002A01018",
		Side:     "BUY",
		Quantity: 100,
		Price:    2500,
		Nonce:    0,
		Deadline: time.Now().Add(10 * time.Minute).Unix(),
	}

	// Verify order digest generation
	digest := mgr.HashSpotOrder(order0)
	if digest == [32]byte{} {
		t.Fatal("Expected non-empty EIP-712 digest")
	}

	// 1. Order with nonce 0 should succeed
	err := mgr.VerifyAndConsumeOrder(order0, trader)
	if err != nil {
		t.Fatalf("Expected order 0 to succeed: %v", err)
	}

	// 2. Replay of nonce 0 must be rejected
	err = mgr.VerifyAndConsumeOrder(order0, trader)
	if err == nil {
		t.Fatal("Expected replay order to be rejected with nonce error")
	}

	// 3. Order with valid sequential nonce 1 should succeed
	order1 := order0
	order1.Nonce = 1
	err = mgr.VerifyAndConsumeOrder(order1, trader)
	if err != nil {
		t.Fatalf("Expected order 1 to succeed: %v", err)
	}

	// 4. Expired order must fail
	expiredOrder := order0
	expiredOrder.Nonce = 2
	expiredOrder.Deadline = time.Now().Add(-1 * time.Minute).Unix()
	err = mgr.VerifyAndConsumeOrder(expiredOrder, trader)
	if err == nil {
		t.Fatal("Expected expired order to fail")
	}
}

func TestEIP712_SessionKeyDelegation(t *testing.T) {
	domain := EIP712Domain{
		Name:              "Growww Exchange",
		Version:           "1.0",
		ChainID:           1337,
		VerifyingContract: "0x1111222233334444555566667777888899990000",
	}

	mgr := NewEIP712Manager(domain)
	trader := "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"
	sessionKey := "0xSessionKeyAddress1234567890123456789012"

	// Register session key with notional limit 500,000
	delegation := SessionKeyDelegationData{
		MasterTrader:     trader,
		SessionKey:       sessionKey,
		ValidUntil:       time.Now().Add(1 * time.Hour).Unix(),
		MaxNotionalLimit: 500000,
		Nonce:            1,
	}

	err := mgr.RegisterSessionKey(delegation)
	if err != nil {
		t.Fatalf("Failed to register session key: %v", err)
	}

	// Order signed by delegated session key within limit
	order := SpotOrderData{
		Trader:   trader,
		ISIN:     "INE002A01018",
		Side:     "BUY",
		Quantity: 10,
		Price:    2500, // notional = 25,000 <= 500,000
		Nonce:    0,
		Deadline: time.Now().Add(10 * time.Minute).Unix(),
	}

	err = mgr.VerifyAndConsumeOrder(order, sessionKey)
	if err != nil {
		t.Fatalf("Expected session key order to succeed: %v", err)
	}

	// Order exceeding notional limit
	largeOrder := SpotOrderData{
		Trader:   trader,
		ISIN:     "INE002A01018",
		Side:     "BUY",
		Quantity: 500,
		Price:    2500, // notional = 1,250,000 > 500,000 limit
		Nonce:    1,
		Deadline: time.Now().Add(10 * time.Minute).Unix(),
	}

	err = mgr.VerifyAndConsumeOrder(largeOrder, sessionKey)
	if err == nil {
		t.Fatal("Expected order exceeding session notional limit to fail")
	}
}
