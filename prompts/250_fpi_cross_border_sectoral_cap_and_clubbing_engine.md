# 250 - Foreign Portfolio Investor (FPI) Real-Time Sectoral Cap & Clubbing Engine (Go / Redis / PostgreSQL)

## Purpose
Inbound foreign investment into Indian capital markets is regulated under SEBI (Foreign Portfolio Investors) Regulations 2019, the Foreign Exchange Management Act (FEMA), and RBI regulations. SEBI mandates that a single Foreign Portfolio Investor (FPI) or investor group cannot hold 10% or more of the total paid-up equity capital of an Indian listed company on a fully diluted basis. Furthermore, aggregate foreign investment in any Indian corporate security cannot exceed the sectoral cap (such as 24%, 49%, 74%, or 100% depending on the sector FDI policy).

To enable frictionless, high-throughput international investing on the National Blockchain Stock Exchange (NBSE) while guaranteeing strict legal and regulatory compliance, the **FPI Real-Time Sectoral Cap & Clubbing Engine** acts as an atomic pre-trade risk and post-trade compliance gate. The engine tracks real-time foreign investment headroom per ISIN, executes graph-based beneficial ownership clubbing across affiliated offshore accounts, evaluates available investment headroom before order matching, and generates automated daily reporting feeds to the NSDL/CDSL Foreign Investment Monitoring Cell.

## What You Are Building
A distributed, sub-millisecond compliance and headroom calculation service (`services/fpi-compliance-engine`) written in Go, utilizing Redis cluster state caches and PostgreSQL for historical records. Key deliverables include:
- **Real-Time Sectoral Headroom Cache:** In-memory tracking of foreign shareholding percentages per ISIN, updated atomically upon trade execution and corporate action events.
- **Single-Investor 10% & Red-Flag Limit Gate:** Pre-trade gate preventing any single FPI entity from breaching the 10% paid-up capital threshold or the 3% red-flag threshold.
- **Investor Group Clubbing Graph Resolver:** Resolves controlling beneficial ownership (> 50% voting rights or common management) across international accounts to aggregate holdings across multiple FPI registration numbers.
- **Breach Mitigation & Divestment Orchestrator:** Triggers automated 5-day divestment notification workflows or mandatory liquidation routing if corporate action rebases cause passive sectoral cap breaches.
- **Depository Regulatory Reporting Generator:** Automated batch and real-time XML/JSON telemetry dispatch to NSDL/CDSL Centralized FPI Monitoring infrastructure.
- **Fixed Platform Fee Invariant:** Seamlessly records the universal 0.00% (Zero Fee) platform fee (0.00% fee at launch; future fee parameters governed by FeeController.sol) on all foreign investor trades while maintaining FIFO tax records for cross-border tax treaty (DTAA) filings.

## Scope Boundaries
- **In Scope:**
  - Pre-order foreign investment headroom check for inbound international orders.
  - Multi-account investor group clubbing and aggregation algorithms.
  - Real-time sectoral limit decrementing and replenishment on order lifecycle events.
  - Integration with SEBI/RBI sectoral limit master data feeds.
  - Audit logging of all FPI compliance checks and regulatory notifications.
- **Out of Scope / Handled Elsewhere:**
  - Order matching and execution (handled in Prompt 205 / Prompt 246).
  - Cross-currency FX margin calculations (handled in Prompt 238 / Prompt 251).
  - KYC and AML sanctions verification (handled in Prompt 201 / Prompt 704).
  - Smart contract token transfers and DvP settlement (handled in Prompt 306 / Prompt 718).

## Technology to Use
- **Primary Language & Framework:** **Go 1.22+** utilizing `pgx/v5` for PostgreSQL and `go-redis/v9` for atomic Redis Lua scripting. Go provides low-latency concurrency and efficient memory management required for microsecond pre-trade gating.
- **In-Memory Headroom Storage:** **Redis Cluster** with atomic Lua scripts (`scripts/check_fpi_headroom.lua`) for zero-race reservation and rollback of foreign investment headroom.
- **Relational Storage:** **PostgreSQL 16+** with relational tables for FPI entity profiles, beneficial ownership hierarchies, and historical headroom snapshots.
- **Message Broker:** **Apache Kafka** for streaming trade executions, depository cap updates, and regulatory audit records.

