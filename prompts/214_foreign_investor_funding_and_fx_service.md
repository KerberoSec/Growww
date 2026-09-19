# 214 - Foreign-Investor Funding & FX Service (GIFT City / IFSCA Gateway)

## Purpose
The Foreign-Investor Funding & FX Service serves as the regulatory and financial gateway connecting global international investors with Indian equity markets through the GIFT City (Gujarat International Finance Tec-City) Special Economic Zone under the regulatory oversight of the International Financial Services Centres Authority (IFSCA). 

This service manages cross-border capital inflows in foreign currencies (USD, EUR, GBP, AED, SGD), interfaces with Authorized Dealer Category-I (AD Cat-I) banking partners for real-time FX rate conversions, enforces Liberalised Remittance Scheme (LRS) and FEMA compliance limits, and coordinates fund movement into segregated offshore escrow accounts. It bridges offshore fiat funding with domestic Indian equity settlement rails while strictly segregating foreign and domestic operational entities.

## What You Are Building
A high-throughput Python/FastAPI microservice (`services/gift-city-fx-service`) providing:
- **Multi-Currency Treasury Manager:** Ingestion and ledgering of international wire transfers (SWIFT MT103 / ISO 20022 `pacs.008`).
- **Real-Time FX Quote & Execution Engine:** FIX 4.4 and REST protocol adapters interfacing with institutional liquidity providers to stream and lock conversion rates (USD/INR, EUR/INR, etc.).
- **Inter-Entity Funding Escrow Coordinator:** Cryptographic reconciliation of offshore investor funds transferred to the domestic regulated clearing entity.
- **IFSCA Regulatory Limit Enforcer:** Automated validation against individual and institutional cross-border investment caps.
- **Artifacts Delivered:**
 - `services/gift-city-fx-service/app/main.py` - FastAPI application entry point.
 - `services/gift-city-fx-service/app/services/fx_engine.py` - FX rate streaming and trade lock engine.
 - `services/gift-city-fx-service/app/services/swift_parser.py` - SWIFT / ISO 20022 cross-border payment parser.
 - `services/gift-city-fx-service/app/services/escrow_manager.py` - GIFT City to domestic bank escrow coordinator.
 - `proto/growww/fx/v1/fx_service.proto` - gRPC contracts for funding and FX conversion.

## Scope Boundaries
- **In Scope:**
 - Ingestion and tracking of multi-currency deposits (USD, EUR, GBP, AED, SGD, JPY).
 - Integration with Authorized Dealer Bank APIs for real-time FX quote locking and trade settlement.
 - GIFT City offshore escrow account balance management and inter-entity ledger reconciliation.
 - Generation of statutory foreign exchange reporting data (RBI Form A2, FETERS, IFSCA returns).
- **Out of Scope / Handled Elsewhere:**
 - Domestic INR payment rails (UPI, IMPS, NEFT, RTGS) handled by Prompt 212.
 - Foreign investor onboarding and passport/sanctions verification handled by Prompt 202.
 - On-chain fractional token issuance handled by Prompt 303.

## Technology to Use
- **Primary Language & Framework:** Python 3.12+ with FastAPI, AsyncIO, and SQLAlchemy 2.0 (async).
- **Justification:** Python is selected for its extensive ecosystem supporting financial modeling, complex decimal precision calculations, native support for FIX protocol integrations (`quickfix`), and high-speed asynchronous REST/WebSocket connectivity required for streaming FX liquidity feeds.
- **Dependencies & Libraries:**
 - PostgreSQL 16+ for transaction ledgering with `asyncpg`.
 - Redis 7.2+ for sub-millisecond FX rate caching and rate lock reservations.
 - `quickfix` / `pyfixest` for institutional banking FIX protocol connectivity.
 - `lxml` and `pydantic-xml` for ISO 20022 XML parsing.
 - Apache Kafka 3.7+ for asynchronous event distribution.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Stores multi-currency investor balances, FX quotes, wire confirmations, and inter-entity settlement logs.
