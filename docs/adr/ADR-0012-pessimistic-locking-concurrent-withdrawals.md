# ADR-0012: Pessimistic Row Locking for Concurrent Withdrawal Safety

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Lead Database Architect, Chief Risk Officer  

---

## 1. Context
Concurrent API requests for fiat withdrawals, order margins, and settlement debits can race against account balances, causing negative balance overdrafts if optimistic balance reads overlap.

---

## 2. Decision
Enforce **Pessimistic Row-Level Locking (`SELECT ... FOR UPDATE`)** inside the PostgreSQL double-entry withdrawal initiation stored procedure (`initiate_withdrawal`).
- Locks the specific account balance record for the duration of the debit transaction.
- Freezes free balance availability before inserting the `WITHDRAWAL_HOLD` journal entry.

---

## 3. Consequences
- **Positive:** Mathematically impossible to overdraw an account via concurrent race conditions.
- **Trade-offs:** Introduces row-level serialization for high-frequency concurrent actions on the same account.
