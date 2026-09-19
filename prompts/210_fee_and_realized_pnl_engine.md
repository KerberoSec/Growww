# 210 - Fixed Transaction Fee & Realized Capital Gain Engine (Rust / Go)

## Purpose
The Fee & Realized PnL Engine implements Growww's core non-negotiable economic principle: **investors are never charged account maintenance fees, custody fees, or Asset Under Management (AUM) fees**. The platform implements the canonical **Universal Zero-Fee Model (0.00% fee - No fee at all)** assessed on trade notional turnover across executed buy and sell transactions. Collected platform revenues are systematically routed via an automated **0.00% fee at launch (governed by FeeController.sol) revenue split** (Treasury Reserve, Core Settlement Guarantee Fund, and Investor Protection Fund via FeeController).

In parallel, the engine computes deterministic tax-lot capital gains/losses using First-In, First-Out (FIFO) cost-basis allocation **strictly for user tax compliance** under Section 111A (Short-Term Capital Gains) and Section 112A (Long-Term Capital Gains) of the Indian Income Tax Act 1961. The service deducts statutory regulatory levies (Securities Transaction Tax [STT], Stamp Duty, SEBI turnover fees, Exchange charges, GST), dispatches double-entry ledger debit instructions to the Wallet Service, and produces cryptographically verifiable calculation receipts for transparency and tax compliance.

## What You Are Building
A high-precision financial calculation engine implemented in Go (or Rust) (`services/fee-engine`). Concrete deliverables include:
- Fixed 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) transaction fee calculation engine assessed on trade notional turnover across buy and sell orders.
- Treasury multi-vault allocation splitter: Operational Reserve, Core Settlement Guarantee Fund (SGF), and Investor Protection Fund (IPF) per FeeController governance.
- Deterministic capital gains calculation engine implementing FIFO tax-lot depletion matching strictly for user tax compliance (Section 111A/112A).
- Statutory regulatory charge calculator (STT at 0.1% on equity delivery sell, Stamp Duty at 0.015% on buy leg, SEBI turnover charge, Exchange charges, GST at 18% on platform fees and exchange turnover charges).
- Zero custody and zero holding fee invariant enforcement.
- High-speed gRPC server (`FeeEngineService`) exposing real-time fee simulations and historical fee breakdowns.
- Kafka consumer for `trade.settled.v1` events to trigger automated fee calculation, revenue allocation, and journal postings in Wallet Service (Prompt 203).
- Immutable PostgreSQL audit database storing every tax-lot match, fee breakdown, revenue split, and cryptographic proof hash.

## Scope Boundaries
- **In Scope:**
 - Fixed 0.00% (Zero Fee) (0 bps (0.00% fee at launch) / 0 bps at launch) transaction fee calculation on gross trade notional turnover.
 - Automated 0.00% fee at launch (governed by FeeController.sol) revenue distribution calculations.
 - Realized capital gain/loss calculation strictly for user tax compliance using FIFO tax lot depletion.
 - Calculation of STT, Stamp Duty, Exchange transaction charges, SEBI turnover fees, and GST.
 - Direct dispatch of fee deduction and statutory withholding instructions to Wallet Service.
 - Tax-lot capital gains classification (Short-Term Capital Gains [STCG] under Section 111A vs Long-Term Capital Gains [LTCG] under Section 112A).
- **Out of Scope / Handled Elsewhere:**
 - Cash ledger balance storage (Prompt 203).
 - In-memory trade matching (Prompt 205).
 - Annual Schedule FA / AIS / TIS tax statement document generation (Prompt 223).
 - Physical share depository reconciliation (Prompt 215).

