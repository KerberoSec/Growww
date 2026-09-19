# Web Pro Trading Terminal & WebAssembly Streaming Architecture Specification

Document Version: 1.0.0  
Status: Approved Architectural Specification  
Target Workspace: `apps/growww_web`  
Classification: Core Trading Workstation Architecture  

---

## 1. Architectural Overview & System Decomposition

The Web Pro Trading Terminal (`apps/growww_web`) is an institutional-grade, browser-native trading workstation built on Next.js 14 (App Router), React 19 Concurrent Mode, TypeScript, and WebAssembly. It delivers zero-install, zero-compromise trading execution directly inside any modern web browser conforming to the HTML5, WebGL 2.0, WebGPU, and WebAssembly specifications.

The workstation addresses the demanding requirements of high-frequency day traders, institutional scalpers, and quantitative analysts who operate across multiple physical displays, execute orders in single-digit milliseconds via physical key chords, and monitor ultra-dense Level 2/Level 3 market depth updates streaming at up to 100,000 updates per second.

```
+--------------------------------------------------------------------------------------------------------------------+
|                                      WEB PRO TRADING WORKSTATION ARCHITECTURE                                      |
+--------------------------------------------------------------------------------------------------------------------+
|                                                                                                                    |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                          REACT 19 / NEXT.JS 14 CONCURRENT WORKSPACE (MAIN WINDOW)                          |   |
|   |                                                                                                            |   |
|   |   +--------------------------+  +-------------------------------------+  +-----------------------------+   |   |
|   |   |   Multi-Ticker Scanner   |  |   TradingView Advanced Charting     |  |   Level 2 / 3 Order Book    |   |   |
|   |   |   - Sparklines & Volumes |  |   - Multi-Timeframe (1s to 1M)      |  |   - 50-Depth Visual Ladder  |   |   |
|   |   |   - Link Group Sync      |  |   - Custom WASM Datafeed Adapter    |  |   - Tabular Zero-CLS Engine |   |   |
|   |   +--------------------------+  +-------------------------------------+  +-----------------------------+   |   |
|   |                                                                                                            |   |
|   |   +--------------------------+  +-------------------------------------+  +-----------------------------+   |   |
|   |   |   WebGL DOM Ladder       |  |   Zero-Fee Order Entry Bar          |  |   Real-Time Blotter         |   |   |
|   |   |   - Single-Click Trade   |  |   - 0.00% Exchange Fee Invariant    |  |   - Live Position PnL       |   |   |
|   |   |   - Visual Volume Ramps  |  |   - 0 Gas Sponsored Paymaster Badge |  |   - Besu Tx Hash Explorer   |   |   |
|   |   +--------------------------+  +-------------------------------------+  +-----------------------------+   |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                        |                                                                 ^                         |
|                        | BroadcastChannel API (`growww_terminal_sync`)                    | Atomics / SharedMemory   |
|                        v                                                                 |                         |
|   +------------------------------------------+  +----------------------------------------+---------------------+   |
|   |   DETACHED POP-OUT WINDOWS (Multi-Screen) |  |   DEDICATED OFF-THREAD STREAMING ENGINE (Web Worker)        |   |
|   |   - Native OS Browser Windows            |  |   - TLS WebSocket (`wss://ws.growww.in/v1/market/depth`)     |   |
|   |   - Detached 4K TV Charting              |  |   - Rust/WASM Binary Protobuf Streaming Decoder             |   |
|   |   - Independent Scalper DOM Ladder       |  |   - Ring Buffer inside `SharedArrayBuffer`                  |   |
|   |   - Backed by SharedWorker Arbiter       |  |   - 50ms Conflation Throttler (20 FPS UI Dispatch)          |   |
|   +------------------------------------------+  +--------------------------------------------------------------+   |
|                                                                                                                    |
+--------------------------------------------------------------------------------------------------------------------+
```

### 1.1 Invariant Architecture Guarantees
1. **Zero Main-Thread Network Parsing:** All network streaming ingestion, binary WebSocket frames, and Protobuf deserialization execute off the main thread inside a dedicated Web Worker.
2. **Zero Allocation Rendering:** Market depth ladder updates modify contiguous binary views in a pre-allocated `SharedArrayBuffer` ring buffer, preventing JavaScript V8 Garbage Collector pauses during volatility spikes.
3. **Zero Cumulative Layout Shift (CLS = 0):** All financial metrics, price ladders, order books, and trade ticks render using strictly monospaced tabular numerals (`JetBrains Mono`, `tabular-nums`) within fixed-dimension CSS grid constraints.
4. **Zero-Fee Execution Transparency:** All order entry surfaces, validation dialogues, and execution receipts display the strict `0.00% Zero-Fee` model and `0 Gas Fee` ERC-4337 Paymaster sponsorship badge.

---

## 2. Next.js 14 App Router & React 19 Concurrent Mode Framework

The client workspace is architected using Next.js 14 App Router paired with the React 19 Concurrent Mode runtime. The design separates server-side static layout scaffolds from dynamic, client-side trading micro-frontends.

### 2.1 Directory Structure & Route Grouping

The directory structure isolates normal marketing pages from latency-critical pro trading terminals and detached pop-out window shells:

```
apps/growww_web/
├── src/
│   ├── app/
│   │   ├── (auth)/
│   │   │   ├── login/
│   │   │   └── webauthn/
│   │   ├── (marketing)/
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx
│   │   ├── (pro-terminal)/
│   │   │   └── trade/
│   │   │       ├── [symbol]/
│   │   │       │   ├── layout.tsx
│   │   │       │   └── page.tsx
│   │   │       └── layout.tsx
│   │   ├── (popout)/
│   │   │   └── popout/
│   │   │       ├── [panelId]/
│   │   │       │   ├── layout.tsx
│   │   │       │   └── page.tsx
│   │   │       └── layout.tsx
│   │   ├── api/
│   │   │   ├── auth/
│   │   │   │   ├── passkey/
│   │   │   │   └── siwe/
│   │   │   └── user/
│   │   │       └── layouts/
│   │   ├── layout.tsx
│   │   └── globals.css
│   ├── components/
│   │   ├── blotter/
│   │   ├── charting/
│   │   ├── dom/
│   │   ├── hotkeys/
│   │   ├── orderbook/
│   │   ├── orderform/
│   │   ├── popout/
│   │   ├── scanner/
│   │   └── workspace/
│   ├── workers/
│   │   ├── market_depth.worker.ts
│   │   ├── shared_coordinator.worker.ts
│   │   └── wasm/
│   │       ├── growww_protobuf_decoder.wasm
│   │       └── growww_protobuf_decoder.d.ts
│   └── lib/
│       ├── atomics/
│       ├── audio/
│       ├── auth/
│       ├── hotkeys/
│       ├── state/
│       └── webgpu/
```

### 2.2 Server vs Client Component Boundaries

To maximize first-contentful-paint (FCP) while maintaining sub-millisecond interactivity:
- **Server Components (RSC):**
  - Page layouts, metadata generation, OpenGraph tags, and theme bootstrap.
  - Server-side pre-fetching of instrument definitions, tick-size rules, leverage tiers, and trading permissions from exchange REST endpoints.
  - Initial layout preset resolution based on user session cookies.
- **Client Components (`'use client'`):**
  - All panels registered with Dockview (`DockviewWorkspace`).
  - Web Worker and SharedWorker lifecycle instantiators.
  - WebGL 2.0 / WebGPU Canvas rendering views.
  - Keyboard shortcut listeners, audio synthesizers, and WebAuthn authenticators.

### 2.3 React 19 Concurrent Mode Primitives

React 19 Concurrent Mode features are leveraged to maintain deterministic UI responsiveness during market surges:

1. **`useTransition` for Non-Blocking Panel Switching:**
   Switching trading pairs or loading heavy layout presets is wrapped in `startTransition`. This allows high-priority user actions (e.g., keyboard order cancellations) to preempt layout recalculations:
   ```typescript
   const [isPending, startTransition] = useTransition();

   function handleSymbolChange(nextSymbol: string) {
     startTransition(() => {
       setActiveSymbol(nextSymbol);
       updateActiveSubscriptions(nextSymbol);
     });
   }
   ```

2. **`useDeferredValue` for Filtered Scanners and Large Blotters:**
   Watchlist filtering, order history searches, and execution logs utilize `useDeferredValue` so that fast typing in search inputs never drops keystrokes while filtering lists of 1,000+ instruments.

3. **`useActionState` and Server Actions for Secure Configuration:**
   Layout preset saves and custom hotkey bindings utilize React 19 actions with optimistic client-side updates and automatic rollback upon network failure.

4. **Independent Suspense Boundaries:**
   Every Dockview panel is isolated within its own `<Suspense fallback={<PanelSkeleton />}>` boundary. A slowdown or initialization stall in the TradingView chart component cannot block the initialization of the order book or DOM ladder.

---

## 3. Dockview Multi-Split Grid Layout & Workspace Persistence

The terminal utilizes `dockview-core` and `@dockview/react` to deliver an adaptable multi-split grid docking interface comparable to institutional desktop terminals.

```
+--------------------------------------------------------------------------------------------------------------------+
| [Preset: Day Trader v] [Save Layout] [Reset (Ctrl+D)]  [Link: Red v]  [Status: Connected (4ms)] [0.00% Zero-Fee]   |
+--------------------------------------------------+-----------------------------------------------+-----------------+
| Multi-Ticker Watchlist                           | TradingView Advanced Chart (BTC-USDT)         | Order Book L2   |
| [Search Symbol...]                               | [1m] [5m] [15m] [1h] [4h] [1D]                | Price   Size    |
| BTC-USDT  $68,420.00  +3.42%                     |                                               | 68425.5 1.420   |
| ETH-USDT   $3,512.10  +2.15%                     | [====== Candlestick Area ======]              | 68424.0 0.812   |
| SOL-USDT     $182.40  -0.84%                     |                                               | 68423.0 3.190   |
|                                                  |                                               | 68421.5 5.011   |
|                                                  |                                               | --------------- |
|                                                  |                                               | 68420.0 0.050   |
|                                                  |                                               | --------------- |
|                                                  |                                               | 68419.0 2.450   |
|                                                  |                                               | 68418.5 4.112   |
+--------------------------------------------------+-----------------------------------------------+-----------------+
| Execution Blotter / Live Positions                                                               | Instant Order   |
| [Active Orders (2)] [Positions (1)] [Fills (14)] [Audit Trail]                                   | [Buy]    [Sell] |
| Sym      Side  Size      Entry Price   Mark Price    PnL (ROE%)        Status       Tx Hash      | Limit Price:    |
| BTC-USDT LONG  0.500 BTC $67,800.00    $68,420.00    +$310.00 (+4.57%) CONFIRMED    0x8f2a...    | [ 68420.00    ] |
|                                                                                                  | Qty: [ 0.100  ] |
|                                                                                                  | Fee: ₹0.00      |
+--------------------------------------------------------------------------------------------------+-----------------+
```

### 3.1 Panel Registry & Capabilities

Every component registered in the Dockview grid implements a standardized interface:

| Panel Identifier | Primary Component | Rendering Tech | Data Feed Source |
| :--- | :--- | :--- | :--- |
| `chart_advanced` | `TradingViewBridge` | HTML5 Canvas / WebGL | WASM Ring Buffer + Candle History API |
| `order_book_l2` | `WebGLDepthLadder` | WebGL 2.0 / Canvas | `SharedArrayBuffer` (50ms Conflation) |
| `dom_ladder` | `DOMClickToTrade` | HTML5 Canvas / DOM | `SharedArrayBuffer` Direct Slice |
| `order_entry` | `QuickOrderBar` | React 19 Client Form | Local State + Hotkey Dispatcher |
| `execution_blotter` | `ExecutionBlotterTable` | React 19 Tabular Virtualizer | WebSocket Private Stream + Besu RPC |
| `market_scanner` | `MarketWatchlist` | React 19 Virtualized List | Conflated Ticker Broadcast |
| `trades_tape` | `TimeAndSalesStream` | React 19 Circular Buffer | High-Frequency Public Trades WS |
| `telemetry_bar` | `SystemStatusBadge` | React 19 Client Component | Web Worker RTT Heartbeat Monitor |

### 3.2 Workspace Presets

The workstation provides three pre-configured layouts optimized for specific trading workflows:

#### 1. Day Trader Preset
- **Target Audience:** Active intraday momentum and breakout traders.
- **Topology:**
  - Left column (20% width): Ticker Watchlist with sparklines and 24-hour volume stats.
  - Center column (55% width): Split vertically. Top: Large TradingView chart (5m/15m). Bottom: Positions blotter and open orders table.
  - Right column (25% width): Split vertically. Top: Level 2 depth order book. Bottom: Quick order entry form with zero-fee calculators.

#### 2. Scalper Preset
- **Target Audience:** Microsecond liquidity takers and market makers.
- **Topology:**
  - Left column (35% width): Centered Depth of Market (DOM) ladder with single-click bid/ask submission and cumulative depth volume profile.
  - Center column (40% width): High-speed 1-second/tick candle chart stacked above real-time Time & Sales stream.
  - Right column (25% width): Compact Level 2 book, immediate market order controls, hotkey cheat-sheet, and 1-click panic cancellation bar.

#### 3. Analyst Preset
- **Target Audience:** Swing traders, macro technical analysts, and portfolio allocators.
- **Topology:**
  - Left column (15% width): Watchlist categorized by asset class (Crypto, Tokenized Equities, Commodities).
  - Center column (70% width): 2x2 multi-chart grid displaying 15m, 1h, 4h, and 1D charts synchronized with crosshairs.
  - Right column (15% width): Technical screener, funding rates, economic calendar, and historical performance blotter.

### 3.3 State Persistence Pipeline

User workspace states are persisted through a three-tier fallback architecture:

```
[ Dockview Workspace State Changes ]
               |
               +---> 1. In-Memory React State (Instant re-render)
               |
               +---> 2. Web Storage (localStorage + IndexedDB debounced at 300ms)
               |
               +---> 3. Remote Cloud Sync (HTTP PATCH /api/v1/user/layouts debounced at 2000ms)
