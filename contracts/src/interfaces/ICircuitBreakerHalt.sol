// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

/**
 * @title ICircuitBreakerHalt
 * @notice Multi-tier emergency pause, timelock, and circuit breaker interface for the Growww NBSE exchange.
 * Compliant with SEBI Master Circular on Market-Wide Circuit Breakers (MWCB) and IFSCA regulations.
 */
interface ICircuitBreakerHalt {
    /// @notice Halt tier levels ordered from least to most severe
    enum HaltTier {
        NONE,               // 0: Normal operations
        TIER_0_ISIN,        // 1: Individual security halted
        TIER_1_SECTOR,      // 2: Entire industry sector halted
        TIER_2_MARKET_WIDE, // 3: System-wide market index halt (SEBI 10%, 15%, 20%)
        TIER_3_GLOBAL       // 4: Total consortium ledger freeze
    }

    /// @notice Statutory and operational trigger reasons
    enum HaltReason {
        MANUAL_GOVERNANCE,        // Administrative or regulatory order
        VOLATILITY_COLLAR,        // LULD / dynamic price band breach
        ORACLE_DIVERGENCE,        // Median-of-N oracle feed mismatch
        DEPOSITORY_DESYNC,        // Physical demat custody vs token balance mismatch
        SURVEILLANCE_ANOMALY,     // Spoofing, layering, or market manipulation alert
        INFRASTRUCTURE_FAULT,     // Bridge exploit, validator failure, or node desync
        SEBI_INDEX_CIRCUIT_10,    // Market-wide index 10% movement
        SEBI_INDEX_CIRCUIT_15,    // Market-wide index 15% movement
        SEBI_INDEX_CIRCUIT_20     // Market-wide index 20% movement
    }

    /// @notice Market trading session state for SEBI MWCB tracking
    enum MarketSessionState {
        NORMAL_TRADING,
        VOLATILITY_HALT,
        CALL_AUCTION_DISCOVERY,
        MARKET_CLOSED_FOR_DAY
    }

    /// @notice State container for an individual ISIN halt
    struct ISINHaltStatus {
        bool isHalted;
        HaltReason reason;
        uint64 haltedAt;
        uint64 expiresAt;
        bytes32 detailsHash;       // IPFS/cryptographic hash of incident dossier
        address haltedBy;
    }

    /// @notice State container for Market-Wide Circuit Breaker (Tier 2)
    struct MarketWideHaltState {
        MarketSessionState sessionState;
        uint8 circuitLevel;        // 10, 15, or 20
        uint64 haltedAt;
        uint64 cooldownExpiresAt;
        uint64 auctionExpiresAt;
        bytes32 triggerIndex;      // e.g., keccak256("NIFTY_50_COMPOSITE")
        uint256 indexLevelPaise;   // Index level at trigger (2 decimal precision)
    }

    // --- Events ---
    event ISINHalted(
        bytes32 indexed isin,
        HaltReason indexed reason,
        uint64 indexed haltedAt,
        uint64 expiresAt,
        bytes32 detailsHash,
        address haltedBy
    );

    event ISINResumed(
        bytes32 indexed isin,
        uint64 indexed resumedAt,
        address resumedBy
    );

    event SectorHalted(
        bytes32 indexed sectorId,
        HaltReason indexed reason,
        uint64 indexed haltedAt,
        bytes32 detailsHash,
        address haltedBy
    );

    event SectorResumed(
        bytes32 indexed sectorId,
        uint64 indexed resumedAt,
        address resumedBy
    );

    event MarketWideHaltTriggered(
        uint8 indexed circuitLevel,
        uint64 indexed haltedAt,
        uint64 cooldownExpiresAt,
        uint64 auctionExpiresAt,
        uint256 indexLevelPaise,
        bytes32 triggerIndex,
        address triggeredBy
    );

    event MarketWideAuctionInitiated(
        uint8 indexed circuitLevel,
        uint64 indexed auctionStartedAt,
        uint64 auctionExpiresAt
    );

    event MarketWideResumed(
        uint8 indexed circuitLevel,
        uint64 indexed resumedAt,
        address resumedBy
    );

    event GlobalEmergencyFreezeTriggered(
        uint64 indexed frozenAt,
        bytes32 indexed reasonHash,
        address triggeredBy
    );

    event GlobalEmergencyFreezeLifted(
        uint64 indexed liftedAt,
        address liftedBy
    );

    event ISINSectorMapped(
        bytes32 indexed isin,
        bytes32 indexed sectorId
    );

    event SentinelAuthorized(
        address indexed sentinel,
        bytes32 indexed role
    );

    event HaltParameterUpdated(
        bytes32 indexed parameterName,
        uint256 oldValue,
        uint256 newValue
    );

    // --- Custom Errors ---
    error CircuitBreakerActive(bytes32 isin, HaltTier tier, uint64 expiresAt);
    error MarketWideHaltActive(uint8 level, uint64 cooldownExpiry);
    error GlobalFreezeActive();
    error GlobalFreezeNotActive();
    error CooldownNotElapsed(uint64 remainingSeconds);
    error InvalidHaltDuration(uint32 duration);
    error UnauthorizedSentinel(address caller, bytes32 requiredRole);
    error ISINAlreadyHalted(bytes32 isin);
    error ISINNotHalted(bytes32 isin);
    error SectorAlreadyHalted(bytes32 sectorId);
    error SectorNotHalted(bytes32 sectorId);
    error InvalidSessionTransition(MarketSessionState current, MarketSessionState next);
    error InvalidCircuitLevel(uint8 level);

    // --- Core Verification Views (Called by Settlement & Token Contracts) ---
    function isHalted(bytes32 isin) external view returns (bool halted, HaltTier tier, uint64 expiresAt);
    function requireNotHalted(bytes32 isin) external view;
    function isGlobalFrozen() external view returns (bool);
    function getMarketWideState() external view returns (MarketWideHaltState memory);
    function getISINHaltStatus(bytes32 isin) external view returns (ISINHaltStatus memory);

    // --- Sector and Mapping Views ---
    function isSectorHalted(bytes32 sectorId) external view returns (bool);
    function getSectorForISIN(bytes32 isin) external view returns (bytes32);

    // --- Autonomous Sentinel Triggers ---
    function triggerDepositoryHalt(bytes32 isin, bytes32 mismatchProofHash) external;
    function triggerOracleDivergenceHalt(bytes32 isin, uint256 deviationBps, bytes32 detailsHash) external;
    function triggerSurveillanceVolatilityHalt(bytes32 isin, uint32 durationSeconds, bytes32 detailsHash) external;
    function triggerMarketWideCircuitBreaker(
        uint8 circuitLevel,
        bytes32 indexId,
        uint256 indexLevelPaise,
        uint64 sessionTimestamp
    ) external;

    // --- Governance & Administrative Functions ---
    function triggerGlobalEmergencyFreeze(bytes32 reasonHash) external;
    function liftGlobalEmergencyFreeze() external;
    function haltISINManual(bytes32 isin, HaltReason reason, uint32 durationSeconds, bytes32 detailsHash) external;
    function resumeISINManual(bytes32 isin) external;
    function haltSector(bytes32 sectorId, HaltReason reason, bytes32 detailsHash) external;
    function resumeSector(bytes32 sectorId) external;
    function transitionMarketWideToAuction() external;
    function resumeMarketWide() external;
    function mapISINToSector(bytes32 isin, bytes32 sectorId) external;
    function authorizeSentinel(address sentinel, bytes32 role) external;
}
