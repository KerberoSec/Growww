# 232 - 24/7 e₹ (Digital Rupee CBDC) & RBI RTGS Instant Liquidity Settlement Adapter (Go / mTLS / ISO 20022)

## Purpose
Traditional equity clearing and settlement cycles ($T+1$ / $T+0$) are constrained by commercial banking operating windows, batch NEFT/RTGS settlement cut-offs, and interbank liquidity latency. To achieve true instant atomic Delivery-versus-Payment (DvP Model 1) 24 hours a day, 7 days a week, the Clearing Corporation requires native integration with sovereign central bank money.

The **24/7 e₹ (Digital Rupee CBDC) & RBI RTGS Instant Liquidity Settlement Adapter** serves as the ultra-secure enterprise gateway connecting Growww's clearing and settlement infrastructure to:
1. **Reserve Bank of India (RBI) e₹-Wholesale (e₹-W) & e₹-Retail (e₹-R) Central Bank Digital Currency (CBDC) Distributed Ledgers.**
2. **RBI Real-Time Gross Settlement (RTGS) 24/7 ISO 20022 Host-to-Host (H2H) Financial Messaging Rails.**

By maintaining tokenized sovereign e₹ liquidity pools and enabling direct central bank money DvP settlement, the adapter eliminates commercial bank credit risk, provides continuous 24/7 instant cash settlement finality, and allows automated collateralized liquidity facility (CLF) rebalancing during volatile trading surges.

## What You Are Building
A mission-critical, banking-grade Go microservice (`services/cbdc-rtgs-adapter`) operating inside an isolated Virtual Private Cloud (VPC) with Hardware Security Module (HSM) signing, mTLS tunnels, and ISO 20022 XML/JSON parsers. Concrete deliverables include:
- **RBI e₹ CBDC Node Gateway:** Bidirectional integration with RBI CBDC Core Ledger APIs (via authorized Sponsor Banks) handling token minting, wallet holding verification, 24/7 atomic peer-to-peer transfers, and burn redemptions.
- **24/7 RTGS ISO 20022 Message Translator:** Robust serialization and deserialization engine processing standard financial messages:
 - `pacs.008.001.10` (Financial Institutional Customer Credit Transfer)
 - `pacs.009.001.10` (Financial Institutional Direct Credit Transfer for Interbank Clearing)
 - `pacs.002.001.12` (Payment Status Report / Settlement Confirmation)
 - `camt.053.001.10` (Bank-to-Customer End-of-Day Statement)
 - `camt.054.001.10` (Debit/Credit Notification)
- **HSM-Backed Digital Signature Engine:** Cryptographic signing of ISO 20022 payloads and CBDC transfer requests using PKI RSA-4096 / ECDSA secp256r1 keys stored in FIPS 140-2 Level 3 CloudHSM.
- **Continuous Liquidity Buffer Manager:** Dynamic liquidity tracker monitoring intraday clearinghouse e₹ and RTGS settlement reserves, executing automated sweeps between commercial bank escrow and RBI central bank accounts.
- **Atomic DvP Cash Settlement Relayer:** Real-time bridge executing instantaneous cash legs for DvP smart contract transactions (`SettlementDvP.sol`, Prompt 306).

## Scope Boundaries
- **In Scope:**
 - Connection management, mTLS handshake, and heartbeat monitoring to RBI CBDC and RTGS H2H endpoints.
 - Parsing, validation, signing, and schema enforcement for ISO 20022 message flows.
 - e₹ CBDC wallet management for Clearing Member pool accounts.
 - Automated liquidity sweeps and collateralized credit facility drawdowns.
 - Inbound webhook processing and outbox pattern delivery for settlement confirmations.
 - End-of-day bank statement reconciliation against internal cash ledgers.
- **Out of Scope / Handled Elsewhere:**
 - Retail UPI Intent and Net Banking payment gateway collections (handled in Prompt 212).
 - Internal wallet ledger double-entry bookkeeping (handled in Prompt 203).
 - On-chain fractional equity token transfers (handled in Prompt 306).
 - Pre-trade VaR and margin adequacy checking (handled in Prompt 229).

