# 212 - Banking & Payment Gateway Integration Service (Go)

## Purpose
The Banking & Payment Gateway Integration Service is the regulated fiat on-ramp and off-ramp gateway for Growww's domestic operations. It connects the Growww investment platform directly to RBI-regulated banking channels - including Unified Payments Interface (UPI Intent, Collect, and Dynamic QR), NetBanking, Immediate Payment Service (IMPS), National Electronic Funds Transfer (NEFT), and Real Time Gross Settlement (RTGS) - via licensed Payment Aggregators (Razorpay, Cashfree) and Scheduled Commercial Bank Nodal APIs (ICICI Bank, HDFC Bank).

In strict compliance with SEBI and RBI anti-money laundering regulations, Growww prohibits third-party deposits: funds can only be deposited from or withdrawn to a verified bank account registered in the investor's own name. This service automates penny-drop bank account verification, processes real-time deposits, validates cryptographically signed bank webhooks, and executes automated withdrawal disbursements from segregated nodal escrow accounts.

## What You Are Building
A high-security, resilient Go microservice (`services/payment-gateway-service`). Concrete deliverables include:
- Multi-rail payment intent generator (UPI deep-links, dynamic QR codes, NetBanking checkout tokens).
- Penny-drop bank verification client (interfacing with NPCI/Bank IMPS APIs to verify account ownership and validate beneficiary name against KYC records using fuzzy matching).
- High-security HTTP webhook ingestion router with cryptographic HMAC-SHA256 signature verification and replay protection.
- Automated payout/withdrawal dispatch engine supporting instant IMPS payouts and bulk RTGS/NEFT batch files.
- Double-entry coordination client invoking Wallet Service (Prompt 203) to credit verified deposits and commit payout debits.
- Asynchronous Kafka event publisher streaming `payment.deposit.confirmed.v1`, `payment.payout.dispatched.v1`, `payment.bank_account.verified.v1`.

## Scope Boundaries
- **In Scope:**
 - Generating UPI payment intents and dynamic QR payloads.
 - Penny-drop account verification (transferring ₹1.00 to verify account active status and legal name).
 - Webhook ingestion, signature verification, and idempotency handling.
 - Withdrawal payout execution and bank UTR (Unique Transaction Reference) tracking.
 - Managing investor registered bank accounts (`bank_accounts`).
- **Out of Scope / Handled Elsewhere:**
 - User INR ledger balance storage and double-entry postings (Prompt 203).
 - Foreign currency (USD/EUR/GBP) SWIFT/Fedwire funding for GIFT City (Prompt 214).
 - Order intake and pre-trade fund reservation (Prompt 204).
 - KYC document extraction and identity verification (Prompt 202).

