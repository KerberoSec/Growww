# 237 - Institutional Multi-Chain MPC-TSS Vault & Custody Service (Rust / C++ / CloudHSM / GG20 & FROST)

## Purpose
In a 24/7 institutional equity clearing, digital asset custody, and multi-chain settlement ecosystem, private key management represents the single most critical security vulnerability. Traditional single-key architectures (where a master private key resides in a single database, server memory, or even a single Hardware Security Module) create an existential single point of failure (SPOF) vulnerable to internal collusion, key theft, or infrastructure compromise. Multi-signature smart contracts partially alleviate this risk on EVM networks but introduce substantial gas overhead, lack multi-chain interoperability (e.g., across Ed25519 and non-EVM chains), and expose governance policy structures publicly on-chain.

The **Institutional Multi-Chain MPC-TSS Vault & Custody Service** eliminates single points of failure by implementing an enterprise-grade Multi-Party Computation Threshold Signature Scheme (MPC-TSS). By utilizing GG20/CGGMP21 for ECDSA Secp256k1 (EVM, Bitcoin, Hyperledger Besu) and FROST (Flexible Round-Optimized Schnorr Threshold) for EdDSA Ed25519 (Solana, Aptos, Polkadot, Algorand, Cosmos), the service allows a distributed set of 5 institutional validator nodes to collectively generate keys, maintain isolated key shares, and produce valid cryptographic signatures via a 3-of-5 threshold quorum.

Crucially, the complete private key is **never assembled, combined, or materialized in memory anywhere** during its entire lifecycle (neither at key generation, nor during signing, nor during storage). Each key share is isolated within FIPS 140-2/3 Level 3 Hardware Security Modules (AWS CloudHSM / Nitro Enclaves), undergoes proactive secret sharing (PSS) offline key refreshes to invalidate compromised historical shares without changing the public address, and is governed by a deterministic, programmatic withdrawal policy engine enforcing multi-signatory quorums, velocity limits, time locks, address whitelisting, and statutory Travel Rule compliance.

## What You Are Building
An ultra-secure, memory-safe, low-latency institutional microservice in **Rust** (with performance-critical constant-time cryptographic extensions in **C++20**) deployed at `services/mpc-tss-vault-service`. Concrete deliverables include:
- **Distributed Key Generation (DKG) Protocol Engine:** Multi-party distributed generation of Secp256k1 and Ed25519 keypairs across 5 institutional nodes using Feldman's / Pedersen's Verifiable Secret Sharing (VSS), producing a shared public address without any node ever knowing the full master private key.
- **3-of-5 Threshold Co-Signing Coordinator & Signer:** High-throughput, round-optimized threshold signing engine executing GG20/CGGMP21 (for ECDSA) and FROST (for Ed25519) signing rounds across authenticated peer nodes within strict sub-second SLAs.
- **Hardware Security Module (CloudHSM) Key Share Wrapper:** Zero-trust cryptographic interface isolating participant key shares inside AWS CloudHSM / Azure Dedicated HSM / AWS Nitro Enclaves via PKCS#11, ensuring key shares are encrypted under hardware root keys and never exposed in plaintext host memory.
- **Proactive Secret Sharing (PSS) & Offline Key Refresh Worker:** Automated and manual key refresh daemon that re-randomizes polynomial shares across the 5 nodes without changing the master public key or on-chain address, rendering old or leaked shares completely useless.
- **Programmatic Withdrawal Policy Engine:** Strict pre-signing rules gate evaluating transaction intent, multi-signatory operational quorums, cumulative 24-hour velocity caps, cooling-off time locks, whitelisted destination addresses, and AML/Travel Rule attestations before releasing co-signing commitments.
- **Internal gRPC Vault Gateway & Kafka Audit Stream:** High-performance internal RPC interface (`MpcVaultService`) for signing orchestration, accompanied by an immutable Kafka event stream recording all DKG events, round aborts, co-signing approvals, and signature dispatches.

