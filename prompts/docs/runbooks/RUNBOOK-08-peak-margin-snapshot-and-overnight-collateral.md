# Runbook 08: SEBI Peak Margin Snapshot & Overnight Collateral Management

**Runbook ID:** RBK-OPS-008  
**Severity Tier:** Operational Standard  
**Authority:** Risk Manager / Settlement Officer  

---

## 1. Description & Regulatory Context
Under SEBI Peak Margin guidelines, clearing members must report 4 intraday margin snapshots randomly selected during exchange trading hours (09:15 to 15:30 IST). Any client margin deficit triggers statutory penalties.

---

## 2. Overnight & Weekend Collateral Partitioning
1. **Trading Session Differentiation:**
   - **Primary Session (09:15 to 15:30 IST):** Standard SPAN margin with intraday credit lines.
   - **Extended / 24/7 Session:** All orders require **100% pre-funded cash or RBI e-Rupee CBDC**.
2. **Snapshot Execution:**
   - `risk-service` captures client margin utilization at random intervals and commits the state to `reporting-service`.
   - Any detected shortfall automatically alerts the client and places the account in `REDUCE_ONLY` status before the next snapshot window.
