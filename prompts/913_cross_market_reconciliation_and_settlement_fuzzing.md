# 913 - Cross-Market Reconciliation & Automated Settlement Fuzzing Suite

## Purpose
Establishes the definitive specification for the end-to-end automated fuzzing, property-based verification, and adversarial chaos simulation suite (`tests/settlement-fuzzing`). Operating a SEBI and IFSCA compliant dual-entity financial market with atomic Delivery-versus-Payment (DvP) settlement, on-chain Settlement Guarantee Fund (SGF) default waterfalls, high-throughput in-memory order matching, cross-venue Smart Order Routing (SOR), and automated corporate action rebases requires formal runtime invariant verification.

This test suite subjects the core transactional engine to extreme synthetic stress, non-deterministic concurrency, network partitions, adversarial transaction ordering, and out-of-band market events. It proves that the distributed system upholds mathematical invariants of double-entry ledger balance, zero phantom token minting, strict price-time priority, bounded cross-market latency arbitrage exposure, and deterministic corporate action execution under all failure modes.

## What You Are Building
A comprehensive, automated fuzzing and chaos simulation framework located in `tests/settlement-fuzzing/`:
- **DvP Atomic Settlement Fuzzer (`tests/settlement-fuzzing/dvp/`)**: Multi-threaded property-based fuzzing engine testing concurrent atomic settlement, non-blocking ledger rollbacks, relayer RPC drops, nonce race conditions, and EVM revert compensation.
- **SGF Default Waterfall & Stress Simulator (`tests/settlement-fuzzing/sgf/`)**: Stress test suite simulating multi-member simultaneous defaults under market flash crashes, validating progressive margin drawdown, SGF tier depletion, and priority loss allocation algorithms.
- **High-Load Matching Engine Invariant Verifier (`tests/settlement-fuzzing/matching/`)**: High-throughput property test harness firing 100,000+ randomized order operations per second (limit, market, cancel, IOC, FOK) into the Rust matching engine, asserting price-time priority, no-crossed-book states, and determinism.
- **Cross-Venue Latency Arbitrage & SOR Resilience Harness (`tests/settlement-fuzzing/arbitrage/`)**: Network-level chaos harness injecting asymmetric microsecond-to-millisecond latency jitter across simulated NSE, BSE, NBSE, and GIFT City venues, asserting slippage bounds, quote freshness, and toxic flow containment.
- **Corporate Action Split Rebase Race Condition Harness (`tests/settlement-fuzzing/corporate-actions/`)**: Concurrency fuzzer triggering stock split and reverse split rebases concurrently with in-flight matching, pending DvP settlements, and portfolio valuation queries, proving total asset value conservation.
- **Automated Invariant Monitor & Chaos Coordinator (`tests/settlement-fuzzing/coordinator/`)**: Central test runner orchestrating chaos injection (via Chaos Mesh and Toxiproxy) while continuously asserting system-wide conservation laws across PostgreSQL, Redis, Kafka, and Hyperledger Besu.

## Scope Boundaries
- **In Scope:**
  - Property-based generative fuzzing of two-legged DvP settlement state machines across Go settlement services and Solidity contracts.
  - Chaos simulation of clearing member insolvency, collateral haircuts, and multi-tier SGF loss absorption waterfalls.
  - Continuous invariant verification of Rust matching engine order book states under high concurrency and randomized event streams.
  - Latency perturbation testing across simulated multi-exchange gateways validating Smart Order Routing and execution bounds.
  - Concurrency testing of corporate action stock split rebases across order books, pending settlement queues, and user holdings.
  - Automated detection of deadlocks, race conditions, double-spend vulnerabilities, and ledger desynchronization.
- **Out of Scope / Handled Elsewhere:**
  - Standard unit and component integration tests (Prompt 901).
  - Production-scale baseline load and capacity benchmarks without failure injection (Prompt 903).
  - Generic Kubernetes infrastructure chaos and node draining (Prompt 904).
  - Mainnet disaster recovery rehearsal and manual cutover protocol (Prompt 911 dress rehearsal / Prompt 908).
  - Cross-chain bridge SPV and derivatives liquidation stress testing (Prompt 912).

