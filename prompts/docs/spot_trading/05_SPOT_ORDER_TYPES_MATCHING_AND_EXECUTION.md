# Spot Order Types, Matching Engine & Execution Rules

## 1. Overview & Execution Engine Architecture

The Growww / NBSE matching engine (`services/matching-engine`) is an ultra-low latency, in-memory execution system written in Rust 2021. It executes orders with strict deterministic **Price-Time Priority (FIFO)** at sub-15 microsecond latency, backed by a memory-mapped write-ahead log (WAL) and hot-warm shadow failover.

---

## 2. Supported Spot Order Types & Behavioral Specifications

### 2.1 Standard Limit Orders
- **Definition:** An order to buy or sell a specified quantity at a specified limit price or better.
- **Queue Priority:** Strict Price-Time priority. Buy orders with higher prices and Sell orders with lower prices take priority. For equal prices, earlier arrival takes priority.
- **In-Place Size Reduction (TRD-06):** If a trader reduces open order quantity ($Q_{new} < Q_{old}$), the order remains at its exact current queue position without loss of priority. Increasing quantity forces the order to the back of the queue.

### 2.2 Market Orders
- **Definition:** An order to execute immediately against resting liquidity at the best available prices in the order book.
- **Execution Modes:** Immediate-or-Cancel (IOC) or Fill-or-Kill (FOK).
- **Mandatory Slippage Guard:** To protect retail users from predatory sweeps in thin books, all market orders carry an automatic slippage collar (default +/-1.0%). Any portion of the order that would execute outside the collar is cancelled immediately.

### 2.3 Stop-Loss & Stop-Limit Orders
- **Stop-Loss Market:** Inactive until the market Last Traded Price (LTP) breaches the trigger price, at which point it converts to an aggressive Market Order.
- **Stop-Limit:** Inactive until LTP breaches the trigger price, at which point it enters the order book as a standard Limit Order at the specified limit price.
- **Sliding Watermark:** Prevents premature stop triggers caused by momentary single-print wicks.

### 2.4 Trailing Stop Orders
- **Definition:** A stop order whose trigger price dynamically tracks favorable market moves by a fixed point value or percentage offset.
- **Long Position Trailing Stop:** Trigger price moves upward as the market price reaches new highs. If the price falls by the trailing delta, the order activates.

### 2.5 Iceberg & Algorithmic Slicing Orders
- **Definition:** An institutional order where only a small display quantity ($Q_{display}$) is visible in the public Level-2 order book, while the remaining reserve ($Q_{reserve}$) is hidden.
- **Replenishment Rules:** When the display quantity is filled, the engine automatically replenishes the display quantity from reserve with a randomized delay (50ms to 200ms) and minor size jitter (+/-10%) to prevent algorithmic detection.

---

## 3. Market Protection & Fairness Mechanisms

### 3.1 Self-Trade Prevention (STP)
Prevents unintentional matching between orders placed by the same beneficial owner (matching PAN or Account Group). Supports 4 standard modes:
1. **Cancel Aggressive (default):** Incoming aggressive order is cancelled; resting passive order remains.
2. **Cancel Passive:** Resting passive order is cancelled; incoming aggressive order enters the book.
3. **Cancel Both:** Both orders are cancelled immediately.
4. **Decrement and Cancel:** The larger order is decremented by the smaller order size, and the smaller order is cancelled.

### 3.2 The 500-Microsecond Asymmetric Speed Bump Guard
To protect retail investors and resting market maker quotes from predatory high-frequency trading (HFT) latency arbitrage:
- **Aggressive Ingress Orders:** Delayed by exactly 500 microseconds before entering the matching engine queue.
- **Passive Cancellations & Limit Orders:** Ingested with 0 microsecond delay.
- This asymmetry ensures market makers can cancel or re-quote during market shifts before aggressive latency-arbitrage bots can pick them off.
