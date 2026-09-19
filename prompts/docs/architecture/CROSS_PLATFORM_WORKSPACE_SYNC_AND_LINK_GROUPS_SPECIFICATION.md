# Cross-Platform Workspace Synchronization & Multi-Monitor Link Groups Specification

**Specification ID:** SPEC-ARCH-043-WORKSPACE-SYNC  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Core Client Architecture, Multi-Window IPC & Distributed State Synchronization  
**Target Services:** `services/workspace-sync-service`, `apps/growww_web`, `apps/growww_flutter` (macOS, Windows, Linux, iOS, Android)  
**Last Updated:** September 2026  

---

## 1. Executive Summary & Architectural Vision

Professional traders, proprietary trading desks, and institutional market operators require highly customized, multi-monitor display environments. A typical pro setup spans two to six physical monitors featuring specialized visual surfaces: high-density Level 2 and Level 3 order book ladders, multi-timeframe TradingView charts, depth-of-market (DOM) click-to-trade panels, real-time trade blotters, and execution telemetry consoles.

The Growww / NBSE Cross-Platform Workspace Synchronization & Multi-Monitor Link Group system solves two fundamental operational challenges across consumer web browsers and native desktop operating systems:
1. **Multi-Monitor Window Detachment & Zero-Latency IPC:** Allowing individual panels (charts, order books, ladders) to detach into independent native child windows across multiple physical monitors, maintaining sub-millisecond symbol synchronization across detached windows via platform-native Inter-Process Communication (IPC).
2. **Ubiquitous Cloud Workspace Synchronization ("Everywhere"):** Seamlessly synchronizing window geometries, monitor arrangements, DPI scale factors, panel splits, and ticker link groups across devices (Web Pro Terminal, macOS Apple Silicon workstations, Windows DirectX 12 rigs, Linux multi-head displays, and mobile devices) with distributed vector clocks for deterministic conflict resolution.

### Core Architectural Invariants:
- **Zero-Latency Local Inter-Window Sync:** When a trader changes a ticker symbol on Monitor 1, all linked panels across Monitors 2, 3, and 4 update within <=1 millisecond via zero-copy IPC without round-tripping to cloud servers.
- **DPI and Display-Agnostic Geometry Normalization:** Window positions and layout trees are saved in a normalized virtual coordinate space with per-monitor DPI scaling awareness, preventing window clipping or off-screen rendering when shifting between different monitor configurations (e.g. 4K 150% scaling to 1440p 100% scaling).
- **Four Immutable Color-Coded Link Groups:** Ticker locking is strictly categorized into 4 high-contrast link channels: Group Red (`#FF3B56`), Group Blue (`#00A3FF`), Group Green (`#00F0A0`), and Group Yellow (`#FFD600`).
- **Distributed Vector Clocks with Last-Write-Wins (LWW):** Multi-device synchronization resolves conflicting layout edits deterministically using monotonically increasing vector clocks combined with microsecond wall-clock tie-breakers.
- **Strict 0.00% Zero-Fee & 0 Gas Presentation:** Every trading terminal panel, order form, and execution blotter immutably incorporates the zero-fee presentation badge (`0.00% Maker / 0.00% Taker`) and the ERC-4337 Account Abstraction 0 Gas sponsorship badge (`0 Gas / Paymaster Sponsored`), guarded by layout constraints that prohibit obscuration.

---

## 2. End-to-End System Topology

