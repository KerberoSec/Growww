package main

import (
	"fmt"
	"testing"
	"time"
)

func TestVaultAuditEngine_FullLifecycle(t *testing.T) {
	engine := NewVaultAuditPortalEngine()

	// 1. Register accredited vault facility
	facility := &VaultFacility{
		VaultID:                "VAULT-MUM-001",
		VaultName:              "Brinks BKC Precious Metals Vault",
		CustodianCompanyName:   "Brinks India Pvt Ltd",
		MCXAccreditationNumber: "MCX-VAULT-2024-089",
		WDRARegistrationNumber: "WDRA-MH-MUM-4421",
		FacilityAddress:        "Bandra Kurla Complex, Mumbai, Maharashtra 400051",
		Latitude:               19.0688,
		Longitude:              72.8708,
		GeofenceRadiusMeters:   150,
		IsActive:               true,
		CreatedAt:              time.Now().UTC(),
	}
	if err := engine.RegisterFacility(facility); err != nil {
		t.Fatalf("failed to register vault facility: %v", err)
	}

	// 2. Ingest WDRA e-NWR record from NERL
	enwr := &WDRAeNWRRecord{
		ENWRNumber:            "NERL-ENWR-2024-998811",
		Repository:            RepoNERL,
		VaultID:               "VAULT-MUM-001",
		BeneficiaryClientCode: "GROWWW-INST-01",
		CommodityType:         CommodityGold9999,
		TotalWeightGrams:      1000.0,
		PurityBps:             9999,
		IsPledged:             false,
		IsLienMarked:          false,
		ReceiptStatus:         "ACTIVE",
		ValidityStart:         time.Now().UTC().AddDate(0, -1, 0),
		ValidityEnd:           time.Now().UTC().AddDate(0, 11, 0),
	}
	if err := engine.IngestWDRAeNWR(enwr); err != nil {
		t.Fatalf("failed to ingest e-NWR: %v", err)
	}

	// Verify valid e-NWR
	verifiedENWR, err := engine.VerifyEnwrReceipt(RepoNERL, "NERL-ENWR-2024-998811", "VAULT-MUM-001", "GROWWW-INST-01")
	if err != nil || !verifiedENWR.ValidityEnd.After(time.Now().UTC()) {
		t.Fatalf("expected valid e-NWR, got error: %v", err)
	}

	// 3. Create Audit Session
	session, err := engine.CreateAuditSession(
		"VAULT-MUM-001",
		AuditPeriodicCycleCount,
		"INSPECTOR-UUID-001",
		"ASSAYER-UUID-002",
		"CUSTODIAN-UUID-003",
		"MANDATE-Q3-2024-GOLD",
		[]string{"BAR-MMTC-001", "BAR-MMTC-002"},
	)
	if err != nil {
		t.Fatalf("failed to create audit session: %v", err)
	}

	// 4. Submit IoT Scale Weighment (Mettler Toledo) with valid variance (< 0.001%)
	calibExpiry := time.Now().UTC().AddDate(0, 6, 0)
	scaleSig := GenerateMockSignature()
	receipt, disc, err := engine.SubmitBarWeighment(
		session.SessionID,
		"VAULT-MUM-001",
		"BAR-MMTC-001",
		CommodityGold9999,
		"SCALE-METTLER-XPR",
		"CERT-NABL-CALIB-882",
		1005.0000,
		5.0000,
		1000.0000,
		1000.0005, // 0.00005% variance <= 0.001%
		"RAW_TELEMETRY_STREAM_HEX_XYZ",
		scaleSig,
		calibExpiry,
	)
	if err != nil || disc != nil || receipt == nil {
		t.Fatalf("expected successful weighment without discrepancy, got err: %v, disc: %+v", err, disc)
	}

	// 5. Submit Assay Attestation (MMTC-PAMP, 9999 fineness, 3240 m/s ultrasonic velocity)
	assayerSig := GenerateMockSignature()
	att, discPurity, err := engine.SubmitAssayAttestation(
		session.SessionID,
		"VAULT-MUM-001",
		"BAR-MMTC-001",
		"MMTC-PAMP",
		"LOT-2024-99",
		"LAB-NABL-099",
		"NABL-TC-5541",
		CommodityGold9999,
		9999,
		3240.0,
		99.99,
		"s3://vault-audit/certificates/bar-001.pdf",
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"0x04abc...",
		assayerSig,
	)
	if err != nil || discPurity != nil || att == nil {
		t.Fatalf("expected successful assay attestation, got err: %v, disc: %+v", err, discPurity)
	}

	// Add bar to inventory
	engine.AddOrUpdateBar(&VaultBar{
		BarID:               "BAR-UUID-001",
		VaultID:             "VAULT-MUM-001",
		BarSerialNumber:     "BAR-MMTC-001",
		RefineryName:        "MMTC-PAMP",
		RefineryBatchNumber: "LOT-2024-99",
		CommodityType:       CommodityGold9999,
		PurityFinenessBps:   9999,
		GrossWeightGrams:    1000.0,
		ENWRNumber:          "NERL-ENWR-2024-998811",
		TamperBagSealNumber: "SEAL-998822",
		IsQuarantined:       false,
	})

	// Add 2nd bar
	engine.AddOrUpdateBar(&VaultBar{
		BarID:               "BAR-UUID-002",
		VaultID:             "VAULT-MUM-001",
		BarSerialNumber:     "BAR-MMTC-002",
		RefineryName:        "MMTC-PAMP",
		RefineryBatchNumber: "LOT-2024-99",
		CommodityType:       CommodityGold9999,
		PurityFinenessBps:   9999,
		GrossWeightGrams:    1000.0,
		ENWRNumber:          "NERL-ENWR-2024-998812",
		TamperBagSealNumber: "SEAL-998823",
		IsQuarantined:       false,
	})

	// 6. Finalize session with 3-of-3 quorum
	finalSession, err := engine.FinalizeAuditSession(
		session.SessionID,
		GenerateMockSignature(),
		GenerateMockSignature(),
		GenerateMockSignature(),
		"s3://vault-audit/reports/session-001.pdf",
		"sha256-report-digest",
	)
	if err != nil {
		t.Fatalf("failed to finalize audit session: %v", err)
	}
	if finalSession.Status != StatusAttestationCompleted {
		t.Fatalf("expected ATTESTATION_COMPLETED, got %s", finalSession.Status)
	}
	if finalSession.InventoryMerkleRoot == "" || finalSession.OnChainAttestationTxHash == "" {
		t.Fatalf("expected non-empty Merkle root and Besu tx hash")
	}

	// 7. Get Proof of Reserve Snapshot
	snapshot, err := engine.GetProofOfReserveSnapshot("VAULT-MUM-001", CommodityGold9999, 1999.0)
	if err != nil {
		t.Fatalf("failed to get PoR snapshot: %v", err)
	}
	if !snapshot["is_fully_backed"].(bool) {
		t.Fatalf("expected 100%% fully backed reserve")
	}
}

