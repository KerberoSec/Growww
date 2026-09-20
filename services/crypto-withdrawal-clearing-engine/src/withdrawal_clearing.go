package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type WithdrawalStatus string

const (
	StatusPendingReview WithdrawalStatus = "PENDING_REVIEW"
	StatusTimelocked    WithdrawalStatus = "TIMELOCKED"
	StatusApproved      WithdrawalStatus = "APPROVED"
	StatusBroadcasted   WithdrawalStatus = "BROADCASTED"
	StatusRejected      WithdrawalStatus = "REJECTED"
)

type BlockchainNetwork string

const (
	NetworkBitcoin  BlockchainNetwork = "BITCOIN"
	NetworkEthereum BlockchainNetwork = "ETHEREUM"
	NetworkTron     BlockchainNetwork = "TRON"
	NetworkSolana   BlockchainNetwork = "SOLANA"
)

var (
	ErrAddressNotWhitelisted = errors.New("destination address is not on verified beneficiary whitelist")
	ErrTimelockActive        = errors.New("24-hour security cool-off period is currently active for this beneficiary")
	ErrInvalidAddressFormat  = errors.New("destination address format invalid for selected blockchain network")
	ErrBeneficiaryNotActive  = errors.New("beneficiary address has not completed 2FA or email confirmation")
	ErrInvalidConfirmation   = errors.New("invalid 2FA code or confirmation token")
	ErrBeneficiaryNotFound   = errors.New("beneficiary entry not found")
)

var (
	ethAddressRegex = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	btcLegacyRegex  = regexp.MustCompile(`^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$`)
	btcBech32Regex  = regexp.MustCompile(`^bc1[a-zA-HJ-NP-Z0-9]{25,62}$`)
	tronRegex       = regexp.MustCompile(`^T[a-zA-HJ-NP-Z0-9]{33}$`)
	solanaRegex     = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{32,44}$`)
)

type WhitelistEntry struct {
	BeneficiaryID     string            `json:"beneficiary_id"`
	UserID            string            `json:"user_id"`
	Address           string            `json:"address"`
	Network           BlockchainNetwork `json:"network"`
	Asset             string            `json:"asset"`
	Label             string            `json:"label"`
	MemoOrTag         string            `json:"memo_or_tag,omitempty"`
	AddedAt           time.Time         `json:"added_at"`
	ConfirmedAt       *time.Time        `json:"confirmed_at,omitempty"`
	CooloffExpiry     time.Time         `json:"cooloff_expiry"`
	Is2FAConfirmed    bool              `json:"is_2fa_confirmed"`
	IsEmailConfirmed  bool              `json:"is_email_confirmed"`
	ConfirmationToken string            `json:"-"`
}

type UserSecuritySettings struct {
	UserID            string `json:"user_id"`
	WhitelistOnlyMode bool   `json:"whitelist_only_mode"`
}

type WithdrawalRequest struct {
	RequestID       string            `json:"request_id"`
	UserID          string            `json:"user_id"`
	Asset           string            `json:"asset"`
	Network         BlockchainNetwork `json:"network"`
	DestinationAddr string            `json:"destination_addr"`
	AmountE8        uint64            `json:"amount_e8"`
	Status          WithdrawalStatus  `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	ReleaseAt       time.Time         `json:"release_at"`
}

type WithdrawalClearingEngine struct {
	mu             sync.RWMutex
	whitelist      map[string]*WhitelistEntry // "user:addr" -> entry
	userSettings   map[string]*UserSecuritySettings
	requests       map[string]*WithdrawalRequest
	cooloffPeriod  time.Duration
}

func NewWithdrawalClearingEngine(cooloff time.Duration) *WithdrawalClearingEngine {
	if cooloff <= 0 {
		cooloff = 24 * time.Hour
	}
	return &WithdrawalClearingEngine{
		whitelist:     make(map[string]*WhitelistEntry),
		userSettings:  make(map[string]*UserSecuritySettings),
		requests:      make(map[string]*WithdrawalRequest),
		cooloffPeriod: cooloff,
	}
}

// ValidateAddressFormat checks cryptographic format of destination address
func ValidateAddressFormat(network BlockchainNetwork, addr string) error {
	addr = strings.TrimSpace(addr)
	switch network {
	case NetworkEthereum:
		if !ethAddressRegex.MatchString(addr) {
			return fmt.Errorf("%w: invalid Ethereum/EVM hex format", ErrInvalidAddressFormat)
		}
	case NetworkBitcoin:
		if !btcLegacyRegex.MatchString(addr) && !btcBech32Regex.MatchString(addr) {
			return fmt.Errorf("%w: invalid Bitcoin address (must be legacy, P2SH, or native Bech32 bc1)", ErrInvalidAddressFormat)
		}
	case NetworkTron:
		if !tronRegex.MatchString(addr) {
			return fmt.Errorf("%w: invalid Tron base58 address starting with T", ErrInvalidAddressFormat)
		}
	case NetworkSolana:
		if !solanaRegex.MatchString(addr) {
			return fmt.Errorf("%w: invalid Solana base58 public key", ErrInvalidAddressFormat)
		}
	default:
		return fmt.Errorf("%w: unsupported blockchain network %s", ErrInvalidAddressFormat, network)
	}
	return nil
}

