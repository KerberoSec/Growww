package main

import (
	"strings"
	"testing"
	"time"
)

func TestDeepfakeVideoKYCFilterEngine_ValidAndSpoofScenarios(t *testing.T) {
	engine := NewDeepfakeVideoKYCFilterEngine()
	userID := "user-kyc-deepfake-001"

	// 1. Issue challenge
	chal, err := engine.IssueLivenessChallenge(userID, ChallengeHeadTurnLeft)
	if err != nil || chal.ChallengeID == "" {
		t.Fatalf("failed to issue liveness challenge: %v", err)
	}

	// 2. Legitimate human verification (High PAD, high rPPG, high A/V sync, low tamper)
	validTelemetry := &VideoStreamTelemetry{
		SessionID:                   "session-valid-001",
		UserID:                      userID,
		DurationSeconds:             5.0,
		FrameRateFPS:                30.0,
		TotalFramesAnalyzed:         150,
		PADScore:                    0.96,
		RPPGLivenessSignalScore:     0.92,
		AVLipSyncCorrelationScore:   0.90,
		BoundaryTamperArtifactScore: 0.02,
		MicroExpressionDynamicsScore: 0.95,
		ChallengeResponseCompleted:  true,
		RawVideoSHA256:              "a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0",
	}

	report, err := engine.EvaluateVideoKYCStream(chal.ChallengeID, validTelemetry)
	if err != nil {
		t.Fatalf("unexpected error during valid stream evaluation: %v", err)
	}
	if report.Status != StatusDeepfakePassed {
		t.Fatalf("expected status PASSED, got %s (reason: %s)", report.Status, report.RejectionReason)
	}
	if report.DeepfakeRiskScore > 0.15 {
		t.Fatalf("expected deepfake risk score <= 0.15, got %.3f", report.DeepfakeRiskScore)
	}
	if !strings.HasPrefix(report.BesuOnChainAnchorTxHash, "0x") {
		t.Fatalf("expected valid Besu on-chain anchor hash")
	}

	// 3. Replay Attack Prevention Check (Re-submitting same challenge)
	_, err = engine.EvaluateVideoKYCStream(chal.ChallengeID, validTelemetry)
	if err == nil {
		t.Fatalf("expected replay attack error when reusing challenge nonce")
	}

	// 4. Deepfake Spoof Scenario (Screen replay / Face-swap: Low PAD, low rPPG, high boundary tamper)
	chal2, _ := engine.IssueLivenessChallenge(userID, ChallengeBlinkTwice)
	spoofTelemetry := &VideoStreamTelemetry{
		SessionID:                   "session-spoof-002",
		UserID:                      userID,
		DurationSeconds:             4.0,
		FrameRateFPS:                30.0,
		TotalFramesAnalyzed:         120,
		PADScore:                    0.65, // Below 0.90 threshold
		RPPGLivenessSignalScore:     0.40, // Below 0.75 threshold
		AVLipSyncCorrelationScore:   0.50, // Lip sync mismatch
		BoundaryTamperArtifactScore: 0.85, // High blending artifact
		MicroExpressionDynamicsScore: 0.30,
		ChallengeResponseCompleted:  true,
		RawVideoSHA256:              "f9e8d7c6b5a432109876543210fedcba09876543210fedcba09876543210fedc",
	}

	spoofReport, err := engine.EvaluateVideoKYCStream(chal2.ChallengeID, spoofTelemetry)
	if err != nil {
		t.Fatalf("unexpected error during spoof evaluation: %v", err)
	}
	if spoofReport.Status != StatusDeepfakeRejected {
		t.Fatalf("expected status REJECTED_SPOOF, got %s", spoofReport.Status)
	}
}

