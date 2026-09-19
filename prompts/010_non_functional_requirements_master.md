# 010 - Non-Functional Requirements & Performance Engineering Master Spec

## Purpose
Operating a high-concurrency, mission-critical financial exchange and permissioned settlement infrastructure demands strict, quantifiable Non-Functional Requirements (NFRs). Sub-millisecond latency in order matching, high-throughput DvP settlement, absolute data durability, zero-loss disaster recovery, and 99.99% availability during market hours are mandatory to meet SEBI cyber resilience standards and deliver an institutional-grade retail trading experience.

This document establishes the master quantitative Service Level Objectives (SLOs), Service Level Agreements (SLAs), latency budgets, throughput targets, capacity models, disaster recovery objectives (RTO/RPO), and reliability engineering standards across the entire Growww technical stack.

## What You Are Building
A foundational Non-Functional Requirements and performance engineering master specification (`docs/architecture/non_functional_requirements.md`) that details:
- Latency Budgets & Percentile Targets across all subsystems (Order Matching Engine p99 < 5ms; API Gateway p99 < 50ms; Hyperledger Besu settlement confirmation p99 < 2s; Flutter UI frame rate 60/120 FPS).
- Throughput & Capacity Targets: baseline 2,500 orders/sec sustained, burst capacity 10,000 orders/sec; 1,500 DvP settlement transactions/sec on the permissioned ledger.
- Availability & Uptime SLAs: 99.99% system availability during Indian market hours (09:00 - 16:00 IST), 99.9% overall platform availability.
- Business Continuity & Disaster Recovery (BCP/DR) Mandates: Recovery Time Objective (RTO) $\le 15$ minutes; Recovery Point Objective (RPO) $= 0$ for ledger state and RPO $\le 1$ second for relational databases.
- Concurrency & Scalability Profiles: support for 1,000,000 concurrently connected WebSocket client sessions.
- Data Durability & Archival Rules: 100% ACID transaction compliance, 7-year immutable WORM audit log retention.

## Scope Boundaries
- **In Scope:**
 - Quantitative NFR tables, latency budgets, and SLA/SLO definitions.
 - Multi-region high-availability topology requirements and disaster recovery drill procedures.
 - Resource sizing benchmarks, load testing targets, and chaos engineering standards.
- **Out of Scope / Handled Elsewhere:**
 - Kubernetes cluster deployment manifests and Terraform scripts (covered in Category 8).
 - Load testing scripts and benchmark harnesses (covered in Prompt 905).
 - Specific microservice performance tuning (covered in Category 2).

## Technology to Use
- Metrics & Observability: OpenTelemetry standard with Prometheus metrics, Jaeger distributed tracing, and Grafana dashboards.
- Load & Stress Testing: k6 and Locust distributed load testing frameworks.
- Specification Format: Markdown (`docs/architecture/non_functional_requirements.md`) with YAML SLO definitions (`docs/architecture/slo_targets.yaml`).
- Justification: Codifying SLOs in machine-readable YAML allows automated continuous verification in CI/CD load testing pipelines, blocking deployments that fail latency or throughput budgets.

## Backend / Infra Touchpoints
- Distributed Kubernetes Cluster (AWS EKS Multi-AZ `ap-south-1`).
- Kafka Event Streaming Cluster (Multi-AZ with replication factor 3, `min.insync.replicas=2`).
- Distributed PostgreSQL Clusters (Patroni HA with synchronous physical replication).
- Redis Enterprise / KeyDB In-Memory Caches.
- Hyperledger Besu 4-Validator QBFT Consortium Cluster.

## Blockchain Interaction
Establishes the quantitative performance invariants for the Hyperledger Besu permissioned ledger:
- **Consensus & Block Finality:** QBFT consensus configured with a fixed 2-second block period, delivering immediate deterministic finality (zero probabilistic forks or chain reorganizations).
- **Settlement Throughput:** Hyperledger Besu network tuned to process a minimum of 1,500 DvP settlement transactions per second (TPS) using optimized JVM memory settings, RocksDB state storage, and multi-threaded transaction validation.
- **Node Redundancy & Quorum:** Minimum 4 validator nodes distributed across separate Availability Zones and entity premises; consensus survives failure of any 1 validator node ($F = (N-1)/3$).
- **Zero-PII Performance Invariant:** Hashing and cryptographic commitments are executed off-chain; on-chain contracts process compact 32-byte hashes, maximizing ledger gas efficiency and execution throughput.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Review SEBI Cyber Security and Cyber Resilience Framework (CSCRF) guidelines regarding stock exchange infrastructure uptime and recovery benchmarks.
2. Establish the System-Wide Latency Budget, allocating maximum acceptable response times from client tap to ledger receipt.
3. Formulate the Microservice Latency Matrix defining p50, p95, p99, and p99.9 latency limits for User Service, Order Service, Matching Engine, and Settlement Engine.
4. Establish the Throughput and Concurrency Profiles: quantify nominal, peak market-open (09:15 IST), and catastrophic volatility burst scenarios.
5. Define the Availability & Reliability SLAs: calculate error budgets and downtime allowances for 99.99% market-hour uptime.
6. Design the Disaster Recovery Architecture: establish synchronous replication between primary data center (Mumbai AZ-1) and hot standby (Mumbai AZ-2), with asynchronous replication to disaster recovery site (Hyderabad).
7. Specify the RTO and RPO Targets: mandate zero data loss ($\text{RPO} = 0$) for financial ledger transactions and maximum 15-minute failover ($\text{RTO} \le 15\text{ min}$).
8. Define the Client Performance Standards for Flutter: 60 FPS minimum on low-end mobile devices (120 FPS on high-refresh displays), cold start time < 1.5 seconds, warm start < 500ms, payload size < 35MB.
9. Establish the Security & Compliance Performance Constraints: HSM signing latency < 10ms per cryptographic signature; field-level encryption overhead < 2ms per query.
10. Define the Chaos Engineering and Resilience Testing Mandates (automated weekly Chaos Mesh experiments testing pod termination, network partitioning, and disk corruption).
11. Construct the machine-readable YAML SLO specification (`slo_targets.yaml`).
12. Review and obtain formal sign-off from Chief Technology Officer, Head of Site Reliability Engineering, and Lead Systems Architect.

## Interfaces / Contracts
```yaml
# Master SLO & Performance Specification (docs/architecture/slo_targets.yaml)
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

## Security & Compliance Notes
- Mandated by SEBI CSCRF: Financial institutions must conduct bi-annual live Disaster Recovery (DR) drills with unannounced failover simulations during non-market hours.
- Automated system health telemetry must report directly to the Chief Information Security Officer (CISO) and Security Operations Center (SOC) 24x7x365 with automated P1 incident escalation within 60 seconds of SLA breach.

## Acceptance Criteria
- [ ] `docs/architecture/non_functional_requirements.md` is complete with comprehensive latency, throughput, availability, and BCP/DR targets.
- [ ] Quantitative performance matrices for all Category 2 microservices and Category 3 blockchain nodes are codified.
- [ ] RTO $\le 15$ min and RPO $= 0$ (ledger) / RPO $\le 1$s (DB) failover architectures are documented.
- [ ] Machine-readable `slo_targets.yaml` is generated and validated for integration into Prometheus alert rules.
- [ ] Formal sign-off obtained from Head of SRE, Chief Technology Officer, and Lead Systems Architect.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 002 (Two-Entity Structure).
- Parallel Tasks: 101 (System Architecture), 801 (Infra Overview), 905 (Load Testing).
- Downstream Blockers: Blocks all infrastructure provisioning prompts in Category 8 and test automation in Category 9.
