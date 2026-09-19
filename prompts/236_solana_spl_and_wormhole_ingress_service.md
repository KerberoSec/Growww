# 236 - Solana SPL and Wormhole Ingress Service (Rust / Go / gRPC)

## Purpose
Democratizing access to Indian equities and tokenized real-world assets through GIFT City IFSCA and international institutional capital channels requires high-throughput, low-latency inbound liquidity corridors. The Solana blockchain network operates as a high-speed execution layer with sub-second slot finality and deep global stablecoin liquidity (USDC, USDT). In addition, cross-chain messaging protocols such as Wormhole Native Token Transfers (NTT) and Pyth Network oracle data bridges enable seamless cross-chain collateral attestation, fiat-backed stablecoin deposits, and real-time verifiable settlement.

However, ingesting deposits and cross-chain messages from high-throughput public blockchains into a regulated clearing and settlement system introduces significant financial, technical, and operational risks:
1. **Slot Reorganization and Fork Invalidation:** Solana consensus (Tower BFT and Proof of History) undergoes micro-forks and optimistic confirmation changes before reaching irreversible finality. Crediting balances prior to the `finalized` commitment level exposes the clearinghouse to double-spend and ghost deposit risks.
2. **Account Creation and Rent Exemption:** SPL tokens require Program Derived Addresses (PDAs) and Associated Token Accounts (ATAs) funded with minimum lamport rent-exempt balances, requiring deterministic deposit address generation and automated rent-harvesting lifecycle controls.
3. **Cross-Chain Bridge Message Verification:** Bridged assets and data via Wormhole NTT and Pyth require rigorous cryptographic verification of Guardian Verified Action Approvals (VAAs), sequence numbers, consistency levels, and emitter addresses before minting or crediting settlement ledgers.

The **Solana SPL and Wormhole Ingress Service** is a dedicated, ultra-low-latency Rust and Go ingestion microservice. It monitors Solana mainnet-beta in real time, validates inbound native SOL, SPL-USDC, and SPL-USDT transfers to dedicated institutional and retail deposit addresses, verifies Wormhole NTT payloads, detects slot reorganizations, and emits cryptographically attested deposit events to the Growww clearing ledger and internal ledger infrastructure.

