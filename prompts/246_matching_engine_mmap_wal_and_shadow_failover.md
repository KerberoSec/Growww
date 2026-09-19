# 246 - Matching Engine Memory-Mapped WAL & Hot-Warm Shadow Failover (Rust)

## Purpose
High-frequency financial market infrastructures, institutional exchanges, and national securities depositories require uninterrupted continuous trading availability with strict guarantees of zero data loss (Recovery Point Objective RPO = 0) and sub-second failover recovery (Recovery Time Objective RTO < 50ms). In an exchange processing tens of thousands of orders per second with microsecond-level execution latencies, conventional relational database transactions, asynchronous disk logging, or naive fsync system calls introduce severe latency jitter, thread blocking, and unacceptable risk of state divergence during ungraceful process crashes, power outages, or hardware faults.

The **Matching Engine Memory-Mapped WAL & Hot-Warm Shadow Failover Subsystem** (`services/matching-engine-wal`) delivers a bare-metal, high-performance persistence and replication architecture for the Rust Order Matching Engine (Prompt 205). It combines a zero-copy append-only memory-mapped (mmap) Write-Ahead Log (WAL) backed by kernel-level synchronous page flushing (`MS_SYNC` / `msync`), lock-free ringbuffer command sequencing, hot-warm shadow replica parallel execution with outbound match suppression, automated sub-millisecond leader lease fencing, and periodic SHA-256 depth state checksum verification executed deterministically every 10,000 matches to guarantee mathematical consistency across replicated order books.

## What You Are Building
A bare-metal optimized, ultra-low-latency Rust persistence and high-availability subsystem (`services/matching-engine-wal`). Concrete deliverables include:
- **Zero-Copy Append-Only Memory-Mapped WAL Engine:** High-speed file-backed memory-mapped log manager utilizing `memmap2` and direct `libc` system calls (`posix_fallocate`, `msync` with `MS_SYNC`, `madvise`, `mlock`) operating on fixed-size, cache-line aligned (64-byte and 128-byte) binary frame layouts to eliminate heap allocations and runtime serialization overhead.
- **Monotonic Sequenced Command Journal & Indexer:** Monotonically increasing 64-bit sequence generator and cyclic segment rotator managing pre-allocated 2GB WAL log extents with nanosecond-precision timing, magic byte validation, and CRC32C frame integrity verification.
- **Hot-Warm Shadow Replica Parallel Matching Engine:** Redundant matching engine replica architecture where the shadow node continuously ingests the identical sequenced command stream, executes deterministic Price-Time FIFO matching in parallel on its local in-memory Limit Order Book (LOB), updates book state in lockstep with the primary node, and strictly suppresses outbound event publication to Apache Kafka until promoted.
- **Deterministic SHA-256 Depth State Checksum Verifier:** Continuous integrity auditor computing cryptographic SHA-256 digests across active bid/ask price levels, resting order queues, and aggregate volume every 10,000 matches. Compares primary and shadow state checksums over an out-of-band peer heartbeat channel to immediately detect and isolate state divergence.
- **Sub-Millisecond Leader Fencing & Failover Controller:** Distributed lease manager backed by etcd/Raft with monotonically increasing epoch fencing tokens, detecting primary heartbeat failure within 15ms, executing atomic shadow promotion, draining uncommitted match queues, and rebinding Kafka event publishing without duplicate trade emissions.
- **Ultra-Fast Crash Recovery & Replay Engine:** Deterministic sequential WAL replay subsystem scanning mmap segments and base snapshots to reconstruct 1,000,000 resting orders and matching state in $< 500\text{ms}$ during node cold starts.
- **Administrative & Telemetry gRPC Interface:** Dedicated control plane service exposing real-time segment metrics, replication lag, checksum audit logs, manual failover triggers, and symbol-level WAL health indicators.

## Scope Boundaries
- **In Scope:**
  - Memory-mapped file management, chunk pre-allocation via `posix_fallocate`, and synchronous page flushing via `msync(..., MS_SYNC)`.
  - Binary WAL record framing for New Orders, Order Cancels, Order Replaces, Mass Cancels, Trades/Matches, Checkpoints, and Epoch Fencing tokens.
  - Hot-warm shadow replica parallel matching and deterministic local state maintenance.
  - Outbound event emission suppression on shadow replicas during normal operation.
  - Periodic SHA-256 order book depth state checksum calculation every 10,000 matches.
  - Out-of-band primary-to-shadow checksum verification, divergence alerting, and automated shadow quarantine.
  - Distributed lease acquisition, heartbeat health monitoring, epoch fencing token validation, and sub-50ms promotion cutover.
  - Cold-start WAL replay and state recovery from local NVMe storage.
  - gRPC control plane APIs and Prometheus metrics instrumentation.
