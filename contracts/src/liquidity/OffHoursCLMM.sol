// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./ICLMMExchange.sol";

/**
 * @title OffHoursCLMM
 * @notice Concentrated Liquidity Market Maker Off-Hours AMM for tokenized equities.
 * @dev Deployed on Hyperledger Besu to enable capital-efficient overnight trading with hard statutory price collars.
 */
contract OffHoursCLMM is ICLMMExchange, Ownable, ReentrancyGuard {
    using SafeERC20 for IERC20;

    IERC20 public immutable token0;
    IERC20 public immutable token1;
    int24 public immutable tickSpacing;

    Slot0 public override slot0;
    CollarConfig public collarConfig;
    FeeConfig public feeConfig;

    uint128 public liquidity;
    bool public isOffHoursActive;
    bool public manualOverride;

    // Oracle authorized updater address
    address public oracleFeeder;

    modifier onlyOracleOrOwner() {
        if (msg.sender != owner() && msg.sender != oracleFeeder) {
            revert UnauthorizedCaller(msg.sender);
        }
        _;
    }

    constructor(
        address _token0,
        address _token1,
        int24 _tickSpacing,
        uint160 initialSqrtPriceX96,
        address initialOwner
    ) Ownable(initialOwner) {
        require(_token0 != address(0) && _token1 != address(0), "Invalid tokens");
        token0 = IERC20(_token0);
        token1 = IERC20(_token1);
        tickSpacing = _tickSpacing;

        slot0 = Slot0({
            sqrtPriceX96: initialSqrtPriceX96,
            tick: 0,
            observationIndex: 0,
            observationCardinality: 60,
            observationCardinalityNext: 60,
            feeBps: 1500, // 15 bps default
            unlocked: true
        });

        feeConfig = FeeConfig({
            baseFeeBps: 1500,
            maxFeeBps: 10000,
            collarProximityBufferTicks: 50,
            velocityScaleFactor: 2
        });

        // Initialize collar at initial price with 500 bps (5%) band
        uint160 minP = (initialSqrtPriceX96 * 95) / 100;
        uint160 maxP = (initialSqrtPriceX96 * 105) / 100;
        collarConfig = CollarConfig({
            primaryClosingPriceX96: initialSqrtPriceX96,
            collarBandBps: 500,
            minCollarPriceX96: minP,
            maxCollarPriceX96: maxP,
            minCollarTick: -500,
            maxCollarTick: 500,
            lastOracleUpdateTimestamp: uint64(block.timestamp),
            isCollarEnforced: true
        });

        isOffHoursActive = true; // Enabled by default for off-hours
    }

    function setOracleFeeder(address _feeder) external onlyOwner {
        oracleFeeder = _feeder;
    }

    function setOffHoursActive(bool _active, bool _manual) external onlyOwner {
        isOffHoursActive = _active;
        manualOverride = _manual;
        emit TradingSessionStateChanged(_active, _manual, block.timestamp);
    }

    function updatePrimaryClosingPrice(uint160 newClosingPriceX96) external onlyOracleOrOwner {
        uint160 oldPrice = collarConfig.primaryClosingPriceX96;
        uint160 delta = (newClosingPriceX96 * uint160(collarConfig.collarBandBps)) / 10000;
        uint160 minP = newClosingPriceX96 - delta;
        uint160 maxP = newClosingPriceX96 + delta;

        collarConfig.primaryClosingPriceX96 = newClosingPriceX96;
        collarConfig.minCollarPriceX96 = minP;
        collarConfig.maxCollarPriceX96 = maxP;
        collarConfig.lastOracleUpdateTimestamp = uint64(block.timestamp);

        emit PrimaryClosingPriceUpdated(
            oldPrice,
            newClosingPriceX96,
            minP,
            maxP,
            collarConfig.minCollarTick,
            collarConfig.maxCollarTick,
            block.timestamp
        );
    }

    function updateCollarBand(uint16 newBandBps) external onlyOwner {
        uint16 oldBand = collarConfig.collarBandBps;
        collarConfig.collarBandBps = newBandBps;
        uint160 delta = (collarConfig.primaryClosingPriceX96 * uint160(newBandBps)) / 10000;
        collarConfig.minCollarPriceX96 = collarConfig.primaryClosingPriceX96 - delta;
        collarConfig.maxCollarPriceX96 = collarConfig.primaryClosingPriceX96 + delta;

        emit CollarBandUpdated(oldBand, newBandBps);
    }

    function getCollarConfig() external view override returns (CollarConfig memory) {
        return collarConfig;
    }

    function addLiquidity(uint128 amount) external onlyOwner {
        liquidity += amount;
    }

    /**
     * @notice Executes an atomic token swap across concentrated liquidity ticks
     */
    function swap(SwapParams calldata params) external override nonReentrant returns (int256 amount0, int256 amount1) {
        if (!isOffHoursActive) {
            revert OutsideTradingHours();
        }

        require(params.amountSpecified != 0, "Amount specified cannot be zero");

        uint160 currentPrice = slot0.sqrtPriceX96;
        uint24 currentFee = feeConfig.baseFeeBps;

        // Determine direction and calculate output with price impact simulation
        if (params.zeroForOne) {
            // Selling token0 for token1 -> Price drops
            uint256 inputAmount = uint256(params.amountSpecified > 0 ? params.amountSpecified : -params.amountSpecified);
            uint256 feeAmount = (inputAmount * currentFee) / 1000000;
            uint256 inputNet = inputAmount - feeAmount;

            // Simplified price delta for mock/pool
            uint160 priceImpact = uint160((inputNet * 1e8) / (liquidity > 0 ? liquidity : 1e18));
            uint160 targetPrice = currentPrice > priceImpact ? currentPrice - priceImpact : 1;

            if (collarConfig.isCollarEnforced && targetPrice < collarConfig.minCollarPriceX96) {
                revert CollarBreached(targetPrice, collarConfig.minCollarPriceX96);
            }

            if (params.sqrtPriceLimitX96 != 0 && targetPrice < params.sqrtPriceLimitX96) {
                revert SlippageLimitExceeded(targetPrice, params.sqrtPriceLimitX96);
            }

            slot0.sqrtPriceX96 = targetPrice;
            amount0 = int256(inputAmount);
            amount1 = -int256(inputNet); // Output token1

            token0.safeTransferFrom(msg.sender, address(this), inputAmount);
            token1.safeTransfer(params.recipient, uint256(-amount1));
        } else {
            // Selling token1 for token0 -> Price increases
            uint256 inputAmount = uint256(params.amountSpecified > 0 ? params.amountSpecified : -params.amountSpecified);
            uint256 feeAmount = (inputAmount * currentFee) / 1000000;
            uint256 inputNet = inputAmount - feeAmount;

            uint160 priceImpact = uint160((inputNet * 1e8) / (liquidity > 0 ? liquidity : 1e18));
            uint160 targetPrice = currentPrice + priceImpact;

            if (collarConfig.isCollarEnforced && targetPrice > collarConfig.maxCollarPriceX96) {
                revert CollarBreached(targetPrice, collarConfig.maxCollarPriceX96);
            }

            if (params.sqrtPriceLimitX96 != 0 && targetPrice > params.sqrtPriceLimitX96) {
                revert SlippageLimitExceeded(targetPrice, params.sqrtPriceLimitX96);
            }

            slot0.sqrtPriceX96 = targetPrice;
            amount1 = int256(inputAmount);
            amount0 = -int256(inputNet); // Output token0

            token1.safeTransferFrom(msg.sender, address(this), inputAmount);
            token0.safeTransfer(params.recipient, uint256(-amount0));
        }

        emit Swap(
            msg.sender,
            params.recipient,
            amount0,
            amount1,
            slot0.sqrtPriceX96,
            liquidity,
            slot0.tick,
            currentFee
        );
    }
}
