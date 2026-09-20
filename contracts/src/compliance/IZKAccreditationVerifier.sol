// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title IZKAccreditationVerifier
 * @notice Interface for Groth16 Zero-Knowledge Verification of Investor Accreditation & Jurisdiction
 * @dev Enforces SEBI & IFSCA accredited investor rules on-chain with zero-knowledge data privacy
 */
interface IZKAccreditationVerifier {

    enum AccreditationTier {
        NONE,
        SEBI_RETAIL_ACCREDITED,        // Net worth >= INR 7.5 Cr / Income >= INR 2 Cr
        SEBI_LARGE_VALUE_ACCREDITED,   // Minimum commitment >= INR 25 Cr (AIF LVAI)
        IFSCA_INDIVIDUAL_ACCREDITED,   // Net worth >= USD 1,000,000 (excluding primary residence)
        IFSCA_INSTITUTIONAL_ACCREDITED // Net assets >= USD 5,000,000
    }

    struct AccreditationAttestation {
        AccreditationTier tier;
        uint64 verifiedAt;
        uint64 expiresAt;
        bytes32 nullifierHash;
        bytes32 merkleRootUsed;
        bool isValid;
    }

    struct ProofPoints {
        uint256[2] a;
        uint256[2][2] b;
        uint256[2] c;
    }

    // Events
    event AccreditationRegistered(
        address indexed investor,
        AccreditationTier indexed tier,
        bytes32 indexed nullifierHash,
        uint64 expiresAt,
        bytes32 merkleRoot
    );

    event AccreditationRevoked(address indexed investor, bytes32 indexed nullifierHash, string reason);
    event CredentialRootAdded(bytes32 indexed root, uint256 validUntil);
    event CredentialRootRevoked(bytes32 indexed root);
    event AgencyKeyWhitelisted(address indexed agencyPubKey, string agencyName);
    event AgencyKeyRemoved(address indexed agencyPubKey);

    // Custom Errors
    error InvalidProof();
    error NullifierAlreadySpent(bytes32 nullifierHash);
    error InvalidCredentialRoot(bytes32 merkleRoot);
    error RootExpired(bytes32 merkleRoot, uint256 expirationTime);
    error TierMismatch(uint8 requestedTier, uint8 provedTier);
    error AttestationExpired(address investor, uint64 expiresAt);
    error AttestationNotFound(address investor);
    error UnauthorizedCaller(address caller);
    error ZeroAddress();

    function verifyProof(
        uint256[2] calldata a,
        uint256[2][2] calldata b,
        uint256[2] calldata c,
        uint256[4] calldata input
    ) external view returns (bool success);

    function registerAccreditation(
        ProofPoints calldata proof,
        uint256[4] calldata publicInputs,
        uint64 attestationDuration
    ) external returns (bool success);

    function isAccredited(address investor, AccreditationTier requiredTier) external view returns (bool isEligible);

    function getAttestation(address investor) external view returns (AccreditationAttestation memory attestation);

    function isNullifierConsumed(bytes32 nullifierHash) external view returns (bool spent);
}