- **Out of Scope / Handled Elsewhere:**
  - In-memory limit order book matching algorithm implementation (handled in Prompt 205).
  - Pre-trade risk policy limit evaluation and margin reservation (handled in Prompt 206).
  - Client REST/WebSocket market data broadcasting (handled in Prompt 207).
  - On-chain Delivery-versus-Payment (DvP) smart contract settlement (handled in Prompt 208 and Prompt 306).
  - Institutional FIX 5.0 / OUCH / ITCH protocol session management (handled in Prompt 225).
  - User wallet cash holds and double-entry ledger persistence (handled in Prompt 203).

## Technology to Use
- **Primary Language & Toolchain:** Rust 1.78+ (2021 Edition) compiled with maximum optimization flags (`opt-level = 3`, `lto = "fat"`, `codegen-units = 1`, `panic = "abort"`, `target-cpu = "native"`). Rust provides deterministic microsecond execution without garbage collection pauses, zero-cost memory safety, and direct hardware memory mapping.
- **Memory-Mapped I/O & System Primitives:** `memmap2` crate for safe memory mapping abstractions; `libc` for direct Linux system calls including `posix_fallocate`, `msync`, `madvise` (`MADV_SEQUENTIAL`, `MADV_WILLNEED`, `MADV_DONTDUMP`), and `mlock` to lock WAL buffers into physical RAM and prevent swap paging.
- **Concurrency & Memory Safety:** `crossbeam-channel`, `parking_lot`, standard atomic primitives (`std::sync::atomic::*`), and lock-free ringbuffers (`rtrb` / `crossbeam-queue`) for zero-contention command routing between network ingestion threads, matching threads, and WAL flusher threads.
- **Hashing & Checksums:** `sha2` (RustCrypto with hardware AVX2/AVX-512 and ARM Neon acceleration) for 10,000-match state depth checksums; `crc32fast` for sub-nanosecond CRC32C frame validation.
- **Binary Serialization & Wire Encodings:** `zerocopy` and `bytemuck` for zero-allocation casting between raw byte slices and C-ABI packed structs (`#[repr(C, packed)]` / `#[repr(C, align(64))]`).
- **Distributed Coordination & Fencing:** `etcd-client` for distributed lease management, heartbeat keep-alive leases, and monotonic epoch fencing tokens.
- **Messaging & Inter-Service RPC:** `tonic` / `prost` for gRPC control plane and peer replication channels; `rdkafka` (librdkafka C bindings) for Kafka event streaming.
- **Relational Persistence (Audit & Metrics):** PostgreSQL 16+ for long-term WAL segment indexing, checksum verification logs, and failover incident records.
- **Storage Hardware Architecture:** Direct-attached NVMe PCIe Gen4/Gen5 solid-state drives formatted with XFS/ext4 filesystems mounted with `noatime`, `nodiratime`, and direct I/O tuning.

## Backend / Infra Touchpoints
- **Direct NVMe Storage Paths:**
  - WAL Segment Directory: `/var/data/growww/matching-engine/wal/` containing pre-allocated 2GB segment files (`wal_segment_00000001.wal`, `wal_segment_00000002.wal`).
  - Checkpoint Snapshot Directory: `/var/data/growww/matching-engine/snapshots/` containing periodic serialized LOB base images.
  - Quarantine Dump Directory: `/var/data/growww/matching-engine/quarantine/` for post-mortem forensic state dumps on checksum mismatch.
- **Apache Kafka Topics:**
  - Consumes: `order.matching.commands.v1` (partitioned by ISIN).
  - Publishes (Primary Leader Only): `engine.matches.v1` (execution reports), `matching.depth.v1` (L2/L3 market depth diffs).
  - Publishes (Both Primary & Shadow): `matching.wal.checksum.v1` (audit state checksums), `engine.failover.events.v1` (leader election and failover transition records).
- **Core Order Matching Engine (Prompt 205):** Direct in-process binding and synchronous WAL pre-commit hook preceding order book state mutation.
- **Pre-Trade Risk & Margin Engine (Prompt 206):** Ingests execution receipts to release or adjust margin reservations.
- **Trade Settlement & DvP Orchestration Service (Prompt 208):** Ingests deduplicated execution match records tagged with immutable WAL sequence identifiers for atomic on-chain settlement.
- **Prometheus & Grafana:** Real-time observability scraping metrics for `mmap_wal_sync_latency_nanoseconds`, `mmap_wal_bytes_written_total`, `shadow_replica_lag_records`, `checksum_verification_status`, and `failover_duration_milliseconds`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Deterministic Match Receipts & Sequence Anchoring:** Every trade execution logged to the WAL is assigned a monotonically increasing 64-bit sequence number (`wal_sequence_id`). When match events are forwarded to the Trade Settlement Service (Prompt 208), this sequence number is embedded into the cryptographic match receipt submitted to `SettlementDvP.sol` (Prompt 306) on Hyperledger Besu under QBFT consensus.
- **On-Chain State Checksum Attestation:** The SHA-256 depth state checksums generated every 10,000 matches can be periodically anchored into an on-chain verification registry (`MatchingEngineProofRegistry.sol`), providing an immutable, tamper-evident public audit trail proving that off-chain order execution was continuous and strictly deterministic.
- **Zero On-Chain PII Invariant:** All WAL records, replication frames, and on-chain attestations contain solely numeric sequence identifiers, UUIDs, ISIN security codes, fixed-point integer quantities, sub-paise prices, and cryptographic SHA-256 hashes. No investor names, PANs, email addresses, or banking details ever enter the WAL binary structure or the blockchain ledger.

