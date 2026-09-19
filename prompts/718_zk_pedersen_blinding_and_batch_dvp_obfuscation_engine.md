# 718 - ZK Pedersen Blinding and Batch DvP Obfuscation Engine (DPDP Act 2023, GDPR, Homomorphic Sums, Timing-Graph De-anonymization Prevention)

## Purpose
Establishes an institutional-grade privacy, cryptographic commitment, and transaction graph obfuscation microservice (`services/privacy-obfuscation-service`). The engine delivers mathematical confidentiality and timing-decorrelation guarantees across all trade clearing, settlement, and reserve notarization pipelines within the Growww ecosystem.

The microservice addresses two critical operational and regulatory requirements:
1. **Zero-Knowledge Proof-of-Reserve Notarization via Pedersen Commitments:** Protects institutional and retail investor privacy by generating homomorphically summable, perfectly hiding Pedersen commitments ($C = g^v \cdot h^r$) over elliptic curve groups (BN254 / BabyJubjub). This allows the platform to cryptographically prove total asset solvency to public verifiers and regulators without exposing individual account balances, fractional positions, or wallet distribution curves.
2. **Multi-Trade Atomic Batch DvP Settlement Obfuscation:** Mitigates timing analysis, transaction graph heuristics, and counterparty profiling attacks on the Hyperledger Besu consortium ledger. By grouping individual matched trade executions into atomic multi-trade Delivery versus Payment (DvP) batches with non-deterministic time-jitter windows, Fisher-Yates permutation shuffling, and zero-net-delta synthetic padding trades, the engine ensures strict compliance with the **Digital Personal Data Protection Act (DPDP Act 2023)** and **General Data Protection Regulation (GDPR Article 25 & 32)**.

## What You Are Building
A high-throughput, fault-tolerant privacy and compliance microservice (`services/privacy-obfuscation-service`) comprising:
- **Pedersen Commitment Blinding Engine:** Computes information-theoretically hiding and computationally binding Pedersen commitments over the BN254 / BabyJubjub elliptic curve with hardware-backed blinding entropy generation.
- **Homomorphic Solvency Aggregator:** Provides zero-knowledge verifiable balance summation across user accounts, generating root blinding values and sum commitments for the Sparse Merkle Sum Tree (SMST) store.
- **Atomic Batch DvP Settlement Obfuscation Coordinator:** Gathers matched trade executions from the Order Matching Engine and Trade Settlement Service, aggregating them into atomic multi-trade batches enforcing a minimum k-anonymity threshold ($k \ge 10$) and participant entropy constraints.
- **Dynamic Timing-Jitter and Graph-Decorrelation Scheduler:** Introduces bounded Poisson delay jitter (50ms to 250ms) and time-bucket quantization to neutralize ledger transaction graph analysis, side-channel snooping, and mempool correlation vectors.
- **Synthetic Padding & Volume Masking Generator:** Automatically injects zero-net-delta liquidity padding transactions during low-liquidity market regimes to maintain deterministic batch sizes and prevent statistical trade-volume fingerprinting.
- **DPDP & GDPR Compliance Escrow Gateway:** Implements verifiable split-key threshold escrow (M-of-N secret sharing) for regulatory auditability, permitting authorized legal disclosure of de-obfuscated transaction graphs to SEBI / FIU-IND under strict court order without compromising the public ledger privacy envelope.
- **Tamper-Evident Privacy Audit Relayer:** Emits cryptographically signed attestation manifests, commitment verification logs, and anonymization entropy metrics to Kafka and ClickHouse.

## Scope Boundaries
- **In Scope:**
  - Computation and verification of elliptic curve Pedersen commitments ($C = v \cdot G + r \cdot H$).
  - Generation and secure storage of high-entropy blinding factors ($r \in \mathbb{F}_q$) with envelope encryption.
  - Homomorphic addition and subtraction mechanics for balance and position delta commitments.
  - Multi-trade batch DvP grouping, topological shuffling, and netting matrix formulation.
  - Dynamic batch windowing, Poisson timing jitter, and k-anonymity enforcement algorithms.
  - Synthetic zero-net-delta trade padding generation for volume normalization.
  - Cryptographic threshold key escrow generation for DPDP Act 2023 / GDPR lawful disclosure.
  - Synchronous gRPC API and asynchronous Kafka event consumers/producers.
  - PostgreSQL 16 schema for blinding factor vaults, batch manifests, and audit trails.
