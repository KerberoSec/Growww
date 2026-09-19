# 339 - Concentrated Liquidity Market Maker (CLMM) Off-Hours AMM Contract (Solidity)

## Purpose
Primary Indian equity exchanges (National Stock Exchange - NSE, and Bombay Stock Exchange - BSE) execute cash and derivatives market trading strictly between 09:15 and 15:30 Indian Standard Time (IST). Outside these statutory regular trading hours (RTH) - encompassing evening hours, overnight windows, weekends, and exchange holidays - domestic and international investors face a complete liquidity vacuum. Institutional asset managers, retail traders, and global participants in Gujarat International Finance Tec-City (GIFT City IFSC) are left exposed to overnight macro events, corporate earnings announcements, geopolitical developments, and global market shocks without an orderly mechanism to manage risk or adjust equity exposure.

Traditional off-market mechanisms, including over-the-counter (OTC) bilateral matching and standard constant product automated market makers ($x \cdot y = k$), suffer from fatal structural defects:
- **Capital Inefficiency:** Constant product AMMs distribute capital uniformly across the price spectrum from zero to infinity ($[0, \infty)$), requiring massive collateral reserves to provide shallow depth around the current market price, resulting in prohibitive slippage on medium-to-large equity swap orders.
- **Unbounded Price Drift & Manipulation:** Without rigid anchor constraints, unpegged off-hours AMMs can be manipulated via low-liquidity spoofing, flash loans, or cross-venue arbitrage, leading to irrational opening imbalances when primary exchanges reopen.
- **Disconnection from Statutory Exchange Rules:** SEBI guidelines mandate strict circuit breaker collars (e.g. +/- 5%, +/- 10%) relative to previous official closing prices to maintain orderly markets and prevent predatory cascades.

The **Concentrated Liquidity Market Maker (CLMM) Off-Hours AMM Contract (`OffHoursCLMM.sol`)** solves these challenges by establishing an institutional-grade, capital-efficient, on-chain liquidity venue for tokenized equities deployed on the permissioned Hyperledger Besu ledger (QBFT consensus). By adopting concentrated liquidity mechanics ($L = \frac{\Delta y}{\Delta \sqrt{P}}$), liquidity providers (LPs) concentrate capital within discrete, customized price ranges around the primary exchange official closing price. The contract integrates dynamic volatility collars strictly bounded by primary exchange closing prices (default +/- 5% hard collar), geometric mean Time-Weighted Average Price (TWAP) oracles, time-gated trading sessions (15:30 to 09:15 IST), and dynamic fee scaling to deliver institutional depth, front-running resistance, and full regulatory harmony.

---

## What You Are Building
A production-grade, upgradeable Solidity smart contract architecture under `contracts/src/liquidity/` comprising:
- **`contracts/src/liquidity/OffHoursCLMM.sol`**: Core concentrated liquidity exchange contract managing tick-based liquidity allocations, virtual reserves, stateful price tick navigation via bitmapped indexes, single-hop and multi-hop atomic swap execution, dynamic fee scaling, and hard dynamic volatility collar enforcement bounded by primary closing prices.
- **`contracts/src/liquidity/CLMMPositionNFT.sol`**: Non-fungible token (ERC-721) position manager tracking concentrated liquidity positions, fee growth inside tick boundaries ($feeGrowthInside0X128, feeGrowthInside1X128$), accrued uncollected fees, and LP position lifecycle operations (mint, add liquidity, remove liquidity, collect).
- **`contracts/src/liquidity/libraries/TickMath.sol`**: High-performance mathematical library computing square root price at tick index $\sqrt{P}(i) = 1.0001^{i/2}$ in Q64.96 fixed-point representation and reverse log base 1.0001 calculations.
- **`contracts/src/liquidity/libraries/FullMath.sol`**: 512-bit intermediate product arithmetic library facilitating $(a \times b) / denominator$ with round-up and round-down modes without overflow.
- **`contracts/src/liquidity/libraries/SqrtPriceMath.sol` & `SwapMath.sol`**: Mathematical functions calculating token deltas ($\Delta x$, $\Delta y$) and step-by-step price transitions within tick intervals.
- **`contracts/src/interfaces/liquidity/ICLMMExchange.sol`**: Primary interface declaring swap methods, slot0 state structures, volatility collar parameters, dynamic fee models, TWAP oracle observations, and administrative controls.
- **`contracts/src/interfaces/liquidity/ISyntheticLiquidityPool.sol`**: Pool interface governing concentrated liquidity provision, fee growth accounting, position tracking, and synthetic asset pairing routines.
- **Foundry Test Suite (`test/liquidity/OffHoursCLMM.t.sol`)**: Comprehensive unit, differential, invariant fuzz, and boundary tests validating tick transitions, collar clamping, fee scaling, TWAP accuracy, and reentrancy resistance.

---

