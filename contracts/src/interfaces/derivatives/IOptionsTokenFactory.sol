// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

interface IOptionsTokenFactory {
    enum OptionType {
        CALL,
        PUT
    }

    enum SettlementType {
        PHYSICAL_DELIVERY,
        CASH_SETTLED
    }

    enum SeriesState {
        UNINITIALIZED,
        ACTIVE,
        EXPIRED_PRICE_LOCKED,
        EXPIRED_WORTHLESS_OTM,
        FULLY_SETTLED,
        CANCELLED
    }

    struct OptionSeriesParams {
        bytes32 seriesId;
        address underlyingToken;
        address settlementToken;
        OptionType optionType;
        SettlementType settlementType;
        uint256 strikePrice;
        uint256 expiryTimestamp;
        bool isAmerican;
        SeriesState state;
    }

    // Events
    event OptionSeriesCreated(
        bytes32 indexed seriesId,
        uint256 indexed tokenId,
        address indexed underlyingToken,
        address settlementToken,
        OptionType optionType,
        SettlementType settlementType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        bool isAmerican
    );
    event OptionsMinted(bytes32 indexed seriesId, uint256 indexed tokenId, address indexed recipient, uint256 amount);
    event OptionsBurned(bytes32 indexed seriesId, uint256 indexed tokenId, address indexed holder, uint256 amount);
    event SeriesStateUpdated(bytes32 indexed seriesId, SeriesState indexed oldState, SeriesState indexed newState);

    // Custom Errors
    error SeriesAlreadyExists(bytes32 seriesId);
    error SeriesDoesNotExist(bytes32 seriesId);
    error InvalidUnderlyingToken(address token);
    error InvalidSettlementToken(address token);
    error InvalidStrikePrice();
    error InvalidExpiryTimestamp(uint256 provided, uint256 current);
    error UnauthorizedCaller(address caller);
    error NonKYCRecipient(address recipient);
    error SeriesNotActive(bytes32 seriesId, SeriesState currentState);

    // Factory Functions
    function createOptionSeries(
        address underlyingToken,
        address settlementToken,
        OptionType optionType,
        SettlementType settlementType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        bool isAmerican
    ) external returns (bytes32 seriesId, uint256 tokenId);

    function mintOption(address recipient, uint256 tokenId, uint256 amount) external;
    function burnOption(address holder, uint256 tokenId, uint256 amount) external;
    function updateSeriesState(bytes32 seriesId, SeriesState newState) external;

    // View Functions
    function getSeriesParams(bytes32 seriesId) external view returns (OptionSeriesParams memory);
    function getSeriesParamsByTokenId(uint256 tokenId) external view returns (OptionSeriesParams memory);
    function computeTokenId(
        address underlyingToken,
        OptionType optionType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        SettlementType settlementType
    ) external pure returns (uint256);
    function computeSeriesId(
        address underlyingToken,
        OptionType optionType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        SettlementType settlementType
    ) external pure returns (bytes32);
    function isSeriesActive(bytes32 seriesId) external view returns (bool);
}