- **Redis 7.2:** Real-time FX quote cache, guaranteed rate locks (30-second TTL), and idempotency keys.
- **Kafka Topics:**
 - Publishes: `fx.quote.locked`, `fx.conversion.completed`, `funding.crossborder.received`, `funding.inter_entity.settled`.
 - Subscribes: `user.onboarded.international`, `order.international.placed`.
- **External Financial Gateways:** GIFT City Unit banking APIs, SWIFT Alliance Lite2 / ISO 20022 banking hubs, FX Liquidity Provider feeds.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium ledger running QBFT consensus.
- **Interaction Model:** When cross-border funds settle in GIFT City escrow, the service notifies the blockchain relayer. The relayer invokes `CrossEntitySettlement.sol` to record the synthetic INR funding credit backing the offshore investor's wallet address.
- **Zero PII Guarantee:** Only investor cryptographic public identifiers (e.g., wallet address / hashed account UUID), foreign currency amount, locked conversion rate, and transaction hashes are referenced on the permissioned ledger. No passport, country of origin, or individual banking details are stored on-chain.
- **Custody Backing:** Guarantees that on-chain minting operations for foreign investors can only proceed after physical foreign currency is converted and settled in regulated escrow accounts.

## Step-by-Step Build Instructions
1. Initialize FastAPI project structure with strict typing, configuration management, and async database session providers.
2. Define Protobuf definitions in `proto/growww/fx/v1/fx_service.proto` and compile gRPC client/server interfaces.
3. Set up PostgreSQL tables for `foreign_accounts`, `fx_quotes`, `cross_border_transfers`, and `escrow_settlements`.
4. Implement the FIX 4.4 and WebSocket client to stream live FX spot and forward rates from AD Cat-I banking partners.
5. Build the FX Quote Lock engine in Redis, providing 30-second guaranteed conversion rates with slippage protection.
6. Implement ISO 20022 `pacs.008.001.08` and SWIFT MT103 incoming wire transfer parsers with cryptographic signature validation.
7. Build the multi-currency ledger module enforcing 18-decimal fixed-point precision math (`decimal.Decimal`) to eliminate floating-point rounding errors.
8. Implement the inter-entity transfer state machine coordinating fund transfers between GIFT City banking entities and domestic clearing accounts.
9. Integrate Kafka event publishers for funding lifecycle events and consumers for international user order triggers.
10. Implement compliance rule engines for IFSCA investment limits, FEMA reporting flags, and automated sanctions re-checks.
11. Add OpenTelemetry distributed tracing and Prometheus metrics tracking FX spread margins, latency, and quote conversion rates.
12. Construct automated integration tests simulating end-to-end SWIFT deposit, FX rate lock, conversion execution, and ledger updates.

## Interfaces / Contracts

