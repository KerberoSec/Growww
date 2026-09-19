# ADR-0035: Perpetual Futures Fair Mark Price & Clamped 8-Hour Funding Rate

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Risk Officer, Head of Quantitative Research, Principal Derivatives Architect  

---

## 1. Context
Derivatives markets require robust protection against price manipulation and flash crashes on isolated order books, while keeping perpetual contract prices tethered to the underlying spot index over time.

---

## 2. Decision
1. **Fair Mark Price:** Liquidations and unrealized P&L are computed strictly using a composite 5-exchange weighted median spot index plus a 30-minute basis moving average, preventing unnatural liquidation wicks.
2. **8-Hour Funding Mechanism:** Funding is calculated every 8 hours clamped between $-0.75\%$ and $+0.75\%$.
3. **Zero Exchange Intermediation Fee:** Funding transfers occur peer-to-peer directly between longs and shorts with 0% exchange fee deduction.

---

## 3. Consequences
- **Positive:** Protects leveraged positions from book illiquidity; aligns perpetual prices tightly with spot.
- **Trade-offs:** Requires low-latency external index price feeds with outlier rejection filters.