## Scope Boundaries
- **In Scope:**
  - Multi-party distributed key generation (DKG) protocols for Secp256k1 and Ed25519.
  - 3-of-5 threshold co-signing round execution, round serialization, zero-knowledge proof verification, and signature reconstruction.
  - FIPS 140-2/3 Level 3 CloudHSM and Nitro Enclave integration for participant key share encryption at rest and in memory.
  - Proactive secret sharing (PSS) offline key refresh cycles.
  - Identifiable abort protocols detecting and isolating malicious, unresponsive, or faulty participant nodes.
  - Programmatic withdrawal policy evaluation (velocity limits, dual control, time locks, address whitelisting, Travel Rule checks).
  - Outbox pattern event broadcasting for vault lifecycle events to Apache Kafka.
- **Out of Scope / Handled Elsewhere:**
  - Retail investor session authentication and WebAuthn/MPIN verification (handled in Prompt 201).
  - End-user INR fiat payment collection and UPI/NEFT rail integrations (handled in Prompt 212 / 232).
  - Solidity smart contract logic for settlement DvP on Hyperledger Besu (handled in Prompt 306).
  - Pre-trade VaR and margin adequacy checking (handled in Prompt 229).
  - On-chain gas fee relayer and broadcast transaction submission to public blockchains (handled in Prompt 319).

## Technology to Use
- **Primary Languages & Systems:** **Rust (1.78+)** for memory safety, concurrency, and zero-cost abstractions, combined with audited **C++20** constant-time cryptographic primitives.
- **MPC-TSS Cryptographic Frameworks:**
  - `cggmp21` / `curv` / `k256` / `libsecp256k1` for Secp256k1 threshold ECDSA (GG20 and CGGMP21 protocols).
  - `frost-dalek` / `ed25519-dalek` for Ed25519 threshold EdDSA (FROST protocol, RFC 9591).
  - Constant-time arithmetic and side-channel resistant BigNum libraries (`subtle`, `num-bigint-dig`).
- **Hardware Security & Secure Enclaves:** AWS CloudHSM / AWS Nitro Enclaves / PKCS#11 (`cryptoki-rs`) for hardware-backed participant key share envelope encryption and ephemeral signing operations.
- **Inter-Node P2P & Transport Security:** `tonic` (gRPC over HTTP/2) with mutual TLS (mTLS 1.3), strict cipher suites (`TLS_AES_256_GCM_SHA384`), client certificate pinning, and Noise Protocol Framework for direct peer-to-peer encrypted MPC round message exchanges.
- **Transactional & Audit Store:** **PostgreSQL 16+** with `sqlx` (async Rust driver) for vault account metadata, policy configs, active session states, and immutable cryptographic audit logs.
- **Distributed State & In-Memory Coordination:** **Redis 7.2** with Redlock for sub-millisecond signing round state coordination, distributed mutex locking, and replay nonce deduplication.
- **Event Streaming & Message Bus:** **Apache Kafka** with `rdkafka` for asynchronous vault event publication and audit integration.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `vault_accounts`, `mpc_dkg_sessions`, `mpc_key_shares`, `mpc_signing_sessions`, `withdrawal_policies`, `vault_audit_logs`.
- **Redis 7.2 Keys:** `mpc:session:{session_id}:state`, `mpc:velocity:{vault_id}:24h_spent`, `mpc:nonce:lock:{vault_id}:{chain_id}`.
- **Apache Kafka Topics:**
  - Consumes: `vault.signing_requested.v1`, `vault.policy_override_requested.v1`.
  - Publishes: `vault.dkg_completed.v1`, `vault.signature_generated.v1`, `vault.signing_failed.v1`, `vault.key_refreshed.v1`, `vault.security_alert.v1`.
- **Upstream Callers:**
  - `services/settlement-service` (Prompt 208): Requests MPC signatures for clearinghouse relayer transactions on Hyperledger Besu.
  - `services/custodian-depository-integration` (Prompt 213): Requests cold/warm vault signatures for institutional depository asset transfers.
  - `services/settlement-guarantee-fund-service` (Prompt 230): Requests multi-sig signatures for SGF collateral rebalancing.
