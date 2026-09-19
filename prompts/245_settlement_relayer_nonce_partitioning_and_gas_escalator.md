# 245 - Settlement Relayer Nonce Partitioning & Gas Escalator (Go)

## Purpose
In high-throughput tokenized capital markets operating on Ethereum Virtual Machine (EVM) ledgers, the sequential account nonce constraint ($0, 1, 2, \dots$) creates a severe structural bottleneck. If a single relayer account broadcasts all Delivery-versus-Payment (DvP) settlement transactions, any unmined, pending, or temporarily stuck transaction blocks all subsequent transactions behind it (head-of-line blocking). Under CPMI-IOSCO Principle 8 (Settlement Finality) and SEBI regulatory settlement cycles, the National Blockchain Stock Exchange (NBSE) requires deterministic, sub-second settlement finality across thousands of concurrent trades without single-point queue blockages.

The **Settlement Relayer Nonce Partitioning & Gas Escalator** service (`services/settlement-relayer`) resolves this concurrency bottleneck by sharding transaction dispatch across 32 dedicated on-chain relayer accounts (`relayer_00` to `relayer_31`). The service uses 32-bit Murmur3 hashing on the security's International Securities Identification Number (ISIN) to deterministically map trades for a given asset to a dedicated relayer partition. Nonce allocation is coordinated through atomic Redis sequence queues backed by PostgreSQL audit ledgers, with transaction signing delegated to Hardware Security Modules (AWS CloudHSM / HashiCorp Vault). An autonomous gas escalator monitors unmined transactions against strict QBFT block confirmation SLAs, executing automatic replacement transactions with a 25% gas price escalation ($\text{GasPrice} \times 1.25$) to eliminate stuck transactions and guarantee uninterrupted settlement flow.

## What You Are Building
A mission-critical, ultra-low-latency Go microservice (`services/settlement-relayer`) managing distributed EVM transaction relaying, nonce partitioning, and dynamic gas pricing for Hyperledger Besu. Concrete deliverables include:
- **32-Partition Relayer Account Pool:** A managed pool of 32 whitelisted EVM relayer addresses (`relayer_00` through `relayer_31`), each backed by a dedicated secp256k1 ECDSA private key secured in AWS CloudHSM or HashiCorp Vault Transit Engine.
- **Murmur3 ISIN Sharding Router:** A deterministic partitioning dispatcher computing $\text{partition\_id} = \text{Murmur3\_32}(\text{isin}) \pmod{32}$ to route all settlement transactions for an ISIN to a dedicated relayer queue, ensuring serial ordering per security while achieving 32x parallel throughput across different securities.
- **Atomic Redis Nonce Sequence Queue:** A high-speed, Lua-orchestrated nonce state machine in Redis 7.2 guaranteeing zero-gap, monotonic nonce issuance, in-flight transaction tracking, lease-based recovery, and atomic synchronization with Hyperledger Besu transaction pools.
- **Dynamic 1.25x Gas Escalator Worker:** An automated background watchdog that detects transactions unmined after a configurable block threshold ($T_{\text{timeout}} = 4\text{ seconds}$ / 2 QBFT blocks) and broadcasts replacement transactions with matching nonces and escalated gas parameters ($\lceil \text{GasPrice} \times 1.25 \rceil$).
- **CloudHSM / Vault Remote Transaction Signer:** An asynchronous signing pipeline communicating with HSM/Vault over mTLS to produce EIP-1559 and legacy EVM raw signed transactions without exposing private key material to application memory.
- **Hyperledger Besu QBFT Block Finality Monitor:** A dual-mode JSON-RPC / WebSocket block listener subscribing to new block headers, verifying transaction receipts, confirming single-block QBFT deterministic finality, and publishing completion events.
- **Resynchronization & Nonce Recovery Daemon:** A self-healing background worker that polls `eth_getTransactionCount` across `latest` and `pending` states to detect and repair nonce drift, transaction drops, or network split artifacts.
- **Kafka Event Streaming Integration:** Consumer for incoming settlement dispatch requests (`settlement.tx_requested.v1`) and publisher for transaction lifecycle events (`relayer.tx_broadcast.v1`, `relayer.tx_mined.v1`, `relayer.tx_escalated.v1`, `relayer.tx_failed.v1`).

