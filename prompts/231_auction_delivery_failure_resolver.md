# 231 - Depository Fail-to-Deliver Auction & Buy-In Resolution Engine (Rust / Kafka / PostgreSQL)

## Purpose
In modern electronic equity clearing and rolling settlement cycles ($T+1$ / $T+0$), a short delivery occurs when a selling participant fails to deliver the promised physical shares or tokenized securities to the clearing corporation pool account by the mandatory settlement pay-in cut-off time. Without a robust failure resolution mechanism, the purchasing counterparty would suffer non-delivery, undermining market integrity.

In compliance with SEBI Master Circulars on Settlement Mechanism and Auction Handling, the **Depository Fail-to-Deliver Auction & Buy-In Resolution Engine** manages short delivery incidents. When pay-in shortages occur, the engine instantly initiates an automated competitive **Buy-In Auction Session** on settlement morning ($T+1$) to procure shares from the open market. If the auction successfully fulfills the short position, shares are delivered to the buyer, and the defaulting seller is debited for the auction acquisition cost plus statutory penalties. If the auction fails or yields zero offers, the engine automatically executes a **Financial Close-Out / Cash Settlement**, crediting the buyer with the statutory close-out valuation and transferring penalties to the Core Settlement Guarantee Fund (SGF) and Investor Protection Fund (IPF).

## What You Are Building
A deterministic, high-throughput failure resolution microservice (`services/auction-resolver`) written in Rust, integrated with Kafka event streams and PostgreSQL. Concrete deliverables include:
- **Pay-In Shortage Detector:** Analyzes depository pay-in confirmations from NSDL/CDSL (Prompt 213) and token balances against matched trade obligations at settlement cut-off time (e.g., 10:30 AM IST on $T+1$).
- **Isolated Auction Order Book Engine:** Runs dedicated low-latency auction order books for shorted ISINs with maximum bid price caps (standard closing price $+ 20\%$, or the highest price during the settlement cycle).
- **Auction Bid Collector & Matcher:** Collects competitive sell offers from market participants wishing to deliver shares into the auction pool and matches bids using Price-Time priority.
- **Valuation Debit & Penalty Calculator:** Computes total financial liability against the defaulting seller:
  $$\text{Defaulter Debit} = \text{Auction Value} + \text{Clearing Penalty (0.5\% - 2.0\%)} + \text{Depository Penalty}$$
- **Automated Financial Close-Out Engine:** When auction fails or yields insufficient volume, executes statutory cash close-out:
  $$\text{Close-Out Price} = \max(\text{Highest Price during Settlement Cycle}, \text{Auction Day Close}) \times 1.20$$
  and credits the buyer's wallet while debiting the defaulter.
- **On-Chain Settlement Rectification Relayer:** Dispatches smart contract calls to `SettlementDvP.sol` (Prompt 306) to transfer auction-acquired tokens or close out failed trade contracts on Hyperledger Besu.

## Scope Boundaries
- **In Scope:**
 - Automated detection of depository delivery shortages at settlement cut-off.
 - Creation, scheduling, and lifecycle management of Buy-In Auction sessions.
 - Auction order book matching with price band ceilings ($\le +20\%$).
 - Debit of purchase cost and penalties from defaulting sellers' collateral/wallets.
 - Automated statutory financial close-out calculation when auction bids fail.
 - Buyer credit distribution and penalty remittance to Core SGF/IPF.
- **Out of Scope / Handled Elsewhere:**
 - Primary exchange continuous order book matching (handled in Prompt 205).
 - Physical NSDL/CDSL depository demat account clearing interface (handled in Prompt 213).
 - Cash wallet ledger double-entry bookkeeping (handled in Prompt 203).
 - Systemic multi-member default waterfall activation (handled in Prompt 230 / Prompt 315).

## Technology to Use
- **Primary Language & Framework:** **Rust 1.78+** utilizing `tokio` asynchronous runtime and `tonic` gRPC framework. Rust provides deterministic performance, zero garbage-collection pauses during auction execution, and memory safety for high-stakes financial close-outs.
- **In-Memory Auction Book:** Custom Rust BTreeMap-based matching book (`src/auction/book.rs`) optimized for rapid multi-lot auction clearing.
- **Event Streaming:** **Apache Kafka** via `rdkafka` for real-time shortages ingestion, auction ticker broadcast, and settlement resolution event dispatch.
- **Database & Storage:** **PostgreSQL 16+** with `sqlx` connection pool for auction configurations, bid history, penalty distributions, and immutable close-out audit trails.

