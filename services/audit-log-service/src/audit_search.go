package main

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"
)

// AuditSearchFilter defines query parameters for searching the immutable audit log.
type AuditSearchFilter struct {
	Category   string    // Filter by event category (AUTH, ORDER, TRADE, SETTLEMENT, REGULATORY)
	ActorID    string    // Filter by pseudonymous actor ID
	Action     string    // Substring match on action field
	StartTime  time.Time // Inclusive lower bound on timestamp
	EndTime    time.Time // Inclusive upper bound on timestamp
	MinIndex   uint64    // Starting index (1-based)
	MaxResults int       // Max records to return (0 = unlimited)
}

// SearchAuditLog searches the append-only ledger with the given filter criteria.
// Returns matching records in insertion order.
func (s *ImmutableWORMStorage) SearchAuditLog(filter AuditSearchFilter) []AuditLogRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	var results []AuditLogRecord
	for _, rec := range s.records {
		if filter.MinIndex > 0 && rec.Index < filter.MinIndex {
			continue
		}
		if filter.Category != "" && rec.EventCategory != filter.Category {
			continue
		}
		if filter.ActorID != "" && rec.ActorID != filter.ActorID {
			continue
		}
		if filter.Action != "" && !strings.Contains(rec.Action, filter.Action) {
			continue
		}
		if !filter.StartTime.IsZero() && rec.Timestamp.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && rec.Timestamp.After(filter.EndTime) {
			continue
		}
		results = append(results, rec)
		if filter.MaxResults > 0 && len(results) >= filter.MaxResults {
			break
		}
	}
	return results
}

// VerifyChainIntegrity walks the entire hash chain and returns the index of the
// first tampered record, or 0 if the chain is fully intact.
func (s *ImmutableWORMStorage) VerifyChainIntegrity() (valid bool, brokenAtIndex uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expectedPrevHash := "0000000000000000000000000000000000000000000000000000000000000000"
	for _, rec := range s.records {
		if rec.PreviousHash != expectedPrevHash {
			return false, rec.Index
		}
		expectedPrevHash = rec.RecordHash
	}
	return true, 0
}

// GetRecordCount returns the total number of immutable audit log entries.
func (s *ImmutableWORMStorage) GetRecordCount() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint64(len(s.records))
}

// ExportComplianceCSV generates a CSV-formatted compliance export of audit records
// matching the given filter. Suitable for SEBI, RBI, and internal compliance reporting.
func (s *ImmutableWORMStorage) ExportComplianceCSV(filter AuditSearchFilter) string {
	records := s.SearchAuditLog(filter)

	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header row
	_ = writer.Write([]string{
		"Index", "PreviousHash", "RecordHash", "EventCategory",
		"ActorID", "Action", "PayloadDigest", "Timestamp",
	})

	for _, rec := range records {
		_ = writer.Write([]string{
			fmt.Sprintf("%d", rec.Index),
			rec.PreviousHash,
			rec.RecordHash,
			rec.EventCategory,
			rec.ActorID,
			rec.Action,
			rec.PayloadDigest,
			rec.Timestamp.Format(time.RFC3339Nano),
		})
	}

	writer.Flush()
	return buf.String()
}