## WAL & Shadow Failover Architecture

### 1. Memory-Mapped WAL Binary Framing & Flush Lifecycle
The WAL operates on pre-allocated contiguous file segments (standard size: 2,147,483,648 bytes / 2GB) mapped into the matching engine's virtual memory space via `mmap`. Write operations execute as direct pointer copies into the mapped buffer without heap allocation or intermediate copying.

Each WAL record is formatted as a 64-byte or 128-byte aligned binary frame consisting of:
1. **Magic Header (4 bytes):** Constant signature `0x57414C31` (`WAL1`) identifying valid frame boundaries.
2. **Epoch / Fencing Token (8 bytes):** Monotonically increasing cluster leadership epoch ($u64$) assigned by the distributed lease coordinator. Writes with a stale epoch are immediately rejected.
3. **Sequence Number (8 bytes):** Monotonically increasing sequence identifier ($u64$) guaranteeing strict global order per trading instrument / ISIN.
4. **Timestamp (8 bytes):** Nanoseconds since Unix epoch ($u64$) recorded via monotonic hardware clock (`CLOCK_MONOTONIC_RAW`).
5. **Frame Type (2 bytes):** Enumeration indicating command type (`ORDER_NEW`, `ORDER_CANCEL`, `ORDER_REPLACE`, `ORDER_MASS_CANCEL`, `TRADE_MATCH`, `CHECKPOINT_START`, `CHECKPOINT_END`, `CHECKSUM_ATTESTATION`).
6. **Payload Length (2 bytes):** Size of the payload payload structure in bytes ($u16$).
7. **Payload Data (Variable / Fixed Padded):** C-ABI packed payload containing fixed-point prices ($10^{-4}$ INR), fractional quantities ($10^{-6}$ shares), user ID UUIDs, order ID UUIDs, and side indicators.
8. **Frame CRC32C (4 bytes):** Hardware-computed CRC32C checksum validating frame payload and header integrity.

Upon copying a frame into the mmap buffer, the WAL writer initiates kernel page synchronization:
- In strict durability mode: Calls `msync(ptr, len, MS_SYNC)` to force immediate hardware flush to NVMe storage before publishing match acknowledgement.
- In batched ringbuffer mode: Calls `msync(ptr, len, MS_ASYNC)` with a dedicated background thread performing `MS_SYNC` flushes every 1,000 microseconds or every 100 accumulated records.

```
+---------------------------------------------------------------------------------------------------+
|                                 64-Byte / 128-Byte Aligned WAL Frame                              |
+-------------------+-------------------+-------------------+-------------------+-------------------+
| Magic (4B) 0x57.. | Epoch Token (8B)  | Sequence ID (8B)  | Timestamp (8B)    | Frame Type (2B)   |
+-------------------+-------------------+-------------------+-------------------+-------------------+
| Payload Len (2B)  | Payload Data (e.g. Price, Qty, Order IDs, ISIN)           | CRC32C (4B)       |
+---------------------------------------------------------------------------------------------------+
```

### 2. Hot-Warm Shadow Replica Parallel Execution Model
High availability is achieved through an active-shadow paired deployment across independent physical availability zones or bare-metal racks:
- **Primary Node (Leader):** Receives incoming orders from the ingress pipeline, writes binary frames to its local mmap WAL, executes matching against its in-memory Limit Order Book (LOB), produces trade executions, logs execution records to WAL, and publishes match events to Kafka (`engine.matches.v1`) and market depth to `matching.depth.v1`.
- **Shadow Node (Follower):** Ingests the identical ordered command stream via a low-latency peer replication stream (or shared sequenced Kafka partition). The shadow logs incoming commands to its local replica WAL, updates its in-memory LOB by running the identical deterministic matching algorithm, but **actively suppresses** all outbound event publishing to Kafka and downstream settlement queues.
- **Outbound Suppression Switch:** The shadow's Kafka producer is held in a dormant, suppressed state (`is_active_leader = false`). Match events generated by the shadow's local matching loop are buffered into an in-memory ringbuffer with a circular sliding window of the last 50,000 matches.

