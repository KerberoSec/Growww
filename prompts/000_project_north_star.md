# 000 - Project North Star: Purpose, Boundaries & Architectural Vision

## Purpose
Growww is a blockchain-based, SEBI/RBI-compliant investment infrastructure designed to democratize access to Indian equities for domestic and international investors. By enabling fractional, 1:1 real-asset-backed digital representations of Indian equities held in regulated custody, Growww solves the structural barriers of high per-share capital requirements and friction-heavy cross-border investment corridors. 

This document establishes the non-negotiable architectural boundaries, governance principles, and regulatory postures for the entire engineering organization. Growww is strictly NOT an unregulated cryptocurrency exchange, nor does it issue synthetic, derivative, or unbacked tokens disconnected from physical depository shares. Every engineer, architect, and contributor must align all design decisions with the invariants codified in this document.

## What You Are Building
A comprehensive foundational architectural mandate and vision document (`docs/architecture/000_north_star.md`) that serves as the root technical constitution for the Growww project. The document must define:
- The 10 Non-Negotiable Architectural & Regulatory Invariants.
- Clear distinction between Growww's asset-backed tokenization model and speculative crypto/synthetic asset models.
- The two-tier legal-technical entity topology (Domestic Regulated Broker/Custodian Entity + GIFT City IFSCA Gateway).
- Complete lifecycle of a fractional unit (from fiat deposit, depository acquisition, token issuance, order matching, DvP settlement, to redemption/burn and fiat payout).
- The Universal Zero-Fee Model (0.00% fee - No fee at all) on trade notional turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol) (zero holding fees, zero AUM-based fees), alongside FIFO capital gains computation strictly for user tax compliance (Section 111A/112A).
- Formal anti-goals preventing scope creep into unregulated or speculative domains.

## Scope Boundaries
- **In Scope:**
 - Codification of system-wide architectural invariants, regulatory principles, and asset backing rules.
 - End-to-end asset lifecycle state transitions across traditional finance rails and permissioned ledger.
 - High-level topology across mobile/desktop Flutter clients, web console, polyglot microservices, and Hyperledger Besu consortium nodes.
 - Architectural definitions of anti-goals and boundary enforcement.
- **Out of Scope / Handled Elsewhere:**
 - Granular API and schema specifications (covered in Category 1 and Category 2).
 - Smart contract Solidity implementation code (covered in Category 3).
 - Specific cloud infrastructure deployment scripts and Terraform manifests (covered in Category 8).
 - Formal regulatory legal filings and petitions (covered in Prompt 003).

## Technology to Use
- Primary Format: Version-controlled Markdown (`docs/architecture/000_north_star.md`) embedded with Mermaid.js architecture and sequence diagrams.
- Documentation Framework: MkDocs Material with Architecture Decision Records (ADR) tooling and Git pre-commit markdown linting (`markdownlint-cli2`).
- Justification: Treating the architectural constitution as version-controlled code ensures that any proposed change to system invariants undergoes rigorous peer review, cryptographic signing, and automated CI validation before adoption.

## Backend / Infra Touchpoints
- Git Repository documentation root (`/docs/architecture/`).
- CI/CD Documentation Pipeline (automated schema validation and ADR changelog generation).
- Internal Developer Portal (searchable index for engineering, compliance, and product teams).

## Blockchain Interaction
Establishes the supreme ledger invariant: The blockchain is strictly a permissioned audit, issuance, and Delivery-versus-Payment (DvP) settlement record layer. It never stores unbacked tokens, private investor PII, or speculative volatile native gas assets.

- **Target Network:** Hyperledger Besu permissioned enterprise consortium ledger.
- **Consensus Mechanism:** QBFT (Quorum Byzantine Fault Tolerance) with a 2-second deterministic block time and immediate finality (zero probabilistic forks).
- **1:1 Custody Backing:** Every digital security token issued on-chain maps strictly to a real underlying security held in a designated SEBI-registered custodian/depository account (NSDL/CDSL). Unbacked minting is cryptographically prevented via multi-party HSM authorization.
- **Zero PII Ledger Invariant:** No investor names, Aadhaar numbers, PANs, email addresses, or phone numbers ever enter ledger transaction payloads, logs, or state storage. Only pseudonymous, salted cryptographic account hashes (`bytes32`) exist on-chain.

