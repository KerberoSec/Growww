# 620 - Next.js 14 BTC/USDT Pro-Trading Terminal & Live Order Book Dashboard

## Purpose
BTC/USDT represents the bedrock trading pair for digital asset price discovery, liquidity formation, and market depth within the Growww institutional and retail financial ecosystem. Operating under dual domestic and cross-border regulatory frameworks (including the International Financial Services Centres Authority - IFSCA at GIFT City and applicable global virtual digital asset guidelines), professional traders, arbitrageurs, and active retail market participants demand an institutional-grade, zero-latency desktop trading terminal. Standard retail interfaces suffer from document object model (DOM) thrashing, excessive React component tree re-renders, slow WebSocket serialization, and clunky mouse-driven execution tickets that introduce costly execution slippage during high-volatility market events.

This prompt specifies the end-to-end architecture, frontend interface engineering, high-throughput streaming pipelines, state management, hotkey execution engine, and on-chain verification mechanisms for the **Next.js 14 BTC/USDT Pro-Trading Terminal & Live Order Book Dashboard** (`apps/growww_web/trade/btc-usdt`). The system delivers an ultra-fast trading cockpit combining the full-featured TradingView Advanced Charts library, a hardware-accelerated streaming Level-2 order book depth ladder, a real-time time and sales trade feed, single-keystroke hotkey order dispatching with strict safety collars, seamless toggling between Demo (Paper) and Real trading modes, and cryptographic Delivery-versus-Payment (DvP) settlement inspection on Hyperledger Besu.

## What You Are Building
A high-performance, modular trading terminal located in `apps/growww_web/trade/btc-usdt` (and App Router route group `apps/growww_web/app/(trade)/trade/btc-usdt/`) engineered with:
- `TradingCockpitGrid`: A fully responsive, multi-column CSS Grid workstation containing dockable and resizable quadrants: Top Market Ticker Ribbon, Main Chart Viewport, Level-2 Order Book Depth Ladder, Live Trade Feed Tape, Fast Order Execution Panel, and Bottom Account & Settlement Blotter.
- `TradingViewAdvancedChart`: Deep integration with the TradingView Advanced Charts (Charting Library) featuring institutional drawing tools, 100+ technical indicators, multi-timeframe resolution switching (1s, 1m, 5m, 15m, 1h, 4h, 1D, 1W), and a custom WebSocket Datafeed adapter streaming live candlestick updates with sub-millisecond precision.
- `StreamingOrderBookLadder`: A virtualized Level-2 market depth ladder rendering 50 to 100 price levels of Bids (green) and Asks (red) with dynamic cumulative volume depth bars, Best Bid/Offer (BBO) spread indicators, real-time tick-size grouping (0.1, 0.5, 1, 5, 10 USDT), and visual resting order markers.
- `LiveTradeFeedTape`: High-frequency Time & Sales tape streaming trade executions from Market Feeder (Prompt 272) with tick-by-tick price, size, timestamp, buyer/seller aggression coloring, and whale trade highlighting.
- `FastExecutionOrderEntry`: A high-density order entry panel supporting Limit, Market, Stop-Limit, Stop-Market, and OCO (One-Cancels-the-Other) orders, featuring percentage slider selectors (25%, 50%, 75%, 100%), leverage/margin configuration, post-only / reduce-only toggles, and pre-trade fee previews.
- `ModeSwitcher`: Instantaneous, state-isolated toggle between "Demo Mode" (paper trading sandbox with simulated virtual balances, mock matching fills, and risk-free experimentation) and "Real Mode" (live capital execution interacting with Spot Order Service Prompt 275 and collateral balances from Wallet Service Prompt 203).
- `HotkeyExecutionSubsystem`: Low-latency keyboard shortcut system supporting single-stroke and two-stroke order injection (e.g., Space = Buy Market, Shift+Space = Sell Market, Shift+B = Limit Buy at BBO, Escape = Emergency Cancel All Orders) with focus isolation, accidental trade safeguards, and interactive HUD cheat-sheet.
- `BottomExecutionBlotter`: Multi-tabbed blotter displaying Open Orders, Order History, Trade Fills, Positions/Holdings, and On-Chain Settlement Proofs with live links to the Hyperledger Besu Consortium Block Explorer.

