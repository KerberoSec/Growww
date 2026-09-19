# ADR-0014: Salted Kafka Partition Hashing for High-Volume Instruments

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Infrastructure Architect, Data Platform Lead  

---

## 1. Context
During major market events (e.g. index rebalancing, earnings releases for bellwether stocks like `RELIANCE`), single-key partitioning (`key = instrument_id`) creates severe traffic hotspots on individual Kafka partition brokers, leading to consumer lag spikes.

---

## 2. Decision
Implement **Dynamic Salted Sub-Partitioning**:
- Standard volume instruments partition strictly by `instrument_id`.
- High-volume flagged instruments apply a deterministic account salt (`Murmur3_32(account_id) % 8`), spreading the order stream across 8 dedicated sub-partitions while preserving in-order delivery per user.

---

## 3. Consequences
- **Positive:** Scales single-stock peak throughput by $8\times$; prevents Kafka partition saturation.
- **Trade-offs:** Matching engine consumer groups must aggregate all 8 sub-partitions per instrument.
