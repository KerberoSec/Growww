# 621 - Next.js 14 Demo Paper Trading Simulator & Performance Analytics Portal

## Purpose
In modern financial markets and regulated digital asset platforms operating under SEBI and IFSCA guidelines, market participants range from novice retail investors seeking to understand order types to quantitative proprietary desks testing automated execution algorithms. Exposing unverified strategies or untrained retail users directly to live capital markets introduces significant risk of immediate capital destruction, psychological churn, and inadvertent regulatory infractions. Furthermore, static historical backtesting cannot replicate the live market microstructure dynamics, real-time order queue latency, and emotional discipline required during live price discovery.

Under ADR-0058 (Risk-Free Paper Trading Simulation & Competency-Based Trader Graduation), Growww implements a dedicated, high-fidelity web paper trading simulator and analytics environment. This portal allows users to execute manual and algorithmic trading strategies against live, real-time Bitcoin (BTC) and INR/USD market feeds with zero financial risk. To transform paper trading from a simple toy into a genuine institutional training and strategy refinement engine, the portal integrates comprehensive algorithmic performance metrics, including trade win rate, annualized Sharpe ratio, Sortino ratio, maximum drawdown (MDD), and profit factor.

This prompt specifies the end-to-end architecture, frontend interface, real-time state synchronization, virtual double-entry accounting, testnet blockchain verification, and educational graduation bridge for the **Next.js 14 Demo Paper Trading Simulator & Performance Analytics Portal** (`apps/growww_web/demo`). The portal delivers an authentic institutional trading experience, complete with virtual balance provisioning via a 1-click testnet faucet, realistic slippage and fee simulation, simulated trade settlement receipts on Hyperledger Besu Testnet, and a seamless graduation path to live money trading once performance criteria are satisfied.

## What You Are Building
An enterprise-grade, high-performance Next.js 14 web application located in `apps/growww_web/demo` (and App Router route group `apps/growww_web/app/(demo)/`) featuring:
- `PersistentDemoWatermarkBanner`: High-contrast, unyielding amber/gold warning header pinned across every view, complemented by a semi-transparent repeating diagonal CSS background watermark ("SIMULATION / PAPER TRADING - NO REAL MONEY INVOLVED") to eliminate any ambiguity between paper and real capital.
- `VirtualPortfolioSummaryCard`: Real-time portfolio cockpit rendering virtual cash balances (vINR / vUSD), virtual Bitcoin holdings (vBTC), equity valuation, free margin, margin utilization percentage, unrealized PnL, and realized daily PnL.
- `OneClickFaucetModal`: Self-service virtual liquidity top-up dialog allowing users to instantly credit 1.00000000 vBTC and 100,000 vUSD (or ₹10,00,000 vINR) with rate-limit cooldown timers and a one-click "Reset Portfolio to Default" trigger.
- `DemoOrderEntryWidget`: Fast execution ticket supporting Market, Limit, Stop-Loss, Take-Profit, and Trailing Stop orders with customizable leverage (1x spot to 20x demo margin), slippage modeling, and simulated execution latency matching live engine parameters.
- `SimulatedPnLChart`: Interactive, responsive performance visualization powered by Recharts, rendering time-series equity curves, cumulative percentage returns against a Bitcoin buy-and-hold benchmark, and daily underwater drawdown area graphs.
- `PerformanceAnalyticsDashboard`: Deep quantitative performance metric panel calculating:
  - Win Rate (% winning trades out of total closed positions).
  - Profit Factor (Gross Profits divided by Gross Losses).
  - Annualized Sharpe Ratio (excess return per unit of volatility relative to risk-free rate).
  - Sortino Ratio (downside deviation risk-adjusted return).
  - Maximum Drawdown (peak-to-trough percentage drop and duration in days/hours).
  - Expectancy, Average Win vs Average Loss ratio, and Maximum Consecutive Losses.
