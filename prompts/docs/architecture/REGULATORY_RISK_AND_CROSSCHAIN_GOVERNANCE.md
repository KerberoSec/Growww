# Regulatory Risk, Compliance & Cross-Chain Governance Specification

**Document Version:** 2.0.0-PROD-SPEC  
**Status:** Approved Architecture Standard  
**Owner:** Regulatory Affairs, Risk Modeling & Cross-Border Governance Group  
**Classification:** Institutional Standard  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  
**Related Specifications:**
- [Financial Domain Specifications & Trading Venue Architecture](./FINANCIAL_AND_DOMAIN_SPECIFICATIONS.md)
- [Growww Core Blockchain Ledger & Web3 Architecture Specification](./BLOCKCHAIN_LEDGER_CORE_SPECIFICATION.md)
- [End-to-End Asset Lifecycle, Failure Modes & Resilience Framework](./END_TO_END_ASSET_LIFECYCLE_AND_RESILIENCE_FRAMEWORK.md)
- [Event-Driven Architecture Specification](./EVENT_DRIVEN_ARCHITECTURE_SPECIFICATION.md)
- [NBSE Master Architecture & Integration Blueprint](./NBSE_MASTER_ARCHITECTURE_AND_INTEGRATION_BLUEPRINT.md)
- [Runbook 07: Cross-Chain Reorganization & Clawback Saga Execution](../runbooks/RUNBOOK-07-crosschain-reorg-and-clawback.md)
- [Runbook 08: SEBI Peak Margin Snapshot & Overnight Collateral Management](../runbooks/RUNBOOK-08-peak-margin-snapshot-and-overnight-collateral.md)

---

## 1. Executive Architectural Summary & Statutory Mandate

The Growww platform operates a hybrid financial market infrastructure uniting continuous 24/7 Web3 trading mechanisms with strict statutory compliance across domestic Indian capital markets and the International Financial Services Centre (GIFT City). To maintain this posture, the architecture enforces mathematical invariants, institutional custody backing, and multi-regulatory reporting pipelines.

```
+---------------------------------------------------------------------------------------------------+
|                         MULTI-JURISDICTIONAL REGULATORY SUPERVISION CORE                          |
|                                                                                                   |
|    +--------------------------------------+       +-----------------------------------------+     |
|    |      DOMESTIC INDIAN OPERATING       |       |       INTERNATIONAL GATEWAY ENTITY      |     |
|    |           ENTITY (ONSHORE)           |       |           (OFFSHORE - GIFT CITY)        |     |
|    +--------------------------------------+       +-----------------------------------------+     |
|    | Regulators: SEBI, RBI, MCA, FIU-IND  |       | Regulators: IFSCA, Global FATF AML      |     |
|    | Base Currency: Indian Rupee (INR)    |       | Base Currency: US Dollar (USD / USDC)   |     |
|    | Depository: NSDL / CDSL (Demat Pool) |       | Custody: Multi-Sig Cold Vaults & Prime  |     |
|    | Legal Basis: Companies Act 2013,     |       | Legal Basis: IFSCA Act 2019, IFSCA      |     |
|    |   SEBI Master Circulars, DPDP 2023   |       |   Capital Market Regulations 2021       |     |
|    +--------------------------------------+       +-----------------------------------------+     |
|                       \                                       /                                   |
|                        \                                     /                                    |
|                         v                                   v                                     |
|    +-----------------------------------------------------------------------------------------+    |
|    |                  CROSS-ENTITY REGULATORY GOVERNANCE ENGINE & GATEWAYS                   |    |
|    |                                                                                         |    |
|    |  [REG-01] SEBI Peak Margin Partitioning   <--->  24/7 Trading & Collateral Quarantine   |    |
|    |  [REG-02] Fractional Proxy Voting Portal  <--->  Democratized Beneficial Ownership      |    |
|    |  [REG-03] Tiered Finality & Reorg Matrix  <--->  Solana / Bitcoin / Ethereum Gateways   |    |
|    +-----------------------------------------------------------------------------------------+    |
|                                             |                                                     |
|                                             v                                                     |
|    +-----------------------------------------------------------------------------------------+    |
|    |         HYPERLEDGER BESU ENTERPRISE CONSORTIUM LEDGER (QBFT CONSENSUS, 2.0s)            |    |
|    |   - Instant Deterministic Finality (0 Forks / 0 Reorgs)                                 |    |
|    |   - ERC-3643 Permissioned Security Tokens & Zero On-Chain PII                           |    |
|    |   - Immutable Merkle State Roots for Regulatory Filings & Proof of Reserve              |    |
|    +-----------------------------------------------------------------------------------------+    |
+---------------------------------------------------------------------------------------------------+
```

### 1.1 Core Regulatory Governance Invariants

1. **INV-REG-1 (Strict Two-Entity Separation):** Domestic onshore accounts, collateral pools, and order books operating under SEBI and RBI jurisdiction are strictly quarantined from International Financial Services Centre (GIFT City) accounts under IFSCA. Collateral, margin lines, and liquidity pools never commingle across legal borders.
2. **INV-REG-2 (Zero Unbacked Fractional Securities):** Every fractional security unit (down to $10^{-6}$ micro-shares) mapped on the internal ledger or tokenized via ERC-3643 smart contracts corresponds strictly 1:1 to physical equity shares custodied in designated demat pool accounts with NSDL or CDSL. Fractional units enjoy full pass-through economic and corporate governance rights.
3. **INV-REG-3 (Continuous Margin Solvency):** At no point during 24/7 market operation may a domestic account run an uncollateralized position. Leverage is restricted to SEBI-approved market hours with pre-funded capital rules governing all after-hours and weekend trading.
4. **INV-REG-4 (Economic Finality Pre-Credit):** Ingestion of assets from external public blockchains (Bitcoin, Ethereum, Solana) guarantees that zero internal ledger purchasing power or synthetic minting is granted until transactions reach mathematically and economically irreversible finality.

### 1.2 Governance Scope Matrix

| Identifier | Subsystem | Target Jurisdiction | Primary Regulators | Architectural Core |
|---|---|---|---|---|
| **REG-01** | Peak Margin Partitioning | Domestic Equity & F&O | SEBI, Clearing Corporations | Dynamic Session Partitioning, 4-Snapshot CSPRNG, Pre-emptive Liquidation State Machine |
| **REG-02** | Fractional Proxy Voting | Corporate Actions & Equity | MCA, SEBI, NSDL, CDSL | EIP-712 Voting Ballots, Hamilton-Hare Integer Apportionment, Depository Adapter |
| **REG-03** | Tiered Cross-Chain Finality | Cross-Border Liquidity | IFSCA, Global Gateways | Multi-Stage Confirmation Thresholds, Watchdog DAG Trackers, 5-Phase Clawback Saga |

---

## 2. SEBI Peak Margin Snapshot Compliance in 24/7 Markets (REG-01)

### 2.1 Statutory Context & Regulatory Problem Statement

Under the Securities and Exchange Board of India (SEBI) Master Circulars on Risk Management (notably circulars `SEBI/HO/MRD2/DCAP/CIR/P/2020/127` and `SEBI/HO/MRD2/DCAP/P/CIR/2021/0598`), clearing members and stockbrokers are legally required to collect upfront initial margins from clients. 

To enforce this throughout the trading day, clearing corporations (CCs) execute **four random snapshot windows** during standard exchange trading hours (09:15 to 15:30 IST). The highest margin requirement across these four windows is defined as the client's **Peak Margin** for the day. If the collected collateral at that snapshot window is less than the peak margin requirement, a statutory short-collection penalty is imposed:

$$\text{Shortfall}(T_k) = \max\left(0, M_{\text{req}}(T_k) - C_{\text{eff}}(T_k)\right)$$

Operating a **24/7 continuous trading venue** introduces severe regulatory and operational challenges under this framework:
1. **Banking Rail Asymmetry:** RTGS/NEFT interbank transfers and banking clearing windows experience weekend and holiday maintenance halts, preventing clients from topping up collateral dynamically outside traditional banking hours.
2. **Valuation Volatility:** Crypto and tokenized global assets fluctuate continuously over weekends, creating risk of phantom margin shortfalls against domestic equities if collateral pools are commingled.
3. **Regulatory Penalty Exposure:** If margin credit were extended 24/7, an adverse market movement at 03:00 IST on a Sunday could trigger a catastrophic margin shortfall penalty during Monday morning's first statutory SEBI snapshot (09:15-10:45 IST), before banking rails open.

### 2.2 Dual-Session Architectural Partitioning Model

To guarantee 100% statutory compliance with SEBI peak margin rules while enabling uninterrupted 24/7 trading, the platform partitions the market day into two distinct operational sessions governed by separate risk engines:

```
+---------------------------------------------------------------------------------------------------+
|                     24/7 TIME PARTITIONING & COLLATERAL GOVERNANCE MODEL                          |
|                                                                                                   |
|  00:00               09:15                             15:30               23:59:59               |
|    +-------------------+---------------------------------+--------------------+                   |
|    |  EXTENDED SESSION |     PRIMARY STATUTORY SESSION   |  EXTENDED SESSION  |                   |
|    |   (PRE-FUNDED)    |     (SPAN + INTRADAY LEVERAGE)  |   (PRE-FUNDED)     |                   |
|    +-------------------+---------------------------------+--------------------+                   |
|              ^                           ^                           ^                            |
|              |                           |                           |                            |
|      No Margin Credit            4 Random Snapshots          No Margin Credit                     |
|      100% Cash / eINR            Window 1: 09:15 - 10:45     100% Cash / eINR                     |
|      Zero Shortfall Risk         Window 2: 10:45 - 12:15     Zero Shortfall Risk                  |
|                                  Window 3: 12:15 - 13:45                                          |
|                                  Window 4: 13:45 - 15:30                                          |
+---------------------------------------------------------------------------------------------------+
```

#### Session A: Primary Statutory Market Session (09:15:00 to 15:30:00 IST)
- **Applicability:** Monday through Friday, excluding statutory clearing holidays declared by NSE/BSE.
- **Risk Policy:** Standard Standard Portfolio Analysis of Risk (SPAN) combined with Extreme Loss Margin (ELM) and Value-at-Risk (VaR) requirements.
- **Credit Lines:** Intraday leverage is permitted against approved, pledged collateral (demat equities, sovereign gold bonds, liquid mutual funds) subject to SEBI-mandated haircut schedules.
- **Snapshot Monitoring:** Real-time computation of client and proprietary positions across the 4 randomized SEBI snapshot intervals.

#### Session B: Extended & 24/7 Weekend Session (15:30:00 to 09:15:00 IST, Weekends & Holidays)
- **Applicability:** Every night, Saturday, Sunday, and statutory exchange holidays.
- **Risk Policy:** **100% Cash Pre-Funding Invariant.** Zero intraday leverage is extended.
- **Permissible Settlement Media:** Available settled INR cash balances, RBI Digital Rupee (eINR) CBDC, or fully paid unencumbered tokenized securities.
- **Collateral Haircut on Volatile Assets:** Pledged non-cash assets cannot be used to open leveraged positions during this window.
- **Transition Buffer Rule:** At 15:15:00 IST (15 minutes prior to Primary Session close), the risk engine executes an automated transition check. If an account has open leveraged positions that are not converted to delivery or 100% cash-backed, the Pre-Emptive Auto-Mitigation Engine systematically initiates square-off to eliminate margin deficit carryover.

### 2.3 Cryptographic CSPRNG Snapshot Schedule Engine

To prevent regulatory gaming and ensure strict compliance with SEBI's mandate for random intraday observation, snapshot timestamps are generated deterministically each morning using a Cryptographically Secure Pseudo-Random Number Generator (CSPRNG).

```
+---------------------------------------------------------------------------------------------------+
|                        CSPRNG SNAPSHOT GENERATION & VERIFICATION PIPELINE                         |
|                                                                                                   |
|   +--------------------------+       +-------------------------+                                  |
|   |  Market Open Besu Block  | ----> | NIST Public Randomness  |                                  |
|   |   Hash at 09:00:00 IST   |       |   Beacon Interop Seed   |                                  |
|   +--------------------------+       +-------------------------+                                  |
|                 \                                 /                                               |
|                  \                               /                                                |
|                   v                             v                                                 |
|   +-----------------------------------------------------------------------------------------+     |
|   |                HMAC-SHA256 DETERMINISTIC SEED EXPANDER (FIPS 140-2 HSM)                 |     |
|   |   Seed = HMAC_SHA256(Key_Compliance, BlockHash_0900 || DateStamp || Nonce)             |     |
|   +-----------------------------------------------------------------------------------------+     |
|                                             |                                                     |
|             +-------------------------------+-------------------------------+                     |
|             |                               |                               |                     |
|             v                               v                               v                     |
|   +--------------------+          +--------------------+          +--------------------+          |
|   |  Window 1 (Bucket) |          |  Window 2 (Bucket) |          |  Window 3 & 4      |          |
|   |  09:15 to 10:45    |          |  10:45 to 12:15    |          |  12:15 to 15:30    |          |
|   |  Offset: Delta_1   |          |  Offset: Delta_2   |          |  Offsets: D_3, D_4 |          |
|   +--------------------+          +--------------------+          +--------------------+          |
|             \                               |                               /                     |
|              +------------------------------+------------------------------+                      |
|                                             |                                                     |
|                                             v                                                     |
|   +-----------------------------------------------------------------------------------------+     |
|   |                     SEALED COMPLIANCE ENVELOPE (AES-256-GCM)                            |     |
|   |   - Unlocks in memory exactly 10 milliseconds before each snapshot bucket fires         |     |
|   |   - Published to WORM S3 Compliance Bucket & Notarized to Hyperledger Besu at 15:35 IST |     |
|   +-----------------------------------------------------------------------------------------+     |
+---------------------------------------------------------------------------------------------------+
```

#### Snapshot Window Distribution
The 375 minutes of primary exchange trading (09:15 to 15:30 IST) are partitioned into four disjoint intervals:
- **Bucket 1:** 09:15:00 to 10:45:00 IST ($T_{\text{start}, 1} = 0$, $\text{Duration} = 90\text{ min} = 5,400\text{ sec}$)
- **Bucket 2:** 10:45:00 to 12:15:00 IST ($T_{\text{start}, 2} = 5,400$, $\text{Duration} = 90\text{ min} = 5,400\text{ sec}$)
- **Bucket 3:** 12:15:00 to 13:45:00 IST ($T_{\text{start}, 3} = 10,800$, $\text{Duration} = 90\text{ min} = 5,400\text{ sec}$)
- **Bucket 4:** 13:45:00 to 15:30:00 IST ($T_{\text{start}, 4} = 16,200$, $\text{Duration} = 105\text{ min} = 6,300\text{ sec}$)

For each bucket $i \in \{1, 2, 3, 4\}$, the exact snapshot second offset $\Delta_i$ is derived as:

$$\Delta_i = \text{uint32}(\text{HMAC\_SHA256}(\text{Seed}, \text{"SNAPSHOT\_WINDOW\_"} \parallel i)[0..4]) \pmod{\text{Duration}_i}$$

$$T_{\text{snapshot}, i} = T_{\text{start}, i} + \Delta_i$$

This mathematical approach guarantees:
1. **Unpredictability:** Trading participants cannot predict snapshot intervals in advance to temporarily deposit capital or manipulate open interest.
2. **Auditability:** At 15:35:00 IST, the seed and snapshot times are committed to the immutable audit log and published to regulatory observer nodes for zero-knowledge verification.

### 2.4 Mathematical Margin Formulation & Penalty Structure

At each snapshot timestamp $T_k$, the Risk Engine computes the total margin requirement $M_{\text{req}, u}$ and effective available collateral $C_{\text{eff}, u}$ for every user account $u$:

$$M_{\text{req}, u}(T_k) = \text{SPAN}_u(T_k) + \text{ELM}_u(T_k) + \text{AddOn}_u(T_k)$$

Where:
- $\text{SPAN}_u(T_k)$: Portfolio risk derived from the worst-case loss across 16 standard exchange risk scenarios.
- $\text{ELM}_u(T_k)$: Extreme Loss Margin, calculated as $\max(3.5\%, 1.5 \times \sigma_{\text{asset}}) \times \text{Gross Notional Exposure}$.
- $\text{AddOn}_u(T_k)$: Ad-hoc surveillance margins, including Additional Surveillance Measure (ASM) and Graded Surveillance Measure (GSM) surcharges.

Effective collateral $C_{\text{eff}, u}$ incorporates statutory haircuts according to SEBI Category specifications:

$$C_{\text{eff}, u}(T_k) = \text{Cash}_u + \text{CBDC}_u + \sum_{j \in \text{Approved Sec}} Q_{u, j} \cdot P_j(T_k) \cdot (1 - H_j)$$

Haircut rules follow SEBI categorisation:
- **Cash & Cash Equivalents (G-Secs, T-Bills, Liquid ETFs):** Haircut $H_j = 10.0\%$.
- **Category 1 Equity (Index constituents with derivatives):** Haircut $H_j = \max(20.0\%, \text{VaR Rate}_j)$.
- **Category 2 Equity (BSE 500 / Nifty 500 non-derivatives):** Haircut $H_j = \max(30.0\%, \text{VaR Rate}_j + 10\%)$.
- **All other securities & unapproved tokens:** Haircut $H_j = 100.0\%$ (Zero margin credit allowed).

