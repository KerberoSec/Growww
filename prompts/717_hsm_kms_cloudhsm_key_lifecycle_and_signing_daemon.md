# 717 - HSM & CloudHSM Key Lifecycle and Cryptographic Signing Daemon (FIPS 140-3 Level 3, EIP-712, BIP-174 PSBT, QBFT Block Signing)

## Purpose
Establishes an institutional-grade, zero-trust cryptographic key lifecycle management and hardware signing microservice (`services/cloudhsm-signer-daemon`). The daemon provides a secure physical and logical isolation boundary for all mission-critical cryptographic operations across the platform, interfacing directly with AWS CloudHSM clusters and dedicated Thales Luna PCIe/Network Hardware Security Modules (HSMs) operating under **FIPS 140-3 Level 3** physical and logical tamper-response standards.

In strict compliance with **SEBI Cybersecurity and Cyber Resilience Framework (CSCRF)**, **RBI Cyber Security Framework for Financial Intermediaries**, **IFSCA Capital Markets Intermediaries Regulations**, and **NIST SP 800-57 / SP 800-90A Key Management Guidelines**, the service guarantees that private keys are generated inside certified hardware boundaries and can never be exported or read in plaintext by application processes, memory dump exploits, or administrative operators (`CKA_EXTRACTABLE = FALSE`, `CKA_SENSITIVE = TRUE`). 

The daemon acts as the centralized, ultra-low-latency signing oracle for:
1. **Hyperledger Besu QBFT Validator Block Signing:** Sub-5 millisecond deterministic block proposal and commit signing with hardware-enforced anti-equivocation (slashing prevention) state tracking.
2. **EIP-712 Structured DvP Settlement Batch Authorization:** Domain-separated cryptographic signing for off-chain trade matching batches, tokenized equity transfers, and regulatory compliance hooks.
3. **BIP-174 Partially Signed Bitcoin Transaction (PSBT) Signing:** Script validation, UTXO ownership verification, and multisig threshold signing for institutional reserve vaults.
4. **Automated Key Lifecycle & Ceremony Orchestration:** Quorum-based M-of-N key generation ceremonies, scheduled zero-downtime key rotation pipelines, automated on-chain governance registration, and instant cryptographic shredding upon revocation.

## What You Are Building
A high-throughput, fault-tolerant cryptographic signing microservice (`services/cloudhsm-signer-daemon`):
- **PKCS#11 Hardware Abstraction & Session Pool Manager:** High-concurrency connection pool communicating over PKCS#11 v2.40 (Cryptoki) with AWS CloudHSM / Thales Luna HSM clusters, managing multi-partition failover, session timeouts, and load balancing across hardware security modules.
- **QBFT Validator Consensus Block Signer & Anti-Equivocation Guard:** High-speed secp256k1 block signing module with an embedded, ACID-compliant local state store that verifies monotonically increasing round and block sequence numbers, mathematically eliminating double-signing risk.
- **EIP-712 Structured Data Settlement Batch Signer:** Parser and validator for EIP-712 typed structured data hashes, verifying verifying-contract addresses, chain IDs, transfer volume velocity caps, and settlement batch Merkle roots before executing ECDSA signatures.
- **BIP-174 PSBT Multisig Transaction Signer:** Bitcoin and UTXO-compatible transaction processor that parses raw PSBT payloads, validates non-witness and witness UTXOs, checks fee rates against maximum allowed tolerance, and computes hardware-isolated signatures.
- **M-of-N Key Ceremony & Automated Rotation Coordinator:** Cryptographic state machine orchestrating key generation ceremonies using Shamir Secret Sharing (SSS) and multi-officer quorum authorization, publishing public key metadata to on-chain governance registries while archiving old keys into verification-only mode.
- **Hardware-Enforced Cryptographic Policy Engine:** Validates that incoming signing requests originate from authenticated microservices over mutual TLS (mTLS), adhere to per-service rate limits, and conform to strict payload validation rules before unlocking HSM signing slots.
- **Tamper-Evident SIEM Audit Log Relayer:** Real-time event streaming pipe dispatching cryptographically hashed audit records for every key access, signing attempt, ceremony stage, and equivocation violation to Apache Kafka and SIEM collectors.

