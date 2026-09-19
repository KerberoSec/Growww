# Runbook 04: Three-Way Reconciliation Break Resolution

**Runbook ID:** RBK-OPS-004  
**Severity Tier:** P1 (Critical)  
**Authority:** Financial Controller / Chief Risk Officer  

---

## 1. Description
Automated 3-way reconciliation compares balances across:
1. Internal Double-Entry Ledger (`wallet-service`)
2. On-Chain Token Supply (`SettlementDvP.sol` / `EquityToken.sol`)
3. Depository Holding Statements (NSDL / CDSL / Custodian)

---

## 2. Break Classification & Immediate Action

| Break Type | Threshold | Immediate Action |
|---|---|---|
| **Fractional Dust Variance** | $\le 1$ minor unit per holder | Auto-logged to reconciliation journal; no operational disruption |
| **In-Flight Settlement Lag** | Pending transactions within standard T+0 window | Monitored until next block/batch confirmation |
| **Material Unbacked Token Divergence** | $> 0$ unverified tokens | **Immediate trading halt on affected instrument.** Escalate to CRO and Custody Board. |

---

## 3. Resolution Workflow
1. Identify the out-of-balance asset and transaction ID from `reconciliation-service`.
2. Inspect bank settlement logs and NSDL/CDSL holding statements.
3. If an erroneous mint occurred, execute four-eyes administrative freeze via `FreezeManager.sol`.
4. Post compensating journal entries and publish an updated Proof of Reserve attestation.
