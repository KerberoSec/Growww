// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/bridge/HTLCSwap.sol";

/**
 * @notice Malicious receiver contract attempting reentrancy into HTLCSwap
 */
contract MaliciousHTLCReceiver {
    HTLCSwap public htlc;
    bytes32 public targetSwapId;
    bytes public targetPreimage;
    uint256 public attackAttempts;
    bool public attackSucceeded;

    constructor(address _htlc) {
        htlc = HTLCSwap(_htlc);
    }

    function setTarget(bytes32 _swapId, bytes calldata _preimage) external {
        targetSwapId = _swapId;
        targetPreimage = _preimage;
        attackAttempts = 0;
        attackSucceeded = false;
    }

    receive() external payable {
        attackAttempts++;
        if (attackAttempts == 1) {
            // Attempt to re-enter withdraw
            try htlc.withdraw(targetSwapId, targetPreimage) {
                attackSucceeded = true;
            } catch {
                // Expected revert: Already withdrawn
            }
        }
    }
}

/**
 * @notice Malicious sender contract attempting reentrancy into HTLCSwap refund
 */
contract MaliciousHTLCSender {
    HTLCSwap public htlc;
    bytes32 public targetSwapId;
    uint256 public attackAttempts;
    bool public attackSucceeded;

    constructor(address _htlc) {
        htlc = HTLCSwap(_htlc);
    }

    function initiate(
        bytes32 _swapId,
        bytes32 _hashLock,
        uint256 _duration,
        address payable _receiver
    ) external payable {
        targetSwapId = _swapId;
        attackAttempts = 0;
        attackSucceeded = false;
        htlc.initiateSwap{value: msg.value}(_swapId, _hashLock, _duration, _receiver);
    }

    receive() external payable {
        attackAttempts++;
        if (attackAttempts == 1) {
            // Attempt to re-enter refund
            try htlc.refund(targetSwapId) {
                attackSucceeded = true;
            } catch {
                // Expected revert: Already refunded
            }
        }
    }
}

/**
 * @notice Recipient that always reverts on receiving ETH
 */
contract HTLCRevertingReceiver {
    receive() external payable {
        revert("Rejection of incoming ETH");
    }
}

