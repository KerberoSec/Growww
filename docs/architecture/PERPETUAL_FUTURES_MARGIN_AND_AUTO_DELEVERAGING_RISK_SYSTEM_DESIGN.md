# Perpetual Futures Margin & Auto-Deleveraging (ADL) Risk System Design

**Specification ID:** SPEC-ARCH-040-PERP  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Derivatives Risk Engine & Liquidation Waterfall  
**Owner:** Quantitative Risk & Derivatives Engineering Group  

---

## 1. Executive Summary & Zero-Fee Trading Invariant
The Perpetual Futures risk engine manages cross-margin and isolated-margin risk across USD-Margined and Crypto-Margined perpetual contracts (BTC/USDT):
- **Universal Zero-Fee Derivatives**: Strictly **0.00% trading fee** (No fee at all for Maker or Taker).
- **Anti-Manipulation Mark Price**: Uses a median-of-three composite index price combined with an exponential moving average (EMA) basis filter to eliminate flash-crash liquidations.
- **Smart Step-Down Liquidation**: Partially liquidates only the minimum contract quantity required to restore the maintenance margin ratio, avoiding unnecessary full-account wipes.
- **Auto-Deleveraging (ADL) Waterfall**: Transparent 5-tier ADL priority queue ranking positions by unrealized PnL percentage and effective leverage.

---

## 2. Mathematical Models & Margin Calculations

### 2.1 Mark Price Formulation
$$\text{MarkPrice} = \text{IndexPrice} + \text{Median}(\text{Basis}_1, \text{Basis}_2, \text{Basis}_3)$$
$$\text{Basis} = \text{EMA}_{30\text{s}}\left(\frac{\text{Bid}_1 + \text{Ask}_1}{2} - \text{IndexPrice}\right)$$

### 2.2 Maintenance Margin Ratio (MMR) & Liquidation Trigger
Liquidation triggers immediately when Account Equity falls below Maintenance Margin:
$$\text{MarginRatio} = \frac{\text{MaintenanceMargin}}{\text{AccountEquity}} \ge 100\%$$
$$\text{AccountEquity} = \text{WalletBalance} + \sum \text{UnrealizedPnL}$$

### 2.3 8-Hour Clamped Funding Rate
$$\text{PremiumIndex } (P) = \frac{\max(0, \text{ImpactBid} - \text{MarkPrice}) - \max(0, \text{MarkPrice} - \text{ImpactAsk})}{\text{IndexPrice}}$$
$$\text{FundingRate } (F) = \text{Clamp}\left(P + \text{Clamp}(r - P, -0.05\%, +0.05\%), -0.75\%, +0.75\%\right)$$
- Interest rate component $r = 0.01\%$ per 8 hours.
- Funding transfers occur directly peer-to-peer between longs and shorts with zero exchange fee deductions.

---

## 3. Liquidation & ADL Waterfall Engine

```
+----------------------------------------------------------------------------------------------------+
| MULTI-STAGE DERIVATIVES LIQUIDATION WATERFALL                                                      |
|                                                                                                    |
|  [ Margin Ratio >= 100% Detected ]                                                                |
|                 |                                                                                  |
|                 v                                                                                  |
|  [ Stage 1: Immediate Order Cancellation ] -> Cancels all open orders in opposite direction        |
|                 |                                                                                  |
|                 v                                                                                  |
|  [ Stage 2: Partial Smart Step-Down ] ------> Liquidates minimum size to drop to lower tier bracket|
|                 |                                                                                  |
|                 v (If Equity Still Insufficient)                                                   |
|  [ Stage 3: Liquidation Engine Takeover ] ---> Takes remaining position at bankruptcy price        |
|                 |                                                                                  |
|                 +----------------------------+-----------------------------+                       |
|                 | (Positive Remainder)                                     | (Deficit)             |
|                 v                                                          v                       |
|  [ Stage 4: Credit Insurance Fund ]                        [ Stage 5: Insurance Fund Deficit ]     |
|  (Excess balance transferred to buffer)                                    |                       |
|                                                                            v                       |
|                                                            [ Stage 6: Auto-Deleveraging (ADL) ]    |
|                                                            (Closes opposing top-ranked positions)  |
+----------------------------------------------------------------------------------------------------+
```

### 3.1 ADL Ranking Formula:
$$\text{RankingFactor} = \text{PnLPercentile} \times \text{EffectiveLeverage}$$
Opposing positions in Quantile 5 (highest leverage and highest profit) are automatically closed at the bankruptcy price of the liquidated account, preventing systemic bad debt.
