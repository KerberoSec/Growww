# 808 - Mission-Critical Alerting Architecture, SLO Framework & On-Call Runbooks

## Purpose
In high-concurrency financial trading and digital security custody, technical incidents have immediate monetary, legal, and regulatory consequences. Unplanned downtime, an unhandled consensus fork in the permissioned blockchain, or a stalled settlement queue can trigger regulatory penalties from SEBI and financial losses.

This prompt establishes Growww's multi-tier alerting architecture, Service Level Objective (SLO) error-budget monitoring framework, and standardized operational incident runbooks. By integrating Prometheus Alertmanager, Sloth for multi-window burn-rate SLO alerting, PagerDuty / Opsgenie escalation rotations, and automated self-healing remediation hooks, the system guarantees that operational anomalies are identified, escalated, and resolved within strict MTTR (Mean Time to Resolution) limits.

## What You Are Building
A production incident response framework, alerting rules, and operational runbook repository:
- `deployments/k8s/monitoring/alertmanager.yaml`: Prometheus Alertmanager configuration with multi-channel routing (PagerDuty, Slack, Opsgenie, Webhooks), deduplication, and suppression trees.
- `deployments/k8s/monitoring/slo-rules.yaml`: Sloth-generated Prometheus recording and alerting rules based on Google SRE multi-window multi-burn-rate SLO frameworks.
- `deployments/k8s/monitoring/blockchain-alerts.yaml`: Dedicated PrometheusRule alerting for Hyperledger Besu validator node health, consensus stalls, peer drops, and relayer wallet balances.
- `docs/runbooks/`: Standardized, step-by-step Markdown runbooks for high-severity operational incidents:
 - `RB-001_matching_engine_latency_spike.md`: Diagnosis and remediation for order matching queues, lock contention, and memory compaction delays.
 - `RB-002_besu_consensus_stall_and_peer_partition.md`: Recovery procedures for QBFT validator forks, round escalations, and node resynchronization.
 - `RB-003_dvp_settlement_deadlock.md`: Triage steps for atomic delivery-versus-payment execution timeouts and custodian API disconnects.
 - `RB-004_nsdl_cdsl_custodian_gateway_outage.md`: Failover to secondary depository gateway and automated trade queuing during exchange-wide custodian outages.
 - `RB-005_upi_payment_aggregator_failure.md`: Dynamic switching between primary and fallback banking rails upon elevated payment webhook drop rates.
- `services/auto-remediation-runner/`: Lightweight Kubernetes operator executing approved self-healing actions (e.g. draining a desynchronized Besu validator node, recycling deadlocked worker pods).

## Scope Boundaries
- **In Scope:**
 - Alertmanager configuration, routing trees, and escalation tiers (P1-Critical to P4-Info).
 - SLI/SLO mathematical rule definitions and error budget burn rate calculations.
 - Dedicated alerting rules for Kubernetes, Microservices, Databases, Kafka, and Blockchain validator nodes.
 - Incident response runbooks with explicit diagnostic commands and recovery steps.
 - Automated self-healing scripts with safety interlocks.
- **Out of Scope / Handled Elsewhere:**
 - Metrics scraping and OpenTelemetry collector setup (Prompt 806).
 - Incident post-mortem governance and SEBI regulatory disclosure filing (Prompt 706).
 - Continuous deployment rollback automation (Prompt 804).

## Technology to Use
- **Prometheus Alertmanager v0.27+**: Alert grouping, deduplication, and notification routing engine. Justification: Native integration with Prometheus/Mimir, robust inhibition rules (e.g., mute child service alerts if the underlying node/cluster is unreachable), and support for high-availability clustering.
- **Sloth (SLO Generator)**: Industry standard tool generating production-ready Prometheus alerting rules implementing Google SRE multi-window multi-burn-rate algorithms.
- **PagerDuty / Opsgenie API**: Incident management platform providing automated on-call schedules, SMS/phone escalations, and incident timeline tracking.
- **Grafana OnCall**: In-cluster on-call management UI integrated directly with Grafana dashboards.

## Backend / Infra Touchpoints
- **Prometheus / Mimir**: Evaluates alert rule expressions every 15 seconds.
- **Alertmanager Cluster**: High-availability 3-replica cluster in `growww-observability` namespace.
- **PagerDuty / Opsgenie Webhooks**: Incident triggers with severity mapping.
- **Slack Alert Channels**: `#alerts-p1-critical`, `#alerts-p2-high`, `#alerts-trading-ops`, `#alerts-blockchain`.

## Blockchain Interaction
Establishes specialized telemetry monitors and incident runbooks for the Hyperledger Besu QBFT network:
- **Consensus Health Alerts**:
 - `BesuConsensusRoundEscalation`: Triggers when `besu_consensus_qbft_round > 0` for more than 2 consecutive blocks (indicating validator voting timeouts or network delay).
 - `BesuPeerCountCritical`: Triggers when `besu_peers_connected_total < 3` (warning of imminent consortium network partition).
 - `BesuBlockProductionStall`: Triggers when `increase(besu_blockchain_block_number[30s]) == 0` (block production halted).
 - `RelayerWalletLowBalance`: Triggers when relayer gas-paying address balance falls below 0.5 ETH / native token.
 - `ValidatorHsmSigningError`: Triggers immediately on any signing failure between the validator node and CloudHSM / Vault.
