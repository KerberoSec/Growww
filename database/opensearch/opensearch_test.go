package opensearch

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBulkIndexerIngestAndFlush(t *testing.T) {
	var flushedItems []interface{}
	var mu sync.Mutex

	callback := func(ctx context.Context, batch []interface{}) error {
		mu.Lock()
		defer mu.Unlock()
		flushedItems = append(flushedItems, batch...)
		return nil
	}

	// Batch size: 5, max buffer: 10
	indexer := NewBulkIndexer(5, 10, 500*time.Millisecond, callback)
	defer indexer.Stop()

	// Ingest 4 items (should remain in buffer)
	for i := 0; i < 4; i++ {
		doc := &TradeDocument{TradeID: "trade-b1"}
		if err := indexer.Ingest(context.Background(), doc); err != nil {
			t.Fatalf("unexpected ingest error: %v", err)
		}
	}

	mu.Lock()
	if len(flushedItems) != 0 {
		t.Errorf("expected 0 flushed items before batch threshold, got %d", len(flushedItems))
	}
	mu.Unlock()

	// Ingest 5th item: batch threshold reached, immediate flush triggered
	_ = indexer.Ingest(context.Background(), &TradeDocument{TradeID: "trade-b5"})

	mu.Lock()
	if len(flushedItems) != 5 {
		t.Errorf("expected 5 flushed items after batch threshold, got %d", len(flushedItems))
	}
	mu.Unlock()

	metrics := indexer.Metrics()
	if metrics.TotalIndexed != 5 {
		t.Errorf("expected TotalIndexed 5, got %d", metrics.TotalIndexed)
	}
}

func TestAuditTrailTamperDetection(t *testing.T) {
	pipeline := NewAuditPipeline()

	event := &AuditTrailDocument{
		EventID:       "evt-101",
		Action:        "ORDER_PLACED",
		ActorID:       "usr-1",
		EntityID:      "ord-1",
		CorrelationID: "corr-1",
		Timestamp:     time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC),
	}

	// 1. Valid event ingestion (computes and sets valid hash)
	if err := pipeline.IngestAuditEvent(event); err != nil {
		t.Fatalf("expected valid ingestion, got %v", err)
	}

	// 2. Tampered event with invalid checksum
	tamperedEvent := &AuditTrailDocument{
		EventID:       "evt-102",
		Action:        "BALANCE_CREDITED",
		ActorID:       "usr-attacker",
		EntityID:      "acc-1",
		CorrelationID: "corr-2",
		TamperHash:    "deadbeefcafebabe00001111222233334444555566667777888899990000aaaa",
		Timestamp:     time.Date(2026, 9, 20, 10, 5, 0, 0, time.UTC),
	}
	if err := pipeline.IngestAuditEvent(tamperedEvent); err != ErrTamperedAuditTrail {
		t.Fatalf("expected ErrTamperedAuditTrail, got %v", err)
	}
}

func TestTradeLifecycleReconstitution(t *testing.T) {
	pipeline := NewAuditPipeline()

	trade := &TradeDocument{
		TradeID:         "trade-lifecycle-1",
		ISIN:            "INE002A01018",
		Symbol:          "RELIANCE",
		BuyerOrderID:    "ord-buyer-1",
		SellerOrderID:   "ord-seller-1",
		PricePerUnitINR: 2500.00,
		FractionalUnits: 10.0,
		ExecutedAt:      time.Date(2026, 9, 20, 10, 1, 0, 0, time.UTC),
	}
	pipeline.IngestTrade(trade)

	// Timeline events
	ev1 := &AuditTrailDocument{
		EventID:       "ev-1",
		Action:        "ORDER_CREATED",
		EntityID:      "ord-buyer-1",
		CorrelationID: "trade-lifecycle-1",
		Timestamp:     time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC),
	}
	ev2 := &AuditTrailDocument{
		EventID:       "ev-2",
		Action:        "ORDER_MATCHED",
		EntityID:      "trade-lifecycle-1",
		CorrelationID: "trade-lifecycle-1",
		Timestamp:     time.Date(2026, 9, 20, 10, 1, 0, 0, time.UTC),
	}
	ev3 := &AuditTrailDocument{
		EventID:       "ev-3",
		Action:        "SETTLEMENT_DVP_CONFIRMED",
		EntityID:      "trade-lifecycle-1",
		CorrelationID: "trade-lifecycle-1",
		Timestamp:     time.Date(2026, 9, 20, 10, 1, 2, 0, time.UTC),
	}

	_ = pipeline.IngestAuditEvent(ev3) // Out of order insert
	_ = pipeline.IngestAuditEvent(ev1)
	_ = pipeline.IngestAuditEvent(ev2)

	reconstituted, err := pipeline.ReconstituteTradeLifecycle(context.Background(), "trade-lifecycle-1")
	if err != nil {
		t.Fatalf("failed to reconstitute trade lifecycle: %v", err)
	}

	if !reconstituted.IntegrityValid {
		t.Errorf("expected lifecycle integrity to be valid")
	}

	if len(reconstituted.AuditEvents) != 3 {
		t.Fatalf("expected 3 audit events, got %d", len(reconstituted.AuditEvents))
	}

	// Must be chronologically sorted
	if reconstituted.AuditEvents[0].Action != "ORDER_CREATED" {
		t.Errorf("expected first action to be ORDER_CREATED, got %s", reconstituted.AuditEvents[0].Action)
	}
	if reconstituted.AuditEvents[2].Action != "SETTLEMENT_DVP_CONFIRMED" {
		t.Errorf("expected last action to be SETTLEMENT_DVP_CONFIRMED, got %s", reconstituted.AuditEvents[2].Action)
	}
}

