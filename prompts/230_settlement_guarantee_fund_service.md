# 230 - Settlement Guarantee Fund & Default Waterfall Service (Go / Temporal / PostgreSQL)

## Purpose
A Clearing Corporation acts as the central counterparty (CCP), interposing itself between buyers and sellers to guarantee trade settlement even if one or more clearing members default. Under CPMI-IOSCO Principle 4 (Credit Risk) and Principle 7 (Liquidity Risk), as well as SEBI (Stock Exchanges and Clearing Corporations) Regulations 2018, the clearing entity must maintain a ring-fenced **Core Settlement Guarantee Fund (Core SGF)** to absorb extreme counterparty defaults.

The **Settlement Guarantee Fund & Default Waterfall Service** orchestrates the end-to-end lifecycle of the SGF. It conducts daily regulatory stress tests (Cover-1 and Cover-2 extreme market shock scenarios), calculates clearing member contribution quotas, monitors collateral health, and executes the deterministic multi-tranche **Default Waterfall** when a participant fails to meet settlement obligations. By coordinating off-chain bank escrow accounts, tokenized collateral reserves, and on-chain SGF smart contracts (`SettlementGuaranteeFund.sol`, Prompt 315), this service ensures zero settlement disruption and preserves platform solvency.

## What You Are Building
A resilient Go orchestration microservice (`services/sgf-service`) integrated with Temporal workflow engine for fault-tolerant default handling, backed by PostgreSQL and Apache Kafka. Key deliverables include:
- **Daily SGF Stress Testing & Sizing Engine:** Evaluates Cover-1 (default of the single largest clearing participant) and Cover-2 (simultaneous default of the two largest participants) stress scenarios across 20+ historical and hypothetical stress market shocks.
- **Member Contribution Quota Calculator:** Computes monthly Minimum Required Corpus (MRC) allocations per participant based on gross turnover, open interest, and peak margin utilization.
- **Deterministic Default Waterfall Orchestrator:** Implements a state-machine-backed Temporal workflow executing the SEBI statutory default waterfall sequence:
  1. *Tranche 1:* Defaulter's Initial & Mark-to-Market Margins.
  2. *Tranche 2:* Defaulter's Core SGF Contribution.
  3. *Tranche 3:* Clearing Corporation (CC) Dedicated Capital Allocation.
  4. *Tranche 4:* Core SGF Pooled Contributions (Non-defaulting members pro-rata).
  5. *Tranche 5:* Clearing Corporation Capital Reserves / Insurance.
  6. *Tranche 6:* Assessment Calls / Haircut Recovery on non-defaulting participants.
- **Collateral Liquidation Manager:** Interfaces with Custodian Depository (Prompt 213) and Treasury Banking (Prompt 212/232) to liquidate defaulter assets within settlement cut-off windows.
- **On-Chain SGF Relayer Bridge:** Dispatches multi-party cryptographic transactions to `SettlementGuaranteeFund.sol` on Hyperledger Besu to lock, replenish, or slash on-chain reserve tranches.

## Scope Boundaries
- **In Scope:**
 - Calculation and rebalancing of Core SGF Minimum Required Corpus (MRC).
 - Daily Cover-1 and Cover-2 stress testing against market shock scenarios.
 - Tracking of member SGF deposit allocations (Cash, e₹ CBDC, Sovereign G-Secs).
 - Automated triggering and step-by-step execution of the multi-stage Default Waterfall workflow.
 - Assessment call generation and tracking for non-defaulting clearing participants.
 - Event logging and regulatory reporting for CCP risk disclosures.
- **Out of Scope / Handled Elsewhere:**
 - Pre-trade and intra-day individual portfolio VaR/ELM margin calculation (handled in Prompt 229).
 - Normal-course Delivery-versus-Payment trade settlement execution (handled in Prompt 208 / Prompt 306).
 - Physical auction and buy-in market resolution for share short deliveries (handled in Prompt 231).
 - Smart contract Solidity logic for on-chain fund locking and slashing (handled in Prompt 315).

