# 341 - On-Chain Circuit Breaker, Timelock & Multi-Tier Emergency Pause Contract (Solidity)

## Purpose
In mission-critical financial market infrastructure, rapid and deterministic containment of systemic risk is paramount. On permissioned distributed ledgers facilitating fractional equity settlements, anomalous market dynamics-such as flash crashes, algorithmic runaway feedback loops, external oracle divergence, smart contract vulnerability exploits, or off-chain depository synchronization desynchronizations (e.g., NSDL/CDSL share count mismatches)-can lead to catastrophic, irreversible settlement default if not halted within milliseconds.

Traditional stock exchanges rely on centralized regulatory halts to maintain orderly markets. The Growww National Blockchain Stock Exchange (NBSE) deployed on Hyperledger Besu requires an immutable, verifiable, on-chain safety mechanism operating in strict compliance with the **SEBI Master Circular on Market-Wide Circuit Breakers (MWCB)** (Circular SEBI/HO/MRD/DP/CIR/P/2018/142 & Circular CIR/MRD/DP/25/2013) and **IFSCA Market Infrastructure Regulations**.

This prompt specifies the enterprise implementation of the **On-Chain Circuit Breaker, Timelock & Multi-Tier Emergency Pause Contract (`CircuitBreakerHalt.sol`)**. The contract provides autonomous and governance-triggered emergency trading halts across multiple granular tiers: single-ISIN securities, entire industry sectors, market-wide indices (enforcing statutory 10%, 15%, and 20% SEBI circuit filters with time-staged cooldowns and call-auction price discovery intervals), and global ledger emergency halts. It enforces asymmetric operational privileges: immediate, single-signer autonomous tripping by authorized automated sentinels, paired with rigorous M-of-N multisig timelocked overrides for safe resumption.

## What You Are Building
A production-grade, modular, upgradeable Solidity smart contract suite under `contracts/src/governance/` comprising:
- `CircuitBreakerHalt.sol`: Master circuit breaker smart contract implementing multi-tiered emergency halt state machines, SEBI index circuit limit tracking, cooldown timers, call-auction readiness flags, and granular ISIN/sector registry mappings.
- `ICircuitBreakerHalt.sol`: Complete Solidity interface exposing query view methods, emergency halt triggers, timelocked resumption routines, custom errors, and comprehensive event schemas.
- Granular Multi-Tier Halt Hierarchy:
  - **Tier 0 (Single-ISIN Halt):** Freezes settlements and secondary transfers for a specific tokenized security due to single-stock volatility collar breach, depository mismatch, or corporate action freeze.
  - **Tier 1 (Sector / Asset Class Halt):** Freezes an entire basket of securities (e.g., Banking, IT, Commodity tokens, Sovereign G-Secs) in response to localized sector contagion.
  - **Tier 2 (Market-Wide Circuit Breaker - MWCB):** Enforces exchange-wide halts matching SEBI index rules (10%, 15%, 20% drops in benchmark composite index), managing scheduled cooldown windows and call-auction pre-opening states.
  - **Tier 3 (Global Emergency Freeze):** Instantaneous ledger-level shutdown of all DvP settlements and token movements across the consortium during critical infrastructure failure, bridge exploits, or validator emergencies.
- Modifier & Guard Injection: Standardized `whenNotHalted(bytes32 isin)` and `requireNotHalted(bytes32 isin)` guards integrated directly into `SettlementDvP.sol` (Prompt 306/329) and `TokenizedEquity.sol` / `DigitalSecurityToken.sol` (Prompt 303).
- Timelock & Governance Integration: Native binding to OpenZeppelin `TimelockController` and `MultiSigGovernance.sol` (Prompt 307) ensuring that resumption of trading requires statutory cooldown completion or 3-of-5 threshold multisig approval.
- Comprehensive Foundry Test Harness (`test/governance/CircuitBreakerHalt.t.sol`): Invariant fuzzing, multi-tier state machine transitions, reentrancy defense, and role-based access verification under high-stress simulated market conditions.

