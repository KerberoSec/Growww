# 261 - High-Frequency Tick-by-Tick Market Replay & Audit Service (Rust / Go)

## Purpose
High-frequency electronic exchanges and regulated financial market infrastructures generate billions of nanosecond-stamped market data updates, order submissions, cancellations, modifications, and execution events daily. Maintaining absolute reproducibility of past market states is vital across three critical institutional dimensions:
1. **Algorithmic Backtesting and Quantitative Model Validation:** Institutional market makers, quantitative desks, and smart order routing (SOR) engines require deterministic, tick-by-tick simulation of limit order book (LOB) dynamics. Historical simulations must reproduce microsecond queue priorities, depth replenishment latencies, and market impact without lookahead bias or timing distortions.
2. **Regulatory Trade Reconstruction & Forensic Audit:** Under the Securities and Exchange Board of India (SEBI) Master Circular on Algorithmic Trading and the Prevention of Money Laundering Act (PMLA), exchanges and brokerages must maintain an immutable 8-year audit trail. Compliance officers must possess the capability to reconstruct the exact microsecond state of any order book during regulatory inquiries into suspected market manipulation (such as spoofing, layering, front-running, quote stuffing, or wash trading).
3. **Disaster Post-Mortems & Resilience Verification:** When extreme market volatility, network anomalies, or system failures occur, engineers must reproduce the precise cascade of inbound packets that triggered the incident. Replaying historical network captures (pcap) and matching engine Write-Ahead Logs (WAL) against shadow or sandbox environments allows deterministic root-cause identification and bug verification.

The **High-Frequency Tick-by-Tick Market Replay & Audit Service** (`services/market-replay-service`) provides an ultra-high-throughput, deterministic replay and audit subsystem built in Rust and Go. It streams historical tick archives and matching engine WAL logs into isolated sandbox matching engines, simulation environments, and surveillance modules at dynamically controllable replay speeds ranging from single-step manual advancement, real-time 1x playback, up to 1000x accelerated execution, and unconstrained dry-run throughput. It incorporates cryptographic SHA-256 hash chains for tamper detection and reconciles historical executions directly against Hyperledger Besu on-chain Delivery-versus-Payment (DvP) settlement transactions and block timestamps.

## What You Are Building
An enterprise-grade, deterministic historical market replay daemon and regulatory trade reconstruction service (`services/market-replay-service`). Concrete functional components include:
- **Zero-Copy Memory-Mapped Tick Reader & Decompressor (Rust):** High-speed stream reader utilizing `memmap2` and streaming Zstandard (zstd) decompression to ingest compressed historical tick extents and matching engine binary WAL segments directly from local NVMe staging storage without heap reallocation.
- **Sparse Nanosecond Index & Seek Engine (Rust):** High-density binary index generator mapping nanosecond timestamps and monotonic sequence numbers to physical byte offsets in compressed archives, enabling sub-millisecond random seeks across multi-terabyte datasets.
- **Precision Time Dilation & Pacing Loop (Rust):** Hardware-clock-driven scheduling engine using `clock_nanosleep(CLOCK_MONOTONIC_RAW)` and hybrid spin-wait polling. Maintains accurate inter-tick arrival intervals at user-defined time dilation factors ($1\times$, $5\times$, $10\times$, $50\times$, $100\times$, $500\times$, $1000\times$), while compensating for timer jitter and system sleep drift.
- **Replay Control Plane & Session Manager (Go):** gRPC microservice exposing fine-grained orchestration endpoints (`StartReplay`, `PauseReplay`, `ResumeReplay`, `SeekReplay`, `SetReplaySpeed`, `StepTick`, `GetReplayStatus`) to coordinate multiple concurrent replay sessions across isolated sandbox environments.
- **Multi-Transport Sandbox Dispatcher (Rust / Go):** Configurable event publication layer routing replayed ticks into dedicated sandbox Kafka topics (`sandbox.market.ticker.v1`, `sandbox.order.matching.commands.v1`), direct POSIX shared-memory ringbuffers for zero-latency in-process sandbox matching engines, or high-throughput client gRPC streaming.
- **SEBI 8-Year Cold Archive Staging Manager (Go):** Automated lifecycle manager orchestrating the retrieval and decompression of historical tick archives from Amazon S3 / S3 Glacier Deep Archive onto local NVMe staging tiers (`/var/data/growww/market-replay/staging`).
- **Cryptographic Hash Chain Tamper Auditor (Rust):** Continuous integrity validator computing cumulative SHA-256 hash chains across sequential ticks, verifying archives against daily Merkle roots notarized on Hyperledger Besu to detect historical data corruption or tampering.
- **On-Chain DvP & Block Timestamp Reconciler (Go):** Forensic auditing engine cross-referencing off-chain matching engine trade execution ticks with permissioned Hyperledger Besu `SettlementDvP.sol` contract transaction logs and QBFT block header timestamps.

