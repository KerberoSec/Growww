package src

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestWSBinaryFramingRoundtrip(t *testing.T) {
	payload := []byte(`{"symbol":"RELIANCE","bid":2950.5,"ask":2950.75}`)
	nanos := uint64(time.Now().UnixNano())

	frame := WSBinaryFrame{
		Opcode:         OpcodeL2DepthSnapshot,
		Flags:          0x01, // compressed
		SequenceNumber: 104523,
		TimestampNanos: nanos,
		Payload:        payload,
	}

	raw, err := EncodeWSBinaryFrame(frame)
	if err != nil {
		t.Fatalf("failed encoding frame: %v", err)
	}

	decoded, err := DecodeWSBinaryFrame(raw)
	if err != nil {
		t.Fatalf("failed decoding frame: %v", err)
	}

	if decoded.Opcode != OpcodeL2DepthSnapshot {
		t.Fatalf("opcode mismatch: expected %d, got %d", OpcodeL2DepthSnapshot, decoded.Opcode)
	}
	if decoded.Flags != 0x01 {
		t.Fatalf("flags mismatch: expected 0x01, got 0x%02x", decoded.Flags)
	}
	if decoded.SequenceNumber != 104523 {
		t.Fatalf("sequence number mismatch: expected 104523, got %d", decoded.SequenceNumber)
	}
	if decoded.TimestampNanos != nanos {
		t.Fatalf("timestamp mismatch: expected %d, got %d", nanos, decoded.TimestampNanos)
	}
	if !bytes.Equal(decoded.Payload, payload) {
		t.Fatalf("payload mismatch")
	}

	// Tamper with payload byte to trigger CRC32 failure
	rawTampered := make([]byte, len(raw))
	copy(rawTampered, raw)
	rawTampered[25] ^= 0xFF

	_, err = DecodeWSBinaryFrame(rawTampered)
	if err == nil || !errors.Is(err, ErrWSChecksumMismatch) {
		t.Fatalf("expected ErrWSChecksumMismatch on tampered frame, got %v", err)
	}

	// Invalid magic byte test
	rawBadMagic := make([]byte, len(raw))
	copy(rawBadMagic, raw)
	rawBadMagic[0] = 0x00
	_, err = DecodeWSBinaryFrame(rawBadMagic)
	if err == nil || !errors.Is(err, ErrWSInvalidMagicByte) {
		t.Fatalf("expected ErrWSInvalidMagicByte, got %v", err)
	}
}
