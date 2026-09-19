# 277 - Bitcoin (BTC) On-Chain UTXO Deposit & Confirmation Listener (Go)

## Purpose
Retail and institutional investors in modern 24/7 electronic digital asset exchanges demand real-time, frictionless deposit experiences when funding their accounts with native Bitcoin (BTC). Unlike account-based blockchain networks (such as Ethereum or Hyperledger Besu) or traditional domestic fiat banking ledgers where a single balance balance variable increments, the Bitcoin blockchain operates strictly on an Unspent Transaction Output (UTXO) state model. Each transaction consumes one or more existing outputs and produces discrete new UTXOs bound to cryptographic locking scripts (scriptPubKey).

The **Bitcoin (BTC) On-Chain UTXO Deposit & Confirmation Listener** is an enterprise-grade, high-throughput worker service engineered in Go. Its core operational mission is to continuously monitor the Bitcoin mainnet blockchain for incoming user deposits, detect unconfirmed (0-conf) mempool transactions for instantaneous user interface feedback, systematically track block confirmations up to required probabilistic finality thresholds, autonomously handle chain reorganizations (reorgs) without financial double-crediting, and coordinate atomic real-money account credits.

Upon reaching irreversible consensus finality (a minimum of 3 confirmations for standard deposits and 6 confirmations for high-value institutional amounts), the service triggers downstream balance crediting in the off-chain double-entry ledger (Prompt 203) and coordinates with the platform custody bridge to mint 1:1 backed wrapped Bitcoin (`wBTC`) tokens on the permissioned Hyperledger Besu ledger. This guarantees mathematical parity, real-time Proof-of-Reserve synchronization, zero on-chain Personally Identifiable Information (PII), and strict statutory compliance with anti-money laundering (AML) and Travel Rule mandates.

---

## What You Are Building
A resilient, fault-tolerant Go worker daemon (`services/btc-deposit-listener`) built with Go 1.22+ that interfaces directly with Bitcoin Core full nodes and fallback block explorer APIs. Concrete deliverables include:

- **Dual-Channel Bitcoin Ingestion Engine (ZMQ + JSON-RPC):** A sub-second block and transaction listener utilizing Bitcoin Core ZeroMQ streams (`zmqpubrawblock`, `zmqpubrawtx`) for low-latency mempool and block notifications, coupled with an active JSON-RPC connection pool (`btcsuite/btcd/rpcclient`) for deep block retrieval, raw transaction deserialization, and script inspection.
- **Monitored Address Registry & Bloom/Hash Filter Cache:** A high-speed, thread-safe in-memory cache and Redis 7.2 index tracking active user deposit addresses across modern Bitcoin address formats: Native SegWit (BIP-84 P2WPKH, `bc1q...`) and Taproot (BIP-86 P2TR, `bc1p...`), synchronized continuously from the Multi-Chain MPC-TSS Vault Service (Prompt 237).
- **Mempool 0-Conf Deposit Detector:** A real-time transaction analyzer that parses unconfirmed mempool transactions, identifies deposits to monitored platform addresses, verifies script integrity, evaluates BIP-125 Replace-By-Fee (RBF) signaling, and emits immediate 0-conf notifications to Flutter and Web clients via Notification Service (Prompt 211).
- **Block Ingestion & UTXO Confirmation Tracker:** A persistent worker that processes newly mined blocks, extracts confirmed UTXOs matching monitored addresses, records block heights and hashes, and increments confirmation counts deterministically on every new block tip extension.
- **Deep Chain Reorganization (Reorg) Detector & Rollback Engine:** A stateful block header tracker that compares incoming block parent hashes against the local database canonical chain. If a chain reorganization is detected (up to 100 blocks deep), the engine identifies the common ancestor, marks orphaned block deposits as `REORG_ORPHANED`, halts uncommitted credits, and triggers automated ledger holds or compensation workflows in Wallet Account Service (Prompt 203).
- **RBF & Double-Spend Surveillance Daemon:** A security monitor that tracks mempool replacement transactions and detects conflicting inputs (double-spend attempts) targeting platform deposit addresses, immediately blacklisting tainted transactions and alerting market surveillance.
- **1:1 Wrapped BTC (`wBTC`) Minting Coordinator:** An orchestration pipeline that, upon verified deposit confirmation (3 or 6 confs), dispatches cryptographically verified deposit attestations to Kafka, triggering Trade Settlement Service (Prompt 208) to mint 1:1 backed `wBTC` on Hyperledger Besu while crediting the user's trading balance in Wallet Account Service (Prompt 203).
- **Internal gRPC API (`BtcDepositService`):** A high-performance Protobuf gRPC interface exposing RPCs to register new deposit addresses, query deposit transaction lifecycles, stream deposit updates, and manually trigger block range re-scans.

---

## Scope Boundaries

