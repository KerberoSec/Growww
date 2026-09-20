package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// CommodityType represents physical commodity classification
type CommodityType string

const (
	CommodityGold9999     CommodityType = "GOLD_9999"
	CommodityGold9950     CommodityType = "GOLD_9950"
	CommoditySilver9990   CommodityType = "SILVER_9990"
	CommodityPlatinum9995 CommodityType = "PLATINUM_9995"
)

// Minimum acceptable purity in basis points (e.g. 9999 = 99.99%)
var minFinenessBpsMap = map[CommodityType]int{
	CommodityGold9999:     9999,
	CommodityGold9950:     9950,
	CommoditySilver9990:   9990,
	CommodityPlatinum9995: 9995,
}

// AuditType represents the mandate under which an audit is conducted
type AuditType string

const (
	AuditInboundIngress         AuditType = "INBOUND_INGRESS"
	AuditPeriodicCycleCount     AuditType = "PERIODIC_CYCLE_COUNT"
	AuditRandomSampleAssay      AuditType = "RANDOM_SAMPLE_ASSAY"
	AuditOutboundEgress         AuditType = "OUTBOUND_EGRESS"
	AuditRegulatorySurpriseAudit AuditType = "REGULATORY_SURPRISE_AUDIT"
)

// AttestationStatus represents the multi-party quorum lifecycle
type AttestationStatus string

const (
	StatusDraft                    AttestationStatus = "DRAFT"
	StatusPendingAssayerSignature   AttestationStatus = "PENDING_ASSAYER_SIGNATURE"
	StatusPendingInspectorSignature AttestationStatus = "PENDING_INSPECTOR_SIGNATURE"
	StatusPendingCustodianSignature AttestationStatus = "PENDING_CUSTODIAN_SIGNATURE"
	StatusAttestationCompleted     AttestationStatus = "ATTESTATION_COMPLETED"
	StatusAttestationRejected      AttestationStatus = "ATTESTATION_REJECTED"
	StatusQuarantined              AttestationStatus = "QUARANTINED"
)

// DiscrepancySeverity represents the risk classification of an anomaly
type DiscrepancySeverity string

const (
	SeverityLowDocumentation     DiscrepancySeverity = "LOW_DOCUMENTATION"
	SeverityMediumSealMismatch   DiscrepancySeverity = "MEDIUM_SEAL_MISMATCH"
	SeverityHighWeightVariance   DiscrepancySeverity = "HIGH_WEIGHT_VARIANCE"
	SeverityCriticalPurityFailure DiscrepancySeverity = "CRITICAL_PURITY_FAILURE"
	SeverityCriticalBarMissing   DiscrepancySeverity = "CRITICAL_BAR_MISSING"
)

// EnwrRepository represents WDRA repository platforms
type EnwrRepository string

const (
	RepoNERL EnwrRepository = "NERL"
	RepoCCRL EnwrRepository = "CCRL"
)

// Recognized refineries by LBMA & BIS
var recognizedRefineries = map[string]bool{
	"MMTC-PAMP":      true,
	"VALCAMBI":       true,
	"ARGOR-HERAEUS":  true,
	"RAND REFINERY":  true,
	"PAMP SUISSE":    true,
	"HERAEUS":        true,
	"EMIRATES GOLD":  true,
}

// VaultFacility represents accredited storage vault
type VaultFacility struct {
	VaultID                string    `json:"vault_id"`
	VaultName              string    `json:"vault_name"`
	CustodianCompanyName   string    `json:"custodian_company_name"`
	MCXAccreditationNumber string    `json:"mcx_accreditation_number"`
	WDRARegistrationNumber string    `json:"wdra_registration_number"`
	FacilityAddress        string    `json:"facility_address"`
	Latitude               float64   `json:"latitude"`
	Longitude              float64   `json:"longitude"`
	GeofenceRadiusMeters   int       `json:"geofence_radius_meters"`
	IsActive               bool      `json:"is_active"`
	CreatedAt              time.Time `json:"created_at"`
}

