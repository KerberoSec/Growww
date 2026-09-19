# 270 - Request-For-Quote (RFQ) & Instant Convert Swap Service (Go / Rust)

## Purpose
Retail and institutional participants in modern electronic financial ecosystems demand a frictionless, single-click conversion experience that abstracts away the complexities of Level-2 Central Limit Order Books (CLOB), bid-ask spread math, order types, and execution slippage. In traditional equity and multi-currency exchanges, retail market orders submitted during volatile market regimes suffer from significant adverse price slippage, latency front-running, and partial fills.

As codified in ADR-0033, the **Request-For-Quote (RFQ) & Instant Convert Swap Service** delivers a zero-slippage, guaranteed-price instant conversion facility. It enables users to execute instantaneous two-way swaps between domestic fiat (INR), official retail and wholesale Central Bank Digital Currency (eINR CBDC), dollar-pegged stablecoins (USDT), and fractional tokenized digital securities (e.g., tokenized RELIANCE, TATA, Government Securities, or MCX Gold vault receipts) deployed on the Hyperledger Besu permissioned ledger.

The service acts as a programmatic pricing and risk-absorbing liquidity principal, backed directly by Treasury market-maker inventory and automated internal hedging protocols. It computes fair-value pricing, issues cryptographically signed, guaranteed price quotes locked for a 5-second execution window, guarantees zero execution slippage upon commitment, and coordinates atomic two-phase settlement execution across off-chain double-entry ledgers and on-chain Delivery-versus-Payment (DvP) contracts.

---

## What You Are Building
A high-throughput, sub-millisecond RFQ conversion engine (`services/rfq-swap-service`) engineered in Go 1.22+ (or Rust with Tokio/Tonic) that functions as the central conversion authority for Growww and NBSE. Concrete deliverables include:

- **Dynamic RFQ Pricing Engine:** A real-time mathematical pricing core that consumes National Best Bid and Offer (NBBO) market data, Consolidated Tape feeds (Prompt 207 / Prompt 258), and Treasury inventory skew levels to compute two-way conversion prices. It incorporates a fair-value spread calculation with dynamic volatility buffers while strictly maintaining the statutory 0.00% (No fee at all) platform fee invariance without hidden price gouging.
- **5-Second Ephemeral Quote Lock Manager:** A Redis-backed state machine utilizing atomic Lua scripts to register, lock, and enforce a 5.00-second guaranteed execution window. Every quote includes an HMAC-SHA256 signature binding the quote ID, user ID, trading pair, direction, guaranteed execution rate, fee breakdown, and millisecond expiration timestamp.
- **Bi-Directional gRPC Quote Streaming Service:** High-frequency, multiplexed server-streaming and bi-directional gRPC channels (`GetQuoteStream`, `StreamLiveQuotes`) enabling Flutter mobile clients and web dashboards to receive sub-50ms tick updates and synchronized execution countdown timers.
- **Two-Phase Atomic Settlement Coordinator (Prepare & Commit):** An atomic execution pipeline that eliminates partial execution risk:
  - *Phase 1 (Prepare):* Pre-execution balance verification and cryptographic hold placement across source funds (INR/eINR/USDT in Wallet Service Prompt 203 or tokenized securities in Portfolio Service Prompt 209).
  - *Phase 2 (Commit):* Atomic ledger transfer, fee debit (0.00% (Zero Fee)), Treasury inventory re-allocation, and on-chain DvP contract invocation on Hyperledger Besu.
  - *Rollback:* Automated timeout release returning locked balances if user fails to commit within the 5-second lock or if system execution aborts.
- **Treasury Market-Maker Inventory & Auto-Hedging Router:** Real-time inventory tracking for the platform Treasury pool across all supported conversion assets. When an instant swap consumes inventory past predefined inventory delta thresholds, the service dispatches asynchronous hedging orders to internal matching engine books (Prompt 205) or institutional liquidity pools.

---

