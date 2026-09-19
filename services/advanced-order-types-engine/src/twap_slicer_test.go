package main

import (
	"sync"
	"testing"
	"time"
)

func TestTWAPSlicer_ExactVolumeConservation(t *testing.T) {
	slicer := NewTWAPSlicerWithSeed(42)

	totalQty := uint64(10000000000) // 100.00000000 units in E8
	duration := 10 * time.Minute
	interval := 1 * time.Minute

	order, err := slicer.CreateTWAP(CreateTWAPRequest{
		ParentOrderID:       "twap-001",
		UserID:              "user-123",
		Symbol:              "BTC-USDT",
		Side:                "BUY",
		TotalQuantityE8:     totalQty,
		Duration:            duration,
		NominalInterval:     interval,
		LimitPriceCapE8:     7000000000000,
		EnablePoissonJitter: true,
		VariancePct:         20,
		StartTime:           time.Now().UTC(),
	})

	if err != nil {
		t.Fatalf("Failed to create TWAP order: %v", err)
	}

	if len(order.Slices) != 10 {
		t.Fatalf("Expected 10 slices, got %d", len(order.Slices))
	}

	// Verify exact volume conservation: sum(slices) must equal totalQty
	var sumSlices uint64
	for i, slice := range order.Slices {
		if slice.TargetQtyE8 == 0 {
			t.Errorf("Slice %d has zero quantity", i)
		}
		sumSlices += slice.TargetQtyE8
	}

	if sumSlices != totalQty {
		t.Fatalf("Volume conservation violated: sum=%d, total=%d, diff=%d", sumSlices, totalQty, int64(sumSlices)-int64(totalQty))
	}
}

func TestTWAPSlicer_DispatchScheduleAndDueSlices(t *testing.T) {
	slicer := NewTWAPSlicerWithSeed(100)

	startTime := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)
	interval := 30 * time.Second
	duration := 5 * time.Minute

	order, err := slicer.CreateTWAP(CreateTWAPRequest{
		ParentOrderID:       "twap-sched-1",
		UserID:              "user-1",
		Symbol:              "ETH-USDT",
		Side:                "BUY",
		TotalQuantityE8:     5000000000,
		Duration:            duration,
		NominalInterval:     interval,
		EnablePoissonJitter: false,
		VariancePct:         0,
		StartTime:           startTime,
	})
	if err != nil {
		t.Fatalf("CreateTWAP failed: %v", err)
	}

	// At start time, exactly 1 slice should be due
	due := slicer.GetDueSlices(startTime)
	if len(due) != 1 {
		t.Fatalf("Expected 1 slice due at start time, got %d", len(due))
	}
	if due[0].Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", due[0].Sequence)
	}

	// At start + 20s, no new slices due
	due2 := slicer.GetDueSlices(startTime.Add(20 * time.Second))
	if len(due2) != 0 {
		t.Fatalf("Expected 0 slices due at +20s, got %d", len(due2))
	}

	// At start + 30s, slice 2 should be due
	due3 := slicer.GetDueSlices(startTime.Add(30 * time.Second))
	if len(due3) != 1 {
		t.Fatalf("Expected 1 slice due at +30s, got %d", len(due3))
	}
	if due3[0].Sequence != 2 {
		t.Errorf("Expected sequence 2, got %d", due3[0].Sequence)
	}

	snap, _ := slicer.GetOrder(order.ParentOrderID)
	if snap.SlicesDispatched != 2 {
		t.Errorf("Expected 2 slices dispatched, got %d", snap.SlicesDispatched)
	}
}

func TestTWAPSlicer_PriceCapValidation(t *testing.T) {
	slicer := NewTWAPSlicer()

	// BUY order with cap of 65,000 USDT (in E8: 65000 * 1e8)
	buyCap := uint64(6500000000000)
	_, err := slicer.CreateTWAP(CreateTWAPRequest{
		ParentOrderID:   "twap-buy-cap",
		UserID:          "user-1",
		Symbol:          "BTC-USDT",
		Side:            "BUY",
		TotalQuantityE8: 100000000,
		Duration:        10 * time.Minute,
		NominalInterval: 1 * time.Minute,
		LimitPriceCapE8: buyCap,
	})
	if err != nil {
		t.Fatalf("CreateTWAP failed: %v", err)
	}

	// Market at 64,000 -> Valid
	valid, err := slicer.ValidatePriceCap("twap-buy-cap", 6400000000000)
	if err != nil || !valid {
		t.Errorf("Expected valid price cap check, got valid=%v, err=%v", valid, err)
	}

	// Market at 66,000 -> Invalid (breaches buy limit cap)
	valid, err = slicer.ValidatePriceCap("twap-buy-cap", 6600000000000)
	if err != nil || valid {
		t.Errorf("Expected invalid price cap check when market exceeds buy cap, got valid=%v", valid)
	}

	// SELL order with cap of 60,000 USDT
	sellCap := uint64(6000000000000)
	_, err = slicer.CreateTWAP(CreateTWAPRequest{
		ParentOrderID:   "twap-sell-cap",
		UserID:          "user-2",
		Symbol:          "BTC-USDT",
		Side:            "SELL",
		TotalQuantityE8: 100000000,
		Duration:        10 * time.Minute,
		NominalInterval: 1 * time.Minute,
		LimitPriceCapE8: sellCap,
	})
	if err != nil {
		t.Fatalf("CreateTWAP failed: %v", err)
	}

	// Market at 59,000 -> Invalid (below minimum sell price)
	valid, err = slicer.ValidatePriceCap("twap-sell-cap", 5900000000000)
	if err != nil || valid {
		t.Errorf("Expected invalid price cap check when market below sell cap, got valid=%v", valid)
	}
}

