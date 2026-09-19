# 235 - EVM Chainlink CCIP Multi-Token Ingress & Canonical Lockbox Bridge Service (Go / CCIP / EIP-712)

## Purpose
To provide global institutional and retail investors with seamless, non-custodial capital ingress and egress across public Ethereum Virtual Machine (EVM) blockchains while maintaining strict accounting integrity, deterministic finality, and 1:1 backing against regulated custody ledgers. 

The **EVM Chainlink CCIP Multi-Token Ingress & Canonical Lockbox Bridge Service** acts as the decentralized multi-chain financial gateway for Growww. It continuously monitors, verifies, and settles inbound and outbound digital asset transfers across five major EVM networks (Ethereum Mainnet, Arbitrum One, Optimism, Base, and Polygon PoS) across five standard institutional crypto assets (ETH/WETH, USDC, USDT, DAI, and WBTC).

By utilizing Chainlink Cross-Chain Interoperability Protocol (CCIP) alongside canonical institutional lockbox smart contracts, the service automates cross-chain token routing, protects the clearinghouse against chain reorganizations through configurable finality watchers, facilitates frictionless user onboarding via EIP-712 permit meta-transactions, dynamically rebases gas fees to prevent stuck bridge transactions, and routes cryptographically attested deposit events into the internal `wallet-service` (Prompt 203) for instant fiat or synthetic ledger crediting.

## What You Are Building
A high-throughput, mission-critical Go microservice (`services/evm-ingress-service`) operating across multi-cloud RPC clusters with Hardware Security Module (HSM) transaction signing, distributed block reorg detection, and transactional outbox event streaming. Concrete deliverables include:
- **Multi-Chain RPC Finality Watchers & Block Ingestion Engine:** Resilient multi-node WebSocket/JSON-RPC block monitors across Ethereum Mainnet (L1), Arbitrum One (L2), Optimism (L2), Base (L2), and Polygon PoS (Sidechain/L2 commit-chain) tracking block confirmations, safe/finalized execution tags, and reorg depths.
- **Canonical Lockbox Smart Contract Watcher:** Real-time event listener monitoring on-chain deposit events (`TokensLocked`, `TokensReleased`, `DepositWithData`) emitted by audited Lockbox smart contracts deployed across all supported EVM chains.
- **Chainlink CCIP Programmable Bridge Orchestrator:** Bidirectional bridge relayer that constructs, quotes, dispatches, and monitors Chainlink CCIP cross-chain token transfers (`CCIPSendRequested`, `ExecutionStateChanged`), handling CCIP Router interactions, TokenPool validations, and Risk Management Network / Arm Network consensus checks.
- **EIP-712 Permit Meta-Transaction Ingress Relayer:** Gasless deposit relayer accepting EIP-2612 / EIP-3009 / Permit2 signatures from end-user wallets, verifying cryptographic digests, and broadcasting batched `depositWithPermit` transactions while managing sponsor gas allowances.
- **Dynamic Gas Fee Rebasing & Transaction Lifecycle Manager:** Automated EIP-1559 gas pricer that monitors pending mempool transactions, applies dynamic Replace-By-Fee (RBF) gas bumps (minimum 15% escalation) during chain congestion spikes, and rebases CCIP execution fee budgets across native tokens and LINK.
- **Atomic Wallet Ledger Event Dispatcher:** Transactional outbox publisher that routes verified, finalized deposit and withdrawal events into Kafka topic `wallet.evm.ingress.v1` for atomic ingestion by `wallet-service` (Prompt 203).

## Scope Boundaries
- **In Scope:**
  - Real-time ingestion and event filtering of ERC-20 and native token deposits (ETH, USDC, USDT, DAI, WBTC) across Ethereum, Arbitrum, Optimism, Base, and Polygon.
  - Deep block reorg tracking, fork detection, and state reconciliation (safe block vs finalized block transitions).
  - Integration with Chainlink CCIP v1.5+ Router and TokenPool contracts for cross-chain bridging.
  - EIP-712 permit parsing, signature verification, and subsidized relayer dispatch.
  - Dynamic gas escalation, stuck transaction cancellation, and gas price rebasing.
  - Double-spend prevention, deposit deduplication, and atomic routing to `wallet-service`.
  - Cryptographic validation of transaction receipts and Merkle proof generation for ledger auditing.