## Scope Boundaries
- **In Scope:**
  - Next.js 14 App Router application structure under `apps/growww_web/trade/btc-usdt`.
  - Multi-column grid layout with customizable viewports and panel collapsing.
  - TradingView Advanced Charts library wrapper and custom JavaScript Datafeed (`onReady`, `resolveSymbol`, `getBars`, `subscribeBars`, `unsubscribeBars`).
  - Streaming Level-2 order book depth ladder with `@tanstack/react-virtual` virtualization and requestAnimationFrame batching.
  - Duplex WebSocket integration with Market Feeder (Prompt 272) and Depth Broadcaster (Prompt 276).
  - Rapid order entry interface dispatching to Spot Order Service (Prompt 275).
  - Configurable hotkey management with keyboard focus trapping, safety price deviation collars, and max size limits.
  - Demo (Paper) vs Real trading mode execution and storage isolation.
  - Blotter tabs for active orders, execution history, account balances, and Besu transaction explorer verification links.
  - Playwright end-to-end integration and Vitest unit testing suites.
- **Out of Scope / Handled Elsewhere:**
  - Spot matching engine core order book matching algorithms (Prompt 205, Prompt 275).
  - Low-level UDP market data multicast ingestion and dissemination gateway (Prompt 272).
  - L2 order book delta aggregation and snapshot generation microservice (Prompt 276).
  - Fiat bank transfer rails, UPI/IMPS deposits, and crypto on-ramp custody (Prompt 212, Prompt 213).
  - Smart contract DvP atomic settlement protocol on Besu (Prompt 306).
  - Institutional FIX protocol gateway session handling (Prompt 225).

## Technology to Use
- **Frontend Framework:** Next.js 14 App Router utilizing React Server Components (RSC) for initial page hydration and dynamic `'use client'` sandboxes for real-time canvas and WebSocket pipelines.
- **Language:** TypeScript 5.4+ with strict null checks, zero `any` assertions, and exact type unions for market states and order parameters.
- **Styling & UI Components:** Tailwind CSS 3.4+ configured with high-contrast institutional dark mode variables (Dark Obsidian `#080B11`, Surface Navy `#0F141F`, Border Slate `#1E2638`, Bid Green `#00C087`, Ask Red `#F6465D`, Accent Amber `#F0B90B`), Radix UI headless accessible primitives, and Lucide React icons.
- **Charting Engine:** TradingView Advanced Charts (standalone charting library) integrated via HTML5 iframe/canvas bridge with custom JS Datafeed implementation.
- **DOM Virtualization:** `@tanstack/react-virtual` (v3.x) for Level-2 depth ladder and trade feed virtualization, rendering only visible rows to sustain 60 FPS without memory leaks.
- **High-Frequency State Management:** Zustand with Immer middleware and external vanilla listeners, allowing WebSocket frames to update chart and ladder buffers outside the standard React reconciliation loop.
- **High-Precision Arithmetic:** Decimal.js for BTC (8 decimal places) and USDT (2 decimal places) calculations, preventing IEEE 754 floating-point rounding errors in order prices, cumulative volumes, and fee previews.
- **Form Validation & Hotkeys:** React Hook Form paired with Zod schemas for order parameters, and a custom `useHotkeys` engine utilizing the native `KeyboardEvent` API with debouncing and focus trapping.

## Backend / Infra Touchpoints
- **Spot Order Service (Prompt 275):**
  - REST/gRPC-Web endpoints:
    - `POST /api/v1/spot/orders`: Dispatches Limit, Market, Stop-Limit, and OCO orders with unique client order IDs (`clientOrderId`) and idempotency keys.
    - `DELETE /api/v1/spot/orders/{orderId}`: Cancels an active resting order.
    - `DELETE /api/v1/spot/orders/all`: Emergency cancel-all for the BTC/USDT symbol.
    - `GET /api/v1/spot/orders/active?symbol=BTCUSDT`: Hydrates active resting orders on initial load.
- **Market Feeder (Prompt 272):**
  - Duplex WebSocket connection (`wss://marketdata.growww.in/ws/v1/market`):
    - Subscribes to `trade:BTCUSDT` for tick-by-tick real-time match execution records.
    - Subscribes to `ticker:BTCUSDT` for 24-hour rolling statistics (24h High, 24h Low, 24h Volume, 24h Price Change Percentage).
    - Subscribes to `kline:BTCUSDT:{resolution}` for candlestick bar updates.
