# Growww Platform Features & Capabilities Catalog

Comprehensive catalogue of all user-facing, institutional, infrastructure, smart contract, and operational features of the **Growww Next-Gen Real World Asset (RWA) Tokenization & 24/7 Trading Platform**.

---

## 1. Fee Model & Economic Policy

### 1.1 Single 0.00% fee (No fee at all) Invariant
- **0 Brokerage:** ₹0 brokerage on all delivery, intraday, and fractional trades across all asset classes.
- **0 Deposit / Withdrawal Fees:** ₹0 platform fees on UPI, IMPS, NEFT, RTGS, and NetBanking transfers.
- **0 Platform Surcharges:** No hidden account maintenance charges (AMC), no call-and-trade fees, no platform markup.
- **Single 0.00% (No fee at all) Platform Fee:** Exactly 0.00% (Zero Fee at launch; 0 bps maker / 0 bps taker) calculated on executed trade notional value:
  $$\text{Total Fee} = \text{Executed Notional} \times 0.0000 \quad (0.00\% \text{ at launch; FeeController governed})$$
- **Automatic Internal Allocation:** The platform's 0.00% fee (No fee at all) internally funds all operations:
  - **Treasury Reserve Pool:** Infrastructure, liquidity provision, and gas sponsorship.
  - **Settlement Guarantee Fund (SGF):** Central counterparty default insurance.
  - **Investor Protection Fund (IPF):** Retail investor fraud and indemnity protection (governed dynamically via FeeController.sol).
- **100% Gasless Trading:** All blockchain interactions (token transfers, DvP settlement, whitelisting) are gasless for end users via EIP-2771 meta-transaction Paymasters sponsored by the Treasury.

### 1.2 Forced Liquidation 1.0% Penalty Fee (Exchange Risk Protection)
- **Normal Trading vs Forced Liquidation:** Normal voluntary trades placed by users strictly pay the single flat **0.00% (Zero Fee)** platform fee. Forced liquidations triggered by the exchange risk engine carry a dedicated **1.0% penalty fee** (100 basis points / 0.0100) on the forcefully liquidated notional value.
- **Liquidation Trigger Threshold:** Triggered automatically when an under-collateralized position breaches its Maintenance Margin Requirement (MMR).
- **Formula:**
  $$\text{Fee}_{liq} = \text{round}\left(\text{Liquidated Notional} \times 0.0100\right)$$
- **Insurance Fund Capitalization:**
  - **80% to SGF / Insurance Fund:** Direct capital buffer to absorb negative balance bankruptcy deficits and protect exchange solvency.
  - **20% to Treasury Relayer Pool:** Covers gas sponsorship and compute overhead for the liquidation execution bots.
- **Zero Trader Clawback Guarantee:** If market slippage causes an account to enter negative equity during forced liquidation, the deficit is absorbed 100% by the Insurance Fund / SGF with zero haircut or socialized clawback from profitable traders.

---

## 2. Trading & Market Microstructure Features

### 2.1 Order Types & Execution Modes
- **Standard Limit Orders:** Strict Price-Time (FIFO) queue priority execution.
- **Market Orders:** Immediate-or-Cancel (IOC) or Fill-or-Kill (FOK) with configurable slippage protection (default $\pm 1\%$).
- **Stop-Loss & Stop-Loss Limit Orders:** Evaluated against real-time Last Traded Price (LTP) with sliding high/low watermarks.
- **Trailing Stop-Loss Orders:** Dynamically tracks favorable market moves with configurable trailing points or percentage trail offsets.
- **Passive Midpoint Peg Orders:** Rerouted to mid-quote with anti-feedback volatility clamping.
- **In-Place Size Reductions:** Reducing open order quantity ($Q_2 < Q_1$) updates size in-place while preserving FIFO queue priority.

### 2.2 24/7 Continuous Trading & Session Management
- **Primary Market Mirroring:** Uninterrupted tracking during Indian exchange hours (09:15 to 15:30 IST).
- **24/7 Synthetic & Off-Hours Secondary Trading:** 24/7 secondary tokenized liquidity with dynamic price bands ($\pm 3\%, \pm 5\%, \pm 8\%$).
- **Indicative Equilibrium Price (IEP) Call Auctions:** 15-minute call auctions for opening sessions and circuit-breaker resumption utilizing Dual Fenwick Trees for $O(\log K)$ price discovery.
- **Self-Trade Prevention (STP):** Beneficial-owner PAN matching with 4 configurable modes: Cancel Aggressive, Cancel Passive, Cancel Both, Decrement and Cancel.

