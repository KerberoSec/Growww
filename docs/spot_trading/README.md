# Spot Trading on Blockchain: Master Architecture & UI/UX Suite (`docs/spot_trading`)

Welcome to the dedicated architectural, system design, and UI/UX documentation suite for the **Growww / NBSE Blockchain Spot Trading Platform**.

This suite provides complete operational specifications for spot trading, the hybrid Indian market + crypto fusion, the attractive dark-mode UI/UX design system, demo paper trading, real-money execution, and institutional security.

---

## Architectural Dossiers Catalog

| Document | Scope & Focus |
| :--- | :--- |
| **[01. Spot Trading Master Blueprint](./01_SPOT_TRADING_EXCHANGE_MASTER_BLUEPRINT.md)** | Core blockchain spot exchange architecture, Hyperledger Besu QBFT, 2s finality, and atomic DvP |
| **[02. Dark-Mode Attractive UI/UX Design System](./02_DARK_MODE_ATTRACTIVE_UI_UX_DESIGN_SYSTEM.md)** | Deep Obsidian dark palette, monospaced tabular typography, fluid TradingView charts, and micro-interactions |
| **[03. Indian Market + Crypto Fusion Architecture](./03_INDIAN_MARKET_AND_CRYPTO_FUSION_ARCHITECTURE.md)** | Reconciling Indian capital market rules (SEBI, NSDL/CDSL, INR, STT, CKYC) with 24/7 global crypto trading |
| **[04. Blockchain Spot Settlement & DvP Engine](./04_BLOCKCHAIN_SPOT_SETTLEMENT_AND_DVP_ENGINE.md)** | Atomic DvP smart contracts, EIP-712 trade signing, gasless Paymaster, and 0.00% fee (No fee at all) allocation |
| **[05. Spot Order Types, Matching & Execution](./05_SPOT_ORDER_TYPES_MATCHING_AND_EXECUTION.md)** | Market, Limit, Stop-Limit, Trailing Stop, Iceberg orders, in-place size reduction, and 500μs speed bumps |
| **[06. Demo Sandbox vs Real-Money Trading](./06_DEMO_SANDBOX_VS_REAL_MONEY_TRADING_ENGINE.md)** | Live external Bitcoin price feed, 10k USDT testnet faucet, portfolio reset, and zero-cross-contamination |
| **[07. Global Market Types & Derivatives Roadmap](./07_GLOBAL_EXCHANGE_MARKET_TYPES_AND_FUTURE_ROADMAP.md)** | Modern exchange market spectrum (Spot, USD-M Futures, Coin-M Futures, Options) and cross-margin roadmap |
| **[08. Settings, Security & User Management](./08_SETTINGS_SECURITY_AND_USER_MANAGEMENT_SPECIFICATION.md)** | 2FA, FIDO2 Passkeys, anti-phishing codes, address whitelisting, Dark Mode defaults, and DPDP Act rights |

---

## Core Operational Invariants
1. **Single Universal fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)):** 0 basis points (0.0000) charged on executed notional at launch. Zero brokerage, zero gas fees.
2. **100% Pre-Funded Spot Trading:** Traded assets are delivered immediately into user accounts via atomic DvP with zero leverage or liquidation risk.
3. **Dark Mode by Default:** Deep Obsidian (`#0B0E14`) theme with Emerald Green (`#00C087`) buys and Crimson Red (`#FF3B30`) sells.
4. **Live-Price Demo Parity:** Demo paper trading moves in lockstep with real-world live Bitcoin market prices.
5. **Zero-PII On Blockchain:** Identity verified via Poseidon zero-knowledge commitments; zero personal data stored on ledger.
