package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// Matches masked Aadhaar e.g. XXXX-XXXX-1234, ********1234, XXXX XXXX 1234, XXXXXXXX1234
	maskedAadhaarRegex = regexp.MustCompile(`^(?:[X*]{4}[-\s]?[X*]{4}[-\s]?[0-9]{4}|[X*]{8}[0-9]{4})$`)
	// Unmasked 12 digit raw Aadhaar pattern
	rawAadhaarRegex = regexp.MustCompile(`^[0-9]{4}[-\s]?[0-9]{4}[-\s]?[0-9]{4}$|^[0-9]{12}$`)
	// CKYCR 14 digit pattern
	ckycRegex = regexp.MustCompile(`^[0-9A-Z]{14}$`)
)

type AadhaarDataVaultValidator struct {
	secretSalt string
}

func NewAadhaarDataVaultValidator(secretSalt string) *AadhaarDataVaultValidator {
	return &AadhaarDataVaultValidator{
		secretSalt: secretSalt,
	}
}

// ValidateMaskedAadhaar ensures that raw 12-digit Aadhaar is strictly rejected and 8-digit masking is enforced
func (v *AadhaarDataVaultValidator) ValidateMaskedAadhaar(maskedNumber string) (last4 string, err error) {
	clean := strings.TrimSpace(maskedNumber)
	
	// If raw 12 digit unmasked Aadhaar is detected, trigger security error
	if rawAadhaarRegex.MatchString(clean) {
		return "", errors.New("CRITICAL COMPLIANCE VIOLATION: Unredacted raw 12-digit Aadhaar detected; UIDAI regulations require 8-digit masking (XXXX-XXXX-1234)")
	}

	if !maskedAadhaarRegex.MatchString(clean) {
		return "", errors.New("invalid masked Aadhaar format; must mask first 8 digits (e.g. XXXX-XXXX-1234 or ********1234)")
	}

	// Extract last 4 digits
	digitsOnly := regexp.MustCompile(`[0-9]`).FindAllString(clean, -1)
	if len(digitsOnly) != 4 {
		return "", errors.New("masked Aadhaar must expose exactly the last 4 digits")
	}

	return strings.Join(digitsOnly, ""), nil
}

// GenerateVaultToken creates an opaque UUID token for the Aadhaar Data Vault
func (v *AadhaarDataVaultValidator) GenerateVaultToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ADV-TOK-%s", hex.EncodeToString(bytes)), nil
}

// HashVaultReference creates a deterministic HMAC/hash of the masked Aadhaar + investor UUID + salt
func (v *AadhaarDataVaultValidator) HashVaultReference(investorUUID, last4 string) string {
	h := sha256.New()
	h.Write([]byte(investorUUID + ":" + last4 + ":" + v.secretSalt))
	return hex.EncodeToString(h.Sum(nil))
}

// ValidateCKYCNumber validates 14-digit Central KYC Registry format
func (v *AadhaarDataVaultValidator) ValidateCKYCNumber(ckyc string) error {
	clean := strings.TrimSpace(strings.ToUpper(ckyc))
	if len(clean) != 14 || !ckycRegex.MatchString(clean) {
		return errors.New("CKYC number must be a valid 14-character CERSAI alphanumeric string")
	}
	return nil
}
