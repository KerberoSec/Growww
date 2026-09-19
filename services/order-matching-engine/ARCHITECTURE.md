# Matching Engine Internals & Dynamic Parameter Specification

## 1. Architecture Overview
The order matching engine is implemented in Rust 1.76, operating on a single-writer thread pinned to a dedicated CPU core via NUMA-aware thread affinity. It processes orders with strict Price-Time Priority (FIFO).

## 2. Dynamic Fee Configuration Engine
- **Starting Fee Policy**: Initialized to strictly **0.00% fee** across all orders (`maker_fee_bps = 0`, `taker_fee_bps = 0`).
- **Hot-Reloadable Fee Architecture**:
  - Engine maintains atomic fee parameters:
    `static MAKER_FEE_BPS: AtomicU16 = AtomicU16::new(0);`
    `static TAKER_FEE_BPS: AtomicU16 = AtomicU16::new(0);`
  - Listens to Kafka configuration topic `growww.governance.fee_updates.v1`.
  - Upon authorized governance update, fees can be hot-reloaded dynamically in sub-microseconds without restarting the matching engine or disrupting order execution.

## 3. Fixed-Point Decimal Arithmetic
- **Price Precision**: 6 decimal places (USDT quote: `1.000000 USDT = 1_000_000`).
- **Quantity Precision**: 8 decimal places (BTC base: `1.00000000 BTC = 100_000_000`).
- **Quote Amount**:
  `QuoteAmount = ((Quantity as u128 * Price as u128) / 100_000_000) as u64` (checked integer arithmetic).

## 4. Market Protection & Risk Controls
- **Price Collar / Banding**: Orders outside +/- 10% of last trade price (or 5-minute TWAP) are rejected immediately.
- **Self-Trade Prevention (STP)**: Supports `CancelNewest`, `CancelOldest`, and `DecrementAndCancel` modes to eliminate wash trades.
- **Rate Limiting**: 5,000 order operations/sec per trader account.

## 5. Latency Budgets
- p50: 1.0 microsecond
- p95: 2.5 microseconds
- p99: 5.0 microseconds
