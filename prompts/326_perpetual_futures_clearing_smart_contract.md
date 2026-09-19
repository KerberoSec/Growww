# 326 - On-Chain Perpetual Futures Clearing & Margin Smart Contract Suite (PerpClearingHouse.sol, PerpVault.sol, FundingRateOracle.sol)

## Purpose
Decentralized and permissioned derivatives market infrastructure requires cryptographic, non-bypassable, and transparent risk clearing and collateral custody. In volatile financial market environments, off-chain clearing mechanisms are vulnerable to settlement opacity, delayed liquidation execution, counterparty default cascades, and arbitrary risk overrides.

The **On-Chain Perpetual Futures Clearing & Margin Smart Contract Suite (`PerpClearingHouse.sol`, `PerpVault.sol`, `FundingRateOracle.sol`)** establishes an immutable, programmable settlement and risk facility deployed on the permissioned Hyperledger Besu ledger (QBFT consensus). It coordinates real-time cross-margin accounting, tokenized multi-collateral custody (e₹ CBDC, tokenized G-Secs, and approved stable reserves), automated periodic funding rate indexing, non-custodial liquidations, bad debt socialization (auto-deleveraging and mutualized loss allocation), and insurance fund backstops in full compliance with SEBI and GIFT City IFSC regulatory risk standards.

## What You Are Building
A production-grade, upgradeable Solidity smart contract suite under `contracts/derivatives/` comprising:
- `PerpClearingHouse.sol`: Core clearinghouse managing perpetual position lifecycles (Long/Short entry, size modification, close, realized PnL accounting, maintenance margin ratio checks, multi-tier partial/full liquidations, auto-deleveraging, and bad debt resolution).
- `PerpVault.sol`: Segregated cross-margin multi-collateral custody vault managing deposit registries, free margin withdraw verifications, haircut-weighted portfolio valuation, dedicated insurance fund capital custody, and protocol fee accounting.
- `FundingRateOracle.sol`: Continuous funding rate computation and cumulative index accumulator tracking time-weighted average price (TWAP) mark/index differentials, clamping limits, and O(1) funding payment calculations.
- `IPerpClearingHouse.sol`, `IPerpVault.sol`, `IFundingRateOracle.sol`: Standardized interfaces declaring all data structures, operational routines, view methods, custom errors, and events.
- Comprehensive Foundry Test Suite (`test/derivatives/PerpClearingSuite.t.sol`): Validating cross-margin invariants, multi-collateral haircut haircuts, automated funding settlement, liquidation waterfalls, insurance fund depletion, bad debt socialization, and reentrancy protections.

## Scope Boundaries
- **In Scope:**
  - On-chain cross-margin position accounting across multiple perpetual market pairs.
  - Multi-asset collateral custody in `PerpVault.sol` supporting e₹ (ERC-20 tokenized cash) and sovereign G-Sec tokens with configurable risk haircuts.
  - O(1) continuous funding payment settlements using cumulative funding rate indices.
  - Multi-tier liquidation engine with partial liquidation tranches and liquidator incentive distribution.
  - Bad debt waterfall: First drawing from defaulting user margin, then protocol Insurance Fund, and finally activating Auto-Deleveraging (ADL) / socialized loss haircut on top-tier profitable counterparties.
  - Strict mathematical invariant enforcement: Total contract collateral equals tracked user allocations plus insurance fund and accrued protocol fees.
  - Detailed audit logging emitting structured events for off-chain blockchain indexers and regulatory telemetry.
- **Out of Scope / Handled Elsewhere:**
  - Off-chain high-frequency order matching engine and order book sequencing (handled in Prompt 205).
  - Off-chain pre-trade Value-at-Risk (VaR) and SPAN margin simulations (handled in Prompt 229).
  - Physical fiat payment rail integration at Reserve Bank of India (handled in Prompt 212 / Prompt 232).
  - Primary cash-equity Delivery-versus-Payment (DvP) settlement (handled in Prompt 306).

