// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {IAccountRules} from "../interfaces/permissioning/IAccountRules.sol";
import {AdminRegistry} from "./AdminRegistry.sol";

/**
 * @title AccountRules
 * @notice Besu on-chain account permissioning rules contract.
 * @dev Governs transaction submission authorization on the consortium ledger.
 *      Only allowlisted addresses (relayers, operators, authorized institutional accounts)
 *      can submit transactions to the transaction pool.
 */
contract AccountRules is Initializable, IAccountRules {
    AdminRegistry public adminRegistry;

    mapping(address => bool) private _allowedAccounts;
    address[] private _accountsList;

    error UnauthorizedCaller(address caller);
    error InvalidZeroAddress();
    error AccountAlreadyAllowed(address account);
    error AccountNotAllowed(address account);

    modifier onlyAdmin() {
        if (!adminRegistry.isAuthorizedAdmin(msg.sender)) {
            revert UnauthorizedCaller(msg.sender);
        }
        _;
    }

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    function initialize(address adminRegistryAddress) external initializer {
        if (adminRegistryAddress == address(0)) revert InvalidZeroAddress();
        adminRegistry = AdminRegistry(adminRegistryAddress);
    }

    function transactionAllowed(
        address sender,
        address /* target */,
        uint256 /* value */,
        uint256 /* gasPrice */,
        uint256 /* gasLimit */,
        bytes calldata /* payload */
    ) external view override returns (bool isAllowed) {
        return _allowedAccounts[sender];
    }

    function isAccountAllowed(address account) external view override returns (bool) {
        return _allowedAccounts[account];
    }

    function addAccount(address account) external override onlyAdmin {
        if (account == address(0)) revert InvalidZeroAddress();
        if (_allowedAccounts[account]) revert AccountAlreadyAllowed(account);

        _allowedAccounts[account] = true;
        _accountsList.push(account);

        emit AccountAdded(true, account);
    }

    function removeAccount(address account) external override onlyAdmin {
        if (!_allowedAccounts[account]) revert AccountNotAllowed(account);

        _allowedAccounts[account] = false;
        emit AccountRemoved(false, account);
    }

    function size() external view override returns (uint256) {
        uint256 count = 0;
        for (uint256 i = 0; i < _accountsList.length; i++) {
            if (_allowedAccounts[_accountsList[i]]) {
                count++;
            }
        }
        return count;
    }
}
