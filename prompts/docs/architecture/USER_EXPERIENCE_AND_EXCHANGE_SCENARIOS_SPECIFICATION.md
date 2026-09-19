# User Experience, Trading Workflows & Exchange Scenarios Specification

**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Product Engineering & Exchange Operations Architecture Board  
**Effective Date:** September 2026  
**Review Cadence:** Quarterly  
**Classification:** Technical & Operational Architectural Blueprint  

---

## 1. Architectural Overview & Persona Topology

### 1.1 Scope, Objectives & Product Segments
The National Blockchain Stock Exchange (NBSE) and Growww Trading Platform provide a high-throughput, institutional-grade, multi-asset digital exchange architecture designed to service both domestic Indian retail investors and global institutional market participants. This specification establishes the end-to-end technical, algorithmic, and user-experience standards governing:

1. **User Onboarding & Identity Governance:** Automated identity verification compliant with SEBI KYC Master Circulars, the Prevention of Money Laundering Act (PMLA 2002), and the Digital Personal Data Protection (DPDP) Act 2023.
2. **Multi-Rail Funding:** Real-time domestic fiat settlement via RBI-regulated payment rails and multi-chain cryptographic asset transfers with automated confirmation tracking and FATF Travel Rule validation.
3. **Execution Paradigms & Trading Workspaces:** Unified execution across high-performance Central Limit Order Books (CLOB), zero-slippage Request-For-Quote (RFQ) Convert engines, and multi-monitor Institutional Pro Workspaces.
4. **Order Execution Microstructure:** Microsecond-deterministic execution across simple and advanced algorithmic order types (Market with slippage protection, Limit, Stop, Trailing Stop, Midpoint Peg, OCO, Bracket Orders).
5. **Withdrawal Security:** Pessimistic balance reserving, cryptographic address book whitelisting, 12-hour cooling windows, and Multi-Party Computation (MPC) threshold signing.
6. **Institutional Sub-Accounts & Developer APIs:** Hierarchical account management, Ed25519/HMAC-SHA256 authenticated REST/WebSocket interfaces, and multi-tier leaky-bucket rate limiting.
7. **Statutory Tax & Portfolio Engine:** Mark-to-Market portfolio analytics, FIFO and Weighted Average Cost tax-lot accounting, real-time portfolio analytics, exportable PnL statements, and zero on-chain tax withholding.
8. **Failure Recovery & Edge Case Protocols:** Resilient automated handling for tagless memo chains, blockchain reorganizations, in-band WebSocket re-authentication, and instant emergency account self-lock kill switches.

### 1.2 User Personas & Permissions Topology

```
+-------------------------------------------------------------------------------------------------------------+
|                                    PLATFORM USER PERSONA TOPOLOGY                                           |
|                                                                                                             |
|  +------------------------+  +------------------------+  +------------------------+  +-------------------+  |
|  |     RETAIL DOMESTIC    |  |     HIGH-NET-WORTH     |  |   INSTITUTIONAL DESK   |  | GIFT CITY FOREIGN |  |
|  |       INVESTOR         |  |      ACTIVE TRADER     |  |   & MARKET MAKER       |  |     INVESTOR      |  |
|  +------------------------+  +------------------------+  +------------------------+  +-------------------+  |
|  | * Web / Mobile UI      |  | * Pro Trading Desktop  |  | * FIX 4.4 / gRPC / WS  |  | * LRS Inflow Rail |  |
|  | * ₹100 Min Investment  |  | * Advanced Multi-Chart |  | * Sub-Accounts (100+)  |  | * USD/EUR/AED Clr |  |
|  | * UPI / Instant IMPS   |  | * Bracket / OCO Orders |  | * 5,000 req/s API Tier |  | * IFSCA FME Reg.  |  |
|  | * RFQ Instant Swap     |  | * Depth Ladder (DOM)   |  | * Ed25519 Signed Auth  |  | * FATF Travel R.  |  |
|  | * In-App Tax P&L       |  | * Custom Layout & Keys |  | * Zero-Withdrawal Scope|  | * Multi-Currency  |  |
|  +------------------------+  +------------------------+  +------------------------+  +-------------------+  |
+-------------------------------------------------------------------------------------------------------------+
```

1. **Retail Domestic Investor:** Accesses mobile (Flutter) and standard web interfaces; operates primarily in INR spot trading, fractionated securities tokens, and zero-slippage RFQ swap modules; utilizes UPI 2.0 and IMPS banking rails.
2. **High-Net-Worth / Active Pro Trader:** Operates on advanced browser/desktop workspaces utilizing multi-window docks, Level-2/Level-3 depth ladders (DOM), complex algorithmic orders (Bracket, Trailing Stops), and high-frequency alerting.
3. **Institutional Prop Desk / Market Maker:** Executes programmatic orders via FIX 4.4, gRPC, and high-frequency WebSockets; utilizes Master/Sub-Account hierarchies with strict operational permission scoping; requires sub-millisecond round-trip latencies and dedicated rate-limit tiers up to 5,000 requests per second.
4. **GIFT City Foreign Investor:** Interacts through the International Gateway Entity regulated by IFSCA; funds via cross-border wire transfers (SWIFT, Fedwire) or the RBI Liberalised Remittance Scheme (LRS); trades in USD and synthetic cross-border pairs under FATF Travel Rule compliance.

### 1.3 High-Level System Architecture Flow

```
                                  +-----------------------+
                                  | Client Applications   |
                                  | (Flutter / Next.js /  |
                                  | FIX / High-Speed SDK) |
                                  +-----------+-----------+
                                              |
                                              | HTTPS / WSS / FIX
                                              v
                                  +-----------------------+
                                  |   API Gateway & BFF   |
                                  | (Envoy / In-Band Auth |
                                  | Leaky-Bucket Limiter) |
                                  +-----------+-----------+
                                              |
                   +--------------------------+--------------------------+
                   |                                                     |
                   v                                                     v
       +-----------------------+                             +-----------------------+
       |   User & Compliance   |                             | Matching Engine Core  |
       | - DigiLocker / Aadhaar|                             | (Rust In-Memory CLOB) |
       | - PAN / CKYC Fetch    |                             | - Nano-Sequencer      |
       | - Video KYC & Liveness|                             | - Level-2/3 DOM       |
       | - AML / Travel Rule   |                             | - SPSC Ring Buffer    |
       +-----------+-----------+                             +-----------+-----------+
                   |                                                     |
                   | State Events                                        | Order Fills & BBO
                   v                                                     v
+------------------------------------------------------------------------------------+
|                               Apache Kafka Streaming Fabric                         |
|   `order.events.v1`  |  `trade.settlement.v1`  |  `marketdata.depth.v1`  |  `kyc`   |
+------------------------------------------------------------------------------------+
       |                                       |                               |
       v                                       v                               v
+--------------+                       +---------------+               +---------------+
| Double-Entry |                       | Ledger Bridge |               | Notification  |
| Journal Core |                       | & Settlement  |               | & Tax Service |
| (PostgreSQL) |                       | (Besu / DvP)  |               | (APNS/TDS 194)|
+--------------+                       +---------------+               +---------------+
```

For core ledger definitions and matching engine mechanics, see [Core Ledger Specification](CORE_LEDGER_AND_MULTI_CURRENCY_SPECIFICATION.md) and [Matching Engine Specification](CORE_MATCHING_ENGINE_AND_ORDER_BOOK_SPECIFICATION.md).

---

## 2. Complete User Journey Scenarios: Onboarding & Identity Verification

### 2.1 Multi-Stage Identity Verification Pipeline Overview
The onboarding pipeline guarantees zero-friction user acquisition while satisfying the stringent mandates of PMLA 2002, SEBI KYC Master Directives, and the DPDP Act 2023. Domestic Indian onboarding is structured into five deterministic, sequential validation gates.

```
+---------------------------------------------------------------------------------------------------+
|                                  DOMESTIC ONBOARDING PIPELINE FLOW                                |
|                                                                                                   |
|  [Step 1]           [Step 2]           [Step 3]           [Step 4]           [Step 5]             |
|  PAN Verification   DigiLocker / CKYC  Biometric Liveness Video KYC (V-KYC)  Bank Penny Drop      |
|  +---------------+  +---------------+  +---------------+  +---------------+  +---------------+    |
|  | NSDL/UTIITSL  |  | Aadhaar OTP / |  | Passive Anti- |  | Live Officer  |  | UPI / IMPS Re1|    |
|  | Active Status |->| CERSAI 14-Dig |->| Spoofing 3D   |->| Geolocation   |->| Jaro-Winkler  |    |
|  | PAN-Aadhaar   |  | 8-Digit Mask  |  | Micro-Express |  | Document OCR  |  | Score >= 85%  |    |
|  +---------------+  +---------------+  +---------------+  +---------------+  +---------------+    |
|                                                                                      |            |
|                                                                                      v            |
|                                                                              [Account Activated]  |
|                                                                              Besu KYC Commitment  |
+---------------------------------------------------------------------------------------------------+
```

