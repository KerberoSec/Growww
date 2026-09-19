# Runbook 15: Depository Batch Delivery Rejection & Compensating SGF Resolution

**Runbook ID:** RBK-OPS-015  
**Severity Tier:** P1 (Critical)  
**Authority:** Custody Operations Lead / Chief Risk Officer  

---

## 1. Description & Trigger Conditions
Triggered when NSDL or CDSL rejects an electronic Delivery Instruction Slip (DIS) during batch clearing for an unconfirmed deposit.

---

## 2. Remediation Workflow

```
[1. Isolate Pre-Settlement Account]
  - Freeze client's unconfirmed pre-settlement balance
  - Cancel any open limit orders in the restricted pre-settlement book
                  |
                  v
[2. Evaluate Secondary Market Exposure]
  - If shares were traded on pre-settlement margin, calculate total counterparty exposure
                  |
                  v
[3. Compensating SGF Settlement Execution]
  - Settlement Guarantee Fund (SGF) steps in as Central Counterparty (CCP)
  - Delivers shares from platform reserve pool to buyer
  - Debits client cash collateral to cover replacement share purchase
                  |
                  v
[4. File Incident Notice]
  - Log audit entry on Hyperledger Besu and notify NSDL/CDSL compliance
```