contract HTLCSwapTest is TestBase {
    HTLCSwap htlc;

    address payable sender = payable(address(0x1111));
    address payable receiver = payable(address(0x2222));
    address stranger = address(0x9999);

    bytes defaultPreimage = "growww_nbse_atomic_secret_2026";
    bytes32 defaultHashLock;
    bytes32 defaultSwapId = keccak256("htlc_swap_001");
    uint256 defaultDuration = 2 hours;
    uint256 defaultAmount = 2 ether;

    function setUp() public {
        htlc = new HTLCSwap();
        defaultHashLock = sha256(defaultPreimage);
        vm.deal(sender, 100 ether);
        vm.deal(receiver, 100 ether);
    }

    /* -------------------------------------------------------------------------- */
    /*                         1. SWAP INITIATION TESTS                           */
    /* -------------------------------------------------------------------------- */

    function test_InitiateSwap_Success() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        (
            bytes32 hashLock,
            uint256 timelock,
            uint256 value,
            address payable s,
            address payable r,
            bool withdrawn,
            bool refunded,
            bytes memory preimage
        ) = htlc.swaps(defaultSwapId);

        assertEq(hashLock, defaultHashLock);
        assertEq(timelock, block.timestamp + defaultDuration);
        assertEq(value, defaultAmount);
        assertEq(s, sender);
        assertEq(r, receiver);
        assertFalse(withdrawn);
        assertFalse(refunded);
        assertEq(preimage.length, 0);
        assertEq(address(htlc).balance, defaultAmount);
    }

    function test_InitiateSwap_RevertZeroValue() public {
        vm.prank(sender);
        vm.expectRevert("No funds provided");
        htlc.initiateSwap{value: 0}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );
    }

    function test_InitiateSwap_RevertZeroReceiver() public {
        vm.prank(sender);
        vm.expectRevert("Invalid receiver");
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            payable(address(0))
        );
    }

    function test_InitiateSwap_RevertShortTimelock() public {
        vm.prank(sender);
        vm.expectRevert("Timelock duration too short");
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            59 minutes, // Less than 1 hour
            receiver
        );
    }

    function test_InitiateSwap_RevertDuplicateSwapId() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        vm.prank(sender);
        vm.expectRevert("Swap already exists");
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );
    }

    /* -------------------------------------------------------------------------- */
    /*                         2. PREIMAGE WITHDRAWAL TESTS                       */
    /* -------------------------------------------------------------------------- */

    function test_Withdraw_Success() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        uint256 receiverBalBefore = receiver.balance;

        // Receiver claims with preimage
        vm.prank(receiver);
        htlc.withdraw(defaultSwapId, defaultPreimage);

        (, , , , , bool withdrawn, bool refunded, bytes memory storedPreimage) = htlc.swaps(defaultSwapId);
        assertTrue(withdrawn, "Must be withdrawn");
        assertFalse(refunded);
        assertEq(storedPreimage, defaultPreimage);
        assertEq(receiver.balance, receiverBalBefore + defaultAmount);
        assertEq(address(htlc).balance, 0);
    }

    function test_Withdraw_SuccessByThirdPartyRelayer() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        uint256 receiverBalBefore = receiver.balance;

        // Third party relayer submits the preimage
        vm.prank(stranger);
        htlc.withdraw(defaultSwapId, defaultPreimage);

        // Funds still go strictly to the receiver
        assertEq(receiver.balance, receiverBalBefore + defaultAmount);
        assertEq(stranger.balance, 0);
    }

    function test_Withdraw_RevertWrongPreimage() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        bytes memory wrongPreimage = "wrong_secret_attempt";
        vm.prank(receiver);
        vm.expectRevert("Hashlock mismatch");
        htlc.withdraw(defaultSwapId, wrongPreimage);
    }

    function test_Withdraw_RevertEmptyPreimage() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        bytes memory emptyPreimage = "";
        vm.prank(receiver);
        vm.expectRevert("Hashlock mismatch");
        htlc.withdraw(defaultSwapId, emptyPreimage);
    }

    function test_Withdraw_RevertNonExistentSwap() public {
        bytes32 nonExistentId = keccak256("non_existent_swap");
        vm.prank(receiver);
        vm.expectRevert("Swap not found");
        htlc.withdraw(nonExistentId, defaultPreimage);
    }

    function test_Withdraw_RevertAlreadyWithdrawn() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        vm.prank(receiver);
        htlc.withdraw(defaultSwapId, defaultPreimage);

        // Second withdrawal attempt
        vm.prank(receiver);
        vm.expectRevert("Already withdrawn");
        htlc.withdraw(defaultSwapId, defaultPreimage);
    }

    /* -------------------------------------------------------------------------- */
    /*                         3. TIMELOCK EXPIRY & REFUND TESTS                  */
    /* -------------------------------------------------------------------------- */

    function test_Refund_SuccessAfterTimelock() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        // Advance time to exactly timelock expiry
        vm.warp(block.timestamp + defaultDuration);

        uint256 senderBalBefore = sender.balance;
        vm.prank(sender);
        htlc.refund(defaultSwapId);

        (, , , , , bool withdrawn, bool refunded, ) = htlc.swaps(defaultSwapId);
        assertFalse(withdrawn);
        assertTrue(refunded, "Must be refunded");
        assertEq(sender.balance, senderBalBefore + defaultAmount);
        assertEq(address(htlc).balance, 0);
    }

    function test_Refund_RevertBeforeTimelock() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        // 1 second before timelock expiry
        vm.warp(block.timestamp + defaultDuration - 1);

        vm.prank(sender);
        vm.expectRevert("Timelock not yet expired");
        htlc.refund(defaultSwapId);
    }

    function test_Refund_RevertAlreadyRefunded() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        vm.warp(block.timestamp + defaultDuration);

        vm.prank(sender);
        htlc.refund(defaultSwapId);

        // Second refund attempt
        vm.prank(sender);
        vm.expectRevert("Already refunded");
        htlc.refund(defaultSwapId);
    }

    function test_Refund_RevertNonExistentSwap() public {
        bytes32 nonExistentId = keccak256("ghost_swap");
        vm.prank(sender);
        vm.expectRevert("Swap not found");
        htlc.refund(nonExistentId);
    }

    /* -------------------------------------------------------------------------- */
    /*                         4. RACE CONDITIONS & MUTUAL EXCLUSIVITY            */
    /* -------------------------------------------------------------------------- */

    function test_Refund_RevertIfAlreadyWithdrawn() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        // Receiver withdraws in time
        vm.prank(receiver);
        htlc.withdraw(defaultSwapId, defaultPreimage);

        // Time passes past timelock
        vm.warp(block.timestamp + defaultDuration + 100);

        // Sender attempts late refund
        vm.prank(sender);
        vm.expectRevert("Already withdrawn");
        htlc.refund(defaultSwapId);
    }

    function test_Withdraw_RevertIfAlreadyRefunded() public {
        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            defaultSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        // Timelock expires and sender refunds
        vm.warp(block.timestamp + defaultDuration + 1);
        vm.prank(sender);
        htlc.refund(defaultSwapId);

        // Receiver attempts late withdrawal with valid preimage
        vm.prank(receiver);
        vm.expectRevert("Already refunded");
        htlc.withdraw(defaultSwapId, defaultPreimage);
    }

    /* -------------------------------------------------------------------------- */
    /*                         5. REENTRANCY ATTACK RESISTANCE                    */
    /* -------------------------------------------------------------------------- */

    function test_Reentrancy_ReceiverCannotDrainOnWithdraw() public {
        MaliciousHTLCReceiver malReceiver = new MaliciousHTLCReceiver(address(htlc));
        bytes32 attackSwapId = keccak256("attack_swap_receiver");

        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            attackSwapId,
            defaultHashLock,
            defaultDuration,
            payable(address(malReceiver))
        );

        malReceiver.setTarget(attackSwapId, defaultPreimage);

        // Trigger withdrawal
        vm.prank(address(malReceiver));
        htlc.withdraw(attackSwapId, defaultPreimage);

        assertFalse(malReceiver.attackSucceeded(), "Reentrant withdraw must fail");
        assertEq(malReceiver.attackAttempts(), 1, "Only one withdrawal occurred");
        assertEq(address(malReceiver).balance, defaultAmount, "Exact amount received");
    }

    function test_Reentrancy_SenderCannotDrainOnRefund() public {
        MaliciousHTLCSender malSender = new MaliciousHTLCSender(address(htlc));
        bytes32 attackSwapId = keccak256("attack_swap_sender");

        // MalSender initiates swap
        malSender.initiate{value: defaultAmount}(
            attackSwapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        // Advance past timelock
        vm.warp(block.timestamp + defaultDuration + 1);

        // MalSender calls refund
        vm.prank(address(malSender));
        htlc.refund(attackSwapId);

        assertFalse(malSender.attackSucceeded(), "Reentrant refund must fail");
        assertEq(malSender.attackAttempts(), 1, "Only one refund occurred");
        assertEq(address(malSender).balance, defaultAmount, "Exact refund received");
    }

    /* -------------------------------------------------------------------------- */
    /*                         6. TRANSFER FAILURE RESISTANCE                     */
    /* -------------------------------------------------------------------------- */

    function test_Withdraw_RevertOnReceiverTransferFailure() public {
        HTLCRevertingReceiver revReceiver = new HTLCRevertingReceiver();
        bytes32 swapId = keccak256("rev_receiver_swap");

        vm.prank(sender);
        htlc.initiateSwap{value: defaultAmount}(
            swapId,
            defaultHashLock,
            defaultDuration,
            payable(address(revReceiver))
        );

        vm.expectRevert("Transfer failed");
        htlc.withdraw(swapId, defaultPreimage);
    }

    function test_Refund_RevertOnSenderTransferFailure() public {
        HTLCRevertingReceiver revSender = new HTLCRevertingReceiver();
        bytes32 swapId = keccak256("rev_sender_swap");

        // Deal ETH to reverting sender
        vm.deal(address(revSender), 10 ether);

        // Use prank to make revSender initiate swap
        vm.prank(address(revSender));
        htlc.initiateSwap{value: defaultAmount}(
            swapId,
            defaultHashLock,
            defaultDuration,
            receiver
        );

        vm.warp(block.timestamp + defaultDuration + 1);

        vm.expectRevert("Refund transfer failed");
        htlc.refund(swapId);
    }
}
