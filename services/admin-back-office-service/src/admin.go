// Package adminbackoffice provides the Admin Back-Office Service for NBSE.
// It handles user management, KYC approval queues, trade monitoring,
// and system configuration for the exchange admin dashboard.
package adminbackoffice

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// KYCStatus represents the KYC verification state of a user.
type KYCStatus string

const (
	KYCPending  KYCStatus = "PENDING"
	KYCApproved KYCStatus = "APPROVED"
	KYCRejected KYCStatus = "REJECTED"
	KYCSuspended KYCStatus = "SUSPENDED"
)

// UserRole defines access levels within the admin system.
type UserRole string

const (
	RoleSuperAdmin UserRole = "SUPER_ADMIN"
	RoleCompliance UserRole = "COMPLIANCE"
	RoleSupport    UserRole = "SUPPORT"
	RoleViewer     UserRole = "VIEWER"
)

// User represents a platform user managed by admins.
type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full_name"`
	KYC           KYCStatus `json:"kyc_status"`
	Role          UserRole  `json:"role"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PANNumber     string    `json:"pan_number,omitempty"`
	AadhaarHash   string    `json:"aadhaar_hash,omitempty"`
	TradingVolume float64   `json:"trading_volume_inr"`
}

// KYCRequest represents a pending KYC approval item.
type KYCRequest struct {
	UserID      string    `json:"user_id"`
	DocumentURL string    `json:"document_url"`
	SubmittedAt time.Time `json:"submitted_at"`
	ReviewedBy  string    `json:"reviewed_by,omitempty"`
	ReviewedAt  time.Time `json:"reviewed_at,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// TradeAlert represents a flagged trade for monitoring.
type TradeAlert struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Symbol     string    `json:"symbol"`
	Side       string    `json:"side"`
	Quantity   float64   `json:"quantity"`
	Price      float64   `json:"price"`
	AlertType  string    `json:"alert_type"` // WASH_TRADE, SPOOFING, LARGE_ORDER, RAPID_FIRE
	Severity   string    `json:"severity"`   // LOW, MEDIUM, HIGH, CRITICAL
	DetectedAt time.Time `json:"detected_at"`
	Resolved   bool      `json:"resolved"`
}