## Scope Boundaries
- **In Scope:**
  - Hardware-isolated private key generation inside FIPS 140-3 Level 3 HSM hardware partitions with unextractable attributes (`CKA_EXTRACTABLE = FALSE`, `CKA_PRIVATE = TRUE`, `CKA_SENSITIVE = TRUE`).
  - High-performance PKCS#11 session management and multi-HSM cluster connection pooling.
  - Hyperledger Besu QBFT validator consensus message signing (Proposal, Prepare, Commit, Round Change) with sub-5ms p99 latency.
  - Local persistent anti-equivocation database with fsync write-ahead logging to prevent double-signing.
  - EIP-712 typed structured data hashing and ECDSA `(v, r, s)` signing for trade settlement batches and compliance directives.
  - BIP-174 PSBT parsing, input validation, fee sanity checks, and hardware signing for Bitcoin multisig custody.
  - Ed25519 signing for Solana / cross-chain operations and RSA-4096 / ECDSA NIST P-384 signing for X.509 PKI certificates.
  - Automated and ceremony-driven key rotation workflows, public key lifecycle state transitions, and cryptographic key destruction.
  - Synchronous gRPC API with mutual TLS (mTLS) client certificate verification and role-based access control.
  - High-frequency audit logging to Kafka, ClickHouse, and PostgreSQL.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain limit order book matching and trade execution (handled in Prompt 205).
  - Off-chain trade clearing and DvP settlement batch aggregation algorithms (handled in Prompt 208).
  - Retail investor client-side MPC key shares and social recovery (handled in Prompt 237).
  - Smart contract settlement state execution and fee collection (handled in Prompt 329).
  - General database field-level envelope encryption SDK for PII masking (handled in Prompt 707).
  - Physical vault weighment and bar assay verification (handled in Prompt 716).

## Technology to Use
- **Core Microservice:** Rust 1.78+ / Go 1.22+ for deterministic execution, zero garbage collection pauses during block consensus signing, strict memory safety, and native C-bindings to vendor PKCS#11 libraries.
- **Hardware Security Modules (HSM):** AWS CloudHSM (LiquidSecurity PCIe cryptographic accelerators) / Thales Luna HSM 7 (FIPS 140-3 Level 3 validated physical security boundary).
- **Cryptographic Specifications & APIs:** PKCS#11 v2.40 (OASIS Cryptoki API), OpenSSL 3.2+ FIPS Provider, libsecp256k1 C-bindings, Rust `cryptoki` / Go `miekg/pkcs11`.
- **Anti-Equivocation Local State Engine:** RocksDB / SQLite in WAL mode with synchronous fsync on dedicated NVMe storage for sub-millisecond atomic block sequence persistence.
- **Relational Metadata & Lifecycle Database:** PostgreSQL 16 with Row-Level Security (RLS) for storing public key registries, key aliases, rotation schedules, and ceremony state logs.
- **Cryptographic Audit Trail & Telemetry:** ClickHouse 24+ / TimescaleDB for immutable append-only storage of high-volume signing audit events and latency metrics.
- **Event Bus & SIEM Ingestion:** Apache Kafka for real-time publishing of key lifecycle notifications, signing alerts, and security anomaly detections.
- **Inter-Service Communication:** gRPC over HTTP/2 with mutual TLS (mTLS) enforcing X.509 client certificate pinning and AES-256-GCM transport encryption.
- **Justification:** Private key custody in capital markets requires uncompromising hardware isolation and sub-millisecond deterministic signing latency. Rust/Go combined with standard PKCS#11 interfaces ensures direct interaction with certified physical HSMs, preventing runtime key extraction while sustaining over 2,000 cryptographic signatures per second per HSM partition.

## Backend / Infra Touchpoints
- **Upstream Microservices & Callers:**
  - Hyperledger Besu Validator Nodes (Prompts 301, 310, 311): Consensus engine dispatching QBFT block proposal and commit signing requests via secure local gRPC/IPC.
  - `services/trade-settlement` (Prompt 208): Settlement engine dispatching EIP-712 DvP batch authorization requests.
  - `services/custody-bridge` (Prompt 319): Institutional custody gateway dispatching BIP-174 PSBT signing requests for cold/warm vault sweeps.
  - `services/transfer-compliance` (Prompt 305): Compliance relayer requesting signed authorization certificates for token transfers.
  - `services/admin-back-office` (Prompts 217, 604, 605): Security administrators and Crypto Officers initiating key generation and rotation ceremonies.
- **Downstream Systems & Infrastructure:**
  - Physical AWS CloudHSM / Thales Luna HSM Clusters: Hardware cryptographic engine executing ECDSA, Ed25519, and RSA operations.
  - Hyperledger Besu Consortium Blockchain: Ingests transactions signed by authorized relayer keys and validator nodes.
  - PostgreSQL 16 & ClickHouse 24+: Stores key metadata, ceremony proofs, and SIEM audit records.
  - OpenTelemetry / Prometheus (Prompt 806): Monitors HSM latency, session pool utilization, and signing failure rates.
- **Messaging Topics (Apache Kafka):**
  - Consumes: `governance.key_ceremony.initiated.v1`, `governance.key_rotation.approved.v1`.
  - Publishes: `crypto.signing.audit.v1`, `crypto.key_lifecycle.event.v1`, `crypto.equivocation_attempt.flagged.v1`, `crypto.key_rotation.completed.v1`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus.
- **Validator Consensus Block Signing:**
  - Dedicated secp256k1 validator private keys reside permanently inside isolated HSM partitions.
  - When the Besu consensus engine proposes or commits to a block, it forwards the canonical QBFT digest to the signer daemon.
  - The daemon checks the local RocksDB anti-equivocation table: if the tuple `(chain_id, validator_address, block_number, sequence, round)` was already signed with a different block hash, the daemon immediately rejects the request, raises a critical security alarm, and halts signing to prevent network slashing.
  - Upon passing validation, the HSM calculates the deterministic ECDSA signature according to RFC 6979 / SEC 1 format, returning `(r, s, v)` to the node.
