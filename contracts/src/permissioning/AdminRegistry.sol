// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

/**
 * @title AdminRegistry
 * @notice Central access governance contract for Hyperledger Besu consortium permissioning rules.
 * @dev Coordinates admin rights across NodeRules and AccountRules permissioning contracts.
 */
contract AdminRegistry is Initializable, AccessControlUpgradeable {
    bytes32 public constant PERMISSIONING_ADMIN_ROLE = keccak256("PERMISSIONING_ADMIN_ROLE");
    bytes32 public constant CONSORTIUM_OPERATOR_ROLE = keccak256("CONSORTIUM_OPERATOR_ROLE");

    error InvalidZeroAddress();
    error NotAdmin(address caller);

    event AdminAdded(address indexed admin);
    event AdminRemoved(address indexed admin);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address initialAdmin) external initializer {
        if (initialAdmin == address(0)) revert InvalidZeroAddress();

        _grantRole(DEFAULT_ADMIN_ROLE, initialAdmin);
        _grantRole(PERMISSIONING_ADMIN_ROLE, initialAdmin);
        _grantRole(CONSORTIUM_OPERATOR_ROLE, initialAdmin);
    }

    function isAuthorizedAdmin(address account) external view returns (bool) {
        return hasRole(DEFAULT_ADMIN_ROLE, account) || hasRole(PERMISSIONING_ADMIN_ROLE, account);
    }

    function addAdmin(address newAdmin) external onlyRole(DEFAULT_ADMIN_ROLE) {
        if (newAdmin == address(0)) revert InvalidZeroAddress();
        _grantRole(PERMISSIONING_ADMIN_ROLE, newAdmin);
        emit AdminAdded(newAdmin);
    }

    function removeAdmin(address admin) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _revokeRole(PERMISSIONING_ADMIN_ROLE, admin);
        emit AdminRemoved(admin);
    }
}
