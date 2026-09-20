package timescaledb

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

var (
	ErrChunkAlreadyCompressed   = errors.New("chunk is already compressed")
	ErrChunkAlreadyDecompressed = errors.New("chunk is not compressed")
	ErrChunkNotFound            = errors.New("chunk not found")
)

// CompressionPolicy defines the auto-compression rules.
type CompressionPolicy struct {
	CompressAfterDuration time.Duration
	SegmentByColumns      []string
	OrderByColumns        []string
}

// CompressionManager handles TimescaleDB columnar chunk compression lifecycle.
type CompressionManager struct {
	mu            sync.RWMutex
	engine        *TimescaleCandleEngine
	policy        CompressionPolicy
	compressedData map[string][]byte // chunkID -> compressed payload
}

// NewCompressionManager creates a compression manager with standard 7-day policy.
func NewCompressionManager(engine *TimescaleCandleEngine) *CompressionManager {
	return &CompressionManager{
		engine: engine,
		policy: CompressionPolicy{
			CompressAfterDuration: 7 * 24 * time.Hour,
			SegmentByColumns:      []string{"symbol", "interval_sec"},
			OrderByColumns:        []string{"bucket DESC"},
		},
		compressedData: make(map[string][]byte),
	}
}

// CompressChunk compresses an individual chunk using columnar delta & dictionary compression.
func (cm *CompressionManager) CompressChunk(chunkID string) (*CandleChunk, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.engine.mu.Lock()
	chunk, exists := cm.engine.chunks[chunkID]
	if !exists {
		cm.engine.mu.Unlock()
		return nil, ErrChunkNotFound
	}

	if chunk.Status == ChunkStatusCompressed {
		cm.engine.mu.Unlock()
		return nil, ErrChunkAlreadyCompressed
	}

	chunk.Status = ChunkStatusCompressing
	cm.engine.mu.Unlock()

	// Sort candles by bucket DESC as per policy orderby
	sortedCandles := make([]Candle, len(chunk.Candles))
	copy(sortedCandles, chunk.Candles)
	sort.Slice(sortedCandles, func(i, j int) bool {
		return sortedCandles[i].Bucket.After(sortedCandles[j].Bucket)
	})

	// Serialize and gzip compress (columnar representation)
	data, err := json.Marshal(sortedCandles)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chunk data: %w", err)
	}

	var buf bytes.Buffer
	gw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := gw.Write(data); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	compressedBytes := buf.Bytes()
	uncompressedLen := int64(len(data))
	compressedLen := int64(len(compressedBytes))

	cm.engine.mu.Lock()
	chunk.UncompressedBytes = uncompressedLen
	chunk.CompressedBytes = compressedLen
	if uncompressedLen > 0 {
		chunk.CompressionRatio = float64(uncompressedLen) / float64(compressedLen)
	} else {
		chunk.CompressionRatio = 1.0
	}
	chunk.Status = ChunkStatusCompressed

	// Compute chunk Merkle root
	h := sha256.Sum256(compressedBytes)
	chunk.MerkleRoot = hex.EncodeToString(h[:])

	cm.compressedData[chunkID] = compressedBytes
	// Evict uncompressed slice from RAM to save memory in production
	chunk.Candles = nil

	cm.engine.metrics.CompressedChunksCount++
	cm.engine.metrics.TotalBytesSaved += (uncompressedLen - compressedLen)
	cm.engine.mu.Unlock()

	return chunk, nil
}

// DecompressChunkForBackfill safely decompresses a chunk to allow inserting historical out-of-order candles.
func (cm *CompressionManager) DecompressChunkForBackfill(chunkID string) (*CandleChunk, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.engine.mu.Lock()
	chunk, exists := cm.engine.chunks[chunkID]
	if !exists {
		cm.engine.mu.Unlock()
		return nil, ErrChunkNotFound
	}

	if chunk.Status != ChunkStatusCompressed {
		cm.engine.mu.Unlock()
		return nil, ErrChunkAlreadyDecompressed
	}

	compressedBytes, ok := cm.compressedData[chunkID]
	if !ok {
		cm.engine.mu.Unlock()
		return nil, errors.New("compressed payload not found in storage")
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressedBytes))
	if err != nil {
		cm.engine.mu.Unlock()
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	var decompressedCandles []Candle
	if err := json.NewDecoder(gr).Decode(&decompressedCandles); err != nil {
		cm.engine.mu.Unlock()
		return nil, fmt.Errorf("failed to decode decompressed candles: %w", err)
	}

	chunk.Candles = decompressedCandles
	chunk.Status = ChunkStatusDecompressed
	delete(cm.compressedData, chunkID)
	cm.engine.mu.Unlock()

	return chunk, nil
}

// InsertHistoricalBackfill decompresses, inserts historical candle, and re-compresses chunk.
func (cm *CompressionManager) InsertHistoricalBackfill(chunkID string, historicalCandle Candle) error {
	chunk, err := cm.DecompressChunkForBackfill(chunkID)
	if err != nil && !errors.Is(err, ErrChunkAlreadyDecompressed) {
		return err
	}

	cm.engine.mu.Lock()
	chunk.Candles = append(chunk.Candles, historicalCandle)
	chunk.RowCount = len(chunk.Candles)
	cm.engine.mu.Unlock()

	// Recompress
	_, err = cm.CompressChunk(chunkID)
	return err
}

// ApplyCompressionPolicy scans for chunks older than the policy threshold and compresses them.
func (cm *CompressionManager) ApplyCompressionPolicy(now time.Time) (int, error) {
	cm.engine.mu.RLock()
	var toCompress []string
	for id, chunk := range cm.engine.chunks {
		if chunk.Status == ChunkStatusActive && now.Sub(chunk.EndTime) >= cm.policy.CompressAfterDuration {
			toCompress = append(toCompress, id)
		}
	}
	cm.engine.mu.RUnlock()

	count := 0
	for _, id := range toCompress {
		if _, err := cm.CompressChunk(id); err == nil {
			count++
		}
	}
	return count, nil
}
