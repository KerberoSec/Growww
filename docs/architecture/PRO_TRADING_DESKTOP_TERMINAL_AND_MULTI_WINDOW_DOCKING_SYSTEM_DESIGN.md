# Pro Trading Web & Desktop Terminal Multi-Window Docking Specification

**Specification ID:** SPEC-ARCH-041-PRO  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Web & Desktop Trading Terminal Architecture  
**Owner:** Web Applications & Terminal Experience Group  

---

## 1. Executive Summary & Multi-Window Architecture
The Pro Web and Desktop Trading Terminal (`apps/growww_web`) provides an institutional-grade, multi-monitor customizable docking workspace for pro traders:
- **Golden Layout / FlexLayout Integration**: Completely customizable docking layout with detachable pop-out windows for external multi-monitor setups.
- **Web Workers & SharedArrayBuffer**: Offloads WebSocket decoding and L2 orderbook calculation to dedicated Web Workers, ensuring the browser main thread remains 100% fluid at 120 FPS.
- **Keyboard Shortcuts & Hotkey Fast Trading**: Zero-latency order execution via single-key commands (e.g. `B` for Buy, `S` for Sell, `Space` for Cancel All).

---

## 2. Web Worker & State Pipeline

```
+----------------------------------------------------------------------------------------------------+
| MULTI-THREADED PRO WEB TRADING TERMINAL ARCHITECTURE                                               |
|                                                                                                    |
|  [ High-Throughput WSS ] ---> [ Dedicated Web Worker ]                                             |
|                               - Binary SBE / Protobuf Parser                                       |
|                               - SharedArrayBuffer Atomic State Updates                             |
|                                                |                                                   |
|                                                v                                                   |
|                                  [ SharedArrayBuffer Memory ]                                      |
|                                                |                                                   |
|                                                v                                                   |
|                               [ Main Window (React 18 Concurrent) ]                                |
|                               - Canvas-based High-Speed Order Ladder                               |
|                               - Fluid TradingView Lightweight Charts                               |
|                                                |                                                   |
|                         +----------------------+----------------------+                            |
|                         |                                             |                            |
|                         v                                             v                            |
|               [ Detached Monitor 1 ]                        [ Detached Monitor 2 ]                 |
|               (Pop-Out Depth Ladder)                        (Dedicated 4K Chart Window)            |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 Multi-Window State Synchronization:
- Detached browser windows communicate via `BroadcastChannel` and `SharedWorker`, maintaining sub-millisecond state coherence across physical monitors without server re-fetching.
