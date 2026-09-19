# 622 - Web Pro Workstation Multi-Dock Grid Layout System

## Purpose
Deliver an institutional-grade, highly performant web interface specification for Web Pro Workstation Multi-Dock Grid Layout System on Next.js 14, providing ultra-low latency, fluid keyboard interactions, and responsive multi-monitor trading layouts.

This document establishes the authoritative technical blueprint, visual UI layout, interactive gestures, state management architecture, failure modes, and acceptance criteria for Web Pro Workstation Multi-Dock Grid Layout System across the Growww / NBSE trading platform.

## What You Are Building
Customizable grid layout manager (Dockview / GoldenLayout) allowing day traders to tile, stack, and float panels.

### Visual UI Wireframe & Web Layout Architecture
```
+-------------------------------------------------------------------------+
| GROWWW PRO TERMINAL | BTC/USDT $64,250.00 (+3.45%) | 24h Vol: 4,210.5 BTC       |
+-------------------------------------------------------------------------+
| [Chart] [Orderbook] [Recent Trades] [Depth] [Orders] [Settings]         |
+----------------------------------------------------+--------------------+
|                                                    | LEVEL 2 ORDERBOOK  |
|         TRADINGVIEW PRO ADVANCED CHART             | 64,252.00  1.45 BTC|
|                                                    | 64,251.00  0.82 BTC|
|                                                    | 64,250.00  Spread  |
|                                                    | 64,249.00  2.10 BTC|
|                                                    +--------------------+
|                                                    | ORDER ENTRY DRAWER |
|                                                    | [BUY]     [SELL]   |
|                                                    | Price: 64,250.00   |
|                                                    | Qty:   0.50 BTC    |
|                                                    | [25%][50%][75%][100]
|                                                    | [PLACE ORDER (Neon)|
+----------------------------------------------------+--------------------+
| OPEN ORDERS (2) | ORDER HISTORY | TRADE LOGS | POSITION VALUE: $32,125  |
+-------------------------------------------------------------------------+
```

Key UI capabilities include:
- **Deep Obsidian Palette**: Master UI theme rendered in Deep Obsidian (`#0B0E14`), Surface (`#121721`), and Card (`#181F2C`) with Neon Green (`#00F0A0`) and Neon Red (`#FF3B56`) accents.
- **Microsecond Tabular Layout**: Zero-layout-shift tabular typography using JetBrains Mono for prices, quantities, and balances.
- **Multi-Monitor Docking**: Detachable grid panels supporting independent monitor windows via BroadcastChannel synchronization.
- **Sub-Millisecond Hotkeys**: Comprehensive keyboard shortcut support for rapid, mouse-free trading operations.

## Scope Boundaries
- **In Scope:**
- Next.js 14 App Router layout component and client components with TypeScript.
- Zustand reactive state stores for real-time WebSocket market data and user state.
- Tailored Tailwind CSS classes implementing the Deep Obsidian design tokens.
- Interactive controls: sliders, steppers, modals, drawers, and detached window coordination.
- **Out of Scope / Handled Elsewhere:**
- Backend matching engine order execution (handled in Prompt 021).
- Hardware security module master key ceremonies (handled in Prompt 020).

## Technology to Use
- Core Technologies: Next.js 14, React 18, TypeScript, Tailwind CSS v3, Zustand, Dockview, Lucide Icons.
- Performance Standards: Sub-16ms frame render times, First Contentful Paint (FCP) < 1.0s, Cumulative Layout Shift (CLS) = 0.
- Design System: `packages/web_ui` shared component library with JetBrains Mono numbers.

## Backend / Infra Touchpoints
- Web Trading Terminal, API Gateway, Market Data WebSocket, User Session Service, Envoy Ingress.

## Blockchain Interaction
Displays real-time blockchain settlement status indicators verified against Hyperledger Besu on-chain state.

