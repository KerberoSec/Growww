# 714 - Cross-Chain AML & Blockchain Forensics Screener (Bitcoin, Ethereum, Solana)

## Purpose
Establishes an automated, institutional-grade crypto forensics and real-time Anti-Money Laundering (AML) transaction screening engine. Operating at the GIFT City International Gateway entity, this service provides continuous surveillance, origin-of-funds verification, and taint analysis for all inbound cryptocurrency funding deposits (Bitcoin, Ethereum / EVM ERC-20, and Solana / SPL tokens) prior to fiat conversion and allocation into tokenized Indian equity investment structures.

In compliance with **IFSCA Anti-Money Laundering and Counter-Financing of Terrorism (AML/CFT) Guidelines**, **FATF Recommendations 15 and 16 (Virtual Assets and Travel Rule)**, and **US OFAC Crypto Sanctions Compliance Guidance**, this service eliminates contamination risks by screening every deposit against global sanctions lists, darknet marketplace clusters, terrorist financing wallets, ransomware syndicates, and privacy-obfuscation protocols (such as Tornado Cash, Sinbad, Blender.io, and Wasabi). Inbound funds linked to high-risk or sanctioned clusters are automatically isolated into quarantine custody wallets, blocked from platform ledgers, and subjected to immediate on-chain account freezes on the permissioned settlement blockchain.

## What You Are Building
A high-throughput, resilient crypto forensics and compliance microservice (`services/crypto-forensics-screener`):
- **Multi-Vendor Blockchain Intelligence Aggregator:** Unified client orchestrator querying Chainalysis (KYT/Reactor), Elliptic (Navigator/Lens), and TRM Labs APIs in parallel with deterministic fallback routing, weighted risk scoring consensus, and circuit breaker protection.
- **Cross-Chain Transaction Unpacker & Taint Evaluator:** Ingests raw deposit events across Bitcoin (UTXO graph traversal), Ethereum (EVM internal transactions, smart contract traces, ERC-20 transfers), and Solana (SPL token transfers and instruction logs), evaluating direct (1-hop) and indirect (up to 5-hop) exposure percentages.
- **Automated Sanctions & Mixer Quarantine Coordinator:** Deterministic state machine that isolates flagged deposits, halts balance crediting in the Wallet Service (Prompt 203), and issues cryptographic freeze directives to the on-chain Compliance Registry (Prompt 305).
- **IVMS 101 Travel Rule & Counterparty VASP Verifier:** Gathers originator wallet ownership proofs, verifies counterparty Virtual Asset Service Provider (VASP) licensing status, and formats Travel Rule data packages conforming to the interVASP IVMS 101 messaging standard.
- **Forensic Case Management & SAR Dossier Generator:** Compiles tamper-evident forensic evidence dossiers containing visual transaction graph paths, vendor risk scores, cluster attribution metadata, and millisecond timestamps for regulatory reporting to IFSCA and FIU-IND.

## Scope Boundaries
- **In Scope:**
  - Real-time pre-credit AML screening of inbound crypto deposits across Bitcoin (BTC), Ethereum (ETH, USDC, USDT), and Solana (SOL, USDC).
  - Integration with Chainalysis, Elliptic, and TRM Labs screening endpoints with resilient multi-vendor consensus.
  - Detection and scoring of risk categories: OFAC SDN Sanctions, Terrorist Financing, Stolen Assets, Mixers/Tumblers, Darknet Markets, Ransomware, High-Risk Exchanges, and Scams.
  - Automated deposit quarantine routing, balance hold enforcement, and compliance alert generation.
  - Real-time dispatch of on-chain account freezing commands to Hyperledger Besu.
  - Persistent caching of address reputations and UTXO/transaction taint scores.
  - Export of standardized SAR/STR forensic evidence packages and IVMS 101 Travel Rule payloads.
- **Out of Scope / Handled Elsewhere:**
  - Private key management, MPC key signing, and multi-sig vault sweeps (handled in Prompts 213 and 319).
  - Crypto-to-fiat FX conversion and cross-border bank settlement (handled in Prompt 214).
  - Domestic Indian investor fiat onboarding and Aadhaar/PAN KYC (handled in Prompts 004 and 202).
  - Secondary market order matching and trading execution (handled in Prompt 205).
  - General fiat transaction monitoring for smurfing and structuring (handled in Prompt 704).

