# Database Architecture, CDC Debezium & Continuous Reconciliation

**Specification ID:** SPEC-ARCH-013-DB  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Financial Persistence & Distributed Ledger Synchronization  
**Owner:** Core Data Platform & Database Reliability Group  

---

## 1. Executive Summary & Polyglot Storage Architecture
The platform enforces a purpose-built polyglot database topology separating instantaneous financial transactions, relational user state, time-series market history, and immutable append-only logs:
- **TigerBeetle Financial Ledger**: Single source of truth for trading balances and double-entry holds. Operates in sub-millisecond memory with strict two-phase commit (2PC) serializability (`Ledger ID 1` for Real Spot, `Ledger ID 2` for Demo).
- **PostgreSQL 16 (Transactional Relational Core)**: Manages non-balance transactional state (user profiles, orders, compliance claims, KYC status) with Debezium Change Data Capture (CDC).
- **Apache Kafka (Event Backbone)**: Event streaming bus with 32 partitions per topic ingesting trade fills and order lifecycle events.
- **TimescaleDB & ClickHouse**: High-throughput time-series engines aggregating tick streams into 1s, 1m, 5m, 1h, and 1d OHLCV candlestick bars.
- **ScyllaDB**: High-velocity distributed NoSQL store persisting immutable audit trails and user activity sessions.

---

## 2. Change Data Capture (CDC) Pipeline & Outbox Pattern

```
+----------------------------------------------------------------------------------------------------+
| TRANSACTIONAL OUTBOX CDC & ASYNCHRONOUS PROJECTION PIPELINE                                        |
|                                                                                                    |
|  [ Microservice Transaction ] ---> [ PostgreSQL ACID Commit ]                                      |
|                                    - Business Entity Table Updates                                 |
|                                    - Append to outbox_events Table                                 |
|                                                |                                                   |
|                                                v                                                   |
|                              [ Debezium PostgreSQL CDC Connector ]                                |
|                              - Reads Postgres Write-Ahead Log (WAL)                                |
|                              - Guarantees Exactly-Once At-Least-Once Delivery                      |
|                                                |                                                   |
|                                                v                                                   |
|                                [ Apache Kafka Event Stream ]                                       |
|                                - Partition Key: user_id / symbol                                   |
|                                                |                                                   |
|                         +----------------------+----------------------+                            |
|                         |                                             |                            |
|                         v                                             v                            |
|             [ TimescaleDB Candle Consumer ]               [ Read-Model Projection Worker ]         |
|             - Generates OHLCV rollups                     - Updates Redis Read Cache               |
+----------------------------------------------------------------------------------------------------+
```

---

## 3. Continuous 3-Way Financial Reconciliation Engine
Every 60 seconds, an automated reconciliation daemon executes an exhaustive balance check:
$$\Delta = |\sum \text{TigerBeetleBalances} - \sum \text{BesuVaultReserves}|$$
1. If $\Delta == 0$: State verified; reconciliation checkpoint anchored to ScyllaDB audit log.
2. If $\Delta > 0$: Circuit breaker triggers instantly:
   - Automated withdrawals halted for the divergent asset.
   - SRE On-Call and Financial Operations alerted with exact account divergence diffs.
   - TigerBeetle and Besu transactions quarantined for manual review.
