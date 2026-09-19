// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/demo/VirtualFaucet.sol";

contract VirtualFaucetTest is TestBase {
    VirtualFaucet public faucet;
    address public deployer = address(0xAA01);
    address public alice = address(0xA11CE);
    address public bob = address(0xB0B);

    function setUp() public {
        vm.prank(deployer);
        faucet = new VirtualFaucet();
    }

    function test_InitialState() public view {
        assertEq(faucet.owner(), deployer, "Deployer should be owner");
        assertEq(faucet.DEFAULT_USDT_GRANT(), 10_000 * 10**18, "USDT grant should be 10,000 * 10^18");
        assertEq(faucet.DEFAULT_BTC_GRANT(), 1 * 10**8, "BTC grant should be 1 * 10^8");
        assertEq(faucet.COOLDOWN_PERIOD(), 24 hours, "Cooldown should be 24 hours");
    }

    function test_ClaimDemoFunds_Success() public {
        vm.prank(alice);
        bool success = faucet.claimDemoFunds();
        assertTrue(success, "Claim should succeed");

        assertEq(faucet.lastDripTimestamp(alice), block.timestamp, "Drip timestamp recorded");
        assertEq(faucet.totalDrippedUsdt(alice), 10_000 * 10**18, "USDT credited");
        assertEq(faucet.totalDrippedBtc(alice), 1 * 10**8, "BTC credited");
    }

    function test_ClaimDemoFunds_RevertCooldown() public {
        vm.prank(alice);
        faucet.claimDemoFunds();

        // Attempting immediate second claim
        vm.prank(alice);
        try faucet.claimDemoFunds() {
            assertTrue(false, "Should revert within cooldown");
        } catch Error(string memory reason) {
            assertEq(reason, "Cooldown: Wait 24 hours between claims");
        }

        // Fast forward 12 hours (still in cooldown)
        vm.warp(block.timestamp + 12 hours);
        vm.prank(alice);
        try faucet.claimDemoFunds() {
            assertTrue(false, "Should revert after 12 hours");
        } catch Error(string memory reason) {
            assertEq(reason, "Cooldown: Wait 24 hours between claims");
        }
    }

    function test_ClaimDemoFunds_SuccessAfterCooldown() public {
        vm.prank(alice);
        faucet.claimDemoFunds();

        // Fast forward 24 hours
        vm.warp(block.timestamp + 24 hours);

        vm.prank(alice);
        faucet.claimDemoFunds();

        assertEq(faucet.totalDrippedUsdt(alice), 20_000 * 10**18, "Cumulative USDT should be 20,000");
        assertEq(faucet.totalDrippedBtc(alice), 2 * 10**8, "Cumulative BTC should be 2 BTC");
    }

    function test_AutoCreditUser_ByOwner() public {
        vm.prank(deployer);
        bool success = faucet.autoCreditUser(bob);
        assertTrue(success, "Auto-credit by owner should succeed");

        assertEq(faucet.lastDripTimestamp(bob), block.timestamp);
        assertEq(faucet.totalDrippedUsdt(bob), 10_000 * 10**18);
        assertEq(faucet.totalDrippedBtc(bob), 1 * 10**8);
    }

    function test_AutoCreditUser_RevertNonOwner() public {
        vm.prank(alice);
        try faucet.autoCreditUser(bob) {
            assertTrue(false, "Non-owner should not auto credit");
        } catch Error(string memory reason) {
            assertEq(reason, "Only owner");
        }
    }

    function test_AutoCreditUser_RevertAlreadyCredited() public {
        vm.prank(deployer);
        faucet.autoCreditUser(bob);

        // Attempting second auto credit
        vm.prank(deployer);
        try faucet.autoCreditUser(bob) {
            assertTrue(false, "Cannot auto-credit already credited account");
        } catch Error(string memory reason) {
            assertEq(reason, "Account already credited");
        }
    }

    function test_AutoCreditUser_RevertIfUserAlreadyClaimed() public {
        vm.prank(bob);
        faucet.claimDemoFunds();

        // Admin tries to auto credit after user claimed
        vm.prank(deployer);
        try faucet.autoCreditUser(bob) {
            assertTrue(false, "Cannot auto-credit user who already claimed");
        } catch Error(string memory reason) {
            assertEq(reason, "Account already credited");
        }
    }

    function test_CanClaim_FreshUser() public view {
        (bool eligible, uint256 remaining) = faucet.canClaim(alice);
        assertTrue(eligible, "Fresh user should be eligible");
        assertEq(remaining, 0, "No wait time for fresh user");
    }

    function test_CanClaim_DuringCooldown() public {
        vm.prank(alice);
        faucet.claimDemoFunds();

        (bool eligible, uint256 remaining) = faucet.canClaim(alice);
        assertFalse(eligible, "Should not be eligible during cooldown");
        assertEq(remaining, 24 hours, "Should have 24 hours remaining");

        vm.warp(block.timestamp + 6 hours);
        (bool eligibleAfter6h, uint256 remainingAfter6h) = faucet.canClaim(alice);
        assertFalse(eligibleAfter6h);
        assertEq(remainingAfter6h, 18 hours);

        vm.warp(block.timestamp + 18 hours); // Total 24 hours
        (bool eligibleAfter24h, uint256 remainingAfter24h) = faucet.canClaim(alice);
        assertTrue(eligibleAfter24h);
        assertEq(remainingAfter24h, 0);
    }
}
