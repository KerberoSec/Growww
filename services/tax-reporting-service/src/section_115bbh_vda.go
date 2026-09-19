package main

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// VDATaxAssessment represents the tax calculation for a single VDA disposal transaction
type VDATaxAssessment struct {
	AssessmentID       string    `json:"assessment_id"`
	UserID             string    `json:"user_id"`
	PAN                string    `json:"pan"`
	Symbol             string    `json:"symbol"`
	SaleProceedsINR    float64   `json:"sale_proceeds_inr"`
	CostOfAcquisition  float64   `json:"cost_of_acquisition_inr"`
	RealizedGainINR    float64   `json:"realized_gain_inr"`
	TaxRatePct         float64   `json:"tax_rate_pct"`       // 30.0% flat under Section 115BBH
	SurchargeCessPct   float64   `json:"surcharge_cess_pct"` // 4.0% Health & Education Cess
	EffectiveTaxRatePct float64  `json:"effective_tax_rate_pct"` // 31.20%
	TotalTaxPayableINR float64   `json:"total_tax_payable_inr"`
	CalculatedAt       time.Time `json:"calculated_at"`
}

// TaxLot tracks an individual purchase acquisition of a VDA for FIFO matching
type TaxLot struct {
	LotID             string    `json:"lot_id"`
	UserID            string    `json:"user_id"`
	Symbol            string    `json:"symbol"`
	Quantity          float64   `json:"quantity"`
	RemainingQuantity float64   `json:"remaining_quantity"`
	CostPerUnitINR    float64   `json:"cost_per_unit_inr"`
	TotalCostINR      float64   `json:"total_cost_inr"`
	AcquiredAt        time.Time `json:"acquired_at"`
}

// LotDepletionRecord tracks an individual FIFO slice matched during disposal
type LotDepletionRecord struct {
	LotID             string    `json:"lot_id"`
	DepletedQuantity  float64   `json:"depleted_quantity"`
	CostPerUnitINR    float64   `json:"cost_per_unit_inr"`
	CostOfAcquisition float64   `json:"cost_of_acquisition_inr"`
	SalePricePerUnit  float64   `json:"sale_price_per_unit"`
	SaleProceedsINR   float64   `json:"sale_proceeds_inr"`
	RealizedGainINR   float64   `json:"realized_gain_inr"`
	TaxableGainINR    float64   `json:"taxable_gain_inr"` // gain if > 0 else 0 (no loss set-off)
	TaxPayableINR     float64   `json:"tax_payable_inr"`  // 31.2% on taxable gain
	AcquiredAt        time.Time `json:"acquired_at"`
}

// ScheduleVDAEntry corresponds directly to the statutory Schedule VDA columns in Indian ITR-2 / ITR-3
type ScheduleVDAEntry struct {
	SerialNo            int       `json:"sr_no"`
	DateOfAcquisition   string    `json:"date_of_acquisition"`
	DateOfTransfer      string    `json:"date_of_transfer"`
	HeadOfIncome        string    `json:"head_of_income"` // "Capital Gains"
	CostOfAcquisition   float64   `json:"cost_of_acquisition_inr"`
	ConsiderationAmount float64   `json:"consideration_amount_inr"`
	RealizedGainOrLoss  float64   `json:"realized_gain_or_loss_inr"`
	TaxableUnder115BBH  float64   `json:"taxable_under_115bbh_inr"`
	TaxPayable31_2Pct   float64   `json:"tax_payable_inr"`
	AssetDescription    string    `json:"asset_description"`
}

// AnnualVDATaxSummary consolidates all trades for a financial year enforcing Section 115BBH invariants
type AnnualVDATaxSummary struct {
	FinancialYear              string             `json:"financial_year"` // e.g. "2026-2027"
	UserID                     string             `json:"user_id"`
	PAN                        string             `json:"pan"`
	TotalGrossConsiderationINR float64            `json:"total_gross_consideration_inr"`
	TotalCostOfAcquisitionINR  float64            `json:"total_cost_of_acquisition_inr"`
	TotalRealizedGainsINR      float64            `json:"total_realized_gains_inr"`      // Sum of positive gains
	TotalRealizedLossesINR     float64            `json:"total_realized_losses_inr"`     // Sum of losses (absolute value)
	TaxableIncome115BBHINR     float64            `json:"taxable_income_115bbh_inr"`     // Strictly equal to TotalRealizedGainsINR (NO SET-OFF)
	BaseTax30PctINR            float64            `json:"base_tax_30_pct_inr"`
	HealthEducationCess4PctINR float64            `json:"cess_4_pct_inr"`
	TotalTaxLiabilityINR       float64            `json:"total_tax_liability_inr"`
	LossCarriedForwardINR      float64            `json:"loss_carried_forward_inr"`      // Strictly 0.0 under Section 115BBH(2)(b)
	DisallowedExpensesINR      float64            `json:"disallowed_expenses_inr"`       // Platform, gas, advisory fees disallowed
	ScheduleVDAEntries         []ScheduleVDAEntry `json:"schedule_vda_entries"`
	GeneratedAt                time.Time          `json:"generated_at"`
}