## Scope Boundaries

### In Scope
- Real-time ingestion of live reference prices, depth curves, and volatility parameters from Market Data Service (Prompt 207) and NBBO Tape (Prompt 258).
- Mathematical calculation of synthetic conversion rates across diverse asset classes: Fiat (INR), CBDC (eINR), Stablecoin (USDT), and fractional tokenized securities (ERC-20/ERC-3643 equivalents on Besu).
- Generation of deterministic, tamper-proof quotes with 5.00-second validity windows locked in memory via Redis 7.2.
- Cryptographic quote signature generation and verification (HMAC-SHA256) preventing quote tampering or rate alteration.
- User tier entitlement and velocity check enforcement (Tier 1 Basic, Tier 2 Verified, Tier 3 Accredited) in coordination with Risk Engine (Prompt 206).
- Two-phase commit execution across off-chain financial ledgers (Prompt 203) and on-chain Besu DvP settlement contracts.
- Statutory 0.00% (No fee at all) platform fee calculation and audit reporting (Prompt 210).
- Emitting atomic execution and inventory rebalance events to Apache Kafka.
- Historical swap statement lookups and compliance audit queries.

### Out of Scope / Handled Elsewhere
- Central Limit Order Book continuous matching algorithms and order book state maintenance (handled in Prompt 205 and Prompt 253).
- Primary fiat banking rails, UPI instant collections, and IMPS/NEFT payment gateways (handled in Prompt 212).
- Core RBI Digital Rupee wholesale/retail host banking integration (handled in Prompt 232).
- Physical depository custody lockups and NSDL/CDSL demat account reconciliation (handled in Prompt 213).
- Primary user identity onboarding and PAN/Aadhaar KYC verification (handled in Prompt 202).
- Cross-chain bridge relays and proof verification (handled in Prompt 235 and Prompt 238).

---

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+ or Rust 1.78+. Go provides lightweight goroutines, deterministic concurrency, and sub-millisecond gRPC processing. Alternatively, Rust with Tokio and Tonic provides zero-cost abstractions, memory safety without garbage collection pauses, and microsecond-level calculation latency.
- **In-Memory Quote Lock Store:** Redis 7.2+ Cluster using Redis Hashes for active quote payloads, atomic Redis Lua scripts for non-blocking single-use claim execution, and Redis Key Expiration notifications (`__keyevent@*__:expired`) to drive cleanup.
- **Relational Persistence & Audit Ledger:** PostgreSQL 16+ utilizing `pgx/v5` (Go) or `sqlx` (Rust) with strict ACID transactions, read replicas for trade history queries, and partitioned execution tables.
- **Message Broker & Event Streaming:** Apache Kafka 3.7+ (`segmentio/kafka-go` or `rdkafka`) for dispatching atomic execution notices (`rfq.swaps.executed.v1`) and treasury hedging commands (`rfq.treasury.hedge.v1`).
- **RPC & Communication Framework:** gRPC (`google.golang.org/grpc` or `tonic`) over HTTP/2 with protobuf serialization for sub-millisecond client streaming and internal microservice orchestration.
- **High-Precision Arithmetic:** `github.com/shopspring/decimal` (Go) or `rust_decimal` (Rust) ensuring exact fixed-point financial calculation with Banker's Rounding (`ROUND_HALF_EVEN`) to eliminate floating-point drift.
- **Blockchain Connectivity:** `go-ethereum/ethclient` or `ethers-rs` connecting to permissioned Hyperledger Besu JSON-RPC nodes with EIP-712 structured signing.

---