## Backend / Infra Touchpoints
- **PostgreSQL 16 Tables:** `short_delivery_records`, `auction_sessions`, `auction_bids`, `auction_trade_allocations`, `financial_closeouts`, `settlement_penalty_ledgers`.
- **Apache Kafka Topics:** Consumes `settlement.short_delivery_detected.v1`, `custodian.payin_report.v1`; publishes `auction.session_created.v1`, `auction.trade_executed.v1`, `settlement.closeout_executed.v1`, `auction.penalties_assessed.v1`.
- **Trade Settlement Service (Prompt 208):** Ingests auction settlement outputs to finalize buyer delivery.
- **Wallet & Account Service (Prompt 203):** Executes automated debit against defaulting seller and credit to buyer.
- **Custodian Depository Integration (Prompt 213):** Directs early pay-in allocation of auction-sold shares.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Settlement DvP Rectification:** When shares are procured via auction, the resolver invokes `SettlementDvP.settleTrade` with the auction seller address substituting the defaulting seller on Hyperledger Besu.
- **Close-Out Token Cancellation:** If an auction fails and cash close-out occurs, the engine triggers `SettlementDvP.cancelTrade(tradeId, "FINANCIAL_CLOSEOUT_SHORT_DELIVERY")` to release the buyer's locked cash hold and mark the on-chain trade as financially settled.
- **Zero PII Exposure:** On-chain events and function calls strictly utilize `trade_id` (`bytes32`), `isin`, `defaulting_address`, and `auction_seller_address`.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Service:** Initialize Rust Cargo workspace `services/auction-resolver` with strict compiler lints (`deny(warnings)`).
2. **Define Protobuf Schema:** Create `proto/growww/auction/v1/auction_service.proto` defining `RegisterShortDelivery`, `CreateAuctionSession`, `SubmitAuctionBid`, `ExecuteAuctionMatching`, and `ExecuteFinancialCloseout`.
3. **Generate gRPC Stubs:** Build Rust gRPC client and server bindings using `tonic-build`.
4. **Design PostgreSQL Schema:** Write migrations creating `short_delivery_records`, `auction_sessions`, `auction_bids`, and `financial_closeouts`.
5. **Implement Short-Delivery Ingestion Engine:** Build Kafka consumer processing `settlement.short_delivery_detected.v1` events emitted by Depository Reconciliation (Prompt 215) and Trade Settlement (Prompt 208).
6. **Implement Auction Parameter Calculator:**
 - Determine standard reference price: $P_{\text{ref}} = \text{Previous Day Closing Price}$.
 - Calculate maximum auction ceiling price: $P_{\text{cap}} = \min(P_{\text{ref}} \times 1.20, \text{Highest Trade Price in Settlement Cycle} \times 1.20)$.
7. **Implement In-Memory Auction Order Book:** Build Rust auction book accepting sell bids with validations:
 - Seller must have verified free physical/tokenized equity balance.
 - Bid price must satisfy $P_{\text{floor}} \le P_{\text{bid}} \le P_{\text{cap}}$.
8. **Build Auction Matching Algorithm:** At auction window close (e.g., 12:00 PM IST), execute batch uniform-price or price-time auction matching:
 - Aggregate cumulative sell volume.
 - Clear matching orders against total short-delivery demand.
9. **Implement Valuation Debit & Penalty Engine:** Calculate defaulting member penalties:
   $$\text{Penalty}_{\text{CC}} = \text{Short Value} \times 0.01 \quad (\text{routed to Core SGF})$$
   $$\text{Penalty}_{\text{IPF}} = \text{Short Value} \times 0.005 \quad (\text{routed to Investor Protection Fund})$$
10. **Implement Automated Financial Close-Out Engine:** For any unfilled short quantity:
 - Compute close-out settlement rate: $P_{\text{close}} = P_{\text{cap}}$.
 - Calculate net buyer compensation: $\text{Buyer Credit} = Q_{\text{unfilled}} \times P_{\text{close}}$.
 - Charge defaulter and credit buyer wallet instantly.
11. **Implement On-Chain DvP Resolver Bridge:** Submit transaction to `SettlementDvP.sol` on Hyperledger Besu to complete or close out on-chain trade records.
12. **Build Kafka Event Broadcasting:** Publish detailed auction clearing reports to `auction.trade_executed.v1` and `settlement.closeout_executed.v1`.
13. **Configure Prometheus Metrics:** Export `auction_sessions_active`, `short_deliveries_total`, `auction_fill_rate_percent`, `closeouts_executed_total`, `penalties_collected_inr`.
14. **Write Integration & Simulation Test Suite:** Write comprehensive Rust integration tests testing 100% auction fulfillment, partial auction fulfillment with close-out, and complete auction failure leading to cash close-out.