```

1. **Local Storage Layer:** Layout changes (splits, panel resizing, tab order, floating states) serialize to JSON via `dockviewApi.toJSON()` and store in `localStorage` under key `growww_layout_${presetId}_v1`.
2. **IndexedDB Backup:** Full panel configurations, indicator parameter sets, and drawing tools serialize into an IndexedDB object store (`growww_terminal_db / layouts`).
3. **Cloud Synchronization:** When authenticated via WebAuthn or SIWE, the layout profile synchronizes to the server backend, enabling identical multi-window layouts across physical workstations.
4. **Docking Geometry Constraints:** Panels enforce strict min/max boundary constraints (e.g., Order Book min-width: 280px; Order Form min-width: 260px) preventing UI overlap or unreadable squishing.

---

## 4. Multi-Monitor Pop-Out Windowing System

Active traders require multi-screen trading environments. The terminal supports unconstrained detachment of any panel into a native browser window spawned via `window.open()`, retaining bidirectional communication with the main coordinator window.

```
+--------------------------------------------------------------------------------------------------------------------+
|                                    MULTI-MONITOR SYNCHRONIZATION TOPOLOGY                                          |
+--------------------------------------------------------------------------------------------------------------------+
|                                                                                                                    |
|   +------------------------------------+             +------------------------------------+                        |
|   |         PHYSICAL MONITOR 1         |             |         PHYSICAL MONITOR 2         |                        |
|   |  MAIN BROWSER WINDOW               |             |  POP-OUT WINDOW A (Detached Chart) |                        |
|   |  Route: /trade/BTC-USDT            |             |  Route: /popout/chart_advanced     |                        |
|   |  - Dockview Layout Coordinator     |             |  - Fullscreen 4K TradingView Canvas|                        |
|   |  - Web Worker Host                 |             |  - Subscribed to Link Group RED    |                        |
|   |  - Active Link Group: RED          |             |  - Independent Viewport Controls   |                        |
|   +------------------------------------+             +------------------------------------+                        |
|                     ^                                                  ^                                           |
|                     |                                                  |                                           |
|                     |             BroadcastChannel API                 |                                           |
|                     +==================================================+                                           |
|                     |          Name: 'growww_terminal_sync'            |                                           |
|                     |                                                  |                                           |
|                     v                                                  v                                           |
|   +------------------------------------+             +------------------------------------+                        |
|   |         PHYSICAL MONITOR 3         |             |         SHARED BACKING LAYER       |                        |
|   |  POP-OUT WINDOW B (Detached DOM)   |             |  SharedWorker Coordinator          |                        |
|   |  Route: /popout/dom_ladder         | <=========> |  Route: /workers/shared.worker.ts  |                        |
|   |  - WebGL Depth of Market Ladder    |             |  - Cross-Window Leader Election    |                        |
|   |  - Subscribed to Link Group BLUE   |             |  - Fallback State Broker           |                        |
|   |  - Hardware Click-to-Trade Engine  |             |  - Single WebSocket Uplink Arbiter |                        |
|   +------------------------------------+             +------------------------------------+                        |
|                                                                                                                    |
+--------------------------------------------------------------------------------------------------------------------+
```

### 4.1 Window Detachment Lifecycle

When the user clicks the "Pop Out" button on any Dockview panel header:
1. The panel state, active symbol, timeframe, and link group are captured.
2. The main window executes:
   ```typescript
   const popoutUrl = `/popout/${panelId}?symbol=${encodeURIComponent(symbol)}&group=${linkGroup}&origin=${encodeURIComponent(window.location.origin)}`;
   const features = 'popup=yes,menubar=no,toolbar=no,status=no,width=1280,height=800,left=100,top=100';
   const childWindow = window.open(popoutUrl, `growww_popout_${panelId}_${Date.now()}`, features);
   ```
3. The main window replaces the docked panel with a placeholder: "Panel detached to external window. [Re-dock Panel]".
4. When the pop-out window unloads (`beforeunload` event), it posts a `DETACHED_WINDOW_CLOSING` message, prompting the main coordinator window to re-dock the panel back into the primary Dockview grid.

### 4.2 Cross-Window Synchronization via BroadcastChannel

All windows instantiate a shared `BroadcastChannel('growww_terminal_sync')`. Messages adhere to a strict typed protocol:

```typescript
export type TerminalSyncMessage =
  | {
      type: 'SYMBOL_CHANGE';
      sourceId: string;
      linkGroup: 'red' | 'blue' | 'green' | 'yellow' | 'neutral';
      symbol: string;
      timestampNs: number;
    }
  | {
      type: 'ORDER_SUBMITTED';
      sourceId: string;
      clientOrderId: string;
      symbol: string;
      side: 'BUY' | 'SELL';
      orderType: 'LIMIT' | 'MARKET' | 'STOP_LOSS';
      price: string;
      quantity: string;
      timestampNs: number;
    }
  | {
      type: 'ORDER_FILLED';
      sourceId: string;
      orderId: string;
      symbol: string;
      side: 'BUY' | 'SELL';
      filledQty: string;
      fillPrice: string;
      timestampNs: number;
    }
  | {
      type: 'PANIC_CANCEL_TRIGGERED';
      sourceId: string;
      symbol?: string; // If undefined, global cancel across all pairs
      timestampNs: number;
    }
  | {
      type: 'HEARTBEAT';
      sourceId: string;
      windowRole: 'COORDINATOR' | 'POPOUT';
      timestampNs: number;
    };
