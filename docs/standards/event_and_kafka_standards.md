# Growww Platform Event Schema & Kafka Topic Standards

## 1. Scope & Governance
This standard establishes the asynchronous event streaming conventions, CloudEvents v1.0 schema envelopes, Kafka topic taxonomy, and deterministic partitioning policies across the Growww / NBSE ecosystem.

---

## 2. Kafka Topic Taxonomy
All Kafka topics must adhere strictly to the formal 6-segment dot-delimited grammar:

```
<env>.<entity>.<bounded-context>.<aggregate>.<event-name>.<version>
```

### 2.1 Segment Definitions
- **`<env>`**: Deployment environment (`prod`, `staging`, `testnet`, `dev`).
- **`<entity>`**: Jurisdiction regulatory entity (`domestic` for SEBI/RBI, `giftcity` for IFSCA offshore).
- **`<bounded-context>`**: Core service domain (`trading`, `settlement`, `compliance`, `custody`, `ledger`).
- **`<aggregate>`**: Domain entity subject (`order`, `trade`, `dvp`, `balance`, `investor`).
- **`<event-name>`**: Past-tense action verb (`placed`, `matched`, `cancelled`, `executed`, `whitelisted`).
- **`<version>`**: Schema evolution major version (`v1`, `v2`).

### 2.2 Canonical Topic Examples
- `prod.domestic.trading.order.matched.v1`
- `prod.giftcity.settlement.dvp.executed.v1`
- `prod.domestic.ledger.besu.block_committed.v1`
- `prod.domestic.compliance.investor.kyc_verified.v1`

---

## 3. CNCF CloudEvents v1.0 Envelope Standard
All business event payloads are wrapped in standard CloudEvents envelopes:

```json
{
  "specversion": "1.0",
  "id": "018d9f45-728b-7000-848f-39589dfd0001",
  "source": "https://trading.growww.trade/matching-engine",
  "type": "growww.trading.order.matched.v1",
  "subject": "ord_98741",
  "time": "2026-09-20T12:00:00.123456Z",
  "datacontenttype": "application/json",
  "correlationid": "corr_887412",
  "actorid": "usr_99014",
  "entityid": "DOMESTIC",
  "data": {
    "execution_id": "exec_0182",
    "symbol": "BTC/USDT",
    "fill_price_e8": 6000000000000,
    "fill_qty_e8": 150000000
  }
}
```

---

## 4. Partition Key Strategy Matrix
Deterministic partition keys guarantee in-order sequential execution per business boundary:

| Event Domain | Partition Key | Guarantee |
| :--- | :--- | :--- |
| User Orders & Cancellations | `user_id` | Guarantees FIFO processing per investor account. |
| Order Book & Trade Matching | `symbol` / `security_isin` | Guarantees deterministic order sequence per asset book. |
| DvP Blockchain Settlement | `settlement_id` | Guarantees serialized multilateral clearing execution. |
| Account Ledger & Balances | `account_id` | Guarantees zero balance mutation race conditions. |

---

## 5. Dead-Letter Queue (DLQ) & Poison Pill Topology
- If a message fails processing after 3 exponential backoff retries (100ms, 500ms, 2000ms):
  - Forwarded immediately to `<original-topic>.dlq`.
  - Enriched with headers: `x-error-reason`, `x-original-partition`, `x-failure-timestamp`.
  - SRE alerts dispatched via Prometheus/PagerDuty.
