package main

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type TWAPOrder struct {
	TWAPID          string        `json:"twap_id"`
	UserID          string        `json:"user_id"`
	Symbol          string        `json:"symbol"`
	Side            string        `json:"side"`
	TotalQuantityE8 uint64        `json:"total_quantity_e8"`
	Duration        time.Duration `json:"duration"`
	Interval        time.Duration `json:"interval"`
	SlicesTotal     uint32        `json:"slices_total"`
	SlicesExecuted  uint32        `json:"slices_executed"`
	FilledQtyE8     uint64        `json:"filled_qty_e8"`
	Active          bool          `json:"active"`
	StartTime       time.Time     `json:"start_time"`
}

type TWAPEngine struct {
	mu     sync.Mutex
	orders map[string]*TWAPOrder
	rng    *rand.Rand
}

func NewTWAPEngine() *TWAPEngine {
	return &TWAPEngine{
		orders: make(map[string]*TWAPOrder),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateTWAP initializes a new time-weighted execution plan
func (e *TWAPEngine) CreateTWAP(id, userID, symbol, side string, totalQtyE8 uint64, duration, interval time.Duration) (*TWAPOrder, error) {
	if totalQtyE8 == 0 || duration <= 0 || interval <= 0 {
		return nil, errors.New("invalid TWAP parameters")
	}

	numSlices := uint32(duration / interval)
	if numSlices == 0 {
		numSlices = 1
	}

	order := &TWAPOrder{
		TWAPID:          id,
		UserID:          userID,
		Symbol:          symbol,
		Side:            side,
		TotalQuantityE8: totalQtyE8,
		Duration:        duration,
		Interval:        interval,
		SlicesTotal:     numSlices,
		SlicesExecuted:  0,
		FilledQtyE8:     0,
		Active:          true,
		StartTime:       time.Now().UTC(),
	}

	e.mu.Lock()
	e.orders[id] = order
	e.mu.Unlock()

	return order, nil
}

// GetNextSliceQuantity calculates child order volume with volume randomized variance
func (e *TWAPEngine) GetNextSliceQuantity(twapID string) (uint64, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	order, exists := e.orders[twapID]
	if !exists || !order.Active {
		return 0, false
	}

	remainingQty := order.TotalQuantityE8 - order.FilledQtyE8
	remainingSlices := order.SlicesTotal - order.SlicesExecuted

	if remainingQty == 0 || remainingSlices == 0 {
		order.Active = false
		return 0, false
	}

	baseSlice := remainingQty / uint64(remainingSlices)
	// Random variance between 90% and 110% of base slice
	factor := 0.90 + (e.rng.Float64() * 0.20)
	sliceQty := uint64(float64(baseSlice) * factor)

	if sliceQty > remainingQty || remainingSlices == 1 {
		sliceQty = remainingQty
	}

	return sliceQty, true
}

// RecordSliceFill updates TWAP progression
func (e *TWAPEngine) RecordSliceFill(twapID string, filledQty uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if order, exists := e.orders[twapID]; exists {
		order.FilledQtyE8 += filledQty
		order.SlicesExecuted++
		if order.FilledQtyE8 >= order.TotalQuantityE8 || order.SlicesExecuted >= order.SlicesTotal {
			order.Active = false
		}
	}
}