## Technology to Use
- **Property-Based Testing & Fuzzing Frameworks:**
  - **Python Hypothesis & Asyncio Fuzz Engine:** For orchestrating high-level multi-service scenarios, non-deterministic state-machine exploration, and automated minimal failure reproducing shrinking.
  - **Rust `proptest` & `cargo-fuzz` (libFuzzer):** For native in-memory matching engine invariant fuzzing, order book state validation, and memory safety audits under synthetic order streams.
  - **Foundry (Echidna / Medusa / Foundry Invariant Fuzzing):** For property-based invariant testing of `SettlementDvP.sol`, `SettlementGuaranteeFund.sol`, and `DigitalSecurityToken.sol`.
- **Fault Injection & Network Chaos:**
  - **Toxiproxy:** Injects programmable TCP latency, bandwidth limits, jitter, packet corruption, and connection drops between microservices, Redis, Kafka, and Besu RPC nodes.
  - **Chaos Mesh:** Kubernetes-native fault injection for pod kills, disk I/O latency, and container CPU throttling during active fuzz runs.
- **Observability & Invariant Telemetry:**
  - OpenTelemetry metrics, Prometheus, and structured assertion loggers capturing state drift, execution latencies, and invariant violation traces.

*Justification:* Combining Rust native fuzzing for sub-microsecond matching engine verification with Python Hypothesis and Foundry invariant testing provides multi-layered coverage across off-chain financial state machines and on-chain EVM settlement contracts.

## Backend / Infra Touchpoints
- **Services Under Test:**
  - Matching Engine Service (`services/matching-engine`, Rust)
  - Trade Settlement Service (`services/settlement-service`, Go)
  - Wallet & Ledger Service (`services/wallet-service`, Go)
  - Risk & Margin Service (`services/risk-service`, Go)
  - Corporate Actions Service (`services/corporate-actions`, Python)
  - Smart Order Routing Adapter (`services/sor-adapter`, Go)
- **Data Stores & Queues:**
  - PostgreSQL 16+ (Double-entry ledgers, settlement records, holding tables)
  - Redis 7+ Cluster (Order book cache, locks, rate limits)
  - Apache Kafka (Topics: `engine.matches.v1`, `trade.settled.v1`, `sgf.default.v1`, `corporate_action.split.v1`)
- **Blockchain Nodes & Contracts:**
  - Hyperledger Besu local multi-node devnet (QBFT consensus, 4 validators, 1-second block time)
  - Smart contracts: `SettlementDvP.sol`, `SettlementGuaranteeFund.sol`, `DigitalSecurityToken.sol`, `NBSESettlementDvP.sol`

## Blockchain Interaction
- **Permissioned EVM Environment:** Hyperledger Besu private network configured with QBFT consensus and 1-second block generation.
- **DvP Settlement Fuzzing:**
  - Injects asynchronous transaction submission bursts, duplicate transaction nonces, out-of-order execution, and artificial RPC timeout errors.
  - Asserts that for every emitted `executeDvPTrade` call, the smart contract either transfers security tokens and verifies off-chain cash authorization or reverts atomically with zero partial state mutations.
- **Settlement Guarantee Fund (SGF) Contract Invariants:**
  - Invokes `drawdownWaterfall(address defaultingMember, uint256 deficitAmount)` under concurrent multi-member insolvency scenarios.
  - Asserts that capital is seized in strict hierarchical order: Defaulter Collateral -> Defaulter SGF -> Core Exchange SGF -> Non-Defaulter SGF Mutualization.
  - Verifies that total fund balances on-chain exactly match off-chain clearing corporation reserve ledgers.
- **Corporate Action Rebase Interfacing:**
  - Calls `rebaseTokenSupply(address tokenAddress, uint256 multiplier, uint256 divisor)` during active token transfer bursts.
  - Verifies that token balances across all whitelisted accounts scale deterministically without precision loss or truncation vulnerabilities.

## Invariant Formulations & Adversarial Vectors
- **Invariant 1: Conservation of Cash and Asset Balances (DvP Solvency)**
  - For any settlement batch $B$ containing matches $M$:
    $$ \sum_{m \in M} \Delta Cash(Buyer_m) + \sum_{m \in M} \Delta Cash(Seller_m) + \sum_{m \in M} Fees(m) = 0 $$
    $$ \sum_{m \in M} \Delta Tokens(Buyer_m) + \sum_{m \in M} \Delta Tokens(Seller_m) = 0 $$
  - Under no circumstances shall total system cash or total minted security tokens deviate from physical custodial backing.
- **Invariant 2: Price-Time Priority & No Crossed Book**
  - For all bids $B_i$ and asks $A_j$ in the active order book:
    $$ \max_{i}(Price(B_i)) < \min_{j}(Price(A_j)) $$
  - If $Price(B_{incoming}) \ge \min_j(Price(A_j))$, matching occurs immediately at $Price(A_j)$ in order of $Timestamp(A_j)$.
