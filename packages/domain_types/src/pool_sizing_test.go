package src

import (
	"errors"
	"testing"
)

func TestPoolSizingComputation(t *testing.T) {
	// 16 cores, SSD, 5000 QPS, 2ms avg latency
	cfg := PoolSizingConfig{
		CPUCores:              16,
		DiskSpindles:          4,
		Mode:                  PoolModeTransaction,
		TargetQPS:             5000,
		AverageQueryLatencyMs: 2.0,
	}

	res := ComputePoolSizing(cfg)

	// optimal = 2*16 + 4 = 36
	if res.OptimalPostgresConnections != 36 {
		t.Fatalf("expected 36 optimal postgres conns, got %d", res.OptimalPostgresConnections)
	}

	// Little's law: 5000 * (2 / 1000) = 10 conns
	if res.LittlesLawRequiredConns != 10.0 {
		t.Fatalf("expected 10.0 required conns, got %.2f", res.LittlesLawRequiredConns)
	}

	// Recommended pgbouncer pool should be at least optimalPostgres = 36
	if res.RecommendedPgBouncerPool < 36 {
		t.Fatalf("expected at least 36 pgbouncer pool, got %d", res.RecommendedPgBouncerPool)
	}

	if res.UtilizationWarning {
		t.Fatalf("workload of 10 conns on 36 core capacity should not trigger warning")
	}
}

func TestConnectionPoolGovernorQueuing(t *testing.T) {
	// Small pool: 2 server connections, max wait queue 2
	gov := NewConnectionPoolGovernor(2, 2)

	// 1. First 2 acquire immediately
	acquired1, err := gov.AcquireConnection()
	if err != nil || !acquired1 {
		t.Fatalf("conn 1 failed")
	}
	acquired2, err := gov.AcquireConnection()
	if err != nil || !acquired2 {
		t.Fatalf("conn 2 failed")
	}

	active, queue := gov.Stats()
	if active != 2 || queue != 0 {
		t.Fatalf("expected 2 active, 0 queued; got %d, %d", active, queue)
	}

	// 2. 3rd and 4th enter queue
	acquired3, err := gov.AcquireConnection()
	if err != nil || acquired3 {
		t.Fatalf("conn 3 should be queued")
	}
	acquired4, err := gov.AcquireConnection()
	if err != nil || acquired4 {
		t.Fatalf("conn 4 should be queued")
	}

	active, queue = gov.Stats()
	if active != 2 || queue != 2 {
		t.Fatalf("expected 2 active, 2 queued; got %d, %d", active, queue)
	}

	// 3. 5th request overflows queue and gets rejected with ErrQueueOverflow
	_, err = gov.AcquireConnection()
	if err == nil || !errors.Is(err, ErrQueueOverflow) {
		t.Fatalf("expected ErrQueueOverflow, got %v", err)
	}

	// 4. Release connection discharges queue
	gov.ReleaseConnection()
	active, queue = gov.Stats()
	if active != 2 || queue != 1 {
		t.Fatalf("expected 2 active, 1 queued after release; got %d, %d", active, queue)
	}
}
