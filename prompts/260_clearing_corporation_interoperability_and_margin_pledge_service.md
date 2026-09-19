# 260 - Clearing Corporation Interoperability & SEBI Margin Pledge/Re-Pledge Gateway (Go)

## Purpose
In modern Indian capital markets, margin trading across Cash Equities, Equity Derivatives (Futures & Options), Currency Derivatives, and Commodities requires strict adherence to central clearing guidelines and statutory depository mechanics. Historically, brokers collected client securities into proprietary pool accounts, exposing client assets to broker default and co-mingling risks. Under the Securities and Exchange Board of India (SEBI) Margin Pledge / Re-Pledge framework (SEBI circular SEBI/HO/MIRSD/DOP/CIR/P/2020/143) and subsequent Client Collateral Segregation & Allocation directives (SEBI/HO/MRD2_DCAP/P/CIR/2021/0598), securities can no longer leave the investor's own Demat account. Instead, collateral is pledged directly in favor of the Trading Member (TM) / Clearing Member (CM) and subsequently re-pledged to the designated Clearing Corporation (CC).

Furthermore, the implementation of Clearing Corporation Interoperability enables market participants to consolidate clearing and settlement of trades executed across multiple exchanges (National Stock Exchange - NSE, and Bombay Stock Exchange - BSE) through a single preferred Clearing Corporation: either NSE Clearing Limited (NSCCL) or Indian Clearing Corporation Limited (ICCL). This requires an intelligent, automated routing gateway capable of handling multi-CCP collateral allocation, real-time haircut calculation, demat authorization handshakes, and seamless collateral transfer between CCPs without double-pledging or capital fragmentation.

The **Clearing Corporation Interoperability & SEBI Margin Pledge/Re-Pledge Gateway** (`services/clearing-pledge-service`) provides an enterprise-grade, high-throughput gateway built in Go 1.22+. It automates depository pledge handshakes with National Securities Depository Limited (NSDL) and Central Depository Services (India) Limited (CDSL), orchestrates CDSL TPIN and NSDL mobile OTP authentication workflows, conducts clearing member collateral valuation against daily Value-at-Risk (VaR) and Extreme Loss Margin (ELM) haircut matrices, reconciles client collateral segregation reports, and interfaces with the National Blockchain Stock Exchange (NBSE) permissioned Hyperledger Besu ledger for on-chain ERC-3643 collateral locking, while assessing the mandatory platform invariant flat 0.00% transaction fee (No fee at all) (0.00% fee at launch; future fee parameters governed by FeeController.sol).

## What You Are Building
A mission-critical, resilient clearing and depository integration microservice (`services/clearing-pledge-service`) written in Go 1.22+, comprising:
- **Depository Pledge Gateway (NSDL & CDSL Adapters):** Interfaces directly with NSDL and CDSL Clearing Corporation Push Service (CCPS) and Electronic Delivery Instruction Slip (e-DIS) APIs using ISO 20022 financial messaging formats (`colr.003`, `colr.004`, `semt.013`, `sese.023`) over mutual TLS (mTLS) with PKCS#7 SHA-256 digital signature signing.
- **Demat TPIN & OTP Authorization Orchestrator:** Manages client-side depository authorization redirects, tracking time-to-live (TTL) ephemeral session states, validating CDSL TPIN submissions and NSDL mobile OTP tokens, and driving automated pledge status transitions.
- **Clearing Corporation Interoperability & Allocation Router:** Optimizes collateral placement across NSCCL and ICCL. Evaluates member-level margin commitments, portfolio open interest, and cross-exchange margin netting benefits to dynamically allocate and re-route re-pledged collateral.
- **Real-Time Collateral Haircut & Valuation Engine:** Ingests daily Clearing Corporation haircut files (`C_VAR1`, `C_VAR2`, illiquidity multipliers, and approved security masters). Computes real-time effective margin credit based on live tick feeds and statutory 50:50 cash-to-non-cash collateral ratios.
- **Pre-Release Risk & Margin Verification Pipeline:** Guarantees that collateral release or unpledge instructions are only dispatched to depositories if the client maintains a minimum 105% margin solvency buffer post-release, synchronously verifying state with the Risk & Margin Service (Prompt 206).
- **Daily SEBI Collateral Segregation & Reporting Engine:** Generates end-of-day (EOD) and intraday snapshot files adhering to SEBI client collateral allocation specifications, proving that zero client collateral is co-mingled or utilized for proprietary trading member obligations.
- **Hyperledger Besu ERC-3643 Collateral Lock Relayer:** For tokenized securities on the permissioned ledger, executes smart contract collateral lockups (`lockCollateral`), records depository pledge references, and sets compliance claim attributes (`COLLATERAL_PLEDGED_NSCCL` / `COLLATERAL_PLEDGED_ICCL`) without exposing PII.
- **Universal 0.00% (No fee at all) Platform Fee Ledger:** Automatically logs and distributes the platform's mandatory 0.00% fee (No fee at all) on collateral turnover (Platform Treasury, Core SGF, and Investor Protection Fund per FeeController governance).

