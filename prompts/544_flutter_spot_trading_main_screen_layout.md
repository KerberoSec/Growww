# 544 - Flutter Spot Trading Main Screen Layout & Tab Navigation

## Purpose
Deliver the core mobile trading workstation interface for BTC/USDT spot trading, seamlessly integrating real-time ticker headers, fluid TradingView charts, Level 2 orderbook ladders, recent trades tape, and bottom order entry sheets.

This document establishes the authoritative technical blueprint, visual UI layout, interactive gestures, state management architecture, failure modes, and acceptance criteria for Flutter Spot Trading Main Screen Layout & Tab Navigation across the Growww / NBSE trading platform.

## What You Are Building
Scaffolding the main responsive spot trade screen in Flutter with sliver app bars, tab controllers, Riverpod state injection, and sticky bottom action drawers.

### Visual UI Wireframe & Layout Architecture
```
+-------------------------------------------------------+
| [<-] BTC/USDT  $64,250.00  +3.45% (24h)        [Search] |
+-------------------------------------------------------+
| 24h High: 65,100  | 24h Low: 62,800  | 24h Vol: 4.2k BTC|
+-------------------------------------------------------+
| [ 1m | 5m | 15m | 1h | 4h | 1D ]   [Indicators] [Depth] |
|                                                       |
|              TRADINGVIEW CANDLESTICK CHART            |
|                                                       |
+-------------------------------------------------------+
| [Orderbook]  [Recent Trades]  [Market Info]            |
| 64,251.50   0.452 BTC  |  Ask Wall (Red)              |
| 64,250.00   0.120 BTC  |  Spread: $0.50 (0.001%)      |
| 64,249.50   1.820 BTC  |  Bid Wall (Green)            |
+-------------------------------------------------------+
| [ BUY BTC (Neon Green) ]     | [ SELL BTC (Neon Red) ] |
+-------------------------------------------------------+
```

Key UI capabilities include:
- **Obsidian Dark Mode**: Engineered on the Deep Obsidian (`#0B0E14`) theme with Neon Green (`#00F0A0`) and Neon Red (`#FF3B56`) accents.
- **Micro-Animations & Haptics**: Subtle tactile haptic triggers on user taps, fills, and mode toggles.
- **60/120 FPS Performance**: Hardware-accelerated rendering with zero layout shift during continuous WebSocket data bursts.
- **Strict Accessibility**: Contrast ratios exceeding WCAG 2.1 Level AA standards on dark surfaces.

## Scope Boundaries
- **In Scope:**
- Responsive layout adapting across mobile (portrait) and tablet/desktop (landscape split-view).
- Persistent top header displaying live price ticker, 24h high/low, and 24h volume.
- Smooth tab navigation between Orderbook, Market Trades, and Asset Profile without reloading chart state.
- Sticky bottom action bar triggering Buy/Sell order entry sheets.
- **Out of Scope / Handled Elsewhere:**
- Backend order matching logic.
- Blockchain validator consensus voting.

## Technology to Use
- Core Technologies: Flutter 3.19+, Riverpod 2.5, CustomScrollView, SliverAppBar, JetBrains Mono tabular font.
- Performance Standards: Sub-16ms frame render times (60 FPS minimum, 120 FPS target), zero-allocation layout pipelines.
- Design System: `packages/growww_ui` shared design library with JetBrains Mono numbers.

## Backend / Infra Touchpoints
- SpotTradeScreen, MarketDataController, OrderbookWidget, TradingViewCanvas, BottomOrderBar.

## Blockchain Interaction
Displays on-chain settlement confirmation badges for completed trades executed on Hyperledger Besu.

