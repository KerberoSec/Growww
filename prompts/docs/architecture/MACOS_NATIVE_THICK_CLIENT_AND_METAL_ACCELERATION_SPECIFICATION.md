# macOS Native Thick Client & Metal Acceleration Architectural Specification

**Specification ID:** SPEC-ARCH-043-MACOS-METAL  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Institutional Desktop Architecture, High-Performance Native Systems  
**Target Platform:** macOS 12.0 Monterey, macOS 13.0 Ventura, macOS 14.0 Sonoma, macOS 15.0 Sequoia and later  
**Target Hardware:** Apple Silicon (M1/M2/M3/M4, Pro, Max, Ultra) & Universal Intel x86_64  
**Display Target:** Apple ProMotion (120Hz locked), Apple Studio Display (5K P3), Apple Pro Display XDR (6K EDR/HDR)  
**Last Updated:** September 2026  

---

## 1. Executive Summary & Core Platform Directives

The Growww / NBSE macOS Native Thick Client represents the pinnacle institutional trading workstation for macOS. Engineered for quantitative desks, algorithmic scalpers, and institutional portfolio managers, this platform combines the expressive rendering capabilities of Flutter 3.22+ Desktop with low-level macOS AppKit/Cocoa primitives, Apple Metal hardware acceleration, Apple Silicon vectorization, and a zero-GC Rust FFI network core.

```
+---------------------------------------------------------------------------------------+
|                                    macOS User Space                                   |
|                                                                                       |
|  +---------------------------+  +---------------------------+  +-------------------+  |
|  |     NSApplication Host    |  |     NSMenu Hierarchy      |  |   Dock Tile View  |  |
|  |   (AppKit / Cocoa Shell)  |  |   (Cmd+1 .. Cmd+9 / Hot)  |  | (Live PnL Badge)  |  |
|  +-------------+-------------+  +-------------+-------------+  +---------+---------+  |
|                |                              |                          |            |
|  +-------------v------------------------------v--------------------------v---------+  |
|  |                     Flutter 3.22+ Desktop Engine (Dart 3.4+)                    |  |
|  |                                                                                 |  |
|  |  +--------------------+  +----------------------+  +-------------------------+  |  |
|  |  |  Trading Workspaces |  |  Multi-Window Engine |  |  0.00% Zero-Fee Engine  |  |  |
|  |  |  (L2/L3 Orderbooks)|  |  (Studio / XDR Detach)|  |  (0 Gas Sponsor Badge)  |  |  |
|  |  +---------+----------+  +----------+-----------+  +------------+------------+  |  |
|  +------------|------------------------|---------------------------|---------------+  |
|               |                        |                           |                  |
|  +------------v------------+  +--------v-------------+  +----------v---------------+  |
|  |  Flutter Impeller Engine|  | Native AppKit Bridge |  | LocalAuthentication (LA) |  |
|  |  Metal Backend (MSL AOT)|  | (NSScreen / Windows) |  | TouchID / Secure Enclave |  |
|  +------------+------------+  +----------------------+  +----------+---------------+  |
|               |                                                    |                  |
|  +------------v----------------------------------------------------v---------------+  |
|  |            Rust FFI Socket Core (librust_trading_core.dylib)                    |  |
|  |    Tokio Async IO | Zero-GC SPSC Ring Buffer | Lock-Free Shared Memory Ring     |  |
|  +-------------------------------------+-------------------------------------------+  |
+----------------------------------------|----------------------------------------------+
                                         |
+----------------------------------------v----------------------------------------------+
|                         Apple Silicon Hardware Subsystems                             |
|                                                                                       |
|  +-------------------+  +--------------------+  +-------------------+  +-----------+  |
|  | Unified Memory    |  | Metal GPU Cores    |  | Secure Enclave    |  | ProMotion |  |
|  | Architecture (UMA)|  | (120 FPS Impeller) |  | Processor (SEP)   |  | (120Hz)   |  |
|  +-------------------+  +--------------------+  +-------------------+  +-----------+  |
+---------------------------------------------------------------------------------------+
```

### Architectural Directives:
1. **Lock-Free Zero-GC Market Streaming:** High-frequency binary market data feeds bypass the Dart garbage-collected heap entirely via a C-compatible foreign function interface (`rust_trading_core.dylib`), feeding lock-free single-producer single-consumer (SPSC) ring buffers directly into memory mapped structures.
2. **Deterministic 120 FPS ProMotion Rendering:** The graphics pipeline leverages Flutter Impeller running exclusively over Apple Metal (MSL), synchronized with `CVDisplayLink` and `CADisplayLink` to eliminate all runtime shader compilation jank and enforce an 8.33ms frame budget across Apple ProMotion panels.
3. **Hardware-Bound Security:** Cryptographic signing keys, authentication tokens, and high-notional order authorizations are anchored to the Apple Secure Enclave Processor (SEP) via `LocalAuthentication` TouchID/FaceID biometrics, prohibiting private key extraction under any operational condition.
4. **Native macOS Workstation Ergonomics:** Complete integration into macOS desktop paradigms including multi-monitor detachment across 5K Studio Displays and 6K Pro Display XDRs with P3 wide color gamut synchronization, native `NSMenu` shortcuts (`Cmd+1` to `Cmd+9`), and real-time Dock tile badging.
5. **Zero-Fee Financial Invariant:** The interface immutably enforces and displays the platform's core commitment: 0.00% Maker, 0.00% Taker, 0.00% Platform Fees, and Sponsored 0 Gas blockchain execution across all order ticket presentations.

