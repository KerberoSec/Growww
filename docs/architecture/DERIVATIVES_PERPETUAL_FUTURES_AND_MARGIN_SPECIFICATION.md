# Perpetual Futures, Derivatives & Margin Trading Specification

This document defines the mathematical models, risk parameters, funding rate mechanisms, mark price calculations, and auto-deleveraging (ADL) algorithms for perpetual futures and margin trading on the Growww RWA Exchange.

---

## 1. Perpetual Futures Mechanics & Mark Price Engine

### 1.1 Composite Index Price & Fair Mark Price
To protect traders from manipulation, wicks, and flash crashes on a single order book, liquidations and unrealized P&L are calculated strictly against the **Fair Mark Price** rather than the Last Traded Price (LTP).
1. **Composite Index Price ($P_{index}$):** Weighted median price aggregated from 5 major external liquidity venues (e.g. Binance, OKX, Bybit, Coinbase, Kraken):
   $$P_{index} = \text{Median}\left(P_1, P_2, P_3, P_4, P_5\right)$$
   Outlier feeds ($|P_i - \text{Median}| > 1.5\%$) are automatically dropped.
2. **Fair Mark Price ($P_{mark}$):**
   $$P_{mark} = P_{index} \times \left(1 + \text{BasisTWAP}_{30m}\right)$$
   Where $\text{BasisTWAP}_{30m} = \frac{1}{1800}\int_{t-1800}^t \left(\frac{P_{mid}(t) - P_{index}(t)}{P_{index}(t)}\right) dt$.

### 1.2 8-Hour Funding Rate Mechanism
Funding payments balance long and short contract demand every 8 hours (at 00:00, 08:00, 16:00 UTC):
1. **Premium Index ($P_{premium}$):**
   $$P_{premium} = \frac{\max(0, \text{ImpactBid} - P_{index}) - \max(0, P_{index} - \text{ImpactAsk})}{P_{index}}$$
2. **Funding Rate ($F$):**
   $$F = \text{clamp}\left(P_{premium} + \text{clamp}\left(I - P_{premium}, -0.05\%, +0.05\%\right), -0.75\%, +0.75\%\right)$$
   Where default interest rate differential $I = 0.01\%$ per 8 hours.
3. **Funding Cash Flow:**
   $$\text{Payment} = \text{PositionSize} \times P_{mark} \times F$$
   - If $F > 0$: Longs pay Shorts.
   - If $F < 0$: Shorts pay Longs.
   - Platform fee on funding payments: **0% (Pure zero-fee peer-to-peer exchange)**.

---

## 2. Margin System: Cross-Margin vs Isolated Margin

### 2.1 Isolated Margin Mode
- Margin allocated to a specific position is strictly quarantined in an isolated account ledger.
- If the isolated margin ratio breaches the Maintenance Margin Requirement (MMR), only the isolated position is liquidated. Total account equity in other assets remains 100% protected.

### 2.2 Cross-Margin Multi-Asset Collateral Pool
- All supported assets in the user's cross-margin account contribute to a unified collateral pool:
  $$\text{Total Collateral Value} = \sum_{i=1}^M \left(\text{Balance}_i \times P_{mark, i} \times (1 - \text{Haircut}_i)\right)$$
- **Statutory Collateral Haircut Matrix:**
  - Tokenized Government Securities (G-Secs): 5% Haircut (95% collateral value).
  - Blue-chip Equities (NIFTY 50 / S&P 500): 20% Haircut (80% collateral value).
  - Bitcoin (BTC) / Ethereum (ETH) / USDT / USDC: 20% Haircut (80% collateral value).
  - Mid-cap Equities & Tier-2 Altcoins: 50% Haircut (50% collateral value).

### 2.3 Tiered Position Brackets & Leverage De-escalation
To manage systemic risk, maximum allowable leverage automatically scales down as position size increases:

| Tier Bracket | Position Notional Range (USDT / INR) | Maximum Leverage | Initial Margin (IMR) | Maintenance Margin (MMR) |
|---|---|---|---|---|
| **Tier 1** | ₹0 to ₹50,00,000 ($0 to $60,000) | 20x | 5.0% | 2.5% |
| **Tier 2** | ₹50,00,001 to ₹2,00,00,000 ($60k to $250k) | 10x | 10.0% | 5.0% |
| **Tier 3** | ₹2,00,00,001 to ₹10,00,00,000 ($250k to $1.2M) | 5x | 20.0% | 10.0% |
| **Tier 4** | > ₹10,00,00,000 (> $1.2M) | 2x | 50.0% | 25.0% |

---

## 3. Liquidation & Auto-Deleveraging (ADL) Protocol

### 3.1 Liquidation Execution & 1.0% Fee
When $\text{Account Margin Ratio} < \text{MMR}$:
1. Matching engine cancels all open unexecuted limit orders to release reserved collateral.
2. Liquidation Engine takes over position and executes aggressive IOC/FOK market orders against resting book liquidity.
3. Levies flat **1.0% forced liquidation fee** on liquidated notional:
   - **80% to Insurance Fund / SGF** to guarantee zero-clawback solvency.
   - **20% to Treasury Relayer Pool** for compute and gas sponsorship.

### 3.2 Auto-Deleveraging (ADL) Fallback
If an extreme market gap occurs where the Insurance Fund / SGF is completely exhausted, the platform initiates ADL:
1. Counterparty positions are ranked by **ADL Score**:
   $$\text{ADL Score} = \text{ProfitRankingPercentile} \times \text{EffectiveLeverage}$$
2. The highest-ranked profitable positions are automatically deleveraged at the bankruptcy price of the liquidated account, preventing systemic insolvency without socialized clawbacks from uninvolved accounts.
