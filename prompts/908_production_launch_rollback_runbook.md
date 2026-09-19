# 908 - Production Launch Runbook & Emergency Rollback Procedures

## Purpose
Establishes the exhaustive, step-by-step operational runbook and emergency rollback protocol for launching the Growww platform into production. Because financial market platforms cannot tolerate prolonged downtime, corrupted balances, or orphaned on-chain states during releases, this runbook provides a battle-tested timeline (T-48 hours to T+24 hours), automated rollback triggers, and precise technical procedures for blue-green cutovers across microservices, databases, and the Hyperledger Besu blockchain.

## What You Are Building
A production-grade launch and rollback operational toolkit (`ops/launch/` and `docs/runbooks/`):
- Master Launch Execution Checklist (`ops/launch/T_MINUS_LAUNCH_SCHEDULE.md`) with explicit role assignments, verification scripts, and sign-off gates.
- Automated Pre-Flight Verification Script (`ops/launch/preflight_check.py`) testing database connectivity, HSM signer availability, Kafka topic partitions, and Besu validator health.
- Emergency Rollback Automation Scripts (`ops/launch/rollback_cutover.sh`) capable of performing blue-green traffic diversion and database backward-compatible rollbacks in < 60 seconds.
- Post-Launch Smoke Verification Test Suite (`tests/smoke/production_smoke.py`) validating core user journeys immediately post-DNS cutover.

## Scope Boundaries
- **In Scope:**
 - Full launch timeline from T-48h through T+24h post-launch hypercare.
 - Production database migration safety checks (zero backward-incompatible DDL statements).
 - Smart contract production deployment, constructor initialization, and multi-sig ownership transfer.
 - DNS cutover and CDN/WAF routing activation (Cloudflare / AWS Route53).
 - Defined rollback decision matrix with quantitative thresholds (error rate > 0.5%, matching latency > 10ms, database lock contention).
- **Out of Scope / Handled Elsewhere:**
 - Pre-launch sandbox pilot trials (Prompt 907).
 - Day-2 continuous monitoring and SLO alerting (Prompt 909).
 - Disaster recovery multi-region restoration (Prompt 710).

## Technology to Use
- **Deployment & Progressive Delivery:** ArgoCD + Flagger for declarative GitOps blue-green and canary deployments on Kubernetes.
- **Database Schema Migrations:** Flyway / Liquibase with forward-compatible migration strategies (expand-and-contract pattern).
- **Blockchain Deployment & Verification:** Foundry (`forge script`) executing deterministic contract deployments via Ledger/AWS CloudHSM hardware signers.
- **Incident & Launch Management:** PagerDuty / OpsGenie + dedicated launch war room communication bridge.

*Justification:* ArgoCD GitOps ensures all infrastructure state is auditable in git, while Flagger blue-green deployments enable instant, automated traffic rollbacks without container cold-start delays.

## Backend / Infra Touchpoints
- Production AWS Multi-AZ / EKS Kubernetes Cluster.
- Production PostgreSQL 16+ Primary & Multi-AZ Standby with Patroni.
- Production Apache Kafka Cluster (6 brokers, RF=3).
- AWS CloudHSM / HashiCorp Vault Transit Engine for production transaction signing.
- Production Hyperledger Besu 4-Node Validator Consortium + 2 Backup RPC nodes.
- Cloudflare CDN, DDoS Shield, and WAF.