```

### 4.3 Link Group Synchronization Channels

Panels can be tagged with one of four color-coded Link Groups:
- **Group Red:** Primary execution instrument (default: `BTC-USDT`).
- **Group Blue:** Secondary momentum pair (e.g., `ETH-USDT`).
- **Group Green:** Arbitrage / Hedge pair.
- **Group Yellow:** Macro index / Tokenized RWA pair.
- **Neutral (Grey):** Independent; ignores external symbol broadcast events.

When a trader clicks `SOL-USDT` in a Red-grouped watchlist:
1. The watchlist dispatches a `SYMBOL_CHANGE` event with `linkGroup: 'red'` to the `BroadcastChannel`.
2. Popped-out charts and depth ladders on secondary monitors tagged with `linkGroup: 'red'` intercept the event and transition to `SOL-USDT` within 2 milliseconds.
3. Panels tagged with `blue` or `neutral` remain untouched.

### 4.4 SharedWorker Coordinator & Leader Election

In multi-window configurations, maintaining separate WebSocket connections in each window leads to duplicated bandwidth and rate-limiting issues. A backing `SharedWorker` (`shared_coordinator.worker.ts`) acts as the state arbiter:
- **Leader Election:** The main browser window asserts itself as the `COORDINATOR`. If the main window closes, one of the pop-out windows is elected coordinator via Bully Algorithm over `SharedWorker` ports.
- **Connection Multiplexing:** The worker maintains a single persistent WebSocket stream for shared market data, fanning binary payloads out to connected window ports via transferable buffers.

---

## 5. Dedicated Web Worker & WebAssembly Protobuf Streaming Pipeline

To eliminate UI thread micro-stutters and garbage collection pauses during extreme volatility bursts (e.g., 100,000 depth deltas per second during market liquidation cascades), all network ingestion, binary parsing, and depth book aggregation execute off-thread.

```
+--------------------------------------------------------------------------------------------------------------------+
|                                    OFF-THREAD STREAMING & MEMORY PIPELINE                                          |
+--------------------------------------------------------------------------------------------------------------------+
|                                                                                                                    |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                                          DEDICATED WEB WORKER THREAD                                       |   |
|   |                                                                                                            |   |
|   |   +--------------------------------+       +-----------------------------------------------------------+   |   |
|   |   | TLS WebSocket Client           |       | WebAssembly Protobuf Decoder Module                       |   |   |
|   |   | wss://ws.growww.in/v1/market/..| ----> | - Rust compiled to WASM via wasm-bindgen / prost          |   |   |
|   |   | - Binary stream (ArrayBuffer)  |       | - Zero JS allocation during message decode                |   |   |
|   |   +--------------------------------+       +-----------------------------------------------------------+   |   |
|   |                                                                  |                                         |   |
|   |                                                                  v                                         |   |
|   |   +----------------------------------------------------------------------------------------------------+   |   |
|   |   | In-Worker Order Book Engine (L2 / L3 Red-Black Tree & BBO Aggregator)                              |   |   |
|   |   | - Applies bid/ask delta updates                                                                    |   |   |
|   |   | - Computes cumulative depth volumes                                                                |   |   |
|   |   +----------------------------------------------------------------------------------------------------+   |   |
|   |                                                                  |                                         |   |
|   |                                                                  v                                         |   |
|   |   +----------------------------------------------------------------------------------------------------+   |   |
|   |   | 50ms Conflation Throttler (20 FPS UI Engine)                                                       |   |   |
|   |   | - Flushes snapshots into SharedArrayBuffer memory segment                                          |   |   |
|   |   | - Executes Atomics.store() and Atomics.notify()                                                    |   |   |
|   |   +----------------------------------------------------------------------------------------------------+   |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                      |                                                             |
|                                                      v                                                             |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                                       SHAREDARRAYBUFFER MEMORY REGION                                      |   |
|   |  [ Header (32B) ] [ Top 50 Bids: Price, Qty, Count (1200B) ] [ Top 50 Asks: Price, Qty, Count (1200B) ]    |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                      |                                                             |
|                                                      v                                                             |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                                       MAIN UI THREAD (WebGL Canvas / React 19)                             |   |
|   |  - requestAnimationFrame() renders directly from SharedArrayBuffer view                                   |   |
|   |  - Zero V8 Garbage Collection overhead; locked 60 / 120 FPS rendering                                      |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                                                                                    |
+--------------------------------------------------------------------------------------------------------------------+
```

### 5.1 Protobuf Binary Schema Specification

Market depth updates stream as binary Protocol Buffers (`market_depth.proto`):

```protobuf
syntax = "proto3";

