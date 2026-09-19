# 269 - P2P Fiat Escrow & Multi-Sig Dispute Arbitration Service (Go)

## Purpose
The P2P Fiat Escrow & Multi-Sig Dispute Arbitration Service facilitates compliant peer-to-peer bank transfers (Unified Payments Interface [UPI], Immediate Payment Service [IMPS], and National Electronic Funds Transfer [NEFT]) between verified investors on the Growww RWA Exchange. The service ensures trustless settlement by coupling off-chain domestic fiat transfers with automated collateral lockup in a segregated escrow account, payment receipt verification, and dual-control dispute arbitration in strict alignment with ADR-0039 and RUNBOOK-27.

Direct fiat-to-token and fiat-to-collateral transactions expose market participants to counterparty default, phantom payment claims, and fraudulent chargeback schemes. Under Securities and Exchange Board of India (SEBI) and Reserve Bank of India (RBI) regulatory frameworks, third-party payment transfers are illegal and violate anti-money laundering mandates. This service enforces complete identity isolation with strict KYC name invariants, automates time-bounded escrow state transitions, secures payment proof evidence in tamper-evident storage, and implements a 2-of-3 multi-signature cryptographic dispute resolution protocol across Hyperledger Besu and internal double-entry ledgers.

## What You Are Building
A high-throughput, fault-tolerant Go microservice (`services/p2p-escrow-service`) responsible for orchestrating the end-to-end P2P trading lifecycle. Deliverables include:
- **P2P Advertisement & Order Matching Desk:** Order book engine allowing verified makers to post buy/sell advertisements with fixed or floating pricing, order volume thresholds, and accepted payment rails (UPI VPA, IMPS IFSC/Account).
- **Escrow State Machine Engine:** Deterministic state machine governing trade states (`ORDER_CREATED`, `COLLATERAL_LOCKED`, `PAYMENT_MARKED_PAID`, `COMPLETED`, `DISPUTE_OPENED`, `DISPUTE_RESOLVED`, `CANCELLED_EXPIRED`).
- **Distributed Countdown Timer Manager:** Redis-backed countdown timers enforcing a 15-minute buyer payment window and a 30-minute seller confirmation window with automated expiration or escalation triggers.
- **Payment Proof Verification & Anti-Fraud Ingestion:** Secure pipeline for uploading bank transfer receipts, extracting Unique Transaction Reference (UTR) numbers, calculating cryptographic SHA-256 hashes, and generating perceptual image hashes (pHash) to detect forged or recycled screenshots.
- **Maker-Checker Dispute Arbitration Desk (RUNBOOK-27):** Operational queue integration enabling compliance officers to evaluate contested trades, perform automated Open Banking UTR verifications, and cast cryptographically signed arbitration votes.
- **On-Chain Multi-Sig Coordinator:** Ethereum ABI client interfacing with `P2PEscrow.sol` on Hyperledger Besu to atomically lock collateral tokens and execute 2-of-3 multi-signature releases upon cooperative confirmation or dispute settlement.
- **Key Deliverable Files:**
  - `services/p2p-escrow-service/cmd/server/main.go` - Microservice entry point and gRPC/REST server.
  - `services/p2p-escrow-service/internal/statemachine/escrow.go` - Escrow state machine transitions and guards.
  - `services/p2p-escrow-service/internal/order/matching.go` - P2P ad matching and order placement.
  - `services/p2p-escrow-service/internal/proof/hashing.go` - SHA-256 and pHash image duplicate detector.
  - `services/p2p-escrow-service/internal/timer/countdown.go` - Redis keyspace notification timer listener.
  - `services/p2p-escrow-service/internal/arbitration/maker_checker.go` - Dual-control arbitration workflow engine.
  - `services/p2p-escrow-service/internal/chain/escrow_client.go` - Hyperledger Besu `P2PEscrow.sol` Go client.
  - `proto/growww/p2p/v1/p2p_escrow_service.proto` - Protobuf gRPC interface contracts.

