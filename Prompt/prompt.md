# Master Build Prompt Specification Index

This document is the authoritative catalog of all declarative build prompt specifications for the National Blockchain Stock Exchange (NBSE) platform (working name: Growww / NBSE). Each prompt sheet defines the system boundaries, architecture, state transitions, mathematical formulas, communication protocols, failure modes, and automated acceptance tests for a specific platform component.

Total Prompt Specifications: 288

---

## 0. High-Level Vision, Domain & Product Specifications (11 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `000` | [`000_project_north_star.md`](./prompts/000_project_north_star.md) | Project North Star: Purpose, Boundaries & Architectural Vision |
| `001` | [`001_glossary_of_domain_terms.md`](./prompts/001_glossary_of_domain_terms.md) | Glossary of Domain Terms & Ubiquitous Language |
| `002` | [`002_two_entity_legal_technical_structure.md`](./prompts/002_two_entity_legal_technical_structure.md) | Two-Entity Legal & Technical Separation Architecture |
| `003` | [`003_regulatory_pathway_overview.md`](./prompts/003_regulatory_pathway_overview.md) | Regulatory Pathway & Sandbox Strategy (SEBI, RBI, IFSCA) |
| `004` | [`004_kyc_aml_policy_domestic.md`](./prompts/004_kyc_aml_policy_domestic.md) | Domestic KYC/AML & Prevention of Money Laundering Policy |
| `005` | [`005_kyc_aml_policy_foreign_gift_city.md`](./prompts/005_kyc_aml_policy_foreign_gift_city.md) | International Investor KYC/AML Policy (GIFT City IFSCA) |
| `006` | [`006_fee_model_specification.md`](./prompts/006_fee_model_specification.md) | Fee Model Specification: Fixed Fee Structure, Realized Gain Engine & Treasury Allocation |
| `007` | [`007_proof_of_reserve_public_disclosure.md`](./prompts/007_proof_of_reserve_public_disclosure.md) | Proof-of-Reserve & Custody Verification Architecture |
| `008` | [`008_data_protection_and_privacy_policy.md`](./prompts/008_data_protection_and_privacy_policy.md) | Data Protection, Privacy & Zero-PII Ledger Policy (DPDP & GDPR) |
| `009` | [`009_risk_disclosure_and_investor_protection.md`](./prompts/009_risk_disclosure_and_investor_protection.md) | Risk Disclosure & Investor Protection Framework |
| `010` | [`010_non_functional_requirements_master.md`](./prompts/010_non_functional_requirements_master.md) | Non-Functional Requirements & Performance Engineering Master Spec |

---

## 1. System Architecture & Foundation Infrastructure (14 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `101` | [`101_system_architecture_overview.md`](./prompts/101_system_architecture_overview.md) | System Architecture Overview & C4 Model |
| `102` | [`102_service_boundary_map_and_bounded_contexts.md`](./prompts/102_service_boundary_map_and_bounded_contexts.md) | Service Boundary Map & Domain-Driven Design (DDD) Bounded Contexts |
| `103` | [`103_api_design_standards.md`](./prompts/103_api_design_standards.md) | API Design Standards (REST & gRPC Conventions, Versioning & Error Formats) |
| `104` | [`104_event_schema_and_kafka_topic_standards.md`](./prompts/104_event_schema_and_kafka_topic_standards.md) | Event Schema & Kafka Topic Naming Standards (CloudEvents, Schema Registry & Partitioning) |
| `105` | [`105_authentication_and_authorization_architecture.md`](./prompts/105_authentication_and_authorization_architecture.md) | Authentication & Authorization Architecture (OAuth2/OIDC, RBAC/ABAC & mTLS Service Mesh) |
| `106` | [`106_monorepo_vs_polyrepo_and_layout.md`](./prompts/106_monorepo_vs_polyrepo_and_layout.md) | Monorepo Architecture, Repository Layout & Tooling Strategy |
| `107` | [`107_coding_standards_and_linting.md`](./prompts/107_coding_standards_and_linting.md) | Polyglot Coding Standards, Code Formatting & Linting Suite |
| `108` | [`108_environment_strategy_and_config_mgmt.md`](./prompts/108_environment_strategy_and_config_mgmt.md) | Environment Strategy, Configuration Management & 12-Factor App Architecture |
| `109` | [`109_secrets_management_architecture.md`](./prompts/109_secrets_management_architecture.md) | Secrets Management, Encryption Keys & HSM Architecture |
| `110` | [`110_inter_entity_secure_communication.md`](./prompts/110_inter_entity_secure_communication.md) | Inter-Entity Secure Communication Design (Domestic Regulated Entity <-> GIFT City Gateway) |
| `111` | [`111_domain_model_core_entities.md`](./prompts/111_domain_model_core_entities.md) | Canonical Domain Model & Core Entity Specifications |
| `112` | [`112_idempotency_and_exactly_once_processing.md`](./prompts/112_idempotency_and_exactly_once_processing.md) | Idempotency & Exactly-Once Processing Strategy Across Distributed Services |
| `113` | [`113_distributed_tracing_and_opentelemetry_standard.md`](./prompts/113_distributed_tracing_and_opentelemetry_standard.md) | Distributed Tracing & OpenTelemetry Microsecond Precision Standard |
| `114` | [`114_schema_evolution_and_backward_compatibility_protocol.md`](./prompts/114_schema_evolution_and_backward_compatibility_protocol.md) | Schema Evolution & Wire Protocol Backward Compatibility Standard |

---