## Scope Boundaries
- **In Scope:**
  - On-chain state machine managing Tier 0 (ISIN), Tier 1 (Sector), Tier 2 (Index MWCB 10%/15%/20%), and Tier 3 (Global) halts.
  - Autonomous trip execution via role-based access control (`SURVEILLANCE_SENTINEL_ROLE`, `ORACLE_SENTINEL_ROLE`, `DEPOSITORY_SENTINEL_ROLE`, `EMERGENCY_GUARDIAN_ROLE`).
  - Safe resumption workflows with mandatory cool-off enforcement, call-auction readiness flags, and M-of-N governance override (`GOVERNANCE_TIMELOCK_ROLE`).
  - Direct modifier integration `whenNotHalted(bytes32 isin)` and `isHalted(bytes32 isin)` for settlement and token contracts.
  - Granular halt state tracking with timestamps, trigger reason codes, breaker levels, and scheduled cooldown expiration.
  - Emitting rich audit events: `ISINHalted`, `ISINResumed`, `MarketWideHaltTriggered`, `MarketWideHaltResumed`, `GlobalEmergencyFreezeTriggered`, `GlobalEmergencyFreezeLifted`, `HaltParameterUpdated`.
  - Foundry unit, property, and invariant fuzz test suites.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain surveillance and statistical anomaly detection algorithms (handled in Prompt 228).
  - Continuous Circuit Breaker Coordination and Call-Auction price discovery state machine (handled in Prompt 713).
  - Off-chain matching engine partition freezing and resting quote cancellation (handled in Prompt 205 and Prompt 226).
  - External exchange Bhavcopy ingestion and daily collar calculation (handled in Prompt 242).
  - Manual UI dashboard for compliance officers and back-office operations (handled in Prompt 217).

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Shanghai/Cancun with transient storage opcodes `TSTORE`/`TLOAD`, custom user-defined value types, and native custom errors for gas minimization).
- **Base Frameworks & Libraries:** 
  - OpenZeppelin Contracts Upgradeable v5.0 (`AccessControlUpgradeable`, `UUPSUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`).
  - OpenZeppelin Contracts v5.0 `TimelockController` for time-delayed governance executions.
- **Development & Testing Toolchain:** **Foundry** (`forge` for compilation, property-based testing, and invariant fuzzing; `cast` for Besu JSON-RPC interaction and diagnostic scripting).
- **Static Analysis & Formal Verification:** Slither, Mythril, and Solhint automated CI pipelines to guarantee zero unhandled edge cases or privilege escalations.

## Backend / Infra Touchpoints
- **Continuous Market Circuit Breaker Coordinator (Prompt 713):** High-throughput microservice monitoring real-time VWAP and index ticks; publishes high-priority signed transactions to trigger Tier 0 or Tier 2 halts within < 5ms of threshold breach.
- **Real-Time Market Surveillance Engine (Prompt 228):** Ingests tick-by-tick order book data, flagging market manipulation (spoofing, layering, runaway order bursts); dispatches emergency pause commands via relayer holding `SURVEILLANCE_SENTINEL_ROLE`.
- **Admin Back-Office Service (Prompt 217):** Web portal utilized by exchange risk officers and compliance directors to view real-time halt states, inspect reason codes, initiate manual pauses, or submit timelocked unpause proposals to the multi-sig.
- **Custodian Depository Integration Service (Prompt 213) & Reconciliation Service (Prompt 215):** Triggers `depositoryHalt(bytes32 isin, bytes32 mismatchHash)` when off-chain demat holdings at NSDL/CDSL disagree with token supply on Besu.
- **Decentralized Multi-Source Oracle Aggregator (Prompt 328):** Invokes `oracleDivergenceHalt(bytes32 isin, uint256 maxDeviationBps)` if multi-oracle median prices deviate beyond confidence boundaries.
- **Blockchain Event Indexer (Prompt 309):** Ingests all halt and resumption events, updating the Redis circuit breaker cache and alerting client apps within 50ms.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Consensus & Finality:** Operates on Hyperledger Besu private consortium network with QBFT consensus, 2-second deterministic block times, and instant absolute finality (zero reorganizations).
- **Modifier Injection:** `SettlementDvP.sol` and `TokenizedEquity.sol` inherit or interface with `CircuitBreakerHalt.sol` using:
  ```solidity
  modifier whenNotHalted(bytes32 isin) {
      if (circuitBreaker.isHalted(isin)) {
          revert TradingHalted(isin);
      }
      _;
  }
  ```