## Scope Boundaries
- **In-Scope:**
  - Implementation of concentrated liquidity mathematical engine using Q64.96 fixed-point arithmetic and FullMath intermediate 512-bit operations.
  - Tick management with spaced tick bitmaps (`TickBitmap.sol`) enabling efficient tick initialization queries and boundary crossings.
  - Time-gated session enforcement: swaps and active liquidity rebalancing permitted exclusively during off-hours (15:30:00 to 09:14:59 IST on exchange trading days; continuous operation on weekends and exchange holidays).
  - Dynamic volatility collar enforcement: rejecting swaps or clamping execution prices that breach the configured collar band (default +/- 5% relative to the primary closing price $P_{close}$).
  - Ingestion and validation of primary exchange official closing prices signed by authorized oracles.
  - Geometric mean TWAP oracle observation ring-buffer (minimum 60 historical observations recording timestamp, tick cumulative, and liquidity cumulative).
  - Dynamic fee scaling: base pool fee dynamically augmented based on instantaneous volatility, collar proximity, and block-level trade intensity.
  - ERC-721 tokenized position tracking for fractional concentrated liquidity providers.
  - Emergency halt hooks integrated with platform-wide `CircuitBreakerTimelock.sol` (Prompt 341).
  - Event emissions compatible with blockchain indexing services (`ChainEventIndexer`, Prompt 309).
- **Out-of-Scope / Handled Elsewhere:**
  - Primary exchange continuous limit order book (CLOB) matching during regular trading hours 09:15 to 15:30 IST (handled in Prompt 205).
  - Depository demat immobilisation and physical stock custody (handled in Custodian Depository Integration Service, Prompt 213).
  - Direct ingestion and cleansing of raw multicast exchange feeds (handled in Market Data Service, Prompt 207).
  - Multi-asset cross-margin portfolio accounting and perpetual futures liquidation engines (handled in Prompt 326).
  - Off-chain clearinghouse tax withholding and eTDS generation (handled in Prompt 336).

---

## Technology to Use
- **Smart Contract Language:** **Solidity ^0.8.24** (Target EVM: Cancun / Shanghai with Yul inline assembly for optimized FullMath and BitMap manipulation).
  *Justification:* Solidity 0.8.24 provides native checked arithmetic, user-defined value types, and supports transient storage opcodes (`TSTORE` / `TLOAD`) for gas-efficient reentrancy locks during batched swap and tick-crossing loops.
- **Fixed-Point Arithmetic:** **Q64.96** (unsigned fixed-point number with 64 bits of integer precision and 96 bits of fractional precision, where $1.0 = 2^{96} \approx 7.9228 \times 10^{28}$).
- **Math Libraries:** Custom audited implementations of `FullMath.sol` (512-bit multiplication and division), `TickMath.sol` (tick to $\sqrt{P}$ conversions), and `FixedPoint96.sol`.
- **Framework & Tooling:** **Foundry (`forge`, `cast`)** for unit testing, differential testing against reference implementations, invariant fuzzing, and gas profiling.
- **Base Standards:** OpenZeppelin Contracts Upgradeable v5.0 (`UUPSUpgradeable`, `AccessControlUpgradeable`, `ReentrancyGuardUpgradeable`, `PausableUpgradeable`, `ERC721EnumerableUpgradeable`, `SafeERC20`).
- **Target Ledger:** Hyperledger Besu permissioned EVM ledger with QBFT consensus, 2-second block finality, and zero gas volatility.

---

## Backend / Infra Touchpoints
- **Market Data Service (Prompt 207):** Ingests official closing prices from NSE and BSE at 15:30 IST, computes volume-weighted benchmarks, signs the closing price payload with an authorized relayer key, and submits the data to update $P_{close}$ and reset volatility collars on-chain.
- **Primary Market Adapter (Prompt 242):** Signals session transitions (Market Close at 15:30 IST, Pre-Market Open at 09:00 IST, Regular Market Open at 09:15 IST), corporate action tick adjustments (e.g. stock splits, bonus issues, dividends), and exchange holiday calendars.
- **Fee and Realized PnL Engine (Prompt 210):** Subscribes to on-chain `Swap` and `Collect` events, aggregates protocol fee shares, calculates liquidity provider yields, and synchronizes real-time off-chain trading statements and tax ledgers.
- **Chain Event Indexer (Prompt 309):** Indexes `Swap`, `Mint`, `Burn`, `CollarUpdated`, and `SessionStateChanged` events to generate off-hours order book depth, candlestick charts, and TWAP telemetry for the Growww web/mobile trading interfaces.
- **Circuit Breaker Timelock (Prompt 341):** Listens to cross-market anomaly alerts; possesses administrative authority to trigger emergency pause hooks if anomalous off-hours price action is detected.

---