## 2. Backend Microservices (Go / Rust / Python) (79 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `201` | [`201_user_service.md`](./prompts/201_user_service.md) | User & Identity Service (FastAPI / PostgreSQL) |
| `202` | [`202_kyc_aml_service.md`](./prompts/202_kyc_aml_service.md) | KYC & Sanctions Screening Service (FastAPI / Celery) |
| `203` | [`203_wallet_account_service.md`](./prompts/203_wallet_account_service.md) | Wallet & Double-Entry Account Ledger Service (Go) |
| `204` | [`204_order_service.md`](./prompts/204_order_service.md) | Order Management & Lifecycle Service (Go) |
| `205` | [`205_order_matching_engine.md`](./prompts/205_order_matching_engine.md) | High-Performance Order Matching Engine (Rust) |
| `206` | [`206_risk_and_margin_checks_service.md`](./prompts/206_risk_and_margin_checks_service.md) | Pre-Trade Risk & Margin Engine (Go / Redis) |
| `207` | [`207_market_data_service.md`](./prompts/207_market_data_service.md) | Real-Time Market Data & WebSocket Streaming Service (Go) |
| `208` | [`208_trade_settlement_service.md`](./prompts/208_trade_settlement_service.md) | Trade Settlement & DvP Orchestration Service (Go) |
| `209` | [`209_portfolio_and_holdings_service.md`](./prompts/209_portfolio_and_holdings_service.md) | Portfolio & Fractional Holdings Accounting Service (Python / FastAPI) |
| `210` | [`210_fee_and_realized_pnl_engine.md`](./prompts/210_fee_and_realized_pnl_engine.md) | Fixed Transaction Fee & Realized Capital Gain Engine (Rust / Go) |
| `211` | [`211_notification_service.md`](./prompts/211_notification_service.md) | Transactional Notification Service (Go) |
| `212` | [`212_payment_gateway_integration_service.md`](./prompts/212_payment_gateway_integration_service.md) | Banking & Payment Gateway Integration Service (Go) |
| `213` | [`213_custodian_depository_integration_service.md`](./prompts/213_custodian_depository_integration_service.md) | Custodian & Depository Integration Service (NSDL/CDSL API Adapter) |
| `214` | [`214_foreign_investor_funding_and_fx_service.md`](./prompts/214_foreign_investor_funding_and_fx_service.md) | Foreign-Investor Funding & FX Service (GIFT City / IFSCA Gateway) |
| `215` | [`215_reconciliation_service.md`](./prompts/215_reconciliation_service.md) | On-Chain vs Off-Chain Ledger Reconciliation Engine (Go / Rust) |
| `216` | [`216_regulatory_reporting_service.md`](./prompts/216_regulatory_reporting_service.md) | Regulatory Reporting & Compliance Service (SEBI / RBI / IFSCA) |
| `217` | [`217_admin_back_office_service.md`](./prompts/217_admin_back_office_service.md) | Admin & Back-Office Service (Exception Handling & Dual-Control Operations) |
| `218` | [`218_audit_log_service.md`](./prompts/218_audit_log_service.md) | Immutable Audit Log Service (High-Throughput Cryptographic Trail) |
| `219` | [`219_api_gateway_and_bff.md`](./prompts/219_api_gateway_and_bff.md) | API Gateway & Backend-for-Frontend (BFF) Layer (Go / Envoy / REST / GraphQL) |
| `220` | [`220_rate_limiting_and_abuse_prevention.md`](./prompts/220_rate_limiting_and_abuse_prevention.md) | Rate Limiting & Abuse Prevention Service (Redis / Token Bucket / Behavioral Scoring) |
| `221` | [`221_search_and_discovery_service.md`](./prompts/221_search_and_discovery_service.md) | Search & Discovery Service (OpenSearch / PostgreSQL / Catalog Engine) |
| `222` | [`222_corporate_actions_service.md`](./prompts/222_corporate_actions_service.md) | Corporate Actions Service (Dividends, Splits, Bonuses & Token Adjustments) |
| `223` | [`223_tax_reporting_statement_service.md`](./prompts/223_tax_reporting_statement_service.md) | Tax Reporting & Capital Gains Statement Service (Income Tax / Sections 111A & 112A) |
| `224` | [`224_referral_growth_service.md`](./prompts/224_referral_growth_service.md) | Referral & Growth Service (SEBI Advertising Code Compliant) |
| `225` | [`225_fix_protocol_gateway.md`](./prompts/225_fix_protocol_gateway.md) | FIX 5.0 SP2 / ITCH / OUCH Low-Latency Binary Trading Gateway (Rust) |
| `226` | [`226_advanced_order_types_engine.md`](./prompts/226_advanced_order_types_engine.md) | Advanced Order Types & Algorithmic Trigger Engine (Go) |
| `227` | [`227_institutional_dark_pool_service.md`](./prompts/227_institutional_dark_pool_service.md) | 24/7 Institutional Dark Pool & Block Crossing Service (Rust) |
| `228` | [`228_real_time_market_surveillance_engine.md`](./prompts/228_real_time_market_surveillance_engine.md) | Real-Time Market Surveillance & Dynamic Volatility Engine (Go) |
| `229` | [`229_real_time_var_margin_engine.md`](./prompts/229_real_time_var_margin_engine.md) | Real-Time Value at Risk (VaR) & Extreme Loss Margin (ELM) Engine (Rust / Redis / SIMD) |
| `230` | [`230_settlement_guarantee_fund_service.md`](./prompts/230_settlement_guarantee_fund_service.md) | Settlement Guarantee Fund & Default Waterfall Service (Go / Temporal / PostgreSQL) |
| `231` | [`231_auction_delivery_failure_resolver.md`](./prompts/231_auction_delivery_failure_resolver.md) | Depository Fail-to-Deliver Auction & Buy-In Resolution Engine (Rust / Kafka / PostgreSQL) |
| `232` | [`232_cbdc_digital_rupee_settlement_adapter.md`](./prompts/232_cbdc_digital_rupee_settlement_adapter.md) | 24/7 e₹ (Digital Rupee CBDC) & RBI RTGS Instant Liquidity Settlement Adapter (Go / mTLS / ISO 20022) |
| `233` | [`233_continuous_24x7_regulatory_reporting.md`](./prompts/233_continuous_24x7_regulatory_reporting.md) | Continuous 24x7 Regulatory Statutory Reporting & Compliance Dispatcher |
| `234` | [`234_bitcoin_lightning_and_taproot_ingress_service.md`](./prompts/234_bitcoin_lightning_and_taproot_ingress_service.md) | Bitcoin (BTC), Lightning Network & Taproot Collateral Ingress Service (Go / Bitcoin Core / LND / PSBT / HSM) |
| `235` | [`235_evm_chainlink_ccip_multi_token_ingress_service.md`](./prompts/235_evm_chainlink_ccip_multi_token_ingress_service.md) | EVM Chainlink CCIP Multi-Token Ingress & Canonical Lockbox Bridge Service (Go / CCIP / EIP-712) |
| `236` | [`236_solana_spl_and_wormhole_ingress_service.md`](./prompts/236_solana_spl_and_wormhole_ingress_service.md) | Solana SPL and Wormhole Ingress Service (Rust / Go / gRPC) |
| `237` | [`237_multichain_mpc_tss_vault_custody_service.md`](./prompts/237_multichain_mpc_tss_vault_custody_service.md) | Institutional Multi-Chain MPC-TSS Vault & Custody Service (Rust / C++ / CloudHSM / GG20 & FROST) |
| `238` | [`238_crosschain_collateral_and_synthetic_fx_router.md`](./prompts/238_crosschain_collateral_and_synthetic_fx_router.md) | Cross-Chain Collateral & Synthetic FX Router Service |
| `239` | [`239_options_pricing_and_greeks_risk_engine.md`](./prompts/239_options_pricing_and_greeks_risk_engine.md) | Real-Time Options Pricing, Greeks & Volatility Risk Engine (C++ / Rust / SIMD) |
| `240` | [`240_perpetuals_and_synthetic_derivatives_engine.md`](./prompts/240_perpetuals_and_synthetic_derivatives_engine.md) | Perpetuals & Synthetic Derivatives Engine (Rust) |
| `241` | [`241_span_portfolio_margin_and_liquidation_engine.md`](./prompts/241_span_portfolio_margin_and_liquidation_engine.md) | Real-Time SPAN Portfolio Margin & Liquidation Engine (Go / Rust / Redis) |
| `242` | [`242_nse_bse_market_data_and_order_routing_adapter.md`](./prompts/242_nse_bse_market_data_and_order_routing_adapter.md) | NSE / BSE Market Data Feeds, Bhavcopy Ingestion & Smart Order Routing (SOR) Adapter (Go) |
| `243` | [`243_mcx_commodity_and_warehouse_receipt_adapter.md`](./prompts/243_mcx_commodity_and_warehouse_receipt_adapter.md) | MCX Commodity & Warehouse Receipt Adapter (Go / Python / WDRA eNWR / NERL / CCRL) |
| `244` | [`244_nbse_fixed_fee_and_revenue_distribution_engine.md`](./prompts/244_nbse_fixed_fee_and_revenue_distribution_engine.md) | NBSE Fixed Fee & Revenue Distribution Engine (Go / Rust) |
| `245` | [`245_settlement_relayer_nonce_partitioning_and_gas_escalator.md`](./prompts/245_settlement_relayer_nonce_partitioning_and_gas_escalator.md) | Settlement Relayer Nonce Partitioning & Gas Escalator (Go) |
| `246` | [`246_matching_engine_mmap_wal_and_shadow_failover.md`](./prompts/246_matching_engine_mmap_wal_and_shadow_failover.md) | Matching Engine Memory-Mapped WAL & Hot-Warm Shadow Failover (Rust) |
| `247` | [`247_crosschain_reorg_invalidation_and_clawback_saga_coordinator.md`](./prompts/247_crosschain_reorg_invalidation_and_clawback_saga_coordinator.md) | Cross-Chain Reorg Invalidation and Clawback Saga Coordinator (Go / Rust) |
| `248` | [`248_exchange_adapter_preopen_auction_and_speed_bump_guard.md`](./prompts/248_exchange_adapter_preopen_auction_and_speed_bump_guard.md) | Exchange Adapter Pre-Open Auction & Asymmetric Speed Bump Guard (Go) |
| `249` | [`249_corporate_action_exdate_order_purge_and_ebce_ledger.md`](./prompts/249_corporate_action_exdate_order_purge_and_ebce_ledger.md) | Corporate Action Ex-Date Order Purge & EBCE Ledger (Go / Rust) |
| `250` | [`250_fpi_cross_border_sectoral_cap_and_clubbing_engine.md`](./prompts/250_fpi_cross_border_sectoral_cap_and_clubbing_engine.md) | Foreign Portfolio Investor (FPI) Real-Time Sectoral Cap & Clubbing Engine (Go / Redis / PostgreSQL) |
| `251` | [`251_cross_currency_collateral_fx_haircut_and_hedging_engine.md`](./prompts/251_cross_currency_collateral_fx_haircut_and_hedging_engine.md) | Cross-Currency Collateral Dynamic FX Haircut & Auto-Hedging Engine (Rust / Redis / PostgreSQL) |
| `252` | [`252_gemini_financial_intelligence_and_literacy_service.md`](./prompts/252_gemini_financial_intelligence_and_literacy_service.md) | Gemini Financial Intelligence & Literacy Service (FastAPI / LangChain / RAG / Hyperledger Besu) |
| `253` | [`253_continuous_24_7_synthetic_market_and_after_hours_gateway.md`](./prompts/253_continuous_24_7_synthetic_market_and_after_hours_gateway.md) | Continuous 24/7 Synthetic Market & After-Hours Liquidity Gateway (Rust / Redis / Besu) |
| `254` | [`254_bond_yield_curve_and_dirty_price_calculator_service.md`](./prompts/254_bond_yield_curve_and_dirty_price_calculator_service.md) | Bond Yield Curve & Dirty Price Calculator Service (Rust / gRPC / SIMD) |
| `255` | [`255_cross_asset_portfolio_margin_and_collateral_optimizer.md`](./prompts/255_cross_asset_portfolio_margin_and_collateral_optimizer.md) | Unified Cross-Asset Portfolio Margining & Dynamic Collateral Optimization Service (Go / Rust) |
| `256` | [`256_limit_up_limit_down_luld_volatility_dampener_service.md`](./prompts/256_limit_up_limit_down_luld_volatility_dampener_service.md) | Limit-Up / Limit-Down (LULD) Dynamic Volatility Dampener & Call Auction Engine (Rust) |
| `257` | [`257_twap_vwap_algorithmic_execution_engine.md`](./prompts/257_twap_vwap_algorithmic_execution_engine.md) | Institutional Algorithmic Execution Engine (TWAP, VWAP, POV & Iceberg) (Rust) |
| `258` | [`258_national_best_bid_offer_nbbo_consolidated_tape_engine.md`](./prompts/258_national_best_bid_offer_nbbo_consolidated_tape_engine.md) | Real-Time National Best Bid and Offer (NBBO) Consolidated Tape Engine (Go / Rust) |
| `259` | [`259_account_abstraction_bundler_and_paymaster_service.md`](./prompts/259_account_abstraction_bundler_and_paymaster_service.md) | ERC-4337 Account Abstraction Bundler & Gasless Paymaster Service (Go) |
| `260` | [`260_clearing_corporation_interoperability_and_margin_pledge_service.md`](./prompts/260_clearing_corporation_interoperability_and_margin_pledge_service.md) | Clearing Corporation Interoperability & SEBI Margin Pledge/Re-Pledge Gateway (Go) |
| `261` | [`261_high_frequency_tick_by_tick_market_replay_service.md`](./prompts/261_high_frequency_tick_by_tick_market_replay_service.md) | High-Frequency Tick-by-Tick Market Replay & Audit Service (Rust / Go) |
| `262` | [`262_automated_liquidity_provisioning_and_market_maker_incentive_engine.md`](./prompts/262_automated_liquidity_provisioning_and_market_maker_incentive_engine.md) | Automated Liquidity Provisioning & Market Maker Incentive Engine (Go) |
| `263` | [`263_sebi_margin_pledge_repledge_depository_gateway.md`](./prompts/263_sebi_margin_pledge_repledge_depository_gateway.md) | SEBI Margin Pledge & Re-Pledge Depository Gateway (Go) |
| `264` | [`264_dynamic_collateral_haircut_and_auto_topup_engine.md`](./prompts/264_dynamic_collateral_haircut_and_auto_topup_engine.md) | Dynamic Collateral Haircut & Margin Call Notification Engine (Go / Rust) |
| `265` | [`265_copy_trading_and_proportional_replication_service.md`](./prompts/265_copy_trading_and_proportional_replication_service.md) | Copy Trading & Proportional Replication Service (Rust / Go) |
| `266` | [`266_rwa_launchpad_and_dutch_auction_engine.md`](./prompts/266_rwa_launchpad_and_dutch_auction_engine.md) | RWA Primary Token Launchpad & Dutch Auction Engine (Go) |
| `267` | [`267_sebi_scores_and_regulatory_grievance_gateway.md`](./prompts/267_sebi_scores_and_regulatory_grievance_gateway.md) | SEBI SCORES 2.0 & Regulatory Grievance Gateway (Python / FastAPI) |
| `268` | [`268_auto_deleveraging_and_margin_default_resolver.md`](./prompts/268_auto_deleveraging_and_margin_default_resolver.md) | Real-Time Auto-Deleveraging (ADL) & Margin Default Resolver (Rust) |
| `269` | [`269_p2p_fiat_escrow_and_dispute_arbitration_service.md`](./prompts/269_p2p_fiat_escrow_and_dispute_arbitration_service.md) | P2P Fiat Escrow & Multi-Sig Dispute Arbitration Service (Go) |
| `270` | [`270_rfq_and_instant_convert_swap_service.md`](./prompts/270_rfq_and_instant_convert_swap_service.md) | Request-For-Quote (RFQ) & Instant Convert Swap Service (Go / Rust) |
| `271` | [`271_affiliate_referral_and_multi_tier_rebate_engine.md`](./prompts/271_affiliate_referral_and_multi_tier_rebate_engine.md) | Multi-Tier Affiliate Referral & Rebate Engine (Go) |
| `272` | [`272_btc_usdt_live_external_market_data_feeder.md`](./prompts/272_btc_usdt_live_external_market_data_feeder.md) | BTC/USDT Real-Time External Market Data Feeder & Aggregator (Go) |
| `273` | [`273_demo_trading_virtual_matching_and_execution_engine.md`](./prompts/273_demo_trading_virtual_matching_and_execution_engine.md) | Demo / Paper Trading Virtual Matching & Execution Engine (Rust / Go) |
| `274` | [`274_demo_trading_faucet_and_virtual_wallet_ledger.md`](./prompts/274_demo_trading_faucet_and_virtual_wallet_ledger.md) | Demo Trading Faucet & Virtual Balance Ledger Service (Go) |
| `275` | [`275_btc_usdt_spot_order_execution_and_lifecycle_service.md`](./prompts/275_btc_usdt_spot_order_execution_and_lifecycle_service.md) | Real-Money BTC/USDT Spot Order Execution & Lifecycle Service (Go) |
| `276` | [`276_btc_usdt_live_orderbook_and_market_depth_broadcaster.md`](./prompts/276_btc_usdt_live_orderbook_and_market_depth_broadcaster.md) | BTC/USDT Real-Time Order Book & Market Depth Broadcaster (Go / Rust) |
| `277` | [`277_btc_deposit_and_onchain_utxo_confirmation_listener.md`](./prompts/277_btc_deposit_and_onchain_utxo_confirmation_listener.md) | Bitcoin (BTC) On-Chain UTXO Deposit & Confirmation Listener (Go) |
| `278` | [`278_usdt_multichain_deposit_and_burn_mint_custody_service.md`](./prompts/278_usdt_multichain_deposit_and_burn_mint_custody_service.md) | USDT Multi-Chain Deposit, Verification & Custody Ingress Service (Go) |
| `279` | [`279_btc_usdt_real_money_withdrawal_and_risk_clearing_engine.md`](./prompts/279_btc_usdt_real_money_withdrawal_and_risk_clearing_engine.md) | BTC & USDT Real-Money Withdrawal & Multi-Sig Risk Clearing Engine (Go) |

