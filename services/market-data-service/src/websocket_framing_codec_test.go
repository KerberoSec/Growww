package main

import (
	"bytes"
	"testing"
)

func TestWebSocketFraming_EncodeDecodeRoundTrip(t *testing.T) {
	originalPayload := []byte(`{"event":"depth","symbol":"BTC/USDT","seq":9821}`)
	frame := BinarySubprotocolFrame{
		Version:        WebSocketProtocolV1,
		Type:           FrameTypeMarketDepth,
		Encoding:       EncodingJSON,
		Flags:          0,
		SequenceNumber: 9821,
		Payload:        originalPayload,
	}

	encoded := EncodeFrame(frame)
	if len(encoded) != FixedHeaderSize+len(originalPayload) {
		t.Fatalf("unexpected encoded length: %d", len(encoded))
	}

	decoded, err := DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("failed to decode frame: %v", err)
	}

	if decoded.Version != WebSocketProtocolV1 {
		t.Fatalf("expected version 1, got %d", decoded.Version)
	}
	if decoded.Type != FrameTypeMarketDepth {
		t.Fatalf("expected FrameTypeMarketDepth, got %d", decoded.Type)
	}
	if decoded.Encoding != EncodingJSON {
		t.Fatalf("expected EncodingJSON, got %d", decoded.Encoding)
	}
	if decoded.SequenceNumber != 9821 {
		t.Fatalf("expected sequence 9821, got %d", decoded.SequenceNumber)
	}
	if !bytes.Equal(decoded.Payload, originalPayload) {
		t.Fatalf("payload mismatch: %s vs %s", string(decoded.Payload), string(originalPayload))
	}
}

func TestWebSocketFraming_ZeroPayloadHeartbeat(t *testing.T) {
	ping := BinarySubprotocolFrame{
		Version:        WebSocketProtocolV1,
		Type:           FrameTypePing,
		Encoding:       EncodingProtobuf,
		SequenceNumber: 1,
		Payload:        nil,
	}

	encoded := EncodeFrame(ping)
	decoded, err := DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("failed to decode ping frame: %v", err)
	}

	if decoded.Type != FrameTypePing {
		t.Fatalf("expected ping frame, got %d", decoded.Type)
	}
	if len(decoded.Payload) != 0 {
		t.Fatalf("expected empty payload, got %d bytes", len(decoded.Payload))
	}
}

func TestWebSocketFraming_CorruptedChecksumError(t *testing.T) {
	payload := []byte("institutional_secure_payload")
	frame := BinarySubprotocolFrame{
		Version:        WebSocketProtocolV1,
		Type:           FrameTypeTradeExecution,
		Encoding:       EncodingProtobuf,
		SequenceNumber: 42,
		Payload:        payload,
	}

	encoded := EncodeFrame(frame)

	// Tamper with a payload byte
	encoded[len(encoded)-1] ^= 0xFF

	_, err := DecodeFrame(encoded)
	if err != ErrChecksumMismatch {
		t.Fatalf("expected ErrChecksumMismatch, got %v", err)
	}
}

func TestWebSocketFraming_HeaderValidationErrors(t *testing.T) {
	// Frame too short
	shortBuf := make([]byte, 10)
	if _, err := DecodeFrame(shortBuf); err != ErrFrameTooShort {
		t.Fatalf("expected ErrFrameTooShort, got %v", err)
	}

	// Invalid magic
	badMagic := make([]byte, FixedHeaderSize)
	badMagic[0] = 0xFF
	badMagic[1] = 0xFE
	if _, err := DecodeFrame(badMagic); err != ErrInvalidMagicBytes {
		t.Fatalf("expected ErrInvalidMagicBytes, got %v", err)
	}
}