- **EIP-712 Settlement Batch Authorization:**
  - Relayer keys sign structured data hashes for atomic DvP trade execution in `NBSESettlementDvP.sol` (Prompt 329).
  - The daemon parses the payload against the canonical EIP-712 domain schema:
    $$\text{DomainSeparator} = \text{keccak256}(\text{abi.encode}(\text{EIP712DomainTypeHash}, \text{nameHash}, \text{versionHash}, \text{chainId}, \text{verifyingContract}))$$
    $$\text{StructHash} = \text{keccak256}(\text{abi.encode}(\text{BatchSettlementTypeHash}, \text{batchId}, \text{merkleRoot}, \text{tradeCount}, \text{grossValueINR}, \text{settlementWindowExpiry}))$$
    $$\text{Digest} = \text{keccak256}(\text{abi.encodePacked("\x19\x01"}, \text{DomainSeparator}, \text{StructHash}))$$
  - The daemon verifies that the verifying contract address matches the approved on-chain deployment whitelist before generating the signature.
- **On-Chain Key Lifecycle & Governance Synchronizer:**
  - Upon completion of a key rotation ceremony, the daemon outputs an attestation payload containing the new public key, certificate digest, and M-of-N Crypto Officer signatures.
  - The governance relayer submits an update transaction to `MultiSigGovernance.sol` (Prompt 307) and `ComplianceRegistry.sol` (Prompt 305) to update the authorized relayer / validator whitelist on-chain.
- **Zero PII on Ledger:** Ledger records contain strictly pseudonymized cryptographic public keys, key IDs, signature digests, and domain hashes. No operator names, officer credentials, or server network addresses are committed to the ledger.

## Step-by-Step Build Instructions
1. Scaffold the `services/cloudhsm-signer-daemon` repository following the standard monorepo structure (Prompt 106) with strict compile-time type checking, static analysis, and zero application code outside authorized packages (Prompt 107).
2. Generate Go, Rust, and Python gRPC stubs from the Protobuf definition in `proto/growww/crypto_signer/v1/signer_daemon.proto`.
3. Design and execute PostgreSQL 16 database migrations for tables `hsm_clusters`, `hsm_key_registry`, `qbft_equivocation_state`, `key_rotation_ceremonies`, `key_ceremony_shares`, and `cryptographic_signing_audit_log`.
4. Configure ClickHouse schema and table engines for immutable, high-throughput cryptographic audit trail logging.
5. Implement the **PKCS#11 Hardware Session Pool Manager**:
   - Initialize connection to AWS CloudHSM / Thales Luna HSM client libraries via `C_Initialize` and `C_OpenSession`.
   - Implement thread-safe connection pooling with keep-alive health checks, automatic reconnects, and failover across primary and standby HSM partitions.
   - Enforce read-only and read-write session segregation for Crypto Officer (CO) and Crypto User (CU) roles.
6. Implement the **Key Lifecycle Management Module**:
   - Hardware key generation for secp256k1 (ECDSA / Schnorr), Ed25519, NIST P-384, and RSA-4096 using the HSM True Random Number Generator (TRNG).
   - Set immutable PKCS#11 attributes on all generated private keys: `CKA_EXTRACTABLE = FALSE`, `CKA_PRIVATE = TRUE`, `CKA_SENSITIVE = TRUE`, `CKA_SIGN = TRUE`, `CKA_TOKEN = TRUE`.
   - Export public keys and X.509 public certificates while maintaining key handle references in PostgreSQL `hsm_key_registry`.
7. Implement the **QBFT Validator Block Signer & Anti-Equivocation Guard**:
   - Embed RocksDB with write-ahead logging (WAL) on NVMe storage for zero-latency local state persistence.
   - Before signing any consensus message (Proposal, Prepare, Commit, RoundChange), atomically query and store the composite key `(chain_id, validator_pubkey, block_number, round, message_type)`.
   - Reject any duplicate signing request with a conflicting block hash, publish an emergency `crypto.equivocation_attempt.flagged.v1` alert to Kafka, and trip the circuit breaker.
   - Execute deterministic ECDSA signature on secp256k1 inside the HSM and return formatted `(r, s, v)` bytes.
8. Implement the **EIP-712 Structured Data Settlement Batch Signer**:
   - Parse inbound EIP-712 JSON/Protobuf payloads and validate domain separator parameters (`chainId`, `verifyingContract`, `name`, `version`).
   - Validate batch bounds: maximum trade count per batch ($\le 100$), gross value INR limit, and expiry timestamp ($t \le t_{\text{current}} + 300\text{s}$).
   - Compute Keccak-256 structured hash and execute ECDSA signature using the designated settlement relayer key handle inside the HSM.