### 2.2 PAN Verification & Tax Residency Validation
All domestic users must provide a 10-character Permanent Account Number (PAN).
1. **Format Validation:** Regular expression match: `^[A-Z]{5}[0-9]{4}[A-Z]{1}$`. Character 4 must be `'P'` (Individual), `'C'` (Company), `'H'` (HUF), or `'F'` (Partnership).
2. **Government Registry Lookup:** Invokes the NSDL / UTIITSL validation API via an encrypted secure channel.
3. **Validation Requirements:**
   - PAN status must be returned as `OPERATIVE`.
   - `pan_aadhaar_linked` flag must equal `TRUE`. Under Indian tax law, transactions originating from an inoperative or unlinked PAN are legally void.
   - User date of birth must verify the applicant is aged 18 or older. Minors are strictly prohibited from holding trading accounts.

### 2.3 Central KYC Registry (CKYCR) Automated Fetch
To deliver a zero-upload onboarding experience for previously verified citizens, the system queries the Central Registry of Securitisation Asset Reconstruction and Security Interest of India (CERSAI):
1. **Search Request:** System sends a query with PAN and Date of Birth to the CKYCR API gateway.
2. **Record Extraction:** If a valid 14-digit CKYC identifier (`CKYC_IN`) is retrieved, the service extracts the certified document bundle (Identity proof, Address proof, photograph).
3. **Data Freshness Check:** If the CKYC record was updated within 24 months and the user's current residential address matches, manual document upload steps are completely bypassed.

### 2.4 DigiLocker Aadhaar e-Sign & XML Verification
When CKYC records are unavailable or outdated, the user is routed to the paperless UIDAI DigiLocker flow:
1. **Consent Redirection:** The user is redirected to the MeitY-authorized DigiLocker consent gateway.
2. **UIDAI OTP Verification:** The user inputs their 12-digit Aadhaar number and confirms the time-based one-time password (OTP) dispatched to their UIDAI-registered mobile number.
3. **Aadhaar Data Vault Invariant:**
   - The first 8 digits of the Aadhaar number are cryptographically masked (`XXXX-XXXX-1234`) across all client UI screens, application logs, database tables, and analytical pipelines.
   - Raw XML documents are unbundled in an ephemeral in-memory worker; name, date of birth, gender, and address are extracted.
   - The plain Aadhaar number is replaced with a universally unique token (`aadhaar_token`) generated by the HSM-backed Aadhaar Data Vault compliant with UIDAI Circular 2017/01.

### 2.5 Biometric Liveness Detection & Passive Anti-Spoofing
To eliminate injection attacks, deepfakes, and static photograph impersonation:
1. **Liveness Model:** The client-side Flutter/Web SDK captures a high-resolution video stream processed through a WebAssembly / native neural engine conforming to ISO/IEC 30107-3 Level 2 (Presentation Attack Detection).
2. **Passive Challenges:**
   - 3D depth-map triangulation through subtle multi-frame lighting variations.
   - Eye-blink rate detection and involuntary ocular micro-saccades.
   - Active challenge response (e.g. asking user to tilt head slightly or read three random single-digit numbers displayed on screen).
3. **Liveness Score Threshold:** A confidence metric $\ge 0.95$ is strictly required. Any frame exhibiting video loop artifacts or digital mask borders triggers an immediate fraud quarantine.

### 2.6 Video KYC (V-KYC) Workflow
For High-Risk categorizations or users whose DigiLocker/CKYC confidence falls below deterministic thresholds, a real-time Video KYC session is conducted:
1. **Infrastructure:** Encrypted WebRTC peer-to-peer connection terminating in a secured compliance recording cluster.
2. **Officer Coordination:** The user is paired with a SEBI-certified compliance officer.
3. **In-Call Verification Checklist:**
   - Automated IP geolocation gating: User coordinates must resolve to sovereign Indian territory. GPS coordinates are captured and cross-referenced with IP BGP routing.
   - Physical PAN card display: The user displays their physical PAN card. The officer client triggers high-resolution OCR, validating holographic micro-print patterns and matching text fields.
   - Dynamic Question Challenge: The officer prompts the user with dynamic security questions derived from credit bureau history (e.g. historical bank provider, birth city).
   - Audio-Visual Recording: The complete session is encrypted with AES-256-GCM and archived in cold WORM (Write Once Read Many) storage with a 10-year statutory retention policy.

### 2.7 Bank Account Verification via Penny Drop
To ensure zero third-party deposit fraud, the investor's designated bank account is verified before any funds are accepted:
1. **Penny Drop Rail:** The system dispatches an instant ₹1.00 transfer via NPCI Immediate Payment Service (IMPS) or Unified Payments Interface (UPI) into the target account.
2. **Beneficiary Extraction:** The response from the destination bank provides the exact legal name registered against the target bank account (`beneficiary_name_raw`).
3. **Fuzzy Name Matching Algorithm:** The extracted name is compared against the name obtained from PAN and Aadhaar records using the Jaro-Winkler distance and Double Metaphone phonetic encoding.

$$\text{Sim}(s_1, s_2) = \text{JaroWinkler}(s_1, s_2)$$

$$\text{NameMatchCondition} = \begin{cases} 
\text{APPROVED} & \text{if } \text{Sim}(s_1, s_2) \ge 0.85 \\
\text{MANUAL\_REVIEW} & \text{if } 0.70 \le \text{Sim}(s_1, s_2) < 0.85 \\
\text{REJECTED} & \text{if } \text{Sim}(s_1, s_2) < 0.70 
\end{cases}$$

### 2.8 Onboarding State Machine & Ledger Whitelisting
Once all gates pass successfully, the onboarding state machine transitions the user:

```
[INITIATED] -> [PAN_VERIFIED] -> [AADHAAR_LINKED] -> [LIVENESS_PASSED] -> [BANK_VERIFIED] -> [ACTIVE]
```

Upon reaching `ACTIVE` status:
- An identity commitment hash is generated: 
  $$\text{Commitment} = \text{keccak256}(\text{pan\_hash} \parallel \text{investor\_uuid} \parallel \text{system\_salt})$$
- The 32-byte commitment is written to `ComplianceRegistry.sol` on the Hyperledger Besu consortium network, permitting atomic DvP security token trading (see [ADR-0005](../adr/ADR-0005-erc3643-permissioned-security-tokens.md)).

---

## 3. Complete User Journey Scenarios: Multi-Rail Deposit Flows

### 3.1 Instant Fiat Deposit Infrastructure (Domestic INR)
Fiat funding is strictly restricted to RBI-regulated banking rails terminating in segregated nodal escrow accounts.

```
+---------------------------------------------------------------------------------------------------+
|                                  DOMESTIC FIAT DEPOSIT RAILS                                      |
|                                                                                                   |
|  [UPI 2.0 AutoPay]          [Dynamic QR Code]             [IMPS / RTGS Virtual Account]           |
|  - Instant deep-link        - Displayed on Web/Terminal   - Dedicated ICICI/HDFC VAN ID           |
|  - Mandate pre-approval     - Scanned by any UPI app      - UTR webhook reconciliation            |
|  - Settlement: < 5s         - Settlement: < 5s            - Settlement: 30s - 3 min               |
|  - Max per txn: ₹2,00,000   - Max per txn: ₹2,00,000      - Unlimited institutional volume        |
+---------------------------------------------------------------------------------------------------+
```

#### 3.1.1 UPI 2.0 AutoPay & Dynamic QR Flows
1. **UPI Intent (Mobile):** The mobile application invokes the native NPCI UPI Intent chooser (`upi://pay?pa=growww.nodal@icici&pn=Growww&am=50000.00&tr=TXN10293847&tn=Deposit...`). The user selects their registered banking application (Google Pay, PhonePe, BHIM, Paytm).
2. **Dynamic QR (Web):** The web terminal generates a unique, single-use SVG/PNG dynamic QR code containing the cryptographically signed transaction reference.
3. **Webhook Verification:** The banking aggregator dispatches an HMAC-SHA256 signed webhook. The backend validates signature integrity, checks replay prevention cache in Redis (`pay:idemp:{ref}`), and posts a double-entry deposit credit to the investor's balance.

#### 3.1.2 Dedicated Virtual Account Number (VAN) Routing (IMPS/NEFT/RTGS)
For high-value transfers exceeding standard UPI limits (up to ₹10,00,00,000):
1. **VAN Generation:** Each investor is assigned a persistent, unique Virtual Account Number mapped to the platform's nodal bank partnership:
   `GROWWW` + `7-digit alphanumeric user identifier` (e.g. `GROWWW7849201`, IFSC: `ICIC0000104`).