- **Depth Broadcaster (Prompt 276):**
  - High-throughput WebSocket connection (`wss://marketdata.growww.in/ws/v1/depth`):
    - Subscribes to `depth:BTCUSDT:l2_snapshot_100ms` for baseline order book recovery.
    - Subscribes to `depth:BTCUSDT:l2_delta` for incremental delta updates (sequence-numbered price level mutations).
- **Wallet & Account Ledger Service (Prompt 203):**
  - Fetches real-time available and reserved balances for BTC and USDT (`GET /api/v1/wallet/balances`).
  - Pre-allocates balance holds during order entry to prevent balance overdrafts.
- **API Gateway & BFF (Prompt 219):**
  - Manages session authentication via HttpOnly JWT tokens, CORS enforcement, and request throttling.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu Consortium Blockchain (QBFT consensus, 2-second block intervals, deterministic zero-reorg finality).
- **Settlement Architecture:**
  - In Real Mode, every matched trade executed by the Spot Matching Engine is submitted to `SpotSettlementDvP.sol` on Hyperledger Besu for immutable Delivery-versus-Payment asset transfer between buyer and seller custodial accounts.
  - The trading terminal subscribes to settlement confirmation events via the Backend BFF and extracts the Besu transaction hash (`txHash`), block number, and gas execution receipt.
- **Explorer Verification:**
  - The Bottom Blotter (Settlement tab) and trade fill toast notifications render direct hyperlinks to the internal Besu Explorer:
    `https://explorer.growww.in/tx/{txHash}`
  - Clicking the explorer link displays verified on-chain details: Block Height, QBFT Validator Signatures, Timestamp, Encrypted Buyer/Seller Demat IDs, and Executed Quantities.
- **Pseudonymity & Compliance:**
  - No Personally Identifiable Information (PII) is committed to the blockchain. All on-chain events reference consortium-approved Ethereum addresses mapped to internal KYC-verified account identifiers in compliance with DPDP Act 2023 regulations.

## Step-by-Step Build Instructions

1. **Scaffold Directory Structure & Terminal Layout Architecture:**
   - Create the route directory at `apps/growww_web/trade/btc-usdt/` and route group `apps/growww_web/app/(trade)/trade/btc-usdt/`.
   - Create layout shell `apps/growww_web/app/(trade)/layout.tsx` providing dark theme configuration, top navigation bar, WebSocket connection status, and mode switcher.
   - Configure multi-column CSS Grid in `apps/growww_web/app/(trade)/trade/btc-usdt/page.tsx` dividing the screen into header ribbon, chart quadrant, order book column, trade tape, order execution panel, and bottom blotter dock.

2. **Define TypeScript Domain Types & State Stores:**
   - Establish typed data models in `apps/growww_web/trade/btc-usdt/types/market.ts` for ticker metrics, order book levels, delta updates, trade records, and order submission payloads.
   - Implement Zustand store in `apps/growww_web/trade/btc-usdt/stores/useMarketStore.ts` managing order book state, recent trades, and 24h ticker metrics.
   - Implement Zustand store in `apps/growww_web/trade/btc-usdt/stores/useOrderEntryStore.ts` managing order forms, hotkey settings, active tabs, and Demo vs Real mode states.

3. **Build Top Market Ticker Ribbon Component (`MarketHeaderRibbon.tsx`):**
   - Render active pair symbol `BTC/USDT`, 24h Mark Price, 24h Index Price, 24h Percentage Change (with green/red indicator), 24h High, 24h Low, and 24h Volume in BTC and USDT.
   - Connect to `useMarketStore` to update ticker metrics without re-rendering adjacent chart or order entry components.
   - Embed Demo Mode vs Real Mode badge indicator and toggle switch.

4. **Integrate TradingView Advanced Charts Library (`TradingViewChart.tsx`):**
   - Load the TradingView Advanced Charts library bundle into `apps/growww_web/public/static/charting_library/`.
   - Implement custom JavaScript Datafeed adapter (`apps/growww_web/trade/btc-usdt/lib/tv-datafeed.ts`):
     - `onReady`: Exposes supported resolutions (1s, 1m, 5m, 15m, 1h, 4h, 1D, 1W) and symbol configuration.
     - `resolveSymbol`: Resolves `BTC/USDT` specifications (pricescale: 100, minmov: 1, timezone: 'Asia/Kolkata').
     - `getBars`: Queries historical OHLCV data from Market Feeder REST API (`GET /api/v1/market/klines`).
     - `subscribeBars`: Subscribes to real-time kline updates from Market Feeder WebSocket (`kline:BTCUSDT:{resolution}`).
     - `unsubscribeBars`: Cleans up active kline subscriptions upon resolution or symbol change.
   - Mount the chart within an iframe/container with responsive resize observer support.