## Scope Boundaries
- **In Scope:**
  - Managing the pool of 32 relayer addresses (`relayer_00` to `relayer_31`) and their respective nonces.
  - Murmur3 32-bit hash partitioning based on asset ISIN strings.
  - Atomic Redis sequence queues with Lua scripts for nonce reservation, committing, and leasing.
  - CloudHSM / HashiCorp Vault Transit Engine signing integration for secp256k1 raw transactions.
  - Autonomous gas price escalator implementing the 25% bump rule ($\text{GasPrice} \times 1.25$) on unmined transactions.
  - Multi-node Hyperledger Besu RPC dispatch with automatic failover and load balancing.
  - Real-time QBFT block header ingestion and receipt confirmation.
  - Nonce gap detection, stuck transaction replacement, and automated resynchronization.
  - PostgreSQL persistence for relayer state, transaction lifecycle, gas escalation logs, and audit metrics.
  - Kafka event consumption and production.
  - gRPC endpoints for settlement dispatch and relayer health telemetry.
- **Out of Scope / Handled Elsewhere:**
  - Matching engine limit order book trade execution (handled in Prompt 205).
  - Off-chain double-entry fiat INR cash ledger debits/credits (handled in Prompt 203).
  - Solidity smart contract implementation for `SettlementDvP.sol` (handled in Prompt 306).
  - SGF default waterfall execution and clearing fund margin calls (handled in Prompt 230).
  - Physical custodian depository reconciliation with NSDL/CDSL (handled in Prompt 213).
  - Token minting, burning, and corporate actions execution (handled in Prompt 222).

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+. Chosen for high-concurrency goroutine scheduling, deterministic memory management, low garbage collection pauses, and battle-tested EVM client libraries.
- **EVM Blockchain Client:** `github.com/ethereum/go-ethereum` (`ethclient`, `core/types`, `common`, `crypto`, `rlp`) configured for Hyperledger Besu JSON-RPC and WebSocket endpoints over mTLS.
- **Sharding & Hashing Library:** `spaolacci/murmur3` or `twmb/murmur3` for 32-bit Murmur3 non-cryptographic high-speed hashing.
- **In-Memory Nonce Store & Sequence Queues:** **Redis 7.2+ Cluster** using `go-redis/v9` with pre-loaded Lua scripts (`EVALSHA`) for atomic nonce allocation and in-flight lease management.
- **Cryptographic Key Management:** **AWS CloudHSM** / **HashiCorp Vault Transit Engine** via PKCS#11 / REST API with secp256k1 ECDSA signing keys. Private keys never leave the secure enclave.
- **Relational Persistence:** **PostgreSQL 16+** using `pgx/v5` connection pooling and `sqlc` compile-time type-safe query generation.
- **Message Broker & Event Streaming:** **Apache Kafka 3.7+** with KRaft using `segmentio/kafka-go` with manual partition offset commits.
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** with mTLS for secure, low-latency RPC invocations.
- **Observability & Metrics:** **Prometheus** client (`prometheus/client_golang`) and **OpenTelemetry** Go SDK for distributed tracing across Kafka, Redis, HSM, and Besu nodes.

## Backend / Infra Touchpoints
- **Hyperledger Besu Validator Cluster:**
  - Primary and backup JSON-RPC / WebSocket endpoints over mTLS (e.g., `https://besu-validator-01.internal:8545`, `wss://besu-validator-01.internal:8546`).
  - QBFT consensus network running with 2-second block period and 0 base fee or dynamic fee market.
- **Redis 7.2 Key Schema:**
  - `relayer:partition:{id}:confirmed_nonce`: Integer string storing the highest contiguous mined nonce confirmed on-chain.
  - `relayer:partition:{id}:next_nonce`: Integer string storing the next assignable in-flight nonce sequence.
  - `relayer:partition:{id}:inflight:{nonce}`: Redis Hash containing active transaction metadata (`tx_hash`, `trade_id`, `gas_price`, `submit_timestamp`, `escalation_count`).
  - `relayer:partition:{id}:queue`: Redis Sorted Set containing queued settlement jobs scored by arrival timestamp.
  - `relayer:partition:{id}:lock`: Distributed lock key (Redlock pattern) for nonce resync operations.
- **PostgreSQL 16 Tables:**
  - `relayer_accounts`: Metadata, Ethereum address, partition index (0 to 31), HSM key alias, active status.
  - `relayer_nonce_sequences`: Monotonic sequence ledger tracking current, pending, and reserved nonces per partition.
  - `settlement_transactions`: Comprehensive transaction lifecycle records, including trade ID, batch ID, ISIN, partition ID, nonce, gas parameters, raw payload, status, and mined block numbers.
  - `gas_escalation_history`: Detailed audit log of every gas escalation attempt, old vs new gas prices, replacement transaction hashes, and reason codes.
  - `relayer_health_snapshots`: Periodic health snapshots capturing pending queue depth, in-flight count, RPC latencies, and error rates.