- **Out of Scope / Handled Elsewhere:**
  - Double-entry internal fiat ledgering and user cash account balances (handled in Prompt 203).
  - Institutional cold-storage key custody and physical bank depository reserves (handled in Prompt 213).
  - Off-chain foreign exchange conversions (USD/INR) and GIFT City wire settlements (handled in Prompt 214).
  - Core trade matching and order execution (handled in Prompt 205).
  - On-chain fractional equity token issuance on permissioned Hyperledger Besu (handled in Prompt 303).

## Technology to Use
- **Primary Language & Runtime:** **Go 1.22+** with `gin-gonic/gin` for administrative APIs and `google.golang.org/grpc` for low-latency inter-service RPC communications.
- **EVM Client Libraries:** `ethereum/go-ethereum` (`geth`) `ethclient`, `rpc`, `accounts/abi`, and `crypto` packages for native ABI encoding, raw transaction serialization, and RPC node failover orchestration.
- **Cross-Chain Protocol:** **Chainlink CCIP Go SDK / Contracts bindings** (`@chainlink/contracts-ccip`) for cross-chain message creation, fee calculation, and execution tracking.
- **Database & Storage:** **PostgreSQL 16+** with `jackc/pgx/v5` connection pool for tracking monitored blocks, deposit transactions, gas rebasing logs, and reorg history.
- **Cache & Distributed Locking:** **Redis 7.2** for sub-millisecond nonce tracking, mempool duplicate submission filters, and distributed lease locks across watcher replicas.
- **Event Streaming:** **Apache Kafka** via `segmentio/kafka-go` for broadcasting guaranteed deposit lifecycle events (`detected`, `confirmed`, `reorg_invalidated`, `bridged`).
- **HSM / KMS Signing:** AWS KMS / HashiCorp Vault Transit engine for securing relayer private keys with strict role-based access policies.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `evm_supported_chains`, `evm_token_registries`, `evm_monitored_blocks`, `evm_ingress_deposits`, `evm_egress_withdrawals`, `ccip_bridge_messages`, `gas_rebasing_history`, `reorg_audit_records`.
- **Redis 7.2 Keys:** `evm:chain:{chain_id}:latest_block`, `evm:nonce:relayer:{chain_id}`, `evm:tx:lock:{chain_id}:{tx_hash}`, `evm:gas:price_cache:{chain_id}`.
- **Apache Kafka Topics:**
  - Publishes: `wallet.evm_deposit.detected.v1`, `wallet.evm_deposit.confirmed.v1`, `wallet.evm_deposit.reorg_invalidated.v1`, `wallet.ccip_bridge.completed.v1`, `wallet.evm_withdrawal.broadcasted.v1`.
  - Consumes: `wallet.evm_withdrawal.requested.v1`, `admin.evm_lockbox.emergency_pause.v1`.
- **Upstream / Peer Services:**
  - `services/wallet-service` (Prompt 203): Consumes deposit events to credit user multi-asset balances and dispatches withdrawal requests.
  - `services/kyc-aml-service` (Prompt 202): Performs on-chain AML wallet screening (Chainalysis/Elliptic oracle) prior to deposit finalization.
  - `services/audit-log-service` (Prompt 218): Ingests immutable Merkle receipts and cryptographic proof of deposit.
