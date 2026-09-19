# Institutional FIX 5.0 SP2 Gateway & Drop Copy Architecture Specification

**Specification ID:** SPEC-ARCH-027-FIX  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** High-Frequency Protocol & Institutional Colocation  
**Owner:** Institutional Connectivity & Quantitative Trading Solutions  

---

## 1. Executive Summary & Performance Standards
The Institutional FIX Gateway provides direct market access (DMA) via the Financial Information eXchange (FIX) 5.0 SP2 protocol for quantitative trading firms, high-frequency market makers, and institutional custodians:
- **Sub-Microsecond Parsing**: Zero-allocation QuickFIX/C++ and Rust parsers delivering order ingestion latency $< 2.5\mu\text{s}$.
- **Isolated Drop Copy Stream**: Separate, real-time read-only FIX Drop Copy connection streaming trade execution reports for post-trade clearing and risk surveillance.
- **Universal Zero-Fee Invariant**: **0.00% trading fees** (No fee at all for institutional participants).

---

## 2. Session Protocol & Sequence Recovery

```
+----------------------------------------------------------------------------------------------------+
| HIGH-FREQUENCY FIX 5.0 SP2 PROTOCOL & SEQUENCE RECOVERY PIPELINE                                   |
|                                                                                                    |
|  [ Client Fix Engine ] ---> [ Kernel-Bypass Solarflare Onload TCP Interface ]                      |
|                                                |                                                   |
|                                                v                                                   |
|                                  [ Lock-Free SPSC FIX Parser ]                                     |
|                                                |                                                   |
|                                                v                                                   |
|                       [ Sequence Number Validation (MsgSeqNum 34) ]                                |
|                                 |                             |                                    |
|                       IN-ORDER  |                             | GAP DETECTED                       |
|                                 v                             v                                    |
|             [ Route to Matching Engine SPSC ]    [ Emit ResendRequest (MsgType 2) ]                |
|                                                  - Replay missing messages from in-memory cache    |
+----------------------------------------------------------------------------------------------------+
```
