# 274 - Demo Trading Faucet & Virtual Balance Ledger Service (Go)

## Purpose
Retail financial platforms and algorithmic trading venues must provide newcomers, retail investors, and quantitative developers with a zero-risk, high-fidelity sandbox environment. Novice traders need to experience market depth, order submission lifecycles, and portfolio dynamics without endangering real capital, while algorithmic trading developers require a deterministic testing ground that accurately mirrors the behavior of production execution venues.

As specified in the platform trading architecture, the **Demo Trading Faucet & Virtual Balance Ledger Service** (`services/demo-wallet-service`) governs virtual account balances, testnet faucet allocations (provisioning an initial baseline of 10,000 virtual USDT and 1.00000000 virtual BTC per user), and immutable double-entry balance accounting for risk-free paper trading.

The service guarantees absolute logical and physical isolation between simulated paper balances and real-world capital. It operates its own dedicated relational datastore (`nbse_demo_wallet`), isolated event streams, and dedicated testnet contract relayers. It enforces strict anti-abuse rate limits (faucet claims restricted to once per 24-hour window per user account and client IP subnet), provides instantaneous portfolio reset capabilities to restore baseline allocations, and acts as the authoritative financial ledger for the Demo Matching Engine.

---

## What You Are Building
A high-performance, low-latency Go microservice (`services/demo-wallet-service`) developed in Go 1.22+ that functions as the authoritative virtual balance management and paper-trading custody engine. Concrete deliverables include:

- **1-Click Testnet Faucet Allocation Engine:** An idempotent faucet claim handler that disburses 10,000 virtual USDT (vUSDT) and 1 virtual BTC (vBTC) to verified user sandbox accounts. It validates eligibility, executes double-entry credits, and triggers asynchronous testnet ERC-20 minting on Hyperledger Besu Testnet.
- **Strict Anti-Abuse Rate-Limiting Guard:** A Redis-backed token-bucket and sliding-window rate limiter enforcing a strict 24-hour cooldown per user account and per client IP address, coupled with device fingerprint checks to prevent testnet sybil attacks and ledger bloat.
- **Immutable Virtual Double-Entry Ledger:** A multi-asset accounting core adhering to strict double-entry principles (total debits equal total credits). Every faucet grant, order hold, trade execution, and portfolio reset produces immutable journal entries with running balance assertions, preventing negative balances and phantom credits.
- **Atomic Portfolio Reset Mechanism:** A 1-click portfolio reset endpoint that cancels active virtual order holds, archives current paper trading positions, and re-initializes user balances to the canonical starting allocation (10,000 vUSDT and 1 vBTC), providing traders a clean slate.
- **High-Throughput Order Reservation & Hold Manager:** Low-latency gRPC APIs (`HoldVirtualBalance`, `ReleaseVirtualBalance`, `SettleVirtualTrade`) invoked by the Demo Matching Engine (Prompt 273) to lock required margin, release funds upon cancellation, and execute bilateral atomic trade settlements.
- **Asynchronous Testnet Blockchain Relayer:** An outbox-pattern event processor that connects to the Testnet Faucet Smart Contract (Prompt 348) on Hyperledger Besu to mint corresponding virtual ERC-20 tokens to the user's non-custodial testnet address.

---

## Scope Boundaries

### In Scope
- Provisioning and managing virtual sandbox accounts for registered users across multiple virtual assets: vUSDT, vBTC, vETH, and vINR.
- Standard 1-click testnet faucet disbursement granting 10,000 vUSDT and 1 vBTC per successful claim.
- Anti-abuse verification: 24-hour cooldown per user ID, IP subnet throttling (maximum 5 claims per hour per `/24` IPv4 or `/48` IPv6 block), and Cloudflare turnstile / captcha verification hooks.
- 1-click portfolio reset restoring virtual balances to baseline allocations while releasing active open order holds.
- Ultra-low latency hold, release, and bilateral settlement operations for the Demo Matching Engine (Prompt 273).
- Strict double-entry accounting in isolated PostgreSQL database (`nbse_demo_wallet`) with debit/credit balance verification.
- Publishing virtual balance change and faucet claim events to isolated Kafka topics (`demo.wallet.*`).
- Asynchronous invocation of Testnet Faucet Smart Contract (Prompt 348) on Hyperledger Besu Testnet via automated relayer keys.
- Querying virtual balances, available margin, open holds, and ledger audit history.