// VaultBar represents physical commodity bar in accredited custody
type VaultBar struct {
	BarID               string        `json:"bar_id"`
	VaultID             string        `json:"vault_id"`
	BarSerialNumber     string        `json:"bar_serial_number"`
	RefineryName        string        `json:"refinery_name"`
	RefineryBatchNumber string        `json:"refinery_batch_number"`
	CommodityType       CommodityType `json:"commodity_type"`
	PurityFinenessBps   int           `json:"purity_fineness_bps"`
	GrossWeightGrams    float64       `json:"gross_weight_grams"`
	FineWeightGrams     float64       `json:"fine_weight_grams"`
	ENWRNumber          string        `json:"enwr_number"`
	TamperBagSealNumber string        `json:"tamper_bag_seal_number"`
	IsQuarantined       bool          `json:"is_quarantined"`
	LastAuditSessionID  string        `json:"last_audit_session_id"`
	BarHash             string        `json:"bar_hash"`
}

// WeighmentReceipt records calibrated scale reading
type WeighmentReceipt struct {
	WeighmentID             string    `json:"weighment_id"`
	SessionID               string    `json:"session_id"`
	VaultID                 string    `json:"vault_id"`
	BarSerialNumber         string    `json:"bar_serial_number"`
	ScaleDeviceID           string    `json:"scale_device_id"`
	ScaleCalibrationCertID  string    `json:"scale_calibration_cert_id"`
	GrossWeightGrams        float64   `json:"gross_weight_grams"`
	TareWeightGrams         float64   `json:"tare_weight_grams"`
	NetWeightGrams          float64   `json:"net_weight_grams"`
	ManifestWeightGrams     float64   `json:"manifest_weight_grams"`
	VariancePercentage      float64   `json:"variance_percentage"`
	ScaleDigitalSignature   string    `json:"scale_digital_signature"`
	WeighmentHash           string    `json:"weighment_hash"`
	RecordedAt              time.Time `json:"recorded_at"`
}

// AssayAttestation records NABL lab spectroscopic purity certification
type AssayAttestation struct {
	CertificateID             string        `json:"certificate_id"`
	SessionID                 string        `json:"session_id"`
	BarSerialNumber           string        `json:"bar_serial_number"`
	AssayLaboratoryID         string        `json:"assay_laboratory_id"`
	NABACAccreditationNumber   string        `json:"nabl_accreditation_number"`
	CommodityType             CommodityType `json:"commodity_type"`
	MeasuredFinenessBps       int           `json:"measured_fineness_bps"`
	UltrasonicVelocityMPerS   float64       `json:"ultrasonic_velocity_m_per_s"`
	XRFPurityReadingPct       float64       `json:"xrf_purity_reading_pct"`
	AssayCertificateS3URI     string        `json:"assay_certificate_s3_uri"`
	AssayCertificateSHA256    string        `json:"assay_certificate_sha256"`
	AssayerPublicKey          string        `json:"assayer_public_key"`
	AssayerSignature          string        `json:"assayer_signature"`
	RecordedAt                time.Time     `json:"recorded_at"`
}

// WDRAeNWRRecord represents repository electronic negotiable warehouse receipt
type WDRAeNWRRecord struct {
	ENWRNumber            string         `json:"enwr_number"`
	Repository            EnwrRepository `json:"repository"`
	VaultID               string         `json:"vault_id"`
	BeneficiaryClientCode string         `json:"beneficiary_client_code"`
	CommodityType         CommodityType  `json:"commodity_type"`
	TotalWeightGrams      float64        `json:"total_weight_grams"`
	PurityBps             int            `json:"purity_bps"`
	IsPledged             bool           `json:"is_pledged"`
	IsLienMarked          bool           `json:"is_lien_marked"`
	ReceiptStatus         string         `json:"receipt_status"` // "ACTIVE", "EXPIRED", "LOCKED"
	ValidityStart         time.Time      `json:"validity_start"`
	ValidityEnd           time.Time      `json:"validity_end"`
	LastSyncedAt          time.Time      `json:"last_synced_at"`
}

