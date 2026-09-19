# 209 - Portfolio & Fractional Holdings Accounting Service (Python / FastAPI)

## Purpose
The Portfolio & Fractional Holdings Accounting Service is the investment ledger of record for Growww investors. It tracks user-level fractional equity positions (accurate to 6 decimal places), calculates cost basis using tax-lot accounting (First-In, First-Out [FIFO] and Weighted Average Cost), tracks unrealized and realized Profit & Loss (P&L), and aggregates portfolio asset allocation metrics.

Because Growww tokenizes Indian equities into fractional units backed 1:1 by custodian demat holdings, this service continuously synchronizes off-chain investor tax-lot portfolios with the cryptographic token balances held in each investor's on-chain public address on Hyperledger Besu.

## What You Are Building
A high-precision, asynchronous Python FastAPI microservice (`services/portfolio-service`). Concrete deliverables include:
- REST and gRPC endpoints providing real-time portfolio summaries, individual holding breakdowns, tax-lot histories, and performance analytics.
- High-precision financial accounting engine managing fractional share balances and cost basis without floating-point errors.
- Real-time mark-to-market valuation engine calculating unrealized P&L against live ticker feeds from the Market Data Service.
- Tax-lot manager tracking every acquisition lot (`lot_id`, `quantity`, `remaining_quantity`, `cost_per_unit_inr`, `acquired_at`) for SEBI/Income Tax compliance.
- Kafka consumer for `trade.settled.v1` and `corporate_action.applied.v1`.
- On-chain token balance verification worker ensuring off-chain database holdings match Hyperledger Besu smart contract balances ($100\%$ reconciliation).

## Scope Boundaries
- **In Scope:**
 - Investor fractional equity holdings state per ISIN.
 - Tax-lot accounting (FIFO allocation upon partial or full liquidation).
 - Unrealized P&L and daily P&L calculation against live market data.
 - Holding holds/reservations for pending sell orders.
 - Synchronizing portfolio state following corporate actions (splits, bonuses).
- **Out of Scope / Handled Elsewhere:**
 - Order intake and matching (Prompt 204, 205).
 - Profit-based platform fee calculation and deduction (Prompt 210).
 - Cash wallet balance management (Prompt 203).
 - Statutory annual capital gains tax filing statement generation (Prompt 223).

## Technology to Use
- **Primary Language & Framework:** Python 3.12 with FastAPI (0.111+). Python is chosen for its native arbitrary-precision decimal support (`decimal.Decimal`), robust mathematical libraries, and async SQLAlchemy integration for complex financial tax-lot queries.
- **Database & Storage:** PostgreSQL 16+ using asyncpg and SQLAlchemy 2.0 (async ORM); Redis 7.2 for caching real-time portfolio valuations.
- **Financial Arithmetic:** Native Python `decimal.Decimal` configured with `ROUND_HALF_UP` precision to 6 decimal places for shares and 4 decimal places for INR.
- **Streaming & Messaging:** `aiokafka` for async event consumption; `web3.py` for direct JSON-RPC on-chain token balance verification.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `user_holdings`, `tax_lots`, `holding_holds`, `portfolio_daily_snapshots`.
- **Redis 7.2:** Caches calculated portfolio summaries with 5-second TTL and on-demand cache busting.
- **Apache Kafka:** Consumes `trade.settled.v1`, `corporate_action.applied.v1`; publishes `portfolio.updated.v1`.
- **Market Data Service (Prompt 207):** Ingests live prices to compute mark-to-market portfolio values.
- **Hyperledger Besu Nodes:** Queries `DigitalSecurityToken.balanceOf(ledgerAddress)` for balance verification.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **1:1 Token Reconciliation Invariant:** For every user holding record in PostgreSQL, the sum of remaining quantities across active tax lots must exactly equal the token balance returned by the corresponding `DigitalSecurityToken.sol` smart contract on Hyperledger Besu:
  $$\sum \text{tax\_lots.remaining\_quantity} = \text{DigitalSecurityToken.balanceOf}(\text{user\_ledger\_address})$$