package growww.market.v1;

enum DepthAction {
  DEPTH_ACTION_UNSPECIFIED = 0;
  DEPTH_ACTION_SNAPSHOT = 1;
  DEPTH_ACTION_UPDATE = 2;
  DEPTH_ACTION_DELETE = 3;
}

enum OrderSide {
  ORDER_SIDE_UNSPECIFIED = 0;
  ORDER_SIDE_BID = 1;
  ORDER_SIDE_ASK = 2;
}

message DepthLevel {
  uint64 price_raw = 1;      // Scaled by 10^8 (e.g. 68420.50 -> 6842050000000)
  uint64 quantity_raw = 2;   // Scaled by 10^8 (e.g. 1.25000000 -> 125000000)
  uint32 order_count = 3;    // Resting order count at price level
  uint32 flags = 4;          // Bitfield: [bit 0: implied liquidity, bit 1: updated]
}

message MarketDepthMessage {
  string symbol = 1;
  uint64 sequence_number = 2;
  uint64 timestamp_ns = 3;
  DepthAction action = 4;
  repeated DepthLevel bids = 5;
  repeated DepthLevel asks = 6;
  uint64 best_bid_price = 7;
  uint64 best_ask_price = 8;
  uint64 best_bid_qty = 9;
  uint64 best_ask_qty = 10;
}
```

### 5.2 WebAssembly (WASM) Binary Decoder

The Web Worker compiles a high-performance WebAssembly module built in Rust:
- **Zero-Copy Byte Decoding:** Rust's `prost` library decodes directly from the incoming WebSocket byte slice into WebAssembly linear memory without intermediate heap string or object allocations.
- **Fast Numerics:** 64-bit integer fixed-point arithmetic (`price_raw`, `quantity_raw`) is handled natively in WASM via 64-bit integer CPU instructions (`i64`), bypassing JavaScript floating-point inaccuracy.

### 5.3 SharedArrayBuffer Ring Buffer Memory Layout

To eliminate postMessage serialization latency, market depth state is stored in a `SharedArrayBuffer` shared between the Web Worker and the UI thread. The memory layout is strictly structured:

```
SharedArrayBuffer Memory Segment (Total: 4,096 Bytes per Trading Pair):

