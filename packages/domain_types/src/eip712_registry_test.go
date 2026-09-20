package src

import (
	"encoding/hex"
	"math/big"
	"testing"
)

func TestEIP712DomainSeparatorComputationAndRegistry(t *testing.T) {
	registry := NewEIP712DomainSeparatorRegistry()

	domain := EIP712Domain{
		Name:              "GrowwwSettlement",
		Version:           "1.0.0",
		ChainID:           big.NewInt(1337),
		VerifyingContract: "0x1111111111111111111111111111111111111111",
	}

	sep, err := registry.RegisterDomain(domain)
	if err != nil {
		t.Fatalf("failed to register domain: %v", err)
	}

	if sep == [32]byte{} {
		t.Fatalf("computed empty separator")
	}

	// Lookup from cache
	cached, found := registry.GetSeparator(big.NewInt(1337), "0x1111111111111111111111111111111111111111")
	if !found || cached != sep {
		t.Fatalf("cache lookup failed or mismatch: found=%v, cached=%x, expected=%x", found, cached, sep)
	}

	// Different chain ID should yield different separator
	domainChain2 := domain
	domainChain2.ChainID = big.NewInt(1338)
	sep2, err := registry.ComputeSeparator(domainChain2)
	if err != nil {
		t.Fatalf("failed computing separator for chain 2: %v", err)
	}
	if sep == sep2 {
		t.Fatalf("different chain IDs must result in different domain separators to prevent replay attacks")
	}

	// HashTypedDataV4 check
	var dummyStructHash [32]byte
	copy(dummyStructHash[:], []byte("order-hash-00000000000000000000"))

	digest := registry.HashTypedDataV4(sep, dummyStructHash)
	if digest == [32]byte{} {
		t.Fatalf("computed empty digest for typed data v4")
	}

	t.Logf("Computed Domain Separator: 0x%s", hex.EncodeToString(sep[:]))
	t.Logf("Computed Typed Data V4 Digest: 0x%s", hex.EncodeToString(digest[:]))
}