#### Peak Margin Identification
The Peak Margin for client $u$ on trade date $t$ is:

$$M_{\text{peak}, u} = \max_{k \in \{1, 2, 3, 4\}} M_{\text{req}, u}(T_k)$$

#### SEBI Statutory Penalty Schedule
If a deficit exists at any snapshot $k$, statutory penalties are assessed per client according to SEBI circular guidelines:

| Shortfall Percentage ($D_u / M_{\text{req}, u}$) | Penalty Rate (% of Shortfall Amount) | Operational Action |
|---|---|---|
| **$< 0.5\%$ (or shortfall $< \text{Rs } 5,000$)** | 0.5% | Warning notification dispatched to client |
| **$\ge 0.5\%$ and $< 5.0\%$** | 1.0% | Client trading account placed in `REDUCE_ONLY` |
| **$\ge 5.0\%$** | 2.0% | Complete suspension of new order entry |
| **Repeated Default ($\ge 3$ consecutive trading days)** | 5.0% + Account Suspension | Regulatory incident report dispatched to clearing corporation |

### 2.5 Pre-Emptive Liquidation & Auto-Mitigation State Machine

To protect clients from statutory penalties and insulate the platform clearing fund from member defaults, `risk-service` executes a continuous real-time monitor calculating Margin Utilization $U_u(t)$:

$$U_u(t) = \frac{M_{\text{req}, u}(t)}{C_{\text{eff}, u}(t)}$$

```
+---------------------------------------------------------------------------------------------------+
|                         DYNAMIC MARGIN UTILIZATION STATE MACHINE                                  |
|                                                                                                   |
|    +--------------------+                                                                         |
|    |      NORMAL        |   U < 80%                                                               |
|    |  All Orders Active |                                                                         |
|    +--------------------+                                                                         |
|              |                                                                                    |
|              |  U >= 80%                                                                          |
|              v                                                                                    |
|    +--------------------+                                                                         |
|    |     WARNING        |   80% <= U < 90%                                                        |
|    | Alert Client / Push|                                                                         |
|    +--------------------+                                                                         |
|              |                                                                                    |
|              |  U >= 90%                                                                          |
|              v                                                                                    |
|    +--------------------+                                                                         |
|    |   REDUCE_ONLY      |   90% <= U < 100%                                                       |
|    | Block Size Increases                                                                         |
|    +--------------------+                                                                         |
|              |                                                                                    |
|              |  U >= 100% (or T_snap - t < 120s and U >= 95%)                                     |
|              v                                                                                    |
|    +--------------------+                                                                         |
|    |  AUTO_MITIGATION   |   Mass Purge Resting Bids -> Liquidate Margin-Heavy Legs                |
|    | Position Restored  |   Target: Restore U < 80% before Snapshot Execution                     |
|    +--------------------+                                                                         |
+---------------------------------------------------------------------------------------------------+
```

#### State Transition Specifications
1. **NORMAL ($U_u < 80\%$):** Standard order entry, modification, and execution across all permitted instruments.
2. **WARNING ($80\% \le U_u < 90\%$):** Automated high-priority alerts emitted via WebSocket and push notification. Margin call advisory requests additional collateral.
3. **REDUCE_ONLY ($90\% \le U_u < 100\%$):** Order gateway rejects any new order that increases gross exposure or SPAN requirements. Only position-closing or risk-reducing orders are accepted.
4. **AUTO_MITIGATION ($U_u \ge 100\%$, or $U_u \ge 95\%$ when time to snapshot bucket expiration is less than 120 seconds):**
   - **Step 1:** The matching engine immediately cancels all unexecuted resting buy orders for client $u$.
   - **Step 2:** Risk engine identifies the position leg with the highest marginal SPAN risk contribution ($\Delta M / \Delta Q$).
   - **Step 3:** Liquidation engine places automated market-to-limit IOC (Immediate-or-Cancel) orders to square off open intraday positions until $U_u$ drops below $80\%$.
   - **Step 4:** Detailed execution records are logged into the audit vault with rationale `SEBI_PEAK_MARGIN_DEFICIT_PREVENTION`.

### 2.6 Snapshot Archival, WORM Compliance & SERG Gateway Dispatch

Every snapshot calculation produces a cryptographically sealed compliance payload.

```
+---------------------------------------------------------------------------------------------------+
|                        REGULATORY REPORTING DISPATCH ARCHITECTURE                                 |
|                                                                                                   |
|   +-----------------------+     Kafka Topic:                       +--------------------------+   |
|   |      risk-service     | -------------------------------------> | continuous-regulatory-   |   |
|   | (Snapshot Calculator) |   compliance.peak_margin_snapshots.v1  |        reporting         |   |
|   +-----------------------+                                        +--------------------------+   |
|                                                                                 |                 |
|                   +-------------------------------------------------------------+                 |
|                   |                                                             |                 |
|                   v                                                             v                 |
|   +-------------------------------+                             +-----------------------------+   |
|   |   FIPS 140-2 Level 3 HSM      |                             |   AWS S3 Object Lock        |   |
|   |   XMLDSig Class-3 Corporate   |                             |   (WORM Compliance Mode)    |   |
|   |   Certificate Sealer          |                             |   8-Year Statutory Hold     |   |
|   +-------------------------------+                             +-----------------------------+   |
|                   |                                                                               |
|                   v                                                                               |
|   +-------------------------------------------------------------------------------------------+   |
|   |         SEBI ELECTRONIC REPORTING GATEWAY (SERG) / CLEARING CORPORATION SFTP              |   |
|   |   - Protocol: SFTP over Mutual TLS 1.3 / AS2 Encrypted Channel                            |   |
|   |   - Format: SEBI Standard XBRL / XML Peak Margin Report Schema (Version 2.4)             |   |
|   |   - Acknowledgment: Cryptographic MDN Receipt Ingested and Stored in PostgreSQL            |   |
|   +-------------------------------------------------------------------------------------------+   |
+---------------------------------------------------------------------------------------------------+
```

---

## 3. Fractional Share Pass-Through Proxy Voting Portal (REG-02)

### 3.1 Legal & Statutory Mandate

Under Section 108 of the Indian Companies Act 2013 read with Rule 20 of the Companies (Management and Administration) Rules 2014, and Regulation 44 of the SEBI (Listing Obligations and Disclosure Requirements) Regulations 2015, every listed company in India is legally mandated to provide remote electronic voting (e-voting) facilities to its shareholders for all shareholder resolutions.

In traditional markets, beneficial ownership is registered at the whole-share level with depositories (NSDL and CDSL). However, the Growww platform permits retail and institutional participants to purchase and hold fractional micro-shares down to six decimal places ($10^{-6}$ units).

```
   1 Whole Equity Share = 1,000,000 Micro-Share Units (Tokenized Base Units)
```

Because NSDL and CDSL infrastructure only records integer shares held in the name of the **Custodian Trustee Segregated Demat Pool Account** (`IN300123-10000001`), an architectural bridge is required to ensure that beneficial fractional holders are not disenfranchised. The Corporate Governance Pass-Through Voting Portal guarantees that voting rights are passed directly through to fractional investors in exact mathematical proportion to their holdings.

### 3.2 End-to-End Governance Architecture & Lifecycle

The lifecycle spans resolution discovery, record-date snapshotting, digital ballot presentation, cryptographic vote casting, proportional mathematical apportionment, depository proxy transmission, and on-chain verification:

```
+---------------------------------------------------------------------------------------------------+
|                   FRACTIONAL PROXY VOTING LIFECYCLE & APPORTIONMENT PIPELINE                      |
|                                                                                                   |
|  [STEP 1: Resolution Ingestion]                                                                   |
|   NSDL/CDSL / Exchange Corporate Action Feed Ingestion via corporate-actions microservice         |
|   Extracts: ISIN, AGM/EGM Date, Resolutions (1..N), Cut-Off Record Date (T_rec), Voting Deadline   |
|                                     |                                                             |
|                                     v                                                             |
|  [STEP 2: Record-Date Fractional Snapshot]                                                        |
|   At T_rec (23:59:59 IST), portfolio-service freezes fractional balances H(u, s) across all users|
|   Constructs Sparse Merkle Tree of holdings; posts Root Hash R_holdings to Hyperledger Besu      |
|                                     |                                                             |
|                                     v                                                             |
|  [STEP 3: Pass-Through Voting Portal (apps/growww_web & apps/growww_flutter)]                     |
|   Renders resolution ballot with management rationale and proxy advisory recommendations         |
|   Investor signs ballot using EIP-712 structured typed data (FOR / AGAINST / ABSTAIN)             |
|                                     |                                                             |
|                                     v                                                             |
|  [STEP 4: Pro-Rata Vote Aggregation & Hamilton-Hare Apportionment]                                |
|   custody-adapter aggregates fractional votes across entire user pool:                            |
|   Calculates raw totals: V_for, V_against, V_abstain, V_unvoted                                    |
|   Applies Largest Remainder (Hamilton-Hare) Method to map fractional sums to integer Demat shares|
|                                     |                                                             |
|                                     v                                                             |
|  [STEP 5: Depository Proxy Execution]                                                             |
|   Custodian Trustee signs consolidated ISO 20022 seev.004 e-Voting instruction file via HSM       |
|   Transmits via secure API / SFTP into NSDL e-Voting / CDSL eVoting before regulatory cutoff      |
|                                     |                                                             |
|                                     v                                                             |
|  [STEP 6: Immutable On-Chain Notarization & Client Verification]                                  |
|   Merkle root of all individual investor ballots + Depository ACK notarized on Besu ledger        |
|   Investors verify inclusion of their vote via client-side Merkle proof in Growww UI              |
+---------------------------------------------------------------------------------------------------+
```

### 3.3 Record Date Cryptographic Snapshot Protocol

At the official record date cutoff timestamp $T_{\text{rec}}$ (typically 23:59:59 IST on the date specified in the issuer notice):
1. `corporate-actions` service issues a synchronization barrier across `portfolio-service` and `settlement-service`.
2. For the target security ISIN, the system extracts the exact fractional balance $H_{u, \text{isin}}$ for every active participant $u$:

$$\sum_{u=1}^{N} H_{u, \text{isin}} = S_{\text{custody\_total}}$$

3. An immutable Sparse Merkle Tree (SMT) of investor entitlements is generated:

$$\text{Leaf}_u = \text{SHA256}(u \parallel \text{ISIN} \parallel H_{u, \text{isin}} \parallel \text{Nonce})$$

4. The entitlement root $R_{\text{entitlement}}$ is committed to PostgreSQL and published as an on-chain event on Hyperledger Besu contract `CorporateGovernanceRegistry.sol`.

### 3.4 Web & Mobile Voting Portal Specifications (`apps/growww_web`)

The voting portal provides an intuitive, transparent interface for retail and institutional shareholders:
- **Resolution Analysis:** Renders full statutory resolution text, explanatory statements under Section 102 of the Companies Act, and board of directors recommendations.
- **Independent Proxy Advisory Integration:** Displays automated consensus guidance and analysis from SEBI-registered proxy advisory firms:
  - Institutional Investor Advisory Services (IiAS)
  - Stakeholders Empowerment Services (SES)
  - InGovern Research Services
- **ESG Governance Scores:** Highlights potential corporate governance risks (excessive director remuneration, related-party transactions, auditor tenure).
- **Cryptographic Ballot Signing:** Voting choices (`FOR`, `AGAINST`, `ABSTAIN`) are cryptographically signed using EIP-712 structured typed data, binding the investor identity, resolution ID, and timestamp:

```json
{
  "types": {
    "EIP712Domain": [
      {"name": "name", "type": "string"},
      {"name": "version", "type": "string"},
      {"name": "chainId", "type": "uint256"},
      {"name": "verifyingContract", "type": "address"}
    ],
    "ProxyVote": [
      {"name": "voterAccount", "type": "address"},
      {"name": "isin", "type": "string"},
      {"name": "meetingId", "type": "string"},
      {"name": "resolutionId", "type": "uint32"},
      {"name": "choice", "type": "uint8"},
      {"name": "fractionalWeight", "type": "uint256"},
      {"name": "nonce", "type": "uint256"},
      {"name": "deadline", "type": "uint256"}
    ]
  },
  "primaryType": "ProxyVote",
  "domain": {
    "name": "Growww Corporate Governance Portal",
    "version": "1.0",
    "chainId": 13371,
    "verifyingContract": "0x1122334455667788990011223344556677889900"
  },
  "message": {
    "voterAccount": "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
    "isin": "INE002A01018",
    "meetingId": "AGM-2026-RELIANCE",
    "resolutionId": 3,
    "choice": 1,
    "fractionalWeight": 1450200,
    "nonce": 42,
    "deadline": 1789804800
  }
}
```

### 3.5 Deterministic Integer Share Apportionment Engine (Largest Remainder / Hamilton-Hare)

Because depositories (NSDL and CDSL) and issuer RTAs only accept integer share votes, the system must translate aggregate fractional voting preferences into discrete whole share votes without statistical bias.

#### Step 1: Raw Fractional Summation
For each resolution $r$, `custody-adapter` sums the exact fractional holdings for all participating voters:

$$V_{\text{for}} = \sum_{u \in \text{Voters}_{\text{FOR}}} H_u$$

$$V_{\text{against}} = \sum_{u \in \text{Voters}_{\text{AGAINST}}} H_u$$

$$V_{\text{abstain}} = \sum_{u \in \text{Voters}_{\text{ABSTAIN}}} H_u$$

Let $V_{\text{voted}} = V_{\text{for}} + V_{\text{against}} + V_{\text{abstain}}$.

#### Step 2: Custodian Neutrality Invariant on Unvoted Shares
If some fractional holders do not cast a ballot prior to the voting deadline, their unvoted weight is:

$$V_{\text{unvoted}} = S_{\text{custody\_total}} - V_{\text{voted}}$$

**Statutory Invariant:** To prevent conflicts of interest, the Custodian Trustee is **strictly prohibited from exercising discretionary voting power** on unvoted shares. Unvoted shares are categorized as `UNVOTED_ABSTAIN` or excluded from the voting return in accordance with the issuer's articles of association and SEBI custodial regulations.

#### Step 3: Hamilton-Hare (Largest Remainder) Apportionment Algorithm
To map the fractional totals to the total whole shares $S_{\text{whole}} = \lfloor S_{\text{custody\_total}} \rfloor$ held in the demat pool:

1. Compute the exact quota of shares for each voting category $c \in \{\text{FOR}, \text{AGAINST}, \text{ABSTAIN}\}$:

$$Q_c = S_{\text{whole}} \times \frac{V_c}{S_{\text{custody\_total}}}$$

2. Allocate the integer base shares:

$$I_c = \lfloor Q_c \rfloor$$

3. Compute the fractional remainder for each category:

$$R_c = Q_c - I_c$$

4. Calculate remaining unassigned whole shares:

$$K = S_{\text{whole}} - \sum_{c} I_c$$

5. Rank categories by descending remainder $R_c$. Allocate $+1$ additional share to the top $K$ categories until all $S_{\text{whole}}$ shares are distributed.

#### Numerical Verification Example
Suppose the Custodian Pool Demat Account holds exactly $10,000$ whole shares of Reliance Industries Ltd (`INE002A01018`). The platform clients hold $10,000.000000$ fractional units. Voting closes with the following aggregate fractional tallies:
- $V_{\text{for}} = 5,432.650000$ shares (54.3265%)
- $V_{\text{against}} = 3,124.800000$ shares (31.2480%)
- $V_{\text{abstain}} = 1,442.550000$ shares (14.4255%)

**Execution:**
1. Quotas:
   - $Q_{\text{for}} = 10,000 \times 0.543265 = 5,432.65 \implies I_{\text{for}} = 5,432$, Remainder $R_{\text{for}} = 0.65$
   - $Q_{\text{against}} = 10,000 \times 0.312480 = 3,124.80 \implies I_{\text{against}} = 3,124$, Remainder $R_{\text{against}} = 0.80$
   - $Q_{\text{abstain}} = 10,000 \times 0.144255 = 1,442.55 \implies I_{\text{abstain}} = 1,442$, Remainder $R_{\text{abstain}} = 0.55$
2. Sum of integer parts: $5,432 + 3,124 + 1,442 = 9,998$ shares.
3. Remaining shares to distribute: $K = 10,000 - 9,998 = 2$ shares.
4. Ordering by remainder:
   - 1st: `AGAINST` with $R = 0.80 \implies +1 \text{ share}$
   - 2nd: `FOR` with $R = 0.65 \implies +1 \text{ share}$
   - 3rd: `ABSTAIN` with $R = 0.55 \implies +0 \text{ shares}$
5. Final consolidated depository vote:
   - **`FOR`:** $5,433$ shares
   - **`AGAINST`:** $3,125$ shares
   - **`ABSTAIN`:** $1,442$ shares
   - **Total:** $10,000$ shares ($100.00\%$ accounted for, zero bias).

### 3.6 Depository Proxy Execution & Cryptographic Verification