## Technology to Use
- **Primary Language & Framework:** **Go 1.22+** with `gin-gonic/gin` and `google.golang.org/grpc` for high-throughput concurrency and minimal memory footprint.
- **Cryptographic Security:** Hardware Security Module (AWS CloudHSM / Azure Dedicated HSM) integration via PKCS#11 C-Go wrappers and Go `crypto/tls` for mTLS 1.3 tunnels with client certificate pinning.
- **Financial Messaging Libraries:** `moov-io/iso20022` and custom XML schema validators validating ISO 20022 schema compliance.
- **Database & Storage:** **PostgreSQL 16+** with `pgx/v5` for encrypted transaction logs, raw ISO payload archives, and settlement reconciliation tables.
- **Distributed Caching & Queues:** **Redis 7.2** for sub-millisecond liquidity state caching and duplicate message idempotency checking.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `cbdc_wallets`, `cbdc_settlement_transactions`, `iso20022_messages`, `liquidity_buffers`, `cbdc_reconciliation_audit`.
- **Redis 7.2 Keys:** `cbdc:wallet:{member_id}:balance`, `rtgs:message:idempotency:{msg_id}`, `cbdc:liquidity:available_total`.
- **Apache Kafka Topics:** Consumes `settlement.cash_leg_requested.v1`, `wallet.sweep_requested.v1`; publishes `cbdc.payment_settled.v1`, `rtgs.transfer_confirmed.v1`, `cbdc.liquidity_alert.v1`.
- **Trade Settlement Service (Prompt 208):** Calls the adapter to execute real-time gross settlement for matched trades.
- **Payment Gateway Service (Prompt 212):** Bridges high-value retail payouts into the RTGS/CBDC rail.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **DvP Cash Leg Proofs:** The CBDC adapter generates cryptographic payment receipts (signed by RBI/Sponsor Bank HSM) and submits transaction hashes to `SettlementDvP.sol` (Prompt 306) on Hyperledger Besu to release tokenized shares atomically.
- **CBDC Tokenized Representation:** Where clearing participants utilize on-chain tokenized e₹ (`eINRToken.sol`) for 100% on-chain DvP, the adapter synchronizes on-chain token minting/burning with physical central bank reserves locked in RBI escrow accounts.
- **Zero PII Standard:** In accordance with DPDP Act 2023 and RBI privacy guidelines, payloads over the message bus and ledger contain only `settlement_id`, `amount_paise`, `cbdc_wallet_id_hash`, and bank routing codes (IFSC).

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/cbdc-rtgs-adapter` with enterprise layout and security hardening.
2. **Define Protobuf Schema:** Author `proto/growww/cbdc/v1/cbdc_adapter.proto` specifying `ExecuteCbdcPayment`, `ExecuteRtgsTransfer`, `GetWalletBalance`, `InitiateLiquiditySweep`, and `QueryPaymentStatus`.
3. **Generate gRPC Stubs:** Compile Protobuf contracts with `protoc-gen-go`.
4. **Design PostgreSQL Schema:** Write database migrations for `cbdc_wallets`, `cbdc_settlement_transactions`, and `iso20022_messages`.
5. **Implement ISO 20022 Message Serialization Engine:** Build XML/JSON converters for `pacs.008`, `pacs.009`, `pacs.002`, `camt.053`, and `camt.054` message schemas with strict XSD validation.
6. **Implement CloudHSM PKCS#11 Signature Provider:** Build Go cryptographic module interfacing with CloudHSM to digitally sign outbound ISO 20022 message headers and verify incoming RBI digital signatures.
7. **Configure Secure mTLS 1.3 Transport:** Establish bidirectional mTLS tunnels connecting to RBI / Sponsor Bank Host-to-Host gateways with strict cipher suites (`TLS_AES_256_GCM_SHA384`) and certificate pinning.
8. **Implement RBI e₹ CBDC API Connector:** Build REST/gRPC client implementing e₹ CBDC protocol:
 - Wallet registration and verification.
 - P2P / P2M atomic transfer execution.
 - Offline token redemption and balance synchronization.
9. **Implement RTGS 24/7 H2H Dispatcher:** Implement high-speed HTTP/2 and SFTP message transport dispatching `pacs.009` interbank settlements with automatic retry and exponential backoff.
10. **Build Distributed Idempotency Filter:** Implement Redis-based two-phase commit idempotency lock preventing duplicate payment dispatches for identical `message_id` or `settlement_id`.
11. **Implement 24/7 Liquidity Buffer Monitor:**
 - Monitor real-time e₹ pool balance vs peak intraday settlement requirements.
 - If pool drops below 25% buffer, trigger automated RTGS sweep from commercial bank reserve account.
12. **Build Inbound Webhook & Callback Receiver:** Build secure webhook endpoint ingesting asynchronous `pacs.002` settlement confirmations, verifying bank signatures, updating PostgreSQL, and publishing to Kafka `cbdc.payment_settled.v1`.
13. **Implement End-of-Day EOD Reconciliation Worker:** Automate daily parsing of `camt.053` bank statements to reconcile every individual settlement transaction against the internal cash ledger.
14. **Configure Prometheus Metrics & Alerts:** Export `cbdc_settlement_latency_seconds`, `rtgs_transfers_total`, `cbdc_liquidity_reserve_inr`, `iso20022_parsing_errors_total`.
15. **Write Comprehensive Mock & Integration Tests:** Write Go test suites simulating RBI CBDC gateway and RTGS endpoints with fault injection (timeouts, duplicate messages, signature mismatches).

## Interfaces / Contracts

### Protobuf Definition (`cbdc_adapter.proto`)
```protobuf
syntax = "proto3";