### Out of Scope / Handled Elsewhere
- Continuous Limit Order Book matching and execution for demo instruments (handled by Demo Matching Engine, Prompt 273).
- Real-money fiat deposits, UPI payment gateways, bank transfers, and custody accounts (handled by Wallet & Account Service, Prompt 203, and Fiat Banking, Prompt 212).
- Production real-asset order matching and institutional risk checks (handled by Matching Engine, Prompt 205, and Risk Service, Prompt 206).
- Production Mainnet blockchain transactions, real ERC-3643 tokens, and DvP settlement (handled by Prompt 303 and Prompt 306).
- Primary user identity creation and real-world KYC PAN/Aadhaar verification (handled by User Service, Prompt 201, and KYC Service, Prompt 202).
- Frontend mobile UI state and sandbox banners (handled by Flutter Environment Switcher, Prompt 527, and Web Dashboard, Prompt 603).

---

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+ utilizing standard library concurrency, `context` propagation, and Go workspaces.
- **Relational Datastore:** PostgreSQL 16+ running in a physically separate database instance (`nbse_demo_wallet`) with `pgx/v5` connection pooling and `sqlc` for compile-time verified, zero-allocation SQL queries.
- **In-Memory Cache & Rate Limiter:** Redis 7.2+ Cluster utilizing Redis Lua scripts for atomic 24-hour sliding-window checks and `redsync` for distributed lock coordination during balance reservations.
- **Event Streaming Broker:** Apache Kafka 3.7+ (`segmentio/kafka-go`) consuming matching events and publishing virtual ledger mutations on dedicated `demo.wallet.*` topics.
- **RPC & Web Gateway Framework:** gRPC (`google.golang.org/grpc`) with HTTP/2 protobuf transport, coupled with `grpc-gateway` to expose RESTful JSON endpoints.
- **High-Precision Arithmetic:** `github.com/shopspring/decimal` for fixed-point math (up to 28 digits of precision, 8 to 18 decimal places) to eliminate IEEE-754 floating-point drift.
- **Blockchain Client:** `go-ethereum/ethclient` communicating with Hyperledger Besu Testnet JSON-RPC endpoints via EIP-1559 transactions with automated nonce and gas management.
- **Observability:** OpenTelemetry Go SDK (`go.opentelemetry.io/otel`) with Prometheus metrics exporter (`prometheus/client_golang`) and structured logging via `uber-go/zap`.

---

## Backend / Infra Touchpoints
- **User Service (Prompt 201):** Validates user authentication tokens and extracts user identifiers (`user_id`, account status). Automatically initializes virtual wallet records upon initial user login or onboarding trigger.
- **Demo Matching Engine (Prompt 273):** Primary downstream caller of balance hold, release, and settlement RPCs. Invokes `HoldVirtualBalance` prior to admitting limit orders into the virtual order book, and calls `SettleVirtualTrade` upon matching counterparty orders.
- **Testnet Faucet Smart Contract (Prompt 348):** Deployed on Hyperledger Besu Testnet. Invoked asynchronously by the relayer worker to mint testnet ERC-20 tokens (vUSDT, vBTC) to the user's testnet wallet address.
- **Demo Market Data Service:** Consumes live or mirrored price feeds to evaluate total virtual portfolio equity and unrealized PnL.
- **Audit & Compliance Logging (Prompt 218):** Logs all administrative balance overrides, faucet rate limit violations, and portfolio resets for sandbox operational integrity.
- **Redis Cluster:** Stores user faucet claim timestamps, IP rate limit counters, and cached virtual balance snapshots for low-latency pre-check validation.
- **Apache Kafka:** Transports event payloads between the Demo Matching Engine, Demo Wallet Service, and testnet blockchain indexing pipelines.

---

## Blockchain Interaction (Permissioned Hyperledger Besu Testnet, QBFT Consensus)
- **Virtual ERC-20 Faucet Contract (`TestnetFaucet.sol` / Prompt 348):** When a user triggers a faucet claim, the service executes the off-chain ledger credit immediately, then enqueues an asynchronous blockchain minting job in PostgreSQL using the transactional outbox pattern.
- **Relayer Key Management:** An isolated, dedicated testnet relayer key held in HashiCorp Vault or AWS KMS signs EIP-1559 transactions calling `mint(address recipient, uint256 amount)` on the Besu testnet ERC-20 contracts.
- **Asset Token Mappings:**
  - `vUSDT`: Testnet ERC-20 contract pegged to 6 decimals ($10,000 \times 10^6$ units).
  - `vBTC`: Testnet ERC-20 contract pegged to 8 decimals ($1.00000000 \times 10^8$ units).
  - `vETH`: Testnet ERC-20 contract pegged to 18 decimals ($10.000000000000000000 \times 10^{18}$ units).