- **Asymmetric Privilege Architecture:** Immediate, single-signer tripping by automated HSM-backed sentinel accounts (`SURVEILLANCE_SENTINEL_ROLE`, `ORACLE_SENTINEL_ROLE`, `DEPOSITORY_SENTINEL_ROLE`) ensures zero delay during flash crashes. In contrast, resumption requires either the automatic expiration of statutory SEBI cooldown timers or explicit 3-of-5 multisig timelock execution (`GOVERNANCE_TIMELOCK_ROLE`).
- **Zero On-Chain PII:** The contract deals strictly with financial instruments and protocol states. All halt triggers record only `bytes32 indexed isin`, `uint8 tier`, `uint8 reasonCode`, `bytes32 detailsHash` (IPFS hash of the off-chain surveillance incident dossier), and `uint256 timestamp`. Zero personal data is stored on-chain.
- **1:1 Custody Backing Assurance:** Whenever a depository synchronization discrepancy is detected, the ISIN is halted instantly, preventing trading of unbacked or duplicate tokens until physical custodial audit passes.

## Step-by-Step Build Instructions (10-15 steps)
1. **Initialize Project & Scaffolding:** Set up Foundry directory structure: `contracts/src/governance/CircuitBreakerHalt.sol`, `contracts/src/interfaces/ICircuitBreakerHalt.sol`, and `test/governance/CircuitBreakerHalt.t.sol`.
2. **Define Data Structures & Enums:** Declare `HaltTier` (TIER_0_ISIN, TIER_1_SECTOR, TIER_2_MARKET_WIDE, TIER_3_GLOBAL), `HaltReason` (VOLATILITY_COLLAR, ORACLE_DIVERGENCE, DEPOSITORY_DESYNC, SURVEILLANCE_ANOMALY, MANUAL_GOVERNANCE, INFRASTRUCTURE_FAULT), and structs for `HaltStatus`, `MarketWideHaltConfig`, and `ISINConfig`.
3. **Configure Role-Based Access Control:** Implement OpenZeppelin `AccessControlUpgradeable` with distinct roles: `DEFAULT_ADMIN_ROLE`, `EMERGENCY_GUARDIAN_ROLE`, `SURVEILLANCE_SENTINEL_ROLE`, `ORACLE_SENTINEL_ROLE`, `DEPOSITORY_SENTINEL_ROLE`, and `GOVERNANCE_TIMELOCK_ROLE`.
4. **Implement Global Emergency Freeze (Tier 3):** Build `triggerGlobalEmergencyFreeze(bytes32 reasonHash)` and `liftGlobalEmergencyFreeze()` functions. Use storage flags and transient storage to instantly block all operations platform-wide.
5. **Implement Market-Wide Circuit Breaker State Machine (Tier 2):**
   - Encode statutory SEBI limits (10%, 15%, 20% index drops).
   - Implement stage-specific cooldown duration tracking based on block timestamp and trading session phase (pre-1:00 PM, 1:00 PM - 2:30 PM, post-2:30 PM).
   - Enforce mandatory call-auction transition state (`CALL_AUCTION_DISCOVERY`) before full resumption to normal trading (`NORMAL_TRADING`).