---

## 2. Native AppKit/Cocoa Application Lifecycle & Flutter 3.22+ Integration

The thick client utilizes a native Objective-C/Swift AppKit host that embeds the Flutter 3.22+ Desktop engine as a primary view controller while retaining direct control over macOS system events, windowing, and the Cocoa run loop.

### 2.1 Cocoa Application Lifecycle Architecture

The application entry point is managed by `NSApplication` and governed by an extended `AppDelegate` adhering to `NSApplicationDelegate`:

```
+------------------------+
| NSApplicationMain(...) |
+-----------+------------+
            |
+-----------v------------+
|  applicationWill-      | ---> Initialize Rust Core Dynamic Library
|  FinishLaunching:      | ---> Validate Mach-O Architecture & CPU Features
+-----------+------------+
            |
+-----------v------------+
|  applicationDid-       | ---> Initialize FlutterEngine & Impeller Metal Layer
|  FinishLaunching:      | ---> Configure Main NSWindow & Vibrant Titlebar
|                        | ---> Register Global Hotkeys & Dock Tile Subsystem
+-----------+------------+
            |
+-----------v------------+      +---------------------------+
| Active Application     | <--> | applicationDidBecomeActive|
| Run Loop (NSEvent)     |      | applicationDidResignActive|
+-----------+------------+      +---------------------------+
            |
+-----------v------------+
|  applicationShould-    | ---> Institutional Invariant: Return NO (Remain in Menu Bar
|  TerminateAfterLast-   |      when windows close, preserving background order routing)
|  WindowClosed:         |
+-----------+------------+
            |
+-----------v------------+
|  applicationWill-      | ---> Flush In-Flight Ledger Journals to SQLite WAL
|  Terminate:            | ---> Graceful TCP Disconnect & Zero-Memory Wipe
+------------------------+
```

### 2.2 NSWindow Configuration and Vibrant Glass Styling

The trading workstation delivers an immersive, dark-mode-first aesthetic (Obsidian Dark `#0B0E14`) utilizing AppKit's native material vibrancy system:

```swift
// Native Cocoa Window Initialization Specification
let windowMask: NSWindow.StyleMask = [
    .titled,
    .closable,
    .miniaturizable,
    .resizable,
    .fullSizeContentView
]

let window = NSWindow(
    contentRect: initialRect,
    styleMask: windowMask,
    backing: .buffered,
    defer: false
)

// Configure Titlebar and Chrome Vibrancy
window.titlebarAppearsTransparent = true
window.titleVisibility = .hidden
window.isMovableByWindowBackground = true
window.backgroundColor = NSColor(calibratedRed: 11/255, green: 14/255, blue: 20/255, alpha: 1.0)

// Visual Effect Substrate (Behind-Window Blending)
let visualEffect = NSVisualEffectView()
visualEffect.blendingMode = .behindWindow
visualEffect.material = .underWindowBackground
visualEffect.state = .active
window.contentView = visualEffect
```

### 2.3 FlutterEngine Embedding and Bridge Protocol

The Flutter desktop runtime is hosted inside an `NSViewController` container. Communication between Dart and Cocoa is handled via two specialized bidirectional channels:

1. **System Control Channel (`growww/macos/system`):** Dispatches asynchronous window management, dock updates, display metrics, and biometric requests.
2. **Native Event Stream (`growww/macos/events`):** High-frequency stream for hardware state changes, thermal warnings, and external display hotplug events.

```
+--------------------+        MethodChannel ("growww/macos/system")        +-------------------+
|                    | --------------------------------------------------> |                   |
|   Dart Runtime     |                                                     |   macOS Cocoa     |
|   (Flutter 3.22+)  | <-------------------------------------------------- |   Host Engine     |
|                    |         EventChannel ("growww/macos/events")        |                   |
+--------------------+                                                     +-------------------+
```

---

## 3. Metal Hardware Acceleration via Impeller Engine & 120 FPS ProMotion Displays

To guarantee institutional-grade execution responsiveness, the thick client relies entirely on Flutter's next-generation **Impeller** rendering engine configured with Apple Metal.

### 3.1 Impeller Metal Pipeline & MSL Ahead-Of-Time Compilation

Traditional Flutter rendering on Skia suffered from runtime compilation jank as OpenGL or Metal shaders compiled on the main rendering thread during first presentation. Impeller eliminates this architectural flaw:

```
Compile-Time / Packaging:
+------------------------+      impellerc      +-------------------------+
| Impeller Shaders (GLSL)| ------------------> | Metal Shading Language  |
| & MSL Source Files     |                     | (MSL .metallib Binary)  |
+------------------------+                     +------------+------------+
                                                            |
Runtime Execution:                                          |
+------------------------+      Pre-Warmed     +------------v------------+
| Render Tree Traversal  | ------------------> | Pipeline State Objects  |
| (Order Book / Charts)  |                     | (MTLRenderPipelineState)|
+-----------+------------+                     +------------+------------+
            |                                               |
            v                                               v
+------------------------------------------------------------------------+
| Metal Command Buffer Encoding (MTLCommandBuffer & MTLRenderCommandEncoder)
| Zero Runtime Shader Compilation Jank | 0 Dropped Frames During Volatility
+------------------------------------------------------------------------+
```

