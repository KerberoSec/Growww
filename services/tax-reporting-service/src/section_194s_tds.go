package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type TDSDeductionRecord struct {
	RecordID        string    `json:"record_id"`
	TradeID         string    `json:"trade_id"`
	SellerPAN       string    `json:"seller_pan"`
	GrossAmountINR  float64   `json:"gross_amount_inr"`
	TDSRatePct      float64   `json:"tds_rate_pct"` // 1.00%
	TDSAmountINR    float64   `json:"tds_amount_inr"`
	NetProceedsINR  float64   `json:"net_proceeds_inr"`
	ChallanBSRCode  string    `json:"challan_bsr_code"`
	TDSQuarter      string    `json:"tds_quarter"` // e.g. Q2-2026
	DeductedAt      time.Time `json:"deducted_at"`
}

type Section194STaxEngine struct {
	mu           sync.Mutex
	records      map[string]*TDSDeductionRecord
	panTurnover  map[string]float64
	exemptionCap float64 // ₹50,000 for specified persons, ₹10,000 general
}

func NewSection194STaxEngine() *Section194STaxEngine {
	return &Section194STaxEngine{
		records:      make(map[string]*TDSDeductionRecord),
		panTurnover:  make(map[string]float64),
		exemptionCap: 10000.0,
	}
}

// ComputeAndDeductTDS calculates 1% TDS under Section 194S on VDA transfer consideration
func (e *Section194STaxEngine) ComputeAndDeductTDS(tradeID, sellerPAN string, grossProceedsINR float64) *TDSDeductionRecord {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.panTurnover[sellerPAN] += grossProceedsINR

	// Compute 1% TDS rounded to nearest integer as required by IT Act
	tds := math.Round(grossProceedsINR * 0.01)
	net := grossProceedsINR - tds

	rec := &TDSDeductionRecord{
		RecordID:       fmt.Sprintf("TDS-%s", tradeID),
		TradeID:        tradeID,
		SellerPAN:      sellerPAN,
		GrossAmountINR: grossProceedsINR,
		TDSRatePct:     1.00,
		TDSAmountINR:   tds,
		NetProceedsINR: net,
		TDSQuarter:     "Q2-2026",
		DeductedAt:     time.Now().UTC(),
	}

	e.records[rec.RecordID] = rec
	fmt.Printf("[Section 194S] Deducted 1%% TDS (₹%.2f) on trade %s for PAN %s\n", tds, tradeID, sellerPAN)
	return rec
}