---

## 3. Blockchain & Smart Contracts (Hyperledger Besu / Solidity / EVM) (50 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `301` | [`301_permissioned_blockchain_evaluation_selection.md`](./prompts/301_permissioned_blockchain_evaluation_selection.md) | Permissioned Blockchain Platform Evaluation & Selection (Hyperledger Besu vs Fabric vs Polygon Supernets) |
| `302` | [`302_network_topology_and_validator_setup.md`](./prompts/302_network_topology_and_validator_setup.md) | Permissioned Network Topology & QBFT Validator Infrastructure Setup |
| `303` | [`303_token_issuance_smart_contract.md`](./prompts/303_token_issuance_smart_contract.md) | Permissioned Asset Token Issuance Smart Contract (ERC-3643 / CMTA) |
| `304` | [`304_token_redemption_smart_contract.md`](./prompts/304_token_redemption_smart_contract.md) | Token Redemption & Custodial Asset Release Smart Contract |
| `305` | [`305_transfer_compliance_hooks_smart_contract.md`](./prompts/305_transfer_compliance_hooks_smart_contract.md) | Smart Contract Transfer Compliance Hooks & KYC/AML Whitelist Registry |
| `306` | [`306_settlement_dvp_smart_contract.md`](./prompts/306_settlement_dvp_smart_contract.md) | Atomic Delivery-versus-Payment (DvP) Settlement Smart Contract (SettlementDvP.sol) |
| `307` | [`307_multisig_governance_smart_contract.md`](./prompts/307_multisig_governance_smart_contract.md) | Multi-Party Authorization & Multisig Governance Smart Contract |
| `308` | [`308_on_chain_proof_of_reserve_publishing.md`](./prompts/308_on_chain_proof_of_reserve_publishing.md) | On-Chain Proof-of-Reserve Attestation & Merkle Registry Smart Contract |
| `309` | [`309_event_indexing_service.md`](./prompts/309_event_indexing_service.md) | Blockchain Event Indexing & Query Service |
| `310` | [`310_chain_node_monitoring_and_alerting.md`](./prompts/310_chain_node_monitoring_and_alerting.md) | Permissioned Blockchain Node Observability & Health Monitoring |
| `311` | [`311_validator_key_management_hsm.md`](./prompts/311_validator_key_management_hsm.md) | Validator & Relayer Key Management via Hardware Security Modules (HSM/KMS) |
| `312` | [`312_chain_upgrade_and_governance.md`](./prompts/312_chain_upgrade_and_governance.md) | Smart Contract Proxy Upgrades & Blockchain Protocol Hardfork Management |
| `313` | [`313_cross_entity_ledger_bridge.md`](./prompts/313_cross_entity_ledger_bridge.md) | Inter-Entity Ledger Bridge (Domestic Depository <-> GIFT City Gateway) |
| `314` | [`314_chain_disaster_recovery_backup.md`](./prompts/314_chain_disaster_recovery_backup.md) | Blockchain Ledger Disaster Recovery, RocksDB Snapshotting & Node Failover |
| `315` | [`315_settlement_guarantee_fund_contract.md`](./prompts/315_settlement_guarantee_fund_contract.md) | On-Chain Settlement Guarantee Fund & Default Waterfall Smart Contract (SettlementGuaranteeFund.sol) |
| `316` | [`316_appchain_rollup_sequencer_architecture.md`](./prompts/316_appchain_rollup_sequencer_architecture.md) | Dedicated Layer-2 Appchain Rollup Sequencer Architecture (50,000 TPS 24/7) |
| `317` | [`317_zk_proof_of_solvency_verifier.md`](./prompts/317_zk_proof_of_solvency_verifier.md) | Zero-Knowledge (ZK) Proof-of-Solvency & Privacy-Preserving Compliance Verifier (Groth16 / Plonk) |
| `318` | [`318_shareholder_voting_governance_contract.md`](./prompts/318_shareholder_voting_governance_contract.md) | Shareholder Proxy Voting & Corporate Governance Smart Contract Suite (ERC-1155 / ERC-5805) |
| `319` | [`319_institutional_custody_bridge.md`](./prompts/319_institutional_custody_bridge.md) | Cross-Chain Institutional Custody Bridge & Interoperability Hub (Chainlink CCIP / LayerZero v2) |
| `320` | [`320_permissioned_testnet_cluster_and_faucet.md`](./prompts/320_permissioned_testnet_cluster_and_faucet.md) | Permissioned Testnet Cluster, Developer Faucet & Mock Depository Minting Gateway |
| `321` | [`321_mainnet_genesis_ceremony_and_validator_onboarding.md`](./prompts/321_mainnet_genesis_ceremony_and_validator_onboarding.md) | Mainnet Genesis Ceremony, FIPS 140-2 Level 3 HSM Key Generation & Multi-Institutional Validator Onboarding |
| `322` | [`322_bitcoin_spv_and_dlc_bridge_contract.md`](./prompts/322_bitcoin_spv_and_dlc_bridge_contract.md) | Bitcoin SPV Header Relay & Discreet Log Contracts (DLC) Settlement Bridge (BitcoinSPVBridge.sol, DLCRegistry.sol) |
| `323` | [`323_evm_crosschain_liquidity_bridge_contract.md`](./prompts/323_evm_crosschain_liquidity_bridge_contract.md) | EVM Cross-Chain Liquidity Bridge Contract (Chainlink CCIP / LayerZero v2 / Multi-Token Collateral) |
| `324` | [`324_solana_besu_light_client_and_bridge_contract.md`](./prompts/324_solana_besu_light_client_and_bridge_contract.md) | Solana-Hyperledger Besu Light Client and Cross-Chain Bridge Contract |
| `325` | [`325_options_clearing_and_exercise_smart_contract.md`](./prompts/325_options_clearing_and_exercise_smart_contract.md) | Options Clearing, Collateral Lockup & Automated In-The-Money Exercise Smart Contracts (OptionsTokenFactory.sol, OptionsClearingHouse.sol, PhysicalAndCashSettler.sol) |
| `326` | [`326_perpetual_futures_clearing_smart_contract.md`](./prompts/326_perpetual_futures_clearing_smart_contract.md) | On-Chain Perpetual Futures Clearing & Margin Smart Contract Suite (PerpClearingHouse.sol, PerpVault.sol, FundingRateOracle.sol) |
| `327` | [`327_multichain_proof_of_reserve_registry_contract.md`](./prompts/327_multichain_proof_of_reserve_registry_contract.md) | Multi-Chain Proof-of-Reserve Registry Smart Contract |
| `328` | [`328_decentralized_oracle_aggregation_smart_contract.md`](./prompts/328_decentralized_oracle_aggregation_smart_contract.md) | Decentralized Multi-Source Oracle Aggregator Smart Contract Suite (OracleAggregator.sol) |
| `329` | [`329_nbse_settlement_dvp_and_fee_collector.md`](./prompts/329_nbse_settlement_dvp_and_fee_collector.md) | NBSE Delivery-versus-Payment (DvP) Settlement & Automated Fee Collector Smart Contracts (NBSESettlementDvP.sol, NBSEFeeCollector.sol) |
| `330` | [`330_mcx_commodity_token_and_vault_registry.md`](./prompts/330_mcx_commodity_token_and_vault_registry.md) | MCX Commodity Token and WDRA Physical Vault Registry Smart Contracts |
| `331` | [`331_primary_market_order_routing_and_clearing_bridge.md`](./prompts/331_primary_market_order_routing_and_clearing_bridge.md) | Primary Market Order Routing & Clearing Bridge Smart Contracts (PrimaryMarketBridge.sol, MarketSessionQueue.sol) |
| `332` | [`332_commodity_scrap_equalizer_and_in_transit_escrow_contract.md`](./prompts/332_commodity_scrap_equalizer_and_in_transit_escrow_contract.md) | Commodity Scrap Equalizer and In-Transit Escrow Smart Contracts |
| `333` | [`333_tokenized_gsec_bonds_and_coupon_accrual_contract.md`](./prompts/333_tokenized_gsec_bonds_and_coupon_accrual_contract.md) | Tokenized G-Sec Bonds & Coupon Accrual Smart Contracts (Solidity / Besu) |
| `334` | [`334_rbi_cbdc_einr_wholesale_and_retail_bridge_contract.md`](./prompts/334_rbi_cbdc_einr_wholesale_and_retail_bridge_contract.md) | RBI CBDC eINR Wholesale and Retail Programmable Bridge Smart Contracts (RBIeINRBridge.sol, WrappedeINR.sol) |
| `335` | [`335_zk_light_client_crosschain_header_verifier_contract.md`](./prompts/335_zk_light_client_crosschain_header_verifier_contract.md) | ZK Light Client Cross-Chain Header & Transaction Inclusion Verifier Contract (ZKCrossChainLightClient.sol) |
| `336` | [`336_automated_tax_withholding_and_etds_ledger_contract.md`](./prompts/336_automated_tax_withholding_and_etds_ledger_contract.md) | Automated Tax Withholding and e-TDS Compliance Ledger Smart Contract (AutomatedTaxLedger.sol, ETDSWithholdingVault.sol) |
| `337` | [`337_erc4337_smart_account_and_session_keys_contract.md`](./prompts/337_erc4337_smart_account_and_session_keys_contract.md) | ERC-4337 Modular Smart Account & Ephemeral Session Keys Contract (Solidity) |
| `338` | [`338_groth16_zk_snark_investor_accreditation_verifier.md`](./prompts/338_groth16_zk_snark_investor_accreditation_verifier.md) | Groth16 ZK-SNARK Investor Accreditation & Jurisdictional Verifier Contract (Solidity / Circom) |
| `339` | [`339_concentrated_liquidity_clmm_synthetic_amm_contract.md`](./prompts/339_concentrated_liquidity_clmm_synthetic_amm_contract.md) | Concentrated Liquidity Market Maker (CLMM) Off-Hours AMM Contract (Solidity) |
| `340` | [`340_quantum_safe_lattice_signature_verifier_contract.md`](./prompts/340_quantum_safe_lattice_signature_verifier_contract.md) | Post-Quantum Lattice Signature (ML-DSA / Dilithium) Verifier Contract (Solidity / Yul) |
| `341` | [`341_circuit_breaker_timelock_and_emergency_halt_contract.md`](./prompts/341_circuit_breaker_timelock_and_emergency_halt_contract.md) | On-Chain Circuit Breaker, Timelock & Multi-Tier Emergency Pause Contract (Solidity) |
| `342` | [`342_fractional_share_rights_and_corporate_action_splitter.md`](./prompts/342_fractional_share_rights_and_corporate_action_splitter.md) | Fractional Share Rights & On-Chain Corporate Action Splitter Contract (Solidity) |
| `343` | [`343_cross_chain_atomic_swap_htlc_settlement_contract.md`](./prompts/343_cross_chain_atomic_swap_htlc_settlement_contract.md) | Cross-Chain Hashed Time-Locked Contract (HTLC) Atomic Settlement (Solidity) |
| `344` | [`344_rwa_dutch_auction_and_linear_vesting_contract.md`](./prompts/344_rwa_dutch_auction_and_linear_vesting_contract.md) | RWA Dutch Auction & On-Chain Linear Vesting Smart Contract (Solidity) |
| `345` | [`345_tokenized_yield_and_staking_vault_contract.md`](./prompts/345_tokenized_yield_and_staking_vault_contract.md) | Fixed-Income Tokenized Yield & Staking Vault Smart Contract (Solidity) |
| `346` | [`346_p2p_escrow_and_multisig_arbitration_contract.md`](./prompts/346_p2p_escrow_and_multisig_arbitration_contract.md) | P2P Escrow & 2-of-3 Multi-Sig Arbitration Smart Contract (Solidity) |
| `347` | [`347_affiliate_rebate_and_commission_splitter_contract.md`](./prompts/347_affiliate_rebate_and_commission_splitter_contract.md) | On-Chain Affiliate Rebate & Commission Splitter Smart Contract (Solidity) |
| `348` | [`348_virtual_faucet_and_demo_token_minting_contract.md`](./prompts/348_virtual_faucet_and_demo_token_minting_contract.md) | Virtual Faucet & Demo Token Minting Smart Contract (Solidity) |
| `349` | [`349_btc_usdt_onchain_atomic_dvp_settlement_contract.md`](./prompts/349_btc_usdt_onchain_atomic_dvp_settlement_contract.md) | Real-Money BTC/USDT Atomic Delivery-versus-Payment (DvP) Settlement Contract (Solidity) |
| `350` | [`350_mpc_tss_hot_cold_vault_multisig_coordinator_contract.md`](./prompts/350_mpc_tss_hot_cold_vault_multisig_coordinator_contract.md) | Institutional MPC-TSS Hot/Cold Vault Multi-Sig Coordinator Contract (Solidity) |

