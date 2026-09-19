# Runbook 23: Iceberg & Algorithmic Order Stuck Slicing Remediation

**Runbook ID:** RBK-OPS-023  
**Severity Tier:** P2 (High)  
**Authority:** Trading Systems Operations Lead / Matching Engine Lead  

---

## 1. Description & Trigger Conditions
Triggered when:
- An Iceberg or TWAP algorithmic parent order encounters a stuck slicing transition (child order filled but subsequent tranche not spawned).
- An OCO multi-leg order experiences a desynchronization where one leg is executed but the counter-leg fails to auto-cancel.

---

## 2. Remediation Workflow

```
[1. Identify Stuck Algorithmic Parent ID]
  - Query metrics: matching_engine_algo_slicing_errors_total, trigger_engine_desync_alerts
  - Identify parent_order_id, order_type (e.g. ICEBERG, OCO, TWAP), and instrument ISIN
                  |
                  v
[2. Query In-Memory Matching Engine State]
  - Query matching-engine administrative endpoint:
      GET /api/v1/engine/admin/order-status?parent_id=<PARENT_ID>
  - Inspect parent remaining quantity, active child tranches, and fill history
                  |
                  v
[3. Trigger Algorithmic Reconciliation / Force-Cancel]
  - If OCO orphan leg is detected in the active book:
      - Submit atomic administrative cancel command:
          POST /api/v1/engine/admin/purge-orphan-leg
          {"parent_id": "<PARENT_ID>", "orphan_leg_id": "<LEG_ID>"}
  - If Iceberg replenishment stalled:
      - Trigger re-slice evaluation or cancel remaining parent quantity
      - Release unexecuted collateral hold back to user available ledger balance
                  |
                  v
[4. Audit Logging & Verification]
  - Verify user account balance and open order list are 100% reconciled
  - Emit operational audit log on Hyperledger Besu and notify trader
```