- **Out of Scope / Handled Elsewhere:**
  - Core order matching and continuous limit order book execution (Prompt 205).
  - High-level trade clearing and gross-to-net settlement scheduling (Prompt 208).
  - Physical depository custody ingestion from NSDL/CDSL (Prompt 213).
  - Sparse Merkle Sum Tree persistent storage in RocksDB (Prompt 409).
  - Physical CloudHSM PKCS#11 key lifecycle daemon (Prompt 717).
  - Smart contract DvP atomic execution logic in `NBSESettlementDvP.sol` (Prompt 329).
  - Frontend investor balance proof verifier UI (Prompts 514, 605).

## Technology to Use
- **Core Microservice:** Rust 1.78+ using `arkworks-rs` (`ark-bn254`, `ark-ec`, `ark-ff`) and `subtle` for constant-time cryptographic primitives; Go 1.22+ for high-concurrency batch orchestration workers.
- **Cryptographic Curves & Primitives:** BN254 (alt_bn128) and BabyJubjub curves; Pedersen commitment generators $G, H$ where $H = \text{MapToCurve}(\text{Keccak256}(G))$ ensuring unknown discrete logarithm $\log_G(H)$; ChaCha20-Poly1305 for blinding factor vault envelope encryption.
- **In-Memory Buffer & Window Staging:** Redis 7.2 Enterprise Cluster with Redis Streams and sub-millisecond sorted sets for low-latency trade accumulation and k-anonymity bucket evaluation.
- **Relational Metadata Database:** PostgreSQL 16 with Row-Level Security (RLS), pgcrypto, and table partitioning for encrypted blinding factor storage and batch manifests.
- **Audit Logging & Telemetry:** ClickHouse 24+ for append-only anonymization audit logs; Prometheus and OpenTelemetry for timing-jitter and batch entropy tracking.
- **Messaging Pipeline:** Apache Kafka (librdkafka) for streaming trade execution events and publishing obfuscated settlement batches.
- **Transport Security:** gRPC over HTTP/2 with mutual TLS (mTLS) enforcing TLS 1.3, AES-256-GCM, and X.509 client certificate pinning.
- **Justification:** Rust guarantees memory safety and constant-time execution for zero-knowledge arithmetic, preventing cache-timing side channels. Redis Streams combined with Go concurrency provides sub-millisecond batch buffering and topological trade shuffling at scale.

## Backend / Infra Touchpoints
- **Upstream Microservices & Callers:**
  - `services/order-matching-engine` (Prompt 205): Forwards executed trade matches (`settlement.trade_matched.v1`) to the batch obfuscation queue.
  - `services/trade-settlement` (Prompt 208): Requests obfuscated batch settlement bundles for clearing cycles.
  - `services/por-smt-generator` (Prompt 409): Dispatches user balance snapshots to generate blinded leaf commitments for Sparse Merkle Sum Trees.
  - `services/wallet-account-service` (Prompt 203): Dispatches wallet balance delta events requiring homomorphic commitment updates.
- **Downstream Systems & Infrastructure:**
  - `services/cloudhsm-signer-daemon` (Prompt 717): Signs obfuscated batch settlement payloads using EIP-712 structured data hashes.
  - `NBSESettlementDvP.sol` (Prompt 329): On-chain smart contract executing the atomic multi-trade batch on Hyperledger Besu.
  - `ProofOfReserveRegistry.sol` (Prompt 007): Receives aggregate blinded balance commitments for on-chain solvency validation.
  - PostgreSQL 16, Redis 7.2, and ClickHouse 24+ data stores.
- **Messaging Topics (Apache Kafka):**
  - Consumes:
    - `settlement.trade_matched.v1`: Ingests real-time matched trade execution records.
    - `por.epoch_snapshot.initiated.v1`: Ingests account balance snapshots for daily PoR generation.
    - `compliance.escrow_reveal.requested.v1`: Ingests authorized regulatory de-anonymization requests.
  - Publishes:
    - `settlement.batch_dvp_obfuscated.v1`: Emits fully shuffled, padded, and aggregated batch DvP settlement bundles.
    - `por.pedersen_commitments.generated.v1`: Emits blinded account commitments and aggregate tree sums.
    - `privacy.graph_decorrelation.audit.v1`: Emits anonymization entropy metrics and batch timing statistics.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating under QBFT consensus.