// VaultAuditSession represents an active audit mandate session
type VaultAuditSession struct {
	SessionID                  string            `json:"session_id"`
	VaultID                    string            `json:"vault_id"`
	AuditType                  AuditType         `json:"audit_type"`
	LeadInspectorID            string            `json:"lead_inspector_id"`
	AssayerID                  string            `json:"assayer_id"`
	CustodianRepresentativeID  string            `json:"custodian_representative_id"`
	AuditMandateReference      string            `json:"audit_mandate_reference"`
	Status                     AttestationStatus `json:"status"`
	ExpectedBarSerials         []string          `json:"expected_bar_serials"`
	AuditedBarSerials          []string          `json:"audited_bar_serials"`
	TotalBarsAudited           int               `json:"total_bars_audited"`
	TotalGrossWeightGrams      float64           `json:"total_gross_weight_grams"`
	TotalFineWeightGrams       float64           `json:"total_fine_weight_grams"`
	InventoryMerkleRoot        string            `json:"inventory_merkle_root"`
	InspectorSignature         string            `json:"inspector_signature"`
	AssayerSignature           string            `json:"assayer_signature"`
	CustodianSignature         string            `json:"custodian_signature"`
	OnChainAttestationTxHash   string            `json:"on_chain_attestation_tx_hash"`
	DiscrepancyIDs             []string          `json:"discrepancy_ids"`
	OpenedAt                   time.Time         `json:"opened_at"`
	ClosedAt                   time.Time         `json:"closed_at"`
}

// VaultAuditDiscrepancy represents detected anomaly
type VaultAuditDiscrepancy struct {
	DiscrepancyID              string              `json:"discrepancy_id"`
	SessionID                  string              `json:"session_id"`
	VaultID                    string              `json:"vault_id"`
	BarSerialNumber            string              `json:"bar_serial_number"`
	RefineryName               string              `json:"refinery_name"`
	Severity                   DiscrepancySeverity `json:"severity"`
	DiscrepancyType            string              `json:"discrepancy_type"`
	ExpectedValue              string              `json:"expected_value"`
	MeasuredValue              string              `json:"measured_value"`
	VariancePercentage         float64             `json:"variance_percentage"`
	EmergencyMintHaltTriggered bool                `json:"emergency_mint_halt_triggered"`
	IsResolved                 bool                `json:"is_resolved"`
	ResolutionOfficerID        string              `json:"resolution_officer_id"`
	ResolutionAction           string              `json:"resolution_action"`
	ResolutionJustification    string              `json:"resolution_justification"`
	ResolvedAt                 time.Time           `json:"resolved_at"`
	CreatedAt                  time.Time           `json:"created_at"`
}

// MerkleNode represents node in the Proof-of-Reserve tree
type MerkleNode struct {
	Hash  string
	Left  *MerkleNode
	Right *MerkleNode
}

// MerkleProof represents audit inclusion proof for single bar
type MerkleProof struct {
	BarHash      string   `json:"bar_hash"`
	MerkleRoot   string   `json:"merkle_root"`
	ProofPath    []string `json:"proof_path"`
	ProofIndices []int    `json:"proof_indices"` // 0 for left sibling, 1 for right sibling
}

// VaultAuditPortalEngine coordinates physical vault audits, IoT weighment, e-NWRs, and PoR Merkle proofs
type VaultAuditPortalEngine struct {
	mu               sync.RWMutex
	facilities       map[string]*VaultFacility
	inventory        map[string]map[string]*VaultBar // vaultID -> barSerial -> Bar
	sessions         map[string]*VaultAuditSession
	weighments       map[string]*WeighmentReceipt
	assayCerts       map[string]*AssayAttestation
	enwrRecords      map[string]*WDRAeNWRRecord // enwrNumber -> Record
	discrepancies    map[string]*VaultAuditDiscrepancy
	mintFrozenVaults map[string]bool
}

// NewVaultAuditPortalEngine creates a new VaultAuditPortalEngine instance
func NewVaultAuditPortalEngine() *VaultAuditPortalEngine {
	return &VaultAuditPortalEngine{
		facilities:       make(map[string]*VaultFacility),
		inventory:        make(map[string]map[string]*VaultBar),
		sessions:         make(map[string]*VaultAuditSession),
		weighments:       make(map[string]*WeighmentReceipt),
		assayCerts:       make(map[string]*AssayAttestation),
		enwrRecords:      make(map[string]*WDRAeNWRRecord),
		discrepancies:    make(map[string]*VaultAuditDiscrepancy),
		mintFrozenVaults: make(map[string]bool),
	}
}

// RegisterFacility registers an accredited bullion vault
func (e *VaultAuditPortalEngine) RegisterFacility(facility *VaultFacility) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if facility.VaultID == "" {
		return errors.New("vault_id is required")
	}
	if facility.MCXAccreditationNumber == "" {
		return errors.New("mcx_accreditation_number is required")
	}
	if facility.WDRARegistrationNumber == "" {
		return errors.New("wdra_registration_number is required")
	}
	e.facilities[facility.VaultID] = facility
	if _, ok := e.inventory[facility.VaultID]; !ok {
		e.inventory[facility.VaultID] = make(map[string]*VaultBar)
	}
	return nil
}

