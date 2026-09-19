# Growww / NBSE Prompt Specification System - Root Architectural Blueprint

## 1. Executive Overview

The Growww / NBSE (National Blockchain Stock Exchange) Prompt Specification System is the authoritative, declarative engineering blueprint for a sovereign, institutional-grade 24/7 financial market infrastructure layer. The platform bridges Indian capital markets - encompassing the National Stock Exchange (NSE), Bombay Stock Exchange (BSE), Multi Commodity Exchange (MCX), National Securities Depository Limited (NSDL), and Central Depository Services Limited (CDSL) - with global institutional and retail liquidity via GIFT City (IFSCA) and a high-performance, permissioned Hyperledger Besu enterprise blockchain ledger.

Modern mission-critical financial systems cannot afford ambiguity, architectural drift, or implementation hallucinations. To solve this, the Growww / NBSE engineering architecture is governed by a declarative specification-first paradigm. Every subsystem, distributed microservice, zero-PII smart contract, risk engine, client application interface, and compliance pipeline is defined in an isolated, immutable specification sheet.

### The Declarative Build Prompt Philosophy

1. **Atomic Independence:** Each prompt document provides a self-contained execution package. An autonomous AI agent or staff software engineer can ingest a single prompt sheet along with its immediate upstream dependencies and implement the target module without needing external tribal knowledge or speculative assumptions.
2. **Determinism and Zero Hallucination:** By providing explicit Protobuf interfaces, database schemas, message-bus topologies, mathematical clearing models, and deterministic acceptance gates, the specification eliminates generative guesswork.
3. **Traceable Financial Provenance:** Every line of production code links back to a unique Prompt ID (`000` through `950`), ensuring end-to-end statutory and architectural traceability across regulatory filings, code commits, CI/CD validation pipelines, and disaster-recovery runbooks.
4. **Sovereign Regulatory Alignment:** The specification system directly embeds the regulatory frameworks of the Securities and Exchange Board of India (SEBI), Reserve Bank of India (RBI), International Financial Services Centres Authority (IFSCA), Prevention of Money Laundering Act (PMLA), and Digital Personal Data Protection Act 2023 (DPDP Act).

---

## 2. Directory Structure and Architectural Assets

The `Prompt/` root directory serves as the engineering operating system for the entire platform. The directory layout and file responsibilities are structured as follows:

```
Prompt/
├── README.md
├── SECURITY.md
├── prompt.md
├── docs/
│   ├── PRODUCT_DEFINITION.md
│   ├── FEATURES.md
│   ├── IMPLEMENTATION_STATUS.md
│   ├── ENGINEERING_STANDARDS_AND_PROVENANCE.md
│   ├── adr/
│   ├── api/
│   ├── architecture/
│   ├── compliance/
│   ├── runbooks/
│   └── vision/
└── prompts/
    ├── 00_INDEX.md
    ├── 000_project_north_star.md
    ├── ...
    └── 950_final_production_readiness_acceptance_gate.md
```

### File and Directory Responsibilities

- **`prompts/` (Atomic Prompt Repository):**
  Houses the complete library of 951 atomic, numbered specification files (ranging from `000` to `950`) plus `00_INDEX.md`. Every file follows an identical 12-section technical template. Files are categorized into ten functional tiers covering product vision, core infrastructure, microservices, smart contracts, client applications, security, DevOps, and quality assurance.
- **`README.md` (Root Architectural Blueprint):**
  This document. It serves as the master navigation index, governing standard, invariant declaration, and agent execution guide across the entire prompt corpus.
- **`SECURITY.md` (Sovereign Security & Cryptographic Policy):**
  Defines the institutional security baseline: FIPS 140-2 Level 3 Hardware Security Module (HSM) key lifecycle management, zero-PII ledger enforcement, Pedersen commitment blinding, mTLS service-mesh encryption, privileged access management (PAM), and Coordinated Vulnerability Disclosure (CVD) protocols.
- **`prompt.md` (Master Prompt Compiler & Catalog Index):**
  A centralized context sheet containing the global platform inventory, executive roles, and cross-cutting architectural mandates utilized when seeding large-context language model workflows.
- **`docs/` (Supporting Architectural Decision Records & Regulatory Dossiers):**
  Houses permanent system documentation, Architecture Decision Records (ADRs), unified API catalogs, compliance runbooks, statutory submission drafts for SEBI/RBI sandboxes, and feature roadmap trackers.

---

## 3. The 10 Core Architectural and Regulatory Invariants

