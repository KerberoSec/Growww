package main

import (
	"crypto/sha256"
	"testing"
)

func TestMPCCMP_TwoOfThreeCeremony(t *testing.T) {
	coord, err := NewMPCCMPCoordinator(2, 3) // 2-of-3 threshold
	if err != nil {
		t.Fatalf("failed to init coordinator: %v", err)
	}

	digest := sha256.Sum256([]byte("bitcoin_withdrawal_tx_digest_9812"))
	sessionID := "SESS_WITHDRAWAL_2OF3_001"

	// 1. Start session with party 1 and 2
	err = coord.StartSigningSession(sessionID, digest, []uint32{1, 2})
	if err != nil {
		t.Fatalf("failed to start signing session: %v", err)
	}

	// 2. Submit party 1 share -> threshold not yet met
	quorumMet, err := coord.SubmitPartialSignature(sessionID, 1)
	if err != nil {
		t.Fatalf("unexpected error submitting party 1: %v", err)
	}
	if quorumMet {
		t.Fatalf("expected threshold not met with only 1 share")
	}

	// Attempt to retrieve signature before threshold
	_, err = coord.GetAggregatedSignature(sessionID)
	if err != ErrThresholdNotMet && err == nil {
		t.Fatalf("expected ErrThresholdNotMet, got %v", err)
	}

	// 3. Submit party 2 share -> threshold met!
	quorumMet, err = coord.SubmitPartialSignature(sessionID, 2)
	if err != nil {
		t.Fatalf("unexpected error submitting party 2: %v", err)
	}
	if !quorumMet {
		t.Fatalf("expected threshold met with 2 shares")
	}

	// 4. Retrieve aggregated signature
	sig, err := coord.GetAggregatedSignature(sessionID)
	if err != nil {
		t.Fatalf("failed to get aggregated signature: %v", err)
	}
	if len(sig) == 0 {
		t.Fatalf("expected non-empty aggregated signature bytes")
	}
}

func TestMPCCMP_InsufficientPartiesRejection(t *testing.T) {
	coord, err := NewMPCCMPCoordinator(3, 5) // 3-of-5 threshold
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	digest := sha256.Sum256([]byte("digest_test"))
	// Attempt with only 2 parties when 3 are required
	err = coord.StartSigningSession("SESS_FAIL", digest, []uint32{1, 2})
	if err == nil {
		t.Fatalf("expected error on insufficient participating parties")
	}
}

func TestMPCCMP_DuplicatePartySubmissionRejection(t *testing.T) {
	coord, _ := NewMPCCMPCoordinator(2, 3)
	digest := sha256.Sum256([]byte("digest_dup"))
	coord.StartSigningSession("SESS_DUP", digest, []uint32{1, 2})

	coord.SubmitPartialSignature("SESS_DUP", 1)

	// Attempt duplicate submission by party 1
	_, err := coord.SubmitPartialSignature("SESS_DUP", 1)
	if err != ErrPartyAlreadySigned {
		t.Fatalf("expected ErrPartyAlreadySigned, got: %v", err)
	}
}
