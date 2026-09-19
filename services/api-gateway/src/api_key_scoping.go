package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net"
	"sync"
	"time"
)

type APIPermission uint8

const (
	PermRead     APIPermission = 1 << 0 // 1
	PermTrade    APIPermission = 1 << 1 // 2
	PermWithdraw APIPermission = 1 << 2 // 4
)

type APIKeyMetadata struct {
	KeyHash       string
	UserID        string
	Permissions   APIPermission
	IPWhitelist   []string // CIDR or exact IP
	RateLimitTier string
	ExpiresAt     time.Time
	Active        bool
}

type APIKeyAuthorizer struct {
	mu   sync.RWMutex
	keys map[string]APIKeyMetadata // keyHash -> metadata
}

func NewAPIKeyAuthorizer() *APIKeyAuthorizer {
	return &APIKeyAuthorizer{
		keys: make(map[string]APIKeyMetadata),
	}
}

// AuthorizeRequest verifies key validity, IP whitelisting, and requested permission
func (a *APIKeyAuthorizer) AuthorizeRequest(rawKey, clientIP string, requiredPerm APIPermission) error {
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	a.mu.RLock()
	meta, exists := a.keys[keyHash]
	a.mu.RUnlock()

	if !exists || !meta.Active {
		return errors.New("invalid or revoked API key")
	}

	if time.Now().UTC().After(meta.ExpiresAt) {
		return errors.New("API key expired")
	}

	// Verify Permission
	if (meta.Permissions & requiredPerm) == 0 {
		return errors.New("forbidden: API key lacks required permission")
	}

	// Verify IP Whitelist
	if len(meta.IPWhitelist) > 0 {
		ip := net.ParseIP(clientIP)
		if ip == nil {
			return errors.New("invalid client IP format")
		}

		matched := false
		for _, entry := range meta.IPWhitelist {
			if _, cidrNet, err := net.ParseCIDR(entry); err == nil {
				if cidrNet.Contains(ip) {
					matched = true
					break
				}
			} else if entry == clientIP {
				matched = true
				break
			}
		}

		if !matched {
			return errors.New("access denied: client IP not whitelisted for this API key")
		}
	}

	return nil
}