- **Invariant 3: SGF Waterfall Exhaustion Strict Ordering**
  - Drawdown sequence must enforce:
    $$ LossAllocated = \min(L, Collat_{def}) + \min(L_1, SGF_{def}) + \min(L_2, SGF_{core}) + \min(L_3, SGF_{mutual}) $$
    where $L_{k}$ represents residual unabsorbed loss after tier $k$. No tier $k+1$ may be debited if tier $k$ has available capacity.
- **Invariant 4: Corporate Action Valuation Invariance**
  - For a stock split ratio $R = N / D$, the aggregate market value of any position $k$ before and after rebase must remain constant within 1 paise precision:
    $$ Value_{post} = Quantity_{post} \times Price_{post} = (Quantity_{pre} \times \frac{N}{D}) \times (Price_{pre} \times \frac{D}{N}) = Value_{pre} $$
- **Adversarial Vectors Injected:**
  - High-frequency order cancellation racing against matching engine trade execution.
  - Sudden killing of settlement relayer process midway through a two-phase commit.
  - Simulating 90% clearing member default cascading across multiple settlement cycles.
  - Microsecond order front-running and quote cancellation under simulated 200ms cross-venue network delays.
  - Triggering a 1:10 stock split while 500 orders for that security are executing in the matching engine.

## Step-by-Step Build Instructions
1. Initialize test directory structure under `tests/settlement-fuzzing/` with subdirectories for `dvp/`, `sgf/`, `matching/`, `arbitrage/`, `corporate-actions/`, `coordinator/`, and `fixtures/`.
2. Configure local fuzzing environment with Docker Compose spinning up Hyperledger Besu (4-node QBFT), PostgreSQL 16, Redis 7, Apache Kafka, and Toxiproxy.
3. Implement the matching engine property fuzzer in `tests/settlement-fuzzing/matching/` using Rust `proptest`, generating random sequences of limit, market, and cancel orders to assert book orderliness and execution invariants.
4. Implement the DvP atomic settlement fuzzing harness in `tests/settlement-fuzzing/dvp/` using Python Hypothesis, generating random buyer/seller match batches with concurrent wallet hold commits, chain submissions, and simulated network dropouts.
5. Implement the SGF default waterfall simulation engine in `tests/settlement-fuzzing/sgf/` testing cascading member defaults, collateral haircut stress, and mutualized risk pool drawdowns.
6. Implement the cross-venue latency arbitrage simulation in `tests/settlement-fuzzing/arbitrage/` using Toxiproxy to perturb network streams between the Smart Order Router and mock exchange gateways.
7. Implement the corporate action split rebase concurrency fuzzer in `tests/settlement-fuzzing/corporate-actions/` executing stock split events simultaneously with randomized matching and settlement operations.
8. Implement Foundry invariant test contracts in `contracts/test/invariants/` defining stateful fuzzing properties for `SettlementDvP.sol` and `SettlementGuaranteeFund.sol`.
9. Develop the central invariant verification coordinator in `tests/settlement-fuzzing/coordinator/` that monitors double-entry database consistency and ledger proofs before and after every chaos cycle.
10. Integrate automated fuzzing pipelines into continuous integration with configurable execution time budgets (e.g. 30-minute quick fuzz in PR CI, 8-hour deep soak in nightly builds).
11. Build an automated failure artifact generator that serializes reproducing seeds, event replay logs, and minimal failing test cases upon any invariant violation.

## Interfaces / Contracts

### 1. Fuzzing Suite Configuration Schema (`tests/settlement-fuzzing/config.yaml`)
```yaml
version: "1.0"
fuzzing_targets:
  dvp_settlement:
    concurrency_workers: 16
    max_batch_size: 500
    fault_injection:
      rpc_timeout_rate: 0.05
      kafka_rebalance_interval_sec: 30
      redis_disconnect_rate: 0.02
  sgf_waterfall:
    clearing_members_count: 50
    default_probability_per_cycle: 0.15
    market_drop_max_pct: 40.0
    haircut_volatility_multiplier: 1.5
  matching_engine:
    ops_per_second: 100000
    symbol_universe_size: 20
    order_types: ["LIMIT", "MARKET", "CANCEL", "IOC", "FOK"]
    seed: "0xdeadbeef12345678"
  latency_arbitrage:
    venues: ["NSE", "BSE", "NBSE", "GIFT_CITY"]
    jitter_min_ms: 0
    jitter_max_ms: 250
    packet_loss_rate: 0.03
  corporate_actions:
    max_concurrent_splits: 5
    split_ratios: ["1:2", "1:5", "1:10", "10:1", "5:1"]
    orders_in_flight_during_rebase: 1000
```

