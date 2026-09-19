# Spot Trading Exchange Master Blueprint

**Document Reference:** GROWWW-SPEC-SPOT-001  
**Classification:** Institutional Architecture Specification  
**Status:** Approved for Implementation  
**Target Ledger:** Hyperledger Besu (QBFT Consensus)  
**Target Engine:** In-Memory Rust Matching Engine (Sub-10 Microsecond CLOB)  
**Version:** 1.0.0-PROD  

---

## 1. Executive Vision and System Boundaries

### 1.1 Strategic Vision
The Growww Spot Exchange architecture unifies regulated Indian Capital Markets (equities, mutual funds, tokenized government securities, and real-world assets) with institutional-grade digital assets (cryptocurrencies) into a single, high-throughput, low-latency trading ecosystem.

Rather than treating spot trading on blockchain as an experimental sidecar, this blueprint establishes blockchain-based atomic Delivery-versus-Payment (DvP) as the primary immediate product and sovereign settlement foundation. Every tokenized instrument traded on the platform represents a strict 1:1 real asset backed in segregated, regulated institutional custody:
- Indian equities, exchange-traded funds (ETFs), and sovereign debt are backed 1:1 by physical shares and certificates immobilized in depository pool accounts with National Securities Depository Limited (NSDL) and Central Depository Services (India) Limited (CDSL).
- Digital assets (BTC, ETH, SOL, regulated stablecoins) are backed 1:1 in multi-party computation (MPC) cold-storage vaults with zero rehypothecation and mathematically verifiable zero-knowledge proofs of solvency.

The architecture provides equal institutional access through two synchronized environments:
1. Live Market Demo Paper Trading: Allows retail and institutional users to practice, backtest, and simulate strategies against live tick data with zero capital risk.
2. Real-Money Mainnet Trading: Live capital execution settling atomically on-chain with deterministic finality, governed by a universal 0.00% (Zero Fee) platform fee model.

```
+----------------------------------------------------------------------------------------------------+
|                                    GROWWW SPOT EXCHANGE ECOSYSTEM                                  |
|                                                                                                    |
|   +--------------------------------------------------------------------------------------------+   |
|   |                               CLIENT INTERFACES & GATEWAYS                                 |   |
|   |   Web (React)  |  Mobile (Flutter)  |  Institutional FIX 4.4 / 5.0  |  REST / WebSocket API    |   |
|   +--------------------------------------------------------------------------------------------+   |
|                                                  |                                                 |
|                                                  v                                                 |
|   +--------------------------------------------------------------------------------------------+   |
|   |                            API GATEWAY & PRE-TRADE RISK ENGINE                             |   |
|   |   - Identity Verification (KYC/AML / ONCHAINID)                                            |   |
|   |   - Route Selector: Demo Paper Track vs Real-Money Mainnet Track                           |   |
|   |   - Pessimistic Balance Locking & Collateral Verification                                  |   |
|   +--------------------------------------------------------------------------------------------+   |
|                                                  |                                                 |
|                        +-------------------------+-------------------------+                       |
|                        |                                                   |                       |
|                        v                                                   v                       |
|   +---------------------------------------+   +------------------------------------------------+   |
|   |     DEMO EXECUTION TRACK (PAPER)      |   |        REAL-MONEY EXECUTION TRACK (MAINNET)    |   |
|   | - Isolated Demo Ledger (PostgreSQL)   |   | - High-Performance Sequencer (LMAX Disruptor)  |   |
|   | - Simulated Fill Engine (Live Feed)   |   | - Sub-10us Rust Central Limit Order Book (CLOB)|   |
|   | - Zero Capital Risk / Virtual Capital |   | - Transactional Outbox to Apache Kafka Streams |   |
|   +---------------------------------------+   +------------------------------------------------+   |
|                                                                            |                       |
|                                                                            v                       |
|                                               +------------------------------------------------+   |
|                                               |           POST-TRADE & CLEARING ENGINE         |   |
|                                               | - Statutory STT / Stamp / TDS Tax Engine       |   |
|                                               | - fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Splitter (Treasury/SGF/IPF)   |   |
|                                               | - BLS Signature Batch Aggregation Worker       |   |
|                                               +------------------------------------------------+   |
|                                                                            |                       |
|                                                                            v                       |
|                                               +------------------------------------------------+   |
|                                               |      HYPERLEDGER BESU BLOCKCHAIN (QBFT)        |   |
|                                               | - 2.0-Second Deterministic Blocks (Zero Reorg) |   |
|                                               | - ERC-3643 Regulated Security Tokens (Equities)|   |
|                                               | - ERC-20 / Wrapped Tokens (Crypto / Stable)    |   |
|                                               | - Model 1 Atomic Delivery-versus-Payment (DvP) |   |
|                                               +------------------------------------------------+   |
|                                                                |                  |                |
|                                      +-------------------------+                  +----+           |
|                                      |                                                 |           |
|                                      v                                                 v           |
|   +----------------------------------------------------+   +-----------------------------------+   |
|   |          INDIAN DEPOSITORY CUSTODY BRIDGE          |   |      INSTITUTIONAL CRYPTO CUSTODY |   |
|   | - NSDL / CDSL Depository Participant (DP) Pools    |   | - Tier-4 MPC Cold Vaults          |   |
|   | - Pre-Settlement Demat Finality Gating (ADR-0017)  |   | - Zero Rehypothecation Guarantee  |   |
|   | - Corporate Actions Oracle (Dividends / Splits)    |   | - ZK-SNARK Solvency Proofs        |   |
|   +----------------------------------------------------+   +-----------------------------------+   |
+----------------------------------------------------------------------------------------------------+
```

