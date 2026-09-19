# Runbook 16: Candlestick Stream Desynchronization & Historical Bar Repair

**Runbook ID:** RBK-OPS-016  
**Severity Tier:** P2 (High)  
**Authority:** Market Data Operations Lead / SRE Lead  

---

## 1. Description & Trigger Conditions
Triggered when candlestick bar stream verification detects:
- Discrepancy between published 1m/5m/1h OHLCV bars and the matching engine trade execution log.
- Late-arriving trades received beyond the 1500ms sliding watermark grace period.
- Negative or inverted candle bounds ($High < Low$, $Open < 0$).

---

## 2. Remediation Workflow

```
[1. Identify Corrupted ISIN & Time Window]
  - Query market-data-service Prometheus metric: market_data_candle_revision_watermark_breach_total
  - Identify ISIN, bar timeframe (e.g., 1m), and target timestamp interval [T_start, T_end]
                  |
                  v
[2. Replay Canonical Trade Journal]
  - Extract canonical trade log from matching engine Raft / WAL for [T_start, T_end]
  - Re-aggregate raw OHLCV bars with 128-bit micro-turnover precision
                  |
                  v
[3. Publish CANDLE_REVISE Event]
  - Emit CANDLE_REVISE payload on Kafka topic: marketdata.candles.revised.v1
  - Overwrite historical bar record in TimescaleDB / ClickHouse
  - Trigger WebSocket update to active client charts without altering subsequent bar open-close bounds
                  |
                  v
[4. Audit & Verification]
  - Verify OHLC integrity invariants: High >= max(Open, Close, Low) and Low <= min(Open, Close, High)
  - Log resolution report in operational audit repository
```
