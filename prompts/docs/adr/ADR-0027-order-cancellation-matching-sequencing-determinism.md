# ADR-0027: Order Cancellation and Matching Sequencing Determinism

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Matching Engine Architect, Head of Trading  

---

## 1. Context
In an asynchronous distributed architecture, an `OrderCancelRequest` in flight over Kafka can race with a match event generated concurrently in the matching engine memory. If the Order Service acknowledges cancellation before the match event is processed, the client receives false cancellation confirmation while their order has been executed, producing phantom state breaks.

---

## 2. Decision
1. **Single-Writer Sequencing Authority:** The Rust matching engine's single-writer SPSC ingress queue is the sole authoritative sequencer for both new order placements and cancellations.
2. **Synchronous Execution / Rejection:** 
   - If the cancellation executes before matching, the engine emits `ORDER_CANCELLED`.
   - If the order has already matched (partially or fully) when the cancel command reaches the head of the sequencer queue, the engine cancels any unexecuted remaining quantity and emits `ORDER_CANCELLED_WITH_FILL` or rejects with `ERR_ORDER_ALREADY_FILLED`.
3. **No Optimistic Cancel ACKs:** The API Gateway and Order Service strictly do NOT acknowledge cancellation to the client until the engine's deterministic response event is received.

---

## 3. Consequences
- **Positive:** Completely eliminates race conditions between order cancellation and execution fills.
- **Trade-offs:** Client cancel round-trip latency includes the sequencer traversal time (typically $< 15\mu\text{s}$).
