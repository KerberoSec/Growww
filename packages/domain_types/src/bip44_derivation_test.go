package src

import (
	"crypto/rand"
	"errors"
	"io"
	"testing"
)

func TestBIP44PathParsingAndValidation(t *testing.T) {
	// Standard Ethereum EVM path
	ethPathStr := "m/44'/60'/0'/0/0"
	p, err := ParseBIP44Path(ethPathStr)
	if err != nil {
		t.Fatalf("failed parsing eth path: %v", err)
	}

	if p.Purpose != 44 || p.CoinType != 60 || p.Account != 0 || p.Change != 0 || p.AddressIndex != 0 {
		t.Fatalf("path fields mismatch: %+v", p)
	}

	if p.Format() != ethPathStr {
		t.Fatalf("format mismatch: expected %s, got %s", ethPathStr, p.Format())
	}

	// Solana path with 'h' hardening notation
	solPathStr := "m/44h/501h/0h/0/5"
	solP, err := ParseBIP44Path(solPathStr)
	if err != nil {
		t.Fatalf("failed parsing sol path: %v", err)
	}
	if solP.CoinType != 501 || solP.AddressIndex != 5 {
		t.Fatalf("solana path fields mismatch: %+v", solP)
	}

	// Invalid path missing levels
	_, err = ParseBIP44Path("m/44'/60'")
	if err == nil || !errors.Is(err, ErrInvalidBIP44Path) {
		t.Fatalf("expected ErrInvalidBIP44Path for short path")
	}

	// Invalid purpose (e.g. 49' instead of 44')
	_, err = ParseBIP44Path("m/49'/0'/0'/0/0")
	if err == nil || !errors.Is(err, ErrInvalidBIP44Path) {
		t.Fatalf("expected ErrInvalidBIP44Path for purpose 49'")
	}
}

func TestBIP32KeyDerivation(t *testing.T) {
	seed := make([]byte, 64)
	if _, err := io.ReadFull(rand.Reader, seed); err != nil {
		t.Fatalf("failed generating seed: %v", err)
	}

	master, err := DeriveMasterKey(seed)
	if err != nil {
		t.Fatalf("failed deriving master key: %v", err)
	}

	if master.Depth != 0 {
		t.Fatalf("master depth must be 0")
	}

	// Derive hardened child (e.g. 44')
	child44 := master.DeriveChild(44 | HardenedBit)
	if child44.Depth != 1 {
		t.Fatalf("child depth must be 1, got %d", child44.Depth)
	}
	if child44.Key == master.Key {
		t.Fatalf("child key must differ from parent key")
	}

	// Derive second child (60')
	child60 := child44.DeriveChild(60 | HardenedBit)
	if child60.Depth != 2 {
		t.Fatalf("child depth must be 2, got %d", child60.Depth)
	}
}