```
                    +---------------------------------------+
                    |  Kafka / Ingress: Inbound Orders      |
                    +-------------------+-------------------+
                                        | (Identical Stream)
                     +------------------+------------------+
                     |                                     |
                     v                                     v
       +----------------------------+        +----------------------------+
       |   Primary Matching Node    |        |   Shadow Matching Node     |
       |  (Active Leader, Epoch N)  |        |  (Warm Replica, Epoch N)   |
       +--------------+-------------+        +--------------+-------------+
       | Local mmap WAL (MS_SYNC)   |        | Local Replica mmap WAL     |
       | In-Memory LOB Matching     |        | In-Memory LOB Matching     |
       | Outbound Kafka: ENABLED    |        | Outbound Kafka: SUPPRESSED |
       +--------------+-------------+        +--------------+-------------+
                      |                                     |
                      | 10,000 Matches Checksum Broadcast   |
                      +------------------+------------------+
                                         |
                                         v
                      +-------------------------------------+
                      | Checksum Comparator & Heartbeat     |
                      |  - Parity: OK                       |
                      |  - Discrepancy: Alert & Quarantine  |
                      +-------------------------------------+
```

### 3. SHA-256 Depth State Checksum Verification (Every 10,000 Matches)
To prevent silent state drift, memory corruption, or non-deterministic execution divergence between the Primary and Shadow nodes, a cryptographic state checksum is computed every 10,000 matches:
1. **Match Counter Trigger:** Each node maintains an atomic match counter per trading instrument (`match_counter`). When `match_counter % 10000 == 0`, a state checksum evaluation is triggered immediately following the match execution.
2. **Canonical State Serialization:** The engine creates an in-memory canonical byte representation of the LOB without allocating heap memory:
   - Header: ISIN (12 bytes), Monotonic Sequence ID (8 bytes), Match Counter (8 bytes), Timestamp (8 bytes).
   - Bids Array: Array of `(Price, TotalQuantity, OrderCount)` sorted descending by Price.
   - Asks Array: Array of `(Price, TotalQuantity, OrderCount)` sorted ascending by Price.
   - Top-K Order Hashes: SHA-256 hash chain of the first 100 resting order IDs and quantities at each top price level.
3. **Cryptographic Checksum Calculation:** The canonical byte stream is hashed using SHA-256:
   $$\text{DepthChecksum} = \text{SHA-256}(\text{Header} \parallel \text{CanonicalBids} \parallel \text{CanonicalAsks} \parallel \text{OrderHashChain})$$
4. **Peer Checksum Exchange & Verification:**
   - The Primary publishes a `ChecksumAttestation` frame over the peer replication channel and emits a `matching.wal.checksum.v1` event.
   - The Shadow computes its local `DepthChecksum` at the exact same sequence point and compares it against the Primary's checksum.
5. **Divergence Handling & Quarantine:** If `PrimaryChecksum != ShadowChecksum`:
   - An immediate high-priority alert (`MATCHING_ENGINE_STATE_DIVERGENCE_CRITICAL`) is broadcast to Prometheus/PagerDuty.
   - The Shadow replica enters a **Quarantine State**: it ceases active synchronization, refuses promotion, dumps a full forensic core snapshot (including raw LOB memory, uncommitted ringbuffers, and recent WAL frames) to `/var/data/growww/matching-engine/quarantine/`, and restarts state re-hydration from a fresh base checkpoint.

### 4. Leader Fencing, Heartbeat Loss & Sub-50ms Failover Workflow
1. **Heartbeat Protocol:** Primary broadcasts a UDP/TCP heartbeat every 5 milliseconds to the Shadow node and maintains an active lease in etcd with a 15-millisecond TTL.
2. **Failure Detection:** If the Shadow node fails to receive 3 consecutive heartbeats (15ms elapsed) and detects etcd lease expiration, it initiates the failover state machine.
3. **Epoch Fencing Token Acquisition:** The Shadow issues an atomic compare-and-swap transaction to etcd to acquire the leadership lease, incrementing the cluster epoch from $N$ to $N + 1$.
4. **State Transition & Match Queue Re-evaluation:**
   - Shadow updates its internal state: `cluster_epoch = N + 1`, `is_active_leader = true`.
   - Shadow scans its local circular sliding window of matches against the last committed sequence confirmed by Kafka consumers.
   - Any matches generated locally during the cutover window that were not yet published to Kafka are immediately drained and published to `engine.matches.v1` with the new Epoch $N + 1$ header.
5. **Ingress Rerouting:** Ingress gateways and Kafka command consumers re-route active command consumption directly to the promoted Shadow node.
6. **Split-Brain Prevention (Fencing):** If the old Primary node recovers or experiences a transient network partition, any attempt to write to its local WAL or emit to Kafka is rejected because its epoch $N$ is less than the active cluster epoch $N + 1$. The old Primary immediately halts its matching loop and enters follower/recovery mode.
7. **Total Failover Time:** Heartbeat timeout (15ms) + Lease acquisition (10ms) + Output drain and activation (15ms) $\le 40\text{ms}$, satisfying the sub-50ms RTO requirement with zero trade loss ($RPO = 0$).

## Step-by-Step Build Instructions (10-15 steps)