Every subsystem specified within the Growww / NBSE platform must strictly adhere to the following ten non-negotiable architectural and regulatory invariants. Any implementation or pull request violating these invariants is immediately rejected at the automated CI/CD gating phase.

### 1. Real-Asset Custodial Backing (1:1 Invariant)
Every digital token or fractional share minted on the NBSE ledger must be strictly backed 1:1 by real, unencumbered physical securities, commodities, or sovereign debt held in licensed custody (NSDL, CDSL, RBI-approved depositories, or WDRA-accredited commodity vaults). Synthetic instruments without verifiable custodial allocation are strictly prohibited. Token issuance and burning are cryptographically bound to custodial deposit and withdrawal receipts.

### 2. Independent 24/7/365 Continuous Trading
The exchange operates continuously without closing intervals. Secondary trading occurs 24 hours a day, 7 days a week, 365 days a year on the NBSE internal matching engine and virtual automated market maker (vAMM) liquidity rails. During traditional domestic market sessions (09:15 to 15:30 IST), the platform synchronizes with NSE NEAT and BSE BOLT DMA gateways. Off-market liquidity is protected via dynamic volatility collars, pre-open auction queues, and liquidity-provider depth pools.

### 3. Google Gemini Conversational Financial Intelligence
Google Gemini 1.5 Pro and Flash models form the human-centric advisory and analytics plane (`services/gemini-advisor-service` and client applications). Gemini powers multilingual conversational navigation across 22 scheduled Indian languages, performs Retrieval-Augmented Generation (RAG) over real-time Proof-of-Reserve attestations, parses regulatory disclosures, and calculates deterministic portfolio stress tests under strict non-advisory SEBI research analyst guardrails.

### 4. Universal Zero-Fee Model (0.00% fee across all trading - No fee at all)
A single, uniform fee of strictly 0.00% (Zero Fee at launch) is charged across all traded notionals, cross-currency foreign exchange conversions, and platform transfers (0 bps maker / 0 bps taker). Future fee adjustments are governed dynamically via `FeeController.sol` (requiring a 48-hour timelock, 3-of-5 institutional multi-sig, and a 50 bps hard safety ceiling). If fees are activated in the future, revenue distribution is routed dynamically to the Corporate Treasury, Core Settlement Guarantee Fund (SGF), and Investor Protection Fund (IPF).
Capital gains are tracked using FIFO cost basis accounting to meet Indian Income Tax Act requirements (Sections 111A, 112A, and 194S).

### 5. Dual-Environment Testnet Sandbox vs Mainnet Institutional Isolation
The architecture enforces complete logical, physical, and cryptographic isolation between the two operating networks:
- **Testnet Sandbox (`nbse-testnet`, Chain ID 13371):** Open to developers, testing suites, mock clearing houses, eINR faucets, and chaos experiments.
- **Mainnet Institutional Production (`nbse-mainnet`, Chain ID 2026):** Production ledger restricted to authorized bank rails, institutional participants, FIPS 140-2 Level 3 HSM nodes, and live depository feeds.
Cross-environment state leakage or key reuse is prevented through compiler-enforced network guards and distinct cryptographic genesis configurations.

### 6. Permissioned Enterprise Blockchain Ledger (Hyperledger Besu QBFT)
The distributed settlement layer runs on Hyperledger Besu utilizing the QBFT (Quorum Byzantine Fault Tolerance) consensus algorithm with a 2.0-second block period, immediate deterministic single-block finality, and zero chain reorganizations. All smart contracts adhere to ERC-3643 (T-REX compliant security tokens). Identity is maintained via off-chain Claim Issuers; zero Personally Identifiable Information (PII) is ever written to the ledger.

### 7. Compliance by Design
Regulatory compliance is programmatic rather than reactive. Every state transition evaluates real-time rules: Foreign Portfolio Investor (FPI) sectoral investment caps under SEBI FPI Regulations 2019 / FEMA, Central KYC (CKYC) registry validation, AML/CFT screening against international watchlists, 500-microsecond asymmetric speed bumps to eliminate predatory high-frequency order front-running, and M-of-N multi-party governance for exceptional actions.

### 8. Multi-Platform Thick Client & Web Portals
Investors and operators access the platform via a unified, platform-adaptive Flutter client targeting Android, iOS, Windows, macOS, and Linux, complemented by a Next.js 14 server-rendered trading terminal and administrative compliance dashboard. All clients share identical domain calculation models, offline order-queuing resilience, and secure hardware keystore abstractions.

