package src

import (
	"crypto/sha256"
	"testing"
	"time"
)

func TestPKCS11_SessionPoolAndHardwareKeyGeneration(t *testing.T) {
	pool := NewPKCS11SessionPool(1, 4)

	keyDesc, err := pool.GenerateHardwareKey("validator-besu-01", AlgorithmSecp256k1)
	if err != nil {
		t.Fatalf("failed to generate hardware key: %v", err)
	}

	if keyDesc.Extractable != false {
		t.Fatalf("expected non-extractable key flag")
	}
	if keyDesc.Sensitive != true {
		t.Fatalf("expected sensitive key flag")
	}
	if len(keyDesc.PublicKeyBytes) == 0 {
		t.Fatalf("expected valid public key bytes")
	}

	// Test signing
	digest := sha256.Sum256([]byte("test_consensus_message"))
	r, s, err := pool.SignDigest("validator-besu-01", digest[:])
	if err != nil {
		t.Fatalf("failed to sign digest: %v", err)
	}
	if r == nil || s == nil {
		t.Fatalf("expected valid (r,s) signature components")
	}
}

func TestQBFT_AntiEquivocationEnforcement(t *testing.T) {
	pool := NewPKCS11SessionPool(1, 4)
	_, err := pool.GenerateHardwareKey("val-primary", AlgorithmSecp256k1)
	if err != nil {
		t.Fatalf("failed to gen key: %v", err)
	}

	signer := NewQBFTBlockSigner(pool, "val-primary")

	hashA := sha256.Sum256([]byte("block_hash_A"))
	hashB := sha256.Sum256([]byte("block_hash_B"))

	req1 := BlockSigningRequest{
		ChainID:          1337,
		ValidatorAddress: "0x1111111111111111111111111111111111111111",
		BlockNumber:      100,
		Round:            0,
		BlockHash:        hashA,
		PayloadDigest:    ComputeCanonicalBlockDigest(1337, 100, 0, hashA),
	}

	resp1, err := signer.SignQBFTBlock(req1)
	if err != nil {
		t.Fatalf("expected successful initial block signing: %v", err)
	}
	if resp1.BlockNumber != 100 {
		t.Fatalf("expected block 100, got %d", resp1.BlockNumber)
	}

	// Equivocation attempt: Same chain, validator, block 100, round 0, but DIFFERENT hash (hashB)
	reqEquivocate := BlockSigningRequest{
		ChainID:          1337,
		ValidatorAddress: "0x1111111111111111111111111111111111111111",
		BlockNumber:      100,
		Round:            0,
		BlockHash:        hashB, // Different hash -> Slashing risk!
		PayloadDigest:    ComputeCanonicalBlockDigest(1337, 100, 0, hashB),
	}

	_, err = signer.SignQBFTBlock(reqEquivocate)
	if err == nil {
		t.Fatalf("expected anti-equivocation error on conflicting block hash, but succeeded")
	}

	// Block regression attempt: Block 99 after block 100
	reqRegress := BlockSigningRequest{
		ChainID:          1337,
		ValidatorAddress: "0x1111111111111111111111111111111111111111",
		BlockNumber:      99,
		Round:            0,
		BlockHash:        hashA,
		PayloadDigest:    ComputeCanonicalBlockDigest(1337, 99, 0, hashA),
	}

	_, err = signer.SignQBFTBlock(reqRegress)
	if err == nil {
		t.Fatalf("expected error on block number regression")
	}

	// Valid advance: Block 101
	hashNext := sha256.Sum256([]byte("block_hash_101"))
	reqNext := BlockSigningRequest{
		ChainID:          1337,
		ValidatorAddress: "0x1111111111111111111111111111111111111111",
		BlockNumber:      101,
		Round:            0,
		BlockHash:        hashNext,
		PayloadDigest:    ComputeCanonicalBlockDigest(1337, 101, 0, hashNext),
	}

	respNext, err := signer.SignQBFTBlock(reqNext)
	if err != nil {
		t.Fatalf("expected valid advance to block 101, got: %v", err)
	}
	if respNext.BlockNumber != 101 {
		t.Fatalf("expected block 101, got %d", respNext.BlockNumber)
	}
}

