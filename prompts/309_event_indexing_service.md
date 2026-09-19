# 309 - Blockchain Event Indexing & Query Service

## Purpose
Directly querying distributed blockchain nodes for historical transactions, token transfer histories, user holdings, and settlement receipts is computationally expensive, lacks flexible filtering/sorting capabilities, and cannot support high-concurrency client application loads. To provide instantaneous user portfolio updates, audit log views, and downstream microservice synchronization, Growww requires a dedicated, resilient blockchain event indexing pipeline.

This prompt specifies the architecture, implementation, and deployment of the **Blockchain Event Indexer & Query Service** (`services/event-indexer/`). The service maintains an active WebSocket / JSON-RPC subscription to Hyperledger Besu nodes, ingests raw block headers and log receipts, decodes application contract events (`Transfer`, `TokensMintedWithCustodyProof`, `DvPSettlementExecuted`, `ReserveAttestationPublished`), populates an indexed PostgreSQL operational datastore, and publishes canonical domain events to Apache Kafka topics.

## What You Are Building
A high-throughput, fault-tolerant indexing service in Rust or Go under `services/event-indexer/` containing:
- Block & Log Ingestion Pipeline: Subscribes to `eth_subscribe("newHeads")` and `eth_getLogs` with automatic exponential reconnection and block gap backfilling.
- ABI Decoder Engine: High-performance ABI unpacker decoding ERC-3643, DvP settlement, compliance, and proof-of-reserve smart contract logs.
- Checkpoint & State Tracking: Transactional database cursor ensuring exactly-once event processing and zero missing blocks across service restarts.
- Kafka Event Publisher: Emits strongly typed Avro/JSON events to Kafka topics (e.g. `growww.ledger.tokens.transferred`, `growww.ledger.settlement.executed`).
- High-Performance Query API: REST and GraphQL endpoints exposing fast historical queries for mobile/web client apps and internal services.

## Scope Boundaries
- **In Scope:**
 - Real-time ingestion of Besu blocks and transaction logs.
 - Historical backfill from block 0 to current block height.
 - Decoding event ABIs for all Growww smart contracts.
 - Storing indexed transactions, transfers, and settlements in PostgreSQL.
 - Streaming normalized ledger events to Apache Kafka.
 - GraphQL and REST query endpoints for token balances and trade histories.
- **Out of Scope / Handled Elsewhere:**
 - Order Matching and Pre-trade validation (handled in Prompts 204 & 205).
 - PostgreSQL schema master definitions (handled in Prompt 401).
 - Kafka cluster infrastructure setup (handled in Prompt 403).

## Technology to Use
- **Language:** **Rust (v1.78+) or Go (v1.22+)**.
  *Justification:* Rust (using `alloy-rs` / `ethers-rs`) or Go (using `go-ethereum/ethclient`) provides zero-cost abstractions, deterministic memory management, ultra-low processing latency (<5ms per block), and exceptional concurrency handling for stream processing.
- **Database:** PostgreSQL 16 (with timescale/partitioned event tables).
- **Messaging:** Apache Kafka with Confluent Schema Registry.
- **API Layer:** Axum / Tonic (Rust) or Gin / Chi (Go) + async GraphQL engine.
- **Observability:** OpenTelemetry, Prometheus metrics exporter.

## Backend / Infra Touchpoints
- **Hyperledger Besu RPC Nodes:** Connects to internal RPC relay nodes (Prompt 302) via WebSocket (`ws://besu-rpc:8546`) and HTTP (`http://besu-rpc:8545`).
- **PostgreSQL 16:** Relational database storing normalized ledger state (`ledger_events`, `token_transfers`, `settlement_records`, `indexer_checkpoints`).
- **Apache Kafka:** Event streaming bus broadcasting ledger updates to Portfolio Service (Prompt 209), Wallet Service (Prompt 203), and Audit Service (Prompt 218).

## Blockchain Interaction
- **Ingestion Mechanisms:** Subscribes to `eth_subscribe("logs", {"address": [...]})` and polls `eth_getBlockByNumber` for full receipt batches.
- **Processed Contract Events:**
 - `DigitalSecurityToken.sol`: `Transfer`, `TokensMintedWithCustodyProof`, `AddressFrozen`.
 - `SettlementDvP.sol`: `DvPSettlementExecuted`, `BatchSettled`, `TradeCancelled`.
 - `IdentityRegistry.sol`: `IdentityRegistered`, `IdentityRemoved`.
 - `ProofOfReserveRegistry.sol`: `ReserveAttestationPublished`, `SolvencyBreachAlert`.
- **Zero PII Guarantee:** The indexing database mirrors strictly on-chain data: addresses, transaction hashes, block numbers, amounts, and metadata hashes. PII correlation is resolved exclusively within isolated off-chain user services.

## Step-by-Step Build Instructions
1. Initialize project under `services/event-indexer/` (Rust Cargo workspace or Go module).
2. Generate type-safe contract bindings from compiled Solidity ABIs using `alloy::sol!` macro (Rust) or `abigen` (Go).
3. Design PostgreSQL schema migrations (`migrations/001_ledger_indexer_tables.sql`):
 - `indexer_checkpoints` (`network_id`, `last_indexed_block`, `updated_at`)
 - `indexed_blocks` (`block_number`, `block_hash`, `parent_hash`, `timestamp`, `tx_count`)
 - `token_transfers` (`tx_hash`, `log_index`, `block_number`, `token_address`, `from_address`, `to_address`, `amount`, `timestamp`)
 - `dvp_settlements` (`tx_hash`, `trade_id`, `token_address`, `buyer_address`, `seller_address`, `token_amount`, `inr_amount`, `fee_amount`, `timestamp`)
 - `por_attestations` (`tx_hash`, `attestation_id`, `merkle_root`, `timestamp`, `assets_count`)