### 3.2 120Hz ProMotion Synchronization via CVDisplayLink and CADisplayLink

Apple ProMotion panels dynamically alternate between 24Hz and 120Hz. For high-frequency trading, frame drop during rapid order book delta flashing or microsecond order ladder updates creates visual fatigue and latency perception.

The client locks the display link to 120 FPS during active trading sessions:

```
ProMotion Dynamic Refresh Arbitration:
- Passive Screen State (No Market Ticks): 30Hz - 60Hz (Energy Preservation)
- Active Order Book Streaming (>50 ticks/sec): Locked 120Hz (8.33ms per frame)
- Chart Dragging / Real-Time Pan: Locked 120Hz (8.33ms per frame)
```

```swift
// CVDisplayLink 120 FPS Synchronization Specification
var displayLink: CVDisplayLink?
CVDisplayLinkCreateWithActiveCGDisplays(&displayLink)

guard let link = displayLink else { return }

CVDisplayLinkSetOutputCallback(link, { (displayLink, inNow, inOutputTime, flagsIn, flagsOut, context) -> CVReturn in
    let targetTime = inOutputTime.pointee
    // Target frame period: 8,333,333 nanoseconds (120 FPS)
    let frameDurationNs = targetTime.videoRefreshPeriod
    
    // Dispatch tick render pass to Flutter Impeller CAMetalLayer
    NotificationCenter.default.post(
        name: .proMotionFrameSignal,
        object: nil,
        userInfo: ["frameTimestamp": targetTime.hostTime]
    )
    return kCVReturnSuccess
}, nil)

CVDisplayLinkStart(link)
```

### 3.3 Frame Budget Allocation Under 120 FPS (8.33ms)

Every micro-frame must strictly conform to the 8.33 millisecond execution envelope:

| Pipeline Stage | Max Allocated Time | Subsystem Responsible | Invariant Behavior |
| :--- | :--- | :--- | :--- |
| **Tick Extraction** | 0.80 ms | Rust FFI SPSC Ring Buffer | Zero heap allocation, pointer swap |
| **State Reconciliation** | 1.50 ms | Dart ViewModel / Riverpod | Diff calculation on order book levels |
| **Layout & Render Tree** | 2.00 ms | Flutter Framework Core | RenderCustomMultiChildLayoutBox |
| **Metal Command Encoding** | 1.80 ms | Impeller Metal Backend | Command buffer population (`MTLCommandBuffer`) |
| **GPU Rasterization** | 2.00 ms | Apple Silicon Metal GPU | Fragment shader execution on unified RAM |
| **Margin / Safety Headroom** | 0.23 ms | Engine Scheduler | Prevents frame drop overruns |
| **Total Target** | **8.33 ms** | **End-to-End Budget** | **120 FPS Locked Cadence** |

### 3.4 Thermal State Management & Dynamic Frame Throttling

To sustain long-running trading sessions on MacBook Pro hardware without inducing system-level fan noise or hardware degradation, the macOS host monitors `ProcessInfo.thermalStateDidChangeNotification`:

```
Nominal State:
- Impeller Target: Locked 120 FPS
- Order Book Depth: Full 100 levels rendered

Fair / Moderate Thermal State:
- Impeller Target: 120 FPS maintained
- Chart Anti-Aliasing: Reduced from 4x MSAA to 2x MSAA

Serious Thermal State:
- Impeller Target: Throttled to 60 FPS (16.67ms frame budget)
- Order Book Depth: Aggregated to top 25 visible levels

Critical Thermal State:
- Impeller Target: Throttled to 30 FPS
- Visual effects and glass blur (`NSVisualEffectView`) disabled
```

---

## 4. Apple Silicon Native Compilation & Intel Universal Binary Architecture

The production thick client is distributed as a Universal Mach-O 2-way fat binary targeting both native Apple Silicon ARM64 and legacy Intel x86_64 architectures.

```
                    Build Automation Pipeline
                                |
          +---------------------+---------------------+
          |                                           |
          v                                           v
Target: aarch64-apple-darwin               Target: x86_64-apple-darwin
- Flutter AOT Engine (ARM64)               - Flutter AOT Engine (x86_64)
- Rust Core (librust_trading_arm64.dylib)  - Rust Core (librust_trading_x86.dylib)
- Native Cocoa Launcher (ARM64)            - Native Cocoa Launcher (x86_64)
          |                                           |
          +---------------------+---------------------+
                                |
                                v
               Universal Mach-O Lipidation Tooling
       lipo -create -output GrowwwUniversal Growww_arm64 Growww_x86
                                |
                                v
          +-------------------------------------------+
          |         Universal Mach-O Binary           |
          |                                           |
          |  [Fat Header (FAT_MAGIC / FAT_MAGIC_64)]  |
          |  +-- Architecture: ARM64 (M1/M2/M3/M4)    |
          |  |   Alignment: 16KB (2^14 bytes)         |
          |  +-- Architecture: x86_64 (Intel 64-bit)  |
          |      Alignment: 4KB (2^12 bytes)          |
          +-------------------------------------------+
```

### 4.1 Apple Silicon Micro-Architectural Optimizations

On Apple Silicon (M1/M2/M3/M4 Pro, Max, and Ultra), the binary unlocks specialized hardware capabilities:

1. **16KB Page Alignment:** The binary is strictly linked with `-Wl,-pagezero_size,0x1000 -Wl,-segalign,0x4000` to satisfy ARM64 Darwin page boundary alignment.
2. **ARM NEON Vector Execution:** Order book depth sorting and VWAP/EMA technical indicators leverage 128-bit NEON SIMD registers, processing four 32-bit floats or two 64-bit doubles per clock cycle.
3. **Unified Memory Architecture (UMA) Zero-Copy Buffering:** Memory allocated for order book charts and depth textures is configured with `MTLResourceStorageModeShared`. Both the CPU (Rust/Dart) and the Metal GPU core address the identical physical memory address, bypassing PCIe bus data transfer overhead.

### 4.2 Code Signing, Hardened Runtime, and Apple Notarization

To comply with macOS Gatekeeper and guarantee tamper-proof execution, the release artifact passes through Apple's strict cryptographic verification pipeline:

```bash
# Hardened Runtime Code Signing
codesign --force --verify --verbose \
  --sign "Developer ID Application: Growww Securities Ltd (NBSE)" \
  --options runtime \
  --entitlements build/macos/Entitlements.plist \
  Universal/Growww.app

# Submission to Apple Notary Service
xcrun notarytool submit Growww.dmg \
  --keychain-profile "NBSE-Notary-Profile" \
  --wait

# Stapling Cryptographic Notarization Ticket
xcrun stapler staple Growww.dmg
```

Required security entitlements in `Entitlements.plist`:
- `com.apple.security.cs.allow-jit` (Required for Dart VM development mode; disabled in production AOT releases).
- `com.apple.security.cs.allow-unsigned-executable-memory` (Disabled).
- `com.apple.security.cs.disable-library-validation` (Disabled).
- `com.apple.security.network.client` (Enabled for high-speed TCP/WebSocket egress).
- `com.apple.security.smartcard` (Enabled for hardware-bound HSM authentication).

---

## 5. Apple Secure Enclave & TouchID/FaceID Biometric Authentication

The thick client utilizes Apple's dedicated hardware security chip (the Secure Enclave Processor or SEP) to authenticate traders, protect private key materials, and authorize financial transactions.

```
+-------------------------------------------------------------------------------+
| User Action: Place High-Notional Order (>100,000 USDT) / Confirm Withdrawal   |
+---------------------------------------+---------------------------------------+
                                        |
                                        v
+-------------------------------------------------------------------------------+
| AppKit Cocoa Layer: LocalAuthentication Framework (LAContext)                 |
| Policy: LAPolicyDeviceOwnerAuthenticationWithBiometrics                       |
+---------------------------------------+---------------------------------------+
                                        |
                                        v
+-------------------------------------------------------------------------------+
| Secure Enclave Hardware Boundary (Isolated Microkernel & Storage)             |
|                                                                               |
| 1. TouchID Sensor reads fingerprint / FaceID sensor scans depth map           |
| 2. Biometric matching occurs strictly inside SEP isolated silicon             |
| 3. Hardware matches template -> Unlocks 256-bit Key inside Secure Enclave     |
| 4. SEP signs the transaction payload using secp256r1 or Ed25519               |
| 5. Raw private key is NEVER exported to macOS kernel or user memory space     |
+---------------------------------------+---------------------------------------+
                                        |
                   +--------------------+--------------------+
                   |                                         |
                   v                                         v
         [Biometric Verified]                      [Biometric Rejected]
                   |                                         |
                   v                                         v
      Returns Cryptographic Signature          Aborts Order Dispatch
      to Rust Network Core                     Displays Visual Error & Logs
```

### 5.1 LocalAuthentication Implementation Specification

```swift
import LocalAuthentication
import Security

public class SecureEnclaveAuthManager {
    public static let shared = SecureEnclaveAuthManager()
    
    public func authorizeTradingAction(
        prompt: String,
        completion: @escaping (Bool, Error?) -> Void
    ) {
        let context = LAContext()
        context.localizedCancelTitle = "Cancel Order"
        context.localizedFallbackTitle = "Use Master Passphrase"
        
        var evalError: NSError?
        let policy: LAPolicy = .deviceOwnerAuthenticationWithBiometrics
        
        if context.canEvaluatePolicy(policy, error: &evalError) {
            context.evaluatePolicy(policy, localizedReason: prompt) { success, error in
                DispatchQueue.main.async {
                    completion(success, error)
                }
            }
        } else {
            // Fallback to Device Owner Authentication (Hardware PIN / System Password)
            context.evaluatePolicy(.deviceOwnerAuthentication, localizedReason: prompt) { success, error in
                DispatchQueue.main.async {
                    completion(success, error)
                }
            }
        }
    }
}
```

### 5.2 Hardware Key Generation & Key Invalidation Safeguards

Keys are generated directly inside the Secure Enclave utilizing `kSecAccessControlBiometryCurrentSet`. If a user alters macOS TouchID biometric enrollments (adds or removes a fingerprint), the operating system permanently invalidates the hardware key, preventing unauthorized access if physical access is compromised:

```swift
let accessControl = SecAccessControlCreateWithFlags(
    kCFAllocatorDefault,
    kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
    [.biometryCurrentSet, .privateKeyUsage],
    nil
)!

let attributes: [String: Any] = [
    kSecClass as String: kSecClassKey,
    kSecAttrKeyType as String: kSecAttrKeyTypeECSECPrimeRandom,
    kSecAttrKeySizeInBits as String: 256,
    kSecAttrTokenID as String: kSecAttrTokenIDSecureEnclave,
    kSecPrivateKeyAttrs as String: [
        kSecAttrIsPermanent as String: true,
        kSecAttrApplicationTag as String: "com.growww.pro.trading.authKey".data(using: .utf8)!,
        kSecAttrAccessControl as String: accessControl
    ]
]
```

