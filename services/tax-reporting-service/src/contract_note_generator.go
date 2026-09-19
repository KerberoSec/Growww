package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type ContractNoteItem struct {
	TradeID     string  `json:"trade_id"`
	OrderTime   string  `json:"order_time"`
	TradeTime   string  `json:"trade_time"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	Quantity    float64 `json:"quantity"`
	Price       float64 `json:"price"`
	GrossTotal  float64 `json:"gross_total"`
	STT         float64 `json:"stt"`
	StampDuty   float64 `json:"stamp_duty"`
	TDS194S     float64 `json:"tds_194s"`
	GST         float64 `json:"gst"`
	PlatformFee float64 `json:"platform_fee"` // 0.00% universal policy
	NetSettled  float64 `json:"net_settled"`
}

type ConsolidatedContractNote struct {
	NoteNumber      string             `json:"note_number"`
	TradeDate       string             `json:"trade_date"`
	InvestorPAN     string             `json:"investor_pan"`
	InvestorName    string             `json:"investor_name"` // Retained in off-chain PDF only
	UCC             string             `json:"ucc"`
	Trades          []ContractNoteItem `json:"trades"`
	TotalGross      float64            `json:"total_gross"`
	TotalTDS194S    float64            `json:"total_tds_194s"`
	TotalSTT        float64            `json:"total_stt"`
	TotalNetPayout  float64            `json:"total_net_payout"`
	DigitalSignHash string             `json:"digital_signature_hash"`
	GeneratedAt     time.Time          `json:"generated_at"`
}

func GenerateContractNote(noteNo, pan, ucc string, items []ContractNoteItem) *ConsolidatedContractNote {
	var totalGross, totalTDS, totalSTT, totalNet float64
	for _, it := range items {
		totalGross += it.GrossTotal
		totalTDS += it.TDS194S
		totalSTT += it.STT
		totalNet += it.NetSettled
	}

	note := &ConsolidatedContractNote{
		NoteNumber:      noteNo,
		TradeDate:       time.Now().Format("2006-01-02"),
		InvestorPAN:     pan,
		UCC:             ucc,
		Trades:          items,
		TotalGross:      totalGross,
		TotalTDS194S:    totalTDS,
		TotalSTT:        totalSTT,
		TotalNetPayout:  totalNet,
		DigitalSignHash: "0x" + fmt.Sprintf("%x", time.Now().UnixNano()),
		GeneratedAt:     time.Now().UTC(),
	}

	bytes, _ := json.MarshalIndent(note, "", "  ")
	_ = bytes
	return note
}
