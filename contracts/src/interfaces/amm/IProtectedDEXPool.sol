// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IProtectedDEXPool
 * @notice Interface for institutional DEX pool with flash loan defense and oracle TWAP manipulation protection.
 */
interface IProtectedDEXPool {
    // =========================================================================
    // Events
    // =========================================================================

    event TwapUpdated(
        uint256 price0Cumulative,
        uint256 price1Cumulative,
        uint32 blockTimestamp
    );

    event FlashLoanExecuted(
        address indexed borrower,
        address indexed token,
        uint256 amount,
        uint256 fee
    );

    event MaxTwapDeviationUpdated(uint256 oldDeviationBps, uint256 newDeviationBps);
    event MaxBlockPriceImpactUpdated(uint256 oldImpactBps, uint256 newImpactBps);
    event FlashLoanFeeUpdated(uint256 oldFeeBps, uint256 newFeeBps);
    event FlashLoanProtectionToggled(bool enabled);

    // =========================================================================
    // Errors
    // =========================================================================

    error SameBlockTradeRestricted(address account, uint256 blockNumber);
    error MaxBlockPriceImpactExceeded(uint256 impactBps, uint256 maxImpactBps);
    error TWAPManipulationDetected(uint256 spotPrice, uint256 twapPrice, uint256 deviationBps);
    error InsufficientFlashLoanLiquidity(address token, uint256 requested, uint256 available);
    error FlashLoanNotRepaid(address token, uint256 required, uint256 actual);
    error UnauthorizedCallbackCaller(address caller, address pool);
    error TwapWindowNotElapsed(uint256 elapsed, uint256 requiredWindow);
    error FlashLoanReentrancyProhibited();

    // =========================================================================
    // View Functions
    // =========================================================================

    function price0CumulativeLast() external view returns (uint256);
    function price1CumulativeLast() external view returns (uint256);
    function blockTimestampLast() external view returns (uint32);
    function maxTwapDeviationBps() external view returns (uint256);
    function maxBlockPriceImpactBps() external view returns (uint256);
    function twapWindow() external view returns (uint256);
    function flashLoanFeeBps() external view returns (uint256);
    function flashLoanProtectionEnabled() external view returns (bool);
    function lastTradeBlock(address account) external view returns (uint256);

    function consultTwap(address token, uint256 amountIn, uint256 period) external view returns (uint256 amountOut);

    // =========================================================================
    // Flash Loan Function
    // =========================================================================

    function flashLoan(
        address receiver,
        address token,
        uint256 amount,
        bytes calldata data
    ) external returns (bool);
}

interface IFlashBorrower {
    function onFlashLoan(
        address initiator,
        address token,
        uint256 amount,
        uint256 fee,
        bytes calldata data
    ) external returns (bytes32);
}
