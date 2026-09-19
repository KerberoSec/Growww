# 203 - Wallet & Double-Entry Account Ledger Service (Go)

## Purpose
The Wallet & Account Ledger Service is the financial core of Growww's domestic operations. It manages Indian Rupee (INR) fiat balances, automated deposits, withdrawal disbursements, and funds reservation (holds) for active buy orders. In strict adherence to SEBI and RBI regulations, all domestic funds are denominated in INR and held in designated nodal escrow bank accounts with RBI-licensed scheduled commercial banks. Growww strictly avoids unregulated stablecoins, crypto on-ramps, or synthetic balances.

The service implements an immutable, cryptographic double-entry accounting ledger where every financial movement is represented by balanced debits and credits ($\sum \text{debits} = \sum \text{credits}$). This guarantees mathematical consistency, auditability, and zero balance drift across millions of transactions.

## What You Are Building
A mission-critical, high-throughput Go microservice (`services/wallet-service`). Concrete deliverables include:
- High-performance gRPC server exposing RPCs for account creation, balance checks, fund holds (reservation), hold releases, and trade settlement commits.
- Strict double-entry ledger database schema (`accounts`, `journal_entries`, `postings`, `funds_holds`) with PostgreSQL serializable isolation.
- Fast, atomic fund reservation engine utilizing pessimistic row-level locking (`SELECT ... FOR UPDATE`) to prevent double-spending or negative balances.
- Kafka consumers for payment gateway events (`payment.deposit.confirmed.v1`) and trade settlement events (`trade.settled.v1`).
- Kafka publisher streaming balance update events (`wallet.balance_updated.v1`, `wallet.hold_created.v1`).

## Scope Boundaries
- **In Scope:**
 - Account chart definition (Asset, Liability, Equity, Revenue, Expense).
 - Real-time balance calculations: $\text{Available Balance} = \text{Ledger Balance} - \text{Total Active Holds}$.
 - Pre-trade fund reservation (`ReserveFunds`) for limit/market buy orders.
 - Hold lifecycle state transitions (`ACTIVE`, `COMMITTED`, `RELEASED`, `EXPIRED`).
 - Idempotent deposit credit and withdrawal debit journal postings.
- **Out of Scope / Handled Elsewhere:**
 - Payment gateway banking integrations (UPI, IMPS, NEFT, RTGS) (Prompt 212).
 - Order matching book and execution logic (Prompt 205).
 - Delivery-versus-Payment (DvP) smart contract settlement coordination (Prompt 208).
 - Capital gains fee calculation engine (Prompt 210).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+. Go is selected for its high concurrency primitives (goroutines/channels), garbage collection predictability, minimal memory footprint, and deterministic low-latency execution under extreme financial transaction loads.