1. **Initialize Subsystem Workspace:** Scaffold the Rust crate `services/matching-engine-wal` within the monorepo, configuring `Cargo.toml` dependencies (`memmap2`, `libc`, `zerocopy`, `bytemuck`, `crc32fast`, `sha2`, `tokio`, `tonic`, `prost`, `etcd-client`, `rdkafka`, `parking_lot`, `criterion`).
2. **Define C-ABI Packed Binary Frame Structures:** Author `crates/wal-types` defining `#[repr(C, align(64))]` and `#[repr(C, packed)]` structures for WAL headers, order command payloads, trade execution records, and state checksum attestations with zero-allocation byte conversion traits.
3. **Implement Memory-Mapped Segment Manager (`MmapWalWriter`):** Build the file allocation and mapping engine. Use `libc::posix_fallocate` to pre-allocate 2GB segment files on direct NVMe storage, map segments with `memmap2::MmapMut`, and set kernel memory flags using `libc::madvise` and `libc::mlock`.
4. **Implement Synchronous Flush & Durability Pipeline:** Implement `msync` flush routines supporting configurable sync modes (`MS_SYNC` for absolute durability on every record, or `MS_ASYNC` with periodic microsecond background flush timers) and monotonic 64-bit sequence ID generation.
5. **Implement WAL Segment Indexer & Log Rotator:** Build automated segment rotation logic that detects when the current 2GB segment boundary is approached, pre-allocates the next sequential segment in a background thread, writes an end-of-segment marker, and transitions mmap pointers atomically.
6. **Integrate Pre-Matching WAL Hook in Order Matching Core:** Embed the WAL writer directly into the matching engine's execution path. Ensure incoming orders are committed to the mmap WAL before being applied to the in-memory `BTreeMap` Limit Order Book (Prompt 205).
7. **Implement Hot-Warm Shadow Replica Subsystem:** Build the secondary replica runner. Configure the shadow instance to ingest the identical command stream, maintain an identical in-memory LOB, and route generated match events into a dormant ringbuffer with outbound Kafka publishing suppressed.
8. **Build High-Performance Peer Replication Channel:** Implement an ultra-low-latency peer-to-peer TCP/UDP streaming protocol between Primary and Shadow nodes to transmit WAL frame sequence acknowledgements and 5ms heartbeat packets.
9. **Implement SHA-256 Depth State Checksum Calculator:** Build the deterministic state serializer. When `match_counter % 10000 == 0`, serialize canonical bid/ask price levels and top order hash chains into a stack-allocated buffer and compute the 256-bit SHA-256 digest.
10. **Build Checksum Verification & Automated Quarantine Engine:** Implement the peer checksum comparator. If the shadow's checksum differs from the primary's checksum at the identical match counter, immediately trigger PagerDuty alerts, isolate the shadow replica, and generate a full binary diagnostic dump to `/var/data/growww/matching-engine/quarantine/`.
11. **Implement Distributed Fencing & Lease Controller:** Integrate `etcd-client` to manage cluster leadership leases, 15ms TTL keep-alive heartbeats, and monotonic epoch fencing tokens to prevent split-brain execution across partitions.
12. **Implement Sub-50ms Shadow Promotion State Machine:** Build the atomic failover orchestrator. Upon leader lease loss, acquire leadership epoch $N + 1$, transition `is_active_leader = true`, drain uncommitted matches from the sliding window ringbuffer, enable Kafka event publication, and resume live order matching.
13. **Implement Crash Recovery & Fast Replay Engine:** Build the sequential mmap WAL reader (`MmapWalReader`). On cold start, load the latest local snapshot checkpoint and sequentially replay uncommitted WAL binary frames to reconstitute 1,000,000 orders into the in-memory LOB in $< 500\text{ms}$.
14. **Expose gRPC Control Plane & Prometheus Telemetry:** Implement the `WalAdminService` gRPC server (`GetWalStatus`, `GetReplicationLag`, `VerifyStateChecksum`, `TriggerManualFailover`) and instrument Prometheus metrics for flush latencies, sequence offsets, replication drift, and failover durations.
15. **Execute Chaos, Benchmark & State Determinism Test Suite:** Author Criterion micro-benchmarks and automated chaos test harnesses simulating network partitions, hard process kills (`SIGKILL`), NVMe write stalls, and random order bursts to verify zero lost trades and $< 50\text{ms}$ failover cutover.

## Interfaces / Contracts

