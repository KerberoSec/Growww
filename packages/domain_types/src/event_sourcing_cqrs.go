package src

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrConcurrencyConflict = errors.New("eventsourcing: optimistic concurrency conflict, version mismatch")
	ErrAggregateNotFound   = errors.New("eventsourcing: aggregate not found")
)

// DomainEvent represents an immutable event envelope in the append-only event log.
type DomainEvent struct {
	EventID        string    `json:"event_id"`
	AggregateID    string    `json:"aggregate_id"`
	SequenceNumber uint64    `json:"sequence_number"`
	EventType      string    `json:"event_type"`
	Payload        []byte    `json:"payload"`
	Timestamp      time.Time `json:"timestamp"`
}

// AggregateSnapshot represents a materialized point-in-time state checkpoint.
type AggregateSnapshot struct {
	AggregateID string    `json:"aggregate_id"`
	Version     uint64    `json:"version"`
	State       []byte    `json:"state"`
	Timestamp   time.Time `json:"timestamp"`
}

// EventStore defines the core append-only storage and retrieval contract for CQRS.
type EventStore interface {
	Append(aggregateID string, expectedVersion uint64, events []DomainEvent) error
	GetEvents(aggregateID string, fromVersion uint64) ([]DomainEvent, error)
	SaveSnapshot(snapshot AggregateSnapshot) error
	GetSnapshot(aggregateID string) (*AggregateSnapshot, error)
}

// InMemEventStore provides a thread-safe in-memory event store with OCC validation.
type InMemEventStore struct {
	mu        sync.RWMutex
	events    map[string][]DomainEvent
	snapshots map[string]*AggregateSnapshot
}

// NewInMemEventStore initializes a new CQRS event store.
func NewInMemEventStore() *InMemEventStore {
	return &InMemEventStore{
		events:    make(map[string][]DomainEvent),
		snapshots: make(map[string]*AggregateSnapshot),
	}
}

// Append validates optimistic concurrency control and writes events atomically.
func (s *InMemEventStore) Append(aggregateID string, expectedVersion uint64, events []DomainEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.events[aggregateID]
	currentVersion := uint64(len(existing))

	if currentVersion != expectedVersion {
		return fmt.Errorf("%w: current version %d does not match expected %d", ErrConcurrencyConflict, currentVersion, expectedVersion)
	}

	for i, evt := range events {
		seq := currentVersion + uint64(i) + 1
		evt.SequenceNumber = seq
		if evt.Timestamp.IsZero() {
			evt.Timestamp = time.Now().UTC()
		}
		existing = append(existing, evt)
	}

	s.events[aggregateID] = existing
	return nil
}

// GetEvents returns events for an aggregate starting after fromVersion.
func (s *InMemEventStore) GetEvents(aggregateID string, fromVersion uint64) ([]DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	existing, exists := s.events[aggregateID]
	if !exists {
		return []DomainEvent{}, nil
	}

	if fromVersion >= uint64(len(existing)) {
		return []DomainEvent{}, nil
	}

	result := make([]DomainEvent, len(existing[fromVersion:]))
	copy(result, existing[fromVersion:])
	return result, nil
}

// SaveSnapshot writes an aggregate checkpoint.
func (s *InMemEventStore) SaveSnapshot(snapshot AggregateSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if snapshot.Timestamp.IsZero() {
		snapshot.Timestamp = time.Now().UTC()
	}
	s.snapshots[snapshot.AggregateID] = &snapshot
	return nil
}

// GetSnapshot retrieves the latest snapshot if available.
func (s *InMemEventStore) GetSnapshot(aggregateID string) (*AggregateSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, exists := s.snapshots[aggregateID]
	if !exists {
		return nil, nil
	}
	return snapshot, nil
}
