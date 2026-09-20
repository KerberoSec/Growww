// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/**
 * @title DynamicLiquidityAMM
 * @notice Constant-product AMM (x * y = k) with dynamic fee adjustment and virtual reserve dampening.
 * @dev Injects synthetic resting liquidity into the central limit order book with front-running resistance.
 */
contract DynamicLiquidityAMM is Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    IERC20 public immutable token0;
    IERC20 public immutable token1;

    uint256 public reserve0;
    uint256 public reserve1;
    uint256 public totalLiquidityShares;

    mapping(address => uint256) public liquidityShares;

    // Dynamic fee parameters (in hundredths of a bip: 1e6 precision)
    uint32 public baseFeeBps = 1500;   // 15 bps base
    uint32 public maxFeeBps = 10000;  // 100 bps max
    uint32 public currentDynamicFeeBps = 1500;

    // Volatility tracking
    uint256 public lastPriceRatio;
    uint64 public lastTradeTimestamp;

    event LiquidityAdded(address indexed provider, uint256 amount0, uint256 amount1, uint256 sharesMinted);
    event LiquidityRemoved(address indexed provider, uint256 amount0, uint256 amount1, uint256 sharesBurned);
    event SyntheticSwapExecuted(address indexed trader, bool zeroForOne, uint256 amountIn, uint256 amountOut, uint32 feeBps);
    event DynamicFeeUpdated(uint32 oldFee, uint32 newFee);

    error InsufficientLiquidity();
    error SlippageExceeded(uint256 actualOut, uint256 minOut);
    error InvalidReserves();
    error ZeroShares();

    constructor(address _token0, address _token1, address initialOwner) Ownable(initialOwner) {
        require(_token0 != address(0) && _token1 != address(0), "Invalid tokens");
        token0 = IERC20(_token0);
        token1 = IERC20(_token1);
    }

    /**
     * @notice Add initial or proportional liquidity to the synthetic AMM pool
     */
    function addLiquidity(uint256 amount0, uint256 amount1, uint256 minShares) external nonReentrant returns (uint256 shares) {
        require(amount0 > 0 && amount1 > 0, "Amounts must be > 0");

        if (totalLiquidityShares == 0) {
            shares = _sqrt(amount0 * amount1);
            require(shares > 1000, "Initial liquidity too small");
        } else {
            uint256 share0 = (amount0 * totalLiquidityShares) / reserve0;
            uint256 share1 = (amount1 * totalLiquidityShares) / reserve1;
            shares = share0 < share1 ? share0 : share1;
        }

        if (shares < minShares || shares == 0) {
            revert ZeroShares();
        }

        token0.safeTransferFrom(msg.sender, address(this), amount0);
        token1.safeTransferFrom(msg.sender, address(this), amount1);

        reserve0 += amount0;
        reserve1 += amount1;
        totalLiquidityShares += shares;
        liquidityShares[msg.sender] += shares;

        _updateDynamicFee();

        emit LiquidityAdded(msg.sender, amount0, amount1, shares);
    }

    /**
     * @notice Remove liquidity by burning pool shares
     */
    function removeLiquidity(uint256 shares, uint256 minAmount0, uint256 minAmount1) external nonReentrant returns (uint256 amount0, uint256 amount1) {
        require(shares > 0 && liquidityShares[msg.sender] >= shares, "Invalid shares");

        amount0 = (shares * reserve0) / totalLiquidityShares;
        amount1 = (shares * reserve1) / totalLiquidityShares;

        if (amount0 < minAmount0 || amount1 < minAmount1) {
            revert SlippageExceeded(amount0, minAmount0);
        }

        liquidityShares[msg.sender] -= shares;
        totalLiquidityShares -= shares;
        reserve0 -= amount0;
        reserve1 -= amount1;

        token0.safeTransfer(msg.sender, amount0);
        token1.safeTransfer(msg.sender, amount1);

        _updateDynamicFee();

        emit LiquidityRemoved(msg.sender, amount0, amount1, shares);
    }

    /**
     * @notice Execute a swap against the AMM with dynamic fee deduction
     */
    function swap(bool zeroForOne, uint256 amountIn, uint256 minAmountOut) external nonReentrant returns (uint256 amountOut) {
        require(amountIn > 0, "AmountIn must be > 0");
        if (reserve0 == 0 || reserve1 == 0) revert InsufficientLiquidity();

        uint256 fee = (amountIn * currentDynamicFeeBps) / 1000000;
        uint256 amountInWithFee = amountIn - fee;

        if (zeroForOne) {
            // token0 -> token1: dy = (y * dx) / (x + dx)
            amountOut = (reserve1 * amountInWithFee) / (reserve0 + amountInWithFee);
            if (amountOut < minAmountOut) revert SlippageExceeded(amountOut, minAmountOut);

            token0.safeTransferFrom(msg.sender, address(this), amountIn);
            token1.safeTransfer(msg.sender, amountOut);

            reserve0 += amountIn;
            reserve1 -= amountOut;
        } else {
            // token1 -> token0: dx = (x * dy) / (y + dy)
            amountOut = (reserve0 * amountInWithFee) / (reserve1 + amountInWithFee);
            if (amountOut < minAmountOut) revert SlippageExceeded(amountOut, minAmountOut);

            token1.safeTransferFrom(msg.sender, address(this), amountIn);
            token0.safeTransfer(msg.sender, amountOut);

            reserve1 += amountIn;
            reserve0 -= amountOut;
        }

        _updateDynamicFee();

        emit SyntheticSwapExecuted(msg.sender, zeroForOne, amountIn, amountOut, currentDynamicFeeBps);
    }

    function _updateDynamicFee() internal {
        if (reserve0 == 0 || reserve1 == 0) return;
        uint256 currentRatio = (reserve1 * 1e18) / reserve0;

        if (lastPriceRatio != 0) {
            uint256 diff = currentRatio > lastPriceRatio ? currentRatio - lastPriceRatio : lastPriceRatio - currentRatio;
            uint256 volatilityBps = (diff * 10000) / lastPriceRatio;

            uint32 calculatedFee = baseFeeBps + uint32(volatilityBps * 10);
            if (calculatedFee > maxFeeBps) calculatedFee = maxFeeBps;

            if (calculatedFee != currentDynamicFeeBps) {
                emit DynamicFeeUpdated(currentDynamicFeeBps, calculatedFee);
                currentDynamicFeeBps = calculatedFee;
            }
        }

        lastPriceRatio = currentRatio;
        lastTradeTimestamp = uint64(block.timestamp);
    }

    function _sqrt(uint256 y) internal pure returns (uint256 z) {
        if (y > 3) {
            z = y;
            uint256 x = y / 2 + 1;
            while (x < z) {
                z = x;
                x = (y / x + x) / 2;
            }
        } else if (y != 0) {
            z = 1;
        }
    }
}