### 9. Polyglot Production Tooling
Technical stacks are selected strictly based on domain performance and reliability profiles:
- **Rust:** Ultra-low latency order matching engine, memory-mapped write-ahead logs, and core cryptographic primitives.
- **Go:** High-throughput microservices, ledger relayers, FIX protocol gateways, and distributed event coordinators.
- **Python:** AI/ML inference, Gemini LangChain integrations, and regulatory analytics pipelines.
- **Solidity:** Enterprise smart contracts audited for EVM execution on Hyperledger Besu.
- **PostgreSQL 16 & TimescaleDB:** Core double-entry relational ledger and time-series tick repository.
- **Redis Cluster:** Sub-millisecond order book cache and pre-trade margin checks.
- **Apache Kafka:** Distributed, immutable event streaming backbone.

### 10. Zero Formatting Violations (Strict ASCII Hyphenation)
To prevent cross-compiler parsing failures, Unicode normalization mismatches, terminal rendering glitches, and encoding bugs across polyglot systems, all documentation, specifications, configuration files, and code repositories must exclusively use standard ASCII hyphens (-). Non-ASCII typography such as em dashes or en dashes is strictly prohibited across the entire project.

---

## 4. The 12-Section Mandatory Prompt Specification Template

Every specification file inside `prompts/` must follow this 12-section technical template without exception. The template enforces complete specification fidelity, ensuring that autonomous AI coding agents and human developers have an unambiguous target.

```markdown
# <Prompt Number> - <Title>

## Purpose
Detailed explanation of why this subsystem exists, the business/financial problem it solves, and its role in the NBSE architecture.

## What You Are Building
Concrete deliverables, microservice paths, Protobuf contracts, smart contracts, screens, or pipelines.

## Scope Boundaries
Explicit declaration of what is IN scope and what is OUT of scope (cross-referenced with other prompt numbers).

## Technology to Use
Primary programming languages, frameworks, version targets, dependencies, and architectural justification.

## Backend / Infra Touchpoints
Databases (PostgreSQL tables), queues (Kafka topics), cache keys (Redis), external APIs (NSDL/CDSL, UPI, RBI CBDC).

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
Specific smart contracts, EIP-712 structured data hashes, events, and DvP settlement mechanics.

## Step-by-Step Build Instructions (10-15 steps)
Deterministic, numbered sequence of implementation instructions from scaffolding to tests.

## Interfaces / Contracts
Declarative Protobuf v3 definitions, PostgreSQL DDL schemas, smart contract function interfaces, or JSON schemas.

## Security & Compliance Notes
Cryptographic controls, HSM key lifecycle, PII masking, SEBI/RBI/DPDP Act 2023 compliance, audit trails.

## Acceptance Criteria
Testable checklist of definition-of-done criteria including performance benchmarks and invariant assertions.

## Suggested Order / Dependencies
Prerequisites, parallel development tasks, and downstream dependencies.
```

### Detailed Breakdown of Template Sections

1. **Purpose:**
   Articulates the business rationale, economic foundation, and operational necessity of the subsystem. It defines the specific problem being addressed within the Indian capital market and international investment pipeline.
2. **What You Are Building:**
   Lists the tangible engineering artifacts produced when the prompt is completed: directory paths, service binaries, gRPC definitions, database tables, Flutter screens, or smart contracts.
3. **Scope Boundaries:**
   Prevents scope creep and cross-subsystem leakage. It explicitly marks boundaries with an `IN SCOPE` checklist and an `OUT OF SCOPE` table referencing other prompt IDs responsible for adjacent components.
4. **Technology to Use:**
   Mandates the exact programming languages, libraries, versions, and runtime constraints. It enforces performance guardrails (such as zero-allocation loops in Rust or connection-pooling parameters in Go).
5. **Backend / Infra Touchpoints:**
   Maps the subsystem to the wider platform infrastructure: specific relational tables, Redis key spaces, Kafka event streams, and external depository or banking APIs.
6. **Blockchain Interaction:**
   Specifies all ledger interactions on Hyperledger Besu. It covers EIP-712 hash structures, smart contract function calls, on-chain event emissions, gas schedules, and zero-PII privacy guarantees.
7. **Step-by-Step Build Instructions (10-15 steps):**
   Provides an implementation plan structured as 10 to 15 sequential actions. It guides the builder from initial project scaffolding through interface design, business logic, storage access, and test suites.
8. **Interfaces / Contracts:**
   Defines the programmatic boundaries: Protobuf v3 schemas, relational DDL schemas, OpenAPI specifications, or Solidity function signatures. Implementations must match these declarations.
