// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import "@openzeppelin/contracts-upgradeable/token/ERC1155/ERC1155Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/PausableUpgradeable.sol";
import "../interfaces/derivatives/IOptionsTokenFactory.sol";
import "../interfaces/IIdentityRegistry.sol";

/**
 * @title OptionsTokenFactory
 * @notice ERC-1155 multi-token factory for standardized Call and Put option series.
 * @dev Manages option series parameters, deterministic token IDs, minting, burning, and KYC compliance.
 */
contract OptionsTokenFactory is
    IOptionsTokenFactory,
    ERC1155Upgradeable,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    PausableUpgradeable
{
    bytes32 public constant CLEARING_HOUSE_ROLE = keccak256("CLEARING_HOUSE_ROLE");
    bytes32 public constant EMERGENCY_GUARDIAN_ROLE = keccak256("EMERGENCY_GUARDIAN_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    mapping(bytes32 => OptionSeriesParams) private _seriesRegistry;
    mapping(uint256 => bytes32) private _tokenToSeries;

    address public identityRegistry;
    address public clearingHouse;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(
        address admin,
        address _clearingHouse,
        address _identityRegistry,
        string memory uri_
    ) external initializer {
        __ERC1155_init(uri_);
        __AccessControl_init();
        __UUPSUpgradeable_init();
        __Pausable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        _grantRole(EMERGENCY_GUARDIAN_ROLE, admin);

        if (_clearingHouse != address(0)) {
            clearingHouse = _clearingHouse;
            _grantRole(CLEARING_HOUSE_ROLE, _clearingHouse);
        }
        identityRegistry = _identityRegistry;
    }

    function setClearingHouse(address _clearingHouse) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (clearingHouse != address(0)) {
            _revokeRole(CLEARING_HOUSE_ROLE, clearingHouse);
        }
        clearingHouse = _clearingHouse;
        if (_clearingHouse != address(0)) {
            _grantRole(CLEARING_HOUSE_ROLE, _clearingHouse);
        }
    }

    function setIdentityRegistry(address _identityRegistry) external onlyRole(DEFAULT_ADMIN_ROLE) {
        identityRegistry = _identityRegistry;
    }

    function setURI(string memory newUri) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _setURI(newUri);
    }

    function pause() external onlyRole(EMERGENCY_GUARDIAN_ROLE) {
        _pause();
    }

    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) {
        _unpause();
    }

    function computeTokenId(
        address underlyingToken,
        OptionType optionType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        SettlementType settlementType
    ) public pure returns (uint256) {
        return uint256(keccak256(abi.encode(underlyingToken, optionType, strikePrice, expiryTimestamp, settlementType)));
    }

    function computeSeriesId(
        address underlyingToken,
        OptionType optionType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        SettlementType settlementType
    ) public pure returns (bytes32) {
        return keccak256(abi.encode(underlyingToken, optionType, strikePrice, expiryTimestamp, settlementType));
    }

    function createOptionSeries(
        address underlyingToken,
        address settlementToken,
        OptionType optionType,
        SettlementType settlementType,
        uint256 strikePrice,
        uint256 expiryTimestamp,
        bool isAmerican
    ) external returns (bytes32 seriesId, uint256 tokenId) {
        if (underlyingToken == address(0)) revert InvalidUnderlyingToken(underlyingToken);
        if (settlementToken == address(0)) revert InvalidSettlementToken(settlementToken);
        if (strikePrice == 0) revert InvalidStrikePrice();
        if (expiryTimestamp <= block.timestamp) revert InvalidExpiryTimestamp(expiryTimestamp, block.timestamp);

        seriesId = computeSeriesId(underlyingToken, optionType, strikePrice, expiryTimestamp, settlementType);
        tokenId = computeTokenId(underlyingToken, optionType, strikePrice, expiryTimestamp, settlementType);

        if (_seriesRegistry[seriesId].state != SeriesState.UNINITIALIZED) {
            revert SeriesAlreadyExists(seriesId);
        }

        OptionSeriesParams memory params = OptionSeriesParams({
            seriesId: seriesId,
            underlyingToken: underlyingToken,
            settlementToken: settlementToken,
            optionType: optionType,
            settlementType: settlementType,
            strikePrice: strikePrice,
            expiryTimestamp: expiryTimestamp,
            isAmerican: isAmerican,
            state: SeriesState.ACTIVE
        });

        _seriesRegistry[seriesId] = params;
        _tokenToSeries[tokenId] = seriesId;

        emit OptionSeriesCreated(
            seriesId,
            tokenId,
            underlyingToken,
            settlementToken,
            optionType,
            settlementType,
            strikePrice,
            expiryTimestamp,
            isAmerican
        );
    }

    function mintOption(address recipient, uint256 tokenId, uint256 amount) external onlyRole(CLEARING_HOUSE_ROLE) {
        bytes32 seriesId = _tokenToSeries[tokenId];
        if (seriesId == bytes32(0)) revert SeriesDoesNotExist(seriesId);
        if (_seriesRegistry[seriesId].state != SeriesState.ACTIVE) {
            revert SeriesNotActive(seriesId, _seriesRegistry[seriesId].state);
        }
        if (identityRegistry != address(0)) {
            if (!IIdentityRegistry(identityRegistry).isVerified(recipient)) {
                revert NonKYCRecipient(recipient);
            }
        }

        _mint(recipient, tokenId, amount, "");
        emit OptionsMinted(seriesId, tokenId, recipient, amount);
    }

    function burnOption(address holder, uint256 tokenId, uint256 amount) external onlyRole(CLEARING_HOUSE_ROLE) {
        bytes32 seriesId = _tokenToSeries[tokenId];
        if (seriesId == bytes32(0)) revert SeriesDoesNotExist(seriesId);

        _burn(holder, tokenId, amount);
        emit OptionsBurned(seriesId, tokenId, holder, amount);
    }

    function updateSeriesState(bytes32 seriesId, SeriesState newState) external {
        if (!hasRole(CLEARING_HOUSE_ROLE, msg.sender) && !hasRole(DEFAULT_ADMIN_ROLE, msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }
        if (_seriesRegistry[seriesId].state == SeriesState.UNINITIALIZED) {
            revert SeriesDoesNotExist(seriesId);
        }

        SeriesState oldState = _seriesRegistry[seriesId].state;
        _seriesRegistry[seriesId].state = newState;
        emit SeriesStateUpdated(seriesId, oldState, newState);
    }

    function getSeriesParams(bytes32 seriesId) external view returns (OptionSeriesParams memory) {
        OptionSeriesParams memory params = _seriesRegistry[seriesId];
        if (params.state == SeriesState.UNINITIALIZED) revert SeriesDoesNotExist(seriesId);
        return params;
    }

    function getSeriesParamsByTokenId(uint256 tokenId) external view returns (OptionSeriesParams memory) {
        bytes32 seriesId = _tokenToSeries[tokenId];
        if (seriesId == bytes32(0)) revert SeriesDoesNotExist(seriesId);
        return _seriesRegistry[seriesId];
    }

    function isSeriesActive(bytes32 seriesId) external view returns (bool) {
        return _seriesRegistry[seriesId].state == SeriesState.ACTIVE;
    }

    function _update(
        address from,
        address to,
        uint256[] memory ids,
        uint256[] memory values
    ) internal override whenNotPaused {
        super._update(from, to, ids, values);

        if (from != address(0) && to != address(0) && identityRegistry != address(0)) {
            if (!IIdentityRegistry(identityRegistry).isVerified(to)) {
                revert NonKYCRecipient(to);
            }
        }
    }

    function supportsInterface(
        bytes4 interfaceId
    ) public view override(ERC1155Upgradeable, AccessControlUpgradeable) returns (bool) {
        return super.supportsInterface(interfaceId);
    }

    function _authorizeUpgrade(address newImplementation) internal override onlyRole(UPGRADER_ROLE) {}

    uint256[45] private __gap;
}
