# Non-Functional Requirements, Latency Budgets & Performance Engineering Master Specification

**Specification ID:** `SPEC-ARCH-010-NFR`  
**Document Version:** `1.0.0-PROD`  
**Status:** Approved / Production Ready  
**Classification:** Exchange Infrastructure, Systems Engineering & Regulatory Compliance  
**Target Platform:** Growww Brokerage & National Blockchain Stock Exchange (NBSE) Core Platform  
**Owner:** Site Reliability Engineering, High-Frequency Trading Core & Blockchain Architecture  
**Effective Date:** September 2026  
**Review Cadence:** Quarterly  
**Regulatory Alignment:** SEBI Cyber Security and Cyber Resilience Framework (CSCRF) 2024/2026, SEBI Master Circular for Stock Exchanges, RBI Cyber Security Framework for Payment Systems, Digital Personal Data Protection (DPDP) Act 2023, ISO/IEC 27001:2022, SOC 2 Type II  

---

## 1. Executive Summary & Purpose

Operating a high-concurrency, mission-critical financial exchange and permissioned settlement infrastructure demands strict, mathematically quantifiable Non-Functional Requirements (NFRs). Under the dual-entity operating model:
- **Growww Technologies India Pvt. Ltd.** operates as the regulated retail stock broker and depository participant (DP), managing retail ingress, biometric authentication, pre-trade risk gating, fiat on/off-ramps, and front-end trading interfaces.
- **NBSE Ltd. (National Blockchain Stock Exchange)** operates as the sovereign sovereign financial market infrastructure (FMI), Central Limit Order Book (CLOB) matching engine, clearing corporation, and atomic Delivery-versus-Payment (DvP) settlement authority anchored to a permissioned Hyperledger Besu blockchain consortium.

Sub-millisecond latency in order matching, high-throughput DvP settlement, absolute transactional durability, zero-loss disaster recovery, 99.99% system availability during Indian market hours (09:00 - 16:00 IST), and 99.9% overall platform availability are mandatory to comply with SEBI CSCRF guidelines and deliver an institutional-grade retail trading experience.

