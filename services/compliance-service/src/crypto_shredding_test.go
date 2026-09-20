package main

import (
	"strings"
	"testing"
	"time"
)

func TestCryptoShreddingEngine_FullLifecycle(t *testing.T) {
	engine := NewCryptoShreddingKMSEngine()
	userID := "user-compliance-dpdp-001"

	// 1. Provision KMS Key
	kmsKey, err := engine.ProvisionUserKMSKey(userID)
	if err != nil || kmsKey.Status != KeyStatusActive {
		t.Fatalf("failed to provision user KMS key: %v", err)
	}

	// 2. Encrypt User PII (Aadhaar, PAN, Biometric, Bank)
	pan := "ABCDE1234F"
	aadhaar := "234567891234"
	biometric := "BIO_TEMPLATE_FACE_LANDMARKS_DATA"
	bankAcct := "HDFC0001234_9876543210"

	envelope, err := engine.EncryptUserPII(userID, pan, aadhaar, biometric, bankAcct)
	if err != nil {
		t.Fatalf("failed to encrypt user PII: %v", err)
	}
	if len(envelope.EncryptedPAN) == 0 || string(envelope.EncryptedPAN) == pan {
		t.Fatalf("PII is not properly encrypted")
	}

	// 3. Verify Decryption while key is ACTIVE
	decPAN, decAadhaar, decBio, decBank, err := engine.DecryptUserPII(userID)
	if err != nil {
		t.Fatalf("decryption failed while key active: %v", err)
	}
	if decPAN != pan || decAadhaar != aadhaar || decBio != biometric || decBank != bankAcct {
		t.Fatalf("decrypted PII does not match original plaintext")
	}

	// 4. Request Shredding when 7-year retention has passed (>7 years ago)
	eightYearsAgo := time.Now().UTC().AddDate(-8, 0, 0)
	req, err := engine.RequestCryptoShredding(userID, eightYearsAgo, eightYearsAgo, false)
	if err != nil || req.Status != ShreddingRequested {
		t.Fatalf("expected valid shredding request after 7 years: %v", err)
	}

	// 5. Dual Authorization and Destruction Execution
	complianceSign := GenerateMockSignature()
	dpoSign := GenerateMockSignature()
	cert, err := engine.AuthorizeAndExecuteShredding(req.RequestID, complianceSign, dpoSign)
	if err != nil {
		t.Fatalf("failed to execute crypto-shredding: %v", err)
	}

	if cert.KeyStatus != KeyStatusDestroyed || !cert.IsUnrecoverable {
		t.Fatalf("expected key status DESTROYED, got %s", cert.KeyStatus)
	}
	if cert.CiphertextEntropy < 6.0 {
		t.Fatalf("expected high Shannon entropy for shredded ciphertext, got %.2f", cert.CiphertextEntropy)
	}
	if !strings.HasPrefix(cert.BesuAttestationHash, "0x") {
		t.Fatalf("expected Besu attestation hash starting with 0x")
	}

	// 6. Verify DECRYPTION NOW FAILS MATHEMATICALLY (Zero Key Material)
	_, _, _, _, err = engine.DecryptUserPII(userID)
	if err == nil {
		t.Fatalf("expected decryption to FAIL after key destruction, but succeeded!")
	}
	if !strings.Contains(err.Error(), "permanently destroyed") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestCryptoShreddingEngine_StatutoryRetentionAndRegulatoryHold(t *testing.T) {
	engine := NewCryptoShreddingKMSEngine()
	userID := "user-recent-tx-002"

	_, err := engine.ProvisionUserKMSKey(userID)
	if err != nil {
		t.Fatalf("failed to provision key: %v", err)
	}

	_, _ = engine.EncryptUserPII(userID, "ABCDE1234F", "123456789012", "BIO", "BANK")

	// 1. Rejection when last transaction was recent (< 7 years ago, e.g. 2 years ago)
	twoYearsAgo := time.Now().UTC().AddDate(-2, 0, 0)
	req, err := engine.RequestCryptoShredding(userID, twoYearsAgo, twoYearsAgo, false)
	if err == nil || req.Status != ShreddingHeld {
		t.Fatalf("expected request to be held due to 7-year statutory retention requirement")
	}

	// 2. Rejection when under active regulatory inquiry
	tenYearsAgo := time.Now().UTC().AddDate(-10, 0, 0)
	req2, err2 := engine.RequestCryptoShredding(userID, tenYearsAgo, tenYearsAgo, true)
	if err2 == nil || req2.Status != ShreddingHeld {
		t.Fatalf("expected request to be held due to active regulatory inquiry")
	}
}
