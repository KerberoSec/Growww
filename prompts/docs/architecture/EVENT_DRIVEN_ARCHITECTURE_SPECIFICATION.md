# Event-Driven Architecture Specification

**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Distributed Systems & Event Streaming Architecture Group  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  

---

## 1. Kafka Topic Taxonomy & Partitioning Strategy (E-01)

Topics follow a deterministic, hierarchical naming schema:
```
growww.<domain>.<entity>.<event_type>.v<version>
```

### 1.1 Topic Catalog & Governance

| Topic Name | Partitions | Partition Key | Replication Factor | Retention Policy | Compaction |
|---|---|---|---|---|---|
| `growww.order.order.created.v1` | 32 | `account_id` | 3 (min.isr=2) | 7 Days | Delete |
| `growww.trade.trade.executed.v1` | 32 | `instrument_id` | 3 (min.isr=2) | Forever (`retention.ms=-1`) | False |
| `growww.wallet.journal.posted.v1` | 16 | `account_id` | 3 (min.isr=2) | Forever (`retention.ms=-1`) | False |
| `growww.settlement.instruction.v1` | 8 | `settlement_id` | 3 (min.isr=2) | 90 Days | Delete |
| `growww.audit.entry.appended.v1` | 8 | `stream_id` | 3 (min.isr=2) | Forever (`retention.ms=-1`) | False |
| `growww.instrument.state.v1` | 4 | `instrument_id` | 3 (min.isr=2) | Forever | Compact |

- **Partition Key Assignment:** Ensures strict in-order message delivery for that specific entity.
- **Financial Immutability:** Financial journals and trade audit streams are retained permanently.

---

## 2. Standardized Event Envelope (E-06)

Every event published to Kafka is wrapped in a standardized CloudEvents-compliant protobuf envelope:

```protobuf
syntax = "proto3";

package growww.events.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/any.proto";

message EventEnvelope {
  string event_id = 1;                  // UUIDv4 unique identifier
  string correlation_id = 2;            // End-to-end request trace ID
  string causation_id = 3;              // ID of the event/command triggering this event
  string source_service = 4;           // Originating microservice name
  string schema_version = 5;            // Semantic version of event payload
  google.protobuf.Timestamp occurred_at = 6;
  google.protobuf.Any payload = 7;     // Typed protobuf domain event
  map<string, string> metadata = 8;     // Tenant, jurisdiction, tracing headers
}
```

---

## 3. Schema Evolution & Registry Governance (E-02)

1. **Compatibility Mode:** Enforced as `BACKWARD_TRANSITIVE` across all topics. New consumer versions can always read historical events.
2. **Schema Registry Validation in CI:** Every pull request runs `buf breaking --against` to detect incompatible schema alterations:
   - Field numbers must never be re-used or changed.
   - Field types must never be modified.
   - Required fields cannot be added; optional fields must provide safe default values.

---

## 4. Poison Message Handling & Dead-Letter Queues (E-03)

Failed message processing follows an automated multi-stage escalation path:

```
[Incoming Message]
       |
       v
[Processing Attempt 1..3]  ---(transient error)---> [In-Process Backoff (100ms, 500ms, 2s)]
       |
       +--(unresolved after 3 attempts)
       |
       v
[growww.<domain>.<topic>.retry.v1]  ---(delayed processing up to 3 attempts)---> [Success]
       |
       +--(unresolved after 6 total attempts)
       |
       v
[growww.<domain>.<topic>.dlq.v1]  ---> [Alert Ops / Manual Inspection]
       |
       v
[Commit Offset to Source Topic]
```

- **DLQ Monitoring:** Any non-zero queue depth on `.dlq.v1` topics triggers an immediate P2 SRE alert.

---

## 5. Security, Replay & Operational Observability (E-04 to E-13)

1. **Authentication & Authorization (E-12):** Production clusters enforce mTLS for broker-to-broker and client-to-broker communication with granular SASL/SCRAM ACLs.
2. **Consumer Lag Monitoring (E-07):** PromQL alerts fire if consumer lag exceeds 5,000 messages or 10 seconds of processing time on critical trading topics.
3. **Event Replay Runbook (E-08):** Read-only projection consumers support rewind replay by provisioning a new unique consumer group ID (`<service>-replay-<timestamp>`) with `auto.offset.reset=earliest`.
