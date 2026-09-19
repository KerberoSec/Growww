# 904 - Chaos Engineering & Resilience Testing Plan

## Purpose
Establishes a systematic chaos engineering and resilience testing framework to proactively identify vulnerabilities, race conditions, and single points of failure across the Growww distributed infrastructure. Because financial systems must guarantee zero state loss and rapid recovery during catastrophic outages (per SEBI and RBI business continuity and disaster recovery mandates), this suite validates that the platform gracefully withstands database master failovers, Kafka broker crashes, validator network partitions, and extreme network latencies.

## What You Are Building
An automated, continuous chaos engineering harness (`tests/chaos/`):
- Chaos Mesh Custom Resource Definitions (CRDs) targeting Kubernetes pods, network layers, disk I/O, and compute resources.
- Automated fault injection game-day scenarios validating Recovery Time Objective (RTO < 30 seconds) and Recovery Point Objective (RPO = 0) for transactional state.
- Hyperledger Besu QBFT consensus resilience test suite injecting validator node partitions, Byzantine delays, and cluster restarts.
- Automated resilience verification assertions that monitor trade matching integrity and database ACID guarantees during active chaos events.

## Scope Boundaries
- **In Scope:**
 - Kubernetes-native chaos experiments (Pod kills, container crashes, node draining, memory/CPU exhaustion).
 - Network fault injection (Packet loss, latency spikes, jitter, DNS failures, split-brain network partitions).
 - PostgreSQL primary failover under active transaction load (Patroni / Cloud RDS Multi-AZ switchover).
 - Kafka broker loss, leader partition rebalancing, and consumer group re-elections.
 - Hyperledger Besu QBFT 4-validator quorum failure scenarios (1-node failure fault tolerance).
 - Redis Sentinel / Cluster node failover during high-volume caching and WebSocket pub/sub.
- **Out of Scope / Handled Elsewhere:**
 - Load and capacity benchmarks without failure injection (Prompt 903).
 - Application-level unit/integration tests (Prompt 901).
 - Formal disaster recovery cross-region failover runbooks (Prompt 710, 802).

## Technology to Use
- **Primary Chaos Framework:** Chaos Mesh (CNCF incubating Kubernetes-native chaos platform) - selected for its CRD-based declarative workflows, fine-grained RBAC controls, and native support for complex network/pod/kernel-level fault injection.
- **Secondary Network Simulator:** Toxiproxy (Shopify) for deterministic TCP fault simulation between application microservices and mock upstream APIs.
- **Observability Integration:** Prometheus alerts, Grafana dashboards, and OpenTelemetry spans capturing error rate spikes and service recovery latencies during experiments.

*Justification:* Chaos Mesh provides declarative, auditable, and automated fault injection directly inside Kubernetes without requiring invasive agent modifications to production containers.

## Backend / Infra Touchpoints
- Staging / Pre-Production Kubernetes Cluster running Chaos Mesh operator.
- PostgreSQL 16+ high-availability cluster managed by Patroni / Cloud Multi-AZ.
- Apache Kafka 3-broker cluster with active replication factor = 3 and `min.insync.replicas` = 2.
- Redis 7+ Cluster (3 primaries, 3 replicas) with automated failover.
- Hyperledger Besu QBFT consortium network (4 validator nodes).

## Blockchain Interaction
- Injects network partition chaos between Hyperledger Besu validator nodes to test QBFT consensus resilience ($N=4$, fault tolerance $F=1$, requiring $2F+1=3$ validator quorum).
- Verifies that when 1 validator is killed or partitioned, the remaining 3 validators continue proposing and committing blocks at 2-second intervals without ledger halt.
- Verifies that when 2 validators are partitioned (loss of quorum), the chain gracefully pauses block production without corrupting state, and immediately resumes normal block production and syncs missing blocks once the network partition heals.
- Tests DvP atomic settlement transaction retries and idempotency key enforcement when the Go settlement relayer experiences sudden RPC disconnections during on-chain execution.

