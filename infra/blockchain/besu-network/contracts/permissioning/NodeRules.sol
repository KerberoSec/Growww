// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {INodeRules} from "../interfaces/permissioning/INodeRules.sol";
import {AdminRegistry} from "./AdminRegistry.sol";

/**
 * @title NodeRules
 * @notice Besu on-chain node permissioning rules contract.
 * @dev Enforces P2P transport layer allowlists. Any node connection handshake whose
 *      enode identity is not allowlisted is immediately rejected by the Besu client.
 */
contract NodeRules is Initializable, INodeRules {
    struct EnodeRecord {
        bytes32 enodeHigh;
        bytes32 enodeLow;
        bytes16 ip;
        uint16 port;
        bool active;
    }

    AdminRegistry public adminRegistry;

    // Mapping from enode hash to EnodeRecord
    mapping(bytes32 => EnodeRecord) private _nodes;
    bytes32[] private _nodeKeys;

    error UnauthorizedCaller(address caller);
    error InvalidZeroAddress();
    error NodeAlreadyExists(bytes32 nodeKey);
    error NodeNotFound(bytes32 nodeKey);
    error InvalidEnodeData();

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

    function computeNodeKey(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) public pure returns (bytes32) {
        return keccak256(abi.encode(enodeHigh, enodeLow, ip, port));
    }

    function enodeAllowed(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external view override returns (bool isAllowed) {
        bytes32 nodeKey = computeNodeKey(enodeHigh, enodeLow, ip, port);
        return _nodes[nodeKey].active;
    }

    function addEnode(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external override onlyAdmin {
        if (enodeHigh == bytes32(0) && enodeLow == bytes32(0)) revert InvalidEnodeData();

        bytes32 nodeKey = computeNodeKey(enodeHigh, enodeLow, ip, port);
        if (_nodes[nodeKey].active) revert NodeAlreadyExists(nodeKey);

        _nodes[nodeKey] = EnodeRecord({
            enodeHigh: enodeHigh,
            enodeLow: enodeLow,
            ip: ip,
            port: port,
            active: true
        });
        _nodeKeys.push(nodeKey);

        emit NodeAdded(true, enodeHigh, enodeLow, ip, port);
    }

    function removeEnode(
        bytes32 enodeHigh,
        bytes32 enodeLow,
        bytes16 ip,
        uint16 port
    ) external override onlyAdmin {
        bytes32 nodeKey = computeNodeKey(enodeHigh, enodeLow, ip, port);
        if (!_nodes[nodeKey].active) revert NodeNotFound(nodeKey);

        _nodes[nodeKey].active = false;
        emit NodeRemoved(false, enodeHigh, enodeLow, ip, port);
    }

    function size() external view override returns (uint256) {
        uint256 activeCount = 0;
        for (uint256 i = 0; i < _nodeKeys.length; i++) {
            if (_nodes[_nodeKeys[i]].active) {
                activeCount++;
            }
        }
        return activeCount;
    }
}
