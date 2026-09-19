# ADR-0016: Lock-Free SPSC Sequencer Ring and NUMA Core Pinning

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Systems Architect, HFT Infrastructure Lead  

---

## 1. Context
Multi-threaded ingress gateways attempting to update atomic sequencer cursors concurrently suffer from CPU cacheline bouncing (`MESI` invalidation), causing order ingress latency spikes exceeding 15ms.

---

## 2. Decision
Implement an **LMAX Disruptor-style Single-Writer Sequencer** using lock-free Single-Producer Single-Consumer (SPSC) ring buffers:
- Network worker threads push raw order payloads to dedicated SPSC input queues.
- A single Sequencer thread (pinned to CPU Core 2 via `isolcpus`) polls input queues in a busy-wait loop, assigns monotonic 64-bit sequence numbers, and dispatches to the matching core (Core 3, same NUMA node 0).
- All pointer structures use 64-byte cacheline alignment (`#[repr(align(64))]`).

---

## 3. Consequences
- **Positive:** Reduces sequencer latency from $15\text{ms}$ down to $< 1.8\mu\text{s}$; eliminates mutex lock contention.
- **Trade-offs:** Requires dedicated CPU core allocation.