- `DemoTradeHistoryTable`: High-throughput TanStack Table v8 rendering active demo orders, open positions with real-time mark-to-market liquidation thresholds, historical filled orders, simulated brokerage/clearing fees, and clickable Besu testnet verification hashes.
- `BesuTestnetReceiptModal`: Cryptographic verification viewer querying simulated trade settlement logs on Hyperledger Besu Testnet via Viem 2.x, displaying transaction hash, block number, and gasless settlement event parameters.
- `GraduateToRealMoneyBridge`: Milestone-driven conversion banner and modal triggered when traders achieve demonstrated competence (e.g., minimum 20 simulated trades, positive Sharpe ratio, win rate > 50%), guiding them directly into the verified KYC onboarding pipeline (`apps/growww_web/app/(onboarding)/kyc`) while securely preserving paper trading records.

## Scope Boundaries
- **In Scope:**
  - Next.js 14 App Router application structure under `apps/growww_web/demo` and route group `apps/growww_web/app/(demo)/`.
  - Client-side and server-rendered virtual portfolio management, position tracking, and margin calculation.
  - Real-time WebSocket connection to Live Market Data Service (Prompt 207) for live BTC/INR and BTC/USD price feeds and L2 order book depth.
  - Bidirectional communication with Demo Matching Engine (Prompt 273) and Demo Wallet Service (Prompt 274).
  - Isolated client state management utilizing Zustand with persistence for demo preferences and offline calculation.
  - Recharts integration for responsive equity curve graphs, drawdown charts, and PnL distribution histograms.
  - TanStack Table v8 implementation for virtual order books, open positions, closed trades, and fee ledgers.
  - Client-side statistical calculation engine computing Sharpe, Sortino, MDD, profit factor, and win rate.
  - Read-only blockchain interaction with Hyperledger Besu Testnet using Viem 2.x to verify simulated settlement receipts.
  - Unambiguous visual demo watermarking and strict session cookie isolation between demo and real trading accounts.
  - Interactive "Graduate to Real Money" modal with qualification rule evaluator and KYC redirection.
  - Comprehensive unit, statistical mathematical, and Playwright end-to-end test suites.
- **Out of Scope / Handled Elsewhere:**
  - Production matching engine and real order books (handled by Prompt 204 and Prompt 205).
  - Production banking, nodal fiat escrow accounts, and real INR UPI/IMPS transfers (handled by Prompt 203 and Prompt 212).
  - Production KYC, PAN verification, and DigiLocker integrations (handled by Prompt 202 and Prompt 602).
  - Real asset custody, cold storage HSMs, and Fireblocks multisig coordination (handled by Prompt 213 and Prompt 307).
  - Backend implementation of Demo Matching Engine microservice (handled by Prompt 273).
  - Backend implementation of Demo Wallet Service microservice (handled by Prompt 274).
  - Historical multi-year backtesting data replay infrastructure (handled by Prompt 260 and Prompt 265).

## Technology to Use
- **Frontend Framework:** Next.js 14 App Router leveraging React Server Components (RSC) for initial page hydration, streaming SSR with Suspense boundaries, and React 18 Client Components for high-frequency interactive canvas and chart widgets.
- **Language:** TypeScript 5.x configured with strict mode, zero `any` assertions, and exact optional property typing.
- **UI Components & Styling:** Tailwind CSS 3.4+ configured with tabular numeric typography (`font-mono`, `tabular-nums`), high-contrast demo theme accents (amber/yellow-500), and accessible headless UI primitives from Radix UI / shadcn/ui (Dialog, Tabs, Card, Badge, Tooltip, Sheet, Progress, Alert).
- **Data Visualization & Charting:** Recharts 2.x for responsive rendering of portfolio equity curves, daily underwater drawdown area graphs, return distributions, and benchmark overlays.
- **Table & Data Grid:** TanStack Table v8 (`@tanstack/react-table`) for virtual order books, active positions, trade execution history, and fee breakdown with sorting, pagination, and column filtering.
- **Real-Time Data & State Management:** Zustand with Immer middleware for local virtual portfolio state, active orders, and WebSocket subscription tracking; TanStack Query (React Query v5) for server-state synchronization.
- **High-Precision Mathematics:** Decimal.js and `BigInt` primitives for all financial calculations, margin ratios, unrealized PnL, liquidation prices, and risk metrics to avoid IEEE 754 floating-point inaccuracies.
- **Web3 Blockchain Client:** Viem 2.x configured for Hyperledger Besu Testnet (Chain ID 1337 / QBFT consortium network) to inspect simulated trade settlement events via JSON-RPC.
- **Form Management & Validation:** React Hook Form integrated with Zod schemas for order parameter validation, leverage boundaries, and faucet claim inputs.

