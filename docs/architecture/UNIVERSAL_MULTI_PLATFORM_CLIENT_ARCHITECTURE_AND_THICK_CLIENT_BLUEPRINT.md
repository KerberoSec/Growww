# Universal Multi-Platform Client Architecture & Thick Client Master Blueprint

**Specification ID:** SPEC-ARCH-042-CLIENT-UNIVERSAL  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Client Applications, Web Terminal & Desktop Thick Client Architecture  
**Target Environments:** Web (Desktop Pro), macOS (Apple Silicon & Intel), Windows (x64/ARM64), Linux (X11/Wayland), Android, iOS  
**Last Updated:** September 2026  

---

## 1. Executive Summary & Core Platform Directives

The Growww / NBSE trading infrastructure delivers a unified, sovereign trading experience across every consumer and institutional computing platform. Rather than treating desktop, web, and mobile as disparate silos, the architecture establishes a shared financial core with specialized, platform-native rendering shells.

### Primary Focus Directives:
1. **Web Trading Terminal (Main Focus - Retail & Day Traders):**
   - High-performance, browser-native Pro Trading Workstation built with Next.js 14 (App Router), React 19 Concurrent Mode, TypeScript, Tailwind CSS, and WebGL/WebGPU.
   - Modular multi-window docking layout via Dockview/FlexLayout with multi-monitor pop-out windows communicating over the `BroadcastChannel` API and `SharedWorker`.
   - Dedicated Web Workers utilizing `SharedArrayBuffer` for off-main-thread binary Protobuf WebSocket stream decoding, ensuring locked 60/120 FPS rendering.
   - WebAuthn/FIDO2 Passkeys, Web3 non-custodial wallet connectivity (EIP-4361 SIWE, WalletConnect), and zero Cumulative Layout Shift (CLS = 0).
2. **macOS Native Thick Client (Main Focus - Institutional & Pro Mac):**
   - Native thick client built with Flutter 3.22+ Desktop and AppKit/Cocoa native extensions, optimized for Apple Silicon (M1/M2/M3/M4 Pro/Max/Ultra) with universal x86_64 support.
   - Hardware-accelerated Metal rendering pipeline via the Flutter Impeller engine, locking 120Hz ProMotion display refresh rates.
   - Deep macOS integration: Apple Secure Enclave & TouchID/FaceID via `LocalAuthentication`, Native Menu Bar shortcuts, Dock icon live PnL badging, and multi-display detachment across Apple Pro Display XDRs.
3. **Windows Native Thick Client (Main Focus - Active Scalpers & Prop Desks):**
   - Native thick client built with Flutter 3.22+ Desktop and Win32/WinUI 3 native extensions, compiled for Windows 10/11 x64 and ARM64.
   - DirectX 12 / Direct3D hardware-accelerated rendering capable of driving high-refresh gaming and institutional trading monitors (144Hz, 240Hz, 360Hz).
   - Deep Windows integration: Windows Hello biometrics (fingerprint & facial recognition), global OS-level keyboard hooks (`RegisterHotKey`) for panic cancel-all even when unfocused, per-monitor DPI v2 scaling, and System Tray quick-ticker monitoring.
4. **Linux Native Thick Client (Everywhere - Developers & Algo Desks):**
   - Native GTK 3/4 & Wayland/X11 multi-head window detachment with Vulkan hardware acceleration, packaged via Flatpak, AppImage, and Snap.
   - System Secret Service API (`libsecret`) integration for encrypted credential storage.
5. **Mobile Applications (Android & iOS - Retail Mobility):**
   - Universal Flutter mobile client (`apps/growww_flutter`) with Obsidian Dark Theme (`#0B0E14`), neon accents, tactile haptics, biometric quick-auth (BiometricPrompt / FaceID), and offline order queuing via encrypted SQLite WAL mode.
6. **Cross-Platform Synchronization ("Everywhere"):**
   - Cloud-synced workspace geometry, color-coded ticker Link Groups (Red, Blue, Green, Yellow), universal sub-paise financial precision (`decimal`), and the immutable **Zero-Fee Invariant (0.00% Maker / 0.00% Taker / 0 Gas / 0 TDS)** displayed across all interfaces.

---

## 2. Multi-Platform Architectural Matrix

