// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {ReentrancyGuardUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";
import {PausableUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ISettlementGuaranteeFund} from "../interfaces/ISettlementGuaranteeFund.sol";

/**
 * @title SettlementGuaranteeFund
 * @notice Immutable, programmable Settlement Guarantee Fund (SGF) & Default Waterfall facility
 *         on Hyperledger Besu (QBFT consensus) for the Growww financial exchange.
 * @dev Implements the statutory SEBI / CPMI-IOSCO Default Waterfall sequence:
 *      Tranche 1: Defaulter Initial & Variation Margin balances.
 *      Tranche 2: Defaulter Core SGF contribution quotas.
 *      Tranche 3: Clearing Corporation (CC) Skin-in-the-Game dedicated capital.
 *      Tranche 4: Mutualized non-defaulting member SGF corpus (pro-rata slashed).
 *      Tranche 5: CC Emergency reserve capital.
 *
 * Invariants:
 *  1. Non-bypassable sequence: Tranche N cannot be executed if Tranche N-1 still has available funds.
 *  2. Clearing Corporation dedicated capital (Tranche 3) must be fully drawn before non-defaulting
 *     members' SGF contributions (Tranche 4) can be slashed.
 *  3. Invariant token backing: Sum of tracked allocations equals contract balance at all times.
 *  4. Zero PII on-chain: Only pseudonymous hashes (`memberIdHash`), wallets, and numeric balances are stored.
 */
