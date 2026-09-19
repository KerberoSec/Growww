# 909 - Post-Launch Monitoring, SLOs & Error Budget Management

## Purpose
Defines the continuous operational observability, Service Level Objective (SLO) tracking, and Error Budget management framework for the Growww production ecosystem. Because financial infrastructure demands uncompromising reliability, uptime, and latency standards (including SEBI's mandate for 99.9% availability and swift incident disclosures), this specification formalizes how Service Level Indicators (SLIs) are measured, how error budgets govern feature release velocity, and how multi-window burn-rate alerting detects incidents before customer impact occurs.

## What You Are Building
A production-grade Site Reliability Engineering (SRE) observability and SLO management suite (`ops/observability/slos/`):
- Declarative SLO-as-Code definitions compiled using Sloth / Pyrra targeting Prometheus for all core user journeys.
- Multi-Window Multi-Burn-Rate alerting rules (1h, 6h, 3-day, 14-day burn windows) minimizing alert fatigue and false positives.
- Real-time Grafana SLO & Error Budget executive dashboards tracking 30-day rolling availability and error budget consumption.
- Automated Engineering Release Freeze Policy triggering automated CI/CD deployment freezes when an error budget is depleted.
- Dedicated Hyperledger Besu Blockchain telemetry and validator consensus monitoring pipelines.

## Scope Boundaries
- **In Scope:**
 - Defining SLIs/SLOs for critical services: Order Gateway (99.95% availability, p99 < 50ms), Matching Engine (99.99% availability, p99 < 1ms), DvP Settlement (99.9% success within 3s), WebSocket Streaming (99.9% delivery < 100ms).
 - Multi-window error budget burn rate alerting (Google SRE standard).
 - Hyperledger Besu validator node health, peer count, gas consumption, and consensus round monitoring.
 - Automated deployment gate integration blocking non-emergency releases during budget exhaustion.
 - Quarterly SEBI compliance uptime and incident reporting automation.
- **Out of Scope / Handled Elsewhere:**
 - Initial monitoring stack deployment and log shipper setup (Prompt 806, 807).
 - Launch day war room operations (Prompt 908).
 - Customer ticket handling and support operations (Prompt 910).

## Technology to Use
- **SLO Engine:** Sloth / Pyrra (Kubernetes-native Prometheus SLO generator) - translates human-readable SLO targets into mathematically rigorous Prometheus recording and alerting rules.
- **Metrics Storage & Querying:** Prometheus + Thanos for durable, long-term metric storage and PromQL queries.
- **Dashboards & Alerting:** Grafana 10+ with custom SRE error budget panels + PagerDuty for on-call engineer paging.
- **Distributed Tracing & Logs:** OpenTelemetry collector + Grafana Tempo + Loki for correlation between SLO violations and root-cause spans/logs.

*Justification:* Sloth generates multi-window burn-rate alerts adhering to Google SRE best practices, eliminating alert noise and ensuring actionable on-call pages.

## Backend / Infra Touchpoints
- Production Kubernetes Cluster Prometheus Operator.
- Envoy API Gateway access log metrics (`envoy_http_downstream_rq_time_bucket`).
- Rust Matching Engine Prometheus metrics (`matching_engine_duration_microseconds`).
- PostgreSQL connection pool exporter and slow query logger.
- Kafka exporter (`kafka_consumergroup_lag`).
- Hyperledger Besu metrics exporter (`besu_blockchain_height`, `besu_peers`, `besu_qbft_round`).

## Blockchain Interaction
- Continuously monitors Hyperledger Besu blockchain consensus health:
 - **Block Production SLI:** Validates that a new block is produced every 2.0 ± 0.2 seconds.
 - **Validator Peer Count:** Alerts if any of the 4 validators connects to fewer than 3 peers.
 - **Mempool & Gas SLI:** Monitors pending transaction pool size and block gas limit saturation (> 80%).
 - **Settlement DvP SLI:** Measures on-chain execution error rate for `SettlementDvP.sol` transactions; alerts immediately if transaction revert rate exceeds 0.00% (Zero Fee).
 - **Indexer Lag SLI:** Measures block height difference between Besu chain head and off-chain PostgreSQL event indexer; alerts if lag exceeds 2 blocks.

## Step-by-Step Build Instructions
1. Scaffold the SLO repository structure (`ops/observability/slos/`) with `definitions/`, `alerts/`, `dashboards/`, and `policies/`.
2. Define the core SLI metric expressions in PromQL for Order Placement, Matching, Market Data WebSocket, DvP Settlement, and KYC Processing.
3. Author Sloth SLO definition manifests (`ops/observability/slos/definitions/order_gateway.yaml`, `matching_engine.yaml`, `dvp_settlement.yaml`).
4. Generate the corresponding Prometheus recording and alerting rules using `sloth generate`.
5. Deploy generated alerting rules to the production Prometheus Operator instance via GitOps (ArgoCD).
6. Configure PagerDuty routing rules mapping Tier-1 alerts (2% burn in 1h, 5% burn in 6h) to High-Urgency on-call pages and Tier-2 alerts (10% burn in 3 days) to Low-Urgency Slack notifications.
7. Build the master Grafana Executive SLO Dashboard (`ops/observability/slos/dashboards/executive_slo.json`) displaying rolling 30-day budget bars and current burn rates.
8. Build the Hyperledger Besu Validator & Ledger Telemetry Dashboard (`ops/observability/slos/dashboards/besu_ledger_health.json`).
9. Configure OpenTelemetry automatic trace-to-metric linking allowing engineers to click directly from an SLO breach alert to offending trace spans in Grafana Tempo.
10. Implement the Error Budget Policy Gate script (`ops/observability/slos/policies/check_error_budget.py`) in CI/CD pipeline, blocking non-critical feature deployments if the 30-day error budget is < 10%.
11. Implement the automated SEBI Platform Uptime & Downtime Reporter generating monthly PDF/CSV disclosures showing 99.9% compliance and zero unannounced maintenance.
12. Conduct a simulated SLO breach game-day to test alert routing, PagerDuty escalation policies, and release freeze enforcement.

## Interfaces / Contracts
```yaml
# Sloth SLO Definition: Order Ingestion & Matching Engine
# File: ops/observability/slos/definitions/matching_engine.yaml

version: "sloth.slok.dev/v1"
service: "growww-trading"
slos:
 - name: "order-matching-latency"
    objective: 99.99
    description: "99.99% of order matching executions complete within 1ms"
    sli:
      events:
        error_query: sum(rate(matching_engine_execution_time_bucket{le="1000"}[{{.window}}]))
        total_query: sum(rate(matching_engine_execution_time_count[{{.window}}]))
    alerting:
      name: "MatchingEngineLatencyHigh"
      page_alert:
        labels:
          severity: "critical"
          team: "core-trading"
      ticket_alert:
        labels:
          severity: "warning"
          team: "core-trading"

 - name: "dvp-settlement-success-rate"
    objective: 99.9
    description: "99.9% of DvP atomic settlements succeed on-chain without revert"
    sli:
      events:
        error_query: sum(rate(dvp_settlement_total{status=~"failed|reverted"}[{{.window}}]))
        total_query: sum(rate(dvp_settlement_total[{{.window}}]))
    alerting:
      name: "DvPSettlementFailureSpike"
      page_alert:
        labels:
          severity: "critical"
          team: "settlement"
```

## Security & Compliance Notes
- Adheres to SEBI guidelines mandating minimum 99.9% uptime for digital investment platforms and submission of root cause analysis (RCA) for any outage lasting > 15 minutes.
- Observability pipelines must never record PII, cleartext PAN/Aadhaar numbers, or user bank account numbers in metric labels or log streams.
- Metrics endpoints (`/metrics`) must be internal-only, accessible exclusively by Prometheus scraper agents over mTLS.

## Acceptance Criteria
- [ ] Sloth manifests generate standard Prometheus alerting rules for all core microservices and blockchain components.
- [ ] Multi-window burn-rate alerts trigger accurately under simulated traffic faults without false alarms.
- [ ] Grafana SLO dashboard displays live rolling 30-day error budgets with real-time burn indicators.
- [ ] Hyperledger Besu blockchain consensus metrics and indexer lag are continuously tracked with alerting on validator disconnection.
- [ ] CI/CD release gate automatically blocks production deployments when a service's error budget is exhausted (< 10% remaining).

## Suggested Order / Dependencies
- **Prerequisites:** 010 (NFRs), 205 (Matching Engine), 310 (Node Monitoring), 806 (Observability Stack), 908.
- **Parallel Tasks:** 910 (Customer Grievance Redressal).