---

## 4. Shared Libraries, Protocols & SDKs (13 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `401` | [`401_postgresql_schema_design.md`](./prompts/401_postgresql_schema_design.md) | PostgreSQL Schema Design (Core Transactional Data) |
| `402` | [`402_redis_patterns_and_caching.md`](./prompts/402_redis_patterns_and_caching.md) | Redis Usage Patterns (Session, Cache, Real-Time Order Book State) |
| `403` | [`403_kafka_cluster_and_topic_partitioning.md`](./prompts/403_kafka_cluster_and_topic_partitioning.md) | Kafka Cluster Design & Topic Partitioning Strategy |
| `404` | [`404_data_warehouse_analytics_pipeline.md`](./prompts/404_data_warehouse_analytics_pipeline.md) | Data Warehouse / Analytics Pipeline (For Compliance & BI) |
| `405` | [`405_data_retention_and_archival.md`](./prompts/405_data_retention_and_archival.md) | Data Retention & Archival Policy Implementation |
| `406` | [`406_backup_restore_automation.md`](./prompts/406_backup_restore_automation.md) | Backup & Restore Automation for All Datastores |
| `407` | [`407_master_data_management.md`](./prompts/407_master_data_management.md) | Master Data Management (Securities Master, Corporate Actions Master) |
| `408` | [`408_historical_market_data_timescale_and_clickhouse_pipeline.md`](./prompts/408_historical_market_data_timescale_and_clickhouse_pipeline.md) | Historical Market Data Pipeline & Time-Series Analytics (TimescaleDB & ClickHouse) |
| `409` | [`409_proof_of_reserve_sparse_merkle_tree_store.md`](./prompts/409_proof_of_reserve_sparse_merkle_tree_store.md) | Proof-of-Reserve Sparse Merkle Sum Tree State Store & ZK Generator (services/por-smt-generator) |
| `410` | [`410_debezium_cdc_event_sourcing_and_outbox_architecture.md`](./prompts/410_debezium_cdc_event_sourcing_and_outbox_architecture.md) | Debezium CDC Transactional Outbox & Event-Sourced Ledger Pipeline |
| `411` | [`411_vector_database_pgvector_rag_financial_knowledge_store.md`](./prompts/411_vector_database_pgvector_rag_financial_knowledge_store.md) | Vector Database & pgvector RAG Financial Knowledge Store |
| `412` | [`412_apache_flink_real_time_stream_analytics_and_feature_store.md`](./prompts/412_apache_flink_real_time_stream_analytics_and_feature_store.md) | Apache Flink Real-Time Stream Analytics & Online Feature Store |
| `413` | [`413_hot_warm_cold_tiered_storage_and_s3_glacier_archival.md`](./prompts/413_hot_warm_cold_tiered_storage_and_s3_glacier_archival.md) | Hot-Warm-Cold Tiered Storage & S3 Glacier Immutable Regulatory Archival |

