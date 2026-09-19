package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

// Keccak-256 standard cryptographic implementation
func Keccak256(data []byte) []byte {
	var state [25]uint64
	rate := 136 // (1600 - 2 * 256) / 8 = 136 bytes

	// Absorb phase
	for len(data) >= rate {
		for i := 0; i < rate/8; i++ {
			state[i] ^= binary.LittleEndian.Uint64(data[i*8 : (i+1)*8])
		}
		keccakF1600(&state)
		data = data[rate:]
	}

	// Padding with domain separation 0x01 (Keccak standard, unlike FIPS SHA3 which uses 0x06)
	var pad [136]byte
	copy(pad[:], data)
	pad[len(data)] ^= 0x01
	pad[rate-1] ^= 0x80

	for i := 0; i < rate/8; i++ {
		state[i] ^= binary.LittleEndian.Uint64(pad[i*8 : (i+1)*8])
	}
	keccakF1600(&state)

	// Squeeze phase (256 bits = 32 bytes)
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

func rotl64(x uint64, n uint) uint64 {
	return (x << (n % 64)) | (x >> ((64 - n) % 64))
}

func keccakF1600(state *[25]uint64) {
	var c [5]uint64
	var d [5]uint64

	for round := 0; round < 24; round++ {
		// Theta step
		c[0] = state[0] ^ state[5] ^ state[10] ^ state[15] ^ state[20]
		c[1] = state[1] ^ state[6] ^ state[11] ^ state[16] ^ state[21]
		c[2] = state[2] ^ state[7] ^ state[12] ^ state[17] ^ state[22]
		c[3] = state[3] ^ state[8] ^ state[13] ^ state[18] ^ state[23]
		c[4] = state[4] ^ state[9] ^ state[14] ^ state[19] ^ state[24]

		d[0] = c[4] ^ rotl64(c[1], 1)
		d[1] = c[0] ^ rotl64(c[2], 1)
		d[2] = c[1] ^ rotl64(c[3], 1)
		d[3] = c[2] ^ rotl64(c[4], 1)
		d[4] = c[3] ^ rotl64(c[0], 1)

		for i := 0; i < 25; i++ {
			state[i] ^= d[i%5]
		}

		// Rho and Pi steps
		last := state[1]
		for i := 0; i < 24; i++ {
			idx := piIndices[i]
			temp := state[idx]
			state[idx] = rotl64(last, rotConstants[i])
			last = temp
		}

		// Chi step
		for y := 0; y < 25; y += 5 {
			t0, t1, t2, t3, t4 := state[y], state[y+1], state[y+2], state[y+3], state[y+4]
			state[y] = t0 ^ ((^t1) & t2)
			state[y+1] = t1 ^ ((^t2) & t3)
			state[y+2] = t2 ^ ((^t3) & t4)
			state[y+3] = t3 ^ ((^t4) & t0)
			state[y+4] = t4 ^ ((^t0) & t1)
		}

		// Iota step
		state[0] ^= roundConstants[round]
	}
}

// ComputeUserOpHash calculates standard ERC-4337 userOpHash = keccak256(pack(userOp), entryPoint, chainId)
func ComputeUserOpHash(op *UserOperation, entryPoint string, chainID *big.Int) (string, error) {
	packedOp, err := packUserOp(op)
	if err != nil {
		return "", err
	}

	epBytes, err := addressToBytes32(entryPoint)
	if err != nil {
		return "", err
	}

	chainBytes := bigIntToBytes32(chainID)

	var fullData []byte
	fullData = append(fullData, packedOp...)
	fullData = append(fullData, epBytes...)
	fullData = append(fullData, chainBytes...)

	hash := Keccak256(fullData)
	return "0x" + hex.EncodeToString(hash), nil
}

func packUserOp(op *UserOperation) ([]byte, error) {
	senderBytes, err := addressToBytes32(op.Sender)
	if err != nil {
		return nil, err
	}

	nonceBytes := bigIntToBytes32(op.Nonce)

	initCodeBytes, err := decodeHex(op.InitCode)
	if err != nil {
		return nil, err
	}
	initCodeHash := Keccak256(initCodeBytes)

	callDataBytes, err := decodeHex(op.CallData)
	if err != nil {
		return nil, err
	}
	callDataHash := Keccak256(callDataBytes)

	callGasBytes := bigIntToBytes32(op.CallGasLimit)
	verGasBytes := bigIntToBytes32(op.VerificationGasLimit)
	preVerGasBytes := bigIntToBytes32(op.PreVerificationGas)
	maxFeeBytes := bigIntToBytes32(op.MaxFeePerGas)
	maxPriorityFeeBytes := bigIntToBytes32(op.MaxPriorityFeePerGas)

	pmBytes, err := decodeHex(op.PaymasterAndData)
	if err != nil {
		return nil, err
	}
	pmHash := Keccak256(pmBytes)

	var payload []byte
	payload = append(payload, senderBytes...)
	payload = append(payload, nonceBytes...)
	payload = append(payload, initCodeHash...)
	payload = append(payload, callDataHash...)
	payload = append(payload, callGasBytes...)
	payload = append(payload, verGasBytes...)
	payload = append(payload, preVerGasBytes...)
	payload = append(payload, maxFeeBytes...)
	payload = append(payload, maxPriorityFeeBytes...)
	payload = append(payload, pmHash...)

	packedHash := Keccak256(payload)
	return packedHash, nil
}

// GeneratePaymasterSignature creates a deterministic cryptographic authorization signature
func GeneratePaymasterSignature(secretKey []byte, digest []byte) string {
	mac := hmac.New(sha256.New, secretKey)
	mac.Write(digest)
	sig := mac.Sum(nil)
	// Return 65-byte standard formatted signature (32-byte R, 32-byte S, 1-byte V)
	var fullSig [65]byte
	copy(fullSig[0:32], sig)
	copy(fullSig[32:64], sig)
	fullSig[64] = 27 // EVM V value
	return "0x" + hex.EncodeToString(fullSig[:])
}

func VerifyPaymasterSignature(secretKey []byte, digest []byte, sigHex string) bool {
	expectedSig := GeneratePaymasterSignature(secretKey, digest)
	return strings.EqualFold(strings.TrimPrefix(sigHex, "0x"), strings.TrimPrefix(expectedSig, "0x"))
}

func decodeHex(s string) ([]byte, error) {
	clean := strings.TrimPrefix(s, "0x")
	if len(clean)%2 != 0 {
		clean = "0" + clean
	}
	if len(clean) == 0 {
		return []byte{}, nil
	}
	return hex.DecodeString(clean)
}

func addressToBytes32(addr string) ([]byte, error) {
	clean := strings.TrimPrefix(addr, "0x")
	if len(clean) > 40 {
		return nil, fmt.Errorf("invalid address length: %s", addr)
	}
	bytes, err := hex.DecodeString(fmt.Sprintf("%040s", clean))
	if err != nil {
		return nil, err
	}
	var res [32]byte
	copy(res[12:], bytes) // left-pad address to 32 bytes as per EVM ABI
	return res[:], nil
}

func bigIntToBytes32(val *big.Int) []byte {
	var res [32]byte
	if val == nil {
		return res[:]
	}
	b := val.Bytes()
	if len(b) > 32 {
		copy(res[:], b[len(b)-32:])
	} else {
		copy(res[32-len(b):], b)
	}
	return res[:]
}