## Technology to Use
- **Primary Language & Framework:** Go 1.22+. Selected for its strict typing, robust cryptographic standard library (`crypto/hmac`, `crypto/sha256`), high-performance HTTP routing (`go-chi/chi`), and low-latency gRPC client/server capabilities.
- **Database & Storage:** PostgreSQL 16+ using `pgx/v5` and `sqlc` for storing bank accounts and transaction audit records.
- **Financial Math & Decimals:** `github.com/shopspring/decimal` for exact monetary amounts in INR.
- **Messaging & Streaming:** `segmentio/kafka-go` for publishing payment state changes.
- **Fuzzy Name Matching:** `github.com/schollz/closestmatch` or Jaro-Winkler implementation for bank name vs KYC name verification (threshold $\ge 85\%$).

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `bank_accounts`, `payment_transactions`, `payout_requests`, `webhook_events`.
- **Redis 7.2:** Webhook idempotency locks (`pay:idemp:{provider}:{event_id}`) with 24-hour TTL.
- **Banking / PA Gateways:** Razorpay / Cashfree APIs, ICICI/HDFC Corporate Nodal APIs (mTLS + API Key).
- **Apache Kafka:** Publishes to `payment.events.v1`.
- **Wallet Service (Prompt 203):** Invoked via gRPC `CreditDeposit` to credit funds upon verified bank receipt.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **100% Real INR Fiat Invariant:** Domestic funding operates exclusively via RBI-licensed banking rails into SEBI-mandated nodal escrow bank accounts. No cryptocurrencies, unregulated stablecoins (USDT/USDC), or decentralized on-ramps are utilized.
- **Deposit-to-Ledger Bridge:** When this service confirms an INR deposit from an investor's verified bank account, it credits their off-chain wallet ledger, enabling them to purchase fractional equity tokens on Hyperledger Besu.
- **Audit-Trail Reconciliations:** Bank UTR references are linked to the downstream trade settlement transaction hashes on Hyperledger Besu, ensuring seamless end-to-end auditability from bank deposit to on-chain asset custody.
- **Zero On-Chain Banking Data:** Under DPDP Act 2023, bank account numbers, IFSC codes, and bank UTRs are never published to the blockchain.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/payment-gateway-service` with strict linter rules, Makefile, and standard layout (`cmd/`, `internal/bank/`, `internal/providers/`, `internal/webhook/`).
2. **Define Protobuf Contracts:** Create `proto/growww/payment/v1/payment_service.proto` defining `InitiateDeposit`, `InitiatePayout`, `VerifyBankAccount`, and `GetTransactionStatus`.
3. **Generate Go Stubs:** Compile protobuf definitions to Go gRPC client and server stubs.
4. **Design PostgreSQL Schema:** Write migrations for `bank_accounts`, `payment_transactions`, `payout_requests`, and `webhook_events`.
5. **Implement Penny-Drop Verification Module:** Build client that initiates ₹1.00 IMPS penny-drop to user's bank account, extracts beneficiary name returned by NPCI/bank, and runs Jaro-Winkler string similarity against KYC full name.
6. **Implement Multi-Provider Deposit Generator:** Build adapters for Razorpay and Cashfree to create UPI Intent URLs (`upi://pay?...`), UPI Collect requests, and dynamic QR base64 strings.
7. **Implement Webhook Ingestion Router:** Build HTTP endpoint `/v1/payments/webhooks/{provider}` using `chi` router. Verify incoming HMAC-SHA256 signature using provider secret keys stored in KMS/Vault.
8. **Implement Webhook Deduplication & Replay Guard:** Check Redis key `pay:idemp:{provider}:{event_id}`. If duplicate, respond HTTP 200 immediately without re-processing.
9. **Build Transaction State Machine:** Manage states (`INITIATED`, `PENDING_USER`, `SUCCESS`, `FAILED`, `EXPIRED`, `REFUNDED`) with optimistic locking.
10. **Implement Wallet Credit Dispatcher:** Upon webhook confirmation of `SUCCESS`, call `WalletService.CreditDeposit` via gRPC with idempotency key.
11. **Implement Automated Payout Engine:** Build withdrawal worker that selects approved `payout_requests`, verifies nodal account balance, and initiates instant IMPS/RTGS transfers via corporate banking APIs.
12. **Implement Payout Status Poller:** Build cron/ticker worker that polls bank APIs for pending payouts every 60 seconds and updates final UTR numbers.
13. **Publish Kafka Events:** Stream payment lifecycle events to `payment.events.v1` for downstream notification and audit logging.
14. **Configure Telemetry & Metrics:** Expose Prometheus metrics for deposit success rate, webhook latency, penny-drop failure reasons, and bank API response times.
15. **Write Comprehensive Test Suite:** Write unit and integration tests with mock bank APIs and testcontainers, testing signature verification, duplicate webhook deduplication, and penny-drop name mismatch rejection.

## Interfaces / Contracts