## Blockchain Interaction
- **Permissioned Ledger Execution:** Deployed on Hyperledger Besu under QBFT consensus. Transactions are processed by authorized consortium validators with 2-second block intervals, completely eliminating public mempool front-running and miner-extractable value (MEV) sandwich attacks by third parties.
- **Atomic Swap Routine:** Swaps execute via `swap(address recipient, bool zeroForOne, int256 amountSpecified, uint160 sqrtPriceLimitX96, bytes calldata data)`:
  - Exact input ($amountSpecified > 0$) or exact output ($amountSpecified < 0$).
  - Strict slippage limit ($sqrtPriceLimitX96$) ensuring the transaction reverts if market impact exceeds user tolerance.
  - Callback pattern allowing flash-swaps or custom collateral transfers prior to pool settlement.
- **Position NFT Management:** Liquidity providers deposit tokenized equity (Token0) and settlement currency (eINR / USDC, Token1). The position manager mints an ERC-721 NFT recording:
  - `tickLower` and `tickUpper` (multiples of `tickSpacing`).
  - Liquidity amount $L = \Delta y / \Delta \sqrt{P}$.
  - Fee growth snapshots for deterministic fee harvesting without historical iteration.
- **Dynamic Circuit Breaker & Collar Hooks:** On every swap step:
  - The proposed target $\sqrt{P}_{next}$ is checked against $\sqrt{P}_{collar\_min}$ and $\sqrt{P}_{collar\_max}$.
  - If the price breaches the +/- 5% collar, the swap is either capped at the collar boundary tick or reverts with custom error `CollarBreached(uint160 targetPrice, uint160 limitPrice)`.
- **Trading Window Gating:** Every state-modifying swap or liquidity modification verifies:
  - `isOffHoursActive == true`.
  - Block timestamp falls within the statutory off-hours window (`15:30 <= time < 09:15 IST` or `isExchangeHoliday == true`).
  - If invoked during primary market hours (09:15 to 15:30 IST), the transaction reverts with `OutsideTradingHours()`.
- **Zero On-Chain PII:** The contract stores solely token contract addresses, integer tick indexes, Q64.96 price coordinates, and anonymized account addresses. No investor personal identity or Demat account numbers reside on-chain.

---

## Step-by-Step Build Instructions (10-15 steps)

1. **Scaffold Directory Structure & Dependencies:**
   - Create directories: `contracts/src/liquidity/`, `contracts/src/liquidity/libraries/`, `contracts/src/interfaces/liquidity/`, `test/liquidity/`, and `script/liquidity/`.
   - Configure `foundry.toml` targeting EVM `cancun`, optimizer enabled with 20,000 runs, and Yul compilation enabled.
   - Install OpenZeppelin Contracts Upgradeable v5.0.

2. **Implement Core Arithmetic Libraries:**
   - Implement `contracts/src/liquidity/libraries/FullMath.sol`: 512-bit intermediate multiplication ($a \times b$) using assembly, modular inverse computation, and exact division with rounding control.
   - Implement `contracts/src/liquidity/libraries/TickMath.sol`: Compute $\sqrt{P}$ from tick index $i \in [-887272, 887272]$ using binary exponentiation in Q64.96 format; compute tick index from $\sqrt{P}$ with ratio approximations.
   - Implement `contracts/src/liquidity/libraries/SqrtPriceMath.sol`: Compute $\Delta x = \Delta (1/\sqrt{P}) \cdot L$ and $\Delta y = \Delta \sqrt{P} \cdot L$ with round-up for inputs and round-down for outputs.
   - Implement `contracts/src/liquidity/libraries/SwapMath.sol`: Compute step-by-step swap amounts, next square root price, and fee deductions within a single tick interval.

3. **Implement Tick and Bitmap Management Libraries:**
   - Implement `contracts/src/liquidity/libraries/TickBitmap.sol`: Efficiently locate initialized ticks using 256-bit word bitmasks and bitwise operations (`mostSignificantBit`, `leastSignificantBit`).
   - Implement tick initialization, reference count tracking, and fee growth outside accounting in `contracts/src/liquidity/libraries/Tick.sol`.

4. **Define Formal Interfaces:**
   - Create `contracts/src/interfaces/liquidity/ICLMMExchange.sol` declaring structs (`Slot0`, `TickInfo`, `PositionInfo`, `CollarConfig`, `FeeConfig`), swap routines, view functions, custom errors, and events.
   - Create `contracts/src/interfaces/liquidity/ISyntheticLiquidityPool.sol` declaring LP liquidity management (`mint`, `burn`, `collect`), fee collection, and synthetic equity token pairing methods.
   - Create `contracts/src/interfaces/liquidity/ICLMMPositionNFT.sol` declaring ERC-721 tokenized position descriptor and transfer methods.

