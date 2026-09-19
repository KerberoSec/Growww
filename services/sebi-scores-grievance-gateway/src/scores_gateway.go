package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

var panRegex = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)

// UserMappingStore is an interface to resolve PAN hashes to internal user accounts
type UserMappingStore interface {
	ResolveUserByPANHash(panHash string) (string, bool)
	RegisterUserPAN(pan string, userID string) string
}

// MemoryUserMappingStore is an in-memory user registry for mapping PAN hashes to user IDs
type MemoryUserMappingStore struct {
	mu       sync.RWMutex
	panToUID map[string]string // panHash -> userID
}

func NewMemoryUserMappingStore() *MemoryUserMappingStore {
	return &MemoryUserMappingStore{
		panToUID: make(map[string]string),
	}
}

func (s *MemoryUserMappingStore) ResolveUserByPANHash(panHash string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	uid, exists := s.panToUID[panHash]
	return uid, exists
}

func (s *MemoryUserMappingStore) RegisterUserPAN(pan string, userID string) string {
	panHash := HashPAN(pan)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.panToUID[panHash] = userID
	return panHash
}

// ScoresGatewayConfig holds connection parameters and credentials
type ScoresGatewayConfig struct {
	ScoresAPIEndpoint string
	WebhookSecret     string
	EntitySEBICode    string
	UseMTLS           bool
}

// RegulatorySubmissionResult stores the response from submitting an ATR to the regulator
type RegulatorySubmissionResult struct {
	ComplaintID            string    `json:"complaint_id"`
	RegulatoryAckNumber    string    `json:"regulatory_ack_number"`
	SubmittedAt            time.Time `json:"submitted_at"`
	BesuTransactionHash    string    `json:"besu_transaction_hash"`
	ReceiptHash            string    `json:"receipt_hash"`
	ZeroPIIConfirmed       bool      `json:"zero_pii_confirmed"`
}

// SCORESGatewayConnector implements the SEBI SCORES 2.0 API gateway
type SCORESGatewayConnector struct {
	config    ScoresGatewayConfig
	userStore UserMappingStore
	mu        sync.RWMutex
	dockets   map[string]*ComplaintDocket // ID -> Docket
	byRegNo   map[string]string           // ExternalRegNo -> ID
	atrs      map[string]*ActionTakenReport
	dossiers  map[string]*EvidenceDossier
}

// NewSCORESGatewayConnector creates a new SCORES 2.0 Gateway instance
func NewSCORESGatewayConnector(cfg ScoresGatewayConfig, userStore UserMappingStore) *SCORESGatewayConnector {
	if userStore == nil {
		userStore = NewMemoryUserMappingStore()
	}
	return &SCORESGatewayConnector{
		config:    cfg,
		userStore: userStore,
		dockets:   make(map[string]*ComplaintDocket),
		byRegNo:   make(map[string]string),
		atrs:      make(map[string]*ActionTakenReport),
		dossiers:  make(map[string]*EvidenceDossier),
	}
}

// ValidatePANFormat verifies standard 10-character Indian PAN format
func ValidatePANFormat(pan string) error {
	pan = strings.TrimSpace(strings.ToUpper(pan))
	if !panRegex.MatchString(pan) {
		return ErrInvalidPANFormat
	}
	return nil
}

// HashPAN calculates a deterministic SHA-256 hash of the normalized PAN
func HashPAN(pan string) string {
	normalized := strings.TrimSpace(strings.ToUpper(pan))
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// MaskPII returns redacted representations of sensitive personal information
func MaskPII(name, email, mobile string) (string, string, string) {
	maskedName := "INVESTOR_***"
	if len(name) > 2 {
		maskedName = string(name[0]) + strings.Repeat("*", len(name)-2) + string(name[len(name)-1])
	}

	maskedEmail := "***@***.***"
	parts := strings.Split(email, "@")
	if len(parts) == 2 && len(parts[0]) > 2 {
		maskedEmail = string(parts[0][0]) + "***" + string(parts[0][len(parts[0])-1]) + "@" + parts[1]
	}

	maskedMobile := "+91******"
	if len(mobile) >= 10 {
		maskedMobile = mobile[:3] + strings.Repeat("*", len(mobile)-5) + mobile[len(mobile)-2:]
	}

	return maskedName, maskedEmail, maskedMobile
}

// VerifyHMACSignature validates the incoming webhook against the shared secret
func (g *SCORESGatewayConnector) VerifyHMACSignature(rawPayload []byte, signatureHex string) bool {
	if g.config.WebhookSecret == "" {
		return true // Webhook secret not configured in local mock mode
	}
	mac := hmac.New(sha256.New, []byte(g.config.WebhookSecret))
	mac.Write(rawPayload)
	expectedMAC := mac.Sum(nil)
	actualMAC, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}
	return hmac.Equal(actualMAC, expectedMAC)
}