func TestTWAPSlicer_ExecutionAndVWAPRecalculation(t *testing.T) {
	slicer := NewTWAPSlicerWithSeed(7)

	totalQty := uint64(200000000) // 2 units
	order, _ := slicer.CreateTWAP(CreateTWAPRequest{
		ParentOrderID:   "twap-exec-1",
		UserID:          "u1",
		Symbol:          "ETH-USDT",
		Side:            "BUY",
		TotalQuantityE8: totalQty,
		Duration:        2 * time.Minute,
		NominalInterval: 1 * time.Minute,
		VariancePct:     0,
	})

	slice1 := order.Slices[0]
	slice2 := order.Slices[1]

	// Slice 1 executes: 1 unit at 3,000 USDT (3000 * 1e8)
	price1 := uint64(300000000000)
	err := slicer.RecordSliceExecution(order.ParentOrderID, slice1.SliceID, slice1.TargetQtyE8, price1)
	if err != nil {
		t.Fatalf("RecordSliceExecution failed: %v", err)
	}

	snap1, _ := slicer.GetOrder(order.ParentOrderID)
	if snap1.ExecutedQuantityE8 != slice1.TargetQtyE8 {
		t.Errorf("Expected executed %d, got %d", slice1.TargetQtyE8, snap1.ExecutedQuantityE8)
	}
	if snap1.AveragePriceE8 != price1 {
		t.Errorf("Expected average price %d, got %d", price1, snap1.AveragePriceE8)
	}
	if snap1.Status != TWAPStatusActive {
		t.Errorf("Expected status ACTIVE, got %s", snap1.Status)
	}

	// Slice 2 executes: 1 unit at 3,200 USDT (3200 * 1e8)
	price2 := uint64(320000000000)
	err = slicer.RecordSliceExecution(order.ParentOrderID, slice2.SliceID, slice2.TargetQtyE8, price2)
	if err != nil {
		t.Fatalf("RecordSliceExecution 2 failed: %v", err)
	}

	snap2, _ := slicer.GetOrder(order.ParentOrderID)
	expectedAvgPrice := (price1 + price2) / 2 // Exactly 3,100 USDT
	if snap2.AveragePriceE8 != expectedAvgPrice {
		t.Errorf("Expected VWAP average %d, got %d", expectedAvgPrice, snap2.AveragePriceE8)
	}
	if snap2.Status != TWAPStatusCompleted {
		t.Errorf("Expected status COMPLETED, got %s", snap2.Status)
	}
	if snap2.RemainingQuantityE8 != 0 {
		t.Errorf("Expected remaining quantity 0, got %d", snap2.RemainingQuantityE8)
	}
}

func TestTWAPSlicer_PauseResumeCancel(t *testing.T) {
	slicer := NewTWAPSlicer()

	order, _ := slicer.CreateTWAP(CreateTWAPRequest{
		ParentOrderID:   "twap-prc",
		UserID:          "u1",
		Symbol:          "BTC-USDT",
		Side:            "BUY",
		TotalQuantityE8: 1000000000,
		Duration:        5 * time.Minute,
		NominalInterval: 1 * time.Minute,
	})

	if err := slicer.PauseTWAP(order.ParentOrderID); err != nil {
		t.Fatalf("Pause failed: %v", err)
	}
	snap, _ := slicer.GetOrder(order.ParentOrderID)
	if snap.Status != TWAPStatusPaused {
		t.Errorf("Expected PAUSED, got %s", snap.Status)
	}

	if err := slicer.ResumeTWAP(order.ParentOrderID); err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	snap, _ = slicer.GetOrder(order.ParentOrderID)
	if snap.Status != TWAPStatusActive {
		t.Errorf("Expected ACTIVE, got %s", snap.Status)
	}

	remaining, err := slicer.CancelTWAP(order.ParentOrderID)
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}
	if remaining != snap.TotalQuantityE8 {
		t.Errorf("Expected cancelled remaining %d, got %d", snap.TotalQuantityE8, remaining)
	}

	snap, _ = slicer.GetOrder(order.ParentOrderID)
	if snap.Status != TWAPStatusCanceled {
		t.Errorf("Expected CANCELED, got %s", snap.Status)
	}
}

func TestTWAPSlicer_ConcurrentExecution(t *testing.T) {
	slicer := NewTWAPSlicerWithSeed(999)

	numOrders := 20
	var wg sync.WaitGroup

	for i := 0; i < numOrders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			parentID := "concurrent-twap-" + string(rune('A'+id))
			order, err := slicer.CreateTWAP(CreateTWAPRequest{
				ParentOrderID:   parentID,
				UserID:          "user-conc",
				Symbol:          "BTC-USDT",
				Side:            "BUY",
				TotalQuantityE8: 50000000,
				Duration:        5 * time.Minute,
				NominalInterval: 1 * time.Minute,
				VariancePct:     15,
			})
			if err != nil {
				t.Errorf("CreateTWAP failed: %v", err)
				return
			}

			// Simulate executing all slices
			for _, sl := range order.Slices {
				slicer.RecordSliceExecution(parentID, sl.SliceID, sl.TargetQtyE8, 6000000000000)
			}

			finalSnap, _ := slicer.GetOrder(parentID)
			if finalSnap.Status != TWAPStatusCompleted {
				t.Errorf("Order %s not completed: %s", parentID, finalSnap.Status)
			}
		}(i)
	}

	wg.Wait()
}