## Technology to Use
- **Core Microservice:** Go 1.22+ for high-concurrency asynchronous API aggregation, low memory footprint, and sub-millisecond internal routing.
- **Stream Ingestion & Event Bus:** Apache Kafka for consuming real-time blockchain node deposit events and publishing compliance screening decisions.
- **In-Memory Risk Matrix & Address Cache:** Redis 7 (Cluster) for sub-millisecond address reputation caching, rate-limit token buckets, and active screening session locks.
- **Forensics Time-Series & Trace Graph Store:** ClickHouse 24+ for storing raw multi-hop transaction traces, cluster attribution logs, and historical forensics queries.
- **Relational Case & Audit Database:** PostgreSQL 16 with Row-Level Security (RLS) for compliance cases, whitelist overrides, vendor API logs, and quarantine state tracking.
- **External Forensics APIs:** Chainalysis KYT API, Elliptic Navigator API v2, TRM Labs Screening & Forensics API v2.
- **Secrets & Key Custody:** HashiCorp Vault / AWS Secrets Manager with HSM backing for external API credentials and mTLS client certificates.
- **Justification:** Screening crypto deposits during active customer funding flows mandates parallel querying across multiple blockchain intelligence vendors with a p99 SLA under 1,500ms. Go provides lightweight goroutine concurrency, predictable low-latency execution without heavy runtime overhead, and strong compile-time type safety.

## Backend / Infra Touchpoints
- **Upstream Microservices:**
  - `services/custodian-adapter` (Prompt 213): Emits deposit detection events upon blockchain node block confirmations.
  - `services/custody-bridge` (Prompt 319): Forwards institutional cross-chain deposit intents and address registrations.
- **Downstream Microservices:**
  - `services/wallet-service` (Prompt 203): Ingests screening approval to credit pending user wallet balances.
  - `services/foreign-funding-fx` (Prompt 214): Authorizes conversion of cleared crypto deposits into USD/INR.
  - `services/risk-engine` (Prompt 206): Blocks trading activity if an existing active investor wallet receives tainted funds.
  - `services/regulatory-reporting` (Prompts 216 & 233): Ingests forensic SAR dossiers for regulatory filing.
  - `services/admin-back-office` (Prompts 217, 604 & 605): Feeds compliance officer investigation workbench.
- **Messaging Topics (Apache Kafka):**
  - Consumes: `crypto.deposit.detected.v1`, `crypto.address.screening_requested.v1`.
  - Publishes: `crypto.deposit.screened.v1`, `crypto.deposit.quarantined.v1`, `compliance.crypto.freeze_command.v1`, `compliance.crypto.sar_alert.v1`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus.
- **Automated On-Chain Compliance Freezing:** When an incoming deposit exhibits direct exposure to sanctioned entities (OFAC SDN), privacy mixers (Tornado Cash), or a composite risk score $\ge 85$, the service publishes a high-priority `compliance.crypto.freeze_command.v1` event. The on-chain Compliance Relayer (Prompt 305) executes an immediate transaction calling `ComplianceRegistry.freezeInvestor(address investorAddress, bytes32 reasonHash)`. This instantly blocks the investor from minting, transferring, or trading fractional tokenized equity units on-chain.
- **Forensic Attestation & Evidence Anchoring:** For every screened deposit, a deterministic SHA-256 Merkle root of the combined forensic report (including vendor risk scores, source txids, UTXO hops, attribution tags, and timestamps) is anchored to `ProofOfReserveRegistry.sol` / `ComplianceAuditRegistry.sol` on Hyperledger Besu. This creates an immutable, legally verifiable audit trail for regulatory examinations.
- **Zero PII on Ledger:** All blockchain records strictly reference pseudonymized investor ledger addresses (`0x...`), external transaction hashes, and cryptographic evidence hashes. No investor names, email addresses, or KYC identity numbers are ever committed to the ledger.

