# 532 - Flutter Desktop Multi-Window Institutional Trading Terminal

## Purpose
Institutional desks, proprietary trading firms, market makers, and professional active traders require an expansive multi-monitor desktop workspace that spans two to six high-resolution displays. Single-window mobile or web viewports create severe cognitive and physical bottlenecks during fast-moving market sessions. Active traders must simultaneously monitor tick-by-tick Level 2 and Level 3 order book depth, track Depth of Market (DOM) click-to-trade ladders, review algorithmic trade execution blotters, observe multi-timeframe candlestick charts, and verify blockchain settlement telemetry without constantly switching tabs or losing contextual focus.

The **Flutter Desktop Multi-Window Institutional Trading Terminal** delivers a production-grade, hardware-accelerated desktop workstation for Windows (Win32 / WinUI), macOS (AppKit / Cocoa), and Linux (X11 / Wayland). Operating across multiple physical monitors with disparate DPI scaling factors, the terminal allows traders to detach any viewport (such as order books, DOM ladders, or execution blotters) into autonomous native operating system windows. The architecture guarantees sub-millisecond inter-process communication (IPC) for symbol synchronization, locks rendering at 60 to 120 frames per second under high-throughput market tick loads via a Rust FFI socket core, registers global keyboard shortcuts for instantaneous execution, and provides real-time cryptographic auditability for Hyperledger Besu on-chain settlements.

## What You Are Building
A modular, high-performance desktop architecture located in `apps/growww_flutter/lib/features/desktop_terminal/` and integrated with native desktop runners:
- `DesktopMultiWindowManager`: Central coordinator orchestrating native OS window lifecycles (spawning, positioning, docking, snapping, minimizing, restoring, closing), multi-monitor display metrics detection, and layout persistence.
- `DetachableOrderBookWindow`: Dedicated auxiliary native window rendering full L2/L3 order books with visual bid/ask volume depth distributions, microsecond price updates, tick-by-tick tape flow, and cumulative volume profile charts.
- `RealTimeDepthLadderWindow`: High-frequency DOM (Depth of Market) price ladder widget featuring static central price rungs, dynamic bid/ask volume bars, single-click order entry/cancellation at exact price increments, auto-centering locks, and hotkey-triggered size scaling.
- `InstitutionalExecutionBlotterWindow`: Auxiliary window displaying real-time working orders, completed executions, algorithmic slicing progress (TWAP/VWAP), net desk positions, realized/unrealized PnL, and manual panic-cancel controls.
- `BlockchainSettlementMonitorWindow`: Live distributed ledger monitor displaying Hyperledger Besu block confirmation progress, on-chain transaction hashes, block explorer deep-links, and gasless Paymaster sponsorship badges (ERC-4337).
- `GlobalShortcutManager`: Desktop-wide keyboard hotkey registry capturing rapid trading commands (such as F1-F12 symbol selection, Spacebar for global panic cancel-all, Ctrl+B/Ctrl+S for instant market entry, and keypad navigation for depth centering).
- `TradingIPCBridge`: High-throughput inter-process communication channel utilizing `desktop_multi_window` message protocols and memory-mapped state synchronization to replicate active ticker symbols, theme palettes, and authorization sessions across child windows.
- `RustFFISocketEngine`: Native dynamic library integration (`rust_trading_core`) written in Rust and bound via Dart FFI to ingest binary WebSocket market data streams, process ring buffers, and offload deserialization overhead from the Dart UI event loop.
- `DesktopWorkspaceStorage`: Encrypted local layout repository persisting multi-monitor window coordinates, dock states, active tickers, and customized shortcut bindings across application restarts.

## Scope Boundaries
- **In Scope:**
  - Multi-window lifecycle management (creating, positioning, docking, detaching, snapping, closing) across Windows, macOS, and Linux.
  - Multi-monitor display topology awareness supporting mixed DPI scaling (e.g. 4K monitor at 150% scaling paired with 1440p monitor at 100% scaling).
  - High-frequency DOM (Depth of Market) price ladder with microsecond visual updates, click-to-trade order entry, and keyboard auto-centering.
  - Inter-window state synchronization (active symbol link groups, workspace configurations, cross-window theme updates, session state).
  - Desktop-wide and window-focused hotkey mapping with conflict resolution and customizable trading action bindings.
  - Rust FFI native bindings (`rust_trading_core`) for high-throughput WebSocket tick ingestion and ring buffer processing.
  - Hardware-accelerated Flutter rendering using Impeller (Metal on macOS, Vulkan/DirectX on Windows and Linux) optimized with `RepaintBoundary` and custom `RenderBox` implementations.
  - Inactivity screen lock synchronously shielding all active windows and isolating sensitive authentication tokens.
  - On-chain settlement confirmation counters, transaction hash displays, and gasless Paymaster badge rendering.
