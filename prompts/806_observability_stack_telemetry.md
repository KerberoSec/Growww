# 806 - Full-Stack Observability, OpenTelemetry & Distributed Tracing

## Purpose
Operating an institutional-grade investment platform combining fractional equity trading, sub-millisecond order matching, real-time banking rails, and a permissioned blockchain ledger requires deep, continuous visibility across all distributed components. A microsecond bottleneck in the matching engine or a delayed block finality confirmation in the settlement contract directly impacts financial integrity and regulatory compliance.

This prompt establishes the end-to-end observability and distributed tracing architecture for Growww. By deploying a unified telemetry pipeline powered by OpenTelemetry (OTel), Prometheus / Grafana Mimir for metrics, Grafana Tempo for distributed tracing, and Grafana Loki for logs, engineering teams gain instant root-cause analysis across mobile clients, backend microservices, Kafka event streams, and Hyperledger Besu validator nodes.

## What You Are Building
A complete observability stack, OpenTelemetry instrumentation framework, and production dashboards:
- `deployments/k8s/observability/otel-collector.yaml`: OpenTelemetry Collector daemonset and gateway configuration with batching, tail-sampling, and PII-redaction processors.
- `deployments/helm/observability-stack/`: Helm deployment specifications for Prometheus / Grafana Mimir, Grafana Tempo, Grafana Loki, and Grafana 11+.
- `deployments/grafana/dashboards/`: Production Grafana dashboard definitions as code (JSON/Grafonnet):
 - `01-executive-overview.json`: Platform throughput, total active users, trade volume in INR, system availability SLO.
 - `02-order-matching-engine-latency.json`: Sub-millisecond matching latency histograms (p50, p90, p99, p99.9), order book depth, cancellation rates.
 - `03-settlement-dvp-pipeline.json`: End-to-end DvP settlement duration, custodian confirmation lag, on-chain transaction confirmation latency.
 - `04-besu-blockchain-consensus.json`: QBFT consensus rounds, block production interval, peer count, transaction pool depth, validator voting round latency.
- `packages/observability/`: Shared OpenTelemetry SDK wrappers for Rust, Go, Python, and Flutter/Dart providing standardized W3C TraceContext propagation, metrics export, and span lifecycle hooks.

## Scope Boundaries
- **In Scope:**
 - OpenTelemetry Collector pipeline configuration (OTLP ingestion via gRPC/HTTP).
 - Distributed tracing infrastructure with Grafana Tempo and W3C trace context header propagation (`traceparent`, `tracestate`).
 - Metrics scraping and long-term storage via Prometheus / Grafana Mimir.
 - Golden signals / RED (Rate, Errors, Duration) and USE (Utilization, Saturation, Errors) dashboards for all services and blockchain nodes.
 - Mobile client telemetry ingestion with automatic PII filtering.
- **Out of Scope / Handled Elsewhere:**
 - Regulatory immutable WORM log archiving and SIEM integration (Prompt 807).
 - PagerDuty incident escalation policies and alerting rules (Prompt 808).
 - FinOps cloud cost analytics (Prompt 809).

## Technology to Use
- **OpenTelemetry Collector Contrib (v0.100+)**: Vendor-neutral telemetry proxy. Justification: Unifies ingestion of metrics, traces, and logs, standardizes semantic conventions, and enables in-flight PII scrubbing via transform processors before data leaves cluster memory.
- **Prometheus & Grafana Mimir**: Horizontally scalable, multi-tenant time-series metrics storage with long-term retention.
- **Grafana Tempo**: High-scale, cost-effective distributed tracing backend backed by object storage (S3).
- **Grafana Loki**: Log aggregation system indexing metadata labels rather than full text, providing high throughput at low storage cost.
- **Grafana 11+**: Unified visualization platform with role-based access control and dashboard-as-code automation.

## Backend / Infra Touchpoints
- **Microservices**: Exporting traces and metrics over OTLP gRPC (`port 4317`) and HTTP (`port 4318`).
- **Kafka Cluster**: JMX metrics exporter capturing consumer group lag, partition offsets, and byte throughput.
- **PostgreSQL / Redis**: `postgres_exporter` and `redis_exporter` scraping connection pools, lock waits, and cache hit ratios.
- **Hyperledger Besu Nodes**: Scraping Besu Prometheus endpoint (`http://besu-node:9545/metrics` or `8545/metrics`).

## Blockchain Interaction
Captures and visualizes real-time consensus telemetry, validator node health, and smart contract execution metrics:
- **Besu Prometheus Scraping**: Scrapes core consensus metrics:
 - `besu_blockchain_block_number`: Current chain height.
 - `besu_consensus_qbft_round`: Current QBFT consensus voting round (an elevation above 0 indicates validator voting delay or network partition).
 - `besu_peers_connected_total`: Number of active authenticated consortium validator peers.
 - `besu_transaction_pool_transactions`: Unconfirmed pending transaction queue size.
 - `besu_blockchain_difficulty_total`: QBFT total block difficulty.
- **Distributed Trace Correlation**: When an off-chain settlement worker submits a transaction to `SettlementDvP.sol`, the OpenTelemetry `trace_id` is passed as an indexed event parameter in Solidity or logged in the relayer receipt. The event indexing service (Prompt 309) correlates the on-chain `tx_hash` with the originating user order `trace_id`, enabling end-to-end trace visualization from mobile tap to on-chain block confirmation in Grafana Tempo.