## Step-by-Step Build Instructions
1. Scaffold the `services/crypto-forensics-screener` repository in Go 1.22+ following the standard workspace layout (Prompt 106) with linting, formatting, and strict type checking (Prompt 107).
2. Generate Go and Python gRPC stubs from the Protobuf definitions in `proto/growww/crypto_forensics/v1/crypto_forensics.proto`.
3. Design and execute PostgreSQL 16 database migrations for tables `crypto_forensics_screenings`, `crypto_quarantined_deposits`, `forensics_vendor_audit_log`, and `forensics_address_risk_cache`.
4. Configure ClickHouse schema and table engines for high-volume storage of multi-hop transaction traces and forensic graph events.
5. Implement vendor client adapters (`adapters/chainalysis`, `adapters/elliptic`, `adapters/trmlabs`) with exponential backoff, circuit breaking, and mock modes for offline unit testing.
6. Implement the **Cross-Chain Transaction Unpacker**:
   - Bitcoin: Traverse UTXO input graph, identifying sending addresses, change outputs, and multi-sig scripts.
   - Ethereum: Parse EVM transaction logs, internal contract call traces, and ERC-20 `Transfer` events.
   - Solana: Parse transaction instruction data, inner instructions, and SPL token program balances.
7. Implement the **Multi-Vendor Consensus & Risk Aggregator**:
   - Query configured vendors concurrently using Go goroutines with a strict 1,200ms timeout.
   - Normalize disparate vendor risk scores (0-100 scale) and risk categories into canonical taxonomy.
   - Compute the composite risk score using weighted conservative aggregation:
     $$\text{Composite Risk} = \max(\text{Risk}_{\text{OFAC}}, \text{Risk}_{\text{Mixer}}, 0.5 \cdot \text{Risk}_{\text{Chainalysis}} + 0.3 \cdot \text{Risk}_{\text{Elliptic}} + 0.2 \cdot \text{Risk}_{\text{TRM}})$$
8. Implement the **Deterministic Screening Decision Engine**:
   - `CLEARED` (Score $< 30$ and Zero Sanctions/Mixer exposure): Emit approval event to credit wallet.
   - `MANUAL_REVIEW` ($30 \le \text{Score} < 75$ or indirect taint $> 10\%$): Place deposit in compliance review hold.
   - `QUARANTINED` ($\text{Score} \ge 75$, direct Sanctions hit, or Mixer exposure): Move funds to quarantine custody, block wallet, and trigger alert.
9. Implement the **Quarantine & Automated Freeze Dispatcher**:
   - Upon `QUARANTINED` decision, emit `compliance.crypto.freeze_command.v1` to trigger on-chain freeze on Hyperledger Besu.
   - Notify `services/wallet-service` to freeze pending balance and prevent outbound withdrawals.
10. Implement the **IVMS 101 Travel Rule Generator**:
    - Validate counterparty VASP decentralized identifiers (DIDs) against verified VASP directories (e.g. TRM VASP Directory, Global Travel Rule Network).
    - Format sender and beneficiary metadata into IVMS 101 compliant JSON/XML payloads for transfers $\ge \$1,000$ equivalent.
11. Build the **Forensics Dossier Export Service**:
    - Serialize complete transaction forensics graphs, attribution tags, and vendor responses into signed PDF/JSON packages for regulatory submission to IFSCA / FIU-IND.
12. Configure Prometheus metrics (`crypto_screenings_total`, `crypto_screening_latency_ms`, `crypto_vendor_errors_total`, `crypto_deposits_quarantined_total`), OpenTelemetry distributed tracing, and structured JSON audit logging.

## Interfaces / Contracts

