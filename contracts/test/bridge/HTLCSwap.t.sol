// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "../../src/bridge/HTLCSwap.sol";

contract HTLCSwapTest is Test {
    HTLCSwap public htlc;

    address payable public alice = payable(address(0xA11CE));
    address payable public bob = payable(address(0xB0B));

    bytes public secretPreimage = "secret_preimage_for_atomic_swap_123";
    bytes32 public hashLock;
    bytes32 public swapId = keccak256("SWAP_ORDER_001");
    uint256 public swapAmount = 5 ether;
    uint256 public duration = 2 hours;

    function setUp() public {
        htlc = new HTLCSwap();
        hashLock = sha256(secretPreimage);
        vm.deal(alice, 10 ether);
        vm.deal(bob, 1 ether);
    }

    function test_InitiateSwap_Success() public {
        vm.prank(alice);
        htlc.initiateSwap{value: swapAmount}(swapId, hashLock, duration, bob);

        (
            bytes32 storedHash,
            uint256 timelock,
            uint256 val,
            address sender,
            address receiver,
            bool withdrawn,
            bool refunded,

        ) = htlc.swaps(swapId);

        assertEq(storedHash, hashLock);
        assertEq(timelock, block.timestamp + duration);
        assertEq(val, swapAmount);
        assertEq(sender, alice);
        assertEq(receiver, bob);
        assertFalse(withdrawn);
        assertFalse(refunded);
    }

    function test_InitiateSwap_RevertZeroValue() public {
        vm.prank(alice);
        vm.expectRevert("No funds provided");
        htlc.initiateSwap{value: 0}(swapId, hashLock, duration, bob);
    }

    function test_InitiateSwap_RevertTooShortDuration() public {
        vm.prank(alice);
        vm.expectRevert("Timelock duration too short");
        htlc.initiateSwap{value: swapAmount}(swapId, hashLock, 30 minutes, bob);
    }

    function test_Withdraw_HappyPath() public {
        vm.prank(alice);
        htlc.initiateSwap{value: swapAmount}(swapId, hashLock, duration, bob);

        uint256 bobBalBefore = bob.balance;

        // Bob withdraws presenting secret preimage
        vm.prank(bob);
        htlc.withdraw(swapId, secretPreimage);

        assertEq(bob.balance, bobBalBefore + swapAmount);

        (, , , , , bool withdrawn, , ) = htlc.swaps(swapId);
        assertTrue(withdrawn);
    }

    function test_Withdraw_RevertWrongPreimage() public {
        vm.prank(alice);
        htlc.initiateSwap{value: swapAmount}(swapId, hashLock, duration, bob);

        bytes memory fakeSecret = "wrong_fake_preimage";
        vm.prank(bob);
        vm.expectRevert("Hashlock mismatch");
        htlc.withdraw(swapId, fakeSecret);
    }

    function test_Refund_HappyPath() public {
        vm.prank(alice);
        htlc.initiateSwap{value: swapAmount}(swapId, hashLock, duration, bob);

        // Attempt refund before expiry -> revert
        vm.prank(alice);
        vm.expectRevert("Timelock not yet expired");
        htlc.refund(swapId);

        // Fast-forward past timelock
        vm.warp(block.timestamp + duration + 1);

        uint256 aliceBalBefore = alice.balance;
        vm.prank(alice);
        htlc.refund(swapId);

        assertEq(alice.balance, aliceBalBefore + swapAmount);

        (, , , , , , bool refunded, ) = htlc.swaps(swapId);
        assertTrue(refunded);
    }
}