## What You Are Building
An enterprise-grade, institutional ingestion service (`services/solana-ingress-service`) designed for sub-second event parsing and zero-loss financial auditing:
- **Real-Time Solana Geyser & RPC Ingestion Engine:** High-performance gRPC block, slot, and account subscription pipeline leveraging Solana Geyser gRPC (Yellowstone Dragon's Mouth / Triton) with fallback to JSON-RPC WebSocket pub/sub (`accountSubscribe`, `programSubscribe`, `blockSubscribe`).
- **Commitment Level Lifecycle Tracker:** State machine managing deposit transactions across confirmation tiers (`processed` $\rightarrow$ `confirmed` $\rightarrow$ `finalized`), holding settlement crediting until root commitment (irreversible supermajority $> 66\%$ stake weight / 32+ confirmation depth) is achieved.
- **Deterministic Custodial ATA Manager:** Derivation and indexing system for user-specific Program Derived Addresses (PDAs) and Associated Token Accounts (ATAs) mapped to the master custodial root authority for USDC, USDT, and custom NTT tokens.
- **Slot Reorganization & Fork Invalidation Detector:** Real-time directed acyclic graph (DAG) slot tracker that identifies orphan slots, dropped optimistic blocks, and chain reorganizations, automatically rolling back unfinalized pending deposits and alerting surveillance.
- **Wormhole NTT & Pyth Bridge Attestation Processor:** Decodes, parses, and validates Wormhole Core Bridge VAA signatures, NTT payload formats (Transfer, TransferWithPayload), Pyth entropy/price messages, sequence numbers, and replay protection nonces.
- **Clearing Ledger Relayer Bridge:** Dispatches standardized, cryptographically signed deposit and bridge attestation events to the Apache Kafka message bus (`wallet.solana_deposit_finalized.v1`) and triggers settlement minting/crediting in the Growww clearing ledger.

## Scope Boundaries
- **In Scope:**
  - Real-time monitoring of Solana mainnet-beta for SOL, SPL-USDC, and SPL-USDT transactions.
  - Streaming ingestion via Solana Geyser gRPC protocol and fallback standard RPC nodes.
  - Tracking commitment states (`processed`, `confirmed`, `finalized`) with deterministic finality barriers.
  - Dynamic derivation, indexing, and rent-exempt validation of Associated Token Accounts (ATAs).
  - Slot reorganization detection, fork resolution, and rollback handling for speculative transactions.
  - Wormhole VAA signature verification (13-of-19 Guardian quorum), NTT message parsing, and emitter verification.
  - Pyth bridge contract attestation and payload verification.
  - Outbox pattern publication of attested deposit events to Kafka and PostgreSQL audit tables.
  - Idempotent deduplication across Solana transaction signatures (`tx_hash`) and Wormhole `(emitter_chain, emitter_address, sequence)` tuples.
- **Out of Scope / Handled Elsewhere:**
  - Outbound withdrawals and Solana transaction signing / key management (handled in Custody Integration Service, Prompt 213).
  - Foreign exchange conversion between USDC/USDT and INR/USD for GIFT City operations (handled in Prompt 214).
  - Internal double-entry user cash ledger crediting and balance reservation (handled in Prompt 203).
  - Smart contract settlement logic on Hyperledger Besu (handled in Prompt 306).
  - Pre-trade sanctions screening of depositor on-chain addresses (handled in Prompts 202 and 703).

## Technology to Use
- **Core Streaming Ingestion Layer:** **Rust 1.78+** using `solana-client`, `solana-sdk`, `yellowstone-grpc-client`, and `tokio` asynchronous runtime for zero-cost memory safety, SIMD-accelerated instruction decoding, and maximum network throughput.
- **Service API & Lifecycle Control Layer:** **Go 1.22+** with `gin-gonic/gin` and `google.golang.org/grpc` for administrative APIs, service health checks, and cross-service orchestration.
- **Solana Network Interface:** Dedicated validator RPC cluster and Yellowstone Dragon's Mouth Geyser gRPC plugin over HTTP/2 with mTLS.
- **Wormhole SDK:** `@wormhole-foundation/sdk-definitions` / Rust Wormhole VAA parser (`wormhole-sdk-rs`) implementing Guardian secp256k1 signature verification.
- **Database & Storage:** **PostgreSQL 16+** with `jackc/pgx/v5` and partition tables for immutable transaction journals, slot progression logs, and deposit records.
- **Distributed Cache & Locking:** **Redis 7.2** for sub-millisecond duplicate transaction signature filtering, active slot fork tracking, and distributed consensus lock states.
- **Message Streaming:** **Apache Kafka 3.7+** for publishing finalized deposit events with guaranteed at-least-once delivery and transactional outbox pattern.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `solana_deposit_transactions`, `solana_custodial_accounts`, `solana_slot_history`, `wormhole_attestations`, `solana_reorg_audit_log`.
- **Redis 7.2 Keys:** `solana:tx:seen:{signature}`, `solana:slot:current_finalized`, `solana:ata:index:{wallet_address}`, `wormhole:vaa:seen:{emitter_chain}:{sequence}`.
- **Apache Kafka Topics:**
  - Consumes: `wallet.account_created.v1`, `custody.sweep_completed.v1`.
  - Publishes: `solana.deposit_detected.v1`, `solana.deposit_finalized.v1`, `wormhole.vaa_attested.v1`, `solana.reorg_detected.v1`.
- **Foreign Investor Funding & FX Service (Prompt 214):** Ingests finalized stablecoin deposit events to trigger USD/INR currency conversion and GIFT City trade allocations.
- **Wallet & Account Ledger Service (Prompt 203):** Credits fiat/stablecoin balances to investor sub-ledgers upon receiving finalized events.
- **Trade Settlement Service (Prompt 208):** Ingests cross-chain settlement attestations for instantaneous Delivery-versus-Payment (DvP) execution.
- **Reconciliation Service (Prompt 215):** Compares on-chain custodial vault balances against internal double-entry ledgers every 60 seconds.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Attestation Relay to Besu:** When an SPL transfer or Wormhole NTT deposit achieves `finalized` commitment on Solana, the service constructs a cryptographic deposit receipt containing:
  - `source_chain_id` (Solana: `1` for Wormhole, `501` for SLIP-0044).
  - `solana_tx_signature` (64-byte base58 string).
  - `slot_number` and `block_hash`.
  - `token_mint_hash` (e.g., hash of USDC mint address `EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v`).
  - `amount_subunits` (uint64 base units, 6 decimals for USDC/USDT).
  - `recipient_member_id_hash` (keccak256 hash of internal investor/clearing ID).
- **Clearing Smart Contract Submission:** The cryptographic attestation is relayed to `SettlementDvP.sol` (Prompt 306) and `ProofOfReserveRegistry.sol` (Prompt 308) on Hyperledger Besu to verify inbound collateral backing 1:1 token minting.
- **Zero PII Standard:** In compliance with DPDP Act 2023 and SEBI/IFSCA regulatory guidelines, public keys, transaction signatures, and slot hashes are transmitted without real-world investor identities. Plaintext identity mapping resides exclusively within encrypted off-chain domestic databases.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Repositories:** Initialize the dual Rust ingestion worker (`crates/solana-ingestor`) and Go gRPC service (`services/solana-ingress-service`) with monorepo build orchestration.
2. **Define Protobuf Contracts:** Author `proto/growww/solana_ingress/v1/solana_ingress_service.proto` specifying gRPC definitions for deposit tracking, account registration, Wormhole VAA verification, and slot status queries.
3. **Generate Language Stubs:** Compile Protobuf schemas to Rust stubs using `tonic-build` / `prost` and Go stubs using `protoc-gen-go` / `protoc-gen-go-grpc`.
4. **Design PostgreSQL Schema:** Implement database migrations for `solana_deposit_transactions`, `solana_custodial_accounts`, `solana_slot_history`, `wormhole_attestations`, and `solana_reorg_audit_log` with TimescaleDB indexing on slot numbers.
5. **Implement Yellowstone Geyser gRPC Ingestion Pipeline (Rust):**
   - Connect to dedicated Solana validator nodes via Yellowstone Dragon's Mouth gRPC stream.
   - Configure account filter subscriptions for master custodial authority and registered Associated Token Accounts (ATAs).
   - Configure transaction filter subscriptions for Wormhole Core Bridge and NTT program IDs (`worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth`).
6. **Implement Fallback RPC Poller & WebSocket Subscriber (Rust):**
   - Implement redundant JSON-RPC connection pool with automatic failover across 3 independent Solana RPC providers.
   - Subscribe to `blockSubscribe` and `programSubscribe` for SPL Token Program (`TokenkegQfeZyiMwajbMRHx5228nW5395NxCU746fy6JP`) and Token-2022 Program (`TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb`).
7. **Implement Solana Transaction & Inner-Instruction Parser (Rust):**
   - Parse native SOL transfers (`SystemProgram::Transfer`).
   - Parse SPL Token `Transfer` and `TransferChecked` instructions from outer and inner CPI instruction trees.
   - Extract and validate mint address, decimals, pre-token balances, and post-token balances to guarantee exact received amount calculation.
8. **Implement Deterministic ATA Address Generator & Indexer (Go/Rust):**
   - Implement standard Solana Associated Token Account derivation logic: `find_program_address(&[wallet_pubkey, TOKEN_PROGRAM_ID, mint_pubkey], ASSOCIATED_TOKEN_PROGRAM_ID)`.
   - Maintain a synchronized Redis and PostgreSQL index of authorized user deposit addresses.
9. **Build Commitment Level Confirmation State Machine (Go):**
   - Track transaction status transitions: `DETECTED` (processed) $\rightarrow$ `CONFIRMED` (optimistic commitment) $\rightarrow$ `FINALIZED` (Tower BFT root reached).
   - Enforce mandatory rule: no deposit is credited or published to the clearing ledger until commitment equals `finalized` (slot confirmed by $> 66\%$ stake weight and at least 32 subsequent slots).
10. **Implement Slot Reorganization & Fork Resolver (Rust/Go):**
    - Maintain a rolling in-memory sliding window of the last 128 slots and their parent blockhashes.
    - If a slot fork is detected where a previously `processed` or `confirmed` transaction signature is absent from the canonical finalized chain, mark the transaction as `FORK_DROPPED`, write to `solana_reorg_audit_log`, and alert risk surveillance.
11. **Implement Wormhole NTT & Pyth Bridge Attestation Engine (Rust):**
    - Implement Guardian set signature validation against known active Guardian public keys (secp256k1 13-of-19 quorum).
    - Parse Wormhole VAA body: `version`, `guardian_set_index`, `timestamp`, `nonce`, `emitter_chain`, `emitter_address`, `sequence`, `consistency_level`, `payload`.
    - Decode NTT payload structs (amount, source token, recipient address, transferee) and verify replay protection via unique sequence database constraints.
12. **Build Transactional Outbox & Kafka Publisher (Go):**
    - Implement PostgreSQL transactional outbox pattern to guarantee exactly-once publication of `solana.deposit_finalized.v1` and `wormhole.vaa_attested.v1` events.
    - Configure Kafka partition keying on `recipient_wallet_id` to guarantee in-order processing.
13. **Implement Custodial Lamport Rent Sweeper & Balance Monitor (Go):**
    - Track minimum rent-exempt balance requirements (approx 0.00203928 SOL per ATA).
    - Flag accounts with surplus native SOL balances for scheduled consolidation sweeps to the master treasury vault.
14. **Configure Observability, Health Checks, and Metrics (Go/Rust):**
    - Export Prometheus metrics: `solana_ingress_slot_height`, `solana_ingress_slot_lag_seconds`, `solana_ingress_deposits_detected_total`, `solana_ingress_deposits_finalized_total`, `solana_ingress_reorgs_detected_total`, `wormhole_vaa_verified_total`.
    - Configure liveness probes asserting RPC/Geyser connection health and stream heartbeats.
15. **Write Comprehensive Mock & Integration Test Suite:**
    - Build test harnesses simulating Solana validator Geyser streams, simulated micro-fork reorgs, duplicate transaction replay attacks, and forged Wormhole VAA payloads.

## Interfaces / Contracts

### Protobuf Definition (`solana_ingress_service.proto`)
```protobuf
syntax = "proto3";

package growww.solana_ingress.v1;

option go_package = "growww/solana_ingress/v1;solana_ingressv1";

service SolanaIngressService {
  rpc RegisterCustodialDepositAccount (RegisterCustodialDepositAccountRequest) returns (RegisterCustodialDepositAccountResponse);
  rpc GetDepositTransactionStatus (GetDepositTransactionStatusRequest) returns (GetDepositTransactionStatusResponse);
  rpc GetSlotProgression (GetSlotProgressionRequest) returns (GetSlotProgressionResponse);
  rpc VerifyWormholeVAA (VerifyWormholeVAARequest) returns (VerifyWormholeVAAResponse);
  rpc ListUnfinalizedDeposits (ListUnfinalizedDepositsRequest) returns (ListUnfinalizedDepositsResponse);
  rpc TriggerManualAttestation (TriggerManualAttestationRequest) returns (TriggerManualAttestationResponse);
}

enum CommitmentLevel {
  COMMITMENT_LEVEL_UNSPECIFIED = 0;
  COMMITMENT_LEVEL_PROCESSED = 1;
  COMMITMENT_LEVEL_CONFIRMED = 2;
  COMMITMENT_LEVEL_FINALIZED = 3;
}

enum IngressDepositStatus {
  INGRESS_DEPOSIT_STATUS_UNSPECIFIED = 0;
  INGRESS_DEPOSIT_STATUS_DETECTED = 1;
  INGRESS_DEPOSIT_STATUS_CONFIRMED = 2;
  INGRESS_DEPOSIT_STATUS_FINALIZED = 3;
  INGRESS_DEPOSIT_STATUS_CREDITED = 4;
  INGRESS_DEPOSIT_STATUS_FORK_DROPPED = 5;
  INGRESS_DEPOSIT_STATUS_REJECTED = 6;
}

enum TokenType {
  TOKEN_TYPE_UNSPECIFIED = 0;
  TOKEN_TYPE_SOL_NATIVE = 1;
  TOKEN_TYPE_SPL_USDC = 2;
  TOKEN_TYPE_SPL_USDT = 3;
  TOKEN_TYPE_WORMHOLE_NTT_CUSTOM = 4;
}

message RegisterCustodialDepositAccountRequest {
  string user_id = 1;
  TokenType token_type = 2;
  string memo_identifier = 3;
  string requested_by_service = 4;
}

message RegisterCustodialDepositAccountResponse {
  string user_id = 1;
  string solana_wallet_pubkey = 2;
  string associated_token_address = 3;
  string token_mint_address = 4;
  int64 rent_exempt_lamports = 5;
  int64 created_at_unix_ms = 6;
}

message GetDepositTransactionStatusRequest {
  string transaction_signature = 1;
}

message GetDepositTransactionStatusResponse {
  string transaction_signature = 1;
  uint64 slot = 2;
  string block_hash = 3;
  IngressDepositStatus status = 4;
  CommitmentLevel commitment = 5;
  string sender_address = 6;
  string recipient_address = 7;
  TokenType token_type = 8;
  string token_mint_address = 9;
  uint64 amount_subunits = 10;
  uint32 token_decimals = 11;
  string amount_decimal = 12;
  int64 detected_at_unix_ms = 13;
  int64 finalized_at_unix_ms = 14;
  int64 confirmations = 15;
}

message GetSlotProgressionRequest {}

message GetSlotProgressionResponse {
  uint64 current_processed_slot = 1;
  uint64 current_confirmed_slot = 2;
  uint64 current_finalized_slot = 3;
  int64 slot_lag = 4;
  int64 active_fork_branches = 5;
  int64 last_reorg_detected_unix_ms = 6;
}

message VerifyWormholeVAARequest {
  bytes raw_vaa_bytes = 1;
  bool submit_attestation_to_clearing = 2;
}

message VerifyWormholeVAAResponse {
  bool is_valid = 1;
  string emitter_chain = 2;
  string emitter_address = 3;
  uint64 sequence = 4;
  uint32 consistency_level = 5;
  string payload_type = 6;
  string recipient_address = 7;
  uint64 amount_subunits = 8;
  string token_mint_address = 9;
  string clearing_attestation_id = 10;
  string error_message = 11;
}

message ListUnfinalizedDepositsRequest {
  uint32 limit = 1;
}

message ListUnfinalizedDepositsResponse {
  repeated GetDepositTransactionStatusResponse pending_deposits = 1;
}

message TriggerManualAttestationRequest {
  string transaction_signature = 1;
  string reason = 2;
  string operator_id = 3;
}

message TriggerManualAttestationResponse {
  bool success = 1;
  string attestation_id = 2;
  string status = 3;
}
```

### PostgreSQL Database Schema DDL
```sql
-- Schema DDL for Solana SPL and Wormhole Ingress Service

CREATE TABLE solana_custodial_accounts (
    account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    wallet_pubkey VARCHAR(44) NOT NULL,
    associated_token_address VARCHAR(44) NOT NULL UNIQUE,
    token_mint_address VARCHAR(44) NOT NULL,
    token_type VARCHAR(32) NOT NULL, -- SOL_NATIVE, SPL_USDC, SPL_USDT, WORMHOLE_NTT_CUSTOM
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    rent_exempt_lamports BIGINT NOT NULL DEFAULT 2039280,
    derivation_path VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE solana_slot_history (
    slot_number BIGINT PRIMARY KEY,
    block_hash VARCHAR(88) NOT NULL,
    parent_slot BIGINT NOT NULL,
    parent_block_hash VARCHAR(88) NOT NULL,
    block_time TIMESTAMPTZ,
    commitment VARCHAR(20) NOT NULL DEFAULT 'PROCESSED', -- PROCESSED, CONFIRMED, FINALIZED
    is_canonical BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finalized_at TIMESTAMPTZ
);

CREATE TABLE solana_deposit_transactions (
    deposit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_signature VARCHAR(88) NOT NULL UNIQUE,
    slot_number BIGINT NOT NULL,
    block_hash VARCHAR(88) NOT NULL,
    sender_address VARCHAR(44) NOT NULL,
    recipient_ata VARCHAR(44) NOT NULL REFERENCES solana_custodial_accounts(associated_token_address),
    user_id UUID NOT NULL,
    token_mint_address VARCHAR(44) NOT NULL,
    token_type VARCHAR(32) NOT NULL,
    amount_subunits BIGINT NOT NULL,
    token_decimals SMALLINT NOT NULL DEFAULT 6,
    amount_decimal NUMERIC(24, 6) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'DETECTED', -- DETECTED, CONFIRMED, FINALIZED, CREDITED, FORK_DROPPED, REJECTED
    commitment VARCHAR(20) NOT NULL DEFAULT 'PROCESSED',
    wormhole_vaa_id UUID,
    clearing_ledger_tx_hash VARCHAR(66),
    credited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wormhole_attestations (
    attestation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vaa_hash VARCHAR(64) NOT NULL UNIQUE,
    emitter_chain SMALLINT NOT NULL, -- e.g. 1 = Solana, 2 = Ethereum
    emitter_address VARCHAR(66) NOT NULL,
    sequence BIGINT NOT NULL,
    consistency_level SMALLINT NOT NULL,
    payload_type VARCHAR(32) NOT NULL, -- TRANSFER, TRANSFER_WITH_PAYLOAD, PYTH_PRICE
    token_address VARCHAR(66) NOT NULL,
    amount_subunits BIGINT NOT NULL,
    recipient_address VARCHAR(66) NOT NULL,
    guardian_set_index INTEGER NOT NULL,
    signatures_count SMALLINT NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    solana_tx_signature VARCHAR(88),
    clearing_settled BOOLEAN NOT NULL DEFAULT FALSE,
    raw_vaa_hex TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMPTZ,
    CONSTRAINT uq_wormhole_emitter_seq UNIQUE (emitter_chain, emitter_address, sequence)
);

CREATE TABLE solana_reorg_audit_log (
    reorg_event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    orphan_slot BIGINT NOT NULL,
    canonical_slot BIGINT NOT NULL,
    orphan_block_hash VARCHAR(88) NOT NULL,
    canonical_block_hash VARCHAR(88) NOT NULL,
    dropped_tx_signatures TEXT[] NOT NULL DEFAULT '{}',
    affected_deposits_count INTEGER NOT NULL DEFAULT 0,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE solana_outbox_events (
    outbox_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id VARCHAR(128) NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,
    retry_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_solana_deposit_status ON solana_deposit_transactions(status, slot_number);
CREATE INDEX idx_solana_deposit_user ON solana_deposit_transactions(user_id);
CREATE INDEX idx_solana_slot_canonical ON solana_slot_history(is_canonical, commitment);
CREATE INDEX idx_solana_outbox_pending ON solana_outbox_events(is_published, created_at) WHERE is_published = FALSE;
```

### Kafka Event Schemas & Wormhole Payload Specifications

#### Event: `solana.deposit_finalized.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SolanaDepositFinalizedEvent",
  "type": "object",
  "required": [
    "event_id",
    "timestamp_ns",
    "deposit_id",
    "transaction_signature",
    "slot_number",
    "block_hash",
    "user_id",
    "token_type",
    "token_mint_address",
    "amount_subunits",
    "token_decimals",
    "amount_decimal",
    "sender_address",
    "recipient_ata",
    "commitment"
  ],
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "timestamp_ns": { "type": "integer" },
    "deposit_id": { "type": "string", "format": "uuid" },
    "transaction_signature": { "type": "string", "maxLength": 88 },
    "slot_number": { "type": "integer" },
    "block_hash": { "type": "string", "maxLength": 88 },
    "user_id": { "type": "string", "format": "uuid" },
    "token_type": { "type": "string", "enum": ["SOL_NATIVE", "SPL_USDC", "SPL_USDT", "WORMHOLE_NTT_CUSTOM"] },
    "token_mint_address": { "type": "string", "maxLength": 44 },
    "amount_subunits": { "type": "integer", "minimum": 1 },
    "token_decimals": { "type": "integer", "enum": [6, 9] },
    "amount_decimal": { "type": "string" },
    "sender_address": { "type": "string", "maxLength": 44 },
    "recipient_ata": { "type": "string", "maxLength": 44 },
    "commitment": { "type": "string", "enum": ["FINALIZED"] },
    "wormhole_vaa_id": { "type": ["string", "null"], "format": "uuid" }
  },
  "additionalProperties": false
}
```

#### Event: `solana.reorg_detected.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SolanaReorgDetectedEvent",
  "type": "object",
  "required": [
    "event_id",
    "timestamp_ns",
    "orphan_slot",
    "canonical_slot",
    "orphan_block_hash",
    "canonical_block_hash",
    "dropped_signatures",
    "affected_deposits_count"
  ],
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "timestamp_ns": { "type": "integer" },
    "orphan_slot": { "type": "integer" },
    "canonical_slot": { "type": "integer" },
    "orphan_block_hash": { "type": "string", "maxLength": 88 },
    "canonical_block_hash": { "type": "string", "maxLength": 88 },
    "dropped_signatures": {
      "type": "array",
      "items": { "type": "string", "maxLength": 88 }
    },
    "affected_deposits_count": { "type": "integer" }
  },
  "additionalProperties": false
}
```

## Security & Compliance Notes
- **Strict Finality Gate:** Deposits must never be credited to internal trading or settlement balances based on `processed` or `confirmed` commitments. Finalization requires explicit Tower BFT root confirmation ($> 66\%$ stake consensus, minimum 32 confirmation slots depth) to prevent double-spending from micro-forks.
- **Wormhole VAA Cryptographic Verification:** All cross-chain messages require 13-of-19 threshold ECDSA secp256k1 signatures from active Wormhole Guardians. Emitter chain IDs and contract addresses must match verified whitelists to prevent spoofing.
- **Deduplication & Replay Protection:** Transaction signatures and Wormhole `(emitter_chain, emitter_address, sequence)` tuples are protected by database unique constraints and Redis distributed idempotency keys with a 72-hour TTL.
- **Zero PII Standard:** In compliance with the Digital Personal Data Protection (DPDP) Act 2023 and SEBI cybersecurity guidelines, on-chain payloads and Kafka events carry only cryptographic user identifiers (`user_id` UUID) and public blockchain addresses.
- **High-Availability RPC Quorum:** Ingestion operates across a quorum of at least 2 independent Geyser streaming endpoints and 3 fallback JSON-RPC providers with automated health detection and latency failover.

## Acceptance Criteria
- [ ] Rust Geyser ingestion worker connects to Solana validator stream and ingests block slots with latency $< 200\text{ms}$.
- [ ] Fallback JSON-RPC poller automatically resumes stream if Yellowstone Geyser disconnects.
- [ ] Associated Token Account (ATA) generator deterministically derives correct PDA addresses for USDC (`EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v`) and USDT (`Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB`).
- [ ] Inbound transactions transition through state machine: `DETECTED` $\rightarrow$ `CONFIRMED` $\rightarrow$ `FINALIZED`.
- [ ] Deposits are credited and published to Kafka `solana.deposit_finalized.v1` only upon achieving `FINALIZED` commitment.
- [ ] Slot reorganization simulation demonstrates dropped orphan slot detection, rollback of unfinalized transactions, and emission of `solana.reorg_detected.v1`.
- [ ] Wormhole VAA parser correctly verifies Guardian signature quorum (13-of-19) and rejects invalid/replayed sequence numbers.
- [ ] Transactional outbox pattern guarantees exactly-once event publication to Kafka without duplicates.
- [ ] Prometheus metrics export accurate slot height, lag, deposit throughput, and verification counters.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Design Standards), Prompt `104` (Event Schemas & Kafka Topic Standards), Prompt `109` (Secrets Management Architecture), Prompt `203` (Wallet & Account Ledger Service).
- **Parallel Tasks:** Prompt `213` (Custodian & Depository Integration Service), Prompt `214` (Foreign Investor Funding & FX Service), Prompt `319` (Cross-Chain Institutional Custody Bridge).
- **Downstream Blockers:** Prompt `208` (Trade Settlement & DvP Service cross-chain stablecoin settlement), Prompt `215` (Reconciliation Service Solana balance verification).
