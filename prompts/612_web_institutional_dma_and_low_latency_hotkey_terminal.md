# 612 - Next.js 14 Institutional Direct Market Access (DMA) Workstation

## Purpose
Institutional proprietary trading desks, quantitative market makers, hedge funds, and agency execution brokers require deterministic, sub-millisecond execution capabilities with zero DOM overhead, millisecond-accurate market depth visualization, tactile keyboard hotkey navigation, and customizable multi-monitor workspace configurations. Operating under SEBI Direct Market Access (DMA) circulars, algorithmic execution governance guidelines, and IFSCA exchange trading rules, institutional market participants cannot tolerate the latency jitter, DOM thrashing, or UI re-render bottlenecks typical of retail web trading applications.

This prompt specifies the architecture, engineering design, and implementation of the Next.js 14 Institutional Direct Market Access (DMA) Workstation (`apps/growww_web/app/(dma)/dma/`). The system delivers an ultra-high-performance trading terminal featuring a hardware-accelerated WebGL / PixiJS Order Book Price Ladder (Depth of Market / DOM), a zero-garbage-collection WebWorker off-thread binary WebSocket parser, a multi-monitor Golden Layout docking panel manager, an institutional low-latency hotkey execution subsystem with safety collars, real-time round-trip latency telemetry across the FIX Protocol Gateway (Prompt 225), Order Service (Prompt 204), and Market Data Service (Prompt 207), and real-time transaction broadcast queue tracking with Hyperledger Besu consortium settlement finality latencies.

## What You Are Building
A standalone, high-performance institutional trading application and panel ecosystem featuring:
- `GoldenLayoutDockingSystem`: Multi-window, multi-monitor dockable panel manager built on modern Golden Layout / FlexLayout primitives, allowing institutional traders to float, dock, split, tab, tear out, and persist custom workspace topologies across multi-monitor display arrays.
- `WebGlOrderBookLadder`: Hardware-accelerated Depth of Market (DOM) / Price Ladder rendered via PixiJS and raw WebGL 2.0 shaders at a constant 60-120 FPS, visualizing 100+ price ticks, cumulative depth histograms, dynamic spread ribbons, resting orders, and single-click or hotkey order injection/amendment/pulling.
- `LowLatencyHotkeyEngine`: High-precision keyboard interaction subsystem with global and panel-scoped hotkey listeners, tactile single-key execution (e.g., Space = Buy Market, Shift+Space = Sell Market, Escape = Cancel All / Emergency Flatten), debounce de-jitter, two-stroke verification for high-impact capital operations, and an on-demand shortcut cheatsheet HUD.
- `OffThreadWebSocketParserWorker`: Dedicated WebWorker operating outside the React main UI thread, consuming high-throughput binary WebSocket feeds (Simple Binary Encoding / Protobuf / raw ArrayBuffers), aggregating delta books, computing sorted price depth, and streaming frame snapshots to the UI via `SharedArrayBuffer` or zero-copy transferable objects.
- `InstitutionalExecutionBlotter`: Virtualized, sub-millisecond execution grid displaying Parent/Child orders, algorithmic execution slices (TWAP, VWAP, POV, Iceberg), fill rates, volume-weighted average price (VWAP) slippage, and instant cancel/replace controls.
- `FixGatewayLatencyMonitor`: Microsecond-precision network telemetry ribbon visualizing round-trip times (RTT) across the browser, FIX Protocol Gateway (Prompt 225), Matching Engine (Prompt 205), and Pre-Trade Risk Engine (Prompt 206), complete with rolling p50/p90/p99 latency histograms.
- `BlockchainBroadcastQueueViewer`: Real-time on-chain queue monitor rendering pending Delivery-versus-Payment (DvP) transactions, mempool propagation stages, Hyperledger Besu block inclusion latencies, and finality confirmations.
- `PreTradeRiskGuardHUD`: High-visibility institutional risk ribbon displaying dynamic buying power, net/gross notional exposure limits, order-per-second velocity throttles (SEBI algorithmic collar compliance), margin utilization meters, and an instant master killswitch.

