package main

import (
	"testing"
)

func TestSpotOrderLifecycle_SubmitAndFill(t *testing.T) {
	svc := NewSpotOrderLifecycleService()

	order, err := svc.SubmitOrder("ord-1", "user-1", "BTC/USDT", "BUY", 6500000000000, 100000000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if order.Status != OrderNew {
		t.Fatalf("Expected status NEW, got %s", order.Status)
	}

	// Partial fill
	order, err = svc.RecordFill("ord-1", 50000000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if order.Status != OrderPartiallyFilled {
		t.Fatalf("Expected PARTIALLY_FILLED, got %s", order.Status)
	}

	// Complete fill
	order, err = svc.RecordFill("ord-1", 50000000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if order.Status != OrderFilled {
		t.Fatalf("Expected FILLED, got %s", order.Status)
	}
}

func TestSpotOrderLifecycle_RejectInvalidSymbol(t *testing.T) {
	svc := NewSpotOrderLifecycleService()

	_, err := svc.SubmitOrder("ord-2", "user-1", "ETH/USDT", "BUY", 300000000000, 100000000)
	if err == nil {
		t.Fatal("Expected error for unsupported symbol")
	}
}

func TestSpotOrderRouter_Validation(t *testing.T) {
	svc := NewSpotOrderLifecycleService()
	router := NewSpotOrderRouter(svc, 1000, 100000000000)

	// Valid order
	result := router.ValidateOrder("BTC/USDT", "BUY", 6500000000000, 100000000)
	if !result.Valid {
		t.Fatalf("Expected valid order, got errors: %v", result.Errors)
	}

	// Invalid symbol
	result = router.ValidateOrder("ETH/USDT", "BUY", 6500000000000, 100000000)
	if result.Valid {
		t.Fatal("Expected invalid for unsupported symbol")
	}

	// Quantity too small
	result = router.ValidateOrder("BTC/USDT", "BUY", 6500000000000, 100)
	if result.Valid {
		t.Fatal("Expected invalid for quantity below minimum")
	}

	// Invalid side
	result = router.ValidateOrder("BTC/USDT", "HOLD", 6500000000000, 100000000)
	if result.Valid {
		t.Fatal("Expected invalid for bad side")
	}
}

func TestSpotOrderRouter_CancelOrder(t *testing.T) {
	svc := NewSpotOrderLifecycleService()
	router := NewSpotOrderRouter(svc, 1000, 100000000000)

	_, err := router.RouteOrder("ord-3", "user-1", "BTC/USDT", "SELL", 6600000000000, 50000000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	cancelResult, err := router.CancelOrder("ord-3")
	if err != nil {
		t.Fatalf("Unexpected error cancelling: %v", err)
	}
	if cancelResult.NewStatus != OrderCanceled {
		t.Fatalf("Expected CANCELED status, got %s", cancelResult.NewStatus)
	}

	// Should not be able to cancel again
	_, err = router.CancelOrder("ord-3")
	if err == nil {
		t.Fatal("Expected error cancelling already cancelled order")
	}
}

func TestSpotOrderRouter_GetOrdersByUser(t *testing.T) {
	svc := NewSpotOrderLifecycleService()
	router := NewSpotOrderRouter(svc, 1000, 100000000000)

	router.RouteOrder("o1", "trader-A", "BTC/USDT", "BUY", 6500000000000, 50000000)
	router.RouteOrder("o2", "trader-A", "BTC/USDT", "SELL", 6600000000000, 30000000)
	router.RouteOrder("o3", "trader-B", "BTC/USDT", "BUY", 6550000000000, 40000000)

	orders := router.GetOrdersByUser("trader-A")
	if len(orders) != 2 {
		t.Fatalf("Expected 2 orders for trader-A, got %d", len(orders))
	}
}
