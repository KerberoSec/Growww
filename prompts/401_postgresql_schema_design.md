# 401 - PostgreSQL Schema Design (Core Transactional Data)

## Purpose
The PostgreSQL database serves as the primary relational and transactional source of truth for the Growww ecosystem. It enforces strict ACID guarantees across all core financial operations, including user identity metadata, INR fiat balances, order life cycles, executed trades, fractional security unit holdings, NSDL/CDSL custody share allocations, and transaction fee tracking. 

In a regulatory environment governed by SEBI, RBI, and IFSCA, the database architecture must guarantee absolute data integrity, deterministic double-entry accounting, auditability, and zero data loss. It bridges traditional Indian financial market clearing with the permissioned Hyperledger Besu blockchain, storing off-chain metadata, KYC verifications, and physical-to-digital asset custody mappings while strictly isolating Personally Identifiable Information (PII) from on-chain transactions.

## What You Are Building
A production-grade PostgreSQL 16+ relational schema definition and database migration pipeline containing:
- **Migration Suite:** Flyway/Alembic migration scripts (`migrations/postgres/`) organized by bounded context schemas (`identity`, `ledger`, `trading`, `custody`, `compliance`, `fees`).
- **Relational Tables & Constraints:** Full DDL for users, bank accounts, fiat ledgers, order book state, trade executions, fractional token holdings, custodian safe-keeping receipts, and fee settlements.
- **Partitioning Strategy:** Declarative range partitioning on high-volume tables (`trading.trades`, `ledger.journal_entries`, `audit.system_events`) by time range (monthly/daily) using `pg_partman`.
- **High-Performance Indexing:** Optimized B-Tree, partial, BRIN, and composite indexes tailored for sub-millisecond lookups and audit queries.
- **Double-Entry Financial Invariants:** Immutable append-only ledger journal with strict CHECK constraints, preventing negative balances and balance discrepancies.
- **Multi-Tenant Row-Level Security (RLS):** Database-level security policies enforcing segregation between Domestic Regulated Entity data and GIFT City Gateway International Entity data.

## Scope Boundaries
- **In Scope:** Complete DDL definitions, table constraints, declarative partitioning, indexing strategies, audit triggers, foreign key integrity rules, Flyway/Alembic migration harnesses, and multi-tenant RLS policies.
- **Out of Scope / Handled Elsewhere:**
 - Application-level ORM/repository implementations (handled in Category 2 microservices: 201, 203, 204, 208, 209).
 - In-memory caching and real-time order-book state (handled in Prompt 402).
 - Analytical OLAP data warehousing and ClickHouse schemas (handled in Prompt 404).
 - Kafka event bus streaming and CDC integration (handled in Prompt 403).

## Technology to Use
PostgreSQL 16+ is selected as the primary relational engine. PostgreSQL provides industry-standard ACID transactional guarantees, native declarative table partitioning with query partition pruning, native JSONB support for semi-structured regulatory payloads, robust Row-Level Security (RLS), and unmatched ecosystem maturity for financial systems. In comparison to MySQL or distributed NoSQL stores, PostgreSQL provides stronger transactional isolation guarantees, advanced BRIN indexing for high-volume time-series data, and native integration with `pg_partman` and `pgcrypto`.

- **Database Engine:** PostgreSQL 16.x Enterprise.
- **Migration Framework:** Flyway 10.x / Alembic 1.13+.
- **Partitioning Extension:** `pg_partman` v5.x.
- **Connection Pooling:** PgBouncer 1.22+ in transaction-pooling mode.
- **Crypto & UUID Extensions:** `pgcrypto`, `uuid-ossp`.

## Backend / Infra Touchpoints
- **Primary-Replica Cluster:** Highly Available PostgreSQL cluster across 3 Availability Zones with synchronous streaming replication to an in-region standby and asynchronous replication to a DR replica.
- **PgBouncer:** Layered connection pooler terminating client connections and managing server pool limits.
- **HashiCorp Vault / AWS KMS:** Dynamic database credential generation and column-level Transparent Data Encryption (TDE).
- **Kafka / Debezium CDC:** Change Data Capture streaming table mutations to the Kafka messaging bus.

## Blockchain Interaction
The PostgreSQL schema stores the canonical off-chain ledger that reconciles 1:1 with the permissioned Hyperledger Besu blockchain network (QBFT consensus, 2-second block finality).

### Detailed On-Chain Integration Mechanics:
- **Contract Reflections:** Stores transaction references for smart contracts including `DigitalSecurityToken.sol` (token balances and fractional units), `SettlementDvP.sol` (atomic trade settlements), `ComplianceRegistry.sol` (whitelisted investor addresses), and `ProofOfReserveRegistry.sol` (Merkle roots of physical share custody).
- **Zero-PII Ledger Mapping:** Maps internal `user_id` and `kyc_id` to pseudonymous Ethereum addresses (`0x...`). No names, PAN numbers, Aadhaar numbers, or bank account details are ever sent to or stored in on-chain events.
- **Custody 1:1 Reconciliation:** The `custody.allocations` table tracks physical demat account holdings in NSDL/CDSL against the total minted fractional tokens recorded in `DigitalSecurityToken.sol`.
- **Proof-of-Reserve Attestation:** Daily Merkle root calculations of off-chain share allocations are stored in PostgreSQL before being committed to the on-chain `ProofOfReserveRegistry.sol`.
- **Audit Hashes:** Every settled trade record stores `tx_hash`, `block_number`, and `block_timestamp` from the Besu ledger.