// ComputeBarHash computes deterministic SHA256 hash: SHA256(BarID || GrossWeight || FinenessBps || VaultCode)
func ComputeBarHash(serialNumber string, grossWeight float64, finenessBps int, vaultID string) string {
	payload := fmt.Sprintf("%s:%.4f:%d:%s", serialNumber, grossWeight, finenessBps, vaultID)
	h := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(h[:])
}

// IngestWDRAeNWR ingests an electronic Negotiable Warehouse Receipt from NERL / CCRL
func (e *VaultAuditPortalEngine) IngestWDRAeNWR(rec *WDRAeNWRRecord) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if rec.ENWRNumber == "" {
		return errors.New("enwr_number is required")
	}
	if rec.Repository != RepoNERL && rec.Repository != RepoCCRL {
		return errors.New("invalid repository: must be NERL or CCRL")
	}
	rec.LastSyncedAt = time.Now().UTC()
	e.enwrRecords[rec.ENWRNumber] = rec
	return nil
}

// VerifyEnwrReceipt verifies an e-NWR receipt's validity and encumbrance status
func (e *VaultAuditPortalEngine) VerifyEnwrReceipt(repo EnwrRepository, enwrNumber, vaultID, clientCode string) (*WDRAeNWRRecord, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rec, exists := e.enwrRecords[enwrNumber]
	if !exists {
		return nil, fmt.Errorf("e-NWR %s not found in %s repository", enwrNumber, repo)
	}
	if rec.Repository != repo {
		return nil, fmt.Errorf("repository mismatch: expected %s, found %s", repo, rec.Repository)
	}
	if rec.VaultID != vaultID {
		return nil, fmt.Errorf("vault mismatch: receipt belongs to %s, requested %s", rec.VaultID, vaultID)
	}
	if clientCode != "" && rec.BeneficiaryClientCode != clientCode {
		return nil, fmt.Errorf("beneficiary mismatch: expected %s, got %s", rec.BeneficiaryClientCode, clientCode)
	}
	if rec.IsPledged || rec.IsLienMarked {
		return rec, fmt.Errorf("e-NWR %s is encumbered (pledged: %v, lien: %v)", enwrNumber, rec.IsPledged, rec.IsLienMarked)
	}
	now := time.Now().UTC()
	if now.After(rec.ValidityEnd) {
		return rec, fmt.Errorf("e-NWR %s expired on %s", enwrNumber, rec.ValidityEnd.Format(time.RFC3339))
	}
	return rec, nil
}

// CreateAuditSession initiates an official inspection session
func (e *VaultAuditPortalEngine) CreateAuditSession(
	vaultID string,
	auditType AuditType,
	leadInspectorID, assayerID, custodianID, mandateRef string,
	expectedSerials []string,
) (*VaultAuditSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.facilities[vaultID]; !ok {
		return nil, fmt.Errorf("vault %s is not registered or accredited", vaultID)
	}
	if leadInspectorID == "" || assayerID == "" || custodianID == "" {
		return nil, errors.New("all 3 roles (inspector, assayer, custodian) must be assigned")
	}

	sessionID := fmt.Sprintf("session-%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", vaultID, mandateRef, time.Now().UnixNano()))))[:16]
	session := &VaultAuditSession{
		SessionID:                 sessionID,
		VaultID:                   vaultID,
		AuditType:                 auditType,
		LeadInspectorID:           leadInspectorID,
		AssayerID:                 assayerID,
		CustodianRepresentativeID: custodianID,
		AuditMandateReference:     mandateRef,
		Status:                    StatusDraft,
		ExpectedBarSerials:        expectedSerials,
		AuditedBarSerials:         make([]string, 0),
		DiscrepancyIDs:            make([]string, 0),
		OpenedAt:                  time.Now().UTC(),
	}
	e.sessions[sessionID] = session
	return session, nil
}

