# Runbook 07: Cross-Chain Reorganization & Clawback Saga Execution

**Runbook ID:** RBK-OPS-007  
**Severity Tier:** P1 (Critical)  
**Authority:** Risk Operations Lead / SRE On-Call  

---

## 1. Description & Triggers
Triggered when an external blockchain (Bitcoin, Ethereum, Solana) experiences a deep reorganization exceeding standard confirmation depth after deposit credit was granted.

---

## 2. 5-Phase Clawback Saga Workflow

```
[Deep Reorg Detected by Gateway Watchdog]
                  |
                  v
[Phase 1: MARGIN_RECOVERY_LOCKED]
  - Target account state set to locked
  - Active open limit orders mass-purged from matching engine
                  |
                  v
[Phase 2: Collateral Valuation Reassessment]
  - Recalculate account SPAN margin without the invalidated deposit
                  |
                  v
[Phase 3: Priority Auto-Liquidation]
  - If margin deficit exists, liquidate positions to restore maintenance margin
                  |
                  v
[Phase 4: SGF Buffer Loss Absorption]
  - If liquidation proceeds fall short, draw down GIFT City SGF buffer
                  |
                  v
[Phase 5: Besu Synthetic Token Burn]
  - Execute on-chain burn of unbacked synthetic representation to restore INV-1
```

---

## 3. Post-Incident Forensic Audit
Export event logs to the compliance vault and file an incident notice with IFSCA within 24 hours.
