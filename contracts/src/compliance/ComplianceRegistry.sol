// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {AccessControlUpgradeable} from "@openzeppelin/contracts-upgradeable/access/AccessControlUpgradeable.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {IComplianceRegistry, IComplianceModule} from "../interfaces/IComplianceRegistry.sol";
import {IIdentityRegistry} from "../interfaces/IIdentityRegistry.sol";

/**
 * @title ComplianceRegistry
 * @notice Modular compliance router coordinating onchain identity checks and pluggable compliance rules.
 * @dev Enforces ERC-3643 pre-transfer validation across bound compliance modules.
 */
contract ComplianceRegistry is
    Initializable,
    AccessControlUpgradeable,
    UUPSUpgradeable,
    IComplianceRegistry
{
    // --- Roles ---
    bytes32 public constant COMPLIANCE_ADMIN_ROLE = keccak256("COMPLIANCE_ADMIN_ROLE");
    bytes32 public constant UPGRADER_ROLE = keccak256("UPGRADER_ROLE");
    bytes32 public constant TOKEN_ROLE = keccak256("TOKEN_ROLE");

    // --- State Variables ---
    IIdentityRegistry public identityRegistry;
    address[] private _modules;
    mapping(address => bool) private _isBound;

    event IdentityRegistryUpdated(address indexed oldRegistry, address indexed newRegistry);

    error InvalidIdentityRegistry();

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initializes the compliance registry.
     * @param admin Address granted admin, compliance, and upgrader privileges.
     * @param identityRegistryAddress Address of the onchain identity registry.
     */
    function initialize(
        address admin,
        address identityRegistryAddress
    ) external initializer {
        if (admin == address(0)) revert InvalidModuleAddress();
        if (identityRegistryAddress == address(0)) revert InvalidIdentityRegistry();

        __AccessControl_init();
        __UUPSUpgradeable_init();

        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(COMPLIANCE_ADMIN_ROLE, admin);
        _grantRole(UPGRADER_ROLE, admin);

        identityRegistry = IIdentityRegistry(identityRegistryAddress);
    }

    /**
     * @notice Updates the associated identity registry.
     * @param newRegistry Address of the new identity registry.
     */
    function setIdentityRegistry(address newRegistry) external onlyRole(COMPLIANCE_ADMIN_ROLE) {
        if (newRegistry == address(0)) revert InvalidIdentityRegistry();
        address old = address(identityRegistry);
        identityRegistry = IIdentityRegistry(newRegistry);
        emit IdentityRegistryUpdated(old, newRegistry);
    }

    /**
     * @notice Binds a new compliance module to the registry.
     * @param moduleAddress Address of the compliance module contract.
     */
    function bindModule(address moduleAddress) external override onlyRole(COMPLIANCE_ADMIN_ROLE) {
        if (moduleAddress == address(0)) revert InvalidModuleAddress();
        if (_isBound[moduleAddress]) revert ModuleAlreadyBound(moduleAddress);

        _isBound[moduleAddress] = true;
        _modules.push(moduleAddress);

        emit ModuleBound(moduleAddress);
    }

    /**
     * @notice Unbinds an active compliance module from the registry.
     * @param moduleAddress Address of the module to unbind.
     */
    function unbindModule(address moduleAddress) external override onlyRole(COMPLIANCE_ADMIN_ROLE) {
        if (!_isBound[moduleAddress]) revert ModuleNotBound(moduleAddress);

        _isBound[moduleAddress] = false;

        uint256 length = _modules.length;
        for (uint256 i = 0; i < length; ++i) {
            if (_modules[i] == moduleAddress) {
                _modules[i] = _modules[length - 1];
                _modules.pop();
                break;
            }
        }

        emit ModuleUnbound(moduleAddress);
    }

    /**
     * @notice Returns all bound compliance modules.
     * @return Array of module addresses.
     */
    function getModules() external view override returns (address[] memory) {
        return _modules;
    }

    /**
     * @notice Pre-transfer compliance verification hook.
     * @param from Sender address.
     * @param to Recipient address.
     * @param amount Token transfer amount.
     * @return True if transfer satisfies identity KYC and all module rules.
     */
    function canTransfer(
        address from,
        address to,
        uint256 amount
    ) external view override returns (bool) {
        // 1. Identity checks
        if (from != address(0)) {
            if (!identityRegistry.isVerified(from)) {
                return false;
            }
        }

        if (to != address(0)) {
            if (!identityRegistry.isVerified(to)) {
                return false;
            }
        }

        // 2. Iterate through all bound compliance modules
        uint256 length = _modules.length;
        for (uint256 i = 0; i < length; ++i) {
            bool passed = IComplianceModule(_modules[i]).moduleCheck(from, to, amount, address(this));
            if (!passed) {
                return false;
            }
        }

        return true;
    }

    /**
     * @notice Post-transfer hook called by authorized security tokens.
     */
    function transferred(address from, address to, uint256 amount) external override {
        uint256 length = _modules.length;
        for (uint256 i = 0; i < length; ++i) {
            IComplianceModule(_modules[i]).moduleTransferAction(from, to, amount);
        }
    }

    /**
     * @notice Post-mint hook called by authorized security tokens.
     */
    function created(address to, uint256 amount) external override {
        uint256 length = _modules.length;
        for (uint256 i = 0; i < length; ++i) {
            IComplianceModule(_modules[i]).moduleMintAction(to, amount);
        }
    }

    /**
     * @notice Post-burn hook called by authorized security tokens.
     */
    function destroyed(address from, uint256 amount) external override {
        uint256 length = _modules.length;
        for (uint256 i = 0; i < length; ++i) {
            IComplianceModule(_modules[i]).moduleBurnAction(from, amount);
        }
    }

    function _authorizeUpgrade(
        address newImplementation
    ) internal override onlyRole(UPGRADER_ROLE) {}
}