---

## 6. Native macOS Menu Bar (NSMenu) Structure & Global Shortcuts

The thick client adheres to Apple Human Interface Guidelines (HIG) by exposing a fully articulated native `NSMenu` hierarchy. All core trading workspaces and risk actions are mapped to physical hardware shortcuts.

```
+----------------------------------------------------------------------------------------------------+
|    Growww Pro  File  Edit  View  Workspaces  Trading  Orders  Analytics  Window  Help  [0.00% FEE]|
+----------------------------------------------------------------------------------------------------+
```

### 6.1 Menu Hierarchy & Keyboard Shortcut Matrix

| Menu Group | Menu Item Label | Key Equivalent | Action Target / Dispatched Route |
| :--- | :--- | :--- | :--- |
| **Workspaces** | Spot Trading Console | `Cmd+1` | `workspace_router.navigate("/spot")` |
| **Workspaces** | Perpetual Futures & Margin | `Cmd+2` | `workspace_router.navigate("/perpetuals")` |
| **Workspaces** | Institutional L3 Microstructure | `Cmd+3` | `workspace_router.navigate("/l3-depth")` |
| **Workspaces** | Options Volatility Matrix | `Cmd+4` | `workspace_router.navigate("/options")` |
| **Workspaces** | RWA Primary Issuance Portal | `Cmd+5` | `workspace_router.navigate("/rwa-bonds")` |
| **Workspaces** | Algorithmic TWAP/VWAP Terminal | `Cmd+6` | `workspace_router.navigate("/algo-desk")` |
| **Workspaces** | Balances, Treasury & Ledger | `Cmd+7` | `workspace_router.navigate("/portfolio")` |
| **Workspaces** | Statutory Tax & 194S TDS Portal | `Cmd+8` | `workspace_router.navigate("/compliance")` |
| **Workspaces** | System Hardware & Secure Enclave | `Cmd+9` | `workspace_router.navigate("/settings")` |
| **Trading** | Quick Buy / Long Ladder | `Cmd+B` | Focuses bid ladder order entry |
| **Trading** | Quick Sell / Short Ladder | `Cmd+S` | Focuses ask ladder order entry |
| **Trading** | Emergency Cancel All Orders | `Cmd+Shift+X` | Global Panic Broadcast (Immediate) |
| **View** | Detach Focused Widget | `Cmd+Shift+D` | Creates independent Cocoa `NSWindow` |
| **View** | Reset to Standard Multi-Pane | `Cmd+0` | Re-docks all detached windows |
| **Window** | Toggle Fullscreen Mode | `Cmd+Ctrl+F` | Enters native macOS space fullscreen |

### 6.2 Global Hotkey Event Tap (System-Wide Intercept)

For institutional risk control, the **Emergency Cancel All Orders** (`Cmd+Shift+X`) hotkey can be configured to trigger globally via `NSEvent.addGlobalMonitorForEvents` or Carbon Event APIs, ensuring immediate risk mitigation even if the user is focused on a separate spreadsheet or terminal:

```swift
// Register System-Wide Risk Hotkey
NSEvent.addGlobalMonitorForEvents(matching: .keyDown) { event in
    // Intercept Cmd+Shift+X (KeyCode 7 for 'X' with Command + Shift modifier masks)
    if event.modifierFlags.contains([.command, .shift]) && event.keyCode == 7 {
        RiskSafetySubsystem.dispatchEmergencyCancelAll(reason: "Global Panic Shortcut")
    }
}
```

---

## 7. Dock Icon Badge Displaying Real-Time Daily PnL & Order Alerts

The client utilizes the native `NSDockTile` subsystem to communicate portfolio status, active order alerts, and daily PnL directly on the macOS Dock without requiring window focus.

```
+-----------------------------------+
|          [ Growww Icon ]          |
|                                   |
|   +---------------------------+   |
|   |   +$14.2K   |  (3 Alerts) |   |  <--- Custom Rendered Retina Dock Tile
|   +---------------------------+   |
+-----------------------------------+
```

### 7.1 Dock Tile Rendering Pipeline

Instead of using basic static text strings via `dockTile.badgeLabel`, the thick client instantiates a custom high-DPI `NSView` that renders a color-coded financial badge:

```swift
public class PortfolioDockTileView: NSView {
    public var dailyPnL: Double = 0.0
    public var activeFillsCount: Int = 0
    
    public override func draw(_ dirtyRect: NSRect) {
        // Step 1: Draw App Icon Substrate
        NSApplication.shared.applicationIconImage?.draw(in: bounds)
        
        // Step 2: Formulate Color Matrix
        let isPositive = dailyPnL >= 0.0
        let badgeColor = isPositive 
            ? NSColor(red: 0.0/255, green: 200.0/255, blue: 5.0/255, alpha: 0.95)   // Emerald Green (#00C805)
            : NSColor(red: 255.0/255, green: 59.0/255, blue: 48.0/255, alpha: 0.95) // Ruby Crimson (#FF3B30)
            
        // Step 3: Draw Lower PnL Pill Badge
        let pillRect = NSRect(x: 4, y: 6, width: bounds.width - 8, height: 32)
        let pillPath = NSBezierPath(roundedRect: pillRect, xRadius: 8, yRadius: 8)
        badgeColor.setFill()
        pillPath.fill()
        
        // Step 4: Render High-Contrast Typography
        let pnlText = (isPositive ? "+" : "") + formatCurrency(dailyPnL)
        let attributes: [NSAttributedString.Key: Any] = [
            .font: NSFont.monospacedDigitSystemFont(ofSize: 18, weight: .black),
            .foregroundColor: NSColor.white
        ]
        let stringSize = pnlText.size(withAttributes: attributes)
        let textOrigin = NSPoint(
            x: pillRect.midX - (stringSize.width / 2.0),
            y: pillRect.midY - (stringSize.height / 2.0)
        )
        pnlText.draw(at: textOrigin, withAttributes: attributes)
        
        // Step 5: Render Active Fills Alert Pill (Top-Right Badge)
        if activeFillsCount > 0 {
            let alertRect = NSRect(x: bounds.width - 34, y: bounds.height - 34, width: 28, height: 28)
            let alertPath = NSBezierPath(ovalIn: alertRect)
            NSColor(red: 0.0/255, green: 122.0/255, blue: 255.0/255, alpha: 1.0).setFill()
            alertPath.fill()
            
            let countText = "\(activeFillsCount)"
            let alertAttrs: [NSAttributedString.Key: Any] = [
                .font: NSFont.boldSystemFont(ofSize: 14),
                .foregroundColor: NSColor.white
            ]
            countText.draw(at: NSPoint(x: alertRect.midX - 5, y: alertRect.midY - 8), withAttributes: alertAttrs)
        }
    }
}
```

### 7.2 Throttling and Update Cadence

To safeguard system resources and prevent battery degradation on MacBooks:
- **Normal Market Operation:** Dock tile redrawing is throttled to 1 Hz (`1 update per second`).
- **Immediate Event Push:** Instant invalidation (`NSApplication.shared.dockTile.display()`) triggers on order executions, liquidation warnings, or stop-loss hits.
- **Privacy Mode:** When macOS screen lock is engaged, the PnL text string immediately masks to `******` to prevent shoulder-surfing.

---

## 8. Multi-Display External Monitor Detachment (Studio Displays & Pro Display XDR)

Professional trading setups demand multi-monitor detachment across external displays. The thick client supports detaching charts, order books, and depth ladders into independent, hardware-accelerated child windows.

```
+---------------------------------------------------------------------------------------+
|                                    NSScreen Subsystem                                 |
|                                                                                       |
|  +---------------------------+  +---------------------------+  +-------------------+  |
|  | Screen 1: Built-in Panel  |  | Screen 2: Studio Display  |  | Screen 3: Pro XDR |  |
|  | Liquid Retina XDR (120Hz) |  | 5K Resolution (60Hz P3)   |  | 6K EDR (1600 nits)|  |
|  | Main Executive Dashboard  |  | Multi-Pair Chart Array    |  | L3 Depth Ladder   |  |
|  +-------------+-------------+  +-------------+-------------+  +---------+---------+  |
|                |                              |                          |            |
|  +-------------v------------------------------v--------------------------v---------+  |
|  |                 Inter-Window Synchronization Bus (Link Groups)                  |  |
|  |                 Symbol Linking: Group Red / Blue / Green / Yellow               |  |
|  |                 Sub-millisecond State Broadcast via Shared Memory               |  |
|  +---------------------------------------------------------------------------------+  |
+---------------------------------------------------------------------------------------+
```

### 8.1 Multi-Window AppKit Lifecycle Management

Each detached window is an independent `NSWindow` containing a distinct or shared `FlutterEngine` instance:

```swift
public class DetachedWindowCoordinator {
    private var childWindows: [UUID: NSWindow] = [:]
    
    public func spawnDetachedWidget(
        widgetIdentifier: String,
        targetScreen: NSScreen,
        linkGroup: String
    ) -> UUID {
        let windowId = UUID()
        
        let screenRect = targetScreen.visibleFrame
        let targetRect = NSRect(
            x: screenRect.origin.x + 100,
            y: screenRect.origin.y + 100,
            width: 1280,
            height: 800
        )
        
        let detachedWindow = NSWindow(
            contentRect: targetRect,
            styleMask: [.titled, .closable, .miniaturizable, .resizable],
            backing: .buffered,
            defer: false
        )
        
        detachedWindow.title = "Growww Pro Detached - \(widgetIdentifier)"
        detachedWindow.isReleasedWhenClosed = false
        detachedWindow.tabbingMode = .disallowed
        
        // Retain and present
        childWindows[windowId] = detachedWindow
        detachedWindow.makeKeyAndOrderFront(nil)
        
        return windowId
    }
}
```

### 8.2 Wide Color Gamut (Display P3) & Extended Dynamic Range (EDR)

Institutional charts and market ladders are displayed across varying hardware panels. The client dynamically queries display color spaces:

1. **Apple Studio Display (5K 5120x2880 @ 60Hz):** Uses the `CGColorSpace.displayP3` color space. Color coordinates are mapped into DCI-P3 to prevent color clipping in neon green/red book indicators.
2. **Apple Pro Display XDR (6K 6016x3384 @ 60Hz / 1600 nits EDR):** When rendering on XDR panels, the Metal fragment shaders unlock extended dynamic range highlights. Liquidation clusters and high-volume nodes illuminate with HDR brilliance (`EDR headroom up to 4.0x`), distinguishing them against standard SDR chart candles.

