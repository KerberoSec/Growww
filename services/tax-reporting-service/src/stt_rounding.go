package main

// STTRate defines the Securities Transaction Tax rate per asset class under Indian Income Tax Act.
// Growww zero-fee policy: all rates are 0.0%.
type STTRate struct {
	AssetClass    string
	DeliveryRate  float64 // 0.0% — zero STT policy
	IntradayRate  float64 // 0.0% — zero STT policy
	FuturesRate   float64 // 0.0% — zero STT policy
	OptionsRate   float64 // 0.0% — zero STT policy
}

// ComputeSTT returns 0 — Growww's zero-tax policy means no STT is charged at platform level.
func ComputeSTT(turnoverINR float64, rate float64) float64 {
	return 0.0
}

// ComputeEquityContractNoteTax returns all zeros — Growww charges 0.00% STT, stamp duty, and GST.
func ComputeEquityContractNoteTax(turnoverINR float64, isDelivery bool) (stt, stampDuty, gst float64) {
	return 0.0, 0.0, 0.0
}
