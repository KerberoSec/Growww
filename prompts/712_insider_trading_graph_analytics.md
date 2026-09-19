# 712 - Insider Trading & UPSI Graph Analytics Engine

## Purpose
Establishes a high-dimensional graph analytics and behavioral surveillance engine to detect and prevent insider trading and illicit front-running of **Unpublished Price Sensitive Information (UPSI)**. Operating under the **SEBI (Prohibition of Insider Trading) Regulations, 2015 (PIT)** and **IFSCA Capital Market Regulations**, this service automates the maintenance of the Structured Digital Database (SDD), maps dynamic social and financial relationship graphs, and correlates abnormal trading spikes immediately preceding corporate announcements, earnings disclosures, mergers, and board resolutions.

In continuous 24/7 trading markets, material non-public information (MNPI) can be exploited across irregular trading hours. This engine constructs multi-hop entity graphs linking Designated Persons (DPs), Connected Persons, immediate relatives, corporate insiders, corporate entities, common directorships, shared IP/device networks, and fund flows to identify suspicious pre-announcement trading clusters and front-running rings.

## What You Are Building
A specialized graph intelligence and anomaly detection microservice (`services/insider-trading-graph`):
- **Structured Digital Database (SDD) & UPSI Event Tracker:** Secure, tamper-evident repository tracking all UPSI creation, sharing workflows, and non-disclosure undertakings with millisecond-accurate timestamps as mandated under SEBI PIT Regulation 3(5).
- **Heterogeneous Entity-Relation Knowledge Graph:** Graph engine combining KYC relational data, CKYC records, bank accounts, family links, corporate directorships (MCA-21 integration), shared device fingerprints, and geolocation coordinates into an active graph topology.
- **Abnormal Pre-Announcement Volume & Alpha Detector:** Quantitative statistical engine scanning trading volumes, open interest buildup, directional bias, and realized profit anomalies in rolling lookback windows ($1\text{h}, 6\text{h}, 24\text{h}, 72\text{h}$) prior to official corporate filings or public disclosures.
- **Multi-Hop Graph Proximity & Path Traversal Engine:** Graph traversal algorithms calculating path distances, clustering coefficients, and centrality scores between accounts executing abnormal pre-UPSI trades and verified Designated Persons.
- **Automated Trading Window Closure Enforcer:** Interfacing with the Pre-Trade Risk Engine (Prompt 206) to enforce algorithmic pre-trade blocks for Designated Persons and connected entities during blackout periods.

## Scope Boundaries
- **In Scope:**
 - Ingestion and maintenance of Structured Digital Database (SDD) entries with cryptographic timestamping.
 - Multi-source entity resolution and knowledge graph construction connecting PANs, family members, company directors, and device fingerprints.
 - Anomaly scoring for pre-UPSI trading: tracking abnormal volume ($Z\text{-score} > 3.0$), concentrated position build-ups, and sudden account activation.
 - Subgraph extraction and shortest-path analysis between flagged trading accounts and UPSI holders.
 - Automated generation of SEBI PIT Investigation Dossiers with visual trade-graph charts and transaction chronologies.
- **Out of Scope / Handled Elsewhere:**
 - Real-time order book micro-manipulation (spoofing/layering) detection (handled in Prompt 711).
 - Corporate action entitlement and fractional distribution processing (handled in Prompt 222).
 - Pre-trade fund balance reservations (handled in Prompt 203).

## Technology to Use
- **Graph Database & Query Engine:** Neo4j Enterprise 5+ / Memgraph with Cypher query language, supplemented by Rust `petgraph` for high-throughput in-memory subgraph traversal.
- **Data Science & Streaming Analytics:** Python 3.12+ / NetworkX / PyTorch Geometric for dynamic link prediction and community detection algorithms.
- **Stream Ingestion & Event Bus:** Apache Kafka for consuming order executions, corporate filing feeds, and SDD access logs.
- **Time-Series / Columnar Analytics:** ClickHouse 24+ for aggregating high-volume trade histories, computing baseline investor behavior, and running historical pre-announcement backtests.
- **Relational Metadata Store:** PostgreSQL 16 with Row-Level Security (RLS) and cryptographic pgcrypto extensions for SDD records and case management.
- **Justification:** Insider trading detection relies on traversing complex, multi-hop relationship webs (family, shared devices, mutual investments, directorships) overlaid with time-series trade execution logs. A dedicated graph database with Cypher allows sub-second $k$-hop path queries that are computationally prohibitive in traditional relational databases.

