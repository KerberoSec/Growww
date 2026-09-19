# Flutter Mobile Architecture: Obsidian Dark Mode, CustomPainter & Haptics

**Specification ID:** SPEC-ARCH-023-MOB  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Mobile Frontend Architecture & High-Frequency UI Rendering  
**Owner:** Mobile Engineering & Client Experience Group  

---

## 1. Executive Summary & Design System
The Flutter application (`apps/growww_flutter`) delivers an ultra-smooth, responsive mobile trading experience for iOS, Android, and Desktop:
- **Visual Design**: Deep Obsidian theme (`#0B0E14` primary background, `#151A23` elevated cards, Neon Green `#00E676` for bids/profit, Coral Crimson `#FF3B30` for asks/loss, Radiant Amber `#FFB300` for Demo Paper Trading).
- **High-Frequency Rendering Invariant**: Guaranteed 120 FPS orderbook depth ladders and TradingView charts through background compute isolates and GPU-accelerated `CustomPainter`.
- **Sensory Haptics**: Rich micro-tactile feedback on order placement, cancellations, slider thresholds, and demo faucet claims.

---

## 2. Multi-Threaded Isolate Architecture & Zero-Jank UI

```
+----------------------------------------------------------------------------------------------------+
| FLUTTER CLIENT THREADING & COMPUTE ISOLATE PIPELINE                                                |
|                                                                                                    |
|  [ WebSocket Raw Stream ] ---> [ Background Dart Isolate (Worker) ]                                |
|   (50,000 Ticks/sec)           - Protobuf Binary Deserialization                                   |
|                                - L2 Orderbook Ladder Aggregation                                   |
|                                - 16.6ms Render Frame Conflation (60/120 FPS)                       |
|                                                |                                                   |
|                                                v                                                   |
|                               [ Immutable Render State Payload ]                                   |
|                                                |                                                   |
|                                                v                                                   |
|                               [ UI Main Thread (Skia / Impeller) ]                                 |
|                               - CustomPainter RepaintBoundary                                      |
|                               - Zero Garbage Collection Jitter                                     |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 Performance Optimizations:
- **RepaintBoundary Isolation**: Orderbook ladder widget is wrapped in `RepaintBoundary` so ticker updates do not trigger layout invalidation of the surrounding screen.
- **Haptic Feedback Patterns**:
  - Limit Order Placed: `HapticFeedback.lightImpact()`.
  - Market Order Filled: `HapticFeedback.mediumImpact()`.
  - Stop-Loss Triggered: `HapticFeedback.heavyImpact()`.
  - Demo Faucet Dispensed: Double haptic burst (`lightImpact()` followed by `mediumImpact()`).