## Scope Boundaries
- **In Scope:**
  - Ingestion, decompression, and parsing of compressed pcap packet captures and binary matching engine WAL logs (Prompt 246).
  - Nanosecond-precision tick replay with controllable speed scaling ($1\times$ to $1000\times$, pause, single-step forward, seek by sequence or timestamp).
  - High-performance sparse index generation for sub-millisecond seeking in multi-terabyte tick archives.
  - Event streaming into isolated Kafka sandbox topics, POSIX shared-memory ringbuffers, and gRPC client streams.
  - Cryptographic verification of historical tick archives using SHA-256 rolling hash chains and Besu on-chain Merkle roots.
  - Reconciliation of historical off-chain trade ticks with on-chain `SettlementDvP.sol` event logs and QBFT block timestamps.
  - Enforcing strict multi-tenant sandbox isolation to prevent replayed events from leaking into live production topics.
  - Generation of SEBI-compliant statutory trade reconstruction audit reports in JSON and Parquet formats.
- **Out of Scope / Handled Elsewhere:**
  - Live production order matching and real-time execution (handled in Prompt 205 Order Matching Engine).
  - Live production market data ticker broadcasting (handled in Prompt 207 Real-Time Market Data Service).
  - Real-time pre-trade risk evaluation for live orders (handled in Prompt 206 Risk and Margin Checks Service).
  - Live institutional FIX 5.0 gateway sessions (handled in Prompt 225 FIX Protocol Gateway).
  - Execution of proprietary quantitative backtesting strategy logic (handled in external client backtest runners).
  - Front-end backtest charting and audit visualization interface (handled in Prompt 601 Web Trading Terminal).

## Technology to Use
- **Core Replay & Decompression Engine (Rust):**
  - **Rust 1.78+ (2021 Edition):** Compiled with low-latency profile (`opt-level = 3`, `lto = "fat"`, `codegen-units = 1`, `panic = "abort"`).
  - **Memory Mapping:** `memmap2` for zero-copy file mapping; `libc` for `posix_fadvise` (`POSIX_FADV_SEQUENTIAL`, `POSIX_FADV_WILLNEED`) and `madvise`.
  - **Decompression:** `zstd` (libzstd C bindings with streaming decoder) supporting multi-threaded dictionary-based decompression.
  - **High-Resolution Clock:** `libc::clock_nanosleep` using `CLOCK_MONOTONIC_RAW` combined with CPU cycle counters (`core::arch::x86_64::_rdtsc`) for sub-microsecond timer calibration.
  - **Zero-Allocation Serialization:** `zerocopy` and `bytemuck` for casting binary WAL frames to in-memory structs without byte copies.
- **Control Plane & Archive Management (Go):**
  - **Go 1.22+:** Efficient concurrency with goroutines and channels for session orchestration and AWS SDK S3 lifecycle pipelines.
  - **gRPC Framework:** `google.golang.org/grpc` with Protocol Buffers v3 for high-throughput client and control APIs.
  - **Cloud Archive Access:** `aws-sdk-go-v2` for high-concurrency multipart byte-range downloads from S3 and Glacier retrieval coordination.
- **Event Streaming & Messaging:**
  - **Apache Kafka 3.7+:** Isolated sandbox cluster or dedicated sandbox topic namespaces (`sandbox.market.*`) using librdkafka (`rdkafka` crate in Rust, `confluent-kafka-go` in Go).
  - **Shared Memory IPC:** POSIX shared-memory (`shm_open`, `mmap`) ringbuffers for ultra-low-latency in-process feeding of sandbox matching engines.
- **Relational Metadata & Audit Store:**
  - **PostgreSQL 16+:** Relational storage for archive catalogs, replay session state, audit reconstruction tasks, and Besu verification proofs.
  - **TimescaleDB Extension:** Efficient storage of sparse index chunks and replay performance telemetry.
- **Blockchain Connectivity:**
  - **Go-Ethereum (`geth`) RPC Client:** Interfacing with permissioned Hyperledger Besu nodes over mTLS JSON-RPC to query `TradeAuditRegistry.sol` and `SettlementDvP.sol`.

## Backend / Infra Touchpoints
- **NVMe Staging Mounts:**
  - Local Staging Directory: `/var/data/growww/market-replay/staging/` storing active uncompressed and Zstandard-compressed chunk archives.
  - Index Directory: `/var/data/growww/market-replay/indexes/` holding pre-built binary sparse offset indexes (`*.rpi`).
  - Audit Export Directory: `/var/data/growww/market-replay/reports/` holding generated regulatory trade reconstruction reports.
- **S3 / Glacier Cloud Archive Paths:**
  - Cold Storage Bucket: `s3://growww-audit-archives-prod/ticks/{year}/{month}/{day}/{isin}.wal.zst`.
  - Daily Digest Bucket: `s3://growww-audit-archives-prod/digests/{year}/{month}/{day}/merkle_root.json`.
- **Kafka Topics:**
  - Sandbox Broadcast Topics: `sandbox.market.ticker.v1`, `sandbox.market.depth.v1`, `sandbox.order.matching.commands.v1`, `sandbox.engine.matches.v1`.
  - Audit & Control Topics: `audit.replay.session_events.v1`, `audit.replay.tamper_alerts.v1`.