- **Atomic Batch DvP On-Chain Settlement:**
  - Individual trade records are never posted to the ledger as discrete transactions. Instead, the obfuscation engine aggregates $N$ individual trades into a single multi-party transaction submitted to `NBSESettlementDvP.sol`.
  - The transaction payload contains the netted token and INR balance transfers across all participants, accompanied by an obfuscated Merkle root and an EIP-712 batch attestation signature.
  - To outside observers and consortium nodes, the transaction appears as a single atomic state transition touching multiple accounts simultaneously, rendering heuristic transaction-graph reconstruction mathematically intractable.
- **Blinded Balance Notarization on `ProofOfReserveRegistry.sol`:**
  - Daily Proof-of-Reserve attestations publish the aggregate blinded balance commitment $C_{total} = \sum C_i = (\sum v_i) \cdot G + (\sum r_i) \cdot H$.
  - The smart contract verifies that $C_{total}$ equals the homomorphic product of all sub-tree commitments and matches physical custodian depository holdings without learning any individual investor balance.
- **Zero-PII Ledger Invariant:**
  - No investor names, Permanent Account Numbers (PAN), Aadhaar references, bank account details, or raw client identifiers are committed to blockchain state.
  - All on-chain accounts use deterministic pseudonymized cryptographic addresses derived through one-way hashing with platform-level master salts.

## Cryptographic Architecture & Mathematical Mechanics

### 1. Pedersen Commitment Scheme Formulation
The service operates over a prime-order elliptic curve group $\mathbb{G}$ of order $q$ with base generator points $G, H \in \mathbb{G}$.

#### Generator Point Derivation (Nothing-Up-My-Sleeve):
To ensure computational binding, the discrete logarithm $\gamma = \log_G(H)$ must be unknown. $H$ is deterministically generated using a verifiable hash-to-curve algorithm:
$$H = \text{MapToCurve}(\text{Keccak256}(\text{EncodePoint}(G) \parallel \text{"GROWWW_POR_PEDERSEN_H_GENERATOR_V1"}))$$

#### Commitment Construction:
For an investor holding balance value $v \in \mathbb{Z}_q$, the service samples a cryptographically secure random blinding factor $r \xleftarrow{\$} \mathbb{Z}_q$ and computes:
$$C(v, r) = v \cdot G + r \cdot H$$

#### Homomorphic Addition Property:
Given two commitments $C(v_1, r_1) = v_1 \cdot G + r_1 \cdot H$ and $C(v_2, r_2) = v_2 \cdot G + r_2 \cdot H$:
$$C(v_1 + v_2, r_1 + r_2) = C(v_1, r_1) + C(v_2, r_2) = (v_1 + v_2) \cdot G + (r_1 + r_2) \cdot H$$

#### Homomorphic Subtraction Property:
For balance debit mutations:
$$C(v_1 - v_2, r_1 - r_2) = C(v_1, r_1) - C(v_2, r_2) = (v_1 - v_2) \cdot G + (r_1 - r_2) \cdot H$$

#### Solvency Conservation Invariant:
For an account balance update involving transfer of amount $\Delta v$ from account $A$ to account $B$, with newly chosen blinding factors $r_A', r_B'$:
$$C(v_A', r_A') + C(v_B', r_B') - (C(v_A, r_A) + C(v_B, r_B)) = 0 \cdot G + (r_A' + r_B' - r_A - r_B) \cdot H$$

### 2. Multi-Trade Batch DvP Settlement Obfuscation Mechanics

#### k-Anonymity Settlement Threshold:
A settlement batch $\mathcal{B}$ is declared ready for dispatch only when satisfying the strict k-anonymity constraints:
$$\text{Size}(\mathcal{B}) \ge k_{\min} \quad (k_{\min} = 10)$$
$$\text{DistinctAccounts}(\mathcal{B}) \ge m_{\min} \quad (m_{\min} = 6)$$
$$\text{Entropy}(\mathcal{B}) = -\sum_{i=1}^{M} p_i \log_2(p_i) \ge E_{\text{threshold}} \quad (E_{\text{threshold}} = 2.5 \text{ bits})$$
where $p_i$ is the proportion of total batch trade volume attributed to participant $i$.

#### Dynamic Poisson Timing Jitter:
To break microsecond-level timing correlations between order matching events and ledger block inclusion, the batch dispatcher schedules dispatch time $t_{\text{dispatch}}$ as:
$$t_{\text{dispatch}} = t_{\text{batch\_ready}} + \Delta t_{\text{jitter}}$$
where $\Delta t_{\text{jitter}} \sim \text{Poisson}(\lambda)$ truncated to $[50\text{ms}, 250\text{ms}]$.

