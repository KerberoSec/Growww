package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
)

var (
	panRegex = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]$`)
	validEntityTypes = map[byte]string{
		'P': "INDIVIDUAL",
		'C': "COMPANY",
		'H': "HINDU_UNDIVIDED_FAMILY",
		'F': "PARTNERSHIP_FIRM",
		'A': "ASSOCIATION_OF_PERSONS",
		'T': "TRUST",
		'B': "BODY_OF_INDIVIDUALS",
		'L': "LOCAL_AUTHORITY",
		'J': "ARTIFICIAL_JURIDICAL_PERSON",
		'G': "GOVERNMENT",
	}
)

type PANValidator struct {
	secretSalt string
}

func NewPANValidator(secretSalt string) *PANValidator {
	return &PANValidator{
		secretSalt: secretSalt,
	}
}

// ValidateFormat checks if the PAN conforms strictly to Income Tax Department rules
func (v *PANValidator) ValidateFormat(pan string) (entityType string, err error) {
	pan = strings.TrimSpace(strings.ToUpper(pan))
	if len(pan) != 10 {
		return "", errors.New("PAN must be exactly 10 characters long")
	}

	if !panRegex.MatchString(pan) {
		return "", errors.New("invalid PAN structure; expected 5 letters, 4 digits, 1 letter")
	}

	entityChar := pan[3]
	eType, exists := validEntityTypes[entityChar]
	if !exists {
		return "", errors.New("invalid 4th character in PAN: unrecognized entity classification")
	}

	return eType, nil
}

// HashPAN returns the SHA-256 hashed PAN with salt for secure storage and zero-PII indexing
func (v *PANValidator) HashPAN(pan string) string {
	pan = strings.TrimSpace(strings.ToUpper(pan))
	h := sha256.New()
	h.Write([]byte(pan + ":" + v.secretSalt))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyPANStatus validates operational status and Aadhaar linking
func (v *PANValidator) VerifyPANStatus(pan string, status PANStatus, aadhaarLinked bool, isIndividual bool) error {
	_, err := v.ValidateFormat(pan)
	if err != nil {
		return err
	}

	if status != PANStatusOperative {
		return errors.New("PAN status is not OPERATIVE (status: " + string(status) + ")")
	}

	if isIndividual && !aadhaarLinked {
		return errors.New("mandatory PAN-Aadhaar linkage missing under Section 139AA of Income Tax Act")
	}

	return nil
}
