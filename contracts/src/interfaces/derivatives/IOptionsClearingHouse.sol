// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "./IOptionsTokenFactory.sol";

interface IOptionsClearingHouse {
    struct CollateralPosition {
        uint256 positionId;
        bytes32 seriesId;
        address writer;
        address collateralToken;
        uint256 lockedAmount;
        uint256 mintedContracts;
        uint256 unexercisedContracts;
        bool isReleased;
    }

    struct SettlementPriceData {
        bytes32 seriesId;
        uint256 settlementPrice;
        uint256 priceTimestamp;
        uint256 sequenceNumber;
    }

    struct BatchExerciseParams {
        bytes32 seriesId;
        address[] holders;
        uint256[] contractAmounts;
        uint256[] costBases;
    }

    // Events
    event CollateralDepositedAndOptionMinted(
        uint256 indexed positionId,
        bytes32 indexed seriesId,
        address indexed writer,
        address collateralToken,
        uint256 collateralAmount,
        uint256 contractsMinted
    );
    event SettlementPriceSubmitted(
        bytes32 indexed seriesId,
        uint256 settlementPrice,
        uint256 priceTimestamp,
        address indexed oracleRelayer
    );
    event AutomatedExerciseExecuted(
        bytes32 indexed seriesId,
        address indexed holder,
        uint256 contractAmount,
        uint256 netPayout,
        uint256 platformFeeAmount
    );
    event CollateralUnlocked(uint256 indexed positionId, bytes32 indexed seriesId, address indexed writer, uint256 amountReleased);
    event OracleRelayerUpdated(address indexed oracleRelayer, bool isWhitelisted);

    // Custom Errors
    error InsufficientCollateralProvided(uint256 required, uint256 provided);
    error ExpiryNotReached(uint256 currentTimestamp, uint256 expiryTimestamp);
    error ExpiryAlreadyProcessed(bytes32 seriesId);
    error StaleOraclePrice(uint256 priceTimestamp, uint256 currentTimestamp);
    error InvalidOracleSignature();
    error OracleNotWhitelisted(address relayer);
    error PositionAlreadyReleased(uint256 positionId);
    error InactiveOptionSeries(bytes32 seriesId);
    error ArrayLengthMismatch();
    error OptionNotITM(bytes32 seriesId, uint256 settlementPrice, uint256 strikePrice);
    error UnauthorizedCaller(address caller);

    // Clearing Operations
    function writeAndMintOption(
        bytes32 seriesId,
        uint256 contractAmount
    ) external returns (uint256 positionId);

    function submitSettlementPrice(
        bytes32 seriesId,
        SettlementPriceData calldata data,
        bytes calldata signature
    ) external;

    function batchAutoExerciseAndSettle(
        BatchExerciseParams calldata params
    ) external;

    function reclaimOTMCollateral(
        uint256 positionId
    ) external;

    function reclaimResidualCollateral(
        uint256 positionId
    ) external;

    // View Functions
    function getCollateralPosition(uint256 positionId) external view returns (CollateralPosition memory);
    function getSettlementPrice(bytes32 seriesId) external view returns (uint256 price, uint256 timestamp, bool isLocked);
    function isOptionITM(bytes32 seriesId) external view returns (bool isITM, uint256 intrinsicPayoffPerUnit);
    function totalLockedCollateral(address token) external view returns (uint256);
}