#### Topological Fisher-Yates Permutation:
Within each batch $\mathcal{B}$, trades are topologically flattened into atomic debit and credit vector entries. The execution order is permuted using a cryptographically secure Fisher-Yates shuffle seeded with hardware entropy:
$$\vec{T}_{\text{obfuscated}} = \text{Permute}(\vec{T}_{\text{raw}}, \text{CSPRNG\_Seed})$$

#### Zero-Net-Delta Synthetic Padding:
When a settlement window expires ($T_{\text{window\_max}} = 500\text{ms}$) but $\text{Size}(\mathcal{B}) < k_{\min}$, the engine automatically injects $P$ synthetic internal liquidity balancing trades between platform treasury sub-accounts:
$$\sum_{j=1}^{P} \Delta \vec{v}_{\text{synthetic}, j} = \vec{0}$$
$$\sum_{j=1}^{P} \Delta \text{Cash}_{\text{synthetic}, j} = 0$$
Synthetic trades normalize batch size to exactly $k_{\min}$ while contributing zero net change to external positions.

### 3. Regulatory Split-Key Escrow Mechanics (DPDP Act 2023 / GDPR)
To comply with legal disclosure mandates under DPDP Act 2023 Section 8(5) and regulatory audits:
- The blinding factor $r$ for every commitment is encrypted under a $(t, n)$ threshold ElGamal / ECIES public key shared among designated Compliance Officers and Judicial Escrow Agents.
- The ciphertext $E(r) = \text{Encrypt}(r, \text{PK}_{\text{escrow}})$ is stored alongside the commitment metadata.
- Reconstruction of $r$ requires at least $t$ of $n$ Crypto Officers to provide verifiable partial decryptions, strictly governed by smart contract multi-sig approval and physical M-of-N hardware token authentication.

## Step-by-Step Build Instructions (10-15 steps)
1. **Repository & Workspace Initialization:** Initialize the `services/privacy-obfuscation-service` microservice under the standard monorepo workspace following Prompt 106 guidelines, configuring Rust toolchain 1.78+ and Go 1.22+.
2. **Cryptographic Primitives Implementation:** Implement the BN254 / BabyJubjub curve wrappers, constant-time point multiplication, nothing-up-my-sleeve generator derivation for $G$ and $H$, and ChaCha20-Poly1305 envelope encryption in `src/crypto/pedersen.rs`.
3. **Database Schema Deployment:** Author and execute PostgreSQL 16 migrations for blinding factor vaults, batch settlement manifests, k-anonymity configuration parameters, and regulatory escrow audit logs.
4. **Blinding Factor Vault Engine:** Implement secure storage and retrieval for account blinding factors with local AES-256-GCM caching, auto-expiring memory allocations, and AWS KMS / CloudHSM key-encryption-key (KEK) wrapping.
5. **Homomorphic Operations Module:** Build the commitment aggregator supporting additive and subtractive operations, batch verification routines, and Sparse Merkle Sum Tree leaf commitment generation in `src/crypto/homomorphic.rs`.
6. **Redis Stream Ingestion Worker:** Implement the high-concurrency Go worker consuming matched trade execution events from Kafka topic `settlement.trade_matched.v1` and buffering them into Redis sorted sets indexed by security ISIN and execution timestamp.
7. **k-Anonymity & Entropy Evaluator:** Build the batch qualification evaluator verifying participant count, volume distribution entropy, and minimum trade thresholds before releasing candidate settlement batches.
8. **Poisson Timing-Jitter Scheduler:** Implement the non-deterministic delay queue introducing truncated Poisson jitter (50ms to 250ms) to decorrelate match timing from batch submission.
9. **Topological Shuffler & Netting Engine:** Build the batch transformation pipeline executing Fisher-Yates shuffling over atomic balance mutations and generating the aggregated netting settlement matrix.
10. **Synthetic Padding Generator:** Implement the zero-net-delta trade injection engine that synthesizes self-canceling treasury liquidity pairs when market volume is insufficient to satisfy k-anonymity within the maximum window timeout.
11. **Regulatory Escrow & Threshold Decryption Engine:** Implement the ECIES / ElGamal M-of-N threshold key escrow encryption module and verifiable partial decryption request coordinator.
12. **gRPC Server Implementation:** Implement the synchronous gRPC service handling client commitment generation, sum verification, and batch obfuscation requests according to `proto/growww/privacy_obfuscation/v1/privacy_obfuscation.proto`.
13. **Kafka Event Producers & Relayers:** Build the transactional Kafka producers publishing obfuscated settlement batches (`settlement.batch_dvp_obfuscated.v1`) and audit records (`privacy.graph_decorrelation.audit.v1`).
14. **Integration & Benchmarking Suite:** Execute comprehensive unit, property-based, and load tests validating cryptographic correctness, zero-knowledge conservation laws, constant-time execution, and sub-10ms batch throughput.