---

## 5. Cross-Platform Mobile & Desktop Client (Flutter) (43 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `501` | [`501_flutter_project_scaffolding.md`](./prompts/501_flutter_project_scaffolding.md) | Flutter Multi-Platform Project Scaffolding (Mobile + Desktop) |
| `502` | [`502_flutter_app_architecture_state_management.md`](./prompts/502_flutter_app_architecture_state_management.md) | Flutter App Architecture & State Management (Riverpod) |
| `503` | [`503_flutter_design_system_theming.md`](./prompts/503_flutter_design_system_theming.md) | Flutter Design System & Theming (Light/Dark, Platform-Adaptive) |
| `504` | [`504_flutter_onboarding_kyc_flow.md`](./prompts/504_flutter_onboarding_kyc_flow.md) | Flutter Onboarding & KYC Flow UI |
| `505` | [`505_flutter_authentication_ui.md`](./prompts/505_flutter_authentication_ui.md) | Flutter Authentication UI (Biometrics, MPIN, OTP, Session) |
| `506` | [`506_flutter_home_dashboard_screen.md`](./prompts/506_flutter_home_dashboard_screen.md) | Flutter Home / Dashboard Screen |
| `507` | [`507_flutter_market_watchlist_screen.md`](./prompts/507_flutter_market_watchlist_screen.md) | Flutter Market & Watchlist Screen with Real-Time Price Updates |
| `508` | [`508_flutter_security_detail_screen.md`](./prompts/508_flutter_security_detail_screen.md) | Flutter Security Detail Screen (TradingView-Style Charts) |
| `509` | [`509_flutter_order_placement_flow.md`](./prompts/509_flutter_order_placement_flow.md) | Flutter Order Placement Flow (Buy/Sell, Order Types, Confirmation) |
| `510` | [`510_flutter_portfolio_holdings_screen.md`](./prompts/510_flutter_portfolio_holdings_screen.md) | Flutter Portfolio & Holdings Screen (Fractional Units, P&L, Proof-of-Reserve) |
| `511` | [`511_flutter_wallet_funds_screen.md`](./prompts/511_flutter_wallet_funds_screen.md) | Flutter Wallet & Funds Screen (Deposit/Withdraw, UPI/Bank Linking) |
| `512` | [`512_flutter_transaction_trade_history_screen.md`](./prompts/512_flutter_transaction_trade_history_screen.md) | Flutter Transaction & Trade History Screen |
| `513` | [`513_flutter_notifications_center_ui.md`](./prompts/513_flutter_notifications_center_ui.md) | Flutter Notifications Center UI |
| `514` | [`514_flutter_settings_profile_ui.md`](./prompts/514_flutter_settings_profile_ui.md) | Flutter Settings, Profile & Regulatory Demat Management UI |
| `515` | [`515_flutter_offline_queued_orders.md`](./prompts/515_flutter_offline_queued_orders.md) | Flutter Offline Resilient & Queued Order (AMO) Management UX |
| `516` | [`516_flutter_android_packaging_signing.md`](./prompts/516_flutter_android_packaging_signing.md) | Platform-Specific Packaging: Android (Google Play Store) Build & Signing Pipeline |
| `517` | [`517_flutter_ios_packaging_signing.md`](./prompts/517_flutter_ios_packaging_signing.md) | Platform-Specific Packaging: iOS (App Store & TestFlight) Build & Signing Pipeline |
| `518` | [`518_flutter_windows_msix_packaging.md`](./prompts/518_flutter_windows_msix_packaging.md) | Platform-Specific Packaging: Windows (MSIX & Microsoft Store) Build Pipeline |
| `519` | [`519_flutter_linux_packaging.md`](./prompts/519_flutter_linux_packaging.md) | Platform-Specific Packaging: Linux (Flatpak, Snap & AppImage) Build Pipeline |
| `520` | [`520_flutter_macos_packaging_notarization.md`](./prompts/520_flutter_macos_packaging_notarization.md) | Platform-Specific Packaging: macOS (Notarized .app & .dmg) Build Pipeline |
| `521` | [`521_flutter_local_secure_storage.md`](./prompts/521_flutter_local_secure_storage.md) | Local Secure Storage (Keychain, KeyStore, DPAPI & Libsecret Abstraction) for Secrets |
| `522` | [`522_flutter_push_notifications.md`](./prompts/522_flutter_push_notifications.md) | Cross-Platform Push Notification Integration (Mobile FCM/APNs & Desktop System Notifications) |
| `523` | [`523_flutter_accessibility_localization.md`](./prompts/523_flutter_accessibility_localization.md) | Accessibility & Multi-Language Localization (i18n & a11y) Architecture |
| `524` | [`524_flutter_crash_reporting_analytics.md`](./prompts/524_flutter_crash_reporting_analytics.md) | Client-Side Crash Reporting, Performance Monitoring & Privacy-Preserving Analytics |
| `525` | [`525_flutter_api_client_layer.md`](./prompts/525_flutter_api_client_layer.md) | Flutter to Backend API Client Layer (Typed Dio, Mutex Token Refresh, Pinning & Retries) |
| `526` | [`526_flutter_deep_linking_universal_links.md`](./prompts/526_flutter_deep_linking_universal_links.md) | Deep Linking & Universal Links Across Mobile and Desktop Platforms |
| `527` | [`527_flutter_environment_switcher_and_sandbox_mode.md`](./prompts/527_flutter_environment_switcher_and_sandbox_mode.md) | Flutter In-App Environment Switcher, Mock KYC Toggle, Sandbox Mode Banner & Network Inspector |
| `528` | [`528_flutter_multichain_crypto_deposit_screen.md`](./prompts/528_flutter_multichain_crypto_deposit_screen.md) | Flutter Multi-Chain Crypto Deposit and Withdrawal Screen (BTC, ETH, SOL, POL, ARB) |
| `529` | [`529_flutter_options_chain_and_derivatives_screen.md`](./prompts/529_flutter_options_chain_and_derivatives_screen.md) | Flutter Options Chain and Derivatives Screen |
| `530` | [`530_flutter_commodity_physical_delivery_flow.md`](./prompts/530_flutter_commodity_physical_delivery_flow.md) | Flutter Commodity Physical Delivery and Vault Redemption Flow (gGOLD, gSILVER, WDRA eNWR, Armored Logistics) |
| `531` | [`531_flutter_gemini_conversational_assistant_screen.md`](./prompts/531_flutter_gemini_conversational_assistant_screen.md) | Flutter Gemini Conversational Assistant Screen & Interactive Drawer (Riverpod / Markdown / WebRTC) |
| `532` | [`532_flutter_multi_window_desktop_institutional_terminal.md`](./prompts/532_flutter_multi_window_desktop_institutional_terminal.md) | Flutter Desktop Multi-Window Institutional Trading Terminal |
| `533` | [`533_flutter_passkeys_and_webauthn_hardware_security_flow.md`](./prompts/533_flutter_passkeys_and_webauthn_hardware_security_flow.md) | Flutter FIDO2 Passkeys & Biometric Hardware Key Signing Flow |
| `534` | [`534_flutter_tradingview_charting_and_technical_indicators_engine.md`](./prompts/534_flutter_tradingview_charting_and_technical_indicators_engine.md) | Flutter High-Performance Custom Canvas & TradingView Charting Engine |
| `535` | [`535_flutter_market_heatmap_and_sector_treemap_screen.md`](./prompts/535_flutter_market_heatmap_and_sector_treemap_screen.md) | Flutter Real-Time Market Heatmap & Sector Treemap Screen |
| `536` | [`536_flutter_copy_trading_and_strategy_leaderboard_screen.md`](./prompts/536_flutter_copy_trading_and_strategy_leaderboard_screen.md) | Flutter Copy Trading & Strategy Leaderboard Screen |
| `537` | [`537_flutter_rwa_launchpad_and_primary_subscription_screen.md`](./prompts/537_flutter_rwa_launchpad_and_primary_subscription_screen.md) | Flutter RWA Launchpad & Primary Subscription Screen |
| `538` | [`538_flutter_p2p_fiat_trading_and_chat_escrow_screen.md`](./prompts/538_flutter_p2p_fiat_trading_and_chat_escrow_screen.md) | Flutter P2P Fiat Trading & Encrypted Chat Escrow Screen |
| `539` | [`539_flutter_crypto_earn_and_fixed_yield_staking_screen.md`](./prompts/539_flutter_crypto_earn_and_fixed_yield_staking_screen.md) | Flutter Fixed-Income Yield & Staking Vault Screen |
| `540` | [`540_flutter_btc_usdt_live_chart_and_tradingview_screen.md`](./prompts/540_flutter_btc_usdt_live_chart_and_tradingview_screen.md) | Flutter BTC/USDT Live TradingView Chart & Market Watch Screen |
| `541` | [`541_flutter_demo_vs_real_trading_mode_switcher_and_faucet_screen.md`](./prompts/541_flutter_demo_vs_real_trading_mode_switcher_and_faucet_screen.md) | Flutter Demo / Paper Trading Mode Switcher & Faucet Screen |
| `542` | [`542_flutter_btc_usdt_buy_sell_order_entry_sheet.md`](./prompts/542_flutter_btc_usdt_buy_sell_order_entry_sheet.md) | Flutter BTC/USDT Buy/Sell Order Entry Bottom Sheet & Execution Modal |
| `543` | [`543_flutter_btc_and_usdt_deposit_withdrawal_modal.md`](./prompts/543_flutter_btc_and_usdt_deposit_withdrawal_modal.md) | Flutter BTC & USDT Deposit & Withdrawal Management Modal |

