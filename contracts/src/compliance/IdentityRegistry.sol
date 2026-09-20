// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {IIdentityRegistry} from "../interfaces/IIdentityRegistry.sol";

/**
 * @title IdentityRegistry
 * @notice On-chain identity claims registry mapping pseudonymous Ethereum addresses to verified investor claims.
 * @dev Stores cryptographic identity hashes without storing investor PII on-chain.
 *      Enforces KYC tier verification and sanctions checks required by SEBI and IFSCA.
 */
contract IdentityRegistry is
    Initializable,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    IIdentityRegistry
{
    // --- Roles ---
    bytes32 public constant AGENT_ROLE = keccak256("AGENT_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");

    // --- State Variables ---
    mapping(address => bytes32) private _identities;
    mapping(bytes32 => InvestorClaim) private _claims;
    mapping(bytes32 => bool) private _sanctionedIdentities;
    uint256 private _totalIdentities;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initializes the Identity Registry contract.
     * @param admin Address granted the default admin and upgrader roles.
     * @param agent Address granted the identity management agent role.
     */
    function initialize(address admin, address agent) external initializer {
        if (admin == address(0) || agent == address(0)) revert InvalidZeroAddress();

        __AccessControl_init();
        __UUPSUpgradeable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);
        _grantRole(AGENT_ROLE, agent);
    }

    /**
     * @notice Registers an investor identity claim for a given wallet address.
     * @param userAddress The Ethereum address of the investor.
     * @param identityId Cryptographic hash representing the investor's off-chain identity.
     * @param countryCode ISO 3166-1 numeric country code of the investor.
     * @param kycTier KYC tier (1: Basic, 2: Domestic Verified, 3: Accredited Foreign).
     */
    function registerIdentity(
        address userAddress,
        bytes32 identityId,
        uint16 countryCode,
        uint8 kycTier
    ) public override onlyRole(AGENT_ROLE) {
        if (userAddress == address(0)) revert InvalidZeroAddress();
        if (identityId == bytes32(0)) revert IdentityNotFound(userAddress);
        if (_identities[userAddress] != bytes32(0)) revert IdentityAlreadyExists(userAddress);

        _identities[userAddress] = identityId;
        _claims[identityId] = InvestorClaim({
            identityId: identityId,
            countryCode: countryCode,
            kycTier: kycTier,
            isSanctioned: _sanctionedIdentities[identityId],
            registeredAt: uint64(block.timestamp)
        });

        _totalIdentities += 1;

        emit IdentityRegistered(userAddress, identityId, countryCode);
    }

    /**
     * @notice Batch registers investor identity claims.
     * @param userAddresses Array of investor Ethereum addresses.
     * @param identityIds Array of identity hashes.
     * @param countryCodes Array of ISO 3166-1 numeric country codes.
     * @param kycTiers Array of KYC tiers.
     */
    function batchRegisterIdentity(
        address[] calldata userAddresses,
        bytes32[] calldata identityIds,
        uint16[] calldata countryCodes,
        uint8[] calldata kycTiers
    ) external override onlyRole(AGENT_ROLE) {
        uint256 length = userAddresses.length;
        if (
            length != identityIds.length ||
            length != countryCodes.length ||
            length != kycTiers.length
        ) {
            revert ArrayLengthMismatch();
        }

        for (uint256 i = 0; i < length; ++i) {
            registerIdentity(userAddresses[i], identityIds[i], countryCodes[i], kycTiers[i]);
        }
    }

    /**
     * @notice Deletes an investor identity record.
     * @param userAddress The investor Ethereum address to delete.
     */
    function deleteIdentity(address userAddress) external override onlyRole(AGENT_ROLE) {
        bytes32 identityId = _identities[userAddress];
        if (identityId == bytes32(0)) revert IdentityNotFound(userAddress);

        delete _identities[userAddress];
        delete _claims[identityId];

        if (_totalIdentities > 0) {
            _totalIdentities -= 1;
        }

        emit IdentityRemoved(userAddress, identityId);
    }

    /**
     * @notice Updates the sanction status of an identity.
     * @param identityId Cryptographic hash representing the investor's identity.
     * @param isSanctioned_ True if the identity is placed on a sanctions list.
     */
    function updateSanctionStatus(
        bytes32 identityId,
        bool isSanctioned_
    ) external override onlyRole(AGENT_ROLE) {
        _sanctionedIdentities[identityId] = isSanctioned_;
        if (_claims[identityId].identityId != bytes32(0)) {
            _claims[identityId].isSanctioned = isSanctioned_;
        }

        emit SanctionStatusUpdated(identityId, isSanctioned_);
    }

    /**
     * @notice Updates the country classification of an investor.
     * @param userAddress The investor Ethereum address.
     * @param newCountryCode New ISO 3166-1 numeric country code.
     */
    function updateCountry(
        address userAddress,
        uint16 newCountryCode
    ) external override onlyRole(AGENT_ROLE) {
        bytes32 identityId = _identities[userAddress];
        if (identityId == bytes32(0)) revert IdentityNotFound(userAddress);

        _claims[identityId].countryCode = newCountryCode;

        emit CountryUpdated(userAddress, newCountryCode);
    }

    /**
     * @notice Updates the KYC tier of an investor.
     * @param userAddress The investor Ethereum address.
     * @param newKycTier New KYC tier.
     */
    function updateKycTier(
        address userAddress,
        uint8 newKycTier
    ) external override onlyRole(AGENT_ROLE) {
        bytes32 identityId = _identities[userAddress];
        if (identityId == bytes32(0)) revert IdentityNotFound(userAddress);

        _claims[identityId].kycTier = newKycTier;

        emit KycTierUpdated(userAddress, newKycTier);
    }

    /**
     * @notice Checks whether an address has a verified, non-sanctioned identity claim.
     * @param userAddress The investor Ethereum address.
     * @return True if the address has an active KYC claim and is not sanctioned.
     */
    function isVerified(address userAddress) external view override returns (bool) {
        bytes32 identityId = _identities[userAddress];
        if (identityId == bytes32(0)) {
            return false;
        }

        InvestorClaim memory claim = _claims[identityId];
        if (claim.isSanctioned || _sanctionedIdentities[identityId]) {
            return false;
        }

        return claim.kycTier > 0;
    }

    /**
     * @notice Retrieves full investor claim details for a wallet address.
     * @param userAddress The investor Ethereum address.
     * @return claim The InvestorClaim struct.
     */
    function getInvestorClaim(
        address userAddress
    ) external view override returns (InvestorClaim memory claim) {
        bytes32 identityId = _identities[userAddress];
        if (identityId == bytes32(0)) revert IdentityNotFound(userAddress);
        claim = _claims[identityId];
    }

    /**
     * @notice Retrieves the ISO 3166-1 numeric country code of an investor.
     * @param userAddress The investor Ethereum address.
     * @return Country code.
     */
    function getInvestorCountry(address userAddress) external view override returns (uint16) {
        bytes32 identityId = _identities[userAddress];
        if (identityId == bytes32(0)) revert IdentityNotFound(userAddress);
        return _claims[identityId].countryCode;
    }

    /**
     * @notice Retrieves the identityId for a given address.
     * @param userAddress The investor Ethereum address.
     * @return identityId bytes32 hash.
     */
    function getIdentityId(address userAddress) external view override returns (bytes32) {
        return _identities[userAddress];
    }

    /**
     * @notice Checks if an address is registered in the registry.
     * @param userAddress The investor Ethereum address.
     * @return True if registered.
     */
    function contains(address userAddress) external view override returns (bool) {
        return _identities[userAddress] != bytes32(0);
    }

    /**
     * @notice Checks if an identityId is sanctioned.
     * @param identityId Cryptographic hash.
     * @return True if sanctioned.
     */
    function isSanctioned(bytes32 identityId) external view override returns (bool) {
        return _sanctionedIdentities[identityId] || _claims[identityId].isSanctioned;
    }

    /**
     * @notice Total registered identities count.
     * @return Count of registered identities.
     */
    function totalIdentities() external view override returns (uint256) {
        return _totalIdentities;
    }

    function _authorizeUpgrade(
        address newImplementation
    ) internal override onlyRole(UPGRADER_ROLE) {}
}