## Backend / Infra Touchpoints
- **Market Data Service (Prompt 207):** Ingests streaming L2 order book top-of-book prices and tick-by-tick trades via gRPC streaming or Kafka to calculate real-time synthetic mid-prices and liquidity depth.
- **Wallet & Account Service (Prompt 203):** Coordinates two-phase fund holds and balances. Pre-allocates and holds user source assets, debits source balances, credits destination balances, and records double-entry journal postings.
- **Fee & Realized PnL Engine (Prompt 210):** Validates and posts the Universal 0.00% Zero-Fee Invariant (No fee at all) on every completed swap, routing fee revenues to exchange treasury accounts.
- **Risk & Margin Checks Service (Prompt 206):** Ingests pre-trade velocity metrics, user tier limits, and maximum single-swap notional thresholds before issuing binding quotes.
- **Treasury Allocation & Inventory Management:** Monitors platform liquidity inventory across INR, eINR, USDT, and digital securities. Emits inventory skew metrics to the pricing engine and triggers automated rebalancing when inventory boundaries are breached.
- **Trade Settlement Service (Prompt 208):** Hand-off coordinator for institutional post-trade reporting and statutory exchange clearing.
- **Audit Log & Surveillance Service (Prompt 218 / Prompt 228):** Ingests quote generation logs, expiration rates, latency metrics, and user execution histories to detect potential arbitrage, latency gaming, or front-running abuse.

---

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Atomic DvP Smart Contract (`RFQAtomicSwap.sol` / `SettlementDvP.sol`):** Swaps involving tokenized fractional securities (ERC-3643 or permissioned ERC-20 backed 1:1 by custodian demat shares) execute through an atomic Delivery-versus-Payment smart contract deployed on Hyperledger Besu under QBFT consensus (2-second deterministic finality).
- **Token-for-Token and Fiat-for-Token Swaps:**
  - *Fiat/CBDC to Token:* The service locks fiat/eINR off-chain in Wallet Service (Prompt 203 / Prompt 232), and calls the DvP contract using the platform treasury relayer key to transfer fractional security tokens from the Treasury Inventory Vault to the user's whitelisted on-chain address.
  - *Token to Fiat/CBDC:* The user's pre-approved allowance or ERC-2612 permit transfers fractional security tokens to the Treasury Vault, unlocking instant fiat or eINR credit in the user's wallet.
  - *Token to Token:* Direct atomic transfer inside `RFQAtomicSwap.sol`, transferring Asset A from user to Treasury while transferring Asset B from Treasury to user in a single atomic transaction.
- **Zero On-Chain PII Invariant:** The Besu ledger only records anonymous cryptographic wallet addresses, token contract identifiers, fractional token amounts (up to 18 decimal places), and the off-chain `swap_id` hash. No user names, PAN numbers, bank accounts, or IP addresses are written to the blockchain.
- **Finality & Idempotency:** The on-chain transaction hash is linked to the off-chain `rfq_swap_executions` record. If an on-chain execution reverts, the two-phase coordinator executes an immediate rollback of the off-chain holds.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Service Workspace:** Initialize the Go module or Rust crate at `services/rfq-swap-service` with standardized directory layout:
   - `cmd/server/`: Main application lifecycle, dependency injection, and signal handling.
   - `internal/pricing/`: Fair-value calculation engine, spread algorithms, and inventory skew calculators.
   - `internal/quotemanager/`: Redis quote lock lifecycle, TTL timers, and HMAC verification.
   - `internal/settlement/`: Two-phase atomic coordinator (Prepare, Commit, Rollback).
   - `internal/treasury/`: Inventory allocation tracker and auto-hedging Kafka publisher.
   - `internal/blockchain/`: Hyperledger Besu DvP client and smart contract bindings.
   - `internal/store/`: PostgreSQL schemas and queries via `sqlc` or `sqlx`.
