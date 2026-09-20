package main

import (
	"crypto/sha256"
	"testing"
)

func TestHSMOrchestrator_LifecycleAndSigning(t *testing.T) {
	custodians := []string{"fob_c1", "fob_c2", "fob_c3", "fob_c4", "fob_c5"}
	orch := NewHSMOrchestrator(BackendYubiHSM2, custodians)

	// 1. Load Root Key
	keyRef, err := orch.LoadRootKey(101, "besu_validator_root", "secp256k1")
	if err != nil {
		t.Fatalf("failed to load root key: %v", err)
	}
	if keyRef.KeyHandle != 101 {
		t.Fatalf("expected handle 101, got %d", keyRef.KeyHandle)
	}

	// 2. Sign Digest
	digest := sha256.Sum256([]byte("qbft_block_proposal_header"))
	sig, err := orch.SignDigest(101, digest)
	if err != nil {
		t.Fatalf("failed to sign digest: %v", err)
	}
	if len(sig) == 0 {
		t.Fatalf("expected non-empty signature")
	}

	// 3. Custodian Key Ceremony (3-of-5)
	// Fail with 2 fobs
	err = orch.ExecuteCustodianKeyCeremony([]string{"fob_c1", "fob_c2"}, "UPDATE_VALIDATOR_SET")
	if err == nil {
		t.Fatalf("expected error with only 2 fobs")
	}

	// Succeed with 3 fobs
	err = orch.ExecuteCustodianKeyCeremony([]string{"fob_c1", "fob_c3", "fob_c5"}, "UPDATE_VALIDATOR_SET")
	if err != nil {
		t.Fatalf("expected success with 3 fobs: %v", err)
	}

	// 4. Emergency Zeroization
	err = orch.EmergencyZeroizeKey(101, []string{"fob_c2", "fob_c3", "fob_c4"})
	if err != nil {
		t.Fatalf("failed to emergency zeroize key: %v", err)
	}

	// Attempt signing after zeroization -> must fail
	_, err = orch.SignDigest(101, digest)
	if err != ErrKeyAlreadyZeroized {
		t.Fatalf("expected ErrKeyAlreadyZeroized, got: %v", err)
	}
}
