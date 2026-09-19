# Runbook 14: Hyperledger Besu Bonsai Trie Storage Maintenance & Pruning

**Runbook ID:** RBK-OPS-014  
**Severity Tier:** P2 (High)  
**Authority:** Blockchain Infrastructure Lead / SRE On-Call  

---

## 1. Description & Alert Triggers
- PromQL Alert: `BesuDiskUsagePercent > 85%` on validator node NVMe volume.
- RocksDB compaction latency $> 250\text{ms}$.

---

## 2. Remediation Workflow
1. **Trigger Manual Bonsai State Pruning:**
   ```bash
   curl -X POST --data '{"jsonrpc":"2.0","method":"admin_pruneState","params":[],"id":1}' \
     -H "Content-Type: application/json" http://127.0.0.1:8545
   ```
2. **Verify Archive Log Export:**
   Confirm historical receipts and logs older than 2,048 blocks have synced to AWS S3 Glacier.
3. **Trigger RocksDB Compaction:**
   ```bash
   curl -X POST --data '{"jsonrpc":"2.0","method":"admin_compactStorage","params":[],"id":1}' \
     -H "Content-Type: application/json" http://127.0.0.1:8545
   ```