2. **Define Protobuf Specifications:** Author `proto/growww/rfq/v1/rfq_service.proto` defining RPCs: `RequestQuote`, `ExecuteQuote`, `GetQuoteStream`, `GetQuoteStatus`, and `ListSwapHistory`.
3. **Compile Protobuf & gRPC Stubs:** Generate type-safe Go/Rust server and client interfaces using `buf` or `protoc` with gRPC validation extensions.
4. **Design PostgreSQL Schema:** Write database migration scripts defining tables: `rfq_quotes`, `rfq_swap_executions`, `treasury_inventory_positions`, and `rfq_tier_limits`. Apply primary keys, foreign key constraints, check constraints, and performance indexes.
5. **Configure `sqlc` or `sqlx` Persistence Layer:** Implement zero-allocation, type-safe database access logic for quote recording, execution state transitions, and audit trails.
6. **Implement Market Data Ingestion Pipeline:** Establish high-throughput gRPC client subscriptions to Market Data Service (Prompt 207) and NBBO Tape (Prompt 258) to maintain an in-memory lock-free cache (`sync.Map` or crossbeam channels) of current bid, ask, mid-price, and rolling volatility.
7. **Build Pricing & Spread Algorithm:** Implement the mathematical pricing core:
   - Calculate synthetic mid-price: $P_{\text{mid}} = \frac{P_{\text{bid}} + P_{\text{ask}}}{2}$.
   - Add dynamic inventory skew buffer: $\Delta_{\text{skew}} = \alpha \cdot \left(\frac{I_{\text{current}} - I_{\text{target}}}{I_{\text{target}}}\right)$.
   - Apply statutory 0.00% (No fee at all) platform fee invariance: $\text{Fee} = \text{Notional} \times 0.0001$.
   - Compute final guaranteed conversion price $P_{\text{quote}}$ with zero additional hidden retail spread markup.
8. **Build Ephemeral Quote Lock Manager in Redis:** Author atomic Redis Lua scripts (`create_quote_lock.lua`, `claim_quote_lock.lua`, `release_quote_lock.lua`) enforcing strict 5.00-second TTLs, preventing race conditions and ensuring that each quote can be claimed exactly once.
9. **Implement HMAC-SHA256 Cryptographic Signing:** Generate tamper-evident signature payloads encompassing: `quote_id`, `user_id`, `from_asset`, `to_asset`, `from_amount`, `to_amount`, `guaranteed_rate`, and `expires_at_unix_ms`.
10. **Implement Pre-Trade Tier and Risk Guards:** Connect to Risk Service (Prompt 206) via gRPC to validate user KYC tier (Tier 1 limit: ₹50,000/day; Tier 2 limit: ₹10,00,000/day; Tier 3: Unlimited/Institutional) and velocity caps before quote issuance.
11. **Build Two-Phase Settlement Coordinator:**
    - *Prepare Step:* Invoke Wallet Service (Prompt 203) or Portfolio Service (Prompt 209) via gRPC to place an atomic financial hold on the user's source funds.
    - *Commit Step:* Transition quote in Redis to `EXECUTING`, debit user source funds, credit destination funds, deduct the 0.00% fee (No fee at all), credit Treasury inventory, and trigger on-chain Besu settlement if tokenized assets are involved.
    - *Failure Recovery:* Release holds immediately if on-chain transaction reverts or network RPC times out.
12. **Implement Hyperledger Besu DvP Adapter:** Build transaction relayer integrating with `RFQAtomicSwap.sol` or `SettlementDvP.sol` on Besu, submitting EIP-1559/legacy transactions under QBFT consensus and awaiting 1-block receipt confirmation.
13. **Implement Treasury Auto-Hedging Publisher:** Calculate cumulative net inventory imbalances. When net exposure exceeds safety parameters, publish an auto-hedge intent event to Kafka topic `rfq.treasury.hedge.v1` for downstream execution on NBSE order books.
14. **Expose gRPC & REST Endpoints:** Implement gRPC handlers with OpenTelemetry distributed tracing and expose RESTful endpoints via `grpc-gateway` for browser and mobile compatibility.
15. **Implement Comprehensive Test Suite:** Author end-to-end integration tests using `testcontainers` for PostgreSQL and Redis, verifying quote expiration after 5.001 seconds, single-use claim idempotency, zero-slippage execution under market volatility, and atomic rollbacks on simulated settlement failure.