### Protobuf Definition (`fx_service.proto`)
```protobuf
syntax = "proto3";

package growww.fx.v1;

option go_package = "github.com/growww/services/gift-city-fx/gen/v1;fxv1";

service ForeignFundingFXService {
  rpc RequestFXQuote (FXQuoteRequest) returns (FXQuoteResponse);
  rpc ExecuteFXConversion (FXConversionRequest) returns (FXConversionResponse);
  rpc IngestCrossBorderTransfer (TransferIngestionRequest) returns (TransferIngestionResponse);
  rpc GetForeignAccountBalance (AccountBalanceRequest) returns (AccountBalanceResponse);
}

message FXQuoteRequest {
  string investor_id = 1;
  string source_currency = 2; // USD, EUR, GBP, AED, SGD
  string target_currency = 3; // INR
  string amount = 4; // High-precision string representation
  string direction = 5; // BUY_INR / SELL_INR
}

message FXQuoteResponse {
  string quote_id = 1;
  string source_currency = 2;
  string target_currency = 3;
  string exchange_rate = 4;
  string inverted_rate = 5;
  string gross_target_amount = 6;
  string bank_fee_inr = 7;
  string net_target_amount = 8;
  int64 expires_at_timestamp = 9;
}

message FXConversionRequest {
  string quote_id = 1;
  string investor_id = 2;
  string idempotency_key = 3;
}

message FXConversionResponse {
  string conversion_id = 1;
  string status = 2; // SETTLED / PENDING_BANK / FAILED
  string source_currency = 3;
  string source_amount = 4;
  string settled_inr_amount = 5;
  string bank_deal_reference = 6;
  int64 settled_at = 7;
}

message TransferIngestionRequest {
  string bank_wire_reference = 1;
  string swift_message_type = 2;
  string raw_wire_payload = 3;
  string sender_bic = 4;
  string receiving_iban = 5;
  string currency = 6;
  string amount = 7;
}

message TransferIngestionResponse {
  string transfer_id = 1;
  string matched_investor_id = 2;
  string status = 3; // CREDITED / HELD_COMPLIANCE / UNMATCHED
  string message = 4;
}

message AccountBalanceRequest {
  string investor_id = 1;
}

message AccountBalanceResponse {
  string investor_id = 1;
  repeated CurrencyBalance balances = 2;
}

message CurrencyBalance {
  string currency = 1;
  string total_balance = 2;
  string available_balance = 3;
  string locked_in_orders = 4;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE foreign_investor_wallets (
    wallet_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investor_id VARCHAR(64) UNIQUE NOT NULL,
    entity_jurisdiction VARCHAR(32) NOT NULL DEFAULT 'GIFT_CITY_IFSCA',
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE currency_ledger_entries (
    entry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES foreign_investor_wallets(wallet_id),
    currency VARCHAR(8) NOT NULL,
    amount NUMERIC(28, 8) NOT NULL,
    entry_type VARCHAR(32) NOT NULL CHECK (entry_type IN ('DEPOSIT', 'WITHDRAWAL', 'FX_LOCK', 'FX_SETTLE', 'FEE')),
    reference_id VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE fx_conversion_records (
    conversion_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quote_id VARCHAR(64) UNIQUE NOT NULL,
    investor_id VARCHAR(64) NOT NULL,
    source_currency VARCHAR(8) NOT NULL,
    target_currency VARCHAR(8) NOT NULL,
    exchange_rate NUMERIC(18, 8) NOT NULL,
    source_amount NUMERIC(28, 8) NOT NULL,
    target_amount_inr NUMERIC(28, 8) NOT NULL,
    ad_bank_reference VARCHAR(128),
    status VARCHAR(32) NOT NULL DEFAULT 'SETTLED',
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

## Security & Compliance Notes
- **IFSCA & FEMA Cross-Border Compliance:** All foreign exchange transactions comply with IFSCA Capital Market Regulations and RBI Master Direction on Foreign Investment in India.
- **Segregated Entity Operations:** GIFT City offshore bank accounts are strictly isolated from domestic clearing accounts, communicating only via cryptographic, audited inter-entity ledger reconciliation.
- **Sanctions & PEP Re-Screening:** Every incoming SWIFT wire is automatically passed through real-time OFAC, UN, and FATF sanctions screening before funds are credited.
- **mTLS & Secret Encryption:** Bank API credentials and SWIFT communication channels utilize mutual TLS and HSM-managed API keys.

## Acceptance Criteria
- [ ] Python FastAPI service starts cleanly with fully configured OpenAPI documentation and async connection pooling.
- [ ] FX rate engine successfully connects to simulated FIX/REST feeds and streams live quotes with sub-50ms latency.
- [ ] 30-second Redis rate locks correctly prevent race conditions and guarantee execution pricing.
- [ ] SWIFT MT103 and ISO 20022 parsers accurately ingest wire payloads with 100% field mapping fidelity.
- [ ] Financial calculations use 18-decimal fixed-point precision with zero floating-point inaccuracies.
- [ ] Kafka events are successfully published to and consumed by downstream wallet and settlement services.
- [ ] Automated test suite achieves >=85% unit and integration test coverage.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 002 (Two-Entity Structure), Prompt 005 (Foreign KYC Policy), Prompt 103 (API Standards), Prompt 110 (Inter-Entity Secure Comms).
- **Subsequent / Parallel Tasks:** Prompt 203 (Wallet Service), Prompt 215 (Reconciliation Service), Prompt 313 (Cross-Entity Ledger Bridge).