---

## 3. Market Data & Candlestick Aggregation

### 3.1 Real-Time Streaming & Level-2/3 Depth
- **Sub-Millisecond Level-2 Depth:** Top 20/50 bid-ask depth streaming with atomic multi-tier batch updates and CRC32 integrity checksums.
- **Ultra-Low Latency Ticker Streams:** Real-time LTP, 24-hour volume, open, high, low, and VWAP broadcasts.
- **Adaptive WebSocket Backpressure:** 256-packet per-client ring buffer with automatic 200ms conflated snapshot mode for constrained mobile networks, eliminating latency drift and server out-of-memory errors.
- **In-Band Token Rotation:** Zero-disconnection session refresh allows 24/7 active streaming without dropping socket connections.

### 3.2 Dual Candlestick Time Series
- **Immutable Raw OHLCV (`raw_ohlcv_1m`):** Exact execution-time trade bars for statutory compliance and regulatory audits.
- **Dynamically Split-Adjusted OHLCV (`adjusted_ohlcv_1m`):** Real-time backward adjusted series reflecting corporate stock splits and consolidations for seamless technical analysis.
- **Late-Arriving Trade Sliding Watermark:** 1500ms grace period emits `CANDLE_REVISE` updates for late fills without disrupting subsequent bar open-close boundaries.
- **Zero-Volume Carry-Forward:** Seamless continuous chart rendering during zero-volume intervals with synthetic flat candles.

---

## 4. Real World Asset (RWA) Tokenization & Custody

### 4.1 Tokenization Domains
- **Tokenized Indian Equities (Phase 1/2):** Discrete, fractional units ($10^{-6}$) backed 1:1 by depository demat balances (NSDL/CDSL).
- **Tokenized Government Securities (G-Secs & T-Bills):** High-yield sovereign debt tokenization with automated daily accrued interest calculations.
- **Tokenized Corporate Bonds:** Fixed and floating rate bonds with automated coupon distribution directly to investor on-chain balances.
- **GIFT City International Assets:** Multi-currency US Equities, Global ETFs, and FX tokenization via IFSC banking rails.

### 4.2 Proof of Reserve (PoR) & Depository Invariants
- **Real-Time Depository Synchronization:** Automated cryptographic reconciliation between on-chain ERC-3643 supply and custodian depository ledgers.
- **Corporate Action Cash-in-Lieu (CIL) Engine:** Automatic pro-rata cash distribution for fractional shares during rights, splits, and dividend events.
- **Pre-Settlement Deposit Finality Gating:** Unconfirmed depository transfers gated to restricted pre-settlement trading books, preventing systemic delivery default unwinds.

---

## 5. Wallet, Multi-Currency & Ledger Engine

### 5.1 Immutable Double-Entry Ledger
- **Fixed-Scale Integer Math:** Micro-currency precision ($1\text{ weINR} = ₹0.0001 = 10^{-4}\text{ INR}$) ensuring zero floating-point rounding errors.
- **Balanced Transaction Constraint:** Atomic multi-leg journal entries with verified $\sum \text{Debits} \equiv \sum \text{Credits}$.
- **Pessimistic Row-Level Lock Isolation:** High-concurrency wallet operations protected by PostgreSQL row-level locks (`SELECT ... FOR UPDATE`), eliminating overdraft and double-spend race conditions.

### 5.2 Payment Rails & Instant Settlement
- **UPI 2.0 AutoPay & QR:** Instant mandate creation and real-time one-click fund transfers.
- **Instant IMPS / RTGS / NEFT:** Automated bank webhook ingestion with deterministic two-phase idempotency.
- **Multi-Currency Escrow (INR / USD / EUR):** Real-time FX conversion with virtual hold escrow for cross-border GIFT City investments.

---

## 6. Client Applications, Web Pro Terminal & Desktop Thick Clients