## Backend / Infra Touchpoints
- **PostgreSQL Tables:** `fpi_investor_profiles`, `fpi_investor_groups`, `fpi_group_memberships`, `isin_sectoral_caps`, `fpi_holding_aggregations`, `fpi_compliance_audit_logs`.
- **Kafka Topics:** Consumes `order.placement.fpi.v1`, `trade.executed.v1`, `corporate_action.rebase.v1`; publishes `order.fpi_check.passed.v1`, `order.fpi_check.rejected.v1`, `fpi.headroom_threshold_alert.v1`, `fpi.regulatory_report.v1`.
- **Matching Engine (Prompt 205):** Calls FPI service via synchronous gRPC (`ValidateFpiOrderHeadroom`) prior to injecting international orders into the order book.
- **Depository Reconciliation (Prompt 215):** Synchronizes off-chain NSDL/CDSL foreign investment data feeds with local Redis headroom state.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **ERC-3643 Compliance Claims Integration:** The FPI engine verifies that foreign investor wallet addresses possess valid on-chain Identity Claims (`CLAIM_FPI_CATEGORY_1` or `CLAIM_FPI_CATEGORY_2`) issued by the NBSE Identity Verification Authority.
- **Privacy-Preserving On-Chain Verification:** On-chain DvP contracts verify country-of-origin compliance without exposing underlying investor corporate structures, using cryptographic hashes registered in `IdentityRegistry.sol`.
- **Zero PII Exposure:** Public ledger logs record strictly pseudonymous wallet addresses, token quantities, and ISIN identifiers.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go microservice `services/fpi-compliance-engine` with clean architecture layers (`domain`, `repository`, `service`, `transport`).
2. **Define Protobuf Schema:** Create `proto/growww/fpi/v1/fpi_service.proto` defining `ValidateFpiOrderHeadroom`, `RegisterFpiEntity`, `UpdateSectoralCap`, and `GetFpiHoldingSummary`.
3. **Generate Go Stubs:** Compile Protobuf schemas into Go gRPC server and client interfaces using `buf`.
4. **Design PostgreSQL Schema:** Write database migration creating `fpi_investor_profiles`, `fpi_investor_groups`, `isin_sectoral_caps`, and audit log tables.
5. **Implement Redis Headroom Manager:** Author atomic Lua script `check_and_reserve_headroom.lua` to check current foreign holding percentage plus order quantity against total issued capital and sectoral ceiling.
6. **Implement Investor Group Clubbing Logic:** Write recursive graph traversal in Go to resolve all affiliated FPI accounts belonging to the same ultimate beneficial owner (UBO) or asset manager.
7. **Implement Pre-Trade Headroom Validator:** Expose gRPC method `ValidateFpiOrderHeadroom` executing single-investor 10% test and aggregate sectoral cap check in under 200 microseconds.
8. **Implement Trade Fill Hook:** Ingest `trade.executed.v1` events to commit reserved headroom permanently and release any unfilled order headroom reservations.
9. **Implement Order Cancellation Hook:** Ingest `order.cancelled.v1` events to release reserved headroom instantly in Redis.
10. **Implement Sectoral Master Data Ingestion:** Build daily synchronizer processing NSDL/CDSL circular data files to update ISIN FDI caps, paid-up shares, and aggregate foreign limit limits.
11. **Implement Automated Red-Flag Alerting:** Trigger Kafka notification when foreign investment reaches 3% below the sectoral ceiling (warning state requiring exchange dissemination).
12. **Implement Mandatory Divestment Workflow:** Build state machine tracking passive corporate action breaches and enforcing SEBI 5-day divestment window.
13. **Integrate Fixed Fee Model:** Record 0.00% (Zero Fee) platform fee (0.00% fee at launch; future fee parameters governed by FeeController.sol) on all foreign order fills and compute DTAA tax withholding estimates.
14. **Write Unit & Integration Tests:** Write test suites verifying zero race conditions during concurrent high-volume order reservations, accurate group clubbing, and graceful fallback on Redis disconnect.

## Interfaces / Contracts

