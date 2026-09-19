# ADR-0013: Statutory Core SGF Default Loss Absorption Waterfall

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Risk Officer, Clearing Corporation Managing Director  

---

## 1. Context
A clearing member default on settlement obligations requires an unambiguous, legally binding capital drawdown waterfall aligned with SEBI Master Regulations (`SEBI/HO/MRD2/DCAP/CIR/P/2020/127`).

---

## 2. Decision
Implement the **5-Tier Statutory Core SGF Waterfall**:
1. Defaulter's Initial Margin & Collateral
2. Defaulter's Contribution to Core SGF
3. Platform First-Loss Capital (25% allocation of retained earnings)
4. Core Settlement Guarantee Fund Corpus
5. Non-Defaulting Members' Pro-Rata Assessment (capped at 100% initial margin)

---

## 3. Consequences
- **Positive:** Full regulatory compliance; ring-fences non-defaulting member assets; preserves market solvency.
- **Trade-offs:** Requires capital allocation commitments from platform treasury.
