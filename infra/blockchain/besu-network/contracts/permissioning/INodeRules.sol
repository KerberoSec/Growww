// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

/**
 * @title INodeRules
 * @notice Interface for Besu on-chain node permissioning rules.
 * @dev Besu calls `enodeAllowed` during the P2P handshake to determine whether a connection
 *      is permitted. Non-whitelisted enodes are rejected at the transport layer.
 */
interface INodeRules {
    event NodeAdded(bool indexed active, bytes32 indexed enodeHigh, bytes32 indexed enodeLow, bytes16 ip, uint16 port);
    event NodeRemoved(bool indexed active, bytes32 indexed enodeHigh, bytes32 indexed enodeLow, bytes16 ip, uint16 port);

    function enodeAllowed(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external view returns (bool isAllowed);

    function addEnode(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external;

    function removeEnode(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external;

    function size() external view returns (uint256);
}
