# Threshold Key Management & MPC-TSS Specification

## 1. Architecture Overview
Manages institutional cryptocurrency and tokenized asset custody using Multi-Party Computation with Threshold Signature Schemes (MPC-TSS, CGGMP21 protocol), ensuring private keys never exist in full memory at any single location.

## 2. Key Management Topology
- **Distributed Key Generation (DKG)**: Generates secp256k1 and ed25519 keypairs without a trusted dealer across five independent security boundaries:
  - Shard 1: Primary Exchange Enclave (AWS Nitro Enclaves - Mumbai).
  - Shard 2: Secondary Cloud Enclave (GCP Confidential Computing - Frankfurt).
  - Shard 3: Third Cloud Enclave (Azure Confidential Computing - Singapore).
  - Shard 4: Independent Custodian Bank HSM (FIPS 140-3 Level 3).
  - Shard 5: Offline Disaster Recovery Vault HSM (Air-Gapped GIFT City Facility).
- **Threshold Policy**: Mandatory 3-of-5 institutional quorum required to produce a valid signature.

## 3. Pre-Signing Pool for Sub-80ms Latency
To prevent multi-round network roundtrips from delaying automated settlement batches:
- The institutional MPC cluster executes offline pre-signing ceremonies during idle periods.
- Nonces ($R$ values) and partial pre-computations are stored in an encrypted offline cache (target capacity: 5,000 pre-computed nonces maintained with auto-replenishment low-watermark at 1,000 nonces).
- Online phase requires only a single network round (<80ms) to combine signature shares into a valid standard ECDSA signature.
- Atomic single-use pop semantics via Redis Lua script (`GETDEL`) combined with hardware monotonic counter checks eliminate any possibility of ECDSA nonce reuse.

## 4. EIP-4361 Sign-In with Ethereum (SIWE) & Ephemeral Sessions
- Institutional and client Web3 authentication follows EIP-4361 with chain ID 13370 enforcement.
- Server-generated nonces expire after 5 minutes (TTL in Redis) and are single-use (`GETDEL`).
- Smart contract wallets are validated via ERC-1271 (`isValidSignature`).
- Verified sessions receive an ephemeral JWT valid for 15 minutes, strictly scoped to Web3 custody operations (`custody:sign_intent`, `custody:view_balance`).

## 5. Security Invariants
- Dynamic key share refreshing (Proactive Secret Sharing, PSS) every 24 hours without changing the public address.
- Zero raw private keys written to disk, memory dumps, or logs.
- Hard velocity caps on automated withdrawals: transactions > 25,000 USDT require manual compliance multi-party sign-off.
- FIPS 140-3 Level 3 compliance across all physical and cloud HSM modules.
