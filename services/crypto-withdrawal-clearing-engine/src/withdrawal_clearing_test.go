package main

import (
	"errors"
	"testing"
	"time"
)

func TestAddressFormatValidation(t *testing.T) {
	// Valid addresses
	if err := ValidateAddressFormat(NetworkEthereum, "0x71C8418320499023631777DD5002AEf089351a4F"); err != nil {
		t.Fatalf("expected valid ETH address, got %v", err)
	}
	if err := ValidateAddressFormat(NetworkBitcoin, "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq"); err != nil {
		t.Fatalf("expected valid BTC bech32 address, got %v", err)
	}
	if err := ValidateAddressFormat(NetworkBitcoin, "1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2"); err != nil {
		t.Fatalf("expected valid BTC legacy address, got %v", err)
	}
	if err := ValidateAddressFormat(NetworkTron, "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"); err != nil {
		t.Fatalf("expected valid Tron address, got %v", err)
	}
	if err := ValidateAddressFormat(NetworkSolana, "7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU"); err != nil {
		t.Fatalf("expected valid Solana address, got %v", err)
	}

	// Invalid addresses
	if err := ValidateAddressFormat(NetworkEthereum, "invalid_eth_address"); !errors.Is(err, ErrInvalidAddressFormat) {
		t.Fatalf("expected ErrInvalidAddressFormat for invalid ETH address, got %v", err)
	}
	if err := ValidateAddressFormat(NetworkBitcoin, "0x71C8418320499023631777DD5002AEf089351a4F"); !errors.Is(err, ErrInvalidAddressFormat) {
		t.Fatalf("expected ErrInvalidAddressFormat for ETH address on Bitcoin network, got %v", err)
	}
}

func TestBeneficiaryConfirmationAndCooloff(t *testing.T) {
	cooloff := 24 * time.Hour
	engine := NewWithdrawalClearingEngine(cooloff)
	now := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)

	ethAddr := "0x71C8418320499023631777DD5002AEf089351a4F"

	// 1. Add beneficiary with 2FA
	entry, err := engine.AddBeneficiary("USER-1", ethAddr, NetworkEthereum, "USDT", "My Ledger Cold Wallet", "", "123456")
	if err != nil {
		t.Fatalf("unexpected error adding beneficiary: %v", err)
	}
	if !entry.Is2FAConfirmed {
		t.Fatal("expected 2FA to be confirmed")
	}
	if entry.IsEmailConfirmed {
		t.Fatal("expected email not to be confirmed yet")
	}

	// 2. Withdrawal should fail before email confirmation
	_, err = engine.RequestWithdrawal("W-1", "USER-1", "USDT", NetworkEthereum, ethAddr, 1000_00000000, now)
	if !errors.Is(err, ErrBeneficiaryNotActive) {
		t.Fatalf("expected ErrBeneficiaryNotActive, got %v", err)
	}

	// 3. Confirm email with token
	err = engine.ConfirmBeneficiaryEmail("USER-1", ethAddr, entry.ConfirmationToken)
	if err != nil {
		t.Fatalf("unexpected error confirming email: %v", err)
	}

	// 4. Withdrawal should fail during 24h cool-off period
	_, err = engine.RequestWithdrawal("W-2", "USER-1", "USDT", NetworkEthereum, ethAddr, 1000_00000000, now.Add(12*time.Hour))
	if !errors.Is(err, ErrTimelockActive) {
		t.Fatalf("expected ErrTimelockActive during cool-off, got %v", err)
	}

	// 5. Withdrawal succeeds after 24h cool-off expires
	afterCooloff := now.Add(25 * time.Hour)
	req, err := engine.RequestWithdrawal("W-3", "USER-1", "USDT", NetworkEthereum, ethAddr, 1000_00000000, afterCooloff)
	if err != nil {
		t.Fatalf("unexpected error requesting withdrawal after cool-off: %v", err)
	}
	if req.Status != StatusPendingReview {
		t.Fatalf("expected StatusPendingReview, got %s", req.Status)
	}
}

func TestWhitelistOnlyMode(t *testing.T) {
	engine := NewWithdrawalClearingEngine(24 * time.Hour)
	now := time.Now().UTC()

	unlistedAddr := "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045" // Vitalik's address

	// Without whitelist-only mode: allowed
	req, err := engine.RequestWithdrawal("W-OPEN", "USER-2", "ETH", NetworkEthereum, unlistedAddr, 500_00000000, now)
	if err != nil || req == nil {
		t.Fatalf("expected open withdrawal allowed when whitelist-only mode disabled, got err %v", err)
	}

	// Enable whitelist-only mode
	engine.SetWhitelistOnlyMode("USER-2", true)

	// Now unlisted address is strictly blocked
	_, err = engine.RequestWithdrawal("W-BLOCKED", "USER-2", "ETH", NetworkEthereum, unlistedAddr, 500_00000000, now)
	if !errors.Is(err, ErrAddressNotWhitelisted) {
		t.Fatalf("expected ErrAddressNotWhitelisted, got %v", err)
	}
}