5. **Implement Dynamic Volatility Collar Module:**
   - In `OffHoursCLMM.sol`, declare `CollarConfig` struct tracking: `primaryClosingPriceX96`, `collarBandBps` (default 500 = 5.00%), `minCollarPriceX96`, `maxCollarPriceX96`, `minCollarTick`, `maxCollarTick`, and `lastOracleUpdateTimestamp`.
   - Implement `setPrimaryClosingPrice(uint160 closingPriceX96, uint256 timestamp, bytes calldata signature)` restricted to `ORACLE_UPDATER_ROLE` (Prompt 207).
   - Recompute `minCollarPriceX96 = closingPriceX96 * (10000 - collarBandBps) / 10000` and `maxCollarPriceX96 = closingPriceX96 * (10000 + collarBandBps) / 10000` and cache their corresponding tick boundaries.

6. **Implement Time-Gated Trading Window Engine:**
   - Implement `checkOffHoursActive()` view helper converting block timestamp to Indian Standard Time (IST = UTC + 5:30).
   - Enforce active window: daily from 15:30:00 to 09:14:59 IST.
   - Implement `setTradingSessionOverride(bool forceActive, bool isHoliday)` restricted to `MARKET_OPERATOR_ROLE` (Prompt 242) for handling scheduled exchange holidays and early market halts.
   - Attach `onlyOffHours` modifier to `swap` and liquidity modification functions.

7. **Implement Dynamic Fee Scaling Algorithm:**
   - Define base fee: e.g. 15 bps ($0.15\%$).
   - Implement dynamic fee multiplier based on two factors:
     1. **Collar Proximity:** As current tick nears `minCollarTick` or `maxCollarTick` within a 10% buffer zone of the collar boundary, linearly scale fee up to maximum fee cap (e.g. 100 bps).
     2. **Tick Velocity:** Track aggregate tick traversal within the current block or across the last 3 blocks; escalate fee if rapid one-sided price movement is detected.
   - Calculate effective swap fee dynamically prior to each swap execution step.

8. **Implement Core Swap Execution (`swap`):**
   - Implement `swap` method in `OffHoursCLMM.sol` with strict reentrancy protection.
   - Validate trading session active and inputs non-zero.
   - Initialize swap state: remaining amount, current $\sqrt{P}$, current tick, fee growth accumulators.
   - Execute while loop across ticks:
     - Find next initialized tick using `TickBitmap`.
     - Ensure next tick does not exceed `minCollarTick` or `maxCollarTick`.
     - Calculate step swap using `SwapMath.computeSwapStep`.
     - Update fee growth and deduct dynamic swap fee.
     - Cross tick if target reached: update tick net liquidity and cross fee growth outside.
   - Enforce final price within user's `sqrtPriceLimitX96` and within dynamic volatility collar.
   - Execute token balance settlements via `SafeERC20` transfers and trigger caller swap callback if requested.
   - Emit `Swap` event with execution details.

9. **Implement Liquidity Provision Lifecycle (`mint`, `burn`, `collect`):**
   - Implement `mint`: Validate lower and upper ticks are valid multiples of `tickSpacing` and `tickLower < tickUpper`. Calculate required token amounts based on current $\sqrt{P}$, transfer tokens to pool, update tick liquidity, and update position state.
   - Implement `burn`: Reduce liquidity from position, update tick net liquidity, transition accrued tokens to `tokensOwed0` and `tokensOwed1`.
   - Implement `collect`: Transfer accrued fees and withdrawn principal to the position owner.
   - Enforce protocol fee split: Deduct protocol fee share (e.g. 10% of collected swap fees) and credit to protocol fee vault.

10. **Implement ERC-721 Tokenized Position Manager (`CLMMPositionNFT.sol`):**
    - Implement ERC-721 compliant position manager contract with metadata descriptor.
    - Expose `mintPosition`, `increaseLiquidity`, `decreaseLiquidity`, and `collectFees` delegating directly to underlying `OffHoursCLMM` pools.
    - Restrict burning and fee collection strictly to the authenticated NFT owner or approved operator.

11. **Implement Geometric Mean TWAP Oracle Ring-Buffer:**
    - Maintain fixed-size circular observation buffer of at least 60 entries tracking: `blockTimestamp`, `tickCumulative`, and `secondsPerLiquidityCumulativeX128`.
    - Write observation on the first swap of each block.
    - Expose `observe(uint32[] calldata secondsAgos)` returning time-weighted tick cumulatives for manipulation-resistant off-chain pricing.

12. **Implement Emergency Controls & Circuit Breakers:**
    - Inherit OpenZeppelin `PausableUpgradeable` and `AccessControlUpgradeable`.
    - Grant `EMERGENCY_GUARDIAN_ROLE` to `CircuitBreakerTimelock.sol` (Prompt 341) to trigger immediate global pool pause.
    - Implement `emergencyWithdraw` with timelock protection in the event of upstream token contract deprecation.