### 8.3 Window Topology State Restoration

The client implements `NSWindowRestoration` to persist multi-monitor layouts:
- Display monitor hardware UUIDs (`CGDisplayCreateUUIDFromDisplayID`) are serialized to SQLite.
- Window coordinates, detached widget IDs, and Link Group subscriptions are automatically restored when monitors are reconnected or the application relaunches.

---

## 9. Rust FFI Socket Core (`rust_trading_core.dylib`) for Zero-GC Streaming

Garbage collection pauses in client runtimes are unacceptable during market volatility. When an exchange dispatches 100,000 order book updates per second, a single 15ms GC pause causes stale pricing and dropped frames.

The thick client delegates network transport, binary parsing, and order book state management to a native dynamic library compiled in Rust (`librust_trading_core.dylib`).

```
+-----------------------------------------------------------------------------------+
|                        TCP / TLS Binary WebSocket Connection                      |
+-----------------------------------------+-----------------------------------------+
                                          |
                                          v
+-----------------------------------------------------------------------------------+
| Rust Core Thread (Tokio Async Runtime - OS Native Thread)                         |
|                                                                                   |
| 1. High-speed socket read into pre-allocated memory buffer                        |
| 2. Zero-copy Protobuf / SBE parsing                                               |
| 3. Directly writes parsed C-struct into Lock-Free SPSC Ring Buffer                |
| 4. Never touches Dart garbage-collected heap                                      |
+-----------------------------------------+-----------------------------------------+
                                          |
                                          v
+-----------------------------------------------------------------------------------+
| Lock-Free SPSC (Single Producer Single Consumer) Ring Buffer                      |
|                                                                                   |
| [Slot 0] [Slot 1] [Slot 2] [Slot 3] ... [Slot N] (#[repr(align(64))] Cache Line)  |
| Atomic Head Pointer (Rust) <---------------> Atomic Tail Pointer (Dart/Metal)    |
+-----------------------------------------+-----------------------------------------+
                                          |
                                          v
+-----------------------------------------------------------------------------------+
| Dart Consumer Thread (dart:ffi NativePort / Direct Memory Access)                 |
|                                                                                   |
| 1. Reads raw memory pointer without object allocation                             |
| 2. Feeds Impeller Metal vertex buffer directly                                    |
| 3. Result: 0.0ms GC Pause Time even during extreme volume spikes                  |
+-----------------------------------------------------------------------------------+
```

### 9.1 SPSC Ring Buffer C-Compatible ABI Specification

The shared memory ring buffer is defined in Rust with strict C representation and 64-byte cache line alignment to prevent false sharing across Apple Silicon performance cores:

```rust
// rust_trading_core/src/ring_buffer.rs

#[repr(C)]
#[repr(align(64))]
pub struct MarketTick {
    pub timestamp_ns: u64,
    pub symbol_id: u32,
    pub price_mantissa: i64,
    pub price_exponent: i32,
    pub size_mantissa: u64,
    pub size_exponent: i32,
    pub side: u8, // 0 = Buy, 1 = Sell
    pub is_trade: u8,
    pub sequence_number: u64,
    pub _padding: [u8; 14], // Enforce exact 64-byte cache line fit
}

#[repr(C)]
pub struct RingBufferHeader {
    pub head: std::sync::atomic::AtomicUsize,
    pub tail: std::sync::atomic::AtomicUsize,
    pub capacity: usize,
    pub mask: usize,
    pub buffer_ptr: *mut MarketTick,
}

#[no_mangle]
pub extern "C" fn growww_ring_buffer_init(capacity: usize) -> *mut RingBufferHeader {
    // Capacity must be a power of two for bitwise wrapping
    assert!(capacity.is_power_of_two());
    let mut storage = Vec::<MarketTick>::with_capacity(capacity);
    let buffer_ptr = storage.as_mut_ptr();
    std::mem::forget(storage);

    let header = Box::new(RingBufferHeader {
        head: std::sync::atomic::AtomicUsize::new(0),
        tail: std::sync::atomic::AtomicUsize::new(0),
        capacity,
        mask: capacity - 1,
        buffer_ptr,
    });

    Box::into_raw(header)
}
```

### 9.2 Dart FFI Zero-Copy Binding

Dart consumes this buffer via `dart:ffi`, bypassing string or object construction:

```dart
// Dart 3.4+ Zero-Copy Consumption Pipeline
import 'dart:ffi' as ffi;

final class MarketTickNative extends ffi.Struct {
  @ffi.Uint64()
  external int timestampNs;

  @ffi.Uint32()
  external int symbolId;

  @ffi.Int64()
  external int priceMantissa;

  @ffi.Int32()
  external int priceExponent;

  @ffi.Uint64()
  external int sizeMantissa;

  @ffi.Int32()
  external int sizeExponent;

  @ffi.Uint8()
  external int side;

  @ffi.Uint8()
  external int isTrade;

  @ffi.Uint64()
  external int sequenceNumber;

  @ffi.Array(14)
  external ffi.Array<ffi.Uint8> padding;
}

class FastTickConsumer {
  final ffi.Pointer<MarketTickNative> _buffer;
  final int _mask;
  int _tail = 0;

  FastTickConsumer(this._buffer, int capacity) : _mask = capacity - 1;

  void drainAvailableTicks(void Function(ffi.Pointer<MarketTickNative>) onTick) {
    // Read directly from unmanaged memory pointer
    final currentHead = readAtomicHead();
    while (_tail != currentHead) {
      final tickPtr = _buffer.elementAt(_tail & _mask);
      onTick(tickPtr);
      _tail++;
    }
  }
}
```

