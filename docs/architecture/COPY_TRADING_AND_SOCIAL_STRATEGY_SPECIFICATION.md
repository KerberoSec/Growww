# Copy Trading & Social Strategy Engine Specification

**Specification ID:** SPEC-ARCH-007-COPY  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Social Trading & Automated Execution Routing  
**Owner:** Client Solutions & Quantitative Strategy Group  

---

## 1. Executive Summary & Invariants
The Copy Trading platform allows retail followers to mirror the trades of verified Master Traders in real-time with microsecond execution:
- **Zero Trading Fees**: Strictly **0.00% exchange trading commission** for all master and follower orders.
- **Proportional Position Sizing**: Automatically sizes follower orders based on the follower's allocated capital relative to the master trader's equity.
- **Slippage Protection Guard**: Follower orders execute only if the execution price is within 0.10% of the master trader's fill price, preventing front-running and adverse selection.
- **High-Water Mark (HWM) Performance Sharing**: Master traders receive voluntary performance rewards (e.g. 10%) calculated strictly on net new profits above the historical equity peak.

---

## 2. Mathematical Position Sizing & Execution Flow

### 2.1 Proportional Sizing Formula:
$$\text{FollowerOrderSize} = \text{MasterOrderSize} \times \left(\frac{\text{FollowerAllocatedEquity}}{\text{MasterTotalEquity}}\right)$$

### 2.2 High-Water Mark Profit Sharing:
$$\text{ProfitShare} = \max(0, \text{CurrentEquity} - \text{HighWaterMark}) \times \text{PerformanceRate}$$
If `CurrentEquity > HighWaterMark`, `HighWaterMark` updates to `CurrentEquity` upon distribution.