13. **Write Comprehensive Foundry Unit & Fuzz Tests (`test/liquidity/OffHoursCLMM.t.sol`):**
    - Unit test tick conversion accuracy and math invariants against Uniswap v3 reference vectors.
    - Unit test off-hours trading window validation: verify swaps succeed at 16:00 IST and fail at 10:00 IST.
    - Invariant fuzz test: ensure total pool token balances always equal or exceed virtual reserve requirements plus uncollected fees ($\text{Balance}_0 \ge \text{Tracked}_0$, $\text{Balance}_1 \ge \text{Tracked}_1$).
    - Collar breach test: verify swap attempting to move price $+5.01\%$ reverts with `CollarBreached`.
    - Dynamic fee test: verify fees increase monotonically as price approaches collar boundaries.

14. **Conduct Gas Profiling & Formal Auditing:**
    - Benchmark gas costs for standard swaps (target $\le 110,000$ gas for single-tick crossing swaps).
    - Run Slither static analyzer to verify absence of reentrancy, uninitialized storage pointers, or precision truncation.
    - Perform symbolic verification with Halmos asserting collar bounds cannot be breached under any sequence of valid swaps.

15. **Prepare Deployment Scripts & Documentation:**
    - Write `script/liquidity/DeployOffHoursCLMM.s.sol` configuring initial tokens, fee tiers, tick spacings, and oracle authorizations for Hyperledger Besu QBFT devnet and mainnet.
    - Publish contract interface documentation, operational runbooks for daily closing price updates, and disaster recovery procedures.

---

## Interfaces / Contracts

### Concentrated Liquidity Exchange Interface (`ICLMMExchange.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title ICLMMExchange
 * @notice Primary interface for Concentrated Liquidity Market Maker Off-Hours AMM
 * @dev Enforces tick-based concentrated liquidity, dynamic volatility collars, and time-gated sessions
 */
interface ICLMMExchange {

    // --- STRUCTS ---

    struct Slot0 {
        uint160 sqrtPriceX96;             // Current square root price as a Q64.96
        int24 tick;                        // Current tick index corresponding to sqrtPriceX96
        uint16 observationIndex;           // Index of the most recently written oracle observation
        uint16 observationCardinality;     // Current maximum number of observations stored
        uint16 observationCardinalityNext; // Next maximum number of observations to be allocated
        uint24 feeBps;                     // Current dynamic swap fee in hundredths of a bip (1e-6)
        bool unlocked;                     // Reentrancy lock status
    }

    struct CollarConfig {
        uint160 primaryClosingPriceX96;    // Official closing price from primary exchange (Q64.96)
        uint16 collarBandBps;              // Allowed price collar deviation in basis points (e.g. 500 = 5%)
        uint160 minCollarPriceX96;         // Lower bound price collar (P_close * (1 - band))
        uint160 maxCollarPriceX96;         // Upper bound price collar (P_close * (1 + band))
        int24 minCollarTick;               // Tick corresponding to minCollarPriceX96
        int24 maxCollarTick;               // Tick corresponding to maxCollarPriceX96
        uint64 lastOracleUpdateTimestamp;  // Timestamp of last primary closing price update
        bool isCollarEnforced;             // Circuit breaker active flag
    }

    struct FeeConfig {
        uint24 baseFeeBps;                 // Base pool fee (e.g. 1500 = 15 bps)
        uint24 maxFeeBps;                  // Maximum dynamic fee cap (e.g. 10000 = 100 bps)
        uint16 collarProximityBufferTicks; // Tick range adjacent to collar where fee scaling triggers
        uint16 velocityScaleFactor;        // Sensitivity multiplier for intraday tick velocity
    }

    struct TickInfo {
        uint128 liquidityGross;            // Total liquidity referencing this tick
        int128 liquidityNet;               // Liquidity to add (lower tick) or subtract (upper tick) on cross
        uint256 feeGrowthOutside0X128;     // Fee growth on the other side of this tick for token0
        uint256 feeGrowthOutside1X128;     // Fee growth on the other side of this tick for token1
        int56 tickCumulativeOutside;       // Cumulative tick value on the other side of this tick
        uint160 secondsPerLiquidityOutsideX128; // Seconds per unit of liquidity on the other side
        uint32 secondsOutside;             // Seconds spent on the other side of this tick
        bool initialized;                  // True if liquidityGross > 0
    }

    struct OracleObservation {
        uint32 blockTimestamp;             // Observation timestamp
        int56 tickCumulative;              // Cumulative tick value (sum of tick * seconds)
        uint160 secondsPerLiquidityCumulativeX128; // Cumulative seconds per liquidity
        bool initialized;                  // True if observation has been written
    }

    struct SwapParams {
        address recipient;                 // Address receiving swapped output tokens
        bool zeroForOne;                   // True if trading token0 for token1, false if token1 for token0
        int256 amountSpecified;            // Exact input (> 0) or exact output (< 0) amount
        uint160 sqrtPriceLimitX96;         // Execution price limit protecting against slippage
        bytes data;                        // Optional callback data passed to swap caller
    }

    // --- EVENTS ---

