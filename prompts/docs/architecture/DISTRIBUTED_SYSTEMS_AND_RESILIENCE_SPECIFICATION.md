# Distributed Systems, Scalability & Resilience Specification

**Specification ID:** SPEC-ARCH-011-RES  
**Document Version:** 2.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Cloud Infrastructure & Reliability Engineering Group  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  

---

## 1. Hyperledger Besu Bonsai Trie & Rolling Pruning Topology (ADR-0020)

At 21.6 million transactions per day (500 settlements/block at 2.0s block time), standard archival blockchain nodes accumulate $> 20\text{ GB}$ of RocksDB state daily.

```
+----------------------------------------------------------------------------------------------------+
| BESU BONSAI TRIE STORAGE & ROLLING STATE PRUNING ARCHITECTURE                                     |
|                                                                                                    |
|  [Hyperledger Besu Validator Node (Hot NVMe Disk)]                                                 |
|  - Storage Engine: Bonsai Trie (--data-storage-format=BONSAI)                                      |
|  - Active Working State: Capped at < 450 GB                                                       |
|  - Rolling Trie Pruning: Retains last 2,048 block states (approx. 68 minutes of reorg buffer)      |
|           |                                                                                        |
|           v (Every 24 Hours / 43,200 Blocks)                                                       |
|  [Audit Event Exporter Daemon]                                                                     |
|  - Streams finalized block receipts and trade logs to immutable AWS S3 Glacier WORM Vault          |
|  - Enforces 8-year statutory SEBI regulatory data retention                                        |
+----------------------------------------------------------------------------------------------------+
```

---

## 2. High-Frequency PostgreSQL WAL & CDC Outbox Partitioning

To sustain 30,000 database row inserts per second without disk IOPS saturation:

### 2.1 Declarative Hourly Outbox Partitioning
```sql
CREATE TABLE outbox_events (
    id              BIGSERIAL,
    aggregate_type  TEXT        NOT NULL,
    aggregate_id    TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    payload         JSONB       NOT NULL,
    headers         JSONB       NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at    TIMESTAMPTZ,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Completed hourly partitions older than 72 hours are dropped via DROP TABLE, eliminating VACUUM bloat.
```

### 2.2 PostgreSQL Group Commit & Storage Configuration
- **Storage Subsystem:** AWS io2 Block Express NVMe SSDs configured for 64,000 IOPS.
- **WAL Flush Policy:** `synchronous_commit = off` on bulk ingestion streams with `commit_delay = 5000` (5ms flush ceiling), batching hundreds of transaction commits into a single disk sync.

---

## 3. Kafka Partition Key Hotspot Mitigation (Salted Partitioning)

Under extreme single-asset volume spikes (e.g. index rebalancing in `RELIANCE`), standard partitioning by `instrument_id` can overload a single Kafka broker partition:

```rust
fn calculate_partition_key(instrument_id: &str, account_id: &str, is_high_volume: bool) -> String {
    if is_high_volume {
        let salt = Murmur3_32(account_id.as_bytes()) % 8;
        format!("{}_s{}", instrument_id, salt)
    } else {
        instrument_id.to_string()
    }
}
```

This distributes high-volume single-asset order streams across 8 parallel sub-partitions while preserving FIFO ordering per user account.

---

## 4. WebSocket Fanout & Connection Storm Mitigation (1,000,000 Clients)

To handle market-open connection storms (09:15 IST) and prevent buffer exhaustion:
1. **Edge Termination:** WebSockets terminate on distributed Envoy/Nginx edge gateways rather than internal microservice pods.
2. **Subscription Deltas & Conflation:** If a mobile client connection throttles, the gateway conflates Level 2 updates using a 100ms ticker coalescing window, reducing broadcast bandwidth by $75\%$ without dropping price discovery integrity.
3. **Connection Jitter:** Reconnecting clients enforce randomized exponential backoff ($T = \min(30\text{s}, 2^{\text{attempt}} \times 500\text{ms} + \text{jitter})$) to prevent thundering herd crashes.