## Backend / Infra Touchpoints
- **Demo Matching Engine (Prompt 273):**
  - Communicates via authenticated REST and low-latency WebSockets (`/api/v1/demo/orders` and `wss://api.growww.in/ws/v1/demo`).
  - Matches demo orders against live real-time market order book depth without submitting orders to the production exchange matching engine.
  - Ingests simulated Market, Limit, and Stop orders, simulates order queue priority and execution slippage, and broadcasts execution reports (`ORDER_FILLED`, `ORDER_PARTIALLY_FILLED`, `ORDER_REJECTED`, `ORDER_CANCELLED`).
- **Demo Wallet Service (Prompt 274):**
  - Manages segregated virtual double-entry balance sheets for paper traders.
  - Endpoints:
    - `GET /api/v1/demo/wallet/balance`: Returns virtual cash balances (vINR/vUSD), locked margins, and virtual coin balances.
    - `POST /api/v1/demo/wallet/faucet`: Claims virtual testnet assets with rate limiting.
    - `POST /api/v1/demo/wallet/reset`: Resets portfolio to initial virtual seed balance ($100,000 vUSD or ₹10,00,000 vINR).
    - `GET /api/v1/demo/wallet/transactions`: Returns virtual double-entry ledger journals.
- **Live Market Data Service (Prompt 207):**
  - Connects to public WebSocket feed (`wss://api.growww.in/ws/v1/marketdata`) to receive real-time Bitcoin price ticks, best bid/ask (BBO), and L2 depth for mark-to-market valuations and realistic order matching.
- **API Gateway & Session Boundary (Prompt 219):**
  - Isolates demo authentication from production sessions.
  - Issues dedicated demo tokens (`growww_demo_token`) scoped strictly to `/demo` and `/api/v1/demo/*` endpoints, preventing cross-environment credential leaking.
- **Audit Log Service (Prompt 218):**
  - Records demo session analytics, milestone achievements, and graduation events for continuous UX improvement.

## Blockchain Interaction
- **Target Network:** Hyperledger Besu Testnet (QBFT consensus, 2-second block intervals, gasless platform sponsorship, Chain ID 1337).
- **Core Smart Contract Interfaced:**
  - `DemoTradeSettlementRegistry.sol`:
    - Deployed on the permissioned Besu testnet to simulate on-chain delivery-versus-payment (DvP) settlement for demo trades.
    - Contract interface:
      ```solidity
      // SPDX-License-Identifier: Apache-2.0
      pragma solidity ^0.8.24;

      interface IDemoTradeSettlementRegistry {
          event DemoTradeSettled(
              bytes32 indexed tradeId,
              address indexed virtualTrader,
              bytes32 indexed marketSymbol,
              uint256 executedPrice,
              uint256 executedQuantity,
              uint8 side, // 0 = BUY, 1 = SELL
              uint256 virtualFeePaid,
              uint256 timestamp
          );

          function recordDemoSettlement(
              bytes32 tradeId,
              address virtualTrader,
              bytes32 marketSymbol,
              uint256 executedPrice,
              uint256 executedQuantity,
              uint8 side,
              uint256 virtualFeePaid
          ) external returns (bool success);

          function getSettlementRecord(bytes32 tradeId) external view returns (
              address virtualTrader,
              bytes32 marketSymbol,
              uint256 executedPrice,
              uint256 executedQuantity,
              uint8 side,
              uint256 virtualFeePaid,
              uint256 timestamp,
              uint256 blockNumber
          );
      }
      ```
- **Client-Side Blockchain Queries:**
  - Viem 2.x `createPublicClient` configured with Besu Testnet RPC URL (`https://besu-testnet.growww.in/rpc`).
  - When an order execution report arrives via WebSocket, the client retrieves the associated transaction hash or queries `DemoTradeSettlementRegistry` logs for `DemoTradeSettled` events.
  - Users can click on any trade execution in the trade history table to open the `BesuTestnetReceiptModal`, displaying on-chain block confirmations, event logs, and cryptographic verification proofs, providing transparent auditability.

