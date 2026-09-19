// Package notification provides a production-grade multi-channel notification
// dispatcher for the NBSE sovereign exchange. It supports email, SMS, push
// notifications, and in-app WebSocket delivery with template rendering,
// delivery tracking, and user preference management.
package notification

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"text/template"
	"time"
)

// Channel represents a notification delivery channel.
type Channel string

const (
	ChannelEmail     Channel = "email"
	ChannelSMS       Channel = "sms"
	ChannelPush      Channel = "push"
	ChannelWebSocket Channel = "websocket"
)

// DeliveryStatus tracks the lifecycle of a notification.
type DeliveryStatus string

const (
	StatusPending   DeliveryStatus = "pending"
	StatusSent      DeliveryStatus = "sent"
	StatusDelivered DeliveryStatus = "delivered"
	StatusFailed    DeliveryStatus = "failed"
	StatusBounced   DeliveryStatus = "bounced"
)

// Priority defines notification urgency.
type Priority int

const (
	PriorityLow    Priority = 0
	PriorityNormal Priority = 1
	PriorityHigh   Priority = 2
	PriorityCritical Priority = 3
)

// UserPreference stores per-user channel preferences and quiet hours.
type UserPreference struct {
	UserID       string
	EnabledChannels map[Channel]bool
	QuietStart   int // hour of day (0-23)
	QuietEnd     int // hour of day (0-23)
	Timezone     string
	UpdatedAt    time.Time
}

// Template represents a notification template with variable substitution.
type Template struct {
	ID       string
	Name     string
	Channel  Channel
	Subject  string // for email
	Body     string // Go text/template syntax
	Version  int
	CreatedAt time.Time
}

// Notification represents a single notification to be dispatched.
type Notification struct {
	ID          string
	UserID      string
	Channel     Channel
	TemplateID  string
	Params      map[string]string
	Priority    Priority
	Subject     string
	Body        string
	Status      DeliveryStatus
	CreatedAt   time.Time
	SentAt      *time.Time
	DeliveredAt *time.Time
	FailReason  string
	RetryCount  int
	MaxRetries  int
}

// ChannelProvider is an interface that channel-specific senders must implement.
type ChannelProvider interface {
	Send(ctx context.Context, n *Notification) error
	Channel() Channel
}

// TemplateEngine renders notification templates with parameters.
type TemplateEngine struct {
	mu        sync.RWMutex
	templates map[string]*Template
	compiled  map[string]*template.Template
}

// NewTemplateEngine creates a new template engine.
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*Template),
		compiled:  make(map[string]*template.Template),
	}
}

// Register adds or updates a template in the engine.
func (te *TemplateEngine) Register(tmpl *Template) error {
	if tmpl.ID == "" || tmpl.Body == "" {
		return errors.New("template ID and body are required")
	}
	compiled, err := template.New(tmpl.ID).Parse(tmpl.Body)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}
	te.mu.Lock()
	defer te.mu.Unlock()
	te.templates[tmpl.ID] = tmpl
	te.compiled[tmpl.ID] = compiled
	return nil
}

// Render renders a template with the given parameters.
func (te *TemplateEngine) Render(templateID string, params map[string]string) (string, error) {
	te.mu.RLock()
	compiled, ok := te.compiled[templateID]
	te.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("template %q not found", templateID)
	}
	var buf bytes.Buffer
	if err := compiled.Execute(&buf, params); err != nil {
		return "", fmt.Errorf("template render failed: %w", err)
	}
	return buf.String(), nil
}

// PreferenceStore manages user notification preferences.
type PreferenceStore struct {
	mu    sync.RWMutex
	prefs map[string]*UserPreference
}

// NewPreferenceStore creates a new preference store.
func NewPreferenceStore() *PreferenceStore {
	return &PreferenceStore{
		prefs: make(map[string]*UserPreference),
	}
}

// Set stores or updates user preferences.
func (ps *PreferenceStore) Set(pref *UserPreference) error {
	if pref.UserID == "" {
		return errors.New("user ID is required")
	}
	pref.UpdatedAt = time.Now()
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.prefs[pref.UserID] = pref
	return nil
}

// Get retrieves user preferences.
func (ps *PreferenceStore) Get(userID string) (*UserPreference, error) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.prefs[userID]
	if !ok {
		return nil, fmt.Errorf("preferences not found for user %q", userID)
	}
	return p, nil
}

// IsChannelEnabled checks if a channel is enabled for a user.
func (ps *PreferenceStore) IsChannelEnabled(userID string, ch Channel) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.prefs[userID]
	if !ok {
		return true // default: all channels enabled
	}
	enabled, exists := p.EnabledChannels[ch]
	if !exists {
		return true // default: enabled if not explicitly set
	}
	return enabled
}

// IsQuietHours checks if it's currently quiet hours for the user.
func (ps *PreferenceStore) IsQuietHours(userID string, now time.Time) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	p, ok := ps.prefs[userID]
	if !ok {
		return false
	}
	loc, err := time.LoadLocation(p.Timezone)
	if err != nil {
		loc = time.UTC
	}
	hour := now.In(loc).Hour()
	if p.QuietStart <= p.QuietEnd {
		return hour >= p.QuietStart && hour < p.QuietEnd
	}
	// wraps midnight, e.g. 22-07
	return hour >= p.QuietStart || hour < p.QuietEnd
}

// DeliveryTracker tracks notification delivery status and history.
type DeliveryTracker struct {
	mu      sync.RWMutex
	records map[string]*Notification
}

