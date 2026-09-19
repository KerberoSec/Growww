package main

import (
	"fmt"
	"testing"
)

func TestEnqueueAndProcess(t *testing.T) {
	d := NewDispatcher(3)
	d.Enqueue(&RegulatoryEvent{EventID: "e1", EventType: "TRADE_REPORT", Target: "SEBI"})
	if d.QueueLen() != 1 { t.Error("expected 1 in queue") }
	err := d.ProcessNext(func(e *RegulatoryEvent) error { return nil })
	if err != nil { t.Fatal(err) }
	if d.SentCount() != 1 { t.Error("expected 1 sent") }
}

func TestRetryAndDLQ(t *testing.T) {
	d := NewDispatcher(2)
	d.Enqueue(&RegulatoryEvent{EventID: "e2", Target: "FIU_IND"})
	fail := func(e *RegulatoryEvent) error { return fmt.Errorf("network error") }
	d.ProcessNext(fail)
	if d.QueueLen() != 1 { t.Errorf("retry should re-queue, got %d", d.QueueLen()) }
	d.ProcessNext(fail)
	if d.DLQLen() != 1 { t.Errorf("expected DLQ after max retries, got %d", d.DLQLen()) }
}

func TestEmptyQueue(t *testing.T) {
	d := NewDispatcher(3)
	err := d.ProcessNext(func(e *RegulatoryEvent) error { return nil })
	if err == nil { t.Error("expected error on empty queue") }
}