func TestCryptoAMLScoringEngine_MultiHopAndSanctions(t *testing.T) {
	engine := NewCryptoAMLScoringEngine()
	engine.AddSanctionedAddress("0xSanctionedLazarusGroup0000000000000001")

	// 1. Direct Sanctions Hit
	sanctionReq := &CryptoAMLScreeningRequest{
		RequestID:     "req-ofac-01",
		UserID:        "user-bad-actor",
		WalletAddress: "0xSanctionedLazarusGroup0000000000000001",
		Blockchain:    "ETHEREUM",
		AmountUSD:     50000.0,
	}

	res, err := engine.ScreenCryptoTransaction(sanctionReq)
	if err != nil || res.Decision != DecisionRejectAndFreeze || !res.STRTriggered {
		t.Fatalf("expected REJECT_AND_FREEZE and STR Triggered for sanctioned address, got: %+v", res)
	}

	// 2. Multi-hop Exposure Screening (Indirect Mixer & Ransomware exposure)
	multiHopReq := &CryptoAMLScreeningRequest{
		RequestID:     "req-multihop-02",
		UserID:        "user-crypto-trader",
		WalletAddress: "0xCleanLookingWallet1234567890abcdef1234",
		Blockchain:    "BESU",
		AmountUSD:     15000.0,
		Exposures: []ExposureDetail{
			{Category: ExposureLicensedVASP, Percentage: 0.60, HopDistance: 1},
			{Category: ExposureMixersTumblers, Percentage: 0.30, HopDistance: 2}, // 75 * 0.30 * 0.5 = 11.25
			{Category: ExposureDarknetMarkets, Percentage: 0.10, HopDistance: 3}, // 85 * 0.10 * 0.25 = 2.125
		},
		EllipticScore: 2.5, // 25.0
	}

	res2, err := engine.ScreenCryptoTransaction(multiHopReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res2.Decision != DecisionApprove && res2.Decision != DecisionReviewEDD {
		t.Fatalf("unexpected decision for low-risk multi-hop exposure: %s", res2.Decision)
	}

	// 3. Smart Contract Transfer Compliance Hook
	allowed, reason := engine.VerifyTransferCompliance("0xCleanWalletA", "0xCleanWalletB", 1000.0)
	if !allowed {
		t.Fatalf("expected transfer allowed between clean wallets, reason: %s", reason)
	}

	allowed2, reason2 := engine.VerifyTransferCompliance("0xSanctionedLazarusGroup0000000000000001", "0xCleanWalletB", 1000.0)
	if allowed2 {
		t.Fatalf("expected transfer blocked from sanctioned address, got reason: %s", reason2)
	}
}

func TestAdverseMediaScreeningNLPEngine_ScreeningAndFuzzyMatching(t *testing.T) {
	engine := NewAdverseMediaScreeningNLPEngine()

	// 1. Ingest regulatory and news articles
	engine.IngestArticle(&AdverseMediaArticle{
		ArticleID:       "art-sebi-001",
		Title:           "SEBI bars Nirav Modi associate over diamond round-tripping money laundering scam",
		Source:          "SEBI Enforcement Orders",
		SourceTier:      SourceTier1RegulatoryPress,
		PublishedAt:     time.Now().UTC().AddDate(0, -2, 0), // 2 months ago
		ExtractedText:   "Investigation revealed that Rajesh Kumar Sharma assisted in funneling ₹500 Crore through overseas shell companies.",
		MatchedEntities: []string{"Rajesh Kumar Sharma", "Nirav Modi"},
		OffenseCategory: OffenseMoneyLaundering,
		SentimentScore:  -0.95,
	})

	engine.IngestArticle(&AdverseMediaArticle{
		ArticleID:       "art-reuters-002",
		Title:           "Interpol red corner notice issued for international terror financing syndicate",
		Source:          "Reuters",
		SourceTier:      SourceTier1RegulatoryPress,
		PublishedAt:     time.Now().UTC().AddDate(0, -6, 0),
		ExtractedText:   "Key operative Dawood Ibrahim associate Tariq Ahmed implicated in Hawala transfers.",
		MatchedEntities: []string{"Tariq Ahmed", "Dawood Ibrahim"},
		OffenseCategory: OffenseTerrorFinancing,
		SentimentScore:  -1.0,
	})

	// 2. Screen entity with exact/substring match
	report, err := engine.ScreenEntity(&AdverseMediaScreeningRequest{
		RequestID:   "req-media-001",
		EntityID:    "entity-9988",
		FullName:    "Rajesh Kumar Sharma",
		Aliases:     []string{"Rajesh Sharma"},
		CountryCode: "IND",
		IsPEP:       false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Decision != AdverseDecisionProhibited && report.Decision != AdverseDecisionEDDRequired {
		t.Fatalf("expected adverse media hit, got: %s (risk score: %.2f)", report.Decision, report.AdverseMediaRiskScore)
	}
	if report.MatchedArticlesCount == 0 {
		t.Fatalf("expected matched article count > 0")
	}

	// 3. Screen entity with high fuzzy similarity (e.g. "Tarique Ahmad" for "Tariq Ahmed")
	fuzzyReport, err := engine.ScreenEntity(&AdverseMediaScreeningRequest{
		RequestID:   "req-media-002",
		EntityID:    "entity-9989",
		FullName:    "Tarique Ahmad",
		CountryCode: "IND",
		IsPEP:       true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fuzzyReport.Decision != AdverseDecisionProhibited {
		t.Fatalf("expected PROHIBITED_MATCH on terror financing fuzzy match, got: %s", fuzzyReport.Decision)
	}

	// 4. Clean Entity
	cleanReport, err := engine.ScreenEntity(&AdverseMediaScreeningRequest{
		RequestID:   "req-media-clean",
		EntityID:    "entity-clean-01",
		FullName:    "Aarav Sundaram",
		CountryCode: "IND",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cleanReport.Decision != AdverseDecisionPass || cleanReport.MatchedArticlesCount != 0 {
		t.Fatalf("expected clean PASS for unrelated entity, got: %s", cleanReport.Decision)
	}
}