- **On-Chain Balance Sync:** The service listens to on-chain `Transfer` events emitted by `DigitalSecurityToken.sol` under QBFT consensus to ensure that all transfers executed during atomic DvP settlements are immediately reflected in user portfolio views.
- **Zero On-Chain PII:** The blockchain tracks token balances by public address (`0x...`); the link between user identity and holdings exists solely within this secure, DPDP-compliant service.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize `services/portfolio-service` using Poetry/uv, Python 3.12, strict `ruff` and `mypy` configurations.
2. **Define Protobuf Contracts:** Create `proto/growww/portfolio/v1/portfolio_service.proto` for portfolio summaries, holding details, and tax lot inspection.
3. **Generate gRPC Stubs:** Compile protobuf schemas to async Python stubs.
4. **Design Database Schema & Migrations:** Write Alembic migrations for `user_holdings`, `tax_lots`, `holding_holds`, and `portfolio_daily_snapshots`.
5. **Implement Fixed-Precision Tax-Lot Engine:** Build core accounting classes using `decimal.Decimal` to manage lot creation on buy fills and FIFO lot depletion on sell fills.
6. **Implement Real-Time Mark-to-Market Calculator:** Build calculation engine computing:
 - Total Portfolio Value $= \sum (\text{holding.quantity} \times \text{market.ltp})$
 - Unrealized P&L $= \text{Total Current Value} - \text{Total Invested Cost}$
 - Total Invested Cost $= \sum (\text{tax\_lot.remaining\_quantity} \times \text{tax\_lot.cost\_per\_unit})$
7. **Implement Share Hold Reservation:** Build gRPC methods (`ReserveHolding`, `ReleaseHolding`, `CommitHolding`) for sell orders, ensuring users cannot place orders for more shares than their available unheld balance.
8. **Build Trade Settlement Consumer:** Implement Kafka consumer for `trade.settled.v1`:
 - If Buy: create new `tax_lot` and increment `user_holdings.total_quantity`.
 - If Sell: deplete oldest active `tax_lots` via FIFO, record realized cost basis, and decrement `user_holdings.total_quantity`.
9. **Build Corporate Actions Processor:** Ingest `corporate_action.applied.v1` to adjust tax lots and holdings proportionately for stock splits and bonus issues.
10. **Implement On-Chain Balance Verifier:** Build background reconciliation job using `web3.py` that queries Besu node smart contracts and raises alerts if off-chain holdings deviate from on-chain tokens.
11. **Implement Redis Caching Layer:** Cache aggregated portfolio summaries in Redis; invalidate immediately on settlement ingestion.
12. **Expose REST & gRPC API Endpoints:** Build FastAPI router exposing `/v1/portfolio/summary`, `/v1/portfolio/holdings`, and gRPC `PortfolioService` endpoints.
13. **Configure Prometheus Metrics & Health:** Expose `/metrics` for portfolio query latency, cache hit ratios, and reconciliation mismatch gauges.
14. **Write Comprehensive Test Suite:** Implement test suite using `pytest-asyncio` verifying FIFO tax-lot depletion, zero floating-point drift over 10,000 randomized trades, and exact on-chain balance parity.

## Interfaces / Contracts