## Scope Boundaries
- **In Scope:**
  - Dedicated Next.js 14 App Router DMA workstation layout and routing (`/dma`).
  - Golden Layout docking framework with multi-window tear-out, layout persistence (localStorage + encrypted remote sync), and workspace presets.
  - WebGL / PixiJS hardware-accelerated Level-2 / Level-3 order book ladder (DOM) running at continuous 60-120 FPS.
  - Dedicated WebWorker binary WebSocket decoder and ring-buffer for high-frequency market data ingestion.
  - Granular hotkey engine with scoped keyboard focus trapping, safety guards, and customizable keymaps.
  - Virtualized institutional execution blotter, active fills stream, and position/P&L tracker.
  - Latency waterfall telemetry component (Browser -> Edge -> FIX Gateway -> Matching Engine -> Besu Ledger).
  - Hyperledger Besu transaction broadcast queue and DvP finality tracker.
  - Pre-trade risk guard HUD with SEBI rate-limit indicators and emergency panic flatten workflows.
  - Enterprise web security hardening: Content Security Policy (CSP), Sub-Resource Integrity (SRI), and anti-clickjacking frame busting.
- **Out of Scope / Handled Elsewhere:**
  - FIX Protocol Gateway backend and TCP session state machine (`services/fix-gateway` - Prompt 225).
  - Core Order Matching Engine and memory-mapped write-ahead logging (`services/matching-engine` - Prompt 205, Prompt 246).
  - Pre-Trade Risk & Margin Checks Microservice backend (`services/risk-margin` - Prompt 206).
  - Market Data WebSocket distribution server (`services/market-data` - Prompt 207).
  - Smart contract DvP atomic settlement and on-chain clearing contracts (`SettlementDvP.sol` - Prompt 306).
  - Algorithmic slice engine backends (TWAP / VWAP / POV container - Prompt 226).

## Technology to Use
- **Next.js 14 App Router (React 18/19, TypeScript 5.4+):** Server-rendered initial application shell for instantaneous terminal bootstrap with strict `'use client'` execution sandboxes for canvas and WebSocket pipelines.
- **PixiJS 8.x / WebGL 2.0:** Hardware-accelerated 2D graphics engine utilizing GPU vertex buffers and custom batch shaders to render price ladder rows, volume depth bars, and order flags at sub-5ms draw cycles without DOM node overhead.
- **Golden Layout (v2.6+) / modern TypeScript port:** Multi-pane docking, splitting, stacking, and popping-out window manager configured with high-contrast institutional dark theme CSS variables.
- **WebWorkers & Transferable ArrayBuffers:** Dedicated background worker threads parsing incoming binary WebSocket frames off the main thread, transferring memory buffers via zero-copy semantics to eliminate UI frame drops.
- **Tailwind CSS 3.4+:** High-contrast institutional dark color palette (Obsidian `#0A0E17`, Terminal Amber `#FFB000`, Execution Cyan `#00E5FF`, Bid Emerald `#00E676`, Ask Crimson `#FF1744`).
- **Zustand with Immer & Vanilla Stores:** Zero-overhead outside-React state stores feeding the WebGL canvas and hotkey dispatchers without triggering React component tree re-renders.
- **Web Audio API:** Ultra-low-latency synthesized audio feedback pings (<10ms latency) for fill confirmations, order rejections, risk breaches, and panic killswitch actuation.
- **Decimal.js:** High-precision fixed-point calculations for prices, fractional quantities, and currency conversions without IEEE 754 floating-point inaccuracies.

## Backend / Infra Touchpoints
- **FIX Protocol Gateway (Prompt 225):** Direct WebSocket/gRPC bridge (`wss://dma.growww.in/ws/v1/fix`) interfacing with QuickFIX/J or Go-based FIX engine for institutional order entry (Tag 35=D New Order Single, Tag 35=F Order Cancel Request, Tag 35=G Cancel/Replace, Tag 35=8 Execution Report).
- **Order Service (Prompt 204):** REST and gRPC-Web endpoint (`https://dma.growww.in/api/v1/orders`) for non-urgent order queries, trade history hydration, and position state recovery.
- **Market Data Service (Prompt 207):** High-throughput binary WebSocket stream (`wss://marketdata.growww.in/ws/v1/l2-l3`) broadcasting SBE/Protobuf Level-2 depth snapshots and incremental deltas.
- **Pre-Trade Risk & Margin Checks Service (Prompt 206):** Ingests pre-trade risk validations, gross notional limits, and margin status updates via high-speed gRPC/WebSocket stream (`wss://risk.growww.in/ws/v1/limits`).
- **Hyperledger Besu Consortium RPC (Prompt 302, Prompt 306):** Subscribes to pending transaction pools and settlement contract event streams (`SettlementDvP.sol`) to monitor block inclusion and finality latency.
- **Institutional Workspace Config Service (Prompt 212):** REST endpoint (`/api/v1/dma/workspaces`) to backup, sync, and version terminal workspace layouts, hotkey maps, and user profiles across trading desks.