9. Implement the **BIP-174 PSBT Multisig Signer**:
   - Parse partially signed Bitcoin transactions, validating input UTXOs, witness scripts, locktime, and change output destination addresses.
   - Check transaction fee rate against allowable network bounds (reject if fee $> 250$ sat/vB or $> 0.5\%$ of output value).
   - Sign designated input indexes using the custody root or derived key handle inside the HSM, attaching partial signatures to the PSBT.
10. Implement the **M-of-N Key Ceremony & Rotation Orchestrator**:
    - Manage ceremony state transitions: `INITIALIZED`, `OFFICERS_AUTHENTICATED`, `KEY_GENERATED`, `QUORUM_ATTESTED`, `ACTIVE`, `DEPRECATED`, `REVOKED`, `SHREDDED`.
    - Implement Shamir Secret Sharing (SSS) share verification over authenticated mTLS channels with FIDO2 / WebAuthn confirmation from authorized Crypto Officers.
    - Implement zero-downtime dual-key overlapping transition windows (new key signs new batches; old key remains active for verification during settlement finality).
    - Trigger cryptographic key destruction (`C_DestroyObject`) upon key revocation or lifecycle decommissioning.
11. Implement the **Mutual TLS (mTLS) Security & Policy Gate**:
    - Terminate mTLS on gRPC endpoints, extracting client Common Name (CN) and Subject Alternative Name (SAN).
    - Match calling microservice identity against granular key access policies (e.g., only `services/trade-settlement` can invoke EIP-712 settlement keys; only `services/besu-validator` can invoke QBFT keys).
    - Enforce Token Bucket rate limiting per key handle to prevent denial-of-service or signing exhaustion attacks.
12. Build the **SIEM & Audit Trail Pipeline**:
    - Stream real-time signing events to Kafka topic `crypto.signing.audit.v1` and ClickHouse audit tables.
    - Ensure zero sensitive key material or private seed data is ever logged or exported.
13. Configure Prometheus metrics (`hsm_session_pool_active`, `hsm_signing_duration_ms`, `hsm_signing_errors_total`, `anti_equivocation_violations_total`), OpenTelemetry tracing, and health check probes.

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/crypto_signer/v1/signer_daemon.proto`)
```protobuf
syntax = "proto3";

package growww.crypto_signer.v1;

option go_package = "github.com/growww/services/cloudhsm-signer-daemon/gen/v1;signerdaemonv1";

service CryptoSignerService {
  rpc SignQbftBlock (SignQbftBlockRequest) returns (SignQbftBlockResponse);
  rpc SignEip712Settlement (SignEip712SettlementRequest) returns (SignEip712SettlementResponse);
  rpc SignBip174Psbt (SignBip174PsbtRequest) returns (SignBip174PsbtResponse);
  rpc SignGenericDigest (SignGenericDigestRequest) returns (SignGenericDigestResponse);
  rpc InitiateKeyCeremony (InitiateKeyCeremonyRequest) returns (InitiateKeyCeremonyResponse);
  rpc SubmitCeremonyShare (SubmitCeremonyShareRequest) returns (SubmitCeremonyShareResponse);
  rpc FinalizeKeyRotation (FinalizeKeyRotationRequest) returns (FinalizeKeyRotationResponse);
  rpc GetKeyMetadata (GetKeyMetadataRequest) returns (GetKeyMetadataResponse);
  rpc ListActiveKeys (ListActiveKeysRequest) returns (ListActiveKeysResponse);
  rpc RevokeKey (RevokeKeyRequest) returns (RevokeKeyResponse);
}

enum KeyAlgorithm {
  KEY_ALGORITHM_UNSPECIFIED = 0;
  SECP256K1_ECDSA = 1;
  SECP256K1_SCHNORR = 2;
  ED25519 = 3;
  NIST_P384_ECDSA = 4;
  RSA_4096 = 5;
}

enum KeyUsageType {
  KEY_USAGE_UNSPECIFIED = 0;
  QBFT_VALIDATOR_CONSENSUS = 1;
  EIP712_SETTLEMENT_RELAYER = 2;
  BIP174_BITCOIN_CUSTODY_MULTISIG = 3;
  TRANSFER_COMPLIANCE_ATTESTATION = 4;
  PKI_INTERNAL_CA_SIGNING = 5;
  PROOF_OF_RESERVE_ORACLE = 6;
}

enum KeyLifecycleState {
  KEY_STATE_UNSPECIFIED = 0;
  PENDING_CEREMONY = 1;
  ACTIVE = 2;
  ROTATING = 3;
  DEPRECATED_VERIFY_ONLY = 4;
  REVOKED = 5;
  SHREDDED = 6;
}

enum QbftMessageType {
  QBFT_MSG_UNSPECIFIED = 0;
  PROPOSAL = 1;
  PREPARE = 2;
  COMMIT = 3;
  ROUND_CHANGE = 4;
}

enum CeremonyStatus {
  CEREMONY_STATUS_UNSPECIFIED = 0;
  INITIALIZED = 1;
  COLLECTING_SHARES = 2;
  QUORUM_REACHED = 3;
  COMPLETED = 4;
  ABORTED = 5;
}