## Step-by-Step Build Instructions
1. Scaffold the `tests/chaos/` directory with subfolders for `experiments/`, `workflows/`, `scripts/`, and `runbooks/`.
2. Deploy Chaos Mesh operator into the staging Kubernetes cluster using Helm charts with restricted namespace RBAC.
3. Implement `01_pod_kill_matching_engine.yaml` to randomly terminate Rust matching engine pods and verify zero order state loss from Redis/Kafka replay.
4. Implement `02_postgres_primary_failover.yaml` to trigger primary database crash during active order submission and assert zero lost trades (RPO = 0) and failover RTO < 20s.
5. Implement `03_kafka_broker_partition.yaml` to kill a Kafka broker and assert that producers seamlessly redirect to insync replicas without dropping order events.
6. Implement `04_network_latency_spike.yaml` injecting 500ms latency and 15% packet drop between API Gateway and downstream microservices to verify circuit breaker trip (Resilience4j / Envoy outlier detection).
7. Implement `05_besu_validator_partition.yaml` isolating 1 Hyperledger Besu QBFT validator and verifying uninterrupted block creation and transaction settlement.
8. Implement `06_besu_quorum_loss_recovery.yaml` isolating 2 Besu validators, verifying chain halt without fork, restoring network connectivity, and asserting fast catch-up synchronization.
9. Implement `07_redis_master_failover.yaml` triggering Redis primary pod termination and validating that WebSocket subscribers reconnect with zero duplicate trade broadcasts.
10. Develop an automated test orchestrator script (`tests/chaos/scripts/run_chaos_suite.py`) that runs k6 background load (5,000 req/s) while executing Chaos Mesh experiments.
11. Add automated assertion checks that query API error budgets and database integrity hashes before, during, and after each chaos injection.
12. Integrate Chaos Mesh workflows into weekly automated staging game days with Slack/PagerDuty notification webhooks and automated post-mortem metric exports.

## Interfaces / Contracts
```yaml
# Chaos Mesh Experiment: Besu QBFT Validator Network Partition
# File: tests/chaos/experiments/05_besu_validator_partition.yaml

apiVersion: chaos-mesh.org/v1alpha1
kind: NetworkChaos
metadata:
  name: besu-validator-isolation
  namespace: growww-ledger-staging
spec:
  action: partition
  mode: fixed
  value: "1"
  selector:
    namespaces:
 - growww-ledger-staging
    labelSelectors:
      app: "besu-validator"
      validator-index: "3"
  direction: both
  duration: "5m"
  scheduler:
    cron: "0 2 * * 2" # Weekly automated run
```

```yaml
# Chaos Mesh Experiment: PostgreSQL Primary Pod Termination Under Load
# File: tests/chaos/experiments/02_postgres_primary_failover.yaml

apiVersion: chaos-mesh.org/v1alpha1
kind: PodChaos
metadata:
  name: postgres-primary-kill
  namespace: growww-storage-staging
spec:
  action: pod-kill
  mode: fixed
  value: "1"
  selector:
    namespaces:
 - growww-storage-staging
    labelSelectors:
      role: "master"
      app: "postgres-patroni"
  duration: "1m"
```

## Security & Compliance Notes
- Chaos experiments must NEVER run against production databases or live NSDL/CDSL/NPCI production connections.
- Chaos Mesh dashboard and API must be secured with multi-factor authentication (MFA) and restricted to authorized Senior SRE and QA Leads.
- Detailed audit logs of all injected faults, experiment timestamps, and affected microservices must be archived for SEBI/RBI compliance inspection.

## Acceptance Criteria
- [ ] Platform withstands random pod kills across all microservices with zero unhandled 500 errors returned to clients (handled via retries/circuit breakers).
- [ ] PostgreSQL primary failover completes within 25 seconds with 0 data loss and zero corrupted financial ledger entries.
- [ ] Hyperledger Besu QBFT consensus continues uninterrupted when 1 validator fails; recovers cleanly from 2-node partition when connectivity is restored.
- [ ] Kafka producer idempotency guarantees exactly-once processing across broker restarts and partition rebalances.
- [ ] All automated chaos experiments generate detailed post-experiment reports logging RTO, RPO, and SLO impact.

## Suggested Order / Dependencies
- **Prerequisites:** 010 (NFR Targets), 302 (Validator Setup), 401-403 (Data/Kafka), 802 (Kubernetes), 903 (Load Testing).
- **Parallel Tasks:** 905 (Security Testing), 908 (Launch Runbook).