Reference Master Blueprint: [`docs/architecture/UNIVERSAL_MULTI_PLATFORM_CLIENT_ARCHITECTURE_AND_THICK_CLIENT_BLUEPRINT.md`](architecture/UNIVERSAL_MULTI_PLATFORM_CLIENT_ARCHITECTURE_AND_THICK_CLIENT_BLUEPRINT.md)

### 6.1 Web Pro Trading Terminal (`apps/growww_web` - Main Focus)
- **Multi-Monitor Window Popping & Docking:** Customizable Dockview multi-split grid with detachable pop-out windows via `BroadcastChannel` and `SharedWorker`, maintaining sub-millisecond state coherence across physical monitors.
- **Off-Thread Web Worker Streaming:** High-frequency binary Protobuf WebSocket stream decodes in WebAssembly inside dedicated Web Workers, updating state via `SharedArrayBuffer` at 50ms conflation to keep the UI thread locked at 60/120 FPS.
- **WebGL / WebGPU Hardware Accelerated Charting:** Canvas-based TradingView Advanced Charts with 100+ technical indicators, depth distribution profiles, and zero Cumulative Layout Shift (CLS = 0).
- **Zero-Latency Keyboard Hotkey Suite:** `Shift+B` (Buy Market), `Shift+S` (Sell Market), `Escape` (Panic Cancel All), `Space` (Ticker Palette), and margin lot selectors (`1`, `2`, `5`, `0`).
- **WebAuthn & Non-Custodial Web3:** Biometric sign-in via FIDO2 Passkeys, MetaMask / WalletConnect support, and EIP-4361 (Sign-In with Ethereum - SIWE).

### 6.2 macOS Native Thick Client (`apps/growww_flutter` Desktop - Main Focus)
- **Metal Hardware Acceleration (Impeller):** Renders 50-depth order books and charts at 120 FPS on Apple ProMotion displays (MacBook Pro, Studio Display, Pro Display XDR).
- **Apple Silicon Native Optimization:** Universal Mach-O binary (`arm64` + `x86_64`) with ARM64 NEON SIMD optimizations compiled directly into the Rust FFI core (`rust_trading_core`).
- **Apple Secure Enclave & TouchID:** Hardware-isolated credential storage in macOS Keychain with biometric authorization via `LocalAuthentication`.
- **Native Cocoa Menu Bar & Dock Integration:** Global menu bar commands (`Cmd+1` to `Cmd+9`), dock icon live PnL badge, and multi-display detachment with mixed DPI scaling.

### 6.3 Windows Native Thick Client (`apps/growww_flutter` Desktop - Main Focus)
- **DirectX 12 / Direct3D Hardware Acceleration:** High-refresh support driving 144Hz, 240Hz, and 360Hz monitors with adaptive G-Sync/FreeSync.
- **Windows Hello Biometric Quick-Auth:** Instant facial recognition and fingerprint validation for high-value trade authorization.
- **Global OS Keyboard Hooks (`RegisterHotKey`):** Win32 global hook allows panic-canceling active orders (`Ctrl+Alt+Space`) even when the app is running in the background.
- **Per-Monitor DPI v2 & System Tray:** Seamless window dragging across disparate DPI displays and background system tray ticker status.

### 6.4 Linux Native Thick Client (`apps/growww_flutter` Desktop - Everywhere)
- **GTK 3/4 & Wayland/X11 Multi-Head:** Vulkan-backed hardware rendering across multi-head setups, packaged via Flatpak, AppImage, and Snap with Linux Secret Service API (`libsecret`) integration.

### 6.5 Mobile Client Applications (`apps/growww_flutter` Mobile - iOS & Android)
- **Ergonomic Trading & Tabular Figures:** Obsidian Dark Theme (`#0B0E14`), neon green (`#00F0A0`), neon red (`#FF3B56`), tabular fonts (`JetBrains Mono`), and single-thumb Swipe-to-Trade.
- **Biometric Quick-Auth & Tactile Haptics:** FaceID / TouchID / Android BiometricPrompt with native selection, impact, and alert haptic pulses.
- **Battery-Optimized Lifecycle & Offline Queue:** Socket graceful background suspension, foreground snapshot + delta rehydration, and encrypted SQLite WAL offline order queuing.

