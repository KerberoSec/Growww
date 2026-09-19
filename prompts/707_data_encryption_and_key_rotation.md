# 707 - Data Encryption Standards & Automated Key Rotation

## Purpose
Defines mandatory cryptographic standards for protecting all investor PII, financial transaction records, order book states, and blockchain private keys across data at rest, data in transit, and data in use. Implements NIST-approved envelope encryption (AES-256-GCM), zero-trust TLS 1.3 with Perfect Forward Secrecy across all internal and external communication channels, FIPS 140-2 Level 3 Hardware Security Module (HSM) key custody for blockchain signing nodes, and automated 90-day cryptographic key rotation pipelines with zero service disruption, in strict compliance with RBI Cyber Security Framework, SEBI CSCRF, and the Digital Personal Data Protection (DPDP) Act 2023.

## What You Are Building
- **Master Encryption Standards Specification:** `docs/security/encryption_standards.md` detailing approved cipher suites, key lengths, derivation functions, and data classification schemas.
- **Polyglot Field-Level Envelope Encryption SDK:** Shared cryptographic libraries (`packages/crypto-envelope/` in Go, Python, and Rust) utilizing Google Tink / AWS Encryption SDK for field-level database encryption (Aadhaar, PAN, Bank Account numbers).
- **Automated KMS & Vault Key Rotation Pipeline:** Infrastructure-as-code and scheduled cron automation rotating Key Encryption Keys (KEKs) every 90 days and background re-encrypting stale Data Encryption Keys (DEKs).
- **HSM PKCS#11 Integration Provider for Blockchain Nodes:** Hardware security module adapter enabling Hyperledger Besu validator nodes and transaction relayers to sign transactions without exposing private keys in plaintext memory.
- **UIDAI-Compliant Aadhaar Masking & Vault Engine:** Specialized tokenization module masking the first 8 digits of Aadhaar numbers and storing hashed reference tokens in isolated Vault storage.

## Scope Boundaries
- **In Scope:**
 - Data classification matrix (Confidential, Restricted, Internal, Public).
 - Field-level encryption for sensitive PII in PostgreSQL, ClickHouse, and Redis.
 - Storage volume encryption (LUKS / AWS KMS-managed EBS) and database Transparent Data Encryption (TDE).
 - Transport Layer Security (TLS 1.3) configuration with strict cipher suite enforcement.
 - Validator and relayer private key custody in FIPS 140-2 Level 3 CloudHSM / Vault Transit.
 - Automated 90-day key rotation and re-encryption workflows.
- **Out of Scope / Handled Elsewhere:**
 - Static secrets and application credential injection (handled in Prompt 109).
 - Database schema definitions (handled in Prompt 401).
 - Local client-side secure enclave storage on Flutter mobile devices (handled in Prompt 521).

## Technology to Use
- **Cryptographic Algorithms:** AES-256-GCM (Authenticated Encryption with Associated Data), RSA-4096 / ECDSA (secp256k1 & NIST P-384), HMAC-SHA256, Argon2id for password hashing.
- **Envelope Encryption SDK:** Google Tink / AWS Encryption SDK in Go, Python, and Rust.
- **Key Management & HSM:** AWS CloudHSM / AWS KMS / HashiCorp Vault Transit Secrets Engine with FIPS 140-2 Level 3 validation.
- **Transport Security:** TLS 1.3 enforcing `TLS_AES_256_GCM_SHA384` and `TLS_CHACHA20_POLY1305_SHA256` with mutual TLS (mTLS) across the Kubernetes service mesh.
- **Justification:** Google Tink provides mathematically proven cryptographic primitives that eliminate dangerous implementation pitfalls such as nonce reuse, padding oracle attacks, and unauthenticated ciphertexts.

## Backend / Infra Touchpoints
- **Datastores:** PostgreSQL 16 (401), Redis 7 (402), ClickHouse (404), S3/MinIO Document Storage (encrypted via SSE-KMS).
- **Key Management:** AWS KMS, HashiCorp Vault Cluster, AWS CloudHSM.
- **Services:** User Service (201), KYC Service (202), Wallet Service (203), Payment Gateway (212), Custodian Adapter (213).

## Blockchain Interaction
HSM key lifecycle management for permissioned Hyperledger Besu consortium nodes:
- **Validator Signing Keys:** Besu QBFT validator block-signing private keys are generated inside dedicated CloudHSM partitions via PKCS#11 interfaces. The validator node process never holds the raw private key in operating system RAM or disk storage.
- **Transaction Relayer Keys:** The relayer service (Prompt 305) signs mint, burn, and DvP settlement transactions by dispatching cryptographic signature requests to Vault Transit Engine with HSM backend.
- **On-Chain Key Rotation Hooks:** When a transaction relayer key rotates, the IAM pipeline executes an automated multisig transaction in `MultiSigGovernance.sol` to update the authorized relayer address whitelist in `ComplianceRegistry.sol`.

## Step-by-Step Build Instructions
1. Author the Master Encryption Standards document (`docs/security/encryption_standards.md`) establishing the Data Classification Schema:
 - **Restricted (Tier 1):** PAN, Aadhaar, Bank Account Numbers, Private Keys, Passwords, Biometric Templates.
 - **Confidential (Tier 2):** User Names, Phone Numbers, Email Addresses, Trade Histories, Portfolio Balances.
 - **Internal (Tier 3):** Service logs, Internal metrics, Order queue telemetry.
 - **Public (Tier 4):** Market ticker prices, Proof-of-reserve Merkle roots, Public disclosures.
