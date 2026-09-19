// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/reserves/MerkleRegistry.sol";
import "../src/compliance/AccountAbstraction.sol";
import "../src/demo/VirtualFaucet.sol";

contract E2ESolvencyAndSettlementTest is TestBase {
    MerkleRegistry public registry;
    VirtualFaucet public faucet;

    address public exchangeAdmin = address(0xAD01);
    address public auditor = address(0xAA01);
    address public entryPoint = address(0xEE01);

    uint256 internal userPrivateKey = 0x11223344;
    address public userOwner;
    AccountAbstraction public userAccount;

    address public guardianA = address(0x61);
    address public guardianB = address(0x62);

    function setUp() public {
        userOwner = vm.addr(userPrivateKey);

        vm.startPrank(exchangeAdmin);
        registry = new MerkleRegistry(auditor);
        faucet = new VirtualFaucet();
        vm.stopPrank();

        address[] memory guardians = new address[](2);
        guardians[0] = guardianA;
        guardians[1] = guardianB;

        userAccount = new AccountAbstraction(userOwner, entryPoint, guardians, 2);
    }

    function hashPair(bytes32 a, bytes32 b) internal pure returns (bytes32) {
        return a <= b ? keccak256(abi.encodePacked(a, b)) : keccak256(abi.encodePacked(b, a));
    }

    function test_E2E_SolvencyProofAndUserInclusion() public {
        // 1. Simulate 4 user liability records on exchange (Zero-PII anonymized user hashes)
        bytes32[4] memory userIds = [
            keccak256("usr_ind_0001"),
            keccak256("usr_ind_0002"),
            keccak256("usr_ind_0003"),
            keccak256("usr_ind_0004")
        ];

        uint256[4] memory balances = [
            uint256(15_50000000), // 15.5 BTC
            uint256(30_00000000), // 30.0 BTC
            uint256(5_25000000),  // 5.25 BTC
            uint256(49_25000000)  // 49.25 BTC
        ];

        uint256 totalLiabilities = balances[0] + balances[1] + balances[2] + balances[3]; // 100 BTC
        uint256 totalReserves = 102_00000000; // 102 BTC (102% backed)

        bytes32[4] memory leaves;
        for (uint256 i = 0; i < 4; i++) {
            leaves[i] = keccak256(abi.encodePacked(userIds[i], balances[i]));
        }

        bytes32 branch01 = hashPair(leaves[0], leaves[1]);
        bytes32 branch23 = hashPair(leaves[2], leaves[3]);
        bytes32 liabilitiesRoot = hashPair(branch01, branch23);

        // 2. Auditor publishes the attested solvency epoch
        vm.prank(auditor);
        uint256 epoch = registry.publishEpoch(
            liabilitiesRoot,
            keccak256("mpc_cold_vault_btc_multisig_utxos"),
            totalLiabilities,
            totalReserves,
            "ipfs://bafybeisolvencyreportq3"
        );
        assertEq(epoch, 1);

        // 3. User 1 independently audits and proves their inclusion
        bytes32[] memory user1Proof = new bytes32[](2);
        user1Proof[0] = leaves[0];
        user1Proof[1] = branch23;

        bool verified = registry.verifyAccountInclusion(epoch, leaves[1], user1Proof);
        assertTrue(verified, "User 1 should be provably included in exchange liabilities root");

        // 4. Verification fails if someone claims different balance
        bytes32 forgedLeaf1 = keccak256(abi.encodePacked(userIds[1], uint256(999_00000000)));
        bool forgedResult = registry.verifyAccountInclusion(epoch, forgedLeaf1, user1Proof);
        assertFalse(forgedResult, "Tampered balance cannot be verified");
    }

    function test_E2E_SmartAccount_FaucetClaim_And_Recovery() public {
        // 1. Smart account claims demo trading funds from virtual faucet via EntryPoint
        bytes memory claimCalldata = abi.encodeWithSignature("claimDemoFunds()");

        vm.prank(entryPoint);
        userAccount.execute(address(faucet), 0, claimCalldata);

        assertEq(faucet.totalDrippedUsdt(address(userAccount)), 10_000 * 10**18, "USDT credited to smart account");
        assertEq(faucet.totalDrippedBtc(address(userAccount)), 1 * 10**8, "BTC credited to smart account");

        // 2. User signs UserOperation to perform gasless transaction
        bytes32 userOpHash = keccak256("userOp_place_btc_buy_order");
        bytes32 ethSignedMessageHash = keccak256(
            abi.encodePacked("\x19Ethereum Signed Message:\n32", userOpHash)
        );
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(userPrivateKey, ethSignedMessageHash);
        bytes memory signature = abi.encodePacked(r, s, v);

        vm.prank(entryPoint);
        uint256 valid = userAccount.validateUserOp(userOpHash, signature, 0);
        assertEq(valid, 0, "UserOp validated by entryPoint");

        // 3. User loses private key, initiates multi-guardian social recovery
        address backupOwner = address(0xCAFE99);

        vm.prank(guardianA);
        userAccount.initiateRecovery(backupOwner);

        vm.prank(guardianB);
        userAccount.initiateRecovery(backupOwner);

        // Fast forward 48 hours timelock
        vm.warp(block.timestamp + 48 hours);
        userAccount.executeRecovery();

        assertEq(userAccount.owner(), backupOwner, "Ownership successfully recovered to backupOwner");
    }
}
