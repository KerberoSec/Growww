# ADR-0023: In-Place Order Quantity Reduction FIFO Priority Preservation

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Principal Matching Engine Architect, Head of Trading  

---

## 1. Context
Canceling and re-inserting an order when a trader reduces quantity destroys time priority in the matching engine queue, degrading market maker quoting incentives.

---

## 2. Decision
Implement an in-place `reduce_order_quantity` method in the Rust matching engine:
- Downward quantity modifications ($Q_2 < Q_1$) decrement price-level total volume without altering queue position or timestamp.
- Upward quantity changes ($Q_2 > Q_1$) or price modifications invalidate priority and move the order to the queue tail.

---

## 3. Consequences
- **Positive:** Encourages active liquidity provision; aligns with institutional market standards.
- **Trade-offs:** Requires doubly-linked list order tracking for in-place modifications.