- **AWS CloudHSM / HashiCorp Vault Transit Engine:**
  - Key aliases: `growww-settlement-relayer-00` through `growww-settlement-relayer-31`.
- **Apache Kafka Topics:**
  - Consumes: `settlement.tx_requested.v1` (from Trade Settlement Service, Prompt 208).
  - Publishes: `relayer.tx_broadcast.v1`, `relayer.tx_mined.v1`, `relayer.tx_escalated.v1`, `relayer.tx_failed.v1`.
- **Trade Settlement & DvP Service (Prompt 208):** Ingests transaction status updates and finality receipts to transition settlement state machines.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Network Topology:** Hyperledger Besu running Enterprise Istanbul/QBFT Byzantine Fault Tolerant consensus with 2-second deterministic block finality (1 block confirmation requirement, zero reorg risk past finality).
- **Relayer Whitelist Registry:** All 32 relayer addresses (`0xRelayer00...` to `0xRelayer31...`) are registered on-chain in `RelayerAccessRegistry.sol` (Prompt 306/315) and granted the `AUTHORIZED_RELAYER_ROLE` on `SettlementDvP.sol`.
- **Smart Contract Interfacing:**
  - `SettlementDvP.executeDvPTrade(bytes32 tradeId, address buyer, address seller, address tokenContract, uint256 tokenUnitsRaw, uint256 inrAmountPaise, bytes relayerSig)`
  - `SettlementDvP.executeDvPBatch(bytes32 batchId, DvPExecutionParams[] trades)`
- **Transaction Formulation:** Constructs EIP-1559 (`DynamicFeeTx`) or legacy (`LegacyTx`) raw transactions with:
  - `ChainID`: Configured Besu network chain ID (e.g., `1337` or `2026`).
  - `Nonce`: Assigned partition nonce from Redis sequence queue.
  - `GasLimit`: Calculated based on trade/batch size (e.g., 250,000 gas per single trade, 2,500,000 per batch of 20).
  - `GasPrice` / `MaxFeePerGas`: Base network fee plus escalator multiplier.
  - `To`: Target `SettlementDvP.sol` contract address.
  - `Data`: ABI-encoded method call payload.
- **Gas Escalator Replacement Mechanics:**
  - An unmined transaction is replaced by creating a new raw transaction with the identical `Nonce`, identical `To`, identical `Data`, identical `Value`, and updated `GasPrice`:
    $$\text{GasPrice}_{\text{new}} = \max\left(\lceil \text{GasPrice}_{\text{current}} \times 1.25 \rceil, \text{BesuBaseFee} \times 1.25\right)$$
  - The replacement transaction is signed via CloudHSM/Vault and broadcast via `eth_sendRawTransaction`. The old transaction hash is superseded on-chain once the replacement is included in a block.
- **Zero On-Chain PII Invariant:** Transactions contain solely pseudonymous Ethereum addresses (`buyer_address`, `seller_address`), security token contract addresses, monetary integer values (paise / token units), and cryptographic hashes. Zero customer PANs, legal names, or banking details are ever broadcast or stored on-chain.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Directory & Layout:** Initialize Go module `services/settlement-relayer` following clean architecture principles (`cmd/relayer/`, `internal/partitioner/`, `internal/nonce/`, `internal/signer/`, `internal/escalator/`, `internal/besu/`, `internal/store/`, `internal/kafka/`, `internal/grpc/`).
2. **Define Protobuf Specifications:** Author `proto/growww/relayer/v1/settlement_relayer.proto` defining RPCs `SubmitSettlementTx`, `GetTransactionStatus`, `GetRelayerPoolStatus`, and `ResyncRelayerNonce`. Generate Go stubs using `protoc-gen-go` and `protoc-gen-go-grpc`.
3. **Design Database Schema Migrations:** Author PostgreSQL migrations creating tables `relayer_accounts`, `relayer_nonce_sequences`, `settlement_transactions`, `gas_escalation_history`, and `relayer_health_snapshots` with appropriate indices and foreign key constraints.
4. **Implement Murmur3 ISIN Partitioner:**
   - Build `ISINPartitioner` module implementing 32-bit Murmur3 hashing on normalized 12-character ISIN strings (e.g., `INE002A01018`).
   - Compute partition index:
     $$\text{partition\_id} = \text{Murmur3\_32}(\text{isin}) \pmod{32}$$
   - Map `partition_id` ($0 \dots 31$) to the corresponding relayer account configuration (`relayer_00` to `relayer_31`).
