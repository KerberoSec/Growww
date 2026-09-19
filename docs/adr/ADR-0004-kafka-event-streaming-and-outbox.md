# ADR-0004: Apache Kafka Event Streaming with Transactional Outbox Pattern

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Architect, Data Platform Lead  

---

## 1. Context
Microservices need to communicate state changes asynchronously across service boundaries without introducing dual-write inconsistencies between PostgreSQL database commits and message queue dispatches.

---

## 2. Decision
Adopt **Apache Kafka** as the event streaming backbone combined with the **Transactional Outbox Pattern** powered by Debezium Change Data Capture (CDC).
- Business transactions write domain state and an outbox event in the same ACID PostgreSQL transaction.
- Debezium reads the PostgreSQL WAL and publishes events to Kafka with strict ordering per aggregate key.

---

## 3. Alternatives Considered
- **Direct Dual-Write (DB + Kafka producer):** Rejected because service crashes between DB commit and Kafka publish cause permanent state divergence and phantom settlements.
- **RabbitMQ / NATS:** Excellent message queuing, but lack Kafka's permanent replayable distributed log capability required for financial event sourcing and regulatory audit.

---

## 4. Consequences
- **Positive:** Guaranteed at-least-once event delivery with zero phantom writes; complete historical replayability.
- **Trade-offs:** Additional operational dependency on Debezium connectors and Kafka broker clusters.
