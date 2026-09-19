package main

import (
	"strings"
	"testing"
)

func TestAppendAuditLog_HashChainLinking(t *testing.T) {
	store := NewImmutableWORMStorage()

	rec1 := store.AppendAuditLog("AUTH", "user-001", "LOGIN", `{"ip":"10.0.0.1"}`)
	rec2 := store.AppendAuditLog("ORDER", "user-001", "PLACE_ORDER", `{"order_id":"ORD-123"}`)

	if rec1.Index != 1 {
		t.Fatalf("Expected index 1, got %d", rec1.Index)
	}
	if rec2.Index != 2 {
		t.Fatalf("Expected index 2, got %d", rec2.Index)
	}

	// rec2.PreviousHash must equal rec1.RecordHash for chain integrity
	if rec2.PreviousHash != rec1.RecordHash {
		t.Fatalf("Hash chain broken: rec2.PreviousHash=%s != rec1.RecordHash=%s",
			rec2.PreviousHash, rec1.RecordHash)
	}
}

func TestAppendAuditLog_GenesisBlockPreviousHash(t *testing.T) {
	store := NewImmutableWORMStorage()
	rec := store.AppendAuditLog("TRADE", "user-002", "EXECUTION", `{"trade_id":"T-1"}`)

	expectedGenesisPrev := "0000000000000000000000000000000000000000000000000000000000000000"
	if rec.PreviousHash != expectedGenesisPrev {
		t.Fatalf("Genesis record PreviousHash should be all zeros, got %s", rec.PreviousHash)
	}
}

func TestSearchAuditLog_FilterByCategory(t *testing.T) {
	store := NewImmutableWORMStorage()
	store.AppendAuditLog("AUTH", "u1", "LOGIN", `{}`)
	store.AppendAuditLog("ORDER", "u1", "PLACE", `{}`)
	store.AppendAuditLog("AUTH", "u2", "LOGOUT", `{}`)
	store.AppendAuditLog("TRADE", "u1", "FILL", `{}`)

	results := store.SearchAuditLog(AuditSearchFilter{Category: "AUTH"})
	if len(results) != 2 {
		t.Fatalf("Expected 2 AUTH records, got %d", len(results))
	}
	for _, r := range results {
		if r.EventCategory != "AUTH" {
			t.Errorf("Unexpected category: %s", r.EventCategory)
		}
	}
}

func TestSearchAuditLog_FilterByActorAndMaxResults(t *testing.T) {
	store := NewImmutableWORMStorage()
	for i := 0; i < 10; i++ {
		store.AppendAuditLog("ORDER", "trader-A", "PLACE", `{}`)
	}
	store.AppendAuditLog("ORDER", "trader-B", "PLACE", `{}`)

	results := store.SearchAuditLog(AuditSearchFilter{ActorID: "trader-A", MaxResults: 5})
	if len(results) != 5 {
		t.Fatalf("Expected 5 results with MaxResults=5, got %d", len(results))
	}
}

func TestVerifyChainIntegrity_IntactChain(t *testing.T) {
	store := NewImmutableWORMStorage()
	store.AppendAuditLog("AUTH", "u1", "LOGIN", `{}`)
	store.AppendAuditLog("ORDER", "u1", "PLACE", `{}`)
	store.AppendAuditLog("TRADE", "u1", "FILL", `{}`)

	valid, brokenAt := store.VerifyChainIntegrity()
	if !valid {
		t.Fatalf("Chain should be valid, broken at index %d", brokenAt)
	}
}

func TestVerifyChainIntegrity_TamperedRecord(t *testing.T) {
	store := NewImmutableWORMStorage()
	store.AppendAuditLog("AUTH", "u1", "LOGIN", `{}`)
	store.AppendAuditLog("ORDER", "u1", "PLACE", `{}`)
	store.AppendAuditLog("TRADE", "u1", "FILL", `{}`)

	// Tamper with the second record's PreviousHash
	store.records[1].PreviousHash = "deadbeef"

	valid, brokenAt := store.VerifyChainIntegrity()
	if valid {
		t.Fatal("Chain should be invalid after tampering")
	}
	if brokenAt != 2 {
		t.Fatalf("Expected break at index 2, got %d", brokenAt)
	}
}

func TestExportComplianceCSV(t *testing.T) {
	store := NewImmutableWORMStorage()
	store.AppendAuditLog("SETTLEMENT", "u1", "T+1_SETTLE", `{"batch":"S-001"}`)
	store.AppendAuditLog("REGULATORY", "system", "SEBI_REPORT", `{"report":"daily"}`)

	csv := store.ExportComplianceCSV(AuditSearchFilter{})
	if !strings.Contains(csv, "SETTLEMENT") {
		t.Error("CSV export should contain SETTLEMENT category")
	}
	if !strings.Contains(csv, "SEBI_REPORT") {
		t.Error("CSV export should contain SEBI_REPORT action")
	}
	// Verify header row
	if !strings.Contains(csv, "Index,PreviousHash,RecordHash") {
		t.Error("CSV export should contain header row")
	}
}