## Step-by-Step Build Instructions

1. **Scaffold Next.js 14 Demo Route Architecture:**
   - Initialize directory structure at `apps/growww_web/demo` and route group `apps/growww_web/app/(demo)/`.
   - Create route files:
     - `layout.tsx`: Root demo layout wrapping child pages with `DemoWatermarkProvider`, `DemoHeaderWithWatermark`, and isolated demo styling.
     - `page.tsx`: Main Demo Paper Trading Simulator dashboard.
     - `analytics/page.tsx`: In-depth strategy performance analytics, drawdown curves, and statistical breakdowns.
     - `history/page.tsx`: Comprehensive historical demo trade ledger, order log, and testnet blockchain receipts.

2. **Implement Persistent Demo Watermarking & Visual Warning Shield:**
   - Build `DemoWatermarkProvider` inserting a persistent, full-viewport CSS repeating pattern overlay:
     - Text: "SIMULATION / PAPER TRADING - NO REAL MONEY" rotated at -35 degrees with 4% opacity, unselectable (`pointer-events-none`).
   - Implement `DemoHeaderWithWatermark`:
     - Pinned amber alert banner (`bg-amber-500 text-black font-semibold text-xs py-1 px-4 text-center tracking-wide uppercase flex items-center justify-center gap-2`).
     - Clear indicator: "Simulation Mode Active | Real Market Data | Virtual Capital | No Financial Risk".
     - Quick-action buttons: "Reset Portfolio", "1-Click Faucet", and "Graduate to Real Money".
   - Inject browser tab title prefix dynamically: `[DEMO] Growww Trading Simulator`.

3. **Configure Session & Cookie Isolation Middleware:**
   - Implement Next.js middleware check under `apps/growww_web/middleware.ts` for `/demo` routes.
   - Restrict authentication verification to `growww_demo_token` stored in HTTP-only, SameSite=Strict cookies with path restricted strictly to `/demo`.
   - Ensure demo sessions are completely decoupled from production JWT cookies (`growww_auth_token`), preventing cross-account state contamination.

4. **Construct Zustand Virtual Portfolio & Demo State Store:**
   - Create `useDemoTradingStore` in `apps/growww_web/lib/stores/demo-trading-store.ts` using Zustand with Immer.
   - Maintain state for:
     - Virtual balances: `virtualUsdBalance`, `virtualInrBalance`, `virtualBtcBalance`, `lockedMargin`.
     - Active orders: Map of open Limit and Stop orders.
     - Open positions: Map of active leveraged or spot positions.
     - Live market prices: Current BTC/USD and BTC/INR tick data, bid/ask spreads.
     - Performance cache: Historical equity time-series and trade logs.

5. **Build 1-Click Testnet Faucet Modal (`OneClickFaucetModal`):**
   - Implement interactive Radix UI dialog accessible from the demo header.
   - Provide presets: "Standard Starter ($10,000 vUSD + 0.5 vBTC)", "Pro Strategy Tester ($100,000 vUSD + 2.0 vBTC)", "Custom Top-Up".
   - Connect to Demo Wallet Service (Prompt 274) endpoint `POST /api/v1/demo/wallet/faucet`.
   - Incorporate a 60-second client-side cooldown countdown to discourage spamming.
   - Add "Reset All Positions & Balances" button with confirmation prompt executing `POST /api/v1/demo/wallet/reset`.

6. **Integrate Real-Time Bitcoin Market Data WebSocket:**
   - Implement `useLiveMarketData` hook subscribing to `wss://api.growww.in/ws/v1/marketdata`.
   - Subscribe to symbols: `BTC-INR` and `BTC-USD`.
   - Ingest real-time trade ticks, best bid/ask (BBO), and 24-hour high/low metrics.
   - Feed incoming ticks directly into `useDemoTradingStore` for sub-second mark-to-market position revaluation.

