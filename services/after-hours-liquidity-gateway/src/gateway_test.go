package afterhours

import (
	"testing"
)

func TestSessionLifecycle(t *testing.T) {
	gw := NewLiquidityGateway()

	if gw.ActiveSession() != SessionClosed {
		t.Errorf("expected CLOSED initially, got %s", gw.ActiveSession())
	}

	if err := gw.OpenSession(SessionPreMarket); err != nil {
		t.Fatalf("OpenSession failed: %v", err)
	}
	if gw.ActiveSession() != SessionPreMarket {
		t.Errorf("expected PRE_MARKET, got %s", gw.ActiveSession())
	}

	gw.CloseSession()
	if gw.ActiveSession() != SessionClosed {
		t.Errorf("expected CLOSED after close, got %s", gw.ActiveSession())
	}

	// Cannot open REGULAR or CLOSED
	if err := gw.OpenSession(SessionRegular); err == nil {
		t.Error("expected error opening REGULAR session")
	}
	if err := gw.OpenSession(SessionClosed); err == nil {
		t.Error("expected error opening CLOSED session")
	}
}

func TestSpreadQuote(t *testing.T) {
	gw := NewLiquidityGateway()
	gw.SetReferencePrice("BTC-INR", 5_000_000)
	gw.OpenSession(SessionPreMarket)

	quote, err := gw.GetSpreadQuote("BTC-INR")
	if err != nil {
		t.Fatalf("GetSpreadQuote failed: %v", err)
	}

	// Pre-market: 2x multiplier, base spread = 0.1% = 5000
	// Widened = 10000, half = 5000
	if quote.Multiplier != 2.0 {
		t.Errorf("expected multiplier 2.0, got %.1f", quote.Multiplier)
	}
	if quote.BidPrice >= quote.AskPrice {
		t.Error("bid should be less than ask")
	}
	if quote.Spread <= 0 {
		t.Error("spread should be positive")
	}

	// No reference price
	_, err = gw.GetSpreadQuote("ETH-INR")
	if err == nil {
		t.Error("expected error for missing reference price")
	}

	// Session closed
	gw.CloseSession()
	_, err = gw.GetSpreadQuote("BTC-INR")
	if err == nil {
		t.Error("expected error when session closed")
	}
}

func TestSubmitOrderValidation(t *testing.T) {
	gw := NewLiquidityGateway()
	gw.SetReferencePrice("BTC-INR", 5_000_000)
	gw.OpenSession(SessionPreMarket)

	// Valid limit order
	order := &Order{
		ID:       "o-1",
		UserID:   "u-1",
		Symbol:   "BTC-INR",
		Side:     SideBuy,
		Type:     OrderLimit,
		Price:    5_000_000,
		Quantity: 0.5,
	}
	if err := gw.SubmitOrder(order); err != nil {
		t.Fatalf("SubmitOrder failed: %v", err)
	}
	if order.Status != "ROUTED" {
		t.Errorf("expected ROUTED, got %s", order.Status)
	}

	// Duplicate order ID
	order2 := &Order{ID: "o-1", UserID: "u-1", Symbol: "BTC-INR", Side: SideBuy, Type: OrderLimit, Price: 5_000_000, Quantity: 0.1}
	if err := gw.SubmitOrder(order2); err == nil {
		t.Error("expected error for duplicate order ID")
	}
}

func TestMarketOrderRejected(t *testing.T) {
	gw := NewLiquidityGateway()
	gw.SetReferencePrice("BTC-INR", 5_000_000)
	gw.OpenSession(SessionPreMarket)

	order := &Order{
		ID:       "o-2",
		UserID:   "u-1",
		Symbol:   "BTC-INR",
		Side:     SideBuy,
		Type:     OrderMarket,
		Price:    5_000_000,
		Quantity: 0.1,
	}
	err := gw.SubmitOrder(order)
	if err == nil {
		t.Error("expected market order to be rejected in pre-market")
	}
	if order.Status != "REJECTED" {
		t.Errorf("expected REJECTED status, got %s", order.Status)
	}
}

func TestOrderSizeLimitExceeded(t *testing.T) {
	gw := NewLiquidityGateway()
	gw.SetReferencePrice("BTC-INR", 5_000_000)
	gw.OpenSession(SessionPreMarket) // max 5M INR

	// 5M price * 2 qty = 10M > 5M limit
	order := &Order{
		ID:       "o-3",
		UserID:   "u-1",
		Symbol:   "BTC-INR",
		Side:     SideBuy,
		Type:     OrderLimit,
		Price:    5_000_000,
		Quantity: 2.0,
	}
	err := gw.SubmitOrder(order)
	if err == nil {
		t.Error("expected order to be rejected for exceeding size limit")
	}
	if order.Status != "REJECTED" {
		t.Errorf("expected REJECTED, got %s", order.Status)
	}
}

func TestCancelOrder(t *testing.T) {
	gw := NewLiquidityGateway()
	gw.SetReferencePrice("BTC-INR", 5_000_000)
	gw.OpenSession(SessionPostMarket)

	order := &Order{
		ID:       "o-4",
		UserID:   "u-1",
		Symbol:   "BTC-INR",
		Side:     SideSell,
		Type:     OrderLimit,
		Price:    5_000_000,
		Quantity: 0.1,
	}
	gw.SubmitOrder(order)

	if err := gw.CancelOrder("o-4"); err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
	o, _ := gw.GetOrder("o-4")
	if o.Status != "CANCELLED" {
		t.Errorf("expected CANCELLED, got %s", o.Status)
	}

	// Cannot cancel already cancelled
	if err := gw.CancelOrder("o-4"); err == nil {
		t.Error("expected error cancelling already cancelled order")
	}
}

func TestNoOrdersWhenClosed(t *testing.T) {
	gw := NewLiquidityGateway()
	order := &Order{
		ID:       "o-5",
		UserID:   "u-1",
		Symbol:   "BTC-INR",
		Side:     SideBuy,
		Type:     OrderLimit,
		Price:    5_000_000,
		Quantity: 0.1,
	}
	err := gw.SubmitOrder(order)
	if err == nil {
		t.Error("expected error submitting order when session closed")
	}
}
