// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {SoulboundTraderReputation} from "../../src/reputation/SoulboundTraderReputation.sol";
import {ISoulboundTraderReputation} from "../../src/interfaces/reputation/ISoulboundTraderReputation.sol";

contract SoulboundTraderReputationTest is Test {
    SoulboundTraderReputation public implementation;
    SoulboundTraderReputation public reputation;

    address public admin = address(0xAD01);
    address public issuer = address(0x1550);
    address public oracle = address(0x08AC);

    address public trader1 = address(0x1001);
    address public trader2 = address(0x1002);

    function setUp() public {
        implementation = new SoulboundTraderReputation();
        bytes memory initData = abi.encodeWithSelector(
            SoulboundTraderReputation.initialize.selector,
            admin,
            "Growww Soulbound Trader Reputation",
            "GSTR",
            "https://reputation.growww.in/metadata/"
        );

        ERC1967Proxy proxy = new ERC1967Proxy(address(implementation), initData);
        reputation = SoulboundTraderReputation(address(proxy));

        vm.startPrank(admin);
        reputation.grantRole(reputation.ISSUER_ROLE(), issuer);
        reputation.grantRole(reputation.ORACLE_ROLE(), oracle);
        vm.stopPrank();
    }

    function test_Initialization() public {
        assertTrue(reputation.hasRole(reputation.DEFAULT_ADMIN_ROLE(), admin));
        assertTrue(reputation.hasRole(reputation.ISSUER_ROLE(), issuer));
        assertTrue(reputation.hasRole(reputation.ORACLE_ROLE(), oracle));
        assertEq(reputation.name(), "Growww Soulbound Trader Reputation");
        assertEq(reputation.symbol(), "GSTR");
    }

    function test_MintSoulboundProfile() public {
        vm.prank(issuer);
        uint256 tokenId = reputation.mintProfile(trader1);

        assertEq(tokenId, 1);
        assertEq(reputation.ownerOf(tokenId), trader1);
        assertTrue(reputation.locked(tokenId));

        ISoulboundTraderReputation.TraderReputation memory profile = reputation.getTraderProfile(trader1);
        assertEq(profile.trader, trader1);
        assertEq(profile.currentStreak, 0);
        assertEq(profile.totalVolumeE8, 0);
        assertEq(uint8(profile.kycTier), uint8(ISoulboundTraderReputation.KycTier.NONE));
        assertEq(uint8(profile.referralTier), uint8(ISoulboundTraderReputation.ReferralTier.STANDARD));
    }

    function test_RevertOnDuplicateMint() public {
        vm.prank(issuer);
        reputation.mintProfile(trader1);

        vm.prank(issuer);
        vm.expectRevert(abi.encodeWithSelector(ISoulboundTraderReputation.ProfileAlreadyExists.selector, trader1));
        reputation.mintProfile(trader1);
    }

    function test_SoulboundTransferIsBlocked() public {
        vm.prank(issuer);
        uint256 tokenId = reputation.mintProfile(trader1);

        // Attempting to transfer must revert with SoulboundTokenLocked
        vm.prank(trader1);
        vm.expectRevert(abi.encodeWithSelector(ISoulboundTraderReputation.SoulboundTokenLocked.selector, tokenId));
        reputation.transferFrom(trader1, trader2, tokenId);

        vm.prank(trader1);
        vm.expectRevert(abi.encodeWithSelector(ISoulboundTraderReputation.SoulboundTokenLocked.selector, tokenId));
        reputation.safeTransferFrom(trader1, trader2, tokenId);
    }

    function test_TradingActivityAndStreakTracking() public {
        vm.prank(issuer);
        reputation.mintProfile(trader1);

        // Record trading activity
        uint256 tradeVolume = 50_000e8; // 50,000 turnover
        vm.prank(oracle);
        reputation.updateTradingActivity(trader1, tradeVolume, true);

        ISoulboundTraderReputation.TraderReputation memory profile = reputation.getTraderProfile(trader1);
        assertEq(profile.totalVolumeE8, tradeVolume);
        assertEq(profile.currentStreak, 1);
        assertEq(profile.maxStreak, 1);
        assertEq(profile.lastTradeTimestamp, block.timestamp);

        // Consecutive trading streak
        vm.warp(block.timestamp + 1 days);
        vm.prank(oracle);
        reputation.updateTradingActivity(trader1, 25_000e8, true);

        profile = reputation.getTraderProfile(trader1);
        assertEq(profile.totalVolumeE8, 75_000e8);
        assertEq(profile.currentStreak, 2);
        assertEq(profile.maxStreak, 2);
    }

    function test_KycAndReferralTierUpdates() public {
        vm.prank(issuer);
        reputation.mintProfile(trader1);

        // Upgrade to KYC Tier 2 (Verified Domestic)
        vm.prank(issuer);
        reputation.updateKycStatus(trader1, ISoulboundTraderReputation.KycTier.TIER_2_VERIFIED);

        // Upgrade to Gold referral tier
        vm.prank(issuer);
        reputation.updateReferralTier(trader1, ISoulboundTraderReputation.ReferralTier.GOLD);

        ISoulboundTraderReputation.TraderReputation memory profile = reputation.getTraderProfile(trader1);
        assertEq(uint8(profile.kycTier), uint8(ISoulboundTraderReputation.KycTier.TIER_2_VERIFIED));
        assertEq(uint8(profile.referralTier), uint8(ISoulboundTraderReputation.ReferralTier.GOLD));
        // Score: KYC Tier 2 (200) + Referral Tier 3 (150) = 350
        assertEq(reputation.getReputationScore(trader1), 350);
    }

    function test_AchievementUnlocking() public {
        vm.prank(issuer);
        reputation.mintProfile(trader1);

        // Unlock achievement 1: First Trade (50 bonus)
        vm.prank(issuer);
        reputation.unlockAchievement(trader1, 1);
        assertTrue(reputation.hasAchievement(trader1, 1));
        assertEq(reputation.getReputationScore(trader1), 50);

        // Unlock achievement 4: Master Trader (1000 bonus)
        vm.prank(issuer);
        reputation.unlockAchievement(trader1, 4);
        assertTrue(reputation.hasAchievement(trader1, 4));
        assertEq(reputation.getReputationScore(trader1), 1050);

        // Cannot claim same achievement twice
        vm.prank(issuer);
        vm.expectRevert(abi.encodeWithSelector(ISoulboundTraderReputation.AchievementAlreadyClaimed.selector, trader1, 1));
        reputation.unlockAchievement(trader1, 1);
    }

    function testFuzz_ReputationScoreMonotonicity(uint32 streak, uint32 volumeUnits) public {
        vm.assume(streak < 1000);
        vm.assume(volumeUnits < 1_000_000);

        vm.prank(issuer);
        reputation.mintProfile(trader1);

        uint256 volumeE8 = uint256(volumeUnits) * 1e8;
        for (uint32 i = 0; i < streak; i++) {
            vm.prank(oracle);
            reputation.updateTradingActivity(trader1, 1e8, true);
        }

        uint256 score = reputation.getReputationScore(trader1);
        assertEq(score, (uint256(streak) * 1) + (uint256(streak) * 10));
    }
}
