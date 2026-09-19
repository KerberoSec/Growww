# 266 - RWA Primary Token Launchpad & Dutch Auction Engine (Go)

## Purpose
In traditional Indian finance and real-world asset (RWA) capital markets, primary issuances of debt, equity, and asset-backed securities (such as non-convertible debentures [NCDs], real estate investment trusts [REITs], infrastructure investment trusts [InvITs], private credit debt pools, and structured commercial paper) have historically been constrained by manual book-building, opaque allocation practices, high investment banking underwriting fees, and multi-day settlement cycles. Furthermore, fixed-price private placements frequently suffer from mispricing: assets are either underpriced (depriving issuers of fair cost of capital) or overpriced (resulting in failed capital calls and unsold inventory).

Under the regulatory frameworks established by the Securities and Exchange Board of India (SEBI) under the SEBI (Issue and Listing of Non-Convertible Securities) Regulations and SEBI (Alternative Investment Funds) Regulations, alongside Section 42 of the Indian Companies Act, 2013 (Private Placement Offer and Allocation Rules), capital mobilization requires strict regulatory compliance, accredited investor limits, fair and transparent price discovery, and segregated escrow accounting.

The **RWA Primary Token Launchpad & Dutch Auction Engine** (`services/launchpad-engine`) is an enterprise-grade, high-throughput issuance and auction settlement service built in Go 1.22+. Designed in accordance with ADR-0044 (Primary Issuance & Dutch Auction Architecture) and RUNBOOK-32 (Launchpad Settlement & Clearing Operations), the engine orchestrates the end-to-end lifecycle of primary asset tokenization. It executes mathematically verifiable linear Dutch auction price descent curves, determines uniform clearing prices through cumulative demand aggregation, allocates over-subscribed issues via deterministic pro-rata distribution, dispatches instantaneous escrow refunds, and provisions programmatic linear vesting schedules via permissioned smart contracts (`DutchAuction.sol` and `LinearVestingVault.sol`) deployed on the National Blockchain Stock Exchange (NBSE) Hyperledger Besu network under QBFT consensus. All capital raised through the platform automatically assesses the platform's universal flat 0.00% transaction fee (No fee at all) (governed dynamically by FeeController.sol (0.00% at launch)).

## What You Are Building
A mission-critical, ultra-reliable primary market issuance microservice (`services/launchpad-engine`) implemented in Go 1.22+, comprising the following core architectural subsystems:

- **Linear Dutch Auction Price Descent Engine:** Evaluates real-time offering prices over a time-bounded window ($[T_{\text{start}}, T_{\text{end}}]$) decaying linearly from an initial ceiling price ($P_{\text{start}}$) to a floor reserve price ($P_{\text{floor}}$). Emits real-time price tick events and governs auction lifecycle state transitions (`SCHEDULED`, `BIDDING_OPEN`, `BIDDING_CLOSED`, `PRICE_DISCOVERY`, `SETTLING`, `FINALIZED`, `CANCELLED`).
- **Anti-Front-Running Commit-Reveal Bid Book Aggregator:** Enforces a two-phase commit-reveal bidding mechanism to eliminate front-running, bid sniping, and miner/validator extractable value (MEV). During the commit phase, investors submit salted cryptographic hashes of their bid parameters backed by 100% locked escrowed cash collateral verified by the Wallet Account Service (Prompt 203). During the reveal phase, submitted nonces are unsealed, validated against on-chain/off-chain commitments, and compiled into an in-memory priority bid book.
- **Uniform Clearing Price Discovery Core:** Aggregates revealed bids into an ordered cumulative demand function against the total primary token supply ($S_{\text{total}}$). Dynamically determines the uniform clearing price ($P^*$) where aggregate demand equals or exceeds supply, guaranteeing that every successful bidder pays the exact same uniform clearing price regardless of their initial maximum bid.
- **Pro-Rata Over-Subscription Allocation Engine:** In events where cumulative demand at the uniform clearing price exceeds remaining unallocated tokens, the engine executes deterministic pro-rata scaling with integer micro-token rounding algorithms, ensuring zero token leakage and complete distribution fairness.
- **Automated Escrow Refund & DvP Settlement Dispatcher:** Integrates synchronously with the Multi-Asset Wallet & Ledger Service (Prompt 203) to execute atomic Delivery-versus-Payment (DvP). Deducts final settled capital at clearing price $P^*$, instantly releases excess escrowed funds (the spread between bid price and clearing price, plus unfilled portions) back to client wallets, and deposits allocated tokens into custody.
- **Linear Vesting Schedule Orchestrator:** Manages on-chain token lockup and stream release terms for issued RWA securities, pushing deterministic distribution schedules with customizable cliff intervals and continuous linear block-by-block vesting curves to `LinearVestingVault.sol`.
- **Statutory Companies Act & SEBI Compliance Guard:** Tracks investor participation across primary private placements, enforcing the statutory hard ceiling of at most 200 distinct eligible investors per financial year per issue under Section 42 of the Companies Act, 2013, while verifying accredited investor status via KYC Service (Prompt 202/005) or zero-knowledge accreditation proofs (Prompt 338).
- **Hyperledger Besu QBFT Settlement Relayer:** Bridges finalized auction parameters, Merkle allocation roots, and token mint instructions to smart contracts (`DutchAuction.sol`, `LinearVestingVault.sol`, and `TokenIssuanceService` Prompt 303) on the permissioned ledger with 1:1 physical custody backing and zero Personally Identifiable Information (PII).
- **Universal 0.00% (No fee at all) Platform Fee Ledger:** Automatically calculates, logs, and splits the mandatory 0.00% (Zero Fee) platform fee across primary market gross capital proceeds (0.00% fee at launch; future fee parameters governed by FeeController.sol).