- **Strict Network Isolation Invariant:** Relayer keys and RPC endpoints are strictly locked to the Besu Testnet Chain ID (e.g., `13371`). Hardcoded runtime assertions prevent the service from ever dispatching transactions to production Besu Mainnet (Chain ID `13370`) or public Ethereum networks.
- **Non-Blocking Settlement:** On-chain testnet minting is decoupled from off-chain trading. If the Besu testnet experiences transient network halts or block delays under QBFT re-election, off-chain paper trading continues uninterrupted.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Service Repository:** Initialize the Go module at `services/demo-wallet-service` with standardized layout:
   - `cmd/server/`: Main application daemon, flags, configuration loading, and graceful shutdown.
   - `cmd/relayer/`: Standalone background worker processing on-chain minting outbox tasks.
   - `internal/config/`: Environment configuration struct with strict validation.
   - `internal/ledger/`: Core double-entry balance accounting domain logic.
   - `internal/faucet/`: Faucet eligibility engine, rate limiters, and claim coordinators.
   - `internal/repository/`: Database models and SQL queries generated via `sqlc`.
   - `internal/blockchain/`: Hyperledger Besu RPC client, contract ABI bindings, and nonce manager.
   - `internal/kafka/`: Kafka event producers and consumer group workers.
   - `internal/grpc/`: gRPC service implementations and interceptors.
2. **Define Protobuf Contracts:** Create `proto/growww/demowallet/v1/demo_wallet_service.proto` containing RPC definitions: `ClaimFaucet`, `GetVirtualBalance`, `ListVirtualBalances`, `ResetVirtualPortfolio`, `HoldVirtualBalance`, `ReleaseVirtualBalance`, `SettleVirtualTrade`, and `GetFaucetEligibility`.
3. **Generate Go Stubs & Gateway:** Compile protobuf files using `buf generate`, producing type-safe gRPC server stubs, validation helpers (`protoc-gen-validate`), and reverse-proxy REST gateways.
4. **Provision Isolated Database Schema:** Author idempotent SQL migration scripts in `migrations/` targeting `nbse_demo_wallet`. Create tables: `virtual_accounts`, `virtual_ledger_entries`, `virtual_transactions`, `faucet_claims`, `virtual_order_holds`, and `blockchain_mint_outbox`. Apply strict foreign keys, check constraints, and unique indices.
5. **Implement Compile-Time Persistence with `sqlc`:** Define SQL query statements in `db/queries/` covering balance checks with row-level locks (`SELECT ... FOR UPDATE`), atomic ledger inserts, hold state transitions, and outbox polling. Generate type-safe Go structs.
6. **Implement Double-Entry Balance Accounting Core:** Construct the ledger domain engine in `internal/ledger`:
   - Enforce fundamental accounting identity: every transaction must consist of balanced DEBIT and CREDIT postings ($\sum \text{Debits} = \sum \text{Credits}$).
   - Manage system balancing equity accounts (e.g., `EQUITY_FAUCET_DISBURSEMENT`, `EQUITY_PORTFOLIO_RESET`) to counterpart all user asset adjustments.
   - Guard against negative balances: available balance must satisfy $\text{available} \ge \text{hold\_amount}$ before placing holds.
7. **Implement Redis Sliding-Window Rate Limiter:** Author atomic Redis Lua scripts (`check_and_record_faucet.lua`) enforcing:
   - 24-hour fixed cooldown per `user_id`: key `faucet:cooldown:user:<user_id>` with TTL of 86400 seconds.
   - Sliding-window rate limit per client IP subnet: key `faucet:rate:ip:<subnet_hash>` tracking claims per rolling 60-minute window (limit: 5).
8. **Build Faucet Claim Coordinator:** Implement `ClaimFaucet` handler:
   - Validate incoming user JWT and extract verified `user_id` and client IP.
   - Execute Redis rate-limit script. If rate-limited, return `RESOURCE_EXHAUSTED` with remaining seconds.
   - Begin PostgreSQL serializable transaction: credit user asset accounts (10,000 vUSDT, 1 vBTC), debit platform faucet equity account, record transaction journal entries, insert `faucet_claims` record, and insert a pending job into `blockchain_mint_outbox`.
   - Publish `demo.wallet.faucet_claimed.v1` event to Kafka.
9. **Implement Portfolio Reset Pipeline:** Implement `ResetVirtualPortfolio` handler:
   - Lock user virtual accounts using `SELECT ... FOR UPDATE`.
   - Verify no open orders are in matching execution; notify Demo Matching Engine to cancel active orders.
   - Release existing virtual holds.
   - Post zeroing debit/credit adjustments transferring all current asset balances to `EQUITY_PORTFOLIO_RESET`.
   - Disburse fresh baseline balances: 10,000 vUSDT and 1 vBTC against `EQUITY_FAUCET_DISBURSEMENT`.
   - Record reset audit event and publish `demo.wallet.portfolio_reset.v1` to Kafka.