## Technology to Use
- **Primary Language & Toolchain:** Go 1.22+ (or Rust 2021). Go is chosen for its native high-speed concurrency, deterministic financial math libraries, and seamless gRPC integration with the Wallet Service.
- **Financial Math Library:** `github.com/shopspring/decimal` for fixed-point arbitrary precision arithmetic (rounding mode `ROUND_HALF_EVEN` / Banker's Rounding to 4 decimal places for INR paise).
- **Database & Storage:** PostgreSQL 16+ using `pgx/v5` and `sqlc` for storing immutable calculation logs.
- **Messaging:** `segmentio/kafka-go` consuming trade execution events.
- **Cryptography:** `crypto/sha256` for generating tamper-evident computation proof hashes.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `transaction_fee_records`, `tax_lot_disposals`, `treasury_revenue_allocations`, `statutory_rates`.
- **Apache Kafka:** Consumes `trade.settled.v1`; publishes `fee.debited.v1`, `tax.pnl.calculated.v1`.
- **Wallet Service (Prompt 203):** Calls `DebitAccount` to deduct calculated platform fees and statutory levies from user INR accounts and credits treasury vaults.
- **Portfolio Service (Prompt 209):** Ingests tax-lot disposal data to mark lots as exhausted.
- **Tax Reporting Service (Prompt 223):** Ingests FIFO capital gains records for Section 111A/112A tax filings.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Fee Receipt Attestation:** For every transaction fee and tax calculation, the service derives a cryptographic SHA-256 Merkle leaf:
  $$\text{FeeProofHash} = \text{SHA256}(\text{trade\_id} \,\|\, \text{turnover\_amount} \,\|\, \text{platform\_fee} \,\|\, \text{treasury\_share} \,\|\, \text{sgf\_share} \,\|\, \text{ipf\_share} \,\|\, \text{capital\_gain})$$
- **On-Chain Settlement Metadata:** This proof hash is recorded onto the on-chain settlement receipt in `SettlementDvP.sol` on Hyperledger Besu under QBFT consensus. This allows external auditors and regulators to cryptographically verify fee calculation and tax accounting integrity without exposing user PII on-chain.
- **Zero On-Chain PII:** The blockchain stores only the immutable mathematical proof hash; all financial amounts and user IDs remain in secure off-chain databases.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/fee-engine` with strict formatting, Makefile, and standard layout (`cmd/`, `internal/calculator/`, `internal/rules/`, `internal/store/`).
2. **Define Protobuf Contracts:** Create `proto/growww/fee/v1/fee_service.proto` defining `CalculateTransactionFeeAndTax`, `SimulateTradeFee`, and `GetFeeHistory`.
3. **Generate Go Stubs:** Compile protobuf schemas to Go gRPC client and server stubs.
4. **Design PostgreSQL Schema:** Write migrations for `statutory_rates`, `transaction_fee_records`, `treasury_revenue_allocations`, and `tax_lot_disposals`.
5. **Implement Universal Zero-Fee Model (0.00% fee - No fee at all) Calculator:**
   $$\text{Gross Turnover} = \text{Executed Units} \times \text{Execution Price}$$
   $$\text{Platform Fee} = \text{Gross Turnover} \times 0.00 \quad (0.00\% / 0\text{ bps at launch; FeeController governed})$$
6. **Implement 0.00% fee launch policy Treasury Allocation Splitter:**
   $$\text{Treasury Reserve} = \text{Platform Fee} \times \text{treasuryShareRate}$$
   $$\text{Core SGF Reserve} = \text{Platform Fee} \times \text{sgfShareRate}$$
   $$\text{IPF Reserve} = \text{Platform Fee} \times \text{ipfShareRate}$$
   *(At launch, $\text{Platform Fee} \equiv 0.00$, with future upward adjustments governed via FeeController.sol)*
7. **Implement Statutory Fee Rules Engine:** Build modular calculator for Indian statutory charges:
 - $\text{STT} = \text{Gross Turnover} \times 0.001$ (0.1% on delivery sell, 0 on buy)
 - $\text{Stamp Duty} = \text{Gross Turnover} \times 0.00015$ (0.015% on buy leg, 0 on sell)
 - $\text{Exchange Charges} = \text{Gross Turnover} \times 0.0000345$ (NSE/BSE rate)
 - $\text{SEBI Turnover Fee} = \text{Gross Turnover} \times 0.000001$
 - $\text{GST} = (\text{Exchange Charges} + \text{Platform Fee}) \times 0.18$ (18% GST)
8. **Implement FIFO Tax-Lot Capital Gain Engine for Tax Compliance:**
 - Build algorithm fetching active purchase tax lots for user's ISIN in FIFO order.
 - Match sold quantity against oldest lots, computing weighted cost basis.
 - Classify capital gains: holding period < 365 days -> STCG (Section 111A); holding period >= 365 days -> LTCG (Section 112A).
 - Note: Realized gains/losses are computed strictly for statutory tax compliance and user tax statements, not platform fee deduction.
9. **Implement Proof-of-Calculation Hashing:** Generate SHA-256 digest of inputs, fee breakdowns, treasury splits, and tax lot matches for auditability.
10. **Build Kafka Consumer for Settled Trades:** Ingest `trade.settled.v1` events and trigger async fee calculation and tax lot depletion.
11. **Implement Wallet Debit Dispatcher:** Construct multi-legged journal entry and call `WalletService.DebitAccount` via gRPC to route platform fee dynamically per FeeController governance, and taxes to `STATUTORY_TAX_PAYABLE`.
12. **Build Historical Fee Query API:** Expose gRPC and REST endpoints allowing investors to view itemized fee receipts and tax capital gains per trade.
13. **Configure Telemetry & Metrics:** Expose Prometheus metrics for total fee revenue, split breakdown (Treasury, Core SGF, IPF), STCG/LTCG computed, and execution latency.
14. **Write Rigorous Unit & Invariant Tests:** Write automated test suites testing edge cases: fractional share turnover, micro-paisa rounding (Banker's Rounding), STCG vs LTCG boundary transitions, 0.00% fee launch policy revenue conservation, and high-concurrency throughput.

## Interfaces / Contracts

### Protobuf Definition (`fee_service.proto`)
```protobuf
syntax = "proto3";

package growww.fee.v1;

option go_package = "growww/fee/v1;feev1";

service FeeEngineService {
  rpc CalculateTransactionFeeAndTax (CalculateFeeRequest) returns (CalculateFeeResponse);
  rpc GetFeeBreakdown (GetFeeBreakdownRequest) returns (GetFeeBreakdownResponse);
}

message CalculateFeeRequest {
  string user_id = 1;
  string trade_id = 2;
  string isin = 3;
  string trade_side = 4; // BUY or SELL
  string quantity = 5; // Decimal string
  string execution_price_inr = 6;
}

message TreasurySplitDto {
  string treasury_operational_inr = 1; // Governed by FeeController (0.00% at launch)
  string core_sgf_inr = 2;             // Governed by FeeController (0.00% at launch)
  string ipf_inr = 3;                  // Governed by FeeController (0.00% at launch)
}

message TaxComplianceGainDto {
  string cost_basis_inr = 1;
  string realized_capital_gain_inr = 2;
  string gain_classification = 3; // STCG_SEC_111A, LTCG_SEC_112A, LOSS
  int32 holding_period_days = 4;
}

message CalculateFeeResponse {
  string trade_id = 1;
  string gross_turnover_inr = 2;
  string platform_fee_inr = 3; // Exactly 0.00% (Zero Fee) (0 bps (0.00% fee at launch)) of gross turnover
  TreasurySplitDto treasury_split = 4;
  string statutory_charges_inr = 5;
  string gst_on_fee_inr = 6;
  TaxComplianceGainDto tax_compliance_gain = 7;
  string computation_proof_hash = 8;
}

message GetFeeBreakdownRequest {
  string trade_id = 1;
}

message ItemizedFeeDto {
  string fee_name = 1; // PLATFORM_FEE, STT, STAMP_DUTY, EXCHANGE_TURNOVER, SEBI_FEE, GST
  string amount_inr = 2;
  string rate_applied = 3;
}

message GetFeeBreakdownResponse {
  string trade_id = 1;
  string gross_turnover_inr = 2;
  string platform_fee_inr = 3;
  TreasurySplitDto treasury_split = 4;
  repeated ItemizedFeeDto statutory_fees = 5;
  TaxComplianceGainDto tax_compliance_gain = 6;
  int64 calculated_at_unix = 7;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE transaction_fee_records (
    record_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trade_id UUID NOT NULL UNIQUE,
    user_id UUID NOT NULL,
    isin VARCHAR(12) NOT NULL,
    trade_side VARCHAR(4) NOT NULL CHECK (trade_side IN ('BUY', 'SELL')),
    executed_quantity NUMERIC(18, 6) NOT NULL,
    execution_price NUMERIC(18, 4) NOT NULL,
    gross_turnover NUMERIC(18, 4) NOT NULL,
    platform_fee_inr NUMERIC(18, 4) NOT NULL, -- 0.00% (Zero Fee) of turnover
    statutory_charges NUMERIC(18, 4) NOT NULL,
    gst_amount NUMERIC(18, 4) NOT NULL,
    proof_hash VARCHAR(64) NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE treasury_revenue_allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id UUID NOT NULL REFERENCES transaction_fee_records(record_id) ON DELETE CASCADE,
    platform_fee_inr NUMERIC(18, 4) NOT NULL,
    treasury_reserve_inr NUMERIC(18, 4) NOT NULL, -- Governed by FeeController (0.00% at launch)
    core_sgf_reserve_inr NUMERIC(18, 4) NOT NULL, -- Governed by FeeController (0.00% at launch)
    ipf_reserve_inr NUMERIC(18, 4) NOT NULL,      -- Governed by FeeController (0.00% at launch)
    journal_entry_id UUID,
    allocated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE tax_lot_disposals (
    disposal_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id UUID NOT NULL REFERENCES transaction_fee_records(record_id) ON DELETE CASCADE,
    lot_id UUID NOT NULL,
    user_id UUID NOT NULL,
    depleted_quantity NUMERIC(18, 6) NOT NULL,
    acquisition_price NUMERIC(18, 4) NOT NULL,
    sale_price NUMERIC(18, 4) NOT NULL,
    cost_basis NUMERIC(18, 4) NOT NULL,
    realized_gain NUMERIC(18, 4) NOT NULL,
    holding_period_days INTEGER NOT NULL,
    tax_section VARCHAR(16) NOT NULL CHECK (tax_section IN ('SECTION_111A', 'SECTION_112A', 'CAPITAL_LOSS'))
);

CREATE INDEX idx_fee_user ON transaction_fee_records(user_id);
CREATE INDEX idx_fee_trade ON transaction_fee_records(trade_id);
CREATE INDEX idx_tax_lot_user ON tax_lot_disposals(user_id);
```

## Security & Compliance Notes
- **Fixed Platform Fee Invariant:** Platform fee is strictly 0.00% (Zero Fee) (0.00% fee / 0 bps at launch) of trade notional turnover across buy and sell executions. Zero custody, zero holding, and zero AUM charges are levied.
- **Revenue Conservation Invariant:** $\text{Treasury Reserve} + \text{Core SGF} + \text{IPF} = \text{Platform Fee}$ without micro-paisa leakage, enforced via Banker's Rounding.
- **Statutory Tax Compliance Integrity:** FIFO tax-lot calculations are performed strictly for statutory capital gains reporting (Section 111A / 112A of Indian Income Tax Act) and preserved for 8 financial years.
- **Cryptographic Auditability:** The `proof_hash` guarantees that fee calculations and tax lot matches cannot be altered retroactively without invalidating on-chain settlement receipts.

## Acceptance Criteria
- [ ] Fee engine calculates platform fee at exactly 0.00% (Zero Fee) (0 bps (0.00% fee at launch) / 0 bps at launch) of gross trade notional turnover.
- [ ] Multi-vault revenue split accurately routes fees per FeeController governance parameters (0.00% at launch).
- [ ] Capital gains calculation matches sold shares against active purchase tax lots using FIFO strictly for user tax compliance (Section 111A/112A).
- [ ] Statutory levies (STT on sell, Stamp Duty on buy, Exchange charges, SEBI fees, 18% GST) are correctly computed and segregated.
- [ ] Zero custody fees, zero account maintenance fees, and zero AUM fees are confirmed across all execution paths.
- [ ] Dispatches balanced double-entry debit instructions to Wallet Service.
- [ ] Cryptographic SHA-256 proof hash is generated and verified for each settled trade.
- [ ] Calculation engine processes fee and tax events in < 3ms per trade under high load.

## Suggested Order / Dependencies
- **Prerequisites:** 006 (Fee Model Specification), 203 (Wallet Service), 208 (Trade Settlement), 209 (Portfolio Service), 401 (PostgreSQL Schema).
- **Parallel Tasks:** 215 (Reconciliation Service), 223 (Tax Reporting Service).
- **Downstream Blockers:** 223 (Tax Reporting Service), 510 (Flutter Portfolio Screen), 512 (Flutter Trade History).