### 2. Matching Engine Invariant Harness Interface (`tests/settlement-fuzzing/matching/engine_fuzz_harness.rs`)
```rust
use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq)]
pub enum OrderSide {
    Buy,
    Sell,
}

#[derive(Debug, Clone, PartialEq)]
pub enum OrderType {
    Limit,
    Market,
    Cancel,
    ImmediateOrCancel,
    FillOrKill,
}

#[derive(Debug, Clone)]
pub struct FuzzOrderEvent {
    pub order_id: u64,
    pub account_id: u64,
    pub symbol: String,
    pub side: OrderSide,
    pub order_type: OrderType,
    pub price_paise: u64,
    pub quantity: u64,
    pub timestamp_ns: u64,
}

#[derive(Debug, Clone)]
pub struct EngineExecutionReport {
    pub match_id: u64,
    pub maker_order_id: u64,
    pub taker_order_id: u64,
    pub execution_price_paise: u64,
    pub execution_quantity: u64,
    pub timestamp_ns: u64,
}

pub trait MatchingEngineInvariantChecker {
    fn verify_no_crossed_book(&self, symbol: &str) -> Result<(), String>;
    fn verify_price_time_priority(&self, symbol: &str, report: &EngineExecutionReport) -> Result<(), String>;
    fn verify_conservation_of_shares(&self, symbol: &str, initial_positions: &HashMap<u64, i64>) -> Result<(), String>;
    fn verify_deterministic_replay(&self, seed: u64, events: &[FuzzOrderEvent]) -> Result<[u8; 32], String>;
}
```

### 3. DvP Settlement Fuzz Coordinator Interface (`tests/settlement-fuzzing/dvp/dvp_fuzz_coordinator.py`)
```python
from typing import List, Dict, Any, Optional, Protocol
from dataclasses import dataclass
from enum import Enum

class SettlementState(Enum):
    PENDING = "PENDING"
    CASH_HELD = "CASH_HELD"
    CHAIN_SUBMITTED = "CHAIN_SUBMITTED"
    SETTLED = "SETTLED"
    ROLLED_BACK = "ROLLED_BACK"

@dataclass(frozen=True)
class FuzzMatchRecord:
    match_id: str
    symbol: str
    buyer_account_id: str
    seller_account_id: str
    quantity: int
    price_paise: int
    trade_timestamp_ns: int

@dataclass(frozen=True)
class SettlementVerificationResult:
    is_solvent: bool
    cash_delta_sum: int
    token_delta_sum: int
    unresolved_discrepancies: List[str]

class DvPSettlementFuzzCoordinator(Protocol):
    async def inject_concurrent_settlements(
        self,
        matches: List[FuzzMatchRecord],
        chaos_drop_rpc: bool,
        chaos_rebalance_kafka: bool,
    ) -> List[SettlementState]:
        ...

    async def verify_dvp_solvency_invariants(
        self,
        initial_balances: Dict[str, Dict[str, int]],
        batch_id: str,
    ) -> SettlementVerificationResult:
        ...

    async def simulate_rpc_disconnect_and_recover(
        self,
        pending_tx_hash: str,
    ) -> SettlementState:
        ...
```

### 4. SGF Default Waterfall Simulation Interface (`tests/settlement-fuzzing/sgf/sgf_waterfall_simulator.py`)
```python
from typing import List, Dict, Protocol
from dataclasses import dataclass

@dataclass(frozen=True)
class MemberCollateralState:
    member_id: str
    cash_margin_paise: int
    approved_securities_value_paise: int
    sgf_contribution_paise: int
    is_defaulted: bool

@dataclass(frozen=True)
class WaterfallLossDistribution:
    defaulter_collateral_used: int
    defaulter_sgf_used: int
    core_sgf_used: int
    mutualized_sgf_used: int
    uncovered_loss: int

class SGFWaterfallSimulator(Protocol):
    def initialize_clearing_pool(
        self,
        members: List[MemberCollateralState],
        core_exchange_sgf_paise: int,
    ) -> None:
        ...

    def execute_member_default(
        self,
        defaulter_id: str,
        settlement_deficit_paise: int,
        haircut_percentage: float,
    ) -> WaterfallLossDistribution:
        ...

    def verify_waterfall_invariants(
        self,
        distribution: WaterfallLossDistribution,
        statutory_core_sgf_floor_paise: int,
    ) -> bool:
        ...
```