message SignQbftBlockRequest {
  string key_id = 1;
  uint64 chain_id = 2;
  uint64 block_number = 3;
  uint32 sequence = 4;
  uint32 round = 5;
  QbftMessageType message_type = 6;
  bytes block_hash = 7; // 32-byte block proposal hash
  bytes payload_digest = 8; // 32-byte Keccak-256 consensus message hash
}

message SignQbftBlockResponse {
  string key_id = 1;
  bytes signature_bytes = 2; // 65-byte compact signature (r || s || v)
  bytes public_key = 3;
  int64 signed_at_utc = 4;
  uint32 signing_latency_micros = 5;
}

message Eip712DomainSpec {
  string name = 1;
  string version = 2;
  uint64 chain_id = 3;
  string verifying_contract = 4;
}

message SignEip712SettlementRequest {
  string key_id = 1;
  Eip712DomainSpec domain = 2;
  string batch_id = 3;
  bytes merkle_root = 4; // 32-byte batch trades Merkle root
  uint32 trade_count = 5;
  double gross_value_inr = 6;
  int64 settlement_window_expiry_utc = 7;
  bytes structured_data_digest = 8; // 32-byte EIP-712 final digest
}

message SignEip712SettlementResponse {
  string key_id = 1;
  string batch_id = 2;
  bytes signature_r = 3; // 32 bytes
  bytes signature_s = 4; // 32 bytes
  uint32 signature_v = 5; // 27 or 28 (or normalized 0/1)
  bytes compact_signature = 6; // 65 bytes
  int64 signed_at_utc = 7;
}

message SignBip174PsbtRequest {
  string key_id = 1;
  bytes raw_psbt_bytes = 2;
  repeated uint32 input_indexes_to_sign = 3;
  uint64 max_fee_satoshis = 4;
  uint32 max_fee_rate_sat_per_vb = 5;
  string change_address_expected = 6;
}

message SignBip174PsbtResponse {
  string key_id = 1;
  bytes signed_psbt_bytes = 2;
  uint32 signatures_added_count = 3;
  uint64 fee_satoshis = 4;
  int64 signed_at_utc = 5;
}

message SignGenericDigestRequest {
  string key_id = 1;
  KeyAlgorithm algorithm = 2;
  bytes digest = 3; // Must match algorithm hash length (32 bytes for SHA-256/Keccak-256, 48 bytes for SHA-384)
  string caller_service_id = 4;
  string request_purpose = 5;
}

message SignGenericDigestResponse {
  string key_id = 1;
  bytes signature = 2;
  bytes public_key = 3;
  int64 signed_at_utc = 4;
}

message InitiateKeyCeremonyRequest {
  string key_alias = 1;
  KeyAlgorithm algorithm = 2;
  KeyUsageType usage_type = 3;
  uint32 required_quorum_m = 4; // e.g. 3
  uint32 total_officers_n = 5; // e.g. 5
  repeated string authorized_officer_ids = 6;
  string mandate_reference = 7;
}

message InitiateKeyCeremonyResponse {
  string ceremony_id = 1;
  string key_alias = 2;
  CeremonyStatus status = 3;
  int64 expires_at_utc = 4;
}

message SubmitCeremonyShareRequest {
  string ceremony_id = 1;
  string officer_id = 2;
  string officer_fido2_signature = 3;
  bytes encrypted_share_payload = 4;
  string officer_public_key_thumbprint = 5;
}

message SubmitCeremonyShareResponse {
  string ceremony_id = 1;
  uint32 submitted_shares_count = 2;
  uint32 required_quorum_m = 3;
  CeremonyStatus status = 4;
}

message FinalizeKeyRotationRequest {
  string ceremony_id = 1;
  string existing_key_id = 2;
  int64 deprecation_grace_period_seconds = 3;
}

message FinalizeKeyRotationResponse {
  string new_key_id = 1;
  string previous_key_id = 2;
  bytes new_public_key = 3;
  string on_chain_registration_payload = 4;
  int64 activated_at_utc = 5;
  int64 previous_key_deprecated_at_utc = 6;
}

message GetKeyMetadataRequest {
  string key_id = 1;
}

message GetKeyMetadataResponse {
  string key_id = 1;
  string key_alias = 2;
  KeyAlgorithm algorithm = 3;
  KeyUsageType usage_type = 4;
  KeyLifecycleState state = 5;
  bytes public_key = 6;
  string public_address_hex = 7; // EVM address (0x...) or Bitcoin address
  string hsm_partition_id = 8;
  int64 created_at_utc = 9;
  int64 last_rotated_at_utc = 10;
  int64 expires_at_utc = 11;
  uint64 total_signatures_executed = 12;
}

message ListActiveKeysRequest {
  KeyUsageType filter_usage = 1;
}

message ListActiveKeysResponse {
  repeated GetKeyMetadataResponse keys = 1;
}

