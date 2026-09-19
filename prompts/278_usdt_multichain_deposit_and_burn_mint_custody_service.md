# 278 - USDT Multi-Chain Deposit, Verification & Custody Ingress Service (Go)

## Purpose
In an institutional digital equity clearing, tokenized asset settlement, and multi-currency investment ecosystem, stablecoin liquidity ingress serves as the fundamental gateway for global and NRI capital. Tether USD (USDT) represents the dominant global stablecoin liquidity rail across multiple major public blockchain infrastructures. However, accepting high-throughput, multi-chain stablecoin deposits directly into an enterprise financial platform introduces critical security, regulatory, and technical challenges:
1. Fragmented Chain Architectures: Ethereum (ERC-20, high gas, Proof-of-Stake finality), Tron (TRC-20, high-speed delegated DPoS, energy/bandwidth fee model), and Polygon PoS (EVM sidechain, Bor/Heimdall consensus with probabilistic deep reorg risks) exhibit divergent RPC semantics, event log structures, and transaction finality timelines.
2. Attack Surface and Fraud: Fraudulent deposits, including spoofed token contracts ("fake USDT"), unconfirmed or dropped zero-gas transactions, flash-loan transaction stuffing, malicious reorg attacks, and non-standard ERC-20 implementations (such as Tether's lack of standard boolean return values on transfer calls), pose continuous threat vectors.
3. Strict Regulatory Compliance: Regulatory frameworks (FIU-IND, FATF Recommendation 16 Travel Rule, PMLA, and GIFT City IFSC regulations) mandate that no un-screened or illicit digital asset may enter the institutional settlement pool. Wallets associated with sanctioned jurisdictions, darknet mixing services, terrorist financing, or ransomware must be intercepted prior to internal ledger crediting.
4. On-Chain Ledger Invariant: To power compliant secondary market trading and DvP settlement on Growww's private, regulatory-approved Hyperledger Besu clearinghouse, incoming verified multi-chain USDT must be securely swept into institutional custody vaults while simultaneously minting an exact 1:1 backed, compliant token (`weUSDT` - Wrapped Equity USDT, an ERC-3643 permissioned asset token) on Hyperledger Besu.

The **USDT Multi-Chain Deposit, Verification & Custody Ingress Service** solves these foundational challenges. Written in high-concurrency, memory-safe Go 1.22+, it operates as an event-driven blockchain ingestion and verification engine that continuously streams, parses, confirms, sanctions-screens, and reconciles USDT deposits across Ethereum, Tron, and Polygon, securely sweeps physical custody funds into MPC-TSS cold vaults, and mints 1:1 custodial-backed `weUSDT` on Hyperledger Besu.

## What You Are Building
An enterprise-grade, event-driven ingestion microservice in **Go 1.22+** deployed at `services/usdt-deposit-service`. Concrete deliverables include:
- **Multi-Chain Node Ingress & WebSocket Block Listeners:** Resilient, self-healing block-and-log ingestion daemons connecting to redundant JSON-RPC and gRPC providers across Ethereum Mainnet (`go-ethereum`), Tron Grid / Tron RPC, and Polygon PoS Bor nodes.
- **Smart Contract Event Log Parser & Filter:** High-performance log filter validating contract addresses, topic hashes, recipient addresses, raw 6-decimal uint256 token values, and non-standard transfer function behaviors against immutable whitelists.
- **Reorg-Resilient Confirmation Tracker:** Dynamic state machine managing confirmation depths per network (e.g., 32 blocks on Ethereum, 19 blocks on Tron, 128 blocks on Polygon) using block reorganization rollback handlers and parent-hash integrity validation.
- **Synchronous Sanctions & AML Screening Pipeline:** Integration gateway with Sanctions Screening Service (Prompt 703) to perform pre-credit blockchain analytics (Chainalysis, TRM Labs, Elliptic), scoring source addresses and transaction provenance before crediting balances.
- **Ledger Ingress & Account Credit Gateway:** Integration with Wallet Service (Prompt 203) using idempotent double-entry ledger commands to update investor fiat/stablecoin custodial balances.
- **Automated Hot-to-Cold MPC Sweep Coordinator:** Balance monitor tracking hot deposit wallet exposure thresholds, triggering automated aggregation and sweeping via MPC Vault Custody Service (Prompt 237) to warm/cold multi-sig reserve vaults.
- **ERC-3643 Besu weUSDT Minting Relayer:** Orchestration client invoking Besu's permissioned `weUSDT` contract through authorized clearing relayer accounts, minting 1:1 asset tokens mapped to verified investor on-chain identities (`ONCHAINID`).
- **Kafka Audit & Telemetry Producer:** Comprehensive outbox event stream emitting real-time telemetry across deposit detection, confirmation milestones, AML decisions, mint authorizations, and sweep executions.

## Scope Boundaries
- **In Scope:**
  - Ingestion and tracking of incoming USDT transactions on Ethereum (ERC-20), Tron (TRC-20), and Polygon PoS.
  - Parsing and cryptographic verification of smart contract `Transfer(address,address,uint256)` event logs.
  - Detection and mitigation of fake-deposit attacks, invalid contract token addresses, zero-value spam, and failed receipts.
  - Multi-stage block confirmation management and chain reorganization rollbacks.
  - Pre-credit AML and sanctions screening dispatch and verdict enforcement (Prompt 703).
  - Calling Wallet Account Service (Prompt 203) for investor ledger accounting.
  - Threshold-based hot-to-cold custody vault sweep dispatch via MPC Vault Custody (Prompt 237).
  - Orchestration of 1:1 `weUSDT` ERC-3643 minting on Hyperledger Besu.
  - Comprehensive PostgreSQL schema for deposit lifecycle, address allocation, confirmations, and audit logs.
  - Internal gRPC API (`UsdtDepositService`) and Kafka event publishing.
- **Out of Scope / Handled Elsewhere:**
  - Key management and raw threshold cryptographic signature generation (handled by MPC-TSS Vault Service in Prompt 237).
  - Fiat INR payment gateway and UPI/IMPS banking rail settlement (handled in Prompt 212 / 214).
  - Secondary market order matching and execution (handled in Prompt 205).
  - Primary clearinghouse DvP settlement logic on Hyperledger Besu (handled in Prompt 208 / 306).
  - User identity document verification and KYC onboarding (handled in Prompt 202).
  - Outbound user USDT withdrawal processing and gas relayer fee calculation (handled in separate withdrawal service).

## Technology to Use
- **Core Runtime & Language:** **Go 1.22+** utilizing structured concurrency (goroutines, channels, `sync.WaitGroup`, `errgroup`), context propagation, and strict memory allocation profiles.
- **Blockchain SDKs & RPC Clients:**
  - `github.com/ethereum/go-ethereum` (`ethclient`, `crypto`, `accounts/abi`) for Ethereum and Polygon JSON-RPC / WebSocket connections.
  - Custom Tron gRPC and HTTP RPC client (`tron-protocol/grpc-api`, `google.golang.org/grpc`) for Tron TRC-20 protocol interaction, Protobuf decoding, and solid block verification.
  - Besu Web3 client bindings (`go-ethereum/ethclient`) for Hyperledger Besu private network interaction.
- **Database & Storage:**
  - **PostgreSQL 16+** with `jackc/pgx/v5` connection pool for ACID-compliant, highly indexed transactional storage.
  - **Redis 7.2** for sub-millisecond distributed locks (`redsync`), block height tracking, and transient replay deduplication.
- **Message Bus & Streaming:**
  - **Apache Kafka** using `segmentio/kafka-go` or `confluent-kafka-go/v2` with transactional producer semantics for guaranteed at-least-once message delivery.
- **Inter-Service Communication:**
  - **gRPC / Protocol Buffers (v3)** via `google.golang.org/grpc` for low-latency internal RPC calls with mutual TLS 1.3.
- **Observability & Health:**
  - OpenTelemetry (`go.opentelemetry.io/otel`), Prometheus client (`prometheus/client_golang`), and structured logging (`go.uber.org/zap`).

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:**
  - `usdt_chain_configs`: Supported blockchain parameters, contract addresses, RPC endpoints, and required confirmation depths.
  - `deposit_addresses`: Allocated user deposit addresses, mapped user IDs, derivation index, chain types, and activation status.
  - `usdt_deposits`: Immutable deposit records, transaction hashes, source/destination addresses, amounts, block numbers, and state.
  - `deposit_confirmations`: Granular per-block confirmation tracking and reorg audit history.
  - `vault_sweeps`: Automated hot-to-cold sweep jobs, MPC session IDs, batch sizes, gas costs, and sweep transaction hashes.
  - `usdt_ingress_audit_logs`: Cryptographically hashed audit log recording state transitions, operator overrides, and AML decisions.
- **Redis 7.2 Keys:**
  - `deposit:lock:{chain_id}:{tx_hash}`: Distributed mutex ensuring single-worker processing per transaction.
  - `deposit:block:processed:{chain_id}`: Monotonically increasing cursor of processed block numbers per chain.
  - `deposit:wallet:pending_balance:{address}`: Aggregated unconfirmed balance for rapid UI reflection.
  - `deposit:ratelimit:{user_id}`: Rate limiting key for address generation requests.
- **Apache Kafka Topics:**
  - Consumes:
    - `wallet.address_generation_requested.v1` (Prompt 203)
    - `compliance.aml_screening_completed.v1` (Prompt 703)
    - `vault.signature_generated.v1` (Prompt 237)
  - Publishes:
    - `usdt.deposit.detected.v1`
    - `usdt.deposit.confirmed.v1`
    - `usdt.deposit.aml_passed.v1`
    - `usdt.deposit.aml_flagged.v1`
    - `usdt.deposit.credited.v1`
    - `usdt.mint.requested.v1`
    - `usdt.mint.completed.v1`
    - `usdt.sweep.triggered.v1`
    - `usdt.sweep.completed.v1`
- **Upstream Callers:**
  - `services/wallet-account-service` (Prompt 203): Requests unique multi-chain deposit address generation for onboarded investors.
  - External Blockchain Validators (Ethereum, Tron, Polygon): Inbound transaction broadcasts.
- **Downstream Consumers:**
  - `services/kyc-aml-service` / `sanctions-screening` (Prompt 703): Ingests deposit metadata for real-time transaction monitoring and counterparty screening.
  - `services/mpc-tss-vault-service` (Prompt 237): Receives sweep requests to sign consolidation transactions moving funds from hot deposit addresses to cold custody.
  - `services/wallet-account-service` (Prompt 203): Updates investor ledger balances upon AML clearance and block confirmation.
  - Hyperledger Besu Permissioned Network: Receives mint transactions to issue `weUSDT`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **External Public Blockchains (Ethereum, Tron, Polygon):**
  - Read-only RPC listeners consume blocks and transaction receipts.
  - Exact Contract Address Filtering:
    - Ethereum USDT (ERC-20): `0xdAC17F958D2ee523a2206206994597C13D831ec7` (Decimals: 6).
    - Tron USDT (TRC-20): `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t` (Decimals: 6).
    - Polygon PoS USDT (ERC-20): `0xc2132D05D31c914a87C6611C10748AEb04B58e8F` (Decimals: 6).
  - Event Topic Filtering: `0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef` representing `Transfer(address,address,uint256)`.
- **Internal Permissioned Clearinghouse (Hyperledger Besu):**
  - Node Architecture: Multi-node QBFT (Quorum Byzantine Fault Tolerance) consortium cluster with 1-second deterministic block times and immediate finality.
  - Token Contract: `weUSDT` implementing the ERC-3643 (Identity-compliant permissioned security token standard), requiring identity registry verification prior to transfer.
  - 1:1 Custodial Backing Invariant: The `weUSDT` token contract grants `MINTER_ROLE` exclusively to the verified Growww Clearinghouse Relayer address managed by the MPC-TSS Vault (Prompt 237). A `mint(address to, uint256 amount)` transaction executes only after the external deposit has achieved required block confirmations, cleared AML sanctions screening, and been recorded in the PostgreSQL audit log.
  - Mathematical Precision Alignment: External USDT transactions (6 decimals, where 1 USDT = 1,000,000 units) are scaled up deterministically by $10^{12}$ to match Besu standard 18-decimal token representation (where 1 `weUSDT` = $10^{18}$ base units), preventing fractional rounding loss.
  - Zero PII Invariant: Besu mint transactions reference strictly the investor's pseudonymized on-chain wallet address and verified identity claims hash. No names, tax IDs (PAN), passport numbers, or email addresses are ever recorded on the blockchain ledger.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Repository:** Initialize Go module `services/usdt-deposit-service` with clean hexagonal architecture (`cmd/`, `internal/domain`, `internal/adapter/blockchain`, `internal/adapter/repository`, `internal/adapter/kafka`, `internal/adapter/grpc`, `internal/config`).
2. **Define Protocol Buffer Schemas:** Author `proto/growww/deposit/v1/usdt_deposit.proto` specifying gRPC procedures for deposit address generation, deposit query, manual verification fallback, sweep triggers, and audit retrieval. Generate Go gRPC code using `protoc-gen-go` and `protoc-gen-go-grpc`.
3. **Design Database Migrations:** Write idempotent PostgreSQL DDL migrations using `golang-migrate/migrate` creating `usdt_chain_configs`, `deposit_addresses`, `usdt_deposits`, `deposit_confirmations`, `vault_sweeps`, and `usdt_ingress_audit_logs`.
4. **Implement Multi-Chain Configuration Loader:** Create a secure configuration loader loading RPC endpoints (primary and fallback), WebSocket URLs, contract addresses, gas caps, and confirmation depth requirements per network from HashiCorp Vault or environment secrets.
5. **Build Ethereum & Polygon Block Streamer:** Implement an event listener utilizing `go-ethereum/ethclient` subscribing to new block headers via WebSocket. Implement an automatic reconnection manager with exponential backoff and backfill poller to bridge disconnected network gaps.
6. **Build Tron TRC-20 Ingress Streamer:** Construct a specialized Tron listener querying Tron Grid / Tron full nodes via gRPC / HTTP. Decode base58/hex addresses, filter contract trigger transactions targeting `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t`, and stream unconfirmed and solid block logs.
7. **Implement Smart Contract Event Log Verifier:** Build a parser extracting `Transfer` event logs from transaction receipts. Verify:
   - Contract address matches configured legitimate USDT contract for the specific chain.
   - Transaction status code is strictly `0x1` (successful).
   - Recipient address (`to`) matches an active assigned address in `deposit_addresses`.
   - Amount is greater than the configured network minimum deposit threshold.
8. **Build Chain Reorganization & Confirmation Tracker:** Implement an asynchronous worker that polls chain progress for pending deposits:
   - Increments confirmation count on each newly mined block.
   - Verifies block hash continuity. If a block hash changes (reorg detected), queries the canonical chain, rolls back orphaned deposit states, and marks unconfirmed transactions as `REORG_DETECTED`.
   - Transitions deposit status to `CONFIRMED` once confirmation depth reaches network safety thresholds (Ethereum: 32 blocks, Tron: 19 blocks, Polygon: 128 blocks).
9. **Integrate Real-Time AML & Sanctions Screening Client:** Implement an event dispatcher emitting `usdt.deposit.detected.v1` to Kafka. Consume `compliance.aml_screening_completed.v1` from Sanctions Screening Service (Prompt 703). If risk score exceeds threshold or matches sanctions lists, immediately freeze deposit, set status to `AML_HOLD`, and alert compliance operations.
10. **Implement Wallet Balance Crediting Gateway:** Build an idempotent gRPC client calling Wallet Account Service (Prompt 203) upon successful AML clearance and confirmation, passing a unique idempotency key (`chain_id:tx_hash:log_index`) to credit the user's custodial balance.
11. **Implement Hyperledger Besu weUSDT Minting Relayer:** Construct an Ethereum client instance connected to Besu nodes. Upon confirmed and AML-cleared deposit, formulate an ABI-encoded `mint(address to, uint256 amount)` call to the `weUSDT` ERC-3643 contract, dispatch the transaction through an MPC-signed relayer wallet, and await Besu receipt confirmation.
12. **Build Automated Hot-to-Cold Vault Sweep Engine:** Implement a periodic sweep daemon evaluating balance thresholds across assigned deposit addresses. When an address accumulates exceeding the sweep trigger threshold (e.g., >= 10,000 USDT):
    - Generates an unsigned sweep transaction sweeping funds to the institutional Cold MPC Vault.
    - Emits a sweep signing request to MPC Vault Custody Service (Prompt 237).
    - Broadcasts the signed transaction once co-signed, recording the sweep in `vault_sweeps`.
13. **Implement Transactional Outbox & Kafka Publisher:** Implement the transactional outbox pattern using PostgreSQL to guarantee atomic storage of deposit state changes and dispatch of Kafka event payloads (`usdt.deposit.confirmed.v1`, `usdt.mint.completed.v1`).
14. **Add Distributed Locking & Idempotency Guards:** Wrap all transaction ingestion, confirmation updates, and sweep operations in Redis distributed locks (`deposit:lock:{chain_id}:{tx_hash}`) to prevent race conditions across multi-replica service deployments.
15. **Construct Comprehensive Test Suite & Observability:** Implement unit tests with mock RPC servers, integration tests with Ganache/Hardhat, reorg simulation test suites, OpenTelemetry tracing spans, and Prometheus metrics for ingestion latency, confirmation delays, and sweep volumes.

## Interfaces / Contracts

### Protobuf Definition (`usdt_deposit.proto`)
```protobuf
syntax = "proto3";

package growww.deposit.v1;

option go_package = "growww/deposit/v1;depositv1";

service UsdtDepositService {
  rpc GenerateDepositAddress (GenerateDepositAddressRequest) returns (GenerateDepositAddressResponse);
  rpc GetDepositStatus (GetDepositStatusRequest) returns (GetDepositStatusResponse);
  rpc ListUserDeposits (ListUserDepositsRequest) returns (ListUserDepositsResponse);
  rpc SimulateDepositVerification (SimulateDepositVerificationRequest) returns (SimulateDepositVerificationResponse);
  rpc TriggerVaultSweep (TriggerVaultSweepRequest) returns (TriggerVaultSweepResponse);
  rpc GetVaultBalance (GetVaultBalanceRequest) returns (GetVaultBalanceResponse);
}

enum BlockchainNetwork {
  BLOCKCHAIN_NETWORK_UNSPECIFIED = 0;
  BLOCKCHAIN_NETWORK_ETHEREUM_MAINNET = 1;
  BLOCKCHAIN_NETWORK_TRON_MAINNET = 2;
  BLOCKCHAIN_NETWORK_POLYGON_POS = 3;
  BLOCKCHAIN_NETWORK_BESU_PRIVATE = 4;
}

enum DepositStatus {
  DEPOSIT_STATUS_UNSPECIFIED = 0;
  DEPOSIT_STATUS_DETECTED = 1;
  DEPOSIT_STATUS_CONFIRMING = 2;
  DEPOSIT_STATUS_CONFIRMED = 3;
  DEPOSIT_STATUS_AML_SCREENING = 4;
  DEPOSIT_STATUS_AML_CLEARED = 5;
  DEPOSIT_STATUS_AML_HOLD = 6;
  DEPOSIT_STATUS_CREDITED = 7;
  DEPOSIT_STATUS_MINT_SUBMITTED = 8;
  DEPOSIT_STATUS_MINT_COMPLETED = 9;
  DEPOSIT_STATUS_REORG_REVERSED = 10;
  DEPOSIT_STATUS_REJECTED_INVALID = 11;
}

enum SweepStatus {
  SWEEP_STATUS_UNSPECIFIED = 0;
  SWEEP_STATUS_PENDING = 1;
  SWEEP_STATUS_SIGNING = 2;
  SWEEP_STATUS_BROADCAST = 3;
  SWEEP_STATUS_CONFIRMED = 4;
  SWEEP_STATUS_FAILED = 5;
}

message GenerateDepositAddressRequest {
  string user_id = 1;
  BlockchainNetwork network = 2;
  string correlation_id = 3;
}

message GenerateDepositAddressResponse {
  string address = 1;
  BlockchainNetwork network = 2;
  string memo_or_tag = 3;
  int64 min_deposit_raw = 4; // Raw 6 decimals (e.g. 10000000 = 10 USDT)
  int64 expires_at_unix = 5;
}

message GetDepositStatusRequest {
  string deposit_id = 1;
  string tx_hash = 2;
  BlockchainNetwork network = 3;
}

message GetDepositStatusResponse {
  string deposit_id = 1;
  string user_id = 2;
  BlockchainNetwork network = 3;
  string tx_hash = 4;
  string from_address = 5;
  string to_address = 6;
  string amount_raw = 7;
  double amount_formatted = 8;
  uint64 current_confirmations = 9;
  uint64 required_confirmations = 10;
  DepositStatus status = 11;
  string aml_risk_score = 12;
  string besu_mint_tx_hash = 13;
  int64 detected_at_unix = 14;
  int64 confirmed_at_unix = 15;
}

message ListUserDepositsRequest {
  string user_id = 1;
  BlockchainNetwork network = 2;
  DepositStatus status_filter = 3;
  int32 page_size = 4;
  string page_token = 5;
}

message ListUserDepositsResponse {
  repeated GetDepositStatusResponse deposits = 1;
  string next_page_token = 2;
  int64 total_count = 3;
}

message SimulateDepositVerificationRequest {
  BlockchainNetwork network = 1;
  string tx_hash = 2;
}

message SimulateDepositVerificationResponse {
  bool is_valid = 1;
  string token_contract = 2;
  string from_address = 3;
  string to_address = 4;
  string amount_raw = 5;
  bool is_assigned_user_address = 6;
  string error_message = 7;
}

message TriggerVaultSweepRequest {
  BlockchainNetwork network = 1;
  string sweep_from_address = 2;
  string destination_vault_address = 3;
  string min_amount_raw = 4;
  string requested_by_operator_id = 5;
}

message TriggerVaultSweepResponse {
  string sweep_job_id = 1;
  SweepStatus status = 2;
  string estimated_sweep_amount_raw = 3;
}

message GetVaultBalanceRequest {
  BlockchainNetwork network = 1;
  string address = 2;
}

message GetVaultBalanceResponse {
  string address = 1;
  BlockchainNetwork network = 2;
  string balance_raw = 3;
  double balance_formatted = 4;
  int64 last_updated_block = 5;
}
```

### PostgreSQL Schema DDL (`000278_usdt_deposit_schema.sql`)
```sql
-- PostgreSQL 16 Schema for USDT Multi-Chain Deposit & Ingress Service

CREATE TYPE blockchain_network AS ENUM (
    'ETHEREUM_MAINNET',
    'TRON_MAINNET',
    'POLYGON_POS',
    'BESU_PRIVATE'
);

CREATE TYPE deposit_status AS ENUM (
    'DETECTED',
    'CONFIRMING',
    'CONFIRMED',
    'AML_SCREENING',
    'AML_CLEARED',
    'AML_HOLD',
    'CREDITED',
    'MINT_SUBMITTED',
    'MINT_COMPLETED',
    'REORG_REVERSED',
    'REJECTED_INVALID'
);

CREATE TYPE sweep_status AS ENUM (
    'PENDING',
    'SIGNING',
    'BROADCAST',
    'CONFIRMED',
    'FAILED'
);

CREATE TABLE usdt_chain_configs (
    network blockchain_network PRIMARY KEY,
    chain_id BIGINT NOT NULL,
    token_contract_address VARCHAR(128) NOT NULL,
    decimals SMALLINT NOT NULL DEFAULT 6,
    required_confirmations INT NOT NULL,
    min_deposit_raw NUMERIC(36, 0) NOT NULL,
    hot_wallet_sweep_threshold_raw NUMERIC(36, 0) NOT NULL,
    cold_vault_address VARCHAR(128) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE deposit_addresses (
    address_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(64) NOT NULL,
    network blockchain_network NOT NULL,
    address VARCHAR(128) NOT NULL,
    derivation_index INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (network, address),
    UNIQUE (user_id, network)
);

CREATE TABLE usdt_deposits (
    deposit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(64) NOT NULL,
    network blockchain_network NOT NULL REFERENCES usdt_chain_configs(network),
    tx_hash VARCHAR(128) NOT NULL,
    log_index INT NOT NULL DEFAULT 0,
    block_number BIGINT NOT NULL,
    block_hash VARCHAR(128) NOT NULL,
    from_address VARCHAR(128) NOT NULL,
    to_address VARCHAR(128) NOT NULL,
    amount_raw NUMERIC(36, 0) NOT NULL,
    amount_scaled NUMERIC(36, 18) NOT NULL, -- Scaled to 18 decimals for Besu
    token_contract VARCHAR(128) NOT NULL,
    current_confirmations INT NOT NULL DEFAULT 0,
    required_confirmations INT NOT NULL,
    status deposit_status NOT NULL DEFAULT 'DETECTED',
    aml_risk_score VARCHAR(32),
    aml_screening_reference VARCHAR(128),
    aml_completed_at TIMESTAMPTZ,
    wallet_credit_reference VARCHAR(128),
    credited_at TIMESTAMPTZ,
    besu_mint_tx_hash VARCHAR(128),
    besu_minted_at TIMESTAMPTZ,
    idempotency_key VARCHAR(256) NOT NULL UNIQUE,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE deposit_confirmations (
    confirmation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deposit_id UUID NOT NULL REFERENCES usdt_deposits(deposit_id) ON DELETE CASCADE,
    block_number BIGINT NOT NULL,
    block_hash VARCHAR(128) NOT NULL,
    parent_block_hash VARCHAR(128) NOT NULL,
    confirmations_recorded INT NOT NULL,
    is_canonical BOOLEAN NOT NULL DEFAULT TRUE,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vault_sweeps (
    sweep_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    network blockchain_network NOT NULL REFERENCES usdt_chain_configs(network),
    from_address VARCHAR(128) NOT NULL,
    to_cold_address VARCHAR(128) NOT NULL,
    amount_raw NUMERIC(36, 0) NOT NULL,
    sweep_tx_hash VARCHAR(128),
    mpc_session_id VARCHAR(128),
    gas_fee_raw NUMERIC(36, 0),
    status sweep_status NOT NULL DEFAULT 'PENDING',
    operator_id VARCHAR(64),
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    broadcast_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    failure_reason TEXT
);

CREATE TABLE usdt_ingress_audit_logs (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deposit_id UUID REFERENCES usdt_deposits(deposit_id),
    network blockchain_network NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    actor_id VARCHAR(128) NOT NULL,
    previous_state VARCHAR(32),
    new_state VARCHAR(32),
    payload JSONB NOT NULL,
    hash_digest VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance and high-throughput lookups
CREATE INDEX idx_deposit_addresses_lookup ON deposit_addresses(network, address);
CREATE INDEX idx_deposit_addresses_user ON deposit_addresses(user_id);
CREATE INDEX idx_usdt_deposits_user_status ON usdt_deposits(user_id, status);
CREATE INDEX idx_usdt_deposits_tx ON usdt_deposits(network, tx_hash, log_index);
CREATE INDEX idx_usdt_deposits_status ON usdt_deposits(status) WHERE status IN ('DETECTED', 'CONFIRMING', 'AML_SCREENING');
CREATE INDEX idx_deposit_confirmations_block ON deposit_confirmations(deposit_id, block_number);
CREATE INDEX idx_vault_sweeps_network_status ON vault_sweeps(network, status);
CREATE INDEX idx_usdt_audit_deposit ON usdt_ingress_audit_logs(deposit_id, created_at DESC);
```

## Security & Compliance Notes
- **Fake-Deposit Attack Prevention:**
  - Strict Token Contract Address Verification: The service maintains an immutable whitelist of official USDT contract addresses (`0xdAC17F958D2ee523a2206206994597C13D831ec7` on Ethereum, `TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t` on Tron, `0xc2132D05D31c914a87C6611C10748AEb04B58e8F` on Polygon). Any transaction emitting a `Transfer` event where the emitting contract does not match the exact canonical address is immediately tagged `REJECTED_INVALID` and discarded.
  - Transaction Receipt Status Verification: Ingestion never parses memory pool pending transactions or raw unmined payloads. It requires a completed receipt from full archive nodes where status code equals `0x1` (success). Reverted transactions that emitted speculative logs are rejected.
  - Non-Standard ERC-20 Return Value Handling: Tether on Ethereum does not adhere to standard ERC-20 return specifications (it does not return a boolean `true` on successful transfer). The Go parser utilizes custom raw-byte ABI decoders rather than standard generated wrappers that crash on missing return buffers.
- **Chain Reorganization & Double-Spending Defenses:**
  - Confirmation Depth Parameterization: To neutralize shallow and deep reorganizations, minimum block confirmation thresholds are enforced:
    - Ethereum Mainnet: 32 blocks (reflecting Casper Proof-of-Stake epoch finality checkpoints).
    - Tron Mainnet: 19 blocks (reflecting 2/3+ Super Representative solid block confirmations).
    - Polygon PoS: 128 blocks (accounting for historical Bor sprint deep reorgs).
  - Parent Hash Continuity Tracking: Every ingested block records `parent_block_hash`. If a discrepancy or orphaned block is identified, the confirmation tracker triggers a recursive parent audit, rolls back unconfirmed states, and re-validates transactions on the canonical chain branch.
- **Travel Rule & Sanctions Screening Compliance (PMLA / FATF / FIU-IND):**
  - Originating Address Sanctions Evaluation: Prior to crediting an investor account or authorizing minting on Hyperledger Besu, the sender address (`from_address`) and transaction hash are screened through the Sanctions Screening Service (Prompt 703).
  - High-Risk Interception: Any association with OFAC-sanctioned addresses, mixing services (e.g., Tornado Cash), ransomware clusters, or darknet markets triggers an automated freeze (`AML_HOLD`), preventing ledger crediting and alerting the compliance MLRO (Money Laundering Reporting Officer).
  - Statutory Travel Rule Metadata Validation: For deposits exceeding regulatory thresholds (USD 1,000 / INR 50,000 equivalent), IVMS 101 counterparty originator information must be validated against registered VASP identities.
- **Hot-to-Cold Vault Risk Containment:**
  - Deposit Address Isolation: Each investor is allocated a unique deterministic deposit address derived from an MPC master public key.
  - Sweep Limit Exposure: When aggregated balance on an active hot deposit address exceeds the configurable threshold (e.g., 10,000 USDT), an automated sweep job transfers funds to the multi-sig cold vault (Prompt 237), ensuring minimal capital at risk in live operational addresses.
- **Idempotency & Exactly-Once Ingress:**
  - The database enforces a strict unique constraint on `idempotency_key` (`network:tx_hash:log_index`). Repeated RPC events, duplicate Kafka messages, or dual-worker processing will hit unique key violations and abort without duplicate credit or double minting.

## Acceptance Criteria
- [ ] Multi-chain WebSocket and RPC listeners ingest incoming USDT transfer events on Ethereum, Tron, and Polygon with end-to-end detection latency $< 2\text{ seconds}$ from block inclusion.
- [ ] Fake-deposit attacks (counterfeit contract emitting Transfer events, reverted transactions, zero-value transfers) are rejected and logged to `usdt_ingress_audit_logs`.
- [ ] Reorganization handling automatically detects deep fork swaps, rolls back unconfirmed deposit states, and preserves canonical consistency without manual intervention.
- [ ] Minimum confirmation depth policies (Ethereum: 32, Tron: 19, Polygon: 128) are enforced prior to transitioning deposit status to `CONFIRMED`.
- [ ] AML screening gateway integration (Prompt 703) runs synchronously prior to ledger crediting; high-risk flagged transactions are quarantined in `AML_HOLD`.
- [ ] Double-entry ledger balance crediting with Wallet Account Service (Prompt 203) executes with zero duplicate credits under extreme replay simulation.
- [ ] Hyperledger Besu `weUSDT` ERC-3643 minting executes 1:1 against confirmed deposits with accurate mathematical decimal scaling ($10^6 \to 10^{18}$).
- [ ] Automated hot-to-cold sweep engine initiates consolidation transactions to MPC Vault Custody (Prompt 237) when address balances reach configured sweep thresholds.
- [ ] Structured transactional outbox guarantees reliable publication of all deposit lifecycle events to Apache Kafka.
- [ ] OpenTelemetry traces and Prometheus metrics expose real-time block lag, processing duration, confirmation queues, and error rates.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Design Standards), Prompt `104` (Event Schema & Kafka Topic Standards), Prompt `111` (Domain Model Core Entities), Prompt `112` (Idempotency Standards), Prompt `203` (Wallet Account Service).
- **Parallel Tasks:** Prompt `703` (Sanctions & PEP Screening Integration), Prompt `218` (Audit Log Service), Prompt `215` (Reconciliation Service).
- **Downstream Blockers:** Prompt `237` (MPC-TSS Vault Custody Service - for signature execution on vault sweeps), Prompt `306` (Settlement DvP Smart Contract relayer execution on Besu), Prompt `319` (Institutional Custody Bridge).