### In Scope
- Continuous monitoring of Bitcoin mainnet (and testnet/regtest in sandbox environments) blocks and mempool.
- Ingestion, parsing, and tracking of Native SegWit (BIP-84 P2WPKH, `bc1q...`) and Taproot (BIP-86 P2TR, `bc1p...`) deposit addresses.
- Sub-second detection of unconfirmed mempool transactions (0-conf) for UI push notifications.
- Stateful lifecycle tracking of UTXO deposits: `DETECTED_0CONF`, `CONFIRMING`, `CONFIRMED_CREDITED`, `REORG_ORPHANED`, `DOUBLE_SPENT`, and `AML_HOLD`.
- Multi-tier confirmation thresholds: minimum 3 confirmations for standard deposits (<= 0.5 BTC), minimum 6 confirmations for high-value institutional deposits (> 0.5 BTC or AML flagged).
- Chain reorganization detection up to 100 blocks deep, canonical chain re-alignment, and orphaned transaction state rollback.
- Detection of BIP-125 Replace-By-Fee (RBF) transactions and mempool input double-spend attempts.
- Publishing structured lifecycle events to Apache Kafka (`crypto.btc.deposit.detected.v1`, `crypto.btc.deposit.confirmed.v1`, `crypto.btc.reorg.detected.v1`, `crypto.btc.doublespend.detected.v1`).
- Triggering double-entry ledger balance credits in Wallet Account Service (Prompt 203).
- Coordinating 1:1 wrapped Bitcoin (`wBTC`) token minting on Hyperledger Besu under QBFT consensus.
- Statutory Travel Rule threshold gating (> INR 50,000 / $1,000 USD equivalent) and AML screening status evaluation.
- PostgreSQL 16 schema design and transactional persistence using `sqlc` with transactional outbox patterns.
- High-availability catch-up synchronization recovering missed blocks after node downtime or network partitions.

### Out of Scope / Handled Elsewhere
- Private key management, MPC threshold signing, and Bitcoin withdrawal transaction broadcasting (handled in Prompt 237 - Institutional Multi-Chain MPC-TSS Vault & Custody Service).
- Automated UTXO consolidation, fee bumping (CPFP/RBF), and sweeping from deposit addresses into platform cold storage (handled in Prompt 237).
- Spot trading, order book matching, and instant RFQ conversions between BTC, INR, and USDT (handled in Prompt 205 and Prompt 270).
- Primary user identity onboarding, KYC verification, and Aadhaar/PAN validation (handled in Prompt 202).
- Smart contract implementation of the `wBTC` ERC-20 contract and Proof-of-Reserve Registry on Besu (handled in Prompt 306 and Prompt 327).
- Direct external integration with third-party blockchain analytics vendors (Chainalysis, Elliptic, TRM Labs) for transaction risk scoring (handled in Compliance & AML Gateway Prompt 216).
- End-user mobile and web deposit interfaces (handled in Prompt 528 and Prompt 610).

---

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+ for concurrent goroutines, low-latency network I/O, garbage collection determinism, and enterprise cryptographic library support.
- **Bitcoin Blockchain SDK:** `github.com/btcsuite/btcd` (sub-packages: `rpcclient`, `wire`, `btcjson`, `chaincfg`, `txscript`) and `github.com/btcsuite/btcd/btcutil` (`bech32`, `address`) for Bitcoin protocol serialization, address validation, script parsing, and JSON-RPC client management.
- **Bitcoin Node Infrastructure:** Bitcoin Core (`bitcoind` v26.0+) configured with `-txindex=1`, `-server=1`, and active ZeroMQ notifications (`-zmqpubrawblock`, `-zmqpubrawtx`).
- **ZeroMQ Messaging Library:** `github.com/pebbe/zmq4` (or pure Go ZeroMQ implementation `go-zeromq/zmq4`) for consuming low-latency raw block and raw transaction binary streams from `bitcoind`.
- **Relational Persistence & Auditing:** PostgreSQL 16+ utilizing `pgx/v5` connection pool with `sqlc` for compile-time verified, zero-allocation database queries and strict ACID transaction guarantees.
- **Distributed Caching & In-Memory Index:** Redis 7.2 Cluster for sub-millisecond deposit address caching, bloom filter membership testing, and distributed mutex locks (`redsync/v2`) during block processing.
- **Message Broker & Event Bus:** Apache Kafka 3.7+ (`segmentio/kafka-go`) with SASL/SCRAM authentication for high-durability event streaming.
- **Inter-Service Communication:** gRPC over HTTP/2 with Protocol Buffers (`google.golang.org/grpc`, `google.golang.org/protobuf`) for internal service APIs.
- **High-Precision Fixed-Point Math:** `github.com/shopspring/decimal` for exact satoshi-to-BTC and BTC-to-INR conversions without floating-point inaccuracies.
- **Observability & Metrics:** OpenTelemetry (`go.opentelemetry.io/otel`) for distributed trace propagation, Prometheus metrics client for tracking block lag, confirmation latency, and reorg depths, and `uber-go/zap` for structured JSON logging.

---

