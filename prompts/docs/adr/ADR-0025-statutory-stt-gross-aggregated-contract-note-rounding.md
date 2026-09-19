# ADR-0025: Statutory STT Calculation on Gross Daily Aggregated Notional

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Head of Regulatory Compliance, Lead Tax Architect, Principal Ledger Engineer  

---

## 1. Context
Calculating Securities Transaction Tax (STT) on individual fractional micro-fills and rounding each transaction independently produces up to 20% cumulative discrepancy compared to statutory clearing house (NSE/BSE/NSDL) contract notes governed by Section 102 of the Finance Act 2004.

---

## 2. Decision
1. Calculate STT strictly on the **gross aggregated daily delivery turnover per ISIN** per client at end-of-day (EOD) contract note generation:
   $$\text{Statutory STT} = \text{round}\left(\left(\sum \text{Executed Traded Value}\right) \times 0.001\right)$$
2. Intra-day execution estimates reserve STT with micro-cent ceiling rounding in virtual hold accounts, with exact compensating debit/credit reconciliation executed during post-market batch settlement.

---

## 3. Consequences
- **Positive:** Guarantees 100% mathematical parity with SEBI and CBDT broker contract note requirements.
- **Trade-offs:** Requires EOD tax reconciliation pass to release excess intra-day STT reserve holds.