- **Consensus & Block Finality:** Hyperledger Besu QBFT consensus with 2-second block intervals and deterministic finality.
- **Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review UX design mockups and web interaction requirements for Web Pro Workstation Multi-Dock Grid Layout System.
2. Construct the responsive Next.js 14 App Router layout component and layout tree.
3. Establish state management stores using Zustand or Jotai for reactive data binding.
4. Implement zero-layout-shift tabular number rendering using JetBrains Mono typography.
5. Connect WebSocket subscriptions to live market tickers, depth diffs, and account updates.
6. Integrate keyboard navigation hotkeys, interactive sliders, and input steppers.
7. Add comprehensive client-side form validation, balance sufficiency, and fee calculations.
8. Establish graceful degradation, network reconnect banners, and toast notification queues.
9. Instrument OpenTelemetry web tracing and Core Web Vitals monitoring metrics.
10. Write Playwright end-to-end integration tests verifying cross-browser compatibility.
11. Verify sub-50ms user input response times and 60 FPS canvas rendering during heavy market load.
12. Review security and accessibility with the Lead Architect and obtain release sign-off.

## Interfaces / Contracts
```protobuf
syntax = "proto3";

package growww.web.ui.v1;

option go_package = "growww/packages/proto/growww/web/ui/v1;webuiv1";

message WebproworkstationmultidocklayoutsystemState {
  string user_id = 1;
  string active_market = 2;
  bool is_window_detached = 3;
  uint64 timestamp_ms = 4;
  map<string, string> component_layout = 5;
}

service WebproworkstationmultidocklayoutsystemService {
  rpc UpdateViewState(WebproworkstationmultidocklayoutsystemState) returns (WebproworkstationmultidocklayoutsystemState);
}
```

## Failure Modes & Edge Cases
| Failure Scenario | Trigger Condition | System Behavior & Mitigation |
| :--- | :--- | :--- |
| WebSocket Disconnection | Browser loses internet connectivity | Top bar displays subtle yellow reconnecting banner; buffers queued actions; auto-reconnects with exponential backoff. |
| Memory Leak on Long Sessions | Terminal open continuously for >24 hours | Circular buffers bound maximum in-memory tick history; old canvas frames garbage collected automatically. |
| Rapid Hotkey Double-Trigger | User presses Shift+B multiple times in 100ms | Request deduplicator drops secondary submissions; enforces 200ms debounce threshold with distinct client nonce. |
| Window Detach Crash | Secondary physical monitor disconnected | Detached window automatically re-docks into main browser container; restores original layout grid seamlessly. |
| Inadequate Trading Balance | Order entry value exceeds available funds | Immediate crimson outline on input field; disables submission button; displays instant deposit drawer link. |
| High Data Burst Lag | Orderbook receives 2,000 updates per second | UI throttles DOM re-renders to 20 FPS using requestAnimationFrame; keeps internal state strictly up-to-date. |
| Cross-Tab State Conflict | User places trade in Tab A while Tab B is open | BroadcastChannel synchronizes open orders and balance state across all open browser tabs within 10ms. |
| Stale Browser Cache | Outdated JavaScript bundle cached by browser | Service worker detects new release hash; displays non-intrusive 'New Version Available - Update' banner. |

## Acceptance Criteria
- [ ] Complete Next.js 14 UI layout, wireframe, and component architecture for Web Pro Workstation Multi-Dock Grid Layout System fully specified.
- [ ] Responsive design verified across desktop widescreen, multi-monitor displays, and laptop screens.
- [ ] Zustand state management stores and WebSocket data synchronization hooks defined.
- [ ] Zero layout shift (CLS = 0) verified during high-frequency market data streaming.
- [ ] 8 comprehensive failure modes and graceful degradation behaviors documented.
- [ ] Strict typography verified: exactly 0 Unicode em dashes or en dashes present in file.
- [ ] Zero implementation code files created on disk; specification strictly ready for build phase.
- [ ] Compliance with SEBI market disclosure and WCAG 2.1 accessibility standards validated.

## Regulatory & Compliance Mapping
- **Regulatory Framework:** SEBI electronic trading guidelines, WCAG 2.1 Level AA Accessibility Standards.
- **Audit & Governance:** 7-year WORM immutable audit trail retention, cryptographic non-repudiation, and automated compliance reporting.
