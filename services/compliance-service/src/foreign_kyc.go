package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	iso3CountryRegex = regexp.MustCompile(`^[A-Z]{3}$`)
)

type ForeignKYCValidator struct {
	sanctionsEngine *SanctionsPEPEngine
	secretSalt      string
}

func NewForeignKYCValidator(sanctions *SanctionsPEPEngine, salt string) *ForeignKYCValidator {
	return &ForeignKYCValidator{
		sanctionsEngine: sanctions,
		secretSalt:      salt,
	}
}

// computeMRZCheckDigit calculates the ICAO 9303 check digit (weights 7, 3, 1, 7, 3, 1...)
func computeMRZCheckDigit(input string) int {
	weights := []int{7, 3, 1}
	sum := 0
	for i, ch := range input {
		val := 0
		if ch >= '0' && ch <= '9' {
			val = int(ch - '0')
		} else if ch >= 'A' && ch <= 'Z' {
			val = int(ch - 'A' + 10)
		} else if ch == '<' {
			val = 0
		}
		weight := weights[i%3]
		sum += val * weight
	}
	return sum % 10
}

// ValidateMRZLines validates 2-line ICAO Doc 9303 Machine Readable Zone strings
func (v *ForeignKYCValidator) ValidateMRZLines(line1, line2 string) (passportNo string, nationality string, expiryDate time.Time, err error) {
	line1 = strings.TrimSpace(strings.ToUpper(line1))
	line2 = strings.TrimSpace(strings.ToUpper(line2))

	if len(line1) != 44 || len(line2) != 44 {
		return "", "", time.Time{}, fmt.Errorf("ICAO 9303 MRZ requires exactly 44 characters per line (Line 1: %d, Line 2: %d)", len(line1), len(line2))
	}

	if line1[0] != 'P' {
		return "", "", time.Time{}, errors.New("ICAO 9303 Line 1 must begin with document type 'P'")
	}

	// Line 2 format:
	// 0..8: Passport number (9 chars)
	// 9: Check digit for passport number
	// 10..12: Nationality (3 chars)
	// 13..18: Date of birth (YYMMDD)
	// 19: Check digit for DOB
	// 20: Sex (M/F/<)
	// 21..26: Expiry date (YYMMDD)
	// 27: Check digit for expiry date
	rawPassportNo := line2[0:9]
	checkDigitPassport := int(line2[9] - '0')
	expectedPassportCheck := computeMRZCheckDigit(rawPassportNo)
	if checkDigitPassport != expectedPassportCheck {
		return "", "", time.Time{}, fmt.Errorf("MRZ checksum mismatch for passport number: expected %d, got %d", expectedPassportCheck, checkDigitPassport)
	}

	passportNo = strings.ReplaceAll(rawPassportNo, "<", "")
	nationality = line2[10:13]

	// Expiry date verification
	rawExpiry := line2[21:27]
	checkDigitExpiry := int(line2[27] - '0')
	expectedExpiryCheck := computeMRZCheckDigit(rawExpiry)
	if checkDigitExpiry != expectedExpiryCheck {
		return "", "", time.Time{}, fmt.Errorf("MRZ checksum mismatch for expiry date: expected %d, got %d", expectedExpiryCheck, checkDigitExpiry)
	}

	yy, _ := strconv.Atoi(rawExpiry[0:2])
	mm, _ := strconv.Atoi(rawExpiry[2:4])
	dd, _ := strconv.Atoi(rawExpiry[4:6])

	// Assume 2000s for YY
	fullYear := 2000 + yy
	expiryDate = time.Date(fullYear, time.Month(mm), dd, 0, 0, 0, 0, time.UTC)

	// Validate remaining validity (> 6 months)
	sixMonthsLater := time.Now().AddDate(0, 6, 0)
	if expiryDate.Before(sixMonthsLater) {
		return passportNo, nationality, expiryDate, errors.New("passport validity must be greater than 6 months from onboarding date")
	}

	return passportNo, nationality, expiryDate, nil
}

// ValidateFATCACRS validates tax classification under US FATCA and OECD CRS
func (v *ForeignKYCValidator) ValidateFATCACRS(taxCountryISO3, tinVaultToken string, isUSPerson bool, fatcaStatus string) error {
	taxCountry := strings.ToUpper(strings.TrimSpace(taxCountryISO3))
	if !iso3CountryRegex.MatchString(taxCountry) {
		return errors.New("tax residency country must be a valid 3-letter ISO-3166-1 alpha-3 code")
	}

	if strings.TrimSpace(tinVaultToken) == "" {
		return errors.New("Tax Identification Number (TIN) vault token is mandatory")
	}

	if isUSPerson {
		if fatcaStatus != "COMPLIANT_W9" {
			return errors.New("US Persons must complete Form W-9 certification (status COMPLIANT_W9)")
		}
	} else {
		if fatcaStatus != "COMPLIANT_W8BEN" && fatcaStatus != "EXEMPT" {
			return errors.New("Non-US foreign investors must complete Form W-8BEN certification")
		}
	}

	return nil
}

// HashPassport produces a salted SHA-256 hash of the passport number
func (v *ForeignKYCValidator) HashPassport(passportNo string) string {
	h := sha256.New()
	h.Write([]byte(passportNo + ":" + v.secretSalt))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyForeignInvestor runs the full GIFT City IFSC foreign investor onboarding pipeline
func (v *ForeignKYCValidator) VerifyForeignInvestor(req *ForeignInvestorProfile, mrz1, mrz2 string) error {
	// 1. FATF Blacklist Check
	if v.sanctionsEngine.IsFATFBlacklisted(req.NationalityISO3) {
		req.FATFBlacklisted = true
		return fmt.Errorf("onboarding prohibited: nationality '%s' is on FATF high-risk blacklist", req.NationalityISO3)
	}

	// 2. Sanctions & PEP Screening
	sanctioned, src, _, _, _, frozen := v.sanctionsEngine.ComprehensiveScreening(req.InvestorUUID, req.NationalityISO3)
	if sanctioned || frozen {
		req.SanctionsClear = false
		req.AccountFrozen = true
		return fmt.Errorf("sanctions violation detected from source: %s", src)
	}
	req.SanctionsClear = true

	// 3. ICAO 9303 MRZ Passport Check
	passportNo, nat, expiry, err := v.ValidateMRZLines(mrz1, mrz2)
	if err != nil {
		return fmt.Errorf("passport verification failed: %w", err)
	}
	req.MRZValid = true
	req.PassportExpiryDate = expiry
	req.PassportHashSHA256 = v.HashPassport(passportNo)
	if nat != req.NationalityISO3 {
		return fmt.Errorf("MRZ nationality '%s' does not match declared nationality '%s'", nat, req.NationalityISO3)
	}

	// 4. Biometric Liveness & Anti-Spoof
	if req.BiometricScore < 0.85 || !req.AntiSpoofPassed {
		return errors.New("biometric 3D face liveness score below required threshold (0.85) or presentation attack detected")
	}

	// 5. FATCA / CRS Verification
	err = v.ValidateFATCACRS(req.TaxResidencyISO3, req.TINVaultToken, req.IsUSPerson, req.FATCAStatus)
	if err != nil {
		return fmt.Errorf("FATCA/CRS tax certification failed: %w", err)
	}

	req.AssignedTier = KYCTier2FullCKYC
	req.JurisdictionFlag = "JURISDICTION_IFSCA"
	req.VerifiedAt = time.Now().UTC()
	req.ReKYCDueAt = req.VerifiedAt.Add(2 * 365 * 24 * time.Hour) // 2 year re-KYC for foreign retail

	return nil
}