| Dimension | Web Pro Terminal (`growww_web`) | macOS Thick Client (`growww_flutter`) | Windows Thick Client (`growww_flutter`) | Linux Thick Client (`growww_flutter`) | iOS Mobile Client (`growww_flutter`) | Android Mobile Client (`growww_flutter`) |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Primary Framework** | Next.js 14 / React 19 / TS | Flutter 3.22+ / Dart 3.4+ / Rust | Flutter 3.22+ / Dart 3.4+ / Rust | Flutter 3.22+ / Dart 3.4+ / Rust | Flutter 3.22+ / Dart 3.4+ | Flutter 3.22+ / Dart 3.4+ |
| **Rendering Backend** | WebGL 2.0 / WebGPU / Canvas | Metal (via Impeller Engine) | DirectX 12 / Direct3D (Impeller) | Vulkan / OpenGL (Impeller) | Metal (Impeller) | Vulkan (Impeller) |
| **Target Refresh Rate** | 60 - 120 FPS | 120 FPS (Apple ProMotion) | 144 - 360 FPS (G-Sync/FreeSync) | 60 - 144 FPS | 60 - 120 FPS (ProMotion) | 60 - 120 FPS |
| **Multi-Window System** | Dockview + Detachable Pop-outs | Native Cocoa Child Windows | Native Win32 Child Windows | Native GTK / Wayland Windows | Single Viewport / Split-Screen | Single Viewport / Foldable |
| **Inter-Window IPC** | `BroadcastChannel` + `SharedWorker` | Named Pipes / Dart `Isolate` | Win32 Named Pipes / Dart `Isolate` | Unix Domain Sockets / `Isolate` | N/A | N/A |
| **Market Data Socket** | Web Worker + Protobuf WASM | Rust FFI (`rust_trading_core`) | Rust FFI (`rust_trading_core`) | Rust FFI (`rust_trading_core`) | Dart WebSocket + SIMD isolate | Dart WebSocket + SIMD isolate |
| **Hardware Biometrics** | WebAuthn / FIDO2 Passkeys | Apple Secure Enclave / TouchID | Windows Hello (Biometrics/PIN) | Secret Service API / FIDO2 | Apple FaceID / TouchID | Android BiometricPrompt |
| **Global OS Hotkeys** | Browser Keyboard Hooks | Carbon / Cocoa Event Monitor | Win32 `RegisterHotKey` API | X11 / Wayland Global Shortcuts | Gesture Quick-Swipes | Gesture Quick-Swipes |
| **Offline Persistence** | IndexedDB (idb) + localStorage | Encrypted SQLite (WAL mode) | Encrypted SQLite (WAL mode) | Encrypted SQLite (WAL mode) | Encrypted SQLite (WAL mode) | Encrypted SQLite (WAL mode) |
| **Binary Packaging** | PWA / Edge CDN deployment | `.dmg` / Universal PKG (Notarized) | MSIX / WiX Installer (Signed) | Flatpak / AppImage / `.deb` | Apple App Store / TestFlight | Google Play Store / APK |

---

## 3. Web Pro Trading Terminal Architecture (Main Focus)

The Pro Web Trading Terminal (`apps/growww_web`) serves active day traders, technical analysts, and retail investors requiring immediate, zero-install institutional capabilities in any modern Chromium, WebKit, or Gecko browser.