### 1.2 System Boundaries and Entity Segregation
To comply with Indian securities laws and international regulatory frameworks, the architecture maintains strict operational and legal ring-fencing between two operating entities:

1. Growww Technologies India Pvt. Ltd. (Phase 1: Retail Broker and Depository Participant):
   - Licensed under SEBI (Stock Brokers) Regulations, 1992 and SEBI (Depositories and Participants) Regulations, 2018.
   - Responsible for client onboarding, Aadhaar/PAN identity verification, FIU-IND AML screening, fiat banking rails (UPI, IMPS, RTGS, NEFT), and retail client relationship management.
   - Holds physical custody of securities in segregated client depository pool accounts at NSDL and CDSL.
   - Acts as a registered Trading Member and Clearing Member on authorized market venues.

2. National Blockchain Stock Exchange Ltd. (NBSE Ltd. - Phase 2: Recognized Exchange & Clearing Corporation):
   - Established as an independent Financial Market Infrastructure (FMI) entity under Section 4 of the Securities Contracts (Regulation) Act, 1956 and SEBI (Stock Exchanges and Clearing Corporations) Regulations, 2018.
   - Operates the sovereign Hyperledger Besu distributed settlement ledger, public Central Limit Order Book (CLOB), Core Settlement Guarantee Fund (SGF), and Investor Protection Fund (IPF).
   - Enforces non-discriminatory market access; all broker participants (including Growww Retail) connect over standardized FIX 4.4/5.0 and binary protocols with equal latency and routing priority. Zero internal internalization or front-running is structurally possible.

---

## 2. The Spot Trading Blockchain Engine

### 2.1 Consensus Protocol and Block Finality
The settlement infrastructure runs on a permissioned Hyperledger Besu network utilizing Quorum Byzantine Fault Tolerance (QBFT) consensus:

- Block Generation Interval: Fixed at 2.0 seconds.
- Round Change Timeout: 8.0 seconds base timeout with exponential backoff on validator communication faults.
- Fault Tolerance Threshold: N >= 3F + 1 nodes, where N is the total validator count and F is the maximum tolerated Byzantine or unreachable nodes. For an active consortium of 10 validator nodes, the network tolerates up to 3 simultaneous Byzantine or failing nodes without liveness interruption.
- Immediate Deterministic Finality: QBFT provides single-block mathematical finality. Once a block is committed with 2F + 1 validator signatures, the state transition is irreversible. The chain reorganization probability is exactly zero, satisfying statutory settlement finality requirements under SEBI and IFSCA regulations.
- EVM Environment: London Hardfork target (`londonBlock: 0`). Zero base transaction fees are maintained on the network layer through Gas-free Meta Transactions (EIP-2771) with authorized Paymaster contracts, eliminating gas volatility from end-user execution.

### 2.2 Validator Topology and Governance
The validator network is distributed across geographically redundant, jurisdictionally compliant secure hosting zones:

| Node Type | Operator / Stakeholder | Location | Role in Network |
|---|---|---|---|
| Core Validator 1-3 | NBSE Clearing Corporation | Mumbai (Tier-4 Datacenter) | Primary block proposal and state commit |
| Core Validator 4-6 | GIFT City IFSC Trust Infrastructure | Gandhinagar (GIFT IFSC) | International settlement and cross-border routing |
| Institutional Validator 7-8 | Consortia Banking Partners | Hyderabad / Bengaluru | Independent transaction verification and attestations |
| Regulatory Watcher 1 | SEBI Regulatory Node | Mumbai | Real-time audit, read-only state, surveillance mirror |
| Regulatory Watcher 2 | FIU-IND AML Node | New Delhi | Real-time transaction monitoring and flag verification |

Validator consensus keys are safeguarded in FIPS 140-2 Level 3 Hardware Security Modules (HSMs). Validator rotation and membership modifications require an on-chain supermajority (75% affirmative votes) from active governance nodes.

### 2.3 Atomic Delivery-versus-Payment (DvP) Architecture
The exchange operates Model 1 DvP settlement (simultaneous gross transfer of both securities/tokens and funds):

```
+----------------------------------------------------------------------------------------------------+
|                                    ATOMIC DVP SETTLEMENT TRANSACTION                               |
|                                                                                                    |
|  [EVM Transaction Batch: DvPSettlementEngine.executeBatchDvP()]                                    |
|                                                                                                    |
|  +-------------------------------------+          +-------------------------------------+          |
|  |             LEG 1: ASSET            |          |             LEG 2: PAYMENT          |          |
|  |                                     |          |                                     |          |
|  | Seller Vault / Custody              |          | Buyer Escrow / Wallet               |          |
|  |  (ERC-3643 Security Token / Crypto) |  ATOMIC  |  (Regulated INR Token / Stablecoin) |          |
|  |                 |                   | <======> |                 |                   |          |
|  |                 v                   |  BALANCE |                 v                   |          |
|  | Buyer Vault / Account               |  MUTATION| Seller Cash Balance                 |          |
|  |  (Demat Mirror / Crypto Wallet)     |          |  (INR / Stablecoin Credit)          |          |
|  +-------------------------------------+          +-------------------------------------+          |
|                                            |                                                       |
|                                            v                                                       |
|  +----------------------------------------------------------------------------------------------+  |
|  |                             LEG 3: UNIVERSAL 0.00% fee (No fee at all) DEDUCTION                             |  |
|  |  0.00% fee at launch (0.0000 notional) governed by FeeController.sol:                      |  |
|  |    -> 60% of collected fees to Platform Treasury (0 at launch)                                           |  |
|  |    -> 25% of collected fees to Settlement Guarantee Fund (SGF) Contract (0 at launch)                     |  |
|  |    -> 15% of collected fees to Investor Protection Fund (IPF) Contract (0 at launch)                      |  |
|  +----------------------------------------------------------------------------------------------+  |
+----------------------------------------------------------------------------------------------------+
```

All three legs occur within a single EVM state transition. If any leg fails (such as compliance check rejection, insufficient balance, or transfer lock), the entire transaction reverts, ensuring no party is exposed to principal risk.

### 2.4 Token Standards and State Storage Architecture
The platform deploys two primary smart contract standards on Hyperledger Besu:

1. Regulated Equities and Securities: ERC-3643 (Permissioned Security Tokens)
   - Identity Registry: Every token holder must link to an on-chain Identity Contract (ONCHAINID) verifying PAN, Aadhaar, KYC status, and investor accreditation.
   - Compliance Engine: Direct transfer restrictions enforce trading window closures, insider trading restrictions, daily price band limits, and foreign portfolio investment (FPI) holding ceilings.
   - Controlled Mint/Burn: Shares immobilized in the depository pool account trigger token minting to the user's verified address; withdrawal to Demat triggers token burning.

2. Digital Assets and Settlement Currencies: ERC-20 with Enterprise Extensions
   - Supported assets: Native wrapped tokens (wBTC, wETH, wSOL), regulated fiat settlement stable tokens, and USDT/USDC bridge tokens.
   - Blacklist and Pausing: Emergency circuit breaker functionality controlled by the independent Risk Committee multisig.

State Storage Engine: Besu Bonsai Trie architecture with rolling pruning. This design guarantees constant-time state reads, prevents disk bloat, and provides deterministic microsecond Merkelized state verification under continuous trading load.

---

## 3. Hybrid Indian Capital Market + Crypto Architecture

