# Settlement Orchestration & Dynamic Fee Settlement Architecture

## 1. Architecture Overview
The `trade-settlement-service` orchestrates high-throughput clearing between the microsecond Rust order matching engine, TigerBeetle double-entry balance ledgers, and Hyperledger Besu.

## 2. Dynamic Fee Parameter & Zero-Fee Launch
- **Initial Launch Policy**: Operates with strictly **0.00% fees everywhere ("0 means 0 in all")**.
- **Dynamic Fee Controller Integration**:
  - Subscribes to `FeeController.sol` events via WebSocket on Besu.
  - Caches current `makerFeeBps` and `takerFeeBps` in memory (default `0`).
  - If governance executes a fee increase in the future, the settlement service updates its netting and balance debit formulas seamlessly without downtime.

## 3. Three-Tier High-Throughput Settlement Pipeline
1. **Tier 1 (Instantaneous Match & Ledger Hold)**:
   - Matching engine emits execution reports to Kafka topic `growww.spot.trades.v1`.
   - TigerBeetle executes two-phase pending transfers (`flags.pending = true`), locking assets in sub-millisecond real-time with dual-entry accounting.
2. **Tier 2 (Continuous Multilateral Netting Engine)**:
   - Aggregates trades over a 10-second netting epoch.
   - Bilateral high-frequency trades are compressed into net balance changes, achieving a 99.8% transaction volume reduction.
3. **Tier 3 (Batch Settlement Submission on Besu)**:
   - Settlement relayer formats netted balances into a single batch calling `DvPAtomicSettlement.settleNetBatch()`.
   - Upon QBFT block finality, TigerBeetle commits pending transfers (`flags.post_pending_transfer = true`).

## 4. Nonce Management & Relayer Sharding
- **32-Worker Relayer Pool**: Distinct relayer accounts (`0xRelayer00` through `0xRelayer31`).
- **Deterministic Sharding**: Trades routed via `Murmur3(TradingPair) % 32`.
- **In-Memory Monotonic Nonce Tracker**: Each relayer maintains an atomic sequence counter synchronized via Redis Lua locks.
- **Stuck Transaction Recovery**: If unmined after 2 blocks (4 seconds), transaction is rebroadcast with identical nonce and a 1.25x (+25%) gas bump (`maxPriorityFeePerGas`).

## 5. Continuous 3-Way Reconciliation Engine
Every 60 seconds, an automated daemon verifies:
`TigerBeetle Ledger Sum == Besu Vault Balance == Kafka Outbox Offset`
