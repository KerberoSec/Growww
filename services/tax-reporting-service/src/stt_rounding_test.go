package main

import "testing"

func TestComputeSTT_Rounding(t *testing.T) {
	// Turnover: 1,23,456 at 0.1% = 123.456 -> rounds to 123
	stt := ComputeSTT(123456, 0.001)
	if stt != 123 {
		t.Errorf("expected 123, got %f", stt)
	}

	// Turnover: 1,23,600 at 0.1% = 123.600 -> rounds to 124
	stt2 := ComputeSTT(123600, 0.001)
	if stt2 != 124 {
		t.Errorf("expected 124, got %f", stt2)
	}
}

func TestEquityContractNoteTax_Delivery(t *testing.T) {
	stt, stamp, gst := ComputeEquityContractNoteTax(1000000, true) // 10 Lakhs
	if stt != 1000 {
		t.Errorf("expected 1000 STT, got %f", stt)
	}
	if stamp != 150 {
		t.Errorf("expected 150 stamp duty, got %f", stamp)
	}
	if gst != 0 {
		t.Errorf("expected 0 GST on zero-fee exchange, got %f", gst)
	}
}