5. **Build Streaming Level-2 Order Book Depth Ladder (`OrderBookLadder.tsx`):**
   - Implement virtualized dual-column ladder displaying Asks in descending price order at the top, Bids in descending price order at the bottom, and the dynamic BBO spread ribbon in the center.
   - Render horizontal depth bars in the background of each row representing the cumulative volume proportion relative to the maximum depth tier.
   - Implement tick-size aggregation dropdown (0.1, 0.5, 1, 5, 10 USDT) to dynamically bucket adjacent price levels.
   - Add click-to-fill handler on any price row to automatically populate the order entry ticket price.
   - Display distinct visual markers (e.g., small amber dots or badges) next to price levels where the user has active resting limit orders.

6. **Implement Depth Broadcaster (Prompt 276) WebSocket Integration (`useDepthStream.ts`):**
   - Establish resilient WebSocket connection to `wss://marketdata.growww.in/ws/v1/depth`.
   - Implement initial snapshot hydration: Fetch complete 100-level snapshot on connection open.
   - Implement sequence-checked incremental delta buffer:
     - Verify incoming delta sequence IDs (`lastUpdateId`) match local book state.
     - If sequence gap is detected, discard buffer, trigger snapshot re-fetch, and re-apply buffered deltas.
   - Throttle UI updates to maximum 50ms intervals using `requestAnimationFrame` and mutable buffer staging to prevent React rendering churn.

7. **Construct Live Trade Feed (Time & Sales Tape) (`LiveTradeFeed.tsx`):**
   - Connect to Market Feeder WebSocket (`trade:BTCUSDT`) to receive real-time execution matches.
   - Build virtualized vertical list rendering Price, Size (BTC), Time (HH:mm:ss.SSS), and Aggressor Side (Buy/Sell).
   - Format row text color based on trade side (Green for buyer-maker, Red for seller-maker).
   - Add animated flash effect for high-notional "whale" trades exceeding 5.0 BTC.

8. **Develop Fast Execution Order Entry Panel (`OrderEntryPanel.tsx`):**
   - Build dual Buy (Long) / Sell (Short) tab interface with distinct emerald and crimson button styling.
   - Implement order type selector: Limit, Market, Stop-Limit, Stop-Market, and OCO.
   - Integrate numeric stepper inputs for Price and Amount with quick-fill percentage buttons (25%, 50%, 75%, 100% of available collateral).
   - Implement interactive slider for rapid fractional position sizing.
   - Add checkboxes for Post-Only (Maker) and Reduce-Only execution constraints.
   - Compute real-time pre-trade fee estimate (0.00% (Zero Fee) standard taker / 0.00% maker fee (Zero Brokerage)) and estimated total notional in USDT using Decimal.js.

9. **Implement Demo Mode vs Real Mode Execution Isolation:**
   - In **Demo Mode**:
     - Maintain simulated virtual portfolio state in local storage and memory (e.g., initial balance of 100,000 virtual USDT and 2 virtual BTC).
     - Order submissions generate synthetic client fills by matching against live BBO prices from the order book.
     - Synthetic fills generate mock trade receipts without dispatching network calls to the live matching engine.
   - In **Real Mode**:
     - Validate active user session and collateral balance via Wallet Service (Prompt 203).
     - Dispatch authenticated order requests to Spot Order Service (Prompt 275) with unique UUIDv4 idempotency keys.
     - Display live execution reports and order status transitions (`NEW`, `PARTIALLY_FILLED`, `FILLED`, `CANCELED`, `REJECTED`).