```
+-------------------------------------------------------------------------------------------------------+
|                                    PRO WEB TRADING TERMINAL ARCHITECTURE                              |
|                                                                                                       |
|   +-----------------------------------------------------------------------------------------------+   |
|   |                         MAIN BROWSER WINDOW (Next.js 14 / React 19)                           |   |
|   |  +------------------------+  +-------------------------------+  +--------------------------+  |   |
|   |  | Multi-Ticker Watchlist |  | TradingView Advanced Charting |  | Level 2 / 3 Order Book   |  |   |
|   |  | - Sparklines & Volume  |  | - 100+ Indicators, PineScript |  | - 50-Depth Visual Ladder |  |   |
|   |  +------------------------+  +-------------------------------+  +--------------------------+  |   |
|   |  +------------------------+  +-------------------------------+  +--------------------------+  |   |
|   |  | DOM Click-to-Trade     |  | Instant Order Entry Form      |  | Real-Time Trade Blotter  |  |   |
|   |  | - Single-Click Scaling |  | - 0.00% Zero-Fee Calculator   |  | - Besu Tx Hash Telemetry |  |   |
|   |  +------------------------+  +-------------------------------+  +--------------------------+  |   |
|   +-----------------------------------------------------------------------------------------------+   |
|                                   |                                    |                              |
|                    BroadcastChannel API / SharedWorker                 |                              |
|                   +---------------+---------------+                    |                              |
|                   |                               |                    |                              |
|                   v                               v                    v                              |
|   +-------------------------------+  +-------------------------------+ |                              |
|   | POP-OUT WINDOW 1 (Monitor 2)  |  | POP-OUT WINDOW 2 (Monitor 3)  | |                              |
|   | Detached 4K Full-Screen Chart |  | Detached DOM Depth Ladder     | |                              |
|   +-------------------------------+  +-------------------------------+ |                              |
|                                                                        |                              |
|                  +-----------------------------------------------------+                              |
|                  |                                                                                    |
|                  v                                                                                    |
|   +-----------------------------------------------------------------------------------------------+   |
|   |                               DEDICATED WEB WORKER ENGINE (WASM)                              |   |
|   |  - High-Throughput WSS Client (`wss://ws.growww.in/v1/market/depth/stream`)                   |   |
|   |  - Protobuf Binary Streaming Decoder (WASM Rust/C compiled)                                   |   |
|   |  - Zero-Copy Level 2 Depth Ring Buffer in `SharedArrayBuffer`                                 |   |
|   |  - 50ms Conflation Throttler: Emits at max 20 DOM frame dispatches/sec to prevent UI locking  |   |
|   +-----------------------------------------------------------------------------------------------+   |
+-------------------------------------------------------------------------------------------------------+
```

### 3.1 Docking Layout System (Dockview / FlexLayout)
The terminal interface is structured on a dynamic, split-pane grid powered by `dockview-core`:
- **Panels:** Watchlist, Chart, Order Book, Depth Ladder, Order Form, Positions Blotter, Trade History, System Status.
- **Detachable Pop-Outs:** Clicking the "Pop Out" icon spawns a secondary native browser window via `window.open('/trade/popout?panel=chart&symbol=BTC-USDT')`.
- **Cross-Window Synchronization:**
  ```typescript
  // Synchronizing symbol selection across popped-out browser windows
  const channel = new BroadcastChannel('growww_workspace_sync');
  channel.postMessage({
    type: 'SYMBOL_CHANGE',
    linkGroup: 'group_red',
    symbol: 'BTC-USDT',
    timestampNs: Date.now() * 1000000
  });

  channel.onmessage = (event) => {
    if (event.data.type === 'SYMBOL_CHANGE' && event.data.linkGroup === currentLinkGroup) {
      setActiveSymbol(event.data.symbol);
    }
  };
  ```

### 3.2 Web Worker & SharedArrayBuffer Ingestion Pipeline
To eliminate UI thread stutter during violent market volatility bursts (e.g. 100,000 depth ticks/second):
1. The WebSocket connection connects directly inside a dedicated `WebWorker`.
2. Binary Protobuf payloads are decoded in WebAssembly (compiled from Rust `prost` or Go).
3. Order book state is maintained in a fixed-size `SharedArrayBuffer` memory segment shared between worker and main thread.
4. An `Atomics`-based state flag notifies the UI thread that a fresh 50ms conflated frame is ready to render via WebGL canvas.

### 3.3 Zero-Latency Keyboard Hotkey Execution
Professional day traders trade without touching the mouse:
- `Shift + B`: Instant Market Buy Order Entry.
- `Shift + S`: Instant Market Sell Order Entry.
- `Space`: Focus Ticker Search Command Palette.
- `Escape`: Panic Cancel All Active Orders for Current Symbol.
- `Shift + Escape`: Global Panic Cancel All Active Orders Across All Symbols.
- `1` / `2` / `5` / `0`: Quick Lot Size Selector (10%, 25%, 50%, 100% of Available Margin).

---

## 4. Native Desktop Thick Client Architecture (macOS & Windows Main Focus)

For proprietary trading desks, professional scalpers, and institutional asset managers, the **Growww Desktop Thick Client** (`apps/growww_flutter` desktop runner) provides an uncompromised, native OS operating system experience that transcends browser sandbox constraints.

```
+-------------------------------------------------------------------------------------------------------+
|                                  GROWWW DESKTOP THICK CLIENT CORE                                     |
|                                                                                                       |
|   +-----------------------------------------------------------------------------------------------+   |
|   |                         MASTER PROCESS (Desktop Multi-Window Controller)                      |   |
|   |  - Coordinates Multi-Monitor Window Geometries (Coordinates, Bounds, DPI, Display Affinity) |   |
|   |  - Owns Master WebSocket Ingestion & Rust FFI Ring Buffer                                     |   |
|   |  - Houses Secure OS Credential Vault (macOS Keychain / Windows Credential Manager)           |   |
|   +-----------------------------------------------------------------------------------------------+   |
|                      |                                                     |                          |
|         Platform-Specific Native Channel                      High-Speed Cross-Process IPC            |
|         (AppKit / Win32 / GTK)                                (Named Pipes / Memory Mapped File)      |
|                      |                                                     |                          |
|         +------------+------------+                       +----------------+----------------+         |
|         |                         |                       |                                 |         |
|         v                         v                       v                                 v         |
|   +-------------------+   +--------------------+    +--------------------+    +--------------------+  |
|   | macOS Subsystem   |   | Windows Subsystem  |    | Detached Window 1  |    | Detached Window 2  |  |
|   | - Metal Impeller  |   | - DirectX 12 Render|    | - DOM Depth Ladder |    | - Multi-Timeframe  |  |
|   | - Apple Silicon   |   | - Windows Hello    |    | - Click-to-Trade   |    |   TradingView Chart|  |
|   | - Secure Enclave  |   | - Win32 HotKey API |    | - Auto-Centering   |    | - 4K Monitor Scaled|  |
|   | - Cocoa Menu Bar  |   | - Tray Min-Monitor |    | - Sub-ms Updates   |    | - Independent DPI  |  |
|   +-------------------+   +--------------------+    +--------------------+    +--------------------+  |
|                                                                                                       |
|   +-----------------------------------------------------------------------------------------------+   |
|   |                        NATIVE RUST FFI CORE (`rust_trading_core.dll` / `.dylib`)              |   |
|   |  - Tokio async WebSocket Client with TCP_NODELAY and SO_RCVBUF = 4MB                          |   |
|   |  - SIMD-accelerated binary Fast-SBE / Protobuf parsing directly into C-ABI ring buffers       |   |
|   |  - Zero Dart garbage collection overhead on market tick hot paths                             |   |
|   +-----------------------------------------------------------------------------------------------+   |
+-------------------------------------------------------------------------------------------------------+
```

### 4.1 macOS Native Implementation (Apple Silicon & Metal)
1. **Metal Acceleration (Impeller Engine):**
   - Renders complex 50-level order book depth profiles, high-frequency tick tapes, and multi-pane TradingView charts at 120 FPS on Apple ProMotion displays (MacBook Pro, Studio Display, Pro Display XDR).
   - Zero GPU hitching or thermal throttling during all-day market sessions.
2. **Apple Silicon Optimization:**
   - Universal binary (`arm64` + `x86_64`) with native ARM64 NEON SIMD optimizations compiled directly into the Rust FFI core.
3. **Apple Secure Enclave & TouchID:**
   - Cryptographic signing keys and API authentication secrets are protected in the macOS Keychain backed by the hardware Secure Enclave.
   - Every withdrawal, high-notional trade, or API key generation triggers biometric validation via Apple `LocalAuthentication`:
     ```swift
     // macOS LocalAuthentication integration
     let context = LAContext()
     var error: NSError?
     if context.canEvaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, error: &error) {
       context.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, localizedReason: "Authorize Instant NBSE Trade Execution") { success, error in
         if success {
           // Release hardware-bound signing token
         }
       }
     }
     ```
4. **Native Cocoa Menu Bar & Global Shortcuts:**
   - Native macOS Menu Bar (`NSMenu`) housing command palettes, workspace layout presets, and keyboard shortcuts (`Cmd+1` to `Cmd+9` workspace switching).
   - Menu Bar extra item (status icon) displaying live ticker price and daily PnL without needing an active foreground window.

### 4.2 Windows Native Implementation (DirectX 12 & Windows Hello)
1. **DirectX 12 / Direct3D Rendering:**
   - Direct hardware acceleration interfacing directly with NVIDIA, AMD, and Intel GPUs.
   - Unlocked refresh rates supporting 144Hz, 240Hz, and 360Hz ultra-fast gaming and trading monitors with adaptive G-Sync and FreeSync synchronization.
2. **Per-Monitor DPI v2 Awareness:**
   - Seamlessly handles multi-monitor trading setups where display scaling factors differ (e.g., a 32-inch 4K monitor at 150% scaling adjacent to a 27-inch 1440p monitor at 100% scaling). Windows detach and drag between monitors with zero visual glitching or coordinate miscalculations.
3. **Windows Hello Biometric Quick-Auth:**
   - Integrates with `Windows.Security.Credentials.KeyCredentialManager` to provide instantaneous facial recognition (Windows Hello Camera) and fingerprint authentication for high-value orders.
4. **OS-Level Global Keyboard Hooks (`RegisterHotKey`):**
   - Active traders can press `Ctrl + Alt + Space` to trigger the Global Panic Cancel-All action from anywhere in Windows, even when another application (such as Excel or Bloomberg Terminal) has foreground focus.
5. **System Tray Integration:**
   - Minimizes to Windows System Tray with dynamic tooltips showing live portfolio value, net daily gain/loss, and Besu ledger block confirmation height.

### 4.3 High-Frequency Depth of Market (DOM) Click-to-Trade Ladder
Both macOS and Windows thick clients feature an institutional DOM price ladder:
- **Fixed Central Price Rungs:** Price ladder remains steady while bid/ask size bars fluctuate dynamically.
- **Single-Click Order Placement:**
  - Left-Click on Bid Column: Submits Limit Buy order at exact clicked price rung.
  - Left-Click on Ask Column: Submits Limit Sell order at exact clicked price rung.
  - Right-Click on Any Working Order: Instant single-order cancellation.
  - Middle-Click / Space: Auto-centers ladder around current Best Bid / Best Offer midpoint.

---

## 5. Mobile Client Architecture (Android & iOS)

The mobile client (`apps/growww_flutter`) brings sovereign exchange capabilities into retail traders' pockets with zero compromise on security or real-time speed.

### 5.1 Mobile Design & Ergonomics
- **Obsidian Dark Theme:** Deep Obsidian (`#0B0E14`) background hierarchy, `#141822` elevated surface cards, `#1F2636` structural borders.
- **Neon Trading Accents:** Precision `#00F0A0` (Buy / Profit / Up-Tick) and `#FF3B56` (Sell / Loss / Down-Tick).
- **Tabular Font Formatting:** `JetBrains Mono` and `Roboto Mono` with `fontFeatures: [FontFeature.tabularFigures()]` preventing layout jitter during price fluctuations.
- **Tactile Haptic Feedback Engine:** Native platform channel triggers micro-vibrations:
  - `HapticFeedback.selectionClick()`: Ticker switching and tab selection.
  - `HapticFeedback.mediumImpact()`: Order submission confirmation.
  - `HapticFeedback.heavyImpact()`: Circuit breaker volatility warning or liquidation margin alert.