- **Downstream Sandboxes & Internal Integrations:**
  - **Order Matching Engine Sandbox (Prompt 205):** Consumes replayed commands to reconstruct historical order book state and match executions in isolation.
  - **Real-Time Surveillance Sandbox (Prompt 228):** Consumes replayed ticks to backtest anomaly detection heuristics and wash trading algorithms against historical market crashes.
  - **Trade Settlement Service (Prompt 208):** Ingests reconstructed trade events for historical clearing settlement reconciliations.
  - **Prometheus & Grafana:** Scrapes real-time replay telemetry: `market_replay_ticks_emitted_total`, `market_replay_pacing_jitter_nanoseconds`, `market_replay_compression_throughput_mb_s`, `market_replay_active_sessions`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Immutable Daily Tick Merkle Root Anchoring:** At the conclusion of every trading day (17:00 IST), the archive pipeline aggregates all matching engine WAL frames and pcap feeds into immutable compressed archives. A Merkle tree is computed over the serialized tick records. The resulting 32-byte Merkle root, total record count, and end-of-day SHA-256 rolling chain hash are committed to `TradeAuditRegistry.sol` on Hyperledger Besu under QBFT consensus.
- **Pre-Replay Tamper Verification:** Before initializing any regulatory trade reconstruction session, `services/market-replay-service` computes the continuous SHA-256 digest of the staged historical archive. It queries `TradeAuditRegistry.sol` on Hyperledger Besu to verify that the archive's digest matches the on-chain notarized root. If a mismatch is detected, the replay is aborted and an alert is broadcast to `audit.replay.tamper_alerts.v1`.
- **On-Chain DvP Execution Reconciliation:** During regulatory trade investigations, off-chain trade execution ticks (`trade_id`, `wal_sequence_id`, `price_paise`, `quantity`) are matched against on-chain Delivery-versus-Payment events emitted by `SettlementDvP.sol` (Prompt 306). The service validates that every off-chain match executed by the matching engine correlates with an on-chain asset settlement token transfer within the matching QBFT block window.
- **Zero-PII Compliance Invariant:** Historical tick archives, hash chains, and on-chain audit entries strictly contain financial instrument codes (ISINs), execution sequence IDs, nanosecond timestamps, paise prices, and quantity integers. Investor identities (PAN, name, demat account number) are never present in tick replays or on-chain records; investor mapping is resolved only within authorized, air-gapped compliance vaults.

## High-Frequency Replay Mechanics & Architecture

### 1. High-Frequency Market Tick Binary Format & Rolling Hash Chain
Historical tick archives are stored as contiguous streams of fixed-size binary frames (64-byte header plus packed payload) compressed with Zstandard. Each frame carries an incremental SHA-256 chaining hash to ensure tamper-evidence:

$$\mathcal{H}_i = \text{SHA-256}\left( \mathcal{H}_{i-1} \parallel \text{Seq}_i \parallel \text{TimestampNS}_i \parallel \text{EventType}_i \parallel \text{Payload}_i \right)$$

```
+---------------------------------------------------------------------------------------------------+
|                               Binary Market Replay Frame (64-Byte Header)                         |
+-------------------+-------------------+-------------------+-------------------+-------------------+
| Magic (4B) 0x5250 | Version (2B)      | Event Type (2B)   | Sequence ID (8B)  | Timestamp NS (8B) |
+-------------------+-------------------+-------------------+-------------------+-------------------+
| Instrument ISIN (12B)                 | Flags (4B)        | Payload Len (4B)  | Prev Hash (16B)   |
+-------------------+-------------------+-------------------+-------------------+-------------------+
| Packed Payload Data (L1/L2/L3 Quote, Trade Execution, Order Command) - Variable Length            |
+---------------------------------------------------------------------------------------------------+
| Frame SHA-256 Checksum / Trailing Digest (16B truncated or 32B full)                              |
+---------------------------------------------------------------------------------------------------+
```

Event types supported in the replay stream:
- `0x01` (`TICK_L1_QUOTE`): Top-of-book best bid/offer updates with aggregate quantities.
- `0x02` (`TICK_L2_DEPTH`): 5-level or 10-level price-aggregated market depth diffs.
- `0x03` (`TICK_L3_ORDER_ADD`): Individual limit order additions entering the book with sequence numbers.
- `0x04` (`TICK_L3_ORDER_MODIFY`): Quantity modifications or price changes.
- `0x05` (`TICK_L3_ORDER_CANCEL`): Order cancellations and expiries.
- `0x06` (`TICK_TRADE_EXECUTION`): Matched trades containing execution ID, buyer/seller order IDs, price, and volume.
- `0x07` (`TICK_AUCTION_INDICATIVE`): Pre-open or LULD call auction indicative equilibrium prices and imbalances.
- `0x08` (`TICK_CIRCUIT_STATE`): Halts, volatility pauses, and band expansions.

### 2. Zero-Copy Memory-Mapped Indexing & Fast Seeking
To seek instantaneously across multi-gigabyte files spanning hundreds of millions of ticks without scanning from the beginning of the file, the service builds a companion sparse index file (`.rpi` - Replay Index):
- **Sparse Index Granularity:** An index entry is generated every 10,000 frames or every 100 milliseconds of trading time.
- **Index Entry Binary Structure (32 bytes):**
  - `timestamp_ns` ($u64$): Monotonic nanosecond timestamp of the indexed tick.
  - `sequence_id` ($u64$): Global matching engine sequence number.
  - `compressed_offset` ($u64$): Byte offset within the `.wal.zst` file pointing to the start of the Zstandard compression frame.
  - `uncompressed_offset` ($u32$): Relative byte offset within the decompressed frame window.
  - `frame_index` ($u32$): Monotonically increasing record index within the day's archive.