- **Database & Driver:** PostgreSQL 16+ utilizing `pgx/v5` connection pool configured with strict transactional isolation (`READ COMMITTED` with explicit row locking or `SERIALIZABLE` for reconciliation sweeps).
- **SQL Code Generation:** `sqlc` for compile-time verified, zero-reflection type-safe SQL queries.
- **Financial Arithmetic:** `github.com/shopspring/decimal` for fixed-point arbitrary precision arithmetic (avoiding floating-point IEEE-754 precision loss).
- **Inter-Service Communication:** `google.golang.org/grpc` for sub-millisecond internal RPCs; `segmentio/kafka-go` for reliable event consumption with manual offset commits.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `accounts`, `journal_entries`, `postings`, `funds_holds`, `idempotency_keys`.
- **Redis 7.2:** Caches user available balances with strict short TTLs (10s) and immediate cache eviction on journal commits.
- **Apache Kafka:** Subscribes to `payment.events.v1`, `trade.settlement.v1`; publishes to `wallet.events.v1`.
- **Order Service (Prompt 204):** Calls `ReserveFunds` and `ReleaseFunds` via gRPC.
- **Trade Settlement Service (Prompt 208):** Calls `CommitHold` via gRPC during atomic DvP execution.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Fiat-to-Ledger Mapping:** All cash balances in this service represent real Indian Rupees held in RBI-regulated nodal escrow bank accounts.
- **DvP Atomic Settlement Link:** During trade settlement, this service coordinates with `SettlementDvP.sol` on Hyperledger Besu via the Trade Settlement Service (Prompt 208). When a buy order is matched, off-chain INR cash is debited from the buyer's account and credited to the seller's account simultaneously as digital equity tokens are transferred on Hyperledger Besu under QBFT consensus.
- **Zero Stablecoins / Zero On-Chain Fiat:** No fiat currency is converted into stablecoins or stored as ERC-20 tokens. The blockchain maintains tokenized equity custody records (ERC-3643) while fiat settlements are cryptographically attested via settlement transaction hashes (`tx_hash`).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize `services/wallet-service` Go module with `golangci-lint` configuration and standard project layout (`cmd/`, `internal/domain/`, `internal/repository/`, `internal/service/`).
2. **Define Protobuf Contracts:** Create `proto/growww/wallet/v1/wallet_service.proto` for balance queries, fund holds, and postings.
3. **Generate Go Protobuf & gRPC Code:** Compile proto definitions using `protoc-gen-go` and `protoc-gen-go-grpc`.
4. **Design PostgreSQL Schema:** Write SQL migration files defining accounts, journal entries, postings, and holds. Enforce check constraints (`balance >= 0` for customer asset accounts).
5. **Configure `sqlc`:** Set up `sqlc.yaml` and generate type-safe Go queries for transactional atomic ledger operations.
6. **Implement Double-Entry Ledger Engine:** Write domain logic ensuring every journal entry validates $\sum \text{debits} - \sum \text{credits} = 0$ before execution.
7. **Implement Pessimistic Fund Reservation (`ReserveFunds`):** Build locking mechanism: select account row `FOR UPDATE`, check `available_balance >= requested_amount`, create hold record, and return unique `hold_id`.
8. **Implement Hold Commit & Release Logic:** Build functions to atomically commit holds into journal postings (upon trade fill) or release holds back to available balance (upon order cancellation).
9. **Implement Idempotency Engine:** Store idempotency keys in PostgreSQL within the same database transaction to prevent duplicate debits or credits.
10. **Build gRPC Service Handlers:** Implement `WalletServiceServer` interface with structured error handling (e.g., `INSUFFICIENT_FUNDS`, `HOLD_NOT_FOUND`).
11. **Implement Kafka Event Consumers:** Build consumer groups for deposit confirmations and DvP trade settlements with retry backoff and dead-letter queues.
12. **Implement Kafka Event Publisher:** Stream balance update notifications on `wallet.events.v1`.
13. **Configure Telemetry & Health Probes:** Expose `/metrics` (Prometheus) with histograms for transaction latency and `/healthz` liveness probes.
14. **Perform High-Concurrency Integration Testing:** Write unit tests and concurrent stress tests using Go's `testing` and `testcontainers-go` verifying zero race conditions and zero negative balances under 500 concurrent threads.

## Interfaces / Contracts