## Interfaces / Contracts

### Protobuf Definition (`auction_service.proto`)
```protobuf
syntax = "proto3";

package growww.auction.v1;

option go_package = "growww/auction/v1;auctionv1";

service AuctionResolverService {
  rpc RegisterShortDelivery (RegisterShortDeliveryRequest) returns (RegisterShortDeliveryResponse);
  rpc CreateAuctionSession (CreateAuctionSessionRequest) returns (CreateAuctionSessionResponse);
  rpc SubmitAuctionBid (SubmitAuctionBidRequest) returns (SubmitAuctionBidResponse);
  rpc ExecuteAuctionMatching (ExecuteAuctionMatchingRequest) returns (ExecuteAuctionMatchingResponse);
  rpc ExecuteFinancialCloseout (ExecuteFinancialCloseoutRequest) returns (ExecuteFinancialCloseoutResponse);
  rpc GetAuctionSessionStatus (GetAuctionSessionStatusRequest) returns (GetAuctionSessionStatusResponse);
}

enum AuctionStatus {
  AUCTION_STATUS_UNSPECIFIED = 0;
  AUCTION_STATUS_SCHEDULED = 1;
  AUCTION_STATUS_OPEN_FOR_BIDS = 2;
  AUCTION_STATUS_MATCHING = 3;
  AUCTION_STATUS_COMPLETED = 4;
  AUCTION_STATUS_PARTIAL_CLOSEOUT = 5;
  AUCTION_STATUS_FULL_CLOSEOUT = 6;
}

message RegisterShortDeliveryRequest {
  string settlement_cycle_id = 1;
  string original_trade_id = 2;
  string defaulting_member_id = 3;
  string defaulting_ledger_address = 4;
  string buyer_member_id = 5;
  string buyer_ledger_address = 6;
  string isin = 7;
  string short_quantity = 8; // Fractional equity quantity
  string original_trade_price = 9;
}

message RegisterShortDeliveryResponse {
  string short_delivery_id = 1;
  string auction_session_id = 2;
  string maximum_auction_price = 3;
  int64 auction_start_time_unix = 4;
  int64 auction_end_time_unix = 5;
}

message CreateAuctionSessionRequest {
  string isin = 1;
  string settlement_cycle_id = 2;
  string total_short_quantity = 3;
  string reference_closing_price = 4;
  string price_cap = 5;
  int64 bid_window_seconds = 6;
}

message CreateAuctionSessionResponse {
  string auction_session_id = 1;
  AuctionStatus status = 2;
  string price_cap = 3;
}

message SubmitAuctionBidRequest {
  string auction_session_id = 1;
  string seller_user_id = 2;
  string seller_ledger_address = 3;
  string offer_quantity = 4;
  string offer_price = 5; // Must be <= price_cap
}

message SubmitAuctionBidResponse {
  string bid_id = 1;
  bool accepted = 2;
  string rejection_reason = 3;
  int64 submitted_at_unix_ns = 4;
}

message ExecuteAuctionMatchingRequest {
  string auction_session_id = 1;
}

message AuctionMatchedTrade {
  string auction_trade_id = 1;
  string bid_id = 2;
  string seller_user_id = 3;
  string matched_quantity = 4;
  string matched_price = 5;
  string total_inr_value = 6;
}

message ExecuteAuctionMatchingResponse {
  string auction_session_id = 1;
  string total_short_quantity = 2;
  string total_procured_quantity = 3;
  string remaining_unfilled_quantity = 4;
  repeated AuctionMatchedTrade trades = 5;
  AuctionStatus resulting_status = 6;
}

message ExecuteFinancialCloseoutRequest {
  string short_delivery_id = 1;
  string unfilled_quantity = 2;
  string closeout_price = 3;
  string penalty_rate_percent = 4;
}

message ExecuteFinancialCloseoutResponse {
  string closeout_id = 1;
  string buyer_credit_amount_inr = 2;
  string defaulter_debit_amount_inr = 3;
  string core_sgf_penalty_inr = 4;
  string ipf_penalty_inr = 5;
  bool is_settled = 6;
}

message GetAuctionSessionStatusRequest {
  string auction_session_id = 1;
}

message GetAuctionSessionStatusResponse {
  string auction_session_id = 1;
  string isin = 2;
  AuctionStatus status = 3;
  string short_quantity = 4;
  string filled_quantity = 5;
  string price_cap = 6;
  int32 total_bids_received = 7;
}
```