### Protobuf Definition (`proto/growww/matching/wal/v1/wal_service.proto`)
```protobuf
syntax = "proto3";

package growww.matching.wal.v1;

option go_package = "growww/matching/wal/v1;walv1";

service WalAdminService {
  rpc GetWalStatus (GetWalStatusRequest) returns (GetWalStatusResponse);
  rpc GetReplicationLag (GetReplicationLagRequest) returns (GetReplicationLagResponse);
  rpc VerifyStateChecksum (VerifyStateChecksumRequest) returns (VerifyStateChecksumResponse);
  rpc TriggerManualFailover (TriggerManualFailoverRequest) returns (TriggerManualFailoverResponse);
  rpc StreamWalFrames (StreamWalFramesRequest) returns (stream WalFrameData);
}

enum NodeRole {
  NODE_ROLE_UNSPECIFIED = 0;
  NODE_ROLE_PRIMARY_LEADER = 1;
  NODE_ROLE_HOT_SHADOW_REPLICA = 2;
  NODE_ROLE_QUARANTINED = 3;
  NODE_ROLE_RECOVERING = 4;
}

enum WalFlushMode {
  WAL_FLUSH_MODE_UNSPECIFIED = 0;
  WAL_FLUSH_MODE_MS_SYNC = 1;
  WAL_FLUSH_MODE_MS_ASYNC_BATCHED = 2;
}

enum ChecksumStatus {
  CHECKSUM_STATUS_UNSPECIFIED = 0;
  CHECKSUM_STATUS_MATCH = 1;
  CHECKSUM_STATUS_MISMATCH_DIVERGENCE = 2;
  CHECKSUM_STATUS_PENDING_PEER = 3;
}

message GetWalStatusRequest {
  string isin = 1;
}

message GetWalStatusResponse {
  string node_id = 1;
  NodeRole role = 2;
  uint64 current_epoch = 3;
  uint64 active_segment_id = 4;
  uint64 last_committed_sequence_id = 5;
  uint64 last_flushed_sequence_id = 6;
  uint64 total_bytes_written = 7;
  WalFlushMode flush_mode = 8;
  int64 uptime_seconds = 9;
  int64 timestamp_unix_ns = 10;
}

message GetReplicationLagRequest {
  string isin = 1;
}

message GetReplicationLagResponse {
  string primary_node_id = 1;
  string shadow_node_id = 2;
  uint64 primary_sequence_id = 3;
  uint64 shadow_sequence_id = 4;
  uint64 sequence_lag_count = 5;
  int64 time_lag_nanoseconds = 6;
  bool is_in_sync = 7;
}

message VerifyStateChecksumRequest {
  string isin = 1;
  uint64 match_sequence_interval = 2; // e.g. 10000
}

message VerifyStateChecksumResponse {
  string isin = 1;
  uint64 match_count = 2;
  uint64 sequence_id = 3;
  string primary_sha256_checksum = 4;
  string shadow_sha256_checksum = 5;
  ChecksumStatus status = 6;
  int64 verified_at_unix_ns = 7;
}

message TriggerManualFailoverRequest {
  string isin = 1;
  string target_shadow_node_id = 2;
  string reason = 3;
}

message TriggerManualFailoverResponse {
  bool success = 1;
  uint64 new_epoch = 2;
  string promoted_node_id = 3;
  int64 cutover_duration_microseconds = 4;
  string message = 5;
}

message StreamWalFramesRequest {
  string isin = 1;
  uint64 start_sequence_id = 2;
  uint32 max_frames = 3;
}

message WalFrameData {
  uint64 epoch = 1;
  uint64 sequence_id = 2;
  uint32 frame_type = 3;
  int64 timestamp_unix_ns = 4;
  bytes raw_payload = 5;
  uint32 crc32c = 6;
}
```

### Binary WAL Record Memory Layout (C-ABI Specification)

```text
================================================================================
C-ABI PACKED BINARY WAL FRAME LAYOUT (64-BYTE / 128-BYTE ALIGNED)
================================================================================

1. COMMON FRAME HEADER (32 Bytes, Fixed):
   Offset 0x00 - 0x03 (4 bytes) : Magic Identifier = 0x57414C31 ('W', 'A', 'L', '1')
   Offset 0x04 - 0x07 (4 bytes) : Frame Version (uint32, e.g. 1)
   Offset 0x08 - 0x0F (8 bytes) : Leadership Epoch Fencing Token (uint64, big-endian)
   Offset 0x10 - 0x17 (8 bytes) : Monotonic Sequence Identifier (uint64, big-endian)
   Offset 0x18 - 0x1F (8 bytes) : Timestamp Nanoseconds (int64, Unix epoch)

2. FRAME METADATA (8 Bytes, Fixed):
   Offset 0x20 - 0x21 (2 bytes) : Frame Type Enumeration (uint16)
                                  0x0001 = ORDER_NEW
                                  0x0002 = ORDER_CANCEL
                                  0x0003 = ORDER_REPLACE
                                  0x0004 = ORDER_MASS_CANCEL
                                  0x0005 = TRADE_MATCH
                                  0x0006 = CHECKPOINT_START
                                  0x0007 = CHECKPOINT_END
                                  0x0008 = CHECKSUM_ATTESTATION
   Offset 0x22 - 0x23 (2 bytes) : Payload Byte Length (uint16)
   Offset 0x24 - 0x27 (4 bytes) : Reserved Padding (zeros)

3. PAYLOAD DATA (Variable / Fixed 88 Bytes for 128-Byte Frame):
   For ORDER_NEW (Frame Type 0x0001, 88 Bytes):
   Offset 0x28 - 0x37 (16 bytes) : Order ID (UUID, 128-bit big-endian)
   Offset 0x38 - 0x47 (16 bytes) : User ID (UUID, 128-bit big-endian)
   Offset 0x48 - 0x53 (12 bytes) : ISIN (ASCII String, 12 bytes, e.g. "INE002A01018")
   Offset 0x54 - 0x54 (1 byte)   : Side (uint8: 1 = BUY, 2 = SELL)
   Offset 0x55 - 0x55 (1 byte)   : Order Type (uint8: 1 = LIMIT, 2 = MARKET, 3 = IOC, 4 = FOK)
   Offset 0x56 - 0x57 (2 bytes)  : Time In Force (uint16)
   Offset 0x58 - 0x5F (8 bytes)  : Limit Price (uint64, fixed-point 10^-4 INR)
   Offset 0x60 - 0x67 (8 bytes)  : Quantity (uint64, fixed-point 10^-6 shares)
   Offset 0x68 - 0x77 (16 bytes) : Client Order Reference ID (16 bytes ASCII / UUID)
   Offset 0x78 - 0x7B (4 bytes)  : STP Policy Flag (uint32)

4. FRAME INTEGRITY FOOTER (4 Bytes, Fixed):
   Offset 0x7C - 0x7F (4 bytes) : Hardware CRC32C Checksum (uint32)
                                  Computed over bytes 0x00 through 0x7B.

Total Frame Size: 128 Bytes (Exact cache-line multiple, zero struct tearing).
```