2. **Bank Clearing Notification:** When the user transfers funds from their verified bank account via RTGS/NEFT/IMPS, the clearing bank pushes an API notification with the Unique Transaction Reference (UTR).
3. **Automated Name Reconciliation:** The remitters bank account and name are validated against the user's KYC record. Upon confirmation, the deposit is credited in $< 30$ seconds.

#### 3.1.3 Strict Third-Party Deposit Rejection Policy
Under SEBI circular SEBI/HO/MIRSD/DOS3/CIR/P/2018/115, any deposit received from a bank account not registered and verified to the user is rejected:
- Funds are intercepted prior to ledger credit.
- An automated reverse clearing instruction is generated.
- Capital is refunded to the originating source account within 2 banking hours with the reversal reason code `REJ_3RD_PARTY_NAME_MISMATCH`.

### 3.2 Crypto Multi-Chain Deposit Architecture
For digital assets, the platform implements a non-custodial hierarchical deterministic deposit pipeline paired with real-time on-chain confirmation tracking (see [ADR-0030](../adr/ADR-0030-multi-chain-hd-wallet-deposit-derivation.md)).

```
+---------------------------------------------------------------------------------------------------+
|                                 CRYPTO DEPOSIT CONFIRMATION TRACKER                               |
|                                                                                                   |
|  Asset: ETH (Ethereum Mainnet)                                                                    |
|  TxHash: 0x8f2a...9c1d                                                                            |
|  Amount: 2.50000000 ETH                                                                           |
|                                                                                                   |
|  [========================== 64/64 Confirmations ==========================] 100%               |
|                                                                                                   |
|  Stage 1: Detected         Stage 2: Casper Finalized     Stage 3: Risk Screened   Stage 4: Credit |
|  [ ✓ ] Block #20784912     [ ✓ ] 2 Epochs Finality       [ ✓ ] TRM Risk Score 12  [ ✓ ] Available |
+---------------------------------------------------------------------------------------------------+
```

#### 3.2.1 Deterministic HD Address Derivation Standards
Addresses are derived on demand using FIPS 140-3 Level 4 HSM-stored master public keys ($O(1)$ derivation without exposing private keys):
- **Bitcoin (BTC):** BIP-84 Native SegWit Bech32 derivation: `m/84'/0'/0'/0/{user_index}`.
- **Ethereum & EVM L2s (Base, Arbitrum, Polygon):** BIP-44 standard: `m/44'/60'/0'/0/{user_index}`.
- **Solana (SOL & SPL Tokens):** BIP-44 Ed25519 standard: `m/44'/501'/0'/{user_index}'`.
- **TRON (TRC-20 USDT):** BIP-44 standard: `m/44'/195'/0'/0/{user_index}`.

#### 3.2.2 Live Confirmation Tracker & State Transitions
The frontend WebSocket stream connects to `wallet-service` and provides a live progress tracker:

| Asset | Rail | Blocks for Detection | Blocks for Unlocked Trading | Finality Standard (ADR-0010) |
| :--- | :--- | :--- | :--- | :--- |
| **BTC** | Bitcoin Native | 0-conf (Mempool) | 3 confirmations | 6 blocks (~60 min) |
| **ETH** | Ethereum POS | 1 block (Unfinalized) | 32 blocks (1 epoch) | 64 blocks (2 epochs, ~12.8 min) |
| **USDC** | Arbitrum One | 1 batch | 20 L2 blocks | L1 Data Poster Finality (~15 min) |
| **SOL** | Solana SPL | 1 slot (Processed) | 16 slots (Confirmed) | 32 slots (Finalized, ~13 sec) |
| **USDT** | TRON TRC-20 | 1 block | 12 blocks | 19 blocks (~57 sec) |

#### 3.2.3 Blockchain AML Risk Scoring & Quarantining
Prior to crediting the user balance, the incoming transaction hash and sender address are evaluated via real-time blockchain intelligence APIs (Chainalysis KYT / TRM Labs) conforming to [ADR-0032](../adr/ADR-0032-travel-rule-and-blockchain-aml-risk-scoring.md):
- If risk score $< 25$: Immediate automated deposit credit.
- If risk score between $25$ and $65$: Account flagged for enhanced monitoring; withdrawal freeze initiated pending manual compliance review.
- If risk score $> 65$ or associated with OFAC sanctioned entities, illicit mixers (Tornado Cash), ransomware, or darknet markets: Funds are routed to a segregated quarantine ledger; an automated Suspicious Transaction Report (STR) is queued for FIU-IND FinGate.

### 3.3 Cross-Border GIFT City LRS & Foreign Funding
For international trading desks and Indian residents investing in foreign assets through the International Financial Services Centres Authority (IFSCA) regulatory framework:
1. **LRS Remittance Rails:** Indian residents remit capital under the RBI Liberalised Remittance Scheme (LRS) up to $\$250,000$ USD per financial year.
2. **Form A2 & Tax Collected at Source (TCS):**
   - The platform generates automated digital Form A2 and Section 206C(1G) declarations.
   - For aggregate remittances exceeding ₹7,00,000 in a financial year, the statutory 20% TCS is computed and withheld for onward remittance to the tax authorities.
3. **SWIFT / Fedwire Gateway:** International institutional funds arrive directly into segregated multi-currency custody accounts (USD, EUR, GBP, AED) hosted at GIFT City IFSC scheduled banks. Real-time SWIFT MT103 and ISO 20022 `pacs.008` messages trigger instant multi-currency ledger credits.

---

## 4. Complete User Journey Scenarios: Trading Modes & Workspaces

### 4.1 Standard Retail Spot Trading Workspace
Designed for visual clarity, safety, and sub-second execution responsiveness for retail traders.

```
+-------------------------------------------------------------------------------------------------------------+
|                                    STANDARD SPOT TRADING WORKSPACE                                          |
|                                                                                                             |
|  +--------------------------------------------------+  +----------------------+  +-----------------------+  |
|  | TradingView Advanced Chart Canvas                |  | Level-2 Order Book   |  | Order Execution Slip  |  |
|  | Pair: BTC/INR | Interval: 15m                    |  | Price      Size      |  | [BUY]     [SELL]      |  |
|  | 5,420,000 | +2.45% | Vol: 184.22 BTC             |  | 5,420,500  0.4512    |  +-----------------------+  |
|  |                                                  |  | 5,420,400  1.1200    |  | Type: [ Limit | Mkt ] |  |
|  |     /\    /\                                     |  | 5,420,200  0.8950    |  | Price:   5,420,000    |  |
|  |    /  \  /  \    /\                              |  | -------------------- |  | Amount:  0.0500 BTC   |  |
|  |        \/    \  /  \                             |  | Spread:    100 INR   |  | [ 25% | 50% | 100% ]  |  |
|  |               \/                                 |  | -------------------- |  | Total:   ₹2,71,000    |  |
|  | Indicators: RSI(14): 58.2 | EMA(200): 5,380,000  |  | 5,420,100  2.1000    |  | Fee:     ₹27.10 (0.00% (Zero Fee))|  |
|  |                                                  |  | 5,420,000  3.4500    |  | TDS:     ₹2,710 (1%)  |  |
|  +--------------------------------------------------+  +----------------------+  +-----------------------+  |
|  | Open Orders (2) | Trade History | Realized P&L: +₹14,250.00 | Net Holdings: 0.1250 BTC               |  |
+-------------------------------------------------------------------------------------------------------------+
```

1. **TradingView Integration:** Embedded TradingView Charting Library supporting 1m to 1M intervals, multi-pane technical indicators (RSI, MACD, Bollinger Bands, Volume Profile), and persistent user drawings synced across devices via cloud state storage.
2. **Level-2 DOM & Depth Visualization:** Aggregated market depth ladder updated via high-conflation WebSockets (100ms conflation buffers, see [ADR-0024](../adr/ADR-0024-websocket-slow-consumer-conflation-backpressure.md)). Displays cumulative bid/ask walls and visual depth bars.
3. **Execution Slip:** Single-click switching between Buy and Sell modes, percentage balance selector buttons (25%, 50%, 75%, 100%), real-time fee preview (0.00% standard fee - No fee at all), and zero on-chain TDS confirmation.

### 4.2 Quick Convert / Instant Swap (Zero-Slippage RFQ Engine)
Retail users requiring simple, instantaneous conversions without navigating order books utilize the zero-slippage RFQ Swap Engine conforming to [ADR-0033](../adr/ADR-0033-rfq-instant-convert-swap-engine.md).

