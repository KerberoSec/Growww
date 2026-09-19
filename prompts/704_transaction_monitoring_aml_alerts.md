# 704 - Real-Time Transaction Monitoring & AML Suspicious Activity Alerting

## Purpose
Establishes a continuous, real-time transaction monitoring and anti-money laundering (AML) surveillance engine. It analyzes high-velocity fiat cash flows (UPI, NEFT, RTGS on-ramps and off-ramps), order book interactions, and on-chain digital security token transfers to detect illicit activity - including smurfing/structuring (deliberate transactions under ₹50,000 threshold to evade reporting), rapid round-tripping, circular trading rings, wash trading, front-running, and sudden volume velocity anomalies. 

The system automates the generation and submission of Suspicious Transaction Reports (STRs) to India's Financial Intelligence Unit (FIU-IND) and IFSCA, ensuring complete compliance with the Prevention of Money Laundering Act (PMLA 2002) and SEBI Market Surveillance regulations.

## What You Are Building
- **AML Transaction Monitoring Engine:** `services/aml-monitoring`, a distributed event stream processing application evaluating transactions against deterministic typology rules and behavioral anomaly models.
- **Typology & Rule Execution Pipeline:** Low-latency rule engine computing stateful sliding window aggregations (1-hour, 24-hour, 7-day, 30-day) over investor fiat and token flows.
- **Circular Trading & Graph Analysis Module:** Graph-based cycle detection service identifying collusive trading patterns between interrelated investor accounts.
- **FIU-IND STR/SAR Report Generator:** Automated reporting module serializing suspicious case records into FIU-IND compliant XML format with digital signatures.
- **Automated Risk Throttling & Freeze Hook:** Triggering real-time fiat withdrawal holds and on-chain permissioned ledger account freezes upon detection of high-risk AML violations.

## Scope Boundaries
- **In Scope:**
 - Real-time Kafka event ingestion across deposits, withdrawals, order placements, trade executions, and token transfers.
 - Stateful window aggregations (e.g. cumulative deposit volume > ₹10,000,000 / month; $>5$ deposits in 1 hour just below ₹50,000).
 - Market manipulation surveillance: wash trading (self-matching via related entities), spoofing, pump-and-dump velocity spikes.
 - Investor dynamic risk scoring and alert generation.
 - Automated STR/SAR dossier compilation with full evidentiary audit trails.
- **Out of Scope / Handled Elsewhere:**
 - Order book execution and matching (handled in Prompt 205).
 - Static identity verification and document OCR (handled in Prompt 202).
 - Admin/Compliance review UI portal (handled in Prompts 604 and 605).

## Technology to Use
- **Stream Processing Engine:** Apache Flink 1.18+ / Faust (Python stream processor) consuming transactional Kafka topics with exactly-once processing guarantees.
- **Analytical Data Store:** ClickHouse 24+ for ultra-fast columnar aggregations and historical pattern backtesting over billions of trade records.
- **In-Memory State & Velocity Counters:** Redis 7 Cluster using Sliding Window Rate Limiters and HyperLogLog counters for real-time velocity tracking.
- **Relational Case Store:** PostgreSQL 16 for structured AML case management, analyst notes, and compliance escalations.
- **Justification:** Apache Flink and ClickHouse allow sub-second stateful stream processing over millions of daily events, enabling immediate detection of structuring and coordinated market manipulation that batch SQL queries cannot catch in time.

## Backend / Infra Touchpoints
- **Microservices:** Payment Gateway Service (212), Order Service (204), Matching Engine (205), Trade Settlement (208), Wallet Service (203), Reporting Service (216).
- **Messaging:** Kafka cluster topics: `orders.placed`, `trades.executed`, `wallet.fiat_deposited`, `wallet.fiat_withdrawn`, `ledger.tokens_transferred`, `aml.alerts.high_priority`.
- **Infrastructure:** ClickHouse OLAP Cluster, PostgreSQL 16, Redis 7 Cluster, HashiCorp Vault.

## Blockchain Interaction
Correlation of on-chain token movements with off-chain fiat rails and compliance freezing:
- **On-Chain/Off-Chain Cross-Reconciliation:** The monitoring engine subscribes to Besu ledger events emitted by `DigitalSecurityToken.sol` (`Transfer`, `TokensMinted`, `TokensRedeemed`) and `SettlementDvP.sol` (`SettlementCompleted`), correlating them with fiat bank transfer timestamps to detect off-chain shadow settlements.
- **Automated On-Chain Account Freezing:** When a critical AML pattern (e.g., circular wash-trading ring or terrorist financing typology) exceeds the severity threshold ($\text{Risk Score} \ge 90$), the service dispatches an emergency freeze command via Kafka to Prompt 305, invoking `ComplianceRegistry.freezeInvestor(address, reasonHash)` on Hyperledger Besu.
- **Tamper-Evident Evidence Anchoring:** The Merkle root of all aggregated transaction records involved in a flagged AML case is anchored to the permissioned blockchain for forensic immutability before submission to FIU-IND.

