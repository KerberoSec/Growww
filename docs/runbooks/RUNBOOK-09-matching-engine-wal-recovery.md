# Runbook 09: Matching Engine Cold-Start & WAL Snapshot Recovery

**Runbook ID:** RBK-OPS-009  
**Severity Tier:** P1 (Critical)  
**Authority:** High-Frequency Trading Systems Lead / SRE On-Call  

---

## 1. Description & Trigger Conditions
Triggered when a matching engine partition pod crashes, restarts, or fails over to its hot-standby replica.

---

## 2. Recovery Workflow

```
[Crash Detected / New Pod Provisioned]
                  |
                  v
[1. Load Memory Snapshot from Disk]
  - Locate latest /var/lib/growww/snapshots/orderbook_<isin>_<timestamp>.bin
  - mmap binary image into address space (MS_SYNC)
                  |
                  v
[2. Query Snapshot High-Water Offset]
  - Extract snapshot_last_kafka_offset from binary header
                  |
                  v
[3. Replay growww.order.events.v1]
  - Seek Kafka consumer to snapshot_last_kafka_offset + 1
  - Replay events to rebuild active order book in memory
                  |
                  v
[4. Validate Cryptographic State Root]
  - Assert in-memory order book Merkle root == audit-service current root
                  |
                  v
[5. Re-Open Gateway Traffic]
  - Notify API Gateway to resume order ingestion for that instrument
```

Target recovery latency: **$< 2.0\text{ seconds}$** for 1,000,000 active limit orders.