## Blockchain Interaction
- **Transaction Broadcast Queue Monitoring:** The workstation continuously tracks settlement transactions emitted by the matching engine into the Hyperledger Besu consortium mempool, displaying the queue depth, gas priority, and transaction lifecycle states (`QUEUED`, `BROADCASTING`, `MEMPOOL_ACCEPTED`, `MINED`, `FINALIZED`).
- **Besu Confirmation Latency Telemetry:** Real-time calculation and visualization of on-chain confirmation latency:
  $$\Delta t_{\text{chain}} = t_{\text{block\_receipt}} - t_{\text{engine\_match}}$$
  Displaying QBFT 2-second block progress indicators and consensus validator signatures.
- **Atomic Delivery-versus-Payment (DvP) Proofs:** Every executed fill displays an interactive DvP status pill linking directly to the internal consortium Besu Explorer (`https://explorer.growww.in/tx/0x...`), rendering the verified transaction hash, block height, token transfer IDs, and smart contract execution logs.
- **Regulatory Transparency & Pseudonymity:** Zero PII is committed to the Besu ledger. Institutional entity codes and trader IDs are cryptographically hashed and mapped to authorized consortium Ethereum addresses (`0x...`) whitelisted in `InstitutionalRegistry.sol`.

## Step-by-Step Build Instructions

1. **Scaffold DMA Application Shell and Route Architecture:**
   - Create route directory `apps/growww_web/app/(dma)/dma/` with `page.tsx`, `layout.tsx`, and `loading.tsx`.
   - Configure metadata with strict anti-indexing tags (`robots: 'noindex, nofollow'`) and configure dynamic rendering (`export const dynamic = 'force-dynamic'`).
   - Create subcomponent directories under `apps/growww_web/components/dma/`: `ladder/`, `blotter/`, `docking/`, `hotkeys/`, `telemetry/`, and `workers/`.

2. **Integrate Golden Layout Multi-Window Docking Manager:**
   - Scaffold `GoldenLayoutContainer.tsx` wrapping the core viewport in a dynamic FlexLayout/Golden Layout component.
   - Define standard panel types: `LADDER` (Price Ladder DOM), `BLOTTER` (Order & Fill Blotter), `CHART` (Lightweight Chart), `TIME_AND_SALES` (Tape), `RISK_HUD` (Risk & Velocity Limits), `LATENCY_MONITOR` (Telemetry), and `BLOCKCHAIN_QUEUE` (Besu Queue).
   - Implement pop-out window capability via `window.open` synchronizing state across multi-monitor displays using `BroadcastChannel` API.
   - Implement local storage layout caching with remote backup synchronization (`POST /api/v1/dma/workspaces`).

3. **Develop Dedicated Off-Thread Market Data WebWorker (`market-data.worker.ts`):**
   - Create WebWorker script handling WebSocket connection to `wss://marketdata.growww.in/ws/v1/l2-l3`.
   - Implement binary packet parser for SBE / Protobuf Level-2 order book snapshots and delta events.
   - Maintain an off-thread contiguous memory buffer of sorted price levels (Bids descending, Asks ascending).
   - Throttle UI frame dispatches to 60 FPS (16.6ms) or 120 FPS (8.3ms) intervals using `requestAnimationFrame` equivalents in workers or high-precision timers (`setInterval`), dispatching array buffers to the main thread via transferable objects.

4. **Construct Hardware-Accelerated PixiJS WebGL Price Ladder (`WebGlOrderBookLadder.tsx`):**
   - Initialize PixiJS `Application` with WebGL 2.0 context, `antialias: false`, and `powerPreference: 'high-performance'`.
   - Render a central fixed-tick Price Column with Bids on the left and Asks on the right.
   - Implement custom PixiJS `Graphics` and `BitmapText` pools to render cumulative depth histograms behind price cells without generating garbage-collected JavaScript objects.
   - Support smooth mousewheel scrolling and auto-centering on the Best Bid/Offer (BBO) spread ribbon.
   - Render dynamic resting order markers displaying user open orders at exact price levels with volume badges.