## Step-by-Step Build Instructions
1. Scaffold the `services/aml-monitoring` service structure and configure dependencies for Flink/Faust and ClickHouse clients.
2. Define Kafka event consumer topologies for `wallet.fiat_deposited`, `wallet.fiat_withdrawn`, `orders.placed`, and `trades.executed`.
3. Design and migrate the ClickHouse schema for `trades_surveillance_stream`, `fiat_flow_stream`, and `investor_velocity_hourly`.
4. Implement the Structuring / Smurfing Rule Engine: detect multiple deposits between ₹45,000 and ₹49,999 aggregating to $> ₹200,000$ within a 72-hour sliding window.
5. Implement the High-Velocity Flow Rule: detect rapid deposit followed by immediate full-balance withdrawal or token redemption within 30 minutes.
6. Implement the Wash Trading & Self-Trading Detector: identify buy and sell orders in `matching_engine` executed between accounts sharing PAN numbers, bank accounts, IP addresses, or device fingerprints.
7. Implement the Circular Trading Cycle Algorithm using graph traversal: detect cycles where Asset $A \to B \to C \to A$ occurs at artificial non-market prices to inflate volume.
8. Implement the Dynamic Risk Scoring Engine: aggregate rule triggers into a composite 0-100 risk score with exponential decay over time.
9. Build the automated Alert Dispatcher: publish high-risk alerts to `aml.alerts.high_priority` Kafka topic and write case records to PostgreSQL.
10. Implement the automated on-chain freeze trigger for alerts with score $> 90$, invoking `ComplianceRegistry.freezeInvestor` through the relayer.
11. Build the FIU-IND STR/SAR XML generator conforming to FIU-IND XML Specification v2.0, embedding investor details, transaction chronology, and typology codes.
12. Implement historical backtesting framework: replay 30 days of archived ClickHouse trade logs to tune false-positive threshold parameters.
13. Set up Prometheus metrics (`aml_events_processed_total`, `aml_alerts_generated_total`, `aml_rule_latency_ms`) and Grafana surveillance dashboards.

## Interfaces / Contracts
```protobuf
// schemas/aml/v1/aml_monitoring.proto
syntax = "proto3";
package growww.aml.v1;

enum AlertSeverity {
  SEVERITY_UNSPECIFIED = 0;
  SEVERITY_LOW = 1;
  SEVERITY_MEDIUM = 2;
  SEVERITY_HIGH = 3;
  SEVERITY_CRITICAL = 4;
}

enum TypologyType {
  TYPOLOGY_UNSPECIFIED = 0;
  STRUCTURING_SMURFING = 1;
  RAPID_MOVEMENT_OF_FUNDS = 2;
  WASH_TRADING_CIRCULAR = 3;
  DORMANT_ACCOUNT_SUDDEN_SPIKE = 4;
  SANCTIONS_ASSOCIATION = 5;
}

message AmlAlertEvent {
  string alert_id = 1;
  string timestamp_utc = 2;
  string investor_id = 3;
  string ethereum_address = 4;
  AlertSeverity severity = 5;
  TypologyType typology = 6;
  double risk_score = 7; // 0.0 to 100.0
  string description = 8;
  repeated string triggering_transaction_ids = 9;
  double total_inr_amount = 10;
  bool automated_freeze_executed = 11;
  string evidence_merkle_root = 12;
}
```

```sql
-- ClickHouse DDL for AML Stream Ingestion
CREATE TABLE default.aml_transaction_events (
    event_time DateTime64(3, 'Asia/Kolkata'),
    event_type LowCardinality(String),
    investor_id UUID,
    fiat_amount Decimal64(2),
    security_symbol LowCardinality(String),
    order_id UUID,
    trade_id UUID,
    counterparty_id UUID,
    ip_address IPv4,
    device_fingerprint String,
    on_chain_tx_hash FixedString(66)
) ENGINE = MergeTree()
ORDER BY (event_type, investor_id, event_time)
SETTINGS index_granularity = 8192;
```

## Security & Compliance Notes
- **PMLA 2002 Section 12:** Mandates reporting of suspicious transactions (cash transactions $\ge ₹10\text{ Lakhs}$ or structured transactions) to Director, FIU-IND within 7 working days.
- **SEBI Market Surveillance:** Complies with SEBI circulars on surveillance of abnormal trading behavior and pre-emptive freezing of suspicious accounts.
- **Data Privacy & Tamper Resistance:** Analytics pipelines process tokenized identifiers; raw PII is stored separately with AES-256 encryption.

## Acceptance Criteria
- [ ] Stream processing pipeline ingests and evaluates transactions with end-to-end latency $< 50\text{ms}$.
- [ ] Structuring detection rule accurately identifies 100% of synthetic test scenarios involving split payments under ₹50,000.
- [ ] Wash trading and circular graph analysis identifies collusive cycles in $< 200\text{ms}$.
- [ ] Critical alerts ($\text{Risk Score} \ge 90$) trigger automated on-chain account freeze and verified on Besu testnet.
- [ ] FIU-IND XML schema exporter produces valid, compliant STR reports verified against official validation tools.
- [ ] ClickHouse analytical queries complete in $< 100\text{ms}$ over a test dataset of 50 million historical trades.
- [ ] Alert deduplication and suppression mechanisms prevent alert fatigue for known operational anomalies.

## Suggested Order / Dependencies
- **Prerequisites:** 004 (Domestic KYC/AML), 104 (Kafka Standards), 203 (Wallet Service), 204 (Order Service), 208 (Settlement Service), 403 (Kafka), 404 (ClickHouse).
- **Parallel Tasks:** 703 (Sanctions Screening), 216 (Reporting Service).
- **Downstream Dependents:** 604 (Admin Review), 605 (Exception Queue), 706 (Incident Response).
