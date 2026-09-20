// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {CustodialProofRegistry} from "../../src/custody/CustodialProofRegistry.sol";
import {ICustodialProofRegistry} from "../../src/interfaces/custody/ICustodialProofRegistry.sol";

contract CustodialProofRegistryTest is Test {
    CustodialProofRegistry public registry;

    address public admin = address(0xAD01);
    address public auditor = address(0xAA01);

    uint256 public custodianPk1 = 0xA11;
    uint256 public custodianPk2 = 0xA22;
    uint256 public custodianPk3 = 0xA33;

    address public custodian1;
    address public custodian2;
    address public custodian3;

    function setUp() public {
        custodian1 = vm.addr(custodianPk1);
        custodian2 = vm.addr(custodianPk2);
        custodian3 = vm.addr(custodianPk3);

        address[] memory custodians = new address[](3);
        custodians[0] = custodian1;
        custodians[1] = custodian2;
        custodians[2] = custodian3;

        vm.prank(admin);
        registry = new CustodialProofRegistry(
            auditor,
            custodians,
            2, // 2-of-3 threshold
            10000, // 100% required reserve ratio
            admin
        );
    }

    function test_Deployment() public view {
        assertEq(registry.auditor(), auditor);
        assertEq(registry.custodianThreshold(), 2);
        assertEq(registry.currentEpoch(), 0);
        assertTrue(registry.isCustodian(custodian1));
        assertTrue(registry.isCustodian(custodian2));
        assertTrue(registry.isCustodian(custodian3));
    }

    function test_PublishAttestedEpoch_Success() public {
        bytes32 liabRoot = keccak256("liab_epoch_1");
        bytes32 assetRoot = keccak256("asset_epoch_1");
        uint256 totalLiabE8 = 100_000 * 1e8;
        uint256 totalResE8 = 105_000 * 1e8; // 105% coverage
        uint256 expiry = block.timestamp + 1 hours;

        bytes32 digest = registry.getEpochDigest(
            liabRoot,
            assetRoot,
            totalLiabE8,
            totalResE8,
            1,
            expiry
        );

        // Sign with Custodian 1 & Custodian 2
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(custodianPk1, digest);
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(custodianPk2, digest);

        bytes[] memory signatures = new bytes[](2);
        signatures[0] = abi.encodePacked(r1, s1, v1);
        signatures[1] = abi.encodePacked(r2, s2, v2);

        vm.prank(auditor);
        uint256 epochId = registry.publishAttestedEpoch(
            liabRoot,
            assetRoot,
            totalLiabE8,
            totalResE8,
            "ipfs://solvency-report-1",
            expiry,
            signatures
        );

        assertEq(epochId, 1);
        assertEq(registry.currentEpoch(), 1);

        ICustodialProofRegistry.SolvencyEpochRecord memory rec = registry.getEpoch(1);
        assertEq(rec.liabilitiesMerkleRoot, liabRoot);
        assertEq(rec.assetsMerkleRoot, assetRoot);
        assertEq(rec.totalLiabilitiesE8, totalLiabE8);
        assertEq(rec.totalReservesE8, totalResE8);
        assertEq(rec.custodianAttestationsCount, 2);

        (bool solvent, uint256 ratioBps) = registry.isSolvent(1);
        assertTrue(solvent);
        assertEq(ratioBps, 10500); // 105.00%
    }

    function test_PublishAttestedEpoch_RevertsInsufficientSignatures() public {
        bytes32 liabRoot = keccak256("liab");
        bytes32 assetRoot = keccak256("asset");
        uint256 expiry = block.timestamp + 1 hours;

        bytes32 digest = registry.getEpochDigest(
            liabRoot,
            assetRoot,
            100,
            100,
            1,
            expiry
        );

        // Only 1 signature
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(custodianPk1, digest);
        bytes[] memory signatures = new bytes[](1);
        signatures[0] = abi.encodePacked(r1, s1, v1);

        vm.prank(auditor);
        vm.expectRevert(
            abi.encodeWithSelector(
                ICustodialProofRegistry.InsufficientCustodianSignatures.selector,
                1,
                2
            )
        );
        registry.publishAttestedEpoch(liabRoot, assetRoot, 100, 100, "", expiry, signatures);
    }

    function test_PublishAttestedEpoch_RevertsDuplicateSignatures() public {
        bytes32 liabRoot = keccak256("liab");
        bytes32 assetRoot = keccak256("asset");
        uint256 expiry = block.timestamp + 1 hours;

        bytes32 digest = registry.getEpochDigest(
            liabRoot,
            assetRoot,
            100,
            100,
            1,
            expiry
        );

        // Custodian 1 signs twice
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(custodianPk1, digest);
        bytes[] memory signatures = new bytes[](2);
        signatures[0] = abi.encodePacked(r1, s1, v1);
        signatures[1] = abi.encodePacked(r1, s1, v1);

        vm.prank(auditor);
        vm.expectRevert(
            abi.encodeWithSelector(
                ICustodialProofRegistry.DuplicateSignature.selector,
                custodian1
            )
        );
        registry.publishAttestedEpoch(liabRoot, assetRoot, 100, 100, "", expiry, signatures);
    }

    function test_PublishAttestedEpoch_RevertsInsolvent() public {
        bytes32 liabRoot = keccak256("liab");
        bytes32 assetRoot = keccak256("asset");
        uint256 expiry = block.timestamp + 1 hours;

        bytes[] memory signatures = new bytes[](2);

        vm.prank(auditor);
        vm.expectRevert(
            abi.encodeWithSelector(
                ICustodialProofRegistry.InsufficientReserves.selector,
                90,
                100
            )
        );
        registry.publishAttestedEpoch(liabRoot, assetRoot, 100, 90, "", expiry, signatures);
    }

    function test_VerifyAssetReserveProof() public {
        ICustodialProofRegistry.AssetReserveLeaf memory leafA = ICustodialProofRegistry.AssetReserveLeaf({
            custodian: custodian1,
            assetId: keccak256("BTC"),
            balanceE8: 50_000 * 1e8,
            depositoryRef: keccak256("CDSL_BATCH_01"),
            timestamp: block.timestamp
        });

        ICustodialProofRegistry.AssetReserveLeaf memory leafB = ICustodialProofRegistry.AssetReserveLeaf({
            custodian: custodian2,
            assetId: keccak256("ETH"),
            balanceE8: 60_000 * 1e8,
            depositoryRef: keccak256("NSDL_BATCH_02"),
            timestamp: block.timestamp
        });

        bytes32 hashA = registry.hashAssetLeaf(leafA);
        bytes32 hashB = registry.hashAssetLeaf(leafB);

        bytes32 root = hashA <= hashB
            ? keccak256(abi.encodePacked(hashA, hashB))
            : keccak256(abi.encodePacked(hashB, hashA));

        bytes32 liabRoot = keccak256("liab");
        uint256 expiry = block.timestamp + 1 hours;

        bytes32 digest = registry.getEpochDigest(
            liabRoot,
            root,
            100_000 * 1e8,
            110_000 * 1e8,
            1,
            expiry
        );

        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(custodianPk1, digest);
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(custodianPk2, digest);

        bytes[] memory signatures = new bytes[](2);
        signatures[0] = abi.encodePacked(r1, s1, v1);
        signatures[1] = abi.encodePacked(r2, s2, v2);

        vm.prank(auditor);
        registry.publishAttestedEpoch(
            liabRoot,
            root,
            100_000 * 1e8,
            110_000 * 1e8,
            "report",
            expiry,
            signatures
        );

        // Verify Leaf A with sibling Leaf B
        bytes32[] memory proofA = new bytes32[](1);
        proofA[0] = hashB;

        bool valid = registry.verifyAssetReserveProof(1, leafA, proofA);
        assertTrue(valid);

        // Tampered leaf amount fails
        leafA.balanceE8 = 99_999 * 1e8;
        bool invalid = registry.verifyAssetReserveProof(1, leafA, proofA);
        assertFalse(invalid);
    }

    function test_ChallengeAndResolve() public {
        bytes32 liabRoot = keccak256("liab");
        bytes32 assetRoot = keccak256("asset");
        uint256 expiry = block.timestamp + 1 hours;

        bytes32 digest = registry.getEpochDigest(liabRoot, assetRoot, 100, 150, 1, expiry);
        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(custodianPk1, digest);
        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(custodianPk2, digest);

        bytes[] memory signatures = new bytes[](2);
        signatures[0] = abi.encodePacked(r1, s1, v1);
        signatures[1] = abi.encodePacked(r2, s2, v2);

        vm.prank(auditor);
        registry.publishAttestedEpoch(liabRoot, assetRoot, 100, 150, "report", expiry, signatures);

        // Challenge epoch
        registry.challengeEpoch(1, "Discrepancy in custodian cold vault balance");
        (bool solvent, ) = registry.isSolvent(1);
        assertFalse(solvent); // Challenged epoch considered not solvent

        // Auditor resolves challenge (dismissed/upheld=false)
        vm.prank(auditor);
        registry.resolveChallenge(1, false);
        (bool solventAfter, ) = registry.isSolvent(1);
        assertTrue(solventAfter);
    }
}