func TestVaultAuditEngine_DiscrepanciesAndCircuitBreaker(t *testing.T) {
	engine := NewVaultAuditPortalEngine()
	_ = engine.RegisterFacility(&VaultFacility{
		VaultID:                "VAULT-DEL-002",
		VaultName:              "Sequel Delhi Bullion Vault",
		CustodianCompanyName:   "Sequel Logistics",
		MCXAccreditationNumber: "MCX-VAULT-2024-099",
		WDRARegistrationNumber: "WDRA-DL-001",
		FacilityAddress:        "IGI Cargo Terminal, New Delhi",
		IsActive:               true,
	})

	session, _ := engine.CreateAuditSession(
		"VAULT-DEL-002",
		AuditInboundIngress,
		"INSP-01", "ASSAY-01", "CUST-01", "MANDATE-TEST",
		[]string{"BAR-FAKE-01"},
	)

	// 1. High Weight Variance (> 0.001%)
	calibExpiry := time.Now().UTC().AddDate(0, 6, 0)
	_, discWeight, err := engine.SubmitBarWeighment(
		session.SessionID, "VAULT-DEL-002", "BAR-BAD-WEIGHT",
		CommodityGold9999, "SCALE-01", "CERT-01",
		1010.0, 10.0, 1000.0, 990.0, // Variance: 1% >> 0.001%
		"RAW", GenerateMockSignature(), calibExpiry,
	)
	if err != nil {
		t.Fatalf("unexpected error during weighment: %v", err)
	}
	if discWeight == nil || discWeight.Severity != SeverityHighWeightVariance {
		t.Fatalf("expected HIGH_WEIGHT_VARIANCE discrepancy, got: %+v", discWeight)
	}

	// 2. Ultrasonic Velocity Anomaly (Tungsten Core)
	_, discDensity, err := engine.SubmitAssayAttestation(
		session.SessionID, "VAULT-DEL-002", "BAR-TUNGSTEN-01",
		"MMTC-PAMP", "BATCH-01", "LAB-01", "NABL-01",
		CommodityGold9999, 9999,
		5100.0, // Tungsten speed of sound ~ 5200 m/s vs Gold 3240 m/s
		99.99, "s3://uri", "sha", "pubkey", "sig",
	)
	if err == nil || discDensity == nil || !discDensity.EmergencyMintHaltTriggered {
		t.Fatalf("expected emergency mint halt on tungsten velocity anomaly")
	}

	// Verify vault mint is frozen
	if !engine.mintFrozenVaults["VAULT-DEL-002"] {
		t.Fatalf("expected vault to be mint frozen")
	}

	// 3. Resolve Discrepancy with MFA
	err = engine.ResolveDiscrepancy(discDensity.DiscrepancyID, "OFFICER-001", "BAR_QUARANTINED_AND_RETURNED", "Bar quarantined and refiner notified", "MFA-OTP-998811")
	if err != nil {
		t.Fatalf("failed to resolve discrepancy: %v", err)
	}

	// Verify vault is unfrozen after resolution
	if engine.mintFrozenVaults["VAULT-DEL-002"] {
		t.Fatalf("expected vault to be unfrozen after discrepancy resolution")
	}
}