## Step-by-Step Build Instructions
1. Scaffold migration directory structure: `migrations/postgres/V1__init_schemas.sql` to `V8__indexes_and_rls.sql`.
2. Configure schema namespaces: `CREATE SCHEMA identity`, `ledger`, `trading`, `custody`, `compliance`, `fees`, `audit`.
3. Enable required PostgreSQL extensions (`uuid-ossp`, `pgcrypto`, `pg_stat_statements`).
4. Implement `identity` schema DDL: `users`, `kyc_records`, `bank_accounts`, `entity_affiliations` (Domestic vs GIFT City).
5. Implement `ledger` schema DDL: immutable double-entry tables `accounts`, `journal_entries`, `postings` with `NUMERIC(18, 4)` currency precision.
6. Implement `trading` schema DDL: `orders`, `order_state_transitions`, and range-partitioned `trades` partitioned monthly by `executed_at`.
7. Implement `custody` schema DDL: `depository_participants`, `demat_accounts`, `physical_custody_holdings`, `fractional_token_allocations`.
8. Implement `fees` schema DDL: `transaction_fees`, `treasury_revenue_allocations` (0.00% fee at launch (governed by FeeController.sol) split on 0.00% (Zero Fee) trade turnover), and `tax_lot_disposals` using FIFO tracking strictly for user tax compliance (Section 111A/112A).
9. Implement `compliance` schema DDL: `investor_whitelists`, `regulatory_freezes`, `sanctions_screening_logs`.
10. Configure `pg_partman` automated partition maintenance routines for `trades` and `journal_entries`.
11. Build composite and partial indexes for hot query paths (`orders.user_id WHERE status = 'ACTIVE'`).
12. Define Row-Level Security (RLS) policies on entity-specific tables to enforce legal boundary separation between domestic and international tenants.
13. Implement automated SQL migration test suite verifying rollback safety, schema linters, and foreign key cycle prevention.

## Interfaces / Contracts