### 5. Corporate Action Rebase Fuzz Interface (`tests/settlement-fuzzing/corporate-actions/rebase_fuzzer.py`)
```python
from typing import List, Dict, Protocol
from dataclasses import dataclass

@dataclass(frozen=True)
class StockSplitEvent:
    symbol: str
    numerator: int
    denominator: int
    effective_timestamp_ns: int

@dataclass(frozen=True)
class PortfolioPosition:
    account_id: str
    symbol: str
    quantity: int
    average_buy_price_paise: int

@dataclass(frozen=True)
class RebaseVerificationReport:
    total_pre_rebase_value_paise: int
    total_post_rebase_value_paise: int
    valuation_delta_paise: int
    rounding_leakage_paise: int
    orders_rebased_count: int
    orders_cancelled_count: int

class CorporateActionFuzzCoordinator(Protocol):
    async def inject_rebase_under_active_trading(
        self,
        split: StockSplitEvent,
        active_order_count: int,
        pending_settlement_count: int,
    ) -> RebaseVerificationReport:
        ...

    def verify_no_valuation_drift(
        self,
        report: RebaseVerificationReport,
        max_allowed_leakage_paise: int = 1,
    ) -> bool:
        ...
```

### 6. Smart Contract Invariant Interface (`contracts/test/invariants/ISettlementInvariants.sol`)
```solidity
// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

interface ISettlementInvariants {
    struct TokenReserveProof {
        address tokenAddress;
        uint256 totalCirculatingTokens;
        uint256 custodialPhysicalShareCount;
        bytes32 merkleRoot;
    }

    function checkDvPSolvencyInvariant(
        bytes32 tradeId,
        address buyer,
        address seller,
        uint256 tokenUnits,
        uint256 inrPaise
    ) external view returns (bool isConserved);

    function checkSGFWaterfallIntegrity(
        address defaultingMember,
        uint256 totalDeficit,
        uint256 coreSGFFloor
    ) external view returns (bool isHierarchyRespected);

    function checkCorporateActionRebaseIntegrity(
        address tokenAddress,
        uint256 multiplier,
        uint256 divisor,
        uint256 preRebaseSupply
    ) external view returns (bool isSupplyAccurate);
}
```

## Security & Compliance Notes
- All fuzzing scenarios must execute strictly in isolated sandbox environments with mock custodial and banking credentials.
- Invariant tests must verify that zero user funds can ever be locked or lost due to unhandled exceptions or state machine race conditions.
- SGF default waterfall algorithms must comply strictly with SEBI Comprehensive Risk Management Framework and SGF minimum capital requirements.
- Full cryptographic audit trails, including execution traces and invariant assertion logs, must be retained for post-run compliance verification.

## Acceptance Criteria
- [ ] DvP settlement fuzzer executes 100,000 randomized trade settlements under 20% simulated network dropouts with zero ledger discrepancy.
- [ ] Matching engine fuzzer validates price-time priority, conservation of shares, and zero crossed order book across 1,000,000 generated operations.
- [ ] SGF simulator correctly resolves cascading multi-member defaults strictly following the 4-tier waterfall without draining core capital below statutory thresholds.
- [ ] Cross-venue latency harness asserts that Smart Order Routing prevents stale quote execution and limits slippage within configured basis point thresholds.
- [ ] Corporate action split rebase fuzzer executes 1:2, 1:5, 1:10, and 10:1 splits during active trading with zero position valuation variance (< 0.01 INR).
- [ ] Foundry invariant test suite runs 50,000 runs without encountering a single state reversion or solvency invariant breach.
- [ ] Automated fuzzer outputs deterministic reproduction seeds and minimal failing traces whenever an invariant check fails.

## Suggested Order / Dependencies
- **Prerequisites:** 205 (Order Matching Engine), 208 (Trade Settlement Service), 215 (Reconciliation Service), 222 (Corporate Actions Service), 241 (SPAN Margin Engine), 306 (SettlementDvP.sol), 315 (SettlementGuaranteeFund.sol), 812 (Mock Depository Suite).
- **Parallel Tasks:** 903 (Load & Performance Testing), 904 (Chaos Engineering Plan), 912 (Cross-Chain Derivatives Stress Testing).
