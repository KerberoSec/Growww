# Runbook 18: STT Tax Contract Note Discrepancy & Reconciliation

**Runbook ID:** RBK-OPS-018  
**Severity Tier:** P2 (High)  
**Authority:** Tax Operations Lead / Clearing & Settlement Lead  

---

## 1. Description & Trigger Conditions
Triggered during End-of-Day (EOD) contract note generation when:
- Discrepancy detected between intra-day estimated STT reserve holds and statutory EOD gross aggregated STT.
- Difference between clearing house contract note STT and platform internal tax ledger exceeds 0 paise ($INR > 0.00$).

---

## 2. Remediation Workflow

```
[1. Lock Account Settlement Batch]
  - Flag client account settlement ledger in PENDING_TAX_RECONCILIATION status
  - Prevent final payout dispatch until tax differential is computed
                  |
                  v
[2. Re-Compute Gross Aggregated STT]
  - Query tax-service: GET /api/v1/tax/recalculate-stt?date=YYYY-MM-DD&client_id=<ID>
  - Re-sum all executed buy/sell turnover per ISIN across trading day
  - Apply statutory 0.1% rate on gross sum: Statutory_STT = round(Gross_Turnover * 0.001)
                  |
                  v
[3. Generate Compensating Tax Ledger Adjustment]
  - Calculate variance: Variance = Sum(IntraDay_STT_Holds) - Statutory_STT
  - If Variance > 0 (over-reserved): Credit difference back to client available cash balance
  - If Variance < 0 (under-reserved): Debit difference from client cash balance
  - Post atomic double-entry journal entry: Account 2100 (STT Payable) vs Account 1100 (Client Cash)
                  |
                  v
[4. Issue Final Reconciled Contract Note]
  - Generate digitally signed PDF/JSON contract note with exact statutory STT breakdown
  - Dispatch to client and regulatory reporting repository
```
