package src

import (
	"testing"
	"time"
)

func TestServiceDiscoveryRegistrationAndResolution(t *testing.T) {
	client := NewConsulDiscoveryClient(5 * time.Second)

	inst1 := &ServiceInstance{
		ID:          "order-service-01",
		ServiceName: "order-service",
		Host:        "10.0.1.10",
		Port:        50051,
		Priority:    10,
		Weight:      100,
		Status:      HealthPassing,
		LatencyMs:   1.2,
	}

	inst2 := &ServiceInstance{
		ID:          "order-service-02",
		ServiceName: "order-service",
		Host:        "10.0.1.11",
		Port:        50051,
		Priority:    10,
		Weight:      100,
		Status:      HealthPassing,
		LatencyMs:   0.8,
	}

	// Register
	if err := client.RegisterInstance(inst1); err != nil {
		t.Fatalf("unexpected error registering inst1: %v", err)
	}
	if err := client.RegisterInstance(inst2); err != nil {
		t.Fatalf("unexpected error registering inst2: %v", err)
	}

	// Resolve
	healthy, err := client.Resolve("order-service")
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if len(healthy) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(healthy))
	}

	// Lowest latency
	best, err := client.GetLowestLatencyInstance("order-service")
	if err != nil {
		t.Fatalf("unexpected lowest latency error: %v", err)
	}
	if best.ID != "order-service-02" {
		t.Fatalf("expected order-service-02 with 0.8ms latency, got %s", best.ID)
	}

	// Round-robin selection
	first, _ := client.GetHealthyInstance("order-service")
	second, _ := client.GetHealthyInstance("order-service")
	if first.ID == second.ID {
		// With 2 instances and round robin, they should rotate
		third, _ := client.GetHealthyInstance("order-service")
		if first.ID == second.ID && second.ID == third.ID {
			t.Fatalf("expected round robin distribution across instances")
		}
	}

	// Health status update: degrade inst2
	if err := client.UpdateHealth("order-service-02", HealthCritical, 999.0); err != nil {
		t.Fatalf("failed to update health: %v", err)
	}

	healthyAfter, err := client.Resolve("order-service")
	if err != nil {
		t.Fatalf("failed resolving after degrade: %v", err)
	}
	if len(healthyAfter) != 1 || healthyAfter[0].ID != "order-service-01" {
		t.Fatalf("expected only order-service-01 healthy, got %+v", healthyAfter)
	}

	// Deregister
	if err := client.DeregisterInstance("order-service-01"); err != nil {
		t.Fatalf("failed to deregister: %v", err)
	}
	_, err = client.Resolve("order-service")
	if err == nil {
		t.Fatalf("expected error when resolving service with no healthy instances")
	}
}