## Step-by-Step Build Instructions
1. Initialize the root architectural documentation repository structure under `docs/architecture/` with ADR tooling.
2. Draft Section 1: "The Core Problem & Mission Statement", detailing the democratization of Indian equities and cross-border accessibility.
3. Codify Section 2: "The 10 Non-Negotiable System Invariants", covering 1:1 custody backing, RBI banking rails, permissioned ledger, compliance-by-design, multi-party authorization, fixed 0.00% transaction fee (No fee at all) model with 0.00% fee launch policy revenue split, two-entity separation, 5-platform Flutter client, polyglot backend, and HSM security.
4. Draft Section 3: "Growww vs. Crypto Exchanges & Synthetics", creating an unambiguous comparison table proving regulatory compliance.
5. Model Section 4: "Two-Entity Governance & Operational Separation", diagramming the domestic SEBI/RBI entity and the GIFT City IFSCA gateway.
6. Detail Section 5: "Fractional Asset Lifecycle & DvP Settlement Flow", mapping states from fiat on-ramp to depository custody, token minting, order execution, DvP settlement, and redemption/burn.
7. Detail Section 6: "Monetization Architecture", formalizing the mathematical rule for the Universal Zero-Fee Model (0.00% fee - No fee at all) on trade notional turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol), clarifying zero holding/AUM fees and that the realized gain engine computes FIFO capital gains strictly for user tax compliance (Section 111A/112A).
8. Draft Section 7: "Platform & Client Architecture Overview", defining the shared Flutter native client (Android, iOS, Windows, macOS, Linux) and Next.js web portal.
9. Draft Section 8: "Enterprise Security & Key Custody Mandate", specifying FIPS 140-2 Level 3 HSM integration and zero-trust mTLS.
10. Draft Section 9: "Explicit Anti-Goals", listing banned architectural patterns (synthetic tokens, unregulated stablecoins, public crypto dex bridges, unbacked fractionalization).
11. Generate Mermaid sequence diagrams illustrating the end-to-end custody and settlement loop.
12. Configure CI markdown linting and link-checking workflows to enforce doc integrity.
13. Conduct formal review and obtain sign-off from Lead Architect, Chief Compliance Officer, and Security Lead.

## Interfaces / Contracts
```yaml
# Invariant Validation Schema (docs/architecture/schema/invariants.yaml)
system_invariants:
  version: "1.0.0"
  asset_backing:
    type: "1:1_physical_custody"
    depository: "NSDL_CDSL"
    synthetic_allowed: false
  fiat_rails:
    domestic: ["UPI", "IMPS", "NEFT", "RTGS"]
    unregulated_stablecoins_allowed: false
  ledger:
    type: "permissioned_consortium"
    node_engine: "Hyperledger_Besu"
    consensus: "QBFT"
    block_time_seconds: 2
    on_chain_pii_allowed: false
  monetization:
    fee_type: "fixed_transaction_turnover"
    platform_fee_rate: 0.0000 # 0.00% Zero Fee at launch (FeeController governed) # 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) on trade notional turnover
    revenue_split:
      governance: "FeeController.sol"
      timelock_hours: 48
      max_fee_ceiling_bps: 50
      maker_fee_bps: 0 # 0.00% Zero-Fee at launch
      taker_fee_bps: 0 # 0.00% Zero-Fee at launch
    tax_compliance_engine: "FIFO_capital_gains_Section_111A_112A"
    holding_fee_rate: 0.0
    aum_fee_rate: 0.0
  security:
    key_custody: "FIPS_140_2_L3_HSM"
    inter_service_auth: "mTLS_SPIFFE_SPIRE"
```

## Security & Compliance Notes
- Strict separation of investor identity from public blockchain addresses; identity-to-address mapping resides exclusively in off-chain, encrypted databases protected by envelope encryption (AES-256-GCM + Vault KMS).
- Root administrative keys do not exist; all governance actions require M-of-N threshold multi-signature execution backed by hardware security modules (HSMs).
- Complete compliance alignment with SEBI Intermediary Regulations, RBI Payment & Settlement Systems Act, and IFSCA Fintech Regulatory Framework.

## Acceptance Criteria
- [ ] `docs/architecture/000_north_star.md` is fully drafted, covering all 10 non-negotiable invariants and anti-goals.
- [ ] Clear architectural comparison matrix between Growww, traditional stockbrokers, and crypto exchanges is documented.
- [ ] End-to-end fractional unit lifecycle (Fiat -> Custody -> Token -> Trade -> Burn -> Fiat) is mapped with Mermaid sequence diagrams.
- [ ] Universal Zero-Fee Model (0.00% fee - No fee at all) on trade notional turnover with 0.00% fees at launch (100% net proceeds credited; future fee adjustments governed by FeeController.sol) is formally defined, with FIFO capital gains computed strictly for user tax compliance (Section 111A/112A) and explicit exclusion of AUM/holding fees.
- [ ] Document passes all CI markdown linter checks and is signed off by Architecture and Compliance leads.

## Suggested Order / Dependencies
- Prerequisites: None (Genesis Architectural Document).
- Parallel Tasks: 001 (Glossary), 002 (Two-Entity Structure), 003 (Regulatory Pathway).
- Downstream Blockers: Blocks all subsequent architecture and implementation prompts across Categories 1 through 9.