---

## 6. Web Applications (Next.js 14 / TypeScript) (21 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `601` | [`601_nextjs_investor_web_app_scaffolding.md`](./prompts/601_nextjs_investor_web_app_scaffolding.md) | Next.js 14 Investor Web App Scaffolding |
| `602` | [`602_web_onboarding_kyc_flow.md`](./prompts/602_web_onboarding_kyc_flow.md) | Web Onboarding, DigiLocker & Camera KYC Flow |
| `603` | [`603_web_trading_dashboard.md`](./prompts/603_web_trading_dashboard.md) | Web Trading Terminal & Lightweight Charts Integration |
| `604` | [`604_admin_user_kyc_review_dashboard.md`](./prompts/604_admin_user_kyc_review_dashboard.md) | Admin Console: User Management & KYC Review Dashboard |
| `605` | [`605_admin_risk_exception_multiparty_approval.md`](./prompts/605_admin_risk_exception_multiparty_approval.md) | Admin Console: Risk Exceptions & Multi-Party Approval UI |
| `606` | [`606_admin_proof_of_reserve_reconciliation.md`](./prompts/606_admin_proof_of_reserve_reconciliation.md) | Admin & Public Proof-of-Reserve Verification Dashboard |
| `607` | [`607_admin_regulatory_reporting_dashboard.md`](./prompts/607_admin_regulatory_reporting_dashboard.md) | Admin Console: Regulatory Reporting & Audit Export Portal |
| `608` | [`608_public_marketing_landing_site.md`](./prompts/608_public_marketing_landing_site.md) | Public Landing Page, Education Hub & Risk Disclosures |
| `609` | [`609_developer_portal_and_testnet_faucet_ui.md`](./prompts/609_developer_portal_and_testnet_faucet_ui.md) | Public Developer Portal, Interactive API Explorer, Web3 Testnet Faucet & Mock Demat Faucet UI |
| `610` | [`610_web_options_chain_and_crosschain_deposit_portal.md`](./prompts/610_web_options_chain_and_crosschain_deposit_portal.md) | Web Options Chain, Strategy Builder & Multi-Chain Web3 Deposit Portal |
| `611` | [`611_web_commodity_physical_delivery_portal.md`](./prompts/611_web_commodity_physical_delivery_portal.md) | Web Commodity Physical Delivery Portal & Institutional Vault Management |
| `612` | [`612_web_institutional_dma_and_low_latency_hotkey_terminal.md`](./prompts/612_web_institutional_dma_and_low_latency_hotkey_terminal.md) | Next.js 14 Institutional Direct Market Access (DMA) Workstation |
| `613` | [`613_web_regulatory_portal_and_sebi_compliance_dashboard.md`](./prompts/613_web_regulatory_portal_and_sebi_compliance_dashboard.md) | Next.js 14 Regulatory Audit & SEBI/IFSCA Supervisory Dashboard |
| `614` | [`614_web_clearing_member_and_broker_capital_adequacy_portal.md`](./prompts/614_web_clearing_member_and_broker_capital_adequacy_portal.md) | Next.js 14 Clearing Member & Broker Capital Adequacy Portal |
| `615` | [`615_web_asset_issuer_and_tokenization_originator_portal.md`](./prompts/615_web_asset_issuer_and_tokenization_originator_portal.md) | Next.js 14 RWA Asset Issuer & Tokenization Originator Portal |
| `616` | [`616_web_copy_trading_and_master_trader_portal.md`](./prompts/616_web_copy_trading_and_master_trader_portal.md) | Next.js 14 Pro-Trader Copy Trading & Strategy Management Portal |
| `617` | [`617_web_sebi_scores_and_investor_grievance_portal.md`](./prompts/617_web_sebi_scores_and_investor_grievance_portal.md) | Next.js 14 Regulatory Grievance & SEBI SCORES 2.0 Management Portal |
| `618` | [`618_web_p2p_dispute_arbitration_and_operator_desk.md`](./prompts/618_web_p2p_dispute_arbitration_and_operator_desk.md) | Next.js 14 P2P Dispute Arbitration & Compliance Desk |
| `619` | [`619_web_rwa_primary_launchpad_investor_portal.md`](./prompts/619_web_rwa_primary_launchpad_investor_portal.md) | Next.js 14 RWA Primary Launchpad & Auction Investor Terminal |
| `620` | [`620_web_btc_usdt_pro_trading_terminal_and_orderbook.md`](./prompts/620_web_btc_usdt_pro_trading_terminal_and_orderbook.md) | Next.js 14 BTC/USDT Pro-Trading Terminal & Live Order Book Dashboard |
| `621` | [`621_web_demo_paper_trading_simulator_and_analytics_portal.md`](./prompts/621_web_demo_paper_trading_simulator_and_analytics_portal.md) | Next.js 14 Demo Paper Trading Simulator & Performance Analytics Portal |