```
+---------------------------------------------------------------------------------------------------+
|                                 QUICK CONVERT / INSTANT SWAP (RFQ)                                |
|                                                                                                   |
|  From Asset:  [ 1,000.00000000 USDT ]      Available: 4,500.00 USDT                               |
|                                                                                                   |
|                               [   ⇅ Invert Assets   ]                                             |
|                                                                                                   |
|  To Asset:    [ 86,520.00 INR        ]     Guaranteed Conversion Rate: 1 USDT = 86.52 INR         |
|                                                                                                   |
|  +---------------------------------------------------------------------------------------------+  |
|  | Quote Lock: 00:05.42 [======    ] | Guaranteed Slippage: 0.00% | Platform Fee: 0.00% (Zero Fee) (Incl) |  |
|  +---------------------------------------------------------------------------------------------+  |
|                                                                                                   |
|                                  [ CONFIRM SWAP (4s REMAINING) ]                                  |
+---------------------------------------------------------------------------------------------------+
```

#### 4.2.1 Two-Way Synthetic Pricing & 6.00-Second Quote Guarantee
1. **Quote Request:** The client dispatches a pricing request:
   `POST /api/v1/convert/quote` (`from_asset: "USDT"`, `to_asset: "INR"`, `amount: "1000.00"`).
2. **Synthetic Pricing Aggregation:** The RFQ pricing engine interrogates internal CLOB order books and external liquidity providers to calculate a guaranteed executable price inclusive of the 0.00% (Zero Fee) platform fee.
3. **Cryptographic Quote Token:** The gateway responds with a signed quote payload:
   ```json
   {
     "quote_id": "rfq_89f02938a1c9",
     "from_asset": "USDT",
     "to_asset": "INR",
     "from_amount": "1000.00000000",
     "to_amount": "86520.00",
     "rate": "86.52000000",
     "expires_at_ms": 1789824006000,
     "signature": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
   }
   ```
4. **Execution Invariant:** If confirmed within the 6.00-second window, execution occurs atomically at the exact quoted price with 0.00% slippage. Price movement occurring during the 6-second window is absorbed by the platform's internal automated inventory hedging engine.

### 4.3 Advanced Institutional Pro Workspace
For high-volume desks requiring multi-asset monitoring, sub-millisecond execution, and customizable desktop screen layouts.

```
+-------------------------------------------------------------------------------------------------------------+
|                                  INSTITUTIONAL PRO WORKSPACE (MULTI-DOCK)                                   |
|                                                                                                             |
|  +--------------------------------+  +-------------------------------+  +--------------------------------+  |
|  | Chart 1: BTC/INR (1m)          |  | Chart 2: ETH/INR (5m)         |  | Level-3 Raw DOM Order Book     |  |
|  | High-speed tick stream         |  | Breakout volume profile       |  | Size    Orders    Price        |  |
|  |                                |  |                               |  | 1.4500     2      5,420,500    |  |
|  +--------------------------------+  +-------------------------------+  | 0.8900     1      5,420,400    |  |
|  | Algorithmic Execution Slip     |  | Real-Time Margin Telemetry    |  | 3.1200     4      5,420,300    |  |
|  | Mode: Institutional Midpoint   |  | IM Required:  ₹1,24,50,000    |  | ------------------------------ |  |
|  | Hotkeys: Shift+B (Buy)         |  | Collateral:   ₹4,80,00,000    |  | Spread: 100 INR | Micro-ticks  |  |
|  | Hotkeys: Shift+S (Sell)        |  | Margin Ratio: 25.93% (HEALTHY)|  | ------------------------------ |  |
|  | Hotkeys: Esc (Cancel All)      |  | Haircut Adj:  ₹4,32,00,000    |  | 2.5000     3      5,420,200    |  |
|  +--------------------------------+  +-------------------------------+  +--------------------------------+  |
|  | Positions | Open Fills | Routing Desk Logs | FIX Drop Copy Status: Connected | Latency: 420us            |  |
+-------------------------------------------------------------------------------------------------------------+
```

1. **Multi-Window Docking Grid:** Built on an asynchronous web docking framework (Dockview) allowing individual chart panels, depth ladders, and position tables to be undocked into independent OS windows across multiple monitor arrays.
2. **Deterministic Keyboard Hotkeys:** Zero-mouse execution bindings:
   - `Shift + B`: Instant Buy at BBO (Best Bid Offer).
   - `Shift + S`: Instant Sell at BBO.
   - `Space`: Flatten active position via Aggressive Market IOC.
   - `Escape`: Cancel all active open orders across the selected instrument.
   - `1` through `5`: Instant lot size presets ($0.1, 0.5, 1.0, 5.0, 10.0$).
3. **Level-3 Market By Order (MBO) Feed:** Institutional desks receive un-aggregated Level-3 order updates showing discrete individual queue priority positions and order queue IDs.

---

## 5. Complete User Journey Scenarios: Order Execution Scenarios

### 5.1 Limit Orders & Time-In-Force Qualifiers
A Limit order buys or sells an instrument at a specified price or better:
1. **Good 'Til Canceled (GTC):** The order remains in the matching engine order book until fully filled or explicitly canceled by the user.
2. **Immediate Or Cancel (IOC):** Any portion of the order that can be filled immediately against resting liquidity at the limit price or better is matched; any remaining unfilled quantity is instantaneously canceled without entering the book.
3. **Fill Or Kill (FOK):** The order must be executed immediately in its entirety against resting liquidity; if the complete quantity cannot be satisfied, the entire order is canceled with zero fills.
4. **Post-Only (Maker-or-Cancel):** Guarantees the order will only add liquidity to the order book. If the order price would cross an existing resting opposite order, the matching engine cancels the order immediately, preventing taker fee charges.

### 5.2 Market Orders with Max Slippage Bounds
To protect users from catastrophic execution prices during flash crashes or illiquid market books:
1. **Slippage Bounding Mechanism:** Pure unrestricted market orders are disallowed. All market orders are automatically wrapped with a mandatory maximum price slippage parameter ($\Delta_{\text{max}} = 1.00\%$ default for retail, configurable up to $5.00\%$ for institutional desks).
2. **Calculation Formula:**
   - **Market Buy Order:** Effective limit cap:
     $$P_{\text{cap}} = P_{\text{BBO\_Ask}} \times (1 + \Delta_{\text{max}})$$
   - **Market Sell Order:** Effective limit floor:
     $$P_{\text{floor}} = P_{\text{BBO\_Bid}} \times (1 - \Delta_{\text{max}})$$
3. **Partial Execution:** The engine matches through the order book until either the quantity is satisfied or the price reaches $P_{\text{cap}}$ / $P_{\text{floor}}$. Any remaining balance is canceled via IOC mechanics, preventing market manipulation slippage.

### 5.3 Stop-Loss Orders (LTP Trigger vs Index Trigger)
Stop orders remain dormant outside the active matching book until market prices cross a designated trigger price:
1. **Trigger References:**
   - **Last Traded Price (LTP):** Triggers upon the most recent trade executed in the local matching engine.
   - **Mark / Index Price:** Triggers upon the volume-weighted median index price calculated across multiple external exchanges, preventing localized wick-induced trigger cascades.
2. **Stop-Loss Market:** When $P_{\text{market}} \le P_{\text{trigger}}$ (for long exits), dispatches a Market Sell Order with slippage protection.
3. **Stop-Loss Limit:** When $P_{\text{market}} \le P_{\text{trigger}}$, injects a Limit Sell Order at $P_{\text{limit}}$ into the matching engine.

### 5.4 Trailing Stop-Loss Orders
Trailing stops allow traders to lock in profits while limiting maximum downside risk by automatically ratcheting the trigger price alongside favorable price movements.

```
Price (INR)
   ^                                      /\ Peak: ₹5,600,000
   |                                     /  \
   |                      /\            /    \   Trailing Stop Trigger
   |                     /  \          /      \- - - - - - - - - - - - - - - - (TRIGGERED!)
   |         /\         /    \        /        \  Ratcheted Stop: ₹5,488,000
   |        /  \       /      \      /          \
   |       /    \     /        \    /
   |      /      \   /          \  /
   |     /        \ /            \/
   |    /          V
   |   /  Trailing Stop: ₹5,096,000 (2% below peak)
   |  /
   | / Entry: ₹5,200,000
   +-----------------------------------------------------------------------------------> Time
```

1. **Parameters:**
   - Base Position: Long or Short.
   - Trailing Offset: Defined either as a percentage ($\delta_{\%} = 2.00\%$) or an absolute rupee tick amount ($\Delta_{\text{INR}} = ₹10,000$).
2. **Algorithmic Ratchet Engine:**
   - For a Long position, the system tracks the highest price reached since order placement ($P_{\text{peak}}$).
   - If current price $P(t) > P_{\text{peak}}$, then:
     $$P_{\text{peak}} \leftarrow P(t)$$
     $$P_{\text{trigger}} = P_{\text{peak}} \times (1 - \delta_{\%})$$
   - The trigger price only moves upward; it never decreases.
   - If market price falls to $P(t) \le P_{\text{trigger}}$, the trailing stop triggers an exit market order.