```sql
-- Core Schema DDL Specification (Sample Bounded Contexts)

CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS ledger;
CREATE SCHEMA IF NOT EXISTS trading;
CREATE SCHEMA IF NOT EXISTS custody;
CREATE SCHEMA IF NOT EXISTS fees;

-- Identity: Investor Account
CREATE TABLE identity.users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(16) NOT NULL CHECK (entity_type IN ('DOMESTIC_RE', 'GIFT_CITY_GW')),
    blockchain_address CHAR(42) NOT NULL UNIQUE, -- Zero-PII pseudonymous address (0x...)
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING_KYC' CHECK (status IN ('PENDING_KYC', 'ACTIVE', 'FROZEN', 'CLOSED')),
    kyc_level VARCHAR(16) NOT NULL DEFAULT 'NONE' CHECK (kyc_level IN ('NONE', 'BASIC', 'FULL_SEBI', 'FULL_IFSCA')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Ledger: Immutable Double-Entry Journal
CREATE TABLE ledger.accounts (
    account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identity.users(user_id) ON DELETE RESTRICT,
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    account_type VARCHAR(32) NOT NULL CHECK (account_type IN ('USER_WALLET', 'SETTLEMENT_ESCROW', 'FEE_REVENUE', 'CUSTODY_RESERVE')),
    balance NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (balance >= 0),
    locked_balance NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (locked_balance >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE ledger.journal_entries (
    entry_id UUID DEFAULT gen_random_uuid(),
    reference_id UUID NOT NULL, -- Order ID, Deposit ID, or Settlement ID
    reference_type VARCHAR(32) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (entry_id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE ledger.postings (
    posting_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL,
    entry_created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    account_id UUID REFERENCES ledger.accounts(account_id) ON DELETE RESTRICT,
    amount NUMERIC(18, 4) NOT NULL CHECK (amount <> 0), -- Positive: Credit, Negative: Debit
    balance_after NUMERIC(18, 4) NOT NULL CHECK (balance_after >= 0)
);

-- Trading: Partitioned Trade Execution Table
CREATE TABLE trading.trades (
    trade_id UUID DEFAULT gen_random_uuid(),
    buyer_order_id UUID NOT NULL,
    seller_order_id UUID NOT NULL,
    isin CHAR(12) NOT NULL,
    symbol VARCHAR(32) NOT NULL,
    token_address CHAR(42) NOT NULL,
    fractional_units NUMERIC(18, 6) NOT NULL CHECK (fractional_units > 0),
    price_per_unit NUMERIC(18, 4) NOT NULL CHECK (price_per_unit > 0),
    gross_amount NUMERIC(18, 4) NOT NULL,
    fee_amount NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    settlement_status VARCHAR(32) NOT NULL DEFAULT 'PENDING' CHECK (settlement_status IN ('PENDING', 'SETTLED_DVP', 'FAILED')),
    on_chain_tx_hash CHAR(66), -- 0x... Besu transaction hash
    on_chain_block_num BIGINT,
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    PRIMARY KEY (trade_id, executed_at)
) PARTITION BY RANGE (executed_at);

-- Custody: 1:1 Physical Demat Backing Allocation
CREATE TABLE custody.allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin CHAR(12) NOT NULL,
    depository VARCHAR(16) NOT NULL CHECK (depository IN ('NSDL', 'CDSL')),
    demat_account_no VARCHAR(32) NOT NULL,
    physical_shares_held NUMERIC(18, 6) NOT NULL CHECK (physical_shares_held >= 0),
    tokens_minted NUMERIC(18, 6) NOT NULL CHECK (tokens_minted >= 0),
    reserve_status VARCHAR(16) NOT NULL DEFAULT 'BALANCED' CHECK (reserve_status IN ('BALANCED', 'DISCREPANCY', 'REBALANCING')),
    last_reconciled_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    CONSTRAINT chk_custody_1_to_1 CHECK (tokens_minted <= physical_shares_held)
);

-- Fees: Fixed Transaction Fee & Treasury Allocation Ledger (0.00% (Zero Fee) on Turnover) & FIFO Tax Tracking
CREATE TABLE fees.transaction_fees (
    fee_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identity.users(user_id),
    trade_id UUID NOT NULL,
    isin CHAR(12) NOT NULL,
    turnover_amount NUMERIC(18, 4) NOT NULL,
    fee_rate NUMERIC(6, 6) NOT NULL DEFAULT 0.000000, -- 0.00% (Zero Fee) (0 bps at launch; FeeController governed)
    total_fee_inr NUMERIC(18, 4) NOT NULL,
    treasury_portion_inr NUMERIC(18, 4) NOT NULL, -- Governed by FeeController (0.00% at launch)
    core_sgf_portion_inr NUMERIC(18, 4) NOT NULL, -- Governed by FeeController (0.00% at launch)
    ipf_portion_inr NUMERIC(18, 4) NOT NULL,      -- Governed by FeeController (0.00% at launch)
    assessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

CREATE TABLE fees.tax_lot_disposals (
    disposal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES identity.users(user_id),
    trade_id UUID NOT NULL,
    isin CHAR(12) NOT NULL,
    cost_basis NUMERIC(18, 4) NOT NULL,
    sale_proceeds NUMERIC(18, 4) NOT NULL,
    realized_capital_gain NUMERIC(18, 4) NOT NULL,
    holding_period_days INTEGER NOT NULL,
    gain_type VARCHAR(8) NOT NULL CHECK (gain_type IN ('STCG', 'LTCG')), -- Section 111A / 112A compliance
    assessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);
```

## Security & Compliance Notes
- **Data Localization:** All production database instances for the domestic entity are hosted exclusively in Indian data centers (e.g. AWS ap-south-1 Mumbai/Hyderabad) compliant with RBI and SEBI mandates.
- **DPDP Act & PII Isolation:** All sensitive investor PII (Aadhaar, PAN, Bank account numbers) is stored encrypted with KMS AES-256 in dedicated isolated vaults. The core trading and ledger tables reference only pseudonymous UUIDs.
- **Immutable Financial Records:** SEBI mandates 8-year audit retention. Row updates or deletions on `ledger.journal_entries`, `ledger.postings`, and `trading.trades` are strictly blocked via PostgreSQL database triggers.
- **Precision Guarantees:** Floating-point data types (`FLOAT`, `DOUBLE`) are strictly forbidden. All monetary balances use `NUMERIC(18, 4)` and all security fractional quantities use `NUMERIC(18, 6)`.

## Acceptance Criteria
- [ ] Complete DDL migration scripts authored and validated using Flyway/Alembic in clean PostgreSQL 16 instances.
- [ ] Table schemas across `identity`, `ledger`, `trading`, `custody`, `compliance`, `fees` initialized with zero foreign key cycle bugs.
- [ ] Declarative monthly partitioning verified on `trading.trades` and `ledger.journal_entries` with automatic partition creation via `pg_partman`.
- [ ] Strict CHECK constraints enforce double-entry invariants and prevent negative balances or fractional token over-issuance (`tokens_minted <= physical_shares_held`).
- [ ] Row-Level Security (RLS) policies tested to isolate domestic records from GIFT City international records.
- [ ] Benchmark tests verify >5,000 writes/sec on `trades` and `postings` with PgBouncer connection pooling.
- [ ] Database triggers successfully block `UPDATE` and `DELETE` on immutable financial tables.
- [ ] Audit logging extension and pg_stat_statements enabled for all DDL and DML operations.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 101 (System Architecture), Prompt 102 (Bounded Contexts), Prompt 111 (Domain Model).
- **Parallel Tasks:** Prompt 402 (Redis Caching), Prompt 403 (Kafka Architecture), Prompt 407 (Master Data Management).
