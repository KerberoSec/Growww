# 911 - Mainnet Dress Rehearsal, Dual-Region Failover Simulation, Shadow Production Traffic Replay, and Regulated Cutover Protocol

## Purpose
Establishes the definitive, production-grade operational framework for conducting full-scale mainnet dress rehearsals, dual-region disaster recovery failover simulations, shadow production traffic replay, and regulated multi-entity cutover protocols for the Growww platform. 

Operating a regulated fractional equity and digital security token infrastructure under SEBI and IFSCA mandates requires absolute certainty that trading, atomic DvP settlement, real-time risk checks, market data distribution, and on-chain blockchain consensus execute flawlessly under peak market conditions and catastrophic infrastructure failures. This specification defines an end-to-end dry run framework (T-72 hours to T+48 hours) that validates zero data loss (RPO = 0), sub-30-second cross-region recovery (RTO < 30s), zero regression under shadowed live traffic, and cryptographically auditable regulatory cutover sign-offs.

## What You Are Building
A production-grade rehearsal, simulation, and cutover automation suite located in `ops/rehearsal/`, `ops/failover/`, `ops/shadow/`, and `docs/cutover/`:
- **Mainnet Dress Rehearsal Automation Harness (`ops/rehearsal/`)**: Full-lifecycle orchestrator executing synthetic trading days, simulated market opens/halts/closes, corporate action batches, and multi-party cryptographic key ceremonies on production-identical infrastructure.
- **Dual-Region Failover Simulation Engine (`ops/failover/`)**: Automated disaster recovery orchestration executing seamless traffic diversion and state reconciliation between AWS Primary (`ap-south-1` Mumbai) and Disaster Recovery (`ap-south-2` Hyderabad / GIFT City) across Kubernetes, PostgreSQL, Kafka MirrorMaker 2, Redis, and Hyperledger Besu.
- **Shadow Production Traffic Replay & Diff Pipeline (`ops/shadow/`)**: Non-intrusive traffic mirroring via Envoy API Gateway and FIX Gateway, anonymizing sensitive PII, replaying tick and order traffic at 1x to 5x production speeds, and performing automated diff analysis on response payloads, latencies, and state mutations.
- **Regulated Cutover Protocol & Digital Sign-off Vault (`docs/cutover/`, `ops/cutover/`)**: Multi-party governance workflow enforcing 4-eyes digital signatures from the Chief Technology Officer (CTO), Chief Compliance Officer (CCO), Lead Database Architect, and External Custodian Trustees prior to live production activation.

## Scope Boundaries
- **In Scope:**
  - T-72h to T+48h end-to-end rehearsal execution timeline with deterministic operational gates.
  - Active-Standby / Active-Active dual-region failover testing between Mumbai (`ap-south-1`) and Hyderabad/GIFT City (`ap-south-2`).
  - Automated PostgreSQL Patroni primary switchover, Kafka MirrorMaker 2 partition alignment, and Redis cluster failover under active transaction load.
  - Hyperledger Besu multi-region QBFT consensus partition simulation, validator isolation, and snap-sync catch-up verification.
  - Envoy API Gateway shadow traffic replication, request deduplication, payload anonymization, and Diffy comparison analysis.
  - Multi-party HSM key ceremony rehearsal and smart contract ownership transfer to `MultiSigGovernance.sol`.
  - Automated Go/No-Go decision matrix and 60-second emergency cutover abort procedures.
- **Out of Scope / Handled Elsewhere:**
  - Initial DevSecOps CI/CD scanning and SAST/DAST pipelines (Prompt 905).
  - Regulatory sandbox pilot participant onboarding (Prompt 907).
  - Standard single-AZ chaos pod termination experiments (Prompt 904).
  - Continuous Day-2 SLO alerting and error budget monitoring (Prompt 909).