### 5.5 Passive Midpoint Peg Orders
Institutional participants requiring execution without crossing the bid-ask spread employ passive midpoint pegging:
1. **Dynamic Re-pricing:** The order automatically floats at the exact mathematical midpoint of the Best Bid and Offer (BBO):
   $$P_{\text{peg}} = \left\lfloor \frac{P_{\text{BBO\_Bid}} + P_{\text{BBO\_Ask}}}{2 \times \text{TickSize}} \right\rfloor \times \text{TickSize}$$
2. **Priority Rule:** Midpoint orders sit in a secondary execution queue behind resting displayed orders at that price level, providing liquidity without price impact.
3. **Execution Invariant:** If the spread narrows to 1 tick, the peg order becomes dormant or reprioritizes based on pre-set offset parameters.

### 5.6 One-Cancels-the-Other (OCO) Orders
An OCO order couples two contingent orders together: typically a take-profit limit order and a stop-loss order.
1. **Structure:**
   - **Leg 1:** Limit Order at $P_{\text{take\_profit}}$ (above current market for longs).
   - **Leg 2:** Stop Order with trigger at $P_{\text{stop\_trigger}}$ and execution at $P_{\text{stop\_limit}}$ (below current market for longs).
2. **Atomic Invariant:**
   - Both legs share a unified reservation lock on the user's available asset balance. Balance is reserved once, avoiding double allocation.
   - If Leg 1 fills or partially fills, Leg 2 is atomically canceled.
   - If the trigger price for Leg 2 is breached, Leg 1 is instantly canceled from the active book, and Leg 2 is submitted as an active limit/market order.

### 5.7 Bracket Orders
A Bracket Order allows complete trade lifecycle automation for active intraday traders by encapsulating an initial entry order alongside two contingent exit orders:
1. **Entry Leg:** Can be Limit or Market.
2. **Upper Exit Leg (Take-Profit):** Placed as a resting Limit Order as soon as the Entry Leg fills.
3. **Lower Exit Leg (Stop-Loss):** Placed as a Stop-Loss (Fixed or Trailing) as soon as the Entry Leg fills.
4. **Execution Cycle:**
   - If the Entry Leg is canceled before execution, both exit child legs are deleted.
   - Upon fill of the Entry Leg, the position quantity is assigned to the exit legs.
   - The exit legs form an OCO relationship: execution of the Take-Profit leg cancels the Stop-Loss leg and vice-versa.

### 5.8 Complete Order Lifecycle State Machine

```
               [NEW]
                 |
                 v (Validation & Risk Checks Passed)
             [ACCEPTED]
                 |
        +--------+--------+
        |                 | (Unconditional IOC/FOK miss)
        v                 v
    [ROUTED]         [REJECTED]
        |
   +----+----+
   |         | (Resting in CLOB)
   v         v
[FILLED]  [PARTIALLY_FILLED]
             |
             +------------+
             |            |
             v            v
         [FILLED]    [CANCELED]
```

State transitions are streamed over Apache Kafka on the `order.events.v1` topic with strict event sourcing determinism (see [ADR-0027](../adr/ADR-0027-order-cancellation-matching-sequencing-determinism.md)).

---

## 6. Complete User Journey Scenarios: Fiat & Crypto Withdrawals

### 6.1 Instant Fiat Payouts (Domestic INR)
Fiat payouts are executed 24/7/365 exclusively to the investor's verified domestic bank account.

```
+---------------------------------------------------------------------------------------------------+
|                                     FIAT WITHDRAWAL LIFECYCLE                                     |
|                                                                                                   |
|  [User Payout Request]                                                                            |
|          |                                                                                        |
|          v                                                                                        |
|  [Pessimistic Balance Reservation] (Available -> Reserved in PostgreSQL Core Ledger)              |
|          |                                                                                        |
|          v                                                                                        |
|  [Risk Engine Evaluation] (Velocity limits, 12h cooling post-KYC/password change, travel rule)    |
|          |                                                                                        |
|    +-----+--------------------+                                                                   |
|    | Amount <= ₹5,00,000      | Amount > ₹5,00,000                                                |
|    v                          v                                                                   |
|  [Instant IMPS / UPI]       [Corporate RTGS Nodal Rail]                                           |
|  SLA: < 15 seconds           SLA: < 15 minutes (Bank Operating Hours)                             |
|    |                          |                                                                   |
|    +-----+--------------------+                                                                   |
|          v                                                                                        |
|  [Bank Webhook / UTR Confirmation Received]                                                       |
|          |                                                                                        |
|          v                                                                                        |
|  [Double-Entry Settlement Commit] (Reserved Debit -> Outflow Clear; see ADR-0012)                 |
+---------------------------------------------------------------------------------------------------+
```

1. **Pessimistic Ledger Reservation:** To prevent double-spend attacks from rapid concurrent withdrawal requests, the user balance is immediately moved from `available` to `reserved_pending_payout` inside a serializable PostgreSQL transaction with row-level locking (`SELECT ... FOR UPDATE`, see [ADR-0012](../adr/ADR-0012-pessimistic-locking-concurrent-withdrawals.md)).
2. **Instant Rail Dispatch:** Amounts up to ₹5,00,000 are dispatched via IMPS / UPI rails delivering an end-to-end user receipt SLA $< 15$ seconds.
3. **High-Value RTGS Routing:** Amounts exceeding ₹5,00,000 are batched and routed via commercial bank corporate RTGS APIs.

### 6.2 Crypto On-Chain Withdrawals & Multi-Layer Security

```
+---------------------------------------------------------------------------------------------------+
|                                  CRYPTO ON-CHAIN WITHDRAWAL FLOW                                  |
|                                                                                                   |
|  [Destination Address Input] -> [Whitelisted in Address Book?]                                    |
|                                         |                                                         |
|               +-------------------------+-------------------------+                               |
|               | YES                                               | NO                            |
|               v                                                   v                               |
|  [12-Hour Cooldown Elapsed?]                        [Add to Address Book Workflow]                |
|               |                                     - FIDO2 WebAuthn / Passkey                    |
|       +-------+-------+                             - TOTP + SMS OTP Verification                 |
|       | YES           | NO                          - Enforce 12-Hour Security Cooldown           |
|       v               v                                                                           |
|  [AML Risk    [Request Blocked]                                                                   |
|   Screening]  Address in 12h cooldown                                                             |
|       |                                                                                           |
|       v                                                                                           |
|  [MPC 2-of-3 Threshold Signing] (AWS Nitro Secure Enclave + Cosigning Engine)                     |
|       |                                                                                           |
|       v                                                                                           |
|  [Broadcast to Blockchain] -> [Live TxHash Tracker] -> [Confirmation Finality]                    |
+---------------------------------------------------------------------------------------------------+
```

#### 6.2.1 Address Book Management & 12-Hour Security Cooldown
To protect users against account takeover and session hijacking attacks:
1. **Mandatory Whitelisting:** Crypto withdrawals can only be dispatched to addresses explicitly saved in the user's cryptographically confirmed Address Book.
2. **12-Hour Activation Cooldown Invariant:**
   - When a new address is added to the Address Book, it enters a non-bypassable `PENDING_ACTIVATION` state.
   - For exactly 12 hours (43,200 seconds), zero funds can be withdrawn to that address.
   - Multi-channel security notifications (SMS, Email, Push) are dispatched immediately. If the user did not initiate the addition, they can freeze their account with a single click.

#### 6.2.2 Multi-Factor Step-Up Authentication
Dispatching a crypto withdrawal requires explicit three-tier step-up authorization:
1. **FIDO2 WebAuthn / Hardware Key:** Biometric Passkey or hardware security key (YubiKey) verification via WebAuthn protocol.
2. **Time-Based One-Time Password (TOTP):** 6-digit Google Authenticator / RFC 6238 token.
3. **Cryptographic Email Confirmation Link:** High-value withdrawals ($> \$5,000$ USD equivalent) require confirming an HMAC-SHA256 signed approval link dispatched to the user's primary registered email address.

#### 6.2.3 MPC 2-of-3 Threshold Signing Architecture
Conforming to [ADR-0031](../adr/ADR-0031-mpc-threshold-signing-and-cold-vault-custody.md), private keys are never assembled in memory:
- **Key Share 1:** Isolated inside AWS Nitro Enclaves running confidential containerized signers.
- **Key Share 2:** Governed by the automated Co-Signing Policy Engine (verifies user 2FA, daily withdrawal velocity limits, and FATF Travel Rule compliance).
- **Key Share 3:** Cold Disaster Recovery Key Share secured in institutional offline custody.
- Transactions are signed cooperatively using Lindell / GG20 threshold ECDSA and FROST EdDSA protocols.

