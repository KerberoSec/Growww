// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {AdminRegistry} from "../../src/permissioning/AdminRegistry.sol";
import {NodeRules} from "../../src/permissioning/NodeRules.sol";
import {AccountRules} from "../../src/permissioning/AccountRules.sol";
import {INodeRules} from "../../src/interfaces/permissioning/INodeRules.sol";
import {IAccountRules} from "../../src/interfaces/permissioning/IAccountRules.sol";

contract NodeRulesTest is Test {
    AdminRegistry public adminRegistry;
    NodeRules public nodeRules;
    AccountRules public accountRules;

    address public consortiumAdmin = address(0xAD01);
    address public subAdmin = address(0xAD02);
    address public relayer = address(0xBE01);
    address public unauthorizedUser = address(0xFA01);

    bytes32 public enodeHigh1 = 0x9e8b7c6d5e4f3a2b1c0d9e8b7c6d5e4f3a2b1c0d9e8b7c6d5e4f3a2b1c0d9e8b;
    bytes32 public enodeLow1 = 0x8d7c6b5a4f3e2d1c0b9a8d7c6b5a4f3e2d1c0b9a8d7c6b5a4f3e2d1c0b9a8d7c;
    bytes16 public ip1 = 0x00000000000000000000ffff7f000001; // 127.0.0.1
    uint16 public port1 = 30303;

    function setUp() public {
        // Deploy AdminRegistry
        AdminRegistry adminImpl = new AdminRegistry();
        bytes memory adminInit = abi.encodeWithSelector(AdminRegistry.initialize.selector, consortiumAdmin);
        ERC1967Proxy adminProxy = new ERC1967Proxy(address(adminImpl), adminInit);
        adminRegistry = AdminRegistry(address(adminProxy));

        // Deploy NodeRules
        NodeRules nodeImpl = new NodeRules();
        bytes memory nodeInit = abi.encodeWithSelector(NodeRules.initialize.selector, address(adminRegistry));
        ERC1967Proxy nodeProxy = new ERC1967Proxy(address(nodeImpl), nodeInit);
        nodeRules = NodeRules(address(nodeProxy));

        // Deploy AccountRules
        AccountRules accountImpl = new AccountRules();
        bytes memory accountInit = abi.encodeWithSelector(AccountRules.initialize.selector, address(adminRegistry));
        ERC1967Proxy accountProxy = new ERC1967Proxy(address(accountImpl), accountInit);
        accountRules = AccountRules(address(accountProxy));
    }

    function test_AdminGovernance() public {
        assertTrue(adminRegistry.isAuthorizedAdmin(consortiumAdmin));
        assertFalse(adminRegistry.isAuthorizedAdmin(subAdmin));

        vm.prank(consortiumAdmin);
        adminRegistry.addAdmin(subAdmin);
        assertTrue(adminRegistry.isAuthorizedAdmin(subAdmin));

        vm.prank(consortiumAdmin);
        adminRegistry.removeAdmin(subAdmin);
        assertFalse(adminRegistry.isAuthorizedAdmin(subAdmin));
    }

    function test_NodePermissioningLifecycle() public {
        assertFalse(nodeRules.enodeAllowed(enodeHigh1, enodeLow1, ip1, port1));
        assertEq(nodeRules.size(), 0);

        // Add node by admin
        vm.prank(consortiumAdmin);
        nodeRules.addEnode(enodeHigh1, enodeLow1, ip1, port1);

        assertTrue(nodeRules.enodeAllowed(enodeHigh1, enodeLow1, ip1, port1));
        assertEq(nodeRules.size(), 1);

        // Remove node by admin
        vm.prank(consortiumAdmin);
        nodeRules.removeEnode(enodeHigh1, enodeLow1, ip1, port1);

        assertFalse(nodeRules.enodeAllowed(enodeHigh1, enodeLow1, ip1, port1));
        assertEq(nodeRules.size(), 0);
    }

    function test_NodePermissioningUnauthorizedReverts() public {
        vm.prank(unauthorizedUser);
        vm.expectRevert();
        nodeRules.addEnode(enodeHigh1, enodeLow1, ip1, port1);
    }

    function test_AccountPermissioningLifecycle() public {
        assertFalse(accountRules.isAccountAllowed(relayer));
        assertFalse(accountRules.transactionAllowed(relayer, address(0x123), 0, 0, 21000, ""));
        assertEq(accountRules.size(), 0);

        vm.prank(consortiumAdmin);
        accountRules.addAccount(relayer);

        assertTrue(accountRules.isAccountAllowed(relayer));
        assertTrue(accountRules.transactionAllowed(relayer, address(0x123), 0, 0, 21000, ""));
        assertEq(accountRules.size(), 1);

        vm.prank(consortiumAdmin);
        accountRules.removeAccount(relayer);

        assertFalse(accountRules.isAccountAllowed(relayer));
        assertFalse(accountRules.transactionAllowed(relayer, address(0x123), 0, 0, 21000, ""));
        assertEq(accountRules.size(), 0);
    }

    function test_AccountPermissioningUnauthorizedReverts() public {
        vm.prank(unauthorizedUser);
        vm.expectRevert();
        accountRules.addAccount(relayer);
    }
}