### Protobuf Definition (`crypto_forensics.proto`)
```protobuf
syntax = "proto3";

package growww.crypto_forensics.v1;

option go_package = "github.com/growww/services/crypto-forensics-screener/gen/v1;cryptoforensicsv1";

service CryptoForensicsService {
  rpc ScreenDeposit (ScreenDepositRequest) returns (ScreenDepositResponse);
  rpc ScreenAddress (ScreenAddressRequest) returns (ScreenAddressResponse);
  rpc GetScreeningReport (GetScreeningReportRequest) returns (GetScreeningReportResponse);
  rpc SubmitAnalystDecision (SubmitAnalystDecisionRequest) returns (SubmitAnalystDecisionResponse);
}

enum BlockchainNetwork {
  NETWORK_UNSPECIFIED = 0;
  BITCOIN_MAINNET = 1;
  ETHEREUM_MAINNET = 2;
  SOLANA_MAINNET = 3;
}

enum ScreeningDecision {
  DECISION_UNSPECIFIED = 0;
  DECISION_CLEARED = 1;
  DECISION_MANUAL_REVIEW = 2;
  DECISION_QUARANTINED = 3;
  DECISION_REJECTED = 4;
}

enum RiskCategory {
  RISK_NONE = 0;
  SANCTIONS_OFAC = 1;
  TERRORIST_FINANCING = 2;
  MIXER_PRIVACY_PROTOCOL = 3;
  DARKNET_MARKET = 4;
  RANSOMWARE = 5;
  STOLEN_FUNDS_HACK = 6;
  SCAM_FRAUD = 7;
  HIGH_RISK_EXCHANGE = 8;
  GAMBLING_UNLICENSED = 9;
}

message ScreenDepositRequest {
  string deposit_id = 1;
  string user_id = 2;
  BlockchainNetwork network = 3;
  string asset_symbol = 4; // e.g., "BTC", "ETH", "USDC", "SOL"
  string transaction_hash = 5;
  string source_address = 6;
  string destination_vault_address = 7;
  string deposit_amount = 8; // Decimal formatted string
  int64 block_height = 9;
  int64 block_timestamp_utc = 10;
}

message VendorRiskScore {
  string vendor_name = 1; // "CHAINALYSIS", "ELLIPTIC", "TRM_LABS"
  double risk_score = 2; // 0.0 to 100.0
  repeated RiskCategory identified_risks = 3;
  int32 direct_hop_distance = 4;
  double direct_taint_percentage = 5;
  double indirect_taint_percentage = 6;
  string vendor_reference_id = 7;
  int64 evaluation_latency_ms = 8;
}

message ScreenDepositResponse {
  string screening_id = 1;
  string deposit_id = 2;
  ScreeningDecision decision = 3;
  double composite_risk_score = 4;
  repeated RiskCategory primary_risk_factors = 5;
  repeated VendorRiskScore vendor_scores = 6;
  bool automated_freeze_triggered = 7;
  string cryptographic_evidence_hash = 8;
  int64 screened_at_utc = 9;
}

message ScreenAddressRequest {
  BlockchainNetwork network = 1;
  string address = 2;
  string asset_symbol = 3;
}

message ScreenAddressResponse {
  string address = 1;
  BlockchainNetwork network = 2;
  double risk_score = 3;
  ScreeningDecision recommended_action = 4;
  repeated RiskCategory risk_categories = 5;
  bool is_sanctioned = 6;
  bool is_mixer = 7;
  string cluster_entity_name = 8;
  int64 last_updated_utc = 9;
}

message GetScreeningReportRequest {
  string screening_id = 1;
}

message GetScreeningReportResponse {
  string screening_id = 1;
  string deposit_id = 2;
  string user_id = 3;
  ScreenDepositRequest original_request = 4;
  ScreenDepositResponse screening_result = 5;
  string forensic_graph_json = 6;
  string travel_rule_ivms101_payload = 7;
  string analyst_notes = 8;
}

message SubmitAnalystDecisionRequest {
  string screening_id = 1;
  string compliance_officer_id = 2;
  ScreeningDecision final_decision = 3;
  string justification = 4;
  string mfa_token = 5;
}

message SubmitAnalystDecisionResponse {
  bool success = 1;
  string screening_id = 2;
  ScreeningDecision applied_decision = 3;
  int64 updated_at_utc = 4;
}
```

