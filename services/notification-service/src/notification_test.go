package notification

import (
	"context"
	"testing"
	"time"
)

func setupDispatcher(t *testing.T) (*Dispatcher, *MockProvider, *MockProvider) {
	t.Helper()
	te := NewTemplateEngine()
	ps := NewPreferenceStore()
	dt := NewDeliveryTracker()

	err := te.Register(&Template{
		ID:   "welcome",
		Name: "Welcome Email",
		Body: "Hello {{.Name}}, welcome to NBSE!",
	})
	if err != nil {
		t.Fatalf("failed to register template: %v", err)
	}

	emailProvider := NewMockProvider(ChannelEmail)
	smsProvider := NewMockProvider(ChannelSMS)

	d := NewDispatcher(te, ps, dt)
	d.RegisterProvider(emailProvider)
	d.RegisterProvider(smsProvider)

	return d, emailProvider, smsProvider
}

func TestTemplateEngine_RegisterAndRender(t *testing.T) {
	te := NewTemplateEngine()

	err := te.Register(&Template{
		ID:   "order_confirm",
		Name: "Order Confirmation",
		Body: "Order #{{.OrderID}} for {{.Symbol}} confirmed at ₹{{.Price}}",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	result, err := te.Render("order_confirm", map[string]string{
		"OrderID": "ORD-12345",
		"Symbol":  "RELIANCE",
		"Price":   "2450.00",
	})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	expected := "Order #ORD-12345 for RELIANCE confirmed at ₹2450.00"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestTemplateEngine_RenderMissingTemplate(t *testing.T) {
	te := NewTemplateEngine()
	_, err := te.Render("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for missing template")
	}
}

func TestTemplateEngine_RegisterInvalidTemplate(t *testing.T) {
	te := NewTemplateEngine()
	err := te.Register(&Template{
		ID:   "bad",
		Body: "{{.Unclosed",
	})
	if err == nil {
		t.Fatal("expected error for invalid template syntax")
	}
}

func TestPreferenceStore_SetAndGet(t *testing.T) {
	ps := NewPreferenceStore()
	pref := &UserPreference{
		UserID: "user-1",
		EnabledChannels: map[Channel]bool{
			ChannelEmail: true,
			ChannelSMS:   false,
		},
		Timezone: "Asia/Kolkata",
	}
	if err := ps.Set(pref); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := ps.Get("user-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !got.EnabledChannels[ChannelEmail] {
		t.Error("email should be enabled")
	}
	if got.EnabledChannels[ChannelSMS] {
		t.Error("sms should be disabled")
	}
}

func TestPreferenceStore_IsChannelEnabled(t *testing.T) {
	ps := NewPreferenceStore()
	_ = ps.Set(&UserPreference{
		UserID: "u1",
		EnabledChannels: map[Channel]bool{
			ChannelEmail: true,
			ChannelSMS:   false,
		},
	})
	if !ps.IsChannelEnabled("u1", ChannelEmail) {
		t.Error("email should be enabled for u1")
	}
	if ps.IsChannelEnabled("u1", ChannelSMS) {
		t.Error("sms should be disabled for u1")
	}
	// Unknown user defaults to enabled
	if !ps.IsChannelEnabled("unknown", ChannelEmail) {
		t.Error("unknown user should default to enabled")
	}
}

func TestPreferenceStore_QuietHours(t *testing.T) {
	ps := NewPreferenceStore()
	_ = ps.Set(&UserPreference{
		UserID:     "u1",
		QuietStart: 22,
		QuietEnd:   7,
		Timezone:   "UTC",
	})

	// 23:00 UTC should be quiet
	quietTime := time.Date(2025, 1, 1, 23, 0, 0, 0, time.UTC)
	if !ps.IsQuietHours("u1", quietTime) {
		t.Error("23:00 UTC should be quiet hours")
	}

	// 12:00 UTC should not be quiet
	activeTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	if ps.IsQuietHours("u1", activeTime) {
		t.Error("12:00 UTC should not be quiet hours")
	}
}

func TestDispatcher_DispatchSuccess(t *testing.T) {
	d, emailProvider, _ := setupDispatcher(t)
	ctx := context.Background()

	n, err := d.Dispatch(ctx, "user-1", ChannelEmail, "welcome", map[string]string{"Name": "Arjun"}, PriorityNormal)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}
	if n.Status != StatusSent {
		t.Errorf("expected status %s, got %s", StatusSent, n.Status)
	}
	if n.Body != "Hello Arjun, welcome to NBSE!" {
		t.Errorf("unexpected body: %s", n.Body)
	}
	if emailProvider.SentCount() != 1 {
		t.Errorf("expected 1 sent, got %d", emailProvider.SentCount())
	}
}

func TestDispatcher_DispatchChannelDisabled(t *testing.T) {
	d, _, _ := setupDispatcher(t)
	ctx := context.Background()

	// Disable email for user
	ps := NewPreferenceStore()
	_ = ps.Set(&UserPreference{
		UserID:          "user-2",
		EnabledChannels: map[Channel]bool{ChannelEmail: false},
	})
	d.preferences = ps

	_, err := d.Dispatch(ctx, "user-2", ChannelEmail, "welcome", map[string]string{"Name": "Test"}, PriorityNormal)
	if err == nil {
		t.Fatal("expected error for disabled channel")
	}
}

func TestDispatcher_DispatchProviderFailure(t *testing.T) {
	d, emailProvider, _ := setupDispatcher(t)
	ctx := context.Background()

	emailProvider.SetFailNext()
	n, err := d.Dispatch(ctx, "user-1", ChannelEmail, "welcome", map[string]string{"Name": "Test"}, PriorityNormal)
	if err == nil {
		t.Fatal("expected error for provider failure")
	}
	if n.Status != StatusFailed {
		t.Errorf("expected status %s, got %s", StatusFailed, n.Status)
	}
}

func TestDeliveryTracker_Lifecycle(t *testing.T) {
	dt := NewDeliveryTracker()

	n := &Notification{
		ID:     "notif-1",
		UserID: "user-1",
		Status: StatusPending,
	}
	dt.Track(n)

	got, err := dt.Get("notif-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got.Status != StatusPending {
		t.Errorf("expected pending, got %s", got.Status)
	}

	_ = dt.UpdateStatus("notif-1", StatusSent, "")
	got, _ = dt.Get("notif-1")
	if got.Status != StatusSent {
		t.Errorf("expected sent, got %s", got.Status)
	}
	if got.SentAt == nil {
		t.Error("SentAt should be set")
	}

	_ = dt.UpdateStatus("notif-1", StatusDelivered, "")
	got, _ = dt.Get("notif-1")
	if got.DeliveredAt == nil {
		t.Error("DeliveredAt should be set")
	}

	byUser := dt.GetByUser("user-1")
	if len(byUser) != 1 {
		t.Errorf("expected 1 notification for user, got %d", len(byUser))
	}
}

func TestDispatcher_MultiChannelDispatch(t *testing.T) {
	te := NewTemplateEngine()
	ps := NewPreferenceStore()
	dt := NewDeliveryTracker()

	_ = te.Register(&Template{ID: "alert", Body: "Price alert: {{.Symbol}}"})

	d := NewDispatcher(te, ps, dt)
	d.RegisterProvider(NewMockProvider(ChannelEmail))
	d.RegisterProvider(NewMockProvider(ChannelSMS))
	d.RegisterProvider(NewMockProvider(ChannelPush))
	d.RegisterProvider(NewMockProvider(ChannelWebSocket))

	ctx := context.Background()
	notifications, errs := d.DispatchMultiChannel(ctx, "user-1", "alert", map[string]string{"Symbol": "INFY"}, PriorityHigh)

	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}
	if len(notifications) != 4 {
		t.Errorf("expected 4 notifications, got %d", len(notifications))
	}
}