    event Swap(
        address indexed sender,
        address indexed recipient,
        int256 amount0,
        int256 amount1,
        uint160 sqrtPriceX96,
        uint128 liquidity,
        int24 tick,
        uint24 effectiveFeeBps
    );

    event PrimaryClosingPriceUpdated(
        uint160 oldClosingPriceX96,
        uint160 newClosingPriceX96,
        uint160 minCollarPriceX96,
        uint160 maxCollarPriceX96,
        int24 minCollarTick,
        int24 maxCollarTick,
        uint256 timestamp
    );

    event TradingSessionStateChanged(
        bool isOffHoursActive,
        bool isManualOverride,
        uint256 timestamp
    );

    event DynamicFeeConfigUpdated(
        uint24 baseFeeBps,
        uint24 maxFeeBps,
        uint16 collarProximityBufferTicks,
        uint16 velocityScaleFactor
    );

    event CollarBandUpdated(uint16 oldBandBps, uint16 newBandBps);
    event ProtocolFeeCollected(address indexed recipient, uint256 amount0, uint256 amount1);

    // --- CUSTOM ERRORS ---

    error OutsideTradingHours();
    error CollarBreached(uint160 targetPriceX96, uint160 boundaryPriceX96);
    error InvalidTickRange(int24 tickLower, int24 tickUpper);
    error InvalidTickSpacing(int24 tick, int24 tickSpacing);
    error SlippageLimitExceeded(uint160 currentPriceX96, uint160 limitPriceX96);
    error ZeroLiquidityMint();
    error StaleOraclePrice(uint256 priceTimestamp, uint256 currentTimestamp);
    error UnauthorizedCaller(address caller);
    error PoolLocked();

    // --- SWAP & VIEW FUNCTIONS ---

    /**
     * @notice Executes an atomic token swap across concentrated liquidity ticks
     * @param params Swap parameters specifying recipient, direction, amount, and price limit
     * @return amount0 Token0 balance change (positive: sent to pool, negative: received from pool)
     * @return amount1 Token1 balance change (positive: sent to pool, negative: received from pool)
     */
    function swap(SwapParams calldata params) external returns (int256 amount0, int256 amount1);

    /**
     * @notice Retrieves current pool state variables
     */
    function slot0() external view returns (
        uint160 sqrtPriceX96,
        int24 tick,
        uint16 observationIndex,
        uint16 observationCardinality,
        uint16 observationCardinalityNext,
        uint24 feeBps,
        bool unlocked
    );

    /**
     * @notice Retrieves active volatility collar configuration and boundary ticks
     */
    function getCollarConfig() external view returns (CollarConfig memory);

    /**
     * @notice Retrieves dynamic fee configuration
     */
    function getFeeConfig() external view returns (FeeConfig memory);

    /**
     * @notice Returns historical time-weighted average price (TWAP) tick cumulatives
     * @param secondsAgos Array of seconds relative to current timestamp to observe
     * @return tickCumulatives Cumulative tick values at specified timestamps
     * @return secondsPerLiquidityCumulativeX128s Cumulative seconds per liquidity values
     */
    function observe(uint32[] calldata secondsAgos)
        external
        view
        returns (int56[] memory tickCumulatives, uint160[] memory secondsPerLiquidityCumulativeX128s);

    /**
     * @notice Checks whether off-hours trading is currently active
     */
    function isOffHoursTradingActive() external view returns (bool isActive);

    // --- ADMINISTRATIVE FUNCTIONS ---

    /**
     * @notice Updates the primary exchange official closing price and recalibrates collar bands
     * @param closingPriceX96 New closing price in Q64.96 format
     * @param timestamp Timestamp of the exchange closing price fix
     * @param signature Cryptographic signature from authorized market data oracle
     */
    function setPrimaryClosingPrice(
        uint160 closingPriceX96,
        uint256 timestamp,
        bytes calldata signature
    ) external;

    /**
     * @notice Configures dynamic volatility collar band in basis points
     * @param newCollarBandBps Deviation band (e.g. 500 = 5%)
     */
    function setCollarBand(uint16 newCollarBandBps) external;

    /**
     * @notice Overrides trading session state for exchange holidays or emergency market interventions
     * @param forceActive True to permit trading, false to suspend
     * @param isHoliday True if overriding for scheduled exchange holiday
     */
    function setTradingSessionOverride(bool forceActive, bool isHoliday) external;
}
```

---

### Synthetic Liquidity Pool Interface (`ISyntheticLiquidityPool.sol`)
```solidity
// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title ISyntheticLiquidityPool
 * @notice Interface governing concentrated liquidity management and synthetic equity pairing
 * @dev Coordinates position lifecycle, fee calculations, and asset backing
 */
interface ISyntheticLiquidityPool {

    // --- STRUCTS ---