- **External RPC Gateways:** Alchemy, Infura, QuickNode, and dedicated validator nodes across Ethereum Mainnet, Arbitrum One, Optimism, Base, and Polygon.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Multi-Chain to Permissioned Ledger Bridge:** When an inbound deposit achieves finality on an external EVM chain (e.g., 50,000 USDC locked on Arbitrum Lockbox), this service verifies the transaction proof and issues an attested deposit voucher to the internal relayer. The relayer invokes `ProofOfReserveRegistry.sol` (Prompt 308) on Hyperledger Besu, establishing 1:1 custody backing before synthetic investment tokens are minted.
- **Zero PII on Public Ledgers:** Public EVM lockboxes and CCIP message payloads contain strictly pseudonymous identifiers: `user_account_uuid_hash` (`bytes32`), `source_chain_id` (`uint64`), `token_address` (`address`), and `amount` (`uint256`). No investor names, email addresses, KYC IDs, or domestic tax identifiers ever touch public mempools.
- **Consensus & Finality Coupling:** Hyperledger Besu QBFT consensus provides instant 2-second deterministic finality once a cross-chain deposit is validated off-chain, ensuring zero ledger rollbacks internally even if an external L2 undergoes delayed batch dispute resolution.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/evm-ingress-service` adhering to hexagonal clean architecture (`cmd/`, `internal/domain/`, `internal/watcher/`, `internal/ccip/`, `internal/relayer/`, `internal/adapter/`).
2. **Compile Smart Contract ABIs:** Generate Go bindings using `abigen` for Canonical Lockbox (`ILockbox.json`), ERC-20 with Permit (`IERC20Permit.json`), Permit2 (`IPermit2.json`), and Chainlink CCIP Router (`IRouterClient.json`).
3. **Define Protobuf Schema:** Create `proto/growww/ingress/v1/evm_ingress_service.proto` specifying gRPC endpoints for querying deposit states, submitting EIP-712 permits, initiating cross-chain CCIP withdrawals, and reading chain finality metrics.
4. **Design PostgreSQL Migrations:** Author database DDL scripts for supported chains, token configurations, block headers, deposit states, CCIP cross-chain message tracking, and reorg event tables.
5. **Implement Resilient Multi-Chain RPC Provider:** Build an adaptive RPC connection pool with automatic round-robin failover, health probing, rate-limit throttling, and fallback across primary and secondary JSON-RPC/WebSocket node endpoints.
6. **Implement Block Ingestion & Finality Watchers:** Build concurrent block stream listeners per chain implementing chain-specific finality models:
   - *Ethereum:* Track `safe` (1 epoch / 32 slots) and `finalized` (2 epochs / 64 slots) execution tags.
   - *Arbitrum One & Optimism:* Monitor sequencer feed for soft confirmation, followed by L1 data availability batch finalization.
   - *Base:* Track L1 rollup state batch settlement.
   - *Polygon PoS:* Monitor Heimdall checkpoint state and Milestones for irreversible finality (128+ confirmations).
7. **Implement Chain Reorganization Detector:**
   - Maintain a sliding window of recent block headers (`parent_hash`, `block_number`, `block_hash`) in PostgreSQL and memory.
   - If an incoming block reveals a `parent_hash` mismatch with the local canonical chain, initiate fork detection.
   - Traverse backwards to identify the common ancestor block, mark invalidated unconfirmed deposits in state `REORG_INVALIDATED`, and emit Kafka alert `wallet.evm_deposit.reorg_invalidated.v1`.
8. **Build Lockbox Event Parser & Filter:** Ingest `TokensLocked(address sender, address token, uint256 amount, bytes32 indexed accountHash, uint64 destinationChainId)` events from canonical lockbox contracts, validating token whitelist rules and minimum deposit thresholds.
9. **Build EIP-712 Permit Meta-Transaction Handler:**
   - Implement EIP-712 domain separator verification and signature unpacker for ERC-20 `permit(owner, spender, value, deadline, v, r, s)` and EIP-3009 `receiveWithAuthorization`.
   - Validate permit expiration timestamps and nonce freshness against target EVM chain state.
   - Assemble lockbox contract call `depositWithPermit` and dispatch via the gasless relayer.
10. **Implement Chainlink CCIP Cross-Chain Bridge Engine:**
    - Interface with Chainlink `IRouterClient` across source and destination chains.
    - Compute CCIP execution fees dynamically using `getFee(destinationChainSelector, message)` with native gas token or LINK payment strategy.
    - Dispatch `ccipSend` transactions for outbound cross-chain liquidity rebalancing.
    - Monitor `CCIPSendRequested` and `ExecutionStateChanged` events emitted by CCIP OffRamps, tracking bridge execution through states `UNTIL_FINALIZED`, `IN_FLIGHT`, `SUCCESS`, and `FAILED`.
11. **Implement Dynamic Gas Rebasing & EIP-1559 Relayer Engine:**
    - Query `eth_feeHistory` and base fee trends per block to compute optimal `maxFeePerGas` and `maxPriorityFeePerGas`.
    - If a broadcasted relayer transaction remains pending beyond a configurable timeout (e.g., 60 seconds on L1, 15 seconds on L2), execute Replace-By-Fee (RBF) by bumping gas fees by at least 15% with identical nonce.
    - Implement emergency transaction cancellation (0 ETH self-transfer with higher gas) if contract execution reverts unexpectedly.
12. **Build Transactional Outbox Event Dispatcher:**
    - Record deposit state transitions (`DETECTED` -> `SAFE_CONFIRMED` -> `FINALIZED`) atomically within PostgreSQL transactions.
    - Poll outbox table and publish verified events to Apache Kafka topic `wallet.evm.ingress.v1` with guaranteed at-least-once delivery.
13. **Implement Admin & Circuit-Breaker Controls:** Provide emergency lockbox pausing, asset deposit threshold caps, and automated relayer wallet balance auto-refill triggers.
14. **Configure Observability & Metrics:** Export Prometheus metrics: `evm_deposit_amount_total`, `evm_reorg_depth_blocks`, `evm_rpc_latency_seconds`, `ccip_bridge_duration_seconds`, `evm_relayer_gas_balance_eth`.
15. **Implement Mock & Multi-Chain Integration Test Suite:** Develop end-to-end integration tests using `testcontainers-go` and local Anvil / Hardhat multi-chain forks simulating block reorgs (depth 1 to 5), EIP-712 permits, and CCIP message delivery.

## Interfaces / Contracts

### Protobuf Definition (`evm_ingress_service.proto`)
```protobuf
syntax = "proto3";