```
+-------------------------------------------------------------------------------------------------------------------------+
|                                    CROSS-PLATFORM WORKSPACE & LINK GROUP TOPOLOGY                                       |
|                                                                                                                         |
|  +-------------------------------------------------------------------------------------------------------------------+  |
|  | PHYSICAL HARDWARE TIER (MULTI-MONITOR SETUP)                                                                      |  |
|  |                                                                                                                   |  |
|  |   +------------------------------------+  +------------------------------------+  +----------------------------+  |  |
|  |   | MONITOR 1 (Primary - 4K UHD @ 1.5x) |  | MONITOR 2 (Secondary - 1440p @ 1x) |  | MONITOR 3 (Vertical - 1080p)|  |
|  |   | [Watchlist Panel] (Group Red)       |  | [TradingView 4K] (Group Red)       |  | [DOM Ladder] (Group Red)   |  |
|  |   | [Order Entry Panel] (Group Blue)    |  | [Trade Blotter] (Group Blue)       |  | [L2 Book] (Group Yellow)   |  |
|  |   +------------------------------------+  +------------------------------------+  +----------------------------+  |  |
|  +-------------------------------------------------------------------------------------------------------------------+  |
|                                     |                                      |                               |            |
|  +-------------------------------------------------------------------------------------------------------------------+  |
|  | LOCAL INTER-PROCESS COMMUNICATION (IPC) TIER (SUB-MILLISECOND LINK GROUP TICKER SYNC)                            |  |
|  |                                                                                                                   |  |
|  |   * Web Browser:        BroadcastChannel API (`growww_link_groups`) + SharedWorker Fallback                       |  |
|  |   * Windows Thick App:  Named Pipes (`\\.\pipe\growww-workspace-sync-{uid}`)                                       |  |
|  |   * macOS / Linux:      Unix Domain Sockets (`/var/run/user/{uid}/growww-workspace.sock`)                         |  |
|  +-------------------------------------------------------------------------------------------------------------------+  |
|                                                         ^                                                               |
|                                                         | Local Encrypted Persistence (SQLite WAL / IndexedDB)          |
|                                                         v                                                               |
|  +-------------------------------------------------------------------------------------------------------------------+  |
|  | CLIENT ENGINE (STATE MACHINE, DOCKING ENGINE & VECTOR CLOCK MANAGER)                                              |  |
|  |                                                                                                                   |  |
|  |   - Golden Layout / Dockview React 19 Core (Web)                                                                   |  |
|  |   - Flutter Native Multi-Window Desktop Engine (Desktop)                                                           |  |
|  |   - Local Vector Clock Generator: V_clock = { device_id: counter }                                                 |  |
|  +-------------------------------------------------------------------------------------------------------------------+  |
|                                                         |                                                               |
|                                      Mutual TLS 1.3 / gRPC & Streaming WSS                                              |
|                                                         |                                                               |
|                                                         v                                                               |
|  +-------------------------------------------------------------------------------------------------------------------+  |
|  | CLOUD BACKEND SERVICE TIER (`services/workspace-sync-service`)                                                     |  |
|  |                                                                                                                   |  |
|  |   +-----------------------------------------------------------------------------------------------------------+   |  |
|  |   | gRPC API Endpoints (`/workspace.v1.WorkspaceSyncService/*`)                                               |   |  |
|  |   | - GetWorkspaceLayout / SaveWorkspaceLayout / StreamWorkspaceUpdates / ResolveConflict                     |   |  |
|  |   +-----------------------------------------------------------------------------------------------------------+   |  |
|  |   | WebSocket Gateway (`wss://ws.growww.in/v1/workspace/stream`)                                               |   |  |
|  |   | - Real-time JSON/Protobuf cross-device layout mutation broadcast                                           |   |  |
|  |   +-----------------------------------------------------------------------------------------------------------+   |  |
|  |   | Conflict Resolution Engine (Vector Clock Causality Matrix & LWW Arbiter)                                  |   |  |
|  |   +-----------------------------------------------------------------------------------------------------------+   |  |
|  +-------------------------------------------------------------------------------------------------------------------+  |
|                                     |                                      |                                            |
|                                     v                                      v                                            |
|                  +------------------------------------+  +------------------------------------+                         |
|                  | REDIS 7 CLUSTER (LAYOUT CACHE)     |  | POSTGRESQL 16 (AUTHORITATIVE STORE)|                         |
|                  | - Sub-millisecond snapshot cache   |  | - Relational layout records        |                         |
|                  | - Pub/Sub inter-node sync bus      |  | - JSONB normalized geometry trees  |                         |
|                  | - Distributed lease locks          |  | - Audit trail & version histories  |                         |
|                  +------------------------------------+  +------------------------------------+                         |
+-------------------------------------------------------------------------------------------------------------------------+
```

---

## 3. Central Workspace Synchronization Service Architecture

The Central Workspace Synchronization Service (`services/workspace-sync-service`) is an ultra-low-latency Go/Rust microservice responsible for maintaining, versioning, and distributing canonical workspace configurations across all connected clients for authenticated traders.

### 3.1 Service Architecture & Internal Components

```
+-------------------------------------------------------------------------------------------------------+
|                          CENTRAL WORKSPACE SYNC SERVICE INTERNALS                                     |
|                                                                                                       |
|   [ Ingress gRPC Traffic ]                   [ Ingress WebSocket Streams ]                            |
|             |                                              |                                          |
|             v                                              v                                          |
|   +-------------------+                          +-------------------+                                |
|   | gRPC Ingress      |                          | WSS Ingress       |                                |
|   | Controller        |                          | Session Hub       |                                |
|   +-------------------+                          +-------------------+                                |
|             |                                              |                                          |
|             +----------------------+-----------------------+                                          |
|                                    |                                                                  |
|                                    v                                                                  |
|   +-----------------------------------------------------------------------------------------------+   |
|   | AUTHENTICATION & SESSION CONTEXT VALIDATOR                                                    |   |
|   | - Validates SPIFFE/SPIRE mTLS X.509 SVIDs (Internal Services)                                 |   |
|   | - Validates ed25519 JWT Session Tokens (Client Terminals)                                     |   |
|   | - Extracts user_id, org_id, device_id, platform_type                                          |   |
|   +-----------------------------------------------------------------------------------------------+   |
|                                    |                                                                  |
|                                    v                                                                  |
|   +-----------------------------------------------------------------------------------------------+   |
|   | VECTOR CLOCK CONFLICT RESOLUTION & ARBITRATION PIPELINE                                        |   |
|   | - Evaluates incoming V_client vs V_server                                                     |   |
|   | - Detects concurrent diverged branches                                                        |   |
|   | - Applies deterministic 3-way layout merge or LWW timestamp arbitration                       |   |
|   +-----------------------------------------------------------------------------------------------+   |
|                                    |                                                                  |
|             +----------------------+----------------------+                                           |
|             |                                             |                                           |
|             v                                             v                                           |
|   +-----------------------------------+         +-----------------------------------+                 |
|   | REDIS WRITE-THROUGH CACHE         |         | POSTGRESQL 16 PERSISTENCE ADAPTER |                 |
|   | - Key: `ws:layout:{user_id}:{id}` |         | - Table: `workspace_profiles`     |                 |
|   | - TTL: 72 Hours Sliding Expiry    |         | - Transactional commit with WAL   |                 |
|   | - Cluster Pub/Sub for Invalidation|         | - Debezium CDC Outbox Eventing    |                 |
|   +-----------------------------------+         +-----------------------------------+                 |
|             |                                                                                         |
|             v                                                                                         |
|   +-----------------------------------------------------------------------------------------------+   |
|   | WEBSOCKET BROADCAST DISPATCHER                                                                |   |
|   | - Identifies other active sessions for user_id (excluding originating device_id)               |   |
|   | - Streams compressed Protobuf delta frames over established WebSockets                        |   |
|   +-----------------------------------------------------------------------------------------------+   |
+-------------------------------------------------------------------------------------------------------+
```

### 3.2 Protobuf & gRPC Service Definitions

The service exposes high-performance gRPC interfaces for synchronous fetching/saving and bidirectional streaming for real-time live synchronization:

```protobuf
syntax = "proto3";

package growww.workspace.v1;

option go_package = "github.com/growww/nbse/services/workspace-sync-service/v1;workspacev1";

import "google/protobuf/timestamp.proto";

// WorkspaceSyncService orchestrates real-time workspace layout synchronization
service WorkspaceSyncService {
  // Retrieves the current canonical workspace layout for a user
  rpc GetWorkspaceLayout(GetWorkspaceLayoutRequest) returns (GetWorkspaceLayoutResponse);

  // Atomically updates or saves a workspace layout profile
  rpc SaveWorkspaceLayout(SaveWorkspaceLayoutRequest) returns (SaveWorkspaceLayoutResponse);

  // Bidirectional streaming channel for real-time layout and link-group synchronization
  rpc StreamWorkspaceUpdates(stream WorkspaceSyncClientMessage) returns (stream WorkspaceSyncServerMessage);

  // List all available workspace layout presets for a user
  rpc ListWorkspaceProfiles(ListWorkspaceProfilesRequest) returns (ListWorkspaceProfilesResponse);

  // Delete a workspace layout profile
  rpc DeleteWorkspaceProfile(DeleteWorkspaceProfileRequest) returns (DeleteWorkspaceProfileResponse);
}

// Canonical color-coded link group identifiers
enum LinkGroupColor {
  LINK_GROUP_UNSPECIFIED = 0;
  LINK_GROUP_RED = 1;     // Scalping / Primary Pair (BTC/USDT)
  LINK_GROUP_BLUE = 2;    // Secondary Major Pair (ETH/USDT)
  LINK_GROUP_GREEN = 3;   // Commodity / Tokenized Asset (w-GOLD/eINR)
  LINK_GROUP_YELLOW = 4;  // Equity / Index Asset (NBSE-50/eINR)
}

// Platform type identifying the rendering client environment
enum ClientPlatform {
  CLIENT_PLATFORM_UNSPECIFIED = 0;
  CLIENT_PLATFORM_WEB_PRO = 1;
  CLIENT_PLATFORM_MACOS_NATIVE = 2;
  CLIENT_PLATFORM_WINDOWS_NATIVE = 3;
  CLIENT_PLATFORM_LINUX_NATIVE = 4;
  CLIENT_PLATFORM_IOS_MOBILE = 5;
  CLIENT_PLATFORM_ANDROID_MOBILE = 6;
}

// Vector Clock entry capturing causal version history
message VectorClock {
  // Map of device_id to monotonically increasing sequence number
  map<string, uint64> entries = 1;
  // Fallback microsecond UNIX epoch timestamp for deterministic LWW arbitration
  int64 wall_clock_epoch_micros = 2;
}

// Bounding box for normalized window geometry
message WindowBounds {
  int32 x = 1;
  int32 y = 2;
  int32 width = 3;
  int32 height = 4;
  bool is_maximized = 5;
  bool is_minimized = 6;
  bool is_fullscreen = 7;
}

// Display monitor characteristics
message DisplayDescriptor {
  string display_id = 1;          // OS-level unique display identifier
  string display_name = 2;        // Human-readable monitor name (e.g., "Pro Display XDR")
  WindowBounds bounds = 3;        // Resolution bounds in virtual screen coordinates
  WindowBounds work_area = 4;     // Usable area excluding OS taskbars/dock
  float dpi_scale_factor = 5;     // Scale factor: 1.0 = 100%, 1.5 = 150%, 2.0 = 200% (Retina)
  bool is_primary = 6;            // Indicates whether this is the OS primary monitor
  int32 refresh_rate_hz = 7;      // Monitor refresh rate (e.g., 60, 120, 144, 240)
}

// Individual trading panel docking state
message PanelState {
  string panel_id = 1;            // Unique panel instance identifier
  string panel_type = 2;          // e.g., "ORDER_BOOK", "CHART", "DOM_LADDER", "BLOTTER", "ORDER_ENTRY"
  string title = 3;               // Display title in tab header
  LinkGroupColor link_group = 4;  // Color-coded link group assignment
  string active_symbol = 5;       // Currently displayed symbol (e.g., "BTC-USDT")
  string json_state = 6;          // Component-specific serializable configuration JSON
  bool zero_fee_badge_visible = 7;// Strict compliance invariant: must evaluate to true
}

// Detached child window descriptor
message DetachedWindowDescriptor {
  string window_id = 1;           // Window identifier
  string target_display_id = 2;   // Associated physical monitor ID
  WindowBounds bounds = 3;        // Window coordinates and dimensions
  string layout_tree_json = 4;    // FlexLayout / Dockview serialized layout tree
  repeated PanelState panels = 5; // Panels docked inside this detached window
}

// Complete canonical workspace profile
message WorkspaceProfile {
  string profile_id = 1;          // UUID v4 identifier
  string user_id = 2;             // User account UUID
  string profile_name = 3;        // e.g., "Day Trader 3-Screen", "Scalper Dual DOM"
  bool is_active = 4;             // Is this the currently loaded workspace
  VectorClock vector_clock = 5;   // Vector clock tracking causal edits
  string root_layout_tree_json = 6;// Root main window layout tree
  repeated DetachedWindowDescriptor detached_windows = 7; // Detached pop-out windows
  repeated DisplayDescriptor display_topology = 8;        // Hardware topology snapshot
  ClientPlatform origin_platform = 9;                     // Platform where profile was created
  google.protobuf.Timestamp updated_at = 10;              // Server commit timestamp
  uint32 schema_version = 11;                             // Geometry schema version (current: 1)
}

message GetWorkspaceLayoutRequest {
  string user_id = 1;
  string profile_id = 2; // Optional: empty loads active profile
}

message GetWorkspaceLayoutResponse {
  WorkspaceProfile profile = 1;
}

message SaveWorkspaceLayoutRequest {
  WorkspaceProfile profile = 1;
  string originating_device_id = 2;
}

message SaveWorkspaceLayoutResponse {
  bool success = 1;
  WorkspaceProfile committed_profile = 2;
  bool conflict_detected = 3;
}

message ListWorkspaceProfilesRequest {
  string user_id = 1;
}

message ListWorkspaceProfilesResponse {
  repeated WorkspaceProfile profiles = 1;
}

message DeleteWorkspaceProfileRequest {
  string user_id = 1;
  string profile_id = 2;
}

message DeleteWorkspaceProfileResponse {
  bool success = 1;
}

// Bi-directional streaming messages
message WorkspaceSyncClientMessage {
  oneof payload {
    WorkspaceHeartbeat heartbeat = 1;
    LayoutDeltaUpdate layout_delta = 2;
    LinkGroupSymbolChange symbol_change = 3;
    DisplayTopologyChange topology_change = 4;
  }
}

message WorkspaceSyncServerMessage {
  oneof payload {
    WorkspaceHeartbeatAck heartbeat_ack = 1;
    WorkspaceProfile remote_layout_applied = 2;
    ConflictResolutionNotice conflict_notice = 3;
    LinkGroupSymbolBroadcast symbol_broadcast = 4;
  }
}

message WorkspaceHeartbeat {
  string device_id = 1;
  int64 timestamp_micros = 2;
}

message WorkspaceHeartbeatAck {
  int64 server_time_micros = 1;
}

message LayoutDeltaUpdate {
  string profile_id = 1;
  string device_id = 2;
  VectorClock client_clock = 3;
  string delta_patch_json = 4; // RFC 6902 JSON Patch
}

message LinkGroupSymbolChange {
  LinkGroupColor link_group = 1;
  string symbol = 2;
  string originating_panel_id = 3;
  int64 timestamp_micros = 4;
}

message LinkGroupSymbolBroadcast {
  LinkGroupColor link_group = 1;
  string symbol = 2;
  string originating_device_id = 3;
  int64 timestamp_micros = 4;
}

message DisplayTopologyChange {
  repeated DisplayDescriptor active_displays = 1;
}

message ConflictResolutionNotice {
  string profile_id = 1;
  string winning_device_id = 2;
  WorkspaceProfile resolved_profile = 3;
  string resolution_reason = 4;
}
```

---

## 4. Canonical Workspace Geometry Schema

Trading workstations utilize varied multi-monitor hardware topologies: high-DPI Apple Pro Display XDRs (3008x1692 logical @ 2.0x), ultra-wide Samsung Odyssey G9s (5120x1440 @ 1.0x), standard dual 1440p displays, and vertical 1080p monitors. The canonical workspace geometry schema defines a hardware-normalized, scale-invariant spatial representation.

### 4.1 Coordinate Space Normalization & Virtual Desktop Canvas

To prevent window placement anomalies (such as pop-out windows launching off-screen when opening a 3-monitor profile on a single laptop display), window coordinates are maintained in a Dual-Coordinate Mapping Model:
1. **Absolute Hardware Canvas Coordinates:** Real pixel bounds registered from OS display APIs (`EnumDisplayMonitors` on Win32, `NSScreen.screens` on macOS, `wl_output` / `XRRGetScreenResources` on Linux).
2. **Normalized Relative Viewport Coordinates:** Fractional percentages `(x_ratio, y_ratio, w_ratio, h_ratio)` relative to the target monitor work area. When restoring to a monitor with differing physical dimensions or aspect ratio, the layout manager defaults to fractional ratio projections and verifies minimum visual boundary clamps (`min_width = 320px`, `min_height = 240px`).

```
+---------------------------------------------------------------------------------------------------+
| VIRTUAL SCREEN SPACE NORMALIZATION                                                                |
|                                                                                                   |
|  DISPLAY 1 (Primary - 4K UHD)                     DISPLAY 2 (Secondary - 1440p)                   |
|  Physical: 3840 x 2160, Scale: 150%               Physical: 2560 x 1440, Scale: 100%             |
|  Logical Work Area: 2560 x 1400                   Logical Work Area: 2560 x 1440                  |
|  Origin: (x: 0, y: 0)                             Origin: (x: 2560, y: 0)                         |
|  +---------------------------------------------+  +--------------------------------------------+  |
|  | Window 1: Main Shell                        |  | Window 2: Detached Chart                   |  |
|  | Logical: (x: 0, y: 0, w: 2560, h: 1400)      |  | Logical: (x: 2560, y: 0, w: 1800, h: 1440) |  |
|  | Ratio:   (rx: 0, ry: 0, rw: 1.0, rh: 1.0)   |  | Ratio:   (rx: 0, ry: 0, rw: 0.70, rh: 1.0) |  |
|  +---------------------------------------------+  +--------------------------------------------+  |
+---------------------------------------------------------------------------------------------------+
```

### 4.2 DPI Scaling & Per-Monitor Display Metrics Matrix

Per-monitor DPI handling is critical for preventing blurriness, layout clipping, and misaligned mouse-hit tests during window movement between mixed-DPI displays:

| Platform | Native DPI Query API | DPI Change Notification Event | Scaling Handling Model |
| :--- | :--- | :--- | :--- |
| **Web Browser (`growww_web`)** | `window.devicePixelRatio` | `window.matchMedia('(resolution: ...)')` | WebGL canvas backbuffer resized via CSS pixel ratio multiplication; DOM scaled natively by browser. |
| **Windows Thick Client** | `GetDpiForMonitor` / Per-Monitor V2 | `WM_DPICHANGED` | Intercepts suggested `RECT` in `lParam`, updates DirectX swapchain scaling matrix, triggers Flutter Impeller re-rasterization. |
| **macOS Thick Client** | `NSScreen.backingScaleFactor` | `NSWindowDidChangeBackingPropertiesNotification` | CoreGraphics backing store re-allocated; Metal layer updates drawable size to match physical display pixels. |
| **Linux Thick Client** | `wl_surface_set_buffer_scale` / `Xft.dpi` | Wayland `preferred_scale` / X11 `RRScreenChangeNotify` | Adapts Vulkan surface dimensions; re-calculates GTK font metrics and panel constraints. |

### 4.3 Canonical JSON Schema Specification

The authoritative JSON serialization for an active workspace profile conforms to the following schema:

```json
{
  "$schema": "https://json-schema.growww.in/v1/workspace-profile.json",
  "profile_id": "8f3b6c20-7e1d-4b82-990a-6e210fa14d91",
  "user_id": "usr_998120485932",
  "profile_name": "Institutional Scalper 3-Screen",
  "is_active": true,
  "schema_version": 1,
  "vector_clock": {
    "entries": {
      "desk_win_workstation_alpha": 412,
      "mbp_m3_max_mobile": 89
    },
    "wall_clock_epoch_micros": 1792482463000000
  },
  "display_topology": [
    {
      "display_id": "disp_dell_u3223qe_primary",
      "display_name": "Dell UltraSharp 32 4K",
      "is_primary": true,
      "dpi_scale_factor": 1.5,
      "refresh_rate_hz": 60,
      "bounds": { "x": 0, "y": 0, "width": 3840, "height": 2160, "is_maximized": false, "is_minimized": false, "is_fullscreen": false },
      "work_area": { "x": 0, "y": 0, "width": 3840, "height": 2100, "is_maximized": false, "is_minimized": false, "is_fullscreen": false }
    },
    {
      "display_id": "disp_asus_rog_pg279qm_secondary",
      "display_name": "ASUS ROG Swift 27 240Hz",
      "is_primary": false,
      "dpi_scale_factor": 1.0,
      "refresh_rate_hz": 240,
      "bounds": { "x": 3840, "y": 0, "width": 2560, "height": 1440, "is_maximized": false, "is_minimized": false, "is_fullscreen": false },
      "work_area": { "x": 3840, "y": 0, "width": 2560, "height": 1440, "is_maximized": false, "is_minimized": false, "is_fullscreen": false }
    }
  ],
  "root_layout": {
    "window_id": "win_root_main",
    "display_id": "disp_dell_u3223qe_primary",
    "bounds": { "x": 0, "y": 0, "width": 2560, "height": 1400, "is_maximized": true, "is_minimized": false, "is_fullscreen": false },
    "dock_tree": {
      "type": "row",
      "weight": 100,
      "children": [
        {
          "type": "column",
          "weight": 25,
          "children": [
            {
              "type": "panel",
              "panel_id": "panel_watchlist_01",
              "panel_type": "WATCHLIST",
              "title": "Scalp Watchlist",
              "link_group": "LINK_GROUP_RED",
              "active_symbol": "BTC-USDT",
              "zero_fee_badge_visible": true,
              "config": { "show_sparkline": true, "sort_column": "change_24h" }
            },
            {
              "type": "panel",
              "panel_id": "panel_order_entry_01",
              "panel_type": "ORDER_ENTRY",
              "title": "Instant Order Execution",
              "link_group": "LINK_GROUP_RED",
              "active_symbol": "BTC-USDT",
              "zero_fee_badge_visible": true,
              "config": { "default_order_type": "LIMIT", "leverage": "10x", "paymaster_gas_sponsorship": true }
            }
          ]
        },
        {
          "type": "column",
          "weight": 75,
          "children": [
            {
              "type": "panel",
              "panel_id": "panel_chart_primary",
              "panel_type": "TRADINGVIEW_CHART",
              "title": "BTC-USDT 1m Pro Chart",
              "link_group": "LINK_GROUP_RED",
              "active_symbol": "BTC-USDT",
              "zero_fee_badge_visible": true,
              "config": { "timeframe": "1m", "chart_type": "candlestick", "indicators": ["EMA_9", "EMA_21", "VWAP"] }
            },
            {
              "type": "panel",
              "panel_id": "panel_blotter_01",
              "panel_type": "TRADE_BLOTTER",
              "title": "Real-Time Trade & Execution Blotter",
              "link_group": "LINK_GROUP_RED",
              "active_symbol": "BTC-USDT",
              "zero_fee_badge_visible": true,
              "config": { "filter": "ALL_WORKING", "auto_scroll": true }
            }
          ]
        }
      ]
    }
  },
  "detached_windows": [
    {
      "window_id": "win_detached_dom_01",
      "target_display_id": "disp_asus_rog_pg279qm_secondary",
      "bounds": { "x": 3840, "y": 0, "width": 1280, "height": 1440, "is_maximized": false, "is_minimized": false, "is_fullscreen": false },
      "dock_tree": {
        "type": "column",
        "weight": 100,
        "children": [
          {
            "type": "panel",
            "panel_id": "panel_dom_ladder_01",
            "panel_type": "DOM_LADDER",
            "title": "DOM Price Ladder - 50 Depth",
            "link_group": "LINK_GROUP_RED",
            "active_symbol": "BTC-USDT",
            "zero_fee_badge_visible": true,
            "config": { "depth_levels": 50, "show_cumulative_volume": true, "single_click_trading": true }
          }
        ]
      }
    },
    {
      "window_id": "win_detached_secondary_02",
      "target_display_id": "disp_asus_rog_pg279qm_secondary",
      "bounds": { "x": 5120, "y": 0, "width": 1280, "height": 1440, "is_maximized": false, "is_minimized": false, "is_fullscreen": false },
      "dock_tree": {
        "type": "column",
        "weight": 100,
        "children": [
          {
            "type": "panel",
            "panel_id": "panel_l2_book_blue",
            "panel_type": "ORDER_BOOK",
            "title": "L2 Book - ETH Major",
            "link_group": "LINK_GROUP_BLUE",
            "active_symbol": "ETH-USDT",
            "zero_fee_badge_visible": true,
            "config": { "aggregation_tick": "0.01", "visual_imbalance": true }
          }
        ]
      }
    }
  ]
}
```

---

## 5. Four Color-Coded Link Groups

To enable institutional-grade workflow ergonomics, all panels within any main window or detached pop-out window can be linked via 4 Color-Coded Link Groups. Changing a ticker symbol in any component instantly locks and updates all other panels sharing that Link Group across all physical displays.

### 5.1 Link Group Taxonomy & Color System

```
+----------------------------------------------------------------------------------------------------+
| COLOR-CODED LINK GROUP MATRIX                                                                      |
|                                                                                                    |
|  [ Group Red ]     Hex: #FF3B56  |  Primary Scalping Pair (BTC/USDT, High-Vol Derivatives)         |
|  [ Group Blue ]    Hex: #00A3FF  |  Secondary Major Pair (ETH/USDT, Layer 1 Assets)                |
|  [ Group Green ]   Hex: #00F0A0  |  Commodity & Sovereign Vault Tokens (w-GOLD/eINR, Silver)       |
|  [ Group Yellow ]  Hex: #FFD600  |  Equities, Structured RWAs & Index Baskets (NBSE-50/eINR)       |
+----------------------------------------------------------------------------------------------------+
```

1. **Group Red (`#FF3B56`):**
   - Designated default link channel for active scalp and momentum instruments (e.g. BTC-USDT Perpetual Futures or Spot).
   - High-contrast crimson badge displayed in panel tab corner and header border.
2. **Group Blue (`#00A3FF`):**
   - Designated channel for secondary institutional major pairs (e.g. ETH-USDT, SOL-USDT).
   - Vivid cyan-blue badge indicator.
3. **Group Green (`#00F0A0`):**
   - Designated channel for tokenized real-world assets, sovereign gold tokens, and commodities (e.g. w-GOLD-eINR, SILVER-eINR).
   - Neon emerald badge indicator.
4. **Group Yellow (`#FFD600`):**
   - Designated channel for equity index tokens, corporate token debt, and basket instruments (e.g. NBSE-50-eINR).
   - High-visibility amber-yellow badge indicator.
5. **Group None (Unlinked / Independent):**
   - Panel retains its symbol independently and ignores link group broadcasts. Allows maintaining a static index chart or macro overview while active groups fluctuate.

### 5.2 Cross-Window & Cross-Device Synchronization Pipeline

```
+-------------------------------------------------------------------------------------------------------+
|                               LINK GROUP BROADCAST & DISPATCH FLOW                                    |
|                                                                                                       |
|   +-----------------------------------------------------------------------------------------------+   |
|   | TRADER ACTION: User clicks "ETH-USDT" in Watchlist Panel (Assigned to LINK_GROUP_BLUE)        |   |
|   +-----------------------------------------------------------------------------------------------+   |
|                                                  |                                                    |
|                                                  v                                                    |
|   +-----------------------------------------------------------------------------------------------+   |
|   | LOCAL EVENT DISPATCHER                                                                        |   |
|   | 1. Constructs LinkGroupSymbolChange event:                                                     |   |
|   |    { group: "LINK_GROUP_BLUE", symbol: "ETH-USDT", source_panel: "watchlist_01", time: ... }    |   |
|   | 2. Broadcasts locally over OS-level IPC:                                                      |   |
|   |    - Web: BroadcastChannel("growww_link_groups").postMessage(...)                             |   |
|   |    - Windows: WriteFile to Named Pipe `\\.\pipe\growww-workspace-sync-{uid}`                  |   |
|   |    - macOS/Linux: write() to Unix Domain Socket `/var/run/user/...`                           |   |
|   +-----------------------------------------------------------------------------------------------+   |
|                        |                                                          |                   |
|       [ Local IPC Dispatch <= 1ms ]                                 [ Cloud Sync Dispatch ]           |
|                        |                                                          |                   |
|                        v                                                          v                   |
|   +-------------------------------------------+      +--------------------------------------------+   |
|   | LOCAL WINDOWS ON ALL PHYSICAL MONITORS    |      | STREAMING WEBSOCKET TO BACKEND             |   |
|   | - Detached Window 1 (Chart Blue):         |      | - Transmits LinkGroupSymbolChange payload  |   |
|   |   Switches chart symbol to "ETH-USDT"     |      | - Backend publishes to Redis layout cache  |   |
|   | - Detached Window 2 (Order Book Blue):    |      | - Broadcasts to user's other active        |   |
|   |   Switches L2 ladder to "ETH-USDT"        |      |   devices (e.g. Mobile iPad / Laptop)      |   |
|   | - Execution Blotter (Blue):               |      +--------------------------------------------+   |
|   |   Filters working orders for "ETH-USDT"   |                                                       |
|   +-------------------------------------------+                                                       |
+-------------------------------------------------------------------------------------------------------+
```

### 5.3 Local IPC Event Payload Specification

Every link group symbol broadcast follows an explicit binary/JSON wire layout:

```json
{
  "event_type": "LINK_GROUP_SYMBOL_TRANSITION",
  "version": 1,
  "timestamp_micros": 1792482463124500,
  "originating_device_id": "desk_win_workstation_alpha",
  "originating_window_id": "win_root_main",
  "originating_panel_id": "panel_watchlist_01",
  "link_group": "LINK_GROUP_BLUE",
  "previous_symbol": "SOL-USDT",
  "target_symbol": "ETH-USDT",
  "market_category": "PERPETUAL_FUTURES",
  "propagation_scope": "LOCAL_AND_CLOUD"
}
```

---

## 6. Multi-Monitor Window Detachment & IPC Communication Protocols

Detaching visual panels into separate OS windows enables traders to distribute charts, ladders, and blotters freely across multiple physical monitors. Because market state moves in microseconds, inter-window communication must not depend on cloud network roundtrips.

### 6.1 Platform-Specific IPC Mechanisms

```
+----------------------------------------------------------------------------------------------------+
| MULTI-PLATFORM IPC IMPLEMENTATION MATRIX                                                           |
|                                                                                                    |
|  Platform       IPC Mechanism           Latency Target   Zero-Copy Buffer   Fault Recovery         |
|  :---           :---                    :---             :---               :---                   |
|  Web Pro        BroadcastChannel API    < 0.8 ms         SharedArrayBuffer  SharedWorker Fallback  |
|  Windows        Win32 Named Pipes       < 0.15 ms        Overlapped I/O     Pipe Auto-Reconnect    |
|  macOS          Unix Domain Sockets     < 0.12 ms        kqueue / POSIX     UDS Rebind & Handshake |
|  Linux          Unix Domain Sockets     < 0.10 ms        epoll / Non-block  Socket Reselect        |
+----------------------------------------------------------------------------------------------------+
```

#### 6.1.1 Web Browser Pop-Out Windows (BroadcastChannel & SharedWorker)
- **Primary Transport:** Modern Chromium, WebKit, and Gecko browsers leverage the `BroadcastChannel` API bound to the channel key `growww_workspace_ipc_v1`.
- **Secondary Shared State Transport:** For high-throughput Level 2 order book streaming (up to 50,000 updates/sec), windows connect to a singleton `SharedWorker`. The `SharedWorker` maintains a shared `SharedArrayBuffer` containing the latest market depth ring buffer and dispatches UI sync pulses via `Atomics.notify()` and `postMessage` with transferable ArrayBuffers.
- **Window Management:** Detached windows open via `window.open('/popout?panelId=...', 'win_id', 'width=1280,height=800,left=1920,top=0')`. The parent window maintains a registry of child window references and monitors window closure via `unload` and `visibilitychange` listeners.

#### 6.1.2 Windows Native Thick Client (Win32 Named Pipes)
- **Pipe Address:** `\\.\pipe\growww-workspace-sync-{user_id}`
- **Operational Mode:** `PIPE_TYPE_MESSAGE | PIPE_READMODE_MESSAGE | PIPE_WAIT` with asynchronous overlapped I/O (`FILE_FLAG_OVERLAPPED`).
- **Architecture:** The primary window hosts the Named Pipe Server instance. When child windows spawn across secondary and tertiary monitors, they connect as pipe clients (`CreateFileW`).
- **Framing:** 4-byte little-endian message length prefix followed by Protobuf/JSON serialized frames.
- **Security:** Strict Windows Access Control Lists (ACLs) permit pipe access only to the current Windows Security Identifier (SID), preventing cross-user eavesdropping on multi-tenant workstations.

#### 6.1.3 macOS & Linux Thick Clients (Unix Domain Sockets)
- **Socket Path:**
  - macOS: `$TMPDIR/growww-workspace-sync-{user_id}.sock` (protected by POSIX file permissions `0600`).
  - Linux: `/var/run/user/{uid}/growww-workspace-sync.sock` (systemd XDG_RUNTIME_DIR compliant).
- **Architecture:** Primary window creates non-blocking listening socket. Event polling utilizes `kqueue` (macOS) or `epoll` (Linux) to monitor connected detached windows.
- **Backpressure & Framing:** Length-delimited framing with a 32-bit unsigned big-endian integer prefix. If a detached window's socket buffer reaches capacity (e.g. during heavy GPU render stutter), frames are conflated using a single-slot latest-value cache for link group transitions.

### 6.2 Window Detachment & Re-Attachment Lifecycle State Machine

```
+-------------------------------------------------------------------------------------------------------+
|                               WINDOW DETACHMENT & DOCKING LIFECYCLE                                   |
|                                                                                                       |
|                     [ Panel Docked in Main Shell ]                                                    |
|                                   |                                                                   |
|                   User Drags Tab Out or Clicks "Detach"                                               |
|                                   v                                                                   |
|                     [ Spawn Pop-Out Window ]                                                          |
|                     - Query current display metrics & work area                                       |
|                     - Set initial bounds clamped to target display                                    |
|                                   |                                                                   |
|                                   v                                                                   |
|                     [ Establish Local IPC Handshake ]                                                 |
|                     - Web: BroadcastChannel ping/pong                                                 |
|                     - Win32: ConnectNamedPipe()                                                       |
|                     - macOS/Linux: connect() UDS                                                      |
|                                   |                                                                   |
|                     +-------------+-------------+                                                     |
|                     |                           |                                                     |
|          [ Handshake Succeeded ]     [ Handshake Timed Out (500ms) ]                                  |
|                     |                           |                                                     |
|                     v                           v                                                     |
|         [ Active Detached State ]     [ Graceful Degradation ]                                        |
|         - Sub-millisecond sync        - Fallback to LocalStorage / File polling                       |
|         - Zero-fee badge rendered     - Display degraded IPC toast notification                       |
|                     |                                                                                 |
|                     +---------------------------+                                                     |
|                                                 |                                                     |
|                               User Drags Back or Closes Window                                        |
|                                                 v                                                     |
|                                  [ Teardown & Re-Docking ]                                            |
|                                  - Flush final panel state to parent                                  |
|                                  - Disconnect IPC handles                                             |
|                                  - Re-dock panel into root layout tree                                |
|                                  - Re-render unified layout within 16ms                               |
+-------------------------------------------------------------------------------------------------------+
```

---

## 7. Cross-Platform State Persistence Architecture

Workspace state is persisted across three distinct tiers to ensure high availability, sub-millisecond local retrieval, and offline resilience.

```
+-------------------------------------------------------------------------------------------------------+
|                                   THREE-TIER STATE PERSISTENCE ARCHITECTURE                           |
|                                                                                                       |
|    +---------------------------------------------------------------------------------------------+    |
|    | TIER 1: AUTHORITATIVE RELATIONAL STORE (PostgreSQL 16)                                      |    |
|    | - Authoritative historical record of user workspace profiles and layout trees               |    |
|    | - ACID transactional integrity, JSONB indexing, foreign keys, user tenant isolation          |    |
|    +---------------------------------------------------------------------------------------------+    |
|                                                  ^                                                    |
|                               Write-Through / Asynchronous Commit                                     |
|                                                  v                                                    |
|    +---------------------------------------------------------------------------------------------+    |
|    | TIER 2: HIGH-THROUGHPUT CLOUD CACHE & PUB/SUB (Redis 7 Cluster)                             |    |
|    | - Key: `ws:layout:{user_id}:{profile_id}`                                                   |    |
|    | - In-memory layout snapshot serving gRPC `GetWorkspaceLayout` in < 2ms                      |    |
|    | - Cluster Pub/Sub channel `ws:events:{user_id}` for instant cross-session push               |    |
|    +---------------------------------------------------------------------------------------------+    |
|                                                  ^                                                    |
|                               Mutual TLS 1.3 Streaming WebSocket Sync                                 |
|                                                  v                                                    |
|    +---------------------------------------------------------------------------------------------+    |
|    | TIER 3: LOCAL ENCRYPTED CLIENT PERSISTENCE (Offline-First)                                  |    |
|    | - Desktop / Mobile: SQLite 3.45+ with SQLCipher (AES-256-GCM) in WAL mode                   |    |
|    | - Web Pro: IndexedDB via `idb` with Web Cryptography API AES-GCM encryption                  |    |
|    | - Instant zero-latency cold-start load (<15ms) prior to cloud network handshake             |    |
|    +---------------------------------------------------------------------------------------------+    |
+-------------------------------------------------------------------------------------------------------+
```

### 7.1 PostgreSQL 16 Relational Schema & DDL

The authoritative relational database stores versioned profiles, display topologies, and JSONB layout trees:

```sql
-- PostgreSQL 16 Authoritative Workspace Storage Schema
CREATE SCHEMA IF NOT EXISTS workspace_mgmt;

CREATE TABLE workspace_mgmt.workspace_profiles (
    profile_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(64) NOT NULL,
    profile_name VARCHAR(128) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    schema_version INTEGER NOT NULL DEFAULT 1,
    vector_clock JSONB NOT NULL DEFAULT '{}'::jsonb,
    wall_clock_epoch_micros BIGINT NOT NULL,
    display_topology JSONB NOT NULL DEFAULT '[]'::jsonb,
    root_layout JSONB NOT NULL DEFAULT '{}'::jsonb,
    detached_windows JSONB NOT NULL DEFAULT '[]'::jsonb,
    origin_platform VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_user_profile_name UNIQUE (user_id, profile_name)
);

-- Indexing for instantaneous profile queries by user
CREATE INDEX idx_workspace_profiles_user_active 
ON workspace_mgmt.workspace_profiles (user_id, is_active);

CREATE INDEX idx_workspace_profiles_updated 
ON workspace_mgmt.workspace_profiles (user_id, updated_at DESC);

-- Historical layout audit log for rollback and disaster recovery
CREATE TABLE workspace_mgmt.workspace_audit_log (
    audit_id BIGSERIAL PRIMARY KEY,
    profile_id UUID NOT NULL REFERENCES workspace_mgmt.workspace_profiles(profile_id) ON DELETE CASCADE,
    user_id VARCHAR(64) NOT NULL,
    device_id VARCHAR(64) NOT NULL,
    action VARCHAR(32) NOT NULL, -- 'CREATED', 'UPDATED', 'MERGED', 'DELETED'
    delta_patch JSONB,
    snapshot JSONB NOT NULL,
    committed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workspace_audit_profile 
ON workspace_mgmt.workspace_audit_log (profile_id, committed_at DESC);
```

### 7.2 Redis 7 Layout Cache & Invalidation Architecture

- **Snapshot Key:** `ws:profile:{user_id}:{profile_id}`
  - Format: Compressed Protobuf binary payload.
  - TTL: 72 hours sliding expiration, refreshed on access.
- **Active Profile Pointer:** `ws:active:{user_id}`
  - Points to the `profile_id` currently active for the user.
- **Session Distributed Lease Lock:** `ws:lock:{user_id}:{profile_id}`
  - Redlock algorithm with a 5-second TTL used during multi-device vector clock merge transactions to prevent split-brain writes.

### 7.3 Local Offline Client Persistence

1. **Desktop Thick Clients (macOS, Windows, Linux):**
   - SQLite 3 in WAL mode (`PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;`).
   - Encrypted with AES-256-GCM via SQLCipher. Key derivation uses PBKDF2 with 256,000 iterations rooted in OS secure credential stores (Apple Keychain, Windows Credential Manager, Linux Secret Service).
   - Instant local boot: Terminal paints complete multi-monitor layout within 15 milliseconds of launch, before establishing network sockets.
2. **Web Pro Terminal (`growww_web`):**
   - HTML5 IndexedDB accessed via `idb` library under database `growww_workspace_store`.
   - Layout cached locally with an AES-GCM symmetric key stored in non-exportable CryptoKey format within the browser.

---

## 8. Distributed Conflict Resolution & Vector Clock Engine

Active traders frequently maintain simultaneous active sessions: an office multi-monitor Windows trading tower, a portable MacBook Pro, and a mobile phone. When layouts are edited simultaneously on different devices while intermittently disconnected, deterministic conflict resolution is mandatory.

### 8.1 Vector Clock Formalism

Every device generates a unique device identifier: `device_id` (e.g. `desk_win_01`, `mbp_mac_02`). Each layout mutation increments the device's counter in the vector clock:

$$V = \{ d_1: c_1, d_2: c_2, \dots, d_n: c_n \}$$

Given two vector clocks $V_A$ and $V_B$:
1. **$V_A$ dominates $V_B$ ($V_A > V_B$):** If for all devices $d$, $V_A[d] \ge V_B[d]$ and there exists at least one $d$ where $V_A[d] > V_B[d]$. $V_A$ causally succeeds $V_B$; $V_A$ is accepted without conflict.
2. **$V_A$ equals $V_B$ ($V_A = V_B$):** Identical state; update is a no-op.
3. **$V_A$ and $V_B$ are concurrent ($V_A \parallel V_B$):** There exists some $d_i$ where $V_A[d_i] > V_B[d_i]$ and some $d_j$ where $V_B[d_j] > V_A[d_j]$. This indicates simultaneous uncoordinated edits.

```
+---------------------------------------------------------------------------------------------------+
| VECTOR CLOCK CAUSALITY MATRIX                                                                     |
|                                                                                                   |
|           Device A (Windows Tower)                   Device B (MacBook Pro)                       |
|           ------------------------                   ----------------------                       |
|           Initial State:                                                                          |
|           V_0 = { A: 10, B: 5 }                                                                   |
|                      |                                          |                                 |
|            Edit: Moves DOM on Mon 2                   Edit: Resizes Chart on Mon 1                |
|            V_A = { A: 11, B: 5 }                      V_B = { A: 10, B: 6 }                       |
|                      |                                          |                                 |
|                      \--------------------+---------------------/                                 |
|                                           |                                                       |
|                                           v                                                       |
|                              [ Conflict Detected: V_A || V_B ]                                    |
|                                           |                                                       |
|                                           v                                                       |
|                        +-------------------------------------+                                    |
|                        | CONFLICT RESOLUTION ARBITER         |                                    |
|                        | 1. Evaluates 3-Way Structural Merge |                                    |
|                        | 2. Fallback: Microsecond LWW Clock  |                                    |
|                        | 3. Resolves & Emits V_Merged        |                                    |
|                        |    V_Merged = { A: 11, B: 6 }       |                                    |
|                        +-------------------------------------+                                    |
+---------------------------------------------------------------------------------------------------+
```

### 8.2 Conflict Resolution Pipeline & Three-Way Merging

When concurrency ($V_A \parallel V_B$) is detected, `services/workspace-sync-service` executes the following deterministic reconciliation sequence:

1. **Non-Overlapping Component Merge (Structural Merge):**
   - If Device A modified panels in Window 1, and Device B modified panels in Window 2, both modifications are preserved.
   - If Device A updated Link Group Red's active symbol, and Device B updated Link Group Blue's active symbol, both symbol assignments merge cleanly into the composite profile.
2. **Overlapping Collision Arbitration (Last-Write-Wins Fallback):**
   - If both devices modified the same panel or window coordinates, the system falls back to the deterministic Last-Write-Wins (LWW) rule using `wall_clock_epoch_micros` (synchronized via NTP / Chrony with <= 500 microsecond drift).
   - In the event of an exact microsecond tie, the lexicographically higher `device_id` string wins deterministically.
3. **Consensus Vector Clock Advancement:**
   - The merged profile receives a merged vector clock:
     
     $$V_{\text{merged}}[d] = \max(V_A[d], V_B[d]) \quad \forall d$$
     
   - The resolved profile is immediately persisted to PostgreSQL, updated in Redis, and pushed to all active sessions with a `ConflictResolutionNotice`.

---

## 9. Regulatory Compliance, Zero-Fee Presentation & Gas Sponsorship Badge

The Growww / NBSE trading terminal is architected on a zero-fee paradigm. To ensure full compliance with regulatory transparency and exchange fair-trading directives, layout engines enforce strict visual invariants.

### 9.1 Zero-Fee Presentation Invariant

All trading panels, including Instant Order Entry, DOM Price Ladders, Quick Execution Bars, and Confirmation Overlays, must visibly display the trading fee rate as exactly `0.00%`:

```
+---------------------------------------------------------------------------------------------------+
| ZERO-FEE TRADING BADGE UI COMPONENT                                                               |
|                                                                                                   |
|   +-------------------------------------------------------------------------------------------+   |
|   |  ORDER CONFIRMATION: BUY 0.25 BTC @ 64,250.00 USDT                                        |   |
|   |                                                                                           |   |
|   |  Order Value:          16,062.50 USDT                                                     |   |
|   |  Exchange Fee Rate:    0.00% (Genesis Zero-Fee Tier)                        [ZERO-FEE]    |   |
|   |  Exchange Fee Total:   0.00 USDT                                                          |   |
|   |  Statutory TDS (1%):   0.00 USDT (Unencumbered Liquidity Route)                           |   |
|   |  Paymaster Gas Fee:    0.00 USDT (ERC-4337 Sponsored)                      [0 GAS FREE]  |   |
|   |  ---------------------------------------------------------------------------------------  |   |
|   |  Total Ingress Debit:  16,062.50 USDT                                                     |   |
|   +-------------------------------------------------------------------------------------------+   |
+---------------------------------------------------------------------------------------------------+
```

### 9.2 ERC-4337 Account Abstraction 0 Gas Sponsorship Badge

All on-chain settlement operations (including Besu ledger settlements, Taproot custody transfers, and smart contract trades) are fully sponsored by exchange paymasters. The user interface mandates an explicit badge:

- **Badge Text:** `0 Gas (ERC-4337 Sponsored)`
- **Badge Styling:**
  - Background: `#00F0A01A` (10% opacity neon emerald)
  - Border: `1px solid #00F0A0`
  - Text Color: `#00F0A0`
  - Icon: Fuel pump glyph with zero-slash overlay

### 9.3 Non-Obscuration Layout Constraint Verification

The layout rendering engine enforces an immutable layout constraint:
1. When resizing or splitting panels, if a panel's width drops below `240px` or height below `160px`, the zero-fee and gas badges do not clip or disappear. Instead, the UI switches to a high-density micro-badge: `[0.00% / 0G]`.
2. Any client-side CSS or native Flutter transform that attempts to set `display: none`, `opacity: 0`, or `visibility: hidden` on the zero-fee compliance container triggers an automated layout reset event and reports a compliance layout anomaly to client telemetry.

---

## 10. Performance Benchmarks & Acceptance Criteria Matrix

To certify that the Cross-Platform Workspace Synchronization & Multi-Monitor Link Group engine meets institutional-grade performance standards, all client platforms and backend services must fulfill the following acceptance criteria:

| Test ID | Metric / Operation | Target Benchmark | Verification Procedure |
| :--- | :--- | :--- | :--- |
| **TC-SYNC-01** | Local IPC Link Group Symbol Change | **< 1.0 ms** (P99) | Measure elapsed time from click on Watchlist in Window 1 to symbol load event in Detached Window 2 across 3 physical displays. |
| **TC-SYNC-02** | Cloud Workspace Profile Fetch | **< 15.0 ms** (P95) | gRPC `GetWorkspaceLayout` benchmark against Redis cache under 10,000 req/sec load. |
| **TC-SYNC-03** | End-to-End Cross-Device Layout Sync | **< 150.0 ms** (P95) | Layout mutation on Windows Workstation streaming over WebSocket, applying to macOS client over 100ms simulated WAN latency. |
| **TC-SYNC-04** | Cold-Start Offline Layout Painting | **< 20.0 ms** (P99) | App launch to complete multi-monitor layout paint from encrypted SQLite WAL / IndexedDB cache prior to network connection. |
| **TC-SYNC-05** | High-DPI Multi-Monitor Window Drag | **0 Dropped Frames** | Drag detached window between 4K 150% scaling display and 1440p 100% scaling display; verify 0 visual tearing or coordinate miscalculations. |
| **TC-SYNC-06** | Vector Clock Concurrent Resolution | **100% Deterministic** | Induce simultaneous conflicting edits on 2 disconnected clients; verify both converge to identical state upon network reconnection. |
| **TC-SYNC-07** | Zero-Fee Badge Invariant Check | **100% Compliance** | Automated layout test resizes 20 panel permutations; verify zero-fee and 0 gas badges remain visible with CLS = 0. |

---

## 11. Appendix: Edge Cases & Failure Recovery Modes

### 11.1 Physical Monitor Disconnection / Reconnection
- **Scenario:** A laptop connected to dual external 4K monitors is unplugged from its Thunderbolt dock.
- **Handling:**
  1. The client intercepts OS display change notifications (`WM_DISPLAYCHANGE` on Windows, `NSApplicationDidChangeScreenParametersNotification` on macOS, `onchange` on browser `screen.orientation`).
  2. All detached child windows residing on disconnected monitors are instantly collapsed into floating docked tabs in the primary screen's layout tree.
  3. Window geometries are saved into a fallback cache. When the Thunderbolt dock is re-attached, the client detects matching `display_id` descriptors and automatically restores the detached multi-monitor configuration.

### 11.2 Split-Brain Network Partition During Trade Execution
- **Scenario:** Local Wi-Fi disconnects while trader is moving panels and submitting orders.
- **Handling:**
  1. Link Group IPC continues functioning with zero degradation across physical monitors via local Named Pipes / Sockets.
  2. Workspace layout edits are queued in local SQLite / IndexedDB with incremented vector clock counters.
  3. Order entry panels enter "Offline Warning Mode", displaying queued status while maintaining strict 0.00% fee invariants.
  4. Upon WAN reconnection, buffered layout vector clocks synchronistically reconcile with the central service via gRPC stream.