5. **Build Interactive Single-Click & Drag Ladder Trading Mechanics:**
   - Implement instantaneous mouse interactions on the WebGL canvas:
     - Left-click on Bid Price column: Place Limit Buy at price tick.
     - Left-click on Ask Price column: Place Limit Sell at price tick.
     - Right-click or middle-click on existing order badge: Instant Cancel (Tag 35=F).
     - Drag-and-drop resting order badge: Instant Cancel/Replace (Tag 35=G) modifying price tick.
     - Column header buttons: "Cancel Bids", "Cancel Asks", "Cancel All", and "Market Flatten".

6. **Develop Low-Latency Hotkey Execution Subsystem (`useHotkeyEngine.ts`):**
   - Implement a centralized keydown event listener with strict event bubbling suppression (`e.preventDefault()`, `e.stopPropagation()`).
   - Define panel-scoped and global hotkey contexts (e.g., Ladder Active vs Blotter Active).
   - Implement default institutional hotkey configuration:
     - `Space`: Buy Market default quantity.
     - `Shift + Space`: Sell Market default quantity.
     - `1` through `9`: Quick-select preset quantity lots.
     - `Up / Down Arrow`: Shift active ladder price tick selection.
     - `Escape`: Panic Cancel All open orders for active ticker.
     - `Shift + Escape`: Master Killswitch (Cancel All + Flatten all positions).
     - `F1 - F8`: Switch active symbol tabs or workspace presets.
   - Implement a visual Hotkey HUD modal (`HotkeyCheatsheetModal.tsx`) toggled via `?`.

7. **Implement Pre-Trade Safety Collars & Fat-Finger Prevention:**
   - Construct client-side safety validation pipeline executing prior to socket dispatch:
     - Price collar check: Rejects orders with limit prices deviating >3% from BBO.
     - Max notional value check: Flags orders exceeding single-ticket institutional threshold (e.g., ₹5,00,00,000).
     - Two-stroke keypress confirmation: Demands a double-tap confirmation sequence (e.g., `Shift+Escape` followed by `Enter` within 750ms) for destructive operations (Flatten All).
   - Provide visual/audio rejection alerts when pre-trade collars are violated.

8. **Construct Virtualized Institutional Execution Blotter (`ExecutionBlotter.tsx`):**
   - Utilize `@tanstack/react-virtual` to display up to 10,000 active, filled, or cancelled orders with zero DOM latency.
   - Columns: Order ID, Client Order ID (ClOrdID), Symbol, Side, Type, Order Qty, CumQty, LeavesQty, AvgPrice, Limit Price, Status, FIX Route, Algorithmic Strategy, Latency, Actions.
   - Implement instantaneous one-click "Cancel" and "Amend" buttons for active leaves.
   - Implement Child Order expansion accordion for algorithmic parent orders (TWAP/VWAP slices).

9. **Build FIX Gateway Latency Waterfall & Telemetry Ribbon (`LatencyTelemetryRibbon.tsx`):**
   - Connect to FIX Gateway telemetry stream computing rolling latency percentiles:
     - $L_{\text{net}}$: Browser WebSocket to Gateway network round-trip.
     - $L_{\text{fix}}$: FIX Gateway serialization and validation time.
     - $L_{\text{risk}}$: Pre-trade risk evaluation time.
     - $L_{\text{match}}$: Matching engine queue and execution tick.
     - $L_{\text{total}}$: End-to-end order acknowledgement round-trip ($T_{\text{ack}} - T_{\text{send}}$).
   - Render real-time color-coded indicators: Green (<5ms), Yellow (5-20ms), Red (>20ms).

10. **Implement Blockchain Broadcast Queue & Besu Settlement Tracker (`BlockchainQueueViewer.tsx`):**
    - Connect to Hyperledger Besu WebSocket RPC and indexer feeds.
    - Display pending DvP transaction queue with transaction hash, order ID reference, gas price, mempool entry timestamp, and block timer.
    - Visualize real-time QBFT consensus round progression and block inclusion latency ($T_{\text{confirmed}} - T_{\text{matched}}$).
    - Provide deep links to internal Besu block explorer for every settled lot.

