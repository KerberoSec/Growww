# 701 - Full System Threat Model & STRIDE Analysis

## Purpose
Establishes a comprehensive, multi-tier threat model for the entire Growww investment infrastructure using the STRIDE methodology (Spoofing, Tampering, Repudiation, Information Disclosure, Denial of Service, Elevation of Privilege). The Growww platform bridges regulated Indian financial markets (SEBI custodians NSDL/CDSL, RBI banking rails) with a permissioned blockchain settlement layer (Hyperledger Besu) across multi-platform client applications (Flutter on Android, iOS, Windows, Linux, macOS) and web dashboards. 

This threat model identifies system attack vectors, evaluates residual risks against SEBI's Cybersecurity and Cyber Resilience Framework (CSCRF) and CERT-In guidelines, and defines precise technical countermeasures before engineering implementation begins.

## What You Are Building
- **Master Threat Model Specification:** `docs/security/threat_model.md` defining threat catalogs, attack trees, trust boundaries, and risk scoring (DREAD & CVSS v3.1).
- **Data Flow Diagrams (DFDs):** Level 0 Context, Level 1 Container, and Level 2 Microservice-to-Ledger data flow diagrams formatted in Mermaid.js with explicit trust boundary overlays.
- **STRIDE Threat Traceability Matrix:** Structured spreadsheet/markdown mapping each identified threat to affected architectural components, mitigation mechanisms, and corresponding Jira/Prompt implementation tickets.
- **Automated Threat Model Validator:** CI-integrated linting script (`scripts/security/validate-threat-model.py`) that checks for unmitigated High/Critical risks and enforces architectural compliance.

## Scope Boundaries
- **In Scope:**
 - Full system STRIDE analysis across all 9 architectural tiers (Flutter Clients, Next.js Web, API Gateway/BFF, Microservices, Event Streaming/Kafka, Relational/NoSQL Stores, HSM/KMS Custody, Permissioned Blockchain Network, and External Banking/Custody Adapters).
 - Trust boundary identification (Client-to-Gateway, Inter-Service mTLS, Fiat Rail Boundaries, Blockchain Validator Network, Domestic vs GIFT City Gateway).
 - Threat actor profiling: external cybercriminals, malicious insiders, state-sponsored entities, colluding validator nodes, and compromised third-party vendor systems.
 - Threat scoring using CVSS v3.1 metrics and DREAD qualitative severity ranking.
- **Out of Scope / Handled Elsewhere:**
 - Implementation of IAM and RBAC controls (handled in Prompt 702).
 - Real-time sanctions and AML alerting engine implementation (handled in Prompts 703 and 704).
 - Penetration testing execution and bug bounty platform launch (handled in Prompt 705).
 - Security incident runbook execution (handled in Prompt 706).

## Technology to Use
- **Modeling Framework:** Microsoft STRIDE taxonomy combined with OWASP Threat Dragon specification format and CVSS v3.1 scoring calculators.
- **Diagramming & Architecture:** Mermaid.js for embedded Git-versioned Data Flow Diagrams (DFD Levels 0-2) and attack trees.
- **Automation & Linting:** Python 3.12 script using `pydantic` and `pyyaml` to validate that all threat catalog entries link to verified mitigation prompts and test suites.
- **Justification:** Maintaining the threat model as version-controlled code (`docs/security/threat_model.md`) ensures continuous synchronization with architectural PRs, preventing the security drift typical of static corporate documentation.

## Backend / Infra Touchpoints
- **Microservices Layer:** User Service (201), KYC Service (202), Wallet Service (203), Order Matching Engine (205), Trade Settlement (208), Custodian Adapter (213), Payment Gateway (212).
- **Infrastructure:** Kubernetes Clusters (802), Envoy/Kong API Gateway (219), HashiCorp Vault (109), Kafka Broker Mesh (403), PostgreSQL 16 (401), Redis 7 (402).
- **External Interfaces:** NSDL/CDSL Depository APIs, RBI NEFT/RTGS/UPI gateways, UIDAI/CKYC verification endpoints.

## Blockchain Interaction
Threat vectors and mitigations specific to the permissioned Hyperledger Besu consortium network:
- **Validator Compromise / Sybil Collusion:** Threat of colluding validator nodes attempting to rewrite historical DvP settlements or mint unbacked tokens; mitigated by QBFT Byzantine Fault Tolerance ($3f+1$ validator quorum across distinct institutional nodes) and FIPS 140-2 Level 3 HSM hardware key protection.
- **Mempool Front-Running / MEV:** Malicious relayer or validator front-running investor orders; mitigated by private consortium peer-to-peer encryption, sequencer FIFO ordering, and batch settlement hashing.
- **Smart Contract Exploit Vectors:** Reentrancy, integer overflow, flash-liquidation attacks, or unauthorized privilege escalation on `DigitalSecurityToken.sol` and `SettlementDvP.sol`; mitigated by OpenZeppelin audited base contracts, reentrancy guards, formal verification (Certora), and multi-party timelock governance (`MultiSigGovernance.sol`).
- **Proof-of-Reserve Attestation Poisoning:** Forging depository reserve certificates; mitigated by dual-signed Merkle root attestations from SEBI-registered custodians verified on-chain via `ProofOfReserveRegistry.sol`.

