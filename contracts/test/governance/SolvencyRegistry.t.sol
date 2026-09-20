// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {SolvencyRegistry} from "../../src/governance/SolvencyRegistry.sol";
import {MessageHashUtils} from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

contract SolvencyRegistryTest is Test {
    using MessageHashUtils for bytes32;

    SolvencyRegistry internal registry;

    address internal admin = address(0xAA1);
    address internal publisher = address(0xBB1);

    uint256 internal auditorPrivateKey = 0xA11CE;
    address internal auditor;

    function setUp() public {
        auditor = vm.addr(auditorPrivateKey);

        vm.startPrank(admin);
        registry = new SolvencyRegistry(admin);
        registry.grantRole(registry.PUBLISHER_ROLE(), publisher);
        registry.grantRole(registry.AUDITOR_ROLE(), auditor);
        vm.stopPrank();
    }

    function test_PublishSolvencySnapshot_Success() public {
        bytes32 merkleRoot = keccak256("MERKLE_ROOT_2026_Q1");
        uint256 totalLiability = 1_000_000 * 1e18;
        uint256 totalReserve = 1_050_000 * 1e18; // 105% solvency

        bytes32 digest = keccak256(abi.encodePacked(merkleRoot, totalLiability, totalReserve, block.chainid));
        bytes32 ethHash = digest.toEthSignedMessageHash();
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(auditorPrivateKey, ethHash);

        vm.prank(publisher);
        uint256 sid = registry.publishSolvencySnapshot(merkleRoot, totalLiability, totalReserve, v, r, s);

        assertEq(sid, 1);
        assertEq(registry.latestMerkleRoot(), merkleRoot);
        SolvencyRegistry.SolvencySnapshot memory snap = registry.getLatestSnapshot();
        assertEq(snap.solvencyRatioBps, 10500); // 105.00%
        assertEq(snap.auditor, auditor);
    }

    function test_InsolventSnapshot_Revert() public {
        bytes32 merkleRoot = keccak256("MERKLE_ROOT_INSOLVENT");
        uint256 totalLiability = 1_000_000 * 1e18;
        uint256 totalReserve = 900_000 * 1e18; // 90% -> insolvent!

        bytes32 digest = keccak256(abi.encodePacked(merkleRoot, totalLiability, totalReserve, block.chainid));
        bytes32 ethHash = digest.toEthSignedMessageHash();
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(auditorPrivateKey, ethHash);

        vm.prank(publisher);
        vm.expectRevert(SolvencyRegistry.InsolventExchange.selector);
        registry.publishSolvencySnapshot(merkleRoot, totalLiability, totalReserve, v, r, s);
    }

    function test_MerkleInclusionVerification() public view {
        // Construct 2-leaf tree: leaf0 and leaf1
        bytes32 leaf0 = keccak256("USER_ALICE_BALANCE_100");
        bytes32 leaf1 = keccak256("USER_BOB_BALANCE_200");

        bytes32 root;
        if (leaf0 <= leaf1) {
            root = keccak256(abi.encodePacked(leaf0, leaf1));
        } else {
            root = keccak256(abi.encodePacked(leaf1, leaf0));
        }

        // Proof for leaf0 is [leaf1]
        bytes32[] memory proof0 = new bytes32[](1);
        proof0[0] = leaf1;

        bool verified = registry.verifyLeafInclusion(root, leaf0, proof0);
        assertTrue(verified);

        // Invalid proof fails
        proof0[0] = keccak256("CORRUPTED_LEAF");
        bool failedVerification = registry.verifyLeafInclusion(root, leaf0, proof0);
        assertFalse(failedVerification);
    }
}
