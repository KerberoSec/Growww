package worm_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"growww/database/worm"
)

func TestWORMStorageEngine_ComplianceModeEnforcement(t *testing.T) {
	engine := worm.NewWORMStorageEngine()
	ctx := context.Background()

	payload := []byte("Trade Confirmation ID: 998823 - 50.0 BTC @ 65000 USDT")
	req := worm.PutObjectRequest{
		Bucket:          "growww-audit-worm",
		Key:             "trades/2026/09/20/trade_998823.json",
		Payload:         payload,
		RetentionMode:   worm.ModeCompliance,
		RetainUntilDate: time.Now().Add(7 * 365 * 24 * time.Hour), // 7-year retention
		LegalHold:       false,
	}

	resp, err := engine.PutObject(ctx, req)
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	// Attempt deletion in compliance mode while locked
	err = engine.DeleteObject(ctx, req.Bucket, req.Key, resp.VersionID, false, "")
	if !errors.Is(err, worm.ErrObjectLockedInCompliance) {
		t.Errorf("expected ErrObjectLockedInCompliance, got: %v", err)
	}

	// Attempt deletion with fake governance bypass in compliance mode
	err = engine.DeleteObject(ctx, req.Bucket, req.Key, resp.VersionID, true, "GOV_BYPASS_AUTHORIZED_SEC_RULE_17A4")
	if !errors.Is(err, worm.ErrObjectLockedInCompliance) {
		t.Errorf("expected compliance mode to reject even with bypass token, got: %v", err)
	}

	// Verify retrieval works and data is intact
	rec, data, err := engine.GetObject(ctx, req.Bucket, req.Key, resp.VersionID)
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if string(data) != string(payload) {
		t.Errorf("expected payload %s, got %s", string(payload), string(data))
	}
	if rec.RetentionMode != worm.ModeCompliance {
		t.Errorf("expected ModeCompliance, got %s", rec.RetentionMode)
	}
}

func TestWORMStorageEngine_GovernanceModeBypass(t *testing.T) {
	engine := worm.NewWORMStorageEngine()
	ctx := context.Background()

	payload := []byte("Temporary Regulatory Staging Document")
	req := worm.PutObjectRequest{
		Bucket:          "growww-staging-worm",
		Key:             "reports/staging_01.json",
		Payload:         payload,
		RetentionMode:   worm.ModeGovernance,
		RetainUntilDate: time.Now().Add(30 * 24 * time.Hour),
		LegalHold:       false,
	}

	resp, err := engine.PutObject(ctx, req)
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	// Unprivileged delete should fail
	err = engine.DeleteObject(ctx, req.Bucket, req.Key, resp.VersionID, false, "")
	if !errors.Is(err, worm.ErrGovernanceBypassDenied) {
		t.Errorf("expected ErrGovernanceBypassDenied, got: %v", err)
	}

	// Privileged delete with authorized bypass token should succeed
	err = engine.DeleteObject(ctx, req.Bucket, req.Key, resp.VersionID, true, "GOV_BYPASS_AUTHORIZED_SEC_RULE_17A4")
	if err != nil {
		t.Fatalf("DeleteObject with valid governance token failed: %v", err)
	}

	// Verify object is gone
	_, _, err = engine.GetObject(ctx, req.Bucket, req.Key, resp.VersionID)
	if !errors.Is(err, worm.ErrObjectNotFound) {
		t.Errorf("expected ErrObjectNotFound, got: %v", err)
	}
}

