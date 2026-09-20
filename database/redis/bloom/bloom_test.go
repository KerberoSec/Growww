package bloom_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"growww/database/redis/bloom"
)

func TestBloomFilter_MathModels(t *testing.T) {
	n := uint64(100000)
	p := 0.01 // 1% false positive

	m, k := bloom.CalculateOptimalParams(n, p)

	if m == 0 || k == 0 {
		t.Fatalf("expected non-zero m and k, got m=%d, k=%d", m, k)
	}

	actualFP := bloom.CalculateFalsePositiveRate(m, k, n)
	if actualFP > 0.02 {
		t.Errorf("expected FP rate around 0.01, got %f", actualFP)
	}
}

func TestStandardBloomFilter_ZeroFalseNegatives(t *testing.T) {
	n := uint64(5000)
	filter := bloom.NewStandardBloomFilter(n, 0.001)

	// Insert 5000 keys
	keys := make([][]byte, n)
	for i := uint64(0); i < n; i++ {
		keys[i] = []byte(fmt.Sprintf("order-uuid-%d", i))
		filter.Add(keys[i])
	}

	// Verify ZERO false negatives: every inserted element MUST be found
	for i := uint64(0); i < n; i++ {
		if !filter.Contains(keys[i]) {
			t.Fatalf("FALSE NEGATIVE: element %s was not found!", string(keys[i]))
		}
	}

	if filter.ElementCount() != n {
		t.Errorf("expected element count %d, got %d", n, filter.ElementCount())
	}

	// Check false positive rate on uninserted elements
	uninsertedCount := 5000
	falsePositives := 0
	for i := 0; i < uninsertedCount; i++ {
		unseen := []byte(fmt.Sprintf("unseen-key-%d", i))
		if filter.Contains(unseen) {
			falsePositives++
		}
	}

	fpRate := float64(falsePositives) / float64(uninsertedCount)
	if fpRate > 0.01 { // Allow slight variance above 0.001 target
		t.Errorf("observed false positive rate too high: %f (%d/%d)", fpRate, falsePositives, uninsertedCount)
	}
}

func TestScalableBloomFilter_DynamicScaling(t *testing.T) {
	// Initial capacity of 100 per layer
	sbf := bloom.NewScalableBloomFilter(100, 0.01)

	// Insert 1000 elements (should trigger multiple layers)
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("scalable-item-%d", i))
		sbf.Add(key)
	}

	if sbf.LayerCount() <= 1 {
		t.Errorf("expected scalable filter to expand layers, got %d", sbf.LayerCount())
	}
	if sbf.TotalElements() != 1000 {
		t.Errorf("expected 1000 total elements, got %d", sbf.TotalElements())
	}

	// Verify zero false negatives
	for i := 0; i < 1000; i++ {
		key := []byte(fmt.Sprintf("scalable-item-%d", i))
		if !sbf.Contains(key) {
			t.Fatalf("scalable filter missed element %d", i)
		}
	}
}

func TestCountingBloomFilter_AddAndRemove(t *testing.T) {
	cbf := bloom.NewCountingBloomFilter(1000, 0.01)

	key1 := []byte("active-session-1")
	key2 := []byte("active-session-2")

	cbf.Add(key1)
	cbf.Add(key2)

	if !cbf.Contains(key1) || !cbf.Contains(key2) {
		t.Fatalf("expected both keys present")
	}
	if cbf.ElementCount() != 2 {
		t.Errorf("expected count 2, got %d", cbf.ElementCount())
	}

	// Remove key1
	err := cbf.Remove(key1)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if cbf.Contains(key1) {
		t.Errorf("expected key1 to be removed from counting filter")
	}
	if !cbf.Contains(key2) {
		t.Errorf("key2 should still be present")
	}
	if cbf.ElementCount() != 1 {
		t.Errorf("expected count 1, got %d", cbf.ElementCount())
	}
}

