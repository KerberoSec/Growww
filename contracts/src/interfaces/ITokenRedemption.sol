// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title ITokenRedemption
 * @notice Interface for two-phase token redemption and custodial asset release.
 * @dev Manages token escrow, custodial EIP-712 settlement proofs, and irreversible burning.
 */
interface ITokenRedemption {
    enum RedemptionStatus {
        NONE,
        ESCROWED,
        SETTLED_BURNED,
        CANCELLED
    }

    struct RedemptionRequest {
        bytes32 id;
        address investor;
        address tokenAddress;
        uint256 amount;
        uint64 requestTimestamp;
        uint64 expiryTimestamp;
        RedemptionStatus status;
        bytes32 dematDebitRef;
    }

    event RedemptionInitiated(
        bytes32 indexed id,
        address indexed investor,
        address indexed token,
        uint256 amount,
        uint64 expiryTimestamp
    );

    event RedemptionCompleted(
        bytes32 indexed id,
        bytes32 indexed dematDebitRef,
        uint256 timestamp
    );

    event RedemptionCancelled(
        bytes32 indexed id,
        string reason
    );

    event TokenSupportUpdated(address indexed token, bool supported);

    // Errors
    error InvalidAmount();
    error TokenNotSupported(address token);
    error RequestNotFound(bytes32 id);
    error InvalidStatus(RedemptionStatus currentStatus);
    error RequestExpired(uint64 expiryTimestamp, uint256 currentTimestamp);
    error RequestNotExpired(uint64 expiryTimestamp, uint256 currentTimestamp);
    error InvalidSignature();
    error InvalidZeroAddress();
    error ArrayLengthMismatch(uint256 expected, uint256 actual);

    // User Operations
    function requestRedemption(address token, uint256 amount) external returns (bytes32 redemptionId);

    // Custodian / Relayer Settlement Operations
    function executeRedemptionWithProof(
        bytes32 redemptionId,
        bytes32 dematDebitRef,
        bytes calldata custodianSignature
    ) external;

    function executeBatchRedemption(
        bytes32[] calldata redemptionIds,
        bytes32[] calldata dematDebitRefs,
        bytes[] calldata signatures
    ) external;

    // Failure / Expiry Recovery
    function cancelExpiredRedemption(bytes32 redemptionId) external;
    function adminCancelRedemption(bytes32 redemptionId, string calldata reason) external;

    // View Functions
    function getRequest(bytes32 redemptionId) external view returns (RedemptionRequest memory);
}
