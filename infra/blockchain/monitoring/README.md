# Hyperledger Besu Consortium Monitoring, Telemetry & Alerting Stack

## Architecture Overview
This module (`infra/blockchain/monitoring/`) provides full-stack observability for the Growww permissioned Hyperledger Besu EVM settlement ledger operating under Quorum Byzantine Fault Tolerance (QBFT).

### Monitored Components & Port Allocations
| Component | Metric Port | Path | Scrape Interval | Description |
| :--- | :--- | :--- | :--- | :--- |
| **Besu Validators** (4 nodes) | `9545` | `/metrics` | `5s` | Native Besu Prometheus metrics (Consensus, Peers, RocksDB, JVM) |
| **Besu RPC Relayers** (2 nodes) | `9545` | `/metrics` | `5s` | Ingress transaction pool & JSON-RPC worker queues |
| **Besu Observers** (2 nodes) | `9545` | `/metrics` | `10s` | Archive node synchronization and compliance query state |
| **Web3Signer Sidecars** (4 instances) | `9001` | `/metrics` | `5s` | HSM signing latency and signature error counters |
| **Host Node Exporters** | `9100` | `/metrics` | `15s` | NVMe disk I/O, CPU, RAM, and network throughput |
| **Consensus Health Prober** | `8080` | `/metrics` | `5s` | Synthetic zero-value JSON-RPC active probing daemon |

---

## Deliverables Summary

1. **`prometheus.yml`**: Prometheus configuration defining high-frequency 5s scrape jobs, relabeling, and Alertmanager routing.
2. **`alerts.yml`**: Alertmanager rules defining critical P1 consensus stall conditions, validator downtime, height divergence, low peer count, and RocksDB storage bottlenecks.
3. **`dashboard.json`**: Institutional Grafana dashboard visualizing live block production, height divergence, peer topology, JVM heap, and Web3Signer latency.
4. **`docs/ops/runbook_consensus_stall.md`**: Operational incident runbook for QBFT consensus stalls, quorum recovery ($N=4, F=1, Q=3$), and SEBI compliance reporting.
5. **`prober/prober.py`**: Active synthetic health prober daemon polling JSON-RPC methods and producing status JSON and Prometheus metrics.
6. **`prober/prober_test.py`**: Unit test suite verifying prober logic, quorum evaluation, and metric generation.
7. **`alertmanager/alertmanager.yml`**: Alertmanager routing dispatching P1 alerts to PagerDuty and P2 alerts to Slack.

---

## Running the Health Prober Unit Tests

```bash
python3 infra/blockchain/monitoring/prober/prober_test.py
```

---

## Security & Compliance
- **Zero PII in Telemetry:** Metric labels and logs do not contain user wallet addresses, transaction payloads, or customer identities.
- **Metrics Network Isolation:** Port `9545` is bound strictly to private VPC/pod IPs and is never exposed to public internet ingress.
- **7-Year Retention:** Long-term metrics are exported to Thanos / VictoriaMetrics for regulatory audit trails under SEBI guidelines.
