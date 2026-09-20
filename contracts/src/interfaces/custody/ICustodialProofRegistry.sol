// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title ICustodialProofRegistry
 * @notice Interface for on-chain custodial proof verification and solvency Merkle registry.
 * @dev Anchors cryptographic Proof of Reserves (PoR) and Proof of Liabilities (PoL) with multi-custodian quorum.
 */
interface ICustodialProofRegistry {
    struct CustodialAttestation {
        bytes32 epochRoot;
        bytes32 assetId;
        uint256 totalReservesE8;
        uint256 nonce;
        uint256 timestamp;
        uint256 expiry;
    }

    struct SolvencyEpochRecord {
        bytes32 liabilitiesMerkleRoot;
        bytes32 assetsMerkleRoot;
        uint256 totalLiabilitiesE8;
        uint256 totalReservesE8;
        uint256 timestamp;
        uint256 custodianAttestationsCount;
        string auditReportUri;
        bool isChallenged;
    }

    struct AssetReserveLeaf {
        address custodian;
        bytes32 assetId;
        uint256 balanceE8;
        bytes32 depositoryRef;
        uint256 timestamp;
    }

    // =========================================================================
    // Events
    // =========================================================================

    event CustodianAdded(address indexed custodian);
    event CustodianRemoved(address indexed custodian);
    event CustodianThresholdUpdated(uint256 oldThreshold, uint256 newThreshold);
    event AuditorUpdated(address indexed oldAuditor, address indexed newAuditor);
    event EpochAttested(
        uint256 indexed epochId,
        bytes32 liabilitiesRoot,
        bytes32 assetsRoot,
        uint256 totalLiabilitiesE8,
        uint256 totalReservesE8,
        uint256 timestamp
    );
    event ReserveLeafVerified(uint256 indexed epochId, bytes32 indexed assetId, address indexed custodian, uint256 amount);
    event LiabilityLeafVerified(uint256 indexed epochId, bytes32 indexed leafHash);
    event SolvencyChallenged(uint256 indexed epochId, address indexed challenger, string reason);
    event SolvencyChallengeResolved(uint256 indexed epochId, bool upheld);

    // =========================================================================
    // Errors
    // =========================================================================

    error UnauthorizedCaller();
    error InvalidZeroAddress();
    error InvalidThreshold(uint256 threshold, uint256 custodiansCount);
    error CustodianAlreadyAdded(address custodian);
    error CustodianNotFound(address custodian);
    error InsufficientReserves(uint256 totalReserves, uint256 requiredReserves);
    error InsufficientCustodianSignatures(uint256 provided, uint256 required);
    error DuplicateSignature(address signer);
    error InvalidSignature(address signer);
    error SignatureExpired(uint256 expiry, uint256 currentTimestamp);
    error EpochDoesNotExist(uint256 epochId);
    error EpochAlreadyChallenged(uint256 epochId);
    error MerkleProofInvalid();
    error InvalidLeafParameters();

    // =========================================================================
    // View Functions
    // =========================================================================

    function currentEpoch() external view returns (uint256);
    function auditor() external view returns (address);
    function custodianThreshold() external view returns (uint256);
    function isCustodian(address custodian) external view returns (bool);
    function getEpoch(uint256 epochId) external view returns (SolvencyEpochRecord memory);

    function verifyAssetReserveProof(
        uint256 epochId,
        AssetReserveLeaf calldata leaf,
        bytes32[] calldata proof
    ) external view returns (bool);

    function verifyLiabilityProof(
        uint256 epochId,
        bytes32 leafHash,
        bytes32[] calldata proof
    ) external view returns (bool);

    function isSolvent(uint256 epochId) external view returns (bool solvent, uint256 coverageRatioBps);

    // =========================================================================
    // State-Changing Functions
    // =========================================================================

    function publishAttestedEpoch(
        bytes32 liabilitiesRoot,
        bytes32 assetsRoot,
        uint256 totalLiabilitiesE8,
        uint256 totalReservesE8,
        string calldata reportUri,
        uint256 expiry,
        bytes[] calldata custodianSignatures
    ) external returns (uint256 epochId);

    function challengeEpoch(uint256 epochId, string calldata reason) external;

    function resolveChallenge(uint256 epochId, bool upheld) external;
}
