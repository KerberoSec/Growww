package main

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

const (
	WebSocketMagicByte0 = 0x47 // 'G'
	WebSocketMagicByte1 = 0x57 // 'W'
	WebSocketProtocolV1 = 0x01

	// HeaderSize: 2 (magic) + 1 (version) + 1 (type) + 1 (encoding) + 1 (flags) + 8 (seq) + 4 (len) + 4 (crc32) = 22 bytes
	FixedHeaderSize = 22
)

type SubprotocolFrameType uint8

const (
	FrameTypePing           SubprotocolFrameType = 0x01
	FrameTypePong           SubprotocolFrameType = 0x02
	FrameTypeSubscribe      SubprotocolFrameType = 0x03
	FrameTypeUnsubscribe    SubprotocolFrameType = 0x04
	FrameTypeMarketDepth    SubprotocolFrameType = 0x05
	FrameTypeTradeExecution SubprotocolFrameType = 0x06
	FrameTypeOrderUpdate    SubprotocolFrameType = 0x07
	FrameTypeError          SubprotocolFrameType = 0x08
)

type SerializationEncoding uint8

const (
	EncodingJSON     SerializationEncoding = 0x01
	EncodingProtobuf SerializationEncoding = 0x02
)

// BinarySubprotocolFrame represents an institutional framed WebSocket message
type BinarySubprotocolFrame struct {
	Version        uint8
	Type           SubprotocolFrameType
	Encoding       SerializationEncoding
	Flags          uint8
	SequenceNumber uint64
	Payload        []byte
}

var (
	ErrFrameTooShort      = errors.New("frame buffer shorter than fixed header size")
	ErrInvalidMagicBytes  = errors.New("invalid websocket subprotocol magic bytes")
	ErrUnsupportedVersion = errors.New("unsupported protocol version")
	ErrChecksumMismatch   = errors.New("frame payload crc32 checksum mismatch")
	ErrPayloadTruncated   = errors.New("frame buffer truncated before payload completion")
)

// EncodeFrame serializes a subprotocol frame into an institutional binary wire format
func EncodeFrame(frame BinarySubprotocolFrame) []byte {
	payloadLen := len(frame.Payload)
	buf := make([]byte, FixedHeaderSize+payloadLen)

	buf[0] = WebSocketMagicByte0
	buf[1] = WebSocketMagicByte1
	buf[2] = frame.Version
	buf[3] = byte(frame.Type)
	buf[4] = byte(frame.Encoding)
	buf[5] = frame.Flags

	binary.BigEndian.PutUint64(buf[6:14], frame.SequenceNumber)
	binary.BigEndian.PutUint32(buf[14:18], uint32(payloadLen))

	// Compute CRC32 of payload
	checksum := crc32.ChecksumIEEE(frame.Payload)
	binary.BigEndian.PutUint32(buf[18:22], checksum)

	if payloadLen > 0 {
		copy(buf[22:], frame.Payload)
	}

	return buf
}

// DecodeFrame parses an institutional binary wire buffer into a BinarySubprotocolFrame
func DecodeFrame(data []byte) (*BinarySubprotocolFrame, error) {
	if len(data) < FixedHeaderSize {
		return nil, ErrFrameTooShort
	}

	if data[0] != WebSocketMagicByte0 || data[1] != WebSocketMagicByte1 {
		return nil, ErrInvalidMagicBytes
	}

	version := data[2]
	if version != WebSocketProtocolV1 {
		return nil, ErrUnsupportedVersion
	}

	frameType := SubprotocolFrameType(data[3])
	encoding := SerializationEncoding(data[4])
	flags := data[5]

	seqNum := binary.BigEndian.Uint64(data[6:14])
	payloadLen := binary.BigEndian.Uint32(data[14:18])
	expectedChecksum := binary.BigEndian.Uint32(data[18:22])

	totalExpected := FixedHeaderSize + int(payloadLen)
	if len(data) < totalExpected {
		return nil, ErrPayloadTruncated
	}

	payload := data[FixedHeaderSize:totalExpected]
	actualChecksum := crc32.ChecksumIEEE(payload)
	if actualChecksum != expectedChecksum {
		return nil, ErrChecksumMismatch
	}

	return &BinarySubprotocolFrame{
		Version:        version,
		Type:           frameType,
		Encoding:       encoding,
		Flags:          flags,
		SequenceNumber: seqNum,
		Payload:        payload,
	}, nil
}
