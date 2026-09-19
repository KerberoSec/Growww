# 545 - Flutter Interactive Orderbook Depth Ladder & Bid/Ask Visualizer

## Purpose
Provide crypto spot traders with an animated, zero-layout-shift Level 2 orderbook ladder displaying bids, asks, spreads, and relative depth volume heatmaps in real-time.

This document establishes the authoritative technical blueprint, visual UI layout, interactive gestures, state management architecture, failure modes, and acceptance criteria for Flutter Interactive Orderbook Depth Ladder & Bid/Ask Visualizer across the Growww / NBSE trading platform.

## What You Are Building
CustomPainter-backed high-performance orderbook ladder rendering horizontal volume bars, animated spread indicators, and interactive tap-to-fill price selection.

### Visual UI Wireframe & Layout Architecture
```
+-------------------------------------------------------+
| LEVEL 2 ORDERBOOK (BTC/USDT)                          |
| Price (USDT)         Size (BTC)          Total (BTC)  |
+-------------------------------------------------------+
| 64,255.00 (Red)      0.852               4.821  [||||]|
| 64,254.00 (Red)      1.240               3.969  [|||] |
| 64,252.50 (Red)      0.510               2.729  [||]  |
| 64,251.00 (Red)      2.219               2.219  [||||]|
+-------------------------------------------------------+
| Spread: $1.00 (0.0015%)   USD: $64,250.00             |
+-------------------------------------------------------+
| 64,250.00 (Green)    1.150               1.150  [|||] |
| 64,249.00 (Green)    3.420               4.570  [||||]|
| 64,247.50 (Green)    0.910               5.480  [||]  |
| 64,246.00 (Green)    2.100               7.580  [||||]|
+-------------------------------------------------------+
| [Precision: 0.1 | 0.01 | 1.0]      [Depth: 10 | 20 | 50] |
+-------------------------------------------------------+
```

Key UI capabilities include:
- **Obsidian Dark Mode**: Engineered on the Deep Obsidian (`#0B0E14`) theme with Neon Green (`#00F0A0`) and Neon Red (`#FF3B56`) accents.
- **Micro-Animations & Haptics**: Subtle tactile haptic triggers on user taps, fills, and mode toggles.
- **60/120 FPS Performance**: Hardware-accelerated rendering with zero layout shift during continuous WebSocket data bursts.
- **Strict Accessibility**: Contrast ratios exceeding WCAG 2.1 Level AA standards on dark surfaces.

## Scope Boundaries
- **In Scope:**
- Top 20 bids and asks rendered at 20 FPS with zero layout shift using tabular figures.
- Interactive price tap: tapping any ask or bid row automatically populates the order entry price field.
- Dynamic depth aggregation selector (0.1, 1.0, 10.0 USDT tick groupings).
- Visual volume heatbars showing relative cumulative liquidity clusters.
- **Out of Scope / Handled Elsewhere:**
- Matching engine matching algorithm.
- Database persistence.

## Technology to Use
- Core Technologies: Flutter CustomPainter, RepaintBoundary, StreamBuilder, WebSocket depth stream.
- Performance Standards: Sub-16ms frame render times (60 FPS minimum, 120 FPS target), zero-allocation layout pipelines.
- Design System: `packages/growww_ui` shared design library with JetBrains Mono numbers.

## Backend / Infra Touchpoints
- OrderbookWidget, DepthPainter, OrderEntryController, WebSocketClient.

## Blockchain Interaction
Visually marks resting limit orders registered in Hyperledger Besu DvP smart contract registry.

- **Consensus & Block Finality:** Hyperledger Besu QBFT consensus with 2-second block intervals and deterministic finality.
- **Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review UX design mockups and interaction requirements for Flutter Interactive Orderbook Depth Ladder & Bid/Ask Visualizer.
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

package growww.ui.orderbook.v1;

option go_package = "growww/packages/proto/growww/ui/orderbook/v1;orderbookv1";

message OrderbookLadderViewState {
  string user_id = 1;
  string trading_pair = 2;
  bool is_active = 3;
  uint64 last_updated_ms = 4;
  map<string, string> visual_tokens = 5;
}

service OrderbookLadderService {
  rpc StreamOrderbookLadder(OrderbookLadderViewState) returns (OrderbookLadderViewState);
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
- [ ] Complete UI layout, ASCII wireframe, and interaction flows for Flutter Interactive Orderbook Depth Ladder & Bid/Ask Visualizer fully specified.
- [ ] Responsive design verified for mobile portrait and tablet/desktop landscape orientations.
- [ ] State management integration defined with Riverpod providers and reactive viewmodels.
- [ ] 60/120 FPS performance budget verified with zero layout shift during continuous price updates.
- [ ] 8 comprehensive UI failure modes and graceful recovery behaviors documented.
- [ ] Strict typography verified: exactly 0 Unicode em dashes or en dashes present in file.
- [ ] Zero implementation code files created on disk; specification strictly ready for build phase.
- [ ] Compliance with SEBI market disclosure and WCAG accessibility standards validated.

## Regulatory & Compliance Mapping
- **Regulatory Framework:** IOSCO Market Transparency Standards, SEBI Orderbook Fairness Guidelines.
- **Audit & Governance:** 7-year WORM immutable audit trail retention, cryptographic non-repudiation, and automated compliance reporting.
