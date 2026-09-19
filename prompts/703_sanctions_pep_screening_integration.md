# 703 - Global Sanctions & Politically Exposed Persons (PEP) Screening Integration

## Purpose
Integrates an automated, low-latency sanctions and Politically Exposed Persons (PEP) screening engine across both domestic Indian and international (GIFT City / IFSCA) investor onboarding and transaction lifecycles. Protects the Growww platform from facilitating transactions involving sanctioned individuals, organizations, or high-risk political figures by screening against United Nations Security Council (UNSC), US Office of Foreign Assets Control (OFAC SDN), European Union (EU CFSP), UK HM Treasury (HMT), and Indian Ministry of Home Affairs (UAPA / FICN / Banned Entities) lists in strict compliance with the Prevention of Money Laundering Act (PMLA 2002), FATF Recommendations, and SEBI AML Master Circulars.

## What You Are Building
- **Sanctions & PEP Screening Service:** `services/sanctions-screener`, a high-throughput microservice providing sub-50ms REST/gRPC endpoints for synchronous pre-onboarding checks and asynchronous Kafka event processing.
- **Fuzzy String & Phonetic Matching Engine:** SIMD-accelerated name-matching pipeline combining Jaro-Winkler distance, Levenshtein edit distance, Double Metaphone phonetic encoding, and Indian/International name variant alias dictionaries.
- **Automated Watchlist Ingestion Pipeline:** Daily differential synchronization worker pulling updated datasets from OpenSanctions, OFAC, UN, and Ministry of Home Affairs feeds with cryptographic hash verification.
- **Daily Batch Re-Screening Worker:** Distributed Celery/Kafka worker that re-screens the entire registered investor database against updated watchlists within a 2-hour nightly batch window.
- **On-Chain Freeze Dispatcher:** Automated event relayer that broadcasts urgent freeze commands to on-chain compliance registries upon confirmed sanctions matches.

## Scope Boundaries
- **In Scope:**
 - Ingestion and normalization of UN, OFAC, EU, UK, and Indian UAPA sanctions and PEP datasets.
 - Multi-attribute fuzzy matching (Full Name, Date of Birth, Country/Nationality, Passport/PAN hash).
 - Synchronous pre-onboarding screening API and pre-deposit/pre-trade screening hooks.
 - Asynchronous daily batch re-screening of the entire active investor base.
 - False-positive reduction heuristics, whitelisting rules, and compliance analyst alert queues.
- **Out of Scope / Handled Elsewhere:**
 - Primary identity document OCR and biometric liveness verification (handled in Prompt 202).
 - Manual compliance case investigation UI (handled in Prompts 604 and 605).
 - On-chain smart contract whitelist implementation (handled in Prompt 305).

## Technology to Use
- **Core Service:** Python 3.12 with FastAPI for async REST endpoints and gRPC interfaces.
- **Matching Algorithms:** `rapidfuzz` (C++ SIMD-accelerated Levenshtein/Jaro-Winkler) and `jellyfish` for phonetic Double Metaphone transformations.
- **Watchlist Sources:** OpenSanctions / World-Check / LexisNexis Bridger API feeds with automated JSON/XML parser pipelines.
- **Data Stores & Caching:** Redis 7+ for low-latency indexed name n-gram caches and PostgreSQL 16 for audit records, match logs, and whitelisting overrides.
- **Task Orchestration:** Celery with Redis broker and Kafka consumers for high-volume streaming events.
- **Justification:** `rapidfuzz` provides sub-millisecond fuzzy matching performance across 500,000+ sanctions records in memory, eliminating expensive SaaS API round-trip latencies during critical investor onboarding and order placement flows.

## Backend / Infra Touchpoints
- **Microservices:** KYC Service (202), User Service (201), Order Service (204), Wallet Service (203), Admin Service (217).
- **Messaging / Queues:** Kafka topics `sanctions.screen.requested`, `sanctions.match.alerted`, and `compliance.account.freeze_command`.
- **Datastores:** PostgreSQL 16 (`sanctions_db`), Redis 7 in-memory n-gram search index, MinIO/S3 for raw watchlist snapshot archives.

## Blockchain Interaction
Automated on-chain compliance freeze hooks and validator awareness:
- **Instant Ledger Freeze:** When an active investor matches a sanctioned entity with $\ge 95\%$ match confidence on mandatory lists (OFAC/UN/UAPA), the service automatically emits a `compliance.account.freeze_command` to the Transaction Relayer (Prompt 305).
- **Smart Contract Execution:** The relayer executes an on-chain transaction calling `ComplianceRegistry.freezeInvestor(address investorAddress, bytes32 reasonHash)`. This immediately blocks the wallet from transferring, trading, receiving, or redeeming any `DigitalSecurityToken` units on the Hyperledger Besu ledger.
- **Proof of Action:** The transaction hash and block receipt are cryptographically committed to the compliance audit record (Prompt 218) to satisfy statutory regulatory reporting to SEBI and FIU-IND.