// FIFOTaxEngine manages user tax lots and calculates capital gains strictly compliant with Section 115BBH
type FIFOTaxEngine struct {
	mu      sync.Mutex
	userLots map[string]map[string][]*TaxLot // userID -> symbol -> FIFO lot queue
	records  map[string]*VDATaxAssessment
}

func NewFIFOTaxEngine() *FIFOTaxEngine {
	return &FIFOTaxEngine{
		userLots: make(map[string]map[string][]*TaxLot),
		records:  make(map[string]*VDATaxAssessment),
	}
}

// Compute115BBHTax calculates statutory 30% flat tax + 4% cess on positive VDA gains (single transaction)
func Compute115BBHTax(assessmentID, userID, pan, symbol string, proceedsINR, costINR float64) *VDATaxAssessment {
	gain := proceedsINR - costINR
	var taxPayable float64

	// Section 115BBH invariant: losses cannot be set off; taxable only if gain > 0
	// Base tax: 30%, Health & Education Cess: 4% -> Effective 31.2%
	if gain > 0 {
		baseTax := gain * 0.30
		cess := baseTax * 0.04
		taxPayable = math.Round(baseTax + cess)
	} else {
		taxPayable = 0.0
	}

	return &VDATaxAssessment{
		AssessmentID:        assessmentID,
		UserID:              userID,
		PAN:                 pan,
		Symbol:              symbol,
		SaleProceedsINR:     proceedsINR,
		CostOfAcquisition:   costINR,
		RealizedGainINR:     gain,
		TaxRatePct:          30.0,
		SurchargeCessPct:    4.0,
		EffectiveTaxRatePct: 31.20,
		TotalTaxPayableINR:  taxPayable,
		CalculatedAt:        time.Now().UTC(),
	}
}

// AddAcquisitionLot adds a purchase lot to the user's FIFO inventory
func (e *FIFOTaxEngine) AddAcquisitionLot(userID, symbol string, quantity, costPerUnitINR float64, acquiredAt time.Time) *TaxLot {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.userLots[userID] == nil {
		e.userLots[userID] = make(map[string][]*TaxLot)
	}

	lotID := fmt.Sprintf("LOT-%s-%d", symbol, acquiredAt.UnixNano())
	lot := &TaxLot{
		LotID:             lotID,
		UserID:            userID,
		Symbol:            symbol,
		Quantity:          quantity,
		RemainingQuantity: quantity,
		CostPerUnitINR:    costPerUnitINR,
		TotalCostINR:      quantity * costPerUnitINR,
		AcquiredAt:        acquiredAt,
	}

	e.userLots[userID][symbol] = append(e.userLots[userID][symbol], lot)
	return lot
}

