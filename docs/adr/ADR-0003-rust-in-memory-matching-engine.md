# ADR-0003: Rust for In-Memory Matching Engine and Audit Log

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Systems Architect, VP Engineering  

---

## 1. Context
The Central Limit Order Book (CLOB) matching engine and cryptographic audit logger require sub-10 microsecond execution latency, deterministic memory usage, and zero runtime garbage collection pauses under high concurrent load.

---

## 2. Decision
Implement `services/matching-engine` and `services/audit-service` in **Rust**, utilizing Tokio for asynchronous event processing, custom slab memory pools, and lock-free data structures.

---

## 3. Alternatives Considered
- **Go:** Rejected for the core matching loop due to non-deterministic GC pauses (20-100 microseconds) which violate fair price-time execution guarantees.
- **C++:** Provides raw performance but lacks Rust's compile-time memory safety, concurrency guarantees, and modern package ecosystem.

---

## 4. Consequences
- **Positive:** Deterministic sub-10μs order matching; memory safety without garbage collection; zero allocation on hot paths.
- **Trade-offs:** Steeper learning curve for general application developers; longer initial compile times in CI.