### 6.3 Zero-Fee Internal Transfers (Growww ID Transfers)
Traders transferring digital balances to other registered Growww users (identified by unique Growww ID, verified mobile number, or email) bypass public blockchain networks completely:
1. **Off-Chain Atomic Book-Entry:** The funds are transferred entirely on the internal double-entry ledger database:
   $$\text{UserA}_{\text{Available}} \leftarrow \text{UserA}_{\text{Available}} - \text{Amount}$$
   $$\text{UserB}_{\text{Available}} \leftarrow \text{UserB}_{\text{Available}} + \text{Amount}$$
2. **Latency & Cost:** Executed with $< 50\text{ms}$ latency at ₹0.00 network gas fees.
3. **Compliance Auditability:** The transaction is assigned an internal book-transfer reference logged in the immutable audit ledger with complete AML origin-to-destination traceability.

---

## 7. Advanced Retail & Institutional Capabilities

### 7.1 Sub-Accounts & Institutional Hierarchy

```
+-------------------------------------------------------------------------------------------------------------+
|                                    INSTITUTIONAL ACCOUNT HIERARCHY                                          |
|                                                                                                             |
|                                        +----------------------------+                                       |
|                                        |  Master Institutional Org  |                                       |
|                                        |  (KYC Verified Legal Corp) |                                       |
|                                        +--------------+-------------+                                       |
|                                                       |                                                     |
|                     +---------------------------------+---------------------------------+                   |
|                     |                                 |                                 |                   |
|                     v                                 v                                 v                   |
|       +---------------------------+     +---------------------------+     +---------------------------+     |
|       | Sub-Account #1: Prop Desk |     | Sub-Account #2: MM Market |     | Sub-Account #3: Arbitrage |     |
|       | - Scopes: Spot, Margin    |     | - Scopes: Spot Trading    |     | - Scopes: Spot Trading    |     |
|       | - NO WITHDRAWAL           |     | - NO WITHDRAWAL           |     | - NO WITHDRAWAL           |     |
|       | - API Rate: 2,000 req/s   |     | - API Rate: 5,000 req/s   |     | - API Rate: 2,000 req/s   |     |
|       | - Collateral: Isolated    |     | - Collateral: Cross-Margin|     | - Collateral: Isolated    |     |
|       +---------------------------+     +---------------------------+     +---------------------------+     |
+-------------------------------------------------------------------------------------------------------------+
```

1. **Master-to-Subaccount Topology:** Corporate institutional accounts can spawn up to 200 child sub-accounts for algorithmic strategy isolation, risk compartmentation, and trader attribution.
2. **Granular Permission Scopes:**
   - `READ_ONLY`: Market data inspection, order status queries, and historical reporting.
   - `SPOT_TRADING`: Order placement, modification, and cancellation across spot books.
   - `MARGIN_TRADING`: Collateral borrowing, leverage modification, and margin positions.
   - `INTERNAL_TRANSFER`: Inter-subaccount collateral rebalancing.
   - `NO_WITHDRAWAL`: Strict architectural guarantee that API keys bound to sub-accounts cannot initiate external fiat or blockchain withdrawals under any circumstance.
3. **Instant Zero-Fee Collateral Sweeping:** Master accounts can sweep and reallocate margin collateral between sub-accounts instantly without touching settlement layers.

### 7.2 Developer & Institutional API Key Infrastructure
High-frequency market makers and algorithmic trading desks interact through low-latency REST and WebSocket APIs.

#### 7.2.1 Cryptographic Authentication: Ed25519 & HMAC-SHA256
API requests support dual cryptographic authentication paradigms:
1. **Ed25519 Asymmetric Signatures (Recommended for Ultra-Low Latency):**
   - The institutional client signs the canonical request string using their local Ed25519 private key:
     $$\text{CanonicalString} = \text{Timestamp} + \text{HTTP\_Method} + \text{Request\_Path} + \text{Body}$$
     $$\text{Signature} = \text{Ed25519\_Sign}(\text{PrivateKey}, \text{CanonicalString})$$
   - The gateway validates the signature against the pre-registered Ed25519 public key in $< 50\mu\text{s}$.
2. **HMAC-SHA256 Symmetric Signatures:**
   - Standard symmetric signature using pre-shared secret key:
     $$\text{Signature} = \text{HMAC-SHA256}(\text{API\_Secret}, \text{CanonicalString})$$
3. **Replay Attack Protection:** All requests require a monotonic millisecond timestamp header (`X-GROWWW-TIMESTAMP`). Requests with timestamps deviating by more than $\pm 5,000\text{ms}$ from atomic exchange NTP servers are rejected.

#### 7.2.2 IP Whitelisting & CIDR Subnet Restriction
API keys can be restricted to up to 20 static IP addresses or CIDR subnets. Requests originating from non-whitelisted IP addresses are dropped at the perimeter edge firewall.

#### 7.2.3 Tiered Rate Limits & Leaky Bucket Governance
Rate limits are enforced at the API Gateway via an in-memory Redis Leaky Bucket algorithm:

| Tier | Eligibility / 30-Day Volume | REST Rate Limit | WebSocket Execution Rate | FIX 4.4 Throughput |
| :--- | :--- | :--- | :--- | :--- |
| **Tier 1 (Retail)** | Default ($< ₹10\text{ Lakhs}$) | 100 req / sec | 50 msgs / sec | Not Eligible |
| **Tier 2 (Pro Trader)** | Volume $\ge ₹10\text{ Lakhs}$ | 500 req / sec | 250 msgs / sec | Not Eligible |
| **Tier 3 (Institutional)** | Volume $\ge ₹10\text{ Crores}$ | 2,000 req / sec | 1,000 msgs / sec | 1,000 msgs / sec |
| **Tier 4 (Market Maker VIP)** | Tiered MM Agreement | 5,000 req / sec | 2,500 msgs / sec | 5,000 msgs / sec |

### 7.3 Real-Time Price Alerts & Smart Notifications
The notification hub operates as an event-driven cluster consuming from Kafka and dispatching across four parallel channels:

```
[Price Ticks / Order Events / Margin Telemetry] 
                     |
                     v (Kafka: `marketdata.ticks.v1`, `order.events.v1`)
          [Notification Rule Engine]
                     |
    +----------------+----------------+----------------+
    |                |                |                |
    v                v                v                v
[WebPush / WSS]  [Mobile APNS/FCM]   [DLT SMS Rail]   [Transactional Email]
In-App Banner    Push Notification   Statutory Alerts Order Fill Statements
Latency: < 50ms  Latency: < 500ms    Latency: < 3s    Latency: < 5s
```

1. **Price Threshold & Volatility Spike Detectors:**
   - Users can configure static alerts ($P \ge X$ or $P \le Y$).
   - Dynamic 5-minute volatility alerts: Triggers when the price moves $> 3.5\%$ within any rolling 300-second window, alerting users to sudden market dislocations.
2. **Execution & Margin Telemetry Alerts:**
   - **Order Fills:** Instantaneous alerts for full fills, partial fills, and stop triggers.
   - **Margin Calls:** Dispatched when portfolio Margin Level drops below $120\%$.
   - **Pre-Liquidation Warning:** Emergency high-priority push, SMS, and email dispatched when Margin Level drops below $105\%$.

### 7.4 Real-Time Portfolio Analytics & Statutory Tax Reporting

#### 7.4.1 Real-Time Mark-to-Market (MTM) P&L Engine
Portfolio holdings are evaluated in real time against the exchange Last Traded Price (LTP):
1. **Unrealized P&L Formula:**
   $$\text{Unrealized P\&L} = \sum_{i=1}^{N} Q_i \times (P_{\text{LTP}, i} - P_{\text{AvgCost}, i})$$
2. **Realized P&L Formula (Tax-Lot Accounting):**
   $$\text{Realized P\&L} = \sum_{j=1}^{M} Q_{\text{sold}, j} \times (P_{\text{sale}, j} - P_{\text{cost\_lot}, j}) - \text{Fees}_j$$

#### 7.4.2 Tax-Lot Accounting Paradigms
1. **First-In-First-Out (FIFO):** Default statutory accounting method under Indian Income Tax regulations. Oldest acquisition cost lots are matched and depleted first upon sale.
2. **Weighted Average Cost (WAC):** Available for institutional portfolio performance analysis.

#### 7.4.3 Decoupled Voluntary Portfolio & Tax Statement Exports
The platform automates tax compliance under the Indian Income Tax Act 1961:
1. **Decoupled Tax Reporting Engine (Zero On-Chain Withholding):**
   - 1% Tax Deducted at Source (TDS) is automatically calculated and withheld on the gross consideration of all Virtual Digital Asset (VDA) sell transactions exceeding statutory annual thresholds (₹50,000 for specified persons, ₹10,000 for others).
   - The withheld TDS is mapped to the user's verified PAN and deposited with the Income Tax Department.