- **Out of Scope / Handled Elsewhere:**
  - Server-side Order Matching Engine core matching algorithms and risk validation (handled in Prompt 205).
  - FIX Protocol Gateway connectivity and institutional binary OUCH/ITCH exchange feeds (handled in Prompt 225).
  - Algorithmic execution engine calculations for TWAP/VWAP child order slicing (handled in Prompt 257).
  - Smart contract settlement logic and Account Abstraction Bundler/Paymaster backend infrastructure (handled in Prompt 259, Prompt 315).
  - Mobile-specific gesture navigation and responsive bottom bar UI (handled in Prompts 506 through 510).

## Technology to Use
- **Primary Framework:** Flutter 3.22+ with Dart 3.4+ targeting desktop operating systems (Windows x86_64, macOS Apple Silicon/x86_64, Linux x86_64).
- **Multi-Window Orchestration:** `desktop_multi_window: ^0.2.0` for spawning native OS windows with isolated Flutter engines sharing OS window handles, paired with `window_manager: ^0.3.9` for window frame manipulation, boundary positioning, and multi-monitor enumeration.
- **High-Performance Native Ingestion:** Rust FFI native dynamic library (`rust_trading_core`) via `flutter_rust_bridge: ^2.0.0` or direct `dart:ffi`, implementing high-throughput Tokio WebSocket client, lock-free ring buffers, and SIMD order book delta aggregation.
- **State Management:** `flutter_riverpod: ^2.5.1` with code-generated `AsyncNotifier` and `StreamNotifier` state primitives, integrated with cross-process IPC dispatchers.
- **Keyboard Shortcut Capture:** `hotkey_manager: ^0.2.0` combined with Flutter native `HardwareKeyboard` and `Shortcuts`/`Actions` system for comprehensive desktop hotkey bindings.
- **Local Cache & Storage:** `flutter_secure_storage: ^9.2.2` (Windows Credential Manager, macOS Keychain, Linux Secret Service) and encrypted SQLite/Isar for multi-monitor workspace layout geometry persistence.
- **Rendering Pipeline:** Hardware-accelerated Flutter Impeller engine (Metal / Vulkan / DirectX) utilizing custom `RenderBox` and `CustomPainter` widgets wrapped in `RepaintBoundary` to prevent canvas invalidation across static UI elements.
- **Fixed-Point Financial Mathematics:** `decimal: ^2.3.3` ensuring sub-paise precision (0.0001 INR) without IEEE-754 floating-point inaccuracies.

## Backend / Infra Touchpoints
- **Market Data WebSocket Service (Prompt 207):**
  - **Endpoint:** `wss://ws.growww.in/v1/market/depth/stream` (binary Protocol Buffer / FlatBuffers format).
  - Delivers sub-10ms Level 2 (20-depth, 50-depth) and Level 3 order book deltas, aggregated trades, and best bid/offer quotes.
  - Ingested directly by the Master Window Rust FFI core and distributed across child windows via IPC.
- **Order Gateway & Execution Service (Prompt 204):**
  - **Endpoints:** `POST /api/v1/orders/quick-entry`, `DELETE /api/v1/orders/cancel-all`, `DELETE /api/v1/orders/{order_id}`.
  - Low-latency HTTP/2 or gRPC order dispatch for DOM ladder click-to-trade and panic cancel hotkeys.
- **Local Secure Storage & Session Storage (Prompt 521):**
  - Master process retrieves session credentials from OS hardware vaults; child worker windows access scoped cryptographic session tokens over IPC without duplicating master secrets in memory.
- **Workspace Layout Cloud Synchronization Service (Prompt 405 / User Profile):**
  - **Endpoints:** `GET /api/v1/desktop/workspaces`, `PUT /api/v1/desktop/workspaces/{workspace_id}`.
  - Synchronizes multi-monitor window geometry profiles, custom color tags, and hotkey presets across trader workstations.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Besu Block Confirmation Telemetry:**
  - The Institutional Execution Blotter and Blockchain Settlement Monitor display live Hyperledger Besu block heights, validator consensus round metrics, and dynamic confirmation counters (e.g. `12 / 12 Confirmations - Finalized`).
