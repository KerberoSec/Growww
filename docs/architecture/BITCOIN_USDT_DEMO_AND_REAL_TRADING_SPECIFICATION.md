# Bitcoin (BTC/USDT) Demo & Real-Money Trading Master Architecture Specification

## 1. Executive Summary & Vision

The Growww / NBSE platform implements a sovereign, institutional-grade trading infrastructure centered on the **BTC/USDT** trading pair. The system provides two fully segregated execution modes operating over the identical, real-time live Bitcoin market price:

1. **Demo / Paper Trading Mode (Testnet - Chain ID 13371):**
   - Enables retail and institutional users to practice BTC/USDT spot and algorithmic trading without financial risk.
   - Trades execute in lockstep with real-world Bitcoin spot market prices streamed from Tier-1 global exchanges.
   - Includes a 1-click testnet faucet dispensing 10,000 virtual USDT and 1.0 virtual BTC, portfolio reset capabilities, and realistic simulated fill models (with synthetic book depth and realistic slippage).
2. **Real-Money Trading Mode (Mainnet - Chain ID 2026):**
   - High-performance, fully audited spot trading with real capital.
   - Direct on-chain deposit/withdrawal ingress for Bitcoin (Native SegWit `bc1q` and Taproot `bc1p`) and USDT (Ethereum ERC-20, Tron TRC-20, and Polygon).
   - In-memory sub-millisecond matching engine execution (Rust) backed by double-entry relational balance accounting (PostgreSQL 16) and atomic Delivery-versus-Payment (DvP) settlement on Hyperledger Besu under QBFT consensus.
   - Enforces the platform's non-negotiable **single flat 0.00% transaction fee (No fee at all)** (zero gas fees, zero hidden spreads).

---

## 2. End-to-End System Architecture

```
                                  +---------------------------------------+
                                  | Global Spot Feeds (Binance, Coinbase, |
                                  | Kraken, OKX) + Chainlink Oracle Feeds |
                                  +---------------------------------------+
                                                      |
                                                      v
                                        +----------------------------+
                                        | services/market-feeder     |
                                        | (Trimmed Median Aggregator)|
                                        +----------------------------+
                                                      |
                          +---------------------------+---------------------------+
                          | Live Ticker & OHLCV Streams (Redis / Kafka)           |
                          v                                                       v
        +-----------------------------------+                   +-----------------------------------+
        | DEMO / PAPER TRADING CLUSTER      |                   | REAL-MONEY TRADING CLUSTER        |
        | (Isolated Testnet Environment)    |                   | (Isolated Mainnet Environment)    |
        +-----------------------------------+                   +-----------------------------------+
        | • services/demo-matching-engine   |                   | • services/matching-engine (Rust) |
        | • services/demo-wallet-service    |                   | • services/btc-order-service (Go) |
        | • Database: nbse_demo_wallet      |                   | • services/wallet-service (Go)    |
        | • Smart Contract: VirtualFaucet   |                   | • Database: nbse_production_core  |
        | • 10,000 vUSDT / 1.0 vBTC Faucet  |                   | • Smart Contract: SettlementDvP   |
        | • Zero Financial Risk             |                   | • MPC-TSS Hot/Cold Vaults         |
        +-----------------------------------+                   +-----------------------------------+
                          |                                                       |
                          +---------------------------+---------------------------+
                                                      |
                                                      v
                                    +-----------------------------------+
                                    | CLIENT APPLICATION LAYER          |
                                    | (Flutter App & Next.js 14 Portal) |
                                    +-----------------------------------+
                                    | • Persistent Mode Switcher Toggle |
                                    | • Live TradingView Chart (BTC)    |
                                    | • Real-Time Level-2 Order Book    |
                                    | • Buy/Sell Order Entry Modal      |
                                    +-----------------------------------+
```

---

## 3. Real-Time External Market Data Feeder (`services/market-feeder`)

