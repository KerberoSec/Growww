package opensearch

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// IndexerMetrics tracks ingestion performance.
type IndexerMetrics struct {
	TotalIndexed uint64
	TotalFailed  uint64
	FlushesCount uint64
}

// BulkIndexer coordinates buffered high-throughput indexing into OpenSearch.
type BulkIndexer struct {
	mu            sync.Mutex
	buffer        []interface{}
	batchSize     int
	flushInterval time.Duration
	maxBufferSize int
	flushCallback func(ctx context.Context, batch []interface{}) error
	metrics       IndexerMetrics
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

// NewBulkIndexer initializes the bulk indexer with automatic background flush.
func NewBulkIndexer(
	batchSize int,
	maxBufferSize int,
	flushInterval time.Duration,
	flushCallback func(ctx context.Context, batch []interface{}) error,
) *BulkIndexer {
	if batchSize <= 0 {
		batchSize = 200
	}
	if maxBufferSize <= 0 {
		maxBufferSize = 2000
	}
	if flushInterval <= 0 {
		flushInterval = 100 * time.Millisecond
	}

	bi := &BulkIndexer{
		buffer:        make([]interface{}, 0, batchSize),
		batchSize:     batchSize,
		maxBufferSize: maxBufferSize,
		flushInterval: flushInterval,
		flushCallback: flushCallback,
		stopCh:        make(chan struct{}),
	}

	bi.wg.Add(1)
	go bi.flushLoop()

	return bi
}

// Ingest queues a document for bulk indexing.
func (bi *BulkIndexer) Ingest(ctx context.Context, doc interface{}) error {
	bi.mu.Lock()
	if len(bi.buffer) >= bi.maxBufferSize {
		bi.mu.Unlock()
		atomic.AddUint64(&bi.metrics.TotalFailed, 1)
		return ErrBufferFull
	}

	bi.buffer = append(bi.buffer, doc)
	shouldFlush := len(bi.buffer) >= bi.batchSize
	bi.mu.Unlock()

	if shouldFlush {
		return bi.Flush(ctx)
	}
	return nil
}

// Flush immediately flushes buffered documents.
func (bi *BulkIndexer) Flush(ctx context.Context) error {
	bi.mu.Lock()
	if len(bi.buffer) == 0 {
		bi.mu.Unlock()
		return nil
	}

	batch := bi.buffer
	bi.buffer = make([]interface{}, 0, bi.batchSize)
	bi.mu.Unlock()

	atomic.AddUint64(&bi.metrics.FlushesCount, 1)
	if bi.flushCallback != nil {
		if err := bi.flushCallback(ctx, batch); err != nil {
			atomic.AddUint64(&bi.metrics.TotalFailed, uint64(len(batch)))
			return err
		}
	}

	atomic.AddUint64(&bi.metrics.TotalIndexed, uint64(len(batch)))
	return nil
}

func (bi *BulkIndexer) flushLoop() {
	defer bi.wg.Done()
	ticker := time.NewTicker(bi.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bi.stopCh:
			_ = bi.Flush(context.Background())
			return
		case <-ticker.C:
			_ = bi.Flush(context.Background())
		}
	}
}

// Stop shuts down the bulk indexer safely after flushing.
func (bi *BulkIndexer) Stop() {
	close(bi.stopCh)
	bi.wg.Wait()
}

// Metrics returns the telemetry counters.
func (bi *BulkIndexer) Metrics() IndexerMetrics {
	return IndexerMetrics{
		TotalIndexed: atomic.LoadUint64(&bi.metrics.TotalIndexed),
		TotalFailed:  atomic.LoadUint64(&bi.metrics.TotalFailed),
		FlushesCount: atomic.LoadUint64(&bi.metrics.FlushesCount),
	}
}
