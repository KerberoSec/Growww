# Universal Multi-Platform Testing & E2E Verification Matrix

**Specification ID:** SPEC-ARCH-045-TESTING-MATRIX  
**Document Version:** 1.0.0-PROD  
**Status:** Approved & Authoritative  
**Classification:** Quality Assurance, Multi-Platform Automated Testing & Acceptance Architecture  
**Target Environments:** Web Pro Terminal (Next.js), macOS Thick Client, Windows Thick Client, Linux Thick Client, iOS Mobile, Android Mobile  
**Last Updated:** September 2026  

---

## 1. Executive Summary & Quality Strategy

The Growww / NBSE trading platform mandates 100% automated end-to-end verification across all six target operating systems prior to any production deployment. Manual QA cannot reliably catch race conditions in 120 FPS canvas orderbooks, cross-window IPC deadlocks, or per-monitor DPI scaling glitches.

This document defines the automated test architecture, testing tools, acceptance criteria, and failure gates for every client platform.

---

## 2. Automated Multi-Platform Test Stack

```
+-------------------------------------------------------------------------------------------------------+
|                                  MULTI-PLATFORM TEST AUTOMATION STACK                                 |
|                                                                                                       |
|  [ Web Pro Terminal ]       ---> Playwright + Lighthouse CI + Web Worker Mock Server                  |
|  [ macOS Thick Client ]     ---> Flutter Driver + AppKit XCTest Native Harness + Metal Frame Tracing  |
|  [ Windows Thick Client ]   ---> WinAppDriver / WinUI 3 UI Automation + DirectX 12 PIX Benchmarking   |
|  [ Linux Thick Client ]     ---> Headless Weston / Xvfb + GTest GTK Harness + Vulkan Validation Layer |
|  [ Android Mobile Client ]  ---> Patrol / Espresso + Android Emulator Matrix (API 26-34)              |
|  [ iOS Mobile Client ]      ---> Patrol / XCUITest + iOS Simulator Matrix (iOS 15-18)                 |
+-------------------------------------------------------------------------------------------------------+
```

---

## 3. Platform-Specific Test Suites & Gate Criteria

### 3.1 Web Pro Trading Terminal (`growww_web`)
- **Automated Tool:** Playwright + Chromium / WebKit / Firefox.
- **Test Scenarios:**
  - **TC-WEB-01 (Multi-Window Popping):** Spawns 3 pop-out windows (`window.open`), emits a ticker change event via `BroadcastChannel`, and asserts all detached windows update within <=20ms.
  - **TC-WEB-02 (Web Worker Stream Ingestion):** Feeds 100,000 mock Protobuf depth packets/second into the worker; asserts the main React UI thread does not drop below 60 FPS (zero long tasks >50ms).
  - **TC-WEB-03 (CLS & Performance Budget):** Lighthouse CI audit verifies Cumulative Layout Shift equals `0.00` and First Input Delay (FID) is under 50ms.
  - **TC-WEB-04 (WebAuthn Passkey Mock):** Injects virtual authenticator via Chrome DevTools Protocol (CDP) and validates biometric trade execution.

### 3.2 macOS Native Thick Client (`apps/growww_flutter` Desktop)
- **Automated Tool:** Flutter Integration Test + XCTest AppKit harness.
- **Test Scenarios:**
  - **TC-MAC-01 (Metal 120 FPS Benchmark):** Renders 50-depth live orderbook ladder; verifies that `CVDisplayLink` records >=119.5 FPS on simulated ProMotion displays without GPU frame drops.
  - **TC-MAC-02 (Secure Enclave TouchID Mock):** Simulates `LAContext` biometric success and failure callbacks, ensuring private keys never leave the hardware enclave.
  - **TC-MAC-03 (Multi-Display Detachment):** Detaches an auxiliary window, moves it across simulated Display P3 monitors, and validates window restoration coordinates after restart.

### 3.3 Windows Native Thick Client (`apps/growww_flutter` Desktop)
- **Automated Tool:** WinAppDriver + MSBuild C++ test harness.
- **Test Scenarios:**
  - **TC-WIN-01 (DirectX 12 Refresh Rate):** Asserts frame pacing across high-refresh monitor profiles (144Hz, 240Hz, 360Hz) with adaptive tearing enabled.
  - **TC-WIN-02 (Windows Hello Integration):** Mocks `KeyCredentialManager` and verifies cryptographic trade payload signing.
  - **TC-WIN-03 (Per-Monitor DPI v2):** Drags window from 4K (150% scaling) to 1440p (100% scaling); asserts `WM_DPICHANGED` handles recalculate bounds without font blur or clipping.
  - **TC-WIN-04 (Global Hotkey Capture):** Focuses a dummy background application (Notepad), presses `Ctrl+Alt+Space`, and verifies that the `RegisterHotKey` listener cancels all active orders in <=15ms.

### 3.4 Linux Native Thick Client (`apps/growww_flutter` Desktop)
- **Automated Tool:** Headless Weston Wayland compositor / Xvfb + Vulkan Validation Layers.
- **Test Scenarios:**
  - **TC-LNX-01 (Vulkan Impeller Rendering):** Confirms zero Vulkan validation layer warnings or memory leaks during high-throughput tick ingestion.
  - **TC-LNX-02 (Secret Service DBus API):** Mocks `org.freedesktop.secrets` service and validates credential storage and retrieval.

### 3.5 Mobile Applications (Android & iOS)
- **Automated Tool:** Patrol test framework + physical device farm.
- **Test Scenarios:**
  - **TC-MOB-01 (15-Second Battery Suspension):** Backgrounds application; verifies WebSocket disconnects cleanly within 15 seconds, releasing Android `WakeLock` and iOS background tasks.
  - **TC-MOB-02 (Offline SQLite WAL Queue):** Cuts network connection (airplane mode), places 5 limit orders, restores network after 60 seconds, and verifies orders are dispatched with valid idempotency UUIDs.
  - **TC-MOB-03 (Tactile Haptic Triggers):** Asserts that selection, medium impact, and volatility alert haptic feedback methods are invoked via platform channels.

---

## 4. Universal Zero-Fee Invariant Assertion Gate

Every automated UI test across all platforms must execute the following non-negotiable assertion on every order form, confirmation modal, and execution blotter:

```python
# Universal Zero-Fee Test Assertion Rule (Pseudocode)
def assert_zero_fee_invariants(rendered_order_screen):
    assert "0.00%" in rendered_order_screen.trading_fee_text
    assert "₹0.00" in rendered_order_screen.platform_commission_text
    assert "0 Gas" in rendered_order_screen.paymaster_sponsorship_badge or "Sponsored" in rendered_order_screen.paymaster_sponsorship_badge
    assert rendered_order_screen.tds_deduction_amount == 0.00
```

Any build or pull request where a non-zero fee or missing gas sponsorship badge is rendered on any platform is automatically rejected by the CI/CD pipeline.
