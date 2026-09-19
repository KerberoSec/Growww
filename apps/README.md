# Client Applications Suite (Web, Desktop Thick Clients & Mobile)

## Executive Overview
The `apps/` directory contains all client-facing trading workstations, institutional desktop applications, mobile apps, and administrative frontends for the Growww / NBSE trading platform. Every application is built with an Obsidian Dark Mode default (`#0B0E14`), high-refresh tabular typography, zero-fee transparency, and fluid hardware-accelerated charting.

Architectural Master Blueprint: [`docs/architecture/UNIVERSAL_MULTI_PLATFORM_CLIENT_ARCHITECTURE_AND_THICK_CLIENT_BLUEPRINT.md`](../docs/architecture/UNIVERSAL_MULTI_PLATFORM_CLIENT_ARCHITECTURE_AND_THICK_CLIENT_BLUEPRINT.md)

## Multi-Platform Target Inventory

| Application | Target Platforms | Architectural Technology | Primary Use Case & Capabilities |
| :--- | :--- | :--- | :--- |
| **`growww_web`** | **Web (Desktop Pro, Tablet, Mobile Web)** | Next.js 14 (App Router) / React 19 / TS / WebGL / WebAssembly / Tailwind | **Main Focus:** Browser-native institutional trading terminal with Dockview multi-window layout, multi-monitor detachable pop-outs (`BroadcastChannel`), off-thread Web Worker Protobuf streaming, TradingView charting, and WebAuthn Passkeys. |
| **`growww_flutter`** (Desktop) | **macOS (Apple Silicon & Intel), Windows (x64/ARM64), Linux (X11/Wayland)** | Flutter 3.22+ Desktop / Rust FFI (`rust_trading_core`) / Impeller / Metal / DirectX 12 | **Main Focus (Thick Client):** Multi-monitor native desktop workstation with independent OS window handles, DOM click-to-trade depth ladder, Apple Secure Enclave / TouchID, Windows Hello biometrics, global OS keyboard hotkeys, and direct ring-buffer tick streaming. |
| **`growww_flutter`** (Mobile) | **iOS (iPhone/iPad), Android (Phone/Tablet/Foldable)** | Flutter 3.22+ / Riverpod 2.5 / Metal & Vulkan / SQLite WAL | Sovereign mobile trading with FaceID / BiometricPrompt, tactile haptics, battery-optimized streaming lifecycle, and encrypted offline order queue. |
| **`growww_admin`** | **Web (Restricted Admin Back-Office)** | Next.js 14 App Router / Edge Middleware / WebAuthn FIDO2 | Back-office compliance operations, KYC reviews, P2P escrow arbitration, and Proof-of-Reserve audits with cryptographic four-eyes authorization. |
| **`growww_marketing`** | **Web (Public Marketing & Developer Portal)** | Next.js 14 Static Generation / SEO / MDX Docs | Public landing pages, developer documentation, API reference, testnet faucet UI, and statutory regulatory risk disclosures. |

## Platform-Specific Native Capabilities

### 1. Web Pro Trading Terminal (`growww_web` - Main Focus)
- **Multi-Monitor Window Popping:** Detach charts, order books, and watchlists into separate browser windows across multiple displays with zero-delay synchronization via `BroadcastChannel` and `SharedWorker`.
- **Worker-Driven Streaming:** WebSocket streams decode inside dedicated Web Workers using WebAssembly Protobuf decoders, feeding the DOM via `SharedArrayBuffer` at locked 60/120 FPS.
- **Fast Trading Hotkeys:** Rapid order entry via `Shift+B` (Buy Market), `Shift+S` (Sell Market), and `Escape` (Panic Cancel All).

### 2. macOS Thick Client (`growww_flutter` - Main Focus)
- **Metal Acceleration:** Render 50-depth order books and charts at 120 FPS on Apple ProMotion displays.
- **Apple Secure Enclave:** Hardware-isolated credential storage and TouchID biometric trade confirmation.
- **Cocoa Menu Bar & Dock:** Global macOS menu integration and Dock icon live PnL badge.

### 3. Windows Thick Client (`growww_flutter` - Main Focus)
- **DirectX 12 Hardware Rendering:** Butter-smooth 144Hz - 360Hz refresh rate support for high-end monitors.
- **Windows Hello Biometrics:** Instant facial recognition and fingerprint validation for order signing.
- **Global Hotkeys:** Win32 `RegisterHotKey` API allows panic-canceling active orders even when unfocused.
- **Per-Monitor DPI v2:** Flawless multi-monitor dragging across mixed DPI displays (e.g. 4K 150% + 1440p 100%).

### 4. Android & iOS Mobile Client (`growww_flutter`)
- **Ergonomic Trading UI:** Single-thumb order placement, swipeable order forms, and responsive bottom sheets.
- **Tactile Haptic Pulses:** Native selection, confirmation impact, and volatility alert vibrations.
- **Offline Resilient Queue:** Encrypted SQLite WAL mode queues orders during cellular tunnel dropouts.

## Local Execution & Development
```bash
# Web Pro Terminal (Next.js 14)
cd apps/growww_web && pnpm install && pnpm dev # http://localhost:3000

# Universal Flutter Client (Desktop macOS / Windows / Linux & Mobile)
cd apps/growww_flutter && flutter pub get
flutter run -d macos     # macOS Native Desktop Thick Client
flutter run -d windows   # Windows Native Desktop Thick Client
flutter run -d linux     # Linux Native Desktop Thick Client
flutter run -d chrome    # Web Fallback Target
flutter run -d ios       # iOS Simulator or Physical Device
flutter run -d android   # Android Emulator or Physical Device
```
