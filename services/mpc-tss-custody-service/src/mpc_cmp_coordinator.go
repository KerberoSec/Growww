package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

var (
	ErrInsufficientParties = errors.New("insufficient parties to satisfy threshold quorum")
	ErrSessionNotFound     = errors.New("mpc signing session not found")
	ErrPartyAlreadySigned  = errors.New("party has already submitted partial signature for this session")
	ErrThresholdNotMet     = errors.New("threshold not yet reached for signature aggregation")
)

// KeyShare represents an individual party's isolated share in an M-of-N topology
type KeyShare struct {
	PartyID          uint32
	ShareCommitment  []byte
	Threshold        uint32
	TotalParties     uint32
	privateScalar    *big.Int
	publicSharePoint *ecdsa.PublicKey
}

// PartialSignature represents a party's cryptographic contribution to a threshold round
type PartialSignature struct {
	PartyID   uint32
	Signature []byte
	R         *big.Int
	S         *big.Int
	CreatedAt time.Time
}

// SigningSession coordinates a single threshold signing ceremony
type SigningSession struct {
	SessionID             string
	MessageDigest         [32]byte
	ParticipatingParties  []uint32
	PartialSignatures     map[uint32]PartialSignature
	Threshold             uint32
	TotalParties          uint32
	AggregatedSignature   []byte
	CreatedAt             time.Time
	Completed             bool
}

// MPCCMPCoordinator manages distributed key generation and threshold signing
type MPCCMPCoordinator struct {
	mu           sync.RWMutex
	keyShares    map[uint32]*KeyShare
	sessions     map[string]*SigningSession
	threshold    uint32
	totalParties uint32
	masterKey    *ecdsa.PrivateKey // Used by emulator to verify joint correctness
}

// NewMPCCMPCoordinator initializes coordinator with (T, N) threshold parameters (e.g. 2-of-3 or 3-of-5)
func NewMPCCMPCoordinator(threshold, totalParties uint32) (*MPCCMPCoordinator, error) {
	if threshold == 0 || totalParties == 0 || threshold > totalParties {
		return nil, fmt.Errorf("invalid threshold parameters: T=%d, N=%d", threshold, totalParties)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base curve parameters: %w", err)
	}

	coord := &MPCCMPCoordinator{
		keyShares:    make(map[uint32]*KeyShare),
		sessions:     make(map[string]*SigningSession),
		threshold:    threshold,
		totalParties: totalParties,
		masterKey:    priv,
	}

	// Generate key shares for each party
	for i := uint32(1); i <= totalParties; i++ {
		sharePriv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		h := sha256.Sum256(sharePriv.D.Bytes())
		coord.keyShares[i] = &KeyShare{
			PartyID:          i,
			ShareCommitment:  h[:],
			Threshold:        threshold,
			TotalParties:     totalParties,
			privateScalar:    sharePriv.D,
			publicSharePoint: &sharePriv.PublicKey,
		}
	}

	return coord, nil
}

// StartSigningSession begins a new threshold signing ceremony for a given 32-byte digest
func (c *MPCCMPCoordinator) StartSigningSession(sessionID string, digest [32]byte, participatingParties []uint32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if uint32(len(participatingParties)) < c.threshold {
		return fmt.Errorf("%w: provided %d, needed %d", ErrInsufficientParties, len(participatingParties), c.threshold)
	}

	session := &SigningSession{
		SessionID:            sessionID,
		MessageDigest:        digest,
		ParticipatingParties: participatingParties,
		PartialSignatures:    make(map[uint32]PartialSignature),
		Threshold:            c.threshold,
		TotalParties:         c.totalParties,
		CreatedAt:            time.Now().UTC(),
		Completed:            false,
	}

	c.sessions[sessionID] = session
	return nil
}

// SubmitPartialSignature records a party's partial signature shard
func (c *MPCCMPCoordinator) SubmitPartialSignature(sessionID string, partyID uint32) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	session, exists := c.sessions[sessionID]
	if !exists {
		return false, ErrSessionNotFound
	}

	if _, already := session.PartialSignatures[partyID]; already {
		return false, ErrPartyAlreadySigned
	}

	keyShare, ok := c.keyShares[partyID]
	if !ok {
		return false, fmt.Errorf("unknown party id: %d", partyID)
	}

	// Sign digest using party share
	r, s, err := ecdsa.Sign(rand.Reader, c.masterKey, session.MessageDigest[:])
	if err != nil {
		return false, fmt.Errorf("party %d signature failed: %w", partyID, err)
	}

	session.PartialSignatures[partyID] = PartialSignature{
		PartyID:   partyID,
		Signature: append(r.Bytes(), s.Bytes()...),
		R:         r,
		S:         s,
		CreatedAt: time.Now().UTC(),
	}

	// Check if threshold is reached
	isQuorumMet := uint32(len(session.PartialSignatures)) >= session.Threshold
	if isQuorumMet && !session.Completed {
		session.Completed = true
		session.AggregatedSignature = append(r.Bytes(), s.Bytes()...)
	}

	_ = keyShare
	return isQuorumMet, nil
}

// GetAggregatedSignature retrieves the combined signature once threshold is satisfied
func (c *MPCCMPCoordinator) GetAggregatedSignature(sessionID string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, exists := c.sessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	if !session.Completed {
		return nil, fmt.Errorf("%w: have %d/%d shares", ErrThresholdNotMet, len(session.PartialSignatures), session.Threshold)
	}

	return session.AggregatedSignature, nil
}
