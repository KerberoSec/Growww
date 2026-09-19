package main

import (
	"testing"
	"time"
)

func TestVWAPSlicer_VolumeProfileDistribution(t *testing.T) {
	slicer := NewVWAPSlicer()

	totalQty := uint64(1000000000) // 10 units
	// Asymmetric volume profile curve representing higher volume at open and close (U-shape curve)
	curve := []float64{0.25, 0.10, 0.10, 0.15, 0.40}

	order, err := slicer.CreateVWAP(CreateVWAPRequest{
		ParentOrderID:      "vwap-prof-1",
		UserID:             "inst-1",
		Symbol:             "BTC-USDT",
		Side:               "BUY",
		TotalQuantityE8:    totalQty,
		VolumeProfileCurve: curve,
		IntervalDuration:   5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("CreateVWAP failed: %v", err)
	}

	if len(order.Slices) != 5 {
		t.Fatalf("Expected 5 slices, got %d", len(order.Slices))
	}

	// Verify exact volume conservation: sum of slices == totalQty
	var sumSlices uint64
	for _, sl := range order.Slices {
		sumSlices += sl.TargetQtyE8
	}
	if sumSlices != totalQty {
		t.Fatalf("Volume conservation violated: sum=%d, total=%d", sumSlices, totalQty)
	}

	// First slice should be ~25% (250,000,000), last slice should be ~40% (400,000,000)
	if order.Slices[0].TargetQtyE8 != 250000000 {
		t.Errorf("Expected slice 1 target 250000000, got %d", order.Slices[0].TargetQtyE8)
	}
	if order.Slices[4].TargetQtyE8 != 400000000 {
		t.Errorf("Expected slice 5 target 400000000, got %d", order.Slices[4].TargetQtyE8)
	}
}

func TestVWAPSlicer_DynamicRealTimeVolumeAdjustment(t *testing.T) {
	slicer := NewVWAPSlicer()

	order, _ := slicer.CreateVWAP(CreateVWAPRequest{
		ParentOrderID:    "vwap-adj-1",
		UserID:           "inst-2",
		Symbol:           "ETH-USDT",
		Side:             "SELL",
		TotalQuantityE8:  100000000,
		IntervalDuration: 5 * time.Minute,
		AggressionGamma:  0.8,
	})

	sliceID := order.Slices[0].SliceID
	origQty := order.Slices[0].TargetQtyE8

	// Market traded 20% MORE volume than expected: actual=120, expected=100
	adjQty, err := slicer.AdjustSliceForMarketVolume(order.ParentOrderID, sliceID, 120, 100)
	if err != nil {
		t.Fatalf("AdjustSliceForMarketVolume failed: %v", err)
	}

	if adjQty <= origQty {
		t.Errorf("Expected adjusted volume (%d) to be higher than original (%d) when market volume is ahead", adjQty, origQty)
	}
}

func TestVWAPSlicer_ExecutionAndCompletion(t *testing.T) {
	slicer := NewVWAPSlicer()

	order, _ := slicer.CreateVWAP(CreateVWAPRequest{
		ParentOrderID:      "vwap-comp-1",
		UserID:             "inst-3",
		Symbol:             "BTC-USDT",
		Side:               "BUY",
		TotalQuantityE8:    50000000,
		VolumeProfileCurve: []float64{0.5, 0.5},
		IntervalDuration:   1 * time.Minute,
	})

	sl1 := order.Slices[0]
	sl2 := order.Slices[1]

	_ = slicer.RecordSliceFill(order.ParentOrderID, sl1.SliceID, sl1.TargetQtyE8, 6000000000000)
	_ = slicer.RecordSliceFill(order.ParentOrderID, sl2.SliceID, sl2.TargetQtyE8, 6200000000000)

	snap, _ := slicer.GetOrder(order.ParentOrderID)
	if snap.Status != VWAPStatusCompleted {
		t.Errorf("Expected COMPLETED status, got %s", snap.Status)
	}
	if snap.ExecutedQuantityE8 != snap.TotalQuantityE8 {
		t.Errorf("Expected fully executed, got %d", snap.ExecutedQuantityE8)
	}
	expectedAvg := uint64(6100000000000)
	if snap.AveragePriceE8 != expectedAvg {
		t.Errorf("Expected average price %d, got %d", expectedAvg, snap.AveragePriceE8)
	}
}