- **Downstream Consumers:**
  - `services/admin-back-office` (Prompt 217): Compliance portal for dual-control signing approvals.
  - `services/audit-log-service` (Prompt 218): Ingests cryptographic audit trails of all signing sessions.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Standard Account Abstraction & Full Node Compatibility:** The MPC-TSS vault outputs standard, aggregate ECDSA (Secp256k1) and EdDSA (Ed25519) signatures. To Hyperledger Besu (and other connected blockchains), the resulting signature is indistinguishable from a signature generated by a conventional single-key wallet. Hyperledger Besu verifies transactions using native `ecrecover` EVM operations without requiring custom on-chain multi-sig contracts, saving gas and maintaining 100% standard tooling compatibility.
- **Custody Pool & Relayer Address Control:** MPC-derived addresses control Growww's primary institutional liquidity pools, cold/warm asset vaults, and settlement relayer accounts on Hyperledger Besu.
- **1:1 Custody Backing Invariant:** Vault transactions authorizing token minting, burning, or high-value DvP settlements (`SettlementDvP.sol`, `DigitalSecurityToken.sol`) execute only when validated against physical depository receipts (NSDL/CDSL).
- **Zero On-Chain PII Invariant:** In compliance with DPDP Act 2023 and SEBI cybersecurity frameworks, raw investor identity data is never contained in signing payloads. Transaction hashes and parameters contain strictly pseudonymized addresses, token contract references, and cryptographic nonces.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Rust cargo workspace `services/mpc-tss-vault-service` with micro-crates (`mpc-core`, `mpc-dkg`, `mpc-signing`, `mpc-hsm`, `mpc-policy`, `mpc-server`), configuring strict compiler flags (`-D warnings`), AddressSanitizer, and constant-time execution linters.
2. **Define Protobuf Specifications:** Author `proto/growww/vault/v1/mpc_vault.proto` specifying `InitiateDkg`, `InitiateThresholdSigning`, `RefreshKeyShares`, `EvaluateWithdrawalPolicy`, `GetVaultAccount`, and `GetSigningSessionStatus`.
3. **Generate gRPC Stubs:** Compile Protobuf contracts into type-safe Rust server and client bindings using `tonic-build` and `prost`.
4. **Design PostgreSQL Schema:** Write database migration scripts creating `vault_accounts`, `mpc_dkg_sessions`, `mpc_key_shares`, `mpc_signing_sessions`, `withdrawal_policies`, and `vault_audit_logs`.
5. **Implement Hardware Security Module (CloudHSM) Wrapper:** Build the `mpc-hsm` module interfacing via PKCS#11 with AWS CloudHSM / AWS Nitro Enclaves to encrypt, store, and decrypt participant key shares using hardware-protected Wrapping Keys (AES-256-GCM / RSA-4096-OAEP).
6. **Implement Multi-Party DKG Protocol for Secp256k1 (GG20/CGGMP21):** Implement distributed key generation across 5 institutional participant nodes:
   - Round 1: Exchange Pedersen/Feldman commitments and ZK-proofs of knowledge.
   - Round 2: Distribute Shamir secret polynomial evaluations over encrypted point-to-point channels.
   - Round 3: Verify consistency, aggregate public key $Y = \sum y_i$, compute public verification vector, and store encrypted local share $x_i$ into CloudHSM.
7. **Implement Multi-Party DKG Protocol for Ed25519 (FROST):** Implement round-optimized FROST distributed key generation:
   - Generate Schnorr proofs of knowledge for secret coefficients.
   - Execute coefficient commitment broadcast and secret share distribution.
   - Compute aggregate group public key and individual participant verification shares.
8. **Establish Secure Multi-Party P2P Transport Layer:** Configure mTLS 1.3 peer connections between the 5 institutional validator nodes (Growww Custody, Custodian Bank, Depository Trustee, Independent Auditor, Regulatory Escrow) with strict certificate pinning and Noise Protocol encryption for intra-round communication.
9. **Implement 3-of-5 Threshold Co-Signing Coordinator (GG20/CGGMP21):** Implement multi-round threshold co-signing:
   - Pre-Signing Round: Sample ephemeral secret shares, compute public ephemeral commitments, and execute zero-knowledge range proofs.
   - Signing Round: Exchange partial signature responses $s_i$, verify non-interactive zero-knowledge proofs (NIZK), and aggregate into final standard ECDSA signature $(r, s, v)$.
   - Identifiable Abort: If any node submits an invalid proof or times out, identify the offending participant node, log the cryptographic fault to PostgreSQL, and abort the session.
