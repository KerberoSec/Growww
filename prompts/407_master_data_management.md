# 407 - Master Data Management (Securities Master, Corporate Actions Master)

## Purpose
In a fractional investment platform trading real asset-backed Indian equities, reliable master data is the foundation of market integrity. Every tradable token must strictly correspond to an authentic, active, and SEBI-listed security. Trading engines, risk check systems, regulatory report generators, and client interfaces all require instantaneous, deterministic access to canonical asset identifiers, trading rules, tick sizes, circuit filter limits, and scheduled corporate actions.

The Master Data Management (MDM) Service maintains the golden record for all tokenized instruments. It automates the ingestion of daily exchange security master files and Bhavcopy data from the National Stock Exchange (NSE), Bombay Stock Exchange (BSE), and depositories (NSDL/CDSL). Furthermore, it tracks corporate action lifecycles (dividends, stock splits, bonus issues, rights offerings), orchestrating downstream token supply adjustments with on-chain smart contracts.

## What You Are Building
A production-grade Master Data microservice and automated exchange ingestion engine containing:
- **Master Data Service:** A Python 3.12 / FastAPI microservice (`services/master-data-service/`) providing high-performance gRPC and REST APIs for security metadata lookups.
- **Exchange Master & Bhavcopy Ingestion Pipeline:** Automated daily workers downloading, validating, and parsing official NSE/BSE security masters, closing prices, and circulars.
- **Corporate Actions Lifecycle Engine:** Stateful tracking of corporate actions from announcement and board approval to ex-date, record date, and execution ratio verification.
- **Sub-Millisecond Redis Master Cache:** In-memory pre-warmed cache providing instant lookup of security trading parameters (tick size, lot size, circuit breakers, status) for matching and risk engines.
- **Kafka Event Publisher:** Emits change notifications to compacted Kafka topics (`market.securities_master`, `market.corporate_actions`) whenever security trading status or corporate action parameters change.
- **Multi-Party Approval (Maker-Checker) Portal:** Administrative API enforcing two-person rule approval before activating newly tokenized securities or confirming corporate action execution ratios.

## Scope Boundaries
- **In Scope:** Canonical security schemas, ISIN/symbol resolution, Bhavcopy parsing, corporate action registry, gRPC/REST endpoints, Redis cache warming, and event publishing.
- **Out of Scope / Handled Elsewhere:**
 - On-chain token mint/burn execution for corporate actions (handled in Prompt 222 and Prompt 303).
 - High-level investor search and catalog filtering (handled in Prompt 221).
 - In-memory order matching engine state (handled in Prompt 205).
 - Physical custodian share settlement (handled in Prompt 213).

## Technology to Use
Python 3.12 with FastAPI, SQLAlchemy 2.0 (async), and gRPC is selected for the Master Data service. Python provides superior ecosystem tooling for parsing diverse financial data feeds (CSV, fixed-width text, XML, JSON), while FastAPI and gRPC deliver high-throughput, low-latency API serving. In-memory caching via Redis 7.2 ensures that pre-trade validation checks execute in microseconds without database round-trips.

- **Service Framework:** Python 3.12, FastAPI 0.110+, SQLAlchemy 2.0 (asyncio).
- **Inter-Service Protocol:** gRPC / Protocol Buffers v3 for internal microservices; REST / JSON for back-office admin.
- **Database Engine:** PostgreSQL 16 (`master_data` schema).
- **Cache Engine:** Redis 7.2 Cluster (Hashes with automated pre-warming).
- **Messaging:** Kafka (compacted topic `market.securities_master`).

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Relational storage for master tables (`securities`, `corporate_actions`, `exchanges`, `sectors`).
- **Redis 7.2:** Hot cache of tradable symbols and trading limits.
- **SFTP / HTTPS Ingestion Connectors:** Secure connectors pulling daily feeds from NSE/BSE/NSDL/CDSL.
- **Kafka Cluster:** Publishes security updates and corporate action announcements.

## Blockchain Interaction
The Master Data service establishes the immutable linkage between traditional physical equities and on-chain ERC-3643 digital tokens on Hyperledger Besu.

