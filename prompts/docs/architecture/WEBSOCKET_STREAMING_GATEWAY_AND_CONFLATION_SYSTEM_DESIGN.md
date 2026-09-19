# WebSocket Streaming Gateway & Conflation Architecture Specification

**Specification ID:** SPEC-ARCH-051-WS  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Real-Time Market Data Distribution & Streaming Infrastructure  
**Owner:** Market Data & Edge Distribution Group  

---

## 1. Executive Summary & Throughput Targets
The WebSocket Streaming Gateway distributes real-time market data (BTC/USDT Level-2 orderbook ladders, public trade prints, 24-hour tickers, and user-specific order updates) to over **1,000,000 concurrent client sessions**.
- **Edge Conflation Window**: 50ms aggregation windows for retail L2 depth, reducing outbound bandwidth by 92% during extreme market volatility.
- **Institutional Level-3 Feed**: Un-conflated tick-by-tick binary SBE (Simple Binary Encoding) streams for direct DMA market makers.
- **Zero-Allocation Ring Buffers**: Lock-free circular queues written in Go and Rust delivering sub-millisecond fanout latency.

---

## 2. Multi-Tier Conflation Engine

```
+----------------------------------------------------------------------------------------------------+
| HIGH-FANOUT WEBSOCKET ARCHITECTURE & CONFLATION TOPOLOGY                                           |
|                                                                                                    |
|  [ Matching Engine ] ---> [ IPC Shared Memory Ring ] ---> [ Internal L2 Depth Aggregator ]         |
|   (50,000 Fills/sec)                                                    |                          |
|                                                                         v                          |
|                                                          [ 50ms Conflation Bucketing ]             |
|                                                                         |                          |
|                               +-----------------------------------------+                          |
|                               |                                                                    |
|                               v                                         v                          |
|                  [ Public Conflated Channel ]             [ Private Execution Channel ]            |
|                  - 20-Level Depth Snapshots               - Instantaneous Trade Fill Receipts      |
|                  - Max 20 Updates/sec per Pair            - Un-conflated Direct Delivery           |
|                               |                                         |                          |
|                               +--------------------+--------------------+                          |
|                                                    |                                               |
|                                                    v                                               |
|                                      [ Edge Envoy Gateway Swarm ]                                  |
|                                      (1,000,000 WebSockets Connected)                              |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 Conflation Algorithms:
1. **Price-Level Delta Conflation**:
   - Within each 50ms window, multiple updates to the same price level are coalesced into a single delta `(price, new_aggregate_volume)`.
   - If volume transitions to `0`, a `DELETE` instruction is emitted.
2. **Backpressure & Slow Consumer Mitigation**:
   - Each client connection possesses a bounded 4MB TCP write buffer.
   - If a slow client drops below 10 KB/s throughput and client buffer utilization reaches 90%, the gateway issues a `DROP_SLOW_CONSUMER` event and terminates the TCP session cleanly, preventing memory exhaustion on the gateway edge.

---

## 3. Protocol Serialization & Wire Standards
- **Web & Mobile Clients**: Compact JSON over WebSocket or Protobuf binary framing.
- **Institutional Colocation Terminals**: FIX 5.0 SP2 and binary Simple Binary Encoding (SBE) streams over kernel-bypass Solarflare Onload TCP.