## Scope Boundaries
- **In Scope:**
  - P2P buy/sell advertisement lifecycle (posting, updating, pausing, cancellation).
  - Taker order intake, balance checks, and collateral locking in the P2P Escrow Account (Account 2800).
  - Real-time payment countdown timer management (15-minute window).
  - Buyer payment marking and seller receipt confirmation flows.
  - Bank payment receipt ingestion, metadata extraction (UTR, timestamp, amount), and screenshot perceptual hashing.
  - Dispute lifecycle management and Maker-Checker operational arbitration workflows (ADR-0039, RUNBOOK-27).
  - Hyperledger Besu on-chain lockup and 2-of-3 multi-signature dispute resolution execution.
  - Kafka event emission for user notifications, audit trails, and financial intelligence monitoring.
- **Out of Scope / Handled Elsewhere:**
  - Double-entry balance sheet account mutation execution (handled by Wallet Service, Prompt 203).
  - Direct banking gateway webhooks and automated corporate disbursement payouts (handled by Payment Gateway Service, Prompt 212).
  - Central Limit Order Book (CLOB) matching for secondary security tokens (handled by Prompt 205).
  - User identity document extraction, Aadhaar/PAN verification, and sanctions screening (handled by KYC/AML Service, Prompt 202).
  - Global administrative RBAC authorization and hardware MFA verification (handled by Admin Service, Prompt 217).
  - Depository Delivery-versus-Payment (DvP) settlement with NSDL/CDSL (handled by Prompt 213).

## Technology to Use
- **Primary Language & Runtime:** Go 1.22+. Selected for lightweight concurrency primitives (goroutines/channels), garbage collection predictability, low memory footprint, and low-latency gRPC services.
- **Database & Persistence:** PostgreSQL 16+ using `pgx/v5` connection pooling with strict transactional isolation (`READ COMMITTED` with pessimistic row-locking `FOR UPDATE` or `SERIALIZABLE` for dispute resolution sweeps).
- **SQL Code Generation:** `sqlc` for compile-time verified, type-safe SQL query generation without runtime reflection overhead.
- **Distributed Caching & Timers:** Redis 7.2+ utilizing Redis Keyspace Notifications (`notify-keyspace-events Ex`) and distributed locks (`redsync/v4`) for high-precision countdown timers.
- **Object Storage:** AWS S3 or MinIO with Server-Side Encryption (SSE-KMS / SSE-S3) and S3 Object Lock (WORM compliance) for immutable dispute evidence storage.
- **Event Streaming:** Apache Kafka 3.7+ via `segmentio/kafka-go` with manual commit acknowledgement after database persistence.
- **Image Hashing & Forensics:** `github.com/corona10/goimagehash` for perceptual hashing (Difference Hash / Wavelet Hash) and `crypto/sha256` for file integrity digests.
- **Blockchain Connectivity:** `github.com/ethereum/go-ethereum` (geth v1.13+) for secp256k1 signature validation, ABI packing, and JSON-RPC over mTLS to Hyperledger Besu.
- **Financial Calculations:** `github.com/shopspring/decimal` for fixed-point decimal arithmetic in INR paise and fractional token quantities.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `p2p_advertisements`, `p2p_trades`, `p2p_disputes`, `p2p_evidence_files`, `p2p_arbitration_votes`.
- **Redis 7.2 Cluster:** Active countdown timer keys (`p2p:timer:pay:{trade_id}`, `p2p:timer:release:{trade_id}`) and mutex locks (`p2p:lock:trade:{trade_id}`).
- **AWS S3 / MinIO:** Bucket `growww-p2p-evidence-vault` storing buyer/seller receipts, statements, and screenshots with 8-year WORM retention.
- **Apache Kafka:**
  - Subscribes to: `user.kyc.verified.v1`, `admin.dispute.voted.v1`.
  - Publishes to: `p2p.trade.created.v1`, `p2p.payment.marked_paid.v1`, `p2p.escrow.released.v1`, `p2p.dispute.opened.v1`, `p2p.dispute.resolved.v1`, `compliance.fiu_ind.suspicious_p2p.v1`.