9. **Security & Compliance Notes:**
   Specifies cryptographic standards, data sanitization rules, PII isolation mechanics, SEBI/RBI audit obligations, and threat mitigations mapped to STRIDE categories.
10. **Acceptance Criteria:**
    A checklist of verifiable assertions required for task sign-off. It combines functional correctness, automated test coverage, performance thresholds (such as 99th-percentile latency), and regulatory invariant tests.
11. **Suggested Order / Dependencies:**
    Defines the directed acyclic graph (DAG) of prerequisites, parallel development opportunities, and downstream systems dependent on this prompt.

---

## 5. Complete Directory Roadmap & Category Breakdown

The NBSE specification system contains 378 standalone build specifications organized into ten logical categories (Categories 0 through 9). Below is the comprehensive architectural roadmap:

```
+---------------------------------------------------------------------------------------------------+
|                            Growww / NBSE Specification Hierarchy                                  |
+---------------------------------------------------------------------------------------------------+
| Category 0: Vision, Domain & Product Specifications            | 11 Prompts (000 - 010)           |
| Category 1: System Architecture & Foundation Infrastructure    | 14 Prompts (101 - 114)           |
| Category 2: Backend Microservices                              | 79 Prompts (201 - 279)           |
| Category 3: Blockchain & Smart Contracts                       | 50 Prompts (301 - 350)           |
| Category 4: Shared Libraries, Protocols & SDKs                 | 13 Prompts (401 - 413)           |
| Category 5: Cross-Platform Mobile & Desktop Client             | 43 Prompts (501 - 543)           |
| Category 6: Web Applications                                   | 21 Prompts (601 - 621)           |
| Category 7: Security, Compliance & Market Surveillance         | 22 Prompts (701 - 722)           |
| Category 8: DevOps, CI/CD, Infrastructure as Code             | 17 Prompts (801 - 817)           |
| Category 9: End-to-End Testing, Chaos & QA Automation          | 18 Prompts (901 - 918)           |
+---------------------------------------------------------------------------------------------------+
| Total Authoritative Specifications                             | 378 Prompts (000 - 918)          |
+---------------------------------------------------------------------------------------------------+
```

### Category 0: High-Level Vision, Domain & Product Specifications (11 Prompts: 000 - 010)
- **Domain Scope:** Defines the legal, regulatory, economic, and sovereign foundation of the exchange.
- **Key Modules:**
  - `000`: Project North Star and high-level architectural boundaries.
  - `001`: Ubiquitous domain language, securities glossaries, and financial terminology.
  - `002`: Two-entity corporate separation (Domestic Indian Broker/Dealer vs GIFT City International Gateway).
  - `003` - `005`: Regulatory sandbox pathways with SEBI, RBI, and IFSCA, accompanied by domestic and cross-border KYC/AML frameworks.
  - `006`: Universal 0.00% fee (No fee at all) allocation, FIFO capital gains accounting, and Treasury/SGF/IPF distribution.
  - `007` - `010`: Proof-of-Reserve transparency, DPDP zero-PII policies, investor disclosures, and master non-functional requirements.

### Category 1: System Architecture & Foundation Infrastructure (14 Prompts: 101 - 114)
- **Domain Scope:** Establishes cross-cutting architectural patterns, inter-service contracts, and monorepo governance.
- **Key Modules:**
  - `101` - `102`: C4 system architecture diagrams and Domain-Driven Design (DDD) bounded contexts.
  - `103` - `105`: Unified REST/gRPC API standards, CloudEvents schema registries on Kafka, and OAuth2/OIDC/mTLS security meshes.
  - `106` - `107`: Monorepo directory topology, polyglot build orchestration, and strict linting rules.
  - `108` - `110`: 12-factor configuration management, HashiCorp Vault secrets management, and cross-entity encrypted channels.
  - `111` - `112`: Canonical domain entity definitions and distributed idempotency patterns.
  - `113` - `114`: Microsecond OpenTelemetry distributed tracing standards and wire protocol schema evolution.