### PostgreSQL Database Schema
```sql
CREATE TYPE blockchain_network_enum AS ENUM (
    'BITCOIN_MAINNET',
    'ETHEREUM_MAINNET',
    'SOLANA_MAINNET'
);

CREATE TYPE screening_decision_enum AS ENUM (
    'DECISION_CLEARED',
    'DECISION_MANUAL_REVIEW',
    'DECISION_QUARANTINED',
    'DECISION_REJECTED'
);

CREATE TABLE crypto_forensics_screenings (
    screening_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deposit_id VARCHAR(64) NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    network blockchain_network_enum NOT NULL,
    asset_symbol VARCHAR(16) NOT NULL,
    transaction_hash VARCHAR(128) NOT NULL,
    source_address VARCHAR(128) NOT NULL,
    destination_vault_address VARCHAR(128) NOT NULL,
    deposit_amount NUMERIC(36, 18) NOT NULL,
    composite_risk_score NUMERIC(5, 2) NOT NULL,
    decision screening_decision_enum NOT NULL,
    is_sanctioned_hit BOOLEAN NOT NULL DEFAULT FALSE,
    is_mixer_hit BOOLEAN NOT NULL DEFAULT FALSE,
    direct_taint_pct NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    indirect_taint_pct NUMERIC(5, 2) NOT NULL DEFAULT 0.00,
    primary_cluster_name VARCHAR(128),
    evidence_hash CHAR(64) NOT NULL,
    on_chain_freeze_tx_hash VARCHAR(66),
    on_chain_attestation_tx_hash VARCHAR(66),
    analyst_id UUID,
    analyst_justification TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE crypto_quarantined_deposits (
    quarantine_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    screening_id UUID NOT NULL REFERENCES crypto_forensics_screenings(screening_id),
    user_id UUID NOT NULL,
    network blockchain_network_enum NOT NULL,
    asset_symbol VARCHAR(16) NOT NULL,
    deposit_amount NUMERIC(36, 18) NOT NULL,
    quarantine_vault_address VARCHAR(128) NOT NULL,
    source_tx_hash VARCHAR(128) NOT NULL,
    quarantine_reason VARCHAR(128) NOT NULL,
    sar_filing_reference VARCHAR(64),
    is_released BOOLEAN NOT NULL DEFAULT FALSE,
    release_authorized_by UUID,
    released_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE forensics_vendor_audit_log (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    screening_id UUID NOT NULL REFERENCES crypto_forensics_screenings(screening_id),
    vendor_name VARCHAR(32) NOT NULL,
    endpoint_called VARCHAR(256) NOT NULL,
    http_status_code INT NOT NULL,
    raw_response_payload JSONB NOT NULL,
    vendor_risk_score NUMERIC(5, 2) NOT NULL,
    latency_ms INT NOT NULL,
    queried_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE forensics_address_risk_cache (
    cache_key VARCHAR(160) PRIMARY KEY, -- network:address
    network blockchain_network_enum NOT NULL,
    address VARCHAR(128) NOT NULL,
    risk_score NUMERIC(5, 2) NOT NULL,
    risk_categories TEXT[] NOT NULL DEFAULT '{}',
    cluster_entity_name VARCHAR(128),
    is_sanctioned BOOLEAN NOT NULL DEFAULT FALSE,
    is_mixer BOOLEAN NOT NULL DEFAULT FALSE,
    vendor_consensus JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cfs_user ON crypto_forensics_screenings (user_id, created_at DESC);
CREATE INDEX idx_cfs_txhash ON crypto_forensics_screenings (transaction_hash);
CREATE INDEX idx_cfs_decision ON crypto_forensics_screenings (decision, created_at DESC);
CREATE INDEX idx_farc_expires ON forensics_address_risk_cache (expires_at);
```

### Kafka Event Schemas

#### Topic: `crypto.deposit.screened.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CryptoDepositScreenedEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "screening_id": { "type": "string", "format": "uuid" },
    "deposit_id": { "type": "string" },
    "user_id": { "type": "string", "format": "uuid" },
    "network": { "type": "string", "enum": ["BITCOIN_MAINNET", "ETHEREUM_MAINNET", "SOLANA_MAINNET"] },
    "asset_symbol": { "type": "string" },
    "amount": { "type": "string" },
    "transaction_hash": { "type": "string" },
    "source_address": { "type": "string" },
    "decision": { "type": "string", "enum": ["DECISION_CLEARED", "DECISION_MANUAL_REVIEW", "DECISION_QUARANTINED", "DECISION_REJECTED"] },
    "composite_risk_score": { "type": "number", "minimum": 0.0, "maximum": 100.0 },
    "evidence_hash": { "type": "string" },
    "timestamp_utc": { "type": "integer" }
  },
  "required": ["event_id", "screening_id", "deposit_id", "user_id", "network", "asset_symbol", "amount", "transaction_hash", "source_address", "decision", "composite_risk_score", "evidence_hash", "timestamp_utc"]
}
```

#### Topic: `crypto.deposit.quarantined.v1`
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "CryptoDepositQuarantinedEvent",
  "type": "object",
  "properties": {
    "event_id": { "type": "string", "format": "uuid" },
    "quarantine_id": { "type": "string", "format": "uuid" },
    "screening_id": { "type": "string", "format": "uuid" },
    "user_id": { "type": "string", "format": "uuid" },
    "network": { "type": "string" },
    "asset_symbol": { "type": "string" },
    "amount": { "type": "string" },
    "source_tx_hash": { "type": "string" },
    "quarantine_vault_address": { "type": "string" },
    "risk_reasons": { "type": "array", "items": { "type": "string" } },
    "on_chain_freeze_dispatched": { "type": "boolean" },
    "timestamp_utc": { "type": "integer" }
  },
  "required": ["event_id", "quarantine_id", "screening_id", "user_id", "network", "asset_symbol", "amount", "source_tx_hash", "quarantine_vault_address", "risk_reasons", "on_chain_freeze_dispatched", "timestamp_utc"]
}
```

