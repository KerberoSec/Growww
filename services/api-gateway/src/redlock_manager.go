package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type DistributedLock struct {
	Resource   string
	Token      string
	ExpiresAt  time.Time
	mu         sync.Mutex
}

type RedlockManager struct {
	mu    sync.Mutex
	locks map[string]*DistributedLock
}

func NewRedlockManager() *RedlockManager {
	return &RedlockManager{locks: make(map[string]*DistributedLock)}
}

func (m *RedlockManager) Acquire(resource string, ttl time.Duration) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UTC()
	if existing, ok := m.locks[resource]; ok {
		if now.Before(existing.ExpiresAt) {
			return "", fmt.Errorf("lock for %s already held", resource)
		}
	}

	b := make([]byte, 16)
	rand.Read(b)
	token := hex.EncodeToString(b)

	m.locks[resource] = &DistributedLock{
		Resource:  resource,
		Token:     token,
		ExpiresAt: now.Add(ttl),
	}
	return token, nil
}

func (m *RedlockManager) Release(resource, token string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	lock, ok := m.locks[resource]
	if !ok {
		return false
	}
	if lock.Token != token {
		return false
	}
	delete(m.locks, resource)
	return true
}
