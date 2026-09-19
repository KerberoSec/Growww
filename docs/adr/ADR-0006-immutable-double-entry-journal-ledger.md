# ADR-0006: Immutable Double-Entry Journal Ledger for Account Balances

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Lead Financial Systems Architect, Chief Financial Officer  

---

## 1. Context
Financial balances for fiat, stablecoins, and token holdings must maintain an unbroken audit trail, eliminate balance corruption from race conditions, and support historical point-in-time reconstruction for statutory regulatory filings.

---

## 2. Decision
Model all account balances strictly as a derived projection over an **immutable, append-only double-entry journal** in PostgreSQL.
- Every economic event creates a transaction composed of at least two balanced legs (`Debit == Credit`).
- Database-level deferred constraint triggers enforce balance conservation per transaction.
- Automated daily trial balances assert global balance consistency across all ledger accounts.

---

## 3. Alternatives Considered
- **Single-Entry Mutable Balances (`UPDATE accounts SET balance = balance + :amount`):** Prone to race conditions, lacks auditability, and fails financial audit standards.

---

## 4. Consequences
- **Positive:** Complete auditable history; mathematical impossibility of unrecorded balance adjustments; automated corruption detection.
- **Trade-offs:** Reading balances requires indexed aggregation or maintaining a synchronized materialized view.