## Backend / Infra Touchpoints
- **Bitcoin Core Full Node Cluster:** Primary source of truth for raw blocks, mempool transactions, and UTXO confirmation states. Connects via ZMQ TCP port 28332/28333 and authenticated JSON-RPC port 8332.
- **Multi-Chain MPC-TSS Vault Service (Prompt 237):** Supplies newly derived P2WPKH (`bc1q`) and P2TR (`bc1p`) deposit addresses for registered users. Emits `mpc.vault.address_generated.v1` consumed by this service.
- **Wallet & Account Service (Prompt 203):** Receives balance credit commands upon confirmed UTXO deposits. Executes atomic double-entry journal postings crediting user crypto balances and debiting exchange vault clearing accounts.
- **Notification Service (Prompt 211):** Consumes deposit events to send instantaneous push notifications, SMS, and emails to users upon 0-conf detection and 3-conf final settlement.
- **Risk Engine & AML Service (Prompt 206 / Prompt 216):** Performs real-time sanctions screening and Travel Rule compliance verification on deposit transactions prior to releasing credited funds for trading or withdrawal.
- **Trade Settlement Service (Prompt 208):** Orchestrates the issuance of 1:1 backed `wBTC` on Hyperledger Besu once Bitcoin mainnet confirmations are locked.
- **Proof-of-Reserve Registry (Prompt 327):** Consumes confirmed UTXO updates to update the exchange's cryptographic Proof-of-Reserve balance attestations on-chain.
- **Audit Log & Surveillance Service (Prompt 218):** Ingests immutable audit records of all detected deposits, reorgs, and credit executions for regulatory inspection by FIU-IND and SEBI.
- **Apache Kafka Topics:**
  - *Consumes:* `mpc.vault.address_generated.v1`, `compliance.aml.cleared.v1`.
  - *Publishes:* `crypto.btc.deposit.detected.v1`, `crypto.btc.deposit.confirming.v1`, `crypto.btc.deposit.confirmed.v1`, `crypto.btc.reorg.detected.v1`, `crypto.btc.doublespend.detected.v1`.

---

## Blockchain Interaction (Reading Bitcoin mainnet UTXOs, minting wrapped 1:1 backed wBTC on Besu Mainnet upon confirmation)

### Reading Bitcoin Mainnet UTXOs
The service interacts with Bitcoin mainnet through a resilient dual-pipe architecture:
1. **Low-Latency ZeroMQ Mempool & Block Stream:**
   - Subscribes to `bitcoind` ZMQ topics `zmqpubrawtx` and `zmqpubrawblock`.
   - On incoming raw transactions, decodes the Bitcoin wire format using `wire.MsgTx`.
   - Inspects all outputs (`TxOut`). Validates each output's `PkScript` against the active monitored address registry.
   - Decodes Native SegWit P2WPKH (witness version 0, 20-byte pubkey hash, Bech32 `bc1q` prefix) and Taproot P2TR (witness version 1, 32-byte Schnorr output key, Bech32m `bc1p` prefix).
   - Identifies matching user deposit addresses and records 0-conf mempool detections within < 500ms of network broadcast.
2. **Canonical Block Height & Header Verification:**
   - When `zmqpubrawblock` emits a new block (or during periodic block polling), decodes `wire.MsgBlock`.
   - Extracts the block header, block hash, previous block hash, Merkle root, timestamp, and transaction count.
   - Verifies canonical chain continuity by checking that `prev_block_hash` matches the currently recorded chain tip in PostgreSQL.
   - Extracts confirmed transactions matching monitored addresses, records the confirmed UTXO details (txid, vout index, satoshi value, scriptPubKey, block height, block hash), and assigns 1 confirmation.
3. **Confirmation Counting Engine:**
   - As subsequent blocks arrive, updates confirmation counts: $\text{Confirmations} = (\text{CurrentTipHeight} - \text{DepositBlockHeight}) + 1$.
   - Evaluates multi-tier thresholds:
     - Standard deposits ($\le 0.5\text{ BTC}$): crediting triggered at $\ge 3$ confirmations ($\approx 30$ minutes).
     - High-value / institutional deposits ($> 0.5\text{ BTC}$ or high-risk AML score): crediting triggered at $\ge 6$ confirmations ($\approx 60$ minutes).

### Minting Wrapped 1:1 Backed `wBTC` on Besu Mainnet
Once a Bitcoin deposit achieves the required confirmation depth:
1. **Deterministic Satoshi Parity:**
   - Bitcoin satoshis are mapped with exact 1:1 mathematical precision. 1 BTC = $100,000,000$ satoshis.
   - The platform `wBTC` smart contract deployed on Hyperledger Besu is configured with 8 decimal places (matching native satoshis directly) or 18 decimal places with standardized $10^{10}$ scaling factor.
2. **Atomic Settlement Dispatch:**
   - The service writes a confirmed deposit record to PostgreSQL within a database transaction and enqueues a settlement task in the outbox table.
   - The settlement task dispatches a command to Kafka topic `settlement.besu.mint_requested.v1`.
   - Trade Settlement Service (Prompt 208), utilizing the authorized platform relayer key or MPC vault coordinator, invokes `wBTC.mint(userBesuAddress, amountSats, depositIdHash)` on Hyperledger Besu under QBFT consensus (2-second block finality).
3. **Zero On-Chain PII Invariant:**
   - The Besu mint transaction records only the user's pseudonymous on-chain address (`0x...`), the minted token amount, and a cryptographic SHA-256 hash of the Bitcoin `txid:vout`.
   - Zero user PII (names, PAN, Aadhaar, email, IP addresses) is ever transmitted to or stored on Hyperledger Besu.
