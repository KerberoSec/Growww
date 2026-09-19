// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IIdentityRegistry
 * @notice ERC-3643 Onchain Identity Registry interface for Growww Indian equity security tokens.
 * @dev Maps pseudonymous onchain addresses to verified investor identity claims without storing PII.
 */
interface IIdentityRegistry {
    struct InvestorClaim {
        bytes32 identityId;
        uint16 countryCode; // ISO 3166-1 numeric
        uint8 kycTier;      // 1 = Basic, 2 = Verified Domestic, 3 = Accredited Foreign
        bool isSanctioned;
        uint64 registeredAt;
    }

    event IdentityRegistered(address indexed userAddress, bytes32 indexed identityId, uint16 countryCode);
    event IdentityRemoved(address indexed userAddress, bytes32 indexed identityId);
    event CountryUpdated(address indexed userAddress, uint16 newCountryCode);
    event KycTierUpdated(address indexed userAddress, uint8 newKycTier);
    event SanctionStatusUpdated(bytes32 indexed identityId, bool isSanctioned);

    error IdentityAlreadyExists(address userAddress);
    error IdentityNotFound(address userAddress);
    error UnauthorizedAgent();
    error InvalidZeroAddress();
    error ArrayLengthMismatch();

    function registerIdentity(
        address userAddress,
        bytes32 identityId,
        uint16 countryCode,
        uint8 kycTier
    ) external;

    function batchRegisterIdentity(
        address[] calldata userAddresses,
        bytes32[] calldata identityIds,
        uint16[] calldata countryCodes,
        uint8[] calldata kycTiers
    ) external;

    function deleteIdentity(address userAddress) external;
    function updateSanctionStatus(bytes32 identityId, bool isSanctioned) external;
    function updateCountry(address userAddress, uint16 newCountryCode) external;
    function updateKycTier(address userAddress, uint8 newKycTier) external;

    function isVerified(address userAddress) external view returns (bool);
    function getInvestorClaim(address userAddress) external view returns (InvestorClaim memory);
    function getInvestorCountry(address userAddress) external view returns (uint16);
    function getIdentityId(address userAddress) external view returns (bytes32);
    function contains(address userAddress) external view returns (bool);
    function isSanctioned(bytes32 identityId) external view returns (bool);
    function totalIdentities() external view returns (uint256);
}