- **Wallet Service (Prompt 203):** Invokes `ReserveFunds` to hold seller tokens in Account 2800 (P2P Escrow) and `CommitHold` to transfer tokens to buyer upon settlement.
- **Payment Gateway Integration Service (Prompt 212):** Invokes Open Banking / Account Aggregator endpoints for automated bank UTR verification.
- **Admin Service (Prompt 217):** Ingests Maker-Checker dual approvals and cryptographically signed operational arbitration decisions.
- **KYC/AML Service (Prompt 202):** Validates CKYC compliance status and bank account holder name matching before order intake.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Network Topology:** Hyperledger Besu private consortium network running Istanbul/QBFT (Quorum Byzantine Fault Tolerance) consensus with 1-second block periods, zero transaction gas fees, and authorized relayer nodes.
- **Smart Contract Interfaced:** `P2PEscrow.sol` deployed on Hyperledger Besu:
  - `lockCollateral(bytes32 tradeId, address seller, address tokenAddress, uint256 tokenAmount, bytes32 proofHash)`: Locks tokens from seller's custody address into the smart contract vault upon order creation.
  - `releaseToBuyer(bytes32 tradeId, bytes sellerSignature)`: Standard cooperative settlement path executed when the seller confirms fiat receipt.
  - `resolveDispute(bytes32 tradeId, address targetRecipient, bytes[2] officerSignatures)`: 2-of-3 multi-signature release executed only when two distinct compliance officers (Maker and Checker) sign off on arbitration outcome.
- **Off-Chain / On-Chain Parity:** The smart contract vault mirrors off-chain double-entry ledger Account 2800 (P2P Escrow Holding Account). Token movements are atomically matched between PostgreSQL postings and Besu transaction hashes.
- **Zero On-Chain PII Invariant:** In compliance with India's Digital Personal Data Protection (DPDP) Act 2023, zero investor names, phone numbers, bank account numbers, IFSC codes, UPI VPAs, or raw dispute images exist on-chain. Only pseudonymous Besu addresses (`0x...`), keccak256 trade hashes, and SHA-256 evidence digests are recorded on the ledger.
- **Cryptographic Relayer & CloudHSM:** Transactions submitted to Besu are signed via authorized platform relayer keys hosted in AWS CloudHSM or HashiCorp Vault Transit Engine.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service Structure:** Initialize Go module `services/p2p-escrow-service` with standardized directory layout (`cmd/`, `internal/domain/`, `internal/repository/`, `internal/service/`, `internal/chain/`, `internal/timer/`) and configure `golangci-lint`.
2. **Compile Solidity ABI Stubs:** Run `abigen` on `P2PEscrow.sol` ABI to generate type-safe Go bindings for collateral locking, cooperative releases, and multi-sig dispute settlements.
3. **Define Protobuf Contracts:** Create `proto/growww/p2p/v1/p2p_escrow_service.proto` defining RPCs for advertisement management, trade initiation, payment state updates, dispute creation, evidence submission, and arbitration voting.
4. **Generate Go Protobuf & gRPC Code:** Compile proto definitions using `protoc-gen-go` and `protoc-gen-go-grpc`.
5. **Design PostgreSQL Schema:** Author migration scripts creating `p2p_advertisements`, `p2p_trades`, `p2p_disputes`, `p2p_evidence_files`, and `p2p_arbitration_votes` with appropriate foreign keys, status ENUMs, and check constraints.
6. **Configure `sqlc`:** Set up `sqlc.yaml` and generate type-safe, compile-time verified Go database queries.
7. **Implement P2P Advertisement Engine:** Build maker advertisement manager enforcing KYC verification checks, price limits, minimum/maximum trade boundaries, and supported payment rails.
8. **Implement Escrow State Machine:** Build core state machine handling deterministic transitions (`ORDER_CREATED` -> `COLLATERAL_LOCKED` -> `PAYMENT_MARKED_PAID` -> `COMPLETED`, or branching to `DISPUTE_OPENED`).
9. **Implement Redis Countdown Timer:** Build timer subsystem using Redis TTL keys and keyspace notification listener. If a buyer fails to mark paid within 15 minutes, automatically cancel order and unlock collateral back to seller.
10. **Implement Proof Ingestion & Perceptual Hashing:** Build file ingestion endpoint storing images in AWS S3 / MinIO. Compute SHA-256 file checksum and perceptual hash (pHash); check Redis/PostgreSQL for duplicate image hashes across existing trades to intercept recycled screenshot scams.
11. **Implement Wallet Service Integration:** Connect gRPC client to Wallet Service (Prompt 203) to atomically invoke `ReserveFunds` (locking seller tokens into Account 2800) and `CommitHold` upon successful completion.
12. **Implement Maker-Checker Arbitration Engine (RUNBOOK-27):** Build dual-control dispute desk. Require two distinct compliance officers: Officer 1 (Maker) proposes release or refund; Officer 2 (Checker) verifies evidence independently and approves.
13. **Implement Besu On-Chain Coordinator:** Build client that signs transactions via AWS CloudHSM / Vault and broadcasts `lockCollateral`, `releaseToBuyer`, or `resolveDispute` with 2-of-3 signatures to `P2PEscrow.sol`.
14. **Implement Kafka Event Streamer:** Build producers streaming `p2p.trade.*`, `p2p.dispute.*`, and `compliance.fiu_ind.suspicious_p2p.v1` events. Connect notification triggers for buyer/seller alerts.
15. **Integration & Stress Testing:** Write comprehensive unit and integration tests using `testcontainers-go` for PostgreSQL, Redis, MinIO, and Besu, verifying zero race conditions, reliable timer cancellations, and tamper-resistant dispute resolutions under 1,000 concurrent trades.

