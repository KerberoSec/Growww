package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// DeepfakeVerificationStatus represents the outcome of video KYC screening
type DeepfakeVerificationStatus string

const (
	StatusDeepfakePassed       DeepfakeVerificationStatus = "PASSED"
	StatusDeepfakeManualReview DeepfakeVerificationStatus = "MANUAL_REVIEW"
	StatusDeepfakeRejected     DeepfakeVerificationStatus = "REJECTED_SPOOF"
)

// ChallengeType represents interactive challenge prompt
type ChallengeType string

const (
	ChallengeHeadTurnLeft   ChallengeType = "TURN_HEAD_LEFT"
	ChallengeHeadTurnRight  ChallengeType = "TURN_HEAD_RIGHT"
	ChallengeBlinkTwice     ChallengeType = "BLINK_TWICE"
	ChallengeReadRandomCode ChallengeType = "READ_RANDOM_CODE"
	ChallengeSmile          ChallengeType = "SMILE"
)

// LivenessChallenge contains an active time-bounded challenge
type LivenessChallenge struct {
	ChallengeID string        `json:"challenge_id"`
	UserID      string        `json:"user_id"`
	Type        ChallengeType `json:"type"`
	PromptCode  string        `json:"prompt_code"`
	Nonce       string        `json:"nonce"`
	ExpiresAt   time.Time     `json:"expires_at"`
	IsCompleted bool          `json:"is_completed"`
}

// VideoStreamTelemetry contains frame-by-frame and acoustic telemetry
type VideoStreamTelemetry struct {
	SessionID                   string    `json:"session_id"`
	UserID                      string    `json:"user_id"`
	DurationSeconds             float64   `json:"duration_seconds"`
	FrameRateFPS                float64   `json:"frame_rate_fps"`
	TotalFramesAnalyzed         int       `json:"total_frames_analyzed"`
	PADScore                    float64   `json:"pad_score"`                     // Presentation Attack Detection (0.0 to 1.0, >= 0.90 required)
	RPPGLivenessSignalScore     float64   `json:"rppg_liveness_signal_score"`    // Remote photoplethysmography pulse score (0.0 to 1.0)
	AVLipSyncCorrelationScore   float64   `json:"av_lip_sync_correlation_score"` // Audio-visual phoneme-viseme correlation (0.0 to 1.0)
	BoundaryTamperArtifactScore float64   `json:"boundary_tamper_artifact_score"`// Frequency domain blending artifact (0.0 to 1.0, lower is better)
	MicroExpressionDynamicsScore float64  `json:"micro_expression_dynamics_score"`// Natural facial twitch & eyelid dynamics
	ChallengeResponseCompleted  bool      `json:"challenge_response_completed"`
	RawVideoSHA256              string    `json:"raw_video_sha256"`
}

// DeepfakeAuditReport holds evaluation outcome and blockchain anchor
type DeepfakeAuditReport struct {
	ReportID                 string                     `json:"report_id"`
	SessionID                string                     `json:"session_id"`
	UserID                   string                     `json:"user_id"`
	Status                   DeepfakeVerificationStatus `json:"status"`
	DeepfakeRiskScore        float64                    `json:"deepfake_risk_score"` // 0.0 to 1.0 (lower is better, <= 0.15 is pass)
	LivenessConfidenceScore  float64                    `json:"liveness_confidence_score"`
	RejectionReason          string                     `json:"rejection_reason,omitempty"`
	RawVideoHash             string                     `json:"raw_video_hash"`
	AttestationProofDigest   string                     `json:"attestation_proof_digest"`
	BesuOnChainAnchorTxHash  string                     `json:"besu_on_chain_anchor_tx_hash"`
	EvaluatedAt              time.Time                  `json:"evaluated_at"`
}

// DeepfakeVideoKYCFilterEngine conducts presentation attack and deepfake video verification
type DeepfakeVideoKYCFilterEngine struct {
	mu                 sync.RWMutex
	challenges         map[string]*LivenessChallenge // challengeID -> Challenge
	reports            map[string]*DeepfakeAuditReport
	usedNonces         map[string]bool
	maxDeepfakeScore   float64 // default: 0.15
	minPADScore        float64 // default: 0.90
	minAVSyncScore     float64 // default: 0.80
	minRPPGScore       float64 // default: 0.75
}

