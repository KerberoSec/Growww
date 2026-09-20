// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title ISecuritiesAMM
 * @notice Interface for ERC-3643 compliant constant-product Automated Market Maker (XYK) pool.
 * @dev Enforces investor KYC/compliance checks on secondary trading and liquidity provision.
 */
interface ISecuritiesAMM {
    // =========================================================================
    // Events
    // =========================================================================

    event LiquidityAdded(
        address indexed provider,
        uint256 amount0,
        uint256 amount1,
        uint256 lpSharesMinted
    );

    event LiquidityRemoved(
        address indexed provider,
        uint256 amount0,
        uint256 amount1,
        uint256 lpSharesBurned,
        address indexed recipient
    );

    event Swap(
        address indexed sender,
        address indexed recipient,
        address indexed tokenIn,
        uint256 amountIn,
        uint256 amountOut,
        uint256 feeAmount
    );

    event ReservesSynced(uint256 reserve0, uint256 reserve1);
    event FeeRateUpdated(uint256 oldFeeBps, uint256 newFeeBps);
    event FeeCollectorUpdated(address oldCollector, address newCollector);
    event IdentityRegistryUpdated(address oldRegistry, address newRegistry);

    // =========================================================================
    // Errors
    // =========================================================================

    error CallerNotCompliant(address caller);
    error RecipientNotCompliant(address recipient);
    error InsufficientLiquidityMinted();
    error InsufficientLiquidityBurned();
    error InsufficientOutputAmount(uint256 requested, uint256 actual);
    error InsufficientInputAmount();
    error SlippageExceeded(uint256 expectedMin, uint256 actual);
    error DeadlineExpired(uint256 deadline, uint256 blockTimestamp);
    error IdenticalTokens();
    error InvalidZeroAddress();
    error InvalidAmount();
    error ExcessiveFeeBps(uint256 feeBps, uint256 maxFeeBps);
    error KInvariantViolated(uint256 currentK, uint256 requiredK);
    error PoolPaused();

    // =========================================================================
    // View Functions
    // =========================================================================

    function token0() external view returns (address);
    function token1() external view returns (address);
    function identityRegistry() external view returns (address);
    function complianceRegistry() external view returns (address);
    function feeBps() external view returns (uint256);
    function protocolFeeBps() external view returns (uint256);
    function feeCollector() external view returns (address);
    function getReserves() external view returns (uint256 reserve0, uint256 reserve1, uint32 blockTimestampLast);
    function getAmountOut(uint256 amountIn, address tokenIn) external view returns (uint256 amountOut, uint256 fee);
    function quote(uint256 amountA, uint256 reserveA, uint256 reserveB) external pure returns (uint256 amountB);

    // =========================================================================
    // State-Changing Functions
    // =========================================================================

    function addLiquidity(
        uint256 amount0Desired,
        uint256 amount1Desired,
        uint256 amount0Min,
        uint256 amount1Min,
        address to,
        uint256 deadline
    ) external returns (uint256 amount0, uint256 amount1, uint256 lpShares);

    function removeLiquidity(
        uint256 lpShares,
        uint256 amount0Min,
        uint256 amount1Min,
        address to,
        uint256 deadline
    ) external returns (uint256 amount0, uint256 amount1);

    function swapExactTokensForTokens(
        uint256 amountIn,
        uint256 amountOutMin,
        address tokenIn,
        address to,
        uint256 deadline
    ) external returns (uint256 amountOut);

    function sync() external;
}
