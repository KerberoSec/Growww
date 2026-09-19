# ADR-0022: Dual Historical OHLCV Time-Series Storage for Corporate Actions

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Chief Financial Data Architect, Compliance Lead  

---

## 1. Context
Corporate actions (stock splits, bonus issues) create sudden 50% price drops on charts if unadjusted, while adjusting raw trades breaks statutory tax and regulatory audit logs.

---

## 2. Decision
Maintain **Dual Historical OHLCV Tables in TimescaleDB**:
1. `raw_ohlcv_1m`: Immutable raw execution records for regulatory compliance and tax audits.
2. `adjusted_ohlcv_1m`: Materialized view dynamically adjusted by cumulative split and dividend factors for frontend charting applications.

---

## 3. Consequences
- **Positive:** Chart indicators display continuous smooth curves; regulatory audits inspect exact unadjusted historical executions.
- **Trade-offs:** Dual storage requires indexing overhead in TimescaleDB.
