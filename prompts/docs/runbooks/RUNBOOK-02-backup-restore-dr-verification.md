# Runbook 02: Disaster Recovery, Database Restore & Blockchain Node Resync

**Runbook ID:** RBK-OPS-002  
**Severity Tier:** P1 (Critical)  
**Authority:** Lead Site Reliability Engineer / Database Administrator  
**RTO / RPO Targets:** RTO < 15 Minutes | RPO = 0 (Financial Ledger)  

---

## 1. PostgreSQL Point-in-Time Recovery (PITR)

### Procedure
1. Provision target RDS Aurora / PostgreSQL instance from the latest continuous WAL archive snapshot.
2. Specify recovery target timestamp immediately prior to corruption event:
   ```ini
   recovery_target_time = '2026-09-18 14:30:00+05:30'
   recovery_target_action = 'promote'
   ```
3. Execute automated trial balance verification:
   ```sql
   SELECT asset_id, SUM(CASE WHEN direction = 'D' THEN amount ELSE -amount END) AS imbalance
   FROM journal_entry
   GROUP BY asset_id
   HAVING SUM(CASE WHEN direction = 'D' THEN amount ELSE -amount END) <> 0;
   ```
4. If zero imbalance rows are returned, promote instance to primary and update service connection strings.

---

## 2. Hyperledger Besu Validator Node Disaster Recovery

### Critical Constraint: Double-Signing Prevention
When restoring a Besu validator from snapshot, the node **must never sign a block at or below its previously recorded height**.

### Procedure
1. Extract validator node private key from AWS CloudHSM / Vault.
2. Restore chain data directory from validated snapshot.
3. Verify `last_signed_height` marker file:
   ```bash
   cat /var/lib/besu/data/last_signed_height.json
   ```
4. Start node in non-validating (archive/syncing) mode until local block height equals current network height.
5. Re-enable validator key once synchronization reaches tip + 1.

---

## 3. Automated Weekly Restore Drills
Every Sunday at 02:00 IST, automated CI pipelines restore production WAL backups into an isolated sandbox, execute ledger balance assertions, and verify cryptographic Merkle audit roots.