## Technology to Use
- **Traffic Mirroring & Diff Analysis:** Envoy Proxy (`request_mirror_policies`), GoReplay / Shadow Replay Agent, Diffy (Open-source REST/gRPC diff engine).
- **Dual-Region Traffic Routing:** AWS Route 53 Application Recovery Controller (ARC), Cloudflare Dynamic Traffic Steering, AWS Global Accelerator.
- **Data Streaming & Database Replication:** PostgreSQL 16+ with Patroni and physical streaming replication, Apache Kafka 3.7+ with MirrorMaker 2.0 (active-passive topic replication with offset mapping), Redis 7+ Sentinel / Global Datastore.
- **Blockchain Infrastructure:** Hyperledger Besu 24.x QBFT consortium (4 validator nodes across Mumbai and Hyderabad/GIFT City), Foundry (`forge script` / `cast`), AWS CloudHSM / HashiCorp Vault Transit Engine.
- **Rehearsal Orchestration & Synthetic Load:** Python 3.12 (`asyncio`, `httpx`, `pydantic`), k6 distributed load testing clusters, Prometheus 2.50+ / Thanos, Grafana 10+, OpenTelemetry Collector.

*Justification:* Envoy native request mirroring combined with GoReplay allows real-time traffic duplication without injecting client latency, while Route 53 ARC and Patroni provide mathematically sound RPO = 0 failover guarantees mandated by SEBI.

## Backend / Infra Touchpoints
- **Primary Cloud Region:** AWS `ap-south-1` (Mumbai) hosting 3-AZ Kubernetes EKS Cluster, Primary PostgreSQL Patroni Cluster, Kafka Cluster (6 brokers), Redis Cluster (6 nodes), and 2 Hyperledger Besu Validators.
- **Disaster Recovery Region:** AWS `ap-south-2` (Hyderabad / GIFT City) hosting replica Kubernetes EKS Cluster, Read/Standby PostgreSQL Cluster, Standby Kafka Cluster (3 brokers), and 2 Hyperledger Besu Validators.
- **Edge Routing & Ingress:** Cloudflare Enterprise WAF, DDoS Shield, Magic Transit, and Route 53 Health Checks.
- **Key Management:** AWS CloudHSM / Dedicated Thales Luna HSM clusters with cross-region key replication for institutional signing.
- **Observability Stack:** Thanos multi-region metric aggregator, Grafana Tempo distributed tracing, and Loki centralized audit logs.

## Blockchain Interaction
- **Genesis & Validator Mesh Dry Run:** Validates that the 4-node Hyperledger Besu QBFT network initializes deterministically across regions (2 nodes in Mumbai, 1 node in GIFT City, 1 node in Hyderabad) maintaining 2-second block intervals with sub-30ms inter-region latency.
- **Cross-Region Consensus Resilience:** Simulates complete WAN disconnection between Mumbai and Hyderabad; validates that QBFT consensus ($N=4$, $F=1$, quorum requirement $2F+1=3$) continues uninterrupted when 1 node drops, and halts safely without forks if 2 nodes drop.
- **Fast-Sync Recovery Rehearsal:** Validates that when a partitioned validator re-joins the consortium, it catches up 5,000 blocks via snap sync in < 45 seconds without degrading active transaction throughput.
- **Smart Contract Deployment Ceremony:** Deploys `DigitalSecurityToken.sol`, `SettlementDvP.sol`, `ComplianceRegistry.sol`, `ProofOfReserveRegistry.sol`, and `MultiSigGovernance.sol` via HSM-signed transactions, executes role assignment, transfers `DEFAULT_ADMIN_ROLE` to the 3-of-5 MultiSig contract, and initiates the 48-hour timelock.
- **Genesis Proof-of-Reserve Attestation:** Executes synthetic attestation transaction publishing root hash of zero circulating tokens matching custodian depository zero-balance certificates.

## Step-by-Step Build Instructions
1. Scaffold directory tree `ops/rehearsal/`, `ops/failover/`, `ops/shadow/`, `ops/cutover/`, and `docs/cutover/`.
2. Author the T-Minus 72-Hour Master Rehearsal Schedule (`ops/rehearsal/TIMELINE_T_MINUS_72H.md`) mapping every phase: environment provisioning, shadow traffic replay, dual-region chaos failover, cutover ceremony, and hypercare.
3. Configure the Envoy API Gateway shadow filter (`ops/shadow/envoy_shadow_filter.yaml`) routing 100% of read traffic and sanitizing/mirroring 10% of write traffic to staging shadow endpoints.
4. Implement the Shadow Payload Sanitizer and Replay Daemon (`ops/shadow/shadow_replay_daemon.py`) replacing live Aadhaar/PAN, bank account numbers, and API tokens with deterministic mock test credentials before sending requests to the shadow backend.
5. Deploy Diffy regression detection service (`ops/shadow/diffy_config.yaml`) comparing response headers, status codes, and JSON bodies between primary and shadow clusters, flagging schema mismatches or latency drift > 5%.
6. Configure PostgreSQL Patroni cross-region standby cluster in `ap-south-2` and configure Kafka MirrorMaker 2.0 pipelines replicating transactional topics with consumer group offset synchronization.
7. Build the Automated Dual-Region Failover Orchestrator (`ops/failover/dual_region_orchestrator.py`) capable of:
   - Detecting primary region degradation via synthetic health probes.
   - Promoting PostgreSQL standby in `ap-south-2` to writable primary.
   - Switching Kafka consumer group offsets via MirrorMaker 2 checkpoints.
   - Updating Route 53 ARC routing controls to divert 100% client traffic to `ap-south-2`.
   - Asserting trade matching engine state continuity with zero lost executions (RPO = 0).