## Interfaces / Contracts

### Protobuf Definition (`p2p_escrow_service.proto`)
```protobuf
syntax = "proto3";

package growww.p2p.v1;

option go_package = "growww/p2p/v1;p2pv1";

service P2PEscrowService {
  rpc CreateAdvertisement (CreateAdvertisementRequest) returns (CreateAdvertisementResponse);
  rpc InitiateP2PTrade (InitiateP2PTradeRequest) returns (InitiateP2PTradeResponse);
  rpc MarkPaymentSent (MarkPaymentSentRequest) returns (MarkPaymentSentResponse);
  rpc ConfirmPaymentReceipt (ConfirmPaymentReceiptRequest) returns (ConfirmPaymentReceiptResponse);
  rpc OpenDispute (OpenDisputeRequest) returns (OpenDisputeResponse);
  rpc UploadDisputeEvidence (UploadDisputeEvidenceRequest) returns (UploadDisputeEvidenceResponse);
  rpc SubmitArbitrationVote (SubmitArbitrationVoteRequest) returns (SubmitArbitrationVoteResponse);
  rpc GetTradeDetails (GetTradeDetailsRequest) returns (GetTradeDetailsResponse);
}

enum TradeSide {
  TRADE_SIDE_UNSPECIFIED = 0;
  TRADE_SIDE_BUY = 1;
  TRADE_SIDE_SELL = 2;
}

enum PaymentMethodType {
  PAYMENT_METHOD_UNSPECIFIED = 0;
  PAYMENT_METHOD_UPI = 1;
  PAYMENT_METHOD_IMPS = 2;
  PAYMENT_METHOD_NEFT = 3;
}

enum TradeStatus {
  TRADE_STATUS_UNSPECIFIED = 0;
  TRADE_STATUS_ORDER_CREATED = 1;
  TRADE_STATUS_COLLATERAL_LOCKED = 2;
  TRADE_STATUS_PAYMENT_MARKED_PAID = 3;
  TRADE_STATUS_COMPLETED = 4;
  TRADE_STATUS_DISPUTE_OPENED = 5;
  TRADE_STATUS_DISPUTE_RESOLVED = 6;
  TRADE_STATUS_CANCELLED_EXPIRED = 7;
}

enum DisputeReason {
  DISPUTE_REASON_UNSPECIFIED = 0;
  DISPUTE_REASON_PAYMENT_NOT_RECEIVED = 1;
  DISPUTE_REASON_AMOUNT_INCORRECT = 2;
  DISPUTE_REASON_THIRD_PARTY_PAYMENT = 3;
  DISPUTE_REASON_FRAUDULENT_RECEIPT = 4;
}

enum ArbitrationDecision {
  ARBITRATION_DECISION_UNSPECIFIED = 0;
  ARBITRATION_DECISION_RELEASE_TO_BUYER = 1;
  ARBITRATION_DECISION_REFUND_TO_SELLER = 2;
}

message CreateAdvertisementRequest {
  string idempotency_key = 1;
  string maker_user_id = 2;
  TradeSide side = 3;
  string token_symbol = 4;
  string token_contract_address = 5;
  string total_quantity = 6;
  string min_trade_amount_inr = 7;
  string max_trade_amount_inr = 8;
  string unit_price_inr = 9;
  repeated PaymentMethodType accepted_payment_methods = 10;
  string payment_details_json = 11;
}

message CreateAdvertisementResponse {
  string advertisement_id = 1;
  string status = 2;
  int64 created_at_unix = 3;
}

message InitiateP2PTradeRequest {
  string idempotency_key = 1;
  string advertisement_id = 2;
  string taker_user_id = 3;
  string fiat_amount_inr = 4;
  PaymentMethodType selected_payment_method = 5;
}

message InitiateP2PTradeResponse {
  string trade_id = 1;
  TradeStatus status = 2;
  string crypto_amount = 3;
  string fiat_amount_inr = 4;
  int64 payment_window_expires_at_unix = 5;
  string seller_payment_details_json = 6;
  string onchain_escrow_tx_hash = 7;
}

message MarkPaymentSentRequest {
  string idempotency_key = 1;
  string trade_id = 2;
  string buyer_user_id = 3;
  string bank_utr_number = 4;
  string sender_bank_account_number = 5;
  string sender_upi_vpa = 6;
}

message MarkPaymentSentResponse {
  string trade_id = 1;
  TradeStatus status = 2;
  int64 seller_confirmation_expires_at_unix = 3;
}

message ConfirmPaymentReceiptRequest {
  string idempotency_key = 1;
  string trade_id = 2;
  string seller_user_id = 3;
}

message ConfirmPaymentReceiptResponse {
  string trade_id = 1;
  TradeStatus status = 2;
  string onchain_release_tx_hash = 3;
  int64 settled_at_unix = 4;
}

message OpenDisputeRequest {
  string idempotency_key = 1;
  string trade_id = 2;
  string disputing_user_id = 3;
  DisputeReason reason = 4;
  string dispute_statement = 5;
}

message OpenDisputeResponse {
  string dispute_id = 1;
  string trade_id = 2;
  TradeStatus status = 3;
  int64 opened_at_unix = 4;
}

message UploadDisputeEvidenceRequest {
  string dispute_id = 1;
  string user_id = 2;
  string file_type = 3; // "IMAGE_JPEG", "IMAGE_PNG", "PDF_STATEMENT"
  bytes file_payload = 4;
  string utr_number = 5;
}

message UploadDisputeEvidenceResponse {
  string evidence_file_id = 1;
  string sha256_checksum = 2;
  string phash_checksum = 3;
  string storage_s3_key = 4;
  bool duplicate_detected = 5;
}

message SubmitArbitrationVoteRequest {
  string idempotency_key = 1;
  string dispute_id = 2;
  string officer_id = 3;
  string officer_role = 4; // "MAKER" or "CHECKER"
  ArbitrationDecision decision = 5;
  string officer_notes = 6;
  bytes officer_signature = 7;
}

message SubmitArbitrationVoteResponse {
  string dispute_id = 1;
  bool is_resolved = 2;
  TradeStatus updated_trade_status = 3;
  string onchain_settlement_tx_hash = 4;
}

message GetTradeDetailsRequest {
  string trade_id = 1;
}

message GetTradeDetailsResponse {
  string trade_id = 1;
  string advertisement_id = 2;
  string buyer_user_id = 3;
  string seller_user_id = 4;
  string token_symbol = 5;
  string crypto_amount = 6;
  string fiat_amount_inr = 7;
  TradeStatus status = 8;
  string bank_utr_number = 9;
  int64 payment_window_expires_at_unix = 10;
  int64 created_at_unix = 11;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE p2p_trade_side_enum AS ENUM ('BUY', 'SELL');
CREATE TYPE p2p_payment_method_enum AS ENUM ('UPI', 'IMPS', 'NEFT');
CREATE TYPE p2p_trade_status_enum AS ENUM (
    'ORDER_CREATED',
    'COLLATERAL_LOCKED',
    'PAYMENT_MARKED_PAID',
    'COMPLETED',
    'DISPUTE_OPENED',
    'DISPUTE_RESOLVED',
    'CANCELLED_EXPIRED'
);
CREATE TYPE p2p_dispute_reason_enum AS ENUM (
    'PAYMENT_NOT_RECEIVED',
    'AMOUNT_INCORRECT',
    'THIRD_PARTY_PAYMENT',
    'FRAUDULENT_RECEIPT'
);
CREATE TYPE p2p_dispute_status_enum AS ENUM (
    'PENDING_EVIDENCE',
    'UNDER_MAKER_REVIEW',
    'UNDER_CHECKER_APPROVAL',
    'RESOLVED_RELEASE_BUYER',
    'RESOLVED_REFUND_SELLER',
    'ESCALATED_TO_MLRO'
);
CREATE TYPE p2p_arbitration_decision_enum AS ENUM ('RELEASE_TO_BUYER', 'REFUND_TO_SELLER');
CREATE TYPE p2p_officer_role_enum AS ENUM ('MAKER', 'CHECKER');

CREATE TABLE p2p_advertisements (
    advertisement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    maker_user_id UUID NOT NULL,
    side p2p_trade_side_enum NOT NULL,
    token_symbol VARCHAR(16) NOT NULL,
    token_contract_address VARCHAR(42) NOT NULL,
    total_quantity NUMERIC(28, 18) NOT NULL CHECK (total_quantity > 0),
    available_quantity NUMERIC(28, 18) NOT NULL CHECK (available_quantity >= 0),
    min_trade_amount_inr NUMERIC(18, 4) NOT NULL CHECK (min_trade_amount_inr > 0),
    max_trade_amount_inr NUMERIC(18, 4) NOT NULL CHECK (max_trade_amount_inr >= min_trade_amount_inr),
    unit_price_inr NUMERIC(18, 4) NOT NULL CHECK (unit_price_inr > 0),
    accepted_payment_methods p2p_payment_method_enum[] NOT NULL,
    payment_details_encrypted TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE p2p_trades (
    trade_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertisement_id UUID NOT NULL REFERENCES p2p_advertisements(advertisement_id),
    buyer_user_id UUID NOT NULL,
    seller_user_id UUID NOT NULL,
    token_symbol VARCHAR(16) NOT NULL,
    crypto_amount NUMERIC(28, 18) NOT NULL CHECK (crypto_amount > 0),
    fiat_amount_inr NUMERIC(18, 4) NOT NULL CHECK (fiat_amount_inr > 0),
    unit_price_inr NUMERIC(18, 4) NOT NULL,
    payment_method p2p_payment_method_enum NOT NULL,
    status p2p_trade_status_enum NOT NULL DEFAULT 'ORDER_CREATED',
    bank_utr_number VARCHAR(64),
    sender_account_or_vpa VARCHAR(128),
    hold_id UUID, -- References Wallet Service funds_holds
    onchain_trade_id_hash BYTEA UNIQUE, -- keccak256 hash passed to P2PEscrow.sol
    onchain_lock_tx_hash VARCHAR(66),
    onchain_release_tx_hash VARCHAR(66),
    payment_window_expires_at TIMESTAMPTZ NOT NULL,
    seller_confirmation_expires_at TIMESTAMPTZ,
    settled_at TIMESTAMPTZ,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_distinct_parties CHECK (buyer_user_id != seller_user_id)
);

CREATE TABLE p2p_disputes (
    dispute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_id UUID NOT NULL UNIQUE REFERENCES p2p_trades(trade_id),
    opened_by_user_id UUID NOT NULL,
    reason p2p_dispute_reason_enum NOT NULL,
    dispute_statement TEXT NOT NULL,
    status p2p_dispute_status_enum NOT NULL DEFAULT 'PENDING_EVIDENCE',
    evidence_deadline TIMESTAMPTZ NOT NULL,
    maker_officer_id UUID,
    maker_decision p2p_arbitration_decision_enum,
    maker_notes TEXT,
    maker_voted_at TIMESTAMPTZ,
    checker_officer_id UUID,
    checker_decision p2p_arbitration_decision_enum,
    checker_notes TEXT,
    checker_voted_at TIMESTAMPTZ,
    fiu_ind_str_reported BOOLEAN NOT NULL DEFAULT FALSE,
    fiu_ind_report_reference VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_distinct_officers CHECK (maker_officer_id IS NULL OR checker_officer_id IS NULL OR maker_officer_id != checker_officer_id)
);

CREATE TABLE p2p_evidence_files (
    evidence_file_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES p2p_disputes(dispute_id) ON DELETE CASCADE,
    uploaded_by_user_id UUID NOT NULL,
    file_type VARCHAR(32) NOT NULL,
    s3_bucket VARCHAR(100) NOT NULL,
    s3_key VARCHAR(512) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    sha256_checksum CHAR(64) NOT NULL,
    phash_checksum VARCHAR(64),
    extracted_utr VARCHAR(64),
    is_duplicate BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE p2p_arbitration_votes (
    vote_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dispute_id UUID NOT NULL REFERENCES p2p_disputes(dispute_id) ON DELETE CASCADE,
    officer_id UUID NOT NULL,
    role p2p_officer_role_enum NOT NULL,
    decision p2p_arbitration_decision_enum NOT NULL,
    officer_signature BYTEA NOT NULL,
    notes TEXT NOT NULL,
    voted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_dispute_officer_role UNIQUE (dispute_id, role)
);

CREATE INDEX idx_p2p_ads_search ON p2p_advertisements(side, token_symbol, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_p2p_trades_buyer ON p2p_trades(buyer_user_id, status);
CREATE INDEX idx_p2p_trades_seller ON p2p_trades(seller_user_id, status);
CREATE INDEX idx_p2p_trades_timer ON p2p_trades(status, payment_window_expires_at) WHERE status IN ('ORDER_CREATED', 'COLLATERAL_LOCKED');
CREATE INDEX idx_p2p_disputes_status ON p2p_disputes(status);
CREATE INDEX idx_p2p_evidence_sha256 ON p2p_evidence_files(sha256_checksum);
CREATE INDEX idx_p2p_evidence_phash ON p2p_evidence_files(phash_checksum) WHERE phash_checksum IS NOT NULL;
CREATE INDEX idx_p2p_evidence_utr ON p2p_evidence_files(extracted_utr) WHERE extracted_utr IS NOT NULL;
```