## Interfaces / Contracts

### 1. Protobuf Interface Specification (`proto/growww/privacy_obfuscation/v1/privacy_obfuscation.proto`)

```protobuf
syntax = "proto3";

package growww.privacy_obfuscation.v1;

option go_package = "github.com/growww/backend/proto/privacy_obfuscation/v1;privacyobfuscationv1";

service PrivacyObfuscationService {
  rpc GeneratePedersenCommitment(GeneratePedersenCommitmentRequest) returns (GeneratePedersenCommitmentResponse);
  rpc BatchGenerateCommitments(BatchGenerateCommitmentsRequest) returns (BatchGenerateCommitmentsResponse);
  rpc VerifyCommitmentSum(VerifyCommitmentSumRequest) returns (VerifyCommitmentSumResponse);
  rpc ObfuscateSettlementBatch(ObfuscateSettlementBatchRequest) returns (ObfuscateSettlementBatchResponse);
  rpc RequestRegulatoryEscrowReveal(RegulatoryEscrowRevealRequest) returns (RegulatoryEscrowRevealResponse);
  rpc GetBatchObfuscationMetrics(GetBatchObfuscationMetricsRequest) returns (GetBatchObfuscationMetricsResponse);
}

message EllipticCurvePoint {
  bytes x = 1;
  bytes y = 2;
  string curve_name = 3;
}

message GeneratePedersenCommitmentRequest {
  string account_id = 1;
  string asset_isin = 2;
  uint64 value = 3;
  uint64 epoch_id = 4;
  bool persist_blinding_factor = 5;
}

message GeneratePedersenCommitmentResponse {
  string commitment_id = 1;
  EllipticCurvePoint commitment = 2;
  bytes commitment_bytes = 3;
  uint64 epoch_id = 4;
  int64 created_at_ns = 5;
}

message AccountBalanceEntry {
  string account_id = 1;
  string asset_isin = 2;
  uint64 value = 3;
}

message BatchGenerateCommitmentsRequest {
  uint64 epoch_id = 1;
  repeated AccountBalanceEntry entries = 2;
}

message BlindedLeafEntry {
  string account_id = 1;
  string asset_isin = 2;
  EllipticCurvePoint commitment = 3;
  bytes commitment_bytes = 4;
}

message BatchGenerateCommitmentsResponse {
  uint64 epoch_id = 1;
  repeated BlindedLeafEntry blinded_leaves = 2;
  EllipticCurvePoint aggregate_commitment = 3;
  bytes aggregate_commitment_bytes = 4;
  uint64 total_value = 5;
}

message VerifyCommitmentSumRequest {
  repeated bytes input_commitment_bytes = 1;
  bytes expected_sum_commitment_bytes = 2;
  string curve_name = 3;
}

message VerifyCommitmentSumResponse {
  bool is_valid = 1;
  string error_message = 2;
}

message RawTradeExecution {
  string trade_id = 1;
  string buyer_account_id = 2;
  string seller_account_id = 3;
  string asset_isin = 4;
  uint64 quantity = 5;
  uint64 price_paise = 6;
  uint64 gross_amount_paise = 7;
  int64 matched_at_ns = 8;
}

message ObfuscateSettlementBatchRequest {
  string batch_id = 1;
  repeated RawTradeExecution trades = 2;
  uint32 min_k_anonymity = 3;
  uint32 max_window_delay_ms = 4;
  bool allow_synthetic_padding = 5;
}

message ObfuscatedTransfer {
  string transfer_id = 1;
  string pseudonymized_account_id = 2;
  string asset_isin = 3;
  int64 net_quantity_delta = 4;
  int64 net_cash_delta_paise = 5;
  bytes commitment_bytes = 6;
}

message ObfuscateSettlementBatchResponse {
  string batch_id = 1;
  string obfuscated_batch_id = 2;
  repeated ObfuscatedTransfer transfers = 3;
  bytes batch_merkle_root = 4;
  uint32 original_trade_count = 5;
  uint32 synthetic_trade_count = 6;
  uint32 participant_count = 7;
  double calculated_entropy = 8;
  int64 scheduled_dispatch_timestamp_ns = 9;
  bytes batch_attestation_signature = 10;
}

message RegulatoryEscrowRevealRequest {
  string warrant_reference_id = 1;
  string requesting_authority = 2;
  string target_account_id = 3;
  uint64 epoch_id_start = 4;
  uint64 epoch_id_end = 5;
  repeated bytes officer_authorization_signatures = 6;
}

message RevealedBlindingFactorEntry {
  uint64 epoch_id = 1;
  string asset_isin = 2;
  bytes blinding_factor = 3;
  uint64 balance_value = 4;
  bytes commitment_bytes = 5;
}

message RegulatoryEscrowRevealResponse {
  string reveal_session_id = 1;
  string target_account_id = 2;
  repeated RevealedBlindingFactorEntry entries = 3;
  int64 revealed_at_ns = 4;
  bytes audit_log_signature = 5;
}

message GetBatchObfuscationMetricsRequest {
  int64 start_time_ms = 1;
  int64 end_time_ms = 2;
}

message GetBatchObfuscationMetricsResponse {
  uint64 total_batches_processed = 1;
  uint64 total_trades_obfuscated = 2;
  uint64 total_synthetic_trades_injected = 3;
  double average_entropy_score = 4;
  double average_jitter_delay_ms = 5;
  double p99_jitter_delay_ms = 6;
}
```

