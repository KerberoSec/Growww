// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title ISettlementGuaranteeFund
 * @notice Standardized interface for the On-Chain Settlement Guarantee Fund & Default Waterfall
 * @dev Enforces statutory SEBI / CPMI-IOSCO Default Waterfall priority:
 *      Tranche 1: Defaulter Initial & Variation Margin balances
 *      Tranche 2: Defaulter Core SGF contribution quota
 *      Tranche 3: Clearing Corporation (CC) Skin-in-the-Game dedicated capital
 *      Tranche 4: Mutualized non-defaulting member SGF corpus (pro-rata slashed)
 *      Tranche 5: CC Emergency reserve capital
 */
interface ISettlementGuaranteeFund {
    enum TrancheType {
        DEFAULTER_MARGINS,
        DEFAULTER_SGF,
        CC_SKIN_IN_THE_GAME,
        POOLED_SGF,
        CC_RESERVES
    }

    enum DefaultStatus {
        UNSPECIFIED,
        DECLARED,
        PROCESSING_WATERFALL,
        RESOLVED,
        FAILED
    }

    struct MemberAllocation {
        bytes32 memberIdHash;
        address depositToken;
        uint256 sgfDepositBalance;
        uint256 lockedMarginBalance;
        uint256 lastUpdatedTimestamp;
        bool isDefaulted;
    }

    struct DefaultRecord {
        bytes32 defaultId;
        bytes32 defaulterIdHash;
        uint256 totalDefaultAmount;
        uint256 totalRecoveredAmount;
        TrancheType currentTranche;
        DefaultStatus status;
        uint256 declaredTimestamp;
        uint256 resolvedTimestamp;
    }

    struct AssessmentCall {
        bytes32 defaultId;
        bytes32 memberIdHash;
        uint256 assessedAmount;
        uint256 paidAmount;
        uint256 deadline;
        bool isFulfilled;
    }

    // Events
    event SGFDepositReceived(bytes32 indexed memberIdHash, address indexed token, uint256 amount);
    event CCContributionDeposited(address indexed token, uint256 amount);
    event DefaultDeclared(bytes32 indexed defaultId, bytes32 indexed defaulterIdHash, uint256 defaultAmount);
    event TrancheSlashed(bytes32 indexed defaultId, TrancheType indexed tranche, uint256 amountDrawn, address destinationVault);
    event ProRataMemberSlashed(bytes32 indexed defaultId, bytes32 indexed memberIdHash, uint256 amountSlashed);
    event DefaultResolved(bytes32 indexed defaultId, uint256 totalRecovered, uint256 remainingDeficit);
    event AssessmentCallIssued(bytes32 indexed defaultId, bytes32 indexed memberIdHash, uint256 amountDue, uint256 deadline);

    // Supplementary Events
    event MarginDepositReceived(bytes32 indexed memberIdHash, address indexed token, uint256 amount);
    event CCReservesDeposited(address indexed token, uint256 amount);
    event AssessmentReplenished(bytes32 indexed defaultId, bytes32 indexed memberIdHash, uint256 amount);
    event MemberSGFWithdrawn(bytes32 indexed memberIdHash, address indexed token, uint256 amount, address recipient);
    event MemberMarginWithdrawn(bytes32 indexed memberIdHash, address indexed token, uint256 amount, address recipient);
    event AssessmentCapUpdated(bytes32 indexed memberIdHash, uint256 newCap);

    // Custom Errors
    error UnauthorizedCaller(address caller);
    error DefaultAlreadyDeclared(bytes32 defaultId);
    error DefaultNotFound(bytes32 defaultId);
    error InvalidTrancheSequence(TrancheType expected, TrancheType provided);
    error InsufficientContractBalance(address token, uint256 available, uint256 requested);
    error MemberAlreadyDefaulted(bytes32 memberIdHash);
    error ZeroDepositAmount();
    error AssessmentExceedsCap(bytes32 memberIdHash, uint256 amount);
    error DefaultNotActive(bytes32 defaultId);
    error DefaultAlreadyResolved(bytes32 defaultId);
    error InvalidAddress();
    error MemberNotFound(bytes32 memberIdHash);
    error InsufficientMemberBalance(bytes32 memberIdHash, uint256 available, uint256 requested);
    error AssessmentDeadlinePassed(uint256 deadline, uint256 currentTimestamp);
    error AssessmentAlreadyFulfilled(bytes32 defaultId, bytes32 memberIdHash);
    error AssessmentNotFound(bytes32 defaultId, bytes32 memberIdHash);
    error TokenMismatch(address expected, address provided);

    // Core Operational Functions
    function depositMemberSGF(bytes32 memberIdHash, address token, uint256 amount) external;
    function depositCCContribution(address token, uint256 amount) external;
    function declareMemberDefault(bytes32 defaultId, bytes32 defaulterIdHash, uint256 defaultAmount) external;
    function executeWaterfallStep(bytes32 defaultId, TrancheType tranche, uint256 drawAmount, address destinationVault) external;
    function resolveDefault(bytes32 defaultId) external;
    function issueAssessmentCall(bytes32 defaultId, bytes32 memberIdHash, uint256 assessedAmount, uint256 deadline) external;

    // Supplementary Operational Functions
    function depositMemberMargin(bytes32 memberIdHash, address token, uint256 amount) external;
    function depositCCReserves(address token, uint256 amount) external;
    function slashDefaulterMargins(bytes32 defaultId, address settlementVault) external;
    function slashDefaulterSGF(bytes32 defaultId, address settlementVault) external;
    function slashCCContribution(bytes32 defaultId, address settlementVault, uint256 amount) external;
    function slashPooledSGF(bytes32 defaultId, address settlementVault, uint256 requiredAmount) external;
    function slashCCReserves(bytes32 defaultId, address settlementVault, uint256 amount) external;
    function replenishAssessment(bytes32 defaultId, bytes32 memberIdHash, uint256 amount) external;
    function withdrawMemberSGF(bytes32 memberIdHash, uint256 amount, address recipient) external;
    function withdrawMemberMargin(bytes32 memberIdHash, uint256 amount, address recipient) external;

    // View Functions
    function getMemberAllocation(bytes32 memberIdHash) external view returns (MemberAllocation memory);
    function getDefaultRecord(bytes32 defaultId) external view returns (DefaultRecord memory);
    function getTotalSGFCorpus(address token) external view returns (uint256 totalMemberDeposits, uint256 ccContribution);
    function getAssessmentCall(bytes32 defaultId, bytes32 memberIdHash) external view returns (AssessmentCall memory);
    function getMemberAssessmentCap(bytes32 memberIdHash) external view returns (uint256);
    function getSolventMemberCount(address token) external view returns (uint256);
    function getTrackedAllocations(address token) external view returns (
        uint256 memberSGFDeposits,
        uint256 memberLockedMargins,
        uint256 ccSkinCapital,
        uint256 ccEmergencyCapital,
        uint256 totalTracked
    );
}