4. **Proof-of-Reserve Attestation:**
   - The newly locked Bitcoin mainnet UTXO is registered in the Proof-of-Reserve Registry (Prompt 327), maintaining a provable 1:1 balance backing between total Bitcoin in custody vaults and total circulating `wBTC` on Besu.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Service Workspace:** Initialize the Go module at `services/btc-deposit-listener` with standardized project structure:
   - `cmd/listener/`: Main daemon entrypoint, configuration parsing, dependency injection, graceful shutdown handler.
   - `internal/config/`: Environment variable and file-based configuration management.
   - `internal/bitcoind/`: Bitcoin Core RPC client, connection pooling, ZMQ listener, block deserializer.
   - `internal/registry/`: Monitored address registry, in-memory cache, Redis bloom filter synchronization.
   - `internal/mempool/`: 0-conf transaction scanner, RBF detector, double-spend analyzer.
   - `internal/indexer/`: Block processor, canonical chain tracker, confirmation updater, reorg detector.
   - `internal/settlement/`: Settlement bridge dispatcher, Besu `wBTC` mint coordinator, Wallet Account Service client.
   - `internal/store/`: PostgreSQL schemas, migrations, and `sqlc`-generated queries.
   - `internal/kafka/`: Kafka consumer and transactional outbox event publisher.
   - `internal/grpc/`: gRPC server handlers implementing `BtcDepositService`.
2. **Define Protobuf Specifications:** Author `proto/growww/btc/v1/btc_deposit_service.proto` defining RPCs for registering addresses, querying deposit statuses, listing UTXOs, streaming updates, and manually requesting block re-indexing. Generate Go stubs using `buf`.
3. **Author PostgreSQL Schema & Configure `sqlc`:** Write database migrations defining tables: `btc_monitored_addresses`, `btc_block_headers`, `btc_utxo_deposits`, `btc_reorg_events`, and `btc_deposit_outbox`. Configure `sqlc.yaml` and generate type-safe Go query methods.
4. **Implement Bitcoin Core RPC & ZMQ Client:** Build resilient Bitcoin Core integration layer wrapping `btcsuite/btcd/rpcclient`. Implement ZMQ subscriber subscribing to `rawtx` and `rawblock` with automatic reconnection, heartbeat monitoring, and fallback to JSON-RPC block polling if ZMQ stalls.
5. **Build Monitored Address Registry & Cache:** Create a concurrent, lock-free in-memory hash map (`sync.Map` or read-heavy RWMutex) storing all active user deposit addresses. Implement background worker polling PostgreSQL and consuming `mpc.vault.address_generated.v1` from Kafka to load newly derived P2WPKH (`bc1q`) and P2TR (`bc1p`) addresses dynamically without restarting the daemon.
6. **Implement Bech32 / Bech32m Address Parser:** Author robust address decoding logic using `btcutil/bech32` and `txscript`. Extract public key hashes from witness v0 scripts and output keys from witness v1 Taproot scripts. Ensure strict rejection of non-standard or malformed scriptPubKeys.
7. **Build 0-Conf Mempool Deposit Scanner:** Connect the ZMQ `rawtx` stream to a worker pool. For every incoming transaction, inspect outputs. If an output scriptPubKey matches a monitored address:
   - Verify transaction structure, ensure value is above the dust threshold (546 satoshis).
   - Check BIP-125 RBF signaling (sequence number < `0xffffffff - 1`).
   - Insert deposit record into PostgreSQL with status `DETECTED_0CONF`.
   - Publish `crypto.btc.deposit.detected.v1` to Kafka and push instant notification to Notification Service (Prompt 211).
8. **Build Block Ingestion & Canonical Header Tracker:** On arrival of each new block (via ZMQ `rawblock` or RPC polling):
   - Parse `wire.MsgBlock` header: block hash, height, previous block hash, timestamp.
   - Fetch the latest known block header from `btc_block_headers`.
   - Verify chain continuity: check if `block.Header.PrevBlock == latest_db_block.Hash`.
   - If continuous, insert new block header into `btc_block_headers` within an atomic transaction.
9. **Implement Deep Chain Reorganization (Reorg) Detector:**
   - If `block.Header.PrevBlock != latest_db_block.Hash`, initiate reorg resolution:
     - Trace backwards along the Bitcoin Core header chain using RPC `getblockheader` until finding the common ancestor block with the local database.
     - Calculate reorg depth: $D = \text{LatestLocalHeight} - \text{CommonAncestorHeight}$.
     - Log reorg event in `btc_reorg_events`. If $D > 100$, raise a critical P1 alert and halt automated processing for manual intervention.
     - For all blocks between common ancestor and old tip, mark status as `ORPHANED`.
     - Find all deposits recorded in orphaned blocks. Transition their status to `REORG_ORPHANED`.
     - If any orphaned deposit was already credited in Wallet Account Service (Prompt 203), dispatch an immediate compensation / freeze event to Kafka (`crypto.btc.reorg.detected.v1`).
     - Rewind database block tip to common ancestor, then replay the new canonical chain blocks forward.
10. **Build UTXO Confirmation Counter & Dynamic Crediting Pipeline:**
    - On every new canonical block tip update, execute a batch SQL query to recalculate confirmations for all uncredited deposits (`CONFIRMING`).
    - Increment confirmation counts.
    - Evaluate crediting policy:
      - If amount $\le 0.5\text{ BTC}$ and confirmations $\ge 3$: mark `CONFIRMED_CREDITED`.
      - If amount $> 0.5\text{ BTC}$ and confirmations $\ge 6$: mark `CONFIRMED_CREDITED`.
      - If AML status is `FLAGGED`, transition to `AML_HOLD` and pause crediting pending compliance review.
