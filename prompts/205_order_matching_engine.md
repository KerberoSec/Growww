# 205 - High-Performance Order Matching Engine (Rust)

## Purpose
The Order Matching Engine is the ultra-low-latency computational core of the Growww trading exchange. It executes deterministic, in-memory limit order book matching based on strict Price-Time Priority (FIFO) rules for Indian equities. Unlike traditional equity exchanges dealing only in whole integer shares, Growww's matching engine natively supports fractional equity units (down to 6 decimal places, e.g., $0.000001$ shares) while maintaining microsecond-level matching latencies and zero floating-point arithmetic errors.

By delivering deterministic execution and producing cryptographic match receipts that feed the on-chain Delivery-versus-Payment (DvP) settlement pipeline, the matching engine enables continuous liquidity and seamless fractional investing in 1:1 asset-backed Indian securities.

## What You Are Building
A standalone, bare-metal optimized Rust service (`services/matching-engine`). Concrete deliverables include:
- In-memory Limit Order Book (LOB) implementation per ISIN utilizing `BTreeMap` for sorted price levels and intrusive doubly-linked lists for FIFO queue management.
- Fixed-point 128-bit integer financial math engine ($10^6$ multiplier for fractional shares, $10^4$ multiplier for INR paise/sub-paise).
- High-throughput asynchronous command ingestion pipeline consuming from Kafka (`order.matching.commands.v1`) and zero-copy gRPC.
- Self-Trade Prevention (STP) engine to cancel or reject crossing orders originating from the same beneficial owner.
- High-speed trade match and market depth event broadcaster streaming to Kafka (`engine.matches.v1`, `matching.depth.v1`).
- RocksDB local state snapshotting and Write-Ahead Log (WAL) replay engine for $< 2$-second cold-start recovery.

## Scope Boundaries
- **In Scope:**
 - In-memory order book matching (Limit, Market, IOC, FOK orders).
 - Price-time priority matching with partial fill allocation.
 - Fractional equity volume matching (fixed-point arithmetic).
 - Order cancellation and replacement processing.
 - Level-2 (top 20 bid/ask price levels) market depth snapshot generation.
 - Deterministic state snapshotting to local NVMe storage.
- **Out of Scope / Handled Elsewhere:**
 - User authentication and session management (Prompt 201).
 - Pre-trade wallet balance holds and database storage (Prompt 203, 204).
 - Pre-trade risk policy limit evaluation (Prompt 206).
 - Client WebSocket fan-out broadcasting (Prompt 207).
 - On-chain smart contract transaction submission (Prompt 208).

## Technology to Use
- **Primary Language & Toolchain:** Rust 2021 Edition (stable 1.78+). Rust is selected over Go/Java for this component to guarantee zero garbage collection (GC) latency spikes, deterministic memory layout, thread-level data race safety, and sub-10-microsecond matching latencies.
- **Async Runtime & Concurrency:** `tokio` (multi-threaded async runtime), `crossbeam-channel` for lock-free intra-thread message passing, `parking_lot` for fast primitives.
- **Data Structures & Fixed Math:** `std::collections::BTreeMap` for price levels; custom intrusive doubly-linked list for cache-friendly $O(1)$ order queue insertions and removals; fixed-point integer representations (no `f32` / `f64`).
- **Persistence & Snapshots:** `rocksdb` Rust bindings for high-speed local NVMe snapshot persistence.
- **Messaging & Serialization:** `rdkafka` (librdkafka C bindings for maximum throughput) and `prost` / `tonic` for Protobuf/gRPC.

