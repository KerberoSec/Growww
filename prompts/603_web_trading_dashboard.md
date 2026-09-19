# 603 - Web Trading Terminal & Lightweight Charts Integration

## Purpose
Builds an institutional-grade, high-performance desktop web trading terminal (`apps/growww_web/app/(app)/trade/`) offering multi-pane market surveillance, interactive TradingView Lightweight Charts with multi-timeframe OHLCV bars, live Level-2 order book depth ladders, fractional equity order execution ticket (INR amount or fractional unit-based), and real-time portfolio holdings tracking with DvP settlement transparency. Empowers retail and professional investors to trade fractional shares of Indian blue-chip equities with sub-second execution feedback and complete transparency into underlying custody backing.

## What You Are Building
- Multi-pane desktop trading terminal interface (`/trade/[ticker]`) with customizable docked panels.
- TradingView Lightweight Charts (v4.x) integration featuring candlestick charts, volume histograms, moving averages (SMA/EMA), and RSI indicators.
- Real-time Level-2 order book depth ladder with visual bid/ask volume depth bars and spread indicator.
- Interactive order ticket component supporting Market, Limit, and Stop-Loss orders with instant INR fractional calculation and pre-trade fee preview (Universal Zero-Fee Model (0.00% fee - No fee at all) on turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol); zero holding fees; FIFO capital gains for Section 111A/112A tax compliance).
- Recent trades ticker feed and market statistics ribbon (24h high/low, volume, 52-week range, underlying custody ISIN, SEBI depository info).
- Bottom dock containing Open Orders, Order History, Holdings/Positions, and DvP Settlement status tabs.

## Scope Boundaries
- **In Scope:**
 - Multi-panel trading layout, responsive resizing, and keyboard navigation.
 - TradingView Lightweight Charts canvas setup, historical data fetching, and WebSocket candle streaming.
 - Level-2 Order book virtualization rendering 50+ price levels at 60 FPS without DOM lag.
 - Client-side fractional share calculation, limit price validation, and order ticket submission.
 - Real-time WebSocket event handling for order lifecycle updates and settlement confirmations.
- **Out of Scope / Handled Elsewhere:**
 - Order matching engine backend (Prompt 205).
 - Pre-trade risk validation service (Prompt 206).
 - Market data WebSocket streaming backend (Prompt 207).
 - Smart contract DvP settlement execution (Prompt 306).

## Technology to Use
- **Next.js 14, React 18/19, TypeScript 5.4+:** Provides server-rendered initial page shell for instant page loads and client components for interactive canvas rendering.
- **`lightweight-charts` (TradingView v4.x):** Hardware-accelerated HTML5 Canvas financial charting library delivering ultra-smooth 60 FPS pan/zoom and minimal bundle size (<45KB gzipped) compared to full TradingView library.
- **`@tanstack/react-virtual`:** Virtualizes large Level-2 order book depth ladders and trade feeds, rendering only visible DOM elements to eliminate rendering overhead on high-frequency price updates.
- **Zustand with Immer:** High-performance ephemeral state management for order book snapshot and delta patching with zero unnecessary React re-renders.
- **Decimal.js / BigNumber.js:** Executes exact fixed-point mathematical calculations for fractional shares (up to 6 decimal places) and INR fiat amounts (2 decimal places) without IEEE 754 floating-point inaccuracies.

## Backend / Infra Touchpoints
- **Market Data WebSocket Service (Prompt 207):** Subscribes to `wss://api.growww.in/ws/v1/market` for live order book deltas, candle bars, and recent match trades.
- **Order Service (Prompt 204):** Submits orders via REST / gRPC-Web (`POST /api/v1/orders`) with idempotency headers.
- **Portfolio & Holdings Service (Prompt 209):** Fetches real-time fractional unit balances and average purchase costs.
- **Fee & Realized-P&L Engine (Prompt 210):** Queries estimated fixed 0.00% transaction fee (No fee at all) schedule for pre-trade preview.

## Blockchain Interaction
- **DvP Settlement Verification:** Displays real-time on-chain transaction hash and confirmation tag (e.g., "Besu DvP Settled • Block #481920") upon trade matching and settlement execution.
- **Ledger Event Subscription:** Connects to Hyperledger Besu WebSocket RPC to listen for `SettlementDvP.SettlementCompleted(bytes32 indexed orderId, address indexed buyer, address indexed seller, uint256 tokenAmount, uint256 fiatAmount)` events to provide cryptographic proof of settlement.

### Detailed On-Chain Integration Mechanics:
- **Target Network:** Hyperledger Besu (Permissioned Consortium Network with QBFT Consensus).
- **Core Smart Contracts Interfaced:**
 - `DigitalSecurityToken.sol` (Queries token contract address for active ISIN, decimals, and whitelist verification).
 - `SettlementDvP.sol` (Tracks settlement execution and atomic delivery-versus-payment state).
- **Transparency:** The web UI embeds a direct link to the internal Besu Block Explorer (`explorer.growww.in/tx/0x...`) for every matched trade execution.

