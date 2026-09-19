# ADR-0043: Hybrid CLOB-AMM Routing & Concentrated Liquidity Pools

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Trading Systems Architect, Head of Market Structure  

---

## 1. Context
Long-tail fractional RWAs (real estate, private debt) suffer from wide spreads in standard limit order books, requiring automated concentrated liquidity pools integrated into the matching engine.

---

## 2. Decision
1. **Concentrated Liquidity AMM:** Implement virtual tick reserves ($P(i) = 1.0001^i$) providing up to 4,000x capital efficiency for custom price bands.
2. **Smart Order Routing (SOR):** Evaluates marginal execution price curves and atomically splits orders across both CLOB and AMM pools.
3. **Internal JIT Arbitrage:** Captures price divergence profits directly for the exchange Settlement Guarantee Fund (SGF).

---

## 3. Consequences
- **Positive:** Deep liquidity for illiquid RWAs; minimal price slippage for traders.
- **Trade-offs:** Dual book evaluation adds minor routing latency ($\approx 25\mu\text{s}$).
