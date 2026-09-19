// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/compliance/AccountAbstraction.sol";

contract MockTarget {
    uint256 public valueReceived;
    bytes public dataReceived;
    uint256 public counter;

    function increment() external payable {
        counter++;
        valueReceived += msg.value;
    }

    function failingFunction() external pure {
        revert("Target operation failed");
    }

    receive() external payable {
        valueReceived += msg.value;
    }
}

contract AccountAbstractionTest is TestBase {
    AccountAbstraction public account;
    MockTarget public target;

    uint256 internal ownerPrivateKey = 0xA11CE01;
    address public owner;
    address public entryPoint = address(0xEEEE);

    address public guardian1 = address(0x1111);
    address public guardian2 = address(0x2222);
    address public guardian3 = address(0x3333);
    address[] public initialGuardians;

    function setUp() public {
        owner = vm.addr(ownerPrivateKey);

        initialGuardians.push(guardian1);
        initialGuardians.push(guardian2);
        initialGuardians.push(guardian3);

        account = new AccountAbstraction(owner, entryPoint, initialGuardians, 2);
        target = new MockTarget();
    }

    // --- Constructor & Initialization Tests ---

    function test_Initialization_Success() public view {
        assertEq(account.owner(), owner, "Owner mismatch");
        assertEq(account.entryPoint(), entryPoint, "EntryPoint mismatch");
        assertEq(account.guardianQuorum(), 2, "Quorum mismatch");
        assertEq(account.nonce(), 0, "Initial nonce should be 0");
        assertTrue(account.isGuardian(guardian1), "Guardian 1 should be active");
        assertTrue(account.isGuardian(guardian2), "Guardian 2 should be active");
        assertTrue(account.isGuardian(guardian3), "Guardian 3 should be active");
        assertFalse(account.isGuardian(address(0x9999)), "Random address is not guardian");
    }

    function test_Initialization_RevertZeroOwner() public {
        try new AccountAbstraction(address(0), entryPoint, initialGuardians, 2) {
            assertTrue(false, "Should revert for address(0) owner");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid owner");
        }
    }

    function test_Initialization_RevertZeroQuorum() public {
        try new AccountAbstraction(owner, entryPoint, initialGuardians, 0) {
            assertTrue(false, "Should revert for 0 quorum");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid quorum");
        }
    }

    function test_Initialization_RevertQuorumExceedsGuardians() public {
        try new AccountAbstraction(owner, entryPoint, initialGuardians, 4) {
            assertTrue(false, "Should revert when quorum > guardian count");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid quorum");
        }
    }

    function test_Initialization_RevertDuplicateGuardian() public {
        address[] memory dups = new address[](2);
        dups[0] = guardian1;
        dups[1] = guardian1; // Duplicate
        try new AccountAbstraction(owner, entryPoint, dups, 1) {
            assertTrue(false, "Should revert on duplicate guardian");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid guardian");
        }
    }

    function test_Initialization_RevertZeroAddressGuardian() public {
        address[] memory zeroG = new address[](2);
        zeroG[0] = guardian1;
        zeroG[1] = address(0);
        try new AccountAbstraction(owner, entryPoint, zeroG, 1) {
            assertTrue(false, "Should revert on address(0) guardian");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid guardian");
        }
    }

    // --- Direct and EntryPoint Execution Tests ---

    function test_Execute_ByOwner_Success() public {
        bytes memory callData = abi.encodeWithSignature("increment()");
        vm.prank(owner);
        bytes memory res = account.execute(address(target), 0, callData);

        assertEq(target.counter(), 1, "Counter should be incremented");
        assertEq(account.nonce(), 1, "Account nonce should increment");
        assertTrue(res.length == 0, "No return data expected");
    }

    function test_Execute_WithValue_Transfer() public {
        vm.deal(address(account), 5 ether);

        bytes memory callData = abi.encodeWithSignature("increment()");
        vm.prank(owner);
        account.execute(address(target), 2 ether, callData);

        assertEq(address(target).balance, 2 ether, "Target should receive 2 ETH");
        assertEq(address(account).balance, 3 ether, "Account balance should be 3 ETH");
        assertEq(target.counter(), 1);
    }

    function test_Execute_ByEntryPoint_Success() public {
        bytes memory callData = abi.encodeWithSignature("increment()");
        vm.prank(entryPoint);
        account.execute(address(target), 0, callData);

        assertEq(target.counter(), 1, "EntryPoint should be authorized to execute");
        assertEq(account.nonce(), 1);
    }

    function test_Execute_RevertUnauthorizedCaller() public {
        bytes memory callData = abi.encodeWithSignature("increment()");
        vm.prank(guardian1);
        try account.execute(address(target), 0, callData) {
            assertTrue(false, "Non-owner non-entrypoint caller must revert");
        } catch Error(string memory reason) {
            assertEq(reason, "Only EntryPoint or Owner");
        }
    }

    function test_Execute_RevertIfTargetFails() public {
        bytes memory callData = abi.encodeWithSignature("failingFunction()");
        vm.prank(owner);
        try account.execute(address(target), 0, callData) {
            assertTrue(false, "Should revert if target call reverts");
        } catch Error(string memory reason) {
            assertEq(reason, "Execution failed");
        }
    }

    // --- ERC-4337 validateUserOp Tests ---

    function test_ValidateUserOp_Success() public {
        bytes32 userOpHash = keccak256("userOp_transfer_tokens_100");
        bytes32 ethSignedMessageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", userOpHash)
        );

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(ownerPrivateKey, ethSignedMessageHash);
        bytes memory signature = abi.encodePacked(r, s, v);

        vm.prank(entryPoint);
        uint256 validationData = account.validateUserOp(userOpHash, signature, 0);

        assertEq(validationData, 0, "Validation should succeed (return 0)");
    }

    function test_ValidateUserOp_InvalidSignatureReturnsOne() public {
        bytes32 userOpHash = keccak256("userOp_transfer_tokens_100");
        bytes32 ethSignedMessageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", userOpHash)
        );

        // Sign with an unauthorized key
        uint256 attackerKey = 0xBADB01;
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(attackerKey, ethSignedMessageHash);
        bytes memory signature = abi.encodePacked(r, s, v);

        vm.prank(entryPoint);
        uint256 validationData = account.validateUserOp(userOpHash, signature, 0);

        assertEq(validationData, 1, "Validation should fail (return 1)");
    }

    function test_ValidateUserOp_RevertNonEntryPointCaller() public {
        bytes32 userOpHash = keccak256("op");
        bytes memory dummySig = new bytes(65);

        vm.prank(owner);
        try account.validateUserOp(userOpHash, dummySig, 0) {
            assertTrue(false, "Should revert for non-EntryPoint caller");
        } catch Error(string memory reason) {
            assertEq(reason, "Only EntryPoint");
        }
    }

    function test_ValidateUserOp_WithMissingAccountFundsReimbursement() public {
        vm.deal(address(account), 1 ether);

        bytes32 userOpHash = keccak256("gas_reimbursed_op");
        bytes32 ethSignedMessageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", userOpHash)
        );
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(ownerPrivateKey, ethSignedMessageHash);
        bytes memory signature = abi.encodePacked(r, s, v);

        uint256 requiredPrefund = 0.05 ether;
        uint256 epBalanceBefore = entryPoint.balance;

        vm.prank(entryPoint);
        uint256 validationData = account.validateUserOp(userOpHash, signature, requiredPrefund);

        assertEq(validationData, 0, "Validation should succeed");
        assertEq(entryPoint.balance - epBalanceBefore, requiredPrefund, "EntryPoint should receive reimbursed gas");
        assertEq(address(account).balance, 0.95 ether, "Account balance debited");
    }

    // --- Social Recovery Flow Tests ---

    function test_SocialRecovery_FullLifecycle() public {
        address newOwner = address(0xCAFE);

        // 1. Guardian 1 initiates recovery
        vm.prank(guardian1);
        account.initiateRecovery(newOwner);

        (address proposed, uint256 approvals, uint256 executeAfter, bool executed) = account.currentRecovery();
        assertEq(proposed, newOwner, "Proposed owner mismatch");
        assertEq(approvals, 1, "First approval count should be 1");
        assertEq(executeAfter, block.timestamp + 48 hours, "Timelock should be +48 hours");
        assertFalse(executed, "Should not be executed yet");

        // 2. Guardian 1 cannot vote again
        vm.prank(guardian1);
        try account.initiateRecovery(newOwner) {
            assertTrue(false, "Guardian 1 should not vote twice");
        } catch Error(string memory reason) {
            assertEq(reason, "Already voted");
        }

        // 3. Trying to execute before quorum met should fail
        try account.executeRecovery() {
            assertTrue(false, "Execute before quorum should revert");
        } catch Error(string memory reason) {
            assertEq(reason, "Quorum not met");
        }

        // 4. Guardian 2 approves recovery (quorum = 2 met)
        vm.prank(guardian2);
        account.initiateRecovery(newOwner);

        (, uint256 approvalsAfterG2, , ) = account.currentRecovery();
        assertEq(approvalsAfterG2, 2, "Quorum reached");

        // 5. Trying to execute before timelock expires should fail
        try account.executeRecovery() {
            assertTrue(false, "Execute before timelock should revert");
        } catch Error(string memory reason) {
            assertEq(reason, "Timelock not expired");
        }

        // 6. Fast forward 48 hours
        vm.warp(block.timestamp + 48 hours);

        // 7. Execute recovery
        account.executeRecovery();

        assertEq(account.owner(), newOwner, "Ownership successfully transferred to newOwner");
        (, , , bool executedAfter) = account.currentRecovery();
        assertTrue(executedAfter, "Recovery marked executed");

        // 8. Trying to execute again reverts
        try account.executeRecovery() {
            assertTrue(false, "Cannot re-execute recovery");
        } catch Error(string memory reason) {
            assertEq(reason, "Recovery already executed");
        }
    }

    function test_SocialRecovery_RevertNonGuardian() public {
        address attacker = address(0xBAD);
        vm.prank(attacker);
        try account.initiateRecovery(attacker) {
            assertTrue(false, "Non-guardian cannot initiate recovery");
        } catch Error(string memory reason) {
            assertEq(reason, "Only guardian");
        }
    }

    function test_SocialRecovery_RevertInvalidNewOwner() public {
        // Zero address
        vm.prank(guardian1);
        try account.initiateRecovery(address(0)) {
            assertTrue(false, "Cannot propose address(0)");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid new owner");
        }

        // Current owner
        vm.prank(guardian1);
        try account.initiateRecovery(owner) {
            assertTrue(false, "Cannot propose current owner");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid new owner");
        }
    }

    function test_SocialRecovery_OwnerCanCancelRecovery() public {
        address rogueOwner = address(0xBAD);

        vm.prank(guardian1);
        account.initiateRecovery(rogueOwner);

        // Owner detects rogue recovery attempt and cancels it
        vm.prank(owner);
        account.cancelRecovery();

        (, , , bool executed) = account.currentRecovery();
        assertTrue(executed, "Recovery should be cancelled/executed");

        // Guardian 2 cannot reach quorum / cannot execute cancelled recovery
        vm.warp(block.timestamp + 48 hours);
        try account.executeRecovery() {
            assertTrue(false, "Cancelled recovery cannot be executed");
        } catch Error(string memory reason) {
            assertEq(reason, "Recovery already executed");
        }

        assertEq(account.owner(), owner, "Owner remains unchanged");
    }

    function test_SocialRecovery_CancelRevertNonOwner() public {
        vm.prank(guardian1);
        try account.cancelRecovery() {
            assertTrue(false, "Only owner can cancel recovery");
        } catch Error(string memory reason) {
            assertEq(reason, "Only owner allowed");
        }
    }
}
