# 547 - Flutter Market Order Confirmation Modal with Slippage Protection

## Purpose
Protect mobile traders from unexpected fill prices in volatile markets by displaying an instantaneous execution confirmation modal with user-configurable slippage tolerance caps.

This document establishes the authoritative technical blueprint, visual UI layout, interactive gestures, state management architecture, failure modes, and acceptance criteria for Flutter Market Order Confirmation Modal with Slippage Protection across the Growww / NBSE trading platform.

## What You Are Building
A high-clarity confirmation dialog displaying estimated fill price, weighted slippage impact across orderbook depth, and customizable slippage caps.

### Visual UI Wireframe & Layout Architecture
```
+-------------------------------------------------------+
| CONFIRM MARKET BUY (BTC)                             |
+-------------------------------------------------------+
| Best Available Ask:       $64,250.00                  |
| Estimated Fill Price:     $64,262.50 (+0.02%)         |
| Order Size:               0.50000000 BTC              |
| Total Estimated Cost:     $32,131.25 USDT             |
+-------------------------------------------------------+
| Slippage Tolerance:                                   |
| [ 0.1% ]     [ (0.5% Default) ]     [ 1.0% ]   [Custom]|
+-------------------------------------------------------+
| [!] Warning: Price may vary if market moves rapidly. |
+-------------------------------------------------------+
| [ Cancel ]             | [ CONFIRM EXECUTION (Green) ]|
+-------------------------------------------------------+
```

Key UI capabilities include:
- **Obsidian Dark Mode**: Engineered on the Deep Obsidian (`#0B0E14`) theme with Neon Green (`#00F0A0`) and Neon Red (`#FF3B56`) accents.
- **Micro-Animations & Haptics**: Subtle tactile haptic triggers on user taps, fills, and mode toggles.
- **60/120 FPS Performance**: Hardware-accelerated rendering with zero layout shift during continuous WebSocket data bursts.
- **Strict Accessibility**: Contrast ratios exceeding WCAG 2.1 Level AA standards on dark surfaces.

## Scope Boundaries
- **In Scope:**
- Real-time calculation of expected slippage based on current L2 orderbook depth.
- Selectable slippage tolerance pills: 0.1%, 0.5% (recommended default), 1.0%, or custom input.
- Visual alert highlight if market depth is insufficient, warning of partial fills or high slippage.
- Sub-300ms transition to trade receipt upon execution.
- **Out of Scope / Handled Elsewhere:**
- Database backup automation.
- Hardware key generation.

## Technology to Use
- Core Technologies: Flutter AlertDialog, Riverpod, BigInt math, OpenTelemetry event tracer.
- Performance Standards: Sub-16ms frame render times (60 FPS minimum, 120 FPS target), zero-allocation layout pipelines.
- Design System: `packages/growww_ui` shared design library with JetBrains Mono numbers.

## Backend / Infra Touchpoints
- MarketOrderModal, DepthService, SlippageCalculator, OrderGatewayClient.

## Blockchain Interaction
Submits signed execution payload to Hyperledger Besu DvP smart contract with maxSlippagePrice parameter.

- **Consensus & Block Finality:** Hyperledger Besu QBFT consensus with 2-second block intervals and deterministic finality.
- **Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review UX design mockups and interaction requirements for Flutter Market Order Confirmation Modal with Slippage Protection.
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

package growww.ui.marketorder.v1;

option go_package = "growww/packages/proto/growww/ui/marketorder/v1;marketorderv1";

message PromptMarketConfirmationViewState {
  string user_id = 1;
  string trading_pair = 2;
  bool is_active = 3;
  uint64 last_updated_ms = 4;
  map<string, string> visual_tokens = 5;
}

service MarketOrderConfirmationService {
  rpc PromptMarketConfirmation(PromptMarketConfirmationViewState) returns (PromptMarketConfirmationViewState);
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
- [ ] Complete UI layout, ASCII wireframe, and interaction flows for Flutter Market Order Confirmation Modal with Slippage Protection fully specified.
- [ ] Responsive design verified for mobile portrait and tablet/desktop landscape orientations.
- [ ] State management integration defined with Riverpod providers and reactive viewmodels.
- [ ] 60/120 FPS performance budget verified with zero layout shift during continuous price updates.
- [ ] 8 comprehensive UI failure modes and graceful recovery behaviors documented.
- [ ] Strict typography verified: exactly 0 Unicode em dashes or en dashes present in file.
- [ ] Zero implementation code files created on disk; specification strictly ready for build phase.
- [ ] Compliance with SEBI market disclosure and WCAG accessibility standards validated.

## Regulatory & Compliance Mapping
- **Regulatory Framework:** SEBI Best Execution Guidelines for Retail Investors, Consumer Financial Protection Standards.
- **Audit & Governance:** 7-year WORM immutable audit trail retention, cryptographic non-repudiation, and automated compliance reporting.
