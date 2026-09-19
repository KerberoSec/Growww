# Runbook 26: Merkle Proof of Solvency Mismatch & Liability Recalculation

**Runbook ID:** RBK-OPS-026  
**Severity Tier:** P1 (Critical)  
**Authority:** Chief Technology Officer / Chief Cryptographer  

---

## 1. Description & Trigger Conditions
Triggered when:
- zk-SNARK proof generation fails during daily solvency snapshot.
- User client-side verification detects a hash path mismatch ($\text{RootComputed} \ne R_{solvency}$).
- Total verified on-chain assets drop below aggregated user liabilities ($Assets < Liabilities$).

---

## 2. Remediation Workflow

```
[1. Lock Snapshot Publishing & Quarantine Proof Engine]
  - Halt automated publishing of the daily Merkle Root to public portals
  - Prevent user-facing proof queries while investigation is active
                  |
                  v
[2. Re-Audit Double-Entry Ledger Balances]
  - Execute full ledger consistency scan:
      GET /api/v1/audit/ledger/verify-double-entry-balances
  - Assert zero negative balances across all user accounts: Balance_u >= 0
  - Compare total ledger liabilities against database snapshot sum
                  |
                  v
[3. Re-Generate zk-SNARK Blinded Merkle Sum Tree]
  - Re-run Groth16 proof circuit with freshly verified account commitments
  - Verify proof against on-chain SolvencyRegistry.sol contract on Hyperledger Besu
                  |
                  v
[4. Publish Verified Root & Open Self-Verification]
  - Publish verified Merkle root and updated on-chain asset signatures
  - Notify compliance auditor and unlock user self-verification API
```