### Protobuf Definition (`wallet_service.proto`)
```protobuf
syntax = "proto3";

package growww.wallet.v1;

option go_package = "growww/wallet/v1;walletv1";

service WalletService {
  rpc GetBalance (GetBalanceRequest) returns (GetBalanceResponse);
  rpc ReserveFunds (ReserveFundsRequest) returns (ReserveFundsResponse);
  rpc ReleaseFunds (ReleaseFundsRequest) returns (ReleaseFundsResponse);
  rpc CommitHold (CommitHoldRequest) returns (CommitHoldResponse);
  rpc CreditDeposit (CreditDepositRequest) returns (CreditDepositResponse);
}

message GetBalanceRequest {
  string user_id = 1;
}

message GetBalanceResponse {
  string user_id = 1;
  string currency = 2; // Always "INR"
  string ledger_balance = 3; // String representation of decimal (e.g., "15000.50")
  string held_balance = 4;
  string available_balance = 5;
  int64 updated_at_unix = 6;
}

message ReserveFundsRequest {
  string idempotency_key = 1;
  string user_id = 2;
  string order_id = 3;
  string amount = 4; // Decimal string in INR
  string purpose = 5; // e.g., "BUY_ORDER_RESERVATION"
}

message ReserveFundsResponse {
  string hold_id = 1;
  string user_id = 2;
  string held_amount = 3;
  string remaining_available_balance = 4;
  int64 expires_at_unix = 5;
}

message ReleaseFundsRequest {
  string idempotency_key = 1;
  string hold_id = 2;
  string reason = 3;
}

message ReleaseFundsResponse {
  bool success = 1;
  string released_amount = 2;
  string updated_available_balance = 3;
}

message CommitHoldRequest {
  string idempotency_key = 1;
  string hold_id = 2;
  string trade_id = 3;
  string debit_amount = 4;
  string counterparty_account_id = 5;
}

message CommitHoldResponse {
  string journal_entry_id = 1;
  bool success = 2;
  string final_balance = 3;
}

message CreditDepositRequest {
  string idempotency_key = 1;
  string user_id = 2;
  string payment_reference = 3;
  string amount = 4;
}

message CreditDepositResponse {
  string journal_entry_id = 1;
  string updated_ledger_balance = 2;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE account_type_enum AS ENUM ('CUSTOMER_INR_WALLET', 'SETTLEMENT_CLEARING', 'NODAL_ESCROW_BANK', 'PLATFORM_REVENUE', 'STATUTORY_TAX_PAYABLE');
CREATE TYPE hold_status_enum AS ENUM ('ACTIVE', 'COMMITTED', 'RELEASED', 'EXPIRED');

CREATE TABLE accounts (
    account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID, -- NULL for platform internal clearing accounts
    account_type account_type_enum NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    ledger_balance NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    held_balance NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_positive_customer_balance CHECK (
        account_type != 'CUSTOMER_INR_WALLET' OR (ledger_balance >= held_balance AND held_balance >= 0)
    )
);

CREATE TABLE journal_entries (
    entry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_reference VARCHAR(100) NOT NULL UNIQUE,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE postings (
    posting_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES journal_entries(entry_id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(account_id),
    amount NUMERIC(18, 4) NOT NULL, -- Positive for Credit, Negative for Debit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE funds_holds (
    hold_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(account_id),
    order_id UUID NOT NULL,
    held_amount NUMERIC(18, 4) NOT NULL CHECK (held_amount > 0),
    status hold_status_enum NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_accounts_user ON accounts(user_id);
CREATE INDEX idx_holds_account_active ON funds_holds(account_id) WHERE status = 'ACTIVE';
```

## Security & Compliance Notes
- **Mathematical Invariant Verification:** Every ledger transaction enforces $\sum \text{postings.amount} = 0$ via database trigger and application-level invariant checks.
- **Pessimistic Concurrency:** Uses PostgreSQL row-level locks on `accounts` to prevent balance race conditions and double-spending across distributed microservices.
- **Nodal Account Segregation:** Domestic client funds are strictly segregated from operational company capital in accordance with RBI Nodal Escrow Account regulations.
- **Audit Logging:** Every journal entry and posting is immutable and append-only; update/delete operations on `postings` are blocked via SQL permission grants.

## Acceptance Criteria
- [ ] Double-entry ledger enforces balance equality across all debit and credit postings.
- [ ] `ReserveFunds` successfully locks funds and returns `INSUFFICIENT_FUNDS` when requested amount exceeds available balance.
- [ ] Concurrent requests against the same user account execute safely without negative balances or race conditions.
- [ ] Releasing or committing holds updates available and ledger balances accurately in a single database transaction.
- [ ] Idempotency keys prevent duplicate deposits, holds, or trade debit executions.
- [ ] gRPC response latency for `GetBalance` and `ReserveFunds` is $< 3\text{ms}$ at 2,000 req/sec.

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Standards), 111 (Domain Model), 112 (Idempotency), 401 (PostgreSQL Schema).
- **Parallel Tasks:** 201 (User Service), 202 (KYC Service), 206 (Risk Engine).
- **Downstream Blockers:** 204 (Order Service), 208 (Trade Settlement Service), 212 (Payment Gateway Service).
