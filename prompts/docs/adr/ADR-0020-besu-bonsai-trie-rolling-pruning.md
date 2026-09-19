# ADR-0020: Hyperledger Besu Bonsai Trie Storage and Rolling Pruning

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Lead Blockchain SRE, Infrastructure Lead  

---

## 1. Context
At 21.6 million transactions per day, unpruned blockchain state accumulates 15-25 GB of RocksDB disk space daily, causing validator storage exhaustion within 90 days.

---

## 2. Decision
Deploy Hyperledger Besu with the **Bonsai Trie Storage Format (`--data-storage-format=BONSAI`)** and enable **Rolling Trie Pruning**:
- Retains working trie state for the last 2,048 blocks (approx. 68 minutes of reorg buffer), capping local NVMe disk usage under 450 GB.
- Completed block receipts and execution logs are archived asynchronously to an AWS S3 Glacier WORM storage vault for 8-year statutory compliance.

---

## 3. Consequences
- **Positive:** Caps validator node disk footprint; prevents RocksDB compaction latency spikes.
- **Trade-offs:** Historical RPC state calls older than 2,048 blocks must query archival indexers.