## Backend / Infra Touchpoints
- **Apache Kafka:** Consumes commands from `order.matching.commands.v1` (partitioned by ISIN); publishes matches to `engine.matches.v1` and L2 depth diffs to `matching.depth.v1`.
- **Local NVMe SSD:** RocksDB directory `/var/data/growww/matching-engine/snapshots/` for periodic order book snapshots (every 10,000 events).
- **Trade Settlement Service (Prompt 208):** Ingests match events from `engine.matches.v1` to orchestrate DvP settlements.
- **Market Data Service (Prompt 207):** Ingests L2 depth diffs to stream live order books to clients.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Deterministic Match Receipts:** For every executed match, the engine produces a cryptographically signed execution receipt containing `match_id`, `buyer_address`, `seller_address`, `isin`, `token_amount_raw` (integer micro-tokens matching ERC-3643 fractional decimals), `trade_price_paise`, and `sequence_number`.
- **On-Chain DvP Feed:** These match receipts are ingested by the Trade Settlement Service (Prompt 208), which submits atomic settlement transactions to `SettlementDvP.sol` on Hyperledger Besu under QBFT consensus.
- **Zero On-Chain PII:** The matching engine operates strictly on pseudonymized `user_id` and public ledger addresses (`0x...`), ensuring zero PII enters the matching stream or ledger.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Rust Project:** Initialize Cargo workspace `services/matching-engine` with `Cargo.toml`, release profiles optimized for performance (`opt-level = 3`, `lto = "fat"`, `codegen-units = 1`, `panic = "abort"`), and strict `clippy` linter settings.
2. **Define Fixed-Point Data Types:** Implement `Price` (u64 representing hundredths of a paisa, $10^{-4}$ INR) and `Quantity` (u64 representing micro-shares, $10^{-6}$ units). Implement checked arithmetic with overflow protection.
3. **Build Core Order Book Data Structures:** Implement `OrderQueue` (FIFO intrusive linked list of orders) and `PriceLevel` structs.
4. **Implement LimitOrderBook:** Build `LimitOrderBook` struct containing `bids: BTreeMap<Price, OrderQueue, Reverse>` and `asks: BTreeMap<Price, OrderQueue>`.
5. **Implement Price-Time Matching Logic:** Write `match_limit_order` function:
 - For Buy Limit: traverse `asks` where `ask_price <= limit_price`. Match available volume at `ask_price`.
 - For Sell Limit: traverse `bids` where `bid_price >= limit_price`. Match available volume at `bid_price`.
 - Support partial fills, generating `Match` events for each matched pair.
6. **Implement Market & Special Order Matching:** Build matching handlers for Market, IOC (Immediate-or-Cancel - cancel unmatched remainder), and FOK (Fill-or-Kill - verify full liquidity exists before matching, else abort).
7. **Implement Self-Trade Prevention (STP):** Add check: if incoming order `user_id == maker_order.user_id`, apply configured STP policy (`CANCEL_INCOMING`, `CANCEL_RESTING`, or `DECREMENT_AND_CANCEL`).
8. **Build Level-2 Market Depth Generator:** Implement snapshot extractor returning top $N$ bid and ask price levels with aggregated quantities.
9. **Implement Snapshotting & RocksDB Persistence:** Implement background worker serializing order book state using `bincode` and writing checkpoints to RocksDB with sequence IDs.
10. **Implement WAL Replay & Cold-Start Recovery:** Build startup recovery procedure: load latest RocksDB checkpoint, then replay Kafka messages from the checkpoint sequence offset to rebuild exact state.
11. **Implement Kafka Command Consumer:** Build single-threaded or per-ISIN pinned thread consumer using `rdkafka` with pre-allocated buffers.
12. **Implement Kafka Match & Depth Producer:** Stream `MatchEvent` and `DepthUpdate` to Kafka with zero-copy serialization.
13. **Implement gRPC Administration & Metrics API:** Expose `tonic` gRPC server for administrative state queries, dynamic symbol halting, and health checks.
14. **Instrument Micro-Benchmarks & Profiling:** Write Criterion micro-benchmarks (`benches/matching_benchmark.rs`) verifying matching latency $\le 10\mu\text{s}$ per order.
15. **Execute Chaos & Stress Tests:** Run high-throughput stress simulation (100,000 orders/sec) verifying state determinism, zero memory leaks, and 100% match reproducibility across replay runs.

## Interfaces / Contracts