11. **Implement BIP-125 RBF & Double-Spend Surveillance Monitor:**
    - Maintain an in-memory index of spent outpoints (`txid:vout`) for all unconfirmed mempool transactions.
    - If a new mempool transaction spends an outpoint already present in another unconfirmed deposit, detect conflict.
    - If BIP-125 replacement: verify if the new transaction still pays the platform address with an equal or greater amount. If the platform output was stripped or replaced, mark deposit as `DOUBLE_SPENT` and emit `crypto.btc.doublespend.detected.v1`.
12. **Implement Real-Money Settlement & Besu `wBTC` Mint Trigger:**
    - When a deposit transitions to `CONFIRMED_CREDITED`:
      - Write a settlement task into `btc_deposit_outbox` within the same PostgreSQL ACID transaction.
      - Asynchronously dispatch gRPC call to Wallet Account Service (Prompt 203) to post double-entry ledger credit to user account.
      - Publish event `crypto.btc.deposit.confirmed.v1` to Kafka, triggering Trade Settlement Service (Prompt 208) to submit `wBTC.mint(...)` on Hyperledger Besu.
13. **Build Catch-Up Block Sync & Background Reconciliation Daemon:**
    - On service startup (or recovery from network disconnect), query Bitcoin Core for current best block height.
    - Compare against local database maximum height.
    - Sequentially fetch and process all missed blocks in batches (e.g., 50 blocks per batch) before enabling live ZMQ processing, ensuring zero gap in UTXO detection.
    - Run an hourly background reconciliation cron comparing all database UTXOs against Bitcoin Core's UTXO set (`gettxout` or `scantxoutset`) to verify state correctness.
14. **Implement gRPC Service & Health Endpoints:** Expose gRPC server implementing `BtcDepositService` with OpenTelemetry tracing interceptors, Prometheus metrics middleware (`/metrics`), and Kubernetes liveness/readiness probes (`/healthz`, `/readyz`).
15. **Execute End-to-End Test Suite:** Author unit and integration tests using `dockertest` or `testcontainers-go` running `bitcoind` in regtest mode:
    - Test Native SegWit (`bc1q`) and Taproot (`bc1p`) deposit detection.
    - Test 0-conf mempool event generation.
    - Mine blocks and verify confirmation progression (0 -> 1 -> 2 -> 3 -> 6).
    - Simulate a 4-block chain reorganization: verify orphan detection, confirmation reset, and rollback safety.
    - Simulate RBF double-spend and verify rejection.
    - Verify atomic outbox publishing to Kafka.

---

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/btc/v1/btc_deposit_service.proto`)

```protobuf
syntax = "proto3";

package growww.btc.v1;

option go_package = "growww/btc/v1;btcv1";

// BtcDepositService manages Bitcoin on-chain UTXO deposit monitoring,
// address registration, confirmation tracking, and settlement handoffs.
service BtcDepositService {
  // Register a newly derived user deposit address for real-time monitoring.
  rpc RegisterDepositAddress (RegisterDepositAddressRequest) returns (RegisterDepositAddressResponse);

  // Get current active deposit address details for a user.
  rpc GetDepositAddress (GetDepositAddressRequest) returns (GetDepositAddressResponse);

  // Retrieve a specific UTXO deposit by transaction hash and output index.
  rpc GetDepositByTxId (GetDepositByTxIdRequest) returns (GetDepositByTxIdResponse);

  // List all UTXO deposits associated with a specific user.
  rpc ListDepositsByUser (ListDepositsByUserRequest) returns (ListDepositsByUserResponse);

  // Server-streaming RPC emitting real-time deposit lifecycle updates.
  rpc StreamDepositUpdates (StreamDepositUpdatesRequest) returns (stream DepositUpdateEvent);

  // Administrative RPC to trigger manual block re-indexing across a height range.
  rpc ReindexBlockRange (ReindexBlockRangeRequest) returns (ReindexBlockRangeResponse);

  // Query service health, Bitcoin Core sync state, and chain tip height.
  rpc GetSyncStatus (GetSyncStatusRequest) returns (GetSyncStatusResponse);
}

enum AddressScriptType {
  ADDRESS_SCRIPT_TYPE_UNSPECIFIED = 0;
  ADDRESS_SCRIPT_TYPE_P2WPKH = 1;     // Native SegWit (BIP-84, bc1q...)
  ADDRESS_SCRIPT_TYPE_P2TR = 2;       // Taproot (BIP-86, bc1p...)
}

enum DepositStatus {
  DEPOSIT_STATUS_UNSPECIFIED = 0;
  DEPOSIT_STATUS_DETECTED_0CONF = 1;  // Seen in mempool, 0 confirmations
  DEPOSIT_STATUS_CONFIRMING = 2;      // Included in block, 1 to N-1 confirmations
  DEPOSIT_STATUS_CONFIRMED_CREDITED = 3; // Met confirmation threshold and credited
  DEPOSIT_STATUS_REORG_ORPHANED = 4;  // Block orphaned by reorg, confirmations lost
  DEPOSIT_STATUS_DOUBLE_SPENT = 5;    // Conflicting input detected in mempool/block
  DEPOSIT_STATUS_AML_HOLD = 6;        // Confirmation halted pending Travel Rule/AML
  DEPOSIT_STATUS_MANUAL_REVIEW = 7;   // Flagged for compliance or engineering audit
}

