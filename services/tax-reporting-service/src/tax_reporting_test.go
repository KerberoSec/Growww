package main

import (
	"strings"
	"testing"
	"time"
)

// TestValidatePAN verifies regex compliance for Indian Income Tax PAN
func TestValidatePAN(t *testing.T) {
	validPANs := []string{"ABCDE1234F", "AAACG7777K", "BNZPK9876Q", "BLRGP1234F"}
	for _, pan := range validPANs {
		if !ValidatePAN(pan) {
			t.Errorf("expected valid PAN for %s", pan)
		}
	}

	invalidPANs := []string{"ABCDE1234", "12345ABCDE", "ABCDE12345", "abcde1234f", "ABCDEF12345G", "BLRG01234F", ""}
	for _, pan := range invalidPANs {
		if ValidatePAN(pan) {
			t.Errorf("expected invalid PAN for %s", pan)
		}
	}
}

// TestSection194S_ExemptionThreshold_General tests the general ₹10,000 FY threshold
func TestSection194S_ExemptionThreshold_General(t *testing.T) {
	engine := NewSection194STaxEngine()
	pan := "ABCDE1234F"

	// Trade 1: ₹4,000 (Turnover: ₹4,000 <= ₹10,000) -> 0% TDS
	rec1 := engine.ComputeAndDeductTDS("TX-001", pan, 4000.0)
	if rec1.TDSRatePct != 0.0 || rec1.TDSAmountINR != 0.0 {
		t.Fatalf("expected 0 TDS on trade 1 under threshold, got rate %.2f amt %.2f", rec1.TDSRatePct, rec1.TDSAmountINR)
	}
	if rec1.NetProceedsINR != 4000.0 {
		t.Fatalf("expected 100%% net proceeds, got %.2f", rec1.NetProceedsINR)
	}

	// Trade 2: ₹5,000 (Cumulative: ₹9,000 <= ₹10,000) -> 0% TDS
	rec2 := engine.ComputeAndDeductTDS("TX-002", pan, 5000.0)
	if rec2.TDSAmountINR != 0.0 {
		t.Fatalf("expected 0 TDS on trade 2 under threshold, got %.2f", rec2.TDSAmountINR)
	}

	// Trade 3: ₹6,000 (Cumulative: ₹15,000 > ₹10,000) -> 1% TDS on ₹6,000 = ₹60
	rec3 := engine.ComputeAndDeductTDS("TX-003", pan, 6000.0)
	if rec3.TDSRatePct != 1.0 || rec3.TDSAmountINR != 60.0 {
		t.Fatalf("expected 1%% TDS (₹60) on trade 3, got rate %.2f amt %.2f", rec3.TDSRatePct, rec3.TDSAmountINR)
	}
	if rec3.NetProceedsINR != 5940.0 {
		t.Fatalf("expected net proceeds ₹5940, got %.2f", rec3.NetProceedsINR)
	}

	// Trade 4: ₹100,000 -> 1% TDS = ₹1000
	rec4 := engine.ComputeAndDeductTDS("TX-004", pan, 100000.0)
	if rec4.TDSAmountINR != 1000.0 || rec4.NetProceedsINR != 99000.0 {
		t.Fatalf("expected ₹1000 TDS on ₹100k, got %.2f", rec4.TDSAmountINR)
	}
}

// TestSection194S_SpecifiedPerson_Threshold tests the higher ₹50,000 threshold for specified persons
func TestSection194S_SpecifiedPerson_Threshold(t *testing.T) {
	engine := NewSection194STaxEngine()
	pan := "SPEC01234F"
	engine.SetSpecifiedPerson(pan, true)

	// Trade 1: ₹35,000 (Turnover: ₹35,000 <= ₹50,000) -> 0% TDS
	rec1 := engine.ComputeAndDeductTDS("TX-SPEC-1", pan, 35000.0)
	if rec1.TDSAmountINR != 0.0 {
		t.Fatalf("expected 0 TDS for specified person under 50k, got %.2f", rec1.TDSAmountINR)
	}
	if !rec1.IsSpecifiedPerson {
		t.Fatalf("expected IsSpecifiedPerson to be true")
	}

	// Trade 2: ₹20,000 (Cumulative: ₹55,000 > ₹50,000) -> 1% TDS on ₹20,000 = ₹200
	rec2 := engine.ComputeAndDeductTDS("TX-SPEC-2", pan, 20000.0)
	if rec2.TDSAmountINR != 200.0 {
		t.Fatalf("expected ₹200 TDS on threshold crossing, got %.2f", rec2.TDSAmountINR)
	}
}