8. Implement the Blockchain Quorum Disaster Simulator (`ops/failover/besu_quorum_simulator.py`) injecting network packet loss, validator halts, and validating rapid block synchronization.
9. Construct the Automated Rehearsal Verification Suite (`ops/rehearsal/rehearsal_verifier.py`) running k6 load scripts (10,000 orders/sec) while executing failover switchovers, asserting that p99 latency remains < 15ms and zero double-spend or state corruption occurs.
10. Develop the Multi-Party Digital Sign-off Ceremony Tool (`ops/cutover/signoff_ceremony.py`) requiring Ed25519/RSA-4096 cryptographic signatures from CTO, CCO, Lead Architect, and Custodian Trustee.
11. Construct the Emergency Cutover Abort & Revert Script (`ops/cutover/cutover_abort.sh`) executing instant rollback of DNS, database connections, and smart contract pause states in < 60 seconds.
12. Conduct a full dry-run execution on pre-production infrastructure, logging every event, artifact checksum, and metric trace to the immutable audit bucket.
13. Publish the Regulated Cutover Protocol Runbook (`docs/cutover/REGULATED_CUTOVER_PROTOCOL.md`) including compliance escalation contacts, SEBI/IFSCA notification templates, and post-mortem review procedures.

## Interfaces / Contracts

### 1. Dual-Region Failover Decision Matrix (`ops/failover/FAILOVER_DECISION_MATRIX.md`)
```markdown
# Dual-Region Failover Decision Matrix & RTO/RPO Targets

| Incident Scenario | Detection Mechanism | Quantitative Trigger | Target RTO | Target RPO | Automated Action | Authorized Approver |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Primary Region Complete Outage** | Route 53 ARC Probes / Multi-AZ Ping | 3 consecutive failed health probes (30s) | < 30 seconds | 0 (Zero Loss) | Auto-promote DR Patroni DB, Shift Route 53 DNS | Incident Commander (Auto) |
| **PostgreSQL Primary Corruption / Split** | Patroni Raft Heartbeat & WAL Check | WAL replication stream halt > 5s | < 20 seconds | 0 (Zero Loss) | Promote Synchronous Standby node | Lead DB Architect (Auto) |
| **Kafka Broker Cluster Partition** | Kafka Lag Exporter / Under-Replicated | Under-replicated partitions > 0 for 2m | < 15 seconds | 0 (Zero Loss) | Failover consumer groups to MirrorMaker DR | Core Platform SRE Lead |
| **Besu Validator Quorum Loss (2 Nodes)** | Prometheus `besu_peers` & Block Rate | Zero blocks produced for > 6 seconds | < 45 seconds | 0 (Zero Loss) | Activate DR Standby Validator, Sync Snap | Blockchain Architect |
| **Settlement DvP State Desynchronization** | Off-chain Reconciliation Engine | Discrepancy between DB and Ledger > 0 | < 60 seconds | 0 (Zero Loss) | MultiSig contract pause, halt matching | Chief Compliance Officer |
```

### 2. Envoy Shadow Traffic Mirroring Configuration (`ops/shadow/envoy_shadow_filter.yaml`)
```yaml
# Envoy RouteConfiguration snippet for Shadow Traffic Replay
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: order-gateway-shadow-routing
  namespace: growww-core
spec:
  hosts:
    - "order-api.growww.internal"
  http:
    - name: "order-placement-mirror"
      match:
        - uri:
            prefix: /v1/orders
      route:
        - destination:
            host: order-service-primary.growww-core.svc.cluster.local
            port:
              number: 8080
          weight: 100
      mirror:
        host: order-service-shadow.growww-shadow.svc.cluster.local
        port:
          number: 8080
      mirrorPercentage:
        value: 100.0
      headers:
        request:
          set:
            x-growww-shadow-mode: "true"
            x-growww-replay-id: "%REQ(X-REQUEST-ID)%"
```

