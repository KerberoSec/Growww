package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type ConsentCategory string

const (
	ConsentEssentialService       ConsentCategory = "ESSENTIAL_TRADING_SERVICE"
	ConsentMarketingAnalytics     ConsentCategory = "MARKETING_ANALYTICS"
	ConsentPromotionalSMS         ConsentCategory = "PROMOTIONAL_SMS"
	ConsentThirdPartyOffers       ConsentCategory = "THIRD_PARTY_OFFERS"
)

type ConsentAction string

const (
	ActionGrant  ConsentAction = "GRANT"
	ActionRevoke ConsentAction = "REVOKE"
)

var (
	ErrEssentialConsentRevocation = errors.New("mandatory essential service consent cannot be revoked while account is active")
	ErrConsentNotFound            = errors.New("consent record not found")
)

type ConsentLogEntry struct {
	EntryID      string          `json:"entry_id"`
	UserID       string          `json:"user_id"`
	Category     ConsentCategory `json:"category"`
	Action       ConsentAction   `json:"action"`
	IPAddress    string          `json:"ip_address"`
	UserAgent    string          `json:"user_agent"`
	Reason       string          `json:"reason,omitempty"`
	Timestamp    time.Time       `json:"timestamp"`
	PreviousHash string          `json:"previous_hash"`
	RecordHash   string          `json:"record_hash"`
}

type UserConsentState struct {
	UserID                   string          `json:"user_id"`
	EssentialTradingService  bool            `json:"essential_trading_service"`
	MarketingAnalytics       bool            `json:"marketing_analytics"`
	PromotionalSMS           bool            `json:"promotional_sms"`
	ThirdPartyOffers         bool            `json:"third_party_offers"`
	LastUpdated              time.Time       `json:"last_updated"`
}

type PortabilityProfile struct {
	UserID        string             `json:"user_id"`
	Email         string             `json:"email"`
	PANStatus     string             `json:"pan_status"`
	AadhaarStatus string             `json:"aadhaar_status"`
	Consents      *UserConsentState  `json:"active_consents"`
	ExportedAt    time.Time          `json:"exported_at"`
	ChecksumSHA256 string            `json:"checksum_sha256"`
}

type DPDPConsentManager struct {
	mu           sync.RWMutex
	states       map[string]*UserConsentState // userID -> state
	auditTrail   map[string][]ConsentLogEntry // userID -> entries
	lastHash     map[string]string            // userID -> last record hash
}

func NewDPDPConsentManager() *DPDPConsentManager {
	return &DPDPConsentManager{
		states:     make(map[string]*UserConsentState),
		auditTrail: make(map[string][]ConsentLogEntry),
		lastHash:   make(map[string]string),
	}
}

// GrantConsent updates user consent state and logs immutable cryptographic entry
func (m *DPDPConsentManager) GrantConsent(userID string, category ConsentCategory, ipAddr, userAgent string) (*ConsentLogEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[userID]
	if !exists {
		state = &UserConsentState{
			UserID: userID,
		}
		m.states[userID] = state
	}

	switch category {
	case ConsentEssentialService:
		state.EssentialTradingService = true
	case ConsentMarketingAnalytics:
		state.MarketingAnalytics = true
	case ConsentPromotionalSMS:
		state.PromotionalSMS = true
	case ConsentThirdPartyOffers:
		state.ThirdPartyOffers = true
	default:
		return nil, errors.New("unknown consent category")
	}

	state.LastUpdated = time.Now().UTC()
	return m.appendLogLocked(userID, category, ActionGrant, ipAddr, userAgent, "")
}

// RevokeConsent updates consent state and logs revocation
func (m *DPDPConsentManager) RevokeConsent(userID string, category ConsentCategory, ipAddr, userAgent, reason string) (*ConsentLogEntry, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.states[userID]
	if !exists {
		return nil, ErrConsentNotFound
	}

	// Essential consent cannot be unilaterally revoked while account remains open
	if category == ConsentEssentialService {
		return nil, ErrEssentialConsentRevocation
	}

	switch category {
	case ConsentMarketingAnalytics:
		state.MarketingAnalytics = false
	case ConsentPromotionalSMS:
		state.PromotionalSMS = false
	case ConsentThirdPartyOffers:
		state.ThirdPartyOffers = false
	default:
		return nil, errors.New("unknown consent category")
	}

	state.LastUpdated = time.Now().UTC()
	return m.appendLogLocked(userID, category, ActionRevoke, ipAddr, userAgent, reason)
}

func (m *DPDPConsentManager) appendLogLocked(
	userID string,
	category ConsentCategory,
	action ConsentAction,
	ipAddr, userAgent, reason string,
) (*ConsentLogEntry, error) {
	now := time.Now().UTC()
	prevHash := m.lastHash[userID]
	if prevHash == "" {
		prevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	}

	entryID := fmt.Sprintf("DPDP-%s-%d", userID, now.UnixNano())
	payload := fmt.Sprintf("%s:%s:%s:%s:%s:%s:%d:%s",
		entryID, userID, category, action, ipAddr, userAgent, now.Unix(), prevHash)
	hash := sha256.Sum256([]byte(payload))
	recordHash := hex.EncodeToString(hash[:])

	entry := ConsentLogEntry{
		EntryID:      entryID,
		UserID:       userID,
		Category:     category,
		Action:       action,
		IPAddress:    ipAddr,
		UserAgent:    userAgent,
		Reason:       reason,
		Timestamp:    now,
		PreviousHash: prevHash,
		RecordHash:   recordHash,
	}

	m.auditTrail[userID] = append(m.auditTrail[userID], entry)
	m.lastHash[userID] = recordHash

	return &entry, nil
}

// GetUserConsentState returns current active opt-in status
func (m *DPDPConsentManager) GetUserConsentState(userID string) (*UserConsentState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[userID]
	if !exists {
		return nil, ErrConsentNotFound
	}
	return state, nil
}

// GetConsentAuditTrail returns chronological immutable audit entries
func (m *DPDPConsentManager) GetConsentAuditTrail(userID string) []ConsentLogEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entries := m.auditTrail[userID]
	result := make([]ConsentLogEntry, len(entries))
	copy(result, entries)
	return result
}

// ExportDataPortabilityArchive compiles machine-readable JSON data archive with SHA256 integrity checksum
func (m *DPDPConsentManager) ExportDataPortabilityArchive(userID, email, panStatus, aadhaarStatus string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state := m.states[userID]
	profile := PortabilityProfile{
		UserID:        userID,
		Email:         email,
		PANStatus:     panStatus,
		AadhaarStatus: aadhaarStatus,
		Consents:      state,
		ExportedAt:    time.Now().UTC(),
	}

	raw, err := json.Marshal(profile)
	if err != nil {
		return nil, err
	}

	checksum := sha256.Sum256(raw)
	profile.ChecksumSHA256 = hex.EncodeToString(checksum[:])

	return json.MarshalIndent(profile, "", "  ")
}