### 2. PostgreSQL 16 Database Schema (`migrations/V718__privacy_obfuscation_init.sql`)

```sql
-- Schema: privacy_obfuscation
CREATE SCHEMA IF NOT EXISTS privacy_obfuscation;

-- Table: pedersen_generator_registry
CREATE TABLE privacy_obfuscation.pedersen_generator_registry (
    curve_name VARCHAR(32) NOT NULL,
    generator_label VARCHAR(64) NOT NULL,
    point_g_x BYTEA NOT NULL,
    point_g_y BYTEA NOT NULL,
    point_h_x BYTEA NOT NULL,
    point_h_y BYTEA NOT NULL,
    seed_specification TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (curve_name, generator_label)
);

-- Table: blinding_factor_vault (Encrypted at Rest with Envelope Encryption)
CREATE TABLE privacy_obfuscation.blinding_factor_vault (
    vault_entry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(64) NOT NULL,
    asset_isin VARCHAR(12) NOT NULL,
    epoch_id BIGINT NOT NULL,
    commitment_bytes BYTEA NOT NULL,
    encrypted_blinding_factor BYTEA NOT NULL,
    blinding_factor_nonce BYTEA NOT NULL,
    key_derivation_tag BYTEA NOT NULL,
    escrow_ciphertext BYTEA NOT NULL,
    balance_value NUMERIC(28, 8) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_blinding_account_asset_epoch UNIQUE (account_id, asset_isin, epoch_id)
) PARTITION BY RANGE (epoch_id);

CREATE INDEX idx_blinding_vault_account ON privacy_obfuscation.blinding_factor_vault (account_id, epoch_id);
CREATE INDEX idx_blinding_vault_commitment ON privacy_obfuscation.blinding_factor_vault (commitment_bytes);

-- Table: batch_dvp_records
CREATE TABLE privacy_obfuscation.batch_dvp_records (
    batch_id VARCHAR(64) PRIMARY KEY,
    obfuscated_batch_id VARCHAR(64) NOT NULL UNIQUE,
    batch_merkle_root BYTEA NOT NULL,
    original_trade_count INT NOT NULL,
    synthetic_trade_count INT NOT NULL DEFAULT 0,
    participant_count INT NOT NULL,
    entropy_score NUMERIC(6, 4) NOT NULL,
    jitter_delay_ms INT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    besu_tx_hash VARCHAR(66),
    scheduled_dispatch_at TIMESTAMPTZ NOT NULL,
    dispatched_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_batch_dvp_status_dispatch ON privacy_obfuscation.batch_dvp_records (status, scheduled_dispatch_at);
CREATE INDEX idx_batch_dvp_merkle_root ON privacy_obfuscation.batch_dvp_records (batch_merkle_root);

-- Table: batch_trade_memberships
CREATE TABLE privacy_obfuscation.batch_trade_memberships (
    membership_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id VARCHAR(64) NOT NULL REFERENCES privacy_obfuscation.batch_dvp_records(batch_id),
    trade_id VARCHAR(64) NOT NULL,
    buyer_account_id VARCHAR(64) NOT NULL,
    seller_account_id VARCHAR(64) NOT NULL,
    asset_isin VARCHAR(12) NOT NULL,
    quantity NUMERIC(18, 4) NOT NULL,
    price_paise BIGINT NOT NULL,
    is_synthetic BOOLEAN NOT NULL DEFAULT FALSE,
    shuffled_position INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_batch_trade_membership UNIQUE (batch_id, trade_id)
);

CREATE INDEX idx_memberships_batch_pos ON privacy_obfuscation.batch_trade_memberships (batch_id, shuffled_position);

-- Table: regulatory_escrow_audit_log
CREATE TABLE privacy_obfuscation.regulatory_escrow_audit_log (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    warrant_reference_id VARCHAR(128) NOT NULL,
    requesting_authority VARCHAR(128) NOT NULL,
    target_account_id VARCHAR(64) NOT NULL,
    epoch_id_start BIGINT NOT NULL,
    epoch_id_end BIGINT NOT NULL,
    revealed_record_count INT NOT NULL,
    authorization_digest BYTEA NOT NULL,
    executed_by_officer_id VARCHAR(64) NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_regulatory_audit_warrant ON privacy_obfuscation.regulatory_escrow_audit_log (warrant_reference_id);
```

