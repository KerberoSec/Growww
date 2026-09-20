package parquet

import (
	"time"
)

// TradeRecord represents a standardized execution record for columnar export.
type TradeRecord struct {
	TradeID         string `json:"trade_id"`
	TimestampNs     int64  `json:"timestamp_ns"`
	Symbol          string `json:"symbol"`
	PriceE8         uint64 `json:"price_e8"`
	QuantityE8      uint64 `json:"quantity_e8"`
	Side            string `json:"side"` // "BUY" or "SELL"
	MakerOrderID    string `json:"maker_order_id"`
	TakerOrderID    string `json:"taker_order_id"`
	FeeE8           uint64 `json:"fee_e8"`
	SettlementBlock uint64 `json:"settlement_block"`
}

// OrderRecord represents an order lifecycle snapshot for audit/analytics.
type OrderRecord struct {
	OrderID     string `json:"order_id"`
	TimestampNs int64  `json:"timestamp_ns"`
	AccountID   string `json:"account_id"`
	Symbol      string `json:"symbol"`
	OrderType   string `json:"order_type"` // "LIMIT", "MARKET", "STOP_LIMIT"
	Side        string `json:"side"`
	PriceE8     uint64 `json:"price_e8"`
	QuantityE8  uint64 `json:"quantity_e8"`
	FilledE8    uint64 `json:"filled_e8"`
	Status      string `json:"status"` // "NEW", "PARTIAL", "FILLED", "CANCELLED"
	LatencyNs   int64  `json:"latency_ns"`
}

// ParquetFileManifest contains metadata for an exported columnar file.
type ParquetFileManifest struct {
	ExportID          string    `json:"export_id"`
	TableName         string    `json:"table_name"`
	PartitionPath     string    `json:"partition_path"`
	FileName          string    `json:"file_name"`
	RowCount          int64     `json:"row_count"`
	UncompressedBytes int64     `json:"uncompressed_bytes"`
	CompressedBytes   int64     `json:"compressed_bytes"`
	CompressionRatio  float64   `json:"compression_ratio"`
	SHA256Checksum    string    `json:"sha256_checksum"`
	StartTimestampNs  int64     `json:"start_timestamp_ns"`
	EndTimestampNs    int64     `json:"end_timestamp_ns"`
	ExportedAt        time.Time `json:"exported_at"`
}

// ColumnarTradeBatch holds trades decomposed into parallel primitive arrays (columns).
type ColumnarTradeBatch struct {
	TradeIDs        []string
	TimestampsNs    []int64
	Symbols         []string
	PricesE8        []uint64
	QuantitiesE8    []uint64
	Sides           []string
	MakerOrderIDs   []string
	TakerOrderIDs   []string
	FeesE8          []uint64
	SettlementBlocks []uint64
}

// NewColumnarTradeBatch initializes empty batch slices.
func NewColumnarTradeBatch(capacity int) *ColumnarTradeBatch {
	return &ColumnarTradeBatch{
		TradeIDs:        make([]string, 0, capacity),
		TimestampsNs:    make([]int64, 0, capacity),
		Symbols:         make([]string, 0, capacity),
		PricesE8:        make([]uint64, 0, capacity),
		QuantitiesE8:    make([]uint64, 0, capacity),
		Sides:           make([]string, 0, capacity),
		MakerOrderIDs:   make([]string, 0, capacity),
		TakerOrderIDs:   make([]string, 0, capacity),
		FeesE8:          make([]uint64, 0, capacity),
		SettlementBlocks: make([]uint64, 0, capacity),
	}
}

// Append adds a record into columnar arrays.
func (b *ColumnarTradeBatch) Append(t TradeRecord) {
	b.TradeIDs = append(b.TradeIDs, t.TradeID)
	b.TimestampsNs = append(b.TimestampsNs, t.TimestampNs)
	b.Symbols = append(b.Symbols, t.Symbol)
	b.PricesE8 = append(b.PricesE8, t.PriceE8)
	b.QuantitiesE8 = append(b.QuantitiesE8, t.QuantityE8)
	b.Sides = append(b.Sides, t.Side)
	b.MakerOrderIDs = append(b.MakerOrderIDs, t.MakerOrderID)
	b.TakerOrderIDs = append(b.TakerOrderIDs, t.TakerOrderID)
	b.FeesE8 = append(b.FeesE8, t.FeeE8)
	b.SettlementBlocks = append(b.SettlementBlocks, t.SettlementBlock)
}

// Len returns the batch length.
func (b *ColumnarTradeBatch) Len() int {
	return len(b.TradeIDs)
}