## Scope Boundaries
- **In Scope:**
  - Automated state machine for primary token issuances across Dutch auctions and fixed-price subscription pools.
  - Linear Dutch auction price descent calculation:
    $$P(t) = P_{\text{start}} - \left( \frac{P_{\text{start}} - P_{\text{floor}}}{T_{\text{end}} - T_{\text{start}}} \right) \cdot (t - T_{\text{start}})$$
  - Cryptographic commit-reveal bidding scheme with salted SHA-256 commitments and zero-leakage reveal validation.
  - 100% pre-funded cash escrow verification and atomic hold reservations via Wallet Account Service (Prompt 203).
  - Uniform clearing price calculation ($P^*$) and clearing volume resolution.
  - Deterministic pro-rata over-subscription rationing with micro-unit rounding conservation.
  - Instant automated dispatch of excess escrow refunds and non-winning bid releases.
  - Linear vesting schedule creation, tracking cliff periods, and streaming distribution configurations.
  - Enforcement of the Companies Act Section 42 private placement 200-investor cap per fiscal year.
  - Off-chain and on-chain coordination with `DutchAuction.sol` and `LinearVestingVault.sol` on Hyperledger Besu.
  - Assessing and ledgering the universal 0.00% (No fee at all) platform transaction fee (0.00% fee launch policy revenue split).
  - High-throughput gRPC APIs and Kafka event streaming for real-time order books and launchpad status.
- **Out of Scope / Handled Elsewhere:**
  - Continuous secondary market order matching and central limit order book (CLOB) trading (handled in Prompt 205 Order Matching Engine).
  - Fiat banking rail integrations, RTGS/NEFT/UPI collection gateways (handled in Prompt 203 Wallet Service and Prompt 214 Foreign Investor Funding Service).
  - Physical asset appraisal, title deeds legal due diligence, and SPV corporate structuring (handled in Prompt 615 Web Asset Issuer Portal).
  - Physical depository demat corporate action execution and NSDL/CDSL cross-listing (handled in Prompt 213 Depository Adapter).
  - Secondary market portfolio margin scanning and dynamic haircut valuation (handled in Prompt 206 Risk Engine and Prompt 264 Dynamic Collateral Haircut Engine).
  - On-chain proof of reserve physical vault auditing (handled in Prompt 308 / Prompt 327).

## Technology to Use
- **Core Language & Runtime:** **Go 1.22+** leveraging structured concurrency, lightweight goroutines, bounded worker pools, and low-latency garbage collection optimized for microsecond financial computation.
- **Database & Persistence:** **PostgreSQL 16+** with `jackc/pgx/v5` connection pooling, strict ACID transactions, transaction-level advisory locks (`pg_advisory_xact_lock`), and schema table partitioning by auction ID.
- **In-Memory Cache & Distributed Mutex:** **Redis 7.2+ Cluster** with Redis Sentinel failover for sub-millisecond price descent caching, bid book staging, and distributed locking (`Redlock`) to guarantee single-flight settlement execution.
- **Message Broker & Event Streaming:** **Apache Kafka 3.7+** (`confluent-kafka-go` or `segmentio/kafka-go`) with strictly ordered, partitioned topics for auction state transitions, bid commit-reveal streams, and refund dispatch events with guaranteed `at-least-once` delivery.
- **Inter-Service Communication:** **gRPC / Protocol Buffers v3** with HTTP/2 transport and mutual TLS (mTLS) for sub-millisecond inter-service RPC invocations.
- **Blockchain Client Integration:** `github.com/ethereum/go-ethereum/ethclient` interfacing over JSON-RPC/WebSockets with Hyperledger Besu QBFT permissioned nodes.
- **Precision Numerical Mathematics:** Go standard `math/big` arbitrary precision integer and rational arithmetic for micro-token allocation, paise-level fiat settlement, and exact pro-rata rounding without floating-point drift.
- **Cryptographic Primitives:** Standard Go `crypto/sha256`, `crypto/hmac`, and `crypto/subtle` for timing-attack-resistant commitment verification and salted identity anonymization.

## Backend / Infra Touchpoints
- **PostgreSQL Database Tables:**
  - `launchpad_auctions`: Master auction configuration, asset ISIN, start/floor prices, token supply, timeline, and lifecycle state.
  - `launchpad_commitments`: Encrypted / hashed bid commitments submitted during the commit window with escrow hold IDs.
  - `launchpad_bids`: Revealed bid records containing plaintext price, quantity, timestamp, investor ID, and eligibility status.
  - `launchpad_allocations`: Final allocation records detailing awarded tokens, uniform clearing price, total cost, and pro-rata factor.
  - `launchpad_refunds`: Detailed log of excess escrow and non-winning bid refund transactions dispatched to user wallets.
  - `launchpad_vesting_schedules`: On-chain and off-chain vesting terms, cliff timestamps, total periods, and claim tracking.
  - `launchpad_investor_caps`: Section 42 tally tracking unique investor counts per issuer and financial year.
  - `launchpad_platform_fee_ledgers`: Records the 0.00% (Zero Fee) platform fee assessed on primary proceeds with 0.00% fee launch policy distribution.
  - `launchpad_audit_trail_ledger`: Immutable append-only log of every auction state change, bid submission, and settlement calculation with SHA-256 hash chaining.