func TestEIP712_SettlementBatchSigning(t *testing.T) {
	pool := NewPKCS11SessionPool(1, 4)
	pool.GenerateHardwareKey("settlement-relayer", AlgorithmSecp256k1)

	whitelistedContract := "0x5FbDB2315678afecb367f032d93F642f64180aa3"
	eipSigner := NewEIP712BatchSigner(
		pool,
		"settlement-relayer",
		[]string{whitelistedContract},
		10000000.0, // ₹1 Crore max batch
	)

	domain := EIP712Domain{
		Name:              "NBSESettlementDvP",
		Version:           "1",
		ChainID:           1337,
		VerifyingContract: whitelistedContract,
	}

	now := time.Now().Unix()

	batch := SettlementBatchData{
		BatchID:         "BATCH-2026-001",
		MerkleRoot:      sha256.Sum256([]byte("trade_merkle_root")),
		TradeCount:      150,
		GrossValueINR:   5000000.0, // ₹50 Lakhs (within limit)
		ExpiryTimestamp: now + 300,
	}

	sigHex, digest, err := eipSigner.SignSettlementBatch(domain, batch, now)
	if err != nil {
		t.Fatalf("failed to sign settlement batch: %v", err)
	}

	if len(sigHex) == 0 {
		t.Fatalf("expected non-empty hex signature")
	}
	if len(digest) != 32 {
		t.Fatalf("expected 32 byte digest")
	}

	// Test unwhitelisted contract rejection
	badDomain := domain
	badDomain.VerifyingContract = "0x9999999999999999999999999999999999999999"
	_, _, err = eipSigner.SignSettlementBatch(badDomain, batch, now)
	if err == nil {
		t.Fatalf("expected rejection of unwhitelisted contract")
	}

	// Test collar exceed rejection (₹2 Crores > ₹1 Crore limit)
	overCollarBatch := batch
	overCollarBatch.GrossValueINR = 20000000.0
	_, _, err = eipSigner.SignSettlementBatch(domain, overCollarBatch, now)
	if err == nil {
		t.Fatalf("expected rejection of batch exceeding gross collar")
	}
}

func TestKeyCeremony_MofNQuorumAndShredding(t *testing.T) {
	pool := NewPKCS11SessionPool(1, 4)
	officers := []string{"officer_1", "officer_2", "officer_3", "officer_4", "officer_5"}
	coordinator := NewKeyCeremonyCoordinator(pool, 3, 5, officers) // 3 of 5 quorum

	// Attempt generation with only 2 officers (must fail)
	_, err := coordinator.ExecuteKeyGenerationCeremony(
		"master_treasury_key",
		AlgorithmSecp256k1,
		[]string{"officer_1", "officer_2"},
	)
	if err != ErrQuorumNotMet && err == nil {
		t.Fatalf("expected ErrQuorumNotMet when only 2 officers sign")
	}

	// Successful generation with 3 officers
	rec, err := coordinator.ExecuteKeyGenerationCeremony(
		"master_treasury_key_v1",
		AlgorithmSecp256k1,
		[]string{"officer_1", "officer_2", "officer_3"},
	)
	if err != nil {
		t.Fatalf("expected successful ceremony with 3 officers: %v", err)
	}
	if rec.State != KeyStateActive {
		t.Fatalf("expected active key state, got %s", rec.State)
	}

	// Generate replacement key
	_, err = coordinator.ExecuteKeyGenerationCeremony(
		"master_treasury_key_v2",
		AlgorithmSecp256k1,
		[]string{"officer_2", "officer_3", "officer_4"},
	)
	if err != nil {
		t.Fatalf("failed to generate v2 key: %v", err)
	}

	// Rotate v1 to v2
	err = coordinator.RotateKey(
		"master_treasury_key_v1",
		"master_treasury_key_v2",
		[]string{"officer_1", "officer_2", "officer_3"},
	)
	if err != nil {
		t.Fatalf("failed to rotate key: %v", err)
	}

	v1Rec, _ := coordinator.GetKeyRecord("master_treasury_key_v1")
	v2Rec, _ := coordinator.GetKeyRecord("master_treasury_key_v2")
	if v1Rec.State != KeyStateArchived {
		t.Fatalf("expected v1 to be ARCHIVED, got %s", v1Rec.State)
	}
	if v2Rec.State != KeyStateActive {
		t.Fatalf("expected v2 to be ACTIVE, got %s", v2Rec.State)
	}

	// Cryptographic shredding
	err = coordinator.ShredKey("master_treasury_key_v1", []string{"officer_1", "officer_3", "officer_5"})
	if err != nil {
		t.Fatalf("failed to shred key: %v", err)
	}

	v1Shredded, _ := coordinator.GetKeyRecord("master_treasury_key_v1")
	if v1Shredded.State != KeyStateRevoked {
		t.Fatalf("expected REVOKED state after shredding, got %s", v1Shredded.State)
	}

	// Ensure shredded key cannot be used in pool
	_, _, err = pool.SignDigest("master_treasury_key_v1", []byte("hash"))
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound for shredded key, got %v", err)
	}
}
