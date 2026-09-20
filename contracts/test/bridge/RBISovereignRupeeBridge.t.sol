// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "../../src/bridge/RBISovereignRupeeBridge.sol";

contract MockCBDCRupee is ERC20 {
    constructor() ERC20("RBI Sovereign e-Rupee", "eINR") {
        _mint(msg.sender, 1_000_000_000_00); // 10 Crore INR in paise
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract RBISovereignRupeeBridgeTest is Test {
    RBISovereignRupeeBridge public bridge;
    MockCBDCRupee public cbdc;
    WrappedDigitalRupee public wrapped;

    address public admin = address(0xAA);
    address public user = address(0xBB);

    function setUp() public {
        cbdc = new MockCBDCRupee();

        vm.startPrank(admin);
        bridge = new RBISovereignRupeeBridge(address(cbdc), admin);
        wrapped = bridge.wrappedToken();
        bridge.setKYCStatus(user, true);
        vm.stopPrank();

        cbdc.mint(user, 100_000_00); // 1 Lakh INR
        vm.startPrank(user);
        cbdc.approve(address(bridge), type(uint256).max);
        wrapped.approve(address(bridge), type(uint256).max);
        vm.stopPrank();
    }

    function test_LockAndMint_Success() public {
        uint256 depositPaise = 50_000_00; // 50,000 INR

        vm.prank(user);
        bridge.lockAndMint(depositPaise);

        assertEq(bridge.totalLockedReserve(), depositPaise);
        assertEq(wrapped.balanceOf(user), depositPaise);
        assertEq(cbdc.balanceOf(address(bridge)), depositPaise);
    }

    function test_LockAndMint_RevertsWithoutKYC() public {
        address unverified = address(0xCC);
        cbdc.mint(unverified, 10_000_00);

        vm.startPrank(unverified);
        cbdc.approve(address(bridge), type(uint256).max);
        vm.expectRevert(abi.encodeWithSelector(RBISovereignRupeeBridge.KYCRequired.selector));
        bridge.lockAndMint(10_000_00);
        vm.stopPrank();
    }

    function test_LockAndMint_EnforcesDailyLimit() public {
        uint256 limit = bridge.dailyLimitPaise();
        cbdc.mint(user, limit * 2);

        vm.startPrank(user);
        bridge.lockAndMint(limit);

        // Attempting to exceed daily volume
        vm.expectRevert(abi.encodeWithSelector(
            RBISovereignRupeeBridge.DailyLimitExceeded.selector,
            limit,
            1_00,
            limit
        ));
        bridge.lockAndMint(1_00);
        vm.stopPrank();
    }

    function test_BurnAndUnlock_Success() public {
        uint256 amount = 25_000_00;

        vm.startPrank(user);
        bridge.lockAndMint(amount);
        bridge.burnAndUnlock(amount, "RTGS_RBI_REF_991823");
        vm.stopPrank();

        assertEq(bridge.totalLockedReserve(), 0);
        assertEq(wrapped.balanceOf(user), 0);
        assertEq(cbdc.balanceOf(user), 100_000_00);
    }

    function test_EmergencyHaltPreventsTransfers() public {
        vm.prank(admin);
        bridge.toggleEmergencyHalt(true);

        vm.prank(user);
        vm.expectRevert(abi.encodeWithSelector(RBISovereignRupeeBridge.BridgePaused.selector));
        bridge.lockAndMint(10_000_00);
    }
}