- **Binary Search Seeking:** Seeking to a timestamp $T_{\text{target}}$ performs a binary search over the memory-mapped `.rpi` table ($O(\log N)$ memory accesses), locates the closest preceding checkpoint entry, initializes the streaming Zstandard decompressor at `compressed_offset`, and discards uncompressed frames up to `sequence_id`. Seek latency is strictly bounded under 2 milliseconds on standard PCIe Gen4 NVMe drives.

### 3. Precision Nanosecond Pacing & Time Dilation Algorithm
Let $\Delta t_{\text{sim}} = t_{i+1} - t_i$ be the historical elapsed nanoseconds between two consecutive ticks. For a configured replay speed multiplier $S \in [0.001, 1000.0]$:

$$\Delta t_{\text{sleep}} = \frac{t_{i+1} - t_i}{S}$$

The pacing loop executes a hybrid sleep-spin schedule to minimize operating system context-switching overhead while eliminating microsecond timer jitter:
1. **Coarse Sleep Phase ($\Delta t_{\text{sleep}} > 100\,\mu\text{s}$):** If the required sleep duration exceeds 100 microseconds, invoke `libc::clock_nanosleep(CLOCK_MONOTONIC_RAW, 0, &ts_coarse, NULL)` for $\Delta t_{\text{sleep}} - 50\,\mu\text{s}$ to yield CPU execution to the operating system scheduler.
2. **Fine Spin-Wait Phase ($\Delta t_{\text{sleep}} \le 50\,\mu\text{s}$):** For the remaining sub-50 microsecond interval, enter a tight spin-wait loop querying the CPU Time Stamp Counter (`rdtsc`) or `clock_gettime(CLOCK_MONOTONIC_RAW)` with compiler memory fences (`std::hint::spin_loop()`) until the exact target nanosecond is reached.
3. **Dynamic Drift Compensation:** If system thread contention or scheduler preemption introduces a timing error $\epsilon_i = t_{\text{actual}} - t_{\text{scheduled}}$, the error is subtracted from the subsequent sleep interval: $\Delta t_{\text{sleep}}' = \Delta t_{\text{sleep}} - \epsilon_i$. If cumulative drift $\sum \epsilon_i$ exceeds 5 milliseconds (under extreme CPU saturation), the engine issues a telemetry warning and resets the timing baseline to prevent runaway catch-up bursts.
4. **Max-Speed / Unconstrained Mode ($S = \infty$):** Bypasses all sleep intervals entirely, pumping ticks directly into the downstream socket or ringbuffer as rapidly as the consumer can accept them (backpressure-limited).

```
          Historical Ticks:      [Tick 1] ---------- (200 µs) ---------- [Tick 2]
                                    |                                       |
    10x Speed Scaling (S=10):       v                                       v
          Scheduled Target:      [Emit 1] ----- (20 µs) ----- [Emit 2]
                                    |              ^
                                    |              | (Spin-Wait Phase: 20 µs)
                                    +--------------+
```

### 4. Sandbox Isolation & Multi-Tenant Routing
To guarantee absolute safety, the replay service operates in strict sandbox isolation mode:
- **Topic Namespace Segregation:** Replay topics are enforced via Kafka ACLs. Replayed streams write exclusively to topics prefixed with `sandbox.` or `audit.reconstruction.`. Under no circumstances can a replay session publish to production topics (`market.ticker.v1` or `order.matching.commands.v1`).
- **Simulated Wall-Clock Invariant:** The replay engine stamps every emitted message with two distinct timestamps:
  1. `simulated_tick_timestamp_ns`: The historical exchange execution time.
  2. `replayed_at_ns`: The current wall-clock time of physical emission.
  Downstream sandbox matching engines and risk evaluators run with mock time drivers keyed to `simulated_tick_timestamp_ns` to ensure temporal consistency during accelerated 100x replay.
- **Session Fencing:** Every replay session is assigned a unique `UUIDv4` `session_id`. Downstream sandbox consumers subscribe exclusively to consumer groups partitioned or tagged by `session_id`, allowing dozens of quantitative researchers to run concurrent backtests over overlapping historical periods without cross-talk.

