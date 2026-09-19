package main

import (
	"errors"
	"fmt"
	"sync"
)

type SandboxGuardrails struct {
	mu                         sync.RWMutex
	maxDomesticUsers           uint32
	maxForeignUsers            uint32
	maxDomesticPortfolioINR    float64
	maxForeignPortfolioUSD     float64
	maxAggregateVolumeINR      float64
	currentDomesticUsers       uint32
	currentForeignUsers        uint32
	currentAggregateVolumeINR  float64
	permittedDomesticRails     map[string]bool
	zeroFeeEnforced            bool
}

func NewSandboxGuardrails() *SandboxGuardrails {
	return &SandboxGuardrails{
		maxDomesticUsers:        10000,
		maxForeignUsers:         5000,
		maxDomesticPortfolioINR: 50000.0,
		maxForeignPortfolioUSD:  10000.0,
		maxAggregateVolumeINR:   500000000.0, // ₹500 Million
		permittedDomesticRails: map[string]bool{
			"UPI":  true,
			"IMPS": true,
			"NEFT": true,
			"RTGS": true,
		},
		zeroFeeEnforced: true,
	}
}

// ValidateDomesticEnrollment checks SEBI sandbox user and portfolio limits
func (g *SandboxGuardrails) ValidateDomesticEnrollment(requestedPortfolioINR float64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.currentDomesticUsers >= g.maxDomesticUsers {
		return fmt.Errorf("SEBI Sandbox enrollment capped at %d domestic retail users", g.maxDomesticUsers)
	}

	if requestedPortfolioINR > g.maxDomesticPortfolioINR {
		return fmt.Errorf("portfolio request ₹%.2f exceeds SEBI sandbox limit of ₹%.2f per user",
			requestedPortfolioINR, g.maxDomesticPortfolioINR)
	}

	g.currentDomesticUsers++
	return nil
}

// ValidateForeignEnrollment checks IFSCA sandbox user and portfolio limits
func (g *SandboxGuardrails) ValidateForeignEnrollment(requestedPortfolioUSD float64) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.currentForeignUsers >= g.maxForeignUsers {
		return fmt.Errorf("IFSCA Sandbox enrollment capped at %d foreign investors", g.maxForeignUsers)
	}

	if requestedPortfolioUSD > g.maxForeignPortfolioUSD {
		return fmt.Errorf("portfolio request $%.2f exceeds IFSCA sandbox limit of $%.2f per user",
			requestedPortfolioUSD, g.maxForeignPortfolioUSD)
	}

	g.currentForeignUsers++
	return nil
}

// ValidateFiatPaymentRail ensures unregulated stablecoins are blocked and only RBI banking rails are used
func (g *SandboxGuardrails) ValidateFiatPaymentRail(railName string) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	rail := stringsToUpper(railName)
	if !g.permittedDomesticRails[rail] {
		return fmt.Errorf("prohibited payment rail '%s': only approved RBI domestic banking rails (UPI, IMPS, NEFT, RTGS) are permitted", railName)
	}
	return nil
}

// ValidateZeroFeeModel verifies that the platform fee rate is strictly 0.00% at launch
func (g *SandboxGuardrails) ValidateZeroFeeModel(makerFeeBps, takerFeeBps int, holdingFeeRate, aumFeeRate float64) error {
	if makerFeeBps != 0 || takerFeeBps != 0 {
		return fmt.Errorf("invariant violation: launch policy mandates 0.00%% transaction fees (maker: %d bps, taker: %d bps)", makerFeeBps, takerFeeBps)
	}

	if holdingFeeRate != 0.0 || aumFeeRate != 0.0 {
		return errors.New("invariant violation: zero holding and zero AUM fees are strictly enforced")
	}

	return nil
}

func stringsToUpper(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		b = append(b, c)
	}
	return string(b)
}