### 3.1 Unified Multi-Asset Catalog
The exchange engine handles Indian domestic capital instruments and global digital assets within a standardized contract abstraction:

```
+----------------------------------------------------------------------------------------------------+
|                                    UNIFIED ASSET CLASS TOPOLOGY                                    |
|                                                                                                    |
|  +-----------------------------------+  +-----------------------------------+  +----------------+  |
|  |      INDIAN CASH EQUITIES & ETFS  |  |       TOKENIZED FIXED INCOME      |  | DIGITAL ASSETS |  |
|  |                                   |  |                                   |  |    (CRYPTO)    |  |
|  |  * Nifty 50 Top Equities          |  |  * Sovereign Green Bonds          |  |  * Bitcoin     |  |
|  |  * Bharat 22 ETF                  |  |  * 91-Day & 364-Day T-Bills       |  |  * Ethereum    |  |
|  |  * Gold BeES Commodity ETF        |  |  * Corporate AAA Commercial Paper |  |  * Solana      |  |
|  |                                   |  |                                   |  |  * Stablecoins |  |
|  |  Custody: CDSL / NSDL Demat Pool  |  |  Custody: RBI Depository / NSDL   |  |  Custody: MPC  |  |
|  |  Token Standard: ERC-3643         |  |  Token Standard: ERC-3643         |  |  Standard:     |  |
|  |  Settlement: INR Fiat / Demat DvP |  |  Settlement: INR Fiat / Demat DvP |  |  ERC-20 / Native|  |
|  +-----------------------------------+  +-----------------------------------+  +----------------+  |
+----------------------------------------------------------------------------------------------------+
```

### 3.2 Regulatory Compliance and Dual Legal Gateways
Trading operations are partitioned into regulatory routing domains governed by applicable legal mandates:

| Dimension | Indian Capital Markets Track | Digital Assets (Crypto / VDA) Track |
|---|---|---|
| Regulatory Jurisdiction | SEBI, Reserve Bank of India (RBI), MCA | FIU-IND (PMLA), Ministry of Finance |
| Tax Deduction at Source (TDS) | Section 194 (Dividends), Section 194-IA | Section 194S (1.0% TDS on gross crypto sells) |
| Transaction Taxes | Securities Transaction Tax (STT), Stamp Duty | No STT; 18% GST on platform commission |
| AML & Sanctions Screening | C-KYC, Aadhaar e-KYC, SEBI Debarred Entities | FIU-IND Suspicious Transaction Reports, Chainalysis |
| Regulatory Reporting Cadence | Real-time to Trade Repository, End-of-Day (EOD) | Monthly VDA reporting, Form 61A, Suspicious Activity |

### 3.3 Demat Custody Bridge Architecture (ADR-0017)
To ensure strict 1:1 asset backing without fractional reserve exposure, the Demat Custody Bridge operates on pre-settlement finality gating:

```
+----------------------------------------------------------------------------------------------------+
|                                    DEMAT CUSTODY SYNCHRONIZATION                                   |
|                                                                                                    |
|  [Step 1: Physical Share Ingress]                                                                  |
|   User instructs transfer of 100 shares of RELIANCE from Personal Demat to Growww DP Pool Account. |
|                                                  |                                                 |
|                                                  v                                                 |
|  [Step 2: Depository Gateway Verification]                                                         |
|   NSDL/CDSL Settlement Confirmation Slip received by DP Gateway worker via ISO 20022 message.       |
|                                                  |                                                 |
|                                                  v                                                 |
|  [Step 3: Pre-Settlement Gating & Token Minting]                                                   |
|   Depository Bridge locks shares in Segregated Pool Vault -> Triggers on-chain minting of        |
|   100.00000000 e-RELIANCE (ERC-3643) to User ONCHAINID address on Hyperledger Besu.                |
|                                                  |                                                 |
|                                                  v                                                 |
|  [Step 4: Continuous Reconciliation Invariant]                                                    |
|   Invariant: Total Tokenized Supply == Total Physical Shares in DP Pool (Reconciled every 60s).    |
|                                                  |                                                 |
|                                                  v                                                 |
|  [Step 5: Share Egress / Redemption]                                                               |
|   User requests redemption -> 100 e-RELIANCE burned on-chain -> DP Gateway executes NSDL/CDSL      |
|   delivery out instruction to User's beneficiary Demat account within T+1 statutory settlement.    |
+----------------------------------------------------------------------------------------------------+
```