## Technology to Use
- **Primary Language & Framework:** **Go 1.22+** with `gin-gonic/gin` for internal administrative APIs and `google.golang.org/grpc` for microservice RPCs.
- **Workflow Engine:** **Temporal.io Go SDK** (`go.temporal.io/sdk`) for durable, long-running, fault-tolerant execution of default handling and liquidation sagas.
- **Database & Storage:** **PostgreSQL 16+** with `jackc/pgx/v5` connection pool for member allocation ledgers, stress test telemetry, and audit histories.
- **In-Memory Cache:** **Redis 7.2** for real-time tracking of current fund balances, tranche utilization levels, and live member exposure.
- **Event Streaming:** **Apache Kafka** via `segmentio/kafka-go` for broadcasting SGF health telemetry, stress breach warnings, and default waterfall execution events.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `sgf_corpus_configs`, `sgf_member_allocations`, `sgf_stress_test_runs`, `default_waterfall_executions`, `waterfall_tranche_deductions`, `sgf_assessment_calls`.
- **Temporal Server:** Cluster managing `DefaultWaterfallWorkflow`, `SgfRebalancingWorkflow`, and `DailyStressTestWorkflow`.
- **Apache Kafka Topics:** Consumes `settlement.default_detected.v1`, `wallet.collateral.v1`; publishes `sgf.stress_results.v1`, `sgf.waterfall_initiated.v1`, `sgf.tranche_slashed.v1`, `sgf.corpus_rebalanced.v1`.
- **Trade Settlement Service (Prompt 208):** Alerts SGF Service upon cash or security delivery failure at settlement cut-off.
- **Real-Time VaR Engine (Prompt 229):** Supplies participant peak margin and portfolio exposure metrics.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **On-Chain SGF Smart Contract Sync:** Connects to `SettlementGuaranteeFund.sol` (Prompt 315) on Hyperledger Besu to verify locked on-chain reserve tranches.
- **Cryptographic Slashing Execution:** When Tranche 2 or Tranche 4 is accessed during a default, the service constructs and submits multi-sig authorized transactions calling `slashMemberTranche` or `slashPooledTranche` on-chain.
- **Zero PII Transmission:** Transactions on Hyperledger Besu contain only `member_id_hash` (`bytes32`), `tranche_id` (`uint8`), `slashed_amount_inr` (`uint256`), and `default_id` (`bytes32`). No member names, PANs, or banking details touch the blockchain.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Go module `services/sgf-service` with structured clean architecture (`cmd/`, `internal/domain/`, `internal/workflow/`, `internal/adapter/`).
2. **Define Protobuf Schema:** Create `proto/growww/sgf/v1/sgf_service.proto` defining endpoints for `GetSgfCorpusStatus`, `ExecuteDailyStressTest`, `InitiateDefaultWaterfall`, and `IssueAssessmentCall`.
3. **Generate gRPC Stubs:** Generate Go stubs using `protoc-gen-go` and `protoc-gen-go-grpc`.
4. **Design PostgreSQL Migrations:** Create migration scripts for SGF corpus records, member allocation ledgers, stress test results, and default execution audit logs.
5. **Implement SGF Corpus Sizing Engine:** Code mathematical sizing module calculating Minimum Required Corpus (MRC):
   $$\text{MRC} = \max(\text{Cover1\_StressLoss}, \text{Cover2\_StressLoss}) + \text{LiquidityBuffer} (20\%)$$
6. **Implement Member Quota Allocation Logic:** Compute individual member contributions:
   $$\text{Quota}_i = \text{MRC} \times \left(0.50 \times \frac{\text{Turnover}_i}{\sum \text{Turnover}} + 0.50 \times \frac{\text{PeakMargin}_i}{\sum \text{PeakMargin}}\right)$$