- **Redis Cache Keys:**
  - `launchpad:auction:price:{auction_id}`: Real-time current descending auction price (TTL: 1s refresh).
  - `launchpad:auction:state:{auction_id}`: Current lifecycle state of the auction.
  - `launchpad:auction:bid_count:{auction_id}`: Monotonically increasing atomic counter of submitted commitments.
  - `lock:launchpad:settlement:{auction_id}`: Distributed mutex ensuring single-flight clearing execution.
  - `lock:launchpad:investor:{auction_id}:{investor_id}`: Mutex preventing concurrent duplicate bids from the same investor.
- **Kafka Topics:**
  - **Consumes:**
    - `wallet.escrow_locked.v1`: Ingests cash escrow reservation confirmations from Wallet Account Service (Prompt 203).
    - `identity.accreditation_verified.v1`: Ingests accredited investor approval events from KYC Service (Prompt 202/005).
    - `blockchain.auction_settled.v1`: Ingests on-chain settlement receipts from Hyperledger Besu relayer.
  - **Publishes:**
    - `launchpad.auction_created.v1`: Emitted upon validation and scheduling of a new primary issuance.
    - `launchpad.price_tick.v1`: Periodic broadcast of descending Dutch auction price ticks.
    - `launchpad.bidding_closed.v1`: Emitted when auction window expires, initiating price discovery.
    - `launchpad.clearing_price_discovered.v1`: Emitted upon determination of uniform clearing price $P^*$.
    - `launchpad.tokens_allocated.v1`: Emitted with final pro-rata allocation distributions.
    - `launchpad.refunds_dispatched.v1`: Emitted when excess escrow release requests are dispatched.
    - `launchpad.vesting_vault_initialized.v1`: Emitted when linear vesting vault parameters are configured.
    - `launchpad.fee_assessed.v1`: Emitted to record the universal 0.00% (No fee at all) platform fee.
- **Internal Microservices:**
  - **Multi-Asset Wallet & Ledger Service (Prompt 203):** Manages pre-funded escrow locks, final settlement debits, and instant unallocated refunds.
  - **Custodian & Depository Integration Service (Prompt 213):** Verifies that underlying physical RWA title deeds, mortgages, or debenture trust receipts are locked 1:1 in custodial vaults before auction launch.
  - **Token Issuance Smart Contract Service (Prompt 303):** Mints or pre-deposits ERC-3643 digital security tokens into the `DutchAuction.sol` smart contract.
  - **KYC & Investor Identity Service (Prompt 202 / 005 / 338):** Validates investor onboarding, Section 42 annual cap eligibility, and Groth16 ZK-SNARK accreditation claims.
  - **Primary Market Order Routing & Clearing Bridge (Prompt 331):** Synchronizes primary issuance completion with secondary market listing queues.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Immediate Finality:** Deployed on the Growww National Blockchain Stock Exchange (NBSE) permissioned Hyperledger Besu consortium network operating under Istanbul QBFT (Quorum Byzantine Fault Tolerance) consensus with 2-second block times and deterministic zero-reorganization finality.
- **On-Chain Smart Contract Interfaces:**
  - `DutchAuction.sol`: The canonical on-chain auction coordinator contract. Maintains immutable auction parameters (token address, total offering supply, ceiling price, floor price, start/end block timestamps). Receives the off-chain settled uniform clearing price ($P^*$) along with an aggregated Merkle root representing all investor allocations. Verifies that total allocated tokens do not exceed escrowed primary tokens.
  - `LinearVestingVault.sol`: Receives primary token allocations for issues with lockup restrictions. Holds tokens in escrow and releases them continuously on a block-by-block linear trajectory after an initial cliff period. Investors invoke `claimVestedTokens()` on-chain, with transfers governed by the underlying ERC-3643 `ComplianceModule.sol`.
- **1:1 Physical Custody Backing Invariant:**
  - Smart contracts strictly prohibit the deployment or initialization of an auction without cryptographic confirmation from the Custodian Depository Adapter (Prompt 213) certifying that underlying physical assets (e.g., registered title deeds held in escrow, charge registered with the Registrar of Companies [RoC], or physical debenture certificates) are locked with 1:1 parity in an institutional depository vault.
- **Zero Personally Identifiable Information (Zero PII):**
  - No investor names, PAN cards, Demat BOIDs, or bank account numbers are ever published on-chain.
  - Investor addresses are mapped using pseudonymous on-chain identities registered in the ERC-3643 Identity Registry, with off-chain records indexed via salted hashes:
    $$\text{InvestorTag} = \text{HMAC-SHA256}(\text{InvestorID}, \text{PlatformSalt})$$
- **Cryptographic Settlement Proofs & Commitments:**
  - In the commit phase, commitment hashes ($\text{CommitmentHash} = \text{SHA256}(P \parallel Q \parallel \text{Salt} \parallel \text{InvestorAddress})$) can optionally be notarized on-chain or stored off-chain in Redis/PostgreSQL with Merkle roots anchored to `DutchAuction.sol` to guarantee absolute tamper-proof auditability.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Architecture:** Initialize Go 1.22+ module `services/launchpad-engine` following standard domain-driven design directory conventions (`cmd/server`, `internal/auction`, `internal/bidding`, `internal/pricing`, `internal/allocation`, `internal/vesting`, `internal/compliance`, `internal/blockchain`, `internal/storage`).