7. **Implement Demo Order Entry Ticket (`DemoOrderEntryWidget`):**
   - Create responsive order placement component with tabs: "Market", "Limit", "Stop-Limit".
   - Include side toggles: "Buy / Long" (green) and "Sell / Short" (red).
   - Add leverage slider: 1x (Spot) up to 20x (Demo Margin) with dynamic margin requirement calculation.
   - Display real-time pre-trade execution preview:
     - Estimated fill price (accounting for simulated order book depth slippage).
     - Required virtual margin.
     - Estimated simulated trading fee (0.00% maker / 0.00% taker - Zero Brokerage).
     - Estimated liquidation price for leveraged positions.
   - Form submission validates inputs via Zod and dispatches payload to Demo Matching Engine (`POST /api/v1/demo/orders` or WebSocket).

8. **Build Virtual Positions & Active Orders TanStack Table:**
   - Implement `DemoPositionsTable` using `@tanstack/react-table`:
     - Columns: Market Symbol, Side, Size (vBTC), Entry Price, Mark Price, Liquidation Price, Margin Ratio, Unrealized PnL ($ and %), Action ("Close Position", "Market Reverse").
     - Dynamic color-coding for positive (green) and negative (red) unrealized PnL with pulsating tick indicators.
   - Implement `DemoActiveOrdersTable`:
     - Columns: Order ID, Time, Symbol, Type, Side, Quantity, Limit Price, Trigger Price, Status, Action ("Cancel Order").

9. **Connect Demo Matching Engine Real-Time Event Stream:**
   - Establish duplex WebSocket connection to Demo Matching Engine (Prompt 273): `wss://api.growww.in/ws/v1/demo`.
   - Handle incoming message types:
     - `DEMO_ORDER_ACK`: Acknowledges receipt of order.
     - `DEMO_ORDER_FILL`: Contains executed price, quantity, fee, and Besu testnet trade ID. Plays subtle audio fill confirmation and displays toast notification.
     - `DEMO_POSITION_UPDATE`: Updates open position sizing and realized PnL.
     - `DEMO_LIQUIDATION_WARNING`: Alerts user when margin utilization exceeds 85%.

10. **Implement Quantitative Performance Analytics Calculation Engine:**
    - Develop `apps/growww_web/lib/analytics/performance-calculator.ts` implementing:
      - **Win Rate:** `(Winning Trades / Total Closed Trades) * 100`.
      - **Profit Factor:** `Sum of Profits / Sum of Losses`.
      - **Annualized Sharpe Ratio:** `(Mean Portfolio Return - Risk Free Rate) / Standard Deviation of Returns * sqrt(365)`.
      - **Sortino Ratio:** `(Mean Portfolio Return - Risk Free Rate) / Downside Semi-Variance * sqrt(365)`.
      - **Maximum Drawdown (MDD):** `max((Peak Value - Trough Value) / Peak Value)` across the historical equity series.
      - **Average Trade Duration & Trade Expectancy:** Mathematical expectation per trade.
    - Ensure zero floating-point division errors by handling edge cases (e.g. zero losing trades, empty trade history).

11. **Construct Recharts Performance Visualization Dashboard:**
    - Build `SimulatedPnLChart` inside `apps/growww_web/demo/components/pnl-chart.tsx`:
      - Primary AreaChart: Virtual portfolio equity curve over time with green/red gradient fill.
      - Benchmark Line: Overlaid Bitcoin buy-and-hold equivalent return line.
      - Underwater Drawdown Chart: Secondary Recharts AreaChart plotting negative drawdown percentages from previous peak equity.
      - Interactive tooltip showing timestamp, equity value, unrealized PnL, and benchmark comparison.

12. **Build Besu Testnet Settlement Receipt Modal:**
    - Implement `BesuTestnetReceiptModal` triggered from trade history rows:
      - Accepts `besuTradeId` and queries Besu Testnet using Viem `publicClient.getLogs`.
      - Displays transaction hash, block number, QBFT validator signature timestamp, gas used (0 gas, subsidized), and verified event arguments (`tradeId`, `virtualTrader`, `executedPrice`, `executedQuantity`).
      - Provides a direct link to the internal Besu testnet block explorer.