+-----------------------+---------------------+---------------------------------------------------------------+
| Byte Offset           | Data Type           | Field Description                                             |
+-----------------------+---------------------+---------------------------------------------------------------+
| 0x0000 - 0x0007       | uint64 (Atomic)     | Sequence Number (Monotonically increasing)                     |
| 0x0008 - 0x000F       | uint64              | Timestamp Nanoseconds (Engine matching time)                  |
| 0x0010 - 0x0013       | uint32              | Total Active Bid Levels Count ($N \le 50$)                    |
| 0x0014 - 0x0017       | uint32              | Total Active Ask Levels Count ($M \le 50$)                    |
| 0x0018 - 0x001B       | uint32 (Atomic)     | Conflation Frame Counter (Increments on 50ms flush)           |
| 0x001C - 0x001F       | uint32              | Status Flags (Bit 0: Crossed Book, Bit 1: Stale Warning)      |
| 0x0020 - 0x04CF       | 50 x 24-byte Struct | Bids Array (Sorted Descending by Price)                       |
| 0x04D0 - 0x097F       | 50 x 24-byte Struct | Asks Array (Sorted Ascending by Price)                        |
| 0x0980 - 0x0FFF       | Reserved            | Alignment padding & telemetry metrics                         |
+-----------------------+---------------------+---------------------------------------------------------------+

Single Level Struct Layout (24 Bytes):
+-----------------------+---------------------+---------------------------------------------------------------+
| Offset                | Data Type           | Field Description                                             |
+-----------------------+---------------------+---------------------------------------------------------------+
| +0x00                 | uint64              | Price Raw ($P \times 10^8$)                                   |
| +0x08                 | uint64              | Quantity Raw ($Q \times 10^8$)                                |
| +0x10                 | uint32              | Order Count                                                   |
| +0x14                 | uint32              | Cumulative Quantity Raw ($Q_{cum} \times 10^8$)                |
+-----------------------+---------------------+---------------------------------------------------------------+
```

### 5.4 Atomics Synchronization & 50ms Conflation Engine

To prevent browser UI threads from locking under high tick volume:
1. The Web Worker updates its internal order book in real-time.
2. Every 50 milliseconds (20 FPS update cadence), the worker writes the top 50 bids and asks into the `SharedArrayBuffer` memory region.
3. The worker issues an atomic store and notification:
   ```typescript
   Atomics.add(atomicHeaderView, ATOMIC_FRAME_COUNTER_INDEX, 1);
   Atomics.notify(atomicHeaderView, ATOMIC_FRAME_COUNTER_INDEX, 1);
   ```
4. The main thread's rendering loop checks for updates via `requestAnimationFrame`:
   ```typescript
   function renderLoop() {
     const currentFrame = Atomics.load(atomicHeaderView, ATOMIC_FRAME_COUNTER_INDEX);
     if (currentFrame !== lastRenderedFrame) {
       renderOrderBookCanvas(sharedBufferView);
       lastRenderedFrame = currentFrame;
     }
     requestAnimationFrame(renderLoop);
   }
   ```
5. If Cross-Origin Isolation (`Cross-Origin-Opener-Policy: same-origin` and `Cross-Origin-Embedder-Policy: require-corp`) is unavailable on older browsers, the worker automatically falls back to Transferable `ArrayBuffer` structured cloning over standard Web Worker messaging.

---

## 6. WebGL 2.0 / WebGPU Canvas Rendering Engine & TradingView Integration

Rendering microsecond depth changes and 100+ indicator technical charts requires hardware acceleration. DOM manipulation using standard HTML `<div>` or `<table>` nodes produces unacceptable reflow penalties; therefore, high-density visualization utilizes WebGL 2.0 / WebGPU with monospaced tabular typography.

```
+--------------------------------------------------------------------------------------------------------------------+
|                                      HARDWARE-ACCELERATED RENDERING PIPELINE                                       |
+--------------------------------------------------------------------------------------------------------------------+
|                                                                                                                    |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                            HTML5 Canvas Element (`ref={canvasRef}`)                                        |   |
|   |                            TransferControlToOffscreen() (Optional Off-Thread Render)                       |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                      |                                                             |
|                                                      v                                                             |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                                         WEBGL 2.0 / WEBGPU PIPELINE                                        |   |
|   |                                                                                                            |   |
|   |   +--------------------------------------+       +-----------------------------------------------------+   |   |
|   |   | Vertex Buffer Objects (VBO)          |       | Instanced Shader Program                            |   |   |
|   |   | - 50 Quad Rectangles for Price Rows  | ----> | - Dynamic Color Grading:                            |   |   |
|   |   | - Normalized Coordinates: [-1.0, 1.0]|       |   Bids: #00D09C (Growww Emerald Green)              |   |   |
|   |   | - Instance Data: Depth Percentages   |       |   Asks: #EB5757 (Institutional Crimson Red)         |   |   |
|   |   +--------------------------------------+       |   Alpha Ramp: Depth Volume Density                  |   |   |
|   |                                                  +-----------------------------------------------------+   |   |
|   |                                                                             |                              |   |
|   |                                                                             v                              |   |
|   |   +----------------------------------------------------------------------------------------------------+   |   |
|   |   | Texture Atlas Glyph Renderer                                                                       |   |   |
|   |   | - Pre-baked JetBrains Mono Numeric Atlas (0-9, ., ,, +, -, K, M, B)                                |   |   |
|   |   | - Fixed 14px cell bounds; tabular numbers; 0 layout shift                                          |   |   |
|   |   +----------------------------------------------------------------------------------------------------+   |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                      |                                                             |
|                                                      v                                                             |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                               HARDWARE RASTERIZER (GPU Framebuffer: 60/120 FPS)                           |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                                                                                    |
+--------------------------------------------------------------------------------------------------------------------+
```

### 6.1 WebGL 2.0 Shader Specifications for Depth Ramps

The Depth Ladder visualizes market volume using an instanced quad shader. Each depth level generates a horizontal bar with width proportional to cumulative quantity:

#### Vertex Shader (`depth_ladder.vert`):
```glsl
#version 300 es
layout (location = 0) in vec2 a_position;      // Standard Quad: [0,0] to [1,1]
layout (location = 1) in vec2 a_instance_pos;  // Row position [y_offset, row_height]
layout (location = 2) in float a_fill_ratio;   // Depth fill width: [0.0, 1.0]
layout (location = 3) in float a_side;         // 0.0 = Bid (Green), 1.0 = Ask (Red)

out vec2 v_uv;
out float v_side;
out float v_fill_ratio;

uniform mat4 u_projection;

void main() {
    v_uv = a_position;
    v_side = a_side;
    v_fill_ratio = a_fill_ratio;
    
    // Scale X by the fill ratio, position Y according to row index
    vec2 scaled_pos = vec2(a_position.x * a_fill_ratio, a_position.y * a_instance_pos.y) + vec2(0.0, a_instance_pos.x);
    gl_Position = u_projection * vec4(scaled_pos, 0.0, 1.0);
}
```

#### Fragment Shader (`depth_ladder.frag`):
```glsl
#version 300 es
precision highp float;

in vec2 v_uv;
in float v_side;
in float v_fill_ratio;

out vec4 fragColor;

