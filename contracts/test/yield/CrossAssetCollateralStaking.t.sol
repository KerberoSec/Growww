// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

import {CrossAssetCollateralStaking} from "../../src/yield/CrossAssetCollateralStaking.sol";
import {ICrossAssetCollateralStaking} from "../../src/interfaces/yield/ICrossAssetCollateralStaking.sol";
import {IIdentityRegistry} from "../../src/interfaces/IIdentityRegistry.sol";

contract MockStakingToken is ERC20 {
    constructor(string memory name, string memory symbol) ERC20(name, symbol) {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract MockStakingKYC is IIdentityRegistry {
    mapping(address => bool) public verified;

    function setVerified(address user, bool status) external {
        verified[user] = status;
    }

    function isVerified(address userAddress) external view override returns (bool) {
        return verified[userAddress];
    }

    function registerIdentity(address, bytes32, uint16, uint8) external override {}
    function batchRegisterIdentity(address[] calldata, bytes32[] calldata, uint16[] calldata, uint8[] calldata) external override {}
    function deleteIdentity(address) external override {}
    function updateSanctionStatus(bytes32, bool) external override {}
    function updateCountry(address, uint16) external override {}
    function updateKycTier(address, uint8) external override {}
    function getInvestorClaim(address) external pure override returns (InvestorClaim memory) {
        return InvestorClaim(bytes32(0), 0, 0, false, 0);
    }
    function getInvestorCountry(address) external pure override returns (uint16) { return 0; }
    function getIdentityId(address) external pure override returns (bytes32) { return bytes32(0); }
    function contains(address) external pure override returns (bool) { return true; }
    function isSanctioned(bytes32) external pure override returns (bool) { return false; }
    function totalIdentities() external pure override returns (uint256) { return 0; }
}

contract CrossAssetCollateralStakingTest is Test {
    CrossAssetCollateralStaking public staking;
    MockStakingToken public tokenA;
    MockStakingToken public securityToken;
    MockStakingToken public usdcReward;
    MockStakingToken public einrReward;
    MockStakingKYC public kycRegistry;

    address public admin = address(0xAD01);
    address public treasury = address(0x7777);
    address public alice = address(0x1001);
    address public bob = address(0x1002);
    address public nonKycUser = address(0x9999);

    function setUp() public {
        kycRegistry = new MockStakingKYC();
        kycRegistry.setVerified(admin, true);
        kycRegistry.setVerified(alice, true);
        kycRegistry.setVerified(bob, true);

        vm.prank(admin);
        staking = new CrossAssetCollateralStaking(address(kycRegistry), treasury, admin);

        tokenA = new MockStakingToken("Collateral A", "COLA");
        securityToken = new MockStakingToken("RWA Debt Bond", "RWAB");
        usdcReward = new MockStakingToken("USDC Yield", "USDC");
        einrReward = new MockStakingToken("eINR Dividend", "eINR");

        // Configure Staking Assets:
        // tokenA: weight 1.0x (10000 bps), non-security
        // securityToken: weight 1.5x (15000 bps), security token
        vm.prank(admin);
        staking.configureStakingAsset(address(tokenA), 10000, false);
        vm.prank(admin);
        staking.configureStakingAsset(address(securityToken), 15000, true);

        // Add Reward Tokens
        vm.prank(admin);
        staking.addRewardToken(address(usdcReward));
        vm.prank(admin);
        staking.addRewardToken(address(einrReward));

        // Mint tokens
        tokenA.mint(alice, 10_000 ether);
        securityToken.mint(alice, 10_000 ether);
        tokenA.mint(bob, 10_000 ether);
        securityToken.mint(nonKycUser, 10_000 ether);

        usdcReward.mint(admin, 100_000 ether);
        einrReward.mint(admin, 100_000 ether);

        // Approvals
        vm.prank(alice);
        tokenA.approve(address(staking), type(uint256).max);
        vm.prank(alice);
        securityToken.approve(address(staking), type(uint256).max);

        vm.prank(bob);
        tokenA.approve(address(staking), type(uint256).max);

        vm.prank(nonKycUser);
        securityToken.approve(address(staking), type(uint256).max);

        vm.prank(admin);
        usdcReward.approve(address(staking), type(uint256).max);
        vm.prank(admin);
        einrReward.approve(address(staking), type(uint256).max);
    }

    function test_DeploymentAndConfiguration() public view {
        assertEq(staking.identityRegistry(), address(kycRegistry));
        assertEq(staking.feeTreasury(), treasury);
        assertTrue(staking.isRewardTokenSupported(address(usdcReward)));
        assertTrue(staking.isRewardTokenSupported(address(einrReward)));
    }

    function test_Stake_FlexibleAndLockDurationMultipliers() public {
        // Alice stakes 1,000 tokenA with FLEXIBLE (1.0x weight, 1.0x duration = 1,000 weight)
        vm.prank(alice);
        staking.stake(address(tokenA), 1_000 ether, ICrossAssetCollateralStaking.LockDuration.FLEXIBLE);

        assertEq(staking.userEffectiveWeight(alice), 1_000 ether);

        // Bob stakes 1,000 tokenA with 90 DAYS lock (1.0x weight * 1.35x duration = 1,350 weight)
        vm.prank(bob);
        staking.stake(address(tokenA), 1_000 ether, ICrossAssetCollateralStaking.LockDuration.DAYS_90);

        assertEq(staking.userEffectiveWeight(bob), 1_350 ether);
        assertEq(staking.totalEffectiveWeight(), 2_350 ether);
    }

    function test_YieldDistribution_ContinuousMultiTokenAccrualAndClaim() public {
        // Alice stakes 1,000 tokenA (1,000 weight)
        vm.prank(alice);
        staking.stake(address(tokenA), 1_000 ether, ICrossAssetCollateralStaking.LockDuration.FLEXIBLE);

        // Bob stakes 1,000 tokenA with 365 DAYS lock (1,000 * 1.75 = 1,750 weight)
        vm.prank(bob);
        staking.stake(address(tokenA), 1_000 ether, ICrossAssetCollateralStaking.LockDuration.DAYS_365);

        // Total weight = 2,750 ether
        assertEq(staking.totalEffectiveWeight(), 2_750 ether);

        // Admin notifies rewards: 2,750 USDC and 5,500 eINR
        vm.prank(admin);
        staking.notifyRewardAmount(address(usdcReward), 2_750 ether);
        vm.prank(admin);
        staking.notifyRewardAmount(address(einrReward), 5_500 ether);

        // Verify earned amounts:
        // Alice has 1000 / 2750 -> 1,000 USDC and 2,000 eINR
        // Bob has 1750 / 2750 -> 1,750 USDC and 3,500 eINR
        assertEq(staking.earned(alice, address(usdcReward)), 1_000 ether);
        assertEq(staking.earned(alice, address(einrReward)), 2_000 ether);
        assertEq(staking.earned(bob, address(usdcReward)), 1_750 ether);
        assertEq(staking.earned(bob, address(einrReward)), 3_500 ether);

        // Alice claims rewards
        address[] memory rewardTokens = new address[](2);
        rewardTokens[0] = address(usdcReward);
        rewardTokens[1] = address(einrReward);

        vm.prank(alice);
        staking.claimRewards(rewardTokens);

        assertEq(usdcReward.balanceOf(alice), 1_000 ether);
        assertEq(einrReward.balanceOf(alice), 2_000 ether);
        assertEq(staking.earned(alice, address(usdcReward)), 0);
    }

    function test_MarginLock_PreventsUnstaking() public {
        vm.prank(alice);
        staking.stake(address(tokenA), 1_000 ether, ICrossAssetCollateralStaking.LockDuration.FLEXIBLE);

        // Alice locks collateral for exchange margin
        vm.prank(alice);
        staking.setMarginLock(address(tokenA), true);

        // Unstake is blocked
        vm.prank(alice);
        vm.expectRevert(ICrossAssetCollateralStaking.PositionLockedForMargin.selector);
        staking.unstake(address(tokenA), 500 ether, false);

        // Release margin lock
        vm.prank(alice);
        staking.setMarginLock(address(tokenA), false);

        // Unstake now succeeds
        vm.prank(alice);
        staking.unstake(address(tokenA), 500 ether, false);
        assertEq(tokenA.balanceOf(alice), 9_500 ether);
    }

    function test_EarlyWithdrawalPenalty() public {
        // Alice stakes 1,000 tokenA with 30-day lock
        vm.prank(alice);
        staking.stake(address(tokenA), 1_000 ether, ICrossAssetCollateralStaking.LockDuration.DAYS_30);

        // Attempting normal unstake before 30 days reverts
        vm.prank(alice);
        vm.expectRevert();
        staking.unstake(address(tokenA), 1_000 ether, false);

        // Force early withdrawal applies 10% penalty
        uint256 treasuryBalBefore = tokenA.balanceOf(treasury);
        vm.prank(alice);
        (uint256 netReceived, uint256 penalty) = staking.unstake(address(tokenA), 1_000 ether, true);

        assertEq(penalty, 100 ether);     // 10% penalty
        assertEq(netReceived, 900 ether);  // 90% net received
        assertEq(tokenA.balanceOf(treasury), treasuryBalBefore + 100 ether);
    }

    function test_Unstake_AfterLockExpires_ZeroPenalty() public {
        vm.prank(alice);
        staking.stake(address(tokenA), 1_000 ether, ICrossAssetCollateralStaking.LockDuration.DAYS_30);

        // Fast forward 31 days
        vm.warp(block.timestamp + 31 days);

        vm.prank(alice);
        (uint256 netReceived, uint256 penalty) = staking.unstake(address(tokenA), 1_000 ether, false);

        assertEq(penalty, 0);
        assertEq(netReceived, 1_000 ether);
    }

    function test_SecurityTokenStaking_KYCEnforced() public {
        // Non-KYC user cannot stake security token
        vm.prank(nonKycUser);
        vm.expectRevert(
            abi.encodeWithSelector(
                ICrossAssetCollateralStaking.UserNotKYCVerified.selector,
                nonKycUser
            )
        );
        staking.stake(address(securityToken), 100 ether, ICrossAssetCollateralStaking.LockDuration.FLEXIBLE);

        // KYC-verified Alice can stake security token with 1.5x asset weight
        vm.prank(alice);
        staking.stake(address(securityToken), 100 ether, ICrossAssetCollateralStaking.LockDuration.FLEXIBLE);

        assertEq(staking.userEffectiveWeight(alice), 150 ether); // 100 * 1.5 = 150
    }
}
