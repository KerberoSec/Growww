package main

import (
	"sync"
	"testing"
	"time"
)

func TestIcebergSlicer_NoDeadlockOnClipFilled(t *testing.T) {
	// Verify that OnClipFilled completes within 100ms and does not deadlock
	slicer := NewIcebergSlicerWithSeed(42)

	order := &IcebergOrder{
		ParentOrderID: "ice-deadlock-test",
		UserID:        "user-1",
		Symbol:        "BTC-USDT",
		Side:          "BUY",
		PriceE8:       6000000000000,
		TotalQtyE8:    500000000, // 5 units
		PeakQtyE8:     100000000, // 1 unit per clip
		VariancePct:   10,
	}

	if err := slicer.RegisterOrder(order); err != nil {
		t.Fatalf("RegisterOrder failed: %v", err)
	}

	clipQty, ok := slicer.NextClip(order.ParentOrderID)
	if !ok || clipQty == 0 {
		t.Fatalf("Failed to generate initial clip: ok=%v, clip=%d", ok, clipQty)
	}

	done := make(chan bool)
	go func() {
		// This previously deadlocked due to re-acquiring s.mu in NextClip!
		nextClip, hasMore := slicer.OnClipFilled(order.ParentOrderID, clipQty)
		if !hasMore || nextClip == 0 {
			t.Errorf("Expected next clip after first fill, got hasMore=%v, nextClip=%d", hasMore, nextClip)
		}
		done <- true
	}()

	select {
	case <-done:
		// Success, no deadlock!
	case <-time.After(500 * time.Millisecond):
		t.Fatal("DEADLOCK DETECTED: OnClipFilled did not return within 500ms!")
	}
}

func TestIcebergSlicer_DisplaySizeRefreshUntilExhausted(t *testing.T) {
	slicer := NewIcebergSlicerWithSeed(123)

	totalQty := uint64(500000000) // 5 units
	peakQty := uint64(100000000)  // 1 unit
	order := &IcebergOrder{
		ParentOrderID: "ice-refresh-1",
		UserID:        "user-2",
		Symbol:        "ETH-USDT",
		Side:          "SELL",
		PriceE8:       300000000000,
		TotalQtyE8:    totalQty,
		PeakQtyE8:     peakQty,
		VariancePct:   15,
	}

	if err := slicer.RegisterOrder(order); err != nil {
		t.Fatalf("RegisterOrder failed: %v", err)
	}

	var cumulativeFilled uint64
	currentClip, ok := slicer.NextClip(order.ParentOrderID)
	if !ok {
		t.Fatal("Initial clip creation failed")
	}

	iterations := 0
	for ok && currentClip > 0 {
		iterations++
		cumulativeFilled += currentClip

		// Verify clip does not exceed remaining volume
		remainingBefore := totalQty - (cumulativeFilled - currentClip)
		if currentClip > remainingBefore {
			t.Fatalf("Clip %d exceeded remaining quantity %d", currentClip, remainingBefore)
		}

		// Fill the current clip and refresh next
		nextClip, hasMore := slicer.OnClipFilled(order.ParentOrderID, currentClip)
		currentClip = nextClip
		ok = hasMore

		if iterations > 20 {
			t.Fatal("Runaway loop detected in iceberg clip refresh")
		}
	}

	if cumulativeFilled != totalQty {
		t.Fatalf("Expected cumulative filled %d, got %d", totalQty, cumulativeFilled)
	}

	snap, exists := slicer.GetOrder(order.ParentOrderID)
	if !exists {
		t.Fatal("Order not found after completion")
	}
	if snap.Status != IcebergStatusCompleted {
		t.Fatalf("Expected status COMPLETED, got %s", snap.Status)
	}
	if snap.ExecutedQtyE8 != totalQty {
		t.Fatalf("Expected executed %d, got %d", totalQty, snap.ExecutedQtyE8)
	}
}

func TestIcebergSlicer_PartialFillHandling(t *testing.T) {
	slicer := NewIcebergSlicerWithSeed(55)

	order := &IcebergOrder{
		ParentOrderID: "ice-partial-1",
		UserID:        "user-3",
		Symbol:        "SOL-USDT",
		Side:          "BUY",
		PriceE8:       15000000000,
		TotalQtyE8:    1000000000, // 10 units
		PeakQtyE8:     200000000,  // 2 units
		VariancePct:   0,
	}
	_ = slicer.RegisterOrder(order)

	clip, _ := slicer.NextClip(order.ParentOrderID)
	if clip != 200000000 {
		t.Fatalf("Expected clip 200000000, got %d", clip)
	}

	// Partial fill of 50,000,000 (0.5 units)
	partialQty := uint64(50000000)
	if err := slicer.OnClipPartialFill(order.ParentOrderID, partialQty); err != nil {
		t.Fatalf("Partial fill failed: %v", err)
	}

	snap, _ := slicer.GetOrder(order.ParentOrderID)
	if snap.ExecutedQtyE8 != partialQty {
		t.Errorf("Expected executed %d, got %d", partialQty, snap.ExecutedQtyE8)
	}
	if snap.ActiveClip.ExecutedQtyE8 != partialQty {
		t.Errorf("Expected active clip executed %d, got %d", partialQty, snap.ActiveClip.ExecutedQtyE8)
	}
	if snap.ActiveClip.Status != "OPEN" {
		t.Errorf("Expected active clip to remain OPEN, got %s", snap.ActiveClip.Status)
	}

	// Fill the remaining 150,000,000 on the clip
	remainingClip := clip - partialQty
	nextClip, hasMore := slicer.OnClipFilled(order.ParentOrderID, remainingClip)
	if !hasMore || nextClip == 0 {
		t.Fatalf("Expected next clip to be refreshed: hasMore=%v, nextClip=%d", hasMore, nextClip)
	}

	snap2, _ := slicer.GetOrder(order.ParentOrderID)
	if snap2.ExecutedQtyE8 != 200000000 {
		t.Errorf("Expected total executed 200000000, got %d", snap2.ExecutedQtyE8)
	}
}

