package ordermatching

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// SyncStatus represents the client-side synchronization state machine
type SyncStatus string

const (
	StatusBufferingDeltas SyncStatus = "BUFFERING_DELTAS"
	StatusSynchronized    SyncStatus = "SYNCHRONIZED"
	StatusDesynchronized  SyncStatus = "DESYNCHRONIZED"
)

// PriceLevel represents an aggregated price level in the L2 orderbook
type PriceLevel struct {
	PriceE8         uint64 `json:"price_e8"`
	TotalQuantityE8 uint64 `json:"total_quantity_e8"`
	OrderCount      uint32 `json:"order_count"`
}

// L2Snapshot represents the full authoritative snapshot of the orderbook
type L2Snapshot struct {
	Symbol       string       `json:"symbol"`
	LastUpdateID uint64       `json:"last_update_id"`
	TimestampNs  uint64       `json:"timestamp_ns"`
	Bids         []PriceLevel `json:"bids"`
	Asks         []PriceLevel `json:"asks"`
}

// PriceDelta represents a single level change (quantity=0 clears level)
type PriceDelta struct {
	PriceE8 uint64 `json:"price_e8"`
	QtyE8   uint64 `json:"quantity_e8"`
}

// BookDelta represents an incremental orderbook delta stream packet
type BookDelta struct {
	Symbol        string       `json:"symbol"`
	FirstUpdateID uint64       `json:"first_update_id"`
	LastUpdateID  uint64       `json:"last_update_id"`
	TimestampNs   uint64       `json:"timestamp_ns"`
	Bids          []PriceDelta `json:"bids"`
	Asks          []PriceDelta `json:"asks"`
}

// OrderBookSync manages client-side delta sync and snapshot reconciliation (Prompt 065)
type OrderBookSync struct {
	mu             sync.RWMutex
	symbol         string
	status         SyncStatus
	lastUpdateID   uint64
	bids           map[uint64]uint64
	asks           map[uint64]uint64
	deltaBuffer    []BookDelta
	maxBufferSize  int
	resyncRequests uint64
}

// NewOrderBookSync initializes a new synchronizer instance
func NewOrderBookSync(symbol string) *OrderBookSync {
	return &OrderBookSync{
		symbol:        symbol,
		status:        StatusBufferingDeltas,
		lastUpdateID:  0,
		bids:          make(map[uint64]uint64),
		asks:          make(map[uint64]uint64),
		deltaBuffer:   make([]BookDelta, 0, 1024),
		maxBufferSize: 10000,
	}
}

// Status returns current sync state
func (s *OrderBookSync) Status() SyncStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// LastUpdateID returns current synced sequence ID
func (s *OrderBookSync) LastUpdateID() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastUpdateID
}

// ResyncCount returns total resynchronization trigger count
func (s *OrderBookSync) ResyncCount() uint64 {
	return atomic.LoadUint64(&s.resyncRequests)
}

// BufferOrApplyDelta receives an incremental delta
func (s *OrderBookSync) BufferOrApplyDelta(delta BookDelta) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.status == StatusSynchronized {
		return s.applyDeltaLocked(delta)
	}

	// Buffering mode before/during snapshot fetch
	if len(s.deltaBuffer) >= s.maxBufferSize {
		s.deltaBuffer = s.deltaBuffer[1:]
	}
	s.deltaBuffer = append(s.deltaBuffer, delta)
	return nil
}

