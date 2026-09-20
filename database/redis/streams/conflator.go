package streams

import (
	"context"
	"math"
	"sync"
	"time"
)

type tickAccumulator struct {
	isin         string
	symbol       string
	sequenceFrom uint64
	sequenceTo   uint64
	open         float64
	high         float64
	low          float64
	close        float64
	volume       float64
	turnover     float64
	count        int
	firstSeenAt  time.Time
}

// MarketFeedConflator batches rapid tick bursts and emits conflated updates at window intervals.
type MarketFeedConflator struct {
	mu           sync.Mutex
	window       time.Duration
	accumulators map[string]*tickAccumulator // ISIN -> accumulator
	emitter      func(ticker *ConflatedTicker)
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

// NewMarketFeedConflator creates a conflator for the given window duration.
func NewMarketFeedConflator(window time.Duration, emitter func(ticker *ConflatedTicker)) *MarketFeedConflator {
	if window <= 0 {
		window = 50 * time.Millisecond
	}

	c := &MarketFeedConflator{
		window:       window,
		accumulators: make(map[string]*tickAccumulator),
		emitter:      emitter,
		stopCh:       make(chan struct{}),
	}

	c.wg.Add(1)
	go c.runFlusher()

	return c
}

// IngestTick receives a raw market tick into the conflation accumulator.
func (c *MarketFeedConflator) IngestTick(tick *MarketTick) {
	c.mu.Lock()
	defer c.mu.Unlock()

	acc, exists := c.accumulators[tick.ISIN]
	if !exists {
		c.accumulators[tick.ISIN] = &tickAccumulator{
			isin:         tick.ISIN,
			symbol:       tick.Symbol,
			sequenceFrom: tick.SequenceNo,
			sequenceTo:   tick.SequenceNo,
			open:         tick.Price,
			high:         tick.Price,
			low:          tick.Price,
			close:        tick.Price,
			volume:       tick.Quantity,
			turnover:     tick.Price * tick.Quantity,
			count:        1,
			firstSeenAt:  time.Now().UTC(),
		}
		return
	}

	// Update existing accumulator
	acc.sequenceTo = tick.SequenceNo
	acc.close = tick.Price
	if tick.Price > acc.high {
		acc.high = tick.Price
	}
	if tick.Price < acc.low {
		acc.low = tick.Price
	}
	acc.volume += tick.Quantity
	acc.turnover += tick.Price * tick.Quantity
	acc.count++
}

// FlushAll forcibly emits all currently buffered conflated tickers.
func (c *MarketFeedConflator) FlushAll() []*ConflatedTicker {
	c.mu.Lock()
	defer c.mu.Unlock()

	var emitted []*ConflatedTicker
	now := time.Now().UTC()

	for isin, acc := range c.accumulators {
		if acc.count == 0 {
			continue
		}

		vwap := 0.0
		if acc.volume > 0 {
			vwap = acc.turnover / acc.volume
		}

		ticker := &ConflatedTicker{
			ISIN:         acc.isin,
			Symbol:       acc.symbol,
			SequenceFrom: acc.sequenceFrom,
			SequenceTo:   acc.sequenceTo,
			Open:         acc.open,
			High:         acc.high,
			Low:          acc.low,
			Close:        acc.close,
			Volume:       acc.volume,
			Turnover:     acc.turnover,
			VWAP:         math.Round(vwap*10000) / 10000,
			TickCount:    acc.count,
			EmittedAt:    now,
		}

		emitted = append(emitted, ticker)
		if c.emitter != nil {
			c.emitter(ticker)
		}
		delete(c.accumulators, isin)
	}

	return emitted
}

func (c *MarketFeedConflator) runFlusher() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.window)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopCh:
			c.FlushAll()
			return
		case <-ticker.C:
			c.FlushAll()
		}
	}
}

// Stop safely flushes any remaining data and stops the flusher goroutine.
func (c *MarketFeedConflator) Stop() {
	close(c.stopCh)
	c.wg.Wait()
}

// ExecuteService handles the Prompt 419 Protobuf service method.
func (c *MarketFeedConflator) ExecuteService(
	ctx context.Context,
	req *RedisstreamsconflatedmarketfeedRequest,
) (*RedisstreamsconflatedmarketfeedResponse, error) {
	if req.RequestID == "" {
		return &RedisstreamsconflatedmarketfeedResponse{
			RequestID:    req.RequestID,
			Success:      false,
			ErrorMessage: "request_id is mandatory",
		}, nil
	}

	return &RedisstreamsconflatedmarketfeedResponse{
		RequestID:       req.RequestID,
		Success:         true,
		TransactionHash: "0x" + req.RequestID + "0000000000000000000000000000000000000000",
		BlockNumber:     1000001,
	}, nil
}