2. **Define Protobuf Service Contracts:** Author `proto/growww/launchpad/v1/launchpad_service.proto` defining RPC methods: `CreateAuction`, `GetAuctionStatus`, `SubmitBidCommitment`, `RevealBid`, `StreamAuctionPriceTicks`, `CalculateClearingPrice`, `FinalizeAuctionSettlement`, and `GetVestingSchedule`.
3. **Generate Go Stubs & Validation Hooks:** Compile protocol buffers using `protoc` with `protoc-gen-go`, `protoc-gen-go-grpc`, and `protoc-gen-validate` to enforce field validation rules (e.g., positive quantities, valid ISIN strings, positive timestamps).
4. **Author PostgreSQL Schema & Migrations:** Implement database migration scripts under `migrations/` defining tables for `launchpad_auctions`, `launchpad_commitments`, `launchpad_bids`, `launchpad_allocations`, `launchpad_refunds`, `launchpad_vesting_schedules`, and fee ledgers with foreign key constraints, check constraints, and performance indexes.
5. **Implement Auction State Machine & Lifecycle Worker:** Develop a thread-safe auction finite state machine (FSM) transitioning states: `SCHEDULED` -> `BIDDING_OPEN` -> `BIDDING_CLOSED` -> `PRICE_DISCOVERY` -> `SETTLING` -> `FINALIZED` (or `CANCELLED`). Implement an asynchronous background ticker that evaluates start/end timestamps and manages state transitions.
6. **Build Linear Dutch Auction Price Descent Calculator:** Implement high-precision price descent evaluation in `internal/pricing/dutch.go`. Calculate the continuous price $P(t)$ from ceiling $P_{\text{start}}$ to floor $P_{\text{floor}}$ across duration $T_{\text{end}} - T_{\text{start}}$, broadcasting descending price ticks to Redis pub/sub and Kafka topic `launchpad.price_tick.v1`.
7. **Implement Commit-Reveal Bidding Pipeline:**
   - *Commit Handler:* Accepts commitment hash, checks eligibility, requests cash escrow lock from Wallet Account Service (Prompt 203) for maximum potential bid liability ($Q \cdot P_{\text{start}}$), and persists commitment.
   - *Reveal Handler:* Ingests revealed parameters ($P$, $Q$, $\text{Salt}$), recomputes SHA-256 hash, validates against stored commitment, confirms timestamp falls within reveal window, and stores revealed bid in the priority bid book.
8. **Implement Companies Act Section 42 Investor Ceiling Guard:** Build the statutory compliance validator in `internal/compliance/private_placement.go`. Verify that the auction does not exceed 200 distinct non-QIB participating investors within a single financial year, rejecting commitments once the regulatory limit is attained.
9. **Build Uniform Clearing Price Discovery Algorithm:** Implement the price discovery logic in `internal/pricing/clearing.go`. Sort all revealed bids in descending price order. Construct the cumulative demand curve:
   $$D(P) = \sum_{i: P_i \ge P} Q_i$$
   Locate the uniform clearing price $P^*$ where $D(P^*) \ge S_{\text{total}}$. If cumulative demand at the floor price $P_{\text{floor}}$ is less than $S_{\text{total}}$, determine whether the auction clears at $P_{\text{floor}}$ (under-subscribed clearing) or fails based on the issue's minimum subscription threshold.
10. **Implement Pro-Rata Over-Subscription Allocation Core:** Build the allocation engine in `internal/allocation/prorata.go`. For bids with $P_i > P^*$, allocate 100% of requested quantity $Q_i$. For marginal bids where $P_i = P^*$, allocate tokens proportionally according to:
    $$\text{Ratio} = \frac{S_{\text{remaining}}}{\sum_{j \in \text{Marginal}} Q_j}, \quad A_i = \lfloor Q_i \cdot \text{Ratio} \rfloor$$
    Distribute any fractional remainder tokens deterministically using largest-remainder or timestamp priority, ensuring zero unallocated token leakage.
11. **Implement Automated Escrow Refund Dispatcher:** Develop the refund coordinator in `internal/auction/refund.go`. Synchronously call Wallet Account Service (Prompt 203) to:
    - Debit settled amount: $\text{SettledAmount}_i = A_i \cdot P^*$.
    - Release excess locked funds: $\text{Refund}_i = \text{EscrowLocked}_i - \text{SettledAmount}_i$.
    - Fully release escrow for all unsuccessful bids ($P_i < P^*$).
12. **Build Linear Vesting Schedule Manager:** Implement vesting configuration logic in `internal/vesting/manager.go`. For issuances with lockups, calculate cliff dates and streaming release durations, construct initial balances, and generate cryptographic Merkle trees representing investor vesting allocations.
13. **Implement Hyperledger Besu Blockchain Relayer:** Construct the Besu transaction relayer in `internal/blockchain/relayer.go` using `go-ethereum/ethclient`. Assemble and dispatch transactions invoking `settleAuction(uint256 clearingPrice, bytes32 allocationMerkleRoot)` on `DutchAuction.sol` and initialize `LinearVestingVault.sol` with root hashes, signed with an enterprise HSM key.
14. **Integrate Universal 0.00% (No fee at all) Platform Fee Splitter:** Implement platform fee accounting in `internal/auction/fee.go`. Assess the 0.00% fee (No fee at all) on total primary capital raised ($S_{\text{allocated}} \cdot P^*$), and record the deterministic split into PostgreSQL: Treasury reserve, Core SGF, and Investor Protection Fund per FeeController governance.
15. **Construct Comprehensive Test & Verification Suite:** Write unit tests for price descent formulas and commit-reveal integrity; write table-driven test cases for pro-rata allocation edge cases; write integration tests with mock Wallet Service and Besu QBFT nodes; execute concurrency chaos tests validating zero duplicate allocations and zero double refunds under high load.

