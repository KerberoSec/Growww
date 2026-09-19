# 710 - Business Continuity Plan (BCP) & Disaster Recovery Framework

## Purpose
Establishes an enterprise-grade Business Continuity Plan (BCP) and Disaster Recovery (DR) architecture for the Growww investment platform. Guarantees business continuity, data preservation, and rapid automated failover across geographically isolated cloud regions (Primary: AWS Mumbai `ap-south-1`, Secondary/DR: AWS Hyderabad `ap-south-2`). 

Enforces strict Recovery Time Objectives (RTO < 15 minutes for critical trading and settlement operations; RTO < 4 hours for non-critical administrative services) and Recovery Point Objectives (RPO = 0 for the permissioned Hyperledger Besu blockchain ledger; RPO < 1 minute for off-chain transactional databases), strictly adhering to SEBI CSCRF Chapter VI and RBI Business Continuity Guidelines.

## What You Are Building
- **Master BCP & Disaster Recovery Plan:** `docs/security/bcp_dr_master_plan.md` defining Crisis Management Team (CMT) hierarchies, disaster declaration criteria, communication trees, and recovery procedures.
- **Multi-Region Failover Orchestration Automation:** Executable runbooks and CLI tools (`scripts/dr/failover-orchestration/`) automating database promotion, DNS routing shifts, Kubernetes workload activation, and Besu validator cluster re-anchoring.
- **Continuous Cross-Region Data Replication Pipeline:** pgBackRest + PostgreSQL Physical Streaming Replication and Kafka MirrorMaker 2 configurations synchronizing state from Mumbai to Hyderabad in near real-time.
- **Geo-Distributed Blockchain Validator Topology:** Multi-region Besu QBFT consortium node topology distributing 7 validator nodes across Mumbai (3 nodes), Hyderabad (3 nodes), and GIFT City (1 node) to ensure uninterrupted consensus during a regional cloud failure.
- **SEBI-Mandated Bi-Annual DR Drill Suite:** Comprehensive test plans, chaos injection playbooks, and automated audit report generators for conducting mandatory bi-annual live-market disaster recovery simulations.

## Scope Boundaries
- **In Scope:**
 - Regional disaster scenarios: complete primary data center/cloud region loss, fiber cuts, catastrophic power/cooling failures, regional cyber incidents.
 - Asynchronous and synchronous data replication topologies across Mumbai and Hyderabad.
 - Automated DNS health-check failover (AWS Route 53 / Cloudflare) with automated TTL degradation.
 - Crisis Management Team (CMT) activation workflows and regulatory notification templates (SEBI, RBI, IFSCA).
 - Split-brain prevention, ledger reconciliation, and failback (re-synchronization from DR back to Primary).
- **Out of Scope / Handled Elsewhere:**
 - Cloud infrastructure provisioning and baseline Terraform modules (handled in Prompt 805).
 - Routine automated database backup cron jobs (handled in Prompt 406).
 - Security incident containment and ransomware isolation (handled in Prompt 706).

## Technology to Use
- **DNS & Traffic Management:** AWS Route 53 Application Recovery Controller (ARC) with latency-based and health-check failover routing.
- **Database Replication:** PostgreSQL 16 Physical Streaming Replication with Patroni for automated failover, combined with pgBackRest cross-region S3 repository synchronization.
- **Message Broker Replication:** Apache Kafka MirrorMaker 2.0 synchronizing critical transactional topics.
- **Kubernetes State Backup:** Velero backing up EKS cluster manifests, volumes, and secrets to encrypted multi-region S3 buckets.
- **Blockchain Consensus:** Hyperledger Besu with QBFT consensus spanning multi-region VPC peering.
- **Justification:** Combining cloud-native DNS failover with asynchronous physical database streaming replication and decentralized Byzantine Fault Tolerant blockchain consensus achieves $RPO=0$ on settled transactions and sub-15-minute $RTO$ without single points of failure.

## Backend / Infra Touchpoints
- **Primary Infrastructure:** AWS Mumbai (`ap-south-1`) VPC, EKS Cluster, RDS/PostgreSQL Primary, Kafka Cluster, CloudHSM.
- **Secondary Infrastructure:** AWS Hyderabad (`ap-south-2`) VPC (Warm Standby), EKS Standby Cluster, RDS Replica, Kafka MirrorMaker 2.
- **Edge Routing:** AWS Route 53, AWS CloudFront, AWS WAF.
- **External Dependencies:** NSDL/CDSL Depository Secondary Gateways, RBI Payment Aggregator Redundant Endpoints.