10. **Build Low-Latency Hotkey Execution Subsystem (`useTradingHotkeys.ts`):**
    - Implement global keyboard listener listening for configured trading hotkeys:
      - `Space`: Execute Market Buy with configured default size.
      - `Shift + Space`: Execute Market Sell with configured default size.
      - `Shift + B`: Populate Limit Buy at Best Bid.
      - `Shift + S`: Populate Limit Sell at Best Ask.
      - `Escape`: Cancel All Open Orders immediately.
      - `1` / `2` / `3` / `4`: Select 25% / 50% / 75% / 100% capital sizing.
    - Implement safety collars:
      - Require two-stroke confirmation or modifier combination for market orders.
      - Check price deviation collar: Reject orders deviating >3% from the current BBO.
      - Check maximum notional size limit (configurable by user, default cap of 2 BTC).
    - Disable hotkey execution automatically whenever the user is actively focused on input fields or text areas.
    - Build modal Hotkey Cheatsheet HUD accessible via `?` key.

11. **Construct Bottom Execution Blotter (`TerminalBlotter.tsx`):**
    - Implement tabbed dock featuring:
      - **Open Orders Tab**: Virtualized table of resting orders with Order ID, Side, Type, Price, Amount, Filled %, Time, and Cancel button.
      - **Order History Tab**: Searchable history of canceled, filled, and expired orders.
      - **Trade History Tab**: Detailed log of executed fills with execution fees and realized P&L.
      - **Positions / Balances Tab**: Current wallet holdings for BTC and USDT, margin allocation, and unrealized value.
      - **On-Chain Settlement Tab**: Displays Besu block height, transaction hash with clickable link to Besu Explorer (`https://explorer.growww.in/tx/{txHash}`), atomic DvP status badge, and cryptographic settlement timestamp.

12. **End-to-End Testing, Performance Profiling, and Verification:**
    - Conduct Vitest unit tests verifying Decimal.js precision, order book delta patching, tick aggregation rounding, and hotkey collar checks.
    - Benchmark order book DOM rendering to guarantee sub-50ms paint times during synthetic burst loads of 5,000 WebSocket updates per second.
    - Implement Playwright E2E test suite covering:
      - Switching between Demo and Real modes.
      - Placing Limit and Market orders via mouse click and hotkey combinations.
      - Canceling individual and all open orders via `Escape` key.
      - Verifying Besu Explorer URL generation on trade settlement.

## Interfaces / Contracts

### TypeScript Domain Models (`market.ts`)
```typescript
export type TradeSide = 'BUY' | 'SELL';

export type OrderType = 'LIMIT' | 'MARKET' | 'STOP_LIMIT' | 'STOP_MARKET' | 'OCO';

export type TimeInForce = 'GTC' | 'IOC' | 'FOK';

export type OrderStatus = 
  | 'PENDING' 
  | 'NEW' 
  | 'PARTIALLY_FILLED' 
  | 'FILLED' 
  | 'CANCELED' 
  | 'REJECTED' 
  | 'EXPIRED';

export interface MarketTicker {
  symbol: string;
  lastPrice: string;
  indexPrice: string;
  markPrice: string;
  priceChange24h: string;
  priceChangePercent24h: string;
  highPrice24h: string;
  lowPrice24h: string;
  volume24hBtc: string;
  volume24hUsdt: string;
  timestamp: number;
}

export interface OrderBookLevel {
  price: string;
  amount: string;
  total: string;
  depthPercent: number;
  myOrdersCount: number;
}

export interface OrderBookSnapshot {
  symbol: string;
  lastUpdateId: number;
  bids: [string, string][]; // [price, amount]
  asks: [string, string][]; // [price, amount]
  timestamp: number;
}

export interface DepthDeltaMessage {
  event: 'depthUpdate';
  symbol: string;
  firstUpdateId: number;
  lastUpdateId: number;
  bids: [string, string][]; // [price, amount] - amount '0' means remove level
  asks: [string, string][];
  timestamp: number;
}

export interface TradeExecutionMessage {
  event: 'trade';
  symbol: string;
  tradeId: string;
  price: string;
  amount: string;
  side: TradeSide;
  isWhale: boolean;
  buyerIsMaker: boolean;
  timestamp: number;
}

export interface SpotOrderSubmission {
  clientOrderId: string;
  symbol: 'BTCUSDT';
  side: TradeSide;
  orderType: OrderType;
  timeInForce: TimeInForce;
  price?: string;
  stopPrice?: string;
  amount: string;
  postOnly?: boolean;
  reduceOnly?: boolean;
  isDemo: boolean;
}

export interface SpotOrderRecord {
  orderId: string;
  clientOrderId: string;
  symbol: string;
  side: TradeSide;
  orderType: OrderType;
  price: string;
  amount: string;
  executedAmount: string;
  cumulativeQuote: string;
  status: OrderStatus;
  feeAmount: string;
  feeAsset: string;
  createdAt: number;
  updatedAt: number;
  besuTxHash?: `0x${string}`;
}

export interface SettlementReceipt {
  tradeId: string;
  orderId: string;
  symbol: string;
  txHash: `0x${string}`;
  blockNumber: number;
  buyerAddress: `0x${string}`;
  sellerAddress: `0x${string}`;
  btcAmount: string;
  usdtAmount: string;
  gasUsed: string;
  settledAt: string;
}
```