### PostgreSQL Database Schema DDL
```sql
CREATE TABLE short_delivery_records (
    short_delivery_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    settlement_cycle_id VARCHAR(50) NOT NULL,
    original_trade_id VARCHAR(64) NOT NULL,
    defaulting_member_id UUID NOT NULL,
    defaulting_ledger_address VARCHAR(42) NOT NULL,
    buyer_member_id UUID NOT NULL,
    buyer_ledger_address VARCHAR(42) NOT NULL,
    isin VARCHAR(12) NOT NULL,
    short_quantity NUMERIC(18, 6) NOT NULL,
    original_trade_price NUMERIC(18, 4) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING_AUCTION', -- PENDING_AUCTION, AUCTION_FILLED, CLOSED_OUT
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE auction_sessions (
    auction_session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    isin VARCHAR(12) NOT NULL,
    settlement_cycle_id VARCHAR(50) NOT NULL,
    target_short_quantity NUMERIC(18, 6) NOT NULL,
    procured_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    reference_price NUMERIC(18, 4) NOT NULL,
    price_cap NUMERIC(18, 4) NOT NULL, -- Max +20%
    status VARCHAR(30) NOT NULL DEFAULT 'SCHEDULED', -- SCHEDULED, OPEN, MATCHING, COMPLETED, FAILED
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE auction_bids (
    bid_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    auction_session_id UUID NOT NULL REFERENCES auction_sessions(auction_session_id),
    seller_user_id UUID NOT NULL,
    seller_ledger_address VARCHAR(42) NOT NULL,
    offer_quantity NUMERIC(18, 6) NOT NULL,
    offer_price NUMERIC(18, 4) NOT NULL,
    matched_quantity NUMERIC(18, 6) NOT NULL DEFAULT 0.000000,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, MATCHED, CANCELLED, REJECTED
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE financial_closeouts (
    closeout_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    short_delivery_id UUID NOT NULL REFERENCES short_delivery_records(short_delivery_id),
    unfilled_quantity NUMERIC(18, 6) NOT NULL,
    closeout_price NUMERIC(18, 4) NOT NULL,
    buyer_credit_inr NUMERIC(18, 2) NOT NULL,
    defaulter_debit_inr NUMERIC(18, 2) NOT NULL,
    sgf_penalty_inr NUMERIC(18, 2) NOT NULL,
    ipf_penalty_inr NUMERIC(18, 2) NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_short_delivery_isin ON short_delivery_records(isin, status);
CREATE INDEX idx_auction_session_isin ON auction_sessions(isin, status);
CREATE INDEX idx_auction_bids_session ON auction_bids(auction_session_id, offer_price ASC);
```

## Security & Compliance Notes
- **SEBI Settlement Norms & Auction Ceilings:** Auction offers exceeding the statutory $20\%$ price band ceiling above the reference price must be automatically rejected at the API gateway layer.
- **Strict Anti-Collusion Safeguard:** The defaulting seller is strictly prohibited from participating as a bidder in the auction session for their own shorted security.
- **Ring-Fenced Penalty Remittance:** Penalties collected from defaulting sellers must be accounted for and deposited directly into the Core SGF and Investor Protection Fund accounts within 24 hours of close-out.
- **Deterministic Cash Finality:** If neither shares nor auction offers are available, financial close-out provides definitive, non-reversible cash settlement to the buyer, preventing open-ended delivery delays.

## Acceptance Criteria
- [ ] Depository pay-in shortages are automatically detected at settlement cut-off and registered into `short_delivery_records`.
- [ ] Dedicated auction sessions are scheduled and price caps calculated ($\le +20\%$) within 500ms of shortage detection.
- [ ] Auction order book accepts, validates, and matches competitive bids in Price-Time priority at auction window close.
- [ ] Defaulter is charged the exact auction execution value $+ 1.0\%$ Core SGF penalty $+ 0.5\%$ IPF penalty.
- [ ] Unfilled auction shortages automatically trigger statutory Financial Close-Out with $\max(\text{Highest Cycle Price}, \text{Auction Close}) \times 1.20$.
- [ ] Buyer wallet is credited with close-out compensation, and on-chain trade is closed out on Hyperledger Besu.
- [ ] PostgreSQL audit tables record every short delivery, bid submission, trade allocation, and financial close-out.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `103` (API Standards), Prompt `203` (Wallet Service), Prompt `208` (Trade Settlement Service), Prompt `213` (Custodian Depository Integration).
- **Parallel Tasks:** Prompt `229` (Real-Time VaR Engine), Prompt `230` (SGF Service).
- **Downstream Blockers:** Prompt `306` (Settlement DvP Smart Contract close-out integration), Prompt `215` (Reconciliation Service).
