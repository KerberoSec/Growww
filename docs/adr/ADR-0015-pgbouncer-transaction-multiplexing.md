# ADR-0015: PgBouncer Transaction Multiplexing & Connection Pooling

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Database Infrastructure Lead, VP Engineering  

---

## 1. Context
At 09:15:00 IST (market open), tens of thousands of concurrent client connections surge into the platform, risking PostgreSQL backend process starvation (`max_connections` exhaustion).

---

## 2. Decision
Deploy **PgBouncer** in **Transaction Pooling Mode (`pool_mode = transaction`)** fronting all PostgreSQL clusters.
- Multiplexes up to 10,000 application client connections over a dedicated pool of 50 persistent server connections per database.
- Enforces strict prohibition of session-level state (`SET`, prepared statements require PgBouncer parameter tuning).

---

## 3. Consequences
- **Positive:** Lowers database memory overhead by $80\%$; eliminates backend connection fork overhead during market-open surges.
- **Trade-offs:** Applications must avoid session-scoped locks outside active transactions.