11. **Implement Web Audio Tactile Sound System (`terminalAudio.ts`):**
    - Use Web Audio API oscillators to generate synthetic micro-pings (<10ms playback latency):
      - High chime (880 Hz sine wave): Order fill confirmation.
      - Crisp click (440 Hz short burst): Order placement acknowledged.
      - Low buzzer (150 Hz saw wave): Order rejected / risk collar breach.
      - Double warning klaxon (300 Hz pulse): Master killswitch actuated.
    - Include terminal sound toggle and volume slider in header controls.

12. **Implement Workspace Layout Persistence & Multi-Monitor Sync:**
    - Store active workspace layout, panel sizes, docking locations, and hotkey customizations in `localStorage`.
    - Provide "Save Layout Preset", "Export to JSON", and "Import from JSON" options.
    - Sync layout updates across detached pop-out windows via `BroadcastChannel('dma_workspace_sync')`.

13. **Harden Application Security, CSP, SRI, and Anti-Clickjacking:**
    - Configure HTTP security headers in `next.config.js` and middleware:
      - `Content-Security-Policy`: Restrict `script-src` to self and trusted origins with nonces; `connect-src` to authorized DMA WebSocket and API domains; `worker-src` to `blob:` and self.
      - `X-Frame-Options: DENY` and `frame-ancestors 'none'` to block iframe embedding and clickjacking attacks.
      - Sub-Resource Integrity (SRI) for external bundles and fonts.

14. **Write Comprehensive Automated Unit and Performance Benchmark Tests:**
    - Vitest unit tests for binary WebWorker parsing, hotkey dispatcher logic, and pre-trade safety collar validations.
    - Benchmarking tests verifying WebGL ladder renders 1,000 price tick mutations in under 16ms without memory leaks.
    - Playwright end-to-end tests validating Golden Layout docking, hotkey order placement, cancel-all workflow, and latency ribbon updates.

## Interfaces / Contracts

### WebWorker Binary Message Protocols
```typescript
export type WorkerInboundAction =
  | 'INIT_SOCKET'
  | 'SUBSCRIBE_SYMBOL'
  | 'UNSUBSCRIBE_SYMBOL'
  | 'TERMINATE';

export interface WorkerInboundMessage {
  action: WorkerInboundAction;
  symbol?: string;
  endpointUrl?: string;
  authToken?: string;
}

export type WorkerOutboundType =
  | 'BOOK_SNAPSHOT'
  | 'BOOK_DELTA'
  | 'TRADE_TICK'
  | 'PARSER_METRICS'
  | 'SOCKET_ERROR';

export interface LadderLevel {
  price: number;
  size: number;
  orderCount: number;
  myOrderSize: number;
}

export interface WorkerBookSnapshotPayload {
  symbol: string;
  sequence: number;
  timestampNs: bigint;
  bids: Float64Array; // Pairs: [price, size, orderCount, myOrderSize, ...]
  asks: Float64Array;
}

export interface WorkerBookDeltaPayload {
  symbol: string;
  sequence: number;
  timestampNs: bigint;
  updatedBids: Float64Array;
  updatedAsks: Float64Array;
}

export interface WorkerOutboundMessage {
  type: WorkerOutboundType;
  payload: WorkerBookSnapshotPayload | WorkerBookDeltaPayload | any;
}
```

### Hotkey Management & Action Dispatch Schemas
```typescript
export type HotkeyActionType =
  | 'BUY_MARKET'
  | 'SELL_MARKET'
  | 'BUY_LIMIT_BEST'
  | 'SELL_LIMIT_BEST'
  | 'CANCEL_ALL'
  | 'CANCEL_BIDS'
  | 'CANCEL_ASKS'
  | 'FLATTEN_POSITION'
  | 'PANIC_KILLSWITCH'
  | 'INCREASE_TICK'
  | 'DECREASE_TICK'
  | 'SET_QUANTITY_LOT'
  | 'CENTER_SPREAD'
  | 'TOGGLE_HOTKEY_HUD';

export type HotkeyScope = 'GLOBAL' | 'LADDER' | 'BLOTTER' | 'CHART';

export interface HotkeyBinding {
  id: string;
  key: string; // e.g. "Space", "Shift+Space", "Escape", "ArrowUp"
  scope: HotkeyScope;
  action: HotkeyActionType;
  description: string;
  requiresTwoStrokeConfirm: boolean;
  param?: number | string;
}

export interface HotkeyProfile {
  profileId: string;
  profileName: string;
  isDefault: boolean;
  bindings: HotkeyBinding[];
}
```

