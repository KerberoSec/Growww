// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {IComplianceModule} from "../../interfaces/IComplianceRegistry.sol";
import {IIdentityRegistry} from "../../interfaces/IIdentityRegistry.sol";

/**
 * @title CountryRestrictModule
 * @notice Enforces jurisdictional investment restrictions based on ISO 3166-1 numeric country codes.
 * @dev Rejects transfers if either sender or receiver resides in a restricted or unapproved jurisdiction.
 */
contract CountryRestrictModule is Ownable, IComplianceModule {
    IIdentityRegistry public identityRegistry;

    mapping(uint16 => bool) public allowedCountries;

    event CountryStatusUpdated(uint16 indexed countryCode, bool allowed);
    event IdentityRegistryUpdated(address indexed oldRegistry, address indexed newRegistry);

    error InvalidIdentityRegistry();

    constructor(address initialOwner, address identityRegistryAddress) Ownable(initialOwner) {
        if (identityRegistryAddress == address(0)) revert InvalidIdentityRegistry();
        identityRegistry = IIdentityRegistry(identityRegistryAddress);
    }

    function setCountryStatus(uint16 countryCode, bool allowed) external onlyOwner {
        allowedCountries[countryCode] = allowed;
        emit CountryStatusUpdated(countryCode, allowed);
    }

    function batchSetCountryStatus(uint16[] calldata countryCodes, bool[] calldata statuses) external onlyOwner {
        require(countryCodes.length == statuses.length, "LengthMismatch");
        for (uint256 i = 0; i < countryCodes.length; ++i) {
            allowedCountries[countryCodes[i]] = statuses[i];
            emit CountryStatusUpdated(countryCodes[i], statuses[i]);
        }
    }

    function setIdentityRegistry(address newRegistry) external onlyOwner {
        if (newRegistry == address(0)) revert InvalidIdentityRegistry();
        address old = address(identityRegistry);
        identityRegistry = IIdentityRegistry(newRegistry);
        emit IdentityRegistryUpdated(old, newRegistry);
    }

    function moduleCheck(
        address from,
        address to,
        uint256 /* amount */,
        address /* complianceRegistry */
    ) external view override returns (bool) {
        // If mint (from == 0), check receiver only
        if (from != address(0)) {
            uint16 fromCountry = identityRegistry.getInvestorCountry(from);
            if (!allowedCountries[fromCountry]) {
                return false;
            }
        }

        // If burn (to == 0), check sender only
        if (to != address(0)) {
            uint16 toCountry = identityRegistry.getInvestorCountry(to);
            if (!allowedCountries[toCountry]) {
                return false;
            }
        }

        return true;
    }

    function moduleTransferAction(address /* from */, address /* to */, uint256 /* amount */) external override {}
    function moduleMintAction(address /* to */, uint256 /* amount */) external override {}
    function moduleBurnAction(address /* from */, uint256 /* amount */) external override {}
}
