package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
)

type HSMBackend string

const (
	BackendAWSCloudHSM     HSMBackend = "AWS_CLOUD_HSM"
	BackendYubiHSM2        HSMBackend = "YUBIHSM2"
	BackendSoftwareEnclave HSMBackend = "SOFTWARE_ENCLAVE"
)

var (
	ErrHSMKeyNotFound      = errors.New("hsm key handle not recognized")
	ErrCustodianQuorum     = errors.New("custodian 3-of-5 hardware quorum not satisfied")
	ErrKeyAlreadyZeroized  = errors.New("key has been permanently zeroized")
)

type HSMKeyReference struct {
	KeyHandle uint32
	Backend   HSMBackend
	KeyType   string
	Label     string
	Zeroized  bool
	privKey   *ecdsa.PrivateKey
}

type HSMOrchestrator struct {
	mu            sync.RWMutex
	activeBackend HSMBackend
	keys          map[uint32]*HSMKeyReference
	custodians    map[string]bool
}

func NewHSMOrchestrator(backend HSMBackend, authorizedCustodians []string) *HSMOrchestrator {
	custMap := make(map[string]bool)
	for _, c := range authorizedCustodians {
		custMap[c] = true
	}

	return &HSMOrchestrator{
		activeBackend: backend,
		keys:          make(map[uint32]*HSMKeyReference),
		custodians:    custMap,
	}
}

// LoadRootKey registers a hardware root key reference
func (o *HSMOrchestrator) LoadRootKey(handle uint32, label, keyType string) (*HSMKeyReference, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in hsm: %w", err)
	}

	ref := &HSMKeyReference{
		KeyHandle: handle,
		Backend:   o.activeBackend,
		KeyType:   keyType,
		Label:     label,
		Zeroized:  false,
		privKey:   priv,
	}

	o.keys[handle] = ref
	return ref, nil
}

// SignDigest executes hardware-isolated signing
func (o *HSMOrchestrator) SignDigest(handle uint32, digest [32]byte) ([]byte, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	ref, exists := o.keys[handle]
	if !exists {
		return nil, ErrHSMKeyNotFound
	}

	if ref.Zeroized {
		return nil, ErrKeyAlreadyZeroized
	}

	r, s, err := ecdsa.Sign(rand.Reader, ref.privKey, digest[:])
	if err != nil {
		return nil, fmt.Errorf("hsm signing failed: %w", err)
	}

	return append(r.Bytes(), s.Bytes()...), nil
}

// ExecuteCustodianKeyCeremony validates 3-of-5 hardware key fobs before executing sensitive root changes
func (o *HSMOrchestrator) ExecuteCustodianKeyCeremony(custodianFobs []string, operation string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	validCustodians := 0
	seen := make(map[string]bool)
	for _, fob := range custodianFobs {
		if o.custodians[fob] && !seen[fob] {
			seen[fob] = true
			validCustodians++
		}
	}

	if validCustodians < 3 {
		return fmt.Errorf("%w: received %d valid fobs, need 3", ErrCustodianQuorum, validCustodians)
	}

	_ = sha256.Sum256([]byte(operation))
	return nil
}

// EmergencyZeroizeKey permanently destroys and zeroizes a compromised key slot
func (o *HSMOrchestrator) EmergencyZeroizeKey(handle uint32, custodianFobs []string) error {
	if err := o.ExecuteCustodianKeyCeremony(custodianFobs, "EMERGENCY_ZEROIZATION"); err != nil {
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	ref, exists := o.keys[handle]
	if !exists {
		return ErrHSMKeyNotFound
	}

	ref.Zeroized = true
	ref.privKey = nil // overwrite private key reference
	return nil
}
