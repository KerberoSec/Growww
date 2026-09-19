# 50 Autonomous Engineering Agents & Subagents Master Roster

This directory contains the authoritative operational charters for all **50 autonomous engineering agents and subagents** responsible for executing, hardening, and verifying the Growww / NBSE trading platform architecture.

## Agent Hierarchy & Domain Division

| Tier | Domain Scope | Agent Count | Assigned Agents |
| :--- | :--- | :---: | :--- |
| **Tier 1** | Core Systems, Matching & Low Latency | **8 Agents** | `AGENT-01` to `AGENT-08` |
| **Tier 2** | Blockchain, Smart Contracts & DvP Clearing | **8 Agents** | `AGENT-09` to `AGENT-16` |
| **Tier 3** | Custody, Bitcoin & Multi-Chain Ingress | **8 Agents** | `AGENT-17` to `AGENT-24` |
| **Tier 4** | Indian Market & Crypto Fusion Architecture | **8 Agents** | `AGENT-25` to `AGENT-32` |
| **Tier 5** | Frontend UI/UX, Dark Mode & Mobile/Web Terminals | **8 Agents** | `AGENT-33` to `AGENT-40` |
| **Tier 6** | Compliance, Security, Data Platform & SRE | **10 Agents** | `AGENT-41` to `AGENT-50` |

---

## Complete Roster of 50 Agents

