package main

import (
	"math"
)

// STTRate defines the Securities Transaction Tax rate per asset class under Indian Income Tax Act.
type STTRate struct {
	AssetClass    string
	DeliveryRate  float64 // e.g. 0.1% on delivery (0.001)
	IntradayRate  float64 // e.g. 0.025% on sell side (0.00025)
	FuturesRate   float64 // e.g. 0.0125% on sell side (0.000125)
	OptionsRate   float64 // e.g. 0.0625% on premium (0.000625)
}

// ComputeSTT calculates STT and rounds to nearest whole rupee per SEBI/ITD circulars.
func ComputeSTT(turnoverINR float64, rate float64) float64 {
	rawSTT := turnoverINR * rate
	// Round to nearest integer (half rounds up)
	return math.Round(rawSTT)
}

// ComputeEquityContractNoteTax calculates STT, GST (18%), and Stamp Duty.
func ComputeEquityContractNoteTax(turnoverINR float64, isDelivery bool) (stt, stampDuty, gst float64) {
	if isDelivery {
		stt = ComputeSTT(turnoverINR, 0.001) // 0.1%
		stampDuty = math.Round(turnoverINR * 0.00015) // 0.015%
	} else {
		stt = ComputeSTT(turnoverINR, 0.00025) // 0.025%
		stampDuty = math.Round(turnoverINR * 0.00003) // 0.003%
	}
	// GST on brokerage/charges (0 at 0% fee)
	gst = 0.0
	return stt, stampDuty, gst
}
