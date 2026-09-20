package src

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

// PgBouncerPoolMode defines PgBouncer connection reuse model.
type PgBouncerPoolMode string

const (
	PoolModeSession     PgBouncerPoolMode = "SESSION"
	PoolModeTransaction PgBouncerPoolMode = "TRANSACTION"
	PoolModeStatement   PgBouncerPoolMode = "STATEMENT"
)

var (
	ErrQueueOverflow     = errors.New("pool: client wait queue overflow, rejecting connection")
	ErrAcquisitionTimeout = errors.New("pool: connection acquisition timed out")
)

// PoolSizingConfig holds hardware and workload parameters for connection pool sizing.
type PoolSizingConfig struct {
	CPUCores             int               `json:"cpu_cores"`
	DiskSpindles         int               `json:"disk_spindles"` // or SSD IO parallelism
	Mode                 PgBouncerPoolMode `json:"mode"`
	TargetQPS            float64           `json:"target_qps"`
	AverageQueryLatencyMs float64           `json:"avg_query_latency_ms"`
	MaxClientConnections int               `json:"max_client_connections"`
	MaxWaitQueueSize     int               `json:"max_wait_queue_size"`
	WaitTimeoutMs        int64             `json:"wait_timeout_ms"`
}

// PoolSizingResult computes optimal PostgreSQL and PgBouncer pool limits.
type PoolSizingResult struct {
	OptimalPostgresConnections int     `json:"optimal_postgres_connections"`
	RecommendedPgBouncerPool   int     `json:"recommended_pgbouncer_pool"`
	LittlesLawRequiredConns    float64 `json:"littles_law_required_conns"`
	MaxClientConnections      int     `json:"max_client_connections"`
	MaxWaitQueueDepth          int     `json:"max_wait_queue_depth"`
	UtilizationWarning         bool    `json:"utilization_warning"`
}

// ComputePoolSizing calculates optimal PostgreSQL server pool size and PgBouncer frontend allocation.
// PostgreSQL rule of thumb: PoolSize = (2 * CPUCores) + SpindleCount
// Little's Law: ActiveConnections = QPS * (AverageLatencySeconds)
func ComputePoolSizing(cfg PoolSizingConfig) PoolSizingResult {
	if cfg.CPUCores <= 0 {
		cfg.CPUCores = 4
	}
	if cfg.DiskSpindles <= 0 {
		cfg.DiskSpindles = 1
	}

	// PostgreSQL backend capacity formula
	optimalPostgres := (2 * cfg.CPUCores) + cfg.DiskSpindles

	// Little's law: L = lambda * W
	latencySec := cfg.AverageQueryLatencyMs / 1000.0
	littlesLawConns := cfg.TargetQPS * latencySec

	recommendedPool := optimalPostgres
	if cfg.Mode == PoolModeTransaction {
		// In transaction pooling, connections are held only during active transactions
		recommendedPool = int(math.Ceil(math.Max(float64(optimalPostgres), littlesLawConns*1.2)))
	}

	warning := false
	if littlesLawConns > float64(optimalPostgres*2) {
		warning = true // Workload exceeds recommended backend CPU capacity
	}

	maxClients := cfg.MaxClientConnections
	if maxClients <= 0 {
		maxClients = recommendedPool * 10
	}

	queueDepth := cfg.MaxWaitQueueSize
	if queueDepth <= 0 {
		queueDepth = recommendedPool * 5
	}

	return PoolSizingResult{
		OptimalPostgresConnections: optimalPostgres,
		RecommendedPgBouncerPool:   recommendedPool,
		LittlesLawRequiredConns:    littlesLawConns,
		MaxClientConnections:      maxClients,
		MaxWaitQueueDepth:          queueDepth,
		UtilizationWarning:         warning,
	}
}

// ConnectionPoolGovernor simulates dynamic connection queuing and backpressure.
type ConnectionPoolGovernor struct {
	mu            sync.Mutex
	maxServerConns int
	activeConns   int
	waitQueue     int
	maxQueueSize  int
}

// NewConnectionPoolGovernor creates a new pool governor.
func NewConnectionPoolGovernor(maxServerConns, maxQueueSize int) *ConnectionPoolGovernor {
	return &ConnectionPoolGovernor{
		maxServerConns: maxServerConns,
		maxQueueSize:  maxQueueSize,
	}
}

// AcquireConnection attempts to reserve an active server connection or enqueues.
func (g *ConnectionPoolGovernor) AcquireConnection() (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.activeConns < g.maxServerConns {
		g.activeConns++
		return true, nil // Immediate acquisition
	}

	if g.waitQueue >= g.maxQueueSize {
		return false, fmt.Errorf("%w: current queue depth %d", ErrQueueOverflow, g.waitQueue)
	}

	g.waitQueue++
	return false, nil // Queued
}

// ReleaseConnection releases an active server connection and discharges queue.
func (g *ConnectionPoolGovernor) ReleaseConnection() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.waitQueue > 0 {
		g.waitQueue--
		// connection transferred directly to next waiter
		return
	}

	if g.activeConns > 0 {
		g.activeConns--
	}
}

// Stats returns current pool utilization.
func (g *ConnectionPoolGovernor) Stats() (int, int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.activeConns, g.waitQueue
}