1. **Submission Payload Generation:** `custody-adapter` serializes the final integer breakdown into the ISO 20022 `seev.004.001.06` Meeting Instruction XML format or the NSDL e-Voting Batch API format.
2. **HSM Digital Signature:** The instruction file is digitally signed with the Custodian Trustee's Class-3 Organization DSC stored in a FIPS 140-2 Level 3 HSM.
3. **Depository Receipt Ingestion:** Upon dispatch to the NSDL/CDSL e-Voting gateway, the returned transaction reference (`DepositoryAckID`) is paired with the internal corporate action ID.
4. **On-Chain Audit Attestation:** The system constructs a Merkle tree of all individual signed ballots and posts the root hash along with the `DepositoryAckID` to the `CorporateGovernanceRegistry.sol` smart contract on Hyperledger Besu.
5. **Investor Proof Verification:** In `apps/growww_web`, the user can click "Verify My Vote". The client downloads the lightweight Merkle audit path and cryptographically verifies that their specific vote and fractional weight were included in the root hash submitted to the depository.

---

## 4. Tiered Cross-Chain Finality & Reorganization Protection Matrix (REG-03)

### 4.1 Cross-Border Threat Modeling & Regulatory Objective

Through the International Gateway Entity operating under the IFSCA regulatory sandbox in GIFT City, foreign investors are permitted to fund trading accounts using major global blockchains: **Bitcoin (BTC)**, **Ethereum (ETH / ERC-20 USDC)**, and **Solana (SOL / SPL USDC)**.

These external public networks operate on fundamentally disparate consensus mechanisms with probabilistic or multi-phase finality. A blockchain reorganization occurs when a network node abandons what was previously considered the canonical chain tip in favor of a longer or heavier competing fork.

```
+---------------------------------------------------------------------------------------------------+
|                        CROSS-CHAIN REORGANIZATION (DOUBLE-SPEND) RISK                             |
|                                                                                                   |
|                      Deposit Tx Observed in Block 100                                             |
|                                     |                                                             |
|   Canonical Tip:  ... -> [Block 99] -> [Block 100] -> [Block 101]                                 |
|                                \                                                                  |
|   Competing Fork:               -> [Block 100'] -> [Block 101'] -> [Block 102']                  |
|                                                                                                   |
|   IF deposit credit was granted at Block 100, and Competing Fork becomes canonical:               |
|   - The deposit transaction disappears from the canonical chain.                                  |
|   - Internal ledger purchasing power becomes unbacked (Breaking INV-REG-2).                       |
|   - Attacker withdraws synthetic securities, creating a balance-sheet deficit.                   |
+---------------------------------------------------------------------------------------------------+
```

To eliminate balance sheet insolvency and comply with IFSCA Capital Market Regulations, the platform enforces a **Tiered Cross-Chain Finality Matrix** that aligns internal purchasing power release with irreversible economic finality.

### 4.2 Consensus Foundations & Theoretical Finality Formulations

#### 4.2.1 Bitcoin: Proof-of-Work (Nakamoto Consensus)
Bitcoin exhibits probabilistic finality governed by the Poisson distribution. Assuming an attacker controls an adversarial fraction of network hash power $q < 0.5$ (with honest hash power $p = 1 - q$), the probability $P(z)$ of an attacker successfully reorganizing a transaction after $z$ confirmations is:

$$P(z) = \sum_{k=0}^{\infty} \frac{\lambda^k e^{-\lambda}}{k!} \min\left(1, \left(\frac{q}{p}\right)^{\max(z - k, 0)}\right), \quad \lambda = z \frac{q}{p}$$

For $q = 0.20$ (a major mining pool attack):
- At $z = 1$ confirmation: $P(1) \approx 0.368$ (High risk)
- At $z = 3$ confirmations: $P(3) \approx 0.051$ (Moderate risk)
- At $z = 6$ confirmations: $P(6) \approx 0.0026$ ($0.26\%$, Standard threshold)
- At $z = 12$ confirmations: $P(12) \approx 0.00001$ ($0.001\%$, Institutional threshold)

#### 4.2.2 Ethereum: Proof-of-Stake (Gasper Consensus)
Ethereum pairs Casper FFG (Friendly Finality Gadget) with LMD-GHOST. Transactions are bundled into slots (12 seconds) and epochs (32 slots = 6.4 minutes).
- **Justification:** An epoch boundary checkpoint is justified when $\ge 2/3$ of all active validator stake votes for it.
- **Finalization:** A justified checkpoint becomes finalized when the immediately succeeding epoch is also justified.
- **Cryptoeconomic Security:** Finality requires two consecutive justified epochs ($\approx 64$ slots $\approx 12.8$ minutes). Reversing a finalized Ethereum block requires an attacker to sacrifice at least $1/3$ of the entire network's staked ETH ($> 10\text{ million ETH} \approx \$30+\text{ billion USD}$) via automatic slashing protocols.

#### 4.2.3 Solana: Proof-of-History (PoH) + Tower BFT
Solana utilizes Proof-of-History as a cryptographic clock combined with Tower BFT. Slots occur every $\sim 400\text{ ms}$.
- **Optimistic Confirmation (Processed):** 1 slot ($\approx 400\text{ ms}$). Observed by local leader, subject to micro-forking.
- **Confirmed:** When $\ge 2/3$ of the active cluster stake votes on the block ($\approx 4\text{ to }8\text{ seconds}$).
- **Finalized (Rooted):** Tower BFT employs exponential lockout doubling ($2^k$ slots). A slot is declared **finalized (rooted)** once it accumulates 31 consecutive descendant confirmations, meaning $> 2/3$ of validators have reached maximal lockout ($\approx 32\text{ slots} \approx 12.8\text{ to }25.0\text{ seconds}$). A finalized slot cannot be rolled back without slashing more than $2/3$ of the network validator stake.

#### 4.2.4 Hyperledger Besu: Enterprise Consortium QBFT
The internal platform settlement ledger runs Quorum Byzantine Fault Tolerance (QBFT).
- **Consensus Rules:** 3-phase consensus (Pre-Prepare, Prepare, Commit) with block time of 2.0 seconds.
- **Fault Tolerance:** Tolerates $f$ Byzantine nodes out of $N = 3f + 1$ validators ($N = 7, f = 2$).
- **Instant Deterministic Finality:** Once a block receives $2f + 1 = 5$ commit signatures, it is final. Reorganizations are mathematically impossible ($0\text{ forks / }0\text{ reorgs}$).

### 4.3 Multi-Tiered Crediting & Collateral Matrix

Purchasing power is released in three distinct tiers based on deposit value and destination purpose:

```
+---------------------------------------------------------------------------------------------------+
|                            MULTI-TIERED CROSS-CHAIN CREDITING PIPELINE                            |
|                                                                                                   |
|    Deposit Detected                                                                               |
|           |                                                                                       |
|           +---> Tier 1: Micro / Fast Trading (< $1,000 USD)                                       |
|           |     - Fast Execution in liquid pairs                                                  |
|           |     - Internal SGF Insurance Buffer guarantees potential reorg loss                   |
|           |                                                                                       |
|           +---> Tier 2: Standard Trading ($1,000 to $100,000 USD)                                 |
|           |     - Unrestricted Spot & Derivative Trading                                          |
|           |     - Granted strictly upon Economic Finality                                         |
|           |                                                                                       |
|           +---> Tier 3: Institutional & Outbound Settlement (> $100,000 USD)                       |
|                 - Cross-Chain Bridge Transfers & Fiat Bank Withdrawals                            |
|                 - Requires Maximal Deep Finality + Manual Compliance Attestation                  |
+---------------------------------------------------------------------------------------------------+
```

| Network / Asset | Consensus Engine | Tier 1: Micro Trading ($< \$1{,}000$) | Tier 2: Standard Trading ($\$1\text{k} - \$100\text{k}$) | Tier 3: Institutional ($> \$100{,}000$) | Latency to Tier 2 (Standard) | Reorg Resistance SLA |
|---|---|---|---|---|---|---|
| **Bitcoin (BTC)** | Proof-of-Work | 2 Confirmations | 6 Confirmations | 12 Confirmations | ~60 Minutes | $99.999\%$ Probabilistic |
| **Ethereum (ETH / ERC-20)** | Casper FFG + LMD-GHOST | 12 Blocks (1 Epoch) | 64 Blocks (2 Epochs, Finalized) | 96 Blocks (3 Epochs) | ~12.8 Minutes | Cryptoeconomic ($> \$30\text{B}$ Slashing) |
| **Solana (SOL / SPL)** | PoH + Tower BFT | 16 Confirmed Slots | 64 Finalized Slots (Rooted) | 128 Finalized Slots | ~25.6 Seconds | Cryptoeconomic ($> 2/3$ Stake Slashing) |
| **Hyperledger Besu (Internal)** | QBFT Consortium | 1 Block | 1 Block | 1 Block | 2.0 Seconds | Mathematical Absolute (0 Reorgs) |

