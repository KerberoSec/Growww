package opensearch

import (
	"context"
	"sort"
	"sync"
)

// AuditPipeline manages immutable audit trail storage and trade reconstitution.
type AuditPipeline struct {
	mu          sync.RWMutex
	auditEvents map[string]*AuditTrailDocument // event_id -> doc
	trades      map[string]*TradeDocument      // trade_id -> doc
	byEntity    map[string][]string            // entity_id -> event_ids
	byTrade     map[string][]string            // trade_id -> event_ids
}

func NewAuditPipeline() *AuditPipeline {
	return &AuditPipeline{
		auditEvents: make(map[string]*AuditTrailDocument),
		trades:      make(map[string]*TradeDocument),
		byEntity:    make(map[string][]string),
		byTrade:     make(map[string][]string),
	}
}

// IngestAuditEvent stores an audit record and verifies its cryptographic integrity.
func (p *AuditPipeline) IngestAuditEvent(event *AuditTrailDocument) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	expectedHash := event.CalculateChecksum()
	if event.TamperHash == "" {
		event.TamperHash = expectedHash
	} else if event.TamperHash != expectedHash {
		return ErrTamperedAuditTrail
	}

	p.auditEvents[event.EventID] = event
	p.byEntity[event.EntityID] = append(p.byEntity[event.EntityID], event.EventID)

	// If correlated with a trade
	if event.CorrelationID != "" {
		p.byTrade[event.CorrelationID] = append(p.byTrade[event.CorrelationID], event.EventID)
	}

	return nil
}

// IngestTrade stores an executed trade.
func (p *AuditPipeline) IngestTrade(trade *TradeDocument) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.trades[trade.TradeID] = trade
}

// ReconstituteTradeLifecycle reconstructs all events associated with a trade chronologically.
func (p *AuditPipeline) ReconstituteTradeLifecycle(ctx context.Context, tradeID string) (*ReconstitutedLifecycle, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	trade, exists := p.trades[tradeID]
	if !exists {
		return nil, ErrLifecycleReconstitute
	}

	// Gather audit events related to trade directly or via buyer/seller orders
	eventIDs := make(map[string]bool)
	for _, id := range p.byTrade[tradeID] {
		eventIDs[id] = true
	}
	for _, id := range p.byEntity[trade.BuyerOrderID] {
		eventIDs[id] = true
	}
	for _, id := range p.byEntity[trade.SellerOrderID] {
		eventIDs[id] = true
	}
	for _, id := range p.byEntity[tradeID] {
		eventIDs[id] = true
	}

	events := make([]*AuditTrailDocument, 0, len(eventIDs))
	integrityValid := true

	for id := range eventIDs {
		ev := p.auditEvents[id]
		if ev != nil {
			if ev.TamperHash != ev.CalculateChecksum() {
				integrityValid = false
			}
			events = append(events, ev)
		}
	}

	// Sort chronologically
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	return &ReconstitutedLifecycle{
		TradeID:        tradeID,
		BuyerOrderID:   trade.BuyerOrderID,
		SellerOrderID:  trade.SellerOrderID,
		AuditEvents:    events,
		Trade:          trade,
		IntegrityValid: integrityValid,
	}, nil
}