package growww.ingress.v1;

option go_package = "growww/ingress/v1;ingressv1";

service EvmIngressService {
  rpc GetDepositStatus (GetDepositStatusRequest) returns (GetDepositStatusResponse);
  rpc SubmitPermitDeposit (SubmitPermitDepositRequest) returns (SubmitPermitDepositResponse);
  rpc EstimateBridgeFee (EstimateBridgeFeeRequest) returns (EstimateBridgeFeeResponse);
  rpc InitiateCcipBridgeWithdrawal (InitiateCcipBridgeWithdrawalRequest) returns (InitiateCcipBridgeWithdrawalResponse);
  rpc GetChainFinalityMetrics (GetChainFinalityMetricsRequest) returns (GetChainFinalityMetricsResponse);
  rpc ListSupportedTokens (ListSupportedTokensRequest) returns (ListSupportedTokensResponse);
}

enum EvmChainId {
  CHAIN_UNSPECIFIED = 0;
  CHAIN_ETHEREUM_MAINNET = 1;
  CHAIN_ARBITRUM_ONE = 42161;
  CHAIN_OPTIMISM_MAINNET = 10;
  CHAIN_BASE_MAINNET = 8453;
  CHAIN_POLYGON_POS = 137;
}

enum TokenType {
  TOKEN_UNSPECIFIED = 0;
  TOKEN_ETH = 1;
  TOKEN_USDC = 2;
  TOKEN_USDT = 3;
  TOKEN_DAI = 4;
  TOKEN_WBTC = 5;
}

enum DepositState {
  DEPOSIT_STATE_UNSPECIFIED = 0;
  DEPOSIT_STATE_DETECTED = 1;
  DEPOSIT_STATE_SAFE_CONFIRMED = 2;
  DEPOSIT_STATE_FINALIZED = 3;
  DEPOSIT_STATE_REORG_INVALIDATED = 4;
  DEPOSIT_STATE_CREDITED = 5;
  DEPOSIT_STATE_REJECTED_AML = 6;
}

