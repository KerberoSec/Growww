package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type IdentityCommitmentGenerator struct {
	systemSalt string
}

func NewIdentityCommitmentGenerator(salt string) *IdentityCommitmentGenerator {
	return &IdentityCommitmentGenerator{
		systemSalt: salt,
	}
}

// GenerateDomesticCommitment creates a 32-byte cryptographic commitment for on-chain ComplianceRegistry.sol
// keccak256/sha256(pan_hash + investor_uuid + salt)
func (g *IdentityCommitmentGenerator) GenerateDomesticCommitment(panHash, investorUUID string) string {
	payload := fmt.Sprintf("%s:%s:%s", panHash, investorUUID, g.systemSalt)
	h := sha256.Sum256([]byte(payload))
	return "0x" + hex.EncodeToString(h[:])
}

// GenerateForeignCommitment creates a 32-byte cryptographic commitment for GIFT City IFSCA foreign investor
func (g *IdentityCommitmentGenerator) GenerateForeignCommitment(passportHash, ifscInvestorID string) string {
	payload := fmt.Sprintf("%s:%s:%s:IFSCA", passportHash, ifscInvestorID, g.systemSalt)
	h := sha256.Sum256([]byte(payload))
	return "0x" + hex.EncodeToString(h[:])
}

// VerifyCommitmentIntegrity checks that a commitment is a valid 32-byte hex string (with 0x prefix)
func (g *IdentityCommitmentGenerator) VerifyCommitmentIntegrity(commitment string) bool {
	if !strings.HasPrefix(commitment, "0x") {
		return false
	}
	hexPart := strings.TrimPrefix(commitment, "0x")
	if len(hexPart) != 64 {
		return false
	}
	_, err := hex.DecodeString(hexPart)
	return err == nil
}
