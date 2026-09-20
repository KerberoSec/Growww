// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

import {ICrossAssetCollateralStaking} from "../interfaces/yield/ICrossAssetCollateralStaking.sol";
import {IIdentityRegistry} from "../interfaces/IIdentityRegistry.sol";

/**
 * @title CrossAssetCollateralStaking
 * @notice Institutional cross-asset collateral staking and multi-token yield distribution contract.
 * @dev Supports flexible and fixed-lock staking with yield boost multipliers, continuous dividend/yield
 *      distribution using $O(1)$ reward-per-token algorithms, margin collateral pledging, and ERC-3643 KYC checks.
 */
contract CrossAssetCollateralStaking is
    ReentrancyGuard,
    Pausable,
    Ownable,
    ICrossAssetCollateralStaking
{
    using SafeERC20 for IERC20;

    // --- Constants ---
    uint256 public constant BPS_DENOMINATOR = 10000;
    uint256 public constant PRECISION = 1e18;
    uint256 public constant EARLY_WITHDRAWAL_PENALTY_BPS = 1000; // 10% penalty for breaking fixed lock

    // Lock duration multipliers in BPS
    uint256 public constant MULTIPLIER_FLEXIBLE = 10000; // 1.00x
    uint256 public constant MULTIPLIER_30_DAYS = 11500;  // 1.15x
    uint256 public constant MULTIPLIER_90_DAYS = 13500;  // 1.35x
    uint256 public constant MULTIPLIER_365_DAYS = 17500; // 1.75x

    // --- Identity Registry ---
    address public identityRegistry;

    // --- Staking Asset Configurations ---
    address[] public supportedAssets;
    mapping(address => StakingAssetConfig) public assetConfigs;

    // --- Reward Token Configurations ---
    address[] public supportedRewardTokens;
    mapping(address => bool) public isRewardTokenSupported;
    mapping(address => uint256) public rewardPerTokenStored;
    mapping(address => mapping(address => uint256)) public userRewardPerTokenPaid; // user => rewardToken => paid
    mapping(address => mapping(address => uint256)) public rewards;                 // user => rewardToken => accrued

    // --- System Weights ---
    uint256 public override totalEffectiveWeight;
    mapping(address => uint256) public override userEffectiveWeight;

    // --- User Stake Positions: user => asset => position ---
    mapping(address => mapping(address => StakePosition)) public userPositions;

    // Fee Treasury for penalty routing
    address public feeTreasury;

    modifier updateReward(address account) {
        for (uint256 i = 0; i < supportedRewardTokens.length; i++) {
            address token = supportedRewardTokens[i];
            if (totalEffectiveWeight > 0) {
                // rewardPerToken is already accumulated via notifyRewardAmount
            }
            if (account != address(0)) {
                rewards[account][token] = earned(account, token);
                userRewardPerTokenPaid[account][token] = rewardPerTokenStored[token];
            }
        }
        _;
    }

    constructor(
        address _identityRegistry,
        address _feeTreasury,
        address initialOwner
    ) Ownable(initialOwner) {
        identityRegistry = _identityRegistry;
        feeTreasury = _feeTreasury == address(0) ? initialOwner : _feeTreasury;
    }

    // =========================================================================
    // View Functions
    // =========================================================================

    function getUserPosition(
        address user,
        address asset
    )
        external
        view
        override
        returns (
            uint256 amount,
            uint256 weightedAmount,
            uint256 lockedUntil,
            LockDuration lockDuration,
            bool lockedForMargin
        )
    {
        StakePosition storage pos = userPositions[user][asset];
        return (
            pos.amount,
            pos.weightedAmount,
            pos.lockedUntil,
            pos.lockDuration,
            pos.lockedForMargin
        );
    }

    function earned(address user, address rewardToken) public view override returns (uint256) {
        uint256 userWeight = userEffectiveWeight[user];
        uint256 rptDelta = rewardPerTokenStored[rewardToken] - userRewardPerTokenPaid[user][rewardToken];
        return ((userWeight * rptDelta) / PRECISION) + rewards[user][rewardToken];
    }

    function getLockMultiplier(LockDuration duration) public pure returns (uint256) {
        if (duration == LockDuration.DAYS_30) return MULTIPLIER_30_DAYS;
        if (duration == LockDuration.DAYS_90) return MULTIPLIER_90_DAYS;
        if (duration == LockDuration.DAYS_365) return MULTIPLIER_365_DAYS;
        return MULTIPLIER_FLEXIBLE;
    }

    function getLockDurationSeconds(LockDuration duration) public pure returns (uint256) {
        if (duration == LockDuration.DAYS_30) return 30 days;
        if (duration == LockDuration.DAYS_90) return 90 days;
        if (duration == LockDuration.DAYS_365) return 365 days;
        return 0;
    }

    // =========================================================================
    // Staking & Unstaking
    // =========================================================================

    function stake(
        address asset,
        uint256 amount,
        LockDuration lockDuration
    ) external override nonReentrant whenNotPaused updateReward(msg.sender) {
        StakingAssetConfig storage config = assetConfigs[asset];
        if (!config.isSupported) revert AssetNotSupported(asset);
        if (amount == 0) revert InvalidAmount();

        if (config.isSecurityToken && identityRegistry != address(0)) {
            if (!IIdentityRegistry(identityRegistry).isVerified(msg.sender)) {
                revert UserNotKYCVerified(msg.sender);
            }
        }

        uint256 lockMultiplier = getLockMultiplier(lockDuration);
        uint256 weightedAmount = (amount * config.weightBps * lockMultiplier) / (BPS_DENOMINATOR * BPS_DENOMINATOR);

        StakePosition storage pos = userPositions[msg.sender][asset];
        uint256 unlockTime = block.timestamp + getLockDurationSeconds(lockDuration);

        pos.amount += amount;
        pos.weightedAmount += weightedAmount;
        if (unlockTime > pos.lockedUntil) {
            pos.lockedUntil = unlockTime;
            pos.lockDuration = lockDuration;
        }

        config.totalStaked += amount;
        config.totalWeightedStake += weightedAmount;

        userEffectiveWeight[msg.sender] += weightedAmount;
        totalEffectiveWeight += weightedAmount;

        IERC20(asset).safeTransferFrom(msg.sender, address(this), amount);

        emit Staked(msg.sender, asset, amount, weightedAmount, lockDuration, pos.lockedUntil);
    }

    function unstake(
        address asset,
        uint256 amount,
        bool forceEarlyWithdrawal
    )
        external
        override
        nonReentrant
        whenNotPaused
        updateReward(msg.sender)
        returns (uint256 netAmountReceived, uint256 penaltyAmount)
    {
        StakePosition storage pos = userPositions[msg.sender][asset];
        if (amount == 0) revert InvalidAmount();
        if (amount > pos.amount) revert InsufficientStakedBalance();
        if (pos.lockedForMargin) revert PositionLockedForMargin();

        StakingAssetConfig storage config = assetConfigs[asset];

        // Check lock duration
        if (block.timestamp < pos.lockedUntil) {
            if (!forceEarlyWithdrawal) {
                revert LockDurationNotElapsed(block.timestamp, pos.lockedUntil);
            }
            // Apply early withdrawal penalty
            penaltyAmount = (amount * EARLY_WITHDRAWAL_PENALTY_BPS) / BPS_DENOMINATOR;
            netAmountReceived = amount - penaltyAmount;
        } else {
            netAmountReceived = amount;
            penaltyAmount = 0;
        }

        uint256 weightToRemove = (amount * pos.weightedAmount) / pos.amount;

        pos.amount -= amount;
        pos.weightedAmount -= weightToRemove;

        config.totalStaked -= amount;
        config.totalWeightedStake -= weightToRemove;

        userEffectiveWeight[msg.sender] -= weightToRemove;
        totalEffectiveWeight -= weightToRemove;

        if (penaltyAmount > 0 && feeTreasury != address(0)) {
            IERC20(asset).safeTransfer(feeTreasury, penaltyAmount);
        }
        IERC20(asset).safeTransfer(msg.sender, netAmountReceived);

        emit Unstaked(msg.sender, asset, amount, penaltyAmount);
    }

    function setMarginLock(address asset, bool isLocked) external override {
        StakePosition storage pos = userPositions[msg.sender][asset];
        if (pos.amount == 0) revert InsufficientStakedBalance();
        pos.lockedForMargin = isLocked;
        emit MarginLockUpdated(msg.sender, asset, isLocked);
    }

    // =========================================================================
    // Yield Distribution & Reward Claiming
    // =========================================================================

    function notifyRewardAmount(
        address rewardToken,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (!isRewardTokenSupported[rewardToken]) revert RewardTokenNotSupported(rewardToken);
        if (amount == 0) revert InvalidAmount();
        if (totalEffectiveWeight == 0) revert InvalidAmount();

        IERC20(rewardToken).safeTransferFrom(msg.sender, address(this), amount);

        rewardPerTokenStored[rewardToken] += (amount * PRECISION) / totalEffectiveWeight;

        emit RewardNotified(rewardToken, amount);
    }

    function claimRewards(
        address[] calldata rewardTokens
    ) external override nonReentrant whenNotPaused updateReward(msg.sender) {
        for (uint256 i = 0; i < rewardTokens.length; i++) {
            address token = rewardTokens[i];
            uint256 reward = rewards[msg.sender][token];
            if (reward > 0) {
                rewards[msg.sender][token] = 0;
                IERC20(token).safeTransfer(msg.sender, reward);
                emit RewardPaid(msg.sender, token, reward);
            }
        }
    }

    // =========================================================================
    // Admin Controls
    // =========================================================================

    function configureStakingAsset(
        address asset,
        uint16 weightBps,
        bool isSecurityToken
    ) external onlyOwner {
        if (asset == address(0)) revert InvalidZeroAddress();
        if (weightBps == 0) revert InvalidAmount();

        if (!assetConfigs[asset].isSupported) {
            supportedAssets.push(asset);
        }

        assetConfigs[asset].isSupported = true;
        assetConfigs[asset].weightBps = weightBps;
        assetConfigs[asset].isSecurityToken = isSecurityToken;

        emit AssetConfigured(asset, weightBps, isSecurityToken);
    }

    function addRewardToken(address rewardToken) external onlyOwner {
        if (rewardToken == address(0)) revert InvalidZeroAddress();
        if (!isRewardTokenSupported[rewardToken]) {
            isRewardTokenSupported[rewardToken] = true;
            supportedRewardTokens.push(rewardToken);
        }
    }

    function setFeeTreasury(address newTreasury) external onlyOwner {
        if (newTreasury == address(0)) revert InvalidZeroAddress();
        feeTreasury = newTreasury;
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
