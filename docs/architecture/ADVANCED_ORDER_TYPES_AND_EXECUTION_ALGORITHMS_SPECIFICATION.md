# Advanced Order Types & Algorithmic Execution Specification

This document defines the architectural state machines, matching engine execution algorithms, mathematical invariants, and validation rules for all supported order types and algorithmic slicing strategies on the Growww RWA Exchange.

---

## 1. Core Order Types & Time-in-Force (TIF)

### 1.1 Limit Orders & Priority Invariants
Limit orders specify an exact maximum purchase price or minimum selling price.
- **Queue Priority:** Strict Price-Time (FIFO) priority within each price level.
- **Time-in-Force (TIF) Instructions:**
  - **GTC (Good-Til-Cancelled):** Remains resting in the order book until fully matched, explicitly cancelled by the user, or expired after 90 days.
  - **IOC (Immediate-or-Cancel):** Immediately matches against resting liquidity up to the limit price; any unexecuted remaining quantity is immediately cancelled without resting in the book.
  - **FOK (Fill-or-Kill):** Must be matched entirely upon arrival at or better than the limit price; if the entire quantity cannot be filled immediately, the entire order is rejected.
  - **DAY (Valid for Day):** Automatically cancelled at market session close (15:30 IST for primary hours, or 23:59:59 UTC for 24/7 synthetic books).
  - **GTD (Good-Til-Date/Time):** Auto-expires at a specified unix timestamp $T_{expire}$.

### 1.2 Market Orders with Slippage Protection
Market orders execute immediately against the best resting counterparty liquidity in the order book.
- **Dynamic Slippage Clamping:** To protect retail users from flash crashes and low-liquidity spikes, market orders are internally converted to an aggressive limit order with a configurable maximum slippage band:
  $$P_{max\_buy} = \text{BestAsk} \times (1 + \text{SlippageTolerance})$$
  $$P_{min\_sell} = \text{BestBid} \times (1 - \text{SlippageTolerance})$$
  - Default retail tolerance: $\pm 1.0\%$ (100 bps).
  - Any portion of the order that would execute beyond the slippage boundary is cancelled via IOC behavior.

### 1.3 Post-Only (Maker-Only) Orders
- **Guarantee:** Guarantees that the order will strictly act as a liquidity maker.
- **Matching Rule:** If the order would immediately cross with an existing resting order upon arrival, the entire order is rejected by the matching engine with code `ERR_POST_ONLY_WOULD_CROSS` without executing any trades or paying taker fees.

---

## 2. Advanced Conditional & Stop Orders

### 2.1 Stop-Loss Market (SL-M) & Stop-Loss Limit (SL-L)
Stop orders remain dormant in the **Trigger Engine** and do not enter the active order book until the market reaches the specified trigger price:
- **Trigger Evaluation Reference:** Evaluated on every matching cycle against the Last Traded Price (LTP).
- **Buy Stop-Loss:** Triggers when $\text{LTP} \ge \text{TriggerPrice}$.
- **Sell Stop-Loss:** Triggers when $\text{LTP} \le \text{TriggerPrice}$.
- **Transition Execution:**
  - **SL-M:** Spawns an aggressive Market IOC order with standard slippage clamping upon trigger.
  - **SL-L:** Injects an active Limit order at `LimitPrice` into the matching engine book with fresh time priority upon trigger.

### 2.2 Trailing Stop-Loss Orders
Dynamically trails favorable price moves, locking in profits while protecting against adverse reversals.
- **Ratcheting State Variables:**
  - `peak_price` (for sell orders): $\max(\text{peak\_price}, \text{LTP})$.
  - `trough_price` (for buy orders): $\min(\text{trough\_price}, \text{LTP})$.
- **Dynamic Trigger Formula:**
  - **Sell Trailing Stop:** $\text{ActiveTrigger} = \text{peak\_price} - \text{TrailOffset}$ (or $\text{peak\_price} \times (1 - \text{TrailPercent})$).
  - **Buy Trailing Stop:** $\text{ActiveTrigger} = \text{trough\_price} + \text{TrailOffset}$ (or $\text{trough\_price} \times (1 + \text{TrailPercent})$).
- **Execution:** When $\text{LTP} \le \text{ActiveTrigger}$ (for sells) or $\text{LTP} \ge \text{ActiveTrigger}$ (for buys), the order triggers and routes to the matching engine.

### 2.3 One-Cancels-the-Other (OCO) Orders
An atomic pairing of a **Take-Profit Limit Order** and a **Stop-Loss Order** (Limit or Market):
- **Atomic Linking:** Both legs share a single parent `oco_id` and reserve collateral once based on the maximum possible exposure.
- **Mutual Cancellation:** As soon as either leg is partially or fully executed (or triggered), the matching engine atomically purges the remaining leg from the order book and trigger engine.

### 2.4 Bracket & Cover Orders
- **Cover Order:** An active Limit or Market entry order paired with a mandatory Stop-Loss leg, allowing higher leverage under strict risk limits.
- **Bracket Order:** An entry order paired simultaneously with a Take-Profit limit leg, a Stop-Loss leg, and an optional Trailing Stop offset.

---

## 3. Institutional Algorithmic Slicing & Execution Strategies

### 3.1 Iceberg Orders (Native Slicing)
Iceberg orders mask large institutional order quantities by exposing only a small visible tranche in the public Level-2 order book.
- **Parameters:**
  - `TotalQuantity` ($Q_{tot}$): Total parent order size.
  - `DisplayQuantity` ($Q_{disp}$): Visible tranche displayed in the order book.
  - `VariancePercent` ($\delta$): Optional randomization factor (e.g. $\pm 10\%$) applied to $Q_{disp}$ to prevent algorithmic detection.
- **Replenishment State Machine:**
  1. Child order with size $Q_{disp}$ is placed into the book with standard price-time priority.
  2. When the child order is fully filled, the matching engine automatically decrements $Q_{tot}$ and spawns the next child tranche with fresh timestamp at the tail of that price level.
  3. Cycle repeats until $Q_{tot} = 0$.

### 3.2 Time-Weighted Average Price (TWAP)
Executes large block orders evenly over a designated time duration $[T_{start}, T_{end}]$ in discrete intervals:
- **Slice Quantity:** $Q_{slice} = \frac{Q_{tot}}{N_{slices}}$ with configurable interval $\Delta t = \frac{T_{end} - T_{start}}{N_{slices}}$.
- **Randomized Jitter:** Each slice applies Gaussian time jitter $\Delta t \pm \varepsilon$ and volume variance to minimize market footprint.

### 3.3 Scaled / Grid Orders
Splits a total order quantity into multiple limit orders distributed evenly across a specified price range $[P_{low}, P_{high}]$:
- **Distribution Modes:** Flat, Linear Ascending, Linear Descending, or Exponential.
- **Use Case:** Market making, automated range accumulation, and dollar-cost averaging (DCA).

---

## 4. Fee Invariant on Advanced Orders

All advanced, algorithmic, and conditional order types strictly adhere to the exchange fee policy:
- **0 Brokerage:** ₹0 on all order types.
- **Single 0.00% (No fee at all) Platform Fee:** Calculated solely on executed fills ($\text{FillNotional} \times 0.0001$).
- **Zero Surcharge on Algo Slicing:** No fee penalties or premiums for Iceberg, TWAP, or Grid child orders.