10. **Implement 3-of-5 FROST Threshold Co-Signing (Ed25519):** Implement 2-round FROST threshold signing:
    - Round 1 (Commitment): Participants publish nonces and binding commitments.
    - Round 2 (Sign): Coordinator aggregates commitments, constructs challenge hash, participants respond with individual signature shares, coordinator verifies and combines into valid Ed25519 signature $(R, S)$.
11. **Implement Proactive Secret Sharing (PSS) Offline Key Refresh:** Build key refresh protocol allowing 5 nodes to compute new polynomial shares of zero $\sum f_i(z) = 0$ and update local shares $x'_i = x_i + \sum f_j(i)$, refreshing secret shares without altering the group public key.
12. **Implement Programmatic Withdrawal Policy Engine:** Build deterministic rule validation pipeline evaluating:
    - Multi-Signatory Quorum: Verification of $M$-of-$N$ human operator approvals via WebAuthn/FIDO2 digital signatures.
    - 24-Hour Cumulative Velocity Limits: Enforcing daily withdrawal thresholds per vault account against PostgreSQL/Redis.
    - Whitelisted Address Enforcement: Restricting destinations strictly to verified depository and settlement pool addresses.
    - Cooldown Time Locks: Enforcing mandatory 60-minute holding delays for high-value cross-entity transactions.
    - Statutory Travel Rule & Sanctions Check: Validating counterparty VASP attestation payloads before initiating signing.
13. **Build Redis Session & Nonce Replay Manager:** Implement distributed locking for signing session concurrency, preventing double-signing attacks, nonce reuse, and race conditions across distributed cluster instances.
14. **Implement Kafka Event Publisher & Transactional Outbox:** Stream all signing session lifecycle events (`vault.dkg_completed.v1`, `vault.signature_generated.v1`, `vault.signing_failed.v1`, `vault.key_refreshed.v1`) using transactional outbox guarantees.
15. **Configure Telemetry, Metrics & Cryptographic Test Suite:** Export Prometheus metrics (`mpc_signing_duration_seconds`, `mpc_dkg_round_latency`, `mpc_aborts_total`, `mpc_policy_violations_total`) and author comprehensive test suites simulating malicious nodes, network latency, packet loss, and Byzantine share corruption.

## Interfaces / Contracts

