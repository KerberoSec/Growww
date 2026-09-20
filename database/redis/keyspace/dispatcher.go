package keyspace

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrNoHandlerRegistered = errors.New("no handler registered for event type")
	ErrExecutionFailed     = errors.New("expiry handler execution failed")
)

// KeyspaceMetrics tracks execution statistics.
type KeyspaceMetrics struct {
	TotalEventsReceived   uint64
	TotalEventsDispatched uint64
	TotalEventsFailed     uint64
	TotalRetries          uint64
	DLQCount              uint64
}

// ExpiryDispatcher coordinates event ingestion, concurrency worker pool dispatching, and retries.
type ExpiryDispatcher struct {
	mu           sync.RWMutex
	handlers     map[EventType][]ExpiryHandler
	dlq          *DeadLetterQueue
	metrics      KeyspaceMetrics
	maxRetries   int
	retryBackoff time.Duration
}

// NewExpiryDispatcher initializes the dispatcher.
func NewExpiryDispatcher(maxRetries int, retryBackoff time.Duration) *ExpiryDispatcher {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	if retryBackoff <= 0 {
		retryBackoff = 10 * time.Millisecond
	}
	return &ExpiryDispatcher{
		handlers:     make(map[EventType][]ExpiryHandler),
		dlq:          NewDeadLetterQueue(),
		maxRetries:   maxRetries,
		retryBackoff: retryBackoff,
	}
}

// RegisterHandler binds an event type to a handler.
func (d *ExpiryDispatcher) RegisterHandler(eventType EventType, handler ExpiryHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Dispatch processes an incoming keyspace event with exponential backoff and DLQ routing.
func (d *ExpiryDispatcher) Dispatch(ctx context.Context, event ExpiryEvent) error {
	d.mu.Lock()
	d.metrics.TotalEventsReceived++
	handlers, exists := d.handlers[event.Type]
	d.mu.Unlock()

	if !exists || len(handlers) == 0 {
		// Log warning or route to default handler
		return fmt.Errorf("%w: %s", ErrNoHandlerRegistered, event.Type)
	}

	var dispatchErr error
	for _, handler := range handlers {
		var execErr error
		attempts := 0

		for attempts <= d.maxRetries {
			attempts++
			execErr = handler.HandleExpiry(ctx, event)
			if execErr == nil {
				d.mu.Lock()
				d.metrics.TotalEventsDispatched++
				d.mu.Unlock()
				break
			}

			if attempts <= d.maxRetries {
				d.mu.Lock()
				d.metrics.TotalRetries++
				d.mu.Unlock()
				time.Sleep(d.retryBackoff * time.Duration(1<<(attempts-1)))
			}
		}

		if execErr != nil {
			d.mu.Lock()
			d.metrics.TotalEventsFailed++
			d.metrics.DLQCount++
			d.mu.Unlock()

			d.dlq.Enqueue(event, execErr, attempts)
			dispatchErr = execErr
		}
	}

	return dispatchErr
}

// GetMetrics returns operational metrics.
func (d *ExpiryDispatcher) GetMetrics() KeyspaceMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	m := d.metrics
	m.DLQCount = uint64(d.dlq.Len())
	return m
}

// GetDLQ returns the dead-letter queue.
func (d *ExpiryDispatcher) GetDLQ() *DeadLetterQueue {
	return d.dlq
}