## Backend / Infra Touchpoints
- **Upstream Microservices:**
 - `services/corporate-actions-service` (Prompt 222): Feeds official board meeting schedules, dividend announcements, M&A filings, and earnings dates.
 - `services/kyc-aml-service` (Prompt 202): Provides verified kinship data, CKYC links, corporate shareholding structures, and PAN metadata.
 - `services/order-matching-engine` (Prompt 205): Stream of trade executions and investor volume profiles.
 - `services/admin-back-office` (Prompt 217): Compliance officer interface for SDD entry registration and Designated Person (DP) list updates.
- **Downstream Microservices:**
 - `services/risk-engine` (Prompt 206): Real-time trading window closure blacklists preventing DP order placement.
 - `services/regulatory-reporting-service` (Prompts 216 & 233): Dispatches automated statutory insider trading alerts to SEBI and IFSCA.
- **Messaging Topics (Apache Kafka):**
 - Consumes: `corporate.upsi.registered`, `corporate.filing.published`, `market.trades.settled`, `user.kyc.relations_updated`.
 - Publishes: `insider_trading.alert.generated`, `trading_window.closure.updated`, `sdd.entry.logged`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Ledger:** Hyperledger Besu permissioned consortium network operating QBFT consensus.
- **Immutable SDD State Commitment:** To comply with SEBI PIT Regulation 3(5) mandate for non-tamperable structured digital databases, a cryptographic SHA-256 Merkle root of all active SDD entries (containing salted hashes of UPSI titles, DPs, recipients, and timestamps) is committed to `StructuredDigitalDatabaseRegistry.sol` at regular 1-hour intervals.
- **On-Chain Pre-Announcement Snapshot Attestation:** When a corporate entity registers a confidential UPSI event, a time-locked cryptographic receipt is anchored on-chain. This provides mathematical, irrefutable proof of the exact millisecond when the information was created and shared internally.
- **Zero PII on Ledger:** All graph node identifiers stored on-chain or transmitted in cryptographic proofs use salted one-way hashes (`entity_hash = HMAC-SHA256(pan_or_identifier, system_secret)`), completely shielding investor identities while maintaining verifiable evidentiary integrity.

## Step-by-Step Build Instructions
1. Scaffold the `services/insider-trading-graph` service architecture containing `sdd/`, `graph/`, `detector/`, `traversal/`, and `reporting/` modules.
2. Define Protobuf definitions in `proto/growww/insider_trading/v1/insider_trading.proto` and compile service interfaces.
3. Configure PostgreSQL schema for the **Structured Digital Database (SDD)** enforcing append-only audit logs and cryptographic hash chaining per entry.
4. Set up Neo4j graph schemas with indexed node labels (`Investor`, `DesignatedPerson`, `Company`, `DeviceFingerprint`, `BankAccount`, `IPAddress`) and relationship edges (`RELATED_TO`, `DIRECTOR_OF`, `SHARED_DEVICE_WITH`, `TRANSFERRED_FUNDS_TO`, `TRADED_SECURITY`).
5. Implement the **SDD Management Module**:
 - Ingest UPSI creation events, capturing information description, nature of UPSI, names/PANs of persons sharing information, and recipient confirmations.
 - Enforce automated time-stamped digital acknowledgments with HSM digital signature verification.
6. Implement the **Entity Graph Resolution Pipeline**:
 - Continuously sync KYC relationship trees, CKYC family records, and device login fingerprints to build dynamic graph connections between accounts.
7. Implement the **Pre-Announcement Anomaly Detector**:
 - When a corporate filing or earnings release occurs, automatically extract trading activity for that ISIN during lookback windows ($1\text{h}, 6\text{h}, 24\text{h}, 72\text{h}, 1\text{w}$).
 - Calculate account-level volume anomaly scores:
     $$\text{Volume Anomaly} = \frac{V_{\text{pre-upsi}} - \mu_{V, 90\text{d}}}{\sigma_{V, 90\text{d}}}$$
 - Filter for accounts exhibiting abnormal directional positioning ($Z\text{-score} \ge 3.0$) or first-time aggressive trades in the subject security.
8. Implement the **Graph Path Traversal Algorithm**:
 - For all flagged anomalous trading accounts, execute multi-hop Cypher queries finding paths of length $k \le 4$ connecting them to any Designated Person possessing that specific UPSI.
 - Compute edge weights based on relationship strength (e.g., Immediate Relative = 1.0, Shared Device = 0.9, Common Bank Account = 0.95, 2nd-degree social link = 0.4).
