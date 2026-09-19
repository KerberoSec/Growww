// Package reconciliation provides a production-grade 3-way reconciliation
// engine for the NBSE sovereign exchange. It compares SQL ledger entries
// against Hyperledger Besu on-chain state and NSDL/CDSL depository records,
// detecting mismatches and generating alerts.
package reconciliation

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// RecordSource identifies where a record originates.
type RecordSource string

const (
	SourceLedger     RecordSource = "sql_ledger"
	SourceBesu       RecordSource = "besu_onchain"
	SourceDepository RecordSource = "nsdl_cdsl"
)

// MismatchSeverity classifies the severity of reconciliation mismatches.
type MismatchSeverity string

const (
	SeverityInfo     MismatchSeverity = "info"
	SeverityWarning  MismatchSeverity = "warning"
	SeverityCritical MismatchSeverity = "critical"
)

// MismatchType classifies the type of reconciliation mismatch.
type MismatchType string

const (
	MismatchMissing    MismatchType = "missing"       // Record exists in one source but not another
	MismatchQuantity   MismatchType = "quantity"       // Quantity differs between sources
	MismatchPrice      MismatchType = "price"          // Price/amount differs
	MismatchStatus     MismatchType = "status"         // Settlement status differs
	MismatchTimestamp  MismatchType = "timestamp"      // Timestamp drift exceeds tolerance
)

// BalanceRecord represents a holding balance from any source.
type BalanceRecord struct {
	AccountID  string
	AssetID    string
	Quantity   float64
	Amount     float64 // monetary value
	Status     string
	Timestamp  time.Time
	Source     RecordSource
	TxHash     string // blockchain transaction hash (Besu)
	DepositoryRef string // NSDL/CDSL reference
}

type recordKey struct {
	AccountID string
	AssetID   string
}

// Mismatch represents a detected discrepancy between sources.
type Mismatch struct {
	ID            string
	JobID         string
	AccountID     string
	AssetID       string
	Type          MismatchType
	Severity      MismatchSeverity
	Sources       []RecordSource
	ExpectedValue string
	ActualValue   string
	Delta         float64
	DetectedAt    time.Time
	Resolved      bool
	ResolvedAt    *time.Time
	Notes         string
}

// ReconciliationJob represents a reconciliation run.
type ReconciliationJob struct {
	ID          string
	StartedAt   time.Time
	CompletedAt *time.Time
	Status      string
	TotalRecords int
	Mismatches  []*Mismatch
	Summary     *ReconciliationSummary
}

// ReconciliationSummary provides aggregate stats for a reconciliation run.
type ReconciliationSummary struct {
	TotalAccounts    int
	TotalAssets      int
	TotalRecords     int
	MismatchCount    int
	CriticalCount    int
	WarningCount     int
	InfoCount        int
	MatchRate        float64 // percentage
	ProcessingTimeMs int64
}

// RecordProvider is an interface for fetching records from a specific source.
type RecordProvider interface {
	// FetchBalances returns all balance records for the given accounts and assets.
	FetchBalances(ctx context.Context, accountIDs, assetIDs []string) ([]BalanceRecord, error)
	Source() RecordSource
}

// AlertHandler is called when mismatches are detected.
type AlertHandler interface {
	OnMismatch(ctx context.Context, mismatch *Mismatch) error
}

// Config holds reconciliation engine configuration.
type Config struct {
	QuantityTolerance float64       // allowed absolute difference in quantities
	PriceTolerance    float64       // allowed absolute difference in prices
	TimestampDrift    time.Duration // max allowed timestamp difference
	BatchSize         int           // number of accounts per batch
}

// DefaultConfig returns sensible production defaults.
func DefaultConfig() Config {
	return Config{
		QuantityTolerance: 0.0001,
		PriceTolerance:    0.01,
		TimestampDrift:    5 * time.Minute,
		BatchSize:         1000,
	}
}

// Engine is the main 3-way reconciliation engine.
type Engine struct {
	config    Config
	providers map[RecordSource]RecordProvider
	alerts    []AlertHandler
	mu        sync.RWMutex
	jobs      map[string]*ReconciliationJob
}

// NewEngine creates a new reconciliation engine.
func NewEngine(cfg Config) *Engine {
	return &Engine{
		config:    cfg,
		providers: make(map[RecordSource]RecordProvider),
		jobs:      make(map[string]*ReconciliationJob),
	}
}

// RegisterProvider registers a record source provider.
func (e *Engine) RegisterProvider(p RecordProvider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.providers[p.Source()] = p
}

// RegisterAlertHandler registers an alert handler.
func (e *Engine) RegisterAlertHandler(h AlertHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.alerts = append(e.alerts, h)
}

