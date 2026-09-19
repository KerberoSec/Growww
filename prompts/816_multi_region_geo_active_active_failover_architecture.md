# 816 - Multi-Region Geo Active-Active Disaster Recovery & Cross-DC Consensus Architecture

## Purpose
In a regulated securities exchange, fractional equity platform, and central counterparty settlement ecosystem, regional infrastructure disruptions (such as metropolitan power grid blackouts, undersea optical fiber cuts, severe seismic events, or localized cyber warfare) represent critical systemic risks. Unplanned downtime or data divergence can destabilize market trust, trigger financial imbalances, and violate statutory capital market regulations.

Under the Securities and Exchange Board of India (SEBI) Business Continuity Plan (BCP) and Disaster Recovery (DR) regulations (SEBI Circulars SEBI/HO/MRD/DMS/CIR/P/2019/43, SEBI/HO/MRD1/DT/CIR/P/2021/539, and subsequent directives for Market Infrastructure Institutions and Qualified Stock Brokers), regulated financial market infrastructure must maintain an alternate Disaster Recovery (DR) site located in a distinct seismic zone at least 250 to 500 kilometers away from the Primary Data Centre (DC). The regulatory framework mandates a Recovery Point Objective (RPO) of zero (zero data loss) for all trade, transaction, and custody records, alongside a Recovery Time Objective (RTO) measured in seconds (sub-second target for Growww's high-frequency matching and settlement fabric).

This prompt specifies the architecture, orchestration, and implementation of Growww's Multi-Region Geo Active-Active Disaster Recovery and Cross-DC Consensus Architecture (`infra/geo-active-active/`). The system bridges Primary Data Centre (DC - Mumbai, AWS ap-south-1 and Equinix MB1/MB2 Colo) and Disaster Recovery Data Centre (DR - GIFT City IFSC / Hyderabad, AWS ap-south-2 and CtrlS Colo) with an independent Cloud Witness site (Bangalore/Chennai). It establishes a zero-RPO synchronous replication mesh, sub-second RTO automated ingress failover, a cross-DC Hyperledger Besu QBFT validator consensus quorum, and automated execution harnesses for mandatory quarterly SEBI DR drills.

## What You Are Building
An enterprise-grade, geographically distributed active-active disaster recovery and consensus infrastructure (`infra/geo-active-active/`):
- **Dedicated Low-Latency Interconnect Fabric:** Dual redundant, carrier-neutral 10 Gbps Direct Connect circuits and dark fiber leased lines connecting Mumbai and GIFT City / Hyderabad with Layer 2 MACsec (802.1AE) hardware encryption, achieving sustained round-trip time (RTT) under 12 milliseconds.
- **Multi-Cluster Kubernetes Federation:** Cross-region Kubernetes cluster topology joined via Cilium Cluster Mesh with eBPF-based transparent WireGuard encryption, providing seamless pod-to-pod flat network routing and cross-DC global service discovery.
- **Cross-DC Synchronous PostgreSQL Patroni Cluster:** High-availability database cluster spanning Mumbai (DC), GIFT City / Hyderabad (DR), and Witness nodes with synchronous replication (`synchronous_commit = on`), distributed etcd consensus, and zero-data-loss automated leader promotion.
- **Active-Active Apache Kafka MirrorMaker 2 Fabric:** Symmetrical Kafka clusters deployed in both data centers with bidirectional MirrorMaker 2 replication, consumer offset translation, and header-based anti-cycling filters.
- **Distributed Hyperledger Besu QBFT Validator Quorum:** 7-node geographically dispersed validator set (3 Mumbai, 3 GIFT City / Hyderabad, 1 Witness) guaranteeing single-block deterministic finality and surviving total regional loss without chain split-brain or halt.
- **Active-Active Redis Cache Mesh:** Multi-region Redis cluster using Conflict-Free Replicated Data Types (CRDTs) and active-standby cross-region replication for session state, API tokens, and rate limits.
- **BGP Anycast & Automated DNS Ingress Routing:** Edge routing fabric utilizing BGP Anycast route health injection and AWS Route53 Application Recovery Controller (ARC) / Cloudflare GSLB routing controls for sub-second traffic re-direction.
- **Automated Failover Controller Daemon (`infra/failover-controller`):** Control plane daemon continuously evaluating cross-DC heartbeats, network split conditions, database replication lag, and matching engine sequence parity, executing automated STONITH fencing and failover state promotion.
- **SEBI BCP/DR Quarterly Drill Automation Suite:** Automated orchestration harness executing unannounced live failover switchover drills during active trading hours, recording cryptographic audit logs, and generating statutory compliance verification reports.

## Scope Boundaries
- **In Scope:**
  - Dedicated network interconnect architecture (Direct Connect, dark fiber, MACsec Layer 2 encryption).
  - Multi-cluster Kubernetes federation and cross-region Cilium Cluster Mesh.
  - Multi-region PostgreSQL Patroni synchronous replication and automated failover.
  - Apache Kafka MirrorMaker 2 active-active topic synchronization and offset translation.
  - Hyperledger Besu cross-DC QBFT validator quorum distribution and dynamic validator re-voting.
  - Redis multi-region active-active caching and split-brain resolution.
  - Edge BGP Anycast routing, health check probes, and automated DNS GSLB failover policies.
  - Automated Failover Controller daemon, STONITH fencing agents, and state reconciliation runbooks.
  - SEBI BCP/DR quarterly drill automation and compliance report generation.
- **Out of Scope / Handled Elsewhere:**
  - Matching Engine in-memory ring-buffer WAL replication and shadow engine failover (handled in Prompt 246).
  - Zero-knowledge proof generation and verification infrastructure (handled in Prompts 720 and 721).
  - Long-term tiered cold storage and S3 Glacier compliance archiving (handled in Prompt 413).
  - AWS CloudHSM cluster synchronization and PKCS#11 key generation (handled in Prompt 717).
  - GitOps canary deployment workflows within regional clusters (handled in Prompt 804).

## Technology to Use
- **Network Interconnect & Hardware Encryption:** Dual 10 Gbps AWS Direct Connect and carrier-neutral dark fiber leased lines with IEEE 802.1AE MACsec hardware-level link encryption, sustaining round-trip latency under 12 ms between Mumbai and GIFT City / Hyderabad.
- **Multi-Cluster Orchestration & Service Mesh:** **Kubernetes 1.30+** managed via **Cilium (eBPF)** with Cilium Cluster Mesh, providing cross-cluster pod-to-pod routing, transparent WireGuard node-to-node encryption, and cluster-aware CoreDNS global service lookup (`*.global`).
- **Distributed Relational Database:** **PostgreSQL 16** orchestrated by **Patroni 3.3+** with **Etcd 3.5+** distributed consensus across 3 failure domains, enforcing `synchronous_commit = on` with `synchronous_standby_names = 'FIRST 1 (standby_dr, standby_witness)'`.
- **Distributed Event Bus:** **Apache Kafka 3.7+ (KRaft mode)** and **Kafka MirrorMaker 2 (MM2)** for cross-DC active-active event streaming, partition replication, and consumer group offset synchronization.
- **Consortium Blockchain Quorum:** **Hyperledger Besu 24.1+** running QBFT consensus across a 7-node validator topology with private TLS P2P peering and dynamic validator election.
- **In-Memory Cache & State Store:** **Redis 7.2+ Cluster / KeyDB Active-Active** leveraging Conflict-Free Replicated Data Types (CRDTs) for active-active multi-region key-value synchronization.
- **Global Ingress & Edge Traffic Management:** **BGP Anycast** via BIRD / FRRouting (FRR) peering with Tier-1 upstream transit providers, integrated with **AWS Route53 Application Recovery Controller (ARC)** and Cloudflare Enterprise GSLB.
- **Distributed Observability:** **Prometheus**, **Thanos** (multi-cluster cross-querying), and **OpenTelemetry** for cross-DC replication latency, block height divergence, and RTO/RPO SLA compliance tracking.

## Backend / Infra Touchpoints
- **Matching Engine Hot-Warm Shadow Failover (Prompt 246):** Direct Connect leased lines transport memory-mapped WAL records and order sequence streams from Mumbai primary matching engine to GIFT City / Hyderabad warm shadow engine in real time (<50 ms shadow promotion time).
- **Microservices Deployment Topology (`services/*`):** Symmetrical deployments across Mumbai and DR Kubernetes clusters. Services operate in active-active read mode and active-standby write mode for stateful state transitions.
- **Relational Data Tier:** Core services connect to local PgBouncer connection pools. PgBouncer routes write transactions across the low-latency Direct Connect link to the active PostgreSQL primary, while local read replicas serve low-latency read queries within each region.
- **Event Streaming Bus:** Microservices publish and consume from their local Kafka cluster. MirrorMaker 2 continuously mirrors events across regions, preserving message ordering and header metadata.
- **Blockchain RPC Gateways:** Regional JSON-RPC Besu nodes provide local query caching for read operations, while state-modifying transactions (`eth_sendRawTransaction`) are routed across the consensus validator network.
- **Edge Ingress Proxies:** Envoy-based ingress gateways deployed in both DCs advertise identical BGP Anycast IP prefixes to local Internet Exchanges (IXs: NIXI, Extreme IX, Mumbai IX).

## Blockchain Interaction
Growww's Hyperledger Besu consortium network is designed to survive total regional loss without halting ledger consensus or suffering fork divergence:
- **Geographically Dispersed Validator Quorum:** A 7-node QBFT validator set is deployed across three independent failure domains:
  - Failure Domain 1 (Primary DC - Mumbai): 3 Validator Nodes (`val-mum-01`, `val-mum-02`, `val-mum-03`) distributed across distinct availability zones / racks.
  - Failure Domain 2 (Disaster Recovery DC - GIFT City / Hyderabad): 3 Validator Nodes (`val-dr-01`, `val-dr-02`, `val-dr-03`) distributed across distinct zones / racks.
  - Failure Domain 3 (Neutral Cloud Witness - Bangalore / Chennai): 1 Validator Node (`val-wit-01`) functioning as an independent consensus tiebreaker.
- **QBFT Consensus Safety & Liveness Invariants:**
  - In a QBFT network with $N$ validators, the maximum number of tolerated Byzantine or failed nodes is $f = \lfloor(N - 1) / 3\rfloor$. For $N = 7$, $f = 2$.
  - The required consensus quorum threshold for proposing, preparing, and committing a block is $Q = 2f + 1 = 5$ signatures.
  - If Mumbai DC suffers complete destruction (3 nodes offline), 4 validators remain active (3 in DR + 1 Witness). Because $4 < 5$, the blockchain enters safe consensus pause, preventing split-brain blocks.
- **Automated Dynamic Validator Set Transition:**
  - Upon confirmed failover declaration, the Failover Controller executes an emergency validator transition runbook using pre-signed governance transactions or dynamic validator voting (`qbft_proposeValidatorVote`).
  - The active validator set is dynamically resized from $N = 7$ to $N = 4$ (removing the 3 offline Mumbai nodes).
  - With $N = 4$, the fault tolerance becomes $f = 1$ and the required quorum threshold adjusts to $Q = 2(1) + 1 = 3$ signatures. The remaining 4 nodes (3 DR + 1 Witness) immediately resume block production and settlement finality without data loss.
- **P2P Transport Resilience:** Validators communicate over dedicated Direct Connect private links with fallback to public internet WireGuard mesh tunnels, eliminating single-point-of-failure network dependencies.

## Step-by-Step Build Instructions
1. **Provision Interconnect & MACsec Hardware Encryption:** Configure dual redundant 10 Gbps AWS Direct Connect circuits between Mumbai (ap-south-1) and GIFT City / Hyderabad (ap-south-2). Enable IEEE 802.1AE MACsec encryption on physical transceivers. Verify round-trip latency is strictly under 12 ms and jitter is under 1 ms.
2. **Deploy Multi-Cluster Kubernetes Infrastructure:** Deploy identical Kubernetes 1.30 clusters in Mumbai and DR sites. Author Helm values and Terraform configurations establishing distinct cluster IDs (`growww-prod-mum-01` and `growww-prod-dr-01`) and CIDR allocations.
3. **Configure Cilium Cluster Mesh:** Install Cilium on both clusters using Helm. Establish cross-cluster mesh peering, enable eBPF WireGuard transparent node-to-node encryption, and configure CoreDNS to resolve cross-cluster services via `*.global` domain suffixes.
4. **Deploy Cross-Region Etcd Distributed Consensus:** Provision a 5-node Etcd cluster distributed across Mumbai (2 nodes), DR (2 nodes), and Witness (1 node) connected via TLS with mutual authentication (mTLS). Tune heartbeats (`heartbeat-interval: 250ms`, `election-timeout: 1250ms`) for cross-region stability.
5. **Deploy PostgreSQL Patroni Synchronous Cluster:** Deploy Patroni-managed PostgreSQL 16 across the two DCs. Configure `patroni.yaml` with DCS settings pointing to the cross-region Etcd cluster. Set `synchronous_commit = on` and `synchronous_standby_names = 'FIRST 1 (standby_dr, standby_witness)'` to guarantee zero-RPO data durability.
6. **Configure PgBouncer Connection Routing:** Deploy PgBouncer sidecars in both Kubernetes clusters. Configure transaction-level pooling and configure dynamic reload hooks so application write traffic is seamlessly repointed during database leader election.
7. **Deploy Apache Kafka MirrorMaker 2 Topology:** Deploy symmetrical 5-broker Kafka KRaft clusters in Mumbai and DR. Configure MirrorMaker 2 with bidirectional replication for non-compacted business topics, consumer offset translation, and regex-based cycle exclusion rules (`(?!.*-(mum|dr)).*`).
8. **Deploy Redis Active-Active In-Memory Cache Mesh:** Deploy Redis Cluster in Mumbai and DR. Configure active-active synchronization with CRDT conflict resolution for user sessions, nonces, and API rate-limiting token buckets.
9. **Deploy Distributed Hyperledger Besu QBFT Validator Quorum:** Provision 7 Besu validator nodes across Mumbai (3), DR (3), and Witness (1). Configure static bootnode discovery, TLS P2P encryption, and genesis block with QBFT block period set to 1000 ms. Verify all nodes achieve synchronized block height.
10. **Integrate Matching Engine Shadow Failover (Prompt 246):** Establish dedicated low-latency TCP/UDP kernel-bypass sockets over Direct Connect between Mumbai matching engine WAL writer and DR shadow matching engine. Verify WAL catch-up lag remains under 100 microseconds under peak trading load.
11. **Author Automated Failover Controller Daemon (`infra/failover-controller`):** Implement a resilient Go daemon running in both regions. The controller polls local and remote health probes (`/healthz/cross-dc-status`), evaluates consensus quorum, detects network partitions, and controls STONITH fencing agents.
12. **Configure Edge BGP Anycast & Route53 ARC Ingress:** Set up BIRD/FRR routing daemons advertising Anycast IP ranges to edge IXs. Configure AWS Route53 Application Recovery Controller (ARC) routing control states and health checks with automated 10-second DNS TTLs.
13. **Implement Split-Brain Fencing and STONITH Isolation:** Configure IPMI / BMC out-of-band management and AWS IAM instance fencing scripts. Ensure that before the DR site promotes itself to primary write authority, the Mumbai primary instance is unequivocally fenced to prevent dual-master data corruption.
14. **Build Automated SEBI BCP/DR Quarterly Drill Orchestration Harness:** Author automated end-to-end switchover scripts (`scripts/dr/execute-quarterly-drill.sh`). The script simulates total Mumbai DC failure, validates sub-second edge rerouting, asserts zero-RPO data integrity across PostgreSQL and Besu, logs cryptographic event proofs, and generates compliance audit documents.
15. **Deploy Multi-Region Observability and Alerting:** Deploy Thanos across Mumbai and DR Prometheus instances. Configure cross-DC Grafana dashboards tracking replication lag, QBFT round changes, Kafka mirror delay, and direct connect RTT, with automated PagerDuty escalation policies.

## Interfaces / Contracts

```json
// infra/failover-controller/spec/health-contract.json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "CrossDcHealthStatusResponse",
  "type": "object",
  "required": [
    "region_id",
    "datacenter_role",
    "timestamp_utc",
    "interconnect_rtt_ms",
    "postgres_replication_lag_bytes",
    "kafka_mirrormaker_lag_records",
    "besu_qbft_block_height",
    "besu_peer_count",
    "matching_engine_wal_seq",
    "health_verdict"
  ],
  "properties": {
    "region_id": {
      "type": "string",
      "enum": ["ap-south-1-mumbai", "ap-south-2-hyderabad", "gift-city-colo"]
    },
    "datacenter_role": {
      "type": "string",
      "enum": ["PRIMARY_ACTIVE", "STANDBY_WARM", "WITNESS"]
    },
    "timestamp_utc": {
      "type": "string",
      "format": "date-time"
    },
    "interconnect_rtt_ms": {
      "type": "number",
      "maximum": 25.0
    },
    "postgres_replication_lag_bytes": {
      "type": "integer",
      "minimum": 0
    },
    "kafka_mirrormaker_lag_records": {
      "type": "integer",
      "minimum": 0
    },
    "besu_qbft_block_height": {
      "type": "integer",
      "minimum": 0
    },
    "besu_peer_count": {
      "type": "integer",
      "minimum": 0
    },
    "matching_engine_wal_seq": {
      "type": "integer",
      "minimum": 0
    },
    "health_verdict": {
      "type": "string",
      "enum": ["HEALTHY", "DEGRADED", "PARTITIONED", "FAILED"]
    }
  }
}
```

```yaml
# deployments/patroni/patroni-multiregion.yaml
scope: growww-postgres-cluster
namespace: /service/growww-patroni
name: postgres-mum-01

etcd3:
  hosts:
    - 10.100.1.10:2379  # Mumbai Etcd 01
    - 10.100.1.11:2379  # Mumbai Etcd 02
    - 10.200.1.10:2379  # DR Etcd 01
    - 10.200.1.11:2379  # DR Etcd 02
    - 10.300.1.10:2379  # Witness Etcd 01
  cacert: /etc/patroni/certs/etcd-ca.crt
  cert: /etc/patroni/certs/etcd-client.crt
  key: /etc/patroni/certs/etcd-client.key

bootstrap:
  dcs:
    ttl: 30
    loop_wait: 10
    retry_timeout: 10
    maximum_lag_on_failover: 0
    synchronous_mode: true
    synchronous_mode_strict: true
    synchronous_node_count: 1
    postgresql:
      use_pg_rewind: true
      use_slots: true
      parameters:
        wal_level: replica
        max_wal_senders: 10
        max_replication_slots: 10
        hot_standby: "on"
        synchronous_commit: "on"
        synchronous_standby_names: "FIRST 1 (postgres-dr-01, postgres-wit-01)"
        archive_mode: "on"
        archive_command: "pgbackrest --stanza=growww archive-push %p"

postgresql:
  listen: 0.0.0.0:5432
  connect_address: 10.100.2.10:5432
  data_dir: /var/lib/postgresql/data
  bin_dir: /usr/lib/postgresql/16/bin
  pg_hba:
    - hostssl replication replicator 10.0.0.0/8 cert
    - hostssl all all 10.0.0.0/8 md5
```

```properties
# deployments/kafka/mirrormaker2-crossdc.properties
clusters = mum, dr

mum.bootstrap.servers = 10.100.3.10:9092,10.100.3.11:9092,10.100.3.12:9092
dr.bootstrap.servers = 10.200.3.10:9092,10.200.3.11:9092,10.200.3.12:9092

mum->dr.enabled = true
mum->dr.topics = orders.*,trades.*,settlement.*,risk.*
mum->dr.topics.exclude = .*-(mum|dr),.*internal.*
mum->dr.groups = .*
mum->dr.emit.heartbeats.enabled = true
mum->dr.emit.checkpoints.enabled = true
mum->dr.sync.group.offsets.enabled = true
mum->dr.sync.group.offsets.interval.seconds = 1
mum->dr.sync.topic.acls.enabled = false

dr->mum.enabled = true
dr->mum.topics = orders.*,trades.*,settlement.*,risk.*
dr->mum.topics.exclude = .*-(mum|dr),.*internal.*
dr->mum.groups = .*
dr->mum.emit.heartbeats.enabled = true
dr->mum.emit.checkpoints.enabled = true
dr->mum.sync.group.offsets.enabled = true
dr->mum.sync.group.offsets.interval.seconds = 1

replication.factor = 3
checkpoints.topic.replication.factor = 3
heartbeats.topic.replication.factor = 3
offset-syncs.topic.replication.factor = 3
offset.storage.replication.factor = 3
status.storage.replication.factor = 3
config.storage.replication.factor = 3
```

```json
// deployments/dns/arc-failover-policy.json
{
  "RoutingControlList": [
    {
      "ControlName": "MumbaiPrimaryIngressControl",
      "ControlArn": "arn:aws:route53-recovery-control::123456789012:control/mum-active-01",
      "Status": "Deployed",
      "HealthCheckConfig": {
        "Inverted": false,
        "FailureThreshold": 2,
        "RequestInterval": 10,
        "ResourcePath": "/healthz/cross-dc-status"
      }
    },
    {
      "ControlName": "DRSecondaryIngressControl",
      "ControlArn": "arn:aws:route53-recovery-control::123456789012:control/dr-standby-01",
      "Status": "Deployed",
      "HealthCheckConfig": {
        "Inverted": false,
        "FailureThreshold": 2,
        "RequestInterval": 10,
        "ResourcePath": "/healthz/cross-dc-status"
      }
    }
  ],
  "SafetyRules": [
    {
      "RuleType": "ASSERTION",
      "Name": "AtLeastOneRegionActiveRule",
      "AssertedControls": [
        "arn:aws:route53-recovery-control::123456789012:control/mum-active-01",
        "arn:aws:route53-recovery-control::123456789012:control/dr-standby-01"
      ],
      "RuleConfig": {
        "Threshold": 1,
        "Type": "ATLEAST",
        "Inverted": false
      }
    }
  ]
}
```

## Security & Compliance Notes
- **SEBI BCP & DR Regulatory Mandates:** The platform complies fully with SEBI Circulars SEBI/HO/MRD/DMS/CIR/P/2019/43 and SEBI/HO/MRD1/DT/CIR/P/2021/539. Critical trading, risk, clearing, and depository interfaces maintain zero RPO (zero data loss) and sub-second RTO.
- **Physical & Seismic Zone Separation:** In strict adherence to statutory rules, the Disaster Recovery site (GIFT City / Hyderabad) is separated by >500 km from the Primary DC (Mumbai). Mumbai is situated in Seismic Zone III/IV, whereas Hyderabad is located in stable Seismic Zone II (and GIFT City Gandhinagar in Zone III inland), safeguarding against concurrent catastrophic earthquake disruption.
- **Mandatory Quarterly DR Drills:** SEBI mandates live DR switchover drills conducted every quarter during active trading hours. The automated drill harness executes failover without scheduled market suspension, records cryptographic timing and ledger receipts, and compiles a signed audit dossier for submission to SEBI within 24 hours of drill completion.
- **Data In-Flight Encryption:** All cross-data-center traffic traversing dark fiber leased lines or Direct Connect circuits is secured via IEEE 802.1AE MACsec hardware-level link encryption. Inter-pod and cross-cluster communications enforce mutual TLS 1.3 with AES-256-GCM cipher suites and transparent WireGuard eBPF encryption.
- **Split-Brain Fencing & STONITH Protection:** To prevent split-brain scenarios where both Mumbai and DR concurrently accept financial write operations, the architecture implements STONITH (Shoot The Other Node In The Head) fencing. Primary instances must be confirmed powered down or isolated via out-of-band IPMI / AWS API fences before DR promotes write authority.

## Acceptance Criteria
- [ ] Direct Connect and dark fiber interconnect sustains round-trip latency under 12 ms with MACsec 802.1AE link-level encryption enabled.
- [ ] Cilium Cluster Mesh establishes seamless cross-cluster pod routing and global service discovery between Mumbai and DR clusters.
- [ ] PostgreSQL Patroni cluster enforces synchronous replication (`synchronous_commit = on`) with verified zero transaction loss (RPO = 0) upon simulated primary DC failure.
- [ ] Kafka MirrorMaker 2 maintains continuous active-active topic replication with cross-DC replication lag under 50 ms under 100,000 msg/sec load.
- [ ] Hyperledger Besu 7-node QBFT validator set maintains consensus across three failure domains, automatically transitioning validator set upon regional failure without ledger reorg or permanent halt.
- [ ] Matching engine warm shadow promotion in DR (Prompt 246) completes in under 50 ms without sequence skips or duplicated trades.
- [ ] BGP Anycast and AWS Route53 ARC routing failover redirects external client ingress traffic within 1000 ms of primary failure detection.
- [ ] Failover Controller daemon executes automated split-brain prevention and STONITH node isolation before promoting DR to write authority.
- [ ] Automated quarterly DR drill script successfully simulates Mumbai DC shutdown, switches traffic to DR, runs end-to-end trading transactions, and generates SEBI-compliant audit documentation.
- [ ] Thanos and Prometheus federated monitoring provides real-time alerting on cross-DC replication lag, consensus round drift, and network partition states.

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 101 (System Architecture Overview)
  - Prompt 104 (Event Schema & Kafka Topic Standards)
  - Prompt 205 (Order Matching Engine)
  - Prompt 246 (Matching Engine Memory-Mapped WAL & Hot-Warm Shadow Failover)
  - Prompt 301 (Hyperledger Besu Architecture & Genesis Configuration)
  - Prompt 401 (PostgreSQL Schema Architecture & Transaction Modeling)
  - Prompt 717 (AWS CloudHSM Cluster Architecture & Key Management)
  - Prompt 804 (GitOps Canary Rollout Pipeline with ArgoCD)
- **Downstream & Parallel Dependencies:**
  - Prompt 817 (Multi-Region Chaos Engineering & Network Partition Injection)
  - Prompt 908 (Production Launch Runbook & Operational Go-Live Checklist)
  - Prompt 910 (SEBI BCP/DR Compliance Audit & Drill Verification Runbook)