10. **Implement Order Hold and Release Handlers:** Author `HoldVirtualBalance` and `ReleaseVirtualBalance`:
    - `HoldVirtualBalance`: Atomically deduct requested amount from `available_balance` and add to `locked_balance` for specified asset. Record hold in `virtual_order_holds`.
    - `ReleaseVirtualBalance`: Reverse the hold, transferring amount from `locked_balance` back to `available_balance`. Mark hold status as `RELEASED`.
11. **Implement Atomic Trade Settlement Handler:** Author `SettleVirtualTrade`:
    - Process matched trades between Buyer and Seller.
    - Buyer: Deduct locked quote currency (e.g., vUSDT), credit base currency (e.g., vBTC).
    - Seller: Deduct locked base currency (e.g., vBTC), credit quote currency (e.g., vUSDT).
    - Execute within a single database transaction with deterministic account ordering to prevent deadlock.
    - Write double-entry journal records and publish `demo.wallet.settlement_completed.v1`.
12. **Build Asynchronous Testnet Blockchain Relayer:** Construct background outbox poller in `cmd/relayer`:
    - Poll `blockchain_mint_outbox` for `PENDING` records using `SELECT ... FOR UPDATE SKIP LOCKED`.
    - Connect to Besu Testnet JSON-RPC using `ethclient`.
    - Assemble, sign, and broadcast EIP-1559 transaction to `TestnetFaucet.sol`.
    - Await 1-block receipt confirmation under QBFT consensus; update outbox status to `CONFIRMED` with `on_chain_tx_hash`.
13. **Configure Kafka Producers and Consumers:** Set up Kafka publisher using `segmentio/kafka-go` with snappy compression and guaranteed `WaitForAll` acking. Wire consumer groups to handle demo order cancellation notifications.
14. **Implement OpenTelemetry Metrics & Health Checks:** Instrument service with Prometheus counters and histograms (`faucet_claims_total`, `faucet_rejections_total`, `ledger_transaction_latency_seconds`, `virtual_hold_duration_seconds`). Expose `/healthz`, `/readyz`, and `/metrics` over HTTP.
15. **Comprehensive Integration & Simulation Testing:** Develop test suites using `testcontainers-go` for PostgreSQL, Redis, and local Hyperledger Besu test nodes. Test concurrent faucet claims, rate-limit edge conditions, ledger balance invariances under 1,000 concurrent settlements, and zero-leakage isolation.

---

## Interfaces / Contracts

### Protobuf Definition (`demo_wallet_service.proto`)

