# ADR-0021: Candlestick Late-Event Sliding Watermark Revision Protocol

**Status:** Accepted  
**Date:** September 2026  
**Deciders:** Market Data Lead, Frontend Charting Architect  

---

## 1. Context
Asynchronous network jitter and distributed partition relaying cause trade events to arrive after a 1-minute candlestick bar has closed, resulting in open-close bar discontinuities on client charts.

---

## 2. Decision
Implement a **1500ms Sliding Watermark Revision Window**:
- Trades arriving within 1500ms of a closed bar recalculate historical high/low/close prices in TimescaleDB and emit a `CANDLE_REVISE` event (`is_revision = true`).
- Client charting engines update the historical bar in-place without shifting active bar open prices.

---

## 3. Consequences
- **Positive:** Guarantees Open-Close bar continuity; prevents algorithmic indicator corruption.
- **Trade-offs:** Requires charting SDKs to handle historical bar mutation events.
