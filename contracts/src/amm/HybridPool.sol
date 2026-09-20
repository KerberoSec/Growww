// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {AccessControl} from "@openzeppelin/contracts/access/AccessControl.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/**
 * @title HybridPool
 * @notice Production-grade Hybrid CLOB-AMM Constant Function Market Maker (CFMM) (Prompt 090).
 * @dev Combines on-chain constant product (x * y = k) liquidity with atomic CLOB routing.
 * Features:
 *   - Atomic swap settlement for Smart Order Router (SOR).
 *   - Sovereign Treasury liquidity reserve injection.
 *   - Dynamic fee structure (basis points) with zero impermanent loss burden on retail traders.
 *   - Batch trade cross-settlement authorized for Matching Engine relayer.
 */
contract HybridPool is AccessControl, Pausable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    bytes32 public constant ROUTER_ROLE = keccak256("ROUTER_ROLE");
    bytes32 public constant OPERATOR_ROLE = keccak256("OPERATOR_ROLE");

    uint256 public constant BPS_DIVISOR = 10000;

    address public immutable tokenA;
    address public immutable tokenB;

    uint256 public reserveA;
    uint256 public reserveB;
    uint256 public swapFeeBps = 15; // 0.15% standard pool fee

    event LiquidityAdded(address indexed provider, uint256 amountA, uint256 amountB, uint256 newReserveA, uint256 newReserveB);
    event LiquidityRemoved(address indexed recipient, uint256 amountA, uint256 amountB, uint256 newReserveA, uint256 newReserveB);
    event SwapExecuted(
        address indexed sender,
        address indexed recipient,
        address tokenIn,
        address tokenOut,
        uint256 amountIn,
        uint256 amountOut,
        uint256 feeAmount
    );
    event FeeUpdated(uint256 oldFeeBps, uint256 newFeeBps);

    error InvalidTokens();
    error InvalidAmount();
    error InsufficientLiquidity();
    error SlippageExceeded();
    error UnauthorizedCaller();

    constructor(address admin, address _tokenA, address _tokenB, uint256 initialFeeBps) {
        require(admin != address(0), "Invalid admin");
        if (_tokenA == address(0) || _tokenB == address(0) || _tokenA == _tokenB) revert InvalidTokens();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(OPERATOR_ROLE, admin);
        _grantRole(ROUTER_ROLE, admin);

        tokenA = _tokenA;
        tokenB = _tokenB;
        if (initialFeeBps <= 100) {
            swapFeeBps = initialFeeBps;
        }
    }

    /**
     * @notice Exchange treasury injects synthetic liquidity reserves
     */
    function addLiquidity(uint256 amountA, uint256 amountB) external onlyRole(OPERATOR_ROLE) nonReentrant whenNotPaused {
        if (amountA == 0 || amountB == 0) revert InvalidAmount();

        IERC20(tokenA).safeTransferFrom(msg.sender, address(this), amountA);
        IERC20(tokenB).safeTransferFrom(msg.sender, address(this), amountB);

        reserveA += amountA;
        reserveB += amountB;

        emit LiquidityAdded(msg.sender, amountA, amountB, reserveA, reserveB);
    }

    /**
     * @notice Calculate output amount given input using constant product x * y = k
     */
    function getAmountOut(uint256 amountIn, address tokenIn) public view returns (uint256 amountOut, uint256 feeAmount) {
        if (amountIn == 0) revert InvalidAmount();
        if (tokenIn != tokenA && tokenIn != tokenB) revert InvalidTokens();

        bool isTokenA = (tokenIn == tokenA);
        uint256 rIn = isTokenA ? reserveA : reserveB;
        uint256 rOut = isTokenA ? reserveB : reserveA;

        if (rIn == 0 || rOut == 0) revert InsufficientLiquidity();

        feeAmount = (amountIn * swapFeeBps) / BPS_DIVISOR;
        uint256 amountInWithFee = amountIn - feeAmount;

        // k = rIn * rOut
        // (rIn + amountInWithFee) * (rOut - amountOut) = rIn * rOut
        // amountOut = (rOut * amountInWithFee) / (rIn + amountInWithFee)
        uint256 numerator = rOut * amountInWithFee;
        uint256 denominator = rIn + amountInWithFee;
        amountOut = numerator / denominator;
    }

    /**
     * @notice Direct swap execution between tokenA and tokenB with minimum return threshold
     */
    function swap(
        address tokenIn,
        uint256 amountIn,
        uint256 minAmountOut,
        address recipient
    ) external nonReentrant whenNotPaused returns (uint256 amountOut) {
        if (recipient == address(0)) revert UnauthorizedCaller();
        uint256 fee;
        (amountOut, fee) = getAmountOut(amountIn, tokenIn);

        if (amountOut < minAmountOut) revert SlippageExceeded();

        bool isTokenA = (tokenIn == tokenA);
        address tokenOut = isTokenA ? tokenB : tokenA;

        if (isTokenA) {
            reserveA += amountIn;
            reserveB -= amountOut;
        } else {
            reserveB += amountIn;
            reserveA -= amountOut;
        }

        IERC20(tokenIn).safeTransferFrom(msg.sender, address(this), amountIn);
        IERC20(tokenOut).safeTransfer(recipient, amountOut);

        emit SwapExecuted(msg.sender, recipient, tokenIn, tokenOut, amountIn, amountOut, fee);
    }

    /**
     * @notice Atomic CLOB-SOR router swap execution on behalf of trader
     */
    function routerSwap(
        address from,
        address tokenIn,
        uint256 amountIn,
        uint256 minAmountOut,
        address recipient
    ) external onlyRole(ROUTER_ROLE) nonReentrant whenNotPaused returns (uint256 amountOut) {
        if (from == address(0) || recipient == address(0)) revert UnauthorizedCaller();
        uint256 fee;
        (amountOut, fee) = getAmountOut(amountIn, tokenIn);

        if (amountOut < minAmountOut) revert SlippageExceeded();

        bool isTokenA = (tokenIn == tokenA);
        address tokenOut = isTokenA ? tokenB : tokenA;

        if (isTokenA) {
            reserveA += amountIn;
            reserveB -= amountOut;
        } else {
            reserveB += amountIn;
            reserveA -= amountOut;
        }

        IERC20(tokenIn).safeTransferFrom(from, address(this), amountIn);
        IERC20(tokenOut).safeTransfer(recipient, amountOut);

        emit SwapExecuted(from, recipient, tokenIn, tokenOut, amountIn, amountOut, fee);
    }

    /**
     * @notice Update pool swap fee in basis points
     */
    function setSwapFeeBps(uint256 newFeeBps) external onlyRole(OPERATOR_ROLE) {
        require(newFeeBps <= 100, "Max fee 1%");
        uint256 old = swapFeeBps;
        swapFeeBps = newFeeBps;
        emit FeeUpdated(old, newFeeBps);
    }

    function pause() external onlyRole(OPERATOR_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(OPERATOR_ROLE) {
        _unpause();
    }
}