### Category 2: Backend Microservices (79 Prompts: 201 - 279)
- **Domain Scope:** High-performance, distributed backend services executing core order matching, trade settlement, banking integrations, and real-time market data distribution.
- **Key Modules:**
  - `201` - `204`: Identity management, sanctions screening, double-entry ledger accounts, and order management services.
  - `205`: Sub-millisecond in-memory Rust matching engine with deterministic price-time priority.
  - `206` - `210`: Pre-trade risk checks, real-time WebSocket market data streaming, DvP trade settlement, fractional portfolio accounting, and capital gains calculation.
  - `211` - `215`: Multi-channel notifications, banking gateways (UPI, IMPS, RTGS), depository adapters (NSDL/CDSL), GIFT City FX services, and on-chain ledger reconcilers.
  - `216` - `224`: Regulatory reporting dispatchers, dual-control back-office operations, audit logging, rate limiting, and corporate action engines.
  - `225` - `233`: FIX 5.0 SP2 binary gateways, advanced algorithmic orders, institutional dark pools, surveillance engines, VaR/ELM margin engines, settlement guarantee fund managers, fail-to-deliver auction resolvers, and 24/7 RBI eINR CBDC adapters.
  - `234` - `238`: Cross-chain multi-token ingress gateways (Bitcoin Lightning/Taproot, EVM CCIP, Solana SPL/Wormhole) and institutional MPC-TSS custody vaults.
  - `239` - `241`: Options pricing with SIMD Greeks, perpetual futures clearing engines, and portfolio SPAN margin calculators.
  - `242` - `249`: Primary market routing to NSE/BSE DMA, MCX warehouse receipt integrations, relayer gas escalators, memory-mapped WAL failover, cross-chain saga coordinators, and speed-bump guards.
  - `250` - `256`: Real-time FPI sectoral caps, FX haircut calculators, Gemini financial intelligence microservices, 24/7 after-hours liquidity gateways, bond yield engines, cross-asset collateral optimizers, and Limit-Up/Limit-Down (LULD) volatility dampeners.
  - `257` - `264`: Institutional algorithmic execution (TWAP/VWAP/POV/Iceberg), real-time NBBO consolidated tape, ERC-4337 bundler & paymasters, clearing corporation interoperability, tick-by-tick market replay daemons, market maker incentive engines, SEBI margin pledge gateways, and dynamic collateral haircut calculators.
  - `265` - `271`: Proportional copy trading replication engines, primary RWA Dutch auction token issuance, SEBI SCORES 2.0 & regulatory grievance gateways, real-time auto-deleveraging (ADL) resolvers, P2P fiat escrow & multi-sig arbitration services, instant RFQ convert swaps, and multi-tier affiliate referral rebate engines.
  - `272` - `279`: Bitcoin (BTC/USDT) live external market data feeder, demo paper trading virtual matching engine, 10k USDT testnet faucet & virtual balance ledger, real-money spot order execution OMS, real-time depth broadcaster, Bitcoin UTXO SegWit/Taproot deposit listener, multi-chain USDT ingress, and real-money crypto withdrawal clearing.

### Category 3: Blockchain & Smart Contracts (50 Prompts: 301 - 350)
- **Domain Scope:** The sovereign, permissioned Hyperledger Besu settlement fabric, ERC-3643 asset tokenization contracts, and cross-chain custody bridges.
- **Key Modules:**
  - `301` - `302`: Hyperledger Besu enterprise evaluation and QBFT 4-node validator topology setup.
  - `303` - `306`: ERC-3643 asset token issuance, redemption contracts, on-chain compliance transfer hooks, and atomic Delivery-versus-Payment (`SettlementDvP.sol`).
  - `307` - `312`: M-of-N multi-signature governance, on-chain Proof-of-Reserve Merkle registries, event indexers, node health monitors, validator CloudHSM key signers, and transparent proxy upgrade patterns.
  - `313` - `315`: Cross-entity bridges, RocksDB ledger disaster-recovery snapshots, and on-chain Settlement Guarantee Fund default waterfall smart contracts.
  - `316` - `319`: 50,000 TPS L2 appchain rollup sequencers, Groth16/Plonk ZK proof-of-solvency verifiers, shareholder proxy voting contracts, and institutional custody bridges.
  - `320` - `321`: Developer testnet cluster faucets and institutional Mainnet Genesis key generation ceremonies.
  - `322` - `324`: Bitcoin SPV header relays, EVM cross-chain liquidity bridges, and Solana-Besu light client contracts.
  - `325` - `328`: On-chain options clearing houses, perpetual futures margin vaults, multi-chain reserve registries, and decentralized multi-source oracle aggregators.
  - `329` - `336`: DvP automated fee splitters, MCX physical vault registries, primary market order bridges, commodity scrap equalizers, tokenized G-Sec bond coupon accrual engines, RBI CBDC wholesale bridge contracts, ZK light-client header verifiers, and automated e-TDS withholding ledgers.
  - `337` - `343`: Modular ERC-4337 smart accounts with session keys, Groth16 ZK investor accreditation verifiers, 24/7 off-hours CLMM AMM contracts, post-quantum ML-DSA lattice signature verifiers, multi-tier emergency circuit breaker contracts, fractional corporate action splitters, and cross-chain HTLC atomic settlement contracts.
  - `344` - `347`: RWA Dutch auction & second-by-second linear streaming vesting contracts, ERC-4626 fixed-income tokenized yield vaults, P2P 2-of-3 multi-sig arbitration escrow contracts, and on-chain affiliate rebate commission splitters.
  - `348` - `350`: Virtual testnet faucet & non-transferable demo token minting contracts, production BTC/USDT atomic Delivery-versus-Payment (DvP) settlement smart contract, and institutional MPC-TSS hot/cold vault multi-sig coordinator.

