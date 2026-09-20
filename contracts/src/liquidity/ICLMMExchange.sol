// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title ICLMMExchange
 * @notice Primary interface for Concentrated Liquidity Market Maker Off-Hours AMM
 * @dev Enforces tick-based concentrated liquidity, dynamic volatility collars, and time-gated sessions
 */
interface ICLMMExchange {

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

    struct SwapParams {
        address recipient;                 // Address receiving swapped output tokens
        bool zeroForOne;                   // True if trading token0 for token1, false if token1 for token0
        int256 amountSpecified;            // Exact input (> 0) or exact output (< 0) amount
        uint160 sqrtPriceLimitX96;         // Execution price limit protecting against slippage
        bytes data;                        // Optional callback data passed to swap caller
    }

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

    error OutsideTradingHours();
    error CollarBreached(uint160 targetPriceX96, uint160 boundaryPriceX96);
    error InvalidTickRange(int24 tickLower, int24 tickUpper);
    error InvalidTickSpacing(int24 tick, int24 tickSpacing);
    error SlippageLimitExceeded(uint160 currentPriceX96, uint160 limitPriceX96);
    error ZeroLiquidityMint();
    error StaleOraclePrice(uint256 priceTimestamp, uint256 currentTimestamp);
    error UnauthorizedCaller(address caller);
    error PoolLocked();

    function swap(SwapParams calldata params) external returns (int256 amount0, int256 amount1);

    function slot0() external view returns (
        uint160 sqrtPriceX96,
        int24 tick,
        uint16 observationIndex,
        uint16 observationCardinality,
        uint16 observationCardinalityNext,
        uint24 feeBps,
        bool unlocked
    );

    function getCollarConfig() external view returns (CollarConfig memory);
}
