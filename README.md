# Growww RWA Trading Platform & Ecosystem

Welcome to the **Growww Next-Gen Real World Asset (RWA) Tokenization & Trading Platform**.

This mono-repository hosts the end-to-end architecture, applications, distributed microservices, smart contracts, shared packages, infrastructure configurations, and regulatory compliance documentation for the Growww ecosystem.

---

## 📁 Repository Structure

```
.
├── apps/                 # Client applications (Mobile, Web, Admin, Marketing)
│   ├── growww_flutter/   # Cross-platform Flutter Mobile & Desktop App (iOS, Android, macOS, Windows, Linux)
│   ├── growww_web/       # Next.js 14+ Institutional & Retail Trading Web Application
│   ├── growww_admin/     # Next.js 14+ Back-Office & Compliance Management Portal
│   └── growww_marketing/ # High-converting Next.js Public Landing Site
│
├── services/             # 24 Mission-Critical Distributed Microservices
│   ├── user-service/              # Identity, Auth, Sessions, MFA & Profile Management (Python/FastAPI)
│   ├── kyc-service/               # Video KYC, Aadhaar e-Sign, PAN Verification, Sanctions (Python/FastAPI)
│   ├── wallet-service/            # Double-Entry Core Ledger, Fiat & Token Wallets (Go)
│   ├── order-service/             # Order Validation, Lifecycle & Routing (Go)
│   ├── matching-engine/           # Sub-10μs In-Memory Order Book & Matching Engine (Rust)
│   ├── risk-service/              # Pre-Trade Margin & Real-Time Circuit Breakers (Go)
│   ├── market-data-service/       # Level-2/3 Depth, Ticker Feeds & High-Fanout WebSockets (Go)
│   ├── settlement-service/        # DvP Atomic On-Chain & Off-Chain Settlement Orchestration (Go)
│   ├── portfolio-service/         # Real-time Valuation, Tax-lot Tracking, NAV Engine (Python/FastAPI)
│   ├── fee-engine/                # Universal Zero-Fee Engine (0.00% fee - No fee at all) (Go)
│   ├── notification-service/      # Multi-Channel Push, SMS (TRAI DLT), Email & In-App Alerts (Go)
│   ├── payment-service/           # UPI 2.0 AutoPay, NetBanking, IMPS, RTGS, NACH Mandates (Go)
│   ├── custody-adapter/           # Integration with NSDL, CDSL & Institutional Custodians (Go)
│   ├── gift-city-funding/         # IFSC Banking, Nostro/Vostro Accounts, FX Escrow & LRS Flow (Python/FastAPI)
│   ├── reconciliation-service/    # Multi-Leg 3-Way Reconciliation (Bank, Ledger, Chain) (Go)
│   ├── reporting-service/         # SEBI, RBI, FIU-IND, IFSCA Statutory Regulatory Filings (Python/FastAPI)
│   ├── admin-service/             # RBAC, Four-Eyes Approval Workflow, Operational Controls (Go)
│   ├── audit-service/             # Append-Only Immutable Cryptographic Audit Log Engine (Rust)
│   ├── api-gateway/               # High-Performance API Gateway, Envoy L7 & BFF Routing (Go)
│   ├── abuse-prevention/          # Rate Limiting, DDoS Mitigation, Bot Shield & Fraud Detection (Go)
│   ├── search-service/            # Full-Text & Faceted Asset Discovery Engine (OpenSearch) (Go)
│   ├── corporate-actions/         # Dividends, Coupon Payouts, Splits, Rights Issuance Engine (Python/FastAPI)
│   ├── tax-service/               # Decoupled Voluntary Tax Statements & PnL Accounting Engine (Go)
│   └── referral-service/          # Growth Campaigns, Referral Tracking & Gamified Rewards (Go)
│
├── contracts/            # Smart Contracts (Solidity / Foundry)
│   ├── tokens/           # ERC-3643 / ERC-1400 Compliant RWA Security Tokens
│   ├── settlement/       # Delivery versus Payment (DvP) Atomic Settlement Contracts
│   ├── compliance/       # Identity Registries & Transfer Approval Compliance Hooks
│   ├── governance/       # Multi-Sig Custody, Emergency Pausing & Upgrade Timelocks
│   ├── reserves/         # Chainlink / Automated Proof of Reserve (PoR) Attestation
│   └── bridge/           # Cross-Entity Ledger Bridge for Domestic <-> GIFT City Flow
│
├── packages/             # Shared Libraries & Tooling
│   ├── proto/            # Protocol Buffers (gRPC & Event Schemas)
│   ├── growww_ui/        # Shared Flutter Design System & UI Components
│   ├── web_ui/           # Shared React/Tailwind UI Component Library
│   ├── crypto_utils/     # Multi-Language Cryptographic Primitives & Signatures
│   └── domain_types/     # Canonical Financial Domain Types & Data Contracts
│
├── infra/                # Infrastructure, Deployment & Observability
│   ├── terraform/        # Infrastructure as Code (AWS/GCP, EKS/GKE, RDS, MSK, Redis)
│   ├── k8s/              # Kubernetes Helm Charts, Manifests, HPA & Network Policies
│   ├── blockchain/       # Hyperledger Besu Validator Node Configurations & Genesis
│   ├── observability/    # Prometheus, Grafana Dashboards, OpenTelemetry, Loki
│   └── local/            # Full-Stack Local Development Docker Compose Setup
│
├── prompts/              # Master System Design & Build Prompt Specifications (951 Prompts)
│   ├── README.md         # Master Prompt Architecture Constitution & Roadmap
│   ├── SECURITY.md       # Zero-PII, FIPS 140-2 Level 3 HSM & Defense-in-Depth Model
│   ├── prompt.md         # Master Compiler Prompt & Build Directive Spec
│   ├── 00_INDEX.md       # Authoritative Master Catalog & Category Index
│   └── 000_*.md to 950_*.md # 951 Self-Contained 12-Section Architecture Build Prompts
│
└── docs/                 # Documentation & Runbooks
    ├── FEATURES.md               # Master Platform Features & Capabilities Catalog
    ├── PRODUCT_DEFINITION.md     # Phase 1 (Broker) vs Phase 2 (Exchange) Definition
    ├── IMPLEMENTATION_STATUS.md  # 285-Defect Audit & Remediation Status Matrix
    ├── architecture/             # Master Blueprints, Core Engine & Ledger Specs
    ├── adr/                      # 46 Architecture Decision Records (ADR-0001 to ADR-0046)
    ├── runbooks/                 # 41 SRE Operational Runbooks (RUNBOOK-01 to RUNBOOK-041)
    ├── compliance/               # SEBI, RBI, FIU-IND & IFSCA Regulatory Rulebooks
    └── api/                      # OpenAPI 3.1 & gRPC Reference Specifications
```