- **On-Chain Transaction Hash Tracking:**
  - Each settled trade execution displays an immutable on-chain transaction hash (e.g. `0x7a3f...9b12`) with a one-click action to inspect the transaction details on the internal Besu Block Explorer.
- **Gasless Paymaster Badge & ERC-4337 Account Abstraction:**
  - A prominent status badge (`[Paymaster Sponsored: ₹0.00 Gas Fee]`) confirms that institutional settlement operations are sponsored by the Growww Enterprise Paymaster contract, abstracting native gas token friction while maintaining cryptographic audit trails.
- **Zero On-Chain PII Guarantee:**
  - The client displays only pseudonymous asset-level telemetry: smart contract addresses, Merkle settlement roots, and cryptographic trade batch identifiers. No trader names, PAN identifiers, or IP addresses are written to or queried from the blockchain.

## Desktop Multi-Window Architecture & Inter-Process Synchronization

### 1. Master-Worker Window Topology
The Flutter desktop architecture employs a Master-Worker topology where the primary application instance acts as the Master Controller, and all auxiliary detached viewports run as isolated Worker processes managed by `desktop_multi_window`:

```
+-----------------------------------------------------------------------------------+
|                                 MASTER PROCESS                                    |
|  +-----------------------------------------------------------------------------+  |
|  | Main Terminal Shell (Watchlists, Navigation, Settings, Layout Manager)     |  |
|  +-----------------------------------------------------------------------------+  |
|  | Rust FFI Socket Core (High-Throughput WebSocket Ingestion & Ring Buffer)    |  |
|  +-----------------------------------------------------------------------------+  |
|  | Authentication Guardian (Hardware Enclave Storage & Session Token Issuer)   |  |
|  +-----------------------------------------------------------------------------+  |
|                                       |                                           |
|                  IPC Message Bus (Native OS Pipes / Ports)                        |
|        +------------------------------+------------------------------+            |
+--------|------------------------------|------------------------------|------------+
         |                              |                              |
         v                              v                              v
+------------------+          +------------------+          +-------------------+
|  WORKER WINDOW 1 |          |  WORKER WINDOW 2 |          |  WORKER WINDOW 3  |
|  Detachable DOM  |          |  Detachable L2/3 |          |  Execution        |
|  Depth Ladder    |          |  Order Book View |          |  Blotter Monitor  |
+------------------+          +------------------+          +-------------------+
```

- **Master Process:** Owns the primary WebSocket connection to the backend market data service, runs the Rust FFI socket ingestion engine, handles user authentication, and orchestrates window layout configurations.
- **Worker Processes (Detached Windows):** Spawned on demand with dedicated native window handles (`HWND` on Windows, `NSWindow` on macOS, `GtkWindow` on Linux). Each runs an isolated Flutter engine with independent widget trees and render pipelines, eliminating main-thread UI contention.
- **IPC Protocol:** Inter-process messages are transmitted using binary or compact JSON packets over platform-native communication channels. The protocol supports:
  - *Broadcast Channel:* Master broadcasts market tick deltas, active symbol switches, theme changes, and global screen lock commands to all active workers.
  - *Command Channel:* Workers send order execution intents, window focus notifications, and detachment requests back to the Master for centralized routing.

