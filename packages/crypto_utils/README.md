# Cryptographic Utilities & Primitives

## Purpose & Scope
`packages/crypto_utils` provides audited cryptographic helper libraries for Web3 signatures, Bitcoin script derivation, and zero-knowledge solvency verification.

## Capabilities
- **EIP-712 Signing**: Structured data hashing and ECDSA secp256k1 recovery.
- **Taproot Address Derivation**: BIP-341/342 key-path and script-path P2TR address generators.
- **Sparse Merkle Tree (SMT)**: High-performance tree construction and inclusion proof verification.
- **Fixed-Point Math**: Safe 64-bit and 128-bit integer arithmetic with 8 decimal places (`e8`).