### WebSocket Client Event Schemas

#### Client Subscription Payloads
```json
{
  "action": "subscribe",
  "channels": [
    "ticker:BTCUSDT",
    "trade:BTCUSDT",
    "depth:BTCUSDT:l2_delta",
    "kline:BTCUSDT:1m"
  ],
  "clientTimestamp": 1726738083000
}
```

#### Inbound Trade Feed Event (`trade:BTCUSDT`)
```json
{
  "event": "trade",
  "symbol": "BTCUSDT",
  "tradeId": "984201948",
  "price": "64250.50",
  "amount": "0.45210000",
  "side": "BUY",
  "isWhale": false,
  "buyerIsMaker": false,
  "timestamp": 1726738083120
}
```

#### Inbound Level-2 Delta Event (`depth:BTCUSDT:l2_delta`)
```json
{
  "event": "depthUpdate",
  "symbol": "BTCUSDT",
  "firstUpdateId": 48920191,
  "lastUpdateId": 48920198,
  "bids": [
    ["64250.00", "1.25000000"],
    ["64249.50", "0.00000000"]
  ],
  "asks": [
    ["64250.50", "0.85000000"],
    ["64251.00", "2.10000000"]
  ],
  "timestamp": 1726738083150
}
```

### REST Endpoints: Spot Order Service (Prompt 275)
- `POST /api/v1/spot/orders`
  - Request: `SpotOrderSubmission`
  - Response: `{ success: true, order: SpotOrderRecord }`
- `DELETE /api/v1/spot/orders/{orderId}`
  - Request: Empty (Header carries Auth JWT and Idempotency Key)
  - Response: `{ success: true, canceledOrderId: string, status: "CANCELED" }`
- `DELETE /api/v1/spot/orders/all?symbol=BTCUSDT`
  - Request: Empty
  - Response: `{ success: true, canceledCount: number }`
- `GET /api/v1/spot/orders/active?symbol=BTCUSDT`
  - Response: `{ orders: SpotOrderRecord[] }`

## Security & Compliance Notes

### Sub-50ms DOM Updates & Re-Render Elimination
- High-frequency market data updates must never cause full React component tree re-renders. The terminal employs:
  1. **Zustand Vanilla Stores:** Updates to the order book and trade tape are committed to non-reactive memory buffers.
  2. **`requestAnimationFrame` Throttling:** Incoming WebSocket deltas are accumulated in a ring-buffer and painted to the screen at a synchronized 60 FPS (16.6ms cadence) or throttled 50ms batching, preventing UI freezes during high-volume volatility.
  3. **Virtual DOM Bypass & DOM Node Reuse:** The order book ladder and trade feed employ `@tanstack/react-virtual` with fixed row heights and `will-change: transform` CSS properties, bounding active DOM elements to fewer than 60 nodes regardless of order book depth.

### Secure Session & Transaction Authentication
- **HttpOnly Cookies:** Authentication tokens (access JWT and refresh tokens) are managed exclusively through HttpOnly, Secure, SameSite=Strict cookies to prevent cross-site scripting (XSS) extraction.
- **Anti-CSRF Tokens:** All mutating order entry endpoints require a cryptographic CSRF token passed via custom HTTP request headers (`X-Growww-CSRF-Token`).
- **Idempotency Safeguards:** Every order submission payload includes a UUIDv4 `clientOrderId`. The Spot Order Service rejects duplicate order IDs within a 24-hour sliding window to eliminate double-fill risks from network retries.