### 3.4 Corporate Actions Propagation
Corporate actions declared by issuers (cash dividends, stock splits, bonus shares, rights issues) are processed through an automated oracle pipeline:
1. Depository Notification: The corporate action record date and entitlement ratio are ingested from exchange feeds.
2. Blockchain Snapshot: At the record date block height, an automated state snapshot captures all on-chain token holders and their exact balances.
3. Distribution Execution:
   - Cash Dividends: Credited directly in INR to the user's primary bank account registered via penny-drop verification.
   - Bonus / Split Shares: Directly minted pro-rata to token holders upon confirmation of depository credit to the pool account.

---

## 4. Dual-Track Execution: Live Market Demo vs Real-Money Mainnet

### 4.1 Architectural Separation and Isolation
The Growww Spot Exchange guarantees complete operational and financial isolation between the Live Market Demo Paper Trading track and the Real-Money Mainnet Trading track.

```
+----------------------------------------------------------------------------------------------------+
|                                   DUAL-TRACK EXECUTION TOPOLOGY                                    |
|                                                                                                    |
|                                 +-------------------------------+                                  |
|                                 |   INCOMING CLIENT API REQUEST |                                  |
|                                 |  Header: X-Execution-Track    |                                  |
|                                 +-------------------------------+                                  |
|                                                 |                                                  |
|                        +------------------------+------------------------+                         |
|                        |                                                 |                         |
|             [Track == "DEMO"]                                 [Track == "REAL"]                    |
|                        |                                                 |                         |
|                        v                                                 v                         |
|   +------------------------------------------+      +------------------------------------------+   |
|   |         DEMO PAPER TRADING TRACK         |      |        REAL-MONEY MAINNET TRACK          |   |
|   |                                          |      |                                          |   |
|   |  * Virtual Account Balance Engine        |      |  * Production Ledger Balance Locking     |   |
|   |  * PostgreSQL Demo State Database        |      |  * PostgreSQL Double-Entry Ledger        |   |
|   |  * Realistic Slippage & Queue Model      |      |  * LMAX Disruptor Sequencer (NUMA-pinned)|   |
|   |  * Subscribed to Live Market L2 Data     |      |  * Rust In-Memory CLOB Matching Engine   |   |
|   |  * Virtual 0.00% fee (No fee at all) Simulation          |      |  * Real-Money 0.00% fee (No fee at all) Deduction        |   |
|   |  * Zero Blockchain Transaction Dispatch  |      |  * Hyperledger Besu Atomic DvP Execution |   |
|   |  * Dedicated Demo API Rate Limits        |      |  * Statutory Tax & Clearing Reporting    |   |
|   +------------------------------------------+      +------------------------------------------+   |
+----------------------------------------------------------------------------------------------------+
```

### 4.2 Realistic Simulation Engine for Demo Trading
Unlike simplistic paper trading systems that assume instant fills at the top-of-book price, the Growww Demo Engine enforces realistic market mechanics:
- Queue Priority Emulation: Simulated limit orders are placed into a virtual depth queue. Orders are not filled until the market trades through their limit price or sufficient real volume transacts at the limit price.
- Market Impact and Partial Fills: Large market orders consume virtual liquidity across multiple depth levels, reflecting true slippage.
- Rejection Simulation: Orders submitted outside exchange price bands (operating range) or during market halts are rejected with authentic error codes.
- Virtual Capital Sandbox: Every demo user receives a virtual allocation of INR 1,000,000 and 1.00000000 BTC/ETH for testing, reset-able through user controls.

### 4.3 Interface Parity and Wire Format Consistency
Both tracks utilize identical Protobuf and JSON wire schemas across REST and WebSocket connections. A single flag distinguishes the session:
- HTTP Header: `X-Execution-Track: REAL` or `X-Execution-Track: DEMO`
- WebSocket Connection Parameter: `wss://api.growww.internal/v1/stream?track=demo` vs `wss://api.growww.internal/v1/stream?track=real`

This uniformity allows institutional clients to develop, backtest, and qualify automated trading bots in the demo sandbox, then switch to real-money trading simply by updating credentials and the track header, with zero code modifications.

---

## 5. Complete Spot Order Lifecycle

### 5.1 Supported Order Types and Execution Semantics
The exchange matching engine supports four primary spot order types:

1. Market Order:
   - Aggressive order executed immediately against resting liquidity in the book.
   - Time-in-Force (TIF): Immediate-or-Cancel (IOC) or Fill-or-Kill (FOK).
   - Protective Collar: Default 2.0% maximum slippage protection against the Last Traded Price (LTP) to protect against fat-finger market orders in thin books.

2. Limit Order:
   - Passive or crossing order specifying maximum purchase price or minimum sale price.
   - Supported TIF: Good-Til-Cancelled (GTC), Immediate-or-Cancel (IOC), Fill-or-Kill (FOK), and Good-Til-Date (GTD).
   - Price-Time Priority: Orders are queued chronologically at each price tick. Quantity reductions maintain book priority; price modifications or quantity increases cause the order to lose priority and re-queue at the tail.

3. Stop-Limit Order:
   - Conditional order containing two price parameters: Stop Price (trigger) and Limit Price.
   - Quiescent State: Stored in an off-book Stop Order Registry in memory. Not visible in the public L2/L3 order book.
   - Activation Trigger: When the Last Traded Price (or dual Mark Price for high volatility assets) reaches or crosses the Stop Price, the order activates and injects into the active matching book as a standard Limit Order.

4. Iceberg Order:
   - Algorithmic slicing order for large volume execution without revealing full market size.
   - Parameters: Total Quantity (`Q_tot`) and Display Quantity (`Q_disp`).
   - Execution Mechanics: Only `Q_disp` is entered into the public order book at the designated limit price. Upon complete fill of `Q_disp`, the engine automatically reloads the next slice of `Q_disp` from the hidden remaining quantity (`Q_rem`). Each newly reloaded slice receives a fresh timestamp and joins the tail of the priority queue at that price level.

### 5.2 Step-by-Step Order Lifecycle: Ingress to Finality

```
[Client App]             [API Gateway]           [Matching Engine]         [Outbox / Kafka]         [Besu Blockchain]
     |                         |                         |                        |                         |
     |-- 1. Submit Order ----->|                         |                        |                         |
     |   (mTLS / Signature)    |-- 2. Validate Risk ---->|                        |                         |
     |                         |   (Lock Collateral)     |                        |                         |
     |                         |                         |                        |                         |
     |                         |-- 3. Sequencer Queue -->|                        |                         |
     |                         |   (SPSC Ring Buffer)    |-- 4. Match Order       |                         |
     |                         |                         |   (Sub-10us CLOB)      |                         |
     |                         |                         |           |            |                         |
     |                         |<-- 5. Execution Report -|           |            |                         |
     |<-- 6. WS Order Update --|   (Trade Event)         |           |            |                         |
     |                         |                         |           v            |                         |
     |                         |                         |-- 7. Emit Trade ------>|                         |
     |                         |                         |                        |-- 8. Batch DvP Worker ->|
     |                         |                         |                        |      (Aggregate BLS)    |-- 9. Mine Block
     |                         |                         |                        |                         |      (2s QBFT)
     |                         |                         |                        |<-- 10. Receipt Emit ----|
     |<-- 11. Final Settlement Notification (On-Chain Confirmed) -------------------------------------------|
```

Detailed Phase Breakdown:
- Phase 1: Ingress and Authentication: Client submits signed payload. Gateway checks HMAC/ECDSA signature, nonces, rate limits, and verifies identity status against local cache.
- Phase 2: Pre-Trade Risk Verification: Risk engine verifies available balance. For buy orders, notional cost plus max fee is locked; for sell orders, underlying asset balance is locked.
- Phase 3: Sequencer Ingress: Order enters single-producer single-consumer (SPSC) Disruptor ring buffer on a CPU core pinned via NUMA architecture, assigning a monotonically increasing 64-bit sequence ID.
- Phase 4: Matching Execution: In-memory Rust matching engine processes order against active book in sub-10 microseconds, producing match events or inserting resting limit orders.
- Phase 5: Event Emission: Match events emit simultaneously to the WebSocket gateway (for immediate UI update) and the transactional outbox table via Apache Kafka.
- Phase 6: On-Chain Batch DvP Settlement: Post-trade worker aggregates trades into 2.0-second settlement batches, calculates exact fee allocations, and submits an atomic `executeBatchDvP` transaction to Hyperledger Besu.
- Phase 7: Deterministic Finality: Besu validators commit the block within 2.0 seconds. Event listeners verify transaction receipt and release pessimistic balance locks into finalized ledger state.