enum CcipBridgeStatus {
  CCIP_STATUS_UNSPECIFIED = 0;
  CCIP_STATUS_PENDING_SUBMISSION = 1;
  CCIP_STATUS_IN_FLIGHT = 2;
  CCIP_STATUS_COMMITTED_ON_DESTINATION = 3;
  CCIP_STATUS_EXECUTED_SUCCESS = 4;
  CCIP_STATUS_FAILED = 5;
}

message GetDepositStatusRequest {
  string deposit_id = 1;
  string tx_hash = 2;
  EvmChainId chain_id = 3;
}

message GetDepositStatusResponse {
  string deposit_id = 1;
  EvmChainId chain_id = 2;
  string tx_hash = 3;
  uint64 log_index = 4;
  string user_account_hash = 5;
  TokenType token = 6;
  string token_address = 7;
  string raw_amount = 8;
  string decimal_amount = 9;
  uint64 block_number = 10;
  uint64 confirmations = 11;
  uint64 required_confirmations = 12;
  DepositState state = 13;
  int64 detected_at_unix_ms = 14;
  int64 finalized_at_unix_ms = 15;
}

message SubmitPermitDepositRequest {
  EvmChainId chain_id = 1;
  string user_account_hash = 2;
  TokenType token = 3;
  string token_address = 4;
  string owner_address = 5;
  string amount = 6;
  uint64 deadline = 7;
  uint32 v = 8;
  string r = 9; // Hex 32 bytes
  string s = 10; // Hex 32 bytes
  string idempotency_key = 11;
}

message SubmitPermitDepositResponse {
  string deposit_id = 1;
  string relayer_tx_hash = 2;
  DepositState initial_state = 3;
  int64 submitted_at_unix_ms = 4;
}

message EstimateBridgeFeeRequest {
  EvmChainId source_chain_id = 1;
  EvmChainId destination_chain_id = 2;
  TokenType token = 3;
  string amount = 4;
  bool pay_in_link = 5;
}

message EstimateBridgeFeeResponse {
  string fee_native_wei = 1;
  string fee_link_juels = 2;
  string estimated_duration_seconds = 3;
  uint64 ccip_chain_selector = 4;
}

message InitiateCcipBridgeWithdrawalRequest {
  string withdrawal_id = 1;
  EvmChainId source_chain_id = 2;
  EvmChainId destination_chain_id = 3;
  TokenType token = 4;
  string recipient_evm_address = 5;
  string amount = 6;
  bool pay_in_link = 7;
  string idempotency_key = 8;
}

message InitiateCcipBridgeWithdrawalResponse {
  string withdrawal_id = 1;
  string ccip_message_id = 2;
  string source_tx_hash = 3;
  CcipBridgeStatus status = 4;
  int64 dispatched_at_unix_ms = 5;
}

message GetChainFinalityMetricsRequest {
  EvmChainId chain_id = 1;
}

message GetChainFinalityMetricsResponse {
  EvmChainId chain_id = 1;
  uint64 latest_block_number = 2;
  uint64 safe_block_number = 3;
  uint64 finalized_block_number = 4;
  uint64 current_confirmation_threshold = 5;
  uint64 reorg_depth_last_24h = 6;
  bool rpc_healthy = 7;
  int64 last_block_timestamp_unix_ms = 8;
}

message ListSupportedTokensRequest {
  EvmChainId chain_id = 1;
}

message TokenInfo {
  TokenType token = 1;
  string symbol = 2;
  string contract_address = 3;
  uint32 decimals = 4;
  bool permit_supported = 5;
  string min_deposit_amount = 6;
  string max_deposit_amount = 7;
  bool is_active = 8;
}