### 3. Kafka Event Schemas

#### Event: `settlement.batch_dvp_obfuscated.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SettlementBatchDvpObfuscatedEvent",
  "type": "object",
  "required": [
    "eventId",
    "batchId",
    "obfuscatedBatchId",
    "merkleRoot",
    "netTransfers",
    "originalTradeCount",
    "syntheticTradeCount",
    "participantCount",
    "entropyScore",
    "timestampNs"
  ],
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "batchId": { "type": "string" },
    "obfuscatedBatchId": { "type": "string" },
    "merkleRoot": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "netTransfers": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["pseudonymizedAccountId", "assetIsin", "netQuantityDelta", "netCashDeltaPaise", "commitmentHex"],
        "properties": {
          "pseudonymizedAccountId": { "type": "string" },
          "assetIsin": { "type": "string", "minLength": 12, "maxLength": 12 },
          "netQuantityDelta": { "type": "string" },
          "netCashDeltaPaise": { "type": "string" },
          "commitmentHex": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64,128}$" }
        }
      }
    },
    "originalTradeCount": { "type": "integer", "minimum": 1 },
    "syntheticTradeCount": { "type": "integer", "minimum": 0 },
    "participantCount": { "type": "integer", "minimum": 1 },
    "entropyScore": { "type": "number", "minimum": 0.0 },
    "timestampNs": { "type": "integer" }
  }
}
```

#### Event: `por.pedersen_commitments.generated.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "PedersenCommitmentsGeneratedEvent",
  "type": "object",
  "required": [
    "eventId",
    "epochId",
    "assetIsin",
    "leafCount",
    "aggregateCommitmentHex",
    "totalValue",
    "timestampNs"
  ],
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "epochId": { "type": "integer" },
    "assetIsin": { "type": "string", "minLength": 12, "maxLength": 12 },
    "leafCount": { "type": "integer", "minimum": 1 },
    "aggregateCommitmentHex": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64,128}$" },
    "totalValue": { "type": "string" },
    "timestampNs": { "type": "integer" }
  }
}
```

### 4. Solidity Interface Specifications (`contracts/interfaces/IPrivacyObfuscationHook.sol`)

```solidity
// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

/**
 * @title IPrivacyObfuscationHook
 * @notice Interface for on-chain verification of obfuscated batch settlement commitments and PoR sums.
 */
interface IPrivacyObfuscationHook {
    struct CurvePoint {
        uint256 x;
        uint256 y;
    }

    struct ObfuscatedBatchHeader {
        bytes32 batchId;
        bytes32 merkleRoot;
        uint32 transferCount;
        uint32 originalTradeCount;
        uint32 syntheticTradeCount;
        uint64 settlementWindowExpiry;
    }

    struct NetBalanceTransfer {
        address account;
        bytes32 assetIsinHash;
        int256 netQuantity;
        int256 netCashPaise;
        CurvePoint balanceCommitment;
    }

    event BatchSettlementExecuted(
        bytes32 indexed batchId,
        bytes32 indexed merkleRoot,
        uint32 transferCount,
        uint32 originalTradeCount,
        uint32 syntheticTradeCount
    );

    event AggregateReserveCommitmentNotarized(
        uint256 indexed epochId,
        bytes32 indexed assetIsinHash,
        CurvePoint aggregateCommitment,
        uint256 totalSharesBacking
    );

    function executeObfuscatedBatchDvP(
        ObfuscatedBatchHeader calldata header,
        NetBalanceTransfer[] calldata transfers,
        bytes calldata relayerSignature
    ) external returns (bool success);