13. **Implement "Graduate to Real Money Trading" Bridge:**
    - Build `GraduateToRealMoneyBridge` component evaluating trader qualification rules:
      - Minimum 20 closed simulated trades.
      - Minimum 7 consecutive active days of paper trading.
      - Win rate >= 50%.
      - Profit factor >= 1.25.
      - Maximum drawdown <= 25%.
    - Render milestone progression bar (e.g. "3 of 5 Graduation Milestones Met").
    - When all milestones are unlocked, present congratulatory modal:
      - Highlights trader's simulated performance summary (Sharpe ratio, net virtual profit).
      - Primary action: "Graduate to Real Money" directing user to `/onboarding/kyc` (Prompt 602).
      - Automatically archives demo portfolio performance for user's personal analytics record.

14. **Testing, Linting & Verification:**
    - Implement Vitest unit tests verifying:
      - Math engine: Sharpe, Sortino, MDD, and win rate calculation accuracy against static test vectors.
      - Virtual margin and liquidation price computations.
      - Zod order validation rules.
    - Implement Playwright end-to-end tests validating:
      - Loading demo page and verifying persistent watermark presence.
      - Claiming testnet funds via Faucet modal and checking balance updates.
      - Submitting simulated Market and Limit orders on BTC.
      - Verifying WebSocket order fill toast, position creation, and PnL chart rendering.
      - Verifying session cookie pathing and isolation from live trading routes.

## Interfaces / Contracts

### TypeScript Models & Client Interfaces

```typescript
export type OrderSide = 'BUY' | 'SELL';
export type OrderType = 'MARKET' | 'LIMIT' | 'STOP_LIMIT' | 'TAKE_PROFIT';
export type OrderStatus = 'PENDING' | 'OPEN' | 'FILLED' | 'PARTIALLY_FILLED' | 'CANCELLED' | 'REJECTED';

export interface VirtualBalance {
  currency: 'vUSD' | 'vINR' | 'vBTC';
  total: string;       // Decimal string representation
  available: string;   // Free balance available for orders
  lockedMargin: string; // Balance committed to open positions / limit orders
}

export interface VirtualPortfolio {
  userId: string;
  accountType: 'PAPER_TRADING';
  createdAt: string;
  balances: Record<'vUSD' | 'vINR' | 'vBTC', VirtualBalance>;
  equityInr: string;
  equityUsd: string;
  unrealizedPnlUsd: string;
  unrealizedPnlPercentage: number;
  realizedPnlUsd: string;
  marginUtilizationRate: number; // 0.0 to 1.0
  isLiquidating: boolean;
}

export interface DemoOrderPayload {
  clientOrderId: string;
  symbol: 'BTC-USD' | 'BTC-INR';
  side: OrderSide;
  orderType: OrderType;
  quantityBtc: string;
  limitPrice?: string;
  stopPrice?: string;
  leverage: number; // 1 to 20
  timeInForce: 'GTC' | 'IOC' | 'FOK';
}

export interface DemoTradeRecord {
  tradeId: string;
  orderId: string;
  clientOrderId: string;
  symbol: string;
  side: OrderSide;
  executedPrice: string;
  executedQuantity: string;
  virtualFeePaid: string;
  realizedPnl?: string;
  executionTimestamp: string;
  besuTestnetTxHash: `0x${string}`;
  besuBlockNumber: number;
}

export interface DemoPosition {
  positionId: string;
  symbol: string;
  side: OrderSide;
  sizeBtc: string;
  entryPrice: string;
  markPrice: string;
  liquidationPrice: string;
  allocatedMargin: string;
  leverage: number;
  unrealizedPnlUsd: string;
  unrealizedPnlPercentage: number;
  createdAt: string;
}

export interface FaucetClaimResponse {
  success: boolean;
  creditedUsd: string;
  creditedBtc: string;
  newUsdBalance: string;
  newBtcBalance: string;
  cooldownExpiresAt: string;
  message: string;
}

export interface PerformanceAnalyticsSummary {
  totalTrades: number;
  winningTrades: number;
  losingTrades: number;
  winRatePercentage: number;
  profitFactor: number;
  grossProfitsUsd: string;
  grossLossesUsd: string;
  netRealizedPnlUsd: string;
  annualizedSharpeRatio: number;
  sortinoRatio: number;
  maxDrawdownPercentage: number;
  maxDrawdownDurationHours: number;
  averageTradeDurationMinutes: number;
  expectancyUsd: string;
  consecutiveWinsMax: number;
  consecutiveLossesMax: number;
}

export interface EquityDataPoint {
  timestamp: string;
  equityUsd: number;
  drawdownPercentage: number;
  benchmarkBtcHoldUsd: number;
}

export interface BesuTestnetTradeReceipt {
  tradeId: string;
  contractAddress: `0x${string}`;
  transactionHash: `0x${string}`;
  blockNumber: number;
  blockTimestamp: number;
  virtualTrader: `0x${string}`;
  marketSymbol: string;
  executedPrice: string;
  executedQuantity: string;
  side: 'BUY' | 'SELL';
  virtualFeePaid: string;
  rawEventSignature: string;
}

export interface GraduationMilestoneStatus {
  totalTradesTarget: number;
  currentTradesCount: number;
  minActiveDaysTarget: number;
  currentActiveDays: number;
  minWinRateTarget: number;
  currentWinRate: number;
  minProfitFactorTarget: number;
  currentProfitFactor: number;
  maxDrawdownLimit: number;
  currentMaxDrawdown: number;
  isEligibleForGraduation: boolean;
}
```

