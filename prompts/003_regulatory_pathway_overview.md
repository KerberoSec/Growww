# 003 - Regulatory Pathway & Sandbox Strategy (SEBI, RBI, IFSCA)

## Purpose
Operating a compliant blockchain-based fractional investment platform in India requires a coordinated multi-regulator engagement strategy across three primary financial authorities: the Securities and Exchange Board of India (SEBI), the Reserve Bank of India (RBI), and the International Financial Services Centres Authority (IFSCA). 

This document defines the master regulatory roadmap, regulatory sandbox entry criteria, compliance milestones, testing parameters, and formal licensing pathways for Growww. It ensures that every technical feature built by the engineering organization is mapped to a recognized regulatory exemption, sandbox cohort, or established intermediary license, eliminating legal ambiguity and regulatory risk.

## What You Are Building
A comprehensive regulatory roadmap and sandbox strategy specification (`docs/compliance/regulatory_pathway.md`) that details:
- The Three-Regulator Strategy Matrix (SEBI Innovation & Regulatory Sandbox, RBI Regulatory Sandbox on Cross-Border Payments, and IFSCA Fintech Regulatory Sandbox).
- Entity licensing roadmap: Stock Broker (SEBI), Depository Participant (NSDL/CDSL), Payment Aggregator / Nodal Escrow (RBI), and Capital Market Intermediary (IFSCA).
- Technical sandbox testing parameters: user caps, volume thresholds, maximum fractional ticket sizes, and emergency exit mechanisms.
- Automated regulatory reporting interfaces, data export formats, and audit checkpoints.
- The graduation roadmap from sandbox testing to full commercial operational authorization.

## Scope Boundaries
- **In Scope:**
 - Technical requirements and quantitative constraints imposed by SEBI, RBI, and IFSCA sandbox frameworks.
 - Architecture of regulatory observer nodes on the permissioned ledger.
 - Automated compliance metric collection, risk triggers, and regulatory reporting schemas.
 - Transition criteria from sandbox cohorts to permanent intermediary licenses.
- **Out of Scope / Handled Elsewhere:**
 - Legal drafting of formal petitions and legal memorandums (handled by General Counsel).
 - Microservice implementation for regulatory reporting exports (covered in Prompt 216).
 - Domestic and international KYC/AML policy definitions (covered in Prompts 004 and 005).

## Technology to Use
- Primary Format: Version-controlled Markdown specification (`docs/compliance/regulatory_pathway.md`) accompanied by YAML-based Regulatory Compliance Traceability Matrix (`docs/compliance/traceability_matrix.yaml`).
- Tooling: Automated CI validation tool checking that every API endpoint and smart contract function references a corresponding regulatory policy rule.
- Justification: Structuring regulatory requirements as a machine-readable traceability matrix enables automated compliance auditing during CI/CD builds, ensuring no unapproved financial flows can be merged into production.

## Backend / Infra Touchpoints
- Regulatory Reporting Service (Prompt 216).
- Immutable Audit Log Service (Prompt 218).
- SEBI / RBI / IFSCA Automated Data Export Gateways (SFTP / S3 Encrypted Buckets).
- Depository Participant APIs (NSDL/CDSL inspection hooks).

## Blockchain Interaction
Establishes the regulatory compliance architecture on the Hyperledger Besu ledger:
- **Dedicated Regulatory Observer Nodes:** SEBI and IFSCA are provisioned with read-only, non-validating observer nodes connected directly to the QBFT consortium network.
- **Real-Time Auditability:** Regulators can inspect real-time block generation, token issuance receipts, DvP settlement logs, and daily Merkle root reserve attestations in real time without human intermediary intervention.
- **Zero-PII Compliance Invariant:** Observer nodes receive cryptographic transaction records with pseudonymous account hashes (`bytes32`). Under judicial or regulatory subpoena, off-chain identity de-anonymization is executed through lawful warrant procedures via off-chain KMS.
- **Consensus & Security Invariant:** Anchored to Hyperledger Besu QBFT permissioned ledger with HSM key management and zero on-chain PII.