package growww.cbdc.v1;

option go_package = "growww/cbdc/v1;cbdcv1";

service CbdcRtgsAdapterService {
  rpc ExecuteCbdcTransfer (ExecuteCbdcTransferRequest) returns (ExecuteCbdcTransferResponse);
  rpc ExecuteRtgsSettlement (ExecuteRtgsSettlementRequest) returns (ExecuteRtgsSettlementResponse);
  rpc QueryPaymentStatus (QueryPaymentStatusRequest) returns (QueryPaymentStatusResponse);
  rpc GetCbdcWalletBalance (GetCbdcWalletBalanceRequest) returns (GetCbdcWalletBalanceResponse);
  rpc InitiateLiquiditySweep (InitiateLiquiditySweepRequest) returns (InitiateLiquiditySweepResponse);
}

enum SettlementRail {
  SETTLEMENT_RAIL_UNSPECIFIED = 0;
  SETTLEMENT_RAIL_CBDC_E_RUPEE = 1;
  SETTLEMENT_RAIL_RBI_RTGS_ISO20022 = 2;
}

enum PaymentStatus {
  PAYMENT_STATUS_UNSPECIFIED = 0;
  PAYMENT_STATUS_PENDING = 1;
  PAYMENT_STATUS_SETTLED_SUCCESS = 2;
  PAYMENT_STATUS_REJECTED = 3;
  PAYMENT_STATUS_TIMED_OUT = 4;
}

message ExecuteCbdcTransferRequest {
  string settlement_id = 1;
  string sender_wallet_id = 2;
  string recipient_wallet_id = 3;
  string amount_inr = 4; // Decimal string with paise precision (e.g., "150000.50")
  string purpose_code = 5; // "SEBI_EQUITY_DVP", "SGF_MARGIN_RESERVE"
  string idempotency_key = 6;
}

message ExecuteCbdcTransferResponse {
  string settlement_id = 1;
  string rbi_transaction_reference = 2;
  PaymentStatus status = 3;
  string amount_inr = 4;
  int64 settled_at_unix_ns = 5;
  string digital_signature_receipt = 6;
}

message ExecuteRtgsSettlementRequest {
  string settlement_id = 1;
  string sender_account_number = 2;
  string sender_ifsc = 3;
  string recipient_account_number = 4;
  string recipient_ifsc = 5;
  string amount_inr = 6;
  string end_to_end_id = 7; // pacs.009 EndToEndId
  string uetr = 8; // Unique End-to-end Transaction Reference (UUID)
}

message ExecuteRtgsSettlementResponse {
  string settlement_id = 1;
  string utr_number = 2; // Unique Transaction Reference from RBI
  PaymentStatus status = 3;
  string amount_inr = 4;
  int64 settled_at_unix_ns = 5;
}

message QueryPaymentStatusRequest {
  string settlement_id = 1;
  SettlementRail rail = 2;
}

message QueryPaymentStatusResponse {
  string settlement_id = 1;
  string bank_reference = 2;
  PaymentStatus status = 3;
  string error_code = 4;
  string error_message = 5;
  int64 updated_at_unix_ns = 6;
}

message GetCbdcWalletBalanceRequest {
  string wallet_id = 1;
}

message GetCbdcWalletBalanceResponse {
  string wallet_id = 1;
  string available_balance_inr = 2;
  string blocked_margin_balance_inr = 3;
  string total_balance_inr = 4;
  int64 queried_at_unix_ns = 5;
}

message InitiateLiquiditySweepRequest {
  string sweep_id = 1;
  string source_rail = 2; // "COMMERCIAL_BANK_ESCROW"
  string destination_rail = 3; // "RBI_CBDC_POOL"
  string amount_inr = 4;
}