### Protobuf Definition (`payment_service.proto`)
```protobuf
syntax = "proto3";

package growww.payment.v1;

option go_package = "growww/payment/v1;paymentv1";

service PaymentGatewayService {
  rpc InitiateDeposit (InitiateDepositRequest) returns (InitiateDepositResponse);
  rpc InitiatePayout (InitiatePayoutRequest) returns (InitiatePayoutResponse);
  rpc VerifyBankAccount (VerifyBankAccountRequest) returns (VerifyBankAccountResponse);
  rpc GetTransactionStatus (GetTransactionStatusRequest) returns (GetTransactionStatusResponse);
}

enum PaymentMethod {
  PAYMENT_METHOD_UNSPECIFIED = 0;
  PAYMENT_METHOD_UPI_INTENT = 1;
  PAYMENT_METHOD_UPI_QR = 2;
  PAYMENT_METHOD_UPI_COLLECT = 3;
  PAYMENT_METHOD_NETBANKING = 4;
  PAYMENT_METHOD_IMPS = 5;
  PAYMENT_METHOD_NEFT_RTGS = 6;
}

enum TransactionStatus {
  TX_STATUS_UNSPECIFIED = 0;
  TX_STATUS_INITIATED = 1;
  TX_STATUS_PENDING = 2;
  TX_STATUS_SUCCESS = 3;
  TX_STATUS_FAILED = 4;
  TX_STATUS_EXPIRED = 5;
}

message InitiateDepositRequest {
  string idempotency_key = 1;
  string user_id = 2;
  string bank_account_id = 3;
  string amount_inr = 4; // Decimal string (e.g. "5000.00")
  PaymentMethod payment_method = 5;
  string upi_vpa = 6; // Required for UPI_COLLECT
}

message InitiateDepositResponse {
  string transaction_id = 1;
  string payment_gateway_reference = 2;
  string upi_intent_url = 3; // For mobile deep-link
  string qr_code_base64 = 4; // For web QR display
  int64 expires_at_unix = 5;
}

message InitiatePayoutRequest {
  string idempotency_key = 1;
  string user_id = 2;
  string bank_account_id = 3;
  string amount_inr = 4;
}

message InitiatePayoutResponse {
  string payout_id = 1;
  TransactionStatus status = 2;
  string message = 3;
  int64 estimated_settlement_unix = 4;
}

message VerifyBankAccountRequest {
  string user_id = 1;
  string account_number = 2;
  string ifsc_code = 3;
  string account_holder_name_input = 4;
}

message VerifyBankAccountResponse {
  string bank_account_id = 1;
  bool is_verified = 2;
  string verified_name_from_bank = 3;
  string name_match_score = 4; // e.g. "0.94"
  string bank_name = 5;
  string rejection_reason = 6;
}

message GetTransactionStatusRequest {
  string transaction_id = 1;
}

message GetTransactionStatusResponse {
  string transaction_id = 1;
  string user_id = 2;
  string amount_inr = 3;
  TransactionStatus status = 4;
  string bank_utr = 5;
  int64 completed_at_unix = 6;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TYPE bank_account_status_enum AS ENUM ('PENDING_VERIFICATION', 'VERIFIED', 'REJECTED', 'DISABLED');
CREATE TYPE tx_direction_enum AS ENUM ('DEPOSIT', 'WITHDRAWAL');
CREATE TYPE payment_tx_status_enum AS ENUM ('INITIATED', 'PENDING', 'SUCCESS', 'FAILED', 'EXPIRED');

CREATE TABLE bank_accounts (
    bank_account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    account_number_encrypted BYTEA NOT NULL, -- AES-256-GCM encrypted
    account_number_last4 VARCHAR(4) NOT NULL,
    ifsc_code VARCHAR(11) NOT NULL,
    bank_name VARCHAR(100) NOT NULL,
    beneficiary_name_bank VARCHAR(255),
    name_match_score NUMERIC(5, 4),
    status bank_account_status_enum NOT NULL DEFAULT 'PENDING_VERIFICATION',
    is_primary BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payment_transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    bank_account_id UUID NOT NULL REFERENCES bank_accounts(bank_account_id),
    direction tx_direction_enum NOT NULL,
    amount_inr NUMERIC(18, 4) NOT NULL CHECK (amount_inr > 0),
    payment_method VARCHAR(50) NOT NULL,
    provider_name VARCHAR(50) NOT NULL, -- RAZORPAY, CASHFREE, ICICI_CORP
    gateway_order_id VARCHAR(255) NOT NULL UNIQUE,
    gateway_payment_id VARCHAR(255),
    bank_utr VARCHAR(100),
    status payment_tx_status_enum NOT NULL DEFAULT 'INITIATED',
    error_code VARCHAR(100),
    error_description TEXT,
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE webhook_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_name VARCHAR(50) NOT NULL,
    provider_event_id VARCHAR(255) NOT NULL,
    payload_hash VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    is_processed BOOLEAN NOT NULL DEFAULT FALSE,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_provider_event UNIQUE (provider_name, provider_event_id)
);

CREATE INDEX idx_payment_user ON payment_transactions(user_id);
CREATE INDEX idx_payment_status ON payment_transactions(status);
```

## Security & Compliance Notes
- **Third-Party Deposit Prevention (SEBI Mandate):** Deposits from bank accounts where the penny-drop beneficiary name does not match the KYC verified name ($\text{score} < 85\%$) must be automatically rejected and refunded to source within 24 hours.
- **HMAC Webhook Signatures:** All incoming webhook payloads must be verified against HMAC-SHA256 signatures before parsing; invalid signatures must result in HTTP 401 Unauthorized.
- **Bank Data Encryption at Rest (DPDP Act 2023):** Full bank account numbers are encrypted at rest using AES-256-GCM envelope encryption. Only the last 4 digits (`account_number_last4`) are stored in plaintext for user display.
- **Nodal Escrow Segregation:** Payout withdrawals are processed exclusively from RBI-regulated nodal escrow bank accounts with dual-authorization API keys.

## Acceptance Criteria
- [ ] Penny-drop integration successfully verifies active Indian bank accounts and computes name similarity against KYC name.
- [ ] Mismatched bank account names ($< 85\%$ match) are rejected with clear error codes.
- [ ] UPI Intent and Dynamic QR codes are generated and functional on mobile and web clients.
- [ ] Webhook receiver verifies HMAC-SHA256 signatures and rejects tampered or replayed payloads.
- [ ] Successful deposits atomically credit the user's INR balance in the Wallet Service.
- [ ] Full bank account numbers are encrypted at rest with zero plaintext exposure in logs.
- [ ] Payout requests trigger IMPS transfers and accurately update transaction UTR numbers.

## Suggested Order / Dependencies
- **Prerequisites:** 103 (API Standards), 112 (Idempotency), 201 (User Service), 203 (Wallet Service), 401 (PostgreSQL Schema).
- **Parallel Tasks:** 202 (KYC Service), 211 (Notification Service).
- **Downstream Blockers:** 511 (Flutter Wallet Screen), 601 (Web App).
