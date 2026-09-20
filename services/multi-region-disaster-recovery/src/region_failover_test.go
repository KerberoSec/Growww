package src

import (
	"testing"
	"time"
)

func TestMultiRegionDR_ZeroLossSyncVerification(t *testing.T) {
	coord := NewMultiRegionDRCoordinator(900.0)

	seq := uint64(500000)
	stateHash := ComputeLedgerHash(seq, 1_000_000_00000000)

	now := time.Now()
	_ = coord.UpdateRegionTelemetry("ap-south-1", seq, stateHash, true, now)
	_ = coord.UpdateRegionTelemetry("ap-south-2", seq, stateHash, true, now)

	// Verify RPO = 0 synchronized
	synced, lag, err := coord.VerifyZeroLossSync("ap-south-1", "ap-south-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !synced || lag != 0 {
		t.Errorf("expected 0 lag sync, got synced=%v lag=%d", synced, lag)
	}

	// Test lag detection when secondary falls behind
	_ = coord.UpdateRegionTelemetry("ap-south-2", seq-10, ComputeLedgerHash(seq-10, 999_000_00000000), true, now)
	synced, lag, err = coord.VerifyZeroLossSync("ap-south-1", "ap-south-2")
	if synced || lag != 10 {
		t.Errorf("expected unsynced with lag 10, got synced=%v lag=%d", synced, lag)
	}
}

func TestMultiRegionDR_ExecuteFailoverSuccess(t *testing.T) {
	coord := NewMultiRegionDRCoordinator(900.0) // 15 mins max RTO

	if coord.GetPrimaryRegion() != "ap-south-1" {
		t.Errorf("expected initial primary ap-south-1")
	}

	start := time.Now()
	record, err := coord.ExecuteFailover("ap-south-2", "AWS Mumbai datacenter connectivity loss", start)
	if err != nil {
		t.Fatalf("failover failed: %v", err)
	}

	if record.SourceRegion != "ap-south-1" || record.TargetRegion != "ap-south-2" {
		t.Errorf("unexpected failover regions: %+v", record)
	}
	if record.RpoSeconds != 0.0 {
		t.Errorf("expected RPO = 0, got %f", record.RpoSeconds)
	}
	if record.RtoElapsedSeconds > 900.0 {
		t.Errorf("RTO exceeded 900s: %f", record.RtoElapsedSeconds)
	}
	if coord.GetPrimaryRegion() != "ap-south-2" {
		t.Errorf("expected new primary ap-south-2, got %s", coord.GetPrimaryRegion())
	}
}

func TestMultiRegionDR_QuorumLossRejection(t *testing.T) {
	coord := NewMultiRegionDRCoordinator(900.0)

	// Mark secondary as lacking quorum
	now := time.Now()
	_ = coord.UpdateRegionTelemetry("ap-south-2", 100, "hash", false, now)

	_, err := coord.ExecuteFailover("ap-south-2", "Test failover", now)
	if err == nil {
		t.Errorf("expected failover to reject target lacking quorum")
	}
}