## Security & Compliance Notes
- **Mandatory CKYC & AML Verification:** In accordance with Prevention of Money Laundering Act (PMLA) and SEBI Master Circulars, P2P order placement and taking are restricted to users with verified CKYC status and clean Sanctions/PEP screening. Unverified users are rejected immediately with `PERMISSION_DENIED`.
- **Strict Name Matching & Prohibition of Third-Party Transfers:** Fiat payments must originate from and be sent to bank accounts/VPAs legally matching the verified KYC name on the user's Growww account. Third-party transfers result in immediate trade disqualification, collateral return to seller, and buyer account freezing under suspicion of money mule activity.
- **Image Forensics & Screenshot Deduplication:** To defeat fraudulent payment claims where bad actors submit fabricated or recycled screenshots, every uploaded file undergoes SHA-256 integrity verification and perceptual hash (pHash) analysis. If a pHash hamming distance $< 5$ or exact SHA-256 is detected across disparate accounts or past trades, the trade is flagged for fraud, and the disputing party's account is quarantined.
- **Dual-Control (Maker-Checker) Arbitration Invariant (RUNBOOK-27):** Collateral locked in dispute cannot be disbursed by a single administrator. Resolution mandates two independent compliance officers: a Maker (who examines proof, verifies bank statements, and proposes resolution) and a Checker (who independently verifies Open Banking UTR records and approves). The smart contract enforces 2-of-3 multi-signature verification on-chain.
- **Statutory FIU-IND Reporting:** When a dispute uncovers forged receipts, repeated UTR spoofing, or confirmed money mule patterns, the service generates an automated Suspicious Transaction Report (STR) payload for the Financial Intelligence Unit - India (FIU-IND) via Kafka topic `compliance.fiu_ind.suspicious_p2p.v1`.
- **WORM Tamper-Evident Evidence Retention:** Dispute evidence files (bank statements, payment screenshots) stored in S3/MinIO utilize S3 Object Lock in Compliance Mode with an 8-year statutory retention period as mandated by Rule 3 of PMLA Maintenance of Records Rules.