void main() {
    vec4 bidColor = vec4(0.0, 0.815, 0.612, 0.25); // Emerald Green #00D09C with 25% alpha
    vec4 askColor = vec4(0.922, 0.341, 0.341, 0.25); // Crimson Red #EB5757 with 25% alpha
    
    vec4 baseColor = mix(bidColor, askColor, v_side);
    
    // Subtle horizontal gradient to accentuate depth edge
    float gradient = smoothstep(0.0, 1.0, v_uv.x);
    fragColor = vec4(baseColor.rgb, baseColor.a * (0.6 + 0.4 * gradient));
}
```

### 6.2 TradingView Advanced Charts Bridge

The charting widget integrates TradingView's Advanced Charting Library (JS API) with a bespoke datafeed bridge:
1. **Custom Datafeed Interface (`IExternalDatafeed`):**
   - Implements `getBars`, `subscribeBars`, `resolveSymbol`, and `getMarks`.
   - Historical candle bars (1s, 1m, 5m, 15m, 1h, 4h, 1D) are retrieved from `/api/v1/market/candles` with LZ4 binary compression.
2. **Real-Time Bar Ingestion:**
   - Real-time trade executions streaming from the WASM worker bypass full chart re-renders, updating the active bar directly via `datafeed.updateBar(bar)`.
3. **Crosshair Coherence:**
   - Crosshair movements on detached popped-out chart windows broadcast cursor coordinates ($t, P$) over `growww_terminal_sync`, synchronizing crosshair lines across all open monitors.

### 6.3 Zero Cumulative Layout Shift (CLS = 0) Invariant

Financial UI elements guarantee strict zero layout shifts during high-frequency data updates:
- **Font Face:** `JetBrains Mono`, `Roboto Mono`, or system monospaced stack.
- **CSS OpenType Attributes:** `font-feature-settings: "tnum" 1, "zero" 1; font-variant-numeric: tabular-nums;`.
- **Dimensional Reservations:** Table cells use strict pixel widths (`w-[100px]`, `min-w-[100px]`, `max-w-[100px]`) with `overflow: hidden` and `text-overflow: clip`. Numeric values are right-aligned with fixed trailing zeros matching instrument tick precision (e.g., `68,420.50`).

---

## 7. WebAuthn FIDO2 Passkeys, Non-Custodial Web3 & SIWE Authentication

Security and authentication in the Pro Terminal combine hardware-backed WebAuthn FIDO2 biometrics with non-custodial Ethereum/Besu Web3 wallet connections.

```
+--------------------------------------------------------------------------------------------------------------------+
|                                          AUTHENTICATION ARCHITECTURE                                               |
+--------------------------------------------------------------------------------------------------------------------+
|                                                                                                                    |
|   +---------------------------------------------+       +------------------------------------------------------+   |
|   |        WEBAUTHN / FIDO2 PASSKEYS            |       |           WEB3 / EIP-4361 SIWE                       |   |
|   |  - Platform Biometrics (Touch ID, Hello)    |       |  - Non-Custodial (MetaMask, WalletConnect v2)        |   |
|   |  - Hardware Security Keys (YubiKey)         |       |  - Sign-In with Ethereum (EIP-4361) Standard         |   |
|   |  - Client-Side PRF Key Derivation           |       |  - EIP-712 Structured Typed Signatures               |   |
|   +---------------------------------------------+       +------------------------------------------------------+   |
|                          |                                                         |                               |
|                          v                                                         v                               |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                                     UNIFIED EXCHANGE AUTHENTICATION GATEWAY                                |   |
|   |  - Validates FIDO2 Signature or Recovered Web3 Address                                                     |   |
|   |  - Evaluates Risk Engine & Session Expiry                                                                 |   |
|   |  - Generates Ephemeral HTTP-Only JWT Session Cookie + WebSocket Ticket                                     |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                      |                                                             |
|                                                      v                                                             |
|   +------------------------------------------------------------------------------------------------------------+   |
|   |                                ERC-4337 ACCOUNT ABSTRACTION SESSION KEYS                                   |   |
|   |  - Client generates local ephemeral ECDSA keypair in memory                                                |   |
|   |  - User signs single UserOperation authorizing session key for 8 hours (with strict limits)                |   |
|   |  - Enables sub-millisecond hotkey trading WITHOUT wallet signature popups on every order                   |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                                                                                    |
+--------------------------------------------------------------------------------------------------------------------+
```

### 7.1 WebAuthn / FIDO2 Passkey Specification

Traders authenticate biometrically in under 500 milliseconds without typing passwords or SMS 2FA codes:
1. **Challenge Request:** Client calls `POST /api/v1/auth/passkey/challenge`, receiving a cryptographically secure 32-byte random challenge.
2. **Credential Assertion:** Browser calls `navigator.credentials.get()`:
   ```typescript
   const credential = await navigator.credentials.get({
     publicKey: {
       challenge: base64ToArrayBuffer(challengeData.challenge),
       rpId: window.location.hostname,
       userVerification: 'required',
       timeout: 60000,
       extensions: {
         prf: { eval: { first: new Uint8Array(32) } } // Pseudo-Random Function extension for enclave key derivation
       }
     }
   }) as PublicKeyCredential;
   ```
3. **Verification:** The exchange gateway verifies the authenticator data, client data JSON, and ECDSA signature against the user's stored public key.

### 7.2 Sign-In with Ethereum (EIP-4361 SIWE) Integration

For decentralized custody and non-custodial trading:
1. **Message Construction:** When connecting via MetaMask, WalletConnect v2, or Coinbase Wallet, a compliant EIP-4361 message is constructed:
   ```
   growww.in wants you to sign in with your Ethereum account:
   0x71C...3a9

   Sign in to Growww Pro Trading Terminal. Zero-Fee Trading Invariant Active.

   URI: https://growww.in/trade
   Version: 1
   Chain ID: 133717
   Nonce: 98a7bc6d5e4f3a21
   Issued At: 2026-09-19T17:47:43.000Z
   Expiration Time: 2026-09-20T01:47:43.000Z
   ```
2. **EIP-191 / EIP-712 Signature:** The wallet signs the message. The client transmits the signature and raw message to `/api/v1/auth/siwe/verify`.
3. **Session Issuance:** The server verifies `ecrecover(hash, signature) == address` and issues an authorized session ticket.

### 7.3 ERC-4337 Session Keys for Frictionless Hotkey Execution

To enable instant hotkey trading (e.g., `Shift + B`) without prompting MetaMask popups for every order:
1. Upon session initialization, the client generates an ephemeral, in-memory secp256k1 keypair.
2. The user executes a single one-time signature approving a **Session Key** on their Smart Contract Account (ERC-4337):
   - **Allowed Contracts:** Growww Settlement Engine & CLOB Router.
   - **Allowed Methods:** `submitOrder`, `cancelOrder`, `batchCancel`.
   - **Max Exposure:** Up to $50,000 equivalent.
   - **Validity Window:** Exactly 8 hours.
3. Subsequent hotkey order submissions sign off-chain EIP-712 order structs using the in-memory session key, transmitting them directly to the matching engine WebSocket gateway in under 2 milliseconds.

---

## 8. Ultra-Low Latency Keyboard Hotkey Engine

Professional traders operate the terminal almost exclusively via physical keyboard commands. The hotkey engine operates at the window capture phase to achieve immediate response times.

```
+--------------------------------------------------------------------------------------------------------------------+
|                                          KEYBOARD HOTKEY EVENT ENGINE                                              |
+--------------------------------------------------------------------------------------------------------------------+
|                                                                                                                    |
|   [ Physical Keypress Event ]                                                                                      |
|              |                                                                                                     |
|              v                                                                                                     |
|   +------------------------------------------------------------------------------------------------------------+   |
|   | Global Window Event Listener (Capture Phase: window.addEventListener('keydown', handler, true))            |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|              |                                                                                                     |
|              +---> Is Focus inside Input, Textarea, or ContentEditable?                                            |
|              |     - YES: Allow default keystroke UNLESS modifier is pressed (e.g. Escape or Shift+Escape)         |
|              |     - NO:  Proceed to Hotkey Dispatch Table                                                         |
|              v                                                                                                     |
|   +------------------------------------------------------------------------------------------------------------+   |
|   | Hotkey Dispatch Table (O(1) Map Lookup)                                                                    |   |
|   | - Shift + B        --> Trigger Instant Market Buy                                                          |   |
|   | - Shift + S        --> Trigger Instant Market Sell                                                         |   |
|   | - Escape           --> Panic Cancel Current Symbol Open Orders                                             |   |
|   | - Shift + Escape   --> Global Panic Cancel All Symbols Across Exchange                                    |   |
|   | - Space            --> Open Quick-Search Ticker Command Palette                                            |   |
|   | - 1 / 2 / 5 / 0    --> Set Lot Sizing to 10% / 25% / 50% / 100% Margin                                     |   |
|   | - C                --> Center DOM Ladder at Last Traded Price (LTP)                                        |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|              |                                                                                                     |
|              +---> 1. Prevent Default & Stop Immediate Propagation                                                |
|              +---> 2. Execute Web Audio Synthesizer (1ms Tactile Audio Feedback)                                   |
|              +---> 3. Dispatch Order / Action via Session Key directly to WebSocket                                |
|              +---> 4. Trigger Visual Flash on Corresponding Screen Element                                         |
|                                                                                                                    |
+--------------------------------------------------------------------------------------------------------------------+
```

### 8.1 Hotkey Bindings Matrix

| Key Combination | Scope | Action Performed | Safety Guard |
| :--- | :--- | :--- | :--- |
| `Shift + B` | Global | Instant Market Buy at best available offer | Configurable one-click confirmation toggle |
| `Shift + S` | Global | Instant Market Sell at best available bid | Configurable one-click confirmation toggle |
| `Escape` | Local Symbol | Panic cancel all active resting orders for active pair | Immediate execution; no confirmation modal |
| `Shift + Escape` | Global | Panic cancel all active orders across all pairs | Immediate execution; atomic batch cancellation |
| `Space` | Global | Focus/Open Spotlight Ticker Search Palette | Suppressed if search palette already active |
| `1` | Order Entry | Allocate 10% of available free margin | Applies to currently active order entry form |
| `2` | Order Entry | Allocate 25% of available free margin | Applies to currently active order entry form |
| `5` | Order Entry | Allocate 50% of available free margin | Applies to currently active order entry form |
| `0` | Order Entry | Allocate 100% of available free margin (Max Size) | Applies to currently active order entry form |
| `Arrow Up` | Order Entry | Increment Limit Price by 1 Tick Size | Continuous press auto-repeats at 20Hz |
| `Arrow Down` | Order Entry | Decrement Limit Price by 1 Tick Size | Continuous press auto-repeats at 20Hz |
| `C` | DOM Ladder | Re-center DOM depth ladder at Last Traded Price | Instantly re-aligns ladder to center row |
| `F11` | Window | Toggle native OS fullscreen mode | Browser permission compliant |
| `Ctrl + D` | Workspace | Reset workspace layout to default Day Trader preset | Reversible via Undo prompt |
| `?` | Global | Display Keyboard Hotkeys Cheat-Sheet Modal | Auto-dismisses on second press or Escape |

### 8.2 Quick-Search Ticker Command Palette (Spacebar)

Pressing `Space` (when not editing an input) opens a modal command bar:
- **Instant Search:** Fuzzy matching across symbol names (`BTC-USDT`), full asset descriptions ("Bitcoin"), and asset classes.
- **Micro-Metrics Display:** Shows 24h change, volume, mark price, and sparkline for each candidate.
- **Keyboard Navigation:** Navigate with `Arrow Down` / `Arrow Up`, confirm with `Enter` to switch the active instrument across all panels linked to the current Link Group.

### 8.3 Tactile Audio Feedback Engine

Physical feedback reduces human error during rapid order entry. The terminal incorporates a lightweight, client-side Web Audio API synthesizer:
- **Order Placed:** 15ms sine burst at 880Hz (A5) fading to 440Hz (A4).
- **Order Filled:** Crisp harmonic chime at 1200Hz + 1600Hz.
- **Panic Cancel:** Distinctive 40ms low-frequency square wave burst at 220Hz.
- **Volume / Mute Control:** Traders can adjust volume or disable sound effects in system settings.

---

## 9. Strict 0.00% Zero-Fee Presentation & Gas Sponsorship Invariant

A fundamental principle of the Growww exchange platform is the **Genesis 0.00% Zero-Fee Invariant** and **Zero-Gas Sponsorship**. The terminal must communicate this invariant unambiguously across all user-facing interfaces.

```
+--------------------------------------------------------------------------------------------------------------------+
|                                    ZERO-FEE & SPONSORSHIP BADGE SPECIFICATION                                      |
+--------------------------------------------------------------------------------------------------------------------+
|                                                                                                                    |
|   +------------------------------------------------------------------------------------------------------------+   |
|   | ORDER ENTRY FORM BADGE                                                                                     |   |
|   | +--------------------------------------------------------------------------------------------------------+ |   |
|   | | Trading Fee: ₹0.00 (0.00% Genesis Zero-Fee)               [? Verified Zero Brokerage]                  | |   |
|   | +--------------------------------------------------------------------------------------------------------+ |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                                                                                    |
|   +------------------------------------------------------------------------------------------------------------+   |
|   | ORDER CONFIRMATION MODAL AUDIT TABLE                                                                       |   |
|   | +-----------------------------------------------------+--------------------------------------------------+ |   |
|   | | Field Description                                   | Presentation Value                               | |   |
|   | +-----------------------------------------------------+--------------------------------------------------+ |   |
|   | | Gross Order Value (0.500 BTC @ $68,420.00)          | ₹28,56,535.00                                    | |   |
|   | | Exchange Trading Fee (Maker/Taker)                  | ₹0.00 (0.00% Free)                               | |   |
|   | | Clearing & Settlement Fee                           | ₹0.00 (Free)                                     | |   |
|   | | On-Chain Gas Fee (Hyperledger Besu Network)         | ₹0.00 [Sponsored by Paymaster Badge]             | |   |
|   | | Statutory Withholding Tax (Section 194S TDS)        | Applicable statutory TDS (Separately Itemized)   | |   |
|   | | Total Out-of-Pocket Transaction Fee                 | ₹0.00                                            | |   |
|   | +-----------------------------------------------------+--------------------------------------------------+ |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                                                                                    |
|   +------------------------------------------------------------------------------------------------------------+   |
|   | EXECUTION BLOTTER AUDIT TRAIL BADGES                                                                       |   |
|   | [Tx: 0x8f2a...c4e1] [Gas: 0 Gwei / Sponsored] [Fee: ₹0.00] [Status: Settled on Besu Block #4,192,840]      |   |
|   +------------------------------------------------------------------------------------------------------------+   |
|                                                                                                                    |
+--------------------------------------------------------------------------------------------------------------------+
```

### 9.1 Zero-Fee Mathematical Invariants

In order entry calculations, the trading fee calculation is mathematically invariant:

$$Fee_{exchange} = Notional \times 0.00000000 = 0.00$$

Where:
- $Notional = Price \times Quantity$.
- Under no circumstances may maker or taker fees evaluate to a non-zero number.
- Any attempt by API responses to inject a non-zero trading fee results in an immediate client-side error alert (`ERR_ZERO_FEE_INVARIANT_VIOLATION`) and halts order dispatch.

### 9.2 ERC-4337 Paymaster Gas Sponsorship Verification

On-chain settlement transactions processed through Hyperledger Besu are sponsored by the exchange Paymaster contract:
1. Every order settlement transaction executed via Account Abstraction includes a `paymasterAndData` payload referencing the Growww Paymaster.
2. The user pays exactly 0 gas from their wallet:
   $$\text{GasCost}_{user} = 0.00000000 \text{ ETH / MATIC / GWEI}$$
3. The blotter and order receipt link directly to the transaction on the block explorer, highlighting:
   - Green Shield Badge: `Paymaster Sponsored`.
   - Effective Gas Paid by User: `0.00 Gwei (₹0.00)`.

### 9.3 Statutory Tax Separation

Statutory taxes required by local regulations (e.g., Indian Income Tax Act Section 194S 1% TDS on crypto transfers) must be presented with clear itemization:
- Statutory TDS is strictly distinguished from platform fees:
  - Platform Fee: `₹0.00 (0.00% Zero-Fee)`.
  - Statutory Withholding: `1.00% Section 194S TDS (Remitted to Tax Authority)`.
- The user is provided with a downloadable tax deduction receipt containing the TAN and transaction hash.

---

## 10. Security, Concurrency & Production Reliability Verification

To maintain security and reliability under continuous market operation, the client architecture enforces strict browser security controls and verification criteria.

### 10.1 Cross-Origin Isolation Headers

To enable `SharedArrayBuffer` without browser security restrictions, the Next.js server serves the following security headers on all terminal routes (`/trade/*`, `/popout/*`):

```http
Cross-Origin-Opener-Policy: same-origin
Cross-Origin-Embedder-Policy: require-corp
Content-Security-Policy: default-src 'self'; connect-src 'self' wss://ws.growww.in https://api.growww.in https://*.tradingview.com; script-src 'self' 'wasm-unsafe-eval'; worker-src 'self' blob:;
```

### 10.2 Memory Leak Prevention & Lifecycle Teardown

Trading workstations are frequently left open for days without refreshing. Memory leak prevention rules are strictly applied:
1. **ArrayBuffer Re-use:** Buffers allocated for WebSocket frame reception are recycled into an object pool rather than newly allocated per frame.
2. **Listener Disposals:** All `BroadcastChannel`, `window.addEventListener`, and `SharedWorker` port listeners implement deterministic disposal callbacks inside React 19 `useEffect` cleanups.
3. **Canvas Context Reclaim:** When closing popped-out windows or un-docking WebGL canvases, the rendering engine calls `gl.getExtension('WEBGL_lose_context')?.loseContext()` to release GPU VRAM.

### 10.3 Architecture Verification Matrix

| Verification Dimension | Criterion | Target Threshold | Validation Method |
| :--- | :--- | :--- | :--- |
| **Market Data Latency** | WebSocket ingress to WASM decode | $< 1.5\text{ ms}$ | Performance.now() telemetry in Web Worker |
| **UI Frame Stability** | Main thread frame rate under 50k ticks/sec | Locked $\ge 60\text{ FPS}$ | Chrome DevTools Performance CPU Profile |
| **Cumulative Layout Shift** | Order book updates and tick flashes | $\text{CLS} = 0.000$ | Lighthouse / Core Web Vitals Auditor |
| **Hotkey Dispatch Latency** | `Shift+B` press to WebSocket wire | $< 5.0\text{ ms}$ | Automated End-to-End Playwright test suite |
| **Cross-Window Sync** | Symbol change broadcast to pop-out render | $< 3.0\text{ ms}$ | BroadcastChannel high-resolution timestamp audit |
| **Zero-Fee Presentation** | Order forms, sheets, and blotters | $100\%$ display ₹0.00 | Snapshot assertion across all test fixtures |
| **Memory Retention** | 24-hour continuous streaming run | $\Delta\text{Memory} < 50\text{ MB}$ | Automated long-duration Puppeteer soak test |

---

## 11. Architectural Summary & Implementation Roadmap

The Web Pro Trading Terminal architecture establishes a modern browser workstation:
- Combines the rapid routing and server rendering of **Next.js 14** with the non-blocking concurrent rendering of **React 19**.
- Enables flexible workspace organization via **Dockview** and multi-monitor setups via **pop-out windows** synchronized over `BroadcastChannel`.
- Eliminates UI stalls during volatility surges by shifting network parsing to a **dedicated Web Worker** with a **WebAssembly Protobuf decoder** and a **SharedArrayBuffer** ring buffer.
- Achieves high visual performance and tabular precision via **WebGL 2.0 / WebGPU** and **TradingView Advanced Charts**.
- Delivers rapid authentication via **WebAuthn FIDO2** and **EIP-4361 SIWE**, paired with **ERC-4337 Session Keys** for hotkey trading without wallet prompts.
- Maintains the platform's core commitment to transparent **0.00% Zero-Fee** trading and **Zero-Gas Paymaster sponsorship**.