### PostgreSQL Database Schema DDL (`wal_failover_metadata`)

```sql
-- WAL Segment Registry Table
CREATE TABLE wal_segment_registry (
    segment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    segment_sequence_number BIGINT NOT NULL,
    file_path VARCHAR(255) NOT NULL,
    start_sequence_id BIGINT NOT NULL,
    end_sequence_id BIGINT,
    file_size_bytes BIGINT NOT NULL DEFAULT 2147483648, -- 2GB
    is_closed BOOLEAN NOT NULL DEFAULT FALSE,
    epoch_token BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMPTZ,
    CONSTRAINT uq_node_segment UNIQUE (node_id, isin, segment_sequence_number)
);

-- SHA-256 Depth State Checksum Log Table (Every 10,000 Matches)
CREATE TABLE depth_checksum_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    match_counter BIGINT NOT NULL,
    sequence_id BIGINT NOT NULL,
    primary_node_id VARCHAR(64) NOT NULL,
    shadow_node_id VARCHAR(64) NOT NULL,
    primary_sha256 VARCHAR(64) NOT NULL,
    shadow_sha256 VARCHAR(64) NOT NULL,
    checksum_matched BOOLEAN NOT NULL,
    bid_level_count INT NOT NULL,
    ask_level_count INT NOT NULL,
    total_resting_volume_shares NUMERIC(28, 6) NOT NULL,
    verified_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- High-Availability Failover Incident Audit Table
CREATE TABLE failover_audit_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    previous_leader_node_id VARCHAR(64) NOT NULL,
    promoted_shadow_node_id VARCHAR(64) NOT NULL,
    previous_epoch BIGINT NOT NULL,
    new_epoch BIGINT NOT NULL,
    failover_trigger_reason VARCHAR(100) NOT NULL, -- HEARTBEAT_TIMEOUT, MANUAL_OPERATOR, CHECKSUM_QUARANTINE
    uncommitted_matches_drained INT NOT NULL DEFAULT 0,
    cutover_latency_microseconds BIGINT NOT NULL,
    failover_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indices for performance and audit reconciliation
CREATE INDEX idx_wal_segments_isin_seq ON wal_segment_registry(isin, segment_sequence_number);
CREATE INDEX idx_checksum_isin_match ON depth_checksum_logs(isin, match_counter DESC);
CREATE INDEX idx_checksum_mismatch ON depth_checksum_logs(checksum_matched) WHERE checksum_matched = FALSE;
CREATE INDEX idx_failover_timestamp ON failover_audit_events(failover_timestamp DESC);
```

### Kafka Event Schemas

