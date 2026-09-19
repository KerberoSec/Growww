# ADR-0017: Pre-Settlement Demat Finality Gating for Depository Transfers

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Risk Officer, Head of Custody Operations  

---

## 1. Context
Asynchronous batch settlement rejections by physical depositories (NSDL/CDSL) risk triggering cascading unwinds across secondary market trades if unconfirmed token deposits trade freely.

---

## 2. Decision
Enforce **Strict Pre-Settlement Demat Finality Gating**:
- Digital security tokens are minted on Hyperledger Besu **only after** physical depository delivery confirmation (DIS settlement) is verified in the custodian pool.
- In-flight unconfirmed deposits trade exclusively in a restricted pre-settlement book backed 100% by platform margin and cannot be withdrawn until final depository confirmation.

---

## 3. Consequences
- **Positive:** Completely eliminates cascading secondary market trade unwinds; preserves 100% physical asset backing invariant.
- **Trade-offs:** Adds a deposit verification window before on-chain token minting.