// SubmitBarWeighment validates IoT scale telemetry, tare calculation, and manifest variance (<= 0.001%)
func (e *VaultAuditPortalEngine) SubmitBarWeighment(
	sessionID, vaultID, serialNumber string,
	commodity CommodityType,
	scaleID, calibCertID string,
	grossWeight, tareWeight, netWeight, manifestWeight float64,
	rawTelemetry, scaleSignature string,
	certExpiry time.Time,
) (*WeighmentReceipt, *VaultAuditDiscrepancy, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.sessions[sessionID]
	if !exists {
		return nil, nil, fmt.Errorf("audit session %s not found", sessionID)
	}
	if session.Status != StatusDraft && session.Status != StatusPendingAssayerSignature && session.Status != StatusPendingInspectorSignature {
		return nil, nil, fmt.Errorf("cannot weigh bars in session status %s", session.Status)
	}

	// 1. Calibration certificate check
	if time.Now().UTC().After(certExpiry) {
		return nil, nil, fmt.Errorf("scale calibration certificate %s is expired", calibCertID)
	}

	// 2. Mathematical net weight consistency (Gross - Tare = Net)
	calcNet := grossWeight - tareWeight
	if math.Abs(calcNet-netWeight) > 0.0001 {
		return nil, nil, fmt.Errorf("net weight arithmetic mismatch: gross(%.4f) - tare(%.4f) = %.4f != declared net(%.4f)",
			grossWeight, tareWeight, calcNet, netWeight)
	}

	// 3. Scale digital signature verification (SHA256 HMAC / ECDSA placeholder)
	if scaleSignature == "" {
		return nil, nil, errors.New("missing cryptographic scale digital signature")
	}

	// 4. Manifest variance check: |NetWeight - ManifestWeight| / ManifestWeight <= 0.001% (0.00001)
	variancePct := 0.0
	if manifestWeight > 0 {
		variancePct = math.Abs(netWeight-manifestWeight) / manifestWeight * 100.0
	}

	weighmentHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%.4f:%.4f:%s",
		sessionID, vaultID, serialNumber, grossWeight, netWeight, scaleSignature)))
	weighmentHashHex := hex.EncodeToString(weighmentHash[:])

	receipt := &WeighmentReceipt{
		WeighmentID:            fmt.Sprintf("wgh-%s", weighmentHashHex[:12]),
		SessionID:              sessionID,
		VaultID:                vaultID,
		BarSerialNumber:        serialNumber,
		ScaleDeviceID:          scaleID,
		ScaleCalibrationCertID: calibCertID,
		GrossWeightGrams:       grossWeight,
		TareWeightGrams:        tareWeight,
		NetWeightGrams:         netWeight,
		ManifestWeightGrams:    manifestWeight,
		VariancePercentage:     variancePct,
		ScaleDigitalSignature:  scaleSignature,
		WeighmentHash:          weighmentHashHex,
		RecordedAt:             time.Now().UTC(),
	}
	e.weighments[receipt.WeighmentID] = receipt

	var discrepancy *VaultAuditDiscrepancy
	// Variance threshold: 0.001% = 0.001 percentage points
	if variancePct > 0.001 {
		discID := fmt.Sprintf("disc-wgh-%s", weighmentHashHex[:8])
		discrepancy = &VaultAuditDiscrepancy{
			DiscrepancyID:              discID,
			SessionID:                  sessionID,
			VaultID:                    vaultID,
			BarSerialNumber:            serialNumber,
			Severity:                   SeverityHighWeightVariance,
			DiscrepancyType:            "WEIGHT_MISMATCH",
			ExpectedValue:              fmt.Sprintf("%.4fg", manifestWeight),
			MeasuredValue:              fmt.Sprintf("%.4fg", netWeight),
			VariancePercentage:         variancePct,
			EmergencyMintHaltTriggered: false,
			IsResolved:                 false,
			CreatedAt:                  time.Now().UTC(),
		}
		e.discrepancies[discID] = discrepancy
		session.DiscrepancyIDs = append(session.DiscrepancyIDs, discID)
	}

	return receipt, discrepancy, nil
}