## Scope Boundaries
- **In Scope:**
  - Direct integration with NSDL and CDSL pledge, re-pledge, unpledge, and invocation protocols via ISO 20022 and REST/SFTP CCPS endpoints.
  - Depository client authorization workflows (CDSL TPIN verification and NSDL mobile OTP authorization).
  - Multi-CCP interoperability routing between NSCCL and ICCL for collateral fungibility and optimal margin credit.
  - Real-time ingestion of CC haircut matrices (VaR + ELM + liquidity adjustments) and collateral valuation.
  - Client collateral segregation accounting and automated EOD SEBI reporting generation.
  - Inter-service gRPC communication with Risk & Margin Service (Prompt 206) and Wallet Ledger (Prompt 203).
  - Hyperledger Besu on-chain collateral freeze and ERC-3643 compliance claim updates.
  - Universal fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) assessment and ledger accounting.
- **Out of Scope / Handled Elsewhere:**
  - Order execution and trade matching (handled in Prompt 205 Order Matching Engine).
  - SPAN margin mathematical scanning and delta-gamma risk calculation (handled in Prompt 241 SPAN Margin Engine).
  - Central Counterparty Default Waterfall liquidation and auctioning (handled in Prompt 230 SGF & Default Waterfall Service).
  - Direct bank fiat payment gateway and UPI auto-debit collection (handled in Prompt 203 / Prompt 214).
  - Cross-border foreign portfolio investor cap enforcement (handled in Prompt 250 FPI Headroom Engine).

## Technology to Use
- **Core Language & Runtime:** **Go 1.22+** utilizing goroutines, structured context propagation, bounded worker pools, and low-latency garbage collection optimized for microsecond IPC.
- **Financial Messaging Protocol:** **ISO 20022** XML/JSON schemas for securities collateral management:
  - `colr.003.001.04` (CollateralProposal)
  - `colr.004.001.04` (CollateralProposalResponse)
  - `semt.013.001.04` (IntraPositionMovementReport)
  - `sese.023.001.09` (SecuritiesSettlementTransactionInstruction)
- **Depository Communication & Security:**
  - Mutual TLS (mTLS) with RSA 4096-bit client certificates.
  - Cryptographic digital signing via PKCS#7 / CMS (Cryptographic Message Syntax) with SHA-256 using Hardware Security Module (HSM) FIPS 140-2 Level 3 integration.
  - NSDL CCPS and CDSL CDAS REST/SOAP API adapters.
- **Database & Persistence:** **PostgreSQL 16+** with `jackc/pgx/v5` connection pooling, strict ACID transactions, table partitioning by clearing date, and advisory locks for preventing concurrent pledge mutation.
- **In-Memory Cache & Ephemeral Sessions:** **Redis 7.2+ Cluster** with Redis Sentinel failover for caching live collateral locks, haircut tables, and 180-second TPIN/OTP session tokens.
- **Message Broker & Event Streaming:** **Apache Kafka 3.7+** with `confluent-kafka-go` (librdkafka) for streaming collateral state transitions and audit events with guaranteed `at-least-once` delivery.
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** with Unix Domain Sockets and mTLS for secure, sub-millisecond internal communications.
- **Blockchain Client:** `github.com/ethereum/go-ethereum/ethclient` interfacing with Hyperledger Besu QBFT permissioned nodes.