6. **Implement Sector & Basket Halts (Tier 1):** Map individual ISINs to unique `bytes32 sectorId`. Implement `haltSector(bytes32 sectorId, HaltReason reason, bytes32 detailsHash)` and `resumeSector(bytes32 sectorId)`.
7. **Implement Granular Single-ISIN Halts (Tier 0):** Build `haltISIN(bytes32 isin, HaltReason reason, uint32 durationSeconds, bytes32 detailsHash)` enabling granular pauses triggered by matching engines, surveillance daemons, or depository bridges.
8. **Implement Specialized Sentinel Entrypoints:**
   - `depositoryHalt(bytes32 isin, bytes32 mismatchProofHash)` restricted to `DEPOSITORY_SENTINEL_ROLE`.
   - `oracleDivergenceHalt(bytes32 isin, uint256 deviationBps)` restricted to `ORACLE_SENTINEL_ROLE`.
   - `surveillanceVolatilityHalt(bytes32 isin, uint256 priceDropBps)` restricted to `SURVEILLANCE_SENTINEL_ROLE`.
9. **Implement Comprehensive Query & Enforcement Views:**
   - Build `isHalted(bytes32 isin) external view returns (bool, HaltTier, uint256 expiresAt)` which hierarchically checks Global Freeze $\to$ Market-Wide Halt $\to$ Sector Halt $\to$ ISIN Halt.
   - Implement `requireNotHalted(bytes32 isin) external view` that reverts with descriptive custom errors if any tier is active.
10. **Implement Timelocked Resumption & Multisig Overrides:** Integrate OpenZeppelin `TimelockController` address. Require that manual unhalting before statutory cooldown expiration can only be executed by `GOVERNANCE_TIMELOCK_ROLE` after the timelock delay has elapsed.
11. **Inject Modifiers into SettlementDvP & TokenizedEquity:** Expose `ICircuitBreakerHalt` reference in `SettlementDvP.sol` and `TokenizedEquity.sol`. Apply `whenNotHalted(isin)` to `settleTrade()`, `settleBatch()`, and ERC-3643 transfer hooks.
12. **Build Foundry Unit & Mock Suite:** Create mocks for `MockSettlementDvP.sol` and `MockOracleAggregator.sol`. Write unit tests covering every transition, role authorization check, and custom error revert.
13. **Implement Invariant & Fuzz Tests:** Write property-based invariant tests in Foundry asserting that:
    - If `globalEmergencyFreeze` is true, no ISIN can ever return `isHalted == false`.
    - An expired Tier 0 halt automatically unlocks without requiring explicit transaction, provided cooldown has passed and no higher tier is active.
    - Non-sentinel addresses can never invoke halt functions.
14. **Conduct Slither & Static Analysis Auditing:** Execute Slither static analysis with zero warnings on reentrancy, access control shadow variables, or state variable layout corruption.
15. **Formulate Besu Deployment Script:** Write Foundry deployment script `script/DeployCircuitBreaker.s.sol` configuring initial multi-sig addresses, timelock parameters, sentinel relays, and wiring into existing settlement contracts.

## Interfaces / Contracts

### Circuit Breaker Interface (`ICircuitBreakerHalt.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

interface ICircuitBreakerHalt {
    /// @notice Halt tier levels ordered from least to most severe
    enum HaltTier {
        NONE,               // 0: Normal operations
        TIER_0_ISIN,        // 1: Individual security halted
        TIER_1_SECTOR,      // 2: Entire industry sector halted
        TIER_2_MARKET_WIDE, // 3: System-wide market index halt (SEBI 10%, 15%, 20%)
        TIER_3_GLOBAL       // 4: Total consortium ledger freeze
    }

    /// @notice Statutory and operational trigger reasons
    enum HaltReason {
        MANUAL_GOVERNANCE,        // Administrative or regulatory order
        VOLATILITY_COLLAR,        // LULD / dynamic price band breach
        ORACLE_DIVERGENCE,        // Median-of-N oracle feed mismatch
        DEPOSITORY_DESYNC,        // Physical demat custody vs token balance mismatch
        SURVEILLANCE_ANOMALY,     // Spoofing, layering, or market manipulation alert
        INFRASTRUCTURE_FAULT,     // Bridge exploit, validator failure, or node desync
        SEBI_INDEX_CIRCUIT_10,    // Market-wide index 10% movement
        SEBI_INDEX_CIRCUIT_15,    // Market-wide index 15% movement
        SEBI_INDEX_CIRCUIT_20     // Market-wide index 20% movement
    }