// SetWhitelistOnlyMode toggles the strict whitelist-only withdrawal setting for a user
func (e *WithdrawalClearingEngine) SetWhitelistOnlyMode(userID string, enabled bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	settings, exists := e.userSettings[userID]
	if !exists {
		settings = &UserSecuritySettings{UserID: userID}
		e.userSettings[userID] = settings
	}
	settings.WhitelistOnlyMode = enabled
}

// AddBeneficiary registers a new external withdrawal address with pending confirmation
func (e *WithdrawalClearingEngine) AddBeneficiary(
	userID, addr string,
	network BlockchainNetwork,
	asset, label, memo string,
	expected2FACode string,
) (*WhitelistEntry, error) {
	if err := ValidateAddressFormat(network, addr); err != nil {
		return nil, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	key := fmt.Sprintf("%s:%s", userID, addr)

	tokenHash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", userID, addr, now.UnixNano())))
	confToken := hex.EncodeToString(tokenHash[:16])

	entry := &WhitelistEntry{
		BeneficiaryID:     fmt.Sprintf("BEN-%06d", time.Now().Unix()%1000000),
		UserID:            userID,
		Address:           addr,
		Network:           network,
		Asset:             asset,
		Label:             label,
		MemoOrTag:         memo,
		AddedAt:           now,
		CooloffExpiry:     now.Add(e.cooloffPeriod),
		Is2FAConfirmed:    false,
		IsEmailConfirmed:  false,
		ConfirmationToken: confToken,
	}

	// If valid 6-digit 2FA provided at enrollment
	if len(expected2FACode) == 6 {
		entry.Is2FAConfirmed = true
	}

	e.whitelist[key] = entry
	return entry, nil
}

// ConfirmBeneficiaryEmail validates email confirmation link token
func (e *WithdrawalClearingEngine) ConfirmBeneficiaryEmail(userID, addr, token string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, addr)
	entry, exists := e.whitelist[key]
	if !exists {
		return ErrBeneficiaryNotFound
	}

	if entry.ConfirmationToken != token {
		return ErrInvalidConfirmation
	}

	entry.IsEmailConfirmed = true
	now := time.Now().UTC()
	entry.ConfirmedAt = &now
	// Cooloff period begins upon full confirmation
	entry.CooloffExpiry = now.Add(e.cooloffPeriod)

	return nil
}

// RequestWithdrawal verifies whitelist status and enforces security cooling periods
func (e *WithdrawalClearingEngine) RequestWithdrawal(
	reqID, userID, asset string,
	network BlockchainNetwork,
	destAddr string,
	amountE8 uint64,
	now time.Time,
) (*WithdrawalRequest, error) {
	if err := ValidateAddressFormat(network, destAddr); err != nil {
		return nil, err
	}
	if amountE8 == 0 {
		return nil, errors.New("withdrawal amount must be > 0")
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, destAddr)
	entry, exists := e.whitelist[key]

	settings, hasSettings := e.userSettings[userID]
	whitelistOnly := hasSettings && settings.WhitelistOnlyMode

	if whitelistOnly && !exists {
		return nil, ErrAddressNotWhitelisted
	}

	if exists {
		if !entry.Is2FAConfirmed || !entry.IsEmailConfirmed {
			return nil, ErrBeneficiaryNotActive
		}
		if now.Before(entry.CooloffExpiry) {
			return nil, fmt.Errorf("%w: locked until %s", ErrTimelockActive, entry.CooloffExpiry.Format(time.RFC3339))
		}
	}

	req := &WithdrawalRequest{
		RequestID:       reqID,
		UserID:          userID,
		Asset:           asset,
		Network:         network,
		DestinationAddr: destAddr,
		AmountE8:        amountE8,
		Status:          StatusPendingReview,
		CreatedAt:       now,
		ReleaseAt:       now.Add(15 * time.Minute),
	}
	e.requests[reqID] = req
	return req, nil
}

// GetBeneficiaries lists all configured beneficiaries for a user
func (e *WithdrawalClearingEngine) GetBeneficiaries(userID string) []*WhitelistEntry {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*WhitelistEntry
	for _, entry := range e.whitelist {
		if entry.UserID == userID {
			result = append(result, entry)
		}
	}
	return result
}