## Backend / Infra Touchpoints
- **PostgreSQL Database Tables:**
  - `collateral_pledge_requests`: Master ledger of client pledge, re-pledge, and unpledge requests.
  - `demat_collateral_holdings`: Real-time tracking of pledged ISINs, quantity, depository, and beneficiary accounts.
  - `clearing_member_collateral_allocations`: Multi-CCP collateral distribution between NSCCL and ICCL.
  - `collateral_haircut_master`: Ingested daily VaR, ELM, and additional haircut percentages per ISIN.
  - `clearing_corporation_interop_routes`: Active CC routing rules, preferred CCP designations, and open exposure balances.
  - `collateral_segregation_snapshots`: Daily EOD client collateral segregation records for SEBI audit.
  - `pledge_audit_trail_ledger`: Immutable append-only log of every depository handshake and response payload.
  - `pledge_platform_fee_ledgers`: Records the 0.00% (Zero Fee) platform fee assessed on collateral transactions.
- **Redis Cache Keys:**
  - `pledge:tpin:session:{session_id}`: Ephemeral depository authorization session (TTL: 180s).
  - `haircut:isin:{isin}`: Cached VaR + ELM haircut basis points and active status.
  - `collateral:holding:{account_id}:{isin}`: Current quantity pledged to TM and re-pledged to CC.
  - `lock:collateral:account:{account_id}`: Distributed mutex ensuring single-flight collateral operations.
- **Kafka Topics:**
  - **Consumes:**
    - `depository.tpin.callback.v1`: Ingests depository authorization webhooks.
    - `market.ticker.v1`: Real-time mark-to-market prices for revaluing pledged collateral.
    - `risk.margin_call.v1`: Ingests critical margin deficit alerts requiring collateral invocation.
    - `clearing.collateral_release_requested.v1`: Ingests client requests to unpledge collateral.
  - **Publishes:**
    - `clearing.pledge_initiated.v1`: Emitted when pledge request is sent to depository.
    - `clearing.pledge_confirmed.v1`: Emitted when depository confirms successful pledge to TM.
    - `clearing.repledge_completed.v1`: Emitted when CC confirms re-pledge allocation.
    - `clearing.pledge_released.v1`: Emitted when unpledge is processed by depository.
    - `margin.collateral_credit_updated.v1`: Emitted to update client non-cash margin credit.
    - `clearing.fee_assessed.v1`: Emitted to log the 0.00% (No fee at all) platform fee.
- **External Depository Endpoints:**
  - NSDL DPM (Depository Participant Module) and CCPS Gateways over dedicated leased lines / IPSec VPN.
  - CDSL CDAS API & e-DIS Gateway with TPIN authentication redirects.
- **Internal Microservices:**
  - **Risk & Margin Service (Prompt 206):** Ingests live collateral margin credit; validates unpledge solvency.
  - **Wallet Ledger Service (Prompt 203):** Records non-cash collateral balances and logs fee distributions.
  - **Cross-Asset Collateral Optimizer (Prompt 255):** Ingests holding data to optimize cross-margining allocations.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain Tokenized Collateral Freeze/Lock:** For securities tokenized on the National Blockchain Stock Exchange (NBSE) permissioned Hyperledger Besu network, pledging initiates an atomic smart contract lock:
  - Invokes `lockCollateral(address token, address investor, uint256 amount, bytes32 depositoryRef)` on `CollateralRegistry.sol`.
  - Freezes the designated tokens in the client's on-chain wallet, preventing secondary market transfers, withdrawals, or lending while pledged.
- **ERC-3643 Collateral Claim Attributes:** Updates the client's decentralized identity attributes in the on-chain Identity Registry:
  - Adds a dynamic claim attribute: `TOPIC_COLLATERAL_PLEDGE` with value encoding the designated clearing corporation (`NSCCL` or `ICCL`), pledge sequence number, and expiration.
  - Interacts with `ComplianceModule.sol` to enforce that transfer restrictions remain active until the Clearing Corporation releases the pledge.