### Direct Market Access (DMA) Order Execution & FIX Schemas
```typescript
export type DmaOrderSide = 'BUY' | 'SELL';
export type DmaOrderType = 'LIMIT' | 'MARKET' | 'STOP_LIMIT' | 'PEGGED_BEST';
export type DmaTimeInForce = 'IOC' | 'FOK' | 'DAY' | 'GTC';

export interface DmaOrderRequest {
  clOrdId: string; // UUIDv4 / Institutional Client Order ID
  origClOrdId?: string; // For Cancel / Replace requests
  symbol: string;
  side: DmaOrderSide;
  orderType: DmaOrderType;
  price?: number;
  quantity: number;
  timeInForce: DmaTimeInForce;
  traderId: string;
  accountId: string;
  postOnly?: boolean;
  algoStrategy?: 'NONE' | 'TWAP' | 'VWAP' | 'ICEBERG';
  algoParams?: {
    displayQty?: number;
    durationSeconds?: number;
    maxSlippageBps?: number;
  };
  clientSentTimestampNs: bigint;
}

export interface DmaExecutionReport {
  clOrdId: string;
  orderId: string;
  execId: string;
  symbol: string;
  side: DmaOrderSide;
  orderStatus: 'NEW' | 'PARTIALLY_FILLED' | 'FILLED' | 'CANCELLED' | 'REJECTED' | 'EXPIRED';
  execType: 'NEW' | 'TRADE' | 'CANCEL' | 'REPLACE' | 'REJECT';
  leavesQty: number;
  cumQty: number;
  lastPx: number;
  lastQty: number;
  avgPx: number;
  rejectReason?: string;
  transactTimeNs: bigint;
  roundTripLatencyMs: number;
}
```

### Telemetry, Latency Breakdown & Blockchain Queue Schemas
```typescript
export interface LatencyBreakdown {
  clOrdId: string;
  clientSentTimestamp: number;
  gatewayIngressTimestamp: number;
  riskValidationDurationUs: number;
  matchingEngineQueueDurationUs: number;
  matchingEngineExecutionTimestamp: number;
  clientAckTimestamp: number;
  totalRoundTripMs: number;
  gatewayToClientMs: number;
}

export interface RollingLatencyStats {
  samplesCount: number;
  p50Ms: number;
  p90Ms: number;
  p99Ms: number;
  p999Ms: number;
  maxMs: number;
  jitterMs: number;
}

export interface BesuDvPQueueItem {
  txHash: `0x${string}`;
  orderId: string;
  clOrdId: string;
  symbol: string;
  tradeAmountFiat: string;
  tokenAmount: string;
  buyerAddress: `0x${string}`;
  sellerAddress: `0x${string}`;
  queueStatus: 'MEMPOOL_QUEUED' | 'PROPOSED' | 'MINED_IN_BLOCK' | 'FINALIZED';
  mempoolIngressTime: number;
  blockNumber?: number;
  blockFinalityLatencyMs?: number;
}
```

### Workspace Docking & Layout State Schema
```typescript
export interface PanelConfig {
  id: string;
  type: 'LADDER' | 'BLOTTER' | 'CHART' | 'TIME_AND_SALES' | 'RISK_HUD' | 'LATENCY_MONITOR' | 'BLOCKCHAIN_QUEUE';
  title: string;
  symbol?: string;
  isDetached: boolean;
  coordinates?: { x: number; y: number; width: number; height: number };
}

export interface WorkspaceLayoutConfig {
  layoutVersion: number;
  layoutName: string;
  activeProfileId: string;
  rootLayout: any; // Golden Layout / FlexLayout serialized JSON tree
  panels: PanelConfig[];
  updatedAt: string;
}
```

## Security & Compliance Notes
- **SEBI Algorithmic & Direct Market Access Compliance:** All orders emitted by the workstation incorporate strict institutional tags: `ClOrdID` (Tag 11), `Account` (Tag 1), `ClientID` (Tag 109), and `OrderOrigination` (Tag 1724 = DMA). Client-side order spoofing or layering detection triggers an immediate UI lockout and audit log dispatch.
- **Content Security Policy (CSP):** The Next.js 14 DMA layout enforces strict CSP directives:
  ```http
  Content-Security-Policy: default-src 'self'; script-src 'self' 'nonce-{RANDOM}' 'strict-dynamic'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' wss://dma.growww.in wss://marketdata.growww.in https://dma.growww.in https://rpc.besu.growww.in; worker-src 'self' blob:; frame-ancestors 'none'; object-src 'none'; base-uri 'self';
  ```
