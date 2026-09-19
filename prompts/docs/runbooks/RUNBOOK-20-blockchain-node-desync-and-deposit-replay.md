# Runbook 20: Blockchain Node Desynchronization & Deposit Block Replay

**Runbook ID:** RBK-OPS-020  
**Severity Tier:** P1 (Critical)  
**Authority:** Blockchain Node Operations Lead / Lead Custody Engineer  

---

## 1. Description & Trigger Conditions
Triggered when:
- An external blockchain full node / RPC client (Bitcoin Core, Geth, Erigon, Solana RPC) experiences block height lag (> 10 blocks) or RPC connection drops.
- User crypto deposits are delayed or missing due to node ingestion desynchronization or chain reorgs.

---

## 2. Remediation Workflow

```
[1. Identify Affected Chain & Stalled Height]
  - Query metrics: custody_node_block_height_lag, custody_deposit_ingestion_errors_total
  - Identify affected blockchain daemon (e.g. Bitcoin, Ethereum Mainnet, Arbitrum)
  - Determine last confirmed block height processed by database ledger: H_db
  - Determine actual network tip block height: H_tip
                  |
                  v
[2. Failover to Redundant RPC Nodes]
  - If local node daemon is frozen/unhealthy, switch custody-adapter traffic to backup RPC cluster
  - Restart local validator/archive node with pruned state verification if corrupted
                  |
                  v
[3. Trigger Historical Block Ingestion Replay]
  - Invoke manual replay API:
      POST /api/v1/custody/admin/replay-blocks
      {
        "chain": "ETHEREUM_MAINNET",
        "start_block": H_db - 10,
        "end_block": H_tip
      }
  - Ingestion scanner re-evaluates all transactions matching derived user deposit addresses
  - Idempotent upsert prevents double crediting (ON CONFLICT (txid, vout) DO NOTHING)
                  |
                  v
[4. Verify Deposit Credit & User Notifications]
  - Confirm all uncredited deposits transition from CONFIRMING -> CREDITED
  - Dispatch in-app push and email notifications to affected depositors
  - Log audit verification report in administrative compliance console
```