### 5.2 Battery & Network Optimization Lifecycle
Mobile devices face constrained battery and variable cellular connections (4G/5G/Wi-Fi):
1. **Background Lifecycle Transition:**
   - When the app is backgrounded, the active WebSocket stream gracefully disconnects within 15 seconds to conserve battery and cellular data.
   - Important alerts (order fills, margin alerts, deposits) transition to low-overhead Apple Push Notification service (APNs) and Firebase Cloud Messaging (FCM).
2. **Foreground Instant Resume:**
   - Upon app foregrounding, the client immediately initiates a parallel reconnection:
     - Fetches a REST L2 depth snapshot (`/api/v1/market/depth/snapshot?symbol=BTC-USDT`).
     - Subscribes to the live WebSocket delta stream.
     - Buffers deltas and applies only those where `delta.sequence_id > snapshot.last_sequence_id`, ensuring zero order book drift.

### 5.3 Offline Resilient Order Queuing (Encrypted SQLite WAL)
If an active trader loses cellular connectivity in a transit tunnel:
- Orders submitted in "Offline Queue Mode" are stored in local encrypted SQLite database (AES-256-GCM in WAL mode).
- As soon as network connectivity is restored, the queue processor verifies order timestamp freshness (<30 seconds validity window) and dispatches orders with unique idempotent `client_order_id` UUIDs.

