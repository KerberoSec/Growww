// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {CryptoEarnStakingVault} from "../../src/yield/CryptoEarnStakingVault.sol";
import {DigitalRupee} from "../../src/tokens/DigitalRupee.sol";

contract CryptoEarnStakingVaultTest is Test {
    CryptoEarnStakingVault internal vault;
    DigitalRupee internal usdt;

    address internal admin = address(0xAA1);
    address internal operator = address(0xBB1);
    address internal user1 = address(0xCC1);
    address internal user2 = address(0xCC2);

    function setUp() public {
        vm.startPrank(admin);
        vault = new CryptoEarnStakingVault(admin);
        vault.grantRole(vault.OPERATOR_ROLE(), operator);

        usdt = new DigitalRupee(admin);

        // Fund users and the vault reward pool
        usdt.mint(user1, 100_000 * 1e18, keccak256("MINT-U1"), "u1@rbi.edr");
        usdt.mint(user2, 100_000 * 1e18, keccak256("MINT-U2"), "u2@rbi.edr");
        usdt.mint(address(vault), 10_000 * 1e18, keccak256("MINT-VAULT"), "vault@rbi.edr");
        vm.stopPrank();
    }

    function test_FlexibleEarn_DailyAccrualAndRedeem() public {
        vm.startPrank(operator);
        // Flexible product: 500 bps = 5% APR, 100 USDT min deposit, 1,000,000 max capacity
        uint256 pid = vault.createProduct(
            address(usdt),
            CryptoEarnStakingVault.StakingType.FLEXIBLE,
            500, // 5% APR
            100 * 1e18,
            1_000_000 * 1e18
        );
        vm.stopPrank();

        uint256 depositAmt = 10_000 * 1e18; // 10,000 USDT

        vm.startPrank(user1);
        usdt.approve(address(vault), depositAmt);
        uint256 sid = vault.deposit(pid, depositAmt);
        vm.stopPrank();

        assertEq(sid, 1);

        // Warp 365 days forward
        vm.warp(block.timestamp + 365 days);

        // Expected gross interest = 10,000 * 5% = 500 USDT
        // Risk reserve = 10% of 500 = 50 USDT
        // Net yield to user = 450 USDT
        (uint256 gross, uint256 net, uint256 reserveCut) = vault.calculateAccruedInterest(sid);
        assertEq(gross, 500 * 1e18);
        assertEq(reserveCut, 50 * 1e18);
        assertEq(net, 450 * 1e18);

        // Redeem flexible savings immediately
        uint256 balBefore = usdt.balanceOf(user1);
        vm.prank(user1);
        uint256 totalReturned = vault.redeem(sid);

        assertEq(totalReturned, depositAmt + net);
        assertEq(usdt.balanceOf(user1), balBefore + totalReturned);
    }

    function test_FixedStaking_MaturityFlow() public {
        vm.startPrank(operator);
        // Fixed 30-day product: 1200 bps = 12% APR
        uint256 pid = vault.createProduct(
            address(usdt),
            CryptoEarnStakingVault.StakingType.FIXED_30,
            1200, // 12% APR
            500 * 1e18,
            500_000 * 1e18
        );
        vm.stopPrank();

        uint256 depositAmt = 20_000 * 1e18;

        vm.startPrank(user2);
        usdt.approve(address(vault), depositAmt);
        uint256 sid = vault.deposit(pid, depositAmt);
        vm.stopPrank();

        // Warp to exact maturity (30 days)
        vm.warp(block.timestamp + 30 days);

        uint256 balBefore = usdt.balanceOf(user2);
        vm.prank(user2);
        uint256 totalReturned = vault.redeem(sid);

        // User got back more than principal
        assertGt(totalReturned, depositAmt);
        assertEq(usdt.balanceOf(user2), balBefore + totalReturned);
    }

    function test_FixedStaking_EarlyRedemptionPenalty() public {
        vm.startPrank(operator);
        uint256 pid = vault.createProduct(
            address(usdt),
            CryptoEarnStakingVault.StakingType.FIXED_30,
            1200,
            500 * 1e18,
            500_000 * 1e18
        );
        vm.stopPrank();

        uint256 depositAmt = 10_000 * 1e18;

        vm.startPrank(user1);
        usdt.approve(address(vault), depositAmt);
        uint256 sid = vault.deposit(pid, depositAmt);

        // Warp only 10 days (before 30 days unlock)
        vm.warp(block.timestamp + 10 days);

        // Early redeem: forfeits unharvested interest, gets only principal back
        uint256 balBefore = usdt.balanceOf(user1);
        uint256 totalReturned = vault.redeem(sid);
        vm.stopPrank();

        assertEq(totalReturned, depositAmt);
        assertEq(usdt.balanceOf(user1), balBefore + depositAmt);
    }

    function test_ClaimYieldSeparately() public {
        vm.startPrank(operator);
        uint256 pid = vault.createProduct(
            address(usdt),
            CryptoEarnStakingVault.StakingType.FLEXIBLE,
            1000, // 10% APR
            100 * 1e18,
            1_000_000 * 1e18
        );
        vm.stopPrank();

        uint256 depositAmt = 10_000 * 1e18;

        vm.startPrank(user1);
        usdt.approve(address(vault), depositAmt);
        uint256 sid = vault.deposit(pid, depositAmt);

        // Warp 180 days
        vm.warp(block.timestamp + 180 days);

        uint256 balBefore = usdt.balanceOf(user1);
        uint256 claimedNet = vault.claimYield(sid);
        vm.stopPrank();

        assertGt(claimedNet, 0);
        assertEq(usdt.balanceOf(user1), balBefore + claimedNet);

        // Verify risk reserve accumulated in vault state
        assertGt(vault.riskReserveBalances(address(usdt)), 0);
    }
}