## Interfaces / Contracts

### Protobuf Service Contract (`proto/growww/launchpad/v1/launchpad_service.proto`)
```protobuf
syntax = "proto3";

package growww.launchpad.v1;

option go_package = "github.com/growww/proto/gen/go/launchpad/v1;launchpadv1";

import "google/protobuf/timestamp.proto";

enum AuctionType {
  AUCTION_TYPE_UNSPECIFIED = 0;
  AUCTION_TYPE_DUTCH_DESCENDING = 1;
  AUCTION_TYPE_FIXED_PRICE_OVERSUBSCRIPTION = 2;
}

enum AuctionStatus {
  AUCTION_STATUS_UNSPECIFIED = 0;
  AUCTION_STATUS_SCHEDULED = 1;
  AUCTION_STATUS_BIDDING_OPEN = 2;
  AUCTION_STATUS_BIDDING_CLOSED = 3;
  AUCTION_STATUS_PRICE_DISCOVERY = 4;
  AUCTION_STATUS_SETTLING = 5;
  AUCTION_STATUS_FINALIZED = 6;
  AUCTION_STATUS_CANCELLED = 7;
}

enum AssetCategory {
  ASSET_CATEGORY_UNSPECIFIED = 0;
  ASSET_CATEGORY_REAL_ESTATE = 1;
  ASSET_CATEGORY_PRIVATE_CREDIT = 2;
  ASSET_CATEGORY_CORPORATE_DEBT = 3;
  ASSET_CATEGORY_INFRASTRUCTURE = 4;
}

message CreateAuctionRequest {
  string isin = 1;
  string asset_symbol = 2;
  AssetCategory asset_category = 3;
  AuctionType auction_type = 4;
  uint64 total_token_supply = 5;
  uint64 start_price_paise = 6;
  uint64 floor_price_paise = 7;
  uint64 min_bid_quantity = 8;
  uint64 max_bid_quantity_per_investor = 9;
  google.protobuf.Timestamp bidding_start_time = 10;
  google.protobuf.Timestamp bidding_end_time = 11;
  google.protobuf.Timestamp reveal_end_time = 12;
  bool vesting_enabled = 13;
  uint32 cliff_duration_days = 14;
  uint32 total_vesting_duration_days = 15;
  string issuer_account_id = 16;
  string token_contract_address = 17;
}

message CreateAuctionResponse {
  string auction_id = 1;
  AuctionStatus status = 2;
  google.protobuf.Timestamp created_at = 3;
}

message GetAuctionStatusRequest {
  string auction_id = 1;
}

message GetAuctionStatusResponse {
  string auction_id = 1;
  string isin = 2;
  AuctionStatus status = 3;
  uint64 current_price_paise = 4;
  uint64 total_token_supply = 5;
  uint64 total_bids_submitted = 6;
  uint64 total_quantity_revealed = 7;
  uint64 uniform_clearing_price_paise = 8;
  google.protobuf.Timestamp bidding_start_time = 9;
  google.protobuf.Timestamp bidding_end_time = 10;
  google.protobuf.Timestamp reveal_end_time = 11;
  bool is_over_subscribed = 12;
}

message SubmitBidCommitmentRequest {
  string auction_id = 1;
  string investor_account_id = 2;
  string commitment_hash = 3; // SHA-256(price_paise || quantity || salt || investor_account_id)
  uint64 max_liability_paise = 4; // Escrow lock amount (quantity * start_price_paise)
  string client_ip = 5;
}

message SubmitBidCommitmentResponse {
  string commitment_id = 1;
  string escrow_lock_id = 2;
  google.protobuf.Timestamp committed_at = 3;
  bool accepted = 4;
}

message RevealBidRequest {
  string auction_id = 1;
  string investor_account_id = 2;
  string commitment_id = 3;
  uint64 bid_price_paise = 4;
  uint64 bid_quantity = 5;
  string salt = 6;
}

message RevealBidResponse {
  string bid_id = 1;
  bool is_valid = 2;
  string rejection_reason = 3;
  google.protobuf.Timestamp revealed_at = 4;
}

message StreamAuctionPriceTicksRequest {
  string auction_id = 1;
}

message PriceTick {
  string auction_id = 1;
  uint64 current_price_paise = 2;
  google.protobuf.Timestamp timestamp = 3;
  uint64 elapsed_seconds = 4;
  uint64 remaining_seconds = 5;
}

message FinalizeAuctionSettlementRequest {
  string auction_id = 1;
}

message FinalizeAuctionSettlementResponse {
  string auction_id = 1;
  uint64 uniform_clearing_price_paise = 2;
  uint64 total_tokens_allocated = 3;
  uint64 gross_capital_raised_paise = 4;
  uint64 total_refunds_paise = 5;
  uint64 platform_fee_paise = 6; // 0.00% (Zero Fee) platform fee
  uint32 total_successful_investors = 7;
  string on_chain_settlement_tx_hash = 8;
  AuctionStatus status = 9;
}

message GetVestingScheduleRequest {
  string auction_id = 1;
  string investor_account_id = 2;
}

message GetVestingScheduleResponse {
  string auction_id = 1;
  string investor_account_id = 2;
  uint64 total_vesting_tokens = 3;
  uint64 claimed_tokens = 4;
  uint64 claimable_tokens = 5;
  google.protobuf.Timestamp cliff_end_time = 6;
  google.protobuf.Timestamp vesting_end_time = 7;
  string vesting_vault_contract_address = 8;
}

service LaunchpadService {
  rpc CreateAuction (CreateAuctionRequest) returns (CreateAuctionResponse);
  rpc GetAuctionStatus (GetAuctionStatusRequest) returns (GetAuctionStatusResponse);
  rpc SubmitBidCommitment (SubmitBidCommitmentRequest) returns (SubmitBidCommitmentResponse);
  rpc RevealBid (RevealBidRequest) returns (RevealBidResponse);
  rpc StreamAuctionPriceTicks (StreamAuctionPriceTicksRequest) returns (stream PriceTick);
  rpc FinalizeAuctionSettlement (FinalizeAuctionSettlementRequest) returns (FinalizeAuctionSettlementResponse);
  rpc GetVestingSchedule (GetVestingScheduleRequest) returns (GetVestingScheduleResponse);
}
```

