# 310 - Permissioned Blockchain Node Observability & Health Monitoring

## Purpose
The Growww permissioned ledger is mission-critical financial market infrastructure. A failure in QBFT consensus, validator network partition, or transaction pool congestion could delay settlement execution, violate SEBI settlement SLAs, or impair proof-of-reserve attestation. Full-stack observability into validator node health, consensus rounds, peer connectivity, and RocksDB storage performance is mandatory for maintaining a 99.99% availability SLA.

This prompt specifies the design, configuration, and deployment of the **Blockchain Node Monitoring, Telemetry, and Alerting Stack** (`infra/blockchain/monitoring/`). The system scrapes Hyperledger Besu native Prometheus metrics, executes active JSON-RPC synthetic health probes, renders institutional Grafana operational dashboards, and configures automated PagerDuty/Slack alert rules for consensus round timeouts, peer count drops, disk exhaustion, and validator desynchronization.

## What You Are Building
A comprehensive monitoring and alerting infrastructure suite containing:
- Prometheus Scrape & Rule Configurations: Scrapes Besu internal metrics (`/metrics`), node exporters, and Web3Signer sidecars across all consortium validators.
- Institutional Grafana Dashboards (`dashboards/besu_consortium_overview.json` and `dashboards/validator_deep_dive.json`): Visualizing live block time, TPS, QBFT consensus rounds, peer topology, memory pools, and RocksDB I/O latency.
- Synthetic Health Prober Daemon (Python / Go): Submits periodic zero-value probe calls (`eth_blockNumber`, `eth_syncing`, `net_peerCount`, `qbft_getValidatorsByBlockNumber`) to verify JSON-RPC responsiveness.
- Alerting Rules (`alerting/rules_besu.yml`): Alert definitions covering critical conditions (Consensus Stalled, Validator Missing in Proposal, Peer Count Below Threshold, High Tx Pool Memory).

## Scope Boundaries
- **In Scope:**
 - Prometheus metric collection from Hyperledger Besu nodes (v24.x+).
 - Monitoring QBFT consensus rounds, commit delays, and round changes.
 - Tracking validator peer connectivity and network partition detection.
 - Monitoring RocksDB database size, compaction times, and disk read/write throughput.
 - Alertmanager routing to PagerDuty (high severity) and Slack/Opsgenie (warnings).
- **Out of Scope / Handled Elsewhere:**
 - Kubernetes cluster-wide node monitoring (handled in Prompt 806).
 - Smart contract transaction indexing and event scraping (handled in Prompt 309).
 - General backend microservices tracing and logs (handled in Prompt 807).

## Technology to Use
- **Metrics Collection:** **Prometheus v2.50+ / VictoriaMetrics**.
  *Justification:* Besu natively exposes comprehensive metrics in Prometheus format on port `9545` (`--metrics-enabled=true`, `--metrics-category=ETH,BLOCKCHAIN,PEER,JVM,PROCESS,RPC,TRANSACTION,QBFT`).
- **Dashboarding:** Grafana v10.x+.
- **Alerting:** Prometheus Alertmanager with PagerDuty and Slack webhooks.
- **Probing Engine:** Go / Python containerized cron prober.

## Backend / Infra Touchpoints
- **Hyperledger Besu Nodes:** Scrapes all 4 institutional validators, 2 RPC relayers, and observer nodes.
- **Web3Signer Sidecars:** Monitors HSM signing latency and signer health endpoints.
- **Kubernetes Ingress & Prometheus Operator:** ServiceMonitors and PodMonitors in the `blockchain-system` namespace.

## Blockchain Interaction
- **Scraped Metrics Categories:**
 - `besu_blockchain_chain_head_block_number`: Current block height per node.
 - `besu_qbft_round_number`: Consensus round number (values $>0$ indicate consensus timeouts or failed primary proposals).
 - `besu_peers_current`: Number of connected peers per validator.
 - `besu_executors_worker_work_queue`: JSON-RPC worker pool queue depth.
 - `besu_transaction_pool_transactions_count`: Total pending transactions in memory pool.
- **Zero PII:** Metrics collect purely operational telemetry (block numbers, byte sizes, millisecond latencies, peer counts). No user identities, addresses, or transaction contents are recorded in metrics.

## Step-by-Step Build Instructions
1. Scaffold repository directory under `infra/blockchain/monitoring/` with subdirectories: `prometheus/`, `grafana/dashboards/`, `prober/`, and `alertmanager/`.
2. Enable Besu metrics flags on all validator and RPC nodes: `--metrics-enabled=true`, `--metrics-host="0.0.0.0"`, `--metrics-port=9545`, `--metrics-category=BLOCKCHAIN,JVM,PROCESS,PEER,RPC,TRANSACTION,QBFT`.
3. Configure Prometheus `ServiceMonitor` definitions to discover Besu pods dynamically in Kubernetes (`besu-validators`, `besu-rpc`, `besu-observers`).
4. Configure scrape intervals: 5-second interval for consensus and block metrics; 15-second interval for JVM and OS metrics.
5. Create Grafana Dashboard 1: **Consortium Executive Overview**:
 - Live Network Block Height across all validators (identifies chain divergence immediately).
 - Instantaneous TPS (Transactions Per Second) and Block Utilization %.
 - Block Propagation Latency heatmap across validator nodes.
 - P2P Mesh Topology matrix (peer counts per node).
6. Create Grafana Dashboard 2: **Validator Node Deep Dive**:
 - JVM Heap Memory utilization and Garbage Collection (G1GC) pause times.
 - RocksDB storage size, read/write IOPS, and compaction duration.
 - QBFT Consensus Round change counter and validator proposal distribution.
 - JSON-RPC request rate and p99 response latency per method (`eth_call`, `eth_sendRawTransaction`).