### Category 4: Shared Libraries, Protocols & SDKs (13 Prompts: 401 - 413)
- **Domain Scope:** Foundation data storage layers, message topologies, and common cryptographic models.
- **Key Modules:**
  - `401` - `403`: Master PostgreSQL 16 schema design, Redis Cluster caching topologies, and Kafka partitioning standards.
  - `404` - `406`: ClickHouse analytics data warehouses, automated cold-storage data retention, and zero-loss backup automation.
  - `407` - `409`: Master security catalogs, TimescaleDB tick pipelines, and Sparse Merkle Sum Tree Proof-of-Reserve state generators.
  - `410` - `413`: Debezium CDC transactional outbox pipelines, pgvector financial RAG knowledge stores, Apache Flink real-time stream analytics feature stores, and S3 Glacier immutable WORM regulatory archival.

### Category 5: Cross-Platform Mobile & Desktop Client (43 Prompts: 501 - 543)
- **Domain Scope:** Multi-platform Flutter applications delivering consumer and pro-trader experiences on mobile (Android, iOS) and desktop (Windows, macOS, Linux).
- **Key Modules:**
  - `501` - `503`: Monorepo Flutter scaffolding, Riverpod state management, and platform-adaptive theme design systems.
  - `504` - `507`: DigiLocker/Aadhaar KYC flows, biometric authentication, home dashboards, and real-time market watchlists.
  - `508` - `513`: TradingView chart integration, order placement flows, fractional portfolio views, wallet fund transfers, trade history logs, and notifications.
  - `514` - `515`: User profile settings and offline-resilient queued order (AMO) management.
  - `516` - `520`: Native platform packaging, signing, and notarization (Google Play, Apple App Store, MSIX, Snap/Flatpak/AppImage, and macOS DMG).
  - `521` - `526`: Local hardware secure storage abstractions (Keychain, KeyStore, DPAPI, Libsecret), push notifications, multi-language localization (22 Indian languages), crash telemetry, typed API client layers, and universal deep links.
  - `527` - `531`: In-app sandbox mode toggles, multi-chain crypto deposit screens, options chains, physical commodity redemption flows, and the Gemini conversational assistant drawer.
  - `532` - `535`: Multi-window desktop institutional terminals, FIDO2 Passkeys hardware key signing flows, high-performance canvas & TradingView charting engines, and real-time sector treemap heatmaps.
  - `536` - `539`: Pro-trader copy trading & strategy leaderboards, RWA Dutch auction primary subscription screens, P2P fiat trading with end-to-end encrypted chat escrows, and fixed-income crypto earn staking screens.
  - `540` - `543`: Flutter BTC/USDT live TradingView chart screen, demo vs real trading mode switcher banner & faucet modal, high-speed buy/sell order entry bottom sheet, and crypto deposit/withdrawal modal.

### Category 6: Web Applications (21 Prompts: 601 - 621)
- **Domain Scope:** Server-rendered web portals for institutional traders, retail web users, exchange operators, and regulatory compliance auditors.
- **Key Modules:**
  - `601` - `603`: Next.js 14 web app scaffolding, web DigiLocker camera onboarding, and pro-trader web terminal.
  - `604` - `607`: Back-office operator consoles: KYC review queues, multi-party risk exception workflows, public/admin Proof-of-Reserve reconciliation portals, and regulatory export dashboards.
  - `608` - `611`: Public marketing education portals, developer documentation hubs with Web3 testnet faucets, web options chain builders, and institutional commodity physical delivery management interfaces.
  - `612` - `615`: Next.js 14 institutional DMA workstations, regulatory audit & SEBI supervisory dashboards, clearing member capital adequacy portals, and RWA tokenization issuer portals.
  - `616` - `619`: Pro-trader copy trading strategy management portals, SEBI SCORES 2.0 investor grievance management portals, P2P dispute arbitration operator desks, and RWA primary launchpad investor terminals.
  - `620` - `621`: Next.js 14 BTC/USDT pro-trading terminal with streaming Level-2 depth ladder, and dedicated demo paper trading simulator & strategy analytics portal.