message RevokeKeyRequest {
  string key_id = 1;
  string revocation_reason = 2;
  repeated string approving_officer_ids = 3;
  repeated string approving_officer_signatures = 4;
  bool emergency_immediate_shred = 5;
}

message RevokeKeyResponse {
  string key_id = 1;
  KeyLifecycleState final_state = 2;
  bool is_hardware_shredded = 3;
  int64 revoked_at_utc = 4;
}
```

### PostgreSQL Database Schema
```sql
CREATE TYPE key_algorithm_enum AS ENUM (
    'SECP256K1_ECDSA',
    'SECP256K1_SCHNORR',
    'ED25519',
    'NIST_P384_ECDSA',
    'RSA_4096'
);

CREATE TYPE key_usage_type_enum AS ENUM (
    'QBFT_VALIDATOR_CONSENSUS',
    'EIP712_SETTLEMENT_RELAYER',
    'BIP174_BITCOIN_CUSTODY_MULTISIG',
    'TRANSFER_COMPLIANCE_ATTESTATION',
    'PKI_INTERNAL_CA_SIGNING',
    'PROOF_OF_RESERVE_ORACLE'
);

CREATE TYPE key_lifecycle_state_enum AS ENUM (
    'PENDING_CEREMONY',
    'ACTIVE',
    'ROTATING',
    'DEPRECATED_VERIFY_ONLY',
    'REVOKED',
    'SHREDDED'
);

CREATE TYPE ceremony_status_enum AS ENUM (
    'INITIALIZED',
    'COLLECTING_SHARES',
    'QUORUM_REACHED',
    'COMPLETED',
    'ABORTED'
);