- **Runbook RB-002**: Detailed recovery protocol outlining how on-call engineers safely reboot a lagging validator node, force fast-sync from a healthy peer, or invoke emergency multi-sig governance to replace a compromised validator address in the QBFT validator set.

## Step-by-Step Build Instructions
1. Scaffold directory `deployments/k8s/monitoring/` and `docs/runbooks/`.
2. Configure High Availability Alertmanager cluster manifests with gossip communication across 3 Kubernetes nodes.
3. Configure Alertmanager notification receivers: PagerDuty (P1/P2), Slack (P3/P4), and Webhook endpoints for automated remediation runners.
4. Define Alertmanager inhibition rules (e.g., suppress all microservice alerts within a namespace if the underlying Node is in `NodeNotReady` state).
5. Define Sloth SLO specifications for core services (e.g., Order Service Availability 99.99%, Order Matching Latency p99 < 5ms).
6. Generate Prometheus recording and alerting rules from Sloth manifests covering 1h, 6h, 24h, and 3-day error budget burn rates.
7. Write `blockchain-alerts.yaml` defining critical alerting rules for Besu validator nodes, consensus finality, and relayer balances.
8. Write `database-alerts.yaml` covering PostgreSQL connection saturation, replication lag, Redis memory pressure, and Kafka consumer group lag.
9. Author runbooks `RB-001` through `RB-005` detailing Symptoms, Root Cause Possibilities, Diagnostic CLI Commands, Step-by-Step Resolution Steps, and Escalation Paths.
10. Implement the `services/auto-remediation-runner` with role-based Kubernetes service accounts, restricted strictly to restarting non-validator pods or scaling up Kafka worker replicas upon specific alert webhooks.
11. Configure PagerDuty escalation schedules with primary on-call, secondary on-call, and engineering lead escalation tiers.
12. Perform simulated chaos testing: simulate a blocked Kafka queue and verify PagerDuty pages the primary on-call engineer within 60 seconds.
13. Simulate a Besu node peer disconnect and verify `BesuPeerCountCritical` alert routes to the dedicated `#alerts-blockchain` channel and triggers the on-call blockchain engineer.
14. Validate all runbook links in alert notifications navigate directly to the correct runbook documentation page in internal engineering wikis.

## Interfaces / Contracts
```yaml
# deployments/k8s/monitoring/blockchain-alerts.yaml (Excerpt)
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: growww-blockchain-alerts
  namespace: growww-observability
  labels:
    role: alert-rules
spec:
  groups:
 - name: HyperledgerBesuQBFTAlerts
      rules:
 - alert: BesuBlockProductionStall
          expr: increase(besu_blockchain_block_number[30s]) == 0
          for: 10s
          labels:
            severity: P1-Critical
            service: blockchain-validator
            tier: infrastructure
          annotations:
            summary: "Hyperledger Besu Block Production Halted"
            description: "No new blocks produced on the permissioned ledger in the last 30 seconds. Consensus may be stalled."
            runbook_url: "https://wiki.growww.internal/runbooks/RB-002"

 - alert: BesuConsensusRoundEscalation
          expr: max(besu_consensus_qbft_round) > 0
          for: 20s
          labels:
            severity: P2-High
            service: blockchain-validator
          annotations:
            summary: "QBFT Consensus Round Escalation Detected"
            description: "Current QBFT round is {{ $value }}. Validators are experiencing voting delays or network latency."
            runbook_url: "https://wiki.growww.internal/runbooks/RB-002"

 - alert: RelayerWalletLowBalance
          expr: besu_wallet_balance_wei{account="settlement-relayer"} < 500000000000000000
          for: 1m
          labels:
            severity: P2-High
            service: trade-settlement-service
          annotations:
            summary: "Settlement Relayer Account Balance Low"
            description: "Settlement relayer balance is below 0.5 ETH. Refill required to avoid settlement transaction stalls."
            runbook_url: "https://wiki.growww.internal/runbooks/RB-003"
```

## Security & Compliance Notes
- SEBI Incident Reporting Compliance: Major outages impacting order placement, matching, or settlement must be recorded in the incident management log within 30 minutes to facilitate mandatory SEBI / exchange reporting.
- Audit Logging of Remediation: Any automated remediation script execution or manual on-call command execution is automatically logged to the immutable audit trail (Prompt 807).
- Secret Security in Alertmanager: All PagerDuty and Slack API tokens are injected dynamically from HashiCorp Vault / Kubernetes Secrets.

## Acceptance Criteria
- [ ] Alertmanager cluster routes P1/P2 alerts to PagerDuty with guaranteed SMS/phone delivery within 60 seconds.
- [ ] Sloth multi-window burn-rate SLO alerts accurately detect fast and slow error budget consumption without false positives.
- [ ] Besu blockchain alerts detect block stalls, round escalations, and peer partitions, alerting the blockchain on-call team.
- [ ] All 5 core operational runbooks (`RB-001` through `RB-005`) are documented with copy-pasteable diagnostic commands and clear resolution steps.
- [ ] Automated remediation runner safely executes approved restart/scale actions upon verified webhook signatures.

## Suggested Order / Dependencies
- Prerequisites: Prompt 010 (NFRs Master Doc), Prompt 806 (Observability Stack), Prompt 807 (Centralized Logging).
- Parallel Tasks: Prompt 809 (Cost Optimization), Prompt 810 (Release Management).