### 5. SEBI Regulatory Trade Reconstruction Workflow
When SEBI or an authorized exchange surveillance officer requests a statutory trade reconstruction:
1. **Scope Definition:** Officer submits instrument ISIN, date, and microsecond observation window $[T_{\text{start}}, T_{\text{end}}]$ via gRPC.
2. **Cryptographic Validation:** Replay service verifies the daily historical archive against Hyperledger Besu `TradeAuditRegistry.sol`.
3. **Deterministic Sandbox Replay:** The engine positions to $T_{\text{start}} - 30\text{ minutes}$ in unconstrained mode to pre-warm the sandbox Order Matching Engine (Prompt 205) limit order book to the exact state, queues, and resting depths present immediately prior to the investigation window.
4. **Step-by-Step Execution:** At $T_{\text{start}}$, the replay switches to precision $1\times$ or single-tick manual stepping mode. Every order submission, cancel, modify, and execution is correlated with trading member identifiers and client PANs (retrieved from the off-chain secure vault).
5. **Audit Dossier Generation:** The service exports a tamper-evident audit dossier containing:
   - Full nanosecond order book depth snapshots before and after every trade.
   - Exact limit queue priority rankings for all participating participants.
   - Cross-reconciliation against Hyperledger Besu `SettlementDvP.sol` on-chain block events.
   - Cryptographic SHA-256 hash proofs certifying that the reconstructed sequence was identical to physical exchange execution.

## Step-by-Step Build Instructions (10-15 steps)
1. Initialize the dual-module workspace layout with Rust high-performance engine in `services/market-replay-service/core` and Go control plane in `services/market-replay-service/control`.
2. Define Protobuf schemas in `proto/growww/replay/v1/market_replay_service.proto` specifying replay control commands, speed parameters, tick frames, and audit report structures.
3. Generate Rust client/server types using `prost` and `tonic-build` in `core/build.rs`, and generate Go gRPC stubs using `protoc-gen-go` and `protoc-gen-go-grpc`.
4. Create PostgreSQL database migration scripts in `services/market-replay-service/migrations/001_market_replay_schema.sql` defining archive catalogs, replay sessions, tamper detection audits, and DvP reconciliation logs.
5. Implement the zero-copy memory-mapped binary reader in Rust utilizing `memmap2` and `madvise` to stream raw WAL extents from NVMe storage.
6. Implement the streaming Zstandard decompression pipeline in Rust with multi-threaded dictionary support to decompress compressed historical tick archives.
7. Implement the sparse binary index builder and search engine (`.rpi`) in Rust, providing $O(\log N)$ seeks on monotonic nanosecond timestamps and sequence IDs.
8. Implement the precision time dilation and pacing engine in Rust, combining `clock_nanosleep(CLOCK_MONOTONIC_RAW)` and CPU cycle spin-wait loops with dynamic drift compensation across speeds from $1\times$ to $1000\times$.
9. Implement the continuous SHA-256 hash chain verifier in Rust, recomputing frame hashes and asserting cryptographic parity against historical archive headers.
10. Implement the sandbox event dispatcher supporting Apache Kafka producer streaming (`sandbox.market.*`), POSIX shared-memory ringbuffers, and direct tonic gRPC server streaming.
11. Implement the Go control plane session manager orchestrating multi-tenant replay states (`IDLE`, `BUFFERING`, `PLAYING`, `PAUSED`, `SEEKING`, `COMPLETED`, `ABORTED`).
12. Implement the Go S3 / Glacier staging lifecycle manager, automating background retrieval, multipart download, and local NVMe cache eviction based on LRU disk quotas.
13. Implement the Hyperledger Besu on-chain verifier in Go using `go-ethereum` RPC to validate daily Merkle roots in `TradeAuditRegistry.sol` and reconcile historical executions with `SettlementDvP.sol` event logs.
14. Implement the SEBI statutory trade reconstruction report generator producing deterministic JSON, Parquet, and CSV audit dossiers with order book depth reconstruction.
15. Build comprehensive unit, benchmark, and regression test suites validating timer precision, zero-drift pacing under CPU load, zero-copy seek performance, and tamper detection on corrupted archives.

## Interfaces / Contracts

