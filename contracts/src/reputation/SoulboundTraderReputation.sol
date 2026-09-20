// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts-upgradeable/token/ERC721/ERC721Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts/utils/Strings.sol";
import "../interfaces/reputation/ISoulboundTraderReputation.sol";

/**
 * @title SoulboundTraderReputation
 * @notice ERC-5192 Soulbound Token smart contract recording verified trading streaks, KYC status, and referral tier badges.
 * @dev Non-transferable token standard preventing secondary market transfer of reputation claims.
 */
contract SoulboundTraderReputation is
    ERC721Upgradeable,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    ISoulboundTraderReputation
{
    using Strings for uint256;

    bytes32 public constant ISSUER_ROLE = keccak256("ISSUER_ROLE");
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    uint256 public nextTokenId;

    mapping(address => uint256) private _traderToTokenId;
    mapping(uint256 => TraderReputation) private _profiles;
    mapping(address => mapping(uint256 => bool)) private _unlockedAchievements;
    mapping(uint256 => Achievement) public achievements;
    uint256 public totalAchievementsCount;

    string private _baseTokenURI;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address admin,
        string memory name_,
        string memory symbol_,
        string memory baseURI_
    ) external initializer {
        __ERC721_init(name_, symbol_);
        __AccessControl_init();
        __UUPSUpgradeable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(ISSUER_ROLE, admin);
        _grantRole(ORACLE_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);

        _baseTokenURI = baseURI_;

        // Seed default achievements
        _registerAchievement(1, "First Trade", "Executed first institutional trade", "ipfs://badge-1", 50);
        _registerAchievement(2, "7-Day Streak", "Traded 7 consecutive days", "ipfs://badge-2", 150);
        _registerAchievement(3, "Whale Trader", "Achieved over 1,000,000 turnover", "ipfs://badge-3", 500);
        _registerAchievement(4, "Master Trader", "Maintained top-tier win rate", "ipfs://badge-4", 1000);
        _registerAchievement(5, "Institutional Pioneer", "Completed Tier-3 KYC and audit onboarding", "ipfs://badge-5", 750);
    }

    function registerAchievement(
        uint256 id,
        string memory name,
        string memory description,
        string memory iconUri,
        uint256 scoreBonus
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _registerAchievement(id, name, description, iconUri, scoreBonus);
    }

    function _registerAchievement(
        uint256 id,
        string memory name,
        string memory description,
        string memory iconUri,
        uint256 scoreBonus
    ) internal {
        achievements[id] = Achievement({
            id: id,
            name: name,
            description: description,
            iconUri: iconUri,
            scoreBonus: scoreBonus
        });
        if (id > totalAchievementsCount) {
            totalAchievementsCount = id;
        }
    }

    function setBaseURI(string memory baseURI_) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _baseTokenURI = baseURI_;
    }

    function locked(uint256 tokenId) external view override returns (bool) {
        _requireOwned(tokenId);
        return true;
    }

    function mintProfile(address trader) external override returns (uint256 tokenId) {
        if (!hasRole(ISSUER_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }
        if (trader == address(0)) revert InvalidZeroAddress();
        if (_traderToTokenId[trader] != 0) revert ProfileAlreadyExists(trader);

        tokenId = ++nextTokenId;
        _traderToTokenId[trader] = tokenId;

        _profiles[tokenId] = TraderReputation({
            tokenId: tokenId,
            trader: trader,
            currentStreak: 0,
            maxStreak: 0,
            lastTradeTimestamp: 0,
            totalVolumeE8: 0,
            kycTier: KycTier.NONE,
            referralTier: ReferralTier.STANDARD,
            reputationScore: 0,
            registeredTimestamp: block.timestamp
        });

        _mint(trader, tokenId);
        emit Locked(tokenId);
        emit ProfileMinted(trader, tokenId);
    }

    function updateTradingActivity(
        address trader,
        uint256 volumeE8,
        bool incrementStreak
    ) external override {
        if (!hasRole(ISSUER_ROLE, msg.sender) && !hasRole(ORACLE_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }
        uint256 tokenId = _traderToTokenId[trader];
        if (tokenId == 0) revert ProfileDoesNotExist(trader);

        TraderReputation storage profile = _profiles[tokenId];
        profile.totalVolumeE8 += volumeE8;
        profile.lastTradeTimestamp = uint64(block.timestamp);

        if (incrementStreak) {
            profile.currentStreak += 1;
            if (profile.currentStreak > profile.maxStreak) {
                profile.maxStreak = profile.currentStreak;
            }
        }

        profile.reputationScore = _calculateReputationScore(profile);

        emit VolumeAttested(trader, volumeE8, profile.totalVolumeE8);
        emit TradingStreakUpdated(trader, profile.currentStreak, profile.maxStreak, block.timestamp);
        emit ReputationScoreRecalculated(trader, profile.reputationScore);
    }

    function updateKycStatus(address trader, KycTier tier) external override {
        if (!hasRole(ISSUER_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }
        uint256 tokenId = _traderToTokenId[trader];
        if (tokenId == 0) revert ProfileDoesNotExist(trader);

        TraderReputation storage profile = _profiles[tokenId];
        profile.kycTier = tier;
        profile.reputationScore = _calculateReputationScore(profile);

        emit KycTierUpdated(trader, uint8(tier));
        emit ReputationScoreRecalculated(trader, profile.reputationScore);
    }

    function updateReferralTier(address trader, ReferralTier tier) external override {
        if (!hasRole(ISSUER_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }
        uint256 tokenId = _traderToTokenId[trader];
        if (tokenId == 0) revert ProfileDoesNotExist(trader);

        TraderReputation storage profile = _profiles[tokenId];
        profile.referralTier = tier;
        profile.reputationScore = _calculateReputationScore(profile);

        emit ReferralTierUpdated(trader, uint8(tier));
        emit ReputationScoreRecalculated(trader, profile.reputationScore);
    }

    function unlockAchievement(address trader, uint256 achievementId) external override {
        if (!hasRole(ISSUER_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }
        uint256 tokenId = _traderToTokenId[trader];
        if (tokenId == 0) revert ProfileDoesNotExist(trader);
        if (_unlockedAchievements[trader][achievementId]) {
            revert AchievementAlreadyClaimed(trader, achievementId);
        }

        _unlockedAchievements[trader][achievementId] = true;
        TraderReputation storage profile = _profiles[tokenId];
        profile.reputationScore += achievements[achievementId].scoreBonus;

        emit AchievementUnlocked(trader, achievementId, achievements[achievementId].name);
        emit ReputationScoreRecalculated(trader, profile.reputationScore);
    }

    function getTraderProfile(address trader) external view override returns (TraderReputation memory) {
        uint256 tokenId = _traderToTokenId[trader];
        if (tokenId == 0) revert ProfileDoesNotExist(trader);
        return _profiles[tokenId];
    }

    function getTraderProfileByTokenId(uint256 tokenId) external view override returns (TraderReputation memory) {
        _requireOwned(tokenId);
        return _profiles[tokenId];
    }

    function hasAchievement(address trader, uint256 achievementId) external view override returns (bool) {
        return _unlockedAchievements[trader][achievementId];
    }

    function getReputationScore(address trader) external view override returns (uint256) {
        uint256 tokenId = _traderToTokenId[trader];
        if (tokenId == 0) return 0;
        return _profiles[tokenId].reputationScore;
    }

    function _calculateReputationScore(TraderReputation storage profile) internal view returns (uint256) {
        uint256 baseScore = (profile.totalVolumeE8 / 1e8) * 1;
        baseScore += uint256(profile.currentStreak) * 10;
        baseScore += uint256(profile.kycTier) * 100;
        baseScore += uint256(profile.referralTier) * 50;

        for (uint256 i = 1; i <= totalAchievementsCount; i++) {
            if (_unlockedAchievements[profile.trader][i]) {
                baseScore += achievements[i].scoreBonus;
            }
        }
        return baseScore;
    }

    function _update(
        address to,
        uint256 tokenId,
        address auth
    ) internal override returns (address) {
        address from = _ownerOf(tokenId);
        // Soulbound: disallow any transfer after initial minting
        if (from != address(0) && to != address(0)) {
            revert SoulboundTokenLocked(tokenId);
        }
        return super._update(to, tokenId, auth);
    }

    function _baseURI() internal view override returns (string memory) {
        return _baseTokenURI;
    }

    function supportsInterface(
        bytes4 interfaceId
    ) public view override(ERC721Upgradeable, AccessControlUpgradeable, IERC165) returns (bool) {
        return interfaceId == type(ISoulboundTraderReputation).interfaceId || super.supportsInterface(interfaceId);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    uint256[45] private __gap;
}
