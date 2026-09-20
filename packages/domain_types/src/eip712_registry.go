package src

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
)

// Keccak256 computes standard Ethereum Keccak-256 hash.
func Keccak256(data []byte) []byte {
	var state [25]uint64
	rate := 136

	for len(data) >= rate {
		for i := 0; i < rate/8; i++ {
			state[i] ^= binary.LittleEndian.Uint64(data[i*8 : (i+1)*8])
		}
		keccakF1600(&state)
		data = data[rate:]
	}

	var pad [136]byte
	copy(pad[:], data)
	pad[len(data)] ^= 0x01
	pad[rate-1] ^= 0x80

	for i := 0; i < rate/8; i++ {
		state[i] ^= binary.LittleEndian.Uint64(pad[i*8 : (i+1)*8])
	}
	keccakF1600(&state)

	out := make([]byte, 32)
	for i := 0; i < 4; i++ {
		binary.LittleEndian.PutUint64(out[i*8:(i+1)*8], state[i])
	}
	return out
}

var roundConstants = [24]uint64{
	0x0000000000000001, 0x0000000000008082, 0x800000000000808A, 0x8000000080008000,
	0x000000000000808B, 0x0000000080000001, 0x8000000080008081, 0x8000000000008009,
	0x000000000000008A, 0x0000000000000088, 0x0000000080008009, 0x000000008000000A,
	0x000000008000808B, 0x800000000000008B, 0x8000000000008089, 0x8000000000008003,
	0x8000000000008002, 0x8000000000000080, 0x000000000000800A, 0x800000008000000A,
	0x8000000080008081, 0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
}

var rotConstants = [24]uint{
	1, 3, 6, 10, 15, 21, 28, 36, 45, 55, 2, 14,
	27, 41, 56, 8, 25, 43, 62, 18, 39, 61, 20, 44,
}

var piIndices = [24]int{
	10, 7, 11, 17, 18, 3, 5, 16, 8, 21, 24, 4,
	15, 23, 19, 13, 12, 2, 20, 14, 22, 9, 6, 1,
}

func rol64(x uint64, n uint) uint64 {
	return (x << (n % 64)) | (x >> ((64 - (n % 64)) % 64))
}

func keccakF1600(state *[25]uint64) {
	var c [5]uint64
	var d [5]uint64
	var b [25]uint64

	for round := 0; round < 24; round++ {
		// Theta step
		for i := 0; i < 5; i++ {
			c[i] = state[i] ^ state[i+5] ^ state[i+10] ^ state[i+15] ^ state[i+20]
		}
		for i := 0; i < 5; i++ {
			d[i] = c[(i+4)%5] ^ rol64(c[(i+1)%5], 1)
		}
		for i := 0; i < 25; i++ {
			state[i] ^= d[i%5]
		}

		// Rho and Pi steps
		b[0] = state[0]
		for i := 0; i < 24; i++ {
			b[piIndices[i]] = rol64(state[piIndices[i]], rotConstants[i])
		}

		// Chi step
		for j := 0; j < 25; j += 5 {
			for i := 0; i < 5; i++ {
				state[j+i] = b[j+i] ^ ((^b[j+(i+1)%5]) & b[j+(i+2)%5])
			}
		}

		// Iota step
		state[0] ^= roundConstants[round]
	}
}

// EIP712Domain represents domain separation fields per EIP-712.
type EIP712Domain struct {
	Name              string   `json:"name"`
	Version           string   `json:"version"`
	ChainID           *big.Int `json:"chain_id"`
	VerifyingContract string   `json:"verifying_contract"`
}

// EIP712DomainSeparatorRegistry caches and validates precomputed EIP-712 domain separators.
type EIP712DomainSeparatorRegistry struct {
	mu        sync.RWMutex
	cache     map[string][32]byte
	typeHash  [32]byte
}

const EIP712DomainTypeString = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"

// NewEIP712DomainSeparatorRegistry initializes registry and computes canonical domain typehash.
func NewEIP712DomainSeparatorRegistry() *EIP712DomainSeparatorRegistry {
	th := Keccak256([]byte(EIP712DomainTypeString))
	var typeHash [32]byte
	copy(typeHash[:], th)

	return &EIP712DomainSeparatorRegistry{
		cache:    make(map[string][32]byte),
		typeHash: typeHash,
	}
}

// ComputeSeparator calculates the 32-byte EIP-712 domain separator.
func (r *EIP712DomainSeparatorRegistry) ComputeSeparator(d EIP712Domain) ([32]byte, error) {
	if d.Name == "" || d.Version == "" || d.ChainID == nil || d.VerifyingContract == "" {
		return [32]byte{}, errors.New("invalid EIP712 domain: required fields missing")
	}

	cleanAddr := strings.ToLower(strings.TrimPrefix(d.VerifyingContract, "0x"))
	addrBytes, err := hex.DecodeString(cleanAddr)
	if err != nil || len(addrBytes) != 20 {
		return [32]byte{}, fmt.Errorf("invalid verifyingContract address %s", d.VerifyingContract)
	}

	nameHash := Keccak256([]byte(d.Name))
	versionHash := Keccak256([]byte(d.Version))

	// ABI encode packed 5 words: typeHash, nameHash, versionHash, chainId, verifyingContract
	var encoded [160]byte
	copy(encoded[0:32], r.typeHash[:])
	copy(encoded[32:64], nameHash)
	copy(encoded[64:96], versionHash)

	// Chain ID as 32-byte big-endian uint256
	chainIDBytes := d.ChainID.Bytes()
	copy(encoded[128-len(chainIDBytes):128], chainIDBytes)

	// Address as 32-byte padded word (12 zero bytes followed by 20 address bytes)
	copy(encoded[140:160], addrBytes)

	sep := Keccak256(encoded[:])
	var res [32]byte
	copy(res[:], sep)
	return res, nil
}

// RegisterDomain computes and registers a domain separator.
func (r *EIP712DomainSeparatorRegistry) RegisterDomain(d EIP712Domain) ([32]byte, error) {
	sep, err := r.ComputeSeparator(d)
	if err != nil {
		return [32]byte{}, err
	}

	key := fmt.Sprintf("%s:%s", d.ChainID.String(), strings.ToLower(d.VerifyingContract))
	r.mu.Lock()
	r.cache[key] = sep
	r.mu.Unlock()

	return sep, nil
}

// GetSeparator returns the cached separator if available.
func (r *EIP712DomainSeparatorRegistry) GetSeparator(chainID *big.Int, contractAddr string) ([32]byte, bool) {
	key := fmt.Sprintf("%s:%s", chainID.String(), strings.ToLower(contractAddr))
	r.mu.RLock()
	defer r.mu.RUnlock()
	val, ok := r.cache[key]
	return val, ok
}

// HashTypedDataV4 computes \x19\x01 || domainSeparator || structHash.
func (r *EIP712DomainSeparatorRegistry) HashTypedDataV4(domainSeparator [32]byte, structHash [32]byte) [32]byte {
	var payload [66]byte
	payload[0] = 0x19
	payload[1] = 0x01
	copy(payload[2:34], domainSeparator[:])
	copy(payload[34:66], structHash[:])

	digest := Keccak256(payload[:])
	var res [32]byte
	copy(res[:], digest)
	return res
}