---

## 10. Strict 0.00% Zero-Fee Presentation & 0 Gas Sponsorship Badge

The platform operates on a mathematically transparent, zero-friction financial model. The macOS Thick Client implements strict presentation invariants to reflect this guarantee across every trading interface.

```
+-------------------------------------------------------------------------------+
|                            ORDER CONFIRMATION HUD                             |
|                                                                               |
|  Pair: BTC/USDT (Perpetual)                     Side: BUY / LONG              |
|  Quantity: 2.5000 BTC                           Price: $64,250.00             |
|  Notional Value: $160,625.00                    Leverage: 20x Cross           |
|                                                                               |
|  ---------------------------------------------------------------------------  |
|  FEE SCHEDULE INVARIANTS:                                                     |
|  Maker Fee:                                   0.00% (FREE)                    |
|  Taker Fee:                                   0.00% (FREE)                    |
|  Exchange Platform Fee:                       0.00% (FREE)                    |
|  Blockchain Settlement Gas:                   $0.00 (SPONSORED BY NBSE)       |
|                                                                               |
|  [ ✓ 0.00% ZERO-FEE VERIFIED ]          [ ⚡ 0 GAS SPONSORED PAYMASTER ]      |
+-------------------------------------------------------------------------------+
```

### 10.1 Mathematical and UI Invariants

1. **Maker & Taker Fee Display:** The UI unconditionally presents `0.00%` or `FREE`. No hidden markups or spread inflations are allowed. Any dynamic calculation yielding fees greater than `0.0000` triggers an immediate UI invariant failure and halts order placement.
2. **0 Gas Sponsorship Badge:** Blockchain settlement transactions executed via account abstraction (ERC-4337 / EIP-712 paymaster) feature a permanent emerald sponsorship badge: `0 Gas Sponsored`.
3. **Menu Bar Status Integration:** The global macOS menu bar right-hand status item displays real-time invariant verification: `NBSE: 0.00% Fee Active`.
4. **Tax Compliance Segregation:** Statutory taxes (such as Indian Section 194S 1% TDS) are explicitly categorized as legal government deductions, kept strictly distinct from the exchange's zero-fee execution schedule.

---

## 11. Security, Hardening & Compliance Invariants

The macOS Thick Client enforces defense-in-depth system security:

```
+-------------------------------------------------------------------------------+
|                        macOS Security Hardening Rings                         |
|                                                                               |
|  [ Ring 0: Secure Enclave ]  - TouchID, Ed25519 Signing, Key Non-Export       |
|  [ Ring 1: Memory Hardening] - 0-GC Rust Ring Buffers, Zero-Wipe on Drop      |
|  [ Ring 2: OS Hardening ]    - Hardened Runtime, Code Signing, Notarization   |
|  [ Ring 3: App Sandbox ]     - Explicit Network Egress, Restricted File IO    |
+-------------------------------------------------------------------------------+
```

1. **Zero-Memory Wipe:** Sensitive cryptographic seeds, session tokens, and unencrypted private keys implement the `Zeroize` trait in Rust and `memset_s` in C, zeroing physical RAM immediately upon deallocation.
2. **Anti-Debugging & Tamper Resistance:** Production Mach-O binaries invoke `ptrace(PT_DENY_ATTACH, 0, 0, 0)` during Cocoa initialization, blocking unauthorized debugger attachment and memory dumping.
3. **Encrypted Local Cache:** Local SQLite databases storing layout states, order histories, and audit trails operate exclusively in Write-Ahead Logging (WAL) mode protected by SQLCipher with hardware AES-256 keys retrieved from the macOS Keychain.

---

## 12. Verification & Performance Benchmarking Matrix

Before any production build is notarized and released, it must pass this verification matrix:

| Metric / Objective | Target Requirement | Profiling Tool / Verification Method |
| :--- | :--- | :--- |
| **Frame Rate (ProMotion)** | 120 FPS Locked (8.33ms) | Xcode Instruments: Metal System Trace |
| **Shader Compilation Jank** | Exactly 0 dropped frames | Impeller AOT Pipeline Validator |
| **Tick Throughput** | >= 100,000 ticks/sec | Rust Benchmark Harness (`criterion`) |
| **Dart GC Pause Time** | 0.00 ms (Zero Heap Alloc) | Dart VM Timeline Profiler |
| **TouchID Response Latency** | < 250 milliseconds | Instruments Time Profiler |
| **Universal Binary Integrity**| ARM64 + x86_64 Validated | `vtool -show-build`, `lipo -info` |
| **Gatekeeper Notarization** | Accepted & Ticket Stapled | `spctl --assess --type execute -v` |
| **Zero-Fee Invariant** | Fee == 0.00% strictly | Automated Ledger Assertion Test Suite |

---

## 13. Architectural Sign-Off & Approvals

This specification serves as the binding architectural standard for the Growww / NBSE macOS Native Thick Client. All future releases must conform to the invariants and pipeline topologies defined herein.

*Approved by the Lead Client Architect, Systems Engineering, and Trading Core Infrastructure Committee.*