---

## Interfaces / Contracts

### Protobuf Definition (`rfq_service.proto`)
```protobuf
syntax = "proto3";

package growww.rfq.v1;

option go_package = "growww/rfq/v1;rfqv1";

// The RFQ Service provides guaranteed-rate conversion and instant swaps.
service RFQService {
  // Request a binding, guaranteed-rate quote locked for 5.00 seconds.
  rpc RequestQuote (RequestQuoteRequest) returns (RequestQuoteResponse);

  // Execute a previously locked quote within its valid 5-second window.
  rpc ExecuteQuote (ExecuteQuoteRequest) returns (ExecuteQuoteResponse);

  // Stream continuous indicative conversion prices and countdown timers.
  rpc GetQuoteStream (GetQuoteStreamRequest) returns (stream QuoteStreamUpdate);

  // Query status and audit details of a specific quote.
  rpc GetQuoteStatus (GetQuoteStatusRequest) returns (GetQuoteStatusResponse);

  // List past swap conversion executions for a specific user.
  rpc ListSwapHistory (ListSwapHistoryRequest) returns (ListSwapHistoryResponse);
}

enum AssetType {
  ASSET_TYPE_UNSPECIFIED = 0;
  ASSET_TYPE_FIAT_INR = 1;
  ASSET_TYPE_CBDC_EINR = 2;
  ASSET_TYPE_STABLECOIN_USDT = 3;
  ASSET_TYPE_TOKENIZED_SECURITY = 4;
}

enum QuoteStatus {
  QUOTE_STATUS_UNSPECIFIED = 0;
  QUOTE_STATUS_LOCKED = 1;     // Active and executable within 5-second lock
  QUOTE_STATUS_EXECUTING = 2;  // Claimed and settlement in-progress
  QUOTE_STATUS_COMPLETED = 3;  // Successfully settled with zero slippage
  QUOTE_STATUS_EXPIRED = 4;    // 5-second window elapsed without execution
  QUOTE_STATUS_REJECTED = 5;   // Rejected due to risk checks or insufficient liquidity
  QUOTE_STATUS_FAILED = 6;     // Settlement execution aborted and rolled back
}

message RequestQuoteRequest {
  string user_id = 1;
  string idempotency_key = 2;
  string from_asset = 3;             // e.g., "INR", "EINR", "USDT", "INE002A01018"
  AssetType from_asset_type = 4;
  string to_asset = 5;               // e.g., "INE002A01018", "USDT", "INR"
  AssetType to_asset_type = 6;
  string from_amount = 7;            // Exact amount user wishes to swap
  bool is_exact_out = 8;             // If true, to_amount is fixed and from_amount is calculated
}

message RequestQuoteResponse {
  string quote_id = 1;
  string user_id = 2;
  string from_asset = 3;
  string to_asset = 4;
  string from_amount = 5;
  string to_amount = 6;              // Guaranteed payout amount received
  string guaranteed_rate = 7;        // Exact conversion multiplier
  string platform_fee_amount = 8;    // Universal 0.00% Zero-Fee Invariant (No fee at all) in source asset
  string platform_fee_bps = 9;       // Always "0.00" (0.00% Zero-Fee at launch)
  int64 valid_duration_ms = 10;      // Exactly 5000 (5 seconds)
  int64 issued_at_unix_ms = 11;
  int64 expires_at_unix_ms = 12;
  string quote_signature = 13;       // HMAC-SHA256 signature verifying quote integrity
  QuoteStatus status = 14;
}

message ExecuteQuoteRequest {
  string quote_id = 1;
  string user_id = 2;
  string quote_signature = 3;        // Validated against server secret
  string idempotency_key = 4;
  int64 client_submitted_at_unix_ms = 5;
}

message ExecuteQuoteResponse {
  string swap_id = 1;
  string quote_id = 2;
  string user_id = 3;
  QuoteStatus status = 4;
  string from_asset = 5;
  string to_asset = 6;
  string from_amount_debited = 7;
  string to_amount_credited = 8;
  string fee_debited = 9;
  string effective_rate = 10;        // Exactly matches guaranteed_rate (zero slippage)
  string on_chain_tx_hash = 11;      // Besu transaction hash for DvP swaps (if applicable)
  int64 executed_at_unix_ms = 12;
}

message GetQuoteStreamRequest {
  string user_id = 1;
  string from_asset = 2;
  string to_asset = 3;
  string indicative_amount = 4;
}

message QuoteStreamUpdate {
  string from_asset = 1;
  string to_asset = 2;
  string indicative_rate = 3;
  string estimated_payout = 4;
  string estimated_fee = 5;
  int64 server_time_unix_ms = 6;
}

message GetQuoteStatusRequest {
  string quote_id = 1;
  string user_id = 2;
}

message GetQuoteStatusResponse {
  string quote_id = 1;
  QuoteStatus status = 2;
  int64 time_remaining_ms = 3;
  RequestQuoteResponse quote_details = 4;
}

message ListSwapHistoryRequest {
  string user_id = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message ListSwapHistoryResponse {
  repeated ExecuteQuoteResponse swaps = 1;
  string next_page_token = 2;
}
```

