# ADR-0038: Institutional FIX Protocol 4.4 / 5.0 SP2 & Binary Feeds

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Head of Institutional Sales, Principal Low-Latency Systems Architect  

---

## 1. Context
Institutional trading desks and prime brokers require standardized Financial Information eXchange (FIX) protocol sessions and ultra-low latency binary feeds for automated market making and order execution.

---

## 2. Decision
1. **FIX 4.4 / 5.0 SP2 Engine:** Provide standard FIX gateways supporting New Order Single (`D`), Order Cancel (`F`), Execution Report (`8`), and Drop Copy monitoring feeds.
2. **Binary Protocols:** Offer high-speed Binary OUCH for sub-10μs order ingress and Binary ITCH for multicast Level-3 tick distribution.

---

## 3. Consequences
- **Positive:** Direct compatibility with institutional Execution Management Systems (EMS) and Order Management Systems (OMS).
- **Trade-offs:** Requires dedicated FIX gateway session persistence and sequence recovery daemons.
