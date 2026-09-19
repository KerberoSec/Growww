package main

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sync"
	"time"
)

// TWAPStatus represents the current state of a TWAP session
type TWAPStatus string

const (
	TWAPStatusPending   TWAPStatus = "PENDING"
	TWAPStatusActive    TWAPStatus = "ACTIVE"
	TWAPStatusPaused    TWAPStatus = "PAUSED"
	TWAPStatusCompleted TWAPStatus = "COMPLETED"
	TWAPStatusCanceled  TWAPStatus = "CANCELED"
)

// TWAPSliceStatus represents the execution state of an individual child slice
type TWAPSliceStatus string

const (
	TWAPSliceScheduled  TWAPSliceStatus = "SCHEDULED"
	TWAPSliceDispatched TWAPSliceStatus = "DISPATCHED"
	TWAPSliceFilled     TWAPSliceStatus = "FILLED"
	TWAPSliceCanceled   TWAPSliceStatus = "CANCELED"
)

// TWAPSlice represents a discrete child slice in the TWAP execution schedule
type TWAPSlice struct {
	SliceID         string          `json:"slice_id"`
	ParentOrderID   string          `json:"parent_order_id"`
	Sequence        uint32          `json:"sequence"`
	ScheduledTime   time.Time       `json:"scheduled_time"`
	DispatchedAt    time.Time       `json:"dispatched_at,omitempty"`
	CompletedAt     time.Time       `json:"completed_at,omitempty"`
	TargetQtyE8     uint64          `json:"target_qty_e8"`
	ExecutedQtyE8   uint64          `json:"executed_qty_e8"`
	LimitPriceCapE8 uint64          `json:"limit_price_cap_e8"`
	ExecutedPriceE8 uint64          `json:"executed_price_e8"`
	Status          TWAPSliceStatus `json:"status"`
}

// TWAPOrder represents the parent TWAP order and execution schedule
type TWAPOrder struct {
	ParentOrderID       string        `json:"parent_order_id"`
	UserID              string        `json:"user_id"`
	Symbol              string        `json:"symbol"`
	Side                string        `json:"side"` // BUY or SELL
	TotalQuantityE8     uint64        `json:"total_quantity_e8"`
	ExecutedQuantityE8  uint64        `json:"executed_quantity_e8"`
	RemainingQuantityE8 uint64        `json:"remaining_quantity_e8"`
	Duration            time.Duration `json:"duration"`
	NominalInterval     time.Duration `json:"nominal_interval"`
	SlicesTotal         uint32        `json:"slices_total"`
	SlicesDispatched    uint32        `json:"slices_dispatched"`
	SlicesFilled        uint32        `json:"slices_filled"`
	LimitPriceCapE8     uint64        `json:"limit_price_cap_e8"` // Price collar (Max Buy / Min Sell)
	EnablePoissonJitter bool          `json:"enable_poisson_jitter"`
	VariancePct         uint32        `json:"variance_pct"` // +/- slice size variance percentage
	Status              TWAPStatus    `json:"status"`
	AveragePriceE8      uint64        `json:"average_price_e8"`
	TotalNotionalE8     uint64        `json:"total_notional_e8"`
	StartTime           time.Time     `json:"start_time"`
	EndTime             time.Time     `json:"end_time"`
	NextDispatchTime    time.Time     `json:"next_dispatch_time"`
	Slices              []*TWAPSlice  `json:"slices"`
	CreatedAt           time.Time     `json:"created_at"`
	UpdatedAt           time.Time     `json:"updated_at"`
}

// CreateTWAPRequest defines parameters required to construct a new TWAP session
type CreateTWAPRequest struct {
	ParentOrderID       string
	UserID              string
	Symbol              string
	Side                string
	TotalQuantityE8     uint64
	Duration            time.Duration
	NominalInterval     time.Duration
	LimitPriceCapE8     uint64
	EnablePoissonJitter bool
	VariancePct         uint32
	StartTime           time.Time
}

// TWAPSlicer coordinates TWAP parent order lifecycle, slicing schedules, and dispatching
type TWAPSlicer struct {
	mu     sync.RWMutex
	orders map[string]*TWAPOrder
	rng    *rand.Rand
}

