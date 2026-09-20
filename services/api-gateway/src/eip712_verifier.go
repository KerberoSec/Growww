package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// EIP712Domain represents standard EIP-712 domain separator
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           uint64
	VerifyingContract string
}

// SpotOrderData represents structured order payload
type SpotOrderData struct {
	Trader   string
	ISIN     string
	Side     string // BUY or SELL
	Quantity uint64
	Price    uint64
	Nonce    uint64
	Deadline int64
}

// SessionKeyDelegationData represents delegated trading key
type SessionKeyDelegationData struct {
	MasterTrader      string
	SessionKey        string
	ValidUntil        int64
	MaxNotionalLimit  uint64
	Nonce             uint64
	ApprovedSignature string
}

// EIP712Manager verifies typed signatures and manages replay nonces
type EIP712Manager struct {
	mu           sync.RWMutex
	domain       EIP712Domain
	domainHash   [32]byte
	accountNonce map[string]uint64                     // trader -> current expected nonce
	sessionKeys  map[string]*SessionKeyDelegationData // sessionKey -> delegation
}

// NewEIP712Manager initializes domain separator and state store
func NewEIP712Manager(domain EIP712Domain) *EIP712Manager {
	domainHash := hashDomain(domain)
	return &EIP712Manager{
		domain:       domain,
		domainHash:   domainHash,
		accountNonce: make(map[string]uint64),
		sessionKeys:  make(map[string]*SessionKeyDelegationData),
	}
}

func hashDomain(d EIP712Domain) [32]byte {
	// TypeHash: EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)
	typeString := "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	typeHash := sha256.Sum256([]byte(typeString))

	nameHash := sha256.Sum256([]byte(d.Name))
	versionHash := sha256.Sum256([]byte(d.Version))

	payload := fmt.Sprintf("%x%x%x%016x%s",
		typeHash,
		nameHash,
		versionHash,
		d.ChainID,
		strings.ToLower(d.VerifyingContract),
	)
	return sha256.Sum256([]byte(payload))
}

// HashSpotOrder calculates the 32-byte typed data digest
func (m *EIP712Manager) HashSpotOrder(order SpotOrderData) [32]byte {
	orderTypeString := "SpotOrder(address trader,string isin,string side,uint256 quantity,uint256 price,uint256 nonce,uint256 deadline)"
	orderTypeHash := sha256.Sum256([]byte(orderTypeString))

	isinHash := sha256.Sum256([]byte(order.ISIN))
	sideHash := sha256.Sum256([]byte(order.Side))

	structPayload := fmt.Sprintf("%x%s%x%x%016x%016x%016x%016x",
		orderTypeHash,
		strings.ToLower(order.Trader),
		isinHash,
		sideHash,
		order.Quantity,
		order.Price,
		order.Nonce,
		order.Deadline,
	)
	structHash := sha256.Sum256([]byte(structPayload))

	// Final EIP-712 digest: \x19\x01 || domainHash || structHash
	finalPayload := append([]byte("\x19\x01"), m.domainHash[:]...)
	finalPayload = append(finalPayload, structHash[:]...)
	return sha256.Sum256(finalPayload)
}

// VerifyAndConsumeOrder validates monotonic nonce, deadline, and execution permissions
func (m *EIP712Manager) VerifyAndConsumeOrder(order SpotOrderData, simulatedSigner string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. Check expiration
	if order.Deadline > 0 && time.Now().Unix() > order.Deadline {
		return errors.New("order signature expired")
	}

	trader := strings.ToLower(order.Trader)
	signer := strings.ToLower(simulatedSigner)

	// 2. Validate signer (must be trader or authorized session key)
	if signer != trader {
		delegation, exists := m.sessionKeys[signer]
		if !exists || strings.ToLower(delegation.MasterTrader) != trader {
			return errors.New("unauthorized signer for trader")
		}
		if time.Now().Unix() > delegation.ValidUntil {
			return errors.New("delegated session key expired")
		}
		notional := order.Quantity * order.Price
		if delegation.MaxNotionalLimit > 0 && notional > delegation.MaxNotionalLimit {
			return errors.New("order exceeds delegated session key notional limit")
		}
	}

	// 3. Monotonic Nonce enforcement (Replay Attack Defense)
	expectedNonce := m.accountNonce[trader]
	if order.Nonce != expectedNonce {
		return fmt.Errorf("invalid nonce: expected %d, got %d (replay detected)", expectedNonce, order.Nonce)
	}

	// Increment monotonic nonce
	m.accountNonce[trader] = expectedNonce + 1
	return nil
}

// RegisterSessionKey enables timed high-frequency order placement without repeated wallet popups
func (m *EIP712Manager) RegisterSessionKey(delegation SessionKeyDelegationData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if delegation.ValidUntil <= time.Now().Unix() {
		return errors.New("cannot register already-expired session key")
	}

	m.sessionKeys[strings.ToLower(delegation.SessionKey)] = &delegation
	return nil
}

// GetAccountNonce returns the current expected nonce for a trader
func (m *EIP712Manager) GetAccountNonce(trader string) uint64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.accountNonce[strings.ToLower(trader)]
}