// TestSection194S_Section206AB_PenalRate tests the 5% penal rate for non-compliant ITR filers
func TestSection194S_Section206AB_PenalRate(t *testing.T) {
	engine := NewSection194STaxEngine()
	pan := "PENAL1234K"
	engine.SetSection206ABStatus(pan, true)

	// Seed turnover to exceed threshold
	engine.ComputeAndDeductTDS("TX-INIT", pan, 20000.0)

	// Next trade: ₹50,000 consideration -> 5% penal rate = ₹2,500
	rec := engine.ComputeAndDeductTDS("TX-PENAL", pan, 50000.0)
	if rec.TDSRatePct != 5.0 || rec.TDSAmountINR != 2500.0 {
		t.Fatalf("expected 5%% penal TDS (₹2500) under Section 206AB, got rate %.2f amt %.2f", rec.TDSRatePct, rec.TDSAmountINR)
	}
	if rec.NetProceedsINR != 47500.0 {
		t.Fatalf("expected net ₹47,500, got %.2f", rec.NetProceedsINR)
	}
}

// TestSection194S_CryptoToCrypto_DualTDS tests dual leg deduction on VDA-for-VDA trades
func TestSection194S_CryptoToCrypto_DualTDS(t *testing.T) {
	engine := NewSection194STaxEngine()
	panA := "BUYER1234A"
	panB := "SELLER1234B"

	// Ensure both have exceeded threshold
	engine.ComputeAndDeductTDS("TX-A-INIT", panA, 15000.0)
	engine.ComputeAndDeductTDS("TX-B-INIT", panB, 15000.0)

	tradeValINR := 50000.0
	recA, recB := engine.ComputeCryptoToCryptoTDS("TRADE-C2C-99", panA, panB, tradeValINR)

	if recA.TDSAmountINR != 500.0 || recB.TDSAmountINR != 500.0 {
		t.Fatalf("expected 1%% TDS on both legs (₹500 each), got A: %.2f, B: %.2f", recA.TDSAmountINR, recB.TDSAmountINR)
	}
	if recA.BuyerPAN != panB || recB.BuyerPAN != panA {
		t.Fatalf("expected cross-counterparty PAN linkage")
	}
}

// TestSection194S_Challan26QE_And_Form26Q verifies statutory filing generation
func TestSection194S_Challan26QE_And_Form26Q(t *testing.T) {
	engine := NewSection194STaxEngine()
	pan := "AAACC1234C" // Company PAN (4th character 'C')

	engine.ComputeAndDeductTDS("TX-001", pan, 50000.0) // 1% = 500
	rec2 := engine.ComputeAndDeductTDS("TX-002", pan, 50000.0) // 1% = 500

	challan, err := engine.GenerateChallan26QE([]*TDSDeductionRecord{rec2})
	if err != nil {
		t.Fatalf("unexpected error generating challan: %v", err)
	}

	if challan.MajorHead != "0020" { // Company Major Head
		t.Fatalf("expected MajorHead 0020 for company PAN, got %s", challan.MajorHead)
	}
	if challan.MinorHead != "800" { // Section 194S VDA
		t.Fatalf("expected MinorHead 800, got %s", challan.MinorHead)
	}
	if challan.BSRCode != "0510304" {
		t.Fatalf("expected BSR 0510304, got %s", challan.BSRCode)
	}

	fy, q := GetCurrentQuarterAndFY(time.Now().UTC())
	form26q := engine.GenerateForm26QReport(fy, q)
	if form26q.TotalRecords < 2 {
		t.Fatalf("expected at least 2 records in Form 26Q, got %d", form26q.TotalRecords)
	}
	if form26q.FilingHash == "" {
		t.Fatalf("expected non-empty filing hash")
	}

	form16A, err := engine.GenerateForm16A(pan, fy, q)
	if err != nil {
		t.Fatalf("unexpected error generating Form 16A: %v", err)
	}
	if form16A.TotalTaxDeducted != 1000.0 {
		t.Fatalf("expected ₹1000 total tax deducted in Form 16A, got %.2f", form16A.TotalTaxDeducted)
	}
}