### 4.4 Ingestion Gateway & Real-Time Reorg Watchdog Architecture

The ingestion infrastructure operates redundant RPC nodes across distinct geographical availability zones:
- **Bitcoin:** 3 x `bitcoind` full nodes with ZeroMQ block and transaction streamers.
- **Ethereum:** 3 x `Geth` execution clients paired with `Prysm` consensus clients (tracking checkpoint finality via beacon API).
- **Solana:** 3 x `solana-validator` nodes subscribing to WebSocket `slotSubscribe` and `programSubscribe` channels with commitment `finalized`.

```
+---------------------------------------------------------------------------------------------------+
|                        CROSS-CHAIN WATCHDOG & INGESTION ARCHITECTURE                              |
|                                                                                                   |
|   +---------------------+     +---------------------+     +---------------------+                 |
|   |  bitcoind Cluster   |     |  Ethereum Geth/Prysm|     |  Solana RPC Cluster |                 |
|   +---------------------+     +---------------------+     +---------------------+                 |
|              \                           |                           /                            |
|               +--------------------------+--------------------------+                             |
|                                          |                                                        |
|                                          v                                                        |
|   +-----------------------------------------------------------------------------------------+     |
|   |                       DISTRIBUTED INGESTION WATCHDOG ENGINE                             |     |
|   |   - Maintains sliding-window DAG of recent block headers (Depth = 256 blocks)            |     |
|   |   - Compares parent hashes on every incoming block/slot                                 |     |
|   |   - Detects canonical chain reorganizations in < 250 milliseconds                        |     |
|   +-----------------------------------------------------------------------------------------+     |
|                     |                                                   |                         |
|   (Normal Finality Confirmation)                                (Reorganization Detected)         |
|                     v                                                   v                         |
|   +------------------------------------+              +-------------------------------------+     |
|   | Topic: crosschain.deposit.finalized|              | Topic: crosschain.reorg.detected    |     |
|   | Credited to User Trading Account   |              | Instantiates 5-Phase Clawback Saga  |     |
|   +------------------------------------+              +-------------------------------------+     |
+---------------------------------------------------------------------------------------------------+
```

### 4.5 5-Phase Cross-Chain Clawback Saga Protocol (Runbook 07 Integration)

In the unprecedented event that an external public blockchain suffers a catastrophic deep reorganization exceeding the standard Tier 2 finality threshold, an invalidated deposit would leave internal ledger balances unbacked.

The platform immediately triggers the automated **5-Phase Clawback Saga**, orchestrating rapid isolation and balance-sheet restoration across microservices:

```
+---------------------------------------------------------------------------------------------------+
|                          5-PHASE CROSS-CHAIN CLAWBACK SAGA WORKFLOW                               |
|                                                                                                   |
|   [Deep Reorganization Detected by Gateway Watchdog]                                              |
|                           |                                                                       |
|                           v                                                                       |
|   +-----------------------------------------------------------------------------------------+     |
|   | PHASE 1: ACCOUNT QUARANTINE & ORDER PURGE (Latency: < 50ms)                             |     |
|   | - Account state mutated to MARGIN_RECOVERY_LOCKED in Redis & PostgreSQL                     |     |
|   | - Matching engine broadcasts mass-cancel order for all active resting limit orders     |     |
|   | - Withdrawal gateways and outbound transfers immediately frozen for target account      |     |
|   +-----------------------------------------------------------------------------------------+     |
|                           |                                                                       |
|                           v                                                                       |
|   +-----------------------------------------------------------------------------------------+     |
|   | PHASE 2: DYNAMIC COLLATERAL VALUATION REASSESSMENT (Latency: < 200ms)                   |     |
|   | - Deduct invalidated deposit from total account collateral ledger                       |     |
|   | - Recalculate portfolio SPAN margin and maintenance requirements                        |     |
|   | - Determine Net Margin Shortfall Deficit: D_clawback = M_maint - C_remaining            |     |
|   +-----------------------------------------------------------------------------------------+     |
|                           |                                                                       |
|                           v                                                                       |
|   +-----------------------------------------------------------------------------------------+     |
|   | PHASE 3: PRIORITY AUTO-LIQUIDATION EXECUTION (Latency: < 2,000ms)                       |     |
|   | - If D_clawback > 0: Liquidation engine submits aggressive IOC limit orders             |     |
|   | - Closes open derivative, equity, and tokenized positions on internal LOB               |     |
|   | - Recovered cash proceeds credited to cover deficit                                     |     |
|   +-----------------------------------------------------------------------------------------+     |
|                           |                                                                       |
|                           v                                                                       |
|   +-----------------------------------------------------------------------------------------+     |
|   | PHASE 4: SETTLEMENT GUARANTEE FUND (SGF) LOSS ABSORPTION (Latency: < 5,000ms)           |     |
|   | - If liquidation proceeds fall short (Deficit remains after full liquidation):          |     |
|   | - Trigger automated claim on GIFT City SGF Reorganization Buffer                       |     |
|   | - Balance sheet neutralized; platform maintains zero insolvency across all members     |     |
|   +-----------------------------------------------------------------------------------------+     |
|                           |                                                                       |
|                           v                                                                       |
|   +-----------------------------------------------------------------------------------------+     |
|   | PHASE 5: HYPERLEDGER BESU SYNTHETIC TOKEN BURN (Latency: < 4,000ms)                     |     |
|   | - Multi-sig custodian bridge contract invokes burnSyntheticRepresentation() on Besu    |     |
|   | - Destroys synthetic wrapped units corresponding to orphaned deposit                    |     |
|   | - Restores INV-REG-2 (1:1 physical and reserve backing invariant)                       |     |
|   +-----------------------------------------------------------------------------------------+     |
|                           |                                                                       |
|                           v                                                                       |
|   [POST-INCIDENT FORENSIC AUDIT & STATUTORY FILING]                                               |
|   - WORM compliance log exported; statutory notification filed with IFSCA within 24 hours       |
+---------------------------------------------------------------------------------------------------+
```

#### Detailed Phase Execution Specifications
1. **Phase 1: MARGIN_RECOVERY_LOCKED**
   - Microservice: `risk-service` and `order-service`.
   - Action: Updates user account status flag to `MARGIN_RECOVERY_LOCKED`. Emits `trading.order.purge_user` to Kafka topic `trading.orders.control`.
   - Matching Engine (`services/matching-engine`): Intercepts the purge directive and evicts all resting limit orders for that user within $< 50\mu\text{s}$, preventing predatory trade fills.
2. **Phase 2: Collateral Valuation Reassessment**
   - Microservice: `risk-service` and `portfolio-service`.
   - Action: Removes the unconfirmed deposit record from double-entry journal table `journal_entry`. Evaluates whether the remaining cash and approved securities satisfy the Maintenance Margin Requirement ($M_{\text{maint}}$).
3. **Phase 3: Priority Auto-Liquidation**
   - Microservice: `services/risk-service` Liquidation Module.
   - Action: If a deficit $D_{\text{clawback}} > 0$ exists, the liquidation engine routes aggressive IOC orders to the matching engine. Positions are liquidated in order of liquidity (highest market depth first) to minimize slippage.
4. **Phase 4: SGF Buffer Loss Absorption**
   - Microservice: `services/settlement-service` and `settlement-guarantee-fund-service`.
   - Action: If total liquidation value fails to clear the deficit (e.g. during a market gap), the remaining shortfall is absorbed by the **GIFT City Settlement Guarantee Fund (SGF) Buffer** (governed under IFSCA clearing rules). This guarantees that neither other market participants nor platform solvency is compromised.
5. **Phase 5: Hyperledger Besu Synthetic Token Burn**
   - Microservice: `custody-adapter` and `services/settlement-service`.
   - Action: Initiates an on-chain transaction calling `burnUnbackedSynthetic(address user, address token, uint256 amount)` on `CrossEntityBridge.sol`. Validator nodes verify the proof of reorg and execute the burn, restoring exact 1:1 parity between on-chain token supply and off-chain vault reserves.

### 4.6 Statutory Notification & Forensic Reporting

Within 24 hours of any clawback saga execution:
1. **IFSCA Formal Notification:** `services/reporting-service` compiles an automated incident package containing raw blockchain reorg logs, wallet addresses, transaction hashes, matching engine trade fills, and SGF drawdown metrics.
2. **Statutory Filing:** Dispatched electronically to the IFSCA Supervision Division via the FINnet 2.0 regulatory gateway.
3. **Forensic Archival:** Stored with an 8-year statutory hold on AWS S3 Object Lock in Compliance Mode.

