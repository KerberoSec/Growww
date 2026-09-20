// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "../../src/compliance/ZKAccreditationVerifier.sol";
import "../../src/compliance/IZKAccreditationVerifier.sol";

contract ZKAccreditationVerifierTest is Test {
    ZKAccreditationVerifier public verifier;
    address public admin = address(0xAD);
    address public agency = address(0xCA);
    address public investor1 = address(0x101);
    address public investor2 = address(0x102);

    bytes32 public sampleMerkleRoot = keccak256("CREDENTIAL_TREE_ROOT_2026_Q3");
    bytes32 public sampleNullifier = keccak256("NULLIFIER_INVESTOR_1_EPOCH_1");

    function setUp() public {
        vm.startPrank(admin);
        verifier = new ZKAccreditationVerifier(admin);
        verifier.setAgencyAuthorization(agency, true, "CDSL Ventures Ltd");
        vm.stopPrank();

        vm.prank(agency);
        verifier.addCredentialRoot(sampleMerkleRoot, block.timestamp + 30 days);
    }

    function test_InitialConfiguration() public {
        assertTrue(verifier.isAgencyAuthorized(agency));
        assertTrue(verifier.isMerkleRootValid(sampleMerkleRoot));
        assertFalse(verifier.isNullifierConsumed(sampleNullifier));
    }

    function test_RegisterAccreditation_Success() public {
        IZKAccreditationVerifier.ProofPoints memory proof;
        proof.a = [uint256(1), uint256(2)];
        proof.b = [[uint256(3), uint256(4)], [uint256(5), uint256(6)]];
        proof.c = [uint256(7), uint256(8)];

        uint256[4] memory publicInputs;
        publicInputs[0] = uint256(sampleMerkleRoot);
        publicInputs[1] = uint256(sampleNullifier);
        publicInputs[2] = uint256(IZKAccreditationVerifier.AccreditationTier.SEBI_RETAIL_ACCREDITED);
        publicInputs[3] = uint256(356); // India ISO 3166-1

        vm.prank(investor1);
        bool success = verifier.registerAccreditation(proof, publicInputs, 365 days);
        assertTrue(success);

        // Check attestation
        IZKAccreditationVerifier.AccreditationAttestation memory att = verifier.getAttestation(investor1);
        assertTrue(att.isValid);
        assertEq(uint8(att.tier), uint8(IZKAccreditationVerifier.AccreditationTier.SEBI_RETAIL_ACCREDITED));
        assertEq(att.nullifierHash, sampleNullifier);

        // Verify eligibility checks
        assertTrue(verifier.isAccredited(investor1, IZKAccreditationVerifier.AccreditationTier.SEBI_RETAIL_ACCREDITED));
        assertFalse(verifier.isAccredited(investor1, IZKAccreditationVerifier.AccreditationTier.SEBI_LARGE_VALUE_ACCREDITED));
        assertTrue(verifier.isNullifierConsumed(sampleNullifier));
    }

    function test_RegisterAccreditation_RevertsOnDuplicateNullifier() public {
        IZKAccreditationVerifier.ProofPoints memory proof;
        proof.a = [uint256(1), uint256(2)];
        proof.b = [[uint256(3), uint256(4)], [uint256(5), uint256(6)]];
        proof.c = [uint256(7), uint256(8)];

        uint256[4] memory publicInputs;
        publicInputs[0] = uint256(sampleMerkleRoot);
        publicInputs[1] = uint256(sampleNullifier);
        publicInputs[2] = uint256(IZKAccreditationVerifier.AccreditationTier.SEBI_RETAIL_ACCREDITED);
        publicInputs[3] = uint256(356);

        vm.prank(investor1);
        verifier.registerAccreditation(proof, publicInputs, 365 days);

        // Replay attempt with same nullifier
        vm.prank(investor2);
        vm.expectRevert(abi.encodeWithSelector(IZKAccreditationVerifier.NullifierAlreadySpent.selector, sampleNullifier));
        verifier.registerAccreditation(proof, publicInputs, 365 days);
    }

    function test_RegisterAccreditation_RevertsOnInvalidRoot() public {
        IZKAccreditationVerifier.ProofPoints memory proof;
        proof.a = [uint256(1), uint256(2)];
        proof.b = [[uint256(3), uint256(4)], [uint256(5), uint256(6)]];
        proof.c = [uint256(7), uint256(8)];

        bytes32 rogueRoot = keccak256("ROGUE_ROOT");
        uint256[4] memory publicInputs;
        publicInputs[0] = uint256(rogueRoot);
        publicInputs[1] = uint256(sampleNullifier);
        publicInputs[2] = uint256(IZKAccreditationVerifier.AccreditationTier.SEBI_RETAIL_ACCREDITED);
        publicInputs[3] = uint256(356);

        vm.prank(investor1);
        vm.expectRevert(abi.encodeWithSelector(IZKAccreditationVerifier.InvalidCredentialRoot.selector, rogueRoot));
        verifier.registerAccreditation(proof, publicInputs, 365 days);
    }

    function test_RevokeAccreditation_Administrative() public {
        IZKAccreditationVerifier.ProofPoints memory proof;
        proof.a = [uint256(1), uint256(2)];
        proof.b = [[uint256(3), uint256(4)], [uint256(5), uint256(6)]];
        proof.c = [uint256(7), uint256(8)];

        uint256[4] memory publicInputs;
        publicInputs[0] = uint256(sampleMerkleRoot);
        publicInputs[1] = uint256(sampleNullifier);
        publicInputs[2] = uint256(IZKAccreditationVerifier.AccreditationTier.SEBI_RETAIL_ACCREDITED);
        publicInputs[3] = uint256(356);

        vm.prank(investor1);
        verifier.registerAccreditation(proof, publicInputs, 365 days);
        assertTrue(verifier.isAccredited(investor1, IZKAccreditationVerifier.AccreditationTier.SEBI_RETAIL_ACCREDITED));

        // Admin revokes
        vm.prank(admin);
        verifier.revokeAccreditation(investor1, "SEBI Enforcement Notice");

        assertFalse(verifier.isAccredited(investor1, IZKAccreditationVerifier.AccreditationTier.SEBI_RETAIL_ACCREDITED));
    }
}