    function verifyAndRecordProofOfReserveCommitment(
        uint256 epochId,
        bytes32 assetIsinHash,
        CurvePoint calldata aggregateCommitment,
        uint256 totalDepositoryShares,
        bytes calldata custodianSignature,
        bytes calldata auditorSignature
    ) external returns (bool isValid);
}
```

## Security & Compliance Notes

### 1. DPDP Act 2023 Compliance
- **Data Minimization (Section 6 & 8):** Individual order routing timestamps, IP addresses, and personal investor identifiers are never broadcast to the consortium blockchain or public APIs.
- **Privacy by Design (Section 8(5)):** All investor account balances are committed via Pedersen blinding before participating in Sparse Merkle Tree aggregation.
- **Lawful Disclosure Procedures:** Split-key threshold escrow ensures that regulatory disclosures under valid judicial warrant require M-of-N executive cryptographic authorization, preventing rogue operator de-anonymization.

### 2. GDPR Compliance (Articles 25 & 32)
- **Data Protection by Default:** Settlement transactions are batched, shuffled, and jittered to prevent graph profiling and deanonymization of European investors trading on the GIFT City International Financial Services Centre (IFSC) exchange.
- **Security of Processing (Article 32):** Cryptographic keys and blinding factors are encrypted using ChaCha20-Poly1305 with keys derived from dedicated HSM partitions (`CKA_EXTRACTABLE = FALSE`).

### 3. Cryptographic Attack Mitigations
- **Constant-Time Execution:** All curve arithmetic operations use constant-time scalar multiplication routines from `arkworks-rs` and `subtle`, eliminating cache-timing side channels.
- **Generator Isolation:** The discrete logarithm between generator points $G$ and $H$ is provably unknown through verifiable nothing-up-my-sleeve point generation, preventing malicious binding forgery.
- **Front-Running & MEV Prevention:** In the permissioned QBFT network, batch Merkle roots and Fisher-Yates shuffled transfer orderings prevent node operators from prioritizing or sandwiching individual matched trades.

## Acceptance Criteria
- [ ] Microservice `services/privacy-obfuscation-service` compiles without warnings on Rust 1.78+ and Go 1.22+.
- [ ] Pedersen commitment generator points $G$ and $H$ are verifiably derived with unknown discrete logarithm $\log_G(H)$.
- [ ] Homomorphic addition and subtraction properties satisfy $C(v_1 + v_2, r_1 + r_2) = C(v_1, r_1) + C(v_2, r_2)$ across $1,000,000$ randomized test vectors.
- [ ] Batch DvP obfuscator enforces minimum k-anonymity ($k \ge 10$) and minimum participant entropy ($E \ge 2.5$ bits) before dispatch.
- [ ] Poisson timing jitter introduces bounded delays strictly within the range of $50\text{ms}$ to $250\text{ms}$.
- [ ] Synthetic padding trade injection correctly outputs zero-net-delta balance changes for asset quantities and cash amounts.
- [ ] Blinding factors are envelope-encrypted at rest and accessible only through authenticated KMS/CloudHSM role tokens.
- [ ] Threshold escrow reveal requires valid M-of-N multi-officer signatures and logs an immutable audit event in PostgreSQL and ClickHouse.
- [ ] Synchronous gRPC API sustains $>5,000$ commitment generations per second per CPU core with p99 latency $<2\text{ms}$.
- [ ] Zero Personal Identifiable Information (PII) is emitted to Kafka settlement topics or Hyperledger Besu smart contracts.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero application implementation code.
- [ ] Full specification strictly uses standard ASCII hyphens with zero em dashes or en dashes.

## Suggested Order / Dependencies
1. **Prerequisite Foundation:**
   - Prompt 007 (Proof-of-Reserve Public Disclosure Specification)
   - Prompt 008 (Data Protection and Privacy Policy)
   - Prompt 106 (Monorepo Layout & Coding Standards)
2. **Upstream Services:**
   - Prompt 205 (Order Matching Engine)
   - Prompt 208 (Trade Settlement Service)
   - Prompt 409 (Proof-of-Reserve Sparse Merkle Tree Store)
   - Prompt 717 (HSM & CloudHSM Key Lifecycle and Signing Daemon)
3. **Downstream Dependencies:**
   - Prompt 329 (NBSE Settlement DvP and Fee Collector Smart Contract)
   - Prompt 215 (Reconciliation Service)
   - Prompt 216 (Regulatory Reporting Service)
   - Prompt 605 (Public Transparency & PoR Verifier Portal)
