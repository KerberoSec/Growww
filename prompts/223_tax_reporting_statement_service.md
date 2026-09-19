# 223 - Tax Reporting & Capital Gains Statement Service (Income Tax / Sections 111A & 112A)

## Purpose
The Tax Reporting & Capital Gains Statement Service calculates, tracks, and certifies tax liabilities arising from equity trading, fractional security dispositions, and dividend distributions across Growww. Under the Indian Income Tax Act (1961) and Finance Act amendments, investors are subject to Short-Term Capital Gains (STCG under Section 111A) and Long-Term Capital Gains (LTCG under Section 112A), alongside Securities Transaction Tax (STT), Stamp Duty, and GST on brokerage/fees.

This service implements high-precision FIFO (First-In, First-Out) cost-basis lot tracking, grandfathering fair market value (FMV) provisions for pre-2018 acquisitions, Section 194 dividend TDS accounting, and generates certified, downloadable tax statements (P&L Statements, Schedule 112A summaries, Form 16A TDS certificates, and Excel tax packs) for both domestic Indian taxpayers and international GIFT City investors.

## What You Are Building
A deterministic, high-precision Go / Python microservice (`services/tax-service`) providing:
- **FIFO Cost-Basis Tax Lot Engine:** Tracks acquisition tranches, purchase prices, settlement dates, and disposal allocations.
- **Capital Gains Calculator:** Automates STCG (holding period $< 12$ months) and LTCG ($\ge 12$ months) computations with statutory tax slab tagging.
- **Statutory Tax Statement Renderer:** Generates PDF, Excel (XLSX), and JSON exports formatted for standard Indian ITR (ITR-2, ITR-3) and international filings.
- **Blockchain Settlement Verification:** Validates trade execution timestamps and settlement costs against Hyperledger Besu transaction receipts.
- **Artifacts Delivered:**
 - `services/tax-service/cmd/server/main.go` - Go microservice entry point.
 - `services/tax-service/internal/fifo/lot_matcher.go` - Deterministic FIFO tax lot matching engine.
 - `services/tax-service/internal/calculator/gains.go` - STCG/LTCG, STT, and grandfathering calculator.
 - `services/tax-service/internal/pdf/tax_statement.go` - Certified PDF generator with digital signatures.
 - `proto/growww/tax/v1/tax_service.proto` - Internal gRPC service contracts.

## Scope Boundaries
- **In Scope:**
 - Real-time maintenance of investor tax lots upon trade execution.
 - FIFO matching of sell transactions against earliest available acquisition lots.
 - Computation of STCG, LTCG, dividend income, STT, turnover, and net realized gains.
 - Generating downloadable annual and quarterly tax reports and Schedule 112A tables.
- **Out of Scope / Handled Elsewhere:**
 - Direct electronic e-filing with the Income Tax Department portal on behalf of users.
 - Real-time profit-based platform fee calculation (handled by Fee Engine, Prompt 210).
 - Corporate action cash dividend crediting (handled by Prompt 222).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+ (or Python 3.12 with Pandas/NumPy) utilizing `pgx/v5`, `tealeg/xlsx` for spreadsheet generation, and `unidoc/unipdf` (or `weasyprint`) for PDF document creation.
- **Justification:** Go ensures deterministic, lightning-fast FIFO tax lot calculations even for active traders with tens of thousands of fractional fills, eliminating floating-point drift through fixed-point integer/big decimal arithmetic while producing signed PDF statements in under 100ms.
- **Dependencies & Libraries:**
 - PostgreSQL 16+ for storing tax lots, realized gain events, and statement metadata.
 - AWS S3 / MinIO for storing generated, encrypted PDF tax statements.
 - Redis 7.2+ for caching pre-computed annual tax summaries.
 - Apache Kafka 3.7+ for consuming trade settlement events.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `tax_lots`, `tax_dispositions`, `capital_gains_summary`, `tax_statements`.