### PostgreSQL Database Schema (`services/launchpad-engine/migrations/001_initial_schema.sql`)
```sql
-- PostgreSQL Migration: RWA Primary Token Launchpad & Dutch Auction Engine Schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enums
CREATE TYPE auction_type_enum AS ENUM (
    'DUTCH_DESCENDING',
    'FIXED_PRICE_OVERSUBSCRIPTION'
);

CREATE TYPE auction_status_enum AS ENUM (
    'SCHEDULED',
    'BIDDING_OPEN',
    'BIDDING_CLOSED',
    'PRICE_DISCOVERY',
    'SETTLING',
    'FINALIZED',
    'CANCELLED'
);

CREATE TYPE asset_category_enum AS ENUM (
    'REAL_ESTATE',
    'PRIVATE_CREDIT',
    'CORPORATE_DEBT',
    'INFRASTRUCTURE'
);

-- Master Auction Table
CREATE TABLE launchpad_auctions (
    auction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    asset_symbol VARCHAR(32) NOT NULL,
    asset_category asset_category_enum NOT NULL,
    auction_type auction_type_enum NOT NULL DEFAULT 'DUTCH_DESCENDING',
    status auction_status_enum NOT NULL DEFAULT 'SCHEDULED',
    total_token_supply BIGINT NOT NULL CHECK (total_token_supply > 0),
    allocated_token_supply BIGINT NOT NULL DEFAULT 0,
    start_price_paise BIGINT NOT NULL CHECK (start_price_paise > 0),
    floor_price_paise BIGINT NOT NULL CHECK (floor_price_paise > 0 AND floor_price_paise <= start_price_paise),
    current_price_paise BIGINT NOT NULL,
    uniform_clearing_price_paise BIGINT,
    min_bid_quantity BIGINT NOT NULL DEFAULT 1,
    max_bid_quantity_per_investor BIGINT NOT NULL,
    bidding_start_time TIMESTAMPTZ NOT NULL,
    bidding_end_time TIMESTAMPTZ NOT NULL CHECK (bidding_end_time > bidding_start_time),
    reveal_end_time TIMESTAMPTZ NOT NULL CHECK (reveal_end_time >= bidding_end_time),
    vesting_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    cliff_duration_days INT NOT NULL DEFAULT 0,
    total_vesting_duration_days INT NOT NULL DEFAULT 0,
    issuer_account_id VARCHAR(64) NOT NULL,
    token_contract_address VARCHAR(42) NOT NULL,
    dutch_auction_contract_address VARCHAR(42),
    vesting_vault_contract_address VARCHAR(42),
    settlement_tx_hash VARCHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Bid Commitments Table (Phase 1: Commit)
CREATE TABLE launchpad_commitments (
    commitment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id UUID NOT NULL REFERENCES launchpad_auctions(auction_id) ON DELETE CASCADE,
    investor_account_id VARCHAR(64) NOT NULL,
    hashed_investor_boid VARCHAR(64) NOT NULL,
    commitment_hash VARCHAR(64) NOT NULL,
    escrow_lock_id VARCHAR(64) NOT NULL,
    max_liability_paise BIGINT NOT NULL CHECK (max_liability_paise > 0),
    is_revealed BOOLEAN NOT NULL DEFAULT FALSE,
    committed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(auction_id, investor_account_id)
);

-- Revealed Bids Table (Phase 2: Reveal)
CREATE TABLE launchpad_bids (
    bid_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    commitment_id UUID NOT NULL REFERENCES launchpad_commitments(commitment_id),
    auction_id UUID NOT NULL REFERENCES launchpad_auctions(auction_id) ON DELETE CASCADE,
    investor_account_id VARCHAR(64) NOT NULL,
    bid_price_paise BIGINT NOT NULL CHECK (bid_price_paise > 0),
    bid_quantity BIGINT NOT NULL CHECK (bid_quantity > 0),
    salt VARCHAR(64) NOT NULL,
    is_valid BOOLEAN NOT NULL DEFAULT TRUE,
    is_successful BOOLEAN NOT NULL DEFAULT FALSE,
    allocated_quantity BIGINT NOT NULL DEFAULT 0,
    revealed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Final Allocations Table
CREATE TABLE launchpad_allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id UUID NOT NULL REFERENCES launchpad_auctions(auction_id) ON DELETE CASCADE,
    investor_account_id VARCHAR(64) NOT NULL,
    bid_id UUID NOT NULL REFERENCES launchpad_bids(bid_id),
    allocated_quantity BIGINT NOT NULL CHECK (allocated_quantity >= 0),
    clearing_price_paise BIGINT NOT NULL CHECK (clearing_price_paise > 0),
    total_cost_paise BIGINT NOT NULL CHECK (total_cost_paise >= 0),
    is_pro_rata BOOLEAN NOT NULL DEFAULT FALSE,
    pro_rata_factor NUMERIC(10, 8) NOT NULL DEFAULT 1.00000000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Refunds Ledger Table
CREATE TABLE launchpad_refunds (
    refund_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id UUID NOT NULL REFERENCES launchpad_auctions(auction_id),
    investor_account_id VARCHAR(64) NOT NULL,
    escrow_lock_id VARCHAR(64) NOT NULL,
    original_escrow_paise BIGINT NOT NULL,
    settled_amount_paise BIGINT NOT NULL,
    refund_amount_paise BIGINT NOT NULL CHECK (refund_amount_paise >= 0),
    refund_tx_id VARCHAR(64),
    status VARCHAR(32) NOT NULL DEFAULT 'COMPLETED',
    dispatched_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Linear Vesting Schedules Table
CREATE TABLE launchpad_vesting_schedules (
    schedule_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id UUID NOT NULL REFERENCES launchpad_auctions(auction_id),
    investor_account_id VARCHAR(64) NOT NULL,
    total_tokens BIGINT NOT NULL CHECK (total_tokens > 0),
    claimed_tokens BIGINT NOT NULL DEFAULT 0,
    cliff_end_time TIMESTAMPTZ NOT NULL,
    vesting_end_time TIMESTAMPTZ NOT NULL CHECK (vesting_end_time > cliff_end_time),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(auction_id, investor_account_id)
);

-- Companies Act Section 42 Investor Ceiling Tracker
CREATE TABLE launchpad_investor_caps (
    cap_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer_account_id VARCHAR(64) NOT NULL,
    financial_year VARCHAR(9) NOT NULL, -- e.g., '2026-2027'
    auction_id UUID NOT NULL REFERENCES launchpad_auctions(auction_id),
    investor_count INT NOT NULL DEFAULT 0 CHECK (investor_count <= 200),
    is_capped BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(issuer_account_id, financial_year, auction_id)
);

-- Platform fixed strictly 0.00% fee for all (No fee at all for Maker and Taker)) Ledger (0.00% fee at launch; future fee parameters governed by FeeController.sol)
CREATE TABLE launchpad_platform_fee_ledgers (
    fee_ledger_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_id UUID NOT NULL REFERENCES launchpad_auctions(auction_id),
    gross_proceeds_paise BIGINT NOT NULL,
    fee_paise BIGINT NOT NULL, -- 0.00% (Zero Fee) of gross proceeds
    treasury_split_paise BIGINT NOT NULL, -- Governed by FeeController (0.00% at launch)
    sgf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    ipf_split_paise BIGINT NOT NULL,      -- Governed by FeeController (0.00% at launch)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Append-Only Cryptographic Audit Trail
CREATE TABLE launchpad_audit_trail_ledger (
    audit_id BIGSERIAL PRIMARY KEY,
    auction_id UUID NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    actor_id VARCHAR(64) NOT NULL,
    prev_hash VARCHAR(64) NOT NULL,
    curr_hash VARCHAR(64) NOT NULL,
    payload_json JSONB NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance Indexes
CREATE INDEX idx_auctions_status_timeline ON launchpad_auctions(status, bidding_start_time, bidding_end_time);
CREATE INDEX idx_commitments_auction ON launchpad_commitments(auction_id, is_revealed);
CREATE INDEX idx_bids_auction_price ON launchpad_bids(auction_id, bid_price_paise DESC, revealed_at ASC);
CREATE INDEX idx_allocations_auction_investor ON launchpad_allocations(auction_id, investor_account_id);
CREATE INDEX idx_refunds_auction ON launchpad_refunds(auction_id, status);
CREATE INDEX idx_vesting_lookup ON launchpad_vesting_schedules(auction_id, investor_account_id);
CREATE INDEX idx_investor_caps_lookup ON launchpad_investor_caps(issuer_account_id, financial_year);
```