9. Implement the **Composite Insider Trading Confidence Scorer**:
 - Combine graph proximity score, trading volume anomaly $Z\text{-score}$, realized/unrealized profit metric, and account inactivity factor into a composite confidence index ($0.0 - 100.0$).
10. Implement the **Automated Trading Window Closure Coordinator**:
 - Manage trading blackout calendars for all listed securities; automatically emit pre-trade block commands to `services/risk-engine` for all mapped DPs and immediate relatives.
11. Build the **SEBI PIT Regulatory Export & Visual Dossier Engine**:
 - Render interactive SVG/PDF relationship graph diagrams and chronological trade analysis dossiers formatted strictly for SEBI Enforcement Directorate and IFSCA submission.
12. Establish continuous health checks, Neo4j memory optimizations (APOC and GDS libraries), and end-to-end integration tests.

## Interfaces / Contracts

### Protobuf Definition (`insider_trading.proto`)
```protobuf
syntax = "proto3";

package growww.insider_trading.v1;

option go_package = "github.com/growww/services/insider-trading-graph/gen/v1;insidertradingv1";

service InsiderTradingGraphService {
  rpc RegisterUpsiEvent (RegisterUpsiEventRequest) returns (RegisterUpsiEventResponse);
  rpc UpdateTradingWindow (UpdateTradingWindowRequest) returns (UpdateTradingWindowResponse);
  rpc AnalyzePreAnnouncementTrades (AnalyzePreAnnouncementRequest) returns (AnalyzePreAnnouncementResponse);
  rpc GetEntityRelationshipGraph (EntityGraphRequest) returns (EntityGraphResponse);
  rpc GeneratePitInvestigationDossier (GeneratePitDossierRequest) returns (GeneratePitDossierResponse);
}

enum UpsiCategory {
  UPSI_CATEGORY_UNSPECIFIED = 0;
  FINANCIAL_RESULTS = 1;
  DIVIDENDS = 2;
  MERGERS_ACQUISITIONS = 3;
  CAPITAL_STRUCTURE_CHANGE = 4;
  KEY_MANAGERIAL_CHANGE = 5;
  MATERIAL_BUSINESS_EXPANSION = 6;
}

message UpsiEntity {
  string upsi_id = 1;
  string company_isin = 2;
  UpsiCategory category = 3;
  string nature_of_upsi = 4;
  string sharing_person_pan_hash = 5;
  repeated string recipient_pan_hashes = 6;
  int64 timestamp_shared_utc = 7;
  string sdd_chain_hash = 8;
}

message RegisterUpsiEventRequest {
  UpsiEntity upsi = 1;
}

message RegisterUpsiEventResponse {
  string sdd_entry_id = 1;
  string blockchain_receipt_hash = 2;
  bool registered = 3;
}

message UpdateTradingWindowRequest {
  string company_isin = 1;
  bool is_window_closed = 2;
  int64 closure_start_utc = 3;
  int64 closure_end_utc = 4;
  repeated string designated_person_ids = 5;
}

message UpdateTradingWindowResponse {
  bool acknowledged = 1;
  int32 blocked_accounts_count = 2;
}

message GraphPathStep {
  string from_node_id = 1;
  string from_node_type = 2; // INVESTOR / DP / DEVICE / BANK_ACCOUNT
  string relationship = 3; // IMMEDIATE_RELATIVE / SHARED_DEVICE / CO_DIRECTOR
  string to_node_id = 4;
  string to_node_type = 5;
  double edge_weight = 6;
}

message SuspiciousInsiderAlert {
  string alert_id = 1;
  string isin = 2;
  string upsi_id = 3;
  string trader_pan_hash = 4;
  string connected_dp_pan_hash = 5;
  double confidence_score = 6; // 0.0 to 100.0
  double abnormal_volume_zscore = 7;
  double estimated_profit_inr = 8;
  repeated GraphPathStep connection_path = 9;
  int64 trade_timestamp_utc = 10;
  int64 disclosure_timestamp_utc = 11;
}

message AnalyzePreAnnouncementRequest {
  string isin = 1;
  int64 announcement_timestamp_utc = 2;
  int32 lookback_hours = 3; // e.g. 24, 72, 168
}

message AnalyzePreAnnouncementResponse {
  repeated SuspiciousInsiderAlert suspicious_alerts = 1;
  int32 total_trades_evaluated = 2;
  string analysis_summary = 3;
}

message EntityGraphRequest {
  string root_pan_hash = 1;
  int32 max_hops = 2;
}

message EntityGraphResponse {
  repeated GraphPathStep graph_edges = 1;
}

message GeneratePitDossierRequest {
  string alert_id = 1;
  string compliance_officer_id = 2;
}

message GeneratePitDossierResponse {
  string dossier_id = 1;
  string dossier_download_url = 2;
  string sha256_checksum = 3;
  string blockchain_attestation_hash = 4;
}
```