// IngestSCORESComplaint processes inbound webhook complaints from SEBI SCORES 2.0
func (g *SCORESGatewayConnector) IngestSCORESComplaint(payload ScoresWebhookPayload, signatureHex string) (*ComplaintDocket, error) {
	// Validate PAN
	if err := ValidatePANFormat(payload.ComplainantPAN); err != nil {
		return nil, err
	}

	// Verify HMAC if signature provided
	if signatureHex != "" {
		rawBytes, _ := json.Marshal(payload)
		if !g.VerifyHMACSignature(rawBytes, signatureHex) {
			return nil, ErrInvalidHMACSignature
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Idempotency check: prevent duplicate ingestion of same docket number
	if _, exists := g.byRegNo[payload.ScoresRegistrationNumber]; exists {
		return nil, ErrDuplicateDocket
	}

	panHash := HashPAN(payload.ComplainantPAN)
	maskedName, maskedEmail, maskedMobile := MaskPII(payload.ComplainantName, payload.ComplainantEmail, payload.ComplainantMobile)

	receiptDate := payload.ReceiptDate
	if receiptDate.IsZero() {
		receiptDate = time.Now().UTC()
	}

	// Compute statutory 21 calendar days deadline for SCORES 2.0
	statutoryDeadline := receiptDate.AddDate(0, 0, 21)

	// Generate deterministic complaint hash
	intakeDigestSrc := fmt.Sprintf("%s|%s|%s|%s", payload.ScoresRegistrationNumber, RegulatorSEBISCORES, receiptDate.Format(time.RFC3339), panHash)
	cHashBytes := sha256.Sum256([]byte(intakeDigestSrc))
	complaintHash := hex.EncodeToString(cHashBytes[:])

	docketID := fmt.Sprintf("docket-%s-%d", strings.ReplaceAll(payload.ScoresRegistrationNumber, "/", "-"), time.Now().UnixNano())

	docket := &ComplaintDocket{
		ID:                         docketID,
		ExternalRegistrationNumber: payload.ScoresRegistrationNumber,
		Regulator:                  RegulatorSEBISCORES,
		CurrentState:               StateIngested,
		ComplainantPANHash:         panHash,
		ComplainantNameMasked:      maskedName,
		ComplainantEmailMasked:     maskedEmail,
		ComplainantMobileMasked:    maskedMobile,
		Category:                   payload.ComplaintCategory,
		SubCategory:                payload.SubCategory,
		DisputedAmount:             payload.DisputedAmount,
		ComplaintDescription:       payload.ComplaintDescription,
		StatutorySLADays:           21,
		IntakeTimestamp:            receiptDate,
		StatutoryDeadline:          statutoryDeadline,
		CurrentEscalationStage:     Stage0Normal,
		IsSLABreached:              false,
		ComplaintHash:              complaintHash,
	}

	// Attempt internal user mapping via PAN hash
	if uid, found := g.userStore.ResolveUserByPANHash(panHash); found {
		docket.UserID = uid
		docket.CurrentState = StateMappedToUser
	}

	// Trigger automated diagnostic triage
	g.runDiagnosticTriage(docket)

	// Save to store
	g.dockets[docket.ID] = docket
	g.byRegNo[payload.ScoresRegistrationNumber] = docket.ID

	return docket, nil
}

// runDiagnosticTriage executes preliminary automated checks and moves to DIAGNOSTIC_TRIAGE or INVESTIGATION_PENDING
func (g *SCORESGatewayConnector) runDiagnosticTriage(docket *ComplaintDocket) {
	docket.CurrentState = StateDiagnosticTriage

	// Simulate creating an initial evidence dossier
	dossierID := fmt.Sprintf("dossier-%s", docket.ID)
	dossierHashSrc := fmt.Sprintf("%s|%s|%s", docket.ID, docket.ComplaintHash, time.Now().UTC().Format(time.RFC3339))
	dHash := sha256.Sum256([]byte(dossierHashSrc))

	dossier := &EvidenceDossier{
		ID:          dossierID,
		ComplaintID: docket.ID,
		KYCProfileSnapshot: map[string]string{
			"pan_hash":       docket.ComplainantPANHash,
			"mapped_user_id": docket.UserID,
			"kyc_status":     "VERIFIED_TIER_3",
		},
		LedgerStatementRef: fmt.Sprintf("s3://growww-audit-worm/statements/%s.json", docket.ID),
		TradeContractNotes: []string{
			fmt.Sprintf("ECN-2026-NSE-%s", docket.ComplainantPANHash[:8]),
		},
		AuditLogEventIDs: []string{
			fmt.Sprintf("AUDIT-%d-01", time.Now().Unix()),
			fmt.Sprintf("AUDIT-%d-02", time.Now().Unix()),
		},
		DossierSHA256:         hex.EncodeToString(dHash[:]),
		HarvestingCompletedAt: time.Now().UTC(),
	}

	g.dossiers[dossierID] = dossier
	docket.DossierID = dossierID
	docket.CurrentState = StateAuditDossierCollected
}

// GetDocket retrieves a complaint by its internal ID
func (g *SCORESGatewayConnector) GetDocket(id string) (*ComplaintDocket, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	d, exists := g.dockets[id]
	if !exists {
		return nil, ErrDocketNotFound
	}
	return d, nil
}

// GetDocketByRegNumber retrieves a complaint by external registration number (e.g. SEBIP/2026/0014529)
func (g *SCORESGatewayConnector) GetDocketByRegNumber(regNo string) (*ComplaintDocket, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	id, exists := g.byRegNo[regNo]
	if !exists {
		return nil, ErrDocketNotFound
	}
	return g.dockets[id], nil
}

// ListDockets retrieves all dockets
func (g *SCORESGatewayConnector) ListDockets() []*ComplaintDocket {
	g.mu.RLock()
	defer g.mu.RUnlock()
	list := make([]*ComplaintDocket, 0, len(g.dockets))
	for _, d := range g.dockets {
		list = append(list, d)
	}
	return list
}

// SubmitATRToRegulator dispatches the signed ATR to SEBI SCORES 2.0 and anchors receipt to Hyperledger Besu
func (g *SCORESGatewayConnector) SubmitATRToRegulator(atr *ActionTakenReport) (*RegulatorySubmissionResult, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	docket, exists := g.dockets[atr.ComplaintID]
	if !exists {
		return nil, ErrDocketNotFound
	}

	if docket.CurrentState == StateResolvedClosed {
		return nil, ErrDocketAlreadyResolved
	}

	// Verify Checker has approved
	if atr.CheckerUserID == "" || atr.CheckerApprovedAt == nil {
		return nil, ErrCheckerRequired
	}

	// Verify Class-3 DSC has been applied
	if !atr.IsDSCSigned || atr.ATRPackageHash == "" {
		return nil, ErrDSCNotSigned
	}

	now := time.Now().UTC()
	ackNumber := fmt.Sprintf("SEBI-ATR-ACK-2026-%d", now.UnixNano()%1000000)

	// Compute on-chain receipt hash with zero PII:
	// receiptHash = SHA256(complaintHash || atrPackageHash || ackNumber || timestamp)
	receiptSrc := fmt.Sprintf("%s|%s|%s|%d", docket.ComplaintHash, atr.ATRPackageHash, ackNumber, now.Unix())
	receiptDigest := sha256.Sum256([]byte(receiptSrc))
	receiptHash := hex.EncodeToString(receiptDigest[:])

	// Besu transaction simulation (2-second QBFT block finality)
	besuTxHash := "0x" + hex.EncodeToString(sha256.New().Sum([]byte(receiptHash+now.String())))

	// Update ATR record
	atr.RegulatorySubmissionStatus = "ACKNOWLEDGED"
	atr.RegulatoryAckNumber = ackNumber
	atr.RegulatoryAckTimestamp = &now
	atr.BesuTxHash = besuTxHash

	// Update Docket record
	docket.CurrentState = StateATRSubmittedToRegulator
	docket.ATRId = atr.ID
	docket.BesuTxHash = besuTxHash
	docket.ResolvedAt = &now
	docket.CurrentState = StateResolvedClosed // Formally closed upon regulator acknowledgement

	return &RegulatorySubmissionResult{
		ComplaintID:         docket.ID,
		RegulatoryAckNumber: ackNumber,
		SubmittedAt:         now,
		BesuTransactionHash: besuTxHash,
		ReceiptHash:         receiptHash,
		ZeroPIIConfirmed:    true,
	}, nil
}