## Blockchain Interaction
Multi-region, fault-tolerant permissioned ledger topology and validator quorum resilience:
- **Consortium Validator Distribution:** 7 validator nodes configured under QBFT consensus ($N=7, F=2$, requiring a supermajority quorum of $2F+1 = 5$ active validators to produce blocks):
 - 3 Validator Nodes in AWS Mumbai (`ap-south-1`)
 - 3 Validator Nodes in AWS Hyderabad (`ap-south-2`)
 - 1 Validator Node in GIFT City / IFSCA On-Premise Data Center
- **Zero RPO Invariant:** If the entire Mumbai cloud region experiences catastrophic failure, the remaining 4 nodes in Hyderabad and GIFT City maintain network availability, while an emergency vote in `MultiSigGovernance.sol` adjusts the validator set to 4 nodes ($N=4, F=1$, quorum 3), continuing block production with zero loss of token mint, burn, or DvP settlement records ($RPO = 0$).
- **Relayer RPC Failover:** Application transaction relayers automatically re-route Web3 RPC traffic from Mumbai node endpoints to healthy Hyderabad Besu node endpoints within $< 5$ seconds.

## Step-by-Step Build Instructions
1. Author the Master BCP and Disaster Recovery Plan (`docs/security/bcp_dr_master_plan.md`), defining Business Impact Analysis (BIA) metrics, RTO/RPO tiering, and disaster classification (Minor disruption, Severe service degradation, Catastrophic regional disaster).
2. Establish the Crisis Management Team (CMT) structure: Incident Commander, CISO, Head of Infrastructure, Lead DB Administrator, Head of Compliance, and PR/Communications Lead.
3. Configure AWS cross-region VPC peering and transit gateways between Mumbai (`ap-south-1`) and Hyderabad (`ap-south-2`) with dedicated encrypted tunnels.
4. Deploy the warm-standby Kubernetes (EKS) cluster in Hyderabad with core ingress controllers, monitoring agents, and minimum baseline pod replicas configured via Terraform (Prompt 805).
5. Implement PostgreSQL 16 cross-region physical streaming replication with Patroni and pgBackRest, configuring replication lag monitoring to alert if replication lag exceeds 30 seconds.
6. Configure Kafka MirrorMaker 2.0 to replicate critical transactional topics (`orders.*`, `trades.*`, `wallet.*`, `settlement.*`) from Mumbai to Hyderabad with active consumer offset synchronization.
7. Deploy the 7-node geo-distributed Hyperledger Besu QBFT validator cluster across Mumbai (3), Hyderabad (3), and GIFT City (1).
8. Configure AWS Route 53 Application Recovery Controller (ARC) health checks monitoring `/healthz` endpoints across Mumbai and Hyderabad, configured with a 30-second failover threshold.
9. Implement the automated DR failover orchestration tool (`scripts/dr/failover-orchestration/execute-failover.py`):
 - Step A: Verify primary outage and confirm disaster declaration approval.
 - Step B: Promote Hyderabad PostgreSQL replica to primary write master.
 - Step C: Switch application database connection strings in Vault/Kubernetes secrets.
 - Step D: Scale up standby microservice deployments in Hyderabad EKS cluster.
 - Step E: Shift Route 53 DNS routing weights 100% to Hyderabad ingress.
 - Step F: Execute automated end-to-end synthetic health checks on order intake and settlement.
10. Implement the reverse failback script (`scripts/dr/failover-orchestration/execute-failback.py`) to re-establish replication from Hyderabad back to Mumbai once the primary region recovers.
11. Build the automated split-brain prevention and data reconciliation tool comparing PostgreSQL sequence IDs, Kafka high watermarks, and Besu blockchain block numbers.
12. Design the SEBI-mandated bi-annual live DR drill schedule: schedule unannounced mock failover tests during market off-hours, simulating regional blackout and verifying RTO/RPO targets.
13. Generate the formal SEBI DR Drill Report template, capturing timestamps, recovery metrics, replication lag at cutoff, data loss analysis, and executive sign-off.