// TestSection115BBH_StrictNoLossSetOff verifies that losses CANNOT offset gains
func TestSection115BBH_StrictNoLossSetOff(t *testing.T) {
	now := time.Now().UTC()
	fy, _ := GetCurrentQuarterAndFY(now)

	trades := []LotDepletionRecord{
		{
			LotID:             "LOT-BTC-1",
			DepletedQuantity:  0.5,
			CostOfAcquisition: 200000.0,
			SaleProceedsINR:   300000.0,
			RealizedGainINR:   100000.0, // Profit: +₹1,00,000
			TaxableGainINR:    100000.0,
			TaxPayableINR:     31200.0,  // 31.2%
			AcquiredAt:        now.Add(-48 * time.Hour),
		},
		{
			LotID:             "LOT-ETH-1",
			DepletedQuantity:  2.0,
			CostOfAcquisition: 100000.0,
			SaleProceedsINR:   60000.0,
			RealizedGainINR:   -40000.0, // Loss: -₹40,000
			TaxableGainINR:    0.0,      // No loss set-off
			TaxPayableINR:     0.0,
			AcquiredAt:        now.Add(-24 * time.Hour),
		},
	}

	summary := GenerateAnnualVDASummary(fy, "USR-100", "ABCDE1234F", trades, 1500.0)

	// INVARIANT 1: Realized Gains = ₹1,00,000, Realized Losses = ₹40,000
	if summary.TotalRealizedGainsINR != 100000.0 {
		t.Fatalf("expected total realized gains ₹100,000, got %.2f", summary.TotalRealizedGainsINR)
	}
	if summary.TotalRealizedLossesINR != 40000.0 {
		t.Fatalf("expected total realized losses ₹40,000, got %.2f", summary.TotalRealizedLossesINR)
	}

	// CRITICAL INVARIANT 2: Taxable income under Section 115BBH is ₹100,000 (NOT ₹60,000)
	if summary.TaxableIncome115BBHINR != 100000.0 {
		t.Fatalf("VIOLATION OF SECTION 115BBH: Taxable income must be ₹100,000 with NO loss set-off, got %.2f", summary.TaxableIncome115BBHINR)
	}

	// INVARIANT 3: 30% Base Tax (₹30,000) + 4% Cess (₹1,200) = ₹31,200
	if summary.BaseTax30PctINR != 30000.0 {
		t.Fatalf("expected base tax ₹30,000, got %.2f", summary.BaseTax30PctINR)
	}
	if summary.HealthEducationCess4PctINR != 1200.0 {
		t.Fatalf("expected cess ₹1,200, got %.2f", summary.HealthEducationCess4PctINR)
	}
	if summary.TotalTaxLiabilityINR != 31200.0 {
		t.Fatalf("expected total liability ₹31,200, got %.2f", summary.TotalTaxLiabilityINR)
	}

	// INVARIANT 4: Loss carried forward is strictly 0.00
	if summary.LossCarriedForwardINR != 0.0 {
		t.Fatalf("VIOLATION OF SECTION 115BBH(2)(b): Loss carried forward must be 0.00, got %.2f", summary.LossCarriedForwardINR)
	}

	// INVARIANT 5: Disallowed expenses must be flagged
	if summary.DisallowedExpensesINR != 1500.0 {
		t.Fatalf("expected ₹1500 in disallowed expenses, got %.2f", summary.DisallowedExpensesINR)
	}
}

// TestSection115BBH_FIFOTaxLotMatching verifies multi-lot FIFO queue accounting
func TestSection115BBH_FIFOTaxLotMatching(t *testing.T) {
	engine := NewFIFOTaxEngine()
	userID := "USR-TRADER"
	pan := "ABCDE1234F"
	symbol := "BTC"

	t0 := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	t1 := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	// Acquisition 1: 0.50 BTC @ ₹50,00,000/BTC -> Cost ₹25,00,000
	engine.AddAcquisitionLot(userID, symbol, 0.50, 5000000.0, t0)
	// Acquisition 2: 0.50 BTC @ ₹60,00,000/BTC -> Cost ₹30,00,000
	engine.AddAcquisitionLot(userID, symbol, 0.50, 6000000.0, t1)

	// Disposal: Sell 0.70 BTC @ ₹70,00,000/BTC -> Proceeds ₹49,00,000
	// FIFO matches:
	// Slice 1: 0.50 BTC from Lot 1 (Cost: ₹25L, Proceeds: ₹35L, Gain: ₹10L)
	// Slice 2: 0.20 BTC from Lot 2 (Cost: 0.20 * 60L = ₹12L, Proceeds: 0.20 * 70L = ₹14L, Gain: ₹2L)
	// Total realized gain: ₹12,00,000. Tax payable (31.2%): ₹3,74,400.
	// Remaining in Lot 2: 0.30 BTC.
	depletions, totalGain, taxPayable, err := engine.ProcessDisposal(userID, pan, symbol, 0.70, 7000000.0, t2)
	if err != nil {
		t.Fatalf("unexpected disposal error: %v", err)
	}

	if len(depletions) != 2 {
		t.Fatalf("expected 2 depleted slices from FIFO, got %d", len(depletions))
	}

	if depletions[0].DepletedQuantity != 0.50 || depletions[0].CostPerUnitINR != 5000000.0 {
		t.Fatalf("slice 0 did not match Lot 1 correctly")
	}
	if depletions[1].DepletedQuantity != 0.20 || depletions[1].CostPerUnitINR != 6000000.0 {
		t.Fatalf("slice 1 did not match Lot 2 correctly")
	}

	if totalGain != 1200000.0 {
		t.Fatalf("expected total gain ₹12,00,000, got %.2f", totalGain)
	}
	expectedTax := 1200000.0 * 0.312
	if taxPayable != expectedTax {
		t.Fatalf("expected tax ₹%.2f, got %.2f", expectedTax, taxPayable)
	}

	// Verify remaining inventory in Lot 2 is exactly 0.30 BTC
	userLots := engine.userLots[userID][symbol]
	if userLots[0].RemainingQuantity != 0.0 {
		t.Fatalf("expected Lot 1 fully depleted, got %.4f", userLots[0].RemainingQuantity)
	}
	if userLots[1].RemainingQuantity != 0.30 {
		t.Fatalf("expected Lot 2 remaining 0.30 BTC, got %.4f", userLots[1].RemainingQuantity)
	}
}