- **Zero PII Protection & Salted BOID Hashes:**
  - On-chain transactions contain zero Personally Identifiable Information (no PAN, name, email, or plain Demat account numbers).
  - Depository Beneficiary Owner IDs (BOIDs) are cryptographically hashed using salted HMAC-SHA256:
    $$\text{AccountTag} = \text{HMAC-SHA256}(\text{BOID}, \text{Salt}_{\text{Ledger}})$$
- **1:1 Custody Notarization & Merkle Reconciliation:**
  - At the end of each clearing settlement cycle, an automated Merkle root representing aggregate depository pledge holdings (`semt.013`) is notarized into `CustodyNotary.sol` on Hyperledger Besu, ensuring tamper-proof cryptographic auditability of 1:1 custodial backing.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go 1.22+ module `services/clearing-pledge-service` with standardized directory layout (`cmd/server`, `internal/depository/cdsl`, `internal/depository/nsdl`, `internal/ccinterop`, `internal/haircut`, `internal/segregation`, `internal/blockchain`, `internal/storage`).
2. **Define Protobuf Service Contracts:** Author `proto/growww/clearing/v1/clearing_pledge_service.proto` specifying RPC definitions for `InitiateMarginPledge`, `VerifyPledgeOtp`, `InitiateMarginRePledge`, `InitiateReleasePledge`, `GetCollateralValuation`, and streaming event feeds.
3. **Generate Go Code Stubs:** Generate Go gRPC and Protobuf bindings using `protoc` with `protoc-gen-go` and `protoc-gen-go-grpc`.
4. **Author PostgreSQL Schema & Database Migrations:** Create SQL migration scripts defining `collateral_pledge_requests`, `demat_collateral_holdings`, `clearing_member_collateral_allocations`, `collateral_haircut_master`, and fee ledgers with appropriate foreign keys and performance indexes.
5. **Implement ISO 20022 Message Translator:** Develop robust XML/JSON marshallers and unmarshallers for ISO 20022 schemas: `colr.003` (Pledge Request), `colr.004` (Pledge Confirmation), `semt.013` (Statement of Pledged Positions), and `sese.023` (Unpledge / Release).
6. **Implement Depository Security & PKCS#7 Signer:** Build hardware/software cryptographic signing utility using PKCS#7 detached signatures and mTLS client certificates for authenticating requests sent to NSDL and CDSL CCPS gateways.
7. **Build CDSL TPIN & NSDL OTP Handshake Engine:** Implement the e-DIS authorization lifecycle: generate secure redirect URLs with unique session nonces, handle callback verification, validate OTP tokens, and enforce a strict 180-second session expiry.
8. **Develop Core Pledge State Machine:** Construct a resilient, event-driven finite state machine (FSM) governing pledge status progression: `INITIATED` -> `AUTH_PENDING` -> `PLEDGED_TO_TM` -> `RE_PLEDGE_INITIATED` -> `RE_PLEDGED_TO_CC` -> `RELEASE_PENDING` -> `RELEASED` / `INVOKED`.
9. **Build Clearing Corporation Interoperability Router:** Create algorithmic collateral routing logic between NSCCL and ICCL, dynamically evaluating current margin requirements, open position exposures, and intraday cash/non-cash ratios across both CCPs.
10. **Implement Dynamic Haircut Ingestion & Valuation Engine:** Build automated parsers for daily Clearing Corporation haircut files (`C_VAR1`, `C_VAR2`, and illiquidity multipliers). Calculate effective margin credit using:
    $$\text{MarginCredit} = \sum_{i=1}^{n} Q_i \cdot P_i \cdot \left(1 - \max(\text{VaR}_i + \text{ELM}_i + \text{AddOn}_i, H_{\min})\right)$$