7. **Build Daily Stress Testing Engine:** Implement stress test matrix simulating 20 historical market crises (e.g., 2008 Lehman collapse, 2020 COVID crash) and hypothetical 10-sigma price shocks, calculating worst-case uncollateralized exposure.
8. **Implement Temporal Workflow for Default Waterfall (`DefaultWaterfallWorkflow`):**
 - *Activity 1:* Lock defaulting member account across Order Service and API Gateway.
 - *Activity 2 (Tranche 1):* Seize and liquidate defaulter's initial and MTM margin collateral.
 - *Activity 3 (Tranche 2):* Slash defaulter's SGF contribution on-chain and off-chain.
 - *Activity 4 (Tranche 3):* If deficit remains, draw down Clearing Corporation dedicated equity capital.
 - *Activity 5 (Tranche 4):* If deficit remains, draw down non-defaulting members' SGF pool pro-rata.
 - *Activity 6 (Tranche 5):* If deficit remains, invoke CCP reserve insurance fund.
 - *Activity 7 (Tranche 6):* If deficit remains, issue binding assessment calls to solvent participants.
9. **Build Temporal Activities for Collateral Liquidation:** Interface with Treasury and Custodian services to liquidate seized G-Secs or Blue-chip equity tokens.
10. **Implement On-Chain Contract Relayer:** Build Ethereum JSON-RPC client using `go-ethereum/ethclient` to submit signed default slashing transactions to `SettlementGuaranteeFund.sol`.
11. **Implement Assessment Call Manager:** Generate formal regulatory demand notices and invoice records when non-defaulting participants must replenish slashed pooled funds within 48 hours.
12. **Build Kafka Event Emitters & Listeners:** Ingest `settlement.default_detected.v1` and publish real-time SGF telemetry events.
13. **Implement Prometheus Metrics:** Export `sgf_total_corpus_inr`, `sgf_stress_test_loss_inr`, `sgf_cover2_ratio`, `sgf_waterfall_executions_total`.
14. **Write Integration & Stress Testing Suites:** Write end-to-end simulation tests in Go verifying the full default waterfall execution from Tranche 1 through Tranche 6 under zero failure.

## Interfaces / Contracts