## Security & Compliance Notes
- **Companies Act, 2013 (Section 42 & Rule 14) Compliance:**
  - Strict compliance with Section 42 of the Indian Companies Act, 2013, governing private placements: a private placement of securities cannot be extended to more than 200 persons in aggregate across a single financial year (excluding Qualified Institutional Buyers [QIBs] and employees offered securities under employee stock option schemes).
  - The engine tracks every participating non-QIB investor in `launchpad_investor_caps`. If an issuance attempts to accept commitments exceeding the 200-investor threshold, the engine immediately rejects subsequent non-QIB commitments, preventing the issuance from being deemed an illegal unlisted public offer.
- **SEBI ICDR & Issue of Debt Securities Regulations:**
  - Compliance with SEBI (Issue of Capital and Disclosure Requirements) and SEBI (Issue and Listing of Non-Convertible Securities) guidelines.
  - Guarantees complete information symmetry by publishing descending price tick feeds in real time without preferential access.
  - Uniform clearing price model guarantees non-discriminatory pricing: no investor pays more than the final market-clearing equilibrium price.
- **Anti-Front-Running Commit-Reveal Cryptographic Protocol:**
  - In a standard open auction, predatory participants and malicious validators can observe pending transactions in the mempool and snipe prices or sandwich bids.
  - The engine strictly bifurcates bidding into two non-overlapping or distinct temporal phases:
    1. *Commit Phase:* Investor submits only $\text{SHA256}(P \parallel Q \parallel \text{Salt} \parallel \text{InvestorID})$. No pricing or quantity information is revealed to other participants or operators.
    2. *Reveal Phase:* Once the commit window closes, participants reveal their plain parameters. Any parameter that fails to hash to the original commitment is permanently discarded.