```protobuf
syntax = "proto3";

package growww.demowallet.v1;

option go_package = "growww/demowallet/v1;demowalletv1";

// DemoWalletService manages virtual sandbox balances and paper-trading accounting.
service DemoWalletService {
  // Disburse baseline virtual funds (10,000 vUSDT and 1 vBTC) subject to 24-hour rate limit.
  rpc ClaimFaucet (ClaimFaucetRequest) returns (ClaimFaucetResponse);

  // Check user faucet eligibility, cooldown expiration, and remaining quota.
  rpc GetFaucetEligibility (GetFaucetEligibilityRequest) returns (GetFaucetEligibilityResponse);

  // Query virtual balances for a single asset.
  rpc GetVirtualBalance (GetVirtualBalanceRequest) returns (GetVirtualBalanceResponse);

  // List all virtual balances and open margin holds for a user.
  rpc ListVirtualBalances (ListVirtualBalancesRequest) returns (ListVirtualBalancesResponse);

  // Reset entire virtual portfolio back to canonical baseline (10,000 vUSDT, 1 vBTC).
  rpc ResetVirtualPortfolio (ResetVirtualPortfolioRequest) returns (ResetVirtualPortfolioResponse);

  // Place a temporary balance hold when a demo order is submitted to the matching engine.
  rpc HoldVirtualBalance (HoldVirtualBalanceRequest) returns (HoldVirtualBalanceResponse);

  // Release a temporary balance hold upon demo order cancellation or rejection.
  rpc ReleaseVirtualBalance (ReleaseVirtualBalanceRequest) returns (ReleaseVirtualBalanceResponse);

  // Atomically settle a matched demo trade between buyer and seller virtual accounts.
  rpc SettleVirtualTrade (SettleVirtualTradeRequest) returns (SettleVirtualTradeResponse);

  // Query historical double-entry virtual ledger journal entries for auditing.
  rpc ListVirtualLedgerEntries (ListVirtualLedgerEntriesRequest) returns (ListVirtualLedgerEntriesResponse);
}

enum VirtualAsset {
  VIRTUAL_ASSET_UNSPECIFIED = 0;
  VIRTUAL_ASSET_VUSDT = 1;
  VIRTUAL_ASSET_VBTC = 2;
  VIRTUAL_ASSET_VETH = 3;
  VIRTUAL_ASSET_VINR = 4;
}

enum HoldStatus {
  HOLD_STATUS_UNSPECIFIED = 0;
  HOLD_STATUS_ACTIVE = 1;
  HOLD_STATUS_SETTLED = 2;
  HOLD_STATUS_RELEASED = 3;
}

enum LedgerEntryType {
  LEDGER_ENTRY_TYPE_UNSPECIFIED = 0;
  LEDGER_ENTRY_TYPE_DEBIT = 1;
  LEDGER_ENTRY_TYPE_CREDIT = 2;
}

message ClaimFaucetRequest {
  string user_id = 1;
  string idempotency_key = 2;
  string client_ip = 3;
  string testnet_wallet_address = 4; // Optional: user on-chain Besu testnet address
}

message ClaimFaucetResponse {
  string claim_id = 1;
  string user_id = 2;
  repeated AssetGrant grants = 3;
  int64 claimed_at_unix_ms = 4;
  int64 next_eligible_at_unix_ms = 5;
  string on_chain_tx_hash = 6;       // Populated if minted on Besu Testnet
}

message AssetGrant {
  VirtualAsset asset = 1;
  string amount = 2;                // e.g., "10000.000000" or "1.00000000"
  string new_available_balance = 3;
}

message GetFaucetEligibilityRequest {
  string user_id = 1;
  string client_ip = 2;
}

message GetFaucetEligibilityResponse {
  string user_id = 1;
  bool is_eligible = 2;
  int64 cooldown_remaining_seconds = 3;
  int64 next_eligible_at_unix_ms = 4;
  string rejection_reason = 5;
}

message GetVirtualBalanceRequest {
  string user_id = 1;
  VirtualAsset asset = 2;
}

message GetVirtualBalanceResponse {
  string user_id = 1;
  VirtualAsset asset = 2;
  string available_balance = 3;
  string locked_balance = 4;
  string total_balance = 5;         // available + locked
  int64 updated_at_unix_ms = 6;
}

message ListVirtualBalancesRequest {
  string user_id = 1;
}

message ListVirtualBalancesResponse {
  string user_id = 1;
  repeated GetVirtualBalanceResponse balances = 2;
  string total_portfolio_equity_vusdt = 3;
}

message ResetVirtualPortfolioRequest {
  string user_id = 1;
  string idempotency_key = 2;
  string reason = 3;
}

message ResetVirtualPortfolioResponse {
  string user_id = 1;
  bool success = 2;
  repeated AssetGrant reset_balances = 3;
  int64 reset_at_unix_ms = 4;
}

message HoldVirtualBalanceRequest {
  string user_id = 1;
  string order_id = 2;
  string idempotency_key = 3;
  VirtualAsset asset = 4;
  string amount = 5;
}

message HoldVirtualBalanceResponse {
  string hold_id = 1;
  string order_id = 2;
  string user_id = 3;
  VirtualAsset asset = 4;
  string held_amount = 5;
  string remaining_available_balance = 6;
  HoldStatus status = 7;
  int64 created_at_unix_ms = 8;
}

message ReleaseVirtualBalanceRequest {
  string hold_id = 1;
  string user_id = 2;
  string order_id = 3;
  string idempotency_key = 4;
  string amount_to_release = 5; // Supports full or partial release
}

message ReleaseVirtualBalanceResponse {
  string hold_id = 1;
  string user_id = 2;
  string order_id = 3;
  string released_amount = 4;
  string new_available_balance = 5;
  HoldStatus status = 6;
  int64 released_at_unix_ms = 7;
}

message SettleVirtualTradeRequest {
  string trade_id = 1;
  string idempotency_key = 2;
  string buyer_user_id = 3;
  string buyer_order_id = 4;
  string seller_user_id = 5;
  string seller_order_id = 6;
  VirtualAsset base_asset = 7;        // e.g., VIRTUAL_ASSET_VBTC
  string base_amount = 8;             // e.g., "0.50000000"
  VirtualAsset quote_asset = 9;       // e.g., VIRTUAL_ASSET_VUSDT
  string quote_amount = 10;           // e.g., "32500.000000"
  string buyer_fee_amount = 11;       // in quote_asset
  string seller_fee_amount = 12;      // in quote_asset
}

message SettleVirtualTradeResponse {
  string trade_id = 1;
  bool settled = 2;
  string buyer_new_base_balance = 3;
  string buyer_new_quote_balance = 4;
  string seller_new_base_balance = 5;
  string seller_new_quote_balance = 6;
  int64 settled_at_unix_ms = 7;
}

message ListVirtualLedgerEntriesRequest {
  string user_id = 1;
  VirtualAsset asset = 2;
  int32 page_size = 3;
  string page_token = 4;
}

message VirtualLedgerEntry {
  string entry_id = 1;
  string transaction_id = 2;
  VirtualAsset asset = 3;
  LedgerEntryType entry_type = 4;
  string amount = 5;
  string balance_after = 6;
  string narrative = 7;
  int64 created_at_unix_ms = 8;
}

message ListVirtualLedgerEntriesResponse {
  repeated VirtualLedgerEntry entries = 1;
  string next_page_token = 2;
}
```

