// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import { Initializable } from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import { AccessControlUpgradeable } from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import { UUPSUpgradeable } from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import { ReentrancyGuardUpgradeable } from "@openzeppelin/contracts-upgradeable/utils/ReentrancyGuardUpgradeable.sol";

import { ICircuitBreakerHalt } from "../interfaces/ICircuitBreakerHalt.sol";

/**
 * @title CircuitBreakerHalt
 * @notice Multi-tier emergency pause and circuit breaker contract for Growww NBSE exchange on Hyperledger Besu.
 * Operates in compliance with SEBI Master Circular on Market-Wide Circuit Breakers (MWCB)
 * and IFSCA Market Infrastructure Regulations.
 *
 * Tiers:
 * - Tier 0: Single-ISIN halts (volatility collars, depository desync, oracle divergence).
 * - Tier 1: Sector / Asset Class halts (basket-level containment).
 * - Tier 2: Market-Wide Circuit Breakers (SEBI 10%, 15%, 20% drops with time-staged cooldowns & call-auctions).
 * - Tier 3: Global Emergency Freeze (ledger-wide operational shutdown).
 */
contract CircuitBreakerHalt is
    Initializable,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    ReentrancyGuardUpgradeable,
    ICircuitBreakerHalt
{
    // --- Roles ---
    bytes32 public constant EMERGENCY_GUARDIAN_ROLE = keccak256("EMERGENCY_GUARDIAN_ROLE");
    bytes32 public constant SURVEILLANCE_SENTINEL_ROLE = keccak256("SURVEILLANCE_SENTINEL_ROLE");
    bytes32 public constant ORACLE_SENTINEL_ROLE = keccak256("ORACLE_SENTINEL_ROLE");
    bytes32 public constant DEPOSITORY_SENTINEL_ROLE = keccak256("DEPOSITORY_SENTINEL_ROLE");
    bytes32 public constant GOVERNANCE_TIMELOCK_ROLE = keccak256("GOVERNANCE_TIMELOCK_ROLE");

    // --- Transient Storage Slot ---
    // keccak256("growww.circuitbreaker.transient.globalfreeze")
    bytes32 private constant GLOBAL_FREEZE_TRANSIENT_SLOT =
        0x8f2d650b86a8b792e3be851c23f18e9fa4cf9e3c7886470eb020bf2dbce19d85;

    // --- State Variables (Packed for Gas Minimization) ---
    // Tier 3
    bool public globalEmergencyFreeze;
    uint64 public globalFrozenAt;
    bytes32 public globalFreezeReasonHash;

    // Tier 2
    MarketWideHaltState public marketWideState;

    // Tier 1 & Tier 0 Registry
    mapping(bytes32 => bytes32) public isinToSector;
    mapping(bytes32 => bool) public sectorHalted;
    mapping(bytes32 => ISINHaltStatus) public isinHalt;

    // Configuration
    uint32 public defaultOracleHaltDuration;
    uint32 public maxHaltDuration;
    address public timelockController;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initializes the upgradeable contract with admin, timelock, and emergency guardian roles.
     * @param admin Exchange admin address
     * @param timelock Timelock controller address for delayed governance execution
     * @param emergencyGuardian Emergency guardian address
     */
    function initialize(
        address admin,
        address timelock,
        address emergencyGuardian
    ) external initializer {
        __AccessControl_init();
        __UUPSUpgradeable_init();
        __ReentrancyGuard_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);

        if (timelock != address(0)) {
            _grantRole(GOVERNANCE_TIMELOCK_ROLE, timelock);
            timelockController = timelock;
        }

        if (emergencyGuardian != address(0)) {
            _grantRole(EMERGENCY_GUARDIAN_ROLE, emergencyGuardian);
        }

        defaultOracleHaltDuration = 3600; // 1 hour default
        maxHaltDuration = 30 days;
    }

    // =========================================================================
    // Core Verification Views (Used by SettlementDvP and Token Contracts)
    // =========================================================================

    /**
     * @notice Hierarchically queries if an ISIN is halted across any tier.
     * Priority: Tier 3 (Global) -> Tier 2 (MWCB) -> Tier 1 (Sector) -> Tier 0 (ISIN).
     * @param isin ISIN identifier
     * @return halted True if halted at any tier
     * @return tier Active halt tier level
     * @return expiresAt Timestamp when halt expires (0 or type(uint64).max if indefinite)
     */
    function isHalted(bytes32 isin)
        public
        view
        override
        returns (bool halted, HaltTier tier, uint64 expiresAt)
    {
        // Tier 3: Global Emergency Freeze
        bool frozen;
        assembly {
            frozen := tload(GLOBAL_FREEZE_TRANSIENT_SLOT)
        }
        if (frozen || globalEmergencyFreeze) {
            return (true, HaltTier.TIER_3_GLOBAL, type(uint64).max);
        }

        // Tier 2: Market-Wide Circuit Breaker
        if (marketWideState.sessionState != MarketSessionState.NORMAL_TRADING) {
            return (true, HaltTier.TIER_2_MARKET_WIDE, marketWideState.cooldownExpiresAt);
        }

        // Tier 1: Sector Halt
        bytes32 sectorId = isinToSector[isin];
        if (sectorId != bytes32(0) && sectorHalted[sectorId]) {
            return (true, HaltTier.TIER_1_SECTOR, type(uint64).max);
        }

        // Tier 0: Single-ISIN Halt
        ISINHaltStatus storage status = isinHalt[isin];
        if (status.isHalted) {
            if (status.expiresAt > 0 && status.expiresAt != type(uint64).max && block.timestamp >= status.expiresAt) {
                // Automatic expiry without state change transaction
                return (false, HaltTier.NONE, 0);
            }
            return (true, HaltTier.TIER_0_ISIN, status.expiresAt);
        }

        return (false, HaltTier.NONE, 0);
    }

    /**
     * @notice Enforces that an ISIN is not halted at any tier, reverting with descriptive custom errors.
     * @param isin ISIN identifier
     */
    function requireNotHalted(bytes32 isin) external view override {
        // Tier 3: Global Emergency Freeze
        bool frozen;
        assembly {
            frozen := tload(GLOBAL_FREEZE_TRANSIENT_SLOT)
        }
        if (frozen || globalEmergencyFreeze) {
            revert GlobalFreezeActive();
        }

        // Tier 2: Market-Wide Circuit Breaker
        if (marketWideState.sessionState != MarketSessionState.NORMAL_TRADING) {
            revert MarketWideHaltActive(
                marketWideState.circuitLevel,
                marketWideState.cooldownExpiresAt
            );
        }

        // Tier 1: Sector Halt
        bytes32 sectorId = isinToSector[isin];
        if (sectorId != bytes32(0) && sectorHalted[sectorId]) {
            revert CircuitBreakerActive(isin, HaltTier.TIER_1_SECTOR, type(uint64).max);
        }

        // Tier 0: Single-ISIN Halt
        ISINHaltStatus storage status = isinHalt[isin];
        if (status.isHalted) {
            if (status.expiresAt == 0 || status.expiresAt == type(uint64).max || block.timestamp < status.expiresAt) {
                revert CircuitBreakerActive(isin, HaltTier.TIER_0_ISIN, status.expiresAt);
            }
        }
    }

    /**
     * @notice Returns whether platform-wide global emergency freeze is active.
     */
    function isGlobalFrozen() external view override returns (bool) {
        bool frozen;
        assembly {
            frozen := tload(GLOBAL_FREEZE_TRANSIENT_SLOT)
        }
        return frozen || globalEmergencyFreeze;
    }

    /**
     * @notice Returns the current market-wide circuit breaker state.
     */
    function getMarketWideState() external view override returns (MarketWideHaltState memory) {
        return marketWideState;
    }

    /**
     * @notice Returns the single-ISIN halt status struct.
     * @param isin ISIN identifier
     */
    function getISINHaltStatus(bytes32 isin) external view override returns (ISINHaltStatus memory) {
        return isinHalt[isin];
    }

    /**
     * @notice Returns whether a specific sector is halted.
     * @param sectorId Industry sector identifier
     */
    function isSectorHalted(bytes32 sectorId) external view override returns (bool) {
        return sectorHalted[sectorId];
    }

    /**
     * @notice Returns the sector mapped to an ISIN.
     * @param isin ISIN identifier
     */
    function getSectorForISIN(bytes32 isin) external view override returns (bytes32) {
        return isinToSector[isin];
    }

    // =========================================================================
    // Autonomous Sentinel Triggers (Tier 0 & Tier 2)
    // =========================================================================

    /**
     * @notice Trips Tier 0 halt immediately upon depository desynchronization between off-chain demat and on-chain supply.
     * @param isin ISIN identifier
     * @param mismatchProofHash Cryptographic hash of depository reconciliation divergence report
     */
    function triggerDepositoryHalt(
        bytes32 isin,
        bytes32 mismatchProofHash
    ) external override nonReentrant {
        if (!hasRole(DEPOSITORY_SENTINEL_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, DEPOSITORY_SENTINEL_ROLE);
        }

        ISINHaltStatus storage status = isinHalt[isin];
        if (status.isHalted && (status.expiresAt == 0 || status.expiresAt == type(uint64).max || block.timestamp < status.expiresAt)) {
            revert ISINAlreadyHalted(isin);
        }

        uint64 haltedAt = uint64(block.timestamp);
        uint64 expiresAt = type(uint64).max; // Depository mismatch requires custodial audit and governance override

        isinHalt[isin] = ISINHaltStatus({
            isHalted: true,
            reason: HaltReason.DEPOSITORY_DESYNC,
            haltedAt: haltedAt,
            expiresAt: expiresAt,
            detailsHash: mismatchProofHash,
            haltedBy: msg.sender
        });

        emit ISINHalted(
            isin,
            HaltReason.DEPOSITORY_DESYNC,
            haltedAt,
            expiresAt,
            mismatchProofHash,
            msg.sender
        );
    }

    /**
     * @notice Trips Tier 0 halt when oracle price feeds diverge beyond threshold boundaries.
     * @param isin ISIN identifier
     * @param deviationBps Basis points deviation detected by aggregator
     * @param detailsHash Incident dossier hash
     */
    function triggerOracleDivergenceHalt(
        bytes32 isin,
        uint256 deviationBps,
        bytes32 detailsHash
    ) external override nonReentrant {
        deviationBps; // Silence unused variable warning while preserving interface
        if (!hasRole(ORACLE_SENTINEL_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, ORACLE_SENTINEL_ROLE);
        }

        ISINHaltStatus storage status = isinHalt[isin];
        if (status.isHalted && (status.expiresAt == 0 || status.expiresAt == type(uint64).max || block.timestamp < status.expiresAt)) {
            revert ISINAlreadyHalted(isin);
        }

        uint64 haltedAt = uint64(block.timestamp);
        uint64 expiresAt = haltedAt + uint64(defaultOracleHaltDuration);

        isinHalt[isin] = ISINHaltStatus({
            isHalted: true,
            reason: HaltReason.ORACLE_DIVERGENCE,
            haltedAt: haltedAt,
            expiresAt: expiresAt,
            detailsHash: detailsHash,
            haltedBy: msg.sender
        });

        emit ISINHalted(
            isin,
            HaltReason.ORACLE_DIVERGENCE,
            haltedAt,
            expiresAt,
            detailsHash,
            msg.sender
        );
    }

    /**
     * @notice Trips Tier 0 halt when automated surveillance flags volatility collar breach or spoofing anomaly.
     * @param isin ISIN identifier
     * @param durationSeconds Duration of pause in seconds
     * @param detailsHash Incident dossier hash
     */
    function triggerSurveillanceVolatilityHalt(
        bytes32 isin,
        uint32 durationSeconds,
        bytes32 detailsHash
    ) external override nonReentrant {
        if (!hasRole(SURVEILLANCE_SENTINEL_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, SURVEILLANCE_SENTINEL_ROLE);
        }

        if (durationSeconds == 0 || durationSeconds > maxHaltDuration) {
            revert InvalidHaltDuration(durationSeconds);
        }

        ISINHaltStatus storage status = isinHalt[isin];
        if (status.isHalted && (status.expiresAt == 0 || status.expiresAt == type(uint64).max || block.timestamp < status.expiresAt)) {
            revert ISINAlreadyHalted(isin);
        }

        uint64 haltedAt = uint64(block.timestamp);
        uint64 expiresAt = haltedAt + uint64(durationSeconds);

        isinHalt[isin] = ISINHaltStatus({
            isHalted: true,
            reason: HaltReason.VOLATILITY_COLLAR,
            haltedAt: haltedAt,
            expiresAt: expiresAt,
            detailsHash: detailsHash,
            haltedBy: msg.sender
        });

        emit ISINHalted(
            isin,
            HaltReason.VOLATILITY_COLLAR,
            haltedAt,
            expiresAt,
            detailsHash,
            msg.sender
        );
    }

    /**
     * @notice Trips Tier 2 Market-Wide Circuit Breaker enforcing SEBI Master Circular rules (10%, 15%, 20%).
     * @param circuitLevel SEBI index circuit limit breached (10, 15, or 20)
     * @param indexId Benchmark index identifier (e.g. keccak256("NIFTY_50"))
     * @param indexLevelPaise Benchmark index value in paise (2 decimal precision)
     * @param sessionTimestamp Timestamp of breach (0 uses block.timestamp)
     */
    function triggerMarketWideCircuitBreaker(
        uint8 circuitLevel,
        bytes32 indexId,
        uint256 indexLevelPaise,
        uint64 sessionTimestamp
    ) external override nonReentrant {
        if (!hasRole(SURVEILLANCE_SENTINEL_ROLE, msg.sender) &&
            !hasRole(EMERGENCY_GUARDIAN_ROLE, msg.sender) &&
            !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, SURVEILLANCE_SENTINEL_ROLE);
        }

        // Do not allow overwriting with a lower or equal circuit breaker
        if (marketWideState.sessionState == MarketSessionState.MARKET_CLOSED_FOR_DAY) {
            revert MarketWideHaltActive(marketWideState.circuitLevel, marketWideState.cooldownExpiresAt);
        }
        if (marketWideState.circuitLevel >= circuitLevel && marketWideState.sessionState != MarketSessionState.NORMAL_TRADING) {
            revert MarketWideHaltActive(marketWideState.circuitLevel, marketWideState.cooldownExpiresAt);
        }

        uint64 effectiveTs = sessionTimestamp != 0 ? sessionTimestamp : uint64(block.timestamp);
        (uint64 cooldownDuration, uint64 auctionDuration, MarketSessionState nextState) =
            _calculateMWCBDurations(circuitLevel, effectiveTs);

        if (nextState == MarketSessionState.NORMAL_TRADING) {
            // 10% breach at or after 2:30 PM: No halt; trading continues with price collars
            emit MarketWideHaltTriggered(
                circuitLevel,
                uint64(block.timestamp),
                uint64(block.timestamp),
                uint64(block.timestamp),
                indexLevelPaise,
                indexId,
                msg.sender
            );
            return;
        }

        uint64 haltedAt = uint64(block.timestamp);
        uint64 cooldownExpiresAt = nextState == MarketSessionState.MARKET_CLOSED_FOR_DAY
            ? type(uint64).max
            : haltedAt + cooldownDuration;
        uint64 auctionExpiresAt = nextState == MarketSessionState.MARKET_CLOSED_FOR_DAY
            ? type(uint64).max
            : cooldownExpiresAt + auctionDuration;

        marketWideState = MarketWideHaltState({
            sessionState: nextState,
            circuitLevel: circuitLevel,
            haltedAt: haltedAt,
            cooldownExpiresAt: cooldownExpiresAt,
            auctionExpiresAt: auctionExpiresAt,
            triggerIndex: indexId,
            indexLevelPaise: indexLevelPaise
        });

        emit MarketWideHaltTriggered(
            circuitLevel,
            haltedAt,
            cooldownExpiresAt,
            auctionExpiresAt,
            indexLevelPaise,
            indexId,
            msg.sender
        );
    }

    // =========================================================================
    // Governance & Administrative Functions
    // =========================================================================

    /**
     * @notice Trips Tier 3 Global Emergency Freeze across the entire exchange ledger.
     * @param reasonHash Cryptographic hash of emergency justification dossier
     */
    function triggerGlobalEmergencyFreeze(bytes32 reasonHash) external override nonReentrant {
        if (!hasRole(EMERGENCY_GUARDIAN_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, EMERGENCY_GUARDIAN_ROLE);
        }

        globalEmergencyFreeze = true;
        globalFrozenAt = uint64(block.timestamp);
        globalFreezeReasonHash = reasonHash;

        assembly {
            tstore(GLOBAL_FREEZE_TRANSIENT_SLOT, 1)
        }

        emit GlobalEmergencyFreezeTriggered(uint64(block.timestamp), reasonHash, msg.sender);
    }

    /**
     * @notice Lifts Tier 3 Global Emergency Freeze. Strictly restricted to GOVERNANCE_TIMELOCK_ROLE or admin.
     */
    function liftGlobalEmergencyFreeze() external override nonReentrant {
        if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, GOVERNANCE_TIMELOCK_ROLE);
        }

        if (!globalEmergencyFreeze) {
            revert GlobalFreezeNotActive();
        }

        globalEmergencyFreeze = false;
        globalFrozenAt = 0;
        globalFreezeReasonHash = bytes32(0);

        assembly {
            tstore(GLOBAL_FREEZE_TRANSIENT_SLOT, 0)
        }

        emit GlobalEmergencyFreezeLifted(uint64(block.timestamp), msg.sender);
    }

    /**
     * @notice Manually pauses trading of a specific ISIN (Tier 0).
     */
    function haltISINManual(
        bytes32 isin,
        HaltReason reason,
        uint32 durationSeconds,
        bytes32 detailsHash
    ) external override nonReentrant {
        if (!hasRole(EMERGENCY_GUARDIAN_ROLE, msg.sender) &&
            !hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) &&
            !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, EMERGENCY_GUARDIAN_ROLE);
        }

        if (durationSeconds == 0 || durationSeconds > maxHaltDuration) {
            revert InvalidHaltDuration(durationSeconds);
        }

        ISINHaltStatus storage status = isinHalt[isin];
        if (status.isHalted && (status.expiresAt == 0 || status.expiresAt == type(uint64).max || block.timestamp < status.expiresAt)) {
            revert ISINAlreadyHalted(isin);
        }

        uint64 haltedAt = uint64(block.timestamp);
        uint64 expiresAt = haltedAt + uint64(durationSeconds);

        isinHalt[isin] = ISINHaltStatus({
            isHalted: true,
            reason: reason,
            haltedAt: haltedAt,
            expiresAt: expiresAt,
            detailsHash: detailsHash,
            haltedBy: msg.sender
        });

        emit ISINHalted(
            isin,
            reason,
            haltedAt,
            expiresAt,
            detailsHash,
            msg.sender
        );
    }

    /**
     * @notice Resumes trading of a single ISIN.
     * If cooldown has not elapsed, execution requires GOVERNANCE_TIMELOCK_ROLE (or admin).
     * If cooldown has elapsed, EMERGENCY_GUARDIAN_ROLE, GOVERNANCE_TIMELOCK_ROLE, or admin can clear the record.
     */
    function resumeISINManual(bytes32 isin) external override nonReentrant {
        ISINHaltStatus storage status = isinHalt[isin];
        if (!status.isHalted) {
            revert ISINNotHalted(isin);
        }

        if (status.expiresAt > 0 && status.expiresAt != type(uint64).max && block.timestamp < status.expiresAt) {
            if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
                revert CooldownNotElapsed(status.expiresAt - uint64(block.timestamp));
            }
        } else if (status.expiresAt == 0 || status.expiresAt == type(uint64).max) {
            if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
                revert UnauthorizedSentinel(msg.sender, GOVERNANCE_TIMELOCK_ROLE);
            }
        } else {
            if (!hasRole(EMERGENCY_GUARDIAN_ROLE, msg.sender) &&
                !hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) &&
                !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
                revert UnauthorizedSentinel(msg.sender, EMERGENCY_GUARDIAN_ROLE);
            }
        }

        status.isHalted = false;
        emit ISINResumed(isin, uint64(block.timestamp), msg.sender);
    }

    /**
     * @notice Freezes an entire industry sector / asset class basket (Tier 1).
     */
    function haltSector(
        bytes32 sectorId,
        HaltReason reason,
        bytes32 detailsHash
    ) external override nonReentrant {
        if (!hasRole(EMERGENCY_GUARDIAN_ROLE, msg.sender) &&
            !hasRole(SURVEILLANCE_SENTINEL_ROLE, msg.sender) &&
            !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, EMERGENCY_GUARDIAN_ROLE);
        }

        if (sectorHalted[sectorId]) {
            revert SectorAlreadyHalted(sectorId);
        }

        sectorHalted[sectorId] = true;
        emit SectorHalted(sectorId, reason, uint64(block.timestamp), detailsHash, msg.sender);
    }

    /**
     * @notice Resumes an industry sector / asset class basket (Tier 1).
     */
    function resumeSector(bytes32 sectorId) external override nonReentrant {
        if (!hasRole(EMERGENCY_GUARDIAN_ROLE, msg.sender) &&
            !hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) &&
            !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, EMERGENCY_GUARDIAN_ROLE);
        }

        if (!sectorHalted[sectorId]) {
            revert SectorNotHalted(sectorId);
        }

        sectorHalted[sectorId] = false;
        emit SectorResumed(sectorId, uint64(block.timestamp), msg.sender);
    }

    /**
     * @notice Advances Market-Wide Circuit Breaker state from VOLATILITY_HALT to CALL_AUCTION_DISCOVERY.
     * Enforces that cooldown duration has elapsed unless called by GOVERNANCE_TIMELOCK_ROLE.
     */
    function transitionMarketWideToAuction() external override nonReentrant {
        MarketWideHaltState storage mw = marketWideState;
        if (mw.sessionState != MarketSessionState.VOLATILITY_HALT) {
            revert InvalidSessionTransition(mw.sessionState, MarketSessionState.CALL_AUCTION_DISCOVERY);
        }

        if (block.timestamp < mw.cooldownExpiresAt) {
            if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
                revert CooldownNotElapsed(mw.cooldownExpiresAt - uint64(block.timestamp));
            }
        }

        mw.sessionState = MarketSessionState.CALL_AUCTION_DISCOVERY;
        emit MarketWideAuctionInitiated(mw.circuitLevel, uint64(block.timestamp), mw.auctionExpiresAt);
    }

    /**
     * @notice Resumes Market-Wide trading back to NORMAL_TRADING.
     * Enforces that mandatory call-auction discovery interval has elapsed unless called by GOVERNANCE_TIMELOCK_ROLE.
     */
    function resumeMarketWide() external override nonReentrant {
        MarketWideHaltState storage mw = marketWideState;
        if (mw.sessionState == MarketSessionState.NORMAL_TRADING) {
            revert InvalidSessionTransition(mw.sessionState, MarketSessionState.NORMAL_TRADING);
        }

        if (mw.sessionState == MarketSessionState.VOLATILITY_HALT) {
            // Mandatory call-auction transition state required unless governance override
            if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
                revert InvalidSessionTransition(mw.sessionState, MarketSessionState.NORMAL_TRADING);
            }
        } else if (mw.sessionState == MarketSessionState.CALL_AUCTION_DISCOVERY) {
            if (block.timestamp < mw.auctionExpiresAt) {
                if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
                    revert CooldownNotElapsed(mw.auctionExpiresAt - uint64(block.timestamp));
                }
            }
        } else if (mw.sessionState == MarketSessionState.MARKET_CLOSED_FOR_DAY) {
            if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
                revert UnauthorizedSentinel(msg.sender, GOVERNANCE_TIMELOCK_ROLE);
            }
        }

        uint8 prevLevel = mw.circuitLevel;
        mw.sessionState = MarketSessionState.NORMAL_TRADING;
        mw.circuitLevel = 0;
        mw.cooldownExpiresAt = 0;
        mw.auctionExpiresAt = 0;

        emit MarketWideResumed(prevLevel, uint64(block.timestamp), msg.sender);
    }

    /**
     * @notice Maps an ISIN to an industry sector identifier.
     * @param isin ISIN identifier
     * @param sectorId Industry sector identifier
     */
    function mapISINToSector(bytes32 isin, bytes32 sectorId) external override {
        if (!hasRole(EMERGENCY_GUARDIAN_ROLE, msg.sender) &&
            !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, DEFAULT_ADMIN_ROLE);
        }

        isinToSector[isin] = sectorId;
        emit ISINSectorMapped(isin, sectorId);
    }

    /**
     * @notice Authorizes a sentinel account with a specific role.
     * @param sentinel Sentinel account address
     * @param role AccessControl role hash
     */
    function authorizeSentinel(address sentinel, bytes32 role) external override {
        if (!hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, DEFAULT_ADMIN_ROLE);
        }
        grantRole(role, sentinel);
        emit SentinelAuthorized(sentinel, role);
    }

    /**
     * @notice Updates the default oracle halt duration.
     * @param newDuration New duration in seconds
     */
    function updateDefaultOracleHaltDuration(uint32 newDuration) external {
        if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, DEFAULT_ADMIN_ROLE);
        }
        if (newDuration == 0 || newDuration > maxHaltDuration) {
            revert InvalidHaltDuration(newDuration);
        }
        uint32 oldDuration = defaultOracleHaltDuration;
        defaultOracleHaltDuration = newDuration;
        emit HaltParameterUpdated(keccak256("defaultOracleHaltDuration"), oldDuration, newDuration);
    }

    /**
     * @notice Updates the maximum permissible halt duration.
     * @param newMaxDuration New maximum duration in seconds
     */
    function updateMaxHaltDuration(uint32 newMaxDuration) external {
        if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, DEFAULT_ADMIN_ROLE);
        }
        if (newMaxDuration == 0) {
            revert InvalidHaltDuration(newMaxDuration);
        }
        uint32 oldMax = maxHaltDuration;
        maxHaltDuration = newMaxDuration;
        emit HaltParameterUpdated(keccak256("maxHaltDuration"), oldMax, newMaxDuration);
    }

    /**
     * @notice Updates the TimelockController address.
     * @param newTimelock New TimelockController address
     */
    function updateTimelockController(address newTimelock) external {
        if (!hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, DEFAULT_ADMIN_ROLE);
        }
        address oldTimelock = timelockController;
        if (oldTimelock != address(0)) {
            _revokeRole(GOVERNANCE_TIMELOCK_ROLE, oldTimelock);
        }
        timelockController = newTimelock;
        if (newTimelock != address(0)) {
            _grantRole(GOVERNANCE_TIMELOCK_ROLE, newTimelock);
        }
        emit HaltParameterUpdated(
            keccak256("timelockController"),
            uint256(uint160(oldTimelock)),
            uint256(uint160(newTimelock))
        );
    }

    // =========================================================================
    // Internal Helper Functions
    // =========================================================================

    /**
     * @dev Calculates SEBI MWCB cooldown duration and auction duration.
     * Complies with SEBI Circular SEBI/HO/MRD/DP/CIR/P/2018/142.
     */
    function _calculateMWCBDurations(uint8 circuitLevel, uint64 sessionTimestamp)
        internal
        pure
        returns (uint64 cooldownDuration, uint64 auctionDuration, MarketSessionState nextState)
    {
        uint64 ts = sessionTimestamp;
        // Convert unix timestamp to IST seconds-of-day if >= 86400.
        // IST is UTC + 5:30 = 19,800 seconds.
        // If < 86400, ts is treated directly as seconds of day.
        uint64 secondsOfDay = ts < 86400 ? ts : (ts + 19800) % 86400;

        // SEBI Timing Thresholds in IST seconds from midnight:
        // 1:00 PM IST = 13:00 = 46,800 seconds
        // 2:00 PM IST = 14:00 = 50,400 seconds
        // 2:30 PM IST = 14:30 = 52,200 seconds

        if (circuitLevel == 10) {
            if (secondsOfDay < 46800) {
                // Before 1:00 PM: 45 min halt, 15 min call-auction
                return (2700, 900, MarketSessionState.VOLATILITY_HALT);
            } else if (secondsOfDay < 52200) {
                // 1:00 PM up to 2:30 PM: 15 min halt, 15 min call-auction
                return (900, 900, MarketSessionState.VOLATILITY_HALT);
            } else {
                // At or after 2:30 PM: No halt; trading continues with price collars
                return (0, 0, MarketSessionState.NORMAL_TRADING);
            }
        } else if (circuitLevel == 15) {
            if (secondsOfDay < 46800) {
                // Before 1:00 PM: 1 hr 45 min (105 min) halt, 15 min call-auction
                return (6300, 900, MarketSessionState.VOLATILITY_HALT);
            } else if (secondsOfDay < 50400) {
                // 1:00 PM up to 2:00 PM: 45 min halt, 15 min call-auction
                return (2700, 900, MarketSessionState.VOLATILITY_HALT);
            } else {
                // At or after 2:00 PM: Remainder of the day
                return (type(uint64).max, type(uint64).max, MarketSessionState.MARKET_CLOSED_FOR_DAY);
            }
        } else if (circuitLevel == 20) {
            // At any time of the day: Remainder of the day
            return (type(uint64).max, type(uint64).max, MarketSessionState.MARKET_CLOSED_FOR_DAY);
        } else {
            revert InvalidCircuitLevel(circuitLevel);
        }
    }

    /**
     * @dev Restricts implementation upgrades to GOVERNANCE_TIMELOCK_ROLE or admin.
     */
    function _authorizeUpgrade(address newImplementation) internal override {
        newImplementation; // Silence unused variable warning
        if (!hasRole(GOVERNANCE_TIMELOCK_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedSentinel(msg.sender, GOVERNANCE_TIMELOCK_ROLE);
        }
    }
}