message InitiateLiquiditySweepResponse {
  string sweep_id = 1;
  bool initiated = 2;
  string status = 3;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE cbdc_wallets (
    wallet_id VARCHAR(64) PRIMARY KEY,
    member_id UUID NOT NULL,
    wallet_type VARCHAR(30) NOT NULL, -- CC_SETTLEMENT_POOL, MEMBER_SETTLEMENT_ACCOUNT, SGF_RESERVE_POOL
    rbi_node_identifier VARCHAR(100) NOT NULL,
    current_balance_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    blocked_margin_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cbdc_settlement_transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_id VARCHAR(64) NOT NULL UNIQUE,
    rail VARCHAR(30) NOT NULL, -- CBDC_E_RUPEE, RBI_RTGS
    sender_identifier VARCHAR(100) NOT NULL,
    recipient_identifier VARCHAR(100) NOT NULL,
    amount_inr NUMERIC(18, 2) NOT NULL,
    rbi_reference_number VARCHAR(100),
    utr_number VARCHAR(50),
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING', -- PENDING, SETTLED, REJECTED, TIMED_OUT
    error_details TEXT,
    idempotency_key VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMPTZ
);

CREATE TABLE iso20022_messages (
    message_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_id VARCHAR(64) NOT NULL,
    message_type VARCHAR(30) NOT NULL, -- PACS_008, PACS_009, PACS_002, CAMT_053, CAMT_054
    direction VARCHAR(10) NOT NULL, -- OUTBOUND, INBOUND
    uetr VARCHAR(64),
    raw_payload_xml TEXT NOT NULL,
    signature_digest VARCHAR(128) NOT NULL,
    ack_received BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE cbdc_reconciliation_audit (
    reconciliation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reconciliation_date DATE NOT NULL,
    statement_total_inr NUMERIC(18, 2) NOT NULL,
    internal_ledger_total_inr NUMERIC(18, 2) NOT NULL,
    variance_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    is_reconciled BOOLEAN NOT NULL,
    discrepancy_details JSONB,
    audited_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cbdc_tx_settlement ON cbdc_settlement_transactions(settlement_id);
CREATE INDEX idx_cbdc_tx_status ON cbdc_settlement_transactions(status, created_at);
CREATE INDEX idx_iso_msg_uetr ON iso20022_messages(uetr);
```

## Security & Compliance Notes
- **FIPS 140-2 Level 3 HSM Requirement:** All cryptographic signing keys for RBI ISO 20022 communications and e₹ CBDC transactions must reside in hardware security modules with strict dual-custody access policies.
- **Idempotency & Nonce Isolation:** Every payment request enforces strict idempotency checking using Redis distributed locks and unique UETR (Universal End-to-End Transaction Reference) identifiers to guarantee zero duplicate debit risk.
- **DPDP Act 2023 & Banking Secrecy:** All ISO 20022 XML messages stored in database tables are encrypted at rest with AES-256-GCM envelope encryption. No plaintext investor PII is exposed over inter-service communications.
- **24/7 Availability SLA:** The service is architected for $99.999\%$ uptime with active-active redundant gateway deployments across dual cloud regions.

## Acceptance Criteria
- [ ] Bidirectional mTLS 1.3 connection established with RBI/Sponsor Bank endpoints with certificate pinning.
- [ ] ISO 20022 messaging engine serializes and parses `pacs.008`, `pacs.009`, and `pacs.002` with 100% XSD schema compliance.
- [ ] Digital signatures generated via CloudHSM PKCS#11 interface in $< 20\text{ms}$.
- [ ] 24/7 e₹ CBDC peer-to-peer transfers execute and confirm settlement in $< 1.5$ seconds.
- [ ] Duplicate payment requests with identical idempotency keys are detected and safely deduplicated.
- [ ] Automated liquidity buffer manager triggers RTGS sweeps when CBDC liquidity drops below 25% threshold.
- [ ] Daily `camt.053` automated reconciliation matches bank statement balances against internal ledgers with zero unflagged variance.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Standards), Prompt `109` (Secrets Management), Prompt `203` (Wallet Service), Prompt `208` (Trade Settlement Service).
- **Parallel Tasks:** Prompt `212` (Payment Gateway Service), Prompt `230` (SGF Service).
- **Downstream Blockers:** Prompt `306` (Settlement DvP Smart Contract instant liquidity integration).