### 1. Protobuf Service Contract (`proto/growww/replay/v1/market_replay_service.proto`)
```protobuf
syntax = "proto3";

package growww.replay.v1;

option go_package = "github.com/growww/proto/gen/go/replay/v1;replayv1";

enum ReplaySessionState {
  REPLAY_SESSION_STATE_UNSPECIFIED = 0;
  REPLAY_SESSION_STATE_INITIALIZING = 1;
  REPLAY_SESSION_STATE_BUFFERING = 2;
  REPLAY_SESSION_STATE_PLAYING = 3;
  REPLAY_SESSION_STATE_PAUSED = 4;
  REPLAY_SESSION_STATE_SEEKING = 5;
  REPLAY_SESSION_STATE_STEPPING = 6;
  REPLAY_SESSION_STATE_COMPLETED = 7;
  REPLAY_SESSION_STATE_ABORTED = 8;
}

enum ReplaySpeed {
  REPLAY_SPEED_UNSPECIFIED = 0;
  REPLAY_SPEED_STEP_SINGLE_TICK = 1;  // Advance one tick on command
  REPLAY_SPEED_0_1X = 2;              // 10x Slow Motion
  REPLAY_SPEED_0_5X = 3;              // 2x Slow Motion
  REPLAY_SPEED_1X = 4;                // Real-Time (1:1)
  REPLAY_SPEED_5X = 5;                // 5x Accelerated
  REPLAY_SPEED_10X = 6;               // 10x Accelerated
  REPLAY_SPEED_50X = 7;               // 50x Accelerated
  REPLAY_SPEED_100X = 8;              // 100x Accelerated
  REPLAY_SPEED_500X = 9;              // 500x Accelerated
  REPLAY_SPEED_1000X = 10;            // 1000x Accelerated
  REPLAY_SPEED_UNCONSTRAINED = 11;    // Maximum Throughput (Dry Run)
}

enum ReplayEventType {
  REPLAY_EVENT_TYPE_UNSPECIFIED = 0;
  REPLAY_EVENT_TYPE_L1_QUOTE = 1;
  REPLAY_EVENT_TYPE_L2_DEPTH_DIFF = 2;
  REPLAY_EVENT_TYPE_L3_ORDER_ADD = 3;
  REPLAY_EVENT_TYPE_L3_ORDER_MODIFY = 4;
  REPLAY_EVENT_TYPE_L3_ORDER_CANCEL = 5;
  REPLAY_EVENT_TYPE_TRADE_EXECUTION = 6;
  REPLAY_EVENT_TYPE_AUCTION_INDICATIVE = 7;
  REPLAY_EVENT_TYPE_CIRCUIT_STATE = 8;
}

message StartReplayRequest {
  string session_name = 1;
  repeated string isins = 2;
  int64 start_timestamp_ns = 3;
  int64 end_timestamp_ns = 4;
  ReplaySpeed initial_speed = 5;
  string destination_kafka_prefix = 6; // Defaults to "sandbox."
  bool enable_shared_memory_ipc = 7;
  bool verify_blockchain_hash_chain = 8;
}

message StartReplayResponse {
  string session_id = 1;
  ReplaySessionState state = 2;
  int64 total_ticks_in_scope = 3;
  int64 estimated_duration_seconds = 4;
  string besu_merkle_root_verified = 5;
  int64 started_at_ns = 6;
}

message PauseReplayRequest {
  string session_id = 1;
}

message ResumeReplayRequest {
  string session_id = 1;
}

message StopReplayRequest {
  string session_id = 1;
  string reason = 2;
}

message SetReplaySpeedRequest {
  string session_id = 1;
  ReplaySpeed new_speed = 2;
}

message SeekReplayRequest {
  string session_id = 1;
  oneof seek_target {
    int64 target_timestamp_ns = 2;
    uint64 target_sequence_id = 3;
  }
}

message StepTickRequest {
  string session_id = 1;
  uint32 tick_count = 2; // Number of ticks to step forward (default 1)
}

message GetReplayStatusRequest {
  string session_id = 1;
}

message ReplayStatusResponse {
  string session_id = 1;
  ReplaySessionState state = 2;
  ReplaySpeed current_speed = 3;
  uint64 current_sequence_id = 4;
  int64 current_timestamp_ns = 5;
  uint64 total_ticks_replayed = 6;
  uint64 total_ticks_remaining = 7;
  double replay_progress_percent = 8;
  int64 cumulative_drift_nanoseconds = 9;
  double actual_replay_speed_multiplier = 10;
  bool is_tamper_verified = 11;
}

message StreamReplayEventsRequest {
  string session_id = 1;
  repeated string isins = 2;
  uint32 batch_size = 3;
}

message ReplayTickFrame {
  uint64 sequence_id = 1;
  int64 simulated_timestamp_ns = 2;
  int64 replayed_at_ns = 3;
  string isin = 4;
  ReplayEventType event_type = 5;
  bytes raw_payload = 6;
  bytes sha256_cumulative_hash = 7;
}

message ReplayTickBatch {
  string session_id = 1;
  repeated ReplayTickFrame frames = 2;
}

message TriggerTradeReconstructionRequest {
  string investigation_case_id = 1;
  string isin = 2;
  int64 window_start_timestamp_ns = 3;
  int64 window_end_timestamp_ns = 4;
  repeated string suspicious_trader_ids = 5;
  bool reconcile_blockchain_dvp = 6;
}

message TradeReconstructionReport {
  string report_id = 1;
  string investigation_case_id = 2;
  string isin = 3;
  int64 window_start_timestamp_ns = 4;
  int64 window_end_timestamp_ns = 5;
  uint64 total_orders_in_window = 6;
  uint64 total_trades_in_window = 7;
  uint64 total_turnover_paise = 8;
  bool blockchain_dvp_reconciled = 9;
  uint32 dvp_matched_events = 10;
  uint32 dvp_unmatched_events = 11;
  string archive_merkle_root = 12;
  string report_download_uri = 13;
  int64 generated_at_ns = 14;
}

service MarketReplayService {
  rpc StartReplay (StartReplayRequest) returns (StartReplayResponse);
  rpc PauseReplay (PauseReplayRequest) returns (ReplayStatusResponse);
  rpc ResumeReplay (ResumeReplayRequest) returns (ReplayStatusResponse);
  rpc StopReplay (StopReplayRequest) returns (ReplayStatusResponse);
  rpc SetReplaySpeed (SetReplaySpeedRequest) returns (ReplayStatusResponse);
  rpc SeekReplay (SeekReplayRequest) returns (ReplayStatusResponse);
  rpc StepTick (StepTickRequest) returns (ReplayStatusResponse);
  rpc GetReplayStatus (GetReplayStatusRequest) returns (ReplayStatusResponse);
  rpc StreamReplayEvents (StreamReplayEventsRequest) returns (stream ReplayTickBatch);
  rpc TriggerTradeReconstruction (TriggerTradeReconstructionRequest) returns (TradeReconstructionReport);
}
```

