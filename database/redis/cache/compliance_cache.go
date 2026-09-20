package cache

import (
	"context"
	"strings"
	"sync"
	"time"
)

// ComplianceCache manages in-memory cached on-chain KYC whitelists and sessions.
type ComplianceCache struct {
	mu           sync.RWMutex
	whitelists   map[string]ComplianceStatus // address (lowercase) -> status
	sessions     map[string]*SessionRecord   // session_key -> record
	revokedJTIs  map[string]time.Time        // jti -> revocation_time
}

func NewComplianceCache() *ComplianceCache {
	return &ComplianceCache{
		whitelists:  make(map[string]ComplianceStatus),
		sessions:    make(map[string]*SessionRecord),
		revokedJTIs: make(map[string]time.Time),
	}
}

// SetWhitelist sets KYC status for a blockchain address.
func (c *ComplianceCache) SetWhitelist(address string, status ComplianceStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.whitelists[strings.ToLower(address)] = status
}

// CheckWhitelist validates if an address is active.
func (c *ComplianceCache) CheckWhitelist(ctx context.Context, address string) (ComplianceStatus, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	status, exists := c.whitelists[strings.ToLower(address)]
	return status, exists
}

// StoreSession caches an active session.
func (c *ComplianceCache) StoreSession(session *SessionRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := session.UserID + ":" + session.DeviceID
	c.sessions[key] = session
}

// RevokeJWT marks a JWT ID as revoked in the blacklist.
func (c *ComplianceCache) RevokeJWT(jti string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.revokedJTIs[jti] = time.Now().UTC()
}

// IsJWTRevoked checks if a JWT is blacklisted.
func (c *ComplianceCache) IsJWTRevoked(jti string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, revoked := c.revokedJTIs[jti]
	return revoked
}
