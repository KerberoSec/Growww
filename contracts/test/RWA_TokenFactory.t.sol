// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import "./TestBase.sol";
import "../DigitalSecurityToken.sol";
import "../TokenFactory.sol";

contract RWA_TokenFactoryTest is TestBase {
    TokenFactory public factory;
    address public admin = address(0xAA11);
    address public compliance = address(0xCC22);
    address public alice = address(0x1111);
    address public bob = address(0x2222);

    function setUp() public {
        vm.prank(admin);
        factory = new TokenFactory(compliance);
    }

    function test_CreateSecurityToken() public {
        vm.prank(admin);
        address tokenAddr = factory.createToken(
            "INE002A01018",
            "Sovereign Gold Bond Token",
            "SGB-2026",
            18,
            1_000_000 * 1e18,
            alice
        );

        DigitalSecurityToken token = DigitalSecurityToken(tokenAddr);
        assertEq(token.name(), "Sovereign Gold Bond Token");
        assertEq(token.symbol(), "SGB-2026");
        assertEq(token.balanceOf(alice), 1_000_000 * 1e18);
        assertTrue(token.isKYCVerified(alice));
    }

    function test_TransferRevertIfRecipientNotKYC() public {
        vm.prank(admin);
        address tokenAddr = factory.createToken("INE001", "DST", "DST", 18, 1000, alice);
        DigitalSecurityToken token = DigitalSecurityToken(tokenAddr);

        vm.expectRevert(abi.encodeWithSelector(DigitalSecurityToken.IdentityNotVerified.selector, bob));
        vm.prank(alice);
        token.transfer(bob, 100);
    }

    function test_TransferSuccessAfterKYC() public {
        vm.prank(admin);
        address tokenAddr = factory.createToken("INE001", "DST", "DST", 18, 1000, alice);
        DigitalSecurityToken token = DigitalSecurityToken(tokenAddr);

        vm.prank(compliance);
        token.setKYCStatus(bob, true);

        vm.prank(alice);
        bool ok = token.transfer(bob, 250);
        assertTrue(ok);
        assertEq(token.balanceOf(bob), 250);
        assertEq(token.balanceOf(alice), 750);
    }
}