### Category 7: Security, Compliance, Auditing & Market Surveillance (22 Prompts: 701 - 722)
- **Domain Scope:** Defense-in-depth infrastructure, market abuse detection, cryptographic key security, and privacy compliance.
- **Key Modules:**
  - `701` - `703`: Platform-wide STRIDE threat models, least-privilege IAM/RBAC models, and automated PEP/sanctions screening.
  - `704` - `706`: Real-time AML transaction monitoring, scheduled penetration testing plans, and SOC incident response playbooks.
  - `707` - `710`: Envelope encryption with automated key rotation, secure SDLC review mandates, vendor checklists, and business continuity disaster recovery plans.
  - `711` - `713`: High-frequency market surveillance (spoofing, layering, wash trading, momentum ignition), insider trading graph analytics, and automated market-wide circuit breakers.
  - `714` - `718`: Blockchain forensic transaction screeners, SEBI cross-market manipulation reporting, MCX physical vault assayer portals, CloudHSM signing daemons, and DPDP/GDPR ZK Pedersen commitment blinding engines.
  - `719` - `721`: Sybil ring & cross-account collusion graph surveillance engines, Zero-Knowledge Proof of Liabilities (ZK-PoL) Sparse Merkle Tree engines, and Post-Quantum Cryptography (PQC) hybrid TLS migrations.

### Category 8: DevOps, CI/CD, Infrastructure as Code & Deployment (17 Prompts: 801 - 817)
- **Domain Scope:** Automated provisioning, immutable deployment pipelines, Kubernetes cluster topologies, and dual-environment infrastructure management.
- **Key Modules:**
  - `801` - `802`: Full-stack Docker Compose developer environments and production multi-tenant Kubernetes configurations.
  - `803` - `805`: Polyglot GitHub Actions CI verification matrix, GitOps ArgoCD progressive canary delivery, and Terraform/OpenTofu cloud provisioning.
  - `806` - `809`: OpenTelemetry distributed tracing, centralized immutable SIEM audit pipelines, mission-critical alerting runbooks, and FinOps cost allocation.
  - `810` - `813`: Multi-environment release management, dual-environment Testnet vs Mainnet CI/CD gating, isolated staging orchestrations, and high-fidelity NSDL/CDSL/banking simulator harnesses.
  - `814` - `816`: Bare-metal DPDK kernel-bypass networking architectures, zero-trust SPIFFE/SPIRE workload identity meshes, and multi-region active-active geo-failover fabrics.

### Category 9: End-to-End Testing, Chaos Engineering & QA Automation (18 Prompts: 901 - 918)
- **Domain Scope:** Comprehensive quality assurance, non-functional performance validation, regulatory sandbox simulations, and fault injection.
- **Key Modules:**
  - `901` - `903`: Unit and integration test suites, Flutter-to-backend integration tests, and matching engine stress benchmarks.
  - `904` - `905`: Chaos Mesh infrastructure resilience experiments and automated CI/CD SAST/DAST security scanners.
  - `906` - `908`: SEBI regulatory sandbox UAT scenarios, pilot launch phasing, and zero-downtime production cutover/rollback runbooks.
  - `909` - `910`: SLO error budget governance and SEBI SCORES grievance redressal integration.
  - `911` - `914`: Mainnet disaster-recovery dress rehearsals, cross-chain bridge stress testing, automated cross-market settlement fuzzers, and the 1-Crore (10,000,000 user) distributed concurrency load harness.
  - `915` - `917`: Deterministic market replay & flash crash simulators, formal verification & symbolic execution suites, and high-concurrency copy trading replication & Dutch auction stress suites.

---

## 6. How Autonomous Coding Agents and Developers Execute Prompts

To maintain uncompromising quality and prevent architectural degradation, human developers and autonomous AI agents must execute each prompt specification according to a structured protocol.