func TestIcebergSlicer_RemainderSmallerThanPeak(t *testing.T) {
	slicer := NewIcebergSlicerWithSeed(99)

	// Total 250 units, Peak 100 units -> Slices should be ~100, ~100, and last slice exactly 50!
	order := &IcebergOrder{
		ParentOrderID: "ice-remainder-1",
		UserID:        "user-4",
		Symbol:        "BTC-USDT",
		Side:          "SELL",
		PriceE8:       6500000000000,
		TotalQtyE8:    250,
		PeakQtyE8:     100,
		VariancePct:   0,
	}
	_ = slicer.RegisterOrder(order)

	// Clip 1: 100
	c1, _ := slicer.NextClip(order.ParentOrderID)
	if c1 != 100 {
		t.Fatalf("Expected clip 1 = 100, got %d", c1)
	}

	// Fill clip 1 -> Clip 2: 100
	c2, _ := slicer.OnClipFilled(order.ParentOrderID, 100)
	if c2 != 100 {
		t.Fatalf("Expected clip 2 = 100, got %d", c2)
	}

	// Fill clip 2 -> Clip 3: exactly remainder 50
	c3, hasMore := slicer.OnClipFilled(order.ParentOrderID, 100)
	if !hasMore || c3 != 50 {
		t.Fatalf("Expected clip 3 = 50, got %d, hasMore=%v", c3, hasMore)
	}

	// Fill clip 3 -> No more clips, completed!
	c4, hasMore := slicer.OnClipFilled(order.ParentOrderID, 50)
	if hasMore || c4 != 0 {
		t.Fatalf("Expected order completion, got hasMore=%v, c4=%d", hasMore, c4)
	}

	snap, _ := slicer.GetOrder(order.ParentOrderID)
	if snap.Status != IcebergStatusCompleted {
		t.Errorf("Expected COMPLETED, got %s", snap.Status)
	}
}

func TestIcebergSlicer_OrderCancellation(t *testing.T) {
	slicer := NewIcebergSlicer()

	order := &IcebergOrder{
		ParentOrderID: "ice-cancel-1",
		UserID:        "user-5",
		Symbol:        "BTC-USDT",
		Side:          "BUY",
		PriceE8:       6000000000000,
		TotalQtyE8:    1000,
		PeakQtyE8:     200,
	}
	_ = slicer.RegisterOrder(order)
	_, _ = slicer.NextClip(order.ParentOrderID)
	_, _ = slicer.OnClipFilled(order.ParentOrderID, 200)

	// Now 200 executed, 800 remaining
	unfilled, err := slicer.CancelOrder(order.ParentOrderID)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
	if unfilled != 800 {
		t.Fatalf("Expected unfilled 800, got %d", unfilled)
	}

	snap, _ := slicer.GetOrder(order.ParentOrderID)
	if snap.Status != IcebergStatusCanceled {
		t.Errorf("Expected status CANCELED, got %s", snap.Status)
	}

	// Subsequent next clip should return false
	_, ok := slicer.NextClip(order.ParentOrderID)
	if ok {
		t.Error("Expected NextClip to return false for cancelled order")
	}
}

func TestIcebergSlicer_ConcurrentClipProcessing(t *testing.T) {
	slicer := NewIcebergSlicerWithSeed(777)

	numOrders := 15
	var wg sync.WaitGroup

	for i := 0; i < numOrders; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			parentID := "ice-conc-" + string(rune('A'+id))
			order := &IcebergOrder{
				ParentOrderID: parentID,
				UserID:        "user-conc",
				Symbol:        "BTC-USDT",
				Side:          "BUY",
				PriceE8:       6000000000000,
				TotalQtyE8:    300,
				PeakQtyE8:     100,
				VariancePct:   10,
			}
			if err := slicer.RegisterOrder(order); err != nil {
				t.Errorf("Register failed: %v", err)
				return
			}

			clip, ok := slicer.NextClip(parentID)
			for ok && clip > 0 {
				nextClip, hasMore := slicer.OnClipFilled(parentID, clip)
				clip = nextClip
				ok = hasMore
			}

			snap, _ := slicer.GetOrder(parentID)
			if snap.Status != IcebergStatusCompleted {
				t.Errorf("Expected COMPLETED for %s, got %s", parentID, snap.Status)
			}
		}(i)
	}

	wg.Wait()
}
