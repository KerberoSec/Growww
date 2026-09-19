# ADR-0042: Copy Trading Low-Latency Replication & High-Water Mark Profit Sharing

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Head of Product, Principal Matching Engine Architect, Risk Operations Lead  

---

## 1. Context
Followers copying Master Traders require sub-15ms execution mirroring without adverse slippage, while Master Traders require automated high-water mark profit sharing.

---

## 2. Decision
1. **Proportional Mirror Engine:** Slices child orders in parallel based on relative follower-to-master equity ratios with $\pm 0.5\%$ slippage clamping.
2. **High-Water Mark (HWM) Profit Sharing:** Master Traders receive 10% to 15% profit-share settled weekly strictly when follower net equity exceeds historical HWM.
3. **Follower Drawdown Circuit Breaker:** Automated emergency detachment when follower drawdown breaches user-configured stop loss (e.g. $-15\%$).

---

## 3. Consequences
- **Positive:** Social trading incentives aligned; automated risk detachment protects follower capital.
- **Trade-offs:** Fast market moves may reject child orders exceeding the 0.5% slippage bound.
