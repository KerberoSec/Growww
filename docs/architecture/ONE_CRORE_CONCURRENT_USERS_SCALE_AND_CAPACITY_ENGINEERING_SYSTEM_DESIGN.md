# One Crore (10 Million) Concurrent Users Scale & Capacity Engineering

**Specification ID:** SPEC-ARCH-037-SCALE  
**Document Version:** 2.0.0-PROD  
**Status:** Approved  
**Classification:** Hyper-Scale Infrastructure & Capacity Engineering  
**Owner:** Cloud Infrastructure & Distributed Systems Reliability Group  

---

## 1. Executive Summary & Capacity Objectives
This specification establishes the capacity models, edge network topology, and distributed caching architecture required to sustain **10,000,000 (1 Crore) concurrent connected users** and **100,000 active trading operations per second**:
- **Edge WebSocket Gateway Swarm**: 100 Envoy gateway pods each handling 100,000 concurrent long-lived connections using kernel `epoll` with bounded memory caps.
- **Network Bandwidth Optimization**: 50ms L2 depth conflation reducing edge outbound traffic by 92% (from 50 Gbps to ~4 Gbps).
- **Zero-Fee Economic Scalability**: The 0.00% fee architecture eliminates complex billing and fee calculation loops, saving 25% CPU overhead on the transactional hot path.

---

## 2. Tiered Hyper-Scale Topology

```
+----------------------------------------------------------------------------------------------------+
| ONE CRORE (10,000,000) CONCURRENT SESSIONS CAPACITY TOPOLOGY                                       |
|                                                                                                    |
|  [ 10,000,000 Connected Mobile & Web Traders ]                                                     |
|                           |                                                                        |
|                           v                                                                        |
|  [ AWS Global Accelerator / Anycast DNS ]                                                          |
|                           |                                                                        |
|                           v                                                                        |
|  [ 100 Envoy Edge Gateway Pods (100,000 Connections/Pod) ]                                         |
|  - TLS Termination & HTTP/2 / WebSocket Framing                                                    |
|  - 50ms Delta Conflation per Subscription                                                          |
|                           |                                                                        |
|                           v                                                                        |
|  [ Internal Service Mesh (Envoy + SPIRE mTLS) ]                                                    |
|                           |                                                                        |
|                           v                                                                        |
|  [ Apache Kafka Cluster (32 Partitions per Pair, 500,000 msgs/sec) ]                               |
|                           |                                                                        |
|                           v                                                                        |
|  [ Sharded Rust Matching Engines (1 Pair per Engine Instance) ]                                    |
|                           |                                                                        |
|                           v                                                                        |
|  [ 6-Node TigerBeetle Financial Ledger Cluster ]                                                   |
+----------------------------------------------------------------------------------------------------+
```
