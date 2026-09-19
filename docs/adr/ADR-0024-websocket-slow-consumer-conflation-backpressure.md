# ADR-0024: WebSocket Slow Consumer Drop-Tail Ring Buffers and Conflation Mode

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Market Data Architect, Lead Network Infrastructure Engineer  

---

## 1. Context
Fast-moving market data feeds (Level-2 book updates, trade ticks) can overwhelm mobile clients on constrained 4G/5G connections or slow networks, causing unbound memory growth on the API gateway/market-data cluster and high client latency.

---

## 2. Decision
Implement bounded 256-packet per-connection ring buffers with dual backpressure states:
1. **Drop-Tail Intermediate Conflation:** When ring buffer occupancy exceeds 80% (204 packets), intermediate Level-2 delta ticks are dropped and conflated into periodic (200ms) full book snapshots.
2. **Hard Disconnect on Desynchronization:** If client acknowledgement lag exceeds 2000ms or buffer overflows 256 packets, the connection is terminated with WebSocket closure code 1008 (Policy Violation) and code `ERR_STREAM_DESYNC`, forcing client reconnect and fresh snapshot bootstrap.

---

## 3. Consequences
- **Positive:** Protects gateway memory from unbounded growth; guarantees sub-millisecond dispatch latency for healthy consumers.
- **Trade-offs:** Slow clients experience coalesced updates during high volatility bursts.