### Detailed On-Chain Integration Mechanics:
- **ISIN-to-Contract Binding:** Binds canonical 12-character ISINs (e.g. `INE002A01018` for Reliance Industries) to their deployed `DigitalSecurityToken.sol` smart contract addresses.
- **Fractional Divisibility Standard:** Enforces standard token divisibility (decimals = 6, allowing trades down to 0.000001 shares) across off-chain master data and on-chain contract parameters.
- **Corporate Action Parameter Commitment:** When a corporate action (e.g. 1:1 stock split) reaches the record date, verified execution parameters are signed via HSM and committed to `MultiSigGovernance.sol` to authorize smart contract supply adjustments.
- **Zero PII:** Master data contains purely public asset metadata; zero customer or investor information is involved.

## Step-by-Step Build Instructions
1. Scaffold project repository: `services/master-data-service/` with `app/`, `proto/`, `migrations/`, and `tests/`.
2. Implement PostgreSQL DDL migrations for `master_data.securities`, `master_data.corporate_actions`, and `master_data.price_history`.
3. Implement ISO 6166 checksum validation algorithm for 12-character Indian ISINs (`IN[A-Z0-9]{10}`).
4. Author daily exchange Bhavcopy ingestion worker parsing NSE/BSE CSV/XML feeds and updating daily reference closing prices.
5. Author corporate actions ingestion worker tracking dividends, splits, bonuses, and rights offerings with state machine (`ANNOUNCED`, `APPROVED`, `EX_DATE_PENDING`, `EXECUTED`, `CANCELLED`).
6. Implement Maker-Checker approval workflow requiring two distinct compliance officer approvals for any manual security parameter override.
7. Define Protobuf definitions (`securities_master.proto`) and implement high-performance gRPC server.
8. Implement REST endpoints for administrative back-office management and audit queries.
9. Implement Redis cache layer pre-warming all active securities into Redis Hashes upon service startup.
10. Integrate Kafka event publisher emitting compacted records to `market.securities_master` upon any status transition (e.g. `HALTED`, `ACTIVE`).
11. Implement comprehensive unit tests verifying ISIN validation, Bhavcopy parsing, and split ratio math.
12. Benchmark gRPC lookup latency: verify >10,000 requests/sec with p99 latency < 2ms under concurrent load.

## Interfaces / Contracts

```protobuf
// Protocol Buffers: securities_master.proto
syntax = "proto3";

package growww.masterdata.v1;

service SecuritiesMasterService {
  rpc GetSecurityByISIN (GetSecurityByISINRequest) returns (SecurityResponse);
  rpc GetSecurityBySymbol (GetSecurityBySymbolRequest) returns (SecurityResponse);
  rpc ListActiveSecurities (ListActiveSecuritiesRequest) returns (ListSecuritiesResponse);
  rpc GetCorporateActions (GetCorporateActionsRequest) returns (CorporateActionsResponse);
}

message SecurityResponse {
  string isin = 1;
  string symbol = 2;
  string name = 3;
  string token_contract_address = 4;
  int32 token_decimals = 5;
  string exchange = 6;
  string sector = 7;
  string tick_size = 8;
  int32 lot_size = 9;
  string upper_circuit_limit = 10;
  string lower_circuit_limit = 11;
  string trading_status = 12; // "ACTIVE", "HALTED", "DELISTED"
  int64 last_updated_at = 13;
}

message GetSecurityByISINRequest {
  string isin = 1;
}

message GetSecurityBySymbolRequest {
  string symbol = 1;
}

message ListActiveSecuritiesRequest {
  int32 page_size = 1;
  string page_token = 2;
}

message ListSecuritiesResponse {
  repeated SecurityResponse securities = 1;
  string next_page_token = 2;
}

message GetCorporateActionsRequest {
  string isin = 1;
  string status = 2;
}

message CorporateActionsResponse {
  repeated CorporateActionItem actions = 1;
}

message CorporateActionItem {
  string action_id = 1;
  string isin = 2;
  string action_type = 3; // "DIVIDEND", "SPLIT", "BONUS", "RIGHTS"
  string ratio_from = 4;
  string ratio_to = 5;
  string dividend_amount_per_share = 6;
  int64 ex_date = 7;
  int64 record_date = 8;
  string status = 9;
}
```