---

## 6. Cross-Platform Synchronization Engine ("Everywhere")

The "Everywhere" architecture guarantees that an active trader moving between an office multi-monitor Windows workstation, a home MacBook Pro, a web browser on a borrowed laptop, and an iPhone on the train experiences absolute state continuity.

```
+-------------------------------------------------------------------------------------------------------+
|                                  CROSS-PLATFORM SYNCHRONIZATION TOPOLOGY                              |
|                                                                                                       |
|    +--------------------+    +--------------------+    +--------------------+    +----------------+   |
|    | Web Pro Terminal   |    | macOS Thick Client |    | Windows Workstation|    | Mobile Apps    |   |
|    | (Next.js 14)       |    | (Flutter / Metal)  |    | (Flutter / DirectX)|    | (iOS / Android)|   |
|    +--------------------+    +--------------------+    +--------------------+    +----------------+   |
|              |                         |                         |                        |           |
|              +-------------------------+------------+------------+------------------------+           |
|                                                     |                                                 |
|                                        gRPC / TLS 1.3 Synchronization                                 |
|                                                     |                                                 |
|                                                     v                                                 |
|                           +---------------------------------------------------+                       |
|                           |      CENTRAL WORKSPACE & STATE SYNC SERVICE       |                       |
|                           |         (`services/workspace-sync-service`)       |                       |
|                           +---------------------------------------------------+                       |
|                                                     |                                                 |
|                       +-----------------------------+-----------------------------+                   |
|                       v                                                           v                   |
|       +-------------------------------+                           +-------------------------------+   |
|       | Encrypted Redis Layout Cache  |                           | PostgreSQL Layout Repository  |   |
|       | (Sub-millisecond State Query) |                           | (Multi-Monitor Coordinates)   |   |
|       +-------------------------------+                           +-------------------------------+   |
+-------------------------------------------------------------------------------------------------------+
```