- **Consensus & Block Finality:** Hyperledger Besu QBFT consensus with 2-second block intervals and deterministic finality.
- **Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review UX design mockups and interaction requirements for Flutter Spot Trading Main Screen Layout & Tab Navigation.
2. Define the responsive Flutter widget hierarchy, layout constraints, and orientation adaptations.
3. Construct the state management controllers and Riverpod providers governing view state.
4. Implement custom painter and zero-layout-shift typography using tabular monospace figures.
5. Bind real-time WebSocket streams, market data updates, and balance event listeners.
6. Integrate tactile haptic feedback triggers and sub-200ms fluid micro-animations.
7. Add input validation, balance sufficiency checks, and biometric security prompts.
8. Establish graceful degradation, offline state caching, and error toast alerts.
9. Connect telemetry, OpenTelemetry UI interaction spans, and user engagement analytics.
10. Write comprehensive unit and golden file widget tests verifying UI visual consistency.
11. Profile frame rendering times ensuring strict 60 FPS on mid-range and 120 FPS on high-refresh displays.
12. Review accessibility with the Design Lead and obtain statutory compliance sign-off.

## Interfaces / Contracts
```protobuf
syntax = "proto3";

package growww.ui.trade.v1;

option go_package = "growww/packages/proto/growww/ui/trade/v1;tradev1";

message SpotTradeLayoutViewState {
  string user_id = 1;
  string trading_pair = 2;
  bool is_active = 3;
  uint64 last_updated_ms = 4;
  map<string, string> visual_tokens = 5;
}

service SpotTradeViewService {
  rpc GetSpotTradeLayout(SpotTradeLayoutViewState) returns (SpotTradeLayoutViewState);
}
```

## Failure Modes & Edge Cases
| Failure Scenario | Trigger Condition | System Behavior & Mitigation |
| :--- | :--- | :--- |
| Network Disconnection | Device loses cellular/Wi-Fi connection | Displays non-blocking amber status bar; caches pending inputs locally; auto-reconnects with exponential backoff. |
| Frame Drop / UI Stutter | Burst of 1,000 WebSocket ticks per second | Stream conflation buffers updates to 50ms intervals; renders latest snapshot via RepaintBoundary. |
| Accidental Double-Tap | User rapidly taps trade action button twice | Button debounce lock (300ms) disables secondary tap; generates unique idempotent client request token. |
| Inadequate Balance | Order entry value exceeds available balance | Immediate red highlight on balance text; disables swipe slider; displays inline deposit shortcut button. |
| Stale Market Data | WebSocket connection stalls without TCP drop | Heartbeat watchdog detects missing ping after 5 seconds; marks quotes as stale and triggers silent socket reconnect. |
| Device Theme Mismatch | OS switches to system Light Mode | Enforces platform-wide invariant: Dark Mode Obsidian is strictly locked; displays explanation modal if requested. |
| Biometric Sensor Failure | TouchID/FaceID sensor fails or is locked out | Gracefully falls back to platform PIN / password entry without interrupting active order entry state. |
| Screen Rotation Drift | Device rotated between portrait and landscape | State preserved via Riverpod persistent controllers; smoothly re-flows layout into landscape pro workstation. |

## Acceptance Criteria
- [ ] Complete UI layout, ASCII wireframe, and interaction flows for Flutter Spot Trading Main Screen Layout & Tab Navigation fully specified.
- [ ] Responsive design verified for mobile portrait and tablet/desktop landscape orientations.
- [ ] State management integration defined with Riverpod providers and reactive viewmodels.
- [ ] 60/120 FPS performance budget verified with zero layout shift during continuous price updates.
- [ ] 8 comprehensive UI failure modes and graceful recovery behaviors documented.
- [ ] Strict typography verified: exactly 0 Unicode em dashes or en dashes present in file.
- [ ] Zero implementation code files created on disk; specification strictly ready for build phase.
- [ ] Compliance with SEBI market disclosure and WCAG accessibility standards validated.

## Regulatory & Compliance Mapping
- **Regulatory Framework:** SEBI Fair Disclosure Norms for Market Quotes, WCAG 2.1 Dark Mode Contrast Standards.
- **Audit & Governance:** 7-year WORM immutable audit trail retention, cryptographic non-repudiation, and automated compliance reporting.