11. **Implement Pre-Release Margin Verification Pipeline:** Connect synchronously to Risk & Margin Service (Prompt 206) via gRPC to evaluate post-release portfolio margin solvency before initiating depository unpledge operations.
12. **Build SEBI Collateral Segregation & Reporting Engine:** Implement the daily EOD collateral reporting engine to generate standard SEBI-compliant client collateral allocation files, verifying that client collateral strictly covers corresponding client margin requirements.
13. **Implement Hyperledger Besu ERC-3643 Relayer:** Integrate `go-ethereum/ethclient` to invoke `lockCollateral()` and set `COLLATERAL_PLEDGED_NSCCL` / `COLLATERAL_PLEDGED_ICCL` claim attributes on permissioned security tokens.
14. **Implement Universal 0.00% (No fee at all) Platform Fee Ledger:** Automatically assess the 0.00% (Zero Fee) platform fee across collateral operations, recording the deterministic 0.00% fee at launch (governed by FeeController.sol) revenue split in PostgreSQL.
15. **Write Unit, Integration, and Chaos Test Suites:** Write unit tests for ISO 20022 serialization, integration tests with mock NSDL/CDSL API stubs, concurrent state race condition tests, and latency benchmarks achieving sub-500 microsecond gRPC responses.

## Interfaces / Contracts

### Protobuf Service Contract (`proto/growww/clearing/v1/clearing_pledge_service.proto`)
```protobuf
syntax = "proto3";

package growww.clearing.v1;

option go_package = "github.com/growww/proto/gen/go/clearing/v1;clearingv1";

enum DepositoryType {
  DEPOSITORY_TYPE_UNSPECIFIED = 0;
  DEPOSITORY_TYPE_NSDL = 1;
  DEPOSITORY_TYPE_CDSL = 2;
}

enum ClearingCorporation {
  CLEARING_CORPORATION_UNSPECIFIED = 0;
  CLEARING_CORPORATION_NSCCL = 1; // NSE Clearing Limited
  CLEARING_CORPORATION_ICCL = 2;  // Indian Clearing Corporation Limited
}

enum PledgeStatus {
  PLEDGE_STATUS_UNSPECIFIED = 0;
  PLEDGE_STATUS_INITIATED = 1;
  PLEDGE_STATUS_AUTH_PENDING = 2;
  PLEDGE_STATUS_PLEDGED_TO_TM = 3;
  PLEDGE_STATUS_RE_PLEDGE_INITIATED = 4;
  PLEDGE_STATUS_RE_PLEDGED_TO_CC = 5;
  PLEDGE_STATUS_RELEASE_PENDING = 6;
  PLEDGE_STATUS_RELEASED = 7;
  PLEDGE_STATUS_INVOKED = 8;
  PLEDGE_STATUS_REJECTED = 9;
}

message CollateralItem {
  string isin = 1;
  uint64 quantity = 2;
  uint64 ltp_paise = 3;
  uint32 haircut_bps = 4; // Basis points (e.g., 2000 = 20.00%)
  uint64 effective_margin_credit_paise = 5;
}

message InitiateMarginPledgeRequest {
  string account_id = 1;
  string hashed_boid = 2;
  DepositoryType depository = 3;
  ClearingCorporation target_cc = 4;
  repeated CollateralItem items = 5;
  string client_ip = 6;
  string user_agent = 7;
}

message InitiateMarginPledgeResponse {
  string pledge_request_id = 1;
  PledgeStatus status = 2;
  string depository_auth_url = 3; // Redirect URL for CDSL TPIN / NSDL OTP
  string session_token = 4;
  int64 expires_at_ns = 5;
}

message VerifyPledgeOtpRequest {
  string pledge_request_id = 1;
  string session_token = 2;
  string otp_or_pin = 3;
}

message VerifyPledgeOtpResponse {
  string pledge_request_id = 1;
  PledgeStatus status = 2;
  string depository_transaction_id = 3;
  uint64 total_collateral_value_paise = 4;
  uint64 effective_margin_credit_paise = 5;
  uint64 platform_fee_paise = 6; // 0.00% (No fee at all) platform fee
  string error_message = 7;
}

message InitiateMarginRePledgeRequest {
  string pledge_request_id = 1;
  ClearingCorporation target_cc = 2;
  repeated CollateralItem items = 3;
}

message InitiateMarginRePledgeResponse {
  string repledge_id = 1;
  PledgeStatus status = 2;
  string cc_confirmation_ref = 3;
  uint64 allocated_margin_credit_paise = 4;
}

message InitiateReleasePledgeRequest {
  string account_id = 1;
  string pledge_request_id = 2;
  string isin = 3;
  uint64 quantity = 4;
}

message InitiateReleasePledgeResponse {
  string release_request_id = 1;
  PledgeStatus status = 2;
  bool margin_check_passed = 3;
  string depository_release_ref = 4;
}

message GetCollateralValuationRequest {
  string account_id = 1;
}

message GetCollateralValuationResponse {
  string account_id = 1;
  uint64 total_gross_value_paise = 2;
  uint64 total_haircut_paise = 3;
  uint64 net_margin_credit_paise = 4;
  uint64 cash_equivalent_paise = 5;
  uint64 non_cash_paise = 6;
  bool meets_50_50_cash_rule = 7;
  repeated CollateralItem holdings = 8;
  int64 evaluated_at_ns = 9;
}

service ClearingPledgeService {
  rpc InitiateMarginPledge (InitiateMarginPledgeRequest) returns (InitiateMarginPledgeResponse);
  rpc VerifyPledgeOtp (VerifyPledgeOtpRequest) returns (VerifyPledgeOtpResponse);
  rpc InitiateMarginRePledge (InitiateMarginRePledgeRequest) returns (InitiateMarginRePledgeResponse);
  rpc InitiateReleasePledge (InitiateReleasePledgeRequest) returns (InitiateReleasePledgeResponse);
  rpc GetCollateralValuation (GetCollateralValuationRequest) returns (GetCollateralValuationResponse);
}
```

