package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type ComplianceEngineService struct {
	mu                  sync.RWMutex
	panValidator        *PANValidator
	aadhaarValidator    *AadhaarDataVaultValidator
	pennyDropValidator  *PennyDropValidator
	sanctionsEngine     *SanctionsPEPEngine
	tierLimitsManager   *TierLimitsManager
	fiuSurveillance     *FIUSurveillanceSystem
	foreignValidator    *ForeignKYCValidator
	sandboxGuardrails   *SandboxGuardrails
	commitmentGenerator *IdentityCommitmentGenerator
	domesticInvestors   map[string]*DomesticInvestorProfile
	foreignInvestors    map[string]*ForeignInvestorProfile
}

func NewComplianceEngineService(secretSalt string) *ComplianceEngineService {
	sanctions := NewSanctionsPEPEngine()
	return &ComplianceEngineService{
		panValidator:        NewPANValidator(secretSalt),
		aadhaarValidator:    NewAadhaarDataVaultValidator(secretSalt),
		pennyDropValidator:  NewPennyDropValidator(0.85),
		sanctionsEngine:     sanctions,
		tierLimitsManager:   NewTierLimitsManager(),
		fiuSurveillance:     NewFIUSurveillanceSystem(),
		foreignValidator:    NewForeignKYCValidator(sanctions, secretSalt),
		sandboxGuardrails:   NewSandboxGuardrails(),
		commitmentGenerator: NewIdentityCommitmentGenerator(secretSalt),
		domesticInvestors:   make(map[string]*DomesticInvestorProfile),
		foreignInvestors:    make(map[string]*ForeignInvestorProfile),
	}
}

// ProcessDomesticKYCOnboarding executes the mandatory 4-step domestic Indian onboarding pipeline
func (s *ComplianceEngineService) ProcessDomesticKYCOnboarding(profile *DomesticInvestorProfile) (*DomesticInvestorProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Sanctions Screening
	sanctioned, source, isPEP, pepCat, eddReq, frozen := s.sanctionsEngine.ComprehensiveScreening(profile.DeclaredName, "IND")
	if sanctioned || frozen {
		profile.SanctionsClear = false
		profile.AccountFrozen = true
		profile.SanctionsHitSource = source
		return profile, fmt.Errorf("domestic onboarding rejected: individual matches sanctions list %s", source)
	}
	profile.SanctionsClear = true
	profile.PEPStatus = pepCat
	profile.EDDApproved = !eddReq // If EDD required, pending Senior Management sign-off

	// 2. PAN Format and Operative Status Validation
	err := s.panValidator.VerifyPANStatus(profile.PAN, profile.PANStatus, profile.AadhaarPANLinked, true)
	if err != nil {
		return profile, fmt.Errorf("PAN validation failed: %w", err)
	}
	panHash := s.panValidator.HashPAN(profile.PAN)

	// 3. Aadhaar Masking and Data Vault Compliance (Strict 8-digit masking check)
	last4, err := s.aadhaarValidator.ValidateMaskedAadhaar(profile.AadhaarMasked)
	if err != nil {
		return profile, fmt.Errorf("Aadhaar Data Vault compliance failed: %w", err)
	}
	_ = last4
	if profile.AadhaarVaultToken == "" {
		token, err := s.aadhaarValidator.GenerateVaultToken()
		if err != nil {
			return profile, err
		}
		profile.AadhaarVaultToken = token
	}

	// 4. Bank Penny Drop Verification (Jaro-Winkler >= 0.85)
	score, err := s.pennyDropValidator.ValidatePennyDrop(profile.IFSC, profile.DeclaredName, profile.BeneficiaryName)
	if err != nil {
		profile.NameMatchScore = score
		return profile, fmt.Errorf("bank account penny drop rejected: %w", err)
	}
	profile.NameMatchScore = score

	// 5. Face Liveness & Anti-Spoof
	if profile.FaceLivenessScore < 0.90 || !profile.AntiSpoofPassed {
		return profile, errors.New("biometric face liveness verification failed: PAD score below 0.90 or anti-spoofing alert")
	}

	// 6. Tier Assignment & Risk Categorization
	if isPEP {
		profile.RiskTier = RiskTierHigh
	} else if profile.AnnualIncomeINR > 2500000.0 {
		profile.RiskTier = RiskTierMedium
	} else {
		profile.RiskTier = RiskTierLow
	}

	// CKYC verification check for Tier 2
	if profile.CKYCNumber != "" && s.aadhaarValidator.ValidateCKYCNumber(profile.CKYCNumber) == nil {
		profile.AssignedTier = KYCTier2FullCKYC
	} else {
		profile.AssignedTier = KYCTier1BasicOTP
	}

	// 7. On-Chain Cryptographic Identity Commitment Generation (Zero-PII)
	profile.CommitmentHash = s.commitmentGenerator.GenerateDomesticCommitment(panHash, profile.InvestorUUID)
	profile.VerifiedAt = time.Now().UTC()
	profile.ReKYCDueAt = s.tierLimitsManager.CalculateReKYCDue(profile.RiskTier, profile.VerifiedAt)

	// Persist in off-chain database
	s.domesticInvestors[profile.InvestorUUID] = profile
	return profile, nil
}

// UpgradeToTier3VCIP processes Video Customer Identification Procedure & CA Net Worth certificate
func (s *ComplianceEngineService) UpgradeToTier3VCIP(investorUUID string, vcipCompleted bool, caNetWorthINR float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	profile, exists := s.domesticInvestors[investorUUID]
	if !exists {
		return errors.New("investor profile not found")
	}

	if profile.AssignedTier < KYCTier2FullCKYC {
		return errors.New("must successfully complete Tier 2 (CKYC) before upgrading to Tier 3")
	}

	if !vcipCompleted {
		return errors.New("V-CIP video recording verification not completed")
	}

	if caNetWorthINR < 10000000.0 { // ₹1 Crore CA certified net worth minimum for institutional tier
		return errors.New("CA certified net worth must be >= ₹1,00,00,000 for Tier 3 Institutional onboarding")
	}

	profile.AssignedTier = KYCTier3VCIP
	return nil
}

// CheckWithdrawalLimit checks if a transaction is permitted under the investor's current tier
func (s *ComplianceEngineService) CheckWithdrawalLimit(investorUUID string, amountINR, dailySpentINR, annualSpentINR float64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profile, exists := s.domesticInvestors[investorUUID]
	if !exists {
		return errors.New("investor not found")
	}

	if profile.AccountFrozen {
		return errors.New("account is frozen due to statutory sanctions / AML compliance restriction")
	}

	return s.tierLimitsManager.CheckLimitEnforcement(profile.AssignedTier, amountINR, dailySpentINR, annualSpentINR)
}

// MonitorTransactionSurveillance checks transactions for cash structuring, CTR threshold, and Travel Rule
func (s *ComplianceEngineService) MonitorTransactionSurveillance(userID string, amountINR float64, recentTxs []float64, travelRule *TravelRulePayload) (*AMLAlert, error) {
	// 1. Structuring Evaluation
	allTxs := append(recentTxs, amountINR)
	strAlert := s.fiuSurveillance.EvaluateStructuring(userID, allTxs)
	if strAlert != nil {
		return strAlert, nil
	}

	// 2. CTR Threshold Evaluation (₹10 Lakh)
	ctrAlert := s.fiuSurveillance.EvaluateCashThreshold(userID, amountINR)
	if ctrAlert != nil {
		return ctrAlert, nil
	}

	// 3. Travel Rule Check for VDA transfers
	if travelRule != nil {
		err := s.fiuSurveillance.ValidateTravelRule(*travelRule)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