7. Write alert rule `rules_besu.yml`:
 - `BesuConsensusStalled`: Alert if `besu_blockchain_chain_head_block_number` does not increase for $>10$ seconds (Critical -> PagerDuty).
 - `BesuValidatorDesynced`: Alert if any single validator falls $>5$ blocks behind consortium median height (Warning -> Slack).
 - `BesuPeerCountLow`: Alert if connected peers $< 3$ on any validator node (Critical -> PagerDuty).
 - `BesuQBFTRoundChange`: Alert if `besu_qbft_round_number > 0` (indicates round timeout / leader proposal failure).
 - `BesuDiskSpaceLow`: Alert if RocksDB persistent volume disk usage exceeds 85% (Warning).
8. Implement synthetic health check prober script in Go (`cmd/prober/main.go`):
 - Query all nodes every 2 seconds.
 - Calculate inter-node block height delta.
 - Verify that all active validators are included in the latest block extra-data signature field.
 - Expose custom Prometheus metrics `growww_consensus_health_status{cluster="prod"}`.
9. Deploy Prometheus Operator, Alertmanager, and Grafana in the Kubernetes management cluster.
10. Configure Alertmanager notification routes and templates formatting critical alerts with diagnostic runbook links.
11. Execute chaos testing: simulate taking 1 validator offline; verify `BesuPeerCountLow` and `BesuValidatorDesynced` alerts trigger within 30 seconds.
12. Simulate JSON-RPC overload on RPC relay node; verify worker queue alert fires.
13. Document operational troubleshooting playbooks (`docs/ops/runbook_consensus_stall.md`).

## Interfaces / Contracts

### Prometheus Alert Rules Configuration (`rules_besu.yml`)
```yaml
groups:
 - name: BesuConsortiumAlerts
    rules:
 - alert: BesuConsensusStalled
        expr: increase(besu_blockchain_chain_head_block_number{job="besu-validators"}[30s]) == 0
        for: 10s
        labels:
          severity: critical
          team: blockchain-ops
        annotations:
          summary: "Besu QBFT Consensus Stalled in Production"
          description: "No new blocks produced on validator {{ $labels.instance }} for more than 10 seconds. Network may be partitioned or halted."
          runbook_url: "https://ops.growww.internal/runbooks/besu-consensus-stall"

 - alert: BesuValidatorHeightDivergence
        expr: (max(besu_blockchain_chain_head_block_number) - min(besu_blockchain_chain_head_block_number)) > 3
        for: 15s
        labels:
          severity: warning
          team: blockchain-ops
        annotations:
          summary: "Validator Block Height Divergence Detected"
          description: "One or more Besu validators are lagging behind the consortium head by more than 3 blocks."

 - alert: BesuPeerCountCriticallyLow
        expr: besu_peers_current{job="besu-validators"} < 3
        for: 20s
        labels:
          severity: critical
          team: blockchain-ops
        annotations:
          summary: "Validator Lost P2P Peer Connectivity"
          description: "Validator {{ $labels.instance }} has fewer than 3 connected peers. Risk of losing consensus quorum."

 - alert: BesuQBFTRoundChangeTriggered
        expr: increase(besu_qbft_round_number[1m]) > 0
        for: 0s
        labels:
          severity: warning
          team: blockchain-ops
        annotations:
          summary: "QBFT Consensus Round Change Triggered"
          description: "Validator leader timed out proposing a block, causing round change to next validator in round-robin."
```

### Prober Health Status JSON Output Schema
```json
{
  "timestamp": "2026-09-18T10:15:30Z",
  "network_status": "HEALTHY",
  "chain_id": 13370,
  "consensus_algorithm": "QBFT",
  "latest_block_number": 4829104,
  "block_time_seconds": 2.0,
  "validators_total": 4,
  "validators_online": 4,
  "nodes": [
    { "name": "val-growww-core", "block": 4829104, "peers": 5, "status": "SYNCED" },
    { "name": "val-custodian-nsdl", "block": 4829104, "peers": 5, "status": "SYNCED" },
    { "name": "val-clearing-corp", "block": 4829104, "peers": 5, "status": "SYNCED" },
    { "name": "val-gift-city", "block": 4829104, "peers": 5, "status": "SYNCED" }
  ]
}
```

## Security & Compliance Notes
- **Metrics Network Isolation:** The Prometheus metrics port (`9545`) is bound to private Kubernetes pod IPs and never exposed via public ingress or load balancers.
- **Zero Sensitive Data in Telemetry:** Metric labels and logs do not contain user wallet addresses, transaction payloads, or customer IDs.
- **SEBI/Auditor SLA Reporting:** Monitoring data is archived in long-term metric storage (VictoriaMetrics / Thanos) for 7 years to prove historical system uptime and settlement availability during regulatory audits.

## Acceptance Criteria
- [ ] Prometheus successfully scrapes metrics from all Besu validators and RPC relay nodes.
- [ ] Grafana Consortium Overview and Validator Deep Dive dashboards render live data with refresh rates $\le 5\text{s}$.
- [ ] Automated alerting tested: simulated node crash triggers PagerDuty incident within 30 seconds.
- [ ] Consensus stall detector triggers reliably when block production halts.
- [ ] Operational runbooks documented and linked directly in alert payloads.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `302` (Network Topology & Validator Setup), Prompt `806` (Observability Stack).
- **Parallel Tasks:** Prompt `311` (Validator Key Management HSM), Prompt `808` (Alerting & On-Call Runbooks).
- **Subsequent Prompts Enabled:** Prompt `314` (Chain Disaster Recovery & Backup).