5. **Implement Redis Atomic Nonce Sequence Engine:**
   - Author and load Redis Lua scripts for atomic operations:
     * `reserve_nonce.lua`: Atomically reads `next_nonce`, increments `next_nonce` by 1, writes in-flight tracking record to `relayer:partition:{id}:inflight:{nonce}`, and returns reserved nonce.
     * `commit_nonce.lua`: Upon block confirmation, advances `confirmed_nonce` if contiguous, deletes in-flight tracking record, and prunes completed queue entries.
     * `rollback_nonce.lua`: Upon fatal submission rejection before broadcast, handles nonce quarantine or rollback if no subsequent nonces have been issued.
   - Configure connection pooling and failover for Redis Cluster.
6. **Implement CloudHSM / Vault Remote Signer:**
   - Build `HSMSigner` interface abstracting AWS CloudHSM PKCS#11 and HashiCorp Vault Transit Engine REST APIs.
   - For an unsigned EVM transaction, compute its RLP-encoded signing hash (Keccak-256).
   - Dispatch signing hash to HSM/Vault key alias `growww-settlement-relayer-{partition_id}` to obtain secp256k1 signature tuple $(R, S, V)$.
   - Assemble final raw signed transaction RLP bytes.
7. **Implement Hyperledger Besu Dispatcher & Node Client:**
   - Build connection manager maintaining active JSON-RPC and WebSocket connections to Besu validator nodes.
   - Implement `BroadcastRawTx(ctx, signedTxBytes)` sending raw transactions via `eth_sendRawTransaction`.
   - Handle immediate RPC responses: accept `tx_hash`, detect `nonce too low` or `already known` errors, and handle node disconnects with automatic fallback to secondary nodes.
8. **Implement QBFT Block Confirmation Monitor:**
   - Open WebSocket subscription to `newHeads` on Besu nodes.
   - For every incoming block header, scan transaction hashes against active in-flight Redis records.
   - Query `eth_getTransactionReceipt` to verify status (`0x1` for success, `0x0` for smart contract revert).
   - Trigger `commit_nonce.lua`, update PostgreSQL transaction record to `MINED` / `FINALIZED`, and emit `relayer.tx_mined.v1` to Kafka.
9. **Implement Autonomous Gas Escalator Watchdog:**
   - Build periodic ticker routine (e.g., every 1000ms) inspecting all in-flight transactions in Redis.
   - Calculate elapsed time since submission: $\Delta t = t_{\text{now}} - t_{\text{submitted}}$.
   - If $\Delta t \ge T_{\text{timeout}}$ ($4\text{ seconds}$) and transaction remains unmined:
     * Check current escalation count against max threshold ($N_{\text{max}} = 4$).
     * Compute escalated gas price: $\text{GasPrice}_{\text{new}} = \lceil \text{GasPrice}_{\text{current}} \times 1.25 \rceil$.
     * Verify $\text{GasPrice}_{\text{new}} \le \text{MaxGasPriceCap}$.
     * Create replacement transaction with matching nonce, re-sign via CloudHSM/Vault, and broadcast.
     * Record escalation entry in `gas_escalation_history`, update Redis in-flight record, and publish `relayer.tx_escalated.v1`.
10. **Implement Nonce Resynchronization & Self-Healing Daemon:**
    - Build background recovery worker polling Besu RPC `eth_getTransactionCount` for each of the 32 relayer addresses across `latest` and `pending` states.
    - If `pending_nonce` exceeds `redis_confirmed_nonce` due to service restart or external transactions, reconcile Redis sequence counters.
    - If an in-flight transaction is dropped from Besu txpool without being mined, trigger re-broadcast or nonce gap bridging.
11. **Build Kafka Consumer for Settlement Requests:**
    - Subscribe to Kafka topic `settlement.tx_requested.v1`.
    - Extract ISIN, compute Murmur3 partition, acquire reserved nonce from Redis, sign transaction payload via HSM, broadcast to Besu, update PostgreSQL, and publish `relayer.tx_broadcast.v1`.
12. **Implement gRPC API Server:**
    - Implement gRPC methods for synchronous trade settlement submission, transaction receipt querying, partition status inspections, and manual operator nonce resync commands.
13. **Configure Telemetry, Logging & Prometheus Metrics:**
    - Instrument Prometheus counters and histograms: `settlement_relayer_tx_submitted_total`, `settlement_relayer_tx_mined_total`, `settlement_relayer_tx_escalated_total`, `settlement_relayer_escalation_rounds_histogram`, `settlement_relayer_mining_latency_seconds`, `settlement_relayer_hsm_signing_latency_ms`, `settlement_relayer_nonce_gap_events_total`.
    - Configure structured JSON logging with Zap/Zerolog containing `partition_id`, `relayer_address`, `isin`, `trade_id`, `nonce`, and `tx_hash`.