2. Build the shared field-level encryption library `packages/crypto-envelope` in Go, Python 3.12, and Rust using Google Tink.
3. Implement Envelope Encryption architecture: generate unique, ephemeral AES-256 Data Encryption Keys (DEK) per record; encrypt DEK with AWS KMS / Vault Master Key Encryption Key (KEK); store `[encrypted_dek + iv + ciphertext + auth_tag]` in the database.
4. Implement the Aadhaar Masking & Tokenization Service: ensure only the last 4 digits of Aadhaar are visible (`XXXX-XXXX-1234`), compute SHA-256 HMAC for indexing/deduplication, and store encrypted raw data in an isolated Vault secure enclave.
5. Configure PostgreSQL Transparent Data Encryption (TDE) and verify all underlying cloud storage volumes (AWS EBS) are encrypted with customer-managed KMS keys.
6. Configure Envoy API Gateway and Kubernetes Istio/Linkerd service mesh to enforce TLS 1.3 exclusively, disabling all legacy TLS versions (1.0, 1.1, 1.2) and weak ciphers.
7. Integrate PKCS#11 HSM provider with Hyperledger Besu validator nodes, configuring hardware ECDSA `secp256k1` block signing.
8. Implement the automated 90-day AWS KMS / HashiCorp Vault KEK rotation pipeline using Terraform and Lambda/cron triggers.
9. Implement the asynchronous database background re-encryption worker (`scripts/crypto/re-encrypt-stale-keys.py`) to lazily re-encrypt historical records under rotated KEK versions without locking database tables.
10. Configure MinIO / AWS S3 Server-Side Encryption with KMS (SSE-KMS) and enforce bucket policies denying unencrypted object uploads (`s3:x-amz-server-side-encryption`).
11. Implement cryptographic audit logging in Kafka `audit.crypto.events` recording key creation, rotation, revocation, and high-frequency decryption anomaly alerts.
12. Build Known-Answer Test (KAT) suites verifying cross-platform encryption/decryption parity between Go, Python, Rust, and Dart implementations.
13. Conduct a cryptographic disaster simulation: test emergency KEK revocation and verify immediate cryptographic shredding of compromised data stores.

## Interfaces / Contracts
```protobuf
// schemas/crypto/v1/envelope.proto
syntax = "proto3";
package growww.crypto.v1;

message EncryptedEnvelope {
  string key_id = 1;          // KMS/Vault KEK identifier (e.g. "arn:aws:kms:ap-south-1:123456:key/abc-123")
  uint32 key_version = 2;     // Key version for rotation tracking
  string algorithm = 3;       // e.g. "AES256_GCM_HKDF_4KB"
  bytes encrypted_dek = 4;    // Ciphertext of the ephemeral Data Encryption Key
  bytes nonce = 5;            // 12-byte initialization vector / nonce
  bytes ciphertext = 6;       // Encrypted payload
  bytes tag = 7;              // 16-byte GCM authentication tag
  bytes associated_data = 8;  // Additional Authenticated Data (AAD) binding context (e.g. "user_id:12345")
}
```

```json
// Sample JSON serialized database field
{
  "v": 1,
  "kid": "kms-mumbai-pii-kek-v3",
  "alg": "AES-GCM-256",
  "edek": "AQIDAHh...",
  "iv": "3z92Gk1...",
  "ct": "k8Pz7...",
  "tag": "mN4b9..."
}
```

## Security & Compliance Notes
- **UIDAI Aadhaar Act 2016 & Circulars:** Mandates strict masking of Aadhaar numbers in all databases, prohibiting storage of raw 12-digit Aadhaar in plaintext or unencrypted format.
- **RBI Cyber Security Framework:** Requires cryptographic protection of payment card/bank data, end-to-end transport encryption, and minimum 90-day key rotation.
- **DPDP Act 2023:** Requires technical and organizational measures including encryption to prevent personal data breaches under penalty of up to ₹250 Crores.

## Acceptance Criteria
- [ ] Master Encryption Standards specification published in `docs/security/encryption_standards.md`.
- [ ] Field-level envelope encryption library (`packages/crypto-envelope`) implemented and verified in Go, Python, and Rust.
- [ ] 100% of sensitive PII (Aadhaar, PAN, Bank IFSC/Account) stored as encrypted envelopes in PostgreSQL and ClickHouse.
- [ ] Aadhaar masking strictly implemented; no raw 12-digit Aadhaar numbers exist anywhere in application databases or logs.
- [ ] TLS 1.3 with Perfect Forward Secrecy enforced across all external APIs and internal Kubernetes mTLS traffic.
- [ ] Blockchain validator signing keys operate inside FIPS 140-2 Level 3 CloudHSM via PKCS#11 interface.
- [ ] Automated 90-day KEK rotation pipeline operational with verified background re-encryption worker.
- [ ] Known-Answer Test (KAT) validation passes across all polyglot microservice runtimes.

## Suggested Order / Dependencies
- **Prerequisites:** 105 (Auth Architecture), 107 (Coding Standards), 109 (Secrets Management), 401 (PostgreSQL Schema).
- **Parallel Tasks:** 701 (Threat Model), 702 (IAM & RBAC).
- **Downstream Dependents:** 201 (User Service), 202 (KYC Service), 203 (Wallet Service), 311 (Validator Key Management), 521 (Flutter Secure Storage).