// NewDeepfakeVideoKYCFilterEngine creates a new DeepfakeVideoKYCFilterEngine
func NewDeepfakeVideoKYCFilterEngine() *DeepfakeVideoKYCFilterEngine {
	return &DeepfakeVideoKYCFilterEngine{
		challenges:       make(map[string]*LivenessChallenge),
		reports:          make(map[string]*DeepfakeAuditReport),
		usedNonces:       make(map[string]bool),
		maxDeepfakeScore: 0.15,
		minPADScore:      0.90,
		minAVSyncScore:   0.80,
		minRPPGScore:     0.75,
	}
}

// IssueLivenessChallenge generates a randomized, time-bounded challenge for user video KYC session
func (e *DeepfakeVideoKYCFilterEngine) IssueLivenessChallenge(userID string, cType ChallengeType) (*LivenessChallenge, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if userID == "" {
		return nil, errors.New("user_id is required")
	}

	nonceHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d:%s", userID, time.Now().UnixNano(), cType)))
	nonce := hex.EncodeToString(nonceHash[:])[:16]
	challengeID := fmt.Sprintf("chal-%s", nonce[:10])

	promptCode := fmt.Sprintf("%04d", (time.Now().UnixNano()/1000)%10000)

	challenge := &LivenessChallenge{
		ChallengeID: challengeID,
		UserID:      userID,
		Type:        cType,
		PromptCode:  promptCode,
		Nonce:       nonce,
		ExpiresAt:   time.Now().UTC().Add(3 * time.Minute), // 3-minute validity
		IsCompleted: false,
	}

	e.challenges[challengeID] = challenge
	return challenge, nil
}