### 2. PostgreSQL Database Schema (`services/market-replay-service/migrations/001_market_replay_schema.sql`)
```sql
-- Historical Archive Catalog Table
CREATE TABLE replay_archive_catalogs (
    archive_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    trading_date DATE NOT NULL,
    archive_type VARCHAR(32) NOT NULL DEFAULT 'MATCHING_ENGINE_WAL', -- MATCHING_ENGINE_WAL, PCAP_CAPTURE
    s3_uri VARCHAR(512) NOT NULL,
    local_staging_path VARCHAR(512),
    compression_algorithm VARCHAR(16) NOT NULL DEFAULT 'ZSTD',
    uncompressed_size_bytes BIGINT NOT NULL,
    compressed_size_bytes BIGINT NOT NULL,
    total_records BIGINT NOT NULL,
    start_sequence_id BIGINT NOT NULL,
    end_sequence_id BIGINT NOT NULL,
    start_timestamp_ns BIGINT NOT NULL,
    end_timestamp_ns BIGINT NOT NULL,
    sha256_root_hash VARCHAR(66) NOT NULL,
    besu_merkle_root_tx VARCHAR(66),
    is_staged_locally BOOLEAN NOT NULL DEFAULT FALSE,
    last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_archive_isin_date UNIQUE (isin, trading_date, archive_type)
);

-- Replay Active Sessions
CREATE TABLE replay_sessions (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_name VARCHAR(128) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'INITIALIZING', -- INITIALIZING, BUFFERING, PLAYING, PAUSED, SEEKING, COMPLETED, ABORTED
    current_speed VARCHAR(32) NOT NULL DEFAULT 'REPLAY_SPEED_1X',
    speed_multiplier DOUBLE PRECISION NOT NULL DEFAULT 1.0,
    current_sequence_id BIGINT NOT NULL DEFAULT 0,
    current_timestamp_ns BIGINT NOT NULL DEFAULT 0,
    total_ticks_replayed BIGINT NOT NULL DEFAULT 0,
    total_ticks_in_scope BIGINT NOT NULL DEFAULT 0,
    destination_kafka_prefix VARCHAR(64) NOT NULL DEFAULT 'sandbox.',
    enable_shared_memory_ipc BOOLEAN NOT NULL DEFAULT FALSE,
    cumulative_drift_ns BIGINT NOT NULL DEFAULT 0,
    created_by VARCHAR(64) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resumed_at TIMESTAMPTZ,
    paused_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    error_message TEXT
);

-- Replay Session Active Instruments
CREATE TABLE replay_session_instruments (
    session_id UUID NOT NULL REFERENCES replay_sessions(session_id) ON DELETE CASCADE,
    isin VARCHAR(12) NOT NULL,
    archive_id UUID NOT NULL REFERENCES replay_archive_catalogs(archive_id),
    ticks_replayed BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (session_id, isin)
);

-- Regulatory Trade Reconstruction Requests & Audits
CREATE TABLE replay_trade_reconstructions (
    reconstruction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    case_id VARCHAR(64) NOT NULL,
    requested_by VARCHAR(64) NOT NULL,
    regulatory_mandate VARCHAR(64) NOT NULL DEFAULT 'SEBI_ALGO_AUDIT_2018',
    isin VARCHAR(12) NOT NULL,
    window_start_ns BIGINT NOT NULL,
    window_end_ns BIGINT NOT NULL,
    total_orders_reconstructed BIGINT NOT NULL DEFAULT 0,
    total_trades_reconstructed BIGINT NOT NULL DEFAULT 0,
    total_turnover_paise BIGINT NOT NULL DEFAULT 0,
    is_tamper_verified BOOLEAN NOT NULL DEFAULT FALSE,
    blockchain_dvp_reconciled BOOLEAN NOT NULL DEFAULT FALSE,
    dvp_matched_count INT NOT NULL DEFAULT 0,
    dvp_unmatched_count INT NOT NULL DEFAULT 0,
    archive_merkle_root VARCHAR(66) NOT NULL,
    besu_block_height BIGINT,
    report_storage_uri VARCHAR(512),
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tamper Detection Audit Event Log
CREATE TABLE replay_tamper_audit_logs (
    audit_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    archive_id UUID NOT NULL REFERENCES replay_archive_catalogs(archive_id),
    sequence_id_anomaly BIGINT,
    expected_hash VARCHAR(66) NOT NULL,
    computed_hash VARCHAR(66) NOT NULL,
    tamper_detected BOOLEAN NOT NULL DEFAULT FALSE,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action_taken VARCHAR(128) NOT NULL
);

-- On-Chain DvP Reconciliation Ledger
CREATE TABLE replay_dvp_reconciliations (
    reconciliation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reconstruction_id UUID NOT NULL REFERENCES replay_trade_reconstructions(reconstruction_id) ON DELETE CASCADE,
    wal_trade_id VARCHAR(64) NOT NULL,
    wal_sequence_id BIGINT NOT NULL,
    isin VARCHAR(12) NOT NULL,
    execution_price_paise BIGINT NOT NULL,
    execution_quantity BIGINT NOT NULL,
    execution_timestamp_ns BIGINT NOT NULL,
    besu_tx_hash VARCHAR(66),
    besu_block_number BIGINT,
    besu_block_timestamp_ns BIGINT,
    settlement_status VARCHAR(32) NOT NULL, -- MATCHED_DVP, UNMATCHED_PENDING, FAILED_ONCHAIN
    latency_delta_ms BIGINT,
    reconciled_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for Ultra-Fast Lookups
CREATE INDEX idx_replay_catalogs_isin_date ON replay_archive_catalogs(isin, trading_date);
CREATE INDEX idx_replay_catalogs_staged ON replay_archive_catalogs(is_staged_locally, last_accessed_at);
CREATE INDEX idx_replay_sessions_status ON replay_sessions(status);
CREATE INDEX idx_reconstruction_case ON replay_trade_reconstructions(case_id);
CREATE INDEX idx_dvp_recon_trade ON replay_dvp_reconciliations(reconstruction_id, wal_trade_id);
CREATE INDEX idx_tamper_archive ON replay_tamper_audit_logs(archive_id, detected_at DESC);
```