14. **Perform Comprehensive Resilience & Load Testing:**
    - Execute end-to-end integration tests simulating 32 concurrent ISIN partition streams delivering 5,000 DvP settlements per second.
    - Test chaos scenarios: sudden Besu validator node restarts, simulated dropped transactions triggering 1.25x gas escalation, HSM transient signing latency, and Redis node failover.

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/relayer/v1/settlement_relayer.proto`)
```protobuf
syntax = "proto3";

package growww.relayer.v1;

option go_package = "growww/relayer/v1;relayerv1";

service SettlementRelayerService {
  rpc SubmitSettlementTx (SubmitSettlementTxRequest) returns (SubmitSettlementTxResponse);
  rpc GetTransactionStatus (GetTransactionStatusRequest) returns (GetTransactionStatusResponse);
  rpc GetRelayerPoolStatus (GetRelayerPoolStatusRequest) returns (GetRelayerPoolStatusResponse);
  rpc ResyncRelayerNonce (ResyncRelayerNonceRequest) returns (ResyncRelayerNonceResponse);
}

enum TransactionState {
  TRANSACTION_STATE_UNSPECIFIED = 0;
  TRANSACTION_STATE_QUEUED = 1;
  TRANSACTION_STATE_NONCE_ASSIGNED = 2;
  TRANSACTION_STATE_SIGNED = 3;
  TRANSACTION_STATE_BROADCAST = 4;
  TRANSACTION_STATE_ESCALATED = 5;
  TRANSACTION_STATE_MINED = 6;
  TRANSACTION_STATE_FINALIZED = 7;
  TRANSACTION_STATE_REVERTED = 8;
  TRANSACTION_STATE_FAILED_EXHAUSTED = 9;
}

message SubmitSettlementTxRequest {
  string request_id = 1;
  string trade_id = 2;
  string settlement_batch_id = 3;
  string isin = 4;
  string contract_address = 5;
  bytes call_data = 6;
  uint64 gas_limit = 7;
  string initial_gas_price_wei = 8;
  int64 priority = 9;
}

message SubmitSettlementTxResponse {
  string request_id = 1;
  string trade_id = 2;
  uint32 partition_id = 3;
  string relayer_address = 4;
  uint64 assigned_nonce = 5;
  string initial_tx_hash = 6;
  TransactionState state = 7;
  int64 submitted_at_unix_ns = 8;
}

message GetTransactionStatusRequest {
  string trade_id = 1;
  string tx_hash = 2;
}

message EscalationRecord {
  uint32 escalation_round = 1;
  string tx_hash = 2;
  string gas_price_wei = 3;
  int64 escalated_at_unix_ns = 4;
  string reason = 5;
}

message GetTransactionStatusResponse {
  string trade_id = 1;
  string settlement_batch_id = 2;
  string isin = 3;
  uint32 partition_id = 4;
  string relayer_address = 5;
  uint64 nonce = 6;
  string current_tx_hash = 7;
  repeated EscalationRecord escalation_history = 8;
  TransactionState state = 9;
  uint64 block_number = 10;
  string block_hash = 11;
  uint64 gas_used = 12;
  bool is_reverted = 13;
  string revert_reason = 14;
  int64 created_at_unix_ns = 15;
  int64 finalized_at_unix_ns = 16;
}

message PartitionHealth {
  uint32 partition_id = 1;
  string relayer_address = 2;
  uint64 confirmed_nonce = 3;
  uint64 next_assignable_nonce = 4;
  uint32 pending_in_flight_count = 5;
  uint32 queued_job_count = 6;
  bool is_healthy = 7;
  int64 last_resync_at_unix_ns = 8;
}

message GetRelayerPoolStatusRequest {}

message GetRelayerPoolStatusResponse {
  repeated PartitionHealth partitions = 1;
  uint32 total_active_partitions = 2;
  uint32 total_in_flight_transactions = 3;
  int64 observed_at_unix_ns = 4;
}

message ResyncRelayerNonceRequest {
  uint32 partition_id = 1;
  bool force_override = 2;
}