### Protobuf Definition (`mpc_vault.proto`)
```protobuf
syntax = "proto3";

package growww.vault.v1;

option go_package = "growww/vault/v1;vaultv1";

service MpcVaultService {
  rpc InitiateDkg (InitiateDkgRequest) returns (InitiateDkgResponse);
  rpc InitiateThresholdSigning (InitiateThresholdSigningRequest) returns (InitiateThresholdSigningResponse);
  rpc RefreshKeyShares (RefreshKeySharesRequest) returns (RefreshKeySharesResponse);
  rpc EvaluateWithdrawalPolicy (EvaluateWithdrawalPolicyRequest) returns (EvaluateWithdrawalPolicyResponse);
  rpc GetVaultAccount (GetVaultAccountRequest) returns (GetVaultAccountResponse);
  rpc GetSigningSessionStatus (GetSigningSessionStatusRequest) returns (GetSigningSessionStatusResponse);
}

enum CurveType {
  CURVE_TYPE_UNSPECIFIED = 0;
  CURVE_TYPE_SECP256K1 = 1; // GG20 / CGGMP21 ECDSA (EVM / Besu / BTC)
  CURVE_TYPE_ED25519 = 2;   // FROST EdDSA (Solana / Aptos / Cosmos)
}

enum VaultPurpose {
  VAULT_PURPOSE_UNSPECIFIED = 0;
  VAULT_PURPOSE_SETTLEMENT_RELAYER = 1;
  VAULT_PURPOSE_CUSTODY_HOT = 2;
  VAULT_PURPOSE_CUSTODY_WARM = 3;
  VAULT_PURPOSE_CUSTODY_COLD = 4;
  VAULT_PURPOSE_SGF_COLLATERAL = 5;
}

enum SigningStatus {
  SIGNING_STATUS_UNSPECIFIED = 0;
  SIGNING_STATUS_PENDING_POLICY = 1;
  SIGNING_STATUS_POLICY_APPROVED = 2;
  SIGNING_STATUS_ROUND_1_IN_PROGRESS = 3;
  SIGNING_STATUS_ROUND_2_IN_PROGRESS = 4;
  SIGNING_STATUS_ROUND_3_IN_PROGRESS = 5;
  SIGNING_STATUS_SIGNATURE_READY = 6;
  SIGNING_STATUS_POLICY_REJECTED = 7;
  SIGNING_STATUS_ABORTED_FAULT = 8;
  SIGNING_STATUS_TIMED_OUT = 9;
}

message ParticipantNode {
  uint32 node_id = 1;
  string party_identifier = 2; // "GROWWW_NODE_1", "CUSTODIAN_BANK_NODE", "DEPOSITORY_TRUSTEE_NODE", "AUDITOR_NODE", "REGULATORY_ESCROW_NODE"
  string endpoint_url = 3;
  string tls_cert_fingerprint = 4;
}

message InitiateDkgRequest {
  string dkg_session_id = 1;
  CurveType curve = 2;
  VaultPurpose purpose = 3;
  uint32 threshold = 4; // 3
  uint32 total_parties = 5; // 5
  repeated ParticipantNode participants = 6;
  string initiated_by = 7;
}

message InitiateDkgResponse {
  string dkg_session_id = 1;
  string vault_id = 2;
  string group_public_key_hex = 3;
  string derived_address = 4;
  CurveType curve = 5;
  int64 completed_at_unix_ns = 6;
}

message InitiateThresholdSigningRequest {
  string signing_session_id = 1;
  string vault_id = 2;
  string transaction_payload_hash_hex = 3; // 32-byte SHA256 / Keccak256 hash
  string destination_address = 4;
  string asset_identifier = 5; // Token contract address or currency code
  string amount_raw = 6; // Decimal string with full unit precision
  string memo = 7;
  repeated string operator_approvals = 8; // Dual-control FIDO2/WebAuthn signatures
  string idempotency_key = 9;
}

message InitiateThresholdSigningResponse {
  string signing_session_id = 1;
  SigningStatus status = 2;
  string signature_hex = 3; // DER format or (r, s, v) / (R, S)
  string raw_r = 4;
  string raw_s = 5;
  uint32 recovery_id = 6;
  int64 signed_at_unix_ns = 7;
}

message RefreshKeySharesRequest {
  string refresh_session_id = 1;
  string vault_id = 2;
  repeated uint32 participating_node_ids = 3;
  string authorization_proof = 4;
}

message RefreshKeySharesResponse {
  string refresh_session_id = 1;
  string vault_id = 2;
  uint32 new_share_version = 3;
  bool success = 4;
  int64 refreshed_at_unix_ns = 5;
}

message EvaluateWithdrawalPolicyRequest {
  string vault_id = 1;
  string destination_address = 2;
  string asset_identifier = 3;
  string amount_raw = 4;
  repeated string operator_signatures = 5;
}

message EvaluateWithdrawalPolicyResponse {
  string vault_id = 1;
  bool is_allowed = 2;
  string rejection_reason = 3;
  bool requires_time_lock = 4;
  int64 cooldown_until_unix_ns = 5;
  uint32 required_approvals = 6;
  uint32 current_approvals = 7;
}

message GetVaultAccountRequest {
  string vault_id = 1;
}

message GetVaultAccountResponse {
  string vault_id = 2;
  CurveType curve = 3;
  VaultPurpose purpose = 4;
  string group_public_key_hex = 5;
  string derived_address = 6;
  uint32 threshold = 7;
  uint32 total_parties = 8;
  uint32 share_version = 9;
  bool is_active = 10;
  int64 created_at_unix_ns = 11;
}

message GetSigningSessionStatusRequest {
  string signing_session_id = 1;
}

message GetSigningSessionStatusResponse {
  string signing_session_id = 1;
  string vault_id = 2;
  SigningStatus status = 3;
  repeated uint32 participating_nodes = 4;
  repeated uint32 faulted_nodes = 5;
  string error_message = 6;
  string signature_hex = 7;
  int64 updated_at_unix_ns = 8;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE curve_type AS ENUM ('SECP256K1', 'ED25519');
CREATE TYPE vault_purpose AS ENUM ('SETTLEMENT_RELAYER', 'CUSTODY_HOT', 'CUSTODY_WARM', 'CUSTODY_COLD', 'SGF_COLLATERAL');
CREATE TYPE signing_status AS ENUM (
    'PENDING_POLICY',
    'POLICY_APPROVED',
    'ROUND_1_IN_PROGRESS',
    'ROUND_2_IN_PROGRESS',
    'ROUND_3_IN_PROGRESS',
    'SIGNATURE_READY',
    'POLICY_REJECTED',
    'ABORTED_FAULT',
    'TIMED_OUT'
);

CREATE TABLE vault_accounts (
    vault_id VARCHAR(64) PRIMARY KEY,
    curve curve_type NOT NULL,
    purpose vault_purpose NOT NULL,
    group_public_key_hex VARCHAR(130) NOT NULL UNIQUE,
    derived_address VARCHAR(128) NOT NULL UNIQUE,
    threshold_k SMALLINT NOT NULL DEFAULT 3,
    total_parties_n SMALLINT NOT NULL DEFAULT 5,
    share_version INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mpc_dkg_sessions (
    dkg_session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_accounts(vault_id),
    curve curve_type NOT NULL,
    threshold_k SMALLINT NOT NULL,
    total_parties_n SMALLINT NOT NULL,
    participating_nodes JSONB NOT NULL,
    session_status VARCHAR(32) NOT NULL DEFAULT 'INITIALIZED',
    public_verification_vector JSONB,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mpc_key_shares (
    share_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_accounts(vault_id),
    node_id SMALLINT NOT NULL,
    share_version INT NOT NULL DEFAULT 1,
    hsm_key_handle VARCHAR(128) NOT NULL,
    hsm_wrapped_share BYTEA NOT NULL,
    public_share_commitment_hex VARCHAR(130) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (vault_id, node_id, share_version)
);

CREATE TABLE withdrawal_policies (
    policy_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_accounts(vault_id),
    max_single_tx_amount_inr NUMERIC(24, 4) NOT NULL,
    velocity_24h_cap_inr NUMERIC(24, 4) NOT NULL,
    time_lock_cooldown_seconds INT NOT NULL DEFAULT 0,
    required_signatories_count SMALLINT NOT NULL DEFAULT 2,
    authorized_signatory_keys JSONB NOT NULL,
    whitelisted_addresses JSONB NOT NULL,
    enforce_travel_rule BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE mpc_signing_sessions (
    signing_session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_accounts(vault_id),
    transaction_hash_hex VARCHAR(66) NOT NULL,
    destination_address VARCHAR(128) NOT NULL,
    asset_identifier VARCHAR(128) NOT NULL,
    amount_raw NUMERIC(36, 0) NOT NULL,
    status signing_status NOT NULL DEFAULT 'PENDING_POLICY',
    participating_node_ids JSONB NOT NULL,
    faulted_node_ids JSONB,
    raw_signature_hex VARCHAR(260),
    raw_r VARCHAR(66),
    raw_s VARCHAR(66),
    recovery_id SMALLINT,
    rejection_reason TEXT,
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    signed_at TIMESTAMPTZ
);

CREATE TABLE vault_audit_logs (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id VARCHAR(64) NOT NULL REFERENCES vault_accounts(vault_id),
    event_type VARCHAR(64) NOT NULL,
    actor_identifier VARCHAR(128) NOT NULL,
    event_details JSONB NOT NULL,
    cryptographic_digest VARCHAR(66) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vault_accounts_address ON vault_accounts(derived_address);
CREATE INDEX idx_mpc_signing_vault_status ON mpc_signing_sessions(vault_id, status);
CREATE INDEX idx_mpc_signing_idempotency ON mpc_signing_sessions(idempotency_key);
CREATE INDEX idx_mpc_shares_vault_node ON mpc_key_shares(vault_id, node_id, share_version);
CREATE INDEX idx_vault_audit_vault_time ON vault_audit_logs(vault_id, created_at DESC);
```

