package crypto_utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// EIP712Domain represents the EIP-712 domain separator structure
type EIP712Domain struct {
	Name              string
	Version           string
	ChainID           *big.Int
	VerifyingContract common.Address
}

// SpotOrderEIP712 represents an off-chain structured spot order for signing
type SpotOrderEIP712 struct {
	UserID    string
	Symbol    string
	Side      string // BUY or SELL
	AmountE8  uint64
	PriceE8   uint64
	Nonce     uint64
	Timestamp uint64
}

// DomainTypeHash is the keccak256 hash of the EIP712Domain type declaration
var DomainTypeHash = crypto.Keccak256([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))

// SpotOrderTypeHash is the keccak256 hash of the SpotOrder type declaration
var SpotOrderTypeHash = crypto.Keccak256([]byte("SpotOrder(string userId,string symbol,string side,uint64 amountE8,uint64 priceE8,uint64 nonce,uint64 timestamp)"))

// HashDomainSeparator computes the EIP-712 domain separator hash
func HashDomainSeparator(domain EIP712Domain) []byte {
	return crypto.Keccak256(
		DomainTypeHash,
		crypto.Keccak256([]byte(domain.Name)),
		crypto.Keccak256([]byte(domain.Version)),
		common.LeftPadBytes(domain.ChainID.Bytes(), 32),
		domain.VerifyingContract.Hash().Bytes(),
	)
}

// HashSpotOrder hashes a structured spot order according to EIP-712
func HashSpotOrder(order SpotOrderEIP712) []byte {
	amountBytes := common.LeftPadBytes(new(big.Int).SetUint64(order.AmountE8).Bytes(), 32)
	priceBytes := common.LeftPadBytes(new(big.Int).SetUint64(order.PriceE8).Bytes(), 32)
	nonceBytes := common.LeftPadBytes(new(big.Int).SetUint64(order.Nonce).Bytes(), 32)
	timestampBytes := common.LeftPadBytes(new(big.Int).SetUint64(order.Timestamp).Bytes(), 32)

	return crypto.Keccak256(
		SpotOrderTypeHash,
		crypto.Keccak256([]byte(order.UserID)),
		crypto.Keccak256([]byte(order.Symbol)),
		crypto.Keccak256([]byte(order.Side)),
		amountBytes,
		priceBytes,
		nonceBytes,
		timestampBytes,
	)
}

// ComputeTypedDataHash computes the \x19\x01 final EIP-712 digest
func ComputeTypedDataHash(domain EIP712Domain, order SpotOrderEIP712) []byte {
	domainSeparator := HashDomainSeparator(domain)
	orderHash := HashSpotOrder(order)

	raw := append([]byte("\x19\x01"), domainSeparator...)
	raw = append(raw, orderHash...)
	return crypto.Keccak256(raw)
}

// VerifySignature verifies an ECDSA secp256k1 signature and recovers the signer public address
func VerifySignature(digest []byte, signatureHex string, expectedAddress common.Address) (bool, error) {
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signatureHex, "0x"))
	if err != nil {
		return false, fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(sigBytes) != 65 {
		return false, errors.New("signature must be 65 bytes")
	}

	// Normalize Ethereum recovery ID v (27/28 -> 0/1)
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	pubKey, err := crypto.SigToPub(digest, sigBytes)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	recoveredAddress := crypto.PubkeyToAddress(*pubKey)
	return recoveredAddress == expectedAddress, nil
}

// SignTypedData signs an EIP-712 digest using a local ECDSA private key
func SignTypedData(digest []byte, privKey *ecdsa.PrivateKey) (string, error) {
	sig, err := crypto.Sign(digest, privKey)
	if err != nil {
		return "", err
	}
	// Transform v from 0/1 to 27/28
	sig[64] += 27
	return "0x" + hex.EncodeToString(sig), nil
}
