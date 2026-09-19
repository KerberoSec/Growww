package main

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
	"time"
)

// VWAPStatus represents the lifecycle of a VWAP order
type VWAPStatus string

const (
	VWAPStatusActive    VWAPStatus = "ACTIVE"
	VWAPStatusPaused    VWAPStatus = "PAUSED"
	VWAPStatusCompleted VWAPStatus = "COMPLETED"
	VWAPStatusCanceled  VWAPStatus = "CANCELED"
)

// VWAPSlice represents a volume-weighted child tranche
type VWAPSlice struct {
	SliceID         string    `json:"slice_id"`
	ParentOrderID   string    `json:"parent_order_id"`
	IntervalIndex   uint32    `json:"interval_index"`
	ExpectedVolPct  float64   `json:"expected_vol_pct"`
	TargetQtyE8     uint64    `json:"target_qty_e8"`
	ExecutedQtyE8   uint64    `json:"executed_qty_e8"`
	LimitPriceCapE8 uint64    `json:"limit_price_cap_e8"`
	ExecutedPriceE8 uint64    `json:"executed_price_e8"`
	ScheduledTime   time.Time `json:"scheduled_time"`
	Status          string    `json:"status"` // "SCHEDULED", "DISPATCHED", "FILLED", "CANCELED"
}

// VWAPOrder represents an institutional Volume-Weighted Average Price execution strategy
type VWAPOrder struct {
	ParentOrderID       string       `json:"parent_order_id"`
	UserID              string       `json:"user_id"`
	Symbol              string       `json:"symbol"`
	Side                string       `json:"side"` // BUY or SELL
	TotalQuantityE8     uint64       `json:"total_quantity_e8"`
	ExecutedQuantityE8  uint64       `json:"executed_quantity_e8"`
	RemainingQuantityE8 uint64       `json:"remaining_quantity_e8"`
	AggressionGamma     float64      `json:"aggression_gamma"` // Typically 0.5 to 1.0
	LimitPriceCapE8     uint64       `json:"limit_price_cap_e8"`
	VolumeProfileCurve  []float64    `json:"volume_profile_curve"` // Normalized fraction per interval (sum == 1.0)
	IntervalDuration    time.Duration `json:"interval_duration"`
	Status              VWAPStatus   `json:"status"`
	AveragePriceE8      uint64       `json:"average_price_e8"`
	TotalNotionalE8     uint64       `json:"total_notional_e8"`
	StartTime           time.Time    `json:"start_time"`
	EndTime             time.Time    `json:"end_time"`
	Slices              []*VWAPSlice `json:"slices"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`
}

// VWAPSlicer manages VWAP execution schedules and dynamic participation correction
type VWAPSlicer struct {
	mu     sync.RWMutex
	orders map[string]*VWAPOrder
}

// NewVWAPSlicer constructs a new VWAP slicer
func NewVWAPSlicer() *VWAPSlicer {
	return &VWAPSlicer{
		orders: make(map[string]*VWAPOrder),
	}
}

// CreateVWAPRequest holds parameters to create a VWAP strategy
type CreateVWAPRequest struct {
	ParentOrderID      string
	UserID             string
	Symbol             string
	Side               string
	TotalQuantityE8    uint64
	AggressionGamma    float64
	LimitPriceCapE8    uint64
	VolumeProfileCurve []float64 // If empty, defaults to equal weights
	IntervalDuration   time.Duration
	StartTime          time.Time
}