### Protobuf Definition (`matching_engine.proto`)
```protobuf
syntax = "proto3";

package growww.matching.v1;

option go_package = "growww/matching/v1;matchingv1";

service MatchingEngineAdmin {
  rpc HaltSymbol (HaltSymbolRequest) returns (HaltSymbolResponse);
  rpc ResumeSymbol (ResumeSymbolRequest) returns (ResumeSymbolResponse);
  rpc GetL2Snapshot (GetL2SnapshotRequest) returns (GetL2SnapshotResponse);
}

message HaltSymbolRequest {
  string isin = 1;
  string reason = 2;
}

message HaltSymbolResponse {
  bool success = 1;
  int64 halted_at_unix_ns = 2;
}

message ResumeSymbolRequest {
  string isin = 1;
}

message ResumeSymbolResponse {
  bool success = 1;
  int64 resumed_at_unix_ns = 2;
}

message GetL2SnapshotRequest {
  string isin = 1;
  int32 depth_levels = 2; // Default 20
}

message PriceLevelDto {
  uint64 price_raw = 1; // Price in 10^-4 INR
  uint64 quantity_raw = 2; // Quantity in 10^-6 shares
  int32 order_count = 3;
}

message GetL2SnapshotResponse {
  string isin = 1;
  int64 sequence_number = 2;
  repeated PriceLevelDto bids = 3;
  repeated PriceLevelDto asks = 4;
  int64 timestamp_ns = 5;
}

message OrderMatchEvent {
  string match_id = 1;
  string isin = 2;
  string maker_order_id = 3;
  string taker_order_id = 4;
  string maker_user_id = 5;
  string taker_user_id = 6;
  string maker_ledger_address = 7;
  string taker_ledger_address = 8;
  uint64 matched_price_raw = 9;
  uint64 matched_quantity_raw = 10;
  bool is_maker_buy = 11;
  int64 sequence_number = 12;
  int64 matched_at_unix_ns = 13;
}
```

### Core Rust In-Memory Layout (Conceptual Reference)
```rust
#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub struct Price(pub u64); // Price in 10^-4 INR (0.01 paisa)

#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Quantity(pub u64); // Quantity in 10^-6 units (micro-shares)

#[derive(Debug, Clone)]
pub struct Order {
    pub order_id: u128,
    pub user_id: u128,
    pub ledger_address: [u8; 20],
    pub isin: [u8; 12],
    pub price: Price,
    pub quantity: Quantity,
    pub remaining_qty: Quantity,
    pub side: Side,
    pub timestamp_ns: u64,
}

pub struct LimitOrderBook {
    pub isin: [u8; 12],
    pub bids: std::collections::BTreeMap<std::cmp::Reverse<Price>, std::collections::VecDeque<Order>>,
    pub asks: std::collections::BTreeMap<Price, std::collections::VecDeque<Order>>,
    pub sequence_number: u64,
}
```

## Security & Compliance Notes
- **Deterministic Replay Guarantee:** Order matching is 100% single-threaded or partition-pinned; replaying the exact Kafka log from snapshot zero produces the identical sequence of trades and balances.
- **SEBI Market Integrity & STP:** Strict Self-Trade Prevention prevents artificial volume inflation or wash trading.
- **Integer Overflow Protection:** All arithmetic operations use Rust's `checked_add`, `checked_sub`, and `checked_mul` or saturation arithmetic to eliminate integer wrapping exploits.
- **Zero Heap Allocations on Hot Path:** Core matching loops utilize pre-allocated slab allocators and object pools to eliminate runtime memory allocation latency.

## Acceptance Criteria
- [ ] Matching engine compiles cleanly with zero warnings under `cargo clippy -- -D warnings`.
- [ ] In-memory limit order book correctly matches limit, market, IOC, and FOK orders according to Price-Time Priority.
- [ ] Fractional equity quantities (down to $10^{-6}$ units) match accurately with zero rounding discrepancies.
- [ ] Self-Trade Prevention blocks matching when buyer and seller share the same user ID.
- [ ] Single-order matching latency on $p99$ is $< 25\mu\text{s}$ under 50,000 orders/sec throughput.
- [ ] Snapshot and WAL replay restores identical order book state and sequence number upon process crash simulation.
- [ ] Kafka match events are delivered with exactly-once semantics to downstream settlement pipelines.

## Suggested Order / Dependencies
- **Prerequisites:** 010 (NFR Targets), 103 (API Standards), 104 (Kafka Standards), 204 (Order Service), 407 (Master Data Management).
- **Parallel Tasks:** 206 (Risk Engine), 207 (Market Data Service).
- **Downstream Blockers:** 208 (Trade Settlement Service), 903 (Load Testing Matching Engine).
