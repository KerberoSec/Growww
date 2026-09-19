package main

import (
	"errors"
	"time"
)

// RegulatorType represents supported regulatory redressal portals
type RegulatorType string

const (
	RegulatorSEBISCORES RegulatorType = "SEBI_SCORES"
	RegulatorRBICMS     RegulatorType = "RBI_CMS"
	RegulatorSMARTODR   RegulatorType = "SMART_ODR"
)

// GrievanceState represents the Finite State Machine lifecycle of a regulatory complaint
type GrievanceState string

const (
	StateIngested                 GrievanceState = "INGESTED"
	StateMappedToUser             GrievanceState = "MAPPED_TO_USER"
	StateDiagnosticTriage         GrievanceState = "DIAGNOSTIC_TRIAGE"
	StateInvestigationPending     GrievanceState = "INVESTIGATION_PENDING"
	StateAuditDossierCollected    GrievanceState = "AUDIT_DOSSIER_COLLECTED"
	StateATRDraftedMaker          GrievanceState = "ATR_DRAFTED_MAKER"
	StateATRApprovedChecker       GrievanceState = "ATR_APPROVED_CHECKER"
	StateATRSubmittedToRegulator  GrievanceState = "ATR_SUBMITTED_TO_REGULATOR"
	StateFirstReviewEscalated     GrievanceState = "FIRST_REVIEW_ESCALATED"
	StateODRConciliation          GrievanceState = "ODR_CONCILIATION"
	StateResolvedClosed           GrievanceState = "RESOLVED_CLOSED"
	StateRejectedInvalid          GrievanceState = "REJECTED_INVALID"
)

// ResolutionDisposition represents the formal resolution classification of the complaint
type ResolutionDisposition string

const (
	DispositionResolvedSatisfied   ResolutionDisposition = "RESOLVED_SATISFIED"
	DispositionResolvedWithRefund  ResolutionDisposition = "RESOLVED_WITH_REFUND"
	DispositionRejectedUnfounded   ResolutionDisposition = "REJECTED_UNFOUNDED"
	DispositionClarificationProvided ResolutionDisposition = "CLARIFICATION_PROVIDED"
	DispositionSettledViaODR       ResolutionDisposition = "SETTLED_VIA_ODR"
)

// EscalationStage represents the statutory 21-day escalation milestone
type EscalationStage int

const (
	Stage0Normal           EscalationStage = 0 // Day 0 - 6
	Stage1Day7Warning      EscalationStage = 1 // Day 7 - 13: Investigator reminder
	Stage2Day14Escalation  EscalationStage = 2 // Day 14 - 17: Principal Compliance Officer (PCO)
	Stage3Day18CriticalP1  EscalationStage = 3 // Day 18 - 19: Senior Legal Lead / RUNBOOK-34 War Room
	Stage4Day20Freeze      EscalationStage = 4 // Day 20: Mandatory submission freeze
	Stage5Day21Breached    EscalationStage = 5 // Day 21+: Statutory SLA breach boundary
)

// ScoresWebhookPayload represents incoming docket payload from SEBI SCORES 2.0 API
type ScoresWebhookPayload struct {
	ScoresRegistrationNumber string    `json:"scores_registration_number"`
	ComplainantName          string    `json:"complainant_name"`
	ComplainantPAN           string    `json:"complainant_pan"`
	ComplainantEmail         string    `json:"complainant_email,omitempty"`
	ComplainantMobile        string    `json:"complainant_mobile,omitempty"`
	ComplaintCategory        string    `json:"complaint_category"`
	SubCategory              string    `json:"sub_category,omitempty"`
	ComplaintDescription     string    `json:"complaint_description"`
	ReceiptDate              time.Time `json:"receipt_date"`
	DisputedAmount           float64   `json:"disputed_amount,omitempty"`
	DematBOID                string    `json:"demat_boid,omitempty"`
}

// ComplaintDocket represents an active regulatory complaint record
type ComplaintDocket struct {
	ID                         string         `json:"id"`
	ExternalRegistrationNumber string         `json:"external_registration_number"`
	Regulator                  RegulatorType  `json:"regulator"`
	CurrentState               GrievanceState `json:"current_state"`

	// Anonymized/hashed PII to satisfy DPDP Act 2023 & SEBI regulations
	ComplainantPANHash      string `json:"complainant_pan_hash"`
	ComplainantNameMasked   string `json:"complainant_name_masked"`
	ComplainantEmailMasked  string `json:"complainant_email_masked"`
	ComplainantMobileMasked string `json:"complainant_mobile_masked"`
	UserID                  string `json:"user_id,omitempty"` // Mapped internal user ID

	Category             string  `json:"category"`
	SubCategory          string  `json:"sub_category,omitempty"`
	DisputedAmount       float64 `json:"disputed_amount"`
	ComplaintDescription string  `json:"complaint_description"`

	// Statutory SLA Timestamps
	StatutorySLADays       int             `json:"statutory_sla_days"` // 21 for SEBI/SMART ODR, 30 for RBI
	IntakeTimestamp        time.Time       `json:"intake_timestamp"`
	StatutoryDeadline      time.Time       `json:"statutory_deadline"`
	ResolvedAt             *time.Time      `json:"resolved_at,omitempty"`
	CurrentEscalationStage EscalationStage `json:"current_escalation_stage"`
	IsSLABreached          bool            `json:"is_sla_breached"`

	AssignedOfficerID string `json:"assigned_officer_id,omitempty"`

	// Evidence & ATR linking
	ComplaintHash string `json:"complaint_hash"` // SHA256 of intake metadata
	DossierID     string `json:"dossier_id,omitempty"`
	ATRId         string `json:"atr_id,omitempty"`
	BesuTxHash    string `json:"besu_tx_hash,omitempty"`
}

