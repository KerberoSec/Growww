package src

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	HardenedBit uint32 = 0x80000000

	CoinTypeBitcoin  uint32 = 0
	CoinTypeEthereum uint32 = 60
	CoinTypeSolana   uint32 = 501
)

var (
	ErrInvalidBIP44Path = errors.New("bip44: invalid derivation path")
	bip32SeedKey        = []byte("Bitcoin seed")
)

// BIP44Path holds decoded path indices: m / 44' / coin_type' / account' / change / address_index
type BIP44Path struct {
	Purpose      uint32
	CoinType     uint32
	Account      uint32
	Change       uint32
	AddressIndex uint32
}

// ParseBIP44Path parses a path string like "m/44'/60'/0'/0/0"
func ParseBIP44Path(path string) (*BIP44Path, error) {
	parts := strings.Split(path, "/")
	if len(parts) != 6 || parts[0] != "m" {
		return nil, fmt.Errorf("%w: path must contain exactly 5 levels starting with 'm'", ErrInvalidBIP44Path)
	}

	indices := make([]uint32, 5)
	for i := 1; i <= 5; i++ {
		raw := parts[i]
		hardened := false
		if strings.HasSuffix(raw, "'") || strings.HasSuffix(raw, "h") {
			hardened = true
			raw = strings.TrimRight(raw, "'h")
		}

		val, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid segment '%s'", ErrInvalidBIP44Path, parts[i])
		}

		idx := uint32(val)
		if hardened {
			idx |= HardenedBit
		}
		indices[i-1] = idx
	}

	// BIP-44 requires purpose = 44'
	if indices[0] != (44 | HardenedBit) {
		return nil, fmt.Errorf("%w: purpose must be 44'", ErrInvalidBIP44Path)
	}

	return &BIP44Path{
		Purpose:      indices[0] ^ HardenedBit,
		CoinType:     indices[1] ^ HardenedBit,
		Account:      indices[2] ^ HardenedBit,
		Change:       indices[3],
		AddressIndex: indices[4],
	}, nil
}

// Format returns canonical string representation of the path.
func (p *BIP44Path) Format() string {
	return fmt.Sprintf("m/44'/%d'/%d'/%d/%d", p.CoinType, p.Account, p.Change, p.AddressIndex)
}

// ExtendedKey represents a private key and its 32-byte chain code.
type ExtendedKey struct {
	Key       [32]byte
	ChainCode [32]byte
	Depth     uint8
}

// DeriveMasterKey generates the master node from a 64-byte mnemonic seed via HMAC-SHA512.
func DeriveMasterKey(seed []byte) (*ExtendedKey, error) {
	if len(seed) < 16 || len(seed) > 128 {
		return nil, errors.New("bip32: seed length must be between 16 and 128 bytes")
	}

	h := hmac.New(sha512.New, bip32SeedKey)
	h.Write(seed)
	sum := h.Sum(nil)

	var key, chainCode [32]byte
	copy(key[:], sum[:32])
	copy(chainCode[:], sum[32:64])

	return &ExtendedKey{
		Key:       key,
		ChainCode: chainCode,
		Depth:     0,
	}, nil
}

// DeriveChild derives a child key using HMAC-SHA512 with hardened/non-hardened index.
func (k *ExtendedKey) DeriveChild(index uint32) *ExtendedKey {
	h := hmac.New(sha512.New, k.ChainCode[:])

	if index&HardenedBit != 0 {
		// Hardened: 0x00 || key || index
		h.Write([]byte{0x00})
		h.Write(k.Key[:])
	} else {
		// Non-hardened: simplified pubkey hash || index
		h.Write(k.Key[:])
	}

	var idxBytes [4]byte
	binary.BigEndian.PutUint32(idxBytes[:], index)
	h.Write(idxBytes[:])

	sum := h.Sum(nil)
	var childKey, childChainCode [32]byte
	copy(childKey[:], sum[:32])
	copy(childChainCode[:], sum[32:64])

	return &ExtendedKey{
		Key:       childKey,
		ChainCode: childChainCode,
		Depth:     k.Depth + 1,
	}
}
