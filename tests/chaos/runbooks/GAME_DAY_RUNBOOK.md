# Chaos Engineering & Disaster Recovery Game Day Runbook (Prompt 904)

## Regulatory Mandate & Compliance Objectives
In compliance with SEBI Circulars on Business Continuity Planning (BCP) and Disaster Recovery (DR) for Financial Market Infrastructures (FMIs) and Digital Asset Exchanges:
- **Recovery Time Objective (RTO < 30 seconds)**: Total service failover must complete within **< 30 seconds**.
- **Recovery Point Objective (RPO)**: **Zero State Loss (RPO = 0)**. No trade executions, order cancellations, or DvP settlements may be lost or duplicated.
- **ACID Transaction Integrity**: Distributed ledger state and PostgreSQL primary state must maintain cryptographic consistency before, during, and after fault injection.

## Game Day Chaos Scenario Matrix

| Scenario ID | Target Component | Chaos Type | Target Failure | Success Criteria | Max Permissible RTO |
|-------------|------------------|------------|----------------|------------------|---------------------|
| `DR-SCEN-01` | Order Matching Engine | `PodChaos` | Pod termination during peak order burst | Standby replica assumes primary, replays WAL from Kafka/Redis, zero lost orders | < 5 seconds |
| `DR-SCEN-02` | PostgreSQL Primary | `PodChaos` | Hard kill of primary database container | Patroni elects standby replica, promotes to primary, zero corrupted ledger rows | < 20 seconds |
| `DR-SCEN-03` | Apache Kafka Cluster | `NetworkChaos` | Partition 1 of 3 brokers | Producers reroute traffic to in-sync replicas (`min.insync.replicas=2`), zero dropped events | < 3 seconds |
| `DR-SCEN-04` | Ingress / Gateway | `NetworkChaos` | 500ms latency + 15% packet loss | Gateway circuit breaker trips, graceful client throttling, outlier ejection | < 10 seconds |
| `DR-SCEN-05` | Besu QBFT Consensus | `NetworkChaos` | Isolate 1 validator ($N=4$, $F=1$) | Quorum of 3 validators continues block creation at 2s interval without chain halt | 0 seconds (Uninterrupted) |
| `DR-SCEN-06` | Besu Quorum Loss | `NetworkChaos` | Isolate 2 validators ($N=4$) | Chain pauses safely without creating forks; resumes and catches up immediately upon network heal | < 15 seconds after heal |
| `DR-SCEN-07` | Redis Cluster | `PodChaos` | Redis primary crash | Replica promoted to primary, WebSocket fan-out reconnects seamlessly | < 10 seconds |

## Abort & Rollback Triggers
The Chaos Commander must immediately abort the game day drill if any of the following occur:
1. Client HTTP 5xx error rate exceeds 0.1% over a 60-second window.
2. Kafka consumer lag exceeds 5,000 uncommitted messages.
3. Matching engine book depth integrity check reports crossed books or negative quantities.
4. Besu block production ceases for more than 30 seconds on non-quorum-loss scenarios.

## Execution Procedure
1. Verify cluster health: `kubectl get pods -A`
2. Launch background synthetic load (5,000 req/s): `python3 tests/benchmarks/engine/load_harness.py`
3. Execute automated chaos runner: `python3 tests/chaos/scripts/run_chaos_suite.py`
4. Review post-experiment report in `tests/chaos/reports/` and archive for compliance audits.
