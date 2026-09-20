// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title ILiquidationEngine
 * @notice Interface for dynamic collateral ratio management and liquidation engine.
 */
interface ILiquidationEngine {
    struct CollateralConfig {
        bool isSupported;
        uint16 baseLtvBps;              // e.g. 7500 = 75% LTV (equivalent to 133% min collateral ratio)
        uint16 liquidationThresholdBps;// e.g. 8000 = 80% (125% collateral ratio)
        uint16 liquidationBonusBps;    // e.g. 500 = 5% incentive for liquidator
        uint16 volatilityIndex;        // 0 to 1000 scale, dynamically bumps required collateral
        bool isSecurityToken;          // requires KYC to seize/receive
        address priceFeed;             // Oracle price feed address
    }

    struct UserPosition {
        mapping(address => uint256) collateralBalances;
        mapping(address => uint256) debtBalances;
    }

    // =========================================================================
    // Events
    // =========================================================================

    event CollateralDeposited(address indexed user, address indexed asset, uint256 amount);
    event CollateralWithdrawn(address indexed user, address indexed asset, uint256 amount);
    event DebtBorrowed(address indexed user, address indexed asset, uint256 amount);
    event DebtRepaid(address indexed user, address indexed asset, uint256 amount);
    event PositionLiquidated(
        address indexed borrower,
        address indexed liquidator,
        address collateralAsset,
        address debtAsset,
        uint256 debtCovered,
        uint256 collateralSeized,
        uint256 liquidationBonus
    );
    event VolatilityUpdated(address indexed asset, uint16 oldVol, uint16 newVol);
    event CollateralConfigured(address indexed asset, uint16 baseLtvBps, uint16 liquidationThresholdBps);
    event BadDebtAbsorbed(address indexed borrower, address indexed debtAsset, uint256 badDebtAmount);

    // =========================================================================
    // Errors
    // =========================================================================

    error PositionHealthy(uint256 healthFactor, uint256 minHealthFactor);
    error HealthFactorBelowMinimum(uint256 healthFactor, uint256 requiredHealthFactor);
    error ExceedsMaxCloseFactor(uint256 debtToCover, uint256 maxDebtToCover);
    error InsufficientCollateral();
    error AssetNotSupported(address asset);
    error InvalidZeroAddress();
    error InvalidAmount();
    error OraclePriceStale(address asset);
    error LiquidatorNotKYCVerified(address liquidator);
    error ReentrancyProhibited();

    // =========================================================================
    // View Functions
    // =========================================================================

    function getHealthFactor(address user) external view returns (uint256 healthFactorBps);
    function getDynamicLiquidationThreshold(address user, address asset) external view returns (uint256 thresholdBps);
    function getUserTotalCollateralValueUSD(address user) external view returns (uint256 totalCollateralUSD);
    function getUserTotalDebtValueUSD(address user) external view returns (uint256 totalDebtUSD);
    function getUserCollateral(address user, address asset) external view returns (uint256);
    function getUserDebt(address user, address asset) external view returns (uint256);
    function getAssetPriceUSD(address asset) external view returns (uint256 priceUSD, uint8 decimals);

    // =========================================================================
    // State-Changing Functions
    // =========================================================================

    function depositCollateral(address asset, uint256 amount) external;
    function withdrawCollateral(address asset, uint256 amount) external;
    function borrow(address debtAsset, uint256 amount) external;
    function repay(address debtAsset, uint256 amount) external;

    function liquidatePosition(
        address borrower,
        address collateralAsset,
        address debtAsset,
        uint256 debtToCover
    ) external returns (uint256 collateralSeized);
}