### 2. Multi-Monitor Coordinate Mapping & Window Snapping
- **Virtual Canvas Calculation:** The system queries display boundaries via `window_manager` and constructs a unified virtual desktop coordinate space that accounts for variable monitor dimensions, relative physical positioning, and per-display DPI scaling factors.
- **Magnetic Window Snapping:** When dragging a detachable window within a threshold distance (e.g. 16 logical pixels) of a monitor edge or adjacent window boundary, the window manager magnetically snaps the frame to align perfectly with adjacent trading widgets.
- **Layout Profiles:** Pre-configured and user-defined window layouts (e.g. "Triple-Monitor Arbitrage Desk", "Dual-Monitor Execution Suite", "Single 4K Ultrawide Grid") can be saved and restored with exact pixel coordinates.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Desktop Feature Directory:** Create `apps/growww_flutter/lib/features/desktop_terminal/` with subdirectories `presentation/windows/`, `presentation/widgets/`, `presentation/controllers/`, `domain/models/`, `domain/interfaces/`, `data/ipc/`, `data/ffi/`, and `data/storage/`.
2. **Configure Native Desktop Runners:** Update native configuration files for Windows (`windows/runner/main.cpp`), macOS (`macos/Runner/MainFlutterWindow.swift`), and Linux (`linux/my_application.cc`) to initialize multi-window hooks and set window boundaries.
3. **Integrate Rust FFI Socket Engine:** Configure `rust_trading_core` dynamic library bindings using `dart:ffi`, implementing native Tokio WebSocket listeners, ring buffer management, and SIMD order book delta aggregation.
4. **Implement Window Manager Coordinator:** Build `DesktopMultiWindowManager` utilizing `desktop_multi_window` and `window_manager` to handle window creation, boundary positioning, monitor enumeration, and termination.
5. **Implement Trading IPC Bridge:** Build `TradingIPCBridge` implementing master-to-worker broadcasting and worker-to-master command dispatching with strict message type safety and serialization schemas.
6. **Build Detachable Order Book Window:** Implement `DetachableOrderBookWindow` rendering real-time L2/L3 bid and ask depth columns, cumulative volume distribution bars, and tick-by-tick trade tape feeds.
7. **Build Real-Time Depth of Market (DOM) Ladder:** Implement `RealTimeDepthLadderWindow` featuring a static central price axis, dynamic volume bars, click-to-trade order entry/cancellation, and auto-centering price locks.
8. **Build Institutional Execution Blotter Window:** Implement `InstitutionalExecutionBlotterWindow` displaying open orders, fills, algorithmic execution progress, net positions, and one-click panic cancel-all buttons.
9. **Build Blockchain Settlement Monitor Widget:** Implement `BlockchainSettlementWidget` displaying real-time Hyperledger Besu block confirmation counts, on-chain transaction hashes, and gasless Paymaster badges.
10. **Implement Global Shortcut Manager:** Build `GlobalShortcutManager` utilizing `hotkey_manager` and Flutter `HardwareKeyboard` to capture trading shortcuts (e.g. F1-F12, Spacebar panic cancel, Ctrl+B/Ctrl+S) with conflict resolution.
11. **Build Multi-Monitor Workspace Persistence Engine:** Implement `DesktopWorkspaceStorage` to serialize window geometries, monitor indices, active symbols, and color-coded link groups to encrypted local storage.
12. **Implement Desktop Inactivity Screen Lock:** Build `DesktopInactivityLockManager` to monitor system-wide user idle time and synchronously blank or lock all active windows with biometric or password unlock prompts.
13. **Optimize Hardware-Accelerated Rendering:** Wrap high-frequency widgets in `RepaintBoundary`, implement custom `RenderBox` layouts for the DOM ladder, and verify locked 60/120 FPS performance under 5,000 updates/sec.
14. **Author Comprehensive Desktop Test Suite:** Write unit tests for IPC serialization and hotkey dispatching, widget tests for DOM ladder interactions, and integration tests for multi-window spawn and close cycles.

## Interfaces / Contracts

### 1. Window State Synchronization Models (`lib/features/desktop_terminal/domain/models/window_models.dart`)
```dart
enum WindowType {
  masterTerminal,
  orderBookL2L3,
  depthOfMarketLadder,
  executionBlotter,
  blockchainSettlement,
  advancedChart,
}

enum SymbolLinkGroup {
  groupA, // Red
  groupB, // Blue
  groupC, // Green
  groupD, // Yellow
  none,
}

class WindowGeometry {
  final double x;
  final double y;
  final double width;
  final double height;
  final int monitorIndex;
  final bool isMaximized;
  final bool isAlwaysOnTop;

  const WindowGeometry({
    required this.x,
    required this.y,
    required this.width,
    required this.height,
    required this.monitorIndex,
    this.isMaximized = false,
    this.isAlwaysOnTop = false,
  });

  Map<String, dynamic> toJson();
  factory WindowGeometry.fromJson(Map<String, dynamic> json);
}

class DesktopWindowState {
  final int windowId;
  final WindowType windowType;
  final String title;
  final WindowGeometry geometry;
  final String activeSymbol;
  final SymbolLinkGroup linkGroup;
  final bool isDetached;
  final DateTime lastActiveTimestamp;

  const DesktopWindowState({
    required this.windowId,
    required this.windowType,
    required this.title,
    required this.geometry,
    required this.activeSymbol,
    this.linkGroup = SymbolLinkGroup.none,
    this.isDetached = false,
    required this.lastActiveTimestamp,
  });

  Map<String, dynamic> toJson();
  factory DesktopWindowState.fromJson(Map<String, dynamic> json);
}
```

