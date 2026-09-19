# Growww / NBSE Complete Repository Directory Structure

This document outlines the authoritative, detailed repository directory layout for the **Growww / NBSE Sovereign Blockchain Spot Exchange Platform**.

All directory paths are established and clean, containing zero executable code files, allowing developers and autonomous agents to implement each subsystem according to its corresponding declarative build specification in `Prompt/prompts/`.

---

## 1. Directory Tree Overview

```
Growww/
├── Prompt/                                # Master Prompt & Specification Repository
│   ├── README.md                          # Root Architectural Blueprint & 10 Invariants
│   ├── SECURITY.md                        # Master Sovereign Security Specification
│   ├── prompt.md                          # Master Prompt Context Compiler
│   └── prompts/                           # 951 Atomic Build Specifications (.md only)
│       ├── 00_INDEX.md                    # Master Numerical Catalog (000 to 918)
│       ├── 000_project_north_star.md
│       ├── ...
│       └── 918_btc_usdt_demo_and_real_trading_e2e_stress_suite.md
│
├── docs/                                  # Architectural Dossiers & Regulatory Submissions
│   ├── README.md                          # Master Documentation Portal
│   ├── FEATURES.md                        # Complete Platform Features Catalog
│   ├── IMPLEMENTATION_STATUS.md           # Governance & Status Tracking Matrix
│   ├── PRODUCT_DEFINITION.md              # Two-Phase Legal & Corporate Definition
│   ├── ENGINEERING_STANDARDS_AND_PROVENANCE.md # Engineering Standards
│   ├── spot_trading/                      # Dedicated Spot Trading on Blockchain Suite
│   │   ├── 01_SPOT_TRADING_EXCHANGE_MASTER_BLUEPRINT.md
│   │   ├── 02_DARK_MODE_ATTRACTIVE_UI_UX_DESIGN_SYSTEM.md
│   │   ├── 03_INDIAN_MARKET_AND_CRYPTO_FUSION_ARCHITECTURE.md
│   │   ├── 04_BLOCKCHAIN_SPOT_SETTLEMENT_AND_DVP_ENGINE.md
│   │   ├── 05_SPOT_ORDER_TYPES_MATCHING_AND_EXECUTION.md
│   │   ├── 06_DEMO_SANDBOX_VS_REAL_MONEY_TRADING_ENGINE.md
│   │   ├── 07_GLOBAL_EXCHANGE_MARKET_TYPES_AND_FUTURE_ROADMAP.md
│   │   ├── 08_SETTINGS_SECURITY_AND_USER_MANAGEMENT_SPECIFICATION.md
│   │   └── README.md
│   ├── architecture/                      # C4 Blueprints, Ledger & Risk Specs (14 dossiers)
│   ├── api/                               # REST OpenAPI 3.1, WSS, gRPC, and FIX 5.0 SP2 Specs
│   ├── compliance/                        # SEBI, IFSCA, RBI, FIU-IND & DPDP Regulatory Filings
│   ├── adr/                               # 46 Architecture Decision Records (ADR-0001 to 0046)
│   ├── runbooks/                          # 34 SRE Operational Runbooks (RUNBOOK-01 to 34)
│   └── vision/                            # Sovereign Infrastructure Strategy Dossiers
│
├── services/                              # 79 Polyglot Microservices (Empty Scaffolding)
│   ├── user-service/                      # User & Identity Service (FastAPI / PostgreSQL)
│   ├── kyc-aml-service/                   # KYC & Sanctions Screening Service (FastAPI)
│   ├── wallet-account-service/            # Double-Entry Wallet Ledger Service (Go)
│   ├── order-service/                     # Spot Order Management & Lifecycle Service (Go)
│   ├── order-matching-engine/             # Sub-15us In-Memory Matching Engine (Rust)
│   ├── risk-margin-service/               # Pre-Trade Risk & Collateral Validation (Go)
│   ├── market-data-service/               # Real-Time Level-2/3 Market Data Service (Go)
│   ├── trade-settlement-service/          # Atomic DvP Settlement Coordinator (Go)
│   ├── portfolio-holdings-service/        # Fractional Portfolio Accounting Service (Python)
│   ├── fee-engine/                        # Universal fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Engine (Rust / Go)
│   ├── notification-service/              # Transactional Push / SMS / Email Service (Go)
│   ├── payment-gateway-service/           # Banking UPI 2.0 / IMPS / RTGS Gateway (Go)
│   ├── custodian-depository-adapter/      # NSDL / CDSL API Adapter Service (Go)
│   ├── gift-city-funding-service/         # GIFT City IFSCA Multi-Currency FX Gateway (Python)
│   ├── reconciliation-service/            # On-Chain vs Off-Chain Ledger Reconciler (Go)
│   ├── regulatory-reporting-service/      # SEBI / RBI Statutory Reporting Service (Python)
│   ├── admin-back-office-service/         # Dual-Control Maker-Checker Admin Service (Go)
│   ├── audit-log-service/                 # Cryptographic Immutable Audit Trail Service (Rust)
│   ├── api-gateway/                       # Unified API Gateway & BFF Layer (Go / Envoy)
│   ├── rate-limiting-service/             # Rate Limiting & Abuse Prevention Service (Redis)
│   ├── search-discovery-service/          # Catalog & Instrument Search Service (OpenSearch)
│   ├── corporate-actions-service/         # Dividend, Split & CIL Action Engine (Python)
│   ├── tax-reporting-service/             # STT & Capital Gains Tax Statement Service (Go)
│   ├── referral-growth-service/           # Multi-Tier Growth & Rebate Service (Go)
│   ├── fix-protocol-gateway/              # FIX 5.0 SP2 Binary Institutional Gateway (Rust)
│   ├── advanced-order-types-engine/       # Algorithmic Trailing & Iceberg Slicing Engine (Go)
│   ├── institutional-dark-pool/           # 24/7 Block Crossing & Dark Pool Service (Rust)
│   ├── market-surveillance-engine/        # Anti-Spoofing & Wash Trading Surveillance (Go)
│   ├── var-margin-engine/                 # Real-Time VaR & Extreme Loss Margin Engine (Rust)
│   ├── sgf-default-waterfall-service/     # Core Settlement Guarantee Fund Service (Go)
│   ├── auction-delivery-failure-resolver/ # Fail-to-Deliver Demat Auction Engine (Rust)
│   ├── cbdc-digital-rupee-adapter/        # 24/7 RBI eINR CBDC Settlement Adapter (Go)
│   ├── continuous-regulatory-dispatcher/  # Continuous 24/7 Regulatory Audit Dispatcher (Go)
│   ├── btc-lightning-taproot-ingress/     # Bitcoin UTXO, Lightning & Taproot Ingress (Go)
│   ├── evm-ccip-ingress/                  # Chainlink CCIP Multi-Token Bridge Service (Go)
│   ├── solana-wormhole-ingress/           # Solana SPL & Wormhole Ingress Service (Rust)
│   ├── mpc-tss-custody-service/           # Institutional MPC-TSS Cold Vault Custody (Rust)
│   ├── crosschain-collateral-router/      # Cross-Chain Collateral & Synthetic FX Router (Go)
│   ├── options-pricing-engine/            # Black-Scholes SIMD Greeks Engine (C++ / Rust)
│   ├── perpetuals-engine/                 # Perpetuals & Synthetic Derivatives Engine (Rust)
│   ├── span-margin-engine/                # SPAN Portfolio Multi-Asset Margin Engine (Go)
│   ├── nse-bse-adapter/                   # Primary Market DMA & Bhavcopy Adapter (Go)
│   ├── mcx-commodity-adapter/             # MCX Physical Commodity & WDRA eNWR Adapter (Go)
│   ├── nbse-fee-distribution-engine/      # Fee Splitter (0.00% fee launch policy Waterfall) (Go / Rust)
│   ├── settlement-relayer-gas-escalator/  # 32-Way Partitioned Relayer Gas Escalator (Go)
│   ├── matching-engine-wal-failover/      # Zero-Copy mmap WAL & Shadow Failover (Rust)
│   ├── crosschain-reorg-saga-coordinator/ # 5-Phase Distributed Clawback Saga Engine (Go)
│   ├── speed-bump-guard/                  # 500us Asymmetric Speed Bump Guard (Go)
│   ├── corporate-action-ebce-ledger/      # EBCE Sub-Ledger & Demat Lag Reconciler (Go)
│   ├── fpi-sectoral-cap-engine/           # Real-Time FEMA & FPI Investment Limit Guard (Go)
│   ├── fx-haircut-calculator/             # Dynamic Multi-Currency FX Haircut Engine (Go)
│   ├── gemini-conversational-advisor/     # Google Gemini RAG Advisory Plane (Python)
│   ├── after-hours-liquidity-gateway/     # 24/7 Off-Hours Secondary Liquidity Gateway (Go)
│   ├── bond-yield-calculator/             # Tokenized G-Sec & Corporate Bond Yield Engine (Go)
│   ├── cross-asset-collateral-optimizer/  # SPAN Collateral Optimization Service (Go)
│   ├── luld-circuit-breaker-engine/       # Limit-Up / Limit-Down Volatility Dampener (Go)
│   ├── institutional-algo-execution-engine/# Institutional TWAP / VWAP / POV Slicing (Go)
│   ├── nbbo-consolidated-tape-engine/     # Sub-Millisecond NBBO Consolidated Tape (Go)
│   ├── erc4337-bundler-paymaster-service/ # Account Abstraction Bundler & Paymaster (Go)
│   ├── clearing-corp-interop-gateway/     # SEBI Clearing Corporation Interoperability (Go)
│   ├── tick-market-replay-service/        # High-Frequency Tick Replay & Audit Daemon (Rust)
│   ├── automated-liquidity-provisioning/  # Automated Liquidity Provisioning Engine (Go)
│   ├── sebi-margin-pledge-gateway/        # SEBI Margin Pledge / Re-Pledge Gateway (Go)
│   ├── collateral-haircut-auto-topup/     # Dynamic Haircut & Margin Call Engine (Go)
│   ├── copy-trading-replication-service/  # Low-Latency Proportional Mirror Engine (Rust)
│   ├── rwa-launchpad-dutch-auction-engine/# RWA Primary Launchpad & Dutch Auction Engine (Go)
│   ├── sebi-scores-grievance-gateway/     # SEBI SCORES 2.0 & Ombudsman Gateway (Python)
│   ├── auto-deleveraging-margin-default/  # Real-Time Auto-Deleveraging (ADL) Engine (Rust)
│   ├── p2p-fiat-escrow-service/           # P2P Fiat Escrow & Multi-Sig Dispute Service (Go)
│   ├── rfq-instant-convert-swap-service/  # Request-For-Quote Instant Convert Swap Engine (Go)
│   ├── affiliate-referral-rebate-engine/  # Multi-Tier Referral & Fee Rebate Engine (Go)
│   ├── btc-usdt-market-data-feeder/       # Live Redundant WebSocket Market Feeder (Go)
│   ├── demo-matching-execution-engine/    # Demo Paper Trading Virtual Matching Engine (Rust)
│   ├── demo-trading-faucet-wallet-service/# 10,000 USDT Testnet Faucet & Virtual Wallet (Go)
│   ├── btc-usdt-spot-order-service/       # Real-Money BTC/USDT Spot Order Execution (Go)
│   ├── btc-usdt-depth-broadcaster/        # Level-2/3 Depth Broadcaster & Ring Buffer (Go)
│   ├── btc-utxo-deposit-listener/         # Bitcoin Native SegWit & Taproot Listener (Go)
│   ├── usdt-multichain-deposit-service/   # Multi-Chain USDT Deposit & Custody Ingress (Go)
│   └── crypto-withdrawal-clearing-engine/ # Real-Money Crypto Withdrawal Clearing Engine (Go)
│
├── contracts/                             # 50 Smart Contract Domains (Empty Scaffolding)
│   └── src/
│       ├── compliance/                    # IdentityRegistry.sol, ClaimTopicsRegistry.sol
│       ├── tokens/                        # EquityToken.sol (ERC-3643), weINR.sol, wBTC.sol
│       ├── settlement/                    # SettlementDvP.sol, BtcUsdtDvPSettlement.sol
│       ├── governance/                    # MultiSigGovernance.sol, TimelockController.sol
│       ├── reserves/                      # ProofOfReserve.sol, MerkleRegistry.sol
│       ├── bridge/                        # CrossChainBridge.sol, HTLCSwap.sol
│       ├── derivatives/                   # PerpetualsMarginVault.sol, OptionsClearing.sol
│       ├── yield/                         # BondYieldVault.sol (ERC-4626), StakingVault.sol
│       ├── launchpad/                     # DutchAuction.sol, LinearVestingVault.sol
│       ├── demo/                          # VirtualFaucet.sol, VirtualTokens.sol (Testnet)
│       ├── custody/                       # MpcVaultCoordinator.sol, ColdVaultSweep.sol
│       └── interfaces/                    # IDvP.sol, IERC3643.sol, IVault.sol
│
├── apps/                                  # Multi-Platform Client Applications (Empty Scaffolding)
│   ├── growww_flutter/                    # Flutter 3.22+ Multi-Platform Client (Android, iOS, Desktop)
│   │   ├── lib/screens/trading/           # Spot Trading, TradingView Chart, Order Entry Sheet
│   │   ├── lib/screens/wallet/            # Deposit, Withdrawal, Balances, Network Selector
│   │   ├── lib/screens/demo/              # Paper Trading Mode Switcher, Faucet Modal
│   │   ├── lib/screens/auth/              # Biometric Face ID, FIDO2 Passkeys, PIN Login
│   │   ├── lib/screens/portfolio/         # Holdings, FIFO Realized PnL, Capital Gains
│   │   ├── lib/screens/settings/          # Security Center, Dark Mode Toggles, DPDP Rights
│   │   ├── lib/controllers/               # Riverpod State Notifiers & Providers
│   │   ├── lib/models/                    # Type-Safe Domain Entities & JSON Decoders
│   │   └── lib/widgets/                   # Tabular Numbers, High-Contrast Pill Toggles
│   ├── growww_web/                        # Next.js 14 Pro Trading Web Terminal
│   │   ├── src/app/trade/btc-usdt/        # Institutional 3-Column Cockpit
│   │   ├── src/app/demo/                  # Web Paper Trading Simulator & PnL Analytics
│   │   ├── src/components/tradingview/    # Fluid Canvas & Charting Widgets
│   │   └── src/components/orderbook/      # Streaming Level-2 DOM Ladder
│   ├── growww_admin/                      # Next.js 14 Operator & Back-Office Portal
│   │   ├── src/app/compliance/            # CKYC Verification & Sanctions Review Queue
│   │   ├── src/app/p2p-disputes/          # P2P Arbitration Maker-Checker Console
│   │   ├── src/app/reserves/              # Proof-of-Reserve Attestation Reconciliation
│   │   └── src/app/grievance/             # SEBI SCORES 2.0 & Ombudsman SLA Dashboard
│   └── growww_marketing/                  # Next.js 14 Public Education & Landing Site
│
├── packages/                              # Shared Polyglot Libraries (Empty Scaffolding)
│   ├── proto/                             # Protocol Buffers v3 Service & Event Schemas
│   ├── domain_types/                      # Shared Rust Domain Types & Micro-Currency Primitives
│   ├── go_common/                         # Shared Go Middleware, Config & HTTP Helpers
│   ├── growww_ui/                         # Flutter Theme Design System & Tokens
│   ├── web_ui/                            # React / Tailwind Component Library Primitives
│   └── crypto_utils/                      # HSM, EIP-712 & Poseidon ZK Primitives
│
└── infra/                                 # Infrastructure as Code (Empty Scaffolding)
    ├── blockchain/besu/                   # Hyperledger Besu QBFT Genesis & Validator Manifests
    ├── k8s/environments/testnet/          # Kubernetes Manifests for Testnet Sandbox (13371)
    ├── k8s/environments/mainnet/          # Kubernetes Manifests for Mainnet Production (2026)
    ├── local/docker-compose/              # Local Development Multi-Container Topologies
    ├── observability/                     # Prometheus, OpenTelemetry & Grafana Dashboards
    └── terraform/                         # AWS ap-south-1 & GIFT City Datacenter Provisioning
```