## Security & Compliance Notes
- **FIPS 140-2/3 Level 3 HSM Key Share Custody:** Individual participant key shares ($x_i$) are encrypted using hardware Wrapping Keys inside dedicated CloudHSM / Nitro Enclave boundaries. Plaintext key shares are never stored in databases, written to disk, or transmitted over the network.
- **Identifiable Abort & Byzantine Fault Tolerance:** The MPC-TSS engine implements zero-knowledge proof verification at every protocol round. If a Byzantine or corrupted node transmits an invalid commitment, falsified partial signature, or forged range proof, the protocol uniquely identifies the malicious party, logs the cryptographically signed fault receipt, and immediately aborts without leaking secret share material.
- **Geographic & Legal Separation of Validator Nodes:** The 5 validator nodes are partitioned across 5 distinct institutional entities with independent cloud accounts and data centers:
  1. Growww Internal Custody Node (AWS India).
  2. Institutional Sponsor Bank Custody Node (Azure India).
  3. Registered Depository Trustee Node (Google Cloud India).
  4. Independent Auditing Firm Attestation Node (Dedicated On-Premise CloudHSM).
  5. Regulatory Escrow Node (GIFT City IFSC Gateway).
  A minimum of 3 independent institutional parties must cryptographically participate to authorize and sign any transaction.
