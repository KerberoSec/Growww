# Financial Domain Specifications & Trading Venue Architecture

**Document Version:** 1.0.0-PROD-SPEC  
**Status:** Approved  
**Owner:** Core Financial Engineering & Risk Architecture Group  
**Review Cadence:** Quarterly  
**Last Review:** September 2026  

---

## 1. Cross-Cutting Financial Domain Standards

### 1.1 Double-Entry Ledger Architecture (D-10)

The core accounting system is strictly an immutable, append-only double-entry journal. Account balances are derived projections and are never updated in place without an underlying journal entry.

#### Journal Entry Schema
```sql
CREATE TABLE journal_entry (
    id             BIGSERIAL       PRIMARY KEY,
    transaction_id UUID            NOT NULL,
    account_id     UUID            NOT NULL,
    asset_id       TEXT            NOT NULL,
    direction      CHAR(1)         NOT NULL CHECK (direction IN ('D', 'C')),
    amount         NUMERIC(38,18)  NOT NULL CHECK (amount > 0),
    entry_type     TEXT            NOT NULL,
    reference      TEXT            NOT NULL,
    created_at     TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_journal_tx ON journal_entry (transaction_id);
CREATE INDEX idx_journal_account_asset ON journal_entry (account_id, asset_id, created_at);
```

#### Balanced Transaction Constraint
```sql
CREATE OR REPLACE FUNCTION assert_transaction_balances()
RETURNS TRIGGER AS $$
DECLARE
    v_imbalance NUMERIC;
BEGIN
    SELECT COALESCE(SUM(CASE WHEN direction = 'D' THEN amount ELSE -amount END), 0)
    INTO v_imbalance
    FROM journal_entry
    WHERE transaction_id = NEW.transaction_id AND asset_id = NEW.asset_id;

    IF v_imbalance <> 0 THEN
        RAISE EXCEPTION 'Transaction % for asset % is unbalanced (delta: %)',
            NEW.transaction_id, NEW.asset_id, v_imbalance;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER enforce_balanced_transaction
AFTER INSERT ON journal_entry
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION assert_transaction_balances();
```

#### Daily Trial Balance Assertion
An automated job executes every trading day at 23:59:59 IST across all assets:
$$\sum_{i} \text{Debits}(a, i) - \sum_{i} \text{Credits}(a, i) = 0, \quad \forall a \in \mathbb{A}$$

---

### 1.2 End-to-End Idempotency & De-duplication (D-03)

Every state-mutating API call requires a unique client-generated `Idempotency-Key` header.

#### Database Idempotency Table
```sql
CREATE TABLE idempotency_keys (
    key             TEXT        PRIMARY KEY,
    account_id      UUID        NOT NULL,
    request_hash    BYTEA       NOT NULL,
    response_status INT,
    response_body   JSONB,
    state           TEXT        NOT NULL CHECK (state IN ('IN_PROGRESS', 'COMPLETED')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_idempotency_expiry ON idempotency_keys (expires_at);
```

#### Consumer-Side De-duplication
Kafka consumers write processed message offsets to a local tracking table in the same transaction as the business write:
```sql
CREATE TABLE processed_events (
    consumer_group  TEXT        NOT NULL,
    topic           TEXT        NOT NULL,
    partition_id    INT         NOT NULL,
    offset_num      BIGINT      NOT NULL,
    event_id        UUID        NOT NULL,
    processed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (consumer_group, topic, partition_id, offset_num)
);
```

---

### 1.3 Transactional Outbox & CDC Architecture (D-04)

Dual-write inconsistencies between PostgreSQL and Kafka are eliminated by persisting all domain events into an outbox table within the active database transaction.

```sql
CREATE TABLE outbox_events (
    id              BIGSERIAL   PRIMARY KEY,
    aggregate_type  TEXT        NOT NULL,
    aggregate_id    TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    payload         JSONB       NOT NULL,
    headers         JSONB       NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at    TIMESTAMPTZ
);

CREATE INDEX idx_outbox_unpublished ON outbox_events (id) WHERE published_at IS NULL;
```

Events are streamed to Kafka via Debezium Change Data Capture (CDC) directly reading the PostgreSQL Write-Ahead Log (WAL), ensuring zero lost events and deterministic ordering partitioned by `aggregate_id`.

---

## 2. Trading Venue & Market Microstructure Rules

### 2.1 Self-Trade Prevention (STP) Framework (D-01)

Self-Trade Prevention is enforced at the matching engine layer before fill generation. Matching evaluates the underlying **Beneficial Owner ID (PAN / Tax ID)** rather than trading account IDs.

#### STP Modes
1. `CANCEL_RESTING`: The passive order resting on the book is cancelled; incoming aggressive order continues matching.
2. `CANCEL_INCOMING`: The incoming aggressive order is rejected; resting passive order remains on the book.
3. `CANCEL_BOTH`: Both resting and incoming orders are cancelled.
4. `DECREMENT_AND_CANCEL`: The smaller quantity is cancelled against the larger; remaining uncrossed quantity of the larger order remains active.

---

### 2.2 Instrument Master & Order Entry Validation (D-02)

Every order is validated against strict instrument parameters prior to order book routing:

```
reject if price % tick_size != 0
reject if quantity % lot_size != 0
reject if quantity < min_quantity or quantity > max_quantity
reject if abs(price - reference_price) / reference_price > dynamic_circuit_band_pct
reject if price * quantity > max_order_notional
reject if instrument.session_state != CONTINUOUS_TRADING
```

---

### 2.3 Clock Discipline & Monotonic Sequence Numbering (D-06)

1. **Monotonic Priority:** Internal order priority uses strict monotonically increasing 64-bit sequence numbers assigned at the single API Gateway entry barrier.
2. **Time Synchronisation:** Trading nodes synchronize clocks via IEEE 1588 Precision Time Protocol (PTP) with NTP fallback, with maximum drift tolerance capped at 50 microseconds.
3. **UTC Persistence:** All timestamps are captured in UTC (`TIMESTAMPTZ`) with microsecond resolution and formatted to IST (UTC+05:30) at the UI presentation boundary.

---

### 2.4 Market Data Feed Protocol (D-07)

Market data broadcasts (Level 2/Level 3) over WebSocket and UDP multicast maintain strictly gapless sequence numbering:

```
Message Envelope: {
  seq: uint64,
  instrument_id: string,
  event_type: "SNAPSHOT" | "DELTA" | "HEARTBEAT",
  timestamp_ns: uint64,
  bids: [[price, qty]],
  asks: [[price, qty]]
}
```

- **Gap Detection:** If `received_seq != expected_seq`, the client immediately invalidates local state and requests an atomic snapshot.
- **Heartbeats:** Broadcast every 1,000 ms. If no message or heartbeat is received for 3,000 ms, the client displays a connection disruption banner and freezes order entry.

---

### 2.5 Order-to-Trade Ratio (OTR) & Surveillance Throttling (D-08)

To prevent quote-stuffing and comply with SEBI algorithmic trading guidelines, OTR is monitored dynamically per participant per instrument session:
$$\text{OTR} = \frac{\text{Orders Submitted} + \text{Orders Modified} + \text{Orders Cancelled}}{\max(1, \text{Executed Trades})}$$

- **Warning Tier:** OTR > 50:1 -> Warning alert.
- **Throttling Tier:** OTR > 100:1 -> Rate limited to 5 orders/sec.
- **Suspension Tier:** OTR > 250:1 -> Order entry suspended for 15 minutes.

---

## 3. Risk Management & Settlement Specifications

### 3.1 SPAN Multi-Collateral Risk Model (D-09)

$$\text{InitialMargin} = \max(\text{SPAN\_Scenario\_Loss}, \text{Notional} \times \text{FloorRate})$$
$$\text{MaintenanceMargin} = \text{InitialMargin} \times 0.75$$
$$\text{CollateralValue} = \sum_{i} \min(Q_i \times P_i \times (1 - H_i), \text{Cap}_i)$$

- **Wrong-Way Risk Prohibition:** An asset cannot be posted as collateral to back positions in the same or highly correlated underlying instruments.
- **Weekend Liquidation Halt:** Automated liquidations on tokenized securities are strictly halted when underlying physical exchanges (NSE/BSE) are closed, preventing predatory liquidations during illiquid off-hours.

---

### 3.2 DvP Settlement Saga State Machine (D-05)

```
[INITIATED]
    |
    v
[FUNDS_RESERVED]  ---(failure)---> [FUNDS_RELEASED] -> [FAILED]
    |
    v
[DEPOSITORY_INSTRUCTED]
    |
    v
[DEPOSITORY_CONFIRMED]
    |
    v
[CHAIN_SUBMITTED] ---(timeout)---> [CHAIN_RECEIPT_QUERY]
    |
    v
[CHAIN_CONFIRMED]
    |
    v
[SETTLED]
```

- **Irreversible Leg Ordering:** Off-chain reversible reservations execute first; irreversible blockchain ledger transfers execute last.
- **Receipt Verification:** Timeouts on chain submissions trigger transaction receipt polling by cryptographic hash, eliminating double-execution.

---

### 3.3 Trading Calendar & Session State Machine (D-11)

```
+-----------------------------------------------------------------------------------------------+
|                               TRADING SESSION STATE MACHINE                                   |
|                                                                                               |
|  [CLOSED] -> [PRE_OPEN] -> [CALL_AUCTION] -> [CONTINUOUS] -> [CIRCUIT_HALT] -> [MAINTENANCE] |
+-----------------------------------------------------------------------------------------------+
```

Sessions are managed per instrument by `trading-calendar-service`, integrating Indian statutory market holidays, clearing settlement windows, and corporate action halts.

---

### 3.4 Regulatory Filings, Tax & Post-Trade Invariants (D-12 to D-19)

1. **FIFO P&L Discipline (D-13):** Capital gains (STCG/LTCG) are calculated strictly using First-In-First-Out (FIFO) tax-lot allocation under Indian Income Tax Act regulations.
2. **Contract Notes (D-15):** Digitally signed Electronic Contract Notes (ECN) containing timestamped trade details, brokerage, STT, GST, and stamp duty are generated within 24 hours of trade execution.
3. **Suspicious Transaction Reporting (D-19):** Automated FIU-IND alerts trigger upon structured cash deposits, rapid volume spikes exceeding 500% of 30-day baseline, and sanctions list matches.