### Protobuf Definition (`sgf_service.proto`)
```protobuf
syntax = "proto3";

package growww.sgf.v1;

option go_package = "growww/sgf/v1;sgfv1";

service SettlementGuaranteeFundService {
  rpc GetSgfCorpusStatus (GetSgfCorpusStatusRequest) returns (GetSgfCorpusStatusResponse);
  rpc CalculateMemberQuotas (CalculateMemberQuotasRequest) returns (CalculateMemberQuotasResponse);
  rpc RunStressTest (RunStressTestRequest) returns (RunStressTestResponse);
  rpc InitiateDefaultWaterfall (InitiateDefaultWaterfallRequest) returns (InitiateDefaultWaterfallResponse);
  rpc GetWaterfallStatus (GetWaterfallStatusRequest) returns (GetWaterfallStatusResponse);
  rpc IssueAssessmentCall (IssueAssessmentCallRequest) returns (IssueAssessmentCallResponse);
}

enum WaterfallStage {
  WATERFALL_STAGE_UNSPECIFIED = 0;
  WATERFALL_STAGE_DEFAULTER_MARGINS = 1;
  WATERFALL_STAGE_DEFAULTER_SGF = 2;
  WATERFALL_STAGE_CC_CAPITAL = 3;
  WATERFALL_STAGE_POOLED_SGF = 4;
  WATERFALL_STAGE_CC_RESERVES_INSURANCE = 5;
  WATERFALL_STAGE_ASSESSMENT_CALLS = 6;
  WATERFALL_STAGE_RESOLVED = 7;
  WATERFALL_STAGE_FAILED = 8;
}

message GetSgfCorpusStatusRequest {}

message GetSgfCorpusStatusResponse {
  string total_corpus_inr = 1;
  string cc_skin_in_the_game_inr = 2;
  string members_total_contribution_inr = 3;
  string available_liquidity_inr = 4;
  string cover1_stress_requirement_inr = 5;
  string cover2_stress_requirement_inr = 6;
  bool is_corpus_adequate = 7;
  int64 updated_at_unix_ns = 8;
}

message CalculateMemberQuotasRequest {
  string calculation_period = 1; // e.g. "2026-09"
}

message MemberQuotaItem {
  string member_id = 1;
  string ledger_address = 2;
  string required_quota_inr = 3;
  string current_deposited_inr = 4;
  string shortfall_inr = 5;
  string turnover_share_percent = 6;
  string peak_margin_share_percent = 7;
}

message CalculateMemberQuotasResponse {
  string total_mrc_inr = 1;
  repeated MemberQuotaItem member_quotas = 2;
}

message RunStressTestRequest {
  string scenario_id = 1; // "LEHMAN_2008", "COVID_2020", "EQUITY_SHOCK_30PCT"
  bool apply_cover2 = 2;
}

message RunStressTestResponse {
  string test_run_id = 1;
  string scenario_name = 2;
  string worst_case_default_loss_inr = 3;
  string largest_defaulter_member_id = 4;
  string second_largest_defaulter_member_id = 5;
  string current_sgf_corpus_inr = 6;
  string corpus_surplus_or_deficit_inr = 7;
  bool is_compliant_with_pfmi = 8;
  int64 executed_at_unix_ns = 9;
}

message InitiateDefaultWaterfallRequest {
  string default_id = 1;
  string defaulting_member_id = 2;
  string defaulting_ledger_address = 3;
  string settlement_cycle_id = 4;
  string total_default_amount_inr = 5;
  string default_reason = 6;
}

message InitiateDefaultWaterfallResponse {
  string waterfall_execution_id = 1;
  string workflow_id = 2;
  WaterfallStage current_stage = 3;
  string total_default_amount_inr = 4;
  string amount_resolved_inr = 5;
  string remaining_deficit_inr = 6;
  int64 started_at_unix_ns = 7;
}

message GetWaterfallStatusRequest {
  string waterfall_execution_id = 1;
}

message TrancheDeductionDetail {
  WaterfallStage stage = 1;
  string tranche_name = 2;
  string amount_drawn_inr = 3;
  string on_chain_tx_hash = 4;
  int64 executed_at_unix_ns = 5;
}

message GetWaterfallStatusResponse {
  string waterfall_execution_id = 1;
  string defaulting_member_id = 2;
  WaterfallStage current_stage = 3;
  string total_default_amount_inr = 4;
  string total_recovered_inr = 5;
  string remaining_deficit_inr = 6;
  repeated TrancheDeductionDetail tranche_deductions = 7;
  bool is_complete = 8;
}

message IssueAssessmentCallRequest {
  string waterfall_execution_id = 1;
  string total_assessment_amount_inr = 2;
  int32 payment_window_hours = 3;
}

message IssueAssessmentCallResponse {
  string assessment_id = 1;
  int32 participants_notified_count = 2;
  string total_assessed_inr = 3;
  int64 due_timestamp_unix = 4;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE sgf_corpus_configs (
    config_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    minimum_required_corpus_inr NUMERIC(18, 2) NOT NULL DEFAULT 500000000.00, -- ₹50 Crore Base
    cc_skin_in_the_game_inr NUMERIC(18, 2) NOT NULL DEFAULT 125000000.00, -- 25% CC capital
    liquidity_buffer_percent NUMERIC(5, 2) NOT NULL DEFAULT 20.00,
    assessment_cap_multiplier NUMERIC(3, 1) NOT NULL DEFAULT 2.0, -- Max 2x contribution
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sgf_member_allocations (
    allocation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id UUID NOT NULL,
    ledger_address VARCHAR(42) NOT NULL,
    required_quota_inr NUMERIC(18, 2) NOT NULL,
    cash_deposited_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    cbdc_deposited_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    gsec_collateral_value_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    total_effective_deposit_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    last_rebalanced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(30) NOT NULL DEFAULT 'COMPLIANT' -- COMPLIANT, SHORTFALL, SUSPENDED
);

CREATE TABLE sgf_stress_test_runs (
    run_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scenario_id VARCHAR(50) NOT NULL,
    scenario_name VARCHAR(100) NOT NULL,
    largest_defaulter_loss_inr NUMERIC(18, 2) NOT NULL,
    second_defaulter_loss_inr NUMERIC(18, 2) NOT NULL,
    total_cover2_loss_inr NUMERIC(18, 2) NOT NULL,
    available_corpus_inr NUMERIC(18, 2) NOT NULL,
    is_corpus_adequate BOOLEAN NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE default_waterfall_executions (
    execution_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    default_id VARCHAR(64) NOT NULL UNIQUE,
    defaulting_member_id UUID NOT NULL,
    defaulting_ledger_address VARCHAR(42) NOT NULL,
    total_default_amount_inr NUMERIC(18, 2) NOT NULL,
    current_stage VARCHAR(50) NOT NULL,
    recovered_amount_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    remaining_deficit_inr NUMERIC(18, 2) NOT NULL,
    temporal_workflow_id VARCHAR(100) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'RUNNING', -- RUNNING, RESOLVED, FAILED
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE waterfall_tranche_deductions (
    deduction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES default_waterfall_executions(execution_id),
    stage_name VARCHAR(50) NOT NULL,
    amount_drawn_inr NUMERIC(18, 2) NOT NULL,
    source_account VARCHAR(100) NOT NULL,
    on_chain_tx_hash VARCHAR(66),
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sgf_assessment_calls (
    assessment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    execution_id UUID NOT NULL REFERENCES default_waterfall_executions(execution_id),
    member_id UUID NOT NULL,
    assessed_amount_inr NUMERIC(18, 2) NOT NULL,
    paid_amount_inr NUMERIC(18, 2) NOT NULL DEFAULT 0.00,
    due_date TIMESTAMPTZ NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING' -- PENDING, PAID, DEFAULTED
);

CREATE INDEX idx_sgf_alloc_member ON sgf_member_allocations(member_id);
CREATE INDEX idx_waterfall_exec_status ON default_waterfall_executions(status);
```