- **MinIO / AWS S3:** Encrypted storage for generated PDF and XLSX tax statements.
- **Kafka Topics:**
 - Subscribes: `trade.settled`, `corporate_action.dividend_distributed`, `fee.assessed`.
 - Publishes: `tax.lot.closed`, `tax.statement.generated`.
- **Hyperledger Besu (QBFT):** Queries `SettlementDvP.sol` transaction receipts to cross-verify trade execution dates and prices.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium network running QBFT consensus.
- **Immutable Timestamp & Price Verification:** When computing capital gains, the Tax Service queries the immutable transaction receipt and block timestamp on Besu for both the original buy and sell legs.
- **Audit Verification Code:** Every generated tax statement embeds a cryptographic SHA-256 verification hash and the corresponding on-chain settlement transaction hashes (`buy_tx_hash`, `sell_tx_hash`).
- **Zero PII on Chain:** All tax calculations correlate on-chain anonymous transaction hashes with off-chain encrypted user PAN and residency records inside the secure database perimeter.

## Step-by-Step Build Instructions
1. Scaffold Go project under `services/tax-service` with clean architectural layers.
2. Define Protobuf definitions in `proto/growww/tax/v1/tax_service.proto` and generate Go gRPC stubs.
3. Configure PostgreSQL schema migrations for `tax_lots`, `tax_dispositions`, and `tax_statements`.
4. Implement the Kafka event consumer ingesting `trade.settled` events to create new buy tax lots.
5. Implement the FIFO Tax Lot Matcher allocating sell order quantities against open buy lots in chronological order.
6. Build the Capital Gains Calculation Engine supporting:
 - STCG classification (holding period $< 365$ days for listed equities).
 - LTCG classification (holding period $\ge 365$ days).
 - Grandfathering rule evaluation for pre-31-Jan-2018 acquisitions.
7. Implement statutory deductions and fee accounting (STT, Stamp Duty, exchange turnover fees, platform fees).
8. Build the Excel statement generator producing ITR Schedule 112A formatted spreadsheets.
9. Build the PDF statement generator rendering official, digitally stamped Capital Gains Statements.
10. Integrate MinIO / S3 client for encrypted storage and temporary signed URL downloads.
11. Add Prometheus metrics (`tax_calculations_total`, `statement_generation_duration_ms`) and health checks.
12. Write comprehensive automated test cases validating FIFO lot depletion, partial lot fills, wash-sale handling, and leap-year holding calculations.

## Interfaces / Contracts

