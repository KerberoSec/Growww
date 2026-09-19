// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title IAccountRules
 * @notice Interface for Besu on-chain account permissioning rules.
 * @dev Besu calls `transactionAllowed` before including a transaction in the transaction pool.
 */
interface IAccountRules {
    event AccountAdded(bool indexed active, address indexed account);
    event AccountRemoved(bool indexed active, address indexed account);

    function transactionAllowed(
        address sender,
        address target,
        uint256 value,
        uint256 gasPrice,
        uint256 gasLimit,
        bytes calldata payload
    ) external view returns (bool isAllowed);

    function addAccount(address account) external;
    function removeAccount(address account) external;
    function isAccountAllowed(address account) external view returns (bool);
    function size() external view returns (uint256);
}