func TestWORMStorageEngine_LegalHoldAndRetentionExtension(t *testing.T) {
	engine := worm.NewWORMStorageEngine()
	ctx := context.Background()

	initialRetain := time.Now().Add(10 * time.Minute)
	req := worm.PutObjectRequest{
		Bucket:          "growww-compliance",
		Key:             "kyc/user_101.pdf",
		Payload:         []byte("KYC Proof Documents"),
		RetentionMode:   worm.ModeCompliance,
		RetainUntilDate: initialRetain,
	}

	resp, err := engine.PutObject(ctx, req)
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	// Test retention shortening rejection
	shorterDate := initialRetain.Add(-5 * time.Minute)
	err = engine.ExtendRetention(ctx, req.Bucket, req.Key, resp.VersionID, shorterDate)
	if !errors.Is(err, worm.ErrRetentionShortening) {
		t.Errorf("expected ErrRetentionShortening, got: %v", err)
	}

	// Test retention extension success
	longerDate := initialRetain.Add(24 * time.Hour)
	err = engine.ExtendRetention(ctx, req.Bucket, req.Key, resp.VersionID, longerDate)
	if err != nil {
		t.Fatalf("ExtendRetention failed: %v", err)
	}

	// Test Legal Hold prevents deletion even if expired
	err = engine.SetLegalHold(ctx, req.Bucket, req.Key, resp.VersionID, true)
	if err != nil {
		t.Fatalf("SetLegalHold failed: %v", err)
	}

	err = engine.DeleteObject(ctx, req.Bucket, req.Key, resp.VersionID, true, "GOV_BYPASS_AUTHORIZED_SEC_RULE_17A4")
	if !errors.Is(err, worm.ErrLegalHoldActive) {
		t.Errorf("expected ErrLegalHoldActive, got: %v", err)
	}
}

func TestWORMStorageEngine_MerkleTreeAndBesuAnchoring(t *testing.T) {
	engine := worm.NewWORMStorageEngine()
	ctx := context.Background()

	bucket := "growww-audit-immutable"
	for i := 1; i <= 5; i++ {
		req := worm.PutObjectRequest{
			Bucket:          bucket,
			Key:             "logs/batch_" + string(rune(i+'0')) + ".log",
			Payload:         []byte("Audit log event payload number " + string(rune(i+'0'))),
			RetentionMode:   worm.ModeCompliance,
			RetainUntilDate: time.Now().Add(7 * 365 * 24 * time.Hour),
		}
		if _, err := engine.PutObject(ctx, req); err != nil {
			t.Fatalf("PutObject failed for log %d: %v", i, err)
		}
	}

	// Seal batch and anchor to Hyperledger Besu
	receipt, err := engine.SealBatchAndAnchor(ctx, bucket)
	if err != nil {
		t.Fatalf("SealBatchAndAnchor failed: %v", err)
	}

	if receipt.MerkleRoot == "" {
		t.Errorf("expected valid MerkleRoot")
	}
	if receipt.TransactionHash == "" || receipt.BlockNumber == 0 {
		t.Errorf("expected valid transaction hash and block number")
	}

	// Verify all records updated with on-chain anchoring info
	rec, _, err := engine.GetObject(ctx, bucket, "logs/batch_1.log", "")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	if rec.BesuTxHash != receipt.TransactionHash {
		t.Errorf("expected Besu tx hash %s, got %s", receipt.TransactionHash, rec.BesuTxHash)
	}
}

func TestMerkleTree_InclusionProof(t *testing.T) {
	data := [][]byte{
		[]byte("tx1"),
		[]byte("tx2"),
		[]byte("tx3"),
		[]byte("tx4"),
	}

	hashes := make([][]byte, len(data))
	for i, d := range data {
		h := sha256.Sum256(d)
		hashes[i] = h[:]
	}

	tree, err := worm.NewMerkleTree(hashes)
	if err != nil {
		t.Fatalf("NewMerkleTree failed: %v", err)
	}

	rootHex := tree.RootHex()

	// Test inclusion proof for each leaf
	for i := 0; i < len(hashes); i++ {
		proof, err := tree.GenerateProof(i)
		if err != nil {
			t.Fatalf("GenerateProof failed for index %d: %v", i, err)
		}

		valid := worm.VerifyProof(proof, rootHex)
		if !valid {
			t.Errorf("Merkle proof verification failed for leaf %d", i)
		}

		// Tamper proof and assert rejection
		proof.LeafHash = hex.EncodeToString([]byte("tampered_leaf_hash_0000000000000"))
		if worm.VerifyProof(proof, rootHex) {
			t.Errorf("tampered proof should fail verification")
		}
	}
}
