package main

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// IcebergStatus defines the lifecycle status of an Iceberg order
type IcebergStatus string

const (
	IcebergStatusActive    IcebergStatus = "ACTIVE"
	IcebergStatusCompleted IcebergStatus = "COMPLETED"
	IcebergStatusCanceled  IcebergStatus = "CANCELED"
)

// IcebergClip represents an active visible child tranche on the orderbook
type IcebergClip struct {
	ClipID        string    `json:"clip_id"`
	ParentOrderID string    `json:"parent_order_id"`
	Sequence      uint32    `json:"sequence"`
	PriceE8       uint64    `json:"price_e8"`
	VisibleQtyE8  uint64    `json:"visible_qty_e8"`
	ExecutedQtyE8 uint64    `json:"executed_qty_e8"`
	Status        string    `json:"status"` // "OPEN", "FILLED", "CANCELED"
	CreatedAt     time.Time `json:"created_at"`
	CompletedAt   time.Time `json:"completed_at,omitempty"`
}

// IcebergOrder represents a large parent order with hidden liquidity
type IcebergOrder struct {
	ParentOrderID string         `json:"parent_order_id"`
	UserID        string         `json:"user_id"`
	Symbol        string         `json:"symbol"`
	Side          string         `json:"side"` // BUY or SELL
	PriceE8       uint64         `json:"price_e8"`
	TotalQtyE8    uint64         `json:"total_qty_e8"`
	PeakQtyE8     uint64         `json:"peak_qty_e8"`    // Maximum / nominal visible clip
	VariancePct   uint32         `json:"variance_pct"`   // Random clip variance (e.g. 15 for +/-15%)
	ExecutedQtyE8 uint64         `json:"executed_qty_e8"`
	CurrentClipE8 uint64         `json:"current_clip_e8"`
	Status        IcebergStatus  `json:"status"`
	ActiveClip    *IcebergClip   `json:"active_clip,omitempty"`
	ClipHistory   []*IcebergClip `json:"clip_history,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// IcebergSlicer coordinates slicing and display tranche refresh
type IcebergSlicer struct {
	mu     sync.Mutex
	orders map[string]*IcebergOrder
	rng    *rand.Rand
}

// NewIcebergSlicer constructs a thread-safe IcebergSlicer
func NewIcebergSlicer() *IcebergSlicer {
	return &IcebergSlicer{
		orders: make(map[string]*IcebergOrder),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewIcebergSlicerWithSeed constructs a deterministic slicer for unit testing
func NewIcebergSlicerWithSeed(seed int64) *IcebergSlicer {
	return &IcebergSlicer{
		orders: make(map[string]*IcebergOrder),
		rng:    rand.New(rand.NewSource(seed)),
	}
}

// RegisterOrder validates and registers a new iceberg order
func (s *IcebergSlicer) RegisterOrder(order *IcebergOrder) error {
	if order == nil {
		return errors.New("order cannot be nil")
	}
	if order.ParentOrderID == "" {
		return errors.New("parent_order_id required")
	}
	if order.TotalQtyE8 == 0 {
		return errors.New("total_qty_e8 must be greater than zero")
	}
	if order.PeakQtyE8 == 0 {
		return errors.New("peak_qty_e8 must be greater than zero")
	}
	if order.PeakQtyE8 > order.TotalQtyE8 {
		order.PeakQtyE8 = order.TotalQtyE8
	}
	if order.Side != "BUY" && order.Side != "SELL" {
		return errors.New("side must be BUY or SELL")
	}
	if order.PriceE8 == 0 {
		return errors.New("price_e8 must be greater than zero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	order.Status = IcebergStatusActive
	order.CreatedAt = now
	order.UpdatedAt = now
	order.ExecutedQtyE8 = 0
	order.ClipHistory = make([]*IcebergClip, 0)

	s.orders[order.ParentOrderID] = order
	return nil
}

// calculateNextClipLocked computes next clip size with randomized display variance while holding lock
func (s *IcebergSlicer) calculateNextClipLocked(order *IcebergOrder) (uint64, bool) {
	if order.Status != IcebergStatusActive {
		return 0, false
	}

	remaining := order.TotalQtyE8 - order.ExecutedQtyE8
	if remaining == 0 {
		order.Status = IcebergStatusCompleted
		order.CurrentClipE8 = 0
		order.ActiveClip = nil
		return 0, false
	}

	clip := order.PeakQtyE8
	if order.VariancePct > 0 {
		// Apply pseudo-random variance between (100 - Variance)% and (100 + Variance)%
		varianceRange := int64(order.PeakQtyE8 * uint64(order.VariancePct) / 100)
		if varianceRange > 0 {
			jitter := (s.rng.Int63n(varianceRange*2 + 1)) - varianceRange
			newClip := int64(clip) + jitter
			if newClip > 0 {
				clip = uint64(newClip)
			}
		}
	}

	if clip > remaining {
		clip = remaining
	}

	// Guarantee minimum of 1 unit if remaining > 0
	if clip == 0 && remaining > 0 {
		clip = 1
	}

	order.CurrentClipE8 = clip
	seq := uint32(len(order.ClipHistory) + 1)
	clipObj := &IcebergClip{
		ClipID:        fmt.Sprintf("%s-clip-%d", order.ParentOrderID, seq),
		ParentOrderID: order.ParentOrderID,
		Sequence:      seq,
		PriceE8:       order.PriceE8,
		VisibleQtyE8:  clip,
		ExecutedQtyE8: 0,
		Status:        "OPEN",
		CreatedAt:     time.Now().UTC(),
	}
	order.ActiveClip = clipObj
	order.ClipHistory = append(order.ClipHistory, clipObj)
	order.UpdatedAt = time.Now().UTC()

	return clip, true
}

// NextClip calculates the next child slice to submit to the public orderbook
func (s *IcebergSlicer) NextClip(parentID string) (uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return 0, false
	}

	return s.calculateNextClipLocked(order)
}

// NextClipDetails returns the full structured IcebergClip object
func (s *IcebergSlicer) NextClipDetails(parentID string) (*IcebergClip, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return nil, false
	}

	_, ok := s.calculateNextClipLocked(order)
	if !ok {
		return nil, false
	}
	return order.ActiveClip, true
}

// OnClipPartialFill records partial fill on the current active clip
func (s *IcebergSlicer) OnClipPartialFill(parentID string, filledQty uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return errors.New("iceberg order not found")
	}
	if order.Status != IcebergStatusActive {
		return fmt.Errorf("iceberg order is %s", order.Status)
	}

	order.ExecutedQtyE8 += filledQty
	if order.ActiveClip != nil {
		order.ActiveClip.ExecutedQtyE8 += filledQty
	}
	order.UpdatedAt = time.Now().UTC()

	if order.ExecutedQtyE8 >= order.TotalQtyE8 {
		order.Status = IcebergStatusCompleted
		if order.ActiveClip != nil {
			order.ActiveClip.Status = "FILLED"
			order.ActiveClip.CompletedAt = time.Now().UTC()
		}
	}
	return nil
}

// OnClipFilled records execution of the visible slice and refreshes the next slice without deadlocking
func (s *IcebergSlicer) OnClipFilled(parentID string, filledQty uint64) (uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return 0, false
	}

	order.ExecutedQtyE8 += filledQty
	if order.ActiveClip != nil {
		order.ActiveClip.ExecutedQtyE8 += filledQty
		order.ActiveClip.Status = "FILLED"
		order.ActiveClip.CompletedAt = time.Now().UTC()
	}

	if order.ExecutedQtyE8 >= order.TotalQtyE8 {
		order.Status = IcebergStatusCompleted
		order.CurrentClipE8 = 0
		order.ActiveClip = nil
		order.UpdatedAt = time.Now().UTC()
		return 0, false
	}

	// Refresh next clip using internal lock-free helper to avoid deadlocks!
	return s.calculateNextClipLocked(order)
}

// OnClipFilledDetails performs the clip fill and returns the newly refreshed IcebergClip
func (s *IcebergSlicer) OnClipFilledDetails(parentID string, filledQty uint64) (*IcebergClip, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return nil, false
	}

	order.ExecutedQtyE8 += filledQty
	if order.ActiveClip != nil {
		order.ActiveClip.ExecutedQtyE8 += filledQty
		order.ActiveClip.Status = "FILLED"
		order.ActiveClip.CompletedAt = time.Now().UTC()
	}

	if order.ExecutedQtyE8 >= order.TotalQtyE8 {
		order.Status = IcebergStatusCompleted
		order.CurrentClipE8 = 0
		order.ActiveClip = nil
		order.UpdatedAt = time.Now().UTC()
		return nil, false
	}

	_, ok := s.calculateNextClipLocked(order)
	if !ok {
		return nil, false
	}
	return order.ActiveClip, true
}

// CancelOrder cancels remaining unsliced quantity and any resting clip
func (s *IcebergSlicer) CancelOrder(parentID string) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return 0, errors.New("iceberg order not found")
	}

	if order.Status != IcebergStatusActive {
		return 0, fmt.Errorf("cannot cancel order in status %s", order.Status)
	}

	remaining := order.TotalQtyE8 - order.ExecutedQtyE8
	order.Status = IcebergStatusCanceled
	if order.ActiveClip != nil && order.ActiveClip.Status == "OPEN" {
		order.ActiveClip.Status = "CANCELED"
		order.ActiveClip.CompletedAt = time.Now().UTC()
	}
	order.CurrentClipE8 = 0
	order.UpdatedAt = time.Now().UTC()

	return remaining, nil
}

// GetOrder returns a read-only snapshot of an iceberg order
func (s *IcebergSlicer) GetOrder(parentID string) (*IcebergOrder, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external mutation
	copied := *order
	return &copied, true
}

// ListOrders returns all registered iceberg orders
func (s *IcebergSlicer) ListOrders() []*IcebergOrder {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := make([]*IcebergOrder, 0, len(s.orders))
	for _, o := range s.orders {
		copied := *o
		list = append(list, &copied)
	}
	return list
}
