# ADR-0005: ERC-3643 Permissioned Security Token Standard for RWA

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Lead Smart Contracts Engineer, Head of Regulatory Compliance  

---

## 1. Context
Tokenized equities and real-world assets on the permissioned Besu ledger must strictly enforce regulatory compliance rules, such as KYC verification, sanctions screening, jurisdiction restrictions, and court-ordered asset freezing/seizure at the token contract layer.

---

## 2. Decision
Implement security tokens using the **ERC-3643 (T-REX)** permissioned token standard.
- Every transfer executes an on-chain compliance hook consulting an `IdentityRegistry`.
- An authorized `FreezeManager` contract provides regulatory freezing and court-ordered asset recovery mechanisms.
- All tokens are backed 1:1 by physical depository shares held in segregated custodian accounts.

---

## 3. Alternatives Considered
- **Standard ERC-20:** Unpermissioned; allows transfers to non-KYC addresses, directly violating SEBI and FATF requirements.
- **ERC-1400:** Highly modular but overly complex partitioned architecture, resulting in excessive gas overhead on EVM settlement calls.

---

## 4. Consequences
- **Positive:** Guaranteed compliance enforcement at the smart contract level; standardized interfaces for institutional custodians and auditors.
- **Trade-offs:** Token transfers require interaction with the on-chain identity registry, increasing transaction execution gas.