## Blockchain Interaction
- Executes the production smart contract deployment ceremony for `DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `ProofOfReserveRegistry.sol`, and `MultiSigGovernance.sol`.
- Transfers contract administrative ownership (`DEFAULT_ADMIN_ROLE`, `UPGRADER_ROLE`) from deployer keys to the 3-of-5 production MultiSig Governance contract held by designated corporate custodians.
- Validates that Besu validators are actively signing blocks with sub-2-second block intervals and gas limits set to 30,000,000.
- Publishes the Initial Genesis Proof-of-Reserve Attestation (attesting zero initial token circulation matching zero initial depository holding).

## Step-by-Step Build Instructions
1. Initialize the `ops/launch/` repository with `checklists/`, `scripts/`, `preflight/`, and `rollback/`.
2. Construct the T-Minus Master Launch Schedule (`ops/launch/T_MINUS_LAUNCH_SCHEDULE.md`) mapping every task from T-48h to T+24h with designated owners.
3. Author the Pre-Flight Verification Script (`ops/launch/preflight_check.py`) validating API health, Kafka topics, DB replication lag (< 100ms), and HSM connectivity.
4. Execute T-48h Gate: Freeze code repositories; tag release candidates (`v1.0.0-prod`); complete final staging security scan.
5. Execute T-24h Gate: Run database pre-migration dry runs on production read replicas; verify backup snapshots and point-in-time recovery (PITR).
6. Execute T-12h Gate: Conduct Key Ceremony; initialize production Hyperledger Besu genesis network; verify all 4 validators establish QBFT peer mesh.
7. Execute T-6h Gate: Deploy production smart contracts using Foundry and AWS CloudHSM signer; verify contract bytecodes on Blockscout; transfer admin roles to MultiSig contract.
8. Execute T-2h Gate: Deploy backend microservices to "Green" Kubernetes environment via ArgoCD; keep "Blue" environment idle.
9. Execute T-1h Gate: Run automated synthetic smoke tests against "Green" private endpoints; verify zero errors across all microservices.
10. Execute T-0 (Go/No-Go Decision): Convene Launch Committee (CTO, CCO, Lead Architect, SRE Lead); obtain unanimous verbal and digital sign-off.
11. Execute T+0 (Cutover): Switch Cloudflare DNS / ALB listener rules to direct 100% traffic to Green environment; enable public investor registration.
12. Execute T+15m Post-Cutover Verification: Run live end-to-end synthetic orders; verify order matching and Besu on-chain settlement receipts.
13. Execute T+1h to T+24h Hypercare: SRE and AppSec teams monitor real-time Prometheus dashboards, error budgets, and system logs in 24/7 war room.
14. In the event of an unresolvable critical anomaly (Rollback Trigger), execute the Rollback Protocol within 60 seconds: divert traffic back to Blue, pause Besu smart contracts via MultiSig, and notify stakeholders.

## Interfaces / Contracts
```markdown
# Rollback Decision Matrix & Triggers (ops/launch/ROLLBACK_MATRIX.md)

| Anomaly Condition | Threshold / Metric | Measurement Window | Action | Approver Required |
| :--- | :--- | :--- | :--- | :--- |
| **API Ingestion Error Spike** | 5xx HTTP Error Rate > 0.5% | 3 consecutive minutes | Automatic Blue Traffic Revert | Lead SRE (Automated) |
| **Matching Engine Degrade** | p99 Order Latency > 10ms | 2 consecutive minutes | Automatic Blue Traffic Revert | Lead SRE (Automated) |
| **Settlement Desync** | Besu Tx Failure Rate > 0.1% | 5 minutes | MultiSig Smart Contract Pause | Chief Compliance Officer |
| **Database Lock Escalation** | Active DB Deadlocks > 10/min | 3 minutes | Immediate DB Connection Drain | Lead Database Architect |
| **Security Breach / Exploit** | Any unauthorized state change | Instant | Immediate System Emergency Freeze | Incident Commander (Any 1 of 3) |
```

```python
# Launch Pre-Flight Checker Contract (ops/launch/preflight_check.py)
class PreflightChecker:
    def check_database_connectivity(self) -> bool: ...
    def check_kafka_topics_and_partitions(self) -> bool: ...
    def check_redis_cluster_health(self) -> bool: ...
    def check_hsm_kms_signers(self) -> bool: ...
    def check_besu_validator_quorum(self) -> bool: ...
    def check_external_custodian_mTLS(self) -> bool: ...
    def run_all(self) -> dict[str, bool]: ...
```

## Security & Compliance Notes
- All launch commands must be executed via auditable bastion hosts with full session keystroke recording.
- No single engineer possesses full production privileges; contract deployments and infrastructure cutovers require two-person verification (4-eyes principle).
- Rollback procedures must preserve transactional audit logs and database change logs to ensure complete post-incident forensic capability.

## Acceptance Criteria
- [ ] Pre-flight validation script verifies 100% of infrastructure prerequisites with zero failures.
- [ ] Smart contract production deployment completes and ownership transfers to MultiSig Governance contract.
- [ ] Blue-green traffic cutover executes with zero dropped TCP connections and zero lost orders.
- [ ] Rollback automation script is tested in pre-production and successfully reverts traffic within < 60 seconds.
- [ ] 24-hour post-launch hypercare period concludes with zero Critical (P0/P1) incidents.

## Suggested Order / Dependencies
- **Prerequisites:** 307 (MultiSig Governance), 804 (CD Pipeline), 808 (Alerting & Runbooks), 906, 907.
- **Parallel Tasks:** 909 (Post-Launch Monitoring & SLOs), 910 (Customer Support).
