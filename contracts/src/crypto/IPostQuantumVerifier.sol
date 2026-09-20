// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

uint8 constant ML_DSA_K = 4;
uint8 constant ML_DSA_L = 4;
uint8 constant ML_DSA_D = 13;
uint32 constant ML_DSA_GAMMA1 = 131072;
uint32 constant ML_DSA_GAMMA2 = 95232;
uint8 constant ML_DSA_TAU = 39;
uint8 constant ML_DSA_BETA = 78;
uint8 constant ML_DSA_OMEGA = 80;

uint16 constant ML_DSA_44_PUBLIC_KEY_BYTES = 1312;
uint16 constant ML_DSA_44_SIGNATURE_BYTES = 2420;

/**
 * @title IPostQuantumVerifier
 * @notice Canonical interface for NIST FIPS 204 ML-DSA-44 (Dilithium2) lattice signature verification and dual-signature settlement.
 */
interface IPostQuantumVerifier {

    struct LatticePublicKey {
        bytes32 rho;                   // 32-byte seed for matrix A
        bytes t1Encoded;               // 1280-byte encoded high-order coefficient vector
        bool isActive;
        uint64 registeredAt;
    }

    struct LatticeSignature {
        bytes32 challengeSeed;         // 32-byte c_tilde
        bytes zEncoded;                // 2304-byte response vector
        bytes hintEncoded;             // 84-byte hint vector
    }

    struct DualSignatureBatch {
        bytes32 batchId;               // Unique settlement batch identifier
        bytes32 stateRoot;             // Merkle state root
        uint256 grossNotionalPaise;    // Gross trade consideration in paise
        uint64 timestamp;              // Batch execution timestamp
        bytes ecdsaSignature;          // 65-byte secp256k1 classical signature
        bytes mldsaSignature;          // 2420-byte ML-DSA-44 signature
        address classicalSigner;       // Expected authorized ECDSA signer address
        bytes32 pqcKeyId;              // Registered identifier of ML-DSA public key
    }

    struct VerificationResult {
        bool isValid;                  // Overall verification boolean
        bool classicalValid;           // Outcome of ECDSA check
        bool pqcValid;                 // Outcome of ML-DSA-44 check
        uint256 gasConsumed;           // Gas consumed during execution
        uint64 verifiedTimestamp;      // Timestamp of verification
    }

    error InvalidPublicKeyLength(uint256 actualLength, uint256 expectedLength);
    error InvalidSignatureLength(uint256 actualLength, uint256 expectedLength);
    error NormOutOfBounds(uint32 coefficientValue, uint32 allowedBound);
    error HintWeightExceeded(uint256 actualWeight, uint256 maximumWeight);
    error ChallengeMismatch(bytes32 expectedChallenge, bytes32 actualChallenge);
    error ClassicalSignatureFailed(address recoveredSigner, address expectedSigner);
    error LatticeKeyNotRegistered(bytes32 pqcKeyId);
    error LatticeKeyRevoked(bytes32 pqcKeyId);
    error InvalidBatchTimestamp(uint64 batchTimestamp, uint64 currentBlockTimestamp);
    error ZeroGrossConsideration();

    event LatticePublicKeyRegistered(
        bytes32 indexed pqcKeyId,
        address indexed authority,
        bytes32 indexed rho,
        uint64 registeredAt
    );

    event LatticePublicKeyRevoked(
        bytes32 indexed pqcKeyId,
        address indexed revokedBy,
        uint64 revokedAt
    );

    event SignatureVerifiedPQC(
        bytes32 indexed messageDigest,
        bytes32 indexed pqcKeyId,
        bool success,
        uint256 gasUsed
    );

    event DualSignatureBatchValidated(
        bytes32 indexed batchId,
        address indexed classicalSigner,
        bytes32 indexed pqcKeyId,
        uint256 grossNotionalPaise,
        uint256 platformFeePaise
    );

    function verifyMLDSA44(
        bytes32 messageDigest,
        bytes memory signatureBytes,
        bytes memory publicKeyBytes
    ) external view returns (bool isValid);

    function verifyWithRegisteredKey(
        bytes32 messageDigest,
        bytes calldata signatureBytes,
        bytes32 pqcKeyId
    ) external view returns (bool isValid);

    function verifyDualSignatureBatch(
        DualSignatureBatch calldata batch
    ) external view returns (VerificationResult memory result);

    function computeKeyId(bytes calldata publicKeyBytes) external pure returns (bytes32 pqcKeyId);
}