---

## 5. Cross-Cutting Regulatory Data Models, Schemas & API Contracts

### 5.1 Protobuf Contract: Regulatory Risk & Governance (`regulatory_governance.proto`)

```protobuf
syntax = "proto3";

package growww.regulatory_governance.v1;

option go_package = "github.com/growww/services/common/gen/v1/governance;governancev1";
option java_multiple_files = true;
option java_package = "com.growww.regulatory_governance.v1";

// Service managing statutory compliance, peak margin snapshots, and voting
service RegulatoryGovernanceService {
  // Peak Margin RPCs (REG-01)
  rpc CapturePeakMarginSnapshot (PeakMarginSnapshotRequest) returns (PeakMarginSnapshotResponse);
  rpc GetClientMarginStatus (ClientMarginStatusRequest) returns (ClientMarginStatusResponse);
  
  // Corporate Proxy Voting RPCs (REG-02)
  rpc IngestResolutionNotice (IngestResolutionRequest) returns (IngestResolutionResponse);
  rpc SubmitProxyVote (SubmitVoteRequest) returns (SubmitVoteResponse);
  rpc AggregateDepositoryVotes (AggregateVotesRequest) returns (AggregateVotesResponse);
  
  // Cross-Chain Finality RPCs (REG-03)
  rpc IngestCrossChainDeposit (CrossChainDepositRequest) returns (CrossChainDepositResponse);
  rpc TriggerClawbackSaga (ClawbackSagaRequest) returns (ClawbackSagaResponse);
}

// --- REG-01 Models ---

message PeakMarginSnapshotRequest {
  string snapshot_id = 1;
  uint32 window_index = 2; // 1, 2, 3, or 4
  int64 timestamp_ns = 3;
  string market_open_block_hash = 4;
}

message PeakMarginSnapshotResponse {
  string snapshot_id = 1;
  uint32 accounts_evaluated = 2;
  uint32 shortfalls_detected = 3;
  string total_shortfall_inr = 4;
  string merkle_root = 5;
  int64 completed_at_ns = 6;
}

message ClientMarginStatusRequest {
  string account_id = 1;
}

message ClientMarginStatusResponse {
  string account_id = 1;
  string span_margin_inr = 2;
  string elm_margin_inr = 3;
  string total_margin_required_inr = 4;
  string effective_collateral_inr = 5;
  double utilization_pct = 6;
  string risk_state = 7; // NORMAL, WARNING, REDUCE_ONLY, AUTO_MITIGATION
  bool is_prefunded_session = 8;
}

// --- REG-02 Models ---

enum VoteChoice {
  VOTE_CHOICE_UNSPECIFIED = 0;
  VOTE_CHOICE_FOR = 1;
  VOTE_CHOICE_AGAINST = 2;
  VOTE_CHOICE_ABSTAIN = 3;
}

message IngestResolutionRequest {
  string isin = 1;
  string company_name = 2;
  string meeting_id = 3;
  int64 record_date_cutoff_ns = 4;
  int64 voting_deadline_ns = 5;
  repeated ResolutionItem resolutions = 6;
}

message ResolutionItem {
  uint32 resolution_id = 1;
  string title = 2;
  string description = 3;
  string resolution_type = 4; // ORDINARY, SPECIAL
  string management_recommendation = 5;
  string proxy_advisory_consensus = 6;
}

message IngestResolutionResponse {
  string meeting_id = 1;
  bool is_registered = 2;
  string holdings_snapshot_root = 3;
}

message SubmitVoteRequest {
  string account_id = 1;
  string isin = 2;
  string meeting_id = 3;
  uint32 resolution_id = 4;
  VoteChoice choice = 5;
  string eip712_signature = 6;
  int64 signed_at_ns = 7;
}

message SubmitVoteResponse {
  string vote_receipt_id = 1;
  bool accepted = 2;
  string fractional_weight_recorded = 3;
  string ballot_hash = 4;
}

message AggregateVotesRequest {
  string meeting_id = 1;
  string isin = 2;
  uint32 resolution_id = 3;
}

message AggregateVotesResponse {
  string meeting_id = 1;
  uint32 resolution_id = 2;
  uint64 total_custody_shares = 3;
  uint64 integer_shares_for = 4;
  uint64 integer_shares_against = 5;
  uint64 integer_shares_abstain = 6;
  uint64 integer_shares_unvoted = 7;
  string depository_xml_payload = 8;
  string besu_notarization_tx = 9;
}

// --- REG-03 Models ---

enum BlockchainNetwork {
  NETWORK_UNSPECIFIED = 0;
  NETWORK_BITCOIN = 1;
  NETWORK_ETHEREUM = 2;
  NETWORK_SOLANA = 3;
  NETWORK_BESU_INTERNAL = 4;
}

message CrossChainDepositRequest {
  BlockchainNetwork network = 1;
  string tx_hash = 2;
  string source_address = 3;
  string destination_vault = 4;
  string asset_symbol = 5;
  string amount = 6;
  uint64 block_or_slot_number = 7;
  uint32 confirmations_observed = 8;
}

message CrossChainDepositResponse {
  string internal_credit_id = 1;
  string status = 2; // DETECTED, TIER1_FAST_CREDITED, TIER2_FINALIZED, LOCKED
  uint32 current_tier = 3;
  bool economic_finality_achieved = 4;
}

message ClawbackSagaRequest {
  BlockchainNetwork network = 1;
  string orphaned_tx_hash = 2;
  string target_account_id = 3;
  string invalidated_amount = 4;
  string asset_symbol = 5;
  string reorg_detected_block_hash = 6;
}

message ClawbackSagaResponse {
  string saga_execution_id = 1;
  string final_saga_state = 2; // COMPLETED, SGF_DRAWDOWN_REQUIRED, MANUAL_ESCALATION
  string liquidated_value_inr = 3;
  string sgf_loss_absorbed_inr = 4;
  string tokens_burned_on_besu = 5;
  int64 resolved_at_ns = 6;
}
```

### 5.2 PostgreSQL & TimescaleDB Database Schema