2. **Downloadable Tax Reports:**
   - **Voluntary PnL Statement:** Exportable transaction history reconciling all executed trades, cost basis, and realized PnL.
   - **Voluntary Tax Lot Summary:** End-of-year tax summary detailing cost of acquisition, sale proceeds, and FIFO capital gains calculated asynchronously off-chain for independent filing.

### 7.5 Proof-of-Stake (PoS) Staking & Yield Infrastructure
Allows users to participate in consensus validation rewards without relinquishing underlying custody ownership:
1. **Non-Custodial Liquid Staking Architecture:**
   - **Ethereum (ETH):** Users stake ETH and receive liquid staking tokens (stETH / rETH). The exchange interfaces with distributed validator nodes (SSV / Obol DVT network).
   - **Solana (SOL):** Staking delegated to high-performance, non-censoring institutional validators with zero commission markups.
2. **Daily Auto-Compounding Yield:** Staking rewards are accrued and compounded daily at 00:00 UTC into the user's staking balance.
3. **Unbonding Timers & Instant Liquidity:**
   - Standard Unbonding: Bounded by native network protocol unbonding delays (e.g. 9-14 days for ETH queue, ~3 days for Solana epoch).
   - Instant Exit Pool: Users can exit instantly with zero waiting time by swapping their staking tokens against the exchange internal liquidity buffer for a 0.15% pool rebalancing fee.

---

## 8. Edge Case Handling, Failure Recovery & Operational Resilience

### 8.1 Missing or Invalid Memo / Tag Recovery Protocol
Certain high-throughput blockchains (Ripple XRP, TON, Stellar XLM) use a single pooled platform deposit address and rely on a 32-bit integer `Destination Tag` or `Memo` to identify the recipient account.

```
+---------------------------------------------------------------------------------------------------+
|                            MISSING / INVALID MEMO RECOVERY WATERFALL                              |
|                                                                                                   |
|  [External Deposit Arrives at Master Address with Missing or Unrecognized Tag]                   |
|                                         |                                                         |
|                                         v                                                         |
|  [Orphan Holding Vault]: Funds credited to platform isolation ledger `orphan_deposits`            |
|                                         |                                                         |
|                                         v                                                         |
|  [User Self-Service Recovery Portal]:                                                             |
|  1. User inputs: Chain (XRP/TON), TxHash, Expected Amount, Originating Wallet Address             |
|  2. Cryptographic Proof of Ownership Challenge:                                                   |
|     - Case A (Self-Custody): User signs challenge nonce with originating private key              |
|     - Case B (External VASP): User sends Satoshi micro-deposit ($0.01) from exact source account  |
|                                         |                                                         |
|                                         v                                                         |
|  [Automated Reconciliation Engine]: Matches TxHash, source address, and amount                     |
|                                         |                                                         |
|                                         v                                                         |
|  [Deduction of 0.10% Recovery Fee] (Minimum ₹100 / $1.50 for manual AML re-screening)              |
|                                         |                                                         |
|                                         v                                                         |
|  [Credit Target User Balance]: Transfer from `orphan_deposits` to User Ledger with audit link     |
+---------------------------------------------------------------------------------------------------+
```

1. **Orphan Vault Isolation:** Transactions arriving with missing, zero, or unassigned memos are intercepted by the node ingestion pipeline and credited to a quarantined ledger (`ledger_account_orphan_deposits`).
2. **Self-Service Verification Challenge:**
   - The user opens the Deposit Recovery Console and submits the transaction hash (`tx_hash`).
   - The user must prove ownership of the sending wallet: for private wallets, they submit an Ed25519/Secp256k1 signature of a platform-generated nonce; for exchange wallets, they initiate a micro-deposit matching an exact penny fractional sum.
3. **Resolution:** Upon cryptographic match, the transaction is processed and credited to the user's account with a 0.10% recovery fee deducted to cover operational re-verification.

### 8.2 External Blockchain Network Reorganization (Reorg) Handling
Deep reorganizations on external Proof-of-Work or Proof-of-Stake blockchains present insolvency risks if unconfirmed deposits are credited prematurely.

```
Canonical Chain:  [Block 100] -> [Block 101] -> [Block 102] -> [Block 103] -> [Block 104]
                                       \
Reorg Orphan:                           +-----> [Block 102'] (User Deposit Tx inside orphan!)
```

#### 8.2.1 Reorg Detection & State Invalidation Protocol
Conforming to [ADR-0010](../adr/ADR-0010-tiered-crosschain-finality-governance.md):
1. **Node Watcher Consensus:** Node ingestion relayers track the canonical block hash sequence across three geographically independent full nodes.
2. **Reorganization Trigger:** If a previously observed block hash is displaced by an alternate heavier chain branch at depth $D \ge 1$:
   - The relayer publishes an immediate emergency event: `blockchain.reorg.detected` (`chain: "ETH"`, `depth: D`, `fork_block: 101`).
3. **Deposit Rollback Mechanics:**
   - If the deposit transaction does not exist on the new canonical branch, the deposit state is updated from `CONFIRMED` to `REORG_INVALIDATED`.
   - If the user has not yet spent the balance, the deposited funds are debited from their available balance.

#### 8.2.2 Negative Balance Containment & Compensation Saga
If the user already executed trades or converted the balance prior to the reorg:
1. **Trading Balance Lock:** The account is placed in a `MARGIN_DEFICIT_CONTAINMENT` state. Available withdrawals are suspended.
2. **Mempool Re-broadcast:** The platform custody relayer attempts to re-broadcast the orphaned transaction directly into the mempool of the new canonical chain branch.
3. **Debt Recovery Waterfall:** If the transaction cannot be confirmed after 24 hours:
   - Any open resting orders are canceled.
   - The user's other positive asset balances are frozen up to the deficit amount.
   - If uncollateralized, the loss is provisioned against the Platform Default Insurance Fund (see [ADR-0013](../adr/ADR-0013-sebi-core-sgf-default-waterfall.md) and [ADR-0029](../adr/ADR-0029-forced-liquidation-penalty-fee-and-insurance-fund.md)), and the user's account remains frozen until manually settled.

### 8.3 Session Management & In-Band Token Refresh Protocol
Active traders maintaining persistent WebSocket subscriptions (streaming Level-2 depth and live order executions) must not experience disconnections or visual blind spots when short-lived authentication credentials expire.

Conforming to [ADR-0028](../adr/ADR-0028-inband-websocket-token-refresh.md):

```
Client App                                 API Gateway / WSS Hub                User Service
    |                                                |                                |
    | ====== (Continuous Level-2 Market Data) ====== |                                |
    |                                                |                                |
    |-- [T-60s to Token Expiry]: Request Refresh --->|                                |
    |   (via out-of-band HTTPS REST) ------------------------------------------------>|
    |                                                |   Validate & Sign New JWT      |
    |<-- Return New JWT Token (15-min validity) --------------------------------------|
    |                                                |                                |
    |-- In-Band WS Refresh Message ----------------->|                                |
    |   `{"action": "auth_refresh", "token": "..."}` |                                |
    |                                                |-- Validate Signature & Claims  |
    |                                                |-- Update Connection Expiry     |
    |<-- In-Band Acknowledgment ---------------------|                                |
    |   `{"status": "AUTH_SUCCESS", "valid_until":..}`                                |
    |                                                |                                |
    | ====== (Zero-Interruption Level-2 Streaming Continues Unaltered) ============= |
```

1. **Proactive In-Band Refresh:** Exactly 60 seconds prior to JWT expiration, the client retrieves a refreshed JWT via REST and transmits it directly inside the existing WebSocket channel.
2. **Zero-Disconnect Handshake:** The WebSocket Gateway validates the new signature, extends the connection context deadline by 15 minutes, and responds with `AUTH_SUCCESS`.
3. **Grace Period & Graceful Termination:** If no refreshed token is submitted within 30 seconds past expiry, the gateway closes the socket with standard WebSocket closure code 4401 (`TOKEN_EXPIRED`).

### 8.4 Emergency Account Freeze & Self-Lock Kill Switch
If an investor observes suspicious activity (e.g. an unauthorized login alert or unexpected 2FA push), they can immediately trigger the platform Emergency Kill Switch.

```
[User Triggers Emergency Kill Switch]
                  |
                  v
+---------------------------------------------------------------------------------------------------+
|                                 ATOMIC BLAST RADIUS CONTAINMENT                                   |
|                                                                                                   |
|  1. Session Termination:   All active web, mobile, and desktop JWT sessions revoked in Redis      |
|  2. API Key Invalidation:  All active API keys instantly suspended (`status = SUSPENDED`)        |
|  3. Order Cancellation:    Asynchronous broadcast to Matching Engine canceling all open orders    |
|  4. Withdrawal Lockdown:   All pending and future fiat/crypto withdrawals frozen                  |
|  5. Sub-Account Cascade:   All associated sub-accounts placed in lockdown                         |
+---------------------------------------------------------------------------------------------------+
                  |
                  v
[Account Placed in `EMERGENCY_FROZEN` State]
                  |
                  v
[Unlocking Workflow]:
- Mandatory 24-Hour Cooling Period
- Live Video KYC Re-Verification with Compliance Officer
- Biometric Liveness Verification + Password / 2FA Cryptographic Reset
```

