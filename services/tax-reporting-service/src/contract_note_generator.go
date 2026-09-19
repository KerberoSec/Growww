package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// RoundToRupee implements CBDT / Indian IT Act statutory rounding to the nearest integer Rupee
func RoundToRupee(val float64) float64 {
	return math.Round(val)
}

// ContractNoteItem represents an individual trade settlement entry on the contract note
type ContractNoteItem struct {
	TradeID        string  `json:"trade_id"`
	OrderTime      string  `json:"order_time"`
	TradeTime      string  `json:"trade_time"`
	Symbol         string  `json:"symbol"`
	Side           string  `json:"side"` // BUY or SELL
	Quantity       float64 `json:"quantity"`
	Price          float64 `json:"price"`
	GrossTotal     float64 `json:"gross_total"`
	STT            float64 `json:"stt"`             // Securities Transaction Tax (CBDT rounded)
	StampDuty      float64 `json:"stamp_duty"`      // Indian Stamp Act (rounded)
	TDS194S        float64 `json:"tds_194s"`        // 1% TDS on VDA sell (rounded)
	GST            float64 `json:"gst"`             // 18% GST on platform fee
	PlatformFee    float64 `json:"platform_fee"`    // 0.00% universal policy
	BrokerageCharges float64 `json:"brokerage_charges"` // 0.00 INR
	NetSettled     float64 `json:"net_settled"`
}

// ConsolidatedContractNote represents the statutory daily contract note compliant with SEBI and CBDT norms
type ConsolidatedContractNote struct {
	NoteNumber         string             `json:"note_number"`
	TradeDate          string             `json:"trade_date"`
	InvestorPAN        string             `json:"investor_pan"`
	InvestorName       string             `json:"investor_name"` // Retained in off-chain PDF/records only
	UCC                string             `json:"ucc"`           // Unique Client Code
	ExchangeCode       string             `json:"exchange_code"` // "NBSE"
	Trades             []ContractNoteItem `json:"trades"`
	TotalGross         float64            `json:"total_gross"`
	TotalTDS194S       float64            `json:"total_tds_194s"`
	TotalSTT           float64            `json:"total_stt"`
	TotalStampDuty     float64            `json:"total_stamp_duty"`
	TotalGST           float64            `json:"total_gst"`
	TotalPlatformFee   float64            `json:"total_platform_fee"`
	TotalNetPayout     float64            `json:"total_net_payout"`
	DigitalSignHash    string             `json:"digital_signature_hash"`
	GeneratedAt        time.Time          `json:"generated_at"`
}

// BuildContractNoteItem calculates statutory taxes and net settled consideration for a trade
func BuildContractNoteItem(tradeID, orderTime, tradeTime, symbol, side string, qty, price float64, isVDA bool) ContractNoteItem {
	gross := qty * price
	var stt, stampDuty, tds, platformFee, gst, net float64

	platformFee = 0.0 // 0.00% Zero platform fee policy
	gst = RoundToRupee(platformFee * 0.18)

	if side == "SELL" {
		if isVDA {
			// Virtual Digital Asset transfer: 1% TDS under Section 194S
			tds = RoundToRupee(gross * 0.01)
			stampDuty = 0.0
			stt = 0.0
		} else {
			// Equity delivery: 0.1% STT on delivery (CBDT rounded), 0.015% stamp duty
			stt = RoundToRupee(gross * 0.001)
			stampDuty = math.Round(gross*0.00015*100) / 100
			tds = 0.0
		}
		// For seller: Net = Gross - TDS - STT - StampDuty - GST - Fee
		net = gross - tds - stt - stampDuty - gst - platformFee
	} else {
		// BUY side
		if !isVDA {
			stt = RoundToRupee(gross * 0.001)
			stampDuty = math.Round(gross*0.00015*100) / 100
		}
		tds = 0.0
		// For buyer: Net = Gross + STT + StampDuty + GST + Fee
		net = gross + stt + stampDuty + gst + platformFee
	}

	return ContractNoteItem{
		TradeID:          tradeID,
		OrderTime:        orderTime,
		TradeTime:        tradeTime,
		Symbol:           symbol,
		Side:             side,
		Quantity:         qty,
		Price:            price,
		GrossTotal:       math.Round(gross*100) / 100,
		STT:              stt,
		StampDuty:        stampDuty,
		TDS194S:          tds,
		GST:              gst,
		PlatformFee:      platformFee,
		BrokerageCharges: 0.0,
		NetSettled:       math.Round(net*100) / 100,
	}
}

// GenerateContractNote aggregates items, computes totals, and generates cryptographic digital signature
func GenerateContractNote(noteNo, pan, ucc string, items []ContractNoteItem) *ConsolidatedContractNote {
	var totalGross, totalTDS, totalSTT, totalStamp, totalGST, totalFee, totalNet float64

	for _, it := range items {
		totalGross += it.GrossTotal
		totalTDS += it.TDS194S
		totalSTT += it.STT
		totalStamp += it.StampDuty
		totalGST += it.GST
		totalFee += it.PlatformFee
		totalNet += it.NetSettled
	}

	now := time.Now().UTC()
	tradeDate := now.Format("2006-01-02")

	// Cryptographic SHA-256 signature binding the contract note content
	sigPayload := fmt.Sprintf("%s:%s:%s:%.2f:%.2f:%.2f:%s",
		noteNo, pan, ucc, totalGross, totalTDS, totalNet, tradeDate)
	hash := sha256.Sum256([]byte(sigPayload))
	digitalSign := "0x" + hex.EncodeToString(hash[:])

	note := &ConsolidatedContractNote{
		NoteNumber:       noteNo,
		TradeDate:        tradeDate,
		InvestorPAN:      pan,
		UCC:              ucc,
		ExchangeCode:     "NBSE",
		Trades:           items,
		TotalGross:       math.Round(totalGross*100) / 100,
		TotalTDS194S:     math.Round(totalTDS*100) / 100,
		TotalSTT:         math.Round(totalSTT*100) / 100,
		TotalStampDuty:   math.Round(totalStamp*100) / 100,
		TotalGST:         math.Round(totalGST*100) / 100,
		TotalPlatformFee: math.Round(totalFee*100) / 100,
		TotalNetPayout:   math.Round(totalNet*100) / 100,
		DigitalSignHash:  digitalSign,
		GeneratedAt:      now,
	}

	return note
}

// SerializeToJSON converts the contract note to formatted JSON
func (n *ConsolidatedContractNote) SerializeToJSON() ([]byte, error) {
	return json.MarshalIndent(n, "", "  ")
}