    /// @notice Market trading session state for SEBI MWCB tracking
    enum MarketSessionState {
        NORMAL_TRADING,
        VOLATILITY_HALT,
        CALL_AUCTION_DISCOVERY,
        MARKET_CLOSED_FOR_DAY
    }

    /// @notice State container for an individual ISIN halt
    struct ISINHaltStatus {
        bool isHalted;
        HaltReason reason;
        uint64 haltedAt;
        uint64 expiresAt;
        bytes32 detailsHash;       // IPFS/cryptographic hash of incident dossier
        address haltedBy;
    }

    /// @notice State container for Market-Wide Circuit Breaker (Tier 2)
    struct MarketWideHaltState {
        MarketSessionState sessionState;
        uint8 circuitLevel;        // 10, 15, or 20
        uint64 haltedAt;
        uint64 cooldownExpiresAt;
        uint64 auctionExpiresAt;
        bytes32 triggerIndex;      // e.g., keccak256("NIFTY_50_COMPOSITE")
        uint256 indexLevelPaise;   // Index level at trigger (2 decimal precision)
    }

    // --- Events ---
    event ISINHalted(
        bytes32 indexed isin,
        HaltReason indexed reason,
        uint64 indexed haltedAt,
        uint64 expiresAt,
        bytes32 detailsHash,
        address haltedBy
    );

    event ISINResumed(
        bytes32 indexed isin,
        uint64 indexed resumedAt,
        address resumedBy
    );

    event SectorHalted(
        bytes32 indexed sectorId,
        HaltReason indexed reason,
        uint64 indexed haltedAt,
        bytes32 detailsHash,
        address haltedBy
    );

    event SectorResumed(
        bytes32 indexed sectorId,
        uint64 indexed resumedAt,
        address resumedBy
    );

    event MarketWideHaltTriggered(
        uint8 indexed circuitLevel,
        uint64 indexed haltedAt,
        uint64 cooldownExpiresAt,
        uint64 auctionExpiresAt,
        uint256 indexLevelPaise,
        bytes32 triggerIndex,
        address triggeredBy
    );

    event MarketWideAuctionInitiated(
        uint8 indexed circuitLevel,
        uint64 indexed auctionStartedAt,
        uint64 auctionExpiresAt
    );

    event MarketWideResumed(
        uint8 indexed circuitLevel,
        uint64 indexed resumedAt,
        address resumedBy
    );

    event GlobalEmergencyFreezeTriggered(
        uint64 indexed frozenAt,
        bytes32 indexed reasonHash,
        address triggeredBy
    );

    event GlobalEmergencyFreezeLifted(
        uint64 indexed liftedAt,
        address liftedBy
    );

    event ISINSectorMapped(
        bytes32 indexed isin,
        bytes32 indexed sectorId
    );

    event SentinelAuthorized(
        address indexed sentinel,
        bytes32 indexed role
    );

    // --- Custom Errors ---
    error CircuitBreakerActive(bytes32 isin, HaltTier tier, uint64 expiresAt);
    error MarketWideHaltActive(uint8 level, uint64 cooldownExpiry);
    error GlobalFreezeActive();
    error CooldownNotElapsed(uint64 remainingSeconds);
    error InvalidHaltDuration(uint32 duration);
    error UnauthorizedSentinel(address caller, bytes32 requiredRole);
    error ISINAlreadyHalted(bytes32 isin);
    error ISINNotHalted(bytes32 isin);
    error SectorAlreadyHalted(bytes32 sectorId);
    error SectorNotHalted(bytes32 sectorId);
    error InvalidSessionTransition(MarketSessionState current, MarketSessionState next);

