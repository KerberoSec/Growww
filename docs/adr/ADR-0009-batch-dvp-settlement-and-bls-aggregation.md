# ADR-0009: Batch DvP Settlement and BLS Signature Aggregation

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Cryptographer, Lead Blockchain Architect  

---

## 1. Context
Individual on-chain Delivery-versus-Payment (DvP) transactions strain CloudHSM hardware signing limits (1,000-2,000 ECDSA signatures/sec per appliance) under peak order book execution throughput (10,000 matches/sec).

---

## 2. Decision
Implement **Batch Multi-Trade DvP Settlements** and **BLS12-381 Aggregate Signature Verification**:
- `settlement-service` aggregates 50 to 100 matched trades into a single atomic smart contract transaction payload.
- Relayers sign batch state roots rather than individual fills, reducing CloudHSM signing operations by $98\%$.
- Smart contracts verify aggregate signatures in constant time, dramatically lowering on-chain gas overhead and eliminating relayer nonce contention.

---

## 3. Consequences
- **Positive:** Scales settlement throughput beyond 10,000 trades/sec within standard CloudHSM capacity; lowers gas costs per fill.
- **Trade-offs:** Adds 50ms batching window prior to on-chain dispatch.