    struct PositionInfo {
        uint128 liquidity;                 // Amount of concentrated liquidity in position
        uint256 feeGrowthInside0LastX128;  // Fee growth checkpoint for token0
        uint256 feeGrowthInside1LastX128;  // Fee growth checkpoint for token1
        uint128 tokensOwed0;               // Uncollected token0 fees accrued to position
        uint128 tokensOwed1;               // Uncollected token1 fees accrued to position
    }

    struct MintParams {
        address recipient;                 // Address receiving the liquidity position credit
        int24 tickLower;                   // Lower price boundary tick
        int24 tickUpper;                   // Upper price boundary tick
        uint128 amount;                    // Desired liquidity units to add
        bytes data;                        // Optional callback payload
    }

    struct BurnParams {
        int24 tickLower;                   // Lower price boundary tick
        int24 tickUpper;                   // Upper price boundary tick
        uint128 amount;                    // Desired liquidity units to withdraw
    }

    struct CollectParams {
        address recipient;                 // Address receiving collected fees/principal
        int24 tickLower;                   // Lower price boundary tick
        int24 tickUpper;                   // Upper price boundary tick
        uint128 amount0Requested;          // Maximum token0 amount to harvest
        uint128 amount1Requested;          // Maximum token1 amount to harvest
    }

    // --- EVENTS ---

    event Mint(
        address indexed sender,
        address indexed owner,
        int24 indexed tickLower,
        int24 tickUpper,
        uint128 amount,
        uint256 amount0,
        uint256 amount1
    );

    event Burn(
        address indexed owner,
        int24 indexed tickLower,
        int24 tickUpper,
        uint128 amount,
        uint256 amount0,
        uint256 amount1
    );

    event Collect(
        address indexed owner,
        address recipient,
        int24 indexed tickLower,
        int24 tickUpper,
        uint128 amount0,
        uint128 amount1
    );

    // --- FUNCTIONS ---

    /**
     * @notice Mints concentrated liquidity within the specified tick range
     * @param params Parameters specifying recipient, tick boundaries, and liquidity units
     * @return amount0 Actual token0 amount required and deposited
     * @return amount1 Actual token1 amount required and deposited
     */
    function mint(MintParams calldata params) external returns (uint256 amount0, uint256 amount1);

    /**
     * @notice Burns concentrated liquidity and credits accrued tokens to position balance
     * @param params Parameters specifying tick range and liquidity units to remove
     * @return amount0 Token0 principal unlocked
     * @return amount1 Token1 principal unlocked
     */
    function burn(BurnParams calldata params) external returns (uint256 amount0, uint256 amount1);

    /**
     * @notice Harvests accumulated fees and unlocked principal from a position
     * @param params Parameters specifying destination and requested withdrawal caps
     * @return amount0 Actual token0 harvested
     * @return amount1 Actual token1 harvested
     */
    function collect(CollectParams calldata params) external returns (uint128 amount0, uint128 amount1);

    /**
     * @notice Retrieves state info for a specific liquidity position key
     * @param owner Address of the position owner (or PositionNFT manager)
     * @param tickLower Lower tick bound
     * @param tickUpper Upper tick bound
     * @return position PositionInfo struct containing liquidity and fee checkpoints
     */
    function getPosition(
        address owner,
        int24 tickLower,
        int24 tickUpper
    ) external view returns (PositionInfo memory position);

    /**
     * @notice Returns the token0 (Tokenized Equity) and token1 (Settlement Stable / eINR) addresses
     */
    function token0() external view returns (address);
    function token1() external view returns (address);

    /**
     * @notice Returns the immutable tick spacing enforced by this pool
     */
    function tickSpacing() external view returns (int24);