message RegisterDepositAddressRequest {
  string user_id = 1;
  string bitcoin_address = 2;
  AddressScriptType script_type = 3;
  string derivation_path = 4;
  string vault_account_id = 5;
  int64 registered_at_unix_ms = 6;
}

message RegisterDepositAddressResponse {
  string address_id = 1;
  string bitcoin_address = 2;
  bool is_active = 3;
  int64 registered_at_unix_ms = 4;
}

message GetDepositAddressRequest {
  string user_id = 1;
  AddressScriptType preferred_script_type = 2;
}

message GetDepositAddressResponse {
  string address_id = 1;
  string user_id = 2;
  string bitcoin_address = 3;
  AddressScriptType script_type = 4;
  string derivation_path = 5;
  bool is_active = 6;
}

message UtxoDepositRecord {
  string deposit_id = 1;
  string user_id = 2;
  string bitcoin_address = 3;
  string txid = 4;
  uint32 vout = 5;
  int64 amount_satoshis = 6;
  string amount_btc = 7;             // Fixed-point string representation, e.g. "0.05000000"
  string script_pubkey_hex = 8;
  int64 block_height = 9;            // 0 if unconfirmed in mempool
  string block_hash = 10;            // Empty if unconfirmed
  int32 confirmations = 11;
  int32 required_confirmations = 12; // 3 for standard, 6 for large
  DepositStatus status = 13;
  bool rbf_signaled = 14;            // BIP-125 Replace-By-Fee flag
  string credit_tx_id = 15;          // Wallet Service journal entry ID
  string besu_wbtc_mint_tx_hash = 16;// Hyperledger Besu mint transaction hash
  int64 detected_at_unix_ms = 17;
  int64 confirmed_at_unix_ms = 18;
  int64 credited_at_unix_ms = 19;
}

message GetDepositByTxIdRequest {
  string txid = 1;
  uint32 vout = 2;
}

message GetDepositByTxIdResponse {
  UtxoDepositRecord deposit = 1;
}

message ListDepositsByUserRequest {
  string user_id = 1;
  DepositStatus status_filter = 2;
  int32 page_size = 3;
  string page_token = 4;
}

message ListDepositsByUserResponse {
  repeated UtxoDepositRecord deposits = 1;
  string next_page_token = 2;
  int64 total_count = 3;
}

message StreamDepositUpdatesRequest {
  string user_id = 1;                // Optional: filter by user, empty for all
}

message DepositUpdateEvent {
  string event_id = 1;
  string event_type = 2;             // e.g. "DEPOSIT_DETECTED", "CONFIRMATION_INCREMENT", "DEPOSIT_CREDITED"
  UtxoDepositRecord deposit = 3;
  int64 timestamp_unix_ms = 4;
}

message ReindexBlockRangeRequest {
  int64 start_height = 1;
  int64 end_height = 2;
  bool force_reprocess = 3;
}

message ReindexBlockRangeResponse {
  int64 blocks_queued = 1;
  string job_id = 2;
}

message GetSyncStatusRequest {}

message GetSyncStatusResponse {
  int64 local_best_height = 1;
  string local_best_hash = 2;
  int64 bitcoind_best_height = 3;
  string bitcoind_best_hash = 4;
  bool is_synced = 5;
  int64 blocks_behind = 6;
  int64 active_monitored_addresses = 7;
  int64 mempool_unconfirmed_deposits = 8;
  int64 server_time_unix_ms = 9;
}
```

---

### PostgreSQL Database Schema DDL

```sql
-- PostgreSQL 16+ DDL Schema for Bitcoin On-Chain UTXO Deposit & Confirmation Listener

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Enumerated Types
CREATE TYPE btc_address_script_type_enum AS ENUM (
    'P2WPKH',
    'P2TR'
);

CREATE TYPE btc_deposit_status_enum AS ENUM (
    'DETECTED_0CONF',
    'CONFIRMING',
    'CONFIRMED_CREDITED',
    'REORG_ORPHANED',
    'DOUBLE_SPENT',
    'AML_HOLD',
    'MANUAL_REVIEW'
);

CREATE TYPE btc_block_status_enum AS ENUM (
    'CANONICAL',
    'ORPHANED'
);