---

### PostgreSQL Database Schema DDL (`nbse_demo_wallet`)

```sql
-- PostgreSQL 16+ DDL Schema for Demo Trading Faucet & Virtual Balance Ledger Service
-- Database Name: nbse_demo_wallet (Physically isolated from real-money ledgers)

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Asset Types for Virtual Trading
CREATE TYPE virtual_asset_enum AS ENUM (
    'VUSDT',
    'VBTC',
    'VETH',
    'VINR'
);

-- Account Classification Types
CREATE TYPE virtual_account_type_enum AS ENUM (
    'USER_AVAILABLE',
    'USER_LOCKED',
    'EQUITY_FAUCET_DISBURSEMENT',
    'EQUITY_PORTFOLIO_RESET',
    'REVENUE_DEMO_EXCHANGE_FEE'
);

-- Transaction Classifications
CREATE TYPE virtual_transaction_type_enum AS ENUM (
    'FAUCET_DISBURSEMENT',
    'PORTFOLIO_RESET',
    'ORDER_LOCK',
    'ORDER_UNLOCK',
    'TRADE_SETTLEMENT',
    'FEE_DEDUCTION'
);

-- Directional Ledger Entry Type
CREATE TYPE ledger_entry_type_enum AS ENUM (
    'DEBIT',
    'CREDIT'
);

-- Order Margin Hold Status
CREATE TYPE virtual_hold_status_enum AS ENUM (
    'ACTIVE',
    'SETTLED',
    'RELEASED'
);

-- Outbox Relay Status for Blockchain Minting
CREATE TYPE outbox_status_enum AS ENUM (
    'PENDING',
    'SUBMITTED',
    'CONFIRMED',
    'FAILED'
);

-- Table 1: Virtual Balance Accounts
CREATE TABLE virtual_accounts (
    account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID, -- NULL for platform equity and revenue accounts
    asset_symbol virtual_asset_enum NOT NULL,
    account_type virtual_account_type_enum NOT NULL,
    balance NUMERIC(28, 8) NOT NULL DEFAULT 0.00000000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_virtual_account UNIQUE (user_id, asset_symbol, account_type),
    CONSTRAINT ck_user_balance_non_negative CHECK (
        account_type NOT IN ('USER_AVAILABLE', 'USER_LOCKED') OR balance >= 0
    )
);

-- Table 2: Virtual Financial Transactions (Header)
CREATE TABLE virtual_transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    transaction_type virtual_transaction_type_enum NOT NULL,
    reference_id VARCHAR(128), -- e.g. order_id, trade_id, or claim_id
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 3: Double-Entry Ledger Journal Entries (Lines)
CREATE TABLE virtual_ledger_entries (
    entry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES virtual_transactions(transaction_id) ON DELETE RESTRICT,
    account_id UUID NOT NULL REFERENCES virtual_accounts(account_id) ON DELETE RESTRICT,
    entry_type ledger_entry_type_enum NOT NULL,
    amount NUMERIC(28, 8) NOT NULL CHECK (amount > 0),
    balance_after NUMERIC(28, 8) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 4: Faucet Allocation Claims and Anti-Abuse Tracking
CREATE TABLE faucet_claims (
    claim_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    client_ip INET NOT NULL,
    client_subnet CIDR NOT NULL,
    vusdt_amount NUMERIC(28, 8) NOT NULL DEFAULT 10000.00000000,
    vbtc_amount NUMERIC(28, 8) NOT NULL DEFAULT 1.00000000,
    testnet_wallet_address VARCHAR(42), -- 0x-prefixed Besu testnet address
    on_chain_tx_hash VARCHAR(66),
    claimed_at_unix_ms BIGINT NOT NULL,
    next_eligible_at_unix_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 5: Virtual Order Holds for Matching Engine Margin Reservation
CREATE TABLE virtual_order_holds (
    hold_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    order_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    asset_symbol virtual_asset_enum NOT NULL,
    held_amount NUMERIC(28, 8) NOT NULL CHECK (held_amount > 0),
    released_amount NUMERIC(28, 8) NOT NULL DEFAULT 0.00000000 CHECK (released_amount >= 0),
    status virtual_hold_status_enum NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_released_lte_held CHECK (released_amount <= held_amount)
);

-- Table 6: Testnet Blockchain Minting Outbox (Hyperledger Besu)
CREATE TABLE blockchain_mint_outbox (
    outbox_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id UUID NOT NULL REFERENCES faucet_claims(claim_id) ON DELETE CASCADE,
    recipient_address VARCHAR(42) NOT NULL,
    asset_symbol virtual_asset_enum NOT NULL,
    token_amount NUMERIC(28, 8) NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    status outbox_status_enum NOT NULL DEFAULT 'PENDING',
    tx_hash VARCHAR(66),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance and Audit Indices
CREATE INDEX idx_virtual_accounts_user ON virtual_accounts (user_id);
CREATE INDEX idx_virtual_ledger_entries_account ON virtual_ledger_entries (account_id, created_at DESC);
CREATE INDEX idx_virtual_ledger_entries_tx ON virtual_ledger_entries (transaction_id);
CREATE INDEX idx_faucet_claims_user ON faucet_claims (user_id, claimed_at_unix_ms DESC);
CREATE INDEX idx_faucet_claims_ip ON faucet_claims (client_ip, claimed_at_unix_ms DESC);
CREATE INDEX idx_faucet_claims_subnet ON faucet_claims (client_subnet, claimed_at_unix_ms DESC);
CREATE INDEX idx_virtual_order_holds_order ON virtual_order_holds (order_id);
CREATE INDEX idx_virtual_order_holds_user_active ON virtual_order_holds (user_id, status) WHERE status = 'ACTIVE';
CREATE INDEX idx_mint_outbox_pending ON blockchain_mint_outbox (status, retry_count) WHERE status IN ('PENDING', 'SUBMITTED');

-- Seed System Balancing Equity Accounts
INSERT INTO virtual_accounts (account_id, user_id, asset_symbol, account_type, balance)
VALUES
    ('00000000-0000-0000-0000-000000000001', NULL, 'VUSDT', 'EQUITY_FAUCET_DISBURSEMENT', 0),
    ('00000000-0000-0000-0000-000000000002', NULL, 'VBTC',  'EQUITY_FAUCET_DISBURSEMENT', 0),
    ('00000000-0000-0000-0000-000000000003', NULL, 'VETH',  'EQUITY_FAUCET_DISBURSEMENT', 0),
    ('00000000-0000-0000-0000-000000000004', NULL, 'VINR',  'EQUITY_FAUCET_DISBURSEMENT', 0),
    ('00000000-0000-0000-0000-000000000011', NULL, 'VUSDT', 'EQUITY_PORTFOLIO_RESET', 0),
    ('00000000-0000-0000-0000-000000000012', NULL, 'VBTC',  'EQUITY_PORTFOLIO_RESET', 0),
    ('00000000-0000-0000-0000-000000000013', NULL, 'VETH',  'EQUITY_PORTFOLIO_RESET', 0),
    ('00000000-0000-0000-0000-000000000014', NULL, 'VINR',  'EQUITY_PORTFOLIO_RESET', 0),
    ('00000000-0000-0000-0000-000000000021', NULL, 'VUSDT', 'REVENUE_DEMO_EXCHANGE_FEE', 0)
ON CONFLICT DO NOTHING;
```

