# 222 - Corporate Actions Service (Dividends, Splits, Bonuses & Token Adjustments)

## Purpose
The Corporate Actions Service automates the detection, entitlement calculation, ledger accounting, and on-chain token supply adjustments for mandatory and voluntary corporate events announced by Indian listed corporations. These events include cash dividends, stock splits (sub-divisions), bonus share issues, rights issues, and mergers/demergers.

In a fractional investment platform where investors hold arbitrary fractional units (e.g. 0.4578 shares of an equity), distributing corporate actions requires mathematical precision down to 18 decimal places. The service ensures that fractional cash dividends are credited accurately to investor INR wallets, while stock splits and bonus issues trigger proportional adjustments to off-chain holdings and on-chain `DigitalSecurityToken.sol` token balances without dilution or loss of 1:1 custody backing.

## What You Are Building
A deterministic Python/Go microservice (`services/corporate-actions-service`) providing:
- **Corporate Action Announcement Ingestion:** Ingests official corporate action notifications from BSE, NSE, and NSDL/CDSL feeds.
- **Record Date Entitlement Snapshot Engine:** Freezes investor fractional holding balances at the official Record Date cutoff timestamp.
- **Fractional Cash Dividend Distributor:** Calculates proportional dividend payouts, and credits user wallets with 100% net dividend without on-chain TDS withholding.
- **On-Chain Token Split & Bonus Relayer:** Interacts with Hyperledger Besu smart contracts to execute token multiplier re-denominations or bonus unit mints.
- **Artifacts Delivered:**
 - `services/corporate-actions-service/cmd/server/main.go` (or Python service).
 - `services/corporate-actions-service/internal/entitlement/calculator.go` - High-precision fractional entitlement engine.
 - `services/corporate-actions-service/internal/dividend/payout_engine.go` - Dividend allocation module (0% TDS).
 - `services/corporate-actions-service/internal/chain/rebase_client.go` - Smart contract token split caller.
 - `proto/growww/corporate_actions/v1/corporate_actions.proto` - Internal gRPC service definitions.

## Scope Boundaries
- **In Scope:**
 - Parsing exchange and depository corporate action announcement feeds.
 - Snapshotting fractional holdings at Record Date.
 - Computing fractional cash dividend entitlements with zero TDS withholding.
 - Orchestrating on-chain stock split rebase or bonus token issuance on Hyperledger Besu.
 - Emitting wallet credit and notification events.
- **Out of Scope / Handled Elsewhere:**
 - Direct banking bank-account payout execution (handled by Prompt 212).
 - Annual capital gains tax report generation (handled by Prompt 223).
 - Custodian Demat pool account synchronization (handled by Prompt 213).

## Technology to Use
- **Primary Language & Framework:** Python 3.12+ (using `decimal.Decimal` arbitrary-precision arithmetic) or Go 1.22+ utilizing `math/big` and `pgx/v5`.
- **Justification:** Arbitrary precision is non-negotiable when distributing fractional corporate action dividends and stock splits across millions of accounts. Using standard IEEE 754 floating-point numbers would introduce rounding errors that violate SEBI fractional balance accounting rules.
- **Dependencies & Libraries:**
 - PostgreSQL 16+ for storing corporate action event lifecycles, entitlement snapshots, and payout ledgers.
 - Hyperledger Besu Web3 client (`web3.py` or `go-ethereum`).
 - Redis 7.2+ for snapshot locking and duplicate distribution prevention.
 - Apache Kafka 3.7+ for emitting financial ledger and notification events.

## Backend / Infra Touchpoints
- **PostgreSQL 16:** Tables `corporate_actions`, `action_snapshots`, `action_entitlements`, `dividend_payouts`.
- **Hyperledger Besu (QBFT):** Invokes split/rebase functions on `DigitalSecurityToken.sol` and records corporate action hashes on `ProofOfReserveRegistry.sol`.
- **Depository Feeds:** NSDL/CDSL corporate action announcement XML/SFTP feeds.
- **Kafka Topics:**
 - Subscribes: `custody.corporate_action.announced`, `admin.action.approved`.
 - Publishes: `corporate_action.entitlements.calculated`, `wallet.credit.dividend`, `corporate_action.completed`.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Target Network:** Hyperledger Besu permissioned consortium network running QBFT consensus.