// NewTWAPSlicer constructs a new TWAP slicing coordinator
func NewTWAPSlicer() *TWAPSlicer {
	return &TWAPSlicer{
		orders: make(map[string]*TWAPOrder),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewTWAPSlicerWithSeed constructs a deterministic slicer for testing
func NewTWAPSlicerWithSeed(seed int64) *TWAPSlicer {
	return &TWAPSlicer{
		orders: make(map[string]*TWAPOrder),
		rng:    rand.New(rand.NewSource(seed)),
	}
}

// CreateTWAP validates and compiles an institutional TWAP execution schedule
func (s *TWAPSlicer) CreateTWAP(req CreateTWAPRequest) (*TWAPOrder, error) {
	if req.ParentOrderID == "" {
		return nil, errors.New("parent_order_id cannot be empty")
	}
	if req.TotalQuantityE8 == 0 {
		return nil, errors.New("total_quantity_e8 must be greater than zero")
	}
	if req.Duration <= 0 {
		return nil, errors.New("duration must be greater than zero")
	}
	if req.NominalInterval <= 0 {
		return nil, errors.New("nominal_interval must be greater than zero")
	}
	if req.NominalInterval > req.Duration {
		req.NominalInterval = req.Duration
	}
	if req.Side != "BUY" && req.Side != "SELL" {
		return nil, errors.New("side must be BUY or SELL")
	}
	if req.StartTime.IsZero() {
		req.StartTime = time.Now().UTC()
	}

	numSlices := uint32(req.Duration / req.NominalInterval)
	if numSlices == 0 {
		numSlices = 1
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order := &TWAPOrder{
		ParentOrderID:       req.ParentOrderID,
		UserID:              req.UserID,
		Symbol:              req.Symbol,
		Side:                req.Side,
		TotalQuantityE8:     req.TotalQuantityE8,
		ExecutedQuantityE8:  0,
		RemainingQuantityE8: req.TotalQuantityE8,
		Duration:            req.Duration,
		NominalInterval:     req.NominalInterval,
		SlicesTotal:         numSlices,
		SlicesDispatched:    0,
		SlicesFilled:        0,
		LimitPriceCapE8:     req.LimitPriceCapE8,
		EnablePoissonJitter: req.EnablePoissonJitter,
		VariancePct:         req.VariancePct,
		Status:              TWAPStatusActive,
		AveragePriceE8:      0,
		TotalNotionalE8:     0,
		StartTime:           req.StartTime,
		EndTime:             req.StartTime.Add(req.Duration),
		NextDispatchTime:    req.StartTime,
		Slices:              make([]*TWAPSlice, 0, numSlices),
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	// Pre-generate slice schedule with randomized volume and time jitter
	remainingVol := req.TotalQuantityE8
	currentTime := req.StartTime

	for i := uint32(1); i <= numSlices; i++ {
		slicesLeft := numSlices - i + 1
		var sliceVol uint64

		if slicesLeft == 1 {
			// Final slice absorbs all remaining quantity to ensure exact volume conservation
			sliceVol = remainingVol
		} else {
			baseVol := remainingVol / uint64(slicesLeft)
			if req.VariancePct > 0 {
				varRange := int64(baseVol * uint64(req.VariancePct) / 100)
				if varRange > 0 {
					jitter := (s.rng.Int63n(varRange*2 + 1)) - varRange
					newVol := int64(baseVol) + jitter
					if newVol > 0 {
						sliceVol = uint64(newVol)
					} else {
						sliceVol = baseVol
					}
				} else {
					sliceVol = baseVol
				}
			} else {
				sliceVol = baseVol
			}

			// Do not exceed remaining
			if sliceVol > remainingVol {
				sliceVol = remainingVol
			}
			// Keep at least 1 satoshi/unit per slice if remaining allows
			if sliceVol == 0 && remainingVol >= uint64(slicesLeft) {
				sliceVol = 1
			}
		}

		remainingVol -= sliceVol

		// Calculate interval jitter
		interval := req.NominalInterval
		if req.EnablePoissonJitter && i > 1 {
			// +/- 20% Poisson-like uniform timing jitter
			jitterMs := int64(interval.Milliseconds()) / 5
			if jitterMs > 0 {
				delta := s.rng.Int63n(jitterMs*2+1) - jitterMs
				interval = interval + time.Duration(delta)*time.Millisecond
			}
		}

		slice := &TWAPSlice{
			SliceID:         fmt.Sprintf("%s-slice-%d", req.ParentOrderID, i),
			ParentOrderID:   req.ParentOrderID,
			Sequence:        i,
			ScheduledTime:   currentTime,
			TargetQtyE8:     sliceVol,
			ExecutedQtyE8:   0,
			LimitPriceCapE8: req.LimitPriceCapE8,
			ExecutedPriceE8: 0,
			Status:          TWAPSliceScheduled,
		}

		order.Slices = append(order.Slices, slice)
		currentTime = currentTime.Add(interval)
	}

	if len(order.Slices) > 0 {
		order.NextDispatchTime = order.Slices[0].ScheduledTime
	}

	s.orders[order.ParentOrderID] = order
	return order, nil
}

// GetDueSlices returns all scheduled slices that are due for dispatch as of timestamp `now`
func (s *TWAPSlicer) GetDueSlices(now time.Time) []*TWAPSlice {
	s.mu.Lock()
	defer s.mu.Unlock()

	var due []*TWAPSlice

	for _, order := range s.orders {
		if order.Status != TWAPStatusActive {
			continue
		}

		for _, slice := range order.Slices {
			if slice.Status == TWAPSliceScheduled && !slice.ScheduledTime.After(now) {
				slice.Status = TWAPSliceDispatched
				slice.DispatchedAt = now
				order.SlicesDispatched++
				order.UpdatedAt = now
				due = append(due, slice)

				// Update next dispatch time
				foundNext := false
				for _, nextSlice := range order.Slices {
					if nextSlice.Status == TWAPSliceScheduled {
						order.NextDispatchTime = nextSlice.ScheduledTime
						foundNext = true
						break
					}
				}
				if !foundNext {
					order.NextDispatchTime = order.EndTime
				}
			}
		}
	}

	return due
}

// GetNextSlice returns the next slice due for dispatch for a specific parent order
func (s *TWAPSlicer) GetNextSlice(parentID string, now time.Time) (*TWAPSlice, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists || order.Status != TWAPStatusActive {
		return nil, false
	}

	for _, slice := range order.Slices {
		if slice.Status == TWAPSliceScheduled && !slice.ScheduledTime.After(now) {
			slice.Status = TWAPSliceDispatched
			slice.DispatchedAt = now
			order.SlicesDispatched++
			order.UpdatedAt = now

			// Update next dispatch time
			foundNext := false
			for _, nextSlice := range order.Slices {
				if nextSlice.Status == TWAPSliceScheduled {
					order.NextDispatchTime = nextSlice.ScheduledTime
					foundNext = true
					break
				}
			}
			if !foundNext {
				order.NextDispatchTime = order.EndTime
			}

			return slice, true
		}
	}

	return nil, false
}

// ValidatePriceCap checks whether the given market price satisfies the parent TWAP limit collar
func (s *TWAPSlicer) ValidatePriceCap(parentID string, currentPriceE8 uint64) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[parentID]
	if !exists {
		return false, errors.New("twap order not found")
	}

	if order.LimitPriceCapE8 == 0 {
		return true, nil // No price collar specified
	}

	if order.Side == "BUY" && currentPriceE8 > order.LimitPriceCapE8 {
		return false, nil // Market price exceeds maximum allowed buy price
	}

	if order.Side == "SELL" && currentPriceE8 < order.LimitPriceCapE8 {
		return false, nil // Market price is below minimum allowed sell price
	}

	return true, nil
}

// RecordSliceExecution updates fill details for a child slice and recalculates parent metrics
func (s *TWAPSlicer) RecordSliceExecution(parentID string, sliceID string, executedQtyE8 uint64, executedPriceE8 uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return errors.New("twap order not found")
	}

	var targetSlice *TWAPSlice
	for _, sl := range order.Slices {
		if sl.SliceID == sliceID {
			targetSlice = sl
			break
		}
	}

	if targetSlice == nil {
		return errors.New("slice not found")
	}

	now := time.Now().UTC()
	targetSlice.ExecutedQtyE8 += executedQtyE8
	targetSlice.ExecutedPriceE8 = executedPriceE8
	targetSlice.Status = TWAPSliceFilled
	targetSlice.CompletedAt = now

	// Update parent order stats
	order.ExecutedQuantityE8 += executedQtyE8
	if order.RemainingQuantityE8 >= executedQtyE8 {
		order.RemainingQuantityE8 -= executedQtyE8
	} else {
		order.RemainingQuantityE8 = 0
	}
	order.SlicesFilled++

	// Recalculate Volume Weighted Average Price (VWAP) across all fills using math/big to prevent overflow
	var bigQty, bigPrice, bigSliceNotional, e8 big.Int
	bigQty.SetUint64(executedQtyE8)
	bigPrice.SetUint64(executedPriceE8)
	e8.SetUint64(1e8)

	bigSliceNotional.Mul(&bigQty, &bigPrice)
	bigSliceNotional.Div(&bigSliceNotional, &e8)
	sliceNotional := bigSliceNotional.Uint64()
	order.TotalNotionalE8 += sliceNotional

	if order.ExecutedQuantityE8 > 0 {
		var bigTotalNotional, bigTotalQty, bigAvgPrice big.Int
		bigTotalNotional.SetUint64(order.TotalNotionalE8)
		bigTotalQty.SetUint64(order.ExecutedQuantityE8)

		bigAvgPrice.Mul(&bigTotalNotional, &e8)
		bigAvgPrice.Div(&bigAvgPrice, &bigTotalQty)
		order.AveragePriceE8 = bigAvgPrice.Uint64()
	}

	order.UpdatedAt = now

	// Check if parent order is completed
	if order.ExecutedQuantityE8 >= order.TotalQuantityE8 || order.SlicesFilled >= order.SlicesTotal {
		order.Status = TWAPStatusCompleted
	}

	return nil
}

// PauseTWAP halts dispatching for a running TWAP order
func (s *TWAPSlicer) PauseTWAP(parentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return errors.New("twap order not found")
	}
	if order.Status != TWAPStatusActive {
		return fmt.Errorf("cannot pause order in status %s", order.Status)
	}

	order.Status = TWAPStatusPaused
	order.UpdatedAt = time.Now().UTC()
	return nil
}

// ResumeTWAP unpauses a paused TWAP order
func (s *TWAPSlicer) ResumeTWAP(parentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return errors.New("twap order not found")
	}
	if order.Status != TWAPStatusPaused {
		return fmt.Errorf("cannot resume order in status %s", order.Status)
	}

	order.Status = TWAPStatusActive
	order.UpdatedAt = time.Now().UTC()
	return nil
}