### 6.6 Universal Cross-Platform Synchronization ("Everywhere")
- **Cloud-Synced Workspace Profiles:** Multi-monitor window geometries, indicator setups, and shortcut bindings sync seamlessly between Web, macOS, Windows, and Linux.
- **Color-Coded Link Groups:** 4 Link Groups (Red, Blue, Green, Yellow) locking symbol synchronization across all detached windows and devices.
- **Zero-Fee Presentation Invariant:** Strict 0.00% maker / 0.00% taker / 0 gas (ERC-4337 Paymaster) / 0 TDS displayed uniformly across all 6 platform interfaces.

---

## 7. Security, Privacy & Regulatory Compliance

### 7.1 DPDP Act & Privacy Shield
- **Crypto-Shredding Erasure:** User Data Encryption Keys (DEKs) destroyed in HSM upon erasure requests, mathematically invalidating historical ciphertexts in append-only storage.
- **PII-Free Telemetry:** Automated redaction of Aadhaar, PAN, and phone numbers in Structlog and OpenTelemetry pipelines.
- **Poseidon ZK Identity Commitments:** On-chain identity verification without revealing investor credentials.

### 7.2 SEBI, RBI, FIU-IND & IFSCA Compliance
- **4-Window SEBI Peak Margin Snapshots:** Automated intraday margin capture across 4 random hourly time slices.
- **AML / Sanctions Screening:** Automated screening against UN, OFAC, and FIU-IND watchlists during onboarding and large withdrawals.
- **Statutory Audit Logs:** Cryptographically signed, append-only immutable audit trail in Rust (`audit-service`) recorded on Hyperledger Besu.

---

## 8. Derivatives, Perpetual Futures & Margin Engine
- **Fair Mark Price Engine:** Liquidations evaluated against 5-exchange weighted median spot index + 30m basis TWAP, eliminating unnatural wick liquidations.
- **8-Hour Clamped Funding Rate:** Clamped within $[-0.75\%, +0.75\%]$ exchanged peer-to-peer with ₹0 platform deduction.
- **Cross-Margin Multi-Asset Collateral:** Unified collateral valuation across G-Secs (95%), Equities (80%), and Crypto (80%).
- **Tiered Leverage Brackets:** Dynamic leverage reduction from 20x for retail positions down to 2x for block trades $> \text{₹10 Crores}$.
- **Auto-Deleveraging (ADL) Queue:** Highest-ranking profitable counterparty positions systematically deleveraged only upon total Insurance Fund depletion.

---

## 9. Cryptographic Merkle-Tree Proof of Solvency
- **zk-SNARK Blinded Merkle Sum Tree:** Proves 100% asset backing ($Assets \ge Liabilities$) with zero negative leaf balances and zero balance leakage.
- **On-Chain Digital Signature Verification:** Cryptographic proof of ownership over cold vault addresses on Bitcoin, Ethereum, and Solana.
- **User Self-Verification:** In-app verification of individual balance inclusion in the published Merkle root in $O(\log N)$ time.

---

## 10. Institutional FIX Protocol & Binary Protocols
- **FIX 4.4 / 5.0 SP2 Gateway:** Standard Tag 35 messages (`D`, `F`, `G`, `8`) and Drop Copy real-time risk feeds.
- **Ultra-Low Latency Binary Feeds:** Binary OUCH order entry ($< 8.5\mu\text{s}$) and Binary ITCH multicast Level-3 tick dissemination.
- **Colocation Cross-Connects:** Direct 10Gbps/25Gbps fiber cross-connects in Tier-4 Equinix Mumbai and GIFT City datacenters.

---

## 11. P2P Fiat Escrow & Dispute Resolution
- **Automated Escrow Lockbox:** Seller crypto locked in smart escrow upon trade initiation with 15-minute countdown timer.
- **Strict KYC Name Invariant:** Mandatory 100% match between buyer bank account name and KYC record (third-party payments strictly prohibited).
- **Maker-Checker Arbitration:** Multi-tier dispute resolution with automated Open Banking UTR settlement verification.

---

