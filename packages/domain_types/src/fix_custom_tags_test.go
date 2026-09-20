package src

import (
	"errors"
	"testing"
)

func TestFIXCustomTagEncodingAndParsing(t *testing.T) {
	msg := NewFIXMessage()
	msg.Set(35, "D") // NewOrderSingle
	msg.Set(11, "CLORD-901")
	msg.Set(55, "TCS")
	msg.Set(54, "1") // Buy
	msg.Set(38, "500")

	// Custom Tags
	msg.Set(TagExchangeOrderID, "NBSE-ORD-887766")
	msg.SetInt(TagAlgoStrategyID, 1) // VWAP
	msg.SetInt(TagExecutionLatencyNanos, 450)
	msg.Set(TagSebiCategory, "FPI_CAT1")

	encoded := msg.Encode()
	if encoded == "" {
		t.Fatalf("empty encoded FIX string")
	}

	parsed, err := ParseFIXMessage(encoded)
	if err != nil {
		t.Fatalf("failed parsing FIX message: %v", err)
	}

	// Verify standard tags
	val, ok := parsed.Get(55)
	if !ok || val != "TCS" {
		t.Fatalf("symbol tag 55 mismatch: %s", val)
	}

	// Verify custom tags
	exOrd, ok := parsed.Get(TagExchangeOrderID)
	if !ok || exOrd != "NBSE-ORD-887766" {
		t.Fatalf("custom tag %d mismatch: %s", TagExchangeOrderID, exOrd)
	}

	algoID, ok := parsed.Get(TagAlgoStrategyID)
	if !ok || algoID != "1" {
		t.Fatalf("custom tag %d mismatch: %s", TagAlgoStrategyID, algoID)
	}

	sebiCat, ok := parsed.Get(TagSebiCategory)
	if !ok || sebiCat != "FPI_CAT1" {
		t.Fatalf("custom tag %d mismatch: %s", TagSebiCategory, sebiCat)
	}

	// Tampered checksum test: flip digits just before terminal SOH to make checksum wrong
	// encoded ends with ...10=XXX\x01, we change only the digit portion
	chkStart := len(encoded) - 4 // last 3 digits + SOH
	wrongChecksum := "255"
	// Make sure it doesn't accidentally match
	correctDigits := encoded[chkStart : len(encoded)-1]
	if correctDigits == wrongChecksum {
		wrongChecksum = "000"
	}
	tampered := encoded[:chkStart] + wrongChecksum + SOH
	_, err = ParseFIXMessage(tampered)
	if err == nil || !errors.Is(err, ErrFIXInvalidChecksum) {
		t.Fatalf("expected ErrFIXInvalidChecksum on tampered FIX checksum, got %v", err)
	}
}
