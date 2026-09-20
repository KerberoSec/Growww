// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface ICopyTradingVault {
    struct VaultConfig {
        address masterTrader;
        address assetToken;
        uint256 performanceFeeBps; // 1000 - 2000 (10% - 20%)
        uint256 maxCapacity;
        uint256 minDeposit;
        uint256 lockupPeriod;
    }

    struct FollowerPosition {
        uint256 shares;
        uint256 depositedAmount;
        uint256 depositTimestamp;
        uint256 highWaterMark;
    }

    // Events
    event FollowerDeposited(address indexed follower, uint256 assets, uint256 sharesMinted);
    event FollowerWithdrawn(address indexed follower, uint256 sharesBurned, uint256 assetsReturned);
    event PerformanceFeeHarvested(
        address indexed masterTrader,
        uint256 grossProfit,
        uint256 feeAmount,
        uint256 newHighWaterMark
    );
    event VaultStrategyExecuted(address indexed target, uint256 amount, bytes data);
    event StrategyFundsReturned(address indexed strategy, uint256 amountReturned);
    event HighWaterMarkUpdated(uint256 oldHWM, uint256 newHWM);
    event MasterTraderUpdated(address indexed oldMaster, address indexed newMaster);
    event PerformanceFeeBpsUpdated(uint256 oldBps, uint256 newBps);

    // Custom Errors
    error InvalidPerformanceFeeBps(uint256 bps);
    error DepositBelowMinimum(uint256 provided, uint256 minimum);
    error CapacityExceeded(uint256 requestedTotal, uint256 maxCapacity);
    error LockupActive(uint256 unlockTimestamp, uint256 currentTimestamp);
    error InsufficientShares(uint256 requested, uint256 available);
    error UnauthorizedMasterTrader(address caller);
    error NoProfitAboveHighWaterMark(uint256 currentPrice, uint256 highWaterMark);
    error InvalidZeroAddress();
    error TargetNotApproved(address target);

    function deposit(uint256 assets, address receiver) external returns (uint256 shares);
    function withdraw(uint256 shares, address receiver) external returns (uint256 assets);
    function harvestPerformanceFee() external returns (uint256 feeShares);
    function executeStrategy(address target, uint256 amount, bytes calldata data) external returns (bytes memory result);
    function returnStrategyFunds(uint256 amount) external;

    function totalAssets() external view returns (uint256);
    function totalShares() external view returns (uint256);
    function masterTrader() external view returns (address);
    function convertToShares(uint256 assets) external view returns (uint256);
    function convertToAssets(uint256 shares) external view returns (uint256);
    function getFollowerPosition(address follower) external view returns (FollowerPosition memory);
    function currentSharePrice() external view returns (uint256);
    function highWaterMark() external view returns (uint256);
}
