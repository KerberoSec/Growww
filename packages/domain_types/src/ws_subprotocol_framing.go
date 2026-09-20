package src

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
)

const (
	WSFrameMagicByte    byte   = 0xAA
	WSFixedHeaderLength int    = 23
	WSTrailerLength     int    = 4
	WSSubprotocolName   string = "nbse.marketdata.v1.binary"
)

// WSOpcode represents framing message types for high-frequency market data streaming.
type WSOpcode byte

const (
	OpcodeSubscribe       WSOpcode = 0x01
	OpcodeUnsubscribe     WSOpcode = 0x02
	OpcodeHeartbeatPing   WSOpcode = 0x03
	OpcodeHeartbeatPong   WSOpcode = 0x04
	OpcodeL2DepthSnapshot WSOpcode = 0x05
	OpcodeL2DepthDelta    WSOpcode = 0x06
	OpcodeTradeTick       WSOpcode = 0x07
)

var (
	ErrWSInvalidMagicByte = errors.New("ws_framing: invalid magic byte")
	ErrWSTruncatedFrame   = errors.New("ws_framing: frame length shorter than fixed header")
	ErrWSChecksumMismatch = errors.New("ws_framing: crc32 checksum verification failed")
	ErrWSPayloadTooLarge  = errors.New("ws_framing: payload exceeds max permissible 16MB")
)

// WSBinaryFrame represents a decoded ultra-low latency WebSocket frame.
type WSBinaryFrame struct {
	Opcode         WSOpcode
	Flags          byte
	SequenceNumber uint64
	TimestampNanos uint64
	Payload        []byte
}

// EncodeWSBinaryFrame serializes frame into raw wire bytes with CRC32 integrity trailer.
func EncodeWSBinaryFrame(frame WSBinaryFrame) ([]byte, error) {
	payloadLen := len(frame.Payload)
	if payloadLen > 16*1024*1024 {
		return nil, ErrWSPayloadTooLarge
	}

	totalLen := WSFixedHeaderLength + payloadLen + WSTrailerLength
	buf := make([]byte, totalLen)

	buf[0] = WSFrameMagicByte
	buf[1] = byte(frame.Opcode)
	buf[2] = frame.Flags
	binary.BigEndian.PutUint64(buf[3:11], frame.SequenceNumber)
	binary.BigEndian.PutUint64(buf[11:19], frame.TimestampNanos)
	binary.BigEndian.PutUint32(buf[19:23], uint32(payloadLen))

	if payloadLen > 0 {
		copy(buf[23:23+payloadLen], frame.Payload)
	}

	// Compute CRC32 over header + payload
	checksum := crc32.ChecksumIEEE(buf[:23+payloadLen])
	binary.BigEndian.PutUint32(buf[23+payloadLen:totalLen], checksum)

	return buf, nil
}

// DecodeWSBinaryFrame decodes raw wire bytes into frame and validates CRC32.
func DecodeWSBinaryFrame(raw []byte) (*WSBinaryFrame, error) {
	if len(raw) < WSFixedHeaderLength+WSTrailerLength {
		return nil, ErrWSTruncatedFrame
	}

	if raw[0] != WSFrameMagicByte {
		return nil, fmt.Errorf("%w: got 0x%02x, expected 0x%02x", ErrWSInvalidMagicByte, raw[0], WSFrameMagicByte)
	}

	opcode := WSOpcode(raw[1])
	flags := raw[2]
	seq := binary.BigEndian.Uint64(raw[3:11])
	timestamp := binary.BigEndian.Uint64(raw[11:19])
	payloadLen := int(binary.BigEndian.Uint32(raw[19:23]))

	expectedTotal := WSFixedHeaderLength + payloadLen + WSTrailerLength
	if len(raw) < expectedTotal {
		return nil, fmt.Errorf("%w: expected %d bytes, got %d", ErrWSTruncatedFrame, expectedTotal, len(raw))
	}

	payload := make([]byte, payloadLen)
	if payloadLen > 0 {
		copy(payload, raw[23:23+payloadLen])
	}

	// Verify CRC32
	storedChecksum := binary.BigEndian.Uint32(raw[23+payloadLen : expectedTotal])
	computedChecksum := crc32.ChecksumIEEE(raw[:23+payloadLen])
	if storedChecksum != computedChecksum {
		return nil, fmt.Errorf("%w: stored=0x%08x computed=0x%08x", ErrWSChecksumMismatch, storedChecksum, computedChecksum)
	}

	return &WSBinaryFrame{
		Opcode:         opcode,
		Flags:          flags,
		SequenceNumber: seq,
		TimestampNanos: timestamp,
		Payload:        payload,
	}, nil
}
