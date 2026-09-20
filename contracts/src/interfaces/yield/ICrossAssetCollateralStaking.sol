// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title ICrossAssetCollateralStaking
 * @notice Interface for institutional cross-asset collateral staking and multi-token yield distribution.
 */
interface ICrossAssetCollateralStaking {
    enum LockDuration {
        FLEXIBLE, // 0 days, 1.00x multiplier
        DAYS_30,  // 30 days, 1.15x multiplier
        DAYS_90,  // 90 days, 1.35x multiplier
        DAYS_365  // 365 days, 1.75x multiplier
    }

    struct StakingAssetConfig {
        bool isSupported;
        uint16 weightBps;         // Relative weight (e.g. 10000 = 1.0x, 15000 = 1.5x)
        bool isSecurityToken;     // ERC-3643 token requiring KYC
        uint256 totalStaked;      // Total raw tokens staked
        uint256 totalWeightedStake;// Total weighted tokens staked
    }

    struct StakePosition {
        uint256 amount;
        uint256 weightedAmount;
        uint256 lockedUntil;
        LockDuration lockDuration;
        bool lockedForMargin;     // Staked collateral designated for margin trading
    }

    // =========================================================================
    // Events
    // =========================================================================

    event Staked(
        address indexed user,
        address indexed asset,
        uint256 amount,
        uint256 weightedAmount,
        LockDuration lockDuration,
        uint256 unlockTimestamp
    );

    event Unstaked(
        address indexed user,
        address indexed asset,
        uint256 amount,
        uint256 penaltyAmount
    );

    event RewardNotified(address indexed rewardToken, uint256 rewardAmount);
    event RewardPaid(address indexed user, address indexed rewardToken, uint256 reward);
    event MarginLockUpdated(address indexed user, address indexed asset, bool isLocked);
    event AssetConfigured(address indexed asset, uint16 weightBps, bool isSecurityToken);

    // =========================================================================
    // Errors
    // =========================================================================

    error AssetNotSupported(address asset);
    error RewardTokenNotSupported(address rewardToken);
    error PositionLockedForMargin();
    error LockDurationNotElapsed(uint256 currentTimestamp, uint256 unlockTimestamp);
    error InsufficientStakedBalance();
    error UserNotKYCVerified(address user);
    error InvalidZeroAddress();
    error InvalidAmount();
    error EarlyWithdrawalPenaltyFailed();

    // =========================================================================
    // View Functions
    // =========================================================================

    function totalEffectiveWeight() external view returns (uint256);
    function userEffectiveWeight(address user) external view returns (uint256);
    function earned(address user, address rewardToken) external view returns (uint256);
    function getUserPosition(address user, address asset) external view returns (
        uint256 amount,
        uint256 weightedAmount,
        uint256 lockedUntil,
        LockDuration lockDuration,
        bool lockedForMargin
    );

    // =========================================================================
    // State-Changing Functions
    // =========================================================================

    function stake(
        address asset,
        uint256 amount,
        LockDuration lockDuration
    ) external;

    function unstake(
        address asset,
        uint256 amount,
        bool forceEarlyWithdrawal
    ) external returns (uint256 netAmountReceived, uint256 penaltyAmount);

    function claimRewards(address[] calldata rewardTokens) external;

    function setMarginLock(address asset, bool isLocked) external;

    function notifyRewardAmount(address rewardToken, uint256 amount) external;
}
