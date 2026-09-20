package keyspace

import (
	"sync"
	"time"
)

// DLQRecord represents a failed keyspace expiration execution.
type DLQRecord struct {
	Event       ExpiryEvent `json:"event"`
	ErrorMsg    string      `json:"error_msg"`
	RetryCount  int         `json:"retry_count"`
	FailedAt    time.Time   `json:"failed_at"`
	Resolved    bool        `json:"resolved"`
}

// DeadLetterQueue stores failed events for manual inspection or automated replay.
type DeadLetterQueue struct {
	mu      sync.Mutex
	records []DLQRecord
}

// NewDeadLetterQueue creates a new DLQ.
func NewDeadLetterQueue() *DeadLetterQueue {
	return &DeadLetterQueue{
		records: make([]DLQRecord, 0),
	}
}

// Enqueue adds a failed event to DLQ.
func (d *DeadLetterQueue) Enqueue(event ExpiryEvent, err error, retries int) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.records = append(d.records, DLQRecord{
		Event:      event,
		ErrorMsg:   err.Error(),
		RetryCount: retries,
		FailedAt:   time.Now(),
		Resolved:   false,
	})
}

// GetRecords returns a copy of DLQ records.
func (d *DeadLetterQueue) GetRecords() []DLQRecord {
	d.mu.Lock()
	defer d.mu.Unlock()

	res := make([]DLQRecord, len(d.records))
	copy(res, d.records)
	return res
}

// Len returns the count of records in DLQ.
func (d *DeadLetterQueue) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.records)
}