// Reconcile performs a 3-way reconciliation across all registered sources.
func (e *Engine) Reconcile(ctx context.Context, jobID string, accountIDs, assetIDs []string) (*ReconciliationJob, error) {
	if len(accountIDs) == 0 || len(assetIDs) == 0 {
		return nil, errors.New("accountIDs and assetIDs must not be empty")
	}

	e.mu.RLock()
	if len(e.providers) < 2 {
		e.mu.RUnlock()
		return nil, errors.New("at least 2 providers are required for reconciliation")
	}
	e.mu.RUnlock()

	startTime := time.Now()
	job := &ReconciliationJob{
		ID:        jobID,
		StartedAt: startTime,
		Status:    "running",
	}

	e.mu.Lock()
	e.jobs[jobID] = job
	e.mu.Unlock()

	// Fetch records from all sources
	allRecords := make(map[RecordSource][]BalanceRecord)
	e.mu.RLock()
	providers := make(map[RecordSource]RecordProvider)
	for k, v := range e.providers {
		providers[k] = v
	}
	e.mu.RUnlock()

	for source, provider := range providers {
		records, err := provider.FetchBalances(ctx, accountIDs, assetIDs)
		if err != nil {
			job.Status = "failed"
			return job, fmt.Errorf("failed to fetch from %s: %w", source, err)
		}
		allRecords[source] = records
	}

	// Build lookup maps: (account, asset) -> record per source
	sourceMaps := make(map[RecordSource]map[recordKey]*BalanceRecord)
	totalRecords := 0
	for source, records := range allRecords {
		m := make(map[recordKey]*BalanceRecord)
		for i := range records {
			k := recordKey{records[i].AccountID, records[i].AssetID}
			r := records[i]
			m[k] = &r
			totalRecords++
		}
		sourceMaps[source] = m
	}
	job.TotalRecords = totalRecords

	// Collect all unique keys
	allKeys := make(map[recordKey]bool)
	for _, m := range sourceMaps {
		for k := range m {
			allKeys[k] = true
		}
	}

	// Compare records across sources
	var mismatches []*Mismatch
	sources := make([]RecordSource, 0, len(sourceMaps))
	for s := range sourceMaps {
		sources = append(sources, s)
	}

	for key := range allKeys {
		keyMismatches := e.compareRecords(jobID, key.AccountID, key.AssetID, sources, sourceMaps)
		mismatches = append(mismatches, keyMismatches...)
	}

	// Fire alerts
	e.mu.RLock()
	alertHandlers := make([]AlertHandler, len(e.alerts))
	copy(alertHandlers, e.alerts)
	e.mu.RUnlock()

	for _, mm := range mismatches {
		for _, handler := range alertHandlers {
			_ = handler.OnMismatch(ctx, mm)
		}
	}

	now := time.Now()
	job.CompletedAt = &now
	job.Status = "completed"
	job.Mismatches = mismatches

	// Build summary
	summary := &ReconciliationSummary{
		TotalAccounts: len(accountIDs),
		TotalAssets:   len(assetIDs),
		TotalRecords:  totalRecords,
		MismatchCount: len(mismatches),
		ProcessingTimeMs: now.Sub(startTime).Milliseconds(),
	}
	for _, mm := range mismatches {
		switch mm.Severity {
		case SeverityCritical:
			summary.CriticalCount++
		case SeverityWarning:
			summary.WarningCount++
		case SeverityInfo:
			summary.InfoCount++
		}
	}
	if len(allKeys) > 0 {
		summary.MatchRate = float64(len(allKeys)-len(mismatches)) / float64(len(allKeys)) * 100
		if summary.MatchRate < 0 {
			summary.MatchRate = 0
		}
	}
	job.Summary = summary

	return job, nil
}