-- Table 1: Monitored User Deposit Addresses
CREATE TABLE btc_monitored_addresses (
    address_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    bitcoin_address VARCHAR(90) NOT NULL UNIQUE,
    script_type btc_address_script_type_enum NOT NULL,
    derivation_path VARCHAR(64) NOT NULL,
    vault_account_id VARCHAR(64) NOT NULL,
    script_pubkey_hex VARCHAR(130) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 2: Canonical and Reorganized Bitcoin Block Headers
CREATE TABLE btc_block_headers (
    block_hash VARCHAR(64) PRIMARY KEY,
    block_height BIGINT NOT NULL,
    prev_block_hash VARCHAR(64) NOT NULL,
    merkle_root VARCHAR(64) NOT NULL,
    block_timestamp TIMESTAMPTZ NOT NULL,
    tx_count INT NOT NULL,
    status btc_block_status_enum NOT NULL DEFAULT 'CANONICAL',
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 3: Persistent UTXO Deposit Records
CREATE TABLE btc_utxo_deposits (
    deposit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    address_id UUID NOT NULL REFERENCES btc_monitored_addresses(address_id) ON DELETE RESTRICT,
    bitcoin_address VARCHAR(90) NOT NULL,
    txid VARCHAR(64) NOT NULL,
    vout INT NOT NULL,
    amount_satoshis BIGINT NOT NULL CHECK (amount_satoshis > 546), -- Must exceed Bitcoin dust threshold
    amount_btc NUMERIC(16, 8) NOT NULL CHECK (amount_btc > 0),
    script_pubkey_hex VARCHAR(130) NOT NULL,
    block_hash VARCHAR(64) REFERENCES btc_block_headers(block_hash),
    block_height BIGINT,
    confirmations INT NOT NULL DEFAULT 0 CHECK (confirmations >= 0),
    required_confirmations INT NOT NULL DEFAULT 3,
    status btc_deposit_status_enum NOT NULL DEFAULT 'DETECTED_0CONF',
    rbf_signaled BOOLEAN NOT NULL DEFAULT FALSE,
    aml_screening_passed BOOLEAN NOT NULL DEFAULT FALSE,
    credit_tx_id UUID,                     -- Foreign reference to Wallet Service journal
    besu_wbtc_mint_tx_hash VARCHAR(66),    -- 0x-prefixed 32-byte hash on Besu
    raw_tx_hex TEXT NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    credited_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_btc_txid_vout UNIQUE (txid, vout)
);

-- Table 4: Chain Reorganization Audit Log
CREATE TABLE btc_reorg_events (
    reorg_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    detected_height BIGINT NOT NULL,
    common_ancestor_height BIGINT NOT NULL,
    common_ancestor_hash VARCHAR(64) NOT NULL,
    orphaned_tip_hash VARCHAR(64) NOT NULL,
    new_tip_hash VARCHAR(64) NOT NULL,
    reorg_depth INT NOT NULL,
    affected_deposit_count INT NOT NULL DEFAULT 0,
    affected_deposits JSONB NOT NULL DEFAULT '[]'::jsonb,
    status VARCHAR(32) NOT NULL DEFAULT 'RESOLVED',
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 5: Transactional Outbox for Kafka Event Publishing
CREATE TABLE btc_deposit_outbox (
    outbox_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(64) NOT NULL,
    aggregate_id VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    dispatched_at TIMESTAMPTZ
);

-- Performance and Query Optimization Indexes
CREATE INDEX idx_btc_monitored_user ON btc_monitored_addresses (user_id);
CREATE INDEX idx_btc_monitored_address ON btc_monitored_addresses (bitcoin_address) WHERE is_active = TRUE;
CREATE INDEX idx_btc_block_height ON btc_block_headers (block_height, status);
CREATE INDEX idx_btc_block_prev ON btc_block_headers (prev_block_hash);
CREATE INDEX idx_btc_utxo_user ON btc_utxo_deposits (user_id, created_at DESC);
CREATE INDEX idx_btc_utxo_status ON btc_utxo_deposits (status) WHERE status IN ('DETECTED_0CONF', 'CONFIRMING');
CREATE INDEX idx_btc_utxo_address ON btc_utxo_deposits (bitcoin_address);
CREATE INDEX idx_btc_utxo_block ON btc_utxo_deposits (block_height) WHERE block_height IS NOT NULL;
CREATE INDEX idx_btc_outbox_pending ON btc_deposit_outbox (created_at ASC) WHERE status = 'PENDING';
```

---

## Security & Compliance Notes

### Reorg Depth Threshold & Confirmation Policy
- **Probabilistic Finality Safeguards:** In proof-of-work systems like Bitcoin, transactions are never mathematically final upon block inclusion due to the possibility of chain forks. To mitigate double-spend risk:
  - Standard deposits ($\le 0.5\text{ BTC}$, equivalent to $\approx \text{₹25,00,000}$): require a minimum of **3 confirmations** before balance crediting.
  - High-value / institutional deposits ($> 0.5\text{ BTC}$): require a minimum of **6 confirmations** before balance crediting.
  - Deposits originating from newly flagged addresses or users undergoing enhanced AML due diligence: locked until **12 confirmations** and manual compliance clearance.
- **Automated Reorg Detection & Compensation:**
  - If a chain reorg occurs, the service automatically halts in-flight confirmation updates and determines the common ancestor block.
  - If an uncredited deposit is orphaned, its confirmation counter resets to 0 and its status reverts to `REORG_ORPHANED`.
  - In the rare event that a reorg exceeds 3 blocks and reverses a previously credited deposit, the service dispatches an emergency compensation event to Wallet Account Service (Prompt 203) to place an immediate hold on the user's balance and alert the risk engineering desk.

### Replace-By-Fee (RBF) & Double-Spend Defense
- **BIP-125 Enforcement:** Transactions signaling opt-in RBF (any input sequence number $< \text{0xffffffff - 1}$) can be replaced in the mempool by a higher-fee transaction. The service marks all 0-conf RBF transactions with `rbf_signaled = TRUE`.
- **Zero-Credit on 0-Conf:** Under no operational circumstance does the platform credit usable trading balances or mint `wBTC` on 0-conf mempool detections. 0-conf detection is strictly utilized for UI display and real-time user notification.
- **Mempool Conflict Detection:** The service indexes all unconfirmed inputs. If a transaction attempts to spend an outpoint already registered to a pending deposit with different outputs, the deposit is immediately marked as `DOUBLE_SPENT`, frozen, and logged for market abuse investigation.

### Travel Rule & Anti-Money Laundering (AML) Compliance
- **FATF Recommendation 16 & PMLA Compliance:** Under statutory guidelines established by the Financial Intelligence Unit - India (FIU-IND) and SEBI:
  - For any deposit exceeding the statutory reporting threshold of **₹50,000 INR** ($\approx \$600\text{ USD}$), the service evaluates Travel Rule attestation status.
  - Incoming deposits trigger an automated pre-credit query to Compliance & AML Gateway (Prompt 216) to perform wallet clustering analysis, verify the sending Virtual Asset Service Provider (VASP), and screen the UTXO history against sanctions lists (OFAC, UN, MHA).
  - If the originator score indicates high risk (e.g., darknet marketplace, ransomware, sanctions mixer), the deposit is transitioned to `AML_HOLD` and funds remain locked in custody until compliance manual review is completed.

### Zero Private Keys & Read-Only Infrastructure
- **Security Invariance:** The `btc-deposit-listener` service contains **zero private keys, zero seed phrases, and zero signing credentials**. It operates strictly as a read-only observer of the public Bitcoin blockchain and a consumer of pre-derived public addresses.
- **Node Redundancy & Air-Gapping:** Bitcoin Core nodes run in an isolated VPC network with RPC credentials managed via HashiCorp Vault. RPC endpoints are restricted via mutual TLS and IP firewalls.

---

## Acceptance Criteria

- [ ] Go worker service (`services/btc-deposit-listener`) compiles cleanly with Go 1.22+ and zero compiler warnings.
- [ ] Connects successfully to Bitcoin Core JSON-RPC and ZeroMQ (`rawtx`, `rawblock`) endpoints with automated reconnection on connection failure.
- [ ] Accurately decodes and validates Native SegWit (BIP-84 P2WPKH, `bc1q...`) and Taproot (BIP-86 P2TR, `bc1p...`) deposit addresses.
- [ ] Detects incoming 0-conf transactions in the Bitcoin mempool within $< 1000\text{ms}$ of network broadcast and publishes `crypto.btc.deposit.detected.v1` to Kafka.
- [ ] Correctly identifies BIP-125 Replace-By-Fee (RBF) signaling on unconfirmed mempool transactions and marks `rbf_signaled = true`.
- [ ] Progressively tracks confirmations as new canonical blocks arrive (0 -> 1 -> 2 -> 3 -> 6).
- [ ] Enforces dynamic confirmation thresholds: credits standard deposits ($\le 0.5\text{ BTC}$) at exactly 3 confirmations and high-value deposits ($> 0.5\text{ BTC}$) at exactly 6 confirmations.
- [ ] Detects chain reorganizations up to 100 blocks deep, identifies common ancestor headers, and correctly transitions orphaned deposits to `REORG_ORPHANED`.
- [ ] Executes transactional outbox pattern in PostgreSQL 16 to guarantee exactly-once event publication to Apache Kafka without event loss.
- [ ] Dispatches double-entry balance credit commands to Wallet Account Service (Prompt 203) upon reaching required confirmation thresholds.
- [ ] Emits confirmed deposit events that trigger Trade Settlement Service (Prompt 208) to mint 1:1 backed `wBTC` on Hyperledger Besu with zero on-chain PII.
- [ ] Evaluates Travel Rule and AML screening statuses, locking high-risk deposits in `AML_HOLD` prior to financial crediting.
- [ ] Successfully performs catch-up synchronization on daemon startup, backfilling all missed blocks without dropping UTXOs or duplicating deposits.
- [ ] Passes all unit and integration tests under `dockertest` / `testcontainers-go` against a live `bitcoind` regtest node.

---

## Suggested Order / Dependencies

- **Prerequisites:**
  - `103_api_design_standards.md`: gRPC and Protobuf service conventions.
  - `104_event_schema_and_kafka_topic_standards.md`: Kafka topic naming and Protobuf schema definitions.
  - `111_domain_model_core_entities.md`: Core asset, wallet account, and currency models.
  - `112_idempotency_and_exactly_once_processing.md`: Distributed idempotency patterns and outbox mechanics.
  - `203_wallet_account_service.md`: Double-entry ledgers, balance holds, and asset credit APIs.
  - `237_multichain_mpc_tss_vault_custody_service.md`: Derivation of user Native SegWit and Taproot deposit addresses.
- **Parallel Tasks:**
  - `211_notification_service.md`: Real-time user push notifications for 0-conf and 3-conf events.
  - `216_compliance_aml_and_travel_rule_service.md`: Originator VASP screening and transaction risk scoring.
  - `208_trade_settlement_service.md`: Hyperledger Besu `wBTC` minting relayer and clearing coordinator.
  - `327_multichain_proof_of_reserve_registry_contract.md`: On-chain Proof-of-Reserve balance verification.
- **Downstream Blockers:**
  - `528_flutter_multichain_crypto_deposit_screen.md`: Mobile Bitcoin deposit QR code and live confirmation tracker UI.
  - `610_web_options_chain_and_crosschain_deposit_portal.md`: Web BTC deposit portal with real-time mempool status indicators.