### 5.3 Order State Transition Diagram

```
                              +-------------------+
                              |    PENDING_NEW    |
                              +-------------------+
                                        |
                   +--------------------+--------------------+
                   | Risk / Sequence OK                      | Risk Check Failed
                   v                                         v
          +-----------------+                       +-----------------+
          |       NEW       |                       |    REJECTED     |
          +-----------------+                       +-----------------+
                   |
     +-------------+-------------+
     | Partial Fill              | Complete Fill
     v                           v
+--------------------+   +-------------------+
|  PARTIALLY_FILLED  |   |      FILLED       |
+--------------------+   +-------------------+
     |              |              ^
     | Partial Fill | Fill Rem     |
     +--------------+--------------+
     |
     | User Cancel / Expire
     v
+--------------------+
|     CANCELLED      |
+--------------------+
```

---

## 6. Single Universal 0.00% (No fee at all) Platform Fee Waterfall

### 6.1 Fee Philosophy and Structure
The Growww Spot Exchange enforces a single, transparent, flat platform fee across all trading pairs and client tiers:
- Fee Rate: Strictly 0.00% (Zero Fee) of gross notional trade value (0.0000 at launch, governed dynamically by FeeController.sol).
- Symmetry: Applied identically to both Maker (liquidity provider) and Taker (liquidity remover) sides, eliminating complex tier structures and rebate gaming.
- Calculation Formula:
  `PlatformFee = TradePrice * TradeQuantity * 0.0000` (0.00 at launch; governed by FeeController.sol)

### 6.2 Fee Distribution Waterfall
Every unit of collected platform fee is partitioned at the moment of on-chain clearing according to a strict three-tier allocation waterfall:

```
+----------------------------------------------------------------------------------------------------+
|                               0.00% (Zero Fee) PLATFORM FEE WATERFALL ALLOCATION                              |
|                                                                                                    |
|                               +-------------------------------+                                    |
|                               | TOTAL COLLECTED PLATFORM FEE  |                                    |
|                               |   0.0000% of Trade Notional   |                                    |
|                               +-------------------------------+                                    |
|                                               |                                                    |
|                   +---------------------------+---------------------------+                        |
|                   |                                                       |                        |
|                   v                                                       v                        |
|   +-------------------------------+                       +-------------------------------+        |
|   |  PLATFORM TREASURY (GOV)      |                       |  RISK & GUARANTEE RESERVES    |        |
|   |  0.0000% at launch (60%)      |                       |  0.0000% at launch (40%)      |        |
|   |                               |                       |                               |        |
|   |  * Exchange Operations        |                       +---------------+---------------+        |
|   |  * Infrastructure & Hardware  |                                       |                        |
|   |  * Node Validation Expenses   |                   +-------------------+--------------------+   |
|   |  * Core Engineering & R&D     |                   |                                        |   |
|   +-------------------------------+                   v                                        v   |
|                                       +-------------------------------+  +---------------------+   |
|                                       | CORE SGF BUFFER               |  | INVESTOR PROT       |   |
|                                       | 0.0000% at launch (25%)       |  |     PROTECTION FUND |   |
|                                       |                               |  | 0.0000% at launch (15%) |   |
|                                       | * Settlement Default Reserve  |  |                     |   |
|                                       | * Multi-Member Loss Waterfall |  | * Fraud Protection  |   |
|                                       | * SEBI SECC Mandatory Corpus  |  | * Insolvency Cover  |   |
|                                       +-------------------------------+  +---------------------+   |
+----------------------------------------------------------------------------------------------------+
```

1. 60% of activated fees (0.0000% at launch) to Platform Treasury:
   - Dedicated to operational maintenance, Tier-4 datacenter co-location, cloud routing infrastructure, Besu validator hardware costs, and continuous protocol research.
2. 25% of activated fees (0.0000% at launch) to Core Settlement Guarantee Fund (SGF):
   - Held in a segregated on-chain smart contract vault governed under SEBI SECC Regulations and ADR-0013.
   - Forms the clearing house first-loss absorption buffer to guarantee completion of settlements in the event of participant default.
3. 15% of activated fees (0.0000% at launch) to Investor Protection Fund (IPF):
   - Held in an autonomous trust fund administered by independent public interest trustees.
   - Provides compensatory relief to retail investors in instances of member insolvency, verified technical faults, or custodial bridge interruptions.