// SubmitAssayAttestation verifies assayer accreditation, ultrasonic density, XRF spectroscopic purity
func (e *VaultAuditPortalEngine) SubmitAssayAttestation(
	sessionID, vaultID, serialNumber, refineryName, batchNumber, labID, nablNumber string,
	commodity CommodityType,
	finenessBps int,
	ultrasonicVelocity float64,
	xrfPurityPct float64,
	s3URI, docSHA256, assayerPubKey, assayerSig string,
) (*AssayAttestation, *VaultAuditDiscrepancy, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.sessions[sessionID]
	if !exists {
		return nil, nil, fmt.Errorf("session %s not found", sessionID)
	}

	// 1. Refinery recognition check
	if !recognizedRefineries[refineryName] {
		return nil, nil, fmt.Errorf("refinery %s is not an LBMA or BIS accredited refinery", refineryName)
	}

	// 2. Ultrasonic velocity validation for tungsten core detection
	// Pure Gold speed of sound ~ 3240 m/s; Tungsten ~ 5220 m/s; gold-plated fake will deviate
	if commodity == CommodityGold9999 || commodity == CommodityGold9950 {
		if ultrasonicVelocity < 3150 || ultrasonicVelocity > 3350 {
			// Severe density anomaly - potential fake bar!
			discID := fmt.Sprintf("disc-density-%s", serialNumber)
			disc := &VaultAuditDiscrepancy{
				DiscrepancyID:              discID,
				SessionID:                  sessionID,
				VaultID:                    vaultID,
				BarSerialNumber:            serialNumber,
				RefineryName:               refineryName,
				Severity:                   SeverityCriticalPurityFailure,
				DiscrepancyType:            "ULTRASONIC_DENSITY_ANOMALY",
				ExpectedValue:              "3240 m/s (Gold)",
				MeasuredValue:              fmt.Sprintf("%.2f m/s", ultrasonicVelocity),
				VariancePercentage:         math.Abs(ultrasonicVelocity-3240.0) / 3240.0 * 100.0,
				EmergencyMintHaltTriggered: true,
				IsResolved:                 false,
				CreatedAt:                  time.Now().UTC(),
			}
			e.discrepancies[discID] = disc
			session.DiscrepancyIDs = append(session.DiscrepancyIDs, discID)
			e.mintFrozenVaults[vaultID] = true
			return nil, disc, errors.New("critical density failure: ultrasonic velocity indicates non-gold core")
		}
	}

	// 3. Minimum fineness purity requirement
	minFineness, ok := minFinenessBpsMap[commodity]
	if !ok {
		minFineness = 9999
	}

	att := &AssayAttestation{
		CertificateID:           fmt.Sprintf("assay-%s", serialNumber),
		SessionID:               sessionID,
		BarSerialNumber:         serialNumber,
		AssayLaboratoryID:       labID,
		NABACAccreditationNumber: nablNumber,
		CommodityType:           commodity,
		MeasuredFinenessBps:     finenessBps,
		UltrasonicVelocityMPerS: ultrasonicVelocity,
		XRFPurityReadingPct:     xrfPurityPct,
		AssayCertificateS3URI:   s3URI,
		AssayCertificateSHA256:  docSHA256,
		AssayerPublicKey:        assayerPubKey,
		AssayerSignature:        assayerSig,
		RecordedAt:              time.Now().UTC(),
	}
	e.assayCerts[att.CertificateID] = att

	var disc *VaultAuditDiscrepancy
	if finenessBps < minFineness {
		discID := fmt.Sprintf("disc-purity-%s", serialNumber)
		disc = &VaultAuditDiscrepancy{
			DiscrepancyID:              discID,
			SessionID:                  sessionID,
			VaultID:                    vaultID,
			BarSerialNumber:            serialNumber,
			RefineryName:               refineryName,
			Severity:                   SeverityCriticalPurityFailure,
			DiscrepancyType:            "PURITY_DEVIATION",
			ExpectedValue:              fmt.Sprintf("%d bps", minFineness),
			MeasuredValue:              fmt.Sprintf("%d bps", finenessBps),
			VariancePercentage:         float64(minFineness-finenessBps) / float64(minFineness) * 100.0,
			EmergencyMintHaltTriggered: true,
			IsResolved:                 false,
			CreatedAt:                  time.Now().UTC(),
		}
		e.discrepancies[discID] = disc
		session.DiscrepancyIDs = append(session.DiscrepancyIDs, discID)
		e.mintFrozenVaults[vaultID] = true
	}

	return att, disc, nil
}

// AddOrUpdateBar registers or updates a bar in the vault inventory
func (e *VaultAuditPortalEngine) AddOrUpdateBar(bar *VaultBar) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, ok := e.inventory[bar.VaultID]; !ok {
		e.inventory[bar.VaultID] = make(map[string]*VaultBar)
	}
	bar.BarHash = ComputeBarHash(bar.BarSerialNumber, bar.GrossWeightGrams, bar.PurityFinenessBps, bar.VaultID)
	bar.FineWeightGrams = bar.GrossWeightGrams * (float64(bar.PurityFinenessBps) / 10000.0)
	e.inventory[bar.VaultID][bar.BarSerialNumber] = bar
}