// ProcessDisposal executes FIFO matching for a sell order, computing realized gains per slice and 115BBH liability
func (e *FIFOTaxEngine) ProcessDisposal(userID, pan, symbol string, soldQty, salePriceINR float64, disposedAt time.Time) ([]LotDepletionRecord, float64, float64, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if soldQty <= 0 {
		return nil, 0, 0, errors.New("sold quantity must be strictly positive")
	}

	symbolLots, ok := e.userLots[userID][symbol]
	if !ok || len(symbolLots) == 0 {
		return nil, 0, 0, fmt.Errorf("no tax lots available for user %s symbol %s", userID, symbol)
	}

	// Verify total available inventory
	var totalAvail float64
	for _, lot := range symbolLots {
		totalAvail += lot.RemainingQuantity
	}
	if totalAvail < soldQty-1e-9 {
		return nil, 0, 0, fmt.Errorf("insufficient lot inventory: requested %.8f, available %.8f", soldQty, totalAvail)
	}

	qtyRemainingToMatch := soldQty
	var depletions []LotDepletionRecord
	var totalTaxableGain float64
	var totalTaxPayable float64

	for _, lot := range symbolLots {
		if qtyRemainingToMatch <= 1e-9 {
			break
		}
		if lot.RemainingQuantity <= 1e-9 {
			continue
		}

		matchedQty := math.Round(math.Min(qtyRemainingToMatch, lot.RemainingQuantity)*1e8) / 1e8
		costOfAcq := matchedQty * lot.CostPerUnitINR
		saleProceeds := matchedQty * salePriceINR
		gain := saleProceeds - costOfAcq

		var taxableGain float64
		var taxPayable float64
		if gain > 0 {
			taxableGain = gain
			// 30% base + 4% cess = 31.2%
			taxPayable = math.Round(gain * 0.312)
		} else {
			taxableGain = 0.0
			taxPayable = 0.0
		}

		lot.RemainingQuantity = math.Round((lot.RemainingQuantity-matchedQty)*1e8) / 1e8
		qtyRemainingToMatch = math.Round((qtyRemainingToMatch-matchedQty)*1e8) / 1e8

		dep := LotDepletionRecord{
			LotID:             lot.LotID,
			DepletedQuantity:  matchedQty,
			CostPerUnitINR:    lot.CostPerUnitINR,
			CostOfAcquisition: costOfAcq,
			SalePricePerUnit:  salePriceINR,
			SaleProceedsINR:   saleProceeds,
			RealizedGainINR:   gain,
			TaxableGainINR:    taxableGain,
			TaxPayableINR:     taxPayable,
			AcquiredAt:        lot.AcquiredAt,
		}
		depletions = append(depletions, dep)
		totalTaxableGain += taxableGain
		totalTaxPayable += taxPayable
	}

	return depletions, totalTaxableGain, totalTaxPayable, nil
}

// GenerateAnnualVDASummary creates a comprehensive statutory tax summary enforcing strict Section 115BBH rules
func GenerateAnnualVDASummary(fy, userID, pan string, trades []LotDepletionRecord, claimedExpensesINR float64) *AnnualVDATaxSummary {
	var totalGross, totalCost, totalGains, totalLosses float64
	var entries []ScheduleVDAEntry

	for i, t := range trades {
		totalGross += t.SaleProceedsINR
		totalCost += t.CostOfAcquisition

		if t.RealizedGainINR > 0 {
			totalGains += t.RealizedGainINR
		} else {
			totalLosses += math.Abs(t.RealizedGainINR)
		}

		entry := ScheduleVDAEntry{
			SerialNo:            i + 1,
			DateOfAcquisition:   t.AcquiredAt.Format("2006-01-02"),
			DateOfTransfer:      time.Now().Format("2006-01-02"),
			HeadOfIncome:        "Capital Gains",
			CostOfAcquisition:   math.Round(t.CostOfAcquisition*100) / 100,
			ConsiderationAmount: math.Round(t.SaleProceedsINR*100) / 100,
			RealizedGainOrLoss:  math.Round(t.RealizedGainINR*100) / 100,
			TaxableUnder115BBH:  math.Round(t.TaxableGainINR*100) / 100,
			TaxPayable31_2Pct:   math.Round(t.TaxPayableINR*100) / 100,
			AssetDescription:    t.LotID,
		}
		entries = append(entries, entry)
	}

	// CRITICAL STATUTORY INVARIANT (Section 115BBH(2)(b)):
	// Losses from VDA CANNOT be set off against gains from any other VDA.
	// Taxable base is strictly sum of positive gains.
	taxableIncome := totalGains
	baseTax := taxableIncome * 0.30
	cess := baseTax * 0.04
	totalLiability := math.Round(baseTax + cess)

	return &AnnualVDATaxSummary{
		FinancialYear:              fy,
		UserID:                     userID,
		PAN:                        pan,
		TotalGrossConsiderationINR: math.Round(totalGross*100) / 100,
		TotalCostOfAcquisitionINR:  math.Round(totalCost*100) / 100,
		TotalRealizedGainsINR:      math.Round(totalGains*100) / 100,
		TotalRealizedLossesINR:     math.Round(totalLosses*100) / 100,
		TaxableIncome115BBHINR:     math.Round(taxableIncome*100) / 100,
		BaseTax30PctINR:            math.Round(baseTax*100) / 100,
		HealthEducationCess4PctINR: math.Round(cess*100) / 100,
		TotalTaxLiabilityINR:       totalLiability,
		LossCarriedForwardINR:      0.0, // Strictly 0.0 - carry forward prohibited
		DisallowedExpensesINR:      claimedExpensesINR, // Any expenses other than CoA disallowed
		ScheduleVDAEntries:         entries,
		GeneratedAt:                time.Now().UTC(),
	}
}
