# ADR-0044: RWA Launchpad Dutch Auctions & Smart Contract Linear Vesting

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Head of Primary Markets, Lead Smart Contracts Architect, Legal Counsel  

---

## 1. Context
Primary offerings of tokenized real estate and private credit require fair price discovery without gas wars or bot front-running, coupled with automated linear vesting.

---

## 2. Decision
1. **Dutch Auction Engine:** Executes uniform price clearing where all winning participants pay the final clearing price $P_{clear}$.
2. **Pro-Rata Over-Subscription Pools:** Guarantees fair proportional allocation with sub-second automated excess refund.
3. **Linear Vesting Vaults:** Implements second-by-second streaming token unlocks on Hyperledger Besu with zero gas claim costs via EIP-2771.

---

## 3. Consequences
- **Positive:** Fair, bot-resistant primary allocations; transparent streaming vesting for investors.
- **Trade-offs:** Requires on-chain smart contract vesting schedules.