---

### PostgreSQL Database Schema DDL

```sql
-- PostgreSQL 16+ DDL Schema for RFQ Instant Convert & Swap Service

CREATE TYPE rfq_asset_type_enum AS ENUM (
    'FIAT_INR',
    'CBDC_EINR',
    'STABLECOIN_USDT',
    'TOKENIZED_SECURITY'
);

CREATE TYPE rfq_quote_status_enum AS ENUM (
    'LOCKED',
    'EXECUTING',
    'COMPLETED',
    'EXPIRED',
    'REJECTED',
    'FAILED'
);

-- Table 1: Ephemeral and Historical RFQ Quotes
CREATE TABLE rfq_quotes (
    quote_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    from_asset VARCHAR(32) NOT NULL,
    from_asset_type rfq_asset_type_enum NOT NULL,
    to_asset VARCHAR(32) NOT NULL,
    to_asset_type rfq_asset_type_enum NOT NULL,
    from_amount NUMERIC(28, 8) NOT NULL CHECK (from_amount > 0),
    to_amount NUMERIC(28, 8) NOT NULL CHECK (to_amount > 0),
    guaranteed_rate NUMERIC(28, 10) NOT NULL CHECK (guaranteed_rate > 0),
    platform_fee_amount NUMERIC(28, 8) NOT NULL CHECK (platform_fee_amount >= 0),
    platform_fee_bps NUMERIC(6, 2) NOT NULL DEFAULT 0.00 CHECK (platform_fee_bps >= 0.00), -- Invariant 0.00% (Zero Fee at launch)
    quote_signature VARCHAR(128) NOT NULL,
    status rfq_quote_status_enum NOT NULL DEFAULT 'LOCKED',
    rejection_reason TEXT,
    issued_at_unix_ms BIGINT NOT NULL,
    expires_at_unix_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 2: Confirmed Swap Executions
CREATE TABLE rfq_swap_executions (
    swap_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id UUID NOT NULL REFERENCES rfq_quotes(quote_id) ON DELETE RESTRICT,
    user_id UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    from_asset VARCHAR(32) NOT NULL,
    to_asset VARCHAR(32) NOT NULL,
    from_amount_debited NUMERIC(28, 8) NOT NULL,
    to_amount_credited NUMERIC(28, 8) NOT NULL,
    fee_debited NUMERIC(28, 8) NOT NULL,
    effective_rate NUMERIC(28, 10) NOT NULL,
    on_chain_tx_hash VARCHAR(66), -- 0x-prefixed 32-byte hash for Besu DvP settlements
    settlement_phase VARCHAR(32) NOT NULL DEFAULT 'COMMITTED',
    executed_at_unix_ms BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 3: Treasury Inventory Positions & Skew Tracking
CREATE TABLE treasury_inventory_positions (
    asset_symbol VARCHAR(32) PRIMARY KEY,
    asset_type rfq_asset_type_enum NOT NULL,
    total_inventory NUMERIC(28, 8) NOT NULL DEFAULT 0.00000000,
    reserved_inventory NUMERIC(28, 8) NOT NULL DEFAULT 0.00000000,
    available_inventory NUMERIC(28, 8) GENERATED ALWAYS AS (total_inventory - reserved_inventory) STORED,
    target_inventory NUMERIC(28, 8) NOT NULL,
    max_inventory_limit NUMERIC(28, 8) NOT NULL,
    min_inventory_limit NUMERIC(28, 8) NOT NULL,
    skew_bps INT NOT NULL DEFAULT 0,
    last_rebalanced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table 4: User KYC Tier Swap Volume & Velocity Limits
CREATE TABLE rfq_tier_limits (
    tier_id INT PRIMARY KEY,
    tier_name VARCHAR(64) NOT NULL,
    max_single_swap_inr NUMERIC(18, 2) NOT NULL,
    daily_volume_limit_inr NUMERIC(18, 2) NOT NULL,
    cooldown_seconds INT NOT NULL DEFAULT 0,
    requires_accreditation BOOLEAN NOT NULL DEFAULT FALSE
);

-- Seed Default KYC Tier Swap Limits
INSERT INTO rfq_tier_limits (tier_id, tier_name, max_single_swap_inr, daily_volume_limit_inr, cooldown_seconds, requires_accreditation)
VALUES
    (1, 'Tier 1 Basic', 50000.00, 100000.00, 5, FALSE),
    (2, 'Tier 2 Verified', 1000000.00, 5000000.00, 0, FALSE),
    (3, 'Tier 3 Institutional', 100000000.00, 500000000.00, 0, TRUE);

-- Indexes for performance and lookup speed
CREATE INDEX idx_rfq_quotes_user_status ON rfq_quotes (user_id, status);
CREATE INDEX idx_rfq_quotes_expires_at ON rfq_quotes (expires_at_unix_ms) WHERE status = 'LOCKED';
CREATE INDEX idx_rfq_swap_executions_user ON rfq_swap_executions (user_id, created_at DESC);
CREATE INDEX idx_rfq_swap_executions_quote ON rfq_swap_executions (quote_id);
```