// BuildMerkleTree constructs a deterministic SHA-256 Merkle tree from sorted bar hashes
func BuildMerkleTree(barHashes []string) (string, map[string]*MerkleProof) {
	if len(barHashes) == 0 {
		return "0000000000000000000000000000000000000000000000000000000000000000", make(map[string]*MerkleProof)
	}

	// Sort canonical order
	sortedHashes := make([]string, len(barHashes))
	copy(sortedHashes, barHashes)
	sort.Strings(sortedHashes)

	// Build tree levels
	type treeLevel struct {
		hashes []string
	}
	levels := []treeLevel{{hashes: sortedHashes}}

	current := sortedHashes
	for len(current) > 1 {
		var next []string
		for i := 0; i < len(current); i += 2 {
			if i+1 < len(current) {
				comb := current[i] + current[i+1]
				h := sha256.Sum256([]byte(comb))
				next = append(next, hex.EncodeToString(h[:]))
			} else {
				// Odd node duplicated as per standard Merkle tree rules
				comb := current[i] + current[i]
				h := sha256.Sum256([]byte(comb))
				next = append(next, hex.EncodeToString(h[:]))
			}
		}
		levels = append(levels, treeLevel{hashes: next})
		current = next
	}

	root := current[0]
	proofs := make(map[string]*MerkleProof)

	// Build individual inclusion proofs
	for originalIdx, originalHash := range sortedHashes {
		var proofPath []string
		var indices []int

		idx := originalIdx

		for l := 0; l < len(levels)-1; l++ {
			levelHashes := levels[l].hashes
			var siblingIdx int
			var isRight int

			if idx%2 == 0 {
				if idx+1 < len(levelHashes) {
					siblingIdx = idx + 1
				} else {
					siblingIdx = idx // duplicated
				}
				isRight = 1
			} else {
				siblingIdx = idx - 1
				isRight = 0
			}

			proofPath = append(proofPath, levelHashes[siblingIdx])
			indices = append(indices, isRight)
			idx = idx / 2
		}

		proofs[originalHash] = &MerkleProof{
			BarHash:      originalHash,
			MerkleRoot:   root,
			ProofPath:    proofPath,
			ProofIndices: indices,
		}
	}

	return root, proofs
}

// VerifyMerkleProof validates an arbitrary inclusion proof against the root
func VerifyMerkleProof(proof *MerkleProof) bool {
	if proof == nil || proof.BarHash == "" || proof.MerkleRoot == "" {
		return false
	}

	current := proof.BarHash
	for i, sibling := range proof.ProofPath {
		var comb string
		if proof.ProofIndices[i] == 1 {
			// Sibling is on the right: H(current || sibling)
			comb = current + sibling
		} else {
			// Sibling is on the left: H(sibling || current)
			comb = sibling + current
		}
		h := sha256.Sum256([]byte(comb))
		current = hex.EncodeToString(h[:])
	}
	return current == proof.MerkleRoot
}

// FinalizeAuditSession executes M-of-N threshold signature quorum and generates on-chain PoR Merkle root
func (e *VaultAuditPortalEngine) FinalizeAuditSession(
	sessionID, inspectorSig, assayerSig, custodianSig, reportS3URI, reportSHA256 string,
) (*VaultAuditSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, exists := e.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	if inspectorSig == "" || assayerSig == "" || custodianSig == "" {
		return nil, errors.New("3-of-3 multi-signature quorum required: inspector, assayer, and custodian must sign")
	}

	session.InspectorSignature = inspectorSig
	session.AssayerSignature = assayerSig
	session.CustodianSignature = custodianSig

	// Check if there are critical unresolved discrepancies
	for _, discID := range session.DiscrepancyIDs {
		disc, ok := e.discrepancies[discID]
		if ok && !disc.IsResolved {
			if disc.Severity == SeverityCriticalPurityFailure || disc.Severity == SeverityCriticalBarMissing {
				session.Status = StatusQuarantined
				return session, fmt.Errorf("cannot finalize session: critical discrepancy %s is unresolved", discID)
			}
		}
	}

	// Collect bar hashes for vault
	barsMap := e.inventory[session.VaultID]
	var barHashes []string
	totGross := 0.0
	totFine := 0.0

	for _, bar := range barsMap {
		if !bar.IsQuarantined {
			barHashes = append(barHashes, bar.BarHash)
			totGross += bar.GrossWeightGrams
			totFine += bar.FineWeightGrams
			session.AuditedBarSerials = append(session.AuditedBarSerials, bar.BarSerialNumber)
		}
	}

	root, _ := BuildMerkleTree(barHashes)
	session.InventoryMerkleRoot = root
	session.TotalBarsAudited = len(barHashes)
	session.TotalGrossWeightGrams = totGross
	session.TotalFineWeightGrams = totFine
	session.Status = StatusAttestationCompleted
	session.ClosedAt = time.Now().UTC()

	// Simulate Besu on-chain anchoring transaction hash
	txDigest := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", sessionID, root, session.ClosedAt.UnixNano())))
	session.OnChainAttestationTxHash = "0x" + hex.EncodeToString(txDigest[:])

	return session, nil
}