### Protobuf Definition (`tax_service.proto`)
```protobuf
syntax = "proto3";

package growww.tax.v1;

option go_package = "github.com/growww/services/tax-service/gen/v1;taxv1";

service TaxReportingService {
  rpc GetCapitalGainsSummary (CapitalGainsSummaryRequest) returns (CapitalGainsSummaryResponse);
  rpc GenerateTaxStatement (GenerateTaxStatementRequest) returns (GenerateTaxStatementResponse);
  rpc GetTaxStatementStatus (TaxStatementStatusRequest) returns (TaxStatementStatusResponse);
  rpc ListTaxLots (ListTaxLotsRequest) returns (ListTaxLotsResponse);
}

message CapitalGainsSummaryRequest {
  string user_id = 1;
  string financial_year = 2; // e.g., "2025-2026"
  string quarter = 3;        // Optional: Q1, Q2, Q3, Q4, or ALL
}

message CapitalGainsSummaryResponse {
  string user_id = 1;
  string financial_year = 2;
  string total_stcg_inr = 3;
  string total_ltcg_inr = 4;
  string total_dividend_income_inr = 5;
  string total_tds_deducted_inr = 6;
  string total_stt_paid_inr = 7;
  string net_realized_pnl_inr = 8;
  repeated SecurityTaxBreakdown security_breakdowns = 9;
}

message SecurityTaxBreakdown {
  string isin = 1;
  string symbol = 2;
  string stcg_inr = 3;
  string ltcg_inr = 4;
  string total_buy_value_inr = 5;
  string total_sell_value_inr = 6;
}

message GenerateTaxStatementRequest {
  string user_id = 1;
  string financial_year = 2;
  string report_format = 3; // PDF / EXCEL / JSON
}

message GenerateTaxStatementResponse {
  string statement_id = 1;
  string status = 2; // GENERATING / READY
  int64 requested_at = 3;
}

message TaxStatementStatusRequest {
  string statement_id = 1;
}

message TaxStatementStatusResponse {
  string statement_id = 1;
  string status = 2; // READY / PROCESSING / FAILED
  string download_url = 3;
  string file_sha256 = 4;
  int64 expires_at = 5;
}

message ListTaxLotsRequest {
  string user_id = 1;
  string isin = 2;
}

message ListTaxLotsResponse {
  repeated TaxLotItem open_lots = 1;
}

message TaxLotItem {
  string lot_id = 1;
  string isin = 2;
  string acquisition_date = 3;
  string original_units = 4;
  string remaining_units = 5;
  string cost_price_per_unit = 6;
  string on_chain_buy_tx_hash = 7;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE tax_lots (
    lot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    acquisition_date DATE NOT NULL,
    original_units NUMERIC(28, 18) NOT NULL,
    remaining_units NUMERIC(28, 18) NOT NULL,
    cost_per_unit_inr NUMERIC(18, 4) NOT NULL,
    total_cost_inr NUMERIC(28, 4) NOT NULL,
    buy_settlement_ref VARCHAR(128) NOT NULL,
    on_chain_buy_tx_hash VARCHAR(66) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tax_dispositions (
    disposition_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lot_id UUID NOT NULL REFERENCES tax_lots(lot_id),
    user_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    disposition_date DATE NOT NULL,
    units_sold NUMERIC(28, 18) NOT NULL,
    sell_price_per_unit_inr NUMERIC(18, 4) NOT NULL,
    gross_proceeds_inr NUMERIC(28, 4) NOT NULL,
    allocated_cost_inr NUMERIC(28, 4) NOT NULL,
    realized_gain_inr NUMERIC(28, 4) NOT NULL,
    gain_type VARCHAR(8) NOT NULL CHECK (gain_type IN ('STCG', 'LTCG')),
    holding_period_days INT NOT NULL,
    stt_paid_inr NUMERIC(18, 4) NOT NULL,
    on_chain_sell_tx_hash VARCHAR(66) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tax_statements (
    statement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(64) NOT NULL,
    financial_year VARCHAR(16) NOT NULL,
    file_format VARCHAR(8) NOT NULL,
    s3_path VARCHAR(256) NOT NULL,
    sha256_hash VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'READY',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Security & Compliance Notes
- **Income Tax Act Compliance:** Complies with Sections 111A, 112A, 115BB, and 194 of the Indian Income Tax Act.
- **DPDP Act & PII Protection:** User Permanent Account Numbers (PAN) and tax data are encrypted at rest using AES-256-GCM and masked in all application logs.
- **Signed Download URLs:** Download links for PDF/Excel tax statements are short-lived presigned S3 URLs (TTL 15 minutes) accessible only by authenticated users.
- **Tamper Resistance:** Statements include verifiable digital signatures and on-chain blockchain transaction references ensuring non-repudiation.

## Acceptance Criteria
- [ ] Go tax service compiles cleanly and passes all linting and test suites.
- [ ] FIFO lot matching correctly depletes earliest lots and handles partial disposals accurately.
- [ ] Capital gains engine categorizes holding periods $< 365$ days as STCG and $\ge 365$ days as LTCG.
- [ ] Grandfathering calculations match official CBDT test vectors for pre-2018 assets.
- [ ] Excel statements match Income Tax e-filing Schedule 112A column schemas exactly.
- [ ] Generated PDF statements render accurately with embedded digital signatures and on-chain hashes.
- [ ] Automated test suite achieves >=85% code coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 203 (Wallet Service), Prompt 208 (Settlement Service), Prompt 209 (Holdings Service), Prompt 210 (Fee Engine).
- **Subsequent / Parallel Tasks:** Prompt 216 (Regulatory Reporting Service), Prompt 601 (Web Client Statements).