## Security & Compliance Notes
- **IFSCA Virtual Asset Regulatory Mandate:** Enforces 100% compliance with IFSCA AML/CFT guidelines, ensuring no unverified, illicit, or tainted crypto enters the regulated capital investment pipeline.
- **Zero-Tolerance OFAC Sanctions Policy:** Any direct or indirect connection ($\ge 1\%$ exposure within 3 hops) to OFAC Specially Designated Nationals (SDNs) triggers automatic deposit quarantine and platform-wide account freezing without manual intervention.
- **Mixer & Privacy Protocol Ban:** Any interaction with privacy-enhancing protocols (Tornado Cash, Railgun, Sinbad, Blender, Wasabi CoinJoin) is categorized as `DECISION_QUARANTINED` by default.
- **Multi-Vendor Redundancy & Anti-Collusion:** Prevents vendor bias and single-point-of-failure outages by requiring score validation across at least two independent forensics providers (Chainalysis, Elliptic, TRM Labs).
- **Dual-Control Quarantine Release:** Releasing funds from quarantine custody requires dual cryptographic authorization from the Head of Compliance and Chief Risk Officer via Multi-Factor Authentication (MFA) and HSM-backed digital signatures.
- **Data Protection & PII Isolation:** Forensics vendor queries only transmit public blockchain addresses and transaction hashes. No investor PII (name, email, PAN, Aadhaar, Passport) is ever shared with external analytics providers.

## Acceptance Criteria
- [ ] End-to-end deposit screening completes with p99 latency $< 1,500\text{ms}$ under standard load.
- [ ] Accurately identifies 100% of test deposit transactions linked to known OFAC SDN addresses and Tornado Cash mixer contracts across Bitcoin, Ethereum, and Solana test suites.
- [ ] Successfully queries Chainalysis, Elliptic, and TRM Labs in parallel and gracefully falls back to available providers if any single vendor experiences an outage.
- [ ] Automatically emits `compliance.crypto.freeze_command.v1` and dispatches on-chain freeze call to `ComplianceRegistry.sol` on Hyperledger Besu for deposits with risk score $\ge 85$.
- [ ] Quarantined funds are successfully locked in designated custody vault addresses with zero balance credited to `services/wallet-service`.
- [ ] Generates valid IVMS 101 compliant Travel Rule message payloads for crypto deposits exceeding $\$1,000$ USD equivalent.
- [ ] ClickHouse logs 100% of multi-hop transaction graph paths and vendor raw responses with zero dropped events across 10,000 synthetic test transactions.
- [ ] Code coverage for unit, integration, and mock vendor screening tests exceeds $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 005 (Foreign KYC/AML Policy), Prompt 103 (API Standards), Prompt 104 (Kafka Standards), Prompt 203 (Wallet Account Service), Prompt 213 (Custodian Depository Integration), Prompt 305 (Transfer Compliance Hooks).
- **Parallel Tasks:** Prompt 214 (Foreign Investor Funding & FX Service), Prompt 703 (Global Sanctions & PEP Screening), Prompt 319 (Institutional Custody Bridge).
- **Downstream Dependents:** Prompt 216 (Regulatory Reporting Service), Prompt 233 (Continuous 24x7 Regulatory Reporting), Prompt 604 / 605 (Admin & Compliance Review Consoles).