### Hotkey Accidental Order Safeguards
- Single-keystroke trade execution introduces operational risk. The terminal enforces multi-tiered safety collars:
  1. **Input Focus Masking:** Keystroke events are strictly ignored whenever focus is within an `HTMLInputElement`, `HTMLTextAreaElement`, or active dropdown.
  2. **Price Deviation Collar:** Orders with limit prices deviating more than 3.0% from the current BBO are blocked and display an alert toast ("Order rejected: Price deviates >3% from market").
  3. **Maximum Order Notional Cap:** Keystroke-initiated orders enforce an un-overrideable single-ticket ceiling (default: 2.0 BTC or 150,000 USDT).
  4. **Panic Killswitch (`Escape`):** The emergency cancel-all hotkey is debounced by 200ms to prevent double-firing and immediately dispatches an asynchronous batch cancel request to the Spot Order Service while optimistically purging resting order markers from the UI.

### On-Chain DvP Integrity & Zero PII Exposure
- Every completed spot match on the live exchange is cryptographically notarized on Hyperledger Besu via `SpotSettlementDvP.sol`.
- In strict adherence to DPDP Act 2023 and international financial privacy frameworks, zero PII (such as legal names, emails, or phone numbers) is recorded on the distributed ledger. Explorer links exclusively reference pseudonymous contract addresses and hashed trade identifiers.

## Acceptance Criteria
- [ ] Next.js 14 application compiles and runs without any TypeScript errors, ESLint warnings, or hydration mismatches under `apps/growww_web/trade/btc-usdt`.
- [ ] Multi-column responsive CSS Grid terminal layout mounts cleanly with dockable quadrants for Chart, Order Book, Tape, Order Entry, and Blotter.
- [ ] TradingView Advanced Charts library initializes with custom Datafeed and correctly streams real-time candlestick bars from Market Feeder (Prompt 272).
- [ ] Level-2 Order Book ladder hydrates from 100-level snapshot and applies incremental deltas from Depth Broadcaster (Prompt 276) with sub-50ms paint latency.
- [ ] Tick-size aggregation selector dynamically buckets order book levels into 0.1, 0.5, 1, 5, and 10 USDT groupings without data loss.
- [ ] Live Trade Feed renders recent trades in real time with correct buyer/seller aggression colors and whale trade indicators.
- [ ] Fast Execution Order Entry supports Limit, Market, Stop-Limit, and OCO order submissions with valid Decimal.js fee and total calculations.
- [ ] Mode Switcher toggles cleanly between Demo Mode (local paper trading simulation) and Real Mode (live network order routing) without cross-state data leaks.
- [ ] Hotkey engine processes configured keyboard shortcuts (`Space`, `Shift+Space`, `Shift+B`, `Shift+S`, `Escape`) with focus isolation and safety collars.
- [ ] Orders violating the 3% price deviation collar or maximum notional size limit are rejected with clear user notifications.
- [ ] Bottom Blotter displays active orders, trade history, positions, and valid on-chain Hyperledger Besu Explorer links (`https://explorer.growww.in/tx/{txHash}`).
- [ ] Performance profiling verifies 60 FPS rendering under high-load synthetic bursts of 5,000 WebSocket updates per second with zero UI frame drops.
- [ ] Vitest unit test suite passes with 100% coverage on order book delta patching, Decimal.js calculations, and hotkey collar rules.
- [ ] Playwright E2E suite validates full user flow: mode switching, order placement, order book update, hotkey emergency cancel, and settlement explorer verification.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 203: Wallet & Double-Entry Account Ledger Service (for balance reservation and fund holds).
  - Prompt 219: API Gateway & BFF (for authenticated session routing and rate limiting).
  - Prompt 272: Market Feeder Service (for WebSocket ticker, trade tape, and kline streaming).
  - Prompt 275: Spot Order Service (for core spot order lifecycle, matching routing, and order blotter state).
  - Prompt 276: Depth Broadcaster Service (for L2 order book delta streams and snapshot dissemination).
  - Prompt 306: Smart Contract DvP Atomic Settlement (for on-chain Besu settlement transactions).
  - Prompt 601: Next.js 14 Investor Web App Scaffolding (for root app layout, design system tokens, and shared UI primitives).
- **Parallel Tasks:**
  - Prompt 603: Web Trading Terminal & Lightweight Charts Integration (for fractional equity trading terminal).
  - Prompt 612: Web Institutional Direct Market Access (DMA) Workstation (for institutional hotkey and FIX terminal).
- **Downstream Blockers:**
  - Prompt 906: Comprehensive UAT Plan & Regulatory Sandbox Scenarios.
  - Prompt 907: Regulatory Sandbox Pilot Launch Plan.
