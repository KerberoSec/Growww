# ADR-0034: Advanced Order Types & Native In-Memory Algorithmic Slicing

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Matching Engine Architect, Head of Trading, Chief Technology Officer  

---

## 1. Context
Professional and institutional traders require advanced execution mechanisms (Iceberg slicing, OCO atomic coupling, Trailing Stop-Loss, TWAP, Post-Only) to execute complex trading strategies with minimal market impact. Handling these client-side introduces network latency and risk of partial execution breaks.

---

## 2. Decision
1. **In-Memory Trigger & Slicing Engine:**
   - Integrate a high-performance **Trigger Engine** and **Algorithmic Slicing Engine** directly alongside the Rust matching engine core.
   - Stop-Loss, Trailing Stop, and OCO triggers are evaluated synchronously in $< 5\mu\text{s}$ upon every trade execution tick.
2. **Native Iceberg Parent-Child Management:**
   - The matching engine tracks parent Iceberg states and spawns child tranches deterministically upon previous tranche fill.
   - Visible child tranches receive standard Price-Time priority; replenishment tranches join the queue tail with new timestamps.
3. **Strict Slippage Bounding on Market Orders:**
   - Market orders enforce an automatic default $\pm 1.0\%$ price band against opposite top-of-book, preventing unintended fills in thin books.
4. **Zero Surcharge Invariant:**
   - Algorithmic and conditional child orders strictly incur the single flat **0.00% (Zero Fee) platform fee** on executed notional with ₹0 additional algorithmic surcharge.

---

## 3. Consequences
- **Positive:** Sub-millisecond execution of algorithmic strategies; zero risk of client disconnection breaks during multi-leg OCO or Trailing Stop management.
- **Trade-offs:** Increases matching engine in-memory state tracking complexity.