### 6.3 Statutory and Regulatory Taxes (Separation of Concerns)
Statutory government levies are calculated separately from the platform fee and are never mixed with platform revenues:
- Securities Transaction Tax (STT): Levied on delivery equity purchases and sales (0.1% each side) pursuant to the Finance Act.
- Stamp Duty: 0.015% on delivery purchases as mandated by the Indian Stamp (Collection and Enforcement) Rules.
- Goods and Services Tax (GST): 18.0% GST applied strictly to the platform's 0.00% (Zero Fee) commission (not on the gross trade value).
- VDA Tax Deduction at Source (TDS): 1.0% under Section 194S of the Income-tax Act, 1961, deducted on all gross digital asset sale proceeds and remitted to the Income Tax Department.

---

## 7. Security, Invariants and Verification Gates

### 7.1 Core Mathematical and Financial Invariants
The exchange runtime enforces four fundamental mathematical invariants. Any breach of an invariant halts trading immediately:

1. The Solvency Invariant (Zero Fractional Reserve):
   `VaultedAssets(t) >= sum(UserLedgerBalances(t)) for all assets at all times t.`
   The total quantity of any asset held in physical custody (Demat pool or MPC cold vault) must equal or exceed the total aggregate balance owed to users across both on-chain smart contracts and internal ledgers.

2. The Double-Entry Conservation Invariant:
   `sum(Debits) == sum(Credits) for every transaction record.`
   No balance can be created or destroyed. Every ledger movement requires balanced debits and credits across verified system accounts.

3. The Non-Negative Balance Invariant:
   `UserBalance(u, a, t) >= 0 for all users u, assets a, and timestamps t.`
   Spot trading does not permit naked short selling or uncollateralized overdrafts. System constraints reject any transaction that would result in a negative balance.

4. The Price-Time Priority Invariant:
   `For any two orders O1 and O2 where Price(O1) == Price(O2), if SequenceTimestamp(O1) < SequenceTimestamp(O2), then O1 MUST be matched before O2.`

### 7.2 Verification Gates and Solvency Proofs
The exchange integrates multi-layered verification gates across hardware, software, and cryptographic proofs:

```
+----------------------------------------------------------------------------------------------------+
|                                  CONTINUOUS VERIFICATION GATES                                     |
|                                                                                                    |
|  [GATE 1: Pre-Trade Balance Lock]                                                                  |
|   Database pessimistic row-level lock ensures funds cannot be double-spent across parallel orders.  |
|                                                  |                                                 |
|                                                  v                                                 |
|  [GATE 2: Pre-Settlement Demat Check]                                                              |
|   Zero-naked-shorting validation verifies physical share immobilization before sell book ingress.   |
|                                                  |                                                 |
|                                                  v                                                 |
|  [GATE 3: On-Chain Smart Contract Bounds]                                                          |
|   Smart contracts enforce require(balance >= amount) and use SafeERC20 / checked arithmetic.       |
|                                                  |                                                 |
|                                                  v                                                 |
|  [GATE 4: Hourly ZK-SNARK Solvency Attestation]                                                    |
|   Automated prover generates Merkle Sum Tree liabilities proof + MPC custody asset attestation.    |
|   Users verify their account inclusion anonymously without revealing portfolio balances.          |
+----------------------------------------------------------------------------------------------------+
```

- ZK-SNARK Merkle Proof of Solvency (ADR-0037): Hourly cryptographic commitments publish Merkle roots of the complete user liability tree to the Besu ledger. Users can generate client-side cryptographic inclusion proofs confirming their balances are fully backed without revealing financial details to third parties.
- Pessimistic Locking and Idempotency (ADR-0012): All balance mutations execute inside serialized database transactions with unique idempotency keys, preventing race conditions or duplicate execution during network retries.

### 7.3 Infrastructure Security and Operational Safeguards
- Custody Security: 100% of cold crypto assets reside in MPC multi-party signature vaults requiring a 3-of-5 threshold signing scheme across air-gapped locations.
- Network Micro-Segmentation: The Rust matching engine, Kafka messaging cluster, and Besu validator nodes operate in private subnets with mutual TLS (mTLS) 1.3 encryption and zero public internet exposure.
- Automated Surveillance and Circuit Breakers: Real-time pattern recognition detects wash trading, spoofing, and layering. Volatility circuit breakers automatically pause trading for a cooling-off period if price movements exceed 10% within a rolling 5-minute window.