### 3. Rehearsal & Cutover Sign-off Interface (`ops/cutover/signoff_ceremony.py`)
```python
"""
Multi-Party Cryptographic Cutover Sign-Off Ceremony
Module: ops/cutover/signoff_ceremony.py
"""

from dataclasses import dataclass
from datetime import datetime, timezone
from enum import Enum
import hashlib
import json
from typing import Dict, List, Optional


class SignerRole(str, Enum):
    CHIEF_TECHNOLOGY_OFFICER = "CTO"
    CHIEF_COMPLIANCE_OFFICER = "CCO"
    LEAD_SYSTEM_ARCHITECT = "LEAD_ARCHITECT"
    LEAD_DATABASE_ARCHITECT = "LEAD_DBA"
    CUSTODIAN_TRUSTEE = "CUSTODIAN_TRUSTEE"


class GateStatus(str, Enum):
    PENDING = "PENDING"
    VERIFIED = "VERIFIED"
    REJECTED = "REJECTED"


@dataclass(frozen=True)
class PreflightCheckResult:
    check_name: str
    passed: bool
    latency_ms: float
    details: str
    timestamp: str


@dataclass(frozen=True)
class DigitalSignatureRecord:
    signer_name: str
    signer_role: SignerRole
    public_key_fingerprint: str
    signature_hex: str
    signed_payload_hash: str
    timestamp_utc: str


class MainnetCutoverCeremony:
    """
    Coordinates and validates formal multi-party sign-offs
    prior to executing production DNS cutover and token deployment.
    """

    def __init__(self, release_version: str, target_environment: str):
        self.release_version = release_version
        self.target_environment = target_environment
        self.preflight_results: List[PreflightCheckResult] = []
        self.signatures: Dict[SignerRole, DigitalSignatureRecord] = {}

    def record_preflight(self, check: PreflightCheckResult) -> None:
        """Records an automated preflight check result."""
        ...

    def verify_all_preflights_passed(self) -> bool:
        """Asserts that 100% of preflight gates are strictly green."""
        ...

    def generate_cutover_payload_manifest(self) -> str:
        """
        Generates canonical JSON manifest containing release hash,
        smart contract bytecodes, database migration version, and preflight outputs.
        """
        ...

    def register_signature(
        self,
        role: SignerRole,
        signer_name: str,
        public_key_fingerprint: str,
        signature_hex: str,
    ) -> bool:
        """
        Verifies cryptographic signature against canonical payload manifest
        using the signer's registered public key.
        """
        ...

    def is_cutover_authorized(self) -> bool:
        """
        Validates that all 5 mandatory roles have submitted valid signatures
        and all preflight verification gates are passed.
        """
        ...

    def export_audit_certificate(self, output_path: str) -> None:
        """Exports immutable cryptographic cutover certificate for SEBI/IFSCA audit compliance."""
        ...
```