### Protobuf Service Contract (`proto/growww/fpi/v1/fpi_service.proto`)
```protobuf
syntax = "proto3";

package growww.fpi.v1;

option go_package = "github.com/growww/proto/gen/go/fpi/v1;fpiv1";

enum FpiCategory {
  FPI_CATEGORY_UNSPECIFIED = 0;
  FPI_CATEGORY_I = 1;  // Sovereign funds, central banks, pension funds
  FPI_CATEGORY_II = 2; // Regulated funds, university funds, family offices
}

enum FpiValidationResult {
  FPI_VALIDATION_RESULT_UNSPECIFIED = 0;
  FPI_VALIDATION_RESULT_APPROVED = 1;
  FPI_VALIDATION_RESULT_REJECTED_SINGLE_CAP_BREACH = 2;
  FPI_VALIDATION_RESULT_REJECTED_SECTORAL_CAP_BREACH = 3;
  FPI_VALIDATION_RESULT_REJECTED_KYC_EXPIRED = 4;
}

message ValidateFpiOrderHeadroomRequest {
  string order_id = 1;
  string fpi_entity_id = 2;
  string isin = 3;
  string order_side = 4; // BUY or SELL
  uint64 order_quantity_e6 = 5;
  uint64 limit_price_paise = 6;
}

message ValidateFpiOrderHeadroomResponse {
  FpiValidationResult result = 1;
  bool is_allowed = 2;
  uint64 current_entity_holding_e6 = 3;
  uint64 max_allowable_entity_holding_e6 = 4;
  uint64 current_sectoral_foreign_holding_e6 = 5;
  uint64 max_allowable_sectoral_holding_e6 = 6;
  uint64 total_paid_up_capital_shares_e6 = 7;
  string rejection_reason = 8;
  int64 checked_at_ns = 9;
}

message SectoralCapUpdate {
  string isin = 1;
  string company_name = 2;
  uint64 total_paid_up_shares_e6 = 3;
  uint32 sectoral_cap_bps = 4; // e.g. 7400 for 74.00%
  uint32 aggregate_fpi_limit_bps = 5;
  bool is_red_flag_active = 6;
  int64 effective_timestamp = 7;
}

service FpiComplianceService {
  rpc ValidateFpiOrderHeadroom (ValidateFpiOrderHeadroomRequest) returns (ValidateFpiOrderHeadroomResponse);
  rpc UpdateSectoralCap (SectoralCapUpdate) returns (ValidateFpiOrderHeadroomResponse);
}
```

### PostgreSQL Schema Definition (`services/fpi-compliance-engine/migrations/001_initial_schema.sql`)
```sql
CREATE TABLE fpi_investor_profiles (
    fpi_entity_id VARCHAR(64) PRIMARY KEY,
    fpi_registration_number VARCHAR(64) NOT NULL UNIQUE,
    entity_name VARCHAR(255) NOT NULL,
    fpi_category VARCHAR(16) NOT NULL,
    country_of_incorporation VARCHAR(3) NOT NULL,
    group_id VARCHAR(64) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE isin_sectoral_caps (
    isin VARCHAR(12) PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL,
    total_paid_up_shares NUMERIC(24, 6) NOT NULL,
    sectoral_cap_percentage NUMERIC(5, 2) NOT NULL,
    aggregate_fpi_limit_percentage NUMERIC(5, 2) NOT NULL,
    current_foreign_holding_shares NUMERIC(24, 6) NOT NULL DEFAULT 0,
    is_in_red_flag_zone BOOLEAN NOT NULL DEFAULT FALSE,
    last_depository_sync_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fpi_compliance_audit_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id VARCHAR(64) NOT NULL,
    fpi_entity_id VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    order_side VARCHAR(4) NOT NULL,
    quantity NUMERIC(24, 6) NOT NULL,
    validation_status VARCHAR(32) NOT NULL,
    rejection_code VARCHAR(32),
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fpi_audit_entity ON fpi_compliance_audit_logs(fpi_entity_id, isin);
CREATE INDEX idx_fpi_profiles_group ON fpi_investor_profiles(group_id);
```

## Security & Compliance Notes
- **FEMA & SEBI FPI Regulations 2019 Adherence:** All inbound foreign orders strictly require atomic headroom checks before matching engine order injection.
- **Race Condition Immunity:** Redis Lua atomic scripts prevent simultaneous orders from over-subscribing available foreign investment headroom.
- **Zero PII Exposure:** On-chain identities use pseudonymous ERC-3643 claim hashes verified against confidential off-chain vault registries.
- **DPDP Act 2023 & GDPR Compliance:** Personal data of offshore beneficial owners is encrypted at rest using AES-256-GCM with KMS-managed keys.

## Acceptance Criteria
- [ ] gRPC service processes `ValidateFpiOrderHeadroom` with latency < 200 microseconds at p99 under 20,000 req/s.
- [ ] Atomic Redis reservation prevents headroom breaches during concurrent execution spikes.
- [ ] Single FPI holdings strictly blocked from exceeding 9.999% of total company paid-up shares.
- [ ] Aggregate foreign holdings strictly blocked from exceeding designated ISIN sectoral cap.
- [ ] Group clubbing algorithm correctly aggregates holdings across multi-entity corporate structures.
- [ ] Universal Zero-Fee Model (0.00% fee - No fee at all) (0.00% fee at launch (future fee parameters governed by FeeController.sol)) logged accurately for foreign trade executions.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 102 (Bounded Contexts), Prompt 201 (Auth & KYC).
- **Parallel Work:** Prompt 205 (Matching Engine), Prompt 238 (Cross-Chain Router), Prompt 251 (FX Haircut Engine).
- **Enables:** Inbound international institutional trading and seamless foreign retail participation into Indian equities.