- **100% Pre-Funded Escrow Lockup Invariant:**
  - Phantom bidding and griefing attacks (where malicious actors submit large bids to artificially inflate clearing prices and subsequently default) are strictly prevented.
  - When submitting a commitment, the engine synchronously locks maximum potential liability ($Q \cdot P_{\text{start}}$) in the investor's cash wallet via the Wallet Account Service (Prompt 203). Unbacked bids are rejected upfront.
- **Zero Personally Identifiable Information (Zero PII):**
  - Depository Beneficiary Owner IDs (BOIDs), Permanent Account Numbers (PANs), and banking details are never stored in plaintext or published to Kafka/Blockchain. Hashed representations (`HMAC-SHA256`) with secure rotating salt keys are used for identity correlation.
- **Universal 0.00% (No fee at all) Platform Fee Integrity:**
  - Platform fee is deterministically deducted at source from gross capital proceeds prior to issuer payout: exactly 0.00% (Zero Fee) (0.00% fee).
  - Allocation is split according to platform governance: Platform Treasury, Core SGF, and IPF per FeeController governance, recorded with immutable transactional audit logs.

## Acceptance Criteria
- [ ] Service implements linear Dutch auction price descent curve adhering precisely to the formula $P(t) = P_{\text{start}} - \left( \frac{P_{\text{start}} - P_{\text{floor}}}{T_{\text{end}} - T_{\text{start}}} \right) \cdot (t - T_{\text{start}})$ without integer division truncation errors.
- [ ] Commit-reveal pipeline successfully accepts hashed commitments during the commit window, requires 100% escrow lockup from Wallet Account Service (Prompt 203), and rejects non-matching reveals or reveals submitted after the reveal deadline.
- [ ] Uniform clearing price ($P^*$) discovery correctly identifies the market equilibrium price where cumulative demand meets or exceeds offering supply across all test scenarios (under-subscribed, exact-fill, and over-subscribed).
- [ ] Over-subscription pro-rata allocation engine deterministically scales marginal bids with micro-token rounding conservation, guaranteeing that the sum of all individual allocations identically matches total allocated tokens with zero leakage.
- [ ] Section 42 private placement compliance guard strictly enforces the statutory cap of at most 200 non-QIB investors per financial year per issue, rejecting the 201st investor with an appropriate compliance error.
- [ ] Automated refund dispatcher calculates exact excess escrow releases ($\text{EscrowLocked} - (A_i \cdot P^*)$) and triggers instantaneous wallet credits for winning and out-of-the-money bidders within 1.5 seconds of auction finalization.
- [ ] Linear vesting engine successfully configures cliff periods and streaming release schedules, outputting valid Merkle roots to `LinearVestingVault.sol` on Hyperledger Besu.
- [ ] Blockchain relayer submits settlement payloads and allocation Merkle roots to `DutchAuction.sol` on Hyperledger Besu with confirmed QBFT finality.
- [ ] Universal 0.00% (No fee at all) platform fee is accurately assessed on gross proceeds and logged with the mandatory 0.00% fee at launch (governed by FeeController.sol) split.
- [ ] gRPC response times for `GetAuctionStatus` and `SubmitBidCommitment` remain sub-millisecond (< 800 microseconds p99 at 5,000 concurrent requests/sec).

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt 101 (System Architecture & Blueprint)
  - Prompt 102 (Bounded Contexts & Domain Models)
  - Prompt 203 (Multi-Asset Wallet & Ledger Service)
  - Prompt 213 (Custodian Depository Integration Service)
  - Prompt 301 (Permissioned Blockchain Evaluation & Selection)
  - Prompt 303 (Token Issuance Smart Contract)
- **Parallel Work:**
  - Prompt 206 (Real-Time Risk & Pre-Trade Margin Engine)
  - Prompt 331 (Primary Market Order Routing & Clearing Bridge)
  - Prompt 338 (Groth16 ZK-SNARK Investor Accreditation Verifier)
  - Prompt 615 (Web Asset Issuer & Tokenization Originator Portal)
- **Enables:**
  - Prompt 205 (Order Matching Engine - secondary market listing and continuous trading)
  - Prompt 208 (Trade Settlement Service)
  - Prompt 309 (Blockchain Event Indexing Service)
  - End-to-end institutional primary issuance and Dutch auction capital formation for tokenized real estate, private credit, and corporate debt.