## Technology to Use
- **Smart Contract Language:** **Solidity 0.8.24** (Target EVM: Cancun / Shanghai with Paris backward compatibility).
  *Justification:* Solidity 0.8.24 provides native checked arithmetic preventing integer overflows/underflows, custom user-defined types, transient storage opcodes (`TSTORE`/`TLOAD`) for gas-efficient reentrancy locks during batched position updates, and mature auditability.
- **Security & Upgradeability:** OpenZeppelin Contracts Upgradeable v5.0 (`Initializable`, `UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `SafeERC20`).
- **Development & Testing Framework:** **Foundry** (`forge`, `cast`) for deterministic testing, invariant property fuzzing, and gas profiling.
- **Static Analysis & Formal Verification:** Slither, Mythril, and Halmos for automated security analysis and symbolic bytecode verification.

## Backend / Infra Touchpoints
- **Derivatives Margin & Risk Engine (Prompt 206 / Prompt 229):** Computes off-chain portfolio risk snapshots and submits batched position settlements to `PerpClearingHouse.sol`.
- **Market Data & Oracle Service (Prompt 207):** Ingests multi-exchange spot and futures depth to relay signed index and mark price updates to `FundingRateOracle.sol`.
- **Settlement Guarantee & Insurance Fund Orchestrator (Prompt 230):** Manages liquidity injections, capital replenishments, and audit oversight for `PerpVault.sol`.
- **Blockchain Event Indexer (Prompt 309):** Subscribes to `PositionOpened`, `PositionClosed`, `PositionLiquidated`, `FundingPaymentSettled`, and `BadDebtSocialized` events to synchronize real-time portfolio states in PostgreSQL.

## Blockchain Interaction (permissioned Hyperledger Besu ledger with 1:1 custody backing, zero PII, QBFT)
- **Direct Asset Movement:** `PerpVault.sol` interacts with whitelisted settlement tokens (e.g., `eINRToken.sol`, `TokenizedGSec.sol`) using `SafeERC20.safeTransfer` and `SafeERC20.safeTransferFrom`.
- **Cross-Margin Portfolio Solvency:** All open positions across diverse perpetual markets share a single unified collateral balance in `PerpVault.sol`. Withdrawals are rejected if `FreeMargin < RequestedWithdrawalValue`.
- **Threshold Multi-Sig Administration:** Critical risk parameter modifications (Initial Margin Ratio, Maintenance Margin Ratio, Max Leverage, Whitelisted Collateral Haircuts) require cryptographic multi-sig authorization from `MultiSigGovernance.sol` (Prompt 307) backed by CloudHSM keys.
- **Zero On-Chain PII:** The contract suite records exclusively `bytes32 marketId`, `address trader`, `uint256 size`, `uint256 entryNotional`, and cryptographic position hashes. No personal identity data exists on-chain.

## Step-by-Step Build Instructions (10-15 steps)
1. **Scaffold Directory Structure:** Create `contracts/derivatives/PerpClearingHouse.sol`, `contracts/derivatives/PerpVault.sol`, `contracts/derivatives/FundingRateOracle.sol`, and corresponding interfaces in `contracts/interfaces/`.
2. **Define Data Structures & Enums:** Define enums (`PositionSide`, `PositionStatus`) and structs (`Position`, `MarketConfig`, `AccountMarginSummary`, `CollateralTokenConfig`, `FundingRateState`) in `contracts/interfaces/`.
3. **Configure ERC-1967 Upgradeability & Roles:** Inherit OpenZeppelin `UUPSUpgradeable` and `AccessControlUpgradeable` across all contracts, configuring storage gap buffers (`uint256[50] __gap`) for deterministic storage layouts.
4. **Implement Access Control Hierarchy:** Establish granular roles: `DEFAULT_ADMIN_ROLE`, `RISK_OPERATOR_ROLE`, `ORACLE_UPDATER_ROLE`, `LIQUIDATOR_ROLE`, and `EMERGENCY_GUARDIAN_ROLE`.
5. **Implement PerpVault Collateral Management:**
   - Define deposit and withdrawal mechanisms for whitelisted ERC-20 collateral tokens.
   - Enforce haircut valuations (e.g., e₹ at 100% value, G-Secs at 90% value, approved equities at 70% value).
   - Implement segregated accounting for user margins, protocol insurance fund capital, and accumulated liquidation/clearing fee reserves.
6. **Implement FundingRateOracle Accumulator:**
   - Implement periodic funding rate computation based on mark price and index price TWAP differentials.
   - Implement continuous cumulative funding index: $I(t) = I(t_0) + \text{Rate} \times (t - t_0)$.
   - Implement clamp limits to prevent unbounded funding spikes during extreme market volatility.
7. **Implement PerpClearingHouse Position Lifecycle:**
   - Implement `openPosition` / `closePosition` with slippage protection, leverage bounds verification, and initial margin requirement (IMR) checks.
   - Implement continuous funding fee settlement per position modification using cumulative index differences: $\text{Fee} = \text{Size} \times (I_{\text{current}} - I_{\text{entry}})$.
   - Implement real-time realized PnL calculations and balance transfers to/from `PerpVault.sol`.
8. **Implement Cross-Margin Account Summary Calculator:**
   - Build unified margin evaluation summing total collateral value, net unrealized PnL across all active positions, aggregate initial margin requirement, and maintenance margin requirement (MMR).
9. **Implement Multi-Tier Liquidation Engine:**
   - Implement `liquidatePosition` callable when account margin ratio drops below MMR.
   - Execute partial liquidations in tranches (e.g., 50% position reduction) to reduce market impact before full liquidation.
   - Deduct liquidation penalties, allocating a fixed portion to the liquidator as an execution incentive and the remainder to the Insurance Fund in `PerpVault.sol`.
10. **Implement Bad Debt Waterfall & Auto-Deleveraging (ADL):**
    - If a liquidated account has negative equity (insolvency), draw from the Insurance Fund to cover the deficit.
    - If the Insurance Fund is exhausted, trigger programmatic Auto-Deleveraging (ADL) to close opposing top-profit ranking positions at the bankruptcy price, socializing losses without protocol insolvency.
11. **Implement Emergency Pause & Risk Circuit Breakers:**
    - Add circuit-breaker controls callable by `EMERGENCY_GUARDIAN_ROLE` to pause trading per market or platform-wide during extreme anomalies.
12. **Write Comprehensive Foundry Unit Tests:**
    - Test single-collateral and multi-collateral deposits, withdrawals, and haircut adjustments.
    - Test long and short position openings, incremental additions, partial closes, and PnL realization.
    - Test continuous funding payments under positive and negative funding rate regimes.
13. **Write Liquidation & Invariant Fuzz Tests:**
    - Test partial and full liquidations across extreme price drop scenarios.
    - Test bad debt resolution with insurance fund drawdowns and multi-party auto-deleveraging.
    - Fuzz test the invariant: $\sum \text{User Account Equities} + \text{Insurance Fund} + \text{Fee Reserves} == \text{Total Vault Collateral}$.
14. **Perform Static Analysis & Formal Verification:**
    - Execute Slither and Mythril to verify absence of reentrancy vectors, arbitrary external calls, or rounding vulnerabilities.
15. **Generate Deployment & Verification Scripts:**
    - Write Foundry deployment scripts (`script/DeployPerpSuite.s.sol`) configured for Hyperledger Besu QBFT network and output machine-readable contract ABIs.

## Interfaces / Contracts

### Perpetual Clearing House Interface (`IPerpClearingHouse.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IPerpClearingHouse {
    enum PositionSide {
        LONG,
        SHORT
    }

    enum PositionStatus {
        OPEN,
        CLOSED,
        LIQUIDATED
    }

    struct Position {
        bytes32 marketId;
        address trader;
        PositionSide side;
        uint256 size;
        uint256 entryNotional;
        int256 entryFundingIndex;
        uint256 lastUpdatedTimestamp;
    }

    struct MarketConfig {
        bytes32 marketId;
        address baseToken;
        uint256 initialMarginRatioBps;
        uint256 maintenanceMarginRatioBps;
        uint256 liquidationPenaltyBps;
        uint256 insuranceFeeBps;
        uint256 maxLeverage;
        bool isTradingEnabled;
    }

    struct AccountMarginSummary {
        uint256 totalCollateralValueUSD;
        int256 totalUnrealizedPnlUSD;
        uint256 totalInitialMarginRequiredUSD;
        uint256 totalMaintenanceMarginRequiredUSD;
        int256 freeMarginUSD;
        uint256 marginRatioBps;
    }

    struct PositionModificationParams {
        bytes32 marketId;
        address trader;
        PositionSide side;
        uint256 sizeDelta;
        uint256 priceLimit;
        bool isClose;
    }

    // Events
    event PositionOpened(
        bytes32 indexed marketId,
        address indexed trader,
        PositionSide side,
        uint256 size,
        uint256 entryPrice
    );
    event PositionClosed(
        bytes32 indexed marketId,
        address indexed trader,
        PositionSide side,
        uint256 size,
        uint256 exitPrice,
        int256 realizedPnl
    );
    event PositionLiquidated(
        bytes32 indexed marketId,
        address indexed trader,
        address indexed liquidator,
        uint256 liquidatedSize,
        uint256 liquidationPrice,
        uint256 penaltyAmount,
        uint256 insuranceFee
    );
    event FundingPaymentSettled(
        bytes32 indexed marketId,
        address indexed trader,
        int256 fundingPayment,
        int256 currentFundingIndex
    );
    event BadDebtSocialized(
        bytes32 indexed marketId,
        address indexed defaulter,
        uint256 badDebtAmount,
        uint256 insuranceDrawn,
        uint256 socializedAmount
    );
    event AutoDeleveraged(
        bytes32 indexed marketId,
        address indexed targetTrader,
        uint256 closedSize,
        uint256 executionPrice,
        int256 realizedPnl
    );
    event MarketConfigUpdated(
        bytes32 indexed marketId,
        uint256 initialMarginRatioBps,
        uint256 maintenanceMarginRatioBps,
        bool isTradingEnabled
    );

    // Custom Errors
    error UnauthorizedCaller(address caller);
    error MarketNotActive(bytes32 marketId);
    error InsufficientCollateral(uint256 required, uint256 available);
    error PositionExceedsMaxLeverage(uint256 requestedLeverage, uint256 maxLeverage);
    error AccountNotLiquidatable(address trader, uint256 marginRatioBps, uint256 mmrBps);
    error SlippageToleranceExceeded(uint256 expectedPrice, uint256 actualPrice);
    error ZeroPositionSize();
    error ZeroAddressNotAllowed();
    error InsuranceFundDepleted();

    // Operational Functions
    function openPosition(PositionModificationParams calldata params) external returns (uint256 executedPrice, int256 realizedPnl);
    function closePosition(PositionModificationParams calldata params) external returns (uint256 executedPrice, int256 realizedPnl);
    function liquidatePosition(bytes32 marketId, address trader, uint256 maxLiquidateSize) external returns (uint256 liquidatedSize, uint256 penaltyAmount);
    function settleFunding(bytes32 marketId, address trader) external returns (int256 fundingPayment);
    function socializeBadDebt(bytes32 marketId, address defaulter, uint256 badDebtAmount) external;

    // View Functions
    function getAccountMarginSummary(address trader) external view returns (AccountMarginSummary memory);
    function getPosition(bytes32 marketId, address trader) external view returns (Position memory);
    function getMarketConfig(bytes32 marketId) external view returns (MarketConfig memory);
    function getUnrealizedPnl(bytes32 marketId, address trader) external view returns (int256 unrealizedPnl);
}
```

### Perpetual Vault Interface (`IPerpVault.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IPerpVault {
    struct CollateralTokenConfig {
        address tokenAddress;
        uint256 haircutBps;
        uint256 depositCap;
        uint256 totalDeposited;
        bool isAllowed;
    }

    struct UserCollateral {
        address tokenAddress;
        uint256 rawAmount;
        uint256 valueUSD;
    }

    // Events
    event CollateralDeposited(
        address indexed trader,
        address indexed token,
        uint256 amount,
        uint256 valuationUSD
    );
    event CollateralWithdrawn(
        address indexed trader,
        address indexed token,
        uint256 amount,
        address indexed recipient
    );
    event InsuranceFundCapitalDeposited(address indexed token, uint256 amount);
    event InsuranceFundDrawn(address indexed token, uint256 amount, address indexed destination);
    event ProtocolFeesSwept(address indexed token, uint256 amount, address indexed treasury);
    event CollateralConfigUpdated(address indexed token, uint256 haircutBps, uint256 depositCap, bool isAllowed);

    // Custom Errors
    error TokenNotSupported(address token);
    error DepositCapExceeded(address token, uint256 currentDeposit, uint256 attemptedDeposit);
    error InsufficientFreeMargin(address trader, int256 freeMargin, uint256 withdrawValueUSD);
    error TransferFailed(address token, address from, address to, uint256 amount);
    error VaultBalanceMismatch(address token, uint256 expected, uint256 actual);
    error CallerNotClearingHouse(address caller);

    // Operational Functions
    function depositCollateral(address token, uint256 amount) external;
    function withdrawCollateral(address token, uint256 amount, address recipient) external;
    function depositInsuranceCapital(address token, uint256 amount) external;
    function drawInsuranceFund(address token, uint256 amount, address destination) external returns (uint256 actualDrawn);
    function transferCollateralToClearingHouse(address token, uint256 amount, address recipient) external;

    // View Functions
    function getTraderCollateralBalance(address trader, address token) external view returns (uint256 rawBalance);
    function getTotalCollateralValueUSD(address trader) external view returns (uint256 totalValueUSD);
    function getInsuranceFundBalance(address token) external view returns (uint256 balance);
    function getCollateralTokenConfig(address token) external view returns (CollateralTokenConfig memory);
}
```

### Funding Rate Oracle Interface (`IFundingRateOracle.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IFundingRateOracle {
    struct FundingRateState {
        bytes32 marketId;
        int256 currentRateBps;
        int256 cumulativeFundingIndex;
        uint256 lastUpdatedTimestamp;
        uint256 fundingIntervalSeconds;
    }

    struct OraclePriceData {
        bytes32 marketId;
        uint256 markPrice;
        uint256 indexPrice;
        uint256 timestamp;
        uint256 confidence;
    }

    // Events
    event FundingRateUpdated(
        bytes32 indexed marketId,
        int256 fundingRateBps,
        int256 cumulativeFundingIndex,
        uint256 markPrice,
        uint256 indexPrice
    );
    event PriceFeedsUpdated(
        bytes32 indexed marketId,
        uint256 markPrice,
        uint256 indexPrice,
        uint256 timestamp
    );
    event FundingParameterConfigured(
        bytes32 indexed marketId,
        uint256 fundingIntervalSeconds,
        int256 maxFundingRateBps,
        uint256 dampenerBps
    );

    // Custom Errors
    error StalePriceFeed(bytes32 marketId, uint256 feedTimestamp, uint256 currentTimestamp);
    error PriceDeviationTooHigh(bytes32 marketId, uint256 markPrice, uint256 indexPrice, uint256 maxAllowedDeviationBps);
    error FundingIntervalNotElapsed(bytes32 marketId, uint256 nextFundingTime, uint256 currentTime);
    error OracleSignerInvalid(address recoveredSigner, address expectedSigner);

    // Operational Functions
    function updateFundingRate(bytes32 marketId) external returns (int256 newRateBps, int256 newCumulativeFundingIndex);
    function submitPriceData(
        bytes32 marketId,
        uint256 markPrice,
        uint256 indexPrice,
        uint256 timestamp,
        bytes calldata signature
    ) external;

    // View Functions
    function getLatestMarkPrice(bytes32 marketId) external view returns (uint256 markPrice, uint256 timestamp);
    function getLatestIndexPrice(bytes32 marketId) external view returns (uint256 indexPrice, uint256 timestamp);
    function getFundingRateState(bytes32 marketId) external view returns (FundingRateState memory);
    function getFundingPayment(
        bytes32 marketId,
        int256 entryFundingIndex,
        uint256 positionSize,
        bool isLong
    ) external view returns (int256 fundingPayment);
}
```

## Security & Compliance Notes
- **SEBI & IFSCA Derivatives Margin Regulations:** The smart contract suite enforces continuous Maintenance Margin Requirement (MMR) checks. Positions with margin ratios breaching statutory minimums are queued for immediate liquidation without grace periods.
- **Fail-Safe Bad Debt Resolution:** When liquidations occur at prices worse than the bankruptcy price, deficits are absorbed strictly in sequence: (1) Defaulting user collateral, (2) Protocol Insurance Fund, and (3) Auto-Deleveraging (ADL) across counterparty positions, preventing protocol-wide insolvencies.
- **Reentrancy Protection & Checks-Effects-Interactions:** All state-modifying methods in `PerpClearingHouse.sol` and `PerpVault.sol` utilize OpenZeppelin `ReentrancyGuardUpgradeable` and follow strict Checks-Effects-Interactions sequencing to guard against reentrancy vectors during cross-contract token transfers.
- **Multi-Oracle Freshness & Manipulation Guards:** `FundingRateOracle.sol` verifies price timestamp freshness and enforces maximum allowable mark-to-index price deviation thresholds, reverting price submissions during flash loan oracle attacks.
- **Institutional Multi-Sig Timelock Controls:** Changes to market margin ratios, maximum leverage, liquidation penalty splits, or collateral haircut tables require 48-hour timelock execution via `MultiSigGovernance.sol` (Prompt 307).

## Acceptance Criteria
- [ ] `PerpClearingHouse.sol`, `PerpVault.sol`, and `FundingRateOracle.sol` deployed and verified on Hyperledger Besu local devnet.
- [ ] Cross-margin account balances accurately track multi-collateral deposits with appropriate risk haircuts applied.
- [ ] Position opening and size modification correctly enforce Initial Margin Requirements and leverage limits.
- [ ] Continuous funding payments accurately settle across open positions using cumulative funding rate indices.
- [ ] Liquidation engine triggers upon margin ratio falling below MMR, distributing penalty incentives to liquidators and the insurance fund.
- [ ] Bad debt waterfall correctly absorbs negative account balances via the Insurance Fund, transitioning to Auto-Deleveraging if the fund is depleted.
- [ ] Invariant fuzz tests pass 10,000 runs in Foundry verifying zero token balance leakage across the vault.
- [ ] Slither and Mythril security scans complete with zero high or medium severity vulnerabilities.

## Suggested Order / Dependencies
- **Prerequisites:** Prompt `301` (Blockchain Selection), Prompt `303` (Token Issuance), Prompt `306` (Settlement DvP Contract), Prompt `307` (MultiSig Governance).
- **Parallel Tasks:** Prompt `206` (Pre-Trade Risk Engine), Prompt `229` (Real-Time VaR Margin Engine), Prompt `230` (SGF & Insurance Fund Service).
- **Subsequent Prompts Enabled:** Prompt `215` (Reconciliation Service), Prompt `308` (Proof of Reserve Publishing), Prompt `309` (Event Indexing Service).