message ResyncRelayerNonceResponse {
  uint32 partition_id = 1;
  string relayer_address = 2;
  uint64 previous_redis_nonce = 3;
  uint64 besu_latest_nonce = 4;
  uint64 besu_pending_nonce = 5;
  uint64 synced_nonce = 6;
  bool success = 7;
  string message = 8;
}
```

### PostgreSQL Database Schema DDL
```sql
-- Relayer account partition registry
CREATE TABLE relayer_accounts (
    partition_id SMALLINT PRIMARY KEY CHECK (partition_id >= 0 AND partition_id <= 31),
    account_alias VARCHAR(50) NOT NULL UNIQUE, -- e.g., 'relayer_00' to 'relayer_31'
    ethereum_address CHAR(42) NOT NULL UNIQUE, -- e.g., '0x1234567890123456789012345678901234567890'
    hsm_key_alias VARCHAR(100) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Monotonic nonce sequence state per partition
CREATE TABLE relayer_nonce_sequences (
    partition_id SMALLINT PRIMARY KEY REFERENCES relayer_accounts(partition_id),
    confirmed_nonce BIGINT NOT NULL DEFAULT 0,
    reserved_nonce BIGINT NOT NULL DEFAULT 0,
    last_resync_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_nonce_ordering CHECK (reserved_nonce >= confirmed_nonce)
);

-- Settlement transaction lifecycle records
CREATE TABLE settlement_transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id VARCHAR(64) NOT NULL UNIQUE,
    trade_id UUID NOT NULL,
    settlement_batch_id UUID,
    isin VARCHAR(12) NOT NULL,
    partition_id SMALLINT NOT NULL REFERENCES relayer_accounts(partition_id),
    relayer_address CHAR(42) NOT NULL,
    nonce BIGINT NOT NULL,
    contract_address CHAR(42) NOT NULL,
    call_data BYTEA NOT NULL,
    gas_limit BIGINT NOT NULL,
    initial_gas_price_wei NUMERIC(36, 0) NOT NULL,
    current_gas_price_wei NUMERIC(36, 0) NOT NULL,
    current_tx_hash CHAR(66) NOT NULL,
    escalation_count INT NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'QUEUED',
    block_number BIGINT,
    block_hash CHAR(66),
    gas_used BIGINT,
    is_reverted BOOLEAN NOT NULL DEFAULT FALSE,
    revert_reason TEXT,
    broadcast_at TIMESTAMPTZ,
    mined_at TIMESTAMPTZ,
    finalized_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_partition_nonce UNIQUE (partition_id, nonce)
);