### 4. Rehearsal Execution Timeline Specification (`ops/rehearsal/TIMELINE_T_MINUS_72H.md`)
```markdown
# Rehearsal & Cutover Operational Timeline

| Timeline Phase | Elapsed Time | Action Items | Success Verification Gate | Owner |
| :--- | :--- | :--- | :--- | :--- |
| **T-72h: Environment Prep** | 00:00 - 04:00 | Stand up isolated pre-prod infrastructure in `ap-south-1` and `ap-south-2`. Deploy latest release tags. | 100% microservice health checks return HTTP 200. | SRE Lead |
| **T-60h: Data Baseline** | 04:00 - 08:00 | Restore anonymized production backup snapshot. Run database forward migrations (Flyway). | Migration checksums match CI build artifacts. | Lead DBA |
| **T-48h: Shadow Traffic Start** | 08:00 - 20:00 | Enable Envoy traffic shadowing (100% read, 10% sanitized write). Run Diffy analyzer. | Diffy reports < 0.00% (Zero Fee) mismatch; p99 latency delta < 2ms. | Core Backend Lead |
| **T-24h: Dual-Region Failover** | 20:00 - 28:00 | Execute automated disaster simulation. Cut Mumbai region; promote Hyderabad/GIFT City. | RTO < 30s; RPO = 0; zero duplicate trades recorded. | Incident Commander |
| **T-12h: Blockchain Key Ceremony**| 28:00 - 32:00 | Multi-party HSM signing ceremony. Deploy smart contracts. Transfer admin roles to MultiSig. | Contracts verified on Blockscout; 3-of-5 signers confirmed. | Blockchain Architect |
| **T-4h: Full Load Simulation** | 32:00 - 36:00 | k6 distributed test suite replaying 50,000 active users, 10,000 orders/sec under market volatility. | Zero 5xx errors; order matching latency p99 < 1ms. | QA Lead |
| **T-1h: Cutover Sign-off Gate** | 36:00 - 37:00 | Convene Cutover Committee. Verify preflight results. Execute cryptographic sign-offs. | 5 of 5 signatures registered and verified in vault. | CCO & CTO |
| **T-0: Live Production Cutover** | 37:00 - 37:05 | Flip Cloudflare / Route 53 ARC to live production ingress. Enable client traffic. | 100% traffic routed; zero dropped connections. | Release Engineer |
| **T+1h to T+48h: Hypercare** | 37:05 - 85:05 | 24/7 War room monitoring. Real-time reconciliation, error budget tracking, and SEBI filing. | Zero P0/P1 incidents; daily ledger reconciliation balances. | SRE & Ops Teams |
```

## Security & Compliance Notes
- **SEBI Cyber Security and Cyber Resilience Framework Compliance:** Complies with mandatory disaster recovery (DR) testing requirements requiring simulated failover with RTO < 30 minutes and RPO = 0 for all transactional and order book states.
- **PII Anonymization & Data Sanitization:** Shadow traffic pipelines must strip or irreversibly hash all Personally Identifiable Information (PAN, Aadhaar, phone numbers, email addresses, and bank accounts) before replaying payloads into staging or shadow environments.
- **Four-Eyes Principle & Non-Repudiation:** Cutover execution requires cryptographic authorization from independent executive roles. Digital signatures must be recorded in an immutable, append-only audit log signed by AWS CloudHSM.
- **IFSCA Regulatory Reporting:** In accordance with GIFT City market infrastructure rules, all rehearsal execution logs, failover latency proofs, and DvP settlement verification hashes must be archived for a minimum of 8 years.

## Acceptance Criteria
- [ ] Mainnet dress rehearsal automation harness executes complete 72-hour simulated timeline on production-identical infrastructure without manual intervention.
- [ ] Dual-region failover engine successfully transfers active workload from Mumbai (`ap-south-1`) to Hyderabad/GIFT City (`ap-south-2`) in under 30 seconds with zero state loss (RPO = 0).
- [ ] PostgreSQL Patroni standby promotion and Kafka MirrorMaker 2 topic alignment execute automatically under 5,000 active orders/sec load.
- [ ] Shadow traffic replay daemon processes 100% of mirrored production read queries and 10% of sanitized write transactions with Diffy reporting zero breaking mismatches.
- [ ] Hyperledger Besu 4-node QBFT network withstands validator isolation, maintaining continuous consensus with remaining quorum and syncing partitioned nodes in < 45 seconds upon re-connection.
- [ ] Multi-party cryptographic sign-off ceremony requires and verifies valid signatures from all 5 designated roles (CTO, CCO, Lead Architect, Lead DBA, Custodian Trustee).
- [ ] Emergency cutover abort script is verified to revert DNS, drain database connections, and pause smart contracts within 60 seconds.
- [ ] Comprehensive documentation and runbooks (`ops/rehearsal/`, `ops/failover/`, `ops/shadow/`, `docs/cutover/`) are generated and verified against SEBI and IFSCA compliance standards.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 010 (NFRs Master Doc), Prompt 108 (Environment Strategy), Prompt 205 (Order Matching Engine), Prompt 307 (MultiSig Governance), Prompt 710 (Disaster Recovery Architecture), Prompt 802 (Kubernetes Architecture), Prompt 805 (Terraform Cloud Infra), Prompt 904 (Chaos Engineering), Prompt 908 (Production Launch Runbook).
- **Parallel Tasks:** Prompt 909 (Post-Launch Monitoring & SLOs), Prompt 910 (Customer Support & Grievance Redressal).
- **Downstream Operations:** Live Production Launch Execution and Ongoing Regulatory Compliance Filings.