### 6.1 Multi-Monitor Workspace Profiles
Traders configure and save multiple workspace profiles:
- **"Day Trader Setup":** 3-screen configuration (Screen 1: Watchlist & L2 Order Book; Screen 2: 4K 1-Minute Candlestick Chart; Screen 3: DOM Price Ladder & Execution Blotter).
- **"Scalper Setup":** Ultra-dense layout focusing on dual DOM ladders (BTC/USDT and eINR/USDT) with 1-click execution hotkeys.
- **"Analytics & Macro":** Multi-chart view (1m, 15m, 1h, 1d) with technical overlays and news feed.

Workspaces serialize into a canonical cross-platform JSON/Protobuf specification and synchronize automatically to the user's account.

### 6.2 Color-Coded Link Groups (Ticker Synchronization)
Windows and panels support 4 distinct Link Groups:
- **Group Red:** Linked to Primary Scalping Pair (e.g. BTC/USDT).
- **Group Blue:** Linked to Secondary Major (e.g. ETH/USDT).
- **Group Green:** Linked to Commodity / Gold Token (e.g. w-GOLD/eINR).
- **Group Yellow:** Linked to Equity / Index Token (e.g. NBSE-50/eINR).

Selecting any instrument in a Watchlist assigned to "Group Red" instantly switches every Chart, Order Book, Depth Ladder, and Blotter assigned to "Group Red" across all physical monitors simultaneously.

### 6.3 Universal Sub-Paise Precision & Invariant Display
Across every platform (Web, macOS, Windows, Linux, iOS, Android):
- **Fixed-Point Arithmetic:** Zero IEEE-754 floating-point calculations in client-side financial math. All figures use 64-bit integer micro-units (`10^-6` eINR / 0.0001 paise) formatted via locale-aware formatters.
- **Zero-Fee Presentation Invariant:**
  - Order Entry Confirmation: `Exchange Trading Fee: ₹0.00 (0.00% Genesis Zero-Fee)`.
  - Blockchain Settlement Status: `Gas Sponsoring Paymaster: ₹0.00 (ERC-4337 Sponsored)`.
  - Statutory TDS Deduction: `On-Chain TDS: ₹0.00 (100% Unencumbered DvP Settlement)`.
