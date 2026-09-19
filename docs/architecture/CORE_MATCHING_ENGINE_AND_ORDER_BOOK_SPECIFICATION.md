# Core Matching Engine & Order Book Architecture Specification

**Specification ID:** SPEC-ARCH-008-ENG  
**Document Version:** 2.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** High-Frequency Trading & Systems Engineering Group  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  

---

## 1. Integer Arithmetic Precision & Numeric Overflow Prevention

### 1.1 Fixed-Point Numeric Invariant
To guarantee zero floating-point non-determinism, price-rounding errors, and integer overflow across polyglot boundaries (Rust, Go, Python, Solidity), all prices, quantities, and notional calculations use fixed-point scaled integers:

$$\text{Price} \in \mathbb{Z}_{\ge 0}, \quad \text{Scale} = 10^4 \quad (1 \text{ tick} = 0.0001 \text{ INR} = 1 \text{ weINR})$$
$$\text{Quantity} \in \mathbb{Z}_{\ge 0}, \quad \text{Scale} = 10^6 \quad (1 \text{ unit} = 0.000001 \text{ share})$$
$$\text{Notional} = \frac{\text{Price} \times \text{Quantity}}{10^6} \quad (\text{Scaled in } 10^4 \text{ weINR})$$

### 1.2 128-Bit Intermediate Multiplication in Rust
In `services/matching-engine`, all multiplication operations execute using 128-bit unsigned integers (`u128`) with checked arithmetic before casting to 64-bit integer values:

```rust
#[inline(always)]
pub fn calculate_notional_weinr(price_weinr: u64, quantity_micro: u64) -> Result<u64, MatchingEngineError> {
    let p = price_weinr as u128;
    let q = quantity_micro as u128;
    
    // Multiply in 128-bit space to prevent overflow
    let product = p.checked_mul(q).ok_or(MatchingEngineError::NumericOverflow)?;
    
    // Scale down by quantity precision factor (1,000,000)
    let notional = product / 1_000_000u128;
    
    if notional > u64::MAX as u128 {
        return Err(MatchingEngineError::NumericOverflow);
    }
    
    Ok(notional as u64)
}
```

---

## 2. Ingress Sequencer Ring & NUMA Cache Pinning (ADR-0016)

To eliminate atomic CPU cacheline bouncing (`MESI` protocol invalidation) across network worker threads and achieve deterministic sub-10 microsecond execution:

```
+----------------------------------------------------------------------------------------------------+
| LMAX-STYLE SINGLE-WRITER SPSC INGRESS SEQUENCER TOPOLOGY                                           |
|                                                                                                    |
|  [Network Worker Threads (Cores 4-15)]                                                             |
|           |                                                                                        |
|           v (Lock-Free SPSC Ring Buffers)                                                          |
|  +----------------------------------------------------------------------------------------------+  |
|  | Single-Writer Sequencer Thread (Pinned to Core 2, isolcpus, nohz_full)                       |  |
|  | - Assigns globally monotonic 64-bit Sequence Number: Seq(O)                                  |  |
|  | - Evaluates Spread Gating & Self-Trade Prevention (STP) in L1 Cache                          |  |
|  +----------------------------------------------------------------------------------------------+  |
|           |                                                                                        |
|           v (Direct L1/L2 Cache Transfer)                                                          |
|  +----------------------------------------------------------------------------------------------+  |
|  | Matching Engine Core (Pinned to Core 3, Same NUMA Node 0)                                    |  |
|  | - Executes Price-Time Priority Matching in Sub-10μs                                          |  |
|  | - Emits Trade Match to Journal WAL Ring Buffer                                               |  |
|  +----------------------------------------------------------------------------------------------+  |
+----------------------------------------------------------------------------------------------------+
```

- **Data Layout:** Sequencer cursor and matching engine pointers use 64-byte cacheline alignment (`#[repr(align(64))]`) to prevent false sharing.

---

## 3. Real-Time Call Auction IEP Matching Algorithm (ADR-0019)

During market open and circuit breaker resumption, the Indicative Equilibrium Price (IEP) is calculated in $O(\log K)$ time using **Dual Fenwick Trees (Binary Indexed Trees)** for cumulative Bid and Ask demand curves:

```rust
pub struct FenwickDemandCurve {
    tree: Vec<u64>,
    max_price_ticks: usize,
}

impl FenwickDemandCurve {
    pub fn update_volume(&mut self, price_tick: usize, delta_qty: i64) {
        let mut idx = price_tick;
        while idx < self.tree.len() {
            if delta_qty >= 0 {
                self.tree[idx] += delta_qty as u64;
            } else {
                self.tree[idx] -= delta_qty.unsigned_abs();
            }
            idx += idx & (!idx + 1);
        }
    }

    pub fn query_cumulative(&self, price_tick: usize) -> u64 {
        let mut sum = 0u64;
        let mut idx = price_tick;
        while idx > 0 {
            sum += self.tree[idx];
            idx -= idx & (!idx + 1);
        }
        sum
    }
}
```

### IEP Intersection Search:
1. Binary search finds the intersection $P_{\text{IEP}}$ maximizing $\min(\text{CumBid}(P), \text{CumAsk}(P))$.
2. Ties resolve by minimum order imbalance $|\text{CumBid}(P) - \text{CumAsk}(P)|$.
3. Secondary ties resolve by closest distance to previous closing reference price.

---

## 4. Advanced Order Types & State Transitions

### 4.1 Stop-Loss, Stop-Limit & Trailing Stop Execution Engine (TRD-05)
1. **Trigger Evaluation:** Evaluated against the Last Traded Price (LTP) or median of the last 5 execution ticks, NOT the Best Bid/Ask quote, preventing phantom liquidations during illiquid quote spread widening.
2. **Spread Ceiling Assertion:** Trailing stop-loss triggers execute only when $\text{BestAsk} - \text{BestBid} \le \text{MaxSpreadThreshold}$.
3. **Circuit Halt Isolation:** When an instrument enters `CIRCUIT_HALT` or `MAINTENANCE`, stop triggers are **quarantined**. Stop orders do not activate during price gaps caused by market halts; they evaluate only against the equilibrium price established during the subsequent **Pre-Open Call Auction Window**.

### 4.2 In-Place Order Quantity Reduction & Priority Preservation (TRD-06)
- **Quantity Reduction Invariant (ADR-0023):** When a trader reduces an order's open quantity from $Q_1$ to $Q_2$ ($Q_2 < Q_1$) at the same price, the engine modifies the quantity in-place and decrements total price-level volume **without altering the order's position or timestamp in the FIFO priority queue**.
- **Quantity Increase / Price Change:** Any quantity increase ($Q_2 > Q_1$) or price modification invalidates priority, assigning a new sequence number and moving the order to the tail of the queue.

### 4.3 Passive Pegged Orders & Feedback Loop Prevention (TRD-07)
- **Non-Quote-Forming Invariant:** Midpoint and Primary Pegged orders rest strictly on an internal non-display book and do NOT alter the public Best Bid/Offer quote, mathematically eliminating recursive pricing feedback loops.
- **Cooldown Throttle:** Re-pegging executes only on actual external trades or manual quote updates, with a minimum 50ms cooldown timer per pegged order.

---

## 5. Cold-Start Initialization & Kafka WAL Snapshot Re-Synchronization (RUNBOOK-09)

```
+----------------------------------------------------------------------------------------------------+
| MATCHING ENGINE COLD-START RECOVERY STATE MACHINE                                                  |
|                                                                                                    |
|  [SHUTDOWN / COLD]                                                                                 |
|         |                                                                                          |
|         v                                                                                          |
|  [LOAD_LATEST_MMAP_SNAPSHOT] ---> Read binary memory dump from disk (MS_SYNC)                      |
|         |                                                                                          |
|         v                                                                                          |
|  [REPLAY_KAFKA_WAL]          ---> Stream growww.order.events.v1 from snapshot offset to high-water |
|         |                                                                                          |
|         v                                                                                          |
|  [CROSS_CHECK_CHECKSUM]      ---> Validate Merkle root against audit-service journal               |
|         |                                                                                          |
|         v                                                                                          |
|  [OPEN_FOR_INCOMING_ORDERS]  ---> Bind TCP/WebSocket listeners and accept gateway traffic          |
+----------------------------------------------------------------------------------------------------+
```

Cold-start replay achieves 10,000,000 order book events replayed in $< 1.2\text{ seconds}$ via mmap direct memory mapping.
