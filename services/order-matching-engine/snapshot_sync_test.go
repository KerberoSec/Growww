package ordermatching

import (
	"sync"
	"testing"
)

func TestGoOrderBookSync_HappyPath(t *testing.T) {
	syncMgr := NewOrderBookSync("BTC-USDT")

	// 1. Buffer deltas 1..2 before snapshot
	err := syncMgr.BufferOrApplyDelta(BookDelta{
		Symbol:        "BTC-USDT",
		FirstUpdateID: 1,
		LastUpdateID:  2,
		Bids:          []PriceDelta{{PriceE8: 50000_00000000, QtyE8: 1_00000000}},
	})
	if err != nil {
		t.Fatalf("unexpected error buffering delta: %v", err)
	}

	// 2. Apply snapshot at lastUpdateID = 2
	snapshot := L2Snapshot{
		Symbol:       "BTC-USDT",
		LastUpdateID: 2,
		Bids: []PriceLevel{
			{PriceE8: 50000_00000000, TotalQuantityE8: 1_00000000, OrderCount: 1},
		},
		Asks: []PriceLevel{
			{PriceE8: 51000_00000000, TotalQuantityE8: 2_00000000, OrderCount: 1},
		},
	}
	if err := syncMgr.ApplySnapshot(snapshot); err != nil {
		t.Fatalf("failed to apply snapshot: %v", err)
	}

	if syncMgr.Status() != StatusSynchronized {
		t.Fatalf("expected status SYNCHRONIZED, got %s", syncMgr.Status())
	}
	if syncMgr.LastUpdateID() != 2 {
		t.Fatalf("expected LastUpdateID 2, got %d", syncMgr.LastUpdateID())
	}

	// 3. Apply next consecutive delta (updateID 3)
	err = syncMgr.BufferOrApplyDelta(BookDelta{
		Symbol:        "BTC-USDT",
		FirstUpdateID: 3,
		LastUpdateID:  3,
		Bids:          []PriceDelta{{PriceE8: 50500_00000000, QtyE8: 3_00000000}},
	})
	if err != nil {
		t.Fatalf("failed to apply consecutive delta: %v", err)
	}

	bids, asks, err := syncMgr.GetL2Depth(5)
	if err != nil {
		t.Fatalf("GetL2Depth error: %v", err)
	}
	if len(bids) != 2 {
		t.Fatalf("expected 2 bids, got %d", len(bids))
	}
	if bids[0].PriceE8 != 50500_00000000 || bids[0].TotalQuantityE8 != 3_00000000 {
		t.Fatalf("unexpected top bid: %+v", bids[0])
	}
	if len(asks) != 1 {
		t.Fatalf("expected 1 ask, got %d", len(asks))
	}
}

func TestGoOrderBookSync_ZeroGhostLiquidity(t *testing.T) {
	syncMgr := NewOrderBookSync("ETH-USDT")

	snapshot := L2Snapshot{
		Symbol:       "ETH-USDT",
		LastUpdateID: 10,
		Bids: []PriceLevel{
			{PriceE8: 3000_00000000, TotalQuantityE8: 10_00000000, OrderCount: 1},
		},
	}
	if err := syncMgr.ApplySnapshot(snapshot); err != nil {
		t.Fatalf("failed to apply snapshot: %v", err)
	}

	// Emit delta clearing the price level (Qty = 0)
	err := syncMgr.BufferOrApplyDelta(BookDelta{
		Symbol:        "ETH-USDT",
		FirstUpdateID: 11,
		LastUpdateID:  11,
		Bids:          []PriceDelta{{PriceE8: 3000_00000000, QtyE8: 0}},
	})
	if err != nil {
		t.Fatalf("failed to apply delta: %v", err)
	}

	bids, _, err := syncMgr.GetL2Depth(5)
	if err != nil {
		t.Fatalf("GetL2Depth error: %v", err)
	}
	if len(bids) != 0 {
		t.Fatalf("expected zero ghost liquidity (0 bids), got %d", len(bids))
	}
}

func TestGoOrderBookSync_GapDetectionAndResync(t *testing.T) {
	syncMgr := NewOrderBookSync("SOL-USDT")

	snapshot := L2Snapshot{
		Symbol:       "SOL-USDT",
		LastUpdateID: 100,
	}
	if err := syncMgr.ApplySnapshot(snapshot); err != nil {
		t.Fatalf("failed to apply snapshot: %v", err)
	}

	// Send delta with gap: expected 101, but received 105
	err := syncMgr.BufferOrApplyDelta(BookDelta{
		Symbol:        "SOL-USDT",
		FirstUpdateID: 105,
		LastUpdateID:  105,
	})
	if err == nil {
		t.Fatalf("expected gap error, got nil")
	}

	if syncMgr.Status() != StatusDesynchronized {
		t.Fatalf("expected DESYNCHRONIZED, got %s", syncMgr.Status())
	}
	if syncMgr.ResyncCount() != 1 {
		t.Fatalf("expected 1 resync request, got %d", syncMgr.ResyncCount())
	}
}

func TestGoOrderBookSync_ConcurrentAccess(t *testing.T) {
	syncMgr := NewOrderBookSync("BTC-USDT")
	_ = syncMgr.ApplySnapshot(L2Snapshot{
		Symbol:       "BTC-USDT",
		LastUpdateID: 0,
	})

	var wg sync.WaitGroup
	numRoutines := 50
	deltasPerRoutine := 100

	// Sequentially apply deltas to maintain strict update IDs
	for i := 1; i <= numRoutines*deltasPerRoutine; i++ {
		_ = syncMgr.BufferOrApplyDelta(BookDelta{
			Symbol:        "BTC-USDT",
			FirstUpdateID: uint64(i),
			LastUpdateID:  uint64(i),
			Bids:          []PriceDelta{{PriceE8: uint64(50000 + i%20), QtyE8: uint64(i)}},
		})
	}

	// Concurrently read depth
	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, _, _ = syncMgr.GetL2Depth(10)
			}
		}()
	}
	wg.Wait()

	if syncMgr.Status() != StatusSynchronized {
		t.Fatalf("expected SYNCHRONIZED status, got %s", syncMgr.Status())
	}
}