// CreateVWAP initializes a VWAP order based on volume profile distribution
func (s *VWAPSlicer) CreateVWAP(req CreateVWAPRequest) (*VWAPOrder, error) {
	if req.ParentOrderID == "" {
		return nil, errors.New("parent_order_id required")
	}
	if req.TotalQuantityE8 == 0 {
		return nil, errors.New("total_quantity_e8 must be greater than zero")
	}
	if req.Side != "BUY" && req.Side != "SELL" {
		return nil, errors.New("side must be BUY or SELL")
	}
	if req.IntervalDuration <= 0 {
		return nil, errors.New("interval_duration must be positive")
	}
	if req.AggressionGamma <= 0 {
		req.AggressionGamma = 0.75 // Default balanced aggression
	}
	if len(req.VolumeProfileCurve) == 0 {
		// Default to 10 intervals with equal 10% weight
		req.VolumeProfileCurve = make([]float64, 10)
		for i := range req.VolumeProfileCurve {
			req.VolumeProfileCurve[i] = 0.10
		}
	}
	if req.StartTime.IsZero() {
		req.StartTime = time.Now().UTC()
	}

	// Normalize curve
	var curveSum float64
	for _, w := range req.VolumeProfileCurve {
		if w < 0 {
			return nil, errors.New("volume weights must be non-negative")
		}
		curveSum += w
	}
	if curveSum == 0 {
		return nil, errors.New("volume profile sum must be greater than zero")
	}
	normalizedCurve := make([]float64, len(req.VolumeProfileCurve))
	for i, w := range req.VolumeProfileCurve {
		normalizedCurve[i] = w / curveSum
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	numIntervals := len(normalizedCurve)
	duration := time.Duration(numIntervals) * req.IntervalDuration

	order := &VWAPOrder{
		ParentOrderID:       req.ParentOrderID,
		UserID:              req.UserID,
		Symbol:              req.Symbol,
		Side:                req.Side,
		TotalQuantityE8:     req.TotalQuantityE8,
		ExecutedQuantityE8:  0,
		RemainingQuantityE8: req.TotalQuantityE8,
		AggressionGamma:     req.AggressionGamma,
		LimitPriceCapE8:     req.LimitPriceCapE8,
		VolumeProfileCurve:  normalizedCurve,
		IntervalDuration:    req.IntervalDuration,
		Status:              VWAPStatusActive,
		AveragePriceE8:      0,
		TotalNotionalE8:     0,
		StartTime:           req.StartTime,
		EndTime:             req.StartTime.Add(duration),
		Slices:              make([]*VWAPSlice, 0, numIntervals),
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	remainingQty := req.TotalQuantityE8
	currentTime := req.StartTime

	for i, weight := range normalizedCurve {
		var sliceQty uint64
		if i == numIntervals-1 {
			sliceQty = remainingQty
		} else {
			sliceQty = uint64(float64(req.TotalQuantityE8) * weight)
			if sliceQty > remainingQty {
				sliceQty = remainingQty
			}
		}
		remainingQty -= sliceQty

		slice := &VWAPSlice{
			SliceID:         fmt.Sprintf("%s-vwap-%d", req.ParentOrderID, i+1),
			ParentOrderID:   req.ParentOrderID,
			IntervalIndex:   uint32(i + 1),
			ExpectedVolPct:  weight,
			TargetQtyE8:     sliceQty,
			ExecutedQtyE8:   0,
			LimitPriceCapE8: req.LimitPriceCapE8,
			ExecutedPriceE8: 0,
			ScheduledTime:   currentTime,
			Status:          "SCHEDULED",
		}
		order.Slices = append(order.Slices, slice)
		currentTime = currentTime.Add(req.IntervalDuration)
	}

	s.orders[order.ParentOrderID] = order
	return order, nil
}

// AdjustSliceForMarketVolume implements dynamic real-time tracking correction:
// beta(t) = actual_market_vol / expected_market_vol
// q_adj = q * beta(t)^gamma
func (s *VWAPSlicer) AdjustSliceForMarketVolume(parentID string, sliceID string, actualMarketVolE8, expectedMarketVolE8 uint64) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return 0, errors.New("vwap order not found")
	}

	var targetSlice *VWAPSlice
	for _, sl := range order.Slices {
		if sl.SliceID == sliceID {
			targetSlice = sl
			break
		}
	}
	if targetSlice == nil {
		return 0, errors.New("slice not found")
	}
	if targetSlice.Status != "SCHEDULED" {
		return targetSlice.TargetQtyE8, nil
	}

	if expectedMarketVolE8 > 0 && actualMarketVolE8 > 0 {
		beta := float64(actualMarketVolE8) / float64(expectedMarketVolE8)
		multiplier := math.Pow(beta, order.AggressionGamma)
		adjustedQty := uint64(float64(targetSlice.TargetQtyE8) * multiplier)

		// Clamp between 50% and 150% of nominal target
		minQty := targetSlice.TargetQtyE8 / 2
		maxQty := targetSlice.TargetQtyE8 * 3 / 2
		if adjustedQty < minQty {
			adjustedQty = minQty
		}
		if adjustedQty > maxQty {
			adjustedQty = maxQty
		}

		if adjustedQty > order.RemainingQuantityE8 {
			adjustedQty = order.RemainingQuantityE8
		}
		targetSlice.TargetQtyE8 = adjustedQty
	}

	return targetSlice.TargetQtyE8, nil
}

// RecordSliceFill updates slice and parent metrics upon fill
func (s *VWAPSlicer) RecordSliceFill(parentID string, sliceID string, executedQtyE8, executedPriceE8 uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return errors.New("vwap order not found")
	}

	var targetSlice *VWAPSlice
	for _, sl := range order.Slices {
		if sl.SliceID == sliceID {
			targetSlice = sl
			break
		}
	}
	if targetSlice == nil {
		return errors.New("slice not found")
	}

	targetSlice.ExecutedQtyE8 = executedQtyE8
	targetSlice.ExecutedPriceE8 = executedPriceE8
	targetSlice.Status = "FILLED"

	order.ExecutedQuantityE8 += executedQtyE8
	if order.RemainingQuantityE8 >= executedQtyE8 {
		order.RemainingQuantityE8 -= executedQtyE8
	} else {
		order.RemainingQuantityE8 = 0
	}

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
	order.UpdatedAt = time.Now().UTC()

	if order.ExecutedQuantityE8 >= order.TotalQuantityE8 {
		order.Status = VWAPStatusCompleted
	}

	return nil
}

// CancelVWAP cancels an active VWAP order and returns remaining unsliced volume
func (s *VWAPSlicer) CancelVWAP(parentID string) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[parentID]
	if !exists {
		return 0, errors.New("vwap order not found")
	}

	if order.Status == VWAPStatusCompleted || order.Status == VWAPStatusCanceled {
		return 0, fmt.Errorf("cannot cancel order in status %s", order.Status)
	}

	order.Status = VWAPStatusCanceled
	for _, sl := range order.Slices {
		if sl.Status == "SCHEDULED" {
			sl.Status = "CANCELED"
		}
	}

	remaining := order.RemainingQuantityE8
	order.RemainingQuantityE8 = 0
	order.UpdatedAt = time.Now().UTC()
	return remaining, nil
}

// GetOrder returns an active VWAP order snapshot
func (s *VWAPSlicer) GetOrder(parentID string) (*VWAPOrder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[parentID]
	if !exists {
		return nil, false
	}
	copied := *order
	return &copied, true
}
