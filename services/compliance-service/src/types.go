package main

import (
	"time"
)

// KYCTier represents the statutory KYC completion tier
type KYCTier int

const (
	KYCTier0Unverified KYCTier = 0
	KYCTier1BasicOTP   KYCTier = 1
	KYCTier2FullCKYC   KYCTier = 2
	KYCTier3VCIP       KYCTier = 3
)

func (t KYCTier) String() string {
	switch t {
	case KYCTier0Unverified:
		return "TIER_0_UNVERIFIED"
	case KYCTier1BasicOTP:
		return "TIER_1_BASIC_OTP"
	case KYCTier2FullCKYC:
		return "TIER_2_FULL_CKYC"
	case KYCTier3VCIP:
		return "TIER_3_INSTITUTIONAL_VCIP"
	default:
		return "UNKNOWN"
	}
}

// RiskTier represents the statutory AML risk category under PMLA 2002
type RiskTier string

const (
	RiskTierLow    RiskTier = "LOW"
	RiskTierMedium RiskTier = "MEDIUM"
	RiskTierHigh   RiskTier = "HIGH"
)

// PEPStatus represents the Politically Exposed Person status
type PEPStatus string

const (
	PEPNone           PEPStatus = "NONE"
	PEPDomestic       PEPStatus = "DOMESTIC_PEP"
	PEPForeign        PEPStatus = "FOREIGN_PEP"
	PEPCloseAssociate PEPStatus = "CLOSE_ASSOCIATE"
)

// SanctionsSource identifies which watchlist was hit
type SanctionsSource string

const (
	SanctionsNone    SanctionsSource = "NONE"
	SanctionsOFAC    SanctionsSource = "OFAC_SDN"
	SanctionsUN      SanctionsSource = "UN_CONSOLIDATED"
	SanctionsEU      SanctionsSource = "EU_FINANCIAL"
	SanctionsUKHMT   SanctionsSource = "UK_HM_TREASURY"
	SanctionsMHAUAPA SanctionsSource = "MHA_UAPA_INDIA"
)

// PANStatus represents the NSDL PAN status
type PANStatus string

const (
	PANStatusOperative   PANStatus = "OPERATIVE"
	PANStatusInoperative PANStatus = "INOPERATIVE"
	PANStatusDeletion    PANStatus = "DELETION"
	PANStatusInvalid     PANStatus = "INVALID"
)

// DomesticInvestorProfile holds off-chain KYC state for an Indian investor
type DomesticInvestorProfile struct {
	InvestorUUID        string    `json:"investor_uuid"`
	PAN                 string    `json:"pan"`
	PANStatus           PANStatus `json:"pan_status"`
	AadhaarPANLinked    bool      `json:"aadhaar_pan_linked"`
	AadhaarMasked       string    `json:"aadhaar_masked"`     // XXXX-XXXX-1234
	AadhaarVaultToken   string    `json:"aadhaar_vault_token"`
	CKYCNumber          string    `json:"ckyc_number"`        // 14 digits
	BankAccountToken    string    `json:"bank_account_token"`
	IFSC                string    `json:"ifsc"`
	BeneficiaryName     string    `json:"beneficiary_name"`
	DeclaredName        string    `json:"declared_name"`
	NameMatchScore      float64   `json:"name_match_score"`   // >= 0.85 required
	FaceLivenessScore   float64   `json:"face_liveness_score"`// >= 0.90 required
	AntiSpoofPassed     bool      `json:"anti_spoof_passed"`
	AnnualIncomeINR     float64   `json:"annual_income_inr"`
	Occupation          string    `json:"occupation"`
	AssignedTier        KYCTier   `json:"assigned_tier"`
	RiskTier            RiskTier  `json:"risk_tier"`
	PEPStatus           PEPStatus `json:"pep_status"`
	SanctionsClear      bool      `json:"sanctions_clear"`
	SanctionsHitSource  SanctionsSource `json:"sanctions_hit_source"`
	AccountFrozen       bool      `json:"account_frozen"`
	EDDApproved         bool      `json:"edd_approved"`
	CommitmentHash      string    `json:"commitment_hash"`    // 32-byte hex string
	VerifiedAt          time.Time `json:"verified_at"`
	ReKYCDueAt          time.Time `json:"re_kyc_due_at"`
}

// ForeignInvestorProfile holds off-chain KYC state for GIFT City foreign investor
type ForeignInvestorProfile struct {
	InvestorUUID         string    `json:"investor_uuid"`
	NationalityISO3      string    `json:"nationality_iso3"`
	PassportHashSHA256   string    `json:"passport_hash_sha256"`
	MRZValid             bool      `json:"mrz_valid"`
	PassportExpiryDate   time.Time `json:"passport_expiry_date"`
	BiometricScore       float64   `json:"biometric_score"`
	AntiSpoofPassed      bool      `json:"anti_spoof_passed"`
	TaxResidencyISO3     string    `json:"tax_residency_iso3"`
	TINVaultToken        string    `json:"tin_vault_token"`
	IsUSPerson           bool      `json:"is_us_person"`
	FATCAStatus          string    `json:"fatca_status"` // COMPLIANT_W8BEN, COMPLIANT_W9, EXEMPT
	FATFBlacklisted      bool      `json:"fatf_blacklisted"`
	SanctionsClear       bool      `json:"sanctions_clear"`
	AccountFrozen        bool      `json:"account_frozen"`
	PEPStatus            PEPStatus `json:"pep_status"`
	AssignedTier         KYCTier   `json:"assigned_tier"`
	CommitmentHash       string    `json:"commitment_hash"`
	JurisdictionFlag     string    `json:"jurisdiction_flag"` // JURISDICTION_IFSCA
	VerifiedAt           time.Time `json:"verified_at"`
	ReKYCDueAt           time.Time `json:"re_kyc_due_at"`
}

// TierLimits holds withdrawal & transaction caps for a KYC tier
type TierLimits struct {
	Tier       KYCTier `json:"tier"`
	DailyINR   float64 `json:"daily_inr"`
	AnnualINR  float64 `json:"annual_inr"`
	StatusDesc string  `json:"status_desc"`
}

// AMLAlert holds statutory suspicious transaction or cash transaction alert
type AMLAlert struct {
	AlertID        string    `json:"alert_id"`
	UserID         string    `json:"user_id"`
	RuleTriggered  string    `json:"rule_triggered"`
	Severity       string    `json:"severity"`
	TransactionAmt float64   `json:"transaction_amt"`
	ReportType     string    `json:"report_type"` // STR, CTR
	Description    string    `json:"description"`
	FiledWithFIU   bool      `json:"filed_with_fiu"`
	CreatedAt      time.Time `json:"created_at"`
}

// TravelRulePayload encapsulates IVMS-101 Travel Rule data for crypto transfers > ₹50,000
type TravelRulePayload struct {
	TransferID             string  `json:"transfer_id"`
	CryptoCurrency         string  `json:"crypto_currency"`
	AmountFiatINR          float64 `json:"amount_fiat_inr"`
	OriginatorName         string  `json:"originator_name"`
	OriginatorAccountID    string  `json:"originator_account_id"`
	OriginatorNationalID   string  `json:"originator_national_id"`
	OriginatorVASP         string  `json:"originator_vasp"`
	BeneficiaryName        string  `json:"beneficiary_name"`
	BeneficiaryWalletAddr  string  `json:"beneficiary_wallet_addr"`
	BeneficiaryVASP        string  `json:"beneficiary_vasp"`
	Compliant              bool    `json:"compliant"`
}