- **Zero PII on Client Displays:** No PAN numbers, Aadhaar numbers, or physical street addresses are ever rendered on trading screens or exposed in client crash logs.

---

## 7. Build, Packaging & Distribution Pipeline

Every target platform follows an automated, cryptographically signed release pipeline:

```
+-------------------------------------------------------------------------------------------------------+
|                                CI/CD MULTI-PLATFORM PACKAGING MATRIX                                  |
|                                                                                                       |
|  [ Source Commit ] ---> [ GitHub Actions Multi-Runner Build Matrix ]                                  |
|                                    |                                                                  |
|       +----------------------------+----------------------------+-----------------------------+       |
|       |                            |                            |                             |       |
|       v                            v                            v                             v       |
|  [ macOS Runner ]          [ Windows Runner ]           [ Ubuntu Linux Runner ]       [ Web Vercel/CDN ]|
|  - Xcode 15 / Metal SDK    - MSVC v143 / Win32 SDK      - GCC / GTK3 / Vulkan         - Next.js 14 App|
|  - Flutter build macos     - Flutter build windows      - Flutter build linux         - Static + Edge |
|  - codesign & Notarization - SignTool (EV Code Sign)   - Flatpak & AppImage build    - Docker Web Host|
|  - Outputs: `.dmg`, `.pkg` - Outputs: `.msix`, `.exe`   - Outputs: `.flatpak`, `.deb` - Edge CDN URLs |
+-------------------------------------------------------------------------------------------------------+
```

### 7.1 Release Artifact Standards:
- **macOS:** Universal Mach-O binary signed with Apple Developer ID and notarized via `xcrun notarytool`. Hardened Runtime enabled with App Sandbox exemptions for multi-window IPC.
- **Windows:** 64-bit executable signed with Extended Validation (EV) Code Signing Certificate. Packaged as both an auto-updating MSIX package and a standalone standalone installer.
- **Linux:** Standalone AppImage and sandboxed Flatpak package with explicit Wayland and X11 socket permissions.
- **Web:** Hosted on multi-region Edge CDN with HTTP/3, Brotli compression, strict Content Security Policy (CSP), and PWA service worker caching.
- **iOS:** Apple App Store release IPA signed with Apple Distribution Certificate, compliant with App Store Review Guideline 3.1.5 (Cryptocurrency & Financial Exchanges).
- **Android:** Android App Bundle (AAB) signed via Google Play App Signing, targeting Android 14 (API level 34) with backward compatibility to Android 8.0 (API level 26).

---

## 8. Verification & Acceptance Criteria

To certify universal platform readiness, the following acceptance test matrix must pass with 100% compliance across all 6 targets:

| Test Case | Description | Target Platforms | Acceptance Threshold |
| :--- | :--- | :--- | :--- |
| **TC-CLI-01** | High-Throughput Depth Rendering | Web, macOS, Windows | 50ms conflated stream renders at >=60 FPS without frame drops under 50,000 updates/sec. |
| **TC-CLI-02** | Multi-Monitor Window Detachment | Web (Pop-out), macOS, Windows | Detaching 4 windows across 3 physical displays maintains sub-1ms symbol link synchronization. |
| **TC-CLI-03** | Global Panic Cancel Hotkey | Web, macOS, Windows | Pressing panic hotkey cancels all working orders in <=15ms from keypress to gateway dispatch. |
| **TC-CLI-04** | Biometric Trade Authorization | macOS, Windows, iOS, Android, Web | TouchID, Windows Hello, FaceID, and WebAuthn successfully sign cryptographic order payloads. |
| **TC-CLI-05** | Offline Network Reconnection | iOS, Android, macOS, Windows | Reconnect after 60-second airplane mode resynchronizes order book snapshot within <=200ms. |
| **TC-CLI-06** | Zero-Fee Display Compliance | All 6 Platforms | All order forms, confirmation modals, and blotters explicitly display ₹0.00 / 0.00% fees. |
| **TC-CLI-07** | Memory Stability Soak Test | Web, macOS, Windows | 24-hour continuous streaming test shows <=150MB baseline RAM with zero memory leaks. |
