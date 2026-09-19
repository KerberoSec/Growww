# ADR-0039: Peer-to-Peer (P2P) Automated Escrow Lockbox & Dispute Arbitration

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Head of P2P Operations, Risk Operations Lead, Legal Counsel  

---

## 1. Context
Direct peer-to-peer fiat-to-crypto trading requires trustless collateral locking to prevent seller default and automated dispute workflows to protect buyers against fraudulent sellers.

---

## 2. Decision
1. **Automated Escrow Lockbox:** Sellers' crypto is locked atomically in the P2P Escrow Account upon order creation with a 15-minute buyer payment timer.
2. **Strict KYC Name Invariant:** Fiat sender bank account name must 100% match the user's KYC record. Third-party payments are strictly prohibited.
3. **Maker-Checker Arbitration:** Disputes are resolved through automated Open Banking UTR verification and compliance officer review.

---

## 3. Consequences
- **Positive:** Enables frictionless fiat on-ramping without chargeback or counterparty default risk.
- **Trade-offs:** Requires 24/7 compliance team coverage for contested payment arbitrations.