- **Smart Contract Adjustments:**
 - **Stock Splits (e.g. 1:10 split):** The service initiates an on-chain proposal to call `rebaseTokenSupply(isin, multiplier)` or `executeSplit(ratio)` on `DigitalSecurityToken.sol`, updating all token balances proportionally while preserving individual ownership percentages.
 - **Bonus Issues (e.g. 1:1 bonus):** Initiates a smart contract mint to distribute bonus tokens to investor wallet addresses in lockstep with custodian physical share receipts.
- **Custody Reconciliation:** Confirms that on-chain total supply post-action equals physical Demat shares in NSDL/CDSL custody.
- **Zero PII:** On-chain calls reference only ISINs, token contract addresses, split ratios, and wallet addresses.

## Step-by-Step Build Instructions
1. Scaffold project repository with clean domain layers and arbitrary-precision math libraries.
2. Define Protobuf definitions in `proto/growww/corporate_actions/v1/corporate_actions.proto` and generate client stubs.
3. Configure PostgreSQL schema migrations for `corporate_actions`, `holding_snapshots`, and `payout_records`.
4. Implement the Corporate Action Feed Ingester parsing exchange notifications (ex-date, record date, dividend amount, split ratio).
5. Build the Record Date Holding Snapshot module that queries the Portfolio Service (Prompt 209) at the exact market cutoff time.
6. Implement the Fractional Dividend Entitlement Calculator with 18-decimal precision:
   $$\text{Gross Dividend} = \text{Fractional Units} \times \text{Dividend Per Share}$$
7. Implement the zero-TDS dividend distribution module crediting 100% of dividends directly to user wallets.
8. Build the Maker-Checker review pipeline requiring Compliance Officer approval before executing mass payouts (Prompt 217).
9. Implement the Kafka event dispatcher emitting `wallet.credit.dividend` events consumed by Wallet Service (Prompt 203).
10. Implement the On-Chain Token Rebase / Bonus Relayer calling `DigitalSecurityToken.sol` on Hyperledger Besu via HSM.
11. Add Prometheus metrics (`corporate_actions_processed_total`, `dividend_distributed_inr_total`) and health checks.
12. Write rigorous automated test cases simulating fractional splits (e.g., 3:7 reverse split, 1:10 split) and dividend rounding edge cases.

## Interfaces / Contracts

### Protobuf Definition (`corporate_actions.proto`)
```protobuf
syntax = "proto3";

package growww.corporate_actions.v1;

option go_package = "github.com/growww/services/corporate-actions/gen/v1;corporateactionsv1";

service CorporateActionsService {
  rpc DeclareCorporateAction (DeclareActionRequest) returns (DeclareActionResponse);
  rpc CalculateEntitlements (CalculateEntitlementsRequest) returns (CalculateEntitlementsResponse);
  rpc ExecuteDividendPayout (ExecuteDividendRequest) returns (ExecuteDividendResponse);
  rpc ExecuteTokenSplit (ExecuteSplitRequest) returns (ExecuteSplitResponse);
  rpc GetCorporateActionStatus (ActionStatusRequest) returns (ActionStatusResponse);
}

message DeclareActionRequest {
  string isin = 1;
  string action_type = 2; // CASH_DIVIDEND / STOCK_SPLIT / BONUS_ISSUE / RIGHTS_ISSUE
  string ex_date = 3;     // YYYY-MM-DD
  string record_date = 4; // YYYY-MM-DD
  string dividend_per_share_inr = 5; // For cash dividends
  string split_ratio_numerator = 6;  // e.g. "10" for 10:1 split
  string split_ratio_denominator = 7; // e.g. "1"
  string announcement_ref = 8;
}

message DeclareActionResponse {
  string action_id = 1;
  string status = 2; // ANNOUNCED / PENDING_RECORD_DATE
  int64 declared_at = 3;
}

message CalculateEntitlementsRequest {
  string action_id = 1;
}

message CalculateEntitlementsResponse {
  string action_id = 1;
  int64 total_eligible_investors = 2;
  string total_gross_payout_inr = 3;
  string total_tds_deducted_inr = 4; // Always 0 (Zero-TDS)
  string status = 5; // ENTITLEMENTS_CALCULATED
}

message ExecuteDividendRequest {
  string action_id = 1;
  string maker_admin_id = 2;
  string checker_admin_id = 3;
}

message ExecuteDividendResponse {
  string action_id = 1;
  int64 total_credited_investors = 2;
  string total_disbursed_inr = 3;
  string status = 4; // COMPLETED
  int64 completed_at = 5;
}

message ExecuteSplitRequest {
  string action_id = 1;
  string maker_admin_id = 2;
  string checker_admin_id = 3;
}

message ExecuteSplitResponse {
  string action_id = 1;
  string on_chain_tx_hash = 2;
  string new_total_token_supply = 3;
  string status = 4; // EXECUTED_ON_CHAIN
}

message ActionStatusRequest {
  string action_id = 1;
}

message ActionStatusResponse {
  string action_id = 1;
  string isin = 2;
  string action_type = 3;
  string status = 4;
  int64 record_date_timestamp = 5;
  string on_chain_tx_hash = 6;
}
```