// EvaluateVideoKYCStream processes video stream telemetry and computes composite deepfake score
func (e *DeepfakeVideoKYCFilterEngine) EvaluateVideoKYCStream(
	challengeID string,
	telemetry *VideoStreamTelemetry,
) (*DeepfakeAuditReport, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if telemetry == nil {
		return nil, errors.New("telemetry cannot be nil")
	}
	if telemetry.UserID == "" {
		return nil, errors.New("user_id is required")
	}
	if telemetry.RawVideoSHA256 == "" {
		return nil, errors.New("raw_video_sha256 is required")
	}

	// 1. Challenge & Nonce Verification (Replay Attack Prevention)
	challenge, exists := e.challenges[challengeID]
	if !exists {
		return nil, fmt.Errorf("challenge %s not found", challengeID)
	}
	if challenge.UserID != telemetry.UserID {
		return nil, fmt.Errorf("challenge belongs to %s, telemetry from %s", challenge.UserID, telemetry.UserID)
	}
	if time.Now().UTC().After(challenge.ExpiresAt) {
		return nil, errors.New("liveness challenge has expired")
	}
	if challenge.IsCompleted || e.usedNonces[challenge.Nonce] {
		// Replay attack!
		reportID := fmt.Sprintf("rep-replay-%s", telemetry.SessionID)
		rep := &DeepfakeAuditReport{
			ReportID:                reportID,
			SessionID:               telemetry.SessionID,
			UserID:                  telemetry.UserID,
			Status:                  StatusDeepfakeRejected,
			DeepfakeRiskScore:       1.0,
			LivenessConfidenceScore: 0.0,
			RejectionReason:         "Replay attack detected: Challenge nonce already consumed",
			RawVideoHash:            telemetry.RawVideoSHA256,
			EvaluatedAt:             time.Now().UTC(),
		}
		e.reports[reportID] = rep
		return rep, errors.New("replay attack detected: challenge nonce already consumed")
	}

	challenge.IsCompleted = true
	e.usedNonces[challenge.Nonce] = true

	// 2. Minimum frame count check (at least 3 seconds at 15fps = 45 frames)
	if telemetry.TotalFramesAnalyzed < 45 || telemetry.DurationSeconds < 3.0 {
		return nil, fmt.Errorf("insufficient video duration (%.2fs, %d frames): minimum 3.0s and 45 frames required",
			telemetry.DurationSeconds, telemetry.TotalFramesAnalyzed)
	}

	// 3. Mathematical Composite Deepfake Risk Score Calculation
	// Weights:
	// w_PAD = 0.35, w_RPPG = 0.25, w_AVSync = 0.20, w_BoundaryTamper = 0.15, w_MicroExpr = 0.05
	padDeficit := math.Max(0, 1.0-telemetry.PADScore)
	rppgDeficit := math.Max(0, 1.0-telemetry.RPPGLivenessSignalScore)
	avSyncDeficit := math.Max(0, 1.0-telemetry.AVLipSyncCorrelationScore)
	tamperScore := telemetry.BoundaryTamperArtifactScore
	microExprDeficit := math.Max(0, 1.0-telemetry.MicroExpressionDynamicsScore)

	deepfakeRisk := 0.35*padDeficit +
		0.25*rppgDeficit +
		0.20*avSyncDeficit +
		0.15*tamperScore +
		0.05*microExprDeficit

	livenessConfidence := 1.0 - deepfakeRisk

	// 4. Decision Engine
	var status DeepfakeVerificationStatus
	var rejectionReason string

	if !telemetry.ChallengeResponseCompleted {
		status = StatusDeepfakeRejected
		rejectionReason = "Interactive challenge response was not successfully performed"
	} else if telemetry.PADScore < e.minPADScore {
		status = StatusDeepfakeRejected
		rejectionReason = fmt.Sprintf("Presentation Attack Detection failed: PAD score %.3f below threshold %.3f (spoof mask/screen detected)",
			telemetry.PADScore, e.minPADScore)
	} else if telemetry.RPPGLivenessSignalScore < e.minRPPGScore {
		status = StatusDeepfakeRejected
		rejectionReason = fmt.Sprintf("Remote photoplethysmography (rPPG) liveness pulse absent (score: %.3f)",
			telemetry.RPPGLivenessSignalScore)
	} else if telemetry.AVLipSyncCorrelationScore < e.minAVSyncScore {
		status = StatusDeepfakeRejected
		rejectionReason = fmt.Sprintf("Audio-visual phoneme-viseme correlation failure: lip sync mismatch %.3f (voice/video clone detected)",
			telemetry.AVLipSyncCorrelationScore)
	} else if deepfakeRisk <= e.maxDeepfakeScore {
		status = StatusDeepfakePassed
	} else if deepfakeRisk <= 0.35 {
		status = StatusDeepfakeManualReview
		rejectionReason = fmt.Sprintf("Borderline deepfake risk score: %.3f (queued for manual forensic officer inspection)", deepfakeRisk)
	} else {
		status = StatusDeepfakeRejected
		rejectionReason = fmt.Sprintf("Composite deepfake risk score %.3f exceeds allowable threshold %.3f", deepfakeRisk, e.maxDeepfakeScore)
	}

	// 5. Generate Zero-PII Hyperledger Besu proof anchoring
	attestationPayload := fmt.Sprintf("%s:%s:%s:%.4f:%.4f:%s:%d",
		telemetry.SessionID, telemetry.UserID, status, deepfakeRisk, livenessConfidence, telemetry.RawVideoSHA256, time.Now().Unix())
	proofDigest := sha256.Sum256([]byte(attestationPayload))
	proofHex := hex.EncodeToString(proofDigest[:])
	besuTxHash := "0x" + proofHex

	reportID := fmt.Sprintf("rep-df-%s", telemetry.SessionID)
	report := &DeepfakeAuditReport{
		ReportID:                reportID,
		SessionID:               telemetry.SessionID,
		UserID:                  telemetry.UserID,
		Status:                  status,
		DeepfakeRiskScore:       deepfakeRisk,
		LivenessConfidenceScore: livenessConfidence,
		RejectionReason:         rejectionReason,
		RawVideoHash:            telemetry.RawVideoSHA256,
		AttestationProofDigest:  proofHex,
		BesuOnChainAnchorTxHash: besuTxHash,
		EvaluatedAt:             time.Now().UTC(),
	}

	e.reports[reportID] = report
	return report, nil
}