## Step-by-Step Build Instructions
1. Scaffold directory `deployments/k8s/observability/` and `deployments/grafana/dashboards/`.
2. Deploy the observability storage backends (Mimir, Tempo, Loki) using Helm charts with S3 object storage backends.
3. Configure the OpenTelemetry Collector Contrib as a Kubernetes `DaemonSet` on every node and a scalable `Deployment` gateway.
4. Implement OTel Collector pipelines with `batch`, `memory_limiter`, and `transform` processors (for scrubbing Indian PAN numbers, Aadhaar IDs, and bank account numbers).
5. Build the Go, Rust, and Python OpenTelemetry SDK helper packages standardizing tracer initialization and auto-instrumenting gRPC/HTTP clients.
6. Configure W3C TraceContext propagation across HTTP headers and Kafka message headers (`X-B3-TraceId` and `traceparent`).
7. Instrument the Rust Order Matching Engine to emit sub-millisecond execution duration histograms and order book depth gauges.
8. Instrument the Go Trade Settlement Service to record spans for NSDL custody lock, banking debit, and Besu on-chain settlement execution.
9. Configure Prometheus scrape configs for PostgreSQL, Redis, Kafka, and Hyperledger Besu validator nodes (`/metrics` on port 9545).
10. Build the `04-besu-blockchain-consensus.json` Grafana dashboard tracking block time intervals, QBFT rounds, and peer topology.
11. Build the `02-order-matching-engine-latency.json` Grafana dashboard displaying p50, p90, p99, and p99.9 latency percentiles.
12. Configure Grafana data sources connecting Mimir, Tempo, and Loki with trace-to-log and log-to-trace cross-navigation links.
13. Deploy synthetic canary probes simulating periodic order placements and health checks from multiple simulated geographic regions.
14. Perform end-to-end trace verification: place an order on the Flutter client and verify the complete trace graph displays in Grafana Tempo spanning API Gateway -> Order Service -> Matching Engine -> Kafka -> Settlement Service -> Besu Node.

## Interfaces / Contracts
```yaml
# deployments/k8s/observability/otel-collector.yaml (Excerpt)
apiVersion: v1
kind: ConfigMap
metadata:
  name: otel-collector-config
  namespace: growww-observability
data:
  config.yaml: |
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
          http:
            endpoint: 0.0.0.0:4318
      prometheus:
        config:
          scrape_configs:
 - job_name: 'besu-validators'
              scrape_interval: 2s
              static_configs:
 - targets: ['besu-validator-0.growww-blockchain:9545', 'besu-validator-1.growww-blockchain:9545']

    processors:
      memory_limiter:
        check_interval: 1s
        limit_percentage: 75
        spike_limit_percentage: 20
      batch:
        send_batch_size: 1024
        timeout: 1s
      transform:
        error_mode: ignore
        log_statements:
 - context: log
            statements:
              # Redact Indian PAN (5 letters, 4 digits, 1 letter)
 - replace_all_patterns(attributes, "value", "[A-Z]{5}[0-9]{4}[A-Z]{1}", "[REDACTED_PAN]")
              # Redact Aadhaar (12 digits)
 - replace_all_patterns(attributes, "value", "\\b[0-9]{12}\\b", "[REDACTED_AADHAAR]")

    exporters:
      otlp/tempo:
        endpoint: tempo-distributor.growww-observability:4317
        tls:
          insecure: true
      prometheus/mimir:
        endpoint: 0.0.0.0:8889
      loki:
        endpoint: http://loki-gateway.growww-observability/loki/api/v1/push

    service:
      pipelines:
        traces:
          receivers: [otlp]
          processors: [memory_limiter, transform, batch]
          exporters: [otlp/tempo]
        metrics:
          receivers: [otlp, prometheus]
          processors: [memory_limiter, batch]
          exporters: [prometheus/mimir]
        logs:
          receivers: [otlp]
          processors: [memory_limiter, transform, batch]
          exporters: [loki]
```

## Security & Compliance Notes
- PII Redaction in Flight: All OpenTelemetry collector pipelines strictly enforce in-flight regex redaction for Aadhaar numbers, PAN cards, bank account numbers, and investor mobile numbers before traces/logs are written to disk.
- Encryption in Transit: All OTLP telemetry transport within the Kubernetes cluster is secured via Cilium WireGuard or TLS.
- RBAC in Grafana: Access to Grafana dashboards is integrated with Okta / Google Workspace OIDC; production trading and blockchain dashboards require Level 2 Ops or SecOps authorization.

## Acceptance Criteria
- [ ] OpenTelemetry Collector successfully receives OTLP traces, metrics, and logs on ports 4317/4318.
- [ ] Distributed trace context successfully propagates from API Gateway across Kafka and settlement services to on-chain transaction hashes.
- [ ] Besu blockchain validator metrics (QBFT round, block height, peer count) are scraped every 2 seconds and rendered on Grafana.
- [ ] In-flight PII redaction verified: simulated PAN and Aadhaar strings in log/trace payloads are replaced with `[REDACTED_*]`.
- [ ] Matching engine p99 latency metrics accurately reflect real-time microsecond-level performance.

## Suggested Order / Dependencies
- Prerequisites: Prompt 101 (Architecture Overview), Prompt 103 (API Standards), Prompt 802 (Kubernetes Architecture).
- Parallel Tasks: Prompt 807 (Centralized Logging & SIEM), Prompt 808 (Alerting & On-Call).