func TestMarketSurveillanceWashTrading(t *testing.T) {
	surveillance := NewMarketSurveillanceEngine()

	// Direct self-trade (buyer == seller address)
	washTrade := &TradeDocument{
		TradeID:         "wash-1",
		ISIN:            "INE002A01018",
		Symbol:          "RELIANCE",
		BuyerAddress:    "0x71C841832046a64ce5560A8146F255aB87C29bA2",
		SellerAddress:   "0x71C841832046a64ce5560A8146F255aB87C29bA2",
		PricePerUnitINR: 2500.0,
		FractionalUnits: 50.0,
	}

	alert := surveillance.AnalyzeTradeForWashTrading(context.Background(), washTrade)
	if alert == nil {
		t.Fatal("expected wash trading alert for identical buyer/seller address, got nil")
	}

	if alert.AlertType != AlertWashTrading || alert.Severity != SeverityCritical {
		t.Errorf("expected Critical WashTrading alert, got %v / %v", alert.AlertType, alert.Severity)
	}
}

func TestMarketSurveillanceSpoofing(t *testing.T) {
	surveillance := NewMarketSurveillanceEngine()

	now := time.Now()
	// User placed 4 orders and immediately cancelled all 4 within 50ms
	events := []OrderEvent{
		{OrderID: "o1", UserID: "usr-spoof", Price: 2500.0, PlacedAt: now, CanceledAt: now.Add(20 * time.Millisecond), IsCanceled: true},
		{OrderID: "o2", UserID: "usr-spoof", Price: 2501.0, PlacedAt: now, CanceledAt: now.Add(30 * time.Millisecond), IsCanceled: true},
		{OrderID: "o3", UserID: "usr-spoof", Price: 2502.0, PlacedAt: now, CanceledAt: now.Add(40 * time.Millisecond), IsCanceled: true},
		{OrderID: "o4", UserID: "usr-spoof", Price: 2503.0, PlacedAt: now, CanceledAt: now.Add(45 * time.Millisecond), IsCanceled: true},
	}

	alert := surveillance.AnalyzeOrdersForSpoofing(context.Background(), "INE002A01018", events, 100*time.Millisecond)
	if alert == nil {
		t.Fatal("expected spoofing alert, got nil")
	}
	if alert.AlertType != AlertSpoofing || alert.Severity != SeverityHigh {
		t.Errorf("expected High Spoofing alert, got %v / %v", alert.AlertType, alert.Severity)
	}
}

func TestQueryBuilderAndOHLCV(t *testing.T) {
	qb := NewQueryBuilder()
	qb.Term("isin", "INE002A01018")
	qb.Match("settlement_status", "SETTLED_DVP")
	qb.SetPagination(0, 25)

	dsl, err := qb.Build()
	if err != nil {
		t.Fatalf("failed to build query DSL: %v", err)
	}
	if !strings.Contains(dsl, "INE002A01018") || !strings.Contains(dsl, "SETTLED_DVP") {
		t.Errorf("query DSL missing expected terms: %s", dsl)
	}

	// OHLCV aggregation
	t0 := time.Date(2026, 9, 20, 10, 0, 5, 0, time.UTC)
	trades := []*TradeDocument{
		{ISIN: "INE002A01018", PricePerUnitINR: 100.0, FractionalUnits: 10.0, ExecutedAt: t0},
		{ISIN: "INE002A01018", PricePerUnitINR: 105.0, FractionalUnits: 5.0, ExecutedAt: t0.Add(10 * time.Second)},
		{ISIN: "INE002A01018", PricePerUnitINR: 98.0, FractionalUnits: 8.0, ExecutedAt: t0.Add(20 * time.Second)},
		{ISIN: "INE002A01018", PricePerUnitINR: 102.0, FractionalUnits: 12.0, ExecutedAt: t0.Add(30 * time.Second)},
	}

	bars := ComputeOHLCV(trades, 1*time.Minute)
	if len(bars) != 1 {
		t.Fatalf("expected 1 bar, got %d", len(bars))
	}

	bar := bars[0]
	if bar.Open != 100.0 || bar.High != 105.0 || bar.Low != 98.0 || bar.Close != 102.0 || bar.Volume != 35.0 {
		t.Errorf("OHLCV mismatch: open=%f high=%f low=%f close=%f vol=%f", bar.Open, bar.High, bar.Low, bar.Close, bar.Volume)
	}
}
