package main

import (
	"math/rand"
	"sync"
	"time"
)

// IcebergOrder represents a large parent order with hidden liquidity
type IcebergOrder struct {
	ParentOrderID string    `json:"parent_order_id"`
	UserID        string    `json:"user_id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	PriceE8       uint64    `json:"price_e8"`
	TotalQtyE8    uint64    `json:"total_qty_e8"`
	PeakQtyE8     uint64    `json:"peak_qty_e8"`    // Maximum visible clip
	VariancePct   uint32    `json:"variance_pct"`   // Random clip variance (e.g. 10%)
	ExecutedQtyE8 uint64    `json:"executed_qty_e8"`
	CurrentClipE8 uint64    `json:"current_clip_e8"`
	CreatedAt     time.Time `json:"created_at"`
}

type IcebergSlicer struct {
	mu     sync.Mutex
	orders map[string]*IcebergOrder
	rng    *rand.Rand
}

func NewIcebergSlicer() *IcebergSlicer {
	return &IcebergSlicer{
		orders: make(map[string]*IcebergOrder),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NextClip calculates the next child slice to submit to the public orderbook
func (s *IcebergSlicer) NextClip(parentID string) (uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return 0, false
	}

	remaining := order.TotalQtyE8 - order.ExecutedQtyE8
	if remaining == 0 {
		return 0, false // Fully completed
	}

	clip := order.PeakQtyE8
	if order.VariancePct > 0 {
		// Apply pseudo-random variance between (100 - Variance)% and (100 + Variance)%
		varianceRange := int64(order.PeakQtyE8 * uint64(order.VariancePct) / 100)
		jitter := (s.rng.Int63n(varianceRange*2 + 1)) - varianceRange
		clip = uint64(int64(clip) + jitter)
	}

	if clip > remaining {
		clip = remaining
	}

	order.CurrentClipE8 = clip
	return clip, true
}

// OnClipFilled records execution of the visible slice and readies the next slice
func (s *IcebergSlicer) OnClipFilled(parentID string, filledQty uint64) (uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return 0, false
	}

	order.ExecutedQtyE8 += filledQty
	if order.ExecutedQtyE8 >= order.TotalQtyE8 {
		delete(s.orders, parentID)
		return 0, false
	}

	return s.NextClip(parentID)
}
