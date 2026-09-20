package parquet

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// PartitionManager computes standard Hive/Presto-compliant partition directory structures:
// year=YYYY/month=MM/day=DD/hour=HH/symbol=XYZ/
type PartitionManager struct{}

// NewPartitionManager initializes the partition manager.
func NewPartitionManager() *PartitionManager {
	return &PartitionManager{}
}

// BuildPartitionPath constructs a deterministic hierarchical partition key.
func (p *PartitionManager) BuildPartitionPath(timestamp time.Time, symbol string) string {
	t := timestamp.UTC()
	return fmt.Sprintf("year=%04d/month=%02d/day=%02d/hour=%02d/symbol=%s",
		t.Year(), t.Month(), t.Day(), t.Hour(), symbol)
}

// ColumnarParquetStorageEngine coordinates export batching and querying.
type ColumnarParquetStorageEngine struct {
	mu         sync.RWMutex
	partitions map[string][]*ParquetFileManifest // partitionPath -> manifests
	dataStore  map[string][]byte                 // fileName -> compressed data
	pm         *PartitionManager
}

// NewColumnarParquetStorageEngine creates a new parquet storage engine.
func NewColumnarParquetStorageEngine() *ColumnarParquetStorageEngine {
	return &ColumnarParquetStorageEngine{
		partitions: make(map[string][]*ParquetFileManifest),
		dataStore:  make(map[string][]byte),
		pm:         NewPartitionManager(),
	}
}

// ExportTradeBatch writes a columnar batch into storage with partition indexing and SHA-256 manifest.
func (e *ColumnarParquetStorageEngine) ExportTradeBatch(tableName string, symbol string, t time.Time, batch *ColumnarTradeBatch) (*ParquetFileManifest, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if batch.Len() == 0 {
		return nil, fmt.Errorf("cannot export empty trade batch")
	}

	block, err := SerializeColumnarTrades(batch)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize columnar batch: %w", err)
	}

	partitionPath := e.pm.BuildPartitionPath(t, symbol)
	exportID := fmt.Sprintf("exp_%d", time.Now().UnixNano())
	fileName := fmt.Sprintf("%s/%s_%s.parquet", partitionPath, tableName, exportID)

	// Compute SHA-256
	h := sha256.Sum256(block.CompressedBytes)
	checksum := hex.EncodeToString(h[:])

	var startTs, endTs int64
	if len(batch.TimestampsNs) > 0 {
		startTs = batch.TimestampsNs[0]
		endTs = batch.TimestampsNs[len(batch.TimestampsNs)-1]
	}

	ratio := 1.0
	if len(block.CompressedBytes) > 0 {
		ratio = float64(block.UncompressedBytes) / float64(len(block.CompressedBytes))
	}

	manifest := &ParquetFileManifest{
		ExportID:          exportID,
		TableName:         tableName,
		PartitionPath:     partitionPath,
		FileName:          fileName,
		RowCount:          int64(block.RowCount),
		UncompressedBytes: block.UncompressedBytes,
		CompressedBytes:   int64(len(block.CompressedBytes)),
		CompressionRatio:  ratio,
		SHA256Checksum:    checksum,
		StartTimestampNs:  startTs,
		EndTimestampNs:    endTs,
		ExportedAt:        time.Now().UTC(),
	}

	e.partitions[partitionPath] = append(e.partitions[partitionPath], manifest)
	e.dataStore[fileName] = block.CompressedBytes

	return manifest, nil
}

// QueryByPredicate scans partitions matching criteria and applies pushdown filters.
func (e *ColumnarParquetStorageEngine) QueryByPredicate(partitionPath string, minPriceE8, maxPriceE8 uint64, sideFilter string) (*ColumnarTradeBatch, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	manifests, exists := e.partitions[partitionPath]
	if !exists || len(manifests) == 0 {
		return NewColumnarTradeBatch(0), nil
	}

	result := NewColumnarTradeBatch(0)

	for _, m := range manifests {
		data, ok := e.dataStore[m.FileName]
		if !ok {
			continue
		}

		batch, err := DeserializeColumnarTrades(data)
		if err != nil {
			return nil, err
		}

		// Apply columnar predicate pushdown
		for i := 0; i < batch.Len(); i++ {
			p := batch.PricesE8[i]
			s := batch.Sides[i]

			if minPriceE8 > 0 && p < minPriceE8 {
				continue
			}
			if maxPriceE8 > 0 && p > maxPriceE8 {
				continue
			}
			if sideFilter != "" && s != sideFilter {
				continue
			}

			result.Append(TradeRecord{
				TradeID:         batch.TradeIDs[i],
				TimestampNs:     batch.TimestampsNs[i],
				Symbol:          batch.Symbols[i],
				PriceE8:         batch.PricesE8[i],
				QuantityE8:      batch.QuantitiesE8[i],
				Side:            batch.Sides[i],
				MakerOrderID:    "",
				TakerOrderID:    "",
				FeeE8:           batch.FeesE8[i],
				SettlementBlock: batch.SettlementBlocks[i],
			})
		}
	}

	return result, nil
}