## Step-by-Step Build Instructions
1. Initialize the security documentation directory tree under `docs/security/threat_model/` and configure Markdown linting.
2. Define the comprehensive system asset catalog, classifying assets by criticality: Tier 1 (Validator signing keys, Master KMS keys, Database master credentials, Depository signing certs), Tier 2 (Investor PII, PAN/Aadhaar data, Live order books, Bank account tokens), Tier 3 (Public market data, Static assets).
3. Identify all threat actors and adversary capabilities (External Attacker, Disgruntled Employee, Compromised Third-Party Vendor, Rogue Validator Partner).
4. Construct DFD Level 0 (Context Diagram) capturing external actors (Investors, Compliance Officers, NSDL/CDSL, Banks) interacting with the Growww trust perimeter.
5. Construct DFD Level 1 (Container Diagram) detailing boundaries between Flutter Client, Web App, API Gateway, Microservice Mesh, Vault/HSM, Databases, Kafka, and Besu Nodes.
6. Construct DFD Level 2 (Detailed Data Flows) for critical transaction paths: (a) User Onboarding & KYC, (b) Fiat Deposit via UPI, (c) Order Placement & Matching, (d) On-Chain DvP Settlement, and (e) Token Redemption to Physical Equity.
7. Conduct systematic STRIDE decomposition for every process, data store, data flow, and external entity across all 9 architectural tiers.
8. Model specific blockchain failure modes: 51% validator partition, invalid state transition injections, relayer private key leakage, and consensus desynchronization between domestic and GIFT City nodes.
9. Formulate attack trees using Mermaid syntax for high-impact misuse cases (e.g., "Unauthorized Minting of Security Tokens", "PII Exfiltration from Database", "Bypassing KYC Whitelist Gates").
10. Calculate CVSS v3.1 base, temporal, and environmental scores for every identified threat and assign DREAD severity ratings.
11. Build the Threat Mitigation Traceability Matrix, mapping every threat to explicit technical controls in Prompts 702 through 710 and core services.
12. Implement the Python validation script (`scripts/security/validate-threat-model.py`) to parse `threat_model.yaml` and verify zero unassigned mitigations in CI.
13. Conduct a threat modeling review session with the Chief Information Security Officer (CISO) and Lead System Architect; sign off on baseline.

## Interfaces / Contracts
```yaml
# Schema for docs/security/threat_model/threat_catalog.yaml
threat_catalog_schema_version: "1.0.0"
threat_entry:
  id: "THR-BESU-001"
  title: "Validator Private Key Compromise leading to Rogue Block Attestation"
  stride_category: "Elevation of Privilege / Tampering"
  target_component: "Blockchain Validator Node (Besu QBFT)"
  trust_boundary: "Consortium Validator Mesh <-> Public Cloud VPC"
  threat_actor: "Advanced Persistent Threat / Malicious Node Operator"
  cvss_v31_vector: "CVSS:3.1/AV:N/AC:H/PR:H/UI:N/S:C/C:H/I:H/A:H"
  cvss_base_score: 8.5
  dread_rating:
    damage_potential: 10
    reproducibility: 4
    exploitability: 4
    affected_users: 10
    discoverability: 3
    total_score: 6.2 # High
  description: "Attacker obtains the ECDSA private key of a consortium validator node, allowing them to sign malicious state transitions or disrupt QBFT consensus."
  mitigations:
 - control_ref: "PROMPT-707"
      description: "Validator signing keys generated and stored in FIPS 140-2 Level 3 CloudHSM with PKCS#11 provider."
 - control_ref: "PROMPT-302"
      description: "QBFT consensus requiring >= 5 of 7 validators to commit blocks; single key cannot forge transactions."
 - control_ref: "PROMPT-706"
      description: "Automated validator eviction runbook via on-chain governance vote upon anomalous block proposal detection."
  residual_risk: "LOW"
  status: "MITIGATED"
  verification_test_ref: "tests/security/test_validator_key_isolation.py"
```

## Security & Compliance Notes
- **SEBI CSCRF Compliance:** Aligns with Cybersecurity and Cyber Resilience Framework guidelines for stock exchanges and market intermediaries regarding periodic threat landscape reviews and attack surface minimization.
- **CERT-In Directions:** Establishes clear technical boundaries and asset categorization to enable statutory 6-hour incident reporting in the event of an exploited vulnerability.
- **DPDP Act 2023 & RBI Localization:** Classifies investor financial and biometric data within India geographic boundaries, mandating zero-trust encryption at all crossed trust perimeters.

## Acceptance Criteria
- [ ] Comprehensive `docs/security/threat_model.md` completed, covering all 9 architectural tiers and both domestic and GIFT City entities.
- [ ] DFD Levels 0, 1, and 2 rendered in valid Mermaid syntax with clear trust boundary indicators.
- [ ] STRIDE analysis documents at least 40 discrete threat vectors with CVSS v3.1 and DREAD scoring.
- [ ] Blockchain-specific attack trees (consensus failure, key leakage, contract reentrancy, oracle poisoning) fully detailed.
- [ ] Mitigation Traceability Matrix maps 100% of Critical and High threats to specific architectural controls and prompt tickets.
- [ ] CI validation script (`scripts/security/validate-threat-model.py`) executes successfully, ensuring no unmitigated risks.
- [ ] Formal review sign-off completed by CISO and Lead Blockchain Architect.

## Suggested Order / Dependencies
- **Prerequisites:** 101 (System Architecture Overview), 102 (Service Boundary Map), 105 (Auth Architecture), 110 (Inter-Entity Communication), 301 (Permissioned Blockchain Platform).
- **Parallel Tasks:** Can be developed concurrently with 702 (IAM & RBAC) and 707 (Data Encryption Standards).
- **Downstream Dependents:** Informs 703, 704, 705, 706, 708, 709, and 710.