## Step-by-Step Build Instructions
1. Scaffold `/trade/[ticker]` layout using a CSS Grid container with 4 main quadrants: Market Header, Chart Viewport, Order Book / Depth, and Order Ticket / Bottom Dock.
2. Implement Top Market Ribbon component displaying Symbol, Company Name, Live Price, 24h Change %, 24h High/Low, Total Volume, ISIN, and Custody Depository (NSDL/CDSL).
3. Integrate TradingView Lightweight Charts in the chart viewport with resize observer to maintain aspect ratio dynamically.
4. Implement timeframe switcher (1m, 5m, 15m, 1h, 1D, 1W) that fetches historical OHLCV candles from the Market Data API and sets the chart series.
5. Connect WebSocket candle stream to the chart series via `candlestickSeries.update(candle)` to render live real-time candle ticking.
6. Implement Level-2 Order Book component with `@tanstack/react-virtual` displaying Bids (green) and Asks (red) sorted by price with cumulative volume depth bars.
7. Build WebSocket order book buffer patching snapshots with live deltas (`orderbook_update` events) using a 100ms throttle buffer.
8. Implement Order Placement Ticket supporting Toggle between "INR Amount" (fractional calculation) and "Exact Units".
9. Add Order Type selector (Market, Limit, Stop-Limit) with price stepper inputs and quick-fill percentages (25%, 50%, 75%, 100% of available INR cash).
10. Build pre-trade fee preview calculator displaying estimated statutory charges (STT, Stamp Duty, Exchange charges), the 0.00% platform fee (No fee at all) on trade notional turnover (operating with 0.00% fees at launch (100% net proceeds credited)), reminding user of Growww's zero holding/AUM fee policy, and noting FIFO capital gains tax compliance (Section 111A/112A).
11. Implement Bottom Dock tabs: "Open Orders" (with cancel button), "Executed Trades", "Holdings / Portfolio", and "DvP On-Chain Settlements".
12. Implement toast notification manager alerting users instantly on order placement, partial fill, complete fill, and DvP settlement on Besu ledger.
13. Add keyboard shortcuts (e.g. `B` for Buy, `S` for Sell, `Esc` to close modal, `Up/Down` to switch active ticker) and test with Vitest and Playwright.

## Interfaces / Contracts
```typescript
export interface OrderBookRow {
  price: string; // Decimal string e.g. "2450.50"
  quantity: string; // Fractional quantity e.g. "12.450000"
  totalQuantity: string; // Cumulative depth
  depthPercent: number; // 0 to 100 for visual depth bar
}

export interface CandleStickPoint {
  time: number; // UNIX timestamp in seconds
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface OrderTicketFormValues {
  ticker: string;
  isin: string;
  side: 'BUY' | 'SELL';
  orderType: 'MARKET' | 'LIMIT' | 'STOP_LIMIT';
  inputMode: 'INR_AMOUNT' | 'UNIT_QUANTITY';
  inrAmount?: string;
  unitQuantity?: string;
  limitPrice?: string;
  stopPrice?: string;
}

export interface OrderSubmissionRequest {
  idempotencyKey: string;
  isin: string;
  side: 'BUY' | 'SELL';
  orderType: 'MARKET' | 'LIMIT' | 'STOP_LIMIT';
  quantity: string; // Exact 6-decimal units
  price?: string; // 2-decimal INR limit price
  timeInForce: 'IOC' | 'GTC' | 'DAY';
}

export interface TradeSettlementFeedItem {
  tradeId: string;
  orderId: string;
  isin: string;
  side: 'BUY' | 'SELL';
  price: string;
  quantity: string;
  totalInr: string;
  realizedPnl?: string;
  platformFeeCharged?: string;
  besuTxHash: `0x${string}`;
  blockNumber: number;
  timestamp: string;
  status: 'PENDING_DVP' | 'SETTLED_ON_CHAIN' | 'FAILED';
}
```

## Security & Compliance Notes
- **Client-Side Idempotency:** Generates unique UUIDv4 `idempotencyKey` for every order request to prevent duplicate order submission on double-click or network retry.
- **Pre-Trade Risk Checks:** Validates order parameters against available INR wallet ledger balance and trading price collars (circuit limits ±10%) before dispatching to API.
- **SEBI High-Value Order Confirmation:** Prompts mandatory secondary confirmation modal for orders exceeding ₹5,00,000.
- **WebSocket Message Authentication:** Secures WebSocket handshake with short-lived ephemeral ticket tokens rather than exposing permanent JWTs.

## Acceptance Criteria
- [ ] Multi-pane trading terminal renders cleanly on desktop (1920x1080, 1440x900) and adapts to tablets.
- [ ] TradingView Lightweight Charts loads historical data and updates seamlessly with live WebSocket candle ticks.
- [ ] Level-2 Order Book updates smoothly without frame drops (>50 updates/sec handled without freezing UI).
- [ ] Fractional share calculations accurately convert INR amount to fractional shares up to 6 decimal places.
- [ ] Pre-trade fee calculator correctly displays 0.00% platform fee (No fee at all) on trade turnover, highlights zero holding/AUM fees, and clarifies FIFO capital gains computed strictly for user tax compliance (Section 111A/112A).
- [ ] Order cancellation sends request with immediate optimistic UI update.
- [ ] On-chain DvP settlement status and transaction hash link to Besu Explorer accurately render in settlement tab.
- [ ] Playwright E2E tests verify order placement, cancel, and chart timeframe switching.

## Suggested Order / Dependencies
- **Prerequisites:** 204 (Order Service), 205 (Matching Engine), 207 (Market Data), 209 (Portfolio Service), 601 (Web Scaffolding).
- **Direct Successors / Parallel:** 508 (Flutter Charts), 509 (Flutter Orders), 606 (Proof of Reserve Dashboard).