CREATE TABLE hsm_clusters (
    cluster_id VARCHAR(64) PRIMARY KEY,
    cluster_name VARCHAR(128) NOT NULL,
    provider_type VARCHAR(32) NOT NULL, -- 'AWS_CLOUDHSM', 'THALES_LUNA'
    fips_certification_level VARCHAR(32) NOT NULL DEFAULT 'FIPS_140_3_LEVEL_3',
    primary_partition_id VARCHAR(64) NOT NULL,
    standby_partition_id VARCHAR(64),
    ip_address VARCHAR(45) NOT NULL,
    pkcs11_lib_path VARCHAR(256) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE hsm_key_registry (
    key_id VARCHAR(64) PRIMARY KEY, -- e.g. "key_secp256k1_qbft_val_01"
    key_alias VARCHAR(128) NOT NULL UNIQUE,
    cluster_id VARCHAR(64) NOT NULL REFERENCES hsm_clusters(cluster_id),
    hsm_object_handle BIGINT NOT NULL,
    hsm_key_label VARCHAR(128) NOT NULL,
    algorithm key_algorithm_enum NOT NULL,
    usage_type key_usage_type_enum NOT NULL,
    state key_lifecycle_state_enum NOT NULL DEFAULT 'ACTIVE',
    public_key_bytes BYTEA NOT NULL,
    public_address_hex VARCHAR(66), -- 0x... for EVM keys or base58/bech32 for Bitcoin
    x509_certificate_pem TEXT,
    key_version INT NOT NULL DEFAULT 1,
    required_quorum_m INT NOT NULL DEFAULT 1,
    total_officers_n INT NOT NULL DEFAULT 1,
    total_signatures_executed BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_rotated_at TIMESTAMPTZ,
    deprecated_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    shredded_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE qbft_equivocation_state (
    validator_address VARCHAR(42) NOT NULL,
    chain_id BIGINT NOT NULL,
    block_number BIGINT NOT NULL,
    sequence_number INT NOT NULL,
    round_number INT NOT NULL,
    message_type VARCHAR(32) NOT NULL,
    block_hash BYTEA NOT NULL,
    signed_digest BYTEA NOT NULL,
    signature BYTEA NOT NULL,
    signed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (validator_address, chain_id, block_number, sequence_number, round_number, message_type)
);

CREATE TABLE key_rotation_ceremonies (
    ceremony_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_alias VARCHAR(128) NOT NULL,
    existing_key_id VARCHAR(64) REFERENCES hsm_key_registry(key_id),
    new_key_id VARCHAR(64) REFERENCES hsm_key_registry(key_id),
    algorithm key_algorithm_enum NOT NULL,
    usage_type key_usage_type_enum NOT NULL,
    required_quorum_m INT NOT NULL,
    total_officers_n INT NOT NULL,
    mandate_reference VARCHAR(128) NOT NULL,
    status ceremony_status_enum NOT NULL DEFAULT 'INITIALIZED',
    shares_collected INT NOT NULL DEFAULT 0,
    initiated_by_officer_id UUID NOT NULL,
    finalized_by_officer_id UUID,
    on_chain_tx_hash VARCHAR(66),
    opened_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE key_ceremony_shares (
    share_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ceremony_id UUID NOT NULL REFERENCES key_rotation_ceremonies(ceremony_id) ON DELETE CASCADE,
    officer_id UUID NOT NULL,
    officer_public_key_thumbprint VARCHAR(64) NOT NULL,
    officer_fido2_signature TEXT NOT NULL,
    encrypted_share_hash CHAR(64) NOT NULL,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_ceremony_officer UNIQUE (ceremony_id, officer_id)
);

CREATE TABLE cryptographic_signing_audit_log (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id VARCHAR(64) NOT NULL REFERENCES hsm_key_registry(key_id),
    usage_type key_usage_type_enum NOT NULL,
    caller_service_name VARCHAR(64) NOT NULL,
    caller_client_ip VARCHAR(45) NOT NULL,
    caller_mtls_san VARCHAR(128) NOT NULL,
    payload_context JSONB NOT NULL, -- structured data metadata, block height, batch ID
    digest_hash CHAR(64) NOT NULL,
    signing_latency_micros INT NOT NULL,
    was_successful BOOLEAN NOT NULL DEFAULT TRUE,
    error_reason TEXT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hkr_state_usage ON hsm_key_registry (state, usage_type);
CREATE INDEX idx_hkr_alias ON hsm_key_registry (key_alias);
CREATE INDEX idx_krc_status ON key_rotation_ceremonies (status, opened_at DESC);
CREATE INDEX idx_csal_key_recorded ON cryptographic_signing_audit_log (key_id, recorded_at DESC);
CREATE INDEX idx_csal_caller ON cryptographic_signing_audit_log (caller_service_name, recorded_at DESC);
```

### Kafka Event Schemas

#### Topic: `crypto.signing.audit.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CryptoSigningAuditEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "key_id": { "type": "string" },
    "key_alias": { "type": "string" },
    "usage_type": {
      "type": "string",
      "enum": [
        "QBFT_VALIDATOR_CONSENSUS",
        "EIP712_SETTLEMENT_RELAYER",
        "BIP174_BITCOIN_CUSTODY_MULTISIG",
        "TRANSFER_COMPLIANCE_ATTESTATION",
        "PKI_INTERNAL_CA_SIGNING",
        "PROOF_OF_RESERVE_ORACLE"
      ]
    },
    "algorithm": {
      "type": "string",
      "enum": ["SECP256K1_ECDSA", "SECP256K1_SCHNORR", "ED25519", "NIST_P384_ECDSA", "RSA_4096"]
    },
    "caller_service_name": { "type": "string" },
    "caller_mtls_san": { "type": "string" },
    "digest_hash": { "type": "string", "minLength": 64, "maxLength": 64 },
    "signing_latency_micros": { "type": "integer", "minimum": 0 },
    "was_successful": { "type": "boolean" },
    "hsm_cluster_id": { "type": "string" },
    "timestamp_utc": { "type": "integer" }
  },
  "required": [
    "event_id",
    "key_id",
    "usage_type",
    "algorithm",
    "caller_service_name",
    "digest_hash",
    "signing_latency_micros",
    "was_successful",
    "hsm_cluster_id",
    "timestamp_utc"
  ]
}
```

#### Topic: `crypto.key_lifecycle.event.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CryptoKeyLifecycleEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "key_id": { "type": "string" },
    "key_alias": { "type": "string" },
    "lifecycle_action": {
      "type": "string",
      "enum": ["GENERATED", "ACTIVATED", "ROTATED", "DEPRECATED", "REVOKED", "SHREDDED"]
    },
    "algorithm": { "type": "string" },
    "usage_type": { "type": "string" },
    "public_key_hex": { "type": "string" },
    "public_address_hex": { "type": "string" },
    "key_version": { "type": "integer", "minimum": 1 },
    "ceremony_id": { "type": "string", "format": "uuid" },
    "on_chain_tx_hash": { "type": "string" },
    "timestamp_utc": { "type": "integer" }
  },
  "required": [
    "event_id",
    "key_id",
    "key_alias",
    "lifecycle_action",
    "algorithm",
    "usage_type",
    "public_key_hex",
    "key_version",
    "timestamp_utc"
  ]
}
```

#### Topic: `crypto.equivocation_attempt.flagged.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CryptoEquivocationAttemptFlaggedEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "validator_address": { "type": "string" },
    "chain_id": { "type": "integer" },
    "block_number": { "type": "integer" },
    "sequence_number": { "type": "integer" },
    "round_number": { "type": "integer" },
    "message_type": { "type": "string" },
    "existing_block_hash": { "type": "string" },
    "conflicting_block_hash": { "type": "string" },
    "caller_service": { "type": "string" },
    "action_taken": { "type": "string", "enum": ["SIGNING_REJECTED_CIRCUIT_BROKEN"] },
    "timestamp_utc": { "type": "integer" }
  },
  "required": [
    "event_id",
    "validator_address",
    "chain_id",
    "block_number",
    "sequence_number",
    "round_number",
    "message_type",
    "existing_block_hash",
    "conflicting_block_hash",
    "action_taken",
    "timestamp_utc"
  ]
}
```

## Security & Compliance Notes
- **FIPS 140-3 Level 3 Hardware Boundary Enforcement:** All physical and logical cryptographic boundaries must conform to FIPS 140-3 Level 3 (or FIPS 140-3 Level 3). Private keys are generated using hardware True Random Number Generators (TRNGs) inside physical tamper-responsive enclosures. Key attributes strictly prohibit plaintext extraction (`CKA_EXTRACTABLE = FALSE`, `CKA_SENSITIVE = TRUE`).
- **Anti-Equivocation & Double-Signing Prevention:** The daemon enforces an immutable write-ahead log on local NVMe storage before executing any consensus block signature. Any attempt by an upstream Besu node or attacker to sign two distinct block proposals or commits for the identical `(chain_id, block_number, round)` tuple is rejected instantly at the hardware gateway, mathematically preventing validator slashing and network forks.
- **M-of-N Threshold Ceremony Controls:** Key generation, key activation, and key destruction require multi-party authorization with an M-of-N threshold (minimum 3 of 5 Crypto Officers). Each officer must authenticate using FIDO2 / WebAuthn hardware security keys over dedicated mutual TLS control channels.
- **Zero Plaintext Memory Footprint:** Cryptographic operations occur strictly within the internal RAM and cryptographic acceleration units of the HSM. The daemon host process receives only input digests and output signatures; private key scalar bytes are never loaded into host RAM or persistent swap files.
- **SEBI CSCRF & RBI Key Rotation Mandate:** In compliance with SEBI CSCRF Annexure B and RBI Cyber Security Guidelines, all operational transaction relayer and CA keys must undergo scheduled rotation every 90 days. Deprecated keys enter a temporary verify-only grace period before permanent cryptographic shredding (`C_DestroyObject`).
- **Mutual TLS & Zero-Trust Access Policy:** All incoming gRPC connections must present valid X.509 client certificates signed by the internal Platform Root CA. Granular authorization policies restrict key usage so that settlement keys cannot be invoked by consensus nodes, and validator keys cannot be invoked by settlement services.
- **Zero PII Leakage:** Public keys, addresses, and signature digests contain zero personal identifying information. All operator interactions and officer identities are referenced solely by UUIDs and cryptographic thumbprints.

## Acceptance Criteria
- [ ] Direct hardware integration with AWS CloudHSM / Thales Luna HSM clusters via PKCS#11 v2.40 completes initialization and health checks across active partitions.
- [ ] Generates secp256k1, Ed25519, NIST P-384, and RSA-4096 keys inside the HSM with hardware-enforced `CKA_EXTRACTABLE = FALSE` and `CKA_SENSITIVE = TRUE`.
- [ ] QBFT validator block signing (`SignQbftBlock`) achieves p99 latency $< 5\text{ms}$ under a continuous load of 500 requests per second.
- [ ] Anti-equivocation state engine successfully blocks 100% of simulated conflicting block proposal and commit signing attempts for identical `(chain_id, block_number, round)` tuples, immediately emitting `crypto.equivocation_attempt.flagged.v1`.
- [ ] EIP-712 structured data signing (`SignEip712Settlement`) strictly validates domain parameters, batch size limits ($\le 100$), and expiry windows before generating valid ECDSA `(v, r, s)` signatures.
- [ ] BIP-174 PSBT multisig signing (`SignBip174Psbt`) correctly verifies input witness UTXOs, enforces fee rate bounds ($\le 250$ sat/vB), and appends valid signatures to partially signed Bitcoin transactions.
- [ ] M-of-N key ceremony orchestrator enforces strict 3-of-5 officer quorum with FIDO2 authentication before transitioning new keys to `ACTIVE` state.
- [ ] Zero-downtime key rotation pipeline maintains old key handles in `DEPRECATED_VERIFY_ONLY` mode during transition windows without interrupting active settlement flows.
- [ ] Revocation workflow executes physical hardware key destruction via `C_DestroyObject`, transitioning database records to `SHREDDED`.
- [ ] 100% of signing operations, ceremony events, and equivocation alerts stream to Kafka topic `crypto.signing.audit.v1` and ClickHouse audit tables.
- [ ] Microservice unit, integration, and mock PKCS#11 test suites achieve $\ge 90\%$ code coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design Standards), Prompt 104 (Kafka Standards), Prompt 105 (Auth & IAM Architecture), Prompt 109 (Secrets Management), Prompt 707 (Data Encryption & Key Rotation).
- **Parallel Tasks:** Prompt 208 (Trade Settlement Service), Prompt 213 (Custodian Depository Integration), Prompt 237 (Multichain MPC TSS Vault Custody Service), Prompt 702 (IAM & RBAC Least Privilege).
- **Downstream Dependents:** Prompt 301 (Permissioned Consortium Ledger Architecture), Prompt 305 (Transfer Compliance Hooks), Prompt 310 (Chain Node Monitoring & Alerting), Prompt 311 (Validator Key Management & Node Slashing Defense), Prompt 319 (Institutional Custody Bridge), Prompt 329 (NBSE Settlement DvP & Fee Collector).