```sql
-- Schema: regulatory_compliance
CREATE SCHEMA IF NOT EXISTS regulatory_compliance;

-- Table: Peak Margin Snapshot Log (REG-01)
CREATE TABLE regulatory_compliance.peak_margin_snapshots (
    id                      BIGSERIAL PRIMARY KEY,
    snapshot_uuid           UUID NOT NULL UNIQUE,
    snapshot_date           DATE NOT NULL,
    window_index            SMALLINT NOT NULL CHECK (window_index BETWEEN 1 AND 4),
    trigger_timestamp       TIMESTAMPTZ NOT NULL,
    completed_timestamp     TIMESTAMPTZ NOT NULL,
    seed_hash               TEXT NOT NULL,
    total_accounts_checked  INTEGER NOT NULL,
    deficit_accounts_count  INTEGER NOT NULL,
    total_shortfall_paise   BIGINT NOT NULL DEFAULT 0,
    merkle_state_root       BYTEA NOT NULL,
    serg_submission_ack     TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_peak_margin_date_window 
    ON regulatory_compliance.peak_margin_snapshots (snapshot_date, window_index);

-- Table: Client Peak Margin Records (REG-01)
CREATE TABLE regulatory_compliance.client_margin_snapshots (
    id                      BIGSERIAL PRIMARY KEY,
    snapshot_id             BIGINT NOT NULL REFERENCES regulatory_compliance.peak_margin_snapshots(id),
    account_id              UUID NOT NULL,
    span_margin_paise       BIGINT NOT NULL,
    elm_margin_paise        BIGINT NOT NULL,
    total_margin_req_paise  BIGINT NOT NULL,
    effective_collateral_paise BIGINT NOT NULL,
    shortfall_paise         BIGINT NOT NULL DEFAULT 0,
    penalty_assessed_paise  BIGINT NOT NULL DEFAULT 0,
    is_prefunded_session    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_client_margin_acct 
    ON regulatory_compliance.client_margin_snapshots (account_id, snapshot_id);

-- Table: Proxy Voting Resolutions (REG-02)
CREATE TABLE regulatory_compliance.voting_resolutions (
    id                      BIGSERIAL PRIMARY KEY,
    isin                    VARCHAR(12) NOT NULL,
    meeting_id              TEXT NOT NULL,
    resolution_id           INTEGER NOT NULL,
    title                   TEXT NOT NULL,
    resolution_type         VARCHAR(32) NOT NULL CHECK (resolution_type IN ('ORDINARY', 'SPECIAL')),
    record_date_cutoff      TIMESTAMPTZ NOT NULL,
    voting_deadline         TIMESTAMPTZ NOT NULL,
    total_pool_shares       BIGINT NOT NULL,
    holdings_merkle_root    BYTEA NOT NULL,
    status                  VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_resolution UNIQUE (isin, meeting_id, resolution_id)
);

-- Table: Individual Client Signed Ballots (REG-02)
CREATE TABLE regulatory_compliance.client_proxy_ballots (
    id                      BIGSERIAL PRIMARY KEY,
    resolution_id           BIGINT NOT NULL REFERENCES regulatory_compliance.voting_resolutions(id),
    account_id              UUID NOT NULL,
    fractional_shares_held  NUMERIC(18, 6) NOT NULL CHECK (fractional_shares_held > 0),
    choice                  SMALLINT NOT NULL CHECK (choice IN (1, 2, 3)), -- 1: FOR, 2: AGAINST, 3: ABSTAIN
    eip712_signature        BYTEA NOT NULL,
    ballot_hash             BYTEA NOT NULL,
    submitted_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_account_resolution UNIQUE (resolution_id, account_id)
);

-- Table: Consolidated Depository Instruction Returns (REG-02)
CREATE TABLE regulatory_compliance.depository_vote_returns (
    id                      BIGSERIAL PRIMARY KEY,
    resolution_id           BIGINT NOT NULL REFERENCES regulatory_compliance.voting_resolutions(id),
    total_depository_shares BIGINT NOT NULL,
    allocated_for_shares    BIGINT NOT NULL,
    allocated_against_shares BIGINT NOT NULL,
    allocated_abstain_shares BIGINT NOT NULL,
    unvoted_shares          BIGINT NOT NULL,
    depository_ack_id       TEXT NOT NULL,
    besu_attestation_tx     TEXT NOT NULL,
    dispatched_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: Cross-Chain Finality & Deposit Tracker (REG-03)
CREATE TABLE regulatory_compliance.crosschain_deposit_tracker (
    id                      BIGSERIAL PRIMARY KEY,
    deposit_uuid            UUID NOT NULL UNIQUE,
    network                 VARCHAR(32) NOT NULL CHECK (network IN ('BITCOIN', 'ETHEREUM', 'SOLANA')),
    tx_hash                 TEXT NOT NULL UNIQUE,
    account_id              UUID NOT NULL,
    asset_symbol            VARCHAR(16) NOT NULL,
    amount_base_units       NUMERIC(38, 18) NOT NULL,
    observed_confirmations  INTEGER NOT NULL DEFAULT 0,
    required_confirmations  INTEGER NOT NULL,
    tier_level              SMALLINT NOT NULL DEFAULT 0, -- 0: Detected, 1: Fast, 2: Standard, 3: Institutional
    is_economic_finality    BOOLEAN NOT NULL DEFAULT FALSE,
    is_orphaned_by_reorg    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deposit_status 
    ON regulatory_compliance.crosschain_deposit_tracker (network, is_economic_finality, is_orphaned_by_reorg);

-- Table: Clawback Saga Audit Log (REG-03)
CREATE TABLE regulatory_compliance.clawback_saga_executions (
    id                      BIGSERIAL PRIMARY KEY,
    saga_uuid               UUID NOT NULL UNIQUE,
    deposit_id              BIGINT NOT NULL REFERENCES regulatory_compliance.crosschain_deposit_tracker(id),
    account_id              UUID NOT NULL,
    orphaned_tx_hash        TEXT NOT NULL,
    phase_reached           VARCHAR(32) NOT NULL,
    initial_deficit_paise   BIGINT NOT NULL,
    liquidated_value_paise  BIGINT NOT NULL DEFAULT 0,
    sgf_loss_absorbed_paise BIGINT NOT NULL DEFAULT 0,
    besu_burn_tx_hash       TEXT,
    ifsca_filing_ref        TEXT,
    completed_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### 5.3 Core Kafka Event Specifications

#### Event 1: `compliance.peak_margin.snapshot_taken.v1`
Emitted by `risk-service` immediately upon completing an intraday snapshot.
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "PeakMarginSnapshotTakenEvent",
  "type": "object",
  "required": [
    "eventId",
    "timestamp",
    "snapshotId",
    "windowIndex",
    "totalAccounts",
    "deficitCount",
    "merkleStateRoot"
  ],
  "properties": {
    "eventId": {"type": "string", "format": "uuid"},
    "timestamp": {"type": "string", "format": "date-time"},
    "snapshotId": {"type": "string", "format": "uuid"},
    "windowIndex": {"type": "integer", "minimum": 1, "maximum": 4},
    "totalAccounts": {"type": "integer"},
    "deficitCount": {"type": "integer"},
    "totalShortfallPaise": {"type": "integer"},
    "merkleStateRoot": {"type": "string", "pattern": "^0x[0-9a-fA-F]{64}$"},
    "isStatutorySession": {"type": "boolean"}
  }
}
```

#### Event 2: `compliance.proxy_vote.aggregated.v1`
Emitted by `corporate-actions` after executing the Hamilton-Hare apportionment.
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "ProxyVoteAggregatedEvent",
  "type": "object",
  "required": [
    "eventId",
    "isin",
    "meetingId",
    "resolutionId",
    "totalDepositoryShares",
    "allocatedFor",
    "allocatedAgainst",
    "allocatedAbstain"
  ],
  "properties": {
    "eventId": {"type": "string", "format": "uuid"},
    "isin": {"type": "string", "minLength": 12, "maxLength": 12},
    "meetingId": {"type": "string"},
    "resolutionId": {"type": "integer"},
    "totalDepositoryShares": {"type": "integer"},
    "allocatedFor": {"type": "integer"},
    "allocatedAgainst": {"type": "integer"},
    "allocatedAbstain": {"type": "integer"},
    "unvotedShares": {"type": "integer"},
    "depositoryPayloadSha256": {"type": "string", "pattern": "^[0-9a-fA-F]{64}$"}
  }
}
```

#### Event 3: `compliance.crosschain.reorg_clawback.v1`
Emitted by `risk-service` when a deep reorg triggers the 5-Phase Clawback Saga.
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CrossChainReorgClawbackEvent",
  "type": "object",
  "required": [
    "eventId",
    "sagaId",
    "network",
    "orphanedTxHash",
    "accountId",
    "phase",
    "currentDeficitPaise"
  ],
  "properties": {
    "eventId": {"type": "string", "format": "uuid"},
    "sagaId": {"type": "string", "format": "uuid"},
    "network": {"type": "string", "enum": ["BITCOIN", "ETHEREUM", "SOLANA"]},
    "orphanedTxHash": {"type": "string"},
    "accountId": {"type": "string", "format": "uuid"},
    "phase": {"type": "string", "enum": ["LOCKED", "REVALUED", "LIQUIDATED", "SGF_ABSORBED", "BURNED"]},
    "currentDeficitPaise": {"type": "integer"},
    "liquidatedProceedsPaise": {"type": "integer"},
    "sgfDrawdownPaise": {"type": "integer"}
  }
}
```

---

## 6. Cross-Document Traceability & Implementation Matrix

| Requirement Key | Architectural Subsystem | Implementation Files & Microservices | Smart Contracts / Ledgers | Runbook & Compliance Reference |
|---|---|---|---|---|
| **REG-01** | Peak Margin Snapshot Partitioning | [services/risk-service/](../../../services/risk-margin-service/), [services/reporting-service/](../../../services/regulatory-reporting-service/), [services/matching-engine/](../../../services/order-matching-engine/) | Hyperledger Besu Merkle State Roots, TimescaleDB Tables | [Runbook 08: SEBI Peak Margin Snapshot & Overnight Collateral Management](../runbooks/RUNBOOK-08-peak-margin-snapshot-and-overnight-collateral.md) |
| **REG-02** | Fractional Pass-Through Proxy Voting | [services/corporate-actions/](../../../services/corporate-actions-service/), [services/custody-adapter/](../../../services/custodian-depository-adapter/), [apps/growww_web/](../../../apps/growww_web/), [apps/growww_flutter/](../../../apps/growww_flutter/) | `contracts/governance/`, `contracts/tokens/DigitalSecurityToken.sol` | [Prompt 222: Corporate Actions Service](../../222_corporate_actions_service.md), Companies Act 2013 Sec 108 |
| **REG-03** | Tiered Cross-Chain Finality & Saga | [services/gift-city-funding/](../../../services/gift-city-funding-service/), [services/risk-service/](../../../services/risk-margin-service/), [services/settlement-service/](../../../services/trade-settlement-service/) | `contracts/bridge/CrossEntityBridge.sol`, Hyperledger Besu QBFT | [Runbook 07: Cross-Chain Reorganization & Clawback Saga Execution](../runbooks/RUNBOOK-07-crosschain-reorg-and-clawback.md), IFSCA Regulations |
