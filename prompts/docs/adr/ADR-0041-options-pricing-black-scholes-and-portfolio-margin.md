# ADR-0041: Real-Time Options Pricing & SPAN-Style Portfolio Margin

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Risk Officer, Head of Derivatives Trading, Principal Options Architect  

---

## 1. Context
Option traders require sub-millisecond Greeks calculations and risk-based portfolio margining to offset correlated option legs against underlying perpetual futures, eliminating excessive capital requirements.

---

## 2. Decision
1. **Black-Scholes-Merton Engine:** Compute analytical Greeks ($\Delta, \Gamma, \Theta, \mathcal{V}, \rho$) on every underlying price tick in $< 5\mu\text{s}$.
2. **SABR Volatility Surface:** Fit dynamic SABR volatility smiles across strike prices and expiry cycles.
3. **16-Scenario Portfolio Margin:** Simulate full portfolio risk across 16 price and volatility shift scenarios, granting up to 85% margin relief for delta-hedged spreads.

---

## 3. Consequences
- **Positive:** Dramatic capital efficiency for market makers and options spread traders.
- **Trade-offs:** Requires continuous compute overhead for real-time multi-scenario portfolio evaluations.