To guarantee that both Demo and Real trading reflect genuine market conditions, the platform ingests tick-by-tick market data from multiple independent external venues:

### 3.1 Redundant Ingestion Venues
- **Binance WebSocket:** `wss://stream.binance.com:9443/ws/btcusdt@trade` & `@depth20@100ms`
- **Coinbase Exchange WebSocket:** `wss://ws-feed.exchange.coinbase.com` (`matches` & `level2`)
- **Kraken WebSocket:** `wss://ws.kraken.com/v2` (`trade` & `book`)
- **OKX WebSocket:** `wss://ws.okx.com:8443/ws/v5/public` (`trades` & `books5`)
- **Chainlink Reference Oracle:** Heartbeat-validated on-chain BTC/USD oracle feed on Besu.

### 3.2 Trimmed Median Fair Index Pricing Algorithm
To prevent anomalous price spikes or deliberate manipulation from any single external exchange, the feeder executes a sliding 5-second trimmed median calculation across incoming venue prices:

$$\text{ValidPrices}(t) = \left\{ P_i(t) \;\middle|\; \left| \frac{P_i(t) - P_{median}(t)}{P_{median}(t)} \right| \le 0.0075 \right\}$$

$$\text{FairIndexPrice}(t) = \frac{\sum_{i \in \text{ValidPrices}} w_i \cdot P_i(t)}{\sum_{i \in \text{ValidPrices}} w_i}$$

Where $w_i$ represents the 24-hour volume weight of exchange venue $i$.

### 3.3 Candlestick (OHLCV) Aggregator
- **Intervals:** 1-second ticks, 1-minute, 5-minute, 15-minute, 1-hour, 4-hour, and 1-day bars.
- **Sliding Watermark:** Handles late-arriving trade ticks with a 1500ms sliding revision window (`CANDLE_REVISE`) without shifting historical bar timestamps.
- **Dissemination:** Broadcasts over Redis Pub/Sub channel `market:btc_usdt:ticker` and Kafka topic `growww.market.btc_usdt.ticks.v1`.

---

## 4. Demo / Paper Trading Engine Architecture (`services/demo-matching-engine`)

The demo trading subsystem gives users an authentic trading experience without risking real capital:

### 4.1 Faucet & Virtual Wallet Ledger (`services/demo-wallet-service`)
- **Initial Faucet Allocation:** Upon activating Demo Trading mode, each verified user receives:
  - `10,000.00 vUSDT` (Virtual Tether USD)
  - `1.00000000 vBTC` (Virtual Bitcoin)
- **1-Click Balance Reset:** Users can reset their virtual portfolio back to initial faucet balances at any time with a 24-hour rate limit.
- **Database Isolation:** All virtual transactions are recorded in the dedicated database `nbse_demo_wallet`, completely separated from the production double-entry ledger.

### 4.2 Simulated Execution Model
The demo matching engine simulates realistic order execution against the live market price:
- **Market Orders:** Filled instantaneously at the current `FairIndexPrice` plus a realistic synthetic slippage model calibrated to order size:
  $$\text{FillPrice}_{buy} = P_{ask} \times \left(1 + \frac{\text{OrderSize}}{20.0} \times 0.0005\right)$$
- **Limit Orders:** Resting orders are placed in the virtual order book. A limit buy is triggered whenever the live external market price touches or drops below the limit price. A limit sell is triggered when the live price touches or rises above the limit price.
- **Simulated Fees:** Deducts the platform standard **0.00% fee (No fee at all)** in virtual currency, training the user on fee mechanics.

---

## 5. Real-Money Trading Architecture (`services/btc-order-service`)

When users switch to Real-Money Trading mode, execution switches to the production exchange core:

### 5.1 Real BTC & USDT Ingress
- **Bitcoin (BTC) Ingress (`services/btc-deposit-listener`):**
  - Generates dedicated Native SegWit (`bc1q...`) and Taproot (`bc1p...`) deposit addresses via FIPS 140-2 Level 3 HSM master xpub keys.
  - Monitors the Bitcoin mempool and blocks. Displays 0-confirmation "Pending" state in the UI.
  - Automatically credits the user's production ledger balance upon **3 block confirmations** (6 confirmations for amounts $\ge 2.0\text{ BTC}$).
  - Mints 1:1 custodial backed `wBTC` on Hyperledger Besu.
- **USDT Ingress (`services/usdt-deposit-service`):**
  - Supports Ethereum (ERC-20), Tron (TRC-20), and Polygon networks.
  - Verifies smart contract transaction events and initiates automated sweeps to MPC-TSS cold storage.
  - Mints 1:1 backed `weUSDT` on Hyperledger Besu.

### 5.2 Real-Money Spot Order Execution Lifecycle
1. **Validation & Pre-Trade Risk:** The order service verifies that the user's available unencumbered balance covers the order value plus the 0.00% fee (No fee at all).
2. **Pessimistic Balance Reservation:** Uses PostgreSQL `SELECT ... FOR UPDATE` to lock funds, preventing double-spend race conditions.
3. **Low-Latency Sequencer Ingress:** The order enters the single-writer sequencer ring buffer and is matched by the Rust matching engine (`matching-engine`) in $< 15\mu\text{s}$.
4. **DvP Settlement Handoff:** Trade fills are batched and dispatched to `SettlementDvP.sol` on Hyperledger Besu.
5. **Fee Allocation:**
   - 60% to Corporate Treasury
   - 25% to Settlement Guarantee Fund (SGF)
   - 15% to Investor Protection Fund (IPF).

---

## 6. Client Experience & User Interfaces

### 6.1 Flutter Mobile & Desktop App (`apps/growww_flutter`)
- **Persistent Mode Switcher Banner:** A prominent status bar at the top of the trading screen displays either:
  - **AMBER / YELLOW BANNER:** `"DEMO TRADING MODE (VIRTUAL FUNDS) - [RESET] [FAUCET]"`
  - **EMERALD / GREEN BANNER:** `"REAL TRADING MODE (MAINNET LIVE)"`
- **TradingView Candlestick Chart:** Sub-second continuous rendering of Bitcoin candles with volume bars, moving averages (EMA 20/50/200), RSI, and MACD indicators.
- **Order Entry Sheet:** Tabbed interface for Buy (Green) and Sell (Red) orders, quick percentage selector chips (25%, 50%, 75%, 100%), and biometric confirmation before submitting real-money orders.

### 6.2 Next.js 14 Pro Trading Terminal (`apps/growww_web`)
- **Institutional 3-Column Cockpit:**
  - Left: Interactive TradingView advanced charting widget with drawing tools.
  - Center: Live Level-2 streaming order book with bid/ask visual depth bars and recent trade tape.
  - Right: Order placement ticket supporting Limit, Market, and Stop-Limit orders with keyboard hotkey shortcuts (`B` for Buy, `S` for Sell, `Esc` to Cancel).

---

## 7. Security, Invariants & Market Integrity

1. **Strict Balance & Environment Segregation:** Zero shared memory, zero shared database connections, and zero credential reuse between the Demo sandbox and Mainnet production systems.
2. **Universal fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Invariant:** Exactly 0.00% fee (no fee at all) charged on executed trade notional in both Demo and Real modes.
3. **Asymmetric 500μs Speed Bump:** Aggressive incoming orders from external low-latency API connections are delayed by 500 microseconds to eliminate predatory front-running against retail users.
4. **Dynamic Volatility Collar:** If internal matching price diverges by $> 3.0\%$ from the global Fair Index Price, continuous trading pauses automatically and transitions to a 5-minute call auction.
5. **MPC-TSS Cold Storage Protection:** 95%+ of user Bitcoin and USDT reserves are secured in multi-party computation threshold vaults requiring 3-of-5 geographically distributed key shards for withdrawal authorization.
