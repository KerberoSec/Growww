# Architecture & Pro Web Trading Terminal Specification

## Executive Overview
`growww_web` is an institutional-grade, browser-native pro trading workstation built on Next.js 14 (App Router), React 19 Concurrent Mode, TypeScript, and Tailwind CSS. It empowers active day traders, scalpers, and technical analysts with zero-install institutional execution directly inside any modern web browser.

Reference Master Blueprint: [`docs/architecture/UNIVERSAL_MULTI_PLATFORM_CLIENT_ARCHITECTURE_AND_THICK_CLIENT_BLUEPRINT.md`](../../docs/architecture/UNIVERSAL_MULTI_PLATFORM_CLIENT_ARCHITECTURE_AND_THICK_CLIENT_BLUEPRINT.md)

## Multi-Window Grid Docking & Detachment (Dockview)
- **Customizable Multi-Split Grid:** Powered by `dockview-core`, allowing traders to dock, split, tab, and float any panel (Chart, Order Book, Depth Ladder, Order Form, Execution Blotter, Watchlist).
- **Multi-Monitor Window Popping:** Clicking the pop-out button spawns independent native browser windows via `window.open()`, enabling dual-monitor or triple-monitor physical trading stations.
- **Cross-Window Synchronization:** Pop-out windows maintain sub-millisecond bidirectional state coherence via the `BroadcastChannel` API and a backing `SharedWorker`:
  - Link Group color tags (Group Red, Blue, Green, Yellow) synchronize active ticker selections across physical screens.
  - Active orders, fills, and cancellation alerts broadcast instantaneously to all detached windows without re-fetching from the server.

## Off-Thread Streaming & Memory Architecture (Web Workers + WASM)
- **Dedicated Web Worker:** The WebSocket stream (`wss://ws.growww.in/v1/market/depth/stream`) terminates inside a dedicated Web Worker thread.
- **WebAssembly Protobuf Decoding:** Ingests high-frequency binary Protobuf messages, decoding them in WebAssembly to avoid JavaScript main thread garbage collection cycles.
- **`SharedArrayBuffer` & Zero-Copy Depth State:** Order book levels update directly inside a fixed-size `SharedArrayBuffer` ring buffer.
- **50ms Conflation Dispatch:** The worker throttles DOM dispatches to a smooth 20 FPS (50ms interval), ensuring the React UI thread stays responsive at a locked 60/120 FPS even during 100,000 tick/second volatility bursts.

## Rendering Pipeline & Hardware Acceleration
- **WebGL 2.0 / WebGPU Canvas:** High-frequency visual Depth of Market (DOM) ladder and Level 2 depth visualizations render directly on HTML5 Canvas powered by WebGL hardware acceleration.
- **TradingView Advanced Charts:** Canvas-based candlestick charting supporting 100+ technical indicators, multi-timeframe overlays (1s, 1m, 5m, 1h, 1d), and drawing tools.
- **Zero Cumulative Layout Shift (CLS = 0):** Strict tabular number formatting with `JetBrains Mono` and reserved pixel-perfect container dimensions.

## Authentication, Web3 & Zero-Fee Invariants
- **WebAuthn / FIDO2 Passkeys:** Hardware biometric sign-in (TouchID, Windows Hello, FaceID) via WebAuthn API.
- **Non-Custodial Web3 Connectivity:** Native support for MetaMask, WalletConnect, Coinbase Wallet, and EIP-4361 (Sign-In with Ethereum - SIWE).
- **Zero-Fee Presentation Invariant:**
  - Order Entry and Confirmation sheets strictly display `Trading Fee: ₹0.00 (0.00% Genesis Zero-Fee)`.
  - On-chain settlement confirms `Paymaster Sponsored: ₹0.00 Gas Fee`.
  - Zero on-chain TDS (`₹0.00`) during settlement.

## Keyboard Hotkey Engine
- `Shift + B`: Instant Market Buy order at best offer.
- `Shift + S`: Instant Market Sell order at best bid.
- `Escape`: Panic Cancel all active orders for current trading pair.
- `Shift + Escape`: Global Panic Cancel across all pairs.
- `Space`: Open Quick-Search Ticker Palette.
- `1` / `2` / `5` / `0`: Pre-set margin lot sizing (10%, 25%, 50%, 100%).

## Responsive Viewport Hierarchy
1. **Multi-Monitor / Ultra-Wide (>1920px):** Full multi-pane docking grid with detachable pop-outs.
2. **Standard Desktop (1280px - 1920px):** 3-column consolidated workspace.
3. **Laptop / Compact (1024px - 1280px):** 2-column tabbed workspace.
4. **Tablet & Mobile Web (<1024px):** Adaptive single-column view with bottom sheets matching native mobile ergonomics.
5. **Progressive Web App (PWA):** Installable to desktop dock / taskbar with offline service worker shell caching.