---

## Security & Compliance Notes

- **Prevention of Quote Front-Running and Latency Arbitrage:**
  - *Server-Enforced Expiration:* The 5.00-second quote lifetime is calculated strictly by the server clock. When `ExecuteQuote` arrives, the service computes $\Delta t = T_{\text{server}} - T_{\text{issued}}$. If $\Delta t > 5000\text{ms}$, the request is unconditionally rejected with code `QUOTE_EXPIRED`.
  - *HMAC-SHA256 Tamper Protection:* All quote attributes are hashed with a revolving CloudHSM-managed secret. Any client-side parameter manipulation (such as tampering with `guaranteed_rate` or `to_amount`) causes immediate cryptographic signature validation failure.
  - *Single-Use Lock via Redis Lua:* The quote lock is claimed via an atomic Redis Lua script that checks if status is `LOCKED` and immediately transitions it to `CLAIMED` in a single uninterrupted operation. Concurrent execution requests referencing the same `quote_id` fail instantly.
- **Strict 0.00% (Zero Fee) Platform Fee Invariance:**
  - In compliance with ADR-0033 and NBSE fair-access rules, the service guarantees a flat, unalterable platform fee of 0.00% (Zero Fee) (1.00 bps) across all conversion pairs.
  - Hidden spread markups, payment-for-order-flow (PFOF) gouging, or asymmetric slippage skews targeting retail users are mathematically precluded: the quoted conversion rate mirrors fair-market value, and the fee is isolated as a transparent 0.00% (Zero Fee) deduction displayed prominently to the user.