// CancelTWAP cancels any remaining unscheduled slices and returns remaining quantity to release holds
func (s *TWAPSlicer) CancelTWAP(parentID string) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return 0, errors.New("twap order not found")
	}
	if order.Status == TWAPStatusCompleted || order.Status == TWAPStatusCanceled {
		return 0, fmt.Errorf("cannot cancel order in status %s", order.Status)
	}

	now := time.Now().UTC()
	order.Status = TWAPStatusCanceled
	order.UpdatedAt = now

	for _, sl := range order.Slices {
		if sl.Status == TWAPSliceScheduled {
			sl.Status = TWAPSliceCanceled
			sl.CompletedAt = now
		}
	}

	remaining := order.RemainingQuantityE8
	order.RemainingQuantityE8 = 0
	return remaining, nil
}

// GetOrder returns a deep copy of the TWAP order state
func (s *TWAPSlicer) GetOrder(parentID string) (*TWAPOrder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[parentID]
	if !exists {
		return nil, false
	}

	copied := *order
	copied.Slices = make([]*TWAPSlice, len(order.Slices))
	for i, sl := range order.Slices {
		slCopy := *sl
		copied.Slices[i] = &slCopy
	}
	return &copied, true
}

// ListOrders returns all registered TWAP orders
func (s *TWAPSlicer) ListOrders() []*TWAPOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*TWAPOrder, 0, len(s.orders))
	for _, o := range s.orders {
		copied := *o
		list = append(list, &copied)
	}
	return list
}
