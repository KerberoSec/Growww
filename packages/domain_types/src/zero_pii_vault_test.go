package src

import (
	"crypto/rand"
	"io"
	"strings"
	"testing"
	"time"
)

func TestZeroPIIVaultTokenizationAndBlindIndex(t *testing.T) {
	saltMgr, err := NewSaltEntropyManager(24 * time.Hour)
	if err != nil {
		t.Fatalf("failed creating salt manager: %v", err)
	}

	var masterKey [32]byte
	if _, err := io.ReadFull(rand.Reader, masterKey[:]); err != nil {
		t.Fatalf("failed generating master key: %v", err)
	}

	vault, err := NewZeroPIIVault(masterKey, saltMgr)
	if err != nil {
		t.Fatalf("failed initializing vault: %v", err)
	}

	pan := "ABCDE1234F"
	token1, blindIndex1, err := vault.Tokenize(TokenTypePAN, pan)
	if err != nil {
		t.Fatalf("tokenization failed: %v", err)
	}

	// Blind index must start with v1:
	if !strings.HasPrefix(blindIndex1, "v1:") {
		t.Fatalf("expected blind index prefix v1:, got %s", blindIndex1)
	}

	// Deterministic blind index check: same PAN generates same blind index with same salt version
	token2, blindIndex2, err := vault.Tokenize(TokenTypePAN, pan)
	if err != nil {
		t.Fatalf("second tokenization failed: %v", err)
	}
	if blindIndex1 != blindIndex2 {
		t.Fatalf("blind index must be deterministic for identical input")
	}
	if token1 == token2 {
		t.Fatalf("tokens must have randomized nonces for identical input")
	}

	// Detokenize check
	decrypted1, err := vault.Detokenize(TokenTypePAN, token1)
	if err != nil || decrypted1 != pan {
		t.Fatalf("detokenize failed: expected %s, got %s, err=%v", pan, decrypted1, err)
	}
	decrypted2, err := vault.Detokenize(TokenTypePAN, token2)
	if err != nil || decrypted2 != pan {
		t.Fatalf("detokenize 2 failed: expected %s, got %s, err=%v", pan, decrypted2, err)
	}

	// Mismatched token type should fail AEAD authentication
	_, err = vault.Detokenize(TokenTypeAadhaar, token1)
	if err == nil {
		t.Fatalf("expected AEAD tag verification failure on token type mismatch")
	}

	// Rotate salt and verify version increment
	newVer, err := saltMgr.RotateSalt()
	if err != nil || newVer != 2 {
		t.Fatalf("salt rotation failed: version=%d, err=%v", newVer, err)
	}

	_, blindIndexV2, _ := vault.Tokenize(TokenTypePAN, pan)
	if !strings.HasPrefix(blindIndexV2, "v2:") {
		t.Fatalf("expected blind index prefix v2: after rotation, got %s", blindIndexV2)
	}
}