### PostgreSQL Database Schema (`services/clearing-pledge-service/migrations/001_initial_schema.sql`)
```sql
-- PostgreSQL Migration: Clearing Corporation Interoperability & Demat Pledge Gateway
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enum Definitions
CREATE TYPE depository_type_enum AS ENUM ('NSDL', 'CDSL');
CREATE TYPE clearing_corporation_enum AS ENUM ('NSCCL', 'ICCL');
CREATE TYPE pledge_status_enum AS ENUM (
    'INITIATED',
    'AUTH_PENDING',
    'PLEDGED_TO_TM',
    'RE_PLEDGE_INITIATED',
    'RE_PLEDGED_TO_CC',
    'RELEASE_PENDING',
    'RELEASED',
    'INVOKED',
    'REJECTED'
);

-- Haircut Master Table
CREATE TABLE collateral_haircut_master (
    isin VARCHAR(12) PRIMARY KEY,
    symbol VARCHAR(32) NOT NULL,
    company_name VARCHAR(128) NOT NULL,
    security_type VARCHAR(16) NOT NULL, -- 'EQUITY', 'GOV_BOND', 'ETF', 'MUTUAL_FUND'
    var_haircut_bps INT NOT NULL DEFAULT 1500, -- Basis points (e.g., 15.00%)
    elm_haircut_bps INT NOT NULL DEFAULT 500,  -- Basis points (e.g., 5.00%)
    additional_haircut_bps INT NOT NULL DEFAULT 0,
    total_haircut_bps INT GENERATED ALWAYS AS (var_haircut_bps + elm_haircut_bps + additional_haircut_bps) STORED,
    is_approved_for_pledge BOOLEAN NOT NULL DEFAULT TRUE,
    is_cash_equivalent BOOLEAN NOT NULL DEFAULT FALSE,
    clearing_corp clearing_corporation_enum NOT NULL DEFAULT 'NSCCL',
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Master Pledge Request Table
CREATE TABLE collateral_pledge_requests (
    pledge_request_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(64) NOT NULL,
    hashed_boid VARCHAR(64) NOT NULL,
    depository depository_type_enum NOT NULL,
    target_cc clearing_corporation_enum NOT NULL DEFAULT 'NSCCL',
    status pledge_status_enum NOT NULL DEFAULT 'INITIATED',
    depository_txn_id VARCHAR(64),
    cc_repledge_ref VARCHAR(64),
    gross_collateral_value_paise BIGINT NOT NULL DEFAULT 0,
    net_margin_credit_paise BIGINT NOT NULL DEFAULT 0,
    platform_fee_paise BIGINT NOT NULL DEFAULT 0, -- 0.00% (No fee at all) platform fee
    auth_session_nonce VARCHAR(64),
    auth_url TEXT,
    auth_expires_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pledged Item Line Details
CREATE TABLE demat_collateral_holdings (
    holding_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pledge_request_id UUID NOT NULL REFERENCES collateral_pledge_requests(pledge_request_id) ON DELETE CASCADE,
    account_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL REFERENCES collateral_haircut_master(isin),
    pledged_quantity BIGINT NOT NULL CHECK (pledged_quantity > 0),
    repledged_to_cc_quantity BIGINT NOT NULL DEFAULT 0,
    ltp_paise BIGINT NOT NULL DEFAULT 0,
    haircut_bps INT NOT NULL DEFAULT 2000,
    effective_margin_paise BIGINT NOT NULL DEFAULT 0,
    is_released BOOLEAN NOT NULL DEFAULT FALSE,
    is_invoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Clearing Member Interoperability Allocation Table
CREATE TABLE clearing_member_collateral_allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id VARCHAR(64) NOT NULL,
    clearing_corp clearing_corporation_enum NOT NULL,
    isin VARCHAR(12) NOT NULL,
    allocated_quantity BIGINT NOT NULL,
    allocated_margin_credit_paise BIGINT NOT NULL,
    cc_collateral_account_id VARCHAR(32) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Daily SEBI Collateral Segregation Reporting Table
CREATE TABLE collateral_segregation_snapshots (
    snapshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_date DATE NOT NULL DEFAULT CURRENT_DATE,
    account_id VARCHAR(64) NOT NULL,
    total_margin_required_paise BIGINT NOT NULL,
    total_collateral_available_paise BIGINT NOT NULL,
    cash_collateral_paise BIGINT NOT NULL,
    non_cash_pledged_paise BIGINT NOT NULL,
    collateral_allocated_to_cc_paise BIGINT NOT NULL,
    collateral_retained_with_tm_paise BIGINT NOT NULL,
    proprietary_co_mingling_flag BOOLEAN NOT NULL DEFAULT FALSE,
    audit_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Platform fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Distribution Ledger
CREATE TABLE pledge_platform_fee_ledgers (
    fee_ledger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pledge_request_id UUID NOT NULL REFERENCES collateral_pledge_requests(pledge_request_id),
    account_id VARCHAR(64) NOT NULL,
    turnover_paise BIGINT NOT NULL,
    fee_paise BIGINT NOT NULL, -- 0.00% (Zero Fee) of turnover
    treasury_split_paise BIGINT NOT NULL, -- Governed by FeeController (0.00% at launch)
    sgf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    ipf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Immutable Audit Trail
CREATE TABLE pledge_audit_trail_ledger (
    audit_id BIGSERIAL PRIMARY KEY,
    pledge_request_id UUID NOT NULL,
    event_name VARCHAR(64) NOT NULL,
    previous_state VARCHAR(32),
    new_state VARCHAR(32) NOT NULL,
    payload_json JSONB NOT NULL,
    signature_bytes BYTEA,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for Microsecond Queries
CREATE INDEX idx_pledge_req_account ON collateral_pledge_requests(account_id, status);
CREATE INDEX idx_demat_holdings_account_isin ON demat_collateral_holdings(account_id, isin);
CREATE INDEX idx_cc_allocations_cc ON clearing_member_collateral_allocations(clearing_corp, account_id);
CREATE INDEX idx_haircut_master_lookup ON collateral_haircut_master(isin, is_approved_for_pledge);
CREATE INDEX idx_segregation_date_account ON collateral_segregation_snapshots(snapshot_date, account_id);
```