### REST & WebSocket API Endpoints

- `GET /api/v1/demo/portfolio`:
  - Returns current `VirtualPortfolio` state, balances, and margin utilization.
- `POST /api/v1/demo/wallet/faucet`:
  - Request: `{ requestedAsset: 'ALL' | 'vUSD' | 'vBTC' }`
  - Response: `FaucetClaimResponse` with updated balances and cooldown timestamp.
- `POST /api/v1/demo/wallet/reset`:
  - Request: `{ confirmReset: boolean }`
  - Resets all open positions, pending orders, and balances back to default $100,000 vUSD.
- `GET /api/v1/demo/orders`:
  - Query: `?status=OPEN&limit=50`
  - Returns list of active open demo orders.
- `POST /api/v1/demo/orders`:
  - Request: `DemoOrderPayload`
  - Submits paper order to Demo Matching Engine; returns order confirmation with `orderId`.
- `DELETE /api/v1/demo/orders/{orderId}`:
  - Cancels active pending demo order; releases virtual locked margin.
- `GET /api/v1/demo/positions`:
  - Returns list of open `DemoPosition` items with mark-to-market valuations.
- `POST /api/v1/demo/positions/{positionId}/close`:
  - Closes an active position at current market price.
- `GET /api/v1/demo/analytics`:
  - Returns calculated `PerformanceAnalyticsSummary` and historical `EquityDataPoint[]`.
- `GET /api/v1/demo/trades/{tradeId}/testnet-receipt`:
  - Queries Besu testnet log cache and returns `BesuTestnetTradeReceipt`.
- `WSS wss://api.growww.in/ws/v1/demo`:
  - Real-time duplex WebSocket stream for demo trading:
    - Inbound: `{ action: 'subscribe', channels: ['demo:orders', 'demo:positions', 'demo:pnl'] }`
    - Outbound: Broadcasts order acknowledgements, executions, mark-to-market ticks, and margin calls.

## Security & Compliance Notes
- **Prominent Persistent Demo Watermark:**
  - In strict alignment with SEBI investor protection regulations and financial advertisement guidelines, paper trading simulators must never be mistaken for real money brokerage accounts.
  - The UI enforces a fixed, high-contrast banner at the top of the viewport and a background SVG/CSS diagonal repeating watermark ("SIMULATION / PAPER TRADING") across all views under `/demo`.
  - The browser document title is persistently prefixed with `[DEMO]` to prevent user confusion across multiple browser tabs.
- **Complete Session & Cookie Isolation:**
  - Demo accounts operate with an isolated authentication token (`growww_demo_token`) set with strict cookie scoping (`Path=/demo`, `SameSite=Strict`, `HttpOnly`, `Secure`).
  - Production credentials (`growww_auth_token`) are never read or accepted by `/api/v1/demo/*` endpoints.
  - Logging into or out of the demo environment has zero impact on live trading sessions, preventing unintentional order dispatch to live markets.
- **Zero Real Capital Commingling:**
  - Demo wallet endpoints are physically isolated in backend architecture (Prompt 274), operating against a distinct sandbox database instance.
  - Virtual assets (`vUSD`, `vINR`, `vBTC`) cannot be transferred, withdrawn, or used as collateral on live exchange order books.
  - Real payment gateway integrations (UPI, IMPS, RTGS) are completely excluded from demo routes.