func TestRotationalSlidingWindowBloomFilter_Rotation(t *testing.T) {
	windowSize := 200 * time.Millisecond
	bucketCount := 2
	rbf := bloom.NewRotationalSlidingWindowBloomFilter(windowSize, bucketCount, 1000, 0.01)

	now := time.Now()
	key := []byte("temporary-dedup-key")

	rbf.Add(key, now)
	if !rbf.Contains(key, now) {
		t.Fatalf("key should be present in current window")
	}

	// Fast forward time beyond windowSize
	futureTime := now.Add(300 * time.Millisecond)
	rbf.RotateIfNeeded(futureTime)

	// After full window expiration, key should be evicted
	if rbf.Contains(key, futureTime) {
		t.Errorf("key should have expired from sliding window")
	}
}

func TestDuplicateDetectorEngine_ReplayAttackMitigation(t *testing.T) {
	engine := bloom.NewDuplicateDetectorEngine()

	entityID := "user-portfolio-123"
	txHash1 := "0x3a4f8d9b1c2e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a"

	// First submission (nonce 1) -> Success
	err := engine.CheckAndRecordTransaction(txHash1, entityID, 1)
	if err != nil {
		t.Fatalf("unexpected error on first submission: %v", err)
	}

	// Replay same txHash -> Blocked
	err = engine.CheckAndRecordTransaction(txHash1, entityID, 2)
	if !errors.Is(err, bloom.ErrDuplicateTransaction) {
		t.Errorf("expected ErrDuplicateTransaction for replayed txHash, got %v", err)
	}

	// Non-monotonic nonce submission -> Blocked
	txHash2 := "0x1111111111111111111111111111111111111111111111111111111111111111"
	err = engine.CheckAndRecordTransaction(txHash2, entityID, 1) // reused nonce 1
	if !errors.Is(err, bloom.ErrNonMonotonicNonce) {
		t.Errorf("expected ErrNonMonotonicNonce for reused nonce, got %v", err)
	}

	// Valid new tx with higher monotonic nonce -> Success
	txHash3 := "0x2222222222222222222222222222222222222222222222222222222222222222"
	err = engine.CheckAndRecordTransaction(txHash3, entityID, 2)
	if err != nil {
		t.Fatalf("unexpected error for valid higher nonce tx: %v", err)
	}

	metrics := engine.GetMetrics()
	if metrics.TotalChecked != 4 {
		t.Errorf("expected 4 total checked, got %d", metrics.TotalChecked)
	}
	if metrics.DuplicatesBlocked != 1 {
		t.Errorf("expected 1 duplicate blocked, got %d", metrics.DuplicatesBlocked)
	}
	if metrics.ReplaysBlocked != 1 {
		t.Errorf("expected 1 replay nonce blocked, got %d", metrics.ReplaysBlocked)
	}

	auditLog := engine.GetAuditLog()
	if len(auditLog) != 2 {
		t.Errorf("expected 2 audit log entries for security violations, got %d", len(auditLog))
	}
}

func TestDuplicateDetectorEngine_ConcurrentStress(t *testing.T) {
	engine := bloom.NewDuplicateDetectorEngine()
	var wg sync.WaitGroup

	numGoroutines := 10
	itemsPerRoutine := 100

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for i := 0; i < itemsPerRoutine; i++ {
				txHash := fmt.Sprintf("tx_hash_%d_%d", routineID, i)
				entity := fmt.Sprintf("entity_%d", routineID)
				nonce := uint64(i + 1)
				_ = engine.CheckAndRecordTransaction(txHash, entity, nonce)
			}
		}(g)
	}

	wg.Wait()

	metrics := engine.GetMetrics()
	expectedTotal := uint64(numGoroutines * itemsPerRoutine)
	if metrics.TotalChecked != expectedTotal {
		t.Errorf("expected %d checked, got %d", expectedTotal, metrics.TotalChecked)
	}
}