- **Anti-Clickjacking Controls:** Prevent terminal hijacking via `X-Frame-Options: DENY` and `Content-Security-Policy: frame-ancestors 'none'`. Include client-side frame-busting guard:
  ```typescript
  if (typeof window !== 'undefined' && window.top !== window.self) {
    window.top.location = window.self.location;
  }
  ```
- **Sub-Resource Integrity (SRI):** All compiled script chunks and font assets must validate against pre-computed cryptographic SHA-384 hashes to mitigate supply-chain injections.
- **Pre-Trade Risk Collars & Fat-Finger Prevention:** Client enforces hard price collars (default $\pm 3\%$ of BBO), maximum single-ticket value limits, and double-keystroke confirmation buffers for destructive actions (`Panic Cancel All`, `Flatten`).
- **Emergency Master Killswitch:** Accessible via `Shift+Escape` or prominent header button, issuing simultaneous bulk cancel requests for all active working orders across all symbols and clearing pending queues within $<10\text{ms}$.
- **Memory Management & Zero-GC Guarantees:** Continuous WebGL rendering requires static object pooling for PixiJS graphics, reusing `Float64Array` buffers, and disposing detached textures to prevent JavaScript garbage collection pause spikes ($>10\text{ms}$).

## Acceptance Criteria
- [ ] Next.js 14 App Router DMA layout (`/dma`) loads the institutional workspace shell in under 800ms.
- [ ] Golden Layout framework supports drag-and-drop docking, panel splitting, tab stacking, and multi-monitor pop-out windows.
- [ ] Pop-out windows maintain bidirectional real-time state synchronization with the parent workstation via `BroadcastChannel`.
- [ ] WebWorker binary decoder parses Level-2 market data off-thread, transferring price buffers to the UI with zero main-thread jank.
- [ ] WebGL PixiJS Price Ladder maintains a steady 60-120 FPS during heavy market data bursts (10,000+ depth updates/second).
- [ ] Single-click on ladder bid/ask price ticks places Limit Orders with immediate visual confirmation markers.
- [ ] Dragging resting order badges on the ladder amends order prices via FIX Cancel/Replace (Tag 35=G) in $<5\text{ms}$.
- [ ] Keyboard hotkey engine processes single-key triggers (`Space`, `Shift+Space`, `Escape`) with $<2\text{ms}$ dispatch latency.
- [ ] Pre-trade risk collars immediately block out-of-bounds fat-finger orders and require two-stroke confirmation for emergency flatten actions.
- [ ] Virtualized execution blotter handles 10,000 orders with instantaneous sorting, filtering, and cancel controls.
- [ ] Latency telemetry ribbon visualizes real-time round-trip breakdown (Browser -> FIX Gateway -> Matching Engine -> Ledger) with rolling percentiles.
- [ ] Blockchain broadcast queue displays pending Besu DvP transactions, mempool stages, and block finality latencies.
- [ ] Web Audio API delivers sub-10ms synthesized auditory feedback for fills, rejects, and killswitch actuation.
- [ ] Strict Content Security Policy, Sub-Resource Integrity, and Anti-Clickjacking headers are fully enforced without console violations.
- [ ] Vitest unit tests and Playwright E2E tests achieve $>90\%$ code coverage across hotkeys, ladder interactions, and telemetry parsing.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 204 (Order Service & Lifecycle Management), Prompt 205 (Low-Latency Matching Engine Core), Prompt 206 (Pre-Trade Risk & Margin Checks Service), Prompt 207 (High-Throughput Market Data Service), Prompt 225 (FIX Protocol Gateway Service), Prompt 601 (Next.js Investor Web App Scaffolding).
- **Parallel Tasks:** Prompt 603 (Web Trading Dashboard), Prompt 610 (Web Options Chain & Multi-Chain Deposit Portal), Prompt 806 (Observability Stack Telemetry).
- **Downstream Blockers:** Prompt 906 (UAT Plan & Regulatory Sandbox Scenarios), Prompt 907 (Regulatory Sandbox Pilot Launch).