// ApplySnapshot applies full snapshot and replays applicable buffered deltas
func (s *OrderBookSync) ApplySnapshot(snapshot L2Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if snapshot.Symbol != s.symbol {
		return fmt.Errorf("symbol mismatch: expected %s, got %s", s.symbol, snapshot.Symbol)
	}

	s.bids = make(map[uint64]uint64)
	s.asks = make(map[uint64]uint64)

	for _, b := range snapshot.Bids {
		if b.TotalQuantityE8 > 0 {
			s.bids[b.PriceE8] = b.TotalQuantityE8
		}
	}
	for _, a := range snapshot.Asks {
		if a.TotalQuantityE8 > 0 {
			s.asks[a.PriceE8] = a.TotalQuantityE8
		}
	}

	s.lastUpdateID = snapshot.LastUpdateID

	// Replay buffered deltas
	firstApplicable := false
	for _, delta := range s.deltaBuffer {
		if delta.LastUpdateID <= s.lastUpdateID {
			// Discard older deltas
			continue
		}

		if !firstApplicable {
			// First applicable delta must bridge snapshot.LastUpdateID + 1
			if delta.FirstUpdateID > s.lastUpdateID+1 {
				s.status = StatusDesynchronized
				atomic.AddUint64(&s.resyncRequests, 1)
				return fmt.Errorf("gap detected on snapshot replay: snapshot lastUpdateID=%d, delta firstUpdateID=%d", s.lastUpdateID, delta.FirstUpdateID)
			}
			firstApplicable = true
		}

		if err := s.applyDeltaInternalLocked(delta); err != nil {
			return err
		}
	}

	s.deltaBuffer = s.deltaBuffer[:0]
	s.status = StatusSynchronized
	return nil
}

func (s *OrderBookSync) applyDeltaLocked(delta BookDelta) error {
	// Sequence validation: delta.FirstUpdateID must equal lastUpdateID + 1
	if delta.FirstUpdateID != s.lastUpdateID+1 {
		s.status = StatusDesynchronized
		atomic.AddUint64(&s.resyncRequests, 1)
		return fmt.Errorf("sequence gap: expected %d, got %d. Triggering resync", s.lastUpdateID+1, delta.FirstUpdateID)
	}

	return s.applyDeltaInternalLocked(delta)
}

func (s *OrderBookSync) applyDeltaInternalLocked(delta BookDelta) error {
	for _, b := range delta.Bids {
		if b.QtyE8 == 0 {
			// Zero ghost liquidity: explicit 0 removes level
			delete(s.bids, b.PriceE8)
		} else {
			s.bids[b.PriceE8] = b.QtyE8
		}
	}

	for _, a := range delta.Asks {
		if a.QtyE8 == 0 {
			// Zero ghost liquidity: explicit 0 removes level
			delete(s.asks, a.PriceE8)
		} else {
			s.asks[a.PriceE8] = a.QtyE8
		}
	}

	s.lastUpdateID = delta.LastUpdateID
	return nil
}

// GetL2Depth returns top N price levels sorted properly
func (s *OrderBookSync) GetL2Depth(depth int) ([]PriceLevel, []PriceLevel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.status != StatusSynchronized {
		return nil, nil, errors.New("cannot fetch depth: orderbook not synchronized")
	}

	// Sort bids descending
	bidPrices := make([]uint64, 0, len(s.bids))
	for p := range s.bids {
		bidPrices = append(bidPrices, p)
	}
	sort.Slice(bidPrices, func(i, j int) bool {
		return bidPrices[i] > bidPrices[j]
	})

	bids := make([]PriceLevel, 0, depth)
	for i := 0; i < len(bidPrices) && i < depth; i++ {
		p := bidPrices[i]
		bids = append(bids, PriceLevel{
			PriceE8:         p,
			TotalQuantityE8: s.bids[p],
			OrderCount:      1,
		})
	}

	// Sort asks ascending
	askPrices := make([]uint64, 0, len(s.asks))
	for p := range s.asks {
		askPrices = append(askPrices, p)
	}
	sort.Slice(askPrices, func(i, j int) bool {
		return askPrices[i] < askPrices[j]
	})

	asks := make([]PriceLevel, 0, depth)
	for i := 0; i < len(askPrices) && i < depth; i++ {
		p := askPrices[i]
		asks = append(asks, PriceLevel{
			PriceE8:         p,
			TotalQuantityE8: s.asks[p],
			OrderCount:      1,
		})
	}

	return bids, asks, nil
}