## Interfaces / Contracts
```yaml
# configs/dr/dr_manifest.yaml
bcp_dr_specification_version: "1.0.0"
disaster_recovery_topology:
  primary_region: "ap-south-1"    # AWS Mumbai
  secondary_region: "ap-south-2"  # AWS Hyderabad
  gift_city_region: "in-gift-1"   # GIFT City IFSC On-Prem
  targets:
    critical_trading_rto_seconds: 900   # 15 minutes
    blockchain_ledger_rpo_seconds: 0    # Zero loss
    database_rpo_seconds: 60            # 1 minute
  replication_matrix:
 - datastore: "PostgreSQL Primary (Core OLTP)"
      mechanism: "Physical Streaming Replication + Patroni"
      sync_mode: "ASYNCHRONOUS"
      max_tolerated_lag_bytes: 10485760 # 10MB
 - datastore: "Apache Kafka"
      mechanism: "MirrorMaker 2.0"
      topics_included: ["orders.*", "trades.*", "wallet.*", "settlement.*", "compliance.*"]
 - datastore: "Blockchain Ledger (Besu)"
      mechanism: "Consortium QBFT Peer-to-Peer Consensus Mesh"
      validator_distribution:
        ap_south_1: 3
        ap_south_2: 3
        in_gift_1: 1
  dns_failover:
    provider: "AWS Route 53 ARC"
    health_check_interval_seconds: 10
    failure_threshold_consecutive_checks: 3
```

```json
// Sample SEBI DR Drill Performance Audit JSON
{
  "drill_id": "DR-2026-Q3-001",
  "drill_execution_date_ist": "2026-09-18T20:00:00+05:30",
  "drill_type": "LIVE_FAILOVER_SIMULATION",
  "simulated_scenario": "Total AWS Mumbai Region Loss",
  "primary_region": "ap-south-1",
  "dr_region": "ap-south-2",
  "metrics": {
    "disaster_declared_at": "20:02:15",
    "dns_switch_completed_at": "20:06:40",
    "db_promoted_at": "20:07:10",
    "first_synthetic_trade_executed_at": "20:11:30",
    "actual_rto_minutes": 9.25,
    "rto_target_met": true,
    "actual_db_rpo_seconds": 12,
    "db_rpo_target_met": true,
    "blockchain_blocks_missed": 0,
    "blockchain_rpo_target_met": true
  },
  "data_integrity_audit": {
    "total_settlement_discrepancies": 0,
    "reconciliation_hash_match": true
  },
  "ciso_signoff": "VERIFIED_COMPLIANT"
}
```

## Security & Compliance Notes
- **SEBI CSCRF Chapter VI (BCP & DR Policy):** Mandates that market intermediaries maintain a secondary DR site located in a different seismic zone, conduct bi-annual live DR drills with market participants, and submit comprehensive drill reports to SEBI within 15 days.
- **RBI Guidelines on BCP for Financial Entities:** Requires board-approved BCP/DR policies, periodic threat-based scenario testing, and maintenance of out-of-band communication channels during primary failure.
- **Data Sovereignty & Encryption:** Cross-region data replication traffic must remain encrypted in transit (TLS 1.3 over AWS private backbone) and comply with Indian data residency mandates.

## Acceptance Criteria
- [ ] Master BCP and Disaster Recovery Plan published in `docs/security/bcp_dr_master_plan.md`.
- [ ] Cross-region PostgreSQL streaming replication and Kafka MirrorMaker 2 operational between Mumbai and Hyderabad.
- [ ] 7-node Hyperledger Besu validator mesh operational across Mumbai, Hyderabad, and GIFT City with verified fault tolerance against 3-node Mumbai region failure.
- [ ] Automated failover orchestration script (`execute-failover.py`) successfully transitions traffic to Hyderabad in $< 15$ minutes in a staging simulation drill.
- [ ] RPO = 0 verified for blockchain token settlement ledger during simulated failover.
- [ ] Route 53 health-check automated failover triggers and updates DNS within 60 seconds of simulated endpoint failure.
- [ ] Bi-annual DR drill operational procedures and SEBI compliance audit reporting templates verified.

## Suggested Order / Dependencies
- **Prerequisites:** 010 (NFRs), 101 (System Architecture), 302 (Validator Node Topology), 314 (Chain DR), 401 (PostgreSQL), 403 (Kafka), 406 (Backup Automation), 802 (Kubernetes), 805 (Terraform).
- **Parallel Tasks:** 706 (Incident Response Runbooks), 808 (Alerting & On-Call).
- **Downstream Dependents:** 904 (Chaos Engineering & Resilience), 906 (UAT Regulatory Scenarios), 908 (Production Launch Runbook).
