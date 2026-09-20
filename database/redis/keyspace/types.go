package keyspace

import (
	"context"
	"strings"
	"time"
)

// EventType categorizes keyspace expirations.
type EventType string

const (
	EventTypeOrderExpiry    EventType = "ORDER_EXPIRY"
	EventTypeMarginCall     EventType = "MARGIN_CALL"
	EventTypeSessionExpiry  EventType = "SESSION_EXPIRY"
	EventTypeRateLimitLock  EventType = "RATE_LIMIT_LOCK"
	EventTypeSLATimer       EventType = "SLA_TIMER"
	EventTypeUnknown        EventType = "UNKNOWN"
)

// ExpiryEvent represents a parsed keyspace expiration event.
type ExpiryEvent struct {
	RawKey      string            `json:"raw_key"`
	Channel     string            `json:"channel"`
	Type        EventType         `json:"type"`
	EntityID    string            `json:"entity_id"`
	SecondaryID string            `json:"secondary_id,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ParseKey extracts event semantics based on the key taxonomy:
// - order:exp:{orderId}:{symbol}
// - margin:call:{userId}:{marginId}
// - session:auth:{userId}:{sessionId}
// - lock:rate:{ip}:{clientId}
// - sla:timer:{taskId}:{type}
func ParseKey(key string) ExpiryEvent {
	parts := strings.Split(key, ":")
	event := ExpiryEvent{
		RawKey:    key,
		Timestamp: time.Now(),
		Metadata:  make(map[string]string),
	}

	if len(parts) >= 4 && parts[0] == "order" && parts[1] == "exp" {
		event.Type = EventTypeOrderExpiry
		event.EntityID = parts[2]
		event.SecondaryID = parts[3]
	} else if len(parts) >= 4 && parts[0] == "margin" && parts[1] == "call" {
		event.Type = EventTypeMarginCall
		event.EntityID = parts[2]
		event.SecondaryID = parts[3]
	} else if len(parts) >= 4 && parts[0] == "session" && parts[1] == "auth" {
		event.Type = EventTypeSessionExpiry
		event.EntityID = parts[2]
		event.SecondaryID = parts[3]
	} else if len(parts) >= 4 && parts[0] == "lock" && parts[1] == "rate" {
		event.Type = EventTypeRateLimitLock
		event.EntityID = parts[2]
		event.SecondaryID = parts[3]
	} else if len(parts) >= 4 && parts[0] == "sla" && parts[1] == "timer" {
		event.Type = EventTypeSLATimer
		event.EntityID = parts[2]
		event.SecondaryID = parts[3]
	} else {
		event.Type = EventTypeUnknown
		event.EntityID = key
	}

	return event
}

// ExpiryHandler processes an expiration event.
type ExpiryHandler interface {
	HandleExpiry(ctx context.Context, event ExpiryEvent) error
}

// ExpiryHandlerFunc allows using plain functions as handlers.
type ExpiryHandlerFunc func(ctx context.Context, event ExpiryEvent) error

func (f ExpiryHandlerFunc) HandleExpiry(ctx context.Context, event ExpiryEvent) error {
	return f(ctx, event)
}