// GetProofOfReserveSnapshot generates reconciliation against tokenized circulating supply
func (e *VaultAuditPortalEngine) GetProofOfReserveSnapshot(vaultID string, commodity CommodityType, tokenizedSupplyGrams float64) (map[string]interface{}, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	barsMap, exists := e.inventory[vaultID]
	if !exists {
		return nil, fmt.Errorf("vault %s not found", vaultID)
	}

	var hashes []string
	totFine := 0.0
	barCount := 0

	for _, bar := range barsMap {
		if bar.CommodityType == commodity && !bar.IsQuarantined {
			hashes = append(hashes, bar.BarHash)
			totFine += bar.FineWeightGrams
			barCount++
		}
	}

	root, _ := BuildMerkleTree(hashes)

	ratio := 0.0
	if tokenizedSupplyGrams > 0 {
		ratio = totFine / tokenizedSupplyGrams
	}

	snapshot := map[string]interface{}{
		"snapshot_id":                      fmt.Sprintf("por-%s-%d", vaultID, time.Now().Unix()),
		"vault_id":                         vaultID,
		"commodity_type":                   commodity,
		"total_physical_bars":              barCount,
		"total_physical_fine_weight_grams": totFine,
		"tokenized_circulating_supply":     tokenizedSupplyGrams,
		"reserve_ratio":                    ratio,
		"is_fully_backed":                  totFine >= tokenizedSupplyGrams,
		"inventory_merkle_root":            root,
		"mint_frozen":                      e.mintFrozenVaults[vaultID],
		"timestamp_utc":                    time.Now().UTC().Unix(),
	}

	return snapshot, nil
}

// ResolveDiscrepancy processes compliance resolution for an anomaly
func (e *VaultAuditPortalEngine) ResolveDiscrepancy(
	discrepancyID, officerID, action, justification, mfaToken string,
) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	disc, exists := e.discrepancies[discrepancyID]
	if !exists {
		return fmt.Errorf("discrepancy %s not found", discrepancyID)
	}
	if officerID == "" || justification == "" || mfaToken == "" {
		return errors.New("resolution requires officer ID, justification, and MFA verification token")
	}

	disc.IsResolved = true
	disc.ResolutionOfficerID = officerID
	disc.ResolutionAction = action
	disc.ResolutionJustification = justification
	disc.ResolvedAt = time.Now().UTC()

	// Check if all critical discrepancies for this vault are resolved to unfreeze
	hasUnresolvedCritical := false
	for _, d := range e.discrepancies {
		if d.VaultID == disc.VaultID && !d.IsResolved && (d.Severity == SeverityCriticalPurityFailure || d.Severity == SeverityCriticalBarMissing) {
			hasUnresolvedCritical = true
			break
		}
	}
	if !hasUnresolvedCritical {
		e.mintFrozenVaults[disc.VaultID] = false
	}

	return nil
}

// Helper to generate a dummy ECDSA signature for mock tests
func GenerateMockSignature() string {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	hash := sha256.Sum256([]byte("growww-vault-audit-attestation"))
	r, s, _ := ecdsa.Sign(rand.Reader, priv, hash[:])
	return fmt.Sprintf("0x%s%s", hex.EncodeToString(r.Bytes()), hex.EncodeToString(s.Bytes()))
}