## Security & Compliance Notes
- **SEBI Margin Pledge/Re-Pledge Framework Mandate:**
  - Compliance with SEBI circulars `SEBI/HO/MIRSD/DOP/CIR/P/2020/143` and `SEBI/HO/MRD2_DCAP/P/CIR/2021/0598`.
  - Securities remain credited exclusively to the client's own Demat BOID. Pledges are registered through depository CCPS infrastructure to the Trading Member Client Collateral Pledge Account (`TMCCPA`) and subsequently re-pledged to the Clearing Member / Clearing Corporation Account. Broker pool accounts are strictly prohibited.
- **Client Collateral Segregation & Zero Co-Mingling:**
  - Every client's non-cash collateral is segregated at the Clearing Corporation level using unique Client Identification (UCC).
  - No client collateral can be utilized to satisfy margin obligations of proprietary accounts or other clients. Daily automated reconciliations enforce zero variance.
- **Hardware Security Module (HSM) Digital Signatures:**
  - All outbound ISO 20022 and CCPS API payloads to NSDL and CDSL are signed with PKCS#7 detached signatures generated inside an HSM (FIPS 140-2 Level 3 compliant) using an enterprise RSA 4096-bit private key.
- **TPIN & OTP Ephemeral Lifecycle:**
  - CDSL TPIN authentication URLs and NSDL OTP verification tokens are tied to single-use cryptographically random 256-bit session nonces.
  - Hard expiry enforced at 180 seconds in Redis. Plaintext TPINs and OTPs are never stored in databases, logs, or persistent storage.