### Protobuf Definition (`portfolio_service.proto`)
```protobuf
syntax = "proto3";

package growww.portfolio.v1;

option go_package = "growww/portfolio/v1;portfoliov1";

service PortfolioService {
  rpc GetPortfolioSummary (GetPortfolioSummaryRequest) returns (GetPortfolioSummaryResponse);
  rpc GetHoldingDetail (GetHoldingDetailRequest) returns (GetHoldingDetailResponse);
  rpc ListHoldings (ListHoldingsRequest) returns (ListHoldingsResponse);
  rpc ReserveHolding (ReserveHoldingRequest) returns (ReserveHoldingResponse);
  rpc ReleaseHolding (ReleaseHoldingRequest) returns (ReleaseHoldingResponse);
}

message GetPortfolioSummaryRequest {
  string user_id = 1;
}

message GetPortfolioSummaryResponse {
  string user_id = 1;
  string total_portfolio_value_inr = 2; // Decimal string
  string total_invested_value_inr = 3;
  string total_unrealized_pnl_inr = 4;
  string total_unrealized_pnl_percent = 5;
  string todays_pnl_inr = 6;
  string todays_pnl_percent = 7;
  int32 total_holdings_count = 8;
  int64 updated_at_unix = 9;
}

message HoldingItemDto {
  string isin = 1;
  string symbol = 2;
  string company_name = 3;
  string total_quantity = 4; // Fractional quantity (e.g. "1.250000")
  string available_quantity = 5;
  string average_cost_price_inr = 6;
  string current_market_price_inr = 7;
  string current_value_inr = 8;
  string unrealized_pnl_inr = 9;
  string unrealized_pnl_percent = 10;
  string token_contract_address = 11;
}

message GetHoldingDetailRequest {
  string user_id = 1;
  string isin = 2;
}

message GetHoldingDetailResponse {
  HoldingItemDto holding = 1;
  repeated TaxLotDto tax_lots = 2;
}

message TaxLotDto {
  string lot_id = 1;
  string original_quantity = 2;
  string remaining_quantity = 3;
  string purchase_price_inr = 4;
  int64 acquired_at_unix = 5;
}

message ListHoldingsRequest {
  string user_id = 1;
}

message ListHoldingsResponse {
  repeated HoldingItemDto holdings = 1;
}

message ReserveHoldingRequest {
  string idempotency_key = 1;
  string user_id = 2;
  string order_id = 3;
  string isin = 4;
  string quantity = 5;
}

message ReserveHoldingResponse {
  string hold_id = 1;
  bool success = 2;
  string remaining_available_quantity = 3;
}

message ReleaseHoldingRequest {
  string hold_id = 1;
  string reason = 2;
}

message ReleaseHoldingResponse {
  bool success = 1;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE user_holdings (
    holding_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    isin VARCHAR(12) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    total_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    held_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_isin UNIQUE(user_id, isin),
    CONSTRAINT chk_positive_holdings CHECK (total_quantity >= held_quantity AND held_quantity >= 0)
);

CREATE TABLE tax_lots (
    lot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    holding_id UUID NOT NULL REFERENCES user_holdings(holding_id) ON DELETE CASCADE,
    trade_id UUID NOT NULL,
    original_quantity NUMERIC(18, 6) NOT NULL,
    remaining_quantity NUMERIC(18, 6) NOT NULL CHECK (remaining_quantity >= 0),
    cost_per_unit_inr NUMERIC(18, 4) NOT NULL,
    acquired_at TIMESTAMPTZ NOT NULL,
    is_exhausted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE holding_holds (
    hold_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    holding_id UUID NOT NULL REFERENCES user_holdings(holding_id) ON DELETE CASCADE,
    order_id UUID NOT NULL,
    held_quantity NUMERIC(18, 6) NOT NULL CHECK (held_quantity > 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_holdings_user ON user_holdings(user_id);
CREATE INDEX idx_tax_lots_holding_active ON tax_lots(holding_id, acquired_at) WHERE NOT is_exhausted;
```

## Security & Compliance Notes
- **Mathematical Decimal Invariant:** All calculations strictly use 128-bit fixed-point decimal objects; no floating-point conversions are permitted anywhere in the code.
- **PMLA / SEBI Auditability:** Tax lot acquisition dates, original execution references, and FIFO depletion records are immutably archived for statutory capital gains auditing.
- **Data Privacy (DPDP Act 2023):** User portfolio balances are accessible strictly via authenticated sessions verified via mTLS and JWT authorization.

## Acceptance Criteria
- [ ] Portfolio service accurately calculates fractional share balances and average cost price per ISIN.
- [ ] FIFO tax-lot depletion correctly matches sold shares against the oldest active purchase lots.
- [ ] Mark-to-market calculations update unrealized P&L in real-time when new market prices arrive.
- [ ] `ReserveHolding` locks fractional shares and prevents users from placing duplicate sell orders exceeding available shares.
- [ ] Background reconciliation worker verifies 100% parity between off-chain database holdings and on-chain Besu token balances.
- [ ] Portfolio summary response time is $< 15\text{ms}$ at 1,000 req/sec.

## Suggested Order / Dependencies
- **Prerequisites:** 111 (Domain Model), 207 (Market Data), 208 (Trade Settlement), 401 (PostgreSQL Schema).
- **Parallel Tasks:** 210 (Fee Engine), 215 (Reconciliation Service).
- **Downstream Blockers:** 210 (Fee Engine), 510 (Flutter Portfolio Screen), 603 (Web Trading Dashboard).