```sql
-- DDL Schema: master_data.securities & corporate_actions
CREATE SCHEMA IF NOT EXISTS master_data;

CREATE TABLE master_data.securities (
    isin CHAR(12) PRIMARY KEY, -- ISO 6166 Compliant (e.g. INE002A01018)
    symbol VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    token_contract_address CHAR(42) NOT NULL UNIQUE,
    token_decimals INT NOT NULL DEFAULT 6,
    exchange VARCHAR(16) NOT NULL CHECK (exchange IN ('NSE', 'BSE', 'BOTH')),
    sector VARCHAR(64) NOT NULL,
    tick_size NUMERIC(8, 4) NOT NULL DEFAULT 0.0500,
    lot_size INT NOT NULL DEFAULT 1,
    upper_circuit_limit NUMERIC(18, 4),
    lower_circuit_limit NUMERIC(18, 4),
    trading_status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE' CHECK (trading_status IN ('ACTIVE', 'HALTED', 'SUSPENDED', 'DELISTED')),
    depository VARCHAR(16) NOT NULL DEFAULT 'BOTH' CHECK (depository IN ('NSDL', 'CDSL', 'BOTH')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE master_data.corporate_actions (
    action_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin CHAR(12) REFERENCES master_data.securities(isin) ON DELETE RESTRICT,
    action_type VARCHAR(32) NOT NULL CHECK (action_type IN ('CASH_DIVIDEND', 'STOCK_SPLIT', 'BONUS_ISSUE', 'RIGHTS_OFFERING')),
    ratio_from INT, -- e.g. 1 in 1:2 split
    ratio_to INT,   -- e.g. 2 in 1:2 split
    cash_amount_per_share NUMERIC(18, 4),
    announcement_date DATE NOT NULL,
    ex_date DATE NOT NULL,
    record_date DATE NOT NULL,
    execution_status VARCHAR(32) NOT NULL DEFAULT 'ANNOUNCED' CHECK (execution_status IN ('ANNOUNCED', 'APPROVED', 'EXECUTED', 'CANCELLED')),
    approved_by_maker UUID,
    approved_by_checker UUID,
    on_chain_tx_hash CHAR(66),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);
```

## Security & Compliance Notes
- **ISO 6166 ISIN Validation:** Strict input sanitation and Luhn-mod-10/ISO-6166 checksum validation prevents injection of bogus or unauthorized securities.
- **Maker-Checker Security Control:** Any manual modification to security parameters (e.g. overriding circuit filters or emergency trading halt) requires multi-party authorization (Maker initiates, distinct Checker approves).
- **SEBI Listing Regulatory Compliance:** Daily reconciliation ensures securities suspended or debarred by SEBI/Exchanges are immediately halted in the matching engine and on-chain transfer gate.
- **Audit Logging:** Every master data modification is permanently recorded in append-only audit tables with timestamp, admin ID, IP, and cryptographic signature.

## Acceptance Criteria
- [ ] Master Data microservice implemented with async FastAPI and gRPC server interfaces.
- [ ] PostgreSQL schemas `master_data.securities` and `master_data.corporate_actions` created with integrity constraints.
- [ ] Daily Bhavcopy and security master parsers tested with authentic NSE/BSE sample feed files.
- [ ] Redis cache warming on startup loads all active securities; gRPC benchmark achieves >10,000 req/sec with p99 < 2ms.
- [ ] Compacted Kafka topic `market.securities_master` receives change events immediately upon security parameter updates.
- [ ] Maker-Checker workflow verified: single-admin approval attempts fail; dual-admin approval successfully promotes status.
- [ ] ISIN-to-contract mapping binds physical shares to Hyperledger Besu `DigitalSecurityToken.sol` addresses accurately.
- [ ] Corporate action lifecycle state machine validated from announcement through record date execution.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 103 (API Standards), Prompt 104 (Kafka Standards), Prompt 111 (Domain Model), Prompt 401 (PostgreSQL), Prompt 402 (Redis).
- **Parallel Tasks:** Prompt 205 (Order Matching Engine), Prompt 221 (Search & Discovery), Prompt 222 (Corporate Actions Service).