### 2. IPC Message Protocol (`lib/features/desktop_terminal/domain/models/ipc_models.dart`)
```dart
enum IPCOpCode {
  broadcastTick,
  broadcastSymbolChange,
  broadcastThemeChange,
  broadcastScreenLock,
  commandSubmitOrder,
  commandCancelOrder,
  commandCancelAll,
  commandFocusWindow,
  commandCloseWindow,
  heartbeatPing,
  heartbeatPong,
}

class IPCMessage {
  final String messageId;
  final IPCOpCode opCode;
  final int sourceWindowId;
  final int? targetWindowId; // null indicates broadcast to all windows
  final Map<String, dynamic> payload;
  final int timestampMicroseconds;

  const IPCMessage({
    required this.messageId,
    required this.opCode,
    required this.sourceWindowId,
    this.targetWindowId,
    required this.payload,
    required this.timestampMicroseconds,
  });

  String serialize();
  factory IPCMessage.deserialize(String raw);
}
```

### 3. Hotkey Configuration Schema (`lib/features/desktop_terminal/domain/models/hotkey_models.dart`)
```dart
enum TradingAction {
  panicCancelAllOrders,
  cancelBidsOnly,
  cancelAsksOnly,
  instantMarketBuy,
  instantMarketSell,
  centerDepthLadder,
  incrementOrderSize,
  decrementOrderSize,
  toggleOrderBookDetached,
  toggleDOMDetached,
  lockTerminalWorkstation,
}

enum KeyModifier {
  ctrl,
  alt,
  shift,
  meta,
}

class HotkeyBinding {
  final TradingAction action;
  final String primaryKey; // e.g. "Space", "F1", "B", "S"
  final List<KeyModifier> modifiers;
  final bool isGlobal; // true: works across entire OS; false: window-focus only
  final String description;

  const HotkeyBinding({
    required this.action,
    required this.primaryKey,
    this.modifiers = const [],
    this.isGlobal = false,
    required this.description,
  });

  Map<String, dynamic> toJson();
  factory HotkeyBinding.fromJson(Map<String, dynamic> json);
}

class HotkeyConfiguration {
  final String profileName;
  final List<HotkeyBinding> bindings;
  final bool conflictWarningDismissed;

  const HotkeyConfiguration({
    required this.profileName,
    required this.bindings,
    this.conflictWarningDismissed = false,
  });

  Map<String, dynamic> toJson();
  factory HotkeyConfiguration.fromJson(Map<String, dynamic> json);
}
```

### 4. Blockchain Settlement Event (`lib/features/desktop_terminal/domain/models/settlement_models.dart`)
```dart
class BlockchainSettlementEvent {
  final String tradeId;
  final String orderId;
  final String isin;
  final String transactionHash;
  final int blockNumber;
  final int requiredConfirmations;
  final int currentConfirmations;
  final bool isFinalized;
  final bool paymasterSponsored;
  final String paymasterAddress;
  final DateTime settlementTimestamp;

  const BlockchainSettlementEvent({
    required this.tradeId,
    required this.orderId,
    required this.isin,
    required this.transactionHash,
    required this.blockNumber,
    required this.requiredConfirmations,
    required this.currentConfirmations,
    required this.isFinalized,
    required this.paymasterSponsored,
    required this.paymasterAddress,
    required this.settlementTimestamp,
  });

  Map<String, dynamic> toJson();
  factory BlockchainSettlementEvent.fromJson(Map<String, dynamic> json);
}
```