- **Tiered Volume Caps & Anti-Structuring Enforcement:**
  - Enforces SEBI and PMLA anti-structuring limits per user tier. Daily cumulative swap volumes are tracked in Redis sliding windows. Submissions exceeding tier boundaries are blocked before quote issuance.
  - Institutional Tier 3 swaps exceeding ₹10,00,000 require multi-factor biometrics or WebAuthn/Passkey signatures.
- **Two-Phase Commit Balance Safety:**
  - Off-chain source funds are never debited without placing a verifiable cryptographic hold first.
  - Destination assets are never released to user accounts without guaranteed confirmation of source asset locking and Treasury reserve availability.
- **Audit Trails & Regulatory Traceability:**
  - Every issued quote, whether executed, expired, or rejected, is permanently recorded in `rfq_quotes` with millisecond timestamps and spread snapshots for SEBI algorithmic trading inspection and market surveillance audits.

---

## Acceptance Criteria

- [ ] High-speed RFQ service generates deterministic conversion quotes within $\le 5\text{ms}$ processing latency upon receiving `RequestQuote`.
- [ ] Ephemeral quote locks strictly expire and become unexecutable after exactly 5.00 seconds ($5000\text{ms}$) as enforced by Redis Lua scripts and server timestamps.
- [ ] Claimed quotes execute with **zero slippage**: the effective execution rate and payout amount in `ExecuteQuoteResponse` match the issued quote to the exact decimal digit.
- [ ] Platform fee is strictly invariant at 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) of the traded notional value, with zero additional hidden retail spread markup.
- [ ] Two-phase atomic settlement coordinates with Wallet Service (Prompt 203) to hold and transfer funds without ledger inconsistencies or orphan balances.
- [ ] Atomic DvP settlements involving tokenized digital securities trigger smart contract transfers on Hyperledger Besu under QBFT consensus and record the on-chain transaction hash.
- [ ] Redis Lua scripts guarantee that a quote lock can be claimed exactly once, preventing double-spend and duplicate execution races under concurrent replay attacks.
- [ ] Front-running and rate tampering attempts fail cryptographically via HMAC-SHA256 signature verification.
- [ ] Treasury inventory positions update in real time; net inventory breaches emit rebalancing trigger events to Kafka topic `rfq.treasury.hedge.v1`.
- [ ] User KYC tier limits and daily volume ceilings are strictly enforced prior to quote locking.
- [ ] All database queries compile without errors via `sqlc` or `sqlx` against PostgreSQL 16+.

---

## Suggested Order / Dependencies

- **Prerequisites:**
  - `103_api_design_standards.md`: gRPC and RESTful API conventions.
  - `104_event_schema_and_kafka_topic_standards.md`: Kafka topic naming and Protobuf schema registry.
  - `111_domain_model_core_entities.md`: Core asset, currency, and user domain definitions.
  - `112_idempotency_and_exactly_once_processing.md`: Distributed idempotency key standards.
  - `203_wallet_account_service.md`: Double-entry ledgers and financial balance holds.
  - `206_risk_and_margin_checks_service.md`: User tier validation and velocity limits.
  - `207_market_data_service.md`: Real-time L2 order book depth and tick feeds.
- **Parallel Tasks:**
  - `210_fee_and_realized_pnl_engine.md`: Platform fee distribution and revenue accounting.
  - `232_cbdc_digital_rupee_settlement_adapter.md`: eINR wholesale and retail bridge integration.
  - `258_national_best_bid_offer_nbbo_consolidated_tape_engine.md`: Consolidated reference pricing.
- **Downstream Blockers:**
  - `509_flutter_order_placement_and_trading_interface.md`: Mobile 1-click Instant Convert & Swap UI with countdown animation.
  - `603_web_trading_dashboard_and_order_entry.md`: Web Instant Swap modal and streaming quote chart.
