package main

import (
	"testing"
)

func TestADLEngine_InsuranceFundAbsorption(t *testing.T) {
	engine := NewADLEngine(10000.0) // $10k insurance

	events, err := engine.ResolveDefault("defaulter-1", 5000.0, 65000.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("Expected 0 ADL events when insurance covers loss, got %d", len(events))
	}
	if engine.GetInsuranceFundBalance() != 5000.0 {
		t.Fatalf("Expected insurance balance 5000, got %.2f", engine.GetInsuranceFundBalance())
	}
}

func TestADLEngine_SocializedLossDistribution(t *testing.T) {
	engine := NewADLEngine(2000.0) // $2k insurance

	// Add counterparties to ADL queue
	engine.AddToQueue("cp-1", "pos-1", "LONG", 5000.0, 10.0)
	engine.AddToQueue("cp-2", "pos-2", "LONG", 3000.0, 5.0)

	// $8k loss: $2k from insurance, $6k from ADL
	events, err := engine.ResolveDefault("defaulter-1", 8000.0, 64000.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if engine.GetInsuranceFundBalance() != 0.0 {
		t.Fatalf("Insurance fund should be fully drained, got %.2f", engine.GetInsuranceFundBalance())
	}
	if len(events) < 1 {
		t.Fatalf("Expected at least 1 ADL event, got %d", len(events))
	}

	// Verify socialized total matches remaining loss
	var totalSocialized float64
	for _, evt := range events {
		totalSocialized += evt.SocializedLoss
	}
	if totalSocialized != 6000.0 {
		t.Fatalf("Expected total socialized loss 6000, got %.2f", totalSocialized)
	}
}

func TestADLEngine_QueuePriorityOrdering(t *testing.T) {
	engine := NewADLEngine(0)

	engine.AddToQueue("low", "pos-low", "LONG", 100.0, 2.0)   // score=200
	engine.AddToQueue("high", "pos-hi", "LONG", 500.0, 10.0)  // score=5000
	engine.AddToQueue("mid", "pos-mid", "LONG", 200.0, 5.0)   // score=1000

	queue := engine.GetQueue()
	if len(queue) != 3 {
		t.Fatalf("Expected 3 entries in queue, got %d", len(queue))
	}
	if queue[0].UserID != "high" {
		t.Errorf("Highest priority should be 'high', got %s", queue[0].UserID)
	}
	if queue[1].UserID != "mid" {
		t.Errorf("Second priority should be 'mid', got %s", queue[1].UserID)
	}
	if queue[2].UserID != "low" {
		t.Errorf("Third priority should be 'low', got %s", queue[2].UserID)
	}
}

func TestADLEngine_EmptyQueueError(t *testing.T) {
	engine := NewADLEngine(0) // No insurance, no queue

	_, err := engine.ResolveDefault("defaulter-1", 5000.0, 60000.0)
	if err == nil {
		t.Fatal("Expected error when ADL queue is empty and no insurance fund")
	}
}
