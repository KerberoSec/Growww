// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts/token/ERC721/IERC721.sol";

/**
 * @title ISoulboundTraderReputation
 * @notice Interface for ERC-5192 Soulbound Trader Reputation and Achievement Badge NFTs.
 */
interface ISoulboundTraderReputation is IERC721 {
    // ERC-5192 Soulbound Events
    event Locked(uint256 tokenId);
    event Unlocked(uint256 tokenId);

    // Reputation and Achievement Events
    event ProfileMinted(address indexed trader, uint256 indexed tokenId);
    event TradingStreakUpdated(address indexed trader, uint256 currentStreak, uint256 maxStreak, uint256 lastTradeTimestamp);
    event VolumeAttested(address indexed trader, uint256 additionalVolume, uint256 totalVolume);
    event KycTierUpdated(address indexed trader, uint8 kycTier);
    event ReferralTierUpdated(address indexed trader, uint8 referralTier);
    event AchievementUnlocked(address indexed trader, uint256 indexed achievementId, string achievementName);
    event ReputationScoreRecalculated(address indexed trader, uint256 newScore);

    // Custom Errors
    error SoulboundTokenLocked(uint256 tokenId);
    error ProfileAlreadyExists(address trader);
    error ProfileDoesNotExist(address trader);
    error InvalidZeroAddress();
    error UnauthorizedCaller(address caller);
    error AchievementAlreadyClaimed(address trader, uint256 achievementId);

    enum KycTier {
        NONE,
        TIER_1_BASIC,
        TIER_2_VERIFIED,
        TIER_3_INSTITUTIONAL
    }

    enum ReferralTier {
        STANDARD,
        BRONZE,
        SILVER,
        GOLD,
        PLATINUM,
        DIAMOND
    }

    struct TraderReputation {
        uint256 tokenId;
        address trader;
        uint32 currentStreak;
        uint32 maxStreak;
        uint64 lastTradeTimestamp;
        uint256 totalVolumeE8;
        KycTier kycTier;
        ReferralTier referralTier;
        uint256 reputationScore;
        uint256 registeredTimestamp;
    }

    struct Achievement {
        uint256 id;
        string name;
        string description;
        string iconUri;
        uint256 scoreBonus;
    }

    // ERC-5192 View
    function locked(uint256 tokenId) external view returns (bool);

    // Core Management
    function mintProfile(address trader) external returns (uint256 tokenId);
    function updateTradingActivity(address trader, uint256 volumeE8, bool incrementStreak) external;
    function updateKycStatus(address trader, KycTier tier) external;
    function updateReferralTier(address trader, ReferralTier tier) external;
    function unlockAchievement(address trader, uint256 achievementId) external;

    // View Functions
    function getTraderProfile(address trader) external view returns (TraderReputation memory);
    function getTraderProfileByTokenId(uint256 tokenId) external view returns (TraderReputation memory);
    function hasAchievement(address trader, uint256 achievementId) external view returns (bool);
    function getReputationScore(address trader) external view returns (uint256);
}