contract SettlementGuaranteeFund is
    Initializable,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    ReentrancyGuardUpgradeable,
    PausableUpgradeable,
    ISettlementGuaranteeFund
{
    using SafeERC20 for IERC20;

    // Roles
    bytes32 public constant RISK_COMMITTEE_ROLE = keccak256("RISK_COMMITTEE_ROLE");
    bytes32 public constant SETTLEMENT_OPERATOR_ROLE = keccak256("SETTLEMENT_OPERATOR_ROLE");
    bytes32 public constant EMERGENCY_GUARDIAN_ROLE = keccak256("EMERGENCY_GUARDIAN_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    uint256 public constant BPS_DENOMINATOR = 10000;

    // Default assessment multiplier: 200% (2x peak SGF contribution)
    uint256 public defaultAssessmentCapMultiplierBps;

    // Accounting mappings
    mapping(bytes32 => MemberAllocation) public memberAllocations;
    mapping(bytes32 => address) public memberWallets;
    bytes32[] private _memberIds;
    mapping(bytes32 => bool) private _isRegisteredMember;

    // Historical peak SGF deposits and custom caps for assessment calculations
    mapping(bytes32 => uint256) public peakSGFDeposit;
    mapping(bytes32 => uint256) public memberCustomAssessmentCap;

    // Per-token segregated corpus accounting
    mapping(address => uint256) public totalSGFMemberDeposits;
    mapping(address => uint256) public totalLockedMargins;
    mapping(address => uint256) public ccSkinInTheGame;
    mapping(address => uint256) public ccEmergencyReserves;

    // Default tracking
    mapping(bytes32 => DefaultRecord) public defaultRecords;
    bytes32[] private _defaultIds;

    // Assessment calls: defaultId => memberIdHash => AssessmentCall
    mapping(bytes32 => mapping(bytes32 => AssessmentCall)) public assessmentCalls;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initializes the SettlementGuaranteeFund proxy.
     * @param admin Initial administrative multi-sig governance address.
     * @param riskCommittee Address authorized for risk committee decisions (defaults, assessments).
     * @param settlementOperator Address authorized to execute waterfall steps and clearing operations.
     * @param emergencyGuardian Address authorized to pause/unpause contract in emergencies.
     */
    function initialize(
        address admin,
        address riskCommittee,
        address settlementOperator,
        address emergencyGuardian
    ) external initializer {
        if (
            admin == address(0) ||
            riskCommittee == address(0) ||
            settlementOperator == address(0) ||
            emergencyGuardian == address(0)
        ) {
            revert InvalidAddress();
        }

        __AccessControl_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();
        __Pausable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        _grantRole(RISK_COMMITTEE_ROLE, riskCommittee);
        _grantRole(SETTLEMENT_OPERATOR_ROLE, settlementOperator);
        _grantRole(EMERGENCY_GUARDIAN_ROLE, emergencyGuardian);

        defaultAssessmentCapMultiplierBps = 20000; // 200% (2x)
    }

    // =========================================================================
    // Collateral Deposit & Allocation Logic
    // =========================================================================

    /**
     * @notice Deposits member core SGF contribution into Tranche 2/4 pool.
     * @param memberIdHash Pseudonymous identifier hash of the clearing member.
     * @param token Address of approved tokenized cash (e.g. eINR CBDC).
     * @param amount Amount of tokens to deposit.
     */
    function depositMemberSGF(
        bytes32 memberIdHash,
        address token,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (amount == 0) revert ZeroDepositAmount();
        if (token == address(0)) revert InvalidAddress();
        if (memberIdHash == bytes32(0)) revert InvalidAddress();

        MemberAllocation storage alloc = memberAllocations[memberIdHash];
        if (alloc.isDefaulted) revert MemberAlreadyDefaulted(memberIdHash);

        if (!_isRegisteredMember[memberIdHash]) {
            _registerMember(memberIdHash, msg.sender, token);
        } else if (alloc.depositToken != token) {
            revert TokenMismatch(alloc.depositToken, token);
        }

        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);

        alloc.sgfDepositBalance += amount;
        alloc.lastUpdatedTimestamp = block.timestamp;
        totalSGFMemberDeposits[token] += amount;

        if (alloc.sgfDepositBalance > peakSGFDeposit[memberIdHash]) {
            peakSGFDeposit[memberIdHash] = alloc.sgfDepositBalance;
        }

        emit SGFDepositReceived(memberIdHash, token, amount);
    }

    /**
     * @notice Deposits initial or variation margin balance for a clearing member into Tranche 1.
     * @param memberIdHash Pseudonymous identifier hash of the clearing member.
     * @param token Address of approved tokenized cash.
     * @param amount Amount of margin tokens to lock.
     */
    function depositMemberMargin(
        bytes32 memberIdHash,
        address token,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (amount == 0) revert ZeroDepositAmount();
        if (token == address(0)) revert InvalidAddress();
        if (memberIdHash == bytes32(0)) revert InvalidAddress();

        MemberAllocation storage alloc = memberAllocations[memberIdHash];
        if (alloc.isDefaulted) revert MemberAlreadyDefaulted(memberIdHash);

        if (!_isRegisteredMember[memberIdHash]) {
            _registerMember(memberIdHash, msg.sender, token);
        } else if (alloc.depositToken != token) {
            revert TokenMismatch(alloc.depositToken, token);
        }

        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);

        alloc.lockedMarginBalance += amount;
        alloc.lastUpdatedTimestamp = block.timestamp;
        totalLockedMargins[token] += amount;

        emit MarginDepositReceived(memberIdHash, token, amount);
    }

    /**
     * @notice Deposits dedicated Clearing Corporation (CC) skin-in-the-game capital (Tranche 3).
     * @param token Address of approved settlement token.
     * @param amount Amount of capital to deposit.
     */
    function depositCCContribution(
        address token,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (amount == 0) revert ZeroDepositAmount();
        if (token == address(0)) revert InvalidAddress();

        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);
        ccSkinInTheGame[token] += amount;

        emit CCContributionDeposited(token, amount);
    }

    /**
     * @notice Deposits Clearing Corporation emergency reserve capital (Tranche 5).
     * @param token Address of approved settlement token.
     * @param amount Amount of emergency reserve capital to deposit.
     */
    function depositCCReserves(
        address token,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        if (amount == 0) revert ZeroDepositAmount();
        if (token == address(0)) revert InvalidAddress();

        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);
        ccEmergencyReserves[token] += amount;

        emit CCReservesDeposited(token, amount);
    }

    // =========================================================================
    // Default Declaration Module
    // =========================================================================

    /**
     * @notice Declares a clearing member in default, freezing their assets and initializing waterfall.
     * @param defaultId Unique identifier for this default incident.
     * @param defaulterIdHash Identifier hash of the defaulting clearing member.
     * @param defaultAmount Total settlement shortfall amount in token wei.
     */
    function declareMemberDefault(
        bytes32 defaultId,
        bytes32 defaulterIdHash,
        uint256 defaultAmount
    ) external override onlyRole(RISK_COMMITTEE_ROLE) whenNotPaused {
        if (defaultId == bytes32(0) || defaulterIdHash == bytes32(0)) revert InvalidAddress();
        if (defaultAmount == 0) revert ZeroDepositAmount();

        DefaultRecord storage record = defaultRecords[defaultId];
        if (record.status != DefaultStatus.UNSPECIFIED) {
            revert DefaultAlreadyDeclared(defaultId);
        }

        if (!_isRegisteredMember[defaulterIdHash]) {
            revert MemberNotFound(defaulterIdHash);
        }

        MemberAllocation storage alloc = memberAllocations[defaulterIdHash];
        if (alloc.isDefaulted) {
            revert MemberAlreadyDefaulted(defaulterIdHash);
        }

        // Freeze defaulter's membership
        alloc.isDefaulted = true;
        alloc.lastUpdatedTimestamp = block.timestamp;

        // Initialize default record
        defaultRecords[defaultId] = DefaultRecord({
            defaultId: defaultId,
            defaulterIdHash: defaulterIdHash,
            totalDefaultAmount: defaultAmount,
            totalRecoveredAmount: 0,
            currentTranche: TrancheType.DEFAULTER_MARGINS,
            status: DefaultStatus.DECLARED,
            declaredTimestamp: block.timestamp,
            resolvedTimestamp: 0
        });

        _defaultIds.push(defaultId);

        emit DefaultDeclared(defaultId, defaulterIdHash, defaultAmount);
    }

    // =========================================================================
    // Default Waterfall Execution Engine
    // =========================================================================

    /**
     * @notice Executes a specific tranche step in the statutory default waterfall sequence.
     * @param defaultId Unique identifier for the default incident.
     * @param tranche Waterfall tranche to draw down.
     * @param drawAmount Amount requested to draw (0 means maximum available/needed for this tranche).
     * @param destinationVault Destination settlement pool (e.g. SettlementDvP) to receive liquidity.
     */
    function executeWaterfallStep(
        bytes32 defaultId,
        TrancheType tranche,
        uint256 drawAmount,
        address destinationVault
    ) public override nonReentrant whenNotPaused {
        _checkWaterfallRole();
        if (destinationVault == address(0)) revert InvalidAddress();

        DefaultRecord storage record = defaultRecords[defaultId];
        if (record.status == DefaultStatus.UNSPECIFIED) {
            revert DefaultNotFound(defaultId);
        }
        if (record.status == DefaultStatus.RESOLVED || record.status == DefaultStatus.FAILED) {
            revert DefaultAlreadyResolved(defaultId);
        }

        if (record.status == DefaultStatus.DECLARED) {
            record.status = DefaultStatus.PROCESSING_WATERFALL;
        }

        address token = memberAllocations[record.defaulterIdHash].depositToken;
        uint256 deficitRemaining = record.totalDefaultAmount - record.totalRecoveredAmount;
        if (deficitRemaining == 0) {
            _markDefaultResolved(record);
            return;
        }

        // Enforce strict non-bypassable sequence and synchronization
        _validateAndSyncTranche(record, tranche, token);

        if (tranche == TrancheType.DEFAULTER_MARGINS) {
            _executeTrancheDefaulterMargins(record, token, drawAmount, deficitRemaining, destinationVault);
        } else if (tranche == TrancheType.DEFAULTER_SGF) {
            _executeTrancheDefaulterSGF(record, token, drawAmount, deficitRemaining, destinationVault);
        } else if (tranche == TrancheType.CC_SKIN_IN_THE_GAME) {
            _executeTrancheCCContribution(record, token, drawAmount, deficitRemaining, destinationVault);
        } else if (tranche == TrancheType.POOLED_SGF) {
            _executeTranchePooledSGF(record, token, drawAmount, deficitRemaining, destinationVault);
        } else if (tranche == TrancheType.CC_RESERVES) {
            _executeTrancheCCReserves(record, token, drawAmount, deficitRemaining, destinationVault);
        }
    }

    // =========================================================================
    // Named Waterfall Step Entrypoints (Forwarders)
    // =========================================================================

    function slashDefaulterMargins(bytes32 defaultId, address settlementVault) external override {
        executeWaterfallStep(defaultId, TrancheType.DEFAULTER_MARGINS, 0, settlementVault);
    }

    function slashDefaulterSGF(bytes32 defaultId, address settlementVault) external override {
        executeWaterfallStep(defaultId, TrancheType.DEFAULTER_SGF, 0, settlementVault);
    }

    function slashCCContribution(bytes32 defaultId, address settlementVault, uint256 amount) external override {
        executeWaterfallStep(defaultId, TrancheType.CC_SKIN_IN_THE_GAME, amount, settlementVault);
    }

    function slashPooledSGF(bytes32 defaultId, address settlementVault, uint256 requiredAmount) external override {
        executeWaterfallStep(defaultId, TrancheType.POOLED_SGF, requiredAmount, settlementVault);
    }

    function slashCCReserves(bytes32 defaultId, address settlementVault, uint256 amount) external override {
        executeWaterfallStep(defaultId, TrancheType.CC_RESERVES, amount, settlementVault);
    }

    // =========================================================================
    // Default Resolution & Assessment Replenishment
    // =========================================================================

    /**
     * @notice Manually marks a default as resolved or failed if waterfall steps are completed.
     * @param defaultId Identifier for default incident.
     */
    function resolveDefault(bytes32 defaultId) external override {
        _checkWaterfallRole();
        DefaultRecord storage record = defaultRecords[defaultId];
        if (record.status == DefaultStatus.UNSPECIFIED) {
            revert DefaultNotFound(defaultId);
        }
        if (record.status == DefaultStatus.RESOLVED || record.status == DefaultStatus.FAILED) {
            revert DefaultAlreadyResolved(defaultId);
        }

        uint256 remainingDeficit = 0;
        if (record.totalRecoveredAmount < record.totalDefaultAmount) {
            remainingDeficit = record.totalDefaultAmount - record.totalRecoveredAmount;
            record.status = DefaultStatus.FAILED;
        } else {
            record.status = DefaultStatus.RESOLVED;
        }

        record.resolvedTimestamp = block.timestamp;
        emit DefaultResolved(defaultId, record.totalRecoveredAmount, remainingDeficit);
    }

    /**
     * @notice Issues a statutory assessment call to replenish mutualized SGF corpus after default.
     * @param defaultId Incident identifier that triggered assessment.
     * @param memberIdHash Solvent clearing member to assess.
     * @param assessedAmount Assessment quota requested.
     * @param deadline Unix timestamp deadline for replenishment.
     */
    function issueAssessmentCall(
        bytes32 defaultId,
        bytes32 memberIdHash,
        uint256 assessedAmount,
        uint256 deadline
    ) external override onlyRole(RISK_COMMITTEE_ROLE) whenNotPaused {
        if (defaultRecords[defaultId].status == DefaultStatus.UNSPECIFIED) {
            revert DefaultNotFound(defaultId);
        }
        if (!_isRegisteredMember[memberIdHash]) {
            revert MemberNotFound(memberIdHash);
        }
        if (memberAllocations[memberIdHash].isDefaulted) {
            revert MemberAlreadyDefaulted(memberIdHash);
        }
        if (deadline <= block.timestamp) {
            revert AssessmentDeadlinePassed(deadline, block.timestamp);
        }
        if (assessedAmount == 0) {
            revert ZeroDepositAmount();
        }

        uint256 cap = getMemberAssessmentCap(memberIdHash);
        if (assessedAmount > cap) {
            revert AssessmentExceedsCap(memberIdHash, assessedAmount);
        }

        assessmentCalls[defaultId][memberIdHash] = AssessmentCall({
            defaultId: defaultId,
            memberIdHash: memberIdHash,
            assessedAmount: assessedAmount,
            paidAmount: 0,
            deadline: deadline,
            isFulfilled: false
        });

        emit AssessmentCallIssued(defaultId, memberIdHash, assessedAmount, deadline);
    }

    /**
     * @notice Replenishes an assessment call by depositing settlement tokens.
     * @param defaultId Incident identifier.
     * @param memberIdHash Clearing member paying assessment.
     * @param amount Amount of tokens being paid.
     */
    function replenishAssessment(
        bytes32 defaultId,
        bytes32 memberIdHash,
        uint256 amount
    ) external override nonReentrant whenNotPaused {
        AssessmentCall storage call = assessmentCalls[defaultId][memberIdHash];
        if (call.assessedAmount == 0) {
            revert AssessmentNotFound(defaultId, memberIdHash);
        }
        if (call.isFulfilled) {
            revert AssessmentAlreadyFulfilled(defaultId, memberIdHash);
        }
        if (block.timestamp > call.deadline) {
            revert AssessmentDeadlinePassed(call.deadline, block.timestamp);
        }
        if (amount == 0) {
            revert ZeroDepositAmount();
        }

        address token = memberAllocations[memberIdHash].depositToken;
        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);

        call.paidAmount += amount;
        if (call.paidAmount >= call.assessedAmount) {
            call.isFulfilled = true;
        }

        MemberAllocation storage alloc = memberAllocations[memberIdHash];
        alloc.sgfDepositBalance += amount;
        alloc.lastUpdatedTimestamp = block.timestamp;
        totalSGFMemberDeposits[token] += amount;

        if (alloc.sgfDepositBalance > peakSGFDeposit[memberIdHash]) {
            peakSGFDeposit[memberIdHash] = alloc.sgfDepositBalance;
        }

        emit AssessmentReplenished(defaultId, memberIdHash, amount);
        emit SGFDepositReceived(memberIdHash, token, amount);
    }

    // =========================================================================
    // Solvent Member Collateral Withdrawals
    // =========================================================================

    /**
     * @notice Withdraws solvent member core SGF deposits.
     * @param memberIdHash Member identifier hash.
     * @param amount Amount of tokens to withdraw.
     * @param recipient Recipient address.
     */
    function withdrawMemberSGF(
        bytes32 memberIdHash,
        uint256 amount,
        address recipient
    ) external override nonReentrant whenNotPaused {
        _validateWithdrawalCaller(memberIdHash);
        if (amount == 0) revert ZeroDepositAmount();
        if (recipient == address(0)) revert InvalidAddress();

        MemberAllocation storage alloc = memberAllocations[memberIdHash];
        if (alloc.isDefaulted) revert MemberAlreadyDefaulted(memberIdHash);
        if (alloc.sgfDepositBalance < amount) {
            revert InsufficientMemberBalance(memberIdHash, alloc.sgfDepositBalance, amount);
        }

        address token = alloc.depositToken;
        alloc.sgfDepositBalance -= amount;
        alloc.lastUpdatedTimestamp = block.timestamp;
        totalSGFMemberDeposits[token] -= amount;

        IERC20(token).safeTransfer(recipient, amount);

        emit MemberSGFWithdrawn(memberIdHash, token, amount, recipient);
    }

    /**
     * @notice Withdraws solvent member locked margin balance.
     * @param memberIdHash Member identifier hash.
     * @param amount Amount of tokens to withdraw.
     * @param recipient Recipient address.
     */
    function withdrawMemberMargin(
        bytes32 memberIdHash,
        uint256 amount,
        address recipient
    ) external override nonReentrant whenNotPaused {
        _validateWithdrawalCaller(memberIdHash);
        if (amount == 0) revert ZeroDepositAmount();
        if (recipient == address(0)) revert InvalidAddress();

        MemberAllocation storage alloc = memberAllocations[memberIdHash];
        if (alloc.isDefaulted) revert MemberAlreadyDefaulted(memberIdHash);
        if (alloc.lockedMarginBalance < amount) {
            revert InsufficientMemberBalance(memberIdHash, alloc.lockedMarginBalance, amount);
        }

        address token = alloc.depositToken;
        alloc.lockedMarginBalance -= amount;
        alloc.lastUpdatedTimestamp = block.timestamp;
        totalLockedMargins[token] -= amount;

        IERC20(token).safeTransfer(recipient, amount);

        emit MemberMarginWithdrawn(memberIdHash, token, amount, recipient);
    }

    // =========================================================================
    // Governance & Parameter Administration
    // =========================================================================

    function setMemberAssessmentCap(
        bytes32 memberIdHash,
        uint256 newCap
    ) external onlyRole(RISK_COMMITTEE_ROLE) {
        if (!_isRegisteredMember[memberIdHash]) revert MemberNotFound(memberIdHash);
        memberCustomAssessmentCap[memberIdHash] = newCap;
        emit AssessmentCapUpdated(memberIdHash, newCap);
    }

    function setDefaultAssessmentCapMultiplierBps(
        uint256 newMultiplierBps
    ) external onlyRole(DEFAULT_ADMIN_ROLE) {
        defaultAssessmentCapMultiplierBps = newMultiplierBps;
    }

    function pause() external onlyRole(EMERGENCY_GUARDIAN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(EMERGENCY_GUARDIAN_ROLE) {
        _unpause();
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    // =========================================================================
    // View Functions
    // =========================================================================

    function getMemberAllocation(bytes32 memberIdHash) external view override returns (MemberAllocation memory) {
        return memberAllocations[memberIdHash];
    }

    function getDefaultRecord(bytes32 defaultId) external view override returns (DefaultRecord memory) {
        return defaultRecords[defaultId];
    }

    function getTotalSGFCorpus(address token) external view override returns (uint256 totalMemberDeposits, uint256 ccContribution) {
        totalMemberDeposits = totalSGFMemberDeposits[token];
        ccContribution = ccSkinInTheGame[token];
    }

    function getAssessmentCall(bytes32 defaultId, bytes32 memberIdHash) external view override returns (AssessmentCall memory) {
        return assessmentCalls[defaultId][memberIdHash];
    }

    function getMemberAssessmentCap(bytes32 memberIdHash) public view override returns (uint256) {
        if (memberCustomAssessmentCap[memberIdHash] > 0) {
            return memberCustomAssessmentCap[memberIdHash];
        }
        uint256 base = peakSGFDeposit[memberIdHash];
        if (base == 0) {
            base = memberAllocations[memberIdHash].sgfDepositBalance;
        }
        return (base * defaultAssessmentCapMultiplierBps) / BPS_DENOMINATOR;
    }

    function getSolventMemberCount(address token) external view override returns (uint256 count) {
        for (uint256 i = 0; i < _memberIds.length; i++) {
            MemberAllocation storage alloc = memberAllocations[_memberIds[i]];
            if (!alloc.isDefaulted && alloc.depositToken == token && alloc.sgfDepositBalance > 0) {
                count++;
            }
        }
    }

    function getTrackedAllocations(address token) external view override returns (
        uint256 memberSGFDeposits,
        uint256 memberLockedMargins,
        uint256 ccSkinCapital,
        uint256 ccEmergencyCapital,
        uint256 totalTracked
    ) {
        memberSGFDeposits = totalSGFMemberDeposits[token];
        memberLockedMargins = totalLockedMargins[token];
        ccSkinCapital = ccSkinInTheGame[token];
        ccEmergencyCapital = ccEmergencyReserves[token];
        totalTracked = memberSGFDeposits + memberLockedMargins + ccSkinCapital + ccEmergencyCapital;
    }

    function getAllMemberIds() external view returns (bytes32[] memory) {
        return _memberIds;
    }

    // =========================================================================
    // Internal Waterfall Logic
    // =========================================================================

    function _executeTrancheDefaulterMargins(
        DefaultRecord storage record,
        address token,
        uint256 drawAmount,
        uint256 deficitRemaining,
        address destinationVault
    ) internal {
        MemberAllocation storage alloc = memberAllocations[record.defaulterIdHash];
        uint256 available = alloc.lockedMarginBalance;
        uint256 amountToDraw;

        if (drawAmount == 0 || drawAmount >= available) {
            amountToDraw = available < deficitRemaining ? available : deficitRemaining;
        } else {
            amountToDraw = drawAmount < deficitRemaining ? drawAmount : deficitRemaining;
        }

        if (amountToDraw > 0) {
            alloc.lockedMarginBalance -= amountToDraw;
            totalLockedMargins[token] -= amountToDraw;
            alloc.lastUpdatedTimestamp = block.timestamp;
            IERC20(token).safeTransfer(destinationVault, amountToDraw);
        }

        record.totalRecoveredAmount += amountToDraw;
        emit TrancheSlashed(record.defaultId, TrancheType.DEFAULTER_MARGINS, amountToDraw, destinationVault);

        if (record.totalRecoveredAmount >= record.totalDefaultAmount) {
            _markDefaultResolved(record);
        } else if (alloc.lockedMarginBalance == 0) {
            record.currentTranche = TrancheType.DEFAULTER_SGF;
        }
    }

    function _executeTrancheDefaulterSGF(
        DefaultRecord storage record,
        address token,
        uint256 drawAmount,
        uint256 deficitRemaining,
        address destinationVault
    ) internal {
        MemberAllocation storage alloc = memberAllocations[record.defaulterIdHash];
        uint256 available = alloc.sgfDepositBalance;
        uint256 amountToDraw;

        if (drawAmount == 0 || drawAmount >= available) {
            amountToDraw = available < deficitRemaining ? available : deficitRemaining;
        } else {
            amountToDraw = drawAmount < deficitRemaining ? drawAmount : deficitRemaining;
        }

        if (amountToDraw > 0) {
            alloc.sgfDepositBalance -= amountToDraw;
            totalSGFMemberDeposits[token] -= amountToDraw;
            alloc.lastUpdatedTimestamp = block.timestamp;
            IERC20(token).safeTransfer(destinationVault, amountToDraw);
        }

        record.totalRecoveredAmount += amountToDraw;
        emit TrancheSlashed(record.defaultId, TrancheType.DEFAULTER_SGF, amountToDraw, destinationVault);

        if (record.totalRecoveredAmount >= record.totalDefaultAmount) {
            _markDefaultResolved(record);
        } else if (alloc.sgfDepositBalance == 0) {
            record.currentTranche = TrancheType.CC_SKIN_IN_THE_GAME;
        }
    }

    function _executeTrancheCCContribution(
        DefaultRecord storage record,
        address token,
        uint256 drawAmount,
        uint256 deficitRemaining,
        address destinationVault
    ) internal {
        uint256 available = ccSkinInTheGame[token];
        uint256 amountToDraw;

        if (drawAmount == 0 || drawAmount >= available) {
            amountToDraw = available < deficitRemaining ? available : deficitRemaining;
        } else {
            amountToDraw = drawAmount < deficitRemaining ? drawAmount : deficitRemaining;
        }

        if (amountToDraw > 0) {
            ccSkinInTheGame[token] -= amountToDraw;
            IERC20(token).safeTransfer(destinationVault, amountToDraw);
        }

        record.totalRecoveredAmount += amountToDraw;
        emit TrancheSlashed(record.defaultId, TrancheType.CC_SKIN_IN_THE_GAME, amountToDraw, destinationVault);

        if (record.totalRecoveredAmount >= record.totalDefaultAmount) {
            _markDefaultResolved(record);
        } else if (ccSkinInTheGame[token] == 0) {
            record.currentTranche = TrancheType.POOLED_SGF;
        }
    }

    function _executeTranchePooledSGF(
        DefaultRecord storage record,
        address token,
        uint256 drawAmount,
        uint256 deficitRemaining,
        address destinationVault
    ) internal {
        uint256 totalSolventSGF = 0;
        for (uint256 i = 0; i < _memberIds.length; i++) {
            MemberAllocation storage alloc = memberAllocations[_memberIds[i]];
            if (!alloc.isDefaulted && alloc.depositToken == token) {
                totalSolventSGF += alloc.sgfDepositBalance;
            }
        }

        uint256 amountToDraw;
        if (drawAmount == 0 || drawAmount >= totalSolventSGF) {
            amountToDraw = totalSolventSGF < deficitRemaining ? totalSolventSGF : deficitRemaining;
        } else {
            amountToDraw = drawAmount < deficitRemaining ? drawAmount : deficitRemaining;
        }

        uint256 actualTotalSlashed = 0;
        if (amountToDraw > 0 && totalSolventSGF > 0) {
            uint256 remainingToSlash = amountToDraw;
            uint256 remainingSolventDeposits = totalSolventSGF;

            for (uint256 i = 0; i < _memberIds.length; i++) {
                bytes32 mId = _memberIds[i];
                MemberAllocation storage alloc = memberAllocations[mId];
                if (alloc.isDefaulted || alloc.depositToken != token || alloc.sgfDepositBalance == 0) {
                    continue;
                }

                uint256 memberBalance = alloc.sgfDepositBalance;
                uint256 memberShare;

                // For the last solvent member or exact match, deduct exact remaining dust
                if (remainingSolventDeposits == memberBalance || remainingToSlash >= remainingSolventDeposits) {
                    memberShare = remainingToSlash < memberBalance ? remainingToSlash : memberBalance;
                } else {
                    memberShare = (remainingToSlash * memberBalance) / remainingSolventDeposits;
                    if (memberShare > memberBalance) {
                        memberShare = memberBalance;
                    }
                }

                if (memberShare > 0) {
                    alloc.sgfDepositBalance -= memberShare;
                    alloc.lastUpdatedTimestamp = block.timestamp;
                    remainingToSlash -= memberShare;
                    remainingSolventDeposits -= memberBalance;
                    actualTotalSlashed += memberShare;
                    emit ProRataMemberSlashed(record.defaultId, mId, memberShare);
                } else {
                    remainingSolventDeposits -= memberBalance;
                }
            }

            totalSGFMemberDeposits[token] -= actualTotalSlashed;
            IERC20(token).safeTransfer(destinationVault, actualTotalSlashed);
        }

        record.totalRecoveredAmount += actualTotalSlashed;
        emit TrancheSlashed(record.defaultId, TrancheType.POOLED_SGF, actualTotalSlashed, destinationVault);

        if (record.totalRecoveredAmount >= record.totalDefaultAmount) {
            _markDefaultResolved(record);
        } else {
            // If deficit still remains, advance to CC reserves
            record.currentTranche = TrancheType.CC_RESERVES;
        }
    }

    function _executeTrancheCCReserves(
        DefaultRecord storage record,
        address token,
        uint256 drawAmount,
        uint256 deficitRemaining,
        address destinationVault
    ) internal {
        uint256 available = ccEmergencyReserves[token];
        uint256 amountToDraw;

        if (drawAmount == 0 || drawAmount >= available) {
            amountToDraw = available < deficitRemaining ? available : deficitRemaining;
        } else {
            amountToDraw = drawAmount < deficitRemaining ? drawAmount : deficitRemaining;
        }

        if (amountToDraw > 0) {
            ccEmergencyReserves[token] -= amountToDraw;
            IERC20(token).safeTransfer(destinationVault, amountToDraw);
        }

        record.totalRecoveredAmount += amountToDraw;
        emit TrancheSlashed(record.defaultId, TrancheType.CC_RESERVES, amountToDraw, destinationVault);

        if (record.totalRecoveredAmount >= record.totalDefaultAmount) {
            _markDefaultResolved(record);
        } else if (ccEmergencyReserves[token] == 0) {
            uint256 finalDeficit = record.totalDefaultAmount - record.totalRecoveredAmount;
            record.status = DefaultStatus.FAILED;
            record.resolvedTimestamp = block.timestamp;
            emit DefaultResolved(record.defaultId, record.totalRecoveredAmount, finalDeficit);
        }
    }

    function _validateAndSyncTranche(
        DefaultRecord storage record,
        TrancheType providedTranche,
        address token
    ) internal {
        MemberAllocation storage alloc = memberAllocations[record.defaulterIdHash];

        // 1. Advance through empty tranches if the provided tranche is further down
        if (record.currentTranche == TrancheType.DEFAULTER_MARGINS && alloc.lockedMarginBalance == 0) {
            if (providedTranche > TrancheType.DEFAULTER_MARGINS) {
                record.currentTranche = TrancheType.DEFAULTER_SGF;
            }
        }
        if (record.currentTranche == TrancheType.DEFAULTER_SGF && alloc.sgfDepositBalance == 0) {
            if (providedTranche > TrancheType.DEFAULTER_SGF) {
                record.currentTranche = TrancheType.CC_SKIN_IN_THE_GAME;
            }
        }
        if (record.currentTranche == TrancheType.CC_SKIN_IN_THE_GAME && ccSkinInTheGame[token] == 0) {
            if (providedTranche > TrancheType.CC_SKIN_IN_THE_GAME) {
                record.currentTranche = TrancheType.POOLED_SGF;
            }
        }

        // Regulatory check: CC skin-in-the-game must be fully exhausted before Pooled SGF can be drawn
        if (providedTranche == TrancheType.POOLED_SGF && ccSkinInTheGame[token] > 0) {
            revert InvalidTrancheSequence(TrancheType.CC_SKIN_IN_THE_GAME, TrancheType.POOLED_SGF);
        }

        if (providedTranche != record.currentTranche) {
            revert InvalidTrancheSequence(record.currentTranche, providedTranche);
        }
    }

    function _markDefaultResolved(DefaultRecord storage record) internal {
        record.status = DefaultStatus.RESOLVED;
        record.resolvedTimestamp = block.timestamp;
        emit DefaultResolved(record.defaultId, record.totalRecoveredAmount, 0);
    }

    function _registerMember(bytes32 memberIdHash, address wallet, address token) internal {
        _isRegisteredMember[memberIdHash] = true;
        _memberIds.push(memberIdHash);
        memberWallets[memberIdHash] = wallet;

        memberAllocations[memberIdHash] = MemberAllocation({
            memberIdHash: memberIdHash,
            depositToken: token,
            sgfDepositBalance: 0,
            lockedMarginBalance: 0,
            lastUpdatedTimestamp: block.timestamp,
            isDefaulted: false
        });
    }

    function _checkWaterfallRole() internal view {
        if (!hasRole(SETTLEMENT_OPERATOR_ROLE, msg.sender) && !hasRole(RISK_COMMITTEE_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }
    }

    function _validateWithdrawalCaller(bytes32 memberIdHash) internal view {
        if (!_isRegisteredMember[memberIdHash]) revert MemberNotFound(memberIdHash);
        if (
            msg.sender != memberWallets[memberIdHash] &&
            !hasRole(SETTLEMENT_OPERATOR_ROLE, msg.sender) &&
            !hasRole(RISK_COMMITTEE_ROLE, msg.sender)
        ) {
            revert UnauthorizedCaller(msg.sender);
        }
    }

    // Storage gap reservation for future upgrades
    uint256[43] private __gap;
}