#### Topic: `matching.wal.checksum.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "DepthStateChecksumEvent",
  "type": "object",
  "required": [
    "event_id",
    "isin",
    "match_counter",
    "sequence_id",
    "epoch_token",
    "primary_node_id",
    "primary_sha256",
    "top_bid_price",
    "top_ask_price",
    "total_bid_volume",
    "total_ask_volume",
    "timestamp_unix_ns"
  ],
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "isin": { "type": "string", "minLength": 12, "maxLength": 12 },
    "match_counter": { "type": "integer", "minimum": 10000 },
    "sequence_id": { "type": "integer", "minimum": 1 },
    "epoch_token": { "type": "integer", "minimum": 1 },
    "primary_node_id": { "type": "string" },
    "primary_sha256": { "type": "string", "pattern": "^[0-9a-fA-F]{64}$" },
    "top_bid_price": { "type": "string" },
    "top_ask_price": { "type": "string" },
    "total_bid_volume": { "type": "string" },
    "total_ask_volume": { "type": "string" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

#### Topic: `engine.failover.events.v1`
```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "EngineFailoverEvent",
  "type": "object",
  "required": [
    "event_id",
    "isin",
    "previous_primary_id",
    "promoted_leader_id",
    "previous_epoch",
    "new_epoch",
    "trigger_reason",
    "last_processed_sequence_id",
    "drained_match_count",
    "cutover_duration_microseconds",
    "timestamp_unix_ns"
  ],
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "isin": { "type": "string", "minLength": 12, "maxLength": 12 },
    "previous_primary_id": { "type": "string" },
    "promoted_leader_id": { "type": "string" },
    "previous_epoch": { "type": "integer" },
    "new_epoch": { "type": "integer" },
    "trigger_reason": {
      "type": "string",
      "enum": ["HEARTBEAT_TIMEOUT", "LEASE_EXPIRATION", "MANUAL_OPERATOR", "PRIMARY_PANIC", "CHECKSUM_QUARANTINE"]
    },
    "last_processed_sequence_id": { "type": "integer" },
    "drained_match_count": { "type": "integer" },
    "cutover_duration_microseconds": { "type": "integer" },
    "timestamp_unix_ns": { "type": "integer" }
  }
}
```

## Security & Compliance Notes
- **Fencing Token Invariant & Split-Brain Elimination:** Monotonically increasing 64-bit epoch tokens acquired from etcd ensure that a partitioned or resurrected zombie primary node cannot corrupt the mmap WAL or emit duplicate execution events to Kafka. All downstream consumers reject messages tagged with an epoch lower than the active cluster epoch.
- **Strict Durability Guarantee ($RPO = 0$):** In `MS_SYNC` mode, memory-mapped pages are flushed to non-volatile physical NVMe storage before order match acknowledgements are emitted. Power loss or kernel panic cannot result in committed order loss.
- **Deterministic State Divergence Quarantine:** The periodic SHA-256 depth state checksum computed every 10,000 matches mathematically verifies that primary and shadow limit order books are bit-for-bit identical. Any divergence immediately isolates the shadow replica, preventing corrupted replicas from assuming leadership during failover.
- **Zero Swap Paging & Memory Safety:** Kernel memory locking via `libc::mlock` and `libc::madvise(..., MADV_DONTDUMP)` prevents sensitive matching memory and WAL data from being paged to disk swap or leaked into untrusted diagnostic core dumps.
- **Zero On-Chain PII Compliance:** All WAL binary frames, Kafka messages, and blockchain attestations operate strictly on pseudonymized UUIDs, numerical sequence identifiers, and fixed-point quantities, ensuring complete compliance with the Indian Digital Personal Data Protection (DPDP) Act 2023 and EU GDPR.

## Acceptance Criteria
- [ ] Memory-mapped WAL append operation achieves a $p99$ latency of $\le 5\mu\text{s}$ using zero-copy binary framing on direct NVMe storage.
- [ ] Direct kernel page flush via `msync` with `MS_SYNC` guarantees physical persistence without intermediate buffer copying or heap allocation.
- [ ] Monotonic 64-bit sequence numbers are strictly contiguous and gap-free across 10,000,000 sequential orders under heavy write concurrency.
- [ ] Hot-warm shadow replica maintains an in-memory Limit Order Book identical to the primary node with 0 outbound trade events emitted to Kafka while in shadow mode.
- [ ] Cryptographic SHA-256 depth state checksums are calculated and verified between Primary and Shadow nodes precisely every 10,000 matches.
- [ ] Simulated state divergence (injected single-order discrepancy) causes the shadow replica to quarantine itself, emit critical alerts, and generate a memory snapshot in $< 10\text{ms}$.
- [ ] Automated failover cutover from shadow replica to active primary completes in $< 50\text{ms}$ upon simulated primary failure ($RTO < 50\text{ms}$, $RPO = 0$).
- [ ] Epoch fencing tokens strictly prevent partitioned primary nodes from appending to WAL or broadcasting matches to Kafka after a cutover.
- [ ] Sequential cold-start recovery via `MmapWalReader` replays 1,000,000 orders and reconstitutes full Limit Order Book state in $< 500\text{ms}$.
- [ ] Protobuf service `WalAdminService` correctly reports real-time segment offsets, replication lag, and checksum logs via gRPC.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero application implementation code and zero em dashes or en dashes.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `101` (System Architecture Overview), Prompt `104` (Event Schema & Kafka Topic Standards), Prompt `205` (High-Performance Order Matching Engine), Prompt `401` (PostgreSQL Database Schema & Migration Engine).
- **Parallel Tasks:** Prompt `206` (Pre-Trade Risk & Margin Checks Service), Prompt `225` (FIX 5.0 SP2 / ITCH / OUCH Binary Trading Gateway), Prompt `227` (24/7 Institutional Dark Pool & Block Crossing Service).
- **Downstream Blockers:** Prompt `208` (Trade Settlement & DvP Orchestration Service), Prompt `215` (On-Chain vs Off-Chain Ledger Reconciliation Engine), Prompt `306` (DvP Settlement Smart Contract).
