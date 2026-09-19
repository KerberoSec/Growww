# L3 In-Memory Matching Engine Internals & NUMA Pinning Specification

**Specification ID:** SPEC-ARCH-029-L3ME  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Core Execution Architecture & Ultra-Low Latency Engineering  
**Owner:** Core Matching Engine & Systems Architecture Group  

---

## 1. Executive Summary & Latency Performance Budgets
The core order matching engine is implemented in Rust 1.76, operating on a single-writer thread pinned to a dedicated CPU core via thread affinity and NUMA memory node binding:
- **Universal Zero-Fee Execution**: Strictly **0.00% fee** across all matched orders (No fee at all for Maker or Taker).
- **Latency Targets**: p50 $< 1.0\mu\text{s}$, p95 $< 2.5\mu\text{s}$, p99 $< 5.0\mu\text{s}$.
- **Throughput Budget**: $> 1,000,000$ operations per second sustained per trading pair.
- **Zero Dynamic Memory Allocation**: Uses a pre-allocated slab memory allocator; zero heap allocations during order placement, matching, or cancellation.

---

## 2. Core Data Structures & NUMA Hardware Pinning

```
+----------------------------------------------------------------------------------------------------+
| ULTRA-LOW LATENCY MATCHING ENGINE INTERNAL ARCHITECTURE                                            |
|                                                                                                    |
|  [ Inbound SPSC Lock-Free Queue ] (Cache-Line Aligned Ring Buffer)                                 |
|                 |                                                                                  |
|                 v                                                                                  |
|  [ Single-Writer Thread (CPU Core 4, NUMA Node 0) ]                                                |
|                 |                                                                                  |
|                 +-----------------------------------+-----------------------------------+          |
|                 |                                                                       |          |
|                 v                                                                       v          |
|   [ BTreeMap<Price, PriceLevel> ] (Bids)                  [ BTreeMap<Price, PriceLevel> ] (Asks)   |
|   - O(log N) price insertion                              - O(log N) price insertion               |
|   - Pointer to Best Bid (O(1) peek)                       - Pointer to Best Ask (O(1) peek)        |
|                 |                                                                       |          |
|                 v                                                                       v          |
|   [ Intrusive Doubly-Linked FIFO Queue ]                  [ Intrusive Doubly-Linked FIFO Queue ]   |
|   - O(1) order insertion at tail                          - O(1) order insertion at tail           |
|   - O(1) order cancellation by pointer                    - O(1) order cancellation by pointer     |
|                 |                                                                       |          |
|                 +-----------------------------------+-----------------------------------+          |
|                                                     |                                              |
|                                                     v                                              |
|                                   [ Outbound SPSC Trade Execution Ring ]                           |
|                                   - Emitted to TigerBeetle and Kafka                               |
+----------------------------------------------------------------------------------------------------+
```

### 2.1 Hardware Optimization Flags:
- Core Isolation: `isolcpus=4-7 nohz_full=4-7 rcu_nocbs=4-7` kernel boot parameters.
- Memory Policy: `numactl --cpunodebind=0 --membind=0` ensuring zero cross-socket QPI/UPI bus traversals.
- Cache Alignment: Structs aligned to 64 bytes (`#[repr(align(64))]`) to eliminate false sharing.