---

## 7. Security, Compliance, Auditing & Market Surveillance (22 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `701` | [`701_full_system_threat_model_stride.md`](./prompts/701_full_system_threat_model_stride.md) | Full System Threat Model & STRIDE Analysis |
| `702` | [`702_iam_rbac_least_privilege.md`](./prompts/702_iam_rbac_least_privilege.md) | Identity & Access Management (IAM), RBAC & Privileged Access Management |
| `703` | [`703_sanctions_pep_screening_integration.md`](./prompts/703_sanctions_pep_screening_integration.md) | Global Sanctions & Politically Exposed Persons (PEP) Screening Integration |
| `704` | [`704_transaction_monitoring_aml_alerts.md`](./prompts/704_transaction_monitoring_aml_alerts.md) | Real-Time Transaction Monitoring & AML Suspicious Activity Alerting |
| `705` | [`705_pentesting_bug_bounty_plan.md`](./prompts/705_pentesting_bug_bounty_plan.md) | Penetration Testing, Red Teaming & Bug Bounty Program Plan |
| `706` | [`706_incident_response_runbook.md`](./prompts/706_incident_response_runbook.md) | Security Incident Response Runbooks & Playbooks |
| `707` | [`707_data_encryption_and_key_rotation.md`](./prompts/707_data_encryption_and_key_rotation.md) | Data Encryption Standards & Automated Key Rotation |
| `708` | [`708_secure_sdlc_code_review_policy.md`](./prompts/708_secure_sdlc_code_review_policy.md) | Secure SDLC, Branch Protection & Code Review Policy |
| `709` | [`709_third_party_vendor_security_checklist.md`](./prompts/709_third_party_vendor_security_checklist.md) | Third-Party Vendor Risk Assessment & Due Diligence Checklist |
| `710` | [`710_bcp_disaster_recovery_plan.md`](./prompts/710_bcp_disaster_recovery_plan.md) | Business Continuity Plan (BCP) & Disaster Recovery Framework |
| `711` | [`711_market_surveillance_anti_manipulation.md`](./prompts/711_market_surveillance_anti_manipulation.md) | Market Surveillance & Anti-Manipulation Engine (Spoofing, Layering, Wash Trading, Momentum Ignition) |
| `712` | [`712_insider_trading_graph_analytics.md`](./prompts/712_insider_trading_graph_analytics.md) | Insider Trading & UPSI Graph Analytics Engine |
| `713` | [`713_continuous_market_circuit_breaker_coordinator.md`](./prompts/713_continuous_market_circuit_breaker_coordinator.md) | Continuous Market Circuit Breaker & Resumption Coordinator |
| `714` | [`714_crosschain_aml_and_blockchain_forensics_screener.md`](./prompts/714_crosschain_aml_and_blockchain_forensics_screener.md) | Cross-Chain AML & Blockchain Forensics Screener (Bitcoin, Ethereum, Solana) |
| `715` | [`715_nbse_market_surveillance_and_sebi_reporting_engine.md`](./prompts/715_nbse_market_surveillance_and_sebi_reporting_engine.md) | NBSE Market Surveillance & SEBI Reporting Engine (Cross-Market Manipulation, Circular Trading, Front-Running) |
| `716` | [`716_mcx_wdra_physical_vault_audit_and_inspector_portal.md`](./prompts/716_mcx_wdra_physical_vault_audit_and_inspector_portal.md) | MCX & WDRA Physical Vault Audit & Inspector Portal (Assayer Attestations, Bar Weighment, e-NWR Verification) |
| `717` | [`717_hsm_kms_cloudhsm_key_lifecycle_and_signing_daemon.md`](./prompts/717_hsm_kms_cloudhsm_key_lifecycle_and_signing_daemon.md) | HSM & CloudHSM Key Lifecycle and Cryptographic Signing Daemon (FIPS 140-2 Level 3, EIP-712, BIP-174 PSBT, QBFT Block Signing) |
| `718` | [`718_zk_pedersen_blinding_and_batch_dvp_obfuscation_engine.md`](./prompts/718_zk_pedersen_blinding_and_batch_dvp_obfuscation_engine.md) | ZK Pedersen Blinding and Batch DvP Obfuscation Engine (DPDP Act 2023, GDPR, Homomorphic Sums, Timing-Graph De-anonymization Prevention) |
| `719` | [`719_sybil_and_collusion_graph_surveillance_engine.md`](./prompts/719_sybil_and_collusion_graph_surveillance_engine.md) | Sybil Ring & Cross-Account Collusion Graph Surveillance Engine (Python / Neo4j) |
| `720` | [`720_zero_knowledge_proof_of_liabilities_zk_pol_engine.md`](./prompts/720_zero_knowledge_proof_of_liabilities_zk_pol_engine.md) | Zero-Knowledge Proof of Liabilities (ZK-PoL) Sparse Merkle Tree Engine |
| `721` | [`721_post_quantum_cryptography_migration_and_hybrid_tls_spec.md`](./prompts/721_post_quantum_cryptography_migration_and_hybrid_tls_spec.md) | Post-Quantum Cryptography (PQC) Migration & Hybrid TLS Architecture |
| `722` | [`722_btc_usdt_anti_frontrunning_and_price_manipulation_guard.md`](./prompts/722_btc_usdt_anti_frontrunning_and_price_manipulation_guard.md) | BTC/USDT Anti-Frontrunning & Price Manipulation Surveillance Guard (Go) |

