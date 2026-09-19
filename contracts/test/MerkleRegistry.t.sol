// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./TestBase.sol";
import "../src/reserves/MerkleRegistry.sol";

contract MerkleRegistryTest is TestBase {
    MerkleRegistry public registry;
    address public admin = address(0xAD314);
    address public auditor = address(0xAA11);
    address public alice = address(0xA11CE);
    address public bob = address(0xB0B);

    function setUp() public {
        vm.prank(admin);
        registry = new MerkleRegistry(auditor);
    }

    function test_Deployment_SetsAdminAndAuditor() public view {
        assertEq(registry.exchangeAdmin(), admin, "Admin should match deployer");
        assertEq(registry.auditor(), auditor, "Auditor should match constructor argument");
        assertEq(registry.currentEpoch(), 0, "Initial epoch should be 0");
    }

    function test_Deployment_RevertOnZeroAuditor() public {
        try new MerkleRegistry(address(0)) {
            assertTrue(false, "Should revert on address(0) auditor");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid auditor");
        }
    }

    function test_SetAuditor_Success() public {
        address newAuditor = address(0x9999);
        vm.prank(admin);
        registry.setAuditor(newAuditor);
        assertEq(registry.auditor(), newAuditor, "Auditor should be updated");
    }

    function test_SetAuditor_RevertNonAdmin() public {
        vm.prank(alice);
        try registry.setAuditor(address(0x123)) {
            assertTrue(false, "Should revert for non-admin");
        } catch Error(string memory reason) {
            assertEq(reason, "Only admin");
        }
    }

    function test_SetAuditor_RevertZeroAddress() public {
        vm.prank(admin);
        try registry.setAuditor(address(0)) {
            assertTrue(false, "Should revert for zero address auditor");
        } catch Error(string memory reason) {
            assertEq(reason, "Invalid address");
        }
    }

    function test_PublishEpoch_ByAuditor() public {
        bytes32 liabilitiesRoot = keccak256("liabilities_epoch_1");
        bytes32 assetsRoot = keccak256("assets_epoch_1");
        uint256 totalLiabilities = 100_000_000_000; // 1,000 BTC in e8
        uint256 totalReserves = 105_000_000_000;    // 1,050 BTC in e8 (105% backed)
        string memory reportUri = "ipfs://QmSolvencyReportEpoch1";

        vm.prank(auditor);
        uint256 epochId = registry.publishEpoch(
            liabilitiesRoot,
            assetsRoot,
            totalLiabilities,
            totalReserves,
            reportUri
        );

        assertEq(epochId, 1, "First epoch should be 1");
        assertEq(registry.currentEpoch(), 1, "currentEpoch should increment to 1");

        (
            bytes32 storedLiabilitiesRoot,
            bytes32 storedAssetsRoot,
            uint256 storedLiabilities,
            uint256 storedReserves,
            uint256 storedTimestamp,
            string memory storedReportUri
        ) = registry.epochs(1);

        assertEq(storedLiabilitiesRoot, liabilitiesRoot);
        assertEq(storedAssetsRoot, assetsRoot);
        assertEq(storedLiabilities, totalLiabilities);
        assertEq(storedReserves, totalReserves);
        assertTrue(storedTimestamp > 0);
        assertEq(storedReportUri, reportUri);
    }

    function test_PublishEpoch_ByAdmin() public {
        bytes32 liabilitiesRoot = keccak256("admin_liab");
        bytes32 assetsRoot = keccak256("admin_asset");

        vm.prank(admin);
        uint256 epochId = registry.publishEpoch(
            liabilitiesRoot,
            assetsRoot,
            500,
            600,
            "ipfs://admin-report"
        );
        assertEq(epochId, 1);
    }

    function test_PublishEpoch_RevertUnauthorized() public {
        vm.prank(alice);
        try registry.publishEpoch(bytes32(0), bytes32(0), 100, 200, "") {
            assertTrue(false, "Should revert unauthorized caller");
        } catch Error(string memory reason) {
            assertEq(reason, "Unauthorized");
        }
    }

    function test_PublishEpoch_RevertInsolvencyReservesLessThanLiabilities() public {
        bytes32 liabilitiesRoot = keccak256("liab");
        bytes32 assetsRoot = keccak256("asset");
        uint256 totalLiabilities = 100_000;
        uint256 totalReserves = 99_999; // Less than liabilities

        vm.prank(auditor);
        try registry.publishEpoch(liabilitiesRoot, assetsRoot, totalLiabilities, totalReserves, "ipfs://insolvent") {
            assertTrue(false, "Should revert on fractional reserve / insolvency");
        } catch Error(string memory reason) {
            assertEq(reason, "Insolvency: Reserves less than liabilities");
        }
    }

    function test_PublishEpoch_SequentialEpochs() public {
        vm.prank(auditor);
        uint256 ep1 = registry.publishEpoch(keccak256("1"), keccak256("1"), 100, 100, "uri1");
        assertEq(ep1, 1);

        vm.prank(auditor);
        uint256 ep2 = registry.publishEpoch(keccak256("2"), keccak256("2"), 200, 250, "uri2");
        assertEq(ep2, 2);
        assertEq(registry.currentEpoch(), 2);
    }

    // --- Merkle Proof Verification Tests ---

    function hashPair(bytes32 a, bytes32 b) internal pure returns (bytes32) {
        if (a <= b) {
            return keccak256(abi.encodePacked(a, b));
        } else {
            return keccak256(abi.encodePacked(b, a));
        }
    }

    function test_VerifyProof_SingleLeafTree() public {
        bytes32 leaf = keccak256(abi.encodePacked("alice", uint256(50_00000000)));
        bytes32 root = leaf; // In single-leaf tree, root == leaf

        vm.prank(auditor);
        registry.publishEpoch(root, bytes32(0), 50_00000000, 60_00000000, "report");

        bytes32[] memory emptyProof = new bytes32[](0);
        bool valid = registry.verifyAccountInclusion(1, leaf, emptyProof);
        assertTrue(valid, "Single leaf with empty proof should be valid");
    }

    function test_VerifyProof_TwoLeaves() public {
        bytes32 leafA = keccak256(abi.encodePacked("userA", uint256(100)));
        bytes32 leafB = keccak256(abi.encodePacked("userB", uint256(200)));

        bytes32 root = hashPair(leafA, leafB);

        vm.prank(auditor);
        registry.publishEpoch(root, bytes32(0), 300, 400, "report");

        // Proof for leafA is [leafB]
        bytes32[] memory proofA = new bytes32[](1);
        proofA[0] = leafB;

        bool validA = registry.verifyAccountInclusion(1, leafA, proofA);
        assertTrue(validA, "Leaf A should be verified");

        // Proof for leafB is [leafA]
        bytes32[] memory proofB = new bytes32[](1);
        proofB[0] = leafA;

        bool validB = registry.verifyAccountInclusion(1, leafB, proofB);
        assertTrue(validB, "Leaf B should be verified");

        // Tampered leaf should fail
        bytes32 tamperedLeaf = keccak256(abi.encodePacked("userA", uint256(999)));
        bool invalid = registry.verifyAccountInclusion(1, tamperedLeaf, proofA);
        assertFalse(invalid, "Tampered leaf should return false");

        // Tampered proof should fail
        bytes32[] memory tamperedProof = new bytes32[](1);
        tamperedProof[0] = keccak256("random_fake_sibling");
        bool invalidProof = registry.verifyAccountInclusion(1, leafA, tamperedProof);
        assertFalse(invalidProof, "Tampered proof should return false");
    }

    function test_VerifyProof_FourLeaves() public {
        bytes32 leaf0 = keccak256(abi.encodePacked("user0", uint256(10)));
        bytes32 leaf1 = keccak256(abi.encodePacked("user1", uint256(20)));
        bytes32 leaf2 = keccak256(abi.encodePacked("user2", uint256(30)));
        bytes32 leaf3 = keccak256(abi.encodePacked("user3", uint256(40)));

        bytes32 node01 = hashPair(leaf0, leaf1);
        bytes32 node23 = hashPair(leaf2, leaf3);
        bytes32 root = hashPair(node01, node23);

        vm.prank(auditor);
        registry.publishEpoch(root, bytes32(0), 100, 150, "4-leaf report");

        // Proof for leaf0: [leaf1, node23]
        bytes32[] memory proof0 = new bytes32[](2);
        proof0[0] = leaf1;
        proof0[1] = node23;

        assertTrue(registry.verifyAccountInclusion(1, leaf0, proof0), "Leaf 0 inclusion proof valid");

        // Proof for leaf3: [leaf2, node01]
        bytes32[] memory proof3 = new bytes32[](2);
        proof3[0] = leaf2;
        proof3[1] = node01;

        assertTrue(registry.verifyAccountInclusion(1, leaf3, proof3), "Leaf 3 inclusion proof valid");
    }

    function test_VerifyProof_EightLeavesBalanced() public {
        bytes32[8] memory leaves;
        for (uint256 i = 0; i < 8; i++) {
            leaves[i] = keccak256(abi.encodePacked("account", i, uint256((i + 1) * 100)));
        }

        bytes32[4] memory level1;
        level1[0] = hashPair(leaves[0], leaves[1]);
        level1[1] = hashPair(leaves[2], leaves[3]);
        level1[2] = hashPair(leaves[4], leaves[5]);
        level1[3] = hashPair(leaves[6], leaves[7]);

        bytes32[2] memory level2;
        level2[0] = hashPair(level1[0], level1[1]);
        level2[1] = hashPair(level1[2], level1[3]);

        bytes32 root = hashPair(level2[0], level2[1]);

        vm.prank(auditor);
        registry.publishEpoch(root, bytes32(0), 3600, 5000, "8-leaf report");

        // Verify Leaf 5 (index 5):
        // Sibling at level 0: leaves[4]
        // Sibling at level 1: level1[3] (pair of 6 and 7)
        // Sibling at level 2: level2[0] (subtree containing 0..3)
        bytes32[] memory proof5 = new bytes32[](3);
        proof5[0] = leaves[4];
        proof5[1] = level1[3];
        proof5[2] = level2[0];

        assertTrue(registry.verifyAccountInclusion(1, leaves[5], proof5), "Leaf 5 verified in 8-leaf tree");
    }

    function test_VerifyProof_RevertNonExistentEpoch() public {
        bytes32 leaf = keccak256("any");
        bytes32[] memory proof = new bytes32[](0);

        // Epoch 0 query reverts
        try registry.verifyAccountInclusion(0, leaf, proof) {
            assertTrue(false, "Epoch 0 should revert");
        } catch Error(string memory reason) {
            assertEq(reason, "Epoch does not exist");
        }

        // Epoch 1 query before publish reverts
        try registry.verifyAccountInclusion(1, leaf, proof) {
            assertTrue(false, "Epoch 1 before publish should revert");
        } catch Error(string memory reason) {
            assertEq(reason, "Epoch does not exist");
        }
    }
}