-- Gas escalation audit log
CREATE TABLE gas_escalation_history (
    escalation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES settlement_transactions(transaction_id) ON DELETE CASCADE,
    escalation_round INT NOT NULL,
    previous_tx_hash CHAR(66) NOT NULL,
    replacement_tx_hash CHAR(66) NOT NULL,
    previous_gas_price_wei NUMERIC(36, 0) NOT NULL,
    escalated_gas_price_wei NUMERIC(36, 0) NOT NULL, -- GasPrice * 1.25
    multiplier_applied NUMERIC(4, 2) NOT NULL DEFAULT 1.25,
    escalation_reason VARCHAR(100) NOT NULL, -- e.g., 'UNMINED_TIMEOUT_4S'
    broadcast_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Periodic relayer partition health snapshots
CREATE TABLE relayer_health_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partition_id SMALLINT NOT NULL REFERENCES relayer_accounts(partition_id),
    confirmed_nonce BIGINT NOT NULL,
    reserved_nonce BIGINT NOT NULL,
    in_flight_count INT NOT NULL,
    queued_count INT NOT NULL,
    rpc_latency_ms NUMERIC(8, 2) NOT NULL,
    is_healthy BOOLEAN NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indices for rapid querying and indexing
CREATE INDEX idx_settlement_tx_trade ON settlement_transactions(trade_id);
CREATE INDEX idx_settlement_tx_isin ON settlement_transactions(isin);
CREATE INDEX idx_settlement_tx_status ON settlement_transactions(status);
CREATE INDEX idx_settlement_tx_partition_status ON settlement_transactions(partition_id, status);
CREATE INDEX idx_settlement_tx_current_hash ON settlement_transactions(current_tx_hash);
CREATE INDEX idx_gas_esc_tx_id ON gas_escalation_history(transaction_id);
CREATE INDEX idx_relayer_health_part_time ON relayer_health_snapshots(partition_id, observed_at DESC);
```

### Redis Key Schema & Lua Script Specifications

#### Redis Key Conventions
- `relayer:p:{id}:confirmed_nonce`: Integer storing highest mined nonce.
- `relayer:p:{id}:next_nonce`: Integer storing next available nonce to assign.
- `relayer:p:{id}:inflight:{nonce}`: Hash storing transaction metadata (`tx_hash`, `trade_id`, `gas_price`, `submit_time_ms`, `esc_count`).
- `relayer:p:{id}:inflight_set`: Sorted Set containing all active nonces scored by submission timestamp.
- `relayer:p:{id}:lock`: Distributed lock for synchronizing with Besu JSON-RPC.

#### Lua Script: `reserve_nonce.lua`
```lua
-- KEYS[1]: relayer:p:{partition_id}:next_nonce
-- KEYS[2]: relayer:p:{partition_id}:inflight:{nonce}
-- KEYS[3]: relayer:p:{partition_id}:inflight_set
-- ARGV[1]: tx_hash
-- ARGV[2]: trade_id
-- ARGV[3]: gas_price_wei
-- ARGV[4]: submit_timestamp_ms
-- ARGV[5]: partition_id

local current_nonce = redis.call('GET', KEYS[1])
if not current_nonce then
    current_nonce = 0
    redis.call('SET', KEYS[1], 0)
end

local assigned_nonce = tonumber(current_nonce)
redis.call('INCR', KEYS[1])

local inflight_key = 'relayer:p:' .. ARGV[5] .. ':inflight:' .. tostring(assigned_nonce)
redis.call('HSET', inflight_key,
    'tx_hash', ARGV[1],
    'trade_id', ARGV[2],
    'gas_price', ARGV[3],
    'submit_time_ms', ARGV[4],
    'esc_count', 0
)

redis.call('ZADD', KEYS[3], tonumber(ARGV[4]), assigned_nonce)

return assigned_nonce
```

#### Lua Script: `commit_mined_nonce.lua`
```lua
-- KEYS[1]: relayer:p:{partition_id}:confirmed_nonce
-- KEYS[2]: relayer:p:{partition_id}:inflight_set
-- ARGV[1]: mined_nonce
-- ARGV[2]: partition_id

local mined_nonce = tonumber(ARGV[1])
local confirmed_nonce = tonumber(redis.call('GET', KEYS[1]) or 0)

local inflight_key = 'relayer:p:' .. ARGV[2] .. ':inflight:' .. tostring(mined_nonce)
redis.call('DEL', inflight_key)
redis.call('ZREM', KEYS[2], mined_nonce)

if mined_nonce >= confirmed_nonce then
    redis.call('SET', KEYS[1], mined_nonce + 1)
end

return redis.call('GET', KEYS[1])
```

### Kafka Event Schemas

#### Topic: `settlement.tx_requested.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "SettlementTxRequestedEvent",
  "type": "object",
  "required": [
    "request_id",
    "trade_id",
    "settlement_batch_id",
    "isin",
    "contract_address",
    "call_data_hex",
    "gas_limit",
    "initial_gas_price_wei",
    "timestamp_unix_ns"
  ],
  "properties": {
    "request_id": { "type": "string" },
    "trade_id": { "type": "string", "format": "uuid" },
    "settlement_batch_id": { "type": "string", "format": "uuid" },
    "isin": { "type": "string", "pattern": "^[A-Z]{2}[A-Z0-9]{9}[0-9]$" },
    "contract_address": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "call_data_hex": { "type": "string", "pattern": "^0x[0-9a-fA-F]+$" },
    "gas_limit": { "type": "integer" },
    "initial_gas_price_wei": { "type": "string" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `relayer.tx_broadcast.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "RelayerTxBroadcastEvent",
  "type": "object",
  "required": [
    "request_id",
    "trade_id",
    "partition_id",
    "relayer_address",
    "nonce",
    "tx_hash",
    "gas_price_wei",
    "gas_limit",
    "broadcast_at_unix_ns"
  ],
  "properties": {
    "request_id": { "type": "string" },
    "trade_id": { "type": "string", "format": "uuid" },
    "partition_id": { "type": "integer", "minimum": 0, "maximum": 31 },
    "relayer_address": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "nonce": { "type": "integer" },
    "tx_hash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "gas_price_wei": { "type": "string" },
    "gas_limit": { "type": "integer" },
    "broadcast_at_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `relayer.tx_escalated.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "RelayerTxEscalatedEvent",
  "type": "object",
  "required": [
    "trade_id",
    "partition_id",
    "relayer_address",
    "nonce",
    "escalation_round",
    "previous_tx_hash",
    "replacement_tx_hash",
    "previous_gas_price_wei",
    "escalated_gas_price_wei",
    "escalated_at_unix_ns"
  ],
  "properties": {
    "trade_id": { "type": "string", "format": "uuid" },
    "partition_id": { "type": "integer", "minimum": 0, "maximum": 31 },
    "relayer_address": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "nonce": { "type": "integer" },
    "escalation_round": { "type": "integer", "minimum": 1, "maximum": 4 },
    "previous_tx_hash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "replacement_tx_hash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "previous_gas_price_wei": { "type": "string" },
    "escalated_gas_price_wei": { "type": "string" },
    "escalated_at_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `relayer.tx_mined.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "RelayerTxMinedEvent",
  "type": "object",
  "required": [
    "trade_id",
    "partition_id",
    "relayer_address",
    "nonce",
    "final_tx_hash",
    "block_number",
    "block_hash",
    "gas_used",
    "is_reverted",
    "mined_at_unix_ns"
  ],
  "properties": {
    "trade_id": { "type": "string", "format": "uuid" },
    "partition_id": { "type": "integer", "minimum": 0, "maximum": 31 },
    "relayer_address": { "type": "string", "pattern": "^0x[0-9a-fA-F]{40}$" },
    "nonce": { "type": "integer" },
    "final_tx_hash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "block_number": { "type": "integer" },
    "block_hash": { "type": "string", "pattern": "^0x[0-9a-fA-F]{64}$" },
    "gas_used": { "type": "integer" },
    "is_reverted": { "type": "boolean" },
    "revert_reason": { "type": "string" },
    "mined_at_unix_ns": { "type": "integer" }
  }
}
```

## Security & Compliance Notes
- **Key Custody & Hardware Security Isolation:** Private keys for all 32 relayer addresses reside strictly within tamper-resistant AWS CloudHSM or HashiCorp Vault Transit Engine enclaves. No private key bytes or seed phrases are ever stored in environment variables, configuration files, memory buffers, or application logs.
- **Strict Nonce Gap & Replay Prevention:** EVM nodes reject transactions if a nonce gap occurs. The Lua-driven Redis state machine combined with PostgreSQL unique constraints `uq_partition_nonce(partition_id, nonce)` guarantees strictly sequential, non-repeating nonce allocations per partition.
- **Gas Runaway Circuit Breaker:** The gas escalator is strictly capped at a maximum of 4 escalation rounds and a hard ceiling `MaxGasPriceCap`. If a transaction remains unmined after 4 escalations, the service trips a partition circuit breaker, halts subsequent nonce issuance for that partition, and pages the Site Reliability Engineering (SRE) team immediately.
- **Deterministic ISIN Sharding Guarantee:** Murmur3 32-bit hashing ensures that every trade for a particular ISIN is executed sequentially on the same relayer partition, preventing out-of-order race conditions on the underlying security token smart contract.
- **Zero On-Chain PII Compliance (DPDP Act 2023 & GDPR):** All transactions submitted by relayer accounts contain strictly pseudonymous EVM contract addresses, hashes, and numeric quantities. Zero investor identities, Permanent Account Numbers (PANs), or bank account numbers are ever transmitted to the blockchain.

## Acceptance Criteria
- [ ] Service initializes 32 relayer partitions (`relayer_00` to `relayer_31`) mapped to dedicated CloudHSM/Vault key handles.
- [ ] Murmur3 32-bit partitioner deterministically routes identical ISIN strings to the exact same partition index ($0 \dots 31$).
- [ ] Redis sequence queues allocate monotonic, zero-gap nonces per partition under concurrent load of 5,000 requests/sec.
- [ ] CloudHSM / Vault integration signs raw EIP-1559 and legacy transactions via remote PKCS#11/REST without in-memory private key exposure.
- [ ] Besu JSON-RPC client successfully broadcasts raw signed transactions to Hyperledger Besu validator nodes over mTLS.
- [ ] WebSocket block listener detects mined transaction receipts and confirms single-block QBFT deterministic finality.
- [ ] Gas escalator watchdog automatically triggers after $T_{\text{timeout}} = 4\text{ seconds}$ for unmined transactions, computing $\lceil \text{GasPrice} \times 1.25 \rceil$.
- [ ] Escalated replacement transactions maintain the identical nonce, recipient address, and payload data while broadcasting with updated gas price.
- [ ] Nonce resync daemon queries `eth_getTransactionCount` and reconciles Redis state upon detecting desynchronization.
- [ ] Kafka consumer ingests `settlement.tx_requested.v1` and emits corresponding broadcast, escalation, and mined events.
- [ ] PostgreSQL audit tables accurately persist transaction lifecycles, gas escalation history, and partition health snapshots.
- [ ] Service contains zero application implementation code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Design Standards), Prompt `104` (Event Schema & Kafka Topic Standards), Prompt `109` (Secrets Management Architecture), Prompt `208` (Trade Settlement & DvP Orchestration Service), Prompt `401` (PostgreSQL Schema Architecture), Prompt `402` (Redis Patterns & Caching Architecture).
- **Parallel Tasks:** Prompt `215` (On-Chain vs Off-Chain Ledger Reconciliation Engine), Prompt `230` (Settlement Guarantee Fund Service), Prompt `306` (DvP Settlement Smart Contract).
- **Downstream Blockers:** Prompt `306` (DvP Settlement Smart Contract on-chain integration), Prompt `315` (Settlement Guarantee Fund on-chain registry).