- **Proactive Secret Sharing (PSS) Defense Against Mobile Adversaries:** Automated monthly key refresh protocols re-randomize polynomial shares across all nodes. Even if an adversary compromises a key share from node A in month 1, and node B in month 3, the refreshed shares cannot be combined because they belong to mathematically independent polynomial sharing instances.
- **Zero On-Chain PII & DPDP Act 2023 Compliance:** All MPC-TSS vault transactions reference strictly pseudonymized cryptographic addresses and asset identifiers. Investor personal data is segregated off-chain and protected via envelope encryption.

## Acceptance Criteria
- [ ] DKG protocol generates valid Secp256k1 (GG20/CGGMP21) and Ed25519 (FROST) group public keys across 5 validator nodes with zero master key reconstruction.
- [ ] 3-of-5 threshold co-signing generates valid, broadcast-ready standard signatures in $< 300\text{ms}$ for FROST and $< 800\text{ms}$ for GG20/CGGMP21 under normal network conditions.
- [ ] CloudHSM integration successfully encrypts and decrypts participant key shares via PKCS#11 with zero plaintext share leakage to host system memory.
- [ ] Proactive Secret Sharing (PSS) key refresh cycle successfully re-randomizes participant shares while keeping the master public address unchanged.
- [ ] Programmatic withdrawal policy engine rejects transactions exceeding 24-hour velocity limits, transactions targeting non-whitelisted addresses, or transactions lacking dual-control operator signatures.
- [ ] Identifiable abort mechanism accurately detects intentional partial signature corruption and reports the exact malicious node ID without protocol deadlock.
- [ ] Replay protection prevents duplicate signing sessions for identical `idempotency_key` or `transaction_payload_hash_hex`.
- [ ] All DKG and co-signing operations emit structured audit telemetry to Apache Kafka topic `vault.signature_generated.v1` and PostgreSQL table `vault_audit_logs`.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Standards), Prompt `105` (Authentication & Authorization Architecture), Prompt `109` (Secrets Management Architecture), Prompt `311` (Validator Key Management HSM).
- **Parallel Tasks:** Prompt `208` (Trade Settlement Service), Prompt `213` (Custodian Depository Integration Service), Prompt `230` (Settlement Guarantee Fund Service).
- **Downstream Blockers:** Prompt `306` (Settlement DvP Smart Contract relayer execution), Prompt `319` (Institutional Custody Bridge), Prompt `321` (Mainnet Genesis Ceremony & Validator Onboarding).