```
+-----------------------------------------------------------------------------------+
|                        Prompt Execution Workflow Cycle                            |
+-----------------------------------------------------------------------------------+
| 1. Ingestion & Boundary Analysis   -> Read specification, verify upstream deps     |
| 2. Interface Generation            -> Scaffold Protobuf / DDL / Solidity specs    |
| 3. Core Implementation             -> Implement logic within designated tech stack |
| 4. Security & Invariant Validation -> Verify 10 Invariants, zero-PII, HSM rules    |
| 5. Automated Verification Gate     -> Run unit, integration, and benchmark tests  |
+-----------------------------------------------------------------------------------+
```

### 1. Ingestion and Scope Analysis
- Read the target specification file completely.
- Review the `Scope Boundaries` section. Confirm understanding of what is strictly `IN SCOPE` and ensure no features marked `OUT OF SCOPE` are implemented.
- Check `Suggested Order / Dependencies`. Verify that all prerequisite prompt specifications have been completed and verified.

### 2. Interface Generation First
- Extract the contracts defined in Section 8 (`Interfaces / Contracts`).
- Implement the interface stubs first (compile Protobuf schemas into language stubs, apply database migrations, or generate Solidity interfaces).
- Enforce strict type safety. Do not alter contract field numbers, message names, or database column types without an authorized change process.

### 3. Implementation and Behavioral Discipline
- Write production code strictly within the declared technology stack and version targets.
- Implement business logic following the steps in Section 7 (`Step-by-Step Build Instructions`).
- Maintain zero external runtime dependencies beyond those explicitly sanctioned in Section 4 (`Technology to Use`).
- Eliminate code duplication by importing common utilities from Category 4 libraries.

### 4. Invariant and Security Auditing
- Verify that every database query and state transition upholds the 10 Core Invariants.
- Confirm that no PII is emitted to Kafka streams or stored in Hyperledger Besu transaction payloads.
- Ensure that monetary calculations, fees, and fractional share allocations use fixed-point arithmetic or high-precision decimal libraries, avoiding floating-point rounding errors.

### 5. Passing Acceptance Gates
- Implement unit tests covering every item listed in Section 10 (`Acceptance Criteria`).
- Execute local performance benchmarks to verify latency and throughput targets (such as matching engine sub-millisecond turnarounds).
- Format all source files and markdown documents to adhere to the zero non-ASCII hyphenation rule.

---

## 7. Verification, Acceptance Testing, and Traceability

To ensure institutional compliance and software robustness, the NBSE repository enforces bidirectional traceability across specifications, implementations, and test assertions.

### Traceability Architecture

Every production artifact must carry an explicit link to its governing Prompt ID:

1. **Commit Message Conventions:**
   Commit subjects must follow Conventional Commits prefixed with the prompt identifier:
   `feat(prompt-205): implement in-memory order book price-time priority matching`
   `test(prompt-306): add dvp settlement atomic delivery rejection test suite`
2. **Source Code Annotation:**
   Root package declarations and module header documentation must reference the authoritative prompt:
   `// Authoritative Spec: Prompt/prompts/205_order_matching_engine.md`
3. **Database Migration Tracking:**
   SQL migration files must prefix migration names with the corresponding prompt number:
   `V203__wallet_account_double_entry_schema.sql`
4. **Protobuf Package Namespaces:**
   Package declarations map directly to prompt domains:
   `package nbse.matching.v1; // Reference: Prompt 205`

### Acceptance Verification Matrix

Before any pull request or agent patch is merged into the `main` branch, the automated CI pipeline (`Prompt 803`) executes a verification matrix:

- **Static Analysis & Linting:** Enforces zero compiler warnings, static analysis passes (using `cargo clippy`, `golangci-lint`, `ruff`, and `flutter analyze`), and strict hyphenation checks verifying the absence of Unicode em or en dashes.
- **Contract Compatibility Check:** Validates that Protobuf definitions and OpenAPI schemas remain backwards-compatible using schema registry linters.
- **Unit and Integration Coverage:** Requires a minimum of 85% branch coverage on core financial calculations, balance tracking, and order matching services.
- **Deterministic Simulation Testing:** High-throughput settlement components are run against simulated mock depository feeds (`Prompt 813`) to verify zero transaction failure under randomized network latency.
- **Formal Invariant Verification:** Automated test harnesses evaluate system states against the 10 Invariants, proving that total issued tokens match depository balances and confirming that every transaction fee splits precisely into the 0.00% fee launch policy ratio.

Through this specification system, the Growww / NBSE platform delivers a reproducible, verifiable, and sovereign financial infrastructure layer engineered for continuous 24/7 financial markets.