// TestContractNote_CBDTRoundingAndGST verifies contract note calculation and signature
func TestContractNote_CBDTRoundingAndGST(t *testing.T) {
	// Trade 1: Sell 0.1 BTC @ ₹60,00,000 = ₹600,000 (VDA) -> 1% TDS = ₹6,000
	item1 := BuildContractNoteItem("TRD-1", "09:30:00", "09:30:01", "BTC-INR", "SELL", 0.1, 6000000.0, true)
	if item1.TDS194S != 6000.0 {
		t.Fatalf("expected 1%% TDS ₹6000 on VDA sell, got %.2f", item1.TDS194S)
	}
	if item1.PlatformFee != 0.0 {
		t.Fatalf("expected zero platform fee under policy, got %.2f", item1.PlatformFee)
	}
	if item1.NetSettled != 594000.0 {
		t.Fatalf("expected net settled ₹594,000, got %.2f", item1.NetSettled)
	}

	// Trade 2: Sell 100 shares RELIANCE @ ₹2950.50 = ₹295,050 (Equity) -> STT 0.1% = ₹295 (rounded)
	item2 := BuildContractNoteItem("TRD-2", "10:15:00", "10:15:01", "RELIANCE.EQ", "SELL", 100, 2950.50, false)
	if item2.STT != 295.0 {
		t.Fatalf("expected CBDT rounded STT ₹295, got %.2f", item2.STT)
	}

	note := GenerateContractNote("CN-2026-0001", "ABCDE1234F", "GROWWW-CL-001", []ContractNoteItem{item1, item2})
	if note.TotalGross != 895050.0 {
		t.Fatalf("expected total gross ₹895,050, got %.2f", note.TotalGross)
	}
	if !strings.HasPrefix(note.DigitalSignHash, "0x") {
		t.Fatalf("expected digital sign starting with 0x, got %s", note.DigitalSignHash)
	}

	jsonBytes, err := note.SerializeToJSON()
	if err != nil || len(jsonBytes) == 0 {
		t.Fatalf("failed to serialize contract note: %v", err)
	}
}

// TestCompute115BBHTax_Direct tests direct calculation function for single transaction
func TestCompute115BBHTax_Direct(t *testing.T) {
	// Gain: ₹10,000 -> 30% = ₹3,000 + 4% cess = ₹120 -> ₹3,120
	gainRes := Compute115BBHTax("ASSESS-1", "USR-1", "ABCDE1234F", "BTC", 50000.0, 40000.0)
	if gainRes.RealizedGainINR != 10000.0 || gainRes.TotalTaxPayableINR != 3120.0 {
		t.Fatalf("expected gain ₹10k, tax ₹3120, got gain %.2f tax %.2f", gainRes.RealizedGainINR, gainRes.TotalTaxPayableINR)
	}
	if gainRes.EffectiveTaxRatePct != 31.20 {
		t.Fatalf("expected effective rate 31.20%%, got %.2f", gainRes.EffectiveTaxRatePct)
	}

	// Loss: -₹5,000 -> Tax = 0.0
	lossRes := Compute115BBHTax("ASSESS-2", "USR-1", "ABCDE1234F", "BTC", 35000.0, 40000.0)
	if lossRes.RealizedGainINR != -5000.0 || lossRes.TotalTaxPayableINR != 0.0 {
		t.Fatalf("expected loss ₹-5000, tax 0.0, got gain %.2f tax %.2f", lossRes.RealizedGainINR, lossRes.TotalTaxPayableINR)
	}
}