## Acceptance Criteria
- [ ] P2P advertisement engine validates maker KYC status, minimum/maximum trade amounts, and valid payment rail configurations before publishing.
- [ ] Taker order initiation locks seller collateral atomically in Wallet Service Account 2800 and on Hyperledger Besu `P2PEscrow.sol`.
- [ ] Redis countdown timer triggers automatic cancellation and collateral unlocking if buyer does not mark payment sent within 15 minutes.
- [ ] Once buyer marks payment sent and submits bank UTR, order transitions to `PAYMENT_MARKED_PAID` and starts the 30-minute seller confirmation countdown.
- [ ] Seller confirmation triggers cooperative release: collateral is transferred to buyer trading balance and confirmed on Besu in a single atomic flow.
- [ ] Dispute opening immediately freezes collateral in `STATUS_IN_DISPUTE`, blocking cancellation or auto-release.
- [ ] Uploaded evidence computes SHA-256 and pHash; duplicate screenshots are intercepted with fraud alerts.
- [ ] Maker-Checker arbitration requires two distinct compliance officers with different IDs; unilateral resolution attempts are rejected.
- [ ] Hyperledger Besu `P2PEscrow.sol` validates 2-of-3 signatures before releasing disputed collateral on-chain.
- [ ] Fraudulent or suspicious dispute outcomes emit `compliance.fiu_ind.suspicious_p2p.v1` events for regulatory reporting.
- [ ] Zero Personal Identifiable Information (PII) is published to the Hyperledger Besu blockchain.

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Design Standards), 104 (Kafka Event Schemas), 111 (Domain Models), 202 (KYC & AML Service), 203 (Wallet Account Service).
- **Parallel Tasks:** 211 (Transactional Notification Service), 212 (Payment Gateway Integration Service), 217 (Admin & Back-Office Service).
- **Downstream Blockers:** 528 (Mobile P2P Trading Screens), 604 (Admin Dispute Resolution Portal), 216 (Regulatory Statutory Reporting Service).