---

## 🚀 Quick Start (Local Development)

To spin up the complete local infrastructure (PostgreSQL 16 with 8 isolated databases, Redis, Apache Kafka, Hyperledger Besu Local Node):

```bash
# Start all infrastructure containers
make infra-up

# Verify connectivity and health
make smoke

# Stop local infrastructure
make infra-down
```

---

## 📜 Architectural Specifications & Governance

- **Platform Features & Capabilities Catalog:** [`docs/FEATURES.md`](docs/FEATURES.md)
- **Options Pricing & SPAN-Style Portfolio Margin:** [`docs/architecture/OPTIONS_AND_VOLATILITY_TRADING_SPECIFICATION.md`](docs/architecture/OPTIONS_AND_VOLATILITY_TRADING_SPECIFICATION.md)
- **Copy Trading & Social Strategy Marketplace:** [`docs/architecture/COPY_TRADING_AND_SOCIAL_STRATEGY_SPECIFICATION.md`](docs/architecture/COPY_TRADING_AND_SOCIAL_STRATEGY_SPECIFICATION.md)
- **Hybrid CLOB-AMM Routing & Concentrated Liquidity:** [`docs/architecture/HYBRID_CLOB_AMM_ROUTING_SPECIFICATION.md`](docs/architecture/HYBRID_CLOB_AMM_ROUTING_SPECIFICATION.md)
- **RWA Launchpad & Primary Asset Offerings:** [`docs/architecture/RWA_LAUNCHPAD_AND_PRIMARY_ISSUANCE_SPECIFICATION.md`](docs/architecture/RWA_LAUNCHPAD_AND_PRIMARY_ISSUANCE_SPECIFICATION.md)
- **MPC Key Ceremonies & Multi-Region Disaster Recovery:** [`docs/architecture/MPC_KEY_CEREMONIES_AND_DISASTER_RECOVERY_SPECIFICATION.md`](docs/architecture/MPC_KEY_CEREMONIES_AND_DISASTER_RECOVERY_SPECIFICATION.md)
- **Regulatory Grievance Redressal & Support Gateway:** [`docs/architecture/REGULATORY_GRIEVANCE_AND_SUPPORT_GATEWAY_SPECIFICATION.md`](docs/architecture/REGULATORY_GRIEVANCE_AND_SUPPORT_GATEWAY_SPECIFICATION.md)
- **Perpetual Futures, Derivatives & Margin Trading:** [`docs/architecture/DERIVATIVES_PERPETUAL_FUTURES_AND_MARGIN_SPECIFICATION.md`](docs/architecture/DERIVATIVES_PERPETUAL_FUTURES_AND_MARGIN_SPECIFICATION.md)
- **Cryptographic Merkle Proof of Solvency:** [`docs/architecture/MERKLE_PROOF_OF_SOLVENCY_AND_RESERVE_SPECIFICATION.md`](docs/architecture/MERKLE_PROOF_OF_SOLVENCY_AND_RESERVE_SPECIFICATION.md)
- **Institutional FIX Protocol & Binary Feeds:** [`docs/architecture/INSTITUTIONAL_FIX_PROTOCOL_AND_BINARY_FEEDS_SPECIFICATION.md`](docs/architecture/INSTITUTIONAL_FIX_PROTOCOL_AND_BINARY_FEEDS_SPECIFICATION.md)
- **Peer-to-Peer (P2P) Fiat Escrow & Dispute Resolution:** [`docs/architecture/P2P_ESCROW_AND_DISPUTE_RESOLUTION_SPECIFICATION.md`](docs/architecture/P2P_ESCROW_AND_DISPUTE_RESOLUTION_SPECIFICATION.md)
- **Crypto Earn, PoS Staking & Yield Vaults:** [`docs/architecture/CRYPTO_EARN_STAKING_AND_YIELD_VAULTS_SPECIFICATION.md`](docs/architecture/CRYPTO_EARN_STAKING_AND_YIELD_VAULTS_SPECIFICATION.md)
- **Multi-Tier Affiliate, Referral & Rebates:** [`docs/architecture/AFFILIATE_REFERRAL_AND_REBATE_ENGINE_SPECIFICATION.md`](docs/architecture/AFFILIATE_REFERRAL_AND_REBATE_ENGINE_SPECIFICATION.md)
- **Advanced Order Types & Algorithmic Execution:** [`docs/architecture/ADVANCED_ORDER_TYPES_AND_EXECUTION_ALGORITHMS_SPECIFICATION.md`](docs/architecture/ADVANCED_ORDER_TYPES_AND_EXECUTION_ALGORITHMS_SPECIFICATION.md)
- **Crypto Deposit, Withdrawal & 3-Tier Custody:** [`docs/architecture/CRYPTO_DEPOSIT_WITHDRAWAL_AND_CUSTODY_SPECIFICATION.md`](docs/architecture/CRYPTO_DEPOSIT_WITHDRAWAL_AND_CUSTODY_SPECIFICATION.md)
- **User Experience & Exchange Scenarios:** [`docs/architecture/USER_EXPERIENCE_AND_EXCHANGE_SCENARIOS_SPECIFICATION.md`](docs/architecture/USER_EXPERIENCE_AND_EXCHANGE_SCENARIOS_SPECIFICATION.md)
- **Product & Regulatory Roadmap:** [`docs/PRODUCT_DEFINITION.md`](docs/PRODUCT_DEFINITION.md)
- **Defect Remediation Matrix (All 285 Findings):** [`docs/IMPLEMENTATION_STATUS.md`](docs/IMPLEMENTATION_STATUS.md)
- **Core Matching Engine Architecture:** [`docs/architecture/CORE_MATCHING_ENGINE_AND_ORDER_BOOK_SPECIFICATION.md`](docs/architecture/CORE_MATCHING_ENGINE_AND_ORDER_BOOK_SPECIFICATION.md)
- **Candlestick Aggregation & Market Data:** [`docs/architecture/CANDLESTICK_AGGREGATION_AND_MARKET_DATA_SPECIFICATION.md`](docs/architecture/CANDLESTICK_AGGREGATION_AND_MARKET_DATA_SPECIFICATION.md)
- **Core Ledger & Precision Standards:** [`docs/architecture/CORE_LEDGER_AND_MULTI_CURRENCY_SPECIFICATION.md`](docs/architecture/CORE_LEDGER_AND_MULTI_CURRENCY_SPECIFICATION.md)
- **DvP Settlement & Default Waterfall:** [`docs/architecture/DVP_SETTLEMENT_AND_CLEARING_WATERFALL_SPECIFICATION.md`](docs/architecture/DVP_SETTLEMENT_AND_CLEARING_WATERFALL_SPECIFICATION.md)
- **Distributed Systems Resilience:** [`docs/architecture/DISTRIBUTED_SYSTEMS_AND_RESILIENCE_SPECIFICATION.md`](docs/architecture/DISTRIBUTED_SYSTEMS_AND_RESILIENCE_SPECIFICATION.md)
- **Smart Contracts & DvP Settlement:** [`docs/architecture/SMART_CONTRACTS_AND_SETTLEMENT_SPECIFICATION.md`](docs/architecture/SMART_CONTRACTS_AND_SETTLEMENT_SPECIFICATION.md)
- **Market Microstructure:** [`docs/architecture/MARKET_MICROSTRUCTURE_AND_MATCHING_SPECIFICATION.md`](docs/architecture/MARKET_MICROSTRUCTURE_AND_MATCHING_SPECIFICATION.md)
- **Regulatory Governance:** [`docs/architecture/REGULATORY_RISK_AND_CROSSCHAIN_GOVERNANCE.md`](docs/architecture/REGULATORY_RISK_AND_CROSSCHAIN_GOVERNANCE.md)
- **Architecture Decision Records (ADRs):** [`docs/adr/`](docs/adr/)
- **SRE Operational Runbooks:** [`docs/runbooks/`](docs/runbooks/)
- **Complete 951-Prompt Catalog:** [`prompts/`](prompts/)
