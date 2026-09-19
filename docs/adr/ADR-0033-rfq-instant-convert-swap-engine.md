# ADR-0033: Request-For-Quote (RFQ) Instant Convert & Zero-Slippage Swap Engine

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Head of Product, Principal Matching Engine Architect, Lead Frontend Engineer  

---

## 1. Context
Retail users frequently prefer a simple, single-click "Convert / Instant Swap" experience (e.g. USDT to INR, BTC to ETH) without navigating Level-2 order books, understanding bid-ask spreads, or managing limit order execution states.

---

## 2. Decision
1. **Zero-Slippage RFQ Architecture:**
   - Client requests a 2-way quote: `POST /api/v1/convert/quote` (`from_asset`, `to_asset`, `amount`).
   - The RFQ pricing engine constructs an instantaneous synthetic quote aggregated across internal matching engine order books and institutional market makers.
   - The quote guarantees the exact payout amount and is locked with a cryptographic signature valid for **6.00 seconds**.
2. **Atomic Execution:**
   - When the user confirms within the 6-second window, the transaction executes atomically against internal inventory / market maker quotes with zero slippage.
   - The standard flat **0.00% (Zero Fee) platform fee** is transparently included in the quoted conversion price with ₹0 additional spread markup.

---

## 3. Consequences
- **Positive:** Delivers a 1-click conversion experience; guarantees zero slippage for retail users.
- **Trade-offs:** The platform holds price risk during the 6-second quote window, mitigated by internal hedging algorithms.