    // --- Core Verification Views (Called by Settlement & Token Contracts) ---
    function isHalted(bytes32 isin) external view returns (bool halted, HaltTier tier, uint64 expiresAt);
    function requireNotHalted(bytes32 isin) external view;
    function isGlobalFrozen() external view returns (bool);
    function getMarketWideState() external view returns (MarketWideHaltState memory);
    function getISINHaltStatus(bytes32 isin) external view returns (ISINHaltStatus memory);

    // --- Autonomous Sentinel Triggers ---
    function triggerDepositoryHalt(bytes32 isin, bytes32 mismatchProofHash) external;
    function triggerOracleDivergenceHalt(bytes32 isin, uint256 deviationBps, bytes32 detailsHash) external;
    function triggerSurveillanceVolatilityHalt(bytes32 isin, uint32 durationSeconds, bytes32 detailsHash) external;
    function triggerMarketWideCircuitBreaker(
        uint8 circuitLevel,
        bytes32 indexId,
        uint256 indexLevelPaise,
        uint64 sessionTimestamp
    ) external;

    // --- Governance & Administrative Functions ---
    function triggerGlobalEmergencyFreeze(bytes32 reasonHash) external;
    function liftGlobalEmergencyFreeze() external;
    function haltISINManual(bytes32 isin, HaltReason reason, uint32 durationSeconds, bytes32 detailsHash) external;
    function resumeISINManual(bytes32 isin) external;
    function haltSector(bytes32 sectorId, HaltReason reason, bytes32 detailsHash) external;
    function resumeSector(bytes32 sectorId) external;
    function transitionMarketWideToAuction() external;
    function resumeMarketWide() external;
    function mapISINToSector(bytes32 isin, bytes32 sectorId) external;
}
```

### Settlement Integration Pattern (`SettlementDvP.sol` snippet)
```solidity
// Integration within contracts/settlement/SettlementDvP.sol
import { ICircuitBreakerHalt } from "../interfaces/ICircuitBreakerHalt.sol";