4. Implement the Besu WebSocket connection manager with auto-reconnection and health heartbeats (`eth_blockNumber`).
5. Implement the block backlog catchup scanner: on startup, read `last_indexed_block` from PostgreSQL; if `current_chain_height > last_indexed_block`, paginate in 500-block chunks via `eth_getLogs` to index all historical events.
6. Implement real-time live streaming engine: listen to new block headers, fetch transaction receipts, and dispatch raw logs to the worker pool.
7. Implement log dispatcher: route logs by contract address and topic signature to specific event decoders.
8. Implement atomic database persistence: wrap all decoded events for a block and the updated `indexer_checkpoints` record in a single PostgreSQL database transaction (`BEGIN...COMMIT`).
9. Implement Kafka publisher pipeline: serialize decoded events to JSON/Avro and publish asynchronously to appropriate Kafka topics with partition keys matching `isin` or `user_address`.
10. Implement GraphQL / REST query service exposing:
 - `getTransfersByAddress(address, limit, cursor)`
 - `getSettlementByTradeId(tradeId)`
 - `getTokenHolders(tokenAddress, limit)`
 - `getLatestProofOfReserve()`
11. Instrument indexing metrics: track `blocks_indexed_total`, `event_processing_latency_ms`, `chain_head_lag_blocks`, and `db_write_latency_ms` in Prometheus.
12. Write comprehensive unit tests for ABI decoding and log parsing.
13. Write end-to-end integration tests using a local test Besu node: emit 1,000 synthetic events and verify 100% data consistency in PostgreSQL and Kafka within 500ms.
14. Build multi-stage Docker container with non-root security context.
15. Deploy to Kubernetes with horizontal read replicas for GraphQL queries and a single active leader for the ingestion pipeline.

## Interfaces / Contracts

### Indexer Database Schema Sketch (PostgreSQL)
```sql
CREATE TABLE indexer_checkpoints (
    chain_id BIGINT PRIMARY KEY,
    last_indexed_block BIGINT NOT NULL,
    last_indexed_hash VARCHAR(66) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE token_transfers (
    id BIGSERIAL PRIMARY KEY,
    tx_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    block_number BIGINT NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    from_address VARCHAR(42) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    amount NUMERIC(38, 18) NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tx_log UNIQUE (tx_hash, log_index)
);
CREATE INDEX idx_transfers_from_to ON token_transfers(from_address, to_address);
CREATE INDEX idx_transfers_token_block ON token_transfers(token_address, block_number DESC);

CREATE TABLE dvp_settlements (
    trade_id VARCHAR(66) PRIMARY KEY,
    tx_hash VARCHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    token_address VARCHAR(42) NOT NULL,
    buyer_address VARCHAR(42) NOT NULL,
    seller_address VARCHAR(42) NOT NULL,
    token_amount NUMERIC(38, 18) NOT NULL,
    inr_gross_amount NUMERIC(18, 2) NOT NULL,
    realized_profit NUMERIC(18, 2) NOT NULL,
    fee_amount NUMERIC(18, 2) NOT NULL,
    settled_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_settlements_buyer_seller ON dvp_settlements(buyer_address, seller_address);
```

### GraphQL Query Schema Definition
```graphql
type TokenTransfer {
  txHash: String!
  logIndex: Int!
  blockNumber: Int!
  tokenAddress: String!
  fromAddress: String!
  toAddress: String!
  amount: String!
  timestamp: String!
}

type DvPSettlement {
  tradeId: String!
  txHash: String!
  tokenAddress: String!
  buyerAddress: String!
  sellerAddress: String!
  tokenAmount: String!
  inrGrossAmount: Float!
  feeAmount: Float!
  settledAt: String!
}

type Query {
  transfers(address: String!, limit: Int = 50, offset: Int = 0): [TokenTransfer!]!
  settlement(tradeId: String!): DvPSettlement
  chainSyncStatus: SyncStatus!
}

type SyncStatus {
  currentChainHead: Int!
  lastIndexedBlock: Int!
  lagBlocks: Int!
  isSynced: Boolean!
}
```

## Security & Compliance Notes
- **Idempotency & Reorg Safety:** While QBFT consensus provides instant 1-block finality with zero chain reorganizations, the indexer verifies block parent hashes and enforces database unique constraints (`UNIQUE(tx_hash, log_index)`) to prevent duplicate insertions.
- **Read-Only Ledger Access:** The indexer connects via read-only JSON-RPC credentials and has zero access to signing keys or private keys.
- **Data Protection:** No personal identifiable information is ingested or served by this service; all records represent pseudonymous on-chain cryptographic state.

## Acceptance Criteria
- [ ] Indexer service boots, connects to Besu node, and completes backfill of all past blocks without error.
- [ ] Real-time block latency from block proposal to PostgreSQL commit is $<200\text{ ms}$.
- [ ] 100% of ERC-3643, DvP settlement, and Proof-of-Reserve events decoded accurately.
- [ ] Kafka events are published reliably with correct topic schemas.
- [ ] GraphQL query API returns paginated transfer history in $<50\text{ ms}$.
- [ ] Integration test simulates node disconnection and verifies automatic recovery and gap filling.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `302` (Network Topology & Validator Setup), Prompt `303` (Token Issuance), Prompt `306` (Settlement DvP), Prompt `401` (PostgreSQL Schema).
- **Parallel Tasks:** Prompt `403` (Kafka Cluster Design), Prompt `209` (Portfolio Service).
- **Subsequent Prompts Enabled:** Prompt `512` (Flutter Trade History Screen), Prompt `603` (Web Trading Dashboard).