---

## Security & Compliance Notes

- **Complete Segregation from Real Money & Production Ledgers:**
  - The demo wallet service runs on a completely separate PostgreSQL database instance (`nbse_demo_wallet`) with dedicated credentials. No database links, cross-schema joins, or shared tables exist between `nbse_demo_wallet` and production financial databases (`nbse_wallet`, `nbse_settlement`).
  - Production payment gateways (UPI, IMPS, NEFT) and fiat collection rails have zero network connectivity to `demo-wallet-service`.
  - Account numbers, transaction IDs, and hold IDs are prefixed with `demo_` or formatted via isolated UUID namespaces to prevent accidental processing in real settlement jobs.
- **Anti-Abuse & Rate-Limiting Protections:**
  - **Account Cooldown:** Each user account is strictly limited to 1 faucet claim every 86,400 seconds (24 hours). This is enforced via both Redis TTL keys (`SET NX EX 86400`) and relational database unique constraints on rolling claim timestamps.
  - **IP Subnet Limiting:** Client IP addresses are parsed into CIDR blocks (`/24` for IPv4, `/48` for IPv6). Maximum 5 faucet claims are permitted per subnet per rolling 60-minute window to prevent sybil botnets from exhausting testnet resources.
  - **Account Verification Prerequisite:** Users must possess a valid, authenticated user session (Prompt 201) with basic email or mobile verification before invoking the faucet endpoint. Anonymous bot requests are rejected at the gateway.