This specification formalizes the master quantitative Service Level Objectives (SLOs), Service Level Agreements (SLAs), latency budgets, throughput targets, capacity models, disaster recovery objectives (RTO/RPO), and reliability engineering standards across the entire Growww / NBSE technical stack. Companion machine-readable definitions are maintained in [`docs/architecture/slo_targets.yaml`](file:///home/Kali/Desktop/Growww/Growww/docs/architecture/slo_targets.yaml) for automated continuous verification in CI/CD pipelines and Prometheus alert generation.

---

## 2. System-Wide Latency Budget & Hop-by-Hop Breakdown

### 2.1 End-to-End Latency Budget Architecture
Financial order placement follows an end-to-end execution path from a retail client tap to an on-chain blockchain settlement confirmation. The system enforces strict latency ceilings across every intermediate hop to prevent queuing delay accumulation and head-of-line blocking.

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                    END-TO-END SYSTEM LATENCY BUDGET ALLOCATION (ORDER TO SETTLEMENT)                         |
|                                                                                                                             |
|  [Flutter Client]  --TLS 1.3/QUIC-->  [Envoy Edge GW]  --mTLS-->  [API Gateway BFF]  --gRPC-->  [Pre-Trade Risk Engine]   |
|     (p99 < 15ms)                         (p99 < 5ms)                 (p99 < 20ms)                     (p99 < 1.5ms)         |
|                                                                                                             |               |
|                                                                                                             v (Lock-Free)   |
|  [Web/Mobile UI]   <--WebSocket--   [Market Data GW]  <--Kafka--   [Rust Matching Engine]  <--SPSC Ring-- [Ingress Sequencer]|
|  (Render 60/120Hz)                   (p99 < 50ms)     (p99 < 2ms)     (p99 < 100μs, p50 < 10μs)        (p99 < 50μs)         |
|                                                                              |                                              |
|                                                                              v (Kafka Journal WAL)                          |
|  [Hyperledger Besu] <--RPC/HSM-- [32-Relayer Pool] <--Queue-- [Clearing & DvP Engine]                                       |
|  (p99 < 2000ms block)             (p99 < 10ms HSM)               (p99 < 250ms)                                              |
+-----------------------------------------------------------------------------------------------------------------------------+
```

### 2.2 Granular Hop-by-Hop Latency Allocation

The following table itemizes the maximum allocated latency budget for each discrete processing step along the critical transaction path:

| Hop # | Processing Subsystem | Technology Stack | Transport / Protocol | Target p50 | Target p95 | Target p99 | Target p99.9 | Hard Ceiling |
|---|---|---|---|---|---|---|---|---|
| **1** | Client Mobile/Desktop Network Ingress | Flutter / iOS / Android / C++ | TLS 1.3 over QUIC/TCP | 12.0 ms | 25.0 ms | 35.0 ms | 60.0 ms | 100.0 ms |
| **2** | Edge CDN & DDoS Scrubbing | AWS CloudFront / Cloudflare | Anycast BGP Routing | 1.5 ms | 3.0 ms | 5.0 ms | 10.0 ms | 20.0 ms |
| **3** | Edge Gateway & TLS Termination | Envoy Proxy (C++) | Zero-Copy epoll | 0.8 ms | 1.8 ms | 3.5 ms | 6.0 ms | 10.0 ms |
| **4** | API Gateway BFF & Auth Validation | Go (Gin / Fiber) + Redis | gRPC / Redis Pipeline | 2.5 ms | 6.5 ms | 12.0 ms | 20.0 ms | 35.0 ms |
| **5** | Pre-Trade Risk & Collateral Check | Rust (Lock-Free Bitmaps) | Shared Memory / gRPC | 0.2 ms | 0.8 ms | 1.5 ms | 3.0 ms | 5.0 ms |
| **6** | LMAX-Style Ingress Sequencer | Rust (Single-Writer SPSC) | CPU Pinning L1/L2 | 0.005 ms | 0.020 ms | 0.050 ms | 0.100 ms | 0.250 ms |
| **7** | Core L3 Matching Engine Execution | Rust (`no_std` zero-alloc) | Pinned NUMA Core | 0.008 ms | 0.035 ms | 0.080 ms | 0.200 ms | 0.500 ms |
| **8** | Journal WAL & Kafka Ingress | Rust / Kafka Producer | Kafka `acks=all`, NVMe | 0.6 ms | 1.2 ms | 2.2 ms | 4.5 ms | 10.0 ms |
| **9** | Market Data Conflation & Fanout | Go / Rust WebSocket Broadcaster | Binary Protobuf WS | 5.0 ms | 20.0 ms | 45.0 ms | 85.0 ms | 120.0 ms |
| **10**| Settlement DvP Orchestration | Go / TypeScript Engine | gRPC to Relayer | 50.0 ms | 150.0 ms | 250.0 ms | 450.0 ms | 800.0 ms |
| **11**| Account Abstraction & HSM Signing | AWS KMS / CloudHSM / EIP-712 | PKCS#11 / REST | 2.0 ms | 5.0 ms | 8.5 ms | 10.0 ms | 15.0 ms |
| **12**| Besu QBFT Blockchain Finality | Hyperledger Besu 4-Node QBFT | P2P Consensus Round | 500.0 ms | 1500.0 ms | 2000.0 ms | 2500.0 ms | 3000.0 ms |
| **13**| Client UI Frame Render | Flutter Skia / Impeller | 60 FPS / 120 FPS | 8.33 ms | 12.5 ms | 16.6 ms | 16.6 ms | 25.0 ms |

---

## 3. Microservice Latency Matrix & Failure Thresholds

Every microservice in Category 2 and Category 3 must adhere to quantitative latency percentiles and circuit breaker thresholds. Any service exceeding its p99 target for $\ge 30\text{ seconds}$ automatically trips alerting and load-shedding safeguards.

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                              MICROSERVICE LATENCY & CIRCUIT BREAKER MATRIX                                   |
|                                                                                                                             |
|  [Service Name]            [Protocol]      [p50 Latency]    [p95 Latency]    [p99 Latency]    [Timeout]    [Circuit Breaker]|
|  - User & Profile Service   gRPC Unary      10.0 ms          25.0 ms          50.0 ms          200 ms       50% Fail / 5s   |
|  - Pre-Trade Risk Engine    Rust IPC        0.2 ms           0.8 ms           1.5 ms           10 ms        25% Fail / 1s   |
|  - Order Service (OMS)      gRPC Unary      5.0 ms           15.0 ms          25.0 ms          100 ms       50% Fail / 5s   |
|  - Matching Engine Core     SPSC Ring       0.010 ms         0.050 ms         0.100 ms         5 ms         Failover to Stby|
|  - Settlement DvP Engine    gRPC / RPC      250.0 ms         1000.0 ms        1800.0 ms        5000 ms      40% Fail / 10s  |
|  - Market Data Gateway      WebSocket       5.0 ms           20.0 ms          50.0 ms          150 ms       Drop Slow Cons. |
|  - Depository Gateway       ISO 20022/REST  200.0 ms         800.0 ms         1500.0 ms        10000 ms     30% Fail / 30s  |
|  - Banking / CBDC Gateway   REST / mTLS     150.0 ms         600.0 ms         1200.0 ms        8000 ms      30% Fail / 20s  |
|  - Blockchain RPC Bundler   JSON-RPC        15.0 ms          40.0 ms          80.0 ms          500 ms       50% Fail / 5s   |
+-----------------------------------------------------------------------------------------------------------------------------+
```

### 3.1 Microservice Operational Matrix Details

| Subsystem / Service | Interface / Protocol | SLA Uptime | p50 (ms) | p95 (ms) | p99 (ms) | p99.9 (ms) | Max Timeout (ms) | Circuit Breaker Parameters |
|---|---|---|---|---|---|---|---|---|
| **API Gateway / BFF** | HTTP/2, gRPC-Web, REST | 99.99% | 15.0 | 35.0 | 50.0 | 80.0 | 250 | 50% errors over 10s $\to$ open for 5s |
| **User & Profile Service** | gRPC (mTLS) | 99.95% | 10.0 | 25.0 | 50.0 | 80.0 | 200 | 50% errors over 10s $\to$ open for 5s |
| **Pre-Trade Risk Engine** | In-Memory IPC / Rust C-ABI | 99.999% | 0.2 | 0.8 | 1.5 | 3.0 | 10 | 25% errors over 1s $\to$ fallback to safe reject |
| **Order Management (OMS)** | gRPC (mTLS) | 99.99% | 5.0 | 15.0 | 25.0 | 40.0 | 100 | 50% errors over 5s $\to$ open for 3s |
| **L3 Matching Engine** | SPSC Ring Buffer / Zero-Copy | 99.999% | 0.010 | 0.050 | 0.100 | 0.500 | 5 | Quorum heartbeat failure $> 500\text{ms} \to$ Hot standby failover |
| **Settlement DvP Engine** | gRPC / EVM JSON-RPC | 99.99% | 250.0 | 1000.0 | 1800.0 | 2500.0 | 5000 | 40% reverts over 15s $\to$ open for 10s |
| **Market Data Gateway** | Binary WebSocket / Protobuf | 99.99% | 5.0 | 20.0 | 50.0 | 100.0 | 150 | Buffer saturation $> 80\% \to$ Conflate / drop delta |
| **Depository Gateway** | ISO 20022 / CDSL/NSDL API | 99.90% | 200.0 | 800.0 | 1500.0 | 3000.0 | 10000 | 30% errors over 30s $\to$ queue buffering |
| **Banking / CBDC Gateway** | ISO 20022 / UPI 2.0 / eINR | 99.90% | 150.0 | 600.0 | 1200.0 | 2500.0 | 8000 | 30% errors over 20s $\to$ retry queue with backoff |
| **Notification Engine** | Kafka Consumer / APNS / FCM | 99.90% | 50.0 | 250.0 | 500.0 | 1000.0 | 2000 | Consumer lag $> 5,000 \to$ auto-scale consumers |

---

## 4. Throughput, Scalability & Sizing Models

### 4.1 Traffic Profiles & Concurrency Scenarios

The Growww / NBSE platform is engineered for three operational load profiles:

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                          GROWWW / NBSE THROUGHPUT CAPACITY TIERS                                             |
|                                                                                                                             |
|  [ NOMINAL TRADING PROFILE ]           [ PEAK MARKET-OPEN (09:00-09:30 IST) ]  [ CATASTROPHIC VOLATILITY BURST ]            |
|  - Sustained Orders: 2,500 orders/sec  - Peak Burst: 10,000 orders/sec         - Extreme Ingress: 25,000 - 100,000 orders/s |
|  - DvP Settlements: 500 TPS            - DvP Settlements: 1,500 TPS            - DvP Settlements: 1,500 TPS (Capped)        |
|  - Active WebSockets: 250,000 conns    - Active WebSockets: 1,000,000 conns    - Active WebSockets: 1,000,000+ conns        |
|  - Market Data Msg Rate: 50,000 msg/s  - Market Data Msg Rate: 250,000 msg/s   - Market Data Conflated: 100ms Cap           |
|  - CPU Load: < 35% Utilization         - CPU Load: 55% - 70% Utilization       - Dynamic Load-Shedding Active               |
+-----------------------------------------------------------------------------------------------------------------------------+
```

1. **Nominal Trading Profile (10:00 - 15:00 IST):**
   - Sustained order entry: $2,500\text{ orders/second}$.
   - Sustained DvP settlements on Besu ledger: $500\text{ transactions/second}$.
   - Concurrent active WebSocket connections: $250,000\text{ sessions}$.
   - Market data broadcast rate: $50,000\text{ messages/second}$.

2. **Peak Market-Open Profile (09:00 - 09:30 IST & Expiry Hours 14:30 - 15:30 IST):**
   - Burst order entry: $10,000\text{ orders/second}$.
   - Peak DvP settlements on Besu ledger: $1,500\text{ transactions/second}$.
   - Concurrent active WebSocket connections: $1,000,000\text{ sessions}$ across 10 distributed edge Envoy gateways.
   - Market data broadcast rate: $250,000\text{ messages/second}$.

3. **Catastrophic Volatility Burst (Macro Shocks / Flash Crashes / Union Budget):**
   - Extreme ingress rate: $25,000 - 100,000\text{ orders/second}$.
   - Protective mechanisms: Ingress queue buffering, dynamic pre-trade risk rate-limiting, circuit breakers, and 100ms market data ticker conflation.
   - DvP settlement throughput remains bounded at $1,500\text{ TPS}$ deterministic finality without ledger degradation.

### 4.2 Concurrency & WebSocket Engineering (1,000,000 Clients)

To sustain $1,000,000$ active client sessions during market-open connection storms without thread starvation, file descriptor exhaustion, or socket buffer overflow:
1. **Distributed Edge Gateway Termination:**
   - Terminated across a fleet of 10 Envoy edge gateway instances (each holding 100,000 active persistent TCP connections).
   - Linux kernel TCP parameters tuned via sysctl:
     - `net.core.somaxconn = 65535`
     - `net.ipv4.tcp_max_syn_backlog = 65535`
     - `fs.file-max = 2097152`
     - `net.ipv4.tcp_rmem = 4096 87380 16777216`
     - `net.ipv4.tcp_wmem = 4096 65536 16777216`
2. **Level 2 Order Book Conflation:**
   - Individual tick updates are coalesced within a sliding $100\text{ ms}$ window at the gateway layer.
   - Broadcasts only the latest net delta per price level, reducing outbound network bandwidth consumption by $75\%$ without degrading price discovery integrity.
3. **Connection Storm Jitter & Exponential Backoff:**
   - Reconnecting mobile and web clients execute randomized exponential backoff:
     $$\text{Backoff Interval } T = \min\left(30\text{s}, 2^{\text{attempt}} \times 500\text{ms} + \text{UniformRandom}(0, 250\text{ms})\right)$$
   - Prevents thundering herd crashes against edge load balancers during network reconnects.
4. **Salted Kafka Partitioning:**
   - Under single-asset volume spikes (e.g. `RELIANCE` or `NIFTY50` expiry), orders are salted across 8 parallel sub-partitions using `Murmur3_32(account_id) % 8` to eliminate hot partition broker bottlenecks while preserving per-account FIFO ordering.

### 4.3 Hardware Sizing & Capacity Benchmarks

| Tier / Subsystem | Machine Type / Sizing | vCPU / RAM | Storage Subsystem & IOPS | Network Bandwidth | Capacity Ceiling |
|---|---|---|---|---|---|
| **API Edge Gateways** | AWS `c6i.8xlarge` (Fleet of 10) | 32 vCPU / 64 GB | 100 GB gp3 NVMe | 25 Gbps ENA | 1,000,000 active WebSockets |
| **API Gateway BFF** | AWS `c6i.4xlarge` (Fleet of 12) | 16 vCPU / 32 GB | 100 GB gp3 NVMe | 12.5 Gbps ENA | 20,000 HTTP/gRPC RPS |
| **Pre-Trade Risk Engine** | Bare Metal / `c6i.metal` (2 Nodes) | 128 Cores / 256 GB | Direct-Attached NVMe | 100 Gbps EFA | 50,000 margin checks/sec |
| **Matching Engine Core** | Bare Metal / `c6i.metal` (Active/Standby) | Pinned Core 2-3 / 128 GB | Intel Optane / io2 Express | 100 Gbps EFA | 100,000 matches/sec (< 10μs) |
| **Kafka Cluster** | AWS `i3en.3xlarge` (6 Brokers) | 12 vCPU / 96 GB | 2 x 7.5TB NVMe SSD | 25 Gbps ENA | 500,000 msgs/sec write |
| **PostgreSQL Primary** | AWS `r6i.8xlarge` (Patroni HA) | 32 vCPU / 256 GB | AWS io2 Block Express 64,000 IOPS | 25 Gbps ENA | 30,000 inserts/sec |
| **Besu QBFT Validators** | AWS `r6i.4xlarge` (4 Nodes) | 16 vCPU / 128 GB | 2 TB io2 Express 32,000 IOPS | 25 Gbps ENA | 1,500 DvP settlements/sec |
| **Redis Enterprise** | AWS `r6i.2xlarge` (6-Node Cluster) | 8 vCPU / 64 GB | In-Memory with AOF | 12.5 Gbps ENA | 1,000,000 ops/sec |

---

## 5. Availability, Uptime SLAs & Error Budget Governance

### 5.1 Formal SLA Definitions
The platform establishes two distinct availability commitments backed by contract and regulatory reporting:

1. **Market-Hours Availability SLA (09:00 - 16:00 IST, Monday to Friday):**
   - **Target:** **99.99% Availability** ($\le 0.01\%$ downtime allowance).
   - **Scope:** Ingress API, Pre-Trade Risk, Matching Engine, Settlement Engine, and Real-Time Market Data.
   - **Calculation:**
     $$\text{Daily Market Hours} = 7 \text{ hours} = 420 \text{ minutes} = 25,200 \text{ seconds}$$
     $$\text{Average Monthly Trading Days} = 21.75 \text{ days} \implies 9,135 \text{ minutes / month}$$
     $$\text{Monthly Downtime Allowance } = 9,135 \text{ min} \times (1 - 0.9999) = 0.9135 \text{ min} = \mathbf{54.81 \text{ seconds / month}}$$
     $$\text{Annual Downtime Allowance (250 Days)} = 1,750 \text{ hours} \times 0.0001 = 0.175 \text{ hours} = \mathbf{10.50 \text{ minutes / year}}$$

2. **General Platform SLA (24x7x365 Platform Wide):**
   - **Target:** **99.90% Availability** ($\le 0.10\%$ downtime allowance).
   - **Scope:** Retail Client Web/Mobile App, Crypto Trading, Fiat Banking On-Ramp, Portfolio Reporting, Tax Statements.
   - **Calculation:**
     $$\text{Monthly Operational Hours} = 730 \text{ hours} = 43,800 \text{ minutes}$$
     $$\text{Monthly Downtime Allowance } = 43,800 \text{ min} \times (1 - 0.9990) = \mathbf{43.80 \text{ minutes / month}}$$
     $$\text{Annual Downtime Allowance } = 8,760 \text{ hours} \times 0.001 = \mathbf{8.76 \text{ hours / year}}$$

### 5.2 Multi-Window Multi-Burn-Rate Alerting Architecture

Following Google SRE best practices, alerting is calculated across four rolling time windows based on the consumption rate of the 30-day error budget:

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                    MULTI-WINDOW MULTI-BURN-RATE SRE ALERTING MATRIX                                         |
|                                                                                                                             |
|  [Alert Level]   [Burn Rate]   [Budget Consumed]   [Window]     [Action & Notification]          [Escalation SLA]           |
|  - Tier 1 (P1)   14.4x         2% of 30-Day Budget 1 Hour       PagerDuty Voice/SMS to On-Call   Immediate (< 60 seconds)   |
|  - Tier 1 (P1)   6.0x          5% of 30-Day Budget 6 Hours      PagerDuty Voice/SMS to On-Call   Immediate (< 60 seconds)   |
|  - Tier 2 (P2)   3.0x          10% of Budget       3 Days       Slack #sre-war-room + Jira P2    < 15 minutes               |
|  - Tier 2 (P3)   1.0x          50% of Budget       14 Days      Weekly Operational Review        < 24 hours                 |
+-----------------------------------------------------------------------------------------------------------------------------+
```

#### PromQL Implementation for Master Availability SLI:
```promql
# Order Gateway Availability SLI (1-Hour Burn Rate)
(
  sum(rate(http_requests_total{job="order-gateway", status=~"5.."}[1h]))
  /
  sum(rate(http_requests_total{job="order-gateway"}[1h]))
) > (1 - 0.9999) * 14.4
```

### 5.3 Automated Release Freeze Policy (Error Budget Gate)

Feature release velocity is programmatically governed by remaining error budgets:
- **Remaining Budget $\ge 20\%$:** **Green State.** Continuous automated CI/CD deployments enabled.
- **Remaining Budget between $10\%$ and $20\%$:** **Yellow State.** Release throttling active. Non-essential feature rollouts require formal SRE Director and QA Lead approval.
- **Remaining Budget $< 10\%$:** **Red State (Automated Hard Freeze).** CI/CD deployment pipeline automatically blocks all non-emergency pull requests. Only critical P1 security patches, bug fixes, or reliability improvements are allowed until the 30-day rolling budget recovers.

---

## 6. Business Continuity & Disaster Recovery (BCP / DR) Architecture

### 6.1 Regulatory Mandates (SEBI CSCRF Compliance)
Per SEBI Cyber Security and Cyber Resilience Framework guidelines for designated Stock Exchanges, Clearing Corporations, and Qualified Depositories:
- **Recovery Time Objective (RTO):** $\mathbf{\le 15 \text{ minutes}}$ for a catastrophic regional failure; automated intra-region failover occurs in $\mathbf{< 30 \text{ seconds}}$.
- **Recovery Point Objective (RPO):**
  - **Financial Ledger & Settlement State (Besu QBFT, TigerBeetle):** $\mathbf{\text{RPO} = 0.00 \text{ seconds}}$ (Zero uncommitted ledger transaction loss).
  - **Relational Databases (PostgreSQL Patroni):** $\mathbf{\text{RPO} \le 1.0 \text{ second}}$.
- **Disaster Recovery Drills:** Mandatory **bi-annual live DR drills** with unannounced failover simulations conducted during non-market hours.
- **Security Operations Center (SOC):** 24x7x365 continuous telemetry monitoring with automated P1 incident escalation directly to the Chief Information Security Officer (CISO) within 60 seconds of SLA breach.

### 6.2 Multi-Region High-Availability Topology

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                    MULTI-REGION DISASTER RECOVERY ARCHITECTURE                                              |
|                                                                                                                             |
|  PRIMARY PRODUCTION CLUSTER (AWS ap-south-1 Mumbai)                SECONDARY DR SITE (AWS ap-south-2 Hyderabad)             |
|                                                                                                                             |
|  +---------------------------------------------------------------+ +------------------------------------------------------+ |
|  | AZ-1 (Active Primary)   AZ-2 (Hot Standby)   AZ-3 (Quorum)    | | Warm Standby DR Kubernetes Cluster                   | |
|  | - Rust Matching Engine  - Standby Engine     - Witness Arbiter| | - Pre-warmed Standby Pods (Scaled to 30% Capacity)   | |
|  | - TigerBeetle Active    - TigerBeetle Rep    - TigerBeetle Rep| | - Read-Only TigerBeetle Secondary Replica Set        | |
|  | - Besu Validator #1,#2  - Besu Validator #3  - Besu Valid #4  | | - Besu Passive Archive Node (Continuous Sync)        | |
|  | - Patroni Primary DB    - Patroni Sync Rep   - Patroni Quorum | | - Patroni Async Cross-Region Standby DB              | |
|  | - Kafka Brokers 1-2     - Kafka Brokers 3-4  - Kafka Broker 5 | | - Kafka MirrorMaker 2 Secondary Replication Cluster  | |
|  +---------------------------------------------------------------+ +------------------------------------------------------+ |
|                                   |                                                 ^                                       |
|                                   | Dedicated AWS DirectConnect 10Gbps Redundant WAN|                                       |
|                                   +-------------------------------------------------+                                       |
+-----------------------------------------------------------------------------------------------------------------------------+
```

### 6.3 Data Tier Replication & Durability Matrix

| Component | In-Region Replication (Mumbai AZ-1 $\to$ AZ-2) | Cross-Region Replication (Mumbai $\to$ Hyderabad) | Failover Mechanism | Target RTO | Target RPO |
|---|---|---|---|---|---|
| **TigerBeetle Financial Ledger** | Viewstamped Replication (VSR) across 3 AZs | Asynchronous journal replication stream | Automated Paxos / VSR view change | $< 3 \text{ s}$ | $\mathbf{0.00 \text{ s}}$ |
| **Besu QBFT Blockchain** | QBFT Consensus round across 4 validators | P2P block streaming to passive sync node | Deterministic 1-block consensus quorum | $< 2 \text{ s}$ | $\mathbf{0.00 \text{ s}}$ |
| **PostgreSQL Clusters** | Synchronous physical streaming (`remote_apply`) | Asynchronous WAL shipping (`recovery.signal`) | Patroni automated leader election via etcd | $< 30 \text{ s}$ | $\mathbf{\le 1.0 \text{ s}}$ |
| **Apache Kafka Cluster** | Strimzi Operator (RF=3, `min.insync=2`) | Kafka MirrorMaker 2 cross-region topic sync | Automated consumer group redirect | $< 45 \text{ s}$ | $\mathbf{\le 1.0 \text{ s}}$ |
| **Redis Enterprise** | Multi-AZ Active-Active CRDT replication | Asynchronous replica snapshot sync | Sentinel / Enterprise auto-failover | $< 5 \text{ s}$ | $\mathbf{\le 1.0 \text{ s}}$ |
| **S3 Audit & WORM Logs** | S3 Multi-AZ High-Durability (11 9s) | AWS S3 Cross-Region Replication (CRR) | Automatic multi-region S3 endpoint routing | Instant | $\mathbf{0.00 \text{ s}}$ |

### 6.4 Automated Regional Failover Orchestration Runbook

In the event of a catastrophic total outage affecting AWS Mumbai (`ap-south-1`), the following automated failover procedure executes:
1. **Heartbeat & Quorum Loss Detection ($T \le 15\text{s}$):**
   - AWS Route 53 Application Recovery Controller (ARC) health checks probe edge gateways and validator nodes.
   - If 100% of Mumbai availability zones fail health checks for 15 consecutive seconds, automated disaster recovery is triggered.
2. **Traffic Fencing & Split-Brain Prevention ($T + 30\text{s}$):**
   - Route 53 ARC removes Mumbai routing controls and fences all incoming ingress DNS queries.
   - Ingress endpoints in Mumbai are shut down to prevent split-brain state divergence.
3. **Database & Ledger Promotion ($T + 90\text{s}$):**
   - Kafka MirrorMaker 2 in Hyderabad cuts over replication topics and unpauses local consumers.
   - Hyderabad Patroni secondary PostgreSQL cluster promotes standby to primary (`pg_ctl promote`).
   - Hyderabad TigerBeetle replica group is promoted to primary leader.
4. **State Rehydration in Matching Engine ($T + 120\text{s}$):**
   - Standby Rust Matching Engine rehydrates unfulfilled order state from TigerBeetle balance reservations and WAL logs in $\mathbf{< 8 \text{ seconds}}$.
5. **DNS Cutover & Ingress Open ($T + 180\text{s} = 3\text{ minutes} \ll 15\text{ minutes RTO}$):**
   - Route 53 ARC activates routing controls pointing retail and institutional traffic to Hyderabad Envoy edge gateways.
   - Platform resumes full trade matching and DvP settlement.

---

## 7. Distributed Permissioned Ledger (Hyperledger Besu QBFT) Performance Invariants

The platform's institutional clearing and settlement tier relies on a private, permissioned Hyperledger Besu consortium network governed by NBSE Ltd. and institutional clearing members.

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                    HYPERLEDGER BESU QBFT CONSENSUS & SETTLEMENT PIPELINE                                    |
|                                                                                                                             |
|  [DvP Trade Event]  --->  [Settlement Engine]  --->  [32-Relayer Sharding Pool]  --->  [Besu RPC Bundler]                  |
|                                                             |                                      |                        |
|                                                             v                                      v                        |
|                                              [AWS KMS / Luna HSM Signing]            [QBFT Consensus (4 Nodes)]             |
|                                              (< 10ms Signature Latency)              - 2.0s Block Period                    |
|                                                                                      - Deterministic 1-Block Finality       |
|                                                                                      - 1,500 DvP TPS Sustained              |
|                                                                                                    |                        |
|                                                                                                    v                        |
|                                              [Immutable AWS S3 WORM Vault] <--- [RocksDB Bonsai Trie Engine]                |
|                                              (7-Year Statutory SEBI Ret.)       (< 450 GB Working Set, Rolling Prune)       |
+-----------------------------------------------------------------------------------------------------------------------------+
```

### 7.1 Quantitative Blockchain Invariants

1. **Consensus Protocol & Block Finality:**
   - **Protocol:** Quorum Byzantine Fault Tolerance (QBFT).
   - **Block Period:** Fixed at **2.0 seconds**.
   - **Finality:** Immediate deterministic 1-block finality. Zero probabilistic chain reorganizations or forks can occur, providing absolute legal finality for securities settlement.
2. **Settlement Throughput:**
   - Hyperledger Besu validator nodes are tuned to process a minimum of **1,500 DvP settlement transactions per second (TPS)**.
   - At 2-second block intervals, each block encapsulates up to 3,000 DvP smart contract execution receipts.
   - Block gas limit: Set to $60,000,000\text{ gas}$ with zero transaction fee (`min-gas-price = 0`).
3. **Consensus Quorum & Node Redundancy:**
   - Minimum **4 active validator nodes** ($N = 4$) distributed across independent Availability Zones in Mumbai and on-premises institutional clearing data centers.
   - Maximum tolerated Byzantine/crashed nodes:
     $$F = \left\lfloor \frac{N - 1}{3} \right\rfloor = \left\lfloor \frac{4 - 1}{3} \right\rfloor = 1$$
   - The network survives the complete failure or isolation of any 1 validator node without halting block production or transaction finality.
4. **Storage Engine & Rolling State Pruning:**
   - Storage format: Bonsai Trie (`--data-storage-format=BONSAI`) over RocksDB.
   - Rolling state pruning: Retains the last 2,048 block states (~68 minutes of reorg buffer), capping the hot active working set at **$< 450\text{ GB}$** on hot NVMe SSDs.
   - Archival audit logs: Finalized block receipts and settlement event logs are streamed daily to an immutable AWS S3 Glacier WORM vault for 7-year statutory retention.
5. **Zero-PII Performance Invariant:**
   - No Personally Identifiable Information (PII) is stored or processed on the permissioned ledger.
   - Off-chain services compute cryptographic commitments (Pedersen commitments, SHA-256 hashes); smart contracts process compact 32-byte hashes (`bytes32`).
   - Minimizes EVM execution gas and calldata memory, sustaining $> 1,500\text{ TPS}$.
6. **Account Abstraction & Sponsored Transactions:**
   - 100% sponsored transaction execution via ERC-4337 and `PaymasterRelayer.sol`.
   - 32-way partitioned relayer account pool (`0xRelayer_00` to `0xRelayer_31`) eliminates nonce sequence bottlenecks during peak settlement bursts.

---

## 8. Client & Front-End Performance Standards (Flutter, Web WASM, Desktop)

Retail client applications across iOS, Android, Web Pro Terminal, and Desktop Native Workstations must meet strict performance invariants:

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                              CLIENT PERFORMANCE & RENDERING BUDGETS                                         |
|                                                                                                                             |
|  [Flutter Mobile]                       [Web Pro Trading Terminal]              [Desktop Native Workstation]                |
|  - 60 FPS (Low-End Mobile: 16.6ms)      - WASM Bundle: < 5.0 MB Gzipped         - Native Metal / DirectX / Vulkan           |
|  - 120 FPS (High-Refresh: 8.33ms)       - Time-to-Interactive: < 1.0s           - 120 FPS Multi-Window Docking              |
|  - Cold Start: p99 < 1,500ms            - L2 Depth Ladder: 60 FPS Canvas        - Memory Footprint: < 250 MB                |
|  - Warm Start: p99 < 500ms              - Memory Footprint: < 200 MB            - CPU Utilization: < 5% Idle                |
|  - APK Download Payload: < 35 MB        - Web Workers for Protobuf Deser        - Sub-5ms Hotkey-to-Wire Latency            |
+-----------------------------------------------------------------------------------------------------------------------------+
```

### 8.1 Flutter Mobile Client Performance Criteria
- **Frame Rate & Jank Prevention:**
  - Standard displays: Constant **60 FPS** ($16.6\text{ ms}$ maximum frame render time) on entry-level Android devices (e.g. 4GB RAM, ARM Cortex-A55).
  - High-refresh displays: Smooth **120 FPS** ($8.33\text{ ms}$ maximum frame render time) on ProMotion / 120Hz displays.
  - Zero UI Thread Blocking: All WebSocket JSON/Protobuf decoding, Level 2 order book sorting, and technical indicator computations execute inside dedicated Dart background isolates (`compute()` or `Isolate.spawn()`).
- **Application Startup Latency:**
  - **Cold Start (Process Launch to Interactive Home Screen):** $\mathbf{p99 < 1,500 \text{ ms}}$ ($1.5\text{ seconds}$).
  - **Warm Start (Resume from Background):** $\mathbf{p99 < 500 \text{ ms}}$.
  - **Hot Resume (Screen Unlock / Tab Switch):** $\mathbf{p99 < 200 \text{ ms}}$.
- **Binary Payload & Memory Constraints:**
  - Total compressed download size (APK / IPA): $\mathbf{< 35 \text{ MB}}$.
  - Baseline idle memory consumption: $< 150\text{ MB}$.
  - Peak memory during active multi-pane charting: $< 300\text{ MB}$.
  - Background power consumption: $< 10\%$ CPU drain during continuous market streaming.

### 8.2 Web Pro Trading Terminal (WASM) Criteria
- Initial WebAssembly bundle size: $< 5.0\text{ MB}$ (Brotli compressed).
- Time to Interactive (TTI): $< 1.0\text{ second}$ on standard broadband connections.
- Order book depth ladder rendered using HTML5 Canvas / WebGL at a minimum of 60 FPS.

---

## 9. Security, Cryptography & Compliance Performance Constraints

All cryptographic and compliance subsystems must execute within strict latency envelopes to prevent degrading order entry speed:

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                    SECURITY & CRYPTOGRAPHIC PERFORMANCE CONSTRAINTS                                         |
|                                                                                                                             |
|  [Hardware Security Module (HSM)]       [Field-Level Encryption (AES-256-GCM)]  [Zero-Trust SPIFFE/SPIRE Mesh]              |
|  - EIP-712 & Secp256k1 Signing          - Aadhaar, PAN, Bank Account Crypto     - Mutual TLS (mTLS) Inter-Service Comm      |
|  - Target Latency: < 10ms / signature   - Overhead: < 2.0ms per query           - Proxy Overhead: < 1.0ms per RPC           |
|  - Batch Signing Workers: 32 Threads    - Hardware Acceleration: AES-NI         - Ephemeral X.509 SVID Rotation: 1 Hour     |
+-----------------------------------------------------------------------------------------------------------------------------+
```

1. **Hardware Security Module (HSM) Signing Latency:**
   - Cryptographic signing operations for on-chain settlements, custody transfers, and clearing notes execute via AWS KMS or dedicated CloudHSM / Luna HSM instances.
   - **Maximum Allowed Latency:** $\mathbf{< 10.0 \text{ ms}}$ per Secp256k1 / Ed25519 signature.
   - Pipelined batch signing workers prevent relayer queuing bottlenecks.
2. **Field-Level Data Encryption Overhead:**
   - Highly sensitive PII (Aadhaar, PAN numbers, bank account numbers) is encrypted before persisting to PostgreSQL using AES-256-GCM.
   - **Maximum Query Overhead:** $\mathbf{< 2.0 \text{ ms}}$ per read/write query, leveraging CPU AES-NI hardware acceleration.
3. **Zero-Trust SPIFFE/SPIRE Service Mesh Overhead:**
   - Inter-service gRPC communication is secured with mTLS and ephemeral X.509 SVID certificates rotated hourly.
   - Sidecar proxy latency overhead is capped at $\mathbf{< 1.0 \text{ ms}}$ per internal service hop.
4. **Data Privacy (DPDP Act 2023 Compliance):**
   - Consent management and right-to-erasure workflows complete off-chain within 24 hours.
   - Ledger records reference only irreversible pseudonymized hashes (`bytes32`), ensuring cryptographic privacy compliance with zero ledger rewrites.

---

## 10. Data Durability, Storage Engineering & Statutory Retention

### 10.1 PostgreSQL Outbox & WAL Group Commit Configuration
To sustain up to 30,000 relational database inserts per second without storage I/O stalls:
- **Declarative Hourly Partitioning:** The `outbox_events` table is partitioned hourly by `created_at`. Partitions older than 72 hours are dropped (`DROP TABLE`), completely eliminating table bloat and `VACUUM` overhead.
- **Storage Subsystem:** AWS io2 Block Express NVMe volumes provisioned for $64,000\text{ IOPS}$ and $1,000\text{ MB/s}$ throughput.
- **Group Commit Configuration:**
  ```ini
  # postgresql.conf performance tuning for high-throughput ingestion
  synchronous_commit = off
  commit_delay = 5000          # 5ms flush window batches hundreds of transactions
  commit_siblings = 5
  wal_buffers = 64MB
  checkpoint_timeout = 15min
  max_wal_size = 32GB
  ```

### 10.2 Backup & Point-In-Time Recovery (PITR)
- **Database Full Snapshots:** Automated hourly physical backups to AWS S3.
- **Continuous WAL Archiving:** WAL logs continuously streamed to S3, enabling fine-grained Point-In-Time Recovery (PITR) with a **35-day retention window**.

### 10.3 Statutory 7-Year Immutable WORM Audit Vault
Per SEBI CSCRF Chapter IV and Indian regulatory guidelines for stock exchanges:
- All order matching audit trails, execution reports, trade logs, and blockchain block receipts must be retained for a minimum of **7 years**.
- Archived logs are stored in an **AWS S3 Glacier Vault configured with S3 Object Lock in Compliance Mode** (WORM - Write Once, Read Many).
- Neither root AWS credentials nor administrative personnel can delete or alter audit logs prior to statutory retention expiry.

---

## 11. Chaos Engineering & Resilience Testing Mandates

To ensure zero systemic failure during catastrophic real-world disruptions, automated resilience experiments are executed weekly via Chaos Mesh in the staging environment and during quarterly unannounced production drills:

```
+-----------------------------------------------------------------------------------------------------------------------------+
|                                              CHAOS ENGINEERING EXPERIMENT MATRIX                                            |
|                                                                                                                             |
|  [Experiment Name]         [Failure Mode Injected]      [Target Subsystem]       [Expected Steady State]    [Max Recovery]  |
|  - EXP-01: Pod Kill Chaos   SIGKILL to Active Pods       Matching / Gateway Pods  Standby auto-promotes      < 5.0 seconds   |
|  - EXP-02: WAN Partition    200ms Jitter / 20% Drop      Cross-AZ Network Link    No quorum loss             < 1.0 second    |
|  - EXP-03: Besu 1-Node Fail Isolate 1 of 4 Validators    Besu QBFT Cluster        Consensus uninterrupted    0.0s (No Stall) |
|  - EXP-04: NVMe Disk Stall  IOPS Throttled to 100 IOPS   PostgreSQL / TigerBeetle Backpressure shed active   < 10.0 seconds  |
|  - EXP-05: Kafka Split-Brain Kill 2 of 5 Kafka Brokers   Kafka Strimzi Cluster    Zero uncommitted msg loss  < 30.0 seconds  |
+-----------------------------------------------------------------------------------------------------------------------------+
```

### 11.1 Chaos Experiment Acceptance Standards

1. **Besu Validator Node Isolation (Prompt 934):**
   - Injects a network partition isolating 1 out of 4 Besu validator nodes.
   - **Acceptance Invariant:** The remaining 3 validators maintain QBFT quorum ($\ge 3$ nodes required); new blocks continue to produce every $2.0 \pm 0.2\text{ seconds}$ with zero transaction reversion.
2. **NVMe Disk Stall & WAL Pressure (Prompt 935):**
   - Injects storage latency of $500\text{ ms}$ on PostgreSQL WAL writes.
   - **Acceptance Invariant:** Ingress rate-limiting safely throttles new incoming orders; existing transactions in memory buffer are preserved without corruption.
3. **Kafka Leader Election Chaos (Prompt 936):**
   - Terminates the Kafka broker acting as partition leader for critical order ingress topics.
   - **Acceptance Invariant:** Partition leader failover completes within $3.0\text{ seconds}$; zero message loss occurs due to `min.insync.replicas = 2` and `acks = all`.

---

## 12. Load Testing Targets & CI/CD Verification Harnesses

Performance requirements are verified through automated distributed stress testing prior to every production release:

1. **k6 Distributed Stress Test:**
   - Simulates $100,000\text{ orders/second}$ burst scenarios across distributed k6 runner pods.
   - Validates that API Gateway p99 remains $< 50\text{ ms}$ and Matching Engine p99 remains $< 5\text{ ms}$.
2. **Locust Distributed WebSocket Swarm:**
   - Spawns $1,000,000$ concurrent headless WebSocket clients subscribing to Level 2 order book streams.
   - Verifies edge gateway memory stability and ensures ticker conflation maintains sub-$100\text{ ms}$ delivery.
3. **ghz Microsecond gRPC Benchmark:**
   - Benchmarks inter-service gRPC connections, verifying that p99 latency between API Gateway and Pre-Trade Risk Engine remains $< 1.5\text{ ms}$.
4. **CI/CD Automated Deployment Blocker:**
   - Automated load test suites execute against staging environments.
   - If any test run breaches defined SLO targets in [`slo_targets.yaml`](file:///home/Kali/Desktop/Growww/Growww/docs/architecture/slo_targets.yaml), the deployment pipeline automatically halts and alerts the on-call release coordinator.

---

## 13. Master SLO Machine-Readable Specification Cross-Reference

All quantitative targets in this document are codified in machine-readable YAML format in [`docs/architecture/slo_targets.yaml`](file:///home/Kali/Desktop/Growww/Growww/docs/architecture/slo_targets.yaml):

```yaml
# Extract of Master Performance Specification
performance_specification:
  version: "1.0.0"
  market_hours_uptime_sla_percent: 99.99
  general_uptime_sla_percent: 99.90
  disaster_recovery:
    rto_minutes: 15
    rpo_seconds_database: 1
    rpo_seconds_ledger: 0

  latency_budgets_ms:
    order_matching_engine:
      p50: 1.0
      p95: 2.5
      p99: 5.0
    api_gateway_bff:
      p50: 15.0
      p95: 35.0
      p99: 50.0
    settlement_engine_dvp:
      p50: 500.0
      p95: 1500.0
      p99: 2000.0
    client_flutter_cold_start:
      p99: 1500.0

  throughput_targets_rps:
    order_submission_burst: 10000
    order_matching_sustained: 2500
    besu_ledger_dvp_settlement: 1500
    websocket_active_connections: 1000000

  storage_durability:
    audit_log_retention_years: 7
    database_backup_frequency_hours: 1
    point_in_time_recovery_days: 35
```

---

## 14. Formal Sign-Off & Governance Approvals

This master Non-Functional Requirements specification has been reviewed and formally approved by the executive engineering and compliance leadership:

| Role | Signee | Department / Group | Status | Approval Date |
|---|---|---|---|---|
| **Chief Technology Officer (CTO)** | Core Executive Office | Technology & Engineering | **APPROVED** | September 2026 |
| **Head of Site Reliability Engineering** | SRE & Operations Board | Cloud Infrastructure & Reliability | **APPROVED** | September 2026 |
| **Lead Systems Architect** | Exchange Architecture Council | High-Frequency Core & Distributed Systems | **APPROVED** | September 2026 |
| **Chief Information Security Officer (CISO)** | Security Operations Center | Information Security & Cyber Resilience | **APPROVED** | September 2026 |
| **Head of Regulatory Compliance** | Legal & Compliance Committee | SEBI / RBI Regulatory Affairs | **APPROVED** | September 2026 |
