// SPDX-License-Identifier: Apache-2.0
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";
import {CopyTradingVault} from "../../src/yield/CopyTradingVault.sol";
import {ICopyTradingVault} from "../../src/interfaces/yield/ICopyTradingVault.sol";

contract MockVaultERC20 is ERC20 {
    constructor() ERC20("Mock USD Tether", "USDT") {}

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}

contract MockStrategyTarget {
    address public token;
    uint256 public balance;

    constructor(address _token) {
        token = _token;
    }

    function trade(uint256 amount) external returns (bool) {
        balance += amount;
        return true;
    }
}

contract CopyTradingVaultTest is Test {
    CopyTradingVault public implementation;
    CopyTradingVault public vault;
    MockVaultERC20 public assetToken;
    MockStrategyTarget public strategyTarget;

    address public admin = address(0xAD01);
    address public masterTrader = address(0xBA57);
    address public follower1 = address(0x1001);
    address public follower2 = address(0x1002);

    uint256 public performanceFeeBps = 1500; // 15% performance fee
    uint256 public maxCapacity = 1_000_000e18;
    uint256 public minDeposit = 100e18;
    uint256 public lockupPeriod = 1 days;

    function setUp() public {
        assetToken = new MockVaultERC20();
        strategyTarget = new MockStrategyTarget(address(assetToken));

        implementation = new CopyTradingVault();
        bytes memory initData = abi.encodeWithSelector(
            CopyTradingVault.initialize.selector,
            admin,
            masterTrader,
            address(assetToken),
            "Growww Alpha Vault Share",
            "GAV-SHARES",
            performanceFeeBps,
            maxCapacity,
            minDeposit,
            lockupPeriod
        );

        ERC1967Proxy proxy = new ERC1967Proxy(address(implementation), initData);
        vault = CopyTradingVault(address(proxy));

        vm.prank(admin);
        vault.setApprovedStrategyTarget(address(strategyTarget), true);

        // Fund followers
        assetToken.mint(follower1, 100_000e18);
        assetToken.mint(follower2, 100_000e18);

        vm.prank(follower1);
        assetToken.approve(address(vault), type(uint256).max);

        vm.prank(follower2);
        assetToken.approve(address(vault), type(uint256).max);
    }

    function test_Initialization() public {
        assertEq(vault.masterTrader(), masterTrader);
        assertEq(vault.performanceFeeBps(), 1500);
        assertEq(vault.highWaterMark(), 1e18);
        assertEq(vault.maxCapacity(), maxCapacity);
    }

    function test_RevertInvalidPerformanceFee() public {
        CopyTradingVault testVault = new CopyTradingVault();
        // Fee < 1000 bps (10%) reverts
        vm.expectRevert(abi.encodeWithSelector(ICopyTradingVault.InvalidPerformanceFeeBps.selector, 500));
        new ERC1967Proxy(
            address(testVault),
            abi.encodeWithSelector(
                CopyTradingVault.initialize.selector,
                admin,
                masterTrader,
                address(assetToken),
                "Share",
                "SHR",
                500,
                maxCapacity,
                minDeposit,
                lockupPeriod
            )
        );
    }

    function test_DepositAndSharesMinting() public {
        uint256 depositAmount = 10_000e18;
        vm.prank(follower1);
        uint256 shares = vault.deposit(depositAmount, follower1);

        assertEq(shares, depositAmount);
        assertEq(vault.balanceOf(follower1), depositAmount);
        assertEq(vault.totalAssets(), depositAmount);
        assertEq(vault.currentSharePrice(), 1e18);
    }

    function test_DepositRevertsBelowMinimum() public {
        vm.prank(follower1);
        vm.expectRevert(abi.encodeWithSelector(ICopyTradingVault.DepositBelowMinimum.selector, 50e18, minDeposit));
        vault.deposit(50e18, follower1);
    }

    function test_WithdrawalAndLockupEnforcement() public {
        uint256 depositAmount = 5_000e18;
        vm.prank(follower1);
        vault.deposit(depositAmount, follower1);

        // Attempting to withdraw before lockup expires must revert
        vm.prank(follower1);
        vm.expectRevert();
        vault.withdraw(depositAmount, follower1);

        // Advance time past lockup period
        vm.warp(block.timestamp + lockupPeriod + 1);

        uint256 balBefore = assetToken.balanceOf(follower1);
        vm.prank(follower1);
        uint256 assetsReturned = vault.withdraw(depositAmount, follower1);

        assertEq(assetsReturned, depositAmount);
        assertEq(assetToken.balanceOf(follower1) - balBefore, depositAmount);
        assertEq(vault.balanceOf(follower1), 0);
    }

    function test_HighWaterMarkPerformanceFeeHarvesting() public {
        uint256 depositAmount = 100_000e18;
        vm.prank(follower1);
        vault.deposit(depositAmount, follower1);

        // Simulate trading profit: 20,000 profit donated/earned into vault
        assetToken.mint(address(vault), 20_000e18);

        // Total assets = 120,000e18. Shares = 100,000e18. Share price = 1.20e18
        assertEq(vault.totalAssets(), 120_000e18);
        assertEq(vault.currentSharePrice(), 1.20e18);

        // Harvest performance fee: 15% of 20,000 profit = 3,000 eINR in fee shares
        uint256 masterSharesBefore = vault.balanceOf(masterTrader);
        uint256 feeShares = vault.harvestPerformanceFee();

        assertTrue(feeShares > 0);
        assertEq(vault.balanceOf(masterTrader) - masterSharesBefore, feeShares);
        assertEq(vault.highWaterMark(), vault.currentSharePrice());

        // Calling harvest again without new profit must revert
        vm.expectRevert(
            abi.encodeWithSelector(
                ICopyTradingVault.NoProfitAboveHighWaterMark.selector,
                vault.currentSharePrice(),
                vault.highWaterMark()
            )
        );
        vault.harvestPerformanceFee();
    }

    function test_StrategyExecutionByMasterTrader() public {
        uint256 depositAmount = 50_000e18;
        vm.prank(follower1);
        vault.deposit(depositAmount, follower1);

        uint256 strategyAmount = 20_000e18;
        bytes memory callData = abi.encodeWithSelector(MockStrategyTarget.trade.selector, strategyAmount);

        vm.prank(masterTrader);
        vault.executeStrategy(address(strategyTarget), strategyAmount, callData);

        assertEq(strategyTarget.balance(), strategyAmount);
        assertEq(vault.deployedStrategyFunds(), strategyAmount);
        assertEq(vault.totalAssets(), depositAmount); // Total assets still includes deployed funds
    }

    function test_StrategyExecutionRevertsUnauthorized() public {
        uint256 depositAmount = 50_000e18;
        vm.prank(follower1);
        vault.deposit(depositAmount, follower1);

        bytes memory callData = abi.encodeWithSelector(MockStrategyTarget.trade.selector, 10_000e18);

        vm.prank(follower1);
        vm.expectRevert(abi.encodeWithSelector(ICopyTradingVault.UnauthorizedMasterTrader.selector, follower1));
        vault.executeStrategy(address(strategyTarget), 10_000e18, callData);
    }
}
