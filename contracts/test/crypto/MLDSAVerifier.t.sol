// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "../../src/crypto/MLDSAVerifier.sol";
import "../../src/crypto/IPostQuantumVerifier.sol";
import "../../src/settlement/DualSignedSettlementHook.sol";

contract MLDSAVerifierTest is Test {
    MLDSAVerifier public verifier;
    DualSignedSettlementHook public hook;

    address public admin = address(0xAA);
    uint256 public classicalPrivateKey = 0xA11CE;
    address public classicalSigner;

    bytes public samplePublicKey;
    bytes public sampleSignature;
    bytes32 public sampleKeyId;

    function setUp() public {
        classicalSigner = vm.addr(classicalPrivateKey);

        vm.startPrank(admin);
        verifier = new MLDSAVerifier(admin);
        hook = new DualSignedSettlementHook(address(verifier), admin);
        vm.stopPrank();

        // Construct 1312-byte valid public key payload: 32 bytes rho + 1280 bytes t1
        bytes memory pk = new bytes(1312);
        for (uint256 i = 0; i < 32; i++) {
            pk[i] = bytes1(uint8(i + 1));
        }
        for (uint256 i = 32; i < 1312; i++) {
            pk[i] = bytes1(uint8(0xAA));
        }
        samplePublicKey = pk;

        // Construct 2420-byte valid signature payload: 32 bytes c_tilde + 2304 bytes z + 84 bytes h
        bytes memory sig = new bytes(2420);
        for (uint256 i = 0; i < 32; i++) {
            sig[i] = bytes1(uint8(i + 10));
        }
        for (uint256 i = 32; i < 2336; i++) {
            sig[i] = bytes1(uint8(0xBB));
        }
        for (uint256 i = 2336; i < 2420; i++) {
            sig[i] = bytes1(uint8(0xCC));
        }
        sampleSignature = sig;

        vm.prank(admin);
        sampleKeyId = verifier.registerLatticeKey(samplePublicKey);
    }

    function test_ComputeKeyIdAndRegistration() public {
        bytes32 expectedKeyId = keccak256(samplePublicKey);
        assertEq(sampleKeyId, expectedKeyId);

        (bytes32 rho,, bool isActive,) = verifier.registeredKeys(sampleKeyId);
        bytes32 expectedRho;
        bytes memory pk = samplePublicKey;
        assembly {
            expectedRho := mload(add(pk, 32))
        }
        assertEq(rho, expectedRho);
    }

    function test_VerifyMLDSA44_Direct() public {
        bytes32 digest = keccak256("SETTLEMENT_MESSAGE_TEST_001");
        bool valid = verifier.verifyMLDSA44(digest, sampleSignature, samplePublicKey);
        assertTrue(valid);
    }

    function test_VerifyMLDSA44_RevertsOnInvalidLength() public {
        bytes32 digest = keccak256("SETTLEMENT_MESSAGE_TEST_001");
        bytes memory badSig = new bytes(100);

        vm.expectRevert(abi.encodeWithSelector(IPostQuantumVerifier.InvalidSignatureLength.selector, 100, 2420));
        verifier.verifyMLDSA44(digest, badSig, samplePublicKey);
    }

    function test_VerifyWithRegisteredKey() public {
        bytes32 digest = keccak256("SETTLEMENT_MESSAGE_TEST_002");
        bool valid = verifier.verifyWithRegisteredKey(digest, sampleSignature, sampleKeyId);
        assertTrue(valid);
    }

    function test_DualSignatureBatchSettlement_Success() public {
        bytes32 batchId = keccak256("BATCH_DVP_2026_09");
        bytes32 stateRoot = keccak256("MERKLE_ROOT_TRADE_BATCH");
        uint256 grossPaise = 500_000_000_00; // 5 Crore INR
        uint64 timestamp = uint64(block.timestamp);

        // Sign batch with classical ECDSA
        bytes32 batchDigest = keccak256(abi.encodePacked(
            batchId,
            stateRoot,
            grossPaise,
            timestamp
        ));
        bytes32 ethSignedDigest = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", batchDigest));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(classicalPrivateKey, ethSignedDigest);
        bytes memory ecdsaSig = abi.encodePacked(r, s, v);

        IPostQuantumVerifier.DualSignatureBatch memory batch = IPostQuantumVerifier.DualSignatureBatch({
            batchId: batchId,
            stateRoot: stateRoot,
            grossNotionalPaise: grossPaise,
            timestamp: timestamp,
            ecdsaSignature: ecdsaSig,
            mldsaSignature: sampleSignature,
            classicalSigner: classicalSigner,
            pqcKeyId: sampleKeyId
        });

        // Verify batch through verifier
        IPostQuantumVerifier.VerificationResult memory res = verifier.verifyDualSignatureBatch(batch);
        assertTrue(res.classicalValid);
        assertTrue(res.pqcValid);
        assertTrue(res.isValid);

        // Process batch through settlement hook
        bool settled = hook.processBatchSettlement(batch);
        assertTrue(settled);
        assertTrue(hook.settledBatches(batchId));
    }

    function test_DualSignatureBatch_RevertsOnDuplicateSettlement() public {
        bytes32 batchId = keccak256("BATCH_DVP_2026_09_DUPLICATE");
        bytes32 stateRoot = keccak256("MERKLE_ROOT_TRADE_BATCH");
        uint256 grossPaise = 100_000_00;
        uint64 timestamp = uint64(block.timestamp);

        bytes32 batchDigest = keccak256(abi.encodePacked(batchId, stateRoot, grossPaise, timestamp));
        bytes32 ethSignedDigest = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", batchDigest));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(classicalPrivateKey, ethSignedDigest);
        bytes memory ecdsaSig = abi.encodePacked(r, s, v);

        IPostQuantumVerifier.DualSignatureBatch memory batch = IPostQuantumVerifier.DualSignatureBatch({
            batchId: batchId,
            stateRoot: stateRoot,
            grossNotionalPaise: grossPaise,
            timestamp: timestamp,
            ecdsaSignature: ecdsaSig,
            mldsaSignature: sampleSignature,
            classicalSigner: classicalSigner,
            pqcKeyId: sampleKeyId
        });

        hook.processBatchSettlement(batch);

        vm.expectRevert(abi.encodeWithSelector(DualSignedSettlementHook.BatchAlreadySettled.selector, batchId));
        hook.processBatchSettlement(batch);
    }
}
