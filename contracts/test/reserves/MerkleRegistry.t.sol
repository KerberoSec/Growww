// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "../../src/reserves/MerkleRegistry.sol";

contract MerkleRegistryTest is Test {
    MerkleRegistry public registry;

    address public admin = address(0xAA1);
    address public auditor = address(0xBB2);
    address public stranger = address(0xCC3);

    function setUp() public {
        vm.prank(admin);
        registry = new MerkleRegistry(auditor);
    }

    function test_Initialization() public view {
        assertEq(registry.exchangeAdmin(), admin);
        assertEq(registry.auditor(), auditor);
        assertEq(registry.currentEpoch(), 0);
    }

    function test_PublishEpoch_Success() public {
        bytes32 leaf1 = keccak256("leaf1");
        bytes32 leaf2 = keccak256("leaf2");

        // Two-leaf Merkle root: keccak256(leaf1 + leaf2) sorted
        bytes32 root;
        if (leaf1 <= leaf2) {
            root = keccak256(abi.encodePacked(leaf1, leaf2));
        } else {
            root = keccak256(abi.encodePacked(leaf2, leaf1));
        }

        bytes32 assetRoot = keccak256("asset_holdings_root");

        vm.prank(auditor);
        uint256 epochId = registry.publishEpoch(
            root,
            assetRoot,
            100000e8, // Liabilities
            105000e8, // Reserves (105% solvency)
            "https://audit.growww.in/por/epoch-1.json"
        );

        assertEq(epochId, 1);
        assertEq(registry.currentEpoch(), 1);

        // Verify account inclusion for leaf1
        bytes32[] memory proof = new bytes32[](1);
        proof[0] = leaf2;

        bool verified = registry.verifyAccountInclusion(1, leaf1, proof);
        assertTrue(verified);

        // Verify with corrupted proof fails
        bytes32[] memory badProof = new bytes32[](1);
        badProof[0] = keccak256("fake_sibling");
        bool badVerified = registry.verifyAccountInclusion(1, leaf1, badProof);
        assertFalse(badVerified);
    }

    function test_PublishEpoch_RevertInsolvent() public {
        vm.prank(auditor);
        vm.expectRevert("Insolvency: Reserves less than liabilities");
        registry.publishEpoch(
            keccak256("root"),
            keccak256("asset"),
            100000e8, // 100k liabilities
            99000e8,  // 99k reserves -> deficit!
            "https://audit.growww.in/por/epoch-bad.json"
        );
    }

    function test_PublishEpoch_RevertUnauthorized() public {
        vm.prank(stranger);
        vm.expectRevert("Unauthorized");
        registry.publishEpoch(
            keccak256("root"),
            keccak256("asset"),
            100000e8,
            100000e8,
            ""
        );
    }
}
