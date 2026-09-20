// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {DigitalRupee} from "../../src/tokens/DigitalRupee.sol";
import {IAccessControl} from "@openzeppelin/contracts/access/IAccessControl.sol";
import {Pausable} from "@openzeppelin/contracts/utils/Pausable.sol";

contract DigitalRupeeTest is Test {
    DigitalRupee internal token;

    address internal admin = address(0xAA1);
    address internal minter = address(0xBB2);
    address internal compliance = address(0xCC3);
    address internal alice = address(0xD01);
    address internal bob = address(0xD02);
    address internal eve = address(0xD03);

    bytes32 internal txRef1 = keccak256("RBI-CBDC-TX-001");
    bytes32 internal txRef2 = keccak256("RBI-CBDC-TX-002");

    event DigitalRupeeMinted(
        address indexed to,
        uint256 amount,
        bytes32 indexed rbiTxRef,
        string payerVPA,
        uint256 timestamp
    );

    event DigitalRupeeRedeemed(
        address indexed from,
        uint256 amount,
        string payeeVPA,
        uint256 timestamp
    );

    event AccountFrozen(address indexed account, string reason);
    event AccountUnfrozen(address indexed account);

    function setUp() public {
        vm.prank(admin);
        token = new DigitalRupee(admin);

        vm.startPrank(admin);
        token.grantRole(token.MINTER_ROLE(), minter);
        token.grantRole(token.COMPLIANCE_ROLE(), compliance);
        vm.stopPrank();
    }

    function test_InitialConfiguration() public view {
        assertEq(token.name(), "RBI Digital Rupee");
        assertEq(token.symbol(), "eINR");
        assertEq(token.decimals(), 18);
        assertEq(token.totalSupply(), 0);
        assertTrue(token.hasRole(token.DEFAULT_ADMIN_ROLE(), admin));
        assertTrue(token.hasRole(token.MINTER_ROLE(), minter));
        assertTrue(token.hasRole(token.COMPLIANCE_ROLE(), compliance));
    }

    function test_MintSuccess() public {
        uint256 mintAmount = 10_000 * 1e18; // 10,000 eINR

        vm.expectEmit(true, true, false, true);
        emit DigitalRupeeMinted(alice, mintAmount, txRef1, "alice@rbi.edr", block.timestamp);

        vm.prank(minter);
        token.mint(alice, mintAmount, txRef1, "alice@rbi.edr");

        assertEq(token.balanceOf(alice), mintAmount);
        assertEq(token.totalSupply(), mintAmount);
        assertTrue(token.isRBITxProcessed(txRef1));
    }

    function test_RevertMint_DuplicateRBITransaction() public {
        uint256 mintAmount = 5_000 * 1e18;

        vm.prank(minter);
        token.mint(alice, mintAmount, txRef1, "alice@rbi.edr");

        vm.prank(minter);
        vm.expectRevert(abi.encodeWithSelector(DigitalRupee.RBITransactionAlreadyProcessed.selector, txRef1));
        token.mint(alice, mintAmount, txRef1, "alice@rbi.edr");
    }

    function test_RevertMint_UnauthorizedMinter() public {
        uint256 mintAmount = 1_000 * 1e18;

        vm.expectRevert(
            abi.encodeWithSelector(
                IAccessControl.AccessControlUnauthorizedAccount.selector,
                eve,
                token.MINTER_ROLE()
            )
        );
        vm.prank(eve);
        token.mint(alice, mintAmount, txRef1, "eve@rbi.edr");
    }

    function test_RevertMint_ZeroAmountOrAddress() public {
        vm.prank(minter);
        vm.expectRevert(DigitalRupee.InvalidAddress.selector);
        token.mint(address(0), 100 * 1e18, txRef1, "alice@rbi.edr");

        vm.prank(minter);
        vm.expectRevert(DigitalRupee.InvalidAmount.selector);
        token.mint(alice, 0, txRef1, "alice@rbi.edr");

        vm.prank(minter);
        vm.expectRevert(DigitalRupee.EmptyVPA.selector);
        token.mint(alice, 100 * 1e18, txRef1, "");
    }

    function test_RedeemSuccess() public {
        uint256 mintAmount = 5_000 * 1e18;
        uint256 redeemAmount = 2_000 * 1e18;

        vm.prank(minter);
        token.mint(alice, mintAmount, txRef1, "alice@rbi.edr");

        vm.expectEmit(true, false, false, true);
        emit DigitalRupeeRedeemed(alice, redeemAmount, "alice@rbi.edr", block.timestamp);

        vm.prank(alice);
        token.redeem(redeemAmount, "alice@rbi.edr");

        assertEq(token.balanceOf(alice), mintAmount - redeemAmount);
        assertEq(token.totalSupply(), mintAmount - redeemAmount);
    }

    function test_RevertRedeem_InsufficientBalanceOrZero() public {
        vm.prank(alice);
        vm.expectRevert(DigitalRupee.InvalidAmount.selector);
        token.redeem(0, "alice@rbi.edr");

        vm.prank(alice);
        vm.expectRevert(DigitalRupee.EmptyVPA.selector);
        token.redeem(100 * 1e18, "");
    }

    function test_AccountFreezeAndSanctions() public {
        uint256 mintAmount = 10_000 * 1e18;
        vm.prank(minter);
        token.mint(alice, mintAmount, txRef1, "alice@rbi.edr");

        // Compliance freezes Alice
        vm.expectEmit(true, false, false, true);
        emit AccountFrozen(alice, "Suspicious CTR Pattern reported to FIU-IND");

        vm.prank(compliance);
        token.freezeAccount(alice, "Suspicious CTR Pattern reported to FIU-IND");
        assertTrue(token.isFrozen(alice));

        // Alice cannot transfer
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(DigitalRupee.AccountIsFrozen.selector, alice));
        token.transfer(bob, 100 * 1e18);

        // Bob cannot transfer to Alice
        vm.prank(minter);
        token.mint(bob, 500 * 1e18, txRef2, "bob@rbi.edr");

        vm.prank(bob);
        vm.expectRevert(abi.encodeWithSelector(DigitalRupee.AccountIsFrozen.selector, alice));
        token.transfer(alice, 50 * 1e18);

        // Alice cannot redeem
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(DigitalRupee.AccountIsFrozen.selector, alice));
        token.redeem(100 * 1e18, "alice@rbi.edr");

        // Unfreeze
        vm.expectEmit(true, false, false, true);
        emit AccountUnfrozen(alice);

        vm.prank(compliance);
        token.unfreezeAccount(alice);
        assertFalse(token.isFrozen(alice));

        // Now transfer succeeds
        vm.prank(alice);
        token.transfer(bob, 100 * 1e18);
        assertEq(token.balanceOf(bob), 600 * 1e18);
    }

    function test_EmergencyCircuitBreakerPause() public {
        uint256 mintAmount = 1_000 * 1e18;
        vm.prank(minter);
        token.mint(alice, mintAmount, txRef1, "alice@rbi.edr");

        vm.prank(admin);
        token.pause();

        // Mint blocked when paused
        vm.prank(minter);
        vm.expectRevert(Pausable.EnforcedPause.selector);
        token.mint(bob, 100 * 1e18, txRef2, "bob@rbi.edr");

        // Transfer blocked when paused
        vm.prank(alice);
        vm.expectRevert(Pausable.EnforcedPause.selector);
        token.transfer(bob, 100 * 1e18);

        // Redeem blocked when paused
        vm.prank(alice);
        vm.expectRevert(Pausable.EnforcedPause.selector);
        token.redeem(100 * 1e18, "alice@rbi.edr");

        // Unpause
        vm.prank(admin);
        token.unpause();

        vm.prank(alice);
        token.transfer(bob, 100 * 1e18);
        assertEq(token.balanceOf(bob), 100 * 1e18);
    }
}
