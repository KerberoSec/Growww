package main

import (
	"fmt"
	"sync"
	"time"
)

type RegulatoryEvent struct {
	EventID   string
	EventType string // TRADE_REPORT, POSITION_REPORT, AML_ALERT, TAX_FILING
	Payload   string
	Target    string // SEBI, RBI, FIU_IND
	CreatedAt time.Time
	Status    string // QUEUED, SENT, FAILED, DLQ
	Retries   int
}

type Dispatcher struct {
	mu     sync.Mutex
	queue  []*RegulatoryEvent
	dlq    []*RegulatoryEvent
	sent   []*RegulatoryEvent
	maxRet int
}

func NewDispatcher(maxRetries int) *Dispatcher {
	return &Dispatcher{maxRet: maxRetries}
}

func (d *Dispatcher) Enqueue(evt *RegulatoryEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()
	evt.Status = "QUEUED"
	evt.CreatedAt = time.Now().UTC()
	d.queue = append(d.queue, evt)
}

func (d *Dispatcher) ProcessNext(sendFn func(*RegulatoryEvent) error) error {
	d.mu.Lock()
	if len(d.queue) == 0 {
		d.mu.Unlock()
		return fmt.Errorf("queue empty")
	}
	evt := d.queue[0]
	d.queue = d.queue[1:]
	d.mu.Unlock()

	err := sendFn(evt)
	d.mu.Lock()
	defer d.mu.Unlock()
	if err != nil {
		evt.Retries++
		if evt.Retries >= d.maxRet {
			evt.Status = "DLQ"
			d.dlq = append(d.dlq, evt)
		} else {
			evt.Status = "QUEUED"
			d.queue = append(d.queue, evt)
		}
		return err
	}
	evt.Status = "SENT"
	d.sent = append(d.sent, evt)
	return nil
}

func (d *Dispatcher) QueueLen() int   { d.mu.Lock(); defer d.mu.Unlock(); return len(d.queue) }
func (d *Dispatcher) DLQLen() int     { d.mu.Lock(); defer d.mu.Unlock(); return len(d.dlq) }
func (d *Dispatcher) SentCount() int  { d.mu.Lock(); defer d.mu.Unlock(); return len(d.sent) }
