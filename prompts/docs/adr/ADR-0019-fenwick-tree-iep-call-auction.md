# ADR-0019: Dual Fenwick Trees for O(log K) Call Auction Equilibrium Pricing

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Algorithms Engineer, Matching Engine Lead  

---

## 1. Context
Calculating the Indicative Equilibrium Price (IEP) in real time during circuit breaker call auctions requires intersecting cumulative Bid and Ask demand curves. Naive array scanning has $O(N \cdot K)$ complexity, causing CPU lockup during market-open order spikes.

---

## 2. Decision
Maintain cumulative Bid and Ask demand curves using **Dual Fenwick Trees (Binary Indexed Trees)** in the Rust matching engine:
- Point volume updates and order cancellations execute in $O(\log K)$ time.
- Binary search over the Fenwick cumulative arrays finds the maximum tradable volume intersection $P_{\text{IEP}}$ in $O(\log K)$ time.

---

## 3. Consequences
- **Positive:** Enables sub-millisecond IEP discovery for 100,000+ price levels; eliminates matching engine stalls during call auctions.
- **Trade-offs:** Requires bounded price tick discretization.
