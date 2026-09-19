# Universal Flutter Client Architecture (macOS, Windows, Linux, iOS, Android)

## Executive Overview
`growww_flutter` is engineered as a clean-architecture, feature-first universal client application powering both retail mobile apps (iOS, Android) and multi-monitor institutional desktop workstations (macOS, Windows, Linux).

Reference Master Blueprint: [`docs/architecture/UNIVERSAL_MULTI_PLATFORM_CLIENT_ARCHITECTURE_AND_THICK_CLIENT_BLUEPRINT.md`](../../docs/architecture/UNIVERSAL_MULTI_PLATFORM_CLIENT_ARCHITECTURE_AND_THICK_CLIENT_BLUEPRINT.md)

## Desktop Thick Client Architecture (macOS, Windows, Linux - Main Focus)

### 1. Multi-Window Lifecycle & Detachment Manager
- **Master-Worker Topology (`desktop_multi_window`):** Master controller manages application state, credential vaults, and root WebSocket feeds. Detached viewports (DOM price ladder, full-screen 4K chart, order blotter) run as autonomous native OS windows.
- **Multi-Monitor Display Awareness:** Detects display topology, per-monitor DPI scaling v2, and display refresh rates (60Hz to 360Hz).
- **Inter-Process IPC:** Sub-millisecond cross-window state synchronization via native OS pipes and memory-mapped files (`growww.desktop.ipc.v1`).

### 2. Platform-Specific Native Rendering & Subsystems
- **macOS (AppKit / Metal):**
  - Impeller Metal graphics pipeline locking 120 FPS on Apple ProMotion displays.
  - Native Cocoa Menu Bar (`NSMenu`) with global shortcut commands (`Cmd+1` - `Cmd+9`).
  - Apple Secure Enclave & TouchID via `LocalAuthentication` framework.
  - Dock icon badge displaying real-time daily PnL and active order alerts.
- **Windows (Win32 / DirectX 12):**
  - Direct3D / DirectX 12 acceleration driving high-refresh gaming and trading monitors (144Hz, 240Hz, 360Hz).
  - Windows Hello biometric quick-auth (facial recognition & fingerprint reader).
  - Global OS-level keyboard hooks (`RegisterHotKey`) enabling instant panic cancel-all even when unfocused.
  - System Tray integration with background ticker status and desktop notifications.
- **Linux (GTK 3/4 & Wayland/X11):**
  - Vulkan-backed hardware rendering across multi-head setups.
  - Linux Secret Service API (`libsecret`) integration for encrypted key storage.

### 3. Native Rust FFI Socket Engine (`rust_trading_core`)
- High-throughput Tokio async WebSocket client bound via Dart FFI.
- Bypasses Dart garbage collector on market tick hot paths, parsing binary Fast-SBE / Protobuf directly into C-ABI ring buffers.
- Maintains locked 60/120 FPS UI frame rates under 50,000+ depth ticks per second.

## Mobile Client Architecture (iOS & Android)

### 1. State Management Topology (Riverpod 2.5)
- **AsyncNotifierProviders:** Manage asynchronous data states with optimistic local updates.
- **WebSocket Stream Providers:** Expose live Level 2 depth diffs and ticker streams directly to UI widgets.
- **Persistent State:** Hydrated local state using SQLite WAL mode ensuring offline responsiveness.

### 2. Navigation & Routing (GoRouter)
- Deep linking support with universal link verification for deposit receipts and referral joins.
- Adaptive navigation: Bottom navigation bar on mobile phones, persistent vertical navigation rail on tablets, foldables, and desktop displays.

### 3. Mobile Battery & Connectivity Lifecycle
- **Background Suspension:** Graceful socket suspension within 15 seconds to eliminate battery drain; background updates route via APNs / FCM.
- **Foreground Fast Resumption:** Reconnects, fetches REST L2 snapshot, and reconciles buffered WebSocket deltas where `sequence_id > snapshot.last_sequence_id`.
- **Encrypted SQLite Offline Queue:** Stores queued orders during cellular dead zones, dispatches upon reconnect with idempotency keys.

## UI / UX Design Tokens
- **Theme:** Pure Deep Obsidian (`#0B0E14`) background hierarchy with zero-layout-shift tabular numbers (`JetBrains Mono`).
- **Precision Trading Accents:** Neon Green (`#00F0A0`) and Neon Red (`#FF3B56`).
- **Tactile Haptic Triggers:** Native selection pulses, medium impact order confirmations, and heavy impact circuit-breaker alerts.
- **Zero-Fee Presentation Invariant:** Strict 0.00% maker / 0.00% taker / 0 gas / 0 TDS displayed on all order entry and confirmation sheets.