- **Anti-Abuse & Rate Limiting:**
  - Faucet endpoints enforce IP and account-level rate limits (maximum 1 claim per 60 seconds; maximum 5 claims per 24 hours per IP address) to prevent resource exhaustion on demo matching nodes.
  - Order entry is rate-limited to 20 orders per second per user to simulate realistic API tier limits.
- **Regulatory Disclaimers & No-Advice Declarations:**
  - Every analytics report and equity curve displays a mandatory statutory disclaimer:
    "Simulated paper trading performance does not guarantee future financial results in live markets. Past hypothetical returns do not account for real-world liquidity limitations, extreme slippage, market impact, or exchange outage risks. Provided strictly for educational and algorithmic testing purposes."

## Acceptance Criteria
- [ ] Next.js 14 application compiles and builds successfully under `apps/growww_web/demo` with zero TypeScript errors or lint warnings.
- [ ] `DemoHeaderWithWatermark` displays prominently at the top of all demo pages with clear warning styling and contrast.
- [ ] Repeating diagonal watermark ("SIMULATION / PAPER TRADING") is present and unselectable across all screen viewports.
- [ ] Browser title dynamically reflects `[DEMO]` prefix on all simulator routes.
- [ ] Demo session cookies are strictly scoped to path `/demo` with zero cross-contamination with live session tokens.
- [ ] `OneClickFaucetModal` dispenses virtual funds correctly, updates `VirtualPortfolioSummaryCard` within 500ms, and enforces a 60-second cooldown timer.
- [ ] Portfolio reset button clears all active positions, pending orders, and restores initial virtual balance accurately.
- [ ] Live Bitcoin price feeds stream via WebSocket from Prompt 207 and trigger real-time mark-to-market updates for open positions.
- [ ] Order entry ticket allows placing Market, Limit, and Stop orders with leverage (1x to 20x), calculating required margin and estimated liquidation price in real time.
- [ ] Order fills broadcast from Demo Matching Engine (Prompt 273), updating order history, position table, and cash balance without page reload.
- [ ] TanStack Table renders open positions, active orders, and trade history with sorting, filtering, and responsive pagination.
- [ ] Performance analytics engine accurately computes Win Rate, Profit Factor, Sharpe Ratio, Sortino Ratio, and Maximum Drawdown matching mathematical test vectors.
- [ ] Recharts equity curve and underwater drawdown graphs update dynamically as trades close.
- [ ] Clicking on a trade history row displays `BesuTestnetReceiptModal` with verified transaction hash and block logs from Hyperledger Besu Testnet.
- [ ] "Graduate to Real Money Trading" component tracks qualification milestones and redirects eligible users to `/onboarding/kyc` with archived performance stats.
- [ ] Vitest test suite passes with 100% coverage on financial and statistical analytics math modules.
- [ ] Playwright E2E tests verify the complete user simulation journey: visiting demo portal -> claiming faucet -> executing market order -> viewing position -> closing trade -> verifying analytics and testnet receipt.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 207: Real-Time Market Data & Order Book L2 Service (for live Bitcoin market tick stream).
  - Prompt 273: Demo Matching Engine Microservice (for paper order routing and execution simulation).
  - Prompt 274: Demo Wallet & Virtual Double-Entry Ledger Service (for virtual balances and faucet management).
  - Prompt 320: Permissioned Besu Testnet Cluster & Faucet Infrastructure (for simulated on-chain trade settlement).
  - Prompt 601: Next.js 14 Investor Web App Scaffolding (for shared Tailwind UI primitives and theme provider).
- **Parallel Tasks:**
  - Prompt 603: Web Trading Dashboard (production counterpart sharing charting and order book UI components).
  - Prompt 609: Public Developer Portal & Testnet Faucet UI (developer-centric API and gas faucet).
- **Downstream Blockers:**
  - Prompt 602: Web Onboarding, DigiLocker & Camera KYC Flow (destination for users graduating to real money).
  - Prompt 906: Comprehensive UAT Plan & Regulatory Sandbox Pilot Scenarios.
