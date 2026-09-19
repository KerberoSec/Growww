package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type BeneficiaryType string

const (
	BeneficiaryBank   BeneficiaryType = "BANK_ACCOUNT"
	BeneficiaryCrypto BeneficiaryType = "CRYPTO_ADDRESS"
)

type BeneficiaryItem struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	Type            BeneficiaryType `json:"type"`
	Identifier      string          `json:"identifier"` // IFSC+Account or Blockchain Address
	Network         string          `json:"network"`
	Label           string          `json:"label"`
	CoolingExpiresAt time.Time      `json:"cooling_expires_at"`
	IsActive        bool            `json:"is_active"`
}

type BeneficiaryManager struct {
	mu            sync.RWMutex
	beneficiaries map[string]*BeneficiaryItem // ID -> item
}

func NewBeneficiaryManager() *BeneficiaryManager {
	return &BeneficiaryManager{
		beneficiaries: make(map[string]*BeneficiaryItem),
	}
}

// AddBeneficiary adds a new destination with mandatory 12-hour anti-drain cooling period
func (m *BeneficiaryManager) AddBeneficiary(id, userID string, bType BeneficiaryType, identifier, network, label string) *BeneficiaryItem {
	m.mu.Lock()
	defer m.mu.Unlock()

	item := &BeneficiaryItem{
		ID:               id,
		UserID:           userID,
		Type:             bType,
		Identifier:       identifier,
		Network:          network,
		Label:            label,
		CoolingExpiresAt: time.Now().UTC().Add(12 * time.Hour),
		IsActive:         false,
	}
	m.beneficiaries[id] = item
	fmt.Printf("[Beneficiary Book] Enrolled %s (%s) for user %s with 12h cooling window\n", identifier, label, userID)
	return item
}

// ValidateForWithdrawal checks that the recipient address has passed cooling
func (m *BeneficiaryManager) ValidateForWithdrawal(id, userID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.beneficiaries[id]
	if !exists || item.UserID != userID {
		return errors.New("beneficiary not found")
	}

	if time.Now().UTC().Before(item.CoolingExpiresAt) {
		return fmt.Errorf("beneficiary in cooling security state until %v", item.CoolingExpiresAt)
	}

	return nil
}