### PostgreSQL Database Schema
```sql
CREATE TABLE corporate_actions (
    action_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    action_type VARCHAR(32) NOT NULL CHECK (action_type IN ('CASH_DIVIDEND', 'STOCK_SPLIT', 'BONUS_ISSUE', 'RIGHTS_ISSUE')),
    ex_date DATE NOT NULL,
    record_date DATE NOT NULL,
    dividend_per_share NUMERIC(18, 4),
    split_numerator INT,
    split_denominator INT,
    status VARCHAR(32) NOT NULL DEFAULT 'ANNOUNCED',
    on_chain_tx_hash VARCHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE action_entitlements (
    entitlement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_id UUID NOT NULL REFERENCES corporate_actions(action_id),
    user_id VARCHAR(64) NOT NULL,
    wallet_address VARCHAR(42) NOT NULL,
    snapshot_holding_units NUMERIC(28, 18) NOT NULL,
    gross_payout_inr NUMERIC(28, 4) NOT NULL,
    tds_deducted_inr NUMERIC(28, 4) NOT NULL DEFAULT 0.0000,
    net_payout_inr NUMERIC(28, 4) NOT NULL,
    payout_status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_action_user UNIQUE (action_id, user_id)
);
```

## Security & Compliance Notes
- **Section 194 IT Act TDS Compliance:** Executes 100% full dividend payout with zero withholding and emits statement records (Prompt 223).
- **Dual-Control Execution:** Payout execution and on-chain token rebase require approval from two distinct compliance officers (Maker-Checker).
- **Arbitrary Precision:** All mathematical computations enforce 18-decimal precision with deterministic rounding modes (`ROUND_HALF_EVEN`) to eliminate fractional value leaks.
- **Cryptographic Audit Trail:** All dividend credit logs and token rebase transactions are anchored to the immutable Audit Log Service (Prompt 218).

## Acceptance Criteria
- [ ] Service builds cleanly and executes with arbitrary decimal precision arithmetic.
- [ ] Record Date snapshot captures exact user fractional holdings with zero discrepancies.
- [ ] Fractional dividend distribution correctly credits 100% payout with zero TDS.
- [ ] Stock split rebase correctly multiplies off-chain holdings and executes `DigitalSecurityToken.sol` rebase on Hyperledger Besu.
- [ ] Post-split on-chain token supply matches physical shares in NSDL/CDSL Demat custody exactly.
- [ ] Dual-control Maker-Checker requirement is strictly enforced before financial disbursement.
- [ ] Test coverage exceeds >=85%.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt 203 (Wallet Service), Prompt 209 (Holdings Service), Prompt 213 (Custody Adapter), Prompt 303 (Token Issuance).
- **Subsequent / Parallel Tasks:** Prompt 215 (Reconciliation Service), Prompt 223 (Tax Reporting Service).