### PostgreSQL & Cypher Graph Schemas
```sql
-- PostgreSQL Structured Digital Database (SDD) Schema (SEBI PIT Reg 3(5))
CREATE TABLE structured_digital_database (
    sdd_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    upsi_reference_code VARCHAR(64) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    upsi_category VARCHAR(64) NOT NULL,
    sharing_person_name VARCHAR(128) NOT NULL,
    sharing_person_pan_hash VARCHAR(64) NOT NULL,
    recipient_name VARCHAR(128) NOT NULL,
    recipient_pan_hash VARCHAR(64) NOT NULL,
    recipient_organization VARCHAR(128) NOT NULL,
    purpose_of_sharing TEXT NOT NULL,
    timestamp_shared TIMESTAMPTZ NOT NULL,
    digital_signature_hash VARCHAR(128) NOT NULL,
    previous_entry_hash VARCHAR(64) NOT NULL,
    current_entry_hash VARCHAR(64) NOT NULL,
    blockchain_tx_hash VARCHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sdd_isin ON structured_digital_database (isin, timestamp_shared DESC);
CREATE INDEX idx_sdd_recipient ON structured_digital_database (recipient_pan_hash);

-- Neo4j Cypher Schema Constraints & Graph Traversal Template
-- CREATE CONSTRAINT FOR (i:Investor) REQUIRE i.pan_hash IS UNIQUE;
-- CREATE CONSTRAINT FOR (d:Device) REQUIRE d.fingerprint IS UNIQUE;
-- CREATE CONSTRAINT FOR (b:BankAccount) REQUIRE b.account_hash IS UNIQUE;

-- Cypher Query for K-Hop Connected Trader Detection:
-- MATCH (trader:Investor {pan_hash: $trader_hash})
-- MATCH (dp:DesignatedPerson {isin: $target_isin})
-- MATCH path = shortestPath((trader)-[*1..4]-(dp))
-- RETURN path, length(path) AS distance, [r IN relationships(path) | type(r)] AS rel_types;
```

## Security & Compliance Notes
- **SEBI PIT Regulations 2015 & Structured Digital Database Mandate:** The SDD must maintain an internally chained cryptographic audit trail with no provision for post-facto modification or deletion. Any attempt to update historical SDD logs is rejected and triggers high-severity security alerts.
- **Confidentiality & Insider Access Control:** Access to raw SDD data and active UPSI investigations is strictly compartmentalized under Chinese Wall protocols. Only designated Compliance Officers holding verified hardware tokens can decrypt investigation logs.
- **Zero PII Leakage:** Kinship graphs and corporate links transmitted across analytical workers are pseudonymized with cryptographic salts, ensuring compliance with the Digital Personal Data Protection (DPDP) Act 2023.
- **Evidentiary Forensic Standards:** Graph traversal outputs and correlation chronologies must be cryptographically hashed and signed to meet Indian Evidence Act requirements for admissible electronic records.

## Acceptance Criteria
- [ ] SDD audit trail maintains unbroken cryptographic hash chaining for 100% of recorded UPSI events.
- [ ] Automated trading window closure instantly locks order placement for all mapped DPs across web, mobile, and API channels within $< 50\text{ms}$.
- [ ] Graph traversal engine successfully resolves connected entity paths up to 4 hops in $< 100\text{ms}$ across a graph containing $\ge 1,000,000$ nodes.
- [ ] Accurately identifies $100\%$ of synthetic front-running rings (e.g. DP $\to$ Spouse $\to$ Relative Account $\to$ Outsized Call/Share Purchase) in historical backtesting suites.
- [ ] Calculates volume anomaly $Z\text{-score}$ and directional concentration index across all trades in the 72 hours prior to corporate disclosures with zero pipeline drops.
- [ ] Produces SEBI PIT compliant digital investigation dossiers with embedded visual graph diagrams and verifiable blockchain attestation receipts.
- [ ] Test coverage across graph traversers, SDD validators, and anomaly models exceeds $\ge 90\%$.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 004 (KYC/AML Policy), Prompt 202 (KYC/AML Service), Prompt 205 (Order Matching Engine), Prompt 222 (Corporate Actions Service).
- **Subsequent / Parallel Tasks:** Prompt 711 (Market Surveillance Engine), Prompt 713 (Circuit Breaker Coordinator), Prompt 233 (Continuous 24x7 Regulatory Reporting).
