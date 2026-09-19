package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// CBDCTokenStatus represents the lifecycle state of a CBDC token
type CBDCTokenStatus string

const (
	TokenIssued   CBDCTokenStatus = "ISSUED"
	TokenActive   CBDCTokenStatus = "ACTIVE"
	TokenRedeemed CBDCTokenStatus = "REDEEMED"
	TokenExpired  CBDCTokenStatus = "EXPIRED"
)

// CBDCToken represents a digital rupee token instance
type CBDCToken struct {
	TokenID       string          `json:"token_id"`
	Type          CBDCType        `json:"type"`
	DenomPaise    uint64          `json:"denom_paise"`
	OwnerVPA      string          `json:"owner_vpa"`
	Status        CBDCTokenStatus `json:"status"`
	IssuedAt      time.Time       `json:"issued_at"`
	LastTransfer  time.Time       `json:"last_transfer"`
	RedemptionRef string          `json:"redemption_ref,omitempty"`
}

// CBDCTokenManager manages the issuance, transfer, and redemption lifecycle
type CBDCTokenManager struct {
	mu       sync.RWMutex
	tokens   map[string]*CBDCToken
	adapter  *CBDCBridgeAdapter
	counter  uint64
}

// NewCBDCTokenManager creates a new token lifecycle manager
func NewCBDCTokenManager(adapter *CBDCBridgeAdapter) *CBDCTokenManager {
	return &CBDCTokenManager{
		tokens:  make(map[string]*CBDCToken),
		adapter: adapter,
	}
}

// IssueToken creates a new digital rupee token
func (m *CBDCTokenManager) IssueToken(tokenType CBDCType, denomPaise uint64, ownerVPA string) (*CBDCToken, error) {
	if denomPaise == 0 {
		return nil, errors.New("token denomination must be > 0")
	}
	if ownerVPA == "" {
		return nil, errors.New("owner VPA required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.counter++
	tokenID := fmt.Sprintf("EINR-%s-%06d", tokenType, m.counter)

	token := &CBDCToken{
		TokenID:    tokenID,
		Type:       tokenType,
		DenomPaise: denomPaise,
		OwnerVPA:   ownerVPA,
		Status:     TokenIssued,
		IssuedAt:   time.Now().UTC(),
	}
	m.tokens[tokenID] = token
	return token, nil
}

// TransferToken transfers ownership of a digital rupee token between VPAs
func (m *CBDCTokenManager) TransferToken(tokenID, fromVPA, toVPA string) (*CBDCTransferReceipt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, exists := m.tokens[tokenID]
	if !exists {
		return nil, errors.New("token not found")
	}
	if token.OwnerVPA != fromVPA {
		return nil, errors.New("sender does not own this token")
	}
	if token.Status == TokenRedeemed || token.Status == TokenExpired {
		return nil, fmt.Errorf("token in %s state cannot be transferred", token.Status)
	}

	// Execute transfer via RBI bridge adapter
	receipt, err := m.adapter.SettleCBDCTransfer(CBDCTransferRequest{
		TransferID:  fmt.Sprintf("TXF-%s", tokenID),
		PayerVPA:    fromVPA,
		PayeeVPA:    toVPA,
		AmountPaise: token.DenomPaise,
		Type:        token.Type,
	})
	if err != nil {
		return nil, fmt.Errorf("CBDC transfer failed: %w", err)
	}

	token.OwnerVPA = toVPA
	token.Status = TokenActive
	token.LastTransfer = time.Now().UTC()

	return receipt, nil
}

// RedeemToken redeems a digital rupee token back to fiat INR
func (m *CBDCTokenManager) RedeemToken(tokenID, ownerVPA string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	token, exists := m.tokens[tokenID]
	if !exists {
		return errors.New("token not found")
	}
	if token.OwnerVPA != ownerVPA {
		return errors.New("only token owner can redeem")
	}
	if token.Status == TokenRedeemed {
		return errors.New("token already redeemed")
	}

	token.Status = TokenRedeemed
	token.RedemptionRef = fmt.Sprintf("REDEEM-%s-%d", tokenID, time.Now().UnixNano())
	return nil
}

// GetToken retrieves a token by ID
func (m *CBDCTokenManager) GetToken(tokenID string) (*CBDCToken, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.tokens[tokenID]
	return t, ok
}

// GetTokensByOwner returns all tokens owned by a VPA
func (m *CBDCTokenManager) GetTokensByOwner(ownerVPA string) []*CBDCToken {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*CBDCToken
	for _, t := range m.tokens {
		if t.OwnerVPA == ownerVPA {
			result = append(result, t)
		}
	}
	return result
}