1. **One-Click Trigger:** Accessible via in-app security settings, emergency SMS links, and transactional email notification footers.
2. **Blast Radius & Revocations:**
   - **Session Purge:** Invalidate all active Redis refresh tokens (`auth:session:{user_id}:*`).
   - **API Key Disablement:** Mark all active institutional API keys as `SUSPENDED`.
   - **Mass Order Cancellation:** Dispatch a prioritized `OrderMassCancel` command to the matching engine core for that user ID.
   - **Withdrawal Lock:** Freeze all outgoing transactions in the settlement pipeline.
3. **Restoration Protocol:** Unlocking requires completing a 24-hour mandatory security cooling period followed by a live Video KYC verification session with an exchange compliance officer and a complete reset of all authentication credentials.

---

## 9. Statutory & Regulatory Invariants

### 9.1 Regulatory Oversight Matrix
The platform operates under a dual-entity institutional architecture:

| Regulatory Authority | Statutory Jurisdiction | Governed Operations | Core Requirements |
| :--- | :--- | :--- | :--- |
| **SEBI** (Securities & Exchange Board of India) | Domestic Equities & Securities Tokens | Brokerage, Clearing, Demat Custody | 100% Asset backing, Pre-Trade DMA, Zero-fee execution |
| **RBI** (Reserve Bank of India) | Domestic Banking & Payments | UPI, IMPS, RTGS, Nodal Accounts | Zero 3rd-party deposits, Penny-drop verification, PMLA 2002 |
| **IFSCA** (GIFT City International Authority) | Cross-Border & Foreign Capital | Multi-Currency Trading, International Desks | FATF Travel Rule, LRS compliance, USD clearing |
| **FIU-IND** (Financial Intelligence Unit India) | Anti-Money Laundering (AML) | All Domestic & International Rails | Suspicious Transaction Reports (STR), Cash Transaction Reports (CTR) |
| **ITD** (Income Tax Department) | Taxation & Revenue | Voluntary Tax Statements | Clean DvP on-chain settlement, voluntary tax exports |

### 9.2 Decoupled Tax Architecture & Zero-Withholding Invariants
1. **Zero Deduction at Source:** Strictly 0.00% TDS is deducted on gross consideration across all spot trades and convert orders. The platform operates on a frictionless DvP model with zero on-chain withholding.
2. **Rounding Rule:** Statutory rounding to the nearest integer rupee conforms to [ADR-0025](../adr/ADR-0025-statutory-stt-gross-aggregated-contract-note-rounding.md).
3. **Voluntary Tax Reporting:** Off-chain analytics compute voluntary FIFO tax lot summaries for traders who file independent returns, without any on-chain deduction.

### 9.3 DPDP Act 2023 Cryptographic Shredding & Privacy Isolation
1. **Zero On-Chain PII Invariant:** In compliance with the Digital Personal Data Protection Act 2023, zero Personally Identifiable Information (Aadhaar, PAN, names, bank account numbers, IP addresses) is stored on the Hyperledger Besu consortium blockchain.
2. **Crypto-Shredding Architecture:** User PII in off-chain databases is encrypted using unique per-user AES-256-GCM data encryption keys (DEKs). Upon verified Right-to-be-Forgotten erasure requests (post statutory 10-year PMLA retention expiration), the user DEK is destroyed, rendering historical archives permanently unrecoverable (see [ADR-0007](../adr/ADR-0007-crypto-shredding-dpdp-compliance.md)).

---

## 10. Traceability Matrix & Architectural References

### 10.1 Architecture Decision Records (ADRs)
- [ADR-0001: Hyperledger Besu QBFT Consensus Architecture](../adr/ADR-0001-hyperledger-besu-qbft.md)
- [ADR-0003: Rust In-Memory Matching Engine Architecture](../adr/ADR-0003-rust-in-memory-matching-engine.md)
- [ADR-0004: Kafka Event Streaming & Transactional Outbox](../adr/ADR-0004-kafka-event-streaming-and-outbox.md)
- [ADR-0005: ERC-3643 Permissioned Security Token Standard](../adr/ADR-0005-erc3643-permissioned-security-tokens.md)
- [ADR-0006: Immutable Double-Entry Journal Ledger](../adr/ADR-0006-immutable-double-entry-journal-ledger.md)
- [ADR-0007: Cryptographic Shredding for DPDP Compliance](../adr/ADR-0007-crypto-shredding-dpdp-compliance.md)
- [ADR-0008: Securities Tax Classification and Statutory TDS](../adr/ADR-0008-securities-tax-classification-and-tds.md)
- [ADR-0010: Tiered Cross-Chain Finality Governance](../adr/ADR-0010-tiered-crosschain-finality-governance.md)
- [ADR-0011: Fixed-Point Integer Arithmetic Precision](../adr/ADR-0011-fixed-point-integer-arithmetic-precision.md)
- [ADR-0012: Pessimistic Locking for Concurrent Withdrawals](../adr/ADR-0012-pessimistic-locking-concurrent-withdrawals.md)
- [ADR-0013: SEBI Core Settlement Guarantee Fund Waterfall](../adr/ADR-0013-sebi-core-sgf-default-waterfall.md)
- [ADR-0024: WebSocket Slow Consumer Conflation & Backpressure](../adr/ADR-0024-websocket-slow-consumer-conflation-backpressure.md)
- [ADR-0025: Statutory STT & TDS Gross Aggregated Rounding](../adr/ADR-0025-statutory-stt-gross-aggregated-contract-note-rounding.md)
- [ADR-0027: Order Cancellation Matching Sequencing Determinism](../adr/ADR-0027-order-cancellation-matching-sequencing-determinism.md)
- [ADR-0028: In-Band WebSocket Token Refresh](../adr/ADR-0028-inband-websocket-token-refresh.md)
- [ADR-0029: Forced Liquidation Penalty Fee & Insurance Fund](../adr/ADR-0029-forced-liquidation-penalty-fee-and-insurance-fund.md)
- [ADR-0030: Multi-Chain HD Wallet Deposit Derivation](../adr/ADR-0030-multi-chain-hd-wallet-deposit-derivation.md)
- [ADR-0031: MPC Threshold Signing & 3-Tier Custody Architecture](../adr/ADR-0031-mpc-threshold-signing-and-cold-vault-custody.md)
- [ADR-0032: FATF Travel Rule Protocol & Blockchain AML Scoring](../adr/ADR-0032-travel-rule-and-blockchain-aml-risk-scoring.md)
- [ADR-0033: RFQ Instant Convert & Zero-Slippage Swap Engine](../adr/ADR-0033-rfq-instant-convert-swap-engine.md)

### 10.2 Core Architecture Specifications
- [Frontend & Client Application Architecture](FRONTEND_AND_CLIENT_ARCHITECTURE.md)
- [Core Matching Engine & Order Book Specification](CORE_MATCHING_ENGINE_AND_ORDER_BOOK_SPECIFICATION.md)
- [Core Ledger & Multi-Currency Specification](CORE_LEDGER_AND_MULTI_CURRENCY_SPECIFICATION.md)
- [Distributed Systems & Resilience Specification](DISTRIBUTED_SYSTEMS_AND_RESILIENCE_SPECIFICATION.md)
- [DvP Settlement & Clearing Waterfall Specification](DVP_SETTLEMENT_AND_CLEARING_WATERFALL_SPECIFICATION.md)
- [Regulatory Risk & Cross-Chain Governance](REGULATORY_RISK_AND_CROSSCHAIN_GOVERNANCE.md)
- [NBSE Master Architecture Blueprint](NBSE_MASTER_ARCHITECTURE_AND_INTEGRATION_BLUEPRINT.md)
- [Candlestick Aggregation & Market Data Specification](CANDLESTICK_AGGREGATION_AND_MARKET_DATA_SPECIFICATION.md)
- [Event-Driven Architecture Specification](EVENT_DRIVEN_ARCHITECTURE_SPECIFICATION.md)
- [Financial & Domain Specifications](FINANCIAL_AND_DOMAIN_SPECIFICATIONS.md)
- [Blockchain Ledger Core Specification](BLOCKCHAIN_LEDGER_CORE_SPECIFICATION.md)

### 10.3 Master Product & Provenance Documentation
- [Product Definition](../PRODUCT_DEFINITION.md)
- [Exchange Features Master Index](../FEATURES.md)
- [Implementation Status Tracker](../IMPLEMENTATION_STATUS.md)
- [Engineering Standards and Provenance](../ENGINEERING_STANDARDS_AND_PROVENANCE.md)
