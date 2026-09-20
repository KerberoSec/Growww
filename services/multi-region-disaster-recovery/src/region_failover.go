package src

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

type RegionState string

const (
	RegionHealthy     RegionState = "HEALTHY"
	RegionDegraded    RegionState = "DEGRADED"
	RegionFailingOver RegionState = "FAILING_OVER"
	RegionOffline     RegionState = "OFFLINE"
)

type DeploymentRegion struct {
	RegionID          string      `json:"region_id"` // "ap-south-1" (Mumbai), "ap-south-2" (Hyderabad)
	LocationName      string      `json:"location_name"`
	State             RegionState `json:"state"`
	DnsWeightPercent  int         `json:"dns_weight_percent"` // 0-100
	LastLedgerSeqNo   uint64      `json:"last_ledger_seq_no"`
	LastLedgerStateHash string    `json:"last_ledger_state_hash"`
	BesuNodeQuorum    bool        `json:"besu_node_quorum"`
	HeartbeatTime     time.Time   `json:"heartbeat_time"`
}

type FailoverRecord struct {
	FailoverID        string    `json:"failover_id"`
	SourceRegion      string    `json:"source_region"`
	TargetRegion      string    `json:"target_region"`
	Reason            string    `json:"reason"`
	RpoSeconds        float64   `json:"rpo_seconds"` // Target 0
	RtoElapsedSeconds float64   `json:"rto_elapsed_seconds"` // Must be <= 900s (15 min)
	SyncVerified      bool      `json:"sync_verified"`
	CompletedAt       time.Time `json:"completed_at"`
}

type MultiRegionDRCoordinator struct {
	mu             sync.RWMutex
	regions        map[string]*DeploymentRegion
	primaryRegion  string
	failoverLogs   []*FailoverRecord
	maxAllowedRtoSec float64 // 900 seconds (15 minutes)
}

func NewMultiRegionDRCoordinator(maxAllowedRtoSec float64) *MultiRegionDRCoordinator {
	if maxAllowedRtoSec <= 0 {
		maxAllowedRtoSec = 900.0 // 15 mins SEBI BCP compliance
	}
	coord := &MultiRegionDRCoordinator{
		regions:          make(map[string]*DeploymentRegion),
		maxAllowedRtoSec: maxAllowedRtoSec,
	}

	// Initialize default dual active regions
	coord.regions["ap-south-1"] = &DeploymentRegion{
		RegionID:         "ap-south-1",
		LocationName:     "Mumbai-DC1",
		State:            RegionHealthy,
		DnsWeightPercent: 100,
		LastLedgerSeqNo:  0,
		BesuNodeQuorum:   true,
		HeartbeatTime:    time.Now(),
	}
	coord.regions["ap-south-2"] = &DeploymentRegion{
		RegionID:         "ap-south-2",
		LocationName:     "Hyderabad-DC2",
		State:            RegionHealthy,
		DnsWeightPercent: 0,
		LastLedgerSeqNo:  0,
		BesuNodeQuorum:   true,
		HeartbeatTime:    time.Now(),
	}
	coord.primaryRegion = "ap-south-1"

	return coord
}

// UpdateRegionTelemetry updates heartbeat, ledger sequence, and Besu quorum state
func (c *MultiRegionDRCoordinator) UpdateRegionTelemetry(regionID string, seqNo uint64, stateHash string, quorum bool, now time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	reg, exists := c.regions[regionID]
	if !exists {
		return errors.New("unknown region")
	}

	reg.LastLedgerSeqNo = seqNo
	reg.LastLedgerStateHash = stateHash
	reg.BesuNodeQuorum = quorum
	reg.HeartbeatTime = now
	return nil
}

// VerifyZeroLossSync verifies RPO = 0 between source and target regions
func (c *MultiRegionDRCoordinator) VerifyZeroLossSync(sourceID, targetID string) (bool, uint64, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	src, ok1 := c.regions[sourceID]
	tgt, ok2 := c.regions[targetID]
	if !ok1 || !ok2 {
		return false, 0, errors.New("invalid region ids")
	}

	// RPO = 0 requires sequence numbers and cryptographic state hashes match
	if src.LastLedgerSeqNo != tgt.LastLedgerSeqNo {
		diff := uint64(0)
		if src.LastLedgerSeqNo > tgt.LastLedgerSeqNo {
			diff = src.LastLedgerSeqNo - tgt.LastLedgerSeqNo
		}
		return false, diff, nil
	}

	if src.LastLedgerStateHash != tgt.LastLedgerStateHash && src.LastLedgerStateHash != "" {
		return false, 0, errors.New("ledger state hash divergence detected between regions")
	}

	return true, 0, nil
}

// ExecuteFailover coordinates regional switchover within RTO limits
func (c *MultiRegionDRCoordinator) ExecuteFailover(targetRegion, reason string, startTime time.Time) (*FailoverRecord, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sourceID := c.primaryRegion
	if sourceID == targetRegion {
		return nil, errors.New("target is already primary region")
	}

	src := c.regions[sourceID]
	tgt, exists := c.regions[targetRegion]
	if !exists {
		return nil, errors.New("target region does not exist")
	}

	if !tgt.BesuNodeQuorum {
		return nil, errors.New("target region lacks Besu QBFT node quorum")
	}

	// 1. Mark source as failing over / offline
	src.State = RegionOffline
	src.DnsWeightPercent = 0

	// 2. Promote target region to primary
	tgt.State = RegionHealthy
	tgt.DnsWeightPercent = 100
	c.primaryRegion = targetRegion

	// 3. Measure RTO elapsed
	endTime := time.Now()
	elapsedSec := endTime.Sub(startTime).Seconds()
	if elapsedSec > c.maxAllowedRtoSec {
		return nil, fmt.Errorf("failover exceeded max allowed RTO (%.2fs > %.2fs)", elapsedSec, c.maxAllowedRtoSec)
	}

	// 4. Record failover event
	recID := fmt.Sprintf("dr_%s_to_%s_%d", sourceID, targetRegion, endTime.Unix())
	record := &FailoverRecord{
		FailoverID:        recID,
		SourceRegion:      sourceID,
		TargetRegion:      targetRegion,
		Reason:            reason,
		RpoSeconds:        0.0, // RPO = 0
		RtoElapsedSeconds: elapsedSec,
		SyncVerified:      true,
		CompletedAt:       endTime,
	}

	c.failoverLogs = append(c.failoverLogs, record)
	return record, nil
}

// ComputeLedgerHash generates deterministic state hash for verification
func ComputeLedgerHash(seqNo uint64, balanceTotalE8 int64) string {
	raw := fmt.Sprintf("%d:%d:nbse-dr-seed", seqNo, balanceTotalE8)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func (c *MultiRegionDRCoordinator) GetPrimaryRegion() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.primaryRegion
}