---

## 8. DevOps, CI/CD, Infrastructure as Code & Deployment (17 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `801` | [`801_docker_compose_local_dev_env.md`](./prompts/801_docker_compose_local_dev_env.md) | Full-Stack Docker Compose Local Development Environment |
| `802` | [`802_kubernetes_cluster_architecture.md`](./prompts/802_kubernetes_cluster_architecture.md) | Multi-Tenant Kubernetes Cluster Architecture & Namespace Segmentation |
| `803` | [`803_ci_pipeline_design.md`](./prompts/803_ci_pipeline_design.md) | Polyglot CI Pipeline Architecture & Automated Verification Matrix |
| `804` | [`804_cd_progressive_delivery_pipeline.md`](./prompts/804_cd_progressive_delivery_pipeline.md) | GitOps Continuous Delivery & Progressive Canary Rollouts |
| `805` | [`805_infrastructure_as_code_terraform.md`](./prompts/805_infrastructure_as_code_terraform.md) | Cloud & Blockchain Infrastructure as Code (Terraform / OpenTofu) |
| `806` | [`806_observability_stack_telemetry.md`](./prompts/806_observability_stack_telemetry.md) | Full-Stack Observability, OpenTelemetry & Distributed Tracing |
| `807` | [`807_centralized_logging_audit_pipeline.md`](./prompts/807_centralized_logging_audit_pipeline.md) | Centralized Logging, Immutable Audit Trail & Regulatory SIEM Pipeline |
| `808` | [`808_alerting_and_oncall_runbooks.md`](./prompts/808_alerting_and_oncall_runbooks.md) | Mission-Critical Alerting Architecture, SLO Framework & On-Call Runbooks |
| `809` | [`809_cost_monitoring_optimization.md`](./prompts/809_cost_monitoring_optimization.md) | FinOps Cloud Cost Monitoring, Allocation & Resource Optimization |
| `810` | [`810_environment_promotion_release_mgmt.md`](./prompts/810_environment_promotion_release_mgmt.md) | Environment Promotion, Release Management & Governance Workflows |
| `811` | [`811_testnet_vs_mainnet_dual_environment_cicd.md`](./prompts/811_testnet_vs_mainnet_dual_environment_cicd.md) | Dual-Environment GitOps CI/CD Pipeline: Testnet vs Mainnet, Safe Bytecode Verification & Rollout Gating |
| `812` | [`812_dual_environment_testnet_sandbox_and_mainnet_isolation.md`](./prompts/812_dual_environment_testnet_sandbox_and_mainnet_isolation.md) | Dual-Environment Testnet Sandbox, Mainnet Isolation & Orchestration Suite (Go / Kubernetes / Besu) |
| `813` | [`813_mock_depository_and_banking_sandbox_suite.md`](./prompts/813_mock_depository_and_banking_sandbox_suite.md) | High-Fidelity NSDL/CDSL Depository Simulator & RBI UPI/e-Rupee Banking Mock Sandbox Engine |
| `814` | [`814_bare_metal_dpdk_kernel_bypass_networking_spec.md`](./prompts/814_bare_metal_dpdk_kernel_bypass_networking_spec.md) | Bare-Metal DPDK & Solarflare Kernel-Bypass Network Architecture |
| `815` | [`815_zero_trust_mesh_spiffe_spire_workload_identity.md`](./prompts/815_zero_trust_mesh_spiffe_spire_workload_identity.md) | Zero-Trust SPIFFE/SPIRE Microservice Workload Identity Architecture |
| `816` | [`816_multi_region_geo_active_active_failover_architecture.md`](./prompts/816_multi_region_geo_active_active_failover_architecture.md) | Multi-Region Geo Active-Active Disaster Recovery & Cross-DC Consensus Architecture |
| `817` | [`817_testnet_sandbox_vs_mainnet_dual_matching_pipeline.md`](./prompts/817_testnet_sandbox_vs_mainnet_dual_matching_pipeline.md) | Dual-Environment Testnet (Demo) vs Mainnet (Real) Infrastructure & Matching Pipeline |

---

## 9. End-to-End Testing, Chaos Engineering & QA Automation (18 Prompts)

| Prompt ID | Specification File | Component Title |
| :--- | :--- | :--- |
| `901` | [`901_unit_integration_testing_strategy.md`](./prompts/901_unit_integration_testing_strategy.md) | Unit & Integration Testing Strategy Across Polyglot Stack |
| `902` | [`902_end_to_end_testing_flutter_backend.md`](./prompts/902_end_to_end_testing_flutter_backend.md) | End-to-End Testing Strategy: Flutter Client & Backend Integration |
| `903` | [`903_load_performance_testing_matching_engine.md`](./prompts/903_load_performance_testing_matching_engine.md) | Load & Performance Testing Plan (Matching Engine & Gateway Focus) |
| `904` | [`904_chaos_engineering_resilience_plan.md`](./prompts/904_chaos_engineering_resilience_plan.md) | Chaos Engineering & Resilience Testing Plan |
| `905` | [`905_security_testing_sast_dast_ci.md`](./prompts/905_security_testing_sast_dast_ci.md) | Automated Security Testing in CI/CD (SAST, DAST & Dependency Scanning) |
| `906` | [`906_uat_plan_regulatory_sandbox_scenarios.md`](./prompts/906_uat_plan_regulatory_sandbox_scenarios.md) | User Acceptance Testing (UAT) Plan for Real-World Regulatory Sandbox Scenarios |
| `907` | [`907_regulatory_sandbox_pilot_launch_plan.md`](./prompts/907_regulatory_sandbox_pilot_launch_plan.md) | Regulatory Sandbox Pilot Launch Plan & Phased Rollout |
| `908` | [`908_production_launch_rollback_runbook.md`](./prompts/908_production_launch_rollback_runbook.md) | Production Launch Runbook & Emergency Rollback Procedures |
| `909` | [`909_post_launch_monitoring_slo_error_budgets.md`](./prompts/909_post_launch_monitoring_slo_error_budgets.md) | Post-Launch Monitoring, SLOs & Error Budget Management |
| `910` | [`910_customer_support_grievance_redressal.md`](./prompts/910_customer_support_grievance_redressal.md) | Customer Support, Dispute Resolution & Grievance Redressal (SEBI SCORES / ODR Integration) |
| `911` | [`911_mainnet_dress_rehearsal_and_disaster_simulation.md`](./prompts/911_mainnet_dress_rehearsal_and_disaster_simulation.md) | Mainnet Dress Rehearsal, Dual-Region Failover Simulation, Shadow Production Traffic Replay, and Regulated Cutover Protocol |
| `912` | [`912_crosschain_bridge_and_derivatives_stress_testing_plan.md`](./prompts/912_crosschain_bridge_and_derivatives_stress_testing_plan.md) | Cross-Chain Bridge and Derivatives Stress Testing Plan |
| `913` | [`913_cross_market_reconciliation_and_settlement_fuzzing.md`](./prompts/913_cross_market_reconciliation_and_settlement_fuzzing.md) | Cross-Market Reconciliation & Automated Settlement Fuzzing Suite |
| `914` | [`914_one_crore_scale_concurrency_and_stress_testing_harness.md`](./prompts/914_one_crore_scale_concurrency_and_stress_testing_harness.md) | One-Crore Scale Concurrency, High-Throughput Stress Testing & Distributed Load Harness (Rust / Locust / k6 / Besu) |
| `915` | [`915_deterministic_market_replay_and_flash_crash_simulator.md`](./prompts/915_deterministic_market_replay_and_flash_crash_simulator.md) | Deterministic Market Replay & Extreme Volatility Flash Crash Simulator |
| `916` | [`916_formal_verification_and_symbolic_execution_testing_spec.md`](./prompts/916_formal_verification_and_symbolic_execution_testing_spec.md) | Formal Verification & Symbolic Execution Suite for Core Settlement Contracts |
| `917` | [`917_copy_trading_and_launchpad_concurrency_stress_suite.md`](./prompts/917_copy_trading_and_launchpad_concurrency_stress_suite.md) | Copy Trading Replication & Primary Dutch Auction Concurrency Stress Suite |
| `918` | [`918_btc_usdt_demo_and_real_trading_e2e_stress_suite.md`](./prompts/918_btc_usdt_demo_and_real_trading_e2e_stress_suite.md) | BTC/USDT Demo Paper Trading & Real-Money Concurrency E2E Stress Suite |

---
