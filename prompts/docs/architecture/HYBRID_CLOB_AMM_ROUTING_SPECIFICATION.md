# Hybrid CLOB-AMM Liquidity Routing & Concentrated Liquidity Engine

**Specification ID:** SPEC-ARCH-025-CLOBAMM  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Liquidity Aggregation & Smart Order Routing (SOR)  
**Owner:** Quantitative Trading & Liquidity Routing Group  

---

## 1. Executive Summary & Zero-Fee Invariant
The Hybrid CLOB-AMM Smart Order Router (SOR) dynamically splits and routes incoming user orders across the Central Limit Order Book (CLOB) and on-chain Concentrated Liquidity Automated Market Maker (CLMM) pools:
- **Universal Zero-Fee Execution**: Strictly **0.00% routing and trading commission** (No fee at all).
- **Optimal Execution Pathfinding**: Solves the convex optimization problem to minimize slippage and price impact across discrete orderbook ladders and continuous liquidity curves.
- **Sub-Millisecond Routing Decisions**: Pre-computed liquidity graphs evaluated in C++20/Rust in $< 15\mu\text{s}$.

---

## 2. Mathematical Splitting Optimization

```
+----------------------------------------------------------------------------------------------------+
| HYBRID CLOB-AMM DYNAMIC ORDER SPLITTING ENGINE                                                     |
|                                                                                                    |
|  [ Inbound Order (Quantity Q) ] ---> [ Smart Order Router (SOR) Engine ]                           |
|                                                    |                                               |
|                                                    v                                               |
|                           [ Convex Optimization: min PriceImpact(q_1, q_2) ]                       |
|                                           such that q_1 + q_2 = Q                                  |
|                                                    |                                               |
|                         +--------------------------+--------------------------+                    |
|                         |                                                     |                    |
|                         v                                                     v                    |
|             [ Route Leg 1: CLOB Orderbook ]                       [ Route Leg 2: CLMM Pool ]       |
|             - Quantity: q_1                                       - Quantity: q_2                  |
|             - Fills top N depth price levels                      - Swaps along concentrated tick  |
|             - Microsecond Rust matching                           - Hyperledger Besu pool swap     |
|                         |                                                     |                    |
|                         +--------------------------+--------------------------+                    |
|                                                    |                                               |
|                                                    v                                               |
|                                     [ Unified Execution Fill Event ]                               |
|                                     - Zero Platform Fees Deducted                                  |
|                                     - Delivered Atomically to Trader                               |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 Optimization Objective:
$$\min_{q_1, q_2} \left( \int_0^{q_1} P_{\text{CLOB}}(x) \, dx + \int_0^{q_2} P_{\text{AMM}}(y) \, dy \right) \quad \text{subject to } q_1 + q_2 = Q, \quad q_1, q_2 \ge 0$$
Where $P_{\text{CLOB}}(x)$ is the marginal price in the order book at depth $x$, and $P_{\text{AMM}}(y)$ is the marginal tick price in the concentrated AMM pool.