message ListSupportedTokensResponse {
  repeated TokenInfo tokens = 1;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE evm_supported_chains (
    chain_id BIGINT PRIMARY KEY,
    chain_name VARCHAR(64) NOT NULL,
    chain_type VARCHAR(32) NOT NULL, -- L1, L2_OPTIMISTIC_ROLLUP, L2_ZK_ROLLUP, SIDECHAIN
    ccip_chain_selector BIGINT NOT NULL UNIQUE,
    lockbox_contract_address VARCHAR(42) NOT NULL,
    ccip_router_address VARCHAR(42) NOT NULL,
    required_confirmations INT NOT NULL DEFAULT 64,
    finality_model VARCHAR(32) NOT NULL, -- ETH_FINALIZED_TAG, L2_BATCH_FINALITY, POLYGON_CHECKPOINT
    rpc_primary_url VARCHAR(256) NOT NULL,
    rpc_secondary_url VARCHAR(256) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE evm_token_registries (
    token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id BIGINT NOT NULL REFERENCES evm_supported_chains(chain_id),
    symbol VARCHAR(16) NOT NULL,
    name VARCHAR(64) NOT NULL,
    contract_address VARCHAR(42) NOT NULL,
    decimals INT NOT NULL,
    supports_eip2612_permit BOOLEAN NOT NULL DEFAULT FALSE,
    supports_permit2 BOOLEAN NOT NULL DEFAULT FALSE,
    min_deposit_limit NUMERIC(36, 18) NOT NULL,
    max_deposit_limit NUMERIC(36, 18) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_chain_token_address UNIQUE (chain_id, contract_address)
);

CREATE TABLE evm_monitored_blocks (
    block_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id BIGINT NOT NULL REFERENCES evm_supported_chains(chain_id),
    block_number BIGINT NOT NULL,
    block_hash VARCHAR(66) NOT NULL,
    parent_hash VARCHAR(66) NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    is_canonical BOOLEAN NOT NULL DEFAULT TRUE,
    is_finalized BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_chain_block_hash UNIQUE (chain_id, block_hash),
    CONSTRAINT uq_chain_block_number_canonical UNIQUE (chain_id, block_number, is_canonical)
);

CREATE TABLE evm_ingress_deposits (
    deposit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id BIGINT NOT NULL REFERENCES evm_supported_chains(chain_id),
    tx_hash VARCHAR(66) NOT NULL,
    log_index BIGINT NOT NULL,
    block_number BIGINT NOT NULL,
    block_hash VARCHAR(66) NOT NULL,
    sender_address VARCHAR(42) NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    token_symbol VARCHAR(16) NOT NULL,
    amount_raw NUMERIC(78, 0) NOT NULL,
    amount_decimal NUMERIC(36, 18) NOT NULL,
    user_account_hash VARCHAR(66) NOT NULL,
    state VARCHAR(32) NOT NULL DEFAULT 'DETECTED', -- DETECTED, SAFE_CONFIRMED, FINALIZED, REORG_INVALIDATED, CREDITED, REJECTED_AML
    aml_screening_status VARCHAR(32) NOT NULL DEFAULT 'PENDING', -- PENDING, PASSED, FLAGGED, REJECTED
    confirmations_observed INT NOT NULL DEFAULT 0,
    required_confirmations INT NOT NULL,
    reorg_depth INT NOT NULL DEFAULT 0,
    kafka_published_at TIMESTAMPTZ,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finalized_at TIMESTAMPTZ,
    credited_at TIMESTAMPTZ,
    CONSTRAINT uq_chain_tx_log UNIQUE (chain_id, tx_hash, log_index)
);

CREATE TABLE evm_egress_withdrawals (
    withdrawal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_chain_id BIGINT NOT NULL REFERENCES evm_supported_chains(chain_id),
    destination_chain_id BIGINT NOT NULL REFERENCES evm_supported_chains(chain_id),
    token_address VARCHAR(42) NOT NULL,
    token_symbol VARCHAR(16) NOT NULL,
    amount_decimal NUMERIC(36, 18) NOT NULL,
    amount_raw NUMERIC(78, 0) NOT NULL,
    recipient_address VARCHAR(42) NOT NULL,
    user_account_hash VARCHAR(66) NOT NULL,
    ccip_message_id VARCHAR(66) UNIQUE,
    source_tx_hash VARCHAR(66),
    destination_tx_hash VARCHAR(66),
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING_SUBMISSION', -- PENDING_SUBMISSION, IN_FLIGHT, COMMITTED_ON_DESTINATION, EXECUTED_SUCCESS, FAILED
    fee_paid_native NUMERIC(36, 18) NOT NULL DEFAULT 0,
    fee_paid_link NUMERIC(36, 18) NOT NULL DEFAULT 0,
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ccip_bridge_messages (
    message_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ccip_message_id VARCHAR(66) NOT NULL UNIQUE,
    source_chain_id BIGINT NOT NULL,
    destination_chain_id BIGINT NOT NULL,
    source_chain_selector BIGINT NOT NULL,
    dest_chain_selector BIGINT NOT NULL,
    sender_address VARCHAR(42) NOT NULL,
    receiver_address VARCHAR(42) NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    token_amount NUMERIC(78, 0) NOT NULL,
    fee_token VARCHAR(42) NOT NULL,
    fee_amount NUMERIC(78, 0) NOT NULL,
    source_block_number BIGINT NOT NULL,
    source_tx_hash VARCHAR(66) NOT NULL,
    destination_tx_hash VARCHAR(66),
    state VARCHAR(32) NOT NULL DEFAULT 'IN_FLIGHT', -- IN_FLIGHT, COMMITTED, EXECUTED, FAILED
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gas_rebasing_history (
    rebase_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id BIGINT NOT NULL,
    original_tx_hash VARCHAR(66) NOT NULL,
    replacement_tx_hash VARCHAR(66) NOT NULL,
    nonce BIGINT NOT NULL,
    old_max_fee_per_gas NUMERIC(36, 0) NOT NULL,
    new_max_fee_per_gas NUMERIC(36, 0) NOT NULL,
    old_priority_fee_per_gas NUMERIC(36, 0) NOT NULL,
    new_priority_fee_per_gas NUMERIC(36, 0) NOT NULL,
    bump_percentage NUMERIC(5, 2) NOT NULL,
    reason VARCHAR(64) NOT NULL, -- STUCK_IN_MEMPOOL, GAS_SPIKE_EXPEDITE
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE reorg_audit_records (
    reorg_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chain_id BIGINT NOT NULL REFERENCES evm_supported_chains(chain_id),
    detected_at_block BIGINT NOT NULL,
    common_ancestor_block BIGINT NOT NULL,
    reorg_depth INT NOT NULL,
    invalidated_block_hashes JSONB NOT NULL,
    affected_deposits_count INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evm_deposit_state ON evm_ingress_deposits(chain_id, state);
CREATE INDEX idx_evm_deposit_account ON evm_ingress_deposits(user_account_hash);
CREATE INDEX idx_evm_deposit_tx ON evm_ingress_deposits(tx_hash);
CREATE INDEX idx_evm_block_chain_num ON evm_monitored_blocks(chain_id, block_number DESC);
CREATE INDEX idx_ccip_msg_state ON ccip_bridge_messages(state);
CREATE INDEX idx_reorg_chain ON reorg_audit_records(chain_id, created_at DESC);
```

### Kafka Event Schemas

#### Topic: `wallet.evm_deposit.confirmed.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "EvmDepositConfirmedEvent",
  "type": "object",
  "required": [
    "eventId",
    "depositId",
    "chainId",
    "txHash",
    "logIndex",
    "blockNumber",
    "userAccountHash",
    "tokenSymbol",
    "tokenAddress",
    "amountDecimal",
    "amountRaw",
    "confirmedAtUnixMs"
  ],
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "depositId": { "type": "string", "format": "uuid" },
    "chainId": { "type": "integer", "enum": [1, 42161, 10, 8453, 137] },
    "txHash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "logIndex": { "type": "integer" },
    "blockNumber": { "type": "integer" },
    "userAccountHash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "tokenSymbol": { "type": "string", "enum": ["ETH", "USDC", "USDT", "DAI", "WBTC"] },
    "tokenAddress": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "amountDecimal": { "type": "string" },
    "amountRaw": { "type": "string" },
    "amlCheckStatus": { "type": "string", "enum": ["PASSED"] },
    "confirmedAtUnixMs": { "type": "integer" }
  }
}
```

#### Topic: `wallet.evm_deposit.reorg_invalidated.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "EvmDepositReorgInvalidatedEvent",
  "type": "object",
  "required": [
    "eventId",
    "depositId",
    "chainId",
    "invalidatedTxHash",
    "reorgDepth",
    "commonAncestorBlock",
    "invalidatedAtUnixMs"
  ],
  "properties": {
    "eventId": { "type": "string", "format": "uuid" },
    "depositId": { "type": "string", "format": "uuid" },
    "chainId": { "type": "integer" },
    "invalidatedTxHash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "reorgDepth": { "type": "integer" },
    "commonAncestorBlock": { "type": "integer" },
    "invalidatedAtUnixMs": { "type": "integer" }
  }
}
```

## Security & Compliance Notes
- **Reorg Protection & Asymmetric Confirmation Thresholds:** To protect against double-spend exploits and deep chain reorganizations, deposits are never credited immediately upon mempool detection. Ethereum Mainnet deposits mandate 64 block confirmations or execution of the `finalized` block tag; Polygon PoS mandates 128 block confirmations; Arbitrum, Optimism, and Base L2s mandate L1 batch publication.
- **EIP-712 Replay Attack Prevention:** All permit meta-transactions enforce strict EIP-712 domain separation including contract address, verifying contract name, and `chainId`. Relayer nonces and permit deadlines are validated both off-chain in Redis and on-chain prior to dispatch.
- **Chainlink CCIP Risk Management & Arm Network Verification:** Outbound and inbound cross-chain bridging operations require cryptographic attestation by the independent Chainlink CCIP Arm Network. Transactions flagged or paused by the Arm Network are halted immediately and escalated to platform compliance.
- **Relayer Key Protection via Cloud KMS / HSM:** All relayer signing keys for gas rebasing, permit submission, and lockbox operations are hosted in FIPS 140-2 Level 3 Hardware Security Modules. No unencrypted private keys exist in memory or container filesystems.
- **Zero PII on Public Ledgers:** In compliance with DPDP Act 2023 and global privacy standards, all on-chain deposit events, lockbox interactions, and CCIP message payloads use 32-byte cryptographic hashes of investor account identifiers.
- **Anti-Money Laundering (AML) Travel Rule Gate:** Every detected EVM deposit is asynchronously evaluated against on-chain AML sanctions oracles (sanctioned addresses, darknet mixers) before reaching `FINALIZED` status. Flagged deposits are quarantined in state `REJECTED_AML`.

## Acceptance Criteria
- [ ] Concurrently monitors and processes block streams across Ethereum Mainnet, Arbitrum One, Optimism, Base, and Polygon PoS with $< 500\text{ms}$ ingestion latency.
- [ ] Successfully parses and decodes deposit events for ETH, USDC, USDT, DAI, and WBTC from canonical lockbox smart contracts.
- [ ] Correctly identifies simulated block reorganizations (depths 1 to 5), invalidating orphaned deposit records and broadcasting `wallet.evm_deposit.reorg_invalidated.v1` events.
- [ ] Validates and executes EIP-712 / EIP-2612 and Permit2 gasless deposit meta-transactions through the sponsored relayer.
- [ ] Quotes, dispatches, and tracks cross-chain token transfers via Chainlink CCIP v1.5+ with automated LINK and native gas fee settlement.
- [ ] Dynamic gas rebasing engine identifies mempool transactions stuck for $> 60\text{s}$ and successfully broadcasts RBF replacement transactions with a $\ge 15\%$ gas price bump.
- [ ] Transactional outbox guarantees exactly-once publishing of deposit events to Kafka with zero message loss.
- [ ] Test coverage exceeds $\ge 85\%$ across all Go unit, watcher, and integration test suites.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Design Standards), Prompt 104 (Kafka Topic Standards), Prompt 109 (Secrets & HSM Architecture), Prompt 203 (Wallet & Account Ledger Service).
- **Parallel Tasks:** Prompt 212 (Banking & Payment Gateway Service), Prompt 214 (Foreign Investor Funding & FX Service), Prompt 308 (On-Chain Proof of Reserve Publishing).
- **Downstream Blockers:** Prompt 306 (Settlement DvP Smart Contract multi-chain liquidity integration), Prompt 319 (Institutional Custody Bridge).
