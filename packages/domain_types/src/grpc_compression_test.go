package src

import (
	"bytes"
	"testing"
)

func TestSnappyCompressionRoundtrip(t *testing.T) {
	compressor := NewSnappyCompressor(1024)

	if compressor.Name() != "snappy" {
		t.Fatalf("expected name snappy, got %s", compressor.Name())
	}

	// Test threshold
	if compressor.ShouldCompress(500) {
		t.Fatalf("expected 500 bytes to be below 1024 threshold")
	}
	if !compressor.ShouldCompress(2048) {
		t.Fatalf("expected 2048 bytes to be above threshold")
	}

	// Construct repetitive sample data (like orderbook depth ticks)
	var rawData []byte
	for i := 0; i < 200; i++ {
		rawData = append(rawData, []byte("NIFTY50-25000-CE-QUOTE-BID-ASK-000000000000000000000000")...)
	}

	compressed, err := compressor.Compress(rawData)
	if err != nil {
		t.Fatalf("compression failed: %v", err)
	}

	if len(compressed) >= len(rawData) {
		t.Fatalf("expected compression saving, raw=%d compressed=%d", len(rawData), len(compressed))
	}

	ratio := compressor.CompressionRatio(len(rawData), len(compressed))
	if ratio < 30.0 {
		t.Fatalf("expected at least 30%% compression ratio, got %.2f%%", ratio)
	}

	decompressed, err := compressor.Decompress(compressed)
	if err != nil {
		t.Fatalf("decompression failed: %v", err)
	}

	if !bytes.Equal(rawData, decompressed) {
		t.Fatalf("decompressed data does not match original data")
	}

	// Empty payload test
	emptyComp, err := compressor.Compress([]byte{})
	if err != nil || len(emptyComp) != 0 {
		t.Fatalf("expected empty result for empty input")
	}
	emptyDecomp, err := compressor.Decompress([]byte{})
	if err != nil || len(emptyDecomp) != 0 {
		t.Fatalf("expected empty result for empty decomp")
	}
}