- **Zero PII Exposure:**
  - Demat Beneficiary Owner IDs (BOIDs) and PAN identifiers are hashed with salted HMAC-SHA256 before persistence or transmission across public message brokers.
- **Immutable WORM Audit Trail:**
  - All depository requests, acknowledgments, ISO 20022 payloads, and error responses are written to an append-only audit trail with cryptographic hash chaining (`prev_hash || curr_payload`).

## Acceptance Criteria
- [ ] Direct integration with NSDL and CDSL CCPS gateways formats and parses ISO 20022 `colr.003`, `colr.004`, `semt.013`, and `sese.023` payloads with 100% schema validation.
- [ ] Demat pledge authorization successfully executes CDSL TPIN redirect and NSDL OTP verification flows within 3 seconds of client input.
- [ ] Clearing Corporation Interoperability router seamlessly shifts collateral allocations between NSCCL and ICCL based on open margin requirements without duplicate pledging.
- [ ] Pre-release safety check blocks collateral unpledge if post-release account margin buffer falls below 105% of initial margin requirement.
- [ ] Collateral haircut engine automatically parses daily CC haircut circulars (`VAR_MARGIN_CSV`), accurately calculating combined VaR + ELM haircuts.
- [ ] Statutory 50:50 cash-to-non-cash collateral ratio is calculated and enforced on every margin revaluation event.
- [ ] EOD SEBI collateral segregation snapshot file generates deterministically with zero proprietary co-mingling detected.
- [ ] Hyperledger Besu ERC-3643 collateral lock smart contract executes `lockCollateral()` and sets on-chain compliance claim attributes with confirmed QBFT finality.
- [ ] 0.00% (No fee at all) platform fee is assessed accurately on collateral turnover with 0.00% fee at launch (governed by FeeController.sol) ledger distribution.
- [ ] Microsecond gRPC response time for `GetCollateralValuation` (< 500 microseconds p99 under 10,000 requests/sec).

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 101 (System Architecture & Blueprint)
  - Prompt 102 (Bounded Contexts & Domain Models)
  - Prompt 203 (Multi-Asset Wallet & Ledger Service)
  - Prompt 206 (Real-Time Risk & Pre-Trade Margin Engine)
- **Parallel Work:**
  - Prompt 213 (Custodian Depository Integration Service)
  - Prompt 241 (SPAN Portfolio Margin Engine)
  - Prompt 255 (Cross-Asset Portfolio Margining & Dynamic Collateral Optimization)
- **Enables:**
  - Prompt 225 (FIX Protocol Gateway & Order Routing)
  - Prompt 230 (Settlement Guarantee Fund & Default Waterfall Service)
  - Comprehensive demat collateral margin utilization for high-frequency equity and derivatives trading.