## Security & Compliance Notes
- **SEBI 8-Year Trade Reconstruction Mandate:** Under SEBI Circulars `SEBI/HO/MRD/DP/CIR/P/2018/62` and PMLA Section 12, all order events, cancellations, modifications, and matched trades must be preserved and fully reconstructible for a minimum of 8 years. The service maintains indexed Zstandard archives with long-term retention policies mapped to Amazon S3 Glacier Flexible Deep Archive.
- **Cryptographic Tamper Detection & SHA-256 Hash Chains:** Every tick archive incorporates an immutable rolling SHA-256 hash chain where each record's hash depends on the preceding record's digest. Any modification, record reordering, insertion, or truncation invalidates the terminal Merkle root, which is checked against Hyperledger Besu `TradeAuditRegistry.sol` before replay.
- **Air-Gapped Sandbox Topic Isolation:** Replay topics are strictly segregated from production trading infrastructure. The service enforces dynamic topic prefixing (`sandbox.*`), independent Kafka user ACLs, and isolated consumer groups, preventing simulated ticks from ever entering live production matching engines or risk calculators.
- **Zero-PII Storage Policy:** Market replay data contains solely anonymized numeric identifiers: ISINs, token codes, internal UUIDs, paise prices, and share quantities. Real investor names, PANs, email addresses, and bank accounts are strictly excluded from the replay stream and stored in encrypted off-chain institutional KYC vaults.
- **Monotonic Clock Jitter & Leap Second Handling:** High-resolution pacing utilizes `CLOCK_MONOTONIC_RAW` on Linux to prevent NTP time stepping or leap-second adjustments from warping replay pacing intervals.

## Acceptance Criteria
- [ ] Protobuf service contracts compile cleanly with `tonic-build` in Rust and `protoc-gen-go` in Go without warnings.
- [ ] PostgreSQL migration script `001_market_replay_schema.sql` applies idempotently and establishes all tables, constraints, and indexes.
- [ ] Memory-mapped Zstandard tick reader decompresses historical WAL archives at throughput $> 500\text{ MB/s}$ on NVMe storage.
- [ ] Sparse binary index (`.rpi`) provides sub-2ms random seek to any nanosecond timestamp across a 50GB tick archive.
- [ ] Precision pacing engine maintains timing accuracy within $\pm 5\,\mu\text{s}$ at $1\times$ real-time replay under benchmark conditions.
- [ ] Speed scaling accurately scales pacing intervals across $1\times$, $5\times$, $10\times$, $50\times$, $100\times$, $500\times$, $1000\times$, and unconstrained modes.
- [ ] SHA-256 rolling hash chain verifier successfully identifies deliberate single-bit tampering or dropped records in test tick archives.
- [ ] Replay control gRPC endpoints (`Start`, `Pause`, `Resume`, `Stop`, `Seek`, `SetSpeed`, `StepTick`) transition session state deterministically.
- [ ] Sandbox Kafka dispatcher strictly enforces `sandbox.` topic naming and prevents writes to production topics.
- [ ] Hyperledger Besu on-chain reconciler correctly matches off-chain matching engine trades to `SettlementDvP.sol` events and block timestamps.
- [ ] Complies strictly with the 12-section prompt specification template with zero raw application code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `104` (Event Schema & Kafka Topic Standards), Prompt `205` (Order Matching Engine), Prompt `246` (Matching Engine Memory-Mapped WAL & Hot-Warm Shadow Failover), Prompt `306` (Atomic Delivery-versus-Payment Settlement Smart Contract).
- **Parallel Work:** Prompt `207` (Real-Time Market Data Service), Prompt `228` (Real-Time Market Surveillance Engine), Prompt `256` (Limit-Up / Limit-Down Volatility Dampener Service).
- **Subsequent Prompts Enabled:** Prompt `501` (Flutter Mobile Trading Interface), Prompt `601` (Web Trading Terminal backtest visualization modules).