// NewDeliveryTracker creates a new tracker.
func NewDeliveryTracker() *DeliveryTracker {
	return &DeliveryTracker{
		records: make(map[string]*Notification),
	}
}

// Track starts tracking a notification.
func (dt *DeliveryTracker) Track(n *Notification) {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	dt.records[n.ID] = n
}

// UpdateStatus updates the delivery status of a notification.
func (dt *DeliveryTracker) UpdateStatus(notifID string, status DeliveryStatus, reason string) error {
	dt.mu.Lock()
	defer dt.mu.Unlock()
	n, ok := dt.records[notifID]
	if !ok {
		return fmt.Errorf("notification %q not found", notifID)
	}
	n.Status = status
	now := time.Now()
	switch status {
	case StatusSent:
		n.SentAt = &now
	case StatusDelivered:
		n.DeliveredAt = &now
	case StatusFailed, StatusBounced:
		n.FailReason = reason
	}
	return nil
}

// Get retrieves the current state of a notification.
func (dt *DeliveryTracker) Get(notifID string) (*Notification, error) {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	n, ok := dt.records[notifID]
	if !ok {
		return nil, fmt.Errorf("notification %q not found", notifID)
	}
	return n, nil
}

// GetByUser retrieves all notifications for a user.
func (dt *DeliveryTracker) GetByUser(userID string) []*Notification {
	dt.mu.RLock()
	defer dt.mu.RUnlock()
	var result []*Notification
	for _, n := range dt.records {
		if n.UserID == userID {
			result = append(result, n)
		}
	}
	return result
}

// Dispatcher is the main notification dispatch engine.
type Dispatcher struct {
	providers   map[Channel]ChannelProvider
	templates   *TemplateEngine
	preferences *PreferenceStore
	tracker     *DeliveryTracker
	mu          sync.RWMutex
}

// NewDispatcher creates a new notification dispatcher.
func NewDispatcher(te *TemplateEngine, ps *PreferenceStore, dt *DeliveryTracker) *Dispatcher {
	return &Dispatcher{
		providers:   make(map[Channel]ChannelProvider),
		templates:   te,
		preferences: ps,
		tracker:     dt,
	}
}

// RegisterProvider registers a channel provider.
func (d *Dispatcher) RegisterProvider(p ChannelProvider) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.providers[p.Channel()] = p
}

// generateID creates a unique notification ID.
func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// Dispatch sends a notification through the appropriate channel.
func (d *Dispatcher) Dispatch(ctx context.Context, userID string, channel Channel, templateID string, params map[string]string, priority Priority) (*Notification, error) {
	// Check user preferences
	if !d.preferences.IsChannelEnabled(userID, channel) {
		return nil, fmt.Errorf("channel %s is disabled for user %s", channel, userID)
	}

	// Check quiet hours for non-critical notifications
	if priority < PriorityCritical && d.preferences.IsQuietHours(userID, time.Now()) {
		return nil, fmt.Errorf("quiet hours active for user %s, notification deferred", userID)
	}

	// Render template
	body, err := d.templates.Render(templateID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	n := &Notification{
		ID:         generateID(),
		UserID:     userID,
		Channel:    channel,
		TemplateID: templateID,
		Params:     params,
		Priority:   priority,
		Body:       body,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		MaxRetries: 3,
	}

	d.tracker.Track(n)

	// Get provider
	d.mu.RLock()
	provider, ok := d.providers[channel]
	d.mu.RUnlock()
	if !ok {
		_ = d.tracker.UpdateStatus(n.ID, StatusFailed, "no provider registered for channel")
		return n, fmt.Errorf("no provider registered for channel %s", channel)
	}

	// Send
	if err := provider.Send(ctx, n); err != nil {
		_ = d.tracker.UpdateStatus(n.ID, StatusFailed, err.Error())
		return n, fmt.Errorf("send failed: %w", err)
	}

	_ = d.tracker.UpdateStatus(n.ID, StatusSent, "")
	return n, nil
}

// DispatchMultiChannel sends a notification through all enabled channels for a user.
func (d *Dispatcher) DispatchMultiChannel(ctx context.Context, userID, templateID string, params map[string]string, priority Priority) ([]*Notification, []error) {
	channels := []Channel{ChannelEmail, ChannelSMS, ChannelPush, ChannelWebSocket}
	var notifications []*Notification
	var errs []error
	for _, ch := range channels {
		n, err := d.Dispatch(ctx, userID, ch, templateID, params, priority)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		notifications = append(notifications, n)
	}
	return notifications, errs
}

// MockProvider is a test-friendly channel provider.
type MockProvider struct {
	ch       Channel
	mu       sync.Mutex
	sent     []*Notification
	failNext bool
}

// NewMockProvider creates a mock channel provider.
func NewMockProvider(ch Channel) *MockProvider {
	return &MockProvider{ch: ch}
}

// Send implements ChannelProvider.
func (mp *MockProvider) Send(_ context.Context, n *Notification) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	if mp.failNext {
		mp.failNext = false
		return errors.New("mock send failure")
	}
	mp.sent = append(mp.sent, n)
	return nil
}

// Channel implements ChannelProvider.
func (mp *MockProvider) Channel() Channel { return mp.ch }

// SetFailNext causes the next Send call to fail.
func (mp *MockProvider) SetFailNext() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.failNext = true
}

// SentCount returns the number of successfully sent notifications.
func (mp *MockProvider) SentCount() int {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	return len(mp.sent)
}
