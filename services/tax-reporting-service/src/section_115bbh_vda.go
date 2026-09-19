package main

import (
	"fmt"
	"math"
	"time"
)

type VDATaxAssessment struct {
	AssessmentID       string    `json:"assessment_id"`
	UserID             string    `json:"user_id"`
	PAN                string    `json:"pan"`
	Symbol             string    `json:"symbol"`
	SaleProceedsINR    float64   `json:"sale_proceeds_inr"`
	CostOfAcquisition  float64   `json:"cost_of_acquisition_inr"`
	RealizedGainINR    float64   `json:"realized_gain_inr"`
	TaxRatePct         float64   `json:"tax_rate_pct"` // 30.0% flat under 115BBH
	SurchargeCessPct   float64   `json:"surcharge_cess_pct"` // 4% Health & Education Cess
	TotalTaxPayableINR float64   `json:"total_tax_payable_inr"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

// Compute115BBHTax calculates statutory 30% flat tax + 4% cess on positive VDA gains
func Compute115BBHTax(assessmentID, userID, pan, symbol string, proceedsINR, costINR float64) *VDATaxAssessment {
	gain := proceedsINR - costINR
	var taxPayable float64

	// Section 115BBH invariant: losses cannot be set off; taxable only if gain > 0
	if gain > 0 {
		baseTax := gain * 0.30
		cess := baseTax * 0.04
		taxPayable = math.Round(baseTax + cess)
	} else {
		taxPayable = 0.0
	}

	return &VDATaxAssessment{
		AssessmentID:       assessmentID,
		UserID:             userID,
		PAN:                pan,
		Symbol:             symbol,
		SaleProceedsINR:    proceedsINR,
		CostOfAcquisition:  costINR,
		RealizedGainINR:    gain,
		TaxRatePct:         30.0,
		SurchargeCessPct:   4.0,
		TotalTaxPayableINR: taxPayable,
		CalculatedAt:       time.Now().UTC(),
	}
}