- **Invariable Double-Entry Invariants:**
  - The accounting engine verifies that for every transaction:
    $$\sum \text{Debits} - \sum \text{Credits} = 0$$
  - User available balances are protected by SQL `CHECK (balance >= 0)` constraints. An order hold or trade settlement that would cause an available balance to drop below zero is rejected immediately with an ACID rollback.
  - All balance-mutating transactions acquire row-level locks (`SELECT ... FOR UPDATE`) sorted deterministically by `account_id` to eliminate deadlocks under high-concurrency order submission.
- **Testnet Blockchain Integrity & Relayer Protection:**
  - The blockchain relayer operates with an isolated private key dedicated to Hyperledger Besu Testnet. The relayer key is strictly forbidden from holding mainnet ether or native gas tokens.
  - Gas price caps and nonce trackers prevent relayer stalling. If the testnet RPC becomes unresponsive, outbox jobs enter exponential backoff without stalling off-chain paper trading.

---

## Acceptance Criteria

- [ ] **Standard Faucet Allocation:** Calling `ClaimFaucet` on an eligible account atomically credits exactly `10,000.00000000` vUSDT and `1.00000000` vBTC to the user's available virtual accounts.
- [ ] **24-Hour Cooldown Enforcement:** Attempting to call `ClaimFaucet` a second time within 24 hours of a prior successful claim returns gRPC status `RESOURCE_EXHAUSTED` with the exact remaining cooldown seconds and does not mutate balances.
- [ ] **IP Subnet Throttling:** Submitting more than 5 faucet requests within 1 hour from the same `/24` IPv4 or `/48` IPv6 subnet triggers rate-limiting rejection, even across different user accounts.
- [ ] **Double-Entry Balance Invariance:** Every faucet disbursement, margin hold, release, trade settlement, and portfolio reset records balanced journal entries where $\sum \text{Debits} = \sum \text{Credits}$.
- [ ] **Negative Balance Prevention:** Any transaction attempting to debit more than the current `available_balance` fails with `FAILED_PRECONDITION` and leaves the ledger untouched.
- [ ] **1-Click Portfolio Reset:** Calling `ResetVirtualPortfolio` cancels active virtual order holds, resets user virtual balances back to 10,000 vUSDT and 1 vBTC, and logs balanced adjustment entries.
- [ ] **Demo Matching Engine Margin Holds:** `HoldVirtualBalance` successfully locks funds by transferring the specified amount from `available_balance` to `locked_balance` in sub-5ms latency; `ReleaseVirtualBalance` returns funds to `available_balance`.
- [ ] **Bilateral Trade Settlement:** `SettleVirtualTrade` atomically transfers quote currency from buyer to seller and base currency from seller to buyer while deducting virtual demo trading fees in a single database transaction.
- [ ] **Testnet Blockchain Relaying:** Successful faucet claims enqueue outbox records that mint virtual ERC-20 tokens on Hyperledger Besu Testnet and update `on_chain_tx_hash` upon confirmation.
- [ ] **Physical Datastore Isolation:** Verified that `services/demo-wallet-service` connects exclusively to `nbse_demo_wallet` and has no network or credential access to production real-money databases.
- [ ] **High-Concurrency Resilience:** Integration test simulates 500 concurrent hold and settlement operations across 50 simulated user accounts without deadlocks, phantom reads, or balance discrepancies.

---

## Suggested Order / Dependencies

- **Prerequisites:**
  - `103_api_design_standards.md`: gRPC and RESTful error code standards.
  - `104_event_schema_and_kafka_topic_standards.md`: Kafka topic naming and Protobuf schema standards.
  - `111_domain_model_core_entities.md`: Core currency and account model conventions.
  - `112_idempotency_and_exactly_once_processing.md`: Idempotency keys and outbox pattern standards.
  - `201_user_account_service.md`: User identification, JWT validation, and account lifecycle.
- **Parallel Tasks:**
  - `273_demo_matching_engine_and_virtual_order_book.md`: Virtual matching engine consuming order holds and trade settlements.
  - `348_testnet_faucet_smart_contract.md`: Solidity ERC-20 faucet contracts on Hyperledger Besu Testnet.
  - `527_flutter_environment_switcher_and_sandbox_mode.md`: Mobile UI environment switching and faucet trigger sheet.
- **Downstream Blockers:**
  - `609_developer_portal_and_testnet_faucet_ui.md`: Web developer portal and 1-click testnet faucet interface.
  - `715_sandbox_end_to_end_paper_trading_simulation.md`: Automated synthetic user simulation executing paper trades against the demo cluster.