contract SettlementDvP is ReentrancyGuardUpgradeable {
    ICircuitBreakerHalt public circuitBreaker;

    modifier whenNotHalted(bytes32 isin) {
        circuitBreaker.requireNotHalted(isin);
        _;
    }

    function settleTrade(
        TradeOrder calldata trade,
        bytes calldata orchestratorSig
    ) external nonReentrant whenNotHalted(trade.isin) {
        // Atomic DvP transfer executes only if trade.isin is not halted
        // at Tier 0 (ISIN), Tier 1 (Sector), Tier 2 (Market), or Tier 3 (Global)
    }

    function settleBatch(
        SettlementBatch calldata batch
    ) external nonReentrant {
        for (uint256 i = 0; i < batch.trades.length; i = _uncheckedInc(i)) {
            // Evaluates circuit breaker status per trade ISIN in the batch loop
            circuitBreaker.requireNotHalted(batch.trades[i].isin);
        }
        // Proceed with atomic batch execution
    }
}
```

## Security & Compliance Notes
- **SEBI Master Circular Alignment (MWCB Timing & Durations):**
  - **10% Index Drop:**
    - Triggered *before 1:00 PM*: Trading halts for 45 minutes, followed by a 15-minute call-auction.
    - Triggered *at or after 1:00 PM up to 2:30 PM*: Trading halts for 15 minutes, followed by a 15-minute call-auction.
    - Triggered *at or after 2:30 PM*: No halt; trading continues with price collars.
  - **15% Index Drop:**
    - Triggered *before 1:00 PM*: Trading halts for 1 hour 45 minutes, followed by a 15-minute call-auction.
    - Triggered *at or after 1:00 PM up to 2:00 PM*: Trading halts for 45 minutes, followed by a 15-minute call-auction.
    - Triggered *at or after 2:00 PM*: Trading halts for the remainder of the trading day.
  - **20% Index Drop:**
    - Triggered *at any time of the day*: Trading halts immediately for the remainder of the day.
- **Role-Based Emergency Triggering with M-of-N Multisig Override:**
  - *Fail-Closed Tripping:* Automated sentinels (Prompt 713, Prompt 228, Prompt 213, Prompt 328) possess individual execution keys protected by HSM/KMS (Prompt 311) to trigger halts instantaneously without awaiting governance consensus.
  - *Fail-Safe Resumption:* Resuming trading prior to statutory cooldown expiration strictly requires authorization from `MultiSigGovernance.sol` (3-of-5 threshold) passed through the OpenZeppelin `TimelockController` (48-hour delay for parameter updates, 1-hour delay for emergency override).
- **Protection Against Cascading Liquidations:** During active halts, all state-changing trade executions and margin liquidations on the affected ISINs revert immediately. Order cancellations are permitted off-chain to reduce investor exposure, while atomic on-chain asset transfers remain locked.
- **Reentrancy & Gas Optimization:** All external calls utilize CEI (Checks-Effects-Interactions) patterns and transient storage caching (`TSTORE`/`TLOAD`) to ensure that calling `requireNotHalted` inside 100-trade batch loops consumes less than 800 gas per check.

## Acceptance Criteria
- [ ] `CircuitBreakerHalt.sol` and `ICircuitBreakerHalt.sol` implemented, compiled with Solidity ^0.8.24, and passing all Slither static analysis audits with zero high or medium severity warnings.
- [ ] Tier 0 (Single-ISIN), Tier 1 (Sector), Tier 2 (SEBI Index MWCB), and Tier 3 (Global Emergency Freeze) halt hierarchy rigorously validated.
- [ ] Direct modifier `whenNotHalted(bytes32 isin)` verified in `SettlementDvP.sol` and `TokenizedEquity.sol`, reverting trade settlement with `CircuitBreakerActive` when tripped.
- [ ] SEBI MWCB 10%, 15%, and 20% timing logic correctly determines cooldown duration based on block timestamp session windows.
- [ ] Sentinel roles (`SURVEILLANCE_SENTINEL_ROLE`, `ORACLE_SENTINEL_ROLE`, `DEPOSITORY_SENTINEL_ROLE`) can trigger immediate halts, while unauthorized accounts revert with `UnauthorizedSentinel`.
- [ ] Resumption before statutory expiry strictly blocked unless executed by `GOVERNANCE_TIMELOCK_ROLE` after timelock delay.
- [ ] Automatic resumption functions correctly when temporary duration expires without requiring manual transaction for Tier 0 halts.
- [ ] Foundry test suite achieves $\ge 95\%$ line and branch test coverage, including invariant fuzz tests proving zero state inconsistency.
- [ ] Gas benchmark proves `isHalted(isin)` executes in $< 800$ gas when no halt is active.

## Suggested Order / Dependencies
- **Prerequisites:** 
  - Prompt `301` (Permissioned Blockchain Evaluation & Selection)
  - Prompt `302` (Network Topology and Validator Setup)
  - Prompt `303` (Token Issuance Smart Contract)
  - Prompt `306` (Atomic Delivery-versus-Payment Settlement Smart Contract)
  - Prompt `307` (Multi-Party Authorization & Multisig Governance Smart Contract)
  - Prompt `311` (Validator & Relayer Key Management via HSM)
- **Parallel Tasks:** 
  - Prompt `713` (Continuous Market Circuit Breaker & Resumption Coordinator)
  - Prompt `228` (Real-Time Market Surveillance Engine)
  - Prompt `328` (Decentralized Multi-Source Oracle Aggregator Smart Contract Suite)
  - Prompt `217` (Admin Back-Office Service)
- **Subsequent Prompts Enabled:** 
  - Prompt `315` (Settlement Guarantee Fund Contract)
  - Prompt `329` (NBSE Delivery-versus-Payment DvP Settlement & Automated Fee Collector)
  - Prompt `331` (Primary Market Order Routing & Clearing Bridge)
  - Prompt `811` (Testnet vs Mainnet Dual Environment CI/CD)