## 12. Crypto Earn, Liquid Staking & Yield Vaults
- **Flexible Savings:** Daily auto-compounding yield on idle USDT/INR backed by sovereign treasury repo yields with instant zero-fee redemption.
- **PoS Validator Staking:** Delegated non-custodial liquid staking (ETH, SOL, POL) backed by an internal Slashing Insurance Fund guaranteeing 100% principal protection.
- **Instant Exit Liquidity Pools:** Instant unbonding swap route at flat 0.05% swap fee.

---

## 13. Multi-Tier Affiliate & Referral Engine
- **2-Tier Commission Hierarchy:** 20% to 40% direct Tier-1 commission and 5% to 10% sub-Tier-2 commission.
- **Daily Automated Settlement:** Real-time commission accrual with automated daily settlement directly to affiliate wallets.
- **Kickback Sharing:** Affiliates can share up to 50% of commissions back to referees as an automated trading fee rebate.

---

## 14. Bitcoin (BTC/USDT) Demo & Real-Money Spot Trading
- **Live Market Price Ingestion:** Redundant WebSocket connectors to Binance, Coinbase, Kraken, and OKX with a 5-second trimmed median filter and sliding OHLCV candlestick aggregation.
- **Demo / Paper Trading Environment (Testnet):**
  - **1-Click Testnet Faucet:** 10,000 virtual USDT and 1.0 virtual BTC per user with 24-hour rate limit.
  - **Live Price Execution:** Simulated orders execute against real-world live Bitcoin market prices with realistic synthetic slippage and depth models.
  - **Zero Financial Risk:** Complete isolation between virtual demo ledgers and real-money balances.
- **Real-Money Spot Trading (Mainnet):**
  - **On-Chain Bitcoin Deposit:** Native SegWit (`bc1q`) and Taproot (`bc1p`) address monitoring with automated credit upon 3 block confirmations.
  - **Multi-Chain USDT Deposit:** ERC-20, TRC-20, and Polygon USDT ingress with automated hot-to-cold vault sweeps.
  - **Sub-Millisecond Execution:** In-memory Rust matching engine with single-writer sequencer and deterministic price-time priority.
  - **Atomic DvP Settlement:** On-chain DvP batch settlement on Hyperledger Besu with automatic 0.00% (Zero Fee) platform fee split (0.00% fee at launch; future fee parameters governed by FeeController.sol).
  - **Institutional MPC-TSS Custody:** 3-of-5 threshold signature security over cold vault reserves.

---

## 15. Copy Trading & Social Strategy Replication
- **Proportional Position Sizing:** Pro-rata replication across up to 10,000 followers per master trader within 15 milliseconds.
- **High-Water Mark (HWM) Profit Sharing:** 10% to 15% performance fee calculated weekly on net new trading profits.
- **Follower Circuit Breakers:** Automatic detachment if slippage exceeds 0.5% or master drawdown breaches configured risk limits.

---

## 16. RWA Primary Launchpad & Dutch Auctions
- **Linear Dutch Auction Price Descent:** Uniform clearing price discovery for tokenized real estate, private debt, and sovereign bonds.
- **Pro-Rata Over-Subscription:** Automated fund return dispatchers and pro-rata token allocation.
- **Continuous Second-by-Second Linear Vesting:** Streaming ERC-20 token vesting vaults with instant claim widgets.

---

## 17. SEBI SCORES 2.0 & Regulatory Grievance Gateway
- **Statutory Complaint Ingestion:** Enterprise integration with SEBI SCORES 2.0 API, RBI Ombudsman CMS, and SMART ODR portals.
- **21-Day SLA Breach Protection:** Automated escalation triggers and draft Action Taken Report (ATR) workflows.
- **Cryptographic Audit Proof:** SHA-256 timestamp anchoring of ATR submission receipts to the Besu audit ledger.

---

## 18. Real-Time Auto-Deleveraging (ADL)
- **Exchange Solvency Protection:** Deterministic deleveraging of highest-profit, highest-leverage counterparty positions when the Insurance Fund / SGF is exhausted.
- **Transparent 5-Bar Queue Indicator:** Real-time visibility into account ADL ranking for active positions.
- **Zero Clawback on Spot:** Unleveraged spot traders are mathematically exempt from ADL haircuts.