## Security & Compliance Notes
- **CPMI-IOSCO PFMI Principle 4 & 7 Compliance:** Core SGF resources must at all times cover the simultaneous default of the two clearing members creating the largest aggregate credit exposure in extreme but plausible market conditions (Cover-2 standard).
- **Immutable Waterfall Sequence:** The Default Waterfall execution order is statutory and hardcoded. The service cannot bypass Tranche 1 or Tranche 3 to prematurely penalize non-defaulting participants in Tranche 4.
- **Ring-Fenced Segregation Invariant:** SGF collateral deposits are legally and technically segregated in bankruptcy-remote custody accounts and smart contracts, completely isolated from Growww platform operational funds.
- **Temporal Workflow Determinism:** All state transitions during default resolution are recorded in Temporal history, ensuring zero lost state during node reboots or network outages.

## Acceptance Criteria
- [ ] SGF sizing calculation dynamically updates Minimum Required Corpus (MRC) based on worst-case Cover-2 stress losses $+ 20\%$ liquidity buffer.
- [ ] Member contribution quotas are computed monthly using 50% turnover weight and 50% peak margin weight.
- [ ] Temporal workflow executes default waterfall sequence (Tranches 1 through 6) with full transactional audit logging.
- [ ] Multi-sig on-chain transaction dispatched to `SettlementGuaranteeFund.sol` on Hyperledger Besu when slashing Tranches 2 and 4.
- [ ] Assessment calls generated with statutory 2x member contribution cap and 48-hour payment window.
- [ ] Real-time Prometheus metrics exported for SGF corpus health, Cover-2 coverage ratio, and stress test losses.
- [ ] Zero single-point-of-failure: service survives sudden worker process crashes and resumes in-flight default workflows seamlessly.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Standards), Prompt `203` (Wallet Service), Prompt `208` (Trade Settlement Service), Prompt `229` (Real-Time VaR Margin Engine).
- **Parallel Tasks:** Prompt `213` (Custodian Depository Integration), Prompt `232` (CBDC Settlement Adapter).
- **Downstream Blockers:** Prompt `315` (Settlement Guarantee Fund Smart Contract), Prompt `605` (Admin Risk Exception Portal).
