# SRE Operational Runbook 036: Web Dockview Multi-Monitor Crash Recovery & Memory Management

**Runbook ID:** RUNBOOK-036-WEB-DOCKVIEW-RECOVERY  
**Severity Classification:** Tier 2 (Client Workstation State Recovery)  
**Document Version:** 1.0.0-PROD  
**Target Platform:** Web Pro Trading Terminal (`apps/growww_web`)  
**Target Browsers:** Google Chrome 120+, Microsoft Edge 120+, Apple Safari 17+, Mozilla Firefox 122+  
**Last Review:** September 2026  

---

## 1. Executive Summary

The Web Pro Trading Terminal allows traders to detach multiple panels into separate browser pop-out windows across multiple physical monitors via `window.open()`. Because each pop-out runs in its own browser window context while sharing state over `BroadcastChannel` and `SharedWorker`, browser crashes, abrupt tab closures, or memory exhaustion (OOM) on a single auxiliary window must never corrupt the user's trading session or drop resting limit orders.

---

## 2. Multi-Window Failure Modes & Recovery Architecture

```
+-------------------------------------------------------------------------------------------------------+
|                                POP-OUT CRASH CONTAINMENT ARCHITECTURE                                 |
|                                                                                                       |
|  [ Physical Display 1 ]              [ Physical Display 2 ]               [ Physical Display 3 ]      |
|  +---------------------------+       +---------------------------+        +-------------------------+ |
|  | MASTER BROWSER WINDOW     |       | DETACHED CHART WINDOW     |        | DETACHED DOM LADDER     | |
|  | - Core Order State        |       | - TradingView Canvas      |        | - High-Speed Canvas     | |
|  | - User Session Vault      |       | - Local Window Context    |        | - Local Window Context  | |
|  +---------------------------+       +---------------------------+        +-------------------------+ |
|               |                                    |                                   |              |
|               | <========== Heartbeat Sync via BroadcastChannel (1000ms) =============> |              |
|               |                                                                                       |
|               | [ DETECT POP-OUT DISCONNECT / CRASH ]                                                 |
|               | - Missing 3 consecutive heartbeats (3000ms)                                           |
|               | - Pop-out reference `window.closed === true`                                          |
|               |                                                                                       |
|               v                                                                                       |
|  [ AUTOMATIC LOCAL FALLBACK ]                                                                         |
|  1. Master window automatically restores collapsed panel into internal Dockview grid.                  |
|  2. Unsubmitted draft orders preserved in localStorage recovery stash.                               |
|  3. Audio alert + subtle toast: "Chart window detached monitor lost. Docked to main workstation."    |
+-------------------------------------------------------------------------------------------------------+
```

---

## 3. Client Memory Leak Prevention Directives

During all-day trading sessions (8+ continuous hours), JavaScript single-page applications can accumulate detached DOM elements and memory leaks:
1. **TradingView Canvas Destruction:**
   - On panel unmount or pop-out window close, the chart widget must invoke `.remove()` to clean up WebGL contexts and framebuffers.
2. **WebSocket Message Event Listener Teardown:**
   - Unsubscribe handlers must remove all event listeners from the `BroadcastChannel` and release object URLs (`URL.revokeObjectURL`).
3. **SharedArrayBuffer Ring Buffer Clamping:**
   - Pre-allocated 4MB memory buffer; never dynamically resized at runtime.
4. **Periodic Garbage Collection Hint:**
   - In Chromium browsers, memory footprint is monitored via `performance.memory.usedJSHeapSize`. If heap exceeds 500MB, non-essential off-screen DOM ladders are recycled.

---

## 4. Operational Recovery Procedures

### Scenario A: Trader Unplugs Secondary External Monitor
1. The browser operating system detects display disconnection (`window.screenX / screenY` coordinates become invalid).
2. The pop-out window auto-repositions to primary display coordinates `(x: 50, y: 50)` within 500ms.
3. If window positioning fails, the Master window re-absorbs the panel into the main docking layout.

### Scenario B: Accidental Browser Tab Close
1. User accidentally closes the detached DOM ladder pop-out.
2. Master window receives `beforeunload` message via `BroadcastChannel`.
3. Master UI renders an immediate "Re-open Detached DOM Ladder" floating action button in the bottom right corner.
4. Active working orders remain safely resting on the central matching engine.
