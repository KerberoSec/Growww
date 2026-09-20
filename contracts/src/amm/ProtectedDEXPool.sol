// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";

import {IProtectedDEXPool, IFlashBorrower} from "../interfaces/amm/IProtectedDEXPool.sol";

/**
 * @title ProtectedDEXPool
 * @notice Institutional DEX pool engineered with flash loan defenses, reentrancy guards,
 *         and oracle TWAP manipulation resistant collars.
 */
contract ProtectedDEXPool is ERC20, ReentrancyGuard, Pausable, Ownable, IProtectedDEXPool {
    using SafeERC20 for IERC20;

    // --- Constants ---
    uint256 public constant MINIMUM_LIQUIDITY = 1000;
    uint256 public constant BPS_DENOMINATOR = 10000;
    uint256 public constant PRECISION = 1e18;
    bytes32 public constant FLASH_LOAN_CALLBACK_SUCCESS = keccak256("ERC3156FlashBorrower.onFlashLoan");

    // --- Pool Tokens ---
    address public immutable token0;
    address public immutable token1;

    // --- Reserves & TWAP State ---
    uint112 private _reserve0;
    uint112 private _reserve1;
    uint32 public override blockTimestampLast;

    uint256 public override price0CumulativeLast;
    uint256 public override price1CumulativeLast;

    // Historical TWAP snapshot for window calculation
    struct TwapSnapshot {
        uint256 price0Cumulative;
        uint256 price1Cumulative;
        uint32 timestamp;
    }
    TwapSnapshot public snapshotStart;

    // --- Security & Manipulation Resistance Parameters ---
    uint256 public override maxTwapDeviationBps;     // e.g. 500 = 5% max deviation
    uint256 public override maxBlockPriceImpactBps;  // e.g. 300 = 3% max impact per block
    uint256 public override twapWindow;              // e.g. 300 seconds
    uint256 public override flashLoanFeeBps;          // e.g. 9 = 0.09%
    bool public override flashLoanProtectionEnabled;

    // Per-block price tracking for block impact limits
    uint256 public blockStartPrice0E18;
    uint256 public currentPriceBlockNumber;

    // Anti-flash loan / sandwich attack tracking
    mapping(address => uint256) public override lastTradeBlock;

    // Pool swap fee
    uint256 public swapFeeBps; // e.g. 30 bps (0.30%)

    // Flash loan active reentrancy flag
    bool private _inFlashLoan;

    event SwapExecuted(
        address indexed sender,
        address indexed recipient,
        address indexed tokenIn,
        uint256 amountIn,
        uint256 amountOut,
        uint256 spotPriceE18
    );
    event LiquidityAdded(address indexed provider, uint256 amount0, uint256 amount1, uint256 lpShares);
    event LiquidityRemoved(address indexed provider, uint256 amount0, uint256 amount1, uint256 lpShares);

    constructor(
        address _tokenA,
        address _tokenB,
        uint256 _swapFeeBps,
        uint256 _flashLoanFeeBps,
        uint256 _maxTwapDeviationBps,
        uint256 _maxBlockPriceImpactBps,
        uint256 _twapWindow,
        address initialOwner
    )
        ERC20("Growww Protected DEX LP Token", "G-PROT-LP")
        Ownable(initialOwner)
    {
        require(_tokenA != address(0) && _tokenB != address(0), "Invalid tokens");
        require(_tokenA != _tokenB, "Identical tokens");
        require(_maxTwapDeviationBps <= 2000, "Deviation too high"); // max 20%
        require(_maxBlockPriceImpactBps <= 1000, "Impact too high"); // max 10%

        if (_tokenA < _tokenB) {
            token0 = _tokenA;
            token1 = _tokenB;
        } else {
            token0 = _tokenB;
            token1 = _tokenA;
        }

        swapFeeBps = _swapFeeBps;
        flashLoanFeeBps = _flashLoanFeeBps;
        maxTwapDeviationBps = _maxTwapDeviationBps;
        maxBlockPriceImpactBps = _maxBlockPriceImpactBps;
        twapWindow = _twapWindow;
        flashLoanProtectionEnabled = true;
    }

    // =========================================================================
    // View Functions
    // =========================================================================

    function getReserves() external view returns (uint112 reserve0, uint112 reserve1, uint32 lastTimestamp) {
        return (_reserve0, _reserve1, blockTimestampLast);
    }

    function consultTwap(
        address tokenIn,
        uint256 amountIn,
        uint256 period
    ) public view override returns (uint256 amountOut) {
        require(amountIn > 0, "Zero amount");
        require(tokenIn == token0 || tokenIn == token1, "Invalid token");
        require(period >= twapWindow, "Period below twap window");
        
        uint32 timeElapsed = uint32(block.timestamp) - snapshotStart.timestamp;
        if (timeElapsed == 0) {
            // Fallback to instantaneous price if snapshot just started
            return _getInstantaneousQuote(tokenIn, amountIn);
        }

        uint256 currentP0Cum = price0CumulativeLast;
        uint256 currentP1Cum = price1CumulativeLast;
        uint32 timeSinceLast = uint32(block.timestamp) - blockTimestampLast;

        if (timeSinceLast > 0 && _reserve0 > 0 && _reserve1 > 0) {
            currentP0Cum += (uint256(_reserve1) * PRECISION / uint256(_reserve0)) * timeSinceLast;
            currentP1Cum += (uint256(_reserve0) * PRECISION / uint256(_reserve1)) * timeSinceLast;
        }

        if (tokenIn == token0) {
            uint256 twapPrice0 = (currentP0Cum - snapshotStart.price0Cumulative) / timeElapsed;
            amountOut = (amountIn * twapPrice0) / PRECISION;
        } else {
            uint256 twapPrice1 = (currentP1Cum - snapshotStart.price1Cumulative) / timeElapsed;
            amountOut = (amountIn * twapPrice1) / PRECISION;
        }
    }

    // =========================================================================
    // Liquidity Management
    // =========================================================================

    function addLiquidity(
        uint256 amount0Desired,
        uint256 amount1Desired,
        address to
    ) external nonReentrant whenNotPaused returns (uint256 amount0, uint256 amount1, uint256 lpShares) {
        require(to != address(0), "Zero address");

        (uint256 currentRes0, uint256 currentRes1) = (_reserve0, _reserve1);

        if (currentRes0 == 0 && currentRes1 == 0) {
            amount0 = amount0Desired;
            amount1 = amount1Desired;
        } else {
            uint256 amount1Optimal = (amount0Desired * currentRes1) / currentRes0;
            if (amount1Optimal <= amount1Desired) {
                amount0 = amount0Desired;
                amount1 = amount1Optimal;
            } else {
                uint256 amount0Optimal = (amount1Desired * currentRes0) / currentRes1;
                require(amount0Optimal <= amount0Desired, "Excess optimal");
                amount0 = amount0Optimal;
                amount1 = amount1Desired;
            }
        }

        IERC20(token0).safeTransferFrom(msg.sender, address(this), amount0);
        IERC20(token1).safeTransferFrom(msg.sender, address(this), amount1);

        uint256 totalLpSupply = totalSupply();
        if (totalLpSupply == 0) {
            uint256 initialLiquidity = Math.sqrt(amount0 * amount1);
            require(initialLiquidity > MINIMUM_LIQUIDITY, "Below min liquidity");
            lpShares = initialLiquidity - MINIMUM_LIQUIDITY;
            _mint(address(0x000000000000000000000000000000000000dEaD), MINIMUM_LIQUIDITY);

            // Initialize TWAP snapshot
            snapshotStart = TwapSnapshot({
                price0Cumulative: 0,
                price1Cumulative: 0,
                timestamp: uint32(block.timestamp)
            });
        } else {
            lpShares = Math.min(
                (amount0 * totalLpSupply) / currentRes0,
                (amount1 * totalLpSupply) / currentRes1
            );
        }

        require(lpShares > 0, "Zero LP shares");
        _mint(to, lpShares);

        _updateReservesAndTwap(
            IERC20(token0).balanceOf(address(this)),
            IERC20(token1).balanceOf(address(this))
        );
        emit LiquidityAdded(msg.sender, amount0, amount1, lpShares);
    }

    function removeLiquidity(
        uint256 lpShares,
        address to
    ) external nonReentrant whenNotPaused returns (uint256 amount0, uint256 amount1) {
        require(to != address(0), "Zero address");
        require(lpShares > 0, "Zero shares");

        uint256 totalLpSupply = totalSupply();
        uint256 balance0 = IERC20(token0).balanceOf(address(this));
        uint256 balance1 = IERC20(token1).balanceOf(address(this));

        amount0 = (lpShares * balance0) / totalLpSupply;
        amount1 = (lpShares * balance1) / totalLpSupply;
        require(amount0 > 0 && amount1 > 0, "Zero withdrawn");

        _burn(msg.sender, lpShares);

        IERC20(token0).safeTransfer(to, amount0);
        IERC20(token1).safeTransfer(to, amount1);

        _updateReservesAndTwap(
            IERC20(token0).balanceOf(address(this)),
            IERC20(token1).balanceOf(address(this))
        );
        emit LiquidityRemoved(msg.sender, amount0, amount1, lpShares);
    }

    // =========================================================================
    // Swaps with Flash Loan & TWAP Manipulation Defenses
    // =========================================================================

    function swap(
        uint256 amountIn,
        uint256 amountOutMin,
        address tokenIn,
        address to
    ) external nonReentrant whenNotPaused returns (uint256 amountOut) {
        require(amountIn > 0, "Zero amount");
        require(to != address(0), "Zero to address");
        require(tokenIn == token0 || tokenIn == token1, "Invalid token");

        // 1. Flash Loan / Sandwich Protection
        if (flashLoanProtectionEnabled) {
            if (lastTradeBlock[msg.sender] == block.number) {
                revert SameBlockTradeRestricted(msg.sender, block.number);
            }
            lastTradeBlock[msg.sender] = block.number;
        }

        // 2. Compute Spot Price and verify against TWAP Collar
        address tokenOut = tokenIn == token0 ? token1 : token0;
        uint256 reserveIn = tokenIn == token0 ? _reserve0 : _reserve1;
        uint256 reserveOut = tokenIn == token0 ? _reserve1 : _reserve0;
        require(reserveIn > 0 && reserveOut > 0, "No liquidity");

        uint256 feeAmount = (amountIn * swapFeeBps) / BPS_DENOMINATOR;
        uint256 amountInWithFee = amountIn - feeAmount;
        amountOut = (amountInWithFee * reserveOut) / (reserveIn + amountInWithFee);
        require(amountOut >= amountOutMin, "Slippage exceeded");

        // Compute simulated post-trade reserves
        uint256 postReserve0 = tokenIn == token0 ? _reserve0 + amountIn : _reserve0 - amountOut;
        uint256 postReserve1 = tokenIn == token0 ? _reserve1 - amountOut : _reserve1 + amountIn;
        uint256 spotPricePost = (postReserve1 * PRECISION) / postReserve0;

        // 3. Check Block Price Impact
        if (currentPriceBlockNumber != block.number) {
            currentPriceBlockNumber = block.number;
            blockStartPrice0E18 = (_reserve1 * PRECISION) / _reserve0;
        } else if (maxBlockPriceImpactBps > 0 && blockStartPrice0E18 > 0) {
            uint256 priceDiff = spotPricePost > blockStartPrice0E18
                ? spotPricePost - blockStartPrice0E18
                : blockStartPrice0E18 - spotPricePost;
            uint256 impactBps = (priceDiff * BPS_DENOMINATOR) / blockStartPrice0E18;
            if (impactBps > maxBlockPriceImpactBps) {
                revert MaxBlockPriceImpactExceeded(impactBps, maxBlockPriceImpactBps);
            }
        }

        // 4. Check TWAP Manipulation Collar if TWAP history exists
        if (maxTwapDeviationBps > 0 && (block.timestamp - snapshotStart.timestamp) >= twapWindow) {
            uint256 twapOut = consultTwap(tokenIn, amountIn, block.timestamp - snapshotStart.timestamp);
            if (twapOut > 0) {
                uint256 diff = amountOut > twapOut ? amountOut - twapOut : twapOut - amountOut;
                uint256 deviationBps = (diff * BPS_DENOMINATOR) / twapOut;
                if (deviationBps > maxTwapDeviationBps) {
                    revert TWAPManipulationDetected(amountOut, twapOut, deviationBps);
                }
            }
        }

        // Execute transfers
        IERC20(tokenIn).safeTransferFrom(msg.sender, address(this), amountIn);
        IERC20(tokenOut).safeTransfer(to, amountOut);

        // Update reserves and TWAP accumulators
        _updateReservesAndTwap(
            IERC20(token0).balanceOf(address(this)),
            IERC20(token1).balanceOf(address(this))
        );

        emit SwapExecuted(msg.sender, to, tokenIn, amountIn, amountOut, spotPricePost);
    }

    // =========================================================================
    // Institutional Flash Loan Facility
    // =========================================================================

    function flashLoan(
        address receiver,
        address token,
        uint256 amount,
        bytes calldata data
    ) external override nonReentrant whenNotPaused returns (bool) {
        if (_inFlashLoan) revert FlashLoanReentrancyProhibited();
        require(token == token0 || token == token1, "Invalid token");
        require(amount > 0, "Zero amount");
        require(receiver != address(0), "Zero receiver");

        uint256 availableBalance = IERC20(token).balanceOf(address(this));
        if (amount > availableBalance) revert InsufficientFlashLoanLiquidity(token, amount, availableBalance);

        uint256 fee = (amount * flashLoanFeeBps) / BPS_DENOMINATOR;
        uint256 balanceBefore = availableBalance;

        _inFlashLoan = true;

        // Transfer flash loan funds to borrower
        IERC20(token).safeTransfer(receiver, amount);

        // Execute callback
        bytes32 callbackResult = IFlashBorrower(receiver).onFlashLoan(
            msg.sender,
            token,
            amount,
            fee,
            data
        );
        if (callbackResult != FLASH_LOAN_CALLBACK_SUCCESS) {
            revert UnauthorizedCallbackCaller(receiver, address(this));
        }

        // Ensure flash loan is repaid with fee
        uint256 balanceAfter = IERC20(token).balanceOf(address(this));
        uint256 requiredBalance = balanceBefore + fee;
        if (balanceAfter < requiredBalance) {
            revert FlashLoanNotRepaid(token, requiredBalance, balanceAfter);
        }

        _inFlashLoan = false;

        // Synchronize reserves without updating TWAP accumulators (neutralize oracle manipulation)
        _reserve0 = uint112(IERC20(token0).balanceOf(address(this)));
        _reserve1 = uint112(IERC20(token1).balanceOf(address(this)));

        emit FlashLoanExecuted(receiver, token, amount, fee);
        return true;
    }

    // =========================================================================
    // Admin Controls
    // =========================================================================

    function setMaxTwapDeviation(uint256 newDeviationBps) external onlyOwner {
        require(newDeviationBps <= 2500, "Deviation too high"); // max 25%
        emit MaxTwapDeviationUpdated(maxTwapDeviationBps, newDeviationBps);
        maxTwapDeviationBps = newDeviationBps;
    }

    function setMaxBlockPriceImpact(uint256 newImpactBps) external onlyOwner {
        require(newImpactBps <= 1500, "Impact too high"); // max 15%
        emit MaxBlockPriceImpactUpdated(maxBlockPriceImpactBps, newImpactBps);
        maxBlockPriceImpactBps = newImpactBps;
    }

    function setFlashLoanFee(uint256 newFeeBps) external onlyOwner {
        require(newFeeBps <= 100, "Fee too high"); // max 1%
        emit FlashLoanFeeUpdated(flashLoanFeeBps, newFeeBps);
        flashLoanFeeBps = newFeeBps;
    }

    function toggleFlashLoanProtection(bool enabled) external onlyOwner {
        flashLoanProtectionEnabled = enabled;
        emit FlashLoanProtectionToggled(enabled);
    }

    function resetTwapSnapshot() external onlyOwner {
        snapshotStart = TwapSnapshot({
            price0Cumulative: price0CumulativeLast,
            price1Cumulative: price1CumulativeLast,
            timestamp: uint32(block.timestamp)
        });
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }

    // =========================================================================
    // Internal Functions
    // =========================================================================

    function _updateReservesAndTwap(uint256 balance0, uint256 balance1) internal {
        require(balance0 <= type(uint112).max && balance1 <= type(uint112).max, "Reserve overflow");

        uint32 blockTimestamp = uint32(block.timestamp % 2**32);
        uint32 timeElapsed = blockTimestamp - blockTimestampLast;

        if (timeElapsed > 0 && _reserve0 != 0 && _reserve1 != 0 && !_inFlashLoan) {
            price0CumulativeLast += (uint256(_reserve1) * PRECISION / uint256(_reserve0)) * timeElapsed;
            price1CumulativeLast += (uint256(_reserve0) * PRECISION / uint256(_reserve1)) * timeElapsed;
            emit TwapUpdated(price0CumulativeLast, price1CumulativeLast, blockTimestamp);
        }

        _reserve0 = uint112(balance0);
        _reserve1 = uint112(balance1);
        blockTimestampLast = blockTimestamp;
    }

    function _getInstantaneousQuote(address tokenIn, uint256 amountIn) internal view returns (uint256) {
        if (_reserve0 == 0 || _reserve1 == 0) return 0;
        if (tokenIn == token0) {
            return (amountIn * uint256(_reserve1)) / uint256(_reserve0);
        } else {
            return (amountIn * uint256(_reserve0)) / uint256(_reserve1);
        }
    }
}