### 5. Multi-Window Manager Interface (`lib/features/desktop_terminal/domain/interfaces/i_desktop_manager.dart`)
```dart
abstract class IDesktopMultiWindowManager {
  Future<void> initialize();
  Future<int> spawnWindow({
    required WindowType type,
    required String symbol,
    SymbolLinkGroup linkGroup = SymbolLinkGroup.none,
    WindowGeometry? initialGeometry,
  });
  Future<void> closeWindow(int windowId);
  Future<void> focusWindow(int windowId);
  Future<void> setAlwaysOnTop(int windowId, bool alwaysOnTop);
  Future<List<DesktopWindowState>> getActiveWindows();
  Future<void> saveWorkspaceLayout(String workspaceId, String name);
  Future<void> loadWorkspaceLayout(String workspaceId);
  Stream<IPCMessage> get ipcMessageStream;
  Future<void> sendIPCMessage(IPCMessage message);
}
```

## Security & Compliance Notes
- **Synchronized Inactivity Screen Lock:** A configurable inactivity timer (defaulting to 5 minutes of system-wide mouse and keyboard idle time) triggers a synchronized lock screen across all active desktop windows simultaneously. While locked, all order entry controls and sensitive financial metrics are blanked, and all global hotkeys are suppressed until the user re-authenticates via system biometric challenge (Windows Hello, macOS Touch ID) or master password.
- **Secure Token Isolation in Child Processes:** Worker windows operate under scoped, ephemeral session tokens issued by the Master Process via IPC. Sensitive root API keys, private signing seeds, and hardware vault handles remain isolated in the Master Process memory space and are never serialized or shared across process boundaries.
- **Encrypted Local Workspace Cache:** All persisted window geometries, monitor arrangements, custom hotkey bindings, and workspace templates stored in local databases are encrypted using AES-GCM-256 with keys stored in the OS credential vault (Windows Credential Manager, macOS Keychain, Linux Secret Service).
- **SEBI Institutional Audit Trail Logging:** Every keyboard shortcut invocation, click-to-trade entry, cancel action, and window detachment event is recorded with microsecond-precision hardware timestamps, source window identifiers, and operator session IDs in an immutable local audit log.
- **Screen Capture Protection:** On Windows and macOS workstations, institutional compliance mode can activate native window display affinity settings (`SetWindowDisplayAffinity` on Windows, `NSWindowSharingNone` on macOS) to prevent unauthorized screen scraping or third-party recording of proprietary trading screens.

## Acceptance Criteria
- [ ] Desktop application launches on Windows, macOS, and Linux, supporting seamless detachment of Order Book, DOM Ladder, and Execution Blotter into independent native OS windows.
- [ ] Multi-monitor window positioning accurately maps virtual desktop coordinates across displays with differing resolutions and DPI scaling factors (e.g. 4K at 150% + 1440p at 100%).
- [ ] DOM (Depth of Market) price ladder renders at a locked 60/120 FPS without UI jank or main-thread frame drops under tick rates exceeding 5,000 updates per second.
- [ ] Inter-process communication (IPC) propagates symbol changes, active link group switches, and theme updates across all child windows in less than 2 milliseconds.
- [ ] Global and window-specific hotkeys (including Spacebar panic cancel-all, F1-F12 symbol selection, and ladder centering) execute reliably with active conflict detection.
- [ ] Single-click order placement and cancellation on the DOM ladder triggers instantaneous order dispatch to the Order Gateway with visual fill confirmations.
- [ ] Blockchain Settlement Monitor displays live Hyperledger Besu block confirmation counts, on-chain transaction hashes, and gasless Paymaster sponsorship badges.
- [ ] Inactivity screen lock triggers synchronously across all open windows after the configured idle duration, shielding sensitive market depth and execution data.
- [ ] Custom multi-monitor workspace configurations can be saved, restored, and synchronized across application restarts with exact pixel geometry preservation.
- [ ] Full specification adheres strictly to the 12 mandatory sections with zero raw application code, zero en dashes, and zero em dashes (using standard ASCII hyphens exclusively).

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `501` (Flutter Project Scaffolding), Prompt `502` (App Architecture & State Management), Prompt `503` (Design System & Theming), Prompt `521` (Local Secure Storage).
- **Backend / Engine Dependencies:** Prompt `204` (Order Gateway & Execution Service), Prompt `207` (Market Data Service), Prompt `225` (FIX Protocol Gateway), Prompt `259` (Account Abstraction Bundler & Paymaster Service).
- **Enables:** Advanced institutional workstation capabilities extending Prompts `507` (Market Watchlist), `508` (Security Detail Screen), and `509` (Order Placement Flow) into professional multi-display desktop environments.