    /**
     * @notice Returns total active in-range liquidity
     */
    function liquidity() external view returns (uint128);
}
```

---

## Security & Compliance Notes
- **SEBI-Compliant Price Collars & Circuit Breakers:**
  - In compliance with SEBI master circular provisions governing equity volatility management, `OffHoursCLMM.sol` strictly enforces a non-bypassable collar (default +/- 5%, configurable up to statutory limits) anchored to the primary exchange closing price ($P_{close}$).
  - Any transaction that would drive $\sqrt{P}$ beyond $[0.95 \times P_{close}, 1.05 \times P_{close}]$ is either clamped at the boundary tick or immediately reverted with `CollarBreached`. This guarantees that opening price imbalances on primary exchanges (at 09:15 IST) cannot be destabilized by rogue off-hours activity.
- **Front-Running & MEV Elimination under QBFT:**
  - Deployed on a permissioned Hyperledger Besu consortium network with QBFT consensus, validator nodes operate under strict institutional governance without public mempools.
  - Swaps specify mandatory `sqrtPriceLimitX96` parameters, ensuring transactions revert deterministically if intermediate slippage exceeds user tolerance.
- **Flash-Loan Manipulation Resistance:**
  - Single-block flash-loan manipulation is mitigated via three synchronized layers:
    1. **Dynamic Fee Scaling:** Intraday tick traversal velocity sharply increases the marginal swap fee, making circular arbitrage or price manipulation economically irrational.
    2. **Geometric Mean TWAP Oracle:** Downstream protocol pricing relies on the integrated TWAP observation buffer rather than instantaneous spot reserves.
    3. **Collar Hard Boundary:** Arbitrage capital cannot dislocate prices beyond statutory collars regardless of capital volume.
- **Transient Storage & Reentrancy Defenses:**
  - All state-modifying swap and liquidity operations utilize reentrancy locks. In EVM Cancun deployments, `TSTORE` / `TLOAD` are leveraged for ultra-low gas overhead reentrancy guards.
  - Follows strict Checks-Effects-Interactions patterns: internal liquidity and tick balances update prior to issuing external token transfer callbacks.
- **Precision Arithmetic & Rounding Invariants:**
  - Intermediate math utilizes `FullMath.mulDiv` to maintain 512-bit precision before final division, eliminating integer overflow and truncation errors.
  - Rounding direction is strictly enforced in favor of the protocol pool:
    - Liquidity minting rounds input requirements UP.
    - Swaps round input requirements UP and output distributions DOWN.
    - Liquidity burns round output distributions DOWN.
    - This ensures pool virtual reserves can never experience fractional token leakage under high-frequency tick oscillation.
- **Regulatory Time-Gating & Audit Integrity:**
  - Swaps are cryptographically restricted to off-hours windows (`15:30:00 <= t < 09:15:00 IST`). Attempted executions during primary market hours immediately revert with `OutsideTradingHours()`.
  - Comprehensive event logging (`Swap`, `Mint`, `Burn`, `CollarUpdated`) provides complete auditability for SEBI/IFSCA regulatory telemetry and automated trade surveillance engines.

---

## Acceptance Criteria
- [ ] **Clean Compilation:** `contracts/src/liquidity/OffHoursCLMM.sol` and `CLMMPositionNFT.sol` compile cleanly under Solidity ^0.8.24 with zero compiler warnings.
- [ ] **Math Precision Verification:** `FullMath`, `TickMath`, and `SqrtPriceMath` pass 50,000 differential fuzz tests in Foundry matching reference Uniswap v3 vector outputs with zero precision loss.
- [ ] **Time-Gated Trading Gating:** Automated tests confirm that swaps succeed at 16:00 IST, succeed on weekends, and strictly revert with `OutsideTradingHours()` at 10:30 IST.
- [ ] **Collar Band Enforcement:** A swap attempting to push the execution price $> 5.00\%$ above or $< 5.00\%$ below $P_{close}$ strictly reverts with `CollarBreached`.
- [ ] **Dynamic Fee Scaling:** Verification tests confirm that swap fees increase monotonically as price approaches within the configured proximity buffer of the collar boundary.
- [ ] **TWAP Oracle Accuracy:** Oracle observation ring-buffer accurately records cumulative ticks across sequential blocks and provides tamper-resistant geometric mean prices over 5-minute and 15-minute windows.
- [ ] **NFT Position Management:** `CLMMPositionNFT.sol` correctly mints ERC-721 positions, tracks uncollected fees, and restricts fee harvesting strictly to the NFT owner.
- [ ] **Reserve Solvency Invariant:** Invariant fuzz tests across 20,000 runs confirm that actual contract token balances always satisfy $\text{Balance}_k \ge \text{VirtualReserves}_k + \text{UncollectedFees}_k + \text{ProtocolFees}_k$.
- [ ] **Gas Target Compliance:** Benchmark tests confirm standard single-tick crossing swaps consume $\le 110,000$ gas on Hyperledger Besu.
- [ ] **Static Analysis Cleanliness:** Slither static analysis reports zero high or medium severity findings.

---

## Suggested Order / Dependencies
- **Prerequisites:**
  - Prompt `207` (Market Data Service - provides signed primary exchange closing price feeds).
  - Prompt `301` (Permissioned Blockchain Evaluation & Selection - establishes Besu QBFT network baseline).
  - Prompt `303` (Token Issuance Smart Contract - provides tokenized equity ERC-20 tokens).
  - Prompt `305` (Transfer Compliance Hooks - verifies token transfer restrictions before pool interactions).
- **Parallel Tasks:**
  - Prompt `210` (Fee and Realized PnL Engine - designs off-chain fee aggregation accounting).
  - Prompt `242` (Primary Market Adapter - integrates exchange session schedules and corporate action signals).
  - Prompt `341` (Circuit Breaker Timelock & Emergency Halt Contract - provides emergency freeze infrastructure).
- **Subsequent Prompts Enabled:**
  - Prompt `309` (Chain Event Indexer - indexes off-hours pool swaps and liquidity metrics).
  - Prompt `315` (Settlement Guarantee Fund Contract - connects off-hours liquidity backing to platform risk pools).
  - Prompt `601` (Trading Terminal UI - displays off-hours order book depth, candlestick charts, and concentrated liquidity visualizers).