// TestFIFOTaxEngine_ErrorCases tests error conditions in disposal
func TestFIFOTaxEngine_ErrorCases(t *testing.T) {
	engine := NewFIFOTaxEngine()
	now := time.Now().UTC()

	// 1. Non-positive quantity
	_, _, _, err := engine.ProcessDisposal("U1", "ABCDE1234F", "BTC", 0.0, 50000.0, now)
	if err == nil {
		t.Fatalf("expected error on zero quantity")
	}

	// 2. Unknown symbol / no lots
	_, _, _, err = engine.ProcessDisposal("U1", "ABCDE1234F", "UNKNOWN", 1.0, 50000.0, now)
	if err == nil {
		t.Fatalf("expected error on unknown symbol")
	}

	// 3. Insufficient inventory
	engine.AddAcquisitionLot("U1", "ETH", 0.5, 200000.0, now)
	_, _, _, err = engine.ProcessDisposal("U1", "ABCDE1234F", "ETH", 1.0, 250000.0, now)
	if err == nil {
		t.Fatalf("expected error on insufficient inventory")
	}
}

// TestSection194S_ErrorCases tests error handling for Challan and Form16A
func TestSection194S_ErrorCases(t *testing.T) {
	engine := NewSection194STaxEngine()

	// Empty records challan
	_, err := engine.GenerateChallan26QE([]*TDSDeductionRecord{})
	if err == nil {
		t.Fatalf("expected error on empty records challan")
	}

	// Non-existent PAN in Form 16A
	_, err = engine.GenerateForm16A("NONEX1234F", "2026-2027", "Q1")
	if err == nil {
		t.Fatalf("expected error for non-existent PAN")
	}
}

// TestGetCurrentQuarterAndFY_AllQuarters verifies all 4 quarters and FY computation
func TestGetCurrentQuarterAndFY_AllQuarters(t *testing.T) {
	// Q1: May
	fy1, q1 := GetCurrentQuarterAndFY(time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC))
	if fy1 != "2026-2027" || q1 != "Q1" {
		t.Fatalf("expected 2026-2027 Q1, got %s %s", fy1, q1)
	}

	// Q2: August
	fy2, q2 := GetCurrentQuarterAndFY(time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
	if fy2 != "2026-2027" || q2 != "Q2" {
		t.Fatalf("expected 2026-2027 Q2, got %s %s", fy2, q2)
	}

	// Q3: November
	fy3, q3 := GetCurrentQuarterAndFY(time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC))
	if fy3 != "2026-2027" || q3 != "Q3" {
		t.Fatalf("expected 2026-2027 Q3, got %s %s", fy3, q3)
	}

	// Q4: February (next calendar year, same FY)
	fy4, q4 := GetCurrentQuarterAndFY(time.Date(2027, 2, 10, 0, 0, 0, 0, time.UTC))
	if fy4 != "2026-2027" || q4 != "Q4" {
		t.Fatalf("expected 2026-2027 Q4, got %s %s", fy4, q4)
	}
}

// TestContractNote_BuySide verifies BUY side contract note calculations for Equity and VDA
func TestContractNote_BuySide(t *testing.T) {
	// Equity BUY: ₹100,000 -> STT 0.1% = ₹100, Stamp Duty 0.015% = ₹15 -> Net = ₹100,115
	eqBuy := BuildContractNoteItem("BUY-EQ", "10:00:00", "10:00:01", "TCS.EQ", "BUY", 25, 4000.0, false)
	if eqBuy.STT != 100.0 || eqBuy.StampDuty != 15.0 || eqBuy.NetSettled != 100115.0 {
		t.Fatalf("unexpected equity buy net: %.2f", eqBuy.NetSettled)
	}

	// VDA BUY: ₹50,000 -> No STT, No TDS on buyer in INR onramp -> Net = ₹50,000
	vdaBuy := BuildContractNoteItem("BUY-VDA", "10:05:00", "10:05:01", "ETH-INR", "BUY", 0.25, 200000.0, true)
	if vdaBuy.TDS194S != 0.0 || vdaBuy.NetSettled != 50000.0 {
		t.Fatalf("unexpected VDA buy net: %.2f", vdaBuy.NetSettled)
	}
}