// SystemConfig holds dynamic exchange configuration parameters.
type SystemConfig struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedBy string    `json:"updated_by"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AdminService provides the core admin back-office operations.
type AdminService struct {
	mu           sync.RWMutex
	users        map[string]*User
	kycQueue     []*KYCRequest
	tradeAlerts  []*TradeAlert
	systemConfig map[string]*SystemConfig
}

// NewAdminService creates a new AdminService instance.
func NewAdminService() *AdminService {
	return &AdminService{
		users:        make(map[string]*User),
		kycQueue:     make([]*KYCRequest, 0),
		tradeAlerts:  make([]*TradeAlert, 0),
		systemConfig: make(map[string]*SystemConfig),
	}
}

// ──────────────────────────────────────────────
// User Management
// ──────────────────────────────────────────────

// CreateUser registers a new user in the system.
func (s *AdminService) CreateUser(id, email, fullName string, role UserRole) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id == "" || email == "" {
		return nil, errors.New("user id and email are required")
	}
	if _, exists := s.users[id]; exists {
		return nil, fmt.Errorf("user %s already exists", id)
	}

	now := time.Now().UTC()
	u := &User{
		ID:        id,
		Email:     email,
		FullName:  fullName,
		KYC:       KYCPending,
		Role:      role,
		Enabled:   true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.users[id] = u
	return u, nil
}

// GetUser retrieves a user by ID.
func (s *AdminService) GetUser(id string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	if !ok {
		return nil, fmt.Errorf("user %s not found", id)
	}
	return u, nil
}

// DisableUser disables a user account (soft-ban).
func (s *AdminService) DisableUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return fmt.Errorf("user %s not found", id)
	}
	u.Enabled = false
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// EnableUser re-enables a disabled user account.
func (s *AdminService) EnableUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return fmt.Errorf("user %s not found", id)
	}
	u.Enabled = true
	u.UpdatedAt = time.Now().UTC()
	return nil
}

// ListUsers returns all users. Supports optional KYC status filter.
func (s *AdminService) ListUsers(kycFilter *KYCStatus) []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		if kycFilter != nil && u.KYC != *kycFilter {
			continue
		}
		result = append(result, u)
	}
	return result
}

// ──────────────────────────────────────────────
// KYC Approval Queue
// ──────────────────────────────────────────────

// SubmitKYC adds a KYC request to the approval queue.
func (s *AdminService) SubmitKYC(userID, documentURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user %s not found", userID)
	}
	if u.KYC == KYCApproved {
		return errors.New("user KYC already approved")
	}

	req := &KYCRequest{
		UserID:      userID,
		DocumentURL: documentURL,
		SubmittedAt: time.Now().UTC(),
	}
	s.kycQueue = append(s.kycQueue, req)
	return nil
}

// ApproveKYC approves a pending KYC request.
func (s *AdminService) ApproveKYC(userID, reviewerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user %s not found", userID)
	}
	if u.KYC != KYCPending {
		return fmt.Errorf("user KYC is %s, not PENDING", u.KYC)
	}

	u.KYC = KYCApproved
	u.UpdatedAt = time.Now().UTC()

	// Mark queue entry as reviewed
	for _, req := range s.kycQueue {
		if req.UserID == userID && req.ReviewedBy == "" {
			req.ReviewedBy = reviewerID
			req.ReviewedAt = time.Now().UTC()
			break
		}
	}
	return nil
}

// RejectKYC rejects a pending KYC request with notes.
func (s *AdminService) RejectKYC(userID, reviewerID, notes string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[userID]
	if !ok {
		return fmt.Errorf("user %s not found", userID)
	}
	if u.KYC != KYCPending {
		return fmt.Errorf("user KYC is %s, not PENDING", u.KYC)
	}

	u.KYC = KYCRejected
	u.UpdatedAt = time.Now().UTC()

	for _, req := range s.kycQueue {
		if req.UserID == userID && req.ReviewedBy == "" {
			req.ReviewedBy = reviewerID
			req.ReviewedAt = time.Now().UTC()
			req.Notes = notes
			break
		}
	}
	return nil
}

// PendingKYCCount returns the number of unreviewed KYC requests.
func (s *AdminService) PendingKYCCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, req := range s.kycQueue {
		if req.ReviewedBy == "" {
			count++
		}
	}
	return count
}

// ──────────────────────────────────────────────
// Trade Monitoring
// ──────────────────────────────────────────────

// RaiseTradeAlert creates a new trade surveillance alert.
func (s *AdminService) RaiseTradeAlert(alert *TradeAlert) error {
	if alert.ID == "" || alert.UserID == "" {
		return errors.New("alert ID and UserID are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	alert.DetectedAt = time.Now().UTC()
	alert.Resolved = false
	s.tradeAlerts = append(s.tradeAlerts, alert)
	return nil
}

// ResolveTradeAlert marks an alert as resolved.
func (s *AdminService) ResolveTradeAlert(alertID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.tradeAlerts {
		if a.ID == alertID {
			a.Resolved = true
			return nil
		}
	}
	return fmt.Errorf("alert %s not found", alertID)
}

// UnresolvedAlerts returns all unresolved trade alerts.
func (s *AdminService) UnresolvedAlerts() []*TradeAlert {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*TradeAlert, 0)
	for _, a := range s.tradeAlerts {
		if !a.Resolved {
			result = append(result, a)
		}
	}
	return result
}

// ──────────────────────────────────────────────
// System Configuration
// ──────────────────────────────────────────────

// SetConfig upserts a system configuration key-value pair.
func (s *AdminService) SetConfig(key, value, updatedBy string) error {
	if key == "" {
		return errors.New("config key is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.systemConfig[key] = &SystemConfig{
		Key:       key,
		Value:     value,
		UpdatedBy: updatedBy,
		UpdatedAt: time.Now().UTC(),
	}
	return nil
}

// GetConfig retrieves a system configuration value by key.
func (s *AdminService) GetConfig(key string) (*SystemConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cfg, ok := s.systemConfig[key]
	if !ok {
		return nil, fmt.Errorf("config key %q not found", key)
	}
	return cfg, nil
}

// AllConfigs returns all system configuration entries.
func (s *AdminService) AllConfigs() []*SystemConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*SystemConfig, 0, len(s.systemConfig))
	for _, c := range s.systemConfig {
		result = append(result, c)
	}
	return result
}