## Step-by-Step Build Instructions
1. Scaffold the `services/sanctions-screener` repository following standard layout (Prompt 106) with linting and type checking (Prompt 107).
2. Design the PostgreSQL relational schema for `watchlists`, `sanctioned_entities`, `pep_records`, `screening_logs`, and `whitelist_exceptions`.
3. Implement the watchlist ingestion pipeline (`ingestion/`): fetch, validate cryptographic checksums, parse, and normalize UN, OFAC, EU, UK, and Indian MHA XML/JSON datasets into standard schema models.
4. Implement the string normalization pipeline: diacritic removal, whitespace collapsing, token reordering, and Double Metaphone phonetic encoding.
5. Build the in-memory indexed search engine using Redis 7 and `rapidfuzz`, indexing name tokens, transliterations, and aliases for rapid candidate retrieval.
6. Implement the composite scoring algorithm: compute weighted similarity scores combining name score (60%), date of birth match (20%), nationality/country match (10%), and government ID/PAN match (10%).
7. Create the synchronous REST `/api/v1/screen/individual` and `/api/v1/screen/entity` endpoints with a p99 latency target $< 30\text{ms}$.
8. Implement Kafka consumer listening to `user.registered`, `kyc.submitted`, and `funds.deposit_initiated` events to trigger real-time asynchronous screening.
9. Build the distributed Celery batch re-screening worker that iterates through active user accounts nightly, benchmarking throughput against a minimum 1,000 checks/second.
10. Implement alert escalation workflows: if score $\ge 90\%$, flag account as `PENDING_REVIEW` and emit alert to Kafka `sanctions.match.alerted`; if score $\ge 95\%$ on strict OFAC/UN lists, trigger automated account lock and ledger freeze.
11. Implement the analyst override and whitelisting API, allowing compliance officers to mark verified false positives with documented justification and expiry dates.
12. Instrument OpenTelemetry tracing, Prometheus metrics (`sanctions_checks_total`, `sanctions_match_rate`, `screening_latency_ms`), and structured JSON logging without exposing plaintext PII.
13. Write comprehensive unit and integration tests using synthetic datasets including known international PEPs, complex phonetic name variations, and edge cases.

## Interfaces / Contracts
```protobuf
// schemas/sanctions/v1/sanctions_service.proto
syntax = "proto3";
package growww.sanctions.v1;

service SanctionsService {
  rpc ScreenIndividual (ScreenIndividualRequest) returns (ScreenIndividualResponse);
  rpc BatchScreen (BatchScreenRequest) returns (BatchScreenResponse);
}

message ScreenIndividualRequest {
  string request_id = 1;
  string full_name = 2;
  string date_of_birth = 3; // YYYY-MM-DD
  string nationality_iso2 = 4;
  string resident_country_iso2 = 5;
  string government_id_hash = 6; // SHA-256 of PAN/Passport
  string entity_id = 7; // User ID or Investor UUID
}

message SanctionsMatch {
  string match_id = 1;
  string watchlist_source = 2; // e.g. "OFAC_SDN", "UN_CONSOLIDATED", "MHA_UAPA", "PEP_GLOBAL"
  string matched_name = 3;
  double similarity_score = 4; // 0.0 to 1.0
  string list_reference_id = 5;
  string risk_category = 6; // "SANCTIONS", "PEP_TIER_1", "TERRORIST_FINANCING"
  bool auto_freeze_triggered = 7;
}

message ScreenIndividualResponse {
  string request_id = 1;
  string screening_status = 2; // "CLEARED", "POTENTIAL_MATCH", "CONFIRMED_SANCTIONS"
  double highest_match_score = 3;
  repeated SanctionsMatch matches = 4;
  int64 screened_at_timestamp = 5;
}
```

```yaml
# Sample Kafka payload for sanctions.match.alerted
event_type: "SANCTIONS_MATCH_DETECTED"
event_id: "evt-77a8b9c0-1234"
timestamp: "2026-09-18T10:15:30Z"
investor_id: "usr_99882211"
ethereum_address: "0x3F5CE5FBFe3E9af3971dD833D26bA9b5C936f0bE"
risk_level: "CRITICAL"
watchlist: "OFAC_SDN"
match_score: 0.985
matched_entity_name: "TARGET_INDIVIDUAL_NAME"
actions_taken:
 - "FIAT_ACCOUNT_LOCKED"
 - "ON_CHAIN_FREEZE_DISPATCHED"
 - "COMPLIANCE_ESCALATION_QUEUED"
```

## Security & Compliance Notes
- **PMLA 2002 & SEBI AML Framework:** Strictly adheres to statutory obligations to report designated individuals/entities to the Financial Intelligence Unit (FIU-IND) and prevent execution of orders by banned entities.
- **OFAC 50 Percent Rule:** Screens associated corporate structures and beneficial owners ($\ge 25\%$ domestic, $\ge 10\%$ high-risk / GIFT City) for indirect sanctions exposure.
- **Data Privacy (DPDP Act):** All PII transmitted to external verification feeds is tokenized or transmitted via zero-log dedicated private tunnels.

## Acceptance Criteria
- [ ] `services/sanctions-screener` microservice operational with automated daily ingestion of UN, OFAC, EU, UK, and MHA lists.
- [ ] P99 response latency $< 30\text{ms}$ for synchronous individual screening requests.
- [ ] Fuzzy matching engine achieves 100% recall on standard synthetic sanctions test suites including transliterations and typos.
- [ ] Nightly batch re-screening processes 100,000 investor profiles in $< 30$ minutes.
- [ ] High-confidence matches ($\ge 95\%$) automatically emit `compliance.account.freeze_command` and verify on-chain ledger freeze in Besu testnet.
- [ ] Compliance analyst override workflow functions with dual-approval logging and full audit trail in PostgreSQL.
- [ ] Prometheus metrics and OpenTelemetry distributed tracing integrated and active.

## Suggested Order / Dependencies
- **Prerequisites:** 004 (Domestic KYC/AML Policy), 005 (Foreign KYC/AML Policy), 104 (Event Schemas), 202 (KYC Service), 305 (Transfer with Compliance Hooks).
- **Parallel Tasks:** 701 (Threat Model), 704 (AML Transaction Monitoring).
- **Downstream Dependents:** 204 (Order Service Pre-Trade Gate), 217 (Admin Back-Office), 604 (Admin Review UI).