func TestVaultAuditEngine_MerkleTree10000BarsPerformance(t *testing.T) {
	// Generate 10,000 bar hashes and verify construction in < 500ms and inclusion proof in < 1ms
	const numBars = 10000
	hashes := make([]string, numBars)
	for i := 0; i < numBars; i++ {
		hashes[i] = ComputeBarHash(fmt.Sprintf("BAR-PERF-%05d", i), 1000.0, 9999, "VAULT-MUM-001")
	}

	start := time.Now()
	root, proofs := BuildMerkleTree(hashes)
	treeDuration := time.Since(start)

	if treeDuration > 500*time.Millisecond {
		t.Fatalf("Merkle tree construction for %d bars took too long: %v (expected < 500ms)", numBars, treeDuration)
	}
	if root == "" || len(proofs) != numBars {
		t.Fatalf("invalid tree: root=%s, proofs count=%d", root, len(proofs))
	}

	// Verify arbitrary inclusion proof
	testProof := proofs[hashes[4242]]
	proofStart := time.Now()
	valid := VerifyMerkleProof(testProof)
	proofDuration := time.Since(proofStart)

	if !valid {
		t.Fatalf("Merkle inclusion proof verification failed")
	}
	if proofDuration > 1*time.Millisecond {
		t.Fatalf("Merkle proof verification took too long: %v (expected < 1ms)", proofDuration)
	}
}
