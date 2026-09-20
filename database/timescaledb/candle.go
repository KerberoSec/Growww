package timescaledb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// Candle represents an institutional-grade financial candlestick (OHLCV).
// Amounts and prices are stored in fixed-point 1e8 precision (e8 satoshi/paise equivalent).
type Candle struct {
	Bucket        time.Time `json:"bucket"`
	Symbol        string    `json:"symbol"`
	IntervalSec   int       `json:"interval_sec"`
	OpenE8        uint64    `json:"open_e8"`
	HighE8        uint64    `json:"high_e8"`
	LowE8         uint64    `json:"low_e8"`
	CloseE8       uint64    `json:"close_e8"`
	VolumeE8      uint64    `json:"volume_e8"`
	QuoteVolumeE8 uint64    `json:"quote_volume_e8"`
	TradeCount    uint64    `json:"trade_count"`
	VWAPE8        uint64    `json:"vwap_e8"`
	StateHash     string    `json:"state_hash"`
	Finalized     bool      `json:"finalized"`
}

// TradeTick represents an incoming raw execution event from the matching engine.
type TradeTick struct {
	TradeID     string    `json:"trade_id"`
	Symbol      string    `json:"symbol"`
	PriceE8     uint64    `json:"price_e8"`
	QuantityE8  uint64    `json:"quantity_e8"`
	Timestamp   time.Time `json:"timestamp"`
	IsBuyerMaker bool     `json:"is_buyer_maker"`
}

// ChunkStatus represents the lifecycle state of a hypertable chunk.
type ChunkStatus string

const (
	ChunkStatusActive     ChunkStatus = "ACTIVE"
	ChunkStatusCompressing ChunkStatus = "COMPRESSING"
	ChunkStatusCompressed ChunkStatus = "COMPRESSED"
	ChunkStatusDecompressed ChunkStatus = "DECOMPRESSED"
	ChunkStatusArchived   ChunkStatus = "ARCHIVED"
)

// CandleChunk represents a logical partition / chunk in TimescaleDB.
type CandleChunk struct {
	ChunkID        string      `json:"chunk_id"`
	Symbol         string      `json:"symbol"`
	IntervalSec    int         `json:"interval_sec"`
	StartTime      time.Time   `json:"start_time"`
	EndTime        time.Time   `json:"end_time"`
	Status         ChunkStatus `json:"status"`
	RowCount       int         `json:"row_count"`
	UncompressedBytes int64   `json:"uncompressed_bytes"`
	CompressedBytes   int64   `json:"compressed_bytes"`
	CompressionRatio  float64 `json:"compression_ratio"`
	MerkleRoot     string      `json:"merkle_root"`
	Candles        []Candle    `json:"candles"`
}

// ComputeStateHash calculates a deterministic SHA-256 cryptographic state hash for a candle.
func (c *Candle) ComputeStateHash() string {
	payload := fmt.Sprintf("%d:%s:%d:%d:%d:%d:%d:%d:%d:%d:%d:%t",
		c.Bucket.UnixNano(),
		c.Symbol,
		c.IntervalSec,
		c.OpenE8,
		c.HighE8,
		c.LowE8,
		c.CloseE8,
		c.VolumeE8,
		c.QuoteVolumeE8,
		c.TradeCount,
		c.VWAPE8,
		c.Finalized,
	)
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

// UpdateWithTick incorporates a new trade tick into an open candle.
func (c *Candle) UpdateWithTick(tick TradeTick) {
	if c.TradeCount == 0 {
		c.OpenE8 = tick.PriceE8
		c.HighE8 = tick.PriceE8
		c.LowE8 = tick.PriceE8
		c.CloseE8 = tick.PriceE8
	} else {
		if tick.PriceE8 > c.HighE8 {
			c.HighE8 = tick.PriceE8
		}
		if tick.PriceE8 < c.LowE8 {
			c.LowE8 = tick.PriceE8
		}
		c.CloseE8 = tick.PriceE8
	}

	// Use math/big for institutional-grade precision without 64-bit integer overflow
	pBig := new(big.Int).SetUint64(tick.PriceE8)
	qBig := new(big.Int).SetUint64(tick.QuantityE8)
	e8Big := big.NewInt(100_000_000)

	// quoteVol = (PriceE8 * QuantityE8) / 1e8
	prod := new(big.Int).Mul(pBig, qBig)
	quoteVolBig := new(big.Int).Div(prod, e8Big)
	quoteVol := quoteVolBig.Uint64()

	c.VolumeE8 += tick.QuantityE8
	c.QuoteVolumeE8 += quoteVol
	c.TradeCount++

	// VWAPE8 = (QuoteVolumeE8 * 1e8) / VolumeE8
	if c.VolumeE8 > 0 {
		totQuoteBig := new(big.Int).SetUint64(c.QuoteVolumeE8)
		totVolBig := new(big.Int).SetUint64(c.VolumeE8)
		num := new(big.Int).Mul(totQuoteBig, e8Big)
		vwapBig := new(big.Int).Div(num, totVolBig)
		c.VWAPE8 = vwapBig.Uint64()
	} else {
		c.VWAPE8 = c.CloseE8
	}

	c.StateHash = c.ComputeStateHash()
}
