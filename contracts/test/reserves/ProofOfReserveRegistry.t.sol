// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "../../src/reserves/ProofOfReserveRegistry.sol";

contract ProofOfReserveRegistryTest is Test {
    ProofOfReserveRegistry public registry;

    uint256 internal custodianPk = 0xA11CE;
    uint256 internal auditorPk = 0xB0B;
    address internal custodianSigner;
    address internal auditorSigner;
    address internal circuitBreaker = address(0x9999);

    string constant TEST_ISIN = "INE002A01018"; // Reliance Industries Limited

    function setUp() public {
        custodianSigner = vm.addr(custodianPk);
        auditorSigner = vm.addr(auditorPk);

        registry = new ProofOfReserveRegistry(
            custodianSigner,
            auditorSigner,
            circuitBreaker
        );
    }

    function test_InitialSigners() public view {
        assertEq(registry.custodianSigner(), custodianSigner);
        assertEq(registry.auditorSigner(), auditorSigner);
        assertEq(registry.globalMintPaused(), false);
    }

    function test_PublishAttestationSuccess() public {
        uint256 depository = 1_000_000;
        uint256 tokenSupply = 1_000_000;
        bytes32 merkleRoot = keccak256("merkle_tree_root_epoch_1");

        bytes32 digest = keccak256(
            abi.encodePacked(
                TEST_ISIN,
                depository,
                tokenSupply,
                merkleRoot,
                block.chainid,
                address(registry)
            )
        );
        bytes32 ethDigest = MessageHashUtils.toEthSignedMessageHash(digest);

        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(custodianPk, ethDigest);
        bytes memory custodianSig = abi.encodePacked(r1, s1, v1);

        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(auditorPk, ethDigest);
        bytes memory auditorSig = abi.encodePacked(r2, s2, v2);

        uint256 epoch = registry.publishAttestation(
            TEST_ISIN,
            depository,
            tokenSupply,
            merkleRoot,
            custodianSig,
            auditorSig
        );

        assertEq(epoch, 1);
        assertEq(registry.latestEpoch(TEST_ISIN), 1);
        assertEq(registry.globalMintPaused(), false);
    }

    function test_PublishAttestationDetectsInsolvency() public {
        uint256 depository = 900_000;
        uint256 tokenSupply = 1_000_000; // 100k deficit!
        bytes32 merkleRoot = keccak256("merkle_tree_root_epoch_deficit");

        bytes32 digest = keccak256(
            abi.encodePacked(
                TEST_ISIN,
                depository,
                tokenSupply,
                merkleRoot,
                block.chainid,
                address(registry)
            )
        );
        bytes32 ethDigest = MessageHashUtils.toEthSignedMessageHash(digest);

        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(custodianPk, ethDigest);
        bytes memory custodianSig = abi.encodePacked(r1, s1, v1);

        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(auditorPk, ethDigest);
        bytes memory auditorSig = abi.encodePacked(r2, s2, v2);

        registry.publishAttestation(
            TEST_ISIN,
            depository,
            tokenSupply,
            merkleRoot,
            custodianSig,
            auditorSig
        );

        // Invariant: global mint paused when deficit exists!
        assertTrue(registry.globalMintPaused());
    }

    function test_VerifyInvestorInclusion() public {
        bytes32 investorCommitment = keccak256("user_pan_hash_12345");
        bytes32 salt = keccak256("random_secure_salt_abcde");
        uint256 balance = 50_000;

        bytes32 leaf1 = keccak256(abi.encodePacked(investorCommitment, salt, balance));
        bytes32 leaf2 = keccak256(abi.encodePacked("other_user_commitment", "salt2", uint256(100_000)));

        // Sort leaves for deterministic canonical tree
        bytes32 root;
        bytes32[] memory proof = new bytes32[](1);
        if (leaf1 <= leaf2) {
            root = keccak256(abi.encodePacked(leaf1, leaf2));
            proof[0] = leaf2;
        } else {
            root = keccak256(abi.encodePacked(leaf2, leaf1));
            proof[0] = leaf2;
        }

        bytes32 digest = keccak256(
            abi.encodePacked(
                TEST_ISIN,
                uint256(200_000),
                uint256(150_000),
                root,
                block.chainid,
                address(registry)
            )
        );
        bytes32 ethDigest = MessageHashUtils.toEthSignedMessageHash(digest);

        (uint8 v1, bytes32 r1, bytes32 s1) = vm.sign(custodianPk, ethDigest);
        bytes memory custodianSig = abi.encodePacked(r1, s1, v1);

        (uint8 v2, bytes32 r2, bytes32 s2) = vm.sign(auditorPk, ethDigest);
        bytes memory auditorSig = abi.encodePacked(r2, s2, v2);

        uint256 epoch = registry.publishAttestation(
            TEST_ISIN,
            200_000,
            150_000,
            root,
            custodianSig,
            auditorSig
        );

        bool isValid = registry.verifyInvestorInclusion(
            TEST_ISIN,
            epoch,
            investorCommitment,
            salt,
            balance,
            proof
        );

        assertTrue(isValid);
    }
}