| Agent ID | Title & Job Role | Primary Focus & Domain | Assigned Prompts |
| :--- | :--- | :--- | :--- |
| `AGENT-01` | [`Lead Matching Engine & Orderbook Core Architect`](AGENT-01_lead_matching_engine_architect.md) | Core Rust L3 in-memory orderbook, Price-Time Priority FIFO, zero-... | Prompts 021, 030, 201, 202, 54... |
| `AGENT-02` | [`Order Lifecycle & Gateway Sequencing Engineer`](AGENT-02_order_lifecycle_gateway_engineer.md) | Client order ingress validation, monotonic 64-bit sequence taggin... | Prompts 022, 023, 203, 204, 54... |
| `AGENT-03` | [`Advanced Order Types & Algorithmic Slicing Specialist`](AGENT-03_advanced_order_algorithms_specialist.md) | Conditional Stop-Loss, Take-Profit, OCO bracket orders, dynamic T... | Prompts 024, 025, 026, 027, 02... |
| `AGENT-04` | [`Pre-Trade Risk Engine & Collateral Margin Guardian`](AGENT-04_pre_trade_risk_margin_guardian.md) | Sub-50 microsecond balance verification, fat-finger collars (+/- ... | Prompts 029, 205, 206, 547, 62... |
| `AGENT-05` | [`High-Availability Raft Sequencer & NVMe WAL Engineer`](AGENT-05_high_availability_raft_sequencer_engineer.md) | Append-only Write-Ahead Log (WAL) replication, Raft leader electi... | Prompts 030, 207, 814, 915... |
| `AGENT-06` | [`Market Data Broadcaster & L2/L3 Depth Specialist`](AGENT-06_market_data_broadcaster_depth_specialist.md) | Real-time ticker computation, rolling 24h high/low/volume metrics... | Prompts 061, 065, 066, 208, 20... |
| `AGENT-07` | [`Institutional FIX 5.0 SP2 Protocol Gateway Engineer`](AGENT-07_fix_protocol_institutional_gateway_engineer.md) | Financial Information eXchange (FIX 5.0 SP2) engine handling NewO... | Prompts 062, 210, 814, 903... |
| `AGENT-08` | [`Low-Latency DPDK & Linux Kernel Tuning Engineer`](AGENT-08_latency_optimization_dpdk_kernel_tuning_engineer.md) | Solarflare Onload user-space TCP acceleration, Mellanox VMA kerne... | Prompts 092, 814, 903, 914... |
| `AGENT-09` | [`Lead Blockchain Settlement & Besu Consortium Architect`](AGENT-09_lead_blockchain_settlement_architect.md) | Hyperledger Besu 4-validator private network, QBFT consensus (2-s... | Prompts 015, 101, 102, 301, 32... |
| `AGENT-10` | [`Atomic DvP Settlement Smart Contracts Engineer`](AGENT-10_atomic_dvp_smart_contracts_engineer.md) | Solidity DvPAtomicSettlement.sol contracts executing simultaneous... | Prompts 039, 302, 329, 349, 35... |
| `AGENT-11` | [`Gasless Paymaster & ERC-4337 Account Abstraction Engineer`](AGENT-11_gasless_paymaster_erc4337_engineer.md) | ERC-4337 EntryPoint v0.7 contracts, exchange-sponsored gas paymas... | Prompts 012, 303, 337, 353, 38... |
| `AGENT-12` | [`Sovereign Asset & Stablecoin Token Standards Engineer`](AGENT-12_sovereign_token_standards_engineer.md) | ERC-20, ERC-1400, and ERC-3643 security token contracts for wrapp... | Prompts 014, 304, 334, 351, 36... |
| `AGENT-13` | [`Settlement Relayer & Dynamic Gas Escalator Engineer`](AGENT-13_settlement_relayer_gas_escalator_engineer.md) | High-throughput batch submission daemon grouping matched trades i... | Prompts 039, 211, 282, 329... |
| `AGENT-14` | [`Virtual Testnet Faucet & Paper Trading Contracts Engineer`](AGENT-14_virtual_faucet_demo_contracts_engineer.md) | Testnet token minting contracts (vUSDT, vBTC), one-click auto-cre... | Prompts 033, 320, 348, 561, 63... |
| `AGENT-15` | [`Zero-Knowledge SNARK & Solvency Verifier Cryptographer`](AGENT-15_zk_snark_solvency_verifier_cryptographer.md) | Sparse Merkle Tree (SMT) construction, Groth16 zk-SNARK liability... | Prompts 018, 097, 338, 364, 39... |
| `AGENT-16` | [`Cross-Chain HTLC Atomic Swaps & Bridge Risk Engineer`](AGENT-16_crosschain_htlc_bridge_engineer.md) | Hashed Timelock Contracts (HTLC) supporting SHA-256 preimages, ti... | Prompts 019, 323, 343, 360, 38... |
| `AGENT-17` | [`Lead Crypto Custody & MPC-TSS Infrastructure Architect`](AGENT-17_lead_crypto_custody_architect.md) | Multi-Party Computation (MPC-CMP) threshold key management, Distr... | Prompts 017, 212, 350, 717... |
| `AGENT-18` | [`Bitcoin Taproot (P2TR) & SPV Header Verification Engineer`](AGENT-18_bitcoin_taproot_utxo_ingress_engineer.md) | Native Bitcoin UTXO tracking, BIP-341/342 Taproot script-path dep... | Prompts 013, 213, 322, 564, 64... |
| `AGENT-19` | [`Bitcoin Lightning Network & Instant Ingress Specialist`](AGENT-19_bitcoin_lightning_network_specialist.md) | Lightning Network channel management (LND), sub-second invoice ge... | Prompts 013, 214, 564, 640... |
| `AGENT-20` | [`USDT Multichain Deposit & Hot-to-Cold Sweeper Engineer`](AGENT-20_usdt_multichain_deposit_sweeper_engineer.md) | Multichain deposit listeners across Ethereum (ERC-20), Tron (TRC-... | Prompts 014, 215, 378, 565, 64... |
| `AGENT-21` | [`Crypto Withdrawal Security, Time-Lock & AML Engineer`](AGENT-21_crypto_withdrawal_aml_timelock_engineer.md) | Mandatory 24-hour security time-locks, address book whitelisting,... | Prompts 040, 077, 216, 567, 56... |
| `AGENT-22` | [`Hardware Security Module (HSM) & FIPS 140-3 Specialist`](AGENT-22_hardware_security_module_hsm_specialist.md) | AWS CloudHSM cluster management, on-premises YubiHSM2 integration... | Prompts 020, 717, 814, 911... |
| `AGENT-23` | [`Cross-Chain Re-org & Saga Compensation Coordinator`](AGENT-23_crosschain_reorg_saga_coordinator_engineer.md) | Monitors public blockchains for chain reorganizations and execute... | Prompts 014, 217, 323, 714... |
| `AGENT-24` | [`Cold Vault Air-Gap & Multisig Ceremony Coordinator`](AGENT-24_cold_vault_airgap_ceremony_coordinator.md) | Deep cold vault air-gapped laptop signing ceremonies, Shamir Secr... | Prompts 096, 717, 814, 911... |
| `AGENT-25` | [`Lead Indian Market & Crypto Fusion Regulatory Architect`](AGENT-25_lead_hybrid_finance_regulatory_architect.md) | Harmonizing 24/7 crypto spot trading with traditional Indian equi... | Prompts 002, 003, 041, 042, 11... |
| `AGENT-26` | [`RBI e-Rupee (CBDC Digital Rupee) Integration Engineer`](AGENT-26_rbi_cbdc_edigital_rupee_bridge_engineer.md) | Interfacing with RBI 2-tier CBDC distributor banks (SBI, ICICI, I... | Prompts 043, 218, 334, 365... |
| `AGENT-27` | [`Fiat INR Banking Rails & NPCI UPI 2.0 Gateway Engineer`](AGENT-27_fiat_inr_banking_gateway_engineer.md) | NPCI UPI 2.0 dynamic QR codes, virtual escrow account IMPS/NEFT a... | Prompts 038, 219, 566, 642... |
| `AGENT-28` | [`Zero-TDS Tax Lot & Decoupled PnL Accounting Engineer`](AGENT-28_section_194s_tds_withholding_tax_engineer.md) | zero on-chain TDS (0% withholding) calculation on crypto gross sale consideration, ... | Prompts 044, 220, 336, 369, 57... |
| `AGENT-29` | [`Section 115BBH 30% Flat VDA Profit/Loss Engine`](AGENT-29_section_115bbh_vda_capital_gains_engineer.md) | FIFO (First-In, First-Out) cost-basis lot matching, computing 30%... | Prompts 045, 221, 576, 651... |
| `AGENT-30` | [`Demat Share Margin Pledge & Depository Linkage Engineer`](AGENT-30_demat_margin_pledge_depository_engineer.md) | NSDL / CDSL electronic Margin Pledge/Repledge APIs, SEBI statutor... | Prompts 046, 222, 370, 577, 65... |
| `AGENT-31` | [`P2P Fiat-to-USDT Smart Contract Escrow & Dispute Engineer`](AGENT-31_p2p_fiat_escrow_dispute_arbitration_engineer.md) | Smart contract escrow on Besu, WebRTC encrypted buyer-seller chat... | Prompts 048, 223, 346, 359, 56... |
| `AGENT-32` | [`Contract Note Generator & Class 3 PKI Signature Engineer`](AGENT-32_contract_note_generator_pki_signature_engineer.md) | Automated daily consolidated Contract Notes formatted per SEBI br... | Prompts 050, 224, 555, 635... |
| `AGENT-33` | [`Lead Frontend & Obsidian Dark Mode Design System Architect`](AGENT-33_lead_frontend_design_system_architect.md) | Deep Obsidian (#0B0E14) palette, high-contrast neon green (#00F0A... | Prompts 051, 052, 503, 593, 66... |
| `AGENT-34` | [`Flutter Mobile Spot Trading Workstation Engineer`](AGENT-34_flutter_mobile_trading_workstation_engineer.md) | BTC/USDT spot trade screens, sliver app bars, custom orderbook wi... | Prompts 506, 509, 542, 544, 54... |
| `AGENT-35` | [`Flutter CustomPainter Orderbook & Depth Ladder Specialist`](AGENT-35_flutter_orderbook_depth_ladder_specialist.md) | CustomPainter-backed Level 2 orderbook ladder rendering horizonta... | Prompts 507, 545, 556... |
| `AGENT-36` | [`TradingView Charting Canvas & Technical Indicators Specialist`](AGENT-36_tradingview_charting_bridge_specialist.md) | Integrating TradingView Advanced Charts library across Next.js an... | Prompts 053, 534, 540, 556, 55... |
| `AGENT-37` | [`Haptic Feedback & Sound Engineering Sensory Specialist`](AGENT-37_haptic_sound_sensory_engineering_specialist.md) | Subtle tactile haptic triggers (selection click, medium impact, c... | Prompts 054, 057, 591, 663... |
| `AGENT-38` | [`Demo Paper Trading HUD & Virtual Faucet UI Engineer`](AGENT-38_demo_paper_trading_hud_faucet_engineer.md) | Persistent golden amber demo mode banner, one-click 10,000 vUSDT ... | Prompts 031, 032, 033, 035, 54... |
| `AGENT-39` | [`Next.js 14 Pro Trading Workstation & Docking Engineer`](AGENT-39_nextjs_pro_workstation_docking_engineer.md) | Dockview customizable grid layouts, detachable multi-monitor wind... | Prompts 059, 603, 620, 622, 62... |
| `AGENT-40` | [`Mobile Lockscreen Live Activities & Push Notification Engineer`](AGENT-40_mobile_lockscreen_live_activities_engineer.md) | iOS 17+ Dynamic Island / Live Activities widgets, Android persist... | Prompts 058, 522, 588, 589, 59... |
| `AGENT-41` | [`FIU-IND AML/CFT & IVMS-101 Travel Rule Compliance Officer`](AGENT-41_fiu_ind_aml_travel_rule_compliance_officer.md) | Real-time transaction surveillance for cash structuring and rapid... | Prompts 004, 049, 704, 714, 67... |
| `AGENT-42` | [`Market Surveillance & Anti-Manipulation Graph Engineer`](AGENT-42_market_surveillance_anti_manipulation_engineer.md) | Detecting wash trading, orderbook spoofing, layering, circular tr... | Prompts 074, 711, 712, 719, 72... |
| `AGENT-43` | [`DPDP Act 2023 Privacy & Crypto-Shredding Engineer`](AGENT-43_dpdp_act_privacy_crypto_shredding_engineer.md) | Granular consent management, data portability exports (encrypted ... | Prompts 008, 079, 585, 659... |
| `AGENT-44` | [`Biometric KYC & AI Facial Liveness Specialist`](AGENT-44_biometric_kyc_facial_liveness_specialist.md) | UIDAI Aadhaar OTP e-KYC, DigiLocker OAuth2 document fetch, NSDL P... | Prompts 076, 504, 583, 584, 60... |
| `AGENT-45` | [`Database HA (Patroni) & Debezium CDC Streaming Architect`](AGENT-45_database_patroni_cdc_streaming_architect.md) | PostgreSQL 16 multi-AZ cluster with Patroni failover, Debezium ch... | Prompts 401, 410, 801, 802... |
| `AGENT-46` | [`Time-Series (TimescaleDB & ClickHouse) Data Architect`](AGENT-46_time_series_timescaledb_clickhouse_architect.md) | Sub-second candlestick querying on TimescaleDB hypertables, 7-yea... | Prompts 064, 408, 412, 413... |
| `AGENT-47` | [`Kubernetes Mesh & Cilium eBPF Infrastructure SRE Lead`](AGENT-47_kubernetes_mesh_cilium_ebpf_sre_lead.md) | EKS Multi-AZ cluster orchestration, Cilium eBPF high-throughput C... | Prompts 099, 802, 814, 815... |
| `AGENT-48` | [`Continuous Chaos Engineering & Network Partition Resilience Lead`](AGENT-48_continuous_chaos_engineering_resilience_lead.md) | Automated Chaos Mesh and LitmusChaos injection pipelines simulati... | Prompts 091, 904, 911, 915... |
| `AGENT-49` | [`One Crore (10 Million) Concurrency Stress Testing Lead`](AGENT-49_one_crore_scale_stress_testing_lead.md) | Distributed load testing clusters simulating 10,000,000 WebSocket... | Prompts 098, 903, 914, 918... |
| `AGENT-50` | [`Master Production Readiness & Regulatory Gatekeeper`](AGENT-50_master_production_readiness_release_gatekeeper.md) | Definitive production release gate: zero critical SAST/DAST vulne... | Prompts 100, 701, 708, 908, 95... |