// compareRecords checks for mismatches across sources for a given (account, asset).
func (e *Engine) compareRecords(jobID, accountID, assetID string, sources []RecordSource, sourceMaps map[RecordSource]map[recordKey]*BalanceRecord) []*Mismatch {
	key := recordKey{accountID, assetID}
	var mismatches []*Mismatch

	// Check for missing records
	var present []RecordSource
	for _, src := range sources {
		if _, ok := sourceMaps[src][key]; ok {
			present = append(present, src)
		}
	}

	if len(present) < len(sources) {
		var missing []RecordSource
		for _, src := range sources {
			found := false
			for _, p := range present {
				if p == src {
					found = true
					break
				}
			}
			if !found {
				missing = append(missing, src)
			}
		}
		mismatches = append(mismatches, &Mismatch{
			ID:         fmt.Sprintf("%s-missing-%s-%s", jobID, accountID, assetID),
			JobID:      jobID,
			AccountID:  accountID,
			AssetID:    assetID,
			Type:       MismatchMissing,
			Severity:   SeverityCritical,
			Sources:    missing,
			DetectedAt: time.Now(),
			Notes:      fmt.Sprintf("record missing from %v", missing),
		})
	}

	// Compare quantity and price between all pairs of present sources
	for i := 0; i < len(present); i++ {
		for j := i + 1; j < len(present); j++ {
			r1 := sourceMaps[present[i]][key]
			r2 := sourceMaps[present[j]][key]

			// Quantity mismatch
			if math.Abs(r1.Quantity-r2.Quantity) > e.config.QuantityTolerance {
				severity := SeverityWarning
				if math.Abs(r1.Quantity-r2.Quantity) > e.config.QuantityTolerance*100 {
					severity = SeverityCritical
				}
				mismatches = append(mismatches, &Mismatch{
					ID:            fmt.Sprintf("%s-qty-%s-%s-%s-%s", jobID, accountID, assetID, present[i], present[j]),
					JobID:         jobID,
					AccountID:     accountID,
					AssetID:       assetID,
					Type:          MismatchQuantity,
					Severity:      severity,
					Sources:       []RecordSource{present[i], present[j]},
					ExpectedValue: fmt.Sprintf("%.4f (%s)", r1.Quantity, present[i]),
					ActualValue:   fmt.Sprintf("%.4f (%s)", r2.Quantity, present[j]),
					Delta:         math.Abs(r1.Quantity - r2.Quantity),
					DetectedAt:    time.Now(),
				})
			}

			// Price/amount mismatch
			if math.Abs(r1.Amount-r2.Amount) > e.config.PriceTolerance {
				mismatches = append(mismatches, &Mismatch{
					ID:            fmt.Sprintf("%s-amt-%s-%s-%s-%s", jobID, accountID, assetID, present[i], present[j]),
					JobID:         jobID,
					AccountID:     accountID,
					AssetID:       assetID,
					Type:          MismatchPrice,
					Severity:      SeverityWarning,
					Sources:       []RecordSource{present[i], present[j]},
					ExpectedValue: fmt.Sprintf("%.2f (%s)", r1.Amount, present[i]),
					ActualValue:   fmt.Sprintf("%.2f (%s)", r2.Amount, present[j]),
					Delta:         math.Abs(r1.Amount - r2.Amount),
					DetectedAt:    time.Now(),
				})
			}

			// Status mismatch
			if r1.Status != "" && r2.Status != "" && r1.Status != r2.Status {
				mismatches = append(mismatches, &Mismatch{
					ID:            fmt.Sprintf("%s-status-%s-%s-%s-%s", jobID, accountID, assetID, present[i], present[j]),
					JobID:         jobID,
					AccountID:     accountID,
					AssetID:       assetID,
					Type:          MismatchStatus,
					Severity:      SeverityCritical,
					Sources:       []RecordSource{present[i], present[j]},
					ExpectedValue: fmt.Sprintf("%s (%s)", r1.Status, present[i]),
					ActualValue:   fmt.Sprintf("%s (%s)", r2.Status, present[j]),
					DetectedAt:    time.Now(),
				})
			}
		}
	}

	return mismatches
}

// GetJob retrieves a reconciliation job by ID.
func (e *Engine) GetJob(jobID string) (*ReconciliationJob, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	job, ok := e.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job %q not found", jobID)
	}
	return job, nil
}

// --- Mock implementations for testing ---

// MockProvider implements RecordProvider for testing.
type MockProvider struct {
	source  RecordSource
	records []BalanceRecord
	err     error
}

// NewMockProvider creates a mock provider.
func NewMockProvider(source RecordSource, records []BalanceRecord) *MockProvider {
	return &MockProvider{source: source, records: records}
}

// FetchBalances implements RecordProvider.
func (mp *MockProvider) FetchBalances(_ context.Context, _, _ []string) ([]BalanceRecord, error) {
	if mp.err != nil {
		return nil, mp.err
	}
	return mp.records, nil
}

// Source implements RecordProvider.
func (mp *MockProvider) Source() RecordSource { return mp.source }

// MockAlertHandler records alerts for testing.
type MockAlertHandler struct {
	mu     sync.Mutex
	alerts []*Mismatch
}

// OnMismatch implements AlertHandler.
func (h *MockAlertHandler) OnMismatch(_ context.Context, m *Mismatch) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.alerts = append(h.alerts, m)
	return nil
}

// AlertCount returns the number of alerts received.
func (h *MockAlertHandler) AlertCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.alerts)
}
