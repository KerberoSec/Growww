// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

import {ILiquidationEngine} from "../interfaces/liquidation/ILiquidationEngine.sol";
import {IIdentityRegistry} from "../interfaces/IIdentityRegistry.sol";

interface IPriceFeed {
    function getLatestPriceUSD() external view returns (uint256 priceUSD, uint8 decimals, uint256 updatedAt);
}

/**
 * @title LiquidationEngine
 * @notice Dynamic Collateral Ratio and Liquidation Engine for digital assets and tokenized securities.
 * @dev Adjusts collateral requirements dynamically based on market volatility, enforces institutional
 *      close factors, liquidation bonuses, bad debt absorption, and ERC-3643 KYC checks on liquidators.
 */
contract LiquidationEngine is ReentrancyGuard, Pausable, Ownable, ILiquidationEngine {
    using SafeERC20 for IERC20;

    // --- Constants ---
    uint256 public constant BPS_DENOMINATOR = 10000;
    uint256 public constant MIN_HEALTH_FACTOR_BPS = 10000; // 1.0 in BPS
    uint256 public constant SEVERELY_UNDERWATER_BPS = 9000; // 0.90 in BPS
    uint256 public constant MAX_CLOSE_FACTOR_BPS = 5000;    // 50% max close factor when 0.90 <= HF < 1.0

    // --- KYC Registry ---
    address public identityRegistry;

    // --- Supported Assets ---
    address[] public supportedCollateralAssets;
    address[] public supportedDebtAssets;

    mapping(address => CollateralConfig) public collateralConfigs;
    mapping(address => bool) public isDebtAssetSupported;
    mapping(address => address) public debtPriceFeeds;

    // User Position Balances: user => asset => balance
    mapping(address => mapping(address => uint256)) private _userCollateral;
    mapping(address => mapping(address => uint256)) private _userDebt;

    // Fallback price feeds for testing/simple setups
    mapping(address => uint256) public fallbackPricesUSD; // 1e8 precision

    constructor(
        address _identityRegistry,
        address initialOwner
    ) Ownable(initialOwner) {
        identityRegistry = _identityRegistry;
    }

    // =========================================================================
    // Collateral & Debt Management
    // =========================================================================

    function depositCollateral(
        address asset,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (!collateralConfigs[asset].isSupported) revert AssetNotSupported(asset);
        if (amount == 0) revert InvalidAmount();

        _userCollateral[msg.sender][asset] += amount;
        IERC20(asset).safeTransferFrom(msg.sender, address(this), amount);

        emit CollateralDeposited(msg.sender, asset, amount);
    }

    function withdrawCollateral(
        address asset,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (amount == 0) revert InvalidAmount();
        uint256 userBal = _userCollateral[msg.sender][asset];
        if (amount > userBal) revert InsufficientCollateral();

        _userCollateral[msg.sender][asset] = userBal - amount;

        // Check health factor after withdrawal if user has debt
        uint256 totalDebt = getUserTotalDebtValueUSD(msg.sender);
        if (totalDebt > 0) {
            uint256 hf = getHealthFactor(msg.sender);
            if (hf < MIN_HEALTH_FACTOR_BPS) {
                revert HealthFactorBelowMinimum(hf, MIN_HEALTH_FACTOR_BPS);
            }
        }

        IERC20(asset).safeTransfer(msg.sender, amount);
        emit CollateralWithdrawn(msg.sender, asset, amount);
    }

    function borrow(
        address debtAsset,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (!isDebtAssetSupported[debtAsset]) revert AssetNotSupported(debtAsset);
        if (amount == 0) revert InvalidAmount();

        _userDebt[msg.sender][debtAsset] += amount;

        uint256 hf = getHealthFactor(msg.sender);
        if (hf < MIN_HEALTH_FACTOR_BPS) {
            revert HealthFactorBelowMinimum(hf, MIN_HEALTH_FACTOR_BPS);
        }

        IERC20(debtAsset).safeTransfer(msg.sender, amount);
        emit DebtBorrowed(msg.sender, debtAsset, amount);
    }

    function repay(
        address debtAsset,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (amount == 0) revert InvalidAmount();
        uint256 currentDebt = _userDebt[msg.sender][debtAsset];
        uint256 repayAmount = amount > currentDebt ? currentDebt : amount;

        _userDebt[msg.sender][debtAsset] = currentDebt - repayAmount;
        IERC20(debtAsset).safeTransferFrom(msg.sender, address(this), repayAmount);

        emit DebtRepaid(msg.sender, debtAsset, repayAmount);
    }

    // =========================================================================
    // Liquidation Mechanism
    // =========================================================================

    function liquidatePosition(
        address borrower,
        address collateralAsset,
        address debtAsset,
        uint256 debtToCover
    ) external override nonReentrant whenNotPaused returns (uint256 collateralSeized) {
        if (borrower == address(0)) revert InvalidZeroAddress();
        if (!collateralConfigs[collateralAsset].isSupported) revert AssetNotSupported(collateralAsset);
        if (!isDebtAssetSupported[debtAsset]) revert AssetNotSupported(debtAsset);
        if (debtToCover == 0) revert InvalidAmount();

        // 1. Health factor check
        uint256 hf = getHealthFactor(borrower);
        if (hf >= MIN_HEALTH_FACTOR_BPS) {
            revert PositionHealthy(hf, MIN_HEALTH_FACTOR_BPS);
        }

        // 2. KYC check if collateral is an ERC-3643 digital security token
        if (collateralConfigs[collateralAsset].isSecurityToken && identityRegistry != address(0)) {
            IIdentityRegistry idReg = IIdentityRegistry(identityRegistry);
            if (!idReg.isVerified(msg.sender)) {
                revert LiquidatorNotKYCVerified(msg.sender);
            }
        }

        // 3. Close factor check
        uint256 userTotalDebt = _userDebt[borrower][debtAsset];
        if (userTotalDebt == 0) revert InvalidAmount();

        uint256 maxCoverable = userTotalDebt;
        if (hf >= SEVERELY_UNDERWATER_BPS) {
            // Moderately underwater (0.90 <= HF < 1.0): Max 50% can be liquidated
            maxCoverable = (userTotalDebt * MAX_CLOSE_FACTOR_BPS) / BPS_DENOMINATOR;
            if (maxCoverable == 0) maxCoverable = userTotalDebt;
        }

        if (debtToCover > maxCoverable) {
            revert ExceedsMaxCloseFactor(debtToCover, maxCoverable);
        }

        // 4. Calculate collateral to seize with liquidation bonus
        (uint256 debtPriceUSD, uint8 debtDecimals) = getAssetPriceUSD(debtAsset);
        (uint256 collPriceUSD, uint8 collDecimals) = getAssetPriceUSD(collateralAsset);

        // Debt value in 1e18 normalized USD
        uint256 debtValueUSD = (debtToCover * debtPriceUSD * 1e18) / ((10**debtDecimals) * 1e8);

        uint256 bonusBps = collateralConfigs[collateralAsset].liquidationBonusBps;
        uint256 collValueToSeizeUSD = (debtValueUSD * (BPS_DENOMINATOR + bonusBps)) / BPS_DENOMINATOR;

        // Collateral amount to seize
        collateralSeized = (collValueToSeizeUSD * (10**collDecimals) * 1e8) / (collPriceUSD * 1e18);

        uint256 borrowerCollateralBal = _userCollateral[borrower][collateralAsset];
        if (collateralSeized > borrowerCollateralBal) {
            // Bad debt scenario: collateral insufficient to cover full seized amount
            collateralSeized = borrowerCollateralBal;
            emit BadDebtAbsorbed(borrower, debtAsset, debtToCover);
        }

        // 5. Update state
        _userDebt[borrower][debtAsset] = userTotalDebt - debtToCover;
        _userCollateral[borrower][collateralAsset] = borrowerCollateralBal - collateralSeized;

        // 6. Execute transfers
        IERC20(debtAsset).safeTransferFrom(msg.sender, address(this), debtToCover);
        IERC20(collateralAsset).safeTransfer(msg.sender, collateralSeized);

        emit PositionLiquidated(
            borrower,
            msg.sender,
            collateralAsset,
            debtAsset,
            debtToCover,
            collateralSeized,
            bonusBps
        );
    }

    // =========================================================================
    // Dynamic Risk & Health Factor Calculations
    // =========================================================================

    function getDynamicLiquidationThreshold(
        address /*user*/,
        address asset
    ) public view override returns (uint256 thresholdBps) {
        CollateralConfig storage cfg = collateralConfigs[asset];
        if (!cfg.isSupported) return 0;

        // Dynamic adjustment: higher volatility index decreases threshold (requires more collateral)
        uint256 volPenalty = cfg.volatilityIndex / 2;
        if (volPenalty >= cfg.liquidationThresholdBps) {
            thresholdBps = 1000; // minimum floor 10%
        } else {
            thresholdBps = cfg.liquidationThresholdBps - volPenalty;
        }
    }

    function getDynamicLtv(address asset) public view returns (uint256 ltvBps) {
        CollateralConfig storage cfg = collateralConfigs[asset];
        if (!cfg.isSupported) return 0;

        uint256 volPenalty = cfg.volatilityIndex / 2;
        if (volPenalty >= cfg.baseLtvBps) {
            ltvBps = 1000;
        } else {
            ltvBps = cfg.baseLtvBps - volPenalty;
        }
    }

    function getHealthFactor(address user) public view override returns (uint256 healthFactorBps) {
        uint256 totalDebtUSD = getUserTotalDebtValueUSD(user);
        if (totalDebtUSD == 0) return type(uint256).max;

        uint256 totalAdjustedCollateralUSD = 0;
        for (uint256 i = 0; i < supportedCollateralAssets.length; i++) {
            address asset = supportedCollateralAssets[i];
            uint256 amount = _userCollateral[user][asset];
            if (amount > 0) {
                (uint256 priceUSD, uint8 decimals) = getAssetPriceUSD(asset);
                uint256 valUSD = (amount * priceUSD * 1e18) / ((10**decimals) * 1e8);
                uint256 thresholdBps = getDynamicLiquidationThreshold(user, asset);
                totalAdjustedCollateralUSD += (valUSD * thresholdBps) / BPS_DENOMINATOR;
            }
        }

        healthFactorBps = (totalAdjustedCollateralUSD * BPS_DENOMINATOR) / totalDebtUSD;
    }

    function getUserTotalCollateralValueUSD(address user) public view override returns (uint256 totalCollateralUSD) {
        for (uint256 i = 0; i < supportedCollateralAssets.length; i++) {
            address asset = supportedCollateralAssets[i];
            uint256 amount = _userCollateral[user][asset];
            if (amount > 0) {
                (uint256 priceUSD, uint8 decimals) = getAssetPriceUSD(asset);
                totalCollateralUSD += (amount * priceUSD * 1e18) / ((10**decimals) * 1e8);
            }
        }
    }

    function getUserTotalDebtValueUSD(address user) public view override returns (uint256 totalDebtUSD) {
        for (uint256 i = 0; i < supportedDebtAssets.length; i++) {
            address asset = supportedDebtAssets[i];
            uint256 debt = _userDebt[user][asset];
            if (debt > 0) {
                (uint256 priceUSD, uint8 decimals) = getAssetPriceUSD(asset);
                totalDebtUSD += (debt * priceUSD * 1e18) / ((10**decimals) * 1e8);
            }
        }
    }

    function getUserCollateral(address user, address asset) external view override returns (uint256) {
        return _userCollateral[user][asset];
    }

    function getUserDebt(address user, address asset) external view override returns (uint256) {
        return _userDebt[user][asset];
    }

    function getAssetPriceUSD(address asset) public view override returns (uint256 priceUSD, uint8 decimals) {
        address feed = collateralConfigs[asset].priceFeed != address(0)
            ? collateralConfigs[asset].priceFeed
            : debtPriceFeeds[asset];

        if (feed != address(0)) {
            (uint256 feedPrice, uint8 feedDecimals, ) = IPriceFeed(feed).getLatestPriceUSD();
            return (feedPrice, feedDecimals);
        }

        uint256 fallbackP = fallbackPricesUSD[asset];
        if (fallbackP == 0) revert OraclePriceStale(asset);

        uint8 dec = 18;
        try IERC20Metadata(asset).decimals() returns (uint8 tokenDec) {
            dec = tokenDec;
        } catch {}

        return (fallbackP, dec);
    }

    // =========================================================================
    // Admin Configurations
    // =========================================================================

    function configureCollateral(
        address asset,
        uint16 baseLtvBps,
        uint16 liquidationThresholdBps,
        uint16 liquidationBonusBps,
        uint16 volatilityIndex,
        bool isSecurityToken,
        address priceFeed
    ) external onlyOwner {
        if (asset == address(0)) revert InvalidZeroAddress();
        require(baseLtvBps < liquidationThresholdBps, "LTV must be < threshold");
        require(liquidationThresholdBps <= 9500, "Threshold must be <= 95%");

        if (!collateralConfigs[asset].isSupported) {
            supportedCollateralAssets.push(asset);
        }

        collateralConfigs[asset] = CollateralConfig({
            isSupported: true,
            baseLtvBps: baseLtvBps,
            liquidationThresholdBps: liquidationThresholdBps,
            liquidationBonusBps: liquidationBonusBps,
            volatilityIndex: volatilityIndex,
            isSecurityToken: isSecurityToken,
            priceFeed: priceFeed
        });

        emit CollateralConfigured(asset, baseLtvBps, liquidationThresholdBps);
    }

    function setDebtAsset(address asset, bool supported, address priceFeed) external onlyOwner {
        if (asset == address(0)) revert InvalidZeroAddress();
        if (supported && !isDebtAssetSupported[asset]) {
            supportedDebtAssets.push(asset);
        }
        isDebtAssetSupported[asset] = supported;
        debtPriceFeeds[asset] = priceFeed;
    }

    function setVolatilityIndex(address asset, uint16 newVol) external onlyOwner {
        if (!collateralConfigs[asset].isSupported) revert AssetNotSupported(asset);
        emit VolatilityUpdated(asset, collateralConfigs[asset].volatilityIndex, newVol);
        collateralConfigs[asset].volatilityIndex = newVol;
    }

    function setFallbackPrice(address asset, uint256 priceE8) external onlyOwner {
        fallbackPricesUSD[asset] = priceE8;
    }

    function setIdentityRegistry(address newRegistry) external onlyOwner {
        identityRegistry = newRegistry;
    }

    function pause() external onlyOwner {
        _pause();
    }

    function unpause() external onlyOwner {
        _unpause();
    }
}
