# Decentralized Oracle Aggregation & Market Feed Arbiter Specification

**Specification ID:** SPEC-ARCH-014-ORACLE  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Market Data Integrity & Oracle Security  
**Owner:** Quantitative Infrastructure & Market Arbiter Group  

---

## 1. Executive Summary & Anti-Manipulation Framework
The Decentralized Oracle Aggregation engine establishes authoritative mark and index prices for spot and derivatives trading, eliminating single-exchange manipulation:
- **Multi-Source Aggregation**: Synthesizes real-time feeds from Binance, Coinbase Pro, OKX, Bybit, Chainlink decentralized nodes, and Pyth Network.
- **Medianizer with Outlier Rejection**: Rejects feeds that deviate by $> 3\sigma$ from the median of all active exchanges.
- **Dynamic Staleness Circuit Breaker**: If market feed updates for a pair drop below 1 update per 2,000ms, trading for that pair enters a protective auction or temporary halt.

---

## 2. Oracle Data Pipeline & Median Formulation

```
+----------------------------------------------------------------------------------------------------+
| MULTI-SOURCE ORACLE INGESTION & OUTLIER FILTERING PIPELINE                                         |
|                                                                                                    |
|  [ Binance WSS ]    [ Coinbase WSS ]    [ OKX WSS ]    [ Chainlink Node ]    [ Pyth Network ]      |
|         |                  |                 |                 |                     |             |
|         +------------------+-----------------+-----------------+---------------------+             |
|                                              |                                                     |
|                                              v                                                     |
|                                [ High-Speed Ingress Arbiter ]                                      |
|                                - Timestamp Validation (< 500ms jitter)                             |
|                                - Format Normalization (6 Decimals USDT)                            |
|                                              |                                                     |
|                                              v                                                     |
|                             [ Outlier Detection & Medianizer ]                                     |
|                             - Compute Median: M = Median(P_1, P_2, ..., P_N)                       |
|                             - Drop Feeds where |P_i - M| / M > 0.50%                               |
|                                              |                                                     |
|                                              v                                                     |
|                             [ Authoritative Index Price Broadcast ]                                |
|                             - Emitted to Matching Engine & Risk Arbiter                            |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 Index Price Formula:
$$\text{IndexPrice} = \sum_{i=1}^K w_i P_i, \quad w_i = \frac{\text{Volume}_i}{\sum \text{Volume}_j}$$
Weights $w_i$ reflect the 24-hour volume of the constituent exchange, ensuring liquidity-weighted price accuracy.
