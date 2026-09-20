package timescaledb

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

var (
	ErrInvalidTick         = errors.New("invalid trade tick payload")
	ErrChunkLocked         = errors.New("chunk is currently locked for compression")
	ErrCandleNotFound      = errors.New("candle not found for specified query")
	ErrIntervalUnsupported = errors.New("unsupported candle interval")
)

// SupportedIntervals define valid aggregation windows in seconds.
var SupportedIntervals = []int{1, 60, 300, 900, 3600, 86400} // 1s, 1m, 5m, 15m, 1h, 1d

// StorageMetrics tracks operational telemetry.
type StorageMetrics struct {
	TotalTicksIngested   uint64
	TotalCandlesCreated  uint64
	CompressedChunksCount uint64
	TotalBytesSaved      int64
	LastReconciliationMs int64
}

// DeadLetterQueue captures malformed or corrupt ticks.
type DeadLetterQueue struct {
	mu     sync.Mutex
	Failed []DeadLetterRecord
}

type DeadLetterRecord struct {
	Tick      TradeTick `json:"tick"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
}

// TimescaleCandleEngine provides the in-memory + hypertable storage coordinator.
type TimescaleCandleEngine struct {
	mu                  sync.RWMutex
	activeCandles       map[string]*Candle       // key: symbol:interval:bucket_unix
	chunks              map[string]*CandleChunk  // key: chunk_id
	compressionManager  *CompressionManager
	reconciliationLoop  *ReconciliationManager
	dlq                 *DeadLetterQueue
	metrics             StorageMetrics
	circuitBreakerOpen  bool
}

// NewTimescaleCandleEngine initializes a production-grade TimescaleDB candle storage engine.
func NewTimescaleCandleEngine() *TimescaleCandleEngine {
	engine := &TimescaleCandleEngine{
		activeCandles:      make(map[string]*Candle),
		chunks:             make(map[string]*CandleChunk),
		dlq:                &DeadLetterQueue{Failed: make([]DeadLetterRecord, 0)},
		circuitBreakerOpen: false,
	}
	engine.compressionManager = NewCompressionManager(engine)
	engine.reconciliationLoop = NewReconciliationManager(engine)
	return engine
}

// IngestTradeTick processes an incoming trade tick and updates candles across all supported intervals.
func (e *TimescaleCandleEngine) IngestTradeTick(ctx context.Context, tick TradeTick) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.circuitBreakerOpen {
		return errors.New("circuit breaker is OPEN: storage engine degraded")
	}

	// Validation
	if tick.Symbol == "" || tick.PriceE8 == 0 || tick.QuantityE8 == 0 || tick.Timestamp.IsZero() {
		e.dlq.mu.Lock()
		e.dlq.Failed = append(e.dlq.Failed, DeadLetterRecord{
			Tick:     tick,
			Reason:   "Malformed tick parameters (zero price/qty or empty symbol)",
			FailedAt: time.Now(),
		})
		e.dlq.mu.Unlock()
		return fmt.Errorf("%w: symbol=%s, price=%d, qty=%d", ErrInvalidTick, tick.Symbol, tick.PriceE8, tick.QuantityE8)
	}

	for _, interval := range SupportedIntervals {
		bucketTime := tick.Timestamp.Truncate(time.Duration(interval) * time.Second)
		key := fmt.Sprintf("%s:%d:%d", tick.Symbol, interval, bucketTime.Unix())

		c, exists := e.activeCandles[key]
		if !exists {
			c = &Candle{
				Bucket:      bucketTime,
				Symbol:      tick.Symbol,
				IntervalSec: interval,
				Finalized:   false,
			}
			e.activeCandles[key] = c
			e.metrics.TotalCandlesCreated++
		}

		c.UpdateWithTick(tick)
	}

	e.metrics.TotalTicksIngested++
	return nil
}

// FinalizeCandle seals a candle bucket and assigns it to its hypertable chunk.
func (e *TimescaleCandleEngine) FinalizeCandle(symbol string, interval int, bucket time.Time) (*Candle, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	bucketTime := bucket.Truncate(time.Duration(interval) * time.Second)
	key := fmt.Sprintf("%s:%d:%d", symbol, interval, bucketTime.Unix())

	c, exists := e.activeCandles[key]
	if !exists {
		return nil, ErrCandleNotFound
	}

	c.Finalized = true
	c.StateHash = c.ComputeStateHash()

	// Assign to chunk (partitioned by 1 day)
	chunkStart := bucketTime.Truncate(24 * time.Hour)
	chunkEnd := chunkStart.Add(24 * time.Hour)
	chunkID := fmt.Sprintf("chunk_%s_%d_%d", symbol, interval, chunkStart.Unix())

	chunk, exists := e.chunks[chunkID]
	if !exists {
		chunk = &CandleChunk{
			ChunkID:     chunkID,
			Symbol:      symbol,
			IntervalSec: interval,
			StartTime:   chunkStart,
			EndTime:     chunkEnd,
			Status:      ChunkStatusActive,
			Candles:     make([]Candle, 0),
		}
		e.chunks[chunkID] = chunk
	}

	chunk.Candles = append(chunk.Candles, *c)
	chunk.RowCount = len(chunk.Candles)
	chunk.UncompressedBytes += 128 // Approximate row size

	return c, nil
}

// QueryCandles retrieves candles matching symbol, interval, and time bounds.
func (e *TimescaleCandleEngine) QueryCandles(symbol string, interval int, from, to time.Time) ([]Candle, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []Candle
	for _, c := range e.activeCandles {
		if c.Symbol == symbol && c.IntervalSec == interval {
			if !c.Bucket.Before(from) && !c.Bucket.After(to) {
				result = append(result, *c)
			}
		}
	}

	// Also query sealed chunks
	for _, chunk := range e.chunks {
		if chunk.Symbol == symbol && chunk.IntervalSec == interval {
			if chunk.EndTime.Before(from) || chunk.StartTime.After(to) {
				continue
			}
			for _, c := range chunk.Candles {
				if !c.Bucket.Before(from) && !c.Bucket.After(to) {
					// Avoid duplicates if already in active
					already := false
					for _, existing := range result {
						if existing.Bucket.Equal(c.Bucket) {
							already = true
							break
						}
					}
					if !already {
						result = append(result, c)
					}
				}
			}
		}
	}

	// Sort chronologically ascending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Bucket.Before(result[j].Bucket)
	})

	return result, nil
}

// GetMetrics returns engine metrics.
func (e *TimescaleCandleEngine) GetMetrics() StorageMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.metrics
}

// SetCircuitBreaker toggles circuit breaker state.
func (e *TimescaleCandleEngine) SetCircuitBreaker(open bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.circuitBreakerOpen = open
}

// GetDLQ returns dead letter queue records.
func (e *TimescaleCandleEngine) GetDLQ() []DeadLetterRecord {
	e.dlq.mu.Lock()
	defer e.dlq.mu.Unlock()
	copied := make([]DeadLetterRecord, len(e.dlq.Failed))
	copy(copied, e.dlq.Failed)
	return copied
}