## Step-by-Step Build Instructions
1. Analyze the regulatory mandates and sandbox guidelines of SEBI (Circular SEBI/HO/MRD1/DSAP/CIR/P/2020/107), RBI (Regulatory Sandbox Master Directions), and IFSCA (Fintech Regulatory Sandbox Framework 2022).
2. Map the technical requirements of Growww to specific sandbox testing exemptions: fractional equity allocation, atomic DvP settlement on distributed ledgers, and automated profit-based fee deductions.
3. Define the quantitative boundaries for the Sandbox Phase 1: maximum 10,000 onboarded retail users, maximum INR 50,000 portfolio value per domestic user, and aggregate platform volume caps.
4. Establish the IFSCA GIFT City Sandbox parameters: maximum $10,000 USD portfolio per foreign retail investor and daily FX conversion volume limits.
5. Define the risk mitigation and consumer protection framework required by sandbox authorities (user consent, negative balance protection, guaranteed manual redemption fallback).
6. Design the architecture for Regulatory Observer Nodes on the Hyperledger Besu network, configuring read-only RPC permissions and automated telemetry feeds.
7. Detail the automated daily regulatory reporting feeds: trade reporting to SEBI clearing corporations, STR/CTR reporting to FIU-IND, and cross-border transaction summaries to RBI and IFSCA.
8. Specify the Emergency Exit and Rollback Protocol: detailed technical procedures for unwinding fractional tokens, liquidating custody shares, and returning 100% of fiat funds to investors in the event of sandbox termination.
9. Construct the Regulatory Compliance Traceability Matrix mapping every microservice endpoint and smart contract method to a regulatory rule ID.
10. Establish the audit and inspection criteria required for full license graduation.
11. Build Markdown documentation and compile YAML schemas.
12. Conduct joint review with External Regulatory Counsel, Chief Compliance Officer, and Chief Technology Officer to finalize the roadmap.

## Interfaces / Contracts
```yaml
# Regulatory Compliance Traceability Schema (docs/compliance/traceability_matrix.yaml)
regulatory_framework:
  version: "1.0.0"
  cohort: "FINTECH_SANDBOX_PHASE_1"
  authorities:
    sebi:
      sandbox_program: "SEBI_INNOVATION_SANDBOX"
      target_license: "STOCK_BROKER_DEPOSITORY_PARTICIPANT"
      constraints:
        max_domestic_users: 10000
        max_portfolio_inr_per_user: 50000.00
        max_aggregate_volume_inr: 500000000.00
        settlement_cycle: "T_PLUS_ZERO_DVP"
    rbi:
      sandbox_program: "RBI_INNOVATION_HUB"
      target_license: "PAYMENT_AGGREGATOR_ESCROW"
      constraints:
        settlement_rails: ["UPI", "IMPS", "NEFT", "RTGS"]
        unregulated_stablecoins_prohibited: true
    ifsca:
      sandbox_program: "IFSCA_FINTECH_SANDBOX"
      target_license: "CAPITAL_MARKET_INTERMEDIARY"
      constraints:
        max_foreign_users: 5000
        max_portfolio_usd_per_user: 10000.00
        sanctions_screening_mandatory: true
```

## Security & Compliance Notes
- All sandbox systems must adhere strictly to the SEBI Cyber Security and Cyber Resilience Framework (CSCRF), including mandatory annual VAPT (Vulnerability Assessment and Penetration Testing) by CERT-In empaneled auditors.
- Immediate suspension triggers: Any unbacked token issuance, custody reconciliation mismatch > 0 shares, or unauthorized PII ledger write triggers automated system halt and regulatory notification within 1 hour.

## Acceptance Criteria
- [ ] `docs/compliance/regulatory_pathway.md` is complete with detailed roadmap milestones for SEBI, RBI, and IFSCA.
- [ ] Quantitative sandbox constraints (user caps, volume limits, capital limits) are formally specified.
- [ ] Emergency Exit & Token Unwinding Protocol is documented and mathematically verified for complete investor capital return.
- [ ] Traceability matrix schema (`traceability_matrix.yaml`) is created and integrated into CI verification pipelines.
- [ ] Sign-off obtained from General Counsel and Lead Regulatory Architect.

## Suggested Order / Dependencies
- Prerequisites: 000 (Project North Star), 002 (Two-Entity Structure).
- Parallel Tasks: 004 (Domestic KYC/AML), 005 (Foreign KYC/AML), 216 (Regulatory Reporting).
- Downstream Blockers: Blocks Prompt 216, Prompt 218, and Category 7 security/compliance prompts.
