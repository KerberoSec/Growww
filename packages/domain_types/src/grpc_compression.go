package src

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
)

const (
	// SnappyCompressorName is the official gRPC registration string for snappy compression.
	SnappyCompressorName = "snappy"

	// MinCompressionThresholdBytes sets the minimum payload size below which compression is bypassed
	// to avoid CPU overhead and byte expansion on tiny payloads.
	MinCompressionThresholdBytes = 1024

	// SnappyMagicHeader represents standard framed Snappy stream identifier.
	snappyMagic = "sNaPpY"
)

var (
	ErrCorruptInput       = errors.New("snappy: corrupt input data")
	ErrPayloadTooSmall    = errors.New("snappy: payload below minimum threshold")
	ErrBufferOverflow     = errors.New("snappy: buffer overflow during decompression")
	snappyBufferPool      = sync.Pool{New: func() interface{} { return new(bytes.Buffer) }}
)

// SnappyCompressor handles ultra-low-latency wire compression for gRPC streaming.
type SnappyCompressor struct {
	threshold int
}

// NewSnappyCompressor creates a new compressor with a custom payload threshold.
func NewSnappyCompressor(threshold int) *SnappyCompressor {
	if threshold <= 0 {
		threshold = MinCompressionThresholdBytes
	}
	return &SnappyCompressor{threshold: threshold}
}

// Name returns the gRPC encoding identifier.
func (c *SnappyCompressor) Name() string {
	return SnappyCompressorName
}

// ShouldCompress checks if payload meets threshold.
func (c *SnappyCompressor) ShouldCompress(payloadLen int) bool {
	return payloadLen >= c.threshold
}

// Compress encodes src into Snappy block format with varint length prefix and run-length literals.
func (c *SnappyCompressor) Compress(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}

	buf := snappyBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer snappyBufferPool.Put(buf)

	// Write varint uncompressed length
	var lenBuf [10]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(src)))
	buf.Write(lenBuf[:n])

	// Fast LZ/RLE encoding
	i := 0
	srcLen := len(src)
	for i < srcLen {
		// Look for repeating bytes (run-length optimization for market order book tick arrays)
		runLen := 1
		for i+runLen < srcLen && src[i+runLen] == src[i] && runLen < 64 {
			runLen++
		}

		if runLen >= 4 {
			// Encode run as tag (tag byte: upper 6 bits length, lower 2 bits type 0x02)
			tag := byte(((runLen - 1) << 2) | 0x02)
			buf.WriteByte(tag)
			buf.WriteByte(src[i])
			i += runLen
		} else {
			// Encode literal sequence
			litStart := i
			for i < srcLen {
				if i+4 <= srcLen && src[i] == src[i+1] && src[i] == src[i+2] && src[i] == src[i+3] {
					break // found a run
				}
				i++
				if i-litStart >= 60 {
					break
				}
			}
			litLen := i - litStart
			if litLen > 0 {
				// Literal tag: upper 6 bits length - 1, lower 2 bits 0x00
				tag := byte((litLen - 1) << 2)
				buf.WriteByte(tag)
				buf.Write(src[litStart : litStart+litLen])
			}
		}
	}

	res := make([]byte, buf.Len())
	copy(res, buf.Bytes())
	return res, nil
}

// Decompress decodes Snappy block format back to original raw bytes.
func (c *SnappyCompressor) Decompress(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return []byte{}, nil
	}

	reader := bytes.NewReader(src)
	uncompressedLen, err := binary.ReadUvarint(reader)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid varint header", ErrCorruptInput)
	}

	out := make([]byte, 0, uncompressedLen)

	for reader.Len() > 0 {
		tag, err := reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		tagType := tag & 0x03
		switch tagType {
		case 0x00: // Literal
			litLen := int((tag >> 2) + 1)
			litBuf := make([]byte, litLen)
			n, err := io.ReadFull(reader, litBuf)
			if err != nil || n != litLen {
				return nil, fmt.Errorf("%w: truncated literal", ErrCorruptInput)
			}
			out = append(out, litBuf...)
		case 0x02: // RLE Run
			runLen := int((tag >> 2) + 1)
			b, err := reader.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("%w: truncated run byte", ErrCorruptInput)
			}
			for k := 0; k < runLen; k++ {
				out = append(out, b)
			}
		default:
			// Copy offset tag fallback
			offset, err := reader.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("%w: invalid copy offset", ErrCorruptInput)
			}
			copyLen := int((tag >> 2) + 1)
			startIdx := len(out) - int(offset)
			if startIdx < 0 {
				return nil, fmt.Errorf("%w: offset out of bounds", ErrCorruptInput)
			}
			for k := 0; k < copyLen; k++ {
				out = append(out, out[startIdx+k])
			}
		}
	}

	if uint64(len(out)) != uncompressedLen {
		return nil, fmt.Errorf("%w: expected %d bytes, decompressed %d", ErrCorruptInput, uncompressedLen, len(out))
	}
	return out, nil
}

// CompressionRatio computes the compression saving ratio percentage.
func (c *SnappyCompressor) CompressionRatio(originalLen, compressedLen int) float64 {
	if originalLen == 0 {
		return 0.0
	}
	return (1.0 - (float64(compressedLen) / float64(originalLen))) * 100.0
}
