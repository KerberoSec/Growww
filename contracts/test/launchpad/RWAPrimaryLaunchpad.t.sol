// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {RWAPrimaryLaunchpad} from "../../src/launchpad/RWAPrimaryLaunchpad.sol";
import {DigitalRupee} from "../../src/tokens/DigitalRupee.sol";
import {DigitalSecurityToken} from "../../src/rwa/DigitalSecurityToken.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

contract RWAPrimaryLaunchpadTest is Test {
    RWAPrimaryLaunchpad internal launchpad;
    DigitalRupee internal usdt; // payment token (18 decimals)
    DigitalSecurityToken internal tBillToken; // RWA token (18 decimals)

    address internal admin = address(0xA1);
    address internal issuer = address(0xB1);
    address internal investor1 = address(0xC1);
    address internal investor2 = address(0xC2);

    function setUp() public {
        vm.startPrank(admin);
        launchpad = new RWAPrimaryLaunchpad(admin);
        launchpad.grantRole(launchpad.OPERATOR_ROLE(), issuer);

        usdt = new DigitalRupee(admin);
        // Mint payment tokens to investors
        usdt.mint(investor1, 1_000_000 * 1e18, keccak256("USDT-INV1"), "inv1@rbi.edr");
        usdt.mint(investor2, 1_000_000 * 1e18, keccak256("USDT-INV2"), "inv2@rbi.edr");
        vm.stopPrank();

        // Create RWA Token (10,000 units initial supply to issuer)
        vm.startPrank(issuer);
        tBillToken = new DigitalSecurityToken(
            "India 91D T-Bill Token",
            "TBILL91",
            18,
            10_000 * 1e18,
            address(0),
            issuer
        );
        // Whitelist investors for ERC-3643 KYC compliance
        tBillToken.setKYCStatus(investor1, true);
        tBillToken.setKYCStatus(investor2, true);
        tBillToken.setKYCStatus(address(launchpad), true);
        vm.stopPrank();
    }

    function test_FixedPriceIssuance_Success() public {
        uint256 offerAmount = 5_000 * 1e18;
        uint256 pricePerToken = 100 * 1e18; // 100 USDT per T-Bill
        uint256 startTime = block.timestamp + 10;
        uint256 endTime = block.timestamp + 1 days;
        uint256 softCap = 2_000 * 1e18;

        vm.startPrank(issuer);
        tBillToken.approve(address(launchpad), offerAmount);
        uint256 issuanceId = launchpad.createIssuance(
            address(tBillToken),
            address(usdt),
            RWAPrimaryLaunchpad.IssuanceType.FIXED_PRICE,
            offerAmount,
            pricePerToken,
            0,
            0,
            startTime,
            endTime,
            softCap
        );
        vm.stopPrank();

        assertEq(issuanceId, 1);
        assertEq(tBillToken.balanceOf(address(launchpad)), offerAmount);

        // Warp into active window
        vm.warp(startTime + 100);

        // Investor 1 subscribes 3,000 tokens (300,000 USDT cost)
        uint256 subAmount = 3_000 * 1e18;
        uint256 expectedCost = 300_000 * 1e18;

        vm.startPrank(investor1);
        usdt.approve(address(launchpad), expectedCost);
        launchpad.subscribe(issuanceId, subAmount);
        vm.stopPrank();

        // Check subscription
        (uint256 committed, uint256 deposited, bool claimed, bool refunded) =
            launchpad.subscriptions(issuanceId, investor1);
        assertEq(committed, subAmount);
        assertEq(deposited, expectedCost);
        assertFalse(claimed);
        assertFalse(refunded);

        // Warp past end time and finalize
        vm.warp(endTime + 1);
        launchpad.finalizeIssuance(issuanceId);

        // Issuer received payment funds (300,000 USDT)
        assertEq(usdt.balanceOf(issuer), expectedCost);
        // Issuer received unsold tokens (5,000 - 3,000 = 2,000 tokens)
        assertEq(tBillToken.balanceOf(issuer), (10_000 - 3_000) * 1e18);

        // Investor 1 claims tokens
        vm.prank(investor1);
        launchpad.claimTokens(issuanceId);
        assertEq(tBillToken.balanceOf(investor1), subAmount);
    }

    function test_SoftCapFailed_RefundFlow() public {
        uint256 offerAmount = 5_000 * 1e18;
        uint256 pricePerToken = 100 * 1e18;
        uint256 startTime = block.timestamp + 10;
        uint256 endTime = block.timestamp + 1 days;
        uint256 softCap = 3_000 * 1e18; // 3,000 min cap

        vm.startPrank(issuer);
        tBillToken.approve(address(launchpad), offerAmount);
        uint256 issuanceId = launchpad.createIssuance(
            address(tBillToken),
            address(usdt),
            RWAPrimaryLaunchpad.IssuanceType.FIXED_PRICE,
            offerAmount,
            pricePerToken,
            0,
            0,
            startTime,
            endTime,
            softCap
        );
        vm.stopPrank();

        vm.warp(startTime + 100);

        // Investor 1 subscribes only 1,000 tokens (< 3,000 soft cap)
        uint256 subAmount = 1_000 * 1e18;
        uint256 cost = 100_000 * 1e18;

        vm.startPrank(investor1);
        usdt.approve(address(launchpad), cost);
        launchpad.subscribe(issuanceId, subAmount);
        vm.stopPrank();

        // Warp to end and finalize
        vm.warp(endTime + 1);
        launchpad.finalizeIssuance(issuanceId);

        // Issuer receives all 5,000 RWA tokens back
        assertEq(tBillToken.balanceOf(issuer), 10_000 * 1e18);

        // Investor claims full refund
        uint256 beforeBal = usdt.balanceOf(investor1);
        vm.prank(investor1);
        launchpad.claimRefund(issuanceId);
        assertEq(usdt.balanceOf(investor1), beforeBal + cost);
    }

    function test_DutchAuction_PriceDecay() public {
        uint256 offerAmount = 1_000 * 1e18;
        uint256 startPrice = 200 * 1e18;
        uint256 floorPrice = 100 * 1e18;
        vm.warp(100);
        uint256 startTime = 200;
        uint256 endTime = 1200;

        vm.startPrank(issuer);
        tBillToken.approve(address(launchpad), offerAmount);
        uint256 issuanceId = launchpad.createIssuance(
            address(tBillToken),
            address(usdt),
            RWAPrimaryLaunchpad.IssuanceType.DUTCH_AUCTION,
            offerAmount,
            0,
            startPrice,
            floorPrice,
            startTime,
            endTime,
            500 * 1e18
        );
        vm.stopPrank();

        // At start time -> price is startPrice (200)
        vm.warp(startTime);
        assertEq(launchpad.getCurrentPrice(issuanceId), startPrice);

        // Halfway through auction (200 + 500 = 700) -> price should be midpoint (150)
        vm.warp(700);
        assertEq(launchpad.getCurrentPrice(issuanceId), 150 * 1e18);

        // At end time -> floor price (100)
        vm.warp(endTime);
        assertEq(launchpad.getCurrentPrice(issuanceId), floorPrice);
    }
}