// EvidenceDossier bundles gathered evidence across microservices
type EvidenceDossier struct {
	ID                       string            `json:"id"`
	ComplaintID              string            `json:"complaint_id"`
	KYCProfileSnapshot       map[string]string `json:"kyc_profile_snapshot"`
	LedgerStatementRef       string            `json:"ledger_statement_ref"`
	TradeContractNotes       []string          `json:"trade_contract_notes"`
	AuditLogEventIDs         []string          `json:"audit_log_event_ids"`
	DossierSHA256            string            `json:"dossier_sha256"`
	HarvestingCompletedAt    time.Time         `json:"harvesting_completed_at"`
}

// ActionTakenReport represents the formal ATR submitted to SEBI SCORES 2.0
type ActionTakenReport struct {
	ID                     string                `json:"id"`
	ComplaintID            string                `json:"complaint_id"`
	Disposition            ResolutionDisposition `json:"disposition"`
	ActionTakenSummary     string                `json:"action_taken_summary"`
	DetailedRedressalText  string                `json:"detailed_redressal_text"`
	RefundAmount           float64               `json:"refund_amount,omitempty"`
	SettlementBankUTR      string                `json:"settlement_bank_utr,omitempty"`
	SupportingDocumentRefs []string              `json:"supporting_document_refs,omitempty"`

	// Dual-Control Maker-Checker Tracking (Must NOT be the same user)
	MakerUserID      string     `json:"maker_user_id"`
	MakerDraftedAt   time.Time  `json:"maker_drafted_at"`
	CheckerUserID    string     `json:"checker_user_id,omitempty"`
	CheckerApprovedAt *time.Time `json:"checker_approved_at,omitempty"`
	CheckerNotes     string     `json:"checker_notes,omitempty"`

	// Class-3 Digital Signature Certificate (DSC) & Packaging
	ATRPackageHash string     `json:"atr_package_hash"`
	IsDSCSigned    bool       `json:"is_dsc_signed"`
	DSCSignerDN    string     `json:"dsc_signer_dn,omitempty"`
	DSCSignedAt    *time.Time `json:"dsc_signed_at,omitempty"`

	// Regulatory Submission & On-Chain Anchoring
	RegulatorySubmissionStatus string     `json:"regulatory_submission_status"` // DRAFT, APPROVED, SUBMITTED, ACKNOWLEDGED, REJECTED
	RegulatoryAckNumber        string     `json:"regulatory_ack_number,omitempty"`
	RegulatoryAckTimestamp     *time.Time `json:"regulatory_ack_timestamp,omitempty"`
	BesuTxHash                 string     `json:"besu_tx_hash,omitempty"`
}

// EscalationAlert represents an SLA stage change or breach alert
type EscalationAlert struct {
	ComplaintID                string          `json:"complaint_id"`
	ExternalRegistrationNumber string          `json:"external_registration_number"`
	Stage                      EscalationStage `json:"stage"`
	DaysElapsed                int             `json:"days_elapsed"`
	DaysRemaining              int             `json:"days_remaining"`
	AlertSeverity              string          `json:"alert_severity"` // INFO, WARNING, HIGH, CRITICAL_P1, SLA_BREACH
	Message                    string          `json:"message"`
	TriggeredAt                time.Time       `json:"triggered_at"`
}

// SLADashboard provides aggregate metrics on open regulatory tickets
type SLADashboard struct {
	TotalActiveDockets      int             `json:"total_active_dockets"`
	CriticalBreachRiskCount int             `json:"critical_breach_risk_count"` // Stage >= 3
	BreachedCount           int             `json:"breached_count"`
	ResolvedCount           int             `json:"resolved_count"`
	DocketsByStage          map[int]int     `json:"dockets_by_stage"`
	CalculatedAt            time.Time       `json:"calculated_at"`
}

// Common Domain Errors
var (
	ErrInvalidPANFormat        = errors.New("invalid PAN format: expected 5 uppercase letters, 4 digits, 1 uppercase letter")
	ErrDuplicateDocket         = errors.New("duplicate docket: complaint already ingested")
	ErrInvalidHMACSignature    = errors.New("invalid webhook HMAC-SHA256 signature")
	ErrDocketNotFound          = errors.New("complaint docket not found")
	ErrATRNotFound             = errors.New("action taken report not found")
	ErrDualControlViolation    = errors.New("maker-checker dual-control violation: maker cannot approve own ATR")
	ErrCheckerRequired         = errors.New("checker approval required before regulatory submission")
	ErrDSCNotSigned            = errors.New("ATR must be digitally signed with Class-3 DSC before submission")
	ErrInvalidStateTransition  = errors.New("invalid grievance state transition")
	ErrDocketAlreadyResolved   = errors.New("complaint is already resolved or closed")
)
